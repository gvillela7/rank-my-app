package mongo

import (
	"context"
	"fmt"

	"github.com/gvillela7/rank-my-app/internal/core/domain"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type publishedOrderRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

func NewPublishedOrderRepository(db *mongo.Database, logger *zap.Logger) ports.PublishedOrderRepository {
	return &publishedOrderRepository{
		collection: db.Collection("published_orders"),
		logger:     logger,
	}
}

func (r *publishedOrderRepository) Create(ctx context.Context, publishedOrder *domain.PublishedOrder) error {
	r.logger.Info("Creating published order record",
		zap.String("order_id", publishedOrder.OrderID),
		zap.Bool("published", publishedOrder.Published),
		zap.String("order_status", publishedOrder.OrderStatus),
	)

	_, err := r.collection.InsertOne(ctx, publishedOrder)
	if err != nil {
		r.logger.Error("Failed to create published order record",
			zap.String("order_id", publishedOrder.OrderID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create published order record: %w", err)
	}

	r.logger.Info("Published order record created successfully",
		zap.String("order_id", publishedOrder.OrderID),
		zap.String("record_id", publishedOrder.ID.Hex()),
	)

	return nil
}
