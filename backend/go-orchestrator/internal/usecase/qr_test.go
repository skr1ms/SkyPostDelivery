package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase/mocks"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/qr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQRUseCase_GenerateQR_Success(t *testing.T) {
	mockQRGenerator := new(mocks.MockQRGenerator)
	mockUserRepo := new(mocks.MockUserRepo)
	mockMinioClient := new(mocks.MockMinioClient)
	mockLogger := new(mocks.MockLogger)

	uc := NewQRUseCase(mockQRGenerator, mockUserRepo, mockMinioClient, mockLogger)

	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"
	name := "Test User"

	now := time.Now()
	qrData := &qr.QRData{
		UserID:    userID,
		Email:     email,
		FullName:  name,
		IssuedAt:  now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	qrImageBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

	user := &entity.User{
		ID:       userID,
		FullName: name,
		Email:    &email,
	}
	mockQRGenerator.On("GenerateQRCode", userID, email, name).Return(qrData, qrImageBase64, nil)
	mockMinioClient.On("UploadFile", ctx, mock.Anything, mock.Anything, mock.Anything, "image/png").Return(nil)

	result, image, err := uc.GenerateQR(ctx, user)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, name, result.Name)
	assert.Equal(t, qrImageBase64, image)
	mockQRGenerator.AssertExpectations(t)
	mockMinioClient.AssertExpectations(t)
}

func TestQRUseCase_GenerateQR_GeneratorError(t *testing.T) {
	mockQRGenerator := new(mocks.MockQRGenerator)
	mockUserRepo := new(mocks.MockUserRepo)
	mockMinioClient := new(mocks.MockMinioClient)
	mockLogger := new(mocks.MockLogger)

	uc := NewQRUseCase(mockQRGenerator, mockUserRepo, mockMinioClient, mockLogger)

	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"
	name := "Test User"

	user := &entity.User{
		ID:       userID,
		FullName: name,
		Email:    &email,
	}
	mockQRGenerator.On("GenerateQRCode", userID, email, name).Return(nil, "", errors.New("qr generation failed"))

	result, image, err := uc.GenerateQR(ctx, user)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Empty(t, image)
	assert.Contains(t, err.Error(), "qr generation failed")
	mockQRGenerator.AssertExpectations(t)
}

func TestQRUseCase_ValidateQR_Success(t *testing.T) {
	mockQRGenerator := new(mocks.MockQRGenerator)
	mockUserRepo := new(mocks.MockUserRepo)
	mockMinioClient := new(mocks.MockMinioClient)
	mockLogger := new(mocks.MockLogger)

	uc := NewQRUseCase(mockQRGenerator, mockUserRepo, mockMinioClient, mockLogger)

	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"
	qrDataJSON := `{"user_id":"123","email":"test@example.com"}`

	qrData := &qr.QRData{
		UserID: userID,
		Email:  email,
	}

	user := &entity.User{
		ID:       userID,
		Email:    &email,
		FullName: "Test User",
	}

	mockQRGenerator.On("ValidateQRCode", qrDataJSON).Return(qrData, nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	result, err := uc.ValidateQR(ctx, qrDataJSON)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, email, *result.Email)
	mockQRGenerator.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestQRUseCase_ValidateQR_InvalidQR(t *testing.T) {
	mockQRGenerator := new(mocks.MockQRGenerator)
	mockUserRepo := new(mocks.MockUserRepo)
	mockMinioClient := new(mocks.MockMinioClient)
	mockLogger := new(mocks.MockLogger)

	uc := NewQRUseCase(mockQRGenerator, mockUserRepo, mockMinioClient, mockLogger)

	ctx := context.Background()
	qrDataJSON := `{"user_id":"invalid"}`

	mockQRGenerator.On("ValidateQRCode", qrDataJSON).Return(nil, errors.New("invalid qr code"))

	result, err := uc.ValidateQR(ctx, qrDataJSON)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid qr code")
	mockQRGenerator.AssertExpectations(t)
}

func TestQRUseCase_ValidateQR_EmailMismatch(t *testing.T) {
	mockQRGenerator := new(mocks.MockQRGenerator)
	mockUserRepo := new(mocks.MockUserRepo)
	mockMinioClient := new(mocks.MockMinioClient)
	mockLogger := new(mocks.MockLogger)

	uc := NewQRUseCase(mockQRGenerator, mockUserRepo, mockMinioClient, mockLogger)

	ctx := context.Background()
	userID := uuid.New()
	qrDataJSON := `{"user_id":"123","email":"test@example.com"}`

	qrData := &qr.QRData{
		UserID: userID,
		Email:  "test@example.com",
	}

	email := "different@example.com"
	user := &entity.User{
		ID:       userID,
		Email:    &email,
		FullName: "Test User",
	}

	mockQRGenerator.On("ValidateQRCode", qrDataJSON).Return(qrData, nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	result, err := uc.ValidateQR(ctx, qrDataJSON)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user mismatch")
	mockQRGenerator.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestQRUseCase_RefreshQR_Success(t *testing.T) {
	mockQRGenerator := new(mocks.MockQRGenerator)
	mockUserRepo := new(mocks.MockUserRepo)
	mockMinioClient := new(mocks.MockMinioClient)
	mockLogger := new(mocks.MockLogger)

	uc := NewQRUseCase(mockQRGenerator, mockUserRepo, mockMinioClient, mockLogger)

	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"
	name := "Test User"

	user := &entity.User{
		ID:       userID,
		Email:    &email,
		FullName: name,
	}

	now := time.Now()
	qrData := &qr.QRData{
		UserID:    userID,
		Email:     email,
		FullName:  name,
		IssuedAt:  now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	qrImageBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockQRGenerator.On("GenerateQRCode", userID, email, name).Return(qrData, qrImageBase64, nil)
	mockMinioClient.On("UploadFile", ctx, mock.Anything, mock.Anything, mock.Anything, "image/png").Return(nil)
	mockUserRepo.On("UpdateQR", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID && u.QRIssuedAt != nil && u.QRExpiresAt != nil
	})).Return(user, nil)

	result, image, err := uc.RefreshQR(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, qrImageBase64, image)
	mockUserRepo.AssertExpectations(t)
	mockQRGenerator.AssertExpectations(t)
	mockMinioClient.AssertExpectations(t)
}

func TestQRUseCase_GetUserFromQR_Success(t *testing.T) {
	mockQRGenerator := new(mocks.MockQRGenerator)
	mockUserRepo := new(mocks.MockUserRepo)
	mockMinioClient := new(mocks.MockMinioClient)
	mockLogger := new(mocks.MockLogger)

	uc := NewQRUseCase(mockQRGenerator, mockUserRepo, mockMinioClient, mockLogger)

	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"

	qrData := &qr.QRData{
		UserID:   userID,
		Email:    email,
		FullName: "Test User",
	}

	user := &entity.User{
		ID:       userID,
		FullName: "Test User",
		Email:    &email,
	}

	qrDataJSON := `{"user_id":"` + userID.String() + `","email":"` + email + `"}`

	mockQRGenerator.On("ValidateQRCode", qrDataJSON).Return(qrData, nil)
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	result, err := uc.GetUserFromQR(ctx, qrDataJSON)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	mockQRGenerator.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestQRUseCase_GetUserFromQR_InvalidQR(t *testing.T) {
	mockQRGenerator := new(mocks.MockQRGenerator)
	mockUserRepo := new(mocks.MockUserRepo)
	mockMinioClient := new(mocks.MockMinioClient)
	mockLogger := new(mocks.MockLogger)

	uc := NewQRUseCase(mockQRGenerator, mockUserRepo, mockMinioClient, mockLogger)

	ctx := context.Background()
	qrDataJSON := "invalid"

	mockQRGenerator.On("ValidateQRCode", qrDataJSON).Return(nil, errors.New("invalid QR"))

	result, err := uc.GetUserFromQR(ctx, qrDataJSON)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockQRGenerator.AssertExpectations(t)
}
