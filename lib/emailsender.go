package lib

import (
	"fmt"
	"net/smtp"

	"github.com/janicduplessis/resultscrawler/config"
)

// EmailSender is a helper for sending mails using smtp
type EmailSender struct {
}

// Send sends a message to the specified email.
func (sender *EmailSender) Send(to, subject, message string) error {
	body := fmt.Sprintf("To: %s\r\nFrom:Results<noreply@resultcrawler.com>\r\nSubject: %s\r\n\r\n%s", to, subject, message)
	auth := smtp.PlainAuth("", config.Config.EmailUser, config.Config.EmailPassword, config.Config.EmailHost)
	err := smtp.SendMail(fmt.Sprintf("%s:%s", config.Config.EmailHost, config.Config.EmailPort),
		auth, "Results<noreply@resultcrawler.com>", []string{to}, []byte(body))

	return err
}
