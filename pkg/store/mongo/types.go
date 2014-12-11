package mongo

import (
	"labix.org/v2/mgo/bson"

	"github.com/janicduplessis/resultscrawler/pkg/api"
)

type (
	mongoUser struct {
		ID            bson.ObjectId      `bson:"_id,omitempty"`
		User          *api.User          `bson:"user"`
		CrawlerConfig *api.CrawlerConfig `bson:"crawler_config"`
		Results       *api.Results       `bson:"results"`
		PasswordHash  string             `bson:"password_hash"`
	}
)
