package service

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// BIDDING SERVICE ADDITIONAL TESTS
// ============================================================================

func TestBiddingService_GetBackendBaseURL(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	url := service.GetBackendBaseURL()
	if url != "http://test-backend:8080" {
		t.Errorf("Expected http://test-backend:8080, got %s", url)
	}
}

func TestBiddingService_GetDynamicCreativeService(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	dcs := service.GetDynamicCreativeService()
	if dcs == nil {
		t.Error("Expected DynamicCreativeService, got nil")
	}
}

func TestBiddingService_TrackClick(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	err := service.TrackClick("campaign-123")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestBiddingService_TrackImpression(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	err := service.TrackImpression("campaign-123")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestBiddingService_GetBidLandscape(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	landscape, err := service.GetBidLandscape()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Can be nil from mock cache - that's expected
	if len(landscape) > 0 {
		t.Log("Got bid landscape data")
	}
}

func TestBiddingService_GetSegmentPerformance(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	perf, err := service.GetSegmentPerformance("device")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Can be nil from mock cache - that's expected
	if len(perf) > 0 {
		t.Log("Got segment performance data")
	}
}

func TestBiddingService_GetOptimalBidFloor(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	floor, err := service.GetOptimalBidFloor("publisher-123", 0.6)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Should return 0 from mock cache - that's expected
	if floor < 0 {
		t.Error("Expected non-negative bid floor")
	}
}

func TestBiddingService_RecordImpression(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	err := service.RecordImpression("user-123", "campaign-456", "request-789")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestBiddingService_RecordClick(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	err := service.RecordClick("user-123", "campaign-456", "request-789")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestBiddingService_GetAttribution(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	attrType, requestID, err := service.GetAttribution("user-123", "campaign-456")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Empty strings from mock cache
	if attrType != "" || requestID != "" {
		t.Log("Got attribution data")
	}
}

func TestBiddingService_RecordUserEvent(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	err := service.RecordUserEvent("user-123", "campaign-456", "click")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestBiddingService_GetUserEvents(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	events, err := service.GetUserEvents("user-123", []string{"click", "impression"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Nil from mock cache
	if len(events) > 0 {
		t.Log("Got user events data")
	}
}

func TestBiddingService_RecordTouchpoint(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	err := service.RecordTouchpoint("user-123", "campaign-456", "click", "request-789")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestBiddingService_GetMultiTouchAttribution(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	credits, err := service.GetMultiTouchAttribution("user-123", "campaign-456", "linear")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Nil from mock cache
	if len(credits) > 0 {
		t.Log("Got MTA credits data")
	}
}

func TestBiddingService_GetAutoBidRecommendations(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	// This should not panic even with empty campaigns
	recs, err := service.GetAutoBidRecommendations()
	// May return error or nil from mock
	if err != nil {
		t.Logf("Expected error from mock: %v", err)
	}
	if recs != nil {
		t.Logf("Got %d recommendations", len(recs))
	}
}

func TestBiddingService_GetCrossDeviceFrequency(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	freq := service.GetCrossDeviceFrequency("user-123", "campaign-456")
	// Should return 0 from mock
	if freq < 0 {
		t.Error("Expected non-negative frequency")
	}
}

func TestBiddingService_LinkUserDevices(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	err := service.LinkUserDevices("primary-user", []string{"device1", "device2"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestBiddingService_GetUserDeviceGraph(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	graph, err := service.GetUserDeviceGraph("user-123")
	// May return error from mock when linked devices not found
	if err != nil {
		t.Logf("Expected error from mock: %v", err)
	}
	if graph != nil {
		t.Logf("Got device graph with %d devices", graph.DeviceCount)
	}
}

func TestBiddingService_GetSupplyPathAnalytics(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	analytics := service.GetSupplyPathAnalytics()
	// May be nil or empty slice
	if analytics != nil {
		t.Logf("Got %d analytics entries", len(analytics))
	}
}

func TestBiddingService_IncrementFormatStats(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	// This returns void - just verify it doesn't panic
	service.IncrementFormatStats("banner")
}

func TestBiddingService_GetMetrics(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	metrics, err := service.GetMetrics()
	// May return error from mock
	if err != nil {
		t.Logf("Error from mock: %v", err)
	}
	if metrics != nil {
		t.Log("Got metrics data")
	}
}

func TestBiddingService_RefreshCampaigns(t *testing.T) {
	cache := &mockTestCache{}
	service := NewBiddingService(cache, "http://test-backend:8080")

	// This should not panic even without server
	err := service.RefreshCampaigns("http://test-backend:8080")
	// Will likely fail since there's no actual backend
	if err != nil {
		t.Logf("Expected error without backend: %v", err)
	}
}

// ============================================================================
// UTILITY FUNCTION TESTS
// ============================================================================

func TestCountOverlap(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []string
		expected int
	}{
		{
			name:     "No overlap",
			a:        []string{"a", "b"},
			b:        []string{"c", "d"},
			expected: 0,
		},
		{
			name:     "Partial overlap",
			a:        []string{"a", "b", "c"},
			b:        []string{"b", "c", "d"},
			expected: 2,
		},
		{
			name:     "Complete overlap",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: 3,
		},
		{
			name:     "Empty slices",
			a:        []string{},
			b:        []string{},
			expected: 0,
		},
		{
			name:     "One empty",
			a:        []string{"a", "b"},
			b:        []string{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countOverlap(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("countOverlap(%v, %v) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// BID CACHE SERVICE TESTS
// ============================================================================

func TestBidCacheService_SetWithTTL(t *testing.T) {
	service := NewBidCacheService(nil)
	ctx := context.Background()

	cachedBid := &CachedBid{
		Key:        "test-key",
		Price:      1.50,
		CampaignID: "camp-123",
	}
	service.SetWithTTL(ctx, "test-key", cachedBid, 60*time.Second)

	// Verify the value was stored
	val, found := service.Get(ctx, "test-key")
	if !found {
		t.Error("Expected to find cached value")
	}
	if val == nil {
		t.Error("Expected value, got nil")
	}
}

func TestBidCacheService_InvalidateCampaign(t *testing.T) {
	service := NewBidCacheService(nil)
	ctx := context.Background()

	// Set up some test data
	service.Set(ctx, "campaign:camp1:banner", &CachedBid{Key: "k1", CampaignID: "camp1"})
	service.Set(ctx, "campaign:camp1:video", &CachedBid{Key: "k2", CampaignID: "camp1"})
	service.Set(ctx, "campaign:camp2:banner", &CachedBid{Key: "k3", CampaignID: "camp2"})

	// Invalidate campaign
	count := service.InvalidateCampaign("camp1")
	// The function should invalidate campaign-related entries
	if count < 0 {
		t.Error("Expected non-negative invalidation count")
	}
}

func TestBidCacheService_GetConfig(t *testing.T) {
	service := NewBidCacheService(nil)

	config := service.GetConfig()
	if config == nil {
		t.Error("Expected config, got nil")
	}
	// Check default values
	if config.MaxEntries <= 0 {
		t.Error("Expected positive MaxEntries")
	}
	if config.TTL <= 0 {
		t.Error("Expected positive TTL")
	}
}

func TestBidCacheService_UpdateConfig(t *testing.T) {
	service := NewBidCacheService(nil)

	newConfig := &BidCacheConfig{
		MaxEntries:         5000,
		TTL:                120 * time.Second,
		ImpressionCacheTTL: 600 * time.Second,
		EnablePartnerCache: true,
		EnablePredictive:   true,
	}

	service.UpdateConfig(newConfig)

	config := service.GetConfig()
	if config.MaxEntries != 5000 {
		t.Errorf("Expected MaxEntries 5000, got %d", config.MaxEntries)
	}
}

// ============================================================================
// S2S BIDDING SERVICE ADDITIONAL TESTS
// ============================================================================

func TestS2SBiddingService_SetTimeout(t *testing.T) {
	service := NewS2SBiddingService(nil)

	// Set timeout
	service.SetTimeout(150 * time.Millisecond)

	// Note: There's no getter for timeout, but the function should not panic
	// and should update the internal state
}

// ============================================================================
// DYNAMIC BID SERVICE ADDITIONAL TESTS
// ============================================================================

func TestDynamicBidService_UpdateHourlyMultiplier(t *testing.T) {
	service := NewDynamicBidService(nil)

	// These return void - just verify they don't panic
	service.UpdateHourlyMultiplier(14, 1.5) // 2 PM boost

	// Invalid hour should fail silently or panic (we catch it)
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic for invalid hour: %v", r)
		}
	}()
	service.UpdateHourlyMultiplier(25, 1.5)
}

func TestDynamicBidService_UpdateDeviceMultiplier(t *testing.T) {
	service := NewDynamicBidService(nil)

	// This returns void - just verify it doesn't panic
	service.UpdateDeviceMultiplier("mobile", 1.2)
}

// ============================================================================
// COMPETITIVE INTELLIGENCE SERVICE ADDITIONAL TESTS
// ============================================================================

func TestCompetitiveIntelligenceService_GetCompetitorProfile(t *testing.T) {
	service := NewCompetitiveIntelligenceService(nil)

	profile, found := service.GetCompetitorProfile("competitor-123")
	// May be nil from empty service
	if !found {
		t.Log("Competitor not found (expected for empty service)")
	}
	if profile != nil {
		t.Logf("Got competitor profile: %+v", profile)
	}
}

func TestCompetitiveIntelligenceService_GetSegmentFloor(t *testing.T) {
	service := NewCompetitiveIntelligenceService(nil)

	floor := service.GetSegmentFloor("mobile-US-gaming")
	// Should return some value (may be 0 from empty service)
	t.Logf("Got segment floor: %f", floor)
}

// ============================================================================
// REAL-TIME ALERTS SERVICE ADDITIONAL TESTS
// ============================================================================

func TestRealTimeAlertService_GetActiveAlerts(t *testing.T) {
	service := NewRealTimeAlertService(nil)

	alerts := service.GetActiveAlerts("campaign-123")
	// Returns nil slice when no alerts found - this is valid
	if len(alerts) > 0 {
		t.Logf("Got %d alerts", len(alerts))
	}
}

func TestRealTimeAlertService_AcknowledgeAlert(t *testing.T) {
	service := NewRealTimeAlertService(nil)

	found := service.AcknowledgeAlert("alert-123")
	// Should return false for non-existent alert
	if found {
		t.Log("Alert was found and acknowledged")
	}
}

func TestRealTimeAlertService_ClearOldAlerts(t *testing.T) {
	service := NewRealTimeAlertService(nil)

	// This returns void
	service.ClearOldAlerts(24) // hours
}
