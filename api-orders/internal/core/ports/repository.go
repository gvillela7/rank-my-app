package ports

import (
	"context"

	"github.com/gvillela7/rank-my-app/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Product, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error
}

type PublishedOrderRepository interface {
	Create(ctx context.Context, publishedOrder *domain.PublishedOrder) error
}
