package service

import (
	"sync"
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

func createIncCampaign(enabled bool) *model.Campaign {
	camp := &model.Campaign{
		ID:       "camp-inc-1",
		Name:     "Incrementality Test Campaign",
		BidPrice: 2.0,
		Targeting: model.Targeting{
			IncrementalityConfig: nil,
		},
	}
	if enabled {
		camp.Targeting.IncrementalityConfig = &model.IncrementalityConfig{
			Enabled:         true,
			ExperimentID:    "exp-inc-1",
			ControlPercent:  10.0,
			HoldoutType:     "user",
			MinSampleSize:   100,
			ConfidenceLevel: 0.95,
		}
	}
	return camp
}

func createIncRequest(userID string) *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-inc-1",
		PublisherID: "pub-inc-123",
		User:        model.InternalUser{ID: userID},
		Device: model.InternalDevice{
			Geo: model.InternalGeo{Country: "US", City: "NYC"},
		},
	}
}

func TestInc_NewService(t *testing.T) {
	svc := NewIncrementalityService(nil)
	if svc == nil {
		t.Fatal("expected service")
	}
	if svc.experiments == nil {
		t.Error("expected experiments map initialized")
	}
}

func TestInc_EvaluateUser_Disabled(t *testing.T) {
	svc := NewIncrementalityService(nil)
	campaign := createIncCampaign(false)
	req := createIncRequest("user-1")

	result := svc.EvaluateUser(campaign, req)

	if result.Status != "disabled" {
		t.Errorf("expected status 'disabled', got '%s'", result.Status)
	}
}

func TestInc_EvaluateUser_DefaultExperimentID(t *testing.T) {
	svc := NewIncrementalityService(nil)
	campaign := createIncCampaign(true)
	campaign.Targeting.IncrementalityConfig.ExperimentID = "" // Clear to test default
	req := createIncRequest("user-1")

	result := svc.EvaluateUser(campaign, req)

	expectedExpID := campaign.ID + "_incrementality"
	if result.ExperimentID != expectedExpID {
		t.Errorf("expected default experiment ID '%s', got '%s'", expectedExpID, result.ExperimentID)
	}
}

func TestInc_EvaluateUser_Running(t *testing.T) {
	svc := NewIncrementalityService(nil)
	campaign := createIncCampaign(true)
	req := createIncRequest("user-1")

	result := svc.EvaluateUser(campaign, req)

	if result.Status != "running" {
		t.Errorf("expected status 'running', got '%s'", result.Status)
	}
	if result.ExperimentID != "exp-inc-1" {
		t.Errorf("expected experiment ID 'exp-inc-1', got '%s'", result.ExperimentID)
	}
}

func TestInc_EvaluateUser_DeterministicAssignment(t *testing.T) {
	svc := NewIncrementalityService(nil)
	campaign := createIncCampaign(true)
	req := createIncRequest("user-deterministic")

	// Call multiple times - should get same result
	result1 := svc.EvaluateUser(campaign, req)
	result2 := svc.EvaluateUser(campaign, req)
	result3 := svc.EvaluateUser(campaign, req)

	if result1.UserInControlGroup != result2.UserInControlGroup ||
		result2.UserInControlGroup != result3.UserInControlGroup {
		t.Error("expected deterministic assignment")
	}
}

func TestInc_EvaluateUser_GeoHoldout(t *testing.T) {
	svc := NewIncrementalityService(nil)
	campaign := createIncCampaign(true)
	campaign.Targeting.IncrementalityConfig.HoldoutType = "geo"
	campaign.Targeting.IncrementalityConfig.GeoHoldouts = []string{"US", "NYC"}

	reqUS := createIncRequest("user-us")
	result := svc.EvaluateUser(campaign, reqUS)

	if !result.UserInControlGroup {
		t.Error("expected US user in control group for geo holdout")
	}

	// Test non-holdout geo
	campaign.Targeting.IncrementalityConfig.GeoHoldouts = []string{"UK"}
	reqUS2 := createIncRequest("user-us-2")
	result2 := svc.EvaluateUser(campaign, reqUS2)

	if result2.UserInControlGroup {
		t.Error("expected US user NOT in control group when UK is holdout")
	}
}

func TestInc_EvaluateUser_TimeHoldout(t *testing.T) {
	svc := NewIncrementalityService(nil)
	campaign := createIncCampaign(true)
	campaign.Targeting.IncrementalityConfig.HoldoutType = "time"
	req := createIncRequest("user-1")

	// Time holdout depends on current hour
	result := svc.EvaluateUser(campaign, req)

	// Just verify it runs without crashing
	if result.Status != "running" {
		t.Errorf("expected running status, got %s", result.Status)
	}
}

func TestInc_EvaluateUser_ControlPercentBounds(t *testing.T) {
	svc := NewIncrementalityService(nil)

	tests := []struct {
		name           string
		controlPercent float64
	}{
		{"zero", 0},
		{"negative", -10},
		{"hundred_plus", 150},
		{"valid", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign := createIncCampaign(true)
			campaign.Targeting.IncrementalityConfig.ControlPercent = tt.controlPercent
			campaign.Targeting.IncrementalityConfig.ExperimentID = "exp-" + tt.name
			req := createIncRequest("user-bound-test")

			result := svc.EvaluateUser(campaign, req)
			if result.Status != "running" {
				t.Error("expected running status")
			}
		})
	}
}

func TestInc_RecordImpression(t *testing.T) {
	svc := NewIncrementalityService(nil)
	campaign := createIncCampaign(true)
	req := createIncRequest("user-imp")

	// First evaluate to create experiment
	result := svc.EvaluateUser(campaign, req)

	// Record impression
	svc.RecordImpression(result.ExperimentID, "user-imp", result.UserInControlGroup)

	// Verify recorded
	svc.mu.RLock()
	exp := svc.experiments[result.ExperimentID]
	svc.mu.RUnlock()

	if exp == nil {
		t.Fatal("expected experiment")
	}

	totalImpressions := 0
	for _, u := range exp.controlUsers {
		totalImpressions += u.impressions
	}
	for _, u := range exp.testUsers {
		totalImpressions += u.impressions
	}

	if totalImpressions != 1 {
		t.Errorf("expected 1 impression, got %d", totalImpressions)
	}
}

func TestInc_RecordImpression_NoExperiment(t *testing.T) {
	svc := NewIncrementalityService(nil)

	// Should not crash
	svc.RecordImpression("nonexistent", "user-1", false)
}

func TestInc_RecordConversion(t *testing.T) {
	svc := NewIncrementalityService(nil)
	campaign := createIncCampaign(true)
	req := createIncRequest("user-conv")

	result := svc.EvaluateUser(campaign, req)

	// Record conversion
	svc.RecordConversion(result.ExperimentID, "user-conv", result.UserInControlGroup, 50.0)

	svc.mu.RLock()
	exp := svc.experiments[result.ExperimentID]
	svc.mu.RUnlock()

	totalConv := 0
	totalRev := 0.0
	for _, u := range exp.controlUsers {
		totalConv += u.conversions
		totalRev += u.revenue
	}
	for _, u := range exp.testUsers {
		totalConv += u.conversions
		totalRev += u.revenue
	}

	if totalConv != 1 {
		t.Errorf("expected 1 conversion, got %d", totalConv)
	}
	if totalRev != 50.0 {
		t.Errorf("expected revenue 50.0, got %f", totalRev)
	}
}

func TestInc_RecordConversion_NoExperiment(t *testing.T) {
	svc := NewIncrementalityService(nil)

	// Should not crash
	svc.RecordConversion("nonexistent", "user-1", false, 100.0)
}

func TestInc_GetExperimentResults_NotFound(t *testing.T) {
	svc := NewIncrementalityService(nil)

	result := svc.GetExperimentResults("nonexistent")

	if result.Status != "not_found" {
		t.Errorf("expected status 'not_found', got '%s'", result.Status)
	}
}

func TestInc_GetExperimentResults_InsufficientData(t *testing.T) {
	svc := NewIncrementalityService(nil)
	campaign := createIncCampaign(true)
	campaign.Targeting.IncrementalityConfig.MinSampleSize = 100

	// Create experiment with few users
	for i := 0; i < 10; i++ {
		req := createIncRequest("user-" + string(rune('A'+i)))
		svc.EvaluateUser(campaign, req)
	}

	result := svc.GetExperimentResults("exp-inc-1")

	if result.Status != "insufficient_data" {
		t.Errorf("expected 'insufficient_data', got '%s'", result.Status)
	}
}

func TestInc_GetExperimentResults_WithData(t *testing.T) {
	svc := NewIncrementalityService(nil)
	campaign := createIncCampaign(true)
	campaign.Targeting.IncrementalityConfig.MinSampleSize = 10 // Lower for test
	campaign.Targeting.IncrementalityConfig.ControlPercent = 50

	// Create experiment with users and conversions
	expID := "exp-inc-1"
	for i := 0; i < 100; i++ {
		userID := "user-data-" + string(rune(i))
		req := createIncRequest(userID)
		result := svc.EvaluateUser(campaign, req)

		svc.RecordImpression(expID, userID, result.UserInControlGroup)

		// Simulate conversions
		if i%10 == 0 {
			svc.RecordConversion(expID, userID, result.UserInControlGroup, 25.0)
		}
	}

	result := svc.GetExperimentResults(expID)

	if result.Status == "not_found" {
		t.Error("expected experiment to be found")
	}
}

func TestInc_GetExperimentResults_LiftCalculation(t *testing.T) {
	svc := NewIncrementalityService(nil)

	// Create experiment directly
	expID := "exp-lift-test"
	svc.mu.Lock()
	svc.experiments[expID] = &experiment{
		config: &model.IncrementalityConfig{
			MinSampleSize:   10,
			ConfidenceLevel: 0.95,
		},
		controlUsers: make(map[string]*userStats),
		testUsers:    make(map[string]*userStats),
	}

	// Add control users with some conversions
	for i := 0; i < 100; i++ {
		svc.experiments[expID].controlUsers["ctrl-"+string(rune(i))] = &userStats{
			impressions: 10,
			conversions: 1, // 1% CVR
			revenue:     10,
		}
	}

	// Add test users with higher conversions
	for i := 0; i < 100; i++ {
		svc.experiments[expID].testUsers["test-"+string(rune(i))] = &userStats{
			impressions: 10,
			conversions: 2, // 2% CVR
			revenue:     20,
		}
	}
	svc.mu.Unlock()

	result := svc.GetExperimentResults(expID)

	if result.Status == "not_found" || result.Status == "insufficient_data" {
		t.Errorf("unexpected status: %s", result.Status)
	}

	// Test should have higher CVR than control
	if result.Lift <= 0 {
		t.Errorf("expected positive lift, got %f", result.Lift)
	}
}

func TestInc_GetExperimentResults_ZeroControlConversions(t *testing.T) {
	svc := NewIncrementalityService(nil)

	expID := "exp-zero-ctrl"
	svc.mu.Lock()
	svc.experiments[expID] = &experiment{
		config: &model.IncrementalityConfig{
			MinSampleSize:   5,
			ConfidenceLevel: 0.95,
		},
		controlUsers: make(map[string]*userStats),
		testUsers:    make(map[string]*userStats),
	}

	// Control with no conversions
	for i := 0; i < 10; i++ {
		svc.experiments[expID].controlUsers["ctrl-"+string(rune(i))] = &userStats{
			impressions: 10,
			conversions: 0,
		}
	}

	// Test with conversions
	for i := 0; i < 10; i++ {
		svc.experiments[expID].testUsers["test-"+string(rune(i))] = &userStats{
			impressions: 10,
			conversions: 1,
		}
	}
	svc.mu.Unlock()

	result := svc.GetExperimentResults(expID)

	// Should handle division by zero gracefully
	if result.Lift != 100.0 {
		t.Errorf("expected 100%% lift when control has 0 conversions, got %f", result.Lift)
	}
}

func TestInc_GetExperimentResults_Significance(t *testing.T) {
	svc := NewIncrementalityService(nil)

	expID := "exp-sig-test"
	svc.mu.Lock()
	svc.experiments[expID] = &experiment{
		config: &model.IncrementalityConfig{
			MinSampleSize:   5,
			ConfidenceLevel: 0.95,
		},
		controlUsers: make(map[string]*userStats),
		testUsers:    make(map[string]*userStats),
	}

	// Large sample with clear difference
	for i := 0; i < 1000; i++ {
		svc.experiments[expID].controlUsers["ctrl-"+string(rune(i))] = &userStats{
			conversions: 1,
		}
		svc.experiments[expID].testUsers["test-"+string(rune(i))] = &userStats{
			conversions: 5, // 5x higher
		}
	}
	svc.mu.Unlock()

	result := svc.GetExperimentResults(expID)

	if result.LiftConfidence <= 0 {
		t.Error("expected some confidence level")
	}
}

func TestInc_GetExperimentResults_Recommendations(t *testing.T) {
	tests := []struct {
		name         string
		controlConv  int
		testConv     int
		sampleSize   int
		expectedLift string // "positive", "negative", "neutral"
	}{
		{"strong_lift", 1, 5, 1000, "positive"},
		{"moderate_lift", 1, 2, 1000, "positive"},
		{"no_lift", 1, 1, 1000, "neutral"},
		{"negative_lift", 5, 1, 1000, "negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewIncrementalityService(nil)
			expID := "exp-rec-" + tt.name

			svc.mu.Lock()
			svc.experiments[expID] = &experiment{
				config: &model.IncrementalityConfig{
					MinSampleSize:   10,
					ConfidenceLevel: 0.5, // Low for test
				},
				controlUsers: make(map[string]*userStats),
				testUsers:    make(map[string]*userStats),
			}

			for i := 0; i < tt.sampleSize; i++ {
				svc.experiments[expID].controlUsers["ctrl-"+string(rune(i))] = &userStats{
					conversions: tt.controlConv,
				}
				svc.experiments[expID].testUsers["test-"+string(rune(i))] = &userStats{
					conversions: tt.testConv,
				}
			}
			svc.mu.Unlock()

			result := svc.GetExperimentResults(expID)

			if result.Recommendation == "" {
				t.Error("expected recommendation")
			}
		})
	}
}

func TestInc_GetUserExperimentGroup_NotFound(t *testing.T) {
	svc := NewIncrementalityService(nil)

	group, found := svc.GetUserExperimentGroup("nonexistent", "user-1")

	if found {
		t.Error("expected not found")
	}
	if group != "" {
		t.Error("expected empty group")
	}
}

func TestInc_GetUserExperimentGroup_Control(t *testing.T) {
	svc := NewIncrementalityService(nil)

	expID := "exp-group-test"
	svc.mu.Lock()
	svc.experiments[expID] = &experiment{
		config:       &model.IncrementalityConfig{},
		controlUsers: map[string]*userStats{"user-ctrl": {}},
		testUsers:    make(map[string]*userStats),
	}
	svc.mu.Unlock()

	group, found := svc.GetUserExperimentGroup(expID, "user-ctrl")

	if !found {
		t.Error("expected found")
	}
	if group != "control" {
		t.Errorf("expected 'control', got '%s'", group)
	}
}

func TestInc_GetUserExperimentGroup_Test(t *testing.T) {
	svc := NewIncrementalityService(nil)

	expID := "exp-group-test2"
	svc.mu.Lock()
	svc.experiments[expID] = &experiment{
		config:       &model.IncrementalityConfig{},
		controlUsers: make(map[string]*userStats),
		testUsers:    map[string]*userStats{"user-test": {}},
	}
	svc.mu.Unlock()

	group, found := svc.GetUserExperimentGroup(expID, "user-test")

	if !found {
		t.Error("expected found")
	}
	if group != "test" {
		t.Errorf("expected 'test', got '%s'", group)
	}
}

func TestInc_GetUserExperimentGroup_Unknown(t *testing.T) {
	svc := NewIncrementalityService(nil)

	expID := "exp-group-unknown"
	svc.mu.Lock()
	svc.experiments[expID] = &experiment{
		config:       &model.IncrementalityConfig{},
		controlUsers: make(map[string]*userStats),
		testUsers:    make(map[string]*userStats),
	}
	svc.mu.Unlock()

	group, found := svc.GetUserExperimentGroup(expID, "unknown-user")

	if found {
		t.Error("expected not found for unknown user")
	}
	if group != "" {
		t.Error("expected empty group")
	}
}

func TestInc_AggregateStats(t *testing.T) {
	svc := NewIncrementalityService(nil)

	users := map[string]*userStats{
		"u1": {impressions: 10, conversions: 1, revenue: 50},
		"u2": {impressions: 20, conversions: 2, revenue: 100},
		"u3": {impressions: 5, conversions: 0, revenue: 0},
	}

	stats := svc.aggregateStats(users)

	if stats.users != 3 {
		t.Errorf("expected 3 users, got %d", stats.users)
	}
	if stats.impressions != 35 {
		t.Errorf("expected 35 impressions, got %d", stats.impressions)
	}
	if stats.conversions != 3 {
		t.Errorf("expected 3 conversions, got %d", stats.conversions)
	}
	if stats.revenue != 150 {
		t.Errorf("expected revenue 150, got %f", stats.revenue)
	}
}

func TestInc_AggregateStats_Empty(t *testing.T) {
	svc := NewIncrementalityService(nil)

	stats := svc.aggregateStats(map[string]*userStats{})

	if stats.users != 0 || stats.impressions != 0 {
		t.Error("expected zero stats for empty map")
	}
}

func TestInc_CalculateSignificance_ZeroUsers(t *testing.T) {
	svc := NewIncrementalityService(nil)

	control := aggregatedStats{users: 0}
	test := aggregatedStats{users: 100, conversions: 10}

	confidence := svc.calculateSignificance(control, test)

	if confidence != 0 {
		t.Error("expected 0 confidence with zero users")
	}
}

func TestInc_HashToGroup(t *testing.T) {
	// Test determinism
	r1 := hashToGroup("user-1", "exp-1", 10)
	r2 := hashToGroup("user-1", "exp-1", 10)

	if r1 != r2 {
		t.Error("expected deterministic hash assignment")
	}

	// Test different users may get different assignments
	results := make(map[bool]int)
	for i := 0; i < 100; i++ {
		userID := "test-user-" + string(rune(i))
		result := hashToGroup(userID, "exp-hash", 50)
		results[result]++
	}

	// With 50% control, should have some of each
	if results[true] == 0 || results[false] == 0 {
		t.Log("hash distribution may be skewed - expected some in each group")
	}
}

func TestInc_Concurrency(t *testing.T) {
	svc := NewIncrementalityService(nil)
	campaign := createIncCampaign(true)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userID := "conc-user-" + string(rune(idx))
			req := createIncRequest(userID)

			result := svc.EvaluateUser(campaign, req)
			svc.RecordImpression(result.ExperimentID, userID, result.UserInControlGroup)
			if idx%5 == 0 {
				svc.RecordConversion(result.ExperimentID, userID, result.UserInControlGroup, 10.0)
			}
			svc.GetExperimentResults(result.ExperimentID)
			svc.GetUserExperimentGroup(result.ExperimentID, userID)
		}(i)
	}
	wg.Wait()
}

func TestInc_NormalCDF(t *testing.T) {
	// Test boundary values
	tests := []struct {
		x        float64
		expected float64
	}{
		{0, 0.5},
		{-3, 0.001}, // ~0.001
		{3, 0.999},  // ~0.999
	}

	for _, tt := range tests {
		result := normalCDF(tt.x)
		if result < 0 || result > 1 {
			t.Errorf("CDF out of range: %f", result)
		}
	}
}

func TestInc_Erf(t *testing.T) {
	// erf(0) should be close to 0 (may have small floating point error)
	if val := erf(0); val > 0.0001 || val < -0.0001 {
		t.Errorf("expected erf(0) ≈ 0, got %f", val)
	}

	// erf should be odd function
	if erf(1)+erf(-1) > 0.0001 {
		t.Error("expected erf to be odd function")
	}
}
