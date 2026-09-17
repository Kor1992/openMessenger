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

func (r *postgresUserRepository) SearchByUsername(ctx context.Context, query string, limit int) ([]User, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.Query(
		ctx,
		`SELECT id, username FROM users
		 WHERE username ILIKE '%' || $1 || '%'
		 LIMIT $2`,
		query, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id string) (string, error) {
	var username string
	err := r.db.QueryRow(ctx, `SELECT username FROM users WHERE id = $1`, id).Scan(&username)
	return username, err
}
