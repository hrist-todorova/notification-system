package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hrist-todorova/notification-system/api/producer"
)

const RETRY_KAFKA_CONNECTION_SECONDS = 5

func main() {
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		log.Fatal("Please set API_PORT")
	}
	bootstrapServer := os.Getenv("BOOTSTRAP_SERVER")
	if bootstrapServer == "" {
		log.Fatal("Please set BOOTSTRAP_SERVER")
	}

	var eventProducer *producer.EventProducer
	var err error

	connectedToKafka := false
	for !connectedToKafka {
		log.Printf("Trying to connect to Kafka on %s\n", bootstrapServer)
		eventProducer, err = producer.NewEventProducer(bootstrapServer)
		if err != nil {
			log.Printf("Failed to initiate a producer: %v", err)
			time.Sleep(RETRY_KAFKA_CONNECTION_SECONDS * time.Second)
			continue
		}
		connectedToKafka = true
	}

	defer eventProducer.Close()
	log.Println("Connected to Kafka")

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/notification", eventProducer.AcceptNotification)
	log.Printf("Starting API service on port %s\n", apiPort)
	http.ListenAndServe(fmt.Sprintf(":%s", apiPort), mux)
}
