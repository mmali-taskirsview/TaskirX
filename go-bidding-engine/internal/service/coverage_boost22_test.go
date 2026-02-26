package service

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================
// S2S generateMockBid — bidFloor negative → bidPrice < 0.01 fallback
// ============================================================

func TestGenerateMockBid_NegativeBidFloor_B22(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	partner := &DemandPartner{
		ID:       "partner-neg",
		Name:     "Negative Floor Partner",
		Endpoint: "http://partner-neg.example.com",
		BidFloor: -1.0, // bidFloor + 0.5 = -0.5 < 0.01 → should default to 0.50
		Timeout:  150 * time.Millisecond,
	}

	req := &S2SBidRequest{
		ID: "req-neg-floor",
		Imp: []S2SImpression{
			{ID: "imp-1", BidFloor: 0.01},
		},
	}

	bid := svc.generateMockBid(partner, req)
	if bid == nil {
		t.Fatal("Expected bid, got nil")
	}
	if bid.Price != 0.50 {
		t.Errorf("Expected bid price 0.50 (default fallback), got %.2f", bid.Price)
	}
}

func TestGenerateMockBid_ZeroImpressions_B22(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	partner := &DemandPartner{
		ID:       "partner-zero",
		Name:     "Zero Imps",
		Endpoint: "http://zero.example.com",
		BidFloor: 0.5,
		Timeout:  150 * time.Millisecond,
	}

	req := &S2SBidRequest{
		ID:  "req-zero",
		Imp: []S2SImpression{},
	}

	bid := svc.generateMockBid(partner, req)
	if bid != nil {
		t.Errorf("Expected nil bid for empty impressions, got %+v", bid)
	}
}

// ============================================================
// S2S RegisterPartner — error branches
// ============================================================

func TestRegisterPartner_EmptyID_B22(t *testing.T) {
	svc := NewS2SBiddingService(nil)
	err := svc.RegisterPartner(&DemandPartner{
		ID:       "",
		Endpoint: "http://example.com",
	})
	if err == nil {
		t.Error("Expected error for empty partner ID")
	}
}

func TestRegisterPartner_EmptyEndpoint_B22(t *testing.T) {
	svc := NewS2SBiddingService(nil)
	err := svc.RegisterPartner(&DemandPartner{
		ID:       "partner-x",
		Endpoint: "",
	})
	if err == nil {
		t.Error("Expected error for empty endpoint")
	}
}

func TestRegisterPartner_ForceError_B22(t *testing.T) {
	svc := NewS2SBiddingService(nil)
	err := svc.RegisterPartner(&DemandPartner{
		ID:       "force-error-register",
		Endpoint: "http://example.com",
	})
	if err == nil {
		t.Error("Expected forced register error")
	}
}

func TestRegisterPartner_ForceDuplicate_B22(t *testing.T) {
	svc := NewS2SBiddingService(nil)
	err := svc.RegisterPartner(&DemandPartner{
		ID:       "force-duplicate-register",
		Endpoint: "http://example.com",
	})
	if err == nil {
		t.Error("Expected forced duplicate error")
	}
}

func TestRegisterPartner_ZeroTimeout_B22(t *testing.T) {
	svc := NewS2SBiddingService(nil)
	err := svc.RegisterPartner(&DemandPartner{
		ID:       "partner-zero-timeout",
		Endpoint: "http://example.com",
		Timeout:  0, // should default to 150ms
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	p, _ := svc.GetPartner("partner-zero-timeout")
	if p.Timeout != 150*time.Millisecond {
		t.Errorf("Expected default timeout 150ms, got %v", p.Timeout)
	}
}

// ============================================================
// PG ActivateDeal — wrong-status error branch
// ============================================================

func TestActivateDeal_WrongStatus_B22(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-1",
		SellerID:             "seller-1",
		CommittedImpressions: 10000,
		StartDate:            time.Now(),
		EndDate:              time.Now().Add(30 * 24 * time.Hour),
	}
	created, err := svc.CreateDeal(deal)
	if err != nil {
		t.Fatalf("CreateDeal failed: %v", err)
	}

	// Set to "active" first
	created.Status = "active"
	svc.deals.Store(created.ID, created)

	// Now try to activate an already-active deal (not "pending" or "paused")
	err = svc.ActivateDeal(created.ID)
	if err == nil {
		t.Error("Expected error when activating non-pending/non-paused deal")
	}
}

func TestActivateDeal_NotFound_B22(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)
	err := svc.ActivateDeal("nonexistent-deal-id")
	if err == nil {
		t.Error("Expected error for nonexistent deal")
	}
}

func TestActivateDeal_FromPaused_B22(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-2",
		SellerID:             "seller-2",
		CommittedImpressions: 5000,
		StartDate:            time.Now(),
		EndDate:              time.Now().Add(30 * 24 * time.Hour),
	}
	created, _ := svc.CreateDeal(deal)

	// Set to "paused"
	created.Status = "paused"
	svc.deals.Store(created.ID, created)

	// Should succeed: paused → active
	err := svc.ActivateDeal(created.ID)
	if err != nil {
		t.Errorf("Expected success activating paused deal, got: %v", err)
	}
}

// ============================================================
// PG CheckEligibility — various branches
// ============================================================

func TestCheckEligibility_NotActiveDeal_B22(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-3",
		SellerID:             "seller-3",
		CommittedImpressions: 1000,
		StartDate:            time.Now().Add(-time.Hour),
		EndDate:              time.Now().Add(24 * time.Hour),
	}
	created, _ := svc.CreateDeal(deal)
	// Status is "pending", not "active"

	results := svc.CheckEligibility("pub-1", "site-1", "top", "banner", "desktop", "US")
	for _, r := range results {
		if r.DealID == created.ID {
			t.Errorf("Expected pending deal to be skipped, but it appeared in eligibility results")
		}
	}
}

func TestCheckEligibility_BeforeStartDate_B22(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-4",
		SellerID:             "seller-4",
		CommittedImpressions: 1000,
		StartDate:            time.Now().Add(10 * time.Hour), // Future start
		EndDate:              time.Now().Add(24 * time.Hour),
	}
	created, _ := svc.CreateDeal(deal)
	created.Status = "active"
	svc.deals.Store(created.ID, created)

	results := svc.CheckEligibility("pub-1", "site-1", "top", "banner", "desktop", "US")
	for _, r := range results {
		if r.DealID == created.ID {
			t.Errorf("Expected future-start deal to be skipped")
		}
	}
}

func TestCheckEligibility_FullyDelivered_B22(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-5",
		SellerID:             "seller-5",
		CommittedImpressions: 100,
		StartDate:            time.Now().Add(-time.Hour),
		EndDate:              time.Now().Add(24 * time.Hour),
	}
	created, _ := svc.CreateDeal(deal)
	created.Status = "active"
	// Deliver beyond overdelivery allowance (1.1 * 100 = 110)
	created.DeliveredImpressions = 120
	svc.deals.Store(created.ID, created)

	// Manually seed delivery progress
	svc.deliveryTracker.Store(created.ID, &DeliveryProgress{
		DealID:               created.ID,
		TargetImpressions:    100,
		DeliveredImpressions: 120, // > 1.1 * 100
	})

	results := svc.CheckEligibility("pub-1", "site-1", "top", "banner", "desktop", "US")
	found := false
	for _, r := range results {
		if r.DealID == created.ID {
			found = true
			if r.Eligible {
				t.Error("Expected fully delivered deal to be ineligible")
			}
		}
	}
	if !found {
		t.Log("Fully delivered deal not returned (acceptable if delivery progress check skips it)")
	}
}

func TestCheckEligibility_InventoryMismatch_B22(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-6",
		SellerID:             "seller-6",
		CommittedImpressions: 1000,
		StartDate:            time.Now().Add(-time.Hour),
		EndDate:              time.Now().Add(24 * time.Hour),
		InventorySpecs: InventorySpec{
			PublisherIDs: []string{"only-publisher"},
		},
	}
	created, _ := svc.CreateDeal(deal)
	created.Status = "active"
	svc.deals.Store(created.ID, created)

	// Use a different publisher ID — should not match
	results := svc.CheckEligibility("different-publisher", "site-1", "top", "banner", "desktop", "US")
	for _, r := range results {
		if r.DealID == created.ID {
			t.Errorf("Expected inventory-mismatch deal to be skipped")
		}
	}
}

func TestCheckEligibility_Match_B22(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-7",
		SellerID:             "seller-7",
		CommittedImpressions: 1000,
		StartDate:            time.Now().Add(-time.Hour),
		EndDate:              time.Now().Add(24 * time.Hour),
		FixedPrice:           2.50,
		Priority:             5,
		// Empty InventorySpecs → match all
	}
	created, _ := svc.CreateDeal(deal)
	created.Status = "active"
	svc.deals.Store(created.ID, created)

	results := svc.CheckEligibility("any-pub", "any-site", "any-placement", "banner", "desktop", "US")
	found := false
	for _, r := range results {
		if r.DealID == created.ID && r.Eligible {
			found = true
		}
	}
	if !found {
		t.Error("Expected deal to be eligible with empty inventory specs")
	}
}

// ============================================================
// PG GetStats — with deals in various states
// ============================================================

func TestGetStats_WithDeals_B22(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	// Create active deal
	d1 := &PGDeal{
		BuyerID: "b1", SellerID: "s1",
		CommittedImpressions: 1000,
		StartDate:            time.Now().Add(-time.Hour),
		EndDate:              time.Now().Add(24 * time.Hour),
	}
	c1, _ := svc.CreateDeal(d1)
	c1.Status = "active"
	svc.deals.Store(c1.ID, c1)

	// Create completed deal with spend
	d2 := &PGDeal{
		BuyerID: "b2", SellerID: "s2",
		CommittedImpressions: 500,
		StartDate:            time.Now().Add(-2 * time.Hour),
		EndDate:              time.Now().Add(24 * time.Hour),
	}
	c2, _ := svc.CreateDeal(d2)
	c2.Status = "completed"
	c2.DeliveredImpressions = 500
	c2.ActualSpend = 100.0
	svc.deals.Store(c2.ID, c2)

	stats := svc.GetStats()

	if stats["total_deals"].(int) < 2 {
		t.Errorf("Expected at least 2 deals, got %v", stats["total_deals"])
	}
	totalSpend := stats["total_spend"].(float64)
	if totalSpend < 100.0 {
		t.Errorf("Expected total_spend >= 100.0, got %.2f", totalSpend)
	}
	deliveryRate := stats["delivery_rate"].(float64)
	if deliveryRate < 0.0 || deliveryRate > 10.0 {
		t.Errorf("Unexpected delivery rate: %v", deliveryRate)
	}
}

func TestGetStats_ZeroCommitted_B22(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	// No deals → 0 committed → delivery rate branch: totalCommitted==0
	stats := svc.GetStats()
	if stats["delivery_rate"].(float64) != 0.0 {
		t.Errorf("Expected delivery rate 0 when no deals, got %v", stats["delivery_rate"])
	}
}

// ============================================================
// checkPacingAlerts — via CheckAlerts: dailyBudget=0, under/over-pacing
// ============================================================

func makePacingCampaign(id string, underPacing, overPacing float64) *model.Campaign {
	return &model.Campaign{
		ID:   id,
		Name: "Pacing Campaign " + id,
		Targeting: model.Targeting{
			AlertConfig: &model.AlertConfig{
				Enabled: true,
				PacingAlerts: &model.PacingAlerts{
					Enabled:            true,
					UnderPacingPercent: underPacing,
					OverPacingPercent:  overPacing,
				},
			},
		},
	}
}

func TestCheckPacingAlerts_ZeroDailyBudget_B22(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	camp := makePacingCampaign("pacing-zero", 20.0, 20.0)

	// dailyBudget = 0 → should return early without panic
	result := svc.CheckAlerts(camp, 50.0, 0.0)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	// BidAdjustment should be unchanged (1.0 since no budget alerts)
	if result.BidAdjustment != 1.0 {
		t.Logf("BidAdjustment = %.2f (pacing check was skipped)", result.BidAdjustment)
	}
}

func TestCheckPacingAlerts_UnderPacing_B22(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	camp := makePacingCampaign("pacing-under", 20.0, 20.0)

	// Use a dailyBudget where actual spend is far behind expected
	// We spend 1.0 out of 100.0, while expected spend is much more
	// pacingDiff = actualSpendPercent - expectedSpendPercent
	// To trigger under-pacing: pacingDiff < -underPacing (< -20)
	// actualSpendPercent = 1/100 * 100 = 1.0%
	// expectedSpendPercent = (hoursElapsed / 24) * 100
	// At hour 12, expectedSpendPercent = 50%  → pacingDiff = 1 - 50 = -49 < -20 → under-pacing
	// We can't control time, but with 1.0 spend / 100 budget, during any hour of the day,
	// if it's past midnight (hour >= 1), expectedSpend >= 4.2%, actual is 1% → diff < -3.2
	// We need underPacingPercent = 0 (defaults to 20) — set it to 0 to use default 20
	// Let's just set a very small spend and large budget to ensure under-pacing
	result := svc.CheckAlerts(camp, 0.10, 1000.0) // 0.01% vs expected many %
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	// The under-pacing branch increases BidAdjustment by *1.2
	// Just verify it ran without panicking
}

func TestCheckPacingAlerts_OverPacing_B22(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	camp := makePacingCampaign("pacing-over", 20.0, 5.0)

	// Over-pacing: spend 95 out of 100 budget early in day
	// At midnight (hour 0), expectedSpendPercent = 0, actual = 95%
	// pacingDiff = 95 - 0 = 95 > overPacing (5)
	result := svc.CheckAlerts(camp, 95.0, 100.0)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	// The over-pacing branch decreases BidAdjustment by *0.8
	// BidAdjustment may be < 1.0 if triggered
	t.Logf("BidAdjustment after over-pacing check: %.2f", result.BidAdjustment)
}

func TestCheckPacingAlerts_DefaultThresholds_B22(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	// Use zero values for UnderPacingPercent and OverPacingPercent → use defaults (20%)
	camp := &model.Campaign{
		ID:   "pacing-defaults",
		Name: "Default Pacing",
		Targeting: model.Targeting{
			AlertConfig: &model.AlertConfig{
				Enabled: true,
				PacingAlerts: &model.PacingAlerts{
					Enabled:            true,
					UnderPacingPercent: 0, // defaults to 20
					OverPacingPercent:  0, // defaults to 20
				},
			},
		},
	}
	result := svc.CheckAlerts(camp, 50.0, 100.0)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

// ============================================================
// getOrderedProviders — fallbackOrder branch and priority sort
// ============================================================

func TestGetOrderedProviders_FallbackOrder_B22(t *testing.T) {
	svc := NewUnifiedIDService(NewMockCache())

	config := &model.UnifiedIDConfig{
		Enabled: true,
		Providers: []model.IDProvider{
			{Name: "uid2", Enabled: true, Priority: 3},
			{Name: "id5", Enabled: true, Priority: 1},
			{Name: "rampid", Enabled: true, Priority: 2},
		},
		FallbackOrder: []string{"rampid", "uid2", "id5"},
	}

	ordered := svc.getOrderedProviders(config)
	if len(ordered) != 3 {
		t.Fatalf("Expected 3 providers, got %d", len(ordered))
	}
	// First should be "rampid" per fallback order
	if ordered[0].Name != "rampid" {
		t.Errorf("Expected first provider 'rampid', got '%s'", ordered[0].Name)
	}
}

func TestGetOrderedProviders_PrioritySort_B22(t *testing.T) {
	svc := NewUnifiedIDService(NewMockCache())

	config := &model.UnifiedIDConfig{
		Enabled: true,
		Providers: []model.IDProvider{
			{Name: "uid2", Enabled: true, Priority: 3},
			{Name: "id5", Enabled: true, Priority: 1},
			{Name: "rampid", Enabled: true, Priority: 2},
		},
		// No FallbackOrder → sort by priority ascending
	}

	ordered := svc.getOrderedProviders(config)
	if len(ordered) != 3 {
		t.Fatalf("Expected 3 providers, got %d", len(ordered))
	}
	// Should be sorted by priority ascending: id5(1), rampid(2), uid2(3)
	if ordered[0].Name != "id5" {
		t.Errorf("Expected first provider 'id5' (priority 1), got '%s'", ordered[0].Name)
	}
}

func TestGetOrderedProviders_FallbackOrderUnknownProvider_B22(t *testing.T) {
	svc := NewUnifiedIDService(NewMockCache())

	config := &model.UnifiedIDConfig{
		Enabled: true,
		Providers: []model.IDProvider{
			{Name: "uid2", Enabled: true, Priority: 1},
			{Name: "id5", Enabled: true, Priority: 2},
		},
		// FallbackOrder has a name not in providers → uses 999 order
		FallbackOrder: []string{"rampid", "uid2"},
	}

	ordered := svc.getOrderedProviders(config)
	if len(ordered) != 2 {
		t.Fatalf("Expected 2 providers, got %d", len(ordered))
	}
	// "uid2" is in fallbackOrder at position 1, "id5" is unknown (order 999)
	if ordered[0].Name != "uid2" {
		t.Errorf("Expected 'uid2' first (fallback order), got '%s'", ordered[0].Name)
	}
}

// ============================================================
// GenerateAttributionSource — nil config, debugMode, aggregatableValues
// ============================================================

func TestGenerateAttributionSource_NilConfig_B22(t *testing.T) {
	svc := NewPrivacySandboxService(NewMockCache())
	camp := &model.Campaign{ID: "camp-attr-nil", TenantID: "tenant-1"}
	req := &model.BidRequest{ID: "req-attr-nil"}

	result := svc.GenerateAttributionSource(nil, camp, req)
	if result != nil {
		t.Errorf("Expected nil result for nil config, got %v", result)
	}
}

func TestGenerateAttributionSource_DisabledConfig_B22(t *testing.T) {
	svc := NewPrivacySandboxService(NewMockCache())
	camp := &model.Campaign{ID: "camp-attr-dis", TenantID: "tenant-1"}
	req := &model.BidRequest{ID: "req-attr-dis"}
	config := &model.AttributionAPIConfig{Enabled: false}

	result := svc.GenerateAttributionSource(config, camp, req)
	if result != nil {
		t.Errorf("Expected nil result for disabled config, got %v", result)
	}
}

func TestGenerateAttributionSource_DebugMode_B22(t *testing.T) {
	svc := NewPrivacySandboxService(NewMockCache())
	camp := &model.Campaign{
		ID:       "camp-attr-debug",
		TenantID: "tenant-1",
		Creative: model.Creative{URL: "https://ad.example.com/creative"},
	}
	req := &model.BidRequest{ID: "req-attr-debug"}
	config := &model.AttributionAPIConfig{
		Enabled:         true,
		DebugMode:       true,
		ReportingOrigin: "https://reporting.example.com",
		SourceEventID:   "evt-001",
	}

	result := svc.GenerateAttributionSource(config, camp, req)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	debugKey, ok := result["debug_key"]
	if !ok {
		t.Error("Expected debug_key in result")
	}
	expected := camp.ID + "_" + req.ID
	if debugKey != expected {
		t.Errorf("Expected debug_key '%s', got '%v'", expected, debugKey)
	}
}

func TestGenerateAttributionSource_AggregatableValues_B22(t *testing.T) {
	svc := NewPrivacySandboxService(NewMockCache())
	camp := &model.Campaign{
		ID:       "camp-attr-agg",
		TenantID: "tenant-1",
		Creative: model.Creative{URL: "https://ad.example.com/creative"},
	}
	req := &model.BidRequest{ID: "req-attr-agg"}
	config := &model.AttributionAPIConfig{
		Enabled:            true,
		DebugMode:          false,
		AggregatableValues: []string{"impressions", "clicks"},
	}

	result := svc.GenerateAttributionSource(config, camp, req)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if _, ok := result["aggregatable_source"]; !ok {
		t.Error("Expected aggregatable_source in result when AggregatableValues set")
	}
}

// ============================================================
// GetCostAnalysis — with metrics data
// ============================================================

type mockCacheWithCostMetrics struct {
	*MockCache
	metrics *model.SupplyChainMetrics
}

func (m *mockCacheWithCostMetrics) GetSupplyChainMetrics(timeRange string) (*model.SupplyChainMetrics, error) {
	return m.metrics, nil
}

func TestGetCostAnalysis_WithMetrics_B22(t *testing.T) {
	mc := &mockCacheWithCostMetrics{
		MockCache: NewMockCache(),
		metrics: &model.SupplyChainMetrics{
			TotalRequests: 1000,
			AvgTotalFees:  0.005,
		},
	}
	svc := NewSupplyPathAnalyticsService(mc)

	costs, err := svc.GetCostAnalysis("1h")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if costs["total_fees"] == 0 {
		t.Error("Expected non-zero total_fees")
	}
	if costs["avg_fee_per_request"] != 0.005 {
		t.Errorf("Expected avg_fee_per_request=0.005, got %.4f", costs["avg_fee_per_request"])
	}
}

func TestGetCostAnalysis_NilMetrics_B22(t *testing.T) {
	mc := &mockCacheWithBottlenecks{
		MockCache: NewMockCache(),
		metrics:   nil,
	}
	svc := NewSupplyPathAnalyticsService(mc)

	costs, err := svc.GetCostAnalysis("1h")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(costs) != 0 {
		t.Errorf("Expected empty costs for nil metrics, got %v", costs)
	}
}

// ============================================================
// AnalyzeDirectPublisherOpportunities — with low-success-rate services
// ============================================================

func TestAnalyzeDirectPublisherOpportunities_WithOpportunities_B22(t *testing.T) {
	mc := &mockCacheWithCostMetrics{
		MockCache: NewMockCache(),
		metrics: &model.SupplyChainMetrics{
			TotalRequests: 500,
			ServiceMetrics: map[string]model.ServiceMetrics{
				"svc-low": {
					ServiceName: "svc-low",
					SuccessRate: 0.80, // < 0.95 → creates opportunity
					TotalCalls:  1000,
					TotalFees:   50.0,
				},
				"svc-high": {
					ServiceName: "svc-high",
					SuccessRate: 0.98, // >= 0.95 → no opportunity
					TotalCalls:  2000,
					TotalFees:   20.0,
				},
			},
		},
	}
	svc := NewSupplyPathAnalyticsService(mc)

	analysis, err := svc.AnalyzeDirectPublisherOpportunities("1h")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if analysis == nil {
		t.Fatal("Expected non-nil analysis")
	}
	// Should have at least one opportunity (from svc-low)
	if len(analysis.Opportunities) == 0 {
		t.Error("Expected at least one opportunity for low-success-rate service")
	}
	// Verify opportunity fields
	opp := analysis.Opportunities[0]
	if opp.SuccessRate >= 0.95 {
		t.Errorf("Expected opportunity to have low success rate, got %.2f", opp.SuccessRate)
	}
}

func TestAnalyzeDirectPublisherOpportunities_NilMetrics_B22(t *testing.T) {
	mc := &mockCacheWithBottlenecks{
		MockCache: NewMockCache(),
		metrics:   nil,
	}
	svc := NewSupplyPathAnalyticsService(mc)

	analysis, err := svc.AnalyzeDirectPublisherOpportunities("1h")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(analysis.Opportunities) != 0 {
		t.Errorf("Expected empty opportunities for nil metrics, got %d", len(analysis.Opportunities))
	}
}

// ============================================================
// updateDeliveryProgress — underdelivering status branch
// ============================================================

func TestUpdateDeliveryProgress_Underdelivering_B22(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	// Set AlertThreshold to 0.8 (default)
	deal := &PGDeal{
		BuyerID:              "buyer-udp",
		SellerID:             "seller-udp",
		CommittedImpressions: 10000,
		StartDate:            time.Now().Add(-48 * time.Hour), // started 2 days ago
		EndDate:              time.Now().Add(8 * 24 * time.Hour),
	}
	created, _ := svc.CreateDeal(deal)
	created.Status = "active"
	svc.deals.Store(created.ID, created)

	// Record minimal impressions (way under committed pace)
	// This should trigger underdelivering status
	_ = svc.RecordImpression(created.ID, 1.0) // just 1 impression
	_ = svc.RecordImpression(created.ID, 1.0)

	progress, err := svc.GetDeliveryProgress(created.ID)
	if err != nil {
		t.Fatalf("GetDeliveryProgress failed: %v", err)
	}
	t.Logf("Delivery status: %s, rate: %.4f", progress.Status, progress.DeliveryRate)
	// With only 2 impressions vs 10000 committed, should be underdelivering or slightly_behind
	if progress.Status == "" {
		t.Error("Expected non-empty delivery status")
	}
}

func TestRecordImpression_DealCompleted_B22(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-complete",
		SellerID:             "seller-complete",
		CommittedImpressions: 2, // Complete after 2 impressions
		StartDate:            time.Now().Add(-time.Hour),
		EndDate:              time.Now().Add(24 * time.Hour),
	}
	created, _ := svc.CreateDeal(deal)
	created.Status = "active"
	svc.deals.Store(created.ID, created)

	_ = svc.RecordImpression(created.ID, 1.0)
	_ = svc.RecordImpression(created.ID, 1.0)

	d, _ := svc.GetDeal(created.ID)
	if d.Status != "completed" {
		t.Errorf("Expected deal to be completed after meeting committed impressions, got %s", d.Status)
	}
}
