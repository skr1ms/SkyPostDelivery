package rabbitmq

import "github.com/google/uuid"

type DeliveryTask struct {
	DroneID              uuid.UUID  `json:"drone_id"`
	DroneIP              string     `json:"drone_ip"`
	OrderID              uuid.UUID  `json:"order_id"`
	GoodID               uuid.UUID  `json:"good_id"`
	ParcelAutomatID      uuid.UUID  `json:"parcel_automat_id"`
	InternalLockerCellID *uuid.UUID `json:"internal_locker_cell_id,omitempty"`
	ArucoID              int        `json:"aruco_id"`
	Coordinates          string     `json:"coordinates"`
	Weight               float64    `json:"weight"`
	Height               float64    `json:"height"`
	Length               float64    `json:"length"`
	Width                float64    `json:"width"`
	Priority             int        `json:"priority"`
	CreatedAt            int64      `json:"created_at"`
}

type DeliveryConfirmation struct {
	OrderID      uuid.UUID `json:"order_id"`
	LockerCellID uuid.UUID `json:"locker_cell_id"`
	ConfirmedAt  int64     `json:"confirmed_at"`
	AutomatID    uuid.UUID `json:"automat_id"`
}

type DroneStatusUpdate struct {
	DroneID      uuid.UUID `json:"drone_id"`
	Status       string    `json:"status"`
	BatteryLevel float64   `json:"battery_level"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Altitude     float64   `json:"altitude"`
	UpdatedAt    int64     `json:"updated_at"`
}
