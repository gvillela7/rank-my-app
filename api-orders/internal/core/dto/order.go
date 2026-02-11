package dto

import (
	"time"

	"github.com/gvillela7/rank-my-app/internal/core/domain"
)

// OrderItemRequest represents an item in the order creation request
type OrderItemRequest struct {
	ProductID string `json:"product_id" validate:"required" example:"698c0a0893c94ce530171bbb"`
	Quantity  int    `json:"quantity" validate:"required,gte=1" example:"6"`
}

// CreateOrderRequest represents the request body for creating an order
type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

// UpdateOrderStatusRequest represents the request body for updating order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=criado em_processamento enviado entregue" example:"enviado"`
}

// OrderItemResponse represents an item in the order response
type OrderItemResponse struct {
	ProductID   string  `json:"product_id" example:"507f1f77bcf86cd799439011"`
	ProductName string  `json:"product_name" example:"Mouse Gamer"`
	Price       float64 `json:"price" example:"199.90"`
	Quantity    int     `json:"quantity" example:"2"`
}

// OrderResponse represents the response body for order operations
type OrderResponse struct {
	ID          string              `json:"_id" example:"507f1f77bcf86cd799439011"`
	OrderNumber string              `json:"order_number" example:"ORD-A1B2C3D4"`
	Items       []OrderItemResponse `json:"items"`
	Total       float64             `json:"total" example:"399.80"`
	Status      string              `json:"status" example:"criado"`
	CreatedAt   time.Time           `json:"created_at" example:"2024-02-10T12:00:00Z"`
	UpdatedAt   time.Time           `json:"updated_at" example:"2024-02-10T12:00:00Z"`
}

// ToOrderResponse converts a domain Order to OrderResponse
func ToOrderResponse(order *domain.Order) *OrderResponse {
	items := make([]OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		}
	}

	return &OrderResponse{
		ID:          order.ID.Hex(),
		OrderNumber: order.OrderNumber,
		Items:       items,
		Total:       order.Total,
		Status:      order.Status,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}
