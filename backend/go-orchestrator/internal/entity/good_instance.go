package entity

import (
	"time"

	"github.com/google/uuid"
)

type GoodInstance struct {
	ID              uuid.UUID
	GoodID          uuid.UUID
	Status          string
	StorageLocation *string
	ReservedAt      *time.Time
	DeliveredAt     *time.Time
	CreatedAt       time.Time
}
