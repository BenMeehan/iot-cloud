package services

import (
	"encoding/json"
	"fmt"

	"github.com/benmeehan/iot-registration-service/internal/constants"
	"github.com/benmeehan/iot-registration-service/internal/database"
	"github.com/benmeehan/iot-registration-service/internal/models"
	"github.com/benmeehan/iot-registration-service/pkg/kafka"
	"github.com/benmeehan/iot-registration-service/pkg/mqtt"
	KAFKA "github.com/confluentinc/confluent-kafka-go/kafka"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RegistrationService manages the device registration process on the cloud side
type RegistrationService struct {
	MqttClient  mqtt.MQTTClient
	KafkaClient *kafka.KafkaClient
	DBClient    database.DB
	PubTopic    string
	SubTopic    string
	QOS         int
	Secret      string
	Logger      *logrus.Logger
	Mode        string
}

// NewRegistrationService creates a new instance of RegistrationService
func NewRegistrationService(mode string, mqttClient mqtt.MQTTClient, kafkaClient *kafka.KafkaClient, dbClient database.DB, pubTopic, subTopic string, qos int, secret string, logger *logrus.Logger) *RegistrationService {
	return &RegistrationService{
		MqttClient:  mqttClient,
		KafkaClient: kafkaClient,
		DBClient:    dbClient,
		PubTopic:    pubTopic,
		SubTopic:    subTopic,
		QOS:         qos,
		Secret:      secret,
		Logger:      logger,
		Mode:        mode,
	}
}

// ListenForDeviceRegistration subscribes to the appropriate messaging system based on the mode
func (rs *RegistrationService) ListenForDeviceRegistration() {
	switch rs.Mode {
	case constants.QUEUE_MODE:
		rs.listenForDeviceRegistrationKafka()
	default:
		rs.listenForDeviceRegistrationMQTT()
	}
}

// listenForDeviceRegistrationMQTT subscribes to the registration topic via MQTT
func (rs *RegistrationService) listenForDeviceRegistrationMQTT() {
	token := rs.MqttClient.Subscribe(rs.SubTopic, byte(rs.QOS), rs.handleRegistrationRequest)
	token.Wait()
	if err := token.Error(); err != nil {
		rs.Logger.WithError(err).Fatal("Failed to subscribe to registration topic")
	}
}

// listenForDeviceRegistrationKafka consumes messages from the Kafka topic
func (rs *RegistrationService) listenForDeviceRegistrationKafka() {
	// Assume KafkaClient has a method to consume messages from a topic
	go func() {
		err := rs.KafkaClient.Subscribe(rs.SubTopic, rs.handleRegistrationRequestKafka)
		if err != nil {
			rs.Logger.WithError(err).Fatal("Failed to consume from Kafka topic")
		}
	}()
}

// handleRegistrationRequest processes incoming device registration requests via MQTT
func (rs *RegistrationService) handleRegistrationRequest(client MQTT.Client, msg MQTT.Message) {
	rs.processRegistrationRequest(msg.Payload())
}

// handleRegistrationRequestKafka processes incoming device registration requests via Kafka
func (rs *RegistrationService) handleRegistrationRequestKafka(message KAFKA.Message) {
	payload := message.Value
	rs.processRegistrationRequest(payload)
}

// processRegistrationRequest is the shared logic for processing registration requests
func (rs *RegistrationService) processRegistrationRequest(payload []byte) {
	var request map[string]string
	if err := json.Unmarshal(payload, &request); err != nil {
		rs.Logger.WithError(err).Error("Error parsing registration request")
		return
	}

	clientID, deviceSecret, err := extractFields(request)
	if err != nil {
		rs.Logger.WithError(err).Error("Failed to extract fields from registration request")
		return
	}

	deviceID, err := rs.registerDevice(clientID, deviceSecret)
	if err != nil {
		rs.Logger.WithError(err).Error("Failed to register device")
		return
	}

	if err := rs.sendRegistrationResponse(clientID, deviceID); err != nil {
		rs.Logger.WithError(err).Error("Failed to send registration response")
	}
}

// extractFields retrieves client ID and device secret from the payload
func extractFields(payload map[string]string) (string, string, error) {
	clientID, exists := payload["client_id"]
	if !exists {
		return "", "", fmt.Errorf("client ID not found in the registration request")
	}

	deviceSecret, exists := payload["device_secret"]
	if !exists {
		return "", "", fmt.Errorf("device secret not found in the registration request")
	}

	return clientID, deviceSecret, nil
}

// registerDevice stores the device information in the database
func (rs *RegistrationService) registerDevice(clientID, deviceSecret string) (string, error) {
	if deviceSecret != rs.Secret {
		return "", fmt.Errorf("invalid device secret provided for client: %s", clientID)
	}

	deviceID, err := generateDeviceID()
	if err != nil {
		return "", err
	}

	device := models.NewDevice(deviceID)
	if err := rs.DBClient.SaveDevice(device); err != nil {
		return "", err
	}

	return deviceID, nil
}

// generateDeviceID creates a new unique device ID
func generateDeviceID() (string, error) {
	id, err := uuid.NewRandom() // Using NewRandom for uniqueness without time-sort
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// sendRegistrationResponse publishes the registration response back to the device
func (rs *RegistrationService) sendRegistrationResponse(clientID, deviceID string) error {
	response := map[string]string{"device_id": deviceID}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	responseTopic := fmt.Sprintf("%s/%s", rs.PubTopic, clientID)
	token := rs.MqttClient.Publish(responseTopic, byte(rs.QOS), false, responseBytes)
	token.Wait()
	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to publish registration response: %w", err)
	}

	rs.Logger.Infof("Device %s registered successfully with ID: %s", clientID, deviceID)
	return nil
}
