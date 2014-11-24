package lib

import "labix.org/v2/mgo/bson"

type (
	// UserInfoStore handles userinfo operations with the datastore.
	UserInfoStore interface {
		FindByID(userID bson.ObjectId) (*UserInfo, error)
		Update(userInfo *UserInfo) error
	}

	// UserInfo contains additional info about the user.
	UserInfo struct {
		UserID bson.ObjectId `bson:"user_id"`
		Email  string        `bson:"email"`
		Code   []byte        `bson:"code"`
		Nip    []byte        `bson:"nip"`
	}
)
