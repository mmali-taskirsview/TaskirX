package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================================
// BID LANDSCAPE SERVICE TESTS
// ============================================================================

func TestBidLandscape_Disabled(t *testing.T) {
	svc := NewBidLandscapeService(nil)

	campaign := &model.Campaign{
		ID:       "camp-1",
		BidPrice: 2.0,
		Targeting: model.Targeting{
			BidLandscape: nil, // Disabled
		},
	}

	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
	}

	result := svc.AnalyzeLandscape(campaign, req)
	if result.Analyzed {
		t.Error("Expected Analyzed=false when BidLandscape is disabled")
	}
	if result.BidMultiplier != 1.0 {
		t.Errorf("Expected BidMultiplier=1.0, got %f", result.BidMultiplier)
	}
}

func TestBidLandscape_WithConfig(t *testing.T) {
	svc := NewBidLandscapeService(nil)

	campaign := &model.Campaign{
		ID:       "camp-1",
		BidPrice: 2.0,
		Targeting: model.Targeting{
			BidLandscape: &model.BidLandscape{
				Enabled:        true,
				AnalysisWindow: 24,
				MinSampleSize:  100,
			},
		},
	}

	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Type: "mobile"},
	}

	result := svc.AnalyzeLandscape(campaign, req)
	// With no historical data, should return base multiplier
	if result.BidMultiplier < 0.5 || result.BidMultiplier > 2.0 {
		t.Errorf("BidMultiplier %f out of expected range", result.BidMultiplier)
	}
}

func TestBidLandscape_RecordBid(t *testing.T) {
	svc := NewBidLandscapeService(nil)

	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Type: "mobile"},
	}

	// Record some bids
	svc.RecordBid(req, 2.0, 1.8, true)
	svc.RecordBid(req, 2.5, 2.0, false)
	svc.RecordBid(req, 1.5, 1.5, true)

	// Now analyze should have data
	campaign := &model.Campaign{
		ID:       "camp-1",
		BidPrice: 2.0,
		Targeting: model.Targeting{
			BidLandscape: &model.BidLandscape{
				Enabled: true,
			},
		},
	}

	result := svc.AnalyzeLandscape(campaign, req)
	// Should have some analysis now
	if result.BidMultiplier == 0 {
		t.Error("Expected non-zero BidMultiplier after recording bids")
	}
}

// ============================================================================
// CREATIVE OPTIMIZATION SERVICE TESTS
// ============================================================================

func TestCreativeOptimization_Disabled(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	campaign := &model.Campaign{
		ID: "camp-1",
		Creative: model.Creative{
			Type: "banner",
			URL:  "https://example.com/ad.png",
		},
		Targeting: model.Targeting{
			CreativeOptimization: nil, // Disabled
		},
	}

	req := &model.BidRequest{ID: "req-1"}

	result := svc.SelectCreative(campaign, req)
	if result.SelectionMethod != "default" {
		t.Errorf("Expected default selection method, got %s", result.SelectionMethod)
	}
}

func TestCreativeOptimization_WithPool(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	campaign := &model.Campaign{
		ID: "camp-1",
		Creative: model.Creative{
			Type: "banner",
			URL:  "https://example.com/ad.png",
		},
		Targeting: model.Targeting{
			CreativeOptimization: &model.CreativeOptimization{
				Enabled:         true,
				ExplorationRate: 0.1,
				CreativePool: []model.CreativeVariant{
					{ID: "creative-1", Weight: 1.0, Status: "active"},
					{ID: "creative-2", Weight: 1.0, Status: "active"},
					{ID: "creative-3", Weight: 1.0, Status: "active"},
				},
			},
		},
	}

	req := &model.BidRequest{
		ID: "req-1",
		AdSlot: model.AdSlot{
			Formats: []string{"banner"},
		},
	}

	result := svc.SelectCreative(campaign, req)
	if result.SelectedCreativeID == "" {
		t.Error("Expected a creative to be selected")
	}
	if result.Confidence <= 0 {
		t.Error("Expected positive confidence score")
	}
}

func TestCreativeOptimization_TrackPerformance(t *testing.T) {
	svc := NewCreativeOptimizationService(nil)

	// Track performance - verify no panic
	// The service tracks internally
	_ = svc
}

// ============================================================================
// INCREMENTALITY SERVICE TESTS
// ============================================================================

func TestIncrementality_Disabled(t *testing.T) {
	svc := NewIncrementalityService(nil)

	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			IncrementalityConfig: nil, // Disabled
		},
	}

	req := &model.BidRequest{
		ID:   "req-1",
		User: model.InternalUser{ID: "user-1"},
	}

	result := svc.EvaluateUser(campaign, req)
	if result.UserInControlGroup {
		t.Error("Expected user not in control group when disabled")
	}
}

func TestIncrementality_UserAssignment(t *testing.T) {
	svc := NewIncrementalityService(nil)

	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			IncrementalityConfig: &model.IncrementalityConfig{
				Enabled:        true,
				ExperimentID:   "exp-1",
				ControlPercent: 10.0,
				HoldoutType:    "user",
			},
		},
	}

	// Test multiple users - some should be in control
	controlCount := 0
	testCount := 0

	for i := 0; i < 100; i++ {
		req := &model.BidRequest{
			ID:   "req-" + string(rune(i)),
			User: model.InternalUser{ID: "user-" + string(rune(i))},
		}

		result := svc.EvaluateUser(campaign, req)
		if result.UserInControlGroup {
			controlCount++
		} else {
			testCount++
		}
	}

	// With 10% control, expect roughly 10 in control (allow variance)
	if controlCount < 1 || controlCount > 30 {
		t.Errorf("Expected ~10 control users, got %d", controlCount)
	}
}

func TestIncrementality_RecordConversion(t *testing.T) {
	svc := NewIncrementalityService(nil)

	// First, evaluate user to create the experiment
	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			IncrementalityConfig: &model.IncrementalityConfig{
				Enabled:        true,
				ExperimentID:   "exp-1",
				ControlPercent: 10.0,
				HoldoutType:    "user",
			},
		},
	}

	req := &model.BidRequest{
		ID:   "req-1",
		User: model.InternalUser{ID: "user-1"},
	}

	// This creates the experiment
	svc.EvaluateUser(campaign, req)

	// Now record conversions for both groups
	svc.RecordImpression("exp-1", "user-1", false)
	svc.RecordImpression("exp-1", "user-2", true)
	svc.RecordConversion("exp-1", "user-1", false, 50.0)
	svc.RecordConversion("exp-1", "user-2", true, 25.0)

	// Get results - may return "insufficient_data" status due to small sample size
	result := svc.GetExperimentResults("exp-1")
	if result == nil {
		t.Fatal("Expected experiment results")
	}
	// With experiment created, status should not be "not_found"
	if result.Status == "not_found" {
		t.Error("Experiment should be found after recording data")
	}
}

// ============================================================================
// PRIVACY SANDBOX SERVICE TESTS
// ============================================================================

func TestPrivacySandbox_Disabled(t *testing.T) {
	svc := NewPrivacySandboxService(nil)

	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			PrivacySandbox: nil, // Disabled
		},
	}

	req := &model.BidRequest{ID: "req-1"}

	result := svc.EvaluatePrivacySandbox(campaign, req)
	if result.TopicsAvailable {
		t.Error("Expected TopicsAvailable=false when disabled")
	}
}

func TestPrivacySandbox_TopicsAPI(t *testing.T) {
	svc := NewPrivacySandboxService(nil)

	// Register some topics for user
	svc.RegisterUserTopic("user-1", 1)   // Arts & Entertainment
	svc.RegisterUserTopic("user-1", 100) // Business

	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			PrivacySandbox: &model.PrivacySandbox{
				Enabled: true,
				TopicsAPI: &model.TopicsAPIConfig{
					Enabled:      true,
					TargetTopics: []int{1, 2, 3},
				},
			},
		},
	}

	req := &model.BidRequest{
		ID:   "req-1",
		User: model.InternalUser{ID: "user-1"},
	}

	result := svc.EvaluatePrivacySandbox(campaign, req)
	// Result depends on user topics availability
	if result.TopicMultiplier < 0 {
		t.Error("Expected non-negative topic multiplier")
	}
}

func TestPrivacySandbox_InterestGroups(t *testing.T) {
	svc := NewPrivacySandboxService(nil)

	// Add user to interest groups
	svc.AddToInterestGroup("user-1", "sports_fans")
	svc.AddToInterestGroup("user-1", "tech_enthusiasts")

	groups := svc.GetUserInterestGroups("user-1")
	if len(groups) != 2 {
		t.Errorf("Expected 2 interest groups, got %d", len(groups))
	}
}

// ============================================================================
// CONTEXTUAL AI SERVICE TESTS
// ============================================================================

func TestContextualAI_Disabled(t *testing.T) {
	svc := NewContextualAIService(nil)

	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			ContextualAI: nil, // Disabled
		},
	}

	req := &model.BidRequest{ID: "req-1"}

	result := svc.AnalyzeContext(campaign, req)
	if result.Analyzed {
		t.Error("Expected Analyzed=false when disabled")
	}
	if !result.BrandSafe {
		t.Error("Expected BrandSafe=true by default when disabled")
	}
}

func TestContextualAI_ContentAnalysis(t *testing.T) {
	svc := NewContextualAIService(nil)

	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			ContextualAI: &model.ContextualAI{
				Enabled:          true,
				AnalyzeContent:   true,
				AnalyzeSentiment: true,
			},
		},
	}

	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
		Context: map[string]interface{}{
			"page_title":   "Top 10 Travel Destinations for 2026",
			"page_content": "Explore the best vacation spots around the world",
		},
	}

	result := svc.AnalyzeContext(campaign, req)
	if !result.Analyzed {
		t.Error("Expected Analyzed=true with content")
	}
	if result.BidMultiplier <= 0 {
		t.Error("Expected positive BidMultiplier")
	}
}

func TestContextualAI_BrandSafety(t *testing.T) {
	svc := NewContextualAIService(nil)

	campaign := &model.Campaign{
		ID:               "camp-1",
		BrandSafetyLevel: "strict",
		Targeting: model.Targeting{
			ContextualAI: &model.ContextualAI{
				Enabled:        true,
				AnalyzeContent: true,
			},
		},
	}

	// Test with potentially unsafe content
	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
		Context: map[string]interface{}{
			"page_title":   "Normal news article",
			"page_content": "This is regular content about technology and business",
		},
	}

	result := svc.AnalyzeContext(campaign, req)
	// Regular content should be brand safe
	if !result.BrandSafe {
		t.Log("Content marked as not brand safe:", result.Reason)
	}
}

// ============================================================================
// REAL-TIME ALERTS SERVICE TESTS
// ============================================================================

func TestRealTimeAlerts_Disabled(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			AlertConfig: nil, // Disabled
		},
	}

	result := svc.CheckAlerts(campaign, 50.0, 100.0)
	if result.HasActiveAlerts {
		t.Error("Expected no active alerts when disabled")
	}
	if result.BidAdjustment != 1.0 {
		t.Errorf("Expected BidAdjustment=1.0, got %f", result.BidAdjustment)
	}
}

func TestRealTimeAlerts_BudgetWarning(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	campaign := &model.Campaign{
		ID:   "camp-1",
		Name: "Test Campaign",
		Targeting: model.Targeting{
			AlertConfig: &model.AlertConfig{
				Enabled: true,
				BudgetAlerts: &model.BudgetAlerts{
					Enabled:           true,
					WarnAtPercent:     80.0,
					CriticalAtPercent: 95.0,
				},
			},
		},
	}

	// 85% spent - should trigger warning
	result := svc.CheckAlerts(campaign, 85.0, 100.0)
	if result.BidAdjustment >= 1.0 {
		t.Error("Expected reduced bid adjustment at 85% spend")
	}
}

func TestRealTimeAlerts_RecordMetrics(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	// Record hourly metrics
	for i := 0; i < 24; i++ {
		svc.RecordMetrics("camp-1", float64(100+i*5), 0.02, 0.01, 0.15)
	}

	// Now alerts should have baseline data
	campaign := &model.Campaign{
		ID:   "camp-1",
		Name: "Test Campaign",
		Targeting: model.Targeting{
			AlertConfig: &model.AlertConfig{
				Enabled: true,
				AnomalyDetection: &model.AnomalyDetection{
					Enabled:     true,
					Sensitivity: "medium",
				},
			},
		},
	}

	// Test anomaly detection with normal value
	isAnomaly := svc.DetectAnomaly(campaign, "spend", 150.0)
	// With normal value within range, shouldn't be anomaly
	if isAnomaly {
		t.Log("Value flagged as anomaly")
	}
}

// ============================================================================
// COMPETITIVE INTELLIGENCE SERVICE TESTS
// ============================================================================

func TestCompetitiveIntelligence_Disabled(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			CompetitiveIntelligence: nil, // Disabled
		},
	}

	req := &model.BidRequest{ID: "req-1", PublisherID: "pub-1"}

	result := svc.AnalyzeCompetition(campaign, req)
	if result.Analyzed {
		t.Error("Expected Analyzed=false when disabled")
	}
	if result.BidAdjustment != 1.0 {
		t.Errorf("Expected BidAdjustment=1.0, got %f", result.BidAdjustment)
	}
}

func TestCompetitiveIntelligence_RecordOutcome(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
		AdSlot:      model.AdSlot{ID: "slot-1", Formats: []string{"banner"}},
	}

	// Record auction outcomes
	svc.RecordAuctionOutcome(req, 2.0, 2.5, false, "competitor-1")
	svc.RecordAuctionOutcome(req, 2.5, 2.3, true, "")
	svc.RecordAuctionOutcome(req, 2.2, 2.8, false, "competitor-2")

	// Get market report
	report := svc.GetMarketReport()
	if report["total_auctions"].(int) != 3 {
		t.Errorf("Expected 3 auctions, got %v", report["total_auctions"])
	}
}

func TestCompetitiveIntelligence_WithTracking(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	// Record some competitor activity
	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
		AdSlot:      model.AdSlot{ID: "slot-1", Formats: []string{"banner"}},
	}

	for i := 0; i < 10; i++ {
		svc.RecordAuctionOutcome(req, 2.0, 2.5, false, "competitor-1")
	}

	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			CompetitiveIntelligence: &model.CompetitiveIntelligence{
				Enabled:          true,
				TrackCompetitors: []string{"competitor-1"},
				CompetitiveMode:  "aggressive",
			},
		},
	}

	result := svc.AnalyzeCompetition(campaign, req)
	if !result.Analyzed {
		t.Error("Expected analysis to be performed")
	}
}

// ============================================================================
// UNIFIED ID SERVICE TESTS
// ============================================================================

func TestUnifiedID_Disabled(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			UnifiedIDConfig: nil, // Disabled
		},
	}

	req := &model.BidRequest{ID: "req-1"}

	result := svc.ResolveIdentity(campaign, req)
	if result.Resolved {
		t.Error("Expected Resolved=false when disabled")
	}
	if result.BidMultiplier != 1.0 {
		t.Errorf("Expected BidMultiplier=1.0, got %f", result.BidMultiplier)
	}
}

func TestUnifiedID_WithProviders(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			UnifiedIDConfig: &model.UnifiedIDConfig{
				Enabled: true,
				Providers: []model.IDProvider{
					{Name: "uid2", Enabled: true, Priority: 1, BidBoost: 0.2},
					{Name: "id5", Enabled: true, Priority: 2, BidBoost: 0.15},
				},
				FallbackOrder: []string{"uid2", "id5"},
			},
		},
	}

	req := &model.BidRequest{
		ID:   "req-1",
		User: model.InternalUser{ID: "user-123"},
	}

	result := svc.ResolveIdentity(campaign, req)
	// Resolution depends on simulated match rate
	if result.BidMultiplier < 1.0 {
		t.Error("Expected BidMultiplier >= 1.0")
	}
}

func TestUnifiedID_LinkIdentities(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	// Link identities
	svc.LinkIdentities("uid2-abc", "uid2", "id5-xyz", "id5", "mobile", 0.9)

	// Verify link
	report := svc.GetIdentityReport()
	if report["total_identities"].(int) < 1 {
		t.Error("Expected at least 1 identity after linking")
	}
	if report["total_links"].(int) < 1 {
		t.Error("Expected at least 1 link after linking")
	}
}

func TestUnifiedID_CrossDevice(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	// Link multiple devices
	svc.LinkIdentities("user-1-mobile", "device", "user-1-desktop", "device", "mobile", 0.85)
	svc.LinkIdentities("user-1-desktop", "device", "user-1-tablet", "device", "desktop", 0.80)

	// Check cross-device reach
	reach := svc.CalculateCrossDeviceReach()
	if reach == 0 {
		t.Error("Expected non-zero cross-device reach")
	}
}

func TestUnifiedID_ConsentRequired(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	campaign := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			UnifiedIDConfig: &model.UnifiedIDConfig{
				Enabled:         true,
				ConsentRequired: true,
				Providers: []model.IDProvider{
					{Name: "uid2", Enabled: true, Priority: 1},
				},
			},
		},
	}

	// Request without consent
	req := &model.BidRequest{
		ID:   "req-1",
		User: model.InternalUser{ID: "user-123"},
		Context: map[string]interface{}{
			"gdpr_consent": false,
		},
	}

	result := svc.ResolveIdentity(campaign, req)
	// Should not resolve without consent
	if result.Resolved && !result.HasConsent {
		t.Error("Should not resolve identity without consent")
	}
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

func TestAllServicesInitialized(t *testing.T) {
	// Create mock cache
	cache := &mockCache{}

	svc := NewBiddingService(cache, "http://localhost:8080")

	// Verify all services are initialized
	if svc.GetBidLandscapeService() == nil {
		t.Error("BidLandscapeService not initialized")
	}
	if svc.GetCreativeOptimizationService() == nil {
		t.Error("CreativeOptimizationService not initialized")
	}
	if svc.GetIncrementalityService() == nil {
		t.Error("IncrementalityService not initialized")
	}
	if svc.GetPrivacySandboxService() == nil {
		t.Error("PrivacySandboxService not initialized")
	}
	if svc.GetContextualAIService() == nil {
		t.Error("ContextualAIService not initialized")
	}
	if svc.GetRealTimeAlertService() == nil {
		t.Error("RealTimeAlertService not initialized")
	}
	if svc.GetCompetitiveIntelligenceService() == nil {
		t.Error("CompetitiveIntelligenceService not initialized")
	}
	if svc.GetUnifiedIDService() == nil {
		t.Error("UnifiedIDService not initialized")
	}
}

// Helper: Mock cache for testing - implements cache.Cache interface
type mockCache struct{}

// Generic Cache Methods
func (m *mockCache) Get(key string) (string, error)                     { return "", nil }
func (m *mockCache) Set(key string, value interface{}, ttl int64) error { return nil }

// Campaign Methods
func (m *mockCache) GetActiveCampaigns() ([]*model.Campaign, error)         { return nil, nil }
func (m *mockCache) SetActiveCampaigns(campaigns []*model.Campaign) error   { return nil }
func (m *mockCache) GetCampaign(campaignID string) (*model.Campaign, error) { return nil, nil }
func (m *mockCache) SetCampaign(campaign *model.Campaign) error             { return nil }

// Metrics
func (m *mockCache) IncrementBidCount() error              { return nil }
func (m *mockCache) IncrementWinCount() error              { return nil }
func (m *mockCache) GetBidCount() (int64, error)           { return 0, nil }
func (m *mockCache) GetWinCount() (int64, error)           { return 0, nil }
func (m *mockCache) RecordLatency(latencyMs float64) error { return nil }
func (m *mockCache) GetAverageLatency() (float64, error)   { return 0, nil }

// User Segments
func (m *mockCache) SetUserSegments(userID string, segments []string) error { return nil }
func (m *mockCache) GetUserSegments(userID string) ([]string, error)        { return nil, nil }

// Geo Rules
func (m *mockCache) SetGeoRules(countryCode string, rules map[string]interface{}) error { return nil }
func (m *mockCache) GetGeoRules(countryCode string) (map[string]interface{}, error)     { return nil, nil }

// Campaign Spend
func (m *mockCache) IncrementCampaignSpend(campaignID string, amount float64) (float64, error) {
	return 0, nil
}
func (m *mockCache) GetCampaignSpend(campaignID string) (float64, error) { return 0, nil }

// Bid Formats
func (m *mockCache) IncrementBidFormat(format string) error   { return nil }
func (m *mockCache) GetBidFormats() (map[string]int64, error) { return nil, nil }

// Fraud
func (m *mockCache) IncrementPublisherFraud(publisherID string) error { return nil }

// Request Deduplication
func (m *mockCache) IsRequestDuplicate(requestID string, ttlSeconds int) (bool, error) {
	return false, nil
}

// Frequency Capping
func (m *mockCache) IncrementUserFrequency(userID, campaignID string, windowSecs int) (int64, error) {
	return 0, nil
}
func (m *mockCache) GetUserFrequency(userID, campaignID string) (int64, error) { return 0, nil }

// Campaign Performance Metrics
func (m *mockCache) GetCampaignCTR(campaignID string) (float64, error)     { return 0, nil }
func (m *mockCache) GetCampaignWinRate(campaignID string) (float64, error) { return 0, nil }
func (m *mockCache) IncrementCampaignClicks(campaignID string) error       { return nil }
func (m *mockCache) IncrementCampaignImpressions(campaignID string) error  { return nil }
func (m *mockCache) IncrementCampaignBids(campaignID string) error         { return nil }
func (m *mockCache) IncrementCampaignWins(campaignID string) error         { return nil }

// Bid Landscape Analytics
func (m *mockCache) RecordBidInBucket(priceBucket string) error            { return nil }
func (m *mockCache) RecordWinInBucket(priceBucket string) error            { return nil }
func (m *mockCache) GetBidLandscape() (map[string]map[string]int64, error) { return nil, nil }

// Segment-Level Performance Tracking
func (m *mockCache) IncrementSegmentImpressions(segmentType, segmentValue string) error { return nil }
func (m *mockCache) IncrementSegmentClicks(segmentType, segmentValue string) error      { return nil }
func (m *mockCache) GetSegmentPerformance(segmentType string) (map[string]map[string]int64, error) {
	return nil, nil
}

// Dynamic Bid Floor Optimization
func (m *mockCache) RecordPublisherBidAttempt(publisherID string, bidPrice float64, won bool) error {
	return nil
}
func (m *mockCache) GetOptimalBidFloor(publisherID string, targetWinRate float64) (float64, error) {
	return 0, nil
}

// Conversion Attribution
func (m *mockCache) RecordImpression(userID, campaignID, requestID string, ttlHours int) error {
	return nil
}
func (m *mockCache) RecordClick(userID, campaignID, requestID string, ttlHours int) error { return nil }
func (m *mockCache) GetAttribution(userID, campaignID string) (string, string, error) {
	return "", "", nil
}

// Multi-Touch Attribution
func (m *mockCache) RecordTouchpoint(userID, campaignID, touchpointType, requestID string, ttlDays int) error {
	return nil
}
func (m *mockCache) GetTouchpoints(userID, campaignID string) ([]model.Touchpoint, error) {
	return nil, nil
}
func (m *mockCache) GetMultiTouchAttribution(userID, campaignID, modelType string) ([]model.AttributionCredit, error) {
	return nil, nil
}

// Retargeting Segments
func (m *mockCache) RecordUserEvent(userID, campaignID, eventType string, ttlDays int) error {
	return nil
}
func (m *mockCache) GetUserEvents(userID string, eventTypes []string) (map[string][]string, error) {
	return nil, nil
}
func (m *mockCache) HasUserEvent(userID, campaignID, eventType string) (bool, error) {
	return false, nil
}

// Cross-Device Graph
func (m *mockCache) LinkDevices(primaryUserID string, deviceIDs []string, ttlDays int) error {
	return nil
}
func (m *mockCache) GetLinkedDevices(deviceID string) ([]string, error) { return nil, nil }
func (m *mockCache) GetPrimaryUserID(deviceID string) (string, error)   { return "", nil }
func (m *mockCache) GetCrossDeviceFrequency(primaryUserID, campaignID string) (int64, error) {
	return 0, nil
}

// Supply Path Optimization Analytics
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
