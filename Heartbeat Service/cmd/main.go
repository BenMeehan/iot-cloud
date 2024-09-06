package main

import (
	"os"

	"github.com/benmeehan/iot-heartbeat-service/internal/database"
	"github.com/benmeehan/iot-heartbeat-service/internal/mqtt"
	config "github.com/benmeehan/iot-heartbeat-service/internal/utils"

	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	logrus.Info("Starting heartbeat service...")

	// Load configuration from YAML file
	file, err := os.ReadFile("config/config.yaml")
	if err != nil {
		logrus.Fatalf("Error reading config file: %v", err)
	}
	config.LoadConfig(file)

	// Initialize TimescaleDB connection
	database.InitDB(config.AppConfig.Database.User,
		config.AppConfig.Database.Password,
		config.AppConfig.Database.Host,
		config.AppConfig.Database.DBName,
		config.AppConfig.Database.SSLMode,
		config.AppConfig.Database.Port)

	// Initialize MQTT client and start listening for heartbeats
	mqtt.InitMQTTClient(config.AppConfig.MQTT.Broker,
		config.AppConfig.MQTT.ClientID,
		config.AppConfig.MQTT.Topic,
		config.AppConfig.MQTT.TLS.CACert)

	// Block the main thread to keep services running
	logrus.Info("Heartbeat service is running...")
	select {}
}
