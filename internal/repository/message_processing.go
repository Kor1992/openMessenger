package repository

import "context"

type MessageProcessor interface {
	ProcessMessageCreated(
		ctx context.Context,
		eventID string,
		messageID string,
		chatID string,
		senderID string,
	) (bool, error)
}
