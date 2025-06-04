package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Analytics Models
func TestRecommendation(t *testing.T) {
	now := time.Now()

	rec := Recommendation{
		ID:          "rec-123",
		Type:        "scaling",
		Priority:    "high",
		Title:       "Scale up API service",
		Description: "API service is experiencing high load",
		Impact:      "High performance improvement",
		Effort:      "medium",
		CreatedAt:   now,
	}

	// Test JSON marshaling
	data, err := json.Marshal(rec)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Scale up API service")

	// Test JSON unmarshaling
	var unmarshaled Recommendation
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, rec.ID, unmarshaled.ID)
	assert.Equal(t, rec.Type, unmarshaled.Type)
	assert.Equal(t, rec.Priority, unmarshaled.Priority)
}

// Test Alerting Models
func TestAlertRule(t *testing.T) {
	now := time.Now()

	rule := AlertRule{
		ID:          "rule-123",
		Name:        "High CPU Usage",
		Description: "Alert when CPU usage exceeds 80%",
		Query:       "cpu_usage > 80",
		Threshold: AlertThreshold{
			Operator: ">",
			Value:    80.0,
		},
		Severity:    "warning",
		Duration:    5 * time.Minute,
		Labels:      map[string]string{"service": "api"},
		Annotations: map[string]string{"runbook": "https://wiki.example.com"},
		Enabled:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(rule)
	require.NoError(t, err)

	var unmarshaled AlertRule
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, rule.ID, unmarshaled.ID)
	assert.Equal(t, rule.Threshold.Value, unmarshaled.Threshold.Value)
}

func TestAlert(t *testing.T) {
	now := time.Now()

	alert := Alert{
		ID:       "alert-123",
		RuleID:   "rule-123",
		Status:   "firing",
		Severity: "warning",
		Message:  "CPU usage is 85%",
		StartsAt: now,
		Value:    85.0,
		Threshold: AlertThreshold{
			Operator: ">",
			Value:    80.0,
		},
	}

	data, err := json.Marshal(alert)
	require.NoError(t, err)

	var unmarshaled Alert
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, alert.Status, unmarshaled.Status)
	assert.Equal(t, alert.Value, unmarshaled.Value)
}

func TestIncident(t *testing.T) {
	now := time.Now()

	incident := Incident{
		ID:              "inc-123",
		Title:           "API Service Down",
		Status:          "open",
		Severity:        "critical",
		AffectedService: "api-service",
		CreatedAt:       now,
		Metrics: IncidentMetrics{
			TimeToDetection: 2 * time.Minute,
			MTTR:            2 * time.Hour,
		},
	}

	data, err := json.Marshal(incident)
	require.NoError(t, err)

	var unmarshaled Incident
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, incident.Title, unmarshaled.Title)
	assert.Equal(t, incident.Metrics.MTTR, unmarshaled.Metrics.MTTR)
}

func TestNotificationChannel(t *testing.T) {
	now := time.Now()

	channel := NotificationChannel{
		ID:   "channel-123",
		Name: "Slack Alerts",
		Type: "slack",
		Config: map[string]interface{}{
			"webhook_url": "https://hooks.slack.com/services/...",
			"channel":     "#alerts",
		},
		Conditions: map[string]interface{}{
			"severity": []string{"warning", "critical"},
		},
		RateLimit: RateLimit{
			MaxAlerts:   10,
			TimeWindow:  time.Hour,
			GroupingKey: "service",
		},
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test JSON marshaling
	data, err := json.Marshal(channel)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Slack Alerts")

	// Test JSON unmarshaling
	var unmarshaled NotificationChannel
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, channel.ID, unmarshaled.ID)
	assert.Equal(t, channel.Type, unmarshaled.Type)
	assert.Equal(t, channel.RateLimit.MaxAlerts, unmarshaled.RateLimit.MaxAlerts)
}

func TestAlertManager(t *testing.T) {
	now := time.Now()

	manager := AlertManager{
		Rules: []AlertRule{
			{
				ID:       "rule-1",
				Name:     "CPU High",
				Severity: "warning",
				Enabled:  true,
			},
		},
		ActiveAlerts: map[string]*Alert{
			"alert-1": {
				ID:       "alert-1",
				RuleID:   "rule-1",
				Status:   "firing",
				StartsAt: now,
			},
		},
		AlertHistory: []*Alert{
			{
				ID:       "alert-old",
				RuleID:   "rule-1",
				Status:   "resolved",
				StartsAt: now.Add(-time.Hour),
			},
		},
		NotificationChannels: []NotificationChannel{
			{
				ID:   "channel-1",
				Name: "Email",
				Type: "email",
			},
		},
		Incidents: map[string]*Incident{
			"inc-1": {
				ID:     "inc-1",
				Title:  "Test Incident",
				Status: "open",
			},
		},
		SilencedRules: map[string]time.Time{
			"rule-2": now.Add(time.Hour),
		},
	}

	// Test JSON marshaling (excluding Mutex)
	data, err := json.Marshal(manager)
	require.NoError(t, err)
	assert.Contains(t, string(data), "CPU High")

	// Test JSON unmarshaling
	var unmarshaled AlertManager
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Len(t, unmarshaled.Rules, 1)
	assert.Len(t, unmarshaled.ActiveAlerts, 1)
	assert.Len(t, unmarshaled.AlertHistory, 1)
	assert.Len(t, unmarshaled.NotificationChannels, 1)
	assert.Len(t, unmarshaled.Incidents, 1)
	assert.Len(t, unmarshaled.SilencedRules, 1)
}

// Test APM Models
func TestServiceDependency(t *testing.T) {
	dependency := ServiceDependency{
		ServiceName:  "user-service",
		Operation:    "get_user",
		ResponseTime: 150 * time.Millisecond,
		StatusCode:   200,
		ErrorRate:    0.05,
		RequestCount: 1000,
		Dependencies: []string{"database", "cache"},
		CustomAttributes: map[string]string{
			"version": "v1.2.3",
			"region":  "us-east-1",
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(dependency)
	require.NoError(t, err)
	assert.Contains(t, string(data), "user-service")

	// Test JSON unmarshaling
	var unmarshaled ServiceDependency
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, dependency.ServiceName, unmarshaled.ServiceName)
	assert.Equal(t, dependency.ResponseTime, unmarshaled.ResponseTime)
	assert.Equal(t, dependency.ErrorRate, unmarshaled.ErrorRate)
	assert.Len(t, unmarshaled.Dependencies, 2)
}

func TestAPMData(t *testing.T) {
	now := time.Now()

	apmData := APMData{
		ServiceName:   "api-service",
		TraceID:       "trace-123",
		OperationName: "handle_request",
		StartTime:     now,
		Duration:      250 * time.Millisecond,
		StatusCode:    200,
		ResourceUsage: ResourceMetrics{
			CPUUsage:    45.5,
			MemoryUsage: 1024 * 1024 * 100,
		},
	}

	data, err := json.Marshal(apmData)
	require.NoError(t, err)

	var unmarshaled APMData
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, apmData.ServiceName, unmarshaled.ServiceName)
	assert.Equal(t, apmData.Duration, unmarshaled.Duration)
}

func TestPerformanceProfile(t *testing.T) {
	profile := PerformanceProfile{
		Operation:       "api_endpoint",
		P50ResponseTime: 100.0,
		P95ResponseTime: 250.0,
		P99ResponseTime: 500.0,
		ErrorRate:       0.02,
		ThroughputRPS:   1000.0,
		ResourceProfile: ResourceMetrics{
			CPUUsage:    30.0,
			MemoryUsage: 1024 * 1024 * 50, // 50MB
		},
		Bottlenecks:     []string{"database_query", "external_api"},
		Recommendations: []string{"Add caching", "Optimize queries"},
	}

	// Test JSON marshaling
	data, err := json.Marshal(profile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "api_endpoint")

	// Test JSON unmarshaling
	var unmarshaled PerformanceProfile
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, profile.Operation, unmarshaled.Operation)
	assert.Equal(t, profile.P99ResponseTime, unmarshaled.P99ResponseTime)
	assert.Equal(t, profile.ThroughputRPS, unmarshaled.ThroughputRPS)
	assert.Len(t, unmarshaled.Bottlenecks, 2)
	assert.Len(t, unmarshaled.Recommendations, 2)
}

// Test Logging Models
func TestLogContext(t *testing.T) {
	context := LogContext{
		RequestID:   "req-123",
		TraceID:     "trace-456",
		SpanID:      "span-789",
		UserID:      "user-123",
		SessionID:   "session-456",
		ServiceName: "api-service",
		Version:     "v1.2.3",
		Environment: "production",
	}

	// Test JSON marshaling
	data, err := json.Marshal(context)
	require.NoError(t, err)
	assert.Contains(t, string(data), "req-123")

	// Test JSON unmarshaling
	var unmarshaled LogContext
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, context.RequestID, unmarshaled.RequestID)
	assert.Equal(t, context.ServiceName, unmarshaled.ServiceName)
	assert.Equal(t, context.Environment, unmarshaled.Environment)
}

func TestLogEntry(t *testing.T) {
	now := time.Now()

	entry := LogEntry{
		Level:     "info",
		Timestamp: now,
		Message:   "Request processed",
		Context: LogContext{
			RequestID:   "req-123",
			ServiceName: "api-service",
		},
		Data: map[string]interface{}{
			"user_id": "user-123",
			"status":  200,
		},
	}

	data, err := json.Marshal(entry)
	require.NoError(t, err)

	var unmarshaled LogEntry
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, entry.Level, unmarshaled.Level)
	assert.Equal(t, entry.Context.RequestID, unmarshaled.Context.RequestID)
}

// Test Context Keys
func TestContextKeys(t *testing.T) {
	assert.Equal(t, "request_id", string(RequestIDKey))
	assert.Equal(t, "trace_id", string(TraceIDKey))
	assert.Equal(t, "user_id", string(UserIDKey))
	assert.Equal(t, "session_id", string(SessionIDKey))
	assert.Equal(t, "start_time", string(StartTimeKey))
}

// Edge Cases and Validation Tests
func TestEmptyStructsJSONMarshaling(t *testing.T) {
	tests := []struct {
		name  string
		model interface{}
	}{
		{"Empty Recommendation", Recommendation{}},
		{"Empty AlertRule", AlertRule{}},
		{"Empty Alert", Alert{}},
		{"Empty Incident", Incident{}},
		{"Empty APMData", APMData{}},
		{"Empty LogEntry", LogEntry{}},
		{"Empty LogContext", LogContext{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.model)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Should be able to unmarshal back to a pointer
			switch tt.model.(type) {
			case Recommendation:
				var target Recommendation
				err = json.Unmarshal(data, &target)
			case AlertRule:
				var target AlertRule
				err = json.Unmarshal(data, &target)
			case Alert:
				var target Alert
				err = json.Unmarshal(data, &target)
			case Incident:
				var target Incident
				err = json.Unmarshal(data, &target)
			case APMData:
				var target APMData
				err = json.Unmarshal(data, &target)
			case LogEntry:
				var target LogEntry
				err = json.Unmarshal(data, &target)
			case LogContext:
				var target LogContext
				err = json.Unmarshal(data, &target)
			}
			require.NoError(t, err)
		})
	}
}

func TestNilPointerHandling(t *testing.T) {
	// Test Alert with nil EndsAt
	alert := Alert{
		ID:       "alert-1",
		Status:   "firing",
		StartsAt: time.Now(),
		EndsAt:   nil, // Should handle nil pointer
	}

	data, err := json.Marshal(alert)
	require.NoError(t, err)

	var unmarshaled Alert
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Nil(t, unmarshaled.EndsAt)

	// Test Incident with nil ResolvedAt and PostMortem
	incident := Incident{
		ID:         "inc-1",
		Status:     "open",
		ResolvedAt: nil,
		PostMortem: nil,
	}

	data, err = json.Marshal(incident)
	require.NoError(t, err)

	var unmarshaledIncident Incident
	err = json.Unmarshal(data, &unmarshaledIncident)
	require.NoError(t, err)
	assert.Nil(t, unmarshaledIncident.ResolvedAt)
	assert.Nil(t, unmarshaledIncident.PostMortem)
}

func TestComplexDataStructures(t *testing.T) {
	// Test nested maps and slices
	entry := LogEntry{
		Level:   "info",
		Message: "Complex data test",
		Data: map[string]interface{}{
			"nested_map": map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": "deep_value",
					"array":  []interface{}{1, 2, "three", true},
				},
			},
			"simple_array": []interface{}{"a", "b", "c"},
			"number":       42.5,
			"boolean":      true,
			"null_value":   nil,
		},
	}

	data, err := json.Marshal(entry)
	require.NoError(t, err)

	var unmarshaled LogEntry
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, entry.Level, unmarshaled.Level)
	assert.Equal(t, entry.Message, unmarshaled.Message)
	assert.NotNil(t, unmarshaled.Data["nested_map"])
	assert.NotNil(t, unmarshaled.Data["simple_array"])
}

// Benchmark tests
func BenchmarkRecommendationMarshal(b *testing.B) {
	rec := Recommendation{
		ID:          "rec-123",
		Type:        "scaling",
		Priority:    "high",
		Title:       "Scale up service",
		Description: "Service needs scaling",
		Impact:      "High",
		Effort:      "medium",
		CreatedAt:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(rec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLogEntryMarshal(b *testing.B) {
	entry := LogEntry{
		Level:     "info",
		Timestamp: time.Now(),
		Message:   "Test message",
		Context: LogContext{
			RequestID:   "req-123",
			ServiceName: "test-service",
		},
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(entry)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAPMDataMarshal(b *testing.B) {
	apmData := APMData{
		ServiceName:   "api-service",
		TraceID:       "trace-123",
		SpanID:        "span-456",
		OperationName: "handle_request",
		StartTime:     time.Now(),
		Duration:      250 * time.Millisecond,
		StatusCode:    200,
		ResourceUsage: ResourceMetrics{
			CPUUsage:    45.5,
			MemoryUsage: 1024 * 1024 * 100,
		},
		CustomTags: map[string]string{
			"user_id": "user-123",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(apmData)
		if err != nil {
			b.Fatal(err)
		}
	}
}
