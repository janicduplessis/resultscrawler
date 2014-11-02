package lib

import "labix.org/v2/mgo"

type Store interface {
	Get() (*mgo.Database, ConnCloser)
}
