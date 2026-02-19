package service

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
)

// LookalikeService generates lookalike audiences based on seed users
type LookalikeService struct {
	cacheClient cache.Cache
	mu          sync.RWMutex

	// User profiles for similarity calculation
	userProfiles map[string]*userProfile

	// Segment profiles (aggregated user traits)
	segmentProfiles map[string]*segmentProfile

	// Pre-computed lookalike audiences
	lookalikeAudiences map[string]*lookalikeAudience

	// Configuration
	config *LookalikeConfig
}

// LookalikeConfig holds configuration for lookalike generation
type LookalikeConfig struct {
	Enabled             bool
	MinSeedSize         int
	MaxAudienceSize     int
	SimilarityThreshold float64
	RefreshIntervalHrs  int
	FeatureWeights      map[string]float64
}

type userProfile struct {
	UserID          string
	Segments        []string
	Interests       []string
	DeviceTypes     []string
	Geo             geoProfile
	BehaviorMetrics behaviorMetrics
	LastSeen        time.Time
	ProfileScore    float64
}

type geoProfile struct {
	Country  string
	Region   string
	City     string
	DMA      string
	Language string
}

type behaviorMetrics struct {
	TotalImpressions int64
	TotalClicks      int64
	TotalConversions int64
	AvgSessionTime   float64
	VisitFrequency   float64
	RecencyDays      int
	EngagementScore  float64
}

type segmentProfile struct {
	SegmentID     string
	Name          string
	UserCount     int64
	AvgCTR        float64
	AvgCVR        float64
	TopInterests  []string
	TopGeos       []string
	TopDevices    []string
	FeatureVector []float64
}

type lookalikeAudience struct {
	ID             string
	Name           string
	SeedSegmentID  string
	SeedSize       int
	AudienceSize   int
	ExpansionRatio float64
	Users          []lookalikeUser
	CreatedAt      time.Time
	ExpiresAt      time.Time
	QualityScore   float64
}

type lookalikeUser struct {
	UserID          string
	SimilarityScore float64
	MatchedFeatures []string
	Rank            int
}

// NewLookalikeService creates a new lookalike audience service
func NewLookalikeService(c cache.Cache) *LookalikeService {
	return &LookalikeService{
		cacheClient:        c,
		userProfiles:       make(map[string]*userProfile),
		segmentProfiles:    make(map[string]*segmentProfile),
		lookalikeAudiences: make(map[string]*lookalikeAudience),
		config: &LookalikeConfig{
			Enabled:             true,
			MinSeedSize:         100,
			MaxAudienceSize:     1000000,
			SimilarityThreshold: 0.6,
			RefreshIntervalHrs:  24,
			FeatureWeights: map[string]float64{
				"interests":    0.30,
				"behavior":     0.25,
				"demographics": 0.20,
				"geo":          0.15,
				"device":       0.10,
			},
		},
	}
}

// LookalikeResult represents the result of lookalike audience generation
type LookalikeResult struct {
	AudienceID     string             `json:"audience_id"`
	Name           string             `json:"name"`
	SeedSize       int                `json:"seed_size"`
	AudienceSize   int                `json:"audience_size"`
	ExpansionRatio float64            `json:"expansion_ratio"`
	QualityScore   float64            `json:"quality_score"`
	TopUsers       []lookalikeUser    `json:"top_users"`
	FeatureOverlap map[string]float64 `json:"feature_overlap"`
	Status         string             `json:"status"`
}

// GenerateLookalike generates a lookalike audience from seed users
func (s *LookalikeService) GenerateLookalike(seedUserIDs []string, name string, expansionFactor float64) *LookalikeResult {
	if !s.config.Enabled {
		return &LookalikeResult{Status: "disabled"}
	}

	if len(seedUserIDs) < s.config.MinSeedSize {
		return &LookalikeResult{
			Status:   "insufficient_seed",
			SeedSize: len(seedUserIDs),
		}
	}

	// Build seed profile
	seedProfile := s.buildSeedProfile(seedUserIDs)
	if seedProfile == nil {
		return &LookalikeResult{Status: "seed_profile_failed"}
	}

	// Find similar users
	targetSize := int(float64(len(seedUserIDs)) * expansionFactor)
	if targetSize > s.config.MaxAudienceSize {
		targetSize = s.config.MaxAudienceSize
	}

	similarUsers := s.findSimilarUsers(seedProfile, seedUserIDs, targetSize)

	// Calculate quality metrics
	qualityScore := s.calculateAudienceQuality(seedProfile, similarUsers)
	featureOverlap := s.calculateFeatureOverlap(seedProfile, similarUsers)

	// Create lookalike audience
	audienceID := generateAudienceID(name)
	audience := &lookalikeAudience{
		ID:             audienceID,
		Name:           name,
		SeedSize:       len(seedUserIDs),
		AudienceSize:   len(similarUsers),
		ExpansionRatio: float64(len(similarUsers)) / float64(len(seedUserIDs)),
		Users:          similarUsers,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(time.Duration(s.config.RefreshIntervalHrs) * time.Hour),
		QualityScore:   qualityScore,
	}

	// Store audience
	s.mu.Lock()
	s.lookalikeAudiences[audienceID] = audience
	s.mu.Unlock()

	// Return top users for preview
	topUsers := similarUsers
	if len(topUsers) > 100 {
		topUsers = topUsers[:100]
	}

	return &LookalikeResult{
		AudienceID:     audienceID,
		Name:           name,
		SeedSize:       len(seedUserIDs),
		AudienceSize:   len(similarUsers),
		ExpansionRatio: audience.ExpansionRatio,
		QualityScore:   qualityScore,
		TopUsers:       topUsers,
		FeatureOverlap: featureOverlap,
		Status:         "created",
	}
}

func (s *LookalikeService) buildSeedProfile(seedUserIDs []string) *segmentProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profile := &segmentProfile{
		SegmentID:     "seed",
		FeatureVector: make([]float64, 50), // 50-dimensional feature vector
	}

	interestCounts := make(map[string]int)
	geoCounts := make(map[string]int)
	deviceCounts := make(map[string]int)

	validUsers := 0
	totalCTR := 0.0
	totalCVR := 0.0

	for _, userID := range seedUserIDs {
		if user, exists := s.userProfiles[userID]; exists {
			validUsers++

			// Aggregate interests
			for _, interest := range user.Interests {
				interestCounts[interest]++
			}

			// Aggregate geo
			if user.Geo.Country != "" {
				geoCounts[user.Geo.Country]++
			}

			// Aggregate devices
			for _, device := range user.DeviceTypes {
				deviceCounts[device]++
			}

			// Aggregate behavior
			if user.BehaviorMetrics.TotalImpressions > 0 {
				ctr := float64(user.BehaviorMetrics.TotalClicks) / float64(user.BehaviorMetrics.TotalImpressions)
				totalCTR += ctr
			}
			if user.BehaviorMetrics.TotalClicks > 0 {
				cvr := float64(user.BehaviorMetrics.TotalConversions) / float64(user.BehaviorMetrics.TotalClicks)
				totalCVR += cvr
			}

			// Build feature vector contribution
			s.addToFeatureVector(profile.FeatureVector, user)
		}
	}

	if validUsers == 0 {
		return nil
	}

	// Normalize feature vector
	for i := range profile.FeatureVector {
		profile.FeatureVector[i] /= float64(validUsers)
	}

	// Set profile metrics
	profile.UserCount = int64(validUsers)
	profile.AvgCTR = totalCTR / float64(validUsers)
	profile.AvgCVR = totalCVR / float64(validUsers)
	profile.TopInterests = topN(interestCounts, 10)
	profile.TopGeos = topN(geoCounts, 5)
	profile.TopDevices = topN(deviceCounts, 3)

	return profile
}

func (s *LookalikeService) addToFeatureVector(vector []float64, user *userProfile) {
	// Interest features (indices 0-19)
	interestMap := map[string]int{
		"sports": 0, "tech": 1, "finance": 2, "travel": 3, "food": 4,
		"fashion": 5, "auto": 6, "health": 7, "entertainment": 8, "gaming": 9,
		"news": 10, "shopping": 11, "education": 12, "music": 13, "movies": 14,
		"fitness": 15, "home": 16, "pets": 17, "beauty": 18, "parenting": 19,
	}
	for _, interest := range user.Interests {
		if idx, ok := interestMap[interest]; ok {
			vector[idx] += 1.0
		}
	}

	// Behavior features (indices 20-29)
	if user.BehaviorMetrics.TotalImpressions > 0 {
		vector[20] += math.Log1p(float64(user.BehaviorMetrics.TotalImpressions)) / 10.0
	}
	if user.BehaviorMetrics.TotalClicks > 0 {
		vector[21] += math.Log1p(float64(user.BehaviorMetrics.TotalClicks)) / 5.0
	}
	if user.BehaviorMetrics.TotalConversions > 0 {
		vector[22] += math.Log1p(float64(user.BehaviorMetrics.TotalConversions)) / 3.0
	}
	vector[23] += user.BehaviorMetrics.EngagementScore
	vector[24] += math.Min(user.BehaviorMetrics.VisitFrequency/10.0, 1.0)

	// Recency feature
	recencyScore := 1.0 - math.Min(float64(user.BehaviorMetrics.RecencyDays)/30.0, 1.0)
	vector[25] += recencyScore

	// Device features (indices 30-34)
	deviceMap := map[string]int{"mobile": 30, "desktop": 31, "tablet": 32, "ctv": 33, "other": 34}
	for _, device := range user.DeviceTypes {
		if idx, ok := deviceMap[device]; ok {
			vector[idx] += 1.0
		}
	}

	// Geo features encoded (indices 35-49)
	// Simplified - in production would use geo embeddings
	vector[35] += 1.0 // Has geo data indicator
}

func (s *LookalikeService) findSimilarUsers(seedProfile *segmentProfile, excludeUsers []string, targetSize int) []lookalikeUser {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create exclude set
	excludeSet := make(map[string]bool)
	for _, id := range excludeUsers {
		excludeSet[id] = true
	}

	// Calculate similarity for all users
	candidates := make([]lookalikeUser, 0)

	for userID, profile := range s.userProfiles {
		if excludeSet[userID] {
			continue
		}

		// Build user feature vector
		userVector := make([]float64, 50)
		s.addToFeatureVector(userVector, profile)

		// Calculate cosine similarity
		similarity := cosineSimilarity(seedProfile.FeatureVector, userVector)

		if similarity >= s.config.SimilarityThreshold {
			matchedFeatures := s.findMatchedFeatures(seedProfile, profile)
			candidates = append(candidates, lookalikeUser{
				UserID:          userID,
				SimilarityScore: similarity,
				MatchedFeatures: matchedFeatures,
			})
		}
	}

	// Sort by similarity (descending)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].SimilarityScore > candidates[j].SimilarityScore
	})

	// Take top N
	if len(candidates) > targetSize {
		candidates = candidates[:targetSize]
	}

	// Assign ranks
	for i := range candidates {
		candidates[i].Rank = i + 1
	}

	return candidates
}

func (s *LookalikeService) findMatchedFeatures(seedProfile *segmentProfile, user *userProfile) []string {
	matched := make([]string, 0)

	// Check interest overlap
	seedInterests := make(map[string]bool)
	for _, interest := range seedProfile.TopInterests {
		seedInterests[interest] = true
	}
	for _, interest := range user.Interests {
		if seedInterests[interest] {
			matched = append(matched, "interest:"+interest)
		}
	}

	// Check geo overlap
	for _, geo := range seedProfile.TopGeos {
		if user.Geo.Country == geo {
			matched = append(matched, "geo:"+geo)
		}
	}

	// Check device overlap
	seedDevices := make(map[string]bool)
	for _, device := range seedProfile.TopDevices {
		seedDevices[device] = true
	}
	for _, device := range user.DeviceTypes {
		if seedDevices[device] {
			matched = append(matched, "device:"+device)
		}
	}

	// Check behavior similarity
	if user.BehaviorMetrics.EngagementScore > 0.5 && seedProfile.AvgCTR > 0.02 {
		matched = append(matched, "behavior:high_engagement")
	}

	return matched
}

func (s *LookalikeService) calculateAudienceQuality(seedProfile *segmentProfile, users []lookalikeUser) float64 {
	if len(users) == 0 {
		return 0.0
	}

	// Average similarity score
	totalSimilarity := 0.0
	for _, user := range users {
		totalSimilarity += user.SimilarityScore
	}
	avgSimilarity := totalSimilarity / float64(len(users))

	// Feature coverage
	featureCoverage := 0.0
	for _, user := range users {
		featureCoverage += float64(len(user.MatchedFeatures)) / 10.0 // Normalize by expected max features
	}
	avgFeatureCoverage := math.Min(featureCoverage/float64(len(users)), 1.0)

	// Size quality (prefer reasonable expansion ratios)
	sizeQuality := 1.0
	expansionRatio := float64(len(users)) / float64(seedProfile.UserCount)
	if expansionRatio > 10 {
		sizeQuality = 0.7 // Too much expansion
	} else if expansionRatio < 1.5 {
		sizeQuality = 0.8 // Too little expansion
	}

	// Weighted quality score
	return avgSimilarity*0.5 + avgFeatureCoverage*0.3 + sizeQuality*0.2
}

func (s *LookalikeService) calculateFeatureOverlap(seedProfile *segmentProfile, users []lookalikeUser) map[string]float64 {
	featureCounts := make(map[string]int)
	totalUsers := len(users)

	for _, user := range users {
		for _, feature := range user.MatchedFeatures {
			featureCounts[feature]++
		}
	}

	overlap := make(map[string]float64)
	for feature, count := range featureCounts {
		overlap[feature] = float64(count) / float64(totalUsers)
	}

	return overlap
}

// CreateUserProfile creates a new user profile (helper for handlers)
func (s *LookalikeService) CreateUserProfile(segments, interests, deviceTypes []string, country, region, city string) *userProfile {
	return &userProfile{
		Segments:    segments,
		Interests:   interests,
		DeviceTypes: deviceTypes,
		Geo: geoProfile{
			Country: country,
			Region:  region,
			City:    city,
		},
		LastSeen: time.Now(),
	}
}

// RegisterUserProfile registers or updates a user profile
func (s *LookalikeService) RegisterUserProfile(userID string, profile *userProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()

	profile.UserID = userID
	profile.LastSeen = time.Now()
	s.userProfiles[userID] = profile
}

// UpdateUserBehavior updates user behavior metrics
func (s *LookalikeService) UpdateUserBehavior(userID string, impressions, clicks, conversions int64, sessionTime float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if profile, exists := s.userProfiles[userID]; exists {
		profile.BehaviorMetrics.TotalImpressions += impressions
		profile.BehaviorMetrics.TotalClicks += clicks
		profile.BehaviorMetrics.TotalConversions += conversions

		// Update engagement score
		ctr := 0.0
		if profile.BehaviorMetrics.TotalImpressions > 0 {
			ctr = float64(profile.BehaviorMetrics.TotalClicks) / float64(profile.BehaviorMetrics.TotalImpressions)
		}
		profile.BehaviorMetrics.EngagementScore = math.Min(ctr*50, 1.0)
		profile.BehaviorMetrics.RecencyDays = 0
		profile.LastSeen = time.Now()
	}
}

// GetLookalikeAudience retrieves a lookalike audience by ID
func (s *LookalikeService) GetLookalikeAudience(audienceID string) *lookalikeAudience {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if audience, exists := s.lookalikeAudiences[audienceID]; exists {
		if time.Now().Before(audience.ExpiresAt) {
			return audience
		}
	}
	return nil
}

// IsUserInLookalike checks if a user is in a lookalike audience
func (s *LookalikeService) IsUserInLookalike(userID, audienceID string) (bool, float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if audience, exists := s.lookalikeAudiences[audienceID]; exists {
		for _, user := range audience.Users {
			if user.UserID == userID {
				return true, user.SimilarityScore
			}
		}
	}
	return false, 0.0
}

// GetLookalikeStats returns statistics about lookalike audiences
func (s *LookalikeService) GetLookalikeStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalAudiences := len(s.lookalikeAudiences)
	totalUsers := int64(0)
	avgQuality := 0.0

	for _, audience := range s.lookalikeAudiences {
		totalUsers += int64(audience.AudienceSize)
		avgQuality += audience.QualityScore
	}

	if totalAudiences > 0 {
		avgQuality /= float64(totalAudiences)
	}

	return map[string]interface{}{
		"total_audiences":      totalAudiences,
		"total_users_profiled": len(s.userProfiles),
		"total_audience_reach": totalUsers,
		"avg_quality_score":    avgQuality,
	}
}

// Helper functions

func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func topN(counts map[string]int, n int) []string {
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

func generateAudienceID(name string) string {
	return "lal_" + name + "_" + time.Now().Format("20060102150405")
}
