package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo"
	grpcError "github.com/skr1ms/SkyPostDelivery/drone-service/pkg/grpc"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/rabbitmq"
)

type DeliveryUseCase struct {
	droneRepo              repo.DroneRepo
	deliveryRepo           repo.DeliveryRepo
	droneManager           *DroneManagerUseCase
	droneNotifier          DroneNotifier
	orchestratorGRPCClient grpcError.OrchestratorGRPCClient
	rabbitmqClient         rabbitmq.RabbitMQClient
	logger                 logger.Interface
}

func NewDeliveryUseCase(
	droneRepo repo.DroneRepo,
	deliveryRepo repo.DeliveryRepo,
	droneManager *DroneManagerUseCase,
	droneNotifier DroneNotifier,
	orchestratorGRPCClient grpcError.OrchestratorGRPCClient,
	rabbitmqClient rabbitmq.RabbitMQClient,
	logger logger.Interface,
) *DeliveryUseCase {
	return &DeliveryUseCase{
		droneRepo:              droneRepo,
		deliveryRepo:           deliveryRepo,
		droneManager:           droneManager,
		droneNotifier:          droneNotifier,
		orchestratorGRPCClient: orchestratorGRPCClient,
		rabbitmqClient:         rabbitmqClient,
		logger:                 logger,
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
) (map[string]any, error) {
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
		uc.logger.Error("DeliveryUseCase - StartDelivery - SaveDeliveryTask", err, map[string]any{
			"deliveryID": deliveryID,
			"orderID":    orderID,
		})
		return map[string]any{
			"success":     false,
			"message":     "Failed to save delivery task",
			"delivery_id": "",
		}, fmt.Errorf("DeliveryUseCase - StartDelivery - SaveDeliveryTask: %w", err)
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
			uc.logger.Error("DeliveryUseCase - StartDelivery - RegisterDrone", err, map[string]any{
				"droneID":    droneID,
				"deliveryID": deliveryID,
			})
			return map[string]any{
				"success":     false,
				"message":     "Failed to register drone",
				"delivery_id": "",
			}, fmt.Errorf("DeliveryUseCase - StartDelivery - RegisterDrone: %w", err)
		}
	}

	if err := uc.droneManager.AssignDeliveryToDrone(ctx, droneID, deliveryID); err != nil {
		uc.logger.Error("DeliveryUseCase - StartDelivery - AssignDeliveryToDrone", err, map[string]any{
			"droneID":    droneID,
			"deliveryID": deliveryID,
		})
		return map[string]any{
			"success":     false,
			"message":     "Failed to assign delivery",
			"delivery_id": "",
		}, fmt.Errorf("DeliveryUseCase - StartDelivery - AssignDeliveryToDrone: %w", err)
	}

	go uc.executeDelivery(task)

	uc.logger.Info("Delivery initiated", nil, map[string]any{
		"deliveryID": deliveryID,
		"droneID":    droneID,
		"orderID":    orderID,
	})

	return map[string]any{
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
	taskData := map[string]any{
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
		message := map[string]any{
			"type":      "delivery_task",
			"timestamp": time.Now().Format(time.RFC3339),
			"payload":   taskData,
		}
		if err := uc.droneNotifier.SendToDrone(ctx, droneID, message); err != nil {
			uc.logger.Error("DeliveryUseCase - ExecuteDelivery - SendToDrone", err, map[string]any{
				"droneID": droneID,
				"orderID": orderID,
			})
			return fmt.Errorf("DeliveryUseCase - ExecuteDelivery - SendToDrone: %w", err)
		}
	}

	return nil
}

func (uc *DeliveryUseCase) executeDelivery(task *entity.DeliveryTask) {
	ctx := context.Background()

	if err := uc.deliveryRepo.UpdateDeliveryStatus(ctx, task.OrderID, entity.DeliveryStatusInProgress, nil); err != nil {
		uc.logger.Error("DeliveryUseCase - executeDelivery - UpdateDeliveryStatus", err, map[string]any{
			"orderID":    task.OrderID,
			"deliveryID": task.DeliveryID,
		})
		_ = uc.deliveryRepo.UpdateDeliveryStatus(ctx, task.OrderID, entity.DeliveryStatusFailed, nil)
		if task.DroneID != nil {
			_ = uc.droneManager.ReleaseDrone(ctx, *task.DroneID)
		}
		return
	}

	taskData := map[string]any{
		"delivery_id":       task.DeliveryID,
		"order_id":          task.OrderID,
		"good_id":           task.GoodID,
		"parcel_automat_id": task.ParcelAutomatID,
		"aruco_id":          task.ArucoID,
		"coordinates":       "",
		"internal_cell_id":  task.InternalLockerCellID,
		"dimensions": map[string]any{
			"weight": task.Dimensions.Weight,
			"height": task.Dimensions.Height,
			"length": task.Dimensions.Length,
			"width":  task.Dimensions.Width,
		},
	}

	if task.DroneID == nil {
		uc.logger.Warn("DeliveryUseCase - executeDelivery - droneID is nil", nil, map[string]any{
			"orderID":    task.OrderID,
			"deliveryID": task.DeliveryID,
		})
		_ = uc.deliveryRepo.UpdateDeliveryStatus(ctx, task.OrderID, entity.DeliveryStatusFailed, nil)
		return
	}

	if uc.droneNotifier != nil {
		message := map[string]any{
			"type":      "delivery_task",
			"timestamp": time.Now().Format(time.RFC3339),
			"payload":   taskData,
		}
		if err := uc.droneNotifier.SendToDrone(ctx, *task.DroneID, message); err != nil {
			uc.logger.Error("DeliveryUseCase - executeDelivery - SendToDrone", err, map[string]any{
				"droneID":    *task.DroneID,
				"orderID":    task.OrderID,
				"deliveryID": task.DeliveryID,
			})
			_ = uc.deliveryRepo.UpdateDeliveryStatus(ctx, task.OrderID, entity.DeliveryStatusFailed, nil)
			_ = uc.droneManager.ReleaseDrone(ctx, *task.DroneID)
		}
	}
}

func (uc *DeliveryUseCase) HandleReturnTask(ctx context.Context, droneID string, deliveryID string, baseMarkerID int) error {
	returnCommand := map[string]any{
		"type": "return_to_base",
		"payload": map[string]any{
			"delivery_id":    deliveryID,
			"base_marker_id": baseMarkerID,
		},
	}

	if uc.droneNotifier != nil {
		message := map[string]any{
			"type":      "command",
			"timestamp": time.Now().Format(time.RFC3339),
			"payload":   returnCommand,
		}
		if err := uc.droneNotifier.SendToDrone(ctx, droneID, message); err != nil {
			uc.logger.Error("DeliveryUseCase - HandleReturnTask - SendToDrone", err, map[string]any{
				"droneID":    droneID,
				"deliveryID": deliveryID,
			})
			return fmt.Errorf("DeliveryUseCase - HandleReturnTask - SendToDrone: %w", err)
		}
	}

	if err := uc.droneManager.ReleaseDrone(ctx, droneID); err != nil {
		uc.logger.Error("DeliveryUseCase - HandleReturnTask - ReleaseDrone", err, map[string]any{
			"droneID": droneID,
		})
		return fmt.Errorf("DeliveryUseCase - HandleReturnTask - ReleaseDrone: %w", err)
	}

	return nil
}

func (uc *DeliveryUseCase) SendReturnCommand(ctx context.Context, droneID string, baseMarkerID int) error {

	returnCommand := map[string]any{
		"type": "return_to_base",
		"payload": map[string]any{
			"base_marker_id": baseMarkerID,
		},
	}

	if uc.droneNotifier != nil {
		message := map[string]any{
			"type":      "command",
			"timestamp": time.Now().Format(time.RFC3339),
			"payload":   returnCommand,
		}
		if err := uc.droneNotifier.SendToDrone(ctx, droneID, message); err != nil {
			uc.logger.Error("DeliveryUseCase - SendReturnCommand - SendToDrone", err, map[string]any{
				"droneID": droneID,
			})
			return fmt.Errorf("DeliveryUseCase - SendReturnCommand - SendToDrone: %w", err)
		}
	}

	return nil
}

func (uc *DeliveryUseCase) HandleDroneArrived(ctx context.Context, droneID string, orderID string, parcelAutomatID string) (map[string]any, error) {
	if uc.orchestratorGRPCClient == nil {
		uc.logger.Error("DeliveryUseCase - HandleDroneArrived - orchestrator client not configured", grpcError.ErrGRPCClientNotReady, map[string]any{
			"orderID":         orderID,
			"parcelAutomatID": parcelAutomatID,
		})
		return map[string]any{
			"success": false,
			"message": "Orchestrator client not configured",
		}, grpcError.ErrGRPCClientNotReady
	}

	response, err := uc.orchestratorGRPCClient.RequestCellOpen(ctx, orderID, parcelAutomatID)
	if err != nil {
		uc.logger.Error("DeliveryUseCase - HandleDroneArrived - RequestCellOpen", err, map[string]any{
			"orderID":         orderID,
			"parcelAutomatID": parcelAutomatID,
		})
		return map[string]any{
			"success": false,
			"message": "Failed to request cell open",
		}, fmt.Errorf("DeliveryUseCase - HandleDroneArrived - RequestCellOpen: %w", err)
	}

	if response.Success {
		command := map[string]any{
			"command":          "drop_cargo",
			"order_id":         orderID,
			"cell_id":          response.CellID,
			"internal_cell_id": response.InternalCellID,
		}

		if uc.droneNotifier != nil {
			message := map[string]any{
				"type":      "command",
				"timestamp": time.Now().Format(time.RFC3339),
				"payload":   command,
			}
			if err := uc.droneNotifier.SendToDrone(ctx, droneID, message); err != nil {
				uc.logger.Warn("DeliveryUseCase - HandleDroneArrived - SendToDrone", err, map[string]any{
					"droneID": droneID,
					"orderID": orderID,
				})
			}
		}

		uc.logger.Info("Cell opened successfully", nil, map[string]any{
			"orderID":         orderID,
			"parcelAutomatID": parcelAutomatID,
			"cellID":          response.CellID,
		})

		return map[string]any{
			"success":          true,
			"cell_id":          response.CellID,
			"internal_cell_id": response.InternalCellID,
		}, nil
	}

	uc.logger.Warn("Cell open failed", nil, map[string]any{
		"orderID":         orderID,
		"parcelAutomatID": parcelAutomatID,
		"message":         response.Message,
	})

	return map[string]any{
		"success": false,
		"message": "Failed to open cell",
	}, fmt.Errorf("DeliveryUseCase - HandleDroneArrived - cell open failed: %s", response.Message)
}

func (uc *DeliveryUseCase) HandleCargoDropped(ctx context.Context, orderID string, lockerCellID string) (map[string]any, error) {
	task, err := uc.deliveryRepo.GetDeliveryTask(ctx, orderID)
	if err != nil {
		uc.logger.Error("DeliveryUseCase - HandleCargoDropped - GetDeliveryTask", err, map[string]any{
			"orderID": orderID,
		})
		return map[string]any{
			"success": false,
			"message": "Delivery task not found",
		}, fmt.Errorf("DeliveryUseCase - HandleCargoDropped - GetDeliveryTask: %w", err)
	}

	if task == nil {
		uc.logger.Warn("DeliveryUseCase - HandleCargoDropped - delivery task not found", entityError.ErrDeliveryTaskNotFound, map[string]any{
			"orderID": orderID,
		})
		return map[string]any{
			"success": false,
			"message": "Delivery task not found",
		}, entityError.ErrDeliveryTaskNotFound
	}

	if err := uc.deliveryRepo.UpdateDeliveryStatus(ctx, orderID, entity.DeliveryStatusCompleted, nil); err != nil {
		uc.logger.Error("DeliveryUseCase - HandleCargoDropped - UpdateDeliveryStatus", err, map[string]any{
			"orderID": orderID,
		})
		return map[string]any{
			"success": false,
			"message": "Failed to update delivery status",
		}, fmt.Errorf("DeliveryUseCase - HandleCargoDropped - UpdateDeliveryStatus: %w", err)
	}

	if task.DroneID != nil {
		if err := uc.droneManager.ReleaseDrone(ctx, *task.DroneID); err != nil {
			uc.logger.Warn("DeliveryUseCase - HandleCargoDropped - ReleaseDrone", err, map[string]any{
				"droneID": *task.DroneID,
			})
		}
	}

	if uc.rabbitmqClient != nil {
		confirmationMessage := map[string]any{
			"order_id":       orderID,
			"locker_cell_id": lockerCellID,
		}
		if lockerCellID == "" && task.LockerCellID != "" {
			confirmationMessage["locker_cell_id"] = task.LockerCellID
		}
		if err := uc.rabbitmqClient.Publish(ctx, "confirmations", confirmationMessage); err != nil {
			uc.logger.Warn("DeliveryUseCase - HandleCargoDropped - Publish", err, map[string]any{
				"orderID": orderID,
			})
		}
	}

	uc.logger.Info("Cargo dropped successfully", nil, map[string]any{
		"orderID":      orderID,
		"lockerCellID": lockerCellID,
	})

	return map[string]any{
		"success": true,
		"message": "Cargo dropped successfully",
	}, nil
}

func (uc *DeliveryUseCase) CancelDelivery(ctx context.Context, deliveryID string) (map[string]any, error) {
	task, err := uc.deliveryRepo.GetDeliveryTask(ctx, deliveryID)
	if err != nil {
		uc.logger.Error("DeliveryUseCase - CancelDelivery - GetDeliveryTask", err, map[string]any{
			"deliveryID": deliveryID,
		})
		return map[string]any{
			"success": false,
			"message": "Delivery not found",
		}, fmt.Errorf("DeliveryUseCase - CancelDelivery - GetDeliveryTask: %w", err)
	}

	if task == nil {
		uc.logger.Warn("DeliveryUseCase - CancelDelivery - delivery task not found", entityError.ErrDeliveryTaskNotFound, map[string]any{
			"deliveryID": deliveryID,
		})
		return map[string]any{
			"success": false,
			"message": "Delivery not found",
		}, entityError.ErrDeliveryTaskNotFound
	}

	if uc.droneNotifier != nil && task.DroneID != nil {
		message := map[string]any{
			"type":      "command",
			"timestamp": time.Now().Format(time.RFC3339),
			"payload": map[string]any{
				"command": "cancel_delivery",
			},
		}
		if err := uc.droneNotifier.SendToDrone(ctx, *task.DroneID, message); err != nil {
			uc.logger.Warn("DeliveryUseCase - CancelDelivery - SendToDrone", err, map[string]any{
				"droneID": *task.DroneID,
			})
		}
	}

	if err := uc.deliveryRepo.UpdateDeliveryStatus(ctx, deliveryID, entity.DeliveryStatusCancelled, nil); err != nil {
		uc.logger.Error("DeliveryUseCase - CancelDelivery - UpdateDeliveryStatus", err, map[string]any{
			"deliveryID": deliveryID,
		})
		return map[string]any{
			"success": false,
			"message": "Failed to cancel delivery",
		}, fmt.Errorf("DeliveryUseCase - CancelDelivery - UpdateDeliveryStatus: %w", err)
	}

	if task.DroneID != nil {
		if err := uc.droneManager.ReleaseDrone(ctx, *task.DroneID); err != nil {
			uc.logger.Warn("DeliveryUseCase - CancelDelivery - ReleaseDrone", err, map[string]any{
				"droneID": *task.DroneID,
			})
		}
	}

	uc.logger.Info("Delivery cancelled", nil, map[string]any{
		"deliveryID": deliveryID,
	})

	return map[string]any{
		"success": true,
		"message": "Delivery cancelled",
	}, nil
}

func (uc *DeliveryUseCase) GetDeliveryStatus(ctx context.Context, deliveryID string) (map[string]any, error) {
	task, err := uc.deliveryRepo.GetDeliveryTask(ctx, deliveryID)
	if err != nil {
		uc.logger.Error("DeliveryUseCase - GetDeliveryStatus - GetDeliveryTask", err, map[string]any{
			"deliveryID": deliveryID,
		})
		return map[string]any{
			"success": false,
			"message": "Delivery not found",
		}, fmt.Errorf("DeliveryUseCase - GetDeliveryStatus - GetDeliveryTask: %w", err)
	}

	if task == nil {
		uc.logger.Warn("DeliveryUseCase - GetDeliveryStatus - delivery task not found", entityError.ErrDeliveryTaskNotFound, map[string]any{
			"deliveryID": deliveryID,
		})
		return map[string]any{
			"success": false,
			"message": "Delivery not found",
		}, entityError.ErrDeliveryTaskNotFound
	}

	status := "unknown"
	if task.Status != "" {
		status = string(task.Status)
	}

	droneID := ""
	if task.DroneID != nil {
		droneID = *task.DroneID
	}

	return map[string]any{
		"success":     true,
		"delivery_id": task.DeliveryID,
		"status":      status,
		"drone_id":    droneID,
	}, nil
}
