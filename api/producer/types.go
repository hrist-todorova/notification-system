package producer

import "github.com/IBM/sarama"

type EventProducer struct {
	producer sarama.SyncProducer
}

type Notification struct {
	Channel string `json:"channel" validate:"required"`
	Message string `json:"message" validate:"required"`
}
