package lib

import (
	"fmt"
	"log"
	"time"

	"labix.org/v2/mgo"

	"github.com/janicduplessis/resultscrawler/config"
)

type MongoStore struct {
	mongoSession *mgo.Session
}

type ConnCloser interface {
	Close()
}

func NewMongoStore() *MongoStore {
	// We need this object to establish a session to our MongoDB.
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{fmt.Sprintf("%s:%s", config.Config.DbURL, config.Config.DbPort)},
		Timeout:  60 * time.Second,
		Database: config.Config.DbName,
		Username: config.Config.DbUser,
		Password: config.Config.DbPassword,
	}

	// Create a session which maintains a pool of socket connections
	// to our MongoDB.
	mongoSession, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		log.Fatalf("CreateSession: %s\n", err)
	}

	// Reads may not be entirely up-to-date, but they will always see the
	// history of changes moving forward, the data read will be consistent
	// across sequential queries in the same session, and modifications made
	// within the session will be observed in following queries (read-your-writes).
	// http://godoc.org/labix.org/v2/mgo#Session.SetMode
	mongoSession.SetMode(mgo.Monotonic, true)

	return &MongoStore{
		mongoSession: mongoSession,
	}
}

func (hndl *MongoStore) Get() (*mgo.Database, ConnCloser) {
	sessionCopy := hndl.mongoSession.Copy()
	return sessionCopy.DB(config.Config.DbName), sessionCopy
}
