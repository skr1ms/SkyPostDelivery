package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewJWTService(t *testing.T) {
	accessSecret := "test-access-secret"
	refreshSecret := "test-refresh-secret"
	accessTTL := 15 * time.Minute
	refreshTTL := 7 * 24 * time.Hour

	service := NewJWTService(accessSecret, refreshSecret, accessTTL, refreshTTL)

	assert.NotNil(t, service)
	assert.Equal(t, []byte(accessSecret), service.accessSecret)
	assert.Equal(t, []byte(refreshSecret), service.refreshSecret)
	assert.Equal(t, accessTTL, service.accessTTL)
	assert.Equal(t, refreshTTL, service.refreshTTL)
}

func TestJWTService_GenerateTokenPair(t *testing.T) {
	service := NewJWTService("test-access", "test-refresh", 15*time.Minute, 24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"
	fullName := "Test User"
	role := "client"

	tokenPair, err := service.GenerateTokenPair(userID, email, fullName, role)

	assert.NoError(t, err)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Greater(t, tokenPair.AccessExpiresAt, time.Now().Unix())
	assert.Greater(t, tokenPair.RefreshExpiresAt, time.Now().Unix())
}

func TestJWTService_ValidateAccessToken_Success(t *testing.T) {
	service := NewJWTService("test-access", "test-refresh", 15*time.Minute, 24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"
	fullName := "Test User"
	role := "admin"

	tokenPair, err := service.GenerateTokenPair(userID, email, fullName, role)
	assert.NoError(t, err)

	claims, err := service.ValidateAccessToken(tokenPair.AccessToken)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, fullName, claims.FullName)
	assert.Equal(t, role, claims.Role)
}

func TestJWTService_ValidateAccessToken_InvalidToken(t *testing.T) {
	service := NewJWTService("test-access", "test-refresh", 15*time.Minute, 24*time.Hour)

	claims, err := service.ValidateAccessToken("invalid.token.here")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_ValidateAccessToken_WrongSecret(t *testing.T) {
	service1 := NewJWTService("secret1", "refresh1", 15*time.Minute, 24*time.Hour)
	service2 := NewJWTService("secret2", "refresh2", 15*time.Minute, 24*time.Hour)

	userID := uuid.New()
	tokenPair, err := service1.GenerateTokenPair(userID, "test@example.com", "Test", "client")
	assert.NoError(t, err)

	claims, err := service2.ValidateAccessToken(tokenPair.AccessToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_ValidateRefreshToken_Success(t *testing.T) {
	service := NewJWTService("test-access", "test-refresh", 15*time.Minute, 24*time.Hour)
	userID := uuid.New()
	email := "test@example.com"
	fullName := "Test User"
	role := "client"

	tokenPair, err := service.GenerateTokenPair(userID, email, fullName, role)
	assert.NoError(t, err)

	claims, err := service.ValidateRefreshToken(tokenPair.RefreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, fullName, claims.FullName)
	assert.Equal(t, role, claims.Role)
}

func TestJWTService_ValidateRefreshToken_InvalidToken(t *testing.T) {
	service := NewJWTService("test-access", "test-refresh", 15*time.Minute, 24*time.Hour)

	claims, err := service.ValidateRefreshToken("invalid.refresh.token")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTService_RefreshTokens_Success(t *testing.T) {
	service := NewJWTService("test-access", "test-refresh", 15*time.Minute, 24*time.Hour)
	userID := uuid.New()

	originalPair, err := service.GenerateTokenPair(userID, "test@example.com", "Test", "client")
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	newPair, err := service.RefreshTokens(originalPair.RefreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, newPair)
	assert.NotEqual(t, originalPair.AccessToken, newPair.AccessToken)
	assert.NotEqual(t, originalPair.RefreshToken, newPair.RefreshToken)
}

func TestJWTService_RefreshTokens_InvalidToken(t *testing.T) {
	service := NewJWTService("test-access", "test-refresh", 15*time.Minute, 24*time.Hour)

	newPair, err := service.RefreshTokens("invalid.token")

	assert.Error(t, err)
	assert.Nil(t, newPair)
}

func TestJWTService_TokenExpiry(t *testing.T) {
	service := NewJWTService("test-access", "test-refresh", 1*time.Second, 2*time.Second)
	userID := uuid.New()

	tokenPair, err := service.GenerateTokenPair(userID, "test@example.com", "Test", "client")
	assert.NoError(t, err)

	claims, err := service.ValidateAccessToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	time.Sleep(2 * time.Second)

	claims, err = service.ValidateAccessToken(tokenPair.AccessToken)
	assert.Error(t, err)
	assert.Nil(t, claims)
}
