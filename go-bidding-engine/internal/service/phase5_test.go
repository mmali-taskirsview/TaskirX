package service

import (
	"testing"
	"time"
)

// =============================================================================
// PROGRAMMATIC GUARANTEED SERVICE TESTS
// =============================================================================

func TestNewProgrammaticGuaranteedService(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)
	if svc == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestPGCreateDeal(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		Name:                 "Test PG Deal",
		BuyerID:              "buyer-001",
		SellerID:             "seller-001",
		CommittedImpressions: 1000000,
		FixedPrice:           5.0,
		StartDate:            time.Now(),
		EndDate:              time.Now().Add(30 * 24 * time.Hour),
	}

	result, err := svc.CreateDeal(deal)
	if err != nil {
		t.Fatalf("CreateDeal failed: %v", err)
	}

	if result.ID == "" {
		t.Error("Expected deal ID to be set")
	}
	if result.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", result.Status)
	}
}

func TestPGCreateDealValidation(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	// Missing buyer ID
	deal := &PGDeal{
		SellerID:             "seller-001",
		CommittedImpressions: 1000000,
	}

	_, err := svc.CreateDeal(deal)
	if err == nil {
		t.Error("Expected error for missing buyer_id")
	}

	// Missing commitment
	deal = &PGDeal{
		BuyerID:  "buyer-001",
		SellerID: "seller-001",
	}

	_, err = svc.CreateDeal(deal)
	if err == nil {
		t.Error("Expected error for missing commitment")
	}
}

func TestPGGetDeal(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-001",
		SellerID:             "seller-001",
		CommittedImpressions: 1000000,
	}

	created, _ := svc.CreateDeal(deal)

	retrieved, err := svc.GetDeal(created.ID)
	if err != nil {
		t.Fatalf("GetDeal failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, retrieved.ID)
	}
}

func TestPGActivateDeal(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-001",
		SellerID:             "seller-001",
		CommittedImpressions: 1000000,
		StartDate:            time.Now(),
		EndDate:              time.Now().Add(30 * 24 * time.Hour),
	}

	created, _ := svc.CreateDeal(deal)

	err := svc.ActivateDeal(created.ID)
	if err != nil {
		t.Fatalf("ActivateDeal failed: %v", err)
	}

	retrieved, _ := svc.GetDeal(created.ID)
	if retrieved.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", retrieved.Status)
	}
}

func TestPGPauseDeal(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-001",
		SellerID:             "seller-001",
		CommittedImpressions: 1000000,
		StartDate:            time.Now(),
		EndDate:              time.Now().Add(30 * 24 * time.Hour),
	}

	created, _ := svc.CreateDeal(deal)
	svc.ActivateDeal(created.ID)

	err := svc.PauseDeal(created.ID)
	if err != nil {
		t.Fatalf("PauseDeal failed: %v", err)
	}

	retrieved, _ := svc.GetDeal(created.ID)
	if retrieved.Status != "paused" {
		t.Errorf("Expected status 'paused', got '%s'", retrieved.Status)
	}
}

func TestPGListDeals(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	// Create multiple deals
	for i := 0; i < 5; i++ {
		deal := &PGDeal{
			BuyerID:              "buyer-001",
			SellerID:             "seller-001",
			CommittedImpressions: 1000000,
		}
		svc.CreateDeal(deal)
	}

	deals := svc.ListDeals("", "", "")
	if len(deals) < 5 {
		t.Errorf("Expected at least 5 deals, got %d", len(deals))
	}
}

func TestPGListDealsFiltered(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	// Create deals for different buyers
	deal1 := &PGDeal{BuyerID: "buyer-A", SellerID: "seller-001", CommittedImpressions: 1000}
	deal2 := &PGDeal{BuyerID: "buyer-B", SellerID: "seller-001", CommittedImpressions: 1000}
	svc.CreateDeal(deal1)
	svc.CreateDeal(deal2)

	// Filter by buyer
	deals := svc.ListDeals("buyer-A", "", "")
	for _, d := range deals {
		if d.BuyerID != "buyer-A" {
			t.Errorf("Expected buyer 'buyer-A', got '%s'", d.BuyerID)
		}
	}
}

func TestPGRecordImpression(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-001",
		SellerID:             "seller-001",
		CommittedImpressions: 1000000,
		FixedPrice:           5.0,
		StartDate:            time.Now(),
		EndDate:              time.Now().Add(30 * 24 * time.Hour),
	}

	created, _ := svc.CreateDeal(deal)
	svc.ActivateDeal(created.ID)

	// Record impression
	err := svc.RecordImpression(created.ID, 5.0)
	if err != nil {
		t.Fatalf("RecordImpression failed: %v", err)
	}

	retrieved, _ := svc.GetDeal(created.ID)
	if retrieved.DeliveredImpressions != 1 {
		t.Errorf("Expected 1 impression, got %d", retrieved.DeliveredImpressions)
	}
	if retrieved.ActualSpend != 5.0 {
		t.Errorf("Expected spend 5.0, got %f", retrieved.ActualSpend)
	}
}

func TestPGCheckEligibility(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-001",
		SellerID:             "seller-001",
		CommittedImpressions: 1000000,
		FixedPrice:           5.0,
		StartDate:            time.Now().Add(-1 * time.Hour),
		EndDate:              time.Now().Add(30 * 24 * time.Hour),
		InventorySpecs: InventorySpec{
			PublisherIDs: []string{"pub-001"},
			AdFormats:    []string{"banner"},
		},
	}

	created, _ := svc.CreateDeal(deal)
	svc.ActivateDeal(created.ID)

	// Check eligibility
	eligible := svc.CheckEligibility("pub-001", "", "", "banner", "", "")

	found := false
	for _, e := range eligible {
		if e.DealID == created.ID && e.Eligible {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected deal to be eligible")
	}
}

func TestPGDeliveryProgress(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-001",
		SellerID:             "seller-001",
		CommittedImpressions: 100,
		FixedPrice:           5.0,
		StartDate:            time.Now().Add(-1 * time.Hour),
		EndDate:              time.Now().Add(24 * time.Hour),
	}

	created, _ := svc.CreateDeal(deal)
	svc.ActivateDeal(created.ID)

	// Record some impressions
	for i := 0; i < 10; i++ {
		svc.RecordImpression(created.ID, 5.0)
	}

	progress, err := svc.GetDeliveryProgress(created.ID)
	if err != nil {
		t.Fatalf("GetDeliveryProgress failed: %v", err)
	}

	if progress.DeliveredImpressions != 10 {
		t.Errorf("Expected 10 impressions, got %d", progress.DeliveredImpressions)
	}
}

func TestPGStats(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	// Create some deals
	deal := &PGDeal{
		BuyerID:              "buyer-001",
		SellerID:             "seller-001",
		CommittedImpressions: 1000,
	}
	svc.CreateDeal(deal)

	stats := svc.GetStats()

	if stats["total_deals"].(int) < 1 {
		t.Error("Expected at least 1 deal in stats")
	}
}

// =============================================================================
// DIRECT PUBLISHER SERVICE TESTS
// =============================================================================

func TestNewDirectPublisherService(t *testing.T) {
	svc := NewDirectPublisherService(nil)
	if svc == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestDPRegisterPublisher(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub := &DirectPublisher{
		Name:             "Test Publisher",
		Domain:           "example.com",
		ViewabilityRate:  0.75,
		IVTRate:          0.02,
		BrandSafetyScore: 0.9,
	}

	result, err := svc.RegisterPublisher(pub)
	if err != nil {
		t.Fatalf("RegisterPublisher failed: %v", err)
	}

	if result.ID == "" {
		t.Error("Expected publisher ID to be set")
	}
	if result.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", result.Status)
	}
	if result.QualityScore == 0 {
		t.Error("Expected quality score to be calculated")
	}
}

func TestDPRegisterPublisherValidation(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	// Missing domain
	pub := &DirectPublisher{
		Name: "Test Publisher",
	}

	_, err := svc.RegisterPublisher(pub)
	if err == nil {
		t.Error("Expected error for missing domain")
	}
}

func TestDPGetPublisher(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub := &DirectPublisher{
		Name:   "Test Publisher",
		Domain: "example.com",
	}

	created, _ := svc.RegisterPublisher(pub)

	retrieved, err := svc.GetPublisher(created.ID)
	if err != nil {
		t.Fatalf("GetPublisher failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID '%s', got '%s'", created.ID, retrieved.ID)
	}
}

func TestDPActivatePublisher(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub := &DirectPublisher{
		Name:             "Test Publisher",
		Domain:           "example.com",
		ViewabilityRate:  0.8,
		BrandSafetyScore: 0.9,
		IsDirectSeller:   true,
	}

	created, _ := svc.RegisterPublisher(pub)

	err := svc.ActivatePublisher(created.ID)
	if err != nil {
		t.Fatalf("ActivatePublisher failed: %v", err)
	}

	retrieved, _ := svc.GetPublisher(created.ID)
	if retrieved.Status != "active" {
		t.Errorf("Expected status 'active', got '%s'", retrieved.Status)
	}
}

func TestDPActivatePublisherLowQuality(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub := &DirectPublisher{
		Name:            "Low Quality Publisher",
		Domain:          "lowquality.com",
		ViewabilityRate: 0.1, // Very low
		IVTRate:         0.5, // Very high IVT
	}

	created, _ := svc.RegisterPublisher(pub)

	err := svc.ActivatePublisher(created.ID)
	if err == nil {
		t.Error("Expected error for low quality publisher")
	}
}

func TestDPListPublishers(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	// Create multiple publishers
	for i := 0; i < 3; i++ {
		pub := &DirectPublisher{
			Name:   "Publisher",
			Domain: "example.com",
		}
		svc.RegisterPublisher(pub)
	}

	publishers := svc.ListPublishers("", 0)
	if len(publishers) < 3 {
		t.Errorf("Expected at least 3 publishers, got %d", len(publishers))
	}
}

func TestDPListPublishersFiltered(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	// Create publishers with different quality
	pub1 := &DirectPublisher{Domain: "high.com", ViewabilityRate: 0.9, BrandSafetyScore: 0.9, IsDirectSeller: true}
	pub2 := &DirectPublisher{Domain: "low.com", ViewabilityRate: 0.3}

	created1, _ := svc.RegisterPublisher(pub1)
	svc.RegisterPublisher(pub2)
	svc.ActivatePublisher(created1.ID)

	// Filter by active status
	publishers := svc.ListPublishers("active", 0)
	for _, p := range publishers {
		if p.Status != "active" {
			t.Errorf("Expected status 'active', got '%s'", p.Status)
		}
	}
}

func TestDPAddIntegration(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub := &DirectPublisher{
		Name:   "Test Publisher",
		Domain: "example.com",
	}
	created, _ := svc.RegisterPublisher(pub)

	integration := &PublisherIntegration{
		PublisherID:     created.ID,
		IntegrationType: "api",
		Endpoint:        "https://example.com/api/bid",
	}

	result, err := svc.AddIntegration(integration)
	if err != nil {
		t.Fatalf("AddIntegration failed: %v", err)
	}

	if result.ID == "" {
		t.Error("Expected integration ID to be set")
	}
	if result.Status != "testing" {
		t.Errorf("Expected status 'testing', got '%s'", result.Status)
	}
}

func TestDPAnalyzeSupplyPath(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub := &DirectPublisher{
		Name:   "Test Publisher",
		Domain: "example.com",
		SupplyChain: []SupplyChainNode{
			{ASI: "ssp1.com", SID: "001", Fee: 0.10},
			{ASI: "ssp2.com", SID: "002", Fee: 0.05},
			{ASI: "exchange.com", SID: "003", Fee: 0.15},
		},
		SellerID: "pub-001",
	}
	created, _ := svc.RegisterPublisher(pub)

	result, err := svc.AnalyzeSupplyPath(created.ID)
	if err != nil {
		t.Fatalf("AnalyzeSupplyPath failed: %v", err)
	}

	if result.CurrentFees == 0 {
		t.Error("Expected current fees to be calculated")
	}
	if len(result.RecommendedPath) == 0 {
		t.Error("Expected recommended path")
	}
	if result.FeeSavings < 0 {
		t.Error("Expected non-negative fee savings")
	}
}

func TestDPGetDirectRate(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	// Create mix of direct and indirect publishers
	pub1 := &DirectPublisher{Domain: "direct.com", IsDirectSeller: true, ViewabilityRate: 0.9, BrandSafetyScore: 0.9}
	pub2 := &DirectPublisher{Domain: "indirect.com", IsDirectSeller: false, ViewabilityRate: 0.9, BrandSafetyScore: 0.9}

	created1, _ := svc.RegisterPublisher(pub1)
	created2, _ := svc.RegisterPublisher(pub2)

	svc.ActivatePublisher(created1.ID)
	svc.ActivatePublisher(created2.ID)

	rate := svc.GetDirectRate()
	if rate < 0 || rate > 1 {
		t.Errorf("Expected rate between 0 and 1, got %f", rate)
	}
}

func TestDPRecordPathMetrics(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	metrics := &SupplyPathMetrics{
		PathKey:     "pub-001:ssp1:exchange",
		PublisherID: "pub-001",
		PathLength:  3,
		TotalFees:   0.25,
		AvgLatency:  45.0,
		WinRate:     0.15,
		Impressions: 10000,
		Spend:       500.0,
	}

	svc.RecordPathMetrics(metrics)

	retrieved, err := svc.GetPathMetrics("pub-001:ssp1:exchange")
	if err != nil {
		t.Fatalf("GetPathMetrics failed: %v", err)
	}

	if retrieved.EffectiveCPM == 0 {
		t.Error("Expected effective CPM to be calculated")
	}
	if retrieved.Recommendation == "" && !retrieved.IsOptimal {
		t.Error("Expected recommendation for non-optimal path")
	}
}

func TestDPStats(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub := &DirectPublisher{
		Name:   "Test Publisher",
		Domain: "example.com",
	}
	svc.RegisterPublisher(pub)

	stats := svc.GetStats()

	if stats["total_publishers"].(int) < 1 {
		t.Error("Expected at least 1 publisher in stats")
	}
}

// =============================================================================
// S2S BIDDING SERVICE TESTS
// =============================================================================

func TestNewS2SBiddingService(t *testing.T) {
	svc := NewS2SBiddingService(nil)
	if svc == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestS2SRegisterPartner(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	partner := &DemandPartner{
		ID:       "partner-001",
		Name:     "Test Partner",
		Endpoint: "https://partner.com/bid",
		BidFloor: 0.5,
	}

	err := svc.RegisterPartner(partner)
	if err != nil {
		t.Fatalf("RegisterPartner failed: %v", err)
	}

	retrieved, err := svc.GetPartner("partner-001")
	if err != nil {
		t.Fatalf("GetPartner failed: %v", err)
	}

	if retrieved.Name != "Test Partner" {
		t.Errorf("Expected name 'Test Partner', got '%s'", retrieved.Name)
	}
	if !retrieved.Enabled {
		t.Error("Expected partner to be enabled by default")
	}
}

func TestS2SRegisterPartnerValidation(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	// Missing ID
	partner := &DemandPartner{
		Name:     "Test Partner",
		Endpoint: "https://partner.com/bid",
	}

	err := svc.RegisterPartner(partner)
	if err == nil {
		t.Error("Expected error for missing ID")
	}

	// Missing endpoint
	partner = &DemandPartner{
		ID:   "partner-001",
		Name: "Test Partner",
	}

	err = svc.RegisterPartner(partner)
	if err == nil {
		t.Error("Expected error for missing endpoint")
	}
}

func TestS2SListPartners(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	for i := 0; i < 3; i++ {
		partner := &DemandPartner{
			ID:       "partner-" + string(rune('A'+i)),
			Name:     "Partner",
			Endpoint: "https://partner.com/bid",
		}
		svc.RegisterPartner(partner)
	}

	partners := svc.ListPartners()
	if len(partners) < 3 {
		t.Errorf("Expected at least 3 partners, got %d", len(partners))
	}
}

func TestS2SRemovePartner(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	partner := &DemandPartner{
		ID:       "partner-to-remove",
		Name:     "Test Partner",
		Endpoint: "https://partner.com/bid",
	}
	svc.RegisterPartner(partner)

	err := svc.RemovePartner("partner-to-remove")
	if err != nil {
		t.Fatalf("RemovePartner failed: %v", err)
	}

	_, err = svc.GetPartner("partner-to-remove")
	if err == nil {
		t.Error("Expected error when getting removed partner")
	}
}

func TestS2SEnableDisablePartner(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	partner := &DemandPartner{
		ID:       "partner-toggle",
		Name:     "Test Partner",
		Endpoint: "https://partner.com/bid",
	}
	svc.RegisterPartner(partner)

	// Disable
	err := svc.DisablePartner("partner-toggle")
	if err != nil {
		t.Fatalf("DisablePartner failed: %v", err)
	}

	retrieved, _ := svc.GetPartner("partner-toggle")
	if retrieved.Enabled {
		t.Error("Expected partner to be disabled")
	}

	// Enable
	err = svc.EnablePartner("partner-toggle")
	if err != nil {
		t.Fatalf("EnablePartner failed: %v", err)
	}

	retrieved, _ = svc.GetPartner("partner-toggle")
	if !retrieved.Enabled {
		t.Error("Expected partner to be enabled")
	}
}

func TestS2SStats(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	partner := &DemandPartner{
		ID:       "partner-stats",
		Name:     "Test Partner",
		Endpoint: "https://partner.com/bid",
	}
	svc.RegisterPartner(partner)

	stats := svc.GetStats()

	if stats["partner_count"].(int) < 1 {
		t.Error("Expected at least 1 partner in stats")
	}
}

// =============================================================================
// BID CACHE SERVICE TESTS
// =============================================================================

func TestNewBidCacheService(t *testing.T) {
	svc := NewBidCacheService(nil)
	if svc == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestBidCacheSetGet(t *testing.T) {
	svc := NewBidCacheService(nil)

	bid := &CachedBid{
		Price:      2.50,
		PartnerID:  "partner-001",
		CampaignID: "campaign-001",
	}

	svc.Set(nil, "test-key", bid)

	retrieved, found := svc.Get(nil, "test-key")
	if !found {
		t.Fatal("Expected to find cached bid")
	}

	if retrieved.Price != 2.50 {
		t.Errorf("Expected price 2.50, got %f", retrieved.Price)
	}
}

func TestBidCacheExpiration(t *testing.T) {
	config := &BidCacheConfig{
		MaxEntries: 100,
		TTL:        50 * time.Millisecond,
	}
	svc := NewBidCacheService(config)

	bid := &CachedBid{
		Price: 2.50,
	}

	svc.Set(nil, "expire-key", bid)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	_, found := svc.Get(nil, "expire-key")
	if found {
		t.Error("Expected bid to be expired")
	}
}

func TestBidCacheHitRate(t *testing.T) {
	svc := NewBidCacheService(nil)

	// Set a bid
	bid := &CachedBid{Price: 2.50}
	svc.Set(nil, "hit-key", bid)

	// Hit
	svc.Get(nil, "hit-key")
	svc.Get(nil, "hit-key")

	// Miss
	svc.Get(nil, "miss-key")

	hitRate := svc.GetHitRate()
	if hitRate < 0.5 {
		t.Errorf("Expected hit rate >= 0.5, got %f", hitRate)
	}
}

func TestBidCacheClear(t *testing.T) {
	svc := NewBidCacheService(nil)

	for i := 0; i < 10; i++ {
		bid := &CachedBid{Price: float64(i)}
		svc.Set(nil, "key-"+string(rune('0'+i)), bid)
	}

	if svc.Size() < 10 {
		t.Errorf("Expected at least 10 entries, got %d", svc.Size())
	}

	svc.Clear()

	if svc.Size() != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", svc.Size())
	}
}

func TestBidCacheStats(t *testing.T) {
	svc := NewBidCacheService(nil)

	bid := &CachedBid{Price: 2.50}
	svc.Set(nil, "stats-key", bid)
	svc.Get(nil, "stats-key")

	stats := svc.GetStats()

	if stats.Hits < 1 {
		t.Error("Expected at least 1 hit")
	}
}

func TestBidCacheGenerateKey(t *testing.T) {
	svc := NewBidCacheService(nil)

	params1 := map[string]interface{}{
		"publisher": "pub-001",
		"device":    "mobile",
	}

	params2 := map[string]interface{}{
		"publisher": "pub-001",
		"device":    "mobile",
	}

	params3 := map[string]interface{}{
		"publisher": "pub-002",
		"device":    "mobile",
	}

	key1 := svc.GenerateCacheKey(params1)
	key2 := svc.GenerateCacheKey(params2)
	key3 := svc.GenerateCacheKey(params3)

	if key1 != key2 {
		t.Error("Expected same params to generate same key")
	}

	if key1 == key3 {
		t.Error("Expected different params to generate different keys")
	}
}
