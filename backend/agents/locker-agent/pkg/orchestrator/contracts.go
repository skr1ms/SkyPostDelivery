package orchestrator

import (
	"context"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity"
)

type ClientInterface interface {
	ValidateQR(ctx context.Context, qrData, parcelAutomatID string) (*entity.QRValidationResponse, error)
	ConfirmPickup(ctx context.Context, cellIDs []uuid.UUID) error
	ConfirmLoaded(ctx context.Context, orderID, lockerCellID uuid.UUID) error
}
