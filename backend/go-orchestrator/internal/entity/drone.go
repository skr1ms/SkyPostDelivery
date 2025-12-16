package entity

import "github.com/google/uuid"

type Drone struct {
	ID        uuid.UUID `json:"id"`
	Model     string    `json:"model"`
	IPAddress string    `json:"ip_address"`
	Status    string    `json:"status"`
}

type DroneStatus struct {
	DroneID      uuid.UUID `json:"drone_id"`
	Status       string    `json:"status"`
	BatteryLevel float64   `json:"battery_level"`
	Position     Position  `json:"position"`
}

type Position struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}
