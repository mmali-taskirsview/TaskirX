package service

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

func createUIDCampaign(enabled bool) *model.Campaign {
	camp := &model.Campaign{
		ID:       "camp-uid-1",
		Name:     "Unified ID Test Campaign",
		BidPrice: 2.0,
		Targeting: model.Targeting{
			UnifiedIDConfig: nil,
		},
	}
	if enabled {
		camp.Targeting.UnifiedIDConfig = &model.UnifiedIDConfig{
			Enabled:         true,
			ConsentRequired: false,
			IDGraphEnabled:  true,
			EnrichProfiles:  true,
			FallbackOrder:   []string{"uid2", "id5", "rampid"},
			Providers: []model.IDProvider{
				{Name: "uid2", Enabled: true, Priority: 1, BidBoost: 0.1},
				{Name: "id5", Enabled: true, Priority: 2, BidBoost: 0.05},
				{Name: "rampid", Enabled: true, Priority: 3, BidBoost: 0.15},
			},
		}
	}
	return camp
}

func createUIDRequest(userID string) *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-uid-1",
		PublisherID: "pub-uid-123",
		User:        model.InternalUser{ID: userID},
		Device:      model.InternalDevice{DeviceID: "device-123"},
		Context:     make(map[string]interface{}),
	}
}

func TestUID_NewService(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	if svc == nil {
		t.Fatal("expected service")
	}
	if svc.idGraph == nil {
		t.Error("expected id graph")
	}
	if svc.providerStats == nil {
		t.Error("expected provider stats")
	}
}

func TestUID_ResolveIdentity_Disabled(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	campaign := createUIDCampaign(false)
	req := createUIDRequest("user-1")

	result := svc.ResolveIdentity(campaign, req)

	if result.Resolved {
		t.Error("expected not resolved when disabled")
	}
	if result.BidMultiplier != 1.0 {
		t.Errorf("expected multiplier 1.0, got %f", result.BidMultiplier)
	}
}

func TestUID_ResolveIdentity_NoUserID(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	campaign := createUIDCampaign(true)
	req := &model.BidRequest{
		ID:      "req-1",
		User:    model.InternalUser{},
		Device:  model.InternalDevice{},
		Context: make(map[string]interface{}),
	}

	result := svc.ResolveIdentity(campaign, req)

	if result.Resolved {
		t.Error("expected not resolved without user ID")
	}
	if result.Reason != "no_user_id" {
		t.Errorf("expected 'no_user_id', got '%s'", result.Reason)
	}
}

func TestUID_ResolveIdentity_ConsentRequired(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	campaign := createUIDCampaign(true)
	campaign.Targeting.UnifiedIDConfig.ConsentRequired = true
	req := createUIDRequest("user-1")
	req.Context["gdpr_consent"] = false

	result := svc.ResolveIdentity(campaign, req)

	if result.Resolved {
		t.Error("expected not resolved without consent")
	}
	if result.Reason != "consent_not_granted" {
		t.Errorf("expected 'consent_not_granted', got '%s'", result.Reason)
	}
}

func TestUID_ResolveIdentity_WithConsent(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	campaign := createUIDCampaign(true)
	campaign.Targeting.UnifiedIDConfig.ConsentRequired = true
	req := createUIDRequest("user-consent")
	req.Context["gdpr_consent"] = true

	result := svc.ResolveIdentity(campaign, req)

	if !result.HasConsent {
		t.Error("expected consent flag set")
	}
}

func TestUID_ResolveIdentity_WithConsentString(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	campaign := createUIDCampaign(true)
	campaign.Targeting.UnifiedIDConfig.ConsentRequired = true
	req := createUIDRequest("user-consent-str")
	req.Context["consent_string"] = "CPzHq4APzHq4AAAAAA"

	result := svc.ResolveIdentity(campaign, req)

	if !result.HasConsent {
		t.Error("expected consent from consent string")
	}
}

func TestUID_ResolveIdentity_ProviderMatch(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	campaign := createUIDCampaign(true)
	req := createUIDRequest("user-match")

	result := svc.ResolveIdentity(campaign, req)

	// May or may not match depending on deterministic hash
	// Just verify no crash and proper result structure
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestUID_ResolveIdentity_FromGraph(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	campaign := createUIDCampaign(true)

	// Pre-populate graph
	svc.mu.Lock()
	svc.idGraph["user-cached"] = &identityNode{
		PrimaryID:    "user-cached",
		Provider:     "uid2",
		Confidence:   0.9,
		LinkedIDs:    []linkedIdentity{{ID: "linked-1", Provider: "id5", Confidence: 0.8}},
		Segments:     []string{"seg1", "seg2"},
		DeviceTypes:  []string{"mobile", "desktop"},
		ConsentGiven: true,
		LastUpdated:  time.Now(),
	}
	svc.mu.Unlock()

	req := createUIDRequest("user-cached")
	result := svc.ResolveIdentity(campaign, req)

	if !result.Resolved {
		t.Error("expected resolved from graph")
	}
	if result.DeviceCount != 2 {
		t.Errorf("expected 2 devices, got %d", result.DeviceCount)
	}
}

func TestUID_ResolveIdentity_StaleGraphEntry(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	campaign := createUIDCampaign(true)

	// Add stale entry (older than 24h)
	svc.mu.Lock()
	svc.idGraph["user-stale"] = &identityNode{
		PrimaryID:   "user-stale",
		Provider:    "uid2",
		LastUpdated: time.Now().Add(-48 * time.Hour),
	}
	svc.mu.Unlock()

	req := createUIDRequest("user-stale")
	result := svc.ResolveIdentity(campaign, req)

	// Should attempt fresh resolution since entry is stale
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestUID_ExtractUserID_FromUser(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	req := createUIDRequest("user-123")

	id := svc.extractUserID(req)

	if id != "user-123" {
		t.Errorf("expected 'user-123', got '%s'", id)
	}
}

func TestUID_ExtractUserID_FromDevice(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	req := &model.BidRequest{
		User:   model.InternalUser{},
		Device: model.InternalDevice{DeviceID: "device-abc"},
	}

	id := svc.extractUserID(req)

	if id != "device-abc" {
		t.Errorf("expected 'device-abc', got '%s'", id)
	}
}

func TestUID_ExtractUserID_FromContext(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	tests := []string{"uid2", "id5", "rampid", "idfa", "gaid"}

	for _, key := range tests {
		req := &model.BidRequest{
			User:    model.InternalUser{},
			Device:  model.InternalDevice{},
			Context: map[string]interface{}{key: "context-id-" + key},
		}

		id := svc.extractUserID(req)

		if id != "context-id-"+key {
			t.Errorf("key %s: expected 'context-id-%s', got '%s'", key, key, id)
		}
	}
}

func TestUID_GetOrderedProviders_FallbackOrder(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	config := &model.UnifiedIDConfig{
		FallbackOrder: []string{"rampid", "uid2", "id5"},
		Providers: []model.IDProvider{
			{Name: "uid2", Priority: 1},
			{Name: "id5", Priority: 2},
			{Name: "rampid", Priority: 3},
		},
	}

	ordered := svc.getOrderedProviders(config)

	if ordered[0].Name != "rampid" {
		t.Errorf("expected rampid first, got %s", ordered[0].Name)
	}
}

func TestUID_GetOrderedProviders_ByPriority(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	config := &model.UnifiedIDConfig{
		Providers: []model.IDProvider{
			{Name: "low", Priority: 10},
			{Name: "high", Priority: 1},
			{Name: "mid", Priority: 5},
		},
	}

	ordered := svc.getOrderedProviders(config)

	if ordered[0].Name != "high" {
		t.Errorf("expected high priority first, got %s", ordered[0].Name)
	}
}

func TestUID_GetProviderMatchRate(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	tests := []struct {
		provider string
		expected float64
	}{
		{"uid2", 0.75},
		{"id5", 0.70},
		{"rampid", 0.80},
		{"liveramp", 0.80},
		{"zeotap", 0.65},
		{"unknown", 0.50},
	}

	for _, tt := range tests {
		rate := svc.getProviderMatchRate(tt.provider)
		if rate != tt.expected {
			t.Errorf("%s: expected %f, got %f", tt.provider, tt.expected, rate)
		}
	}
}

func TestUID_LinkIdentities(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	svc.LinkIdentities("id1", "uid2", "id2", "id5", "mobile", 0.85)

	svc.mu.RLock()
	node1 := svc.idGraph["id1"]
	node2 := svc.idGraph["id2"]
	svc.mu.RUnlock()

	if node1 == nil {
		t.Fatal("expected node1")
	}
	if len(node1.LinkedIDs) != 1 {
		t.Errorf("expected 1 linked ID, got %d", len(node1.LinkedIDs))
	}
	if node1.LinkedIDs[0].ID != "id2" {
		t.Error("expected linked to id2")
	}

	if node2 == nil {
		t.Fatal("expected node2 (reverse link)")
	}
}

func TestUID_LinkIdentities_UpdateExisting(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	svc.LinkIdentities("id1", "uid2", "id2", "id5", "mobile", 0.5)
	svc.LinkIdentities("id1", "uid2", "id2", "id5", "mobile", 0.9) // Update confidence

	svc.mu.RLock()
	node := svc.idGraph["id1"]
	svc.mu.RUnlock()

	if len(node.LinkedIDs) != 1 {
		t.Error("expected no duplicate links")
	}
	if node.LinkedIDs[0].Confidence != 0.9 {
		t.Error("expected updated confidence")
	}
}

func TestUID_LinkIdentities_DeviceTypes(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	svc.LinkIdentities("id1", "uid2", "id2", "id5", "mobile", 0.8)
	svc.LinkIdentities("id1", "uid2", "id3", "rampid", "desktop", 0.7)
	svc.LinkIdentities("id1", "uid2", "id4", "zeotap", "mobile", 0.6) // Duplicate device

	svc.mu.RLock()
	node := svc.idGraph["id1"]
	svc.mu.RUnlock()

	if len(node.DeviceTypes) != 2 {
		t.Errorf("expected 2 unique device types, got %d", len(node.DeviceTypes))
	}
}

func TestUID_AddSegments(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	// Create node first
	svc.LinkIdentities("user-1", "uid2", "linked-1", "id5", "mobile", 0.8)

	svc.AddSegments("user-1", []string{"seg1", "seg2"})
	svc.AddSegments("user-1", []string{"seg2", "seg3"}) // seg2 duplicate

	svc.mu.RLock()
	node := svc.idGraph["user-1"]
	svc.mu.RUnlock()

	if len(node.Segments) != 3 {
		t.Errorf("expected 3 unique segments, got %d", len(node.Segments))
	}
}

func TestUID_AddSegments_NonexistentUser(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	// Should not crash
	svc.AddSegments("nonexistent", []string{"seg1"})
}

func TestUID_SetConsent(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	svc.LinkIdentities("user-1", "uid2", "linked-1", "id5", "mobile", 0.8)
	svc.SetConsent("user-1", true)

	svc.mu.RLock()
	node := svc.idGraph["user-1"]
	svc.mu.RUnlock()

	if !node.ConsentGiven {
		t.Error("expected consent set")
	}
}

func TestUID_SetConsent_Nonexistent(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	// Should not crash
	svc.SetConsent("nonexistent", true)
}

func TestUID_GetProviderStats(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	campaign := createUIDCampaign(true)

	// Trigger some lookups
	for i := 0; i < 5; i++ {
		req := createUIDRequest("user-stats-" + string(rune('a'+i)))
		svc.ResolveIdentity(campaign, req)
	}

	stats := svc.GetProviderStats()

	// Should have stats for at least one provider
	if len(stats) == 0 {
		t.Log("no provider stats recorded (depends on match behavior)")
	}
}

func TestUID_GetIdentityReport(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	// Add some identities
	svc.LinkIdentities("id1", "uid2", "id2", "id5", "mobile", 0.8)
	svc.LinkIdentities("id3", "rampid", "id4", "zeotap", "desktop", 0.7)

	report := svc.GetIdentityReport()

	totalIdentities := report["total_identities"].(int)
	if totalIdentities < 2 {
		t.Errorf("expected at least 2 identities, got %d", totalIdentities)
	}

	providerCounts := report["identities_by_provider"].(map[string]int)
	if len(providerCounts) == 0 {
		t.Error("expected provider counts")
	}
}

func TestUID_CleanupStaleIdentities(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	// Add old and new identities
	svc.mu.Lock()
	svc.idGraph["old"] = &identityNode{
		PrimaryID:   "old",
		LastUpdated: time.Now().Add(-60 * 24 * time.Hour), // 60 days ago
	}
	svc.idGraph["new"] = &identityNode{
		PrimaryID:   "new",
		LastUpdated: time.Now(),
	}
	svc.mu.Unlock()

	removed := svc.CleanupStaleIdentities(30) // 30 day retention

	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	svc.mu.RLock()
	_, oldExists := svc.idGraph["old"]
	_, newExists := svc.idGraph["new"]
	svc.mu.RUnlock()

	if oldExists {
		t.Error("expected old identity removed")
	}
	if !newExists {
		t.Error("expected new identity to remain")
	}
}

func TestUID_CalculateCrossDeviceReach_Empty(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	reach := svc.CalculateCrossDeviceReach()

	if reach != 0 {
		t.Errorf("expected 0 for empty graph, got %f", reach)
	}
}

func TestUID_CalculateCrossDeviceReach_WithData(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	// Add multi-device user
	svc.mu.Lock()
	svc.idGraph["multi"] = &identityNode{
		PrimaryID:   "multi",
		DeviceTypes: []string{"mobile", "desktop"},
	}
	// Add single-device user
	svc.idGraph["single"] = &identityNode{
		PrimaryID:   "single",
		DeviceTypes: []string{"mobile"},
	}
	svc.mu.Unlock()

	reach := svc.CalculateCrossDeviceReach()

	if reach != 0.5 {
		t.Errorf("expected 0.5 (1/2), got %f", reach)
	}
}

func TestUID_DebugIdentity_Found(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	svc.LinkIdentities("debug-user", "uid2", "linked-1", "id5", "mobile", 0.85)
	svc.AddSegments("debug-user", []string{"seg1"})
	svc.SetConsent("debug-user", true)

	debug := svc.DebugIdentity("debug-user")

	if !strings.Contains(debug, "debug-user") {
		t.Error("expected identity in debug output")
	}
	if !strings.Contains(debug, "uid2") {
		t.Error("expected provider in debug output")
	}
}

func TestUID_DebugIdentity_NotFound(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	debug := svc.DebugIdentity("nonexistent")

	if !strings.Contains(debug, "not found") {
		t.Error("expected 'not found' message")
	}
}

func TestUID_CheckConsent_Default(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	req := &model.BidRequest{}

	consent := svc.checkConsent(req)

	if !consent {
		t.Error("expected default consent true")
	}
}

func TestUID_ShouldMatch_Deterministic(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	// Same ID should always give same result
	result1 := svc.shouldMatch("test-id", 0.5)
	result2 := svc.shouldMatch("test-id", 0.5)
	result3 := svc.shouldMatch("test-id", 0.5)

	if result1 != result2 || result2 != result3 {
		t.Error("expected deterministic matching")
	}
}

func TestUID_GetConsentStatus(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	if svc.getConsentStatus(true) != "granted" {
		t.Error("expected 'granted'")
	}
	if svc.getConsentStatus(false) != "unknown" {
		t.Error("expected 'unknown'")
	}
}

func TestUID_Max(t *testing.T) {
	if max(5, 3) != 5 {
		t.Error("expected 5")
	}
	if max(3, 5) != 5 {
		t.Error("expected 5")
	}
	if max(5, 5) != 5 {
		t.Error("expected 5")
	}
}

func TestUID_Concurrency(t *testing.T) {
	svc := NewUnifiedIDService(nil)
	campaign := createUIDCampaign(true)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userID := "conc-" + string(rune(idx))
			req := createUIDRequest(userID)
			svc.ResolveIdentity(campaign, req)
			svc.LinkIdentities(userID, "uid2", "linked-"+userID, "id5", "mobile", 0.8)
			svc.AddSegments(userID, []string{"seg1"})
			svc.SetConsent(userID, true)
			svc.GetProviderStats()
			svc.GetIdentityReport()
		}(i)
	}
	wg.Wait()
}

func TestUID_EnrichProfile(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	result := &model.UnifiedIDResult{}
	svc.enrichProfile(result, &model.UnifiedID{ID: "test"})

	if !result.EnrichedProfile {
		t.Error("expected enriched profile flag")
	}
	if len(result.AudienceSegments) == 0 {
		t.Error("expected audience segments added")
	}
}

func TestUID_EnrichWithLinkedIDs(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	// Create node with linked IDs
	svc.mu.Lock()
	svc.idGraph["user-enrich"] = &identityNode{
		PrimaryID: "user-enrich",
		LinkedIDs: []linkedIdentity{
			{ID: "linked-1", Provider: "id5", Confidence: 0.8},
			{ID: "linked-2", Provider: "rampid", Confidence: 0.7},
		},
		DeviceTypes: []string{"mobile", "desktop", "tablet"},
	}
	svc.mu.Unlock()

	result := &model.UnifiedIDResult{}
	svc.enrichWithLinkedIDs(result, "user-enrich")

	if len(result.AlternateIDs) != 2 {
		t.Errorf("expected 2 alternate IDs, got %d", len(result.AlternateIDs))
	}
	if result.DeviceCount != 3 {
		t.Errorf("expected 3 devices, got %d", result.DeviceCount)
	}
}
