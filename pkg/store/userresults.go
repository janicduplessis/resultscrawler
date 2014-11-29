package store

import "labix.org/v2/mgo/bson"

type (
	UserResultsStore interface {
		FindByID(userID bson.ObjectId) (*UserResults, error)
		Update(results *UserResults) error
		Insert(results *UserResults) error
	}

	UserResults struct {
		UserID  bson.ObjectId `bson:"user_id"`
		Classes []Class       `bson:"classes"`
	}

	// Class is an entity for a class.
	Class struct {
		ID      bson.ObjectId `bson:"_id"`
		Name    string        `bson:"name"`
		Group   string        `bson:"group"`
		Year    string        `bson:"year"`
		Results []Result      `bson:"results"`
	}

	// Result is an entity for storing a result
	Result struct {
		Name    string `bson:"name"`
		Result  string `bson:"result"`
		Average string `bson:"average"`
	}
)
