package error

import "errors"

var (
	ErrHardwareNotAvailable    = errors.New("hardware not available")
	ErrArduinoConnectionFailed = errors.New("arduino connection failed")
	ErrArduinoCommandFailed    = errors.New("arduino command failed")
	ErrDisplayConnectionFailed = errors.New("display connection failed")
	ErrDisplayCommandFailed    = errors.New("display command failed")
	ErrCameraConnectionFailed  = errors.New("camera connection failed")
	ErrCameraScanFailed        = errors.New("camera scan failed")
	ErrCameraNotInMockMode     = errors.New("camera not in mock mode")
	ErrCameraChannelFull       = errors.New("camera result channel full")
)
