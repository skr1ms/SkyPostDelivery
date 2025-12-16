package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
)

type NotificationUseCase struct {
	deviceRepo repo.DeviceRepo
	sender     repo.Sender
	logger     logger.Interface
}

type DeliveryNotifier interface {
	NotifyOrderDelivered(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, lockerCellID *uuid.UUID) error
}

func NewNotificationUseCase(deviceRepo repo.DeviceRepo, sender repo.Sender, logger logger.Interface) *NotificationUseCase {
	return &NotificationUseCase{
		deviceRepo: deviceRepo,
		sender:     sender,
		logger:     logger,
	}
}

func (uc *NotificationUseCase) RegisterDevice(ctx context.Context, device *entity.Device) error {
	if device.Token == "" {
		return entityError.ErrNotificationInvalidToken
	}
	if err := uc.deviceRepo.Upsert(ctx, device); err != nil {
		return fmt.Errorf("NotificationUseCase - RegisterDevice - Upsert: %w", err)
	}

	uc.logger.Info("Device registered", nil, map[string]any{
		"userID":   device.UserID,
		"platform": device.Platform,
	})

	return nil
}

func (uc *NotificationUseCase) NotifyOrderDelivered(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, lockerCellID *uuid.UUID) error {
	devices, err := uc.deviceRepo.ListByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("NotificationUseCase - NotifyOrderDelivered - ListDevices: %w", err)
	}

	if len(devices) == 0 {
		uc.logger.Debug("No devices registered for user", nil, map[string]any{
			"userID":  userID,
			"orderID": orderID,
		})
		return nil
	}

	tokens := make([]string, 0, len(devices))
	for _, device := range devices {
		if device.Token != "" {
			tokens = append(tokens, device.Token)
		}
	}

	if len(tokens) == 0 {
		uc.logger.Warn("All device tokens are empty for user", nil, map[string]any{
			"userID":       userID,
			"orderID":      orderID,
			"devicesCount": len(devices),
		})
		return nil
	}

	var lockerCellIDString *string
	if lockerCellID != nil {
		str := lockerCellID.String()
		lockerCellIDString = &str
	}

	invalidTokens, err := uc.sender.SendDeliveryNotification(ctx, tokens, orderID.String(), lockerCellIDString)
	if err != nil {
		uc.logger.Error("NotificationUseCase - NotifyOrderDelivered - SendNotification", err, map[string]any{
			"userID":      userID,
			"orderID":     orderID,
			"tokensCount": len(tokens),
		})
		return fmt.Errorf("NotificationUseCase - NotifyOrderDelivered - SendNotification: %w", err)
	}

	uc.logger.Info("Delivery notification sent", nil, map[string]any{
		"userID":        userID,
		"orderID":       orderID,
		"sentTo":        len(tokens),
		"invalidTokens": len(invalidTokens),
	})

	if len(invalidTokens) > 0 {
		for _, invalidToken := range invalidTokens {
			if err := uc.deviceRepo.DeleteByToken(ctx, invalidToken); err != nil {
				uc.logger.Warn("NotificationUseCase - NotifyOrderDelivered - DeleteInvalidToken", err, map[string]any{
					"token": invalidToken,
				})
			}
		}
	}

	return nil
}
