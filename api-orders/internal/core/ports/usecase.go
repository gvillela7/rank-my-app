package ports

import (
	"context"

	"github.com/gvillela7/rank-my-app/internal/core/dto"
)

type ProductUseCase interface {
	CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
}

type OrderUseCase interface {
	CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetOrderByID(ctx context.Context, id string) (*dto.OrderResponse, error)
	UpdateOrderStatus(ctx context.Context, id string, req *dto.UpdateOrderStatusRequest) (*dto.OrderResponse, error)
}
