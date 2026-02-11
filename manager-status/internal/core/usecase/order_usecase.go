package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gvillela7/rank-my-app/internal/core/domain"
	"github.com/gvillela7/rank-my-app/internal/core/dto"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type orderUseCase struct {
	repository               ports.OrderRepository
	publishedOrderRepository ports.PublishedOrderRepository
	logger                   *zap.Logger
}

func NewOrderUseCase(
	repository ports.OrderRepository,
	publishedOrderRepository ports.PublishedOrderRepository,
	logger *zap.Logger,
) ports.OrderUseCase {
	return &orderUseCase{
		repository:               repository,
		publishedOrderRepository: publishedOrderRepository,
		logger:                   logger,
	}
}

// ProcessOrderStatusMessage processes a message from the order-status queue
func (uc *orderUseCase) ProcessOrderStatusMessage(ctx context.Context, message *dto.OrderStatusMessage) error {
	uc.logger.Info("========== STARTING MESSAGE PROCESSING ==========",
		zap.String("order_id_from_message", message.OrderID),
		zap.String("status_from_message", message.Status),
		zap.Time("timestamp_from_message", message.Timestamp),
	)

	// Parse order ID
	orderID, err := primitive.ObjectIDFromHex(message.OrderID)
	if err != nil {
		uc.logger.Error("Invalid order ID in message - cannot parse to ObjectID",
			zap.String("order_id_string", message.OrderID),
			zap.Error(err),
		)
		return fmt.Errorf("invalid order ID: %w", err)
	}

	uc.logger.Info("Parsed order ID successfully",
		zap.String("order_id_hex", orderID.Hex()),
		zap.String("order_id_original", message.OrderID),
	)

	// Check if order exists
	order, err := uc.repository.FindByID(ctx, orderID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			uc.logger.Error("Order not found",
				zap.String("order_id", message.OrderID),
			)
			return errors.New("order not found")
		}
		uc.logger.Error("Failed to find order",
			zap.String("order_id", message.OrderID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to find order: %w", err)
	}

	// Determine the new status
	newStatus := message.Status

	uc.logger.Info("Checking status transformation",
		zap.String("order_id", message.OrderID),
		zap.String("message_status", message.Status),
	)

	// Transform status: "criada" or "criado" -> "em_processamento"
	if message.Status == "criada" || message.Status == "criado" {
		newStatus = "em_processamento"
		uc.logger.Info("Transforming status to 'em_processamento'",
			zap.String("order_id", message.OrderID),
			zap.String("from_status", message.Status),
		)
	} else {
		uc.logger.Info("No status transformation needed",
			zap.String("order_id", message.OrderID),
			zap.String("status", message.Status),
		)
	}

	// Update order status in database
	if err := uc.repository.UpdateStatus(ctx, orderID, newStatus); err != nil {
		uc.logger.Error("Failed to update order status",
			zap.String("order_id", message.OrderID),
			zap.String("new_status", newStatus),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update order status: %w", err)
	}

	uc.logger.Info("Order status updated successfully",
		zap.String("order_id", message.OrderID),
		zap.String("old_status", order.Status),
		zap.String("new_status", newStatus),
	)

	// Check if there's a published order record
	publishedOrder, err := uc.publishedOrderRepository.FindByOrderID(ctx, orderID)
	if err != nil {
		uc.logger.Error("Failed to find published order record",
			zap.String("order_id", message.OrderID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to find published order record: %w", err)
	}

	// Update or create published order record
	if publishedOrder == nil {
		// Create new published order record
		uc.logger.Info("Creating new published order record",
			zap.String("order_id", message.OrderID),
		)

		newPublishedOrder := &domain.PublishedOrder{
			ID:          primitive.NewObjectID(),
			OrderID:     orderID,
			Published:   true,
			OrderStatus: newStatus,
			Timestamp:   float64(message.Timestamp.UnixNano()) / 1e9,
			PublishedAt: time.Now(),
		}

		if err := uc.publishedOrderRepository.Create(ctx, newPublishedOrder); err != nil {
			uc.logger.Error("Failed to create published order record",
				zap.String("order_id", message.OrderID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to create published order record: %w", err)
		}
	} else if !publishedOrder.Published {
		// Update existing record to mark as published
		uc.logger.Info("Updating published order record",
			zap.String("order_id", message.OrderID),
			zap.Bool("old_published", publishedOrder.Published),
		)

		if err := uc.publishedOrderRepository.UpdatePublishedStatus(ctx, orderID, true); err != nil {
			uc.logger.Error("Failed to update published order status",
				zap.String("order_id", message.OrderID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to update published order status: %w", err)
		}
	} else {
		uc.logger.Info("Published order already marked as published",
			zap.String("order_id", message.OrderID),
		)
	}

	uc.logger.Info("========== MESSAGE PROCESSING COMPLETED ==========",
		zap.String("order_id", message.OrderID),
		zap.String("final_status", newStatus),
		zap.Bool("created_published_order", publishedOrder == nil),
		zap.Bool("updated_published_order", publishedOrder != nil && !publishedOrder.Published),
	)

	return nil
}
