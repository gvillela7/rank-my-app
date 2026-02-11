package ports

import "context"

type MessageProducer interface {
	PublishOrderStatus(ctx context.Context, orderID, status string, timestamp float64) error
}
