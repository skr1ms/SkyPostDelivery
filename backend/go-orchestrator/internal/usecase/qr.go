package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/minio"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/qr"
)

type QRUseCase struct {
	qrGenerator qr.QRGeneratorContract
	userRepo    repo.UserRepo
	minioClient minio.MinioClient
}

func NewQRUseCase(qrGenerator qr.QRGeneratorContract, userRepo repo.UserRepo, minioClient minio.MinioClient) *QRUseCase {
	return &QRUseCase{
		qrGenerator: qrGenerator,
		userRepo:    userRepo,
		minioClient: minioClient,
	}
}

type QRInfo struct {
	UserID    uuid.UUID
	Email     string
	Name      string
	IssuedAt  int64
	ExpiresAt int64
}

func (uc *QRUseCase) GenerateQR(ctx context.Context, userID uuid.UUID, email, name string) (*QRInfo, string, error) {
	qrData, qrImageBase64, err := uc.qrGenerator.GenerateQRCode(userID, email, name)
	if err != nil {
		return nil, "", fmt.Errorf("qr usecase - GenerateQR - qrGenerator.GenerateQRCode: %w", err)
	}

	imageData, err := base64.StdEncoding.DecodeString(qrImageBase64)
	if err != nil {
		return nil, "", fmt.Errorf("qr usecase - GenerateQR - base64.Decode: %w", err)
	}

	objectName := fmt.Sprintf("%s.png", userID.String())
	reader := bytes.NewReader(imageData)

	if err := uc.minioClient.UploadFile(ctx, objectName, reader, int64(len(imageData)), "image/png"); err != nil {
		return nil, "", fmt.Errorf("qr usecase - GenerateQR - minioClient.UploadFile: %w", err)
	}

	qrInfo := &QRInfo{
		UserID:    qrData.UserID,
		Email:     qrData.Email,
		Name:      qrData.FullName,
		IssuedAt:  qrData.IssuedAt.Unix(),
		ExpiresAt: qrData.ExpiresAt.Unix(),
	}

	return qrInfo, qrImageBase64, nil
}

func (uc *QRUseCase) ValidateQR(ctx context.Context, qrDataJSON string) (*entity.User, error) {
	qrData, err := uc.qrGenerator.ValidateQRCode(qrDataJSON)
	if err != nil {
		return nil, fmt.Errorf("qr usecase - ValidateQR - qrGenerator.ValidateQRCode: %w", err)
	}

	user, err := uc.userRepo.GetByID(ctx, qrData.UserID)
	if err != nil {
		return nil, fmt.Errorf("qr usecase - ValidateQR - userRepo.GetByID: %w", err)
	}

	userEmail := ""
	if user.Email != nil {
		userEmail = *user.Email
	}

	if userEmail != qrData.Email {
		return nil, fmt.Errorf("qr usecase - ValidateQR - user email mismatch")
	}

	return user, nil
}

func (uc *QRUseCase) RefreshQR(ctx context.Context, userID uuid.UUID) (*QRInfo, string, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("qr usecase - RefreshQR - userRepo.GetByID: %w", err)
	}

	userEmail := ""
	if user.Email != nil {
		userEmail = *user.Email
	}

	qrInfo, qrImageBase64, err := uc.GenerateQR(ctx, user.ID, userEmail, user.FullName)
	if err != nil {
		return nil, "", err
	}

	issuedAt := time.Unix(qrInfo.IssuedAt, 0)
	expiresAt := time.Unix(qrInfo.ExpiresAt, 0)

	_, err = uc.userRepo.UpdateQR(ctx, userID, issuedAt, expiresAt)
	if err != nil {
		return nil, "", fmt.Errorf("qr usecase - RefreshQR - userRepo.UpdateQR: %w", err)
	}

	return qrInfo, qrImageBase64, nil
}

func (uc *QRUseCase) GetUserFromQR(ctx context.Context, qrDataJSON string) (*entity.User, error) {
	user, err := uc.ValidateQR(ctx, qrDataJSON)
	if err != nil {
		return nil, fmt.Errorf("qr usecase - GetUserFromQR - ValidateQR: %w", err)
	}

	return user, nil
}
