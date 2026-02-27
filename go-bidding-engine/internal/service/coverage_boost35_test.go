package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// BOOST 35: Target 3 functions with 88-90% coverage
// 1. GetMetrics (88.9%)
// 2. GetUserDeviceGraph (88.9%)
// 3. getOrderedProviders (89.5%)

func newBiddingSvc_B35() *BiddingService {
	mc := NewMockCache()
	return NewBiddingService(mc, "")
}

// GetMetrics Tests
// Note: MockCache GetBidCount/GetWinCount/GetAverageLatency return 0, nil by default
// GetMetrics will see 0 bids, 0 wins, 0.0 latency → winRate = 0.0
// We test the simulated error path and the zero division protection
func TestB35_Metrics_ZeroBids(t *testing.T) {
	svc := newBiddingSvc_B35()

	metrics, err := svc.GetMetrics()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	winRate := metrics["win_rate"].(float64)
	if winRate != 0.0 {
		t.Errorf("Expected 0%% win rate when no bids, got %f", winRate)
	}
	if metrics["total_bids"].(int64) != 0 {
		t.Errorf("Expected 0 bids, got %v", metrics["total_bids"])
	}
}

func TestB35_Metrics_SimulatedError(t *testing.T) {
	svc := newBiddingSvc_B35()
	mc := svc.cache.(*MockCache)
	mc.kv["SIMULATE_METRICS_ERROR"] = "1"

	_, err := svc.GetMetrics()
	if err == nil {
		t.Errorf("Expected error when SIMULATE_METRICS_ERROR is set")
	}
	if err.Error() != "simulated metrics error" {
		t.Errorf("Expected 'simulated metrics error', got %v", err)
	}
}

// GetUserDeviceGraph Tests
func TestB35_DeviceGraph_Success(t *testing.T) {
	svc := newBiddingSvc_B35()
	mc := svc.cache.(*MockCache)
	mc.primaryUserID["device123"] = "user-primary-1"
	mc.linkedDevices["user-primary-1"] = []string{"device123", "device456", "device789"}

	result, err := svc.GetUserDeviceGraph("device123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.PrimaryUserID != "user-primary-1" {
		t.Errorf("Expected primary ID 'user-primary-1', got '%s'", result.PrimaryUserID)
	}
	if result.DeviceCount != 3 {
		t.Errorf("Expected 3 devices, got %d", result.DeviceCount)
	}
}

func TestB35_DeviceGraph_EmptyUserID(t *testing.T) {
	svc := newBiddingSvc_B35()

	_, err := svc.GetUserDeviceGraph("")
	if err == nil || err.Error() != "userID required" {
		t.Errorf("Expected 'userID required' error")
	}
}

func TestB35_DeviceGraph_NoPrimaryID(t *testing.T) {
	svc := newBiddingSvc_B35()
	mc := svc.cache.(*MockCache)
	mc.linkedDevices["user123"] = []string{"device1", "device2"}

	result, err := svc.GetUserDeviceGraph("user123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.PrimaryUserID != "user123" {
		t.Errorf("Expected primary ID 'user123', got '%s'", result.PrimaryUserID)
	}
}

func TestB35_DeviceGraph_NoLinkedDevices(t *testing.T) {
	svc := newBiddingSvc_B35()
	mc := svc.cache.(*MockCache)
	mc.primaryUserID["user123"] = "user-primary-1"

	result, err := svc.GetUserDeviceGraph("user123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.DeviceCount != 0 {
		t.Errorf("Expected 0 devices, got %d", result.DeviceCount)
	}
}

// getOrderedProviders Tests
func TestB35_OrderProviders_DefaultPriority(t *testing.T) {
	mc := NewMockCache()
	svc := NewUnifiedIDService(mc)

	config := &model.UnifiedIDConfig{
		Providers: []model.IDProvider{
			{Name: "Provider-C", Priority: 3},
			{Name: "Provider-A", Priority: 1},
			{Name: "Provider-B", Priority: 2},
		},
	}

	ordered := svc.getOrderedProviders(config)
	if len(ordered) != 3 {
		t.Fatalf("Expected 3 providers, got %d", len(ordered))
	}
	if ordered[0].Name != "Provider-A" {
		t.Errorf("Expected 'Provider-A' first, got '%s'", ordered[0].Name)
	}
	if ordered[1].Name != "Provider-B" {
		t.Errorf("Expected 'Provider-B' second, got '%s'", ordered[1].Name)
	}
	if ordered[2].Name != "Provider-C" {
		t.Errorf("Expected 'Provider-C' third, got '%s'", ordered[2].Name)
	}
}

func TestB35_OrderProviders_FallbackOrder(t *testing.T) {
	mc := NewMockCache()
	svc := NewUnifiedIDService(mc)

	config := &model.UnifiedIDConfig{
		Providers: []model.IDProvider{
			{Name: "Provider-A", Priority: 1},
			{Name: "Provider-B", Priority: 2},
			{Name: "Provider-C", Priority: 3},
		},
		FallbackOrder: []string{"Provider-C", "Provider-A", "Provider-B"},
	}

	ordered := svc.getOrderedProviders(config)
	if ordered[0].Name != "Provider-C" {
		t.Errorf("Expected 'Provider-C' first, got '%s'", ordered[0].Name)
	}
	if ordered[1].Name != "Provider-A" {
		t.Errorf("Expected 'Provider-A' second, got '%s'", ordered[1].Name)
	}
	if ordered[2].Name != "Provider-B" {
		t.Errorf("Expected 'Provider-B' third, got '%s'", ordered[2].Name)
	}
}

func TestB35_OrderProviders_CaseInsensitive(t *testing.T) {
	mc := NewMockCache()
	svc := NewUnifiedIDService(mc)

	config := &model.UnifiedIDConfig{
		Providers: []model.IDProvider{
			{Name: "Provider-A", Priority: 1},
			{Name: "Provider-B", Priority: 2},
		},
		FallbackOrder: []string{"provider-b", "PROVIDER-A"},
	}

	ordered := svc.getOrderedProviders(config)
	if ordered[0].Name != "Provider-B" {
		t.Errorf("Expected 'Provider-B' first (case-insensitive), got '%s'", ordered[0].Name)
	}
}

func TestB35_OrderProviders_PartialFallback(t *testing.T) {
	mc := NewMockCache()
	svc := NewUnifiedIDService(mc)

	config := &model.UnifiedIDConfig{
		Providers: []model.IDProvider{
			{Name: "Provider-A", Priority: 1},
			{Name: "Provider-B", Priority: 2},
			{Name: "Provider-C", Priority: 3},
		},
		FallbackOrder: []string{"Provider-B"},
	}

	ordered := svc.getOrderedProviders(config)
	if ordered[0].Name != "Provider-B" {
		t.Errorf("Expected 'Provider-B' first, got '%s'", ordered[0].Name)
	}
	// Others sorted by priority
	if ordered[1].Name != "Provider-A" {
		t.Errorf("Expected 'Provider-A' second, got '%s'", ordered[1].Name)
	}
}

func TestB35_OrderProviders_SamePriorityTiebreaker(t *testing.T) {
	mc := NewMockCache()
	svc := NewUnifiedIDService(mc)

	config := &model.UnifiedIDConfig{
		Providers: []model.IDProvider{
			{Name: "Provider-A", Priority: 1},
			{Name: "Provider-B", Priority: 1},
		},
		FallbackOrder: []string{"Provider-B", "Provider-A"},
	}

	ordered := svc.getOrderedProviders(config)
	if ordered[0].Name != "Provider-B" {
		t.Errorf("Expected 'Provider-B' first, got '%s'", ordered[0].Name)
	}
}
