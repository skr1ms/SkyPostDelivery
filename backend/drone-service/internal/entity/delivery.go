package entity

import (
	"time"

	"github.com/google/uuid"
)

type DeliveryStatus string

const (
	DeliveryStatusPending    DeliveryStatus = "pending"
	DeliveryStatusInProgress DeliveryStatus = "in_progress"
	DeliveryStatusCompleted  DeliveryStatus = "completed"
	DeliveryStatusFailed     DeliveryStatus = "failed"
	DeliveryStatusCancelled  DeliveryStatus = "cancelled"
)

type GoodDimensions struct {
	Weight float64 `json:"weight"`
	Height float64 `json:"height"`
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
}

type DeliveryTask struct {
	DeliveryID           string         `json:"delivery_id"`
	OrderID              string         `json:"order_id"`
	GoodID               string         `json:"good_id"`
	LockerCellID         string         `json:"locker_cell_id"`
	ParcelAutomatID      string         `json:"parcel_automat_id"`
	Dimensions           GoodDimensions `json:"dimensions"`
	CreatedAt            time.Time      `json:"created_at"`
	InternalLockerCellID *string        `json:"internal_locker_cell_id,omitempty"`
	StartedAt            *time.Time     `json:"started_at,omitempty"`
	CompletedAt          *time.Time     `json:"completed_at,omitempty"`
	DroneID              *string        `json:"drone_id,omitempty"`
	ErrorMessage         *string        `json:"error_message,omitempty"`
	ArucoID              *int           `json:"aruco_id,omitempty"`
	Status               DeliveryStatus `json:"status"`
}

type Delivery struct {
	ID                   uuid.UUID
	OrderID              uuid.UUID
	DroneID              *uuid.UUID
	ParcelAutomatID      uuid.UUID
	LockerCellID         uuid.UUID
	InternalLockerCellID *uuid.UUID
	Status               string
	StartedAt            *time.Time
	CompletedAt          *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
