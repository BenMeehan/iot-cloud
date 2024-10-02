package main

import (
	"fmt"
	"os"

	"github.com/benmeehan/iot-registration-service/internal/constants"
	"github.com/benmeehan/iot-registration-service/internal/database"
	"github.com/benmeehan/iot-registration-service/internal/services"
	"github.com/benmeehan/iot-registration-service/internal/utils"
	"github.com/benmeehan/iot-registration-service/pkg/file"
	"github.com/benmeehan/iot-registration-service/pkg/kafka"
	"github.com/benmeehan/iot-registration-service/pkg/mqtt"
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

	// Initialize the database connection
	dBClient := database.NewDatabase(log)
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.DB.Host, config.DB.Port, config.DB.User, config.DB.Password, config.DB.Name, config.DB.SSLMode)

	if err := dBClient.Connect(connStr); err != nil {
		log.WithError(err).Fatal("Failed to initialize database connection")
	}
	defer dBClient.Close()

	// Initialize file operations handler
	fileClient := file.NewFileService(log)

	secret, err := fileClient.ReadFile(config.Device.SecretFile)
	if err != nil {
		log.WithError(err).Fatal("Failed to read auth secret")
	}

	var kafkaClient *kafka.KafkaClient

	// Initialize Kafka client only if the mode is queue
	if config.Service.Mode == constants.QUEUE_MODE {
		kafkaClient, err = kafka.NewKafkaClient(
			config.Kafka.SecurityProtocol,
			config.Kafka.SSL.CACert,
			config.Kafka.SSL.Cert,
			config.Kafka.SSL.Key,
			config.Kafka.SASL.Mechanism,
			config.Kafka.SASL.Username,
			config.Kafka.SASL.Password,
			config.Kafka.Brokers,
			config.Kafka.GroupID,
			log,
		)
		if err != nil {
			log.WithError(err).Fatal("Failed to initialize Kafka Client")
		}
	}

	registraionService := services.NewRegistrationService(config.Service.Mode, mqttClient, kafkaClient, dBClient, config.MQTT.Topics.Response, config.MQTT.Topics.Request, config.MQTT.QOS, secret, log)
	registraionService.ListenForDeviceRegistration()

	// Block the main thread to keep services running
	log.Info("Registration service is running...")
	select {}
}
