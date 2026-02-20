package service

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestLRU_NewCache(t *testing.T) {
	cache := NewLRUCache(100)
	if cache == nil {
		t.Fatal("expected cache")
	}
	if cache.capacity != 100 {
		t.Errorf("expected capacity 100, got %d", cache.capacity)
	}
	if cache.Len() != 0 {
		t.Errorf("expected empty cache, got %d entries", cache.Len())
	}
}

func TestLRU_PutAndGet(t *testing.T) {
	cache := NewLRUCache(10)

	bid := &CachedBid{
		Key:   "test-key",
		Price: 2.50,
	}

	evicted := cache.Put("key1", bid)

	if evicted {
		t.Error("unexpected eviction on first put")
	}

	retrieved, ok := cache.Get("key1")
	if !ok {
		t.Fatal("expected to find key1")
	}
	if retrieved.Price != 2.50 {
		t.Errorf("expected price 2.50, got %f", retrieved.Price)
	}
}

func TestLRU_GetNotFound(t *testing.T) {
	cache := NewLRUCache(10)

	_, ok := cache.Get("nonexistent")

	if ok {
		t.Error("expected not found for nonexistent key")
	}
}

func TestLRU_Eviction(t *testing.T) {
	cache := NewLRUCache(3)

	cache.Put("key1", &CachedBid{Price: 1.0})
	cache.Put("key2", &CachedBid{Price: 2.0})
	cache.Put("key3", &CachedBid{Price: 3.0})

	// This should evict key1
	evicted := cache.Put("key4", &CachedBid{Price: 4.0})

	if !evicted {
		t.Error("expected eviction")
	}
	if cache.Len() != 3 {
		t.Errorf("expected 3 entries, got %d", cache.Len())
	}

	// key1 should be evicted
	_, ok := cache.Get("key1")
	if ok {
		t.Error("key1 should have been evicted")
	}

	// key4 should exist
	_, ok = cache.Get("key4")
	if !ok {
		t.Error("key4 should exist")
	}
}

func TestLRU_MoveToFront(t *testing.T) {
	cache := NewLRUCache(3)

	cache.Put("key1", &CachedBid{Price: 1.0})
	cache.Put("key2", &CachedBid{Price: 2.0})
	cache.Put("key3", &CachedBid{Price: 3.0})

	// Access key1, moving it to front
	cache.Get("key1")

	// Now add key4, which should evict key2 (oldest after key1 was accessed)
	cache.Put("key4", &CachedBid{Price: 4.0})

	_, ok := cache.Get("key1")
	if !ok {
		t.Error("key1 should still exist")
	}

	_, ok = cache.Get("key2")
	if ok {
		t.Error("key2 should have been evicted")
	}
}

func TestLRU_Remove(t *testing.T) {
	cache := NewLRUCache(10)

	cache.Put("key1", &CachedBid{Price: 1.0})
	cache.Put("key2", &CachedBid{Price: 2.0})

	cache.Remove("key1")

	if cache.Len() != 1 {
		t.Errorf("expected 1 entry, got %d", cache.Len())
	}

	_, ok := cache.Get("key1")
	if ok {
		t.Error("key1 should have been removed")
	}
}

func TestLRU_Clear(t *testing.T) {
	cache := NewLRUCache(10)

	cache.Put("key1", &CachedBid{Price: 1.0})
	cache.Put("key2", &CachedBid{Price: 2.0})
	cache.Put("key3", &CachedBid{Price: 3.0})

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("expected empty cache, got %d entries", cache.Len())
	}
}

func TestLRU_HitCount(t *testing.T) {
	cache := NewLRUCache(10)

	cache.Put("key1", &CachedBid{Price: 1.0})

	cache.Get("key1")
	cache.Get("key1")
	cache.Get("key1")

	bid, _ := cache.Get("key1")
	if bid.HitCount != 4 {
		t.Errorf("expected hit count 4, got %d", bid.HitCount)
	}
}

func TestBidCache_NewService(t *testing.T) {
	svc := NewBidCacheService(nil)
	if svc == nil {
		t.Fatal("expected service")
	}
	if svc.config.MaxEntries != 10000 {
		t.Errorf("expected default max entries 10000, got %d", svc.config.MaxEntries)
	}
	if svc.config.TTL != 30*time.Second {
		t.Errorf("expected default TTL 30s, got %v", svc.config.TTL)
	}
}

func TestBidCache_NewServiceWithConfig(t *testing.T) {
	config := &BidCacheConfig{
		MaxEntries: 5000,
		TTL:        60 * time.Second,
	}

	svc := NewBidCacheService(config)

	if svc.config.MaxEntries != 5000 {
		t.Errorf("expected max entries 5000, got %d", svc.config.MaxEntries)
	}
}

func TestBidCache_GenerateCacheKey(t *testing.T) {
	svc := NewBidCacheService(nil)

	params1 := map[string]interface{}{
		"publisher": "pub-1",
		"slot":      "slot-1",
	}
	params2 := map[string]interface{}{
		"publisher": "pub-1",
		"slot":      "slot-1",
	}
	params3 := map[string]interface{}{
		"publisher": "pub-2",
		"slot":      "slot-1",
	}

	key1 := svc.GenerateCacheKey(params1)
	key2 := svc.GenerateCacheKey(params2)
	key3 := svc.GenerateCacheKey(params3)

	if key1 != key2 {
		t.Error("same params should produce same key")
	}
	if key1 == key3 {
		t.Error("different params should produce different key")
	}
	if len(key1) != 32 {
		t.Errorf("expected key length 32, got %d", len(key1))
	}
}

func TestBidCache_SetAndGet(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	bid := &CachedBid{
		Price:      3.50,
		CampaignID: "camp-1",
	}

	svc.Set(ctx, "test-key", bid)

	retrieved, ok := svc.Get(ctx, "test-key")
	if !ok {
		t.Fatal("expected to find cached bid")
	}
	if retrieved.Price != 3.50 {
		t.Errorf("expected price 3.50, got %f", retrieved.Price)
	}
}

func TestBidCache_GetMiss(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	_, ok := svc.Get(ctx, "nonexistent")

	if ok {
		t.Error("expected cache miss")
	}

	stats := svc.GetStats()
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
}

func TestBidCache_SetWithTTL(t *testing.T) {
	config := &BidCacheConfig{
		MaxEntries:        1000,
		TTL:               30 * time.Second,
		StaleServeEnabled: false, // Disable stale serve for this test
	}
	svc := NewBidCacheService(config)
	ctx := context.Background()

	bid := &CachedBid{Price: 2.0}
	svc.SetWithTTL(ctx, "key1", bid, 100*time.Millisecond)

	// Should exist immediately
	_, ok := svc.Get(ctx, "key1")
	if !ok {
		t.Error("expected to find key1")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, ok = svc.Get(ctx, "key1")
	if ok {
		t.Error("expected key1 to be expired")
	}
}

func TestBidCache_StaleServe(t *testing.T) {
	config := &BidCacheConfig{
		MaxEntries:        1000,
		TTL:               50 * time.Millisecond,
		StaleServeEnabled: true,
		StaleServeTTL:     100 * time.Millisecond,
	}
	svc := NewBidCacheService(config)
	ctx := context.Background()

	bid := &CachedBid{Price: 2.0}
	svc.Set(ctx, "key1", bid)

	// Wait for TTL to expire but within stale serve window
	time.Sleep(75 * time.Millisecond)

	retrieved, ok := svc.Get(ctx, "key1")
	if !ok {
		t.Fatal("expected stale serve")
	}
	if !retrieved.Stale {
		t.Error("expected bid to be marked stale")
	}

	stats := svc.GetStats()
	if stats.StaleServes != 1 {
		t.Errorf("expected 1 stale serve, got %d", stats.StaleServes)
	}
}

func TestBidCache_ImpressionCache(t *testing.T) {
	svc := NewBidCacheService(nil)

	bid := &CachedBid{Price: 1.50, CreativeID: "creative-1"}
	svc.CacheImpression("imp-123", bid)

	retrieved, ok := svc.GetImpressionBid("imp-123")
	if !ok {
		t.Fatal("expected to find impression bid")
	}
	if retrieved.CreativeID != "creative-1" {
		t.Errorf("expected creative-1, got %s", retrieved.CreativeID)
	}
}

func TestBidCache_ImpressionCacheExpired(t *testing.T) {
	config := &BidCacheConfig{
		ImpressionCacheTTL: 50 * time.Millisecond,
	}
	svc := NewBidCacheService(config)

	bid := &CachedBid{Price: 1.50}
	svc.CacheImpression("imp-123", bid)

	time.Sleep(75 * time.Millisecond)

	_, ok := svc.GetImpressionBid("imp-123")
	if ok {
		t.Error("expected impression bid to be expired")
	}
}

func TestBidCache_PartnerCache(t *testing.T) {
	config := &BidCacheConfig{
		EnablePartnerCache: true,
		TTL:                30 * time.Second,
	}
	svc := NewBidCacheService(config)

	bid := &CachedBid{Price: 2.00, CampaignID: "camp-1"}
	svc.CachePartnerBid("partner-1", "key-1", bid)

	retrieved, ok := svc.GetPartnerBid("partner-1", "key-1")
	if !ok {
		t.Fatal("expected to find partner bid")
	}
	if retrieved.PartnerID != "partner-1" {
		t.Errorf("expected partner-1, got %s", retrieved.PartnerID)
	}
}

func TestBidCache_PartnerCacheDisabled(t *testing.T) {
	config := &BidCacheConfig{
		EnablePartnerCache: false,
	}
	svc := NewBidCacheService(config)

	bid := &CachedBid{Price: 2.00}
	svc.CachePartnerBid("partner-1", "key-1", bid)

	_, ok := svc.GetPartnerBid("partner-1", "key-1")
	if ok {
		t.Error("expected partner cache to be disabled")
	}
}

func TestBidCache_WarmCache(t *testing.T) {
	config := &BidCacheConfig{
		WarmupEnabled: true,
		TTL:           30 * time.Second,
		MaxEntries:    1000,
	}
	svc := NewBidCacheService(config)
	ctx := context.Background()

	bids := []*CachedBid{
		{Key: "key-1", Price: 1.0},
		{Key: "key-2", Price: 2.0},
		{Key: "key-3", Price: 3.0},
		{Key: "", Price: 4.0}, // Empty key should be skipped
	}

	warmed := svc.WarmCache(ctx, bids)

	if warmed != 3 {
		t.Errorf("expected 3 warmed entries, got %d", warmed)
	}

	_, ok := svc.Get(ctx, "key-1")
	if !ok {
		t.Error("expected to find warmed key-1")
	}
}

func TestBidCache_WarmCacheDisabled(t *testing.T) {
	config := &BidCacheConfig{
		WarmupEnabled: false,
	}
	svc := NewBidCacheService(config)
	ctx := context.Background()

	bids := []*CachedBid{
		{Key: "key-1", Price: 1.0},
	}

	warmed := svc.WarmCache(ctx, bids)

	if warmed != 0 {
		t.Errorf("expected 0 warmed when disabled, got %d", warmed)
	}
}

func TestBidCache_Invalidate(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	svc.Set(ctx, "key-1", &CachedBid{Price: 1.0})
	svc.Set(ctx, "key-2", &CachedBid{Price: 2.0})

	svc.Invalidate("key-1")

	_, ok := svc.Get(ctx, "key-1")
	if ok {
		t.Error("key-1 should have been invalidated")
	}

	_, ok = svc.Get(ctx, "key-2")
	if !ok {
		t.Error("key-2 should still exist")
	}
}

func TestBidCache_InvalidatePartner(t *testing.T) {
	config := &BidCacheConfig{
		EnablePartnerCache: true,
		TTL:                30 * time.Second,
	}
	svc := NewBidCacheService(config)

	svc.CachePartnerBid("partner-1", "key-1", &CachedBid{Price: 1.0})
	svc.CachePartnerBid("partner-1", "key-2", &CachedBid{Price: 2.0})
	svc.CachePartnerBid("partner-2", "key-3", &CachedBid{Price: 3.0})

	count := svc.InvalidatePartner("partner-1")

	if count != 2 {
		t.Errorf("expected 2 invalidated, got %d", count)
	}

	_, ok := svc.GetPartnerBid("partner-1", "key-1")
	if ok {
		t.Error("partner-1 bids should be invalidated")
	}

	_, ok = svc.GetPartnerBid("partner-2", "key-3")
	if !ok {
		t.Error("partner-2 bids should still exist")
	}
}

func TestBidCache_InvalidateCampaign(t *testing.T) {
	config := &BidCacheConfig{
		EnablePartnerCache: true,
		TTL:                30 * time.Second,
	}
	svc := NewBidCacheService(config)

	svc.CachePartnerBid("partner-1", "key-1", &CachedBid{Price: 1.0, CampaignID: "camp-1"})
	svc.CachePartnerBid("partner-1", "key-2", &CachedBid{Price: 2.0, CampaignID: "camp-2"})
	svc.CachePartnerBid("partner-2", "key-3", &CachedBid{Price: 3.0, CampaignID: "camp-1"})

	count := svc.InvalidateCampaign("camp-1")

	if count != 2 {
		t.Errorf("expected 2 invalidated, got %d", count)
	}
}

func TestBidCache_CleanExpired(t *testing.T) {
	config := &BidCacheConfig{
		EnablePartnerCache: true,
		TTL:                50 * time.Millisecond,
		ImpressionCacheTTL: 50 * time.Millisecond,
	}
	svc := NewBidCacheService(config)

	svc.CacheImpression("imp-1", &CachedBid{Price: 1.0})
	svc.CachePartnerBid("partner-1", "key-1", &CachedBid{Price: 2.0})

	time.Sleep(75 * time.Millisecond)

	cleaned := svc.CleanExpired()

	if cleaned != 2 {
		t.Errorf("expected 2 cleaned, got %d", cleaned)
	}
}

func TestBidCache_GetStats(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	svc.Set(ctx, "key-1", &CachedBid{Price: 1.0})
	svc.Get(ctx, "key-1") // Hit
	svc.Get(ctx, "key-2") // Miss

	stats := svc.GetStats()

	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
	if stats.TotalEntries != 1 {
		t.Errorf("expected 1 entry, got %d", stats.TotalEntries)
	}
}

func TestBidCache_GetHitRate(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	// 0 hits, 0 misses
	if svc.GetHitRate() != 0 {
		t.Error("expected 0 hit rate initially")
	}

	svc.Set(ctx, "key-1", &CachedBid{Price: 1.0})
	svc.Get(ctx, "key-1") // Hit
	svc.Get(ctx, "key-1") // Hit
	svc.Get(ctx, "key-2") // Miss
	svc.Get(ctx, "key-3") // Miss

	rate := svc.GetHitRate()
	if rate != 0.5 {
		t.Errorf("expected hit rate 0.5, got %f", rate)
	}
}

func TestBidCache_ResetStats(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	svc.Set(ctx, "key-1", &CachedBid{Price: 1.0})
	svc.Get(ctx, "key-1")
	svc.Get(ctx, "key-2")

	svc.ResetStats()

	stats := svc.GetStats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("stats should be reset")
	}
}

func TestBidCache_Clear(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	svc.Set(ctx, "key-1", &CachedBid{Price: 1.0})
	svc.Set(ctx, "key-2", &CachedBid{Price: 2.0})
	svc.CacheImpression("imp-1", &CachedBid{Price: 1.5})

	svc.Clear()

	if svc.Size() != 0 {
		t.Errorf("expected empty cache, got %d entries", svc.Size())
	}
}

func TestBidCache_GetConfig(t *testing.T) {
	config := &BidCacheConfig{
		MaxEntries: 5000,
		TTL:        60 * time.Second,
	}
	svc := NewBidCacheService(config)

	retrieved := svc.GetConfig()

	if retrieved.MaxEntries != 5000 {
		t.Errorf("expected max entries 5000, got %d", retrieved.MaxEntries)
	}
}

func TestBidCache_UpdateConfig(t *testing.T) {
	svc := NewBidCacheService(nil)

	newConfig := &BidCacheConfig{
		MaxEntries: 20000,
		TTL:        120 * time.Second,
	}

	svc.UpdateConfig(newConfig)

	config := svc.GetConfig()
	if config.MaxEntries != 20000 {
		t.Errorf("expected max entries 20000, got %d", config.MaxEntries)
	}
}

func TestBidCache_Size(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	if svc.Size() != 0 {
		t.Error("expected size 0 initially")
	}

	svc.Set(ctx, "key-1", &CachedBid{Price: 1.0})
	svc.Set(ctx, "key-2", &CachedBid{Price: 2.0})

	if svc.Size() != 2 {
		t.Errorf("expected size 2, got %d", svc.Size())
	}
}

func TestBidCache_Concurrency(t *testing.T) {
	svc := NewBidCacheService(nil)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := "key-" + string(rune(idx%10))
			bid := &CachedBid{Price: float64(idx)}
			svc.Set(ctx, key, bid)
			svc.Get(ctx, key)
			svc.Get(ctx, "nonexistent")
		}(i)
	}
	wg.Wait()

	stats := svc.GetStats()
	if stats.Hits+stats.Misses != 200 {
		t.Errorf("expected 200 total operations, got %d", stats.Hits+stats.Misses)
	}
}

func TestLRU_Concurrency(t *testing.T) {
	cache := NewLRUCache(100)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := "key-" + string(rune(idx))
			cache.Put(key, &CachedBid{Price: float64(idx)})
			cache.Get(key)
			cache.Remove(key)
			cache.Put(key, &CachedBid{Price: float64(idx * 2)})
		}(i)
	}
	wg.Wait()

	if cache.Len() > 50 {
		t.Errorf("unexpected cache size: %d", cache.Len())
	}
}
