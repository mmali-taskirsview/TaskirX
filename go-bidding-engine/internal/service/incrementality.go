package service

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// IncrementalityService manages lift measurement experiments
type IncrementalityService struct {
	cache       cache.Cache
	mu          sync.RWMutex
	experiments map[string]*experiment
}

type experiment struct {
	config       *model.IncrementalityConfig
	controlUsers map[string]*userStats
	testUsers    map[string]*userStats
	startTime    time.Time
	lastUpdated  time.Time
}

type userStats struct {
	impressions int
	clicks      int
	conversions int
	revenue     float64
	lastSeen    time.Time
}

// NewIncrementalityService creates a new incrementality testing service
func NewIncrementalityService(c cache.Cache) *IncrementalityService {
	return &IncrementalityService{
		cache:       c,
		experiments: make(map[string]*experiment),
	}
}

// EvaluateUser determines if user is in control/test and returns result
func (s *IncrementalityService) EvaluateUser(campaign *model.Campaign, req *model.BidRequest) *model.IncrementalityResult {
	config := campaign.Targeting.IncrementalityConfig
	if config == nil || !config.Enabled {
		return &model.IncrementalityResult{
			Status:             "disabled",
			UserInControlGroup: false,
		}
	}

	experimentID := config.ExperimentID
	if experimentID == "" {
		experimentID = campaign.ID + "_incrementality"
	}

	// Get or create experiment
	s.mu.Lock()
	exp, exists := s.experiments[experimentID]
	if !exists {
		exp = &experiment{
			config:       config,
			controlUsers: make(map[string]*userStats),
			testUsers:    make(map[string]*userStats),
			startTime:    time.Now(),
		}
		s.experiments[experimentID] = exp
	}
	s.mu.Unlock()

	// Determine user assignment
	userID := req.User.ID
	isControl := s.assignToGroup(userID, config, req)

	return &model.IncrementalityResult{
		ExperimentID:       experimentID,
		Status:             "running",
		UserInControlGroup: isControl,
	}
}

// assignToGroup deterministically assigns user to control or test
func (s *IncrementalityService) assignToGroup(userID string, config *model.IncrementalityConfig, req *model.BidRequest) bool {
	controlPercent := config.ControlPercent
	if controlPercent <= 0 {
		controlPercent = 10.0 // Default 10% control
	}
	if controlPercent >= 100 {
		controlPercent = 10.0
	}

	switch config.HoldoutType {
	case "geo":
		// Geo-based holdout using Device.Geo (only Country and City available)
		for _, holdoutGeo := range config.GeoHoldouts {
			if req.Device.Geo.Country == holdoutGeo || req.Device.Geo.City == holdoutGeo {
				return true // In control
			}
		}
		return false

	case "time":
		// Time-based holdout (e.g., certain hours are control)
		hour := time.Now().Hour()
		// Control hours: 2am-4am (low traffic, minimal impact)
		if hour >= 2 && hour <= 4 {
			return true
		}
		return false

	default: // "user" - default user-based random assignment
		// Deterministic hash-based assignment
		hash := sha256.Sum256([]byte(userID + config.ExperimentID))
		hashInt := int(hash[0])<<8 + int(hash[1])
		bucket := float64(hashInt%1000) / 10.0 // 0-100

		return bucket < controlPercent
	}
}

// RecordImpression records an impression for incrementality tracking
func (s *IncrementalityService) RecordImpression(experimentID, userID string, isControl bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exp, exists := s.experiments[experimentID]
	if !exists {
		return
	}

	var users map[string]*userStats
	if isControl {
		users = exp.controlUsers
	} else {
		users = exp.testUsers
	}

	if _, exists := users[userID]; !exists {
		users[userID] = &userStats{}
	}

	users[userID].impressions++
	users[userID].lastSeen = time.Now()
	exp.lastUpdated = time.Now()
}

// RecordConversion records a conversion for incrementality tracking
func (s *IncrementalityService) RecordConversion(experimentID, userID string, isControl bool, revenue float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exp, exists := s.experiments[experimentID]
	if !exists {
		return
	}

	var users map[string]*userStats
	if isControl {
		users = exp.controlUsers
	} else {
		users = exp.testUsers
	}

	if _, exists := users[userID]; !exists {
		users[userID] = &userStats{}
	}

	users[userID].conversions++
	users[userID].revenue += revenue
	users[userID].lastSeen = time.Now()
	exp.lastUpdated = time.Now()
}

// GetExperimentResults calculates and returns experiment results
func (s *IncrementalityService) GetExperimentResults(experimentID string) *model.IncrementalityResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exp, exists := s.experiments[experimentID]
	if !exists {
		return &model.IncrementalityResult{
			ExperimentID: experimentID,
			Status:       "not_found",
		}
	}

	// Calculate metrics for each group
	controlStats := s.aggregateStats(exp.controlUsers)
	testStats := s.aggregateStats(exp.testUsers)

	// Check sample size
	minSampleSize := exp.config.MinSampleSize
	if minSampleSize <= 0 {
		minSampleSize = 1000
	}

	if controlStats.users < minSampleSize || testStats.users < minSampleSize {
		return &model.IncrementalityResult{
			ExperimentID:     experimentID,
			Status:           "insufficient_data",
			ControlGroupSize: controlStats.users,
			TestGroupSize:    testStats.users,
		}
	}

	// Calculate conversion rates
	controlCVR := 0.0
	testCVR := 0.0
	if controlStats.users > 0 {
		controlCVR = float64(controlStats.conversions) / float64(controlStats.users)
	}
	if testStats.users > 0 {
		testCVR = float64(testStats.conversions) / float64(testStats.users)
	}

	// Calculate lift
	lift := 0.0
	if controlCVR > 0 {
		lift = (testCVR - controlCVR) / controlCVR * 100
	} else if testCVR > 0 {
		lift = 100.0 // Infinite lift if control has 0 conversions
	}

	// Calculate statistical significance (simplified z-test)
	confidence := s.calculateSignificance(controlStats, testStats)

	// Calculate incremental metrics
	incrementalConv := testStats.conversions - int(float64(testStats.users)*controlCVR)
	if incrementalConv < 0 {
		incrementalConv = 0
	}

	incrementalRevenue := testStats.revenue - (float64(testStats.users) * controlStats.revenue / float64(controlStats.users))
	if incrementalRevenue < 0 {
		incrementalRevenue = 0
	}

	// Calculate iROAS (incremental ROAS)
	// This would need spend data - using placeholder
	iROAS := 0.0
	if incrementalRevenue > 0 {
		iROAS = incrementalRevenue / (incrementalRevenue * 0.3) // Assume 30% of revenue as spend
	}

	// Determine status
	status := "running"
	requiredConfidence := exp.config.ConfidenceLevel
	if requiredConfidence <= 0 {
		requiredConfidence = 0.95
	}
	if confidence >= requiredConfidence {
		status = "complete"
	}

	// Generate recommendation
	recommendation := s.generateRecommendation(lift, confidence, status)

	return &model.IncrementalityResult{
		ExperimentID:       experimentID,
		Status:             status,
		ControlGroupSize:   controlStats.users,
		TestGroupSize:      testStats.users,
		Lift:               lift,
		LiftConfidence:     confidence,
		IncrementalConv:    incrementalConv,
		IncrementalRevenue: incrementalRevenue,
		ROAS:               iROAS, // Use ROAS field (iROAS is unexported)
		Recommendation:     recommendation,
	}
}

type aggregatedStats struct {
	users       int
	impressions int
	conversions int
	revenue     float64
}

func (s *IncrementalityService) aggregateStats(users map[string]*userStats) aggregatedStats {
	stats := aggregatedStats{
		users: len(users),
	}
	for _, u := range users {
		stats.impressions += u.impressions
		stats.conversions += u.conversions
		stats.revenue += u.revenue
	}
	return stats
}

func (s *IncrementalityService) calculateSignificance(control, test aggregatedStats) float64 {
	if control.users == 0 || test.users == 0 {
		return 0
	}

	// Calculate conversion rates
	p1 := float64(control.conversions) / float64(control.users)
	p2 := float64(test.conversions) / float64(test.users)
	n1 := float64(control.users)
	n2 := float64(test.users)

	// Pooled proportion
	pPooled := float64(control.conversions+test.conversions) / (n1 + n2)

	// Standard error
	se := math.Sqrt(pPooled * (1 - pPooled) * (1/n1 + 1/n2))
	if se == 0 {
		return 0
	}

	// Z-score
	z := math.Abs(p2-p1) / se

	// Convert to confidence (approximation)
	// Using error function approximation
	confidence := 1 - 2*normalCDF(-z)

	return confidence
}

// normalCDF approximates the cumulative distribution function of standard normal
func normalCDF(x float64) float64 {
	// Approximation using error function
	return 0.5 * (1 + erf(x/math.Sqrt(2)))
}

// erf approximates the error function
func erf(x float64) float64 {
	// Horner form approximation
	a1 := 0.254829592
	a2 := -0.284496736
	a3 := 1.421413741
	a4 := -1.453152027
	a5 := 1.061405429
	p := 0.3275911

	sign := 1.0
	if x < 0 {
		sign = -1.0
	}
	x = math.Abs(x)

	t := 1.0 / (1.0 + p*x)
	y := 1.0 - (((((a5*t+a4)*t)+a3)*t+a2)*t+a1)*t*math.Exp(-x*x)

	return sign * y
}

func (s *IncrementalityService) generateRecommendation(lift, confidence float64, status string) string {
	if status != "complete" {
		return "Continue experiment to reach statistical significance"
	}

	if confidence < 0.9 {
		return "Results inconclusive. Consider extending the experiment."
	}

	if lift > 20 {
		return "Strong positive lift detected. Advertising is highly effective for this campaign."
	} else if lift > 5 {
		return "Moderate positive lift detected. Advertising provides measurable incremental value."
	} else if lift > 0 {
		return "Small positive lift detected. Consider optimizing targeting to improve efficiency."
	} else if lift > -5 {
		return "No significant lift detected. Consider reviewing campaign strategy."
	} else {
		return "Negative lift detected. Review campaign targeting and creative strategy."
	}
}

// GetUserExperimentGroup returns the experiment group for a user (for API use)
func (s *IncrementalityService) GetUserExperimentGroup(experimentID, userID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exp, exists := s.experiments[experimentID]
	if !exists {
		return "", false
	}

	if _, inControl := exp.controlUsers[userID]; inControl {
		return "control", true
	}
	if _, inTest := exp.testUsers[userID]; inTest {
		return "test", true
	}

	return "", false
}

// hashToGroup creates deterministic group assignment
func hashToGroup(userID, experimentID string, controlPercent float64) bool {
	hash := sha256.Sum256([]byte(userID + experimentID))
	hashHex := hex.EncodeToString(hash[:8])
	// Convert first 8 hex chars to int
	var hashInt int64
	for _, c := range hashHex {
		hashInt = hashInt*16 + int64(c-'0')
		if c >= 'a' {
			hashInt = hashInt - 49 + 10 // Adjust for hex letters
		}
	}
	bucket := float64(hashInt%1000) / 10.0
	return bucket < controlPercent
}
