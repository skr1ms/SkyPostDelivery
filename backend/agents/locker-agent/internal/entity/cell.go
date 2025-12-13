package entity

import (
	"time"

	"github.com/google/uuid"
)

type CellType string

const (
	CellTypeExternal CellType = "external"
	CellTypeInternal CellType = "internal"
)

type Cell struct {
	Number          int
	UUID            uuid.UUID
	Type            CellType
	ParcelAutomatID uuid.UUID
}

type CellMapping struct {
	ParcelAutomatID uuid.UUID
	ExternalCells   map[int]uuid.UUID
	InternalCells   map[int]uuid.UUID
	LastSyncTime    time.Time
	Initialized     bool
}

type SyncCellsRequest struct {
	ParcelAutomatID string   `json:"parcel_automat_id" binding:"required"`
	CellsOut        []string `json:"cells_out" binding:"required"`
	CellsInternal   []string `json:"cells_internal"`
}

type SyncCellsResponse struct {
	Message            string `json:"message"`
	CellsCount         int    `json:"cells_count"`
	InternalCellsCount int    `json:"internal_cells_count"`
	ParcelAutomatID    string `json:"parcel_automat_id"`
}

type OpenCellRequest struct {
	CellNumber  int    `json:"cell_number" binding:"required,min=1"`
	OrderNumber string `json:"order_number"`
}

type OpenCellResponse struct {
	Success       bool      `json:"success"`
	CellNumber    int       `json:"cell_number"`
	CellUUID      uuid.UUID `json:"cell_uuid"`
	Action        string    `json:"action"`
	Type          string    `json:"type"`
	ArduinoStatus string    `json:"arduino_status,omitempty"`
}

type CellMappingResponse struct {
	Mapping            map[string]CellInfo `json:"mapping"`
	CellsCount         int                 `json:"cells_count"`
	InternalMapping    map[string]CellInfo `json:"internal_mapping"`
	InternalCellsCount int                 `json:"internal_cells_count"`
	ParcelAutomatID    string              `json:"parcel_automat_id"`
	Initialized        bool                `json:"initialized"`
}

type CellInfo struct {
	CellUUID        string `json:"cell_uuid"`
	ParcelAutomatID string `json:"parcel_automat_id"`
}

type PrepareCellRequest struct {
	CellID string `json:"cell_id" binding:"required"`
}

type PrepareCellResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	CellNumber int    `json:"cell_number"`
	CellUUID   string `json:"cell_uuid"`
	Type       string `json:"type"`
}

type CellsCountResponse struct {
	CellsCount          int `json:"cells_count"`
	MappedCells         int `json:"mapped_cells"`
	InternalCellsCount  int `json:"internal_cells_count"`
	MappedInternalCells int `json:"mapped_internal_cells"`
}
