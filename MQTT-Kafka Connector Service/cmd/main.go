package main

import (
	"os"

	"github.com/benmeehan/mqtt-kafka-connector-service/internal/services"
	"github.com/benmeehan/mqtt-kafka-connector-service/internal/utils"
	"github.com/benmeehan/mqtt-kafka-connector-service/pkg/kafka"
	"github.com/benmeehan/mqtt-kafka-connector-service/pkg/mqtt"
	"github.com/google/uuid"

	"github.com/sirupsen/logrus"
)

func main() {
	// Set up structured logging with JSON output
	var log = &logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}

	// Load configuration from file
	config, err := utils.LoadConfig("config/config.yaml", log)
	if err != nil {
		log.WithError(err).Fatal("Failed to load configuration")
	}

	// Generate a unique MQTT Client ID by appending a UUID
	config.MQTT.ClientID = config.MQTT.ClientID + "-" + uuid.New().String()
	log.Infof("Using MQTT Client ID: %s", config.MQTT.ClientID)

	// Initialize the shared MQTT connection
	mqttClient := mqtt.NewMqttService(log)
	err = mqttClient.Initialize(config.MQTT.Broker, config.MQTT.ClientID, config.MQTT.TLS.CACert)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize MQTT connection")
	}

	MQTTtoKafkaTopicMappings := make(map[string]string)
	for _, t := range config.TopicMappings {
		MQTTtoKafkaTopicMappings[t.MQTTTopic] = t.KafkaTopic
	}

	kafkaClient, err := kafka.NewKafkaClient(
		config.Kafka.SecurityProtocol,
		config.Kafka.SSL.CACert,
		config.Kafka.SSL.Cert,
		config.Kafka.SSL.Key,
		config.Kafka.SASL.Mechanism,
		config.Kafka.SASL.Username,
		config.Kafka.SASL.Password,
		config.Kafka.Brokers,
		log,
	)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize Kafka Client")
	}

	connector := services.NewMqttKafkaConnector(mqttClient, kafkaClient, MQTTtoKafkaTopicMappings, log)
	connector.Start()

	// Block the main thread to keep services running
	log.Info("MQTT to Kafka Source Connector service is running...")
	select {}
}
