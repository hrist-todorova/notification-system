package consumer

import (
	"github.com/redis/go-redis/v9"
)

type EventConsumer struct {
	Topic    string
	KV       *redis.Client
	Notifier Notifier
}

type Notifier interface {
	Notify(message string) (err error)
}

type Email struct {
	User      string
	Password  string
	FromEmail string
	ToEmail   string
	SmtpHost  string
	SmtpPort  string
}

type SMS struct {
	AccountSid string
	AuthToken  string
	FromPhone  string
	ToPhone    string
}

type Slack struct {
	WebhookURL string
}
