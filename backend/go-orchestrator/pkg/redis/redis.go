package redis

import (
	"strings"

	redis "github.com/go-redis/redis/v8"
	"github.com/skr1ms/SkyPostDelivery/go-orchestrator/config"
)

func New(cfg *config.Redis) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     strings.TrimPrefix(cfg.URL, "redis://"),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}
