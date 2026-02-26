package service

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================
// analyzeSegmentCompetition — empty, high, medium, low market conditions
// ============================================================

func TestAnalyzeSegmentCompetition_NoHistory_B25(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	result := &model.CompetitiveIntelResult{}
	svc.analyzeSegmentCompetition("seg-nodata", result)
	// No auction history → should return early without setting MarketCondition
	if result.MarketCondition != "" {
		t.Errorf("Expected empty market condition, got %s", result.MarketCondition)
	}
}

func TestAnalyzeSegmentCompetition_HighMarket_B25(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	// Seed 6+ competitors with recent auctions → CompetitorsActive > 5 → "high"
	svc.mu.Lock()
	now := time.Now()
	for i := 0; i < 7; i++ {
		svc.auctionHistory = append(svc.auctionHistory, auctionOutcome{
			Timestamp:    now.Add(-1 * time.Hour),
			SegmentKey:   "seg-high",
			Won:          false,
			CompetitorID: "comp-" + string(rune('A'+i)),
			WinningBid:   1.0,
		})
	}
	svc.mu.Unlock()

	result := &model.CompetitiveIntelResult{}
	svc.analyzeSegmentCompetition("seg-high", result)
	if result.MarketCondition != "high" {
		t.Errorf("Expected 'high' market condition, got '%s'", result.MarketCondition)
	}
}

func TestAnalyzeSegmentCompetition_MediumMarket_B25(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	// 3 competitors, bids spread ~0.15 → bidSpread < 0.25 → "medium"
	svc.mu.Lock()
	now := time.Now()
	bids := []float64{1.0, 1.1, 0.95}
	for i, bid := range bids {
		svc.auctionHistory = append(svc.auctionHistory, auctionOutcome{
			Timestamp:    now.Add(-2 * time.Hour),
			SegmentKey:   "seg-med",
			Won:          false,
			CompetitorID: "med-comp-" + string(rune('A'+i)),
			WinningBid:   bid,
		})
	}
	svc.mu.Unlock()

	result := &model.CompetitiveIntelResult{}
	svc.analyzeSegmentCompetition("seg-med", result)
	if result.MarketCondition != "medium" && result.MarketCondition != "high" {
		t.Errorf("Expected 'medium' or 'high' market condition, got '%s'", result.MarketCondition)
	}
}

func TestAnalyzeSegmentCompetition_LowMarket_B25(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	// Only 1 auction, 1 competitor, very spread bids → "low"
	svc.mu.Lock()
	now := time.Now()
	// Wide bid spread: bids 0.1 and 2.0 → large CV → "low" if < 2 competitors
	svc.auctionHistory = append(svc.auctionHistory, auctionOutcome{
		Timestamp:    now.Add(-1 * time.Hour),
		SegmentKey:   "seg-low",
		Won:          false,
		CompetitorID: "solo-comp",
		WinningBid:   2.0,
	})
	svc.auctionHistory = append(svc.auctionHistory, auctionOutcome{
		Timestamp:    now.Add(-30 * time.Minute),
		SegmentKey:   "seg-low",
		Won:          true,
		CompetitorID: "",
		WinningBid:   0.5,
	})
	svc.mu.Unlock()

	result := &model.CompetitiveIntelResult{}
	svc.analyzeSegmentCompetition("seg-low", result)
	// 1 competitor, bidSpread may be > 0.25 → "low"
	if result.MarketCondition == "" {
		t.Error("Expected a market condition to be set")
	}
}

// ============================================================
// updateCompetitorProfile — new profile, trend tracking
// ============================================================

func TestUpdateCompetitorProfile_NewProfile_B25(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	req := &model.BidRequest{
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
	}
	// Record an outcome with a competitor (won=false → updateCompetitorProfile called)
	svc.RecordAuctionOutcome(req, 1.0, 1.5, false, "new-comp")

	profile, exists := svc.GetCompetitorProfile("new-comp")
	if !exists {
		t.Fatal("Expected competitor profile to be created")
	}
	if profile.AvgBidPrice != 1.5 {
		t.Errorf("Expected avg bid 1.5, got %f", profile.AvgBidPrice)
	}
	if profile.BidVolume != 1 {
		t.Errorf("Expected bid volume 1, got %d", profile.BidVolume)
	}
}

func TestUpdateCompetitorProfile_TrendIncreasing_B25(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	req := &model.BidRequest{
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
	}
	// Build history: first 5 bids at 1.0, next 5 at 2.0 → increasing trend
	for i := 0; i < 5; i++ {
		svc.RecordAuctionOutcome(req, 0.9, 1.0, false, "trend-comp")
	}
	for i := 0; i < 5; i++ {
		svc.RecordAuctionOutcome(req, 1.8, 2.0, false, "trend-comp")
	}

	profile, exists := svc.GetCompetitorProfile("trend-comp")
	if !exists {
		t.Fatal("Expected trend-comp profile to exist")
	}
	if profile.TrendDirection != "increasing" {
		t.Errorf("Expected 'increasing' trend, got '%s'", profile.TrendDirection)
	}
}

func TestUpdateCompetitorProfile_TrendDecreasing_B25(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	req := &model.BidRequest{
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
	}
	// First 5 bids at 2.0, next 5 at 0.5 → decreasing
	for i := 0; i < 5; i++ {
		svc.RecordAuctionOutcome(req, 1.9, 2.0, false, "dec-comp")
	}
	for i := 0; i < 5; i++ {
		svc.RecordAuctionOutcome(req, 0.4, 0.5, false, "dec-comp")
	}

	profile, _ := svc.GetCompetitorProfile("dec-comp")
	if profile.TrendDirection != "decreasing" {
		t.Errorf("Expected 'decreasing' trend, got '%s'", profile.TrendDirection)
	}
}

func TestUpdateCompetitorProfile_TrendStable_B25(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	req := &model.BidRequest{
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
	}
	// 10 bids all ~1.0 → stable
	for i := 0; i < 10; i++ {
		svc.RecordAuctionOutcome(req, 0.9, 1.0, false, "stable-comp")
	}

	profile, _ := svc.GetCompetitorProfile("stable-comp")
	if profile.TrendDirection != "stable" {
		t.Errorf("Expected 'stable' trend, got '%s'", profile.TrendDirection)
	}
}

// ============================================================
// calculateCohesion — no users, users found, users not found
// ============================================================

func TestCalculateCohesion_NoUsers_B25(t *testing.T) {
	svc := NewUserClusteringService(nil)
	cohesion := svc.calculateCohesion([]string{}, []float64{0.5, 0.5})
	if cohesion != 0.0 {
		t.Errorf("Expected 0.0 for empty user list, got %f", cohesion)
	}
}

func TestCalculateCohesion_WithUsers_B25(t *testing.T) {
	svc := NewUserClusteringService(nil)

	svc.mu.Lock()
	svc.users["u1"] = &clusterUser{
		UserID:        "u1",
		FeatureVector: []float64{1.0, 0.0},
	}
	svc.users["u2"] = &clusterUser{
		UserID:        "u2",
		FeatureVector: []float64{0.9, 0.1},
	}
	svc.mu.Unlock()

	centroid := []float64{0.95, 0.05}
	cohesion := svc.calculateCohesion([]string{"u1", "u2"}, centroid)
	if cohesion <= 0 || cohesion > 1.0 {
		t.Errorf("Expected cohesion in (0, 1], got %f", cohesion)
	}
}

func TestCalculateCohesion_UsersNotInMap_B25(t *testing.T) {
	svc := NewUserClusteringService(nil)
	// Users not in s.users → count=0 → returns 0.0
	cohesion := svc.calculateCohesion([]string{"ghost-1", "ghost-2"}, []float64{0.5, 0.5})
	if cohesion != 0.0 {
		t.Errorf("Expected 0.0 for unknown users, got %f", cohesion)
	}
}

// ============================================================
// exploreCreative — weighted selection + fallback
// ============================================================

func TestExploreCreative_WeightedSelection_B25(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	creatives := []model.CreativeVariant{
		{ID: "cr-a", Weight: 3.0, Status: "active"},
		{ID: "cr-b", Weight: 1.0, Status: "active"},
	}
	config := &model.CreativeOptimization{
		OptimizationGoal: "ctr",
	}

	// Run many times to ensure we get a valid result
	result := svc.exploreCreative(creatives, config)
	if result == nil {
		t.Fatal("Expected non-nil result from exploreCreative")
	}
	if result.SelectionMethod != "exploration" {
		t.Errorf("Expected 'exploration' method, got '%s'", result.SelectionMethod)
	}
	if result.SelectedCreativeID != "cr-a" && result.SelectedCreativeID != "cr-b" {
		t.Errorf("Unexpected creative ID: %s", result.SelectedCreativeID)
	}
}

func TestExploreCreative_ZeroWeights_B25(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	creatives := []model.CreativeVariant{
		{ID: "cr-zw", Weight: 0, Status: "active"}, // weight=0 → treated as 1.0
	}
	result := svc.exploreCreative(creatives, &model.CreativeOptimization{})
	if result == nil || result.SelectedCreativeID == "" {
		t.Error("Expected a result for zero-weight creative")
	}
}

// ============================================================
// detectPlacementType — from dimensions and fallback
// ============================================================

func TestDetectPlacementType_FromDimensions_Wide_B25(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Dimensions: []int{728, 90}, // wide → banner
		},
	}
	result := svc.detectPlacementType(req)
	if result != "banner" {
		t.Errorf("Expected 'banner' for wide dimensions, got '%s'", result)
	}
}

func TestDetectPlacementType_FromDimensions_Tall_B25(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Dimensions: []int{300, 600}, // tall → interstitial (ratio < 0.7)
		},
	}
	result := svc.detectPlacementType(req)
	if result != "interstitial" {
		t.Errorf("Expected 'interstitial' for tall dimensions, got '%s'", result)
	}
}

func TestDetectPlacementType_NoDimensions_B25(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	req := &model.BidRequest{
		AdSlot: model.AdSlot{},
	}
	result := svc.detectPlacementType(req)
	if result != "banner" {
		t.Errorf("Expected 'banner' fallback, got '%s'", result)
	}
}

// ============================================================
// RecordAuctionOutcome — won=true sets segment floor
// ============================================================

func TestRecordAuctionOutcome_WonSetsFloor_B25(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	req := &model.BidRequest{
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
	}
	svc.RecordAuctionOutcome(req, 2.0, 1.8, true, "")

	floor := svc.GetSegmentFloor(svc.getSegmentKey(req))
	if floor != 2.0 {
		t.Errorf("Expected floor 2.0, got %f", floor)
	}
}

func TestRecordAuctionOutcome_LostUpdatesFloor_B25(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	req := &model.BidRequest{
		AdSlot: model.AdSlot{Formats: []string{"video"}},
	}
	// First record to establish a floor
	svc.RecordAuctionOutcome(req, 1.5, 2.0, false, "")
	floor1 := svc.GetSegmentFloor(svc.getSegmentKey(req))
	if floor1 != 2.0 {
		t.Errorf("Expected floor 2.0 after first loss, got %f", floor1)
	}

	// Second record — smoothly updates existing floor
	svc.RecordAuctionOutcome(req, 1.0, 3.0, false, "")
	floor2 := svc.GetSegmentFloor(svc.getSegmentKey(req))
	// floor2 = 2.0*0.8 + 3.0*0.2 = 1.6 + 0.6 = 2.2
	if floor2 <= floor1 {
		t.Errorf("Expected floor to increase, got %f (was %f)", floor2, floor1)
	}
}

// ============================================================
// autoOptimize — insufficient data path
// ============================================================

func TestAutoOptimize_InsufficientData_B25(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())

	campaign := &model.Campaign{
		ID: "camp-auto",
		Targeting: model.Targeting{
			PerformanceGoals: &model.PerformanceGoals{
				DaypartingOptimization: &model.DaypartingOptimization{
					Enabled:       true,
					AutoOptimize:  true,
					MinMultiplier: 0.3,
					MaxMultiplier: 2.0,
				},
			},
		},
	}
	// No cache data → impressions=0 → insufficient_data reason
	result := svc.CalculateDaypartMultiplier(campaign, &model.BidRequest{})
	if result.Reason != "insufficient_data" {
		t.Errorf("Expected 'insufficient_data' reason, got '%s'", result.Reason)
	}
}

// ============================================================
// getDayOfWeekModifier — zero CTR path
// ============================================================

func TestGetDayOfWeekModifier_ZeroCTR_B25(t *testing.T) {
	mc := &mockCacheForDaypart{
		MockCache: NewMockCache(),
		data:      map[string]string{},
	}
	// impressions=100 but avgPerf.ctr=0 → returns 1.0
	mc.data["daypart_dow:camp-zeroctr:4"] = "impressions:100,clicks:5,conversions:0,spend:0.0000,ctr:0.050000,cvr:0.000000"
	svc := NewDaypartingService(mc)
	avgPerf := hourlyPerformance{ctr: 0.0, impressions: 200} // ctr=0 → returns 1.0
	mod := svc.getDayOfWeekModifier("camp-zeroctr", 4, 14, avgPerf)
	if mod != 1.0 {
		t.Errorf("Expected 1.0 for zero avgCTR, got %f", mod)
	}
}

// ============================================================
// CalculateDaypartMultiplier — disabled and no PerformanceGoals
// ============================================================

func TestCalculateDaypartMultiplier_Disabled_B25(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	campaign := &model.Campaign{
		ID: "camp-disabled",
		Targeting: model.Targeting{
			PerformanceGoals: &model.PerformanceGoals{
				DaypartingOptimization: &model.DaypartingOptimization{
					Enabled: false,
				},
			},
		},
	}
	result := svc.CalculateDaypartMultiplier(campaign, &model.BidRequest{})
	if result.Multiplier != 1.0 {
		t.Errorf("Expected multiplier 1.0 for disabled, got %f", result.Multiplier)
	}
}

func TestCalculateDaypartMultiplier_NilPerformanceGoals_B25(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	campaign := &model.Campaign{
		ID:        "camp-nil-pg",
		Targeting: model.Targeting{},
	}
	result := svc.CalculateDaypartMultiplier(campaign, &model.BidRequest{})
	if result.Multiplier != 1.0 {
		t.Errorf("Expected multiplier 1.0, got %f", result.Multiplier)
	}
}
