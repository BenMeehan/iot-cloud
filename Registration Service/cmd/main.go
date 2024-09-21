package main

import (
	"fmt"
	"os"

	"github.com/benmeehan/iot-registration-service/internal/database"
	"github.com/benmeehan/iot-registration-service/internal/services"
	"github.com/benmeehan/iot-registration-service/internal/utils"
	"github.com/benmeehan/iot-registration-service/pkg/file"
	"github.com/benmeehan/iot-registration-service/pkg/mqtt"
	"github.com/google/uuid"

	"github.com/sirupsen/logrus"
)

func main() {
	// Set up structured logging with JSON output
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	// Load configuration from file
	config, err := utils.LoadConfig("config/config.yaml")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Generate a unique MQTT Client ID by appending a UUID
	config.MQTT.ClientID = config.MQTT.ClientID + "-" + uuid.New().String()
	logrus.Infof("Using MQTT Client ID: %s", config.MQTT.ClientID)

	// Initialize the shared MQTT connection
	mqttClient := mqtt.NewMqttService()
	err = mqttClient.Initialize(config.MQTT.Broker, config.MQTT.ClientID, config.MQTT.TLS.CACert)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize MQTT connection")
	}

	// Initialize the database connection
	dBClient := database.NewDatabase()
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.DB.Host, config.DB.Port, config.DB.User, config.DB.Password, config.DB.Name, config.DB.SSLMode)

	if err := dBClient.Connect(connStr); err != nil {
		logrus.WithError(err).Fatal("Failed to initialize database connection")
	}
	defer dBClient.Close()

	// Initialize file operations handler
	fileClient := file.NewFileService()

	secret, err := fileClient.ReadFile(config.Device.SecretFile)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to read auth secret")
	}

	registraionService := &services.RegistrationService{
		MqttClient: mqttClient,
		DBClient:   dBClient,
		QOS:        config.MQTT.QOS,
		SubTopic:   config.MQTT.Topics.Request,
		PubTopic:   config.MQTT.Topics.Response,
		Secret:     secret,
	}

	registraionService.ListenForDeviceRegistration()

	// Block the main thread to keep services running
	logrus.Info("Registration service is running...")
	select {}
}
