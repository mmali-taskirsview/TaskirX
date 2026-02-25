package service

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// BidLandscapeService provides bid distribution analysis and optimization
type BidLandscapeService struct {
	cache      cache.Cache
	mu         sync.RWMutex
	landscapes map[string]*landscapeData // key: placement+geo hash
}

type landscapeData struct {
	bids         []bidRecord
	lastAnalysis time.Time
	cachedResult *model.BidLandscapeResult
}

type bidRecord struct {
	bidPrice   float64
	clearPrice float64
	won        bool
	timestamp  time.Time
	placement  string
	deviceType string
}

// NewBidLandscapeService creates a new bid landscape service
func NewBidLandscapeService(c cache.Cache) *BidLandscapeService {
	return &BidLandscapeService{
		cache:      c,
		landscapes: make(map[string]*landscapeData),
	}
}

// AnalyzeLandscape performs bid landscape analysis for optimal bidding
func (s *BidLandscapeService) AnalyzeLandscape(campaign *model.Campaign, req *model.BidRequest) *model.BidLandscapeResult {
	config := campaign.Targeting.BidLandscape
	if config == nil || !config.Enabled {
		return &model.BidLandscapeResult{
			Analyzed:      false,
			BidMultiplier: 1.0,
			Reason:        "bid_landscape_disabled",
		}
	}

	// Generate landscape key based on placement characteristics
	landscapeKey := s.generateLandscapeKey(req)

	s.mu.RLock()
	data, exists := s.landscapes[landscapeKey]
	s.mu.RUnlock()

	if !exists || len(data.bids) < config.MinSampleSize {
		return &model.BidLandscapeResult{
			Analyzed:        false,
			SampleSize:      0,
			BidMultiplier:   1.0,
			Confidence:      0,
			MarketCondition: "unknown",
			Reason:          "insufficient_data",
		}
	}

	// Use cached result if recent
	if data.cachedResult != nil && time.Since(data.lastAnalysis) < 15*time.Minute {
		return data.cachedResult
	}

	// Perform analysis
	result := s.performAnalysis(data, campaign, config)

	// Cache result
	s.mu.Lock()
	data.cachedResult = result
	data.lastAnalysis = time.Now()
	s.mu.Unlock()

	return result
}

func (s *BidLandscapeService) performAnalysis(data *landscapeData, campaign *model.Campaign, config *model.BidLandscape) *model.BidLandscapeResult {
	// Filter to analysis window
	windowHours := config.AnalysisWindow
	if windowHours <= 0 {
		windowHours = 24
	}
	cutoff := time.Now().Add(-time.Duration(windowHours) * time.Hour)

	var recentBids []bidRecord
	for _, b := range data.bids {
		if b.timestamp.After(cutoff) {
			recentBids = append(recentBids, b)
		}
	}

	if len(recentBids) < config.MinSampleSize {
		return &model.BidLandscapeResult{
			Analyzed:        false,
			SampleSize:      len(recentBids),
			BidMultiplier:   1.0,
			MarketCondition: "unknown",
			Reason:          "insufficient_recent_data",
		}
	}

	// Sort bids by price
	sort.Slice(recentBids, func(i, j int) bool {
		return recentBids[i].bidPrice < recentBids[j].bidPrice
	})

	// Calculate percentiles
	percentiles := s.calculatePercentiles(recentBids)

	// Find optimal bid range
	optimalRange := s.findOptimalRange(recentBids, percentiles)

	// Determine market condition
	marketCondition := s.assessMarketCondition(recentBids, percentiles)

	// Calculate recommended bid
	recommendedBid, multiplier := s.calculateRecommendedBid(campaign.BidPrice, optimalRange, marketCondition)

	// Calculate confidence based on sample size and consistency
	confidence := s.calculateConfidence(len(recentBids), config.MinSampleSize, recentBids)

	return &model.BidLandscapeResult{
		Analyzed:        true,
		SampleSize:      len(recentBids),
		RecommendedBid:  recommendedBid,
		BidMultiplier:   multiplier,
		Confidence:      confidence,
		OptimalRange:    optimalRange,
		MarketCondition: marketCondition,
		Reason:          "analysis_complete",
	}
}

func (s *BidLandscapeService) calculatePercentiles(bids []bidRecord) []model.BidPercentile {
	percentiles := []int{10, 25, 50, 75, 90}
	result := make([]model.BidPercentile, 0, len(percentiles))

	for _, p := range percentiles {
		idx := (len(bids) * p) / 100
		if idx >= len(bids) {
			idx = len(bids) - 1
		}

		// Calculate win rate at this percentile
		winsAtOrAbove := 0
		totalAtOrAbove := 0
		var clearPrices []float64

		threshold := bids[idx].bidPrice
		for _, b := range bids {
			if b.bidPrice >= threshold {
				totalAtOrAbove++
				if b.won {
					winsAtOrAbove++
					clearPrices = append(clearPrices, b.clearPrice)
				}
			}
		}

		winRate := 0.0
		if totalAtOrAbove > 0 {
			winRate = float64(winsAtOrAbove) / float64(totalAtOrAbove)
		}

		avgClear := 0.0
		if len(clearPrices) > 0 {
			for _, cp := range clearPrices {
				avgClear += cp
			}
			avgClear /= float64(len(clearPrices))
		}

		result = append(result, model.BidPercentile{
			Percentile:  p,
			BidPrice:    threshold,
			WinRate:     winRate,
			AvgClearPrc: avgClear,
		})
	}

	return result
}

func (s *BidLandscapeService) findOptimalRange(_ []bidRecord, percentiles []model.BidPercentile) *model.OptimalBidRange {
	if len(percentiles) < 3 {
		return nil
	}

	// Find the sweet spot: highest win rate / bid price ratio
	var bestEfficiency float64
	var sweetSpot float64
	var expectedWinRate float64

	for _, p := range percentiles {
		if p.BidPrice > 0 {
			efficiency := p.WinRate / p.BidPrice
			if efficiency > bestEfficiency {
				bestEfficiency = efficiency
				sweetSpot = p.BidPrice
				expectedWinRate = p.WinRate
			}
		}
	}

	// Min bid: 25th percentile (competitive floor)
	minBid := percentiles[1].BidPrice
	// Max bid: 90th percentile (diminishing returns)
	maxBid := percentiles[len(percentiles)-1].BidPrice

	// Expected CPM at sweet spot
	expectedCPM := sweetSpot * 1000 / expectedWinRate
	if expectedWinRate <= 0 {
		expectedCPM = 0
	}

	return &model.OptimalBidRange{
		MinBid:          minBid,
		MaxBid:          maxBid,
		SweetSpot:       sweetSpot,
		ExpectedWinRate: expectedWinRate,
		ExpectedCPM:     expectedCPM,
	}
}

func (s *BidLandscapeService) assessMarketCondition(bids []bidRecord, percentiles []model.BidPercentile) string {
	if len(percentiles) < 3 {
		return "unknown"
	}

	// Calculate bid spread
	p25 := percentiles[1].BidPrice
	p75 := percentiles[3].BidPrice

	spread := 0.0
	if p25 > 0 {
		spread = (p75 - p25) / p25
	}

	// Calculate overall win rate
	wins := 0
	for _, b := range bids {
		if b.won {
			wins++
		}
	}
	winRate := float64(wins) / float64(len(bids))

	// Determine market condition
	if winRate > 0.5 && spread < 0.3 {
		return "soft" // Easy to win, low competition
	} else if winRate < 0.2 || spread > 0.8 {
		return "aggressive" // Hard to win, high competition
	}
	return "competitive" // Normal competition
}

func (s *BidLandscapeService) calculateRecommendedBid(currentBid float64, optimalRange *model.OptimalBidRange, marketCondition string) (float64, float64) {
	if optimalRange == nil {
		return currentBid, 1.0
	}

	recommended := optimalRange.SweetSpot

	// Adjust based on market condition
	switch marketCondition {
	case "soft":
		// Can bid lower in soft markets
		recommended = optimalRange.MinBid + (optimalRange.SweetSpot-optimalRange.MinBid)*0.5
	case "aggressive":
		// Need to bid higher in aggressive markets
		recommended = optimalRange.SweetSpot + (optimalRange.MaxBid-optimalRange.SweetSpot)*0.3
	}

	// Calculate multiplier
	multiplier := 1.0
	if currentBid > 0 {
		multiplier = recommended / currentBid
		// Cap multiplier to reasonable range
		if multiplier > 2.0 {
			multiplier = 2.0
		}
		if multiplier < 0.5 {
			multiplier = 0.5
		}
	}

	return recommended, multiplier
}

func (s *BidLandscapeService) calculateConfidence(sampleSize, minSampleSize int, bids []bidRecord) float64 {
	// Base confidence on sample size
	sizeRatio := float64(sampleSize) / float64(minSampleSize*5)
	if sizeRatio > 1.0 {
		sizeRatio = 1.0
	}

	// Adjust for data consistency (standard deviation)
	if len(bids) < 2 {
		return sizeRatio * 0.5
	}

	var sum, sumSq float64
	for _, b := range bids {
		sum += b.bidPrice
		sumSq += b.bidPrice * b.bidPrice
	}
	mean := sum / float64(len(bids))
	variance := (sumSq / float64(len(bids))) - (mean * mean)
	stdDev := math.Sqrt(variance)

	// Lower consistency if high variance
	consistencyFactor := 1.0
	if mean > 0 {
		cv := stdDev / mean // Coefficient of variation
		if cv > 0.5 {
			consistencyFactor = 0.7
		} else if cv > 0.3 {
			consistencyFactor = 0.85
		}
	}

	return sizeRatio * consistencyFactor
}

func (s *BidLandscapeService) generateLandscapeKey(req *model.BidRequest) string {
	// Create key from placement characteristics
	key := req.PublisherID
	if req.Device.Type != "" {
		key += "_" + req.Device.Type
	}
	if req.Device.Geo.Country != "" {
		key += "_" + req.Device.Geo.Country
	}
	return key
}

// RecordBid records a bid for landscape analysis
func (s *BidLandscapeService) RecordBid(req *model.BidRequest, bidPrice, clearPrice float64, won bool) {
	landscapeKey := s.generateLandscapeKey(req)

	record := bidRecord{
		bidPrice:   bidPrice,
		clearPrice: clearPrice,
		won:        won,
		timestamp:  time.Now(),
		placement:  req.PublisherID,
		deviceType: req.Device.Type,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.landscapes[landscapeKey]; !exists {
		s.landscapes[landscapeKey] = &landscapeData{
			bids: make([]bidRecord, 0, 1000),
		}
	}

	data := s.landscapes[landscapeKey]
	data.bids = append(data.bids, record)

	// Keep only last 10000 bids
	if len(data.bids) > 10000 {
		data.bids = data.bids[len(data.bids)-10000:]
	}

	// Invalidate cache
	data.cachedResult = nil
}
