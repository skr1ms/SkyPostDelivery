package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueDeliveries         = "deliveries"
	QueueDeliveriesPriority = "deliveries.priority"
	QueueDeliveryReturn     = "delivery.return"
)

type DeliveryWorker struct {
	client          *Client
	deliveryHandler DeliveryHandler
}

func NewDeliveryWorker(client *Client, deliveryHandler DeliveryHandler) *DeliveryWorker {
	return &DeliveryWorker{
		client:          client,
		deliveryHandler: deliveryHandler,
	}
}

func (w *DeliveryWorker) Start(ctx context.Context) error {
	log.Println("Starting delivery worker, setting up consumers...")

	if err := w.client.Consume(ctx, QueueDeliveries, w.handleDeliveryTask); err != nil {
		return fmt.Errorf("failed to consume from %s: %w", QueueDeliveries, err)
	}

	if err := w.client.Consume(ctx, QueueDeliveriesPriority, w.handleDeliveryTask); err != nil {
		return fmt.Errorf("failed to consume from %s: %w", QueueDeliveriesPriority, err)
	}

	if err := w.client.Consume(ctx, QueueDeliveryReturn, w.handleReturnTask); err != nil {
		return fmt.Errorf("failed to consume from %s: %w", QueueDeliveryReturn, err)
	}

	log.Printf("Delivery worker started successfully (consuming: %s, %s, %s)",
		QueueDeliveries, QueueDeliveriesPriority, QueueDeliveryReturn)

	return nil
}

func (w *DeliveryWorker) handleDeliveryTask(ctx context.Context, delivery amqp.Delivery) error {
	var message map[string]interface{}
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		log.Printf("Failed to unmarshal delivery message: %v", err)
		return err
	}

	log.Printf("Received delivery message: %v", message)

	droneID, _ := message["drone_id"].(string)
	orderID, _ := message["order_id"].(string)
	goodID, _ := message["good_id"].(string)
	parcelAutomatID, _ := message["parcel_automat_id"].(string)
	arucoID, _ := message["aruco_id"].(float64)
	coordinates, _ := message["coordinates"].(string)
	weight, _ := message["weight"].(float64)
	height, _ := message["height"].(float64)
	length, _ := message["length"].(float64)
	width, _ := message["width"].(float64)

	var internalCellID *string
	if internalCellIDStr, ok := message["internal_locker_cell_id"].(string); ok && internalCellIDStr != "" {
		internalCellID = &internalCellIDStr
	}

	log.Printf("Processing delivery task: drone_id=%s, order_id=%s, aruco_id=%d, coordinates=%s",
		droneID, orderID, int(arucoID), coordinates)

	if err := w.deliveryHandler.ExecuteDelivery(
		ctx,
		droneID,
		orderID,
		goodID,
		parcelAutomatID,
		int(arucoID),
		coordinates,
		weight,
		height,
		length,
		width,
		internalCellID,
	); err != nil {
		log.Printf("Failed to execute delivery: %v", err)
		return err
	}

	log.Printf("Successfully handed over delivery task for order %s", orderID)
	return nil
}

func (w *DeliveryWorker) handleReturnTask(ctx context.Context, delivery amqp.Delivery) error {
	var message map[string]interface{}
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		log.Printf("Failed to unmarshal return message: %v", err)
		return err
	}

	log.Printf("Received return message: %v", message)

	droneID, _ := message["drone_id"].(string)
	deliveryID, _ := message["delivery_id"].(string)

	arucoID := 131
	if arucoIDVal, ok := message["aruco_id"]; ok {
		switch v := arucoIDVal.(type) {
		case float64:
			arucoID = int(v)
		case int:
			arucoID = v
		}
	}

	if err := w.deliveryHandler.HandleReturnTask(ctx, droneID, deliveryID, arucoID); err != nil {
		log.Printf("Failed to handle return task: %v", err)
		return err
	}

	log.Printf("Successfully processed return task for delivery %s", deliveryID)
	return nil
}
