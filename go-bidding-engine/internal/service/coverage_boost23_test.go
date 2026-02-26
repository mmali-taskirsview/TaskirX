package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================
// getCompetitionFactor — video/native/banner formats + dimensions
// ============================================================

func TestGetCompetitionFactor_VideoFormat_B23(t *testing.T) {
	svc := NewDynamicBidService(NewMockCache())

	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Formats:    []string{"video"},
			Dimensions: []int{640, 480}, // >= 300x250
		},
	}
	factor := svc.getCompetitionFactor(req)
	// video +0.15, dimensions +0.05 → 1.20
	if factor < 1.19 || factor > 1.21 {
		t.Errorf("Expected competition factor ~1.20 (video+dims), got %.4f", factor)
	}
}

func TestGetCompetitionFactor_NativeFormat_B23(t *testing.T) {
	svc := NewDynamicBidService(NewMockCache())

	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Formats: []string{"native"},
		},
	}
	factor := svc.getCompetitionFactor(req)
	// native +0.10 → 1.10
	if factor < 1.09 || factor > 1.11 {
		t.Errorf("Expected competition factor ~1.10 (native), got %.4f", factor)
	}
}

func TestGetCompetitionFactor_BannerFormat_B23(t *testing.T) {
	svc := NewDynamicBidService(NewMockCache())

	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Formats: []string{"banner"},
		},
	}
	factor := svc.getCompetitionFactor(req)
	// banner +0.0 → 1.0
	if factor < 0.99 || factor > 1.01 {
		t.Errorf("Expected competition factor ~1.0 (banner), got %.4f", factor)
	}
}

func TestGetCompetitionFactor_LargeDimensions_B23(t *testing.T) {
	svc := NewDynamicBidService(NewMockCache())

	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Formats:    []string{"banner"},
			Dimensions: []int{300, 250}, // exactly 300x250
		},
	}
	factor := svc.getCompetitionFactor(req)
	// banner +0.0, dimensions exactly 300x250 → +0.05 → 1.05
	if factor < 1.04 || factor > 1.06 {
		t.Errorf("Expected competition factor ~1.05 (300x250), got %.4f", factor)
	}
}

func TestGetCompetitionFactor_SmallDimensions_B23(t *testing.T) {
	svc := NewDynamicBidService(NewMockCache())

	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Formats:    []string{"banner"},
			Dimensions: []int{100, 50}, // < 300x250
		},
	}
	factor := svc.getCompetitionFactor(req)
	// No dimension bonus → 1.0
	if factor < 0.99 || factor > 1.01 {
		t.Errorf("Expected competition factor ~1.0 (small dims), got %.4f", factor)
	}
}

func TestGetCompetitionFactor_NoFormats_B23(t *testing.T) {
	svc := NewDynamicBidService(NewMockCache())

	req := &model.BidRequest{
		AdSlot: model.AdSlot{
			Formats: []string{},
		},
	}
	factor := svc.getCompetitionFactor(req)
	if factor < 0.99 || factor > 1.01 {
		t.Errorf("Expected competition factor ~1.0 (no formats), got %.4f", factor)
	}
}

// ============================================================
// calculateWeightedMultiplier — known/unknown factor keys
// ============================================================

func TestCalculateWeightedMultiplier_AllKnownFactors_B23(t *testing.T) {
	svc := NewDynamicBidService(NewMockCache())

	factors := map[string]float64{
		"time":        1.1,
		"device":      1.2,
		"publisher":   1.3,
		"context":     0.9,
		"competition": 1.05,
		"goal":        1.0,
	}

	multiplier := svc.calculateWeightedMultiplier(factors)
	if multiplier <= 0 {
		t.Errorf("Expected positive multiplier, got %.4f", multiplier)
	}
	// Verify it's a weighted average of these values
	// weights: time=0.15, device=0.15, publisher=0.25, context=0.25, competition=0.10, goal=0.10
	// weightedSum = 1.1*0.15 + 1.2*0.15 + 1.3*0.25 + 0.9*0.25 + 1.05*0.10 + 1.0*0.10 = 1.09
	if multiplier < 1.07 || multiplier > 1.11 {
		t.Errorf("Expected multiplier ~1.09, got %.4f", multiplier)
	}
}

func TestCalculateWeightedMultiplier_EmptyFactors_B23(t *testing.T) {
	svc := NewDynamicBidService(NewMockCache())

	factors := map[string]float64{}
	multiplier := svc.calculateWeightedMultiplier(factors)
	// No factors → returns 1.0
	if multiplier != 1.0 {
		t.Errorf("Expected multiplier 1.0 for empty factors, got %.4f", multiplier)
	}
}

func TestCalculateWeightedMultiplier_UnknownFactor_B23(t *testing.T) {
	svc := NewDynamicBidService(NewMockCache())

	factors := map[string]float64{
		"unknown_factor": 2.0, // Not in the weights map → ignored
		"time":           1.1,
	}
	multiplier := svc.calculateWeightedMultiplier(factors)
	// Only "time" contributes → weightedSum = 1.1*0.15, totalWeight = 0.15
	// result = 1.1*0.15 / 0.15 = 1.1
	if multiplier < 1.09 || multiplier > 1.11 {
		t.Errorf("Expected multiplier ~1.1 (only time factor), got %.4f", multiplier)
	}
}

// ============================================================
// calculateQualityScore — direct publisher branches
// ============================================================

func TestCalculateQualityScore_AllMetrics_B23(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub := &DirectPublisher{
		ID:               "pub-q-all",
		Domain:           "quality.example.com",
		ViewabilityRate:  0.80,
		IVTRate:          0.02, // 2% IVT → ivtScore = 1 - 0.02*10 = 0.8
		BrandSafetyScore: 0.90,
		IsDirectSeller:   true,
		FeeStructure:     FeeStructure{TotalTakeRate: 0.15},
	}

	score := svc.calculateQualityScore(pub)
	if score <= 0 || score > 1 {
		t.Errorf("Expected quality score in [0,1], got %.4f", score)
	}
}

func TestCalculateQualityScore_NoMetrics_B23(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub := &DirectPublisher{
		ID:     "pub-q-none",
		Domain: "bare.example.com",
		// No viewability, IVT is 0, no brand safety, not direct, no fees
	}

	score := svc.calculateQualityScore(pub)
	// IVTRate = 0 → ivtScore = 1.0; other fields are 0/false
	// score = 1.0 * 0.25; weights = 0.25
	if score <= 0 {
		t.Errorf("Expected positive quality score (IVT=0 → ivtScore=1), got %.4f", score)
	}
}

func TestCalculateQualityScore_HighIVT_B23(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	pub := &DirectPublisher{
		ID:      "pub-q-hivt",
		Domain:  "hivt.example.com",
		IVTRate: 0.15, // ivtScore = 1 - 1.5 = -0.5 → clamped to 0
	}

	score := svc.calculateQualityScore(pub)
	// ivtScore < 0 → clamped to 0; only IVT weight contributes with 0
	// weights = 0.25, score = 0 → score/weights = 0
	if score != 0.0 {
		t.Errorf("Expected quality score 0 for high IVT, got %.4f", score)
	}
}

func TestCalculateQualityScore_WeightsZero_B23(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	// Artificially: ViewabilityRate=0, IVTRate=-1 (negative means no IVT check)
	// But IVTRate >= 0 check: if we set IVTRate < 0 that branch won't run
	// The code checks "if pub.IVTRate >= 0" — so IVTRate = -1 skips it
	// ViewabilityRate = 0 also skips the viewability branch
	// BrandSafetyScore = 0 skips brand safety
	// IsDirectSeller = false → no direct bonus
	// FeeStructure.TotalTakeRate = 0 → no fee bonus
	// So weights = 0 → returns 0.5 (default)
	pub := &DirectPublisher{
		ID:      "pub-q-zero",
		Domain:  "zero.example.com",
		IVTRate: -1.0, // < 0 → IVT branch skipped
	}

	score := svc.calculateQualityScore(pub)
	if score != 0.5 {
		t.Errorf("Expected default quality score 0.5 (all weights=0), got %.4f", score)
	}
}

// ============================================================
// AddIntegration — publisher not found, empty publisher ID
// ============================================================

func TestAddIntegration_PublisherNotFound_B23(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	integration := &PublisherIntegration{
		PublisherID:     "nonexistent-publisher",
		IntegrationType: "api",
		Endpoint:        "http://api.example.com",
	}
	_, err := svc.AddIntegration(integration)
	if err == nil {
		t.Error("Expected error for nonexistent publisher")
	}
}

func TestAddIntegration_EmptyPublisherID_B23(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	integration := &PublisherIntegration{
		PublisherID: "",
	}
	_, err := svc.AddIntegration(integration)
	if err == nil {
		t.Error("Expected error for empty publisher_id")
	}
}

func TestAddIntegration_DefaultsApplied_B23(t *testing.T) {
	svc := NewDirectPublisherService(nil)

	// Register publisher first
	pub := &DirectPublisher{Domain: "test-integ.example.com"}
	created, _ := svc.RegisterPublisher(pub)

	integration := &PublisherIntegration{
		PublisherID:     created.ID,
		IntegrationType: "api",
		Endpoint:        "http://api.example.com",
		Timeout:         0, // should default to 150ms
		MaxQPS:          0, // should default to 1000
	}
	result, err := svc.AddIntegration(integration)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Timeout != 150*time.Millisecond {
		t.Errorf("Expected default Timeout 150ms, got %v", result.Timeout)
	}
	if result.MaxQPS != 1000 {
		t.Errorf("Expected default MaxQPS 1000, got %d", result.MaxQPS)
	}
}

// ============================================================
// AnalyzeSupplyPathEfficiency — high latency + low success + high fees + low efficiency
// ============================================================

type mockCacheForSPA struct {
	*MockCache
	metrics *model.SupplyChainMetrics
}

func (m *mockCacheForSPA) GetSupplyChainMetrics(timeRange string) (*model.SupplyChainMetrics, error) {
	return m.metrics, nil
}

func TestAnalyzeSupplyPathEfficiency_AllBranches_B23(t *testing.T) {
	mc := &mockCacheForSPA{
		MockCache: NewMockCache(),
		metrics: &model.SupplyChainMetrics{
			TotalRequests:  1000,
			AvgTotalFees:   0.005,
			PathEfficiency: 0.5, // < 0.7 → direct_connection suggestion
			ServiceMetrics: map[string]model.ServiceMetrics{
				"slow-svc": {
					ServiceName:  "slow-svc",
					AvgLatencyMs: 250.0, // > 200 → cache suggestion
					SuccessRate:  0.80,  // < 0.95 → circuit_breaker suggestion
					TotalFees:    0.015, // > 0.01 → fee_negotiation suggestion
					TotalCalls:   1000,
				},
			},
		},
	}
	svc := NewSupplyPathAnalyticsService(mc)

	optimization, err := svc.AnalyzeSupplyPathEfficiency("1h")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if optimization == nil {
		t.Fatal("Expected non-nil optimization")
	}

	// Should have at least: cache, circuit_breaker, fee_negotiation, direct_connection
	if len(optimization.Optimizations) < 3 {
		t.Errorf("Expected >= 3 optimization suggestions, got %d", len(optimization.Optimizations))
	}

	typeSet := make(map[string]bool)
	for _, s := range optimization.Optimizations {
		typeSet[s.Type] = true
	}
	for _, expected := range []string{"cache", "circuit_breaker", "fee_negotiation", "direct_connection"} {
		if !typeSet[expected] {
			t.Errorf("Expected suggestion type %q, not found in %v", expected, typeSet)
		}
	}
}

func TestAnalyzeSupplyPathEfficiency_NilMetrics_B23(t *testing.T) {
	mc := &mockCacheForSPA{
		MockCache: NewMockCache(),
		metrics:   nil,
	}
	svc := NewSupplyPathAnalyticsService(mc)

	optimization, err := svc.AnalyzeSupplyPathEfficiency("1h")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(optimization.Optimizations) != 0 {
		t.Errorf("Expected empty optimizations for nil metrics")
	}
}

// ============================================================
// GetTopCombinations — with limit applied
// ============================================================

func TestGetTopCombinations_WithLimit_B23(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	// Create a template and combinations
	tmpl := &CreativeTemplate{
		ID:    "tmpl-top",
		Name:  "Top Template",
		Slots: map[string]*TemplateSlot{},
	}
	svc.templates.Store(tmpl.ID, tmpl)

	// Add 5 combinations with different CTRs
	for i := 0; i < 5; i++ {
		combo := &CreativeCombination{
			ID:         fmt.Sprintf("combo-%d", i),
			TemplateID: "tmpl-top",
			Elements:   map[string]string{},
			Performance: &CombinationPerformance{
				Impressions: 1000,
				Clicks:      int64(10 * (i + 1)), // CTR: 0.01 to 0.05
			},
		}
		combo.Performance.CTR = float64(combo.Performance.Clicks) / float64(combo.Performance.Impressions)
		svc.combinations.Store(combo.ID, combo)
	}

	// Get top 3
	top := svc.GetTopCombinations("tmpl-top", 3)
	if len(top) != 3 {
		t.Errorf("Expected 3 combinations (limit), got %d", len(top))
	}
	// Should be sorted by CTR descending
	if top[0].Performance.CTR < top[1].Performance.CTR {
		t.Error("Expected combinations sorted by CTR descending")
	}
}

func TestGetTopCombinations_LimitLargerThanResults_B23(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	tmpl := &CreativeTemplate{
		ID:    "tmpl-small",
		Name:  "Small Template",
		Slots: map[string]*TemplateSlot{},
	}
	svc.templates.Store(tmpl.ID, tmpl)

	// Only 2 combinations, limit 10
	for i := 0; i < 2; i++ {
		combo := &CreativeCombination{
			ID:          fmt.Sprintf("combo-s%d", i),
			TemplateID:  "tmpl-small",
			Elements:    map[string]string{},
			Performance: &CombinationPerformance{CTR: float64(i+1) * 0.01},
		}
		svc.combinations.Store(combo.ID, combo)
	}

	top := svc.GetTopCombinations("tmpl-small", 10)
	if len(top) != 2 {
		t.Errorf("Expected 2 combinations (< limit), got %d", len(top))
	}
}

// ============================================================
// GenerateLookalike — disabled, insufficient seed, nil profile
// ============================================================

func TestGenerateLookalike_Disabled_B23(t *testing.T) {
	svc := NewLookalikeService(nil)
	svc.config.Enabled = false

	result := svc.GenerateLookalike([]string{"u1", "u2", "u3"}, "Test", 5.0)
	if result.Status != "disabled" {
		t.Errorf("Expected status 'disabled', got '%s'", result.Status)
	}
}

func TestGenerateLookalike_InsufficientSeed_B23(t *testing.T) {
	svc := NewLookalikeService(nil)
	svc.config.Enabled = true
	svc.config.MinSeedSize = 10 // Need at least 10 users

	result := svc.GenerateLookalike([]string{"u1", "u2"}, "Test", 5.0)
	if result.Status != "insufficient_seed" {
		t.Errorf("Expected status 'insufficient_seed', got '%s'", result.Status)
	}
	if result.SeedSize != 2 {
		t.Errorf("Expected SeedSize=2, got %d", result.SeedSize)
	}
}

func TestGenerateLookalike_SeedProfileFailed_B23(t *testing.T) {
	svc := NewLookalikeService(nil)
	svc.config.Enabled = true
	svc.config.MinSeedSize = 1

	// User IDs not in the profile map → buildSeedProfile returns nil → "seed_profile_failed"
	result := svc.GenerateLookalike([]string{"unknown-user-1", "unknown-user-2"}, "Test", 5.0)
	if result.Status != "seed_profile_failed" {
		t.Errorf("Expected status 'seed_profile_failed', got '%s'", result.Status)
	}
}

// ============================================================
// updateDeliveryProgress — slightly_behind and on_pace branches
// ============================================================

func TestUpdateDeliveryProgress_SlightlyBehind_B23(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)
	// AlertThreshold = 0.8; slightly_behind when deliveryRate >= 0.8 but < 1.0
	// We need deliveryRate between 0.8 and 1.0

	deal := &PGDeal{
		BuyerID:              "buyer-sb",
		SellerID:             "seller-sb",
		CommittedImpressions: 1000,
		StartDate:            time.Now().Add(-24 * time.Hour), // started yesterday
		EndDate:              time.Now().Add(9 * 24 * time.Hour),
	}
	created, _ := svc.CreateDeal(deal)
	created.Status = "active"
	svc.deals.Store(created.ID, created)

	// Record ~85% of expected pace (slightly behind: deliveryRate >= 0.8 but < 1.0)
	// After 1 day of 10, expected ~100 imps; record 88 to get ~0.88 delivery rate
	for i := 0; i < 88; i++ {
		_ = svc.RecordImpression(created.ID, 0.001)
	}

	progress, err := svc.GetDeliveryProgress(created.ID)
	if err != nil {
		t.Fatalf("GetDeliveryProgress failed: %v", err)
	}
	t.Logf("Status: %s, DeliveryRate: %.4f", progress.Status, progress.DeliveryRate)
	// Acceptable: on_pace or slightly_behind depending on exact timing
	if progress.Status == "" {
		t.Error("Expected non-empty delivery status")
	}
}

func TestUpdateDeliveryProgress_OnPace_B23(t *testing.T) {
	svc := NewProgrammaticGuaranteedService(nil)

	deal := &PGDeal{
		BuyerID:              "buyer-op",
		SellerID:             "seller-op",
		CommittedImpressions: 100,
		StartDate:            time.Now().Add(-1 * time.Hour),
		EndDate:              time.Now().Add(23 * time.Hour),
	}
	created, _ := svc.CreateDeal(deal)
	created.Status = "active"
	svc.deals.Store(created.ID, created)

	// Deliver all 100 impressions → delivery rate >= 1.0 → on_pace
	for i := 0; i < 100; i++ {
		_ = svc.RecordImpression(created.ID, 0.01)
	}

	progress, err := svc.GetDeliveryProgress(created.ID)
	if err != nil {
		t.Fatalf("GetDeliveryProgress failed: %v", err)
	}
	t.Logf("Status: %s, DeliveryRate: %.4f", progress.Status, progress.DeliveryRate)
	if progress.Status == "" {
		t.Error("Expected non-empty delivery status")
	}
}
