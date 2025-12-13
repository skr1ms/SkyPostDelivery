package usecase

import (
	"context"
	"log"
	"time"

	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/rabbitmq"
)

func (uc *OrderUseCase) StartPendingOrdersWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Println("Starting pending orders worker...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping pending orders worker...")
			return
		case <-ticker.C:
			uc.processPendingOrders(ctx)
		}
	}
}

func (uc *OrderUseCase) processPendingOrders(ctx context.Context) {
	deliveries, err := uc.deliveryRepo.ListByStatus(ctx, "awaiting_drone")
	if err != nil {
		log.Printf("Error fetching awaiting_drone deliveries: %v", err)
		return
	}

	if len(deliveries) == 0 {
		return
	}

	log.Printf("Found %d pending deliveries, trying to assign drones...", len(deliveries))

	for _, delivery := range deliveries {
		drone, err := uc.droneRepo.GetAvailable(ctx)
		if err != nil {
			log.Printf("No available drones yet for delivery %s", delivery.ID)
			break
		}

		if err := uc.droneRepo.UpdateStatus(ctx, drone.ID, "busy"); err != nil {
			log.Printf("Failed to update drone %s status: %v", drone.ID, err)
			continue
		}

		delivery.DroneID = &drone.ID
		if err := uc.deliveryRepo.UpdateDrone(ctx, delivery.ID, drone.ID); err != nil {
			log.Printf("Failed to assign drone to delivery %s: %v", delivery.ID, err)
			_ = uc.droneRepo.UpdateStatus(ctx, drone.ID, "idle")
			continue
		}

		if _, err := uc.deliveryRepo.UpdateStatus(ctx, delivery.ID, "pending"); err != nil {
			log.Printf("Failed to update delivery %s status: %v", delivery.ID, err)
			continue
		}

		order, err := uc.orderRepo.GetByID(ctx, delivery.OrderID)
		if err != nil {
			log.Printf("Failed to get order %s: %v", delivery.OrderID, err)
			continue
		}

		good, err := uc.goodRepo.GetByID(ctx, order.GoodID)
		if err != nil {
			log.Printf("Failed to get good %s: %v", order.GoodID, err)
			continue
		}

		parcelAutomat, err := uc.parcelAutomatRepo.GetByID(ctx, delivery.ParcelAutomatID)
		if err != nil {
			log.Printf("Failed to get parcel automat %s: %v", delivery.ParcelAutomatID, err)
			continue
		}

		deliveryTask := rabbitmq.DeliveryTask{
			DroneID:         drone.ID,
			DroneIP:         drone.IPAddress,
			OrderID:         order.ID,
			GoodID:          order.GoodID,
			ParcelAutomatID: parcelAutomat.ID,
			ArucoID:         parcelAutomat.ArucoID,
			Coordinates:     parcelAutomat.Coordinates,
			Weight:          good.Weight,
			Height:          good.Height,
			Length:          good.Length,
			Width:           good.Width,
			Priority:        0,
			CreatedAt:       time.Now().Unix(),
		}

		queueName := rabbitmq.QueueDeliveries
		if deliveryTask.Priority > 5 {
			queueName = rabbitmq.QueueDeliveriesPriority
		}

		if err := uc.rabbitmqClient.Publish(ctx, queueName, deliveryTask); err != nil {
			log.Printf("Failed to publish delivery task: %v", err)
			_ = uc.droneRepo.UpdateStatus(ctx, drone.ID, "idle")
			_, _ = uc.deliveryRepo.UpdateStatus(ctx, delivery.ID, "awaiting_drone")
			continue
		}

		log.Printf("Successfully assigned drone %s to order %s", drone.ID, order.ID)
	}
}
