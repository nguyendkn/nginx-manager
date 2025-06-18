package controllers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nguyendkn/nginx-manager/internal/models"
	"github.com/nguyendkn/nginx-manager/internal/services"
	"github.com/nguyendkn/nginx-manager/pkg/response"
)

// AnalyticsController handles analytics and historical data endpoints
type AnalyticsController struct {
	analyticsService *services.AnalyticsService
}

// NewAnalyticsController creates a new analytics controller
func NewAnalyticsController(analyticsService *services.AnalyticsService) *AnalyticsController {
	return &AnalyticsController{
		analyticsService: analyticsService,
	}
}

// QueryMetrics handles POST /api/v1/analytics/metrics/query
func (ac *AnalyticsController) QueryMetrics(c *gin.Context) {
	var query services.MetricQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid query parameters", err)
		return
	}

	// Validate time range
	if query.TimeRange.Start.IsZero() || query.TimeRange.End.IsZero() {
		response.BadRequestJSONWithLog(c, "Start and end time are required", nil)
		return
	}

	// Validate time range is not too large (max 90 days)
	if query.TimeRange.End.Sub(query.TimeRange.Start) > 90*24*time.Hour {
		response.BadRequestJSONWithLog(c, "Time range cannot exceed 90 days", nil)
		return
	}

	dataPoints, err := ac.analyticsService.QueryMetrics(query)
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to query metrics", err)
		return
	}

	result := gin.H{
		"data_points": dataPoints,
		"count":       len(dataPoints),
		"query":       query,
		"timestamp":   time.Now(),
	}

	response.SuccessJSONWithLog(c, result, "Metrics queried successfully")
}

// GetHistoricalMetrics handles GET /api/v1/analytics/metrics/{type}/{name}
func (ac *AnalyticsController) GetHistoricalMetrics(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	// Parse query parameters
	startTime := c.Query("start")
	endTime := c.Query("end")
	aggregation := c.DefaultQuery("aggregation", "avg")
	groupBy := c.Query("group_by")
	limitStr := c.DefaultQuery("limit", "1000")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid limit parameter", err)
		return
	}

	// Parse time range
	var timeRange services.TimeRange
	if startTime != "" && endTime != "" {
		start, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			response.BadRequestJSONWithLog(c, "Invalid start time format", err)
			return
		}
		end, err := time.Parse(time.RFC3339, endTime)
		if err != nil {
			response.BadRequestJSONWithLog(c, "Invalid end time format", err)
			return
		}
		timeRange = services.TimeRange{Start: start, End: end}
	} else {
		// Default to last 24 hours
		timeRange = services.TimeRange{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
		}
	}

	query := services.MetricQuery{
		MetricType:  metricType,
		MetricName:  metricName,
		TimeRange:   timeRange,
		Aggregation: aggregation,
		GroupBy:     groupBy,
		Limit:       limit,
	}

	dataPoints, err := ac.analyticsService.QueryMetrics(query)
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to query historical metrics", err)
		return
	}

	result := gin.H{
		"metric_type": metricType,
		"metric_name": metricName,
		"data_points": dataPoints,
		"count":       len(dataPoints),
		"time_range":  timeRange,
		"aggregation": aggregation,
		"timestamp":   time.Now(),
	}

	response.SuccessJSONWithLog(c, result, "Historical metrics retrieved successfully")
}

// GetSystemMetricsSummary handles GET /api/v1/analytics/system/summary
func (ac *AnalyticsController) GetSystemMetricsSummary(c *gin.Context) {
	// Parse time range
	rangeStr := c.DefaultQuery("range", "24h")

	var duration time.Duration
	switch rangeStr {
	case "1h":
		duration = time.Hour
	case "24h":
		duration = 24 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	default:
		duration = 24 * time.Hour
	}

	timeRange := services.TimeRange{
		Start: time.Now().Add(-duration),
		End:   time.Now(),
	}

	// Generate comprehensive system metrics summary
	summary := gin.H{
		"time_range": timeRange,
		"metrics": gin.H{
			"cpu":    ac.getMetricSummary("system", "cpu_usage", timeRange),
			"memory": ac.getMetricSummary("system", "memory_usage", timeRange),
			"disk":   ac.getMetricSummary("system", "disk_usage", timeRange),
		},
		"timestamp": time.Now(),
	}

	response.SuccessJSONWithLog(c, summary, "System metrics summary retrieved successfully")
}

// CreateAlertRule handles POST /api/v1/analytics/alerts/rules
func (ac *AnalyticsController) CreateAlertRule(c *gin.Context) {
	var alertRule models.AlertRule
	if err := c.ShouldBindJSON(&alertRule); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid alert rule data", err)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}
	alertRule.UserID = userID.(uint)

	// Validate alert rule
	if alertRule.Name == "" || alertRule.MetricType == "" || alertRule.MetricName == "" {
		response.BadRequestJSONWithLog(c, "Name, metric type, and metric name are required", nil)
		return
	}

	if alertRule.Condition == "" || alertRule.Severity == "" {
		response.BadRequestJSONWithLog(c, "Condition and severity are required", nil)
		return
	}

	// Create the alert rule using database operations
	if err := ac.analyticsService.CreateAlertRule(&alertRule); err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to create alert rule", err)
		return
	}

	response.SuccessJSONWithLog(c, alertRule, "Alert rule created successfully")
}

// GetAlertRules handles GET /api/v1/analytics/alerts/rules
func (ac *AnalyticsController) GetAlertRules(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	alertRules, err := ac.analyticsService.GetAlertRules(userID.(uint))
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to get alert rules", err)
		return
	}

	result := gin.H{
		"alert_rules": alertRules,
		"count":       len(alertRules),
		"timestamp":   time.Now(),
	}

	response.SuccessJSONWithLog(c, result, "Alert rules retrieved successfully")
}

// UpdateAlertRule handles PUT /api/v1/analytics/alerts/rules/{id}
func (ac *AnalyticsController) UpdateAlertRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid alert rule ID", err)
		return
	}

	var alertRule models.AlertRule
	if err := c.ShouldBindJSON(&alertRule); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid alert rule data", err)
		return
	}

	alertRule.ID = uint(id)

	// Verify ownership
	userID, exists := c.Get("user_id")
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	if err := ac.analyticsService.UpdateAlertRule(&alertRule, userID.(uint)); err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to update alert rule", err)
		return
	}

	response.SuccessJSONWithLog(c, alertRule, "Alert rule updated successfully")
}

// DeleteAlertRule handles DELETE /api/v1/analytics/alerts/rules/{id}
func (ac *AnalyticsController) DeleteAlertRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid alert rule ID", err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	if err := ac.analyticsService.DeleteAlertRule(uint(id), userID.(uint)); err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to delete alert rule", err)
		return
	}

	response.SuccessJSONWithLog(c, gin.H{"id": id}, "Alert rule deleted successfully")
}

// GetAlertInstances handles GET /api/v1/analytics/alerts/instances
func (ac *AnalyticsController) GetAlertInstances(c *gin.Context) {
	// Parse query parameters
	status := c.Query("status")
	severity := c.Query("severity")
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid limit parameter", err)
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid offset parameter", err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	instances, total, err := ac.analyticsService.GetAlertInstances(userID.(uint), status, severity, limit, offset)
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to get alert instances", err)
		return
	}

	result := gin.H{
		"alert_instances": instances,
		"total":           total,
		"limit":           limit,
		"offset":          offset,
		"timestamp":       time.Now(),
	}

	response.SuccessJSONWithLog(c, result, "Alert instances retrieved successfully")
}

// CreateDashboard handles POST /api/v1/analytics/dashboards
func (ac *AnalyticsController) CreateDashboard(c *gin.Context) {
	var dashboard models.Dashboard
	if err := c.ShouldBindJSON(&dashboard); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid dashboard data", err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}
	dashboard.UserID = userID.(uint)

	if dashboard.Name == "" {
		response.BadRequestJSONWithLog(c, "Dashboard name is required", nil)
		return
	}

	if err := ac.analyticsService.CreateDashboard(&dashboard); err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to create dashboard", err)
		return
	}

	response.SuccessJSONWithLog(c, dashboard, "Dashboard created successfully")
}

// GetDashboards handles GET /api/v1/analytics/dashboards
func (ac *AnalyticsController) GetDashboards(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	dashboards, err := ac.analyticsService.GetDashboards(userID.(uint))
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to get dashboards", err)
		return
	}

	result := gin.H{
		"dashboards": dashboards,
		"count":      len(dashboards),
		"timestamp":  time.Now(),
	}

	response.SuccessJSONWithLog(c, result, "Dashboards retrieved successfully")
}

// GetDashboard handles GET /api/v1/analytics/dashboards/{id}
func (ac *AnalyticsController) GetDashboard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid dashboard ID", err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	dashboard, err := ac.analyticsService.GetDashboard(uint(id), userID.(uint))
	if err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to get dashboard", err)
		return
	}

	response.SuccessJSONWithLog(c, dashboard, "Dashboard retrieved successfully")
}

// UpdateDashboard handles PUT /api/v1/analytics/dashboards/{id}
func (ac *AnalyticsController) UpdateDashboard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid dashboard ID", err)
		return
	}

	var dashboard models.Dashboard
	if err := c.ShouldBindJSON(&dashboard); err != nil {
		response.BadRequestJSONWithLog(c, "Invalid dashboard data", err)
		return
	}

	dashboard.ID = uint(id)

	userID, exists := c.Get("user_id")
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	if err := ac.analyticsService.UpdateDashboard(&dashboard, userID.(uint)); err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to update dashboard", err)
		return
	}

	response.SuccessJSONWithLog(c, dashboard, "Dashboard updated successfully")
}

// DeleteDashboard handles DELETE /api/v1/analytics/dashboards/{id}
func (ac *AnalyticsController) DeleteDashboard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequestJSONWithLog(c, "Invalid dashboard ID", err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		response.UnauthorizedJSONWithLog(c, "User not authenticated")
		return
	}

	if err := ac.analyticsService.DeleteDashboard(uint(id), userID.(uint)); err != nil {
		response.InternalServerErrorJSONWithLog(c, "Failed to delete dashboard", err)
		return
	}

	response.SuccessJSONWithLog(c, gin.H{"id": id}, "Dashboard deleted successfully")
}

// Helper method to get metric summary
func (ac *AnalyticsController) getMetricSummary(metricType, metricName string, timeRange services.TimeRange) gin.H {
	query := services.MetricQuery{
		MetricType:  metricType,
		MetricName:  metricName,
		TimeRange:   timeRange,
		Aggregation: "avg",
		GroupBy:     "1h",
		Limit:       200,
	}

	dataPoints, err := ac.analyticsService.QueryMetrics(query)
	if err != nil || len(dataPoints) == 0 {
		return gin.H{
			"current":     0,
			"average":     0,
			"peak":        0,
			"trend":       "unknown",
			"data_points": 0,
		}
	}

	// Calculate basic statistics
	var sum, min, max float64
	min = dataPoints[0].Value
	max = dataPoints[0].Value

	for _, point := range dataPoints {
		sum += point.Value
		if point.Value < min {
			min = point.Value
		}
		if point.Value > max {
			max = point.Value
		}
	}

	average := sum / float64(len(dataPoints))
	current := dataPoints[len(dataPoints)-1].Value

	// Simple trend calculation
	trend := "stable"
	if len(dataPoints) > 1 {
		firstHalf := dataPoints[:len(dataPoints)/2]
		secondHalf := dataPoints[len(dataPoints)/2:]

		var firstSum, secondSum float64
		for _, point := range firstHalf {
			firstSum += point.Value
		}
		for _, point := range secondHalf {
			secondSum += point.Value
		}

		firstAvg := firstSum / float64(len(firstHalf))
		secondAvg := secondSum / float64(len(secondHalf))

		if secondAvg > firstAvg*1.1 {
			trend = "increasing"
		} else if secondAvg < firstAvg*0.9 {
			trend = "decreasing"
		}
	}

	return gin.H{
		"current":     current,
		"average":     average,
		"peak":        max,
		"minimum":     min,
		"trend":       trend,
		"data_points": len(dataPoints),
	}
}
