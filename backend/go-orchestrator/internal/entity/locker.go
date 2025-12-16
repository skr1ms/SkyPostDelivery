package entity

import "github.com/google/uuid"

type ParcelAutomat struct {
	ID            uuid.UUID `json:"id"`
	IPAddress     string    `json:"ip_address"`
	City          string    `json:"city"`
	Address       string    `json:"address"`
	NumberOfCells int       `json:"number_of_cells"`
	Coordinates   string    `json:"coordinates"`
	ArucoID       int       `json:"aruco_id"`
	IsWorking     bool      `json:"is_working"`
}

type LockerCell struct {
	ID     uuid.UUID `json:"id"`
	PostID uuid.UUID `json:"post_id"`
	Height float64   `json:"height"`
	Length float64   `json:"length"`
	Width  float64   `json:"width"`
	Status string    `json:"status"`
}
