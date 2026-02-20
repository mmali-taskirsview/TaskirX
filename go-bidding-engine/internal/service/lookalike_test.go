package service

import (
	"sync"
	"testing"
	"time"
)

func TestLA_NewService(t *testing.T) {
	svc := NewLookalikeService(nil)
	if svc == nil {
		t.Fatal("expected service")
	}
	if svc.userProfiles == nil {
		t.Error("expected user profiles map")
	}
	if svc.lookalikeAudiences == nil {
		t.Error("expected audiences map")
	}
	if svc.config == nil {
		t.Error("expected config")
	}
}

func TestLA_RegisterUserProfile(t *testing.T) {
	svc := NewLookalikeService(nil)

	profile := svc.CreateUserProfile(
		[]string{"segment-1"},
		[]string{"tech", "sports"},
		[]string{"mobile"},
		"US", "CA", "SF",
	)
	svc.RegisterUserProfile("user-1", profile)

	svc.mu.RLock()
	stored := svc.userProfiles["user-1"]
	svc.mu.RUnlock()

	if stored == nil {
		t.Fatal("expected profile stored")
	}
	if stored.UserID != "user-1" {
		t.Error("expected user ID set")
	}
	if len(stored.Interests) != 2 {
		t.Error("expected 2 interests")
	}
}

func TestLA_CreateUserProfile(t *testing.T) {
	svc := NewLookalikeService(nil)

	profile := svc.CreateUserProfile(
		[]string{"seg1", "seg2"},
		[]string{"finance"},
		[]string{"desktop"},
		"UK", "England", "London",
	)

	if profile == nil {
		t.Fatal("expected profile created")
	}
	if profile.Geo.Country != "UK" {
		t.Error("expected UK country")
	}
	if len(profile.Segments) != 2 {
		t.Error("expected 2 segments")
	}
}

func TestLA_UpdateUserBehavior(t *testing.T) {
	svc := NewLookalikeService(nil)

	profile := svc.CreateUserProfile(nil, nil, nil, "", "", "")
	svc.RegisterUserProfile("user-1", profile)

	svc.UpdateUserBehavior("user-1", 100, 5, 1, 10.0)

	svc.mu.RLock()
	stored := svc.userProfiles["user-1"]
	svc.mu.RUnlock()

	if stored.BehaviorMetrics.TotalImpressions != 100 {
		t.Errorf("expected 100 impressions, got %d", stored.BehaviorMetrics.TotalImpressions)
	}
	if stored.BehaviorMetrics.TotalClicks != 5 {
		t.Errorf("expected 5 clicks, got %d", stored.BehaviorMetrics.TotalClicks)
	}
}

func TestLA_UpdateUserBehavior_NonexistentUser(t *testing.T) {
	svc := NewLookalikeService(nil)

	// Should not crash
	svc.UpdateUserBehavior("nonexistent", 100, 5, 1, 10.0)
}

func TestLA_GenerateLookalike_Disabled(t *testing.T) {
	svc := NewLookalikeService(nil)
	svc.config.Enabled = false

	result := svc.GenerateLookalike([]string{"u1", "u2"}, "test", 2.0)

	if result.Status != "disabled" {
		t.Errorf("expected 'disabled', got '%s'", result.Status)
	}
}

func TestLA_GenerateLookalike_InsufficientSeed(t *testing.T) {
	svc := NewLookalikeService(nil)
	svc.config.MinSeedSize = 100

	result := svc.GenerateLookalike([]string{"u1", "u2", "u3"}, "test", 2.0)

	if result.Status != "insufficient_seed" {
		t.Errorf("expected 'insufficient_seed', got '%s'", result.Status)
	}
}

func TestLA_GenerateLookalike_NoProfiles(t *testing.T) {
	svc := NewLookalikeService(nil)
	svc.config.MinSeedSize = 2

	// Users exist in seed but have no profiles
	result := svc.GenerateLookalike([]string{"u1", "u2"}, "test", 2.0)

	if result.Status != "seed_profile_failed" {
		t.Errorf("expected 'seed_profile_failed', got '%s'", result.Status)
	}
}

func TestLA_GenerateLookalike_Success(t *testing.T) {
	svc := NewLookalikeService(nil)
	svc.config.MinSeedSize = 3
	svc.config.SimilarityThreshold = 0.1 // Low threshold for test

	// Create seed users
	for i := 0; i < 5; i++ {
		userID := "seed-" + string(rune('a'+i))
		profile := svc.CreateUserProfile(
			nil,
			[]string{"tech", "sports"},
			[]string{"mobile"},
			"US", "CA", "SF",
		)
		profile.BehaviorMetrics = behaviorMetrics{
			TotalImpressions: 100,
			TotalClicks:      10,
			EngagementScore:  0.5,
		}
		svc.RegisterUserProfile(userID, profile)
	}

	// Create candidate users
	for i := 0; i < 10; i++ {
		userID := "cand-" + string(rune('a'+i))
		profile := svc.CreateUserProfile(
			nil,
			[]string{"tech"},
			[]string{"mobile"},
			"US", "CA", "LA",
		)
		profile.BehaviorMetrics = behaviorMetrics{
			TotalImpressions: 50,
			TotalClicks:      5,
			EngagementScore:  0.4,
		}
		svc.RegisterUserProfile(userID, profile)
	}

	seedIDs := []string{"seed-a", "seed-b", "seed-c", "seed-d", "seed-e"}
	result := svc.GenerateLookalike(seedIDs, "test-audience", 2.0)

	if result.Status != "created" {
		t.Errorf("expected 'created', got '%s'", result.Status)
	}
	if result.AudienceID == "" {
		t.Error("expected audience ID")
	}
	if result.SeedSize != 5 {
		t.Errorf("expected seed size 5, got %d", result.SeedSize)
	}
}

func TestLA_GetLookalikeAudience_NotFound(t *testing.T) {
	svc := NewLookalikeService(nil)

	audience := svc.GetLookalikeAudience("nonexistent")

	if audience != nil {
		t.Error("expected nil for nonexistent")
	}
}

func TestLA_GetLookalikeAudience_Expired(t *testing.T) {
	svc := NewLookalikeService(nil)

	svc.mu.Lock()
	svc.lookalikeAudiences["expired"] = &lookalikeAudience{
		ID:        "expired",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}
	svc.mu.Unlock()

	audience := svc.GetLookalikeAudience("expired")

	if audience != nil {
		t.Error("expected nil for expired audience")
	}
}

func TestLA_GetLookalikeAudience_Valid(t *testing.T) {
	svc := NewLookalikeService(nil)

	svc.mu.Lock()
	svc.lookalikeAudiences["valid"] = &lookalikeAudience{
		ID:        "valid",
		Name:      "Test",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	svc.mu.Unlock()

	audience := svc.GetLookalikeAudience("valid")

	if audience == nil {
		t.Fatal("expected audience")
	}
	if audience.Name != "Test" {
		t.Error("expected name Test")
	}
}

func TestLA_IsUserInLookalike_NotFound(t *testing.T) {
	svc := NewLookalikeService(nil)

	found, _ := svc.IsUserInLookalike("user-1", "nonexistent")

	if found {
		t.Error("expected not found")
	}
}

func TestLA_IsUserInLookalike_UserInAudience(t *testing.T) {
	svc := NewLookalikeService(nil)

	svc.mu.Lock()
	svc.lookalikeAudiences["aud-1"] = &lookalikeAudience{
		ID: "aud-1",
		Users: []lookalikeUser{
			{UserID: "user-1", SimilarityScore: 0.8},
			{UserID: "user-2", SimilarityScore: 0.7},
		},
	}
	svc.mu.Unlock()

	found, score := svc.IsUserInLookalike("user-1", "aud-1")

	if !found {
		t.Error("expected found")
	}
	if score != 0.8 {
		t.Errorf("expected score 0.8, got %f", score)
	}
}

func TestLA_IsUserInLookalike_UserNotInAudience(t *testing.T) {
	svc := NewLookalikeService(nil)

	svc.mu.Lock()
	svc.lookalikeAudiences["aud-1"] = &lookalikeAudience{
		ID: "aud-1",
		Users: []lookalikeUser{
			{UserID: "user-1", SimilarityScore: 0.8},
		},
	}
	svc.mu.Unlock()

	found, _ := svc.IsUserInLookalike("user-99", "aud-1")

	if found {
		t.Error("expected not found for user-99")
	}
}

func TestLA_GetLookalikeStats_Empty(t *testing.T) {
	svc := NewLookalikeService(nil)

	stats := svc.GetLookalikeStats()

	if stats["total_audiences"].(int) != 0 {
		t.Error("expected 0 audiences")
	}
}

func TestLA_GetLookalikeStats_WithData(t *testing.T) {
	svc := NewLookalikeService(nil)

	svc.mu.Lock()
	svc.userProfiles["u1"] = &userProfile{}
	svc.userProfiles["u2"] = &userProfile{}
	svc.lookalikeAudiences["a1"] = &lookalikeAudience{
		AudienceSize: 100,
		QualityScore: 0.8,
	}
	svc.mu.Unlock()

	stats := svc.GetLookalikeStats()

	if stats["total_audiences"].(int) != 1 {
		t.Error("expected 1 audience")
	}
	if stats["total_users_profiled"].(int) != 2 {
		t.Error("expected 2 profiled users")
	}
}

func TestLA_BuildSeedProfile_Empty(t *testing.T) {
	svc := NewLookalikeService(nil)

	profile := svc.buildSeedProfile([]string{})

	if profile != nil {
		t.Error("expected nil for empty seed")
	}
}

func TestLA_BuildSeedProfile_NoValidUsers(t *testing.T) {
	svc := NewLookalikeService(nil)

	profile := svc.buildSeedProfile([]string{"nonexistent-1", "nonexistent-2"})

	if profile != nil {
		t.Error("expected nil when no valid users")
	}
}

func TestLA_BuildSeedProfile_WithUsers(t *testing.T) {
	svc := NewLookalikeService(nil)

	// Add users
	for i := 0; i < 5; i++ {
		userID := "user-" + string(rune('a'+i))
		profile := svc.CreateUserProfile(
			nil,
			[]string{"tech"},
			[]string{"mobile"},
			"US", "", "",
		)
		profile.BehaviorMetrics.TotalImpressions = 100
		profile.BehaviorMetrics.TotalClicks = 5
		svc.RegisterUserProfile(userID, profile)
	}

	seedProfile := svc.buildSeedProfile([]string{"user-a", "user-b", "user-c"})

	if seedProfile == nil {
		t.Fatal("expected profile")
	}
	if seedProfile.UserCount != 3 {
		t.Errorf("expected 3 users, got %d", seedProfile.UserCount)
	}
}

func TestLA_CosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
	}{
		{"identical", []float64{1, 0, 0}, []float64{1, 0, 0}, 1.0},
		{"orthogonal", []float64{1, 0, 0}, []float64{0, 1, 0}, 0.0},
		{"opposite", []float64{1, 0}, []float64{-1, 0}, -1.0},
		{"diff_length", []float64{1, 0}, []float64{1, 0, 0}, 0.0},
		{"zero_vector", []float64{0, 0}, []float64{1, 0}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cosineSimilarity(tt.a, tt.b)
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("expected ~%f, got %f", tt.expected, result)
			}
		})
	}
}

func TestLA_TopN(t *testing.T) {
	counts := map[string]int{
		"a": 5,
		"b": 10,
		"c": 3,
		"d": 8,
	}

	result := topN(counts, 2)

	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	if result[0] != "b" {
		t.Error("expected 'b' first")
	}
}

func TestLA_TopN_EmptyMap(t *testing.T) {
	result := topN(map[string]int{}, 5)

	if len(result) != 0 {
		t.Error("expected empty result")
	}
}

func TestLA_TopN_LessItems(t *testing.T) {
	counts := map[string]int{"a": 5}

	result := topN(counts, 10)

	if len(result) != 1 {
		t.Errorf("expected 1 item, got %d", len(result))
	}
}

func TestLA_GenerateAudienceID(t *testing.T) {
	id := generateAudienceID("test")

	if id == "" {
		t.Error("expected non-empty ID")
	}
	if len(id) < 10 {
		t.Error("expected longer ID")
	}
}

func TestLA_CalculateAudienceQuality_Empty(t *testing.T) {
	svc := NewLookalikeService(nil)

	quality := svc.calculateAudienceQuality(nil, []lookalikeUser{})

	if quality != 0 {
		t.Error("expected 0 for empty")
	}
}

func TestLA_CalculateAudienceQuality_WithUsers(t *testing.T) {
	svc := NewLookalikeService(nil)

	seedProfile := &segmentProfile{UserCount: 10}
	users := []lookalikeUser{
		{SimilarityScore: 0.8, MatchedFeatures: []string{"a", "b"}},
		{SimilarityScore: 0.7, MatchedFeatures: []string{"a"}},
	}

	quality := svc.calculateAudienceQuality(seedProfile, users)

	if quality <= 0 || quality > 1 {
		t.Errorf("expected quality 0-1, got %f", quality)
	}
}

func TestLA_CalculateFeatureOverlap(t *testing.T) {
	svc := NewLookalikeService(nil)

	seedProfile := &segmentProfile{}
	users := []lookalikeUser{
		{MatchedFeatures: []string{"a", "b"}},
		{MatchedFeatures: []string{"a", "c"}},
	}

	overlap := svc.calculateFeatureOverlap(seedProfile, users)

	if overlap["a"] != 1.0 {
		t.Errorf("expected 'a' overlap 1.0, got %f", overlap["a"])
	}
	if overlap["b"] != 0.5 {
		t.Errorf("expected 'b' overlap 0.5, got %f", overlap["b"])
	}
}

func TestLA_AddToFeatureVector(t *testing.T) {
	svc := NewLookalikeService(nil)

	user := &userProfile{
		Interests:   []string{"sports", "tech"},
		DeviceTypes: []string{"mobile"},
		BehaviorMetrics: behaviorMetrics{
			TotalImpressions: 100,
			TotalClicks:      10,
			EngagementScore:  0.5,
		},
	}

	vector := make([]float64, 50)
	svc.addToFeatureVector(vector, user)

	// Sports is index 0, tech is index 1
	if vector[0] != 1.0 {
		t.Error("expected sports feature")
	}
	if vector[1] != 1.0 {
		t.Error("expected tech feature")
	}
	// Mobile is index 30
	if vector[30] != 1.0 {
		t.Error("expected mobile feature")
	}
}

func TestLA_FindMatchedFeatures(t *testing.T) {
	svc := NewLookalikeService(nil)

	seedProfile := &segmentProfile{
		TopInterests: []string{"tech", "sports"},
		TopGeos:      []string{"US"},
		TopDevices:   []string{"mobile"},
	}

	user := &userProfile{
		Interests:   []string{"tech", "food"},
		DeviceTypes: []string{"mobile", "desktop"},
		Geo:         geoProfile{Country: "US"},
		BehaviorMetrics: behaviorMetrics{
			EngagementScore: 0.6,
		},
	}

	matched := svc.findMatchedFeatures(seedProfile, user)

	hasInterest := false
	hasGeo := false
	hasDevice := false
	for _, m := range matched {
		if m == "interest:tech" {
			hasInterest = true
		}
		if m == "geo:US" {
			hasGeo = true
		}
		if m == "device:mobile" {
			hasDevice = true
		}
	}

	if !hasInterest {
		t.Error("expected interest match")
	}
	if !hasGeo {
		t.Error("expected geo match")
	}
	if !hasDevice {
		t.Error("expected device match")
	}
}

func TestLA_Concurrency(t *testing.T) {
	svc := NewLookalikeService(nil)
	svc.config.MinSeedSize = 1

	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userID := "conc-" + string(rune(idx))
			profile := svc.CreateUserProfile(nil, []string{"tech"}, []string{"mobile"}, "US", "", "")
			svc.RegisterUserProfile(userID, profile)
			svc.UpdateUserBehavior(userID, 10, 1, 0, 5.0)
			svc.GetLookalikeStats()
		}(i)
	}
	wg.Wait()
}
