package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skr1ms/SkyPostDelivery/locker-agent/pkg/logger"
)

func Logger(log logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		log.Info("HTTP request", nil, map[string]any{
			"method":   method,
			"path":     path,
			"status":   status,
			"duration": duration.String(),
			"ip":       c.ClientIP(),
		})
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func RequireInitialized(isInitialized func() bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isInitialized() {
			c.JSON(503, gin.H{
				"error":   "Service not initialized",
				"message": "Cell mapping not synced. Please call POST /api/cells/sync first or wait for orchestrator webhook",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
