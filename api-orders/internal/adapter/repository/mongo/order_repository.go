package mongo

import (
	"context"
	"time"

	"github.com/gvillela7/rank-my-app/internal/core/domain"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type orderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) ports.OrderRepository {
	return &orderRepository{
		collection: db.Collection("orders"),
	}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	order.ID = primitive.NewObjectID()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, order)
	return err
}

func (r *orderRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Order, error) {
	var order domain.Order
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
