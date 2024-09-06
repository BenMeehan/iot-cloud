package config

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	MQTT struct {
		Broker   string `yaml:"broker"`
		ClientID string `yaml:"client_id"`
		Topic    string `yaml:"topic"`
		TLS      struct {
			CACert string `yaml:"ca_cert"`
		} `yaml:"tls"`
	} `yaml:"mqtt"`

	Database struct {
		Host     string `yaml:"host"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		Port     int    `yaml:"port"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`
}

var AppConfig Config

func LoadConfig(file []byte) {
	err := yaml.Unmarshal(file, &AppConfig)
	if err != nil {
		logrus.Fatalf("Error parsing config file: %v", err)
	}
}
