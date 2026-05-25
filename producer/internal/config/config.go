package config

import (
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Port       string   `yaml:"port"`
	PortGRPC   string   `yaml:"port_grpc"`
	GRPC       bool     `yaml:"grpc"`
	Brokers    []string `yaml:"brokers"`
	Topic      string   `yaml:"topic"`
	AutoSender bool     `yaml:"auto_sender"`
	Searches   []string `yaml:"searches"`
	Users      []string `yaml:"users"`
	LogFile    string   `yaml:"log_file"`
}

func GetConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	if port := os.Getenv("PRODUCER_PORT"); port != "" {
		config.Port = port
	}
	if portGRPC := os.Getenv("PRODUCER_PORT_GRPC"); portGRPC != "" {
		config.PortGRPC = portGRPC
	}
	if grpc := os.Getenv("PRODUCER_GRPC"); grpc == "true" {
		config.GRPC = true
	} else if grpc != "" {
		config.GRPC = false
	}
	if brokers := os.Getenv("PRODUCER_BROKERS"); brokers != "" {
		config.Brokers = strings.Split(brokers, ",")
	}
	if topic := os.Getenv("PRODUCER_TOPIC"); topic != "" {
		config.Topic = topic
	}
	if autoSender := os.Getenv("PRODUCER_AUTO_SENDER"); autoSender == "true" {
		config.AutoSender = true
	}
	if searches := os.Getenv("PRODUCER_SEARCHES"); searches != "" {
		config.Searches = strings.Split(searches, ",")
	}
	if users := os.Getenv("PRODUCER_USERS"); users != "" {
		config.Users = strings.Split(users, ",")
	}
	if logFile := os.Getenv("PRODUCER_LOG_FILE"); logFile != "" {
		config.LogFile = logFile
	}

	return config, nil
}
