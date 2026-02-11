package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/gvillela7/rank-my-app/internal/core/domain"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type publishedOrderRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// NewPublishedOrderRepository creates a new instance of PublishedOrderRepository
func NewPublishedOrderRepository(db *mongo.Database, logger *zap.Logger) ports.PublishedOrderRepository {
	return &publishedOrderRepository{
		collection: db.Collection("published_orders"),
		logger:     logger,
	}
}

// Create saves a new published order record to MongoDB
func (r *publishedOrderRepository) Create(ctx context.Context, publishedOrder *domain.PublishedOrder) error {
	r.logger.Info("Creating published order record",
		zap.String("order_id", publishedOrder.OrderID.Hex()),
		zap.Bool("published", publishedOrder.Published),
		zap.String("order_status", publishedOrder.OrderStatus),
	)

	_, err := r.collection.InsertOne(ctx, publishedOrder)
	if err != nil {
		r.logger.Error("Failed to create published order record",
			zap.String("order_id", publishedOrder.OrderID.Hex()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create published order record: %w", err)
	}

	r.logger.Info("Published order record created successfully",
		zap.String("order_id", publishedOrder.OrderID.Hex()),
		zap.String("record_id", publishedOrder.ID.Hex()),
	)

	return nil
}

// FindByOrderID retrieves a published order record by order ID
func (r *publishedOrderRepository) FindByOrderID(ctx context.Context, orderID primitive.ObjectID) (*domain.PublishedOrder, error) {
	r.logger.Info("Finding published order record by order ID",
		zap.String("order_id", orderID.Hex()),
	)

	var publishedOrder domain.PublishedOrder
	filter := bson.M{"order_id": orderID}

	err := r.collection.FindOne(ctx, filter).Decode(&publishedOrder)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Info("Published order record not found",
				zap.String("order_id", orderID.Hex()),
			)
			return nil, nil
		}

		r.logger.Error("Failed to find published order record",
			zap.String("order_id", orderID.Hex()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to find published order record: %w", err)
	}

	r.logger.Info("Published order record found",
		zap.String("order_id", orderID.Hex()),
		zap.Bool("published", publishedOrder.Published),
	)

	return &publishedOrder, nil
}

// UpdatePublishedStatus updates the published status of a published order record
func (r *publishedOrderRepository) UpdatePublishedStatus(ctx context.Context, orderID primitive.ObjectID, published bool) error {
	r.logger.Info("Updating published order status",
		zap.String("order_id", orderID.Hex()),
		zap.Bool("published", published),
	)

	filter := bson.M{"order_id": orderID}
	update := bson.M{
		"$set": bson.M{
			"published":    published,
			"published_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update published order status",
			zap.String("order_id", orderID.Hex()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update published order status: %w", err)
	}

	if result.MatchedCount == 0 {
		r.logger.Warn("No published order record found to update",
			zap.String("order_id", orderID.Hex()),
		)
		return fmt.Errorf("no published order record found for order_id: %s", orderID.Hex())
	}

	r.logger.Info("Published order status updated successfully",
		zap.String("order_id", orderID.Hex()),
		zap.Int64("modified_count", result.ModifiedCount),
	)

	return nil
}
