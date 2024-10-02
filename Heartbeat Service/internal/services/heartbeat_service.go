package services

import (
	"encoding/json"

	"github.com/benmeehan/iot-heartbeat-service/internal/database"
	"github.com/benmeehan/iot-heartbeat-service/internal/models"
	"github.com/benmeehan/iot-heartbeat-service/pkg/mqtt"
	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/sirupsen/logrus"
)

type HeartbeatService struct {
	MqttClient mqtt.MQTTClient
	DBClient   database.DB
	SubTopic   string
	QOS        int
	Logger     *logrus.Logger
}

func (h *HeartbeatService) ListenForDeviceHeartbeats() {
	token := h.MqttClient.Subscribe(h.SubTopic, byte(h.QOS), h.handleMessage)
	token.Wait()
	if token.Error() != nil {
		h.Logger.Errorf("Error subscribing to topic %s: %v", h.SubTopic, token.Error())
	} else {
		h.Logger.Infof("Subscribed to topic: %s", h.SubTopic)
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

func (h *HeartbeatService) insertHeartbeat(hb models.Heartbeat) error {
	result := h.DBClient.GetConn().Create(&hb)
	return result.Error
}
