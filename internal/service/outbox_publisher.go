package service

import (
	"context"
	"log"
	"messanger/internal/kafka"
	"messanger/internal/repository"
	"time"
)

type OutboxPublisher struct {
	outbox    repository.OutboxRepository
	producer  *kafka.Producer
	interval  time.Duration
	batchSize int
}

func NewOutboxPublisher(
	outbox repository.OutboxRepository,
	producer *kafka.Producer,
) *OutboxPublisher {
	return &OutboxPublisher{
		outbox:    outbox,
		producer:  producer,
		interval:  1 * time.Second,
		batchSize: 100,
	}
}

func (p *OutboxPublisher) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		if err := p.publish(ctx); err != nil {
			log.Printf("outbox publisher error: %v", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
		}
	}
}

func (p *OutboxPublisher) publish(ctx context.Context) error {
	events, err := p.outbox.GetUnpublished(ctx, p.batchSize)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := p.producer.Publish(
			ctx,
			event.Payload,
		); err != nil {
			return err
		}

		if err := p.outbox.MarkPublished(ctx, event.ID); err != nil {
			return err
		}

		log.Printf(
			"outbox event published: id=%s type=%s aggregate_id=%s",
			event.ID,
			event.EventType,
			event.AggregateID,
		)
	}

	return nil
}
