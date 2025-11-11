package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/skr1ms/hitech-ekb/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGoodUseCase_ListGoods_Success(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	expectedGoods := []*entity.Good{
		{
			ID:     uuid.New(),
			Name:   "iPhone 15",
			Weight: 0.5,
			Height: 15.0,
			Length: 8.0,
			Width:  1.0,
		},
		{
			ID:     uuid.New(),
			Name:   "MacBook Pro",
			Weight: 2.0,
			Height: 35.0,
			Length: 25.0,
			Width:  2.0,
		},
	}

	mockRepo.On("List", mock.Anything).Return(expectedGoods, nil)

	goods, err := uc.ListGoods(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedGoods, goods)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_ListGoods_Error(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	expectedError := errors.New("database error")
	mockRepo.On("List", mock.Anything).Return(nil, expectedError)

	goods, err := uc.ListGoods(context.Background())

	assert.Error(t, err)
	assert.Nil(t, goods)
	assert.Contains(t, err.Error(), "goodRepo.List")
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_GetGood_Success(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	goodID := uuid.New()
	expectedGood := &entity.Good{
		ID:     goodID,
		Name:   "iPhone 15",
		Weight: 0.5,
		Height: 15.0,
		Length: 8.0,
		Width:  1.0,
	}

	mockRepo.On("GetByID", mock.Anything, goodID).Return(expectedGood, nil)

	good, err := uc.GetGood(context.Background(), goodID)

	assert.NoError(t, err)
	assert.Equal(t, expectedGood, good)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_GetGood_NotFound(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	goodID := uuid.New()
	expectedError := errors.New("good not found")

	mockRepo.On("GetByID", mock.Anything, goodID).Return(nil, expectedError)

	good, err := uc.GetGood(context.Background(), goodID)

	assert.Error(t, err)
	assert.Nil(t, good)
	assert.Contains(t, err.Error(), "goodRepo.GetByID")
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_CreateGood_Success(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	goodID := uuid.New()

	expectedGood := &entity.Good{
		ID:                goodID,
		Name:              "iPhone 15",
		Weight:            0.5,
		Height:            15.0,
		Length:            8.0,
		Width:             1.0,
		QuantityAvailable: 1,
	}

	mockRepo.On("Create", mock.Anything, "iPhone 15", 0.5, 15.0, 8.0, 1.0, 1).Return(expectedGood, nil)

	good, err := uc.CreateGood(context.Background(), "iPhone 15", 0.5, 15.0, 8.0, 1.0)

	assert.NoError(t, err)
	assert.NotNil(t, good)
	assert.Equal(t, expectedGood.Name, good.Name)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_CreateGood_Error(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	expectedError := errors.New("database error")
	mockRepo.On("Create", mock.Anything, "iPhone 15", 0.5, 15.0, 8.0, 1.0, 1).Return(nil, expectedError)

	good, err := uc.CreateGood(context.Background(), "iPhone 15", 0.5, 15.0, 8.0, 1.0)

	assert.Error(t, err)
	assert.Nil(t, good)
	assert.Contains(t, err.Error(), "GoodUseCase - CreateGood")
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_UpdateGood_Success(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	goodID := uuid.New()
	expectedGood := &entity.Good{
		ID:     goodID,
		Name:   "iPhone 16",
		Weight: 0.6,
		Height: 16.0,
		Length: 9.0,
		Width:  1.2,
	}

	mockRepo.On("Update", mock.Anything, goodID, "iPhone 16", 0.6, 16.0, 9.0, 1.2).Return(expectedGood, nil)

	good, err := uc.UpdateGood(context.Background(), goodID, "iPhone 16", 0.6, 16.0, 9.0, 1.2)

	assert.NoError(t, err)
	assert.Equal(t, expectedGood, good)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_UpdateGood_Error(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	goodID := uuid.New()
	expectedError := errors.New("update failed")

	mockRepo.On("Update", mock.Anything, goodID, "iPhone 16", 0.6, 16.0, 9.0, 1.2).Return(nil, expectedError)

	good, err := uc.UpdateGood(context.Background(), goodID, "iPhone 16", 0.6, 16.0, 9.0, 1.2)

	assert.Error(t, err)
	assert.Nil(t, good)
	assert.Contains(t, err.Error(), "goodRepo.Update")
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_DeleteGood_Success(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	goodID := uuid.New()

	mockRepo.On("Delete", mock.Anything, goodID).Return(nil)

	err := uc.DeleteGood(context.Background(), goodID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_DeleteGood_Error(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	goodID := uuid.New()
	expectedError := errors.New("delete failed")

	mockRepo.On("Delete", mock.Anything, goodID).Return(expectedError)

	err := uc.DeleteGood(context.Background(), goodID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "goodRepo.Delete")
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_CreateGoods_Success(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	quantity := 5
	name := "Banana"
	weight := 0.15
	height := 7.0
	length := 10.0
	width := 3.0

	goodID := uuid.New()
	expectedGood := &entity.Good{
		ID:                goodID,
		Name:              name,
		Weight:            weight,
		Height:            height,
		Length:            length,
		Width:             width,
		QuantityAvailable: quantity,
	}

	mockRepo.On("Create", mock.Anything, name, weight, height, length, width, quantity).Return(expectedGood, nil).Once()

	goods, err := uc.CreateGoods(context.Background(), name, weight, height, length, width, quantity)

	assert.NoError(t, err)
	assert.NotNil(t, goods)
	assert.Len(t, goods, 1)
	assert.Equal(t, name, goods[0].Name)
	assert.Equal(t, quantity, goods[0].QuantityAvailable)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_CreateGoods_SingleItem(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	quantity := 1
	name := "Apple"
	weight := 0.2
	height := 5.0
	length := 8.0
	width := 5.0

	goodID := uuid.New()
	expectedGood := &entity.Good{
		ID:                goodID,
		Name:              name,
		Weight:            weight,
		Height:            height,
		Length:            length,
		Width:             width,
		QuantityAvailable: 1,
	}

	mockRepo.On("Create", mock.Anything, name, weight, height, length, width, quantity).Return(expectedGood, nil).Once()

	goods, err := uc.CreateGoods(context.Background(), name, weight, height, length, width, quantity)

	assert.NoError(t, err)
	assert.NotNil(t, goods)
	assert.Len(t, goods, 1)
	assert.Equal(t, expectedGood.Name, goods[0].Name)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_CreateGoods_Error(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	quantity := 3
	name := "Orange"
	weight := 0.3
	height := 6.0
	length := 9.0
	width := 6.0

	expectedError := errors.New("database error")
	mockRepo.On("Create", mock.Anything, name, weight, height, length, width, quantity).Return(nil, expectedError).Once()

	goods, err := uc.CreateGoods(context.Background(), name, weight, height, length, width, quantity)

	assert.Error(t, err)
	assert.Nil(t, goods)
	assert.Contains(t, err.Error(), "GoodUseCase - CreateGoods")
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_CreateGoods_LargeQuantity(t *testing.T) {
	mockRepo := new(MockGoodRepo)
	uc := NewGoodUseCase(mockRepo)

	quantity := 100
	name := "Box"
	weight := 1.0
	height := 30.0
	length := 40.0
	width := 30.0

	goodID := uuid.New()
	expectedGood := &entity.Good{
		ID:                goodID,
		Name:              name,
		Weight:            weight,
		Height:            height,
		Length:            length,
		Width:             width,
		QuantityAvailable: quantity,
	}

	mockRepo.On("Create", mock.Anything, name, weight, height, length, width, quantity).Return(expectedGood, nil).Once()

	goods, err := uc.CreateGoods(context.Background(), name, weight, height, length, width, quantity)

	assert.NoError(t, err)
	assert.NotNil(t, goods)
	assert.Len(t, goods, 1)
	assert.Equal(t, quantity, goods[0].QuantityAvailable)
	mockRepo.AssertExpectations(t)
}
