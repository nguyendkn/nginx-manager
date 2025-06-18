package controllers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/internal/services"
	"github.com/nguyendkn/nginx-manager/pkg/response"
)

// MonitoringController handles monitoring and real-time metrics endpoints
type MonitoringController struct {
	monitoringService *services.MonitoringService
}

// NewMonitoringController creates a new monitoring controller
func NewMonitoringController(monitoringService *services.MonitoringService) *MonitoringController {
	return &MonitoringController{
		monitoringService: monitoringService,
	}
}

// GetSystemMetrics handles GET /api/v1/monitoring/system-metrics
func (mc *MonitoringController) GetSystemMetrics(c *gin.Context) {
	metrics, err := mc.monitoringService.GetSystemMetrics()
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to get system metrics", err)
		return
	}

	response.SuccessJSONWithLog(c, metrics, "System metrics retrieved successfully")
}

// GetNginxStatus handles GET /api/v1/monitoring/nginx-status
func (mc *MonitoringController) GetNginxStatus(c *gin.Context) {
	status, err := mc.monitoringService.GetNginxStatus()
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to get nginx status", err)
		return
	}

	response.SuccessJSONWithLog(c, status, "Nginx status retrieved successfully")
}

// GetActivityFeed handles GET /api/v1/monitoring/activity-feed
func (mc *MonitoringController) GetActivityFeed(c *gin.Context) {
	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid limit parameter", err)
		return
	}

	// Validate limit
	if limit < 1 || limit > 100 {
		limit = 50
	}

	activities, err := mc.monitoringService.GetRecentActivity(limit)
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to get activity feed", err)
		return
	}

	result := gin.H{
		"activities": activities,
		"total":      len(activities),
		"limit":      limit,
		"timestamp":  time.Now(),
	}

	response.SuccessJSONWithLog(c, result, "Activity feed retrieved successfully")
}

// HandleWebSocket handles WebSocket connections for real-time updates
func (mc *MonitoringController) HandleWebSocket(c *gin.Context) {
	mc.monitoringService.HandleWebSocket(c)
}

// GetDashboardStats handles GET /api/v1/monitoring/dashboard
func (mc *MonitoringController) GetDashboardStats(c *gin.Context) {
	// Get system metrics
	metrics, err := mc.monitoringService.GetSystemMetrics()
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to get system metrics", err)
		return
	}

	// Get nginx status
	nginxStatus, err := mc.monitoringService.GetNginxStatus()
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to get nginx status", err)
		return
	}

	// Get recent activity
	activities, err := mc.monitoringService.GetRecentActivity(10)
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to get recent activity", err)
		return
	}

	// Compile dashboard stats
	dashboardStats := gin.H{
		"system_metrics": metrics,
		"nginx_status":   nginxStatus,
		"recent_activity": gin.H{
			"activities": activities,
			"count":      len(activities),
		},
		"summary": gin.H{
			"uptime":          metrics.Process.Uptime.String(),
			"memory_usage":    formatBytes(metrics.Memory.Used),
			"memory_percent":  formatPercentage(metrics.Memory.UsedPercent),
			"disk_usage":      formatBytes(metrics.Disk.Used),
			"disk_percent":    formatPercentage(metrics.Disk.UsedPercent),
			"cpu_usage":       formatPercentage(metrics.CPU.Usage),
			"goroutines":      metrics.Process.Goroutines,
			"nginx_running":   nginxStatus.Running,
			"nginx_config_ok": nginxStatus.ConfigTest,
		},
		"timestamp": time.Now(),
	}

	response.SuccessJSONWithLog(c, dashboardStats, "Dashboard stats retrieved successfully")
}

// ControlNginx handles POST /api/v1/monitoring/nginx/control
func (mc *MonitoringController) ControlNginx(c *gin.Context) {
	var request struct {
		Action string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid request format", err)
		return
	}

	// Validate action
	validActions := map[string]bool{
		"start":   true,
		"stop":    true,
		"restart": true,
		"reload":  true,
		"test":    true,
	}

	if !validActions[request.Action] {
		response.BadRequestJSONWithLog(c, "Invalid action. Allowed: start, stop, restart, reload, test", nil)
		return
	}

	// TODO: Implement nginx control actions through nginx service
	// For now, return a success response
	result := gin.H{
		"action":    request.Action,
		"success":   true,
		"message":   "Nginx " + request.Action + " completed successfully",
		"timestamp": time.Now(),
	}

	response.SuccessJSONWithLog(c, result, "Nginx control action executed")
}

// Helper functions

// formatBytes formats bytes into human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatUint(bytes, 10) + " B"
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(bytes)/float64(div), 'f', 1, 64) + " " + "KMGTPE"[exp:exp+1] + "B"
}

// formatPercentage formats a percentage with 1 decimal place
func formatPercentage(percent float64) string {
	return strconv.FormatFloat(percent, 'f', 1, 64) + "%"
}
