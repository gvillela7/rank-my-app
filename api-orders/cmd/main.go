package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gvillela7/rank-my-app/configs"
	_ "github.com/gvillela7/rank-my-app/docs" // Import Swagger docs
	"github.com/gvillela7/rank-my-app/wire"
	"go.uber.org/zap"
)

// @title           Order Management API
// @version         1.0
// @description     API backend escalável para gerenciamento de produtos e pedidos construída com Go, Gin Framework, MongoDB e RabbitMQ seguindo Arquitetura Hexagonal (Clean Architecture).
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8000
// @BasePath  /api/v1

// @schemes   http https

// @tag.name Products
// @tag.description Operações relacionadas a produtos

// @tag.name Orders
// @tag.description Operações relacionadas a pedidos

// @tag.name Health
// @tag.description Health check da aplicação

// @accept   json
// @produce  json

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := config.Load(); err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	ctx := context.Background()

	app, cleanup, err := wire.InitializeApp(ctx)
	if err != nil {
		logger.Fatal("Failed to initialize application", zap.Error(err))
	}
	defer cleanup()

	defer func() {
		if err := app.DB.Disconnect(context.Background()); err != nil {
			logger.Error("Failed to disconnect from MongoDB", zap.Error(err))
		}
		if err := app.RabbitMQConn.Close(context.Background()); err != nil {
			logger.Error("Failed to close RabbitMQ connection", zap.Error(err))
		}
	}()

	cfg := config.GetAPIConfig()
	address := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	srv := &http.Server{
		Addr:           address,
		Handler:        app.Router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		logger.Info("Starting server",
			zap.String("address", address),
			zap.String("environment", cfg.Environment),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}
