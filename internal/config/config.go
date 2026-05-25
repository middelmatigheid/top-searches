package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Port           string   `yaml:"port"`
	PortGRPC       string   `yaml:"port_grpc"`
	GRPC           bool     `yaml:"grpc"`
	Timespan       int64    `yaml:"timespan"`
	BucketTimespan int64    `yaml:"bucket_timespan"`
	Cooldown       int64    `yaml:"cooldown"`
	Brokers        []string `yaml:"brokers"`
	Topic          string   `yaml:"topic"`
	Stoplist       []string `yaml:"stoplist"`
	LogFile        string   `yaml:"log_file"`
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

	if port := os.Getenv("PORT"); port != "" {
		config.Port = port
	}
	if portGRPC := os.Getenv("PORT_GRPC"); portGRPC != "" {
		config.PortGRPC = portGRPC
	}
	if grpc := os.Getenv("GRPC"); grpc == "true" {
		config.GRPC = true
	}
	if timespan := os.Getenv("TIMESPAN"); timespan != "" {
		time, err := strconv.ParseInt(timespan, 10, 64)
		if err != nil {
			return nil, err
		}
		config.Timespan = time
	}
	if bucketTimespan := os.Getenv("BUCKET_TIMESPAN"); bucketTimespan != "" {
		time, err := strconv.ParseInt(bucketTimespan, 10, 64)
		if err != nil {
			return nil, err
		}
		config.BucketTimespan = time
	}
	if cooldown := os.Getenv("COOLDOWN"); cooldown != "" {
		time, err := strconv.ParseInt(cooldown, 10, 64)
		if err != nil {
			return nil, err
		}
		config.Cooldown = time
	}
	if brokers := os.Getenv("BROKERS"); brokers != "" {
		config.Brokers = strings.Split(brokers, ",")
	}
	if topic := os.Getenv("TOPIC"); topic != "" {
		config.Topic = topic
	}
	if stoplist := os.Getenv("STOPLIST"); stoplist != "" {
		config.Stoplist = strings.Split(stoplist, ",")
	}
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		config.LogFile = logFile
	}

	if config.Timespan%config.BucketTimespan != 0 {
		return nil, errors.New("Bucket timespan should be divisor of Timespan")
	}

	return config, nil
}
