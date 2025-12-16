package error

import "errors"

var (
	ErrQRGenerateFailed = errors.New("failed to generate QR code")
)
