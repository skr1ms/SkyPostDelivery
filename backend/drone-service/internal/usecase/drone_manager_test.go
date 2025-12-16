package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDroneManagerUseCase_RegisterDrone_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	uc := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"

	err := uc.RegisterDrone(ctx, droneID)

	assert.NoError(t, err)
	drones := uc.GetRegisteredDrones()
	assert.Contains(t, drones, droneID)
}

func TestDroneManagerUseCase_UnregisterDrone_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	uc := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"

	_ = uc.RegisterDrone(ctx, droneID)
	assert.Contains(t, uc.GetRegisteredDrones(), droneID)

	err := uc.UnregisterDrone(ctx, droneID)
	assert.NoError(t, err)
	assert.NotContains(t, uc.GetRegisteredDrones(), droneID)
}

func TestDroneManagerUseCase_GetFreeDrone_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	uc := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"

	_ = uc.RegisterDrone(ctx, droneID)

	state := &entity.DroneState{
		DroneID:           droneID,
		Status:            entity.DroneStatusIdle,
		BatteryLevel:      85.0,
		CurrentDeliveryID: nil,
	}

	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(state, nil)

	freeDrone, err := uc.GetFreeDrone(ctx)

	assert.NoError(t, err)
	assert.Equal(t, droneID, freeDrone)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneManagerUseCase_GetFreeDrone_NoFreeDrone(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	uc := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"

	_ = uc.RegisterDrone(ctx, droneID)

	state := &entity.DroneState{
		DroneID:           droneID,
		Status:            entity.DroneStatusDelivering,
		BatteryLevel:      85.0,
		CurrentDeliveryID: stringPtr("delivery-123"),
	}

	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(state, nil)

	freeDrone, err := uc.GetFreeDrone(ctx)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, entityError.ErrDroneNotAvailable))
	assert.Empty(t, freeDrone)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneManagerUseCase_GetFreeDrone_LowBattery(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	uc := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"

	_ = uc.RegisterDrone(ctx, droneID)

	state := &entity.DroneState{
		DroneID:           droneID,
		Status:            entity.DroneStatusIdle,
		BatteryLevel:      20.0,
		CurrentDeliveryID: nil,
	}

	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(state, nil)

	freeDrone, err := uc.GetFreeDrone(ctx)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, entityError.ErrDroneNotAvailable))
	assert.Empty(t, freeDrone)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneManagerUseCase_AssignDeliveryToDrone_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	uc := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"
	deliveryID := "delivery-456"

	state := &entity.DroneState{
		DroneID:           droneID,
		Status:            entity.DroneStatusIdle,
		BatteryLevel:      85.0,
		CurrentDeliveryID: nil,
	}

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(state, nil)
	mockDroneRepo.On("SaveDroneState", ctx, mock.MatchedBy(func(s *entity.DroneState) bool {
		return s.DroneID == droneID &&
			s.Status == entity.DroneStatusTakingOff &&
			s.CurrentDeliveryID != nil &&
			*s.CurrentDeliveryID == deliveryID
	})).Return(nil)

	err := uc.AssignDeliveryToDrone(ctx, droneID, deliveryID)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneManagerUseCase_AssignDeliveryToDrone_GetStateError(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	uc := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"
	deliveryID := "delivery-456"

	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(nil, errors.New("database error"))

	err := uc.AssignDeliveryToDrone(ctx, droneID, deliveryID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GetDroneState")
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneManagerUseCase_ReleaseDrone_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	uc := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"
	deliveryID := "delivery-456"

	state := &entity.DroneState{
		DroneID:           droneID,
		Status:            entity.DroneStatusDelivering,
		BatteryLevel:      85.0,
		CurrentDeliveryID: &deliveryID,
	}

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(state, nil)
	mockDroneRepo.On("SaveDroneState", ctx, mock.MatchedBy(func(s *entity.DroneState) bool {
		return s.DroneID == droneID &&
			s.Status == entity.DroneStatusIdle &&
			s.CurrentDeliveryID == nil
	})).Return(nil)

	err := uc.ReleaseDrone(ctx, droneID)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneManagerUseCase_GetDroneState_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	uc := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"

	expectedState := &entity.DroneState{
		DroneID:      droneID,
		Status:       entity.DroneStatusIdle,
		BatteryLevel: 85.0,
	}

	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(expectedState, nil)

	state, err := uc.GetDroneState(ctx, droneID)

	assert.NoError(t, err)
	assert.Equal(t, expectedState, state)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneManagerUseCase_GetDroneState_Error(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	uc := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"

	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(nil, errors.New("database error"))

	state, err := uc.GetDroneState(ctx, droneID)

	assert.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "GetDroneState")
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneManagerUseCase_GetAllDrones(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	uc := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()

	_ = uc.RegisterDrone(ctx, "drone-1")
	_ = uc.RegisterDrone(ctx, "drone-2")
	_ = uc.RegisterDrone(ctx, "drone-3")

	drones := uc.GetAllDrones()

	assert.Len(t, drones, 3)
	assert.Contains(t, drones, "drone-1")
	assert.Contains(t, drones, "drone-2")
	assert.Contains(t, drones, "drone-3")
}
