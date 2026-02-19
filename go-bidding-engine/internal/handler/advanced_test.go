package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/taskirx/go-bidding-engine/internal/model"
	"github.com/taskirx/go-bidding-engine/internal/service"
)

// ============================================================================
// TEST SETUP
// ============================================================================

func init() {
	gin.SetMode(gin.TestMode)
}

// mockCache implements cache.Cache for testing
type mockCache struct{}

func (m *mockCache) Get(key string) (string, error)                       { return "", nil }
func (m *mockCache) Set(key string, value interface{}, ttl int64) error   { return nil }
func (m *mockCache) GetActiveCampaigns() ([]*model.Campaign, error)       { return nil, nil }
func (m *mockCache) SetActiveCampaigns(campaigns []*model.Campaign) error { return nil }
func (m *mockCache) GetCampaign(campaignID string) (*model.Campaign, error) {
	return nil, nil
}
func (m *mockCache) SetCampaign(campaign *model.Campaign) error                         { return nil }
func (m *mockCache) IncrementBidCount() error                                           { return nil }
func (m *mockCache) IncrementWinCount() error                                           { return nil }
func (m *mockCache) GetBidCount() (int64, error)                                        { return 0, nil }
func (m *mockCache) GetWinCount() (int64, error)                                        { return 0, nil }
func (m *mockCache) RecordLatency(latencyMs float64) error                              { return nil }
func (m *mockCache) GetAverageLatency() (float64, error)                                { return 0, nil }
func (m *mockCache) SetUserSegments(userID string, segments []string) error             { return nil }
func (m *mockCache) GetUserSegments(userID string) ([]string, error)                    { return nil, nil }
func (m *mockCache) SetGeoRules(countryCode string, rules map[string]interface{}) error { return nil }
func (m *mockCache) GetGeoRules(countryCode string) (map[string]interface{}, error)     { return nil, nil }
func (m *mockCache) IncrementCampaignSpend(campaignID string, amount float64) (float64, error) {
	return 0, nil
}
func (m *mockCache) GetCampaignSpend(campaignID string) (float64, error) { return 0, nil }
func (m *mockCache) IncrementBidFormat(format string) error              { return nil }
func (m *mockCache) GetBidFormats() (map[string]int64, error)            { return nil, nil }
func (m *mockCache) IncrementPublisherFraud(publisherID string) error    { return nil }
func (m *mockCache) IsRequestDuplicate(requestID string, ttlSeconds int) (bool, error) {
	return false, nil
}
func (m *mockCache) IncrementUserFrequency(userID, campaignID string, windowSecs int) (int64, error) {
	return 0, nil
}
func (m *mockCache) GetUserFrequency(userID, campaignID string) (int64, error)          { return 0, nil }
func (m *mockCache) GetCampaignCTR(campaignID string) (float64, error)                  { return 0, nil }
func (m *mockCache) GetCampaignWinRate(campaignID string) (float64, error)              { return 0, nil }
func (m *mockCache) IncrementCampaignClicks(campaignID string) error                    { return nil }
func (m *mockCache) IncrementCampaignImpressions(campaignID string) error               { return nil }
func (m *mockCache) IncrementCampaignBids(campaignID string) error                      { return nil }
func (m *mockCache) IncrementCampaignWins(campaignID string) error                      { return nil }
func (m *mockCache) RecordBidInBucket(priceBucket string) error                         { return nil }
func (m *mockCache) RecordWinInBucket(priceBucket string) error                         { return nil }
func (m *mockCache) GetBidLandscape() (map[string]map[string]int64, error)              { return nil, nil }
func (m *mockCache) IncrementSegmentImpressions(segmentType, segmentValue string) error { return nil }
func (m *mockCache) IncrementSegmentClicks(segmentType, segmentValue string) error      { return nil }
func (m *mockCache) GetSegmentPerformance(segmentType string) (map[string]map[string]int64, error) {
	return nil, nil
}
func (m *mockCache) RecordPublisherBidAttempt(publisherID string, bidPrice float64, won bool) error {
	return nil
}
func (m *mockCache) GetOptimalBidFloor(publisherID string, targetWinRate float64) (float64, error) {
	return 0, nil
}
func (m *mockCache) RecordImpression(userID, campaignID, requestID string, ttlHours int) error {
	return nil
}
func (m *mockCache) RecordClick(userID, campaignID, requestID string, ttlHours int) error { return nil }
func (m *mockCache) GetAttribution(userID, campaignID string) (string, string, error) {
	return "", "", nil
}
func (m *mockCache) RecordTouchpoint(userID, campaignID, touchpointType, requestID string, ttlDays int) error {
	return nil
}
func (m *mockCache) GetTouchpoints(userID, campaignID string) ([]model.Touchpoint, error) {
	return nil, nil
}
func (m *mockCache) GetMultiTouchAttribution(userID, campaignID, modelType string) ([]model.AttributionCredit, error) {
	return nil, nil
}
func (m *mockCache) RecordUserEvent(userID, campaignID, eventType string, ttlDays int) error {
	return nil
}
func (m *mockCache) GetUserEvents(userID string, eventTypes []string) (map[string][]string, error) {
	return nil, nil
}
func (m *mockCache) HasUserEvent(userID, campaignID, eventType string) (bool, error) {
	return false, nil
}
func (m *mockCache) LinkDevices(primaryUserID string, deviceIDs []string, ttlDays int) error {
	return nil
}
func (m *mockCache) GetLinkedDevices(deviceID string) ([]string, error) { return nil, nil }
func (m *mockCache) GetPrimaryUserID(deviceID string) (string, error)   { return "", nil }
func (m *mockCache) GetCrossDeviceFrequency(primaryUserID, campaignID string) (int64, error) {
	return 0, nil
}
func (m *mockCache) StoreBidPathAnalytics(analytics *model.BidPathAnalytics) error { return nil }
func (m *mockCache) GetBidPathAnalytics(requestID string) (*model.BidPathAnalytics, error) {
	return nil, nil
}
func (m *mockCache) GetSupplyChainMetrics(timeRange string) (*model.SupplyChainMetrics, error) {
	return nil, nil
}
func (m *mockCache) GetServiceMetrics(serviceName string, timeRange string) (*model.ServiceMetrics, error) {
	return nil, nil
}

func setupTestHandler() (*AdvancedHandler, *gin.Engine) {
	cache := &mockCache{}
	biddingService := service.NewBiddingService(cache, "http://localhost:8080")
	handler := NewAdvancedHandler(biddingService)

	router := gin.New()
	return handler, router
}

// ============================================================================
// BID LANDSCAPE TESTS
// ============================================================================

func TestHandleBidLandscapeAnalysis(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/bid-landscape/analyze", handler.HandleBidLandscapeAnalysis)

	reqBody := BidLandscapeRequest{
		CampaignID:  "camp-1",
		PublisherID: "pub-1",
		DeviceType:  "mobile",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/bid-landscape/analyze", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleBidLandscapeAnalysis_InvalidRequest(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/bid-landscape/analyze", handler.HandleBidLandscapeAnalysis)

	// Missing required campaign_id
	reqBody := map[string]string{"publisher_id": "pub-1"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/bid-landscape/analyze", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleRecordBid(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/bid-landscape/record", handler.HandleRecordBid)

	reqBody := RecordBidRequest{
		PublisherID: "pub-1",
		DeviceType:  "mobile",
		BidPrice:    2.5,
		WinPrice:    2.0,
		Won:         true,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/bid-landscape/record", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["status"] != "recorded" {
		t.Errorf("Expected status 'recorded', got '%s'", response["status"])
	}
}

// ============================================================================
// CREATIVE OPTIMIZATION TESTS
// ============================================================================

func TestHandleCreativeSelect(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/creative/select", handler.HandleCreativeSelect)

	reqBody := CreativeSelectRequest{
		CampaignID: "camp-1",
		Formats:    []string{"banner", "video"},
		UserID:     "user-123",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/creative/select", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// ============================================================================
// INCREMENTALITY TESTS
// ============================================================================

func TestHandleIncrementalityEval(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/incrementality/evaluate", handler.HandleIncrementalityEval)

	reqBody := IncrementalityEvalRequest{
		CampaignID:     "camp-1",
		ExperimentID:   "exp-1",
		UserID:         "user-123",
		ControlPercent: 10.0,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/incrementality/evaluate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleGetExperimentResults(t *testing.T) {
	handler, router := setupTestHandler()
	router.GET("/api/advanced/incrementality/results/:experiment_id", handler.HandleGetExperimentResults)

	req, _ := http.NewRequest("GET", "/api/advanced/incrementality/results/exp-1", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleRecordConversion(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/incrementality/conversion", handler.HandleRecordConversion)

	reqBody := RecordConversionRequest{
		ExperimentID: "exp-1",
		UserID:       "user-123",
		IsControl:    false,
		Revenue:      50.0,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/incrementality/conversion", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// ============================================================================
// PRIVACY SANDBOX TESTS
// ============================================================================

func TestHandleRegisterTopic(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/privacy/topic", handler.HandleRegisterTopic)

	reqBody := TopicRegistrationRequest{
		UserID:  "user-123",
		TopicID: 42,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/privacy/topic", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleAddToInterestGroup(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/privacy/interest-group", handler.HandleAddToInterestGroup)

	reqBody := InterestGroupRequest{
		UserID:  "user-123",
		GroupID: "sports_fans",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/privacy/interest-group", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleGetInterestGroups(t *testing.T) {
	handler, router := setupTestHandler()
	router.GET("/api/advanced/privacy/interest-groups/:user_id", handler.HandleGetInterestGroups)

	req, _ := http.NewRequest("GET", "/api/advanced/privacy/interest-groups/user-123", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// ============================================================================
// CONTEXTUAL AI TESTS
// ============================================================================

func TestHandleContextAnalysis(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/contextual/analyze", handler.HandleContextAnalysis)

	reqBody := ContextAnalysisRequest{
		CampaignID:       "camp-1",
		PublisherID:      "pub-1",
		BrandSafetyLevel: "standard",
		Context: map[string]interface{}{
			"page_title":   "Top Travel Destinations",
			"page_content": "Explore the best vacation spots",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/contextual/analyze", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// ============================================================================
// REAL-TIME ALERTS TESTS
// ============================================================================

func TestHandleCheckAlerts(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/alerts/check", handler.HandleCheckAlerts)

	reqBody := AlertCheckRequest{
		CampaignID:   "camp-1",
		CurrentSpend: 850.0,
		Budget:       1000.0,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/alerts/check", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleRecordMetrics(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/alerts/metrics", handler.HandleRecordMetrics)

	reqBody := RecordMetricsRequest{
		CampaignID: "camp-1",
		Spend:      100.0,
		CTR:        0.02,
		CVR:        0.01,
		WinRate:    0.15,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/alerts/metrics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// ============================================================================
// COMPETITIVE INTELLIGENCE TESTS
// ============================================================================

func TestHandleCompetitiveAnalysis(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/competitive/analyze", handler.HandleCompetitiveAnalysis)

	reqBody := CompetitiveAnalysisRequest{
		CampaignID:      "camp-1",
		PublisherID:     "pub-1",
		AdSlotID:        "slot-1",
		CompetitiveMode: "aggressive",
		Competitors:     []string{"competitor-1", "competitor-2"},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/competitive/analyze", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleRecordAuctionOutcome(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/competitive/outcome", handler.HandleRecordAuctionOutcome)

	reqBody := AuctionOutcomeRequest{
		PublisherID:  "pub-1",
		AdSlotID:     "slot-1",
		BidPrice:     2.5,
		WinningPrice: 2.8,
		Won:          false,
		WinnerID:     "competitor-1",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/competitive/outcome", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleGetMarketReport(t *testing.T) {
	handler, router := setupTestHandler()
	router.GET("/api/advanced/competitive/report", handler.HandleGetMarketReport)

	req, _ := http.NewRequest("GET", "/api/advanced/competitive/report", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// ============================================================================
// UNIFIED ID TESTS
// ============================================================================

func TestHandleResolveIdentity(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/identity/resolve", handler.HandleResolveIdentity)

	reqBody := IdentityResolveRequest{
		CampaignID:      "camp-1",
		UserID:          "user-123",
		DeviceID:        "device-abc",
		Providers:       []string{"uid2", "id5"},
		ConsentRequired: true,
		HasConsent:      true,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/identity/resolve", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleLinkIdentities(t *testing.T) {
	handler, router := setupTestHandler()
	router.POST("/api/advanced/identity/link", handler.HandleLinkIdentities)

	reqBody := LinkIdentitiesRequest{
		ID1:        "uid2-abc",
		Provider1:  "uid2",
		ID2:        "id5-xyz",
		Provider2:  "id5",
		DeviceType: "mobile",
		Confidence: 0.9,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/advanced/identity/link", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleGetIdentityReport(t *testing.T) {
	handler, router := setupTestHandler()
	router.GET("/api/advanced/identity/report", handler.HandleGetIdentityReport)

	req, _ := http.NewRequest("GET", "/api/advanced/identity/report", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleGetCrossDeviceReach(t *testing.T) {
	handler, router := setupTestHandler()
	router.GET("/api/advanced/identity/cross-device-reach", handler.HandleGetCrossDeviceReach)

	req, _ := http.NewRequest("GET", "/api/advanced/identity/cross-device-reach", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// ============================================================================
// STATUS TESTS
// ============================================================================

func TestHandleAdvancedServicesStatus(t *testing.T) {
	handler, router := setupTestHandler()
	router.GET("/api/advanced/status", handler.HandleAdvancedServicesStatus)

	req, _ := http.NewRequest("GET", "/api/advanced/status", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if _, ok := response["healthy"]; !ok {
		t.Error("Expected 'healthy' field in response")
	}
	if _, ok := response["services"]; !ok {
		t.Error("Expected 'services' field in response")
	}
}
