package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"os"

	"github.com/benmeehan/iot-heartbeat-service/internal/database"
	"github.com/benmeehan/iot-heartbeat-service/internal/models"
	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func InitMQTTClient(broker, clientID, topic, CACert string) mqtt.Client {
	// Load CA certificate
	caCert, err := os.ReadFile(CACert)
	if err != nil {
		logrus.Fatalf("Error loading CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Configure TLS
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)

	clientID = clientID + uuid.New().String()
	logrus.Info("Using MQTT client ID ", clientID)

	opts.SetClientID(clientID)
	opts.SetTLSConfig(tlsConfig)
	opts.OnConnect = func(client mqtt.Client) {
		logrus.Info("Connected to MQTT broker")
		subscribe(client, topic)
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		logrus.Errorf("Connection lost: %v", err)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logrus.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}

	return client
}

func subscribe(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, handleMessage)
	token.Wait()
	if token.Error() != nil {
		logrus.Errorf("Error subscribing to topic %s: %v", topic, token.Error())
	} else {
		logrus.Infof("Subscribed to topic: %s", topic)
	}
}

func handleMessage(client mqtt.Client, msg mqtt.Message) {
	logrus.WithFields(logrus.Fields{
		"topic":   msg.Topic(),
		"payload": string(msg.Payload()),
	}).Info("Received message")

	var hb models.Heartbeat
	err := json.Unmarshal(msg.Payload(), &hb)
	if err != nil {
		logrus.Errorf("Failed to decode message: %v", err)
		return
	}

	if err := insertHeartbeat(hb); err != nil {
		logrus.Errorf("Error inserting heartbeat into DB: %v", err)
	} else {
		logrus.Infof("Inserted heartbeat for device: %s", hb.DeviceID)
	}
}

func insertHeartbeat(hb models.Heartbeat) error {
	result := database.DB.Create(&hb)
	return result.Error
}
