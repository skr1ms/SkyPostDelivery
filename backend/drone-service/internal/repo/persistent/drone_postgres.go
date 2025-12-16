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
		if isNoRows(err) {
			return "", entityError.ErrDroneNotFound
		}
		return "", fmt.Errorf("DroneRepo - GetDroneIDByIP: %w", err)
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
		if isNoRows(err) {
			return entityError.ErrDroneNotFound
		}
		return fmt.Errorf("DroneRepo - UpdateDroneBattery: %w", err)
	}

	return nil
}

func (r *DroneRepo) SaveDroneState(ctx context.Context, state *entity.DroneState) error {
	droneUUID, err := uuid.Parse(state.DroneID)
	if err != nil {
		return fmt.Errorf("DroneRepo - SaveDroneState - uuid.Parse: %w", err)
	}

	var currentDeliveryID pgtype.UUID
	if state.CurrentDeliveryID != nil {
		deliveryUUID, err := uuid.Parse(*state.CurrentDeliveryID)
		if err == nil {
			currentDeliveryID = pgtype.UUID{
				Bytes: deliveryUUID,
				Valid: true,
			}
		}
	}

	if err := r.q.SaveDroneState(ctx, sqlc.SaveDroneStateParams{
		ID:                droneUUID,
		Status:            string(state.Status),
		BatteryLevel:      float64ToNumeric(state.BatteryLevel),
		Latitude:          float64ToNumeric(state.CurrentPosition.Latitude),
		Longitude:         float64ToNumeric(state.CurrentPosition.Longitude),
		Altitude:          float64ToNumeric(state.CurrentPosition.Altitude),
		Speed:             float64ToNumeric(state.Speed),
		CurrentDeliveryID: currentDeliveryID,
		ErrorMessage:      state.ErrorMessage,
	}); err != nil {
		if isPgForeignKeyViolation(err) {
			return entityError.ErrDroneCreateFailed
		}
		return fmt.Errorf("DroneRepo - SaveDroneState: %w", err)
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
		if isNoRows(err) {
			return nil, entityError.ErrDroneStateNotFound
		}
		return nil, fmt.Errorf("DroneRepo - GetDroneState: %w", err)
	}

	state := &entity.DroneState{
		DroneID:      droneID,
		Status:       entity.DroneStatus(result.Status),
		BatteryLevel: numericToFloat64(result.BatteryLevel),
		CurrentPosition: entity.Position{
			Latitude:  numericToFloat64(result.Latitude),
			Longitude: numericToFloat64(result.Longitude),
			Altitude:  numericToFloat64(result.Altitude),
		},
		Speed: numericToFloat64(result.Speed),
	}

	if result.UpdatedAt.Valid {
		state.LastUpdated = result.UpdatedAt.Time
	} else {
		state.LastUpdated = time.Now()
	}

	if result.CurrentDeliveryID.Valid {
		deliveryID := uuid.UUID(result.CurrentDeliveryID.Bytes).String()
		state.CurrentDeliveryID = &deliveryID
	}

	if result.ErrorMessage != nil {
		state.ErrorMessage = result.ErrorMessage
	}

	return state, nil
}

func float64ToNumeric(f float64) pgtype.Numeric {
	n := pgtype.Numeric{}
	_ = n.Scan(fmt.Sprintf("%.2f", f))
	return n
}

func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	return f.Float64
}
