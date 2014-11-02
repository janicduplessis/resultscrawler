package crawler

import (
	"log"
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

type userInfo struct {
	ID         bson.ObjectId
	LastUpdate time.Time
}

// Scheduler handles scheduling crawler runs for every users.
type Scheduler struct {
	Crawlers  []*Crawler
	UserStore lib.UserStore

	usersInfo []*userInfo
	queueCh   chan *lib.User
	doneCh    chan bool
}

// NewScheduler creates a new scuduler object.
func NewScheduler(crawlers []*Crawler, userStore lib.UserStore) *Scheduler {
	queueCh := make(chan *lib.User)
	doneCh := make(chan bool)

	return &Scheduler{
		Crawlers:  crawlers,
		UserStore: userStore,

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
				// Update results
				user.Classes = res
				err := s.UserStore.Update(user)
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
