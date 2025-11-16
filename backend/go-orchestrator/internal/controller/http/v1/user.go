package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/request"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/response"
	_ "github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/entity"
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

// @Summary      Регистрация пользователя
// @Description  Создает нового пользователя и отправляет SMS с кодом подтверждения
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body request.CreateUser true "Данные пользователя"
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

	_, err := r.userUC.Register(c.Request.Context(), req.FullName, req.Email, req.Phone, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response.PhoneCodeSent{
		Message: "Registration successful. SMS code sent to " + req.Phone,
	})
}

// @Summary      Подтверждение телефона
// @Description  Проверяет SMS-код и активирует аккаунт пользователя
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body request.VerifyPhone true "Телефон и код"
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
		c.JSON(http.StatusUnauthorized, response.Error{Error: err.Error()})
		return
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	tokenPair, err := r.jwtService.GenerateTokenPair(user.ID, email, user.FullName, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, response.Login{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		QRCode:       qrCode,
	})
}

// @Summary      Вход по email
// @Description  Аутентификация пользователя с email и паролем
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body request.Login true "Email и пароль"
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
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Invalid credentials"})
		return
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	tokenPair, err := r.jwtService.GenerateTokenPair(user.ID, email, user.FullName, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, response.Login{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		QRCode:       qrCode,
	})
}

// @Summary      Вход по телефону
// @Description  Отправляет SMS-код для входа пользователя по номеру телефона
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body request.LoginByPhone true "Номер телефона"
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
		errMsg := err.Error()
		if errMsg == "user with this phone not found" {
			c.JSON(http.StatusNotFound, response.Error{Error: "User not found"})
			return
		}
		if errMsg == "phone not verified" {
			c.JSON(http.StatusBadRequest, response.Error{Error: "Phone not verified"})
			return
		}
		if len(errMsg) > 100 {
			c.JSON(http.StatusInternalServerError, response.Error{Error: "Failed to send SMS. Please try again later"})
		} else {
			c.JSON(http.StatusInternalServerError, response.Error{Error: errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, response.PhoneCodeSent{
		Message: "Login code sent to your phone",
	})
}

// @Summary      Запросить сброс пароля
// @Description  Отправляет SMS-код для сброса пароля
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body request.RequestPasswordReset true "Номер телефона"
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
		errMsg := err.Error()
		if errMsg == "user with this phone not found" {
			c.JSON(http.StatusNotFound, response.Error{Error: "User not found"})
			return
		}
		if errMsg == "phone not verified" {
			c.JSON(http.StatusBadRequest, response.Error{Error: "Phone not verified"})
			return
		}
		if len(errMsg) > 100 {
			c.JSON(http.StatusInternalServerError, response.Error{Error: "Failed to send SMS. Please try again later"})
		} else {
			c.JSON(http.StatusInternalServerError, response.Error{Error: errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, response.PhoneCodeSent{
		Message: "Password reset code sent to your phone",
	})
}

// @Summary      Сбросить пароль
// @Description  Устанавливает новый пароль после проверки SMS-кода
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body request.ResetPassword true "Телефон, код и новый пароль"
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
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessMessage{
		Message: "Password reset successfully",
	})
}

// @Summary      Обновить токены
// @Description  Получает новую пару access и refresh токенов используя refresh token
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
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Invalid or expired refresh token"})
		return
	}

	newTokenPair, err := r.jwtService.GenerateTokenPair(
		claims.UserID,
		claims.Email,
		claims.FullName,
		claims.Role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: "Failed to generate new tokens"})
		return
	}

	c.JSON(http.StatusOK, response.RefreshToken{
		AccessToken:  newTokenPair.AccessToken,
		RefreshToken: newTokenPair.RefreshToken,
		ExpiresAt:    newTokenPair.ExpiresAt,
	})
}

// @Summary      Текущий пользователь
// @Description  Возвращает информацию о текущем авторизованном пользователе
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
		c.JSON(http.StatusNotFound, response.Error{Error: "User not found"})
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

	if err := r.notificationUC.RegisterDevice(c.Request.Context(), targetUserID, req.Token, req.Platform); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessMessage{Message: "Device registered"})
}
