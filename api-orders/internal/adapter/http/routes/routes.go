package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gvillela7/rank-my-app/internal/adapter/http/handlers"
	"github.com/gvillela7/rank-my-app/internal/adapter/http/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type RouterConfig struct {
	ProductHandler *handlers.ProductHandler
	OrderHandler   *handlers.OrderHandler
	HealthHandler  *handlers.HealthHandler
	Logger         *zap.Logger
	AllowOrigin    string
	Environment    string
}

func SetupRouter(config *RouterConfig) *gin.Engine {
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(middleware.Recovery(config.Logger))
	router.Use(middleware.Logger(config.Logger))
	router.Use(middleware.CORS(config.AllowOrigin))

	api := router.Group("/api/v1")
	{
		products := api.Group("/products")
		{
			products.POST("", config.ProductHandler.CreateProduct)
		}

		orders := api.Group("/orders")
		{
			orders.POST("", config.OrderHandler.CreateOrder)
			orders.GET("/:id", config.OrderHandler.GetOrderByID)
			orders.PATCH("/:id/status", config.OrderHandler.UpdateOrderStatus)
		}
	}

	// Health check endpoint
	router.GET("/health", config.HealthHandler.HealthCheck)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
