package service

import (
	"sync"
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

func TestDynBid_NewService(t *testing.T) {
	svc := NewDynamicBidService(nil)
	if svc == nil {
		t.Fatal("expected service")
	}
	if svc.config == nil {
		t.Error("expected config")
	}
	if !svc.config.Enabled {
		t.Error("expected enabled by default")
	}
	if svc.config.LearningRate != 0.01 {
		t.Errorf("expected learning rate 0.01, got %f", svc.config.LearningRate)
	}
}

func createDynBidCampaign() *model.Campaign {
	return &model.Campaign{
		ID:       "camp-db-1",
		Status:   "active",
		Budget:   1000,
		BidPrice: 2.50,
		GoalType: "CPM",
	}
}

func createDynBidRequest() *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-db-1",
		PublisherID: "pub-1",
		AdSlot: model.AdSlot{
			ID:         "slot-1",
			Formats:    []string{"banner"},
			Dimensions: []int{300, 250},
		},
		Device: model.InternalDevice{
			Type: "desktop",
		},
	}
}

func TestDynBid_CalculateDynamicBid_Disabled(t *testing.T) {
	svc := NewDynamicBidService(nil)
	svc.config.Enabled = false
	campaign := createDynBidCampaign()
	req := createDynBidRequest()

	result := svc.CalculateDynamicBid(campaign, req)

	if result.AdjustedBid != campaign.BidPrice {
		t.Errorf("expected unchanged bid when disabled, got %f", result.AdjustedBid)
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected multiplier 1.0 when disabled, got %f", result.Multiplier)
	}
}

func TestDynBid_CalculateDynamicBid_Enabled(t *testing.T) {
	svc := NewDynamicBidService(nil)
	campaign := createDynBidCampaign()
	req := createDynBidRequest()

	result := svc.CalculateDynamicBid(campaign, req)

	if result.OriginalBid != campaign.BidPrice {
		t.Errorf("expected original bid %f, got %f", campaign.BidPrice, result.OriginalBid)
	}
	if result.AdjustedBid <= 0 {
		t.Error("expected positive adjusted bid")
	}
	if len(result.Factors) == 0 {
		t.Error("expected factors in result")
	}
}

func TestDynBid_DeviceMultipliers(t *testing.T) {
	svc := NewDynamicBidService(nil)

	tests := []struct {
		device   string
		expected float64
	}{
		{"mobile", 1.0},
		{"desktop", 1.1},
		{"tablet", 0.9},
		{"ctv", 1.3},
		{"other", 0.8},
		{"unknown", 1.0}, // Default
	}

	for _, tt := range tests {
		mult := svc.getDeviceMultiplier(tt.device)
		if mult != tt.expected {
			t.Errorf("device %s: expected %f, got %f", tt.device, tt.expected, mult)
		}
	}
}

func TestDynBid_HourlyMultipliers(t *testing.T) {
	svc := NewDynamicBidService(nil)

	// Check peak hours
	if svc.hourlyMultipliers[19] <= 1.0 {
		t.Error("expected peak hour (19) to have multiplier > 1.0")
	}

	// Check off-peak hours
	if svc.hourlyMultipliers[3] >= 1.0 {
		t.Error("expected off-peak hour (3) to have multiplier < 1.0")
	}
}

func TestDynBid_ClampMultiplier(t *testing.T) {
	svc := NewDynamicBidService(nil)

	tests := []struct {
		input    float64
		expected float64
	}{
		{0.3, 0.5}, // Below min
		{0.5, 0.5}, // At min
		{1.0, 1.0}, // Normal
		{2.0, 2.0}, // At max
		{3.0, 2.0}, // Above max
	}

	for _, tt := range tests {
		result := svc.clampMultiplier(tt.input)
		if result != tt.expected {
			t.Errorf("input %f: expected %f, got %f", tt.input, tt.expected, result)
		}
	}
}

func TestDynBid_PublisherFactor_NoData(t *testing.T) {
	svc := NewDynamicBidService(nil)

	factor := svc.getPublisherFactor("unknown-pub")

	if factor != 1.0 {
		t.Errorf("expected factor 1.0 for unknown publisher, got %f", factor)
	}
}

func TestDynBid_PublisherFactor_WithData(t *testing.T) {
	svc := NewDynamicBidService(nil)

	// Add publisher data
	svc.publisherStats["pub-1"] = &publisherPerf{
		winRate:     0.3,
		avgWinPrice: 2.0,
		quality:     0.8,
		samples:     1000,
	}

	factor := svc.getPublisherFactor("pub-1")

	// Should be > 1.0 due to high quality
	if factor <= 1.0 {
		t.Errorf("expected factor > 1.0 for high quality publisher, got %f", factor)
	}
}

func TestDynBid_ContextFactor_NoData(t *testing.T) {
	svc := NewDynamicBidService(nil)
	req := createDynBidRequest()

	factor := svc.getContextFactor("camp-1", req)

	if factor != 1.0 {
		t.Errorf("expected factor 1.0 with no data, got %f", factor)
	}
}

func TestDynBid_ContextFactor_GoodPerformance(t *testing.T) {
	svc := NewDynamicBidService(nil)
	campaign := createDynBidCampaign()
	req := createDynBidRequest()

	// Add good performance data
	contextKey := svc.buildContextKey(campaign.ID, req)
	svc.contextPerformance[contextKey] = &contextStats{
		impressions: 1000,
		clicks:      50, // 5% CTR (good)
		conversions: 10, // 20% CVR (very good)
		spend:       100.0,
		revenue:     200.0, // 100% ROI (excellent)
	}

	factor := svc.getContextFactor(campaign.ID, req)

	// Should be > 1.0 due to good performance
	if factor <= 1.0 {
		t.Errorf("expected factor > 1.0 for good performance, got %f", factor)
	}
}

func TestDynBid_ContextFactor_PoorPerformance(t *testing.T) {
	svc := NewDynamicBidService(nil)
	campaign := createDynBidCampaign()
	req := createDynBidRequest()

	// Add poor performance data
	contextKey := svc.buildContextKey(campaign.ID, req)
	svc.contextPerformance[contextKey] = &contextStats{
		impressions: 1000,
		clicks:      5, // 0.5% CTR (poor)
		conversions: 0, // 0% CVR (very poor)
		spend:       100.0,
		revenue:     50.0, // -50% ROI (bad)
	}

	factor := svc.getContextFactor(campaign.ID, req)

	// Should be < 1.0 due to poor ROI
	if factor >= 1.0 {
		t.Errorf("expected factor < 1.0 for poor performance, got %f", factor)
	}
}

func TestDynBid_CompetitionFactor(t *testing.T) {
	svc := NewDynamicBidService(nil)

	// Video format should increase competition
	videoReq := &model.BidRequest{
		AdSlot: model.AdSlot{
			Formats: []string{"video"},
		},
	}
	videoFactor := svc.getCompetitionFactor(videoReq)
	if videoFactor <= 1.0 {
		t.Errorf("expected video competition factor > 1.0, got %f", videoFactor)
	}

	// Banner with large dimensions
	bannerReq := &model.BidRequest{
		AdSlot: model.AdSlot{
			Formats:    []string{"banner"},
			Dimensions: []int{300, 250},
		},
	}
	bannerFactor := svc.getCompetitionFactor(bannerReq)
	if bannerFactor <= 1.0 {
		t.Errorf("expected large banner competition factor > 1.0, got %f", bannerFactor)
	}
}

func TestDynBid_GoalFactor(t *testing.T) {
	svc := NewDynamicBidService(nil)

	tests := []struct {
		goalType  string
		minFactor float64
		maxFactor float64
	}{
		{"CPA", 0.8, 1.0},
		{"CPM", 1.0, 1.2},
		{"CPC", 0.9, 1.1},
		{"CPCV", 1.1, 1.2},
		{"unknown", 0.9, 1.1},
	}

	for _, tt := range tests {
		campaign := &model.Campaign{GoalType: tt.goalType}
		factor := svc.getGoalFactor(campaign)
		if factor < tt.minFactor || factor > tt.maxFactor {
			t.Errorf("goal %s: expected factor in [%f, %f], got %f", tt.goalType, tt.minFactor, tt.maxFactor, factor)
		}
	}
}

func TestDynBid_CalculateConfidence(t *testing.T) {
	svc := NewDynamicBidService(nil)

	// No data = low confidence
	conf := svc.calculateConfidence("camp-1", "pub-new")
	if conf > 0.2 {
		t.Errorf("expected low confidence with no data, got %f", conf)
	}

	// Add publisher data
	svc.publisherStats["pub-1"] = &publisherPerf{
		samples: 5000,
	}

	conf = svc.calculateConfidence("camp-1", "pub-1")
	if conf < 0.5 {
		t.Errorf("expected higher confidence with data, got %f", conf)
	}
}

func TestDynBid_PredictWinRate(t *testing.T) {
	svc := NewDynamicBidService(nil)

	// No data - use default estimation
	rate := svc.predictWinRate(2.0, "unknown-pub")
	if rate <= 0 || rate >= 1 {
		t.Errorf("expected win rate in (0, 1), got %f", rate)
	}

	// With publisher data
	svc.publisherStats["pub-1"] = &publisherPerf{
		avgWinPrice: 2.0,
	}

	// Bid at average should give ~50% win rate
	rate = svc.predictWinRate(2.0, "pub-1")
	if rate < 0.4 || rate > 0.6 {
		t.Errorf("expected ~50%% win rate at avg price, got %f", rate)
	}

	// Higher bid should give higher win rate
	highRate := svc.predictWinRate(3.0, "pub-1")
	if highRate <= rate {
		t.Error("expected higher win rate with higher bid")
	}
}

func TestDynBid_PredictROI(t *testing.T) {
	svc := NewDynamicBidService(nil)
	campaign := createDynBidCampaign()

	roi := svc.predictROI(campaign, 2.0, 0.5)

	// ROI should be a reasonable number
	if roi < -1 || roi > 100 {
		t.Errorf("unexpected ROI value: %f", roi)
	}
}

func TestDynBid_GenerateRecommendation(t *testing.T) {
	svc := NewDynamicBidService(nil)

	tests := []struct {
		multiplier float64
		confidence float64
		roi        float64
		expected   string
	}{
		{1.0, 0.1, 0.5, "insufficient_data"},
		{1.0, 0.5, 0.8, "increase_bid"},
		{1.0, 0.5, -0.3, "decrease_bid"},
		{1.5, 0.5, 0.3, "aggressive_bid"},
		{0.7, 0.5, 0.1, "conservative_bid"},
		{1.0, 0.5, 0.1, "maintain_bid"},
	}

	for _, tt := range tests {
		result := svc.generateRecommendation(tt.multiplier, tt.confidence, tt.roi)
		if result != tt.expected {
			t.Errorf("mult=%f conf=%f roi=%f: expected %s, got %s", tt.multiplier, tt.confidence, tt.roi, tt.expected, result)
		}
	}
}

func TestDynBid_RecordOutcome_Win(t *testing.T) {
	svc := NewDynamicBidService(nil)
	campaign := createDynBidCampaign()
	req := createDynBidRequest()

	svc.RecordOutcome(campaign.ID, req, 2.5, 2.0, true, true, true, 10.0)

	// Check context stats
	contextKey := svc.buildContextKey(campaign.ID, req)
	stats := svc.contextPerformance[contextKey]

	if stats.bids != 1 {
		t.Errorf("expected 1 bid, got %d", stats.bids)
	}
	if stats.wins != 1 {
		t.Errorf("expected 1 win, got %d", stats.wins)
	}
	if stats.clicks != 1 {
		t.Errorf("expected 1 click, got %d", stats.clicks)
	}
	if stats.conversions != 1 {
		t.Errorf("expected 1 conversion, got %d", stats.conversions)
	}
	if stats.revenue != 10.0 {
		t.Errorf("expected revenue 10.0, got %f", stats.revenue)
	}
}

func TestDynBid_RecordOutcome_Loss(t *testing.T) {
	svc := NewDynamicBidService(nil)
	campaign := createDynBidCampaign()
	req := createDynBidRequest()

	svc.RecordOutcome(campaign.ID, req, 2.5, 0, false, false, false, 0)

	contextKey := svc.buildContextKey(campaign.ID, req)
	stats := svc.contextPerformance[contextKey]

	if stats.bids != 1 {
		t.Errorf("expected 1 bid, got %d", stats.bids)
	}
	if stats.wins != 0 {
		t.Errorf("expected 0 wins, got %d", stats.wins)
	}
}

func TestDynBid_UpdatePublisherStats(t *testing.T) {
	svc := NewDynamicBidService(nil)

	// Record multiple outcomes
	for i := 0; i < 100; i++ {
		won := i%2 == 0
		clicked := i%10 == 0
		converted := i%20 == 0
		svc.updatePublisherStats("pub-1", won, 2.0, clicked, converted)
	}

	stats := svc.publisherStats["pub-1"]

	if stats.samples != 100 {
		t.Errorf("expected 100 samples, got %d", stats.samples)
	}
	if stats.winRate <= 0 || stats.winRate > 1 {
		t.Errorf("expected valid win rate, got %f", stats.winRate)
	}
}

func TestDynBid_UpdateHourlyMultiplier(t *testing.T) {
	svc := NewDynamicBidService(nil)

	initialMult := svc.hourlyMultipliers[10]
	svc.UpdateHourlyMultiplier(10, 1.5)

	newMult := svc.hourlyMultipliers[10]
	if newMult == initialMult {
		t.Error("expected multiplier to change")
	}

	// Invalid hour should not crash
	svc.UpdateHourlyMultiplier(-1, 1.0)
	svc.UpdateHourlyMultiplier(24, 1.0)
}

func TestDynBid_UpdateDeviceMultiplier(t *testing.T) {
	svc := NewDynamicBidService(nil)

	initialMult := svc.deviceMultipliers["mobile"]
	svc.UpdateDeviceMultiplier("mobile", 1.2)

	newMult := svc.deviceMultipliers["mobile"]
	if newMult == initialMult {
		t.Error("expected multiplier to change")
	}

	// New device type
	svc.UpdateDeviceMultiplier("smartwatch", 0.5)
	if svc.deviceMultipliers["smartwatch"] != 0.5 {
		t.Error("expected new device multiplier to be set")
	}
}

func TestDynBid_GetBidAnalytics(t *testing.T) {
	svc := NewDynamicBidService(nil)
	campaign := createDynBidCampaign()
	req := createDynBidRequest()

	// Record some outcomes
	for i := 0; i < 10; i++ {
		won := i%2 == 0
		svc.RecordOutcome(campaign.ID, req, 2.5, 2.0, won, won, false, 0)
	}

	analytics := svc.GetBidAnalytics()

	if analytics["total_bids"].(int64) != 10 {
		t.Errorf("expected 10 total bids, got %v", analytics["total_bids"])
	}
	if analytics["total_wins"].(int64) != 5 {
		t.Errorf("expected 5 wins, got %v", analytics["total_wins"])
	}
	if analytics["contexts_tracked"].(int) != 1 {
		t.Errorf("expected 1 context tracked, got %v", analytics["contexts_tracked"])
	}
}

func TestDynBid_SetConfig(t *testing.T) {
	svc := NewDynamicBidService(nil)

	newConfig := &DynamicBidConfig{
		Enabled:          false,
		LearningRate:     0.05,
		MinBidMultiplier: 0.3,
		MaxBidMultiplier: 3.0,
	}

	svc.SetConfig(newConfig)
	config := svc.GetConfig()

	if config.Enabled {
		t.Error("expected disabled")
	}
	if config.LearningRate != 0.05 {
		t.Errorf("expected learning rate 0.05, got %f", config.LearningRate)
	}
}

func TestDynBid_BuildContextKey(t *testing.T) {
	svc := NewDynamicBidService(nil)

	req1 := &model.BidRequest{
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Type: "mobile"},
	}
	req2 := &model.BidRequest{
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Type: "desktop"},
	}
	req3 := &model.BidRequest{
		PublisherID: "pub-2",
		Device:      model.InternalDevice{Type: "mobile"},
	}

	key1 := svc.buildContextKey("camp-1", req1)
	key2 := svc.buildContextKey("camp-1", req2)
	key3 := svc.buildContextKey("camp-1", req3)

	if key1 == key2 {
		t.Error("different devices should produce different keys")
	}
	if key1 == key3 {
		t.Error("different publishers should produce different keys")
	}
}

func TestDynBid_Concurrency(t *testing.T) {
	svc := NewDynamicBidService(nil)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			campaign := createDynBidCampaign()
			req := createDynBidRequest()
			req.PublisherID = "pub-" + string(rune(idx%10))

			// Calculate bid
			svc.CalculateDynamicBid(campaign, req)

			// Record outcome
			won := idx%2 == 0
			svc.RecordOutcome(campaign.ID, req, 2.5, 2.0, won, won, false, 0)

			// Get analytics
			svc.GetBidAnalytics()
		}(i)
	}
	wg.Wait()

	analytics := svc.GetBidAnalytics()
	if analytics["total_bids"].(int64) != 100 {
		t.Errorf("expected 100 bids after concurrent ops, got %v", analytics["total_bids"])
	}
}

func TestDynBid_FactorsInResult(t *testing.T) {
	svc := NewDynamicBidService(nil)
	campaign := createDynBidCampaign()
	req := createDynBidRequest()

	result := svc.CalculateDynamicBid(campaign, req)

	expectedFactors := []string{"time", "device", "publisher", "context", "competition", "goal"}
	for _, f := range expectedFactors {
		if _, exists := result.Factors[f]; !exists {
			t.Errorf("expected factor %s in result", f)
		}
	}
}
