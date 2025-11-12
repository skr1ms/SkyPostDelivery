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
			return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
		}
		log.Printf("Queue declared: %s", queueName)
	}

	return nil
}

func (c *Client) Publish(ctx context.Context, queueName string, message interface{}) error {
	c.mu.RLock()
	if !c.isReady {
		c.mu.RUnlock()
		return fmt.Errorf("rabbitmq client is not ready")
	}
	ch := c.channel
	confirms := c.notifyConfirm
	c.mu.RUnlock()

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
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
		return fmt.Errorf("failed to publish message: %w", err)
	}

	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return fmt.Errorf("message was nacked by RabbitMQ")
		}
		log.Printf("Message published to %s and confirmed (tag: %d)", queueName, confirm.DeliveryTag)
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while waiting for confirm: %w", ctx.Err())
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for publish confirmation")
	}
}

func (c *Client) Consume(queueName string, handler func([]byte) error) error {
	c.mu.RLock()
	if !c.isReady {
		c.mu.RUnlock()
		return fmt.Errorf("rabbitmq client is not ready")
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
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-c.done:
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Consumer channel closed, stopping consumer")
					return
				}

				if err := handler(msg.Body); err != nil {
					log.Printf("Error handling message from %s: %v", queueName, err)
					msg.Nack(false, true)
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	log.Printf("Consumer registered for queue: %s", queueName)
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
			log.Printf("Message returned from RabbitMQ: queue=%s, reason=%d, reply_text=%s",
				ret.RoutingKey, ret.ReplyCode, ret.ReplyText)
		}
	}
}

func (c *Client) handleReconnect(url string) {
	for {
		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		closeCh := make(chan *amqp.Error)
		conn.NotifyClose(closeCh)

		select {
		case <-c.done:
			return
		case err := <-closeCh:
			if err != nil {
				log.Printf("RabbitMQ connection closed: %v. Attempting to reconnect...", err)

				c.mu.Lock()
				c.isReady = false
				c.mu.Unlock()

				for {
					select {
					case <-c.done:
						return
					case <-time.After(5 * time.Second):
						log.Println("Attempting to reconnect to RabbitMQ...")

						conn, err := c.connect(url)
						if err != nil {
							log.Printf("Failed to reconnect: %v", err)
							continue
						}

						if err := c.init(conn); err != nil {
							log.Printf("Failed to initialize after reconnect: %v", err)
							conn.Close()
							continue
						}

						log.Println("Reconnected to RabbitMQ successfully")
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

	log.Println("RabbitMQ client closed")
	return err
}
