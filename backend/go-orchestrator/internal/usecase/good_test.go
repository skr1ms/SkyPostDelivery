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

func TestGoodUseCase_List_Success(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

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

	goods, err := uc.List(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedGoods, goods)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_List_Error(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

	expectedError := errors.New("database error")
	mockRepo.On("List", mock.Anything).Return(nil, expectedError)

	goods, err := uc.List(context.Background())

	assert.Error(t, err)
	assert.Nil(t, goods)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_GetByID_Success(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

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

	good, err := uc.GetByID(context.Background(), goodID)

	assert.NoError(t, err)
	assert.Equal(t, expectedGood, good)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_GetByID_NotFound(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

	goodID := uuid.New()
	expectedError := errors.New("good not found")

	mockRepo.On("GetByID", mock.Anything, goodID).Return(nil, expectedError)

	good, err := uc.GetByID(context.Background(), goodID)

	assert.Error(t, err)
	assert.Nil(t, good)
	assert.Contains(t, err.Error(), "good not found")
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_Create_Success(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

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

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(g *entity.Good) bool {
		return g.Name == "iPhone 15" && g.Weight == 0.5 && g.Height == 15.0 && g.Length == 8.0 && g.Width == 1.0 && g.QuantityAvailable == 1
	})).Return(expectedGood, nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	good, err := uc.Create(context.Background(), "iPhone 15", 0.5, 15.0, 8.0, 1.0)

	assert.NoError(t, err)
	assert.NotNil(t, good)
	assert.Equal(t, expectedGood.Name, good.Name)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_Create_Error(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

	expectedError := errors.New("database error")
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(g *entity.Good) bool {
		return g.Name == "iPhone 15" && g.Weight == 0.5 && g.Height == 15.0 && g.Length == 8.0 && g.Width == 1.0 && g.QuantityAvailable == 1
	})).Return(nil, expectedError)

	good, err := uc.Create(context.Background(), "iPhone 15", 0.5, 15.0, 8.0, 1.0)

	assert.Error(t, err)
	assert.Nil(t, good)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_Update_Success(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

	goodID := uuid.New()
	expectedGood := &entity.Good{
		ID:     goodID,
		Name:   "iPhone 16",
		Weight: 0.6,
		Height: 16.0,
		Length: 9.0,
		Width:  1.2,
	}

	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(g *entity.Good) bool {
		return g.ID == goodID && g.Name == "iPhone 16" && g.Weight == 0.6 && g.Height == 16.0 && g.Length == 9.0 && g.Width == 1.2
	})).Return(expectedGood, nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	good, err := uc.Update(context.Background(), goodID, "iPhone 16", 0.6, 16.0, 9.0, 1.2)

	assert.NoError(t, err)
	assert.Equal(t, expectedGood, good)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_Update_Error(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

	goodID := uuid.New()
	expectedError := errors.New("update failed")

	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(g *entity.Good) bool {
		return g.ID == goodID && g.Name == "iPhone 16" && g.Weight == 0.6 && g.Height == 16.0 && g.Length == 9.0 && g.Width == 1.2
	})).Return(nil, expectedError)

	good, err := uc.Update(context.Background(), goodID, "iPhone 16", 0.6, 16.0, 9.0, 1.2)

	assert.Error(t, err)
	assert.Nil(t, good)
	assert.Contains(t, err.Error(), "update failed")
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_Delete_Success(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

	goodID := uuid.New()

	mockRepo.On("Delete", mock.Anything, goodID).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	err := uc.Delete(context.Background(), goodID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_Delete_Error(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

	goodID := uuid.New()
	expectedError := errors.New("delete failed")

	mockRepo.On("Delete", mock.Anything, goodID).Return(expectedError)

	err := uc.Delete(context.Background(), goodID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_CreateWithQuantity_Success(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

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

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(g *entity.Good) bool {
		return g.Name == name && g.Weight == weight && g.Height == height && g.Length == length && g.Width == width && g.QuantityAvailable == quantity
	})).Return(expectedGood, nil).Once()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	good, err := uc.CreateWithQuantity(context.Background(), name, weight, height, length, width, quantity)

	assert.NoError(t, err)
	assert.NotNil(t, good)
	assert.Equal(t, name, good.Name)
	assert.Equal(t, quantity, good.QuantityAvailable)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_CreateWithQuantity_SingleItem(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

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

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(g *entity.Good) bool {
		return g.Name == name && g.Weight == weight && g.Height == height && g.Length == length && g.Width == width && g.QuantityAvailable == quantity
	})).Return(expectedGood, nil).Once()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	good, err := uc.CreateWithQuantity(context.Background(), name, weight, height, length, width, quantity)

	assert.NoError(t, err)
	assert.NotNil(t, good)
	assert.Equal(t, expectedGood.Name, good.Name)
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_CreateWithQuantity_Error(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

	quantity := 3
	name := "Orange"
	weight := 0.3
	height := 6.0
	length := 9.0
	width := 6.0

	expectedError := errors.New("database error")
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(g *entity.Good) bool {
		return g.Name == name && g.Weight == weight && g.Height == height && g.Length == length && g.Width == width && g.QuantityAvailable == quantity
	})).Return(nil, expectedError).Once()

	good, err := uc.CreateWithQuantity(context.Background(), name, weight, height, length, width, quantity)

	assert.Error(t, err)
	assert.Nil(t, good)
	assert.Contains(t, err.Error(), "GoodUseCase - CreateWithQuantity")
	mockRepo.AssertExpectations(t)
}

func TestGoodUseCase_CreateWithQuantity_LargeQuantity(t *testing.T) {
	mockRepo := new(mocks.MockGoodRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewGoodUseCase(mockRepo, mockLogger)

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

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(g *entity.Good) bool {
		return g.Name == name && g.Weight == weight && g.Height == height && g.Length == length && g.Width == width && g.QuantityAvailable == quantity
	})).Return(expectedGood, nil).Once()
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	good, err := uc.CreateWithQuantity(context.Background(), name, weight, height, length, width, quantity)

	assert.NoError(t, err)
	assert.NotNil(t, good)
	assert.Equal(t, quantity, good.QuantityAvailable)
	mockRepo.AssertExpectations(t)
}
