package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/pkg/response"
)

// RateLimiter represents a simple in-memory rate limiter
type RateLimiter struct {
	visitors map[string]*Visitor
	mutex    *sync.RWMutex
	rate     int           // requests per minute
	capacity int           // burst capacity
	cleanup  time.Duration // cleanup interval
}

// Visitor represents a visitor's rate limit state
type Visitor struct {
	tokens    int
	lastSeen  time.Time
	resetTime time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, capacity int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		mutex:    &sync.RWMutex{},
		rate:     rate,
		capacity: capacity,
		cleanup:  time.Minute * 5, // cleanup every 5 minutes
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Get or create visitor
	visitor, exists := rl.visitors[key]
	if !exists {
		visitor = &Visitor{
			tokens:    rl.capacity,
			lastSeen:  now,
			resetTime: now.Add(time.Minute),
		}
		rl.visitors[key] = visitor
	}

	// Update last seen
	visitor.lastSeen = now

	// Reset tokens if minute has passed
	if now.After(visitor.resetTime) {
		visitor.tokens = rl.capacity
		visitor.resetTime = now.Add(time.Minute)
	}

	// Check if request is allowed
	if visitor.tokens > 0 {
		visitor.tokens--
		return true
	}

	return false
}

// cleanupVisitors removes old visitors to prevent memory leaks
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mutex.Lock()
			now := time.Now()
			cutoff := now.Add(-time.Hour) // Remove visitors not seen for 1 hour

			for key, visitor := range rl.visitors {
				if visitor.lastSeen.Before(cutoff) {
					delete(rl.visitors, key)
				}
			}
			rl.mutex.Unlock()
		}
	}
}

// Global rate limiters
var (
	generalLimiter = NewRateLimiter(60, 60) // 60 requests per minute
	authLimiter    = NewRateLimiter(10, 15) // 10 requests per minute for auth endpoints
	strictLimiter  = NewRateLimiter(5, 10)  // 5 requests per minute for sensitive endpoints
)

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Use client IP as the key
		key := getClientKey(c)

		if !limiter.Allow(key) {
			response.ErrorJSONWithLog(c, http.StatusTooManyRequests, "Rate limit exceeded", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// GeneralRateLimitMiddleware applies general rate limiting
func GeneralRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddleware(generalLimiter)
}

// AuthRateLimitMiddleware applies stricter rate limiting for auth endpoints
func AuthRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddleware(authLimiter)
}

// StrictRateLimitMiddleware applies very strict rate limiting for sensitive endpoints
func StrictRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddleware(strictLimiter)
}

// getClientKey generates a unique key for the client
func getClientKey(c *gin.Context) string {
	// Try to get authenticated user ID first
	if userID, exists := c.Get("user_id"); exists {
		return "user:" + strconv.FormatUint(uint64(userID.(uint)), 10)
	}

	// Fall back to IP address
	clientIP := c.ClientIP()

	// Check for forwarded IP headers
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		return "ip:" + forwarded
	}

	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return "ip:" + realIP
	}

	return "ip:" + clientIP
}

// UserBasedRateLimitMiddleware creates user-specific rate limiting
func UserBasedRateLimitMiddleware(rate, capacity int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, capacity)

	return gin.HandlerFunc(func(c *gin.Context) {
		// This middleware should be used after authentication middleware
		userID, exists := c.Get("user_id")
		if !exists {
			// If no user, fall back to IP-based limiting
			key := getClientKey(c)
			if !limiter.Allow(key) {
				response.ErrorJSONWithLog(c, http.StatusTooManyRequests, "Rate limit exceeded", nil)
				c.Abort()
				return
			}
		} else {
			// Use user-specific limiting
			key := "user:" + strconv.FormatUint(uint64(userID.(uint)), 10)
			if !limiter.Allow(key) {
				response.ErrorJSONWithLog(c, http.StatusTooManyRequests, "Rate limit exceeded", nil)
				c.Abort()
				return
			}
		}

		c.Next()
	})
}

// APIKeyRateLimitMiddleware creates API key-specific rate limiting
func APIKeyRateLimitMiddleware(rate, capacity int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, capacity)

	return gin.HandlerFunc(func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Fall back to general rate limiting
			key := getClientKey(c)
			if !limiter.Allow(key) {
				response.ErrorJSONWithLog(c, http.StatusTooManyRequests, "Rate limit exceeded", nil)
				c.Abort()
				return
			}
		} else {
			// Use API key-specific limiting
			key := "api:" + apiKey
			if !limiter.Allow(key) {
				response.ErrorJSONWithLog(c, http.StatusTooManyRequests, "API rate limit exceeded", nil)
				c.Abort()
				return
			}
		}

		c.Next()
	})
}

// DynamicRateLimitMiddleware creates dynamic rate limiting based on endpoint sensitivity
func DynamicRateLimitMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		var limiter *RateLimiter

		// Determine which limiter to use based on endpoint
		switch {
		case isAuthEndpoint(path):
			limiter = authLimiter
		case isSensitiveEndpoint(path, method):
			limiter = strictLimiter
		default:
			limiter = generalLimiter
		}

		key := getClientKey(c)
		if !limiter.Allow(key) {
			response.ErrorJSONWithLog(c, http.StatusTooManyRequests, "Rate limit exceeded", nil)
			c.Abort()
			return
		}

		c.Next()
	})
}

// isAuthEndpoint checks if the path is an authentication endpoint
func isAuthEndpoint(path string) bool {
	authPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
		"/api/v1/auth/forgot-password",
		"/api/v1/auth/reset-password",
	}

	for _, authPath := range authPaths {
		if path == authPath {
			return true
		}
	}

	return false
}

// isSensitiveEndpoint checks if the endpoint is sensitive and needs strict limiting
func isSensitiveEndpoint(path, method string) bool {
	// DELETE operations are generally sensitive
	if method == "DELETE" {
		return true
	}

	// Sensitive paths
	sensitivePaths := []string{
		"/api/v1/users",        // User management
		"/api/v1/settings",     // System settings
		"/api/v1/certificates", // Certificate operations
	}

	for _, sensitivePath := range sensitivePaths {
		if path == sensitivePath && (method == "POST" || method == "PUT" || method == "PATCH") {
			return true
		}
	}

	return false
}
