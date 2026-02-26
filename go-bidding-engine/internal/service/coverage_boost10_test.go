package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================================
// calculateAutoBidMultiplier — all 5 branches
// ============================================================================

func TestCalculateAutoBidMultiplier_HighCTRLowWin(t *testing.T) {
	// CTR > 2.0%, WinRate < 30% → +20% boost
	mc := NewMockCache()
	mc.ctr["camp-auto"] = 0.025    // 2.5%
	mc.winRate["camp-auto"] = 0.20 // 20%
	s := NewBiddingService(mc, "")
	camp := newCampaign(1.0)
	camp.ID = "camp-auto"
	result := s.calculateAutoBidMultiplier(camp)
	if result != 1.20 {
		t.Errorf("expected 1.20, got %f", result)
	}
}

func TestCalculateAutoBidMultiplier_LowCTRHighWin(t *testing.T) {
	// CTR < 0.5%, WinRate > 70% → -20% reduction
	mc := NewMockCache()
	mc.ctr["camp-auto"] = 0.003    // 0.3%
	mc.winRate["camp-auto"] = 0.80 // 80%
	s := NewBiddingService(mc, "")
	camp := newCampaign(1.0)
	camp.ID = "camp-auto"
	result := s.calculateAutoBidMultiplier(camp)
	if result != 0.80 {
		t.Errorf("expected 0.80, got %f", result)
	}
}

func TestCalculateAutoBidMultiplier_ModerateCTRVeryLowWin(t *testing.T) {
	// CTR 1-3%, WinRate < 20% → +10% boost
	mc := NewMockCache()
	mc.ctr["camp-auto"] = 0.015    // 1.5%
	mc.winRate["camp-auto"] = 0.10 // 10%
	s := NewBiddingService(mc, "")
	camp := newCampaign(1.0)
	camp.ID = "camp-auto"
	result := s.calculateAutoBidMultiplier(camp)
	if result != 1.10 {
		t.Errorf("expected 1.10, got %f", result)
	}
}

func TestCalculateAutoBidMultiplier_LowCTRModerateWin(t *testing.T) {
	// CTR < 1.0%, WinRate 50-70% → -10% reduction
	mc := NewMockCache()
	mc.ctr["camp-auto"] = 0.008    // 0.8%
	mc.winRate["camp-auto"] = 0.60 // 60%
	s := NewBiddingService(mc, "")
	camp := newCampaign(1.0)
	camp.ID = "camp-auto"
	result := s.calculateAutoBidMultiplier(camp)
	if result != 0.90 {
		t.Errorf("expected 0.90, got %f", result)
	}
}

func TestCalculateAutoBidMultiplier_DefaultNeutral(t *testing.T) {
	// Mid-range values, none of the special branches → 1.0
	mc := NewMockCache()
	mc.ctr["camp-auto"] = 0.015    // 1.5% — not >2.0, not <0.5, in 1-3 range
	mc.winRate["camp-auto"] = 0.40 // 40% — not <20 (for case3), not 50-70 (for case4)
	s := NewBiddingService(mc, "")
	camp := newCampaign(1.0)
	camp.ID = "camp-auto"
	result := s.calculateAutoBidMultiplier(camp)
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

// ============================================================================
// calculateDealTargetingMultiplier — DealBidAdjustments, PublisherDeals,
// Exclusive block, no-best-deal fallback, cap/floor
// ============================================================================

func TestCalculateDealTargeting_DealBidAdjustment(t *testing.T) {
	// DealBidAdjustment for matched deal applies multiplier
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		DealBidAdjustments: []model.DealBidAdjust{
			{DealID: "deal-adj-1", BidMultiplier: 1.5},
		},
	}
	req := &model.BidRequest{
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal-adj-1", BidFloor: 0.5, At: 1},
			},
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if !result.Matched {
		t.Error("expected deal matched")
	}
	if result.Multiplier < 1.4 {
		t.Errorf("expected multiplier >= 1.4 from DealBidAdjust, got %f", result.Multiplier)
	}
}

func TestCalculateDealTargeting_DealBidAdjustmentZeroMultiplier(t *testing.T) {
	// DealBidAdjustment with BidMultiplier=0 should default to 1.0 (no change)
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		DealBidAdjustments: []model.DealBidAdjust{
			{DealID: "deal-adj-zero", BidMultiplier: 0},
		},
	}
	req := &model.BidRequest{
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal-adj-zero", BidFloor: 0.3, At: 1},
			},
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if !result.Matched {
		t.Error("expected deal matched")
	}
	// BidMultiplier=0 treated as 1.0, so multiplier stays 1.0
	if result.Multiplier < 0.9 {
		t.Errorf("expected multiplier ~1.0, got %f", result.Multiplier)
	}
}

func TestCalculateDealTargeting_PublisherDealsBidBoost(t *testing.T) {
	// PublisherDeals match → BidBoost applied
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		PublisherDeals: []model.PublisherDeal{
			{PublisherID: "pub-abc", BidBoost: 1.3, Exclusive: false},
		},
	}
	req := &model.BidRequest{
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal-pub-1", BidFloor: 0.2, At: 1},
			},
		},
		Context: map[string]interface{}{
			"publisher_id": "pub-abc",
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.2 {
		t.Errorf("expected multiplier >= 1.2 from publisher BidBoost, got %f", result.Multiplier)
	}
}

func TestCalculateDealTargeting_PublisherDealsExclusiveNotMatched(t *testing.T) {
	// PublisherDeals: Exclusive=true, matched deal not in publisher's DealIDs → blocked
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		PublisherDeals: []model.PublisherDeal{
			{
				PublisherID: "pub-excl",
				DealIDs:     []string{"deal-excl-only"},
				BidBoost:    1.1,
				Exclusive:   true,
			},
		},
	}
	req := &model.BidRequest{
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				// This deal is NOT in the publisher's exclusive DealIDs
				{ID: "deal-other", BidFloor: 0.3, At: 1},
			},
		},
		Context: map[string]interface{}{
			"publisher_id": "pub-excl",
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked: publisher exclusive deal not matched")
	}
	if result.Reason != "publisher_exclusive_deal_not_matched" {
		t.Errorf("expected reason publisher_exclusive_deal_not_matched, got %s", result.Reason)
	}
}

func TestCalculateDealTargeting_PublisherDealsExclusiveMatched(t *testing.T) {
	// PublisherDeals: Exclusive=true, matched deal IS in publisher's DealIDs → allowed
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		PublisherDeals: []model.PublisherDeal{
			{
				PublisherID: "pub-excl",
				DealIDs:     []string{"deal-allowed"},
				BidBoost:    1.1,
				Exclusive:   true,
			},
		},
	}
	req := &model.BidRequest{
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal-allowed", BidFloor: 0.3, At: 1},
			},
		},
		Context: map[string]interface{}{
			"publisher_id": "pub-excl",
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
}

func TestCalculateDealTargeting_NoBestDealRequiredFallback(t *testing.T) {
	// findBestDeal returns nil + RequireDeal=true + FallbackToOpen=true → open, not blocked
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	// DealTypes=["preferred"] but deal is "programmatic_guaranteed" (At=3) → no match
	camp.Targeting.DealTargeting = &model.DealTargeting{
		RequireDeal:    true,
		FallbackToOpen: true,
		DealTypes:      []string{"preferred"}, // only preferred accepted
	}
	req := &model.BidRequest{
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal-not-preferred", BidFloor: 0.3, At: 3}, // programmatic_guaranteed, not preferred
			},
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	// With FallbackToOpen, should not be blocked
	if result.Blocked {
		t.Errorf("expected not blocked with FallbackToOpen=true, reason: %s", result.Reason)
	}
}

func TestCalculateDealTargeting_NoBestDealRequiredNoFallback(t *testing.T) {
	// findBestDeal returns nil + RequireDeal=true + FallbackToOpen=false → blocked
	// Use DealTypes=["preferred"] but provide only private_auction deal (At=0, no WSeat)
	// so findBestDeal's type filter rejects it and returns nil
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		RequireDeal:    true,
		FallbackToOpen: false,
		DealTypes:      []string{"preferred"}, // only "preferred" type accepted
	}
	req := &model.BidRequest{
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				// At=3 → "programmatic_guaranteed", not "preferred" → rejected by type filter
				{ID: "deal-not-preferred", BidFloor: 0.3, At: 3},
			},
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked: no matching deal of preferred type and RequireDeal=true with no fallback")
	}
}

func TestCalculateDealTargeting_MultiplierCapAt2(t *testing.T) {
	// Multiplier > 2.0 should be capped at 2.0
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		PreferPG: true,
		DealBidAdjustments: []model.DealBidAdjust{
			{DealID: "deal-pg-high", BidMultiplier: 1.8},
		},
		PreferredDealIDs: []string{"deal-pg-high"},
		PublisherDeals: []model.PublisherDeal{
			{PublisherID: "pub-x", BidBoost: 1.5, Exclusive: false},
		},
	}
	// Deal with type=programmatic_guaranteed (At=1 with private marketplace)
	req := &model.BidRequest{
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal-pg-high", BidFloor: 0.1, At: 1},
			},
		},
		Context: map[string]interface{}{
			"publisher_id": "pub-x",
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier > 2.0 {
		t.Errorf("expected multiplier capped at 2.0, got %f", result.Multiplier)
	}
}

func TestCalculateDealTargeting_PubIDFromPubIDContext(t *testing.T) {
	// getPublisherID reads from "pub_id" context key
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		PublisherDeals: []model.PublisherDeal{
			{PublisherID: "pub-via-pub_id", BidBoost: 1.2, Exclusive: false},
		},
	}
	req := &model.BidRequest{
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal-x", BidFloor: 0.2, At: 1},
			},
		},
		Context: map[string]interface{}{
			"pub_id": "pub-via-pub_id",
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.1 {
		t.Errorf("expected boost from pub_id publisher, got %f", result.Multiplier)
	}
}

// ============================================================================
// calculateWeatherMultiplier — humidity, default boost, cap
// ============================================================================

func TestCalculateWeatherMultiplier_HumidityBelowMin(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	humMin := 40
	camp.Targeting.WeatherTargeting = &model.WeatherTargeting{
		HumidityMin: &humMin,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"humidity": float64(20)},
	}
	result := s.calculateWeatherMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for humidity below min")
	}
	if result.Reason != "humidity_below_min" {
		t.Errorf("expected reason humidity_below_min, got %s", result.Reason)
	}
}

func TestCalculateWeatherMultiplier_HumidityAboveMax(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	humMax := 60
	camp.Targeting.WeatherTargeting = &model.WeatherTargeting{
		HumidityMax: &humMax,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"humidity": float64(80)},
	}
	result := s.calculateWeatherMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for humidity above max")
	}
	if result.Reason != "humidity_above_max" {
		t.Errorf("expected reason humidity_above_max, got %s", result.Reason)
	}
}

func TestCalculateWeatherMultiplier_DefaultBoostUsed(t *testing.T) {
	// Condition matched with Boost=0 → falls back to DefaultBoost
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.WeatherTargeting = &model.WeatherTargeting{
		Conditions: []model.WeatherCondition{
			{Condition: "cloudy", Boost: 0}, // Boost=0 → use DefaultBoost
		},
		DefaultBoost: 1.4,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"weather": "cloudy"},
	}
	result := s.calculateWeatherMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.35 {
		t.Errorf("expected multiplier ~1.4 from DefaultBoost, got %f", result.Multiplier)
	}
}

func TestCalculateWeatherMultiplier_FallbackToHardcodedDefault(t *testing.T) {
	// Condition matched, Boost=0 and DefaultBoost=0 → hardcoded 1.3
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.WeatherTargeting = &model.WeatherTargeting{
		Conditions: []model.WeatherCondition{
			{Condition: "rainy", Boost: 0},
		},
		DefaultBoost: 0,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"weather_condition": "rainy"},
	}
	result := s.calculateWeatherMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.25 {
		t.Errorf("expected multiplier ~1.3 from hardcoded default, got %f", result.Multiplier)
	}
}

func TestCalculateWeatherMultiplier_CapAt2_5(t *testing.T) {
	// Multiple conditions all matching should cap at 2.5
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.WeatherTargeting = &model.WeatherTargeting{
		Conditions: []model.WeatherCondition{
			{Condition: "sunny", Boost: 1.8},
			{Condition: "hot", Boost: 1.8},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"weather":     "sunny",
			"temperature": float64(38), // hot = temp > 30
		},
	}
	result := s.calculateWeatherMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier > 2.5 {
		t.Errorf("expected multiplier capped at 2.5, got %f", result.Multiplier)
	}
}

func TestCalculateWeatherMultiplier_RequiredConditionMatched(t *testing.T) {
	// Required condition IS present → not blocked
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.WeatherTargeting = &model.WeatherTargeting{
		Conditions: []model.WeatherCondition{
			{Condition: "sunny", Required: true, Boost: 1.3},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"weather": "sunny"},
	}
	result := s.calculateWeatherMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked when required condition matched, reason: %s", result.Reason)
	}
}

// ============================================================================
// calculateDemographicMultiplier — missing branches
// ============================================================================

func TestCalculateDemographicMultiplier_ExcludeAgeRange(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		ExcludeAgeRanges: []model.AgeRange{
			{MinAge: 13, MaxAge: 17},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"age": float64(15)},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded age range")
	}
	if result.Reason != "age_excluded" {
		t.Errorf("expected reason age_excluded, got %s", result.Reason)
	}
}

func TestCalculateDemographicMultiplier_RequiredAgeNotMatched(t *testing.T) {
	// Age 50 doesn't match Required age range 25-35 → blocked
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		AgeRanges: []model.AgeRange{
			{MinAge: 25, MaxAge: 35, Required: true, Boost: 1.3},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"age": float64(50)},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for missing required age range")
	}
	if result.Reason != "missing_required_age_range" {
		t.Errorf("expected reason missing_required_age_range, got %s", result.Reason)
	}
}

func TestCalculateDemographicMultiplier_UnknownAgeDiscount(t *testing.T) {
	// Age=0 (unknown) + AgeRanges defined → UnknownAgeBoost applied
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		AgeRanges: []model.AgeRange{
			{MinAge: 25, MaxAge: 45, Boost: 1.2},
		},
		UnknownAgeBoost: 0.75,
	}
	req := &model.BidRequest{} // no age set → age=0
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked for unknown age")
	}
	if result.Multiplier > 0.9 {
		t.Errorf("expected multiplier < 0.9 (unknown age discount), got %f", result.Multiplier)
	}
}

func TestCalculateDemographicMultiplier_UnknownAgeDefaultDiscount(t *testing.T) {
	// UnknownAgeBoost=0 → default 0.8 applied
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		AgeRanges: []model.AgeRange{
			{MinAge: 20, MaxAge: 30, Boost: 1.1},
		},
		UnknownAgeBoost: 0, // will default to 0.8
	}
	req := &model.BidRequest{}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	// Default 0.8 discount → multiplier = 0.8
	if result.Multiplier > 0.85 || result.Multiplier < 0.75 {
		t.Errorf("expected multiplier ~0.8, got %f", result.Multiplier)
	}
}

func TestCalculateDemographicMultiplier_RequiredGenderNotMatched(t *testing.T) {
	// Gender "other" doesn't match Required "female" → blocked
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		Genders: []model.GenderRule{
			{Gender: "female", Required: true, Boost: 1.2},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"gender": "other"},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for missing required gender")
	}
	if result.Reason != "missing_required_gender" {
		t.Errorf("expected reason missing_required_gender, got %s", result.Reason)
	}
}

func TestCalculateDemographicMultiplier_UnknownGenderDiscount(t *testing.T) {
	// Gender="" (normalizes to "unknown") + Genders defined → UnknownGenderBoost
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		Genders: []model.GenderRule{
			{Gender: "male", Boost: 1.2},
		},
		UnknownGenderBoost: 0.7,
	}
	req := &model.BidRequest{} // no gender → unknown
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked for unknown gender")
	}
	if result.Multiplier > 0.8 {
		t.Errorf("expected multiplier <= 0.8 (unknown gender discount), got %f", result.Multiplier)
	}
}

func TestCalculateDemographicMultiplier_UnknownGenderDefaultDiscount(t *testing.T) {
	// UnknownGenderBoost=0 → default 0.8
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		Genders: []model.GenderRule{
			{Gender: "female", Boost: 1.1},
		},
		UnknownGenderBoost: 0,
	}
	req := &model.BidRequest{}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier > 0.85 || result.Multiplier < 0.75 {
		t.Errorf("expected multiplier ~0.8, got %f", result.Multiplier)
	}
}

func TestCalculateDemographicMultiplier_IncomeLevelMatch(t *testing.T) {
	// Income level "high" matches rule → boost applied
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		IncomeLevels: []model.IncomeRule{
			{Level: "high", Boost: 1.3},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"income_level": "high"},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.2 {
		t.Errorf("expected income boost >= 1.2, got %f", result.Multiplier)
	}
}

func TestCalculateDemographicMultiplier_RequiredIncomeMissing(t *testing.T) {
	// Income "medium" doesn't match Required "high" → blocked
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		IncomeLevels: []model.IncomeRule{
			{Level: "affluent", Required: true, Boost: 1.5},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"income": "medium"},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for missing required income level")
	}
	if result.Reason != "missing_required_income_level" {
		t.Errorf("expected reason missing_required_income_level, got %s", result.Reason)
	}
}

func TestCalculateDemographicMultiplier_IncomeLevelDefaultBoost(t *testing.T) {
	// IncomeRule.Boost=0 → default 1.2
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		IncomeLevels: []model.IncomeRule{
			{Level: "low", Boost: 0}, // default 1.2
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"income_level": "low"},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.15 {
		t.Errorf("expected multiplier ~1.2 from default boost, got %f", result.Multiplier)
	}
}

func TestCalculateDemographicMultiplier_CapAt3(t *testing.T) {
	// Multiple boosts exceed 3.0 → capped at 3.0
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		AgeRanges: []model.AgeRange{
			{MinAge: 20, MaxAge: 60, Boost: 2.0},
		},
		Genders: []model.GenderRule{
			{Gender: "male", Boost: 2.0},
		},
		IncomeLevels: []model.IncomeRule{
			{Level: "high", Boost: 1.5},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"age":          float64(30),
			"gender":       "male",
			"income_level": "high",
		},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier > 3.0 {
		t.Errorf("expected multiplier capped at 3.0, got %f", result.Multiplier)
	}
}

func TestCalculateDemographicMultiplier_FloorAt0_3(t *testing.T) {
	// Multiple discounts → floor at 0.3
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		AgeRanges: []model.AgeRange{
			{MinAge: 20, MaxAge: 60, Boost: 0.4},
		},
		Genders: []model.GenderRule{
			{Gender: "male", Boost: 0.4},
		},
		IncomeLevels: []model.IncomeRule{
			{Level: "low", Boost: 0.4},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"age":          float64(35),
			"gender":       "male",
			"income_level": "low",
		},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier < 0.3 {
		t.Errorf("expected multiplier floored at 0.3, got %f", result.Multiplier)
	}
}

func TestCalculateDemographicMultiplier_AgeFromUserStruct(t *testing.T) {
	// Age from req.User.Age (not context) — exercises extractDemographicInfo
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		AgeRanges: []model.AgeRange{
			{MinAge: 25, MaxAge: 50, Boost: 1.2},
		},
	}
	req := &model.BidRequest{
		User: model.InternalUser{Age: 30},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.1 {
		t.Errorf("expected age boost >= 1.1, got %f", result.Multiplier)
	}
}

func TestCalculateDemographicMultiplier_GenderFromUserStruct(t *testing.T) {
	// Gender from req.User.Gender (not context)
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		Genders: []model.GenderRule{
			{Gender: "female", Boost: 1.25},
		},
	}
	req := &model.BidRequest{
		User: model.InternalUser{Gender: "female"},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.2 {
		t.Errorf("expected gender boost >= 1.2, got %f", result.Multiplier)
	}
}

func TestCalculateDemographicMultiplier_GenderExcludeFromUserStruct(t *testing.T) {
	// ExcludeGenders using req.User.Gender — exercises full path
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		ExcludeGenders: []string{"male"},
	}
	req := &model.BidRequest{
		User: model.InternalUser{Gender: "m"}, // "m" normalizes to "male"
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded gender from user struct")
	}
}
