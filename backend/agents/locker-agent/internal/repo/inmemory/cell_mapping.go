package inmemory

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity/error"
)

type CellMappingRepo struct {
	mu            sync.RWMutex
	automatID     uuid.UUID
	externalCells map[int]uuid.UUID
	internalCells map[int]uuid.UUID
	initialized   bool
	lastSyncTime  time.Time
}

func NewCellMappingRepo() *CellMappingRepo {
	return &CellMappingRepo{
		externalCells: make(map[int]uuid.UUID),
		internalCells: make(map[int]uuid.UUID),
		initialized:   false,
	}
}

func (r *CellMappingRepo) Sync(automatID uuid.UUID, external, internal []uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.automatID = automatID

	r.externalCells = make(map[int]uuid.UUID)
	for i, cellUUID := range external {
		r.externalCells[i+1] = cellUUID
	}

	r.internalCells = make(map[int]uuid.UUID)
	for i, cellUUID := range internal {
		r.internalCells[i+1] = cellUUID
	}

	r.initialized = true
	r.lastSyncTime = time.Now()

	return nil
}

func (r *CellMappingRepo) IsInitialized() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.initialized
}

func (r *CellMappingRepo) GetCellUUID(cellNumber int) (uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return uuid.Nil, entityError.ErrCellNotInitialized
	}

	cellUUID, exists := r.externalCells[cellNumber]
	if !exists {
		return uuid.Nil, entityError.ErrCellNotFound
	}

	return cellUUID, nil
}

func (r *CellMappingRepo) GetInternalCellUUID(doorNumber int) (uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return uuid.Nil, entityError.ErrCellNotInitialized
	}

	cellUUID, exists := r.internalCells[doorNumber]
	if !exists {
		return uuid.Nil, entityError.ErrCellNotFound
	}

	return cellUUID, nil
}

func (r *CellMappingRepo) GetCellNumber(cellUUID uuid.UUID) (int, entity.CellType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for num, uuid := range r.externalCells {
		if uuid == cellUUID {
			return num, entity.CellTypeExternal, nil
		}
	}

	for num, uuid := range r.internalCells {
		if uuid == cellUUID {
			return num, entity.CellTypeInternal, nil
		}
	}

	return 0, "", entityError.ErrCellNotFound
}

func (r *CellMappingRepo) GetAllExternalCells() map[int]uuid.UUID {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cells := make(map[int]uuid.UUID, len(r.externalCells))
	for k, v := range r.externalCells {
		cells[k] = v
	}
	return cells
}

func (r *CellMappingRepo) GetAllInternalCells() map[int]uuid.UUID {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cells := make(map[int]uuid.UUID, len(r.internalCells))
	for k, v := range r.internalCells {
		cells[k] = v
	}
	return cells
}

func (r *CellMappingRepo) GetParcelAutomatID() uuid.UUID {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.automatID
}

func (r *CellMappingRepo) GetMapping() *entity.CellMapping {
	r.mu.RLock()
	defer r.mu.RUnlock()

	externalCells := make(map[int]uuid.UUID, len(r.externalCells))
	for k, v := range r.externalCells {
		externalCells[k] = v
	}

	internalCells := make(map[int]uuid.UUID, len(r.internalCells))
	for k, v := range r.internalCells {
		internalCells[k] = v
	}

	return &entity.CellMapping{
		ParcelAutomatID: r.automatID,
		ExternalCells:   externalCells,
		InternalCells:   internalCells,
		LastSyncTime:    r.lastSyncTime,
		Initialized:     r.initialized,
	}
}

func (r *CellMappingRepo) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.automatID = uuid.Nil
	r.externalCells = make(map[int]uuid.UUID)
	r.internalCells = make(map[int]uuid.UUID)
	r.initialized = false
	r.lastSyncTime = time.Time{}
}
