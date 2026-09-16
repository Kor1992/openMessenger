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
