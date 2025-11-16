package usecase

import (
	"context"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/grpc"
)

type (
	DroneStateRepo interface {
		SaveDroneState(ctx context.Context, state *entity.DroneState) error
		GetDroneState(ctx context.Context, droneID string) (*entity.DroneState, error)
		GetDroneIDByIP(ctx context.Context, ipAddress string) (string, error)
		UpdateDroneBattery(ctx context.Context, droneID string, batteryLevel float64) error
	}

	DeliveryTaskRepo interface {
		SaveDeliveryTask(ctx context.Context, task *entity.DeliveryTask) error
		GetDeliveryTask(ctx context.Context, deliveryID string) (*entity.DeliveryTask, error)
		UpdateDeliveryStatus(ctx context.Context, deliveryID string, status entity.DeliveryStatus, errorMessage *string) error
	}

	DroneManager interface {
		RegisterDrone(ctx context.Context, droneID string) error
		GetFreeDrone(ctx context.Context) (string, error)
		AssignDeliveryToDrone(ctx context.Context, droneID string, deliveryID string) error
		ReleaseDrone(ctx context.Context, droneID string) error
		UnregisterDrone(ctx context.Context, droneID string) error
		GetDroneState(ctx context.Context, droneID string) (*entity.DroneState, error)
		GetAllDrones() []string
		GetRegisteredDrones() []string
	}

	DroneWebSocketHandler interface {
		SendTaskToDrone(ctx context.Context, droneID string, task map[string]interface{}) error
		SendCommandToDrone(ctx context.Context, droneID string, command map[string]interface{}) error
	}

	OrchestratorGRPCClient interface {
		UpdateDeliveryStatus(ctx context.Context, deliveryID string, status string) error
		RequestCellOpen(ctx context.Context, orderID string, parcelAutomatID string) (*grpc.CellOpenResponse, error)
	}

	RabbitMQPublisher interface {
		Publish(ctx context.Context, queue string, message interface{}) error
	}
)
