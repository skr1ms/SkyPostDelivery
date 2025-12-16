package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrderUseCase_processPendingOrders_Success(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockGoodRepo := new(mocks.MockGoodRepo)
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockRabbitMQClient := new(mocks.MockRabbitMQClient)
	mockLogger := new(mocks.MockLogger)

	uc := NewOrderUseCase(
		mockOrderRepo,
		mockGoodRepo,
		mockDroneRepo,
		mockDeliveryRepo,
		mockParcelAutomatRepo,
		mockLockerRepo,
		mockInternalLockerRepo,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx := context.Background()
	deliveryID := uuid.New()
	orderID := uuid.New()
	droneID := uuid.New()
	goodID := uuid.New()
	parcelAutomatID := uuid.New()

	delivery := &entity.Delivery{
		ID:              deliveryID,
		OrderID:         orderID,
		ParcelAutomatID: parcelAutomatID,
		Status:          "awaiting_drone",
	}

	order := &entity.Order{
		ID:              orderID,
		GoodID:          goodID,
		ParcelAutomatID: parcelAutomatID,
		Status:          "pending",
	}

	good := &entity.Good{
		ID:     goodID,
		Name:   "Test Good",
		Weight: 1.5,
		Height: 10,
		Length: 20,
		Width:  15,
	}

	parcelAutomat := &entity.ParcelAutomat{
		ID:      parcelAutomatID,
		City:    "Test City",
		Address: "Test Address",
	}

	drone := &entity.Drone{
		ID:     droneID,
		Model:  "DJI Mavic Pro",
		Status: "idle",
	}

	mockDeliveryRepo.On("ListByStatus", ctx, "awaiting_drone").Return([]*entity.Delivery{delivery}, nil)
	mockDroneRepo.On("GetAvailable", ctx).Return(drone, nil)
	mockDroneRepo.On("UpdateStatus", ctx, mock.MatchedBy(func(d *entity.Drone) bool {
		return d.ID == droneID && d.Status == "busy"
	})).Return(nil)
	mockDeliveryRepo.On("UpdateDrone", ctx, mock.MatchedBy(func(d *entity.Delivery) bool {
		return d.ID == deliveryID && d.DroneID != nil && *d.DroneID == droneID
	})).Return(nil)
	mockDeliveryRepo.On("UpdateStatus", ctx, mock.MatchedBy(func(d *entity.Delivery) bool {
		return d.ID == deliveryID && d.Status == "pending"
	})).Return(delivery, nil)
	mockOrderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	mockGoodRepo.On("GetByID", ctx, goodID).Return(good, nil)
	mockParcelAutomatRepo.On("GetByID", ctx, parcelAutomatID).Return(parcelAutomat, nil)
	mockRabbitMQClient.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	uc.processPendingOrders(ctx)

	mockDeliveryRepo.AssertExpectations(t)
	mockDroneRepo.AssertExpectations(t)
	mockRabbitMQClient.AssertExpectations(t)
}

func TestOrderUseCase_processPendingOrders_NoAvailableDrone(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockGoodRepo := new(mocks.MockGoodRepo)
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockRabbitMQClient := new(mocks.MockRabbitMQClient)
	mockLogger := new(mocks.MockLogger)
	uc := NewOrderUseCase(
		mockOrderRepo,
		mockGoodRepo,
		mockDroneRepo,
		mockDeliveryRepo,
		mockParcelAutomatRepo,
		mockLockerRepo,
		mockInternalLockerRepo,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx := context.Background()
	deliveryID := uuid.New()
	orderID := uuid.New()
	parcelAutomatID := uuid.New()

	delivery := &entity.Delivery{
		ID:              deliveryID,
		OrderID:         orderID,
		ParcelAutomatID: parcelAutomatID,
		Status:          "awaiting_drone",
	}

	mockDeliveryRepo.On("ListByStatus", ctx, "awaiting_drone").Return([]*entity.Delivery{delivery}, nil)
	mockDroneRepo.On("GetAvailable", ctx).Return(nil, assert.AnError)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	uc.processPendingOrders(ctx)

	mockDeliveryRepo.AssertExpectations(t)
	mockDroneRepo.AssertExpectations(t)
}

func TestOrderUseCase_processPendingOrders_NoDeliveries(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockGoodRepo := new(mocks.MockGoodRepo)
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockRabbitMQClient := new(mocks.MockRabbitMQClient)

	uc := NewOrderUseCase(
		mockOrderRepo,
		mockGoodRepo,
		mockDroneRepo,
		mockDeliveryRepo,
		mockParcelAutomatRepo,
		mockLockerRepo,
		mockInternalLockerRepo,
		mockRabbitMQClient,
		nil,
	)

	ctx := context.Background()

	mockDeliveryRepo.On("ListByStatus", ctx, "awaiting_drone").Return([]*entity.Delivery{}, nil)

	uc.processPendingOrders(ctx)

	mockDeliveryRepo.AssertExpectations(t)
}

func TestOrderUseCase_StartPendingOrdersWorker_Cancellation(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockGoodRepo := new(mocks.MockGoodRepo)
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockRabbitMQClient := new(mocks.MockRabbitMQClient)
	mockLogger := new(mocks.MockLogger)

	uc := NewOrderUseCase(
		mockOrderRepo,
		mockGoodRepo,
		mockDroneRepo,
		mockDeliveryRepo,
		mockParcelAutomatRepo,
		mockLockerRepo,
		mockInternalLockerRepo,
		mockRabbitMQClient,
		mockLogger,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	mockDeliveryRepo.On("ListByStatus", mock.Anything, "awaiting_drone").Return([]*entity.Delivery{}, nil).Maybe()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

	done := make(chan struct{})
	go func() {
		uc.StartPendingOrdersWorker(ctx, 50*time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Worker did not stop after context cancellation")
	}
}
