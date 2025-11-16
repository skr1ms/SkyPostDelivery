package entity

import (
	"time"

	"github.com/google/uuid"
)

type DroneStatus string

const (
	DroneStatusIdle        DroneStatus = "idle"
	DroneStatusTakingOff   DroneStatus = "taking_off"
	DroneStatusPickingUp   DroneStatus = "picking_up"
	DroneStatusInTransit   DroneStatus = "in_transit"
	DroneStatusDelivering  DroneStatus = "delivering"
	DroneStatusReturning   DroneStatus = "returning"
	DroneStatusLanding     DroneStatus = "landing"
	DroneStatusCharging    DroneStatus = "charging"
	DroneStatusError       DroneStatus = "error"
	DroneStatusMaintenance DroneStatus = "maintenance"
)

type Position struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

type DroneState struct {
	DroneID           string      `json:"drone_id"`
	Status            DroneStatus `json:"status"`
	BatteryLevel      float64     `json:"battery_level"`
	CurrentPosition   Position    `json:"current_position"`
	Speed             float64     `json:"speed"`
	LastUpdated       time.Time   `json:"last_updated"`
	CurrentDeliveryID *string     `json:"current_delivery_id,omitempty"`
	ErrorMessage      *string     `json:"error_message,omitempty"`
}

type Drone struct {
	ID           uuid.UUID
	Model        string
	Status       string
	BatteryLevel float64
	IPAddress    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
