package tools

import "github.com/anachronistic/apns"

const (
	apnsURL    = "gateway.push.apple.com:2195"
	sandboxURL = "gateway.sandbox.push.apple.com:2195"
)

type APNSSender struct {
	client *apns.Client
}

func NewAPNSSender(cert string, key string, sandbox bool) *APNSSender {
	url := apnsURL
	if sandbox {
		url = sandboxURL
	}
	client := apns.NewClient(url, cert, key)
	return &APNSSender{client}
}

func (s *APNSSender) Send(to, subject, message string) error {
	payload := apns.NewPayload()
	payload.Alert = message

	pn := apns.NewPushNotification()
	pn.DeviceToken = to
	pn.AddPayload(payload)

	resp := s.client.Send(pn)
	if resp.Error != nil {
		return resp.Error
	}

	return nil
}
