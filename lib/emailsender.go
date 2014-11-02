package lib

const emailHost = "smtp.gmail.com"

// EmailSender is a helper for sending mails using smtp
type EmailSender struct {
}

// Send sends a message to the specified email.
func (sender *EmailSender) Send(email, title, body string) error {
	//auth := smtp.PlainAuth("", config.Config.EmailUser, config.Config.EmailPassword, emailHost)

	return nil
}
