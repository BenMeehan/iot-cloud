package services

import (
	"encoding/json"
	"fmt"

	"github.com/benmeehan/iot-registration-service/internal/database"
	"github.com/benmeehan/iot-registration-service/internal/models"
	"github.com/benmeehan/iot-registration-service/pkg/mqtt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RegistrationService manages the device registration process on the cloud side
type RegistrationService struct {
	MqttClient mqtt.MQTTClient
	DBClient   database.DB
	PubTopic   string
	SubTopic   string
	QOS        int
	Secret     string
	Logger     *logrus.Logger
}

// listenForDeviceRegistration subscribes to the registration topic and processes incoming registration requests
func (rs *RegistrationService) ListenForDeviceRegistration() {
	// Subscribe to the registration topic
	token := rs.MqttClient.Subscribe(rs.SubTopic, byte(rs.QOS), rs.handleRegistrationRequest)
	token.Wait()
	if err := token.Error(); err != nil {
		rs.Logger.WithError(err).Fatal("Failed to subscribe to registration topic")
	}
}

// handleRegistrationRequest processes incoming device registration requests
func (rs *RegistrationService) handleRegistrationRequest(client MQTT.Client, msg MQTT.Message) {
	var payload map[string]string
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		rs.Logger.WithError(err).Error("Error parsing registration request")
		return
	}

	// Extract fields from the payload
	clientID, exists := payload["client_id"]
	if !exists {
		rs.Logger.Error("Client ID not found in the registration request")
		return
	}

	deviceSecret, exists := payload["device_secret"]
	if !exists {
		rs.Logger.Error("Device secret not found in the registration request")
		return
	}

	// Save the device information to the database
	deviceID, err := rs.registerDevice(clientID, deviceSecret)
	if err != nil {
		rs.Logger.WithError(err).Error("Failed to register device")
		return
	}

	// Publish a registration response back to the device
	response := map[string]string{"device_id": deviceID}
	responseBytes, _ := json.Marshal(response)
	responseTopic := fmt.Sprintf("%s/%s", rs.PubTopic, clientID)
	rs.MqttClient.Publish(responseTopic, byte(rs.QOS), false, responseBytes)
	rs.Logger.Infof("Device %s registered successfully with ID: %s", clientID, deviceID)
}

// registerDevice stores the device information in the database
func (rs *RegistrationService) registerDevice(clientID, deviceSecret string) (string, error) {
	// Compare the provided device secret with the stored secret
	if deviceSecret != rs.Secret {
		return "", fmt.Errorf("invalid device secret provided for client: %s", clientID)
	}

	// Generate a new device ID
	deviceID, err := generateDeviceID()
	if err != nil {
		return "", err
	}

	// Create a new Device instance
	device := models.NewDevice(deviceID)

	// Save the device information in the database
	if err := rs.DBClient.SaveDevice(device); err != nil {
		return "", err
	}
	return deviceID, nil
}

// generateDeviceID creates a new unique device ID
// using uuid v7 as it is time sortable
func generateDeviceID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
