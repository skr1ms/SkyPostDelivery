package usecase

import (
	"context"
	"testing"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDroneTelemetryUseCase_ProcessHeartbeat_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)

	uc := NewDroneTelemetryUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"
	payload := map[string]any{
		"status":              "flying",
		"battery_level":       75.5,
		"current_delivery_id": "delivery-456",
		"position": map[string]any{
			"latitude":  55.7558,
			"longitude": 37.6173,
			"altitude":  100.0,
		},
		"speed": 15.5,
	}

	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDroneRepo.On("UpdateDroneBattery", ctx, droneID, mock.AnythingOfType("float64")).Return(nil)
	mockDroneRepo.On("SaveDroneState", ctx, mock.MatchedBy(func(s *entity.DroneState) bool {
		return s.DroneID == droneID &&
			s.Status == entity.DroneStatus("flying") &&
			s.CurrentDeliveryID != nil &&
			*s.CurrentDeliveryID == "delivery-456"
	})).Return(nil)

	err := uc.ProcessHeartbeat(ctx, droneID, payload)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneTelemetryUseCase_ProcessHeartbeat_DefaultValues(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)

	uc := NewDroneTelemetryUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"
	payload := map[string]any{}

	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDroneRepo.On("UpdateDroneBattery", ctx, droneID, 100.0).Return(nil)
	mockDroneRepo.On("SaveDroneState", ctx, mock.MatchedBy(func(s *entity.DroneState) bool {
		return s.DroneID == droneID &&
			s.Status == entity.DroneStatus("idle") &&
			s.BatteryLevel == 100.0
	})).Return(nil)

	err := uc.ProcessHeartbeat(ctx, droneID, payload)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneTelemetryUseCase_ProcessStatusUpdate_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockLogger := mocks.NewMockLogger(t)

	uc := NewDroneTelemetryUseCase(mockDroneRepo, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"
	payload := map[string]any{
		"status":              "delivering",
		"battery_level":       80.0,
		"current_delivery_id": "delivery-789",
		"position": map[string]any{
			"latitude":  55.7558,
			"longitude": 37.6173,
			"altitude":  150.0,
		},
		"speed": 20.0,
	}

	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDroneRepo.On("SaveDroneState", ctx, mock.MatchedBy(func(s *entity.DroneState) bool {
		return s.DroneID == droneID &&
			s.Status == entity.DroneStatus("delivering")
	})).Return(nil)

	err := uc.ProcessStatusUpdate(ctx, droneID, payload)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}
