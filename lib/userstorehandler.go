package lib

import "labix.org/v2/mgo/bson"

const userKey = "user"

type UserStoreHandler struct {
	store Store
}

func NewUserStoreHandler(store Store) *UserStoreHandler {
	return &UserStoreHandler{
		store: store,
	}
}

func (hndl *UserStoreHandler) FindById(id bson.ObjectId) (*User, error) {
	db, conn := hndl.store.Get()
	defer conn.Close()

	user := User{}
	err := db.C(userKey).FindId(id).One(&user)
	return &user, err
}

func (hndl *UserStoreHandler) FindAll() ([]*User, error) {
	db, conn := hndl.store.Get()
	defer conn.Close()

	var users []*User
	err := db.C(userKey).Find(nil).All(&users)
	return users, err
}

func (hndl *UserStoreHandler) Update(user *User) error {
	db, conn := hndl.store.Get()
	defer conn.Close()

	err := db.C(userKey).UpdateId(user.ID, user)
	return err
}

func (hndl *UserStoreHandler) Insert(user *User) error {
	db, conn := hndl.store.Get()
	defer conn.Close()

	user.ID = bson.NewObjectId()

	err := db.C(userKey).Insert(user)
	return err
}
