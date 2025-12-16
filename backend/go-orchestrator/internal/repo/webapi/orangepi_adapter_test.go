package webapi

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOrangePIAdapter_SendCellUUIDs_EmptyIP(t *testing.T) {
	adapter := NewOrangePIAdapter()
	ctx := context.Background()

	parcelAutomatID := uuid.New()
	cellUUIDs := []uuid.UUID{uuid.New()}
	internalCellUUIDs := []uuid.UUID{uuid.New()}

	err := adapter.SendCellUUIDs(ctx, "", parcelAutomatID, cellUUIDs, internalCellUUIDs)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "IP address is empty")
}

func TestOrangePIAdapter_SendCellUUIDs_RequestFailed(t *testing.T) {
	adapter := NewOrangePIAdapter()
	ctx := context.Background()

	parcelAutomatID := uuid.New()
	cellUUIDs := []uuid.UUID{uuid.New()}
	internalCellUUIDs := []uuid.UUID{uuid.New()}

	err := adapter.SendCellUUIDs(ctx, "invalid-host:9999", parcelAutomatID, cellUUIDs, internalCellUUIDs)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OrangePIAdapter - SendCellUUIDs - Do")
}

func TestOrangePIAdapter_OpenCell_EmptyIP(t *testing.T) {
	adapter := NewOrangePIAdapter()
	ctx := context.Background()
	cellID := uuid.New()

	err := adapter.OpenCell(ctx, "", cellID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "IP address is empty")
}

func TestOrangePIAdapter_OpenCell_RequestFailed(t *testing.T) {
	adapter := NewOrangePIAdapter()
	ctx := context.Background()
	cellID := uuid.New()

	err := adapter.OpenCell(ctx, "invalid-host:9999", cellID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OrangePIAdapter - OpenCell - Do")
}

func TestNewOrangePIAdapter(t *testing.T) {
	adapter := NewOrangePIAdapter()

	assert.NotNil(t, adapter)
	assert.NotNil(t, adapter.httpClient)
	assert.Equal(t, 10000000000, int(adapter.httpClient.Timeout))
}
