package service

import (
	"math"
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// DynamicBidService provides ML-based real-time bid adjustments
type DynamicBidService struct {
	cacheClient cache.Cache
	mu          sync.RWMutex

	// Historical performance data
	contextPerformance map[string]*contextStats  // key: context hash
	publisherStats     map[string]*publisherPerf // key: publisher_id
	hourlyMultipliers  map[int]float64           // hour -> multiplier
	deviceMultipliers  map[string]float64        // device_type -> multiplier

	// ML model weights (simplified linear model)
	weights *bidModelWeights

	// Configuration
	config *DynamicBidConfig
}

// DynamicBidConfig holds configuration for dynamic bidding
type DynamicBidConfig struct {
	Enabled             bool
	LearningRate        float64
	ExplorationRate     float64
	MinBidMultiplier    float64
	MaxBidMultiplier    float64
	HistoryWindowHours  int
	MinSamplesForUpdate int
}

type contextStats struct {
	impressions   int64
	clicks        int64
	conversions   int64
	spend         float64
	revenue       float64
	wins          int64
	bids          int64
	totalBidPrice float64
	totalWinPrice float64
	lastUpdated   time.Time
}

type publisherPerf struct {
	winRate     float64
	avgWinPrice float64
	avgCTR      float64
	avgCVR      float64
	quality     float64 // 0-1 quality score
	samples     int64
}

type bidModelWeights struct {
	baseWeight        float64
	ctrWeight         float64
	cvrWeight         float64
	winRateWeight     float64
	recencyWeight     float64
	qualityWeight     float64
	competitionWeight float64
}

// NewDynamicBidService creates a new dynamic bid service
func NewDynamicBidService(c cache.Cache) *DynamicBidService {
	return &DynamicBidService{
		cacheClient:        c,
		contextPerformance: make(map[string]*contextStats),
		publisherStats:     make(map[string]*publisherPerf),
		hourlyMultipliers:  initHourlyMultipliers(),
		deviceMultipliers:  initDeviceMultipliers(),
		weights:            initDefaultWeights(),
		config: &DynamicBidConfig{
			Enabled:             true,
			LearningRate:        0.01,
			ExplorationRate:     0.1,
			MinBidMultiplier:    0.5,
			MaxBidMultiplier:    2.0,
			HistoryWindowHours:  168, // 7 days
			MinSamplesForUpdate: 100,
		},
	}
}

func initHourlyMultipliers() map[int]float64 {
	// Default hourly multipliers based on typical traffic patterns
	return map[int]float64{
		0: 0.7, 1: 0.6, 2: 0.5, 3: 0.5, 4: 0.6, 5: 0.7,
		6: 0.8, 7: 0.9, 8: 1.0, 9: 1.1, 10: 1.2, 11: 1.2,
		12: 1.1, 13: 1.0, 14: 1.0, 15: 1.1, 16: 1.2, 17: 1.3,
		18: 1.4, 19: 1.5, 20: 1.4, 21: 1.3, 22: 1.1, 23: 0.9,
	}
}

func initDeviceMultipliers() map[string]float64 {
	return map[string]float64{
		"mobile":  1.0,
		"desktop": 1.1,
		"tablet":  0.9,
		"ctv":     1.3,
		"other":   0.8,
	}
}

func initDefaultWeights() *bidModelWeights {
	return &bidModelWeights{
		baseWeight:        1.0,
		ctrWeight:         0.25,
		cvrWeight:         0.30,
		winRateWeight:     0.20,
		recencyWeight:     0.10,
		qualityWeight:     0.10,
		competitionWeight: 0.05,
	}
}

// DynamicBidResult represents the result of dynamic bid calculation
type DynamicBidResult struct {
	OriginalBid     float64            `json:"original_bid"`
	AdjustedBid     float64            `json:"adjusted_bid"`
	Multiplier      float64            `json:"multiplier"`
	Confidence      float64            `json:"confidence"`
	Factors         map[string]float64 `json:"factors"`
	Recommendation  string             `json:"recommendation"`
	ExpectedWinRate float64            `json:"expected_win_rate"`
	ExpectedROI     float64            `json:"expected_roi"`
}

// CalculateDynamicBid calculates the optimal bid price based on context
func (s *DynamicBidService) CalculateDynamicBid(campaign *model.Campaign, req *model.BidRequest) *DynamicBidResult {
	if !s.config.Enabled {
		return &DynamicBidResult{
			OriginalBid: campaign.BidPrice,
			AdjustedBid: campaign.BidPrice,
			Multiplier:  1.0,
			Confidence:  0.0,
		}
	}

	baseBid := campaign.BidPrice
	factors := make(map[string]float64)

	// 1. Time-based adjustment
	hour := time.Now().Hour()
	timeFactor := s.hourlyMultipliers[hour]
	factors["time"] = timeFactor

	// 2. Device-based adjustment
	deviceFactor := s.getDeviceMultiplier(req.Device.Type)
	factors["device"] = deviceFactor

	// 3. Publisher performance adjustment
	pubFactor := s.getPublisherFactor(req.PublisherID)
	factors["publisher"] = pubFactor

	// 4. Historical context performance
	contextFactor := s.getContextFactor(campaign.ID, req)
	factors["context"] = contextFactor

	// 5. Competition level adjustment
	competitionFactor := s.getCompetitionFactor(req)
	factors["competition"] = competitionFactor

	// 6. Goal-based adjustment
	goalFactor := s.getGoalFactor(campaign)
	factors["goal"] = goalFactor

	// Calculate weighted multiplier
	multiplier := s.calculateWeightedMultiplier(factors)

	// Apply bounds
	multiplier = s.clampMultiplier(multiplier)

	// Calculate adjusted bid
	adjustedBid := baseBid * multiplier

	// Calculate confidence based on data availability
	confidence := s.calculateConfidence(campaign.ID, req.PublisherID)

	// Expected metrics
	expectedWinRate := s.predictWinRate(adjustedBid, req.PublisherID)
	expectedROI := s.predictROI(campaign, adjustedBid, expectedWinRate)

	// Generate recommendation
	recommendation := s.generateRecommendation(multiplier, confidence, expectedROI)

	return &DynamicBidResult{
		OriginalBid:     baseBid,
		AdjustedBid:     adjustedBid,
		Multiplier:      multiplier,
		Confidence:      confidence,
		Factors:         factors,
		Recommendation:  recommendation,
		ExpectedWinRate: expectedWinRate,
		ExpectedROI:     expectedROI,
	}
}

func (s *DynamicBidService) getDeviceMultiplier(deviceType string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if mult, ok := s.deviceMultipliers[deviceType]; ok {
		return mult
	}
	return 1.0
}

func (s *DynamicBidService) getPublisherFactor(publisherID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if stats, ok := s.publisherStats[publisherID]; ok {
		// Higher quality publishers get higher bids
		qualityBonus := stats.quality * 0.3
		// Adjust based on win rate (bid more if win rate is low but quality is high)
		winRateAdjust := 1.0
		if stats.winRate < 0.2 && stats.quality > 0.7 {
			winRateAdjust = 1.15 // Bid more aggressively
		} else if stats.winRate > 0.5 {
			winRateAdjust = 0.95 // Can afford to bid less
		}
		return (1.0 + qualityBonus) * winRateAdjust
	}
	return 1.0 // No data, use base bid
}

func (s *DynamicBidService) getContextFactor(campaignID string, req *model.BidRequest) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	contextKey := s.buildContextKey(campaignID, req)
	if stats, ok := s.contextPerformance[contextKey]; ok {
		if stats.impressions < 100 {
			return 1.0 // Not enough data
		}

		// Calculate performance score
		ctr := float64(stats.clicks) / float64(stats.impressions)
		cvr := 0.0
		if stats.clicks > 0 {
			cvr = float64(stats.conversions) / float64(stats.clicks)
		}

		// ROI calculation
		roi := 0.0
		if stats.spend > 0 {
			roi = (stats.revenue - stats.spend) / stats.spend
		}

		// Score based on performance
		score := 1.0
		if ctr > 0.02 {
			score += 0.2 // Good CTR bonus
		}
		if cvr > 0.05 {
			score += 0.3 // Good CVR bonus
		}
		if roi > 0.5 {
			score += 0.2 // Good ROI bonus
		} else if roi < -0.2 {
			score -= 0.3 // Poor ROI penalty
		}

		return score
	}
	return 1.0
}

func (s *DynamicBidService) getCompetitionFactor(req *model.BidRequest) float64 {
	// Estimate competition level based on ad slot characteristics
	factor := 1.0

	// Premium inventory typically has more competition
	if len(req.AdSlot.Formats) > 0 {
		for _, format := range req.AdSlot.Formats {
			switch format {
			case "video":
				factor += 0.15 // Video is competitive
			case "native":
				factor += 0.10 // Native is competitive
			case "banner":
				factor += 0.0 // Standard competition
			}
		}
	}

	// Larger ad slots often have more competition
	// Dimensions is [width, height]
	if len(req.AdSlot.Dimensions) >= 2 && req.AdSlot.Dimensions[0] >= 300 && req.AdSlot.Dimensions[1] >= 250 {
		factor += 0.05
	}

	return factor
}

func (s *DynamicBidService) getGoalFactor(campaign *model.Campaign) float64 {
	// Adjust based on campaign goal type
	switch campaign.GoalType {
	case "CPA", "CPL":
		// Performance campaigns - be more conservative
		return 0.9
	case "CPM", "vCPM":
		// Awareness campaigns - can bid more freely
		return 1.1
	case "CPC":
		// Click campaigns - moderate
		return 1.0
	case "CPCV", "CPV":
		// Video completion - premium
		return 1.15
	default:
		return 1.0
	}
}

func (s *DynamicBidService) calculateWeightedMultiplier(factors map[string]float64) float64 {
	// Weighted average of all factors
	weights := map[string]float64{
		"time":        0.15,
		"device":      0.15,
		"publisher":   0.25,
		"context":     0.25,
		"competition": 0.10,
		"goal":        0.10,
	}

	totalWeight := 0.0
	weightedSum := 0.0

	for factor, value := range factors {
		if weight, ok := weights[factor]; ok {
			weightedSum += value * weight
			totalWeight += weight
		}
	}

	if totalWeight > 0 {
		return weightedSum / totalWeight
	}
	return 1.0
}

func (s *DynamicBidService) clampMultiplier(multiplier float64) float64 {
	if multiplier < s.config.MinBidMultiplier {
		return s.config.MinBidMultiplier
	}
	if multiplier > s.config.MaxBidMultiplier {
		return s.config.MaxBidMultiplier
	}
	return multiplier
}

func (s *DynamicBidService) calculateConfidence(_, publisherID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	samples := int64(0)

	// Check publisher data
	if stats, ok := s.publisherStats[publisherID]; ok {
		samples += stats.samples
	}

	// Calculate confidence based on sample size
	// Confidence approaches 1.0 as samples increase
	if samples == 0 {
		return 0.1 // Low confidence with no data
	}

	confidence := 1.0 - math.Exp(-float64(samples)/1000.0)
	return math.Min(confidence, 0.95)
}

func (s *DynamicBidService) predictWinRate(bidPrice float64, publisherID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if stats, ok := s.publisherStats[publisherID]; ok {
		if stats.avgWinPrice > 0 {
			// Win rate increases as bid approaches/exceeds average win price
			ratio := bidPrice / stats.avgWinPrice
			// Sigmoid function for smooth win rate prediction
			return 1.0 / (1.0 + math.Exp(-2.0*(ratio-1.0)))
		}
	}

	// Default estimate based on bid level
	if bidPrice > 3.0 {
		return 0.7
	} else if bidPrice > 2.0 {
		return 0.5
	} else if bidPrice > 1.0 {
		return 0.3
	}
	return 0.15
}

func (s *DynamicBidService) predictROI(_ *model.Campaign, bidPrice float64, winRate float64) float64 {
	// Simplified ROI prediction
	// ROI = (Expected Revenue - Cost) / Cost

	estimatedCTR := 0.02   // 2% baseline CTR
	estimatedCVR := 0.05   // 5% baseline CVR
	estimatedValue := 10.0 // $10 average conversion value

	expectedRevenue := winRate * estimatedCTR * estimatedCVR * estimatedValue
	cost := bidPrice / 1000.0 // CPM to per-impression cost

	if cost > 0 {
		return (expectedRevenue - cost) / cost
	}
	return 0.0
}

func (s *DynamicBidService) generateRecommendation(multiplier, confidence, expectedROI float64) string {
	if confidence < 0.3 {
		return "insufficient_data"
	}

	if expectedROI > 0.5 && multiplier < 1.5 {
		return "increase_bid"
	} else if expectedROI < -0.2 {
		return "decrease_bid"
	} else if multiplier > 1.3 && expectedROI > 0.2 {
		return "aggressive_bid"
	} else if multiplier < 0.8 {
		return "conservative_bid"
	}

	return "maintain_bid"
}

func (s *DynamicBidService) buildContextKey(campaignID string, req *model.BidRequest) string {
	return campaignID + ":" + req.PublisherID + ":" + req.Device.Type
}

// RecordOutcome records the outcome of a bid for learning
func (s *DynamicBidService) RecordOutcome(campaignID string, req *model.BidRequest, bidPrice, winPrice float64, won bool, clicked bool, converted bool, revenue float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update context performance
	contextKey := s.buildContextKey(campaignID, req)
	if _, exists := s.contextPerformance[contextKey]; !exists {
		s.contextPerformance[contextKey] = &contextStats{}
	}

	stats := s.contextPerformance[contextKey]
	stats.bids++
	stats.totalBidPrice += bidPrice

	if won {
		stats.wins++
		stats.impressions++
		stats.totalWinPrice += winPrice
		stats.spend += winPrice / 1000.0 // CPM to actual cost

		if clicked {
			stats.clicks++
		}
		if converted {
			stats.conversions++
			stats.revenue += revenue
		}
	}
	stats.lastUpdated = time.Now()

	// Update publisher stats
	s.updatePublisherStats(req.PublisherID, won, winPrice, clicked, converted)
}

func (s *DynamicBidService) updatePublisherStats(publisherID string, won bool, winPrice float64, clicked, converted bool) {
	if _, exists := s.publisherStats[publisherID]; !exists {
		s.publisherStats[publisherID] = &publisherPerf{
			quality: 0.5, // Start neutral
		}
	}

	stats := s.publisherStats[publisherID]
	stats.samples++

	// Update win rate (exponential moving average)
	alpha := 0.01 // Learning rate
	wonVal := 0.0
	if won {
		wonVal = 1.0
	}
	stats.winRate = stats.winRate*(1-alpha) + wonVal*alpha

	// Update average win price
	if won && winPrice > 0 {
		if stats.avgWinPrice == 0 {
			stats.avgWinPrice = winPrice
		} else {
			stats.avgWinPrice = stats.avgWinPrice*(1-alpha) + winPrice*alpha
		}
	}

	// Update quality based on engagement
	if won {
		engagementBonus := 0.0
		if clicked {
			engagementBonus += 0.02
		}
		if converted {
			engagementBonus += 0.05
		}
		stats.quality = math.Min(1.0, stats.quality*(1-alpha)+engagementBonus+0.5*alpha)
	}
}

// UpdateHourlyMultiplier updates the multiplier for a specific hour based on performance
func (s *DynamicBidService) UpdateHourlyMultiplier(hour int, performanceScore float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if hour >= 0 && hour < 24 {
		// Blend with existing multiplier
		alpha := 0.1
		current := s.hourlyMultipliers[hour]
		s.hourlyMultipliers[hour] = current*(1-alpha) + performanceScore*alpha
	}
}

// UpdateDeviceMultiplier updates the multiplier for a device type
func (s *DynamicBidService) UpdateDeviceMultiplier(deviceType string, performanceScore float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	alpha := 0.1
	if current, exists := s.deviceMultipliers[deviceType]; exists {
		s.deviceMultipliers[deviceType] = current*(1-alpha) + performanceScore*alpha
	} else {
		s.deviceMultipliers[deviceType] = performanceScore
	}
}

// GetBidAnalytics returns analytics about bidding performance
func (s *DynamicBidService) GetBidAnalytics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalBids := int64(0)
	totalWins := int64(0)
	totalSpend := 0.0
	totalRevenue := 0.0

	for _, stats := range s.contextPerformance {
		totalBids += stats.bids
		totalWins += stats.wins
		totalSpend += stats.spend
		totalRevenue += stats.revenue
	}

	winRate := 0.0
	if totalBids > 0 {
		winRate = float64(totalWins) / float64(totalBids)
	}

	roi := 0.0
	if totalSpend > 0 {
		roi = (totalRevenue - totalSpend) / totalSpend
	}

	return map[string]interface{}{
		"total_bids":         totalBids,
		"total_wins":         totalWins,
		"win_rate":           winRate,
		"total_spend":        totalSpend,
		"total_revenue":      totalRevenue,
		"roi":                roi,
		"contexts_tracked":   len(s.contextPerformance),
		"publishers_tracked": len(s.publisherStats),
		"hourly_multipliers": s.hourlyMultipliers,
		"device_multipliers": s.deviceMultipliers,
	}
}

// SetConfig updates the dynamic bid configuration
func (s *DynamicBidService) SetConfig(config *DynamicBidConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// GetConfig returns the current configuration
func (s *DynamicBidService) GetConfig() *DynamicBidConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}
