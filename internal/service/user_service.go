package service

import (
	"context"
	"errors"
	"messanger/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameTaken = errors.New("username already taken")
	ErrInvalidInput  = errors.New("invalid input")
)

type UserService interface {
	Register(ctx context.Context, username, password string) (string, error)
	SearchByUsername(ctx context.Context, query string) ([]repository.User, error)
	GetByID(ctx context.Context, id string) (string, error)
}

type userService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) UserService {
	return &userService{
		users: users,
	}
}

func (s *userService) Register(
	ctx context.Context,
	username, password string,
) (string, error) {
	if username == "" || password == "" {
		return "", ErrInvalidInput
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	id, err := s.users.Create(ctx, username, string(passwordHash))
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", ErrUsernameTaken
		}

		return "", err
	}

	return id, nil
}

func (s *userService) SearchByUsername(ctx context.Context, query string) ([]repository.User, error) {
	return s.users.SearchByUsername(ctx, query, 20)
}

func (s *userService) GetByID(ctx context.Context, id string) (string, error) {
	return s.users.GetByID(ctx, id)
}
