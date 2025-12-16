package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/config"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
)

type RabbitMQClient interface {
	Consume(queueName string, handler func([]byte) error) error
	Publish(ctx context.Context, queueName string, message any) error
	Close() error
}

type Client struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	done          chan struct{}
	mu            sync.RWMutex
	isReady       bool
	notifyConfirm chan amqp.Confirmation
	notifyReturn  chan amqp.Return
	consumers     map[string]func([]byte) error
	consumerMu    sync.RWMutex
	logger        logger.Interface
}

func NewClient(cfg *config.RabbitMQ, logger logger.Interface) (*Client, error) {
	client := &Client{
		done:          make(chan struct{}),
		notifyConfirm: make(chan amqp.Confirmation, 1),
		notifyReturn:  make(chan amqp.Return, 10),
		consumers:     make(map[string]func([]byte) error),
		logger:        logger,
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

	c.consumerMu.RLock()
	for queueName, handler := range c.consumers {
		if err := c.startConsumer(queueName, handler); err != nil {
			c.logger.Error("Failed to restart consumer after reconnect", err, map[string]any{
				"queue": queueName,
			})
		}
	}
	c.consumerMu.RUnlock()

	c.logger.Info("RabbitMQ client initialized successfully", nil)
	return nil
}

func (c *Client) declareQueues(ch *amqp.Channel) error {
	queues := map[string]amqp.Table{
		QueueDeliveries: {
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": QueueDeliveriesDLQ,
			"x-message-ttl":             int32(3600000),
			"x-max-priority":            int32(10),
		},
		QueueDeliveriesPriority: {
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": QueueDeliveriesDLQ,
			"x-message-ttl":             int32(3600000),
			"x-max-priority":            int32(10),
		},
		QueueConfirmations:  {},
		QueueDeliveriesDLQ:  {},
		QueueDeliveryReturn: {},
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
		c.logger.Info("Queue declared", nil, map[string]any{"queue": queueName})
	}

	return nil
}

func (c *Client) Publish(ctx context.Context, queueName string, message any) error {
	c.mu.RLock()
	if !c.isReady {
		c.mu.RUnlock()
		return ErrClientNotReady
	}
	ch := c.channel
	confirms := c.notifyConfirm
	c.mu.RUnlock()

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("RabbitMQClient - Publish - Marshal: %w", err)
	}

	priority := uint8(0)
	if queueName == QueueDeliveriesPriority {
		priority = 10
	}

	err = ch.PublishWithContext(
		ctx,
		"",
		queueName,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Priority:     priority,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("RabbitMQClient - Publish - PublishWithContext: %w", err)
	}

	select {
	case confirm, ok := <-confirms:
		if !ok {
			return ErrClientNotReady
		}
		if !confirm.Ack {
			return ErrMessageNacked
		}
		c.logger.Info("Message published and confirmed", nil, map[string]any{
			"queue": queueName,
			"tag":   confirm.DeliveryTag,
		})
		return nil
	case <-ctx.Done():
		return fmt.Errorf("RabbitMQClient - Publish - ContextCancelled: %w", ctx.Err())
	case <-time.After(5 * time.Second):
		return ErrPublishTimeout
	}
}

func (c *Client) Consume(queueName string, handler func([]byte) error) error {
	c.consumerMu.Lock()
	c.consumers[queueName] = handler
	c.consumerMu.Unlock()

	return c.startConsumer(queueName, handler)
}

func (c *Client) startConsumer(queueName string, handler func([]byte) error) error {
	c.mu.RLock()
	if !c.isReady {
		c.mu.RUnlock()
		return ErrClientNotReady
	}
	ch := c.channel
	c.mu.RUnlock()

	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("RabbitMQClient - startConsumer - Consume: %w", err)
	}

	go func() {
		for {
			select {
			case <-c.done:
				return
			case msg, ok := <-msgs:
				if !ok {
					c.logger.Warn("Consumer channel closed", nil, map[string]any{"queue": queueName})
					return
				}

				if err := handler(msg.Body); err != nil {
					c.logger.Error("Error handling message", err, map[string]any{"queue": queueName})
					_ = msg.Nack(false, true)
				} else {
					_ = msg.Ack(false)
				}
			}
		}
	}()

	c.logger.Info("Consumer registered for queue", nil, map[string]any{"queue": queueName})
	return nil
}

func (c *Client) handleReturns() {
	for {
		select {
		case <-c.done:
			return
		case ret, ok := <-c.notifyReturn:
			if !ok {
				return
			}
			c.logger.Warn("Message returned from RabbitMQ", nil, map[string]any{
				"queue":      ret.RoutingKey,
				"reply_code": ret.ReplyCode,
				"reply_text": ret.ReplyText,
			})
		}
	}
}

func (c *Client) handleReconnect(url string) {
	for {
		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			select {
			case <-c.done:
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		closeCh := conn.NotifyClose(make(chan *amqp.Error, 1))

		select {
		case <-c.done:
			return
		case err, ok := <-closeCh:
			if !ok {
				continue
			}

			if err != nil {
				c.logger.Warn("RabbitMQ connection closed, attempting to reconnect", err, nil)

				c.mu.Lock()
				c.isReady = false
				c.mu.Unlock()

				for {
					select {
					case <-c.done:
						return
					case <-time.After(5 * time.Second):
						c.logger.Info("Attempting to reconnect to RabbitMQ", nil, nil)

						conn, err := c.connect(url)
						if err != nil {
							c.logger.Error("Failed to reconnect", err, nil)
							continue
						}

						if err := c.init(conn); err != nil {
							c.logger.Error("Failed to initialize after reconnect", err, nil)
							_ = conn.Close()
							continue
						}

						c.logger.Info("Reconnected to RabbitMQ successfully", nil, nil)
						goto reconnected
					}
				}
			reconnected:
			}
		}
	}
}

func (c *Client) Close() error {
	close(c.done)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.isReady = false

	var err error
	if c.channel != nil {
		if e := c.channel.Close(); e != nil {
			err = e
		}
	}

	if c.conn != nil {
		if e := c.conn.Close(); e != nil && err == nil {
			err = e
		}
	}

	c.logger.Info("RabbitMQ client closed", nil, nil)
	return err
}
