package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/drone-service/internal/usecase/mocks"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeliveryUseCase_StartDelivery_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)
	mockNotifier := new(mocks.MockDroneNotifier)
	mockGRPCClient := new(mocks.MockOrchestratorGRPCClient)
	mockRabbitMQClient := mocks.NewMockRabbitMQClient(t)

	uc := NewDeliveryUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockNotifier,
		mockGRPCClient,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx := context.Background()
	droneID := "drone-123"
	orderID := "order-456"
	goodID := "good-789"
	lockerCellID := "cell-123"
	parcelAutomatID := "automat-456"
	arucoID := 131
	dimensions := entity.GoodDimensions{
		Weight: 1.5,
		Height: 10.0,
		Length: 20.0,
		Width:  15.0,
	}

	_ = mockDroneManager.RegisterDrone(ctx, droneID)

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDeliveryRepo.On("SaveDeliveryTask", ctx, mock.MatchedBy(func(task *entity.DeliveryTask) bool {
		return task.OrderID == orderID &&
			task.GoodID == goodID &&
			task.DroneID != nil &&
			*task.DroneID == droneID
	})).Return(nil)

	state := &entity.DroneState{
		DroneID:           droneID,
		Status:            entity.DroneStatusIdle,
		BatteryLevel:      85.0,
		CurrentDeliveryID: nil,
	}

	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(state, nil)
	mockDroneRepo.On("SaveDroneState", ctx, mock.Anything).Return(nil)

	mockDeliveryRepo.On("UpdateDeliveryStatus", ctx, orderID, entity.DeliveryStatusInProgress, (*string)(nil)).Return(nil).Maybe()
	mockNotifier.On("SendToDrone", ctx, droneID, mock.Anything).Return(nil).Maybe()

	result, err := uc.StartDelivery(
		ctx,
		droneID,
		orderID,
		goodID,
		lockerCellID,
		parcelAutomatID,
		arucoID,
		dimensions,
		nil,
	)

	assert.NoError(t, err)
	assert.True(t, result["success"].(bool))
	assert.NotEmpty(t, result["delivery_id"])

	time.Sleep(50 * time.Millisecond)

	mockDeliveryRepo.AssertExpectations(t)
	mockDroneRepo.AssertExpectations(t)
}

func TestDeliveryUseCase_StartDelivery_SaveTaskError(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)
	mockNotifier := new(mocks.MockDroneNotifier)
	mockGRPCClient := new(mocks.MockOrchestratorGRPCClient)
	mockRabbitMQClient := mocks.NewMockRabbitMQClient(t)

	uc := NewDeliveryUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockNotifier,
		mockGRPCClient,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx := context.Background()
	dimensions := entity.GoodDimensions{Weight: 1.0, Height: 10.0, Length: 20.0, Width: 15.0}

	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockDeliveryRepo.On("SaveDeliveryTask", ctx, mock.Anything).Return(errors.New("database error"))

	result, err := uc.StartDelivery(
		ctx,
		"drone-123",
		"order-456",
		"good-789",
		"cell-123",
		"automat-456",
		131,
		dimensions,
		nil,
	)

	assert.Error(t, err)
	assert.False(t, result["success"].(bool))
	mockDeliveryRepo.AssertExpectations(t)
}

func TestDeliveryUseCase_ExecuteDelivery_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)
	mockNotifier := new(mocks.MockDroneNotifier)
	mockGRPCClient := new(mocks.MockOrchestratorGRPCClient)
	mockRabbitMQClient := mocks.NewMockRabbitMQClient(t)

	uc := NewDeliveryUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockNotifier,
		mockGRPCClient,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx := context.Background()
	droneID := "drone-123"
	internalCellID := "internal-cell-123"

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockNotifier.On("SendToDrone", ctx, droneID, mock.MatchedBy(func(msg map[string]any) bool {
		return msg["type"] == "delivery_task" &&
			msg["payload"] != nil
	})).Return(nil)

	err := uc.ExecuteDelivery(
		ctx,
		droneID,
		"order-456",
		"good-789",
		"automat-456",
		131,
		"55.7558,37.6173",
		1.5,
		10.0,
		20.0,
		15.0,
		&internalCellID,
	)

	assert.NoError(t, err)
	mockNotifier.AssertExpectations(t)
}

func TestDeliveryUseCase_ExecuteDelivery_NotifierError(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)
	mockNotifier := new(mocks.MockDroneNotifier)
	mockGRPCClient := new(mocks.MockOrchestratorGRPCClient)
	mockRabbitMQClient := mocks.NewMockRabbitMQClient(t)

	uc := NewDeliveryUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockNotifier,
		mockGRPCClient,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx := context.Background()
	droneID := "drone-123"

	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockNotifier.On("SendToDrone", ctx, droneID, mock.Anything).Return(errors.New("connection error"))

	err := uc.ExecuteDelivery(
		ctx,
		droneID,
		"order-456",
		"good-789",
		"automat-456",
		131,
		"55.7558,37.6173",
		1.5,
		10.0,
		20.0,
		15.0,
		nil,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SendToDrone")
	mockNotifier.AssertExpectations(t)
}

func TestDeliveryUseCase_HandleReturnTask_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)
	mockNotifier := new(mocks.MockDroneNotifier)
	mockGRPCClient := new(mocks.MockOrchestratorGRPCClient)
	mockRabbitMQClient := mocks.NewMockRabbitMQClient(t)

	uc := NewDeliveryUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockNotifier,
		mockGRPCClient,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx := context.Background()
	droneID := "drone-123"
	deliveryID := "delivery-456"
	baseMarkerID := 131

	state := &entity.DroneState{
		DroneID:           droneID,
		Status:            entity.DroneStatusDelivering,
		BatteryLevel:      85.0,
		CurrentDeliveryID: stringPtr(deliveryID),
	}

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockNotifier.On("SendToDrone", ctx, droneID, mock.MatchedBy(func(msg map[string]any) bool {
		return msg["type"] == "command"
	})).Return(nil)

	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(state, nil)
	mockDroneRepo.On("SaveDroneState", ctx, mock.MatchedBy(func(s *entity.DroneState) bool {
		return s.Status == entity.DroneStatusIdle && s.CurrentDeliveryID == nil
	})).Return(nil)

	err := uc.HandleReturnTask(ctx, droneID, deliveryID, baseMarkerID)

	assert.NoError(t, err)
	mockNotifier.AssertExpectations(t)
	mockDroneRepo.AssertExpectations(t)
}

func TestDeliveryUseCase_HandleDroneArrived_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)
	mockNotifier := new(mocks.MockDroneNotifier)
	mockGRPCClient := new(mocks.MockOrchestratorGRPCClient)
	mockRabbitMQClient := mocks.NewMockRabbitMQClient(t)

	uc := NewDeliveryUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockNotifier,
		mockGRPCClient,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx := context.Background()
	droneID := "drone-123"
	orderID := "order-456"
	parcelAutomatID := "automat-456"

	response := &grpc.CellOpenResponse{
		Success:        true,
		Message:        "Cell opened",
		CellID:         "cell-123",
		InternalCellID: "internal-cell-123",
	}

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockGRPCClient.On("RequestCellOpen", ctx, orderID, parcelAutomatID).Return(response, nil)
	mockNotifier.On("SendToDrone", ctx, droneID, mock.MatchedBy(func(msg map[string]any) bool {
		payload, ok := msg["payload"].(map[string]any)
		return ok && payload["command"] == "drop_cargo"
	})).Return(nil)

	result, err := uc.HandleDroneArrived(ctx, droneID, orderID, parcelAutomatID)

	assert.NoError(t, err)
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "cell-123", result["cell_id"])
	mockGRPCClient.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}

func TestDeliveryUseCase_HandleDroneArrived_GRPCError(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)
	mockNotifier := new(mocks.MockDroneNotifier)
	mockGRPCClient := new(mocks.MockOrchestratorGRPCClient)
	mockRabbitMQClient := mocks.NewMockRabbitMQClient(t)

	uc := NewDeliveryUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockNotifier,
		mockGRPCClient,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx := context.Background()
	droneID := "drone-123"
	orderID := "order-456"
	parcelAutomatID := "automat-456"

	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockGRPCClient.On("RequestCellOpen", ctx, orderID, parcelAutomatID).Return(nil, errors.New("grpc error"))

	result, err := uc.HandleDroneArrived(ctx, droneID, orderID, parcelAutomatID)

	assert.Error(t, err)
	assert.False(t, result["success"].(bool))
	mockGRPCClient.AssertExpectations(t)
}

func TestDeliveryUseCase_HandleCargoDropped_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)
	mockNotifier := new(mocks.MockDroneNotifier)
	mockGRPCClient := new(mocks.MockOrchestratorGRPCClient)
	mockRabbitMQClient := mocks.NewMockRabbitMQClient(t)

	uc := NewDeliveryUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockNotifier,
		mockGRPCClient,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx := context.Background()
	orderID := "order-456"
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

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDeliveryRepo.On("GetDeliveryTask", ctx, orderID).Return(task, nil)
	mockDeliveryRepo.On("UpdateDeliveryStatus", ctx, orderID, entity.DeliveryStatusCompleted, (*string)(nil)).Return(nil)
	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(state, nil)
	mockDroneRepo.On("SaveDroneState", ctx, mock.Anything).Return(nil)
	mockRabbitMQClient.On("Publish", ctx, "confirmations", mock.Anything).Return(nil)

	result, err := uc.HandleCargoDropped(ctx, orderID, lockerCellID)

	assert.NoError(t, err)
	assert.True(t, result["success"].(bool))
	mockDeliveryRepo.AssertExpectations(t)
	mockDroneRepo.AssertExpectations(t)
	mockRabbitMQClient.AssertExpectations(t)
}

func TestDeliveryUseCase_CancelDelivery_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)
	mockNotifier := new(mocks.MockDroneNotifier)
	mockGRPCClient := new(mocks.MockOrchestratorGRPCClient)
	mockRabbitMQClient := mocks.NewMockRabbitMQClient(t)

	uc := NewDeliveryUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockNotifier,
		mockGRPCClient,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx := context.Background()
	deliveryID := "delivery-456"
	droneID := "drone-123"

	task := &entity.DeliveryTask{
		DeliveryID: deliveryID,
		OrderID:    "order-456",
		DroneID:    &droneID,
		Status:     entity.DeliveryStatusInProgress,
	}

	state := &entity.DroneState{
		DroneID:           droneID,
		Status:            entity.DroneStatusDelivering,
		BatteryLevel:      85.0,
		CurrentDeliveryID: stringPtr(deliveryID),
	}

	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDeliveryRepo.On("GetDeliveryTask", ctx, deliveryID).Return(task, nil)
	mockNotifier.On("SendToDrone", ctx, droneID, mock.Anything).Return(nil)
	mockDeliveryRepo.On("UpdateDeliveryStatus", ctx, deliveryID, entity.DeliveryStatusCancelled, (*string)(nil)).Return(nil)
	mockDroneRepo.On("GetDroneState", ctx, droneID).Return(state, nil)
	mockDroneRepo.On("SaveDroneState", ctx, mock.Anything).Return(nil)

	result, err := uc.CancelDelivery(ctx, deliveryID)

	assert.NoError(t, err)
	assert.True(t, result["success"].(bool))
	mockDeliveryRepo.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
	mockDroneRepo.AssertExpectations(t)
}

func TestDeliveryUseCase_GetDeliveryStatus_Success(t *testing.T) {
	mockDroneRepo := mocks.NewMockDroneRepo(t)
	mockDeliveryRepo := mocks.NewMockDeliveryRepo(t)
	mockLogger := mocks.NewMockLogger(t)
	mockDroneManager := NewDroneManagerUseCase(mockDroneRepo, mockLogger)
	mockNotifier := new(mocks.MockDroneNotifier)
	mockGRPCClient := new(mocks.MockOrchestratorGRPCClient)
	mockRabbitMQClient := mocks.NewMockRabbitMQClient(t)

	uc := NewDeliveryUseCase(
		mockDroneRepo,
		mockDeliveryRepo,
		mockDroneManager,
		mockNotifier,
		mockGRPCClient,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx := context.Background()
	deliveryID := "delivery-456"
	droneID := "drone-123"

	task := &entity.DeliveryTask{
		DeliveryID: deliveryID,
		OrderID:    "order-456",
		DroneID:    &droneID,
		Status:     entity.DeliveryStatusInProgress,
	}

	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	mockDeliveryRepo.On("GetDeliveryTask", ctx, deliveryID).Return(task, nil)

	result, err := uc.GetDeliveryStatus(ctx, deliveryID)

	assert.NoError(t, err)
	assert.True(t, result["success"].(bool))
	assert.Equal(t, deliveryID, result["delivery_id"])
	assert.Equal(t, string(entity.DeliveryStatusInProgress), result["status"])
	assert.Equal(t, droneID, result["drone_id"])
	mockDeliveryRepo.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}
