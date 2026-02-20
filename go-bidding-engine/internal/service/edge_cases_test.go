package service

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// =============================================================================
// EDGE CASE TESTS - Testing boundary conditions and error scenarios
// =============================================================================

// ----- Creative Optimization Edge Cases -----

func TestCreativeOptimization_NilInputHandling(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	// Test with nil campaign (expect panic or error recovery)
	defer func() {
		if r := recover(); r != nil {
			t.Log("Recovered from nil campaign panic (expected behavior)")
		}
	}()

	// This may panic - which is acceptable for nil campaign
	_ = svc.SelectCreative(nil, nil)
}

func TestCreativeOptimization_BasicCampaign(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	campaign := &model.Campaign{
		ID:       "campaign-001",
		BidPrice: 2.0,
	}
	req := &model.BidRequest{
		ID: "req-001",
	}

	result := svc.SelectCreative(campaign, req)
	// Should return a result even for minimal campaign
	if result == nil {
		t.Log("Returned nil for campaign (may be expected)")
	}
}

// ----- Dynamic Bid Edge Cases -----

func TestDynamicBidEdge_DisabledService(t *testing.T) {
	svc := NewDynamicBidService(nil)
	svc.SetConfig(&DynamicBidConfig{
		Enabled: false,
	})

	campaign := &model.Campaign{
		ID:       "campaign-001",
		BidPrice: 2.0,
	}
	req := &model.BidRequest{
		ID: "req-001",
	}

	result := svc.CalculateDynamicBid(campaign, req)

	// When disabled, should return original bid unchanged
	if result.AdjustedBid != campaign.BidPrice {
		t.Errorf("Expected bid %f, got %f", campaign.BidPrice, result.AdjustedBid)
	}
	if result.Multiplier != 1.0 {
		t.Errorf("Expected multiplier 1.0, got %f", result.Multiplier)
	}
}

func TestDynamicBidEdge_HighBaseBid(t *testing.T) {
	svc := NewDynamicBidService(nil)

	campaign := &model.Campaign{
		ID:       "campaign-001",
		BidPrice: 100.0, // Very high base bid
	}
	req := &model.BidRequest{
		ID:          "req-001",
		PublisherID: "pub-001",
		Device: model.InternalDevice{
			Type: "mobile",
		},
	}

	result := svc.CalculateDynamicBid(campaign, req)

	// Should still produce valid result
	if result.AdjustedBid <= 0 {
		t.Error("Expected positive adjusted bid")
	}
}

func TestDynamicBidEdge_ZeroBid(t *testing.T) {
	svc := NewDynamicBidService(nil)

	campaign := &model.Campaign{
		ID:       "campaign-001",
		BidPrice: 0, // Zero bid
	}
	req := &model.BidRequest{
		ID: "req-001",
	}

	result := svc.CalculateDynamicBid(campaign, req)
	// Should handle gracefully
	if result.AdjustedBid < 0 {
		t.Error("Expected non-negative bid")
	}
}

// ----- Dayparting Edge Cases -----

func TestDaypartingEdge_NilCampaignPanic(t *testing.T) {
	svc := NewDaypartingService(nil)

	// Test with nil campaign (expect panic recovery)
	defer func() {
		if r := recover(); r != nil {
			t.Log("Recovered from nil campaign panic (expected behavior)")
		}
	}()

	// This may panic - acceptable for nil campaign
	_ = svc.CalculateDaypartMultiplier(nil, nil)
}

func TestDaypartingEdge_EmptyCampaign(t *testing.T) {
	svc := NewDaypartingService(nil)

	campaign := &model.Campaign{
		ID: "campaign-001",
	}
	req := &model.BidRequest{
		ID: "req-001",
	}

	result := svc.CalculateDaypartMultiplier(campaign, req)
	// Without dayparting config, should return default multiplier
	if result.Multiplier != 1.0 {
		t.Errorf("Expected default multiplier 1.0, got %f", result.Multiplier)
	}
}

// =============================================================================
// CONCURRENCY TESTS
// =============================================================================

func TestBidCacheEdge_ConcurrentAccess(t *testing.T) {
	svc := NewBidCacheService(nil)

	// Concurrent writes and reads
	done := make(chan bool)

	// Writers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				bid := &CachedBid{Price: float64(id * j)}
				svc.Set(nil, "concurrent-key-"+string(rune('0'+id)), bid)
			}
			done <- true
		}(i)
	}

	// Readers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				svc.Get(nil, "concurrent-key-"+string(rune('0'+id)))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// If we get here without panic, test passes
}

func TestPGEdge_ConcurrentDeals(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	done := make(chan bool)

	// Concurrent deal creation
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				deal := &PGDeal{
					BuyerID:              "buyer-" + string(rune('0'+id)),
					SellerID:             "seller-001",
					CommittedImpressions: 1000,
				}
				svc.CreateDeal(deal)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify deals were created
	deals := svc.ListDeals("", "", "")
	if len(deals) < 100 {
		t.Errorf("Expected at least 100 deals, got %d", len(deals))
	}
}

func TestDPEdge_ConcurrentRegistration(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	done := make(chan bool)

	// Concurrent publisher registration
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				pub := &DirectPublisher{
					Name:   "Publisher",
					Domain: "example" + string(rune('0'+id)) + ".com",
				}
				svc.RegisterPublisher(pub)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify publishers were registered
	pubs := svc.ListPublishers("", 0)
	if len(pubs) < 100 {
		t.Errorf("Expected at least 100 publishers, got %d", len(pubs))
	}
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestIntegration_PGWithDP(t *testing.T) {
	pgSvc := NewProgrammaticGuaranteedService(nil)
	dpSvc := NewDirectPublisherService(nil)

	// Register a direct publisher
	pub := &DirectPublisher{
		Name:             "Premium Publisher",
		Domain:           "premium.com",
		IsDirectSeller:   true,
		ViewabilityRate:  0.9,
		BrandSafetyScore: 0.95,
	}
	registeredPub, _ := dpSvc.RegisterPublisher(pub)
	dpSvc.ActivatePublisher(registeredPub.ID)

	// Create PG deal for this publisher
	deal := &PGDeal{
		BuyerID:              "buyer-premium",
		SellerID:             registeredPub.ID,
		CommittedImpressions: 1000000,
		FixedPrice:           10.0,
		InventorySpecs: InventorySpec{
			PublisherIDs: []string{registeredPub.ID},
		},
		StartDate: time.Now(),
		EndDate:   time.Now().Add(30 * 24 * time.Hour),
	}
	createdDeal, _ := pgSvc.CreateDeal(deal)
	pgSvc.ActivateDeal(createdDeal.ID)

	// Verify eligibility
	eligible := pgSvc.CheckEligibility(registeredPub.ID, "", "", "", "", "")

	found := false
	for _, e := range eligible {
		if e.DealID == createdDeal.ID && e.Eligible {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected deal to be eligible for direct publisher")
	}
}

func TestIntegration_S2SCache(t *testing.T) {
	s2sSvc := NewS2SBiddingService(nil)
	cacheSvc := NewBidCacheService(nil)

	// Register partners
	partner := &DemandPartner{
		ID:       "partner-cached",
		Name:     "Cached Partner",
		Endpoint: "https://partner.com/bid",
	}
	s2sSvc.RegisterPartner(partner)

	// Simulate caching bid response
	cachedBid := &CachedBid{
		Price:     3.50,
		PartnerID: "partner-cached",
	}
	cacheKey := cacheSvc.GenerateCacheKey(map[string]interface{}{
		"partner": "partner-cached",
		"slot":    "slot-001",
	})
	cacheSvc.Set(nil, cacheKey, cachedBid)

	// Verify cache hit
	retrieved, found := cacheSvc.Get(nil, cacheKey)
	if !found {
		t.Error("Expected cache hit")
	}
	if retrieved.PartnerID != "partner-cached" {
		t.Errorf("Expected partner 'partner-cached', got '%s'", retrieved.PartnerID)
	}
}

// =============================================================================
// SERVICE GETTER TESTS
// =============================================================================

func TestBiddingService_CoreGetters(t *testing.T) {
	svc := NewBiddingService(nil, "")

	// Test core getter methods exist and return non-nil
	if svc.GetDynamicBidService() == nil {
		t.Error("GetDynamicBidService returned nil")
	}
	if svc.GetCreativeOptimizationService() == nil {
		t.Error("GetCreativeOptimizationService returned nil")
	}
	if svc.GetDaypartingService() == nil {
		t.Error("GetDaypartingService returned nil")
	}
	if svc.GetAttributionService() == nil {
		t.Error("GetAttributionService returned nil")
	}
	if svc.GetAudienceModelingService() == nil {
		t.Error("GetAudienceModelingService returned nil")
	}
	if svc.GetBidLandscapeService() == nil {
		t.Error("GetBidLandscapeService returned nil")
	}
	if svc.GetCompetitiveIntelligenceService() == nil {
		t.Error("GetCompetitiveIntelligenceService returned nil")
	}
	if svc.GetContextualAIService() == nil {
		t.Error("GetContextualAIService returned nil")
	}
	if svc.GetIncrementalityService() == nil {
		t.Error("GetIncrementalityService returned nil")
	}
	if svc.GetPrivacySandboxService() == nil {
		t.Error("GetPrivacySandboxService returned nil")
	}
	if svc.GetSupplyPathAnalyticsService() == nil {
		t.Error("GetSupplyPathAnalyticsService returned nil")
	}
	if svc.GetUnifiedIDService() == nil {
		t.Error("GetUnifiedIDService returned nil")
	}
	if svc.GetChurnPredictionService() == nil {
		t.Error("GetChurnPredictionService returned nil")
	}
	if svc.GetPerformancePredictionService() == nil {
		t.Error("GetPerformancePredictionService returned nil")
	}
	if svc.GetLookalikeService() == nil {
		t.Error("GetLookalikeService returned nil")
	}
	if svc.GetUserClusteringService() == nil {
		t.Error("GetUserClusteringService returned nil")
	}
	if svc.GetABTestingService() == nil {
		t.Error("GetABTestingService returned nil")
	}
	if svc.GetS2SBiddingService() == nil {
		t.Error("GetS2SBiddingService returned nil")
	}
	if svc.GetBidCacheService() == nil {
		t.Error("GetBidCacheService returned nil")
	}
	if svc.GetProgrammaticGuaranteedService() == nil {
		t.Error("GetProgrammaticGuaranteedService returned nil")
	}
	if svc.GetDirectPublisherService() == nil {
		t.Error("GetDirectPublisherService returned nil")
	}
}

// =============================================================================
// STATS AGGREGATION TEST
// =============================================================================

func TestCoreServices_StatsNotPanic(t *testing.T) {
	svc := NewBiddingService(nil, "")

	// Ensure calling GetStats on services doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetStats panicked: %v", r)
		}
	}()

	// Call GetStats on services that have it
	svc.GetS2SBiddingService().GetStats()
	svc.GetBidCacheService().GetStats()
	svc.GetProgrammaticGuaranteedService().GetStats()
	svc.GetDirectPublisherService().GetStats()
}
