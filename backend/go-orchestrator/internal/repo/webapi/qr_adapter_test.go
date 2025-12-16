package webapi

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/config"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/qr"
	"github.com/stretchr/testify/assert"
)

func TestQRAdapter_GenerateQR_Success(t *testing.T) {
	qrGen := qr.NewQRGenerator(&config.QR{HMACSecret: "test-secret-key"})
	adapter := NewQRAdapter(qrGen)

	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"
	name := "Test User"

	result, err := adapter.GenerateQR(ctx, userID, email, name)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestQRAdapter_GenerateQR_ValidatesOutput(t *testing.T) {
	qrGen := qr.NewQRGenerator(&config.QR{HMACSecret: "test-secret-key"})
	adapter := NewQRAdapter(qrGen)

	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"
	name := "Test User"

	qrBase64, err := adapter.GenerateQR(ctx, userID, email, name)

	assert.NoError(t, err)
	assert.NotEmpty(t, qrBase64)
	assert.Regexp(t, `^[A-Za-z0-9+/=]+$`, qrBase64)
}

func TestNewQRAdapter(t *testing.T) {
	qrGen := qr.NewQRGenerator(&config.QR{HMACSecret: "test-secret"})
	adapter := NewQRAdapter(qrGen)

	assert.NotNil(t, adapter)
	assert.NotNil(t, adapter.qrGenerator)
}
