package qr

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/config"
	"github.com/stretchr/testify/assert"
)

func TestNewQRGenerator(t *testing.T) {
	cfg := &config.QR{HMACSecret: "test-secret"}
	generator := NewQRGenerator(cfg)

	assert.NotNil(t, generator)
	assert.Equal(t, cfg.HMACSecret, generator.hmacSecret)
}

func TestQRGenerator_GenerateQRCode_Success(t *testing.T) {
	generator := NewQRGenerator(&config.QR{HMACSecret: "test-secret"})
	userID := uuid.New()
	email := "test@example.com"
	fullName := "Test User"

	qrData, qrImageBase64, err := generator.GenerateQRCode(userID, email, fullName)

	assert.NoError(t, err)
	assert.NotNil(t, qrData)
	assert.NotEmpty(t, qrImageBase64)
	assert.Equal(t, userID, qrData.UserID)
	assert.Equal(t, email, qrData.Email)
	assert.Equal(t, fullName, qrData.FullName)
	assert.NotEmpty(t, qrData.Signature)
	assert.False(t, qrData.IssuedAt.IsZero())
	assert.False(t, qrData.ExpiresAt.IsZero())
	assert.True(t, qrData.ExpiresAt.After(qrData.IssuedAt))
}

func TestQRGenerator_ValidateQRCode_Success(t *testing.T) {
	generator := NewQRGenerator(&config.QR{HMACSecret: "test-secret"})
	userID := uuid.New()
	email := "test@example.com"
	fullName := "Test User"

	qrData, _, err := generator.GenerateQRCode(userID, email, fullName)
	assert.NoError(t, err)

	qrDataJSON, err := json.Marshal(qrData)
	assert.NoError(t, err)

	validatedData, err := generator.ValidateQRCode(string(qrDataJSON))

	assert.NoError(t, err)
	assert.NotNil(t, validatedData)
	assert.Equal(t, qrData.UserID, validatedData.UserID)
	assert.Equal(t, qrData.Email, validatedData.Email)
	assert.Equal(t, qrData.FullName, validatedData.FullName)
	assert.Equal(t, qrData.Signature, validatedData.Signature)
}

func TestQRGenerator_ValidateQRCode_ExpiredQR(t *testing.T) {
	generator := NewQRGenerator(&config.QR{HMACSecret: "test-secret"})
	userID := uuid.New()

	qrData := &QRData{
		UserID:    userID,
		Email:     "test@example.com",
		FullName:  "Test User",
		IssuedAt:  time.Now().Add(-25 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	signature, err := generator.generateSignature(qrData)
	assert.NoError(t, err)
	qrData.Signature = signature

	qrDataJSON, err := json.Marshal(qrData)
	assert.NoError(t, err)

	validatedData, err := generator.ValidateQRCode(string(qrDataJSON))

	assert.Error(t, err)
	assert.Nil(t, validatedData)
	assert.Contains(t, err.Error(), "expired")
}

func TestQRGenerator_ValidateQRCode_InvalidSignature(t *testing.T) {
	generator := NewQRGenerator(&config.QR{HMACSecret: "test-secret"})
	userID := uuid.New()

	qrData := &QRData{
		UserID:    userID,
		Email:     "test@example.com",
		FullName:  "Test User",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Signature: "invalid-signature",
	}

	qrDataJSON, err := json.Marshal(qrData)
	assert.NoError(t, err)

	validatedData, err := generator.ValidateQRCode(string(qrDataJSON))

	assert.Error(t, err)
	assert.Nil(t, validatedData)
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestQRGenerator_ValidateQRCode_TamperedData(t *testing.T) {
	generator := NewQRGenerator(&config.QR{HMACSecret: "test-secret"})
	userID := uuid.New()

	qrData, _, err := generator.GenerateQRCode(userID, "test@example.com", "Test User")
	assert.NoError(t, err)

	qrData.Email = "tampered@example.com"

	qrDataJSON, err := json.Marshal(qrData)
	assert.NoError(t, err)

	validatedData, err := generator.ValidateQRCode(string(qrDataJSON))

	assert.Error(t, err)
	assert.Nil(t, validatedData)
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestQRGenerator_ValidateQRCode_InvalidJSON(t *testing.T) {
	generator := NewQRGenerator(&config.QR{HMACSecret: "test-secret"})

	validatedData, err := generator.ValidateQRCode("{invalid json")

	assert.Error(t, err)
	assert.Nil(t, validatedData)
}

func TestQRGenerator_ValidateQRCode_DifferentSecret(t *testing.T) {
	generator1 := NewQRGenerator(&config.QR{HMACSecret: "secret1"})
	generator2 := NewQRGenerator(&config.QR{HMACSecret: "secret2"})

	userID := uuid.New()
	qrData, _, err := generator1.GenerateQRCode(userID, "test@example.com", "Test User")
	assert.NoError(t, err)

	qrDataJSON, err := json.Marshal(qrData)
	assert.NoError(t, err)

	validatedData, err := generator2.ValidateQRCode(string(qrDataJSON))

	assert.Error(t, err)
	assert.Nil(t, validatedData)
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestQRGenerator_RefreshQRCode(t *testing.T) {
	generator := NewQRGenerator(&config.QR{HMACSecret: "test-secret"})
	userID := uuid.New()
	email := "test@example.com"
	fullName := "Test User"

	qrData1, image1, err := generator.GenerateQRCode(userID, email, fullName)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	qrData2, image2, err := generator.RefreshQRCode(userID, email, fullName)
	assert.NoError(t, err)

	assert.Equal(t, userID, qrData2.UserID)
	assert.Equal(t, email, qrData2.Email)
	assert.Equal(t, fullName, qrData2.FullName)
	assert.NotEqual(t, qrData1.IssuedAt, qrData2.IssuedAt)
	assert.NotEqual(t, qrData1.Signature, qrData2.Signature)
	assert.NotEmpty(t, image1)
	assert.NotEmpty(t, image2)
}

func TestQRGenerator_generateSignature(t *testing.T) {
	generator := NewQRGenerator(&config.QR{HMACSecret: "test-secret"})
	userID := uuid.New()
	now := time.Now()

	qrData := &QRData{
		UserID:    userID,
		Email:     "test@example.com",
		FullName:  "Test User",
		IssuedAt:  now,
		ExpiresAt: now.Add(24 * time.Hour),
	}

	signature1, err := generator.generateSignature(qrData)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature1)

	signature2, err := generator.generateSignature(qrData)
	assert.NoError(t, err)
	assert.Equal(t, signature1, signature2)

	qrData.Email = "different@example.com"
	signature3, err := generator.generateSignature(qrData)
	assert.NoError(t, err)
	assert.NotEqual(t, signature1, signature3)
}
