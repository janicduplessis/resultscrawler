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
	updateIntervalMin time.Duration = 30
)

type userInfo struct {
	ID         bson.ObjectId
	LastUpdate time.Time
}

// Scheduler handles scheduling crawler runs for every users.
type Scheduler struct {
	userStore lib.UserStore
	usersInfo []*userInfo
	queueCh   chan *lib.User
	doneCh    chan bool
}

// NewScheduler creates a new scuduler object.
func NewScheduler(userStore lib.UserStore) *Scheduler {
	queueCh := make(chan *lib.User)
	doneCh := make(chan bool)

	return &Scheduler{
		userStore: userStore,

		queueCh: queueCh,
		doneCh:  doneCh,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {

	users, err := s.userStore.FindAll()
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

	go s.mainLoop()

	crawler := new(Crawler)
	for {
		select {
		case user := <-s.queueCh:
			// Get results
			res, err := crawler.Run(user)
			if err != nil {
				log.Println(err.Error())
				break
			}
			// Check if results changed
			if s.hasNewResults(user, res) {
				log.Println("New results")
				// Update results
				user.Classes = res
				err := s.userStore.Update(user)
				if err != nil {
					log.Println(err.Error())
				}
			}
		// Stop the program
		case <-s.doneCh:
			return
		}
	}
}

// Queue tells the scheduler do a run for a user
func (s *Scheduler) Queue(userID bson.ObjectId) {
	user, err := s.userStore.FindByID(userID)
	if err != nil {
		log.Println(err.Error())
		return
	}
	s.queueCh <- user
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
