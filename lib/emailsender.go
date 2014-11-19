package lib

import (
	"fmt"
	"net/smtp"
)

// EmailSender is a helper for sending mails using smtp
type EmailSender struct {
	config *EmailConfig
}

// EmailConfig contains email server configuration.
type EmailConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

// NewEmailSender creates a new email sender object.
func NewEmailSender(config *EmailConfig) *EmailSender {
	return &EmailSender{
		config,
	}
}

// Send sends a message to the specified email.
func (sender *EmailSender) Send(to, subject, message string) error {
	body := fmt.Sprintf("From: \"Results\" <noreply@resultcrawler.com>\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s", to, subject, message)
	auth := smtp.PlainAuth("", sender.config.User, sender.config.Password, sender.config.Host)
	err := smtp.SendMail(fmt.Sprintf("%s:%s", sender.config.Host, sender.config.Port),
		auth, "\"Results\" <noreply@resultcrawler.com>", []string{to}, []byte(body))

	return err
}
