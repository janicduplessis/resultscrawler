package crawler

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/janicduplessis/resultscrawler/pkg/api"
	"github.com/janicduplessis/resultscrawler/pkg/store/fakestore"
)

type FakeCrawler struct{}

func (c *FakeCrawler) Run(user *User) []RunResult {
	if getResultsFunc != nil {
		return getResultsFunc()
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
		scheduler.QueueAsync(user, nil)
	}

	wg.Wait()

	scheduler.Stop()

	end()
}

func start() (*Scheduler, *fakestore.FakeStore) {
	config := new(SchedulerConfig)
	for i := 0; i < 10; i++ {
		config.ResultGetters = append(config.ResultGetters, &FakeCrawler{})
	}

	store := new(fakestore.FakeStore)
	store.Data = make(map[string]*fakestore.TestUser)

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

func init() {
	// Working directory is different in test so we have to fix the path of
	// the template file.
	msgTemplatePath = "../../crawler/msgtemplate.html"
}
