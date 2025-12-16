package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
)

const (
	QueueDeliveries         = "deliveries"
	QueueDeliveriesPriority = "deliveries.priority"
	QueueDeliveryReturn     = "delivery.return"
)

type DeliveryWorker struct {
	client          *Client
	deliveryHandler DeliveryHandler
	logger          logger.Interface
}

func NewDeliveryWorker(client *Client, deliveryHandler DeliveryHandler, log logger.Interface) *DeliveryWorker {
	return &DeliveryWorker{
		client:          client,
		deliveryHandler: deliveryHandler,
		logger:          log,
	}
}

func (w *DeliveryWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting delivery worker, setting up consumers...", nil, nil)

	if err := w.client.Consume(ctx, QueueDeliveries, w.handleDeliveryTask); err != nil {
		return fmt.Errorf("DeliveryWorker - Start - Consume[%s]: %w", QueueDeliveries, err)
	}

	if err := w.client.Consume(ctx, QueueDeliveriesPriority, w.handleDeliveryTask); err != nil {
		return fmt.Errorf("DeliveryWorker - Start - Consume[%s]: %w", QueueDeliveriesPriority, err)
	}

	if err := w.client.Consume(ctx, QueueDeliveryReturn, w.handleReturnTask); err != nil {
		return fmt.Errorf("DeliveryWorker - Start - Consume[%s]: %w", QueueDeliveryReturn, err)
	}

	w.logger.Info("Delivery worker started successfully", nil, map[string]any{
		"queues": []string{QueueDeliveries, QueueDeliveriesPriority, QueueDeliveryReturn},
	})

	return nil
}

func (w *DeliveryWorker) handleDeliveryTask(ctx context.Context, delivery amqp.Delivery) error {
	var message map[string]any
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		w.logger.Error("Failed to unmarshal delivery message", err, nil)
		return err
	}

	w.logger.Info("Received delivery message", nil, map[string]any{"message": message})

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

	w.logger.Info("Processing delivery task", nil, map[string]any{
		"drone_id":    droneID,
		"order_id":    orderID,
		"aruco_id":    int(arucoID),
		"coordinates": coordinates,
	})

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
		w.logger.Error("Failed to execute delivery", err, nil)
		return err
	}

	w.logger.Info("Successfully handed over delivery task", nil, map[string]any{"order_id": orderID})
	return nil
}

func (w *DeliveryWorker) handleReturnTask(ctx context.Context, delivery amqp.Delivery) error {
	var message map[string]any
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		w.logger.Error("Failed to unmarshal return message", err, nil)
		return err
	}

	w.logger.Info("Received return message", nil, map[string]any{"message": message})

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
		w.logger.Error("Failed to handle return task", err, nil)
		return err
	}

	w.logger.Info("Successfully processed return task", nil, map[string]any{"delivery_id": deliveryID})
	return nil
}
