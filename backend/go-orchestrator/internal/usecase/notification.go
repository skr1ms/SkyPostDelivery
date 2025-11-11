package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/hitech-ekb/pkg/push"
)

type NotificationUseCase struct {
	deviceRepo DeviceRepo
	sender     push.Sender
}

type DeliveryNotifier interface {
	NotifyOrderDelivered(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, lockerCellID *uuid.UUID) error
}

func NewNotificationUseCase(deviceRepo DeviceRepo, sender push.Sender) *NotificationUseCase {
	return &NotificationUseCase{
		deviceRepo: deviceRepo,
		sender:     sender,
	}
}

func (uc *NotificationUseCase) RegisterDevice(ctx context.Context, userID uuid.UUID, token, platform string) error {
	if token == "" {
		return nil
	}
	if err := uc.deviceRepo.Upsert(ctx, userID, token, platform); err != nil {
		return fmt.Errorf("NotificationUseCase - RegisterDevice - Upsert: %w", err)
	}
	return nil
}

func (uc *NotificationUseCase) NotifyOrderDelivered(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, lockerCellID *uuid.UUID) error {
	devices, err := uc.deviceRepo.ListByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("NotificationUseCase - NotifyOrderDelivered - ListByUserID: %w", err)
	}
	if len(devices) == 0 {
		return nil
	}

	tokens := make([]string, 0, len(devices))
	for _, device := range devices {
		if device.Token != "" {
			tokens = append(tokens, device.Token)
		}
	}
	if len(tokens) == 0 {
		return nil
	}

	payload := push.DeliveryPayload{
		OrderID: orderID.String(),
	}
	if lockerCellID != nil {
		id := lockerCellID.String()
		payload.LockerCellID = &id
	}

	invalidTokens, err := uc.sender.SendDeliveryNotification(ctx, tokens, payload)
	if err != nil {
		return fmt.Errorf("NotificationUseCase - NotifyOrderDelivered - SendDeliveryNotification: %w", err)
	}

	for _, invalid := range invalidTokens {
		_ = uc.deviceRepo.DeleteByToken(ctx, invalid)
	}

	return nil
}
