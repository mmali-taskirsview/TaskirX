package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// BOOST 34: Target 3 functions with 85-89% coverage
// 1. calculateAutoBidMultiplier (88.2%)
// 2. predictCPL (85.7%)
// 3. estimateDaysUntilChurn (87.5%)

func makeMinCamp_B34() *model.Campaign {
	return &model.Campaign{
		ID:       "camp-b34",
		BidPrice: 2.0,
	}
}

func makeMinReq_B34() *model.BidRequest {
	return &model.BidRequest{
		ID:      "req-b34",
		User:    model.InternalUser{ID: "user123"},
		Device:  model.InternalDevice{Type: "desktop", IP: "1.2.3.4"}, // desktop = 0.95x CTR
		Context: make(map[string]interface{}),
	}
}

func newBiddingSvc_B34() *BiddingService {
	mc := NewMockCache()
	return NewBiddingService(mc, "")
}

// calculateAutoBidMultiplier Tests
func TestB34_AutoBid_HighCTRLowWinRate(t *testing.T) {
	svc := newBiddingSvc_B34()
	camp := makeMinCamp_B34()
	mc := svc.cache.(*MockCache)
	mc.ctr[camp.ID] = 0.025
	mc.winRate[camp.ID] = 0.25

	mult := svc.calculateAutoBidMultiplier(camp)
	if mult != 1.20 {
		t.Errorf("Expected 1.20, got %f", mult)
	}
}

func TestB34_AutoBid_LowCTRHighWinRate(t *testing.T) {
	svc := newBiddingSvc_B34()
	camp := makeMinCamp_B34()
	mc := svc.cache.(*MockCache)
	mc.ctr[camp.ID] = 0.004
	mc.winRate[camp.ID] = 0.75

	mult := svc.calculateAutoBidMultiplier(camp)
	if mult != 0.80 {
		t.Errorf("Expected 0.80, got %f", mult)
	}
}

func TestB34_AutoBid_ModerateCTRVeryLowWinRate(t *testing.T) {
	svc := newBiddingSvc_B34()
	camp := makeMinCamp_B34()
	mc := svc.cache.(*MockCache)
	mc.ctr[camp.ID] = 0.020
	mc.winRate[camp.ID] = 0.15

	mult := svc.calculateAutoBidMultiplier(camp)
	if mult != 1.10 {
		t.Errorf("Expected 1.10, got %f", mult)
	}
}

func TestB34_AutoBid_LowCTRModerateWinRate(t *testing.T) {
	svc := newBiddingSvc_B34()
	camp := makeMinCamp_B34()
	mc := svc.cache.(*MockCache)
	mc.ctr[camp.ID] = 0.008
	mc.winRate[camp.ID] = 0.60

	mult := svc.calculateAutoBidMultiplier(camp)
	if mult != 0.90 {
		t.Errorf("Expected 0.90, got %f", mult)
	}
}

func TestB34_AutoBid_AcceptablePerformance(t *testing.T) {
	svc := newBiddingSvc_B34()
	camp := makeMinCamp_B34()
	mc := svc.cache.(*MockCache)
	mc.ctr[camp.ID] = 0.015
	mc.winRate[camp.ID] = 0.45

	mult := svc.calculateAutoBidMultiplier(camp)
	if mult != 1.0 {
		t.Errorf("Expected 1.0, got %f", mult)
	}
}

func TestB34_AutoBid_NoCTRData(t *testing.T) {
	svc := newBiddingSvc_B34()
	camp := makeMinCamp_B34()

	mult := svc.calculateAutoBidMultiplier(camp)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 when no CTR data, got %f", mult)
	}
}

// predictCPL Tests
func TestB34_PredictCPL_Default(t *testing.T) {
	svc := newBiddingSvc_B34()
	camp := makeMinCamp_B34()
	camp.BidPrice = 5.0
	req := makeMinReq_B34()
	perf := performanceData{ctr: 0.02} // desktop 0.95x → 0.019

	cpl := svc.predictCPL(camp, req, perf)
	// predictCTR: 0.02 * 0.95 = 0.019
	// CPL = 5.0 / (0.019 * 0.05) ≈ 5263.16
	expected := 5.0 / (0.02 * 0.95 * 0.05)
	if cpl != expected {
		t.Errorf("Expected %f, got %f", expected, cpl)
	}
}

func TestB34_PredictCPL_HistoricalLeadRate(t *testing.T) {
	svc := newBiddingSvc_B34()
	camp := makeMinCamp_B34()
	camp.BidPrice = 5.0
	req := makeMinReq_B34()
	req.Context["historical_lead_rate"] = 0.10
	perf := performanceData{ctr: 0.02}

	cpl := svc.predictCPL(camp, req, perf)
	// predictCTR: 0.02 * 0.95 = 0.019
	// CPL = 5.0 / (0.019 * 0.10) = 2631.58
	expected := 5.0 / (0.02 * 0.95 * 0.10)
	if cpl != expected {
		t.Errorf("Expected %f, got %f", expected, cpl)
	}
}

func TestB34_PredictCPL_B2BAdjustment(t *testing.T) {
	svc := newBiddingSvc_B34()
	camp := makeMinCamp_B34()
	camp.BidPrice = 5.0
	req := makeMinReq_B34()
	req.Context["is_b2b"] = true
	perf := performanceData{ctr: 0.02}

	cpl := svc.predictCPL(camp, req, perf)
	// predictCTR: 0.02 * 0.95 = 0.019
	// lead rate: 0.05 * 0.7 = 0.035
	// CPL = 5.0 / (0.019 * 0.035)
	expected := 5.0 / (0.02 * 0.95 * 0.05 * 0.7)
	tolerance := 0.01
	if cpl < expected-tolerance || cpl > expected+tolerance {
		t.Errorf("Expected %f, got %f", expected, cpl)
	}
}

func TestB34_PredictCPL_ZeroCTR(t *testing.T) {
	svc := newBiddingSvc_B34()
	camp := makeMinCamp_B34()
	camp.BidPrice = 5.0
	req := makeMinReq_B34()
	perf := performanceData{ctr: 0.0}

	cpl := svc.predictCPL(camp, req, perf)
	// predictCTR uses default 0.01, then 0.01 * 0.95 = 0.0095
	// CPL = 5.0 / (0.0095 * 0.05)
	expected := 5.0 / (0.01 * 0.95 * 0.05)
	if cpl != expected {
		t.Errorf("Expected %f, got %f", expected, cpl)
	}
}

func TestB34_PredictCPL_HistoricalWithB2B(t *testing.T) {
	svc := newBiddingSvc_B34()
	camp := makeMinCamp_B34()
	camp.BidPrice = 10.0
	req := makeMinReq_B34()
	req.Context["historical_lead_rate"] = 0.08
	req.Context["is_b2b"] = true
	perf := performanceData{ctr: 0.03}

	cpl := svc.predictCPL(camp, req, perf)
	// predictCTR: 0.03 * 0.95 = 0.0285
	// lead rate: 0.08 * 0.7 = 0.056
	// CPL = 10.0 / (0.0285 * 0.056)
	expected := 10.0 / (0.03 * 0.95 * 0.08 * 0.7)
	tolerance := 0.01
	if cpl < expected-tolerance || cpl > expected+tolerance {
		t.Errorf("Expected %f, got %f", expected, cpl)
	}
}

// estimateDaysUntilChurn Tests
func TestB34_ChurnDays_LowRisk(t *testing.T) {
	mc := NewMockCache()
	svc := NewChurnPredictionService(mc)
	features := map[string]float64{"recent_activity_drop": 0.2}

	days := svc.estimateDaysUntilChurn(0.25, features)
	if days != 90 {
		t.Errorf("Expected 90, got %d", days)
	}
}

func TestB34_ChurnDays_HighRiskNoActivityDrop(t *testing.T) {
	mc := NewMockCache()
	svc := NewChurnPredictionService(mc)
	features := map[string]float64{"recent_activity_drop": 0.3}

	days := svc.estimateDaysUntilChurn(0.75, features)
	if days != 15 {
		t.Errorf("Expected 15, got %d", days)
	}
}

func TestB34_ChurnDays_HighRiskWithActivityDrop(t *testing.T) {
	mc := NewMockCache()
	svc := NewChurnPredictionService(mc)
	features := map[string]float64{"recent_activity_drop": 0.6}

	days := svc.estimateDaysUntilChurn(0.75, features)
	if days != 7 {
		t.Errorf("Expected 7, got %d", days)
	}
}

func TestB34_ChurnDays_VeryHighRisk(t *testing.T) {
	mc := NewMockCache()
	svc := NewChurnPredictionService(mc)
	features := map[string]float64{"recent_activity_drop": 0.8}

	days := svc.estimateDaysUntilChurn(0.95, features)
	if days != 1 {
		t.Errorf("Expected 1, got %d", days)
	}
}

func TestB34_ChurnDays_ModerateRisk(t *testing.T) {
	mc := NewMockCache()
	svc := NewChurnPredictionService(mc)
	features := map[string]float64{"recent_activity_drop": 0.4}

	days := svc.estimateDaysUntilChurn(0.50, features)
	if days != 30 {
		t.Errorf("Expected 30, got %d", days)
	}
}
