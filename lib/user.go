package lib

import "labix.org/v2/mgo/bson"

type (
	// UserStore handles user related operations in the datastore.
	UserStore interface {
		FindByID(id bson.ObjectId) (*User, error)
		FindAll() ([]*User, error)
		Update(user *User) error
		Insert(user *User) error
	}

	// The User entity.
	User struct {
		ID           bson.ObjectId `bson:"_id,omitempty"`
		UserName     string        `bson:"user_name"`
		PasswordHash string        `bson:"password_hash"`
	}
)
