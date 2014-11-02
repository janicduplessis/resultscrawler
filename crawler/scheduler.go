package crawler

import (
	"log"
	"time"

	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/lib"
)

const (
	checkIntervalSec  time.Duration = 30
	updateIntervalMin time.Duration = 1
)

type userInfo struct {
	ID         bson.ObjectId
	LastUpdate time.Time
}

type Scheduler struct {
	userStore lib.UserStore
	usersInfo []*userInfo
	queueCh   chan *lib.User
	doneCh    chan bool
}

func NewScheduler(userStore lib.UserStore) *Scheduler {
	queueCh := make(chan *lib.User)
	doneCh := make(chan bool)

	return &Scheduler{
		userStore: userStore,

		queueCh: queueCh,
		doneCh:  doneCh,
	}
}

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

func (s *Scheduler) Queue(userId bson.ObjectId) {
	user, err := s.userStore.FindById(userId)
	if err != nil {
		log.Println(err.Error())
		return
	}
	s.queueCh <- user
}

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
