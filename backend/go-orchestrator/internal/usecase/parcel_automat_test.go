package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/skr1ms/hitech-ekb/internal/controller/http/v1/request"
	"github.com/skr1ms/hitech-ekb/internal/entity"
	"github.com/skr1ms/hitech-ekb/pkg/qr"
	"github.com/stretchr/testify/assert"
)

func TestParcelAutomatUseCase_Create_Success(t *testing.T) {
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

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

	mockParcelAutomatRepo.On("Create", ctx, city, address, numberOfCells, ipAddress, coordinates, 101, true).Return(automat, nil)

	cellUUIDs := make([]uuid.UUID, numberOfCells)
	for i := 0; i < numberOfCells; i++ {
		cellUUIDs[i] = uuid.New()
		mockLockerRepo.On("Create", ctx, automat.ID, 40.0, 40.0, 40.0).Return(&entity.LockerCell{
			ID:     cellUUIDs[i],
			PostID: automat.ID,
			Height: 40.0,
			Length: 40.0,
			Width:  40.0,
			Status: "available",
		}, nil).Once()
	}

	mockOrangePIWebAPI.On("SendCellUUIDs", ctx, ipAddress, cellUUIDs).Return(nil).Once()

	result, err := uc.Create(ctx, city, address, ipAddress, coordinates, numberOfCells, 101, cells)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, city, result.City)
	assert.Equal(t, address, result.Address)
	assert.Equal(t, numberOfCells, result.NumberOfCells)
	assert.True(t, result.IsWorking)
	mockParcelAutomatRepo.AssertExpectations(t)
	mockLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_Create_Error(t *testing.T) {
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

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

	mockParcelAutomatRepo.On("Create", ctx, city, address, numberOfCells, "", "", 0, true).Return(nil, errors.New("database error"))

	result, err := uc.Create(ctx, city, address, "", "", numberOfCells, 0, cells)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	mockParcelAutomatRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_Create_CustomCellDimensions(t *testing.T) {
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

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

	mockParcelAutomatRepo.On("Create", ctx, city, address, numberOfCells, "", "", 42, true).Return(automat, nil)

	for i := 0; i < numberOfCells; i++ {
		mockLockerRepo.On("Create", ctx, automat.ID, cells[i].Height, cells[i].Length, cells[i].Width).Return(&entity.LockerCell{
			ID:     uuid.New(),
			PostID: automat.ID,
			Height: cells[i].Height,
			Length: cells[i].Length,
			Width:  cells[i].Width,
			Status: "available",
		}, nil).Once()
	}

	result, err := uc.Create(ctx, city, address, "", "", numberOfCells, 42, cells)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, city, result.City)
	assert.Equal(t, address, result.Address)
	assert.Equal(t, numberOfCells, result.NumberOfCells)
	assert.True(t, result.IsWorking)
	mockParcelAutomatRepo.AssertExpectations(t)
	mockLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_Create_ErrorCreatingCell(t *testing.T) {
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

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

	mockParcelAutomatRepo.On("Create", ctx, city, address, numberOfCells, "", "", 199, true).Return(automat, nil)
	mockLockerRepo.On("Create", ctx, automat.ID, cells[0].Height, cells[0].Length, cells[0].Width).Return(nil, errors.New("cell creation error")).Once()

	result, err := uc.Create(ctx, city, address, "", "", numberOfCells, 199, cells)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "lockerRepo.Create")
	mockParcelAutomatRepo.AssertExpectations(t)
	mockLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_GetByID_Success(t *testing.T) {
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

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
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

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
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

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
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

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

	mockParcelAutomatRepo.On("UpdateStatus", ctx, automatID, isWorking).Return(automat, nil)

	result, err := uc.UpdateStatus(ctx, automatID, isWorking)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, automatID, result.ID)
	assert.False(t, result.IsWorking)
	mockParcelAutomatRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_Delete_Success(t *testing.T) {
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

	ctx := context.Background()
	automatID := uuid.New()

	mockParcelAutomatRepo.On("Delete", ctx, automatID).Return(nil)

	err := uc.Delete(ctx, automatID)

	assert.NoError(t, err)
	mockParcelAutomatRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_GetAutomatCells_Success(t *testing.T) {
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

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
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRGenerator := new(MockQRGenerator)
	mockUserRepo := new(MockUserRepo)
	mockMinioClient := new(MockMinioClient)

	qrUseCase := NewQRUseCase(mockQRGenerator, mockUserRepo, mockMinioClient)
	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, qrUseCase, new(MockOrangePIWebAPI))

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
	mockLockerRepo.On("UpdateCellStatus", ctx, cellID, "opened").Return(nil)

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
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRGenerator := new(MockQRGenerator)
	mockUserRepo := new(MockUserRepo)
	mockMinioClient := new(MockMinioClient)

	qrUseCase := NewQRUseCase(mockQRGenerator, mockUserRepo, mockMinioClient)
	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, qrUseCase, new(MockOrangePIWebAPI))

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
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

	ctx := context.Background()
	cellID := uuid.New()

	cell := &entity.LockerCell{
		ID:     cellID,
		Status: "opened",
	}

	mockLockerRepo.On("GetCellByID", ctx, cellID).Return(cell, nil)
	mockLockerRepo.On("UpdateCellStatus", ctx, cellID, "available").Return(nil)

	err := uc.ConfirmPickup(ctx, []uuid.UUID{cellID})

	assert.NoError(t, err)
	mockLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_ConfirmPickup_CellNotOpened(t *testing.T) {
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

	ctx := context.Background()
	cellID := uuid.New()

	cell := &entity.LockerCell{
		ID:     cellID,
		Status: "occupied",
	}

	mockLockerRepo.On("GetCellByID", ctx, cellID).Return(cell, nil)

	err := uc.ConfirmPickup(ctx, []uuid.UUID{cellID})

	assert.NoError(t, err)
	mockLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_Update_Success(t *testing.T) {
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

	ctx := context.Background()
	automatID := uuid.New()

	automat := &entity.ParcelAutomat{
		ID:          automatID,
		City:        "Updated City",
		Address:     "Updated Address",
		IPAddress:   "192.168.1.2",
		Coordinates: "55.7558,37.6173",
	}

	mockParcelAutomatRepo.On("Update", ctx, automatID, "Updated City", "Updated Address", "192.168.1.2", "55.7558,37.6173").Return(automat, nil)

	result, err := uc.Update(ctx, automatID, "Updated City", "Updated Address", "192.168.1.2", "55.7558,37.6173")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, automatID, result.ID)
	mockParcelAutomatRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_UpdateCell_Success(t *testing.T) {
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

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
	mockLockerRepo.On("UpdateDimensions", ctx, cellID, 15.0, 15.0, 15.0).Return(updatedCell, nil)

	result, err := uc.UpdateCell(ctx, cellID, 15.0, 15.0, 15.0)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, cellID, result.ID)
	assert.Equal(t, 15.0, result.Height)
	mockLockerRepo.AssertExpectations(t)
}

func TestParcelAutomatUseCase_PrepareCell_Success(t *testing.T) {
	mockParcelAutomatRepo := new(MockParcelAutomatRepo)
	mockLockerRepo := new(MockLockerRepo)
	mockOrderRepo := new(MockOrderRepo)
	mockDeliveryRepo := new(MockDeliveryRepo)
	mockQRUseCase := &QRUseCase{}
	mockOrangePIWebAPI := new(MockOrangePIWebAPI)

	uc := NewParcelAutomatUseCase(mockParcelAutomatRepo, mockLockerRepo, mockOrderRepo, mockDeliveryRepo, mockQRUseCase, mockOrangePIWebAPI)

	ctx := context.Background()
	orderID := uuid.New()
	parcelAutomatID := uuid.New()
	cellID := uuid.New()

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

	mockOrderRepo.On("GetByID", ctx, orderID).Return(order, nil)
	mockLockerRepo.On("GetCellByID", ctx, cellID).Return(cell, nil)
	mockParcelAutomatRepo.On("GetByID", ctx, parcelAutomatID).Return(automat, nil)
	mockOrangePIWebAPI.On("OpenCell", ctx, "192.168.1.1", cellID).Return(nil)
	mockLockerRepo.On("UpdateCellStatus", ctx, cellID, "opened").Return(nil)

	result, err := uc.PrepareCell(ctx, orderID, parcelAutomatID)

	assert.NoError(t, err)
	assert.Equal(t, cellID, result)
	mockOrderRepo.AssertExpectations(t)
	mockLockerRepo.AssertExpectations(t)
	mockParcelAutomatRepo.AssertExpectations(t)
	mockOrangePIWebAPI.AssertExpectations(t)
}
