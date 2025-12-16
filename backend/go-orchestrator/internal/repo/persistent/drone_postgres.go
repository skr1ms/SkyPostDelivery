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

type DroneRepo struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewDroneRepo(db *pgxpool.Pool) *DroneRepo {
	return &DroneRepo{db: db, q: sqlc.New(db)}
}

func toEntityDrone(d sqlc.Drone) *entity.Drone {
	return &entity.Drone{
		ID:        d.ID,
		Model:     d.Model,
		IPAddress: d.IpAddress,
		Status:    d.Status,
	}
}

func (r *DroneRepo) Create(ctx context.Context, drone *entity.Drone) (*entity.Drone, error) {
	d, err := r.q.CreateDrone(ctx, sqlc.CreateDroneParams{
		Model:     drone.Model,
		Status:    drone.Status,
		IpAddress: drone.IPAddress,
	})
	if err != nil {
		if isPgUniqueViolation(err) {
			return nil, entityError.ErrDroneCreateFailed
		}
		return nil, fmt.Errorf("DroneRepo - Create: %w", err)
	}
	return toEntityDrone(d), nil
}

func (r *DroneRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Drone, error) {
	d, err := r.q.GetDroneByID(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrDroneNotFound
		}
		return nil, fmt.Errorf("DroneRepo - GetByID: %w", err)
	}
	return toEntityDrone(d), nil
}

func (r *DroneRepo) GetAvailable(ctx context.Context) (*entity.Drone, error) {
	d, err := r.q.GetAvailableDrone(ctx)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrDroneNotAvailable
		}
		return nil, fmt.Errorf("DroneRepo - GetAvailable: %w", err)
	}
	return toEntityDrone(d), nil
}

func (r *DroneRepo) List(ctx context.Context) ([]*entity.Drone, error) {
	rows, err := r.q.ListDrones(ctx)
	if err != nil {
		return nil, fmt.Errorf("DroneRepo - List: %w", err)
	}
	drones := make([]*entity.Drone, 0, len(rows))
	for _, d := range rows {
		drones = append(drones, toEntityDrone(d))
	}
	return drones, nil
}

func (r *DroneRepo) UpdateStatus(ctx context.Context, drone *entity.Drone) error {
	_, err := r.q.UpdateDroneStatus(ctx, sqlc.UpdateDroneStatusParams{
		ID:     drone.ID,
		Status: drone.Status,
	})
	if err != nil {
		if isNoRows(err) {
			return entityError.ErrDroneNotFound
		}
		return fmt.Errorf("DroneRepo - UpdateStatus: %w", err)
	}
	return nil
}

func (r *DroneRepo) Update(ctx context.Context, drone *entity.Drone) (*entity.Drone, error) {
	d, err := r.q.UpdateDrone(ctx, sqlc.UpdateDroneParams{
		ID:        drone.ID,
		Model:     drone.Model,
		IpAddress: drone.IPAddress,
		Status:    drone.Status,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrDroneNotFound
		}
		return nil, fmt.Errorf("DroneRepo - Update: %w", err)
	}
	return toEntityDrone(d), nil
}

func (r *DroneRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteDrone(ctx, id); err != nil {
		if isNoRows(err) {
			return entityError.ErrDroneNotFound
		}
		return fmt.Errorf("DroneRepo - Delete: %w", err)
	}
	return nil
}
