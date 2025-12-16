package error

import "errors"

var (
	ErrSMSServiceUnavailable = errors.New("SMS service is temporarily unavailable")
	ErrSMSInvalidPhone       = errors.New("invalid phone number")
	ErrSMSRateLimitExceeded  = errors.New("SMS rate limit exceeded")
	ErrSMSInsufficientFunds  = errors.New("insufficient SMS balance")
	ErrSMSSendFailed         = errors.New("failed to send SMS")
)
