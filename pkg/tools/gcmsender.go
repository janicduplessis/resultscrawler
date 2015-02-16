package tools

import "github.com/alexjlockwood/gcm"

type GCMSender struct {
	sender *gcm.Sender
}

func NewGCMSender(apiKey string) *GCMSender {
	sender := &gcm.Sender{ApiKey: apiKey}
	return &GCMSender{sender}
}

func (s *GCMSender) Send(to, subject, message string) error {
	data := map[string]interface{}{"message": message}
	msg := gcm.NewMessage(data, to)
	_, err := s.sender.Send(msg, 2)
	if err != nil {
		return err
	}

	return nil
}
