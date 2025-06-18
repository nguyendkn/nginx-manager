package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// HistoricalMetric stores time-series data for analytics
type HistoricalMetric struct {
	BaseModel
	Timestamp    time.Time              `gorm:"index:idx_timestamp" json:"timestamp"`
	MetricType   string                 `gorm:"index:idx_metric_type" json:"metric_type"`
	MetricName   string                 `gorm:"index:idx_metric_name" json:"metric_name"`
	Value        float64                `json:"value"`
	Tags         map[string]interface{} `gorm:"type:jsonb" json:"tags"`
	Source       string                 `json:"source"`    // system, nginx, proxy_host, certificate
	SourceID     *uint                  `json:"source_id"` // ID of related entity
	Unit         string                 `json:"unit"`      // bytes, percent, requests/sec, etc.
	Description  string                 `json:"description"`
	RetentionEnd *time.Time             `json:"retention_end"` // when this metric should be deleted
}

// AlertRule defines threshold-based alerting rules
type AlertRule struct {
	BaseModel
	Name                 string                 `gorm:"not null" json:"name"`
	Description          string                 `json:"description"`
	MetricType           string                 `gorm:"not null;index" json:"metric_type"`
	MetricName           string                 `gorm:"not null;index" json:"metric_name"`
	Condition            string                 `gorm:"not null" json:"condition"` // gt, lt, eq, ne, between
	Threshold            float64                `json:"threshold"`
	ThresholdMax         *float64               `json:"threshold_max"`            // for 'between' condition
	Severity             string                 `gorm:"not null" json:"severity"` // info, warning, critical
	IsEnabled            bool                   `gorm:"default:true" json:"is_enabled"`
	EvaluationWindow     int                    `gorm:"default:300" json:"evaluation_window"` // seconds
	NotificationChannels []NotificationChannel  `gorm:"many2many:alert_rule_channels;" json:"notification_channels"`
	Tags                 map[string]interface{} `gorm:"type:jsonb" json:"tags"`
	LastTriggered        *time.Time             `json:"last_triggered"`
	UserID               uint                   `gorm:"index" json:"user_id"`
	User                 User                   `json:"user,omitempty"`
}

// AlertInstance represents a triggered alert
type AlertInstance struct {
	BaseModel
	AlertRuleID       uint                   `gorm:"not null;index" json:"alert_rule_id"`
	AlertRule         AlertRule              `json:"alert_rule,omitempty"`
	TriggeredAt       time.Time              `gorm:"not null" json:"triggered_at"`
	ResolvedAt        *time.Time             `json:"resolved_at"`
	Status            string                 `gorm:"not null" json:"status"` // triggered, resolved, suppressed
	CurrentValue      float64                `json:"current_value"`
	ThresholdValue    float64                `json:"threshold_value"`
	Message           string                 `json:"message"`
	Context           map[string]interface{} `gorm:"type:jsonb" json:"context"`
	NotificationsSent int                    `gorm:"default:0" json:"notifications_sent"`
}

// NotificationChannel defines how alerts are delivered
type NotificationChannel struct {
	BaseModel
	Name          string                 `gorm:"not null" json:"name"`
	Type          string                 `gorm:"not null" json:"type"` // email, slack, webhook, teams
	IsEnabled     bool                   `gorm:"default:true" json:"is_enabled"`
	Configuration map[string]interface{} `gorm:"type:jsonb" json:"configuration"`
	UserID        uint                   `gorm:"index" json:"user_id"`
	User          User                   `json:"user,omitempty"`
}

// Dashboard represents a customizable analytics dashboard
type Dashboard struct {
	BaseModel
	Name        string                 `gorm:"not null" json:"name"`
	Description string                 `json:"description"`
	IsDefault   bool                   `gorm:"default:false" json:"is_default"`
	IsPublic    bool                   `gorm:"default:false" json:"is_public"`
	Layout      map[string]interface{} `gorm:"type:jsonb" json:"layout"`
	Widgets     []DashboardWidget      `json:"widgets"`
	UserID      uint                   `gorm:"index" json:"user_id"`
	User        User                   `json:"user,omitempty"`
	SharedWith  []User                 `gorm:"many2many:dashboard_shares;" json:"shared_with,omitempty"`
}

// DashboardWidget represents a widget on a dashboard
type DashboardWidget struct {
	BaseModel
	DashboardID     uint                   `gorm:"not null;index" json:"dashboard_id"`
	Dashboard       Dashboard              `json:"dashboard,omitempty"`
	Type            string                 `gorm:"not null" json:"type"` // chart, metric, table, gauge
	Title           string                 `gorm:"not null" json:"title"`
	Position        map[string]interface{} `gorm:"type:jsonb" json:"position"` // x, y, width, height
	Configuration   map[string]interface{} `gorm:"type:jsonb" json:"configuration"`
	DataSource      string                 `json:"data_source"`                        // metrics, logs, nginx_status
	Query           string                 `json:"query"`                              // metric query or filter
	RefreshInterval int                    `gorm:"default:30" json:"refresh_interval"` // seconds
	IsVisible       bool                   `gorm:"default:true" json:"is_visible"`
}

// PerformanceInsight represents analyzed performance data
type PerformanceInsight struct {
	BaseModel
	Type            string                 `gorm:"not null;index" json:"type"` // trend, anomaly, recommendation
	Severity        string                 `gorm:"not null" json:"severity"`   // info, warning, critical
	Title           string                 `gorm:"not null" json:"title"`
	Description     string                 `json:"description"`
	Category        string                 `gorm:"index" json:"category"` // performance, security, resources
	Source          string                 `json:"source"`                // system, nginx, certificate, proxy_host
	SourceID        *uint                  `json:"source_id"`
	Data            map[string]interface{} `gorm:"type:jsonb" json:"data"`
	Recommendations []string               `gorm:"type:jsonb" json:"recommendations"`
	IsResolved      bool                   `gorm:"default:false" json:"is_resolved"`
	ResolvedAt      *time.Time             `json:"resolved_at"`
	ViewedBy        []User                 `gorm:"many2many:insight_views;" json:"viewed_by,omitempty"`
}

// TrafficAnalytics stores aggregated traffic data
type TrafficAnalytics struct {
	BaseModel
	Timestamp       time.Time              `gorm:"index" json:"timestamp"`
	ProxyHostID     *uint                  `gorm:"index" json:"proxy_host_id"`
	ProxyHost       *ProxyHost             `json:"proxy_host,omitempty"`
	RequestCount    int64                  `json:"request_count"`
	BytesIn         int64                  `json:"bytes_in"`
	BytesOut        int64                  `json:"bytes_out"`
	AvgResponseTime float64                `json:"avg_response_time"`
	ErrorCount      int64                  `json:"error_count"`
	StatusCodes     map[string]interface{} `gorm:"type:jsonb" json:"status_codes"`
	Countries       map[string]interface{} `gorm:"type:jsonb" json:"countries"`
	UserAgents      map[string]interface{} `gorm:"type:jsonb" json:"user_agents"`
	Referrers       map[string]interface{} `gorm:"type:jsonb" json:"referrers"`
	TimeWindow      string                 `gorm:"index" json:"time_window"` // hour, day, week, month
}

// MetricAggregation stores pre-calculated aggregated metrics
type MetricAggregation struct {
	BaseModel
	MetricType   string                 `gorm:"not null;index" json:"metric_type"`
	MetricName   string                 `gorm:"not null;index" json:"metric_name"`
	TimeWindow   string                 `gorm:"not null;index" json:"time_window"` // 5m, 1h, 1d, 1w, 1M
	Timestamp    time.Time              `gorm:"index" json:"timestamp"`
	Count        int64                  `json:"count"`
	Sum          float64                `json:"sum"`
	Avg          float64                `json:"avg"`
	Min          float64                `json:"min"`
	Max          float64                `json:"max"`
	P50          float64                `json:"p50"`
	P95          float64                `json:"p95"`
	P99          float64                `json:"p99"`
	StdDev       float64                `json:"std_dev"`
	Tags         map[string]interface{} `gorm:"type:jsonb" json:"tags"`
	RetentionEnd *time.Time             `json:"retention_end"`
}

// Methods for HistoricalMetric
func (hm *HistoricalMetric) BeforeCreate(tx *gorm.DB) error {
	if hm.Timestamp.IsZero() {
		hm.Timestamp = time.Now()
	}
	return nil
}

func (hm *HistoricalMetric) SetRetention(duration time.Duration) {
	retentionEnd := time.Now().Add(duration)
	hm.RetentionEnd = &retentionEnd
}

// Methods for AlertRule
func (ar *AlertRule) BeforeCreate(tx *gorm.DB) error {
	if ar.EvaluationWindow == 0 {
		ar.EvaluationWindow = 300 // default 5 minutes
	}
	return nil
}

func (ar *AlertRule) EvaluateCondition(value float64) bool {
	switch ar.Condition {
	case "gt":
		return value > ar.Threshold
	case "lt":
		return value < ar.Threshold
	case "eq":
		return value == ar.Threshold
	case "ne":
		return value != ar.Threshold
	case "between":
		if ar.ThresholdMax == nil {
			return false
		}
		return value >= ar.Threshold && value <= *ar.ThresholdMax
	default:
		return false
	}
}

// Methods for Dashboard
func (d *Dashboard) BeforeCreate(tx *gorm.DB) error {
	// No specific logic needed for Dashboard creation
	return nil
}

func (d *Dashboard) MarshalJSON() ([]byte, error) {
	type Alias Dashboard
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias:     (*Alias)(d),
		CreatedAt: d.BaseModel.CreatedAt.Format(time.RFC3339),
		UpdatedAt: d.BaseModel.UpdatedAt.Format(time.RFC3339),
	})
}

// Methods for TrafficAnalytics
func (ta *TrafficAnalytics) BeforeCreate(tx *gorm.DB) error {
	if ta.Timestamp.IsZero() {
		ta.Timestamp = time.Now()
	}
	return nil
}

func (ta *TrafficAnalytics) GetErrorRate() float64 {
	if ta.RequestCount == 0 {
		return 0
	}
	return float64(ta.ErrorCount) / float64(ta.RequestCount) * 100
}

// Methods for MetricAggregation
func (ma *MetricAggregation) BeforeCreate(tx *gorm.DB) error {
	if ma.Timestamp.IsZero() {
		ma.Timestamp = time.Now()
	}
	return nil
}

func (ma *MetricAggregation) SetRetention(duration time.Duration) {
	retentionEnd := time.Now().Add(duration)
	ma.RetentionEnd = &retentionEnd
}
