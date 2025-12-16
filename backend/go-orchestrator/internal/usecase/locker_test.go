package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLockerUseCase_GetCell_Success(t *testing.T) {
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewLockerUseCase(mockLockerRepo, mockLogger)

	ctx := context.Background()
	cellID := uuid.New()

	cell := &entity.LockerCell{
		ID:     cellID,
		PostID: uuid.New(),
		Status: "available",
		Height: 10.0,
		Length: 10.0,
		Width:  10.0,
	}

	mockLockerRepo.On("GetCellByID", ctx, cellID).Return(cell, nil)

	result, err := uc.GetCell(ctx, cellID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, cellID, result.ID)
	mockLockerRepo.AssertExpectations(t)
}

func TestLockerUseCase_ListCells_Success(t *testing.T) {
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewLockerUseCase(mockLockerRepo, mockLogger)

	ctx := context.Background()
	postID := uuid.New()

	cells := []*entity.LockerCell{
		{ID: uuid.New(), PostID: postID, Status: "available"},
		{ID: uuid.New(), PostID: postID, Status: "occupied"},
	}

	mockLockerRepo.On("ListCellsByPostID", ctx, postID).Return(cells, nil)

	result, err := uc.ListCells(ctx, postID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	mockLockerRepo.AssertExpectations(t)
}

func TestLockerUseCase_UpdateCellStatus_Success(t *testing.T) {
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockLogger := new(mocks.MockLogger)
	uc := NewLockerUseCase(mockLockerRepo, mockLogger)

	ctx := context.Background()
	cellID := uuid.New()
	status := "occupied"

	cell := &entity.LockerCell{ID: cellID, Status: "available"}
	mockLockerRepo.On("GetCellByID", ctx, cellID).Return(cell, nil)
	mockLockerRepo.On("UpdateCellStatus", ctx, mock.MatchedBy(func(c *entity.LockerCell) bool {
		return c.ID == cellID && c.Status == status
	})).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return()

	err := uc.UpdateCellStatus(ctx, cellID, status)

	assert.NoError(t, err)
	mockLockerRepo.AssertExpectations(t)
}
