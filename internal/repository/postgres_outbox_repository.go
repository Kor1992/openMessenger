package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresOutboxRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOutboxRepository(db *pgxpool.Pool) OutboxRepository {
	return &postgresOutboxRepository{
		db: db,
	}
}

func (r *postgresOutboxRepository) GetUnpublished(
	ctx context.Context,
	limit int,
) ([]OutboxEvent, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
			id,
			event_type,
			aggregate_id,
			payload
		FROM outbox_events
		WHERE published_at IS NULL
		ORDER BY created_at
		LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []OutboxEvent

	for rows.Next() {
		var event OutboxEvent

		if err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.AggregateID,
			&event.Payload,
		); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *postgresOutboxRepository) MarkPublished(
	ctx context.Context,
	id string,
) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE outbox_events
		 SET published_at = NOW()
		 WHERE id = $1
		   AND published_at IS NULL`,
		id,
	)

	return err
}

func (r *postgresOutboxRepository) Cleanup(
	ctx context.Context,
	olderThanDays int,
) (int64, error) {
	result, err := r.db.Exec(
		ctx,
		`DELETE FROM outbox_events
		 WHERE published_at IS NOT NULL
		   AND published_at < NOW() - ($1 || ' days')::INTERVAL`,
		olderThanDays,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
