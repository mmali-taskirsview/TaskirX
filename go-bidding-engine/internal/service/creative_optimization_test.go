package service

import (
	"sync"
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

func TestCreativeOpt_NewService(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	if svc == nil {
		t.Fatal("expected service")
	}
	if svc.creativePerf == nil {
		t.Error("expected creative perf map")
	}
	if svc.placementPerf == nil {
		t.Error("expected placement perf map")
	}
}

func createCreativeOptCampaign() *model.Campaign {
	return &model.Campaign{
		ID:          "camp-co-1",
		Status:      "active",
		Budget:      1000,
		DailyBudget: 100,
		BidPrice:    2.0,
		Creative: model.Creative{
			Type: "banner",
		},
		Targeting: model.Targeting{
			CreativeOptimization: &model.CreativeOptimization{
				Enabled:          true,
				OptimizationGoal: "ctr",
				ExplorationRate:  0.1,
				MinImpressions:   100,
				AutoPause:        true,
				PauseThreshold:   0.3,
				CreativePool: []model.CreativeVariant{
					{ID: "creative-1", Name: "Banner A", Status: "active", Weight: 0.5},
					{ID: "creative-2", Name: "Banner B", Status: "active", Weight: 0.5},
				},
			},
		},
	}
}

func createCreativeOptRequest() *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-co-1",
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

func TestCreativeOpt_SelectCreative_Disabled(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	campaign := createCreativeOptCampaign()
	campaign.Targeting.CreativeOptimization.Enabled = false
	req := createCreativeOptRequest()

	result := svc.SelectCreative(campaign, req)

	if result.SelectionMethod != "default" {
		t.Errorf("expected 'default', got '%s'", result.SelectionMethod)
	}
	if result.Reason != "creative_optimization_disabled" {
		t.Errorf("expected reason 'creative_optimization_disabled', got '%s'", result.Reason)
	}
}

func TestCreativeOpt_SelectCreative_NilConfig(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	campaign := createCreativeOptCampaign()
	campaign.Targeting.CreativeOptimization = nil
	req := createCreativeOptRequest()

	result := svc.SelectCreative(campaign, req)

	if result.SelectionMethod != "default" {
		t.Errorf("expected 'default', got '%s'", result.SelectionMethod)
	}
}

func TestCreativeOpt_SelectCreative_EmptyPool(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	campaign := createCreativeOptCampaign()
	campaign.Targeting.CreativeOptimization.CreativePool = []model.CreativeVariant{}
	req := createCreativeOptRequest()

	result := svc.SelectCreative(campaign, req)

	if result.SelectionMethod != "default" {
		t.Errorf("expected 'default', got '%s'", result.SelectionMethod)
	}
}

func TestCreativeOpt_SelectCreative_NoActiveCreatives(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	campaign := createCreativeOptCampaign()
	campaign.Targeting.CreativeOptimization.CreativePool = []model.CreativeVariant{
		{ID: "creative-1", Status: "paused"},
		{ID: "creative-2", Status: "inactive"},
	}
	req := createCreativeOptRequest()

	result := svc.SelectCreative(campaign, req)

	if result.SelectionMethod != "default" {
		t.Errorf("expected 'default', got '%s'", result.SelectionMethod)
	}
	if result.Reason != "no_active_creatives" {
		t.Errorf("expected reason 'no_active_creatives', got '%s'", result.Reason)
	}
}

func TestCreativeOpt_SelectCreative_PlacementRule(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	campaign := createCreativeOptCampaign()
	campaign.Targeting.CreativeOptimization.PlacementRules = []model.PlacementCreativeRule{
		{PlacementType: "banner", CreativeIDs: []string{"creative-1"}},
	}
	req := createCreativeOptRequest()

	result := svc.SelectCreative(campaign, req)

	if result.SelectionMethod != "rule_based" {
		t.Errorf("expected 'rule_based', got '%s'", result.SelectionMethod)
	}
	if result.SelectedCreativeID != "creative-1" {
		t.Errorf("expected 'creative-1', got '%s'", result.SelectedCreativeID)
	}
}

func TestCreativeOpt_SelectCreative_Exploration(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	campaign := createCreativeOptCampaign()
	campaign.Targeting.CreativeOptimization.ExplorationRate = 1.0 // Always explore
	req := createCreativeOptRequest()

	result := svc.SelectCreative(campaign, req)

	if result.SelectionMethod != "exploration" {
		t.Errorf("expected 'exploration', got '%s'", result.SelectionMethod)
	}
}

func TestCreativeOpt_SelectCreative_Exploitation(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	campaign := createCreativeOptCampaign()
	// Note: Service uses 0.1 as default when ExplorationRate <= 0
	// We need enough data to hit the "optimized" path
	campaign.Targeting.CreativeOptimization.ExplorationRate = 0.001 // Very low exploration
	req := createCreativeOptRequest()

	// Add significant performance data to trigger exploitation
	for i := 0; i < 500; i++ {
		svc.RecordImpression("creative-1", "banner")
		svc.RecordImpression("creative-2", "banner")
	}
	// Creative-1 has better CTR
	for i := 0; i < 50; i++ {
		svc.RecordClick("creative-1", "banner")
	}
	for i := 0; i < 10; i++ {
		svc.RecordClick("creative-2", "banner")
	}

	// Run multiple times - most should be optimized with such low exploration rate
	optimizedCount := 0
	for i := 0; i < 50; i++ {
		result := svc.SelectCreative(campaign, req)
		if result.SelectionMethod == "optimized" {
			optimizedCount++
		}
	}

	// With 0.1% exploration rate, expect most to be optimized
	if optimizedCount < 40 {
		t.Errorf("expected most selections to be optimized, got %d/50", optimizedCount)
	}
}

func TestCreativeOpt_RecordImpression(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	svc.RecordImpression("creative-1", "banner")
	svc.RecordImpression("creative-1", "banner")
	svc.RecordImpression("creative-1", "banner")

	perf := svc.GetCreativePerformance("creative-1")
	if perf == nil {
		t.Fatal("expected performance data")
	}
	if perf.Impressions != 3 {
		t.Errorf("expected 3 impressions, got %d", perf.Impressions)
	}
}

func TestCreativeOpt_RecordClick(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	svc.RecordImpression("creative-1", "banner")
	svc.RecordImpression("creative-1", "banner")
	svc.RecordClick("creative-1", "banner")

	perf := svc.GetCreativePerformance("creative-1")
	if perf == nil {
		t.Fatal("expected performance data")
	}
	if perf.Clicks != 1 {
		t.Errorf("expected 1 click, got %d", perf.Clicks)
	}
	if perf.CTR != 0.5 {
		t.Errorf("expected CTR 0.5, got %f", perf.CTR)
	}
}

func TestCreativeOpt_RecordConversion(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	svc.RecordImpression("creative-1", "banner")
	svc.RecordClick("creative-1", "banner")
	svc.RecordConversion("creative-1", "banner")

	perf := svc.GetCreativePerformance("creative-1")
	if perf == nil {
		t.Fatal("expected performance data")
	}
	if perf.Conversions != 1 {
		t.Errorf("expected 1 conversion, got %d", perf.Conversions)
	}
	if perf.CVR != 1.0 {
		t.Errorf("expected CVR 1.0, got %f", perf.CVR)
	}
}

func TestCreativeOpt_RecordEngagement(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	svc.RecordImpression("creative-1", "banner")
	svc.RecordEngagement("creative-1", "banner", 15.5)

	perf := svc.GetCreativePerformance("creative-1")
	if perf == nil {
		t.Fatal("expected performance data")
	}
	if perf.EngagementRate != 1.0 {
		t.Errorf("expected engagement rate 1.0, got %f", perf.EngagementRate)
	}
	if perf.AvgTimeViewed != 15.5 {
		t.Errorf("expected avg view time 15.5, got %f", perf.AvgTimeViewed)
	}
}

func TestCreativeOpt_GetCreativePerformance_NotFound(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	perf := svc.GetCreativePerformance("nonexistent")

	if perf != nil {
		t.Error("expected nil for nonexistent creative")
	}
}

func TestCreativeOpt_CheckAutoPause_Disabled(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	config := &model.CreativeOptimization{
		AutoPause: false,
	}

	result := svc.CheckAutoPause(config)

	if result != nil {
		t.Error("expected nil when auto pause disabled")
	}
}

func TestCreativeOpt_CheckAutoPause_NilConfig(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	result := svc.CheckAutoPause(nil)

	if result != nil {
		t.Error("expected nil for nil config")
	}
}

func TestCreativeOpt_CheckAutoPause_PoorPerformers(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	// Record good performance for creative-1
	for i := 0; i < 100; i++ {
		svc.RecordImpression("creative-1", "banner")
		if i%5 == 0 {
			svc.RecordClick("creative-1", "banner")
		}
	}

	// Record poor performance for creative-2
	for i := 0; i < 100; i++ {
		svc.RecordImpression("creative-2", "banner")
		if i%100 == 0 {
			svc.RecordClick("creative-2", "banner") // Very low CTR
		}
	}

	config := &model.CreativeOptimization{
		AutoPause:        true,
		PauseThreshold:   0.3,
		OptimizationGoal: "ctr",
	}

	result := svc.CheckAutoPause(config)

	if len(result) == 0 {
		t.Error("expected poor performer to be flagged for pause")
	}
}

func TestCreativeOpt_CheckAutoPause_InsufficientData(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	// Record only 10 impressions (below threshold of 100)
	for i := 0; i < 10; i++ {
		svc.RecordImpression("creative-1", "banner")
	}

	config := &model.CreativeOptimization{
		AutoPause:        true,
		PauseThreshold:   0.3,
		OptimizationGoal: "ctr",
	}

	result := svc.CheckAutoPause(config)

	if result != nil {
		t.Error("expected nil when insufficient data")
	}
}

func TestCreativeOpt_FilterActiveCreatives(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	pool := []model.CreativeVariant{
		{ID: "c1", Status: "active"},
		{ID: "c2", Status: "testing"},
		{ID: "c3", Status: "paused"},
		{ID: "c4", Status: ""}, // Empty = active
		{ID: "c5", Status: "inactive"},
	}

	active := svc.filterActiveCreatives(pool)

	if len(active) != 3 {
		t.Errorf("expected 3 active creatives, got %d", len(active))
	}
}

func TestCreativeOpt_DetectPlacementType_FromFormats(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Formats: []string{"video", "banner"},
		},
	}

	placementType := svc.detectPlacementType(req)

	if placementType != "video" {
		t.Errorf("expected 'video' from formats, got '%s'", placementType)
	}
}

func TestCreativeOpt_DetectPlacementType_WideRatio(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Dimensions: []int{728, 90}, // Wide banner
		},
	}

	placementType := svc.detectPlacementType(req)

	if placementType != "banner" {
		t.Errorf("expected 'banner' for wide ratio, got '%s'", placementType)
	}
}

func TestCreativeOpt_DetectPlacementType_TallRatio(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Dimensions: []int{320, 480}, // Tall interstitial
		},
	}

	placementType := svc.detectPlacementType(req)

	if placementType != "interstitial" {
		t.Errorf("expected 'interstitial' for tall ratio, got '%s'", placementType)
	}
}

func TestCreativeOpt_CalculateScore_CTR(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	perf := &creativePerformance{
		impressions: 100,
		clicks:      10,
	}

	score := svc.calculateCreativeScore(perf, "ctr")

	if score != 0.1 {
		t.Errorf("expected CTR score 0.1, got %f", score)
	}
}

func TestCreativeOpt_CalculateScore_CVR(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	perf := &creativePerformance{
		impressions: 100,
		clicks:      10,
		conversions: 2,
	}

	score := svc.calculateCreativeScore(perf, "cvr")

	if score != 0.2 {
		t.Errorf("expected CVR score 0.2, got %f", score)
	}
}

func TestCreativeOpt_CalculateScore_Engagement(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	perf := &creativePerformance{
		impressions: 100,
		engagements: 50,
	}

	score := svc.calculateCreativeScore(perf, "engagement")

	if score != 0.5 {
		t.Errorf("expected engagement score 0.5, got %f", score)
	}
}

func TestCreativeOpt_CalculateScore_Viewability(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	perf := &creativePerformance{
		impressions: 100,
		viewTime:    1500.0, // 15 seconds average
	}

	score := svc.calculateCreativeScore(perf, "viewability")

	if score != 0.5 {
		t.Errorf("expected viewability score 0.5, got %f", score)
	}
}

func TestCreativeOpt_CalculateScore_Composite(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	perf := &creativePerformance{
		impressions: 100,
		clicks:      10,
		conversions: 2,
		engagements: 20,
		viewTime:    1500.0,
	}

	score := svc.calculateCreativeScore(perf, "composite")

	if score <= 0 {
		t.Error("expected positive composite score")
	}
}

func TestCreativeOpt_CalculateScore_ZeroImpressions(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	perf := &creativePerformance{
		impressions: 0,
	}

	score := svc.calculateCreativeScore(perf, "ctr")

	if score != 0 {
		t.Errorf("expected 0 for zero impressions, got %f", score)
	}
}

func TestCreativeOpt_ExploitWithInsufficientData(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	campaign := createCreativeOptCampaign()
	campaign.Targeting.CreativeOptimization.ExplorationRate = 0.0
	req := createCreativeOptRequest()

	// Record minimal data (below minImpressions)
	for i := 0; i < 10; i++ {
		svc.RecordImpression("creative-1", "banner")
	}

	result := svc.SelectCreative(campaign, req)

	// Should still select something
	if result.SelectedCreativeID == "" {
		t.Error("expected a creative selection")
	}
}

func TestCreativeOpt_PlacementTracking(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	// Record to different placements
	svc.RecordImpression("creative-1", "banner")
	svc.RecordImpression("creative-1", "banner")
	svc.RecordImpression("creative-1", "sidebar")
	svc.RecordClick("creative-1", "banner")

	// Verify global tracking
	perf := svc.GetCreativePerformance("creative-1")
	if perf.Impressions != 3 {
		t.Errorf("expected 3 total impressions, got %d", perf.Impressions)
	}
	if perf.Clicks != 1 {
		t.Errorf("expected 1 click, got %d", perf.Clicks)
	}

	// Verify placement-specific tracking
	svc.mu.RLock()
	bannerPerf := svc.placementPerf["banner"]["creative-1"]
	svc.mu.RUnlock()

	if bannerPerf.impressions != 2 {
		t.Errorf("expected 2 banner impressions, got %d", bannerPerf.impressions)
	}
}

func TestCreativeOpt_Concurrency(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			creativeID := "creative-1"
			if idx%2 == 0 {
				creativeID = "creative-2"
			}
			svc.RecordImpression(creativeID, "banner")
			if idx%5 == 0 {
				svc.RecordClick(creativeID, "banner")
			}
			if idx%20 == 0 {
				svc.RecordConversion(creativeID, "banner")
			}
		}(i)
	}
	wg.Wait()

	perf1 := svc.GetCreativePerformance("creative-1")
	perf2 := svc.GetCreativePerformance("creative-2")

	if perf1 == nil || perf2 == nil {
		t.Fatal("expected performance data for both creatives")
	}

	total := perf1.Impressions + perf2.Impressions
	if total != 100 {
		t.Errorf("expected 100 total impressions, got %d", total)
	}
}

func TestCreativeOpt_AlternativeCreatives(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	campaign := createCreativeOptCampaign()
	campaign.Targeting.CreativeOptimization.ExplorationRate = 0.0
	campaign.Targeting.CreativeOptimization.CreativePool = []model.CreativeVariant{
		{ID: "creative-1", Status: "active"},
		{ID: "creative-2", Status: "active"},
		{ID: "creative-3", Status: "active"},
		{ID: "creative-4", Status: "active"},
	}
	req := createCreativeOptRequest()

	// Add performance data for all creatives
	for i := 0; i < 200; i++ {
		svc.RecordImpression("creative-1", "banner")
		svc.RecordImpression("creative-2", "banner")
		svc.RecordImpression("creative-3", "banner")
		svc.RecordImpression("creative-4", "banner")
	}
	for i := 0; i < 40; i++ {
		svc.RecordClick("creative-1", "banner") // Best
	}
	for i := 0; i < 30; i++ {
		svc.RecordClick("creative-2", "banner")
	}
	for i := 0; i < 20; i++ {
		svc.RecordClick("creative-3", "banner")
	}
	for i := 0; i < 10; i++ {
		svc.RecordClick("creative-4", "banner") // Worst
	}

	result := svc.SelectCreative(campaign, req)

	if result.SelectedCreativeID != "creative-1" {
		t.Errorf("expected best performer 'creative-1', got '%s'", result.SelectedCreativeID)
	}
	if len(result.AlternativeIDs) != 3 {
		t.Errorf("expected 3 alternatives, got %d", len(result.AlternativeIDs))
	}
}

func TestCreativeOpt_DefaultExplorationRate(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	campaign := createCreativeOptCampaign()
	campaign.Targeting.CreativeOptimization.ExplorationRate = 0 // Will default to 0.1
	req := createCreativeOptRequest()

	// Run multiple selections and verify exploration happens sometimes
	explorationCount := 0
	for i := 0; i < 100; i++ {
		result := svc.SelectCreative(campaign, req)
		if result.SelectionMethod == "exploration" {
			explorationCount++
		}
	}

	// With 10% exploration, expect roughly 5-15 explorations
	if explorationCount < 1 || explorationCount > 50 {
		t.Errorf("unexpected exploration count: %d", explorationCount)
	}
}
