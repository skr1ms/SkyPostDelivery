package entity

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	Platform  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
