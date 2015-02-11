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
	updateInterval time.Duration = 10 * time.Minute
)

var (
	msgTemplatePath = "msgtemplate.html"
	msgTemplate     *template.Template
)

// ResultGetter is an interface for something that fetches results.
type ResultGetter interface {
	// Run fetches results for a user.
	Run(user *User) []RunResult
}

// User contains info about the user of a ResultGetter run.
type User struct {
	ID      string
	Classes []api.Class
	Nip     string
	Code    string
	Name    string
	Email   string
	DoneCh  chan bool
}

// RunResult contains the result of a ResultGetter run for a class.
type RunResult struct {
	ClassIndex int
	Class      *api.Class
	Err        error
}

// SchedulerConfig initializes the scheduler.
type SchedulerConfig struct {
	ResultGetters      []ResultGetter
	UserStore          user.Store
	CrawlerConfigStore crawlerconfig.Store
	UserResultsStore   results.Store
	Sender             tools.Sender
}

// Scheduler handles scheduling crawler runs for every user.
type Scheduler struct {
	resultGetters      []ResultGetter
	userStore          user.Store
	crawlerConfigStore crawlerconfig.Store
	userResultsStore   results.Store
	sender             tools.Sender

	queueCh chan *User
	doneCh  chan bool
}

// NewScheduler creates a new scuduler object.
func NewScheduler(config *SchedulerConfig) *Scheduler {
	if msgTemplate == nil {
		msgTemplate = template.Must(template.New("msgtemplate.html").ParseFiles(msgTemplatePath))
	}

	queueCh := make(chan *User, len(config.ResultGetters))
	doneCh := make(chan bool)

	return &Scheduler{
		config.ResultGetters,
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
	for _, getter := range s.resultGetters {
		go s.crawlerLoop(getter)
	}

	s.mainLoop()
}

// Stop stops the scheduler.
func (s *Scheduler) Stop() {
	// Stop the main loop.
	s.doneCh <- true
	// Stop the crawlers.
	for i := 0; i < len(s.resultGetters); i++ {
		s.doneCh <- true
	}
}

// Queue tells the scheduler do a run for a user
func (s *Scheduler) Queue(user *api.User) {
	doneCh := make(chan bool)
	s.QueueAsync(user, doneCh)
	<-doneCh
}

// QueueAsync tells the scheduler do a run for a user async
func (s *Scheduler) QueueAsync(user *api.User, doneCh chan bool) {
	// Get the user current results
	results, err := s.userResultsStore.GetResults(user.ID)
	if err != nil {
		log.Println(err)
		return
	}

	s.queueInternal(user, results, doneCh)
}

func (s *Scheduler) crawlerLoop(crawler ResultGetter) {
	for {
		select {
		case user := <-s.queueCh:
			s.run(user, crawler)
		case <-s.doneCh:
			return
		}
	}
}

// The scheduler main loop.
// Checks if any user needs to be updated every checkInterval.
func (s *Scheduler) mainLoop() {
	ticker := time.NewTicker(checkInterval)
	for {
		select {
		case <-ticker.C:
			users, err := s.userStore.ListUsers()
			if err != nil {
				log.Println(err)
				break
			}

			// Check which users need to update
			for _, user := range users {
				// Get the user current results
				results, err := s.userResultsStore.GetResults(user.ID)
				if err != nil {
					log.Println(err)
					continue
				}

				// Check last update time.
				if time.Now().Sub(results.LastUpdate) < updateInterval {
					continue
				}

				s.queueInternal(user, results, nil)
			}
		case <-s.doneCh:
			// Stop the program
			return
		}
	}
}

func (s *Scheduler) queueInternal(user *api.User, results *api.Results, doneCh chan bool) {
	// Get crawler config.
	crawlerConfig, err := s.crawlerConfigStore.GetCrawlerConfig(user.ID)
	if err != nil {
		log.Println(err)
		return
	}

	s.queueCh <- &User{
		ID:      user.ID,
		Classes: results.Classes,
		Code:    crawlerConfig.Code,
		Nip:     crawlerConfig.Nip,
		Email:   crawlerConfig.NotificationEmail,
		Name:    fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		DoneCh:  doneCh,
	}
}

func (s *Scheduler) run(user *User, crawler ResultGetter) {
	// Notify the caller when the run is done.
	defer func() {
		if user.DoneCh != nil {
			user.DoneCh <- true
			close(user.DoneCh)
		}
	}()
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

	err := s.userResultsStore.UpdateResults(&api.Results{
		UserID:     user.ID,
		Classes:    user.Classes,
		LastUpdate: time.Now(),
	})
	if err != nil {
		log.Println(err)
	}
}

// sendEmail notifies the user by email when he has new results.
func (s *Scheduler) sendEmail(user *User, newResults []api.Class) error {
	var msg bytes.Buffer
	data := struct {
		User       *User
		NewClasses []api.Class
	}{
		user,
		newResults,
	}
	err := msgTemplate.Execute(&msg, data)
	if err != nil {
		return err
	}

	return s.sender.Send(user.Email, "You have new results!", string(msg.Bytes()))
}

// getNewResults compares current results to new results fetched by a ResultGetter
// and returns results that have changed.
func getNewResults(user *User, newResults []RunResult) []api.Class {
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
