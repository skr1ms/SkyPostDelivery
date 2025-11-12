package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/hitech-ekb/internal/entity"
	"github.com/skr1ms/hitech-ekb/internal/usecase/repo/sqlc"
)

type DeliveryRepo struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewDeliveryRepo(db *pgxpool.Pool) *DeliveryRepo {
	return &DeliveryRepo{db: db, q: sqlc.New(db)}
}

func pgUUIDToPtrUUID(pu pgtype.UUID) *uuid.UUID {
	if !pu.Valid {
		return nil
	}
	u := uuid.UUID(pu.Bytes)
	return &u
}

func ptrUUIDToPgUUID(pu *uuid.UUID) pgtype.UUID {
	if pu == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *pu, Valid: true}
}

func toEntityDelivery(d sqlc.Delivery) *entity.Delivery {
	return &entity.Delivery{
		ID:                   d.ID,
		OrderID:              d.OrderID,
		DroneID:              pgUUIDToPtrUUID(d.DroneID),
		ParcelAutomatID:      d.ParcelAutomatID,
		InternalLockerCellID: pgUUIDToPtrUUID(d.InternalLockerCellID),
		Status:               d.Status,
	}
}

func (r *DeliveryRepo) Create(ctx context.Context, orderID uuid.UUID, droneID *uuid.UUID, parcelAutomatID uuid.UUID, internalLockerCellID *uuid.UUID, status string) (*entity.Delivery, error) {
	d, err := r.q.CreateDelivery(ctx, sqlc.CreateDeliveryParams{
		OrderID:              orderID,
		DroneID:              ptrUUIDToPgUUID(droneID),
		ParcelAutomatID:      parcelAutomatID,
		InternalLockerCellID: ptrUUIDToPgUUID(internalLockerCellID),
		Status:               status,
	})
	if err != nil {
		return nil, fmt.Errorf("DeliveryRepo - Create - q.CreateDelivery: %w", err)
	}
	return toEntityDelivery(d), nil
}

func (r *DeliveryRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Delivery, error) {
	d, err := r.q.GetDeliveryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("DeliveryRepo - GetByID - q.GetDeliveryByID: %w", err)
	}
	return toEntityDelivery(d), nil
}

func (r *DeliveryRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.Delivery, error) {
	d, err := r.q.GetDeliveryByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("DeliveryRepo - GetByOrderID - q.GetDeliveryByOrderID: %w", err)
	}
	return toEntityDelivery(d), nil
}

func (r *DeliveryRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*entity.Delivery, error) {
	d, err := r.q.UpdateDeliveryStatus(ctx, sqlc.UpdateDeliveryStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("DeliveryRepo - UpdateStatus - q.UpdateDeliveryStatus: %w", err)
	}
	return toEntityDelivery(d), nil
}

func (r *DeliveryRepo) ListByStatus(ctx context.Context, status string) ([]*entity.Delivery, error) {
	rows, err := r.q.ListDeliveriesByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("DeliveryRepo - ListByStatus - q.ListDeliveriesByStatus: %w", err)
	}
	deliveries := make([]*entity.Delivery, 0, len(rows))
	for _, d := range rows {
		deliveries = append(deliveries, toEntityDelivery(d))
	}
	return deliveries, nil
}

func (r *DeliveryRepo) UpdateDrone(ctx context.Context, id uuid.UUID, droneID uuid.UUID) error {
	_, err := r.q.UpdateDeliveryDrone(ctx, sqlc.UpdateDeliveryDroneParams{
		ID:      id,
		DroneID: ptrUUIDToPgUUID(&droneID),
	})
	if err != nil {
		return fmt.Errorf("DeliveryRepo - UpdateDrone - q.UpdateDeliveryDrone: %w", err)
	}
	return nil
}
