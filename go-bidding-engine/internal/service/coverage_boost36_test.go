package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// BOOST 36: Target 2 functions with 87% coverage
// 1. optimizeForCPA (87.5%)
// 2. optimizeForCPR (87.5%)

func newBiddingSvc_B36() *BiddingService {
	mc := NewMockCache()
	return NewBiddingService(mc, "")
}

func makePerfGoals_B36_CPA(targetCPA float64) *model.PerformanceGoals {
	return &model.PerformanceGoals{
		TargetCPA: targetCPA,
	}
}

func makePerfGoals_B36_CPR(targetCPR float64) *model.PerformanceGoals {
	return &model.PerformanceGoals{
		TargetCPR: targetCPR,
	}
}

func makePerfData_B36() performanceData {
	return performanceData{}
}

// optimizeForCPA Tests
func TestB36_OptimizeCPA_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B36()
	camp := &model.Campaign{
		ID:       "camp-b36-cpa-1",
		BidPrice: 5.0,
	}
	req := &model.BidRequest{
		ID:     "req-b36-cpa-1",
		Device: model.InternalDevice{Type: "mobile"},
	}
	pg := makePerfGoals_B36_CPA(0) // No target
	perf := makePerfData_B36()

	ratio := svc.optimizeForCPA(camp, req, pg, perf)
	if ratio != 1.0 {
		t.Errorf("Expected 1.0 ratio when no target CPA, got %f", ratio)
	}
}

func TestB36_OptimizeCPA_LowRatioFloor(t *testing.T) {
	svc := newBiddingSvc_B36()
	camp := &model.Campaign{
		ID:       "camp-b36-cpa-2",
		BidPrice: 100.0, // High bid price
	}
	req := &model.BidRequest{
		ID:     "req-b36-cpa-2",
		Device: model.InternalDevice{Type: "mobile"},
	}
	pg := makePerfGoals_B36_CPA(1.0) // Low target CPA
	perf := makePerfData_B36()

	ratio := svc.optimizeForCPA(camp, req, pg, perf)
	// Should be floored at 0.3
	if ratio != 0.3 {
		t.Errorf("Expected 0.3 (floor), got %f", ratio)
	}
}

func TestB36_OptimizeCPA_ZeroBidPrice(t *testing.T) {
	svc := newBiddingSvc_B36()
	camp := &model.Campaign{
		ID:       "camp-b36-cpa-3",
		BidPrice: 0, // Zero bid price
	}
	req := &model.BidRequest{
		ID:     "req-b36-cpa-3",
		Device: model.InternalDevice{Type: "mobile"},
	}
	pg := makePerfGoals_B36_CPA(10.0)
	perf := makePerfData_B36()

	ratio := svc.optimizeForCPA(camp, req, pg, perf)
	if ratio != 1.0 {
		t.Errorf("Expected 1.0 when bid price is zero, got %f", ratio)
	}
}

func TestB36_OptimizeCPA_NormalRange(t *testing.T) {
	svc := newBiddingSvc_B36()
	camp := &model.Campaign{
		ID:       "camp-b36-cpa-4",
		BidPrice: 5.0,
	}
	req := &model.BidRequest{
		ID:     "req-b36-cpa-4",
		Device: model.InternalDevice{Type: "mobile"},
	}
	pg := makePerfGoals_B36_CPA(10.0)
	perf := makePerfData_B36()

	ratio := svc.optimizeForCPA(camp, req, pg, perf)
	// Should be between 0.3 and 2.0
	if ratio < 0.3 || ratio > 2.0 {
		t.Errorf("Expected ratio between 0.3 and 2.0, got %f", ratio)
	}
}

// optimizeForCPR Tests
func TestB36_OptimizeCPR_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B36()
	camp := &model.Campaign{
		ID:       "camp-b36-cpr-1",
		BidPrice: 5.0,
	}
	req := &model.BidRequest{
		ID:     "req-b36-cpr-1",
		Device: model.InternalDevice{Type: "mobile"},
	}
	pg := makePerfGoals_B36_CPR(0) // No target
	perf := makePerfData_B36()

	ratio := svc.optimizeForCPR(camp, req, pg, perf)
	if ratio != 1.0 {
		t.Errorf("Expected 1.0 ratio when no target CPR, got %f", ratio)
	}
}

func TestB36_OptimizeCPR_LowRatioFloor(t *testing.T) {
	svc := newBiddingSvc_B36()
	camp := &model.Campaign{
		ID:       "camp-b36-cpr-2",
		BidPrice: 100.0, // High bid price
	}
	req := &model.BidRequest{
		ID:     "req-b36-cpr-2",
		Device: model.InternalDevice{Type: "mobile"},
	}
	pg := makePerfGoals_B36_CPR(1.0) // Low target CPR
	perf := makePerfData_B36()

	ratio := svc.optimizeForCPR(camp, req, pg, perf)
	// Should be floored at 0.3
	if ratio != 0.3 {
		t.Errorf("Expected 0.3 (floor), got %f", ratio)
	}
}

func TestB36_OptimizeCPR_ZeroBidPrice(t *testing.T) {
	svc := newBiddingSvc_B36()
	camp := &model.Campaign{
		ID:       "camp-b36-cpr-3",
		BidPrice: 0, // Zero bid price
	}
	req := &model.BidRequest{
		ID:     "req-b36-cpr-3",
		Device: model.InternalDevice{Type: "mobile"},
	}
	pg := makePerfGoals_B36_CPR(10.0)
	perf := makePerfData_B36()

	ratio := svc.optimizeForCPR(camp, req, pg, perf)
	if ratio != 1.0 {
		t.Errorf("Expected 1.0 when bid price is zero, got %f", ratio)
	}
}

func TestB36_OptimizeCPR_NormalRange(t *testing.T) {
	svc := newBiddingSvc_B36()
	camp := &model.Campaign{
		ID:       "camp-b36-cpr-4",
		BidPrice: 5.0,
	}
	req := &model.BidRequest{
		ID:     "req-b36-cpr-4",
		Device: model.InternalDevice{Type: "mobile"},
	}
	pg := makePerfGoals_B36_CPR(10.0)
	perf := makePerfData_B36()

	ratio := svc.optimizeForCPR(camp, req, pg, perf)
	// Should be between 0.3 and 2.0
	if ratio < 0.3 || ratio > 2.0 {
		t.Errorf("Expected ratio between 0.3 and 2.0, got %f", ratio)
	}
}
