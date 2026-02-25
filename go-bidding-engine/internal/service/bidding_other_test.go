package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================================
// optimizeForDCPM
// ============================================================================

func TestOptimizeForDCPM_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForDCPM(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetDCPM: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for zero TargetDCPM, got %f", result)
	}
}

func TestOptimizeForDCPM_NoBidPrice(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetDCPM: 5.0}
	result := s.optimizeForDCPM(newCampaign(0), newReq(), pg, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for zero bid price, got %f", result)
	}
}

func TestOptimizeForDCPM_HighWinRate(t *testing.T) {
	// winRate > 0.4 → ratio * 0.8
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetDCPM: 5.0}
	perf := performanceData{winRate: 0.5, engagementRate: 0.04}
	result := s.optimizeForDCPM(newCampaign(1.0), newReq(), pg, perf)
	if result < 0.3 || result > 2.5 {
		t.Errorf("unexpected result %f", result)
	}
}

func TestOptimizeForDCPM_LowWinRate(t *testing.T) {
	// winRate < 0.1 → ratio * 1.3
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetDCPM: 5.0}
	perf := performanceData{winRate: 0.05, engagementRate: 0.04}
	result := s.optimizeForDCPM(newCampaign(1.0), newReq(), pg, perf)
	if result < 0.3 || result > 2.5 {
		t.Errorf("unexpected result %f", result)
	}
}

// ============================================================================
// optimizeForCPIAAP
// ============================================================================

func TestOptimizeForCPIAAP_NoTarget(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.optimizeForCPIAAP(newCampaign(1.0), newReq(), &model.PerformanceGoals{TargetCPIAAP: 0}, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for zero target, got %f", result)
	}
}

func TestOptimizeForCPIAAP_NoBidPrice(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPIAAP: 10.0}
	result := s.optimizeForCPIAAP(newCampaign(0), newReq(), pg, newPerfData())
	if result != 1.0 {
		t.Errorf("expected 1.0 for zero bid, got %f", result)
	}
}

func TestOptimizeForCPIAAP_MobileHighIAP(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPIAAP: 10.0}
	req := &model.BidRequest{
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"historical_iap_rate": 0.05,
			"purchase_propensity": 0.5,
			"avg_iap_value":       float64(25),
		},
	}
	result := s.optimizeForCPIAAP(newCampaign(1.0), req, pg, newPerfData())
	if result < 0.2 || result > 3.0 {
		t.Errorf("unexpected result %f", result)
	}
}

func TestOptimizeForCPIAAP_DesktopPenalty(t *testing.T) {
	// Desktop → maxBid * 0.05 → very low ratio → floor 0.2
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPIAAP: 10.0}
	req := &model.BidRequest{Device: model.InternalDevice{Type: "desktop"}}
	result := s.optimizeForCPIAAP(newCampaign(1.0), req, pg, newPerfData())
	if result != 0.2 {
		t.Errorf("expected floor 0.2 for desktop, got %f", result)
	}
}

func TestOptimizeForCPIAAP_WhaleSegment(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	pg := &model.PerformanceGoals{TargetCPIAAP: 100.0}
	req := &model.BidRequest{
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"user_segments": []interface{}{"whale_user"},
		},
	}
	result := s.optimizeForCPIAAP(newCampaign(0.1), req, pg, newPerfData())
	if result < 0.2 || result > 3.0 {
		t.Errorf("unexpected result %f", result)
	}
}

// ============================================================================
// calculateDealTargetingMultiplier
// ============================================================================

func TestCalculateDealTargeting_NilDealTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = nil
	result := s.calculateDealTargetingMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked with nil DealTargeting")
	}
}

func TestCalculateDealTargeting_NilPMPNoDeals(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{RequireDeal: false}
	req := &model.BidRequest{}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked when no deals and RequireDeal=false")
	}
	if result.DealType != "open" {
		t.Errorf("expected deal_type=open, got %s", result.DealType)
	}
}

func TestCalculateDealTargeting_RequireDealNoDeals(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		RequireDeal:    true,
		FallbackToOpen: false,
	}
	req := &model.BidRequest{}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked when RequireDeal=true but no deals available")
	}
}

func TestCalculateDealTargeting_RequireDealFallback(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		RequireDeal:    true,
		FallbackToOpen: true,
	}
	req := &model.BidRequest{}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked with FallbackToOpen=true")
	}
}

func TestCalculateDealTargeting_WithMatchingDeal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		PreferredDealIDs: []string{"deal-123"},
	}
	req := &model.BidRequest{
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal-123", BidFloor: 0.5, At: 1},
			},
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if !result.Matched {
		t.Error("expected deal matched")
	}
	if result.MatchedDealID != "deal-123" {
		t.Errorf("expected deal-123, got %s", result.MatchedDealID)
	}
	if result.Multiplier < 1.0 {
		t.Errorf("expected multiplier >= 1.0 for preferred deal, got %f", result.Multiplier)
	}
}

func TestCalculateDealTargeting_LegacyDealID(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.DealID = "legacy-deal-1"
	camp.Targeting.DealTargeting = nil
	req := &model.BidRequest{
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "legacy-deal-1", BidFloor: 0.5, At: 1},
			},
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	_ = result // just ensure no panic
}

// ============================================================================
// GetAutoBidRecommendations
// ============================================================================

func TestGetAutoBidRecommendations_NoCampaigns(t *testing.T) {
	mc := NewMockCache()
	mc.campaigns = []*model.Campaign{}
	s := NewBiddingService(mc, "")
	recs, err := s.GetAutoBidRecommendations()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(recs) != 0 {
		t.Errorf("expected empty recs, got %d", len(recs))
	}
}

func TestGetAutoBidRecommendations_HighCTRLowWin(t *testing.T) {
	mc := NewMockCache()
	mc.campaigns = []*model.Campaign{
		{ID: "camp-1", BidPrice: 1.0, Status: "active"},
	}
	mc.ctr["camp-1"] = 0.08   // High CTR
	mc.winRate["camp-1"] = 0.05 // Low win rate → increase
	s := NewBiddingService(mc, "")
	recs, err := s.GetAutoBidRecommendations()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(recs) == 0 {
		t.Error("expected at least one recommendation for high CTR/low win rate")
	}
	if len(recs) > 0 {
		action, _ := recs[0]["action"].(string)
		if action != "increase" {
			t.Errorf("expected action=increase, got %s", action)
		}
	}
}

func TestGetAutoBidRecommendations_LowCTRHighWin(t *testing.T) {
	mc := NewMockCache()
	mc.campaigns = []*model.Campaign{
		{ID: "camp-2", BidPrice: 2.0, Status: "active"},
	}
	mc.ctr["camp-2"] = 0.001   // Low CTR
	mc.winRate["camp-2"] = 0.60 // High win rate → decrease
	s := NewBiddingService(mc, "")
	recs, err := s.GetAutoBidRecommendations()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(recs) == 0 {
		t.Error("expected at least one recommendation for low CTR/high win rate")
	}
	if len(recs) > 0 {
		action, _ := recs[0]["action"].(string)
		if action != "decrease" {
			t.Errorf("expected action=decrease, got %s", action)
		}
	}
}

// ============================================================================
// ResolveCrossDeviceUser
// ============================================================================

func TestResolveCrossDeviceUser_EmptyDeviceID(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.ResolveCrossDeviceUser("", nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.PrimaryUserID != "" {
		t.Errorf("expected empty primary user ID for empty deviceID, got %s", result.PrimaryUserID)
	}
}

func TestResolveCrossDeviceUser_NewDeviceNotInCache(t *testing.T) {
	// Device not in primaryUserID map → primaryID == "" → use deviceID as default
	mc := NewMockCache()
	s := NewBiddingService(mc, "")
	result := s.ResolveCrossDeviceUser("dev-001", nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.PrimaryUserID != "dev-001" {
		t.Errorf("expected primary=dev-001, got %s", result.PrimaryUserID)
	}
	if result.DeviceCount < 1 {
		t.Errorf("expected at least 1 device, got %d", result.DeviceCount)
	}
}

func TestResolveCrossDeviceUser_KnownDevice(t *testing.T) {
	mc := NewMockCache()
	mc.primaryUserID["dev-001"] = "user-primary-1"
	mc.linkedDevices["user-primary-1"] = []string{"dev-001", "dev-002", "dev-003"}
	s := NewBiddingService(mc, "")
	result := s.ResolveCrossDeviceUser("dev-001", nil)
	if result.PrimaryUserID != "user-primary-1" {
		t.Errorf("expected user-primary-1, got %s", result.PrimaryUserID)
	}
	if result.DeviceCount != 3 {
		t.Errorf("expected 3 linked devices, got %d", result.DeviceCount)
	}
}

func TestResolveCrossDeviceUser_DeterministicLink(t *testing.T) {
	mc := NewMockCache()
	// email_hash:abc123 → user-existing
	mc.primaryUserID["email_hash:abc123"] = "user-existing"
	mc.linkedDevices["user-existing"] = []string{"dev-old"}
	s := NewBiddingService(mc, "")
	signals := map[string]string{"email_hash": "abc123"}
	result := s.ResolveCrossDeviceUser("dev-new", signals)
	if result.PrimaryUserID != "user-existing" {
		t.Errorf("expected deterministic link to user-existing, got %s", result.PrimaryUserID)
	}
	if !result.IsNewDevice {
		t.Error("expected IsNewDevice=true for newly linked device")
	}
}

// ============================================================================
// RefreshCampaigns
// ============================================================================

func TestRefreshCampaigns_Success(t *testing.T) {
	campaigns := `[{"id":"c1","bid_price":1.5,"status":"active"},{"id":"c2","bid_price":2.5,"status":"active"}]`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, campaigns)
	}))
	defer ts.Close()

	s := NewBiddingService(NewMockCache(), "")
	err := s.RefreshCampaigns(ts.URL)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRefreshCampaigns_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := NewBiddingService(NewMockCache(), "")
	err := s.RefreshCampaigns(ts.URL)
	if err == nil {
		t.Error("expected error for server 500 response")
	}
}

func TestRefreshCampaigns_InvalidURL(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	err := s.RefreshCampaigns("http://127.0.0.1:1")
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestRefreshCampaigns_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "not json")
	}))
	defer ts.Close()

	s := NewBiddingService(NewMockCache(), "")
	err := s.RefreshCampaigns(ts.URL)
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}
