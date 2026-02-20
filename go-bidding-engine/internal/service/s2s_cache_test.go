package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestS2SBiddingService_RegisterPartner(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	partner := &DemandPartner{
		ID:       "partner-1",
		Name:     "Test Partner",
		Endpoint: "https://partner1.example.com/bid",
		BidFloor: 0.50,
		QPS:      100,
	}

	err := svc.RegisterPartner(partner)
	if err != nil {
		t.Fatalf("RegisterPartner failed: %v", err)
	}

	// Verify partner was registered
	retrieved, err := svc.GetPartner("partner-1")
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

func TestS2SBiddingService_RegisterPartnerValidation(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	// Test missing ID
	err := svc.RegisterPartner(&DemandPartner{
		Endpoint: "https://example.com",
	})
	if err == nil {
		t.Error("Expected error for missing ID")
	}

	// Test missing endpoint
	err = svc.RegisterPartner(&DemandPartner{
		ID: "partner-1",
	})
	if err == nil {
		t.Error("Expected error for missing endpoint")
	}
}

func TestS2SBiddingService_UpdatePartner(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	// Register partner
	partner := &DemandPartner{
		ID:       "partner-1",
		Name:     "Original Name",
		Endpoint: "https://example.com/bid",
	}
	svc.RegisterPartner(partner)

	// Update partner
	updated := &DemandPartner{
		ID:       "partner-1",
		Name:     "Updated Name",
		Endpoint: "https://new-endpoint.com/bid",
		QPS:      200,
	}
	err := svc.UpdatePartner(updated)
	if err != nil {
		t.Fatalf("UpdatePartner failed: %v", err)
	}

	// Verify update
	retrieved, _ := svc.GetPartner("partner-1")
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected updated name, got '%s'", retrieved.Name)
	}
}

func TestS2SBiddingService_RemovePartner(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	// Register partner
	svc.RegisterPartner(&DemandPartner{
		ID:       "partner-1",
		Name:     "Test Partner",
		Endpoint: "https://example.com/bid",
	})

	// Remove partner
	err := svc.RemovePartner("partner-1")
	if err != nil {
		t.Fatalf("RemovePartner failed: %v", err)
	}

	// Verify removal
	_, err = svc.GetPartner("partner-1")
	if err == nil {
		t.Error("Expected error for removed partner")
	}
}

func TestS2SBiddingService_ListPartners(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	// Register multiple partners
	for i := 1; i <= 3; i++ {
		svc.RegisterPartner(&DemandPartner{
			ID:       string(rune('a' + i - 1)),
			Name:     "Partner",
			Endpoint: "https://example.com/bid",
		})
	}

	partners := svc.ListPartners()
	if len(partners) != 3 {
		t.Errorf("Expected 3 partners, got %d", len(partners))
	}
}

func TestS2SBiddingService_ProcessBidRequest(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	// Register partners
	svc.RegisterPartner(&DemandPartner{
		ID:       "partner-1",
		Name:     "Partner 1",
		Endpoint: "https://partner1.com/bid",
		BidFloor: 0.50,
	})
	svc.RegisterPartner(&DemandPartner{
		ID:       "partner-2",
		Name:     "Partner 2",
		Endpoint: "https://partner2.com/bid",
		BidFloor: 0.75,
	})

	// Create bid request
	req := &S2SBidRequest{
		ID: "test-request-1",
		Imp: []S2SImpression{
			{
				ID: "imp-1",
				Banner: &S2SBanner{
					W: 300,
					H: 250,
				},
				BidFloor: 0.25,
			},
		},
		Site: &S2SSite{
			ID:     "site-1",
			Domain: "example.com",
		},
		Timeout: 100,
	}

	ctx := context.Background()
	resp, err := svc.ProcessBidRequest(ctx, req)
	if err != nil {
		t.Fatalf("ProcessBidRequest failed: %v", err)
	}

	if resp.ID != "test-request-1" {
		t.Errorf("Response ID mismatch: expected 'test-request-1', got '%s'", resp.ID)
	}

	if len(resp.SeatBid) == 0 {
		t.Error("Expected at least one seat bid")
	}

	if len(resp.PartnerBids) != 2 {
		t.Errorf("Expected 2 partner bids, got %d", len(resp.PartnerBids))
	}
}

func TestS2SBiddingService_ProcessBidRequest_NoPartners(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	req := &S2SBidRequest{
		ID: "test-request",
		Imp: []S2SImpression{
			{ID: "imp-1"},
		},
	}

	_, err := svc.ProcessBidRequest(context.Background(), req)
	if err == nil {
		t.Error("Expected error when no partners available")
	}
}

func TestS2SBiddingService_SelectWinningBid(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	// Register partners for stats tracking
	svc.RegisterPartner(&DemandPartner{
		ID:       "partner-1",
		Name:     "Partner 1",
		Endpoint: "https://example.com",
	})

	response := &S2SBidResponse{
		ID: "resp-1",
		SeatBid: []S2SSeatBid{
			{
				Seat: "partner-1",
				Bid: []S2SBid{
					{ID: "bid-1", ImpID: "imp-1", Price: 1.50},
				},
			},
			{
				Seat: "partner-2",
				Bid: []S2SBid{
					{ID: "bid-2", ImpID: "imp-1", Price: 2.00},
				},
			},
		},
	}

	winningBid, winningSeat, err := svc.SelectWinningBid(response)
	if err != nil {
		t.Fatalf("SelectWinningBid failed: %v", err)
	}

	if winningBid.Price != 2.00 {
		t.Errorf("Expected winning price 2.00, got %.2f", winningBid.Price)
	}
	if winningSeat != "partner-2" {
		t.Errorf("Expected winner 'partner-2', got '%s'", winningSeat)
	}
}

func TestS2SBiddingService_SelectWinningBid_NoBids(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	response := &S2SBidResponse{
		ID:      "resp-1",
		SeatBid: []S2SSeatBid{},
	}

	_, _, err := svc.SelectWinningBid(response)
	if err == nil {
		t.Error("Expected error when no bids")
	}
}

func TestS2SBiddingService_EnableDisablePartner(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	svc.RegisterPartner(&DemandPartner{
		ID:       "partner-1",
		Name:     "Partner 1",
		Endpoint: "https://example.com",
	})

	// Disable partner
	err := svc.DisablePartner("partner-1")
	if err != nil {
		t.Fatalf("DisablePartner failed: %v", err)
	}

	partner, _ := svc.GetPartner("partner-1")
	if partner.Enabled {
		t.Error("Partner should be disabled")
	}

	// Enable partner
	err = svc.EnablePartner("partner-1")
	if err != nil {
		t.Fatalf("EnablePartner failed: %v", err)
	}

	partner, _ = svc.GetPartner("partner-1")
	if !partner.Enabled {
		t.Error("Partner should be enabled")
	}
}

func TestS2SBiddingService_GetStats(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	// Register and process some requests
	svc.RegisterPartner(&DemandPartner{
		ID:       "partner-1",
		Endpoint: "https://example.com",
	})

	stats := svc.GetStats()
	if stats["partner_count"].(int) != 1 {
		t.Error("Stats should show 1 partner")
	}
}

// Bid Cache Tests

func TestBidCacheService_SetAndGet(t *testing.T) {
	svc := NewBidCacheService(nil)

	key := "test-key-1"
	bid := &CachedBid{
		Price:       1.50,
		CampaignID:  "campaign-1",
		BidResponse: json.RawMessage(`{"price": 1.50}`),
	}

	ctx := context.Background()
	svc.Set(ctx, key, bid)

	// Get the cached bid
	cached, found := svc.Get(ctx, key)
	if !found {
		t.Fatal("Expected to find cached bid")
	}

	if cached.Price != 1.50 {
		t.Errorf("Expected price 1.50, got %.2f", cached.Price)
	}
}

func TestBidCacheService_TTLExpiration(t *testing.T) {
	config := &BidCacheConfig{
		MaxEntries:        1000,
		TTL:               50 * time.Millisecond,
		StaleServeEnabled: false,
	}
	svc := NewBidCacheService(config)

	key := "expiring-key"
	bid := &CachedBid{Price: 1.00}

	ctx := context.Background()
	svc.Set(ctx, key, bid)

	// Should be found initially
	_, found := svc.Get(ctx, key)
	if !found {
		t.Fatal("Expected to find bid before expiration")
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should not be found after expiration
	_, found = svc.Get(ctx, key)
	if found {
		t.Error("Expected bid to be expired")
	}
}

func TestBidCacheService_StaleServe(t *testing.T) {
	config := &BidCacheConfig{
		MaxEntries:        1000,
		TTL:               50 * time.Millisecond,
		StaleServeEnabled: true,
		StaleServeTTL:     100 * time.Millisecond,
	}
	svc := NewBidCacheService(config)

	key := "stale-key"
	bid := &CachedBid{Price: 1.00}

	ctx := context.Background()
	svc.Set(ctx, key, bid)

	// Wait past TTL but within stale TTL
	time.Sleep(60 * time.Millisecond)

	// Should still be served (stale)
	cached, found := svc.Get(ctx, key)
	if !found {
		t.Fatal("Expected stale serve to return bid")
	}
	if !cached.Stale {
		t.Error("Expected bid to be marked as stale")
	}
}

func TestBidCacheService_GenerateCacheKey(t *testing.T) {
	svc := NewBidCacheService(nil)

	params := map[string]interface{}{
		"campaign_id": "camp-1",
		"ad_size":     "300x250",
		"country":     "US",
	}

	key1 := svc.GenerateCacheKey(params)
	key2 := svc.GenerateCacheKey(params)

	if key1 != key2 {
		t.Error("Same params should generate same key")
	}

	// Different params should generate different key
	params["country"] = "UK"
	key3 := svc.GenerateCacheKey(params)
	if key1 == key3 {
		t.Error("Different params should generate different key")
	}
}

func TestBidCacheService_ImpressionCache(t *testing.T) {
	svc := NewBidCacheService(nil)

	impID := "imp-123"
	bid := &CachedBid{
		Price:      2.00,
		CreativeID: "creative-1",
	}

	svc.CacheImpression(impID, bid)

	cached, found := svc.GetImpressionBid(impID)
	if !found {
		t.Fatal("Expected to find impression bid")
	}
	if cached.Price != 2.00 {
		t.Errorf("Expected price 2.00, got %.2f", cached.Price)
	}
}

func TestBidCacheService_PartnerCache(t *testing.T) {
	config := &BidCacheConfig{
		MaxEntries:         1000,
		TTL:                30 * time.Second,
		EnablePartnerCache: true,
	}
	svc := NewBidCacheService(config)

	partnerID := "partner-1"
	key := "partner-key-1"
	bid := &CachedBid{
		Price:      1.75,
		PartnerID:  partnerID,
		CampaignID: "camp-1",
	}

	svc.CachePartnerBid(partnerID, key, bid)

	cached, found := svc.GetPartnerBid(partnerID, key)
	if !found {
		t.Fatal("Expected to find partner bid")
	}
	if cached.Price != 1.75 {
		t.Errorf("Expected price 1.75, got %.2f", cached.Price)
	}
}

func TestBidCacheService_WarmCache(t *testing.T) {
	config := &BidCacheConfig{
		MaxEntries:    1000,
		TTL:           30 * time.Second,
		WarmupEnabled: true,
	}
	svc := NewBidCacheService(config)

	bids := []*CachedBid{
		{Key: "warm-1", Price: 1.00},
		{Key: "warm-2", Price: 1.50},
		{Key: "warm-3", Price: 2.00},
	}

	warmed := svc.WarmCache(context.Background(), bids)
	if warmed != 3 {
		t.Errorf("Expected 3 warmed entries, got %d", warmed)
	}

	// Verify entries are in cache
	if svc.Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", svc.Size())
	}
}

func TestBidCacheService_Invalidate(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	svc.Set(ctx, "key-1", &CachedBid{Price: 1.00})
	svc.Set(ctx, "key-2", &CachedBid{Price: 2.00})

	svc.Invalidate("key-1")

	_, found := svc.Get(ctx, "key-1")
	if found {
		t.Error("Invalidated key should not be found")
	}

	_, found = svc.Get(ctx, "key-2")
	if !found {
		t.Error("Non-invalidated key should still be found")
	}
}

func TestBidCacheService_InvalidatePartner(t *testing.T) {
	config := &BidCacheConfig{
		MaxEntries:         1000,
		TTL:                30 * time.Second,
		EnablePartnerCache: true,
	}
	svc := NewBidCacheService(config)

	svc.CachePartnerBid("partner-1", "k1", &CachedBid{Price: 1.00})
	svc.CachePartnerBid("partner-1", "k2", &CachedBid{Price: 1.50})
	svc.CachePartnerBid("partner-2", "k3", &CachedBid{Price: 2.00})

	removed := svc.InvalidatePartner("partner-1")
	if removed != 2 {
		t.Errorf("Expected 2 removed, got %d", removed)
	}

	// Partner 2 bids should still exist
	_, found := svc.GetPartnerBid("partner-2", "k3")
	if !found {
		t.Error("Partner 2 bid should still exist")
	}
}

func TestBidCacheService_CleanExpired(t *testing.T) {
	config := &BidCacheConfig{
		MaxEntries:         1000,
		TTL:                30 * time.Second,
		ImpressionCacheTTL: 10 * time.Millisecond,
	}
	svc := NewBidCacheService(config)

	svc.CacheImpression("imp-1", &CachedBid{Price: 1.00})
	svc.CacheImpression("imp-2", &CachedBid{Price: 2.00})

	time.Sleep(20 * time.Millisecond)

	cleaned := svc.CleanExpired()
	if cleaned < 2 {
		t.Errorf("Expected at least 2 cleaned, got %d", cleaned)
	}
}

func TestBidCacheService_GetStats(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	// Generate some hits and misses
	svc.Set(ctx, "key-1", &CachedBid{Price: 1.00})
	svc.Get(ctx, "key-1") // hit
	svc.Get(ctx, "key-2") // miss

	stats := svc.GetStats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
}

func TestBidCacheService_GetHitRate(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	svc.Set(ctx, "key-1", &CachedBid{Price: 1.00})
	svc.Get(ctx, "key-1") // hit
	svc.Get(ctx, "key-1") // hit
	svc.Get(ctx, "key-2") // miss

	hitRate := svc.GetHitRate()
	expected := 2.0 / 3.0
	if hitRate < expected-0.01 || hitRate > expected+0.01 {
		t.Errorf("Expected hit rate ~%.2f, got %.2f", expected, hitRate)
	}
}

func TestBidCacheService_Clear(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	svc.Set(ctx, "key-1", &CachedBid{Price: 1.00})
	svc.Set(ctx, "key-2", &CachedBid{Price: 2.00})
	svc.CacheImpression("imp-1", &CachedBid{Price: 1.50})

	svc.Clear()

	if svc.Size() != 0 {
		t.Errorf("Expected empty cache after clear, got %d entries", svc.Size())
	}
}

func TestBidCacheService_LRUEviction(t *testing.T) {
	config := &BidCacheConfig{
		MaxEntries: 3,
		TTL:        30 * time.Second,
	}
	svc := NewBidCacheService(config)
	ctx := context.Background()

	// Fill cache to capacity
	svc.Set(ctx, "key-1", &CachedBid{Price: 1.00})
	svc.Set(ctx, "key-2", &CachedBid{Price: 2.00})
	svc.Set(ctx, "key-3", &CachedBid{Price: 3.00})

	// Access key-1 to make it recently used
	svc.Get(ctx, "key-1")

	// Add new entry, should evict key-2 (least recently used)
	svc.Set(ctx, "key-4", &CachedBid{Price: 4.00})

	// key-2 should be evicted
	_, found := svc.Get(ctx, "key-2")
	if found {
		t.Error("LRU eviction should have removed key-2")
	}

	// key-1 should still exist
	_, found = svc.Get(ctx, "key-1")
	if !found {
		t.Error("key-1 should still exist after LRU eviction")
	}
}

func TestBidCacheService_ResetStats(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	svc.Set(ctx, "key-1", &CachedBid{Price: 1.00})
	svc.Get(ctx, "key-1")

	svc.ResetStats()

	stats := svc.GetStats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("Stats should be reset to 0")
	}
}
