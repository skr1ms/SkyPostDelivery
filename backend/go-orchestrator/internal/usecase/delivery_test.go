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

func TestDeliveryUseCase_GetByID_Success(t *testing.T) {
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockRabbitMQClient := new(mocks.MockRabbitMQClient)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewDeliveryUseCase(mockDeliveryRepo, mockOrderRepo, mockLockerRepo, mockInternalLockerRepo, mockRabbitMQClient, nil, mockLogger)

	ctx := context.Background()
	deliveryID := uuid.New()

	delivery := &entity.Delivery{
		ID:     deliveryID,
		Status: "pending",
	}

	mockDeliveryRepo.On("GetByID", ctx, deliveryID).Return(delivery, nil)

	result, err := uc.GetByID(ctx, deliveryID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, deliveryID, result.ID)
	mockDeliveryRepo.AssertExpectations(t)
}

func TestDeliveryUseCase_GetByID_NotFound(t *testing.T) {
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockRabbitMQClient := new(mocks.MockRabbitMQClient)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewDeliveryUseCase(mockDeliveryRepo, mockOrderRepo, mockLockerRepo, mockInternalLockerRepo, mockRabbitMQClient, nil, mockLogger)

	ctx := context.Background()
	deliveryID := uuid.New()

	mockDeliveryRepo.On("GetByID", ctx, deliveryID).Return(nil, errors.New("delivery not found"))

	result, err := uc.GetByID(ctx, deliveryID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "delivery not found")
	mockDeliveryRepo.AssertExpectations(t)
}

func TestDeliveryUseCase_UpdateStatus_Success(t *testing.T) {
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockRabbitMQClient := new(mocks.MockRabbitMQClient)
	mockLogger := new(mocks.MockLogger)
	uc := NewDeliveryUseCase(mockDeliveryRepo, mockOrderRepo, mockLockerRepo, mockInternalLockerRepo, mockRabbitMQClient, nil, mockLogger)

	ctx := context.Background()
	deliveryID := uuid.New()
	orderID := uuid.New()

	delivery := &entity.Delivery{
		ID:      deliveryID,
		OrderID: orderID,
		Status:  "pending",
	}

	updatedDelivery := &entity.Delivery{
		ID:      deliveryID,
		OrderID: orderID,
		Status:  "in_transit",
	}

	order := &entity.Order{
		ID:     orderID,
		Status: "pending",
	}

	updatedOrder := &entity.Order{
		ID:     orderID,
		Status: "in_progress",
	}

	mockDeliveryRepo.On("GetByID", ctx, deliveryID).Return(delivery, nil)
	mockDeliveryRepo.On("UpdateStatus", ctx, mock.MatchedBy(func(d *entity.Delivery) bool {
		return d.ID == deliveryID && d.Status == "in_transit"
	})).Return(updatedDelivery, nil)
	mockOrderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	mockOrderRepo.On("UpdateStatus", ctx, mock.MatchedBy(func(o *entity.Order) bool {
		return o.ID == orderID && o.Status == "in_progress"
	})).Return(updatedOrder, nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	err := uc.UpdateStatus(ctx, deliveryID, "in_transit")

	assert.NoError(t, err)
	mockDeliveryRepo.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
}

func TestDeliveryUseCase_UpdateStatus_Delivered(t *testing.T) {
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockRabbitMQClient := new(mocks.MockRabbitMQClient)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewDeliveryUseCase(mockDeliveryRepo, mockOrderRepo, mockLockerRepo, mockInternalLockerRepo, mockRabbitMQClient, nil, mockLogger)

	ctx := context.Background()
	deliveryID := uuid.New()
	orderID := uuid.New()

	delivery := &entity.Delivery{
		ID:              deliveryID,
		OrderID:         orderID,
		ParcelAutomatID: uuid.New(),
		Status:          "in_transit",
	}

	updatedDelivery := &entity.Delivery{
		ID:              deliveryID,
		OrderID:         orderID,
		ParcelAutomatID: uuid.New(),
		Status:          "delivered",
	}

	order := &entity.Order{
		ID:     orderID,
		Status: "in_progress",
	}

	updatedOrder := &entity.Order{
		ID:     orderID,
		Status: "delivered",
	}

	mockDeliveryRepo.On("GetByID", ctx, deliveryID).Return(delivery, nil)
	mockDeliveryRepo.On("UpdateStatus", ctx, mock.MatchedBy(func(d *entity.Delivery) bool {
		return d.ID == deliveryID && d.Status == "delivered"
	})).Return(updatedDelivery, nil)
	mockOrderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	mockOrderRepo.On("UpdateStatus", ctx, mock.MatchedBy(func(o *entity.Order) bool {
		return o.ID == orderID && o.Status == "delivered"
	})).Return(updatedOrder, nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	err := uc.UpdateStatus(ctx, deliveryID, "delivered")

	assert.NoError(t, err)
	mockDeliveryRepo.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
}

func TestDeliveryUseCase_ListByStatus_Success(t *testing.T) {
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockRabbitMQClient := new(mocks.MockRabbitMQClient)
	mockLogger := new(mocks.MockLogger)
	uc := NewDeliveryUseCase(mockDeliveryRepo, mockOrderRepo, mockLockerRepo, mockInternalLockerRepo, mockRabbitMQClient, nil, mockLogger)

	ctx := context.Background()
	status := "pending"

	deliveries := []*entity.Delivery{
		{ID: uuid.New(), Status: "pending"},
		{ID: uuid.New(), Status: "pending"},
	}

	mockDeliveryRepo.On("ListByStatus", ctx, status).Return(deliveries, nil)

	result, err := uc.ListByStatus(ctx, status)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	mockDeliveryRepo.AssertExpectations(t)
}

func TestDeliveryUseCase_ConfirmGoodsLoaded_Success(t *testing.T) {
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockRabbitMQClient := new(mocks.MockRabbitMQClient)
	mockLogger := new(mocks.MockLogger)
	uc := NewDeliveryUseCase(mockDeliveryRepo, mockOrderRepo, mockLockerRepo, mockInternalLockerRepo, mockRabbitMQClient, nil, mockLogger)

	ctx := context.Background()
	orderID := uuid.New()
	lockerCellID := uuid.New()
	deliveryID := uuid.New()

	internalCellID := uuid.New()
	delivery := &entity.Delivery{
		ID:                   deliveryID,
		OrderID:              orderID,
		ParcelAutomatID:      uuid.New(),
		InternalLockerCellID: &internalCellID,
		Status:               "in_transit",
	}

	updatedDelivery := &entity.Delivery{
		ID:              deliveryID,
		OrderID:         orderID,
		ParcelAutomatID: uuid.New(),
		Status:          "delivered",
	}

	order := &entity.Order{
		ID:           orderID,
		Status:       "delivered",
		LockerCellID: &lockerCellID,
	}

	mockDeliveryRepo.On("GetByOrderID", ctx, orderID).Return(delivery, nil)
	mockDeliveryRepo.On("UpdateStatus", ctx, mock.MatchedBy(func(d *entity.Delivery) bool {
		return d.ID == deliveryID && d.Status == "delivered"
	})).Return(updatedDelivery, nil)
	mockOrderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	mockOrderRepo.On("UpdateStatus", ctx, mock.MatchedBy(func(o *entity.Order) bool {
		return o.ID == orderID && o.Status == "delivered"
	})).Return(order, nil)
	lockerCell := &entity.LockerCell{
		ID:     lockerCellID,
		Status: "available",
	}
	mockLockerRepo.On("GetCellByID", ctx, lockerCellID).Return(lockerCell, nil)
	mockLockerRepo.On("UpdateCellStatus", ctx, mock.MatchedBy(func(c *entity.LockerCell) bool {
		return c.ID == lockerCellID && c.Status == "occupied"
	})).Return(nil)
	internalCell := &entity.LockerCell{
		ID:     internalCellID,
		Status: "available",
	}
	mockInternalLockerRepo.On("GetCellByID", ctx, internalCellID).Return(internalCell, nil)
	mockInternalLockerRepo.On("UpdateCellStatus", ctx, mock.MatchedBy(func(c *entity.LockerCell) bool {
		return c.ID == internalCellID && c.Status == "occupied"
	})).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Maybe().Return()

	err := uc.ConfirmGoodsLoaded(ctx, orderID, lockerCellID)

	assert.NoError(t, err)
	mockDeliveryRepo.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
	mockLockerRepo.AssertExpectations(t)
}

func TestDeliveryUseCase_ConfirmGoodsLoaded_DeliveryNotFound(t *testing.T) {
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockRabbitMQClient := new(mocks.MockRabbitMQClient)
	mockLogger := new(mocks.MockLogger)
	uc := NewDeliveryUseCase(mockDeliveryRepo, mockOrderRepo, mockLockerRepo, mockInternalLockerRepo, mockRabbitMQClient, nil, mockLogger)

	ctx := context.Background()
	orderID := uuid.New()
	lockerCellID := uuid.New()

	mockDeliveryRepo.On("GetByOrderID", ctx, orderID).Return(nil, errors.New("delivery not found"))

	err := uc.ConfirmGoodsLoaded(ctx, orderID, lockerCellID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delivery not found")
	mockDeliveryRepo.AssertExpectations(t)
}
