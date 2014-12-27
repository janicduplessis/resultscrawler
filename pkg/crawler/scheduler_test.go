package crawler

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/pkg/api"
)

type FakeCrawler struct{}

func (c *FakeCrawler) Run(user *User) []RunResult {
	if getResultsFunc != nil {
		return getResultsFunc()
	}
	return nil
}

type TestUser struct {
	User          *api.User
	CrawlerConfig *api.CrawlerConfig
	Results       *api.Results
}

type FakeStore struct {
	Data map[string]*TestUser
	mut  sync.RWMutex
}

func (s *FakeStore) GetCrawlerConfig(userID string) (*api.CrawlerConfig, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.Data[userID].CrawlerConfig, nil
}

func (s *FakeStore) UpdateCrawlerConfig(crawlerConfig *api.CrawlerConfig) error {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.Data[crawlerConfig.UserID].CrawlerConfig = crawlerConfig
	return nil
}

func (s *FakeStore) GetResults(userID string) (*api.Results, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.Data[userID].Results, nil
}

func (s *FakeStore) UpdateResults(userResults *api.Results) error {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.Data[userResults.UserID].Results = userResults
	return nil
}

func (s *FakeStore) GetUser(id string) (*api.User, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.Data[id].User, nil
}

func (s *FakeStore) GetUserForLogin(email string) (*api.User, string, error) {
	panic("Not implemented")
}

func (s *FakeStore) ListUsers() ([]*api.User, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	var users []*api.User
	for _, u := range s.Data {
		users = append(users, u.User)
	}
	return users, nil
}

func (s *FakeStore) UpdateUser(user *api.User) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	return nil
}

func (s *FakeStore) CreateUser(user *api.User, password string) error {
	s.mut.Lock()
	defer s.mut.Unlock()
	user.ID = bson.NewObjectId().Hex()
	s.Data[user.ID] = &TestUser{
		user,
		&api.CrawlerConfig{
			UserID:            user.ID,
			Status:            true,
			NotificationEmail: user.Email,
		},
		&api.Results{
			UserID: user.ID,
		},
	}
	return nil
}

type FakeSender struct {
}

func (s *FakeSender) Send(to, subject, message string) error {
	if sendFunc != nil {
		sendFunc(to, subject, message)
	}

	return nil
}

var (
	sendFunc       func(string, string, string)
	getResultsFunc func() []RunResult
)

func TestSchedulerNewResults(t *testing.T) {
	scheduler, store := start()

	getResultsFunc = func() (res []RunResult) {
		// Fake data returned by the fake crawler.
		res = append(res, RunResult{
			ClassIndex: 0,
			Class: &api.Class{
				ID:    "randomid",
				Name:  "Random Class",
				Group: "21",
				Year:  "20142",
				Results: []api.Result{
					api.Result{
						Name:     "A result",
						Normal:   api.ResultInfo{},
						Weighted: api.ResultInfo{},
					},
				},
				Total: api.ResultInfo{},
				Final: "",
			},
			Err: nil,
		})
		return res
	}

	var messageSent bool

	sendFunc = func(to, subject, message string) {
		messageSent = true
	}

	user := &api.User{
		Email:     "random@user.com",
		FirstName: "random",
		LastName:  "user",
	}

	store.CreateUser(user, "")
	results, _ := store.GetResults(user.ID)
	results.Classes = []api.Class{
		api.Class{
			ID:    "randomid",
			Name:  "Random Class",
			Group: "21",
			Year:  "20142",
		},
	}

	go scheduler.Start()

	scheduler.Queue(user)

	scheduler.Stop()

	if len(results.Classes[0].Results) == 0 {
		t.Error("Results not updated")
	}

	if !messageSent {
		t.Error("Message not sent.")
	}

	end()
}

func TestSchedulerNoNewResults(t *testing.T) {
	scheduler, store := start()

	getResultsFunc = func() (res []RunResult) {
		// Fake data returned by the fake crawler.
		res = append(res, RunResult{
			ClassIndex: 0,
			Class: &api.Class{
				ID:    "randomid",
				Name:  "Random Class",
				Group: "21",
				Year:  "20142",
				Results: []api.Result{
					api.Result{
						Name:     "A result",
						Normal:   api.ResultInfo{},
						Weighted: api.ResultInfo{},
					},
				},
				Total: api.ResultInfo{},
				Final: "",
			},
			Err: nil,
		})
		return res
	}

	var messageSent bool
	sendFunc = func(to, subject, message string) {
		messageSent = true
	}

	user := &api.User{
		Email:     "random@user.com",
		FirstName: "random",
		LastName:  "user",
	}

	store.CreateUser(user, "")
	results, _ := store.GetResults(user.ID)
	results.Classes = []api.Class{
		api.Class{
			ID:    "randomid",
			Name:  "Random Class",
			Group: "21",
			Year:  "20142",
			Results: []api.Result{
				api.Result{
					Name:     "A result",
					Normal:   api.ResultInfo{},
					Weighted: api.ResultInfo{},
				},
			},
		},
	}

	go scheduler.Start()

	scheduler.Queue(user)

	scheduler.Stop()

	if len(results.Classes[0].Results) != 1 {
		t.Error("Results not updated")
	}

	if messageSent {
		t.Error("Sent an email when it was not supposed to.")
	}

	end()
}

func TestSchedulerLoad(t *testing.T) {
	scheduler, store := start()
	wg := sync.WaitGroup{}
	getResultsFunc = func() (res []RunResult) {
		waitTime := 100 + rand.Intn(200)
		time.Sleep(time.Duration(waitTime) * time.Millisecond)
		// Fake data returned by the fake crawler.
		res = append(res, RunResult{
			ClassIndex: 0,
			Class: &api.Class{
				ID:    "randomid",
				Name:  "Random Class",
				Group: "21",
				Year:  "20142",
				Results: []api.Result{
					api.Result{
						Name:     "A result",
						Normal:   api.ResultInfo{},
						Weighted: api.ResultInfo{},
					},
				},
				Total: api.ResultInfo{},
				Final: "",
			},
			Err: nil,
		})
		wg.Done()
		return res
	}

	user := &api.User{
		Email:     "random@user.com",
		FirstName: "random",
		LastName:  "user",
	}

	store.CreateUser(user, "")
	results, _ := store.GetResults(user.ID)
	results.Classes = []api.Class{
		api.Class{
			ID:    "randomid",
			Name:  "Random Class",
			Group: "21",
			Year:  "20142",
			Results: []api.Result{
				api.Result{
					Name:     "A result",
					Normal:   api.ResultInfo{},
					Weighted: api.ResultInfo{},
				},
			},
		},
	}

	go scheduler.Start()
	wg.Add(100)
	for i := 0; i < 100; i++ {
		scheduler.Queue(user)
	}

	wg.Wait()

	scheduler.Stop()

	end()
}

func start() (*Scheduler, *FakeStore) {
	config := new(SchedulerConfig)
	for i := 0; i < 10; i++ {
		config.ResultGetters = append(config.ResultGetters, &FakeCrawler{})
	}

	store := new(FakeStore)
	store.Data = make(map[string]*TestUser)

	config.UserStore = store
	config.UserResultsStore = store
	config.CrawlerConfigStore = store
	config.Sender = new(FakeSender)

	return NewScheduler(config), store
}

func end() {
	getResultsFunc = nil
	sendFunc = nil
}
