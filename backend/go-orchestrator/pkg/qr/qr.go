package qr

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

type QRGeneratorContract interface {
	GenerateQRCode(userID uuid.UUID, email, name string) (qrData *QRData, qrImageBase64 string, err error)
	ValidateQRCode(qrDataJSON string) (qrData *QRData, err error)
	RefreshQRCode(userID uuid.UUID, email, name string) (qrData *QRData, qrImageBase64 string, err error)
}

const TTL = 7 * 24 * time.Hour

type QRGenerator struct {
	hmacSecret string
}

func NewQRGenerator(hmacSecret string) *QRGenerator {
	return &QRGenerator{
		hmacSecret: hmacSecret,
	}
}

type QRData struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"name"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Signature string    `json:"signature"`
}

func (g *QRGenerator) GenerateQRCode(userID uuid.UUID, email, name string) (qrData *QRData, qrImageBase64 string, err error) {
	now := time.Now()
	expiresAt := now.Add(TTL)

	qrData = &QRData{
		UserID:    userID,
		Email:     email,
		FullName:  name,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
	}

	signature, err := g.generateSignature(qrData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate signature: %w", err)
	}

	qrData.Signature = signature

	jsonData, err := json.Marshal(qrData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal qr data: %w", err)
	}

	qr, err := qrcode.New(string(jsonData), qrcode.Medium)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create qr code: %w", err)
	}

	qr.DisableBorder = false
	img := qr.Image(256)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, "", fmt.Errorf("failed to encode png: %w", err)
	}

	qrImageBase64 = base64.StdEncoding.EncodeToString(buf.Bytes())

	return qrData, qrImageBase64, nil
}

func (g *QRGenerator) ValidateQRCode(qrDataJSON string) (*QRData, error) {
	var qrData QRData
	if err := json.Unmarshal([]byte(qrDataJSON), &qrData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal qr data: %w", err)
	}

	if time.Now().After(qrData.ExpiresAt) {
		return nil, fmt.Errorf("qr code expired")
	}

	originalSignature := qrData.Signature
	qrData.Signature = ""

	expectedSignature, err := g.generateSignature(&qrData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate expected signature: %w", err)
	}

	if !hmac.Equal([]byte(originalSignature), []byte(expectedSignature)) {
		return nil, fmt.Errorf("invalid signature")
	}

	qrData.Signature = originalSignature

	return &qrData, nil
}

func (g *QRGenerator) RefreshQRCode(userID uuid.UUID, email, name string) (qrData *QRData, qrImageBase64 string, err error) {
	return g.GenerateQRCode(userID, email, name)
}

func (g *QRGenerator) generateSignature(qrData *QRData) (string, error) {
	dataToSign := fmt.Sprintf("%s:%s:%s:%d:%d",
		qrData.UserID.String(),
		qrData.Email,
		qrData.FullName,
		qrData.IssuedAt.Unix(),
		qrData.ExpiresAt.Unix(),
	)

	h := hmac.New(sha256.New, []byte(g.hmacSecret))
	if _, err := h.Write([]byte(dataToSign)); err != nil {
		return "", err
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}
