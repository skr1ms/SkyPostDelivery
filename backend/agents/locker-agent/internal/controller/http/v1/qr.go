package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/usecase"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/logger"
)

type qrRoutes struct {
	qrScanner *usecase.QRScannerUseCase
	logger    logger.Interface
}

func newQRRoutes(group *gin.RouterGroup, qrScanner *usecase.QRScannerUseCase, log logger.Interface) {
	r := &qrRoutes{
		qrScanner: qrScanner,
		logger:    log,
	}

	group.POST("/scan", r.scanQR)
	group.POST("/confirm-pickup", r.confirmPickup)
	group.POST("/confirm-loaded", r.confirmLoaded)
}

func (r *qrRoutes) scanQR(c *gin.Context) {
	var req entity.QRScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Error("Failed to bind QR scan request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := r.qrScanner.ProcessQRScan(c.Request.Context(), req.QRData)
	if err != nil {
		r.logger.Error("Failed to process QR scan", err)
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (r *qrRoutes) confirmPickup(c *gin.Context) {
	var req entity.ConfirmPickupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Error("Failed to bind confirm pickup request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.qrScanner.ConfirmPickup(c.Request.Context(), req.CellIDs); err != nil {
		r.logger.Error("Failed to confirm pickup", err)
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pickup confirmed successfully",
	})
}

func (r *qrRoutes) confirmLoaded(c *gin.Context) {
	var req entity.ConfirmLoadedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		r.logger.Error("Failed to bind confirm loaded request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.qrScanner.ConfirmLoaded(c.Request.Context(), req.OrderID, req.LockerCellID); err != nil {
		r.logger.Error("Failed to confirm loaded", err)
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Load confirmed successfully",
	})
}
