package service

import (
	"context"
	"messanger/internal/repository"
)

type ChatService interface {
	CreateChat(ctx context.Context, creatorID string) (string, error)
	AddMember(ctx context.Context, chatID, userID string) error
}

type chatService struct {
	chats repository.ChatRepository
}

func NewChatService(chats repository.ChatRepository) ChatService {
	return &chatService{
		chats: chats,
	}
}

func (s *chatService) CreateChat(ctx context.Context, creatorID string) (string, error) {
	return s.chats.Create(ctx, creatorID)
}

func (s *chatService) AddMember(ctx context.Context, chatID, userID string) error {
	return s.chats.AddMember(ctx, chatID, userID)
}
