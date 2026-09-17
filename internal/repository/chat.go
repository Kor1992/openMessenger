package repository

import "context"

type ChatInfo struct {
	ID           string
	CreatedAt    string
	LastMessage  string
	LastSenderID string
	LastMsgTime  string
	MemberCount  int
}

type ChatRepository interface {
	Create(ctx context.Context, creatorID string) (string, error)
	AddMember(ctx context.Context, chatID, userID string) error
	IsMember(ctx context.Context, chatID, userID string) (bool, error)
	GetMemberIDs(ctx context.Context, chatID string) ([]string, error)
	ListByUser(ctx context.Context, userID string) ([]ChatInfo, error)
}
