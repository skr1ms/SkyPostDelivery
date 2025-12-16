package usecase

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
)

type VideoHandler interface {
	HandleVideoFrame(ctx context.Context, droneID string, frameData []byte, deliveryID string) error
}

type DroneDeliveryUseCase struct {
	deliveryRepo    repo.DeliveryRepo
	droneManager    *DroneManagerUseCase
	deliveryUseCase *DeliveryUseCase
	videoHandler    VideoHandler
	logger          logger.Interface
}

func NewDroneDeliveryUseCase(
	deliveryRepo repo.DeliveryRepo,
	droneManager *DroneManagerUseCase,
	deliveryUseCase *DeliveryUseCase,
	videoHandler VideoHandler,
	logger logger.Interface,
) *DroneDeliveryUseCase {
	return &DroneDeliveryUseCase{
		deliveryRepo:    deliveryRepo,
		droneManager:    droneManager,
		deliveryUseCase: deliveryUseCase,
		videoHandler:    videoHandler,
		logger:          logger,
	}
}

func (uc *DroneDeliveryUseCase) ProcessDeliveryUpdate(ctx context.Context, droneID string, payload map[string]any) error {
	var dup entity.DeliveryUpdatePayload
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &dup,
	})
	if err := decoder.Decode(payload); err != nil {
		return fmt.Errorf("DroneDeliveryUseCase - ProcessDeliveryUpdate - Decode: %w", err)
	}

	droneStatus := dup.DroneStatus
	orderID := dup.OrderID
	parcelAutomatID := dup.ParcelAutomatID

	if droneStatus == "arrived_at_destination" {
		if orderID == "" || parcelAutomatID == "" {
			return fmt.Errorf("DroneDeliveryUseCase - ProcessDeliveryUpdate - Validate[arrived_at_destination]: %w", entityError.ErrMissingRequiredField)
		}

		if uc.deliveryUseCase == nil {
			uc.logger.Warn("DroneDeliveryUseCase - ProcessDeliveryUpdate - deliveryUseCase not configured", nil, map[string]any{
				"droneID": droneID,
			})
			return nil
		}

		_, err := uc.deliveryUseCase.HandleDroneArrived(ctx, droneID, orderID, parcelAutomatID)
		if err != nil {
			uc.logger.Error("DroneDeliveryUseCase - ProcessDeliveryUpdate - HandleDroneArrived", err, map[string]any{
				"droneID":         droneID,
				"orderID":         orderID,
				"parcelAutomatID": parcelAutomatID,
			})
			return fmt.Errorf("DroneDeliveryUseCase - ProcessDeliveryUpdate - HandleDroneArrived: %w", err)
		}
		return nil
	}

	deliveryID := dup.DeliveryID
	if deliveryID == "" {
		return nil
	}

	switch droneStatus {
	case "arrived_at_locker":
		if err := uc.deliveryRepo.UpdateDeliveryStatus(ctx, deliveryID, entity.DeliveryStatusInProgress, nil); err != nil {
			uc.logger.Error("DroneDeliveryUseCase - ProcessDeliveryUpdate - UpdateDeliveryStatus", err, map[string]any{
				"deliveryID": deliveryID,
				"droneID":    droneID,
			})
			return fmt.Errorf("DroneDeliveryUseCase - ProcessDeliveryUpdate - UpdateDeliveryStatus: %w", err)
		}
	case "returning":
		if err := uc.droneManager.ReleaseDrone(ctx, droneID); err != nil {
			uc.logger.Error("DroneDeliveryUseCase - ProcessDeliveryUpdate - ReleaseDrone", err, map[string]any{
				"droneID": droneID,
			})
			return fmt.Errorf("DroneDeliveryUseCase - ProcessDeliveryUpdate - ReleaseDrone: %w", err)
		}
	}

	return nil
}

func (uc *DroneDeliveryUseCase) ProcessArrivedAtDestination(ctx context.Context, droneID string, payload map[string]any) error {
	var aad entity.ArrivedAtDestinationPayload
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &aad,
	})
	if err := decoder.Decode(payload); err != nil {
		return fmt.Errorf("DroneDeliveryUseCase - ProcessArrivedAtDestination - Decode: %w", err)
	}

	orderID := aad.OrderID
	parcelAutomatID := aad.ParcelAutomatID

	if orderID == "" || parcelAutomatID == "" {
		return fmt.Errorf("DroneDeliveryUseCase - ProcessArrivedAtDestination - Validate: %w", entityError.ErrMissingRequiredField)
	}

	if uc.deliveryUseCase == nil {
		uc.logger.Warn("DroneDeliveryUseCase - ProcessArrivedAtDestination - deliveryUseCase not configured", nil, map[string]any{
			"droneID": droneID,
		})
		return nil
	}

	_, err := uc.deliveryUseCase.HandleDroneArrived(ctx, droneID, orderID, parcelAutomatID)
	if err != nil {
		uc.logger.Error("DroneDeliveryUseCase - ProcessArrivedAtDestination - HandleDroneArrived", err, map[string]any{
			"droneID":         droneID,
			"orderID":         orderID,
			"parcelAutomatID": parcelAutomatID,
		})
		return fmt.Errorf("DroneDeliveryUseCase - ProcessArrivedAtDestination - HandleDroneArrived: %w", err)
	}
	return nil
}

func (uc *DroneDeliveryUseCase) ProcessCargoDropped(ctx context.Context, payload map[string]any) error {
	var cd entity.CargoDroppedPayload
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &cd,
	})
	if err := decoder.Decode(payload); err != nil {
		return fmt.Errorf("DroneDeliveryUseCase - ProcessCargoDropped - Decode: %w", err)
	}

	orderID := cd.OrderID
	lockerCellID := cd.LockerCellID

	if orderID == "" {
		return fmt.Errorf("DroneDeliveryUseCase - ProcessCargoDropped - Validate: %w", entityError.ErrMissingRequiredField)
	}

	if uc.deliveryUseCase == nil {
		uc.logger.Warn("DroneDeliveryUseCase - ProcessCargoDropped - deliveryUseCase not configured", nil, map[string]any{
			"orderID": orderID,
		})
		return nil
	}

	_, err := uc.deliveryUseCase.HandleCargoDropped(ctx, orderID, lockerCellID)
	if err != nil {
		uc.logger.Error("DroneDeliveryUseCase - ProcessCargoDropped - HandleCargoDropped", err, map[string]any{
			"orderID":      orderID,
			"lockerCellID": lockerCellID,
		})
		return fmt.Errorf("DroneDeliveryUseCase - ProcessCargoDropped - HandleCargoDropped: %w", err)
	}
	return nil
}

func (uc *DroneDeliveryUseCase) ProcessVideoFrame(ctx context.Context, droneID string, frameData []byte, deliveryID string) error {
	if uc.videoHandler == nil {
		uc.logger.Warn("DroneDeliveryUseCase - ProcessVideoFrame - videoHandler not configured", nil, map[string]any{
			"droneID":    droneID,
			"deliveryID": deliveryID,
		})
		return nil
	}

	if err := uc.videoHandler.HandleVideoFrame(ctx, droneID, frameData, deliveryID); err != nil {
		uc.logger.Error("DroneDeliveryUseCase - ProcessVideoFrame - HandleVideoFrame", err, map[string]any{
			"droneID":    droneID,
			"deliveryID": deliveryID,
		})
		return fmt.Errorf("DroneDeliveryUseCase - ProcessVideoFrame - HandleVideoFrame: %w", err)
	}

	return nil
}
