package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) UserRepository {
	return &postgresUserRepository{
		db: db,
	}
}

func (r *postgresUserRepository) Create(
	ctx context.Context,
	username, passwordHash string,
) (string, error) {
	var id string

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO users (username, password_hash)
		 VALUES ($1, $2)
		 RETURNING id`,
		username,
		passwordHash,
	).Scan(&id)

	return id, err
}

func (r *postgresUserRepository) GetByUsername(
	ctx context.Context,
	username string,
) (string, string, error) {
	var id string
	var passwordHash string

	err := r.db.QueryRow(
		ctx,
		`SELECT id, password_hash
		 FROM users
		 WHERE username = $1`,
		username,
	).Scan(&id, &passwordHash)

	return id, passwordHash, err
}
