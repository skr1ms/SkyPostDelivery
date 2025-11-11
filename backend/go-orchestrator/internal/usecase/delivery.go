package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/hitech-ekb/internal/entity"
	"github.com/skr1ms/hitech-ekb/pkg/rabbitmq"
)

type DeliveryUseCase struct {
	deliveryRepo   DeliveryRepo
	orderRepo      OrderRepo
	lockerRepo     LockerRepo
	rabbitmqClient RabbitMQClient
	notifier       DeliveryNotifier
}

func NewDeliveryUseCase(
	deliveryRepo DeliveryRepo,
	orderRepo OrderRepo,
	lockerRepo LockerRepo,
	rabbitmqClient RabbitMQClient,
	notifier DeliveryNotifier,
) *DeliveryUseCase {
	uc := &DeliveryUseCase{
		deliveryRepo:   deliveryRepo,
		orderRepo:      orderRepo,
		lockerRepo:     lockerRepo,
		rabbitmqClient: rabbitmqClient,
		notifier:       notifier,
	}

	return uc
}

func (uc *DeliveryUseCase) StartConfirmationConsumer() {
	go func() {
		err := uc.rabbitmqClient.Consume(
			rabbitmq.QueueConfirmations,
			uc.handleDeliveryConfirmation,
		)
		if err != nil {
			fmt.Printf("Failed to start delivery confirmation consumer: %v", err)
		}
	}()
}

func (uc *DeliveryUseCase) handleDeliveryConfirmation(body []byte) error {
	var confirmation rabbitmq.DeliveryConfirmation
	if err := json.Unmarshal(body, &confirmation); err != nil {
		return fmt.Errorf("failed to unmarshal confirmation: %w", err)
	}

	ctx := context.Background()
	return uc.ConfirmGoodsLoaded(ctx, confirmation.OrderID, confirmation.LockerCellID)
}

func (uc *DeliveryUseCase) GetDelivery(ctx context.Context, id uuid.UUID) (*entity.Delivery, error) {
	delivery, err := uc.deliveryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("delivery usecase - GetDelivery - deliveryRepo.GetByID: %w", err)
	}
	return delivery, nil
}

func (uc *DeliveryUseCase) UpdateStatus(ctx context.Context, deliveryID uuid.UUID, status string) error {
	delivery, err := uc.deliveryRepo.UpdateStatus(ctx, deliveryID, status)
	if err != nil {
		return fmt.Errorf("delivery usecase - UpdateStatus - deliveryRepo.UpdateStatus: %w", err)
	}

	order, err := uc.orderRepo.UpdateStatus(ctx, delivery.OrderID, status)
	if err != nil {
		return fmt.Errorf("delivery usecase - UpdateStatus - orderRepo.UpdateStatus: %w", err)
	}

	if status == "delivered" && uc.notifier != nil {
		if notifyErr := uc.notifier.NotifyOrderDelivered(ctx, order.UserID, order.ID, order.LockerCellID); notifyErr != nil {
			fmt.Printf("delivery usecase - UpdateStatus - notifier.NotifyOrderDelivered: %v\n", notifyErr)
		}
	}

	return nil
}

func (uc *DeliveryUseCase) ListByStatus(ctx context.Context, status string) ([]*entity.Delivery, error) {
	deliveries, err := uc.deliveryRepo.ListByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("delivery usecase - ListByStatus - deliveryRepo.ListByStatus: %w", err)
	}
	return deliveries, nil
}

func (uc *DeliveryUseCase) ConfirmGoodsLoaded(ctx context.Context, orderID, lockerCellID uuid.UUID) error {
	delivery, err := uc.deliveryRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("delivery usecase - ConfirmGoodsLoaded - deliveryRepo.GetByOrderID: %w", err)
	}

	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("delivery usecase - ConfirmGoodsLoaded - orderRepo.GetByID: %w", err)
	}

	if _, err := uc.deliveryRepo.UpdateStatus(ctx, delivery.ID, "delivered"); err != nil {
		return fmt.Errorf("delivery usecase - ConfirmGoodsLoaded - deliveryRepo.UpdateStatus: %w", err)
	}

	updatedOrder, err := uc.orderRepo.UpdateStatus(ctx, order.ID, "delivered")
	if err != nil {
		return fmt.Errorf("delivery usecase - ConfirmGoodsLoaded - orderRepo.UpdateStatus: %w", err)
	}

	if order.LockerCellID != nil {
		if err := uc.lockerRepo.UpdateCellStatus(ctx, *order.LockerCellID, "occupied"); err != nil {
			return fmt.Errorf("delivery usecase - ConfirmGoodsLoaded - lockerRepo.UpdateCellStatus: %w", err)
		}
	}

	if uc.notifier != nil {
		if notifyErr := uc.notifier.NotifyOrderDelivered(ctx, updatedOrder.UserID, updatedOrder.ID, updatedOrder.LockerCellID); notifyErr != nil {
			fmt.Printf("delivery usecase - ConfirmGoodsLoaded - notifier.NotifyOrderDelivered: %v\n", notifyErr)
		}
	}

	return nil
}
