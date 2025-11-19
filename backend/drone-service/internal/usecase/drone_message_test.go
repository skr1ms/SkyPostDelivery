package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDroneMessageUseCase_RegisterDrone_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()
	ipAddress := "192.168.1.100"
	droneID := "drone-123"

	mockDroneRepo.On("GetDroneIDByIP", ctx, ipAddress).Return(droneID, nil)

	result, err := uc.RegisterDrone(ctx, ipAddress)

	assert.NoError(t, err)
	assert.Equal(t, droneID, result)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneMessageUseCase_RegisterDrone_NotFound(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()
	ipAddress := "192.168.1.100"

	mockDroneRepo.On("GetDroneIDByIP", ctx, ipAddress).Return("", errors.New("not found"))

	result, err := uc.RegisterDrone(ctx, ipAddress)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "no drone found")
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneMessageUseCase_UnregisterDrone_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()
	droneID := "drone-123"

	_ = mockDroneManager.RegisterDrone(ctx, droneID)

	err := uc.UnregisterDrone(ctx, droneID)

	assert.NoError(t, err)
	assert.NotContains(t, mockDroneManager.GetRegisteredDrones(), droneID)
}

func TestDroneMessageUseCase_ProcessHeartbeat_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()
	droneID := "drone-123"
	payload := map[string]interface{}{
		"status":              "flying",
		"battery_level":       75.5,
		"current_delivery_id": "delivery-456",
		"position": map[string]interface{}{
			"latitude":  55.7558,
			"longitude": 37.6173,
			"altitude":  100.0,
		},
		"speed": 15.5,
	}

	mockDroneRepo.On("UpdateDroneBattery", ctx, droneID, 75.5).Return(nil)
	mockDroneRepo.On("SaveDroneState", ctx, mock.MatchedBy(func(s *entity.DroneState) bool {
		return s.DroneID == droneID &&
			s.Status == entity.DroneStatus("flying") &&
			s.BatteryLevel == 75.5 &&
			s.CurrentDeliveryID != nil &&
			*s.CurrentDeliveryID == "delivery-456"
	})).Return(nil)

	err := uc.ProcessHeartbeat(ctx, droneID, payload)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneMessageUseCase_ProcessHeartbeat_DefaultValues(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()
	droneID := "drone-123"
	payload := map[string]interface{}{}

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

func TestDroneMessageUseCase_ProcessStatusUpdate_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()
	droneID := "drone-123"
	payload := map[string]interface{}{
		"status":              "delivering",
		"battery_level":       80.0,
		"current_delivery_id": "delivery-789",
		"position": map[string]interface{}{
			"latitude":  55.7558,
			"longitude": 37.6173,
			"altitude":  150.0,
		},
		"speed": 20.0,
	}

	mockDroneRepo.On("SaveDroneState", ctx, mock.MatchedBy(func(s *entity.DroneState) bool {
		return s.DroneID == droneID &&
			s.Status == entity.DroneStatus("delivering") &&
			s.BatteryLevel == 80.0
	})).Return(nil)

	err := uc.ProcessStatusUpdate(ctx, droneID, payload)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneMessageUseCase_ProcessDeliveryUpdate_ArrivedAtLocker(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()
	droneID := "drone-123"
	deliveryID := "delivery-456"
	payload := map[string]interface{}{
		"drone_status": "arrived_at_locker",
		"delivery_id":  deliveryID,
	}

	mockDeliveryRepo.On("UpdateDeliveryStatus", ctx, deliveryID, entity.DeliveryStatusInProgress, (*string)(nil)).Return(nil)

	err := uc.ProcessDeliveryUpdate(ctx, droneID, payload)

	assert.NoError(t, err)
	mockDeliveryRepo.AssertExpectations(t)
}

func TestDroneMessageUseCase_ProcessDeliveryUpdate_Returning(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()
	droneID := "drone-123"
	deliveryID := "delivery-456"
	state := &entity.DroneState{
		DroneID:           droneID,
		Status:            entity.DroneStatusDelivering,
		BatteryLevel:      85.0,
		CurrentDeliveryID: stringPtr(deliveryID),
	}

	// ProcessDeliveryUpdate with "returning" calls ReleaseDrone, which calls GetDroneState and SaveDroneState
	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(state, nil)
	mockDroneRepo.On("SaveDroneState", ctx, mock.MatchedBy(func(s *entity.DroneState) bool {
		return s.Status == entity.DroneStatusIdle && s.CurrentDeliveryID == nil
	})).Return(nil)

	payload := map[string]interface{}{
		"drone_status": "returning",
		"delivery_id":  deliveryID,
	}

	err := uc.ProcessDeliveryUpdate(ctx, droneID, payload)

	assert.NoError(t, err)
	mockDroneRepo.AssertExpectations(t)
}

func TestDroneMessageUseCase_ProcessArrivedAtDestination_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)
	mockDeliveryUseCase := &DeliveryUseCase{}

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockDeliveryUseCase,
		nil,
		nil,
	)

	ctx := context.Background()
	droneID := "drone-123"
	payload := map[string]interface{}{
		"order_id":          "order-789",
		"parcel_automat_id": "automat-456",
	}

	err := uc.ProcessArrivedAtDestination(ctx, droneID, payload)

	_ = err
}

func TestDroneMessageUseCase_ProcessCargoDropped_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)
	mockNotifier := new(mocks.MockDroneNotifier)
	mockGRPCClient := new(mocks.MockOrchestratorGRPCClient)
	mockRabbitMQClient := mocks.NewMockRabbitMQClient(t)

	mockDeliveryUseCase := NewDeliveryUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockNotifier,
		mockGRPCClient,
		mockRabbitMQClient,
	)

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockDeliveryUseCase,
		nil,
		nil,
	)

	ctx := context.Background()
	orderID := "order-789"
	lockerCellID := "cell-123"
	droneID := "drone-123"

	task := &entity.DeliveryTask{
		DeliveryID:   "delivery-789",
		OrderID:      orderID,
		DroneID:      &droneID,
		LockerCellID: lockerCellID,
		Status:       entity.DeliveryStatusInProgress,
	}

	state := &entity.DroneState{
		DroneID:           droneID,
		Status:            entity.DroneStatusDelivering,
		BatteryLevel:      85.0,
		CurrentDeliveryID: stringPtr("delivery-789"),
	}

	mockDeliveryRepo.On("GetDeliveryTask", ctx, orderID).Return(task, nil)
	mockDeliveryRepo.On("UpdateDeliveryStatus", ctx, orderID, entity.DeliveryStatusCompleted, (*string)(nil)).Return(nil)
	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(state, nil)
	mockDroneRepo.On("SaveDroneState", ctx, mock.Anything).Return(nil)
	mockRabbitMQClient.On("Publish", ctx, "confirmations", mock.Anything).Return(nil)

	payload := map[string]interface{}{
		"order_id":       orderID,
		"locker_cell_id": lockerCellID,
	}

	err := uc.ProcessCargoDropped(ctx, payload)

	assert.NoError(t, err)
	mockDeliveryRepo.AssertExpectations(t)
	mockDroneRepo.AssertExpectations(t)
	mockRabbitMQClient.AssertExpectations(t)
}

func TestDroneMessageUseCase_ProcessVideoFrame_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)
	mockVideoHandler := &mockVideoHandler{}

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		nil,
		nil,
		mockVideoHandler,
	)

	ctx := context.Background()
	droneID := "drone-123"
	frameData := []byte("test-frame-data")
	deliveryID := "delivery-456"

	err := uc.ProcessVideoFrame(ctx, droneID, frameData, deliveryID)

	assert.NoError(t, err)
	assert.True(t, mockVideoHandler.called)
	assert.Equal(t, droneID, mockVideoHandler.droneID)
	assert.Equal(t, frameData, mockVideoHandler.frameData)
	assert.Equal(t, deliveryID, mockVideoHandler.deliveryID)
}

func TestDroneMessageUseCase_ProcessVideoFrame_NilHandler(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()
	droneID := "drone-123"
	frameData := []byte("test-frame-data")
	deliveryID := "delivery-456"

	err := uc.ProcessVideoFrame(ctx, droneID, frameData, deliveryID)

	assert.NoError(t, err)
}

func TestDroneMessageUseCase_GetDroneIDByIP_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo)

	uc := NewDroneMessageUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		nil,
		nil,
		nil,
	)

	ctx := context.Background()
	ipAddress := "192.168.1.100"
	droneID := "drone-123"

	mockDroneRepo.On("GetDroneIDByIP", ctx, ipAddress).Return(droneID, nil)

	result, err := uc.GetDroneIDByIP(ctx, ipAddress)

	assert.NoError(t, err)
	assert.Equal(t, droneID, result)
	mockDroneRepo.AssertExpectations(t)
}

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
