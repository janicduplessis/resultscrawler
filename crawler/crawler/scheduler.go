package crawler

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/lib"
)

const (
	// Time between checks to see if a user needs an update in seconds
	checkIntervalSec time.Duration = 30
	// Time between updates for each user in minutes
	updateIntervalMin time.Duration = 30
)

var msgTemplate = template.Must(template.New("msgtemplate.html").ParseFiles("crawler/msgtemplate.html"))

type userInfo struct {
	ID         bson.ObjectId
	LastUpdate time.Time
}

// Scheduler handles scheduling crawler runs for every users.
type Scheduler struct {
	Crawlers  []*Crawler
	UserStore lib.UserStore
	Sender    lib.Sender

	usersInfo       []*userInfo
	queueCh         chan *lib.User
	doneCh          chan bool
	messageTemplate *template.Template
}

// NewScheduler creates a new scuduler object.
func NewScheduler(crawlers []*Crawler, userStore lib.UserStore, sender lib.Sender) *Scheduler {
	queueCh := make(chan *lib.User)
	doneCh := make(chan bool)

	return &Scheduler{
		Crawlers:  crawlers,
		UserStore: userStore,
		Sender:    sender,

		queueCh: queueCh,
		doneCh:  doneCh,
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
		log.Println(err.Error())
		return
	}
	s.queueCh <- user
}

func (s *Scheduler) crawlerLoop(crawler *Crawler) {
	for {
		select {
		case user := <-s.queueCh:
			// Get results
			results := crawler.Run(user)
			// Check if results changed
			newRes := s.getNewResults(user, results)
			if len(newRes) > 0 {
				log.Println(fmt.Sprintf("Found difference: %+v", newRes))
				log.Println(fmt.Sprintf("Old results: %+v", user.Classes))
				log.Println(fmt.Sprintf("New results: %+v", results))
				err := s.sendEmail(user, newRes)
				if err != nil {
					log.Println(err.Error())
				}
				// Update results
				for _, res := range results {
					// Ignore results with errors
					if res.Err == nil {
						user.Classes[res.ClassIndex].Results = res.Results
					}
				}
				log.Println(fmt.Sprintf("Classes before update: %+v", user.Classes))
				err = s.UserStore.Update(user)
				if err != nil {
					log.Println(err.Error())
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

func (s *Scheduler) getNewResults(user *lib.User, newResults []runResult) []lib.Class {
	var resClasses []lib.Class
	for i, resInfo := range newResults {
		if resInfo.Err != nil {
			continue
		}

		var curResults []lib.Result
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
			resClasses = append(resClasses, lib.Class{
				Name:    classInfo.Name,
				Group:   classInfo.Group,
				Year:    classInfo.Year,
				Results: curResults,
			})
		}
	}
	return resClasses
}

func (s *Scheduler) sendEmail(user *lib.User, newResults []lib.Class) error {
	var msg bytes.Buffer
	data := struct {
		User       *lib.User
		NewClasses []lib.Class
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
