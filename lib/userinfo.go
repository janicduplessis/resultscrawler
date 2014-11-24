package lib

import "labix.org/v2/mgo/bson"

type (
	// UserInfoStore handles userinfo operations with the datastore.
	UserInfoStore interface {
		FindByID(userID bson.ObjectId) (*UserInfo, error)
		Update(userInfo *UserInfo) error
		Insert(userInfo *UserInfo) error
	}

	// UserInfo contains additional info about the user.
	UserInfo struct {
		UserID    bson.ObjectId `bson:"user_id"`
		FirstName string        `bson:"first_name"`
		LastName  string        `bson:"last_name"`
		CrawlerOn bool          `bson:"crawler_on"`
		Code      []byte        `bson:"code"`
		Nip       []byte        `bson:"nip"`
	}
)
