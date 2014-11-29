package crawler

import (
	"bytes"
	"log"
	"text/template"
	"time"

	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/pkg/crypto"
	"github.com/janicduplessis/resultscrawler/pkg/logger"
	"github.com/janicduplessis/resultscrawler/pkg/store"
	"github.com/janicduplessis/resultscrawler/pkg/utils"
)

const (
	// Time between checks to see if a user needs an update in seconds
	checkIntervalSec time.Duration = 30
	// Time between updates for each user in minutes
	updateIntervalMin time.Duration = 30
)

var msgTemplatePath = "msgtemplate.html"
var msgTemplate *template.Template

type userInfo struct {
	ID         bson.ObjectId
	LastUpdate time.Time
}

type crawlerUser struct {
	ID      bson.ObjectId
	Classes []store.Class
	Nip     string
	Code    string
	Email   string
}

// SchedulerConfig initializes the scheduler.
type SchedulerConfig struct {
	Crawlers           []*Crawler
	UserStore          store.UserStore
	CrawlerConfigStore store.CrawlerConfigStore
	UserResultsStore   store.UserResultsStore
	Crypto             crypto.Crypto
	Sender             utils.Sender
	Logger             logger.Logger
}

// Scheduler handles scheduling crawler runs for every user.
type Scheduler struct {
	Crawlers           []*Crawler
	UserStore          store.UserStore
	CrawlerConfigStore store.CrawlerConfigStore
	UserResultsStore   store.UserResultsStore
	Crypto             crypto.Crypto
	Sender             utils.Sender
	Logger             logger.Logger

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
		config.Crypto,
		config.Sender,
		config.Logger,

		nil,
		queueCh,
		doneCh,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	users, err := s.UserStore.FindAll()
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
func (s *Scheduler) Queue(userID bson.ObjectId) {
	user, err := s.UserStore.FindByID(userID)
	if err != nil {
		s.Logger.Error(err)
		return
	}
	// Get crawler config.
	crawlerConfig, err := s.CrawlerConfigStore.FindByID(userID)
	if err != nil {
		s.Logger.Error(err)
		return
	}

	// Decrypt the user code and nip
	data, err := s.Crypto.AESDecrypt(crawlerConfig.Code)
	if err != nil {
		s.Logger.Error(err)
		return
	}
	userCode := string(data)
	data, err = s.Crypto.AESDecrypt(crawlerConfig.Nip)
	if err != nil {
		s.Logger.Error(err)
		return
	}
	userNip := string(data)

	// Get the user current results
	results, err := s.UserResultsStore.FindByID(userID)
	if err != nil {
		s.Logger.Error(err)
		return
	}

	s.queueCh <- &crawlerUser{
		ID:      userID,
		Classes: results.Classes,
		Code:    userCode,
		Nip:     userNip,
		Email:   user.Email,
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
				s.Logger.Logf("Found difference: %+v", newRes)
				s.Logger.Logf("Old results: %+v", user.Classes)
				s.Logger.Logf("New results: %+v", results)
				err := s.sendEmail(user, newRes)
				if err != nil {
					s.Logger.Error(err)
				}
				// Update results
				for _, res := range results {
					// Ignore results with errors
					if res.Err == nil {
						user.Classes[res.ClassIndex].Results = res.Results
					}
				}
				s.Logger.Logf("Classes before update: %+v", user.Classes)
				err = s.UserResultsStore.Update(&store.UserResults{
					UserID:  user.ID,
					Classes: user.Classes,
				})
				if err != nil {
					s.Logger.Error(err)
				}
			}
		}
	}
}

// The scheduler main loop
func (s *Scheduler) mainLoop() {
	ticker := time.NewTicker(checkIntervalSec * time.Second)
	for {
		select {
		case <-ticker.C:
			// Check which users need to update
			for _, userInfo := range s.usersInfo {
				if time.Now().Sub(userInfo.LastUpdate) > updateIntervalMin*time.Minute {
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

func (s *Scheduler) sendEmail(user *crawlerUser, newResults []store.Class) error {
	var msg bytes.Buffer
	data := struct {
		User       *crawlerUser
		NewClasses []store.Class
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

func getNewResults(user *crawlerUser, newResults []runResult) []store.Class {
	var resClasses []store.Class
	for i, resInfo := range newResults {
		if resInfo.Err != nil {
			continue
		}

		var curResults []store.Result
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
			resClasses = append(resClasses, store.Class{
				Name:    classInfo.Name,
				Group:   classInfo.Group,
				Year:    classInfo.Year,
				Results: curResults,
			})
		}
	}
	return resClasses
}
