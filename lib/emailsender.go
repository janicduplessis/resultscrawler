package lib

const emailHost = "smtp.gmail.com"

type EmailSender struct {
}

func (sender *EmailSender) Send(email, title, body string) error {
	//auth := smtp.PlainAuth("", config.Config.EmailUser, config.Config.EmailPassword, emailHost)

	return nil
}
