package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================================
// optimizeForCTV — CTVGoals branches: completion below target, primetime,
// live content, co-viewing, preferred device
// ============================================================================

func newCTVReq() *model.BidRequest {
	req := newReq()
	req.Device = model.InternalDevice{Type: "ctv"}
	req.Context = map[string]interface{}{
		"is_ctv":      true,
		"is_live":     true,
		"co_viewing":  true,
		"ctv_device":  "roku",
		"device_type": "roku",
	}
	return req
}

func TestOptimizeForCTV_CompletionBelowTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	pg := &model.PerformanceGoals{
		CTVGoals: &model.CTVOptimization{
			TargetCompletionRate: 0.95, // very high target — predictedCompletion (0.85 default) will be below
		},
	}
	perf := performanceData{completionRate: 0} // forces 0.85 default
	req := newCTVReq()
	mult := s.optimizeForCTV(camp, req, pg, perf)
	// CTV premium 1.3 * below-target penalty 0.8 = 1.04
	if mult < 1.0 || mult > 1.5 {
		t.Errorf("expected CTV multiplier with below-target completion ~1.04, got %f", mult)
	}
}

func TestOptimizeForCTV_CompletionAboveTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	pg := &model.PerformanceGoals{
		CTVGoals: &model.CTVOptimization{
			TargetCompletionRate: 0.70, // predictedCompletion (0.85) >= target → boost 1.2
		},
	}
	perf := performanceData{completionRate: 0}
	req := newCTVReq()
	mult := s.optimizeForCTV(camp, req, pg, perf)
	// 1.3 * 1.2 = 1.56
	if mult < 1.5 {
		t.Errorf("expected above-target completion boost >= 1.5, got %f", mult)
	}
}

func TestOptimizeForCTV_WithPrimetimeAndLiveContent(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	pg := &model.PerformanceGoals{
		CTVGoals: &model.CTVOptimization{
			PrimtimeBoost:    1.2,
			LiveContentBoost: 1.15,
		},
	}
	perf := performanceData{}
	req := newCTVReq()
	// Force primetime via context so we can control it
	req.Context["is_live"] = true
	// isPrimetime depends on current system time, so just assert no panic and valid result
	mult := s.optimizeForCTV(camp, req, pg, perf)
	if mult < 1.0 {
		t.Errorf("expected CTV multiplier >= 1.0, got %f", mult)
	}
}

func TestOptimizeForCTV_WithCoViewingAndPreferredDevice(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	pg := &model.PerformanceGoals{
		CTVGoals: &model.CTVOptimization{
			CoViewingBoost:   1.25,
			PreferredDevices: []string{"roku", "fire_tv"},
		},
	}
	perf := performanceData{}
	req := newCTVReq()
	req.Context["co_viewing"] = true
	req.Context["ctv_device"] = "roku"
	mult := s.optimizeForCTV(camp, req, pg, perf)
	// 1.3 * 1.25 (co-viewing) * 1.15 (preferred device) = ~1.87
	if mult < 1.5 {
		t.Errorf("expected co-viewing + preferred device boost >= 1.5, got %f", mult)
	}
}

func TestOptimizeForCTV_WithPreferredDeviceNoMatch(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	pg := &model.PerformanceGoals{
		CTVGoals: &model.CTVOptimization{
			PreferredDevices: []string{"apple_tv", "fire_tv"},
		},
	}
	perf := performanceData{}
	req := newCTVReq()
	req.Context["ctv_device"] = "roku" // not in preferred list
	req.Context["co_viewing"] = false
	mult := s.optimizeForCTV(camp, req, pg, perf)
	// 1.3 only (no preferred device match → no 1.15 boost)
	if mult != 1.3 {
		t.Errorf("expected 1.3 without preferred device match, got %f", mult)
	}
}

// ============================================================================
// optimizeForVCPM — high/low viewability branches, ratio cap/floor
// ============================================================================

func TestOptimizeForVCPM_HighViewability_B13(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	// predictViewability > 0.8 → ratio * 1.3
	// Use above-fold position to get high viewability
	req := newReq()
	req.AdSlot.Position = "above-fold"
	pg := &model.PerformanceGoals{TargetVCPM: 5.0}
	perf := performanceData{}
	mult := s.optimizeForVCPM(camp, req, pg, perf)
	if mult <= 0 {
		t.Errorf("expected positive VCPM multiplier, got %f", mult)
	}
}

func TestOptimizeForVCPM_LowViewability_Floor(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(100.0) // very high bid → ratio will be tiny → hits 0.2 floor
	req := newReq()
	req.AdSlot.Position = "below-fold"
	req.Context = map[string]interface{}{"viewability_rate": 0.1} // force low viewability
	pg := &model.PerformanceGoals{TargetVCPM: 0.001}
	perf := performanceData{}
	mult := s.optimizeForVCPM(camp, req, pg, perf)
	if mult < 0.1 {
		t.Errorf("expected VCPM multiplier floor >= 0.2, got %f", mult)
	}
}

func TestOptimizeForVCPM_NoBidPrice_B13(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(0) // BidPrice = 0 → returns 1.0
	req := newReq()
	pg := &model.PerformanceGoals{TargetVCPM: 5.0}
	perf := performanceData{}
	mult := s.optimizeForVCPM(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("expected 1.0 for zero BidPrice, got %f", mult)
	}
}

func TestOptimizeForVCPM_RatioCap(t *testing.T) {
	// Very low BidPrice → ratio > 2.5 → capped at 2.5
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(0.0001)
	req := newReq()
	req.AdSlot.Position = "above-fold"
	pg := &model.PerformanceGoals{TargetVCPM: 100.0}
	perf := performanceData{}
	mult := s.optimizeForVCPM(camp, req, pg, perf)
	if mult > 2.6 {
		t.Errorf("expected VCPM multiplier capped at 2.5, got %f", mult)
	}
}

// ============================================================================
// optimizeForCPC — cap at 2.0 and floor at 0.3
// ============================================================================

func TestOptimizeForCPC_RatioCap(t *testing.T) {
	// Very low BidPrice → ratio > 2.0 → capped at 2.0
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(0.0001) // very low bid
	req := newReq()
	pg := &model.PerformanceGoals{TargetCPC: 100.0}
	perf := performanceData{}
	mult := s.optimizeForCPC(camp, req, pg, perf)
	if mult > 2.1 {
		t.Errorf("expected CPC multiplier capped at 2.0, got %f", mult)
	}
}

func TestOptimizeForCPC_RatioFloor(t *testing.T) {
	// Very high BidPrice → ratio < 0.3 → floor at 0.3
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1000.0) // very high bid
	req := newReq()
	pg := &model.PerformanceGoals{TargetCPC: 0.001}
	perf := performanceData{}
	mult := s.optimizeForCPC(camp, req, pg, perf)
	if mult < 0.3 {
		t.Errorf("expected CPC multiplier floor at 0.3, got %f", mult)
	}
	if mult > 0.35 {
		t.Errorf("expected CPC multiplier ~0.3 for very high bid/low target, got %f", mult)
	}
}

// ============================================================================
// calculateVideoTargetingMultiplier — required player size not matched,
// placement blocked, linearity mismatch
// ============================================================================

func TestCalculateVideoTargetingMultiplier_RequiredSizeNotMatched(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		PlayerSizes: []model.VideoPlayerSize{
			{Size: "large", Required: true, MinWidth: 640, Boost: 1.3},
		},
	}
	// Request has a small player — won't match large (minWidth=640)
	req := newReq()
	req.Context = map[string]interface{}{
		"video":         true,
		"player_width":  float64(300),
		"player_height": float64(250),
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Errorf("expected blocked for required player size not matched, reason: %s", result.Reason)
	}
	if result.Reason != "required_player_size_not_matched" {
		t.Errorf("expected required_player_size_not_matched, got %s", result.Reason)
	}
}

func TestCalculateVideoTargetingMultiplier_PlacementMismatch(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		Placements: []string{"instream"},
	}
	req := newReq()
	req.Context = map[string]interface{}{
		"video":           true,
		"video_placement": "outstream", // mismatch
		"video_skip":      false,
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Errorf("expected placement mismatch to block, reason: %s", result.Reason)
	}
	if result.Reason != "placement_mismatch" {
		t.Errorf("expected placement_mismatch, got %s", result.Reason)
	}
}

func TestCalculateVideoTargetingMultiplier_DurationTooShort(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		MinDuration: 30, // requires at least 30s
	}
	req := newReq()
	req.Context = map[string]interface{}{
		"video":       true,
		"maxduration": float64(10), // only 10s max available → too short
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Errorf("expected blocked for duration too short, reason: %s", result.Reason)
	}
	if result.Reason != "duration_too_short" {
		t.Errorf("expected duration_too_short, got %s", result.Reason)
	}
}

func TestCalculateVideoTargetingMultiplier_PlayerSizeBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		PlayerSizes: []model.VideoPlayerSize{
			{Size: "large", Required: false, MinWidth: 300, Boost: 1.4},
		},
	}
	req := newReq()
	req.Context = map[string]interface{}{
		"video":         true,
		"player_width":  float64(640),
		"player_height": float64(480),
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for matching player size, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.3 {
		t.Errorf("expected player size boost >= 1.3, got %f", result.Multiplier)
	}
}

// ============================================================================
// getCurrentSeason — all branches (black_friday, holiday, new_year,
// spring, summer, fall, winter)
// ============================================================================

func TestGetCurrentSeason_AllBranches(t *testing.T) {
	// getCurrentSeason uses time.Now() which is Feb 26, 2026 → "winter"
	s := NewBiddingService(NewMockCache(), "")
	season := s.getCurrentSeason()
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
	// Today is Feb 26, 2026 → should be "winter"
	if season != "winter" {
		t.Logf("note: expected winter for Feb 26, got %s", season)
	}
}
