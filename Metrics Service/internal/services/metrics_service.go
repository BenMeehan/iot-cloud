package services

import (
	"encoding/json"

	"github.com/benmeehan/iot-metrics-service/internal/constants"
	"github.com/benmeehan/iot-metrics-service/internal/database"
	"github.com/benmeehan/iot-metrics-service/internal/models"
	"github.com/benmeehan/iot-metrics-service/pkg/kafka"
	"github.com/benmeehan/iot-metrics-service/pkg/mqtt"
	KAFKA "github.com/confluentinc/confluent-kafka-go/kafka"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

// MetricsService manages the reception and storage of metrics data
type MetricsService struct {
	MqttClient  mqtt.MQTTClient
	KafkaClient *kafka.KafkaClient
	DBClient    database.DB
	SubTopic    string
	QOS         int
	Logger      *logrus.Logger
	Mode        string
}

// NewMetricsService creates a new instance of MetricsService
func NewMetricsService(mode string, mqttClient mqtt.MQTTClient, KafkaClient *kafka.KafkaClient, dbClient database.DB, subTopic string, qos int, logger *logrus.Logger) *MetricsService {
	return &MetricsService{
		MqttClient:  mqttClient,
		KafkaClient: KafkaClient,
		DBClient:    dbClient,
		SubTopic:    subTopic,
		QOS:         qos,
		Logger:      logger,
		Mode:        mode,
	}
}

// ListenForDeviceMetrics listens for incoming metrics data either via Kafka or MQTT based on the Mode
func (m *MetricsService) ListenForDeviceMetrics() {
	switch m.Mode {
	case constants.QUEUE_MODE:
		if err := m.KafkaClient.Subscribe(m.SubTopic, m.handleKafkaMessage); err != nil {
			m.Logger.Errorf("Error subscribing to Kafka topic %s: %v", m.SubTopic, err)
		} else {
			m.Logger.Infof("Subscribed to Kafka topic: %s", m.SubTopic)
		}
	default:
		// Subscribe to MQTT
		token := m.MqttClient.Subscribe(m.SubTopic, byte(m.QOS), m.handleMessage)
		token.Wait()
		if token.Error() != nil {
			m.Logger.Errorf("Error subscribing to topic %s: %v", m.SubTopic, token.Error())
		} else {
			m.Logger.Infof("Subscribed to topic: %s", m.SubTopic)
		}
	}
}

// handleMessage processes the metrics data received via MQTT
func (m *MetricsService) handleMessage(client MQTT.Client, msg MQTT.Message) {
	m.Logger.WithFields(logrus.Fields{
		"topic":   msg.Topic(),
		"payload": string(msg.Payload()),
	}).Info("Received message")

	var metrics models.SystemMetrics
	err := json.Unmarshal(msg.Payload(), &metrics)
	if err != nil {
		m.Logger.Errorf("Failed to decode message: %v", err)
		return
	}

	// Insert metrics into the database
	if err := m.insertMetrics(metrics); err != nil {
		m.Logger.Errorf("Error inserting metrics into DB: %v", err)
	} else {
		m.Logger.Infof("Inserted metrics for device: %s", metrics.DeviceID)
	}
}

// handleKafkaMessage processes the metrics data received via Kafka
func (m *MetricsService) handleKafkaMessage(msg *KAFKA.Message) {
	m.Logger.WithFields(logrus.Fields{
		"topic":   *msg.TopicPartition.Topic,
		"payload": string(msg.Value),
	}).Info("Received message from Kafka")

	var metrics models.SystemMetrics
	err := json.Unmarshal(msg.Value, &metrics)
	if err != nil {
		m.Logger.Errorf("Failed to decode Kafka message: %v", err)
		return
	}

	// Insert metrics into the database
	if err := m.insertMetrics(metrics); err != nil {
		m.Logger.Errorf("Error inserting metrics into DB: %v", err)
	} else {
		m.Logger.Infof("Inserted metrics for device: %s", metrics.DeviceID)
	}
}

// insertMetrics inserts the received system and process metrics into the database
func (m *MetricsService) insertMetrics(metrics models.SystemMetrics) error {
	// Insert the main SystemMetrics record
	result := m.DBClient.GetConn().Create(&metrics)
	if result.Error != nil {
		return result.Error
	}

	// Insert each ProcessMetrics for the system
	for processName, processMetrics := range metrics.Processes {
		if processMetrics != nil {
			processMetricsEntry := models.ProcessMetrics{
				DeviceID:    metrics.DeviceID,
				Timestamp:   metrics.Timestamp,
				ProcessName: processName,
				CPUUsage:    processMetrics.CPUUsage,
				Memory:      processMetrics.Memory,
			}
			// Insert process metrics into process_metrics table
			if err := m.DBClient.GetConn().Create(&processMetricsEntry).Error; err != nil {
				m.Logger.Errorf("Failed to insert process metrics for process: %s", processName)
			}
		}
	}
	return nil
}
