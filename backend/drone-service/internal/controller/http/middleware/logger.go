package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/skr1ms/SkyPostDelivery/drone-service/pkg/logger"
	"time"
)

func Logger(log logger.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)

		fields := map[string]any{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"query":      query,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"latency":    latency.Milliseconds(),
		}

		if c.Writer.Status() >= 500 {
			log.Error("HTTP request", nil, fields)
		} else if c.Writer.Status() >= 400 {
			log.Warn("HTTP request", nil, fields)
		} else {
			log.Info("HTTP request", nil, fields)
		}
	}
}
