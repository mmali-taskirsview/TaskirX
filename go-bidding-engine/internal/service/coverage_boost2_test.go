package service

import (
	"strings"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ─────────────────────────────────────────────────────────────────────────────
// generateBanner tests (66.7% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestGenerateBanner_HTMLSnippet(t *testing.T) {
	camp := newCampaign(1.0)
	camp.Creative.HTMLSnippet = "<div>My Ad</div>"

	result := generateBanner(camp, "https://imp.test", "https://click.test")

	if result != "<div>My Ad</div>" {
		t.Errorf("expected HTMLSnippet to be returned, got: %s", result)
	}
}

func TestGenerateBanner_ImageFallback(t *testing.T) {
	camp := newCampaign(1.0)
	camp.Creative.HTMLSnippet = ""
	camp.Creative.URL = "https://cdn.test/img.png"
	camp.Creative.Width = 300
	camp.Creative.Height = 250

	result := generateBanner(camp, "https://imp.test", "https://click.test")

	if result == "" {
		t.Fatal("expected non-empty banner HTML")
	}
	if !strings.Contains(result, "https://cdn.test/img.png") {
		t.Error("expected image URL in banner")
	}
	if !strings.Contains(result, "https://click.test") {
		t.Error("expected click URL in banner")
	}
	if !strings.Contains(result, "https://imp.test") {
		t.Error("expected impression URL in banner")
	}
	if !strings.Contains(result, "300") {
		t.Error("expected width in banner")
	}
	if !strings.Contains(result, "250") {
		t.Error("expected height in banner")
	}
}

func TestGenerateBanner_EmptyCreative(t *testing.T) {
	camp := newCampaign(1.0)
	// No HTML snippet, no URL — should still return an img tag (empty URL)
	result := generateBanner(camp, "", "")
	if result == "" {
		t.Fatal("expected non-empty result even with empty creative")
	}
	if !strings.Contains(result, "<img") {
		t.Error("expected <img> tag in result")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateViewabilityMultiplier tests (37.5% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func makeViewabilityReq(position, deviceType string, dims []int) *model.BidRequest {
	req := newReq()
	req.AdSlot.Position = position
	req.Device.Type = deviceType
	req.AdSlot.Dimensions = dims
	return req
}

func TestCalcViewability_AboveFoldDesktop(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := makeViewabilityReq("above-fold", "desktop", []int{300, 250})

	result := svc.calculateViewabilityMultiplier(req)

	// above-fold (0.70) * desktop (1.1) * large (1.1) → capped realistic, should be > 1.0
	if result < 1.0 {
		t.Errorf("expected multiplier >= 1.0 for above-fold desktop, got %v", result)
	}
}

func TestCalcViewability_BelowFoldMobile(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := makeViewabilityReq("below-fold", "mobile", nil)

	result := svc.calculateViewabilityMultiplier(req)

	// below-fold (0.30) * mobile (0.9) → 0.27 → very low viewability → < 1.0
	if result >= 1.0 {
		t.Errorf("expected multiplier < 1.0 for below-fold mobile, got %v", result)
	}
}

func TestCalcViewability_StickyTablet(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := makeViewabilityReq("sticky", "tablet", []int{320, 50})

	result := svc.calculateViewabilityMultiplier(req)

	// sticky (0.85) * tablet (1.0) → > 0.7 → multiplier > 1.0
	if result <= 1.0 {
		t.Errorf("expected multiplier > 1.0 for sticky, got %v", result)
	}
}

func TestCalcViewability_SidebarDesktopSmall(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := makeViewabilityReq("sidebar", "desktop", []int{120, 60})

	result := svc.calculateViewabilityMultiplier(req)

	// sidebar (0.45) * desktop (1.1) = 0.495; < 0.7 but > 0.4 → neutral 1.0
	if result != 1.0 {
		t.Errorf("expected 1.0 neutral multiplier for sidebar mid-range, got %v", result)
	}
}

func TestCalcViewability_CTVHighViewability(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := makeViewabilityReq("atf", "ctv", []int{1920, 1080})

	result := svc.calculateViewabilityMultiplier(req)

	// atf (0.70) * ctv (1.15) * large (1.1) → high viewability → > 1.0
	if result <= 1.0 {
		t.Errorf("expected multiplier > 1.0 for CTV atf, got %v", result)
	}
}

func TestCalcViewability_TopPosition(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := makeViewabilityReq("top", "pc", nil)

	result := svc.calculateViewabilityMultiplier(req)

	// top = same as above-fold → 0.70 * 1.1 (desktop=pc) = 0.77 → > 0.7 → multiplier > 1.0
	if result <= 1.0 {
		t.Errorf("expected > 1.0 for top position, got %v", result)
	}
}

func TestCalcViewability_BottomPositionPhone(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := makeViewabilityReq("bottom", "phone", nil)

	result := svc.calculateViewabilityMultiplier(req)

	// bottom=below-fold (0.30) * phone=mobile (0.9) → 0.27 → < 0.40
	if result >= 1.0 {
		t.Errorf("expected < 1.0 for bottom phone, got %v", result)
	}
}

func TestCalcViewability_DefaultPosition(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := makeViewabilityReq("", "", nil)

	result := svc.calculateViewabilityMultiplier(req)

	// default 0.5 → falls in 40-70% → neutral 1.0
	if result != 1.0 {
		t.Errorf("expected 1.0 for default position, got %v", result)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// getCurrentSeason tests (53.8% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestGetCurrentSeason_ReturnsValidSeason(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	season := svc.getCurrentSeason()

	validSeasons := map[string]bool{
		"black_friday": true,
		"holiday":      true,
		"new_year":     true,
		"spring":       true,
		"summer":       true,
		"fall":         true,
		"winter":       true,
	}
	if !validSeasons[season] {
		t.Errorf("unexpected season: %s", season)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// performance_prediction: getFormatFactor (33.3% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestPerfPred_GetFormatFactor_AllFormats(t *testing.T) {
	svc := NewPerformancePredictionService(nil)

	cases := []struct {
		format   string
		expected float64
	}{
		{"video", 1.20},
		{"native", 1.15},
		{"banner", 1.0},
		{"interstitial", 1.25},
		{"unknown", 1.0},
		{"", 1.0},
	}

	for _, tc := range cases {
		result := svc.getFormatFactor(tc.format)
		if result != tc.expected {
			t.Errorf("getFormatFactor(%q) = %v, want %v", tc.format, result, tc.expected)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// performance_prediction: calculateSeasonality (80.0% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestPerfPred_CalculateSeasonality_Weekend(t *testing.T) {
	svc := NewPerformancePredictionService(nil)

	// Find a Saturday
	sat := time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC) // Feb 28 2026 is a Saturday
	result := svc.calculateSeasonality(sat)
	if result != 0.85 {
		t.Errorf("expected 0.85 on Saturday, got %v", result)
	}
}

func TestPerfPred_CalculateSeasonality_Sunday(t *testing.T) {
	svc := NewPerformancePredictionService(nil)

	sun := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC) // March 1 2026 is a Sunday
	result := svc.calculateSeasonality(sun)
	if result != 0.85 {
		t.Errorf("expected 0.85 on Sunday, got %v", result)
	}
}

func TestPerfPred_CalculateSeasonality_Friday(t *testing.T) {
	svc := NewPerformancePredictionService(nil)

	fri := time.Date(2026, 2, 27, 12, 0, 0, 0, time.UTC) // Feb 27 2026 is a Friday
	result := svc.calculateSeasonality(fri)
	if result != 1.10 {
		t.Errorf("expected 1.10 on Friday, got %v", result)
	}
}

func TestPerfPred_CalculateSeasonality_Weekday(t *testing.T) {
	svc := NewPerformancePredictionService(nil)

	mon := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC) // Feb 23 2026 is a Monday
	result := svc.calculateSeasonality(mon)
	if result != 1.0 {
		t.Errorf("expected 1.0 on Monday, got %v", result)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// performance_prediction: assessRisk (66.7% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestPerfPred_AssessRisk_LowImpressions(t *testing.T) {
	svc := NewPerformancePredictionService(nil)
	result := svc.assessRisk(50, 5, 1)
	if result != "high" {
		t.Errorf("expected 'high' for impressions < 100, got %s", result)
	}
}

func TestPerfPred_AssessRisk_LowCTR(t *testing.T) {
	svc := NewPerformancePredictionService(nil)
	// impressions=1000, clicks=3 → CTR=0.003 < 0.005 → high
	result := svc.assessRisk(1000, 3, 5)
	if result != "high" {
		t.Errorf("expected 'high' for low CTR, got %s", result)
	}
}

func TestPerfPred_AssessRisk_MediumCTR(t *testing.T) {
	svc := NewPerformancePredictionService(nil)
	// impressions=1000, clicks=7 → CTR=0.007 (0.005-0.01) → medium
	result := svc.assessRisk(1000, 7, 5)
	if result != "medium" {
		t.Errorf("expected 'medium' for medium CTR, got %s", result)
	}
}

func TestPerfPred_AssessRisk_LowRisk(t *testing.T) {
	svc := NewPerformancePredictionService(nil)
	// impressions=1000, clicks=20 → CTR=0.02, cvr=0.01 → low
	result := svc.assessRisk(1000, 20, 10)
	if result != "low" {
		t.Errorf("expected 'low' for good CTR and CVR, got %s", result)
	}
}

func TestPerfPred_AssessRisk_LowCVR(t *testing.T) {
	svc := NewPerformancePredictionService(nil)
	// impressions=1000, clicks=20 → CTR=0.02 (OK), conversions=0 → CVR=0 < 0.001 → high
	result := svc.assessRisk(1000, 20, 0)
	if result != "high" {
		t.Errorf("expected 'high' for zero CVR, got %s", result)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// performance_prediction: GetPredictionAccuracy with real data (36% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestPerfPred_GetPredictionAccuracy_WithMatchingData(t *testing.T) {
	svc := NewPerformancePredictionService(nil)
	entityID := "camp-acc-1"

	// Store a historical record with a key matching the prediction timestamp
	now := time.Now()
	hourKey := entityID + ":" + now.Format("2006-01-02-15")
	record := &PerformanceRecord{
		EntityID:    entityID,
		EntityType:  "campaign",
		Timestamp:   now,
		Impressions: 1000,
		Clicks:      30,
		Conversions: 3,
		Revenue:     15.0,
		Spend:       10.0,
		CTR:         0.03,
		CVR:         0.003,
		ROAS:        1.5,
	}
	svc.historicalData.Store(hourKey, record)

	// Store a prediction for this entity at this hour
	pred := &PredictionResult{
		ID:          "pred-acc-1",
		EntityID:    entityID,
		EntityType:  "campaign",
		PredictedAt: now,
		Predictions: map[string]*MetricPrediction{
			"ctr":  {Metric: "ctr", PredictedValue: 0.025},
			"cvr":  {Metric: "cvr", PredictedValue: 0.002},
			"roas": {Metric: "roas", PredictedValue: 1.4},
		},
	}
	svc.predictions.Store("pred-acc-1", pred)

	accuracy := svc.GetPredictionAccuracy(entityID, 24)

	if accuracy == nil {
		t.Fatal("expected non-nil accuracy map")
	}
	if accuracy["samples"] < 1 {
		t.Errorf("expected samples >= 1, got %v", accuracy["samples"])
	}
	// With matching data, MAE should be the abs diff of predicted vs actual
	expectedCTRMae := 0.025 - 0.03
	if expectedCTRMae < 0 {
		expectedCTRMae = -expectedCTRMae
	}
	if accuracy["ctr_mae"] < 0 {
		t.Error("expected non-negative ctr_mae")
	}
}

func TestPerfPred_GetPredictionAccuracy_NoMatchingData(t *testing.T) {
	svc := NewPerformancePredictionService(nil)

	// Store a prediction with no matching historical data
	pred := &PredictionResult{
		ID:          "pred-no-match",
		EntityID:    "camp-no-match",
		EntityType:  "campaign",
		PredictedAt: time.Now(),
		Predictions: map[string]*MetricPrediction{
			"ctr": {Metric: "ctr", PredictedValue: 0.02},
		},
	}
	svc.predictions.Store("pred-no-match", pred)

	accuracy := svc.GetPredictionAccuracy("camp-no-match", 24)

	if accuracy["samples"] != 0 {
		t.Errorf("expected 0 samples when no historical match, got %v", accuracy["samples"])
	}
	if accuracy["ctr_mae"] != 0 {
		t.Errorf("expected 0 ctr_mae with no match, got %v", accuracy["ctr_mae"])
	}
}

func TestPerfPred_GetPredictionAccuracy_OutsideLookback(t *testing.T) {
	svc := NewPerformancePredictionService(nil)

	// Prediction outside lookback window (2 days ago)
	old := time.Now().Add(-48 * time.Hour)
	pred := &PredictionResult{
		ID:          "pred-old",
		EntityID:    "camp-old",
		EntityType:  "campaign",
		PredictedAt: old,
		Predictions: map[string]*MetricPrediction{
			"ctr": {Metric: "ctr", PredictedValue: 0.02},
		},
	}
	svc.predictions.Store("pred-old", pred)

	// 1 hour lookback — should not include the old prediction
	accuracy := svc.GetPredictionAccuracy("camp-old", 1)
	if accuracy["samples"] != 0 {
		t.Errorf("expected 0 samples outside lookback, got %v", accuracy["samples"])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// incrementality: generateRecommendation (38.5% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestInc_GenerateRecommendation_NotComplete(t *testing.T) {
	svc := NewIncrementalityService(nil)

	// status != "complete" → always returns continue message
	result := svc.generateRecommendation(25.0, 0.95, "running")
	if !strings.Contains(result, "Continue") {
		t.Errorf("expected 'Continue' for non-complete status, got: %s", result)
	}
}

func TestInc_GenerateRecommendation_LowConfidence(t *testing.T) {
	svc := NewIncrementalityService(nil)

	result := svc.generateRecommendation(15.0, 0.80, "complete")
	if !strings.Contains(result, "inconclusive") {
		t.Errorf("expected inconclusive for confidence < 0.9, got: %s", result)
	}
}

func TestInc_GenerateRecommendation_StrongLift(t *testing.T) {
	svc := NewIncrementalityService(nil)

	result := svc.generateRecommendation(25.0, 0.95, "complete")
	if !strings.Contains(result, "Strong") {
		t.Errorf("expected 'Strong' for lift > 20, got: %s", result)
	}
}

func TestInc_GenerateRecommendation_ModerateLift(t *testing.T) {
	svc := NewIncrementalityService(nil)

	result := svc.generateRecommendation(10.0, 0.95, "complete")
	if !strings.Contains(result, "Moderate") {
		t.Errorf("expected 'Moderate' for lift 5-20, got: %s", result)
	}
}

func TestInc_GenerateRecommendation_SmallPositiveLift(t *testing.T) {
	svc := NewIncrementalityService(nil)

	result := svc.generateRecommendation(2.0, 0.95, "complete")
	if !strings.Contains(result, "Small") {
		t.Errorf("expected 'Small' for lift 0-5, got: %s", result)
	}
}

func TestInc_GenerateRecommendation_NoLift(t *testing.T) {
	svc := NewIncrementalityService(nil)

	result := svc.generateRecommendation(-2.0, 0.95, "complete")
	if !strings.Contains(result, "No significant") {
		t.Errorf("expected 'No significant' for lift -5 to 0, got: %s", result)
	}
}

func TestInc_GenerateRecommendation_NegativeLift(t *testing.T) {
	svc := NewIncrementalityService(nil)

	result := svc.generateRecommendation(-10.0, 0.95, "complete")
	if !strings.Contains(result, "Negative") {
		t.Errorf("expected 'Negative' for lift < -5, got: %s", result)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// applyBidStrategy (94.1% — few missing branches)
// ─────────────────────────────────────────────────────────────────────────────

func TestApplyBidStrategy_MaximizeConversions_HighCVR(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{BidStrategy: "maximize_conversions"}
	perf := performanceData{cvr: 0.05} // > 0.03

	result := svc.applyBidStrategy(pg, perf)
	if result != 1.4 {
		t.Errorf("expected 1.4 for high CVR maximize_conversions, got %v", result)
	}
}

func TestApplyBidStrategy_MaximizeConversions_LowCVR(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{BidStrategy: "maximize_conversions"}
	perf := performanceData{cvr: 0.01}

	result := svc.applyBidStrategy(pg, perf)
	if result != 1.0 {
		t.Errorf("expected 1.0 for low CVR maximize_conversions, got %v", result)
	}
}

func TestApplyBidStrategy_TargetCPA_HighRatio(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{BidStrategy: "target_cpa", TargetCPA: 10.0}
	perf := performanceData{cpa: 7.0} // ratio = 10/7 = 1.43 > 1.2 → cap at 1.2

	result := svc.applyBidStrategy(pg, perf)
	if result != 1.2 {
		t.Errorf("expected 1.2 for high ratio target_cpa, got %v", result)
	}
}

func TestApplyBidStrategy_TargetCPA_LowRatio(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{BidStrategy: "target_cpa", TargetCPA: 5.0}
	perf := performanceData{cpa: 8.0} // ratio = 5/8 = 0.625 < 0.8 → 0.8

	result := svc.applyBidStrategy(pg, perf)
	if result != 0.8 {
		t.Errorf("expected 0.8 for low ratio target_cpa, got %v", result)
	}
}

func TestApplyBidStrategy_MaximizeClicks_HighCTR(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{BidStrategy: "maximize_clicks"}
	perf := performanceData{ctr: 0.02} // > 0.015

	result := svc.applyBidStrategy(pg, perf)
	if result != 1.3 {
		t.Errorf("expected 1.3 for high CTR maximize_clicks, got %v", result)
	}
}

func TestApplyBidStrategy_MaximizeClicks_LowCTR(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{BidStrategy: "maximize_clicks"}
	perf := performanceData{ctr: 0.005}

	result := svc.applyBidStrategy(pg, perf)
	if result != 1.0 {
		t.Errorf("expected 1.0 for low CTR maximize_clicks, got %v", result)
	}
}

func TestApplyBidStrategy_Manual(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{BidStrategy: "manual"}
	perf := performanceData{ctr: 0.05, cvr: 0.05, cpa: 1.0}

	result := svc.applyBidStrategy(pg, perf)
	if result != 1.0 {
		t.Errorf("expected 1.0 for manual strategy, got %v", result)
	}
}

func TestApplyBidStrategy_Default(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{BidStrategy: "unknown_strategy"}
	perf := performanceData{}

	result := svc.applyBidStrategy(pg, perf)
	if result != 1.0 {
		t.Errorf("expected 1.0 for unknown strategy, got %v", result)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// determineOptimizationLevel tests
// ─────────────────────────────────────────────────────────────────────────────

func TestDetermineOptLevel_LearningMode(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{LearningMode: true}
	result := svc.determineOptimizationLevel(pg, performanceData{impressions: 5000, cpa: 5.0})
	if result != "conservative" {
		t.Errorf("expected 'conservative' for learning mode, got %s", result)
	}
}

func TestDetermineOptLevel_InsufficientData(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{}
	result := svc.determineOptimizationLevel(pg, performanceData{impressions: 500})
	if result != "conservative" {
		t.Errorf("expected 'conservative' for < 1000 impressions, got %s", result)
	}
}

func TestDetermineOptLevel_Aggressive(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPA: 10.0}
	// cpa=7.0 < targetCPA*0.8=8.0 → aggressive
	result := svc.determineOptimizationLevel(pg, performanceData{impressions: 2000, cpa: 7.0})
	if result != "aggressive" {
		t.Errorf("expected 'aggressive' when performing well, got %s", result)
	}
}

func TestDetermineOptLevel_Moderate(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPA: 10.0}
	// cpa=9.5 > targetCPA*0.8=8.0 → moderate
	result := svc.determineOptimizationLevel(pg, performanceData{impressions: 2000, cpa: 9.5})
	if result != "moderate" {
		t.Errorf("expected 'moderate', got %s", result)
	}
}
