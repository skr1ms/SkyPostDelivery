package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
)

func TestOrderUseCase_CreateOrder_GoodNotFound(t *testing.T) {
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
	userID := uuid.New()
	goodID := uuid.New()

	mockGoodRepo.On("GetByID", ctx, goodID).Return(nil, errors.New("good not found"))

	result, err := uc.CreateOrder(ctx, userID, goodID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "good not found")
	mockGoodRepo.AssertExpectations(t)
}

func TestOrderUseCase_GetOrder_Success(t *testing.T) {
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockGoodRepo := new(mocks.MockGoodRepo)
	mockDroneRepo := new(mocks.MockDroneRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockRabbitMQClient := new(mocks.MockRabbitMQClient)

	uc := NewOrderUseCase(
		mockOrderRepo,
		mockGoodRepo,
		mockDroneRepo,
		mockDeliveryRepo,
		mockParcelAutomatRepo,
		mockLockerRepo,
		nil,
		mockRabbitMQClient,
		nil,
	)

	ctx := context.Background()
	orderID := uuid.New()

	order := &entity.Order{
		ID:     orderID,
		Status: "pending",
	}

	mockOrderRepo.On("GetByID", ctx, orderID).Return(order, nil)

	result, err := uc.GetOrder(ctx, orderID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, orderID, result.ID)
	mockOrderRepo.AssertExpectations(t)
}

func TestOrderUseCase_GetUserOrders_Success(t *testing.T) {
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
	userID := uuid.New()

	orders := []*entity.Order{
		{ID: uuid.New(), UserID: userID, Status: "pending"},
		{ID: uuid.New(), UserID: userID, Status: "delivered"},
	}

	mockOrderRepo.On("ListByUserID", ctx, userID).Return(orders, nil)

	result, err := uc.GetUserOrders(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	mockOrderRepo.AssertExpectations(t)
}

func TestOrderUseCase_GetUserOrdersWithGoods_Success(t *testing.T) {
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
	userID := uuid.New()
	goodID1 := uuid.New()
	goodID2 := uuid.New()

	ordersWithGoods := []struct {
		Order *entity.Order
		Good  *entity.Good
	}{
		{
			Order: &entity.Order{
				ID:     uuid.New(),
				UserID: userID,
				GoodID: goodID1,
				Status: "pending",
			},
			Good: &entity.Good{
				ID:                goodID1,
				Name:              "Яблоко",
				Weight:            0.5,
				Height:            10.0,
				Length:            10.0,
				Width:             10.0,
				QuantityAvailable: 50,
			},
		},
		{
			Order: &entity.Order{
				ID:     uuid.New(),
				UserID: userID,
				GoodID: goodID2,
				Status: "delivered",
			},
			Good: &entity.Good{
				ID:                goodID2,
				Name:              "Банан",
				Weight:            0.3,
				Height:            15.0,
				Length:            5.0,
				Width:             5.0,
				QuantityAvailable: 30,
			},
		},
	}

	mockOrderRepo.On("ListByUserIDWithGoods", ctx, userID).Return(ordersWithGoods, nil)

	result, err := uc.GetUserOrdersWithGoods(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "Яблоко", result[0].Good.Name)
	assert.Equal(t, "Банан", result[1].Good.Name)
	assert.Equal(t, "pending", result[0].Order.Status)
	assert.Equal(t, "delivered", result[1].Order.Status)
	mockOrderRepo.AssertExpectations(t)
}

func TestOrderUseCase_GetUserOrdersWithGoods_Error(t *testing.T) {
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
	userID := uuid.New()

	mockOrderRepo.On("ListByUserIDWithGoods", ctx, userID).Return(nil, errors.New("database error"))

	result, err := uc.GetUserOrdersWithGoods(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	mockOrderRepo.AssertExpectations(t)
}

func TestOrderUseCase_CreateMultipleOrders_Success(t *testing.T) {
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
	userID := uuid.New()
	goodID1 := uuid.New()
	goodID2 := uuid.New()

	mockGoodRepo.On("GetByID", ctx, goodID1).Return(nil, errors.New("good not found"))
	mockGoodRepo.On("GetByID", ctx, goodID2).Return(nil, errors.New("good not found"))

	result, err := uc.CreateMultipleOrders(ctx, userID, []uuid.UUID{goodID1, goodID2})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "good not found")
}
