package kafka

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

type KafkaClient struct {
	Producer *kafka.Producer
	Logger   *logrus.Logger
}

func NewKafkaClient(securityProtocol, CACert, cert, key, mechanism, username, password string, brokers []string, logger *logrus.Logger) (*KafkaClient, error) {
	// Kafka producer configuration
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": brokers[0],
		// "security.protocol": securityProtocol,
		// "ssl.ca.location":          CACert,
		// "ssl.certificate.location": cert,
		// "ssl.key.location":         key,
		// "sasl.mechanism":           mechanism,
		// "sasl.username":            username,
		// "sasl.password":            password,
	}

	// Create a new producer
	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		return nil, err
	}

	logger.Info("Kafka producer created successfully")

	return &KafkaClient{Producer: producer, Logger: logger}, nil
}

// PublishMessage publishes a message to a specified Kafka topic
func (k *KafkaClient) PublishMessage(topic string, key string, value []byte) error {
	// Create a new message
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          value,
	}

	// Produce the message asynchronously
	deliveryChan := make(chan kafka.Event)

	err := k.Producer.Produce(message, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Wait for delivery report
	go func() {
		e := <-deliveryChan
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			k.Logger.Errorf("Delivery failed for message: %v", m.TopicPartition.Error)
		} else {
			k.Logger.Infof("Message delivered to %v [%v] at offset %v\n", *m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
		}
	}()

	return nil
}

// Close cleans up the Kafka producer
func (k *KafkaClient) Close() {
	k.Producer.Flush(15 * 1000)
	k.Producer.Close()
}
