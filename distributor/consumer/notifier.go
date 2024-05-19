package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

func (e *Email) Notify(message string) (err error) {
	text := []byte("" +
		"From: SumUp Task <" + e.FromEmail + ">\r\n" +
		"To: " + e.ToEmail + " <" + e.ToEmail + ">\r\n" +
		"Subject: New notification\r\n" +
		"\r\n" +
		message +
		"\r\n")

	auth := smtp.PlainAuth("", e.User, e.Password, e.SmtpHost)
	err = smtp.SendMail(fmt.Sprintf("%s:%s", e.SmtpHost, e.SmtpPort), auth, e.FromEmail, []string{e.ToEmail}, text)
	if err != nil {
		return fmt.Errorf("failed to send email message: %w", err)
	}
	return
}

func (s *SMS) Notify(message string) (err error) {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: s.AccountSid,
		Password: s.AuthToken,
	})

	params := &twilioApi.CreateMessageParams{}
	params.SetFrom(s.FromPhone)
	params.SetTo(s.ToPhone)
	params.SetBody(message)

	_, err = client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("failed to send SMS message: %w", err)
	}

	return
}

func (s *Slack) Notify(message string) (err error) {
	jsonMessage, err := json.Marshal(map[string]string{"text": message})
	if err != nil {
		return fmt.Errorf("failed to marshall json: %w", err)
	}

	resp, err := http.Post(s.WebhookURL, "application/json", bytes.NewBuffer(jsonMessage))
	if err != nil {
		return fmt.Errorf("failed to send Slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned error code: %d", resp.StatusCode)
	}

	return
}
