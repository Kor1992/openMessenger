package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresMessageProcessor struct {
	db *pgxpool.Pool
}

func NewPostgresMessageProcessor(
	db *pgxpool.Pool,
) MessageProcessor {
	return &postgresMessageProcessor{db: db}
}

func (r *postgresMessageProcessor) ProcessMessageCreated(
	ctx context.Context,
	eventID string,
	messageID string,
	chatID string,
	senderID string,
) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	var inserted bool

	err = tx.QueryRow(
		ctx,
		`INSERT INTO processed_events (event_id)
		 VALUES ($1)
		 ON CONFLICT (event_id) DO NOTHING
		 RETURNING true`,
		eventID,
	).Scan(&inserted)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return false, nil
		}

		return false, err
	}

	rows, err := tx.Query(
		ctx,
		`SELECT user_id
		 FROM chat_members
		 WHERE chat_id = $1
		   AND user_id <> $2`,
		chatID,
		senderID,
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string

		if err := rows.Scan(&userID); err != nil {
			return false, err
		}

		_, err = tx.Exec(
			ctx,
			`INSERT INTO message_deliveries (message_id, user_id)
			 VALUES ($1, $2)
			 ON CONFLICT (message_id, user_id) DO NOTHING`,
			messageID,
			userID,
		)
		if err != nil {
			return false, err
		}
	}

	if err := rows.Err(); err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	return inserted, nil
}
