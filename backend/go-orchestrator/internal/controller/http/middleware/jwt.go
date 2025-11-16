package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/jwt"
)

const (
	AuthorizationHeader = "Authorization"
	BearerSchema        = "Bearer "
	UserIDKey           = "user_id"
	EmailKey            = "email"
	FullNameKey         = "full_name"
	RoleKey             = "role"
	ClaimsKey           = "jwt_claims"
)

type JWTMiddleware struct {
	JWTService *jwt.JWTService
}

func NewJWTMiddleware(jwtService *jwt.JWTService) *JWTMiddleware {
	return &JWTMiddleware{
		JWTService: jwtService,
	}
}

func (j *JWTMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := j.extractToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: " + err.Error(),
			})
			c.Abort()
			return
		}

		claims, err := j.JWTService.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: Invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(EmailKey, claims.Email)
		c.Set(FullNameKey, claims.FullName)
		c.Set(RoleKey, claims.Role)
		c.Set(ClaimsKey, claims)

		c.Next()
	}
}

func (j *JWTMiddleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(ClaimsKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: Token not found",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*jwt.CustomClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: Invalid token claims",
			})
			c.Abort()
			return
		}

		for _, role := range allowedRoles {
			if userClaims.Role == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Forbidden: Insufficient permissions",
		})
		c.Abort()
	}
}

func (j *JWTMiddleware) AdminOnly() gin.HandlerFunc {
	return j.RequireRole("admin")
}

func (j *JWTMiddleware) ClientOrAdmin() gin.HandlerFunc {
	return j.RequireRole("client", "admin")
}

func (j *JWTMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := j.extractToken(c)
		if err != nil {
			c.Next()
			return
		}

		claims, err := j.JWTService.ValidateAccessToken(token)
		if err != nil {
			c.Next()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(EmailKey, claims.Email)
		c.Set(FullNameKey, claims.FullName)
		c.Set(RoleKey, claims.Role)
		c.Set(ClaimsKey, claims)

		c.Next()
	}
}

func (j *JWTMiddleware) extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader(AuthorizationHeader)
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	if !strings.HasPrefix(authHeader, BearerSchema) {
		return "", errors.New("invalid authorization header format")
	}

	return strings.TrimPrefix(authHeader, BearerSchema), nil
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	return userID.(uuid.UUID), true
}

func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(EmailKey)
	if !exists {
		return "", false
	}
	return email.(string), true
}

func GetUserFullName(c *gin.Context) (string, bool) {
	fullName, exists := c.Get(FullNameKey)
	if !exists {
		return "", false
	}
	return fullName.(string), true
}

func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(RoleKey)
	if !exists {
		return "", false
	}
	return role.(string), true
}

func GetClaims(c *gin.Context) (*jwt.CustomClaims, bool) {
	claims, exists := c.Get(ClaimsKey)
	if !exists {
		return nil, false
	}
	return claims.(*jwt.CustomClaims), true
}

func IsAdmin(c *gin.Context) bool {
	role, exists := GetUserRole(c)
	return exists && role == "admin"
}

func MustGetUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		c.Abort()
		return uuid.Nil
	}

	switch v := userID.(type) {
	case uuid.UUID:
		return v
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return uuid.Nil
		}
		return id
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID type"})
		c.Abort()
		return uuid.Nil
	}
}
