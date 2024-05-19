package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	c "github.com/hrist-todorova/notification-system/distributor/consumer"
	"github.com/redis/go-redis/v9"
)

const RETRY_KAFKA_CONNECTION_SECONDS = 5

func main() {
	bootstrapServer := os.Getenv("BOOTSTRAP_SERVER")
	if bootstrapServer == "" {
		log.Fatal("Please set BOOTSTRAP_SERVER")
	}
	redisAddress := os.Getenv("REDIS_ADDRESS")
	if redisAddress == "" {
		log.Fatal("Please set REDIS_ADDRESS")
	}
	topic := os.Getenv("TOPIC")
	if topic == "" {
		log.Fatal("Please set TOPIC")
	}
	consumerGroupId := os.Getenv("CONSUMER_GROUP_ID")
	if consumerGroupId == "" {
		log.Fatal("Please set CONSUMER_GROUP_ID")
	}
	failureWait := os.Getenv("FAILURE_WAIT")
	if failureWait == "" {
		log.Fatal("Please set FAILURE_WAIT")
	}

	waitSeconds, err := strconv.Atoi(failureWait)
	if err != nil {
		log.Fatal("Please set FAILURE_WAIT to an integer value")
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V3_6_0_0
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = false

	var consumerGroup sarama.ConsumerGroup

	connectedToKafka := false
	for !connectedToKafka {
		log.Printf("Trying to connect to Kafka on %s\n", bootstrapServer)
		consumerGroup, err = sarama.NewConsumerGroup([]string{bootstrapServer}, consumerGroupId, saramaConfig)
		if err != nil {
			log.Printf("Failed to create consumer group: %v", err)
			time.Sleep(RETRY_KAFKA_CONNECTION_SECONDS * time.Second)
			continue
		}
		connectedToKafka = true
	}
	defer consumerGroup.Close()
	log.Println("Connected to Kafka")

	consumer := &c.EventConsumer{
		Topic: topic,
		KV: redis.NewClient(&redis.Options{
			Addr: redisAddress,
		}),
	}
	defer consumer.KV.Close()

	switch consumer.Topic {
	case "email":
		consumer.Notifier = &c.Email{
			User:      os.Getenv("USER"),
			Password:  os.Getenv("PASSWORD"),
			FromEmail: os.Getenv("FROM_EMAIL"),
			ToEmail:   os.Getenv("TO_EMAIL"),
			SmtpHost:  os.Getenv("SMTP_HOST"),
			SmtpPort:  os.Getenv("SMTP_PORT"),
		}
	case "sms":
		consumer.Notifier = &c.SMS{
			AccountSid: os.Getenv("ACCOUNT_SID"),
			AuthToken:  os.Getenv("AUTH_TOKEN"),
			FromPhone:  os.Getenv("FROM_PHONE"),
			ToPhone:    os.Getenv("TO_PHONE"),
		}
	case "slack":
		consumer.Notifier = &c.Slack{
			WebhookURL: os.Getenv("WEBHOOK_URL"),
		}
	default:
		log.Fatalf("Unknown topic: %s", consumer.Topic)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Listening for events ...")
	for {
		err = consumerGroup.Consume(ctx, []string{consumer.Topic}, consumer)
		if err != nil {
			log.Printf("Encountered an error while consuming events: %v\n", err)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
		}
	}
}
