package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestLockerUseCase_GetCell_Success(t *testing.T) {
	mockLockerRepo := new(MockLockerRepo)
	uc := NewLockerUseCase(mockLockerRepo)

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
	mockLockerRepo := new(MockLockerRepo)
	uc := NewLockerUseCase(mockLockerRepo)

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
	mockLockerRepo := new(MockLockerRepo)
	uc := NewLockerUseCase(mockLockerRepo)

	ctx := context.Background()
	cellID := uuid.New()
	status := "occupied"

	mockLockerRepo.On("UpdateCellStatus", ctx, cellID, status).Return(nil)

	err := uc.UpdateCellStatus(ctx, cellID, status)

	assert.NoError(t, err)
	mockLockerRepo.AssertExpectations(t)
}
