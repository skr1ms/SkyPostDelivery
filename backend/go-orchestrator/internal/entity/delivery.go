package entity

import "github.com/google/uuid"

type Delivery struct {
	ID              uuid.UUID
	OrderID         uuid.UUID
	DroneID         *uuid.UUID
	ParcelAutomatID uuid.UUID
	Status          string
}
