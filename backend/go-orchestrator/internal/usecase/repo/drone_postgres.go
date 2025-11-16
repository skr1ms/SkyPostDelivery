package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase/repo/sqlc"
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

func (r *DroneRepo) Create(ctx context.Context, model, status, ipAddress string) (*entity.Drone, error) {
	d, err := r.q.CreateDrone(ctx, sqlc.CreateDroneParams{
		Model:     model,
		Status:    status,
		IpAddress: ipAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("DroneRepo - Create - q.CreateDrone: %w", err)
	}
	return toEntityDrone(d), nil
}

func (r *DroneRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Drone, error) {
	d, err := r.q.GetDroneByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("DroneRepo - GetByID - q.GetDroneByID: %w", err)
	}
	return toEntityDrone(d), nil
}

func (r *DroneRepo) GetAvailable(ctx context.Context) (*entity.Drone, error) {
	d, err := r.q.GetAvailableDrone(ctx)
	if err != nil {
		return nil, fmt.Errorf("DroneRepo - GetAvailable - q.GetAvailableDrone: %w", err)
	}
	return toEntityDrone(d), nil
}

func (r *DroneRepo) List(ctx context.Context) ([]*entity.Drone, error) {
	rows, err := r.q.ListDrones(ctx)
	if err != nil {
		return nil, fmt.Errorf("DroneRepo - List - q.ListDrones: %w", err)
	}
	drones := make([]*entity.Drone, 0, len(rows))
	for _, d := range rows {
		drones = append(drones, toEntityDrone(d))
	}
	return drones, nil
}

func (r *DroneRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.q.UpdateDroneStatus(ctx, sqlc.UpdateDroneStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("DroneRepo - UpdateStatus - q.UpdateDroneStatus: %w", err)
	}
	return nil
}

func (r *DroneRepo) Update(ctx context.Context, id uuid.UUID, model, ipAddress string) (*entity.Drone, error) {
	d, err := r.q.UpdateDrone(ctx, sqlc.UpdateDroneParams{
		ID:        id,
		Model:     model,
		IpAddress: ipAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("DroneRepo - Update - q.UpdateDrone: %w", err)
	}
	return toEntityDrone(d), nil
}

func (r *DroneRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteDrone(ctx, id); err != nil {
		return fmt.Errorf("DroneRepo - Delete - q.DeleteDrone: %w", err)
	}
	return nil
}
