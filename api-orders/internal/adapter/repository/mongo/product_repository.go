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

type productRepository struct {
	collection *mongo.Collection
}

func NewProductRepository(db *mongo.Database) ports.ProductRepository {
	return &productRepository{
		collection: db.Collection("products"),
	}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	product.ID = primitive.NewObjectID()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, product)
	return err
}

func (r *productRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Product, error) {
	var product domain.Product
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}
