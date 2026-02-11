package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PublishedOrder struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	OrderID     string             `bson:"order_id"`
	Published   bool               `bson:"published"`
	OrderStatus string             `bson:"order_status"`
	Timestamp   float64            `bson:"ts"`
	PublishedAt time.Time          `bson:"published_at"`
}

func NewPublishedOrder(orderID, orderStatus string, published bool, timestamp float64) *PublishedOrder {
	return &PublishedOrder{
		ID:          primitive.NewObjectID(),
		OrderID:     orderID,
		Published:   published,
		OrderStatus: orderStatus,
		Timestamp:   timestamp,
		PublishedAt: time.Now(),
	}
}
