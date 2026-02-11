package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	Quantity    int                `bson:"quantity"`
	Price       float64            `bson:"price"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}
