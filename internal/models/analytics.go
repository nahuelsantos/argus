package models

import (
	"time"
)

// Intelligence & Analytics Models

// AnomalyDetectionModel represents ML model for anomaly detection
type AnomalyDetectionModel struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"` // statistical, isolation_forest, lstm, etc.
	Status       string                 `json:"status"`
	Accuracy     float64                `json:"accuracy"`
	TrainingData TrainingDataset        `json:"training_data"`
	Parameters   map[string]interface{} `json:"parameters"`
	LastTrained  time.Time              `json:"last_trained"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// TrainingDataset represents training data for ML models
type TrainingDataset struct {
	Source      string                 `json:"source"`
	Timerange   TimeRange              `json:"timerange"`
	Metrics     []string               `json:"metrics"`
	SampleCount int64                  `json:"sample_count"`
	Features    []string               `json:"features"`
	Labels      map[string]interface{} `json:"labels"`
}

// TimeRange represents a time range for data queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// AnomalyScore represents anomaly detection results
type AnomalyScore struct {
	Timestamp  time.Time              `json:"timestamp"`
	MetricName string                 `json:"metric_name"`
	Value      float64                `json:"value"`
	Score      float64                `json:"score"` // 0-1, higher = more anomalous
	Threshold  float64                `json:"threshold"`
	IsAnomaly  bool                   `json:"is_anomaly"`
	Confidence float64                `json:"confidence"`
	Context    map[string]interface{} `json:"context"`
	ModelID    string                 `json:"model_id"`
}

// PredictiveAlert represents predictive alerting
type PredictiveAlert struct {
	ID              string           `json:"id"`
	RuleID          string           `json:"rule_id"`
	Prediction      Prediction       `json:"prediction"`
	Probability     float64          `json:"probability"`
	TimeToEvent     time.Duration    `json:"time_to_event"`
	Severity        string           `json:"severity"`
	Status          string           `json:"status"`
	Recommendations []Recommendation `json:"recommendations"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// Prediction represents a forecasted event
type Prediction struct {
	Type           string   `json:"type"` // threshold_breach, resource_exhaustion, etc.
	Description    string   `json:"description"`
	Metric         string   `json:"metric"`
	CurrentValue   float64  `json:"current_value"`
	PredictedValue float64  `json:"predicted_value"`
	Threshold      float64  `json:"threshold"`
	Confidence     float64  `json:"confidence"`
	Factors        []string `json:"factors"` // Contributing factors
}

// Recommendation represents actionable insights
type Recommendation struct {
	ID          string                `json:"id"`
	Type        string                `json:"type"`     // scaling, optimization, configuration
	Priority    string                `json:"priority"` // high, medium, low
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Impact      string                `json:"impact"`
	Effort      string                `json:"effort"` // low, medium, high
	Actions     []RecommendedAction   `json:"actions"`
	Metrics     RecommendationMetrics `json:"metrics"`
	CreatedAt   time.Time             `json:"created_at"`
}

// RecommendedAction represents specific actions to take
type RecommendedAction struct {
	Type        string                 `json:"type"` // scale_up, tune_parameter, etc.
	Description string                 `json:"description"`
	Command     string                 `json:"command,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Automated   bool                   `json:"automated"`
}

// RecommendationMetrics represents expected impact metrics
type RecommendationMetrics struct {
	CostSavings     float64 `json:"cost_savings,omitempty"`
	PerformanceGain float64 `json:"performance_gain,omitempty"`
	ResourceSavings float64 `json:"resource_savings,omitempty"`
	ROI             float64 `json:"roi,omitempty"`
}

// RootCauseAnalysis represents automated incident investigation
type RootCauseAnalysis struct {
	ID           string          `json:"id"`
	IncidentID   string          `json:"incident_id"`
	Status       string          `json:"status"`
	Confidence   float64         `json:"confidence"`
	RootCauses   []RootCause     `json:"root_causes"`
	Timeline     []TimelineEvent `json:"timeline"`
	Correlations []Correlation   `json:"correlations"`
	CreatedAt    time.Time       `json:"created_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
}

// RootCause represents identified root cause
type RootCause struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"` // configuration, resource, dependency, etc.
	Component   string     `json:"component"`
	Description string     `json:"description"`
	Evidence    []Evidence `json:"evidence"`
	Probability float64    `json:"probability"`
	Impact      string     `json:"impact"`
}

// Evidence represents supporting evidence for root cause
type Evidence struct {
	Type        string      `json:"type"` // metric, log, trace, etc.
	Source      string      `json:"source"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
	Timestamp   time.Time   `json:"timestamp"`
	Relevance   float64     `json:"relevance"`
}

// TimelineEvent represents events in incident timeline
type TimelineEvent struct {
	Timestamp   time.Time   `json:"timestamp"`
	Type        string      `json:"type"`
	Component   string      `json:"component"`
	Description string      `json:"description"`
	Severity    string      `json:"severity"`
	Data        interface{} `json:"data"`
}

// Correlation represents correlations between metrics/events
type Correlation struct {
	MetricA     string        `json:"metric_a"`
	MetricB     string        `json:"metric_b"`
	Coefficient float64       `json:"coefficient"` // -1 to 1
	Strength    string        `json:"strength"`    // weak, moderate, strong
	Type        string        `json:"type"`        // positive, negative
	Timelag     time.Duration `json:"timelag"`
}

// PerformanceInsight represents performance optimization insights
type PerformanceInsight struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"` // bottleneck, optimization, pattern
	Component   string            `json:"component"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Severity    string            `json:"severity"`
	Impact      PerformanceImpact `json:"impact"`
	Suggestions []Recommendation  `json:"suggestions"`
	Metrics     InsightMetrics    `json:"metrics"`
	CreatedAt   time.Time         `json:"created_at"`
}

// PerformanceImpact represents impact on performance
type PerformanceImpact struct {
	Latency     float64 `json:"latency_ms"`
	Throughput  float64 `json:"throughput_reduction_percent"`
	ErrorRate   float64 `json:"error_rate_increase_percent"`
	ResourceUse float64 `json:"resource_overhead_percent"`
}

// InsightMetrics represents metrics related to insights
type InsightMetrics struct {
	BaselineLatency    float64 `json:"baseline_latency_ms"`
	CurrentLatency     float64 `json:"current_latency_ms"`
	BaselineThroughput float64 `json:"baseline_throughput_rps"`
	CurrentThroughput  float64 `json:"current_throughput_rps"`
	TrendDirection     string  `json:"trend_direction"` // improving, degrading, stable
}

// CapacityPlan represents capacity planning recommendations
type CapacityPlan struct {
	ID              string                   `json:"id"`
	Service         string                   `json:"service"`
	TimeHorizon     time.Duration            `json:"time_horizon"`
	Forecast        ResourceForecast         `json:"forecast"`
	Recommendations []CapacityRecommendation `json:"recommendations"`
	CostAnalysis    CostAnalysis             `json:"cost_analysis"`
	CreatedAt       time.Time                `json:"created_at"`
}

// ResourceForecast represents forecasted resource needs
type ResourceForecast struct {
	CPU     ResourceProjection `json:"cpu"`
	Memory  ResourceProjection `json:"memory"`
	Storage ResourceProjection `json:"storage"`
	Network ResourceProjection `json:"network"`
}

// ResourceProjection represents resource usage projection
type ResourceProjection struct {
	Current    float64     `json:"current"`
	Projected  float64     `json:"projected"`
	Peak       float64     `json:"peak"`
	Average    float64     `json:"average"`
	Trend      string      `json:"trend"` // increasing, decreasing, stable
	Confidence float64     `json:"confidence"`
	Timeline   []DataPoint `json:"timeline"`
}

// DataPoint represents a data point in time series
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// CapacityRecommendation represents capacity planning recommendation
type CapacityRecommendation struct {
	Type       string                 `json:"type"` // scale_up, scale_down, optimize
	Component  string                 `json:"component"`
	Action     string                 `json:"action"`
	Timing     time.Time              `json:"timing"`
	Parameters map[string]interface{} `json:"parameters"`
	CostImpact float64                `json:"cost_impact"`
	Urgency    string                 `json:"urgency"` // low, medium, high, critical
}

// CostAnalysis represents cost optimization analysis
type CostAnalysis struct {
	CurrentCost   float64            `json:"current_cost"`
	ProjectedCost float64            `json:"projected_cost"`
	Savings       float64            `json:"potential_savings"`
	Breakdown     map[string]float64 `json:"cost_breakdown"`
	Optimizations []CostOptimization `json:"optimizations"`
}

// CostOptimization represents cost optimization opportunity
type CostOptimization struct {
	Type        string  `json:"type"` // rightsizing, unused_resources, etc.
	Description string  `json:"description"`
	Savings     float64 `json:"potential_savings"`
	Effort      string  `json:"implementation_effort"`
	Risk        string  `json:"risk_level"`
	Priority    int     `json:"priority"`
}

// IntelligenceMetrics represents metrics
type IntelligenceMetrics struct {
	AnomaliesDetected       int64   `json:"anomalies_detected"`
	PredictionsGenerated    int64   `json:"predictions_generated"`
	RecommendationsCreated  int64   `json:"recommendations_created"`
	AccuracyRate            float64 `json:"accuracy_rate"`
	FalsePositiveRate       float64 `json:"false_positive_rate"`
	TimeToDetection         float64 `json:"time_to_detection_ms"`
	CostSavingsRealized     float64 `json:"cost_savings_realized"`
	PerformanceImprovements float64 `json:"performance_improvements_percent"`
}
