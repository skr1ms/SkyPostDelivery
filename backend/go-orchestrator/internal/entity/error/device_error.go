package error

import "errors"

var (
	ErrDeviceNotFound     = errors.New("device not found")
	ErrDeviceCreateFailed = errors.New("failed to create device")
	ErrDeviceUpdateFailed = errors.New("failed to update device")
)
