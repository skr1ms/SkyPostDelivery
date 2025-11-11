package entity

import "github.com/google/uuid"

type Good struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Weight            float64   `json:"weight"`
	Height            float64   `json:"height"`
	Length            float64   `json:"length"`
	Width             float64   `json:"width"`
	QuantityAvailable int       `json:"quantity_available"`
}
