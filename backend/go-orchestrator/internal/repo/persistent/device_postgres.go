package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo/persistent/sqlc"
)

type DeviceRepo struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewDeviceRepo(db *pgxpool.Pool) *DeviceRepo {
	return &DeviceRepo{db: db, q: sqlc.New(db)}
}

func toEntityDevice(d sqlc.UserDevice) *entity.Device {
	return &entity.Device{
		ID:        d.ID,
		UserID:    d.UserID,
		Token:     d.Token,
		Platform:  d.Platform,
		CreatedAt: d.CreatedAt.Time,
		UpdatedAt: d.UpdatedAt.Time,
	}
}

func (r *DeviceRepo) Upsert(ctx context.Context, device *entity.Device) error {
	_, err := r.q.UpsertDevice(ctx, sqlc.UpsertDeviceParams{
		UserID:   device.UserID,
		Token:    device.Token,
		Platform: device.Platform,
	})
	if err != nil {
		return fmt.Errorf("DeviceRepo - Upsert: %w", err)
	}
	return nil
}

func (r *DeviceRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Device, error) {
	rows, err := r.q.ListDevicesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("DeviceRepo - ListByUserID: %w", err)
	}
	devices := make([]*entity.Device, 0, len(rows))
	for _, d := range rows {
		devices = append(devices, toEntityDevice(d))
	}
	return devices, nil
}

func (r *DeviceRepo) DeleteByToken(ctx context.Context, token string) error {
	if err := r.q.DeleteDeviceByToken(ctx, token); err != nil {
		if isNoRows(err) {
			return entityError.ErrDeviceNotFound
		}
		return fmt.Errorf("DeviceRepo - DeleteByToken: %w", err)
	}
	return nil
}
