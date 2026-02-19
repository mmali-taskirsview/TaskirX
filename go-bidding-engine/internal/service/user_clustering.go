package service

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
)

// UserClusteringService provides ML-based user segmentation and clustering
type UserClusteringService struct {
	cacheClient cache.Cache
	mu          sync.RWMutex

	// User data for clustering
	users map[string]*clusterUser

	// Cluster definitions
	clusters map[string]*userCluster

	// Configuration
	config *ClusteringConfig

	// K-means state
	centroids [][]float64
}

// ClusteringConfig holds configuration for user clustering
type ClusteringConfig struct {
	Enabled            bool
	NumClusters        int
	MinClusterSize     int
	MaxIterations      int
	ConvergenceThresh  float64
	FeatureDimensions  int
	RefreshIntervalHrs int
}

type clusterUser struct {
	UserID          string
	FeatureVector   []float64
	ClusterID       string
	ClusterDistance float64
	LastUpdated     time.Time
}

type userCluster struct {
	ID          string
	Name        string
	Centroid    []float64
	UserCount   int
	AvgDistance float64
	Cohesion    float64 // Intra-cluster similarity
	Features    clusterFeatures
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type clusterFeatures struct {
	TopInterests    []string
	TopDevices      []string
	TopGeos         []string
	AvgEngagement   float64
	AvgRecency      float64
	BehaviorProfile string
	ValueSegment    string
}

// NewUserClusteringService creates a new user clustering service
func NewUserClusteringService(c cache.Cache) *UserClusteringService {
	return &UserClusteringService{
		cacheClient: c,
		users:       make(map[string]*clusterUser),
		clusters:    make(map[string]*userCluster),
		config: &ClusteringConfig{
			Enabled:            true,
			NumClusters:        10,
			MinClusterSize:     50,
			MaxIterations:      100,
			ConvergenceThresh:  0.001,
			FeatureDimensions:  30,
			RefreshIntervalHrs: 24,
		},
	}
}

// ClusteringResult represents the result of clustering operation
type ClusteringResult struct {
	NumClusters     int              `json:"num_clusters"`
	TotalUsers      int              `json:"total_users"`
	Iterations      int              `json:"iterations"`
	Converged       bool             `json:"converged"`
	SilhouetteScore float64          `json:"silhouette_score"`
	Clusters        []ClusterSummary `json:"clusters"`
	Status          string           `json:"status"`
}

// ClusterSummary provides summary info about a cluster
type ClusterSummary struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	UserCount       int      `json:"user_count"`
	Cohesion        float64  `json:"cohesion"`
	TopInterests    []string `json:"top_interests"`
	BehaviorProfile string   `json:"behavior_profile"`
	ValueSegment    string   `json:"value_segment"`
}

// RunClustering performs K-means clustering on all users
func (s *UserClusteringService) RunClustering() *ClusteringResult {
	if !s.config.Enabled {
		return &ClusteringResult{Status: "disabled"}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.users) < s.config.MinClusterSize*s.config.NumClusters {
		return &ClusteringResult{
			Status:     "insufficient_data",
			TotalUsers: len(s.users),
		}
	}

	// Extract feature vectors
	userIDs := make([]string, 0, len(s.users))
	featureVectors := make([][]float64, 0, len(s.users))

	for id, user := range s.users {
		if len(user.FeatureVector) == s.config.FeatureDimensions {
			userIDs = append(userIDs, id)
			featureVectors = append(featureVectors, user.FeatureVector)
		}
	}

	if len(featureVectors) < s.config.MinClusterSize*s.config.NumClusters {
		return &ClusteringResult{
			Status:     "insufficient_valid_users",
			TotalUsers: len(featureVectors),
		}
	}

	// Run K-means
	centroids, assignments, iterations, converged := s.kMeans(featureVectors, s.config.NumClusters)
	s.centroids = centroids

	// Build clusters
	clusterUsers := make(map[int][]string)
	for i, clusterIdx := range assignments {
		clusterUsers[clusterIdx] = append(clusterUsers[clusterIdx], userIDs[i])
	}

	// Create cluster objects
	s.clusters = make(map[string]*userCluster)
	summaries := make([]ClusterSummary, 0)

	for clusterIdx, centroid := range centroids {
		users := clusterUsers[clusterIdx]
		if len(users) < s.config.MinClusterSize {
			continue
		}

		clusterID := generateClusterID(clusterIdx)
		features := s.extractClusterFeatures(users, centroid)
		cohesion := s.calculateCohesion(users, centroid)

		cluster := &userCluster{
			ID:        clusterID,
			Name:      s.generateClusterName(features),
			Centroid:  centroid,
			UserCount: len(users),
			Cohesion:  cohesion,
			Features:  features,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		s.clusters[clusterID] = cluster

		// Update user assignments
		for _, userID := range users {
			if user, exists := s.users[userID]; exists {
				user.ClusterID = clusterID
				user.ClusterDistance = s.euclideanDistance(user.FeatureVector, centroid)
			}
		}

		summaries = append(summaries, ClusterSummary{
			ID:              clusterID,
			Name:            cluster.Name,
			UserCount:       cluster.UserCount,
			Cohesion:        cohesion,
			TopInterests:    features.TopInterests,
			BehaviorProfile: features.BehaviorProfile,
			ValueSegment:    features.ValueSegment,
		})
	}

	// Calculate silhouette score
	silhouette := s.calculateSilhouetteScore(featureVectors, assignments)

	return &ClusteringResult{
		NumClusters:     len(s.clusters),
		TotalUsers:      len(featureVectors),
		Iterations:      iterations,
		Converged:       converged,
		SilhouetteScore: silhouette,
		Clusters:        summaries,
		Status:          "completed",
	}
}

// kMeans implements K-means clustering algorithm
func (s *UserClusteringService) kMeans(data [][]float64, k int) ([][]float64, []int, int, bool) {
	n := len(data)
	dim := len(data[0])

	// Initialize centroids using K-means++ initialization
	centroids := s.kMeansPlusPlusInit(data, k)

	assignments := make([]int, n)
	iterations := 0
	converged := false

	for iter := 0; iter < s.config.MaxIterations; iter++ {
		iterations++

		// Assignment step: assign each point to nearest centroid
		for i, point := range data {
			minDist := math.MaxFloat64
			minIdx := 0
			for j, centroid := range centroids {
				dist := s.euclideanDistance(point, centroid)
				if dist < minDist {
					minDist = dist
					minIdx = j
				}
			}
			assignments[i] = minIdx
		}

		// Update step: recalculate centroids
		newCentroids := make([][]float64, k)
		counts := make([]int, k)

		for j := 0; j < k; j++ {
			newCentroids[j] = make([]float64, dim)
		}

		for i, point := range data {
			cluster := assignments[i]
			counts[cluster]++
			for d := 0; d < dim; d++ {
				newCentroids[cluster][d] += point[d]
			}
		}

		// Normalize centroids
		for j := 0; j < k; j++ {
			if counts[j] > 0 {
				for d := 0; d < dim; d++ {
					newCentroids[j][d] /= float64(counts[j])
				}
			}
		}

		// Check convergence
		maxShift := 0.0
		for j := 0; j < k; j++ {
			shift := s.euclideanDistance(centroids[j], newCentroids[j])
			if shift > maxShift {
				maxShift = shift
			}
		}

		centroids = newCentroids

		if maxShift < s.config.ConvergenceThresh {
			converged = true
			break
		}
	}

	return centroids, assignments, iterations, converged
}

// kMeansPlusPlusInit implements K-means++ initialization
func (s *UserClusteringService) kMeansPlusPlusInit(data [][]float64, k int) [][]float64 {
	n := len(data)
	dim := len(data[0])

	centroids := make([][]float64, k)

	// Choose first centroid randomly
	firstIdx := rand.Intn(n)
	centroids[0] = make([]float64, dim)
	copy(centroids[0], data[firstIdx])

	// Choose remaining centroids
	for i := 1; i < k; i++ {
		// Calculate distances to nearest centroid
		distances := make([]float64, n)
		totalDist := 0.0

		for j, point := range data {
			minDist := math.MaxFloat64
			for c := 0; c < i; c++ {
				dist := s.euclideanDistance(point, centroids[c])
				if dist < minDist {
					minDist = dist
				}
			}
			distances[j] = minDist * minDist // Square distance for probability
			totalDist += distances[j]
		}

		// Choose next centroid with probability proportional to distance^2
		target := rand.Float64() * totalDist
		cumDist := 0.0
		chosenIdx := 0
		for j, dist := range distances {
			cumDist += dist
			if cumDist >= target {
				chosenIdx = j
				break
			}
		}

		centroids[i] = make([]float64, dim)
		copy(centroids[i], data[chosenIdx])
	}

	return centroids
}

func (s *UserClusteringService) euclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.MaxFloat64
	}

	sum := 0.0
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

func (s *UserClusteringService) extractClusterFeatures(userIDs []string, centroid []float64) clusterFeatures {
	features := clusterFeatures{}

	interestCounts := make(map[string]int)
	deviceCounts := make(map[string]int)
	geoCounts := make(map[string]int)
	totalEngagement := 0.0
	totalRecency := 0.0

	// Interest feature indices (0-9)
	interestNames := []string{"sports", "tech", "finance", "travel", "food",
		"fashion", "auto", "health", "entertainment", "gaming"}

	// Determine top interests from centroid
	for i, val := range centroid[:10] {
		if val > 0.3 {
			interestCounts[interestNames[i]] = int(val * 100)
		}
	}

	// Determine behavior profile from centroid
	engagementIdx := 15 // Index for engagement score in centroid
	recencyIdx := 16    // Index for recency score in centroid

	if engagementIdx < len(centroid) {
		totalEngagement = centroid[engagementIdx]
	}
	if recencyIdx < len(centroid) {
		totalRecency = centroid[recencyIdx]
	}

	// Device profile from centroid (indices 20-24)
	deviceNames := []string{"mobile", "desktop", "tablet", "ctv", "other"}
	for i := 0; i < 5 && 20+i < len(centroid); i++ {
		if centroid[20+i] > 0.2 {
			deviceCounts[deviceNames[i]] = int(centroid[20+i] * 100)
		}
	}

	features.TopInterests = topNStrings(interestCounts, 5)
	features.TopDevices = topNStrings(deviceCounts, 3)
	features.TopGeos = topNStrings(geoCounts, 3)
	features.AvgEngagement = totalEngagement
	features.AvgRecency = totalRecency

	// Determine behavior profile
	if totalEngagement > 0.7 {
		features.BehaviorProfile = "high_engagement"
	} else if totalEngagement > 0.4 {
		features.BehaviorProfile = "moderate_engagement"
	} else {
		features.BehaviorProfile = "low_engagement"
	}

	// Determine value segment
	if totalEngagement > 0.6 && totalRecency > 0.7 {
		features.ValueSegment = "high_value"
	} else if totalEngagement > 0.3 || totalRecency > 0.5 {
		features.ValueSegment = "medium_value"
	} else {
		features.ValueSegment = "low_value"
	}

	return features
}

func (s *UserClusteringService) calculateCohesion(userIDs []string, centroid []float64) float64 {
	if len(userIDs) == 0 {
		return 0.0
	}

	totalDist := 0.0
	count := 0

	for _, userID := range userIDs {
		if user, exists := s.users[userID]; exists {
			totalDist += s.euclideanDistance(user.FeatureVector, centroid)
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	avgDist := totalDist / float64(count)
	// Convert distance to cohesion (lower distance = higher cohesion)
	return 1.0 / (1.0 + avgDist)
}

func (s *UserClusteringService) calculateSilhouetteScore(data [][]float64, assignments []int) float64 {
	n := len(data)
	if n < 2 {
		return 0.0
	}

	silhouetteSum := 0.0

	for i := 0; i < n; i++ {
		// Calculate a(i) - average distance to same cluster
		sameCluster := 0.0
		sameCount := 0
		for j := 0; j < n; j++ {
			if i != j && assignments[i] == assignments[j] {
				sameCluster += s.euclideanDistance(data[i], data[j])
				sameCount++
			}
		}
		a := 0.0
		if sameCount > 0 {
			a = sameCluster / float64(sameCount)
		}

		// Calculate b(i) - minimum average distance to other clusters
		b := math.MaxFloat64
		clusterDists := make(map[int][]float64)
		for j := 0; j < n; j++ {
			if assignments[i] != assignments[j] {
				clusterDists[assignments[j]] = append(clusterDists[assignments[j]], s.euclideanDistance(data[i], data[j]))
			}
		}
		for _, dists := range clusterDists {
			avgDist := 0.0
			for _, d := range dists {
				avgDist += d
			}
			avgDist /= float64(len(dists))
			if avgDist < b {
				b = avgDist
			}
		}

		// Calculate silhouette for this point
		if b == math.MaxFloat64 {
			b = 0
		}
		silhouette := 0.0
		if math.Max(a, b) > 0 {
			silhouette = (b - a) / math.Max(a, b)
		}
		silhouetteSum += silhouette
	}

	return silhouetteSum / float64(n)
}

func (s *UserClusteringService) generateClusterName(features clusterFeatures) string {
	name := features.ValueSegment + "_"
	if len(features.TopInterests) > 0 {
		name += features.TopInterests[0] + "_"
	}
	name += features.BehaviorProfile
	return name
}

// RegisterUser registers a user with their feature vector
func (s *UserClusteringService) RegisterUser(userID string, featureVector []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[userID] = &clusterUser{
		UserID:        userID,
		FeatureVector: featureVector,
		LastUpdated:   time.Now(),
	}
}

// BuildUserFeatures builds a feature vector from user attributes
func (s *UserClusteringService) BuildUserFeatures(
	interests []string,
	deviceType string,
	geo string,
	impressions, clicks, conversions int64,
	engagement, recency float64,
) []float64 {
	vector := make([]float64, s.config.FeatureDimensions)

	// Interest features (0-9)
	interestMap := map[string]int{
		"sports": 0, "tech": 1, "finance": 2, "travel": 3, "food": 4,
		"fashion": 5, "auto": 6, "health": 7, "entertainment": 8, "gaming": 9,
	}
	for _, interest := range interests {
		if idx, ok := interestMap[interest]; ok {
			vector[idx] = 1.0
		}
	}

	// Behavior features (10-14)
	vector[10] = math.Min(math.Log1p(float64(impressions))/10.0, 1.0)
	vector[11] = math.Min(math.Log1p(float64(clicks))/5.0, 1.0)
	vector[12] = math.Min(math.Log1p(float64(conversions))/3.0, 1.0)

	// CTR/CVR
	if impressions > 0 {
		vector[13] = math.Min(float64(clicks)/float64(impressions)*20.0, 1.0)
	}
	if clicks > 0 {
		vector[14] = math.Min(float64(conversions)/float64(clicks)*10.0, 1.0)
	}

	// Engagement and recency (15-16)
	vector[15] = engagement
	vector[16] = recency

	// Device features (20-24)
	deviceMap := map[string]int{"mobile": 20, "desktop": 21, "tablet": 22, "ctv": 23, "other": 24}
	if idx, ok := deviceMap[deviceType]; ok {
		vector[idx] = 1.0
	}

	return vector
}

// GetUserCluster returns the cluster assignment for a user
func (s *UserClusteringService) GetUserCluster(userID string) (*userCluster, float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if user, exists := s.users[userID]; exists {
		if cluster, cExists := s.clusters[user.ClusterID]; cExists {
			return cluster, user.ClusterDistance
		}
	}
	return nil, 0.0
}

// GetClusterUsers returns users in a specific cluster
func (s *UserClusteringService) GetClusterUsers(clusterID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]string, 0)
	for userID, user := range s.users {
		if user.ClusterID == clusterID {
			users = append(users, userID)
		}
	}
	return users
}

// GetClusteringStats returns statistics about clustering
func (s *UserClusteringService) GetClusteringStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clusterSizes := make(map[string]int)
	totalCohesion := 0.0

	for id, cluster := range s.clusters {
		clusterSizes[id] = cluster.UserCount
		totalCohesion += cluster.Cohesion
	}

	avgCohesion := 0.0
	if len(s.clusters) > 0 {
		avgCohesion = totalCohesion / float64(len(s.clusters))
	}

	return map[string]interface{}{
		"total_users":   len(s.users),
		"num_clusters":  len(s.clusters),
		"cluster_sizes": clusterSizes,
		"avg_cohesion":  avgCohesion,
		"config":        s.config,
	}
}

// Helper functions

func topNStrings(counts map[string]int, n int) []string {
	type kv struct {
		Key   string
		Value int
	}

	pairs := make([]kv, 0, len(counts))
	for k, v := range counts {
		pairs = append(pairs, kv{k, v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[j].Value
	})

	result := make([]string, 0, n)
	for i := 0; i < n && i < len(pairs); i++ {
		result = append(result, pairs[i].Key)
	}
	return result
}

func generateClusterID(idx int) string {
	return "cluster_" + time.Now().Format("20060102") + "_" + string(rune('A'+idx))
}
