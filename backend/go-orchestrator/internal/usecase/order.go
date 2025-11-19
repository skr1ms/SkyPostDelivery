package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/rabbitmq"
)

type OrderUseCase struct {
	orderRepo          repo.OrderRepo
	goodRepo           repo.GoodRepo
	droneRepo          repo.DroneRepo
	deliveryRepo       repo.DeliveryRepo
	parcelAutomatRepo  repo.ParcelAutomatRepo
	lockerRepo         repo.LockerRepo
	internalLockerRepo repo.InternalLockerRepo
	rabbitmqClient     rabbitmq.RabbitMQClient
}

func NewOrderUseCase(
	orderRepo repo.OrderRepo,
	goodRepo repo.GoodRepo,
	droneRepo repo.DroneRepo,
	deliveryRepo repo.DeliveryRepo,
	parcelAutomatRepo repo.ParcelAutomatRepo,
	lockerRepo repo.LockerRepo,
	internalLockerRepo repo.InternalLockerRepo,
	rabbitmqClient rabbitmq.RabbitMQClient,
) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:          orderRepo,
		goodRepo:           goodRepo,
		droneRepo:          droneRepo,
		deliveryRepo:       deliveryRepo,
		parcelAutomatRepo:  parcelAutomatRepo,
		lockerRepo:         lockerRepo,
		internalLockerRepo: internalLockerRepo,
		rabbitmqClient:     rabbitmqClient,
	}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, userID, goodID uuid.UUID) (*entity.Order, error) {
	good, err := uc.goodRepo.GetByID(ctx, goodID)
	if err != nil {
		return nil, fmt.Errorf("order usecase - CreateOrder - goodRepo.GetByID: %w", err)
	}

	if good.QuantityAvailable <= 0 {
		return nil, fmt.Errorf("order usecase - CreateOrder: good is out of stock")
	}

	automats, err := uc.parcelAutomatRepo.ListWorking(ctx)
	if err != nil || len(automats) == 0 {
		return nil, fmt.Errorf("order usecase - CreateOrder: no working parcel automats available")
	}
	parcelAutomat := automats[0]

	cell, err := uc.lockerRepo.FindAvailableCell(ctx, good.Height, good.Length, good.Width)
	if err != nil {
		return nil, fmt.Errorf("order usecase - CreateOrder: no available cell for good dimensions: %w", err)
	}

	if err := uc.lockerRepo.UpdateCellStatus(ctx, cell.ID, "reserved"); err != nil {
		return nil, fmt.Errorf("order usecase - CreateOrder - lockerRepo.UpdateCellStatus: %w", err)
	}

	if _, err := uc.goodRepo.UpdateQuantity(ctx, goodID, -1); err != nil {
		uc.lockerRepo.UpdateCellStatus(ctx, cell.ID, "available")
		return nil, fmt.Errorf("order usecase - CreateOrder - goodRepo.UpdateQuantity: %w", err)
	}

	var internalCellID *uuid.UUID
	var internalCellReserved bool
	if uc.internalLockerRepo != nil {
		allExternalCells, err := uc.lockerRepo.ListCellsByPostID(ctx, parcelAutomat.ID)
		if err == nil {
			allInternalCells, err := uc.internalLockerRepo.ListCellsByPostID(ctx, parcelAutomat.ID)
			if err == nil && len(allExternalCells) == len(allInternalCells) {
				for idx, extCell := range allExternalCells {
					if extCell.ID == cell.ID && idx < len(allInternalCells) {
						internalCell := allInternalCells[idx]
						if internalCell.Status == "available" {
							if err := uc.internalLockerRepo.UpdateCellStatus(ctx, internalCell.ID, "reserved"); err != nil {
								log.Printf("[ORDER] Failed to reserve internal cell %s: %v", internalCell.ID, err)
							} else {
								internalCellReserved = true
								internalCellID = &internalCell.ID
								log.Printf("[ORDER] Reserved internal cell %s (same index as external cell %s)", internalCell.ID, cell.ID)
							}
						}
						break
					}
				}
			}
		}

		if internalCellID == nil {
			log.Printf("[ORDER] Could not find matching internal cell for external cell %s, trying any available", cell.ID)
			internalCell, findErr := uc.internalLockerRepo.FindAvailableCell(ctx, 0, 0, 0)
			if findErr != nil {
				log.Printf("[ORDER] Failed to find internal locker cell for automat %s: %v", parcelAutomat.ID, findErr)
			} else {
				if err := uc.internalLockerRepo.UpdateCellStatus(ctx, internalCell.ID, "reserved"); err != nil {
					log.Printf("[ORDER] Failed to reserve internal locker cell %s for automat %s: %v", internalCell.ID, parcelAutomat.ID, err)
				} else {
					internalCellReserved = true
					internalCellID = &internalCell.ID
				}
			}
		}
	}

	order, err := uc.orderRepo.CreateWithCell(ctx, userID, goodID, parcelAutomat.ID, &cell.ID, "pending")
	if err != nil {
		uc.goodRepo.UpdateQuantity(ctx, goodID, 1)
		uc.lockerRepo.UpdateCellStatus(ctx, cell.ID, "available")
		if uc.internalLockerRepo != nil && internalCellID != nil {
			uc.internalLockerRepo.UpdateCellStatus(ctx, *internalCellID, "available")
		}
		return nil, fmt.Errorf("order usecase - CreateOrder - orderRepo.CreateWithCell: %w", err)
	}

	drone, err := uc.droneRepo.GetAvailable(ctx)
	if err != nil {
		log.Printf("[ORDER] No available drone for order %s: %v", order.ID, err)
		_, deliveryErr := uc.deliveryRepo.Create(ctx, order.ID, nil, parcelAutomat.ID, internalCellID, "awaiting_drone")
		if deliveryErr != nil {
			fmt.Printf("Warning: failed to create delivery record for order %s: %v\n", order.ID, deliveryErr)
		}
		return order, nil
	}

	log.Printf("[ORDER] Selected drone %s (%s) for order %s", drone.ID, drone.IPAddress, order.ID)

	if err := uc.droneRepo.UpdateStatus(ctx, drone.ID, "busy"); err != nil {
		log.Printf("[ORDER] Failed to update status to busy for drone %s (order %s): %v", drone.ID, order.ID, err)
		return order, nil
	}

	log.Printf("[ORDER] Drone %s status updated to busy for order %s", drone.ID, order.ID)

	delivery, err := uc.deliveryRepo.Create(ctx, order.ID, &drone.ID, parcelAutomat.ID, internalCellID, "pending")
	if err != nil {
		log.Printf("[ORDER] Failed to create delivery for order %s with drone %s: %v", order.ID, drone.ID, err)
		uc.droneRepo.UpdateStatus(ctx, drone.ID, "idle")
		if internalCellReserved && uc.internalLockerRepo != nil && internalCellID != nil {
			uc.internalLockerRepo.UpdateCellStatus(ctx, *internalCellID, "available")
		}
		return order, nil
	}

	log.Printf("[ORDER] Created delivery %s for order %s (drone %s)", delivery.ID, order.ID, drone.ID)

	deliveryTask := rabbitmq.DeliveryTask{
		DroneID:              drone.ID,
		DroneIP:              drone.IPAddress,
		OrderID:              order.ID,
		GoodID:               goodID,
		ParcelAutomatID:      parcelAutomat.ID,
		InternalLockerCellID: internalCellID,
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

	log.Printf("[ORDER] Publishing task for order %s to queue %s (drone %s, aruco %d)", order.ID, queueName, drone.ID, deliveryTask.ArucoID)

	if err := uc.rabbitmqClient.Publish(ctx, queueName, deliveryTask); err != nil {
		log.Printf("[ORDER] Publish failed for order %s: %v", order.ID, err)
		uc.droneRepo.UpdateStatus(ctx, drone.ID, "idle")
		uc.deliveryRepo.UpdateStatus(ctx, delivery.ID, "failed")
		if internalCellReserved && uc.internalLockerRepo != nil && internalCellID != nil {
			uc.internalLockerRepo.UpdateCellStatus(ctx, *internalCellID, "available")
		}
		return order, nil
	}

	log.Printf("[ORDER] Task published for order %s (queue %s)", order.ID, queueName)

	return order, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	order, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("order usecase - GetOrder - orderRepo.GetByID: %w", err)
	}
	return order, nil
}

func (uc *OrderUseCase) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*entity.Order, error) {
	orders, err := uc.orderRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("order usecase - GetUserOrders - orderRepo.ListByUserID: %w", err)
	}
	return orders, nil
}

func (uc *OrderUseCase) GetUserOrdersWithGoods(ctx context.Context, userID uuid.UUID) ([]struct {
	Order *entity.Order
	Good  *entity.Good
}, error) {
	result, err := uc.orderRepo.ListByUserIDWithGoods(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("order usecase - GetUserOrdersWithGoods - orderRepo.ListByUserIDWithGoods: %w", err)
	}
	return result, nil
}

func (uc *OrderUseCase) CreateMultipleOrders(ctx context.Context, userID uuid.UUID, goodIDs []uuid.UUID) ([]*entity.Order, error) {
	orders := make([]*entity.Order, 0, len(goodIDs))
	var lastErr error

	for _, goodID := range goodIDs {
		order, err := uc.CreateOrder(ctx, userID, goodID)
		if err != nil {
			lastErr = err
			continue
		}
		orders = append(orders, order)
	}

	if len(orders) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("order usecase - CreateMultipleOrders - failed to create any orders. Last error: %w", lastErr)
		}
		return nil, fmt.Errorf("order usecase - CreateMultipleOrders - failed to create any orders")
	}

	return orders, nil
}

func (uc *OrderUseCase) ReturnOrder(ctx context.Context, orderID, userID uuid.UUID) error {
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order usecase - ReturnOrder - orderRepo.GetByID: %w", err)
	}

	if order.UserID != userID {
		return fmt.Errorf("order usecase - ReturnOrder: order does not belong to user")
	}

	if order.Status != "pending" && order.Status != "in_progress" {
		return fmt.Errorf("order usecase - ReturnOrder: order cannot be returned (status: %s)", order.Status)
	}

	delivery, err := uc.deliveryRepo.GetByOrderID(ctx, orderID)
	var droneID *uuid.UUID
	if err == nil && delivery != nil && delivery.DroneID != nil {
		droneID = delivery.DroneID

		if delivery.Status == "in_transit" || delivery.Status == "pending" {
			returnTask := rabbitmq.DeliveryTask{
				DroneID:         *droneID,
				DroneIP:         "",
				GoodID:          uuid.Nil,
				ParcelAutomatID: uuid.Nil,
				ArucoID:         131,
				Coordinates:     "0,0",
				Weight:          0,
				Height:          0,
				Length:          0,
				Width:           0,
				Priority:        10,
				CreatedAt:       time.Now().Unix(),
			}

			if err := uc.rabbitmqClient.Publish(ctx, "delivery.return", returnTask); err != nil {
				fmt.Printf("Failed to publish return task: %v\n", err)
			}
		}

		if _, err := uc.deliveryRepo.UpdateStatus(ctx, delivery.ID, "cancelled"); err != nil {
			return fmt.Errorf("order usecase - ReturnOrder - deliveryRepo.UpdateStatus: %w", err)
		}

		if err := uc.droneRepo.UpdateStatus(ctx, *droneID, "returning"); err != nil {
			fmt.Printf("Failed to update drone status: %v\n", err)
		}
	}

	if order.LockerCellID != nil {
		if err := uc.lockerRepo.UpdateCellStatus(ctx, *order.LockerCellID, "available"); err != nil {
			return fmt.Errorf("order usecase - ReturnOrder - lockerRepo.UpdateCellStatus: %w", err)
		}
	}

	if uc.internalLockerRepo != nil && delivery != nil && delivery.InternalLockerCellID != nil {
		if err := uc.internalLockerRepo.UpdateCellStatus(ctx, *delivery.InternalLockerCellID, "available"); err != nil {
			log.Printf("[ORDER] Failed to release internal locker cell %s during return: %v", *delivery.InternalLockerCellID, err)
		}
	}

	if _, err := uc.goodRepo.UpdateQuantity(ctx, order.GoodID, 1); err != nil {
		return fmt.Errorf("order usecase - ReturnOrder - goodRepo.UpdateQuantity: %w", err)
	}

	if _, err := uc.orderRepo.UpdateStatus(ctx, orderID, "cancelled"); err != nil {
		return fmt.Errorf("order usecase - ReturnOrder - orderRepo.UpdateStatus: %w", err)
	}

	return nil
}
