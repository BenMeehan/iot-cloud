package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/benmeehan/mqtt-kafka-connector-service/pkg/kafka"
	"github.com/benmeehan/mqtt-kafka-connector-service/pkg/mqtt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

// MqttKafkaConnector connects MQTT and Kafka services
type MqttKafkaConnector struct {
	mqttService  *mqtt.MqttService
	kafkaService *kafka.KafkaClient
	TopicMapping map[string]string
	Logger       *logrus.Logger
}

// NewMqttKafkaConnector creates a new instance of MqttKafkaConnector
func NewMqttKafkaConnector(mqttService *mqtt.MqttService, kafkaService *kafka.KafkaClient, topicMapping map[string]string, logger *logrus.Logger) *MqttKafkaConnector {
	return &MqttKafkaConnector{
		mqttService:  mqttService,
		kafkaService: kafkaService,
		TopicMapping: topicMapping,
		Logger:       logger,
	}
}

// Start begins subscribing to MQTT topics and publishing messages to Kafka
func (c *MqttKafkaConnector) Start() error {
	for mqttTopic, kafkaTopic := range c.TopicMapping {
		// Define a callback function for handling incoming MQTT messages
		mqttCallback := func(client MQTT.Client, msg MQTT.Message) {
			c.handleMqttMessage(msg, kafkaTopic)
		}

		// Subscribe to the MQTT topic
		token := c.mqttService.Subscribe(mqttTopic, 1, mqttCallback)
		if token.Wait() && token.Error() != nil {
			return fmt.Errorf("failed to subscribe to MQTT topic %s: %w", mqttTopic, token.Error())
		}

		c.Logger.Infof("Subscribed to MQTT topic: %s", mqttTopic)
	}

	return nil
}

// handleMqttMessage processes incoming MQTT messages and publishes them to Kafka
func (c *MqttKafkaConnector) handleMqttMessage(msg MQTT.Message, kafkaTopic string) {
	// Log the received MQTT message
	c.Logger.Infof("Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))

	// Prepare the message payload for Kafka
	message := map[string]interface{}{
		"payload":   string(msg.Payload()),
		"timestamp": time.Now().Unix(),
	}

	// Serialize the message to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		c.Logger.WithError(err).Error("Failed to marshal message to JSON")
		return
	}

	// Publish the message to Kafka
	if err := c.kafkaService.PublishMessage(kafkaTopic, msg.Topic(), messageJSON); err != nil {
		c.Logger.WithError(err).Error("Failed to publish message to Kafka")
	} else {
		c.Logger.Infof("Published message to Kafka topic: %s", kafkaTopic)
	}
}
