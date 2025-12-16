package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/rabbitmq"
)

type DeliveryUseCase struct {
	deliveryRepo       repo.DeliveryRepo
	orderRepo          repo.OrderRepo
	lockerRepo         repo.LockerRepo
	internalLockerRepo repo.InternalLockerRepo
	rabbitmqClient     rabbitmq.RabbitMQClient
	notifier           DeliveryNotifier
	logger             logger.Interface
}

func NewDeliveryUseCase(
	deliveryRepo repo.DeliveryRepo,
	orderRepo repo.OrderRepo,
	lockerRepo repo.LockerRepo,
	internalLockerRepo repo.InternalLockerRepo,
	rabbitmqClient rabbitmq.RabbitMQClient,
	notifier DeliveryNotifier,
	logger logger.Interface,
) *DeliveryUseCase {
	return &DeliveryUseCase{
		deliveryRepo:       deliveryRepo,
		orderRepo:          orderRepo,
		lockerRepo:         lockerRepo,
		internalLockerRepo: internalLockerRepo,
		rabbitmqClient:     rabbitmqClient,
		notifier:           notifier,
		logger:             logger,
	}
}

var validDeliveryStatuses = map[string]bool{
	"pending":        true,
	"awaiting_drone": true,
	"in_transit":     true,
	"delivered":      true,
	"failed":         true,
	"cancelled":      true,
}

var deliveryToOrderStatus = map[string]string{
	"pending":        "in_progress",
	"awaiting_drone": "pending",
	"in_transit":     "in_progress",
	"delivered":      "delivered",
	"failed":         "failed",
	"cancelled":      "cancelled",
}

func (uc *DeliveryUseCase) StartConfirmationConsumer(ctx context.Context) {
	go func() {
		uc.logger.Info("Delivery confirmation consumer started", nil)

		err := uc.rabbitmqClient.Consume(
			rabbitmq.QueueConfirmations,
			uc.handleDeliveryConfirmation,
		)

		if err != nil {
			uc.logger.Error("DeliveryUseCase - StartConfirmationConsumer - Consume", err, nil)
		}

		<-ctx.Done()
		uc.logger.Info("Delivery confirmation consumer stopped", nil)
	}()
}

func (uc *DeliveryUseCase) handleDeliveryConfirmation(body []byte) error {
	var confirmation rabbitmq.DeliveryConfirmation
	if err := json.Unmarshal(body, &confirmation); err != nil {
		uc.logger.Error("DeliveryUseCase - handleDeliveryConfirmation - Unmarshal", err, map[string]any{
			"body": string(body),
		})
		return fmt.Errorf("DeliveryUseCase - handleDeliveryConfirmation - Unmarshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uc.logger.Info("Processing delivery confirmation", nil, map[string]any{
		"orderID":      confirmation.OrderID,
		"lockerCellID": confirmation.LockerCellID,
	})

	return uc.ConfirmGoodsLoaded(ctx, confirmation.OrderID, confirmation.LockerCellID)
}

func (uc *DeliveryUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Delivery, error) {
	delivery, err := uc.deliveryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("DeliveryUseCase - GetByID: %w", err)
	}
	return delivery, nil
}

func (uc *DeliveryUseCase) UpdateStatus(ctx context.Context, deliveryID uuid.UUID, status string) error {
	if !validDeliveryStatuses[status] {
		return entityError.ErrDeliveryInvalidStatus
	}

	delivery, err := uc.deliveryRepo.GetByID(ctx, deliveryID)
	if err != nil {
		return fmt.Errorf("DeliveryUseCase - UpdateStatus - GetByID: %w", err)
	}

	delivery.Status = status
	updatedDelivery, err := uc.deliveryRepo.UpdateStatus(ctx, delivery)
	if err != nil {
		return fmt.Errorf("DeliveryUseCase - UpdateStatus: %w", err)
	}
	delivery = updatedDelivery

	orderStatus, ok := deliveryToOrderStatus[status]
	if !ok {
		orderStatus = status
	}

	order, err := uc.orderRepo.GetByID(ctx, delivery.OrderID)
	if err != nil {
		return fmt.Errorf("DeliveryUseCase - UpdateStatus - GetByID: %w", err)
	}

	order.Status = orderStatus
	updatedOrder, err := uc.orderRepo.UpdateStatus(ctx, order)
	if err != nil {
		return fmt.Errorf("DeliveryUseCase - UpdateStatus - update order status: %w", err)
	}
	order = updatedOrder

	uc.logger.Info("Delivery status updated", nil, map[string]any{
		"deliveryID":     deliveryID,
		"deliveryStatus": status,
		"orderID":        delivery.OrderID,
		"orderStatus":    orderStatus,
	})

	if status == "delivered" {
		uc.notifyOrderDelivered(ctx, order.UserID, order.ID, order.LockerCellID)
	}

	return nil
}

func (uc *DeliveryUseCase) ListByStatus(ctx context.Context, status string) ([]*entity.Delivery, error) {
	deliveries, err := uc.deliveryRepo.ListByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("DeliveryUseCase - ListByStatus: %w", err)
	}
	return deliveries, nil
}

func (uc *DeliveryUseCase) ConfirmGoodsLoaded(ctx context.Context, orderID, lockerCellID uuid.UUID) error {
	delivery, err := uc.deliveryRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("DeliveryUseCase - ConfirmGoodsLoaded - GetByOrderID: %w", err)
	}

	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("DeliveryUseCase - ConfirmGoodsLoaded - GetByID: %w", err)
	}

	delivery.Status = "delivered"
	if _, err := uc.deliveryRepo.UpdateStatus(ctx, delivery); err != nil {
		return fmt.Errorf("DeliveryUseCase - ConfirmGoodsLoaded - UpdateStatus: %w", err)
	}

	order.Status = "delivered"
	updatedOrder, err := uc.orderRepo.UpdateStatus(ctx, order)
	if err != nil {
		return fmt.Errorf("DeliveryUseCase - ConfirmGoodsLoaded - update order status: %w", err)
	}

	if order.LockerCellID != nil {
		cell, err := uc.lockerRepo.GetCellByID(ctx, *order.LockerCellID)
		if err != nil {
			return fmt.Errorf("DeliveryUseCase - ConfirmGoodsLoaded - GetLockerCell: %w", err)
		}
		cell.Status = "occupied"
		if err := uc.lockerRepo.UpdateCellStatus(ctx, cell); err != nil {
			return fmt.Errorf("DeliveryUseCase - ConfirmGoodsLoaded - update locker cell status: %w", err)
		}
	}

	if delivery.InternalLockerCellID != nil {
		internalCell, err := uc.internalLockerRepo.GetCellByID(ctx, *delivery.InternalLockerCellID)
		if err != nil {
			uc.logger.Warn("DeliveryUseCase - ConfirmGoodsLoaded - GetInternalCell", err, map[string]any{
				"cellID": *delivery.InternalLockerCellID,
			})
		} else {
			internalCell.Status = "occupied"
			if err := uc.internalLockerRepo.UpdateCellStatus(ctx, internalCell); err != nil {
				uc.logger.Warn("DeliveryUseCase - ConfirmGoodsLoaded - UpdateInternalCellStatus", err, map[string]any{
					"cellID": *delivery.InternalLockerCellID,
				})
			}
		}
	}

	uc.logger.Info("Goods loaded confirmed", nil, map[string]any{
		"orderID":      orderID,
		"deliveryID":   delivery.ID,
		"lockerCellID": lockerCellID,
	})

	uc.notifyOrderDelivered(ctx, updatedOrder.UserID, updatedOrder.ID, updatedOrder.LockerCellID)

	return nil
}

func (uc *DeliveryUseCase) notifyOrderDelivered(ctx context.Context, userID, orderID uuid.UUID, lockerCellID *uuid.UUID) {
	if uc.notifier == nil {
		return
	}

	if err := uc.notifier.NotifyOrderDelivered(ctx, userID, orderID, lockerCellID); err != nil {
		uc.logger.Warn("DeliveryUseCase - notifyOrderDelivered", err, map[string]any{
			"userID":  userID,
			"orderID": orderID,
		})
	}
}
