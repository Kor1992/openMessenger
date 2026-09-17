package service

import (
	"context"
	"errors"
	"messanger/internal/repository"
)

var (
	ErrNotChatMember = errors.New("user is not a chat member")
	ErrChatNotFound  = errors.New("chat not found")
)

type ChatService interface {
	CreateChat(ctx context.Context, creatorID string) (string, error)
	AddMember(ctx context.Context, chatID, userID, callerID string) error
	IsMember(ctx context.Context, chatID, userID string) (bool, error)
	ListByUser(ctx context.Context, userID string) ([]repository.ChatInfo, error)
	GetMemberIDs(ctx context.Context, chatID string) ([]string, error)
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

func (s *chatService) AddMember(ctx context.Context, chatID, userID, callerID string) error {
	isMember, err := s.chats.IsMember(ctx, chatID, callerID)
	if err != nil {
		return err
	}

	if !isMember {
		return ErrNotChatMember
	}

	return s.chats.AddMember(ctx, chatID, userID)
}

func (s *chatService) IsMember(ctx context.Context, chatID, userID string) (bool, error) {
	return s.chats.IsMember(ctx, chatID, userID)
}

func (s *chatService) ListByUser(ctx context.Context, userID string) ([]repository.ChatInfo, error) {
	return s.chats.ListByUser(ctx, userID)
}

func (s *chatService) GetMemberIDs(ctx context.Context, chatID string) ([]string, error) {
	return s.chats.GetMemberIDs(ctx, chatID)
}
