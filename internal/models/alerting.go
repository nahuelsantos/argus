package models

import (
	"sync"
	"time"
)

// AlertRule represents an alert rule configuration
type AlertRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Query       string            `json:"query"`
	Threshold   AlertThreshold    `json:"threshold"`
	Severity    string            `json:"severity"` // "info", "warning", "critical"
	Duration    time.Duration     `json:"duration"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// AlertThreshold represents alert threshold configuration
type AlertThreshold struct {
	Operator string  `json:"operator"` // ">", "<", ">=", "<=", "=="
	Value    float64 `json:"value"`
}

// Alert represents a fired alert instance
type Alert struct {
	ID           string            `json:"id"`
	RuleID       string            `json:"rule_id"`
	RuleName     string            `json:"rule_name"`
	Status       string            `json:"status"` // "firing", "resolved"
	Severity     string            `json:"severity"`
	Message      string            `json:"message"`
	StartsAt     time.Time         `json:"starts_at"`
	EndsAt       *time.Time        `json:"ends_at,omitempty"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	Value        float64           `json:"value"`
	Threshold    AlertThreshold    `json:"threshold"`
	GeneratorURL string            `json:"generator_url"`
}

// Incident represents an incident created from alerts
type Incident struct {
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	Status          string           `json:"status"` // "open", "investigating", "resolved", "closed"
	Severity        string           `json:"severity"`
	Priority        string           `json:"priority"` // "low", "medium", "high", "critical"
	Assignee        string           `json:"assignee,omitempty"`
	AffectedService string           `json:"affected_service"`
	RelatedAlerts   []string         `json:"related_alerts"`
	Tags            []string         `json:"tags"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	ResolvedAt      *time.Time       `json:"resolved_at,omitempty"`
	Timeline        []IncidentUpdate `json:"timeline"`
	Metrics         IncidentMetrics  `json:"metrics"`
	PostMortem      *PostMortem      `json:"post_mortem,omitempty"`
}

// IncidentUpdate represents an update in the incident timeline
type IncidentUpdate struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Author    string                 `json:"author"`
	Type      string                 `json:"type"` // "status_change", "comment", "assignment", "resolution"
	Message   string                 `json:"message"`
	OldValue  string                 `json:"old_value,omitempty"`
	NewValue  string                 `json:"new_value,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// IncidentMetrics represents incident performance metrics
type IncidentMetrics struct {
	TimeToDetection      time.Duration `json:"time_to_detection"`
	TimeToAcknowledgment time.Duration `json:"time_to_acknowledgment"`
	TimeToResolution     time.Duration `json:"time_to_resolution"`
	MTTR                 time.Duration `json:"mttr"` // Mean Time To Resolution
	DowntimeDuration     time.Duration `json:"downtime_duration"`
}

// PostMortem represents a post-incident analysis
type PostMortem struct {
	ID               string     `json:"id"`
	IncidentID       string     `json:"incident_id"`
	Summary          string     `json:"summary"`
	RootCause        string     `json:"root_cause"`
	Timeline         string     `json:"timeline"`
	ImpactAssessment string     `json:"impact_assessment"`
	ActionItems      []string   `json:"action_items"`
	LessonsLearned   []string   `json:"lessons_learned"`
	CreatedBy        string     `json:"created_by"`
	CreatedAt        time.Time  `json:"created_at"`
	ReviewedBy       []string   `json:"reviewed_by"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
}

// NotificationChannel represents a notification channel configuration
type NotificationChannel struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // "slack", "email", "webhook", "pagerduty"
	Config     map[string]interface{} `json:"config"`
	Conditions map[string]interface{} `json:"conditions"` // When to use this channel
	RateLimit  RateLimit              `json:"rate_limit"`
	Enabled    bool                   `json:"enabled"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// RateLimit represents rate limiting configuration for notifications
type RateLimit struct {
	MaxAlerts   int           `json:"max_alerts"`
	TimeWindow  time.Duration `json:"time_window"`
	GroupingKey string        `json:"grouping_key"`
}

// AlertManager represents the central alert management system
type AlertManager struct {
	Rules                []AlertRule           `json:"rules"`
	ActiveAlerts         map[string]*Alert     `json:"active_alerts"`
	AlertHistory         []*Alert              `json:"alert_history"`
	NotificationChannels []NotificationChannel `json:"notification_channels"`
	Incidents            map[string]*Incident  `json:"incidents"`
	SilencedRules        map[string]time.Time  `json:"silenced_rules"`
	Mutex                sync.RWMutex          `json:"-"`
}
