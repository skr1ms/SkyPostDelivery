package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/request"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/response"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/jwt"
)

type authRoutes struct {
	userUC         *usecase.UserUseCase
	jwtService     *jwt.JWTService
	notificationUC *usecase.NotificationUseCase
}

func newUserRoutes(g *gin.RouterGroup, userUC *usecase.UserUseCase, jwtService *jwt.JWTService, notificationUC *usecase.NotificationUseCase, protectedGroup *gin.RouterGroup, authRateLimiter gin.HandlerFunc) {
	r := &authRoutes{userUC: userUC, jwtService: jwtService, notificationUC: notificationUC}

	auth := g.Group("/auth")
	{
		auth.POST("/register", authRateLimiter, r.register)
		auth.POST("/verify/phone", authRateLimiter, r.verifyPhone)
		auth.POST("/login", authRateLimiter, r.login)
		auth.POST("/login/phone", authRateLimiter, r.loginByPhone)
		auth.POST("/password/reset/request", authRateLimiter, r.requestPasswordReset)
		auth.POST("/password/reset", authRateLimiter, r.resetPassword)
		auth.POST("/refresh", r.refresh)
	}

	protectedAuth := protectedGroup.Group("/auth")
	{
		protectedAuth.GET("/me", r.me)
	}

	protectedUsers := protectedGroup.Group("/users")
	{
		protectedUsers.POST("/:id/devices", r.registerDevice)
	}
}

// @Summary      User registration
// @Description  Creates a new user and sends SMS with verification code
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body request.CreateUser true "User data"
// @Success      201 {object} response.PhoneCodeSent
// @Failure      400 {object} response.Error
// @Failure      500 {object} response.Error
// @Router       /auth/register [post]
func (r *authRoutes) register(c *gin.Context) {
	var req request.CreateUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	user := &entity.User{
		FullName:    req.FullName,
		Email:       &req.Email,
		PhoneNumber: &req.Phone,
		PassHash:    &req.Password,
	}

	_, err := r.userUC.Register(c.Request.Context(), user)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.PhoneCodeSent{
		Message: "Registration successful. SMS code sent to " + req.Phone,
	})
}

// @Summary      Phone verification
// @Description  Verifies SMS code and activates user account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body request.VerifyPhone true "Phone and code"
// @Success      200 {object} response.Login
// @Failure      400 {object} response.Error
// @Failure      401 {object} response.Error
// @Failure      500 {object} response.Error
// @Router       /auth/verify/phone [post]
func (r *authRoutes) verifyPhone(c *gin.Context) {
	var req request.VerifyPhone
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	user, qrCode, err := r.userUC.VerifyPhoneCode(c.Request.Context(), req.Phone, req.Code)
	if err != nil {
		handleError(c, err)
		return
	}

	tokenPair, err := r.jwtService.GenerateTokenPair(user.ID, user.GetEmail(), user.FullName, user.Role)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Login{
		User:             user,
		AccessToken:      tokenPair.AccessToken,
		RefreshToken:     tokenPair.RefreshToken,
		AccessExpiresAt:  tokenPair.AccessExpiresAt,
		RefreshExpiresAt: tokenPair.RefreshExpiresAt,
		QRCode:           qrCode,
	})
}

// @Summary      Login by email
// @Description  Authenticates user with email and password
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body request.Login true "Email and password"
// @Success      200 {object} response.Login
// @Failure      400 {object} response.Error
// @Failure      401 {object} response.Error
// @Failure      500 {object} response.Error
// @Router       /auth/login [post]
func (r *authRoutes) login(c *gin.Context) {
	var req request.Login
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	user, qrCode, err := r.userUC.LoginByCredentials(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		handleError(c, err)
		return
	}

	tokenPair, err := r.jwtService.GenerateTokenPair(user.ID, user.GetEmail(), user.FullName, user.Role)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Login{
		User:             user,
		AccessToken:      tokenPair.AccessToken,
		RefreshToken:     tokenPair.RefreshToken,
		AccessExpiresAt:  tokenPair.AccessExpiresAt,
		RefreshExpiresAt: tokenPair.RefreshExpiresAt,
		QRCode:           qrCode,
	})
}

// @Summary      Login by phone
// @Description  Sends SMS code for user login by phone number
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body request.LoginByPhone true "Phone number"
// @Success      200 {object} response.PhoneCodeSent
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Failure      500 {object} response.Error
// @Router       /auth/login/phone [post]
func (r *authRoutes) loginByPhone(c *gin.Context) {
	var req request.LoginByPhone
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	err := r.userUC.LoginByPhone(c.Request.Context(), req.Phone)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.PhoneCodeSent{
		Message: "Login code sent to your phone",
	})
}

// @Summary      Request password reset
// @Description  Sends SMS code for password reset
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body request.RequestPasswordReset true "Phone number"
// @Success      200 {object} response.PhoneCodeSent
// @Failure      400 {object} response.Error
// @Failure      404 {object} response.Error
// @Failure      500 {object} response.Error
// @Router       /auth/password/reset/request [post]
func (r *authRoutes) requestPasswordReset(c *gin.Context) {
	var req request.RequestPasswordReset
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	err := r.userUC.RequestPasswordReset(c.Request.Context(), req.Phone)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.PhoneCodeSent{
		Message: "Password reset code sent to your phone",
	})
}

// @Summary      Reset password
// @Description  Sets new password after SMS code verification
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body request.ResetPassword true "Phone, code and new password"
// @Success      200 {object} response.SuccessMessage
// @Failure      400 {object} response.Error
// @Router       /auth/password/reset [post]
func (r *authRoutes) resetPassword(c *gin.Context) {
	var req request.ResetPassword
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	err := r.userUC.ResetPassword(c.Request.Context(), req.Phone, req.Code, req.NewPassword)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.SuccessMessage{
		Message: "Password reset successfully",
	})
}

// @Summary      Refresh tokens
// @Description  Gets new access and refresh token pair using refresh token
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200 {object} response.RefreshToken
// @Failure      401 {object} response.Error
// @Failure      500 {object} response.Error
// @Security     Bearer
// @Router       /auth/refresh [post]
func (r *authRoutes) refresh(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Authorization header required"})
		return
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Invalid authorization format"})
		return
	}

	refreshToken := authHeader[7:]

	claims, err := r.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		handleError(c, err)
		return
	}

	newTokenPair, err := r.jwtService.GenerateTokenPair(
		claims.UserID,
		claims.Email,
		claims.FullName,
		claims.Role,
	)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.RefreshToken{
		AccessToken:      newTokenPair.AccessToken,
		RefreshToken:     newTokenPair.RefreshToken,
		AccessExpiresAt:  newTokenPair.AccessExpiresAt,
		RefreshExpiresAt: newTokenPair.RefreshExpiresAt,
	})
}

// @Summary      Current user
// @Description  Returns information about the current authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200 {object} response.UserWithQR
// @Failure      401 {object} response.Error
// @Failure      404 {object} response.Error
// @Security     Bearer
// @Router       /auth/me [get]
func (r *authRoutes) me(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error{Error: "User ID not found in context"})
		return
	}

	var userID string
	switch v := userIDValue.(type) {
	case uuid.UUID:
		userID = v.String()
	case string:
		userID = v
	default:
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Invalid user ID type"})
		return
	}

	user, qrCode, err := r.userUC.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.UserWithQR{
		User:   user,
		QRCode: qrCode,
	})
}

func (r *authRoutes) registerDevice(c *gin.Context) {
	var req request.RegisterDevice
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error{Error: "User ID not found in context"})
		return
	}

	paramID := c.Param("id")
	targetUserID, err := uuid.Parse(paramID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: "Invalid user ID"})
		return
	}

	var currentUserID uuid.UUID
	switch v := userIDValue.(type) {
	case uuid.UUID:
		currentUserID = v
	case string:
		id, parseErr := uuid.Parse(v)
		if parseErr != nil {
			c.JSON(http.StatusUnauthorized, response.Error{Error: "Invalid user ID type"})
			return
		}
		currentUserID = id
	default:
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Invalid user ID type"})
		return
	}

	if currentUserID != targetUserID {
		c.JSON(http.StatusForbidden, response.Error{Error: "Forbidden"})
		return
	}

	if r.notificationUC == nil {
		c.JSON(http.StatusOK, response.SuccessMessage{Message: "Notifications are disabled"})
		return
	}

	device := &entity.Device{
		UserID:   targetUserID,
		Token:    req.Token,
		Platform: req.Platform,
	}

	if err := r.notificationUC.RegisterDevice(c.Request.Context(), device); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.SuccessMessage{Message: "Device registered"})
}
