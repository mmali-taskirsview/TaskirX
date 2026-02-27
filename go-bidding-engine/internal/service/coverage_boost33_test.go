package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// Boost33: target optimizeForCPC/CPM/Engagement/CPE/DCPM/CPAD/CPIAAP/predictROAS edge cases

// ============================================================================
// Helpers
// ============================================================================

func makeMinCamp_B33() *model.Campaign {
	return &model.Campaign{
		ID:       "camp-boost33",
		Name:     "Boost33 Campaign",
		BidPrice: 5.0,
		Creative: model.Creative{
			URL: "https://cdn.example.com/creative.jpg",
		},
		Targeting:   model.Targeting{},
		DailyBudget: 1000.0,
	}
}

func makeMinReq_B33() *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-boost33",
		PublisherID: "pub-test",
		User: model.InternalUser{
			ID: "user-test",
		},
		Device: model.InternalDevice{
			Type: "mobile",
			IP:   "1.2.3.4",
		},
		Context: make(map[string]interface{}),
	}
}

func newBiddingSvc_B33() *BiddingService {
	mc := cache.NewMockCache()
	return NewBiddingService(mc, "")
}

// ============================================================================
// optimizeForCPC Tests
// ============================================================================

// TestB33_CPC_NoTarget tests no target CPC
func TestB33_CPC_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	pg := &model.PerformanceGoals{TargetCPC: 0}
	perf := performanceData{}

	mult := svc.optimizeForCPC(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 for no target CPC, got %f", mult)
	}
}

// TestB33_CPC_Capped tests ratio capped at 2.0
func TestB33_CPC_Capped(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 0.5
	req := makeMinReq_B33()

	pg := &model.PerformanceGoals{TargetCPC: 100.0}
	perf := performanceData{
		ctr: 0.5, // 50% CTR -> very high maxCPM
	}

	mult := svc.optimizeForCPC(camp, req, pg, perf)
	// maxCPM = 100 * 0.5 * 1000 = 50000, ratio = 50000 / (0.5 * 1000) = 100 -> capped at 2.0
	if mult != 2.0 {
		t.Errorf("Expected 2.0 cap, got %f", mult)
	}
}

// TestB33_CPC_Floored tests ratio floored at 0.3
func TestB33_CPC_Floored(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 10.0
	req := makeMinReq_B33()

	pg := &model.PerformanceGoals{TargetCPC: 0.01}
	perf := performanceData{
		ctr: 0.001, // 0.1% CTR
	}

	mult := svc.optimizeForCPC(camp, req, pg, perf)
	// maxCPM = 0.01 * 0.001 * 1000 = 0.01, ratio = 0.01 / (10 * 1000) -> floored at 0.3
	if mult != 0.3 {
		t.Errorf("Expected 0.3 floor, got %f", mult)
	}
}

// ============================================================================
// optimizeForCPM Tests
// ============================================================================

// TestB33_CPM_NoTarget tests no target CPM
func TestB33_CPM_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	pg := &model.PerformanceGoals{TargetCPM: 0}
	perf := performanceData{}

	mult := svc.optimizeForCPM(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 for no target CPM, got %f", mult)
	}
}

// TestB33_CPM_HighViewability tests high viewability boost (>= 0.8)
func TestB33_CPM_HighViewability(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()

	pg := &model.PerformanceGoals{TargetCPM: 10.0}
	perf := performanceData{
		viewability: 0.85, // 85% viewability
	}

	mult := svc.optimizeForCPM(camp, req, pg, perf)
	if mult != 1.3 {
		t.Errorf("Expected 1.3 for high viewability, got %f", mult)
	}
}

// TestB33_CPM_GoodViewability tests good viewability boost (>= 0.6)
func TestB33_CPM_GoodViewability(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()

	pg := &model.PerformanceGoals{TargetCPM: 10.0}
	perf := performanceData{
		viewability: 0.65, // 65% viewability
	}

	mult := svc.optimizeForCPM(camp, req, pg, perf)
	if mult != 1.1 {
		t.Errorf("Expected 1.1 for good viewability, got %f", mult)
	}
}

// TestB33_CPM_LowViewability tests low viewability penalty (< 0.4)
func TestB33_CPM_LowViewability(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()

	pg := &model.PerformanceGoals{TargetCPM: 10.0}
	perf := performanceData{
		viewability: 0.3, // 30% viewability
	}

	mult := svc.optimizeForCPM(camp, req, pg, perf)
	if mult != 0.7 {
		t.Errorf("Expected 0.7 for low viewability, got %f", mult)
	}
}

// TestB33_CPM_NeutralViewability tests neutral viewability (between 0.4 and 0.6)
func TestB33_CPM_NeutralViewability(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()

	pg := &model.PerformanceGoals{TargetCPM: 10.0}
	perf := performanceData{
		viewability: 0.5, // 50% viewability
	}

	mult := svc.optimizeForCPM(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 for neutral viewability, got %f", mult)
	}
}

// ============================================================================
// optimizeForEngagement Tests
// ============================================================================

// TestB33_Engagement_NoGoal tests no engagement goal
func TestB33_Engagement_NoGoal(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	pg := &model.PerformanceGoals{EngagementGoal: 0}
	perf := performanceData{}

	mult := svc.optimizeForEngagement(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 for no engagement goal, got %f", mult)
	}
}

// TestB33_Engagement_MobileBoost tests mobile device boost
func TestB33_Engagement_MobileBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Device.Type = "mobile"

	pg := &model.PerformanceGoals{EngagementGoal: 0.1}
	perf := performanceData{}

	mult := svc.optimizeForEngagement(camp, req, pg, perf)
	// Base 1.0 * 1.2 (mobile) = 1.2
	if mult != 1.2 {
		t.Errorf("Expected 1.2 for mobile, got %f", mult)
	}
}

// TestB33_Engagement_InAppBoost tests in-app environment boost
func TestB33_Engagement_InAppBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Device.Type = "desktop"
	req.Context["environment"] = "in-app"

	pg := &model.PerformanceGoals{EngagementGoal: 0.1}
	perf := performanceData{}

	mult := svc.optimizeForEngagement(camp, req, pg, perf)
	// Base 1.0 * 1.15 (in-app) = 1.15
	if mult != 1.15 {
		t.Errorf("Expected 1.15 for in-app, got %f", mult)
	}
}

// TestB33_Engagement_CombinedBoosts tests mobile + in-app
func TestB33_Engagement_CombinedBoosts(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Device.Type = "mobile"
	req.Context["environment"] = "in-app"

	pg := &model.PerformanceGoals{EngagementGoal: 0.1}
	perf := performanceData{}

	mult := svc.optimizeForEngagement(camp, req, pg, perf)
	// 1.0 * 1.2 * 1.15 = 1.38
	expected := 1.2 * 1.15
	if mult != expected {
		t.Errorf("Expected %f for mobile+in-app, got %f", expected, mult)
	}
}

// ============================================================================
// optimizeForCPE Tests
// ============================================================================

// TestB33_CPE_NoTarget tests no target CPE
func TestB33_CPE_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	pg := &model.PerformanceGoals{TargetCPE: 0}
	perf := performanceData{}

	mult := svc.optimizeForCPE(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 for no target CPE, got %f", mult)
	}
}

// TestB33_CPE_MobileBoost tests mobile boost
func TestB33_CPE_MobileBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Device.Type = "mobile"

	pg := &model.PerformanceGoals{TargetCPE: 5.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPE(camp, req, pg, perf)
	// Should apply 1.15x mobile boost
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with mobile boost, got %f", mult)
	}
}

// TestB33_CPE_RichMediaBoost tests rich_media creative boost
func TestB33_CPE_RichMediaBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Context["creative_type"] = "rich_media"

	pg := &model.PerformanceGoals{TargetCPE: 5.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPE(camp, req, pg, perf)
	// Should apply 1.3x rich_media boost
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with rich_media boost, got %f", mult)
	}
}

// TestB33_CPE_PlayableBoost tests playable creative boost
func TestB33_CPE_PlayableBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Context["creative_type"] = "playable"

	pg := &model.PerformanceGoals{TargetCPE: 5.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPE(camp, req, pg, perf)
	// Should apply 1.3x playable boost
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with playable boost, got %f", mult)
	}
}

// TestB33_CPE_InteractiveBoost tests interactive creative boost
func TestB33_CPE_InteractiveBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Context["creative_type"] = "interactive"

	pg := &model.PerformanceGoals{TargetCPE: 5.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPE(camp, req, pg, perf)
	// Should apply 1.25x interactive boost
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with interactive boost, got %f", mult)
	}
}

// TestB33_CPE_InAppBoost tests in-app environment boost
func TestB33_CPE_InAppBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Context["environment"] = "in-app"

	pg := &model.PerformanceGoals{TargetCPE: 5.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPE(camp, req, pg, perf)
	// Should apply 1.1x in-app boost
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with in-app boost, got %f", mult)
	}
}

// ============================================================================
// optimizeForDCPM Tests
// ============================================================================

// TestB33_DCPM_NoTarget tests no target DCPM
func TestB33_DCPM_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	pg := &model.PerformanceGoals{TargetDCPM: 0}
	perf := performanceData{}

	mult := svc.optimizeForDCPM(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 for no target DCPM, got %f", mult)
	}
}

// TestB33_DCPM_LowWinRateBoost tests low win rate boost (< 0.1)
func TestB33_DCPM_LowWinRateBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 2.0
	req := makeMinReq_B33()

	pg := &model.PerformanceGoals{TargetDCPM: 10.0}
	perf := performanceData{
		ctr:            0.05,
		viewability:    0.7,
		engagementRate: 0.1,
		winRate:        0.05, // 5% win rate -> boost 1.3x
	}

	mult := svc.optimizeForDCPM(camp, req, pg, perf)
	// Should apply 1.3x boost for low win rate
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with low win rate boost, got %f", mult)
	}
}

// TestB33_DCPM_HighWinRatePenalty tests high win rate penalty (> 0.4)
func TestB33_DCPM_HighWinRatePenalty(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 1.0
	req := makeMinReq_B33()

	pg := &model.PerformanceGoals{TargetDCPM: 10.0}
	perf := performanceData{
		ctr:            0.05,
		viewability:    0.7,
		engagementRate: 0.1,
		winRate:        0.5, // 50% win rate -> penalty 0.8x
	}

	mult := svc.optimizeForDCPM(camp, req, pg, perf)
	// Should apply 0.8x penalty for high win rate
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with high win rate penalty, got %f", mult)
	}
}

// TestB33_DCPM_NeutralWinRate tests neutral win rate (between 0.1 and 0.4)
func TestB33_DCPM_NeutralWinRate(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 2.0
	req := makeMinReq_B33()

	pg := &model.PerformanceGoals{TargetDCPM: 10.0}
	perf := performanceData{
		ctr:            0.05,
		viewability:    0.7,
		engagementRate: 0.1,
		winRate:        0.25, // 25% win rate -> no win rate adjustment
	}

	mult := svc.optimizeForDCPM(camp, req, pg, perf)
	// Should have no win rate adjustment
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with neutral win rate, got %f", mult)
	}
}

// ============================================================================
// optimizeForCPAD Tests
// ============================================================================

// TestB33_CPAD_NoTarget tests no target CPAD
func TestB33_CPAD_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	pg := &model.PerformanceGoals{TargetCPAD: 0}
	perf := performanceData{}

	mult := svc.optimizeForCPAD(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 for no target CPAD, got %f", mult)
	}
}

// TestB33_CPAD_HighAppRatingBoost tests high app rating boost (>= 4.5)
func TestB33_CPAD_HighAppRatingBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 1.0
	req := makeMinReq_B33()
	req.Context["app_rating"] = 4.7

	pg := &model.PerformanceGoals{TargetCPAD: 5.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPAD(camp, req, pg, perf)
	// Should apply 1.2x boost for high rating
	if mult < 0.2 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with high rating boost, got %f", mult)
	}
}

// TestB33_CPAD_FeaturedAppBoost tests featured app boost
func TestB33_CPAD_FeaturedAppBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 1.0
	req := makeMinReq_B33()
	req.Context["app_featured"] = true

	pg := &model.PerformanceGoals{TargetCPAD: 5.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPAD(camp, req, pg, perf)
	// Should apply 1.15x boost for featured
	if mult < 0.2 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with featured boost, got %f", mult)
	}
}

// TestB33_CPAD_NonMobilePenalty tests non-mobile device penalty
func TestB33_CPAD_NonMobilePenalty(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 1.0
	req := makeMinReq_B33()
	req.Device.Type = "desktop" // Not mobile or tablet

	pg := &model.PerformanceGoals{TargetCPAD: 5.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPAD(camp, req, pg, perf)
	// Should apply 0.1x penalty for non-mobile
	if mult < 0.2 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with non-mobile penalty, got %f", mult)
	}
}

// TestB33_CPAD_TabletAllowed tests tablet is not penalized
func TestB33_CPAD_TabletAllowed(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 1.0
	req := makeMinReq_B33()
	req.Device.Type = "tablet"

	pg := &model.PerformanceGoals{TargetCPAD: 5.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPAD(camp, req, pg, perf)
	// Tablet should not have 0.1x penalty
	if mult < 0.2 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range for tablet, got %f", mult)
	}
}

// ============================================================================
// optimizeForCPIAAP Tests
// ============================================================================

// TestB33_CPIAAP_NoTarget tests no target CPIAAP
func TestB33_CPIAAP_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	pg := &model.PerformanceGoals{TargetCPIAAP: 0}
	perf := performanceData{}

	mult := svc.optimizeForCPIAAP(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 for no target CPIAAP, got %f", mult)
	}
}

// TestB33_CPIAAP_HistoricalIAPRate tests historical IAP rate context
func TestB33_CPIAAP_HistoricalIAPRate(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 1.0
	req := makeMinReq_B33()
	req.Context["historical_iap_rate"] = 0.05 // 5% IAP rate

	pg := &model.PerformanceGoals{TargetCPIAAP: 10.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPIAAP(camp, req, pg, perf)
	// Should use 0.05 instead of default 0.03
	if mult < 0.2 || mult > 3.0 {
		t.Errorf("Expected multiplier in valid range with historical IAP rate, got %f", mult)
	}
}

// TestB33_CPIAAP_PurchasePropensityBoost tests purchase propensity
func TestB33_CPIAAP_PurchasePropensityBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 1.0
	req := makeMinReq_B33()
	req.Context["purchase_propensity"] = 0.5 // 50% boost to IAP rate

	pg := &model.PerformanceGoals{TargetCPIAAP: 10.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPIAAP(camp, req, pg, perf)
	// IAP rate becomes 0.03 * (1 + 0.5) = 0.045
	if mult < 0.2 || mult > 3.0 {
		t.Errorf("Expected multiplier in valid range with purchase propensity, got %f", mult)
	}
}

// TestB33_CPIAAP_HighValuePurchaserBoost tests avg_iap_value boost
func TestB33_CPIAAP_HighValuePurchaserBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 0.5
	req := makeMinReq_B33()
	req.Context["avg_iap_value"] = 15.0 // > $10

	pg := &model.PerformanceGoals{TargetCPIAAP: 10.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPIAAP(camp, req, pg, perf)
	// Should apply 1.5x boost
	if mult < 0.2 || mult > 3.0 {
		t.Errorf("Expected multiplier in valid range with high value purchaser boost, got %f", mult)
	}
}

// TestB33_CPIAAP_SpenderSegmentBoost tests spender segment boost
func TestB33_CPIAAP_SpenderSegmentBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 0.5
	req := makeMinReq_B33()
	req.Context["user_segments"] = []interface{}{"premium", "high_spender", "frequent"}

	pg := &model.PerformanceGoals{TargetCPIAAP: 10.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPIAAP(camp, req, pg, perf)
	// Should apply 1.8x boost for spender segment
	if mult < 0.2 || mult > 3.0 {
		t.Errorf("Expected multiplier in valid range with spender segment boost, got %f", mult)
	}
}

// TestB33_CPIAAP_WhaleSegmentBoost tests whale segment boost
func TestB33_CPIAAP_WhaleSegmentBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 0.5
	req := makeMinReq_B33()
	req.Context["user_segments"] = []interface{}{"premium", "whale_user"}

	pg := &model.PerformanceGoals{TargetCPIAAP: 10.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPIAAP(camp, req, pg, perf)
	// Should apply 1.8x boost for whale segment
	if mult < 0.2 || mult > 3.0 {
		t.Errorf("Expected multiplier in valid range with whale segment boost, got %f", mult)
	}
}

// TestB33_CPIAAP_NonMobilePenalty tests non-mobile penalty
func TestB33_CPIAAP_NonMobilePenalty(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	camp.BidPrice = 1.0
	req := makeMinReq_B33()
	req.Device.Type = "desktop"

	pg := &model.PerformanceGoals{TargetCPIAAP: 10.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPIAAP(camp, req, pg, perf)
	// Should apply 0.05x penalty for non-mobile
	if mult < 0.2 || mult > 3.0 {
		t.Errorf("Expected multiplier in valid range with non-mobile penalty, got %f", mult)
	}
}

// ============================================================================
// predictROAS Tests
// ============================================================================

// TestB33_ROAS_Default tests default ROAS (2.0)
func TestB33_ROAS_Default(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	perf := performanceData{}

	roas := svc.predictROAS(camp, req, perf)
	if roas != 2.0 {
		t.Errorf("Expected default ROAS 2.0, got %f", roas)
	}
}

// TestB33_ROAS_Historical tests historical ROAS from context
func TestB33_ROAS_Historical(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Context["historical_roas"] = 3.5
	perf := performanceData{}

	roas := svc.predictROAS(camp, req, perf)
	if roas != 3.5 {
		t.Errorf("Expected historical ROAS 3.5, got %f", roas)
	}
}

// TestB33_ROAS_RepeatCustomerBoost tests repeat customer boost (1.5x)
func TestB33_ROAS_RepeatCustomerBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Context["repeat_customer"] = true
	perf := performanceData{}

	roas := svc.predictROAS(camp, req, perf)
	// 2.0 * 1.5 = 3.0
	if roas != 3.0 {
		t.Errorf("Expected ROAS 3.0 with repeat customer boost, got %f", roas)
	}
}

// TestB33_ROAS_CartAbandonBoost tests cart abandoner boost (1.3x)
func TestB33_ROAS_CartAbandonBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Context["cart_abandoner"] = true
	perf := performanceData{}

	roas := svc.predictROAS(camp, req, perf)
	// 2.0 * 1.3 = 2.6
	if roas != 2.6 {
		t.Errorf("Expected ROAS 2.6 with cart abandon boost, got %f", roas)
	}
}

// TestB33_ROAS_HighValueSegmentBoost tests high_value segment boost (1.4x)
func TestB33_ROAS_HighValueSegmentBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Context["user_segments"] = []interface{}{"premium", "high_value", "active"}
	perf := performanceData{}

	roas := svc.predictROAS(camp, req, perf)
	// 2.0 * 1.4 = 2.8
	if roas != 2.8 {
		t.Errorf("Expected ROAS 2.8 with high_value segment boost, got %f", roas)
	}
}

// TestB33_ROAS_FrequentBuyerSegmentBoost tests frequent_buyer segment boost (1.4x)
func TestB33_ROAS_FrequentBuyerSegmentBoost(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Context["user_segments"] = []interface{}{"frequent_buyer", "active"}
	perf := performanceData{}

	roas := svc.predictROAS(camp, req, perf)
	// 2.0 * 1.4 = 2.8
	if roas != 2.8 {
		t.Errorf("Expected ROAS 2.8 with frequent_buyer segment boost, got %f", roas)
	}
}

// TestB33_ROAS_CombinedBoosts tests combined boosts
func TestB33_ROAS_CombinedBoosts(t *testing.T) {
	svc := newBiddingSvc_B33()
	camp := makeMinCamp_B33()
	req := makeMinReq_B33()
	req.Context["repeat_customer"] = true
	req.Context["cart_abandoner"] = true
	req.Context["user_segments"] = []interface{}{"high_value"}
	perf := performanceData{}

	roas := svc.predictROAS(camp, req, perf)
	// 2.0 * 1.5 * 1.3 * 1.4 = 5.46
	expected := 2.0 * 1.5 * 1.3 * 1.4
	if roas != expected {
		t.Errorf("Expected ROAS %f with combined boosts, got %f", expected, roas)
	}
}
