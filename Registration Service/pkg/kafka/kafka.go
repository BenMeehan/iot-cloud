package kafka

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

type KafkaClient struct {
	Consumer *kafka.Consumer
	Logger   *logrus.Logger
}

// NewKafkaClient creates a new Kafka consumer
func NewKafkaClient(securityProtocol, CACert, cert, key, mechanism, username, password string, brokers []string, groupID string, logger *logrus.Logger) (*KafkaClient, error) {
	// Kafka consumer configuration
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": brokers[0],
		"group.id":          groupID,
		"auto.offset.reset": "earliest", // Start reading at the earliest message
		// Uncomment the following lines to enable SSL and SASL
		// "security.protocol": securityProtocol,
		// "ssl.ca.location":          CACert,
		// "ssl.certificate.location": cert,
		// "ssl.key.location":         key,
		// "sasl.mechanism":           mechanism,
		// "sasl.username":            username,
		// "sasl.password":            password,
	}

	// Create a new consumer
	consumer, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		return nil, err
	}

	logger.Info("Kafka consumer created successfully")

	return &KafkaClient{Consumer: consumer, Logger: logger}, nil
}

// Subscribe subscribes to a Kafka topic and starts polling messages with a handler function
func (k *KafkaClient) Subscribe(topic string, handler func(*kafka.Message)) error {
	err := k.Consumer.Subscribe(topic, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}
	k.Logger.Infof("Subscribed to topic: %s", topic)

	// Start a goroutine to poll messages and invoke the handler
	go func() {
		for {
			message, err := k.Consumer.ReadMessage(-1) // -1 means blocking until a message is available
			if err == nil {
				handler(message)
			} else {
				k.Logger.Errorf("Error while receiving message: %v", err)
			}
		}
	}()
	return nil
}

// Close cleans up the Kafka consumer
func (k *KafkaClient) Close() {
	k.Consumer.Close()
	k.Logger.Info("Kafka consumer closed")
}
