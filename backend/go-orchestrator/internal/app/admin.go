package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/config"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	repo "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo/persistent"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo/webapi"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/qr"
	"golang.org/x/crypto/bcrypt"
)

type adminCredentials struct {
	FullName  string
	Email     string
	Phone     string
	Password  string
	CreatedAt string
}

func ensureAdminExists(pg *pgxpool.Pool, cfg *config.Config, qrGenerator *qr.QRGenerator, logger logger.Interface) error {
	admins := []adminCredentials{
		{
			FullName:  cfg.FirstAdmin.FullName,
			Email:     cfg.FirstAdmin.Email,
			Phone:     cfg.FirstAdmin.Phone,
			Password:  cfg.FirstAdmin.Password,
			CreatedAt: cfg.FirstAdmin.CreatedAt,
		},
		{
			FullName:  cfg.SecondAdmin.FullName,
			Email:     cfg.SecondAdmin.Email,
			Phone:     cfg.SecondAdmin.Phone,
			Password:  cfg.SecondAdmin.Password,
			CreatedAt: cfg.SecondAdmin.CreatedAt,
		},
	}

	for _, admin := range admins {
		if err := createAdminUser(pg, admin, qrGenerator, logger); err != nil {
			return err
		}
	}

	return nil
}

func createAdminUser(pg *pgxpool.Pool, admin adminCredentials, qrGenerator *qr.QRGenerator, logger logger.Interface) error {
	if admin.Email == "" || admin.Phone == "" || admin.Password == "" || admin.FullName == "" {
		logger.Info("Admin credentials not provided, skipping admin creation", nil, nil)
		return nil
	}

	ctx := context.Background()
	userRepo := repo.NewUserRepo(pg)

	existingUser, err := userRepo.GetByEmail(ctx, admin.Email)
	if err == nil && existingUser != nil {
		logger.Info("Admin user already exists, skipping creation", nil, nil)
		return nil
	}

	existingUserByPhone, err := userRepo.GetByPhone(ctx, admin.Phone)
	if err == nil && existingUserByPhone != nil {
		logger.Info("User with admin phone already exists, skipping creation", nil, nil)
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("createAdminUser - GenerateFromPassword: %w", err)
	}

	hashedPasswordStr := string(hashedPassword)

	var createdAt time.Time
	if admin.CreatedAt != "" {
		dateStr := admin.CreatedAt
		for i, r := range dateStr {
			if r == ' ' {
				dateStr = dateStr[:i]
				break
			}
		}

		parsedTime, err := time.Parse("02.01.2006", dateStr)
		if err != nil {
			logger.Warn("Failed to parse admin created_at date, using current time", err, map[string]any{
				"createdAt": admin.CreatedAt,
			})
			createdAt = time.Now()
		} else {
			createdAt = parsedTime
		}
	} else {
		createdAt = time.Now()
	}

	user := &entity.User{
		FullName:    admin.FullName,
		Email:       &admin.Email,
		PhoneNumber: &admin.Phone,
		PassHash:    &hashedPasswordStr,
		Role:        "admin",
	}

	if admin.CreatedAt != "" {
		_, err = userRepo.CreateWithCustomDate(ctx, user, createdAt)
	} else {
		_, err = userRepo.Create(ctx, user)
	}
	if err != nil {
		return fmt.Errorf("createAdminUser - Create: %w", err)
	}

	user, err = userRepo.GetByEmail(ctx, admin.Email)
	if err != nil {
		return fmt.Errorf("createAdminUser - GetByEmail: %w", err)
	}

	_, err = userRepo.VerifyPhone(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("createAdminUser - VerifyPhone: %w", err)
	}

	qrAdapter := webapi.NewQRAdapter(qrGenerator)

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	_, err = qrAdapter.GenerateQR(ctx, user.ID, email, user.FullName)
	if err != nil {
		logger.Warn("Failed to generate QR code for admin user", err, map[string]any{
			"userID": user.ID,
		})
	} else {
		now := time.Now()
		expiresAt := now.Add(7 * 24 * time.Hour)

		user.QRIssuedAt = &now
		user.QRExpiresAt = &expiresAt

		_, err = userRepo.UpdateQR(ctx, user)
		if err != nil {
			logger.Warn("Failed to update QR timestamps for admin user", err, map[string]any{
				"userID": user.ID,
			})
		} else {
			logger.Info("Admin QR code generated successfully", nil, map[string]any{
				"userID": user.ID,
			})
		}
	}

	logger.Info("Admin user created successfully", nil, map[string]any{
		"userID": user.ID,
		"email":  admin.Email,
	})
	return nil
}
