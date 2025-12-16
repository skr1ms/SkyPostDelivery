package error

import "errors"

var (
	ErrDeliveryNotFound      = errors.New("delivery not found")
	ErrDeliveryInvalidStatus = errors.New("invalid delivery status")
	ErrDeliveryCreateFailed  = errors.New("failed to create delivery")
	ErrDeliveryUpdateFailed  = errors.New("failed to update delivery")
	ErrDeliveryTaskNotFound  = errors.New("delivery task not found")
)
