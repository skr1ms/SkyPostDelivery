package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCellManagerUseCase_SyncCells(t *testing.T) {
	tests := []struct {
		name        string
		request     *entity.SyncCellsRequest
		setupMocks  func(*mocks.MockCellMappingInterface)
		expectedErr bool
	}{
		{
			name: "successful sync with external and internal cells",
			request: &entity.SyncCellsRequest{
				ParcelAutomatID: uuid.New().String(),
				CellsOut:        []string{uuid.New().String(), uuid.New().String()},
				CellsInternal:   []string{uuid.New().String()},
			},
			setupMocks: func(repo *mocks.MockCellMappingInterface) {
				repo.On("Sync", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "invalid automat ID",
			request: &entity.SyncCellsRequest{
				ParcelAutomatID: "invalid-uuid",
				CellsOut:        []string{uuid.New().String()},
			},
			setupMocks:  func(repo *mocks.MockCellMappingInterface) {},
			expectedErr: true,
		},
		{
			name: "invalid cell UUID",
			request: &entity.SyncCellsRequest{
				ParcelAutomatID: uuid.New().String(),
				CellsOut:        []string{"invalid-uuid"},
			},
			setupMocks:  func(repo *mocks.MockCellMappingInterface) {},
			expectedErr: true,
		},
		{
			name: "repo sync error",
			request: &entity.SyncCellsRequest{
				ParcelAutomatID: uuid.New().String(),
				CellsOut:        []string{uuid.New().String()},
			},
			setupMocks: func(repo *mocks.MockCellMappingInterface) {
				repo.On("Sync", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("sync failed"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockCellMappingInterface(t)
			mockArduino := mocks.NewMockArduinoInterface(t)
			mockDisplay := mocks.NewMockDisplayInterface(t)

			tt.setupMocks(mockRepo)

			uc := &CellManagerUseCase{
				cellRepo: mockRepo,
				arduino:  mockArduino,
				display:  mockDisplay,
				logger:   &mockLogger{},
			}

			err := uc.SyncCells(context.Background(), tt.request)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCellManagerUseCase_OpenCell(t *testing.T) {
	cellUUID := uuid.New()

	tests := []struct {
		name        string
		cellNumber  int
		orderNumber string
		setupMocks  func(*mocks.MockCellMappingInterface, *mocks.MockArduinoInterface, *mocks.MockDisplayInterface)
		expectedErr bool
	}{
		{
			name:        "successful open cell",
			cellNumber:  1,
			orderNumber: "ORDER-123",
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface, display *mocks.MockDisplayInterface) {
				repo.On("IsInitialized").Return(true)
				repo.On("GetCellUUID", 1).Return(cellUUID, nil)
				display.On("ShowCellOpening", 1, "ORDER-123").Return(nil)
				arduino.On("OpenCell", 1).Return(nil)
				display.On("ShowCellOpened", 1).Return(nil)
			},
			expectedErr: false,
		},
		{
			name:       "not initialized",
			cellNumber: 1,
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface, display *mocks.MockDisplayInterface) {
				repo.On("IsInitialized").Return(false)
			},
			expectedErr: true,
		},
		{
			name:       "cell not found",
			cellNumber: 999,
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface, display *mocks.MockDisplayInterface) {
				repo.On("IsInitialized").Return(true)
				repo.On("GetCellUUID", 999).Return(uuid.Nil, entityError.ErrCellNotFound)
			},
			expectedErr: true,
		},
		{
			name:       "arduino error",
			cellNumber: 1,
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface, display *mocks.MockDisplayInterface) {
				repo.On("IsInitialized").Return(true)
				repo.On("GetCellUUID", 1).Return(cellUUID, nil)
				display.On("ShowCellOpening", 1, "").Return(nil)
				arduino.On("OpenCell", 1).Return(errors.New("arduino error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockCellMappingInterface(t)
			mockArduino := mocks.NewMockArduinoInterface(t)
			mockDisplay := mocks.NewMockDisplayInterface(t)

			tt.setupMocks(mockRepo, mockArduino, mockDisplay)

			uc := &CellManagerUseCase{
				cellRepo: mockRepo,
				arduino:  mockArduino,
				display:  mockDisplay,
				logger:   &mockLogger{},
			}

			resp, err := uc.OpenCell(context.Background(), tt.cellNumber, tt.orderNumber)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.Success)
				assert.Equal(t, tt.cellNumber, resp.CellNumber)
			}
		})
	}
}

func TestCellManagerUseCase_OpenInternalDoor(t *testing.T) {
	doorUUID := uuid.New()

	tests := []struct {
		name        string
		doorNumber  int
		setupMocks  func(*mocks.MockCellMappingInterface, *mocks.MockArduinoInterface)
		expectedErr bool
	}{
		{
			name:       "successful open internal door",
			doorNumber: 1,
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface) {
				repo.On("IsInitialized").Return(true)
				repo.On("GetInternalCellUUID", 1).Return(doorUUID, nil)
				arduino.On("OpenInternalDoor", 1).Return(nil)
			},
			expectedErr: false,
		},
		{
			name:       "not initialized",
			doorNumber: 1,
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface) {
				repo.On("IsInitialized").Return(false)
			},
			expectedErr: true,
		},
		{
			name:       "door not found",
			doorNumber: 999,
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface) {
				repo.On("IsInitialized").Return(true)
				repo.On("GetInternalCellUUID", 999).Return(uuid.Nil, entityError.ErrCellNotFound)
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockCellMappingInterface(t)
			mockArduino := mocks.NewMockArduinoInterface(t)

			tt.setupMocks(mockRepo, mockArduino)

			uc := &CellManagerUseCase{
				cellRepo: mockRepo,
				arduino:  mockArduino,
				logger:   &mockLogger{},
			}

			resp, err := uc.OpenInternalDoor(context.Background(), tt.doorNumber)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.Success)
				assert.Equal(t, "internal", resp.Type)
			}
		})
	}
}

func TestCellManagerUseCase_PrepareCell(t *testing.T) {
	cellUUID := uuid.New()

	tests := []struct {
		name        string
		cellID      string
		cellType    entity.CellType
		setupMocks  func(*mocks.MockCellMappingInterface, *mocks.MockArduinoInterface, *mocks.MockDisplayInterface)
		expectedErr bool
	}{
		{
			name:     "successful prepare external cell",
			cellID:   cellUUID.String(),
			cellType: entity.CellTypeExternal,
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface, display *mocks.MockDisplayInterface) {
				repo.On("IsInitialized").Return(true)
				repo.On("GetCellNumber", cellUUID).Return(1, entity.CellTypeExternal, nil)
				repo.On("GetCellUUID", 1).Return(cellUUID, nil)
				display.On("ShowCellOpening", 1, "").Return(nil)
				arduino.On("OpenCell", 1).Return(nil)
				display.On("ShowCellOpened", 1).Return(nil)
			},
			expectedErr: false,
		},
		{
			name:     "successful prepare internal cell",
			cellID:   cellUUID.String(),
			cellType: entity.CellTypeInternal,
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface, display *mocks.MockDisplayInterface) {
				repo.On("IsInitialized").Return(true)
				repo.On("GetCellNumber", cellUUID).Return(1, entity.CellTypeInternal, nil)
				repo.On("GetInternalCellUUID", 1).Return(cellUUID, nil)
				arduino.On("OpenInternalDoor", 1).Return(nil)
			},
			expectedErr: false,
		},
		{
			name:   "invalid UUID",
			cellID: "invalid-uuid",
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface, display *mocks.MockDisplayInterface) {
				repo.On("IsInitialized").Return(true)
			},
			expectedErr: true,
		},
		{
			name:   "cell not found",
			cellID: cellUUID.String(),
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface, display *mocks.MockDisplayInterface) {
				repo.On("IsInitialized").Return(true)
				repo.On("GetCellNumber", cellUUID).Return(0, entity.CellType(""), entityError.ErrCellNotFound)
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockCellMappingInterface(t)
			mockArduino := mocks.NewMockArduinoInterface(t)
			mockDisplay := mocks.NewMockDisplayInterface(t)

			tt.setupMocks(mockRepo, mockArduino, mockDisplay)

			uc := &CellManagerUseCase{
				cellRepo: mockRepo,
				arduino:  mockArduino,
				display:  mockDisplay,
				logger:   &mockLogger{},
			}

			resp, err := uc.PrepareCell(context.Background(), tt.cellID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.Success)
			}
		})
	}
}

func TestCellManagerUseCase_GetCellsCount(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockCellMappingInterface, *mocks.MockArduinoInterface)
		expected      *entity.CellsCountResponse
		expectedError bool
	}{
		{
			name: "successful get cells count",
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface) {
				repo.On("GetMapping").Return(&entity.CellMapping{
					ExternalCells: map[int]uuid.UUID{1: uuid.New(), 2: uuid.New()},
					InternalCells: map[int]uuid.UUID{1: uuid.New()},
				})
				arduino.On("GetCellsCount").Return(5, nil)
			},
			expected: &entity.CellsCountResponse{
				CellsCount:          5,
				MappedCells:         2,
				InternalCellsCount:  0,
				MappedInternalCells: 1,
			},
		},
		{
			name: "arduino error",
			setupMocks: func(repo *mocks.MockCellMappingInterface, arduino *mocks.MockArduinoInterface) {
				repo.On("GetMapping").Return(&entity.CellMapping{
					ExternalCells: map[int]uuid.UUID{},
					InternalCells: map[int]uuid.UUID{},
				})
				arduino.On("GetCellsCount").Return(0, errors.New("arduino error"))
			},
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockCellMappingInterface(t)
			mockArduino := mocks.NewMockArduinoInterface(t)

			tt.setupMocks(mockRepo, mockArduino)

			uc := &CellManagerUseCase{
				cellRepo: mockRepo,
				arduino:  mockArduino,
				logger:   &mockLogger{},
			}

			resp, err := uc.GetCellsCount()

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expected.CellsCount, resp.CellsCount)
				assert.Equal(t, tt.expected.MappedCells, resp.MappedCells)
			}
		})
	}
}

type mockLogger struct{}

func (m *mockLogger) Info(message string, err error, fields ...map[string]any)  {}
func (m *mockLogger) Warn(message string, err error, fields ...map[string]any)  {}
func (m *mockLogger) Error(message string, err error, fields ...map[string]any) {}
func (m *mockLogger) Debug(message string, err error, fields ...map[string]any) {}
func (m *mockLogger) Fatal(message string, err error, fields ...map[string]any) {}
