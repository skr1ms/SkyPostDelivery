package repo

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
)

type MemoryStateRepo struct {
	drones     map[string]*entity.DroneState
	deliveries map[string]*entity.DeliveryTask
	droneIPMap map[string]string
	mu         sync.RWMutex
}

func NewMemoryStateRepo() *MemoryStateRepo {
	return &MemoryStateRepo{
		drones:     make(map[string]*entity.DroneState),
		deliveries: make(map[string]*entity.DeliveryTask),
		droneIPMap: make(map[string]string),
	}
}

func (r *MemoryStateRepo) SaveDroneState(ctx context.Context, state *entity.DroneState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	stateCopy := *state
	r.drones[state.DroneID] = &stateCopy
	return nil
}

func (r *MemoryStateRepo) GetDroneState(ctx context.Context, droneID string) (*entity.DroneState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, exists := r.drones[droneID]
	if !exists {
		return nil, nil
	}

	stateCopy := *state
	return &stateCopy, nil
}

func (r *MemoryStateRepo) GetDroneIDByIP(ctx context.Context, ipAddress string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	droneID, exists := r.droneIPMap[ipAddress]
	if !exists {
		return "", nil
	}

	return droneID, nil
}

func (r *MemoryStateRepo) UpdateDroneBattery(ctx context.Context, droneID string, batteryLevel float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, exists := r.drones[droneID]
	if exists && state != nil {
		state.BatteryLevel = batteryLevel
	}

	return nil
}

func (r *MemoryStateRepo) SaveDeliveryTask(ctx context.Context, task *entity.DeliveryTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	taskCopy := *task
	r.deliveries[task.DeliveryID] = &taskCopy
	return nil
}

func (r *MemoryStateRepo) GetDeliveryTask(ctx context.Context, deliveryID string) (*entity.DeliveryTask, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.deliveries[deliveryID]
	if !exists {
		return nil, nil
	}

	taskCopy := *task
	return &taskCopy, nil
}

func (r *MemoryStateRepo) UpdateDeliveryStatus(ctx context.Context, deliveryID string, status entity.DeliveryStatus, errorMessage *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.deliveries[deliveryID]
	if exists && task != nil {
		task.Status = status
		if errorMessage != nil {
			task.ErrorMessage = errorMessage
		}
	}

	return nil
}

func (r *MemoryStateRepo) MarshalJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data := map[string]interface{}{
		"drones":     r.drones,
		"deliveries": r.deliveries,
	}

	return json.Marshal(data)
}
