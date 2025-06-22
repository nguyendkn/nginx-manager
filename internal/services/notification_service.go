package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/nguyendkn/nginx-manager/internal/models"
	"github.com/nguyendkn/nginx-manager/pkg/logger"
)

// NotificationService handles alert notifications via multiple channels
type NotificationService struct {
	emailTemplates map[string]*template.Template
	httpClient     *http.Client
}

// EmailConfig represents email configuration
type EmailConfig struct {
	SMTPHost    string   `json:"smtp_host"`
	SMTPPort    int      `json:"smtp_port"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	FromAddress string   `json:"from_address"`
	FromName    string   `json:"from_name"`
	ToAddresses []string `json:"to_addresses"`
	UseTLS      bool     `json:"use_tls"`
}

// SlackConfig represents Slack webhook configuration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	IconEmoji  string `json:"icon_emoji"`
}

// WebhookConfig represents generic webhook configuration
type WebhookConfig struct {
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	Timeout    int               `json:"timeout"`
	RetryCount int               `json:"retry_count"`
}

// TeamsConfig represents Microsoft Teams webhook configuration
type TeamsConfig struct {
	WebhookURL string `json:"webhook_url"`
	Title      string `json:"title"`
	ThemeColor string `json:"theme_color"`
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	ns := &NotificationService{
		emailTemplates: make(map[string]*template.Template),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Initialize email templates
	ns.initializeEmailTemplates()

	return ns
}

// SendAlert sends an alert notification through the specified channel
func (ns *NotificationService) SendAlert(channel models.NotificationChannel, alert *models.AlertInstance, rule *models.AlertRule) error {
	switch channel.Type {
	case "email":
		return ns.sendEmailAlert(channel, alert, rule)
	case "slack":
		return ns.sendSlackAlert(channel, alert, rule)
	case "webhook":
		return ns.sendWebhookAlert(channel, alert, rule)
	case "teams":
		return ns.sendTeamsAlert(channel, alert, rule)
	default:
		return fmt.Errorf("unsupported notification channel type: %s", channel.Type)
	}
}

// sendEmailAlert sends an alert via email
func (ns *NotificationService) sendEmailAlert(channel models.NotificationChannel, alert *models.AlertInstance, rule *models.AlertRule) error {
	var emailConfig EmailConfig
	if err := ns.parseConfig(channel.Configuration, &emailConfig); err != nil {
		return fmt.Errorf("invalid email configuration: %v", err)
	}

	// Generate email content
	subject := fmt.Sprintf("[%s] %s Alert: %s",
		strings.ToUpper(rule.Severity), "Nginx Manager", rule.Name)

	body, err := ns.generateEmailBody(alert, rule)
	if err != nil {
		return fmt.Errorf("failed to generate email body: %v", err)
	}

	// Send email
	return ns.sendEmail(emailConfig, subject, body)
}

// sendSlackAlert sends an alert to Slack
func (ns *NotificationService) sendSlackAlert(channel models.NotificationChannel, alert *models.AlertInstance, rule *models.AlertRule) error {
	webhookURL, ok := channel.Configuration["webhook_url"].(string)
	if !ok {
		return fmt.Errorf("missing webhook_url in Slack configuration")
	}

	payload := map[string]interface{}{
		"text": fmt.Sprintf("🚨 *%s Alert: %s*\n%s\nCurrent Value: %.2f\nThreshold: %.2f",
			strings.Title(rule.Severity), rule.Name, alert.Message, alert.CurrentValue, alert.ThresholdValue),
	}

	return ns.sendWebhookRequest(webhookURL, payload)
}

// sendWebhookAlert sends an alert via generic webhook
func (ns *NotificationService) sendWebhookAlert(channel models.NotificationChannel, alert *models.AlertInstance, rule *models.AlertRule) error {
	url, ok := channel.Configuration["url"].(string)
	if !ok {
		return fmt.Errorf("missing url in webhook configuration")
	}

	payload := map[string]interface{}{
		"alert_id":      alert.ID,
		"rule_name":     rule.Name,
		"severity":      rule.Severity,
		"message":       alert.Message,
		"current_value": alert.CurrentValue,
		"threshold":     alert.ThresholdValue,
		"triggered_at":  alert.TriggeredAt,
	}

	return ns.sendWebhookRequest(url, payload)
}

// sendTeamsAlert sends an alert to Microsoft Teams
func (ns *NotificationService) sendTeamsAlert(channel models.NotificationChannel, alert *models.AlertInstance, rule *models.AlertRule) error {
	var teamsConfig TeamsConfig
	if err := ns.parseConfig(channel.Configuration, &teamsConfig); err != nil {
		return fmt.Errorf("invalid Teams configuration: %v", err)
	}

	// Create Teams message payload
	payload := map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": ns.getTeamsSeverityColor(rule.Severity),
		"summary":    fmt.Sprintf("%s Alert: %s", strings.Title(rule.Severity), rule.Name),
		"sections": []map[string]interface{}{
			{
				"activityTitle":    fmt.Sprintf("%s Alert", strings.Title(rule.Severity)),
				"activitySubtitle": rule.Name,
				"activityImage":    "https://nginx.org/favicon.ico",
				"text":             alert.Message,
				"facts": []map[string]interface{}{
					{
						"name":  "Metric",
						"value": rule.MetricName,
					},
					{
						"name":  "Current Value",
						"value": fmt.Sprintf("%.2f", alert.CurrentValue),
					},
					{
						"name":  "Threshold",
						"value": fmt.Sprintf("%.2f", alert.ThresholdValue),
					},
					{
						"name":  "Triggered At",
						"value": alert.TriggeredAt.Format("2006-01-02 15:04:05"),
					},
				},
			},
		},
	}

	return ns.sendWebhookRequest(teamsConfig.WebhookURL, payload)
}

// sendEmail sends an email using SMTP
func (ns *NotificationService) sendEmail(config EmailConfig, subject, body string) error {
	// Create message
	message := fmt.Sprintf("From: %s <%s>\r\n", config.FromName, config.FromAddress)
	message += fmt.Sprintf("To: %s\r\n", strings.Join(config.ToAddresses, ","))
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n"
	message += body

	// Setup authentication
	auth := smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)

	// Send email
	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
	return smtp.SendMail(addr, auth, config.FromAddress, config.ToAddresses, []byte(message))
}

// sendWebhookRequest sends a webhook request
func (ns *NotificationService) sendWebhookRequest(url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := ns.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook request failed with status: %d", resp.StatusCode)
	}

	logger.Info("Alert notification sent successfully",
		logger.String("url", url),
		logger.Int("status_code", resp.StatusCode))

	return nil
}

// generateEmailBody generates HTML email body for alerts
func (ns *NotificationService) generateEmailBody(alert *models.AlertInstance, rule *models.AlertRule) (string, error) {
	tmpl, exists := ns.emailTemplates["alert"]
	if !exists {
		return "", fmt.Errorf("alert email template not found")
	}

	data := struct {
		Alert *models.AlertInstance
		Rule  *models.AlertRule
	}{
		Alert: alert,
		Rule:  rule,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// initializeEmailTemplates initializes email templates
func (ns *NotificationService) initializeEmailTemplates() {
	alertTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        .header { background-color: {{ .SeverityColor }}; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .metric-info { background-color: #f8f9fa; padding: 15px; border-radius: 4px; margin: 15px 0; }
        .footer { background-color: #6c757d; color: white; padding: 15px; text-align: center; font-size: 12px; }
        .severity-{{ .Rule.Severity }} { border-left: 4px solid {{ .SeverityColor }}; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>🚨 {{ .Rule.Severity | title }} Alert</h2>
            <h3>{{ .Rule.Name }}</h3>
        </div>
        <div class="content">
            <p><strong>Alert Message:</strong></p>
            <p>{{ .Alert.Message }}</p>

            <div class="metric-info">
                <h4>Metric Details</h4>
                <p><strong>Metric:</strong> {{ .Rule.MetricName }}</p>
                <p><strong>Current Value:</strong> {{ printf "%.2f" .Alert.CurrentValue }}</p>
                <p><strong>Threshold:</strong> {{ printf "%.2f" .Alert.ThresholdValue }}</p>
                <p><strong>Condition:</strong> {{ .Rule.Condition }}</p>
                <p><strong>Triggered At:</strong> {{ .Alert.TriggeredAt.Format "2006-01-02 15:04:05" }}</p>
            </div>

            <p><em>This alert was generated by Nginx Manager monitoring system.</em></p>
        </div>
        <div class="footer">
            Nginx Manager Alert System
        </div>
    </div>
</body>
</html>
`

	tmpl, err := template.New("alert").Parse(alertTemplate)
	if err != nil {
		logger.Error("Failed to parse alert email template", logger.Err(err))
		return
	}

	ns.emailTemplates["alert"] = tmpl
}

// Helper functions
func (ns *NotificationService) parseConfig(config map[string]interface{}, target interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func (ns *NotificationService) getTeamsSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return "FF0000"
	case "warning":
		return "FFA500"
	case "info":
		return "008000"
	default:
		return "FFA500"
	}
}
