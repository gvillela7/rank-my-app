package ports

import "context"

// MessageConsumer defines the interface for consuming messages from a message broker
type MessageConsumer interface {
	ConsumeOrderStatus(ctx context.Context) error

	Close() error
}
