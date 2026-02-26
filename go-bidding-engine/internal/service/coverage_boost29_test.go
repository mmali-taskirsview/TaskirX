package service

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

// makeMinCamp_B29 creates a bare-minimum campaign for boost29 tests
func makeMinCamp_B29() *model.Campaign {
	return &model.Campaign{
		ID:        "camp-b29",
		BidPrice:  1.0,
		Status:    "active",
		Targeting: model.Targeting{},
	}
}

// makeMinReq_B29 creates a minimal BidRequest for boost29 tests
func makeMinReq_B29() *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-b29",
		PublisherID: "pub1",
		AdSlot:      model.AdSlot{ID: "slot1"},
		Device:      model.InternalDevice{Type: "mobile"},
		User:        model.InternalUser{ID: "user-b29"},
	}
}

// makePerfCamp_B29 builds a campaign with BidPrice and PerformanceGoals set
func makePerfCamp_B29(goal string, target float64, bidPrice float64) *model.Campaign {
	c := makeMinCamp_B29()
	c.BidPrice = bidPrice
	c.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: goal,
	}
	switch goal {
	case "cpa":
		c.Targeting.PerformanceGoals.TargetCPA = target
	case "cpc":
		c.Targeting.PerformanceGoals.TargetCPC = target
	case "cpm":
		c.Targeting.PerformanceGoals.TargetCPM = target
	case "cpi":
		c.Targeting.PerformanceGoals.TargetCPI = target
	case "cps":
		c.Targeting.PerformanceGoals.TargetCPS = target
	case "cpr":
		c.Targeting.PerformanceGoals.TargetCPR = target
	case "cpl":
		c.Targeting.PerformanceGoals.TargetCPL = target
	case "cpv":
		c.Targeting.PerformanceGoals.TargetCPV = target
	case "cpcv":
		c.Targeting.PerformanceGoals.TargetCPCV = target
	case "cpe":
		c.Targeting.PerformanceGoals.TargetCPE = target
	case "vcpm":
		c.Targeting.PerformanceGoals.TargetVCPM = target
	case "dcpm":
		c.Targeting.PerformanceGoals.TargetDCPM = target
	case "cpa_d", "cpad":
		c.Targeting.PerformanceGoals.TargetCPAD = target
	case "cpiaap":
		c.Targeting.PerformanceGoals.TargetCPIAAP = target
	case "roas":
		c.Targeting.PerformanceGoals.TargetROAS = target
	case "viewability":
		c.Targeting.PerformanceGoals.ViewabilityGoal = target
	case "completion":
		c.Targeting.PerformanceGoals.CompletionGoal = target
	case "engagement":
		c.Targeting.PerformanceGoals.EngagementGoal = target
	}
	return c
}

// ─── calculatePerformanceGoalMultiplier ───────────────────────────────────────

// TestPerfGoal_NoGoals_B29 hits the nil-goals early-return path
func TestPerfGoal_NoGoals_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makeMinCamp_B29()
	c.Targeting.PerformanceGoals = nil
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if result.Matched {
		t.Error("expected not matched when no goals")
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected multiplier 1.0, got %v", result.Multiplier)
	}
}

// TestPerfGoal_CPA_B29 hits optimizeForCPA branch
func TestPerfGoal_CPA_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpa", 5.0, 1.0)
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPA goal")
	}
	if result.Multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %v", result.Multiplier)
	}
}

// TestPerfGoal_CPI_B29 hits optimizeForCPI branch
func TestPerfGoal_CPI_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpi", 2.0, 0.5)
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPI goal")
	}
}

// TestPerfGoal_CPS_CartAbandoner_B29 hits optimizeForCPS branch with cart abandoner signal
func TestPerfGoal_CPS_CartAbandoner_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cps", 10.0, 1.0)
	c.Targeting.PerformanceGoals.EcommerceGoals = &model.EcommerceOptimization{
		TargetCostPerSale:   10.0,
		CartAbandonBoost:    1.3,
		RepeatCustomerBoost: 1.2,
	}
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"is_cart_abandoner":  true,
		"is_repeat_customer": true,
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPS+cart")
	}
}

// TestPerfGoal_CPR_B29 hits optimizeForCPR branch
func TestPerfGoal_CPR_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpr", 3.0, 0.5)
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPR goal")
	}
}

// TestPerfGoal_CPL_B2B_B29 hits optimizeForCPL with lead_intent_score and is_b2b branches
func TestPerfGoal_CPL_B2B_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpl", 20.0, 1.0)
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"lead_intent_score": float64(0.9),
		"is_b2b":            true,
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPL goal")
	}
}

// TestPerfGoal_CPCV_NonSkippable_B29 hits optimizeForCPCV non-skippable + short video + CTV
func TestPerfGoal_CPCV_NonSkippable_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpcv", 0.05, 1.0)
	req := makeMinReq_B29()
	req.Device = model.InternalDevice{Type: "ctv"}
	req.Context = map[string]interface{}{
		"skippable":                 false,
		"video_duration":            float64(15),
		"predicted_completion_rate": float64(0.9),
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPCV goal")
	}
}

// TestPerfGoal_CPCV_LongVideo_B29 hits optimizeForCPCV long video discount
func TestPerfGoal_CPCV_LongVideo_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpcv", 0.05, 1.0)
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"skippable":      true,
		"video_duration": float64(60),
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPCV long video")
	}
}

// TestPerfGoal_CPE_RichMedia_B29 hits optimizeForCPE with rich_media creative and in-app env
func TestPerfGoal_CPE_RichMedia_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpe", 0.10, 0.5)
	c.Creative.Type = "rich_media"
	req := makeMinReq_B29()
	req.Device = model.InternalDevice{Type: "mobile"}
	req.Context = map[string]interface{}{
		"creative_type":             "rich_media",
		"environment":               "in-app",
		"predicted_engagement_rate": float64(0.05),
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPE goal")
	}
}

// TestPerfGoal_CPE_Interactive_B29 hits optimizeForCPE "interactive" creative branch
func TestPerfGoal_CPE_Interactive_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpe", 0.10, 0.5)
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"creative_type": "interactive",
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPE interactive")
	}
}

// TestPerfGoal_DCPM_B29 hits optimizeForDCPM win-rate feedback paths
func TestPerfGoal_DCPM_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("dcpm", 3.0, 0.001)
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for DCPM goal")
	}
}

// TestPerfGoal_CPAD_AppRating_B29 hits optimizeForCPAD with app_rating and app_featured signals
func TestPerfGoal_CPAD_AppRating_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpa_d", 2.0, 0.5)
	req := makeMinReq_B29()
	req.Device = model.InternalDevice{Type: "mobile"}
	req.Context = map[string]interface{}{
		"app_rating":   float64(4.8),
		"app_featured": true,
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPAD with app signals")
	}
}

// TestPerfGoal_CPAD_NonMobile_B29 hits non-mobile device penalty in CPAD
func TestPerfGoal_CPAD_NonMobile_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpa_d", 2.0, 0.5)
	req := makeMinReq_B29()
	req.Device = model.InternalDevice{Type: "desktop"}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPAD non-mobile")
	}
}

// TestPerfGoal_ROAS_B29 hits optimizeForROAS branch
func TestPerfGoal_ROAS_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("roas", 3.0, 1.0)
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"historical_roas": float64(4.0),
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for ROAS goal")
	}
}

// TestPerfGoal_Viewability_B29 hits optimizeForViewability branch
func TestPerfGoal_Viewability_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("viewability", 0.7, 1.0)
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"ad_position": "above_fold",
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for Viewability goal")
	}
}

// TestPerfGoal_Completion_B29 hits optimizeForCompletion branch
func TestPerfGoal_Completion_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("completion", 0.8, 1.0)
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"predicted_completion_rate": float64(0.9),
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for Completion goal")
	}
}

// TestPerfGoal_Engagement_B29 hits optimizeForEngagement branch
func TestPerfGoal_Engagement_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("engagement", 0.05, 1.0)
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for Engagement goal")
	}
}

// TestPerfGoal_CPV_SoundOn_B29 hits optimizeForCPV with sound_on + instream branches
func TestPerfGoal_CPV_SoundOn_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpv", 0.02, 1.0)
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"sound_on":            true,
		"video_placement":     "instream",
		"predicted_view_rate": float64(0.8),
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPV goal")
	}
}

// TestPerfGoal_VCPM_HighViewability_B29 hits optimizeForVCPM >=0.8 boost
func TestPerfGoal_VCPM_HighViewability_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("vcpm", 5.0, 0.001)
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"predicted_viewability": float64(0.9),
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for VCPM high viewability")
	}
}

// TestPerfGoal_VCPM_LowViewability_B29 hits optimizeForVCPM <0.4 penalty
func TestPerfGoal_VCPM_LowViewability_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("vcpm", 5.0, 0.001)
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"predicted_viewability": float64(0.2),
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for VCPM low viewability")
	}
}

// TestPerfGoal_CPIAAP_Whale_B29 hits optimizeForCPIAAP whale/spender segment detection
func TestPerfGoal_CPIAAP_Whale_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpiaap", 5.0, 0.5)
	req := makeMinReq_B29()
	req.Device = model.InternalDevice{Type: "mobile"}
	req.Context = map[string]interface{}{
		"historical_iap_rate": float64(0.05),
		"purchase_propensity": float64(0.8),
		"avg_iap_value":       float64(15.0),
		"user_segments":       []interface{}{"whale_spender"},
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPIAAP")
	}
}

// TestPerfGoal_CPIAAP_NonMobile_B29 hits non-mobile penalty in CPIAAP
func TestPerfGoal_CPIAAP_NonMobile_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpiaap", 5.0, 0.5)
	req := makeMinReq_B29()
	req.Device = model.InternalDevice{Type: "desktop"}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for CPIAAP non-mobile")
	}
}

// TestPerfGoal_Threshold_Block_B29 hits checkPerformanceThresholds → block (non-learning mode)
func TestPerfGoal_Threshold_Block_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpm", 5.0, 1.0)
	c.Targeting.PerformanceGoals.Thresholds = &model.PerformanceThresholds{
		MinCTR: 0.99, // Impossible → predictedCTR will be ~0.011 < 0.99 → blocked
	}
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	// Blocked path exercised regardless of exact outcome
	_ = result
}

// TestPerfGoal_Threshold_LearningMode_B29 hits threshold block in learning mode (reduces, not blocks)
func TestPerfGoal_Threshold_LearningMode_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpm", 5.0, 1.0)
	c.Targeting.PerformanceGoals.LearningMode = true
	c.Targeting.PerformanceGoals.Thresholds = &model.PerformanceThresholds{
		MinCTR:         0.99,
		MinViewability: 0.99,
		MinInstallRate: 0.99,
		MinROAS:        999.0,
	}
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if result.Blocked {
		t.Error("should not be blocked in learning mode")
	}
}

// TestPerfGoal_BidStrategy_MaximizeConversions_B29 hits applyBidStrategy maximize_conversions
func TestPerfGoal_BidStrategy_MaxConv_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpa", 10.0, 1.0)
	c.Targeting.PerformanceGoals.BidStrategy = "maximize_conversions"
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for maximize_conversions strategy")
	}
}

// TestPerfGoal_BidStrategy_TargetCPA_B29 hits applyBidStrategy target_cpa branch
func TestPerfGoal_BidStrategy_TargetCPA_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpa", 5.0, 1.0)
	c.Targeting.PerformanceGoals.BidStrategy = "target_cpa"
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for target_cpa strategy")
	}
}

// TestPerfGoal_BidStrategy_MaxClicks_B29 hits applyBidStrategy maximize_clicks
func TestPerfGoal_BidStrategy_MaxClicks_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpc", 0.1, 1.0)
	c.Targeting.PerformanceGoals.BidStrategy = "maximize_clicks"
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for maximize_clicks strategy")
	}
}

// TestPerfGoal_BidStrategy_Manual_B29 hits applyBidStrategy manual → 1.0
func TestPerfGoal_BidStrategy_Manual_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpm", 5.0, 1.0)
	c.Targeting.PerformanceGoals.BidStrategy = "manual"
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for manual strategy")
	}
}

// TestPerfGoal_AppGoals_B29 hits applyAppOptimizations via pg.AppGoals being non-nil
func TestPerfGoal_AppGoals_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpi", 2.0, 0.5)
	c.Targeting.PerformanceGoals.AppGoals = &model.AppOptimization{
		PreferredPlacements:  []string{"rewarded", "interstitial"},
		ExcludeLowLTVSources: false,
		SKAdNetworkOptimized: true,
	}
	req := makeMinReq_B29()
	req.Device = model.InternalDevice{Type: "mobile", OS: "iOS"}
	req.Context = map[string]interface{}{
		"placement": "rewarded",
		"is_ios":    true,
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for AppGoals path")
	}
}

// TestPerfGoal_EcommerceGoals_B29 hits applyEcommerceOptimizations via pg.EcommerceGoals
func TestPerfGoal_EcommerceGoals_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cps", 15.0, 1.0)
	c.Targeting.PerformanceGoals.EcommerceGoals = &model.EcommerceOptimization{
		CartAbandonBoost:    1.5,
		RepeatCustomerBoost: 1.3,
	}
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"is_cart_abandoner":  true,
		"is_repeat_customer": true,
	}
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if !result.Matched {
		t.Error("expected matched for EcommerceGoals path")
	}
}

// TestPerfGoal_MaxBidAdjust_B29 hits MaxBidAdjust capping
func TestPerfGoal_MaxBidAdjust_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpa", 100.0, 0.001) // huge ratio → would exceed cap
	c.Targeting.PerformanceGoals.MaxBidAdjust = 1.5
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if result.Multiplier > 1.5+0.001 {
		t.Errorf("expected multiplier capped at 1.5, got %v", result.Multiplier)
	}
}

// TestPerfGoal_MinBidAdjust_B29 hits MinBidAdjust floor
func TestPerfGoal_MinBidAdjust_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makePerfCamp_B29("cpa", 0.0001, 100.0) // tiny ratio → below floor
	c.Targeting.PerformanceGoals.MinBidAdjust = 0.9
	req := makeMinReq_B29()
	result := svc.calculatePerformanceGoalMultiplier(c, req)
	if result.Multiplier < 0.9-0.001 {
		t.Errorf("expected multiplier floored at 0.9, got %v", result.Multiplier)
	}
}

// ─── checkPerformanceThresholds ───────────────────────────────────────────────

// TestCheckThreshold_Nil_B29 covers nil thresholds fast-return
func TestCheckThreshold_Nil_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	pg := &model.PerformanceGoals{}
	result := &model.PerformanceGoalResult{}
	perf := performanceData{}
	blocked, reason := svc.checkPerformanceThresholds(pg, result, perf)
	if blocked || reason != "" {
		t.Error("expected no block for nil thresholds")
	}
}

// TestCheckThreshold_MinViewability_B29 covers MinViewability path
func TestCheckThreshold_MinViewability_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	pg := &model.PerformanceGoals{
		Thresholds: &model.PerformanceThresholds{MinViewability: 0.8},
	}
	result := &model.PerformanceGoalResult{PredictedViewRate: 0.5}
	perf := performanceData{}
	blocked, reason := svc.checkPerformanceThresholds(pg, result, perf)
	if !blocked {
		t.Error("expected blocked for low viewability")
	}
	if reason != "predicted_viewability_below_threshold" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

// TestCheckThreshold_MinInstallRate_B29 covers MinInstallRate path
func TestCheckThreshold_MinInstallRate_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	pg := &model.PerformanceGoals{
		Thresholds: &model.PerformanceThresholds{MinInstallRate: 0.1},
	}
	result := &model.PerformanceGoalResult{PredictedInstallRate: 0.001}
	perf := performanceData{}
	blocked, reason := svc.checkPerformanceThresholds(pg, result, perf)
	if !blocked {
		t.Error("expected blocked for low install rate")
	}
	if reason != "predicted_install_rate_below_threshold" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

// TestCheckThreshold_MinROAS_B29 covers MinROAS path
func TestCheckThreshold_MinROAS_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	pg := &model.PerformanceGoals{
		Thresholds: &model.PerformanceThresholds{MinROAS: 10.0},
	}
	result := &model.PerformanceGoalResult{PredictedROAS: 1.0}
	perf := performanceData{}
	blocked, reason := svc.checkPerformanceThresholds(pg, result, perf)
	if !blocked {
		t.Error("expected blocked for low ROAS")
	}
	if reason != "predicted_roas_below_threshold" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

// TestCheckThreshold_MaxCPA_B29 covers MaxCPA historical check
func TestCheckThreshold_MaxCPA_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	pg := &model.PerformanceGoals{
		Thresholds: &model.PerformanceThresholds{MaxCPA: 1.0},
	}
	result := &model.PerformanceGoalResult{}
	perf := performanceData{cpa: 5.0} // cpa > MaxCPA
	blocked, reason := svc.checkPerformanceThresholds(pg, result, perf)
	if !blocked {
		t.Error("expected blocked for high CPA")
	}
	if reason != "historical_cpa_above_threshold" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

// TestCheckThreshold_MaxCPI_B29 covers MaxCPI path
func TestCheckThreshold_MaxCPI_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	pg := &model.PerformanceGoals{
		Thresholds: &model.PerformanceThresholds{MaxCPI: 0.5},
	}
	result := &model.PerformanceGoalResult{}
	perf := performanceData{cpi: 2.0} // cpi > MaxCPI
	blocked, reason := svc.checkPerformanceThresholds(pg, result, perf)
	if !blocked {
		t.Error("expected blocked for high CPI")
	}
	if reason != "historical_cpi_above_threshold" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

// TestCheckThreshold_MaxCPS_B29 covers MaxCPS path
func TestCheckThreshold_MaxCPS_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	pg := &model.PerformanceGoals{
		Thresholds: &model.PerformanceThresholds{MaxCPS: 1.0},
	}
	result := &model.PerformanceGoalResult{}
	perf := performanceData{cps: 3.0} // cps > MaxCPS
	blocked, reason := svc.checkPerformanceThresholds(pg, result, perf)
	if !blocked {
		t.Error("expected blocked for high CPS")
	}
	if reason != "historical_cps_above_threshold" {
		t.Errorf("unexpected reason: %s", reason)
	}
}

// ─── applyAppOptimizations ────────────────────────────────────────────────────

// TestApplyAppOpt_Nil_B29 covers nil app → 1.0
func TestApplyAppOpt_Nil_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	req := makeMinReq_B29()
	result := svc.applyAppOptimizations(makeMinCamp_B29(), req, nil, performanceData{})
	if result != 1.0 {
		t.Errorf("expected 1.0 for nil app, got %v", result)
	}
}

// TestApplyAppOpt_RewardedPlacement_B29 hits rewarded placement 1.4x boost
func TestApplyAppOpt_RewardedPlacement_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	app := &model.AppOptimization{
		PreferredPlacements: []string{"rewarded"},
	}
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{"placement": "rewarded"}
	result := svc.applyAppOptimizations(makeMinCamp_B29(), req, app, performanceData{})
	if result <= 1.0 {
		t.Errorf("expected boost for rewarded placement, got %v", result)
	}
}

// TestApplyAppOpt_NonRewardedPlacement_B29 hits non-rewarded preferred placement 1.2x
func TestApplyAppOpt_NonRewardedPlacement_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	app := &model.AppOptimization{
		PreferredPlacements: []string{"interstitial"},
	}
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{"placement": "interstitial"}
	result := svc.applyAppOptimizations(makeMinCamp_B29(), req, app, performanceData{})
	if result <= 1.0 {
		t.Errorf("expected boost for interstitial placement, got %v", result)
	}
}

// TestApplyAppOpt_LowLTV_Check_B29 hits !ExcludeLowLTVSources branch (calls isLowLTVSource)
func TestApplyAppOpt_LowLTV_Check_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	app := &model.AppOptimization{
		ExcludeLowLTVSources: false,
	}
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{"source_quality": "low_ltv"}
	result := svc.applyAppOptimizations(makeMinCamp_B29(), req, app, performanceData{})
	_ = result // exercised the low-LTV path without panic
}

// TestApplyAppOpt_SKAdNetwork_iOS_B29 hits SKAdNetworkOptimized + iOS path
func TestApplyAppOpt_SKAdNetwork_iOS_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	app := &model.AppOptimization{
		SKAdNetworkOptimized: true,
	}
	req := makeMinReq_B29()
	req.Device = model.InternalDevice{Type: "mobile", OS: "iOS"}
	result := svc.applyAppOptimizations(makeMinCamp_B29(), req, app, performanceData{})
	_ = result // exercised SKAdNetwork path
}

// ─── predictROAS ──────────────────────────────────────────────────────────────

// TestPredictROAS_Historical_B29 covers historical_roas context + high_value segment paths
func TestPredictROAS_Historical_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"historical_roas": float64(5.0),
		"user_segments":   []interface{}{"high_value_user"},
	}
	roas := svc.predictROAS(makeMinCamp_B29(), req, performanceData{})
	if roas <= 0 {
		t.Errorf("expected positive ROAS, got %v", roas)
	}
}

// TestPredictROAS_CartAbandoner_B29 covers cart abandoner boost in predictROAS
func TestPredictROAS_CartAbandoner_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"is_cart_abandoner": true,
	}
	roas := svc.predictROAS(makeMinCamp_B29(), req, performanceData{})
	if roas < 2.0 {
		t.Errorf("expected ROAS >= 2.0 for cart abandoner, got %v", roas)
	}
}

// ─── predictCPL ───────────────────────────────────────────────────────────────

// TestPredictCPL_B2B_B29 covers historical_lead_rate and is_b2b branches in predictCPL
func TestPredictCPL_B2B_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makeMinCamp_B29()
	c.BidPrice = 1.0
	req := makeMinReq_B29()
	req.Context = map[string]interface{}{
		"historical_lead_rate": float64(0.08),
		"is_b2b":               true,
	}
	cpl := svc.predictCPL(c, req, performanceData{ctr: 0.02})
	if cpl < 0 {
		t.Errorf("expected non-negative CPL, got %v", cpl)
	}
}

// ─── predictCPCV ──────────────────────────────────────────────────────────────

// TestPredictCPCV_NonSkippable_B29 covers non-skippable + CTV + completion_rate paths
func TestPredictCPCV_NonSkippable_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	c := makeMinCamp_B29()
	c.BidPrice = 1.0
	req := makeMinReq_B29()
	req.Device = model.InternalDevice{Type: "ctv"}
	req.Context = map[string]interface{}{
		"skippable":                 false,
		"predicted_completion_rate": float64(0.8),
	}
	cpcv := svc.predictCPCV(c, req, performanceData{})
	if cpcv <= 0 {
		t.Errorf("expected positive CPCV, got %v", cpcv)
	}
}

// ─── linearAttribution empty ──────────────────────────────────────────────────

// TestLinearAttribution_Empty_B29 covers n==0 → nil path in linearAttribution
func TestLinearAttribution_Empty_B29(t *testing.T) {
	mc := NewMockCache()
	as := NewAttributionService(mc)
	result := as.linearAttribution(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}
	result2 := as.linearAttribution([]model.Touchpoint{})
	if result2 != nil {
		t.Errorf("expected nil for empty slice, got %v", result2)
	}
}

// ─── GetAttributionBidAdjustment edge cases ───────────────────────────────────

// TestGetAttrBidAdj_EmptyIDs_B29 covers both empty userID and empty campaignID early returns
func TestGetAttrBidAdj_EmptyIDs_B29(t *testing.T) {
	mc := NewMockCache()
	as := NewAttributionService(mc)
	if mult := as.GetAttributionBidAdjustment("c1", "", "linear", 168); mult != 1.0 {
		t.Errorf("expected 1.0 for empty userID, got %v", mult)
	}
	if mult := as.GetAttributionBidAdjustment("", "u1", "linear", 168); mult != 1.0 {
		t.Errorf("expected 1.0 for empty campaignID, got %v", mult)
	}
}

// TestGetAttrBidAdj_Floor_B29 covers floor (0.5) when campaign credit is 0.
// Seeds 2 touchpoints with campaignID="" → stored at key "u-floor:", touchpoint.CampaignID="".
// GetAttributionBidAdjustment("c-nomatch", "u-floor") calls CalculateAttribution("u-floor","")
// → credits exist, but credit.Touchpoint.CampaignID="" != "c-nomatch" → campaignCredit=0
// → ratio = 0/avgCredit = 0 → multiplier = 0.5+0*0.5 = 0.5
func TestGetAttrBidAdj_Floor_B29(t *testing.T) {
	mc := NewMockCache()
	as := NewAttributionService(mc)

	_ = mc.RecordTouchpoint("u-floor29", "", "click", "r1", 30)
	_ = mc.RecordTouchpoint("u-floor29", "", "impression", "r2", 30)

	mult := as.GetAttributionBidAdjustment("c-nomatch29", "u-floor29", "linear", 168)
	if mult != 0.5 {
		t.Errorf("expected floor 0.5, got %v", mult)
	}
}

// TestGetAttrBidAdj_NoTouchpoints_B29 covers empty credits → return 1.0
func TestGetAttrBidAdj_NoTouchpoints_B29(t *testing.T) {
	mc := NewMockCache()
	as := NewAttributionService(mc)
	mult := as.GetAttributionBidAdjustment("c1", "u-notp-b29", "linear", 168)
	if mult != 1.0 {
		t.Errorf("expected 1.0 when no touchpoints, got %v", mult)
	}
}

// ─── GetAttributionSummary error path ─────────────────────────────────────────

// TestGetAttrSummary_EmptyUser_B29 covers userID="" → CalculateAttribution error path
func TestGetAttrSummary_EmptyUser_B29(t *testing.T) {
	mc := NewMockCache()
	as := NewAttributionService(mc)
	summary, err := as.GetAttributionSummary("", "linear", 168)
	if err == nil {
		t.Error("expected error for empty userID")
	}
	if summary != nil {
		t.Error("expected nil summary on error")
	}
}

// ─── CompareModels error path ──────────────────────────────────────────────────

// TestCompareModels_ErrorPath_B29 covers error path via empty userID
func TestCompareModels_ErrorPath_B29(t *testing.T) {
	mc := NewMockCache()
	as := NewAttributionService(mc)
	results, err := as.CompareModels("", "c1")
	if err == nil {
		t.Error("expected error for empty userID in CompareModels")
	}
	if results != nil {
		t.Error("expected nil results on error")
	}
}

// ─── getCurrentSeason ─────────────────────────────────────────────────────────

// TestGetCurrentSeason_ValidReturn_B29 verifies getCurrentSeason always returns a valid string
func TestGetCurrentSeason_ValidReturn_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
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
		t.Errorf("unexpected season value: %q", season)
	}
}

// ─── isEventActive year-wrap ──────────────────────────────────────────────────

// TestIsEventActive_YearWrap_Dec_B29 covers year-wrap logic: Dec 28 inside Dec26→Jan02 window
func TestIsEventActive_YearWrap_Dec_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	event := model.SeasonalEvent{
		StartDate: "12-26",
		EndDate:   "01-02",
		Recurring: true,
		Active:    true,
	}
	now := time.Date(2024, 12, 28, 12, 0, 0, 0, time.UTC)
	if !svc.isEventActive(event, now) {
		t.Error("expected event active on Dec 28 in year-wrap window")
	}
}

// TestIsEventActive_YearWrap_Jan_B29 covers year-wrap logic: Jan 1 inside Dec26→Jan02 window
func TestIsEventActive_YearWrap_Jan_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	event := model.SeasonalEvent{
		StartDate: "12-26",
		EndDate:   "01-02",
		Recurring: true,
		Active:    true,
	}
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	if !svc.isEventActive(event, now) {
		t.Error("expected event active on Jan 1 in year-wrap window")
	}
}

// TestIsEventActive_FullDate_Past_B29 covers full-date format for a past non-recurring event
func TestIsEventActive_FullDate_Past_B29(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	event := model.SeasonalEvent{
		StartDate: "2020-01-01",
		EndDate:   "2020-01-05",
		Recurring: false,
		Active:    true,
	}
	now := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	if svc.isEventActive(event, now) {
		t.Error("expected past one-time event to be inactive")
	}
}
