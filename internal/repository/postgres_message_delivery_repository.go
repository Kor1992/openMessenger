package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresMessageDeliveryRepository struct {
	db *pgxpool.Pool
}

func NewPostgresMessageDeliveryRepository(
	db *pgxpool.Pool,
) MessageDeliveryRepository {
	return &postgresMessageDeliveryRepository{db: db}
}

func (r *postgresMessageDeliveryRepository) Create(
	ctx context.Context,
	messageID, userID string,
) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO message_deliveries (message_id, user_id)
		 VALUES ($1, $2)
		 ON CONFLICT (message_id, user_id) DO NOTHING`,
		messageID,
		userID,
	)

	return err
}
