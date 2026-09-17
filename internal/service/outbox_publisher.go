package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"messanger/internal/kafka"
	"messanger/internal/repository"
	"time"
)

type OutboxPublisher struct {
	outbox          repository.OutboxRepository
	producer        *kafka.Producer
	interval        time.Duration
	batchSize       int
	cleanupDays     int
	consecutiveErrs int
}

func NewOutboxPublisher(
	outbox repository.OutboxRepository,
	producer *kafka.Producer,
) *OutboxPublisher {
	return &OutboxPublisher{
		outbox:      outbox,
		producer:    producer,
		interval:    1 * time.Second,
		batchSize:   100,
		cleanupDays: 7,
	}
}

func (p *OutboxPublisher) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		if err := p.publish(ctx); err != nil {
			p.consecutiveErrs++
			slog.Error("outbox publisher error",
				"error", err,
				"consecutive_errors", p.consecutiveErrs,
			)
		} else {
			p.consecutiveErrs = 0
		}

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
		case <-cleanupTicker.C:
			p.cleanup(ctx)
		}
	}
}

func (p *OutboxPublisher) cleanup(ctx context.Context) {
	deleted, err := p.outbox.Cleanup(ctx, p.cleanupDays)
	if err != nil {
		slog.Error("outbox cleanup failed", "error", err)
		return
	}

	if deleted > 0 {
		slog.Info("outbox cleanup completed", "deleted_events", deleted)
	}
}

func (p *OutboxPublisher) publish(ctx context.Context) error {
	events, err := p.outbox.GetUnpublished(ctx, p.batchSize)
	if err != nil {
		return err
	}

	for _, event := range events {
		var parsed struct {
			ChatID string `json:"chat_id"`
		}

		key := event.AggregateID
		if err := json.Unmarshal(event.Payload, &parsed); err == nil && parsed.ChatID != "" {
			key = parsed.ChatID
		}

		if err := p.producer.Publish(ctx, key, event.Payload); err != nil {
			return err
		}

		if err := p.outbox.MarkPublished(ctx, event.ID); err != nil {
			return err
		}

		slog.Info("outbox event published",
			"id", event.ID,
			"type", event.EventType,
			"aggregate_id", event.AggregateID,
		)
	}

	return nil
}
