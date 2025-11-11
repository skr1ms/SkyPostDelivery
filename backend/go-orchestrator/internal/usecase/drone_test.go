package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/skr1ms/hitech-ekb/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestDroneUseCase_GetStatus_Success(t *testing.T) {
	mockDroneRepo := new(MockDroneRepo)

	uc := NewDroneUseCase(mockDroneRepo)

	ctx := context.Background()
	droneID := uuid.New()

	drone := &entity.Drone{
		ID:     droneID,
		Status: "flying",
	}

	mockDroneRepo.On("GetByID", ctx, droneID).Return(drone, nil)

	result, err := uc.GetStatus(ctx, droneID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, droneID, result.DroneID)
	assert.Equal(t, "flying", result.Status)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_GetStatus_Error(t *testing.T) {
	mockDroneRepo := new(MockDroneRepo)

	uc := NewDroneUseCase(mockDroneRepo)

	ctx := context.Background()
	droneID := uuid.New()

	mockDroneRepo.On("GetByID", ctx, droneID).Return(nil, errors.New("drone not found"))

	result, err := uc.GetStatus(ctx, droneID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "drone not found")
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_ListDrones_Success(t *testing.T) {
	mockDroneRepo := new(MockDroneRepo)

	uc := NewDroneUseCase(mockDroneRepo)

	ctx := context.Background()

	drones := []*entity.Drone{
		{ID: uuid.New(), Model: "DJI-1", Status: "idle"},
		{ID: uuid.New(), Model: "DJI-2", Status: "flying"},
	}

	mockDroneRepo.On("List", ctx).Return(drones, nil)

	result, err := uc.ListDrones(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_UpdateDroneStatus_Success(t *testing.T) {
	mockDroneRepo := new(MockDroneRepo)

	uc := NewDroneUseCase(mockDroneRepo)

	ctx := context.Background()
	droneID := uuid.New()
	status := "flying"

	mockDroneRepo.On("UpdateStatus", ctx, droneID, status).Return(nil)

	err := uc.UpdateDroneStatus(ctx, droneID, status)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_UpdateDroneStatus_Error(t *testing.T) {
	mockDroneRepo := new(MockDroneRepo)

	uc := NewDroneUseCase(mockDroneRepo)

	ctx := context.Background()
	droneID := uuid.New()
	status := "flying"

	mockDroneRepo.On("UpdateStatus", ctx, droneID, status).Return(errors.New("database error"))

	err := uc.UpdateDroneStatus(ctx, droneID, status)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_CreateDrone_Success(t *testing.T) {
	mockDroneRepo := new(MockDroneRepo)

	uc := NewDroneUseCase(mockDroneRepo)

	ctx := context.Background()
	model := "DJI-Phantom-5"
	ipAddress := "192.168.10.1"
	expectedDrone := &entity.Drone{
		ID:        uuid.New(),
		Model:     model,
		IPAddress: ipAddress,
		Status:    "idle",
	}

	mockDroneRepo.On("Create", ctx, model, "idle", ipAddress).Return(expectedDrone, nil)

	drone, err := uc.CreateDrone(ctx, model, ipAddress)

	assert.NoError(t, err)
	assert.NotNil(t, drone)
	assert.Equal(t, model, drone.Model)
	assert.Equal(t, ipAddress, drone.IPAddress)
	assert.Equal(t, "idle", drone.Status)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_CreateDrone_Error(t *testing.T) {
	mockDroneRepo := new(MockDroneRepo)

	uc := NewDroneUseCase(mockDroneRepo)

	ctx := context.Background()
	model := "DJI-Phantom-5"
	ipAddress := "192.168.10.1"

	mockDroneRepo.On("Create", ctx, model, "idle", ipAddress).Return(nil, errors.New("database error"))

	drone, err := uc.CreateDrone(ctx, model, ipAddress)

	assert.Error(t, err)
	assert.Nil(t, drone)
	assert.Contains(t, err.Error(), "database error")
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_UpdateDrone_Success(t *testing.T) {
	mockDroneRepo := new(MockDroneRepo)

	uc := NewDroneUseCase(mockDroneRepo)

	ctx := context.Background()
	droneID := uuid.New()
	status := "charging"
	expectedDrone := &entity.Drone{
		ID:     droneID,
		Model:  "DJI-Phantom-5",
		Status: status,
	}

	mockDroneRepo.On("UpdateStatus", ctx, droneID, status).Return(nil)
	mockDroneRepo.On("GetByID", ctx, droneID).Return(expectedDrone, nil)

	drone, err := uc.UpdateDrone(ctx, droneID, status)

	assert.NoError(t, err)
	assert.NotNil(t, drone)
	assert.Equal(t, droneID, drone.ID)
	assert.Equal(t, status, drone.Status)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_UpdateDrone_UpdateError(t *testing.T) {
	mockDroneRepo := new(MockDroneRepo)

	uc := NewDroneUseCase(mockDroneRepo)

	ctx := context.Background()
	droneID := uuid.New()
	status := "charging"

	mockDroneRepo.On("UpdateStatus", ctx, droneID, status).Return(errors.New("update failed"))

	drone, err := uc.UpdateDrone(ctx, droneID, status)

	assert.Error(t, err)
	assert.Nil(t, drone)
	assert.Contains(t, err.Error(), "update failed")
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_DeleteDrone_Success(t *testing.T) {
	mockDroneRepo := new(MockDroneRepo)

	uc := NewDroneUseCase(mockDroneRepo)

	ctx := context.Background()
	droneID := uuid.New()

	mockDroneRepo.On("Delete", ctx, droneID).Return(nil)

	err := uc.DeleteDrone(ctx, droneID)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_DeleteDrone_Error(t *testing.T) {
	mockDroneRepo := new(MockDroneRepo)

	uc := NewDroneUseCase(mockDroneRepo)

	ctx := context.Background()
	droneID := uuid.New()

	mockDroneRepo.On("Delete", ctx, droneID).Return(errors.New("drone in use"))

	err := uc.DeleteDrone(ctx, droneID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "drone in use")
	mockDroneRepo.AssertExpectations(t)
}
