package webapi

import (
	"context"

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
	return qrImageBase64, err
}
