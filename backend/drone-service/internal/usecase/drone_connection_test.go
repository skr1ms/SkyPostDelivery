package usecase

import (
	"context"
	"errors"
	"testing"

	entityError "github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDroneConnectionUseCase_RegisterDrone_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	uc := NewDroneConnectionUseCase(mockDroneRepo, mockDroneManager, mockLogger)

	ctx := context.Background()
	ipAddress := "192.168.1.100"
	droneID := "drone-123"

	mockDroneRepo.On("GetDroneIDByIP", ctx, ipAddress).Return(droneID, nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	result, err := uc.RegisterDrone(ctx, ipAddress)

	assert.NoError(t, err)
	assert.Equal(t, droneID, result)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneConnectionUseCase_RegisterDrone_NotFound(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	uc := NewDroneConnectionUseCase(mockDroneRepo, mockDroneManager, mockLogger)

	ctx := context.Background()
	ipAddress := "192.168.1.100"

	mockDroneRepo.On("GetDroneIDByIP", ctx, ipAddress).Return("", nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	result, err := uc.RegisterDrone(ctx, ipAddress)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, entityError.ErrDroneNotFound))
	assert.Empty(t, result)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneConnectionUseCase_UnregisterDrone_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	uc := NewDroneConnectionUseCase(mockDroneRepo, mockDroneManager, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"

	_ = mockDroneManager.RegisterDrone(ctx, droneID)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	err := uc.UnregisterDrone(ctx, droneID)

	assert.NoError(t, err)
	assert.NotContains(t, mockDroneManager.GetRegisteredDrones(), droneID)
}

func TestDroneConnectionUseCase_GetDroneIDByIP_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	uc := NewDroneConnectionUseCase(mockDroneRepo, mockDroneManager, mockLogger)

	ctx := context.Background()
	ipAddress := "192.168.1.100"
	droneID := "drone-123"

	mockDroneRepo.On("GetDroneIDByIP", ctx, ipAddress).Return(droneID, nil)

	result, err := uc.GetDroneIDByIP(ctx, ipAddress)

	assert.NoError(t, err)
	assert.Equal(t, droneID, result)
	mockDroneRepo.AssertExpectations(t)
}
