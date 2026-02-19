package service

import (
	"math"
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
)

// ChurnPredictionService predicts user churn probability based on engagement patterns
type ChurnPredictionService struct {
	cacheClient    cache.Cache
	users          map[string]*churnUser
	predictions    map[string]*churnPrediction
	config         *ChurnConfig
	modelWeights   *churnModelWeights
	featureStats   *featureStatistics
	mu             sync.RWMutex
}

// ChurnConfig holds configuration for churn prediction
type ChurnConfig struct {
	Enabled              bool
	PredictionWindowDays int     // Days to look ahead for churn
	HighRiskThreshold    float64 // Score above which user is high risk
	MediumRiskThreshold  float64 // Score above which user is medium risk
	MinDataPoints        int     // Minimum events before prediction
	RecalculateInterval  time.Duration
	FeatureWeights       map[string]float64
}

// churnUser represents a user's engagement data for churn modeling
type churnUser struct {
	UserID            string
	FirstSeen         time.Time
	LastSeen          time.Time
	TotalImpressions  int64
	TotalClicks       int64
	TotalConversions  int64
	SessionCount      int
	AvgSessionLength  float64 // minutes
	DaysSinceLastSeen int
	EngagementTrend   float64 // positive = increasing, negative = decreasing
	DeviceTypes       map[string]int
	ActiveDays        map[string]bool // dates when user was active
	WeeklyActivity    []float64       // last 12 weeks of activity scores
	Features          map[string]float64
	UpdatedAt         time.Time
}

// churnPrediction holds the prediction result for a user
type churnPrediction struct {
	UserID           string
	ChurnProbability float64
	RiskLevel        string // "high", "medium", "low"
	TopFactors       []churnFactor
	PredictedAt      time.Time
	Confidence       float64
	DaysUntilChurn   int
	RecommendedAction string
}

// churnFactor represents a factor contributing to churn risk
type churnFactor struct {
	Name        string
	Impact      float64 // positive = increases churn risk
	Description string
	Value       float64
}

// churnModelWeights holds the learned model weights
type churnModelWeights struct {
	Intercept           float64
	DaysSinceLastSeen   float64
	EngagementTrend     float64
	SessionFrequency    float64
	ClickThroughRate    float64
	ConversionRate      float64
	DeviceDiversity     float64
	WeeklyConsistency   float64
	TenureDays          float64
	RecentActivityDrop  float64
	UpdatedAt           time.Time
}

// featureStatistics holds normalization stats for features
type featureStatistics struct {
	Means  map[string]float64
	StdDev map[string]float64
}

// ChurnPredictionResult is returned from prediction calls
type ChurnPredictionResult struct {
	UserID           string        `json:"user_id"`
	ChurnProbability float64       `json:"churn_probability"`
	RiskLevel        string        `json:"risk_level"`
	Confidence       float64       `json:"confidence"`
	TopFactors       []churnFactor `json:"top_factors"`
	DaysUntilChurn   int           `json:"days_until_churn"`
	RecommendedAction string       `json:"recommended_action"`
	PredictedAt      time.Time     `json:"predicted_at"`
}

// NewChurnPredictionService creates a new churn prediction service
func NewChurnPredictionService(c cache.Cache) *ChurnPredictionService {
	return &ChurnPredictionService{
		cacheClient: c,
		users:       make(map[string]*churnUser),
		predictions: make(map[string]*churnPrediction),
		config: &ChurnConfig{
			Enabled:              true,
			PredictionWindowDays: 30,
			HighRiskThreshold:    0.7,
			MediumRiskThreshold:  0.4,
			MinDataPoints:        5,
			RecalculateInterval:  24 * time.Hour,
			FeatureWeights: map[string]float64{
				"days_since_last_seen": 0.25,
				"engagement_trend":     0.20,
				"session_frequency":    0.15,
				"click_through_rate":   0.10,
				"weekly_consistency":   0.15,
				"recent_activity_drop": 0.15,
			},
		},
		modelWeights: &churnModelWeights{
			Intercept:          -1.5,
			DaysSinceLastSeen:  0.08,
			EngagementTrend:    -0.5,
			SessionFrequency:   -0.3,
			ClickThroughRate:   -0.4,
			ConversionRate:     -0.6,
			DeviceDiversity:    -0.2,
			WeeklyConsistency:  -0.35,
			TenureDays:         -0.01,
			RecentActivityDrop: 0.7,
			UpdatedAt:          time.Now(),
		},
		featureStats: &featureStatistics{
			Means: map[string]float64{
				"days_since_last_seen": 7.0,
				"session_frequency":    2.5,
				"click_through_rate":   0.02,
				"weekly_consistency":   0.6,
			},
			StdDev: map[string]float64{
				"days_since_last_seen": 10.0,
				"session_frequency":    2.0,
				"click_through_rate":   0.015,
				"weekly_consistency":   0.25,
			},
		},
	}
}

// RecordUserActivity records a user engagement event
func (s *ChurnPredictionService) RecordUserActivity(userID string, eventType string, metadata map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		user = &churnUser{
			UserID:       userID,
			FirstSeen:    time.Now(),
			DeviceTypes:  make(map[string]int),
			ActiveDays:   make(map[string]bool),
			WeeklyActivity: make([]float64, 12),
			Features:     make(map[string]float64),
		}
		s.users[userID] = user
	}

	now := time.Now()
	user.LastSeen = now
	user.UpdatedAt = now

	// Record active day
	dateKey := now.Format("2006-01-02")
	user.ActiveDays[dateKey] = true

	// Update counts based on event type
	switch eventType {
	case "impression":
		user.TotalImpressions++
	case "click":
		user.TotalClicks++
	case "conversion":
		user.TotalConversions++
	case "session_start":
		user.SessionCount++
	}

	// Track device type
	if deviceType, ok := metadata["device_type"].(string); ok {
		user.DeviceTypes[deviceType]++
	}

	// Update session length if provided
	if sessionLength, ok := metadata["session_length"].(float64); ok {
		// Running average
		if user.SessionCount > 0 {
			user.AvgSessionLength = (user.AvgSessionLength*float64(user.SessionCount-1) + sessionLength) / float64(user.SessionCount)
		} else {
			user.AvgSessionLength = sessionLength
		}
	}

	// Update weekly activity
	s.updateWeeklyActivity(user)
}

// updateWeeklyActivity updates the weekly activity scores
func (s *ChurnPredictionService) updateWeeklyActivity(user *churnUser) {
	now := time.Now()
	
	// Calculate activity for each of the last 12 weeks
	for i := 0; i < 12; i++ {
		weekStart := now.AddDate(0, 0, -7*(i+1))
		weekEnd := now.AddDate(0, 0, -7*i)
		
		activeDays := 0
		for dateStr := range user.ActiveDays {
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				continue
			}
			if date.After(weekStart) && date.Before(weekEnd) {
				activeDays++
			}
		}
		
		user.WeeklyActivity[i] = float64(activeDays) / 7.0 // Normalize to 0-1
	}

	// Calculate engagement trend (slope of recent activity)
	if len(user.WeeklyActivity) >= 4 {
		recent := (user.WeeklyActivity[0] + user.WeeklyActivity[1]) / 2
		older := (user.WeeklyActivity[2] + user.WeeklyActivity[3]) / 2
		user.EngagementTrend = recent - older
	}
}

// PredictChurn predicts churn probability for a user
func (s *ChurnPredictionService) PredictChurn(userID string) *ChurnPredictionResult {
	if !s.config.Enabled {
		return &ChurnPredictionResult{
			UserID:    userID,
			RiskLevel: "unknown",
		}
	}

	s.mu.RLock()
	user, exists := s.users[userID]
	s.mu.RUnlock()

	if !exists {
		return &ChurnPredictionResult{
			UserID:           userID,
			ChurnProbability: 0.5, // Unknown users get neutral score
			RiskLevel:        "unknown",
			Confidence:       0.0,
		}
	}

	// Calculate features
	features := s.calculateFeatures(user)
	
	// Apply logistic regression model
	logit := s.modelWeights.Intercept
	
	logit += s.modelWeights.DaysSinceLastSeen * s.normalizeFeature("days_since_last_seen", features["days_since_last_seen"])
	logit += s.modelWeights.EngagementTrend * features["engagement_trend"]
	logit += s.modelWeights.SessionFrequency * s.normalizeFeature("session_frequency", features["session_frequency"])
	logit += s.modelWeights.ClickThroughRate * s.normalizeFeature("click_through_rate", features["click_through_rate"])
	logit += s.modelWeights.WeeklyConsistency * s.normalizeFeature("weekly_consistency", features["weekly_consistency"])
	logit += s.modelWeights.RecentActivityDrop * features["recent_activity_drop"]
	logit += s.modelWeights.DeviceDiversity * features["device_diversity"]
	logit += s.modelWeights.TenureDays * features["tenure_days"]

	// Sigmoid function
	churnProb := 1.0 / (1.0 + math.Exp(-logit))

	// Determine risk level
	riskLevel := "low"
	if churnProb >= s.config.HighRiskThreshold {
		riskLevel = "high"
	} else if churnProb >= s.config.MediumRiskThreshold {
		riskLevel = "medium"
	}

	// Calculate confidence based on data quantity
	dataPoints := user.TotalImpressions + user.TotalClicks + int64(user.SessionCount)
	confidence := math.Min(1.0, float64(dataPoints)/float64(s.config.MinDataPoints*10))

	// Identify top factors
	topFactors := s.identifyTopFactors(features)

	// Estimate days until churn
	daysUntilChurn := s.estimateDaysUntilChurn(churnProb, features)

	// Recommend action
	recommendedAction := s.recommendAction(riskLevel, topFactors)

	result := &ChurnPredictionResult{
		UserID:           userID,
		ChurnProbability: churnProb,
		RiskLevel:        riskLevel,
		Confidence:       confidence,
		TopFactors:       topFactors,
		DaysUntilChurn:   daysUntilChurn,
		RecommendedAction: recommendedAction,
		PredictedAt:      time.Now(),
	}

	// Cache prediction
	s.mu.Lock()
	s.predictions[userID] = &churnPrediction{
		UserID:           userID,
		ChurnProbability: churnProb,
		RiskLevel:        riskLevel,
		TopFactors:       topFactors,
		PredictedAt:      time.Now(),
		Confidence:       confidence,
		DaysUntilChurn:   daysUntilChurn,
		RecommendedAction: recommendedAction,
	}
	s.mu.Unlock()

	return result
}

// calculateFeatures computes features for the model
func (s *ChurnPredictionService) calculateFeatures(user *churnUser) map[string]float64 {
	features := make(map[string]float64)
	now := time.Now()

	// Days since last seen
	features["days_since_last_seen"] = now.Sub(user.LastSeen).Hours() / 24

	// Engagement trend
	features["engagement_trend"] = user.EngagementTrend

	// Session frequency (sessions per week over last 4 weeks)
	tenureDays := now.Sub(user.FirstSeen).Hours() / 24
	if tenureDays > 0 {
		features["session_frequency"] = float64(user.SessionCount) / (tenureDays / 7)
	}

	// Click-through rate
	if user.TotalImpressions > 0 {
		features["click_through_rate"] = float64(user.TotalClicks) / float64(user.TotalImpressions)
	}

	// Conversion rate
	if user.TotalClicks > 0 {
		features["conversion_rate"] = float64(user.TotalConversions) / float64(user.TotalClicks)
	}

	// Device diversity (entropy of device usage)
	features["device_diversity"] = s.calculateDiversity(user.DeviceTypes)

	// Weekly consistency (how regularly user engages)
	features["weekly_consistency"] = s.calculateWeeklyConsistency(user.WeeklyActivity)

	// Recent activity drop (comparing last 2 weeks to previous 2 weeks)
	features["recent_activity_drop"] = s.calculateActivityDrop(user.WeeklyActivity)

	// Tenure in days
	features["tenure_days"] = tenureDays

	return features
}

// normalizeFeature applies z-score normalization
func (s *ChurnPredictionService) normalizeFeature(name string, value float64) float64 {
	mean := s.featureStats.Means[name]
	stdDev := s.featureStats.StdDev[name]
	if stdDev == 0 {
		return 0
	}
	return (value - mean) / stdDev
}

// calculateDiversity computes entropy-based diversity
func (s *ChurnPredictionService) calculateDiversity(counts map[string]int) float64 {
	total := 0
	for _, count := range counts {
		total += count
	}
	if total == 0 {
		return 0
	}

	entropy := 0.0
	for _, count := range counts {
		p := float64(count) / float64(total)
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	// Normalize to 0-1 (max entropy for 5 device types = log2(5) ≈ 2.32)
	return entropy / 2.32
}

// calculateWeeklyConsistency computes how consistent weekly activity is
func (s *ChurnPredictionService) calculateWeeklyConsistency(weeklyActivity []float64) float64 {
	if len(weeklyActivity) < 4 {
		return 0.5
	}

	// Calculate variance of weekly activity
	sum := 0.0
	for _, activity := range weeklyActivity[:4] {
		sum += activity
	}
	mean := sum / 4.0

	variance := 0.0
	for _, activity := range weeklyActivity[:4] {
		variance += (activity - mean) * (activity - mean)
	}
	variance /= 4.0

	// Lower variance = higher consistency
	// Convert to 0-1 scale where 1 = most consistent
	return 1.0 - math.Min(1.0, math.Sqrt(variance)*2)
}

// calculateActivityDrop computes recent vs older activity ratio
func (s *ChurnPredictionService) calculateActivityDrop(weeklyActivity []float64) float64 {
	if len(weeklyActivity) < 4 {
		return 0
	}

	recent := (weeklyActivity[0] + weeklyActivity[1]) / 2
	older := (weeklyActivity[2] + weeklyActivity[3]) / 2

	if older == 0 {
		return 0
	}

	drop := (older - recent) / older
	return math.Max(0, math.Min(1, drop)) // Clamp to 0-1
}

// identifyTopFactors identifies the most impactful churn factors
func (s *ChurnPredictionService) identifyTopFactors(features map[string]float64) []churnFactor {
	factors := []churnFactor{}

	// Days since last seen
	if features["days_since_last_seen"] > 7 {
		factors = append(factors, churnFactor{
			Name:        "inactivity",
			Impact:      features["days_since_last_seen"] / 30,
			Description: "User has been inactive for extended period",
			Value:       features["days_since_last_seen"],
		})
	}

	// Engagement trend
	if features["engagement_trend"] < -0.1 {
		factors = append(factors, churnFactor{
			Name:        "declining_engagement",
			Impact:      -features["engagement_trend"],
			Description: "User engagement is declining over time",
			Value:       features["engagement_trend"],
		})
	}

	// Activity drop
	if features["recent_activity_drop"] > 0.3 {
		factors = append(factors, churnFactor{
			Name:        "activity_drop",
			Impact:      features["recent_activity_drop"],
			Description: "Significant drop in recent activity",
			Value:       features["recent_activity_drop"],
		})
	}

	// Low consistency
	if features["weekly_consistency"] < 0.3 {
		factors = append(factors, churnFactor{
			Name:        "inconsistent_engagement",
			Impact:      1 - features["weekly_consistency"],
			Description: "User engagement is inconsistent",
			Value:       features["weekly_consistency"],
		})
	}

	// Sort by impact (descending)
	for i := 0; i < len(factors)-1; i++ {
		for j := i + 1; j < len(factors); j++ {
			if factors[j].Impact > factors[i].Impact {
				factors[i], factors[j] = factors[j], factors[i]
			}
		}
	}

	// Return top 3
	if len(factors) > 3 {
		return factors[:3]
	}
	return factors
}

// estimateDaysUntilChurn estimates when user might churn
func (s *ChurnPredictionService) estimateDaysUntilChurn(churnProb float64, features map[string]float64) int {
	if churnProb < 0.3 {
		return 90 // Low risk, estimate 90+ days
	}

	// Base estimation on probability and activity trend
	baseDays := int((1 - churnProb) * 60)
	
	// Adjust based on recent activity drop
	if features["recent_activity_drop"] > 0.5 {
		baseDays = int(float64(baseDays) * 0.5)
	}

	// Minimum 1 day
	if baseDays < 1 {
		baseDays = 1
	}

	return baseDays
}

// recommendAction provides actionable recommendation based on risk
func (s *ChurnPredictionService) recommendAction(riskLevel string, factors []churnFactor) string {
	switch riskLevel {
	case "high":
		if len(factors) > 0 && factors[0].Name == "inactivity" {
			return "Send re-engagement campaign with personalized offer"
		}
		return "Immediate outreach with exclusive offer or incentive"
	case "medium":
		return "Include in nurture campaign with relevant content"
	default:
		return "Continue standard engagement strategy"
	}
}

// GetHighRiskUsers returns users with high churn probability
func (s *ChurnPredictionService) GetHighRiskUsers(limit int) []*ChurnPredictionResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*ChurnPredictionResult, 0)

	for userID, prediction := range s.predictions {
		if prediction.RiskLevel == "high" {
			results = append(results, &ChurnPredictionResult{
				UserID:           userID,
				ChurnProbability: prediction.ChurnProbability,
				RiskLevel:        prediction.RiskLevel,
				Confidence:       prediction.Confidence,
				TopFactors:       prediction.TopFactors,
				DaysUntilChurn:   prediction.DaysUntilChurn,
				RecommendedAction: prediction.RecommendedAction,
				PredictedAt:      prediction.PredictedAt,
			})
		}
	}

	// Sort by churn probability descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].ChurnProbability > results[i].ChurnProbability {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if limit > 0 && len(results) > limit {
		return results[:limit]
	}
	return results
}

// GetChurnStats returns overall churn statistics
func (s *ChurnPredictionService) GetChurnStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalUsers := len(s.users)
	totalPredictions := len(s.predictions)
	
	highRisk := 0
	mediumRisk := 0
	lowRisk := 0
	totalChurnProb := 0.0

	for _, pred := range s.predictions {
		switch pred.RiskLevel {
		case "high":
			highRisk++
		case "medium":
			mediumRisk++
		case "low":
			lowRisk++
		}
		totalChurnProb += pred.ChurnProbability
	}

	avgChurnProb := 0.0
	if totalPredictions > 0 {
		avgChurnProb = totalChurnProb / float64(totalPredictions)
	}

	return map[string]interface{}{
		"total_users":           totalUsers,
		"total_predictions":     totalPredictions,
		"high_risk_users":       highRisk,
		"medium_risk_users":     mediumRisk,
		"low_risk_users":        lowRisk,
		"avg_churn_probability": avgChurnProb,
		"risk_distribution": map[string]float64{
			"high":   float64(highRisk) / math.Max(1, float64(totalPredictions)),
			"medium": float64(mediumRisk) / math.Max(1, float64(totalPredictions)),
			"low":    float64(lowRisk) / math.Max(1, float64(totalPredictions)),
		},
	}
}

// GetConfig returns the current configuration
func (s *ChurnPredictionService) GetConfig() *ChurnConfig {
	return s.config
}

// SetConfig updates the configuration
func (s *ChurnPredictionService) SetConfig(config *ChurnConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// BatchPredict predicts churn for multiple users
func (s *ChurnPredictionService) BatchPredict(userIDs []string) []*ChurnPredictionResult {
	results := make([]*ChurnPredictionResult, 0, len(userIDs))
	for _, userID := range userIDs {
		results = append(results, s.PredictChurn(userID))
	}
	return results
}
