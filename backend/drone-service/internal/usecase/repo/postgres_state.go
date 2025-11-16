package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/postgres"
)

type PostgresStateRepo struct {
	*postgres.Postgres
}

func NewPostgresStateRepo(pg *postgres.Postgres) *PostgresStateRepo {
	return &PostgresStateRepo{Postgres: pg}
}

func (r *PostgresStateRepo) GetDroneIDByIP(ctx context.Context, ipAddress string) (string, error) {
	var droneID uuid.UUID

	query := `SELECT id FROM drones WHERE ip_address = $1`
	err := r.Pool.QueryRow(ctx, query, ipAddress).Scan(&droneID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("postgres - GetDroneIDByIP - QueryRow: %w", err)
	}

	return droneID.String(), nil
}

func (r *PostgresStateRepo) UpdateDroneBattery(ctx context.Context, droneID string, batteryLevel float64) error {
	droneUUID, err := uuid.Parse(droneID)
	if err != nil {
		return fmt.Errorf("postgres - UpdateDroneBattery - uuid.Parse: %w", err)
	}

	query := `SELECT update_drone_battery($1, $2)`
	_, err = r.Pool.Exec(ctx, query, droneUUID, batteryLevel)
	if err != nil {
		return fmt.Errorf("postgres - UpdateDroneBattery - Exec: %w", err)
	}

	return nil
}

func (r *PostgresStateRepo) SaveDroneState(ctx context.Context, state *entity.DroneState) error {
	droneUUID, err := uuid.Parse(state.DroneID)
	if err != nil {
		return fmt.Errorf("postgres - SaveDroneState - uuid.Parse: %w", err)
	}

	query := `
		UPDATE drones
		SET 
			status = $2,
			battery_level = $3,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err = r.Pool.Exec(ctx, query, droneUUID, state.Status, state.BatteryLevel)
	if err != nil {
		return fmt.Errorf("postgres - SaveDroneState - Exec: %w", err)
	}

	return nil
}

func (r *PostgresStateRepo) GetDroneState(ctx context.Context, droneID string) (*entity.DroneState, error) {
	droneUUID, err := uuid.Parse(droneID)
	if err != nil {
		return nil, fmt.Errorf("postgres - GetDroneState - uuid.Parse: %w", err)
	}

	query := `
		SELECT id, status, battery_level
		FROM drones
		WHERE id = $1
	`

	var id uuid.UUID
	var status string
	var batteryLevel float64

	err = r.Pool.QueryRow(ctx, query, droneUUID).Scan(&id, &status, &batteryLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("postgres - GetDroneState - QueryRow: %w", err)
	}

	return &entity.DroneState{
		DroneID:      id.String(),
		Status:       entity.DroneStatus(status),
		BatteryLevel: batteryLevel,
		LastUpdated:  time.Now(),
	}, nil
}

func (r *PostgresStateRepo) SaveDeliveryTask(ctx context.Context, task *entity.DeliveryTask) error {
	deliveryUUID, err := uuid.Parse(task.DeliveryID)
	if err != nil {
		return fmt.Errorf("postgres - SaveDeliveryTask - uuid.Parse delivery: %w", err)
	}

	var droneUUID *uuid.UUID
	if task.DroneID != nil {
		parsed, err := uuid.Parse(*task.DroneID)
		if err != nil {
			return fmt.Errorf("postgres - SaveDeliveryTask - uuid.Parse drone: %w", err)
		}
		droneUUID = &parsed
	}

	query := `
		UPDATE deliveries
		SET 
			drone_id = $2,
			status = $3,
			started_at = COALESCE(started_at, CURRENT_TIMESTAMP)
		WHERE id = $1
	`

	_, err = r.Pool.Exec(ctx, query, deliveryUUID, droneUUID, task.Status)
	if err != nil {
		return fmt.Errorf("postgres - SaveDeliveryTask - Exec: %w", err)
	}

	return nil
}

func (r *PostgresStateRepo) GetDeliveryTask(ctx context.Context, deliveryID string) (*entity.DeliveryTask, error) {
	return nil, nil
}

func (r *PostgresStateRepo) UpdateDeliveryStatus(ctx context.Context, deliveryID string, status entity.DeliveryStatus, errorMessage *string) error {
	deliveryUUID, err := uuid.Parse(deliveryID)
	if err != nil {
		return fmt.Errorf("postgres - UpdateDeliveryStatus - uuid.Parse: %w", err)
	}

	var query string
	var args []interface{}

	if status == entity.DeliveryStatusCompleted {
		query = `
			UPDATE deliveries
			SET 
				status = $2,
				completed_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`
		args = []interface{}{deliveryUUID, status}
	} else {
		query = `
			UPDATE deliveries
			SET status = $2
			WHERE id = $1
		`
		args = []interface{}{deliveryUUID, status}
	}

	_, err = r.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("postgres - UpdateDeliveryStatus - Exec: %w", err)
	}

	return nil
}
