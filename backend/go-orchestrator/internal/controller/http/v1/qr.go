package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skr1ms/hitech-ekb/internal/controller/http/middleware"
	"github.com/skr1ms/hitech-ekb/internal/controller/http/v1/request"
	"github.com/skr1ms/hitech-ekb/internal/controller/http/v1/response"
	"github.com/skr1ms/hitech-ekb/internal/usecase"
)

type qrRoutes struct {
	uc            *usecase.QRUseCase
	jwtMiddleware *middleware.JWTMiddleware
}

func newQRRoutes(g *gin.RouterGroup, uc *usecase.QRUseCase, jwtMiddleware *middleware.JWTMiddleware, qrRateLimiter gin.HandlerFunc) {
	r := &qrRoutes{
		uc:            uc,
		jwtMiddleware: jwtMiddleware,
	}

	group := g.Group("/qr")
	{
		group.POST("/validate", qrRateLimiter, r.validate)

		protected := group.Group("")
		protected.Use(jwtMiddleware.RequireAuth())
		{
			protected.POST("/refresh", r.refresh)
		}
	}
}

// @Summary      Валидация QR-кода
// @Description  Проверяет валидность QR-кода клиента (используется постаматом)
// @Tags         qr
// @Accept       json
// @Produce      json
// @Param        request body request.ValidateRequest true "Данные QR-кода для валидации"
// @Success      200 {object} response.ValidateResponse
// @Failure      400 {object} response.Error
// @Failure      401 {object} response.Error
// @Router       /qr/validate [post]
func (qr *qrRoutes) validate(c *gin.Context) {
	var req request.ValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
		return
	}

	user, err := qr.uc.ValidateQR(c.Request.Context(), req.QRData)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Invalid or expired QR code"})
		return
	}

	expiresAt := int64(0)
	if user.QRExpiresAt != nil {
		expiresAt = user.QRExpiresAt.Unix()
	}

	userEmail := ""
	if user.Email != nil {
		userEmail = *user.Email
	}

	c.JSON(http.StatusOK, response.ValidateResponse{
		Valid:     true,
		UserID:    user.ID.String(),
		Email:     userEmail,
		FullName:  user.FullName,
		ExpiresAt: expiresAt,
	})
}

// @Summary      Обновить QR-код
// @Description  Обновляет QR-код клиента (продлевает срок действия)
// @Tags         qr
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200 {object} response.RefreshResponse
// @Failure      401 {object} response.Error
// @Failure      500 {object} response.Error
// @Router       /qr/refresh [post]
func (qr *qrRoutes) refresh(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.Error{Error: "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, response.Error{Error: "Invalid user ID"})
		return
	}

	qrInfo, qrCodeBase64, err := qr.uc.RefreshQR(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.RefreshResponse{
		QRCode:    qrCodeBase64,
		ExpiresAt: qrInfo.ExpiresAt,
	})
}
