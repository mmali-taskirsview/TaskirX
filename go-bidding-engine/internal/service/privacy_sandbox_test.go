package service

import (
	"sync"
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

func createPSCampaign(enabled bool) *model.Campaign {
	camp := &model.Campaign{
		ID:       "camp-ps-1",
		Name:     "PS Test Campaign",
		TenantID: "tenant-1",
		BidPrice: 2.0,
		Creative: model.Creative{
			URL: "https://ads.example.com/creative",
		},
		Targeting: model.Targeting{
			PrivacySandbox: nil,
		},
	}
	if enabled {
		camp.Targeting.PrivacySandbox = &model.PrivacySandbox{
			Enabled:          true,
			FledgeEnabled:    true,
			FallbackStrategy: "contextual",
			TopicsAPI: &model.TopicsAPIConfig{
				Enabled:       true,
				TargetTopics:  []int{6, 7, 23}, // Computers, Finance, Shopping
				ExcludeTopics: []int{9},        // Games
				TopicBidBoosts: []model.TopicBidBoost{
					{TopicID: 7, Multiplier: 1.5},
				},
			},
			AttributionAPI: &model.AttributionAPIConfig{
				Enabled:         true,
				SourceEventID:   "src-123",
				ReportingOrigin: "https://attribution.example.com",
				DebugMode:       true,
			},
		}
	}
	return camp
}

func createPSRequest(userID string) *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-ps-1",
		PublisherID: "pub-ps-123",
		User:        model.InternalUser{ID: userID},
		Device: model.InternalDevice{
			UserAgent: "Mozilla/5.0 Chrome/120",
		},
		Context: make(map[string]interface{}),
	}
}

func TestPS_NewService(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	if svc == nil {
		t.Fatal("expected service")
	}
	if svc.topicCache == nil {
		t.Error("expected topic cache")
	}
	if svc.fledgeAudiences == nil {
		t.Error("expected fledge audiences")
	}
}

func TestPS_EvaluatePrivacySandbox_Disabled(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(false)
	req := createPSRequest("user-1")

	result := svc.EvaluatePrivacySandbox(campaign, req)

	if result.TopicsAvailable {
		t.Error("expected topics not available when disabled")
	}
	if result.Reason != "privacy_sandbox_disabled" {
		t.Errorf("expected disabled reason, got '%s'", result.Reason)
	}
}

func TestPS_EvaluatePrivacySandbox_NoTopics(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)
	req := createPSRequest("user-1")
	req.Device.UserAgent = "Mozilla/5.0 Firefox/120" // Non-Chrome for no FLEDGE

	result := svc.EvaluatePrivacySandbox(campaign, req)

	if result.TopicsAvailable {
		t.Error("expected no topics available")
	}
	if !result.FallbackUsed {
		t.Error("expected fallback used")
	}
}

func TestPS_EvaluatePrivacySandbox_TopicMatch(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)
	req := createPSRequest("user-1")

	// Register a target topic for user
	svc.RegisterUserTopic("user-1", 7) // Finance

	result := svc.EvaluatePrivacySandbox(campaign, req)

	if !result.TopicsAvailable {
		t.Error("expected topics available")
	}
	if !result.TopicMatch {
		t.Error("expected topic match")
	}
	if result.TopicMultiplier != 1.5 {
		t.Errorf("expected multiplier 1.5, got %f", result.TopicMultiplier)
	}
}

func TestPS_EvaluatePrivacySandbox_TopicMatchDefaultBoost(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)
	req := createPSRequest("user-1")

	// Register a target topic without specific boost
	svc.RegisterUserTopic("user-1", 23) // Shopping

	result := svc.EvaluatePrivacySandbox(campaign, req)

	if !result.TopicMatch {
		t.Error("expected topic match")
	}
	if result.TopicMultiplier != 1.2 {
		t.Errorf("expected default boost 1.2, got %f", result.TopicMultiplier)
	}
}

func TestPS_EvaluatePrivacySandbox_ExcludedTopic(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)
	req := createPSRequest("user-1")

	// Register an excluded topic
	svc.RegisterUserTopic("user-1", 9) // Games (excluded)

	result := svc.EvaluatePrivacySandbox(campaign, req)

	if result.TopicMatch {
		t.Error("expected no match for excluded topic")
	}
	if result.TopicMultiplier != 0 {
		t.Errorf("expected multiplier 0 for excluded, got %f", result.TopicMultiplier)
	}
	if result.Reason != "excluded_topic_match" {
		t.Errorf("expected excluded reason, got '%s'", result.Reason)
	}
}

func TestPS_EvaluatePrivacySandbox_TopicsFromContext(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)
	req := createPSRequest("user-1")

	// Set topics in request context
	req.Context["chrome_topics"] = []interface{}{float64(7), float64(23)}

	result := svc.EvaluatePrivacySandbox(campaign, req)

	if !result.TopicsAvailable {
		t.Error("expected topics available from context")
	}
	if !result.TopicMatch {
		t.Error("expected topic match from context")
	}
}

func TestPS_EvaluatePrivacySandbox_Attribution(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)
	req := createPSRequest("user-1")

	result := svc.EvaluatePrivacySandbox(campaign, req)

	if !result.AttributionEnabled {
		t.Error("expected attribution enabled")
	}
}

func TestPS_EvaluatePrivacySandbox_Fledge(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)
	req := createPSRequest("user-1")
	req.Context["fledge_supported"] = true

	result := svc.EvaluatePrivacySandbox(campaign, req)

	if !result.FledgeEligible {
		t.Error("expected FLEDGE eligible")
	}
}

func TestPS_EvaluatePrivacySandbox_ProtectedAudience(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)
	req := createPSRequest("user-1")
	req.Context["protected_audience"] = true

	result := svc.EvaluatePrivacySandbox(campaign, req)

	if !result.FledgeEligible {
		t.Error("expected Protected Audience eligible")
	}
}

func TestPS_RegisterUserTopic(t *testing.T) {
	svc := NewPrivacySandboxService(nil)

	svc.RegisterUserTopic("user-1", 7)
	svc.RegisterUserTopic("user-1", 23)

	svc.mu.RLock()
	topics := svc.topicCache["user-1"]
	svc.mu.RUnlock()

	if len(topics) != 2 {
		t.Errorf("expected 2 topics, got %d", len(topics))
	}
}

func TestPS_RegisterUserTopic_MaxFive(t *testing.T) {
	svc := NewPrivacySandboxService(nil)

	for i := 1; i <= 7; i++ {
		svc.RegisterUserTopic("user-1", i)
	}

	svc.mu.RLock()
	topics := svc.topicCache["user-1"]
	svc.mu.RUnlock()

	if len(topics) != 5 {
		t.Errorf("expected max 5 topics, got %d", len(topics))
	}
	// Should have topics 3-7 (first 2 removed)
	if topics[0] != 3 {
		t.Errorf("expected first topic 3, got %d", topics[0])
	}
}

func TestPS_AddToInterestGroup(t *testing.T) {
	svc := NewPrivacySandboxService(nil)

	svc.AddToInterestGroup("user-1", "sports_fans")
	svc.AddToInterestGroup("user-1", "tech_enthusiasts")

	groups := svc.GetUserInterestGroups("user-1")

	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}

func TestPS_AddToInterestGroup_NoDuplicates(t *testing.T) {
	svc := NewPrivacySandboxService(nil)

	svc.AddToInterestGroup("user-1", "sports_fans")
	svc.AddToInterestGroup("user-1", "sports_fans")
	svc.AddToInterestGroup("user-1", "sports_fans")

	groups := svc.GetUserInterestGroups("user-1")

	if len(groups) != 1 {
		t.Errorf("expected 1 group (no duplicates), got %d", len(groups))
	}
}

func TestPS_GetUserInterestGroups_Empty(t *testing.T) {
	svc := NewPrivacySandboxService(nil)

	groups := svc.GetUserInterestGroups("nonexistent")

	if groups != nil {
		t.Error("expected nil for nonexistent user")
	}
}

func TestPS_GetTopicName(t *testing.T) {
	svc := NewPrivacySandboxService(nil)

	tests := []struct {
		topicID  int
		expected string
	}{
		{7, "Finance"},
		{24, "Sports"},
		{999, "Unknown"},
	}

	for _, tt := range tests {
		name := svc.GetTopicName(tt.topicID)
		if name != tt.expected {
			t.Errorf("topic %d: expected '%s', got '%s'", tt.topicID, tt.expected, name)
		}
	}
}

func TestPS_GenerateAttributionSource_Disabled(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(false)
	req := createPSRequest("user-1")

	source := svc.GenerateAttributionSource(nil, campaign, req)

	if source != nil {
		t.Error("expected nil when disabled")
	}
}

func TestPS_GenerateAttributionSource_Enabled(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)
	req := createPSRequest("user-1")

	source := svc.GenerateAttributionSource(
		campaign.Targeting.PrivacySandbox.AttributionAPI,
		campaign,
		req,
	)

	if source == nil {
		t.Fatal("expected source")
	}
	if source["source_event_id"] != "src-123" {
		t.Error("expected source event ID")
	}
	if source["debug_key"] == "" {
		t.Error("expected debug key when debug mode enabled")
	}
}

func TestPS_GenerateAttributionSource_NoDebug(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)
	campaign.Targeting.PrivacySandbox.AttributionAPI.DebugMode = false
	req := createPSRequest("user-1")

	source := svc.GenerateAttributionSource(
		campaign.Targeting.PrivacySandbox.AttributionAPI,
		campaign,
		req,
	)

	if source["debug_key"] != "" {
		t.Error("expected empty debug key when debug disabled")
	}
}

func TestPS_GenerateFledgeBid(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)
	req := createPSRequest("user-1")

	bid := svc.GenerateFledgeBid(campaign, req)

	if bid == nil {
		t.Fatal("expected bid")
	}
	if bid["interest_group_owner"] != campaign.TenantID {
		t.Error("expected tenant ID as owner")
	}
	if bid["bid_currency"] != "USD" {
		t.Error("expected USD currency")
	}
}

func TestPS_GenerateFledgeBid_NoTenant(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)
	campaign.TenantID = ""
	req := createPSRequest("user-1")

	bid := svc.GenerateFledgeBid(campaign, req)

	if bid["interest_group_owner"] != campaign.ID {
		t.Error("expected campaign ID as fallback owner")
	}
}

func TestPS_GeneratePrivateAggregationReport(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)

	report := svc.GeneratePrivateAggregationReport(campaign, 12345, 100)

	if report == nil {
		t.Fatal("expected report")
	}
	contributions := report["contributions"].([]map[string]interface{})
	if len(contributions) != 1 {
		t.Error("expected 1 contribution")
	}
	if contributions[0]["bucket"].(int64) != 12345 {
		t.Error("expected bucket 12345")
	}
}

func TestPS_SimulateSharedStorage(t *testing.T) {
	svc := NewPrivacySandboxService(nil)

	value, exists := svc.SimulateSharedStorage("user-1", "frequency_cap")

	// Current implementation returns empty
	if exists {
		t.Error("expected not exists")
	}
	if value != "" {
		t.Error("expected empty value")
	}
}

func TestPS_CleanupExpiredTopics(t *testing.T) {
	svc := NewPrivacySandboxService(nil)

	// Add some topics
	svc.RegisterUserTopic("user-1", 7)
	svc.RegisterUserTopic("user-2", 23)

	// Cleanup (doesn't actually remove in current implementation)
	svc.CleanupExpiredTopics(21)

	// Just verify no crash
}

func TestPS_ContainsString(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"Chrome/120", "Chrome/", true},
		{"Mozilla/5.0", "Chrome/", false},
		{"Chrome/", "Chrome/", true},
		{"", "Chrome/", false},
	}

	for _, tt := range tests {
		result := containsString(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("containsString(%q, %q): expected %v, got %v", tt.s, tt.substr, tt.expected, result)
		}
	}
}

func TestPS_CheckFledgeEligibility_Chrome(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	req := createPSRequest("user-1")
	req.Device.UserAgent = "Mozilla/5.0 Chrome/120.0.0.0"

	eligible := svc.checkFledgeEligibility(req)

	if !eligible {
		t.Error("expected Chrome to be FLEDGE eligible")
	}
}

func TestPS_CheckFledgeEligibility_NonChrome(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	req := createPSRequest("user-1")
	req.Device.UserAgent = "Mozilla/5.0 Firefox/120"

	eligible := svc.checkFledgeEligibility(req)

	if eligible {
		t.Error("expected Firefox not FLEDGE eligible")
	}
}

func TestPS_Concurrency(t *testing.T) {
	svc := NewPrivacySandboxService(nil)
	campaign := createPSCampaign(true)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userID := "conc-" + string(rune(idx))
			svc.RegisterUserTopic(userID, idx%25+1)
			svc.AddToInterestGroup(userID, "group-"+string(rune(idx%5)))

			req := createPSRequest(userID)
			svc.EvaluatePrivacySandbox(campaign, req)
			svc.GetUserInterestGroups(userID)
			svc.GetTopicName(idx % 25)
		}(i)
	}
	wg.Wait()
}

func TestPS_TopicTaxonomy(t *testing.T) {
	// Verify all expected topics exist
	expectedTopics := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}

	for _, id := range expectedTopics {
		if _, exists := topicTaxonomy[id]; !exists {
			t.Errorf("expected topic %d in taxonomy", id)
		}
	}
}
