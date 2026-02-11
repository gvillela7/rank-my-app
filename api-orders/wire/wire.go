//go:build wireinject
// +build wireinject

package wire

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/wire"
	"github.com/gvillela7/rank-my-app/configs"
	"github.com/gvillela7/rank-my-app/internal/adapter/http/handlers"
	"github.com/gvillela7/rank-my-app/internal/adapter/http/routes"
	"github.com/gvillela7/rank-my-app/internal/adapter/messages/producers"
	mongoRepo "github.com/gvillela7/rank-my-app/internal/adapter/repository/mongo"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
	"github.com/gvillela7/rank-my-app/internal/core/usecase"
	dbMongo "github.com/gvillela7/rank-my-app/internal/infra/database/mongo"
	"github.com/gvillela7/rank-my-app/internal/infra/rabbitmq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type App struct {
	Router      *gin.Engine
	DB          *dbMongo.MongoDBConnection
	RabbitMQConn *rabbitmq.RabbitMQConnection
}

func InitializeApp(ctx context.Context) (*App, func(), error) {
	wire.Build(
		ProvideMongoConnection,
		ProvideMongoDatabase,
		ProvideRabbitMQConnection,
		ProvideValidator,
		ProvideLogger,
		ProvideProductRepository,
		ProvideProductUseCase,
		ProvideProductHandler,
		ProvideOrderRepository,
		ProvidePublishedOrderRepository,
		ProvideMessageProducer,
		ProvideOrderUseCase,
		ProvideOrderHandler,
		ProvideHealthHandler,
		ProvideRouter,
		ProvideApp,
	)
	return nil, nil, nil
}

func ProvideApp(router *gin.Engine, conn *dbMongo.MongoDBConnection, rabbitConn *rabbitmq.RabbitMQConnection) *App {
	return &App{
		Router:      router,
		DB:          conn,
		RabbitMQConn: rabbitConn,
	}
}

func ProvideMongoConnection(ctx context.Context) (*dbMongo.MongoDBConnection, error) {
	return dbMongo.NewMongoDBConnection(ctx)
}

func ProvideMongoDatabase(conn *dbMongo.MongoDBConnection) (*mongo.Database, error) {
	return conn.Client()
}

func ProvideValidator() *validator.Validate {
	return validator.New()
}

func ProvideLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func ProvideProductRepository(db *mongo.Database) ports.ProductRepository {
	return mongoRepo.NewProductRepository(db)
}

func ProvideProductUseCase(repo ports.ProductRepository) ports.ProductUseCase {
	return usecase.NewProductUseCase(repo)
}

func ProvideProductHandler(uc ports.ProductUseCase, validator *validator.Validate, logger *zap.Logger) *handlers.ProductHandler {
	return handlers.NewProductHandler(uc, validator, logger)
}

func ProvideOrderRepository(db *mongo.Database) ports.OrderRepository {
	return mongoRepo.NewOrderRepository(db)
}

func ProvideOrderUseCase(orderRepo ports.OrderRepository, productRepo ports.ProductRepository, messageProducer ports.MessageProducer) ports.OrderUseCase {
	return usecase.NewOrderUseCase(orderRepo, productRepo, messageProducer)
}

func ProvideOrderHandler(uc ports.OrderUseCase, validator *validator.Validate, logger *zap.Logger) *handlers.OrderHandler {
	return handlers.NewOrderHandler(uc, validator, logger)
}

func ProvideHealthHandler(rabbitConn *rabbitmq.RabbitMQConnection) *handlers.HealthHandler {
	return handlers.NewHealthHandler(rabbitConn)
}

func ProvideRouter(productHandler *handlers.ProductHandler, orderHandler *handlers.OrderHandler, healthHandler *handlers.HealthHandler, logger *zap.Logger) *gin.Engine {
	cfg := config.GetAPIConfig()
	return routes.SetupRouter(&routes.RouterConfig{
		ProductHandler: productHandler,
		OrderHandler:   orderHandler,
		HealthHandler:  healthHandler,
		Logger:         logger,
		AllowOrigin:    cfg.Origin,
		Environment:    cfg.Environment,
	})
}

func ProvideRabbitMQConnection(logger *zap.Logger) (*rabbitmq.RabbitMQConnection, error) {
	cfg := config.GetRabbitMQConfig()

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

func ProvideMessageProducer(rabbitConn *rabbitmq.RabbitMQConnection, publishedOrderRepo ports.PublishedOrderRepository, logger *zap.Logger) (ports.MessageProducer, error) {
	return producers.NewOrderProducer(rabbitConn, publishedOrderRepo, logger)
}
