package repository

import "context"

type ProcessedEventRepository interface {
	MarkProcessed(ctx context.Context, eventID string) (bool, error)
}
