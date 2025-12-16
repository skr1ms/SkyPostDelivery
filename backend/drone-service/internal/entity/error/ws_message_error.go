package error

import "errors"

var (
	ErrInvalidPayload       = errors.New("invalid payload format")
	ErrMissingRequiredField = errors.New("missing required field")
	ErrInvalidValue         = errors.New("invalid value provided")
	ErrUnknownMessageType   = errors.New("unknown message type")
	ErrActionNotAllowed     = errors.New("action not allowed in current state")
)
