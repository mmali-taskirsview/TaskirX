package service

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// CreativeOptimizationService handles automatic creative selection and optimization
type CreativeOptimizationService struct {
	cache           cache.Cache
	mu              sync.RWMutex
	creativePerf    map[string]*creativePerformance            // key: creativeID
	placementPerf   map[string]map[string]*creativePerformance // key: placement -> creativeID -> perf
	explorationSeed *rand.Rand
}

type creativePerformance struct {
	impressions int64
	clicks      int64
	conversions int64
	viewTime    float64
	engagements int64
	lastUpdated time.Time
	score       float64
	isExploring bool
}

// NewCreativeOptimizationService creates a new creative optimization service
func NewCreativeOptimizationService(c cache.Cache) *CreativeOptimizationService {
	return &CreativeOptimizationService{
		cache:           c,
		creativePerf:    make(map[string]*creativePerformance),
		placementPerf:   make(map[string]map[string]*creativePerformance),
		explorationSeed: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SelectCreative selects the optimal creative for the request
func (s *CreativeOptimizationService) SelectCreative(campaign *model.Campaign, req *model.BidRequest) *model.CreativeOptimizationResult {
	config := campaign.Targeting.CreativeOptimization
	if config == nil || !config.Enabled || len(config.CreativePool) == 0 {
		// Use campaign ID + creative type as identifier since Creative doesn't have ID
		defaultCreativeID := campaign.ID + "_" + campaign.Creative.Type
		return &model.CreativeOptimizationResult{
			SelectedCreativeID: defaultCreativeID,
			SelectionMethod:    "default",
			Confidence:         1.0,
			Reason:             "creative_optimization_disabled",
		}
	}

	// Filter active creatives
	activeCreatives := s.filterActiveCreatives(config.CreativePool)
	if len(activeCreatives) == 0 {
		defaultCreativeID := campaign.ID + "_" + campaign.Creative.Type
		return &model.CreativeOptimizationResult{
			SelectedCreativeID: defaultCreativeID,
			SelectionMethod:    "default",
			Confidence:         1.0,
			Reason:             "no_active_creatives",
		}
	}

	// Check placement-specific rules first
	if ruleResult := s.checkPlacementRules(config.PlacementRules, req, activeCreatives); ruleResult != nil {
		return ruleResult
	}

	// Determine if we should explore or exploit
	explorationRate := config.ExplorationRate
	if explorationRate <= 0 {
		explorationRate = 0.1 // Default 10% exploration
	}

	if s.explorationSeed.Float64() < explorationRate {
		return s.exploreCreative(activeCreatives, config)
	}

	// Exploit: select best performing creative
	return s.exploitBestCreative(activeCreatives, config, req)
}

func (s *CreativeOptimizationService) filterActiveCreatives(pool []model.CreativeVariant) []model.CreativeVariant {
	var active []model.CreativeVariant
	for _, cv := range pool {
		if cv.Status == "" || cv.Status == "active" || cv.Status == "testing" {
			active = append(active, cv)
		}
	}
	return active
}

func (s *CreativeOptimizationService) checkPlacementRules(rules []model.PlacementCreativeRule, req *model.BidRequest, creatives []model.CreativeVariant) *model.CreativeOptimizationResult {
	if len(rules) == 0 {
		return nil
	}

	placementType := s.detectPlacementType(req)

	for _, rule := range rules {
		if rule.PlacementType == placementType {
			// Find first matching creative from rule
			for _, creativeID := range rule.CreativeIDs {
				for _, cv := range creatives {
					if cv.ID == creativeID {
						return &model.CreativeOptimizationResult{
							SelectedCreativeID: cv.ID,
							SelectionMethod:    "rule_based",
							Confidence:         0.9,
							Reason:             "placement_rule_match",
						}
					}
				}
			}
		}
	}

	return nil
}

func (s *CreativeOptimizationService) detectPlacementType(req *model.BidRequest) string {
	// Infer from AdSlot formats
	if len(req.AdSlot.Formats) > 0 {
		return req.AdSlot.Formats[0] // Return primary format
	}
	// Infer from dimensions if available
	if len(req.AdSlot.Dimensions) >= 2 {
		width := req.AdSlot.Dimensions[0]
		height := req.AdSlot.Dimensions[1]
		if width > 0 && height > 0 {
			ratio := float64(width) / float64(height)
			if ratio > 2.5 {
				return "banner"
			} else if ratio < 0.7 {
				return "interstitial"
			}
		}
	}
	return "banner"
}

func (s *CreativeOptimizationService) exploreCreative(creatives []model.CreativeVariant, _ *model.CreativeOptimization) *model.CreativeOptimizationResult {
	// Weight-based random selection for exploration
	var totalWeight float64
	for _, cv := range creatives {
		weight := cv.Weight
		if weight <= 0 {
			weight = 1.0
		}
		totalWeight += weight
	}

	r := s.explorationSeed.Float64() * totalWeight
	var cumWeight float64
	for _, cv := range creatives {
		weight := cv.Weight
		if weight <= 0 {
			weight = 1.0
		}
		cumWeight += weight
		if r <= cumWeight {
			return &model.CreativeOptimizationResult{
				SelectedCreativeID: cv.ID,
				SelectionMethod:    "exploration",
				Confidence:         0.5,
				Reason:             "exploration_test",
			}
		}
	}

	// Fallback
	return &model.CreativeOptimizationResult{
		SelectedCreativeID: creatives[0].ID,
		SelectionMethod:    "exploration",
		Confidence:         0.5,
		Reason:             "exploration_fallback",
	}
}

func (s *CreativeOptimizationService) exploitBestCreative(creatives []model.CreativeVariant, config *model.CreativeOptimization, _ *model.BidRequest) *model.CreativeOptimizationResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type scoredCreative struct {
		id         string
		score      float64
		confidence float64
		ctr        float64
		cvr        float64
	}

	var scored []scoredCreative
	minImpressions := config.MinImpressions
	if minImpressions <= 0 {
		minImpressions = 100
	}

	for _, cv := range creatives {
		perf, exists := s.creativePerf[cv.ID]

		if !exists || perf.impressions < int64(minImpressions) {
			// Not enough data, use exploration score
			scored = append(scored, scoredCreative{
				id:         cv.ID,
				score:      0.5, // Neutral score
				confidence: 0.3,
			})
			continue
		}

		// Calculate score based on optimization goal
		score := s.calculateCreativeScore(perf, config.OptimizationGoal)
		ctr := float64(perf.clicks) / float64(perf.impressions)
		cvr := 0.0
		if perf.clicks > 0 {
			cvr = float64(perf.conversions) / float64(perf.clicks)
		}

		// Confidence based on sample size
		confidence := math.Min(float64(perf.impressions)/float64(minImpressions*10), 1.0)

		scored = append(scored, scoredCreative{
			id:         cv.ID,
			score:      score,
			confidence: confidence,
			ctr:        ctr,
			cvr:        cvr,
		})
	}

	if len(scored) == 0 {
		return &model.CreativeOptimizationResult{
			SelectedCreativeID: creatives[0].ID,
			SelectionMethod:    "default",
			Confidence:         0.5,
			Reason:             "no_scored_creatives",
		}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	best := scored[0]

	// Get alternatives
	var alternatives []string
	for i := 1; i < len(scored) && i < 4; i++ {
		alternatives = append(alternatives, scored[i].id)
	}

	return &model.CreativeOptimizationResult{
		SelectedCreativeID: best.id,
		SelectionMethod:    "optimized",
		PredictedCTR:       best.ctr,
		PredictedCVR:       best.cvr,
		Confidence:         best.confidence,
		AlternativeIDs:     alternatives,
		Reason:             "best_performer",
	}
}

func (s *CreativeOptimizationService) calculateCreativeScore(perf *creativePerformance, goal string) float64 {
	if perf.impressions == 0 {
		return 0
	}

	ctr := float64(perf.clicks) / float64(perf.impressions)
	cvr := 0.0
	if perf.clicks > 0 {
		cvr = float64(perf.conversions) / float64(perf.clicks)
	}
	engagementRate := float64(perf.engagements) / float64(perf.impressions)
	avgViewTime := perf.viewTime / float64(perf.impressions)

	switch goal {
	case "ctr":
		return ctr
	case "cvr":
		return cvr
	case "engagement":
		return engagementRate
	case "viewability":
		return avgViewTime / 30.0 // Normalize to 30 second view
	default:
		// Composite score
		return ctr*0.4 + cvr*0.3 + engagementRate*0.2 + (avgViewTime/30.0)*0.1
	}
}

// RecordImpression records a creative impression
func (s *CreativeOptimizationService) RecordImpression(creativeID string, placement string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.creativePerf[creativeID]; !exists {
		s.creativePerf[creativeID] = &creativePerformance{}
	}
	s.creativePerf[creativeID].impressions++
	s.creativePerf[creativeID].lastUpdated = time.Now()

	// Also track placement-specific
	if _, exists := s.placementPerf[placement]; !exists {
		s.placementPerf[placement] = make(map[string]*creativePerformance)
	}
	if _, exists := s.placementPerf[placement][creativeID]; !exists {
		s.placementPerf[placement][creativeID] = &creativePerformance{}
	}
	s.placementPerf[placement][creativeID].impressions++
}

// RecordClick records a creative click
func (s *CreativeOptimizationService) RecordClick(creativeID string, placement string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.creativePerf[creativeID]; !exists {
		s.creativePerf[creativeID] = &creativePerformance{}
	}
	s.creativePerf[creativeID].clicks++
	s.creativePerf[creativeID].lastUpdated = time.Now()

	if _, exists := s.placementPerf[placement]; exists {
		if _, exists := s.placementPerf[placement][creativeID]; !exists {
			s.placementPerf[placement][creativeID] = &creativePerformance{}
		}
		s.placementPerf[placement][creativeID].clicks++
	}
}

// RecordConversion records a creative conversion
func (s *CreativeOptimizationService) RecordConversion(creativeID string, placement string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.creativePerf[creativeID]; !exists {
		s.creativePerf[creativeID] = &creativePerformance{}
	}
	s.creativePerf[creativeID].conversions++
	s.creativePerf[creativeID].lastUpdated = time.Now()

	if _, exists := s.placementPerf[placement]; exists {
		if _, exists := s.placementPerf[placement][creativeID]; !exists {
			s.placementPerf[placement][creativeID] = &creativePerformance{}
		}
		s.placementPerf[placement][creativeID].conversions++
	}
}

// RecordEngagement records creative engagement
func (s *CreativeOptimizationService) RecordEngagement(creativeID string, placement string, viewTime float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.creativePerf[creativeID]; !exists {
		s.creativePerf[creativeID] = &creativePerformance{}
	}
	s.creativePerf[creativeID].engagements++
	s.creativePerf[creativeID].viewTime += viewTime
	s.creativePerf[creativeID].lastUpdated = time.Now()

	if _, exists := s.placementPerf[placement]; exists {
		if _, exists := s.placementPerf[placement][creativeID]; !exists {
			s.placementPerf[placement][creativeID] = &creativePerformance{}
		}
		s.placementPerf[placement][creativeID].engagements++
		s.placementPerf[placement][creativeID].viewTime += viewTime
	}
}

// CheckAutoPause checks if a creative should be paused based on performance
func (s *CreativeOptimizationService) CheckAutoPause(config *model.CreativeOptimization) []string {
	if config == nil || !config.AutoPause {
		return nil
	}

	threshold := config.PauseThreshold
	if threshold <= 0 {
		threshold = 0.3 // Default: pause if below 30% of average
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate average score
	var totalScore float64
	var count int
	for _, perf := range s.creativePerf {
		if perf.impressions >= 100 {
			score := s.calculateCreativeScore(perf, config.OptimizationGoal)
			totalScore += score
			count++
		}
	}

	if count == 0 {
		return nil
	}

	avgScore := totalScore / float64(count)
	pauseThreshold := avgScore * threshold

	var toPause []string
	for creativeID, perf := range s.creativePerf {
		if perf.impressions >= 100 {
			score := s.calculateCreativeScore(perf, config.OptimizationGoal)
			if score < pauseThreshold {
				toPause = append(toPause, creativeID)
			}
		}
	}

	return toPause
}

// GetCreativePerformance returns performance data for a creative
func (s *CreativeOptimizationService) GetCreativePerformance(creativeID string) *model.CreativePerf {
	s.mu.RLock()
	defer s.mu.RUnlock()

	perf, exists := s.creativePerf[creativeID]
	if !exists {
		return nil
	}

	ctr := 0.0
	cvr := 0.0
	engRate := 0.0
	avgView := 0.0

	if perf.impressions > 0 {
		ctr = float64(perf.clicks) / float64(perf.impressions)
		engRate = float64(perf.engagements) / float64(perf.impressions)
		avgView = perf.viewTime / float64(perf.impressions)
	}
	if perf.clicks > 0 {
		cvr = float64(perf.conversions) / float64(perf.clicks)
	}

	return &model.CreativePerf{
		Impressions:    perf.impressions,
		Clicks:         perf.clicks,
		Conversions:    perf.conversions,
		CTR:            ctr,
		CVR:            cvr,
		EngagementRate: engRate,
		AvgTimeViewed:  avgView,
		Score:          perf.score,
	}
}
