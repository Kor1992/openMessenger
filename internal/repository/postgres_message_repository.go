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

func (r *postgresMessageRepository) ListByChat(ctx context.Context, chatID string, limit, offset int) ([]Message, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(
		ctx,
		`SELECT id, chat_id, sender_id, text, created_at::text
		 FROM messages
		 WHERE chat_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		chatID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.Text, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// reverse to chronological order
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}
