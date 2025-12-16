package entity

import (
	"time"

	"github.com/google/uuid"
)

type QRScanRequest struct {
	QRData string `json:"qr_data" binding:"required"`
}

type QRValidationRequest struct {
	QRData          string `json:"qr_data"`
	ParcelAutomatID string `json:"parcel_automat_id"`
}

type QRValidationResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	CellIDs []string `json:"cell_ids"`
}

type ConfirmPickupRequest struct {
	CellIDs []string `json:"cell_ids" binding:"required"`
}

type ConfirmLoadedRequest struct {
	OrderID      string    `json:"order_id" binding:"required"`
	LockerCellID string    `json:"locker_cell_id" binding:"required"`
	Timestamp    time.Time `json:"timestamp"`
}

type QRScanResponse struct {
	Success     bool               `json:"success"`
	Message     string             `json:"message"`
	CellsOpened []OpenCellResponse `json:"cells_opened"`
	CellCount   int                `json:"cell_count"`
}

type QRScanResult struct {
	QRData    string
	UserID    uuid.UUID
	Timestamp time.Time
}
