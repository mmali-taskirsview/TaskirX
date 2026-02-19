package service

import (
	"fmt"
	"math"
	"strings"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// AudienceModelingService provides lookalike audience expansion,
// audience suppression, and propensity-based bid adjustments.
type AudienceModelingService struct {
	cache cache.Cache
}

// NewAudienceModelingService creates a new audience modeling service
func NewAudienceModelingService(cache cache.Cache) *AudienceModelingService {
	return &AudienceModelingService{cache: cache}
}

// EvaluateAudienceModeling applies audience modeling logic to a bid request
// Returns a result with multiplier adjustments and audience classification
func (a *AudienceModelingService) EvaluateAudienceModeling(campaign *model.Campaign, req *model.BidRequest) model.AudienceModelingResult {
	result := model.AudienceModelingResult{
		Matched:    false,
		Suppressed: false,
		Multiplier: 1.0,
	}

	pg := campaign.Targeting.PerformanceGoals
	if pg == nil || pg.AudienceModeling == nil {
		return result
	}

	am := pg.AudienceModeling

	// 1. Audience Suppression (check first - can block the bid entirely)
	if am.SuppressionEnabled {
		suppressed, reason := a.evaluateSuppression(campaign, req, am)
		if suppressed {
			result.Suppressed = true
			result.Multiplier = 0
			result.Reason = reason
			return result
		}
	}

	// 2. Check if user is in seed audience (highest value)
	userSegments := a.getUserSegments(req)
	isSeedUser := a.isSeedAudience(userSegments, am.SeedSegments)

	if isSeedUser {
		result.Matched = true
		result.AudienceTier = "seed"
		result.Multiplier = 1.5 // Seed users are highest value
		result.Reason = "seed_audience_match"
		return result
	}

	// 3. Lookalike Expansion
	if am.LookalikeEnabled && len(am.SeedSegments) > 0 {
		lookalikeResult := a.evaluateLookalike(campaign, req, am, userSegments)
		if lookalikeResult.IsLookalike {
			result.Matched = true
			result.IsLookalike = true
			result.SimilarityScore = lookalikeResult.SimilarityScore
			result.AudienceTier = lookalikeResult.AudienceTier
			result.Multiplier = lookalikeResult.Multiplier
			result.Reason = lookalikeResult.Reason
			return result
		}
	}

	// 4. Propensity Scoring
	if am.ScoringEnabled {
		propensityResult := a.evaluatePropensityScore(campaign, req, am)
		if propensityResult.PropensityScore > 0 {
			result.Matched = true
			result.PropensityScore = propensityResult.PropensityScore
			result.Multiplier = propensityResult.Multiplier
			result.AudienceTier = propensityResult.AudienceTier
			result.Reason = propensityResult.Reason
			return result
		}
	}

	// 5. Prospecting (no match but still bidding)
	result.AudienceTier = "prospecting"
	result.Multiplier = 0.8 // Lower bid for unknown audiences
	result.Reason = "prospecting_no_audience_match"

	return result
}

// evaluateSuppression checks if the user should be suppressed (excluded) from targeting
func (a *AudienceModelingService) evaluateSuppression(campaign *model.Campaign, req *model.BidRequest, am *model.AudienceModeling) (bool, string) {
	userID := req.User.ID
	if userID == "" {
		return false, ""
	}

	// Check suppression segments
	if len(am.SuppressionSegments) > 0 {
		userSegments := a.getUserSegments(req)
		for _, suppressSeg := range am.SuppressionSegments {
			for _, userSeg := range userSegments {
				if strings.EqualFold(userSeg, suppressSeg) {
					return true, "suppressed_segment:" + suppressSeg
				}
			}
		}
	}

	// Check suppression events (e.g., user already purchased)
	if len(am.SuppressionEvents) > 0 {
		windowDays := am.SuppressionWindowDays
		if windowDays <= 0 {
			windowDays = 30
		}

		for _, eventType := range am.SuppressionEvents {
			hasEvent, err := a.cache.HasUserEvent(userID, campaign.ID, eventType)
			if err == nil && hasEvent {
				return true, "suppressed_event:" + eventType
			}
		}
	}

	// Check conversion suppression (don't target users who already converted)
	if req.Context != nil {
		if converted, ok := req.Context["has_converted"].(bool); ok && converted {
			return true, "suppressed_already_converted"
		}
	}

	return false, ""
}

// evaluateLookalike determines if a user is similar to seed audience and calculates
// a similarity score and bid multiplier
func (a *AudienceModelingService) evaluateLookalike(campaign *model.Campaign, req *model.BidRequest, am *model.AudienceModeling, userSegments []string) model.AudienceModelingResult {
	result := model.AudienceModelingResult{
		Multiplier: 1.0,
	}

	if len(am.SeedSegments) == 0 {
		return result
	}

	// Calculate similarity score based on feature overlap
	similarityScore := a.calculateSimilarityScore(req, am, userSegments)

	threshold := am.SimilarityThreshold
	if threshold <= 0 {
		threshold = 0.7
	}

	// Expansion factor widens the threshold (higher expansion = lower threshold)
	expansion := am.LookalikeExpansion
	if expansion > 0 && expansion <= 10 {
		// Expansion of 1 = strict (threshold stays), 10 = very broad (threshold * 0.3)
		adjustedThreshold := threshold * (1.0 - (expansion-1)*0.08)
		if adjustedThreshold < 0.2 {
			adjustedThreshold = 0.2
		}
		threshold = adjustedThreshold
	}

	if similarityScore >= threshold {
		result.IsLookalike = true
		result.SimilarityScore = similarityScore

		// Determine tier based on similarity
		boost := am.LookalikeBoost
		if boost <= 0 {
			boost = 1.3
		}

		if similarityScore >= 0.9 {
			result.AudienceTier = "lookalike_high"
			result.Multiplier = boost * 1.2 // Premium lookalikes
		} else if similarityScore >= 0.7 {
			result.AudienceTier = "lookalike_medium"
			result.Multiplier = boost
		} else {
			result.AudienceTier = "lookalike_low"
			result.Multiplier = boost * 0.8
		}

		result.Reason = fmt.Sprintf("lookalike_match_score=%.2f_tier=%s", similarityScore, result.AudienceTier)
	}

	return result
}

// calculateSimilarityScore computes how similar a user is to the seed audience
// Uses feature-based similarity across demographics, interests, behavior, geo, device
func (a *AudienceModelingService) calculateSimilarityScore(req *model.BidRequest, am *model.AudienceModeling, userSegments []string) float64 {
	features := am.LookalikeFeatures
	if len(features) == 0 {
		features = []string{"demographics", "interests", "behavior", "geo", "device"}
	}

	totalScore := 0.0
	featureCount := 0

	for _, feature := range features {
		switch strings.ToLower(feature) {
		case "demographics":
			score := a.demographicSimilarity(req)
			totalScore += score
			featureCount++

		case "interests":
			score := a.interestSimilarity(req, userSegments, am.SeedSegments)
			totalScore += score
			featureCount++

		case "behavior":
			score := a.behavioralSimilarity(req)
			totalScore += score
			featureCount++

		case "geo":
			score := a.geoSimilarity(req)
			totalScore += score
			featureCount++

		case "device":
			score := a.deviceSimilarity(req)
			totalScore += score
			featureCount++
		}
	}

	if featureCount == 0 {
		return 0
	}

	return totalScore / float64(featureCount)
}

// demographicSimilarity scores demographic match
func (a *AudienceModelingService) demographicSimilarity(req *model.BidRequest) float64 {
	score := 0.5 // Base score

	if req.Context != nil {
		// Age bracket matching
		if age, ok := req.Context["age_bracket"].(string); ok {
			// Check against seed audience demographics stored in cache
			seedDemKey := "seed_demo_age"
			seedAge, err := a.cache.Get(seedDemKey)
			if err == nil && seedAge != "" && strings.Contains(seedAge, age) {
				score += 0.3
			}
		}

		// Income level matching
		if income, ok := req.Context["income_level"].(string); ok {
			seedIncKey := "seed_demo_income"
			seedInc, err := a.cache.Get(seedIncKey)
			if err == nil && seedInc != "" && strings.Contains(seedInc, income) {
				score += 0.2
			}
		}
	}

	// User age if available
	if req.User.Age > 0 {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

// interestSimilarity scores interest/segment overlap with seed audience
func (a *AudienceModelingService) interestSimilarity(req *model.BidRequest, userSegments, seedSegments []string) float64 {
	if len(userSegments) == 0 || len(seedSegments) == 0 {
		return 0.3 // Low base score with no data
	}

	// Jaccard similarity: |intersection| / |union|
	intersectionCount := 0
	unionSet := make(map[string]bool)

	for _, seg := range seedSegments {
		unionSet[strings.ToLower(seg)] = true
	}
	for _, seg := range userSegments {
		lower := strings.ToLower(seg)
		if unionSet[lower] {
			intersectionCount++
		}
		unionSet[lower] = true
	}

	if len(unionSet) == 0 {
		return 0
	}

	return float64(intersectionCount) / float64(len(unionSet))
}

// behavioralSimilarity scores behavioral similarity based on engagement patterns
func (a *AudienceModelingService) behavioralSimilarity(req *model.BidRequest) float64 {
	score := 0.4 // Base score

	if req.Context != nil {
		// Session engagement signals
		if sessionDuration, ok := req.Context["session_duration"].(float64); ok {
			if sessionDuration > 300 { // >5 min session
				score += 0.2
			} else if sessionDuration > 60 { // >1 min
				score += 0.1
			}
		}

		// Pages viewed
		if pageViews, ok := req.Context["pages_viewed"].(float64); ok {
			if pageViews > 5 {
				score += 0.2
			} else if pageViews > 2 {
				score += 0.1
			}
		}

		// Previous engagement with similar campaigns
		if engagementScore, ok := req.Context["engagement_score"].(float64); ok {
			score += engagementScore * 0.3
		}

		// Return visitor
		if isReturn, ok := req.Context["return_visitor"].(bool); ok && isReturn {
			score += 0.15
		}
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

// geoSimilarity scores geographic similarity to seed audience
func (a *AudienceModelingService) geoSimilarity(req *model.BidRequest) float64 {
	score := 0.3 // Base score

	if req.User.Country != "" {
		// Check if user's country matches seed audience's top geos
		seedGeoKey := "seed_geo_countries"
		seedGeos, err := a.cache.Get(seedGeoKey)
		if err == nil && seedGeos != "" {
			if strings.Contains(strings.ToLower(seedGeos), strings.ToLower(req.User.Country)) {
				score += 0.5
			}
		}
	}

	// City-level matching for higher precision
	if req.Device.Geo.City != "" {
		seedCityKey := "seed_geo_cities"
		seedCities, err := a.cache.Get(seedCityKey)
		if err == nil && seedCities != "" {
			if strings.Contains(strings.ToLower(seedCities), strings.ToLower(req.Device.Geo.City)) {
				score += 0.3
			}
		}
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

// deviceSimilarity scores device profile similarity
func (a *AudienceModelingService) deviceSimilarity(req *model.BidRequest) float64 {
	score := 0.4 // Base score

	// Device type matching
	seedDeviceKey := "seed_device_types"
	seedDevices, err := a.cache.Get(seedDeviceKey)
	if err == nil && seedDevices != "" {
		if strings.Contains(strings.ToLower(seedDevices), strings.ToLower(req.Device.Type)) {
			score += 0.3
		}
	}

	// OS matching
	seedOSKey := "seed_device_os"
	seedOS, err := a.cache.Get(seedOSKey)
	if err == nil && seedOS != "" {
		if strings.Contains(strings.ToLower(seedOS), strings.ToLower(req.Device.OS)) {
			score += 0.2
		}
	}

	// Browser matching
	if req.Device.Browser != "" {
		seedBrowserKey := "seed_device_browsers"
		seedBrowsers, err := a.cache.Get(seedBrowserKey)
		if err == nil && seedBrowsers != "" {
			if strings.Contains(strings.ToLower(seedBrowsers), strings.ToLower(req.Device.Browser)) {
				score += 0.1
			}
		}
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

// evaluatePropensityScore calculates a user's propensity score and maps it to a bid multiplier
func (a *AudienceModelingService) evaluatePropensityScore(campaign *model.Campaign, req *model.BidRequest, am *model.AudienceModeling) model.AudienceModelingResult {
	result := model.AudienceModelingResult{
		Multiplier: 1.0,
	}

	// Calculate propensity based on scoring model type
	var propensity float64
	switch strings.ToLower(am.ScoringModel) {
	case "propensity":
		propensity = a.calculateConversionPropensity(req)
	case "ltv":
		propensity = a.calculateLTVPropensity(req)
	case "churn_risk":
		propensity = a.calculateChurnRiskPropensity(req)
	default:
		propensity = a.calculateConversionPropensity(req)
	}

	result.PropensityScore = propensity

	// Check minimum score threshold
	if am.MinScore > 0 && propensity < am.MinScore {
		result.Reason = fmt.Sprintf("propensity_below_threshold_%.2f<%.2f", propensity, am.MinScore)
		result.Multiplier = 0.5 // Low multiplier for below-threshold users
		return result
	}

	// Apply score-to-bid mapping if configured
	if len(am.ScoreBidMapping) > 0 {
		for _, mapping := range am.ScoreBidMapping {
			if propensity >= mapping.MinScore && propensity < mapping.MaxScore {
				result.Multiplier = mapping.Multiplier
				break
			}
		}
	} else {
		// Default linear mapping: propensity 0-1 maps to multiplier 0.5-2.0
		result.Multiplier = 0.5 + propensity*1.5
	}

	// Determine tier
	if propensity >= 0.8 {
		result.AudienceTier = "high_propensity"
	} else if propensity >= 0.5 {
		result.AudienceTier = "medium_propensity"
	} else {
		result.AudienceTier = "low_propensity"
	}

	result.Reason = fmt.Sprintf("propensity_%s_score=%.2f", am.ScoringModel, propensity)
	return result
}

// calculateConversionPropensity estimates the likelihood a user will convert
func (a *AudienceModelingService) calculateConversionPropensity(req *model.BidRequest) float64 {
	score := 0.3 // Base propensity

	if req.Context != nil {
		// Previous engagement signals
		if engScore, ok := req.Context["engagement_score"].(float64); ok {
			score += engScore * 0.2
		}

		// Intent signals
		if intent, ok := req.Context["intent_score"].(float64); ok {
			score += intent * 0.2
		}

		// Return visitor
		if isReturn, ok := req.Context["return_visitor"].(bool); ok && isReturn {
			score += 0.1
		}

		// Cart abandoner (very high intent)
		if abandon, ok := req.Context["cart_abandoner"].(bool); ok && abandon {
			score += 0.2
		}

		// High-value segments
		if segments, ok := req.Context["user_segments"].([]interface{}); ok {
			for _, seg := range segments {
				if segStr, ok := seg.(string); ok {
					lower := strings.ToLower(segStr)
					if strings.Contains(lower, "high_intent") || strings.Contains(lower, "converter") {
						score += 0.15
						break
					}
				}
			}
		}
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

// calculateLTVPropensity estimates the lifetime value potential of a user
func (a *AudienceModelingService) calculateLTVPropensity(req *model.BidRequest) float64 {
	score := 0.2 // Base LTV propensity

	if req.Context != nil {
		// Purchase history
		if purchaseCount, ok := req.Context["purchase_count"].(float64); ok {
			if purchaseCount > 5 {
				score += 0.3 // Repeat buyer
			} else if purchaseCount > 0 {
				score += 0.15
			}
		}

		// Average order value
		if aov, ok := req.Context["avg_order_value"].(float64); ok {
			if aov > 100 {
				score += 0.2
			} else if aov > 50 {
				score += 0.1
			}
		}

		// Recency of last purchase
		if daysSincePurchase, ok := req.Context["days_since_purchase"].(float64); ok {
			if daysSincePurchase < 7 {
				score += 0.15
			} else if daysSincePurchase < 30 {
				score += 0.1
			}
		}

		// Customer tenure
		if tenureDays, ok := req.Context["customer_tenure_days"].(float64); ok {
			if tenureDays > 365 {
				score += 0.1
			}
		}
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}

// calculateChurnRiskPropensity estimates the risk of user churning (inverse: higher = less likely to churn)
func (a *AudienceModelingService) calculateChurnRiskPropensity(req *model.BidRequest) float64 {
	// Start with high retention (low churn risk) and reduce based on signals
	score := 0.7

	if req.Context != nil {
		// Decreasing engagement signals churn risk
		if engTrend, ok := req.Context["engagement_trend"].(string); ok {
			switch engTrend {
			case "increasing":
				score += 0.2
			case "stable":
				score += 0.0
			case "decreasing":
				score -= 0.3
			}
		}

		// Days since last activity
		if daysSinceActivity, ok := req.Context["days_since_activity"].(float64); ok {
			if daysSinceActivity > 30 {
				score -= 0.3 // High churn risk
			} else if daysSinceActivity > 14 {
				score -= 0.15
			}
		}

		// Subscription status
		if subStatus, ok := req.Context["subscription_status"].(string); ok {
			if subStatus == "cancelled" || subStatus == "expired" {
				score -= 0.4
			}
		}
	}

	if score < 0 {
		score = 0
	}
	if score > 1.0 {
		score = 1.0
	}
	return score
}

// isSeedAudience checks if user is part of the seed audience
func (a *AudienceModelingService) isSeedAudience(userSegments, seedSegments []string) bool {
	if len(seedSegments) == 0 {
		return false
	}

	for _, seed := range seedSegments {
		for _, userSeg := range userSegments {
			if strings.EqualFold(seed, userSeg) {
				return true
			}
		}
	}
	return false
}

// getUserSegments extracts user segments from request context
func (a *AudienceModelingService) getUserSegments(req *model.BidRequest) []string {
	segments := []string{}

	if req.Context != nil {
		if segs, ok := req.Context["user_segments"].([]interface{}); ok {
			for _, seg := range segs {
				if segStr, ok := seg.(string); ok {
					segments = append(segments, segStr)
				}
			}
		}
	}

	// Also include interest categories as segments
	if len(req.User.Categories) > 0 {
		segments = append(segments, req.User.Categories...)
	}

	return segments
}

// Ensure math import is used
var _ = math.Abs
