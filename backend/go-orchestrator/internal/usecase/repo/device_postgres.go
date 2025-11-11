package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/hitech-ekb/internal/entity"
)

type DeviceRepo struct {
	db *pgxpool.Pool
}

func NewDeviceRepo(db *pgxpool.Pool) *DeviceRepo {
	return &DeviceRepo{db: db}
}

func (r *DeviceRepo) Upsert(ctx context.Context, userID uuid.UUID, token, platform string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_devices (user_id, token, platform)
		VALUES ($1, $2, $3)
		ON CONFLICT (token) DO UPDATE
		SET user_id = EXCLUDED.user_id,
		    platform = EXCLUDED.platform,
		    updated_at = NOW()
	`, userID, token, platform)
	if err != nil {
		return fmt.Errorf("DeviceRepo - Upsert: %w", err)
	}
	return nil
}

func (r *DeviceRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Device, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, token, platform, created_at, updated_at
		FROM user_devices
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("DeviceRepo - ListByUserID: %w", err)
	}
	defer rows.Close()

	devices := make([]*entity.Device, 0)
	for rows.Next() {
		var device entity.Device
		if err := rows.Scan(&device.ID, &device.UserID, &device.Token, &device.Platform, &device.CreatedAt, &device.UpdatedAt); err != nil {
			return nil, fmt.Errorf("DeviceRepo - ListByUserID - rows.Scan: %w", err)
		}
		devices = append(devices, &device)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("DeviceRepo - ListByUserID - rows.Err: %w", err)
	}
	return devices, nil
}

func (r *DeviceRepo) DeleteByToken(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM user_devices
		WHERE token = $1
	`, token)
	if err != nil {
		return fmt.Errorf("DeviceRepo - DeleteByToken: %w", err)
	}
	return nil
}
