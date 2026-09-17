package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresChatRepository struct {
	db *pgxpool.Pool
}

func NewPostgresChatRepository(db *pgxpool.Pool) ChatRepository {
	return &postgresChatRepository{db: db}
}

func (r *postgresChatRepository) Create(ctx context.Context, creatorID string) (string, error) {
	tx, err := r.db.Begin(ctx)

	if err != nil {
		return "", err
	}

	defer tx.Rollback(ctx)

	var chatID string

	err = tx.QueryRow(
		ctx,
		`INSERT INTO chats DEFAULT VALUES RETURNING id`,
	).Scan(&chatID)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO chat_members (chat_id, user_id)
		 VALUES ($1, $2)`,
		chatID,
		creatorID,
	)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return chatID, nil

}

func (r *postgresChatRepository) AddMember(ctx context.Context, chatID, userID string) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO chat_members (chat_id, user_id)
		 VALUES ($1, $2)`,
		chatID,
		userID,
	)

	return err
}

func (r *postgresChatRepository) IsMember(
	ctx context.Context,
	chatID string,
	userID string,
) (bool, error) {
	var exists bool

	err := r.db.QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM chat_members
			WHERE chat_id = $1 AND user_id = $2
		)`,
		chatID,
		userID,
	).Scan(&exists)

	return exists, err
}

func (r *postgresChatRepository) GetMemberIDs(
	ctx context.Context,
	chatID string,
) ([]string, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT user_id
		 FROM chat_members
		 WHERE chat_id = $1`,
		chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string

	for rows.Next() {
		var userID string

		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}

		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (r *postgresChatRepository) ListByUser(ctx context.Context, userID string) ([]ChatInfo, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT
			c.id,
			c.created_at::text,
			COALESCE(m.text, '') AS last_message,
			COALESCE(m.sender_id, '') AS last_sender_id,
			COALESCE(m.created_at::text, '') AS last_msg_time,
			(SELECT count(*) FROM chat_members WHERE chat_id = c.id) AS member_count
		 FROM chats c
		 JOIN chat_members cm ON cm.chat_id = c.id
		 LEFT JOIN messages m ON m.chat_id = c.id
		   AND m.created_at = (
		     SELECT MAX(created_at) FROM messages WHERE chat_id = c.id
		   )
		 WHERE cm.user_id = $1
		 ORDER BY COALESCE(m.created_at, c.created_at) DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []ChatInfo
	for rows.Next() {
		var ci ChatInfo
		if err := rows.Scan(&ci.ID, &ci.CreatedAt, &ci.LastMessage, &ci.LastSenderID, &ci.LastMsgTime, &ci.MemberCount); err != nil {
			return nil, err
		}
		chats = append(chats, ci)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chats, nil
}
