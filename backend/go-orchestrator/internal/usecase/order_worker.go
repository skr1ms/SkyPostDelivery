package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/rabbitmq"
)

func (uc *OrderUseCase) StartPendingOrdersWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	uc.logger.Info("Pending orders worker started", nil, map[string]any{
		"interval": interval.String(),
	})

	for {
		select {
		case <-ctx.Done():
			uc.logger.Info("Pending orders worker stopped", nil)
			return
		case <-ticker.C:
			uc.processPendingOrders(ctx)
		}
	}
}

func (uc *OrderUseCase) processPendingOrders(ctx context.Context) {
	deliveries, err := uc.deliveryRepo.ListByStatus(ctx, "awaiting_drone")
	if err != nil {
		uc.logger.Error("OrderUseCase - processPendingOrders - ListByStatus", err)
		return
	}

	if len(deliveries) == 0 {
		return
	}

	uc.logger.Info("Processing pending deliveries", nil, map[string]any{
		"count": len(deliveries),
	})

	for _, delivery := range deliveries {
		if err := uc.processSingleDelivery(ctx, delivery); err != nil {
			continue
		}
	}
}

func (uc *OrderUseCase) processSingleDelivery(ctx context.Context, delivery *entity.Delivery) error {
	drone, err := uc.droneRepo.GetAvailable(ctx)
	if err != nil {
		if errors.Is(err, entityError.ErrDroneNotAvailable) {
			uc.logger.Debug("No available drones, will retry later", nil, map[string]any{
				"deliveryID": delivery.ID,
			})
			return err
		}
		uc.logger.Error("OrderUseCase - processSingleDelivery - GetAvailableDrone", err, map[string]any{
			"deliveryID": delivery.ID,
		})
		return err
	}

	drone.Status = "busy"
	if err := uc.droneRepo.UpdateStatus(ctx, drone); err != nil {
		uc.logger.Error("OrderUseCase - processSingleDelivery - UpdateDroneStatus", err, map[string]any{
			"droneID":    drone.ID,
			"deliveryID": delivery.ID,
		})
		return err
	}

	delivery.DroneID = &drone.ID
	if err := uc.deliveryRepo.UpdateDrone(ctx, delivery); err != nil {
		uc.logger.Error("OrderUseCase - processSingleDelivery - UpdateDrone", err, map[string]any{
			"deliveryID": delivery.ID,
			"droneID":    drone.ID,
		})
		drone.Status = "idle"
		_ = uc.droneRepo.UpdateStatus(ctx, drone)
		return err
	}

	delivery.Status = "pending"
	if _, err := uc.deliveryRepo.UpdateStatus(ctx, delivery); err != nil {
		uc.logger.Error("OrderUseCase - processSingleDelivery - UpdateDeliveryStatus", err, map[string]any{
			"deliveryID": delivery.ID,
		})
		return err
	}

	order, err := uc.orderRepo.GetByID(ctx, delivery.OrderID)
	if err != nil {
		uc.logger.Error("OrderUseCase - processSingleDelivery - GetOrder", err, map[string]any{
			"orderID": delivery.OrderID,
		})
		return err
	}

	good, err := uc.goodRepo.GetByID(ctx, order.GoodID)
	if err != nil {
		uc.logger.Error("OrderUseCase - processSingleDelivery - GetGood", err, map[string]any{
			"goodID": order.GoodID,
		})
		return err
	}

	parcelAutomat, err := uc.parcelAutomatRepo.GetByID(ctx, delivery.ParcelAutomatID)
	if err != nil {
		uc.logger.Error("OrderUseCase - processSingleDelivery - GetParcelAutomat", err, map[string]any{
			"parcelAutomatID": delivery.ParcelAutomatID,
		})
		return err
	}

	deliveryTask := rabbitmq.DeliveryTask{
		DroneID:              drone.ID,
		DroneIP:              drone.IPAddress,
		OrderID:              order.ID,
		GoodID:               order.GoodID,
		ParcelAutomatID:      parcelAutomat.ID,
		InternalLockerCellID: delivery.InternalLockerCellID,
		ArucoID:              parcelAutomat.ArucoID,
		Coordinates:          parcelAutomat.Coordinates,
		Weight:               good.Weight,
		Height:               good.Height,
		Length:               good.Length,
		Width:                good.Width,
		Priority:             0,
		CreatedAt:            time.Now().Unix(),
	}

	queueName := rabbitmq.QueueDeliveries
	if deliveryTask.Priority > 5 {
		queueName = rabbitmq.QueueDeliveriesPriority
	}

	if err := uc.rabbitmqClient.Publish(ctx, queueName, deliveryTask); err != nil {
		uc.logger.Error("OrderUseCase - processSingleDelivery - Publish", err, map[string]any{
			"queue":      queueName,
			"orderID":    order.ID,
			"droneID":    drone.ID,
			"deliveryID": delivery.ID,
		})
		drone.Status = "idle"
		_ = uc.droneRepo.UpdateStatus(ctx, drone)
		delivery.Status = "awaiting_drone"
		_, _ = uc.deliveryRepo.UpdateStatus(ctx, delivery)
		return err
	}

	uc.logger.Info("Drone assigned to order", nil, map[string]any{
		"droneID":    drone.ID,
		"orderID":    order.ID,
		"deliveryID": delivery.ID,
	})

	return nil
}
