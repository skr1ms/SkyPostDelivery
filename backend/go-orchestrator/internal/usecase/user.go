package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo"
	webapiError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo/webapi/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/jwt"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/validator"
	"golang.org/x/crypto/bcrypt"
)

type UserUseCase struct {
	userRepo   repo.UserRepo
	smsWebAPI  repo.SMSAeroWebAPI
	qrWebAPI   repo.QRWebAPI
	jwtService *jwt.JWTService
	validator  validator.Validator
	logger     logger.Interface
}

func NewUserUseCase(
	userRepo repo.UserRepo,
	smsWebAPI repo.SMSAeroWebAPI,
	qrWebAPI repo.QRWebAPI,
	jwtService *jwt.JWTService,
	validator validator.Validator,
	logger logger.Interface,
) *UserUseCase {
	return &UserUseCase{
		userRepo:   userRepo,
		smsWebAPI:  smsWebAPI,
		qrWebAPI:   qrWebAPI,
		jwtService: jwtService,
		validator:  validator,
		logger:     logger,
	}
}

func (uc *UserUseCase) generateCode() string {
	return fmt.Sprintf("%04d", rand.Intn(10000))
}

func (uc *UserUseCase) handleSMSError(err error, operation string) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, webapiError.ErrSMSRateLimitExceeded):
		return webapiError.ErrSMSRateLimitExceeded
	case errors.Is(err, webapiError.ErrSMSInvalidPhone):
		return webapiError.ErrSMSInvalidPhone
	case errors.Is(err, webapiError.ErrSMSInsufficientFunds):
		uc.logger.Error("SMS balance insufficient!", err, map[string]any{"operation": operation})
		return webapiError.ErrSMSServiceUnavailable
	case errors.Is(err, webapiError.ErrSMSServiceUnavailable):
		uc.logger.Warn("SMS service unavailable", err, map[string]any{"operation": operation})
		return webapiError.ErrSMSServiceUnavailable
	default:
		return fmt.Errorf("UserUseCase - %s - SendSMS: %w", operation, err)
	}
}

func (uc *UserUseCase) VerifyPhoneCode(ctx context.Context, phone, code string) (*entity.User, string, error) {
	user, err := uc.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, "", err
	}

	if !user.IsCodeValid(code) {
		return nil, "", entityError.ErrInvalidVerificationCode
	}

	if user.IsCodeExpired() {
		return nil, "", entityError.ErrVerificationCodeExpired
	}

	user, err = uc.userRepo.VerifyPhone(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("UserUseCase - VerifyPhoneCode - VerifyPhone: %w", err)
	}

	qrCode, err := uc.qrWebAPI.GenerateQR(ctx, user.ID, user.GetEmail(), user.FullName)
	if err != nil {
		return nil, "", fmt.Errorf("UserUseCase - VerifyPhoneCode - GenerateQR: %w", err)
	}

	return user, qrCode, nil
}

func (uc *UserUseCase) LoginByPhone(ctx context.Context, phone string) error {
	user, err := uc.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return err
	}

	if !user.PhoneVerified {
		return entityError.ErrPhoneNotVerified
	}

	code := uc.generateCode()
	expiresAt := time.Now().Add(10 * time.Minute)
	codePtr := &code

	user.VerificationCode = codePtr
	user.CodeExpiresAt = &expiresAt

	_, err = uc.userRepo.UpdateVerificationCode(ctx, user)
	if err != nil {
		return fmt.Errorf("UserUseCase - LoginByPhone - UpdateCode: %w", err)
	}

	if err := uc.smsWebAPI.SendSMS(ctx, user.GetPhoneNumber(), code); err != nil {
		return uc.handleSMSError(err, "LoginByPhone")
	}

	return nil
}

func (uc *UserUseCase) LoginByCredentials(ctx context.Context, login, password string) (*entity.User, string, error) {
	user, err := uc.userRepo.GetByEmail(ctx, login)
	if err != nil {
		if errors.Is(err, entityError.ErrUserNotFoundByEmail) {
			return nil, "", entityError.ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("UserUseCase - LoginByCredentials - GetByEmail: %w", err)
	}

	if user.PassHash == nil || *user.PassHash == "" {
		return nil, "", entityError.ErrPasswordNotSet
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PassHash), []byte(password)); err != nil {
		return nil, "", entityError.ErrInvalidCredentials
	}

	if !user.PhoneVerified {
		code := uc.generateCode()
		expiresAt := time.Now().Add(10 * time.Minute)
		codePtr := &code

		user.VerificationCode = codePtr
		user.CodeExpiresAt = &expiresAt

		_, err = uc.userRepo.UpdateVerificationCode(ctx, user)
		if err != nil {
			return nil, "", fmt.Errorf("UserUseCase - LoginByCredentials - UpdateCode: %w", err)
		}

		if user.PhoneNumber != nil {
			if err := uc.smsWebAPI.SendSMS(ctx, *user.PhoneNumber, code); err != nil {
				return nil, "", uc.handleSMSError(err, "LoginByCredentials")
			}
		}

		return nil, "", entityError.ErrPhoneNotVerified
	}

	qrCode, err := uc.qrWebAPI.GenerateQR(ctx, user.ID, user.GetEmail(), user.FullName)
	if err != nil {
		return nil, "", fmt.Errorf("UserUseCase - LoginByCredentials - GenerateQR: %w", err)
	}

	return user, qrCode, nil
}

func (uc *UserUseCase) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	phone := ""
	if user.PhoneNumber != nil {
		phone = *user.PhoneNumber
	}
	phone = validator.NormalizeRussianPhone(phone)
	err := uc.validator.ValidateVar(phone, "russian_phone")
	if err != nil {
		return nil, err
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	err = uc.validator.ValidateVar(email, "custom_email")
	if err != nil {
		return nil, err
	}

	password := ""
	if user.PassHash != nil {
		password = *user.PassHash
	}
	err = uc.validator.ValidateVar(password, "strong_password")
	if err != nil {
		return nil, err
	}

	existingUser, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, entityError.ErrUserNotFoundByEmail) {
		return nil, fmt.Errorf("UserUseCase - Register - CheckEmail: %w", err)
	}
	if existingUser != nil {
		return nil, entityError.ErrUserEmailAlreadyExists
	}

	existingPhone, err := uc.userRepo.GetByPhone(ctx, phone)
	if err != nil && !errors.Is(err, entityError.ErrUserNotFoundByPhone) {
		return nil, fmt.Errorf("UserUseCase - Register - CheckPhone: %w", err)
	}
	if existingPhone != nil {
		return nil, entityError.ErrUserPhoneAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - Register - HashPassword: %w", err)
	}

	hashedPasswordStr := string(hashedPassword)
	phonePtr := &phone
	emailPtr := &email
	user.PassHash = &hashedPasswordStr
	user.PhoneNumber = phonePtr
	user.Email = emailPtr
	user.Role = "client"

	createdUser, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - Register - Create: %w", err)
	}

	code := uc.generateCode()
	expiresAt := time.Now().Add(10 * time.Minute)
	codePtr := &code

	createdUser.VerificationCode = codePtr
	createdUser.CodeExpiresAt = &expiresAt

	_, err = uc.userRepo.UpdateVerificationCode(ctx, createdUser)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - Register - UpdateCode: %w", err)
	}

	if err := uc.smsWebAPI.SendSMS(ctx, phone, code); err != nil {
		return nil, uc.handleSMSError(err, "Register")
	}

	return createdUser, nil
}

func (uc *UserUseCase) RequestPasswordReset(ctx context.Context, phone string) error {
	user, err := uc.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return err
	}

	if !user.PhoneVerified {
		return entityError.ErrPhoneNotVerified
	}

	code := uc.generateCode()
	expiresAt := time.Now().Add(10 * time.Minute)
	codePtr := &code

	user.VerificationCode = codePtr
	user.CodeExpiresAt = &expiresAt

	_, err = uc.userRepo.UpdateVerificationCode(ctx, user)
	if err != nil {
		return fmt.Errorf("UserUseCase - RequestPasswordReset - UpdateCode: %w", err)
	}

	if err := uc.smsWebAPI.SendSMS(ctx, phone, code); err != nil {
		return uc.handleSMSError(err, "RequestPasswordReset")
	}

	return nil
}

func (uc *UserUseCase) ResetPassword(ctx context.Context, phone, code, newPassword string) error {
	user, err := uc.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return err
	}

	if !user.IsCodeValid(code) {
		return entityError.ErrInvalidVerificationCode
	}

	if user.IsCodeExpired() {
		return entityError.ErrVerificationCodeExpired
	}

	err = uc.validator.ValidateVar(newPassword, "strong_password")
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("UserUseCase - ResetPassword - HashPassword: %w", err)
	}

	hashedPasswordStr := string(hashedPassword)
	user.PassHash = &hashedPasswordStr

	_, err = uc.userRepo.UpdatePassword(ctx, user)
	if err != nil {
		return fmt.Errorf("UserUseCase - ResetPassword - UpdatePassword: %w", err)
	}

	emptyCode := ""
	now := time.Now()
	user.VerificationCode = &emptyCode
	user.CodeExpiresAt = &now

	_, err = uc.userRepo.UpdateVerificationCode(ctx, user)
	if err != nil {
		return fmt.Errorf("UserUseCase - ResetPassword - ClearCode: %w", err)
	}

	return nil
}

func (uc *UserUseCase) GetUser(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *UserUseCase) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *UserUseCase) ValidatePassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (uc *UserUseCase) GetUserByID(ctx context.Context, userIDStr string) (*entity.User, string, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, "", fmt.Errorf("UserUseCase - GetUserByID - ParseUserID: %w", err)
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}

	var qrCode string
	qrExpired := user.QRExpiresAt == nil || time.Now().After(*user.QRExpiresAt)

	if qrExpired {
		uc.logger.Info("QR expired or not set, regenerating", nil, map[string]any{"userID": user.ID, "qrExpiresAt": user.QRExpiresAt})

		qrImageBase64, err := uc.qrWebAPI.GenerateQR(ctx, user.ID, user.GetEmail(), user.FullName)
		if err != nil {
			uc.logger.Error("UserUseCase - GetUserByID - GenerateQR", err, map[string]any{"userID": user.ID})
			return user, "", nil
		}

		now := time.Now()
		expiresAt := now.Add(7 * 24 * time.Hour)

		user.QRIssuedAt = &now
		user.QRExpiresAt = &expiresAt

		updatedUser, err := uc.userRepo.UpdateQR(ctx, user)
		if err != nil {
			uc.logger.Warn("UserUseCase - GetUserByID - UpdateQR", err, map[string]any{"userID": user.ID})
		} else {
			user = updatedUser
		}

		qrCode = qrImageBase64
	} else {
		qrCode, err = uc.qrWebAPI.GenerateQR(ctx, user.ID, user.GetEmail(), user.FullName)
		if err != nil {
			uc.logger.Error("UserUseCase - GetUserByID - GenerateQR", err, map[string]any{"userID": user.ID})
			return user, "", nil
		}
	}

	return user, qrCode, nil
}
