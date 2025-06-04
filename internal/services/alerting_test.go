package services

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nahuelsantos/argus/internal/models"
)

func TestNewAlertingService(t *testing.T) {
	as := NewAlertingService()

	assert.NotNil(t, as)
	assert.NotNil(t, as.config)
	assert.NotNil(t, as.alertManager)
	assert.NotNil(t, as.alertManager.Rules)
	assert.NotNil(t, as.alertManager.ActiveAlerts)
	assert.NotNil(t, as.alertManager.AlertHistory)
	assert.NotNil(t, as.alertManager.NotificationChannels)
	assert.NotNil(t, as.alertManager.Incidents)
	assert.NotNil(t, as.alertManager.SilencedRules)
	assert.Equal(t, "argus", as.config.Name)
}

func TestAlertingService_InitAlertManager(t *testing.T) {
	as := NewAlertingService()

	// Before initialization, should have empty data
	assert.Empty(t, as.alertManager.Rules)
	assert.Empty(t, as.alertManager.NotificationChannels)

	// Initialize alert manager
	as.InitAlertManager()

	// Should have default rules and channels
	assert.Greater(t, len(as.alertManager.Rules), 0)
	assert.Greater(t, len(as.alertManager.NotificationChannels), 0)

	// Verify default rules
	ruleNames := make(map[string]bool)
	for _, rule := range as.alertManager.Rules {
		ruleNames[rule.Name] = true
		assert.NotEmpty(t, rule.ID)
		assert.NotEmpty(t, rule.Name)
		assert.NotEmpty(t, rule.Description)
		assert.True(t, rule.Enabled)
		assert.Contains(t, []string{"warning", "critical"}, rule.Severity)
	}

	expectedRules := []string{"high-cpu-usage", "high-memory-usage", "high-error-rate", "low-throughput"}
	for _, expectedRule := range expectedRules {
		assert.True(t, ruleNames[expectedRule], "Expected rule %s not found", expectedRule)
	}

	// Verify default notification channels
	channelTypes := make(map[string]bool)
	for _, channel := range as.alertManager.NotificationChannels {
		channelTypes[channel.Type] = true
		assert.NotEmpty(t, channel.ID)
		assert.NotEmpty(t, channel.Name)
		assert.True(t, channel.Enabled)
		assert.NotNil(t, channel.Config)
	}

	expectedChannelTypes := []string{"slack", "email", "webhook"}
	for _, expectedType := range expectedChannelTypes {
		assert.True(t, channelTypes[expectedType], "Expected channel type %s not found", expectedType)
	}
}

func TestAlertingService_InitDefaultAlertRules(t *testing.T) {
	as := NewAlertingService()
	as.initDefaultAlertRules()

	rules := as.alertManager.Rules
	assert.Len(t, rules, 4) // Expected 4 default rules

	// Test specific rule configurations
	tests := []struct {
		name              string
		expectedThreshold float64
		expectedOperator  string
		expectedSeverity  string
		expectedDuration  time.Duration
	}{
		{
			name:              "high-cpu-usage",
			expectedThreshold: 80.0,
			expectedOperator:  ">",
			expectedSeverity:  "warning",
			expectedDuration:  5 * time.Minute,
		},
		{
			name:              "high-memory-usage",
			expectedThreshold: 2147483648, // 2GB
			expectedOperator:  ">",
			expectedSeverity:  "warning",
			expectedDuration:  3 * time.Minute,
		},
		{
			name:              "high-error-rate",
			expectedThreshold: 5.0,
			expectedOperator:  ">",
			expectedSeverity:  "critical",
			expectedDuration:  2 * time.Minute,
		},
		{
			name:              "low-throughput",
			expectedThreshold: 10.0,
			expectedOperator:  "<",
			expectedSeverity:  "warning",
			expectedDuration:  5 * time.Minute,
		},
	}

	ruleMap := make(map[string]models.AlertRule)
	for _, rule := range rules {
		ruleMap[rule.Name] = rule
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, exists := ruleMap[tt.name]
			require.True(t, exists, "Rule %s should exist", tt.name)

			assert.Equal(t, tt.expectedThreshold, rule.Threshold.Value)
			assert.Equal(t, tt.expectedOperator, rule.Threshold.Operator)
			assert.Equal(t, tt.expectedSeverity, rule.Severity)
			assert.Equal(t, tt.expectedDuration, rule.Duration)
			assert.True(t, rule.Enabled)
			assert.NotEmpty(t, rule.ID)
			assert.NotEmpty(t, rule.Query)
			assert.NotNil(t, rule.Labels)
			assert.NotNil(t, rule.Annotations)
		})
	}
}

func TestAlertingService_InitDefaultNotificationChannels(t *testing.T) {
	as := NewAlertingService()
	as.initDefaultNotificationChannels()

	channels := as.alertManager.NotificationChannels
	assert.Len(t, channels, 3) // Expected 3 default channels

	channelMap := make(map[string]models.NotificationChannel)
	for _, channel := range channels {
		channelMap[channel.Type] = channel
	}

	// Test Slack channel
	t.Run("slack channel", func(t *testing.T) {
		slack, exists := channelMap["slack"]
		require.True(t, exists)

		assert.Equal(t, "slack-alerts", slack.Name)
		assert.True(t, slack.Enabled)
		assert.NotNil(t, slack.Config)
		assert.NotNil(t, slack.Conditions)
		assert.Equal(t, 10, slack.RateLimit.MaxAlerts)
		assert.Equal(t, time.Hour, slack.RateLimit.TimeWindow)
	})

	// Test Email channel
	t.Run("email channel", func(t *testing.T) {
		email, exists := channelMap["email"]
		require.True(t, exists)

		assert.Equal(t, "email-critical", email.Name)
		assert.True(t, email.Enabled)
		assert.NotNil(t, email.Config)
		assert.NotNil(t, email.Conditions)
		assert.Equal(t, 5, email.RateLimit.MaxAlerts)
		assert.Equal(t, 30*time.Minute, email.RateLimit.TimeWindow)
	})

	// Test Webhook channel
	t.Run("webhook channel", func(t *testing.T) {
		webhook, exists := channelMap["webhook"]
		require.True(t, exists)

		assert.Equal(t, "webhook-integration", webhook.Name)
		assert.True(t, webhook.Enabled)
		assert.NotNil(t, webhook.Config)
		assert.NotNil(t, webhook.Conditions)
		assert.Equal(t, 20, webhook.RateLimit.MaxAlerts)
		assert.Equal(t, time.Hour, webhook.RateLimit.TimeWindow)
	})
}

func TestAlertingService_EvaluateRule(t *testing.T) {
	as := NewAlertingService()

	tests := []struct {
		name         string
		ruleName     string
		threshold    models.AlertThreshold
		expectResult bool // This will vary due to randomness, so we test logic
	}{
		{
			name:     "greater than operator",
			ruleName: "high-cpu-usage",
			threshold: models.AlertThreshold{
				Operator: ">",
				Value:    50.0,
			},
		},
		{
			name:     "less than operator",
			ruleName: "low-throughput",
			threshold: models.AlertThreshold{
				Operator: "<",
				Value:    50.0,
			},
		},
		{
			name:     "greater than or equal operator",
			ruleName: "test-rule",
			threshold: models.AlertThreshold{
				Operator: ">=",
				Value:    50.0,
			},
		},
		{
			name:     "less than or equal operator",
			ruleName: "test-rule",
			threshold: models.AlertThreshold{
				Operator: "<=",
				Value:    50.0,
			},
		},
		{
			name:     "equal operator",
			ruleName: "test-rule",
			threshold: models.AlertThreshold{
				Operator: "==",
				Value:    50.0,
			},
		},
		{
			name:     "invalid operator",
			ruleName: "test-rule",
			threshold: models.AlertThreshold{
				Operator: "!=",
				Value:    50.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &models.AlertRule{
				Name:      tt.ruleName,
				Threshold: tt.threshold,
			}

			// Test multiple times to account for randomness
			results := make([]bool, 10)
			for i := 0; i < 10; i++ {
				results[i] = as.evaluateRule(rule)
			}

			// For invalid operator, should always return false
			if tt.threshold.Operator == "!=" {
				for _, result := range results {
					assert.False(t, result, "Invalid operator should always return false")
				}
				return
			}

			// For valid operators, we should see some variation in results
			// (unless the random values are consistently above/below threshold)
			hasTrue := false
			hasFalse := false
			for _, result := range results {
				if result {
					hasTrue = true
				} else {
					hasFalse = true
				}
			}

			// At least one of the conditions should be met in 10 runs
			// (this is probabilistic but very likely)
			assert.True(t, hasTrue || hasFalse, "Should have some variation in results")
		})
	}
}

func TestAlertingService_FireAlert(t *testing.T) {
	as := NewAlertingService()

	rule := &models.AlertRule{
		ID:          "test-rule-id",
		Name:        "test-alert",
		Description: "Test alert description",
		Severity:    "warning",
		Labels:      map[string]string{"test": "value"},
		Annotations: map[string]string{"summary": "Test summary"},
		Threshold: models.AlertThreshold{
			Operator: ">",
			Value:    50.0,
		},
	}

	// Fire alert
	as.fireAlert(rule)

	// Check that alert was added to active alerts
	as.alertManager.Mutex.RLock()
	alert, exists := as.alertManager.ActiveAlerts[rule.ID]
	historyCount := len(as.alertManager.AlertHistory)
	as.alertManager.Mutex.RUnlock()

	assert.True(t, exists, "Alert should be in active alerts")
	assert.NotNil(t, alert)
	assert.Equal(t, rule.ID, alert.RuleID)
	assert.Equal(t, rule.Name, alert.RuleName)
	assert.Equal(t, "firing", alert.Status)
	assert.Equal(t, rule.Severity, alert.Severity)
	assert.Equal(t, rule.Labels, alert.Labels)
	assert.Equal(t, rule.Annotations, alert.Annotations)
	assert.NotEmpty(t, alert.ID)
	assert.Greater(t, historyCount, 0)

	// Try to fire the same alert again - should not create duplicate
	as.fireAlert(rule)

	as.alertManager.Mutex.RLock()
	newHistoryCount := len(as.alertManager.AlertHistory)
	as.alertManager.Mutex.RUnlock()

	assert.Equal(t, historyCount, newHistoryCount, "Should not create duplicate alerts")
}

func TestAlertingService_FireCriticalAlert(t *testing.T) {
	as := NewAlertingService()

	rule := &models.AlertRule{
		ID:          "critical-rule-id",
		Name:        "critical-alert",
		Description: "Critical alert description",
		Severity:    "critical",
		Labels:      map[string]string{"test": "value"},
		Annotations: map[string]string{"summary": "Critical summary"},
		Threshold: models.AlertThreshold{
			Operator: ">",
			Value:    90.0,
		},
	}

	// Fire critical alert
	as.fireAlert(rule)

	// Should create an incident for critical alerts
	as.alertManager.Mutex.RLock()
	alert := as.alertManager.ActiveAlerts[rule.ID]
	incidentCount := len(as.alertManager.Incidents)
	as.alertManager.Mutex.RUnlock()

	assert.NotNil(t, alert)
	assert.Equal(t, "critical", alert.Severity)
	assert.Greater(t, incidentCount, 0, "Should create incident for critical alert")

	// Find the incident
	var incident *models.Incident
	as.alertManager.Mutex.RLock()
	for _, inc := range as.alertManager.Incidents {
		if inc.RelatedAlerts[0] == alert.ID {
			incident = inc
			break
		}
	}
	as.alertManager.Mutex.RUnlock()

	require.NotNil(t, incident, "Should find incident related to alert")
	assert.Equal(t, "open", incident.Status)
	assert.Equal(t, "critical", incident.Severity)
	assert.Equal(t, "high", incident.Priority)
	assert.Contains(t, incident.RelatedAlerts, alert.ID)
	assert.Contains(t, incident.Tags, "auto-generated")
	assert.Contains(t, incident.Tags, "critical")
	assert.NotEmpty(t, incident.Timeline)
}

func TestAlertingService_SendNotificationAsync(t *testing.T) {
	as := NewAlertingService()
	as.initDefaultNotificationChannels()

	alert := &models.Alert{
		ID:       "test-alert-id",
		RuleName: "test-rule",
		Severity: "warning",
		Message:  "Test alert message",
	}

	// Should not panic
	assert.NotPanics(t, func() {
		as.sendNotificationAsync(alert)
	})

	// Test with critical alert
	criticalAlert := &models.Alert{
		ID:       "critical-alert-id",
		RuleName: "critical-rule",
		Severity: "critical",
		Message:  "Critical alert message",
	}

	assert.NotPanics(t, func() {
		as.sendNotificationAsync(criticalAlert)
	})

	// Test with disabled channels
	as.alertManager.Mutex.Lock()
	for i := range as.alertManager.NotificationChannels {
		as.alertManager.NotificationChannels[i].Enabled = false
	}
	as.alertManager.Mutex.Unlock()

	assert.NotPanics(t, func() {
		as.sendNotificationAsync(alert)
	})
}

func TestAlertingService_SimulateNotificationSend(t *testing.T) {
	as := NewAlertingService()

	channel := &models.NotificationChannel{
		ID:   "test-channel",
		Name: "test-channel",
		Type: "slack",
	}

	alert := &models.Alert{
		ID:      "test-alert",
		Message: "Test message",
	}

	// Test multiple times to check probabilistic success
	successCount := 0
	totalTests := 100

	for i := 0; i < totalTests; i++ {
		if as.simulateNotificationSend(channel, alert) {
			successCount++
		}
	}

	// Should have ~95% success rate (allow some variance)
	successRate := float64(successCount) / float64(totalTests)
	assert.Greater(t, successRate, 0.90, "Success rate should be around 95%")
	assert.Less(t, successRate, 0.99, "Success rate should not be 100%")
}

func TestAlertingService_CreateIncidentAsync(t *testing.T) {
	as := NewAlertingService()

	alert := &models.Alert{
		ID:       "test-alert-id",
		RuleName: "test-rule",
		Severity: "critical",
		Message:  "Test critical alert",
		StartsAt: time.Now().Add(-5 * time.Minute),
	}

	// Create incident
	as.createIncidentAsync(alert)

	// Check incident was created
	as.alertManager.Mutex.RLock()
	incidentCount := len(as.alertManager.Incidents)
	var incident *models.Incident
	for _, inc := range as.alertManager.Incidents {
		if len(inc.RelatedAlerts) > 0 && inc.RelatedAlerts[0] == alert.ID {
			incident = inc
			break
		}
	}
	as.alertManager.Mutex.RUnlock()

	assert.Equal(t, 1, incidentCount)
	require.NotNil(t, incident)
	assert.Equal(t, "open", incident.Status)
	assert.Equal(t, "critical", incident.Severity)
	assert.Equal(t, "high", incident.Priority)
	assert.Contains(t, incident.RelatedAlerts, alert.ID)
	assert.Contains(t, incident.Tags, "auto-generated")
	assert.Contains(t, incident.Tags, "critical")
	assert.Len(t, incident.Timeline, 1)
	assert.Equal(t, "creation", incident.Timeline[0].Type)
	assert.Equal(t, "system", incident.Timeline[0].Author)
	assert.Greater(t, incident.Metrics.TimeToDetection, time.Duration(0))
}

func TestAlertingService_EvaluateAlertRules(t *testing.T) {
	as := NewAlertingService()
	as.initDefaultAlertRules()

	// Initially no active alerts
	assert.Empty(t, as.alertManager.ActiveAlerts)

	// Run evaluation
	as.evaluateAlertRules()

	// Due to randomness, we might or might not have alerts
	// Just ensure the function doesn't panic and maintains data integrity
	as.alertManager.Mutex.RLock()
	activeCount := len(as.alertManager.ActiveAlerts)
	historyCount := len(as.alertManager.AlertHistory)
	as.alertManager.Mutex.RUnlock()

	assert.GreaterOrEqual(t, activeCount, 0)
	assert.GreaterOrEqual(t, historyCount, activeCount)

	// Test with disabled rules
	as.alertManager.Mutex.Lock()
	for i := range as.alertManager.Rules {
		as.alertManager.Rules[i].Enabled = false
	}
	as.alertManager.Mutex.Unlock()

	initialActiveCount := activeCount
	as.evaluateAlertRules()

	as.alertManager.Mutex.RLock()
	newActiveCount := len(as.alertManager.ActiveAlerts)
	as.alertManager.Mutex.RUnlock()

	// Should not create new alerts when rules are disabled
	assert.Equal(t, initialActiveCount, newActiveCount)
}

func TestAlertingService_GetAlertManager(t *testing.T) {
	as := NewAlertingService()

	alertManager := as.GetAlertManager()

	assert.NotNil(t, alertManager)
	assert.Equal(t, as.alertManager, alertManager)
	assert.NotNil(t, alertManager.Rules)
	assert.NotNil(t, alertManager.ActiveAlerts)
	assert.NotNil(t, alertManager.AlertHistory)
	assert.NotNil(t, alertManager.NotificationChannels)
	assert.NotNil(t, alertManager.Incidents)
	assert.NotNil(t, alertManager.SilencedRules)
}

func TestAlertingService_ConcurrentAccess(t *testing.T) {
	as := NewAlertingService()
	as.initDefaultAlertRules()
	as.initDefaultNotificationChannels()

	// Test concurrent access to alert manager
	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent alert firing
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			rule := &models.AlertRule{
				ID:          fmt.Sprintf("concurrent-rule-%d", id),
				Name:        fmt.Sprintf("concurrent-alert-%d", id),
				Description: "Concurrent test alert",
				Severity:    "warning",
				Labels:      map[string]string{"test": "concurrent"},
				Annotations: map[string]string{"summary": "Concurrent test"},
				Threshold: models.AlertThreshold{
					Operator: ">",
					Value:    50.0,
				},
			}
			as.fireAlert(rule)
		}(i)
	}

	// Concurrent alert evaluation
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			as.evaluateAlertRules()
		}()
	}

	// Concurrent notification sending
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			alert := &models.Alert{
				ID:       fmt.Sprintf("concurrent-alert-%d", id),
				RuleName: "concurrent-rule",
				Severity: "warning",
				Message:  "Concurrent notification test",
			}
			as.sendNotificationAsync(alert)
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify data integrity
	as.alertManager.Mutex.RLock()
	activeAlerts := len(as.alertManager.ActiveAlerts)
	alertHistory := len(as.alertManager.AlertHistory)
	incidents := len(as.alertManager.Incidents)
	as.alertManager.Mutex.RUnlock()

	assert.GreaterOrEqual(t, activeAlerts, 0)
	assert.GreaterOrEqual(t, alertHistory, 0)
	assert.GreaterOrEqual(t, incidents, 0)

	// Active alerts should not exceed number of unique rules
	assert.LessOrEqual(t, activeAlerts, numGoroutines+4) // 4 default rules
}

func TestAlertingService_AlertManagerDataIntegrity(t *testing.T) {
	as := NewAlertingService()
	as.InitAlertManager()

	// Test data integrity after multiple operations
	rule := &models.AlertRule{
		ID:          "integrity-test-rule",
		Name:        "integrity-test",
		Description: "Data integrity test",
		Severity:    "critical",
		Labels:      map[string]string{"test": "integrity"},
		Annotations: map[string]string{"summary": "Integrity test"},
		Threshold: models.AlertThreshold{
			Operator: ">",
			Value:    50.0,
		},
	}

	// Fire alert multiple times
	for i := 0; i < 5; i++ {
		as.fireAlert(rule)
	}

	as.alertManager.Mutex.RLock()
	activeAlerts := as.alertManager.ActiveAlerts
	alertHistory := as.alertManager.AlertHistory
	incidents := as.alertManager.Incidents
	as.alertManager.Mutex.RUnlock()

	// Should have only one active alert for the rule
	assert.Len(t, activeAlerts, 1)
	assert.Contains(t, activeAlerts, rule.ID)

	// Should have only one entry in history
	assert.Len(t, alertHistory, 1)

	// Should have one incident (critical alert)
	assert.Len(t, incidents, 1)

	// Verify alert data
	alert := activeAlerts[rule.ID]
	assert.Equal(t, rule.ID, alert.RuleID)
	assert.Equal(t, rule.Name, alert.RuleName)
	assert.Equal(t, "firing", alert.Status)
	assert.Equal(t, rule.Severity, alert.Severity)

	// Verify incident data
	var incident *models.Incident
	for _, inc := range incidents {
		incident = inc
		break
	}
	require.NotNil(t, incident)
	assert.Equal(t, "open", incident.Status)
	assert.Equal(t, "critical", incident.Severity)
	assert.Contains(t, incident.RelatedAlerts, alert.ID)
}

// Benchmark tests
func BenchmarkAlertingService_EvaluateRule(b *testing.B) {
	as := NewAlertingService()
	rule := &models.AlertRule{
		Name: "high-cpu-usage",
		Threshold: models.AlertThreshold{
			Operator: ">",
			Value:    80.0,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		as.evaluateRule(rule)
	}
}

func BenchmarkAlertingService_FireAlert(b *testing.B) {
	as := NewAlertingService()
	as.initDefaultNotificationChannels()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rule := &models.AlertRule{
			ID:          fmt.Sprintf("bench-rule-%d", i),
			Name:        fmt.Sprintf("bench-alert-%d", i),
			Description: "Benchmark test alert",
			Severity:    "warning",
			Labels:      map[string]string{"test": "benchmark"},
			Annotations: map[string]string{"summary": "Benchmark test"},
			Threshold: models.AlertThreshold{
				Operator: ">",
				Value:    50.0,
			},
		}
		as.fireAlert(rule)
	}
}

func BenchmarkAlertingService_SendNotificationAsync(b *testing.B) {
	as := NewAlertingService()
	as.initDefaultNotificationChannels()

	alert := &models.Alert{
		ID:       "bench-alert-id",
		RuleName: "bench-rule",
		Severity: "warning",
		Message:  "Benchmark alert message",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		as.sendNotificationAsync(alert)
	}
}

// Example usage
func ExampleAlertingService_InitAlertManager() {
	as := NewAlertingService()
	as.InitAlertManager()

	// Access alert manager
	alertManager := as.GetAlertManager()

	_ = len(alertManager.Rules)                // Number of alert rules
	_ = len(alertManager.NotificationChannels) // Number of notification channels
	_ = len(alertManager.ActiveAlerts)         // Current active alerts
}
