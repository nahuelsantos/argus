package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Intelligence & Analytics Metrics
var (
	AnomaliesDetectedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dinky_anomalies_detected_total",
			Help: "Total number of anomalies detected by ML models",
		},
		[]string{"model_type", "metric_name", "severity"},
	)

	PredictiveAlertsGenerated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dinky_predictive_alerts_generated_total",
			Help: "Total number of predictive alerts generated",
		},
		[]string{"metric_name", "severity", "probability_range"},
	)

	RecommendationsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dinky_recommendations_created_total",
			Help: "Total number of recommendations created",
		},
		[]string{"type", "priority", "component"},
	)

	ModelAccuracy = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dinky_ml_model_accuracy",
			Help: "Accuracy of ML models",
		},
		[]string{"model_id", "model_type"},
	)

	AnomalyDetectionLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dinky_anomaly_detection_duration_seconds",
			Help:    "Time taken to run anomaly detection",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
		},
		[]string{"model_type"},
	)

	RootCauseAnalysisActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "dinky_root_cause_analysis_active",
			Help: "Number of active root cause analyses",
		},
	)

	CapacityPlanningForecasts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dinky_capacity_forecasts_generated_total",
			Help: "Total number of capacity forecasts generated",
		},
		[]string{"service", "resource_type"},
	)

	CostOptimizationSavings = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "dinky_cost_optimization_savings_dollars",
			Help: "Potential cost savings identified",
		},
	)

	PerformanceInsightsGenerated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dinky_performance_insights_generated_total",
			Help: "Total number of performance insights generated",
		},
		[]string{"type", "severity", "component"},
	)

	IntelligenceServiceDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dinky_intelligence_service_duration_seconds",
			Help:    "Duration of intelligence service operations",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10),
		},
		[]string{"operation", "status"},
	)
)

// RegisterIntelligenceMetrics registers metrics
func RegisterIntelligenceMetrics() {
	prometheus.MustRegister(
		AnomaliesDetectedTotal,
		PredictiveAlertsGenerated,
		RecommendationsCreated,
		ModelAccuracy,
		AnomalyDetectionLatency,
		RootCauseAnalysisActive,
		CapacityPlanningForecasts,
		CostOptimizationSavings,
		PerformanceInsightsGenerated,
		IntelligenceServiceDuration,
	)
}
