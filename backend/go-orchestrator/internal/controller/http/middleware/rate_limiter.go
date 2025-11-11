package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	attempts map[string]*attemptInfo
	mu       sync.RWMutex
	cleanup  time.Duration
}

type attemptInfo struct {
	count        int
	lastAttempt  time.Time
	blockedUntil time.Time
}

type RateLimitConfig struct {
	MaxAttempts   int
	Window        time.Duration
	BlockDuration time.Duration
}

var (
	authLimiter = &rateLimiter{
		attempts: make(map[string]*attemptInfo),
		cleanup:  time.Minute * 5,
	}

	apiLimiter = &rateLimiter{
		attempts: make(map[string]*attemptInfo),
		cleanup:  time.Minute,
	}
)

func init() {
	go authLimiter.cleanupLoop()
	go apiLimiter.cleanupLoop()
}

func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, info := range rl.attempts {
			if now.Sub(info.lastAttempt) > rl.cleanup {
				delete(rl.attempts, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) check(ip string, cfg RateLimitConfig) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.attempts[ip]

	if !exists {
		rl.attempts[ip] = &attemptInfo{
			count:       1,
			lastAttempt: now,
		}
		return true
	}

	if now.Before(info.blockedUntil) {
		return false
	}

	if now.Sub(info.lastAttempt) > cfg.Window {
		info.count = 1
		info.lastAttempt = now
		info.blockedUntil = time.Time{}
		return true
	}

	info.count++
	info.lastAttempt = now

	if info.count > cfg.MaxAttempts {
		info.blockedUntil = now.Add(cfg.BlockDuration)
		return false
	}

	return true
}

func (rl *rateLimiter) getBlockedUntil(ip string) time.Time {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if info, exists := rl.attempts[ip]; exists {
		return info.blockedUntil
	}
	return time.Time{}
}

func AuthRateLimiter() gin.HandlerFunc {
	cfg := RateLimitConfig{
		MaxAttempts:   10,
		Window:        time.Minute * 15,
		BlockDuration: time.Minute * 30,
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !authLimiter.check(ip, cfg) {
			blockedUntil := authLimiter.getBlockedUntil(ip)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many authentication attempts",
				"retry_after": int(time.Until(blockedUntil).Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func APIRateLimiter() gin.HandlerFunc {
	cfg := RateLimitConfig{
		MaxAttempts:   500,
		Window:        time.Minute,
		BlockDuration: time.Minute * 5,
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !apiLimiter.check(ip, cfg) {
			blockedUntil := apiLimiter.getBlockedUntil(ip)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": int(time.Until(blockedUntil).Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func QRScanRateLimiter() gin.HandlerFunc {
	cfg := RateLimitConfig{
		MaxAttempts:   5,
		Window:        time.Minute,
		BlockDuration: time.Minute * 10,
	}

	limiter := &rateLimiter{
		attempts: make(map[string]*attemptInfo),
		cleanup:  time.Minute * 5,
	}
	go limiter.cleanupLoop()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.check(ip, cfg) {
			blockedUntil := limiter.getBlockedUntil(ip)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many QR scan attempts",
				"retry_after": int(time.Until(blockedUntil).Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func OrderCreationRateLimiter() gin.HandlerFunc {
	cfg := RateLimitConfig{
		MaxAttempts:   20,
		Window:        time.Minute,
		BlockDuration: time.Minute * 5,
	}

	limiter := &rateLimiter{
		attempts: make(map[string]*attemptInfo),
		cleanup:  time.Minute * 5,
	}
	go limiter.cleanupLoop()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.check(ip, cfg) {
			blockedUntil := limiter.getBlockedUntil(ip)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many order creation attempts",
				"retry_after": int(time.Until(blockedUntil).Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
