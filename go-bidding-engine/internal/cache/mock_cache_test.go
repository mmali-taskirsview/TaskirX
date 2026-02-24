package cache

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

func TestMockCache_ImplementsInterface(t *testing.T) {
	// Verify MockCache implements Cache interface
	var _ Cache = (*MockCache)(nil)
}

func TestMockCache_CampaignOperations(t *testing.T) {
	cache := NewMockCache()

	// Test empty cache
	campaigns, err := cache.GetActiveCampaigns()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(campaigns) != 0 {
		t.Errorf("expected 0 campaigns, got %d", len(campaigns))
	}

	// Test set and get active campaigns
	testCampaigns := []*model.Campaign{
		{ID: "camp-1", Name: "Campaign 1", Budget: 1000.0},
		{ID: "camp-2", Name: "Campaign 2", Budget: 2000.0},
	}

	err = cache.SetActiveCampaigns(testCampaigns)
	if err != nil {
		t.Fatalf("unexpected error setting campaigns: %v", err)
	}

	campaigns, err = cache.GetActiveCampaigns()
	if err != nil {
		t.Fatalf("unexpected error getting campaigns: %v", err)
	}
	if len(campaigns) != 2 {
		t.Errorf("expected 2 campaigns, got %d", len(campaigns))
	}

	// Test single campaign
	campaign := &model.Campaign{ID: "camp-3", Name: "Campaign 3", Budget: 3000.0}
	err = cache.SetCampaign(campaign)
	if err != nil {
		t.Fatalf("unexpected error setting campaign: %v", err)
	}

	retrieved, err := cache.GetCampaign("camp-3")
	if err != nil {
		t.Fatalf("unexpected error getting campaign: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected campaign, got nil")
	}
	if retrieved.Name != "Campaign 3" {
		t.Errorf("expected name 'Campaign 3', got '%s'", retrieved.Name)
	}

	// Test cache miss
	notFound, err := cache.GetCampaign("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for nonexistent campaign")
	}
}

func TestMockCache_BidMetrics(t *testing.T) {
	cache := NewMockCache()

	// Initial state
	bidCount, _ := cache.GetBidCount()
	if bidCount != 0 {
		t.Errorf("expected 0 bids, got %d", bidCount)
	}

	// Increment bids
	for i := 0; i < 10; i++ {
		cache.IncrementBidCount()
	}

	bidCount, _ = cache.GetBidCount()
	if bidCount != 10 {
		t.Errorf("expected 10 bids, got %d", bidCount)
	}

	// Increment wins
	for i := 0; i < 3; i++ {
		cache.IncrementWinCount()
	}

	winCount, _ := cache.GetWinCount()
	if winCount != 3 {
		t.Errorf("expected 3 wins, got %d", winCount)
	}
}

func TestMockCache_Latency(t *testing.T) {
	cache := NewMockCache()

	// Empty latency
	avg, _ := cache.GetAverageLatency()
	if avg != 0 {
		t.Errorf("expected 0 average latency, got %f", avg)
	}

	// Record latencies
	cache.RecordLatency(10.0)
	cache.RecordLatency(20.0)
	cache.RecordLatency(30.0)

	avg, _ = cache.GetAverageLatency()
	if avg != 20.0 {
		t.Errorf("expected 20.0 average latency, got %f", avg)
	}
}

func TestMockCache_UserSegments(t *testing.T) {
	cache := NewMockCache()

	segments := []string{"tech", "sports", "news"}
	err := cache.SetUserSegments("user-123", segments)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, err := cache.GetUserSegments("user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("expected 3 segments, got %d", len(retrieved))
	}

	// Non-existent user
	empty, _ := cache.GetUserSegments("nonexistent")
	if len(empty) != 0 {
		t.Errorf("expected 0 segments for nonexistent user, got %d", len(empty))
	}
}

func TestMockCache_GeoRules(t *testing.T) {
	cache := NewMockCache()

	rules := map[string]interface{}{
		"min_bid":    1.5,
		"max_bid":    10.0,
		"blocked":    false,
		"categories": []string{"tech", "finance"},
	}

	err := cache.SetGeoRules("US", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, err := cache.GetGeoRules("US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved["min_bid"].(float64) != 1.5 {
		t.Errorf("expected min_bid 1.5, got %v", retrieved["min_bid"])
	}
}

func TestMockCache_CampaignSpend(t *testing.T) {
	cache := NewMockCache()

	// Initial spend
	spend, _ := cache.GetCampaignSpend("camp-1")
	if spend != 0 {
		t.Errorf("expected 0 spend, got %f", spend)
	}

	// Increment spend
	newSpend, err := cache.IncrementCampaignSpend("camp-1", 50.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newSpend != 50.0 {
		t.Errorf("expected 50.0 spend, got %f", newSpend)
	}

	// Increment again
	newSpend, _ = cache.IncrementCampaignSpend("camp-1", 25.0)
	if newSpend != 75.0 {
		t.Errorf("expected 75.0 spend, got %f", newSpend)
	}
}

func TestMockCache_BidFormats(t *testing.T) {
	cache := NewMockCache()

	// Record formats
	cache.IncrementBidFormat("banner")
	cache.IncrementBidFormat("banner")
	cache.IncrementBidFormat("video")
	cache.IncrementBidFormat("native")

	formats, err := cache.GetBidFormats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if formats["banner"] != 2 {
		t.Errorf("expected 2 banner bids, got %d", formats["banner"])
	}
	if formats["video"] != 1 {
		t.Errorf("expected 1 video bid, got %d", formats["video"])
	}
}

func TestMockCache_GenericGetSet(t *testing.T) {
	cache := NewMockCache()

	// String value
	err := cache.Set("key1", "value1", 3600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got '%s'", val)
	}

	// Object value
	data := map[string]interface{}{"foo": "bar"}
	cache.Set("key2", data, 3600)
	val2, _ := cache.Get("key2")
	if val2 == "" {
		t.Error("expected non-empty value for object")
	}
}

func TestMockCache_RequestDeduplication(t *testing.T) {
	cache := NewMockCache()

	// First request
	isDup, err := cache.IsRequestDuplicate("req-123", 60)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isDup {
		t.Error("first request should not be duplicate")
	}

	// Same request again
	isDup, _ = cache.IsRequestDuplicate("req-123", 60)
	if !isDup {
		t.Error("second request should be duplicate")
	}

	// Different request
	isDup, _ = cache.IsRequestDuplicate("req-456", 60)
	if isDup {
		t.Error("different request should not be duplicate")
	}
}

func TestMockCache_FrequencyCapping(t *testing.T) {
	cache := NewMockCache()

	// Initial frequency
	freq, _ := cache.GetUserFrequency("user-1", "camp-1")
	if freq != 0 {
		t.Errorf("expected 0 frequency, got %d", freq)
	}

	// Increment
	newFreq, err := cache.IncrementUserFrequency("user-1", "camp-1", 3600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newFreq != 1 {
		t.Errorf("expected 1, got %d", newFreq)
	}

	// Increment again
	newFreq, _ = cache.IncrementUserFrequency("user-1", "camp-1", 3600)
	if newFreq != 2 {
		t.Errorf("expected 2, got %d", newFreq)
	}
}

func TestMockCache_CampaignPerformance(t *testing.T) {
	cache := NewMockCache()

	// Record activity
	for i := 0; i < 100; i++ {
		cache.IncrementCampaignImpressions("camp-1")
		cache.IncrementCampaignBids("camp-1")
	}
	for i := 0; i < 5; i++ {
		cache.IncrementCampaignClicks("camp-1")
		cache.IncrementCampaignWins("camp-1")
	}

	// Check CTR
	ctr, _ := cache.GetCampaignCTR("camp-1")
	if ctr != 0.05 {
		t.Errorf("expected CTR 0.05, got %f", ctr)
	}

	// Check win rate
	winRate, _ := cache.GetCampaignWinRate("camp-1")
	if winRate != 0.05 {
		t.Errorf("expected win rate 0.05, got %f", winRate)
	}

	// Zero impressions
	ctr, _ = cache.GetCampaignCTR("nonexistent")
	if ctr != 0 {
		t.Errorf("expected 0 CTR for nonexistent campaign, got %f", ctr)
	}
}

func TestMockCache_BidLandscape(t *testing.T) {
	cache := NewMockCache()

	// Record bids and wins
	cache.RecordBidInBucket("1.00-1.50")
	cache.RecordBidInBucket("1.00-1.50")
	cache.RecordBidInBucket("1.50-2.00")
	cache.RecordWinInBucket("1.00-1.50")

	landscape, err := cache.GetBidLandscape()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if landscape["1.00-1.50"]["bids"] != 2 {
		t.Errorf("expected 2 bids in 1.00-1.50, got %d", landscape["1.00-1.50"]["bids"])
	}
	if landscape["1.00-1.50"]["wins"] != 1 {
		t.Errorf("expected 1 win in 1.00-1.50, got %d", landscape["1.00-1.50"]["wins"])
	}
}

func TestMockCache_SegmentPerformance(t *testing.T) {
	cache := NewMockCache()

	// Record segment activity
	cache.IncrementSegmentImpressions("device", "mobile")
	cache.IncrementSegmentImpressions("device", "mobile")
	cache.IncrementSegmentImpressions("device", "desktop")
	cache.IncrementSegmentClicks("device", "mobile")

	perf, err := cache.GetSegmentPerformance("device")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if perf["mobile"]["impressions"] != 2 {
		t.Errorf("expected 2 mobile impressions, got %d", perf["mobile"]["impressions"])
	}
	if perf["mobile"]["clicks"] != 1 {
		t.Errorf("expected 1 mobile click, got %d", perf["mobile"]["clicks"])
	}
}

func TestMockCache_BidFloorOptimization(t *testing.T) {
	cache := NewMockCache()

	// Record bid attempts
	cache.RecordPublisherBidAttempt("pub-1", 2.0, true)
	cache.RecordPublisherBidAttempt("pub-1", 1.5, true)
	cache.RecordPublisherBidAttempt("pub-1", 1.0, false)

	floor, err := cache.GetOptimalBidFloor("pub-1", 0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be slightly below average win price
	if floor <= 0 || floor >= 2.0 {
		t.Errorf("expected floor between 0 and 2.0, got %f", floor)
	}

	// Default floor for unknown publisher
	floor, _ = cache.GetOptimalBidFloor("unknown", 0.5)
	if floor != 1.0 {
		t.Errorf("expected default floor 1.0, got %f", floor)
	}
}

func TestMockCache_Attribution(t *testing.T) {
	cache := NewMockCache()

	// Record impression
	cache.RecordImpression("user-1", "camp-1", "req-1", 24)

	// Check VTA attribution
	attrType, reqID, err := cache.GetAttribution("user-1", "camp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attrType != "VTA" {
		t.Errorf("expected VTA attribution, got %s", attrType)
	}
	if reqID != "req-1" {
		t.Errorf("expected req-1, got %s", reqID)
	}

	// Record click (should override impression)
	cache.RecordClick("user-1", "camp-1", "req-2", 24)

	attrType, reqID, _ = cache.GetAttribution("user-1", "camp-1")
	if attrType != "CTA" {
		t.Errorf("expected CTA attribution, got %s", attrType)
	}
	if reqID != "req-2" {
		t.Errorf("expected req-2, got %s", reqID)
	}

	// No attribution
	attrType, _, _ = cache.GetAttribution("user-2", "camp-1")
	if attrType != "" {
		t.Errorf("expected empty attribution, got %s", attrType)
	}
}

func TestMockCache_MultiTouchAttribution(t *testing.T) {
	cache := NewMockCache()

	// Record touchpoints
	cache.RecordTouchpoint("user-1", "camp-1", "impression", "req-1", 30)
	cache.RecordTouchpoint("user-1", "camp-1", "click", "req-2", 30)
	cache.RecordTouchpoint("user-1", "camp-1", "conversion", "req-3", 30)

	// Get touchpoints
	touchpoints, err := cache.GetTouchpoints("user-1", "camp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(touchpoints) != 3 {
		t.Errorf("expected 3 touchpoints, got %d", len(touchpoints))
	}

	// Get attribution credits
	credits, err := cache.GetMultiTouchAttribution("user-1", "camp-1", "linear")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 3 {
		t.Errorf("expected 3 credits, got %d", len(credits))
	}

	// Each touchpoint should get equal credit
	expectedCredit := 1.0 / 3.0
	for _, credit := range credits {
		if credit.Credit < expectedCredit-0.01 || credit.Credit > expectedCredit+0.01 {
			t.Errorf("expected credit ~%f, got %f", expectedCredit, credit.Credit)
		}
	}
}

func TestMockCache_UserEvents(t *testing.T) {
	cache := NewMockCache()

	// Record events
	cache.RecordUserEvent("user-1", "camp-1", "view", 30)
	cache.RecordUserEvent("user-1", "camp-1", "click", 30)
	cache.RecordUserEvent("user-1", "camp-2", "view", 30)

	// Check events
	hasView, _ := cache.HasUserEvent("user-1", "camp-1", "view")
	if !hasView {
		t.Error("expected user to have view event")
	}

	hasClick, _ := cache.HasUserEvent("user-1", "camp-1", "click")
	if !hasClick {
		t.Error("expected user to have click event")
	}

	hasConversion, _ := cache.HasUserEvent("user-1", "camp-1", "conversion")
	if hasConversion {
		t.Error("expected user NOT to have conversion event")
	}

	// Get events
	events, _ := cache.GetUserEvents("user-1", []string{"view", "click"})
	if len(events["view"]) != 2 {
		t.Errorf("expected 2 view events, got %d", len(events["view"]))
	}
}

func TestMockCache_CrossDeviceGraph(t *testing.T) {
	cache := NewMockCache()

	// Link devices
	err := cache.LinkDevices("primary-user", []string{"device-1", "device-2", "device-3"}, 365)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get linked devices
	devices, _ := cache.GetLinkedDevices("device-1")
	if len(devices) != 3 {
		t.Errorf("expected 3 linked devices, got %d", len(devices))
	}

	// Get primary user
	primary, _ := cache.GetPrimaryUserID("device-2")
	if primary != "primary-user" {
		t.Errorf("expected 'primary-user', got '%s'", primary)
	}

	// Unknown device
	primary, _ = cache.GetPrimaryUserID("unknown-device")
	if primary != "" {
		t.Errorf("expected empty string for unknown device, got '%s'", primary)
	}
}

func TestMockCache_CrossDeviceFrequency(t *testing.T) {
	cache := NewMockCache()

	// Link devices
	cache.LinkDevices("primary-user", []string{"device-1", "device-2"}, 365)

	// Record frequency on different devices
	cache.IncrementUserFrequency("device-1", "camp-1", 3600)
	cache.IncrementUserFrequency("device-1", "camp-1", 3600)
	cache.IncrementUserFrequency("device-2", "camp-1", 3600)

	// Get cross-device frequency
	freq, _ := cache.GetCrossDeviceFrequency("primary-user", "camp-1")
	if freq != 3 {
		t.Errorf("expected cross-device frequency 3, got %d", freq)
	}
}

func TestMockCache_BidPathAnalytics(t *testing.T) {
	cache := NewMockCache()

	analytics := &model.BidPathAnalytics{
		RequestID:      "req-123",
		PublisherID:    "pub-1",
		AdSlotID:       "slot-1",
		TotalHops:      3,
		TotalFees:      0.15,
		FinalBidPrice:  2.13,
		TotalLatencyMs: 45,
		WonAuction:     true,
	}

	err := cache.StoreBidPathAnalytics(analytics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, err := cache.GetBidPathAnalytics("req-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected analytics, got nil")
	}
	if retrieved.TotalHops != 3 {
		t.Errorf("expected total hops 3, got %d", retrieved.TotalHops)
	}

	// Not found
	notFound, _ := cache.GetBidPathAnalytics("nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent request")
	}
}

func TestMockCache_SupplyChainMetrics(t *testing.T) {
	cache := NewMockCache()

	metrics, err := cache.GetSupplyChainMetrics("24h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.TotalRequests != 100 {
		t.Errorf("expected 100 total requests, got %d", metrics.TotalRequests)
	}
	if metrics.SuccessfulBids != 70 {
		t.Errorf("expected 70 successful bids, got %d", metrics.SuccessfulBids)
	}
}

func TestMockCache_ServiceMetrics(t *testing.T) {
	cache := NewMockCache()

	metrics, err := cache.GetServiceMetrics("bidding", "1h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.ServiceName != "bidding" {
		t.Errorf("expected service name 'bidding', got '%s'", metrics.ServiceName)
	}
	if metrics.TotalCalls != 1000 {
		t.Errorf("expected 1000 total calls, got %d", metrics.TotalCalls)
	}
}

func TestMockCache_PublisherFraud(t *testing.T) {
	cache := NewMockCache()

	err := cache.IncrementPublisherFraud("pub-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cache.IncrementPublisherFraud("pub-1")
	cache.IncrementPublisherFraud("pub-1")

	// Verify it's tracked (internal state)
	if cache.publisherFraud["pub-1"] != 3 {
		t.Errorf("expected 3 fraud incidents, got %d", cache.publisherFraud["pub-1"])
	}
}

func TestMockCache_Concurrency(t *testing.T) {
	cache := NewMockCache()

	// Run concurrent operations
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				cache.IncrementBidCount()
				cache.IncrementWinCount()
				cache.RecordLatency(10.0)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	bidCount, _ := cache.GetBidCount()
	if bidCount != 1000 {
		t.Errorf("expected 1000 bids after concurrent operations, got %d", bidCount)
	}
}
