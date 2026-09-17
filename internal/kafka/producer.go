package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type MessageCreatedEvent struct {
	EventID   string    `json:"event_id"`
	MessageID string    `json:"message_id"`
	ChatID    string    `json:"chat_id"`
	SenderID  string    `json:"sender_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    topic,
			Balancer: &kafka.Hash{},
		},
	}
}

func (p *Producer) PublishMessageCreated(
	ctx context.Context,
	event MessageCreatedEvent,
) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte(event.ChatID),
			Value: data,
		},
	)
}

func (p *Producer) Publish(
	ctx context.Context,
	key string,
	payload []byte,
) error {
	return p.writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte(key),
			Value: payload,
		},
	)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
