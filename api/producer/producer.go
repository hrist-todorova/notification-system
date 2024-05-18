package producer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/IBM/sarama"
)

func NewEventProducer(bootstrapServer string) (e *EventProducer, err error) {
	config := sarama.NewConfig()
	config.Metadata.AllowAutoTopicCreation = true
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true

	e = &EventProducer{}
	e.producer, err = sarama.NewSyncProducer([]string{bootstrapServer}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create a producer: %w", err)
	}
	return
}

func (e *EventProducer) Close() {
	e.producer.Close()
}

func (e *EventProducer) AcceptNotification(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("Failed to read body: %v\n", err)
		return
	}
	log.Printf("Received: %s\n", string(body))

	var n Notification
	err = json.Unmarshal(body, &n)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("Failed to unmarshall body: %v\n", err)
		return
	}

	message := &sarama.ProducerMessage{
		Topic: n.Channel,
		Value: sarama.StringEncoder(n.Message),
	}
	_, _, err = e.producer.SendMessage(message)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("Failed to accept notification: %v\n", err)
		return
	}
}
