package repository

import "context"

type MessageRepository interface {
	CreateMessageWithOutbox(
		ctx context.Context,
		messageID string,
		chatID string,
		senderID string,
		text string,
		payload []byte,
	) error
}
