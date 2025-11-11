package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/hitech-ekb/internal/entity"
	"github.com/skr1ms/hitech-ekb/internal/usecase/repo/sqlc"
)

type ParcelAutomatRepo struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewParcelAutomatRepo(db *pgxpool.Pool) *ParcelAutomatRepo {
	return &ParcelAutomatRepo{db: db, q: sqlc.New(db)}
}

func toEntityParcelAutomat(p sqlc.ParcelAutomat) *entity.ParcelAutomat {
	return &entity.ParcelAutomat{
		ID:            p.ID,
		IPAddress:     p.IpAddress,
		City:          p.City,
		Address:       p.Address,
		NumberOfCells: int(p.NumberOfCells),
		Coordinates:   p.Coordinates,
		ArucoID:       int(p.ArucoID),
		IsWorking:     p.IsWorking,
	}
}

func (r *ParcelAutomatRepo) Create(ctx context.Context, city, address string, numberOfCells int, ipAddress, coordinates string, arucoID int, isWorking bool) (*entity.ParcelAutomat, error) {
	p, err := r.q.CreateParcelAutomat(ctx, sqlc.CreateParcelAutomatParams{
		City:          city,
		Address:       address,
		NumberOfCells: int32(numberOfCells),
		IpAddress:     ipAddress,
		Coordinates:   coordinates,
		ArucoID:       int32(arucoID),
		IsWorking:     isWorking,
	})
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatRepo - Create - q.CreateParcelAutomat: %w", err)
	}
	return toEntityParcelAutomat(p), nil
}

func (r *ParcelAutomatRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.ParcelAutomat, error) {
	p, err := r.q.GetParcelAutomatByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatRepo - GetByID - q.GetParcelAutomatByID: %w", err)
	}
	return toEntityParcelAutomat(p), nil
}

func (r *ParcelAutomatRepo) List(ctx context.Context) ([]*entity.ParcelAutomat, error) {
	rows, err := r.q.ListParcelAutomats(ctx)
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatRepo - List - q.ListParcelAutomats: %w", err)
	}
	automats := make([]*entity.ParcelAutomat, 0, len(rows))
	for _, p := range rows {
		automats = append(automats, toEntityParcelAutomat(p))
	}
	return automats, nil
}

func (r *ParcelAutomatRepo) ListWorking(ctx context.Context) ([]*entity.ParcelAutomat, error) {
	rows, err := r.q.ListWorkingParcelAutomats(ctx)
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatRepo - ListWorking - q.ListWorkingParcelAutomats: %w", err)
	}
	automats := make([]*entity.ParcelAutomat, 0, len(rows))
	for _, p := range rows {
		automats = append(automats, toEntityParcelAutomat(p))
	}
	return automats, nil
}

func (r *ParcelAutomatRepo) UpdateStatus(ctx context.Context, id uuid.UUID, isWorking bool) (*entity.ParcelAutomat, error) {
	p, err := r.q.UpdateParcelAutomatStatus(ctx, sqlc.UpdateParcelAutomatStatusParams{
		ID:        id,
		IsWorking: isWorking,
	})
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatRepo - UpdateStatus - q.UpdateParcelAutomatStatus: %w", err)
	}
	return toEntityParcelAutomat(p), nil
}

func (r *ParcelAutomatRepo) Update(ctx context.Context, id uuid.UUID, city, address, ipAddress, coordinates string) (*entity.ParcelAutomat, error) {
	p, err := r.q.UpdateParcelAutomat(ctx, sqlc.UpdateParcelAutomatParams{
		ID:          id,
		City:        city,
		Address:     address,
		IpAddress:   ipAddress,
		Coordinates: coordinates,
	})
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatRepo - Update - q.UpdateParcelAutomat: %w", err)
	}
	return toEntityParcelAutomat(p), nil
}

func (r *ParcelAutomatRepo) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeleteParcelAutomat(ctx, id)
	if err != nil {
		return fmt.Errorf("ParcelAutomatRepo - Delete - q.DeleteParcelAutomat: %w", err)
	}
	return nil
}
