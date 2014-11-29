package store

import "labix.org/v2/mgo/bson"

type (
	// UserStore handles user related operations in the datastore.
	UserStore interface {
		FindByID(id bson.ObjectId) (*User, error)
		FindByEmail(email string) (*User, error)
		FindAll() ([]*User, error)
		Update(user *User) error
		Insert(user *User) error
	}

	// The User entity.
	User struct {
		ID           bson.ObjectId `bson:"_id,omitempty"`
		Email        string        `bson:"email"`
		PasswordHash string        `bson:"password_hash"`
		FirstName    string        `bson:"first_name"`
		LastName     string        `bson:"last_name"`
	}
)
