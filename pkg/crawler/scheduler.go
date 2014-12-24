package crawler

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/janicduplessis/resultscrawler/pkg/api"
	"github.com/janicduplessis/resultscrawler/pkg/store/crawlerconfig"
	"github.com/janicduplessis/resultscrawler/pkg/store/results"
	"github.com/janicduplessis/resultscrawler/pkg/store/user"
	"github.com/janicduplessis/resultscrawler/pkg/tools"
)

const (
	// Time between checks to see if a user needs an update in seconds
	checkInterval time.Duration = 30 * time.Second
	// Time between updates for each user in minutes
	updateInterval time.Duration = 1 * time.Minute
)

var msgTemplatePath = "msgtemplate.html"
var msgTemplate *template.Template

type crawlerUser struct {
	ID      string
	Classes []api.Class
	Nip     string
	Code    string
	Name    string
	Email   string
}

// SchedulerConfig initializes the scheduler.
type SchedulerConfig struct {
	Crawlers           []*Crawler
	UserStore          user.Store
	CrawlerConfigStore crawlerconfig.Store
	UserResultsStore   results.Store
	Sender             tools.Sender
}

// Scheduler handles scheduling crawler runs for every user.
type Scheduler struct {
	Crawlers           []*Crawler
	UserStore          user.Store
	CrawlerConfigStore crawlerconfig.Store
	UserResultsStore   results.Store
	Sender             tools.Sender

	queueCh chan *crawlerUser
	doneCh  chan bool
}

// NewScheduler creates a new scuduler object.
func NewScheduler(config *SchedulerConfig) *Scheduler {
	if msgTemplate == nil {
		msgTemplate = template.Must(template.New("msgtemplate.html").ParseFiles(msgTemplatePath))
	}

	queueCh := make(chan *crawlerUser)
	doneCh := make(chan bool)

	return &Scheduler{
		config.Crawlers,
		config.UserStore,
		config.CrawlerConfigStore,
		config.UserResultsStore,
		config.Sender,

		queueCh,
		doneCh,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	for _, crawler := range s.Crawlers {
		go s.crawlerLoop(crawler)
	}

	s.mainLoop()
}

// Queue tells the scheduler do a run for a user
func (s *Scheduler) Queue(user *api.User) {
	// Get the user current results
	results, err := s.UserResultsStore.GetResults(user.ID)
	if err != nil {
		log.Println(err)
		return
	}

	// Check la update time.
	if time.Now().Sub(results.LastUpdate) < updateInterval {
		return
	}

	// Get crawler config.
	crawlerConfig, err := s.CrawlerConfigStore.GetCrawlerConfig(user.ID)
	if err != nil {
		log.Println(err)
		return
	}

	s.queueCh <- &crawlerUser{
		ID:      user.ID,
		Classes: results.Classes,
		Code:    crawlerConfig.Code,
		Nip:     crawlerConfig.Nip,
		Email:   crawlerConfig.NotificationEmail,
		Name:    fmt.Sprintf("%s %s", user.FirstName, user.LastName),
	}
}

func (s *Scheduler) crawlerLoop(crawler *Crawler) {
	for {
		select {
		case user := <-s.queueCh:
			// Get results
			results := crawler.Run(user)
			// Check if results changed
			newRes := getNewResults(user, results)
			if len(newRes) > 0 {
				log.Printf("New results for user %s", user.Email)
				if len(user.Email) > 0 {
					err := s.sendEmail(user, newRes)
					if err != nil {
						log.Println(err)
					}
				}
				// Update results
				for _, res := range results {
					// Ignore results with errors
					if res.Err == nil {
						user.Classes[res.ClassIndex].Results = res.Class.Results
						user.Classes[res.ClassIndex].Total = res.Class.Total
						user.Classes[res.ClassIndex].Final = res.Class.Final
					}
				}
			}

			err := s.UserResultsStore.UpdateResults(&api.Results{
				UserID:     user.ID,
				Classes:    user.Classes,
				LastUpdate: time.Now(),
			})
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// The scheduler main loop
func (s *Scheduler) mainLoop() {
	ticker := time.NewTicker(checkInterval)
	for {
		select {
		case <-ticker.C:
			users, err := s.UserStore.ListUsers()
			if err != nil {
				log.Println(err)
				break
			}

			// Check which users need to update
			for _, user := range users {
				s.Queue(user)
			}
		case <-s.doneCh:
			// Stop the program
			return
		}
	}
}

func (s *Scheduler) sendEmail(user *crawlerUser, newResults []api.Class) error {
	var msg bytes.Buffer
	data := struct {
		User       *crawlerUser
		NewClasses []api.Class
	}{
		user,
		newResults,
	}
	err := msgTemplate.Execute(&msg, data)
	if err != nil {
		return err
	}

	return s.Sender.Send(user.Email, "You have new results!", string(msg.Bytes()))
}

func getNewResults(user *crawlerUser, newResults []runResult) []api.Class {
	var resClasses []api.Class
	for i, resInfo := range newResults {
		if resInfo.Err != nil {
			continue
		}

		var classChanged bool
		var curResults []api.Result
		for j, res := range resInfo.Class.Results {
			if user.Classes[i].Final != resInfo.Class.Final ||
				user.Classes[i].Total.Result != resInfo.Class.Total.Result ||
				user.Classes[i].Total.Average != resInfo.Class.Total.Average {
				classChanged = true
			}
			if len(user.Classes[i].Results) <= j {
				// If the is a new result
				curResults = append(curResults, res)
			} else {
				// Check if a result changed
				oldRes := user.Classes[i].Results[j]
				if oldRes.Name != res.Name ||
					oldRes.Normal.Average != res.Normal.Average ||
					oldRes.Normal.Result != res.Normal.Result {
					curResults = append(curResults, res)
				}
			}
		}
		if len(curResults) > 0 || classChanged {
			classInfo := user.Classes[i]
			resClasses = append(resClasses, api.Class{
				Name:    classInfo.Name,
				Group:   classInfo.Group,
				Year:    classInfo.Year,
				Results: curResults,
				Final:   resInfo.Class.Final,
				Total:   resInfo.Class.Total,
			})
		}
	}
	return resClasses
}
