package repo

import (
	"context"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
)

type (
	DroneRepo interface {
		SaveDroneState(ctx context.Context, state *entity.DroneState) error
		GetDroneState(ctx context.Context, droneID string) (*entity.DroneState, error)
		GetDroneIDByIP(ctx context.Context, ipAddress string) (string, error)
		UpdateDroneBattery(ctx context.Context, droneID string, batteryLevel float64) error
	}

	DeliveryRepo interface {
		SaveDeliveryTask(ctx context.Context, task *entity.DeliveryTask) error
		GetDeliveryTask(ctx context.Context, deliveryID string) (*entity.DeliveryTask, error)
		UpdateDeliveryStatus(ctx context.Context, deliveryID string, status entity.DeliveryStatus, errorMessage *string) error
	}
)
