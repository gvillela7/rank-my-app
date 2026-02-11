package ports

import (
	"context"

	"github.com/gvillela7/rank-my-app/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error
}

type PublishedOrderRepository interface {
	Create(ctx context.Context, publishedOrder *domain.PublishedOrder) error
	FindByOrderID(ctx context.Context, orderID primitive.ObjectID) (*domain.PublishedOrder, error)
	UpdatePublishedStatus(ctx context.Context, orderID primitive.ObjectID, published bool) error
}
