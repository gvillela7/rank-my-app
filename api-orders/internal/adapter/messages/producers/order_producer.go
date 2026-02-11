package producers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gvillela7/rank-my-app/internal/core/domain"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
	"github.com/gvillela7/rank-my-app/internal/infra/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	exchangeName    = "orders"
	exchangeType    = "direct"
	queueName       = "order-status"
	routingKey      = "order-status"
	dlxExchangeName = "orders.dlx"
	dlqName         = "order-status.dlq"
)

type orderProducer struct {
	rabbitConn          *rabbitmq.RabbitMQConnection
	publishedOrderRepo  ports.PublishedOrderRepository
	logger              *zap.Logger
	exchangeInitialized bool
	queueInitialized    bool
}

type OrderStatusMessage struct {
	OrderID   string  `json:"order_id"`
	Timestamp float64 `json:"ts"`
	Status    string  `json:"status"`
}

func NewOrderProducer(
	rabbitConn *rabbitmq.RabbitMQConnection,
	publishedOrderRepo ports.PublishedOrderRepository,
	logger *zap.Logger,
) (ports.MessageProducer, error) {
	producer := &orderProducer{
		rabbitConn:         rabbitConn,
		publishedOrderRepo: publishedOrderRepo,
		logger:             logger,
	}

	if err := producer.setupInfrastructure(); err != nil {
		logger.Error("Failed to setup RabbitMQ infrastructure", zap.Error(err))
	}

	return producer, nil
}

func (p *orderProducer) setupInfrastructure() error {
	channel, err := p.rabbitConn.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	err = channel.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}
	p.logger.Info("Exchange declared successfully", zap.String("exchange", exchangeName))

	err = channel.ExchangeDeclare(
		dlxExchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLX: %w", err)
	}
	p.logger.Info("DLX declared successfully", zap.String("dlx", dlxExchangeName))

	dlq, err := channel.QueueDeclare(
		dlqName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}
	p.logger.Info("DLQ declared successfully", zap.String("dlq", dlqName))

	err = channel.QueueBind(
		dlq.Name,
		"",
		dlxExchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind DLQ to DLX: %w", err)
	}
	p.logger.Info("DLQ bound to DLX successfully")

	queueArgs := amqp.Table{
		"x-dead-letter-exchange": dlxExchangeName,
	}
	queue, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		queueArgs,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	p.queueInitialized = true
	p.logger.Info("Queue declared successfully with DLX",
		zap.String("queue", queueName),
		zap.String("dlx", dlxExchangeName),
	)

	err = channel.QueueBind(
		queue.Name,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}
	p.logger.Info("Queue bound to exchange successfully",
		zap.String("queue", queueName),
		zap.String("exchange", exchangeName),
		zap.String("routing_key", routingKey),
	)

	p.exchangeInitialized = true

	return nil
}

func (p *orderProducer) PublishOrderStatus(ctx context.Context, orderID, status string, timestamp float64) error {
	p.logger.Info("Publishing order status",
		zap.String("order_id", orderID),
		zap.String("status", status),
		zap.Float64("ts", timestamp),
	)

	message := OrderStatusMessage{
		OrderID:   orderID,
		Timestamp: timestamp,
		Status:    status,
	}

	messageBody, err := json.Marshal(message)
	if err != nil {
		p.logger.Error("Failed to marshal message",
			zap.String("order_id", orderID),
			zap.Error(err),
		)

		p.savePublicationRecord(ctx, orderID, status, false, timestamp)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	published := p.publishToRabbitMQ(ctx, messageBody)

	p.savePublicationRecord(ctx, orderID, status, published, timestamp)

	if !published {
		return fmt.Errorf("failed to publish message to RabbitMQ")
	}

	p.logger.Info("Order status published successfully",
		zap.String("order_id", orderID),
		zap.String("status", status),
		zap.Float64("ts", timestamp),
	)

	return nil
}

func (p *orderProducer) publishToRabbitMQ(ctx context.Context, messageBody []byte) bool {
	if !p.exchangeInitialized || !p.queueInitialized {
		p.logger.Warn("RabbitMQ infrastructure not initialized, attempting setup")
		if err := p.setupInfrastructure(); err != nil {
			p.logger.Error("Failed to setup infrastructure", zap.Error(err))
			return false
		}
	}

	channel, err := p.rabbitConn.GetChannel()
	if err != nil {
		p.logger.Error("Failed to get RabbitMQ channel", zap.Error(err))
		return false
	}

	err = channel.PublishWithContext(
		ctx,
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         messageBody,
		},
	)

	if err != nil {
		p.logger.Error("Failed to publish message to RabbitMQ", zap.Error(err))
		return false
	}

	return true
}

func (p *orderProducer) savePublicationRecord(ctx context.Context, orderID, status string, published bool, timestamp float64) {
	publishedOrder := domain.NewPublishedOrder(orderID, status, published, timestamp)

	err := p.publishedOrderRepo.Create(ctx, publishedOrder)
	if err != nil {
		p.logger.Error("Failed to save publication record to database",
			zap.String("order_id", orderID),
			zap.Bool("published", published),
			zap.Float64("ts", timestamp),
			zap.Error(err),
		)
	}

	p.logger.Info("Publication record saved to database",
		zap.String("order_id", orderID),
		zap.Bool("published", published),
		zap.Float64("ts", timestamp),
	)
}
