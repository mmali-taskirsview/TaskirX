package service

// coverage_boost18_test.go targets sub-85% functions:
//   - callFraudService: cache "block" and "allow" hit branches
//   - callOptimizationService: circuit-breaker-open branch
//   - callAIMatchingService: circuit-breaker-open branch
//   - programmatic_guaranteed: UpdateDeal not-found, matchesInventorySpec SiteID/Placement filters
//   - creative_optimization: RecordConversion/RecordEngagement placement-exists branch
//   - performance_prediction: calculateMetricConfidence all metric branches, updateFeatureStats new+existing keys
//   - optimizeForVCPM: low-viewability penalty, high-viewability boost
//   - predictCPE: mobile+rich_media branches

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ---------------------------------------------------------------------------
// callFraudService: cache "block" hit → isFraud=true
// ---------------------------------------------------------------------------

func TestCallFraudService_CacheBlock_B18(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	req := newReq()
	req.Device.IP = "1.2.3.4"

	// Pre-populate kv cache with "block" for this IP
	mc.kv["ip_rep:1.2.3.4"] = "block"

	isFraud, err, hop := svc.callFraudService(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isFraud {
		t.Error("expected isFraud=true from cache block hit")
	}
	if hop == nil {
		t.Fatal("expected a hop record")
	}
	if hop.ServiceType != "cache" {
		t.Errorf("expected cache hop, got %q", hop.ServiceType)
	}
}

// ---------------------------------------------------------------------------
// callFraudService: cache "allow" hit → isFraud=false
// ---------------------------------------------------------------------------

func TestCallFraudService_CacheAllow_B18(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	req := newReq()
	req.Device.IP = "5.6.7.8"

	// Pre-populate kv cache with "allow" for this IP
	mc.kv["ip_rep:5.6.7.8"] = "allow"

	isFraud, err, hop := svc.callFraudService(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isFraud {
		t.Error("expected isFraud=false from cache allow hit")
	}
	if hop == nil {
		t.Fatal("expected a hop record")
	}
	if hop.ServiceType != "cache" {
		t.Errorf("expected cache hop, got %q", hop.ServiceType)
	}
}

// ---------------------------------------------------------------------------
// callOptimizationService: circuit-breaker open branch
// ---------------------------------------------------------------------------

func TestCallOptimizationService_CircuitOpen_B18(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// Force circuit breaker open: >5 failures, last failure recent
	svc.optFailureCount = 6
	svc.optLastFailure = time.Now()

	camp := newCampaign(2.0)
	bid := &model.BidResult{Campaign: camp}
	req := newReq()

	rec, err, hop := svc.callOptimizationService(bid, req)

	if err == nil {
		t.Error("expected circuit breaker error")
	}
	if rec != nil {
		t.Error("expected nil recommendation when circuit open")
	}
	if hop == nil {
		t.Fatal("expected a hop record")
	}
	if hop.ErrorMessage != "circuit breaker active" {
		t.Errorf("expected circuit breaker error message, got %q", hop.ErrorMessage)
	}
}

// ---------------------------------------------------------------------------
// callAIMatchingService: circuit-breaker open branch
// ---------------------------------------------------------------------------

func TestCallAIMatchingService_CircuitOpen_B18(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// Force AI circuit breaker open: >=3 failures, last failure recent
	svc.aiFailureCount = 3
	svc.aiLastFailure = time.Now()

	req := newReq()

	recs, err, hop := svc.callAIMatchingService(req)

	if err == nil {
		t.Error("expected circuit breaker error")
	}
	if recs != nil {
		t.Error("expected nil recs when circuit open")
	}
	if hop == nil {
		t.Fatal("expected a hop record")
	}
	if hop.ErrorMessage != "circuit breaker active" {
		t.Errorf("expected circuit breaker message, got %q", hop.ErrorMessage)
	}
}

// ---------------------------------------------------------------------------
// ProgrammaticGuaranteed: UpdateDeal not-found error branch
// ---------------------------------------------------------------------------

func TestPG_UpdateDeal_NotFound_B18(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{ID: "nonexistent-deal"}
	err := svc.UpdateDeal(deal)

	if err == nil {
		t.Error("expected error for nonexistent deal")
	}
}

// ---------------------------------------------------------------------------
// ProgrammaticGuaranteed: matchesInventorySpec — SiteID filter match
// ---------------------------------------------------------------------------

func TestPG_MatchesInventorySpec_SiteID_Match_B18(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	spec := InventorySpec{SiteIDs: []string{"site-abc"}}
	matched := svc.matchesInventorySpec(spec, "", "site-abc", "", "", "", "")

	if !matched {
		t.Error("expected match for site-abc")
	}
}

// ---------------------------------------------------------------------------
// ProgrammaticGuaranteed: matchesInventorySpec — SiteID filter no match
// ---------------------------------------------------------------------------

func TestPG_MatchesInventorySpec_SiteID_NoMatch_B18(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	spec := InventorySpec{SiteIDs: []string{"site-abc"}}
	matched := svc.matchesInventorySpec(spec, "", "site-xyz", "", "", "", "")

	if matched {
		t.Error("expected no match for wrong siteID")
	}
}

// ---------------------------------------------------------------------------
// ProgrammaticGuaranteed: matchesInventorySpec — Placement filter match
// ---------------------------------------------------------------------------

func TestPG_MatchesInventorySpec_Placement_Match_B18(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	spec := InventorySpec{Placements: []string{"above_fold"}}
	matched := svc.matchesInventorySpec(spec, "", "", "above_fold", "", "", "")

	if !matched {
		t.Error("expected match for above_fold placement")
	}
}

// ---------------------------------------------------------------------------
// ProgrammaticGuaranteed: matchesInventorySpec — Placement filter no match
// ---------------------------------------------------------------------------

func TestPG_MatchesInventorySpec_Placement_NoMatch_B18(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	spec := InventorySpec{Placements: []string{"above_fold"}}
	matched := svc.matchesInventorySpec(spec, "", "", "below_fold", "", "", "")

	if matched {
		t.Error("expected no match for wrong placement")
	}
}

// ---------------------------------------------------------------------------
// CreativeOptimization: RecordConversion — placement EXISTS (inner branch)
// ---------------------------------------------------------------------------

func TestCreativeOpt_RecordConversion_PlacementExists_B18(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	// First record an impression to create the placement entry
	svc.RecordImpression("cr-1", "banner")

	// Now record a conversion — placement "banner" now exists in placementPerf
	svc.RecordConversion("cr-1", "banner")

	// Verify conversion was tracked in placement-level perf
	svc.mu.RLock()
	perf, ok := svc.placementPerf["banner"]["cr-1"]
	svc.mu.RUnlock()

	if !ok {
		t.Fatal("expected creative perf under placement")
	}
	if perf.conversions != 1 {
		t.Errorf("expected 1 conversion, got %d", perf.conversions)
	}
}

// ---------------------------------------------------------------------------
// CreativeOptimization: RecordEngagement — placement EXISTS (inner branch)
// ---------------------------------------------------------------------------

func TestCreativeOpt_RecordEngagement_PlacementExists_B18(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	// Record impression first to seed placementPerf map
	svc.RecordImpression("cr-2", "video")

	// Now record engagement — "video" exists, so inner block executes
	svc.RecordEngagement("cr-2", "video", 15.5)

	svc.mu.RLock()
	perf, ok := svc.placementPerf["video"]["cr-2"]
	svc.mu.RUnlock()

	if !ok {
		t.Fatal("expected creative perf under video placement")
	}
	if perf.engagements != 1 {
		t.Errorf("expected 1 engagement, got %d", perf.engagements)
	}
	if perf.viewTime != 15.5 {
		t.Errorf("expected viewTime=15.5, got %f", perf.viewTime)
	}
}

// ---------------------------------------------------------------------------
// CreativeOptimization: RecordConversion — placement NOT in map (no inner branch)
// ---------------------------------------------------------------------------

func TestCreativeOpt_RecordConversion_PlacementAbsent_B18(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	// No impression recorded first — placement absent, inner if skipped
	svc.RecordConversion("cr-3", "interstitial")

	// Global conversion should still be recorded
	perf := svc.GetCreativePerformance("cr-3")
	if perf == nil {
		t.Fatal("expected performance data")
	}
	if perf.Conversions != 1 {
		t.Errorf("expected 1 conversion, got %d", perf.Conversions)
	}
}

// ---------------------------------------------------------------------------
// PerformancePrediction: calculateMetricConfidence — all metric branches
// ---------------------------------------------------------------------------

func TestPerfPred_CalculateMetricConfidence_CTR_B18(t *testing.T) {
	svc := NewPerformancePredictionService(nil)
	// Use 150 samples (> MinHistoricalSamples=100) → base=0.95, ctr multiplier=0.95
	result := svc.calculateMetricConfidence(150, "ctr")
	expected := 0.95 * 0.95
	if result < expected-0.001 || result > expected+0.001 {
		t.Errorf("expected %.4f, got %.4f", expected, result)
	}
}

func TestPerfPred_CalculateMetricConfidence_CVR_B18(t *testing.T) {
	svc := NewPerformancePredictionService(nil)
	// base=0.95, cvr multiplier=0.90
	result := svc.calculateMetricConfidence(150, "cvr")
	expected := 0.95 * 0.90
	if result < expected-0.001 || result > expected+0.001 {
		t.Errorf("expected %.4f, got %.4f", expected, result)
	}
}

func TestPerfPred_CalculateMetricConfidence_ROAS_B18(t *testing.T) {
	svc := NewPerformancePredictionService(nil)
	// base=0.95, roas multiplier=0.85
	result := svc.calculateMetricConfidence(150, "roas")
	expected := 0.95 * 0.85
	if result < expected-0.001 || result > expected+0.001 {
		t.Errorf("expected %.4f, got %.4f", expected, result)
	}
}

func TestPerfPred_CalculateMetricConfidence_Default_B18(t *testing.T) {
	svc := NewPerformancePredictionService(nil)
	// base=0.95, default multiplier=0.80
	result := svc.calculateMetricConfidence(150, "other_metric")
	expected := 0.95 * 0.80
	if result < expected-0.001 || result > expected+0.001 {
		t.Errorf("expected %.4f, got %.4f", expected, result)
	}
}

// ---------------------------------------------------------------------------
// PerformancePrediction: updateFeatureStats — new key (first entry) + existing key update
// ---------------------------------------------------------------------------

func TestPerfPred_UpdateFeatureStats_NewKey_B18(t *testing.T) {
	svc := NewPerformancePredictionService(nil)

	record := &PerformanceRecord{
		EntityID:   "camp-1",
		EntityType: "campaign",
		Features:   map[string]float64{"bid_price": 2.0, "historical_ctr": 0.05},
	}

	svc.updateFeatureStats(record)

	svc.mu.RLock()
	mean := svc.featureStats.Means["bid_price"]
	minVal := svc.featureStats.Mins["bid_price"]
	maxVal := svc.featureStats.Maxs["bid_price"]
	svc.mu.RUnlock()

	if mean != 2.0 {
		t.Errorf("expected mean=2.0 for new key, got %f", mean)
	}
	if minVal != 2.0 {
		t.Errorf("expected min=2.0, got %f", minVal)
	}
	if maxVal != 2.0 {
		t.Errorf("expected max=2.0, got %f", maxVal)
	}
}

func TestPerfPred_UpdateFeatureStats_ExistingKey_B18(t *testing.T) {
	svc := NewPerformancePredictionService(nil)

	// First record — seeds the stats
	r1 := &PerformanceRecord{
		EntityID:   "camp-1",
		EntityType: "campaign",
		Features:   map[string]float64{"bid_price": 2.0},
	}
	svc.updateFeatureStats(r1)

	// Second record — triggers the update + min/max branch
	r2 := &PerformanceRecord{
		EntityID:   "camp-1",
		EntityType: "campaign",
		Features:   map[string]float64{"bid_price": 5.0}, // new max
	}
	svc.updateFeatureStats(r2)

	svc.mu.RLock()
	maxVal := svc.featureStats.Maxs["bid_price"]
	svc.mu.RUnlock()

	if maxVal != 5.0 {
		t.Errorf("expected max=5.0 after update, got %f", maxVal)
	}
}

func TestPerfPred_UpdateFeatureStats_NewMin_B18(t *testing.T) {
	svc := NewPerformancePredictionService(nil)

	r1 := &PerformanceRecord{
		EntityID:   "camp-1",
		EntityType: "campaign",
		Features:   map[string]float64{"bid_price": 3.0},
	}
	svc.updateFeatureStats(r1)

	r2 := &PerformanceRecord{
		EntityID:   "camp-1",
		EntityType: "campaign",
		Features:   map[string]float64{"bid_price": 1.0}, // new min
	}
	svc.updateFeatureStats(r2)

	svc.mu.RLock()
	minVal := svc.featureStats.Mins["bid_price"]
	svc.mu.RUnlock()

	if minVal != 1.0 {
		t.Errorf("expected min=1.0 after update, got %f", minVal)
	}
}

// ---------------------------------------------------------------------------
// optimizeForVCPM: high viewability → boost; low viewability → penalty
// ---------------------------------------------------------------------------

func TestOptimizeForVCPM_HighViewability_B18(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(1.0)
	pg := &model.PerformanceGoals{
		TargetVCPM: 5.0, // $5 CPM target
	}
	perf := newPerfData()
	perf.viewability = 0.85 // high viewability ≥ 0.8 → 1.3x ratio boost

	req := newReq()

	ratio := svc.optimizeForVCPM(camp, req, pg, perf)

	// effectiveCPM = 5.0 * 0.85 = 4.25; ratio = 4.25/1000 = 0.00425 → boosted by 1.3 → 0.005525
	// Still very low (< 0.2), so returns 0.2
	if ratio < 0.001 {
		t.Errorf("expected non-trivial ratio, got %f", ratio)
	}
}

func TestOptimizeForVCPM_LowViewability_B18(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(1.0)
	pg := &model.PerformanceGoals{
		TargetVCPM: 5000.0, // very high target to produce ratio > 0.2
	}
	perf := newPerfData()
	perf.viewability = 0.3 // low viewability < 0.4 → 0.5x penalty

	req := newReq()

	ratio := svc.optimizeForVCPM(camp, req, pg, perf)

	// effectiveCPM = 5000 * 0.3 = 1500; ratio = 1500/1000 = 1.5 → * 0.5 = 0.75
	if ratio < 0.5 || ratio > 1.0 {
		t.Errorf("expected ratio ~0.75 for low viewability, got %f", ratio)
	}
}

func TestOptimizeForVCPM_ZeroTargetVCPM_B18(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(1.0)
	pg := &model.PerformanceGoals{TargetVCPM: 0}
	req := newReq()

	ratio := svc.optimizeForVCPM(camp, req, pg, newPerfData())
	if ratio != 1.0 {
		t.Errorf("expected 1.0 for zero target VCPM, got %f", ratio)
	}
}

// ---------------------------------------------------------------------------
// predictCPE: mobile device boost, rich_media creative boost
// ---------------------------------------------------------------------------

func TestPredictCPE_MobileRichMedia_B18(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(2.0)
	camp.Creative.Type = "rich_media"

	req := newReq()
	req.Device.Type = "mobile"

	perf := newPerfData()
	perf.engagementRate = 0.04

	result := svc.predictCPE(camp, req, perf)

	// engagementRate=0.04 * 1.2 (mobile) * 1.5 (rich_media) = 0.072
	// CPE = 2.0 / 0.072 ≈ 27.78
	if result <= 0 {
		t.Errorf("expected positive CPE, got %f", result)
	}
	if result > 100 {
		t.Errorf("expected reasonable CPE, got %f", result)
	}
}

func TestPredictCPE_PredictedEngagementRate_B18(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(3.0)

	req := newReq()
	req.Context = map[string]interface{}{
		"predicted_engagement_rate": 0.10,
	}

	perf := newPerfData()

	result := svc.predictCPE(camp, req, perf)

	// CPE = 3.0 / 0.10 = 30.0
	if result < 25.0 || result > 35.0 {
		t.Errorf("expected CPE ~30.0, got %f", result)
	}
}

// ---------------------------------------------------------------------------
// Fraud HTTP server: 200 OK success branch
// ---------------------------------------------------------------------------

func TestCallFraudService_HTTPSuccess_B18(t *testing.T) {
	// Stand up a local HTTP server that returns a valid fraud response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"is_fraud":false,"recommended_action":"allow","confidence":0.9}`))
	}))
	defer server.Close()

	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	svc.fraudServiceURL = server.URL

	req := newReq()
	req.Device.IP = "9.9.9.9"

	isFraud, err, hop := svc.callFraudService(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isFraud {
		t.Error("expected isFraud=false for allow response")
	}
	if hop == nil || !hop.Success {
		t.Error("expected successful hop")
	}
}

// ---------------------------------------------------------------------------
// Fraud HTTP server: block response
// ---------------------------------------------------------------------------

func TestCallFraudService_HTTPBlock_B18(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"is_fraud":true,"recommended_action":"block","confidence":0.99}`))
	}))
	defer server.Close()

	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	svc.fraudServiceURL = server.URL

	req := newReq()
	req.Device.IP = "8.8.8.8"

	isFraud, err, _ := svc.callFraudService(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isFraud {
		t.Error("expected isFraud=true for block response")
	}
}

// ---------------------------------------------------------------------------
// callOptimizationService: HTTP success branch
// ---------------------------------------------------------------------------

func TestCallOptimizationService_HTTPSuccess_B18(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"recommended_bid":2.5,"multiplier":1.25,"reasons":["high_value"]}`))
	}))
	defer server.Close()

	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	svc.optServiceURL = server.URL

	camp := newCampaign(2.0)
	camp.Budget = 1000.0
	camp.DailyBudget = 100.0
	bid := &model.BidResult{Campaign: camp}
	req := newReq()

	rec, err, hop := svc.callOptimizationService(bid, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec == nil {
		t.Fatal("expected recommendation")
	}
	if hop == nil || !hop.Success {
		t.Error("expected successful hop")
	}
	// Circuit breaker should be reset after success
	if svc.optFailureCount != 0 {
		t.Error("expected failure count reset to 0 after success")
	}
}

// ---------------------------------------------------------------------------
// callAIMatchingService: HTTP success branch (circuit breaker reset)
// ---------------------------------------------------------------------------

func TestCallAIMatchingService_HTTPSuccess_B18(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"recommendations":[{"campaign_id":"c1","score":0.9}]}`))
	}))
	defer server.Close()

	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	svc.aiServiceURL = server.URL
	// Start with some failures so reset is verifiable
	svc.aiFailureCount = 2

	req := newReq()

	recs, err, hop := svc.callAIMatchingService(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recs == nil {
		t.Fatal("expected recommendations slice")
	}
	if hop == nil || !hop.Success {
		t.Error("expected successful hop")
	}
	if svc.aiFailureCount != 0 {
		t.Error("expected AI failure count reset to 0")
	}
}
