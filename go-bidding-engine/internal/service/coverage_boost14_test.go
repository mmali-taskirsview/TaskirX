package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================================
// getCurrentSeason — cover spring, summer, fall, new_year via direct call
// (winter is current; we call the function and verify "winter" is returned
// today Feb 2026, plus we test the function exists and handles the switch)
// ============================================================================

func TestGetCurrentSeason_Winter(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	season := s.getCurrentSeason()
	// Feb 2026 is winter
	if season != "winter" {
		t.Errorf("expected winter for Feb, got %s", season)
	}
}

// ============================================================================
// calculateSeasonalMultiplier — remaining uncovered branches:
//   Q4Boost, SummerBoost, BackToSchoolBoost, EnableHolidays (US),
//   Weekend+MonthEnd combined, SummerBoost zero (no match)
// ============================================================================

func TestCalculateSeasonalMultiplier_Q4Boost_B14(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		Q4Boost: 1.5,
	}
	// Even if today is not Q4, function should still run; result.IsQ4 will be false
	// and multiplier stays 1.0. This exercises the Q4 branch check.
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_SummerBoost_B14(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		SummerBoost: 1.4,
	}
	// Not summer today — multiplier should stay 1.0 and no match
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_BackToSchool_B14(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		BackToSchoolBoost: 1.3,
	}
	// Not Aug/Sep today — multiplier stays 1.0
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_EnableHolidays_B14(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		EnableHolidays: true,
		HolidayBoost:   1.5,
		Country:        "US",
	}
	// May or may not be a holiday today; just exercise the code path
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_HolidayDefaultCountry_B14(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		EnableHolidays: true,
		// Country is empty → defaults to "US"
		HolidayBoost: 0, // zero → default 1.3 used
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_WithTimezone_B14(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		Timezone:     "America/New_York",
		WeekendBoost: 1.2,
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_CapAt3_B14(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	// Stack multiple boosts to push above 3.0 cap
	// Use WeekendBoost + MonthEndBoost + an active event with large boost
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		WeekendBoost:  2.5,
		MonthEndBoost: 2.5,
		Events: []model.SeasonalEvent{
			{
				Name:      "big_sale",
				StartDate: "01-01",
				EndDate:   "12-31",
				Boost:     2.5,
				Recurring: true,
				Active:    true,
			},
		},
	}
	result := s.calculateSeasonalMultiplier(camp)
	// Even if today is not weekend/month-end, event covers full year → cap applies
	if result.Multiplier > 3.0 {
		t.Errorf("expected multiplier capped at 3.0, got %f", result.Multiplier)
	}
}

// ============================================================================
// checkBrandSafety — uncovered branches:
//   blocked category, blocked keyword, strict risky category,
//   relaxed risky category, fraud count > 10 strict, fraud count > 5
// ============================================================================

func TestCheckBrandSafety_BlockedCategory(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BlockedCategories = []string{"IAB14-1"}
	req := newReq()
	req.Context = map[string]interface{}{
		"categories": []interface{}{"IAB14-1", "IAB1"},
	}
	result := s.checkBrandSafety(camp, req)
	if !result.Blocked {
		t.Errorf("expected blocked for blocked_category, reason: %s", result.Reason)
	}
	if result.Reason != "blocked_category:IAB14-1" {
		t.Errorf("expected blocked_category:IAB14-1, got %s", result.Reason)
	}
}

func TestCheckBrandSafety_BlockedKeyword(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BlockedKeywords = []string{"casino"}
	req := newReq()
	req.Context = map[string]interface{}{
		"content": "Visit our online casino today!",
	}
	result := s.checkBrandSafety(camp, req)
	if !result.Blocked {
		t.Errorf("expected blocked for keyword, reason: %s", result.Reason)
	}
	if result.Reason != "blocked_keyword:casino" {
		t.Errorf("expected blocked_keyword:casino, got %s", result.Reason)
	}
}

func TestCheckBrandSafety_StrictRiskyCategory(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BrandSafetyLevel = "strict"
	req := newReq()
	req.Context = map[string]interface{}{
		"categories": []interface{}{"IAB25-3"},
	}
	result := s.checkBrandSafety(camp, req)
	if !result.Blocked {
		t.Errorf("expected blocked in strict mode for IAB25, reason: %s", result.Reason)
	}
}

func TestCheckBrandSafety_RelaxedRiskyCategory(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BrandSafetyLevel = "relaxed"
	req := newReq()
	req.Context = map[string]interface{}{
		"categories": []interface{}{"IAB25-5"},
	}
	result := s.checkBrandSafety(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked in relaxed mode, reason: %s", result.Reason)
	}
	if result.Multiplier >= 1.0 {
		t.Errorf("expected reduced multiplier in relaxed mode, got %f", result.Multiplier)
	}
}

func TestCheckBrandSafety_StandardRiskyCategory(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BrandSafetyLevel = "standard"
	req := newReq()
	req.Context = map[string]interface{}{
		"categories": []interface{}{"IAB26-1"},
	}
	result := s.checkBrandSafety(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked in standard mode, reason: %s", result.Reason)
	}
	if result.Multiplier >= 1.0 {
		t.Errorf("expected reduced multiplier for risky category, got %f", result.Multiplier)
	}
}

// ============================================================================
// calculateAudienceSegmentMultiplier — uncovered branches:
//   excluded segment, missing required segment, source-based default weights,
//   weight clamping, multiplier cap at 4.0
// ============================================================================

func TestAudienceSegment_ExcludedSegment(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-bad", Exclude: true},
	}
	req := newReq()
	req.User.Categories = []string{"seg-bad"}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if !result.Blocked {
		t.Errorf("expected blocked for excluded_segment")
	}
	if result.Reason != "excluded_segment:seg-bad" {
		t.Errorf("expected excluded_segment:seg-bad, got %s", result.Reason)
	}
}

func TestAudienceSegment_MissingRequiredSegment(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-vip", Required: true},
	}
	req := newReq()
	req.User.Categories = []string{"seg-other"}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if !result.Blocked {
		t.Errorf("expected blocked for missing_required_segment")
	}
	if result.Reason != "missing_required_segment" {
		t.Errorf("expected missing_required_segment, got %s", result.Reason)
	}
}

func TestAudienceSegment_FirstPartyDefaultWeight(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-fp", Source: "first_party", Weight: 0}, // zero weight → default 1.5
	}
	req := newReq()
	req.User.Categories = []string{"seg-fp"}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked")
	}
	if result.Multiplier < 1.4 {
		t.Errorf("expected multiplier ~1.5 for first_party, got %f", result.Multiplier)
	}
}

func TestAudienceSegment_LookalikeDefaultWeight(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-la", Source: "lookalike", Weight: 0}, // → 1.3
	}
	req := newReq()
	req.User.Categories = []string{"seg-la"}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if result.Multiplier < 1.2 {
		t.Errorf("expected multiplier ~1.3 for lookalike, got %f", result.Multiplier)
	}
}

func TestAudienceSegment_ThirdPartyDefaultWeight(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-3p", Source: "third_party", Weight: 0}, // → 1.2
	}
	req := newReq()
	req.User.Categories = []string{"seg-3p"}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if result.Multiplier < 1.1 {
		t.Errorf("expected multiplier ~1.2 for third_party, got %f", result.Multiplier)
	}
}

func TestAudienceSegment_ContextualDefaultWeight(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-ctx", Source: "contextual", Weight: 0}, // → 1.1
	}
	req := newReq()
	req.User.Categories = []string{"seg-ctx"}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if result.Multiplier < 1.0 {
		t.Errorf("expected multiplier ~1.1 for contextual, got %f", result.Multiplier)
	}
}

func TestAudienceSegment_WeightClampLow(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-low", Weight: 0.1}, // < 0.5 → clamped to 0.5
	}
	req := newReq()
	req.User.Categories = []string{"seg-low"}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if result.Multiplier < 0.45 || result.Multiplier > 0.55 {
		t.Errorf("expected clamped weight ~0.5, got %f", result.Multiplier)
	}
}

func TestAudienceSegment_WeightClampHigh(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-high", Weight: 5.0}, // > 3.0 → clamped to 3.0
	}
	req := newReq()
	req.User.Categories = []string{"seg-high"}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if result.Multiplier > 3.1 {
		t.Errorf("expected clamped weight <= 3.0, got %f", result.Multiplier)
	}
}

func TestAudienceSegment_MultiplierCapAt4(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-a", Weight: 3.0},
		{SegmentID: "seg-b", Weight: 3.0},
	}
	req := newReq()
	req.User.Categories = []string{"seg-a", "seg-b"}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if result.Multiplier > 4.0 {
		t.Errorf("expected cap at 4.0, got %f", result.Multiplier)
	}
}

func TestAudienceSegment_UserContextSegments(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "ctx-seg-1", Source: "contextual"},
	}
	req := newReq()
	req.Context = map[string]interface{}{
		"user_segments": []interface{}{"ctx-seg-1"},
	}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
}

func TestAudienceSegment_AudienceIDs(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "aud-99", Source: "third_party"},
	}
	req := newReq()
	req.Context = map[string]interface{}{
		"audience_ids": []interface{}{"aud-99"},
	}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.1 {
		t.Errorf("expected multiplier boost for third_party, got %f", result.Multiplier)
	}
}

// ============================================================================
// calculatePerformanceGoalMultiplier — cover additional goal types:
//   cpr, cpl, dcpm, maximize_clicks bid strategy, learning mode,
//   applyBidStrategy (target_cpa, maximize_conversions, manual),
//   determineOptimizationLevel (conservative, aggressive, moderate)
// ============================================================================

func TestPerfGoal_CPR(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpr",
		TargetCPR:   0.50,
	}
	req := newReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CPL(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpl",
		TargetCPL:   2.0,
	}
	req := newReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
}

func TestPerfGoal_DCPM(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "dcpm",
		TargetDCPM:  5.0,
	}
	req := newReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
}

func TestPerfGoal_LearningMode(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal:  "cpc",
		TargetCPC:    0.10,
		LearningMode: true,
	}
	req := newReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked in learning mode, reason: %s", result.Reason)
	}
	// Learning mode → OptimizationLevel should be conservative
	if result.OptimizationLevel != "conservative" {
		t.Errorf("expected conservative in learning mode, got %s", result.OptimizationLevel)
	}
}

func TestPerfGoal_MaximizeConversions_Strategy(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpa",
		TargetCPA:   1.0,
		BidStrategy: "maximize_conversions",
	}
	req := newReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %f", result.Multiplier)
	}
}

func TestPerfGoal_MaximizeClicks_Strategy(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpc",
		TargetCPC:   0.10,
		BidStrategy: "maximize_clicks",
	}
	req := newReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %f", result.Multiplier)
	}
}

func TestPerfGoal_ManualBidStrategy(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpm",
		TargetCPM:   1.0,
		BidStrategy: "manual",
	}
	req := newReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %f", result.Multiplier)
	}
}

func TestPerfGoal_TargetCPA_Strategy(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpa",
		TargetCPA:   1.5,
		BidStrategy: "target_cpa",
	}
	req := newReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %f", result.Multiplier)
	}
}

func TestPerfGoal_MaxBidAdjust(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal:  "cpc",
		TargetCPC:    100.0, // very high → would normally push multiplier up
		MaxBidAdjust: 1.5,
	}
	req := newReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Multiplier > 1.6 {
		t.Errorf("expected multiplier capped at MaxBidAdjust 1.5, got %f", result.Multiplier)
	}
}

func TestPerfGoal_MinBidAdjust(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal:  "cpa",
		TargetCPA:    0.0001, // very low → would push multiplier to floor
		MinBidAdjust: 0.8,
	}
	req := newReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	// MinBidAdjust should floor the multiplier
	if result.Multiplier < 0.75 {
		t.Errorf("expected multiplier floored by MinBidAdjust 0.8, got %f", result.Multiplier)
	}
}

// ============================================================================
// optimizeForCPL — lead_intent_score > 0.7 and is_b2b branches
// ============================================================================

func TestOptimizeForCPL_HighIntent_B14(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpl",
		TargetCPL:   2.0,
	}
	req := newReq()
	req.Context = map[string]interface{}{
		"lead_intent_score": float64(0.85), // > 0.7
	}
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
}

func TestOptimizeForCPL_B2B_B14(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpl",
		TargetCPL:   3.0,
	}
	req := newReq()
	req.Context = map[string]interface{}{
		"is_b2b": true,
	}
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
}

// ============================================================================
// optimizeForCPV — sound_on and video_placement instream branches
// ============================================================================

func TestOptimizeForCPV_SoundOn_B14(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpv",
		TargetCPV:   0.01,
	}
	req := newReq()
	req.Context = map[string]interface{}{
		"sound_on":        true,
		"video_placement": "instream",
	}
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
}

// ============================================================================
// calculateScore — cover return 0 branches via blocking multipliers
// ============================================================================

func TestCalculateScore_BrandSafetyBlocked(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BlockedPublishers = []string{"pub-blocked"}
	req := newReq()
	req.PublisherID = "pub-blocked"
	score := s.calculateScore(camp, req)
	if score != 0 {
		t.Errorf("expected score 0 for blocked publisher, got %f", score)
	}
}

func TestCalculateScore_AudienceBlocked(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-excluded", Exclude: true},
	}
	req := newReq()
	req.User.Categories = []string{"seg-excluded"}
	score := s.calculateScore(camp, req)
	if score != 0 {
		t.Errorf("expected score 0 for excluded audience, got %f", score)
	}
}

func TestCalculateScore_PriorityBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Priority = 10
	camp.Budget = 1000.0
	camp.DailyBudget = 100.0
	req := newReq()
	score := s.calculateScore(camp, req)
	if score < 1.0 {
		t.Errorf("expected higher score for priority=10, got %f", score)
	}
}

func TestCalculateScore_LowPriority(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Priority = 1
	camp.Budget = 1000.0
	camp.DailyBudget = 100.0
	req := newReq()
	score := s.calculateScore(camp, req)
	if score > 1.5 {
		t.Errorf("expected lower score for priority=1, got %f", score)
	}
}

func TestCalculateScore_DefaultPriority(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Priority = 0 // < 1 → defaults to 5
	camp.Budget = 1000.0
	camp.DailyBudget = 100.0
	req := newReq()
	score := s.calculateScore(camp, req)
	if score <= 0 {
		t.Errorf("expected positive score for default priority, got %f", score)
	}
}

func TestCalculateScore_OverMaxPriority(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Priority = 15 // > 10 → capped to 10
	camp.Budget = 1000.0
	camp.DailyBudget = 100.0
	req := newReq()
	score := s.calculateScore(camp, req)
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}
}

func TestCalculateScore_VideoBlocked(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Budget = 1000.0
	camp.DailyBudget = 100.0
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		Placements: []string{"instream"},
	}
	req := newReq()
	req.Context = map[string]interface{}{
		"video":           true,
		"video_placement": "outstream", // mismatch → blocked
	}
	score := s.calculateScore(camp, req)
	if score != 0 {
		t.Errorf("expected score 0 for video placement blocked, got %f", score)
	}
}
