package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresProcessedEventRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProcessedEventRepository(
	db *pgxpool.Pool,
) ProcessedEventRepository {
	return &postgresProcessedEventRepository{
		db: db,
	}
}

func (r *postgresProcessedEventRepository) MarkProcessed(
	ctx context.Context,
	eventID string,
) (bool, error) {
	result, err := r.db.Exec(
		ctx,
		`INSERT INTO processed_events (event_id)
		 VALUES ($1)
		 ON CONFLICT (event_id) DO NOTHING`,
		eventID,
	)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() == 1, nil
}
