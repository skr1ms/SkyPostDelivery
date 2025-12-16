package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo/persistent/sqlc"
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

func (r *DeliveryRepo) Create(ctx context.Context, delivery *entity.Delivery) (*entity.Delivery, error) {
	d, err := r.q.CreateDelivery(ctx, sqlc.CreateDeliveryParams{
		OrderID:              delivery.OrderID,
		DroneID:              ptrUUIDToPgUUID(delivery.DroneID),
		ParcelAutomatID:      delivery.ParcelAutomatID,
		InternalLockerCellID: ptrUUIDToPgUUID(delivery.InternalLockerCellID),
		Status:               delivery.Status,
	})
	if err != nil {
		if isPgForeignKeyViolation(err) {
			return nil, entityError.ErrDeliveryCreateFailed
		}
		return nil, fmt.Errorf("DeliveryRepo - Create: %w", err)
	}
	return toEntityDelivery(d), nil
}

func (r *DeliveryRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Delivery, error) {
	d, err := r.q.GetDeliveryByID(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrDeliveryNotFound
		}
		return nil, fmt.Errorf("DeliveryRepo - GetByID: %w", err)
	}
	return toEntityDelivery(d), nil
}

func (r *DeliveryRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.Delivery, error) {
	d, err := r.q.GetDeliveryByOrderID(ctx, orderID)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrDeliveryNotFound
		}
		return nil, fmt.Errorf("DeliveryRepo - GetByOrderID: %w", err)
	}
	return toEntityDelivery(d), nil
}

func (r *DeliveryRepo) UpdateStatus(ctx context.Context, delivery *entity.Delivery) (*entity.Delivery, error) {
	d, err := r.q.UpdateDeliveryStatus(ctx, sqlc.UpdateDeliveryStatusParams{
		ID:     delivery.ID,
		Status: delivery.Status,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrDeliveryNotFound
		}
		return nil, fmt.Errorf("DeliveryRepo - UpdateStatus: %w", err)
	}
	return toEntityDelivery(d), nil
}

func (r *DeliveryRepo) ListByStatus(ctx context.Context, status string) ([]*entity.Delivery, error) {
	rows, err := r.q.ListDeliveriesByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("DeliveryRepo - ListByStatus: %w", err)
	}
	deliveries := make([]*entity.Delivery, 0, len(rows))
	for _, d := range rows {
		deliveries = append(deliveries, toEntityDelivery(d))
	}
	return deliveries, nil
}

func (r *DeliveryRepo) UpdateDrone(ctx context.Context, delivery *entity.Delivery) error {
	_, err := r.q.UpdateDeliveryDrone(ctx, sqlc.UpdateDeliveryDroneParams{
		ID:      delivery.ID,
		DroneID: ptrUUIDToPgUUID(delivery.DroneID),
	})
	if err != nil {
		if isNoRows(err) {
			return entityError.ErrDeliveryNotFound
		}
		return fmt.Errorf("DeliveryRepo - UpdateDrone: %w", err)
	}
	return nil
}
