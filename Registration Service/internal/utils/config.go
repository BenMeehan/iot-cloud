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
			CACert string `yaml:"ca_cert"`
		} `yaml:"tls"` // Path to the CA certificate
		Topics struct {
			Request  string `yaml:"request"`
			Response string `yaml:"response"`
		} `yaml:"topics"`
	} `yaml:"mqtt"`

	DB struct {
		Host     string `yaml:"host"`     // Database host address
		Port     int    `yaml:"port"`     // Database port
		User     string `yaml:"user"`     // Database user
		Password string `yaml:"password"` // Database password
		Name     string `yaml:"dbname"`   // Database name
		SSLMode  string `yaml:"sslmode"`  // SSL mode for the connection
	} `yaml:"database"`

	Device struct {
		SecretFile string `yaml:"secret_file"` // Device secret location
	} `yaml:"device"`

	Kafka struct {
		Topic            string   `yaml:"brokers"`           // Kafka topic
		Brokers          []string `yaml:"brokers"`           // List of Kafka brokers
		ClientID         string   `yaml:"client_id"`         // Kafka client ID
		SecurityProtocol string   `yaml:"security_protocol"` // Security protocol
		GroupID          string   `yaml:"group_id"`          // Consumer Group ID
		SSL              struct {
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

	Service struct {
		Mode string `yaml:"mode"` // Direct MQTT or Queue mode
	} `yaml:"service"`
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
