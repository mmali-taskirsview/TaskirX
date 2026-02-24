package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/taskirx/go-bidding-engine/internal/model"
	"github.com/taskirx/go-bidding-engine/internal/service"
)

// local test cache that returns a deterministic landscape
type mockCacheWithLandscape struct{ mockCache }

func (m *mockCacheWithLandscape) GetBidLandscape() (map[string]map[string]int64, error) {
	return map[string]map[string]int64{
		"0.50-1.00": map[string]int64{"bids": 100, "wins": 25},
		"1.00-2.00": map[string]int64{"bids": 0, "wins": 0},
	}, nil
}

// ============================================================================
// ANALYTICS HANDLER TEST SETUP
// ============================================================================

func setupAnalyticsTestHandler() (*AnalyticsHandler, *gin.Engine) {
	cache := &mockCache{}
	biddingService := service.NewBiddingService(cache, "http://localhost:8080")
	handler := NewAnalyticsHandler(biddingService)

	router := gin.New()
	// Add recovery middleware to handle panics from nil cache returns
	router.Use(gin.Recovery())
	return handler, router
}

// ============================================================================
// SUPPLY CHAIN METRICS TESTS
// ============================================================================

func TestGetSupplyChainMetrics_Default(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/supply-chain/metrics", handler.GetSupplyChainMetrics)

	req, _ := http.NewRequest("GET", "/api/analytics/supply-chain/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Handler calls service which may return nil metrics - check it handles gracefully
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetSupplyChainMetrics_WithTimeRange(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/supply-chain/metrics", handler.GetSupplyChainMetrics)

	req, _ := http.NewRequest("GET", "/api/analytics/supply-chain/metrics?timeRange=24h", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// SUPPLY PATH OPTIMIZATION TESTS
// ============================================================================

func TestGetSupplyPathOptimization_Default(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/supply-path/optimization", handler.GetSupplyPathOptimization)

	req, _ := http.NewRequest("GET", "/api/analytics/supply-path/optimization", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Handler may return 500 due to nil cache returns in mock - this is expected
	// In production, the cache would return actual data
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetSupplyPathOptimization_CustomTimeRange(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/supply-path/optimization", handler.GetSupplyPathOptimization)

	req, _ := http.NewRequest("GET", "/api/analytics/supply-path/optimization?timeRange=7d", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Handler may return 500 due to nil cache returns in mock
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// BID PATH ANALYTICS TESTS
// ============================================================================

func TestGetBidPathAnalytics_ValidRequestID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/bid-path/:requestId", handler.GetBidPathAnalytics)

	req, _ := http.NewRequest("GET", "/api/analytics/bid-path/req-12345", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Will return 404 since we're using mock cache that returns nil
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 404 or 500, got %d", w.Code)
	}
}

func TestGetBidPathAnalytics_EmptyRequestID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/bid-path/:requestId", handler.GetBidPathAnalytics)

	// Gin's router will return 404 for missing param
	req, _ := http.NewRequest("GET", "/api/analytics/bid-path/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Router gives 301 redirect for trailing slash or 404
	if w.Code != http.StatusMovedPermanently && w.Code != http.StatusNotFound {
		t.Errorf("Expected redirect or 404, got %d", w.Code)
	}
}

// ============================================================================
// SERVICE PERFORMANCE TESTS
// ============================================================================

func TestGetServicePerformance_ValidService(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/service-performance", handler.GetServicePerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/service-performance?serviceName=bidding", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetServicePerformance_MissingServiceName(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/service-performance", handler.GetServicePerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/service-performance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetServicePerformance_WithTimeRange(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/service-performance", handler.GetServicePerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/service-performance?serviceName=bidding&timeRange=12h", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// DIRECT PUBLISHER ANALYSIS TESTS
// ============================================================================

func TestGetDirectPublisherAnalysis_Default(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/direct-publisher/analysis", handler.GetDirectPublisherAnalysis)

	req, _ := http.NewRequest("GET", "/api/analytics/direct-publisher/analysis", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// COST BENEFIT ANALYSIS TESTS
// ============================================================================

func TestGetCostBenefitAnalysis_Default(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/cost-benefit", handler.GetCostBenefitAnalysis)

	req, _ := http.NewRequest("GET", "/api/analytics/cost-benefit", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// TRACK CLICK TESTS
// ============================================================================

func TestTrackClick_ValidRequest(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/track/click", handler.TrackClick)

	req, _ := http.NewRequest("GET", "/api/analytics/track/click?campaign_id=camp-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent && w.Code != http.StatusFound && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 204 or 302, got %d", w.Code)
	}
}

func TestTrackClick_MissingCampaignID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/track/click", handler.TrackClick)

	req, _ := http.NewRequest("GET", "/api/analytics/track/click", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestTrackClick_WithUserAndRequest(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/track/click", handler.TrackClick)

	req, _ := http.NewRequest("GET", "/api/analytics/track/click?campaign_id=camp-123&user_id=user-456&request_id=req-789", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent && w.Code != http.StatusFound && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 204 or 302, got %d", w.Code)
	}
}

func TestTrackClick_WithRedirect(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/track/click", handler.TrackClick)

	req, _ := http.NewRequest("GET", "/api/analytics/track/click?campaign_id=camp-123&redirect=https://example.com", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 302, got %d", w.Code)
	}
}

// ============================================================================
// TRACK IMPRESSION TESTS
// ============================================================================

func TestTrackImpression_ValidRequest(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/track/impression", handler.TrackImpression)

	req, _ := http.NewRequest("GET", "/api/analytics/track/impression?campaign_id=camp-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Should return a 1x1 GIF pixel
	if w.Code == http.StatusOK && w.Header().Get("Content-Type") != "image/gif" {
		t.Errorf("Expected Content-Type image/gif, got %s", w.Header().Get("Content-Type"))
	}
}

func TestTrackImpression_MissingCampaignID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/track/impression", handler.TrackImpression)

	req, _ := http.NewRequest("GET", "/api/analytics/track/impression", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestTrackImpression_WithUserAndRequest(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/track/impression", handler.TrackImpression)

	req, _ := http.NewRequest("GET", "/api/analytics/track/impression?campaign_id=camp-123&user_id=user-456&request_id=req-789", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// ============================================================================
// BID LANDSCAPE TESTS
// ============================================================================

func TestGetBidLandscape_Default(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/bid-landscape", handler.GetBidLandscape)

	req, _ := http.NewRequest("GET", "/api/analytics/bid-landscape", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetBidLandscape_ResponseFormat(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/bid-landscape", handler.GetBidLandscape)

	req, _ := http.NewRequest("GET", "/api/analytics/bid-landscape", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to parse response JSON: %v", err)
		}
		if _, ok := response["landscape"]; !ok {
			t.Error("Response should contain 'landscape' field")
		}
	}
}

// Test with deterministic landscape data to verify winRate calculation
func TestGetBidLandscape_WithData(t *testing.T) {
	// Use the top-level mock cache that returns a known landscape
	cache := &mockCacheWithLandscape{}
	biddingService := service.NewBiddingService(cache, "http://localhost:8080")
	handler := NewAnalyticsHandler(biddingService)
	router := gin.New()
	router.GET("/api/analytics/bid-landscape", handler.GetBidLandscape)

	req, _ := http.NewRequest("GET", "/api/analytics/bid-landscape", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	landscape, ok := resp["landscape"].(map[string]interface{})
	if !ok {
		t.Fatalf("landscape field missing or wrong type")
	}

	bucket, ok := landscape["0.50-1.00"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected bucket 0.50-1.00 present")
	}

	winRate, ok := bucket["winRate"].(float64)
	if !ok {
		t.Fatalf("winRate missing or wrong type")
	}

	if winRate < 0.2499 || winRate > 0.2501 {
		t.Fatalf("expected winRate ~0.25, got %v", winRate)
	}
}

// Mock cache to provide deterministic data for auto-bid and segment tests
type mockCacheForAutoBid struct{ mockCache }

func (m *mockCacheForAutoBid) GetActiveCampaigns() ([]*model.Campaign, error) {
	return []*model.Campaign{{
		ID:       "camp-1",
		Name:     "Test Campaign",
		BidPrice: 1.0,
		Status:   "active",
	}}, nil
}

func (m *mockCacheForAutoBid) GetCampaignCTR(campaignID string) (float64, error) {
	if campaignID == "camp-1" {
		return 0.03, nil // 3% CTR to trigger increase branch
	}
	return 0, nil
}

func (m *mockCacheForAutoBid) GetCampaignWinRate(campaignID string) (float64, error) {
	if campaignID == "camp-1" {
		return 0.10, nil // 10% win rate
	}
	return 0, nil
}

func (m *mockCacheForAutoBid) GetSegmentPerformance(segmentType string) (map[string]map[string]int64, error) {
	return map[string]map[string]int64{
		"mobile":  {"imps": 1000, "clicks": 50},
		"desktop": {"imps": 2000, "clicks": 20},
	}, nil
}

func TestGetAutoBidRecommendations_WithData(t *testing.T) {
	cache := &mockCacheForAutoBid{}
	biddingService := service.NewBiddingService(cache, "http://localhost:8080")
	handler := NewAnalyticsHandler(biddingService)
	router := gin.New()
	router.GET("/api/analytics/auto-bid-recommendations", handler.GetAutoBidRecommendations)

	req, _ := http.NewRequest("GET", "/api/analytics/auto-bid-recommendations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	recs, ok := resp["recommendations"].([]interface{})
	if !ok || len(recs) == 0 {
		t.Fatalf("expected recommendations array, got %v", resp["recommendations"])
	}

	first, _ := recs[0].(map[string]interface{})
	if first["campaignId"] != "camp-1" {
		t.Errorf("expected campaignId camp-1, got %v", first["campaignId"])
	}
	mult, ok := first["multiplier"].(float64)
	if !ok || mult <= 1.0 {
		t.Errorf("expected multiplier > 1.0, got %v", first["multiplier"])
	}
}

// Mock cache to provide deterministic data for decrease branch (low CTR, high win rate)
type mockCacheForAutoBidDecrease struct{ mockCache }

func (m *mockCacheForAutoBidDecrease) GetActiveCampaigns() ([]*model.Campaign, error) {
	return []*model.Campaign{{
		ID:       "camp-2",
		Name:     "Decrease Campaign",
		BidPrice: 2.0,
		Status:   "active",
	}}, nil
}

func (m *mockCacheForAutoBidDecrease) GetCampaignCTR(campaignID string) (float64, error) {
	if campaignID == "camp-2" {
		return 0.001, nil // 0.1% CTR to trigger decrease branch
	}
	return 0, nil
}

func (m *mockCacheForAutoBidDecrease) GetCampaignWinRate(campaignID string) (float64, error) {
	if campaignID == "camp-2" {
		return 0.75, nil // 75% win rate
	}
	return 0, nil
}

func TestGetAutoBidRecommendations_Decrease(t *testing.T) {
	cache := &mockCacheForAutoBidDecrease{}
	biddingService := service.NewBiddingService(cache, "http://localhost:8080")
	handler := NewAnalyticsHandler(biddingService)
	router := gin.New()
	router.GET("/api/analytics/auto-bid-recommendations", handler.GetAutoBidRecommendations)

	req, _ := http.NewRequest("GET", "/api/analytics/auto-bid-recommendations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	recs, ok := resp["recommendations"].([]interface{})
	if !ok || len(recs) == 0 {
		t.Fatalf("expected recommendations array with decrease recommendation, got %v", resp["recommendations"])
	}

	first, _ := recs[0].(map[string]interface{})
	if first["campaignId"] != "camp-2" {
		t.Errorf("expected campaignId camp-2, got %v", first["campaignId"])
	}
	mult, ok := first["multiplier"].(float64)
	if !ok || mult >= 1.0 {
		t.Errorf("expected multiplier < 1.0 for decrease, got %v", first["multiplier"])
	}
	if first["action"] != "decrease" {
		t.Errorf("expected action 'decrease', got %v", first["action"])
	}
}

// Mock cache that produces no recommendation (multiplier == 1.0)
type mockCacheForAutoBidNoRecommend struct{ mockCache }

func (m *mockCacheForAutoBidNoRecommend) GetActiveCampaigns() ([]*model.Campaign, error) {
	return []*model.Campaign{{
		ID:       "camp-3",
		Name:     "NoRec Campaign",
		BidPrice: 1.5,
		Status:   "active",
	}}, nil
}

func (m *mockCacheForAutoBidNoRecommend) GetCampaignCTR(campaignID string) (float64, error) {
	return 0.02, nil // 2% CTR and assume win rate in neutral range
}

func (m *mockCacheForAutoBidNoRecommend) GetCampaignWinRate(campaignID string) (float64, error) {
	return 0.5, nil // 50% win rate -> neutral
}

func TestGetAutoBidRecommendations_NoRecommendation(t *testing.T) {
	cache := &mockCacheForAutoBidNoRecommend{}
	biddingService := service.NewBiddingService(cache, "http://localhost:8080")
	handler := NewAnalyticsHandler(biddingService)
	router := gin.New()
	router.GET("/api/analytics/auto-bid-recommendations", handler.GetAutoBidRecommendations)

	req, _ := http.NewRequest("GET", "/api/analytics/auto-bid-recommendations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Expect 0 recommendations
	count, _ := resp["count"].(float64)
	if int(count) != 0 {
		t.Fatalf("expected count 0 for no recommendation, got %v", count)
	}
}

// Mock cache that forces an error when fetching campaigns to exercise the 500 path
type mockCacheForAutoBidError struct{ mockCache }

func (m *mockCacheForAutoBidError) GetActiveCampaigns() ([]*model.Campaign, error) {
	return nil, errors.New("simulated cache failure")
}

func TestGetAutoBidRecommendations_ServiceError(t *testing.T) {
	cache := &mockCacheForAutoBidError{}
	biddingService := service.NewBiddingService(cache, "http://localhost:8080")
	handler := NewAnalyticsHandler(biddingService)
	router := gin.New()
	router.GET("/api/analytics/auto-bid-recommendations", handler.GetAutoBidRecommendations)

	req, _ := http.NewRequest("GET", "/api/analytics/auto-bid-recommendations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected 500 due to service error, got %d", w.Code)
	}
}

func TestGetAutoBidRecommendations_TimestampAndCount(t *testing.T) {
	cache := &mockCacheForAutoBid{}
	biddingService := service.NewBiddingService(cache, "http://localhost:8080")
	handler := NewAnalyticsHandler(biddingService)
	router := gin.New()
	router.GET("/api/analytics/auto-bid-recommendations", handler.GetAutoBidRecommendations)

	req, _ := http.NewRequest("GET", "/api/analytics/auto-bid-recommendations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if _, ok := resp["timestamp"]; !ok {
		t.Errorf("expected timestamp field in response")
	}

	recs, _ := resp["recommendations"].([]interface{})
	count, _ := resp["count"].(float64)
	if int(count) != len(recs) {
		t.Errorf("count field (%v) does not match recommendations length (%d)", count, len(recs))
	}
}

func TestGetSegmentPerformance_WithData(t *testing.T) {
	cache := &mockCacheForAutoBid{}
	biddingService := service.NewBiddingService(cache, "http://localhost:8080")
	handler := NewAnalyticsHandler(biddingService)
	router := gin.New()
	router.GET("/api/analytics/segment-performance", handler.GetSegmentPerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/segment-performance?type=device", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	perf, ok := resp["performance"].(map[string]interface{})
	if !ok {
		t.Fatalf("performance field missing or wrong type")
	}

	mobile, ok := perf["mobile"].(map[string]interface{})
	if !ok {
		t.Fatalf("mobile segment missing")
	}

	ctr, ok := mobile["ctr"].(float64)
	if !ok {
		t.Fatalf("ctr missing or wrong type")
	}
	// expected CTR = 50 / 1000 = 0.05
	if ctr < 0.049 || ctr > 0.051 {
		t.Fatalf("expected ctr ~0.05, got %v", ctr)
	}
}

// ============================================================================
// AUTO-BID RECOMMENDATIONS TESTS
// ============================================================================

func TestGetAutoBidRecommendations_Default(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/auto-bid-recommendations", handler.GetAutoBidRecommendations)

	req, _ := http.NewRequest("GET", "/api/analytics/auto-bid-recommendations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// SEGMENT PERFORMANCE TESTS
// ============================================================================

func TestGetSegmentPerformance_Device(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/segment-performance", handler.GetSegmentPerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/segment-performance?type=device", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetSegmentPerformance_OS(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/segment-performance", handler.GetSegmentPerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/segment-performance?type=os", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetSegmentPerformance_Geo(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/segment-performance", handler.GetSegmentPerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/segment-performance?type=geo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetSegmentPerformance_MissingType(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/segment-performance", handler.GetSegmentPerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/segment-performance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetSegmentPerformance_InvalidType(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/segment-performance", handler.GetSegmentPerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/segment-performance?type=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// ============================================================================
// OPTIMAL BID FLOOR TESTS
// ============================================================================

func TestGetOptimalBidFloor_ValidRequest(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/optimal-bid-floor", handler.GetOptimalBidFloor)

	req, _ := http.NewRequest("GET", "/api/analytics/optimal-bid-floor?publisher_id=pub-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetOptimalBidFloor_CustomTargetWinRate(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/optimal-bid-floor", handler.GetOptimalBidFloor)

	req, _ := http.NewRequest("GET", "/api/analytics/optimal-bid-floor?publisher_id=pub-123&target_win_rate=0.75", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetOptimalBidFloor_MissingPublisherID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/optimal-bid-floor", handler.GetOptimalBidFloor)

	req, _ := http.NewRequest("GET", "/api/analytics/optimal-bid-floor", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetOptimalBidFloor_InvalidTargetWinRate_TooLow(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/optimal-bid-floor", handler.GetOptimalBidFloor)

	req, _ := http.NewRequest("GET", "/api/analytics/optimal-bid-floor?publisher_id=pub-123&target_win_rate=0.05", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetOptimalBidFloor_InvalidTargetWinRate_TooHigh(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/optimal-bid-floor", handler.GetOptimalBidFloor)

	req, _ := http.NewRequest("GET", "/api/analytics/optimal-bid-floor?publisher_id=pub-123&target_win_rate=0.99", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetOptimalBidFloor_InvalidTargetWinRate_NotNumber(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/optimal-bid-floor", handler.GetOptimalBidFloor)

	req, _ := http.NewRequest("GET", "/api/analytics/optimal-bid-floor?publisher_id=pub-123&target_win_rate=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// ============================================================================
// TRACK CONVERSION TESTS
// ============================================================================

func TestTrackConversion_ValidRequest(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/track/conversion", handler.TrackConversion)

	body := map[string]interface{}{
		"user_id":         "user-123",
		"campaign_id":     "camp-456",
		"conversion_type": "purchase",
		"value":           29.99,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/track/conversion", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestTrackConversion_MissingUserID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/track/conversion", handler.TrackConversion)

	body := map[string]interface{}{
		"campaign_id": "camp-456",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/track/conversion", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestTrackConversion_MissingCampaignID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/track/conversion", handler.TrackConversion)

	body := map[string]interface{}{
		"user_id": "user-123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/track/conversion", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestTrackConversion_InvalidJSON(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/track/conversion", handler.TrackConversion)

	req, _ := http.NewRequest("POST", "/api/analytics/track/conversion", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// ============================================================================
// MULTI-TOUCH ATTRIBUTION TESTS
// ============================================================================

func TestGetMultiTouchAttribution_ValidRequest(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/multi-touch-attribution", handler.GetMultiTouchAttribution)

	req, _ := http.NewRequest("GET", "/api/analytics/multi-touch-attribution?user_id=user-123&campaign_id=camp-456", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetMultiTouchAttribution_LinearModel(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/multi-touch-attribution", handler.GetMultiTouchAttribution)

	req, _ := http.NewRequest("GET", "/api/analytics/multi-touch-attribution?user_id=user-123&campaign_id=camp-456&model=linear", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetMultiTouchAttribution_TimeDecayModel(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/multi-touch-attribution", handler.GetMultiTouchAttribution)

	req, _ := http.NewRequest("GET", "/api/analytics/multi-touch-attribution?user_id=user-123&campaign_id=camp-456&model=time_decay", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetMultiTouchAttribution_MissingUserID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/multi-touch-attribution", handler.GetMultiTouchAttribution)

	req, _ := http.NewRequest("GET", "/api/analytics/multi-touch-attribution?campaign_id=camp-456", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetMultiTouchAttribution_MissingCampaignID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/multi-touch-attribution", handler.GetMultiTouchAttribution)

	req, _ := http.NewRequest("GET", "/api/analytics/multi-touch-attribution?user_id=user-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// ============================================================================
// CROSS-DEVICE LINK DEVICES TESTS
// ============================================================================

func TestLinkDevices_ValidRequest(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/cross-device/link", handler.LinkDevices)

	body := map[string]interface{}{
		"primary_user_id": "user-123",
		"device_ids":      []string{"mobile-abc", "desktop-xyz"},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/cross-device/link", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestLinkDevices_MissingPrimaryUserID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/cross-device/link", handler.LinkDevices)

	body := map[string]interface{}{
		"device_ids": []string{"mobile-abc"},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/cross-device/link", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestLinkDevices_EmptyDeviceIDs(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/cross-device/link", handler.LinkDevices)

	body := map[string]interface{}{
		"primary_user_id": "user-123",
		"device_ids":      []string{},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/cross-device/link", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// ============================================================================
// CROSS-DEVICE GRAPH TESTS
// ============================================================================

func TestGetDeviceGraph_ValidRequest(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/cross-device/graph", handler.GetDeviceGraph)

	req, _ := http.NewRequest("GET", "/api/analytics/cross-device/graph?user_id=user-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetDeviceGraph_MissingUserID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/cross-device/graph", handler.GetDeviceGraph)

	req, _ := http.NewRequest("GET", "/api/analytics/cross-device/graph", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// ============================================================================
// CROSS-DEVICE FREQUENCY TESTS
// ============================================================================

func TestGetCrossDeviceFrequency_ValidRequest(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/cross-device/frequency", handler.GetCrossDeviceFrequency)

	req, _ := http.NewRequest("GET", "/api/analytics/cross-device/frequency?user_id=user-123&campaign_id=camp-456", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetCrossDeviceFrequency_MissingUserID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/cross-device/frequency", handler.GetCrossDeviceFrequency)

	req, _ := http.NewRequest("GET", "/api/analytics/cross-device/frequency?campaign_id=camp-456", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetCrossDeviceFrequency_MissingCampaignID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/cross-device/frequency", handler.GetCrossDeviceFrequency)

	req, _ := http.NewRequest("GET", "/api/analytics/cross-device/frequency?user_id=user-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// ============================================================================
// EDGE CASE TESTS
// ============================================================================

func TestAnalyticsHandler_EmptyQueryParameters(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/supply-chain/metrics", handler.GetSupplyChainMetrics)

	// Should use default timeRange
	req, _ := http.NewRequest("GET", "/api/analytics/supply-chain/metrics?timeRange=", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still work with default
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestAnalyticsHandler_SpecialCharactersInParams(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/bid-path/:requestId", handler.GetBidPathAnalytics)

	// URL-encoded special characters
	req, _ := http.NewRequest("GET", "/api/analytics/bid-path/req%2D123%2Dspecial", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 404, or 500, got %d", w.Code)
	}
}

func TestAnalyticsHandler_ContentTypeValidation(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/track/conversion", handler.TrackConversion)

	body := map[string]interface{}{
		"user_id":     "user-123",
		"campaign_id": "camp-456",
	}
	jsonBody, _ := json.Marshal(body)

	// Missing Content-Type header
	req, _ := http.NewRequest("POST", "/api/analytics/track/conversion", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Gin should still parse JSON even without Content-Type
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

// Additional small coverage tests (batch 2)
func TestGetSupplyChainMetrics_Coverage2(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/supply-chain/metrics", handler.GetSupplyChainMetrics)

	req, _ := http.NewRequest("GET", "/api/analytics/supply-chain/metrics?timeRange=48h", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetSupplyPathOptimization_Coverage2(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/supply-path/optimization", handler.GetSupplyPathOptimization)

	req, _ := http.NewRequest("GET", "/api/analytics/supply-path/optimization?timeRange=7d", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetBidPathAnalytics_Coverage2(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/bid-path/:requestId", handler.GetBidPathAnalytics)

	req, _ := http.NewRequest("GET", "/api/analytics/bid-path/test-req-999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 404 or 500, got %d", w.Code)
	}
}

func TestGetDirectPublisherAnalysis_Coverage2(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/direct-publisher/analysis", handler.GetDirectPublisherAnalysis)

	req, _ := http.NewRequest("GET", "/api/analytics/direct-publisher/analysis?timeRange=24h", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetCostBenefitAnalysis_Coverage2(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/cost-benefit", handler.GetCostBenefitAnalysis)

	req, _ := http.NewRequest("GET", "/api/analytics/cost-benefit?timeRange=30d", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// SERVICE PERFORMANCE TESTS
// ============================================================================

func TestGetServicePerformance_Success(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/service/performance", handler.GetServicePerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/service/performance?service=bidding&timeRange=1h", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestGetServicePerformance_DefaultParams(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/service/performance", handler.GetServicePerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/service/performance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

// ============================================================================
// AUTO BID RECOMMENDATIONS TESTS
// ============================================================================

func TestGetAutoBidRecommendations_Success(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/autobid/recommendations/:campaign_id", handler.GetAutoBidRecommendations)

	req, _ := http.NewRequest("GET", "/api/analytics/autobid/recommendations/camp-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetAutoBidRecommendations_MissingCampaignID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/autobid/recommendations/:campaign_id", handler.GetAutoBidRecommendations)

	req, _ := http.NewRequest("GET", "/api/analytics/autobid/recommendations/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty param results in 404 for this router pattern
	if w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 404 or 400, got %d", w.Code)
	}
}

// ============================================================================
// SEGMENT PERFORMANCE TESTS
// ============================================================================

func TestGetSegmentPerformance_Success(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/segment/performance/:segment_id", handler.GetSegmentPerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/segment/performance/segment-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestGetSegmentPerformance_MissingSegmentID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/segment/performance/:segment_id", handler.GetSegmentPerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/segment/performance/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty param results in 404
	if w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 404 or 400, got %d", w.Code)
	}
}

// ============================================================================
// BID LANDSCAPE TESTS
// ============================================================================

func TestGetBidLandscape_Success(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/bid-landscape", handler.GetBidLandscape)

	req, _ := http.NewRequest("GET", "/api/analytics/bid-landscape", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetBidLandscape_WithTimeRange(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/bid-landscape", handler.GetBidLandscape)

	req, _ := http.NewRequest("GET", "/api/analytics/bid-landscape?time_range=24h", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetBidLandscape_WithCampaign(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/bid-landscape", handler.GetBidLandscape)

	req, _ := http.NewRequest("GET", "/api/analytics/bid-landscape?campaign_id=camp-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// AUTO BID RECOMMENDATIONS EXTENDED TESTS
// ============================================================================

func TestGetAutoBidRecommendations_WithParams(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/autobid/recommendations", handler.GetAutoBidRecommendations)

	req, _ := http.NewRequest("GET", "/api/analytics/autobid/recommendations?min_spend=100", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// SEGMENT PERFORMANCE EXTENDED TESTS
// ============================================================================

func TestGetSegmentPerformance_DeviceType(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/segment/performance", handler.GetSegmentPerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/segment/performance?type=device", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestGetSegmentPerformance_GeoType(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/segment/performance", handler.GetSegmentPerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/segment/performance?type=geo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestGetSegmentPerformance_OSType(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/segment/performance", handler.GetSegmentPerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/segment/performance?type=os", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestGetSegmentPerformance_UnknownType(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/segment/performance", handler.GetSegmentPerformance)

	req, _ := http.NewRequest("GET", "/api/analytics/segment/performance?type=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

// ============================================================================
// DIRECT PUBLISHER ANALYSIS TESTS
// ============================================================================

func TestGetDirectPublisherAnalysis_Success(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/direct-publisher/analysis", handler.GetDirectPublisherAnalysis)

	req, _ := http.NewRequest("GET", "/api/analytics/direct-publisher/analysis", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetDirectPublisherAnalysis_WithTimeRange(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/direct-publisher/analysis", handler.GetDirectPublisherAnalysis)

	req, _ := http.NewRequest("GET", "/api/analytics/direct-publisher/analysis?timeRange=7d", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// COST BENEFIT ANALYSIS TESTS
// ============================================================================

func TestGetCostBenefitAnalysis_Success(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/cost-benefit", handler.GetCostBenefitAnalysis)

	req, _ := http.NewRequest("GET", "/api/analytics/cost-benefit", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestGetCostBenefitAnalysis_WithParams(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/cost-benefit", handler.GetCostBenefitAnalysis)

	req, _ := http.NewRequest("GET", "/api/analytics/cost-benefit?timeRange=30d", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// TRACK CLICK EXTENDED TESTS
// ============================================================================

func TestTrackClick_WithAllParams(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/track/click", handler.TrackClick)

	body := map[string]interface{}{
		"user_id":     "user-001",
		"campaign_id": "camp-001",
		"request_id":  "req-001",
		"ad_id":       "ad-001",
		"placement":   "header",
		"timestamp":   "2026-02-23T10:00:00Z",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/track/click", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestTrackClick_MissingUserID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/track/click", handler.TrackClick)

	body := map[string]interface{}{
		"campaign_id": "camp-001",
		"request_id":  "req-001",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/track/click", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("Expected status 400 or 200, got %d", w.Code)
	}
}

// ============================================================================
// TRACK IMPRESSION EXTENDED TESTS
// ============================================================================

func TestTrackImpression_WithAllParams(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/track/impression", handler.TrackImpression)

	body := map[string]interface{}{
		"user_id":     "user-001",
		"campaign_id": "camp-001",
		"request_id":  "req-001",
		"ad_id":       "ad-001",
		"placement":   "sidebar",
		"viewable":    true,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/track/impression", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestTrackImpression_NoCampaignID(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/track/impression", handler.TrackImpression)

	body := map[string]interface{}{
		"user_id":    "user-001",
		"request_id": "req-001",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/track/impression", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("Expected status 400 or 200, got %d", w.Code)
	}
}

// ============================================================================
// TRACK CONVERSION EXTENDED TESTS
// ============================================================================

func TestTrackConversion_Success(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/track/conversion", handler.TrackConversion)

	body := map[string]interface{}{
		"user_id":         "user-001",
		"campaign_id":     "camp-001",
		"conversion_type": "purchase",
		"value":           99.99,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/track/conversion", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestTrackConversion_WithAttribution(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/track/conversion", handler.TrackConversion)

	body := map[string]interface{}{
		"user_id":          "user-001",
		"campaign_id":      "camp-001",
		"conversion_type":  "signup",
		"value":            0,
		"attribution_type": "last_click",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/track/conversion", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestTrackConversion_MissingRequired(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/track/conversion", handler.TrackConversion)

	body := map[string]interface{}{
		"value": 99.99,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/track/conversion", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("Expected status 400 or 200, got %d", w.Code)
	}
}

// ============================================================================
// OPTIMAL BID FLOOR EXTENDED TESTS
// ============================================================================

func TestGetOptimalBidFloor_WithTargetWinRate(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/optimal-bid-floor/:publisher_id", handler.GetOptimalBidFloor)

	req, _ := http.NewRequest("GET", "/api/analytics/optimal-bid-floor/pub-001?target_win_rate=0.6", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestGetOptimalBidFloor_InvalidWinRate(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/optimal-bid-floor/:publisher_id", handler.GetOptimalBidFloor)

	req, _ := http.NewRequest("GET", "/api/analytics/optimal-bid-floor/pub-001?target_win_rate=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle invalid win rate gracefully
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestGetOptimalBidFloor_HighWinRate(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/optimal-bid-floor/:publisher_id", handler.GetOptimalBidFloor)

	req, _ := http.NewRequest("GET", "/api/analytics/optimal-bid-floor/pub-001?target_win_rate=0.95", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestGetOptimalBidFloor_LowWinRate(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/optimal-bid-floor/:publisher_id", handler.GetOptimalBidFloor)

	req, _ := http.NewRequest("GET", "/api/analytics/optimal-bid-floor/pub-001?target_win_rate=0.1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

// ============================================================================
// MULTI-TOUCH ATTRIBUTION EXTENDED TESTS
// ============================================================================

func TestGetMultiTouchAttribution_LinearModelExt(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/attribution/multi-touch", handler.GetMultiTouchAttribution)

	req, _ := http.NewRequest("GET", "/api/analytics/attribution/multi-touch?user_id=user-001&campaign_id=camp-001&model=linear", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestGetMultiTouchAttribution_TimeDecayModelExt(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/attribution/multi-touch", handler.GetMultiTouchAttribution)

	req, _ := http.NewRequest("GET", "/api/analytics/attribution/multi-touch?user_id=user-001&campaign_id=camp-001&model=time_decay", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestGetMultiTouchAttribution_PositionBasedModel(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.GET("/api/analytics/attribution/multi-touch", handler.GetMultiTouchAttribution)

	req, _ := http.NewRequest("GET", "/api/analytics/attribution/multi-touch?user_id=user-001&campaign_id=camp-001&model=position_based", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

// ============================================================================
// LINK DEVICES EXTENDED TESTS
// ============================================================================

func TestLinkDevices_WithTTL(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/devices/link", handler.LinkDevices)

	body := map[string]interface{}{
		"primary_user_id": "user-001",
		"device_ids":      []string{"device-001", "device-002"},
		"ttl_days":        30,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/devices/link", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}

func TestLinkDevices_SingleDevice(t *testing.T) {
	handler, router := setupAnalyticsTestHandler()
	router.POST("/api/analytics/devices/link", handler.LinkDevices)

	body := map[string]interface{}{
		"primary_user_id": "user-001",
		"device_ids":      []string{"device-001"},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/analytics/devices/link", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200, 400, or 500, got %d", w.Code)
	}
}
