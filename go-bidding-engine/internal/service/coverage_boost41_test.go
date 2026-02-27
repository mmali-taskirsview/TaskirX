package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// Coverage Boost 41: Target 90-94% functions to push them over 95%
// Functions targeted:
// 1. matchWeatherCondition (93.8%) - temperature-based condition branches
// 2. optimizeForEngagement (90.9%) - context nil path
// 3. categorizePlayerSize (90.9%) - negative size edge case, exact boundaries// ===============================================================================
// matchWeatherCondition Tests
// ===============================================================================

// TestB41_Weather_HotTemperature tests hot condition detected by temperature > 30
func TestB41_Weather_HotTemperature(t *testing.T) {
	svc := newTestBiddingService()

	weather := &WeatherData{
		Condition:   "partly cloudy", // Not explicitly "hot"
		Temperature: 35.0,            // > 30 triggers hot
	}

	result := svc.matchWeatherCondition(weather, "hot")
	if !result {
		t.Errorf("Expected match for hot temperature (35°C)")
	}
}

// TestB41_Weather_ColdTemperature tests cold condition detected by temperature < 5
func TestB41_Weather_ColdTemperature(t *testing.T) {
	svc := newTestBiddingService()

	weather := &WeatherData{
		Condition:   "clear", // Not explicitly "cold"
		Temperature: 2.0,     // < 5 triggers cold
	}

	result := svc.matchWeatherCondition(weather, "cold")
	if !result {
		t.Errorf("Expected match for cold temperature (2°C)")
	}
}

// TestB41_Weather_SynonymContains tests substring matching for synonyms
func TestB41_Weather_SynonymContains(t *testing.T) {
	svc := newTestBiddingService()

	// Test "precipitation" as substring in rainy synonym list
	weather := &WeatherData{
		Condition:   "heavy precipitation expected",
		Temperature: 20.0,
	}

	result := svc.matchWeatherCondition(weather, "rainy")
	if !result {
		t.Errorf("Expected match for rainy via precipitation synonym")
	}
}

// TestB41_Weather_TargetContainsCondition tests target substring in condition
func TestB41_Weather_TargetContainsCondition(t *testing.T) {
	svc := newTestBiddingService()

	// Test target keyword contained in longer condition string
	weather := &WeatherData{
		Condition:   "scattered showers and rainy periods",
		Temperature: 18.0,
	}

	result := svc.matchWeatherCondition(weather, "rainy")
	if !result {
		t.Errorf("Expected match when condition contains target keyword 'rainy'")
	}
}

// TestB41_Weather_NoMatch tests no match scenario
func TestB41_Weather_NoMatch(t *testing.T) {
	svc := newTestBiddingService()

	weather := &WeatherData{
		Condition:   "partly cloudy",
		Temperature: 20.0, // Not hot (>30) or cold (<5)
	}

	result := svc.matchWeatherCondition(weather, "snowy")
	if result {
		t.Errorf("Expected no match for snowy condition")
	}
}

// ===============================================================================
// optimizeForEngagement Tests
// ===============================================================================

// TestB41_Engagement_NeitherMobileNorInApp tests baseline with no boosts
func TestB41_Engagement_NeitherMobileNorInApp(t *testing.T) {
	mc := cache.NewMockCache()
	svc := &BiddingService{cache: mc}

	camp := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			RetargetingMode: "", // No retargeting
		},
	}

	req := &model.BidRequest{
		User:    model.InternalUser{ID: "user-123"},
		Device:  model.InternalDevice{Type: "desktop"}, // Not mobile
		Context: map[string]interface{}{
			// No environment set
		},
	}

	pg := &model.PerformanceGoals{EngagementGoal: 100}
	perf := performanceData{}

	mult := svc.optimizeForEngagement(camp, req, pg, perf)

	// Should be baseline 1.0 (no mobile, no in-app, no retargeting)
	if mult != 1.0 {
		t.Errorf("Expected multiplier 1.0 with no boosts, got %f", mult)
	}
}

// TestB41_Engagement_ContextNil tests nil context path
func TestB41_Engagement_ContextNil(t *testing.T) {
	mc := cache.NewMockCache()
	svc := &BiddingService{cache: mc}

	camp := &model.Campaign{
		ID: "camp-2",
	}

	req := &model.BidRequest{
		User:    model.InternalUser{ID: "user-456"},
		Device:  model.InternalDevice{Type: "mobile"}, // Mobile boost
		Context: nil,                                  // Nil context
	}

	pg := &model.PerformanceGoals{EngagementGoal: 50}
	perf := performanceData{}

	mult := svc.optimizeForEngagement(camp, req, pg, perf)

	// Should only have mobile boost (1.2)
	if mult != 1.2 {
		t.Errorf("Expected multiplier 1.2 (mobile only), got %f", mult)
	}
}

// ===============================================================================
// categorizePlayerSize Tests
// ===============================================================================

// TestB41_PlayerSize_NegativeValues tests negative width/height edge case
func TestB41_PlayerSize_NegativeValues(t *testing.T) {
	svc := newTestBiddingService()

	// Edge case: negative values should fall through to default
	size := svc.categorizePlayerSize(-100, -50)
	if size != "unknown" {
		t.Errorf("Expected 'unknown' for negative dimensions, got %s", size)
	}
}

// TestB41_PlayerSize_ExactBoundaries tests exact boundary values
func TestB41_PlayerSize_ExactBoundaries(t *testing.T) {
	svc := newTestBiddingService()

	// Test exact boundary at 1280
	size := svc.categorizePlayerSize(1280, 0)
	if size != "xlarge" {
		t.Errorf("Expected 'xlarge' for width=1280, got %s", size)
	}

	// Test exact boundary at 640
	size = svc.categorizePlayerSize(640, 0)
	if size != "large" {
		t.Errorf("Expected 'large' for width=640, got %s", size)
	}

	// Test exact boundary at 400
	size = svc.categorizePlayerSize(400, 0)
	if size != "medium" {
		t.Errorf("Expected 'medium' for width=400, got %s", size)
	}

	// Test just above 0
	size = svc.categorizePlayerSize(1, 0)
	if size != "small" {
		t.Errorf("Expected 'small' for width=1, got %s", size)
	}
}

// TestB41_PlayerSize_ZeroWidth_UseHeight tests using height when width is 0
func TestB41_PlayerSize_ZeroWidth_UseHeight(t *testing.T) {
	svc := newTestBiddingService()

	// Width is 0, should use height=1500 (xlarge)
	size := svc.categorizePlayerSize(0, 1500)
	if size != "xlarge" {
		t.Errorf("Expected 'xlarge' using height=1500, got %s", size)
	}

	// Width is 0, should use height=700 (large)
	size = svc.categorizePlayerSize(0, 700)
	if size != "large" {
		t.Errorf("Expected 'large' using height=700, got %s", size)
	}
}

// ===============================================================================
// Helper Functions
// ===============================================================================

func newTestBiddingService() *BiddingService {
	mc := cache.NewMockCache()
	return &BiddingService{cache: mc}
}
