package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
	"github.com/taskirx/go-bidding-engine/pkg/metrics"
)

// BiddingService handles bid logic
type BiddingService struct {
	cache          cache.Cache
	backendBaseURL string
	aiServiceURL   string
	fraudServiceURL string
	optServiceURL   string
}

// NewBiddingService creates a new bidding service
func NewBiddingService(cache cache.Cache, backendBaseURL string) *BiddingService {
	// Default to internal docker DNS name if not specified via env in main (which we'll do later)
	// For now, hardcode or accept as param. Modified to accept it? 
	// To minimize changes to main.go right now, I'll default it here, but ideally should be passed.
	
	// Check if we can get it from env in a cleaner way or just iterate constructor
	return &BiddingService{
		cache:          cache,
		backendBaseURL: backendBaseURL,
		// Default to docker service name.
		aiServiceURL:   "http://ad-matching:6002/api", // Updated default port
		fraudServiceURL: "http://fraud-detection:6001/api",
		optServiceURL:   "http://bid-optimizer:6003/api",
	}
}

// SetAIServiceURL allows overriding the AI service URL
func (s *BiddingService) SetAIServiceURL(url string) {
	s.aiServiceURL = url
}

// SetFraudServiceURL allows overriding the Fraud service URL
func (s *BiddingService) SetFraudServiceURL(url string) {
	s.fraudServiceURL = url
}

// SetOptimizationServiceURL allows overriding the Optimization service URL
func (s *BiddingService) SetOptimizationServiceURL(url string) {
	s.optServiceURL = url
}

// BackendBaseURL returns the configured backend API base URL.
func (s *BiddingService) BackendBaseURL() string {
	return s.backendBaseURL
}

// ProcessBid processes a bid request and returns the best bid
func (s *BiddingService) ProcessBid(req *model.BidRequest) (*model.BidResponse, error) {
	startTime := time.Now()

	// Get active campaigns from cache
	campaigns, err := s.cache.GetActiveCampaigns()
	if err != nil {
		return nil, fmt.Errorf("failed to get campaigns: %w", err)
	}

	if len(campaigns) == 0 {
		return nil, fmt.Errorf("no active campaigns")
	}

	// Find matching campaigns
	var matchingCampaigns []*model.BidResult
	
	// --- AI: Fraud Check (Fail Fast) ---
	isFraud, err := s.callFraudService(req)
	if err == nil && isFraud {
		metrics.FraudBlockedTotal.Inc()
		return nil, fmt.Errorf("request flagged as fraud")
	} else if err != nil {
		fmt.Printf("Warning: Fraud Service call failed: %v\n", err)
	}
	// -----------------------------------

	for _, campaign := range campaigns {
		if campaign.IsMatch(req) {
			score := s.calculateScore(campaign, req)
			if score > 0 {
				matchingCampaigns = append(matchingCampaigns, &model.BidResult{
					Campaign:   campaign,
					Score:      score,
					BidPrice:   campaign.BidPrice,
					MatchScore: score,
				})
			}
		}
	}

	if len(matchingCampaigns) == 0 {
		return nil, fmt.Errorf("no matching campaigns")
	}

	// Sort by score (highest first)
	sort.Slice(matchingCampaigns, func(i, j int) bool {
		return matchingCampaigns[i].Score > matchingCampaigns[j].Score
	})

	// --- AI / ML Re-Ranking Integration ---
	// Attempt to use Python AI service to re-rank the candidates
	// We only do this if we have candidates
	if len(matchingCampaigns) > 1 {
		aiRecommendations, err := s.callAIMatchingService(req)
		if err == nil && len(aiRecommendations) > 0 {
			// Map AI scores back to our candidates
			// Create a map for O(1) lookup
			aiScoreMap := make(map[string]float64)
			for _, rec := range aiRecommendations {
				aiScoreMap[rec.CampaignID] = rec.OverallScore
			}

			// Boost score based on AI recommendation
			// We combine the base logic score with AI score
			for _, result := range matchingCampaigns {
				if aiScore, exists := aiScoreMap[result.Campaign.ID]; exists {
					// AI Score is 0.0-1.0. We multiply our base score.
					// Or we can replace it. Let's apply a boost.
					result.Score = result.Score * (1.0 + aiScore)
				}
			}

			// Re-sort after AI adjustment
			sort.Slice(matchingCampaigns, func(i, j int) bool {
				return matchingCampaigns[i].Score > matchingCampaigns[j].Score
			})
		} else if err != nil {
			// Log error but continue with fallback logic
			metrics.AdMatchErrorsTotal.Inc()
			fmt.Printf("Warning: AI Service call failed: %v\n", err)
		}
	}
	// --------------------------------------

	// Select best campaign
	winner := matchingCampaigns[0]

	// --- AI: Dynamic Bid Optimization ---
	optimizedBid, err := s.callOptimizationService(winner, req)
	if err == nil {
		winner.BidPrice = optimizedBid.RecommendedBid
		fmt.Printf("Optimized Bid: %.4f (Multiplier: %.2f) Reason: %v\n", 
			optimizedBid.RecommendedBid, optimizedBid.BidMultiplier, optimizedBid.Reasoning)
	} else {
		metrics.OptimizationErrorsTotal.Inc()
		fmt.Printf("Warning: Optimization Service call failed: %v\n", err)
	}
	// ------------------------------------

	// Record metrics
	metrics.BidsPlacedTotal.Inc()
	s.cache.IncrementBidCount()
	s.cache.IncrementWinCount()
	
	latency := time.Since(startTime).Milliseconds()
	s.cache.RecordLatency(float64(latency))

	// Build response
	response := &model.BidResponse{
		RequestID:   req.ID,
		CampaignID:  winner.Campaign.ID,
		BidPrice:    winner.BidPrice,
		CreativeURL: winner.Campaign.CreativeURL,
		ImpressionURL: fmt.Sprintf("%s/api/analytics/track/impression?campaign_id=%s&request_id=%s", 
			s.backendBaseURL, winner.Campaign.ID, req.ID),
		ClickURL: fmt.Sprintf("%s/api/analytics/track/click?campaign_id=%s&request_id=%s", 
			s.backendBaseURL, winner.Campaign.ID, req.ID),
		TTL:       300,
		Timestamp: time.Now(),
	}

	return response, nil
}

// calculateScore calculates matching score for a campaign
func (s *BiddingService) calculateScore(campaign *model.Campaign, req *model.BidRequest) float64 {
	score := campaign.BidPrice // Base score is bid price

	// Boost for exact matches
	if len(campaign.Targeting.Countries) > 0 {
		for _, country := range campaign.Targeting.Countries {
			if country == req.User.Country {
				score *= 1.2 // 20% boost for country match
				break
			}
		}
	}

	if len(campaign.Targeting.Devices) > 0 {
		for _, device := range campaign.Targeting.Devices {
			if device == req.Device.Type {
				score *= 1.1 // 10% boost for device match
				break
			}
		}
	}

	// Boost for category overlap
	if len(campaign.Targeting.Categories) > 0 && len(req.User.Categories) > 0 {
		overlap := countOverlap(campaign.Targeting.Categories, req.User.Categories)
		if overlap > 0 {
			score *= (1.0 + float64(overlap)*0.05) // 5% boost per matching category
		}
	}

	// Budget availability check
	remainingBudget := campaign.Budget - campaign.Spent
	if remainingBudget < campaign.BidPrice*100 { // Less than 100 impressions left
		score *= 0.5 // Reduce score for low budget
	}

	// Check Geo Rules
	geoRules, _ := s.cache.GetGeoRules(req.User.Country)
	if geoRules != nil {
		if blocked, ok := geoRules["blocked"].(bool); ok && blocked {
			score = 0 // Blocked country
		} else if boost, ok := geoRules["boost_multiplier"].(float64); ok {
			score *= boost
		}
	}

	// User Segment Check
	userSegments, _ := s.cache.GetUserSegments(req.User.ID)
	if len(userSegments) > 0 && len(campaign.Targeting.Categories) > 0 {
		overlap := countOverlap(campaign.Targeting.Categories, userSegments)
		if overlap > 0 {
			score *= (1.0 + float64(overlap)*0.10) // 10% boost per matching segment
		}
	}

	// Real-Time Budget Check
	// We check the daily spend in Redis which is updated atomically
	dailySpend, err := s.cache.GetCampaignSpend(campaign.ID)
	if err == nil {
		// Assume daily budget is 10% of total budget if not explicitly defined (simplification)
		// in production, DailyBudget should be a field on Campaign
		dailyBudget := campaign.Budget * 0.10 
		if dailySpend >= dailyBudget {
			return 0 // Campaign has exceeded daily budget
		}
		
		// Pacing: if we are near the daily limit, lower the bid score to slow down
		if dailySpend >= dailyBudget*0.90 {
			score *= 0.5
		}
	}

	return score
}

// countOverlap counts overlapping items in two slices
func countOverlap(slice1, slice2 []string) int {
	count := 0
	for _, item1 := range slice1 {
		for _, item2 := range slice2 {
			if item1 == item2 {
				count++
			}
		}
	}
	return count
}

// GetMetrics returns current metrics
func (s *BiddingService) GetMetrics() (map[string]interface{}, error) {
	bidCount, _ := s.cache.GetBidCount()
	winCount, _ := s.cache.GetWinCount()
	avgLatency, _ := s.cache.GetAverageLatency()

	winRate := 0.0
	if bidCount > 0 {
		winRate = float64(winCount) / float64(bidCount) * 100
	}

	return map[string]interface{}{
		"total_bids":       bidCount,
		"total_wins":       winCount,
		"win_rate":         winRate,
		"avg_latency_ms":   avgLatency,
		"timestamp":        time.Now(),
	}, nil
}

// RefreshCampaigns fetches fresh campaigns from backend API
func (s *BiddingService) RefreshCampaigns(backendURL string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fmt.Sprintf("%s/internal/campaigns/active", backendURL))
	if err != nil {
		return fmt.Errorf("failed to fetch campaigns from backend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("backend returned status: %d", resp.StatusCode)
	}

	var campaigns []*model.Campaign
	if err := json.NewDecoder(resp.Body).Decode(&campaigns); err != nil {
		return fmt.Errorf("failed to decode campaigns: %w", err)
	}

	// Transform backend campaigns to bidding engine model if necessary
	// Assuming the structures are compatible or using the same model definition
	return s.cache.SetActiveCampaigns(campaigns)
}

// callFraudService calls the Fraud Detection Service
func (s *BiddingService) callFraudService(req *model.BidRequest) (bool, error) {
	fraudReq := model.FraudCheckRequest{
		RequestID:   req.ID,
		Timestamp:   time.Now(),
		IPAddress:   req.Device.IP,
		PublisherID: req.PublisherID,
		Device: model.FraudDeviceInfo{
			Type:      req.Device.Type,
			OS:        req.Device.OS,
			UserAgent: req.Device.UserAgent,
		},
		Geo: model.FraudGeoInfo{
			Country: req.User.Country,
		},
	}

	jsonData, err := json.Marshal(fraudReq)
	if err != nil {
		return false, err
	}

	client := &http.Client{Timeout: 50 * time.Millisecond}
	resp, err := client.Post(s.fraudServiceURL+"/check", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("fraud service status: %d", resp.StatusCode)
	}

	var fraudResp model.FraudCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&fraudResp); err != nil {
		return false, err
	}

	if fraudResp.RecommendedAction == "block" || fraudResp.IsFraud {
		return true, nil
	}
	return false, nil
}

// callOptimizationService calls the Bid Optimization Service
func (s *BiddingService) callOptimizationService(bid *model.BidResult, req *model.BidRequest) (*model.BidRecommendation, error) {
	// Simplified context construction
	optReq := model.BidOptimizationRequest{
		RequestID: req.ID,
		Timestamp: time.Now(),
		Strategy:  "maximize_conversions",
		Context: model.OptimizationContext{
			CampaignID: bid.Campaign.ID,
			BaseBid:    bid.Campaign.BidPrice,
			HourOfDay:  time.Now().Hour(),
			DayOfWeek:  int(time.Now().Weekday()),
			Performance: model.CampaignPerformance{
				CampaignID: bid.Campaign.ID,
				WinRate:    0.5, // Placeholder: fetch from cache/metrics
				CTR:        0.02, // Placeholder
			},
			Budget: model.BudgetStatus{
				CampaignID:  bid.Campaign.ID,
				DailyBudget: bid.Campaign.Budget / 30.0, // approx
				PacingRatio: 1.0, 
			},
		},
	}

	jsonData, err := json.Marshal(optReq)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 50 * time.Millisecond}
	resp, err := client.Post(s.optServiceURL+"/optimize", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("optimization service status: %d", resp.StatusCode)
	}

	var optResp model.BidRecommendation
	if err := json.NewDecoder(resp.Body).Decode(&optResp); err != nil {
		return nil, err
	}

	return &optResp, nil
}

// callAIMatchingService calls the external Python AI service
func (s *BiddingService) callAIMatchingService(req *model.BidRequest) ([]model.AIAdRecommendation, error) {
	// 1. Construct payload
	aiReq := model.AIMatchRequest{
		RequestID: req.ID,
		Timestamp: time.Now(),
		User: model.AIUserProfile{
			UserID:     req.User.ID,
			Country:    req.User.Country,
			DeviceType: req.Device.Type,
			Categories: req.User.Categories,
		},
		AdSlot: model.AIAdSlotInfo{
			SlotID:     "slot_default", // In real RTB, would come from req.Imp
			Dimensions: []int{300, 250}, // Default MREC
			Format:     "banner",
		},
		Context: model.AICampaignContext{
			PublisherID: req.PublisherID,
		},
		Strategy:   "hybrid",
		MaxResults: 5,
	}

	jsonData, err := json.Marshal(aiReq)
	if err != nil {
		return nil, err
	}

	// 2. Execute Request with specific timeout
	// AI Service expected at /match endpoint
	client := &http.Client{Timeout: 100 * time.Millisecond} 
	resp, err := client.Post(s.aiServiceURL+"/match", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI service returned status: %d", resp.StatusCode)
	}

	// 3. Decode Response
	var aiResp model.AIMatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return nil, err
	}

	return aiResp.Recommendations, nil
}
