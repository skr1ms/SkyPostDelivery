package inmemory

import (
	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity"
)

type CellMappingInterface interface {
	Sync(automatID uuid.UUID, external, internal []uuid.UUID) error
	GetCellUUID(cellNumber int) (uuid.UUID, error)
	GetInternalCellUUID(doorNumber int) (uuid.UUID, error)
	GetCellNumber(cellUUID uuid.UUID) (int, entity.CellType, error)
	GetMapping() *entity.CellMapping
	IsInitialized() bool
}
