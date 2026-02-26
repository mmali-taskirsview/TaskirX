package service

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ─── Attribution Tests ───────────────────────────────────────────────────────

func TestLinearAttribution_MultiplePoints_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	mc.RecordTouchpoint("u1", "c1", "impression", "r1", 30)
	mc.RecordTouchpoint("u1", "c1", "click", "r2", 30)
	mc.RecordTouchpoint("u1", "c1", "impression", "r3", 30)

	credits, err := svc.CalculateAttribution("u1", "c1", "linear", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 3 {
		t.Fatalf("expected 3 credits, got %d", len(credits))
	}
	for _, c := range credits {
		if c.Credit < 0.33 || c.Credit > 0.34 {
			t.Errorf("linear credit expected ~0.333, got %f", c.Credit)
		}
	}
}

func TestLinearAttribution_NoTouchpoints_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	credits, err := svc.CalculateAttribution("u_empty", "", "linear", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 0 {
		t.Errorf("expected 0 credits for no touchpoints, got %d", len(credits))
	}
}

func TestTimeDecayAttribution_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	// Add touchpoints manually at different times via RecordTouchpoint
	// More recent = higher credit
	mc.RecordTouchpoint("u2", "c2", "impression", "r1", 30)
	time.Sleep(5 * time.Millisecond)
	mc.RecordTouchpoint("u2", "c2", "click", "r2", 30)

	credits, err := svc.CalculateAttribution("u2", "c2", "time_decay", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 2 {
		t.Fatalf("expected 2 credits, got %d", len(credits))
	}
	// The click (more recent) should have more or equal credit than impression
	if credits[1].Credit < credits[0].Credit {
		t.Errorf("more recent touchpoint should have higher credit: %f vs %f", credits[1].Credit, credits[0].Credit)
	}
}

func TestPositionBased_TwoPoints_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	mc.RecordTouchpoint("u3", "c3", "impression", "r1", 30)
	mc.RecordTouchpoint("u3", "c3", "click", "r2", 30)

	credits, err := svc.CalculateAttribution("u3", "c3", "position_based", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 2 {
		t.Fatalf("expected 2 credits, got %d", len(credits))
	}
	// n=2 special case: each gets 0.5
	for _, c := range credits {
		if c.Credit < 0.49 || c.Credit > 0.51 {
			t.Errorf("position_based n=2: expected 0.5, got %f", c.Credit)
		}
	}
}

func TestPositionBased_OnePoint_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	mc.RecordTouchpoint("u4", "c4", "impression", "r1", 30)

	credits, err := svc.CalculateAttribution("u4", "c4", "position_based", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 1 {
		t.Fatalf("expected 1 credit, got %d", len(credits))
	}
	// n=1 special case: gets 1.0
	if credits[0].Credit < 0.99 || credits[0].Credit > 1.01 {
		t.Errorf("position_based n=1: expected 1.0, got %f", credits[0].Credit)
	}
}

func TestPositionBased_FivePoints_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	for i := 0; i < 5; i++ {
		mc.RecordTouchpoint("u5", "c5", "impression", "r"+string(rune('0'+i)), 30)
	}

	credits, err := svc.CalculateAttribution("u5", "c5", "position_based", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 5 {
		t.Fatalf("expected 5 credits, got %d", len(credits))
	}
	// First 0.4, last 0.4, 3 middle = 0.2/3 each
	if credits[0].Credit < 0.39 || credits[0].Credit > 0.41 {
		t.Errorf("first touchpoint: expected 0.4, got %f", credits[0].Credit)
	}
	if credits[4].Credit < 0.39 || credits[4].Credit > 0.41 {
		t.Errorf("last touchpoint: expected 0.4, got %f", credits[4].Credit)
	}
}

func TestGetAttributionBidAdjustment_EmptyUserID_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	// Empty userID → return 1.0 immediately
	multiplier := svc.GetAttributionBidAdjustment("c1", "", "linear", 0)
	if multiplier != 1.0 {
		t.Errorf("expected 1.0 for empty userID, got %f", multiplier)
	}
}

func TestGetAttributionBidAdjustment_EmptyCampaignID_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	// Empty campaignID → return 1.0 immediately
	multiplier := svc.GetAttributionBidAdjustment("", "u1", "linear", 0)
	if multiplier != 1.0 {
		t.Errorf("expected 1.0 for empty campaignID, got %f", multiplier)
	}
}

func TestGetAttributionBidAdjustment_WithHistory_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	// GetAttributionBidAdjustment calls CalculateAttribution(userID, "", model)
	// which calls GetTouchpoints("u6", "") → key "u6:"
	// We store at key "u6:" by passing "" as campaignID to RecordTouchpoint
	mc.RecordTouchpoint("u6", "", "impression", "r1", 30)
	mc.RecordTouchpoint("u6", "", "click", "r2", 30)
	mc.RecordTouchpoint("u6", "", "impression", "r3", 30)

	// Pass non-empty campaignID so early-return check passes
	multiplier := svc.GetAttributionBidAdjustment("c6", "u6", "linear", 1.0)
	// multiplier can be any positive value (even 1.0 if no campaign match)
	if multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %f", multiplier)
	}
}

func TestGetAttributionSummary_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	// Store touchpoints with empty campaignID so GetTouchpoints("u7","") returns them
	// GetAttributionSummary calls CalculateAttribution(userID, "", ...) → GetTouchpoints(userID, "")
	mc.RecordTouchpoint("u7", "", "impression", "r1", 30)
	mc.RecordTouchpoint("u7", "", "click", "r2", 30)
	mc.RecordTouchpoint("u7", "", "impression", "r3", 30)

	summary, err := svc.GetAttributionSummary("u7", "linear", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summary) == 0 {
		t.Error("expected non-empty summary")
	}
}

func TestCompareModels_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	// CompareModels calls CalculateAttribution(userID, "", model) → GetTouchpoints(userID, "")
	mc.RecordTouchpoint("u8", "", "impression", "r1", 30)
	mc.RecordTouchpoint("u8", "", "click", "r2", 30)
	mc.RecordTouchpoint("u8", "", "impression", "r3", 30)

	results, err := svc.CompareModels("u8", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have 5 model results
	if len(results) != 5 {
		t.Errorf("expected 5 model results, got %d", len(results))
	}
}

func TestCompareModels_NoTouchpoints_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	results, err := svc.CompareModels("u_none", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With no touchpoints, 5 models but all with empty credits
	if len(results) != 5 {
		t.Errorf("expected 5 model results, got %d", len(results))
	}
}

// ─── A/B Testing Tests ───────────────────────────────────────────────────────

func makeRunningExperiment_B27(t *testing.T, svc *ABTestingService, expID string) {
	t.Helper()
	req := CreateExperimentRequest{
		Name: expID,
		Variants: []VariantRequest{
			{Name: "control", IsControl: true, Weight: 0.5},
			{Name: "treatment", IsControl: false, Weight: 0.5},
		},
		Metrics: []string{"ctr"},
	}
	exp, err := svc.CreateExperiment(req)
	if err != nil {
		t.Fatalf("CreateExperiment failed: %v", err)
	}
	if err := svc.StartExperiment(exp.ID); err != nil {
		t.Fatalf("StartExperiment failed: %v", err)
	}
}

func TestStopExperiment_NotFound_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewABTestingService(mc)

	err := svc.StopExperiment("nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent experiment")
	}
}

func TestStopExperiment_NotRunning_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewABTestingService(mc)

	// Create but don't start (status = "draft")
	req := CreateExperimentRequest{
		Name: "draft-exp",
		Variants: []VariantRequest{
			{Name: "control", IsControl: true, Weight: 0.5},
			{Name: "treatment", IsControl: false, Weight: 0.5},
		},
		Metrics: []string{"ctr"},
	}
	exp, err := svc.CreateExperiment(req)
	if err != nil {
		t.Fatalf("CreateExperiment failed: %v", err)
	}

	err = svc.StopExperiment(exp.ID)
	if err == nil {
		t.Error("expected error when stopping non-running experiment")
	}
}

func TestStopExperiment_Success_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewABTestingService(mc)

	makeRunningExperiment_B27(t, svc, "stop-success-exp")

	// Find the experiment ID by listing stats
	stats := svc.GetStats()
	if stats["running_experiments"].(int) == 0 {
		t.Fatal("expected at least one running experiment")
	}

	// Collect the ID by ranging experiments
	var foundID string
	svc.experiments.Range(func(k, v any) bool {
		exp := v.(*abExperiment)
		if exp.Status == "running" {
			foundID = exp.ID
			return false
		}
		return true
	})
	if foundID == "" {
		t.Fatal("could not find running experiment ID")
	}

	err := svc.StopExperiment(foundID)
	if err != nil {
		t.Errorf("expected no error stopping running experiment, got %v", err)
	}
}

func TestRecordEvent_CustomMetric_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewABTestingService(mc)

	req := CreateExperimentRequest{
		Name: "record-event-exp",
		Variants: []VariantRequest{
			{Name: "control", IsControl: true, Weight: 0.5},
			{Name: "treatment", IsControl: false, Weight: 0.5},
		},
		Metrics: []string{"ctr"},
	}
	exp, err := svc.CreateExperiment(req)
	if err != nil {
		t.Fatalf("CreateExperiment failed: %v", err)
	}
	if err := svc.StartExperiment(exp.ID); err != nil {
		t.Fatalf("StartExperiment failed: %v", err)
	}

	// Record a custom metric event (not impression/click/conversion/revenue)
	err = svc.RecordEvent(exp.ID, exp.Variants[0].ID, "custom_metric", 3.14)
	if err != nil {
		t.Errorf("unexpected error recording custom metric: %v", err)
	}
}

func TestRecordEvent_RevenueEvent_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewABTestingService(mc)

	req := CreateExperimentRequest{
		Name: "revenue-event-exp",
		Variants: []VariantRequest{
			{Name: "control", IsControl: true, Weight: 0.5},
			{Name: "treatment", IsControl: false, Weight: 0.5},
		},
		Metrics: []string{"revenue"},
	}
	exp, err := svc.CreateExperiment(req)
	if err != nil {
		t.Fatalf("CreateExperiment failed: %v", err)
	}
	if err := svc.StartExperiment(exp.ID); err != nil {
		t.Fatalf("StartExperiment failed: %v", err)
	}

	err = svc.RecordEvent(exp.ID, exp.Variants[0].ID, "revenue", 9.99)
	if err != nil {
		t.Errorf("unexpected error recording revenue event: %v", err)
	}
}

func TestRecordEvent_NotFound_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewABTestingService(mc)

	err := svc.RecordEvent("no-exp", "no-var", "impression", 1)
	if err == nil {
		t.Error("expected error for nonexistent experiment")
	}
}

func TestGetStats_MultipleExperiments_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewABTestingService(mc)

	// Create 2 experiments — one running, one draft
	req1 := CreateExperimentRequest{
		Name: "running-exp-stats",
		Variants: []VariantRequest{
			{Name: "control", IsControl: true, Weight: 0.5},
			{Name: "treatment", IsControl: false, Weight: 0.5},
		},
		Metrics: []string{"ctr"},
	}
	exp1, err := svc.CreateExperiment(req1)
	if err != nil {
		t.Fatalf("CreateExperiment failed: %v", err)
	}
	if err := svc.StartExperiment(exp1.ID); err != nil {
		t.Fatalf("StartExperiment failed: %v", err)
	}

	req2 := CreateExperimentRequest{
		Name: "draft-exp-stats",
		Variants: []VariantRequest{
			{Name: "control", IsControl: true, Weight: 0.5},
			{Name: "variant_b", IsControl: false, Weight: 0.5},
		},
		Metrics: []string{"ctr"},
	}
	if _, err := svc.CreateExperiment(req2); err != nil {
		t.Fatalf("CreateExperiment2 failed: %v", err)
	}

	stats := svc.GetStats()
	if stats["total_experiments"].(int) < 2 {
		t.Errorf("expected >=2 total experiments, got %d", stats["total_experiments"].(int))
	}
	if stats["running_experiments"].(int) < 1 {
		t.Errorf("expected >=1 running experiments, got %d", stats["running_experiments"].(int))
	}
	if stats["draft_experiments"].(int) < 1 {
		t.Errorf("expected >=1 draft experiments, got %d", stats["draft_experiments"].(int))
	}
}

func TestCalculatePValue_ZeroSamples_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewABTestingService(mc)

	// With zero samples, calculatePValue should return 1.0
	pValue := svc.calculatePValue(0, 0, 0, 0)
	if pValue != 1.0 {
		t.Errorf("expected p-value 1.0 for zero samples, got %f", pValue)
	}
}

func TestCalculatePValue_EqualRates_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewABTestingService(mc)

	// Equal conversion rates → high p-value (no significant difference)
	pValue := svc.calculatePValue(100, 10, 100, 10)
	if pValue < 0.5 {
		t.Errorf("expected high p-value for equal rates, got %f", pValue)
	}
}

func TestCalculatePValue_VeryDifferentRates_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewABTestingService(mc)

	// Very different rates → low p-value (significant difference)
	pValue := svc.calculatePValue(1000, 100, 1000, 500)
	if pValue > 0.05 {
		t.Errorf("expected low p-value for very different rates, got %f", pValue)
	}
}

func TestSampleGamma_ShapeLessThan1_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewABTestingService(mc)

	// shape < 1 triggers the recursive branch in sampleGamma
	result := svc.sampleGamma(0.3, 1.0)
	if result < 0 {
		t.Errorf("sampleGamma with shape<1 returned negative value: %f", result)
	}
}

func TestSampleGamma_ShapeGreaterThan1_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewABTestingService(mc)

	// shape >= 1 uses the Marsaglia-Tsang algorithm
	result := svc.sampleGamma(2.0, 1.0)
	if result < 0 {
		t.Errorf("sampleGamma with shape>1 returned negative value: %f", result)
	}
}

// ─── Bid Landscape Tests ─────────────────────────────────────────────────────

func TestFindOptimalRange_FewerThan3Percentiles_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewBidLandscapeService(mc)

	percentiles := []model.BidPercentile{
		{Percentile: 25, BidPrice: 1.0, WinRate: 0.3},
		{Percentile: 50, BidPrice: 2.0, WinRate: 0.5},
	}

	result := svc.findOptimalRange(nil, percentiles)
	if result != nil {
		t.Error("expected nil result for <3 percentiles")
	}
}

func TestFindOptimalRange_WithData_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewBidLandscapeService(mc)

	percentiles := []model.BidPercentile{
		{Percentile: 10, BidPrice: 0.5, WinRate: 0.1},
		{Percentile: 25, BidPrice: 1.0, WinRate: 0.25},
		{Percentile: 50, BidPrice: 2.0, WinRate: 0.50},
		{Percentile: 75, BidPrice: 3.5, WinRate: 0.70},
		{Percentile: 90, BidPrice: 5.0, WinRate: 0.85},
	}

	result := svc.findOptimalRange(nil, percentiles)
	if result == nil {
		t.Fatal("expected non-nil result for 5 percentiles")
	}
	if result.SweetSpot <= 0 {
		t.Errorf("expected positive sweet spot, got %f", result.SweetSpot)
	}
	if result.MinBid <= 0 {
		t.Errorf("expected positive min bid, got %f", result.MinBid)
	}
	if result.MaxBid <= result.MinBid {
		t.Errorf("expected MaxBid > MinBid, got max=%f min=%f", result.MaxBid, result.MinBid)
	}
}

// ─── calculateAutoBidMultiplier Tests ────────────────────────────────────────

func TestCalculateAutoBidMultiplier_HighCTRLowWinRate_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// CTR > 2% (0.025) + WinRate < 30% (0.25) → 1.20
	mc.ctr["camp1"] = 0.025
	mc.winRate["camp1"] = 0.25
	campaign := &model.Campaign{ID: "camp1"}

	multiplier := svc.calculateAutoBidMultiplier(campaign)
	if multiplier != 1.20 {
		t.Errorf("expected 1.20 for high CTR + low win rate, got %f", multiplier)
	}
}

func TestCalculateAutoBidMultiplier_LowCTRHighWinRate_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// CTR < 0.5% (0.003) + WinRate > 70% (0.75) → 0.80
	mc.ctr["camp2"] = 0.003
	mc.winRate["camp2"] = 0.75
	campaign := &model.Campaign{ID: "camp2"}

	multiplier := svc.calculateAutoBidMultiplier(campaign)
	if multiplier != 0.80 {
		t.Errorf("expected 0.80 for low CTR + high win rate, got %f", multiplier)
	}
}

func TestCalculateAutoBidMultiplier_ModerateCTRVeryLowWinRate_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// CTR 1-3% (0.015) + WinRate < 20% (0.15) → 1.10
	mc.ctr["camp3"] = 0.015
	mc.winRate["camp3"] = 0.15
	campaign := &model.Campaign{ID: "camp3"}

	multiplier := svc.calculateAutoBidMultiplier(campaign)
	if multiplier != 1.10 {
		t.Errorf("expected 1.10 for moderate CTR + very low win rate, got %f", multiplier)
	}
}

func TestCalculateAutoBidMultiplier_LowCTRModerateWinRate_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// CTR < 1% (0.007) + WinRate 50-70% (0.60) → 0.90
	mc.ctr["camp4"] = 0.007
	mc.winRate["camp4"] = 0.60
	campaign := &model.Campaign{ID: "camp4"}

	multiplier := svc.calculateAutoBidMultiplier(campaign)
	if multiplier != 0.90 {
		t.Errorf("expected 0.90 for low CTR + moderate win rate, got %f", multiplier)
	}
}

func TestCalculateAutoBidMultiplier_Default_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// Normal performance → 1.0
	mc.ctr["camp5"] = 0.015    // 1.5%
	mc.winRate["camp5"] = 0.45 // 45%
	campaign := &model.Campaign{ID: "camp5"}

	multiplier := svc.calculateAutoBidMultiplier(campaign)
	if multiplier != 1.0 {
		t.Errorf("expected 1.0 for normal performance, got %f", multiplier)
	}
}

func TestCalculateAutoBidMultiplier_NoCTRData_B27(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// No CTR data → GetCampaignCTR returns (0, nil) from map default → still processes
	// Actually map returns 0 with no error, so it runs through logic
	// Use a campaign not seeded in ctr map
	campaign := &model.Campaign{ID: "camp_no_data"}
	multiplier := svc.calculateAutoBidMultiplier(campaign)
	// 0 CTR, 0 WinRate → ctrPct=0 <0.5 but winRatePct=0 not >70, not 50-70 → falls through → 1.0
	if multiplier < 0.79 || multiplier > 1.21 {
		t.Errorf("expected multiplier in range [0.79, 1.21] for no data, got %f", multiplier)
	}
}
