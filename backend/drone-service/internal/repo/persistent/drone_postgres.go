package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/repo/persistent/sqlc"
)

type DroneRepo struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewDroneRepo(pg *pgxpool.Pool) *DroneRepo {
	return &DroneRepo{
		db: pg,
		q:  sqlc.New(pg),
	}
}

func (r *DroneRepo) GetDroneIDByIP(ctx context.Context, ipAddress string) (string, error) {
	droneID, err := r.q.GetDroneIDByIP(ctx, ipAddress)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("DroneRepo - GetDroneIDByIP - q.GetDroneIDByIP: %w", err)
	}

	return droneID.String(), nil
}

func (r *DroneRepo) UpdateDroneBattery(ctx context.Context, droneID string, batteryLevel float64) error {
	droneUUID, err := uuid.Parse(droneID)
	if err != nil {
		return fmt.Errorf("DroneRepo - UpdateDroneBattery - uuid.Parse: %w", err)
	}

	if err := r.q.UpdateDroneBattery(ctx, sqlc.UpdateDroneBatteryParams{
		ID:           droneUUID,
		BatteryLevel: float64ToNumeric(batteryLevel),
	}); err != nil {
		return fmt.Errorf("DroneRepo - UpdateDroneBattery - q.UpdateDroneBattery: %w", err)
	}

	return nil
}

func (r *DroneRepo) SaveDroneState(ctx context.Context, state *entity.DroneState) error {
	droneUUID, err := uuid.Parse(state.DroneID)
	if err != nil {
		return fmt.Errorf("DroneRepo - SaveDroneState - uuid.Parse: %w", err)
	}

	if err := r.q.SaveDroneState(ctx, sqlc.SaveDroneStateParams{
		ID:           droneUUID,
		Status:       string(state.Status),
		BatteryLevel: float64ToNumeric(state.BatteryLevel),
	}); err != nil {
		return fmt.Errorf("DroneRepo - SaveDroneState - q.SaveDroneState: %w", err)
	}

	return nil
}

func (r *DroneRepo) GetDroneState(ctx context.Context, droneID string) (*entity.DroneState, error) {
	droneUUID, err := uuid.Parse(droneID)
	if err != nil {
		return nil, fmt.Errorf("DroneRepo - GetDroneState - uuid.Parse: %w", err)
	}

	result, err := r.q.GetDroneState(ctx, droneUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("DroneRepo - GetDroneState - q.GetDroneState: %w", err)
	}

	return &entity.DroneState{
		DroneID:      result.ID.String(),
		Status:       entity.DroneStatus(result.Status),
		BatteryLevel: numericToFloat64(result.BatteryLevel),
		LastUpdated:  time.Now(),
	}, nil
}

func float64ToNumeric(f float64) pgtype.Numeric {
	n := pgtype.Numeric{}
	n.Scan(fmt.Sprintf("%.2f", f))
	return n
}

func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	return f.Float64
}
