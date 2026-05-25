package consumer

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/middelmatigheid/top-searches/internal/models"
)

type Service interface {
	AddSearch(search string, user string, searchTime time.Time) error
}

type Consumer struct {
	consumer sarama.Consumer
	topic    string
	service  Service
	logger   *slog.Logger
}

func NewConsumer(brokers []string, topic string, service Service, logger *slog.Logger) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: consumer,
		topic:    topic,
		service:  service,
		logger:   logger,
	}, nil
}

func (c *Consumer) Start() error {
	partitions, err := c.consumer.Partitions(c.topic)
	if err != nil {
		return err
	}

	for _, partition := range partitions {
		pc, err := c.consumer.ConsumePartition(c.topic, partition, sarama.OffsetNewest)
		if err != nil {
			return err
		}

		go func(pc sarama.PartitionConsumer) {
			for message := range pc.Messages() {
				c.handleMessage(message)
			}
		}(pc)
	}

	return nil
}

func (c *Consumer) handleMessage(msg *sarama.ConsumerMessage) {
	var event models.SearchEvent

	err := json.Unmarshal(msg.Value, &event)
	if err != nil {
		c.logger.Error("Error while parsing the message", "error", err, "message", msg.Value)
		return
	}

	err = c.service.AddSearch(event.Search, event.User, event.Timestamp)
	if err != nil {
		c.logger.Error("Error while adding search", "error", err, "search", event.Search, "timestamp", event.Timestamp)
	} else {
		c.logger.Debug("Event was handled successfully", "search", event.Search, "timestamp", event.Timestamp)
	}
}
