package service

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// AttributionService provides multi-touch attribution models for conversion credit assignment.
// Supports: last_click, first_click, linear, time_decay, position_based
type AttributionService struct {
	cache cache.Cache
}

// NewAttributionService creates a new attribution service
func NewAttributionService(cache cache.Cache) *AttributionService {
	return &AttributionService{cache: cache}
}

// CalculateAttribution computes attribution credits for a conversion based on the specified model.
// Parameters:
//   - userID: the converting user
//   - campaignID: the campaign to attribute (empty string = all campaigns)
//   - modelType: one of "last_click", "first_click", "linear", "time_decay", "position_based"
//   - halfLifeHours: half-life for time_decay model (default 168 = 7 days if 0)
//
// Returns a slice of AttributionCredit with normalized weights summing to 1.0
func (a *AttributionService) CalculateAttribution(userID, campaignID, modelType string, halfLifeHours float64) ([]model.AttributionCredit, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required for attribution")
	}

	// Get touchpoints from cache
	touchpoints, err := a.cache.GetTouchpoints(userID, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get touchpoints: %w", err)
	}

	if len(touchpoints) == 0 {
		return nil, nil
	}

	// Sort touchpoints by timestamp (ascending)
	sort.Slice(touchpoints, func(i, j int) bool {
		return touchpoints[i].Timestamp.Before(touchpoints[j].Timestamp)
	})

	// Assign positions
	for i := range touchpoints {
		touchpoints[i].Position = i + 1
	}

	// Normalize model type
	modelType = strings.ToLower(strings.TrimSpace(modelType))
	if modelType == "" {
		modelType = "last_click"
	}

	switch modelType {
	case "last_click", "last_touch":
		return a.lastClickAttribution(touchpoints), nil
	case "first_click", "first_touch":
		return a.firstClickAttribution(touchpoints), nil
	case "linear":
		return a.linearAttribution(touchpoints), nil
	case "time_decay":
		if halfLifeHours <= 0 {
			halfLifeHours = 168 // Default 7 days
		}
		return a.timeDecayAttribution(touchpoints, halfLifeHours), nil
	case "position_based":
		return a.positionBasedAttribution(touchpoints), nil
	default:
		return a.lastClickAttribution(touchpoints), nil
	}
}

// lastClickAttribution assigns 100% credit to the last touchpoint before conversion.
// Most common model, used as the default in most ad platforms.
func (a *AttributionService) lastClickAttribution(touchpoints []model.Touchpoint) []model.AttributionCredit {
	credits := make([]model.AttributionCredit, len(touchpoints))
	for i, tp := range touchpoints {
		credit := 0.0
		if i == len(touchpoints)-1 {
			credit = 1.0
		}
		credits[i] = model.AttributionCredit{
			Touchpoint: tp,
			Credit:     credit,
			Model:      "last_click",
		}
	}
	return credits
}

// firstClickAttribution assigns 100% credit to the first touchpoint in the journey.
// Useful for understanding acquisition channels and top-of-funnel performance.
func (a *AttributionService) firstClickAttribution(touchpoints []model.Touchpoint) []model.AttributionCredit {
	credits := make([]model.AttributionCredit, len(touchpoints))
	for i, tp := range touchpoints {
		credit := 0.0
		if i == 0 {
			credit = 1.0
		}
		credits[i] = model.AttributionCredit{
			Touchpoint: tp,
			Credit:     credit,
			Model:      "first_click",
		}
	}
	return credits
}

// linearAttribution distributes credit equally across all touchpoints.
// Each touchpoint receives 1/N of the total credit, where N is the number of touchpoints.
func (a *AttributionService) linearAttribution(touchpoints []model.Touchpoint) []model.AttributionCredit {
	n := len(touchpoints)
	if n == 0 {
		return nil
	}

	equalCredit := 1.0 / float64(n)
	credits := make([]model.AttributionCredit, n)
	for i, tp := range touchpoints {
		credits[i] = model.AttributionCredit{
			Touchpoint: tp,
			Credit:     equalCredit,
			Model:      "linear",
		}
	}
	return credits
}

// timeDecayAttribution assigns credit based on how close each touchpoint is to the conversion.
// Uses exponential decay: credit = 2^(-t/halfLife) where t is time before conversion.
// More recent touchpoints receive proportionally more credit.
func (a *AttributionService) timeDecayAttribution(touchpoints []model.Touchpoint, halfLifeHours float64) []model.AttributionCredit {
	n := len(touchpoints)
	if n == 0 {
		return nil
	}

	// Reference time is the last touchpoint (closest to conversion)
	conversionTime := touchpoints[n-1].Timestamp

	// Calculate raw weights using exponential decay
	rawWeights := make([]float64, n)
	totalWeight := 0.0

	for i, tp := range touchpoints {
		// Time difference in hours from this touchpoint to conversion
		hoursBeforeConversion := conversionTime.Sub(tp.Timestamp).Hours()
		if hoursBeforeConversion < 0 {
			hoursBeforeConversion = 0
		}

		// Exponential decay: weight = 2^(-t/halfLife)
		weight := math.Pow(2, -hoursBeforeConversion/halfLifeHours)
		rawWeights[i] = weight
		totalWeight += weight
	}

	// Normalize weights to sum to 1.0
	credits := make([]model.AttributionCredit, n)
	for i, tp := range touchpoints {
		normalizedCredit := 0.0
		if totalWeight > 0 {
			normalizedCredit = rawWeights[i] / totalWeight
		}
		credits[i] = model.AttributionCredit{
			Touchpoint: tp,
			Credit:     normalizedCredit,
			Model:      "time_decay",
		}
	}

	return credits
}

// positionBasedAttribution (U-shaped) assigns 40% to first, 40% to last, and distributes
// the remaining 20% equally among middle touchpoints.
// This model values both the introduction and closing touchpoints.
func (a *AttributionService) positionBasedAttribution(touchpoints []model.Touchpoint) []model.AttributionCredit {
	n := len(touchpoints)
	if n == 0 {
		return nil
	}

	credits := make([]model.AttributionCredit, n)

	if n == 1 {
		credits[0] = model.AttributionCredit{
			Touchpoint: touchpoints[0],
			Credit:     1.0,
			Model:      "position_based",
		}
		return credits
	}

	if n == 2 {
		credits[0] = model.AttributionCredit{
			Touchpoint: touchpoints[0],
			Credit:     0.5,
			Model:      "position_based",
		}
		credits[1] = model.AttributionCredit{
			Touchpoint: touchpoints[1],
			Credit:     0.5,
			Model:      "position_based",
		}
		return credits
	}

	// First: 40%, Last: 40%, Middle: 20% split equally
	middleCredit := 0.20 / float64(n-2)

	for i, tp := range touchpoints {
		var credit float64
		switch {
		case i == 0:
			credit = 0.40
		case i == n-1:
			credit = 0.40
		default:
			credit = middleCredit
		}
		credits[i] = model.AttributionCredit{
			Touchpoint: tp,
			Credit:     credit,
			Model:      "position_based",
		}
	}

	return credits
}

// GetAttributionSummary returns a summary of attribution credit by campaign for a user's conversion.
// Aggregates credits across all touchpoints by campaign ID.
func (a *AttributionService) GetAttributionSummary(userID, modelType string, halfLifeHours float64) (map[string]float64, error) {
	credits, err := a.CalculateAttribution(userID, "", modelType, halfLifeHours)
	if err != nil {
		return nil, err
	}

	summary := make(map[string]float64)
	for _, credit := range credits {
		summary[credit.Touchpoint.CampaignID] += credit.Credit
	}

	return summary, nil
}

// GetAttributionBidAdjustment returns a bid multiplier based on attribution model insights.
// Campaigns that historically contribute more to conversions receive higher multipliers.
func (a *AttributionService) GetAttributionBidAdjustment(campaignID, userID, modelType string, halfLifeHours float64) float64 {
	if userID == "" || campaignID == "" {
		return 1.0
	}

	credits, err := a.CalculateAttribution(userID, "", modelType, halfLifeHours)
	if err != nil || len(credits) == 0 {
		return 1.0
	}

	// Sum credit for this specific campaign
	campaignCredit := 0.0
	totalCredit := 0.0
	for _, credit := range credits {
		totalCredit += credit.Credit
		if credit.Touchpoint.CampaignID == campaignID {
			campaignCredit += credit.Credit
		}
	}

	if totalCredit <= 0 {
		return 1.0
	}

	// Convert attribution share to bid multiplier
	// If campaign has above-average attribution, boost bid
	avgCredit := totalCredit / float64(len(credits))
	if avgCredit <= 0 {
		return 1.0
	}

	ratio := campaignCredit / avgCredit
	// Map ratio to a multiplier in range [0.5, 2.0]
	multiplier := 0.5 + ratio*0.5
	if multiplier > 2.0 {
		multiplier = 2.0
	}
	if multiplier < 0.5 {
		multiplier = 0.5
	}

	return multiplier
}

// RecordConversionTouchpoint records a touchpoint event for later attribution
func (a *AttributionService) RecordConversionTouchpoint(userID, campaignID, touchpointType, requestID string, ttlDays int) error {
	if ttlDays <= 0 {
		ttlDays = 30 // Default 30-day attribution window
	}
	return a.cache.RecordTouchpoint(userID, campaignID, touchpointType, requestID, ttlDays)
}

// CompareModels returns attribution results for all models side-by-side, useful for analysis
func (a *AttributionService) CompareModels(userID, campaignID string) (map[string][]model.AttributionCredit, error) {
	models := []string{"last_click", "first_click", "linear", "time_decay", "position_based"}
	results := make(map[string][]model.AttributionCredit)

	for _, m := range models {
		credits, err := a.CalculateAttribution(userID, campaignID, m, 168)
		if err != nil {
			return nil, fmt.Errorf("failed for model %s: %w", m, err)
		}
		results[m] = credits
	}

	return results, nil
}

// Ensure time import is used
var _ = time.Now
