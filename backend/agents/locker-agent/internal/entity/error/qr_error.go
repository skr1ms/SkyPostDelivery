package error

import "errors"

var (
	ErrQRValidationFailed    = errors.New("qr validation failed")
	ErrQRInvalidFormat       = errors.New("invalid qr data format")
	ErrQRScanFailed          = errors.New("qr scan failed")
	ErrQRConfirmPickupFailed = errors.New("failed to confirm pickup")
	ErrQRConfirmLoadedFailed = errors.New("failed to confirm loaded")
)
