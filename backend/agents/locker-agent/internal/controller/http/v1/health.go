package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type healthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

type serviceInfoResponse struct {
	Service string `json:"service"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "locker-agent",
	})
}

func serviceInfo(c *gin.Context) {
	c.JSON(http.StatusOK, serviceInfoResponse{
		Service: "Locker Agent Service",
		Version: "1.0.0",
		Status:  "running",
	})
}
