package crawler

import (
	"github.com/janicduplessis/resultscrawler/lib"
)

type Scheduler struct {
	UserStore lib.UserStore

	queueCh chan *lib.User
	doneCh  chan bool
}

func NewScheduler(userStore lib.UserStore) *Scheduler {
	queueCh := make(chan *lib.User)
	doneCh := make(chan bool)

	return &Scheduler{
		UserStore: userStore,

		queueCh: queueCh,
		doneCh:  doneCh,
	}
}

func (s *Scheduler) Start() {
	crawler := new(Crawler)
	for {
		select {
		case user := <-s.queueCh:
			crawler.Run(user)
		case <-s.doneCh:
			return
		}
	}
}

func (s *Scheduler) Queue(user *lib.User) {
	s.queueCh <- user
}
