package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDroneUseCase_GetStatus_Success(t *testing.T) {
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockLogger := new(mocks.MockLogger)

	uc := NewDroneUseCase(mockDroneRepo, mockLogger)

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
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockLogger := new(mocks.MockLogger)

	uc := NewDroneUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := uuid.New()

	mockDroneRepo.On("GetByID", ctx, droneID).Return(nil, errors.New("drone not found"))

	result, err := uc.GetStatus(ctx, droneID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "drone not found")
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_List_Success(t *testing.T) {
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockLogger := new(mocks.MockLogger)

	uc := NewDroneUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()

	drones := []*entity.Drone{
		{ID: uuid.New(), Model: "DJI-1", Status: "idle"},
		{ID: uuid.New(), Model: "DJI-2", Status: "flying"},
	}

	mockDroneRepo.On("List", ctx).Return(drones, nil)

	result, err := uc.List(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_UpdateStatus_Success(t *testing.T) {
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockLogger := new(mocks.MockLogger)

	uc := NewDroneUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := uuid.New()
	status := "busy"

	drone := &entity.Drone{ID: droneID, Status: "idle"}
	mockDroneRepo.On("GetByID", ctx, droneID).Return(drone, nil)
	mockDroneRepo.On("UpdateStatus", ctx, mock.MatchedBy(func(d *entity.Drone) bool {
		return d.ID == droneID && d.Status == status
	})).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	err := uc.UpdateStatus(ctx, droneID, status)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_UpdateStatus_Error(t *testing.T) {
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockLogger := new(mocks.MockLogger)

	uc := NewDroneUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := uuid.New()
	status := "busy"

	drone := &entity.Drone{ID: droneID, Status: "idle"}
	mockDroneRepo.On("GetByID", ctx, droneID).Return(drone, nil)
	mockDroneRepo.On("UpdateStatus", ctx, mock.MatchedBy(func(d *entity.Drone) bool {
		return d.ID == droneID && d.Status == status
	})).Return(errors.New("database error"))

	err := uc.UpdateStatus(ctx, droneID, status)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_Create_Success(t *testing.T) {
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockLogger := new(mocks.MockLogger)

	uc := NewDroneUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	model := "DJI-Phantom-5"
	ipAddress := "192.168.10.1"
	expectedDrone := &entity.Drone{
		ID:        uuid.New(),
		Model:     model,
		IPAddress: ipAddress,
		Status:    "idle",
	}

	mockDroneRepo.On("Create", ctx, mock.MatchedBy(func(d *entity.Drone) bool {
		return d.Model == model && d.IPAddress == ipAddress && d.Status == "idle"
	})).Return(expectedDrone, nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	drone, err := uc.Create(ctx, model, ipAddress)

	assert.NoError(t, err)
	assert.NotNil(t, drone)
	assert.Equal(t, model, drone.Model)
	assert.Equal(t, ipAddress, drone.IPAddress)
	assert.Equal(t, "idle", drone.Status)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_Create_Error(t *testing.T) {
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockLogger := new(mocks.MockLogger)

	uc := NewDroneUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	model := "DJI-Phantom-5"
	ipAddress := "192.168.10.1"

	mockDroneRepo.On("Create", ctx, mock.MatchedBy(func(d *entity.Drone) bool {
		return d.Model == model && d.IPAddress == ipAddress && d.Status == "idle"
	})).Return(nil, errors.New("database error"))

	drone, err := uc.Create(ctx, model, ipAddress)

	assert.Error(t, err)
	assert.Nil(t, drone)
	assert.Contains(t, err.Error(), "database error")
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_Update_Success(t *testing.T) {
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockLogger := new(mocks.MockLogger)

	uc := NewDroneUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := uuid.New()
	status := "charging"
	expectedDrone := &entity.Drone{
		ID:     droneID,
		Model:  "DJI-Phantom-5",
		Status: "idle",
	}

	updatedDrone := &entity.Drone{
		ID:     droneID,
		Model:  "DJI-Phantom-5",
		Status: status,
	}

	mockDroneRepo.On("GetByID", ctx, droneID).Return(expectedDrone, nil)
	mockDroneRepo.On("Update", ctx, mock.MatchedBy(func(d *entity.Drone) bool {
		return d.ID == droneID && d.Model == expectedDrone.Model && d.IPAddress == expectedDrone.IPAddress && d.Status == status
	})).Return(updatedDrone, nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	drone, err := uc.Update(ctx, droneID, "", "", status)

	assert.NoError(t, err)
	assert.NotNil(t, drone)
	assert.Equal(t, droneID, drone.ID)
	assert.Equal(t, status, drone.Status)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_Update_UpdateError(t *testing.T) {
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockLogger := new(mocks.MockLogger)

	uc := NewDroneUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := uuid.New()
	status := "charging"

	expectedDrone := &entity.Drone{
		ID:     droneID,
		Model:  "DJI-Phantom-5",
		Status: "idle",
	}

	mockDroneRepo.On("GetByID", ctx, droneID).Return(expectedDrone, nil)
	mockDroneRepo.On("Update", ctx, mock.MatchedBy(func(d *entity.Drone) bool {
		return d.ID == droneID && d.Model == expectedDrone.Model && d.IPAddress == expectedDrone.IPAddress && d.Status == status
	})).Return(nil, errors.New("update failed"))

	drone, err := uc.Update(ctx, droneID, "", "", status)

	assert.Error(t, err)
	assert.Nil(t, drone)
	assert.Contains(t, err.Error(), "update failed")
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_Delete_Success(t *testing.T) {
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockLogger := new(mocks.MockLogger)

	uc := NewDroneUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := uuid.New()

	mockDroneRepo.On("GetByID", ctx, droneID).Return(&entity.Drone{ID: droneID, Status: "idle"}, nil)
	mockDroneRepo.On("Delete", ctx, droneID).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	err := uc.Delete(ctx, droneID)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneUseCase_Delete_Error(t *testing.T) {
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockLogger := new(mocks.MockLogger)

	uc := NewDroneUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := uuid.New()

	mockDroneRepo.On("GetByID", ctx, droneID).Return(&entity.Drone{ID: droneID, Status: "idle"}, nil)
	mockDroneRepo.On("Delete", ctx, droneID).Return(errors.New("drone in use"))

	err := uc.Delete(ctx, droneID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "drone in use")
	mockDroneRepo.AssertExpectations(t)
}
