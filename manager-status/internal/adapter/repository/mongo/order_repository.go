package mongo

import (
	"context"
	"time"

	"github.com/gvillela7/rank-my-app/internal/core/domain"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type orderRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

func NewOrderRepository(db *mongo.Database, logger *zap.Logger) ports.OrderRepository {
	return &orderRepository{
		collection: db.Collection("orders"),
		logger:     logger,
	}
}

func (r *orderRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Order, error) {
	r.logger.Info("Finding order by ID",
		zap.String("order_id", id.Hex()),
	)

	var order domain.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Warn("Order not found",
				zap.String("order_id", id.Hex()),
			)
		} else {
			r.logger.Error("Error finding order",
				zap.String("order_id", id.Hex()),
				zap.Error(err),
			)
		}
		return nil, err
	}

	r.logger.Info("Order found",
		zap.String("order_id", id.Hex()),
		zap.String("current_status", order.Status),
	)

	return &order, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	r.logger.Info("Updating order status",
		zap.String("order_id", id.Hex()),
		zap.String("new_status", status),
	)

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		r.logger.Error("Failed to update order status",
			zap.String("order_id", id.Hex()),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("Order status update result",
		zap.String("order_id", id.Hex()),
		zap.Int64("matched_count", result.MatchedCount),
		zap.Int64("modified_count", result.ModifiedCount),
	)

	if result.MatchedCount == 0 {
		r.logger.Warn("No order found to update",
			zap.String("order_id", id.Hex()),
		)
		return mongo.ErrNoDocuments
	}

	r.logger.Info("Order status updated successfully",
		zap.String("order_id", id.Hex()),
		zap.String("new_status", status),
	)

	return nil
}
