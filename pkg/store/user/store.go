package user

import "github.com/janicduplessis/resultscrawler/pkg/api"

// Store handles user related operations in the datastore.
type Store interface {
	GetUser(id string) (*api.User, error)
	GetUserForLogin(email string) (*api.User, string, error)
	ListUsers() ([]*api.User, error)
	UpdateUser(user *api.User) error
	CreateUser(user *api.User, password string) error
}
