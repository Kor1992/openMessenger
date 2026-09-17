package repository

import "context"

type OutboxEvent struct {
	ID          string
	EventType   string
	AggregateID string
	Payload     []byte
}

type OutboxRepository interface {
	GetUnpublished(
		ctx context.Context,
		limit int,
	) ([]OutboxEvent, error)

	MarkPublished(
		ctx context.Context,
		id string,
	) error

	Cleanup(
		ctx context.Context,
		olderThanDays int,
	) (int64, error)
}
