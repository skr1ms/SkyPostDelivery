package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/jwt"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/validator"
	"golang.org/x/crypto/bcrypt"
)

type UserUseCase struct {
	userRepo   repo.UserRepo
	smsWebAPI  repo.SMSAeroWebAPI
	qrWebAPI   repo.QRWebAPI
	jwtService *jwt.JWTService
	validator  validator.Validator
}

func NewUserUseCase(
	userRepo repo.UserRepo,
	smsWebAPI repo.SMSAeroWebAPI,
	qrWebAPI repo.QRWebAPI,
	jwtService *jwt.JWTService,
	validator validator.Validator,
) *UserUseCase {
	return &UserUseCase{
		userRepo:   userRepo,
		smsWebAPI:  smsWebAPI,
		qrWebAPI:   qrWebAPI,
		jwtService: jwtService,
		validator:  validator,
	}
}

func (uc *UserUseCase) generateCode() string {
	return fmt.Sprintf("%04d", rand.Intn(10000))
}

func (uc *UserUseCase) VerifyPhoneCode(ctx context.Context, phone, code string) (*entity.User, string, error) {
	user, err := uc.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, "", fmt.Errorf("UserUseCase - VerifyPhoneCode - GetByPhone: %w", err)
	}

	if !user.IsCodeValid(code) {
		return nil, "", fmt.Errorf("invalid verification code")
	}

	if user.IsCodeExpired() {
		return nil, "", fmt.Errorf("verification code expired")
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
		return fmt.Errorf("user with this phone not found")
	}

	if !user.PhoneVerified {
		return fmt.Errorf("phone not verified")
	}

	code := uc.generateCode()
	expiresAt := time.Now().Add(10 * time.Minute)

	_, err = uc.userRepo.UpdateVerificationCode(ctx, user.ID, code, expiresAt)
	if err != nil {
		return fmt.Errorf("UserUseCase - LoginByPhone - UpdateVerificationCode: %w", err)
	}

	if err := uc.smsWebAPI.SendSMS(ctx, phone, code); err != nil {
		return fmt.Errorf("UserUseCase - LoginByPhone - SendSMS: %w", err)
	}

	return nil
}

func (uc *UserUseCase) LoginByCredentials(ctx context.Context, login, password string) (*entity.User, string, error) {
	user, err := uc.userRepo.GetByEmail(ctx, login)
	if err != nil {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	if user.PassHash == nil || *user.PassHash == "" {
		return nil, "", fmt.Errorf("password not set for this user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PassHash), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	if !user.PhoneVerified {
		code := uc.generateCode()
		expiresAt := time.Now().Add(10 * time.Minute)

		_, err = uc.userRepo.UpdateVerificationCode(ctx, user.ID, code, expiresAt)
		if err != nil {
			return nil, "", fmt.Errorf("UserUseCase - LoginByCredentials - UpdateVerificationCode: %w", err)
		}

		if user.PhoneNumber != nil {
			if err := uc.smsWebAPI.SendSMS(ctx, *user.PhoneNumber, code); err != nil {
				return nil, "", fmt.Errorf("UserUseCase - LoginByCredentials - SendSMS: %w", err)
			}
		}

		return nil, "", fmt.Errorf("PHONE_NOT_VERIFIED:%s", *user.PhoneNumber)
	}

	qrCode, err := uc.qrWebAPI.GenerateQR(ctx, user.ID, user.GetEmail(), user.FullName)
	if err != nil {
		return nil, "", fmt.Errorf("UserUseCase - LoginByCredentials - GenerateQR: %w", err)
	}

	return user, qrCode, nil
}

func (uc *UserUseCase) Register(ctx context.Context, fullName, email, phone, password string) (*entity.User, error) {
	phone = validator.NormalizeRussianPhone(phone)
	err := uc.validator.ValidateVar(phone, "russian_phone")
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - Register - ValidateVar: %w", err)
	}
	
	err = uc.validator.ValidateVar(email, "custom_email")
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - Register - ValidateVar: %w", err)
	}
	
	err = uc.validator.ValidateVar(password, "strong_password")
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - Register - ValidateVar: %w", err)
	}

	existingUser, _ := uc.userRepo.GetByEmail(ctx, email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	existingPhone, _ := uc.userRepo.GetByPhone(ctx, phone)
	if existingPhone != nil {
		return nil, fmt.Errorf("user with this phone already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - Register - bcrypt: %w", err)
	}

	user, err := uc.userRepo.Create(ctx, fullName, email, phone, string(hashedPassword), "client")
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - Register - Create: %w", err)
	}

	code := uc.generateCode()
	expiresAt := time.Now().Add(10 * time.Minute)

	_, err = uc.userRepo.UpdateVerificationCode(ctx, user.ID, code, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - Register - UpdateVerificationCode: %w", err)
	}

	err = uc.smsWebAPI.SendSMS(ctx, phone, code)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - Register - SendSMS: %w", err)
	}

	return user, nil
}

func (uc *UserUseCase) RequestPasswordReset(ctx context.Context, phone string) error {
	user, err := uc.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return fmt.Errorf("user with this phone not found")
	}

	if !user.PhoneVerified {
		return fmt.Errorf("phone not verified")
	}

	code := uc.generateCode()
	expiresAt := time.Now().Add(10 * time.Minute)

	_, err = uc.userRepo.UpdateVerificationCode(ctx, user.ID, code, expiresAt)
	if err != nil {
		return fmt.Errorf("UserUseCase - RequestPasswordReset - UpdateVerificationCode: %w", err)
	}

	if err := uc.smsWebAPI.SendSMS(ctx, phone, code); err != nil {
		return fmt.Errorf("UserUseCase - RequestPasswordReset - SendSMS: %w", err)
	}

	return nil
}

func (uc *UserUseCase) ResetPassword(ctx context.Context, phone, code, newPassword string) error {
	user, err := uc.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if !user.IsCodeValid(code) {
		return fmt.Errorf("invalid verification code")
	}

	if user.IsCodeExpired() {
		return fmt.Errorf("verification code expired")
	}

	err = uc.validator.ValidateVar(newPassword, "strong_password")
	if err != nil {
		return fmt.Errorf("UserUseCase - ResetPassword - ValidateVar: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("UserUseCase - ResetPassword - bcrypt: %w", err)
	}

	_, err = uc.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("UserUseCase - ResetPassword - UpdatePassword: %w", err)
	}

	_, err = uc.userRepo.UpdateVerificationCode(ctx, user.ID, "", time.Now())
	if err != nil {
		return fmt.Errorf("UserUseCase - ResetPassword - clear code: %w", err)
	}

	return nil
}

func (uc *UserUseCase) GetUser(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - GetUser - GetByID: %w", err)
	}
	return user, nil
}

func (uc *UserUseCase) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - GetUserByEmail - GetByEmail: %w", err)
	}
	return user, nil
}

func (uc *UserUseCase) ValidatePassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (uc *UserUseCase) GetUserByID(ctx context.Context, userIDStr string) (*entity.User, string, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, "", fmt.Errorf("UserUseCase - GetUserByID - uuid.Parse: %w", err)
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("UserUseCase - GetUserByID - GetByID: %w", err)
	}

	var qrCode string
	qrExpired := user.QRExpiresAt == nil || time.Now().After(*user.QRExpiresAt)

	if qrExpired {
		fmt.Printf("QR expired or not set for user %s. Regenerating QR. QRExpiresAt: %v", user.ID, user.QRExpiresAt)

		qrImageBase64, err := uc.qrWebAPI.GenerateQR(ctx, user.ID, user.GetEmail(), user.FullName)
		if err != nil {
			fmt.Printf("Failed to generate QR for user %s: %v", user.ID, err)
			return user, "", nil
		}

		now := time.Now()
		expiresAt := now.Add(7 * 24 * time.Hour)

		updatedUser, err := uc.userRepo.UpdateQR(ctx, user.ID, now, expiresAt)
		if err != nil {
			fmt.Printf("Failed to update QR timestamps for user %s: %v", user.ID, err)
		} else {
			user = updatedUser
		}

		qrCode = qrImageBase64
	} else {
		qrCode, err = uc.qrWebAPI.GenerateQR(ctx, user.ID, user.GetEmail(), user.FullName)
		if err != nil {
			fmt.Printf("Failed to generate QR for user %s: %v", user.ID, err)
			return user, "", nil
		}
	}

	return user, qrCode, nil
}
