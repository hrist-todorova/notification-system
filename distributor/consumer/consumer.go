package consumer

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/IBM/sarama"
)

func (e EventConsumer) Setup(s sarama.ConsumerGroupSession) (err error) {
	var keys []string
	keys, _, err = e.KV.Scan(s.Context(), 0, fmt.Sprintf("%s:*", e.Topic), 0).Result()
	if err != nil {
		return fmt.Errorf("failed to find previous offsets in the database: %w", err)
	}
	if len(keys) == 0 {
		return
	}

	result, err := e.KV.MGet(s.Context(), keys...).Result()
	if err != nil {
		return fmt.Errorf("failed to fetch previous offsets from the database: %w", err)
	}

	for i, key := range keys {
		keyParts := strings.Split(key, ":")

		partitionInt, err := strconv.ParseInt(keyParts[1], 10, 32)
		if err != nil {
			return fmt.Errorf("failed to convert %s to int32: %w", keyParts[1], err)
		}

		offsetInt, err := strconv.ParseInt(result[i].(string), 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to int64: %w", result[i].(string), err)
		}

		topic := keyParts[0]
		partition := int32(partitionInt)
		nextOffset := int64(offsetInt) + 1

		s.MarkOffset(topic, partition, nextOffset, "")
		log.Printf("Next expected offset for topic: %s, partition: %d is %d\n", topic, partition, nextOffset)
	}

	return
}

func (e EventConsumer) Cleanup(_ sarama.ConsumerGroupSession) (err error) {
	return
}

func (e EventConsumer) ConsumeClaim(s sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) (err error) {
	for message := range claim.Messages() {
		log.Printf("Received message from topic: %s, partition: %d, offset: %d\n",
			message.Topic, message.Partition, message.Offset)

		err = e.Notifier.Notify(string(message.Value))
		if err != nil {
			err = fmt.Errorf("failed to process message from topic: %s, partition: %d, offset: %d: %s",
				message.Topic, message.Partition, message.Offset, err)
			log.Println(err)
			return
		}

		s.MarkMessage(message, "")
		log.Printf("Processed message from topic: %s, partition: %d, offset: %d\n",
			message.Topic, message.Partition, message.Offset)

		key := fmt.Sprintf("%s:%d", message.Topic, message.Partition)

		err = e.KV.Set(s.Context(), key, message.Offset, 0).Err()
		if err != nil {
			err = fmt.Errorf("failed to save offset %d for topic %s with partition %d in the database: %w",
				message.Offset, message.Topic, message.Partition, err)
			log.Println(err)
			return
		}
	}
	return
}
