package service

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// BidCacheService provides intelligent bid caching strategies
// to reduce latency and improve performance
type BidCacheService struct {
	mu              sync.RWMutex
	lruCache        *LRUCache
	impressionCache map[string]*CachedBid
	partnerCache    map[string]map[string]*CachedBid // partner -> key -> bid
	config          *BidCacheConfig
	stats           *BidCacheStats
}

// BidCacheConfig holds cache configuration
type BidCacheConfig struct {
	MaxEntries         int           `json:"max_entries"`
	TTL                time.Duration `json:"ttl"`
	ImpressionCacheTTL time.Duration `json:"impression_cache_ttl"`
	EnablePartnerCache bool          `json:"enable_partner_cache"`
	EnablePredictive   bool          `json:"enable_predictive"`
	EvictionPolicy     string        `json:"eviction_policy"` // "lru", "lfu", "ttl"
	WarmupEnabled      bool          `json:"warmup_enabled"`
	StaleServeEnabled  bool          `json:"stale_serve_enabled"`
	StaleServeTTL      time.Duration `json:"stale_serve_ttl"`
}

// CachedBid represents a cached bid response
type CachedBid struct {
	Key         string          `json:"key"`
	BidResponse json.RawMessage `json:"bid_response"`
	Price       float64         `json:"price"`
	PartnerID   string          `json:"partner_id"`
	CampaignID  string          `json:"campaign_id"`
	CreativeID  string          `json:"creative_id"`
	CreatedAt   time.Time       `json:"created_at"`
	ExpiresAt   time.Time       `json:"expires_at"`
	HitCount    int64           `json:"hit_count"`
	LastAccess  time.Time       `json:"last_access"`
	Stale       bool            `json:"stale"`
}

// BidCacheStats tracks cache performance metrics
type BidCacheStats struct {
	Hits         int64     `json:"hits"`
	Misses       int64     `json:"misses"`
	Evictions    int64     `json:"evictions"`
	Expirations  int64     `json:"expirations"`
	StaleServes  int64     `json:"stale_serves"`
	TotalEntries int64     `json:"total_entries"`
	MemoryUsage  int64     `json:"memory_usage"`
	AvgLatency   float64   `json:"avg_latency_ms"`
	LastReset    time.Time `json:"last_reset"`
}

// LRUCache implements Least Recently Used cache
type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex
}

// cacheEntry for LRU list
type cacheEntry struct {
	key   string
	value *CachedBid
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Get retrieves a value from LRU cache
func (c *LRUCache) Get(key string) (*CachedBid, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		c.list.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		entry.value.HitCount++
		entry.value.LastAccess = time.Now()
		return entry.value, true
	}
	return nil, false
}

// Put adds or updates a value in LRU cache
func (c *LRUCache) Put(key string, value *CachedBid) (evicted bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		c.list.MoveToFront(elem)
		elem.Value.(*cacheEntry).value = value
		return false
	}

	if c.list.Len() >= c.capacity {
		c.evictOldest()
		evicted = true
	}

	entry := &cacheEntry{key: key, value: value}
	elem := c.list.PushFront(entry)
	c.cache[key] = elem
	return evicted
}

// evictOldest removes the oldest entry
func (c *LRUCache) evictOldest() {
	elem := c.list.Back()
	if elem != nil {
		entry := c.list.Remove(elem).(*cacheEntry)
		delete(c.cache, entry.key)
	}
}

// Remove deletes a key from cache
func (c *LRUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		c.list.Remove(elem)
		delete(c.cache, key)
	}
}

// Len returns the number of entries
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list.Len()
}

// Clear removes all entries
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*list.Element)
	c.list.Init()
}

// NewBidCacheService creates a new bid cache service
func NewBidCacheService(config *BidCacheConfig) *BidCacheService {
	if config == nil {
		config = &BidCacheConfig{
			MaxEntries:         10000,
			TTL:                30 * time.Second,
			ImpressionCacheTTL: 60 * time.Second,
			EnablePartnerCache: true,
			EnablePredictive:   true,
			EvictionPolicy:     "lru",
			WarmupEnabled:      true,
			StaleServeEnabled:  true,
			StaleServeTTL:      5 * time.Second,
		}
	}

	return &BidCacheService{
		lruCache:        NewLRUCache(config.MaxEntries),
		impressionCache: make(map[string]*CachedBid),
		partnerCache:    make(map[string]map[string]*CachedBid),
		config:          config,
		stats: &BidCacheStats{
			LastReset: time.Now(),
		},
	}
}

// GenerateCacheKey creates a unique cache key from request parameters
func (s *BidCacheService) GenerateCacheKey(params map[string]interface{}) string {
	// Sort and serialize parameters for consistent key generation
	data, _ := json.Marshal(params)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:16])
}

// Get retrieves a cached bid
func (s *BidCacheService) Get(ctx context.Context, key string) (*CachedBid, bool) {
	start := time.Now()

	// Try LRU cache first
	if cached, ok := s.lruCache.Get(key); ok {
		if time.Now().Before(cached.ExpiresAt) {
			s.recordHit(start)
			return cached, true
		}

		// Check if stale serve is enabled
		if s.config.StaleServeEnabled && time.Now().Before(cached.ExpiresAt.Add(s.config.StaleServeTTL)) {
			cached.Stale = true
			s.mu.Lock()
			s.stats.StaleServes++
			s.mu.Unlock()
			s.recordHit(start)
			return cached, true
		}

		// Expired, remove from cache
		s.lruCache.Remove(key)
		s.mu.Lock()
		s.stats.Expirations++
		s.mu.Unlock()
	}

	s.recordMiss(start)
	return nil, false
}

// Set stores a bid in cache
func (s *BidCacheService) Set(ctx context.Context, key string, bid *CachedBid) {
	bid.Key = key
	bid.CreatedAt = time.Now()
	if bid.ExpiresAt.IsZero() {
		bid.ExpiresAt = time.Now().Add(s.config.TTL)
	}
	bid.LastAccess = time.Now()

	evicted := s.lruCache.Put(key, bid)
	if evicted {
		s.mu.Lock()
		s.stats.Evictions++
		s.mu.Unlock()
	}

	s.mu.Lock()
	s.stats.TotalEntries = int64(s.lruCache.Len())
	s.mu.Unlock()
}

// SetWithTTL stores a bid with custom TTL
func (s *BidCacheService) SetWithTTL(ctx context.Context, key string, bid *CachedBid, ttl time.Duration) {
	bid.ExpiresAt = time.Now().Add(ttl)
	s.Set(ctx, key, bid)
}

// CacheImpression caches bid for a specific impression
func (s *BidCacheService) CacheImpression(impressionID string, bid *CachedBid) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bid.ExpiresAt = time.Now().Add(s.config.ImpressionCacheTTL)
	s.impressionCache[impressionID] = bid
}

// GetImpressionBid retrieves cached bid for impression
func (s *BidCacheService) GetImpressionBid(impressionID string) (*CachedBid, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if bid, ok := s.impressionCache[impressionID]; ok {
		if time.Now().Before(bid.ExpiresAt) {
			return bid, true
		}
	}
	return nil, false
}

// CachePartnerBid caches bid from specific partner
func (s *BidCacheService) CachePartnerBid(partnerID, key string, bid *CachedBid) {
	if !s.config.EnablePartnerCache {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.partnerCache[partnerID]; !ok {
		s.partnerCache[partnerID] = make(map[string]*CachedBid)
	}

	bid.PartnerID = partnerID
	bid.ExpiresAt = time.Now().Add(s.config.TTL)
	s.partnerCache[partnerID][key] = bid
}

// GetPartnerBid retrieves cached bid from specific partner
func (s *BidCacheService) GetPartnerBid(partnerID, key string) (*CachedBid, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if partnerBids, ok := s.partnerCache[partnerID]; ok {
		if bid, ok := partnerBids[key]; ok {
			if time.Now().Before(bid.ExpiresAt) {
				return bid, true
			}
		}
	}
	return nil, false
}

// WarmCache pre-populates cache with predicted bids
func (s *BidCacheService) WarmCache(ctx context.Context, bids []*CachedBid) int {
	if !s.config.WarmupEnabled {
		return 0
	}

	warmed := 0
	for _, bid := range bids {
		if bid.Key != "" {
			s.Set(ctx, bid.Key, bid)
			warmed++
		}
	}
	return warmed
}

// Invalidate removes a specific key from cache
func (s *BidCacheService) Invalidate(key string) {
	s.lruCache.Remove(key)
}

// InvalidatePartner removes all cached bids for a partner
func (s *BidCacheService) InvalidatePartner(partnerID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if partnerBids, ok := s.partnerCache[partnerID]; ok {
		count := len(partnerBids)
		delete(s.partnerCache, partnerID)
		return count
	}
	return 0
}

// InvalidateCampaign removes all cached bids for a campaign
func (s *BidCacheService) InvalidateCampaign(campaignID string) int {
	// For LRU cache, we need to iterate and remove matching entries
	// This is expensive but necessary for cache consistency
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear partner cache entries for campaign
	count := 0
	for _, partnerBids := range s.partnerCache {
		for key, bid := range partnerBids {
			if bid.CampaignID == campaignID {
				delete(partnerBids, key)
				count++
			}
		}
	}
	return count
}

// CleanExpired removes expired entries from cache
func (s *BidCacheService) CleanExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleaned := 0
	now := time.Now()

	// Clean impression cache
	for id, bid := range s.impressionCache {
		if now.After(bid.ExpiresAt) {
			delete(s.impressionCache, id)
			cleaned++
		}
	}

	// Clean partner cache
	for _, partnerBids := range s.partnerCache {
		for key, bid := range partnerBids {
			if now.After(bid.ExpiresAt) {
				delete(partnerBids, key)
				cleaned++
			}
		}
	}

	s.stats.Expirations += int64(cleaned)
	return cleaned
}

// GetStats returns cache statistics
func (s *BidCacheService) GetStats() *BidCacheStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := *s.stats
	stats.TotalEntries = int64(s.lruCache.Len())

	// Calculate hit rate
	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.MemoryUsage = stats.TotalEntries * 500 // Approximate bytes per entry
	}

	return &stats
}

// GetHitRate returns cache hit rate
func (s *BidCacheService) GetHitRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := s.stats.Hits + s.stats.Misses
	if total == 0 {
		return 0
	}
	return float64(s.stats.Hits) / float64(total)
}

// ResetStats resets cache statistics
func (s *BidCacheService) ResetStats() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats = &BidCacheStats{
		LastReset: time.Now(),
	}
}

// Clear removes all entries from cache
func (s *BidCacheService) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lruCache.Clear()
	s.impressionCache = make(map[string]*CachedBid)
	s.partnerCache = make(map[string]map[string]*CachedBid)
	s.stats.TotalEntries = 0
}

// recordHit updates hit statistics
func (s *BidCacheService) recordHit(start time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.Hits++
	latency := float64(time.Since(start).Microseconds()) / 1000.0
	s.stats.AvgLatency = (s.stats.AvgLatency*float64(s.stats.Hits-1) + latency) / float64(s.stats.Hits)
}

// recordMiss updates miss statistics
func (s *BidCacheService) recordMiss(start time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.Misses++
}

// GetConfig returns current cache configuration
func (s *BidCacheService) GetConfig() *BidCacheConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	config := *s.config
	return &config
}

// UpdateConfig updates cache configuration
func (s *BidCacheService) UpdateConfig(config *BidCacheConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// Size returns the current number of entries in cache
func (s *BidCacheService) Size() int {
	return s.lruCache.Len()
}
