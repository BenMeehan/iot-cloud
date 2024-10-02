package services

import (
	"encoding/json"

	"github.com/benmeehan/iot-heartbeat-service/internal/constants"
	"github.com/benmeehan/iot-heartbeat-service/internal/database"
	"github.com/benmeehan/iot-heartbeat-service/internal/models"
	"github.com/benmeehan/iot-heartbeat-service/pkg/kafka"
	"github.com/benmeehan/iot-heartbeat-service/pkg/mqtt"
	KAFKA "github.com/confluentinc/confluent-kafka-go/kafka"
	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/sirupsen/logrus"
)

type HeartbeatService struct {
	MqttClient  mqtt.MQTTClient
	KafkaClient *kafka.KafkaClient
	DBClient    database.DB
	SubTopic    string
	QOS         int
	Logger      *logrus.Logger
	Mode        string
}

// NewHeartbeatService creates a new instance of HeartbeatService
func NewHeartbeatService(mode string, mqttClient mqtt.MQTTClient, KafkaClient *kafka.KafkaClient, dbClient database.DB, subTopic string, qos int, logger *logrus.Logger) *HeartbeatService {
	return &HeartbeatService{
		MqttClient:  mqttClient,
		KafkaClient: KafkaClient,
		DBClient:    dbClient,
		SubTopic:    subTopic,
		QOS:         qos,
		Logger:      logger,
		Mode:        mode,
	}
}

func (h *HeartbeatService) ListenForDeviceHeartbeats() {
	switch h.Mode {
	case constants.QUEUE_MODE:
		if err := h.KafkaClient.Subscribe(h.SubTopic, h.handleKafkaMessage); err != nil {
			h.Logger.Errorf("Error subscribing to Kafka topic %s: %v", h.SubTopic, err)
		} else {
			h.Logger.Infof("Subscribed to Kafka topic: %s", h.SubTopic)
		}
	default:
		// Subscribe to MQTT
		token := h.MqttClient.Subscribe(h.SubTopic, byte(h.QOS), h.handleMessage)
		token.Wait()
		if token.Error() != nil {
			h.Logger.Errorf("Error subscribing to topic %s: %v", h.SubTopic, token.Error())
		} else {
			h.Logger.Infof("Subscribed to topic: %s", h.SubTopic)
		}
	}
}

func (h *HeartbeatService) handleMessage(client MQTT.Client, msg MQTT.Message) {
	h.Logger.WithFields(logrus.Fields{
		"topic":   msg.Topic(),
		"payload": string(msg.Payload()),
	}).Info("Received message")

	var hb models.Heartbeat
	err := json.Unmarshal(msg.Payload(), &hb)
	if err != nil {
		h.Logger.Errorf("Failed to decode message: %v", err)
		return
	}

	if err := h.insertHeartbeat(hb); err != nil {
		h.Logger.Errorf("Error inserting heartbeat into DB: %v", err)
	} else {
		h.Logger.Infof("Inserted heartbeat for device: %s", hb.DeviceID)
	}
}

func (h *HeartbeatService) handleKafkaMessage(msg *KAFKA.Message) {
	h.Logger.WithFields(logrus.Fields{
		"topic":   *msg.TopicPartition.Topic,
		"payload": string(msg.Value),
	}).Info("Received message from Kafka")

	var hb models.Heartbeat
	err := json.Unmarshal(msg.Value, &hb)
	if err != nil {
		h.Logger.Errorf("Failed to decode Kafka message: %v", err)
		return
	}

	// Insert heartbeat into the database
	if err := h.insertHeartbeat(hb); err != nil {
		h.Logger.Errorf("Error inserting heartbeat into DB: %v", err)
	} else {
		h.Logger.Infof("Inserted heartbeat for device: %s", hb.DeviceID)
	}
}

func (h *HeartbeatService) insertHeartbeat(hb models.Heartbeat) error {
	result := h.DBClient.GetConn().Create(&hb)
	return result.Error
}
