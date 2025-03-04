package main

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	DEPOSIT_INITIATED string = "deposit_initiated"
	DEPOSIT_COMPLETED string = "deposit_completed"
)

// ensureTopics ensures the specified Kafka topics exist, creating them if necessary.
func ensureTopics(topics []string) error {
	// Connect to Kafka leader to manage topics
	conn, err := kafka.Dial("tcp", "kafka:9092")
	if err != nil {
		return err
	}
	defer conn.Close()

	// List existing topics
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return err
	}

	// Build a set of existing topics
	existingTopics := make(map[string]struct{})
	for _, p := range partitions {
		existingTopics[p.Topic] = struct{}{}
	}

	// Define topic configurations
	topicConfigs := make([]kafka.TopicConfig, 0, len(topics))
	for _, topic := range topics {
		if _, exists := existingTopics[topic]; !exists {
			log.Printf("Topic %s does not exist, creating it...", topic)
			topicConfigs = append(topicConfigs, kafka.TopicConfig{
				Topic:             topic,
				NumPartitions:     3,
				ReplicationFactor: 1,
			})
		}
	}

	// Create missing topics
	if len(topicConfigs) > 0 {
		err = conn.CreateTopics(topicConfigs...)
		if err != nil {
			return err
		}
		log.Printf("Created topics: %v", topics)
	} else {
		log.Printf("All topics already exist: %v", topics)
	}

	return nil
}
