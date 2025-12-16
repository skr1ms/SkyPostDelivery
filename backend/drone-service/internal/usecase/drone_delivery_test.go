package usecase

import (
	"context"
	"testing"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockVideoHandler struct {
	called     bool
	droneID    string
	frameData  []byte
	deliveryID string
}

func (m *mockVideoHandler) HandleVideoFrame(ctx context.Context, droneID string, frameData []byte, deliveryID string) error {
	m.called = true
	m.droneID = droneID
	m.frameData = frameData
	m.deliveryID = deliveryID
	return nil
}

func TestDroneDeliveryUseCase_ProcessDeliveryUpdate_ArrivedAtLocker(t *testing.T) {
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	uc := NewDroneDeliveryUseCase(mockDeliveryRepo, mockDroneManager, nil, nil, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"
	deliveryID := "delivery-456"
	payload := map[string]any{
		"drone_status": "arrived_at_locker",
		"delivery_id":  deliveryID,
	}

	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDeliveryRepo.On("UpdateDeliveryStatus", ctx, deliveryID, entity.DeliveryStatusInProgress, (*string)(nil)).Return(nil)

	err := uc.ProcessDeliveryUpdate(ctx, droneID, payload)

	assert.NoError(t, err)
	mockDeliveryRepo.AssertExpectations(t)
}

func TestDroneDeliveryUseCase_ProcessDeliveryUpdate_Returning(t *testing.T) {
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	uc := NewDroneDeliveryUseCase(mockDeliveryRepo, mockDroneManager, nil, nil, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"
	deliveryID := "delivery-456"
	state := &entity.DroneState{
		DroneID:           droneID,
		Status:            entity.DroneStatusDelivering,
		BatteryLevel:      85.0,
		CurrentDeliveryID: stringPtr(deliveryID),
	}

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(state, nil)
	mockDroneRepo.On("SaveDroneState", ctx, mock.MatchedBy(func(s *entity.DroneState) bool {
		return s.Status == entity.DroneStatusIdle && s.CurrentDeliveryID == nil
	})).Return(nil)

	payload := map[string]any{
		"drone_status": "returning",
		"delivery_id":  deliveryID,
	}

	err := uc.ProcessDeliveryUpdate(ctx, droneID, payload)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneDeliveryUseCase_ProcessVideoFrame_Success(t *testing.T) {
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)
	mockVideoHandler := &mockVideoHandler{}

	uc := NewDroneDeliveryUseCase(mockDeliveryRepo, mockDroneManager, nil, mockVideoHandler, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"
	frameData := []byte("test-frame-data")
	deliveryID := "delivery-456"

	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	err := uc.ProcessVideoFrame(ctx, droneID, frameData, deliveryID)

	assert.NoError(t, err)
	assert.True(t, mockVideoHandler.called)
	assert.Equal(t, droneID, mockVideoHandler.droneID)
	assert.Equal(t, frameData, mockVideoHandler.frameData)
	assert.Equal(t, deliveryID, mockVideoHandler.deliveryID)
}

func TestDroneDeliveryUseCase_ProcessVideoFrame_NilHandler(t *testing.T) {
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)

	uc := NewDroneDeliveryUseCase(mockDeliveryRepo, mockDroneManager, nil, nil, mockLogger)

	ctx := context.Background()
	droneID := "drone-123"
	frameData := []byte("test-frame-data")
	deliveryID := "delivery-456"

	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	err := uc.ProcessVideoFrame(ctx, droneID, frameData, deliveryID)

	assert.NoError(t, err)
}
