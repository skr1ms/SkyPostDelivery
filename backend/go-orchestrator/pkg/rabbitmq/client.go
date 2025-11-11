package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan struct{}
}

func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	client := &Client{
		conn:    conn,
		channel: ch,
		done:    make(chan struct{}),
	}

	if err := client.declareQueues(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to declare queues: %w", err)
	}

	go client.handleReconnect(url)

	return client, nil
}

func (c *Client) declareQueues() error {
	queues := []string{
		QueueDeliveries,
		QueueDeliveriesPriority,
		QueueConfirmations,
		QueueDeliveriesDLQ,
	}

	for _, queueName := range queues {
		args := amqp.Table{}

		if queueName == QueueDeliveries || queueName == QueueDeliveriesPriority {
			args["x-dead-letter-exchange"] = ""
			args["x-dead-letter-routing-key"] = QueueDeliveriesDLQ
			args["x-message-ttl"] = int32(3600000)
		}

		_, err := c.channel.QueueDeclare(
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

func (c *Client) Publish(ctx context.Context, queueName string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	priority := uint8(0)
	if queueName == QueueDeliveriesPriority {
		priority = 10
	}

	err = c.channel.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
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

	return nil
}

func (c *Client) Consume(queueName string, handler func([]byte) error) error {
	msgs, err := c.channel.Consume(
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
					return
				}

				if err := handler(msg.Body); err != nil {
					log.Printf("Error handling message: %v", err)
					msg.Nack(false, true)
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	return nil
}

func (c *Client) handleReconnect(url string) {
	closeCh := make(chan *amqp.Error)
	c.conn.NotifyClose(closeCh)

	for {
		select {
		case <-c.done:
			return
		case err := <-closeCh:
			if err != nil {
				log.Printf("RabbitMQ connection closed: %v. Attempting to reconnect...", err)

				for {
					time.Sleep(5 * time.Second)
					conn, err := amqp.Dial(url)
					if err != nil {
						log.Printf("Failed to reconnect: %v", err)
						continue
					}

					ch, err := conn.Channel()
					if err != nil {
						conn.Close()
						log.Printf("Failed to open channel: %v", err)
						continue
					}

					c.conn = conn
					c.channel = ch

					if err := c.declareQueues(); err != nil {
						log.Printf("Failed to declare queues: %v", err)
						ch.Close()
						conn.Close()
						continue
					}

					log.Println("Reconnected to RabbitMQ successfully")
					closeCh = make(chan *amqp.Error)
					c.conn.NotifyClose(closeCh)
					break
				}
			}
		}
	}
}

func (c *Client) Close() error {
	close(c.done)

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return err
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}

	return nil
}
