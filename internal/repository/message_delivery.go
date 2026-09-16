package repository

import "context"

type MessageDeliveryRepository interface {
	Create(ctx context.Context, messageID, userID string) error
}
