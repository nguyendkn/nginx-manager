package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/nguyendkn/nginx-manager/internal/models"
	"github.com/nguyendkn/nginx-manager/pkg/logger"
	"gorm.io/gorm"
)

// AnalyticsService handles historical data, alerting, and performance insights
type AnalyticsService struct {
	db                  *gorm.DB
	monitoringService   *MonitoringService
	notificationService *NotificationService
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// MetricQuery represents a query for historical metrics
type MetricQuery struct {
	MetricType  string            `json:"metric_type"`
	MetricName  string            `json:"metric_name"`
	TimeRange   TimeRange         `json:"time_range"`
	Aggregation string            `json:"aggregation"` // avg, sum, min, max, p50, p95, p99
	GroupBy     string            `json:"group_by"`    // time window: 5m, 1h, 1d, 1w
	Tags        map[string]string `json:"tags"`
	Limit       int               `json:"limit"`
}

// MetricDataPoint represents a single metric data point
type MetricDataPoint struct {
	Timestamp time.Time   `json:"timestamp"`
	Value     float64     `json:"value"`
	Tags      interface{} `json:"tags,omitempty"`
}

// TrendAnalysis represents trend analysis results
type TrendAnalysis struct {
	MetricName    string            `json:"metric_name"`
	TimeRange     TimeRange         `json:"time_range"`
	Trend         string            `json:"trend"` // increasing, decreasing, stable
	ChangePercent float64           `json:"change_percent"`
	Confidence    float64           `json:"confidence"` // 0-100
	Anomalies     []Anomaly         `json:"anomalies"`
	Forecast      []MetricDataPoint `json:"forecast,omitempty"`
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	Timestamp   time.Time `json:"timestamp"`
	Value       float64   `json:"value"`
	ExpectedMin float64   `json:"expected_min"`
	ExpectedMax float64   `json:"expected_max"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
}

// PerformanceReport represents a comprehensive performance analysis
type PerformanceReport struct {
	TimeRange        TimeRange                   `json:"time_range"`
	SystemHealth     SystemHealthScore           `json:"system_health"`
	ResourceAnalysis ResourceAnalysis            `json:"resource_analysis"`
	TrafficInsights  TrafficInsights             `json:"traffic_insights"`
	Alerts           []models.AlertInstance      `json:"alerts"`
	Recommendations  []models.PerformanceInsight `json:"recommendations"`
	GeneratedAt      time.Time                   `json:"generated_at"`
}

// SystemHealthScore represents overall system health
type SystemHealthScore struct {
	Overall     float64            `json:"overall"`    // 0-100
	Components  map[string]float64 `json:"components"` // nginx, database, disk, memory, cpu
	Trend       string             `json:"trend"`      // improving, declining, stable
	LastChecked time.Time          `json:"last_checked"`
}

// ResourceAnalysis represents resource usage analysis
type ResourceAnalysis struct {
	CPU     ResourceMetric `json:"cpu"`
	Memory  ResourceMetric `json:"memory"`
	Disk    ResourceMetric `json:"disk"`
	Network NetworkMetric  `json:"network"`
}

// ResourceMetric represents analysis of a specific resource
type ResourceMetric struct {
	Current       float64   `json:"current"`
	Average       float64   `json:"average"`
	Peak          float64   `json:"peak"`
	Trend         string    `json:"trend"`
	PredictedPeak float64   `json:"predicted_peak"`
	AlertLevel    string    `json:"alert_level"` // normal, warning, critical
	LastUpdated   time.Time `json:"last_updated"`
}

// NetworkMetric represents network analysis
type NetworkMetric struct {
	ResourceMetric
	BytesInRate  float64 `json:"bytes_in_rate"`
	BytesOutRate float64 `json:"bytes_out_rate"`
	PacketLoss   float64 `json:"packet_loss"`
	Latency      float64 `json:"latency"`
}

// TrafficInsights represents traffic analysis
type TrafficInsights struct {
	TotalRequests   int64               `json:"total_requests"`
	AvgResponseTime float64             `json:"avg_response_time"`
	ErrorRate       float64             `json:"error_rate"`
	TopEndpoints    []EndpointStats     `json:"top_endpoints"`
	GeographicData  map[string]int64    `json:"geographic_data"`
	UserAgentStats  map[string]int64    `json:"user_agent_stats"`
	StatusCodeDist  map[string]int64    `json:"status_code_distribution"`
	TrafficTrends   []TrafficTrendPoint `json:"traffic_trends"`
}

// EndpointStats represents statistics for a specific endpoint
type EndpointStats struct {
	ProxyHostID      uint    `json:"proxy_host_id"`
	Domain           string  `json:"domain"`
	RequestCount     int64   `json:"request_count"`
	AvgResponseTime  float64 `json:"avg_response_time"`
	ErrorRate        float64 `json:"error_rate"`
	BytesTransferred int64   `json:"bytes_transferred"`
}

// TrafficTrendPoint represents a point in traffic trend data
type TrafficTrendPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	RequestCount int64     `json:"request_count"`
	ResponseTime float64   `json:"response_time"`
	ErrorRate    float64   `json:"error_rate"`
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *gorm.DB, monitoringService *MonitoringService, notificationService *NotificationService) *AnalyticsService {
	return &AnalyticsService{
		db:                  db,
		monitoringService:   monitoringService,
		notificationService: notificationService,
	}
}

// StoreMetric stores a historical metric
func (as *AnalyticsService) StoreMetric(metric *models.HistoricalMetric) error {
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	// Set default retention (1 year for raw metrics)
	if metric.RetentionEnd == nil {
		metric.SetRetention(365 * 24 * time.Hour)
	}

	if err := as.db.Create(metric).Error; err != nil {
		logger.Error("Failed to store metric", logger.Err(err))
		return err
	}

	// Check if this metric triggers any alerts
	go as.checkAlerts(metric)

	// Create aggregations asynchronously
	go as.createAggregations(metric)

	return nil
}

// StoreSystemMetrics stores current system metrics as historical data
func (as *AnalyticsService) StoreSystemMetrics() error {
	metrics, err := as.monitoringService.GetSystemMetrics()
	if err != nil {
		return err
	}

	timestamp := time.Now()

	// Store CPU metrics
	cpuMetrics := []*models.HistoricalMetric{
		{
			Timestamp:   timestamp,
			MetricType:  "system",
			MetricName:  "cpu_usage",
			Value:       metrics.CPU.Usage,
			Unit:        "percent",
			Source:      "system",
			Description: "CPU usage percentage",
		},
		{
			Timestamp:   timestamp,
			MetricType:  "system",
			MetricName:  "load_avg_1",
			Value:       metrics.CPU.LoadAvg1,
			Unit:        "load",
			Source:      "system",
			Description: "1-minute load average",
		},
	}

	// Store Memory metrics
	memoryMetrics := []*models.HistoricalMetric{
		{
			Timestamp:   timestamp,
			MetricType:  "system",
			MetricName:  "memory_usage",
			Value:       metrics.Memory.UsedPercent,
			Unit:        "percent",
			Source:      "system",
			Description: "Memory usage percentage",
		},
		{
			Timestamp:   timestamp,
			MetricType:  "system",
			MetricName:  "memory_used_bytes",
			Value:       float64(metrics.Memory.Used),
			Unit:        "bytes",
			Source:      "system",
			Description: "Memory used in bytes",
		},
	}

	// Store Disk metrics
	diskMetrics := []*models.HistoricalMetric{
		{
			Timestamp:   timestamp,
			MetricType:  "system",
			MetricName:  "disk_usage",
			Value:       metrics.Disk.UsedPercent,
			Unit:        "percent",
			Source:      "system",
			Description: "Disk usage percentage",
		},
		{
			Timestamp:   timestamp,
			MetricType:  "system",
			MetricName:  "disk_used_bytes",
			Value:       float64(metrics.Disk.Used),
			Unit:        "bytes",
			Source:      "system",
			Description: "Disk used in bytes",
		},
	}

	// Store Process metrics
	processMetrics := []*models.HistoricalMetric{
		{
			Timestamp:   timestamp,
			MetricType:  "system",
			MetricName:  "goroutines",
			Value:       float64(metrics.Process.Goroutines),
			Unit:        "count",
			Source:      "system",
			Description: "Number of Go routines",
		},
	}

	// Combine all metrics
	allMetrics := append(cpuMetrics, memoryMetrics...)
	allMetrics = append(allMetrics, diskMetrics...)
	allMetrics = append(allMetrics, processMetrics...)

	// Store all metrics
	for _, metric := range allMetrics {
		if err := as.StoreMetric(metric); err != nil {
			logger.Error("Failed to store system metric",
				logger.String("metric_name", metric.MetricName),
				logger.Err(err))
		}
	}

	return nil
}

// QueryMetrics queries historical metrics with aggregation
func (as *AnalyticsService) QueryMetrics(query MetricQuery) ([]MetricDataPoint, error) {
	if query.Limit == 0 {
		query.Limit = 1000
	}

	db := as.db.Model(&models.HistoricalMetric{}).
		Where("metric_type = ? AND metric_name = ?", query.MetricType, query.MetricName).
		Where("timestamp BETWEEN ? AND ?", query.TimeRange.Start, query.TimeRange.End)

	// Apply tag filters
	for key, value := range query.Tags {
		db = db.Where("tags ->> ? = ?", key, value)
	}

	var metrics []models.HistoricalMetric

	if query.GroupBy != "" {
		// Use aggregated data if available
		return as.queryAggregatedMetrics(query)
	}

	// Query raw metrics
	if err := db.Order("timestamp ASC").Limit(query.Limit).Find(&metrics).Error; err != nil {
		return nil, err
	}

	// Convert to data points
	dataPoints := make([]MetricDataPoint, len(metrics))
	for i, metric := range metrics {
		dataPoints[i] = MetricDataPoint{
			Timestamp: metric.Timestamp,
			Value:     metric.Value,
			Tags:      metric.Tags,
		}
	}

	return dataPoints, nil
}

// queryAggregatedMetrics queries pre-calculated aggregated metrics
func (as *AnalyticsService) queryAggregatedMetrics(query MetricQuery) ([]MetricDataPoint, error) {
	var aggregations []models.MetricAggregation

	db := as.db.Model(&models.MetricAggregation{}).
		Where("metric_type = ? AND metric_name = ? AND time_window = ?",
			query.MetricType, query.MetricName, query.GroupBy).
		Where("timestamp BETWEEN ? AND ?", query.TimeRange.Start, query.TimeRange.End)

	if err := db.Order("timestamp ASC").Limit(query.Limit).Find(&aggregations).Error; err != nil {
		return nil, err
	}

	dataPoints := make([]MetricDataPoint, len(aggregations))
	for i, agg := range aggregations {
		var value float64
		switch query.Aggregation {
		case "avg":
			value = agg.Avg
		case "sum":
			value = agg.Sum
		case "min":
			value = agg.Min
		case "max":
			value = agg.Max
		case "p50":
			value = agg.P50
		case "p95":
			value = agg.P95
		case "p99":
			value = agg.P99
		default:
			value = agg.Avg
		}

		dataPoints[i] = MetricDataPoint{
			Timestamp: agg.Timestamp,
			Value:     value,
			Tags:      agg.Tags,
		}
	}

	return dataPoints, nil
}

// AnalyzeTrends performs trend analysis on metrics
func (as *AnalyticsService) AnalyzeTrends(metricType, metricName string, timeRange TimeRange) (*TrendAnalysis, error) {
	query := MetricQuery{
		MetricType: metricType,
		MetricName: metricName,
		TimeRange:  timeRange,
		GroupBy:    "1h", // Hourly aggregation for trend analysis
		Limit:      1000,
	}

	dataPoints, err := as.QueryMetrics(query)
	if err != nil {
		return nil, err
	}

	if len(dataPoints) < 2 {
		return &TrendAnalysis{
			MetricName: metricName,
			TimeRange:  timeRange,
			Trend:      "insufficient_data",
		}, nil
	}

	// Calculate trend
	trend := as.calculateTrend(dataPoints)

	// Detect anomalies
	anomalies := as.detectAnomalies(dataPoints)

	return &TrendAnalysis{
		MetricName:    metricName,
		TimeRange:     timeRange,
		Trend:         trend.Direction,
		ChangePercent: trend.ChangePercent,
		Confidence:    trend.Confidence,
		Anomalies:     anomalies,
	}, nil
}

// calculateTrend calculates trend direction and confidence
func (as *AnalyticsService) calculateTrend(dataPoints []MetricDataPoint) struct {
	Direction     string
	ChangePercent float64
	Confidence    float64
} {
	if len(dataPoints) < 2 {
		return struct {
			Direction     string
			ChangePercent float64
			Confidence    float64
		}{"unknown", 0, 0}
	}

	// Simple linear regression
	n := float64(len(dataPoints))
	var sumX, sumY, sumXY, sumX2 float64

	for i, point := range dataPoints {
		x := float64(i)
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Calculate change percentage
	firstValue := dataPoints[0].Value
	lastValue := dataPoints[len(dataPoints)-1].Value
	changePercent := 0.0
	if firstValue != 0 {
		changePercent = ((lastValue - firstValue) / firstValue) * 100
	}

	// Determine trend direction
	direction := "stable"
	if math.Abs(slope) > 0.1 { // Threshold for significance
		if slope > 0 {
			direction = "increasing"
		} else {
			direction = "decreasing"
		}
	}

	// Calculate confidence based on correlation
	confidence := math.Min(math.Abs(slope)*100, 100)

	return struct {
		Direction     string
		ChangePercent float64
		Confidence    float64
	}{direction, changePercent, confidence}
}

// detectAnomalies detects anomalies in metric data
func (as *AnalyticsService) detectAnomalies(dataPoints []MetricDataPoint) []Anomaly {
	if len(dataPoints) < 10 {
		return []Anomaly{}
	}

	// Calculate moving average and standard deviation
	windowSize := 10
	var anomalies []Anomaly

	for i := windowSize; i < len(dataPoints); i++ {
		window := dataPoints[i-windowSize : i]

		// Calculate mean and std dev for window
		var sum, sumSq float64
		for _, point := range window {
			sum += point.Value
			sumSq += point.Value * point.Value
		}

		mean := sum / float64(len(window))
		variance := (sumSq / float64(len(window))) - (mean * mean)
		stdDev := math.Sqrt(variance)

		// Check if current point is anomalous (> 2 standard deviations)
		currentValue := dataPoints[i].Value
		expectedMin := mean - 2*stdDev
		expectedMax := mean + 2*stdDev

		if currentValue < expectedMin || currentValue > expectedMax {
			severity := "warning"
			if currentValue < mean-3*stdDev || currentValue > mean+3*stdDev {
				severity = "critical"
			}

			anomalies = append(anomalies, Anomaly{
				Timestamp:   dataPoints[i].Timestamp,
				Value:       currentValue,
				ExpectedMin: expectedMin,
				ExpectedMax: expectedMax,
				Severity:    severity,
				Description: fmt.Sprintf("Value %.2f outside expected range [%.2f, %.2f]",
					currentValue, expectedMin, expectedMax),
			})
		}
	}

	return anomalies
}

// checkAlerts checks if a metric triggers any alert rules
func (as *AnalyticsService) checkAlerts(metric *models.HistoricalMetric) {
	var alertRules []models.AlertRule

	err := as.db.Where("metric_type = ? AND metric_name = ? AND is_enabled = ?",
		metric.MetricType, metric.MetricName, true).Find(&alertRules).Error
	if err != nil {
		logger.Error("Failed to query alert rules", logger.Err(err))
		return
	}

	for _, rule := range alertRules {
		if rule.EvaluateCondition(metric.Value) {
			// Create alert instance
			alertInstance := &models.AlertInstance{
				AlertRuleID:    rule.ID,
				TriggeredAt:    metric.Timestamp,
				Status:         "triggered",
				CurrentValue:   metric.Value,
				ThresholdValue: rule.Threshold,
				Message: fmt.Sprintf("Alert '%s' triggered: %s value %.2f %s threshold %.2f",
					rule.Name, metric.MetricName, metric.Value, rule.Condition, rule.Threshold),
				Context: map[string]interface{}{
					"metric_type": metric.MetricType,
					"metric_name": metric.MetricName,
					"source":      metric.Source,
					"source_id":   metric.SourceID,
					"tags":        metric.Tags,
				},
			}

			if err := as.db.Create(alertInstance).Error; err != nil {
				logger.Error("Failed to create alert instance", logger.Err(err))
				continue
			}

			// Update rule's last triggered time
			now := time.Now()
			rule.LastTriggered = &now
			as.db.Save(&rule)

			// Send notifications
			go as.sendAlertNotifications(alertInstance, &rule)
		}
	}
}

// sendAlertNotifications sends notifications for an alert
func (as *AnalyticsService) sendAlertNotifications(alert *models.AlertInstance, rule *models.AlertRule) {
	if as.notificationService == nil {
		logger.Warn("Notification service not available")
		return
	}

	// Load notification channels
	var channels []models.NotificationChannel
	if err := as.db.Model(rule).Association("NotificationChannels").Find(&channels); err != nil {
		logger.Error("Failed to load notification channels", logger.Err(err))
		return
	}

	for _, channel := range channels {
		if !channel.IsEnabled {
			continue
		}

		err := as.notificationService.SendAlert(channel, alert, rule)
		if err != nil {
			logger.Error("Failed to send alert notification",
				logger.String("channel", channel.Name),
				logger.Err(err))
		} else {
			alert.NotificationsSent++
		}
	}

	// Update alert instance
	as.db.Save(alert)
}

// createAggregations creates time-window aggregations for a metric
func (as *AnalyticsService) createAggregations(metric *models.HistoricalMetric) {
	timeWindows := []string{"5m", "1h", "1d", "1w"}

	for _, window := range timeWindows {
		as.createAggregation(metric, window)
	}
}

// createAggregation creates aggregation for a specific time window
func (as *AnalyticsService) createAggregation(metric *models.HistoricalMetric, timeWindow string) {
	windowStart := as.getWindowStart(metric.Timestamp, timeWindow)
	windowEnd := as.getWindowEnd(windowStart, timeWindow)

	// Check if aggregation already exists
	var existingAgg models.MetricAggregation
	err := as.db.Where("metric_type = ? AND metric_name = ? AND time_window = ? AND timestamp = ?",
		metric.MetricType, metric.MetricName, timeWindow, windowStart).First(&existingAgg).Error

	if err == gorm.ErrRecordNotFound {
		// Create new aggregation
		agg := &models.MetricAggregation{
			MetricType: metric.MetricType,
			MetricName: metric.MetricName,
			TimeWindow: timeWindow,
			Timestamp:  windowStart,
		}

		// Calculate aggregation values
		as.calculateAggregationValues(agg, windowStart, windowEnd)

		// Set retention (longer for aggregated data)
		retentionDuration := as.getRetentionForWindow(timeWindow)
		agg.SetRetention(retentionDuration)

		as.db.Create(agg)
	} else if err == nil {
		// Update existing aggregation
		as.calculateAggregationValues(&existingAgg, windowStart, windowEnd)
		as.db.Save(&existingAgg)
	}
}

// getWindowStart calculates the start of a time window
func (as *AnalyticsService) getWindowStart(timestamp time.Time, window string) time.Time {
	switch window {
	case "5m":
		return timestamp.Truncate(5 * time.Minute)
	case "1h":
		return timestamp.Truncate(time.Hour)
	case "1d":
		return timestamp.Truncate(24 * time.Hour)
	case "1w":
		// Start of week (Monday)
		weekday := timestamp.Weekday()
		if weekday == 0 {
			weekday = 7 // Sunday = 7
		}
		return timestamp.AddDate(0, 0, -int(weekday-1)).Truncate(24 * time.Hour)
	default:
		return timestamp
	}
}

// getWindowEnd calculates the end of a time window
func (as *AnalyticsService) getWindowEnd(start time.Time, window string) time.Time {
	switch window {
	case "5m":
		return start.Add(5 * time.Minute)
	case "1h":
		return start.Add(time.Hour)
	case "1d":
		return start.Add(24 * time.Hour)
	case "1w":
		return start.Add(7 * 24 * time.Hour)
	default:
		return start.Add(time.Hour)
	}
}

// getRetentionForWindow returns appropriate retention duration for aggregation window
func (as *AnalyticsService) getRetentionForWindow(window string) time.Duration {
	switch window {
	case "5m":
		return 30 * 24 * time.Hour // 30 days
	case "1h":
		return 90 * 24 * time.Hour // 90 days
	case "1d":
		return 365 * 24 * time.Hour // 1 year
	case "1w":
		return 5 * 365 * 24 * time.Hour // 5 years
	default:
		return 365 * 24 * time.Hour
	}
}

// calculateAggregationValues calculates aggregation statistics
func (as *AnalyticsService) calculateAggregationValues(agg *models.MetricAggregation, start, end time.Time) {
	var metrics []models.HistoricalMetric

	err := as.db.Where("metric_type = ? AND metric_name = ? AND timestamp BETWEEN ? AND ?",
		agg.MetricType, agg.MetricName, start, end).Find(&metrics).Error
	if err != nil {
		logger.Error("Failed to query metrics for aggregation", logger.Err(err))
		return
	}

	if len(metrics) == 0 {
		return
	}

	// Calculate basic statistics
	values := make([]float64, len(metrics))
	var sum float64

	agg.Count = int64(len(metrics))
	agg.Min = math.Inf(1)
	agg.Max = math.Inf(-1)

	for i, metric := range metrics {
		value := metric.Value
		values[i] = value
		sum += value

		if value < agg.Min {
			agg.Min = value
		}
		if value > agg.Max {
			agg.Max = value
		}
	}

	agg.Sum = sum
	agg.Avg = sum / float64(len(metrics))

	// Calculate percentiles
	sort.Float64s(values)
	agg.P50 = as.percentile(values, 0.5)
	agg.P95 = as.percentile(values, 0.95)
	agg.P99 = as.percentile(values, 0.99)

	// Calculate standard deviation
	var variance float64
	for _, value := range values {
		variance += math.Pow(value-agg.Avg, 2)
	}
	agg.StdDev = math.Sqrt(variance / float64(len(values)))
}

// percentile calculates the percentile value from sorted slice
func (as *AnalyticsService) percentile(sortedValues []float64, p float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}

	index := p * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedValues[lower]
	}

	weight := index - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

// StartMetricsCollection starts automated metrics collection
func (as *AnalyticsService) StartMetricsCollection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info("Started metrics collection", logger.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping metrics collection")
			return
		case <-ticker.C:
			if err := as.StoreSystemMetrics(); err != nil {
				logger.Error("Failed to store system metrics", logger.Err(err))
			}
		}
	}
}

// CleanupExpiredMetrics removes expired metrics based on retention policies
func (as *AnalyticsService) CleanupExpiredMetrics() error {
	now := time.Now()

	// Clean up historical metrics
	result := as.db.Where("retention_end IS NOT NULL AND retention_end < ?", now).
		Delete(&models.HistoricalMetric{})
	if result.Error != nil {
		return result.Error
	}

	logger.Info("Cleaned up expired historical metrics",
		logger.Int64("deleted_count", result.RowsAffected))

	// Clean up metric aggregations
	result = as.db.Where("retention_end IS NOT NULL AND retention_end < ?", now).
		Delete(&models.MetricAggregation{})
	if result.Error != nil {
		return result.Error
	}

	logger.Info("Cleaned up expired metric aggregations",
		logger.Int64("deleted_count", result.RowsAffected))

	return nil
}

// CreateAlertRule creates a new alert rule
func (as *AnalyticsService) CreateAlertRule(alertRule *models.AlertRule) error {
	return as.db.Create(alertRule).Error
}

// GetAlertRules retrieves alert rules for a user
func (as *AnalyticsService) GetAlertRules(userID uint) ([]models.AlertRule, error) {
	var alertRules []models.AlertRule
	err := as.db.Where("user_id = ?", userID).
		Preload("NotificationChannels").
		Find(&alertRules).Error
	return alertRules, err
}

// UpdateAlertRule updates an existing alert rule
func (as *AnalyticsService) UpdateAlertRule(alertRule *models.AlertRule, userID uint) error {
	// Verify ownership
	var existingRule models.AlertRule
	if err := as.db.Where("id = ? AND user_id = ?", alertRule.ID, userID).First(&existingRule).Error; err != nil {
		return err
	}

	return as.db.Save(alertRule).Error
}

// DeleteAlertRule deletes an alert rule
func (as *AnalyticsService) DeleteAlertRule(ruleID, userID uint) error {
	return as.db.Where("id = ? AND user_id = ?", ruleID, userID).Delete(&models.AlertRule{}).Error
}

// GetAlertInstances retrieves alert instances with filtering and pagination
func (as *AnalyticsService) GetAlertInstances(userID uint, status, severity string, limit, offset int) ([]models.AlertInstance, int64, error) {
	query := as.db.Model(&models.AlertInstance{}).
		Joins("JOIN alert_rules ON alert_instances.alert_rule_id = alert_rules.id").
		Where("alert_rules.user_id = ?", userID).
		Preload("AlertRule")

	// Apply filters
	if status != "" {
		query = query.Where("alert_instances.status = ?", status)
	}
	if severity != "" {
		query = query.Where("alert_rules.severity = ?", severity)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get paginated results
	var instances []models.AlertInstance
	err := query.Order("alert_instances.triggered_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&instances).Error

	return instances, total, err
}

// CreateDashboard creates a new dashboard
func (as *AnalyticsService) CreateDashboard(dashboard *models.Dashboard) error {
	return as.db.Create(dashboard).Error
}

// GetDashboards retrieves dashboards for a user
func (as *AnalyticsService) GetDashboards(userID uint) ([]models.Dashboard, error) {
	var dashboards []models.Dashboard
	err := as.db.Where("user_id = ? OR is_public = ?", userID, true).
		Preload("Widgets").
		Find(&dashboards).Error
	return dashboards, err
}

// GetDashboard retrieves a specific dashboard
func (as *AnalyticsService) GetDashboard(dashboardID, userID uint) (*models.Dashboard, error) {
	var dashboard models.Dashboard
	err := as.db.Where("id = ? AND (user_id = ? OR is_public = ?)", dashboardID, userID, true).
		Preload("Widgets").
		First(&dashboard).Error
	return &dashboard, err
}

// UpdateDashboard updates an existing dashboard
func (as *AnalyticsService) UpdateDashboard(dashboard *models.Dashboard, userID uint) error {
	// Verify ownership
	var existingDashboard models.Dashboard
	if err := as.db.Where("id = ? AND user_id = ?", dashboard.ID, userID).First(&existingDashboard).Error; err != nil {
		return err
	}

	return as.db.Save(dashboard).Error
}

// DeleteDashboard deletes a dashboard
func (as *AnalyticsService) DeleteDashboard(dashboardID, userID uint) error {
	return as.db.Where("id = ? AND user_id = ?", dashboardID, userID).Delete(&models.Dashboard{}).Error
}
