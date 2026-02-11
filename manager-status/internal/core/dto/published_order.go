package dto

import "time"

// PublishedOrderResponse represents the response for a published order record
type PublishedOrderResponse struct {
	ID          string    `json:"_id" example:"507f1f77bcf86cd799439011"`
	OrderID     string    `json:"order_id" example:"ORD-20250211-001"`
	Published   bool      `json:"published" example:"true"`
	OrderStatus string    `json:"order_status" example:"criada"`
	Timestamp   float64   `json:"ts" example:"1872367127.98399"`
	PublishedAt time.Time `json:"published_at" example:"2025-02-11T10:30:00Z"`
}
