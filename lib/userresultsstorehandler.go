package lib

import "labix.org/v2/mgo/bson"

const userResultsKey = "user_results"

// UserResultsStoreHandler implements the UserResultsStore interface.
type UserResultsStoreHandler struct {
	store Store
}

// NewUserResultsStoreHandler creates a new UserResultsStoreHandler.
func NewUserResultsStoreHandler(store Store) *UserResultsStoreHandler {
	return &UserResultsStoreHandler{
		store: store,
	}
}

func (hndl *UserResultsStoreHandler) FindByID(userID bson.ObjectId) (*UserResults, error) {
	db, conn := hndl.store.Get()
	defer conn.Close()

	userResults := UserResults{}
	err := db.C(userResultsKey).FindId(userID).One(&userResults)
	return &userResults, err
}

func (hndl *UserResultsStoreHandler) Update(userResults *UserResults) error {
	db, conn := hndl.store.Get()
	defer conn.Close()

	err := db.C(userResultsKey).UpdateId(userResults.UserID, userResults)
	return err
}
