package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================================
// calculateGoalPacingMultiplier — cover past-deadline, goal-met, slightly-behind
// ============================================================================

func TestGoalPacing_PastDeadline_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.GoalTarget = 1000
	camp.GoalDelivered = 500
	camp.GoalEndDate = "2020-01-01" // Past date
	result := s.calculateGoalPacingMultiplier(camp)
	if result != 0.5 {
		t.Errorf("expected 0.5 for past deadline, got %f", result)
	}
}

func TestGoalPacing_GoalMet_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.GoalTarget = 1000
	camp.GoalDelivered = 1000 // already met
	camp.GoalEndDate = time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	result := s.calculateGoalPacingMultiplier(camp)
	if result != 0.3 {
		t.Errorf("expected 0.3 for goal already met, got %f", result)
	}
}

func TestGoalPacing_InvalidDate_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.GoalTarget = 1000
	camp.GoalDelivered = 0
	camp.GoalEndDate = "not-a-date"
	result := s.calculateGoalPacingMultiplier(camp)
	if result != 1.0 {
		t.Errorf("expected 1.0 for invalid date, got %f", result)
	}
}

func TestGoalPacing_NoGoal_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.GoalTarget = 0
	result := s.calculateGoalPacingMultiplier(camp)
	if result != 1.0 {
		t.Errorf("expected 1.0 when no goal, got %f", result)
	}
}

func TestGoalPacing_SlightlyBehind_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.GoalTarget = 10000
	camp.GoalDelivered = 1 // very behind
	camp.GoalEndDate = time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	result := s.calculateGoalPacingMultiplier(camp)
	// deliveryRatio will be very small → 1.5 or 1.2
	if result < 1.1 {
		t.Errorf("expected pacing boost for behind-goal, got %f", result)
	}
}

// ============================================================================
// optimizeForDCPM — win rate < 0.1 (underbidding) and > 0.4 (overpaying) branches
// ============================================================================

func TestOptimizeForDCPM_LowWinRate_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	pg := &model.PerformanceGoals{TargetDCPM: 5.0}
	perf := newPerfData()
	perf.winRate = 0.05 // < 0.1 → should apply 1.3x boost
	result := s.optimizeForDCPM(camp, newReq(), pg, perf)
	if result <= 0 {
		t.Errorf("expected positive result, got %f", result)
	}
}

func TestOptimizeForDCPM_HighWinRate_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	pg := &model.PerformanceGoals{TargetDCPM: 5.0}
	perf := newPerfData()
	perf.winRate = 0.50 // > 0.4 → should apply 0.8x reduction
	result := s.optimizeForDCPM(camp, newReq(), pg, perf)
	if result <= 0 {
		t.Errorf("expected positive result, got %f", result)
	}
}

func TestOptimizeForDCPM_ZeroTarget_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetDCPM: 0}
	result := s.optimizeForDCPM(newCampaign(1.0), newReq(), pg, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for zero target, got %f", result)
	}
}

func TestOptimizeForDCPM_ZeroBidPrice_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetDCPM: 5.0}
	result := s.optimizeForDCPM(newCampaign(0), newReq(), pg, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for zero bid price, got %f", result)
	}
}

// ============================================================================
// predictCPV — autoplay branch
// ============================================================================

func TestPredictCPV_Autoplay_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	perf := newPerfData()
	perf.viewability = 0.5
	req := newReq()
	req.Context = map[string]interface{}{
		"autoplay": true,
	}
	result := s.predictCPV(camp, req, perf)
	if result <= 0 {
		t.Errorf("expected positive CPV, got %f", result)
	}
}

func TestPredictCPV_PredictedViewRate_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(2.0)
	perf := newPerfData()
	req := newReq()
	req.Context = map[string]interface{}{
		"predicted_view_rate": float64(0.8),
	}
	result := s.predictCPV(camp, req, perf)
	expected := 2.0 / 0.8
	if result < expected*0.9 || result > expected*1.1 {
		t.Errorf("expected CPV ~%f, got %f", expected, result)
	}
}

// ============================================================================
// optimizeForCPR — cap/floor branches
// ============================================================================

func TestOptimizeForCPR_CapAt2_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(0.001) // very low bid → high ratio
	pg := &model.PerformanceGoals{TargetCPR: 100.0}
	result := s.optimizeForCPR(camp, newReq(), pg, newPerfData())
	if result > 2.01 {
		t.Errorf("expected cap at 2.0, got %f", result)
	}
}

func TestOptimizeForCPR_FloorAt0_3_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(100.0) // very high bid → low ratio
	pg := &model.PerformanceGoals{TargetCPR: 0.001}
	result := s.optimizeForCPR(camp, newReq(), pg, newPerfData())
	if result < 0.29 {
		t.Errorf("expected floor at 0.3, got %f", result)
	}
}

// ============================================================================
// optimizeForCPA — cap/floor branches
// ============================================================================

func TestOptimizeForCPA_Cap_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(0.001) // very low bid → high ratio
	pg := &model.PerformanceGoals{TargetCPA: 100.0}
	result := s.optimizeForCPA(camp, newReq(), pg, newPerfData())
	if result > 2.01 {
		t.Errorf("expected cap at 2.0, got %f", result)
	}
}

// ============================================================================
// optimizeForCPCV — non-skippable and CTV completion branches
// ============================================================================

func TestOptimizeForCPCV_NonSkippable_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	pg := &model.PerformanceGoals{
		PrimaryGoal: "cpcv",
		TargetCPCV:  0.05,
	}
	camp.Targeting.PerformanceGoals = pg
	req := newReq()
	req.Context = map[string]interface{}{
		"skippable": false,
	}
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
}

func TestOptimizeForCPCV_CTVCompletion_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	pg := &model.PerformanceGoals{
		PrimaryGoal: "cpcv",
		TargetCPCV:  0.05,
	}
	camp.Targeting.PerformanceGoals = pg
	req := newCTVReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
}

// ============================================================================
// optimizeForCPS — e-commerce goals branch
// ============================================================================

func TestPerfGoal_CPS_WithEcommerceGoals_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cps",
		TargetCPS:   5.0,
		EcommerceGoals: &model.EcommerceOptimization{
			TargetCostPerSale: 5.0,
		},
	}
	req := newReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
}

// ============================================================================
// optimizeForCPI — app goals branch
// ============================================================================

func TestPerfGoal_CPI_WithAppGoals_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpi",
		TargetCPI:   2.0,
		AppGoals: &model.AppOptimization{
			TargetCostPerInstall: 2.0,
		},
	}
	req := newReq()
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
}

// ============================================================================
// calculatePerformanceGoalMultiplier — maximize_conversions high CVR branch,
// target_cpa ratio > 1.2 and < 0.8 branches
// ============================================================================

func TestApplyBidStrategy_MaximizeConversions_HighCVR_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{BidStrategy: "maximize_conversions"}
	perf := newPerfData()
	perf.cvr = 0.05 // > 0.03 → 1.4 boost
	result := s.applyBidStrategy(pg, perf)
	if result != 1.4 {
		t.Errorf("expected 1.4 for high-CVR maximize_conversions, got %f", result)
	}
}

func TestApplyBidStrategy_MaximizeConversions_LowCVR_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{BidStrategy: "maximize_conversions"}
	perf := newPerfData()
	perf.cvr = 0.01 // < 0.03 → 1.0
	result := s.applyBidStrategy(pg, perf)
	if result != 1.0 {
		t.Errorf("expected 1.0 for low-CVR maximize_conversions, got %f", result)
	}
}

func TestApplyBidStrategy_TargetCPA_HighRatio_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{
		BidStrategy: "target_cpa",
		TargetCPA:   10.0,
	}
	perf := newPerfData()
	perf.cpa = 5.0 // ratio = 10/5 = 2.0 > 1.2 → capped at 1.2
	result := s.applyBidStrategy(pg, perf)
	if result != 1.2 {
		t.Errorf("expected 1.2 for high ratio target_cpa, got %f", result)
	}
}

func TestApplyBidStrategy_TargetCPA_LowRatio_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{
		BidStrategy: "target_cpa",
		TargetCPA:   5.0,
	}
	perf := newPerfData()
	perf.cpa = 10.0 // ratio = 5/10 = 0.5 < 0.8 → capped at 0.8
	result := s.applyBidStrategy(pg, perf)
	if result != 0.8 {
		t.Errorf("expected 0.8 for low ratio target_cpa, got %f", result)
	}
}

func TestApplyBidStrategy_MaximizeClicks_HighCTR_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{BidStrategy: "maximize_clicks"}
	perf := newPerfData()
	perf.ctr = 0.02 // > 0.015 → 1.3
	result := s.applyBidStrategy(pg, perf)
	if result != 1.3 {
		t.Errorf("expected 1.3 for high-CTR maximize_clicks, got %f", result)
	}
}

func TestApplyBidStrategy_MaximizeClicks_LowCTR_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{BidStrategy: "maximize_clicks"}
	perf := newPerfData()
	perf.ctr = 0.005 // < 0.015 → 1.0
	result := s.applyBidStrategy(pg, perf)
	if result != 1.0 {
		t.Errorf("expected 1.0 for low-CTR maximize_clicks, got %f", result)
	}
}

// ============================================================================
// determineOptimizationLevel — aggressive (low CPA vs target)
// ============================================================================

func TestDetermineOptimizationLevel_Aggressive_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPA: 10.0}
	perf := newPerfData()
	perf.cpa = 5.0          // < 10 * 0.8 = 8.0 → aggressive
	perf.impressions = 2000 // > 1000
	result := s.determineOptimizationLevel(pg, perf)
	if result != "aggressive" {
		t.Errorf("expected aggressive, got %s", result)
	}
}

func TestDetermineOptimizationLevel_LowImpressions_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPA: 10.0}
	perf := newPerfData()
	perf.impressions = 500 // < 1000 → conservative
	result := s.determineOptimizationLevel(pg, perf)
	if result != "conservative" {
		t.Errorf("expected conservative for low impressions, got %s", result)
	}
}

// ============================================================================
// RefreshCampaigns — non-dev env error, success path, non-200 status
// ============================================================================

func TestRefreshCampaigns_Success_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	// Serve a valid campaign JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `[{"id":"c1","bidPrice":1.0,"status":"active","budget":100}]`)
	}))
	defer server.Close()
	err := s.RefreshCampaigns(server.URL)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestRefreshCampaigns_Non200_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	err := s.RefreshCampaigns(server.URL)
	if err == nil {
		t.Errorf("expected error for non-200 status")
	}
}

func TestRefreshCampaigns_BadJSON_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{not valid json}`)
	}))
	defer server.Close()
	err := s.RefreshCampaigns(server.URL)
	if err == nil {
		t.Errorf("expected error for bad JSON")
	}
}

func TestRefreshCampaigns_NetworkError_B15(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	// Use an unreachable address to trigger network error
	err := s.RefreshCampaigns("http://127.0.0.1:1")
	if err == nil {
		t.Errorf("expected network error")
	}
}

// ============================================================================
// checkBrandSafety — high fraud count (>10 standard and >5) via cache seeding
// ============================================================================

func TestCheckBrandSafety_HighFraudCount_Standard_B15(t *testing.T) {
	mc := NewMockCache()
	s := NewBiddingService(mc, "")
	// Seed fraud count > 10
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("fraud:publisher:%s:%s:count", "pub-fraudy", today)
	_ = mc.Set(key, "15", 3600)
	camp := newCampaign(1.0)
	camp.BrandSafetyLevel = "standard"
	req := newReq()
	req.PublisherID = "pub-fraudy"
	result := s.checkBrandSafety(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked in standard for high fraud, reason: %s", result.Reason)
	}
	if result.Multiplier >= 1.0 {
		t.Errorf("expected reduced multiplier for high fraud, got %f", result.Multiplier)
	}
}

func TestCheckBrandSafety_HighFraudCount_Strict_B15(t *testing.T) {
	mc := NewMockCache()
	s := NewBiddingService(mc, "")
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("fraud:publisher:%s:%s:count", "pub-fraudy2", today)
	_ = mc.Set(key, "15", 3600)
	camp := newCampaign(1.0)
	camp.BrandSafetyLevel = "strict"
	req := newReq()
	req.PublisherID = "pub-fraudy2"
	result := s.checkBrandSafety(camp, req)
	if !result.Blocked {
		t.Errorf("expected blocked in strict mode for high fraud count")
	}
}

func TestCheckBrandSafety_MediumFraudCount_B15(t *testing.T) {
	mc := NewMockCache()
	s := NewBiddingService(mc, "")
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("fraud:publisher:%s:%s:count", "pub-mediocre", today)
	_ = mc.Set(key, "7", 3600)
	camp := newCampaign(1.0)
	req := newReq()
	req.PublisherID = "pub-mediocre"
	result := s.checkBrandSafety(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for medium fraud count, reason: %s", result.Reason)
	}
	if result.Multiplier >= 1.0 {
		t.Errorf("expected reduced multiplier for medium fraud, got %f", result.Multiplier)
	}
}
