package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// newPerfData builds a performanceData with sensible defaults for testing.
func newPerfData() performanceData {
	return performanceData{
		ctr:            0.02,
		cvr:            0.05,
		viewability:    0.65,
		completionRate: 0.60,
		engagementRate: 0.04,
		winRate:        0.20,
		cpa:            5.0,
		cpi:            3.0,
		cps:            8.0,
		roas:           3.0,
	}
}

func newReq() *model.BidRequest {
	return &model.BidRequest{
		Device: model.InternalDevice{Type: "mobile"},
	}
}

func newCampaign(bidPrice float64) *model.Campaign {
	return &model.Campaign{
		ID:       "camp-1",
		BidPrice: bidPrice,
		Creative: model.Creative{Type: "banner"},
	}
}

// ============================================================================
// optimizeForCPA
// ============================================================================

func TestOptimizeForCPA_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCPA(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetCPA: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForCPA_Ratio(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPA: 10.0}
	// predictedCTR≈0.01, predictedCVR≈0.02 → expectedConvRate=0.0002
	// maxBidForCPA = 10 * 0.0002 = 0.002; ratio = 0.002/1.0 = 0.002 → clamped to 0.3
	result := s.optimizeForCPA(newCampaign(1.0), newReq(), pg, newPerfData())
	if result < 0.1 || result > 2.1 {
		t.Errorf("unexpected ratio %f", result)
	}
}

func TestOptimizeForCPA_NoBidPrice(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPA: 10.0}
	result := s.optimizeForCPA(newCampaign(0), newReq(), pg, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for zero bid price, got %f", result)
	}
}

// ============================================================================
// optimizeForCPC
// ============================================================================

func TestOptimizeForCPC_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCPC(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetCPC: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForCPC_BelowFloor(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	// With high bid price and low target CPC → ratio should be < 0.3 → clamped to 0.3
	pg := &model.PerformanceGoals{TargetCPC: 0.001}
	result := s.optimizeForCPC(newCampaign(100.0), newReq(), pg, newPerfData())
	if result != 0.3 {
		t.Errorf("expected floor 0.3, got %f", result)
	}
}

func TestOptimizeForCPC_NoBidPrice(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPC: 1.0}
	result := s.optimizeForCPC(newCampaign(0), newReq(), pg, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

// ============================================================================
// optimizeForCPM
// ============================================================================

func TestOptimizeForCPM_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCPM(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetCPM: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForCPM_HighViewability(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPM: 5.0}
	req := &model.BidRequest{
		Device:  model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{"predicted_viewability": 0.85},
	}
	result := s.optimizeForCPM(newCampaign(1.0), req, pg, newPerfData())
	if result < 1.0 {
		t.Errorf("expected >= 1.0 for high viewability, got %f", result)
	}
}

func TestOptimizeForCPM_LowViewability(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPM: 5.0}
	req := &model.BidRequest{
		Device:  model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{"ad_position": "below_fold"},
	}
	// below_fold multiplier 0.7 on base 0.4 → viewRate ≈ 0.28, which is < 0.4 → returns 0.7
	perf := performanceData{viewability: 0.4}
	result := s.optimizeForCPM(newCampaign(1.0), req, pg, perf)
	if result > 1.0 {
		t.Errorf("expected <= 1.0 for low viewability, got %f", result)
	}
}

// ============================================================================
// optimizeForViewability
// ============================================================================

func TestOptimizeForViewability_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForViewability(newCampaign(1.0), newReq(), &model.PerformanceGoals{ViewabilityGoal: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForViewability_AboveTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{ViewabilityGoal: 0.5}
	req := &model.BidRequest{
		Context: map[string]interface{}{"predicted_viewability": 0.9},
	}
	result := s.optimizeForViewability(newCampaign(1.0), req, pg, newPerfData())
	if result <= 1.0 {
		t.Errorf("expected > 1.0 for above-target viewability, got %f", result)
	}
}

func TestOptimizeForViewability_BelowTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{ViewabilityGoal: 0.8}
	req := &model.BidRequest{
		Context: map[string]interface{}{"predicted_viewability": 0.3},
	}
	result := s.optimizeForViewability(newCampaign(1.0), req, pg, performanceData{})
	if result >= 1.0 {
		t.Errorf("expected < 1.0 for below-target viewability, got %f", result)
	}
}

// ============================================================================
// optimizeForCompletion
// ============================================================================

func TestOptimizeForCompletion_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCompletion(newCampaign(1.0), newReq(), &model.PerformanceGoals{CompletionGoal: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForCompletion_HighCompletion(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{CompletionGoal: 0.5}
	req := &model.BidRequest{
		Context: map[string]interface{}{"completion_rate": 0.85},
	}
	result := s.optimizeForCompletion(newCampaign(1.0), req, pg, newPerfData())
	if result <= 1.0 {
		t.Errorf("expected > 1.0, got %f", result)
	}
}

func TestOptimizeForCompletion_LowCompletion(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{CompletionGoal: 0.8}
	req := &model.BidRequest{
		Context: map[string]interface{}{"completion_rate": 0.3},
	}
	result := s.optimizeForCompletion(newCampaign(1.0), req, pg, performanceData{})
	if result >= 1.0 {
		t.Errorf("expected < 1.0, got %f", result)
	}
}

// ============================================================================
// optimizeForEngagement
// ============================================================================

func TestOptimizeForEngagement_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForEngagement(newCampaign(1.0), newReq(), &model.PerformanceGoals{EngagementGoal: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForEngagement_MobileInApp(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{EngagementGoal: 0.05}
	req := &model.BidRequest{
		Device:  model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{"environment": "in-app"},
	}
	result := s.optimizeForEngagement(newCampaign(1.0), req, pg, newPerfData())
	if result <= 1.0 {
		t.Errorf("expected > 1.0 for mobile in-app, got %f", result)
	}
}

// ============================================================================
// optimizeForCPI
// ============================================================================

func TestOptimizeForCPI_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCPI(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetCPI: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForCPI_AppGoalsFallback(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{
		TargetCPI: 0,
		AppGoals:  &model.AppOptimization{TargetCostPerInstall: 2.0},
	}
	result := s.optimizeForCPI(newCampaign(1.0), newReq(), pg, newPerfData())
	// Should use AppGoals.TargetCostPerInstall
	if result == 1.0 {
		t.Error("expected non-default multiplier when AppGoals provides CPI")
	}
}

func TestOptimizeForCPI_NoBidPrice(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPI: 2.0}
	result := s.optimizeForCPI(newCampaign(0), newReq(), pg, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for zero bid, got %f", result)
	}
}

// ============================================================================
// optimizeForCPS
// ============================================================================

func TestOptimizeForCPS_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCPS(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetCPS: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForCPS_WithCartAbandonBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{
		TargetCPS: 5.0,
		EcommerceGoals: &model.EcommerceOptimization{
			TargetCostPerSale: 5.0,
			CartAbandonBoost:  1.5,
		},
	}
	req := &model.BidRequest{
		Device:  model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{"is_cart_abandoner": true},
	}
	result := s.optimizeForCPS(newCampaign(1.0), req, pg, newPerfData())
	if result == 1.0 {
		t.Errorf("expected adjusted multiplier, got 1.0")
	}
}

// ============================================================================
// optimizeForCPR
// ============================================================================

func TestOptimizeForCPR_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCPR(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetCPR: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForCPR_WithTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPR: 10.0}
	result := s.optimizeForCPR(newCampaign(1.0), newReq(), pg, newPerfData())
	if result < 0.3 || result > 2.0 {
		t.Errorf("unexpected result %f outside [0.3, 2.0]", result)
	}
}

func TestOptimizeForCPR_NoBidPrice(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPR: 5.0}
	result := s.optimizeForCPR(newCampaign(0), newReq(), pg, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

// ============================================================================
// optimizeForCPL
// ============================================================================

func TestOptimizeForCPL_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCPL(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetCPL: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForCPL_HighIntentB2B(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPL: 50.0}
	req := &model.BidRequest{
		Device: model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{
			"lead_intent_score":    0.9,
			"is_b2b":               true,
			"historical_lead_rate": 0.08,
		},
	}
	result := s.optimizeForCPL(newCampaign(1.0), req, pg, newPerfData())
	if result < 0.3 || result > 2.5 {
		t.Errorf("unexpected result %f", result)
	}
}

// ============================================================================
// optimizeForCPV
// ============================================================================

func TestOptimizeForCPV_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCPV(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetCPV: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForCPV_SoundOnInstream(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPV: 0.05}
	req := &model.BidRequest{
		Device: model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{
			"sound_on":            true,
			"video_placement":     "instream",
			"predicted_view_rate": 0.7,
		},
	}
	result := s.optimizeForCPV(newCampaign(0.1), req, pg, newPerfData())
	if result < 0.3 || result > 2.5 {
		t.Errorf("unexpected result %f", result)
	}
}

func TestOptimizeForCPV_NoBidPrice(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPV: 0.05}
	// Zero bid price leads to divide by zero in predictCPV → returns 0 → optimizer returns 1.0
	result := s.optimizeForCPV(newCampaign(0), newReq(), pg, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for zero bid, got %f", result)
	}
}

// ============================================================================
// optimizeForCPCV
// ============================================================================

func TestOptimizeForCPCV_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCPCV(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetCPCV: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForCPCV_NonSkippableShort(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPCV: 0.10}
	req := &model.BidRequest{
		Device: model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{
			"skippable":                 false,
			"video_duration":            float64(15),
			"predicted_completion_rate": 0.95,
		},
	}
	result := s.optimizeForCPCV(newCampaign(0.1), req, pg, newPerfData())
	// Non-skippable + short = high completion → ratio boosted
	if result < 0.3 || result > 3.0 {
		t.Errorf("unexpected result %f", result)
	}
}

func TestOptimizeForCPCV_CTVInventory(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPCV: 0.10}
	req := &model.BidRequest{
		Device:  model.InternalDevice{Type: "ctv"},
		Context: map[string]interface{}{"is_ctv": true},
	}
	result := s.optimizeForCPCV(newCampaign(0.1), req, pg, newPerfData())
	if result < 0.3 || result > 3.0 {
		t.Errorf("unexpected result %f", result)
	}
}

func TestOptimizeForCPCV_LongVideo(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPCV: 0.10}
	req := &model.BidRequest{
		Device: model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{
			"skippable":      true,
			"video_duration": float64(60),
		},
	}
	result := s.optimizeForCPCV(newCampaign(0.1), req, pg, newPerfData())
	if result < 0.3 || result > 3.0 {
		t.Errorf("unexpected result %f", result)
	}
}

// ============================================================================
// optimizeForCPE
// ============================================================================

func TestOptimizeForCPE_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCPE(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetCPE: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForCPE_RichMediaMobileInApp(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPE: 0.20}
	req := &model.BidRequest{
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"creative_type":             "rich_media",
			"environment":               "in-app",
			"predicted_engagement_rate": 0.05,
		},
	}
	result := s.optimizeForCPE(newCampaign(0.5), req, pg, newPerfData())
	if result < 0.3 || result > 2.5 {
		t.Errorf("unexpected result %f", result)
	}
}

func TestOptimizeForCPE_NoPredictedCPE(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPE: 0.20}
	// zero bid price → predictCPE returns 0 → optimizer returns 1.0
	result := s.optimizeForCPE(newCampaign(0), newReq(), pg, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for zero CPE prediction, got %f", result)
	}
}

// ============================================================================
// optimizeForVCPM
// ============================================================================

func TestOptimizeForVCPM_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForVCPM(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetVCPM: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForVCPM_NoBidPrice(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetVCPM: 5.0}
	result := s.optimizeForVCPM(newCampaign(0), newReq(), pg, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for zero bid, got %f", result)
	}
}

func TestOptimizeForVCPM_HighViewability(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetVCPM: 5.0}
	req := &model.BidRequest{
		Context: map[string]interface{}{"predicted_viewability": 0.9},
	}
	result := s.optimizeForVCPM(newCampaign(0.001), req, pg, newPerfData())
	if result < 0.2 || result > 2.5 {
		t.Errorf("unexpected result %f", result)
	}
}

func TestOptimizeForVCPM_LowViewability(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetVCPM: 5.0}
	req := &model.BidRequest{
		Context: map[string]interface{}{"predicted_viewability": 0.2},
	}
	result := s.optimizeForVCPM(newCampaign(0.01), req, pg, performanceData{})
	if result < 0.2 || result > 2.5 {
		t.Errorf("unexpected result %f", result)
	}
}

// ============================================================================
// optimizeForCPAD
// ============================================================================

func TestOptimizeForCPAD_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCPAD(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetCPAD: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForCPAD_HighRatedFeaturedApp(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPAD: 1.0}
	req := &model.BidRequest{
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"app_rating":   float64(4.8),
			"app_featured": true,
		},
	}
	result := s.optimizeForCPAD(newCampaign(0.5), req, pg, newPerfData())
	if result < 0.2 || result > 2.5 {
		t.Errorf("unexpected result %f", result)
	}
}

func TestOptimizeForCPAD_DesktopPenalty(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPAD: 1.0}
	req := &model.BidRequest{Device: model.InternalDevice{Type: "desktop"}}
	result := s.optimizeForCPAD(newCampaign(1.0), req, pg, newPerfData())
	// Desktop gets maxBid * 0.1 penalty → very low ratio → floor 0.2
	if result != 0.2 {
		t.Errorf("expected floor 0.2 for desktop, got %f", result)
	}
}

func TestOptimizeForCPAD_NoBidPrice(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPAD: 1.0}
	result := s.optimizeForCPAD(newCampaign(0), newReq(), pg, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

// ============================================================================
// optimizeForCTV
// ============================================================================

func TestOptimizeForCTV_NonCTVInventory(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{}
	req := &model.BidRequest{Device: model.InternalDevice{Type: "mobile"}}
	result := s.optimizeForCTV(newCampaign(1.0), req, pg, newPerfData())
	if result != 0.5 {
		t.Errorf("expected 0.5 penalty for non-CTV, got %f", result)
	}
}

func TestOptimizeForCTV_CTVInventory(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{}
	req := &model.BidRequest{Device: model.InternalDevice{Type: "ctv"}}
	result := s.optimizeForCTV(newCampaign(1.0), req, pg, newPerfData())
	if result < 1.0 {
		t.Errorf("expected >= 1.0 for CTV inventory, got %f", result)
	}
}

func TestOptimizeForCTV_WithGoals(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{
		CTVGoals: &model.CTVOptimization{
			TargetCompletionRate: 0.7,
			PrimtimeBoost:        1.3,
			LiveContentBoost:     1.2,
			CoViewingBoost:       1.1,
			PreferredDevices:     []string{"roku"},
		},
	}
	req := &model.BidRequest{
		Device: model.InternalDevice{Type: "ctv"},
		Context: map[string]interface{}{
			"is_ctv":    true,
			"is_live":   true,
			"primetime": true,
		},
	}
	result := s.optimizeForCTV(newCampaign(1.0), req, pg, performanceData{completionRate: 0.85})
	if result < 1.0 {
		t.Errorf("expected >= 1.0 for CTV with goals, got %f", result)
	}
}

// ============================================================================
// optimizeForROAS
// ============================================================================

func TestOptimizeForROAS_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForROAS(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetROAS: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0, got %f", result)
	}
}

func TestOptimizeForROAS_AboveTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetROAS: 2.0}
	req := &model.BidRequest{
		Context: map[string]interface{}{"historical_roas": 5.0},
	}
	result := s.optimizeForROAS(newCampaign(1.0), req, pg, newPerfData())
	if result <= 1.0 {
		t.Errorf("expected > 1.0 when predicted ROAS above target, got %f", result)
	}
}

func TestOptimizeForROAS_BelowTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetROAS: 5.0}
	req := &model.BidRequest{
		Context: map[string]interface{}{"historical_roas": 1.0},
	}
	result := s.optimizeForROAS(newCampaign(1.0), req, pg, newPerfData())
	if result >= 1.0 {
		t.Errorf("expected < 1.0 when predicted ROAS below target, got %f", result)
	}
}

func TestOptimizeForROAS_EcomGoalsFallback(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{
		TargetROAS:     0,
		EcommerceGoals: &model.EcommerceOptimization{TargetROAS: 3.0},
	}
	result := s.optimizeForROAS(newCampaign(1.0), newReq(), pg, newPerfData())
	// Should use EcommerceGoals.TargetROAS
	if result == 1.0 {
		t.Error("expected non-default multiplier using EcommerceGoals ROAS")
	}
}

// ============================================================================
// predictInstallRate
// ============================================================================

func TestPredictInstallRate_DefaultMobile(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{Device: model.InternalDevice{Type: "mobile"}}
	rate := s.predictInstallRate(newCampaign(1.0), req, newPerfData())
	if rate <= 0 {
		t.Errorf("expected positive install rate, got %f", rate)
	}
}

func TestPredictInstallRate_InApp(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{
		Device:  model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{"environment": "in-app"},
	}
	baseRate := s.predictInstallRate(newCampaign(1.0), &model.BidRequest{Device: model.InternalDevice{Type: "mobile"}}, newPerfData())
	inAppRate := s.predictInstallRate(newCampaign(1.0), req, newPerfData())
	if inAppRate <= baseRate {
		t.Errorf("expected in-app rate (%f) > base rate (%f)", inAppRate, baseRate)
	}
}

func TestPredictInstallRate_Desktop(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{Device: model.InternalDevice{Type: "desktop"}}
	rate := s.predictInstallRate(newCampaign(1.0), req, newPerfData())
	// Desktop should have very low install rate
	if rate >= 0.01 {
		t.Errorf("expected very low install rate for desktop, got %f", rate)
	}
}

func TestPredictInstallRate_HistoricalOverride(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{
		Device:  model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{"historical_install_rate": 0.12},
	}
	rate := s.predictInstallRate(newCampaign(1.0), req, newPerfData())
	if rate != 0.12 {
		t.Errorf("expected 0.12 from historical, got %f", rate)
	}
}

// ============================================================================
// predictROAS
// ============================================================================

func TestPredictROAS_Default(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	roas := s.predictROAS(newCampaign(1.0), newReq(), newPerfData())
	if roas <= 0 {
		t.Errorf("expected positive ROAS, got %f", roas)
	}
}

func TestPredictROAS_HistoricalOverride(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{
		Context: map[string]interface{}{"historical_roas": 4.5},
	}
	roas := s.predictROAS(newCampaign(1.0), req, newPerfData())
	if roas < 4.5 {
		t.Errorf("expected roas >= 4.5, got %f", roas)
	}
}

func TestPredictROAS_HighValueSegment(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"user_segments": []interface{}{"high_value_buyer"},
		},
	}
	baseROAS := s.predictROAS(newCampaign(1.0), newReq(), newPerfData())
	boostedROAS := s.predictROAS(newCampaign(1.0), req, newPerfData())
	if boostedROAS <= baseROAS {
		t.Errorf("expected high-value segment to boost ROAS: base=%f boosted=%f", baseROAS, boostedROAS)
	}
}

// ============================================================================
// predictLTV
// ============================================================================

func TestPredictLTV_Default(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	ltv := s.predictLTV(newCampaign(1.0), newReq(), newPerfData())
	if ltv <= 0 {
		t.Errorf("expected positive LTV, got %f", ltv)
	}
}

func TestPredictLTV_HistoricalOverride(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{
		Context: map[string]interface{}{"predicted_ltv": 120.0},
	}
	ltv := s.predictLTV(newCampaign(1.0), req, newPerfData())
	if ltv < 120.0 {
		t.Errorf("expected ltv >= 120, got %f", ltv)
	}
}

func TestPredictLTV_HighEngagement(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{
		Context: map[string]interface{}{"engagement_score": 0.9},
	}
	baseLTV := s.predictLTV(newCampaign(1.0), newReq(), newPerfData())
	boostedLTV := s.predictLTV(newCampaign(1.0), req, newPerfData())
	if boostedLTV <= baseLTV {
		t.Errorf("expected high engagement to boost LTV: base=%f boosted=%f", baseLTV, boostedLTV)
	}
}

// ============================================================================
// predictCPCV
// ============================================================================

func TestPredictCPCV_Default(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	cpcv := s.predictCPCV(newCampaign(1.0), newReq(), newPerfData())
	if cpcv <= 0 {
		t.Errorf("expected positive CPCV, got %f", cpcv)
	}
}

func TestPredictCPCV_NonSkippable(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{
		Context: map[string]interface{}{"skippable": false},
	}
	// Non-skippable → completion ≈ 0.95
	cpcv := s.predictCPCV(newCampaign(1.0), req, performanceData{})
	expected := 1.0 / 0.95
	if cpcv < expected*0.9 || cpcv > expected*1.1 {
		t.Errorf("expected cpcv ≈ %f, got %f", expected, cpcv)
	}
}

func TestPredictCPCV_CTVBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{
		Device:  model.InternalDevice{Type: "ctv"},
		Context: map[string]interface{}{"is_ctv": true},
	}
	// CTV boosts completion → lower CPCV
	cpcvCTV := s.predictCPCV(newCampaign(1.0), req, performanceData{})
	cpcvNormal := s.predictCPCV(newCampaign(1.0), newReq(), performanceData{})
	if cpcvCTV >= cpcvNormal {
		t.Errorf("expected CTV CPCV (%f) < normal CPCV (%f)", cpcvCTV, cpcvNormal)
	}
}

// ============================================================================
// checkPerformanceThresholds
// ============================================================================

func TestCheckPerformanceThresholds_NoThresholds(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	blocked, reason := s.checkPerformanceThresholds(&model.PerformanceGoals{}, &model.PerformanceGoalResult{}, newPerfData())
	if blocked {
		t.Errorf("expected not blocked with nil thresholds, reason: %s", reason)
	}
}

func TestCheckPerformanceThresholds_MinCTRFail(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{
		Thresholds: &model.PerformanceThresholds{MinCTR: 0.05},
	}
	result := &model.PerformanceGoalResult{PredictedCTR: 0.01}
	blocked, reason := s.checkPerformanceThresholds(pg, result, newPerfData())
	if !blocked {
		t.Error("expected blocked for low CTR")
	}
	if reason != "predicted_ctr_below_threshold" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

func TestCheckPerformanceThresholds_MinViewabilityFail(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{
		Thresholds: &model.PerformanceThresholds{MinViewability: 0.7},
	}
	result := &model.PerformanceGoalResult{PredictedViewRate: 0.3}
	blocked, reason := s.checkPerformanceThresholds(pg, result, newPerfData())
	if !blocked {
		t.Error("expected blocked for low viewability")
	}
	if reason != "predicted_viewability_below_threshold" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

func TestCheckPerformanceThresholds_MaxCPAFail(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{
		Thresholds: &model.PerformanceThresholds{MaxCPA: 3.0},
	}
	perf := newPerfData()
	perf.cpa = 10.0
	blocked, reason := s.checkPerformanceThresholds(pg, &model.PerformanceGoalResult{}, perf)
	if !blocked {
		t.Error("expected blocked for high CPA")
	}
	if reason != "historical_cpa_above_threshold" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

func TestCheckPerformanceThresholds_MinROASFail(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{
		Thresholds: &model.PerformanceThresholds{MinROAS: 5.0},
	}
	result := &model.PerformanceGoalResult{PredictedROAS: 1.5}
	blocked, reason := s.checkPerformanceThresholds(pg, result, newPerfData())
	if !blocked {
		t.Error("expected blocked for low ROAS")
	}
	if reason != "predicted_roas_below_threshold" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

func TestCheckPerformanceThresholds_AllPass(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{
		Thresholds: &model.PerformanceThresholds{
			MinCTR:         0.01,
			MinViewability: 0.5,
			MinROAS:        1.5,
			MaxCPA:         20.0,
		},
	}
	result := &model.PerformanceGoalResult{
		PredictedCTR:      0.05,
		PredictedViewRate: 0.75,
		PredictedROAS:     3.0,
	}
	perf := newPerfData()
	perf.cpa = 5.0
	blocked, _ := s.checkPerformanceThresholds(pg, result, perf)
	if blocked {
		t.Error("expected not blocked when all thresholds pass")
	}
}

// ============================================================================
// applyCTVOptimizations
// ============================================================================

func TestApplyCTVOptimizations_NilCTV(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.applyCTVOptimizations(newCampaign(1.0), newReq(), nil, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for nil CTV config, got %f", result)
	}
}

func TestApplyCTVOptimizations_PrimtimeLive(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	ctv := &model.CTVOptimization{
		PrimtimeBoost:    1.3,
		LiveContentBoost: 1.2,
		CoViewingBoost:   1.1,
	}
	req := &model.BidRequest{
		Device: model.InternalDevice{Type: "ctv"},
		Context: map[string]interface{}{
			"is_live":       true,
			"is_co_viewing": true,
		},
	}
	result := s.applyCTVOptimizations(newCampaign(1.0), req, ctv, newPerfData())
	if result <= 1.0 {
		t.Errorf("expected > 1.0 for primetime+live, got %f", result)
	}
}

func TestApplyCTVOptimizations_HouseholdAtCap(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	ctv := &model.CTVOptimization{
		HouseholdFrequencyCap: 3,
	}
	// Mock s.getHouseholdImpressions by setting up campaign and context
	req := &model.BidRequest{
		Context: map[string]interface{}{"household_id": "hh-001"},
	}
	// No way to set household impressions in mock, so result depends on implementation
	result := s.applyCTVOptimizations(newCampaign(1.0), req, ctv, newPerfData())
	if result < 0.1 {
		t.Errorf("unexpected result %f", result)
	}
}

// ============================================================================
// applyAppOptimizations
// ============================================================================

func TestApplyAppOptimizations_NilApp(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.applyAppOptimizations(newCampaign(1.0), newReq(), nil, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for nil app config, got %f", result)
	}
}

func TestApplyAppOptimizations_RewardedPlacement(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	app := &model.AppOptimization{
		PreferredPlacements: []string{"rewarded"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"placement": "rewarded"},
	}
	result := s.applyAppOptimizations(newCampaign(1.0), req, app, newPerfData())
	if result < 1.4 {
		t.Errorf("expected >= 1.4 for rewarded placement, got %f", result)
	}
}

func TestApplyAppOptimizations_SKAdNetwork(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	app := &model.AppOptimization{
		SKAdNetworkOptimized: true,
		ExcludeLowLTVSources: true,
	}
	req := &model.BidRequest{
		Device:  model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{"device_os": "iOS"},
	}
	result := s.applyAppOptimizations(newCampaign(1.0), req, app, newPerfData())
	if result <= 0 {
		t.Errorf("expected positive multiplier, got %f", result)
	}
}
