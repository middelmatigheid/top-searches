package service

import (
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"github.com/middelmatigheid/top-searches/producer/internal/models"
)

type Producer interface {
	SendMessage()
}

type Service struct {
	producer sarama.SyncProducer
	topic    string
}

func NewService(producer sarama.SyncProducer, topic string) *Service {
	return &Service{
		producer: producer,
		topic:    topic,
	}
}

func (s *Service) SendSearch(search models.Search) error {
	event := models.SearchEvent{
		Search:    search.Search,
		User:      search.User,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: s.topic,
		Key:   sarama.StringEncoder(event.User),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = s.producer.SendMessage(msg)
	return err
}
