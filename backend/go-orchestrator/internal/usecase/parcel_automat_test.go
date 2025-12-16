package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/request"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase/mocks"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/qr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestParcelAutomatUseCase_Create_Success(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)
	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	city := "Moscow"
	address := "Red Square, 1"
	ipAddress := "192.168.1.100"
	coordinates := "55.7558,37.6173"
	numberOfCells := 20

	cells := make([]request.CellDimensions, numberOfCells)
	for i := 0; i < numberOfCells; i++ {
		cells[i] = request.CellDimensions{
			Height: 40.0,
			Length: 40.0,
			Width:  40.0,
		}
	}

	automat := &entity.ParcelAutomat{
		ID:            uuid.New(),
		City:          city,
		Address:       address,
		IPAddress:     ipAddress,
		Coordinates:   coordinates,
		NumberOfCells: numberOfCells,
		IsWorking:     true,
	}

	mockParcelAutomatRepo.On("Create", ctx, mock.MatchedBy(func(a *entity.ParcelAutomat) bool {
		return a.City == city && a.Address == address && a.IPAddress == ipAddress && a.Coordinates == coordinates && a.NumberOfCells == numberOfCells && a.ArucoID == 101 && a.IsWorking == true
	})).Return(automat, nil)

	for i := 0; i < numberOfCells; i++ {
		mockLockerRepo.On("CreateWithNumber", ctx, mock.MatchedBy(func(c *entity.LockerCell) bool {
			return c.PostID == automat.ID && c.Height == 40.0 && c.Length == 40.0 && c.Width == 40.0
		}), i+1).Return(&entity.LockerCell{
			ID:     uuid.New(),
			PostID: automat.ID,
			Height: 40.0,
			Length: 40.0,
			Width:  40.0,
			Status: "available",
		}, nil).Once()
	}

	for i := 0; i < defaultInternalDoorCount; i++ {
		mockInternalLockerRepo.On("CreateWithNumber", ctx, mock.MatchedBy(func(c *entity.LockerCell) bool {
			return c.PostID == automat.ID && c.Height == 0.0 && c.Length == 0.0 && c.Width == 0.0
		}), i+1).Return(&entity.LockerCell{
			ID:     uuid.New(),
			PostID: automat.ID,
			Status: "available",
		}, nil).Once()
	}

	mockOrangePIWebAPI.On("SendCellUUIDs", ctx, ipAddress, automat.ID, mock.Anything, mock.Anything).Return(nil).Once()

	automatEntity := &entity.ParcelAutomat{
		City:          city,
		Address:       address,
		IPAddress:     ipAddress,
		Coordinates:   coordinates,
		NumberOfCells: numberOfCells,
		ArucoID:       101,
	}
	result, err := uc.Create(ctx, automatEntity, cells)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, city, result.City)
	assert.Equal(t, address, result.Address)
	assert.Equal(t, numberOfCells, result.NumberOfCells)
	assert.True(t, result.IsWorking)
	mockParcelAutomatRepo.AssertExpectations(t)
	mockLockerRepo.AssertExpectations(t)
	mockInternalLockerRepo.AssertExpectations(t)
	mockOrangePIWebAPI.AssertExpectations(t)
}

func TestParcelAutomatUseCase_Create_Error(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	city := "Moscow"
	address := "Red Square, 1"
	numberOfCells := 20

	cells := make([]request.CellDimensions, numberOfCells)
	for i := 0; i < numberOfCells; i++ {
		cells[i] = request.CellDimensions{
			Height: 40.0,
			Length: 40.0,
			Width:  40.0,
		}
	}

	mockParcelAutomatRepo.On("Create", ctx, mock.MatchedBy(func(a *entity.ParcelAutomat) bool {
		return a.City == city && a.Address == address && a.IPAddress == "" && a.Coordinates == "" && a.NumberOfCells == numberOfCells && a.ArucoID == 0 && a.IsWorking == true
	})).Return(nil, errors.New("database error"))

	automatEntity := &entity.ParcelAutomat{
		City:          city,
		Address:       address,
		IPAddress:     "",
		Coordinates:   "",
		NumberOfCells: numberOfCells,
		ArucoID:       0,
	}
	result, err := uc.Create(ctx, automatEntity, cells)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	mockParcelAutomatRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_Create_CustomCellDimensions(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	city := "Ekaterinburg"
	address := "Lenina St, 10"
	numberOfCells := 3

	cells := []request.CellDimensions{
		{Height: 20.0, Length: 30.0, Width: 25.0},
		{Height: 40.0, Length: 50.0, Width: 45.0},
		{Height: 60.0, Length: 70.0, Width: 65.0},
	}

	automat := &entity.ParcelAutomat{
		ID:            uuid.New(),
		City:          city,
		Address:       address,
		NumberOfCells: numberOfCells,
		IsWorking:     true,
	}

	mockParcelAutomatRepo.On("Create", ctx, mock.MatchedBy(func(a *entity.ParcelAutomat) bool {
		return a.City == city && a.Address == address && a.IPAddress == "" && a.Coordinates == "" && a.NumberOfCells == numberOfCells && a.ArucoID == 42 && a.IsWorking == true
	})).Return(automat, nil)

	for i := 0; i < numberOfCells; i++ {
		mockLockerRepo.On("CreateWithNumber", ctx, mock.MatchedBy(func(c *entity.LockerCell) bool {
			return c.PostID == automat.ID && c.Height == cells[i].Height && c.Length == cells[i].Length && c.Width == cells[i].Width
		}), i+1).Return(&entity.LockerCell{
			ID:     uuid.New(),
			PostID: automat.ID,
			Height: cells[i].Height,
			Length: cells[i].Length,
			Width:  cells[i].Width,
			Status: "available",
		}, nil).Once()
	}

	for i := 0; i < defaultInternalDoorCount; i++ {
		mockInternalLockerRepo.On("CreateWithNumber", ctx, mock.MatchedBy(func(c *entity.LockerCell) bool {
			return c.PostID == automat.ID && c.Height == 0.0 && c.Length == 0.0 && c.Width == 0.0
		}), i+1).Return(&entity.LockerCell{
			ID:     uuid.New(),
			PostID: automat.ID,
			Status: "available",
		}, nil).Once()
	}

	automatEntity := &entity.ParcelAutomat{
		City:          city,
		Address:       address,
		IPAddress:     "",
		Coordinates:   "",
		NumberOfCells: numberOfCells,
		ArucoID:       42,
	}
	result, err := uc.Create(ctx, automatEntity, cells)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, city, result.City)
	assert.Equal(t, address, result.Address)
	assert.Equal(t, numberOfCells, result.NumberOfCells)
	assert.True(t, result.IsWorking)
	mockParcelAutomatRepo.AssertExpectations(t)
	mockLockerRepo.AssertExpectations(t)
	mockInternalLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_Create_ErrorCreatingCell(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	city := "Moscow"
	address := "Test St, 5"
	numberOfCells := 2

	cells := []request.CellDimensions{
		{Height: 40.0, Length: 40.0, Width: 40.0},
		{Height: 40.0, Length: 40.0, Width: 40.0},
	}

	automat := &entity.ParcelAutomat{
		ID:            uuid.New(),
		City:          city,
		Address:       address,
		NumberOfCells: numberOfCells,
		IsWorking:     true,
	}

	mockParcelAutomatRepo.On("Create", ctx, mock.MatchedBy(func(a *entity.ParcelAutomat) bool {
		return a.City == city && a.Address == address && a.IPAddress == "" && a.Coordinates == "" && a.NumberOfCells == numberOfCells && a.ArucoID == 199 && a.IsWorking == true
	})).Return(automat, nil)
	mockLockerRepo.On("CreateWithNumber", ctx, mock.MatchedBy(func(c *entity.LockerCell) bool {
		return c.PostID == automat.ID && c.Height == cells[0].Height && c.Length == cells[0].Length && c.Width == cells[0].Width
	}), 1).Return(nil, errors.New("cell creation error")).Once()
	mockParcelAutomatRepo.On("Delete", ctx, automat.ID).Return(nil)

	automatEntity := &entity.ParcelAutomat{
		City:          city,
		Address:       address,
		IPAddress:     "",
		Coordinates:   "",
		NumberOfCells: numberOfCells,
		ArucoID:       199,
	}
	result, err := uc.Create(ctx, automatEntity, cells)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "CreateLockerCell")
	mockParcelAutomatRepo.AssertExpectations(t)
	mockLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_GetByID_Success(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	automatID := uuid.New()

	automat := &entity.ParcelAutomat{
		ID:            automatID,
		City:          "Moscow",
		Address:       "Test Address",
		NumberOfCells: 20,
		IsWorking:     true,
	}

	mockParcelAutomatRepo.On("GetByID", ctx, automatID).Return(automat, nil)

	result, err := uc.GetByID(ctx, automatID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, automatID, result.ID)
	mockParcelAutomatRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_List_Success(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()

	automats := []*entity.ParcelAutomat{
		{ID: uuid.New(), City: "Moscow", Address: "Address 1", NumberOfCells: 20, IsWorking: true},
		{ID: uuid.New(), City: "SPB", Address: "Address 2", NumberOfCells: 15, IsWorking: false},
	}

	mockParcelAutomatRepo.On("List", ctx).Return(automats, nil)

	result, err := uc.List(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	mockParcelAutomatRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_ListWorking_Success(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()

	automats := []*entity.ParcelAutomat{
		{ID: uuid.New(), City: "Moscow", Address: "Address 1", NumberOfCells: 20, IsWorking: true},
	}

	mockParcelAutomatRepo.On("ListWorking", ctx).Return(automats, nil)

	result, err := uc.ListWorking(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.True(t, result[0].IsWorking)
	mockParcelAutomatRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_UpdateStatus_Success(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	automatID := uuid.New()
	isWorking := false

	automat := &entity.ParcelAutomat{
		ID:            automatID,
		City:          "Moscow",
		Address:       "Test Address",
		NumberOfCells: 20,
		IsWorking:     isWorking,
	}

	automat.IsWorking = isWorking
	mockParcelAutomatRepo.On("UpdateStatus", ctx, mock.MatchedBy(func(a *entity.ParcelAutomat) bool {
		return a.ID == automatID && a.IsWorking == isWorking
	})).Return(automat, nil)

	result, err := uc.UpdateStatus(ctx, automat)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, automatID, result.ID)
	assert.False(t, result.IsWorking)
	mockParcelAutomatRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_Delete_Success(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	automatID := uuid.New()

	mockParcelAutomatRepo.On("Delete", ctx, automatID).Return(nil)

	err := uc.Delete(ctx, automatID)

	assert.NoError(t, err)
	mockParcelAutomatRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_GetAutomatCells_Success(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	automatID := uuid.New()

	cells := []*entity.LockerCell{
		{ID: uuid.New(), PostID: automatID, Status: "available"},
		{ID: uuid.New(), PostID: automatID, Status: "occupied"},
	}

	mockLockerRepo.On("ListCellsByPostID", ctx, automatID).Return(cells, nil)

	result, err := uc.GetAutomatCells(ctx, automatID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	mockLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_ProcessQRScan_Success(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRGenerator := new(mocks.MockQRGenerator)
	mockUserRepo := new(mocks.MockUserRepo)
	mockMinioClient := new(mocks.MockMinioClient)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)

	mockLogger := new(mocks.MockLogger)
	qrUseCase := NewQRUseCase(mockQRGenerator, mockUserRepo, mockMinioClient, mockLogger)
	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, qrUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	userID := uuid.New()
	orderID := uuid.New()
	deliveryID := uuid.New()
	cellID := uuid.New()
	automatID := uuid.New()

	email := "test@test.com"
	user := &entity.User{
		ID:       userID,
		FullName: "Test User",
		Email:    &email,
	}

	orders := []*entity.Order{
		{ID: orderID, UserID: userID, Status: "delivered", ParcelAutomatID: automatID, LockerCellID: &cellID},
	}

	delivery := &entity.Delivery{
		ID:              deliveryID,
		OrderID:         orderID,
		ParcelAutomatID: automatID,
		Status:          "delivered",
	}

	cell := &entity.LockerCell{
		ID:     cellID,
		PostID: automatID,
		Status: "occupied",
	}

	qrData := `{"user_id":"` + userID.String() + `","email":"test@test.com","full_name":"Test User","exp":9999999999}`

	mockQRGenerator.On("ValidateQRCode", qrData).Return(&qr.QRData{UserID: userID, Email: email, FullName: "Test User"}, nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockOrderRepo.On("ListByUserID", ctx, userID).Return(orders, nil)
	mockDeliveryRepo.On("GetByOrderID", ctx, orderID).Return(delivery, nil)
	mockLockerRepo.On("GetCellByID", ctx, cellID).Return(cell, nil)
	mockLockerRepo.On("UpdateCellStatus", ctx, mock.MatchedBy(func(c *entity.LockerCell) bool {
		return c.ID == cellID && c.Status == "opened"
	})).Return(nil)

	result, err := uc.ProcessQRScan(ctx, qrData, automatID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, cellID)
	mockQRGenerator.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
	mockDeliveryRepo.AssertExpectations(t)
	mockLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_ProcessQRScan_InvalidQR(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRGenerator := new(mocks.MockQRGenerator)
	mockUserRepo := new(mocks.MockUserRepo)
	mockMinioClient := new(mocks.MockMinioClient)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)

	mockLogger := new(mocks.MockLogger)
	qrUseCase := NewQRUseCase(mockQRGenerator, mockUserRepo, mockMinioClient, mockLogger)
	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, qrUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	automatID := uuid.New()
	qrData := `invalid json`

	mockQRGenerator.On("ValidateQRCode", qrData).Return(nil, errors.New("invalid QR code"))

	result, err := uc.ProcessQRScan(ctx, qrData, automatID)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockQRGenerator.AssertExpectations(t)
}

func TestParcelAutomatUseCase_ConfirmPickup_Success(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	cellID := uuid.New()

	cell := &entity.LockerCell{
		ID:     cellID,
		Status: "opened",
	}

	mockLockerRepo.On("GetCellByID", ctx, cellID).Return(cell, nil)
	mockLockerRepo.On("UpdateCellStatus", ctx, mock.MatchedBy(func(c *entity.LockerCell) bool {
		return c.ID == cellID && c.Status == "available"
	})).Return(nil)
	mockOrderRepo.On("GetByLockerCellID", ctx, cellID).Return(nil, errors.New("order not found"))
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	err := uc.ConfirmPickup(ctx, []uuid.UUID{cellID})

	assert.NoError(t, err)
	mockLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_ConfirmPickup_CellNotOpened(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	cellID := uuid.New()

	cell := &entity.LockerCell{
		ID:     cellID,
		Status: "occupied",
	}

	mockLockerRepo.On("GetCellByID", ctx, cellID).Return(cell, nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	err := uc.ConfirmPickup(ctx, []uuid.UUID{cellID})

	assert.NoError(t, err)
	mockLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_Update_Success(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	automatID := uuid.New()

	automat := &entity.ParcelAutomat{
		ID:          automatID,
		City:        "Updated City",
		Address:     "Updated Address",
		IPAddress:   "192.168.1.2",
		Coordinates: "55.7558,37.6173",
	}

	automatEntity := &entity.ParcelAutomat{
		ID:          automatID,
		City:        "Updated City",
		Address:     "Updated Address",
		IPAddress:   "192.168.1.2",
		Coordinates: "55.7558,37.6173",
	}
	mockParcelAutomatRepo.On("Update", ctx, mock.MatchedBy(func(a *entity.ParcelAutomat) bool {
		return a.ID == automatID && a.City == "Updated City" && a.Address == "Updated Address" && a.IPAddress == "192.168.1.2" && a.Coordinates == "55.7558,37.6173"
	})).Return(automat, nil)
	mockLockerRepo.On("ListCellsByPostID", ctx, automatID).Return([]*entity.LockerCell{}, nil).Once()
	mockInternalLockerRepo.On("ListCellsByPostID", ctx, automatID).Return([]*entity.LockerCell{}, nil).Once()
	mockOrangePIWebAPI.On("SendCellUUIDs", ctx, "192.168.1.2", automatID, mock.Anything, mock.Anything).Return(nil).Once()
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	result, err := uc.Update(ctx, automatEntity)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, automatID, result.ID)
	mockParcelAutomatRepo.AssertExpectations(t)
	mockLockerRepo.AssertExpectations(t)
	mockInternalLockerRepo.AssertExpectations(t)
	mockOrangePIWebAPI.AssertExpectations(t)
}

func TestParcelAutomatUseCase_UpdateCell_Success(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	cellID := uuid.New()

	existingCell := &entity.LockerCell{
		ID:     cellID,
		Status: "available",
		Height: 10.0,
		Length: 10.0,
		Width:  10.0,
	}

	updatedCell := &entity.LockerCell{
		ID:     cellID,
		Status: "available",
		Height: 15.0,
		Length: 15.0,
		Width:  15.0,
	}

	mockLockerRepo.On("GetCellByID", ctx, cellID).Return(existingCell, nil)
	mockLockerRepo.On("UpdateDimensions", ctx, mock.MatchedBy(func(c *entity.LockerCell) bool {
		return c.ID == cellID && c.Height == 15.0 && c.Length == 15.0 && c.Width == 15.0
	})).Return(updatedCell, nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	result, err := uc.UpdateCell(ctx, cellID, 15.0, 15.0, 15.0)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, cellID, result.ID)
	assert.Equal(t, 15.0, result.Height)
	mockLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_PrepareCell_Success(t *testing.T) {
	mockParcelAutomatRepo := new(mocks.MockParcelAutomatRepo)
	mockLockerRepo := new(mocks.MockLockerRepo)
	mockInternalLockerRepo := new(mocks.MockInternalLockerRepo)
	mockOrderRepo := new(mocks.MockOrderRepo)
	mockDeliveryRepo := new(mocks.MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(mocks.MockOrangePIWebAPI)
	mockLogger := new(mocks.MockLogger)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockInternalLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI, mockLogger)

	ctx := context.Background()
	orderID := uuid.New()
	parcelAutomatID := uuid.New()
	cellID := uuid.New()
	internalDoorID := uuid.New()

	order := &entity.Order{
		ID:           orderID,
		LockerCellID: &cellID,
	}

	cell := &entity.LockerCell{
		ID:     cellID,
		Status: "reserved",
	}

	automat := &entity.ParcelAutomat{
		ID:        parcelAutomatID,
		IPAddress: "192.168.1.1",
	}

	delivery := &entity.Delivery{
		ID:                   uuid.New(),
		OrderID:              orderID,
		ParcelAutomatID:      parcelAutomatID,
		InternalLockerCellID: &internalDoorID,
	}

	internalCell := &entity.LockerCell{
		ID:     internalDoorID,
		Status: "available",
	}
	mockOrderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	mockLockerRepo.On("GetCellByID", ctx, cellID).Return(cell, nil)
	mockParcelAutomatRepo.On("GetByID", ctx, parcelAutomatID).Return(automat, nil)
	mockDeliveryRepo.On("GetByOrderID", ctx, orderID).Return(delivery, nil)
	mockOrangePIWebAPI.On("OpenCell", ctx, "192.168.1.1", internalDoorID).Return(nil)
	mockInternalLockerRepo.On("GetCellByID", ctx, internalDoorID).Return(internalCell, nil)
	mockInternalLockerRepo.On("UpdateCellStatus", ctx, mock.MatchedBy(func(c *entity.LockerCell) bool {
		return c.ID == internalDoorID && c.Status == "opened"
	})).Return(nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return()

	result, internalResult, err := uc.PrepareCell(ctx, orderID, parcelAutomatID)

	assert.NoError(t, err)
	assert.Equal(t, cellID, result)
	assert.NotNil(t, internalResult)
	assert.Equal(t, internalDoorID, *internalResult)
	mockOrderRepo.AssertExpectations(t)
	mockLockerRepo.AssertExpectations(t)
	mockInternalLockerRepo.AssertExpectations(t)
	mockParcelAutomatRepo.AssertExpectations(t)
	mockDeliveryRepo.AssertExpectations(t)
	mockOrangePIWebAPI.AssertExpectations(t)
}
