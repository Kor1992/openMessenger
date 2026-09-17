package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"messanger/internal/repository"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	maxRetries     = 5
	baseRetryDelay = 500 * time.Millisecond
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
			slog.Error("failed to decode message",
				"error", err,
				"offset", msg.Offset,
			)

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				return err
			}

			continue
		}

		if err := c.processWithRetry(ctx, event, msg); err != nil {
			return err
		}
	}
}

func (c *Consumer) processWithRetry(
	ctx context.Context,
	event MessageCreatedEvent,
	msg kafka.Message,
) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		inserted, err := c.processor.ProcessMessageCreated(
			ctx,
			event.EventID,
			event.MessageID,
			event.ChatID,
			event.SenderID,
		)
		if err != nil {
			lastErr = err
			delay := baseRetryDelay * time.Duration(1<<uint(attempt))
			slog.Warn("processing failed, retrying",
				"event_id", event.EventID,
				"attempt", attempt+1,
				"max_retries", maxRetries,
				"delay", delay,
				"error", err,
			)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}

			continue
		}

		if !inserted {
			slog.Info("event already processed",
				"event_id", event.EventID,
			)
		} else {
			slog.Info("message processed",
				"event_id", event.EventID,
				"message_id", event.MessageID,
				"chat_id", event.ChatID,
				"sender_id", event.SenderID,
			)
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}

		return nil
	}

	slog.Error("exhausted retries for event",
		"event_id", event.EventID,
		"error", lastErr,
	)

	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		return err
	}

	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
