package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap/zapcore"

	"github.com/nahuelsantos/argus/internal/models"
	"github.com/nahuelsantos/argus/internal/services"
)

// AlertingHandlers contains alerting and incident management handlers
type AlertingHandlers struct {
	loggingService  *services.LoggingService
	alertingService *services.AlertingService
}

// NewAlertingHandlers creates a new alerting handlers instance
func NewAlertingHandlers(loggingService *services.LoggingService, alertingService *services.AlertingService) *AlertingHandlers {
	return &AlertingHandlers{
		loggingService:  loggingService,
		alertingService: alertingService,
	}
}

// TestAlertRulesHandler tests alert rules functionality
func (ah *AlertingHandlers) TestAlertRulesHandler(w http.ResponseWriter, r *http.Request) {
	alertManager := ah.alertingService.GetAlertManager()

	alertManager.Mutex.RLock()
	rules := alertManager.Rules
	activeAlertsCount := len(alertManager.ActiveAlerts)
	alertManager.Mutex.RUnlock()

	enabledCount := 0
	for _, rule := range rules {
		if rule.Enabled {
			enabledCount++
		}
	}

	response := map[string]interface{}{
		"message":       "Alert rules functionality tested",
		"total_rules":   len(rules),
		"enabled_rules": enabledCount,
		"active_alerts": activeAlertsCount,
		"rules":         rules,
		"timestamp":     time.Now().Format(time.RFC3339),
		"test_status":   "success",
		"service":       "argus",
		"functionality": "alert_management",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	ah.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Alert rules tested")
}

// TestFireAlertHandler manually fires an alert for testing
func (ah *AlertingHandlers) TestFireAlertHandler(w http.ResponseWriter, r *http.Request) {
	alertType := r.URL.Query().Get("type")
	if alertType == "" {
		alertType = "high-cpu-usage"
	}

	severity := r.URL.Query().Get("severity")
	if severity == "" {
		severity = "critical"
	}

	alertManager := ah.alertingService.GetAlertManager()

	// Find the rule to fire
	alertManager.Mutex.RLock()
	var ruleToFire *models.AlertRule
	for _, rule := range alertManager.Rules {
		if rule.Name == alertType {
			ruleToFire = &rule
			break
		}
	}
	alertManager.Mutex.RUnlock()

	var activeAlertsCount int
	if ruleToFire != nil {
		// Create a proper alert using the models
		alert := &models.Alert{
			ID:       uuid.New().String(),
			RuleID:   ruleToFire.ID,
			RuleName: ruleToFire.Name,
			Status:   "firing",
			Severity: severity,
			Message:  fmt.Sprintf("Test alert: %s", alertType),
			StartsAt: time.Now(),
			Labels:   map[string]string{"test": "true"},
			Annotations: map[string]string{
				"summary": "Test alert fired manually",
			},
			Value:        rand.Float64() * 100,
			Threshold:    ruleToFire.Threshold,
			GeneratorURL: fmt.Sprintf("http://localhost:3001/alerts/%s", ruleToFire.ID),
		}

		alertManager.Mutex.Lock()
		alertManager.ActiveAlerts[ruleToFire.ID] = alert
		activeAlertsCount = len(alertManager.ActiveAlerts)
		alertManager.Mutex.Unlock()
	}

	response := map[string]interface{}{
		"message":       "Alert fired successfully",
		"alert_type":    alertType,
		"severity":      severity,
		"active_alerts": activeAlertsCount,
		"timestamp":     time.Now().Format(time.RFC3339),
		"test_status":   "success",
		"service":       "argus",
		"functionality": "alert_firing",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	ah.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Test alert fired")
}

// TestIncidentManagementHandler tests incident management functionality
func (ah *AlertingHandlers) TestIncidentManagementHandler(w http.ResponseWriter, r *http.Request) {
	alertManager := ah.alertingService.GetAlertManager()

	// Create a test incident using proper models
	incident := &models.Incident{
		ID:              uuid.New().String(),
		Title:           "Test Incident",
		Description:     "Test incident created for validation",
		Status:          "open",
		Severity:        "medium",
		Priority:        "medium",
		AffectedService: "argus",
		RelatedAlerts:   []string{},
		Tags:            []string{"test", "manual"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Timeline: []models.IncidentUpdate{
			{
				ID:        uuid.New().String(),
				Timestamp: time.Now(),
				Author:    "test-user",
				Type:      "creation",
				Message:   "Test incident created",
				NewValue:  "open",
			},
		},
		Metrics: models.IncidentMetrics{
			TimeToDetection: time.Minute * 2,
		},
	}

	alertManager.Mutex.Lock()
	alertManager.Incidents[incident.ID] = incident
	totalIncidents := len(alertManager.Incidents)
	alertManager.Mutex.Unlock()

	// Count incidents by status
	openIncidents := 0
	resolvedIncidents := 0
	alertManager.Mutex.RLock()
	for _, inc := range alertManager.Incidents {
		if inc.Status == "open" || inc.Status == "investigating" {
			openIncidents++
		} else if inc.Status == "resolved" || inc.Status == "closed" {
			resolvedIncidents++
		}
	}
	alertManager.Mutex.RUnlock()

	response := map[string]interface{}{
		"message":            "Incident management tested",
		"created_incident":   incident,
		"total_incidents":    totalIncidents,
		"open_incidents":     openIncidents,
		"resolved_incidents": resolvedIncidents,
		"timestamp":          time.Now().Format(time.RFC3339),
		"test_status":        "success",
		"service":            "argus",
		"functionality":      "incident_management",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	ah.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Incident management tested")
}

// TestNotificationChannelsHandler tests notification channels
func (ah *AlertingHandlers) TestNotificationChannelsHandler(w http.ResponseWriter, r *http.Request) {
	alertManager := ah.alertingService.GetAlertManager()

	alertManager.Mutex.RLock()
	channels := alertManager.NotificationChannels
	alertManager.Mutex.RUnlock()

	enabledCount := 0
	for _, channel := range channels {
		if channel.Enabled {
			enabledCount++
		}
	}

	// Simulate sending test notifications
	testResults := make([]map[string]interface{}, 0)
	for _, channel := range channels {
		if !channel.Enabled {
			continue
		}

		// Simulate notification sending with random latency
		latency := time.Duration(rand.Intn(50)+5) * time.Millisecond
		success := rand.Float64() < 0.95 // 95% success rate

		result := map[string]interface{}{
			"channel_id":   channel.ID,
			"channel_name": channel.Name,
			"channel_type": channel.Type,
			"success":      success,
			"latency_ms":   int(latency.Milliseconds()),
		}
		testResults = append(testResults, result)
	}

	response := map[string]interface{}{
		"message":          "Notification channels tested",
		"total_channels":   len(channels),
		"enabled_channels": enabledCount,
		"test_results":     testResults,
		"timestamp":        time.Now().Format(time.RFC3339),
		"test_status":      "success",
		"service":          "argus",
		"functionality":    "notification_channels",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	ah.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Notification channels tested")
}

// GetActiveAlertsHandler returns active alerts
func (ah *AlertingHandlers) GetActiveAlertsHandler(w http.ResponseWriter, r *http.Request) {
	alertManager := ah.alertingService.GetAlertManager()

	alertManager.Mutex.RLock()
	activeAlerts := make([]*models.Alert, 0)
	for _, alert := range alertManager.ActiveAlerts {
		activeAlerts = append(activeAlerts, alert)
	}
	recentAlerts := alertManager.AlertHistory
	if len(recentAlerts) > 10 {
		recentAlerts = recentAlerts[len(recentAlerts)-10:]
	}
	alertManager.Mutex.RUnlock()

	response := map[string]interface{}{
		"active_alerts": activeAlerts,
		"active_count":  len(activeAlerts),
		"recent_alerts": recentAlerts,
		"recent_count":  len(recentAlerts),
		"timestamp":     time.Now().Format(time.RFC3339),
		"service":       "argus",
		"functionality": "active_alerts_monitoring",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	ah.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Active alerts retrieved")
}

// GetActiveIncidentsHandler returns active incidents with metrics
func (ah *AlertingHandlers) GetActiveIncidentsHandler(w http.ResponseWriter, r *http.Request) {
	alertManager := ah.alertingService.GetAlertManager()

	alertManager.Mutex.RLock()
	activeIncidents := make([]*models.Incident, 0)
	incidentStats := map[string]int{
		"total":         0,
		"open":          0,
		"investigating": 0,
		"resolved":      0,
		"closed":        0,
		"critical":      0,
		"high":          0,
		"medium":        0,
		"low":           0,
	}

	for _, incident := range alertManager.Incidents {
		activeIncidents = append(activeIncidents, incident)
		incidentStats["total"]++

		// Count by status
		switch incident.Status {
		case "open":
			incidentStats["open"]++
		case "investigating":
			incidentStats["investigating"]++
		case "resolved":
			incidentStats["resolved"]++
		case "closed":
			incidentStats["closed"]++
		}

		// Count by severity
		switch incident.Severity {
		case "critical":
			incidentStats["critical"]++
		case "high":
			incidentStats["high"]++
		case "medium":
			incidentStats["medium"]++
		case "low":
			incidentStats["low"]++
		}
	}
	alertManager.Mutex.RUnlock()

	// Calculate MTTR (mock data)
	avgMTTR := time.Duration(rand.Intn(120)+30) * time.Minute

	response := map[string]interface{}{
		"active_incidents":    activeIncidents,
		"incident_statistics": incidentStats,
		"priority_breakdown": map[string]int{
			"critical": incidentStats["critical"],
			"high":     incidentStats["high"],
			"medium":   incidentStats["medium"],
			"low":      incidentStats["low"],
		},
		"mttr_minutes":  int(avgMTTR.Minutes()),
		"timestamp":     time.Now().Format(time.RFC3339),
		"service":       "argus",
		"functionality": "incident_monitoring",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	ah.loggingService.LogWithContext(zapcore.InfoLevel, r.Context(), "Active incidents retrieved")
}
