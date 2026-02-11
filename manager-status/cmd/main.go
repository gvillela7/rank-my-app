package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gvillela7/rank-my-app/configs"
	"github.com/gvillela7/rank-my-app/wire"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting Manager Status Consumer Service")

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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		logger.Info("Closing resources...")

		if err := app.Consumer.Close(); err != nil {
			logger.Error("Failed to close consumer", zap.Error(err))
		}

		if err := app.DB.Disconnect(shutdownCtx); err != nil {
			logger.Error("Failed to disconnect from MongoDB", zap.Error(err))
		}

		if err := app.RabbitMQConn.Close(shutdownCtx); err != nil {
			logger.Error("Failed to close RabbitMQ connection", zap.Error(err))
		}

		logger.Info("All resources closed")
	}()

	consumerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		logger.Info("Starting order status consumer...")
		if err := app.Consumer.ConsumeOrderStatus(consumerCtx); err != nil {
			if err == context.Canceled {
				logger.Info("Consumer stopped by context cancellation")
			} else {
				logger.Error("Consumer error", zap.Error(err))
			}
		}
	}()

	logger.Info("Manager Status Consumer is running. Press Ctrl+C to stop.")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutdown signal received, stopping consumer...")

	cancel()

	time.Sleep(5 * time.Second)

	logger.Info("Service stopped gracefully")
}
