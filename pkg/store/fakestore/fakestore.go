package fakestore

import (
	"sync"

	"github.com/janicduplessis/resultscrawler/pkg/api"
	"github.com/janicduplessis/resultscrawler/pkg/crypto"
	"labix.org/v2/mgo/bson"
)

type TestUser struct {
	User          *api.User
	CrawlerConfig *api.CrawlerConfig
	Results       *api.Results
	Password      string
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
	s.mut.RLock()
	defer s.mut.RUnlock()
	for id, user := range s.Data {
		if user.User.Email == email {
			return s.Data[id].User, s.Data[id].Password, nil
		}
	}
	return nil, "", nil
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
	passHash, err := crypto.GenerateFromPassword(password)
	if err != nil {
		return err
	}
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
		passHash,
	}
	return nil
}
