package error

import "errors"

var (
	ErrQRGenerationFailed   = errors.New("failed to generate QR code")
	ErrQRValidationFailed   = errors.New("QR code validation failed")
	ErrQRUserMismatch       = errors.New("QR code user mismatch")
	ErrQRStorageUnavailable = errors.New("QR storage is temporarily unavailable")
)
