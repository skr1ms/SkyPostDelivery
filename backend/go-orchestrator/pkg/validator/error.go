package validator

import "errors"

var (
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidPhone     = errors.New("invalid phone number")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrPasswordTooShort = errors.New("password is too short")
	ErrPasswordTooLong  = errors.New("password is too long")
)
