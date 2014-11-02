package lib

type Sender interface {
	Send(email, title, body string) error
}
