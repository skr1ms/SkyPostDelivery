package entity

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	GoodID          uuid.UUID  `json:"good_id"`
	ParcelAutomatID uuid.UUID  `json:"parcel_automat_id"`
	LockerCellID    *uuid.UUID `json:"locker_cell_id,omitempty"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
}
