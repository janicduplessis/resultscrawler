package lib

import "labix.org/v2/mgo"

// Store is an interface to a nosql datastore.
// TODO: it is incomplete since it returns an object from the mgo package.
type Store interface {
	Get() (*mgo.Database, ConnCloser)
}
