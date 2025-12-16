package webapi

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/qr"
)

type QRAdapter struct {
	qrGenerator *qr.QRGenerator
}

func NewQRAdapter(qrGenerator *qr.QRGenerator) *QRAdapter {
	return &QRAdapter{
		qrGenerator: qrGenerator,
	}
}

func (a *QRAdapter) GenerateQR(ctx context.Context, userID uuid.UUID, email, name string) (string, error) {
	_, qrImageBase64, err := a.qrGenerator.GenerateQRCode(userID, email, name)
	if err != nil {
		return "", fmt.Errorf("QRAdapter - GenerateQR - GenerateQRCode: %w", err)
	}
	return qrImageBase64, nil
}
