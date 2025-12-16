package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo/persistent/sqlc"
)

type DeliveryRepo struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewDeliveryRepo(pg *pgxpool.Pool) *DeliveryRepo {
	return &DeliveryRepo{
		db: pg,
		q:  sqlc.New(pg),
	}
}

func (r *DeliveryRepo) SaveDeliveryTask(ctx context.Context, task *entity.DeliveryTask) error {
	deliveryUUID, err := uuid.Parse(task.DeliveryID)
	if err != nil {
		return fmt.Errorf("DeliveryRepo - SaveDeliveryTask - uuid.Parse delivery: %w", err)
	}

	var droneUUID *uuid.UUID
	if task.DroneID != nil {
		parsed, err := uuid.Parse(*task.DroneID)
		if err != nil {
			return fmt.Errorf("DeliveryRepo - SaveDeliveryTask - uuid.Parse drone: %w", err)
		}
		droneUUID = &parsed
	}

	if err := r.q.SaveDeliveryTask(ctx, sqlc.SaveDeliveryTaskParams{
		ID:      deliveryUUID,
		DroneID: ptrUUIDToPgUUID(droneUUID),
		Status:  string(task.Status),
	}); err != nil {
		if isPgForeignKeyViolation(err) {
			return entityError.ErrDeliveryCreateFailed
		}
		return fmt.Errorf("DeliveryRepo - SaveDeliveryTask: %w", err)
	}

	return nil
}

func (r *DeliveryRepo) GetDeliveryTask(ctx context.Context, deliveryID string) (*entity.DeliveryTask, error) {
	deliveryUUID, err := uuid.Parse(deliveryID)
	if err != nil {
		return nil, fmt.Errorf("DeliveryRepo - GetDeliveryTask - uuid.Parse: %w", err)
	}

	result, err := r.q.GetDeliveryTask(ctx, deliveryUUID)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrDeliveryTaskNotFound
		}
		return nil, fmt.Errorf("DeliveryRepo - GetDeliveryTask: %w", err)
	}

	var droneID *string
	if result.DroneID.Valid {
		id := uuid.UUID(result.DroneID.Bytes).String()
		droneID = &id
	}

	var internalLockerCellID *string
	if result.InternalLockerCellID.Valid {
		id := uuid.UUID(result.InternalLockerCellID.Bytes).String()
		internalLockerCellID = &id
	}

	var lockerCellID string
	if result.LockerCellID.Valid {
		lockerCellID = uuid.UUID(result.LockerCellID.Bytes).String()
	}

	goodID := result.GoodID.String()

	return &entity.DeliveryTask{
		DeliveryID:           result.ID.String(),
		OrderID:              result.OrderID.String(),
		GoodID:               goodID,
		LockerCellID:         lockerCellID,
		ParcelAutomatID:      result.ParcelAutomatID.String(),
		InternalLockerCellID: internalLockerCellID,
		Status:               entity.DeliveryStatus(result.Status),
		DroneID:              droneID,
		StartedAt:            timestampToTimePtr(result.StartedAt),
		CompletedAt:          timestampToTimePtr(result.CompletedAt),
		CreatedAt:            time.Now(),
	}, nil
}

func timestampToTimePtr(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

func (r *DeliveryRepo) UpdateDeliveryStatus(ctx context.Context, deliveryID string, status entity.DeliveryStatus, errorMessage *string) error {
	deliveryUUID, err := uuid.Parse(deliveryID)
	if err != nil {
		return fmt.Errorf("DeliveryRepo - UpdateDeliveryStatus - uuid.Parse: %w", err)
	}

	if err := r.q.UpdateDeliveryStatus(ctx, sqlc.UpdateDeliveryStatusParams{
		ID:     deliveryUUID,
		Status: string(status),
	}); err != nil {
		if isNoRows(err) {
			return entityError.ErrDeliveryNotFound
		}
		return fmt.Errorf("DeliveryRepo - UpdateDeliveryStatus: %w", err)
	}

	return nil
}

func ptrUUIDToPgUUID(pu *uuid.UUID) pgtype.UUID {
	if pu == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *pu, Valid: true}
}
