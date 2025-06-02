package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/nahuelsantos/argus/internal/metrics"
	"github.com/nahuelsantos/argus/internal/models"
	"github.com/nahuelsantos/argus/internal/services"
)

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// IntelligenceHandler handles intelligence endpoints
type IntelligenceHandler struct {
	logger              *zap.Logger
	intelligenceService *services.IntelligenceService
}

// NewIntelligenceHandler creates a new intelligence handler
func NewIntelligenceHandler(logger *zap.Logger, intelligenceService *services.IntelligenceService) *IntelligenceHandler {
	return &IntelligenceHandler{
		logger:              logger,
		intelligenceService: intelligenceService,
	}
}

// RegisterRoutes registers intelligence routes
func (h *IntelligenceHandler) RegisterRoutes(mux *http.ServeMux) {
	// Anomaly Detection
	mux.HandleFunc("/test-anomaly-detection", h.TestAnomalyDetection)
	mux.HandleFunc("/anomaly-models", h.GetAnomalyModels)
	mux.HandleFunc("/anomaly-scores", h.GetAnomalyScores)

	// Predictive Alerts
	mux.HandleFunc("/test-predictive-alerts", h.TestPredictiveAlerts)
	mux.HandleFunc("/predictive-alerts", h.GetPredictiveAlerts)

	// Root Cause Analysis
	mux.HandleFunc("/test-root-cause-analysis", h.TestRootCauseAnalysis)
	mux.HandleFunc("/root-cause-analysis", h.GetRootCauseAnalysis)

	// Performance Insights
	mux.HandleFunc("/test-performance-insights", h.TestPerformanceInsights)
	mux.HandleFunc("/performance-insights", h.GetPerformanceInsights)

	// Capacity Planning
	mux.HandleFunc("/test-capacity-planning", h.TestCapacityPlanning)
	mux.HandleFunc("/capacity-plans", h.GetCapacityPlans)

	// Recommendations
	mux.HandleFunc("/recommendations", h.GetRecommendations)

	// Intelligence Metrics & Dashboard
	mux.HandleFunc("/intelligence-metrics", h.GetIntelligenceMetrics)
	mux.HandleFunc("/intelligence-dashboard", h.GetIntelligenceDashboard)
}

// TestAnomalyDetection tests ML-powered anomaly detection
func (h *IntelligenceHandler) TestAnomalyDetection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.IntelligenceServiceDuration.WithLabelValues("anomaly_detection", "success").Observe(duration)
	}()

	h.logger.Info("Testing anomaly detection")

	// Generate sample time series data
	metricName := "cpu_usage"
	values := h.generateSampleMetricData(100)
	timestamps := make([]time.Time, len(values))
	now := time.Now()
	for i := range timestamps {
		timestamps[i] = now.Add(time.Duration(-len(values)+i) * time.Minute)
	}

	// Run anomaly detection
	scores, err := h.intelligenceService.DetectAnomalies(r.Context(), metricName, values, timestamps)
	if err != nil {
		h.logger.Error("Failed to detect anomalies", zap.Error(err))
		http.Error(w, fmt.Sprintf("Anomaly detection failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Update metrics
	for _, score := range scores {
		severity := "normal"
		if score.IsAnomaly {
			if score.Score > 0.8 {
				severity = "high"
			} else if score.Score > 0.5 {
				severity = "medium"
			} else {
				severity = "low"
			}
			metrics.AnomaliesDetectedTotal.WithLabelValues(
				"statistical", // model type from score.ModelID lookup
				score.MetricName,
				severity,
			).Inc()
		}
	}

	response := map[string]interface{}{
		"success":         true,
		"message":         "Anomaly detection completed",
		"metric_name":     metricName,
		"data_points":     len(values),
		"anomalies_found": h.countAnomalies(scores),
		"detection_time":  fmt.Sprintf("%.2fms", time.Since(start).Seconds()*1000),
		"models_used":     len(h.intelligenceService.GetActiveModels()),
		"scores":          scores[:min(len(scores), 10)], // Return first 10 for brevity
		"timestamp":       time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetAnomalyModels returns active ML models
func (h *IntelligenceHandler) GetAnomalyModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	models := h.intelligenceService.GetActiveModels()

	response := map[string]interface{}{
		"success":   true,
		"models":    models,
		"count":     len(models),
		"timestamp": time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetAnomalyScores returns recent anomaly scores
func (h *IntelligenceHandler) GetAnomalyScores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simulate recent anomaly scores
	scores := h.generateSampleAnomalyScores(20)

	response := map[string]interface{}{
		"success":   true,
		"scores":    scores,
		"count":     len(scores),
		"timestamp": time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// TestPredictiveAlerts tests predictive alerting system
func (h *IntelligenceHandler) TestPredictiveAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.IntelligenceServiceDuration.WithLabelValues("predictive_alerts", "success").Observe(duration)
	}()

	h.logger.Info("Testing predictive alerts")

	// Generate sample metric trends
	metricData := map[string][]float64{
		"cpu_usage":     h.generateTrendingData(50, 45.0, 2.0),  // Trending upward
		"memory_usage":  h.generateTrendingData(50, 70.0, 1.5),  // Trending upward
		"disk_usage":    h.generateTrendingData(50, 85.0, 0.8),  // Slowly trending up
		"error_rate":    h.generateTrendingData(50, 2.0, 0.3),   // Trending upward
		"response_time": h.generateTrendingData(50, 120.0, 8.0), // Trending upward
	}

	// Generate predictive alerts
	alerts, err := h.intelligenceService.GeneratePredictiveAlerts(r.Context(), metricData)
	if err != nil {
		h.logger.Error("Failed to generate predictive alerts", zap.Error(err))
		http.Error(w, fmt.Sprintf("Predictive alert generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Update metrics
	for _, alert := range alerts {
		probabilityRange := "low"
		if alert.Probability > 0.7 {
			probabilityRange = "high"
		} else if alert.Probability > 0.4 {
			probabilityRange = "medium"
		}

		metrics.PredictiveAlertsGenerated.WithLabelValues(
			alert.Prediction.Metric,
			alert.Severity,
			probabilityRange,
		).Inc()
	}

	response := map[string]interface{}{
		"success":          true,
		"message":          "Predictive alerts generated",
		"alerts_generated": len(alerts),
		"processing_time":  fmt.Sprintf("%.2fms", time.Since(start).Seconds()*1000),
		"metrics_analyzed": len(metricData),
		"alerts":           alerts,
		"timestamp":        time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetPredictiveAlerts returns active predictive alerts
func (h *IntelligenceHandler) GetPredictiveAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	alerts := h.intelligenceService.GetPredictiveAlerts()

	response := map[string]interface{}{
		"success":   true,
		"alerts":    alerts,
		"count":     len(alerts),
		"timestamp": time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// TestRootCauseAnalysis tests automated root cause analysis
func (h *IntelligenceHandler) TestRootCauseAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.IntelligenceServiceDuration.WithLabelValues("root_cause_analysis", "success").Observe(duration)
	}()

	h.logger.Info("Testing root cause analysis")

	// Simulate incident ID
	incidentID := fmt.Sprintf("INC-%d", time.Now().Unix())

	// Perform root cause analysis
	analysis, err := h.intelligenceService.PerformRootCauseAnalysis(r.Context(), incidentID)
	if err != nil {
		h.logger.Error("Failed to perform root cause analysis", zap.Error(err))
		http.Error(w, fmt.Sprintf("Root cause analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Update metrics
	metrics.RootCauseAnalysisActive.Set(1) // Simulate active analysis

	response := map[string]interface{}{
		"success":         true,
		"message":         "Root cause analysis completed",
		"incident_id":     incidentID,
		"analysis_time":   fmt.Sprintf("%.2fs", time.Since(start).Seconds()),
		"confidence":      analysis.Confidence,
		"root_causes":     len(analysis.RootCauses),
		"correlations":    len(analysis.Correlations),
		"timeline_events": len(analysis.Timeline),
		"analysis":        analysis,
		"timestamp":       time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetRootCauseAnalysis returns recent root cause analyses
func (h *IntelligenceHandler) GetRootCauseAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For now, return sample data
	response := map[string]interface{}{
		"success":   true,
		"analyses":  h.generateSampleRootCauseAnalyses(),
		"timestamp": time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// TestPerformanceInsights tests performance insights generation
func (h *IntelligenceHandler) TestPerformanceInsights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.IntelligenceServiceDuration.WithLabelValues("performance_insights", "success").Observe(duration)
	}()

	h.logger.Info("Testing performance insights")

	// Generate performance insights
	insights, err := h.intelligenceService.GeneratePerformanceInsights(r.Context())
	if err != nil {
		h.logger.Error("Failed to generate performance insights", zap.Error(err))
		http.Error(w, fmt.Sprintf("Performance insight generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Update metrics
	for _, insight := range insights {
		metrics.PerformanceInsightsGenerated.WithLabelValues(
			insight.Type,
			insight.Severity,
			insight.Component,
		).Inc()
	}

	response := map[string]interface{}{
		"success":           true,
		"message":           "Performance insights generated",
		"insights_found":    len(insights),
		"analysis_time":     fmt.Sprintf("%.2fms", time.Since(start).Seconds()*1000),
		"services_analyzed": 3, // api_service, database_service, cache_service
		"insights":          insights,
		"timestamp":         time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetPerformanceInsights returns recent performance insights
func (h *IntelligenceHandler) GetPerformanceInsights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	insights, err := h.intelligenceService.GeneratePerformanceInsights(r.Context())
	if err != nil {
		h.logger.Error("Failed to get performance insights", zap.Error(err))
		http.Error(w, "Failed to get performance insights", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"insights":  insights,
		"count":     len(insights),
		"timestamp": time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// TestCapacityPlanning tests capacity planning system
func (h *IntelligenceHandler) TestCapacityPlanning(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.IntelligenceServiceDuration.WithLabelValues("capacity_planning", "success").Observe(duration)
	}()

	h.logger.Info("Testing capacity planning")

	// Get query parameters
	serviceName := r.URL.Query().Get("service")
	if serviceName == "" {
		serviceName = "api_service"
	}

	horizonParam := r.URL.Query().Get("horizon")
	horizon := 30 * 24 * time.Hour // Default 30 days
	if horizonParam != "" {
		if days, err := strconv.Atoi(horizonParam); err == nil {
			horizon = time.Duration(days) * 24 * time.Hour
		}
	}

	// Create capacity plan
	plan, err := h.intelligenceService.CreateCapacityPlan(r.Context(), serviceName, horizon)
	if err != nil {
		h.logger.Error("Failed to create capacity plan", zap.Error(err))
		http.Error(w, fmt.Sprintf("Capacity planning failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Update metrics
	resourceTypes := []string{"cpu", "memory", "storage", "network"}
	for _, resourceType := range resourceTypes {
		metrics.CapacityPlanningForecasts.WithLabelValues(serviceName, resourceType).Inc()
	}

	metrics.CostOptimizationSavings.Set(plan.CostAnalysis.Savings)

	response := map[string]interface{}{
		"success":           true,
		"message":           "Capacity plan generated",
		"service":           serviceName,
		"time_horizon":      horizon.String(),
		"planning_time":     fmt.Sprintf("%.2fms", time.Since(start).Seconds()*1000),
		"recommendations":   len(plan.Recommendations),
		"potential_savings": plan.CostAnalysis.Savings,
		"projected_cost":    plan.CostAnalysis.ProjectedCost,
		"plan":              plan,
		"timestamp":         time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetCapacityPlans returns all capacity plans
func (h *IntelligenceHandler) GetCapacityPlans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	plans := h.intelligenceService.GetCapacityPlans()

	response := map[string]interface{}{
		"success":   true,
		"plans":     plans,
		"count":     len(plans),
		"timestamp": time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetRecommendations returns all recommendations
func (h *IntelligenceHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	recommendations := h.intelligenceService.GetRecommendations()

	response := map[string]interface{}{
		"success":         true,
		"recommendations": recommendations,
		"count":           len(recommendations),
		"timestamp":       time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetIntelligenceMetrics returns metrics
func (h *IntelligenceHandler) GetIntelligenceMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := h.intelligenceService.GetIntelligenceMetrics()

	response := map[string]interface{}{
		"success":   true,
		"metrics":   metrics,
		"timestamp": time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetIntelligenceDashboard returns Phase 5 dashboard data
func (h *IntelligenceHandler) GetIntelligenceDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dashboard := map[string]interface{}{
		"overview": map[string]interface{}{
			"active_models":        len(h.intelligenceService.GetActiveModels()),
			"predictive_alerts":    len(h.intelligenceService.GetPredictiveAlerts()),
			"recommendations":      len(h.intelligenceService.GetRecommendations()),
			"capacity_plans":       len(h.intelligenceService.GetCapacityPlans()),
			"intelligence_metrics": h.intelligenceService.GetIntelligenceMetrics(),
		},
		"models":            h.intelligenceService.GetActiveModels(),
		"predictive_alerts": h.intelligenceService.GetPredictiveAlerts(),
		"recommendations":   h.intelligenceService.GetRecommendations()[:min(len(h.intelligenceService.GetRecommendations()), 5)],
		"timestamp":         time.Now(),
	}

	writeJSON(w, http.StatusOK, dashboard)
}

// Helper functions
func (h *IntelligenceHandler) generateSampleMetricData(count int) []float64 {
	data := make([]float64, count)
	baseValue := 45.0 // Base CPU usage

	for i := 0; i < count; i++ {
		// Normal variation
		data[i] = baseValue + rand.Float64()*10 - 5

		// Add some anomalies
		if rand.Float64() < 0.05 { // 5% chance of anomaly
			data[i] = baseValue + rand.Float64()*40 + 20 // Spike
		}
	}

	return data
}

func (h *IntelligenceHandler) generateTrendingData(count int, baseValue, trendRate float64) []float64 {
	data := make([]float64, count)

	for i := 0; i < count; i++ {
		// Add trend + noise
		data[i] = baseValue + float64(i)*trendRate/float64(count) + rand.Float64()*5 - 2.5
	}

	return data
}

func (h *IntelligenceHandler) generateSampleAnomalyScores(count int) []*models.AnomalyScore {
	scores := make([]*models.AnomalyScore, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		score := rand.Float64()
		scores[i] = &models.AnomalyScore{
			Timestamp:  now.Add(time.Duration(-count+i) * time.Minute),
			MetricName: "cpu_usage",
			Value:      45.0 + rand.Float64()*30,
			Score:      score,
			Threshold:  0.5,
			IsAnomaly:  score > 0.5,
			Confidence: score,
			Context: map[string]interface{}{
				"method": "statistical",
			},
			ModelID: "sample-model-id",
		}
	}

	return scores
}

func (h *IntelligenceHandler) generateSampleRootCauseAnalyses() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":           "rca-001",
			"incident_id":  "INC-001",
			"status":       "completed",
			"confidence":   0.87,
			"root_causes":  2,
			"completed_at": time.Now().Add(-2 * time.Hour),
		},
		{
			"id":          "rca-002",
			"incident_id": "INC-002",
			"status":      "in_progress",
			"confidence":  0.0,
			"root_causes": 0,
			"created_at":  time.Now().Add(-30 * time.Minute),
		},
	}
}

func (h *IntelligenceHandler) countAnomalies(scores []*models.AnomalyScore) int {
	count := 0
	for _, score := range scores {
		if score.IsAnomaly {
			count++
		}
	}
	return count
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
