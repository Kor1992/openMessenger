package service

import (
	"context"
	"errors"
	"messanger/internal/auth"
	"messanger/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService interface {
	Login(ctx context.Context, username, password string) (string, error)
}

type authService struct {
	users      repository.UserRepository
	jwtManager *auth.JWTManager
}

func NewAuthService(
	users repository.UserRepository,
	jwtManager *auth.JWTManager,
) AuthService {
	return &authService{
		users:      users,
		jwtManager: jwtManager,
	}
}

func (s *authService) Login(ctx context.Context, username, password string) (string, error) {
	if username == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	userID, passwordHash, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password),
	); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.jwtManager.Generate(userID)
	if err != nil {
		return "", err
	}

	return token, nil
}
