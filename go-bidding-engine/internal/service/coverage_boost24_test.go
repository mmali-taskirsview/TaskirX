package service

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================
// checkSpendSpike — seed baselineSpend and trigger spike
// ============================================================

func TestCheckSpendSpike_NoHistory_B24(t *testing.T) {
	svc := NewRealTimeAlertService(NewMockCache())

	campaign := &model.Campaign{
		ID: "camp-spike-none",
		Targeting: model.Targeting{
			AlertConfig: &model.AlertConfig{
				Enabled: true,
				BudgetAlerts: &model.BudgetAlerts{
					Enabled:         true,
					UnexpectedSpike: true,
				},
			},
		},
	}
	// No campaignMetrics entry → spike check returns early without alert
	result := svc.CheckAlerts(campaign, 100.0, 500.0)
	if result.HasActiveAlerts {
		t.Error("Expected no active alerts with no history")
	}
}

func TestCheckSpendSpike_BelowThreshold_B24(t *testing.T) {
	svc := NewRealTimeAlertService(NewMockCache())

	campaign := &model.Campaign{
		ID: "camp-spike-low",
		Targeting: model.Targeting{
			AlertConfig: &model.AlertConfig{
				Enabled: true,
				BudgetAlerts: &model.BudgetAlerts{
					Enabled:         true,
					UnexpectedSpike: true,
					SpikeThreshold:  50.0,
				},
			},
		},
	}
	// baselineSpend = 2400 → hourlyAvg = 100.0; currentSpend=120 → spike=20% < 50%
	svc.mu.Lock()
	svc.campaignMetrics[campaign.ID] = &campaignMetricsHistory{
		baselineSpend: 2400.0,
	}
	svc.mu.Unlock()

	result := svc.CheckAlerts(campaign, 120.0, 500.0)
	if result.HasActiveAlerts {
		t.Error("Expected no alert when spike is below threshold")
	}
}

func TestCheckSpendSpike_AboveThreshold_B24(t *testing.T) {
	svc := NewRealTimeAlertService(NewMockCache())

	campaign := &model.Campaign{
		ID: "camp-spike-high",
		Targeting: model.Targeting{
			AlertConfig: &model.AlertConfig{
				Enabled: true,
				BudgetAlerts: &model.BudgetAlerts{
					Enabled:         true,
					UnexpectedSpike: true,
					SpikeThreshold:  30.0,
				},
			},
		},
	}
	// baselineSpend = 240 → hourlyAvg = 10.0; currentSpend=20 → spike=100% > 30%
	svc.mu.Lock()
	svc.campaignMetrics[campaign.ID] = &campaignMetricsHistory{
		baselineSpend: 240.0,
	}
	svc.mu.Unlock()

	svc.CheckAlerts(campaign, 20.0, 200.0)
	// Alert should have been created
	svc.mu.RLock()
	var found bool
	for _, alert := range svc.activeAlerts {
		if alert.CampaignID == campaign.ID && alert.Metric == "spend_spike" {
			found = true
		}
	}
	svc.mu.RUnlock()
	if !found {
		t.Error("Expected spend_spike alert to be created")
	}
}

func TestCheckSpendSpike_DefaultThreshold_B24(t *testing.T) {
	svc := NewRealTimeAlertService(NewMockCache())

	campaign := &model.Campaign{
		ID: "camp-spike-default",
		Targeting: model.Targeting{
			AlertConfig: &model.AlertConfig{
				Enabled: true,
				BudgetAlerts: &model.BudgetAlerts{
					Enabled:         true,
					UnexpectedSpike: true,
					SpikeThreshold:  0, // → default 50.0
				},
			},
		},
	}
	// hourlyAvg = 1200/24 = 50; currentSpend=100 → spike = 100% > 50%
	svc.mu.Lock()
	svc.campaignMetrics[campaign.ID] = &campaignMetricsHistory{
		baselineSpend: 1200.0,
	}
	svc.mu.Unlock()

	svc.CheckAlerts(campaign, 100.0, 1200.0)
	svc.mu.RLock()
	var found bool
	for _, alert := range svc.activeAlerts {
		if alert.CampaignID == campaign.ID && alert.Metric == "spend_spike" {
			found = true
		}
	}
	svc.mu.RUnlock()
	if !found {
		t.Error("Expected spend_spike alert with default threshold")
	}
}

// ============================================================
// matchesInventorySpec — exercise all filter branches
// ============================================================

func TestMatchesInventorySpec_AllFiltersMatch_B24(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)
	spec := InventorySpec{
		PublisherIDs: []string{"pub1"},
		SiteIDs:      []string{"site1"},
		Placements:   []string{"header"},
		AdFormats:    []string{"banner"},
		DeviceTypes:  []string{"desktop"},
		GeoTargets:   []string{"US"},
	}
	if !svc.matchesInventorySpec(spec, "pub1", "site1", "header", "banner", "desktop", "US") {
		t.Error("Expected all filters to match")
	}
}

func TestMatchesInventorySpec_PublisherMismatch_B24(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)
	spec := InventorySpec{
		PublisherIDs: []string{"pub1"},
	}
	if svc.matchesInventorySpec(spec, "pub2", "", "", "", "", "") {
		t.Error("Expected publisher mismatch to return false")
	}
}

func TestMatchesInventorySpec_SiteMismatch_B24(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)
	spec := InventorySpec{
		SiteIDs: []string{"site1"},
	}
	if svc.matchesInventorySpec(spec, "", "site2", "", "", "", "") {
		t.Error("Expected site mismatch to return false")
	}
}

func TestMatchesInventorySpec_PlacementMismatch_B24(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)
	spec := InventorySpec{
		Placements: []string{"header"},
	}
	if svc.matchesInventorySpec(spec, "", "", "footer", "", "", "") {
		t.Error("Expected placement mismatch to return false")
	}
}

func TestMatchesInventorySpec_AdFormatMismatch_B24(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)
	spec := InventorySpec{
		AdFormats: []string{"banner"},
	}
	if svc.matchesInventorySpec(spec, "", "", "", "video", "", "") {
		t.Error("Expected ad format mismatch to return false")
	}
}

func TestMatchesInventorySpec_DeviceTypeMismatch_B24(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)
	spec := InventorySpec{
		DeviceTypes: []string{"desktop"},
	}
	if svc.matchesInventorySpec(spec, "", "", "", "", "mobile", "") {
		t.Error("Expected device type mismatch to return false")
	}
}

func TestMatchesInventorySpec_GeoMismatch_B24(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)
	spec := InventorySpec{
		GeoTargets: []string{"US"},
	}
	if svc.matchesInventorySpec(spec, "", "", "", "", "", "UK") {
		t.Error("Expected geo mismatch to return false")
	}
}

func TestMatchesInventorySpec_NoFilters_B24(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)
	spec := InventorySpec{}
	if !svc.matchesInventorySpec(spec, "pub1", "site1", "header", "banner", "desktop", "US") {
		t.Error("Expected empty spec to match everything")
	}
}

// ============================================================
// checkPlacementRules — rule match and no-match paths
// ============================================================

func TestCheckPlacementRules_RuleMatch_B24(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	rules := []model.PlacementCreativeRule{
		{
			PlacementType: "banner",
			CreativeIDs:   []string{"cr-1", "cr-2"},
		},
	}
	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Formats: []string{"banner"},
		},
	}
	creatives := []model.CreativeVariant{
		{ID: "cr-1", Status: "active"},
		{ID: "cr-2", Status: "active"},
	}
	result := svc.checkPlacementRules(rules, req, creatives)
	if result == nil {
		t.Fatal("Expected a rule-based result, got nil")
	}
	if result.SelectionMethod != "rule_based" {
		t.Errorf("Expected 'rule_based', got '%s'", result.SelectionMethod)
	}
}

func TestCheckPlacementRules_NoMatchingCreative_B24(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	rules := []model.PlacementCreativeRule{
		{
			PlacementType: "video",
			CreativeIDs:   []string{"cr-99"}, // not in creatives list
		},
	}
	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Formats: []string{"video"},
		},
	}
	creatives := []model.CreativeVariant{
		{ID: "cr-1", Status: "active"},
	}
	result := svc.checkPlacementRules(rules, req, creatives)
	if result != nil {
		t.Error("Expected nil when no creative ID matches")
	}
}

func TestCheckPlacementRules_EmptyRules_B24(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	result := svc.checkPlacementRules(nil, &model.BidRequest{}, nil)
	if result != nil {
		t.Error("Expected nil for empty rules")
	}
}

// ============================================================
// RecordConversion (CreativeOptimizationService) — new placement
// ============================================================

func TestCreativeRecordConversion_NewPlacement_B24(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	// Pre-seed placement map with placement but without this creative
	svc.mu.Lock()
	svc.placementPerf["home"] = map[string]*creativePerformance{}
	svc.mu.Unlock()

	svc.RecordConversion("cr-conv-1", "home")

	svc.mu.RLock()
	perf := svc.placementPerf["home"]["cr-conv-1"]
	svc.mu.RUnlock()
	if perf == nil || perf.conversions != 1 {
		t.Error("Expected 1 conversion recorded for new placement creative")
	}
}

func TestCreativeRecordConversion_UnknownPlacement_B24(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	// Placement not in map → placement branch skipped, but global perf should be updated
	svc.RecordConversion("cr-conv-2", "unknown-placement")

	svc.mu.RLock()
	perf := svc.creativePerf["cr-conv-2"]
	svc.mu.RUnlock()
	if perf == nil || perf.conversions != 1 {
		t.Error("Expected 1 conversion in global creative perf")
	}
}

// ============================================================
// parseDaypartCache — valid and invalid strings
// ============================================================

func TestParseDaypartCache_ValidString_B24(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())

	cached := "impressions:500,clicks:25,conversions:5,spend:12.5000,ctr:0.050000,cvr:0.200000,win_rate:0.450000"
	perf := svc.parseDaypartCache(cached)

	if perf.impressions != 500 {
		t.Errorf("Expected 500 impressions, got %d", perf.impressions)
	}
	if perf.clicks != 25 {
		t.Errorf("Expected 25 clicks, got %d", perf.clicks)
	}
	if perf.conversions != 5 {
		t.Errorf("Expected 5 conversions, got %d", perf.conversions)
	}
	if perf.winRate != 0.45 {
		t.Errorf("Expected win_rate 0.45, got %f", perf.winRate)
	}
}

func TestParseDaypartCache_InvalidPairs_B24(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	// Malformed pair without colon
	perf := svc.parseDaypartCache("badentry,impressions:100")
	if perf.impressions != 100 {
		t.Errorf("Expected impressions=100, got %d", perf.impressions)
	}
}

func TestParseDaypartCache_Empty_B24(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	perf := svc.parseDaypartCache("")
	if perf.impressions != 0 {
		t.Error("Expected zero impressions for empty cache string")
	}
}

// ============================================================
// incrementDaypartMetric — click, conversion, win, impression
// ============================================================

func TestIncrementDaypartMetric_Click_B24(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	current := "impressions:100,clicks:10,conversions:2,spend:5.0000,ctr:0.100000,cvr:0.200000"
	result := svc.incrementDaypartMetric(current, "click")
	after := svc.parseDaypartCache(result)
	if after.clicks != 11 {
		t.Errorf("Expected 11 clicks, got %d", after.clicks)
	}
}

func TestIncrementDaypartMetric_Conversion_B24(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	current := "impressions:100,clicks:10,conversions:2,spend:5.0000,ctr:0.100000,cvr:0.200000"
	result := svc.incrementDaypartMetric(current, "conversion")
	after := svc.parseDaypartCache(result)
	if after.conversions != 3 {
		t.Errorf("Expected 3 conversions, got %d", after.conversions)
	}
}

func TestIncrementDaypartMetric_Win_B24(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	result := svc.incrementDaypartMetric("impressions:50,clicks:5,conversions:1,spend:0.0000,ctr:0.100000,cvr:0.200000", "win")
	// win just passes through, no counter incremented
	after := svc.parseDaypartCache(result)
	if after.impressions != 50 {
		t.Errorf("Expected impressions unchanged at 50, got %d", after.impressions)
	}
}

func TestIncrementDaypartMetric_Impression_B24(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	result := svc.incrementDaypartMetric("", "impression")
	after := svc.parseDaypartCache(result)
	if after.impressions != 1 {
		t.Errorf("Expected 1 impression, got %d", after.impressions)
	}
}

// ============================================================
// getDayOfWeekModifier — with valid day data and ratio branches
// ============================================================

type mockCacheForDaypart struct {
	*MockCache
	data map[string]string
}

func (m *mockCacheForDaypart) Get(key string) (string, error) {
	if v, ok := m.data[key]; ok {
		return v, nil
	}
	return "", nil
}

func TestGetDayOfWeekModifier_RatioAbove1_2_B24(t *testing.T) {
	mc := &mockCacheForDaypart{
		MockCache: NewMockCache(),
		data:      map[string]string{},
	}
	// dayPerf CTR=0.20, impressions=100 → > 50 → ratio = 0.20 / 0.10 = 2.0 → clamped 1.2
	mc.data["daypart_dow:camp-dow:1"] = "impressions:100,clicks:20,conversions:2,spend:5.0000,ctr:0.200000,cvr:0.200000"

	svc := NewDaypartingService(mc)
	avgPerf := hourlyPerformance{ctr: 0.10, impressions: 200}
	mod := svc.getDayOfWeekModifier("camp-dow", 1, 14, avgPerf)
	if mod != 1.2 {
		t.Errorf("Expected 1.2 cap, got %f", mod)
	}
}

func TestGetDayOfWeekModifier_RatioBelow0_8_B24(t *testing.T) {
	mc := &mockCacheForDaypart{
		MockCache: NewMockCache(),
		data:      map[string]string{},
	}
	// dayPerf CTR=0.02, impressions=100 → ratio = 0.02 / 0.10 = 0.2 → clamped 0.8
	mc.data["daypart_dow:camp-dow2:2"] = "impressions:100,clicks:2,conversions:0,spend:0.0000,ctr:0.020000,cvr:0.000000"

	svc := NewDaypartingService(mc)
	avgPerf := hourlyPerformance{ctr: 0.10, impressions: 200}
	mod := svc.getDayOfWeekModifier("camp-dow2", 2, 14, avgPerf)
	if mod != 0.8 {
		t.Errorf("Expected 0.8 floor, got %f", mod)
	}
}

func TestGetDayOfWeekModifier_NoCacheData_B24(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	avgPerf := hourlyPerformance{ctr: 0.10}
	mod := svc.getDayOfWeekModifier("camp-missing", 3, 14, avgPerf)
	if mod != 1.0 {
		t.Errorf("Expected 1.0 for missing cache, got %f", mod)
	}
}

func TestGetDayOfWeekModifier_InsufficientImpressions_B24(t *testing.T) {
	mc := &mockCacheForDaypart{
		MockCache: NewMockCache(),
		data:      map[string]string{},
	}
	// impressions=40 < 50 → returns 1.0
	mc.data["daypart_dow:camp-insuf:0"] = "impressions:40,clicks:2,conversions:0,spend:0.0000,ctr:0.050000,cvr:0.000000"
	svc := NewDaypartingService(mc)
	avgPerf := hourlyPerformance{ctr: 0.05}
	mod := svc.getDayOfWeekModifier("camp-insuf", 0, 14, avgPerf)
	if mod != 1.0 {
		t.Errorf("Expected 1.0 for insufficient impressions, got %f", mod)
	}
}

// ============================================================
// GetOptimalHours — with and without cache data
// ============================================================

func TestGetOptimalHours_WithData_B24(t *testing.T) {
	mc := &mockCacheForDaypart{
		MockCache: NewMockCache(),
		data:      map[string]string{},
	}
	// Seed hour 10 and hour 14 with enough impressions
	mc.data["daypart_perf:camp-opt:10"] = "impressions:500,clicks:50,conversions:10,spend:25.0000,ctr:0.100000,cvr:0.200000"
	mc.data["daypart_perf:camp-opt:14"] = "impressions:600,clicks:30,conversions:6,spend:20.0000,ctr:0.050000,cvr:0.200000"
	// avg cache missing → computed from hours
	svc := NewDaypartingService(mc)
	results := svc.GetOptimalHours("camp-opt", 3)
	// Should return up to 2 hours (only 2 have data) and topN=3 won't truncate
	if len(results) == 0 {
		t.Error("Expected at least 1 optimal hour result")
	}
}

func TestGetOptimalHours_NoData_B24(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	results := svc.GetOptimalHours("camp-nodata", 6)
	if len(results) != 0 {
		t.Errorf("Expected 0 results with no cache data, got %d", len(results))
	}
}

func TestGetOptimalHours_DefaultTopN_B24(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	// topN=0 → defaults to 6; no cache data → empty
	results := svc.GetOptimalHours("camp-default", 0)
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

// ============================================================
// ListPublishers — with status and quality filters
// ============================================================

func TestListPublishers_StatusFilter_B24(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub1 := &DirectPublisher{ID: "pub-1", Status: "active", QualityScore: 0.8}
	pub2 := &DirectPublisher{ID: "pub-2", Status: "suspended", QualityScore: 0.9}
	svc.publishers.Store(pub1.ID, pub1)
	svc.publishers.Store(pub2.ID, pub2)

	results := svc.ListPublishers("active", 0.0)
	if len(results) != 1 || results[0].ID != "pub-1" {
		t.Errorf("Expected 1 active publisher, got %d", len(results))
	}
}

func TestListPublishers_QualityFilter_B24(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub1 := &DirectPublisher{ID: "pub-3", Status: "active", QualityScore: 0.5}
	pub2 := &DirectPublisher{ID: "pub-4", Status: "active", QualityScore: 0.9}
	svc.publishers.Store(pub1.ID, pub1)
	svc.publishers.Store(pub2.ID, pub2)

	results := svc.ListPublishers("", 0.7)
	if len(results) != 1 || results[0].ID != "pub-4" {
		t.Errorf("Expected 1 high quality publisher, got %d", len(results))
	}
}

func TestListPublishers_SortedByQuality_B24(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub1 := &DirectPublisher{ID: "pub-5", Status: "active", QualityScore: 0.6}
	pub2 := &DirectPublisher{ID: "pub-6", Status: "active", QualityScore: 0.95}
	svc.publishers.Store(pub1.ID, pub1)
	svc.publishers.Store(pub2.ID, pub2)

	results := svc.ListPublishers("active", 0.0)
	if len(results) < 2 || results[0].QualityScore < results[1].QualityScore {
		t.Error("Expected publishers sorted by quality score descending")
	}
}

// ============================================================
// getAveragePerformance — builds from hour data when no avg cache
// ============================================================

func TestGetAveragePerformance_BuildsFromHours_B24(t *testing.T) {
	mc := &mockCacheForDaypart{
		MockCache: NewMockCache(),
		data:      map[string]string{},
	}
	// Seed two hours with data
	mc.data["daypart_perf:camp-avgbuild:8"] = "impressions:200,clicks:20,conversions:4,spend:10.0000,ctr:0.100000,cvr:0.200000"
	mc.data["daypart_perf:camp-avgbuild:12"] = "impressions:400,clicks:40,conversions:8,spend:20.0000,ctr:0.100000,cvr:0.200000"
	svc := NewDaypartingService(mc)

	avg := svc.getAveragePerformance("camp-avgbuild", 14)
	if avg.impressions == 0 {
		t.Error("Expected non-zero average impressions built from hourly data")
	}
}

func TestGetAveragePerformance_FromCache_B24(t *testing.T) {
	mc := &mockCacheForDaypart{
		MockCache: NewMockCache(),
		data:      map[string]string{},
	}
	mc.data["daypart_avg:camp-avgcache"] = "impressions:300,clicks:30,conversions:6,spend:15.0000,ctr:0.100000,cvr:0.200000"
	svc := NewDaypartingService(mc)

	avg := svc.getAveragePerformance("camp-avgcache", 14)
	if avg.impressions != 300 {
		t.Errorf("Expected 300 from cache, got %d", avg.impressions)
	}
}

// ============================================================
// CheckAutoPause (CreativeOptimizationService)
// ============================================================

func TestCheckAutoPause_BelowThreshold_B24(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	// Seed creatives: cr-good has high score, cr-bad has low score
	svc.mu.Lock()
	svc.creativePerf["cr-good"] = &creativePerformance{impressions: 200, clicks: 40, conversions: 10, score: 0.8}
	svc.creativePerf["cr-bad"] = &creativePerformance{impressions: 150, clicks: 3, conversions: 0, score: 0.1}
	svc.mu.Unlock()

	config := &model.CreativeOptimization{
		AutoPause:        true,
		PauseThreshold:   0.3,
		OptimizationGoal: "ctr",
	}
	toPause := svc.CheckAutoPause(config)
	if len(toPause) == 0 {
		t.Error("Expected at least one creative to be paused")
	}
}

func TestCheckAutoPause_NilConfig_B24(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	result := svc.CheckAutoPause(nil)
	if result != nil {
		t.Error("Expected nil for nil config")
	}
}

func TestCheckAutoPause_AutoPauseDisabled_B24(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	config := &model.CreativeOptimization{AutoPause: false}
	result := svc.CheckAutoPause(config)
	if result != nil {
		t.Error("Expected nil when auto-pause disabled")
	}
}

func TestCheckAutoPause_NoCreativesWithMinImpressions_B24(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)
	svc.mu.Lock()
	svc.creativePerf["cr-low"] = &creativePerformance{impressions: 50} // < 100 threshold
	svc.mu.Unlock()

	config := &model.CreativeOptimization{AutoPause: true}
	result := svc.CheckAutoPause(config)
	if result != nil {
		t.Error("Expected nil when no creatives meet min impressions")
	}
}

// ============================================================
// CalculateDaypartMultiplier — manual and day-specific paths
// ============================================================

func TestCalculateDaypartMultiplier_ManualHourly_B24(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())

	now := time.Now()
	hour := now.Hour()

	campaign := &model.Campaign{
		ID: "camp-manual",
		Targeting: model.Targeting{
			PerformanceGoals: &model.PerformanceGoals{
				DaypartingOptimization: &model.DaypartingOptimization{
					Enabled:           true,
					HourlyMultipliers: map[int]float64{hour: 1.5},
					MinMultiplier:     0.3,
					MaxMultiplier:     2.0,
				},
			},
		},
	}
	result := svc.CalculateDaypartMultiplier(campaign, &model.BidRequest{})
	if result.Multiplier != 1.5 {
		t.Errorf("Expected manual multiplier 1.5, got %f", result.Multiplier)
	}
}

func TestCalculateDaypartMultiplier_DaySpecific_B24(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())

	now := time.Now()
	hour := now.Hour()
	dayName := dayNames[int(now.Weekday())]

	campaign := &model.Campaign{
		ID: "camp-day",
		Targeting: model.Targeting{
			PerformanceGoals: &model.PerformanceGoals{
				DaypartingOptimization: &model.DaypartingOptimization{
					Enabled: true,
					DaySpecific: map[string]map[int]float64{
						dayName: {hour: 1.3},
					},
					MinMultiplier: 0.3,
					MaxMultiplier: 2.0,
				},
			},
		},
	}
	result := svc.CalculateDaypartMultiplier(campaign, &model.BidRequest{})
	if result.Multiplier != 1.3 {
		t.Errorf("Expected day-specific multiplier 1.3, got %f", result.Multiplier)
	}
}
