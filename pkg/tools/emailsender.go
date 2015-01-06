package tools

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailSender is a helper for sending mails using smtp
type EmailSender struct {
	config *EmailConfig
}

// EmailConfig contains email server configuration.
type EmailConfig struct {
	URL      string
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
	auth := smtp.PlainAuth("", sender.config.User, sender.config.Password, sender.config.URL[:strings.Index(sender.config.URL, ":")])
	err := smtp.SendMail(sender.config.URL,
		auth, "\"Results\" <noreply@resultcrawler.com>", []string{to}, []byte(body))

	return err
}
