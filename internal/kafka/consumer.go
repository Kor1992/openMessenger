package kafka

import (
	"context"
	"encoding/json"
	"log"
	"messanger/internal/repository"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader    *kafka.Reader
	processor repository.MessageProcessor
}

func NewConsumer(
	broker string,
	topic string,
	groupID string,
	processor repository.MessageProcessor,
) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{broker},
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		processor: processor,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			return err
		}

		var event MessageCreatedEvent

		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf(
				"failed to decode message: %v",
				err,
			)

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				return err
			}

			continue
		}

		inserted, err := c.processor.ProcessMessageCreated(
			ctx,
			event.EventID,
			event.MessageID,
			event.ChatID,
			event.SenderID,
		)
		if err != nil {
			return err
		}

		if !inserted {
			log.Printf(
				"event already processed: event_id=%s",
				event.EventID,
			)

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				return err
			}

			continue
		}

		log.Printf(
			"MessageCreated: event_id=%s message_id=%s chat_id=%s sender_id=%s text=%q",
			event.EventID,
			event.MessageID,
			event.ChatID,
			event.SenderID,
			event.Text,
		)

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
