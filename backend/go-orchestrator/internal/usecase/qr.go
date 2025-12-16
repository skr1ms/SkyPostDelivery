package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	entityError "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity/error"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/repo"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/minio"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/qr"
)

type QRUseCase struct {
	qrGenerator qr.QRGeneratorContract
	userRepo    repo.UserRepo
	minioClient minio.MinioClient
	logger      logger.Interface
}

func NewQRUseCase(qrGenerator qr.QRGeneratorContract, userRepo repo.UserRepo, minioClient minio.MinioClient, logger logger.Interface) *QRUseCase {
	return &QRUseCase{
		qrGenerator: qrGenerator,
		userRepo:    userRepo,
		minioClient: minioClient,
		logger:      logger,
	}
}

type QRInfo struct {
	UserID    uuid.UUID
	Email     string
	Name      string
	IssuedAt  int64
	ExpiresAt int64
}

func (uc *QRUseCase) GenerateQR(ctx context.Context, user *entity.User) (*QRInfo, string, error) {
	qrData, qrImageBase64, err := uc.qrGenerator.GenerateQRCode(user.ID, user.GetEmail(), user.FullName)
	if err != nil {
		return nil, "", fmt.Errorf("QRUseCase - GenerateQR - GenerateQRCode: %w", err)
	}

	imageData, err := base64.StdEncoding.DecodeString(qrImageBase64)
	if err != nil {
		return nil, "", fmt.Errorf("QRUseCase - GenerateQR - DecodeString: %w", err)
	}

	objectName := fmt.Sprintf("%s.png", user.ID.String())
	reader := bytes.NewReader(imageData)

	if err := uc.minioClient.UploadFile(ctx, objectName, reader, int64(len(imageData)), "image/png"); err != nil {
		uc.logger.Warn("QRUseCase - GenerateQR - UploadFile", err, map[string]any{
			"userID":     user.ID,
			"objectName": objectName,
		})
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
		return nil, fmt.Errorf("QRUseCase - ValidateQR - ValidateQRCode: %w", err)
	}

	user, err := uc.userRepo.GetByID(ctx, qrData.UserID)
	if err != nil {
		return nil, err
	}

	if user.GetEmail() != qrData.Email {
		return nil, entityError.ErrQRUserMismatch
	}

	return user, nil
}

func (uc *QRUseCase) refreshQRInternal(ctx context.Context, user *entity.User) (*QRInfo, string, error) {
	qrInfo, qrImageBase64, err := uc.GenerateQR(ctx, user)
	if err != nil {
		return nil, "", err
	}

	issuedAt := time.Unix(qrInfo.IssuedAt, 0)
	expiresAt := time.Unix(qrInfo.ExpiresAt, 0)

	user.QRIssuedAt = &issuedAt
	user.QRExpiresAt = &expiresAt

	_, err = uc.userRepo.UpdateQR(ctx, user)
	if err != nil {
		uc.logger.Warn("QRUseCase - RefreshQR - UpdateQR", err, map[string]any{
			"userID": user.ID,
		})
	}

	return qrInfo, qrImageBase64, nil
}

func (uc *QRUseCase) RefreshQR(ctx context.Context, userID uuid.UUID) (*QRInfo, string, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}

	return uc.refreshQRInternal(ctx, user)
}

func (uc *QRUseCase) GetUserFromQR(ctx context.Context, qrDataJSON string) (*entity.User, error) {
	user, err := uc.ValidateQR(ctx, qrDataJSON)
	if err != nil {
		return nil, fmt.Errorf("QRUseCase - GetUserFromQR - ValidateQR: %w", err)
	}

	return user, nil
}

func (uc *QRUseCase) GetOrRefreshQR(ctx context.Context, userID uuid.UUID) (*QRInfo, string, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}

	needsRefresh := user.QRExpiresAt == nil || time.Now().After(*user.QRExpiresAt)
	if needsRefresh {
		return uc.refreshQRInternal(ctx, user)
	}

	qrCodeBase64, err := uc.getQRFromStorage(ctx, userID)
	if err != nil {
		uc.logger.Warn("QRUseCase - GetOrRefreshQR - GetFromStorage", err, map[string]any{
			"userID": userID,
		})
		return uc.refreshQRInternal(ctx, user)
	}

	if user.QRIssuedAt == nil || user.QRExpiresAt == nil {
		return uc.refreshQRInternal(ctx, user)
	}

	qrInfo := &QRInfo{
		UserID:    user.ID,
		Email:     user.GetEmail(),
		Name:      user.FullName,
		IssuedAt:  user.QRIssuedAt.Unix(),
		ExpiresAt: user.QRExpiresAt.Unix(),
	}

	return qrInfo, qrCodeBase64, nil
}

func (uc *QRUseCase) getQRFromStorage(ctx context.Context, userID uuid.UUID) (string, error) {
	objectName := fmt.Sprintf("%s.png", userID.String())

	reader, err := uc.minioClient.GetFile(ctx, objectName)
	if err != nil {
		return "", fmt.Errorf("QRUseCase - getQRFromStorage: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	imageData, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("QRUseCase - getQRFromStorage - ReadFile: %w", err)
	}

	qrCodeBase64 := base64.StdEncoding.EncodeToString(imageData)
	return qrCodeBase64, nil
}
