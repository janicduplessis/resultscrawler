package store

import "labix.org/v2/mgo/bson"

type (
	// CrawlerConfigStore handles crawler configuration operations with the datastore.
	CrawlerConfigStore interface {
		FindByID(userID bson.ObjectId) (*CrawlerConfig, error)
		Update(crawlerConfig *CrawlerConfig) error
		Insert(crawlerConfig *CrawlerConfig) error
	}

	// CrawlerConfig contains info about the crawler configuration.
	CrawlerConfig struct {
		UserID            bson.ObjectId `bson:"user_id"`
		Status            bool          `bson:"status"`
		Code              []byte        `bson:"code"`
		Nip               []byte        `bson:"nip"`
		NotificationEmail string        `bson:"notification_email"`
	}
)
