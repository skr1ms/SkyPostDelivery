package error

import "errors"

var (
	ErrNotificationInvalidToken = errors.New("invalid notification token")
	ErrNotificationSendFailed   = errors.New("failed to send notification")
)
