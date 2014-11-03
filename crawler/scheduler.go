package crawler

import (
	"bytes"
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
	updateIntervalMin time.Duration = 1
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
			res, err := crawler.Run(user)
			if err != nil {
				log.Println(err.Error())
			}
			// Check if results changed
			if s.hasNewResults(user, res) {
				log.Println("New results")
				newRes := s.getNewResults(user, res)
				err := s.sendEmail(user, newRes)
				if err != nil {
					log.Println(err.Error())
				}
				// Update results
				user.Classes = res
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

// Compare the user results with a new set of results and check if they are
// different.
func (s *Scheduler) hasNewResults(user *lib.User, newClasses []lib.Class) bool {
	if len(user.Classes) != len(newClasses) {
		return true
	}
	for i, class := range user.Classes {
		if len(class.Results) != len(newClasses[i].Results) {
			return true
		}
		for j, res := range class.Results {
			newRes := newClasses[i].Results[j]
			if newRes.Name != res.Name ||
				newRes.Average != res.Average ||
				newRes.Result != res.Result {
				return true
			}
		}
	}
	return false
}

func (s *Scheduler) getNewResults(user *lib.User, newClasses []lib.Class) []lib.Class {
	results := make([]lib.Class, len(newClasses))
	for i, class := range newClasses {
		// Copy class info
		results[i] = class
		results[i].Results = nil
		for j, res := range class.Results {
			if len(user.Classes[i].Results) <= j {
				// If the is a new result
				results[i].Results = append(results[i].Results, res)
			} else {
				// Check if a result changed
				oldRes := user.Classes[i].Results[j]
				if oldRes.Name != res.Name ||
					oldRes.Average != res.Average ||
					oldRes.Result != res.Result {
					results[i].Results = append(results[i].Results, res)
				}
			}
		}
	}
	return results
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
