package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=3" example:"Mouse Gamer"`
	Description string  `json:"description" validate:"required" example:"Mouse Gamer RGB 16000 DPI"`
	Quantity    int     `json:"quantity" validate:"required,gte=0" example:"50"`
	Price       float64 `json:"price" validate:"required,gt=0" example:"199.90"`
}

type ProductResponse struct {
	ID          string    `json:"_id" example:"507f1f77bcf86cd799439011"`
	Name        string    `json:"name" example:"Mouse Gamer"`
	Description string    `json:"description" example:"Mouse Gamer RGB 16000 DPI"`
	Quantity    int       `json:"quantity" example:"50"`
	Price       float64   `json:"price" example:"199.90"`
	CreatedAt   time.Time `json:"created_at" example:"2024-02-10T12:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-02-10T12:00:00Z"`
}

func ToProductResponse(id primitive.ObjectID, name, description string, quantity int, price float64, createdAt, updatedAt time.Time) *ProductResponse {
	return &ProductResponse{
		ID:          id.Hex(),
		Name:        name,
		Description: description,
		Quantity:    quantity,
		Price:       price,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
