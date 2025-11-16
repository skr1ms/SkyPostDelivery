package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		if statusCode >= 400 {
			c.Writer.Header().Set("X-Request-Duration", duration.String())
		}
	}
}
