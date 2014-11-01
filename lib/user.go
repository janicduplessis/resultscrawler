package lib

type UserStore interface {
	FindById(id int64) error
	Store(user *User) error
}

type User struct {
	ID       int64
	UserName string
	Code     string
	Nip      string
	Classes  []Class
}
