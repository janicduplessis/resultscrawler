package lib

import "labix.org/v2/mgo/bson"

const userKey = "user"

// UserStoreHandler implements the UserStore interface.
type UserStoreHandler struct {
	store Store
}

// NewUserStoreHandler creates a new UserStoreHandler.
func NewUserStoreHandler(store Store) *UserStoreHandler {
	return &UserStoreHandler{
		store: store,
	}
}

// FindByID returns a user with the specified id.
func (hndl *UserStoreHandler) FindByID(id bson.ObjectId) (*User, error) {
	db, conn := hndl.store.Get()
	defer conn.Close()

	user := User{}
	err := db.C(userKey).FindId(id).One(&user)
	return &user, err
}

// FindByEmail return a user with the specified email.
func (hndl *UserStoreHandler) FindByEmail(email string) (*User, error) {
	db, conn := hndl.store.Get()
	defer conn.Close()

	user := User{}
	err := db.C(userKey).Find(bson.M{"email": email}).One(&user)
	return &user, err
}

// FindAll returns all users.
func (hndl *UserStoreHandler) FindAll() ([]*User, error) {
	db, conn := hndl.store.Get()
	defer conn.Close()

	var users []*User
	err := db.C(userKey).Find(nil).All(&users)
	return users, err
}

// Update updates a user.
func (hndl *UserStoreHandler) Update(user *User) error {
	db, conn := hndl.store.Get()
	defer conn.Close()

	err := db.C(userKey).UpdateId(user.ID, user)
	return err
}

// Insert adds a user.
func (hndl *UserStoreHandler) Insert(user *User) error {
	db, conn := hndl.store.Get()
	defer conn.Close()

	user.ID = bson.NewObjectId()

	err := db.C(userKey).Insert(user)
	return err
}
