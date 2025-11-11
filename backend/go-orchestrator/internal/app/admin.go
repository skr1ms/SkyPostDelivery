package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/hitech-ekb/config"
	"github.com/skr1ms/hitech-ekb/internal/usecase/repo"
	"github.com/skr1ms/hitech-ekb/internal/usecase/webapi"
	"github.com/skr1ms/hitech-ekb/pkg/qr"
	"golang.org/x/crypto/bcrypt"
)

type adminCredentials struct {
	FullName  string
	Email     string
	Phone     string
	Password  string
	CreatedAt string
}

func ensureAdminExists(pg *pgxpool.Pool, cfg *config.Config, qrGenerator *qr.QRGenerator) error {
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
		if err := createAdminUser(pg, admin, qrGenerator); err != nil {
			return err
		}
	}

	return nil
}

func createAdminUser(pg *pgxpool.Pool, admin adminCredentials, qrGenerator *qr.QRGenerator) error {
	if admin.Email == "" || admin.Phone == "" || admin.Password == "" || admin.FullName == "" {
		log.Println("Admin credentials not provided, skipping admin creation")
		return nil
	}

	ctx := context.Background()
	userRepo := repo.NewUserRepo(pg)

	existingUser, err := userRepo.GetByEmail(ctx, admin.Email)
	if err == nil && existingUser != nil {
		log.Println("Admin user already exists, skipping creation")
		return nil
	}

	existingUserByPhone, err := userRepo.GetByPhone(ctx, admin.Phone)
	if err == nil && existingUserByPhone != nil {
		log.Println("User with admin phone already exists, skipping creation")
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

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
			log.Printf("Warning: Failed to parse admin created_at date '%s', using current time: %v", admin.CreatedAt, err)
			createdAt = time.Now()
		} else {
			createdAt = parsedTime
		}
	} else {
		createdAt = time.Now()
	}

	if admin.CreatedAt != "" {
		_, err = userRepo.CreateWithCustomDate(
			ctx,
			admin.FullName,
			admin.Email,
			admin.Phone,
			string(hashedPassword),
			"admin",
			createdAt,
		)
	} else {
		_, err = userRepo.Create(
			ctx,
			admin.FullName,
			admin.Email,
			admin.Phone,
			string(hashedPassword),
			"admin",
		)
	}
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	_, err = userRepo.GetByEmail(ctx, admin.Email)
	if err != nil {
		return fmt.Errorf("failed to verify admin user: %w", err)
	}

	user, err := userRepo.GetByEmail(ctx, admin.Email)
	if err != nil {
		return fmt.Errorf("failed to get created admin user: %w", err)
	}

	_, err = userRepo.VerifyPhone(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to verify admin phone: %w", err)
	}

	qrAdapter := webapi.NewQRAdapter(qrGenerator)

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	_, err = qrAdapter.GenerateQR(ctx, user.ID, email, user.FullName)
	if err != nil {
		log.Printf("Warning: Failed to generate QR code for admin user: %v", err)
	} else {
		now := time.Now()
		expiresAt := now.Add(24 * time.Hour)

		_, err = userRepo.UpdateQR(ctx, user.ID, now, expiresAt)
		if err != nil {
			log.Printf("Warning: Failed to update QR timestamps for admin user: %v", err)
		} else {
			log.Println("Admin QR code generated successfully")
		}
	}

	log.Println("Admin user created successfully")
	return nil
}
