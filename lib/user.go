package lib

import "labix.org/v2/mgo/bson"

type UserStore interface {
	FindById(id bson.ObjectId) (*User, error)
	FindAll() ([]*User, error)
	Update(user *User) error
	Insert(user *User) error
}

type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	UserName string        `bson:"user_name"`
	Email    string        `bson:"email"`
	Code     string        `bson:"code"`
	Nip      string        `bson:"nip"`
	Classes  []Class       `bson:"classes"`
}
