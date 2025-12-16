package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/internal/controller/http/v1/response"
)

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, response.Health{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "locker-agent",
	})
}

func serviceInfo(c *gin.Context) {
	c.JSON(http.StatusOK, response.ServiceInfo{
		Service: "Locker Agent Service",
		Version: "1.0.0",
		Status:  "running",
	})
}
