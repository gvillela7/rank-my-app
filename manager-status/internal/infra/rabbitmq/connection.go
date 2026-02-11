package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type RabbitMQConnection struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	url            string
	logger         *zap.Logger
	mu             sync.RWMutex
	reconnectDelay time.Duration
	maxRetries     int
	connected      bool
}

// NewRabbitMQConnection creates a new RabbitMQ connection with auto-reconnect capability
func NewRabbitMQConnection(host string, port int, username, password, vhost string, logger *zap.Logger) (*RabbitMQConnection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", username, password, host, port, vhost)

	logger.Info("Initializing RabbitMQ connection",
		zap.String("host", host),
		zap.Int("port", port),
		zap.String("username", username),
		zap.String("vhost", vhost),
		zap.String("connection_url", fmt.Sprintf("amqp://%s:***@%s:%d/%s", username, host, port, vhost)),
	)

	r := &RabbitMQConnection{
		url:            url,
		logger:         logger,
		reconnectDelay: 5 * time.Second,
		maxRetries:     5,
		connected:      false,
	}

	if err := r.connect(); err != nil {
		return nil, fmt.Errorf("failed to establish initial connection: %w", err)
	}

	go r.monitorConnection()

	return r, nil
}

// connect establishes connection and channel to RabbitMQ
func (r *RabbitMQConnection) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var err error

	r.logger.Info("Connecting to RabbitMQ...")

	r.conn, err = amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		r.conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	r.connected = true
	r.logger.Info("Successfully connected to RabbitMQ")

	return nil
}

// reconnect attempts to reconnect to RabbitMQ with retry logic
func (r *RabbitMQConnection) reconnect() {
	r.mu.Lock()
	if !r.connected {
		r.mu.Unlock()
		return
	}
	r.connected = false
	r.mu.Unlock()

	r.logger.Warn("Connection lost, attempting to reconnect...")

	for attempt := 1; attempt <= r.maxRetries; attempt++ {
		r.logger.Info("Reconnection attempt", zap.Int("attempt", attempt), zap.Int("max_retries", r.maxRetries))

		if err := r.connect(); err != nil {
			r.logger.Error("Reconnection failed",
				zap.Error(err),
				zap.Int("attempt", attempt),
			)

			if attempt < r.maxRetries {
				time.Sleep(r.reconnectDelay)
			}
		} else {
			r.logger.Info("Successfully reconnected to RabbitMQ")
			return
		}
	}

	r.logger.Error("Failed to reconnect after maximum retries", zap.Int("max_retries", r.maxRetries))
}

// monitorConnection monitors the connection and triggers reconnection if needed
func (r *RabbitMQConnection) monitorConnection() {
	for {
		r.mu.RLock()
		conn := r.conn
		r.mu.RUnlock()

		if conn == nil {
			time.Sleep(r.reconnectDelay)
			continue
		}

		// Listen for connection close events
		notifyClose := conn.NotifyClose(make(chan *amqp.Error))

		select {
		case err := <-notifyClose:
			if err != nil {
				r.logger.Error("RabbitMQ connection closed", zap.Error(err))
				r.reconnect()
			}
		}

		time.Sleep(1 * time.Second)
	}
}

// GetChannel returns the RabbitMQ channel (safe for concurrent access)
func (r *RabbitMQConnection) GetChannel() (*amqp.Channel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.connected || r.channel == nil {
		return nil, fmt.Errorf("not connected to RabbitMQ")
	}

	return r.channel, nil
}

// IsConnected checks if the connection is active
func (r *RabbitMQConnection) IsConnected(ctx context.Context) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.connected && r.conn != nil && !r.conn.IsClosed()
}

// Close gracefully closes the RabbitMQ connection
func (r *RabbitMQConnection) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger.Info("Closing RabbitMQ connection...")

	var errs []error

	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	r.connected = false

	if len(errs) > 0 {
		return fmt.Errorf("errors while closing connection: %v", errs)
	}

	r.logger.Info("RabbitMQ connection closed successfully")
	return nil
}
