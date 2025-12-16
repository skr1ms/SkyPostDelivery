package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/middleware"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/request"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/controller/http/v1/response"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/internal/usecase"
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
			protected.GET("/me", r.getMyQR)
			protected.POST("/refresh", r.refresh)
		}
	}
}

// @Summary      QR code validation
// @Description  Validates client QR code (used by parcel automat)
// @Tags         qr
// @Accept       json
// @Produce      json
// @Param        request body request.ValidateRequest true "QR code data for validation"
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
		handleError(c, err)
		return
	}

	var accessExpiresAt, refreshExpiresAt int64
	if user.QRIssuedAt != nil {
		accessExpiresAt = user.QRIssuedAt.Unix()
	}
	if user.QRExpiresAt != nil {
		refreshExpiresAt = user.QRExpiresAt.Unix()
	}

	c.JSON(http.StatusOK, response.ValidateResponse{
		Valid:            true,
		UserID:           user.ID.String(),
		Email:            user.GetEmail(),
		FullName:         user.FullName,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	})
}

// @Summary      Refresh QR code
// @Description  Refreshes client QR code (extends expiration time)
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
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.RefreshResponse{
		QRCode:           qrCodeBase64,
		IssuedAt:         qrInfo.IssuedAt,
		AccessExpiresAt:  qrInfo.IssuedAt,
		RefreshExpiresAt: qrInfo.ExpiresAt,
	})
}

// @Summary      Get my QR code
// @Description  Returns current user QR code, automatically refreshes if expired
// @Tags         qr
// @Produce      json
// @Security     Bearer
// @Success      200 {object} response.QRResponse
// @Failure      401 {object} response.Error
// @Failure      500 {object} response.Error
// @Router       /qr/me [get]
func (qr *qrRoutes) getMyQR(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	// Получить QR (с автообновлением если истёк)
	qrInfo, qrCodeBase64, err := qr.uc.GetOrRefreshQR(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.QRResponse{
		QRCode:    qrCodeBase64,
		IssuedAt:  qrInfo.IssuedAt,
		ExpiresAt: qrInfo.ExpiresAt,
	})
}
