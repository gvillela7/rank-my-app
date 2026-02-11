package mongo

import (
	"context"

	config "github.com/gvillela7/rank-my-app/configs"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBConnection implements the MongoDBConnection interface
type MongoDBConnection struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewMongoDBConnection(ctx context.Context) (*MongoDBConnection, error) {
	cfg := config.GetDBMongo()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	database := client.Database(cfg.Database)

	return &MongoDBConnection{
		client:   client,
		database: database,
	}, nil
}

// IsConnected checks if the database is connected
func (m *MongoDBConnection) IsConnected(ctx context.Context) bool {
	if m.client == nil {
		return false
	}
	resul := m.client.Ping(ctx, nil) == nil
	return resul
}

func (m *MongoDBConnection) Client() (*mongo.Database, error) {
	client := m.database
	return client, nil
}

// Disconnect closes the MongoDB connection gracefully
func (m *MongoDBConnection) Disconnect(ctx context.Context) error {
	if m.client != nil {
		return m.client.Disconnect(ctx)
	}
	return nil
}
