package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gvillela7/rank-my-app/internal/adapter/http/handlers"
	"github.com/gvillela7/rank-my-app/internal/core/domain"
	"github.com/gvillela7/rank-my-app/internal/core/dto"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type orderUseCase struct {
	orderRepository   ports.OrderRepository
	productRepository ports.ProductRepository
	messageProducer   ports.MessageProducer
}

func NewOrderUseCase(orderRepository ports.OrderRepository, productRepository ports.ProductRepository, messageProducer ports.MessageProducer) ports.OrderUseCase {
	return &orderUseCase{
		orderRepository:   orderRepository,
		productRepository: productRepository,
		messageProducer:   messageProducer,
	}
}

func (uc *orderUseCase) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {

	items := make([]domain.OrderItem, len(req.Items))
	for i, itemReq := range req.Items {
		productID, err := primitive.ObjectIDFromHex(itemReq.ProductID)
		if err != nil {
			return nil, handlers.NotFoundError(fmt.Sprintf("Invalid product ID: %s", itemReq.ProductID))
		}

		product, err := uc.productRepository.FindByID(ctx, productID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, handlers.NotFoundError(fmt.Sprintf("Product not found: %s", itemReq.ProductID))
			}
			return nil, fmt.Errorf("failed to fetch product: %w", err)
		}

		if product.Quantity < itemReq.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %s: available %d, requested %d",
				product.Name, product.Quantity, itemReq.Quantity)
		}

		items[i] = domain.OrderItem{
			ProductID:   itemReq.ProductID,
			ProductName: product.Name,
			Price:       product.Price,
			Quantity:    itemReq.Quantity,
		}
	}

	order := &domain.Order{
		OrderNumber: generateOrderNumber(),
		Items:       items,
		Status:      "criado",
	}

	order.CalculateTotal()

	if err := uc.orderRepository.Create(ctx, order); err != nil {
		return nil, err
	}

	timestamp := float64(time.Now().UnixNano()) / 1e9

	if err := uc.messageProducer.PublishOrderStatus(ctx, order.ID.Hex(), order.Status, timestamp); err != nil {
	}

	return dto.ToOrderResponse(order), nil
}

func (uc *orderUseCase) GetOrderByID(ctx context.Context, id string) (*dto.OrderResponse, error) {

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, handlers.NotFoundError("Invalid order ID")
	}

	order, err := uc.orderRepository.FindByID(ctx, objectID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, handlers.NotFoundError("Order not found")
		}
		return nil, err
	}

	return dto.ToOrderResponse(order), nil
}

func (uc *orderUseCase) UpdateOrderStatus(ctx context.Context, id string, req *dto.UpdateOrderStatusRequest) (*dto.OrderResponse, error) {

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, handlers.NotFoundError("Invalid order ID")
	}

	if err := uc.orderRepository.UpdateStatus(ctx, objectID, req.Status); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, handlers.NotFoundError("Order not found")
		}
		return nil, err
	}

	order, err := uc.orderRepository.FindByID(ctx, objectID)
	if err != nil {
		return nil, err
	}

	timestamp := float64(time.Now().UnixNano()) / 1e9

	if err := uc.messageProducer.PublishOrderStatus(ctx, order.ID.Hex(), order.Status, timestamp); err != nil {

	}

	return dto.ToOrderResponse(order), nil
}

func generateOrderNumber() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("ORD-%d", primitive.NewObjectID().Timestamp().Unix())
	}
	return fmt.Sprintf("ORD-%s", hex.EncodeToString(bytes))
}
