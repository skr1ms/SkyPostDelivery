package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo/persistent/sqlc"
)

type OrderRepo struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewOrderRepo(db *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{db: db, q: sqlc.New(db)}
}

func toEntityOrder(o sqlc.Order) *entity.Order {
	var lockerCellID *uuid.UUID
	if o.LockerCellID.Valid {
		id := uuid.UUID(o.LockerCellID.Bytes)
		lockerCellID = &id
	}
	return &entity.Order{
		ID:              o.ID,
		UserID:          o.UserID,
		GoodID:          o.GoodID,
		ParcelAutomatID: o.ParcelAutomatID,
		LockerCellID:    lockerCellID,
		Status:          o.Status,
		CreatedAt:       o.CreatedAt.Time,
	}
}

func (r *OrderRepo) Create(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	return r.CreateWithCell(ctx, order)
}

func (r *OrderRepo) CreateWithCell(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	var cellID pgtype.UUID
	if order.LockerCellID != nil {
		cellID = pgtype.UUID{
			Bytes: *order.LockerCellID,
			Valid: true,
		}
	} else {
		cellID = pgtype.UUID{Valid: false}
	}

	o, err := r.q.CreateOrder(ctx, sqlc.CreateOrderParams{
		UserID:          order.UserID,
		GoodID:          order.GoodID,
		ParcelAutomatID: order.ParcelAutomatID,
		LockerCellID:    cellID,
		Status:          order.Status,
	})
	if err != nil {
		if isPgForeignKeyViolation(err) {
			return nil, entityError.ErrOrderCreateFailed
		}
		return nil, fmt.Errorf("OrderRepo - CreateWithCell: %w", err)
	}
	return toEntityOrder(o), nil
}

func (r *OrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	o, err := r.q.GetOrderByID(ctx, id)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrOrderNotFound
		}
		return nil, fmt.Errorf("OrderRepo - GetByID: %w", err)
	}
	return toEntityOrder(o), nil
}

func (r *OrderRepo) GetByLockerCellID(ctx context.Context, lockerCellID uuid.UUID) (*entity.Order, error) {
	cellID := pgtype.UUID{
		Bytes: lockerCellID,
		Valid: true,
	}
	o, err := r.q.GetOrderByLockerCellID(ctx, cellID)
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrOrderNotFound
		}
		return nil, fmt.Errorf("OrderRepo - GetByLockerCellID: %w", err)
	}
	return toEntityOrder(o), nil
}

func (r *OrderRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Order, error) {
	rows, err := r.q.ListOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("OrderRepo - ListByUserID: %w", err)
	}
	orders := make([]*entity.Order, 0, len(rows))
	for _, row := range rows {
		var lockerCellID *uuid.UUID
		if row.LockerCellID.Valid {
			id := uuid.UUID(row.LockerCellID.Bytes)
			lockerCellID = &id
		}
		orders = append(orders, &entity.Order{
			ID:              row.ID,
			UserID:          row.UserID,
			GoodID:          row.GoodID,
			ParcelAutomatID: row.ParcelAutomatID,
			LockerCellID:    lockerCellID,
			Status:          row.Status,
			CreatedAt:       row.CreatedAt.Time,
		})
	}
	return orders, nil
}

func toOrderWithGood(row sqlc.ListOrdersByUserIDRow) (*entity.Order, *entity.Good) {
	var lockerCellID *uuid.UUID
	if row.LockerCellID.Valid {
		id := uuid.UUID(row.LockerCellID.Bytes)
		lockerCellID = &id
	}
	order := &entity.Order{
		ID:              row.ID,
		UserID:          row.UserID,
		GoodID:          row.GoodID,
		ParcelAutomatID: row.ParcelAutomatID,
		LockerCellID:    lockerCellID,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt.Time,
	}

	var good *entity.Good
	if row.GoodName != nil {
		var goodID uuid.UUID
		if row.GoodID_2.Valid {
			goodID = uuid.UUID(row.GoodID_2.Bytes)
		}

		good = &entity.Good{
			ID:                goodID,
			Name:              *row.GoodName,
			Weight:            parseNumeric(row.GoodWeight),
			Height:            parseNumeric(row.GoodHeight),
			Length:            parseNumeric(row.GoodLength),
			Width:             parseNumeric(row.GoodWidth),
			QuantityAvailable: int(*row.GoodQuantityAvailable),
		}
	}

	return order, good
}

func parseNumeric(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	return f.Float64
}

func (r *OrderRepo) ListByUserIDWithGoods(ctx context.Context, userID uuid.UUID) ([]struct {
	Order *entity.Order
	Good  *entity.Good
}, error) {
	rows, err := r.q.ListOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("OrderRepo - ListByUserIDWithGoods: %w", err)
	}

	result := make([]struct {
		Order *entity.Order
		Good  *entity.Good
	}, 0, len(rows))

	for _, row := range rows {
		order, good := toOrderWithGood(row)
		result = append(result, struct {
			Order *entity.Order
			Good  *entity.Good
		}{Order: order, Good: good})
	}

	return result, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	o, err := r.q.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		ID:     order.ID,
		Status: order.Status,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, entityError.ErrOrderNotFound
		}
		return nil, fmt.Errorf("OrderRepo - UpdateStatus: %w", err)
	}
	return toEntityOrder(o), nil
}
