package service

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ========================================================================
// Dynamic Bid Service Tests
// ========================================================================

func TestDynamicBidService_NewService(t *testing.T) {
	cache := NewMockCache()
	svc := NewDynamicBidService(cache)

	if svc == nil {
		t.Fatal("Expected non-nil DynamicBidService")
	}
	if svc.config == nil {
		t.Fatal("Expected non-nil config")
	}
	if !svc.config.Enabled {
		t.Error("Expected service to be enabled by default")
	}
}

func TestDynamicBidService_CalculateDynamicBid(t *testing.T) {
	cache := NewMockCache()
	svc := NewDynamicBidService(cache)

	campaign := &model.Campaign{
		ID:       "camp-123",
		Name:     "Test Campaign",
		BidPrice: 2.50,
		GoalType: "CPA",
	}

	req := &model.BidRequest{
		ID:          "req-123",
		PublisherID: "pub-456",
		Device: model.InternalDevice{
			Type: "mobile",
		},
		AdSlot: model.AdSlot{
			ID:         "slot-1",
			Dimensions: []int{300, 250},
			Formats:    []string{"banner"},
		},
	}

	result := svc.CalculateDynamicBid(campaign, req)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.OriginalBid != 2.50 {
		t.Errorf("Expected original bid 2.50, got %f", result.OriginalBid)
	}
	if result.AdjustedBid <= 0 {
		t.Error("Expected positive adjusted bid")
	}
	if result.Multiplier <= 0 {
		t.Error("Expected positive multiplier")
	}
}

func TestDynamicBidService_RecordOutcome(t *testing.T) {
	cache := NewMockCache()
	svc := NewDynamicBidService(cache)

	req := &model.BidRequest{
		ID:          "req-123",
		PublisherID: "pub-456",
		Device:      model.InternalDevice{Type: "mobile"},
		AdSlot:      model.AdSlot{ID: "slot-1"},
	}

	svc.RecordOutcome("camp-123", req, 2.50, 2.00, true, true, false, 0.0)

	// Verify stats were updated
	svc.mu.RLock()
	stats := svc.publisherStats["pub-456"]
	svc.mu.RUnlock()

	if stats == nil {
		t.Fatal("Expected publisher stats to be created")
	}
}

func TestDynamicBidService_HourlyMultipliers(t *testing.T) {
	cache := NewMockCache()
	svc := NewDynamicBidService(cache)

	// Check that prime time hours have higher multipliers
	morningMultiplier := svc.hourlyMultipliers[3] // 3 AM
	primeMultiplier := svc.hourlyMultipliers[19]  // 7 PM

	if primeMultiplier <= morningMultiplier {
		t.Error("Expected prime time to have higher multiplier than early morning")
	}
}

func TestDynamicBidService_SetConfig(t *testing.T) {
	cache := NewMockCache()
	svc := NewDynamicBidService(cache)

	newConfig := &DynamicBidConfig{
		Enabled:          false,
		LearningRate:     0.05,
		MinBidMultiplier: 0.3,
		MaxBidMultiplier: 3.0,
	}

	svc.SetConfig(newConfig)

	if svc.config.Enabled {
		t.Error("Expected service to be disabled")
	}
}

func TestDynamicBidService_GetConfig(t *testing.T) {
	cache := NewMockCache()
	svc := NewDynamicBidService(cache)

	config := svc.GetConfig()
	if config == nil {
		t.Fatal("Expected non-nil config")
	}
}

func TestDynamicBidService_GetBidAnalytics(t *testing.T) {
	cache := NewMockCache()
	svc := NewDynamicBidService(cache)

	analytics := svc.GetBidAnalytics()
	if analytics == nil {
		t.Fatal("Expected non-nil analytics")
	}
}

// ========================================================================
// Lookalike Service Tests
// ========================================================================

func TestLookalikeService_NewService(t *testing.T) {
	cache := NewMockCache()
	svc := NewLookalikeService(cache)

	if svc == nil {
		t.Fatal("Expected non-nil LookalikeService")
	}
	if svc.config == nil {
		t.Fatal("Expected non-nil config")
	}
	if !svc.config.Enabled {
		t.Error("Expected service to be enabled by default")
	}
}

func TestLookalikeService_RegisterUserProfile(t *testing.T) {
	cache := NewMockCache()
	svc := NewLookalikeService(cache)

	profile := &userProfile{
		UserID:   "user-123",
		Segments: []string{"sports", "technology"},
		LastSeen: time.Now(),
	}

	svc.RegisterUserProfile("user-123", profile)

	svc.mu.RLock()
	_, exists := svc.userProfiles["user-123"]
	svc.mu.RUnlock()

	if !exists {
		t.Error("Expected user profile to be added")
	}
}

func TestLookalikeService_GenerateLookalike_InsufficientSeed(t *testing.T) {
	cache := NewMockCache()
	svc := NewLookalikeService(cache)

	seedUsers := []string{"user-1", "user-2"}
	result := svc.GenerateLookalike(seedUsers, "test-audience", 2.0)

	if result.Status != "insufficient_seed" {
		t.Errorf("Expected 'insufficient_seed' status, got '%s'", result.Status)
	}
}

func TestLookalikeService_GenerateLookalike_Disabled(t *testing.T) {
	cache := NewMockCache()
	svc := NewLookalikeService(cache)
	svc.config.Enabled = false

	seedUsers := []string{"user-1", "user-2"}
	result := svc.GenerateLookalike(seedUsers, "test-audience", 2.0)

	if result.Status != "disabled" {
		t.Errorf("Expected 'disabled' status, got '%s'", result.Status)
	}
}

func TestLookalikeService_GetLookalikeAudience(t *testing.T) {
	cache := NewMockCache()
	svc := NewLookalikeService(cache)

	svc.mu.Lock()
	svc.lookalikeAudiences["test-aud"] = &lookalikeAudience{
		ID:             "test-aud",
		Name:           "Test Audience",
		SeedSize:       2,
		AudienceSize:   5,
		ExpansionRatio: 2.5,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour), // Set expiry in future
		QualityScore:   0.85,
	}
	svc.mu.Unlock()

	audience := svc.GetLookalikeAudience("test-aud")
	if audience == nil {
		t.Fatal("Expected non-nil audience")
	}

	if audience.ID != "test-aud" {
		t.Errorf("Expected audience ID test-aud, got %s", audience.ID)
	}
}

func TestLookalikeService_IsUserInLookalike(t *testing.T) {
	cache := NewMockCache()
	svc := NewLookalikeService(cache)

	svc.mu.Lock()
	svc.lookalikeAudiences["test-aud"] = &lookalikeAudience{
		ID: "test-aud",
		Users: []lookalikeUser{
			{UserID: "user-1", SimilarityScore: 0.95},
			{UserID: "user-2", SimilarityScore: 0.88},
			{UserID: "user-3", SimilarityScore: 0.75},
		},
	}
	svc.mu.Unlock()

	isMember, score := svc.IsUserInLookalike("user-2", "test-aud")
	if !isMember {
		t.Error("Expected user-2 to be a member")
	}
	if score <= 0 {
		t.Error("Expected positive score for member")
	}

	isMember, _ = svc.IsUserInLookalike("user-99", "test-aud")
	if isMember {
		t.Error("Expected user-99 to not be a member")
	}
}

func TestLookalikeService_GetLookalikeStats(t *testing.T) {
	cache := NewMockCache()
	svc := NewLookalikeService(cache)

	svc.mu.Lock()
	svc.lookalikeAudiences["aud-1"] = &lookalikeAudience{
		ID:            "aud-1",
		Name:          "Audience 1",
		SeedSegmentID: "seg-1",
		SeedSize:      2,
		Users:         []lookalikeUser{{UserID: "u3", SimilarityScore: 0.9}},
		CreatedAt:     time.Now(),
	}
	svc.mu.Unlock()

	stats := svc.GetLookalikeStats()
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}
}

// ========================================================================
// User Clustering Service Tests
// ========================================================================

func TestUserClusteringService_NewService(t *testing.T) {
	cache := NewMockCache()
	svc := NewUserClusteringService(cache)

	if svc == nil {
		t.Fatal("Expected non-nil UserClusteringService")
	}
	if svc.config == nil {
		t.Fatal("Expected non-nil config")
	}
	if !svc.config.Enabled {
		t.Error("Expected service to be enabled by default")
	}
}

func TestUserClusteringService_RegisterUser(t *testing.T) {
	cache := NewMockCache()
	svc := NewUserClusteringService(cache)

	featureVector := []float64{0.5, 0.6, 0.7, 0.8}
	svc.RegisterUser("user-123", featureVector)

	svc.mu.RLock()
	user, exists := svc.users["user-123"]
	svc.mu.RUnlock()

	if !exists {
		t.Error("Expected user to be added")
	}
	if len(user.FeatureVector) != 4 {
		t.Errorf("Expected 4 features, got %d", len(user.FeatureVector))
	}
}

func TestUserClusteringService_RunClustering(t *testing.T) {
	cache := NewMockCache()
	svc := NewUserClusteringService(cache)

	// Use default feature dimensions (30) to match production behavior
	// Generate users with 30-dimensional feature vectors
	for i := 0; i < 100; i++ {
		group := float64(i / 25)
		featureVector := make([]float64, 30)
		// Set some interesting features for clustering
		featureVector[0] = group*0.25 + 0.1 // Interest weight
		featureVector[1] = group*0.20 + 0.2 // Another interest
		featureVector[2] = float64(i%10) * 0.1
		featureVector[15] = 0.5 + group*0.1 // Engagement score
		featureVector[16] = 0.6             // Recency
		featureVector[20] = 0.8             // Mobile device
		svc.RegisterUser(genUserID(i), featureVector)
	}

	svc.config.NumClusters = 4
	svc.config.MinClusterSize = 10
	svc.config.MaxIterations = 50

	result := svc.RunClustering()

	if result.TotalUsers != 100 {
		t.Errorf("Expected 100 total users, got %d", result.TotalUsers)
	}
	if result.Status != "completed" {
		t.Errorf("Expected completed status, got %s", result.Status)
	}
}

func genUserID(i int) string {
	return "user-" + string(rune('A'+i%26)) + string(rune('0'+i/26))
}

func TestUserClusteringService_GetUserCluster(t *testing.T) {
	cache := NewMockCache()
	svc := NewUserClusteringService(cache)

	svc.mu.Lock()
	svc.users["user-123"] = &clusterUser{
		UserID:        "user-123",
		FeatureVector: []float64{0.5, 0.5, 0.5},
		ClusterID:     "cluster-1",
	}
	svc.clusters["cluster-1"] = &userCluster{
		ID:        "cluster-1",
		Name:      "Test Cluster",
		Centroid:  []float64{0.5, 0.5, 0.5},
		UserCount: 1,
	}
	svc.mu.Unlock()

	cluster, confidence := svc.GetUserCluster("user-123")
	if cluster == nil {
		t.Fatal("Expected non-nil cluster")
	}

	if cluster.ID != "cluster-1" {
		t.Errorf("Expected cluster ID cluster-1, got %s", cluster.ID)
	}
	if confidence < 0 {
		t.Error("Expected non-negative confidence")
	}
}

func TestUserClusteringService_GetClusterUsers(t *testing.T) {
	cache := NewMockCache()
	svc := NewUserClusteringService(cache)

	svc.mu.Lock()
	svc.clusters["cluster-1"] = &userCluster{
		ID:        "cluster-1",
		Name:      "High Value Users",
		Centroid:  []float64{0.8, 0.9, 0.7},
		UserCount: 3,
	}
	// Add users assigned to cluster-1
	svc.users["user-1"] = &clusterUser{UserID: "user-1", ClusterID: "cluster-1"}
	svc.users["user-2"] = &clusterUser{UserID: "user-2", ClusterID: "cluster-1"}
	svc.users["user-3"] = &clusterUser{UserID: "user-3", ClusterID: "cluster-1"}
	svc.mu.Unlock()

	users := svc.GetClusterUsers("cluster-1")
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
}

func TestUserClusteringService_GetClusteringStats(t *testing.T) {
	cache := NewMockCache()
	svc := NewUserClusteringService(cache)

	svc.mu.Lock()
	svc.clusters["cluster-1"] = &userCluster{ID: "cluster-1", Name: "Cluster 1", UserCount: 2}
	svc.clusters["cluster-2"] = &userCluster{ID: "cluster-2", Name: "Cluster 2", UserCount: 1}
	svc.mu.Unlock()

	stats := svc.GetClusteringStats()
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}
}

func TestUserClusteringService_EuclideanDistance(t *testing.T) {
	cache := NewMockCache()
	svc := NewUserClusteringService(cache)

	vec1 := []float64{0, 0, 0}
	vec2 := []float64{3, 4, 0}

	distance := svc.euclideanDistance(vec1, vec2)
	expected := 5.0

	if distance != expected {
		t.Errorf("Expected distance %f, got %f", expected, distance)
	}
}

func TestUserClusteringService_RunClustering_InsufficientUsers(t *testing.T) {
	cache := NewMockCache()
	svc := NewUserClusteringService(cache)

	svc.RegisterUser("user-1", []float64{0.1, 0.2})
	svc.RegisterUser("user-2", []float64{0.3, 0.4})

	result := svc.RunClustering()

	if result.Status != "insufficient_data" {
		t.Errorf("Expected 'insufficient_data' status, got '%s'", result.Status)
	}
}

func TestUserClusteringService_RunClustering_Disabled(t *testing.T) {
	cache := NewMockCache()
	svc := NewUserClusteringService(cache)
	svc.config.Enabled = false

	result := svc.RunClustering()

	if result.Status != "disabled" {
		t.Errorf("Expected 'disabled' status, got '%s'", result.Status)
	}
}
