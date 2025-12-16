package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
)

type OrderWithGood struct {
	ID              uuid.UUID    `json:"id"`
	UserID          uuid.UUID    `json:"user_id"`
	GoodID          uuid.UUID    `json:"good_id"`
	ParcelAutomatID uuid.UUID    `json:"parcel_automat_id"`
	Status          string       `json:"status"`
	CreatedAt       time.Time    `json:"created_at"`
	Good            *entity.Good `json:"good,omitempty"`
}
