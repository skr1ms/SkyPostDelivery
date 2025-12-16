package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	entityError "github.com/skr1ms/SkyPostDelivery/locker-agent/internal/entity/error"
)

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, entityError.ErrCellInvalidNumber),
		errors.Is(err, entityError.ErrCellInvalidUUID),
		errors.Is(err, entityError.ErrQRInvalidFormat),
		errors.Is(err, entityError.ErrConfigValidationFailed),
		errors.Is(err, entityError.ErrConfigMissingRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	case errors.Is(err, entityError.ErrCellNotFound),
		errors.Is(err, entityError.ErrQRValidationFailed):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})

	case errors.Is(err, entityError.ErrCellNotInitialized):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})

	case errors.Is(err, entityError.ErrOrchestratorConnectionFailed),
		errors.Is(err, entityError.ErrOrchestratorRequestFailed),
		errors.Is(err, entityError.ErrOrchestratorTimeout),
		errors.Is(err, entityError.ErrHardwareNotAvailable),
		errors.Is(err, entityError.ErrArduinoConnectionFailed),
		errors.Is(err, entityError.ErrDisplayConnectionFailed),
		errors.Is(err, entityError.ErrCameraConnectionFailed):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
