package error

import "errors"

var (
	ErrConfigValidationFailed = errors.New("config validation failed")
	ErrConfigMissingRequired  = errors.New("missing required config field")
)
