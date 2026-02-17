package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// MockCache implements cache.Cache for testing
type MockCache struct {
	campaigns    []*model.Campaign
	userSegments map[string][]string
	geoRules     map[string]map[string]interface{}
	spend        map[string]float64
}

func NewMockCache() *MockCache {
	return &MockCache{
		userSegments: make(map[string][]string),
		geoRules:     make(map[string]map[string]interface{}),
		spend:        make(map[string]float64),
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
		Device: model.Device{IP: "5.6.7.8", Type: "mobile"},
		User:   model.User{Country: "US"},
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
		Device: model.Device{IP: "1.2.3.4", Type: "mobile"},
		User:   model.User{Country: "US"},
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
	}
	mockCache.SetActiveCampaigns([]*model.Campaign{campaign})

	req := &model.BidRequest{
		ID:     "req-opt",
		Device: model.Device{Type: "mobile", IP: "10.0.0.1"},
		User:   model.User{Country: "US"},
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
		Status: "active",
	}
	mockCache.SetActiveCampaigns([]*model.Campaign{campaign})

	req := &model.BidRequest{
		ID: "req-1",
		Device: model.Device{
			Type: "mobile",
			OS:   "android",
		},
		User: model.User{
			Country: "US",
		},
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
	}

	// Campaign B: Higher bid ($10.5), no segment match -> score 10.5
	campaignB := &model.Campaign{
		ID:       "camp-generic",
		BidPrice: 10.5,
		Budget:   1000.0,
		Targeting: model.Targeting{
			Countries: []string{"US"},
		},
	}

	mockCache.SetActiveCampaigns([]*model.Campaign{campaignA, campaignB})

	userID := "user-vip"
	mockCache.SetUserSegments(userID, []string{"luxury"})

	req := &model.BidRequest{
		ID: "req-vip",
		User: model.User{
			ID:      userID,
			Country: "US",
			// Must include "luxury" to pass hard filter if we want to test segment boost on top
			Categories: []string{"luxury"},
		},
		Device: model.Device{Type: "mobile"},
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
	}
	mockCache.SetActiveCampaigns([]*model.Campaign{campaign})

	req := &model.BidRequest{
		ID: "req-geo",
		User: model.User{
			Country: "CA", // User from Canada
		},
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
	}
	// Campaign B: Will get AI Boost
	camp2 := &model.Campaign{
		ID:        "camp-ai",
		BidPrice:  8.0, // Lower base bid/score
		Budget:    1000.0,
		Targeting: model.Targeting{Countries: []string{"US"}},
	}

	mockCache.SetActiveCampaigns([]*model.Campaign{camp1, camp2})

	req := &model.BidRequest{
		ID:     "req-ai-test",
		User:   model.User{Country: "US"},
		Device: model.Device{Type: "mobile"},
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
		User:   model.User{Country: "US"},
		Device: model.Device{Type: "mobile", IP: "10.0.0.1"},
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
		User:   model.User{Country: "US"},
		Device: model.Device{Type: "mobile", IP: "10.0.0.1"},
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
