package utils

import (
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Config represents the structure of the configuration file.
type Config struct {
	MQTT struct {
		Broker   string `yaml:"broker"`    // MQTT broker address
		ClientID string `yaml:"client_id"` // MQTT client ID
		QOS      int    `yaml:"QOS"`       // MQTT Quality of Service
		TLS      struct {
			CACert string `yaml:"ca_cert"` // Path to the CA certificate
		} `yaml:"tls"`
		Topic string `yaml:"topic"` // MQTT topic
	} `yaml:"mqtt"`

	Kafka struct {
		Brokers         []string `yaml:"brokers"`         // List of Kafka brokers
		ClientID        string   `yaml:"client_id"`      // Kafka client ID
		SecurityProtocol string   `yaml:"security_protocol"` // Security protocol
		SSL             struct {
			CACert string `yaml:"ca_cert"` // Path to the CA certificate
			Cert   string `yaml:"cert"`    // Path to the client certificate
			Key    string `yaml:"key"`     // Path to the client key
		} `yaml:"ssl"`
		SASL struct {
			Mechanism string `yaml:"mechanism"` // SASL mechanism
			Username  string `yaml:"username"`  // SASL username
			Password  string `yaml:"password"`  // SASL password
		} `yaml:"sasl"`
	} `yaml:"kafka"`

	TopicMappings []struct {
		MQTTTopic  string `yaml:"mqtt_topic"`  // MQTT topic
		KafkaTopic string `yaml:"kafka_topic"` // Kafka topic
	} `yaml:"topic_mappings"`
}


// LoadConfig loads the YAML configuration from the specified file.
// It returns a pointer to the Config struct and an error if loading fails.
func LoadConfig(filename string, logger *logrus.Logger) (*Config, error) {
	logger.Infof("Loading configuration from file: %s", filename)

	file, err := os.Open(filename)
	if err != nil {
		logger.Errorf("Failed to open config file: %v", err)
		return nil, err
	}
	defer file.Close()

	// Decode the YAML file into the Config struct
	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		logger.Errorf("Failed to decode config file: %v", err)
		return nil, err
	}

	logger.Infof("Configuration loaded successfully from %s", filename)
	return &config, nil
}
