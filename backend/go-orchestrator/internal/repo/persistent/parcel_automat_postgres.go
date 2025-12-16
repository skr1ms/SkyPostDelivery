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

func (r *ParcelAutomatRepo) Create(ctx context.Context, automat *entity.ParcelAutomat) (*entity.ParcelAutomat, error) {
	p, err := r.q.CreateParcelAutomat(ctx, sqlc.CreateParcelAutomatParams{
		City:          automat.City,
		Address:       automat.Address,
		NumberOfCells: int32(automat.NumberOfCells),
		IpAddress:     automat.IPAddress,
		Coordinates:   automat.Coordinates,
		ArucoID:       int32(automat.ArucoID),
		IsWorking:     automat.IsWorking,
	})
	if err != nil {
		if isPgUniqueViolation(err) {
			return nil, entityError.ErrParcelAutomatCreateFailed
		}
		return nil, fmt.Errorf("ParcelAutomatRepo - Create: %w", err)
	}
	return toEntityParcelAutomat(p), nil
}

func (r *ParcelAutomatRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.ParcelAutomat, error) {
	p, err := r.q.GetParcelAutomatByID(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrParcelAutomatNotFound
		}
		return nil, fmt.Errorf("ParcelAutomatRepo - GetByID: %w", err)
	}
	return toEntityParcelAutomat(p), nil
}

func (r *ParcelAutomatRepo) List(ctx context.Context) ([]*entity.ParcelAutomat, error) {
	rows, err := r.q.ListParcelAutomats(ctx)
	if err != nil {
		return nil, fmt.Errorf("ParcelAutomatRepo - List: %w", err)
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
		return nil, fmt.Errorf("ParcelAutomatRepo - ListWorking: %w", err)
	}
	automats := make([]*entity.ParcelAutomat, 0, len(rows))
	for _, p := range rows {
		automats = append(automats, toEntityParcelAutomat(p))
	}
	return automats, nil
}

func (r *ParcelAutomatRepo) UpdateStatus(ctx context.Context, automat *entity.ParcelAutomat) (*entity.ParcelAutomat, error) {
	p, err := r.q.UpdateParcelAutomatStatus(ctx, sqlc.UpdateParcelAutomatStatusParams{
		ID:        automat.ID,
		IsWorking: automat.IsWorking,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrParcelAutomatNotFound
		}
		return nil, fmt.Errorf("ParcelAutomatRepo - UpdateStatus: %w", err)
	}
	return toEntityParcelAutomat(p), nil
}

func (r *ParcelAutomatRepo) Update(ctx context.Context, automat *entity.ParcelAutomat) (*entity.ParcelAutomat, error) {
	p, err := r.q.UpdateParcelAutomat(ctx, sqlc.UpdateParcelAutomatParams{
		ID:          automat.ID,
		City:        automat.City,
		Address:     automat.Address,
		IpAddress:   automat.IPAddress,
		Coordinates: automat.Coordinates,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrParcelAutomatNotFound
		}
		return nil, fmt.Errorf("ParcelAutomatRepo - Update: %w", err)
	}
	return toEntityParcelAutomat(p), nil
}

func (r *ParcelAutomatRepo) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeleteParcelAutomat(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return entityError.ErrParcelAutomatNotFound
		}
		return fmt.Errorf("ParcelAutomatRepo - Delete: %w", err)
	}
	return nil
}
