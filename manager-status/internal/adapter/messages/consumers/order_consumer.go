package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gvillela7/rank-my-app/internal/core/dto"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
	"github.com/gvillela7/rank-my-app/internal/infra/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	exchangeName = "orders"
	exchangeType = "direct"
	queueName    = "order-status"
	routingKey   = "order-status"
	consumerTag  = "manager-status-consumer"
)

type orderConsumer struct {
	rabbitMQConn *rabbitmq.RabbitMQConnection
	useCase      ports.OrderUseCase
	logger       *zap.Logger
	mu           sync.Mutex
	deliveries   <-chan amqp.Delivery
}

// NewOrderConsumer creates a new instance of OrderConsumer
func NewOrderConsumer(
	rabbitMQConn *rabbitmq.RabbitMQConnection,
	useCase ports.OrderUseCase,
	logger *zap.Logger,
) ports.MessageConsumer {
	return &orderConsumer{
		rabbitMQConn: rabbitMQConn,
		useCase:      useCase,
		logger:       logger,
	}
}

// ConsumeOrderStatus starts consuming messages from the order-status queue
func (c *orderConsumer) ConsumeOrderStatus(ctx context.Context) error {
	c.logger.Info("Starting order status consumer",
		zap.String("queue", queueName),
		zap.String("exchange", exchangeName),
	)

	if err := c.setupInfrastructure(); err != nil {
		return fmt.Errorf("failed to setup infrastructure: %w", err)
	}

	if err := c.startConsuming(); err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.logger.Info("Consumer is ready to process messages")
	return c.processMessages(ctx)
}

// setupInfrastructure declares exchange, queue, and bindings
func (c *orderConsumer) setupInfrastructure() error {
	channel, err := c.rabbitMQConn.GetChannel()
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

	c.logger.Info("Exchange declared", zap.String("exchange", exchangeName))

	err = channel.ExchangeDeclare(
		exchangeName+".dlx",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.logger.Warn("Failed to declare DLX exchange (may already exist)", zap.Error(err))
	} else {
		c.logger.Info("Dead Letter Exchange declared", zap.String("exchange", exchangeName+".dlx"))
	}

	_, err = channel.QueueDeclare(
		queueName+".dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.logger.Warn("Failed to declare DLQ (may already exist)", zap.Error(err))
	} else {
		c.logger.Info("Dead Letter Queue declared", zap.String("queue", queueName+".dlq"))

		err = channel.QueueBind(
			queueName+".dlq",
			"",
			exchangeName+".dlx",
			false,
			nil,
		)
		if err != nil {
			c.logger.Warn("Failed to bind DLQ to DLX", zap.Error(err))
		} else {
			c.logger.Info("Dead Letter Queue bound to DLX")
		}
	}

	args := amqp.Table{
		"x-dead-letter-exchange": exchangeName + ".dlx",
	}

	_, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	c.logger.Info("Queue declared with DLX", zap.String("queue", queueName))

	err = channel.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	c.logger.Info("Queue bound to exchange",
		zap.String("queue", queueName),
		zap.String("exchange", exchangeName),
		zap.String("routing_key", routingKey),
	)

	return nil
}

// startConsuming starts consuming messages from the queue
func (c *orderConsumer) startConsuming() error {
	channel, err := c.rabbitMQConn.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	err = channel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	deliveries, err := channel.Consume(
		queueName,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.mu.Lock()
	c.deliveries = deliveries
	c.mu.Unlock()

	c.logger.Info("Started consuming messages",
		zap.String("consumer_tag", consumerTag),
	)

	return nil
}

// processMessages processes incoming messages
func (c *orderConsumer) processMessages(ctx context.Context) error {
	c.mu.Lock()
	deliveries := c.deliveries
	c.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Context cancelled, stopping consumer")
			return ctx.Err()

		case delivery, ok := <-deliveries:
			if !ok {
				c.logger.Warn("Delivery channel closed, attempting to reconnect")

				time.Sleep(5 * time.Second)

				if err := c.startConsuming(); err != nil {
					c.logger.Error("Failed to reconnect", zap.Error(err))
					return fmt.Errorf("failed to reconnect: %w", err)
				}

				c.mu.Lock()
				deliveries = c.deliveries
				c.mu.Unlock()

				continue
			}

			c.handleMessage(ctx, delivery)
		}
	}
}

// handleMessage processes a single message
func (c *orderConsumer) handleMessage(ctx context.Context, delivery amqp.Delivery) {
	c.logger.Info("========== RECEIVED RAW MESSAGE ==========",
		zap.String("message_id", delivery.MessageId),
		zap.Time("timestamp", delivery.Timestamp),
		zap.ByteString("raw_body", delivery.Body),
	)

	var message dto.OrderStatusMessage
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		c.logger.Error("Failed to unmarshal message",
			zap.Error(err),
			zap.ByteString("body", delivery.Body),
		)
		_ = delivery.Nack(false, false)
		return
	}

	c.logger.Info("Parsed message",
		zap.String("order_id", message.OrderID),
		zap.String("status", message.Status),
		zap.Time("timestamp", message.Timestamp),
	)

	if err := c.useCase.ProcessOrderStatusMessage(ctx, &message); err != nil {
		c.logger.Error("Failed to process message",
			zap.String("order_id", message.OrderID),
			zap.Error(err),
		)

		if isOrderNotFoundError(err) {
			c.logger.Warn("Order not found, sending to DLQ",
				zap.String("order_id", message.OrderID),
			)

			_ = delivery.Nack(false, false)
		} else {
			c.logger.Warn("Temporary error, requeuing message",
				zap.String("order_id", message.OrderID),
			)
			_ = delivery.Nack(false, true)
		}
		return
	}

	if err := delivery.Ack(false); err != nil {
		c.logger.Error("Failed to acknowledge message",
			zap.String("order_id", message.OrderID),
			zap.Error(err),
		)
		return
	}

	c.logger.Info("Message processed successfully",
		zap.String("order_id", message.OrderID),
	)
}

// Close gracefully shuts down the consumer
func (c *orderConsumer) Close() error {
	c.logger.Info("Closing order consumer")

	channel, err := c.rabbitMQConn.GetChannel()
	if err != nil {
		c.logger.Warn("Failed to get channel for canceling consumer", zap.Error(err))
		return nil
	}

	if err := channel.Cancel(consumerTag, false); err != nil {
		c.logger.Warn("Failed to cancel consumer", zap.Error(err))
	}

	c.logger.Info("Order consumer closed")
	return nil
}

// isOrderNotFoundError checks if the error is an order not found error
func isOrderNotFoundError(err error) bool {
	return err != nil && (err.Error() == "order not found" ||
		fmt.Sprint(err) == "order not found")
}
