package repository

import "context"

type MessageRepository interface {
	CreateMessage(
		ctx context.Context,
		chatID string,
		senderID string,
		text string,
	) (string, error)

	CreateMessageWithOutbox(
		ctx context.Context,
		messageID string,
		chatID string,
		senderID string,
		text string,
		payload []byte,
	) error
}
