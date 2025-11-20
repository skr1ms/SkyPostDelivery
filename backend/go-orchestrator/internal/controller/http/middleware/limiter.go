package middleware

import (
	limiter "github.com/davidleitw/gin-limiter"
	"github.com/gin-gonic/gin"
	redis "github.com/go-redis/redis/v8"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/pkg/logger"
)

type Limiter struct {
	redis      *redis.Client
	dispatcher *limiter.Dispatcher
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

func NewLimiter(redis *redis.Client) *Limiter {
	dispatcher, err := limiter.LimitDispatcher(DefaultPeriod, DefaultLimit, redis)
	if err != nil {
		logger.New("error").Error("rate_limiter - NewLimiter - limiter.LimitDispatcher", err)
		panic(err)
	}
	return &Limiter{redis: redis, dispatcher: dispatcher}
}

func (l *Limiter) MiddleWare(key string, limit int) gin.HandlerFunc {
	return l.dispatcher.MiddleWare(key, limit)
}
