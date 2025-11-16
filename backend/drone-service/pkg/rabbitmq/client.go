package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	done          chan struct{}
	mu            sync.RWMutex
	isReady       bool
	notifyConfirm chan amqp.Confirmation
	notifyReturn  chan amqp.Return
}

func NewClient(url string) (*Client, error) {
	client := &Client{
		done:          make(chan struct{}),
		notifyConfirm: make(chan amqp.Confirmation, 1),
		notifyReturn:  make(chan amqp.Return, 10),
	}

	conn, err := client.connect(url)
	if err != nil {
		return nil, err
	}

	if err := client.init(conn); err != nil {
		conn.Close()
		return nil, err
	}

	go client.handleReconnect(url)
	go client.handleReturns()

	return client, nil
}

func (c *Client) connect(url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return conn, nil
}

func (c *Client) init(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		ch.Close()
		return fmt.Errorf("failed to put channel into confirm mode: %w", err)
	}

	if err := c.declareQueues(ch); err != nil {
		ch.Close()
		return fmt.Errorf("failed to declare queues: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.channel = ch
	c.isReady = true
	c.notifyConfirm = ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	c.notifyReturn = ch.NotifyReturn(make(chan amqp.Return, 10))
	c.mu.Unlock()

	log.Println("RabbitMQ client initialized successfully")
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
			return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
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
			log.Println("Attempting to reconnect to RabbitMQ...")
			conn, err := c.connect(url)
			if err != nil {
				log.Printf("Failed to reconnect: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if err := c.init(conn); err != nil {
				log.Printf("Failed to initialize after reconnect: %v", err)
				conn.Close()
				time.Sleep(5 * time.Second)
				continue
			}

			log.Println("Successfully reconnected to RabbitMQ")
		}

		time.Sleep(1 * time.Second)
	}
}

func (c *Client) handleReturns() {
	for ret := range c.notifyReturn {
		log.Printf("Message returned: %s", string(ret.Body))
	}
}

func (c *Client) Consume(ctx context.Context, queue string, handler func(context.Context, amqp.Delivery) error) error {
	c.mu.RLock()
	ch := c.channel
	c.mu.RUnlock()

	if ch == nil {
		return fmt.Errorf("channel is not initialized")
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
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			if err := handler(ctx, d); err != nil {
				log.Printf("Error handling message: %v", err)
				d.Nack(false, true)
			} else {
				d.Ack(false)
			}
		}
	}()

	return nil
}

func (c *Client) Publish(ctx context.Context, queue string, message interface{}) error {
	c.mu.RLock()
	ch := c.channel
	isReady := c.isReady
	c.mu.RUnlock()

	if !isReady {
		return fmt.Errorf("client is not ready")
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
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
		c.channel.Close()
	}

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}
