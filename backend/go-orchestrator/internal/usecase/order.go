package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
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
	logger             logger.Interface
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
	logger logger.Interface,
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
		logger:             logger,
	}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, userID, goodID uuid.UUID) (*entity.Order, error) {
	good, err := uc.goodRepo.GetByID(ctx, goodID)
	if err != nil {
		return nil, err
	}

	if good.QuantityAvailable <= 0 {
		return nil, entityError.ErrGoodOutOfStock
	}

	automats, err := uc.parcelAutomatRepo.ListWorking(ctx)
	if err != nil || len(automats) == 0 {
		return nil, entityError.ErrOrderNoWorkingAutomats
	}
	parcelAutomat := automats[0]

	cell, err := uc.lockerRepo.FindAvailableCell(ctx, good.Height, good.Length, good.Width)
	if err != nil {
		if errors.Is(err, entityError.ErrLockerCellNotFound) {
			return nil, entityError.ErrOrderNoAvailableCell
		}
		return nil, fmt.Errorf("OrderUseCase - CreateOrder - FindCell: %w", err)
	}

	cell.Status = "reserved"
	if err := uc.lockerRepo.UpdateCellStatus(ctx, cell); err != nil {
		return nil, fmt.Errorf("OrderUseCase - CreateOrder - UpdateCellStatus: %w", err)
	}

	if _, err := uc.goodRepo.UpdateQuantity(ctx, goodID, -1); err != nil {
		cell.Status = "available"
		_ = uc.lockerRepo.UpdateCellStatus(ctx, cell)
		return nil, fmt.Errorf("OrderUseCase - CreateOrder - UpdateQuantity: %w", err)
	}

	var internalCellID *uuid.UUID
	var internalCellReserved bool
	if uc.internalLockerRepo != nil {
		cellID, err := uc.reserveInternalCell(ctx, parcelAutomat.ID, cell.ID)
		if err != nil {
			uc.logger.Warn("OrderUseCase - CreateOrder - ReserveInternalCell", err, map[string]any{
				"automatID": parcelAutomat.ID,
				"cellID":    cell.ID,
			})
		} else if cellID != nil {
			internalCellReserved = true
			internalCellID = cellID
		}
	}

	order := &entity.Order{
		UserID:          userID,
		GoodID:          goodID,
		ParcelAutomatID: parcelAutomat.ID,
		LockerCellID:    &cell.ID,
		Status:          "pending",
	}

	createdOrder, err := uc.orderRepo.CreateWithCell(ctx, order)
	if err != nil {
		_, _ = uc.goodRepo.UpdateQuantity(ctx, goodID, 1)
		cell.Status = "available"
		_ = uc.lockerRepo.UpdateCellStatus(ctx, cell)
		if uc.internalLockerRepo != nil && internalCellID != nil {
			internalCell, _ := uc.internalLockerRepo.GetCellByID(ctx, *internalCellID)
			if internalCell != nil {
				internalCell.Status = "available"
				_ = uc.internalLockerRepo.UpdateCellStatus(ctx, internalCell)
			}
		}
		return nil, err
	}

	drone, err := uc.droneRepo.GetAvailable(ctx)
	if err != nil {
		uc.logger.Warn("OrderUseCase - CreateOrder - GetAvailableDrone", err, map[string]any{"orderID": createdOrder.ID})
		deliveryEntity := &entity.Delivery{
			OrderID:              createdOrder.ID,
			DroneID:              nil,
			ParcelAutomatID:      parcelAutomat.ID,
			InternalLockerCellID: internalCellID,
			Status:               "awaiting_drone",
		}
		_, deliveryErr := uc.deliveryRepo.Create(ctx, deliveryEntity)
		if deliveryErr != nil {
			uc.logger.Error("OrderUseCase - CreateOrder - CreateDelivery", deliveryErr, map[string]any{"orderID": createdOrder.ID})
		}
		return createdOrder, nil
	}

	drone.Status = "busy"
	if err := uc.droneRepo.UpdateStatus(ctx, drone); err != nil {
		uc.logger.Warn("OrderUseCase - CreateOrder - UpdateDroneStatus", err, map[string]any{"droneID": drone.ID, "orderID": createdOrder.ID})
		return createdOrder, nil
	}

	deliveryEntity := &entity.Delivery{
		OrderID:              createdOrder.ID,
		DroneID:              &drone.ID,
		ParcelAutomatID:      parcelAutomat.ID,
		InternalLockerCellID: internalCellID,
		Status:               "pending",
	}
	delivery, err := uc.deliveryRepo.Create(ctx, deliveryEntity)
	if err != nil {
		uc.logger.Error("OrderUseCase - CreateOrder - CreateDelivery", err, map[string]any{"orderID": createdOrder.ID, "droneID": drone.ID})
		drone.Status = "idle"
		_ = uc.droneRepo.UpdateStatus(ctx, drone)
		if internalCellReserved && uc.internalLockerRepo != nil && internalCellID != nil {
			internalCell, _ := uc.internalLockerRepo.GetCellByID(ctx, *internalCellID)
			if internalCell != nil {
				internalCell.Status = "available"
				_ = uc.internalLockerRepo.UpdateCellStatus(ctx, internalCell)
			}
		}
		return createdOrder, nil
	}

	deliveryTask := rabbitmq.DeliveryTask{
		DroneID:              drone.ID,
		DroneIP:              drone.IPAddress,
		OrderID:              createdOrder.ID,
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

	if err := uc.rabbitmqClient.Publish(ctx, queueName, deliveryTask); err != nil {
		uc.logger.Error("OrderUseCase - CreateOrder - Publish", err, map[string]any{"orderID": createdOrder.ID, "queueName": queueName})
		drone.Status = "idle"
		_ = uc.droneRepo.UpdateStatus(ctx, drone)
		delivery.Status = "failed"
		_, _ = uc.deliveryRepo.UpdateStatus(ctx, delivery)
		_, _ = uc.goodRepo.UpdateQuantity(ctx, goodID, 1)
		cell.Status = "available"
		_ = uc.lockerRepo.UpdateCellStatus(ctx, cell)
		if internalCellReserved && uc.internalLockerRepo != nil && internalCellID != nil {
			internalCell, _ := uc.internalLockerRepo.GetCellByID(ctx, *internalCellID)
			if internalCell != nil {
				internalCell.Status = "available"
				_ = uc.internalLockerRepo.UpdateCellStatus(ctx, internalCell)
			}
		}
		createdOrder.Status = "failed"
		_, _ = uc.orderRepo.UpdateStatus(ctx, createdOrder)
		return nil, fmt.Errorf("OrderUseCase - CreateOrder - Publish: %w", err)
	}

	return createdOrder, nil
}

func (uc *OrderUseCase) reserveInternalCell(ctx context.Context, automatID, externalCellID uuid.UUID) (*uuid.UUID, error) {
	if uc.internalLockerRepo == nil {
		return nil, nil
	}

	allExternalCells, err := uc.lockerRepo.ListCellsByPostID(ctx, automatID)
	if err != nil {
		return nil, fmt.Errorf("OrderUseCase - reserveInternalCell - ListCellsByPostID[external]: %w", err)
	}

	allInternalCells, err := uc.internalLockerRepo.ListCellsByPostID(ctx, automatID)
	if err != nil {
		return nil, fmt.Errorf("OrderUseCase - reserveInternalCell - ListCellsByPostID[internal]: %w", err)
	}

	if len(allExternalCells) == len(allInternalCells) {
		for idx, extCell := range allExternalCells {
			if extCell.ID == externalCellID && idx < len(allInternalCells) {
				internalCell := allInternalCells[idx]
				if internalCell.Status == "available" {
					internalCell.Status = "reserved"
					if err := uc.internalLockerRepo.UpdateCellStatus(ctx, internalCell); err != nil {
						return nil, fmt.Errorf("OrderUseCase - reserveInternalCell - UpdateCellStatus: %w", err)
					}
					return &internalCell.ID, nil
				}
				break
			}
		}
	}

	internalCell, err := uc.internalLockerRepo.FindAvailableCell(ctx, 0, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("OrderUseCase - reserveInternalCell - FindAvailableCell: %w", err)
	}

	internalCell.Status = "reserved"
	if err := uc.internalLockerRepo.UpdateCellStatus(ctx, internalCell); err != nil {
		return nil, fmt.Errorf("OrderUseCase - reserveInternalCell - UpdateCellStatus[fallback]: %w", err)
	}

	return &internalCell.ID, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	order, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (uc *OrderUseCase) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*entity.Order, error) {
	orders, err := uc.orderRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("OrderUseCase - GetUserOrders: %w", err)
	}
	return orders, nil
}

func (uc *OrderUseCase) GetUserOrdersWithGoods(ctx context.Context, userID uuid.UUID) ([]struct {
	Order *entity.Order
	Good  *entity.Good
}, error) {
	result, err := uc.orderRepo.ListByUserIDWithGoods(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("OrderUseCase - GetUserOrdersWithGoods: %w", err)
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
			return nil, lastErr
		}
		return nil, entityError.ErrOrderCreateMultipleFailed
	}

	return orders, nil
}

func (uc *OrderUseCase) ReturnOrder(ctx context.Context, orderID, userID uuid.UUID) error {
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if order.UserID != userID {
		return entityError.ErrOrderNotBelongsToUser
	}

	if order.Status != "pending" && order.Status != "in_progress" {
		return entityError.ErrOrderCannotBeReturned
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
				uc.logger.Error("OrderUseCase - ReturnOrder - PublishReturnTask", err)
			}
		}

		delivery.Status = "cancelled"
		if _, err := uc.deliveryRepo.UpdateStatus(ctx, delivery); err != nil {
			return fmt.Errorf("OrderUseCase - ReturnOrder - UpdateDeliveryStatus: %w", err)
		}

		drone, err := uc.droneRepo.GetByID(ctx, *droneID)
		if err != nil {
			uc.logger.Warn("OrderUseCase - ReturnOrder - GetDrone", err, map[string]any{"droneID": *droneID})
		} else {
			drone.Status = "returning"
			if err := uc.droneRepo.UpdateStatus(ctx, drone); err != nil {
				uc.logger.Warn("OrderUseCase - ReturnOrder - UpdateDroneStatus", err, map[string]any{"droneID": *droneID})
			}
		}
	}

	if order.LockerCellID != nil {
		cell, err := uc.lockerRepo.GetCellByID(ctx, *order.LockerCellID)
		if err != nil {
			return fmt.Errorf("OrderUseCase - ReturnOrder - GetLockerCell: %w", err)
		}
		cell.Status = "available"
		if err := uc.lockerRepo.UpdateCellStatus(ctx, cell); err != nil {
			return fmt.Errorf("OrderUseCase - ReturnOrder - UpdateLockerCellStatus: %w", err)
		}
	}

	if uc.internalLockerRepo != nil && delivery != nil && delivery.InternalLockerCellID != nil {
		internalCell, err := uc.internalLockerRepo.GetCellByID(ctx, *delivery.InternalLockerCellID)
		if err != nil {
			uc.logger.Warn("OrderUseCase - ReturnOrder - GetInternalCell", err, map[string]any{"cellID": *delivery.InternalLockerCellID})
		} else {
			internalCell.Status = "available"
			if err := uc.internalLockerRepo.UpdateCellStatus(ctx, internalCell); err != nil {
				uc.logger.Warn("OrderUseCase - ReturnOrder - ReleaseInternalCell", err, map[string]any{"cellID": *delivery.InternalLockerCellID})
			}
		}
	}

	if _, err := uc.goodRepo.UpdateQuantity(ctx, order.GoodID, 1); err != nil {
		return fmt.Errorf("OrderUseCase - ReturnOrder - UpdateQuantity: %w", err)
	}

	order.Status = "cancelled"
	if _, err := uc.orderRepo.UpdateStatus(ctx, order); err != nil {
		return fmt.Errorf("OrderUseCase - ReturnOrder - UpdateStatus: %w", err)
	}

	return nil
}
