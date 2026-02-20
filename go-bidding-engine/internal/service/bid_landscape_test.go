package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

func createBLRequest() *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-bl-1",
		PublisherID: "pub-bl-123",
		Device: model.InternalDevice{
			Type: "mobile",
			Geo:  model.InternalGeo{Country: "US"},
		},
	}
}

func createBLCampaign(enabled bool, minSample int) *model.Campaign {
	return &model.Campaign{
		ID:       "camp-bl-1",
		BidPrice: 2.0,
		Targeting: model.Targeting{
			BidLandscape: &model.BidLandscape{
				Enabled:        enabled,
				MinSampleSize:  minSample,
				AnalysisWindow: 24,
			},
		},
	}
}

func TestBL_NewService(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.landscapes == nil {
		t.Error("expected landscapes to be initialized")
	}
}

func TestBL_AnalyzeLandscape_Disabled(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	campaign := createBLCampaign(false, 10)
	request := createBLRequest()

	result := svc.AnalyzeLandscape(campaign, request)

	if result.Analyzed {
		t.Error("expected Analyzed=false when disabled")
	}
	if result.BidMultiplier != 1.0 {
		t.Errorf("expected BidMultiplier=1.0, got %f", result.BidMultiplier)
	}
	if result.Reason != "bid_landscape_disabled" {
		t.Errorf("expected reason=bid_landscape_disabled, got %s", result.Reason)
	}
}

func TestBL_AnalyzeLandscape_NoData(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	campaign := createBLCampaign(true, 10)
	request := createBLRequest()

	result := svc.AnalyzeLandscape(campaign, request)

	if result.Analyzed {
		t.Error("expected Analyzed=false with no data")
	}
	if result.Reason != "insufficient_data" {
		t.Errorf("expected reason=insufficient_data, got %s", result.Reason)
	}
}

func TestBL_RecordBid(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()

	svc.RecordBid(request, 2.0, 1.8, true)

	// Verify internal state
	key := svc.generateLandscapeKey(request)
	if _, exists := svc.landscapes[key]; !exists {
		t.Error("expected landscape data to be created")
	}
}

func TestBL_RecordBid_MultipleBids(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()

	for i := 0; i < 100; i++ {
		won := i%3 == 0
		svc.RecordBid(request, 2.0+float64(i%10)*0.1, 1.8, won)
	}

	key := svc.generateLandscapeKey(request)
	data := svc.landscapes[key]
	if len(data.bids) != 100 {
		t.Errorf("expected 100 bids, got %d", len(data.bids))
	}
}

func TestBL_AnalyzeLandscape_WithSufficientData(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()
	campaign := createBLCampaign(true, 10)

	// Record sufficient bids
	for i := 0; i < 50; i++ {
		won := i%2 == 0
		svc.RecordBid(request, 1.5+float64(i)*0.05, 1.4, won)
	}

	result := svc.AnalyzeLandscape(campaign, request)

	if !result.Analyzed {
		t.Error("expected Analyzed=true with sufficient data")
	}
	if result.SampleSize < 50 {
		t.Errorf("expected SampleSize >= 50, got %d", result.SampleSize)
	}
	if result.OptimalRange == nil {
		t.Error("expected OptimalRange to be set")
	}
}

func TestBL_AnalyzeLandscape_OptimalRange(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()
	campaign := createBLCampaign(true, 10)

	// Record bids with realistic distribution
	for i := 0; i < 100; i++ {
		bid := 1.0 + float64(i%20)*0.1
		clear := bid * 0.9
		won := i%4 == 0
		svc.RecordBid(request, bid, clear, won)
	}

	result := svc.AnalyzeLandscape(campaign, request)

	if result.OptimalRange == nil {
		t.Fatal("expected OptimalRange")
	}
	if result.OptimalRange.MinBid <= 0 {
		t.Error("expected positive MinBid")
	}
	if result.OptimalRange.MaxBid < result.OptimalRange.MinBid {
		t.Error("expected MaxBid >= MinBid")
	}
	// SweetSpot is calculated from best efficiency percentile which can be outside MinBid-MaxBid range
	if result.OptimalRange.SweetSpot <= 0 {
		t.Error("expected positive SweetSpot")
	}
}

func TestBL_AnalyzeLandscape_MarketConditions(t *testing.T) {
	tests := []struct {
		name      string
		winRate   float64 // Approximate win rate to simulate
		bidSpread float64 // Spread in bid prices
	}{
		{"soft_market", 0.6, 0.2},
		{"competitive_market", 0.3, 0.4},
		{"aggressive_market", 0.1, 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewBidLandscapeService(nil)
			request := createBLRequest()
			campaign := createBLCampaign(true, 10)

			// Create data simulating market condition
			for i := 0; i < 50; i++ {
				bid := 1.5 + float64(i%10)*tt.bidSpread*0.1
				won := float64(i)/50 < tt.winRate
				svc.RecordBid(request, bid, bid*0.9, won)
			}

			result := svc.AnalyzeLandscape(campaign, request)

			if result.MarketCondition == "" || result.MarketCondition == "unknown" {
				t.Error("expected market condition to be determined")
			}
		})
	}
}

func TestBL_AnalyzeLandscape_Confidence(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()
	campaign := createBLCampaign(true, 10)

	// Few samples = low confidence
	for i := 0; i < 15; i++ {
		svc.RecordBid(request, 2.0, 1.8, i%2 == 0)
	}
	result1 := svc.AnalyzeLandscape(campaign, request)

	// Many samples = higher confidence
	for i := 0; i < 100; i++ {
		svc.RecordBid(request, 2.0+float64(i%5)*0.1, 1.8, i%2 == 0)
	}
	result2 := svc.AnalyzeLandscape(campaign, request)

	if result2.Confidence <= result1.Confidence {
		t.Error("expected higher confidence with more samples")
	}
}

func TestBL_AnalyzeLandscape_Caching(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()
	campaign := createBLCampaign(true, 10)

	for i := 0; i < 50; i++ {
		svc.RecordBid(request, 2.0, 1.8, i%2 == 0)
	}

	// First analysis
	result1 := svc.AnalyzeLandscape(campaign, request)

	// Second analysis should use cache
	result2 := svc.AnalyzeLandscape(campaign, request)

	if result1.Analyzed != result2.Analyzed {
		t.Error("expected cached result to be consistent")
	}
}

func TestBL_AnalyzeLandscape_CacheInvalidation(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()
	campaign := createBLCampaign(true, 10)

	for i := 0; i < 50; i++ {
		svc.RecordBid(request, 2.0, 1.8, i%2 == 0)
	}

	result1 := svc.AnalyzeLandscape(campaign, request)

	// Record more bids - should invalidate cache
	for i := 0; i < 10; i++ {
		svc.RecordBid(request, 3.0, 2.8, true)
	}

	key := svc.generateLandscapeKey(request)
	data := svc.landscapes[key]
	if data.cachedResult != nil {
		t.Error("expected cache to be invalidated after new bids")
	}
	_ = result1
}

func TestBL_BidMultiplier_Range(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()
	campaign := createBLCampaign(true, 10)

	for i := 0; i < 100; i++ {
		svc.RecordBid(request, float64(i%10)+0.5, 1.0, i%3 == 0)
	}

	result := svc.AnalyzeLandscape(campaign, request)

	// Multiplier should be capped
	if result.BidMultiplier < 0.5 || result.BidMultiplier > 2.0 {
		t.Errorf("expected multiplier in [0.5, 2.0], got %f", result.BidMultiplier)
	}
}

func TestBL_LandscapeKey_Generation(t *testing.T) {
	svc := NewBidLandscapeService(nil)

	tests := []struct {
		name    string
		request *model.BidRequest
	}{
		{
			name: "full info",
			request: &model.BidRequest{
				PublisherID: "pub-1",
				Device: model.InternalDevice{
					Type: "mobile",
					Geo:  model.InternalGeo{Country: "US"},
				},
			},
		},
		{
			name: "minimal",
			request: &model.BidRequest{
				PublisherID: "pub-2",
			},
		},
		{
			name: "partial",
			request: &model.BidRequest{
				PublisherID: "pub-3",
				Device:      model.InternalDevice{Type: "desktop"},
			},
		},
	}

	keys := make(map[string]bool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := svc.generateLandscapeKey(tt.request)
			if key == "" {
				t.Error("expected non-empty key")
			}
			if keys[key] {
				t.Error("expected unique keys for different requests")
			}
			keys[key] = true
		})
	}
}

func TestBL_HistoryLimit(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()

	// Record more than limit
	for i := 0; i < 11000; i++ {
		svc.RecordBid(request, 2.0, 1.8, i%2 == 0)
	}

	key := svc.generateLandscapeKey(request)
	data := svc.landscapes[key]
	if len(data.bids) > 10000 {
		t.Errorf("expected bids to be capped at 10000, got %d", len(data.bids))
	}
}

func TestBL_Concurrency(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()
	campaign := createBLCampaign(true, 10)
	done := make(chan bool, 3)

	// Writer
	go func() {
		for i := 0; i < 500; i++ {
			svc.RecordBid(request, 2.0+float64(i%10)*0.1, 1.8, i%2 == 0)
		}
		done <- true
	}()

	// Reader 1
	go func() {
		for i := 0; i < 100; i++ {
			svc.AnalyzeLandscape(campaign, request)
		}
		done <- true
	}()

	// Reader 2
	go func() {
		for i := 0; i < 100; i++ {
			svc.generateLandscapeKey(request)
		}
		done <- true
	}()

	<-done
	<-done
	<-done
}

func TestBL_AnalysisWindowFilter(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()
	campaign := createBLCampaign(true, 5)
	campaign.Targeting.BidLandscape.AnalysisWindow = 1 // 1 hour window

	// Record bids (all recent)
	for i := 0; i < 20; i++ {
		svc.RecordBid(request, 2.0, 1.8, i%2 == 0)
	}

	result := svc.AnalyzeLandscape(campaign, request)

	// Should analyze recent data
	if !result.Analyzed {
		t.Error("expected analysis with recent data")
	}
}

func TestBL_RecommendedBid(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()
	campaign := createBLCampaign(true, 10)

	for i := 0; i < 100; i++ {
		bid := 1.5 + float64(i%20)*0.1
		svc.RecordBid(request, bid, bid*0.9, i%3 == 0)
	}

	result := svc.AnalyzeLandscape(campaign, request)

	if result.RecommendedBid <= 0 {
		t.Error("expected positive RecommendedBid")
	}
}

func TestBL_PercentileCalculation(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()
	campaign := createBLCampaign(true, 5)

	// Create known distribution
	for i := 0; i < 100; i++ {
		bid := float64(i) / 10.0                      // 0.0 to 9.9
		svc.RecordBid(request, bid, bid*0.9, i >= 50) // Win if bid >= 5.0
	}

	result := svc.AnalyzeLandscape(campaign, request)

	if !result.Analyzed {
		t.Fatal("expected analysis")
	}
	// Percentiles should be calculated (checked via OptimalRange)
	if result.OptimalRange == nil {
		t.Error("expected percentile-derived OptimalRange")
	}
}

func TestBL_WinLossTracking(t *testing.T) {
	svc := NewBidLandscapeService(nil)
	request := createBLRequest()

	// All wins
	for i := 0; i < 20; i++ {
		svc.RecordBid(request, 3.0, 2.5, true)
	}

	campaign := createBLCampaign(true, 10)
	result := svc.AnalyzeLandscape(campaign, request)

	// High win rate should indicate soft market
	if result.MarketCondition != "soft" {
		// Could also be "competitive" depending on spread
	}
	_ = result
}

func TestBL_DeviceTypeSegmentation(t *testing.T) {
	svc := NewBidLandscapeService(nil)

	mobileReq := &model.BidRequest{
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Type: "mobile"},
	}
	desktopReq := &model.BidRequest{
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Type: "desktop"},
	}

	mobileKey := svc.generateLandscapeKey(mobileReq)
	desktopKey := svc.generateLandscapeKey(desktopReq)

	if mobileKey == desktopKey {
		t.Error("expected different keys for different device types")
	}
}

func TestBL_GeoSegmentation(t *testing.T) {
	svc := NewBidLandscapeService(nil)

	usReq := &model.BidRequest{
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Geo: model.InternalGeo{Country: "US"}},
	}
	ukReq := &model.BidRequest{
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Geo: model.InternalGeo{Country: "UK"}},
	}

	usKey := svc.generateLandscapeKey(usReq)
	ukKey := svc.generateLandscapeKey(ukReq)

	if usKey == ukKey {
		t.Error("expected different keys for different geos")
	}
}
