package lib

import "labix.org/v2/mgo/bson"

// UserStore handles user related operations in the datastore.
type UserStore interface {
	FindByID(id bson.ObjectId) (*User, error)
	FindAll() ([]*User, error)
	Update(user *User) error
	Insert(user *User) error
}

// The User entity.
type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	UserName string        `bson:"user_name"`
	Email    string        `bson:"email"`
	Code     []byte        `bson:"code"`
	Nip      []byte        `bson:"nip"`
	Classes  []Class       `bson:"classes"`
}
