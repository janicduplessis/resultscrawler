package lib

import (
	"errors"

	"labix.org/v2/mgo"
)

// ErrNotFound happens when no data is found in a datastore request.
var ErrNotFound = errors.New("Not found")

// Store is an interface to a nosql datastore.
// TODO: it is incomplete since it returns an object from the mgo package.
type Store interface {
	Get() (*mgo.Database, ConnCloser)
}
