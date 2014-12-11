package tools

import (
	"fmt"
	"log"
	"time"

	"labix.org/v2/mgo"
)

// MongoHelper handles connection to a mongodb database.
type MongoHelper struct {
	mongoSession *mgo.Session
	config       *MongoConfig
}

// MongoConfig contains the configuration of the database server.
type MongoConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// The ConnCloser interface provides an abstraction to close a db connection.
type ConnCloser interface {
	Close()
}

// NewMongoHelper creates a new MongoStore object.
func NewMongoHelper(dbConfig *MongoConfig) *MongoHelper {
	// We need this object to establish a session to our MongoDB.
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{fmt.Sprintf("%s:%s", dbConfig.Host, dbConfig.Port)},
		Timeout:  60 * time.Second,
		Database: dbConfig.Name,
		Username: dbConfig.User,
		Password: dbConfig.Password,
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

	return &MongoHelper{
		mongoSession: mongoSession,
		config:       dbConfig,
	}
}

// Client returns a connection from the connection pool using the database
// in the configuration. It is the caller's responsability to close the
// connection using the ConnCloser.
func (hndl *MongoHelper) Client() (*mgo.Database, ConnCloser) {
	sessionCopy := hndl.mongoSession.Copy()
	return sessionCopy.DB(hndl.config.Name), sessionCopy
}
