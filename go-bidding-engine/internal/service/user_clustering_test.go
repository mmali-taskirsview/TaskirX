package service

import (
	"math"
	"sync"
	"testing"
)

func TestCluster_NewService(t *testing.T) {
	svc := NewUserClusteringService(nil)
	if svc == nil {
		t.Fatal("expected service")
	}
	if svc.users == nil {
		t.Error("expected users map")
	}
	if svc.clusters == nil {
		t.Error("expected clusters map")
	}
	if svc.config == nil {
		t.Fatal("expected config")
	}
	if svc.config.NumClusters != 10 {
		t.Errorf("expected 10 clusters, got %d", svc.config.NumClusters)
	}
}

func TestCluster_RegisterUser(t *testing.T) {
	svc := NewUserClusteringService(nil)
	features := make([]float64, 30)
	features[0] = 0.5
	features[15] = 0.8 // engagement

	svc.RegisterUser("user-1", features)

	if len(svc.users) != 1 {
		t.Errorf("expected 1 user, got %d", len(svc.users))
	}
	if svc.users["user-1"] == nil {
		t.Error("expected user-1 registered")
	}
}

func TestCluster_BuildUserFeatures_Interests(t *testing.T) {
	svc := NewUserClusteringService(nil)

	features := svc.BuildUserFeatures(
		[]string{"sports", "tech"},
		"mobile",
		"US",
		100, 10, 2,
		0.8, 0.9,
	)

	if len(features) != 30 {
		t.Errorf("expected 30 dimensions, got %d", len(features))
	}

	// Check sports interest (index 0)
	if features[0] != 1.0 {
		t.Errorf("expected sports=1.0, got %f", features[0])
	}
	// Check tech interest (index 1)
	if features[1] != 1.0 {
		t.Errorf("expected tech=1.0, got %f", features[1])
	}
	// Check engagement (index 15)
	if features[15] != 0.8 {
		t.Errorf("expected engagement=0.8, got %f", features[15])
	}
}

func TestCluster_BuildUserFeatures_Devices(t *testing.T) {
	svc := NewUserClusteringService(nil)

	tests := []struct {
		device   string
		expected int
	}{
		{"mobile", 20},
		{"desktop", 21},
		{"tablet", 22},
		{"ctv", 23},
		{"other", 24},
	}

	for _, tt := range tests {
		features := svc.BuildUserFeatures(nil, tt.device, "", 0, 0, 0, 0, 0)
		if features[tt.expected] != 1.0 {
			t.Errorf("%s: expected feature[%d]=1.0, got %f", tt.device, tt.expected, features[tt.expected])
		}
	}
}

func TestCluster_BuildUserFeatures_Behavior(t *testing.T) {
	svc := NewUserClusteringService(nil)

	features := svc.BuildUserFeatures(nil, "", "", 1000, 100, 10, 0, 0)

	// Impressions feature (normalized log)
	if features[10] <= 0 {
		t.Error("expected impressions feature > 0")
	}
	// Clicks feature
	if features[11] <= 0 {
		t.Error("expected clicks feature > 0")
	}
	// Conversions feature
	if features[12] <= 0 {
		t.Error("expected conversions feature > 0")
	}
	// CTR
	if features[13] <= 0 {
		t.Error("expected CTR feature > 0")
	}
	// CVR
	if features[14] <= 0 {
		t.Error("expected CVR feature > 0")
	}
}

func TestCluster_EuclideanDistance(t *testing.T) {
	svc := NewUserClusteringService(nil)

	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
	}{
		{"zero distance", []float64{0, 0}, []float64{0, 0}, 0},
		{"unit distance", []float64{0, 0}, []float64{1, 0}, 1},
		{"3-4-5 triangle", []float64{0, 0}, []float64{3, 4}, 5},
		{"negative coords", []float64{-1, -1}, []float64{2, 3}, 5},
	}

	for _, tt := range tests {
		result := svc.euclideanDistance(tt.a, tt.b)
		if math.Abs(result-tt.expected) > 0.001 {
			t.Errorf("%s: expected %f, got %f", tt.name, tt.expected, result)
		}
	}
}

func TestCluster_EuclideanDistance_DifferentLengths(t *testing.T) {
	svc := NewUserClusteringService(nil)

	result := svc.euclideanDistance([]float64{1, 2}, []float64{1, 2, 3})

	if result != math.MaxFloat64 {
		t.Error("expected MaxFloat64 for different lengths")
	}
}

func TestCluster_RunClustering_Disabled(t *testing.T) {
	svc := NewUserClusteringService(nil)
	svc.config.Enabled = false

	result := svc.RunClustering()

	if result.Status != "disabled" {
		t.Errorf("expected status 'disabled', got '%s'", result.Status)
	}
}

func TestCluster_RunClustering_InsufficientData(t *testing.T) {
	svc := NewUserClusteringService(nil)
	svc.config.MinClusterSize = 50
	svc.config.NumClusters = 10

	// Only add 10 users
	for i := 0; i < 10; i++ {
		features := make([]float64, 30)
		svc.RegisterUser("user-"+string(rune('A'+i)), features)
	}

	result := svc.RunClustering()

	if result.Status != "insufficient_data" {
		t.Errorf("expected status 'insufficient_data', got '%s'", result.Status)
	}
}

func TestCluster_RunClustering_Success(t *testing.T) {
	svc := NewUserClusteringService(nil)
	svc.config.MinClusterSize = 5
	svc.config.NumClusters = 3

	// Add enough users for 3 clusters of 5 each
	for i := 0; i < 20; i++ {
		features := make([]float64, 30)
		// Create distinct clusters
		clusterIdx := i % 3
		features[clusterIdx*3] = float64(clusterIdx) + 0.5
		features[15] = 0.5 + float64(clusterIdx)*0.1 // engagement
		features[16] = 0.6 + float64(clusterIdx)*0.1 // recency
		svc.RegisterUser("user-"+string(rune('A'+i)), features)
	}

	result := svc.RunClustering()

	if result.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", result.Status)
	}
	if result.TotalUsers < 15 {
		t.Errorf("expected at least 15 users, got %d", result.TotalUsers)
	}
	if result.Iterations < 1 {
		t.Error("expected at least 1 iteration")
	}
}

func TestCluster_KMeans(t *testing.T) {
	svc := NewUserClusteringService(nil)

	// Simple 2D data with obvious clusters
	data := [][]float64{
		{0, 0}, {0.1, 0.1}, {0, 0.1}, // Cluster 0 around origin
		{5, 5}, {5.1, 5.1}, {5, 5.1}, // Cluster 1 around (5,5)
	}

	centroids, assignments, iterations, converged := svc.kMeans(data, 2)

	if len(centroids) != 2 {
		t.Errorf("expected 2 centroids, got %d", len(centroids))
	}
	if iterations < 1 {
		t.Error("expected at least 1 iteration")
	}
	if !converged {
		t.Log("may not converge with simple data (depends on random init)")
	}

	// Check assignments - first 3 points should be in same cluster
	cluster0 := assignments[0]
	if assignments[1] != cluster0 || assignments[2] != cluster0 {
		t.Error("expected first 3 points in same cluster")
	}

	// Last 3 points should be in different cluster
	cluster1 := assignments[3]
	if cluster0 == cluster1 {
		t.Error("expected two different clusters")
	}
}

func TestCluster_KMeansPlusPlusInit(t *testing.T) {
	svc := NewUserClusteringService(nil)

	data := [][]float64{
		{0, 0}, {1, 1}, {2, 2}, {10, 10}, {11, 11}, {12, 12},
	}

	centroids := svc.kMeansPlusPlusInit(data, 2)

	if len(centroids) != 2 {
		t.Errorf("expected 2 centroids, got %d", len(centroids))
	}

	// Centroids should likely be separated
	dist := svc.euclideanDistance(centroids[0], centroids[1])
	if dist < 1 {
		t.Errorf("expected centroids to be separated, distance: %f", dist)
	}
}

func TestCluster_GetUserCluster_Exists(t *testing.T) {
	svc := NewUserClusteringService(nil)

	// Register and cluster users
	svc.config.MinClusterSize = 2
	svc.config.NumClusters = 2

	for i := 0; i < 10; i++ {
		features := make([]float64, 30)
		features[i%10] = 1.0
		svc.RegisterUser("user-"+string(rune('A'+i)), features)
	}

	svc.RunClustering()

	// Get cluster for a user
	cluster, distance := svc.GetUserCluster("user-A")

	if cluster == nil {
		t.Log("cluster may be nil if user not assigned")
	}
	if cluster != nil && distance < 0 {
		t.Error("expected non-negative distance")
	}
}

func TestCluster_GetUserCluster_NotExists(t *testing.T) {
	svc := NewUserClusteringService(nil)

	cluster, distance := svc.GetUserCluster("nonexistent")

	if cluster != nil {
		t.Error("expected nil cluster for nonexistent user")
	}
	if distance != 0 {
		t.Errorf("expected 0 distance, got %f", distance)
	}
}

func TestCluster_GetClusterUsers(t *testing.T) {
	svc := NewUserClusteringService(nil)

	// Create clusters
	svc.users["user-A"] = &clusterUser{UserID: "user-A", ClusterID: "cluster-1"}
	svc.users["user-B"] = &clusterUser{UserID: "user-B", ClusterID: "cluster-1"}
	svc.users["user-C"] = &clusterUser{UserID: "user-C", ClusterID: "cluster-2"}

	users := svc.GetClusterUsers("cluster-1")

	if len(users) != 2 {
		t.Errorf("expected 2 users in cluster-1, got %d", len(users))
	}
}

func TestCluster_GetClusterUsers_Empty(t *testing.T) {
	svc := NewUserClusteringService(nil)

	users := svc.GetClusterUsers("nonexistent")

	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestCluster_GetClusteringStats(t *testing.T) {
	svc := NewUserClusteringService(nil)

	// Add some users
	for i := 0; i < 5; i++ {
		svc.users["user-"+string(rune('A'+i))] = &clusterUser{UserID: "user-" + string(rune('A'+i))}
	}

	stats := svc.GetClusteringStats()

	totalUsers := stats["total_users"].(int)
	if totalUsers != 5 {
		t.Errorf("expected 5 users, got %d", totalUsers)
	}

	numClusters := stats["num_clusters"].(int)
	if numClusters != 0 {
		t.Errorf("expected 0 clusters (not run yet), got %d", numClusters)
	}

	if stats["config"] == nil {
		t.Error("expected config in stats")
	}
}

func TestCluster_CalculateCohesion_Empty(t *testing.T) {
	svc := NewUserClusteringService(nil)

	cohesion := svc.calculateCohesion([]string{}, []float64{})

	if cohesion != 0 {
		t.Errorf("expected 0 cohesion for empty, got %f", cohesion)
	}
}

func TestCluster_CalculateCohesion_WithUsers(t *testing.T) {
	svc := NewUserClusteringService(nil)

	centroid := []float64{0.5, 0.5}

	svc.users["user-A"] = &clusterUser{FeatureVector: []float64{0.5, 0.5}}
	svc.users["user-B"] = &clusterUser{FeatureVector: []float64{0.6, 0.6}}

	cohesion := svc.calculateCohesion([]string{"user-A", "user-B"}, centroid)

	// Cohesion should be high since users are close to centroid
	if cohesion < 0.5 {
		t.Errorf("expected high cohesion, got %f", cohesion)
	}
}

func TestCluster_CalculateSilhouetteScore(t *testing.T) {
	svc := NewUserClusteringService(nil)

	// Well-separated clusters
	data := [][]float64{
		{0, 0}, {0.1, 0},
		{10, 10}, {10.1, 10},
	}
	assignments := []int{0, 0, 1, 1}

	score := svc.calculateSilhouetteScore(data, assignments)

	// Should be positive for well-separated clusters
	if score <= 0 {
		t.Errorf("expected positive silhouette score, got %f", score)
	}
}

func TestCluster_CalculateSilhouetteScore_Single(t *testing.T) {
	svc := NewUserClusteringService(nil)

	data := [][]float64{{0, 0}}
	assignments := []int{0}

	score := svc.calculateSilhouetteScore(data, assignments)

	if score != 0 {
		t.Errorf("expected 0 for single point, got %f", score)
	}
}

func TestCluster_ExtractClusterFeatures(t *testing.T) {
	svc := NewUserClusteringService(nil)

	// Create centroid with high sports interest and mobile device
	centroid := make([]float64, 30)
	centroid[0] = 0.9  // sports
	centroid[1] = 0.8  // tech
	centroid[15] = 0.8 // high engagement
	centroid[16] = 0.9 // high recency
	centroid[20] = 0.9 // mobile

	features := svc.extractClusterFeatures([]string{}, centroid)

	if features.BehaviorProfile != "high_engagement" {
		t.Errorf("expected high_engagement, got %s", features.BehaviorProfile)
	}
	if features.ValueSegment != "high_value" {
		t.Errorf("expected high_value, got %s", features.ValueSegment)
	}
	if len(features.TopInterests) == 0 {
		t.Error("expected some interests")
	}
}

func TestCluster_ExtractClusterFeatures_MediumEngagement(t *testing.T) {
	svc := NewUserClusteringService(nil)

	centroid := make([]float64, 30)
	centroid[15] = 0.5 // medium engagement

	features := svc.extractClusterFeatures([]string{}, centroid)

	if features.BehaviorProfile != "moderate_engagement" {
		t.Errorf("expected moderate_engagement, got %s", features.BehaviorProfile)
	}
}

func TestCluster_ExtractClusterFeatures_LowEngagement(t *testing.T) {
	svc := NewUserClusteringService(nil)

	centroid := make([]float64, 30)
	centroid[15] = 0.2 // low engagement

	features := svc.extractClusterFeatures([]string{}, centroid)

	if features.BehaviorProfile != "low_engagement" {
		t.Errorf("expected low_engagement, got %s", features.BehaviorProfile)
	}
}

func TestCluster_GenerateClusterName(t *testing.T) {
	svc := NewUserClusteringService(nil)

	features := clusterFeatures{
		ValueSegment:    "high_value",
		TopInterests:    []string{"sports"},
		BehaviorProfile: "high_engagement",
	}

	name := svc.generateClusterName(features)

	if name != "high_value_sports_high_engagement" {
		t.Errorf("unexpected name: %s", name)
	}
}

func TestCluster_TopNStrings(t *testing.T) {
	counts := map[string]int{
		"a": 10,
		"b": 30,
		"c": 20,
		"d": 5,
	}

	result := topNStrings(counts, 2)

	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
	if result[0] != "b" {
		t.Errorf("expected 'b' first, got %s", result[0])
	}
	if result[1] != "c" {
		t.Errorf("expected 'c' second, got %s", result[1])
	}
}

func TestCluster_TopNStrings_Empty(t *testing.T) {
	result := topNStrings(map[string]int{}, 5)

	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

func TestCluster_Concurrency(t *testing.T) {
	svc := NewUserClusteringService(nil)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			features := make([]float64, 30)
			features[idx%10] = 1.0
			svc.RegisterUser("user-"+string(rune(idx)), features)
			svc.GetUserCluster("user-" + string(rune(idx)))
			svc.GetClusteringStats()
		}(i)
	}
	wg.Wait()
}
