package jwt

import "errors"

var (
	ErrTokenGenerationFailed  = errors.New("failed to generate token")
	ErrTokenValidationFailed  = errors.New("failed to validate token")
	ErrTokenInvalid           = errors.New("invalid token")
	ErrTokenExpired           = errors.New("token expired")
	ErrTokenInvalidType       = errors.New("invalid token type")
	ErrTokenUnexpectedSigning = errors.New("unexpected signing method")
	ErrRefreshTokenInvalid    = errors.New("invalid refresh token")
	ErrAccessTokenInvalid     = errors.New("invalid access token")
)
