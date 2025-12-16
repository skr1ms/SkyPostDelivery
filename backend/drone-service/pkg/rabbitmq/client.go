package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/skr1ms/SkyPostDelivery/drone-service/config"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
)

type RabbitMQClient interface {
	Publish(ctx context.Context, queue string, message any) error
}

type DeliveryHandler interface {
	ExecuteDelivery(
		ctx context.Context,
		droneID string,
		orderID string,
		goodID string,
		parcelAutomatID string,
		arucoID int,
		coordinates string,
		weight float64,
		height float64,
		length float64,
		width float64,
		internalLockerCellID *string,
	) error
	HandleReturnTask(ctx context.Context, droneID string, deliveryID string, baseMarkerID int) error
}

type Client struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	done          chan struct{}
	mu            sync.RWMutex
	isReady       bool
	notifyConfirm chan amqp.Confirmation
	notifyReturn  chan amqp.Return
	logger        logger.Interface
}

func NewClient(cfg *config.RabbitMQ, log logger.Interface) (*Client, error) {
	client := &Client{
		done:          make(chan struct{}),
		notifyConfirm: make(chan amqp.Confirmation, 1),
		notifyReturn:  make(chan amqp.Return, 10),
		logger:        log,
	}

	conn, err := client.connect(cfg.URL)
	if err != nil {
		return nil, err
	}

	if err := client.init(conn); err != nil {
		_ = conn.Close()
		return nil, err
	}

	go client.handleReconnect(cfg.URL)
	go client.handleReturns()

	return client, nil
}

func (c *Client) connect(url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("RabbitMQClient - connect - Dial: %w", err)
	}
	return conn, nil
}

func (c *Client) init(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("RabbitMQClient - init - Channel: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		return fmt.Errorf("RabbitMQClient - init - Confirm: %w", err)
	}

	if err := c.declareQueues(ch); err != nil {
		_ = ch.Close()
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.channel = ch
	c.isReady = true
	c.notifyConfirm = ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	c.notifyReturn = ch.NotifyReturn(make(chan amqp.Return, 10))
	c.mu.Unlock()

	c.logger.Info("RabbitMQ client initialized successfully", nil, nil)
	return nil
}

func (c *Client) declareQueues(ch *amqp.Channel) error {
	queues := map[string]amqp.Table{
		"deliveries": {
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": "deliveries.dlq",
			"x-message-ttl":             int32(3600000),
			"x-max-priority":            int32(10),
		},
		"deliveries.priority": {
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": "deliveries.dlq",
			"x-message-ttl":             int32(3600000),
			"x-max-priority":            int32(10),
		},
		"delivery.return": {},
	}

	for queueName, args := range queues {
		_, err := ch.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			args,
		)
		if err != nil {
			return fmt.Errorf("RabbitMQClient - declareQueues - QueueDeclare[%s]: %w", queueName, err)
		}
	}

	return nil
}

func (c *Client) handleReconnect(url string) {
	for {
		c.mu.RLock()
		isReady := c.isReady
		c.mu.RUnlock()

		if !isReady {
			c.logger.Info("Attempting to reconnect to RabbitMQ...", nil, nil)
			conn, err := c.connect(url)
			if err != nil {
				c.logger.Warn("Failed to reconnect", err, nil)
				time.Sleep(5 * time.Second)
				continue
			}

			if err := c.init(conn); err != nil {
				c.logger.Warn("Failed to initialize after reconnect", err, nil)
				_ = conn.Close()
				time.Sleep(5 * time.Second)
				continue
			}

			c.logger.Info("Successfully reconnected to RabbitMQ", nil, nil)
		}

		time.Sleep(1 * time.Second)
	}
}

func (c *Client) handleReturns() {
	for ret := range c.notifyReturn {
		c.logger.Warn("Message returned", nil, map[string]any{"body": string(ret.Body)})
	}
}

func (c *Client) Consume(ctx context.Context, queue string, handler func(context.Context, amqp.Delivery) error) error {
	c.mu.RLock()
	ch := c.channel
	c.mu.RUnlock()

	if ch == nil {
		return fmt.Errorf("RabbitMQClient - Consume - ChannelNotInitialized")
	}

	msgs, err := ch.Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("RabbitMQClient - Consume - Consume: %w", err)
	}

	go func() {
		for d := range msgs {
			if err := handler(ctx, d); err != nil {
				c.logger.Error("Error handling message", err, nil)
				_ = d.Nack(false, true)
			} else {
				_ = d.Ack(false)
			}
		}
	}()

	return nil
}

func (c *Client) Publish(ctx context.Context, queue string, message any) error {
	c.mu.RLock()
	ch := c.channel
	isReady := c.isReady
	c.mu.RUnlock()

	if !isReady {
		return fmt.Errorf("RabbitMQClient - Publish - ClientNotReady")
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("RabbitMQClient - Publish - Marshal: %w", err)
	}

	return ch.PublishWithContext(
		ctx,
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (c *Client) Close() error {
	close(c.done)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil {
		_ = c.channel.Close()
	}

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}
