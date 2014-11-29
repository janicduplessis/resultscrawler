package utils

// Sender is an interface for sending a message
type Sender interface {
	Send(email, title, body string) error
}
