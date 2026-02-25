package service

import (
	"math"
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// CompetitiveIntelligenceService tracks competitor behavior and adjusts bids
type CompetitiveIntelligenceService struct {
	cache            cache.Cache
	mu               sync.RWMutex
	competitorData   map[string]*localCompetitorProfile // key: competitorID
	segmentBidFloors map[string]float64                 // key: segment identifier
	auctionHistory   []auctionOutcome
}

type localCompetitorProfile struct {
	CompetitorID     string
	DomainPatterns   []string
	AvgBidPrice      float64
	BidVolume        int
	WinRate          float64
	PeakHours        []int
	PreferredFormats []string
	LastSeen         time.Time
	BidHistory       []float64
	TrendDirection   string // "increasing", "decreasing", "stable"
}

type auctionOutcome struct {
	Timestamp     time.Time
	OurBid        float64
	WinningBid    float64
	Won           bool
	CompetitorID  string
	InventoryType string
	SegmentKey    string
}

// NewCompetitiveIntelligenceService creates a new competitive intelligence service
func NewCompetitiveIntelligenceService(c cache.Cache) *CompetitiveIntelligenceService {
	return &CompetitiveIntelligenceService{
		cache:            c,
		competitorData:   make(map[string]*localCompetitorProfile),
		segmentBidFloors: make(map[string]float64),
		auctionHistory:   make([]auctionOutcome, 0, 10000),
	}
}

// AnalyzeCompetition analyzes competitive landscape and recommends bid adjustments
func (s *CompetitiveIntelligenceService) AnalyzeCompetition(campaign *model.Campaign, request *model.BidRequest) *model.CompetitiveIntelResult {
	config := campaign.Targeting.CompetitiveIntelligence
	if config == nil || !config.Enabled {
		return &model.CompetitiveIntelResult{
			Analyzed:      false,
			BidAdjustment: 1.0,
		}
	}

	result := &model.CompetitiveIntelResult{
		Analyzed:        true,
		BidAdjustment:   1.0,
		MarketCondition: "unknown",
	}

	// Calculate segment key for this inventory
	segmentKey := s.getSegmentKey(request)

	// Get competitive metrics for this segment
	s.analyzeSegmentCompetition(segmentKey, result)

	// Track known competitors if configured
	if len(config.TrackCompetitors) > 0 {
		s.analyzeKnownCompetitors(config.TrackCompetitors, result)
	}

	// Generate bid adjustment based on competitive mode
	result.BidAdjustment = s.calculateBidMultiplier(config, result)

	// Market share info
	result.OurShareOfVoice = s.calculateMarketShare()

	// Determine recommended action
	result.RecommendedAction = s.determineRecommendedAction(config, result)

	return result
}

func (s *CompetitiveIntelligenceService) getSegmentKey(request *model.BidRequest) string {
	// Create segment key based on inventory characteristics
	key := "pub_" + request.PublisherID

	if request.AdSlot.ID != "" {
		key += "_slot_" + request.AdSlot.ID
	}

	if len(request.AdSlot.Formats) > 0 {
		key += "_" + request.AdSlot.Formats[0]
	}

	return key
}

func (s *CompetitiveIntelligenceService) analyzeSegmentCompetition(segmentKey string, result *model.CompetitiveIntelResult) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter auctions for this segment
	var segmentAuctions []auctionOutcome
	recentCutoff := time.Now().Add(-24 * time.Hour)

	for _, auction := range s.auctionHistory {
		if auction.SegmentKey == segmentKey && auction.Timestamp.After(recentCutoff) {
			segmentAuctions = append(segmentAuctions, auction)
		}
	}

	if len(segmentAuctions) == 0 {
		return
	}

	// Calculate metrics
	var wins int
	competitorSet := make(map[string]bool)

	for _, auction := range segmentAuctions {
		if auction.Won {
			wins++
		}
		if auction.CompetitorID != "" {
			competitorSet[auction.CompetitorID] = true
		}
	}

	result.CompetitorsActive = len(competitorSet)

	// Determine market condition based on competition level
	bidSpread := s.calculateBidSpread(segmentAuctions)
	if bidSpread < 0.1 || result.CompetitorsActive > 5 {
		result.MarketCondition = "high"
	} else if bidSpread < 0.25 || result.CompetitorsActive > 2 {
		result.MarketCondition = "medium"
	} else {
		result.MarketCondition = "low"
	}
}

func (s *CompetitiveIntelligenceService) calculateBidSpread(auctions []auctionOutcome) float64 {
	if len(auctions) < 2 {
		return 0
	}

	var sum, mean float64
	for _, a := range auctions {
		sum += a.WinningBid
	}
	mean = sum / float64(len(auctions))

	if mean == 0 {
		return 0
	}

	// Calculate coefficient of variation
	var variance float64
	for _, a := range auctions {
		variance += (a.WinningBid - mean) * (a.WinningBid - mean)
	}
	variance /= float64(len(auctions))

	return math.Sqrt(variance) / mean
}

func (s *CompetitiveIntelligenceService) analyzeKnownCompetitors(competitorIDs []string, result *model.CompetitiveIntelResult) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profiles := make([]model.CompetitorProfile, 0, len(competitorIDs))
	var leadingCompetitor string
	var highestWinRate float64

	for _, compID := range competitorIDs {
		if localProfile, exists := s.competitorData[compID]; exists {
			profile := model.CompetitorProfile{
				ID:             compID,
				AvgBidPrice:    localProfile.AvgBidPrice,
				WinRateAgainst: localProfile.WinRate,
				PeakHours:      localProfile.PeakHours,
				LastSeen:       localProfile.LastSeen,
				BiddingPattern: localProfile.TrendDirection,
			}

			profiles = append(profiles, profile)

			// Track leading competitor
			if localProfile.WinRate > highestWinRate {
				highestWinRate = localProfile.WinRate
				leadingCompetitor = compID
			}
		}
	}

	result.CompetitorProfiles = profiles
	result.LeadingCompetitor = leadingCompetitor
}

func (s *CompetitiveIntelligenceService) calculateBidMultiplier(config *model.CompetitiveIntelligence, result *model.CompetitiveIntelResult) float64 {
	multiplier := 1.0

	// Adjust based on market condition
	switch result.MarketCondition {
	case "high":
		multiplier *= 1.15 // Bid higher in very competitive segments
	case "medium":
		multiplier *= 1.05
	case "low":
		multiplier *= 0.95 // Can bid lower in less competitive segments
	}

	// Adjust based on competitive mode
	switch config.CompetitiveMode {
	case "aggressive":
		multiplier *= 1.2
	case "defensive":
		multiplier *= 0.85
	case "balanced":
		// No change
	}

	// Adjust based on market share goal
	if config.MarketShareGoal > 0 && result.OurShareOfVoice > 0 {
		if result.OurShareOfVoice < config.MarketShareGoal {
			// Below target - increase bids
			gap := config.MarketShareGoal - result.OurShareOfVoice
			multiplier *= (1.0 + gap*0.5)
		} else if result.OurShareOfVoice > config.MarketShareGoal*1.2 {
			// Above target - can reduce
			excess := result.OurShareOfVoice - config.MarketShareGoal
			multiplier *= (1.0 - excess*0.3)
		}
	}

	// Apply caps
	if multiplier > 1.5 {
		multiplier = 1.5
	}
	if multiplier < 0.7 {
		multiplier = 0.7
	}

	return multiplier
}

func (s *CompetitiveIntelligenceService) determineRecommendedAction(config *model.CompetitiveIntelligence, result *model.CompetitiveIntelResult) string {
	if result.MarketCondition == "high" && result.OurShareOfVoice < 0.1 {
		return "increase_budget"
	}

	if result.CompetitorsActive > 5 {
		return "optimize_targeting"
	}

	if config.MarketShareGoal > 0 && result.OurShareOfVoice < config.MarketShareGoal*0.5 {
		return "aggressive_bidding"
	}

	if result.MarketCondition == "low" {
		return "maintain_position"
	}

	return "no_change"
}

func (s *CompetitiveIntelligenceService) calculateMarketShare() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate based on recent win rate
	recentCutoff := time.Now().Add(-24 * time.Hour)
	var wins, total int

	for _, auction := range s.auctionHistory {
		if auction.Timestamp.After(recentCutoff) {
			total++
			if auction.Won {
				wins++
			}
		}
	}

	if total == 0 {
		return 0
	}

	return float64(wins) / float64(total)
}

// RecordAuctionOutcome records the result of an auction for learning
func (s *CompetitiveIntelligenceService) RecordAuctionOutcome(request *model.BidRequest, ourBid, winningBid float64, won bool, competitorID string) {
	outcome := auctionOutcome{
		Timestamp:    time.Now(),
		OurBid:       ourBid,
		WinningBid:   winningBid,
		Won:          won,
		CompetitorID: competitorID,
		SegmentKey:   s.getSegmentKey(request),
	}

	if len(request.AdSlot.Formats) > 0 {
		outcome.InventoryType = request.AdSlot.Formats[0]
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.auctionHistory = append(s.auctionHistory, outcome)

	// Maintain max history size
	maxHistory := 10000
	if len(s.auctionHistory) > maxHistory {
		s.auctionHistory = s.auctionHistory[len(s.auctionHistory)-maxHistory:]
	}

	// Update segment bid floors
	if won {
		s.segmentBidFloors[outcome.SegmentKey] = ourBid
	} else if winningBid > 0 {
		// Smoothly update floor based on winning bid
		existingFloor := s.segmentBidFloors[outcome.SegmentKey]
		if existingFloor > 0 {
			s.segmentBidFloors[outcome.SegmentKey] = existingFloor*0.8 + winningBid*0.2
		} else {
			s.segmentBidFloors[outcome.SegmentKey] = winningBid
		}
	}

	// Update competitor profile if known
	if competitorID != "" && !won {
		s.updateCompetitorProfile(competitorID, winningBid)
	}
}

func (s *CompetitiveIntelligenceService) updateCompetitorProfile(competitorID string, bidPrice float64) {
	// Called with lock already held

	if _, exists := s.competitorData[competitorID]; !exists {
		s.competitorData[competitorID] = &localCompetitorProfile{
			CompetitorID: competitorID,
			BidHistory:   make([]float64, 0, 100),
		}
	}

	profile := s.competitorData[competitorID]
	profile.BidHistory = append(profile.BidHistory, bidPrice)
	profile.BidVolume++
	profile.LastSeen = time.Now()

	// Maintain history size
	if len(profile.BidHistory) > 100 {
		profile.BidHistory = profile.BidHistory[len(profile.BidHistory)-100:]
	}

	// Recalculate average
	var sum float64
	for _, b := range profile.BidHistory {
		sum += b
	}
	profile.AvgBidPrice = sum / float64(len(profile.BidHistory))

	// Determine trend
	if len(profile.BidHistory) >= 10 {
		recent := profile.BidHistory[len(profile.BidHistory)-5:]
		older := profile.BidHistory[len(profile.BidHistory)-10 : len(profile.BidHistory)-5]

		var recentSum, olderSum float64
		for _, b := range recent {
			recentSum += b
		}
		for _, b := range older {
			olderSum += b
		}

		recentAvg := recentSum / 5
		olderAvg := olderSum / 5

		if recentAvg > olderAvg*1.1 {
			profile.TrendDirection = "increasing"
		} else if recentAvg < olderAvg*0.9 {
			profile.TrendDirection = "decreasing"
		} else {
			profile.TrendDirection = "stable"
		}
	}
}

// GetCompetitorProfile returns detailed profile of a competitor
func (s *CompetitiveIntelligenceService) GetCompetitorProfile(competitorID string) (*localCompetitorProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profile, exists := s.competitorData[competitorID]
	return profile, exists
}

// GetSegmentFloor returns the recommended bid floor for a segment
func (s *CompetitiveIntelligenceService) GetSegmentFloor(segmentKey string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.segmentBidFloors[segmentKey]
}

// GetMarketReport generates a market report
func (s *CompetitiveIntelligenceService) GetMarketReport() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report := make(map[string]interface{})

	// Total auctions
	report["total_auctions"] = len(s.auctionHistory)

	// Win rate
	var wins int
	for _, a := range s.auctionHistory {
		if a.Won {
			wins++
		}
	}
	if len(s.auctionHistory) > 0 {
		report["overall_win_rate"] = float64(wins) / float64(len(s.auctionHistory))
	}

	// Active competitors
	report["tracked_competitors"] = len(s.competitorData)

	// Top segments
	segmentCounts := make(map[string]int)
	for _, a := range s.auctionHistory {
		segmentCounts[a.SegmentKey]++
	}
	report["segment_counts"] = segmentCounts

	return report
}
