package mongo

import (
	"errors"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/pkg/api"
	"github.com/janicduplessis/resultscrawler/pkg/crypto"
	"github.com/janicduplessis/resultscrawler/pkg/tools"
)

// Store implements the interfaces for storing users, crawlerconfigs and results
// in a mongo data store.
type Store struct {
	helper *tools.MongoHelper
}

const (
	userKey = "user"
)

// New returns a new mongo store.
func New(helper *tools.MongoHelper) *Store {
	return &Store{
		helper,
	}
}

// GetCrawlerConfig returns the crawler config for the specified user.
func (s *Store) GetCrawlerConfig(userID string) (*api.CrawlerConfig, error) {
	db, conn := s.helper.Client()
	defer conn.Close()

	user := mongoUser{}
	err := db.C(userKey).
		FindId(bson.ObjectIdHex(userID)).
		Select(bson.M{"crawler_config": 1}).
		One(&user)
	if err != nil {
		return nil, err
	}

	err = decryptCrawlerConfig(user.CrawlerConfig)
	if err != nil {
		return nil, err
	}

	return user.CrawlerConfig, err
}

// UpdateCrawlerConfig updates the crawler config with the specified config.
func (s *Store) UpdateCrawlerConfig(crawlerConfig *api.CrawlerConfig) error {
	err := encryptCrawlerConfig(crawlerConfig)
	if err != nil {
		return err
	}

	db, conn := s.helper.Client()
	defer conn.Close()
	return db.C(userKey).UpdateId(bson.ObjectIdHex(crawlerConfig.UserID), bson.M{"$set": bson.M{"crawler_config": crawlerConfig}})
}

// GetResults returns results for a user.
func (s *Store) GetResults(userID string) (*api.Results, error) {
	db, conn := s.helper.Client()
	defer conn.Close()
	user := mongoUser{}
	err := db.C(userKey).
		FindId(bson.ObjectIdHex(userID)).
		Select(bson.M{"results.": 1}).
		One(&user)
	return user.Results, err
}

// UpdateResults updates results for a user.
func (s *Store) UpdateResults(userResults *api.Results) error {
	db, conn := s.helper.Client()
	defer conn.Close()

	return db.C(userKey).UpdateId(bson.ObjectIdHex(userResults.UserID), bson.M{"$set": bson.M{"results": userResults}})
}

// GetUser returns a user with the specified id.
func (s *Store) GetUser(id string) (*api.User, error) {
	db, conn := s.helper.Client()
	defer conn.Close()

	user := &mongoUser{}
	err := db.C(userKey).FindId(bson.ObjectIdHex(id)).Select(bson.M{"user": 1}).One(&user)
	return user.User, err
}

// GetUserForLogin return a user by email with a password hash.
func (s *Store) GetUserForLogin(email string) (*api.User, string, error) {
	db, conn := s.helper.Client()
	defer conn.Close()

	user := mongoUser{}
	err := db.C(userKey).
		Find(bson.M{"user.email": email}).
		Select(bson.M{"user": 1, "password_hash": 1}).
		One(&user)

	if err == mgo.ErrNotFound {
		return nil, "", nil
	}

	return user.User, user.PasswordHash, err
}

// ListUsers returns all users.
func (s *Store) ListUsers() ([]*api.User, error) {
	db, conn := s.helper.Client()
	defer conn.Close()

	mongoUsers := []mongoUser{}
	err := db.C(userKey).Find(nil).Select(bson.M{"_id": 0, "user": 1}).All(&mongoUsers)
	users := make([]*api.User, len(mongoUsers))
	for i, u := range mongoUsers {
		users[i] = u.User
	}
	return users, err
}

// UpdateUser updates a user.
func (s *Store) UpdateUser(user *api.User) error {
	db, conn := s.helper.Client()
	defer conn.Close()

	return db.C(userKey).UpdateId(bson.ObjectIdHex(user.ID), bson.M{"$set": bson.M{"user": user}})
}

// CreateUser adds a new user.
func (s *Store) CreateUser(user *api.User, password string) error {
	db, conn := s.helper.Client()
	defer conn.Close()

	id := bson.NewObjectId()
	hexID := id.Hex()
	user.ID = hexID
	crawlerConfig := &api.CrawlerConfig{UserID: hexID, NotificationEmail: user.Email}
	results := &api.Results{UserID: hexID}
	mongoUser := mongoUser{
		id,
		user,
		crawlerConfig,
		results,
		password,
	}
	return db.C(userKey).Insert(&mongoUser)
}

func encryptCrawlerConfig(crawlerConfig *api.CrawlerConfig) error {
	// Encrypt code and nip before saving.
	userCode, err := crypto.AESEncrypt([]byte(crawlerConfig.Code))
	if err != nil {
		return err
	}
	crawlerConfig.Code = string(userCode)
	userNip, err := crypto.AESEncrypt([]byte(crawlerConfig.Nip))
	if err != nil {
		return err
	}
	crawlerConfig.Nip = string(userNip)
	return nil
}

func decryptCrawlerConfig(crawlerConfig *api.CrawlerConfig) error {
	// Encrypt code and nip before saving.
	userCode, err := crypto.AESDecrypt([]byte(crawlerConfig.Code))
	if err != nil {
		return err
	}
	crawlerConfig.Code = string(userCode)
	userNip, err := crypto.AESDecrypt([]byte(crawlerConfig.Nip))
	if err != nil {
		return err
	}
	crawlerConfig.Nip = string(userNip)
	return nil
}

func toOID(id string) (bson.ObjectId, error) {
	if !bson.IsObjectIdHex(id) {
		return bson.ObjectId(""), errors.New("Invalid object ID")
	}

	return bson.ObjectIdHex(id), nil
}
