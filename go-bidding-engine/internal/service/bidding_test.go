package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// MockCache implements cache.Cache for testing
type MockCache struct {
	campaigns       []*model.Campaign
	userSegments    map[string][]string
	geoRules        map[string]map[string]interface{}
	spend           map[string]float64
	kv              map[string]string
	touchpoints     map[string][]model.Touchpoint
	userEvents      map[string]bool // key: "userID:campaignID:eventType"
	frequencies     map[string]int64
	ctr             map[string]float64
	winRate         map[string]float64
	primaryUserID   map[string]string   // deviceID -> primaryUserID mapping
	linkedDevices   map[string][]string // primaryUserID -> []deviceIDs mapping
	crossDeviceFreq map[string]int64    // "primaryUserID:campaignID" -> frequency
}

func NewMockCache() *MockCache {
	return &MockCache{
		userSegments:    make(map[string][]string),
		geoRules:        make(map[string]map[string]interface{}),
		spend:           make(map[string]float64),
		kv:              make(map[string]string),
		touchpoints:     make(map[string][]model.Touchpoint),
		userEvents:      make(map[string]bool),
		frequencies:     make(map[string]int64),
		ctr:             make(map[string]float64),
		winRate:         make(map[string]float64),
		primaryUserID:   make(map[string]string),
		linkedDevices:   make(map[string][]string),
		crossDeviceFreq: make(map[string]int64),
	}
}

func (m *MockCache) GetActiveCampaigns() ([]*model.Campaign, error) {
	return m.campaigns, nil
}
func (m *MockCache) SetActiveCampaigns(c []*model.Campaign) error {
	m.campaigns = c
	return nil
}
func (m *MockCache) GetCampaign(id string) (*model.Campaign, error) { return nil, nil }
func (m *MockCache) SetCampaign(c *model.Campaign) error            { return nil }
func (m *MockCache) IncrementBidCount() error                       { return nil }
func (m *MockCache) IncrementWinCount() error                       { return nil }
func (m *MockCache) GetBidCount() (int64, error)                    { return 0, nil }
func (m *MockCache) GetWinCount() (int64, error)                    { return 0, nil }
func (m *MockCache) RecordLatency(l float64) error                  { return nil }
func (m *MockCache) GetAverageLatency() (float64, error)            { return 0, nil }

func (m *MockCache) SetUserSegments(userID string, segments []string) error {
	m.userSegments[userID] = segments
	return nil
}
func (m *MockCache) GetUserSegments(userID string) ([]string, error) {
	return m.userSegments[userID], nil
}

func (m *MockCache) SetGeoRules(code string, rules map[string]interface{}) error {
	m.geoRules[code] = rules
	return nil
}
func (m *MockCache) GetGeoRules(code string) (map[string]interface{}, error) {
	return m.geoRules[code], nil
}

func (m *MockCache) IncrementCampaignSpend(id string, amount float64) (float64, error) {
	m.spend[id] += amount
	return m.spend[id], nil
}
func (m *MockCache) GetCampaignSpend(id string) (float64, error) {
	return m.spend[id], nil
}

func (m *MockCache) IncrementBidFormat(format string) error { return nil }
func (m *MockCache) GetBidFormats() (map[string]int64, error) {
	return make(map[string]int64), nil
}

// Generic Cache Methods
func (m *MockCache) Get(key string) (string, error) {
	return m.kv[key], nil
}
func (m *MockCache) Set(key string, value interface{}, ttl int64) error {
	if s, ok := value.(string); ok {
		m.kv[key] = s
	}
	return nil
}

// Fraud
func (m *MockCache) IncrementPublisherFraud(publisherID string) error {
	return nil
}

// Request Deduplication
func (m *MockCache) IsRequestDuplicate(requestID string, ttlSeconds int) (bool, error) {
	return false, nil
}

// Frequency Capping
func (m *MockCache) IncrementUserFrequency(userID, campaignID string, windowSecs int) (int64, error) {
	key := userID + ":" + campaignID
	m.frequencies[key]++
	return m.frequencies[key], nil
}
func (m *MockCache) GetUserFrequency(userID, campaignID string) (int64, error) {
	return m.frequencies[userID+":"+campaignID], nil
}

// Campaign performance metrics
func (m *MockCache) GetCampaignCTR(campaignID string) (float64, error) {
	return m.ctr[campaignID], nil
}
func (m *MockCache) GetCampaignWinRate(campaignID string) (float64, error) {
	return m.winRate[campaignID], nil
}
func (m *MockCache) IncrementCampaignClicks(campaignID string) error      { return nil }
func (m *MockCache) IncrementCampaignImpressions(campaignID string) error { return nil }
func (m *MockCache) IncrementCampaignBids(campaignID string) error        { return nil }
func (m *MockCache) IncrementCampaignWins(campaignID string) error        { return nil }

// Bid Landscape Analytics
func (m *MockCache) RecordBidInBucket(priceBucket string) error { return nil }
func (m *MockCache) RecordWinInBucket(priceBucket string) error { return nil }
func (m *MockCache) GetBidLandscape() (map[string]map[string]int64, error) {
	return make(map[string]map[string]int64), nil
}

// Segment-Level Performance Tracking
func (m *MockCache) IncrementSegmentImpressions(segmentType, segmentValue string) error { return nil }
func (m *MockCache) IncrementSegmentClicks(segmentType, segmentValue string) error      { return nil }
func (m *MockCache) GetSegmentPerformance(segmentType string) (map[string]map[string]int64, error) {
	return make(map[string]map[string]int64), nil
}

// Dynamic Bid Floor Optimization
func (m *MockCache) RecordPublisherBidAttempt(publisherID string, bidPrice float64, won bool) error {
	return nil
}
func (m *MockCache) GetOptimalBidFloor(publisherID string, targetWinRate float64) (float64, error) {
	return 0, nil
}

// Conversion Attribution
func (m *MockCache) RecordImpression(userID, campaignID, requestID string, ttlHours int) error {
	return nil
}
func (m *MockCache) RecordClick(userID, campaignID, requestID string, ttlHours int) error {
	return nil
}
func (m *MockCache) GetAttribution(userID, campaignID string) (string, string, error) {
	return "", "", nil
}

// Multi-Touch Attribution
func (m *MockCache) RecordTouchpoint(userID, campaignID, touchpointType, requestID string, ttlDays int) error {
	key := userID + ":" + campaignID
	tp := model.Touchpoint{
		Type:       touchpointType,
		RequestID:  requestID,
		CampaignID: campaignID,
		Timestamp:  time.Now(),
	}
	m.touchpoints[key] = append(m.touchpoints[key], tp)
	return nil
}
func (m *MockCache) GetTouchpoints(userID, campaignID string) ([]model.Touchpoint, error) {
	key := userID + ":" + campaignID
	return m.touchpoints[key], nil
}
func (m *MockCache) GetMultiTouchAttribution(userID, campaignID, modelType string) ([]model.AttributionCredit, error) {
	return nil, nil
}

// Retargeting Segments
func (m *MockCache) RecordUserEvent(userID, campaignID, eventType string, ttlDays int) error {
	key := userID + ":" + campaignID + ":" + eventType
	m.userEvents[key] = true
	return nil
}
func (m *MockCache) GetUserEvents(userID string, eventTypes []string) (map[string][]string, error) {
	return make(map[string][]string), nil
}
func (m *MockCache) HasUserEvent(userID, campaignID, eventType string) (bool, error) {
	key := userID + ":" + campaignID + ":" + eventType
	return m.userEvents[key], nil
}

// Cross-Device Graph
func (m *MockCache) LinkDevices(primaryUserID string, deviceIDs []string, ttlDays int) error {
	if m.linkedDevices == nil {
		m.linkedDevices = make(map[string][]string)
	}
	if m.primaryUserID == nil {
		m.primaryUserID = make(map[string]string)
	}
	m.linkedDevices[primaryUserID] = append(m.linkedDevices[primaryUserID], deviceIDs...)
	for _, deviceID := range deviceIDs {
		m.primaryUserID[deviceID] = primaryUserID
	}
	return nil
}
func (m *MockCache) GetLinkedDevices(primaryUserID string) ([]string, error) {
	if m.linkedDevices == nil {
		return nil, nil
	}
	return m.linkedDevices[primaryUserID], nil
}
func (m *MockCache) GetPrimaryUserID(deviceID string) (string, error) {
	if m.primaryUserID == nil {
		return "", nil
	}
	return m.primaryUserID[deviceID], nil
}
func (m *MockCache) GetCrossDeviceFrequency(primaryUserID, campaignID string) (int64, error) {
	if m.crossDeviceFreq == nil {
		return 0, nil
	}
	key := primaryUserID + ":" + campaignID
	return m.crossDeviceFreq[key], nil
}

// Supply Path Optimization Analytics
func (m *MockCache) StoreBidPathAnalytics(analytics *model.BidPathAnalytics) error { return nil }
func (m *MockCache) GetBidPathAnalytics(requestID string) (*model.BidPathAnalytics, error) {
	return nil, nil
}
func (m *MockCache) GetSupplyChainMetrics(timeRange string) (*model.SupplyChainMetrics, error) {
	return nil, nil
}
func (m *MockCache) GetServiceMetrics(serviceName string, timeRange string) (*model.ServiceMetrics, error) {
	return nil, nil
}

// SetTouchpoints is a test helper to inject touchpoints into the mock
func (m *MockCache) SetTouchpoints(userID, campaignID string, tps []model.Touchpoint) {
	key := userID + ":" + campaignID
	m.touchpoints[key] = tps
}

// SetKV is a test helper to set arbitrary key-value pairs
func (m *MockCache) SetKV(key, value string) {
	m.kv[key] = value
}

// SetUserEvent is a test helper to set a user event
func (m *MockCache) SetUserEvent(userID, campaignID, eventType string) {
	key := userID + ":" + campaignID + ":" + eventType
	m.userEvents[key] = true
}

// Helper to create a mock AI server
func createMockAIServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Just return empty recommendations typically
		json.NewEncoder(w).Encode(model.AIMatchResponse{
			Recommendations: []model.AIAdRecommendation{},
		})
	}))
}

// Compile check
var _ cache.Cache = (*MockCache)(nil)

func TestProcessBid_FraudCheck(t *testing.T) {
	mockCache := NewMockCache()
	service := NewBiddingService(mockCache, "http://backend")

	// Mock Fraud Service
	fraudServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req model.FraudCheckRequest
		json.NewDecoder(r.Body).Decode(&req)

		isFraud := req.IPAddress == "1.2.3.4" // Flag this specific IP

		resp := model.FraudCheckResponse{
			RequestID: req.RequestID,
			IsFraud:   isFraud,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer fraudServer.Close()

	// Mock AI Service (needed because it's initialized in NewBiddingService)
	aiServer := createMockAIServer()
	defer aiServer.Close()
	service.SetAIServiceURL(aiServer.URL)

	service.SetFraudServiceURL(fraudServer.URL)

	campaign := &model.Campaign{
		ID:        "camp-1",
		BidPrice:  1.0,
		Targeting: model.Targeting{Countries: []string{"US"}},
		Status:    "active",
		Budget:    1000,
		Creative: model.Creative{
			Type:   "banner",
			URL:    "http://ads.com/banner.jpg",
			Width:  300,
			Height: 250,
		},
	}
	mockCache.SetActiveCampaigns([]*model.Campaign{campaign})

	// Case 1: Legitimate User
	reqLegit := &model.BidRequest{
		ID:     "req-legit",
		Device: model.InternalDevice{IP: "5.6.7.8", Type: "mobile"},
		User:   model.InternalUser{Country: "US"},
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
	}

	resp, err := service.ProcessBid(reqLegit)
	if err != nil {
		t.Errorf("Expected bid for legit user, got error: %v", err)
	}
	if resp == nil {
		t.Error("Expected bid response for legit user")
	}

	// Case 2: Fraudulent User
	reqFraud := &model.BidRequest{
		ID:     "req-fraud",
		Device: model.InternalDevice{IP: "1.2.3.4", Type: "mobile"},
		User:   model.InternalUser{Country: "US"},
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
	}

	respFraud, errFraud := service.ProcessBid(reqFraud)
	if errFraud == nil {
		t.Error("Expected error for fraud user, got nil")
	} else if errFraud.Error() != "request flagged as fraud" {
		t.Errorf("Expected fraud error, got: %v", errFraud)
	}
	if respFraud != nil {
		t.Error("Expected no bid response for fraud user")
	}
}

func TestProcessBid_BidOptimization(t *testing.T) {
	mockCache := NewMockCache()
	service := NewBiddingService(mockCache, "http://backend")

	// Mock Optimization Service
	optServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := model.BidRecommendation{
			RecommendedBid: 2.50,
			BidMultiplier:  2.5,
			Reasoning:      []string{"high_value_user"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer optServer.Close()

	// Mock other services
	aiServer := createMockAIServer()
	defer aiServer.Close()
	service.SetAIServiceURL(aiServer.URL)

	// Dummy fraud service (always allow)
	fraudServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(model.FraudCheckResponse{IsFraud: false})
	}))
	defer fraudServer.Close()
	service.SetFraudServiceURL(fraudServer.URL)

	service.SetOptimizationServiceURL(optServer.URL)

	baseBid := 1.0
	campaign := &model.Campaign{
		ID:        "camp-opt",
		BidPrice:  baseBid,
		Targeting: model.Targeting{Countries: []string{"US"}},
		Status:    "active",
		Budget:    1000,
		Creative:  model.Creative{Type: "banner"},
	}
	mockCache.SetActiveCampaigns([]*model.Campaign{campaign})

	req := &model.BidRequest{
		ID:     "req-opt",
		Device: model.InternalDevice{Type: "mobile", IP: "10.0.0.1"},
		User:   model.InternalUser{Country: "US"},
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
	}

	resp, err := service.ProcessBid(req)
	if err != nil {
		t.Fatalf("ProcessBid failed: %v", err)
	}

	if resp.BidPrice != 2.50 {
		t.Errorf("Expected optimized bid 2.50, got %f", resp.BidPrice)
	}
}

func TestProcessBid_BudgetEnforcement(t *testing.T) {
	// Setup
	mockCache := NewMockCache()
	service := NewBiddingService(mockCache, "http://backend")

	campaign := &model.Campaign{
		ID:       "camp-1",
		BidPrice: 1.0,
		Budget:   100.0, // Daily budget implied as 10% -> 10.0
		Targeting: model.Targeting{
			Countries: []string{"US"},
			Devices:   []string{"mobile"},
		},
		Status:   "active",
		Creative: model.Creative{Type: "banner"},
	}
	mockCache.SetActiveCampaigns([]*model.Campaign{campaign})

	req := &model.BidRequest{
		ID: "req-1",
		Device: model.InternalDevice{
			Type: "mobile",
			OS:   "android",
		},
		User: model.InternalUser{
			Country: "US",
		},
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
	}

	// 1. Normal Bid
	resp, err := service.ProcessBid(req)
	if err != nil {
		t.Fatalf("Expected bid, got error: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected bid response, got nil")
	}

	// 2. Exceed Budget
	// Daily budget is 10% of 100 = 10.0
	// Set spend to 11.0
	mockCache.spend["camp-1"] = 11.0

	// Now it should fail (return no matching campaigns error)
	_, err = service.ProcessBid(req)
	if err == nil {
		t.Fatal("Expected error due to budget exceeded, got nil")
	}
}

func TestProcessBid_UserSegments(t *testing.T) {
	mockCache := NewMockCache()
	service := NewBiddingService(mockCache, "http://backend")

	// Campaign A: Lower bid ($10), but matches User Segment (gets 10% boost -> score 11.0)
	campaignA := &model.Campaign{
		ID:       "camp-seg",
		BidPrice: 10.0,
		Budget:   1000.0,
		Targeting: model.Targeting{
			Countries:  []string{"US"},
			Categories: []string{"luxury"},
		},
		Creative: model.Creative{Type: "banner"},
	}

	// Campaign B: Higher bid ($10.5), no segment match -> score 10.5
	campaignB := &model.Campaign{
		ID:       "camp-generic",
		BidPrice: 10.5,
		Budget:   1000.0,
		Targeting: model.Targeting{
			Countries: []string{"US"},
		},
		Creative: model.Creative{Type: "banner"},
	}

	mockCache.SetActiveCampaigns([]*model.Campaign{campaignA, campaignB})

	userID := "user-vip"
	mockCache.SetUserSegments(userID, []string{"luxury"})

	req := &model.BidRequest{
		ID: "req-vip",
		User: model.InternalUser{
			ID:      userID,
			Country: "US",
			// Must include "luxury" to pass hard filter if we want to test segment boost on top
			Categories: []string{"luxury"},
		},
		Device: model.InternalDevice{Type: "mobile"},
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
	}

	resp, err := service.ProcessBid(req)
	if err != nil {
		t.Fatalf("Expected bid, got error: %v", err)
	}

	// Expect Campaign A to win due to boost (11.0 > 10.5)
	if resp.CampaignID != "camp-seg" {
		t.Errorf("Expected Campaign A (camp-seg) to win due to segment boost, but got %s (BidPrice: %f)", resp.CampaignID, resp.BidPrice)
	}
}

func TestProcessBid_GeoBlocked(t *testing.T) {
	mockCache := NewMockCache()
	service := NewBiddingService(mockCache, "http://backend")

	campaign := &model.Campaign{
		ID:       "camp-geo",
		BidPrice: 5.0,
		Budget:   1000.0,
		Targeting: model.Targeting{
			Countries: []string{"US"}, // Only US
		},
		Creative: model.Creative{Type: "banner"},
	}
	mockCache.SetActiveCampaigns([]*model.Campaign{campaign})

	req := &model.BidRequest{
		ID: "req-geo",
		User: model.InternalUser{
			Country: "CA", // User from Canada
		},
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
	}

	_, err := service.ProcessBid(req)
	if err == nil {
		t.Error("Expected error (no matching campaigns) due to geo-blocking, but got a bid")
	}
}

func TestProcessBid_AIScoring(t *testing.T) {
	// 1. Setup Mock AI Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request content
		var aiReq model.AIMatchRequest
		json.NewDecoder(r.Body).Decode(&aiReq)

		// Return specific recommendation
		resp := model.AIMatchResponse{
			Recommendations: []model.AIAdRecommendation{
				{
					CampaignID:   "camp-ai",
					OverallScore: 0.5, // 50% boost
					BidPrice:     2.0,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// 2. Setup Service
	mockCache := NewMockCache()
	service := NewBiddingService(mockCache, "http://backend")
	service.SetAIServiceURL(server.URL) // Use mock server

	// Campaign A: No AI Boost
	camp1 := &model.Campaign{
		ID:        "camp-generic",
		BidPrice:  10.0,
		Budget:    1000.0,
		Targeting: model.Targeting{Countries: []string{"US"}},
		Creative:  model.Creative{Type: "banner"},
	}
	// Campaign B: Will get AI Boost
	camp2 := &model.Campaign{
		ID:        "camp-ai",
		BidPrice:  8.0, // Lower base bid/score
		Budget:    1000.0,
		Targeting: model.Targeting{Countries: []string{"US"}},
		Creative:  model.Creative{Type: "banner"},
	}

	mockCache.SetActiveCampaigns([]*model.Campaign{camp1, camp2})

	req := &model.BidRequest{
		ID:     "req-ai-test",
		User:   model.InternalUser{Country: "US"},
		Device: model.InternalDevice{Type: "mobile"},
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
	}

	// 3. Execution
	resp, err := service.ProcessBid(req)

	// 4. Verification
	if err != nil {
		t.Fatalf("ProcessBid failed: %v", err)
	}

	// Without AI: camp-generic (10.0) > camp-ai (8.0)
	// With AI: camp-ai (8.0 * (1+0.5) = 12.0) > camp-generic (10.0)
	if resp.CampaignID != "camp-ai" {
		t.Errorf("Expected 'camp-ai' to win due to ML boost. Winner was: %s", resp.CampaignID)
	}
}

func TestProcessBid_VideoVAST(t *testing.T) {
	mockCache := NewMockCache()
	service := NewBiddingService(mockCache, "http://backend-url")

	// Create Video Campaign
	videoCamp := &model.Campaign{
		ID:        "camp-video",
		BidPrice:  5.0,
		Budget:    1000,
		Targeting: model.Targeting{Countries: []string{"US"}},
		Creative: model.Creative{
			Type:     "video",
			URL:      "http://ads.com/video.mp4",
			Duration: 30,
			MimeType: "video/mp4",
			Width:    640,
			Height:   480,
		},
		Status: "active",
	}

	mockCache.SetActiveCampaigns([]*model.Campaign{videoCamp})

	// Setup Mock Services (Fraud/AI/Opt) to pass through
	setupMockServices(service)

	req := &model.BidRequest{
		ID:     "req-video-1",
		User:   model.InternalUser{Country: "US"},
		Device: model.InternalDevice{Type: "mobile", IP: "10.0.0.1"},
		AdSlot: model.AdSlot{Formats: []string{"video"}},
	}

	resp, err := service.ProcessBid(req)
	if err != nil {
		t.Fatalf("ProcessBid failed: %v", err)
	}

	if resp.AdMarkup == "" {
		t.Error("Expected VAST AdMarkup for video campaign, got empty string")
	}

	if resp.CreativeURL != "http://ads.com/video.mp4" {
		t.Errorf("Expected CreativeURL http://ads.com/video.mp4, got %s", resp.CreativeURL)
	}
}

func TestProcessBid_Native(t *testing.T) {
	mockCache := NewMockCache()
	service := NewBiddingService(mockCache, "http://backend-url")

	// Create Native Campaign
	nativeCamp := &model.Campaign{
		ID:        "camp-native",
		BidPrice:  3.5,
		Budget:    1000,
		Targeting: model.Targeting{Countries: []string{"US"}},
		Creative: model.Creative{
			Type:        "native",
			Title:       "Native Title",
			Description: "Native Description",
			URL:         "http://ads.com/main.jpg",
			IconURL:     "http://ads.com/icon.png",
			CTAText:     "Install Now",
			Width:       1200,
			Height:      627,
		},
		Status: "active",
	}

	mockCache.SetActiveCampaigns([]*model.Campaign{nativeCamp})
	setupMockServices(service)

	req := &model.BidRequest{
		ID:     "req-native-1",
		User:   model.InternalUser{Country: "US"},
		Device: model.InternalDevice{Type: "mobile", IP: "10.0.0.1"},
		AdSlot: model.AdSlot{Formats: []string{"native"}},
	}

	resp, err := service.ProcessBid(req)
	if err != nil {
		t.Fatalf("ProcessBid failed: %v", err)
	}

	if resp.AdMarkup == "" {
		t.Error("Expected Native JSON AdMarkup, got empty string")
	}

	// Simple check for native structure
	var markup map[string]interface{}
	if err := json.Unmarshal([]byte(resp.AdMarkup), &markup); err != nil {
		t.Errorf("Failed to parse Native JSON: %v", err)
	}

	nativeData, ok := markup["native"].(map[string]interface{})
	if !ok {
		t.Error("Missing 'native' key in markup")
	}

	assets, ok := nativeData["assets"].([]interface{})
	if !ok || len(assets) < 5 {
		t.Error("Expected at least 5 assets in native ad")
	}
}

// Helper to setup mock services efficiently
func setupMockServices(s *BiddingService) {
	// Fraud: always safe
	fraud := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(model.FraudCheckResponse{IsFraud: false})
	}))
	s.SetFraudServiceURL(fraud.URL)

	// AI: no specific boost
	ai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(model.AIMatchResponse{})
	}))
	s.SetAIServiceURL(ai.URL)

	// Opt: no change
	opt := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(model.BidRecommendation{RecommendedBid: 0}) // 0 means no opt
	}))
	s.SetOptimizationServiceURL(opt.URL)
}
