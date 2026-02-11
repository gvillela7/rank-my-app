//go:build wireinject
// +build wireinject

package wire

import (
	"context"
	"strconv"

	"github.com/google/wire"
	"github.com/gvillela7/rank-my-app/configs"
	"github.com/gvillela7/rank-my-app/internal/adapter/messages/consumers"
	mongoRepo "github.com/gvillela7/rank-my-app/internal/adapter/repository/mongo"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
	"github.com/gvillela7/rank-my-app/internal/core/usecase"
	dbMongo "github.com/gvillela7/rank-my-app/internal/infra/database/mongo"
	"github.com/gvillela7/rank-my-app/internal/infra/rabbitmq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type App struct {
	Consumer     ports.MessageConsumer
	DB           *dbMongo.MongoDBConnection
	RabbitMQConn *rabbitmq.RabbitMQConnection
	Logger       *zap.Logger
}

func InitializeApp(ctx context.Context) (*App, func(), error) {
	wire.Build(
		ProvideMongoConnection,
		ProvideMongoDatabase,
		ProvideRabbitMQConnection,
		ProvideLogger,
		ProvideOrderRepository,
		ProvidePublishedOrderRepository,
		ProvideOrderUseCase,
		ProvideMessageConsumer,
		ProvideApp,
	)
	return nil, nil, nil
}

func ProvideApp(
	consumer ports.MessageConsumer,
	conn *dbMongo.MongoDBConnection,
	rabbitConn *rabbitmq.RabbitMQConnection,
	logger *zap.Logger,
) *App {
	return &App{
		Consumer:     consumer,
		DB:           conn,
		RabbitMQConn: rabbitConn,
		Logger:       logger,
	}
}

func ProvideMongoConnection(ctx context.Context) (*dbMongo.MongoDBConnection, error) {
	return dbMongo.NewMongoDBConnection(ctx)
}

func ProvideMongoDatabase(conn *dbMongo.MongoDBConnection) (*mongo.Database, error) {
	return conn.Client()
}

func ProvideLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func ProvideOrderRepository(db *mongo.Database, logger *zap.Logger) ports.OrderRepository {
	return mongoRepo.NewOrderRepository(db, logger)
}

func ProvideOrderUseCase(
	repo ports.OrderRepository,
	publishedOrderRepo ports.PublishedOrderRepository,
	logger *zap.Logger,
) ports.OrderUseCase {
	return usecase.NewOrderUseCase(repo, publishedOrderRepo, logger)
}

func ProvideRabbitMQConnection(logger *zap.Logger) (*rabbitmq.RabbitMQConnection, error) {
	cfg := config.GetRabbitMQConfig()

	logger.Info("Loading RabbitMQ configuration",
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.String("username", cfg.Username),
		zap.String("vhost", cfg.VHost),
	)

	// Convert port from string to int
	port, err := strconv.Atoi(cfg.Port)
	if err != nil {
		logger.Error("Invalid RabbitMQ port configuration", zap.String("port", cfg.Port), zap.Error(err))
		port = 5672 // default RabbitMQ port
	}

	return rabbitmq.NewRabbitMQConnection(
		cfg.Host,
		port,
		cfg.Username,
		cfg.Password,
		cfg.VHost,
		logger,
	)
}

func ProvidePublishedOrderRepository(db *mongo.Database, logger *zap.Logger) ports.PublishedOrderRepository {
	return mongoRepo.NewPublishedOrderRepository(db, logger)
}

func ProvideMessageConsumer(
	rabbitConn *rabbitmq.RabbitMQConnection,
	useCase ports.OrderUseCase,
	logger *zap.Logger,
) ports.MessageConsumer {
	return consumers.NewOrderConsumer(rabbitConn, useCase, logger)
}
