package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	ProductID   string  `bson:"product_id"`
	ProductName string  `bson:"product_name"`
	Price       float64 `bson:"price"`
	Quantity    int     `bson:"quantity"`
}

type Order struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	OrderNumber string             `bson:"order_number"`
	Items       []OrderItem        `bson:"items"`
	Total       float64            `bson:"total"`
	Status      string             `bson:"status"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

func (o *Order) CalculateTotal() {
	total := 0.0
	for _, item := range o.Items {
		total += item.Price * float64(item.Quantity)
	}
	o.Total = total
}
