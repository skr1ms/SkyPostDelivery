package request

import "github.com/google/uuid"

type CreateOrder struct {
	GoodID uuid.UUID `json:"good_id" binding:"required"`
}

type CreateMultipleOrders struct {
	GoodIDs []uuid.UUID `json:"good_ids" binding:"required,min=1"`
}
