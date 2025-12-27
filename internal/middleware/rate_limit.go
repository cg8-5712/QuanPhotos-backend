package middleware

import (
	"net/http"
	"sync"
	"time"

	"QuanPhotos/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	// Requests per window
	Limit int
	// Window duration
	Window time.Duration
	// Key function to identify clients (e.g., by IP or user ID)
	KeyFunc func(*gin.Context) string
}

// rateLimiterEntry tracks requests for a key
type rateLimiterEntry struct {
	count     int
	resetTime time.Time
}

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	config  RateLimiterConfig
	entries map[string]*rateLimiterEntry
	mu      sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		config:  config,
		entries: make(map[string]*rateLimiterEntry),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup periodically removes expired entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.Window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.entries {
			if now.After(entry.resetTime) {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]

	if !exists || now.After(entry.resetTime) {
		rl.entries[key] = &rateLimiterEntry{
			count:     1,
			resetTime: now.Add(rl.config.Window),
		}
		return true
	}

	if entry.count >= rl.config.Limit {
		return false
	}

	entry.count++
	return true
}

// Middleware returns a gin middleware function
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rl.config.KeyFunc(c)

		if !rl.Allow(key) {
			response.Error(c, http.StatusTooManyRequests, response.CodeTooManyRequests, "too many requests, please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ============================================
// Predefined Rate Limiters
// ============================================

// IPKeyFunc returns the client IP as the key
func IPKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}

// UserKeyFunc returns the user ID as the key (requires authentication)
func UserKeyFunc(c *gin.Context) string {
	userID := c.GetInt64("userID")
	if userID == 0 {
		return c.ClientIP()
	}
	return "user:" + string(rune(userID))
}

// GlobalRateLimiter creates a global rate limiter (per IP)
func GlobalRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	rl := NewRateLimiter(RateLimiterConfig{
		Limit:   limit,
		Window:  window,
		KeyFunc: IPKeyFunc,
	})
	return rl.Middleware()
}

// LoginRateLimiter creates a rate limiter for login attempts
func LoginRateLimiter() gin.HandlerFunc {
	rl := NewRateLimiter(RateLimiterConfig{
		Limit:   5,           // 5 attempts
		Window:  time.Minute, // per minute
		KeyFunc: IPKeyFunc,
	})
	return rl.Middleware()
}

// UploadRateLimiter creates a rate limiter for file uploads
func UploadRateLimiter() gin.HandlerFunc {
	rl := NewRateLimiter(RateLimiterConfig{
		Limit:  10,          // 10 uploads
		Window: time.Hour,   // per hour
		KeyFunc: func(c *gin.Context) string {
			userID := c.GetInt64("userID")
			if userID == 0 {
				return "upload:" + c.ClientIP()
			}
			return "upload:user:" + string(rune(userID))
		},
	})
	return rl.Middleware()
}

// APIRateLimiter creates a general API rate limiter
func APIRateLimiter() gin.HandlerFunc {
	rl := NewRateLimiter(RateLimiterConfig{
		Limit:   100,         // 100 requests
		Window:  time.Minute, // per minute
		KeyFunc: IPKeyFunc,
	})
	return rl.Middleware()
}
