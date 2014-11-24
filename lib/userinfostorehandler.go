package lib

import "labix.org/v2/mgo/bson"

const userInfoKey = "user_info"

// UserInfoStoreHandler implements the UserInfoStore interface.
type UserInfoStoreHandler struct {
	store Store
}

// NewUserInfoStoreHandler creates a new UserInfoStoreHandler.
func NewUserInfoStoreHandler(store Store) *UserInfoStoreHandler {
	return &UserInfoStoreHandler{
		store: store,
	}
}

func (hndl *UserInfoStoreHandler) FindByID(userID bson.ObjectId) (*UserInfo, error) {
	db, conn := hndl.store.Get()
	defer conn.Close()

	userInfo := UserInfo{}
	err := db.C(userInfoKey).FindId(userID).One(&userInfo)
	return &userInfo, err
}

func (hndl *UserInfoStoreHandler) Update(userInfo *UserInfo) error {
	db, conn := hndl.store.Get()
	defer conn.Close()

	err := db.C(userInfoKey).UpdateId(userInfo.UserID, userInfo)
	return err
}
