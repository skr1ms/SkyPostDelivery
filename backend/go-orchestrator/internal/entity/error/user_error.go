package error

import "errors"

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrUserNotFoundByPhone     = errors.New("user with this phone not found")
	ErrUserNotFoundByEmail     = errors.New("user not found by email")
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrUserEmailAlreadyExists  = errors.New("user with this email already exists")
	ErrUserPhoneAlreadyExists  = errors.New("user with this phone already exists")
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrVerificationCodeExpired = errors.New("verification code expired")
	ErrPhoneNotVerified        = errors.New("phone not verified")
	ErrPasswordNotSet          = errors.New("password not set for this user")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUserEmailMismatch       = errors.New("user email mismatch")
)
