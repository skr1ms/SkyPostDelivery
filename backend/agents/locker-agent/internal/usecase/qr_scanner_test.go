package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQRScannerUseCase_ProcessQRScan(t *testing.T) {
	tests := []struct {
		name        string
		qrData      string
		setupMocks  func(*mocks.MockClientInterface, *mocks.MockDisplayInterface, *mocks.MockCellMappingInterface)
		expectedErr bool
	}{
		{
			name:   "successful QR scan processing",
			qrData: "valid-qr-data",
			setupMocks: func(orchestrator *mocks.MockClientInterface, display *mocks.MockDisplayInterface, cellRepo *mocks.MockCellMappingInterface) {
				cellUUID := uuid.New()
				cellRepo.On("GetMapping").Return(&entity.CellMapping{
					ParcelAutomatID: uuid.New(),
					ExternalCells:   map[int]uuid.UUID{1: cellUUID},
				})
				display.On("ShowScanning").Return(nil)
				orchestrator.On("ValidateQR", mock.Anything, mock.Anything, mock.Anything).Return(&entity.QRValidationResponse{
					Success: true,
					Message: "Valid",
					CellIDs: []string{cellUUID.String()},
				}, nil)
				display.On("ShowSuccess", mock.Anything).Return(nil)
				cellRepo.On("GetCellNumber", cellUUID).Return(1, entity.CellTypeExternal, nil)
				cellRepo.On("GetCellUUID", 1).Return(cellUUID, nil)
				display.On("ShowCellOpening", 1, "").Return(nil)
				display.On("ShowCellOpened", 1).Return(nil)
				display.On("ShowPleaseClose").Return(nil)
			},
			expectedErr: false,
		},
		{
			name:   "orchestrator validation failed",
			qrData: "some-qr-data",
			setupMocks: func(orchestrator *mocks.MockClientInterface, display *mocks.MockDisplayInterface, cellRepo *mocks.MockCellMappingInterface) {
				cellRepo.On("GetMapping").Return(&entity.CellMapping{
					ParcelAutomatID: uuid.New(),
					ExternalCells:   map[int]uuid.UUID{},
				})
				display.On("ShowScanning").Return(nil)
				orchestrator.On("ValidateQR", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("validation failed"))
				display.On("ShowError", mock.Anything).Return(nil)
			},
			expectedErr: true,
		},
		{
			name:   "orchestrator returned invalid status",
			qrData: "invalid-qr",
			setupMocks: func(orchestrator *mocks.MockClientInterface, display *mocks.MockDisplayInterface, cellRepo *mocks.MockCellMappingInterface) {
				cellRepo.On("GetMapping").Return(&entity.CellMapping{
					ParcelAutomatID: uuid.New(),
					ExternalCells:   map[int]uuid.UUID{},
				})
				display.On("ShowScanning").Return(nil)
				orchestrator.On("ValidateQR", mock.Anything, mock.Anything, mock.Anything).Return(&entity.QRValidationResponse{
					Success: false,
					Message: "Invalid QR",
				}, nil)
				display.On("ShowInvalid").Return(nil)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrchestrator := mocks.NewMockClientInterface(t)
			mockDisplay := mocks.NewMockDisplayInterface(t)
			mockRepo := mocks.NewMockCellMappingInterface(t)
			mockArduino := mocks.NewMockArduinoInterface(t)

			mockCellManager := &CellManagerUseCase{
				cellRepo: mockRepo,
				arduino:  mockArduino,
				display:  mockDisplay,
				logger:   &mockLogger{},
			}

			mockRepo.On("IsInitialized").Return(true)
			mockArduino.On("OpenCell", mock.Anything).Return(nil).Maybe()
			tt.setupMocks(mockOrchestrator, mockDisplay, mockRepo)

			uc := &QRScannerUseCase{
				cellManager:        mockCellManager,
				orchestratorClient: mockOrchestrator,
				display:            mockDisplay,
				cellRepo:           mockRepo,
				logger:             &mockLogger{},
			}

			resp, err := uc.ProcessQRScan(context.Background(), tt.qrData)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestQRScannerUseCase_ConfirmPickup(t *testing.T) {
	cellIDs := []string{uuid.New().String(), uuid.New().String()}

	tests := []struct {
		name        string
		cellIDs     []string
		setupMocks  func(*mocks.MockClientInterface, *mocks.MockDisplayInterface, *mocks.MockCellMappingInterface)
		expectedErr bool
	}{
		{
			name:    "successful pickup confirmation",
			cellIDs: cellIDs,
			setupMocks: func(orchestrator *mocks.MockClientInterface, display *mocks.MockDisplayInterface, cellRepo *mocks.MockCellMappingInterface) {
				orchestrator.On("ConfirmPickup", mock.Anything, mock.Anything).Return(nil)
				display.On("ShowThankYou").Return(nil)
			},
			expectedErr: false,
		},
		{
			name:    "orchestrator error",
			cellIDs: cellIDs,
			setupMocks: func(orchestrator *mocks.MockClientInterface, display *mocks.MockDisplayInterface, cellRepo *mocks.MockCellMappingInterface) {
				orchestrator.On("ConfirmPickup", mock.Anything, mock.Anything).Return(errors.New("orchestrator error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrchestrator := mocks.NewMockClientInterface(t)
			mockDisplay := mocks.NewMockDisplayInterface(t)
			mockRepo := mocks.NewMockCellMappingInterface(t)

			tt.setupMocks(mockOrchestrator, mockDisplay, mockRepo)

			uc := &QRScannerUseCase{
				orchestratorClient: mockOrchestrator,
				display:            mockDisplay,
				cellRepo:           mockRepo,
				logger:             &mockLogger{},
			}

			err := uc.ConfirmPickup(context.Background(), tt.cellIDs)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestQRScannerUseCase_ConfirmLoaded(t *testing.T) {
	orderID := uuid.New().String()
	cellID := uuid.New().String()

	tests := []struct {
		name        string
		orderID     string
		cellID      string
		setupMocks  func(*mocks.MockClientInterface)
		expectedErr bool
	}{
		{
			name:    "successful load confirmation",
			orderID: orderID,
			cellID:  cellID,
			setupMocks: func(orchestrator *mocks.MockClientInterface) {
				orchestrator.On("ConfirmLoaded", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedErr: false,
		},
		{
			name:    "orchestrator error",
			orderID: orderID,
			cellID:  cellID,
			setupMocks: func(orchestrator *mocks.MockClientInterface) {
				orchestrator.On("ConfirmLoaded", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("orchestrator error"))
			},
			expectedErr: true,
		},
		{
			name:    "invalid order ID",
			orderID: "invalid-uuid",
			cellID:  cellID,
			setupMocks: func(orchestrator *mocks.MockClientInterface) {
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrchestrator := mocks.NewMockClientInterface(t)

			tt.setupMocks(mockOrchestrator)

			uc := &QRScannerUseCase{
				orchestratorClient: mockOrchestrator,
				logger:             &mockLogger{},
			}

			err := uc.ConfirmLoaded(context.Background(), tt.orderID, tt.cellID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
