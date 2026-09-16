package repository

import "context"

type UserRepository interface {
	Create(ctx context.Context, username, passwordHash string) (string, error)
	GetByUsername(ctx context.Context, username string) (string, string, error)
}
