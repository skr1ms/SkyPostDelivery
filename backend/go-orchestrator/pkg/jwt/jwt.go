package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

type CustomClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role,omitempty"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	AccessExpiresAt  int64  `json:"access_expires_at"`
	RefreshExpiresAt int64  `json:"refresh_expires_at"`
}

func NewJWTService(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (j *JWTService) GenerateTokenPair(userID uuid.UUID, email, name, role string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(j.accessTTL)
	refreshExpiry := now.Add(j.refreshTTL)

	accessClaims := &CustomClaims{
		UserID:    userID,
		Email:     email,
		FullName:  name,
		Role:      role,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "skypost-delivery-orchestrator",
		},
	}

	refreshClaims := &CustomClaims{
		UserID:    userID,
		Email:     email,
		FullName:  name,
		Role:      role,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "skypost-delivery-orchestrator",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("JWTService - GenerateTokenPair - SignedString[access]: %w", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("JWTService - GenerateTokenPair - SignedString[refresh]: %w", err)
	}

	return &TokenPair{
		AccessToken:      accessTokenString,
		RefreshToken:     refreshTokenString,
		AccessExpiresAt:  accessExpiry.Unix(),
		RefreshExpiresAt: refreshExpiry.Unix(),
	}, nil
}

func (j *JWTService) ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("JWTService - ValidateAccessToken - ValidateSigningMethod[alg=%v]: %w", token.Header["alg"], ErrTokenUnexpectedSigning)
		}
		return j.accessSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("JWTService - ValidateAccessToken - ParseWithClaims: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		if claims.TokenType != TokenTypeAccess {
			return nil, ErrTokenInvalidType
		}
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

func (j *JWTService) ValidateRefreshToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("JWTService - ValidateRefreshToken - ValidateSigningMethod[alg=%v]: %w", token.Header["alg"], ErrTokenUnexpectedSigning)
		}
		return j.refreshSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("JWTService - ValidateRefreshToken - ParseWithClaims: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		if claims.TokenType != TokenTypeRefresh {
			return nil, ErrTokenInvalidType
		}
		return claims, nil
	}

	return nil, ErrRefreshTokenInvalid
}

func (j *JWTService) RefreshTokens(refreshTokenString string) (*TokenPair, error) {
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	return j.GenerateTokenPair(claims.UserID, claims.Email, claims.FullName, claims.Role)
}
