package controllers

import (
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/configs"
	"github.com/nguyendkn/nginx-manager/pkg/logger"
	"github.com/nguyendkn/nginx-manager/pkg/response"
)

// HealthController handles health check related endpoints
type HealthController struct {
	env *configs.Environment
}

// NewHealthController creates a new health controller instance
func NewHealthController(env *configs.Environment) *HealthController {
	return &HealthController{
		env: env,
	}
}

// HealthCheck handles the health check endpoint
// @Summary Health check endpoint
// @Description Returns the health status of the service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /health [get]
func (hc *HealthController) HealthCheck(c *gin.Context) {
	// Get system information
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	healthData := gin.H{
		"status":    "healthy",
		"service":   hc.env.GetAppName(),
		"version":   hc.env.GetAppVersion(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
		"system": gin.H{
			"go_version":   runtime.Version(),
			"goroutines":   runtime.NumGoroutine(),
			"memory_alloc": memStats.Alloc,
			"memory_total": memStats.TotalAlloc,
			"memory_sys":   memStats.Sys,
			"gc_runs":      memStats.NumGC,
		},
		"environment": gin.H{
			"app_env":    hc.env.GetAppEnvironment(),
			"gin_mode":   hc.env.GetGinMode(),
			"host":       hc.env.GetHost(),
			"port":       hc.env.GetPort(),
			"log_level":  hc.env.GetLogLevel(),
			"log_format": hc.env.GetLogEncoding(),
		},
		"cors": gin.H{
			"allowed_origins": hc.env.GetCORSAllowedOrigins(),
			"allowed_methods": hc.env.GetCORSAllowedMethods(),
			"allowed_headers": hc.env.GetCORSAllowedHeaders(),
		},
	}

	// Log health check request
	logger.Info("Health check requested",
		logger.String("client_ip", c.ClientIP()),
		logger.String("user_agent", c.Request.UserAgent()),
	)

	response.SuccessJSONWithLog(c, healthData, "Health check successful")
}

// Ping handles the ping endpoint for simple connectivity check
// @Summary Ping endpoint
// @Description Simple ping endpoint for connectivity check
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /ping [get]
func (hc *HealthController) Ping(c *gin.Context) {
	pongData := gin.H{
		"message":   "pong",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Log ping request
	logger.Debug("Ping requested",
		logger.String("client_ip", c.ClientIP()),
	)

	response.SuccessJSONWithLog(c, pongData, "Pong response")
}

// startTime tracks when the application started
var startTime = time.Now()
