package repository

import "context"

type Message struct {
	ID        string
	ChatID    string
	SenderID  string
	Text      string
	CreatedAt string
}

type MessageRepository interface {
	CreateMessageWithOutbox(
		ctx context.Context,
		messageID string,
		chatID string,
		senderID string,
		text string,
		payload []byte,
	) error
	ListByChat(ctx context.Context, chatID string, limit, offset int) ([]Message, error)
}
