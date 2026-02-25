package service

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
)

// PerformancePredictionService provides ML-based performance prediction
// for campaigns, creatives, and ad placements
type PerformancePredictionService struct {
	cache          cache.Cache
	historicalData sync.Map // key -> *PerformanceRecord
	predictions    sync.Map // predictionID -> *PredictionResult
	modelWeights   *PredictionModelWeights
	featureStats   *FeatureStatistics
	mu             sync.RWMutex
	config         PredictionConfig
}

// PredictionConfig holds configuration for performance prediction
type PredictionConfig struct {
	MinHistoricalSamples int                `json:"min_historical_samples"`
	ConfidenceThreshold  float64            `json:"confidence_threshold"`
	PredictionHorizon    int                `json:"prediction_horizon_hours"`
	ModelUpdateInterval  time.Duration      `json:"model_update_interval"`
	FeatureImportance    map[string]float64 `json:"feature_importance"`
	EnableRealTimeUpdate bool               `json:"enable_real_time_update"`
}

// PerformanceRecord holds historical performance data
type PerformanceRecord struct {
	EntityID    string             `json:"entity_id"`
	EntityType  string             `json:"entity_type"` // campaign, creative, placement
	Timestamp   time.Time          `json:"timestamp"`
	Impressions int64              `json:"impressions"`
	Clicks      int64              `json:"clicks"`
	Conversions int64              `json:"conversions"`
	Revenue     float64            `json:"revenue"`
	Spend       float64            `json:"spend"`
	CTR         float64            `json:"ctr"`
	CVR         float64            `json:"cvr"`
	CPC         float64            `json:"cpc"`
	CPM         float64            `json:"cpm"`
	ROAS        float64            `json:"roas"`
	Features    map[string]float64 `json:"features"`
}

// PredictionModelWeights holds the learned model weights
type PredictionModelWeights struct {
	CTRWeights      map[string]float64 `json:"ctr_weights"`
	CVRWeights      map[string]float64 `json:"cvr_weights"`
	ROASWeights     map[string]float64 `json:"roas_weights"`
	Intercepts      map[string]float64 `json:"intercepts"`
	LastTrainedAt   time.Time          `json:"last_trained_at"`
	TrainingSamples int                `json:"training_samples"`
}

// FeatureStatistics holds normalization statistics
type FeatureStatistics struct {
	Means   map[string]float64 `json:"means"`
	StdDevs map[string]float64 `json:"std_devs"`
	Mins    map[string]float64 `json:"mins"`
	Maxs    map[string]float64 `json:"maxs"`
}

// PredictionRequest represents a prediction request
type PredictionRequest struct {
	EntityID   string             `json:"entity_id"`
	EntityType string             `json:"entity_type"`
	Features   map[string]float64 `json:"features"`
	Context    PredictionContext  `json:"context"`
	Horizon    int                `json:"horizon"` // Hours ahead to predict
	Metrics    []string           `json:"metrics"` // Which metrics to predict
}

// PredictionContext provides context for prediction
type PredictionContext struct {
	TimeOfDay      string  `json:"time_of_day"`
	DayOfWeek      string  `json:"day_of_week"`
	DeviceType     string  `json:"device_type"`
	GeoRegion      string  `json:"geo_region"`
	AdFormat       string  `json:"ad_format"`
	BidPrice       float64 `json:"bid_price"`
	TargetAudience string  `json:"target_audience"`
	Seasonality    float64 `json:"seasonality"` // 0-1 seasonal factor
}

// PredictionResult holds the prediction output
type PredictionResult struct {
	ID                string                       `json:"id"`
	EntityID          string                       `json:"entity_id"`
	EntityType        string                       `json:"entity_type"`
	PredictedAt       time.Time                    `json:"predicted_at"`
	HorizonHours      int                          `json:"horizon_hours"`
	Predictions       map[string]*MetricPrediction `json:"predictions"`
	Confidence        float64                      `json:"confidence"`
	FeatureImportance map[string]float64           `json:"feature_importance"`
	Recommendations   []PredictionRecommendation   `json:"recommendations"`
}

// MetricPrediction holds prediction for a single metric
type MetricPrediction struct {
	Metric         string  `json:"metric"`
	PredictedValue float64 `json:"predicted_value"`
	LowerBound     float64 `json:"lower_bound"`
	UpperBound     float64 `json:"upper_bound"`
	Confidence     float64 `json:"confidence"`
	TrendDirection string  `json:"trend_direction"` // up, down, stable
	TrendStrength  float64 `json:"trend_strength"`  // 0-1
	HistoricalMean float64 `json:"historical_mean"`
	PercentChange  float64 `json:"percent_change"`
}

// PredictionRecommendation provides actionable recommendations
type PredictionRecommendation struct {
	Type        string  `json:"type"`
	Priority    string  `json:"priority"` // high, medium, low
	Description string  `json:"description"`
	Impact      float64 `json:"impact"`
	Action      string  `json:"action"`
}

// PerformanceForecast holds multi-period forecast
type PerformanceForecast struct {
	EntityID   string             `json:"entity_id"`
	EntityType string             `json:"entity_type"`
	StartTime  time.Time          `json:"start_time"`
	EndTime    time.Time          `json:"end_time"`
	Intervals  []ForecastInterval `json:"intervals"`
	Summary    ForecastSummary    `json:"summary"`
}

// ForecastInterval holds prediction for a time interval
type ForecastInterval struct {
	StartTime       time.Time             `json:"start_time"`
	EndTime         time.Time             `json:"end_time"`
	Predictions     map[string]float64    `json:"predictions"`
	ConfidenceBands map[string][2]float64 `json:"confidence_bands"` // [lower, upper]
}

// ForecastSummary provides summary statistics
type ForecastSummary struct {
	TotalImpressions int64   `json:"total_impressions"`
	TotalClicks      int64   `json:"total_clicks"`
	TotalConversions int64   `json:"total_conversions"`
	TotalRevenue     float64 `json:"total_revenue"`
	TotalSpend       float64 `json:"total_spend"`
	ExpectedCTR      float64 `json:"expected_ctr"`
	ExpectedCVR      float64 `json:"expected_cvr"`
	ExpectedROAS     float64 `json:"expected_roas"`
	RiskLevel        string  `json:"risk_level"`
}

// NewPerformancePredictionService creates a new prediction service
func NewPerformancePredictionService(c cache.Cache) *PerformancePredictionService {
	return &PerformancePredictionService{
		cache: c,
		modelWeights: &PredictionModelWeights{
			CTRWeights:  defaultCTRWeights(),
			CVRWeights:  defaultCVRWeights(),
			ROASWeights: defaultROASWeights(),
			Intercepts:  map[string]float64{"ctr": 0.02, "cvr": 0.01, "roas": 1.0},
		},
		featureStats: &FeatureStatistics{
			Means:   make(map[string]float64),
			StdDevs: make(map[string]float64),
			Mins:    make(map[string]float64),
			Maxs:    make(map[string]float64),
		},
		config: PredictionConfig{
			MinHistoricalSamples: 100,
			ConfidenceThreshold:  0.7,
			PredictionHorizon:    24,
			ModelUpdateInterval:  time.Hour,
			FeatureImportance: map[string]float64{
				"historical_ctr": 0.25,
				"historical_cvr": 0.20,
				"time_of_day":    0.10,
				"device_type":    0.10,
				"ad_format":      0.10,
				"bid_price":      0.15,
				"seasonality":    0.10,
			},
			EnableRealTimeUpdate: true,
		},
	}
}

func defaultCTRWeights() map[string]float64 {
	return map[string]float64{
		"historical_ctr":    0.40,
		"historical_clicks": 0.15,
		"time_factor":       0.10,
		"device_factor":     0.10,
		"format_factor":     0.10,
		"bid_price":         0.05,
		"seasonality":       0.10,
	}
}

func defaultCVRWeights() map[string]float64 {
	return map[string]float64{
		"historical_cvr":  0.35,
		"historical_ctr":  0.15,
		"landing_quality": 0.15,
		"audience_match":  0.15,
		"time_factor":     0.10,
		"device_factor":   0.10,
	}
}

func defaultROASWeights() map[string]float64 {
	return map[string]float64{
		"historical_roas": 0.30,
		"avg_order_value": 0.20,
		"cvr_prediction":  0.20,
		"cpc":             0.15,
		"margin":          0.15,
	}
}

// RecordPerformance records historical performance data
func (s *PerformancePredictionService) RecordPerformance(record *PerformanceRecord) {
	key := record.EntityID + ":" + record.Timestamp.Format("2006-01-02-15")

	// Calculate derived metrics
	if record.Impressions > 0 {
		record.CTR = float64(record.Clicks) / float64(record.Impressions)
		record.CVR = float64(record.Conversions) / float64(record.Impressions)
		record.CPM = (record.Spend / float64(record.Impressions)) * 1000
	}
	if record.Clicks > 0 {
		record.CPC = record.Spend / float64(record.Clicks)
	}
	if record.Spend > 0 {
		record.ROAS = record.Revenue / record.Spend
	}

	s.historicalData.Store(key, record)

	// Update feature statistics
	s.updateFeatureStats(record)
}

// Predict generates performance predictions
func (s *PerformancePredictionService) Predict(req PredictionRequest) (*PredictionResult, error) {
	// Get historical data
	historicalRecords := s.getHistoricalData(req.EntityID, req.EntityType)

	// Extract features
	features := s.extractFeatures(req, historicalRecords)

	// Calculate confidence based on data availability
	confidence := s.calculateConfidence(len(historicalRecords))

	// Generate predictions for requested metrics
	predictions := make(map[string]*MetricPrediction)

	for _, metric := range req.Metrics {
		pred := s.predictMetric(metric, features, historicalRecords, req.Context)
		predictions[metric] = pred
	}

	// Generate recommendations
	recommendations := s.generateRecommendations(predictions, features, req.Context)

	result := &PredictionResult{
		ID:                generateID(),
		EntityID:          req.EntityID,
		EntityType:        req.EntityType,
		PredictedAt:       time.Now(),
		HorizonHours:      req.Horizon,
		Predictions:       predictions,
		Confidence:        confidence,
		FeatureImportance: s.config.FeatureImportance,
		Recommendations:   recommendations,
	}

	// Store prediction
	s.predictions.Store(result.ID, result)

	return result, nil
}

// Forecast generates multi-period performance forecast
func (s *PerformancePredictionService) Forecast(entityID, entityType string, hours int) (*PerformanceForecast, error) {
	historicalRecords := s.getHistoricalData(entityID, entityType)

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(hours) * time.Hour)

	// Generate hourly intervals
	intervals := make([]ForecastInterval, 0, hours)

	var totalImp, totalClicks, totalConv int64
	var totalRev, totalSpend float64

	for h := 0; h < hours; h++ {
		intervalStart := startTime.Add(time.Duration(h) * time.Hour)
		intervalEnd := intervalStart.Add(time.Hour)

		// Context for this interval
		ctx := PredictionContext{
			TimeOfDay:   getTimeOfDay(intervalStart),
			DayOfWeek:   intervalStart.Weekday().String(),
			Seasonality: s.calculateSeasonality(intervalStart),
		}

		features := s.extractFeaturesWithContext(historicalRecords, ctx)

		// Predict metrics
		ctrPred := s.predictMetric("ctr", features, historicalRecords, ctx)
		cvrPred := s.predictMetric("cvr", features, historicalRecords, ctx)
		impPred := s.predictMetric("impressions", features, historicalRecords, ctx)

		interval := ForecastInterval{
			StartTime: intervalStart,
			EndTime:   intervalEnd,
			Predictions: map[string]float64{
				"ctr":         ctrPred.PredictedValue,
				"cvr":         cvrPred.PredictedValue,
				"impressions": impPred.PredictedValue,
			},
			ConfidenceBands: map[string][2]float64{
				"ctr":         {ctrPred.LowerBound, ctrPred.UpperBound},
				"cvr":         {cvrPred.LowerBound, cvrPred.UpperBound},
				"impressions": {impPred.LowerBound, impPred.UpperBound},
			},
		}

		intervals = append(intervals, interval)

		// Aggregate
		impressions := int64(impPred.PredictedValue)
		totalImp += impressions
		totalClicks += int64(float64(impressions) * ctrPred.PredictedValue)
		totalConv += int64(float64(impressions) * cvrPred.PredictedValue)
	}

	// Calculate summary
	summary := ForecastSummary{
		TotalImpressions: totalImp,
		TotalClicks:      totalClicks,
		TotalConversions: totalConv,
		TotalRevenue:     totalRev,
		TotalSpend:       totalSpend,
		RiskLevel:        s.assessRisk(totalImp, totalClicks, totalConv),
	}

	if totalImp > 0 {
		summary.ExpectedCTR = float64(totalClicks) / float64(totalImp)
		summary.ExpectedCVR = float64(totalConv) / float64(totalImp)
	}
	if totalSpend > 0 {
		summary.ExpectedROAS = totalRev / totalSpend
	}

	return &PerformanceForecast{
		EntityID:   entityID,
		EntityType: entityType,
		StartTime:  startTime,
		EndTime:    endTime,
		Intervals:  intervals,
		Summary:    summary,
	}, nil
}

// GetPredictionAccuracy calculates prediction accuracy
func (s *PerformancePredictionService) GetPredictionAccuracy(entityID string, lookbackHours int) map[string]float64 {
	accuracy := map[string]float64{
		"ctr_mae":  0.0,
		"cvr_mae":  0.0,
		"roas_mae": 0.0,
		"samples":  0.0,
	}

	// Compare predictions with actual results
	var ctrErrors, cvrErrors, roasErrors []float64

	cutoff := time.Now().Add(-time.Duration(lookbackHours) * time.Hour)

	s.predictions.Range(func(key, value any) bool {
		pred := value.(*PredictionResult)
		if pred.EntityID != entityID || pred.PredictedAt.Before(cutoff) {
			return true
		}

		// Get actual values for this time period
		actualKey := entityID + ":" + pred.PredictedAt.Format("2006-01-02-15")
		if actualVal, ok := s.historicalData.Load(actualKey); ok {
			actual := actualVal.(*PerformanceRecord)

			if ctrPred, ok := pred.Predictions["ctr"]; ok {
				ctrErrors = append(ctrErrors, math.Abs(ctrPred.PredictedValue-actual.CTR))
			}
			if cvrPred, ok := pred.Predictions["cvr"]; ok {
				cvrErrors = append(cvrErrors, math.Abs(cvrPred.PredictedValue-actual.CVR))
			}
			if roasPred, ok := pred.Predictions["roas"]; ok {
				roasErrors = append(roasErrors, math.Abs(roasPred.PredictedValue-actual.ROAS))
			}
		}

		return true
	})

	if len(ctrErrors) > 0 {
		accuracy["ctr_mae"] = mean(ctrErrors)
	}
	if len(cvrErrors) > 0 {
		accuracy["cvr_mae"] = mean(cvrErrors)
	}
	if len(roasErrors) > 0 {
		accuracy["roas_mae"] = mean(roasErrors)
	}
	accuracy["samples"] = float64(len(ctrErrors))

	return accuracy
}

// GetStats returns prediction service statistics
func (s *PerformancePredictionService) GetStats() map[string]any {
	stats := map[string]any{
		"total_records":          0,
		"total_predictions":      0,
		"model_last_trained":     s.modelWeights.LastTrainedAt,
		"training_samples":       s.modelWeights.TrainingSamples,
		"confidence_threshold":   s.config.ConfidenceThreshold,
		"prediction_horizon":     s.config.PredictionHorizon,
		"records_by_entity_type": make(map[string]int),
	}

	recordsByType := make(map[string]int)
	s.historicalData.Range(func(key, value any) bool {
		record := value.(*PerformanceRecord)
		stats["total_records"] = stats["total_records"].(int) + 1
		recordsByType[record.EntityType]++
		return true
	})
	stats["records_by_entity_type"] = recordsByType

	s.predictions.Range(func(key, value any) bool {
		stats["total_predictions"] = stats["total_predictions"].(int) + 1
		return true
	})

	return stats
}

// UpdateConfig updates prediction configuration
func (s *PerformancePredictionService) UpdateConfig(config PredictionConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// GetConfig returns current configuration
func (s *PerformancePredictionService) GetConfig() PredictionConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// Helper functions

func (s *PerformancePredictionService) getHistoricalData(entityID, entityType string) []*PerformanceRecord {
	var records []*PerformanceRecord

	s.historicalData.Range(func(key, value any) bool {
		record := value.(*PerformanceRecord)
		if record.EntityID == entityID && record.EntityType == entityType {
			records = append(records, record)
		}
		return true
	})

	// Sort by timestamp
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.Before(records[j].Timestamp)
	})

	return records
}

func (s *PerformancePredictionService) extractFeatures(req PredictionRequest, records []*PerformanceRecord) map[string]float64 {
	features := make(map[string]float64)

	// Add request features
	for k, v := range req.Features {
		features[k] = v
	}

	// Calculate historical features
	if len(records) > 0 {
		var sumCTR, sumCVR, sumROAS float64
		var totalImp, totalClicks int64

		for _, r := range records {
			sumCTR += r.CTR
			sumCVR += r.CVR
			sumROAS += r.ROAS
			totalImp += r.Impressions
			totalClicks += r.Clicks
		}

		n := float64(len(records))
		features["historical_ctr"] = sumCTR / n
		features["historical_cvr"] = sumCVR / n
		features["historical_roas"] = sumROAS / n
		features["historical_impressions"] = float64(totalImp) / n
		features["historical_clicks"] = float64(totalClicks) / n

		// Recent trend
		if len(records) >= 2 {
			recent := records[len(records)-1]
			previous := records[len(records)-2]
			features["ctr_trend"] = recent.CTR - previous.CTR
			features["cvr_trend"] = recent.CVR - previous.CVR
		}
	}

	// Context features
	features["time_factor"] = s.getTimeFactor(req.Context.TimeOfDay)
	features["device_factor"] = s.getDeviceFactor(req.Context.DeviceType)
	features["format_factor"] = s.getFormatFactor(req.Context.AdFormat)
	features["seasonality"] = req.Context.Seasonality
	features["bid_price"] = req.Context.BidPrice

	return features
}

func (s *PerformancePredictionService) extractFeaturesWithContext(records []*PerformanceRecord, ctx PredictionContext) map[string]float64 {
	features := make(map[string]float64)

	if len(records) > 0 {
		var sumCTR, sumCVR float64
		var totalImp int64

		for _, r := range records {
			sumCTR += r.CTR
			sumCVR += r.CVR
			totalImp += r.Impressions
		}

		n := float64(len(records))
		features["historical_ctr"] = sumCTR / n
		features["historical_cvr"] = sumCVR / n
		features["historical_impressions"] = float64(totalImp) / n
	}

	features["time_factor"] = s.getTimeFactor(ctx.TimeOfDay)
	features["seasonality"] = ctx.Seasonality

	return features
}

func (s *PerformancePredictionService) predictMetric(metric string, features map[string]float64, records []*PerformanceRecord, _ PredictionContext) *MetricPrediction {
	var weights map[string]float64
	var intercept float64

	switch metric {
	case "ctr":
		weights = s.modelWeights.CTRWeights
		intercept = s.modelWeights.Intercepts["ctr"]
	case "cvr":
		weights = s.modelWeights.CVRWeights
		intercept = s.modelWeights.Intercepts["cvr"]
	case "roas":
		weights = s.modelWeights.ROASWeights
		intercept = s.modelWeights.Intercepts["roas"]
	case "impressions":
		// Use historical average with time adjustment
		if avg, ok := features["historical_impressions"]; ok {
			timeFactor := features["time_factor"]
			predicted := avg * timeFactor
			return &MetricPrediction{
				Metric:         metric,
				PredictedValue: predicted,
				LowerBound:     predicted * 0.8,
				UpperBound:     predicted * 1.2,
				Confidence:     0.7,
				TrendDirection: "stable",
			}
		}
		return &MetricPrediction{Metric: metric, PredictedValue: 1000}
	default:
		return &MetricPrediction{Metric: metric}
	}

	// Linear prediction
	prediction := intercept
	for feature, weight := range weights {
		if val, ok := features[feature]; ok {
			prediction += weight * val
		}
	}

	// Ensure valid range
	prediction = math.Max(0, prediction)
	if metric == "ctr" || metric == "cvr" {
		prediction = math.Min(1.0, prediction)
	}

	// Calculate bounds
	stdDev := prediction * 0.15 // Approximate
	lowerBound := math.Max(0, prediction-1.96*stdDev)
	upperBound := prediction + 1.96*stdDev
	if metric == "ctr" || metric == "cvr" {
		upperBound = math.Min(1.0, upperBound)
	}

	// Trend analysis
	trend := "stable"
	trendStrength := 0.0
	if trendVal, ok := features[metric+"_trend"]; ok {
		if trendVal > 0.001 {
			trend = "up"
			trendStrength = math.Min(math.Abs(trendVal)*100, 1.0)
		} else if trendVal < -0.001 {
			trend = "down"
			trendStrength = math.Min(math.Abs(trendVal)*100, 1.0)
		}
	}

	// Historical mean
	historicalMean := 0.0
	if val, ok := features["historical_"+metric]; ok {
		historicalMean = val
	}

	// Percent change
	percentChange := 0.0
	if historicalMean > 0 {
		percentChange = ((prediction - historicalMean) / historicalMean) * 100
	}

	return &MetricPrediction{
		Metric:         metric,
		PredictedValue: prediction,
		LowerBound:     lowerBound,
		UpperBound:     upperBound,
		Confidence:     s.calculateMetricConfidence(len(records), metric),
		TrendDirection: trend,
		TrendStrength:  trendStrength,
		HistoricalMean: historicalMean,
		PercentChange:  percentChange,
	}
}

func (s *PerformancePredictionService) calculateConfidence(sampleCount int) float64 {
	if sampleCount >= s.config.MinHistoricalSamples {
		return 0.95
	}
	if sampleCount >= s.config.MinHistoricalSamples/2 {
		return 0.80
	}
	if sampleCount >= 10 {
		return 0.60
	}
	return 0.40
}

func (s *PerformancePredictionService) calculateMetricConfidence(sampleCount int, metric string) float64 {
	base := s.calculateConfidence(sampleCount)

	// Adjust based on metric complexity
	switch metric {
	case "ctr":
		return base * 0.95
	case "cvr":
		return base * 0.90
	case "roas":
		return base * 0.85
	default:
		return base * 0.80
	}
}

func (s *PerformancePredictionService) generateRecommendations(predictions map[string]*MetricPrediction, features map[string]float64, ctx PredictionContext) []PredictionRecommendation {
	var recommendations []PredictionRecommendation

	// CTR-based recommendations
	if ctrPred, ok := predictions["ctr"]; ok {
		if ctrPred.TrendDirection == "down" && ctrPred.TrendStrength > 0.5 {
			recommendations = append(recommendations, PredictionRecommendation{
				Type:        "creative_refresh",
				Priority:    "high",
				Description: "CTR is declining - consider refreshing creatives",
				Impact:      0.15,
				Action:      "Test new ad creatives or update messaging",
			})
		}
	}

	// CVR-based recommendations
	if cvrPred, ok := predictions["cvr"]; ok {
		if cvrPred.PredictedValue < 0.01 {
			recommendations = append(recommendations, PredictionRecommendation{
				Type:        "landing_page",
				Priority:    "high",
				Description: "Low conversion rate predicted - review landing page",
				Impact:      0.20,
				Action:      "Optimize landing page or improve audience targeting",
			})
		}
	}

	// Bid optimization
	if bidPrice, ok := features["bid_price"]; ok {
		if ctrPred, ctrOk := predictions["ctr"]; ctrOk && ctrPred.PredictedValue > 0.02 && bidPrice < 2.0 {
			recommendations = append(recommendations, PredictionRecommendation{
				Type:        "bid_increase",
				Priority:    "medium",
				Description: "Strong CTR expected - consider increasing bid",
				Impact:      0.10,
				Action:      "Increase bid price by 10-20% to win more impressions",
			})
		}
	}

	// Time-based recommendations
	timeFactor := s.getTimeFactor(ctx.TimeOfDay)
	if timeFactor < 0.8 {
		recommendations = append(recommendations, PredictionRecommendation{
			Type:        "day_parting",
			Priority:    "low",
			Description: "Current time slot has lower engagement",
			Impact:      0.05,
			Action:      "Consider reducing budget allocation for this time period",
		})
	}

	return recommendations
}

func (s *PerformancePredictionService) updateFeatureStats(record *PerformanceRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update running statistics for normalization
	for feature, value := range record.Features {
		if _, ok := s.featureStats.Means[feature]; !ok {
			s.featureStats.Means[feature] = value
			s.featureStats.Mins[feature] = value
			s.featureStats.Maxs[feature] = value
		} else {
			// Running mean update
			s.featureStats.Means[feature] = 0.99*s.featureStats.Means[feature] + 0.01*value
			if value < s.featureStats.Mins[feature] {
				s.featureStats.Mins[feature] = value
			}
			if value > s.featureStats.Maxs[feature] {
				s.featureStats.Maxs[feature] = value
			}
		}
	}
}

func (s *PerformancePredictionService) getTimeFactor(timeOfDay string) float64 {
	switch timeOfDay {
	case "morning":
		return 1.05
	case "afternoon":
		return 1.10
	case "evening":
		return 1.15
	case "night":
		return 0.80
	default:
		return 1.0
	}
}

func (s *PerformancePredictionService) getDeviceFactor(deviceType string) float64 {
	switch deviceType {
	case "mobile":
		return 1.10
	case "desktop":
		return 0.95
	case "tablet":
		return 1.0
	default:
		return 1.0
	}
}

func (s *PerformancePredictionService) getFormatFactor(adFormat string) float64 {
	switch adFormat {
	case "video":
		return 1.20
	case "native":
		return 1.15
	case "banner":
		return 1.0
	case "interstitial":
		return 1.25
	default:
		return 1.0
	}
}

func (s *PerformancePredictionService) calculateSeasonality(t time.Time) float64 {
	// Simple seasonality based on day of week
	weekday := t.Weekday()
	switch weekday {
	case time.Saturday, time.Sunday:
		return 0.85
	case time.Friday:
		return 1.10
	default:
		return 1.0
	}
}

func (s *PerformancePredictionService) assessRisk(impressions, clicks, conversions int64) string {
	if impressions < 100 {
		return "high"
	}

	ctr := float64(clicks) / float64(impressions)
	cvr := float64(conversions) / float64(impressions)

	if ctr < 0.005 || cvr < 0.001 {
		return "high"
	}
	if ctr < 0.01 || cvr < 0.005 {
		return "medium"
	}
	return "low"
}

func getTimeOfDay(t time.Time) string {
	hour := t.Hour()
	switch {
	case hour >= 6 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 17:
		return "afternoon"
	case hour >= 17 && hour < 21:
		return "evening"
	default:
		return "night"
	}
}

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
