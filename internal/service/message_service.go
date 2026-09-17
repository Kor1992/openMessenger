package service

import (
	"context"
	"encoding/json"
	"errors"
	"messanger/internal/kafka"
	"messanger/internal/repository"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
)

type MessageService interface {
	SendMessage(
		ctx context.Context,
		chatID string,
		senderID string,
		text string,
	) (string, error)
	ListByChat(ctx context.Context, chatID string, limit, offset int) ([]repository.Message, error)
}

type messageService struct {
	messages repository.MessageRepository
	chats    repository.ChatRepository
}

func NewMessageService(
	messages repository.MessageRepository,
	chats repository.ChatRepository,
) MessageService {
	return &messageService{
		messages: messages,
		chats:    chats,
	}
}

func (s *messageService) SendMessage(
	ctx context.Context,
	chatID string,
	senderID string,
	text string,
) (string, error) {
	if chatID == "" || senderID == "" || text == "" {
		return "", ErrInvalidMessage
	}

	isMember, err := s.chats.IsMember(
		ctx,
		chatID,
		senderID,
	)
	if err != nil {
		return "", err
	}

	if !isMember {
		return "", ErrNotChatMember
	}

	messageID := uuid.New().String()

	event := kafka.MessageCreatedEvent{
		EventID:   uuid.New().String(),
		MessageID: messageID,
		ChatID:    chatID,
		SenderID:  senderID,
		Text:      text,
		CreatedAt: time.Now(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return "", err
	}

	err = s.messages.CreateMessageWithOutbox(
		ctx,
		messageID,
		chatID,
		senderID,
		text,
		payload,
	)
	if err != nil {
		return "", err
	}

	return messageID, nil
}

func (s *messageService) ListByChat(ctx context.Context, chatID string, limit, offset int) ([]repository.Message, error) {
	return s.messages.ListByChat(ctx, chatID, limit, offset)
}
