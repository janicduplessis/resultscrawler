package tools

// Sender provides an interface for sending messages.
type Sender interface {
	Send(to, subject, message string) error
}
