package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================================
// SUPPLY PATH ANALYTICS SERVICE TESTS
// ============================================================================

func TestNewSupplyPathAnalyticsService(t *testing.T) {
	service := NewSupplyPathAnalyticsService(nil)
	if service == nil {
		t.Error("Expected service to be created, got nil")
	}
}

func TestSupplyPathAnalytics_GetSupplyChainMetrics_NilCache(t *testing.T) {
	service := NewSupplyPathAnalyticsService(&mockTestCache{})
	
	// mockTestCache returns nil for GetSupplyChainMetrics
	metrics, err := service.GetSupplyChainMetrics("1h")
	
	// Should handle nil gracefully
	if err != nil {
		// This is acceptable - cache might return error
	}
	// If no error, metrics should be nil
	if err == nil && metrics != nil {
		t.Error("Expected nil metrics from mock cache")
	}
}

func TestSupplyPathAnalytics_GetServicePerformance_NilCache(t *testing.T) {
	service := NewSupplyPathAnalyticsService(&mockTestCache{})
	
	metrics, err := service.GetServicePerformance("bidding", "1h")
	
	// Should handle nil gracefully
	if err != nil {
		// This is acceptable
	}
	if err == nil && metrics != nil {
		t.Error("Expected nil metrics from mock cache")
	}
}

func TestSupplyPathAnalytics_GetBidPathAnalytics_NilCache(t *testing.T) {
	service := NewSupplyPathAnalyticsService(&mockTestCache{})
	
	analytics, err := service.GetBidPathAnalytics("req-123")
	
	// Should handle nil gracefully
	if err != nil {
		// This is acceptable
	}
	if err == nil && analytics != nil {
		t.Error("Expected nil analytics from mock cache")
	}
}

func TestSupplyPathAnalytics_AnalyzeSupplyPathEfficiency_NilMetrics(t *testing.T) {
	service := NewSupplyPathAnalyticsService(&mockTestCache{})
	
	optimization, err := service.AnalyzeSupplyPathEfficiency("1h")
	
	// Should not panic and return valid result
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if optimization == nil {
		t.Error("Expected optimization result, got nil")
	}
	// Empty optimizations when no metrics
	if len(optimization.Optimizations) != 0 {
		t.Errorf("Expected 0 optimizations, got %d", len(optimization.Optimizations))
	}
}

func TestSupplyPathAnalytics_AnalyzeSupplyPathEfficiency_WithMetrics(t *testing.T) {
	cache := &mockTestCacheWithMetrics{
		supplyChainMetrics: &model.SupplyChainMetrics{
			TotalRequests:  1000,
			AvgLatencyMs:   50.0,
			AvgTotalFees:   0.005,
			PathEfficiency: 0.85,
			WinRate:        0.25,
			ServiceMetrics: map[string]model.ServiceMetrics{
				"bidding": {
					ServiceName:   "bidding",
					TotalCalls:    1000,
					SuccessRate:   0.99,
					AvgLatencyMs:  30.0,
					TotalFees:     0.001,
				},
				"ad-server": {
					ServiceName:   "ad-server",
					TotalCalls:    800,
					SuccessRate:   0.92, // Low - should trigger circuit breaker suggestion
					AvgLatencyMs:  250.0, // High - should trigger cache suggestion
					TotalFees:     0.02, // High - should trigger fee negotiation
				},
			},
		},
	}
	service := NewSupplyPathAnalyticsService(cache)
	
	optimization, err := service.AnalyzeSupplyPathEfficiency("1h")
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if optimization == nil {
		t.Error("Expected optimization result, got nil")
	}
	// Should have suggestions for the ad-server service
	if len(optimization.Optimizations) == 0 {
		t.Error("Expected at least one optimization suggestion")
	}
	
	// Check for expected suggestion types
	hasCache := false
	hasCircuitBreaker := false
	hasFeeNegotiation := false
	for _, opt := range optimization.Optimizations {
		switch opt.Type {
		case "cache":
			hasCache = true
		case "circuit_breaker":
			hasCircuitBreaker = true
		case "fee_negotiation":
			hasFeeNegotiation = true
		}
	}
	
	if !hasCache {
		t.Error("Expected cache suggestion for high-latency service")
	}
	if !hasCircuitBreaker {
		t.Error("Expected circuit_breaker suggestion for low success rate service")
	}
	if !hasFeeNegotiation {
		t.Error("Expected fee_negotiation suggestion for high-fee service")
	}
}

func TestSupplyPathAnalytics_GetTopBottlenecks_NilMetrics(t *testing.T) {
	service := NewSupplyPathAnalyticsService(&mockTestCache{})
	
	bottlenecks, err := service.GetTopBottlenecks("1h", 5)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if bottlenecks == nil {
		t.Error("Expected empty list, got nil")
	}
	if len(bottlenecks) != 0 {
		t.Errorf("Expected 0 bottlenecks, got %d", len(bottlenecks))
	}
}

func TestSupplyPathAnalytics_GetTopBottlenecks_WithMetrics(t *testing.T) {
	cache := &mockTestCacheWithMetrics{
		supplyChainMetrics: &model.SupplyChainMetrics{
			ServiceMetrics: map[string]model.ServiceMetrics{
				"slow-service": {AvgLatencyMs: 300.0},
				"medium-service": {AvgLatencyMs: 150.0},
				"fast-service": {AvgLatencyMs: 20.0},
			},
		},
	}
	service := NewSupplyPathAnalyticsService(cache)
	
	// Get top 2 bottlenecks
	bottlenecks, err := service.GetTopBottlenecks("1h", 2)
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(bottlenecks) != 2 {
		t.Errorf("Expected 2 bottlenecks, got %d", len(bottlenecks))
	}
	// Should be sorted by latency (highest first)
	if bottlenecks[0].AvgLatencyMs < bottlenecks[1].AvgLatencyMs {
		t.Error("Bottlenecks should be sorted by latency (highest first)")
	}
}

func TestSupplyPathAnalytics_GetCostAnalysis_NilMetrics(t *testing.T) {
	service := NewSupplyPathAnalyticsService(&mockTestCache{})
	
	costs, err := service.GetCostAnalysis("1h")
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if costs == nil {
		t.Error("Expected empty map, got nil")
	}
	if len(costs) != 0 {
		t.Errorf("Expected 0 cost entries, got %d", len(costs))
	}
}

func TestSupplyPathAnalytics_GetCostAnalysis_WithMetrics(t *testing.T) {
	cache := &mockTestCacheWithMetrics{
		supplyChainMetrics: &model.SupplyChainMetrics{
			TotalRequests: 10000,
			AvgTotalFees:  0.005,
		},
	}
	service := NewSupplyPathAnalyticsService(cache)
	
	costs, err := service.GetCostAnalysis("1h")
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if costs == nil {
		t.Error("Expected cost map, got nil")
	}
	if _, ok := costs["total_fees"]; !ok {
		t.Error("Expected 'total_fees' in cost analysis")
	}
	if _, ok := costs["avg_fee_per_request"]; !ok {
		t.Error("Expected 'avg_fee_per_request' in cost analysis")
	}
	if _, ok := costs["fee_efficiency"]; !ok {
		t.Error("Expected 'fee_efficiency' in cost analysis")
	}
	
	// Verify calculations
	expectedTotal := 0.005 * 10000
	if costs["total_fees"] != expectedTotal {
		t.Errorf("Expected total_fees %.2f, got %.2f", expectedTotal, costs["total_fees"])
	}
}

func TestSupplyPathAnalytics_AnalyzeDirectPublisherOpportunities_NilMetrics(t *testing.T) {
	service := NewSupplyPathAnalyticsService(&mockTestCache{})
	
	analysis, err := service.AnalyzeDirectPublisherOpportunities("1h")
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if analysis == nil {
		t.Error("Expected analysis result, got nil")
	}
	if len(analysis.Opportunities) != 0 {
		t.Errorf("Expected 0 opportunities, got %d", len(analysis.Opportunities))
	}
}

func TestSupplyPathAnalytics_AnalyzeDirectPublisherOpportunities_WithMetrics(t *testing.T) {
	cache := &mockTestCacheWithMetrics{
		supplyChainMetrics: &model.SupplyChainMetrics{
			ServiceMetrics: map[string]model.ServiceMetrics{
				"low-success-ssp": {
					ServiceName:  "low-success-ssp",
					TotalCalls:   1000,
					SuccessRate:  0.90, // Below 0.95 threshold
					TotalFees:    50.0,
				},
				"high-success-ssp": {
					ServiceName:  "high-success-ssp",
					TotalCalls:   1000,
					SuccessRate:  0.99, // Above threshold - no opportunity
					TotalFees:    30.0,
				},
			},
		},
	}
	service := NewSupplyPathAnalyticsService(cache)
	
	analysis, err := service.AnalyzeDirectPublisherOpportunities("1h")
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if analysis == nil {
		t.Error("Expected analysis result, got nil")
	}
	// Should only identify low-success-ssp as opportunity
	if len(analysis.Opportunities) != 1 {
		t.Errorf("Expected 1 opportunity, got %d", len(analysis.Opportunities))
	}
	if len(analysis.Opportunities) > 0 {
		opp := analysis.Opportunities[0]
		if opp.ServiceName != "low-success-ssp" {
			t.Errorf("Expected low-success-ssp, got %s", opp.ServiceName)
		}
		if opp.Priority != "high" {
			t.Errorf("Expected high priority, got %s", opp.Priority)
		}
		if opp.ROI <= 0 {
			t.Error("Expected positive ROI for opportunity")
		}
	}
}

func TestSupplyPathAnalytics_CalculateCostBenefitAnalysis_NilMetrics(t *testing.T) {
	service := NewSupplyPathAnalyticsService(&mockTestCache{})
	
	analysis, err := service.CalculateCostBenefitAnalysis("1h")
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if analysis == nil {
		t.Error("Expected analysis result, got nil")
	}
	if len(analysis.Scenarios) != 0 {
		t.Errorf("Expected 0 scenarios, got %d", len(analysis.Scenarios))
	}
}

func TestSupplyPathAnalytics_CalculateCostBenefitAnalysis_WithMetrics(t *testing.T) {
	cache := &mockTestCacheWithMetrics{
		supplyChainMetrics: &model.SupplyChainMetrics{
			TotalRequests:  100000,
			AvgLatencyMs:   80.0,
			AvgTotalFees:   0.008,
			WinRate:        0.22,
		},
	}
	service := NewSupplyPathAnalyticsService(cache)
	
	analysis, err := service.CalculateCostBenefitAnalysis("1h")
	
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if analysis == nil {
		t.Error("Expected analysis result, got nil")
	}
	
	// Should have 3 scenarios: Direct, Optimization, Hybrid
	if len(analysis.Scenarios) != 3 {
		t.Errorf("Expected 3 scenarios, got %d", len(analysis.Scenarios))
	}
	
	// Verify scenario names
	scenarioNames := make(map[string]bool)
	for _, s := range analysis.Scenarios {
		scenarioNames[s.Name] = true
		// Each scenario should have positive net benefit
		if s.NetBenefit <= 0 {
			t.Errorf("Expected positive net benefit for scenario %s", s.Name)
		}
	}
	
	if !scenarioNames["Direct Publisher Connections"] {
		t.Error("Missing 'Direct Publisher Connections' scenario")
	}
	if !scenarioNames["Service Performance Optimization"] {
		t.Error("Missing 'Service Performance Optimization' scenario")
	}
	if !scenarioNames["Hybrid Optimization"] {
		t.Error("Missing 'Hybrid Optimization' scenario")
	}
	
	// Verify current metrics are recorded
	if analysis.CurrentWinRate != 0.22 {
		t.Errorf("Expected CurrentWinRate 0.22, got %f", analysis.CurrentWinRate)
	}
}

// ============================================================================
// MOCK CACHE FOR SUPPLY PATH ANALYTICS TESTS
// ============================================================================

// mockTestCache returns nil for all methods (tests nil-safe paths)
type mockTestCache struct{}

func (m *mockTestCache) Get(key string) (string, error)                                 { return "", nil }
func (m *mockTestCache) Set(key string, value interface{}, ttl int64) error             { return nil }
func (m *mockTestCache) GetActiveCampaigns() ([]*model.Campaign, error)                 { return nil, nil }
func (m *mockTestCache) SetActiveCampaigns(campaigns []*model.Campaign) error           { return nil }
func (m *mockTestCache) GetCampaign(campaignID string) (*model.Campaign, error)         { return nil, nil }
func (m *mockTestCache) SetCampaign(campaign *model.Campaign) error                     { return nil }
func (m *mockTestCache) IncrementBidCount() error                                       { return nil }
func (m *mockTestCache) IncrementWinCount() error                                       { return nil }
func (m *mockTestCache) GetBidCount() (int64, error)                                    { return 0, nil }
func (m *mockTestCache) GetWinCount() (int64, error)                                    { return 0, nil }
func (m *mockTestCache) RecordLatency(latencyMs float64) error                          { return nil }
func (m *mockTestCache) GetAverageLatency() (float64, error)                            { return 0, nil }
func (m *mockTestCache) SetUserSegments(userID string, segments []string) error         { return nil }
func (m *mockTestCache) GetUserSegments(userID string) ([]string, error)                { return nil, nil }
func (m *mockTestCache) SetGeoRules(countryCode string, rules map[string]interface{}) error {
	return nil
}
func (m *mockTestCache) GetGeoRules(countryCode string) (map[string]interface{}, error) { return nil, nil }
func (m *mockTestCache) IncrementCampaignSpend(campaignID string, amount float64) (float64, error) {
	return 0, nil
}
func (m *mockTestCache) GetCampaignSpend(campaignID string) (float64, error) { return 0, nil }
func (m *mockTestCache) IncrementBidFormat(format string) error              { return nil }
func (m *mockTestCache) GetBidFormats() (map[string]int64, error)            { return nil, nil }
func (m *mockTestCache) IncrementPublisherFraud(publisherID string) error    { return nil }
func (m *mockTestCache) IsRequestDuplicate(requestID string, ttlSeconds int) (bool, error) {
	return false, nil
}
func (m *mockTestCache) IncrementUserFrequency(userID, campaignID string, windowSecs int) (int64, error) {
	return 0, nil
}
func (m *mockTestCache) GetUserFrequency(userID, campaignID string) (int64, error) { return 0, nil }
func (m *mockTestCache) GetCampaignCTR(campaignID string) (float64, error)         { return 0, nil }
func (m *mockTestCache) GetCampaignWinRate(campaignID string) (float64, error)     { return 0, nil }
func (m *mockTestCache) IncrementCampaignClicks(campaignID string) error           { return nil }
func (m *mockTestCache) IncrementCampaignImpressions(campaignID string) error      { return nil }
func (m *mockTestCache) IncrementCampaignBids(campaignID string) error             { return nil }
func (m *mockTestCache) IncrementCampaignWins(campaignID string) error             { return nil }
func (m *mockTestCache) RecordBidInBucket(priceBucket string) error                { return nil }
func (m *mockTestCache) RecordWinInBucket(priceBucket string) error                { return nil }
func (m *mockTestCache) GetBidLandscape() (map[string]map[string]int64, error)     { return nil, nil }
func (m *mockTestCache) IncrementSegmentImpressions(segmentType, segmentValue string) error {
	return nil
}
func (m *mockTestCache) IncrementSegmentClicks(segmentType, segmentValue string) error { return nil }
func (m *mockTestCache) GetSegmentPerformance(segmentType string) (map[string]map[string]int64, error) {
	return nil, nil
}
func (m *mockTestCache) RecordPublisherBidAttempt(publisherID string, bidPrice float64, won bool) error {
	return nil
}
func (m *mockTestCache) GetOptimalBidFloor(publisherID string, targetWinRate float64) (float64, error) {
	return 0, nil
}
func (m *mockTestCache) RecordImpression(userID, campaignID, requestID string, ttlHours int) error {
	return nil
}
func (m *mockTestCache) RecordClick(userID, campaignID, requestID string, ttlHours int) error {
	return nil
}
func (m *mockTestCache) GetAttribution(userID, campaignID string) (string, string, error) {
	return "", "", nil
}
func (m *mockTestCache) RecordTouchpoint(userID, campaignID, touchpointType, requestID string, ttlDays int) error {
	return nil
}
func (m *mockTestCache) GetTouchpoints(userID, campaignID string) ([]model.Touchpoint, error) {
	return nil, nil
}
func (m *mockTestCache) GetMultiTouchAttribution(userID, campaignID, modelType string) ([]model.AttributionCredit, error) {
	return nil, nil
}
func (m *mockTestCache) RecordUserEvent(userID, campaignID, eventType string, ttlDays int) error {
	return nil
}
func (m *mockTestCache) GetUserEvents(userID string, eventTypes []string) (map[string][]string, error) {
	return nil, nil
}
func (m *mockTestCache) HasUserEvent(userID, campaignID, eventType string) (bool, error) {
	return false, nil
}
func (m *mockTestCache) LinkDevices(primaryUserID string, deviceIDs []string, ttlDays int) error {
	return nil
}
func (m *mockTestCache) GetLinkedDevices(deviceID string) ([]string, error) { return nil, nil }
func (m *mockTestCache) GetPrimaryUserID(deviceID string) (string, error)   { return "", nil }
func (m *mockTestCache) GetCrossDeviceFrequency(primaryUserID, campaignID string) (int64, error) {
	return 0, nil
}
func (m *mockTestCache) StoreBidPathAnalytics(analytics *model.BidPathAnalytics) error { return nil }
func (m *mockTestCache) GetBidPathAnalytics(requestID string) (*model.BidPathAnalytics, error) {
	return nil, nil
}
func (m *mockTestCache) GetSupplyChainMetrics(timeRange string) (*model.SupplyChainMetrics, error) {
	return nil, nil
}
func (m *mockTestCache) GetServiceMetrics(serviceName string, timeRange string) (*model.ServiceMetrics, error) {
	return nil, nil
}

// mockTestCacheWithMetrics returns configurable metrics data
type mockTestCacheWithMetrics struct {
	mockTestCache
	supplyChainMetrics *model.SupplyChainMetrics
	serviceMetrics     *model.ServiceMetrics
}

func (m *mockTestCacheWithMetrics) GetSupplyChainMetrics(timeRange string) (*model.SupplyChainMetrics, error) {
	return m.supplyChainMetrics, nil
}

func (m *mockTestCacheWithMetrics) GetServiceMetrics(serviceName string, timeRange string) (*model.ServiceMetrics, error) {
	return m.serviceMetrics, nil
}
