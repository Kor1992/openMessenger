package repository

import "context"

type User struct {
	ID       string
	Username string
}

type UserRepository interface {
	Create(ctx context.Context, username, passwordHash string) (string, error)
	GetByUsername(ctx context.Context, username string) (string, string, error)
	SearchByUsername(ctx context.Context, query string, limit int) ([]User, error)
	GetByID(ctx context.Context, id string) (string, error)
}
