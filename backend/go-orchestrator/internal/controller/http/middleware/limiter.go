package middleware

import (
	"context"

	limiter "github.com/davidleitw/gin-limiter"
	"github.com/gin-gonic/gin"
	redis "github.com/go-redis/redis/v8"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
)

type Limiter struct {
	redis      *redis.Client
	dispatcher *limiter.Dispatcher
	logger     logger.Interface
}

const (
	UserRateLimit  = 200
	QrRateLimit    = 30
	OrderRateLimit = 300

	UserPeriod  = "5-M"
	QrPeriod    = "1-M"
	OrderPeriod = "5-M"

	DefaultPeriod = "24-M"
	DefaultLimit  = 200
)

func NewLimiter(redis *redis.Client, log logger.Interface) *Limiter {
	ctx := context.Background()
	if err := redis.Ping(ctx).Err(); err != nil {
		log.Error("rate_limiter - NewLimiter - failed to connect to Redis", err, nil)
		return &Limiter{
			redis:  redis,
			logger: log,
		}
	}

	dispatcher, err := limiter.LimitDispatcher(DefaultPeriod, DefaultLimit, redis)
	if err != nil {
		log.Error("rate_limiter - NewLimiter - limiter.LimitDispatcher", err, nil)
		return &Limiter{
			redis:  redis,
			logger: log,
		}
	}
	return &Limiter{redis: redis, dispatcher: dispatcher, logger: log}
}

func (l *Limiter) MiddleWare(key string, limit int) gin.HandlerFunc {
	if l.dispatcher == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return l.dispatcher.MiddleWare(key, limit)
}
