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
	updateInterval time.Duration = 30 * time.Minute
)

var msgTemplatePath = "msgtemplate.html"
var msgTemplate *template.Template

type userInfo struct {
	ID         string
	LastUpdate time.Time
}

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

	usersInfo []*userInfo
	queueCh   chan *crawlerUser
	doneCh    chan bool
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

		nil,
		queueCh,
		doneCh,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	users, err := s.UserStore.ListUsers()
	if err != nil {
		log.Fatal(err)
	}
	s.usersInfo = make([]*userInfo, len(users))
	for i, user := range users {
		s.usersInfo[i] = &userInfo{
			ID:         user.ID,
			LastUpdate: time.Now(),
		}
	}

	for _, crawler := range s.Crawlers {
		go s.crawlerLoop(crawler)
	}

	s.mainLoop()
}

// Queue tells the scheduler do a run for a user
func (s *Scheduler) Queue(userID string) {
	user, err := s.UserStore.GetUser(userID)
	if err != nil {
		log.Println(err)
		return
	}
	// Get crawler config.
	crawlerConfig, err := s.CrawlerConfigStore.GetCrawlerConfig(userID)
	if err != nil {
		log.Println(err)
		return
	}

	// Get the user current results
	results, err := s.UserResultsStore.GetResults(userID)
	if err != nil {
		log.Println(err)
		return
	}

	s.queueCh <- &crawlerUser{
		ID:      userID,
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
						user.Classes[res.ClassIndex].Results = res.Results
					}
				}
				err := s.UserResultsStore.UpdateResults(&api.Results{
					UserID:  user.ID,
					Classes: user.Classes,
				})
				if err != nil {
					log.Println(err)
				}
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
			// Check which users need to update
			for _, userInfo := range s.usersInfo {
				if time.Now().Sub(userInfo.LastUpdate) > updateInterval {
					s.Queue(userInfo.ID)
					userInfo.LastUpdate = time.Now()
				}
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

		var curResults []api.Result
		for j, res := range resInfo.Results {
			if len(user.Classes[i].Results) <= j {
				// If the is a new result
				curResults = append(curResults, res)
			} else {
				// Check if a result changed
				oldRes := user.Classes[i].Results[j]
				if oldRes.Name != res.Name ||
					oldRes.Average != res.Average ||
					oldRes.Result != res.Result {
					curResults = append(curResults, res)
				}
			}
		}
		if len(curResults) > 0 {
			classInfo := user.Classes[i]
			resClasses = append(resClasses, api.Class{
				Name:    classInfo.Name,
				Group:   classInfo.Group,
				Year:    classInfo.Year,
				Results: curResults,
			})
		}
	}
	return resClasses
}
