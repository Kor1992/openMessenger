package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresMessageRepository struct {
	db *pgxpool.Pool
}

func NewPostgresMessageRepository(db *pgxpool.Pool) MessageRepository {
	return &postgresMessageRepository{db: db}
}

func (r *postgresMessageRepository) CreateMessage(
	ctx context.Context,
	chatID string,
	senderID string,
	text string,
) (string, error) {
	var id string

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO messages (chat_id, sender_id, text)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		chatID,
		senderID,
		text,
	).Scan(&id)

	return id, err
}

func (r *postgresMessageRepository) CreateMessageWithOutbox(
	ctx context.Context,
	messageID string,
	chatID string,
	senderID string,
	text string,
	payload []byte,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`INSERT INTO messages (id, chat_id, sender_id, text)
		 VALUES ($1, $2, $3, $4)`,
		messageID,
		chatID,
		senderID,
		text,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO outbox_events (
			event_type,
			aggregate_id,
			payload
		)
		VALUES ($1, $2, $3)`,
		"MessageCreated",
		messageID,
		payload,
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
