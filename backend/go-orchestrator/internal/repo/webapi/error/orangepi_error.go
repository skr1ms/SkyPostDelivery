package error

import "errors"

var (
	ErrOrangePIEmptyIP            = errors.New("IP address is empty")
	ErrOrangePIServiceUnavailable = errors.New("OrangePI service is temporarily unavailable")
	ErrOrangePISendFailed         = errors.New("failed to send request to OrangePI")
	ErrOrangePIOpenCellFailed     = errors.New("failed to open OrangePI cell")
)
