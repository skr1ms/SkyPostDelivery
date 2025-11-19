package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/grpc"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/rabbitmq"
)

type DeliveryUseCase struct {
	droneRepo              repo.DroneRepo
	deliveryRepo           repo.DeliveryRepo
	droneManager           *DroneManagerUseCase
	droneNotifier          DroneNotifier
	orchestratorGRPCClient grpc.OrchestratorGRPCClient
	rabbitmqClient         rabbitmq.RabbitMQClient
}

func NewDeliveryUseCase(
	droneRepo repo.DroneRepo,
	deliveryRepo repo.DeliveryRepo,
	droneManager *DroneManagerUseCase,
	droneNotifier DroneNotifier,
	orchestratorGRPCClient grpc.OrchestratorGRPCClient,
	rabbitmqClient rabbitmq.RabbitMQClient,
) *DeliveryUseCase {
	return &DeliveryUseCase{
		droneRepo:              droneRepo,
		deliveryRepo:           deliveryRepo,
		droneManager:           droneManager,
		droneNotifier:          droneNotifier,
		orchestratorGRPCClient: orchestratorGRPCClient,
		rabbitmqClient:         rabbitmqClient,
	}
}

func (uc *DeliveryUseCase) StartDelivery(
	ctx context.Context,
	droneID string,
	orderID string,
	goodID string,
	lockerCellID string,
	parcelAutomatID string,
	arucoID int,
	dimensions entity.GoodDimensions,
	internalLockerCellID *string,
) (map[string]interface{}, error) {
	deliveryID := uuid.New().String()

	task := &entity.DeliveryTask{
		DeliveryID:           deliveryID,
		OrderID:              orderID,
		GoodID:               goodID,
		LockerCellID:         lockerCellID,
		InternalLockerCellID: internalLockerCellID,
		ParcelAutomatID:      parcelAutomatID,
		Dimensions:           dimensions,
		CreatedAt:            time.Now(),
		DroneID:              &droneID,
		ArucoID:              &arucoID,
		Status:               entity.DeliveryStatusPending,
	}

	if err := uc.deliveryRepo.SaveDeliveryTask(ctx, task); err != nil {
		return map[string]interface{}{
			"success":     false,
			"message":     "Failed to save delivery task",
			"delivery_id": "",
		}, fmt.Errorf("delivery usecase - StartDelivery - deliveryRepo.SaveDeliveryTask: %w", err)
	}

	registeredDrones := uc.droneManager.GetRegisteredDrones()
	droneRegistered := false
	for _, id := range registeredDrones {
		if id == droneID {
			droneRegistered = true
			break
		}
	}

	if !droneRegistered {
		if err := uc.droneManager.RegisterDrone(ctx, droneID); err != nil {
			return map[string]interface{}{
				"success":     false,
				"message":     "Failed to register drone",
				"delivery_id": "",
			}, fmt.Errorf("delivery usecase - StartDelivery - droneManager.RegisterDrone: %w", err)
		}
	}

	if err := uc.droneManager.AssignDeliveryToDrone(ctx, droneID, deliveryID); err != nil {
		return map[string]interface{}{
			"success":     false,
			"message":     "Failed to assign delivery",
			"delivery_id": "",
		}, fmt.Errorf("delivery usecase - StartDelivery - droneManager.AssignDeliveryToDrone: %w", err)
	}

	go uc.executeDelivery(task)

	return map[string]interface{}{
		"success":     true,
		"message":     fmt.Sprintf("Delivery initiated with drone %s", droneID),
		"delivery_id": deliveryID,
	}, nil
}

func (uc *DeliveryUseCase) ExecuteDelivery(
	ctx context.Context,
	droneID string,
	orderID string,
	goodID string,
	parcelAutomatID string,
	arucoID int,
	coordinates string,
	weight float64,
	height float64,
	length float64,
	width float64,
	internalLockerCellID *string,
) error {
	taskData := map[string]interface{}{
		"drone_id":          droneID,
		"order_id":          orderID,
		"good_id":           goodID,
		"parcel_automat_id": parcelAutomatID,
		"aruco_id":          arucoID,
		"coordinates":       coordinates,
		"weight":            weight,
		"height":            height,
		"length":            length,
		"width":             width,
		"internal_cell_id":  internalLockerCellID,
	}

	if uc.droneNotifier != nil {
		message := map[string]interface{}{
			"type":      "delivery_task",
			"timestamp": time.Now().Format(time.RFC3339),
			"payload":   taskData,
		}
		if err := uc.droneNotifier.SendToDrone(ctx, droneID, message); err != nil {
			return fmt.Errorf("delivery usecase - ExecuteDelivery - droneNotifier.SendToDrone: %w", err)
		}
	}

	return nil
}

func (uc *DeliveryUseCase) executeDelivery(task *entity.DeliveryTask) {
	ctx := context.Background()

	if err := uc.deliveryRepo.UpdateDeliveryStatus(ctx, task.OrderID, entity.DeliveryStatusInProgress, nil); err != nil {
		_ = uc.deliveryRepo.UpdateDeliveryStatus(ctx, task.OrderID, entity.DeliveryStatusFailed, nil)
		if task.DroneID != nil {
			_ = uc.droneManager.ReleaseDrone(ctx, *task.DroneID)
		}
		return
	}

	taskData := map[string]interface{}{
		"delivery_id":       task.DeliveryID,
		"order_id":          task.OrderID,
		"good_id":           task.GoodID,
		"parcel_automat_id": task.ParcelAutomatID,
		"aruco_id":          task.ArucoID,
		"coordinates":       "",
		"internal_cell_id":  task.InternalLockerCellID,
		"dimensions": map[string]interface{}{
			"weight": task.Dimensions.Weight,
			"height": task.Dimensions.Height,
			"length": task.Dimensions.Length,
			"width":  task.Dimensions.Width,
		},
	}

	if task.DroneID == nil {
		_ = uc.deliveryRepo.UpdateDeliveryStatus(ctx, task.OrderID, entity.DeliveryStatusFailed, nil)
		return
	}

	if uc.droneNotifier != nil {
		message := map[string]interface{}{
			"type":      "delivery_task",
			"timestamp": time.Now().Format(time.RFC3339),
			"payload":   taskData,
		}
		if err := uc.droneNotifier.SendToDrone(ctx, *task.DroneID, message); err != nil {
			_ = uc.deliveryRepo.UpdateDeliveryStatus(ctx, task.OrderID, entity.DeliveryStatusFailed, nil)
			_ = uc.droneManager.ReleaseDrone(ctx, *task.DroneID)
		}
	}
}

func (uc *DeliveryUseCase) HandleReturnTask(ctx context.Context, droneID string, deliveryID string, baseMarkerID int) error {
	returnCommand := map[string]interface{}{
		"type": "return_to_base",
		"payload": map[string]interface{}{
			"delivery_id":    deliveryID,
			"base_marker_id": baseMarkerID,
		},
	}

	if uc.droneNotifier != nil {
		message := map[string]interface{}{
			"type":      "command",
			"timestamp": time.Now().Format(time.RFC3339),
			"payload":   returnCommand,
		}
		if err := uc.droneNotifier.SendToDrone(ctx, droneID, message); err != nil {
			return fmt.Errorf("delivery usecase - HandleReturnTask - droneNotifier.SendToDrone: %w", err)
		}
	}

	if err := uc.droneManager.ReleaseDrone(ctx, droneID); err != nil {
		return fmt.Errorf("delivery usecase - HandleReturnTask - droneManager.ReleaseDrone: %w", err)
	}

	return nil
}

func (uc *DeliveryUseCase) SendReturnCommand(ctx context.Context, droneID string, baseMarkerID int) error {

	returnCommand := map[string]interface{}{
		"type": "return_to_base",
		"payload": map[string]interface{}{
			"base_marker_id": baseMarkerID,
		},
	}

	if uc.droneNotifier != nil {
		message := map[string]interface{}{
			"type":      "command",
			"timestamp": time.Now().Format(time.RFC3339),
			"payload":   returnCommand,
		}
		if err := uc.droneNotifier.SendToDrone(ctx, droneID, message); err != nil {
			return fmt.Errorf("delivery usecase - SendReturnCommand - droneNotifier.SendToDrone: %w", err)
		}
	}

	return nil
}

func (uc *DeliveryUseCase) HandleDroneArrived(ctx context.Context, droneID string, orderID string, parcelAutomatID string) (map[string]interface{}, error) {
	if uc.orchestratorGRPCClient == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Orchestrator client not configured",
		}, fmt.Errorf("delivery usecase - HandleDroneArrived - orchestrator client not configured")
	}

	response, err := uc.orchestratorGRPCClient.RequestCellOpen(ctx, orderID, parcelAutomatID)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to request cell open",
		}, fmt.Errorf("delivery usecase - HandleDroneArrived - orchestratorGRPCClient.RequestCellOpen: %w", err)
	}

	if response.Success {
		command := map[string]interface{}{
			"command":          "drop_cargo",
			"order_id":         orderID,
			"cell_id":          response.CellID,
			"internal_cell_id": response.InternalCellID,
		}

		if uc.droneNotifier != nil {
			message := map[string]interface{}{
				"type":      "command",
				"timestamp": time.Now().Format(time.RFC3339),
				"payload":   command,
			}
			if err := uc.droneNotifier.SendToDrone(ctx, droneID, message); err != nil {
				return map[string]interface{}{
					"success": false,
					"message": "Failed to send drop_cargo command",
				}, fmt.Errorf("delivery usecase - HandleDroneArrived - droneNotifier.SendToDrone: %w", err)
			}
		}

		return map[string]interface{}{
			"success":          true,
			"cell_id":          response.CellID,
			"internal_cell_id": response.InternalCellID,
		}, nil
	}

	return map[string]interface{}{
		"success": false,
		"message": "Failed to open cell",
	}, fmt.Errorf("delivery usecase - HandleDroneArrived - cell open failed: %s", response.Message)
}

func (uc *DeliveryUseCase) HandleCargoDropped(ctx context.Context, orderID string, lockerCellID string) (map[string]interface{}, error) {
	task, err := uc.deliveryRepo.GetDeliveryTask(ctx, orderID)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Delivery task not found",
		}, fmt.Errorf("delivery usecase - HandleCargoDropped - deliveryTaskRepo.GetDeliveryTask: %w", err)
	}

	if task == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Delivery task not found",
		}, fmt.Errorf("delivery usecase - HandleCargoDropped - delivery task not found")
	}

	if err := uc.deliveryRepo.UpdateDeliveryStatus(ctx, orderID, entity.DeliveryStatusCompleted, nil); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to update delivery status",
		}, fmt.Errorf("delivery usecase - HandleCargoDropped - deliveryTaskRepo.UpdateDeliveryStatus: %w", err)
	}

	if task.DroneID != nil {
		_ = uc.droneManager.ReleaseDrone(ctx, *task.DroneID)
	}

	if uc.rabbitmqClient != nil {
		confirmationMessage := map[string]interface{}{
			"order_id":       orderID,
			"locker_cell_id": lockerCellID,
		}
		if lockerCellID == "" && task.LockerCellID != "" {
			confirmationMessage["locker_cell_id"] = task.LockerCellID
		}
		_ = uc.rabbitmqClient.Publish(ctx, "confirmations", confirmationMessage)
	}

	return map[string]interface{}{
		"success": true,
		"message": "Cargo dropped successfully",
	}, nil
}

func (uc *DeliveryUseCase) CancelDelivery(ctx context.Context, deliveryID string) (map[string]interface{}, error) {
	task, err := uc.deliveryRepo.GetDeliveryTask(ctx, deliveryID)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Delivery not found",
		}, fmt.Errorf("delivery usecase - CancelDelivery - deliveryTaskRepo.GetDeliveryTask: %w", err)
	}

	if task == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Delivery not found",
		}, fmt.Errorf("delivery usecase - CancelDelivery - delivery not found")
	}

	if uc.droneNotifier != nil && task.DroneID != nil {
		message := map[string]interface{}{
			"type":      "command",
			"timestamp": time.Now().Format(time.RFC3339),
			"payload": map[string]interface{}{
				"command": "cancel_delivery",
			},
		}
		_ = uc.droneNotifier.SendToDrone(ctx, *task.DroneID, message)
	}

	if err := uc.deliveryRepo.UpdateDeliveryStatus(ctx, deliveryID, entity.DeliveryStatusCancelled, nil); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to cancel delivery",
		}, fmt.Errorf("delivery usecase - CancelDelivery - deliveryTaskRepo.UpdateDeliveryStatus: %w", err)
	}

	if task.DroneID != nil {
		_ = uc.droneManager.ReleaseDrone(ctx, *task.DroneID)
	}

	return map[string]interface{}{
		"success": true,
		"message": "Delivery cancelled",
	}, nil
}

func (uc *DeliveryUseCase) GetDeliveryStatus(ctx context.Context, deliveryID string) (map[string]interface{}, error) {
	task, err := uc.deliveryRepo.GetDeliveryTask(ctx, deliveryID)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Delivery not found",
		}, fmt.Errorf("delivery usecase - GetDeliveryStatus - deliveryTaskRepo.GetDeliveryTask: %w", err)
	}

	if task == nil {
		return map[string]interface{}{
			"success": false,
			"message": "Delivery not found",
		}, fmt.Errorf("delivery usecase - GetDeliveryStatus - delivery not found")
	}

	status := "unknown"
	if task.Status != "" {
		status = string(task.Status)
	}

	droneID := ""
	if task.DroneID != nil {
		droneID = *task.DroneID
	}

	return map[string]interface{}{
		"success":     true,
		"delivery_id": task.DeliveryID,
		"status":      status,
		"drone_id":    droneID,
	}, nil
}
