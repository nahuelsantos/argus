package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/nahuelsantos/argus/internal/config"
	"github.com/nahuelsantos/argus/internal/metrics"
	"github.com/nahuelsantos/argus/internal/models"
)

// AlertingService handles all alerting operations
type AlertingService struct {
	config       *config.ServiceConfig
	alertManager *models.AlertManager
}

// NewAlertingService creates a new alerting service
func NewAlertingService() *AlertingService {
	return &AlertingService{
		config: config.GetServiceConfig(),
		alertManager: &models.AlertManager{
			Rules:                []models.AlertRule{},
			ActiveAlerts:         make(map[string]*models.Alert),
			AlertHistory:         []*models.Alert{},
			NotificationChannels: []models.NotificationChannel{},
			Incidents:            make(map[string]*models.Incident),
			SilencedRules:        make(map[string]time.Time),
		},
	}
}

// InitAlertManager initializes the alert manager with default rules and channels
func (as *AlertingService) InitAlertManager() {
	as.initDefaultAlertRules()
	as.initDefaultNotificationChannels()

	// Start background processes
	go as.alertEvaluationEngine()
	go as.notificationProcessor()
}

// InitDefaultAlertRules creates default alert rules
func (as *AlertingService) initDefaultAlertRules() {
	rules := []models.AlertRule{
		{
			ID:          uuid.New().String(),
			Name:        "high-cpu-usage",
			Description: "Alert when CPU usage exceeds 80%",
			Query:       "cpu_usage_percent > 80",
			Threshold: models.AlertThreshold{
				Operator: ">",
				Value:    80.0,
			},
			Severity: "warning",
			Duration: 5 * time.Minute,
			Labels:   map[string]string{"team": "infrastructure", "service": "system"},
			Annotations: map[string]string{
				"summary":     "High CPU usage detected",
				"description": "CPU usage has been above 80% for more than 5 minutes",
				"runbook":     "https://runbooks.example.com/high-cpu",
			},
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "high-memory-usage",
			Description: "Alert when memory usage exceeds 2GB",
			Query:       "memory_usage_bytes > 2147483648",
			Threshold: models.AlertThreshold{
				Operator: ">",
				Value:    2147483648, // 2GB
			},
			Severity: "warning",
			Duration: 3 * time.Minute,
			Labels:   map[string]string{"team": "infrastructure", "service": "system"},
			Annotations: map[string]string{
				"summary":     "High memory usage detected",
				"description": "Memory usage has exceeded 2GB",
			},
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "high-error-rate",
			Description: "Alert when error rate exceeds 5%",
			Query:       "error_rate_percent > 5",
			Threshold: models.AlertThreshold{
				Operator: ">",
				Value:    5.0,
			},
			Severity: "critical",
			Duration: 2 * time.Minute,
			Labels:   map[string]string{"team": "backend", "service": "api"},
			Annotations: map[string]string{
				"summary":     "High error rate detected",
				"description": "Error rate has exceeded 5% for more than 2 minutes",
				"impact":      "User experience degradation",
			},
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "low-throughput",
			Description: "Alert when request throughput drops below 10 RPS",
			Query:       "requests_per_second < 10",
			Threshold: models.AlertThreshold{
				Operator: "<",
				Value:    10.0,
			},
			Severity: "warning",
			Duration: 5 * time.Minute,
			Labels:   map[string]string{"team": "backend", "service": "api"},
			Annotations: map[string]string{
				"summary":     "Low throughput detected",
				"description": "Request throughput has dropped below 10 RPS",
			},
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	as.alertManager.Mutex.Lock()
	as.alertManager.Rules = rules
	as.alertManager.Mutex.Unlock()
}

// InitDefaultNotificationChannels creates default notification channels
func (as *AlertingService) initDefaultNotificationChannels() {
	channels := []models.NotificationChannel{
		{
			ID:   uuid.New().String(),
			Name: "slack-alerts",
			Type: "slack",
			Config: map[string]interface{}{
				"webhook_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
				"channel":     "#alerts",
				"username":    "AlertBot",
			},
			Conditions: map[string]interface{}{
				"severity": []string{"warning", "critical"},
			},
			RateLimit: models.RateLimit{
				MaxAlerts:   10,
				TimeWindow:  time.Hour,
				GroupingKey: "rule_name",
			},
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:   uuid.New().String(),
			Name: "email-critical",
			Type: "email",
			Config: map[string]interface{}{
				"smtp_server": "smtp.example.com:587",
				"username":    "alerts@example.com",
				"password":    "password",
				"to":          []string{"oncall@example.com", "team-lead@example.com"},
				"from":        "alerts@example.com",
			},
			Conditions: map[string]interface{}{
				"severity": []string{"critical"},
			},
			RateLimit: models.RateLimit{
				MaxAlerts:   5,
				TimeWindow:  30 * time.Minute,
				GroupingKey: "severity",
			},
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:   uuid.New().String(),
			Name: "webhook-integration",
			Type: "webhook",
			Config: map[string]interface{}{
				"url":     "https://api.example.com/webhooks/alerts",
				"method":  "POST",
				"headers": map[string]string{"Authorization": "Bearer token123"},
			},
			Conditions: map[string]interface{}{
				"severity": []string{"warning", "critical"},
			},
			RateLimit: models.RateLimit{
				MaxAlerts:   20,
				TimeWindow:  time.Hour,
				GroupingKey: "service",
			},
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	as.alertManager.Mutex.Lock()
	as.alertManager.NotificationChannels = channels
	as.alertManager.Mutex.Unlock()
}

// AlertEvaluationEngine runs the alert evaluation loop
func (as *AlertingService) alertEvaluationEngine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		as.evaluateAlertRules()
	}
}

// EvaluateAlertRules evaluates all alert rules
func (as *AlertingService) evaluateAlertRules() {
	as.alertManager.Mutex.RLock()
	rules := as.alertManager.Rules
	as.alertManager.Mutex.RUnlock()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		if as.evaluateRule(&rule) {
			as.fireAlert(&rule)
		}
	}
}

// EvaluateRule evaluates a single alert rule
func (as *AlertingService) evaluateRule(rule *models.AlertRule) bool {
	// Simulate metric evaluation
	var currentValue float64

	switch rule.Name {
	case "high-cpu-usage":
		currentValue = float64(rand.Intn(100))
	case "high-memory-usage":
		currentValue = float64(rand.Intn(4) * 1024 * 1024 * 1024) // 0-4GB
	case "high-error-rate":
		currentValue = float64(rand.Intn(20))
	case "low-throughput":
		currentValue = float64(rand.Intn(50))
	default:
		currentValue = float64(rand.Intn(100))
	}

	switch rule.Threshold.Operator {
	case ">":
		return currentValue > rule.Threshold.Value
	case "<":
		return currentValue < rule.Threshold.Value
	case ">=":
		return currentValue >= rule.Threshold.Value
	case "<=":
		return currentValue <= rule.Threshold.Value
	case "==":
		return currentValue == rule.Threshold.Value
	default:
		return false
	}
}

// FireAlert fires an alert
func (as *AlertingService) fireAlert(rule *models.AlertRule) {
	// Check if alert already exists first (without lock)
	as.alertManager.Mutex.RLock()
	_, exists := as.alertManager.ActiveAlerts[rule.ID]
	as.alertManager.Mutex.RUnlock()

	if exists {
		return
	}

	alert := &models.Alert{
		ID:           uuid.New().String(),
		RuleID:       rule.ID,
		RuleName:     rule.Name,
		Status:       "firing",
		Severity:     rule.Severity,
		Message:      fmt.Sprintf("Alert: %s - %s", rule.Name, rule.Description),
		StartsAt:     time.Now(),
		Labels:       rule.Labels,
		Annotations:  rule.Annotations,
		Value:        rand.Float64() * 100,
		Threshold:    rule.Threshold,
		GeneratorURL: fmt.Sprintf("http://localhost:3001/alerts/%s", rule.ID),
	}

	// Add to active alerts and history
	as.alertManager.Mutex.Lock()
	// Double-check after acquiring lock
	if _, exists := as.alertManager.ActiveAlerts[rule.ID]; exists {
		as.alertManager.Mutex.Unlock()
		return
	}
	as.alertManager.ActiveAlerts[rule.ID] = alert
	as.alertManager.AlertHistory = append(as.alertManager.AlertHistory, alert)
	as.alertManager.Mutex.Unlock()

	// Send notification (no locks here)
	as.sendNotificationAsync(alert)

	// Create incident for critical alerts (separate lock)
	if alert.Severity == "critical" {
		as.createIncidentAsync(alert)
	}

	// Update metrics (no locks)
	metrics.AlertsTotal.WithLabelValues(rule.Name, rule.Severity, "firing").Inc()
}

// SendNotificationAsync sends notifications for an alert without holding locks
func (as *AlertingService) sendNotificationAsync(alert *models.Alert) {
	// Get channels snapshot
	as.alertManager.Mutex.RLock()
	channels := make([]models.NotificationChannel, len(as.alertManager.NotificationChannels))
	copy(channels, as.alertManager.NotificationChannels)
	as.alertManager.Mutex.RUnlock()

	for _, channel := range channels {
		if !channel.Enabled {
			continue
		}

		// Check conditions
		if conditions, ok := channel.Conditions["severity"].([]string); ok {
			found := false
			for _, severity := range conditions {
				if severity == alert.Severity {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Simulate notification sending
		success := as.simulateNotificationSend(&channel, alert)

		status := "success"
		if !success {
			status = "failed"
		}

		metrics.NotificationsSent.WithLabelValues(channel.Type, alert.Severity, status).Inc()

		// Simulate latency
		latency := time.Duration(rand.Intn(50)+5) * time.Millisecond
		metrics.NotificationLatency.WithLabelValues(channel.Type).Observe(latency.Seconds())
	}
}

// SimulateNotificationSend simulates sending a notification
func (as *AlertingService) simulateNotificationSend(channel *models.NotificationChannel, alert *models.Alert) bool {
	// Simulate 95% success rate
	return rand.Float64() < 0.95
}

// CreateIncidentAsync creates an incident from a critical alert without holding main lock
func (as *AlertingService) createIncidentAsync(alert *models.Alert) {
	incident := &models.Incident{
		ID:              uuid.New().String(),
		Title:           fmt.Sprintf("Critical Alert: %s", alert.RuleName),
		Description:     fmt.Sprintf("Incident created from critical alert: %s", alert.Message),
		Status:          "open",
		Severity:        alert.Severity,
		Priority:        "high",
		AffectedService: as.config.Name,
		RelatedAlerts:   []string{alert.ID},
		Tags:            []string{"auto-generated", "critical"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Timeline: []models.IncidentUpdate{
			{
				ID:        uuid.New().String(),
				Timestamp: time.Now(),
				Author:    "system",
				Type:      "creation",
				Message:   "Incident automatically created from critical alert",
				NewValue:  "open",
			},
		},
		Metrics: models.IncidentMetrics{
			TimeToDetection: time.Since(alert.StartsAt),
		},
	}

	as.alertManager.Mutex.Lock()
	as.alertManager.Incidents[incident.ID] = incident
	as.alertManager.Mutex.Unlock()

	// Update metrics
	metrics.IncidentsTotal.WithLabelValues(incident.Severity, incident.Status, incident.AffectedService).Inc()
}

// NotificationProcessor processes notification queue
func (as *AlertingService) notificationProcessor() {
	// This would process a notification queue in a real implementation
	// For now, it just updates health metrics
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		metrics.AlertManagerHealth.WithLabelValues("notification_processor").Set(1)
		metrics.AlertManagerHealth.WithLabelValues("alert_evaluator").Set(1)
		metrics.AlertManagerHealth.WithLabelValues("incident_manager").Set(1)
	}
}

// GetAlertManager returns the alert manager instance
func (as *AlertingService) GetAlertManager() *models.AlertManager {
	return as.alertManager
}
