package error

import "errors"

var (
	ErrSenderInitFailed      = errors.New("failed to initialize push sender")
	ErrSenderMessagingFailed = errors.New("failed to get messaging client")
	ErrSenderSendFailed      = errors.New("failed to send push notification")
)
