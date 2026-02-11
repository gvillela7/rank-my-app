package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PublishedOrder represents a record of an order publication attempt to RabbitMQ
type PublishedOrder struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	OrderID     primitive.ObjectID `bson:"order_id"`
	Published   bool               `bson:"published"`
	OrderStatus string             `bson:"order_status"`
	Timestamp   float64            `bson:"ts"`
	PublishedAt time.Time          `bson:"published_at"`
}

// NewPublishedOrder creates a new PublishedOrder instance
func NewPublishedOrder(orderID primitive.ObjectID, orderStatus string, published bool, timestamp float64) *PublishedOrder {
	return &PublishedOrder{
		ID:          primitive.NewObjectID(),
		OrderID:     orderID,
		Published:   published,
		OrderStatus: orderStatus,
		Timestamp:   timestamp,
		PublishedAt: time.Now(),
	}
}
