package dto

import "time"

// OrderStatusMessage representa a mensagem recebida da fila RabbitMQ order-status
type OrderStatusMessage struct {
	OrderID   string    `json:"order_id" validate:"required"`
	Status    string    `json:"status" validate:"required,oneof=criada criado em_processamento enviado entregue"`
	Timestamp time.Time `json:"timestamp"`
}
