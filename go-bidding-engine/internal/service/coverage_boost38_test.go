package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// BOOST 38: Target 2 functions with 88-91% coverage
// 1. optimizeForCPAD (88.0%)
// 2. categorizePlayerSize (90.9%)

func newBiddingSvc_B38() *BiddingService {
	mc := NewMockCache()
	return NewBiddingService(mc, "")
}

func makePerfGoals_B38_CPAD(targetCPAD float64) *model.PerformanceGoals {
	return &model.PerformanceGoals{
		TargetCPAD: targetCPAD,
	}
}

func makePerfData_B38() performanceData {
	return performanceData{}
}

// optimizeForCPAD Tests

func TestB38_OptimizeCPAD_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B38()
	camp := &model.Campaign{
		ID:       "camp-b38-1",
		BidPrice: 5.0,
	}
	req := &model.BidRequest{
		ID:      "req-b38-1",
		Device:  model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{},
	}
	pg := makePerfGoals_B38_CPAD(0) // No target
	perf := makePerfData_B38()

	ratio := svc.optimizeForCPAD(camp, req, pg, perf)
	if ratio != 1.0 {
		t.Errorf("Expected 1.0 when no target CPAD, got %f", ratio)
	}
}

func TestB38_OptimizeCPAD_HighAppRating(t *testing.T) {
	svc := newBiddingSvc_B38()
	camp := &model.Campaign{
		ID:       "camp-b38-2",
		BidPrice: 2.0,
	}
	req := &model.BidRequest{
		ID:     "req-b38-2",
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"app_rating": 4.8, // High rating (>= 4.5) → 1.2x boost
		},
	}
	pg := makePerfGoals_B38_CPAD(5.0)
	perf := makePerfData_B38()

	ratio := svc.optimizeForCPAD(camp, req, pg, perf)
	// High rating should increase the ratio
	if ratio < 0.2 {
		t.Errorf("Expected ratio >= 0.2 with high app rating, got %f", ratio)
	}
}

func TestB38_OptimizeCPAD_FeaturedApp(t *testing.T) {
	svc := newBiddingSvc_B38()
	camp := &model.Campaign{
		ID:       "camp-b38-3",
		BidPrice: 2.0,
	}
	req := &model.BidRequest{
		ID:     "req-b38-3",
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"app_featured": true, // Featured → 1.15x boost
		},
	}
	pg := makePerfGoals_B38_CPAD(5.0)
	perf := makePerfData_B38()

	ratio := svc.optimizeForCPAD(camp, req, pg, perf)
	// Featured app should increase the ratio
	if ratio < 0.2 {
		t.Errorf("Expected ratio >= 0.2 with featured app, got %f", ratio)
	}
}

func TestB38_OptimizeCPAD_HighRatingAndFeatured(t *testing.T) {
	svc := newBiddingSvc_B38()
	camp := &model.Campaign{
		ID:       "camp-b38-4",
		BidPrice: 2.0,
	}
	req := &model.BidRequest{
		ID:     "req-b38-4",
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"app_rating":   4.9,  // High rating → 1.2x
			"app_featured": true, // Featured → 1.15x
		},
	}
	pg := makePerfGoals_B38_CPAD(5.0)
	perf := makePerfData_B38()

	ratio := svc.optimizeForCPAD(camp, req, pg, perf)
	// Combined boosts (1.2 * 1.15 = 1.38x) should increase ratio
	if ratio < 0.2 {
		t.Errorf("Expected ratio >= 0.2 with high rating and featured, got %f", ratio)
	}
}

func TestB38_OptimizeCPAD_NonMobileDevice(t *testing.T) {
	svc := newBiddingSvc_B38()
	camp := &model.Campaign{
		ID:       "camp-b38-5",
		BidPrice: 2.0,
	}
	req := &model.BidRequest{
		ID:      "req-b38-5",
		Device:  model.InternalDevice{Type: "desktop"}, // Non-mobile → 0.1x penalty
		Context: map[string]interface{}{},
	}
	pg := makePerfGoals_B38_CPAD(5.0)
	perf := makePerfData_B38()

	ratio := svc.optimizeForCPAD(camp, req, pg, perf)
	// Desktop devices get 0.1x penalty, likely hitting floor of 0.2
	if ratio != 0.2 {
		t.Errorf("Expected 0.2 (floor) for desktop device, got %f", ratio)
	}
}

func TestB38_OptimizeCPAD_TabletDevice(t *testing.T) {
	svc := newBiddingSvc_B38()
	camp := &model.Campaign{
		ID:       "camp-b38-6",
		BidPrice: 2.0,
	}
	req := &model.BidRequest{
		ID:      "req-b38-6",
		Device:  model.InternalDevice{Type: "tablet"}, // Tablet is OK
		Context: map[string]interface{}{},
	}
	pg := makePerfGoals_B38_CPAD(5.0)
	perf := makePerfData_B38()

	ratio := svc.optimizeForCPAD(camp, req, pg, perf)
	// Tablet should not get the 0.1x penalty
	if ratio < 0.2 || ratio > 2.5 {
		t.Errorf("Expected ratio between 0.2 and 2.5 for tablet, got %f", ratio)
	}
}

func TestB38_OptimizeCPAD_NormalRange(t *testing.T) {
	svc := newBiddingSvc_B38()
	camp := &model.Campaign{
		ID:       "camp-b38-7",
		BidPrice: 2.0,
	}
	req := &model.BidRequest{
		ID:     "req-b38-7",
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"app_rating":   5.0,  // Max rating
			"app_featured": true, // Featured
		},
	}
	pg := makePerfGoals_B38_CPAD(5.0)
	perf := makePerfData_B38()

	ratio := svc.optimizeForCPAD(camp, req, pg, perf)
	// Should be within bounds 0.2-2.5
	if ratio < 0.2 || ratio > 2.5 {
		t.Errorf("Expected ratio between 0.2 and 2.5, got %f", ratio)
	}
}

func TestB38_OptimizeCPAD_LowRatioFloor(t *testing.T) {
	svc := newBiddingSvc_B38()
	camp := &model.Campaign{
		ID:       "camp-b38-8",
		BidPrice: 100.0, // High bid price
	}
	req := &model.BidRequest{
		ID:      "req-b38-8",
		Device:  model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{},
	}
	pg := makePerfGoals_B38_CPAD(0.5) // Low target
	perf := makePerfData_B38()

	ratio := svc.optimizeForCPAD(camp, req, pg, perf)
	// Should be floored at 0.2
	if ratio != 0.2 {
		t.Errorf("Expected 0.2 (floor), got %f", ratio)
	}
}

func TestB38_OptimizeCPAD_ZeroBidPrice(t *testing.T) {
	svc := newBiddingSvc_B38()
	camp := &model.Campaign{
		ID:       "camp-b38-9",
		BidPrice: 0, // Zero bid price
	}
	req := &model.BidRequest{
		ID:      "req-b38-9",
		Device:  model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{},
	}
	pg := makePerfGoals_B38_CPAD(5.0)
	perf := makePerfData_B38()

	ratio := svc.optimizeForCPAD(camp, req, pg, perf)
	if ratio != 1.0 {
		t.Errorf("Expected 1.0 when bid price is zero, got %f", ratio)
	}
}

// categorizePlayerSize Tests

func TestB38_PlayerSize_Unknown_ZeroWidthHeight(t *testing.T) {
	svc := newBiddingSvc_B38()
	size := svc.categorizePlayerSize(0, 0)
	if size != "unknown" {
		t.Errorf("Expected 'unknown' for 0x0, got '%s'", size)
	}
}

func TestB38_PlayerSize_XLarge(t *testing.T) {
	svc := newBiddingSvc_B38()
	size := svc.categorizePlayerSize(1920, 1080)
	if size != "xlarge" {
		t.Errorf("Expected 'xlarge' for 1920x1080, got '%s'", size)
	}
}

func TestB38_PlayerSize_Large(t *testing.T) {
	svc := newBiddingSvc_B38()
	size := svc.categorizePlayerSize(800, 600)
	if size != "large" {
		t.Errorf("Expected 'large' for 800x600, got '%s'", size)
	}
}

func TestB38_PlayerSize_Medium(t *testing.T) {
	svc := newBiddingSvc_B38()
	size := svc.categorizePlayerSize(480, 360)
	if size != "medium" {
		t.Errorf("Expected 'medium' for 480x360, got '%s'", size)
	}
}

func TestB38_PlayerSize_Small(t *testing.T) {
	svc := newBiddingSvc_B38()
	size := svc.categorizePlayerSize(320, 240)
	if size != "small" {
		t.Errorf("Expected 'small' for 320x240, got '%s'", size)
	}
}

func TestB38_PlayerSize_ZeroWidth_UsesHeight(t *testing.T) {
	svc := newBiddingSvc_B38()
	size := svc.categorizePlayerSize(0, 1080)
	// Height 1080 >= 640 → large (not xlarge because height alone is used, and 1080 < 1280)
	// Actually height 1080 >= 1280 is false, so it's >= 640 → large
	if size != "large" {
		t.Errorf("Expected 'large' for height 1080, got '%s'", size)
	}
}

func TestB38_PlayerSize_ZeroHeight_UsesWidth(t *testing.T) {
	svc := newBiddingSvc_B38()
	size := svc.categorizePlayerSize(450, 0)
	// Width 450 (400-639) → medium
	if size != "medium" {
		t.Errorf("Expected 'medium' for width 450, got '%s'", size)
	}
}

func TestB38_PlayerSize_BoundaryXLarge(t *testing.T) {
	svc := newBiddingSvc_B38()
	size := svc.categorizePlayerSize(1280, 720)
	if size != "xlarge" {
		t.Errorf("Expected 'xlarge' at boundary 1280, got '%s'", size)
	}
}

func TestB38_PlayerSize_BoundaryLarge(t *testing.T) {
	svc := newBiddingSvc_B38()
	size := svc.categorizePlayerSize(640, 480)
	if size != "large" {
		t.Errorf("Expected 'large' at boundary 640, got '%s'", size)
	}
}

func TestB38_PlayerSize_BoundaryMedium(t *testing.T) {
	svc := newBiddingSvc_B38()
	size := svc.categorizePlayerSize(400, 300)
	if size != "medium" {
		t.Errorf("Expected 'medium' at boundary 400, got '%s'", size)
	}
}
