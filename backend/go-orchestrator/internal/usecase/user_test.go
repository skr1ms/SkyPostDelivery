package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase/mocks"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/jwt"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUseCase_VerifyPhoneCode_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	phone := "79991234567"
	code := "1234"
	email := "test@example.com"
	userID := uuid.New()
	expiresAt := time.Now().Add(10 * time.Minute)

	mockUserRepo.On("GetByPhone", ctx, phone).Return(&entity.User{
		ID:               userID,
		Email:            &email,
		PhoneNumber:      &phone,
		VerificationCode: &code,
		CodeExpiresAt:    &expiresAt,
	}, nil)
	mockUserRepo.On("VerifyPhone", ctx, userID).Return(&entity.User{
		ID:       userID,
		Email:    &email,
		FullName: "Test",
	}, nil)
	mockQRWebAPI.On("GenerateQR", ctx, userID, email, "Test").Return("qr-code", nil)

	user, qrCode, err := uc.VerifyPhoneCode(ctx, phone, code)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "qr-code", qrCode)
	mockUserRepo.AssertExpectations(t)
	mockQRWebAPI.AssertExpectations(t)
}

func TestUserUseCase_VerifyPhoneCode_InvalidCode(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	phone := "79991234567"
	storedCode := "1234"
	expiresAt := time.Now().Add(10 * time.Minute)

	mockUserRepo.On("GetByPhone", ctx, phone).Return(&entity.User{
		VerificationCode: &storedCode,
		CodeExpiresAt:    &expiresAt,
	}, nil)

	user, qrCode, err := uc.VerifyPhoneCode(ctx, phone, "9999")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, qrCode)
	assert.Contains(t, err.Error(), "invalid")
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_VerifyPhoneCode_ExpiredCode(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	phone := "79991234567"
	code := "1234"
	expiresAt := time.Now().Add(-1 * time.Minute)

	mockUserRepo.On("GetByPhone", ctx, phone).Return(&entity.User{
		VerificationCode: &code,
		CodeExpiresAt:    &expiresAt,
	}, nil)

	user, qrCode, err := uc.VerifyPhoneCode(ctx, phone, code)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, qrCode)
	assert.Contains(t, err.Error(), "expired")
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_LoginByPhone_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	phone := "79991234567"
	userID := uuid.New()

	user := &entity.User{
		ID:            userID,
		PhoneVerified: true,
		PhoneNumber:   &phone,
	}
	mockUserRepo.On("GetByPhone", ctx, phone).Return(user, nil)
	mockUserRepo.On("UpdateVerificationCode", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID && u.VerificationCode != nil && u.CodeExpiresAt != nil
	})).Return(user, nil)
	mockSMSWebAPI.On("SendSMS", ctx, phone, mock.MatchedBy(func(code string) bool {
		return len(code) == 4
	})).Return(nil)

	err := uc.LoginByPhone(ctx, phone)

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockSMSWebAPI.AssertExpectations(t)
}

func TestUserUseCase_LoginByPhone_NotVerified(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	phone := "79991234567"

	mockUserRepo.On("GetByPhone", ctx, phone).Return(&entity.User{
		PhoneVerified: false,
	}, nil)

	err := uc.LoginByPhone(ctx, phone)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not verified")
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_LoginByCredentials_ByEmail_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	userID := uuid.New()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashStr := string(hashedPassword)

	mockUserRepo.On("GetByEmail", ctx, email).Return(&entity.User{
		ID:            userID,
		Email:         &email,
		FullName:      "Test User",
		PassHash:      &hashStr,
		PhoneVerified: true,
	}, nil)
	mockQRWebAPI.On("GenerateQR", ctx, userID, email, "Test User").Return("qr-code", nil)

	user, qrCode, err := uc.LoginByCredentials(ctx, email, password)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "qr-code", qrCode)
	mockUserRepo.AssertExpectations(t)
	mockQRWebAPI.AssertExpectations(t)
}

func TestUserUseCase_LoginByCredentials_InvalidPassword(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	email := "test@example.com"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	hashStr := string(hashedPassword)

	mockUserRepo.On("GetByEmail", ctx, email).Return(&entity.User{
		Email:         &email,
		PassHash:      &hashStr,
		PhoneVerified: true,
	}, nil)

	user, qrCode, err := uc.LoginByCredentials(ctx, email, "wrong")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, qrCode)
	assert.Contains(t, err.Error(), "invalid credentials")
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_LoginByCredentials_PhoneNotVerified(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	phone := "79991234567"
	userID := uuid.New()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashStr := string(hashedPassword)

	user := &entity.User{
		ID:            userID,
		Email:         &email,
		PhoneNumber:   &phone,
		FullName:      "Test User",
		PassHash:      &hashStr,
		PhoneVerified: false,
	}
	mockUserRepo.On("GetByEmail", ctx, email).Return(user, nil)
	mockUserRepo.On("UpdateVerificationCode", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID && u.VerificationCode != nil && u.CodeExpiresAt != nil
	})).Return(user, nil)
	mockSMSWebAPI.On("SendSMS", ctx, phone, mock.MatchedBy(func(code string) bool {
		return len(code) == 4
	})).Return(nil)

	user, qrCode, err := uc.LoginByCredentials(ctx, email, password)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, qrCode)
	assert.ErrorIs(t, err, entityError.ErrPhoneNotVerified)
	mockUserRepo.AssertExpectations(t)
	mockSMSWebAPI.AssertExpectations(t)
}

func TestUserUseCase_LoginByCredentials_UserNotFound(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	login := "nonexistent@example.com"

	mockUserRepo.On("GetByEmail", ctx, login).Return(nil, errors.New("not found"))

	user, qrCode, err := uc.LoginByCredentials(ctx, login, "password")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, qrCode)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_RequestPasswordReset_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	phone := "79991234567"
	userID := uuid.New()

	user := &entity.User{
		ID:            userID,
		PhoneVerified: true,
		PhoneNumber:   &phone,
	}
	mockUserRepo.On("GetByPhone", ctx, phone).Return(user, nil)
	mockUserRepo.On("UpdateVerificationCode", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID && u.VerificationCode != nil && u.CodeExpiresAt != nil
	})).Return(user, nil)
	mockSMSWebAPI.On("SendSMS", ctx, phone, mock.MatchedBy(func(code string) bool {
		return len(code) == 4
	})).Return(nil)

	err := uc.RequestPasswordReset(ctx, phone)

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockSMSWebAPI.AssertExpectations(t)
}

func TestUserUseCase_RequestPasswordReset_UserNotFound(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	phone := "79991234567"

	mockUserRepo.On("GetByPhone", ctx, phone).Return(nil, errors.New("not found"))

	err := uc.RequestPasswordReset(ctx, phone)

	assert.Error(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_ResetPassword_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	phone := "79991234567"
	code := "1234"
	newPassword := "newpassword123"
	userID := uuid.New()
	expiresAt := time.Now().Add(10 * time.Minute)

	user := &entity.User{
		ID:               userID,
		VerificationCode: &code,
		CodeExpiresAt:    &expiresAt,
		PhoneNumber:      &phone,
	}
	mockUserRepo.On("GetByPhone", ctx, phone).Return(user, nil)
	mockValidator.On("ValidateVar", newPassword, "strong_password").Return(nil)
	mockUserRepo.On("UpdatePassword", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID && u.PassHash != nil
	})).Return(user, nil)
	mockUserRepo.On("UpdateVerificationCode", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID && u.VerificationCode != nil && *u.VerificationCode == ""
	})).Return(user, nil)

	err := uc.ResetPassword(ctx, phone, code, newPassword)

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_ResetPassword_InvalidCode(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	phone := "79991234567"
	storedCode := "1234"
	expiresAt := time.Now().Add(10 * time.Minute)

	mockUserRepo.On("GetByPhone", ctx, phone).Return(&entity.User{
		VerificationCode: &storedCode,
		CodeExpiresAt:    &expiresAt,
	}, nil)

	err := uc.ResetPassword(ctx, phone, "9999", "newpassword")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_GetUser_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"

	mockUserRepo.On("GetByID", ctx, userID).Return(&entity.User{
		ID:       userID,
		Email:    &email,
		FullName: "Test User",
	}, nil)

	user, err := uc.GetUser(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_GetUserByEmail_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	email := "test@example.com"
	userID := uuid.New()

	mockUserRepo.On("GetByEmail", ctx, email).Return(&entity.User{
		ID:       userID,
		Email:    &email,
		FullName: "Test User",
	}, nil)

	user, err := uc.GetUserByEmail(ctx, email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, *user.Email)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_GenerateCode(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	code := uc.generateCode()

	assert.Len(t, code, 4)
	assert.Regexp(t, `^\d{4}$`, code)
}

func TestUserUseCase_GetUserByID_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"
	phone := "79991234567"
	fullName := "Test User"
	qrExpiresAt := time.Now().Add(24 * time.Hour)

	expectedUser := &entity.User{
		ID:          userID,
		FullName:    "Test User",
		Email:       &email,
		PhoneNumber: &phone,
		Role:        "client",
		QRExpiresAt: &qrExpiresAt,
	}

	mockUserRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)
	mockQRWebAPI.On("GenerateQR", ctx, userID, email, fullName).Return("base64qrcode", nil)

	user, qrCode, err := uc.GetUserByID(ctx, userID.String())

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	assert.NotEmpty(t, qrCode)
	mockUserRepo.AssertExpectations(t)
	mockQRWebAPI.AssertExpectations(t)
}

func TestUserUseCase_GetUserByID_InvalidUUID(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	invalidUUID := "not-a-valid-uuid"

	user, qrCode, err := uc.GetUserByID(ctx, invalidUUID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, qrCode)
	assert.Contains(t, err.Error(), "ParseUserID")
}

func TestUserUseCase_GetUserByID_NotFound(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	userID := uuid.New()

	mockUserRepo.On("GetByID", ctx, userID).Return(nil, errors.New("user not found"))

	user, qrCode, err := uc.GetUserByID(ctx, userID.String())

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, qrCode)
	assert.Contains(t, err.Error(), "user not found")
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_Register_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	fullName := "Test User"
	email := "test@example.com"
	phone := "+79991234567"
	password := "password123"

	userID := uuid.New()
	user := &entity.User{
		ID:       userID,
		FullName: fullName,
		Email:    &email,
	}

	normalizedPhone := validator.NormalizeRussianPhone(phone)
	mockValidator.On("ValidateVar", normalizedPhone, "russian_phone").Return(nil)
	mockValidator.On("ValidateVar", email, "custom_email").Return(nil)
	mockValidator.On("ValidateVar", password, "strong_password").Return(nil)
	mockUserRepo.On("GetByEmail", ctx, email).Return(nil, entityError.ErrUserNotFoundByEmail)
	mockUserRepo.On("GetByPhone", ctx, normalizedPhone).Return(nil, entityError.ErrUserNotFoundByPhone)
	mockUserRepo.On("Create", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.FullName == fullName && *u.Email == email && *u.PhoneNumber == normalizedPhone && u.Role == "client"
	})).Return(user, nil)
	mockUserRepo.On("UpdateVerificationCode", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID
	})).Return(user, nil)
	mockSMSWebAPI.On("SendSMS", ctx, normalizedPhone, mock.Anything).Return(nil)

	passwordStr := password
	userEntity := &entity.User{
		FullName:    fullName,
		Email:       &email,
		PhoneNumber: &phone,
		PassHash:    &passwordStr,
	}
	result, err := uc.Register(ctx, userEntity)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	mockUserRepo.AssertExpectations(t)
	mockSMSWebAPI.AssertExpectations(t)
}

func TestUserUseCase_Register_EmailExists(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	email := "existing@example.com"

	existingUser := &entity.User{
		ID:    uuid.New(),
		Email: &email,
	}

	phone := "+79991234567"
	password := "password"
	normalizedPhone := validator.NormalizeRussianPhone(phone)
	mockValidator.On("ValidateVar", normalizedPhone, "russian_phone").Return(nil)
	mockValidator.On("ValidateVar", email, "custom_email").Return(nil)
	mockValidator.On("ValidateVar", password, "strong_password").Return(nil)
	mockUserRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)

	passwordStr := password
	userEntity := &entity.User{
		FullName:    "Test",
		Email:       &email,
		PhoneNumber: &phone,
		PassHash:    &passwordStr,
	}
	result, err := uc.Register(ctx, userEntity)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email already exists")
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_Register_PhoneExists(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepo)
	mockSMSWebAPI := new(mocks.MockSMSAeroWebAPI)
	mockQRWebAPI := new(mocks.MockQRWebAPI)
	mockValidator := new(mocks.MockValidator)
	jwtService := jwt.NewJWTService("test", "test", time.Hour, time.Hour)
	mockLogger := new(mocks.MockLogger)

	uc := NewUserUseCase(mockUserRepo, mockSMSWebAPI, mockQRWebAPI, jwtService, mockValidator, mockLogger)

	ctx := context.Background()
	email := "test@example.com"
	phone := "+79991234567"

	existingUser := &entity.User{
		ID: uuid.New(),
	}

	normalizedPhone := validator.NormalizeRussianPhone(phone)
	password := "password"
	mockValidator.On("ValidateVar", normalizedPhone, "russian_phone").Return(nil)
	mockValidator.On("ValidateVar", email, "custom_email").Return(nil)
	mockValidator.On("ValidateVar", password, "strong_password").Return(nil)
	mockUserRepo.On("GetByEmail", ctx, email).Return(nil, entityError.ErrUserNotFoundByEmail)
	mockUserRepo.On("GetByPhone", ctx, normalizedPhone).Return(existingUser, nil)

	passwordStr := password
	userEntity := &entity.User{
		FullName:    "Test",
		Email:       &email,
		PhoneNumber: &phone,
		PassHash:    &passwordStr,
	}
	result, err := uc.Register(ctx, userEntity)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "phone already exists")
	mockUserRepo.AssertExpectations(t)
}
