package service

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================================
// GetMetrics — simulated error and zero-bids win-rate paths
// ============================================================================

func TestGetMetrics_SimulatedError(t *testing.T) {
	mc := NewMockCache()
	mc.kv["SIMULATE_METRICS_ERROR"] = "1"
	s := NewBiddingService(mc, "")
	_, err := s.GetMetrics()
	if err == nil {
		t.Error("expected error for SIMULATE_METRICS_ERROR")
	}
}

func TestGetMetrics_ZeroBids(t *testing.T) {
	mc := NewMockCache()
	s := NewBiddingService(mc, "")
	metrics, err := s.GetMetrics()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	winRate, _ := metrics["win_rate"].(float64)
	if winRate != 0.0 {
		t.Errorf("expected win_rate=0.0 when bids=0, got %f", winRate)
	}
}

func TestGetMetrics_WithBids(t *testing.T) {
	mc := NewMockCache()
	// Inject bid/win counts via IncrementBidCount/IncrementWinCount
	// Since MockCache tracks these in-memory, we just call the service
	s := NewBiddingService(mc, "")
	s.cache.IncrementBidCount()
	s.cache.IncrementBidCount()
	s.cache.IncrementWinCount()
	metrics, err := s.GetMetrics()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := metrics["total_bids"]; !ok {
		t.Error("expected total_bids key in metrics")
	}
}

// ============================================================================
// callFraudService — cache "block" and "allow" paths
// ============================================================================

func TestCallFraudService_CacheBlock(t *testing.T) {
	mc := NewMockCache()
	mc.kv["ip_rep:10.0.0.1"] = "block"
	s := NewBiddingService(mc, "")
	req := &model.BidRequest{
		Device: model.InternalDevice{IP: "10.0.0.1"},
	}
	blocked, err, hop := s.callFraudService(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !blocked {
		t.Error("expected blocked=true from cache 'block'")
	}
	if hop == nil {
		t.Error("expected non-nil hop")
	}
	if hop.ServiceName != "fraud-detection-cache" {
		t.Errorf("expected fraud-detection-cache hop, got %s", hop.ServiceName)
	}
}

func TestCallFraudService_CacheAllow(t *testing.T) {
	mc := NewMockCache()
	mc.kv["ip_rep:10.0.0.2"] = "allow"
	s := NewBiddingService(mc, "")
	req := &model.BidRequest{
		Device: model.InternalDevice{IP: "10.0.0.2"},
	}
	blocked, err, hop := s.callFraudService(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if blocked {
		t.Error("expected blocked=false from cache 'allow'")
	}
	if hop == nil {
		t.Error("expected non-nil hop")
	}
}

func TestCallFraudService_CacheMiss_FallbackToAPI(t *testing.T) {
	// No cache entry → falls through to HTTP call (will fail, but should not panic)
	mc := NewMockCache()
	s := NewBiddingService(mc, "")
	req := &model.BidRequest{
		Device: model.InternalDevice{IP: "192.168.1.100"},
	}
	// API call will fail (no server), but function must return without panic
	_, _, _ = s.callFraudService(req)
}

// ============================================================================
// callAIMatchingService — circuit breaker tripped
// ============================================================================

func TestCallAIMatchingService_CircuitBreakerOpen(t *testing.T) {
	mc := NewMockCache()
	s := NewBiddingService(mc, "")

	// Trip the circuit breaker: 3+ failures within 10 minutes
	s.aiMutex.Lock()
	s.aiFailureCount = 5
	s.aiLastFailure = time.Now()
	s.aiMutex.Unlock()

	req := newReq()
	recs, err, hop := s.callAIMatchingService(req)
	if err == nil {
		t.Error("expected error from open circuit breaker")
	}
	if recs != nil {
		t.Error("expected nil recs from open circuit breaker")
	}
	if hop == nil {
		t.Error("expected non-nil hop from circuit breaker")
	}
}

// ============================================================================
// GetCrossDeviceFrequency — error returns 0
// ============================================================================

func TestGetCrossDeviceFrequency_ErrorReturnsZero(t *testing.T) {
	// MockCache returns 0,nil for unknown primary users
	mc := NewMockCache()
	s := NewBiddingService(mc, "")
	freq := s.GetCrossDeviceFrequency("unknown-user", "camp-x")
	if freq != 0 {
		t.Errorf("expected 0 for unknown cross-device, got %d", freq)
	}
}

func TestGetCrossDeviceFrequency_WithData(t *testing.T) {
	mc := NewMockCache()
	mc.crossDeviceFreq["primary-1:camp-1"] = 5
	s := NewBiddingService(mc, "")
	freq := s.GetCrossDeviceFrequency("primary-1", "camp-1")
	if freq != 5 {
		t.Errorf("expected 5, got %d", freq)
	}
}

// ============================================================================
// GetUserDeviceGraph — empty userID error, cache error, success
// ============================================================================

func TestGetUserDeviceGraph_EmptyUserID(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	_, err := s.GetUserDeviceGraph("")
	if err == nil {
		t.Error("expected error for empty userID")
	}
}

func TestGetUserDeviceGraph_WithPrimaryUser(t *testing.T) {
	mc := NewMockCache()
	mc.primaryUserID["device-abc"] = "primary-user-1"
	mc.linkedDevices["primary-user-1"] = []string{"device-abc", "device-def"}
	s := NewBiddingService(mc, "")
	result, err := s.GetUserDeviceGraph("device-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PrimaryUserID != "primary-user-1" {
		t.Errorf("expected primary-user-1, got %s", result.PrimaryUserID)
	}
	if result.DeviceCount != 2 {
		t.Errorf("expected 2 devices, got %d", result.DeviceCount)
	}
}

func TestGetUserDeviceGraph_UnknownUser_UsesAsOwnPrimary(t *testing.T) {
	// No primaryUserID mapping → userID used as its own primary
	mc := NewMockCache()
	mc.linkedDevices["orphan-device"] = []string{"orphan-device"}
	s := NewBiddingService(mc, "")
	result, err := s.GetUserDeviceGraph("orphan-device")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PrimaryUserID != "orphan-device" {
		t.Errorf("expected orphan-device as primary, got %s", result.PrimaryUserID)
	}
}

// ============================================================================
// getHouseholdID — hh_id and ifa context keys
// ============================================================================

func TestGetHouseholdID_HhIdKey(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"hh_id": "household-99",
		},
	}
	id := s.getHouseholdID(req)
	if id != "household-99" {
		t.Errorf("expected household-99 from hh_id, got %s", id)
	}
}

func TestGetHouseholdID_IfaKey(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"ifa": "ifa-abc-123",
		},
	}
	id := s.getHouseholdID(req)
	if id != "ifa-abc-123" {
		t.Errorf("expected ifa-abc-123 from ifa key, got %s", id)
	}
}

func TestGetHouseholdID_NilContext(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	req := &model.BidRequest{}
	id := s.getHouseholdID(req)
	if id != "" {
		t.Errorf("expected empty string for nil context, got %s", id)
	}
}

// ============================================================================
// PauseDeal / CancelDeal — error paths
// ============================================================================

func newActivePGDeal(id string) *PGDeal {
	return &PGDeal{
		ID:        id,
		Status:    "active",
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}
}

func TestPauseDeal_NotFoundError(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	err := svc.PauseDeal("nonexistent-deal")
	if err == nil {
		t.Error("expected error for non-existent deal")
	}
}

func TestPauseDeal_AlreadyPaused(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	deal := newActivePGDeal("deal-p1")
	deal.Status = "paused"
	svc.deals.Store("deal-p1", deal)
	err := svc.PauseDeal("deal-p1")
	if err == nil {
		t.Error("expected error when pausing already-paused deal")
	}
}

func TestPauseDeal_Success(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	deal := newActivePGDeal("deal-p2")
	svc.deals.Store("deal-p2", deal)
	err := svc.PauseDeal("deal-p2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated, _ := svc.GetDeal("deal-p2")
	if updated.Status != "paused" {
		t.Errorf("expected paused, got %s", updated.Status)
	}
}

func TestCancelDeal_NotFoundError(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	err := svc.CancelDeal("nonexistent-deal")
	if err == nil {
		t.Error("expected error for non-existent deal")
	}
}

func TestCancelDeal_AlreadyCompleted(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	deal := newActivePGDeal("deal-c1")
	deal.Status = "completed"
	svc.deals.Store("deal-c1", deal)
	err := svc.CancelDeal("deal-c1")
	if err == nil {
		t.Error("expected error when cancelling completed deal")
	}
}

func TestCancelDeal_AlreadyCancelled(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	deal := newActivePGDeal("deal-c2")
	deal.Status = "cancelled"
	svc.deals.Store("deal-c2", deal)
	err := svc.CancelDeal("deal-c2")
	if err == nil {
		t.Error("expected error when cancelling already-cancelled deal")
	}
}

func TestCancelDeal_Success(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	deal := newActivePGDeal("deal-c3")
	svc.deals.Store("deal-c3", deal)
	err := svc.CancelDeal("deal-c3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated, _ := svc.GetDeal("deal-c3")
	if updated.Status != "cancelled" {
		t.Errorf("expected cancelled, got %s", updated.Status)
	}
}

// ============================================================================
// CheckEligibility — active deal, overdelivered, inventory mismatch, eligible
// ============================================================================

func TestCheckEligibility_NilDeals(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	result := svc.CheckEligibility("pub1", "site1", "banner", "display", "mobile", "US")
	if len(result) != 0 {
		t.Errorf("expected empty eligibility with no deals, got %d", len(result))
	}
}

func TestCheckEligibility_InactiveDeal(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	deal := newActivePGDeal("deal-elig-1")
	deal.Status = "paused"
	svc.deals.Store("deal-elig-1", deal)
	result := svc.CheckEligibility("pub1", "", "", "", "", "")
	if len(result) != 0 {
		t.Errorf("expected no eligible deals for inactive deal")
	}
}

func TestCheckEligibility_ExpiredDeal(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	deal := newActivePGDeal("deal-elig-2")
	deal.EndDate = time.Now().Add(-1 * time.Hour) // already expired
	svc.deals.Store("deal-elig-2", deal)
	result := svc.CheckEligibility("pub1", "", "", "", "", "")
	if len(result) != 0 {
		t.Errorf("expected no eligible deals for expired deal")
	}
}

func TestCheckEligibility_EligibleDeal(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	deal := newActivePGDeal("deal-elig-3")
	deal.FixedPrice = 2.5
	deal.Priority = 8
	// Empty InventorySpecs matches everything
	svc.deals.Store("deal-elig-3", deal)
	result := svc.CheckEligibility("any-pub", "any-site", "any-placement", "banner", "mobile", "US")
	if len(result) == 0 {
		t.Error("expected at least one eligible deal")
	}
	if !result[0].Eligible {
		t.Errorf("expected Eligible=true, reason: %s", result[0].Reason)
	}
}

// ============================================================================
// GetDeliveryProgress (public) — not found and found
// ============================================================================

func TestGetDeliveryProgress_Public_NotFound(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	_, err := svc.GetDeliveryProgress("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent delivery progress")
	}
}

func TestGetDeliveryProgress_Public_Found(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(NewMockCache())
	progress := &DeliveryProgress{DealID: "deal-dp1", Status: "on_pace"}
	svc.deliveryTracker.Store("deal-dp1", progress)
	result, err := svc.GetDeliveryProgress("deal-dp1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DealID != "deal-dp1" {
		t.Errorf("expected deal-dp1, got %s", result.DealID)
	}
}

// ============================================================================
// GetExperiment — not found error path
// ============================================================================

func TestGetExperiment_NotFound(t *testing.T) {
	svc := NewABTestingService(nil)
	_, err := svc.GetExperiment("nonexistent-exp")
	if err == nil {
		t.Error("expected error for non-existent experiment")
	}
}

func TestGetExperiment_Found(t *testing.T) {
	svc := NewABTestingService(nil)
	exp, _ := svc.CreateExperiment(CreateExperimentRequest{
		Name: "test-exp",
		Type: "ab",
		Variants: []VariantRequest{
			{Name: "control", Weight: 0.5, IsControl: true},
			{Name: "variant1", Weight: 0.5},
		},
	})
	result, err := svc.GetExperiment(exp.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != exp.ID {
		t.Errorf("expected %s, got %s", exp.ID, result.ID)
	}
}

// ============================================================================
// determineRecommendedAction — all 5 branches
// ============================================================================

func TestDetermineRecommendedAction_HighMarketLowShare(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	config := &model.CompetitiveIntelligence{
		Enabled:         true,
		CompetitiveMode: "balanced",
	}
	result := &model.CompetitiveIntelResult{
		MarketCondition:   "high",
		OurShareOfVoice:   0.05, // < 0.1
		CompetitorsActive: 2,
	}
	action := svc.determineRecommendedAction(config, result)
	if action != "increase_budget" {
		t.Errorf("expected increase_budget, got %s", action)
	}
}

func TestDetermineRecommendedAction_ManyCompetitors(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	config := &model.CompetitiveIntelligence{
		Enabled:         true,
		CompetitiveMode: "balanced",
	}
	result := &model.CompetitiveIntelResult{
		MarketCondition:   "medium",
		OurShareOfVoice:   0.3,
		CompetitorsActive: 6, // > 5
	}
	action := svc.determineRecommendedAction(config, result)
	if action != "optimize_targeting" {
		t.Errorf("expected optimize_targeting, got %s", action)
	}
}

func TestDetermineRecommendedAction_AggressiveBidding(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	config := &model.CompetitiveIntelligence{
		Enabled:         true,
		CompetitiveMode: "balanced",
		MarketShareGoal: 0.4,
	}
	result := &model.CompetitiveIntelResult{
		MarketCondition:   "medium",
		OurShareOfVoice:   0.1, // < 0.4 * 0.5 = 0.2
		CompetitorsActive: 2,
	}
	action := svc.determineRecommendedAction(config, result)
	if action != "aggressive_bidding" {
		t.Errorf("expected aggressive_bidding, got %s", action)
	}
}

func TestDetermineRecommendedAction_MaintainPosition(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	config := &model.CompetitiveIntelligence{
		Enabled:         true,
		CompetitiveMode: "balanced",
	}
	result := &model.CompetitiveIntelResult{
		MarketCondition:   "low",
		OurShareOfVoice:   0.5,
		CompetitorsActive: 1,
	}
	action := svc.determineRecommendedAction(config, result)
	if action != "maintain_position" {
		t.Errorf("expected maintain_position, got %s", action)
	}
}

func TestDetermineRecommendedAction_NoChange(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	config := &model.CompetitiveIntelligence{
		Enabled:         true,
		CompetitiveMode: "balanced",
	}
	result := &model.CompetitiveIntelResult{
		MarketCondition:   "medium",
		OurShareOfVoice:   0.3,
		CompetitorsActive: 3,
	}
	action := svc.determineRecommendedAction(config, result)
	if action != "no_change" {
		t.Errorf("expected no_change, got %s", action)
	}
}

// ============================================================================
// calculateSeasonalMultiplier — Q4, Summer, BackToSchool, Events, Holidays
// ============================================================================

func TestCalculateSeasonalMultiplier_WeekendBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		WeekendBoost: 1.4,
	}
	result := s.calculateSeasonalMultiplier(camp)
	// Test runs on any day; just verify no panic and multiplier >= 1.0
	if result.Multiplier < 0.5 {
		t.Errorf("expected multiplier >= 0.5, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_Q4Boost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		Q4Boost: 1.5,
	}
	result := s.calculateSeasonalMultiplier(camp)
	// Multiplier will only differ in Q4 months (Oct-Dec), but no panic
	if result.Multiplier < 0.5 {
		t.Errorf("expected valid multiplier, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_SummerBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		SummerBoost: 1.3,
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier < 0.5 {
		t.Errorf("expected valid multiplier, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_BackToSchoolBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		BackToSchoolBoost: 1.2,
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier < 0.5 {
		t.Errorf("expected valid multiplier, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_ActiveCustomEvent(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		Events: []model.SeasonalEvent{
			{
				Name:      "spring-sale",
				Active:    true,
				StartDate: yesterday,
				EndDate:   tomorrow,
				Boost:     1.6,
			},
		},
	}
	result := s.calculateSeasonalMultiplier(camp)
	if !result.Matched {
		t.Error("expected matched for active event")
	}
	if result.Multiplier < 1.5 {
		t.Errorf("expected boost >= 1.5 for active event, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_InactiveEvent(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		Events: []model.SeasonalEvent{
			{
				Name:   "old-sale",
				Active: false, // inactive
				Boost:  2.0,
			},
		},
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier != 1.0 {
		t.Errorf("expected multiplier=1.0 for inactive event, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_DefaultEventBoost(t *testing.T) {
	// Boost=0 → default 1.5
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		Events: []model.SeasonalEvent{
			{
				Name:      "no-boost-event",
				Active:    true,
				StartDate: yesterday,
				EndDate:   tomorrow,
				Boost:     0, // default 1.5
			},
		},
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier < 1.4 {
		t.Errorf("expected default event boost ~1.5, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_HolidayEnabled(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		EnableHolidays: true,
		HolidayBoost:   1.3,
		Country:        "US",
	}
	// Just ensure no panic; holiday will only boost if today is a US holiday
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier < 0.5 {
		t.Errorf("expected valid multiplier, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_HolidayDefaultCountry(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		EnableHolidays: true,
		HolidayBoost:   1.3,
		Country:        "", // default to US
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier < 0.5 {
		t.Errorf("expected valid multiplier, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_HolidayDefaultBoost(t *testing.T) {
	// HolidayBoost=0 → default 1.3
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		EnableHolidays: true,
		HolidayBoost:   0, // default 1.3
		Country:        "US",
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier < 0.5 {
		t.Errorf("expected valid multiplier, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_CapAt3(t *testing.T) {
	// Multiple big boosts → cap at 3.0
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		WeekendBoost:      2.0,
		Q4Boost:           2.0,
		SummerBoost:       2.0,
		BackToSchoolBoost: 2.0,
		Events: []model.SeasonalEvent{
			{Name: "big-sale", Active: true,
				StartDate: yesterday, EndDate: tomorrow, Boost: 2.0},
		},
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier > 3.0 {
		t.Errorf("expected multiplier capped at 3.0, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_Timezone(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		Timezone:     "America/New_York",
		WeekendBoost: 1.2,
	}
	// Should not panic with valid timezone
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier < 0.5 {
		t.Errorf("expected valid multiplier with timezone, got %f", result.Multiplier)
	}
}

// ============================================================================
// calculateCarrierMultiplier — carrier rules required, ISP required,
// ConnectionTypes filter, WiFiOnly block
// ============================================================================

func TestCalculateCarrierMultiplier_WiFiOnlyBlock(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		WiFiOnly: true,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"connection_type": "cellular"},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for WiFiOnly with cellular connection")
	}
	if result.Reason != "wifi_only" {
		t.Errorf("expected reason wifi_only, got %s", result.Reason)
	}
}

func TestCalculateCarrierMultiplier_ConnectionTypeNotAllowed(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		ConnectionTypes: []string{"wifi", "ethernet"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"connection_type": "cellular"},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for connection_type not in allowed list")
	}
	if result.Reason != "connection_type_not_allowed" {
		t.Errorf("expected connection_type_not_allowed, got %s", result.Reason)
	}
}

func TestCalculateCarrierMultiplier_ExcludedISP(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		ExcludeISPs: []string{"comcast"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"isp":             "Comcast",
			"connection_type": "wifi",
		},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded ISP")
	}
}

func TestCalculateCarrierMultiplier_RequiredCarrierMissing(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		Carriers: []model.CarrierRule{
			{Name: "verizon", Required: true, Boost: 1.3},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"carrier":         "t-mobile",
			"connection_type": "cellular",
		},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for missing required carrier")
	}
	if result.Reason != "missing_required_carrier" {
		t.Errorf("expected missing_required_carrier, got %s", result.Reason)
	}
}

func TestCalculateCarrierMultiplier_RequiredCarrierMatched(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		Carriers: []model.CarrierRule{
			{Name: "verizon", Required: true, Boost: 1.3},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"carrier":         "verizon",
			"connection_type": "cellular",
		},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for matched required carrier, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.2 {
		t.Errorf("expected carrier boost >= 1.2, got %f", result.Multiplier)
	}
}

func TestCalculateCarrierMultiplier_RequiredISPMissing(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		ISPs: []model.ISPRule{
			{Name: "comcast", Required: true, Boost: 1.2},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"isp":             "spectrum",
			"connection_type": "wifi",
		},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for missing required ISP")
	}
	if result.Reason != "missing_required_isp" {
		t.Errorf("expected missing_required_isp, got %s", result.Reason)
	}
}

func TestCalculateCarrierMultiplier_ISPBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		ISPs: []model.ISPRule{
			{Name: "comcast", Boost: 1.4},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"isp":             "comcast",
			"connection_type": "wifi",
		},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.3 {
		t.Errorf("expected ISP boost >= 1.3, got %f", result.Multiplier)
	}
}

func TestCalculateCarrierMultiplier_ISPDefaultBoost(t *testing.T) {
	// ISP Boost=0 → default 1.2
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		ISPs: []model.ISPRule{
			{Name: "verizon_fios", Boost: 0},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"isp":             "verizon_fios",
			"connection_type": "wifi",
		},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.15 {
		t.Errorf("expected default ISP boost ~1.2, got %f", result.Multiplier)
	}
}

// ============================================================================
// DirectPublisherService GetStats — with data
// ============================================================================

func TestDirectPublisherGetStats_WithPublisher(t *testing.T) {
	mc := NewMockCache()
	svc := NewDirectPublisherService(mc)

	pub := &DirectPublisher{
		ID:               "pub-stat-1",
		Status:           "active",
		IsDirectSeller:   true,
		QualityScore:     0.9,
		TotalImpressions: 1000,
		TotalRevenue:     50.0,
	}
	svc.publishers.Store("pub-stat-1", pub)

	integration := &PublisherIntegration{
		PublisherID: "pub-stat-1",
		Status:      "active",
	}
	svc.integrations.Store("integ-1", integration)

	stats := svc.GetStats()
	if stats["total_publishers"].(int) != 1 {
		t.Errorf("expected 1 total publisher, got %d", stats["total_publishers"].(int))
	}
	if stats["active_publishers"].(int) != 1 {
		t.Errorf("expected 1 active publisher, got %d", stats["active_publishers"].(int))
	}
	if stats["total_integrations"].(int) != 1 {
		t.Errorf("expected 1 total integration, got %d", stats["total_integrations"].(int))
	}
}

func TestDirectPublisherGetStats_Empty(t *testing.T) {
	svc := NewDirectPublisherService(NewMockCache())
	stats := svc.GetStats()
	if stats["total_publishers"].(int) != 0 {
		t.Errorf("expected 0 publishers, got %d", stats["total_publishers"].(int))
	}
}
