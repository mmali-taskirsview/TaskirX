// Place these after type declarations, e.g. after BiddingService or performanceData
package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
	"github.com/taskirx/go-bidding-engine/pkg/metrics"
)

// BiddingService handles bid logic
type BiddingService struct {
	cache           cache.Cache
	backendBaseURL  string
	aiServiceURL    string
	fraudServiceURL string
	optServiceURL   string

	// Circuit Breaker for AI Service
	aiMutex        sync.RWMutex
	aiFailureCount int
	aiLastFailure  time.Time

	// Circuit Breaker for Optimization Service
	optMutex        sync.RWMutex
	optFailureCount int
	optLastFailure  time.Time

	// Supply Path Optimization tracking
	spoEnabled   bool
	spoAnalytics []*model.BidPathAnalytics
	spoMutex     sync.RWMutex
	spoService   *SupplyPathAnalyticsService

	// Advanced Services - Phase 1
	attributionService      *AttributionService
	daypartingService       *DaypartingService
	audienceModelingService *AudienceModelingService

	// Advanced Services - Phase 2 (8 New Features)
	bidLandscapeService     *BidLandscapeService
	creativeOptService      *CreativeOptimizationService
	incrementalityService   *IncrementalityService
	privacySandboxService   *PrivacySandboxService
	contextualAIService     *ContextualAIService
	realTimeAlertService    *RealTimeAlertService
	competitiveIntelService *CompetitiveIntelligenceService
	unifiedIDService        *UnifiedIDService

	// Advanced Services - Phase 3 (ML Features)
	dynamicBidService     *DynamicBidService
	lookalikeService      *LookalikeService
	userClusteringService *UserClusteringService

	// Advanced Services - Phase 4 (ML Features Extended)
	churnPredictionService *ChurnPredictionService
	abTestingService       *ABTestingService

	// Advanced Services - Phase 4 (DCO & Prediction)
	dynamicCreativeService       *DynamicCreativeService
	performancePredictionService *PerformancePredictionService

	// Advanced Services - Phase 5 (S2S & Caching)
	s2sBiddingService *S2SBiddingService
	bidCacheService   *BidCacheService

	// Advanced Services - Phase 5 (PMP & SPO)
	programmaticGuaranteedService *ProgrammaticGuaranteedService
	directPublisherService        *DirectPublisherService
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
		aiServiceURL:    "http://ad-matching:6002/api", // Updated default port
		fraudServiceURL: "http://fraud-detection:6001/api",
		optServiceURL:   "http://bid-optimizer:6003/api",
		spoEnabled:      true, // Enable SPO tracking by default
		spoAnalytics:    make([]*model.BidPathAnalytics, 0),
		spoService:      NewSupplyPathAnalyticsService(cache),
		// Initialize advanced services - Phase 1
		attributionService:      NewAttributionService(cache),
		daypartingService:       NewDaypartingService(cache),
		audienceModelingService: NewAudienceModelingService(cache),
		// Initialize advanced services - Phase 2 (8 New Features)
		bidLandscapeService:     NewBidLandscapeService(cache),
		creativeOptService:      NewCreativeOptimizationService(cache),
		incrementalityService:   NewIncrementalityService(cache),
		privacySandboxService:   NewPrivacySandboxService(cache),
		contextualAIService:     NewContextualAIService(cache),
		realTimeAlertService:    NewRealTimeAlertService(cache),
		competitiveIntelService: NewCompetitiveIntelligenceService(cache),
		unifiedIDService:        NewUnifiedIDService(cache),
		// Initialize advanced services - Phase 3 (ML Features)
		dynamicBidService:     NewDynamicBidService(cache),
		lookalikeService:      NewLookalikeService(cache),
		userClusteringService: NewUserClusteringService(cache),
		// Initialize advanced services - Phase 4 (ML Features Extended)
		churnPredictionService: NewChurnPredictionService(cache),
		abTestingService:       NewABTestingService(cache),
		// Initialize advanced services - Phase 4 (DCO & Prediction)
		dynamicCreativeService:       NewDynamicCreativeService(cache),
		performancePredictionService: NewPerformancePredictionService(cache),
		// Initialize advanced services - Phase 5 (S2S & Caching)
		s2sBiddingService: nil, // Initialize after bidding service is created
		bidCacheService:   NewBidCacheService(nil),
		// Initialize advanced services - Phase 5 (PMP & SPO)
		programmaticGuaranteedService: NewProgrammaticGuaranteedService(cache),
		directPublisherService:        NewDirectPublisherService(cache),
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

// GetBackendBaseURL returns the backend base URL
func (s *BiddingService) GetBackendBaseURL() string {
	return s.backendBaseURL
}

// GetSupplyPathAnalyticsService returns the SPO analytics service
func (s *BiddingService) GetSupplyPathAnalyticsService() *SupplyPathAnalyticsService {
	return s.spoService
}

// GetAttributionService returns the attribution service
func (s *BiddingService) GetAttributionService() *AttributionService {
	return s.attributionService
}

// GetDaypartingService returns the dayparting optimization service
func (s *BiddingService) GetDaypartingService() *DaypartingService {
	return s.daypartingService
}

// GetAudienceModelingService returns the audience modeling service
func (s *BiddingService) GetAudienceModelingService() *AudienceModelingService {
	return s.audienceModelingService
}

// GetBidLandscapeService returns the bid landscape analysis service
func (s *BiddingService) GetBidLandscapeService() *BidLandscapeService {
	return s.bidLandscapeService
}

// GetCreativeOptimizationService returns the creative optimization service
func (s *BiddingService) GetCreativeOptimizationService() *CreativeOptimizationService {
	return s.creativeOptService
}

// GetIncrementalityService returns the incrementality testing service
func (s *BiddingService) GetIncrementalityService() *IncrementalityService {
	return s.incrementalityService
}

// GetPrivacySandboxService returns the privacy sandbox service
func (s *BiddingService) GetPrivacySandboxService() *PrivacySandboxService {
	return s.privacySandboxService
}

// GetContextualAIService returns the contextual AI service
func (s *BiddingService) GetContextualAIService() *ContextualAIService {
	return s.contextualAIService
}

// GetRealTimeAlertService returns the real-time alert service
func (s *BiddingService) GetRealTimeAlertService() *RealTimeAlertService {
	return s.realTimeAlertService
}

// GetCompetitiveIntelligenceService returns the competitive intelligence service
func (s *BiddingService) GetCompetitiveIntelligenceService() *CompetitiveIntelligenceService {
	return s.competitiveIntelService
}

// GetUnifiedIDService returns the unified ID service
func (s *BiddingService) GetUnifiedIDService() *UnifiedIDService {
	return s.unifiedIDService
}

// GetDynamicBidService returns the dynamic bid adjustments service
func (s *BiddingService) GetDynamicBidService() *DynamicBidService {
	return s.dynamicBidService
}

// GetLookalikeService returns the lookalike audience service
func (s *BiddingService) GetLookalikeService() *LookalikeService {
	return s.lookalikeService
}

// GetUserClusteringService returns the user clustering service
func (s *BiddingService) GetUserClusteringService() *UserClusteringService {
	return s.userClusteringService
}

// GetChurnPredictionService returns the churn prediction service
func (s *BiddingService) GetChurnPredictionService() *ChurnPredictionService {
	return s.churnPredictionService
}

// GetABTestingService returns the A/B testing service
func (s *BiddingService) GetABTestingService() *ABTestingService {
	return s.abTestingService
}

// GetDynamicCreativeService returns the dynamic creative optimization service
func (s *BiddingService) GetDynamicCreativeService() *DynamicCreativeService {
	return s.dynamicCreativeService
}

// GetPerformancePredictionService returns the performance prediction service
func (s *BiddingService) GetPerformancePredictionService() *PerformancePredictionService {
	return s.performancePredictionService
}

// GetS2SBiddingService returns the server-to-server bidding service
func (s *BiddingService) GetS2SBiddingService() *S2SBiddingService {
	if s.s2sBiddingService == nil {
		s.s2sBiddingService = NewS2SBiddingService(s)
	}
	return s.s2sBiddingService
}

// GetBidCacheService returns the bid caching service
func (s *BiddingService) GetBidCacheService() *BidCacheService {
	return s.bidCacheService
}

// GetProgrammaticGuaranteedService returns the PG deals service
func (s *BiddingService) GetProgrammaticGuaranteedService() *ProgrammaticGuaranteedService {
	return s.programmaticGuaranteedService
}

// GetDirectPublisherService returns the direct publisher relationships service
func (s *BiddingService) GetDirectPublisherService() *DirectPublisherService {
	return s.directPublisherService
}

// GetSupplyPathAnalytics returns the collected SPO analytics
func (s *BiddingService) GetSupplyPathAnalytics() []*model.BidPathAnalytics {
	s.spoMutex.RLock()
	defer s.spoMutex.RUnlock()

	// Return a copy to avoid race conditions
	analytics := make([]*model.BidPathAnalytics, len(s.spoAnalytics))
	copy(analytics, s.spoAnalytics)
	return analytics
}

// ProcessBid processes a bid request and returns the best bid
func (s *BiddingService) ProcessBid(req *model.BidRequest) (*model.BidResponse, error) {
	startTime := time.Now()

	// Request deduplication check (300 second window)
	isDuplicate, err := s.cache.IsRequestDuplicate(req.ID, 300)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Deduplication check failed: %v\n", err)
	} else if isDuplicate {
		metrics.NoBidTotal.WithLabelValues("duplicate_request").Inc()
		return nil, fmt.Errorf("duplicate request ID: %s", req.ID)
	}

	// Initialize SPO analytics tracking
	analytics := &model.BidPathAnalytics{
		RequestID:   req.ID,
		PublisherID: req.PublisherID,
		AdSlotID:    req.AdSlot.ID,
		Hops:        make([]model.SupplyPathHop, 0),
		Metadata:    make(map[string]interface{}),
		Timestamp:   startTime,
	}

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
	isFraud, err, fraudHop := s.callFraudService(req)
	if err == nil && isFraud {
		// Track the fraud check hop
		if fraudHop != nil {
			fraudHop.Sequence = len(analytics.Hops) + 1
			analytics.Hops = append(analytics.Hops, *fraudHop)
			analytics.TotalHops++
			analytics.TotalLatencyMs += fraudHop.LatencyMs
			analytics.TotalFees += fraudHop.Fee
		}

		metrics.FraudBlockedTotal.Inc()
		s.cache.IncrementPublisherFraud(req.PublisherID)
		return nil, fmt.Errorf("request flagged as fraud")
	} else if err != nil {
		// Track failed fraud check hop
		if fraudHop != nil {
			fraudHop.Sequence = len(analytics.Hops) + 1
			analytics.Hops = append(analytics.Hops, *fraudHop)
			analytics.TotalHops++
			analytics.TotalLatencyMs += fraudHop.LatencyMs
			analytics.TotalFees += fraudHop.Fee
		}

		fmt.Printf("Warning: Fraud Service call failed: %v\n", err)
	} else {
		// Track successful fraud check hop
		if fraudHop != nil {
			fraudHop.Sequence = len(analytics.Hops) + 1
			analytics.Hops = append(analytics.Hops, *fraudHop)
			analytics.TotalHops++
			analytics.TotalLatencyMs += fraudHop.LatencyMs
			analytics.TotalFees += fraudHop.Fee
		}
	}
	// -----------------------------------

	// --- Advanced Feature: Unified ID Resolution ---
	// Resolve user identity across multiple ID providers early for better targeting
	var unifiedIDResult *model.UnifiedIDResult
	// We'll apply this per-campaign since it depends on campaign config
	// -------------------------------------------

	for _, campaign := range campaigns {
		if campaign.IsMatch(req) {
			// --- Cross-Device Frequency Cap Check ---
			if campaign.Targeting.CrossDeviceEnabled && campaign.Targeting.FreqCapImpressions > 0 {
				exceeded, _ := s.checkCrossDeviceFreqCap(campaign, req)
				if exceeded {
					continue // User has seen this campaign too many times across all devices
				}
			} else if campaign.Targeting.FreqCapImpressions > 0 && req.User.ID != "" {
				// --- Standard Frequency Cap Check (single device) ---
				windowSecs := campaign.Targeting.FreqCapWindowSecs
				if windowSecs <= 0 {
					windowSecs = 86400 // Default: 24h window
				}
				freq, err := s.cache.GetUserFrequency(req.User.ID, campaign.ID)
				if err == nil && freq >= int64(campaign.Targeting.FreqCapImpressions) {
					continue // User has seen this campaign too many times
				}
			}
			// ---------------------------

			// --- Retargeting Check ---
			if campaign.Targeting.RetargetingMode != "" && req.User.ID != "" {
				userEligible := s.checkRetargetingEligibility(campaign, req.User.ID)
				if campaign.Targeting.RetargetingMode == "include" && !userEligible {
					continue // Only target users with matching events, user doesn't have them
				}
				if campaign.Targeting.RetargetingMode == "exclude" && userEligible {
					continue // Exclude users with matching events, user has them
				}
			}
			// ------------------------

			// --- Dayparting Check ---
			if len(campaign.Targeting.HourSchedule) > 0 {
				currentHour := time.Now().Hour() // 0-23
				hourAllowed := false
				for _, allowedHour := range campaign.Targeting.HourSchedule {
					if currentHour == allowedHour {
						hourAllowed = true
						break
					}
				}
				if !hourAllowed {
					continue // Campaign not active during this hour
				}
			}
			// -----------------------

			// --- Budget Check (Real-Time) ---
			// Check against Total Budget (using lifetime metrics if available)
			// Total Spend = campaign.Spent (historical + synced) + (currentDailySpend - lastSynced)
			// However, we can simply rely on `campaign.Spent` being "up to the last minute".
			// And add the "unsynced delta" from Redis.

			// For simplicity in this iteration:
			// We assume `campaign.Spent` covers historical AND synced daily spend.
			// We fetch `daily_spent` from Redis.
			// We fetch `synced_spent` from Redis (`campaign:spend:synced:<id>:<date>`).
			// RealTimeTotal = campaign.Spent + (daily_spent - synced_spent).

			dateStr := time.Now().Format("2006-01-02")
			dailySpend, err := s.cache.GetCampaignSpend(campaign.ID)
			if err == nil {
				var syncedSpend float64 = 0
				syncedKey := fmt.Sprintf("campaign:spend:synced:%s:%s", campaign.ID, dateStr)
				if val, err := s.cache.Get(syncedKey); err == nil && val != "" {
					fmt.Sscanf(val, "%f", &syncedSpend)
				}

				unsyncedDelta := dailySpend - syncedSpend
				if unsyncedDelta < 0 {
					unsyncedDelta = 0
				}

				totalRealTimeSpend := campaign.Spent + unsyncedDelta

				if totalRealTimeSpend >= campaign.Budget {
					// Skip this campaign, it's over budget
					continue
				}
			}
			// --------------------------------

			score := s.calculateScore(campaign, req)
			if score > 0 {
				bidPrice := campaign.BidPrice
				bidMultiplier := 1.0

				// --- Advanced Feature: Unified ID Resolution ---
				if s.unifiedIDService != nil {
					unifiedIDResult = s.unifiedIDService.ResolveIdentity(campaign, req)
					if unifiedIDResult != nil && unifiedIDResult.Resolved {
						bidMultiplier *= unifiedIDResult.BidMultiplier
					}
				}

				// --- Advanced Feature: Bid Landscape Analysis ---
				if s.bidLandscapeService != nil {
					landscapeResult := s.bidLandscapeService.AnalyzeLandscape(campaign, req)
					if landscapeResult != nil && landscapeResult.Analyzed {
						bidMultiplier *= landscapeResult.BidMultiplier
					}
				}

				// --- Advanced Feature: Contextual AI ---
				if s.contextualAIService != nil {
					contextResult := s.contextualAIService.AnalyzeContext(campaign, req)
					if contextResult != nil && contextResult.Analyzed && contextResult.BrandSafe {
						bidMultiplier *= contextResult.BidMultiplier
					} else if contextResult != nil && !contextResult.BrandSafe {
						continue // Skip non-brand-safe inventory
					}
				}

				// --- Advanced Feature: Competitive Intelligence ---
				if s.competitiveIntelService != nil {
					compResult := s.competitiveIntelService.AnalyzeCompetition(campaign, req)
					if compResult != nil && compResult.Analyzed {
						bidMultiplier *= compResult.BidAdjustment
					}
				}

				// --- Advanced Feature: Incrementality Testing ---
				if s.incrementalityService != nil {
					incResult := s.incrementalityService.EvaluateUser(campaign, req)
					if incResult != nil && incResult.UserInControlGroup {
						continue // Don't bid on control group users
					}
				}

				// --- Advanced Feature: Real-time Alerts ---
				if s.realTimeAlertService != nil {
					dailySpend, _ := s.cache.GetCampaignSpend(campaign.ID)
					alertResult := s.realTimeAlertService.CheckAlerts(campaign, dailySpend, campaign.DailyBudget)
					if alertResult != nil {
						if alertResult.ShouldPauseBid {
							continue // Critical alerts - pause bidding
						}
						bidMultiplier *= alertResult.BidAdjustment
					}
				}

				// Apply combined multiplier to bid price
				bidPrice *= bidMultiplier

				matchingCampaigns = append(matchingCampaigns, &model.BidResult{
					Campaign:   campaign,
					Score:      score,
					BidPrice:   bidPrice,
					MatchScore: score,
				})
				// Record that this campaign entered the auction (bid counter for win-rate denominator)
				go func(cid string) { _ = s.cache.IncrementCampaignBids(cid) }(campaign.ID)
				// Track bid in price bucket for bid landscape analytics
				bucket := getPriceBucket(campaign.BidPrice)
				go func(b string) { _ = s.cache.RecordBidInBucket(b) }(bucket)
				// Track publisher bid floor optimization (bid attempt, not yet known if won)
				// We'll track the win separately when winner is selected
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
		// Circuit Breaker Check
		s.aiMutex.RLock()
		isCircuitOpen := s.aiFailureCount > 5 && time.Since(s.aiLastFailure) < 30*time.Second
		s.aiMutex.RUnlock()

		if !isCircuitOpen {
			aiRecommendations, err, aiHop := s.callAIMatchingService(req)
			if err == nil && len(aiRecommendations) > 0 {
				// Track successful AI matching hop
				if aiHop != nil {
					aiHop.Sequence = len(analytics.Hops) + 1
					analytics.Hops = append(analytics.Hops, *aiHop)
					analytics.TotalHops++
					analytics.TotalLatencyMs += aiHop.LatencyMs
					analytics.TotalFees += aiHop.Fee
				}

				// Success, reset failure count
				s.aiMutex.Lock()
				s.aiFailureCount = 0
				s.aiMutex.Unlock()

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
						result.Score = result.Score * (1.0 + aiScore)
					}
				}

				// Re-sort after AI adjustment
				sort.Slice(matchingCampaigns, func(i, j int) bool {
					return matchingCampaigns[i].Score > matchingCampaigns[j].Score
				})
			} else if err != nil {
				// Track failed AI matching hop
				if aiHop != nil {
					aiHop.Sequence = len(analytics.Hops) + 1
					analytics.Hops = append(analytics.Hops, *aiHop)
					analytics.TotalHops++
					analytics.TotalLatencyMs += aiHop.LatencyMs
					analytics.TotalFees += aiHop.Fee
				}

				// Record failure
				s.aiMutex.Lock()
				s.aiFailureCount++
				s.aiLastFailure = time.Now()
				s.aiMutex.Unlock()

				metrics.AdMatchErrorsTotal.Inc()
				fmt.Printf("Warning: AI Service call failed (Count: %d): %v\n", s.aiFailureCount, err)
			}
		} else {
			// Optional: Increment a "CircuitBreakerOpen" metric if desired
			// fmt.Println("AI Service Circuit Breaker Open")
		}
	}
	// --------------------------------------

	// Select best campaign
	winner := matchingCampaigns[0]

	// --- Advanced Feature: Creative Optimization ---
	// Select the optimal creative variant for this placement
	if s.creativeOptService != nil {
		creativeResult := s.creativeOptService.SelectCreative(winner.Campaign, req)
		if creativeResult != nil && creativeResult.SelectedCreativeID != "" {
			// Store selected creative info in analytics metadata
			analytics.Metadata["selected_creative"] = creativeResult.SelectedCreativeID
			analytics.Metadata["creative_selection_method"] = creativeResult.SelectionMethod
		}
	}

	// --- Advanced Feature: Privacy Sandbox Integration ---
	// Check if Privacy Sandbox APIs should be used
	if s.privacySandboxService != nil {
		psResult := s.privacySandboxService.EvaluatePrivacySandbox(winner.Campaign, req)
		if psResult != nil && psResult.TopicsAvailable {
			analytics.Metadata["privacy_sandbox_enabled"] = true
			analytics.Metadata["topics_available"] = psResult.TopicsAvailable
			analytics.Metadata["topic_match"] = psResult.TopicMatch
			analytics.Metadata["fledge_eligible"] = psResult.FledgeEligible
		}
	}

	// Track losing bids for floor optimization
	go func(pid string, candidates []*model.BidResult, winnerID string) {
		for _, candidate := range candidates {
			if candidate.Campaign.ID != winnerID {
				_ = s.cache.RecordPublisherBidAttempt(pid, candidate.BidPrice, false)
			}
		}
	}(req.PublisherID, matchingCampaigns, winner.Campaign.ID)

	// --- AI: Dynamic Bid Optimization ---
	optimizedBid, err, optHop := s.callOptimizationService(winner, req)
	if err == nil {
		// Track successful optimization hop
		if optHop != nil {
			optHop.Sequence = len(analytics.Hops) + 1
			analytics.Hops = append(analytics.Hops, *optHop)
			analytics.TotalHops++
			analytics.TotalLatencyMs += optHop.LatencyMs
			analytics.TotalFees += optHop.Fee
		}

		winner.BidPrice = optimizedBid.RecommendedBid
		fmt.Printf("Optimized Bid: %.4f (Multiplier: %.2f) Reason: %v\n",
			optimizedBid.RecommendedBid, optimizedBid.BidMultiplier, optimizedBid.Reasoning)
	} else {
		// Track failed optimization hop
		if optHop != nil {
			optHop.Sequence = len(analytics.Hops) + 1
			analytics.Hops = append(analytics.Hops, *optHop)
			analytics.TotalHops++
			analytics.TotalLatencyMs += optHop.LatencyMs
			analytics.TotalFees += optHop.Fee
		}

		metrics.OptimizationErrorsTotal.Inc()
		fmt.Printf("Warning: Optimization Service call failed: %v\n", err)
	}
	// ------------------------------------

	// --- Bid Shading (Second-Price Auctions) ---
	// If auction type is 2nd price (AT=2) and we have multiple bidders,
	// shade the bid down toward second-highest bid to save advertiser money
	if req.AuctionType == 2 && len(matchingCampaigns) > 1 {
		secondHighestBid := matchingCampaigns[1].BidPrice
		shadedBid := s.calculateShadedBid(winner.BidPrice, secondHighestBid, req.AdSlot.BidFloor)
		if shadedBid > 0 {
			winner.BidPrice = shadedBid
		}
	}
	// -------------------------------------------

	// --- Bid Floor Enforcement ---
	// Reject if our final bid price is below the publisher's minimum floor
	if req.AdSlot.BidFloor > 0 && winner.BidPrice < req.AdSlot.BidFloor {
		fmt.Printf("No bid: price %.4f below floor %.4f for slot %s\n",
			winner.BidPrice, req.AdSlot.BidFloor, req.AdSlot.ID)
		metrics.NoBidTotal.WithLabelValues("below_floor").Inc()
		return nil, fmt.Errorf("bid price %.4f is below floor %.4f", winner.BidPrice, req.AdSlot.BidFloor)
	}
	// ------------------------------

	// Record metrics
	metrics.BidsPlacedTotal.WithLabelValues(winner.Campaign.Creative.Type).Inc()
	s.cache.IncrementBidCount()
	s.cache.IncrementWinCount()
	// Track per-campaign win for win-rate calculation
	go func() { _ = s.cache.IncrementCampaignWins(winner.Campaign.ID) }()
	// Track win in price bucket for bid landscape analytics
	winBucket := getPriceBucket(winner.BidPrice)
	go func(b string) { _ = s.cache.RecordWinInBucket(b) }(winBucket)
	// Track publisher bid floor optimization (won auction)
	go func(pid string, price float64) {
		_ = s.cache.RecordPublisherBidAttempt(pid, price, true)
	}(req.PublisherID, winner.BidPrice)

	// Increment frequency cap counter for winner
	if winner.Campaign.Targeting.FreqCapImpressions > 0 && req.User.ID != "" {
		windowSecs := winner.Campaign.Targeting.FreqCapWindowSecs
		if windowSecs <= 0 {
			windowSecs = 86400
		}
		_, _ = s.cache.IncrementUserFrequency(req.User.ID, winner.Campaign.ID, windowSecs)
	}

	// Track per-campaign impression for CTR calculation
	go func() {
		_ = s.cache.IncrementCampaignImpressions(winner.Campaign.ID)
	}()

	// Track segment-level impressions (device, OS, geo)
	go func() {
		_ = s.cache.IncrementSegmentImpressions("device", req.Device.Type)
		_ = s.cache.IncrementSegmentImpressions("os", req.Device.OS)
		_ = s.cache.IncrementSegmentImpressions("geo", req.User.Country)
	}()

	latency := time.Since(startTime).Milliseconds()
	s.cache.RecordLatency(float64(latency))

	// Build response
	response := &model.BidResponse{
		RequestID:   req.ID,
		CampaignID:  winner.Campaign.ID,
		BidPrice:    winner.BidPrice,
		DealID:      winner.Campaign.DealID, // Include Deal ID if present
		CreativeURL: winner.Campaign.Creative.URL,
		ImpressionURL: fmt.Sprintf("%s/api/analytics/track/impression?campaign_id=%s&request_id=%s&price=%.4f",
			s.backendBaseURL, winner.Campaign.ID, req.ID, winner.BidPrice),
		ClickURL: fmt.Sprintf("%s/api/analytics/track/click?campaign_id=%s&request_id=%s",
			s.backendBaseURL, winner.Campaign.ID, req.ID),
		TTL:       300,
		Timestamp: time.Now(),
	}

	// Generate VAST for video
	switch winner.Campaign.Creative.Type {
	case "video", "ctv", "rewarded":
		response.AdMarkup = generateVideoVAST(winner.Campaign, response.ImpressionURL, response.ClickURL)
	case "native":
		// Check context for Native Request payload
		nativeReq := ""
		if val, ok := req.Context["native_request"]; ok {
			if strVal, ok := val.(string); ok {
				nativeReq = strVal
			}
		}
		response.AdMarkup = generateNative(winner.Campaign, response.ImpressionURL, response.ClickURL, nativeReq)
	case "audio":
		response.AdMarkup = generateAudioVAST(winner.Campaign, response.ImpressionURL, response.ClickURL)
	case "rich_media", "interstitial", "ar", "vr", "360_video":
		// AR/VR are treated as Rich Media (HTML5 containers)
		response.AdMarkup = generateRichMedia(winner.Campaign, response.ImpressionURL, response.ClickURL)
	case "playable":
		response.AdMarkup = generatePlayable(winner.Campaign, response.ImpressionURL, response.ClickURL)
	case "pop", "popup", "popunder":
		response.AdMarkup = generatePop(winner.Campaign, response.ImpressionURL, response.ClickURL)
	case "push", "notification":
		response.AdMarkup = generatePush(winner.Campaign, response.ImpressionURL, response.ClickURL)
	case "banner", "display":
		response.AdMarkup = generateBanner(winner.Campaign, response.ImpressionURL, response.ClickURL)
	default:
		// Default to banner format
		response.AdMarkup = generateBanner(winner.Campaign, response.ImpressionURL, response.ClickURL)
	}

	// Finalize SPO analytics
	analytics.FinalBidPrice = winner.BidPrice
	analytics.WonAuction = true
	analytics.CampaignID = winner.Campaign.ID
	analytics.DealID = winner.Campaign.DealID

	// Store analytics in cache
	if err := s.cache.StoreBidPathAnalytics(analytics); err != nil {
		fmt.Printf("Warning: Failed to store SPO analytics: %v\n", err)
	}

	return response, nil
} // generateVideoVAST creates a VAST 3.0/4.0 XML for video/CTV/Rewarded
func generateVideoVAST(c *model.Campaign, impURL, clickURL string) string {
	// For rounded tracking, we need the base tracking URL
	// impURL: .../api/analytics/track/impression?campaign=...
	// videoURL: .../api/analytics/track/video?event={event}&campaign=...
	videoBaseURL := strings.Replace(impURL, "/impression", "/video", 1)

	// Helper to generate tracking event pixels
	const trackingTemplate = `<Tracking event="%s"><![CDATA[%s&event=%s]]></Tracking>`

	// Standard events
	startURL := fmt.Sprintf(trackingTemplate, "start", videoBaseURL, "start")
	firstQuartileURL := fmt.Sprintf(trackingTemplate, "firstQuartile", videoBaseURL, "firstQuartile")
	midpointURL := fmt.Sprintf(trackingTemplate, "midpoint", videoBaseURL, "midpoint")
	thirdQuartileURL := fmt.Sprintf(trackingTemplate, "thirdQuartile", videoBaseURL, "thirdQuartile")
	completeURL := fmt.Sprintf(trackingTemplate, "complete", videoBaseURL, "complete")

	// For rewarded video, we might add a specific extension
	// but standard VAST handles most playback.
	// Optionally add a <Extension type="reward"> block
	extension := ""
	if c.Creative.Rewarded {
		extension = fmt.Sprintf(`
      <Extensions>
        <Extension type="reward">
          <Reward amount="%d" type="%s"/>
        </Extension>
      </Extensions>`, c.Creative.RewardAmt, c.Creative.RewardType)
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<VAST version="4.0">
  <Ad id="%s">
    <InLine>
      <AdSystem>TaskirX</AdSystem>
      <AdTitle>%s</AdTitle>
      <Impression><![CDATA[%s]]></Impression>
      <Creatives>
        <Creative>
          <Linear>
            <Duration>00:00:%02d</Duration>
            <TrackingEvents>
              %s
              %s
              %s
              %s
              %s
            </TrackingEvents>
            <VideoClicks>
              <ClickThrough><![CDATA[%s]]></ClickThrough>
            </VideoClicks>
            <MediaFiles>
              <MediaFile delivery="progressive" type="%s" width="%d" height="%d">
                <![CDATA[%s]]>
              </MediaFile>
            </MediaFiles>
          </Linear>
        </Creative>
      </Creatives>%s
    </InLine>
  </Ad>
</VAST>`,
		c.ID,
		c.Name,
		impURL,
		c.Creative.Duration,
		startURL, firstQuartileURL, midpointURL, thirdQuartileURL, completeURL,
		clickURL,
		c.Creative.MimeType,
		c.Creative.Width,
		c.Creative.Height,
		c.Creative.URL,
		extension)
}

// generateNative creates a Native Ad JSON based on request assets or defaults
func generateNative(c *model.Campaign, impURL, clickURL string, requestRaw string) string {
	var nativeReq model.OpenRTBNativeRequest

	// Try parsing if request is provided
	useDefault := true
	if requestRaw != "" {
		if err := json.Unmarshal([]byte(requestRaw), &nativeReq); err == nil && len(nativeReq.Assets) > 0 {
			useDefault = false
		}
	}

	assets := []map[string]interface{}{}

	if useDefault {
		// Legacy / Default mapping if no request structure found
		assets = []map[string]interface{}{
			{
				"id": 1,
				"title": map[string]string{
					"text": c.Creative.Title,
				},
			},
			{
				"id": 2,
				"img": map[string]interface{}{
					"url": c.Creative.URL, // Main image
					"w":   c.Creative.Width,
					"h":   c.Creative.Height,
				},
			},
			{
				"id": 3,
				"img": map[string]interface{}{
					"url":  c.Creative.IconURL, // Icon
					"type": 1,                  // Icon type
				},
			},
			{
				"id": 4,
				"data": map[string]string{
					"value": c.Creative.Description,
				},
			},
			{
				"id": 5,
				"data": map[string]string{
					"value": c.Creative.CTAText,
				},
			},
		}
	} else {
		// New: Map specific request assets to our creative fields
		for _, assetReq := range nativeReq.Assets {
			assetResp := map[string]interface{}{
				"id": assetReq.ID,
			}
			matched := false

			// Title
			if assetReq.Title != nil {
				assetResp["title"] = map[string]string{"text": c.Creative.Title}
				matched = true
			}

			// Image (Main=3, Icon=1)
			if assetReq.Img != nil {
				switch assetReq.Img.Type {
				case 1: // Icon
					assetResp["img"] = map[string]interface{}{"url": c.Creative.IconURL}
					matched = true
				case 3: // Main
					assetResp["img"] = map[string]interface{}{
						"url": c.Creative.URL,
						"w":   c.Creative.Width,
						"h":   c.Creative.Height,
					}
					matched = true
				default:
					// Fallback to Main Image for unspecifed or other types
					assetResp["img"] = map[string]interface{}{"url": c.Creative.URL}
					matched = true
				}
			}

			// Data (Sponsored=1, Desc=2, Rating=3, Likes=4, Downloads=5, Price=6, SalePrice=7, Phone=8, Address=9, Desc2=10, DispURL=11, CTA=12)
			if assetReq.Data != nil {
				val := ""
				switch assetReq.Data.Type {
				case 1: // Sponsored By
					val = "TaskirX"
				case 2: // Desc
					val = c.Creative.Description
				case 12: // CTA Text
					val = c.Creative.CTAText
				case 11: // Display URL
					val = extractDomain(c.Creative.URL)
				default:
					val = "" // Skip unsupported types
				}

				if val != "" {
					assetResp["data"] = map[string]string{"value": val}
					matched = true
				}
			}

			if matched {
				assets = append(assets, assetResp)
			}
		}
	}

	nativeObj := map[string]interface{}{
		"native": map[string]interface{}{
			"ver": "1.2",
			"link": map[string]string{
				"url": clickURL,
			},
			"imptrackers": []string{impURL},
			"assets":      assets,
		},
	}

	bytes, _ := json.Marshal(nativeObj)
	return string(bytes)
}

func extractDomain(_ string) string {
	// Simple helper for display URL
	// For production, use net/url
	return "taskirx-ad.com"
}

// generateAudioVAST creates a VAST XML for audio ads (DAAST/VAST)
func generateAudioVAST(c *model.Campaign, impURL, clickURL string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<VAST version="4.0">
  <Ad id="%s">
    <InLine>
      <AdSystem>TaskirX Audio</AdSystem>
      <AdTitle>%s</AdTitle>
      <Impression><![CDATA[%s]]></Impression>
      <Creatives>
        <Creative>
          <Linear>
            <Duration>00:00:%02d</Duration>
            <AudioClicks>
              <ClickThrough><![CDATA[%s]]></ClickThrough>
            </AudioClicks>
            <MediaFiles>
              <MediaFile delivery="progressive" type="%s" bitrate="%d">
                <![CDATA[%s]]>
              </MediaFile>
            </MediaFiles>
          </Linear>
        </Creative>
      </Creatives>
    </InLine>
  </Ad>
</VAST>`,
		c.ID,
		c.Name,
		impURL,
		c.Creative.Duration,
		clickURL,
		c.Creative.MimeType,
		c.Creative.Bitrate,
		c.Creative.URL)
}

// generateRichMedia creates HTML5 or snippet based markup
func generateRichMedia(c *model.Campaign, impURL, clickURL string) string {
	// If HTML snippet is provided, use it and inject macros
	if c.Creative.HTMLSnippet != "" {
		// Replace macros
		snippet := c.Creative.HTMLSnippet
		// Simple string replace for macros
		// Note: In production, use a more robust template engine
		// Replace {CLICK_URL}, {IMP_URL}, {CACHEBUSTER}
		return fmt.Sprintf(`<!-- TaskirX Rich Media -->
<div id="ad-%s" style="width:%dpx;height:%dpx;position:relative;">
%s
<img src="%s" style="display:none;width:0;height:0;" />
</div>
<script>
  // Simple click handler injection if needed
  document.getElementById('ad-%s').addEventListener('click', function() {
    window.open('%s', '_blank');
  });
</script>`,
			c.ID, c.Creative.Width, c.Creative.Height, snippet, impURL, c.ID, clickURL)
	}

	// Fallback to iframe if just URL
	return fmt.Sprintf(`<iframe src="%s" width="%d" height="%d" frameborder="0" scrolling="no" marginheight="0" marginwidth="0"></iframe><img src="%s" style="display:none;" />`,
		c.Creative.URL, c.Creative.Width, c.Creative.Height, impURL)
}

// generatePlayable creates HTML markup for playable ads (MRAID usually)
func generatePlayable(c *model.Campaign, impURL, clickURL string) string {
	// Simplified MRAID container
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<script src="mraid.js"></script>
<style>body{margin:0;padding:0;overflow:hidden;}</style>
</head>
<body>
<iframe src="%s" width="100%%" height="100%%" frameborder="0"></iframe>
<script>
// TaskirX Playable Wrapper
var clickUrl = "%s";
var impUrl = "%s";
// Fire impression
(new Image()).src = impUrl;

// Handle MRAID ready
if (typeof mraid !== 'undefined') {
    if (mraid.getState() === 'loading') {
        mraid.addEventListener('ready', function() { mraid.useCustomClose(true); });
    } else {
        mraid.useCustomClose(true);
    }
}
</script>
</body>
</html>`, c.Creative.URL, clickURL, impURL)
}

// generateBanner creates standard banner markup
func generateBanner(c *model.Campaign, impURL, clickURL string) string {
	if c.Creative.HTMLSnippet != "" {
		// Use provided snippet
		return c.Creative.HTMLSnippet
	}
	// Fallback to image tag
	return fmt.Sprintf(`<a href="%s" target="_blank"><img src="%s" width="%d" height="%d" border="0" alt="" /></a><img src="%s" width="1" height="1" style="display:none" />`,
		clickURL, c.Creative.URL, c.Creative.Width, c.Creative.Height, impURL)
}

// generatePop creates popup/popunder JS code
func generatePop(_ *model.Campaign, impURL, clickURL string) string {
	return fmt.Sprintf(`<script>
(function() {
    var url = "%s";
    var imp = "%s";
    var fired = false;

    function deploy() {
        if (fired) return;
        fired = true;

        // Fire impression
        (new Image()).src = imp;
        
        // Popunder logic
        var w = window.open(url, "pop", "width=" + screen.width + ",height=" + screen.height + ",top=0,left=0");
        if(w) {
             w.blur();
             window.focus();
        }
    }

    // Attach to user interaction events to bypass popup blockers
    document.addEventListener('click', deploy);
    document.addEventListener('touchstart', deploy);
})();
</script>`, clickURL, impURL)
}

// generatePush creates a JSON object for push notifications
func generatePush(c *model.Campaign, impURL, clickURL string) string {
	pushObj := map[string]interface{}{
		"title": c.Creative.Title,
		"body":  c.Creative.Description,
		"icon":  c.Creative.IconURL,
		"image": c.Creative.URL,
		"url":   clickURL,
		"imp":   impURL,
	}
	bytes, _ := json.Marshal(pushObj)
	return string(bytes)
}

// calculateScore calculates matching score for a campaign
func (s *BiddingService) calculateScore(campaign *model.Campaign, req *model.BidRequest) float64 {
	score := campaign.BidPrice // Base score is bid price

	// Deal-based price and priority adjustment
	dealResult := s.calculateDealMultiplier(campaign, req)
	if dealResult.UsesDealPrice {
		score = dealResult.DealPrice // Override with deal-specific price
	}
	score *= dealResult.Multiplier // Apply deal priority boost

	// Priority boost: Higher priority campaigns (1-10) get score multiplier
	// Priority 1 = 0.6x, Priority 5 (default) = 1.0x, Priority 10 = 2.0x
	priority := campaign.Priority
	if priority < 1 {
		priority = 5 // Default priority
	} else if priority > 10 {
		priority = 10 // Cap at 10
	}
	priorityMultiplier := 0.6 + (float64(priority-1) * 0.155) // Linear scale from 0.6 to 2.0
	score *= priorityMultiplier

	// Auto-bid optimization: adjust bid price based on performance
	bidMultiplier := s.calculateAutoBidMultiplier(campaign)
	score *= bidMultiplier

	// Viewability-based bid adjustment
	viewabilityMultiplier := s.calculateViewabilityMultiplier(req)
	score *= viewabilityMultiplier

	// Brand Safety Check
	brandSafetyResult := s.checkBrandSafety(campaign, req)
	if brandSafetyResult.Blocked {
		return 0 // Do not bid on unsafe inventory
	}
	score *= brandSafetyResult.Multiplier

	// Contextual Targeting Check
	contextualResult := s.calculateContextualMultiplier(campaign, req)
	if contextualResult.Blocked {
		return 0 // Page contains excluded keywords
	}
	score *= contextualResult.Multiplier

	// Audience Segment Scoring
	audienceResult := s.calculateAudienceSegmentMultiplier(campaign, req)
	if audienceResult.Blocked {
		return 0 // User is in excluded segment or missing required segment
	}
	score *= audienceResult.Multiplier

	// Weather-Based Targeting
	weatherResult := s.calculateWeatherMultiplier(campaign, req)
	if weatherResult.Blocked {
		return 0 // Weather conditions don't match required conditions
	}
	score *= weatherResult.Multiplier

	// POI (Point-of-Interest) Targeting
	poiResult := s.calculatePOIMultiplier(campaign, req)
	if poiResult.Blocked {
		return 0 // User not near required POI or violates distance constraints
	}
	score *= poiResult.Multiplier

	// Carrier/ISP Targeting
	carrierResult := s.calculateCarrierMultiplier(campaign, req)
	if carrierResult.Blocked {
		return 0 // User's carrier/ISP doesn't match requirements
	}
	score *= carrierResult.Multiplier

	// Language Targeting
	languageResult := s.calculateLanguageMultiplier(campaign, req)
	if languageResult.Blocked {
		return 0 // User's language doesn't match requirements
	}
	score *= languageResult.Multiplier

	// Day-of-Week Targeting
	dayOfWeekResult := s.calculateDayOfWeekMultiplier(campaign)
	if !dayOfWeekResult.Allowed {
		return 0 // Campaign not active on this day
	}
	score *= dayOfWeekResult.Multiplier

	// Ad Position Targeting
	adPositionResult := s.calculateAdPositionMultiplier(campaign, req)
	if adPositionResult.Blocked {
		return 0 // Ad position doesn't match requirements
	}
	score *= adPositionResult.Multiplier

	// App Category Targeting
	appTargetingResult := s.calculateAppTargetingMultiplier(campaign, req)
	if appTargetingResult.Blocked {
		return 0 // App doesn't match requirements
	}
	score *= appTargetingResult.Multiplier

	// Seasonal/Event Targeting
	seasonalResult := s.calculateSeasonalMultiplier(campaign)
	score *= seasonalResult.Multiplier

	// Demographic Targeting (age, gender, income)
	demographicResult := s.calculateDemographicMultiplier(campaign, req)
	if demographicResult.Blocked {
		return 0 // User demographics don't match requirements
	}
	score *= demographicResult.Multiplier

	// Video Ad Targeting (player size, placement, completion rates)
	videoResult := s.calculateVideoTargetingMultiplier(campaign, req)
	if videoResult.Blocked {
		return 0 // Video requirements not met
	}
	score *= videoResult.Multiplier

	// Performance Goal Optimization (CPA, CPC, viewability, completion)
	perfGoalResult := s.calculatePerformanceGoalMultiplier(campaign, req)
	if perfGoalResult.Blocked {
		return 0 // Performance thresholds not met
	}
	score *= perfGoalResult.Multiplier

	// Inventory Quality Targeting (brand safety, fraud protection, quality scores)
	invQualityResult := s.calculateInventoryQualityMultiplier(campaign, req)
	if invQualityResult.Blocked {
		return 0 // Inventory quality requirements not met
	}
	score *= invQualityResult.Multiplier

	// Deal/PMP Targeting (programmatic guaranteed, preferred, private auction)
	advDealResult := s.calculateDealTargetingMultiplier(campaign, req)
	if advDealResult.Blocked {
		return 0 // Deal targeting requirements not met
	}
	score *= advDealResult.Multiplier

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

	// Real-Time Budget Check with Time-Based Pacing
	dailySpend, err := s.cache.GetCampaignSpend(campaign.ID)
	if err == nil {
		// Determine daily budget: use explicit DailyBudget or default to Budget/30
		dailyBudget := campaign.DailyBudget
		if dailyBudget <= 0 {
			dailyBudget = campaign.Budget / 30.0
		}

		// Hard cutoff: campaign exceeded daily budget
		if dailySpend >= dailyBudget {
			return 0
		}

		// Calculate pacing multiplier based on strategy
		pacingMultiplier := s.calculatePacingMultiplier(campaign.PacingStrategy, dailySpend, dailyBudget)
		score *= pacingMultiplier
	}

	// Goal-Based Delivery Pacing
	if campaign.GoalTarget > 0 && campaign.GoalEndDate != "" {
		goalMultiplier := s.calculateGoalPacingMultiplier(campaign)
		score *= goalMultiplier
	}

	return score
}

// calculateAutoBidMultiplier adjusts bid price based on real-time campaign performance.
// Logic:
//   - High CTR (>2%) + Low Win Rate (<30%) → Increase bids by 20% (underbidding)
//   - Low CTR (<0.5%) + High Win Rate (>70%) → Decrease bids by 20% (overbidding)
//   - Good performance (CTR 1-3%, Win Rate 40-60%) → Maintain current bid (neutral)
func (s *BiddingService) calculateAutoBidMultiplier(campaign *model.Campaign) float64 {
	// Get real-time performance metrics
	ctr, err := s.cache.GetCampaignCTR(campaign.ID)
	if err != nil {
		return 1.0 // No data, maintain current bid
	}

	winRate, err := s.cache.GetCampaignWinRate(campaign.ID)
	if err != nil {
		return 1.0
	}

	// Convert to percentages for clearer logic
	ctrPct := ctr * 100         // e.g., 0.025 → 2.5%
	winRatePct := winRate * 100 // e.g., 0.25 → 25%

	// Case 1: High engagement but losing auctions → Increase bid
	if ctrPct > 2.0 && winRatePct < 30.0 {
		return 1.20 // +20% bid boost
	}

	// Case 2: Low engagement but winning too many → Decrease bid
	if ctrPct < 0.5 && winRatePct > 70.0 {
		return 0.80 // -20% bid reduction
	}

	// Case 3: Moderate CTR but very low win rate → Moderate increase
	if ctrPct >= 1.0 && ctrPct <= 3.0 && winRatePct < 20.0 {
		return 1.10 // +10% boost
	}

	// Case 4: Low CTR and moderate win rate → Moderate decrease
	if ctrPct < 1.0 && winRatePct >= 50.0 && winRatePct <= 70.0 {
		return 0.90 // -10% reduction
	}

	// Default: Performance is acceptable, maintain current bid
	return 1.0
}

// DealResult represents the result of deal/PMP evaluation
type DealResult struct {
	UsesDealPrice bool    // Whether to override base bid price with deal price
	DealPrice     float64 // The deal-specific price to use
	Multiplier    float64 // Score multiplier based on deal type and priority
	DealType      string  // Type of deal matched
}

// calculateDealMultiplier evaluates PMP/deal-based scoring adjustments
// Deal types and their multipliers:
//   - "guaranteed": 2.5x (must win, highest priority)
//   - "preferred": 1.8x (priority access, preferred pricing)
//   - "private_auction": 1.5x (limited competition)
//   - "open_auction": 1.0x (default, no boost)
func (s *BiddingService) calculateDealMultiplier(campaign *model.Campaign, req *model.BidRequest) DealResult {
	result := DealResult{
		UsesDealPrice: false,
		DealPrice:     0,
		Multiplier:    1.0,
		DealType:      "open_auction",
	}

	// No deal configured for this campaign
	if campaign.DealID == "" {
		return result
	}

	// Verify deal exists in request
	if req.Pmp == nil || len(req.Pmp.Deals) == 0 {
		return result
	}

	// Find matching deal
	var matchedDeal *model.Deal
	for i := range req.Pmp.Deals {
		if req.Pmp.Deals[i].ID == campaign.DealID {
			matchedDeal = &req.Pmp.Deals[i]
			break
		}
	}

	if matchedDeal == nil {
		return result // Deal not found in request
	}

	// Apply deal type multiplier
	dealType := campaign.DealType
	if dealType == "" {
		dealType = "private_auction" // Default for deals
	}
	result.DealType = dealType

	switch dealType {
	case "guaranteed":
		// Programmatic Guaranteed: Must win, highest priority
		result.Multiplier = 2.5
	case "preferred":
		// Preferred Deal: Priority access at fixed price
		result.Multiplier = 1.8
	case "private_auction":
		// Private Auction: Limited competition
		result.Multiplier = 1.5
	default:
		result.Multiplier = 1.0
	}

	// Apply deal-specific priority boost (1-10 scale)
	dealPriority := campaign.DealPriority
	if dealPriority > 0 {
		// Additional boost: 0% to 50% based on deal priority
		result.Multiplier *= (1.0 + float64(dealPriority)*0.05)
	}

	// Check for deal-specific price override
	if campaign.DealPrice > 0 {
		result.UsesDealPrice = true
		result.DealPrice = campaign.DealPrice
	} else if matchedDeal.BidFloor > 0 {
		// Use deal floor as minimum price if no explicit deal price
		if campaign.BidPrice < matchedDeal.BidFloor {
			result.UsesDealPrice = true
			result.DealPrice = matchedDeal.BidFloor
		}
	}

	return result
}

// calculateViewabilityMultiplier adjusts bid price based on expected ad viewability.
// Factors considered:
//   - Ad position (above-fold vs below-fold)
//   - Device type (mobile tends to have lower viewability)
//   - Ad slot size (larger ads tend to be more viewable)
func (s *BiddingService) calculateViewabilityMultiplier(req *model.BidRequest) float64 {
	baseViewability := 0.5 // Default 50% viewability

	// Position factor: Above-fold has ~70% viewability, below-fold ~30%
	switch req.AdSlot.Position {
	case "above-fold", "atf", "top":
		baseViewability = 0.70
	case "below-fold", "btf", "bottom":
		baseViewability = 0.30
	case "sidebar":
		baseViewability = 0.45
	case "sticky", "fixed":
		baseViewability = 0.85 // Sticky ads have highest viewability
	}

	// Device factor: Desktop has higher viewability than mobile
	switch req.Device.Type {
	case "desktop", "pc":
		baseViewability *= 1.1
	case "mobile", "phone":
		baseViewability *= 0.9
	case "tablet":
		baseViewability *= 1.0
	case "ctv", "tv":
		baseViewability *= 1.15 // CTV typically has high viewability
	}

	// Size factor: Larger ads are more viewable
	if len(req.AdSlot.Dimensions) >= 2 {
		area := req.AdSlot.Dimensions[0] * req.AdSlot.Dimensions[1]
		if area >= 300*250 { // Large rectangle or bigger
			baseViewability *= 1.1
		} else if area < 160*600 { // Small banners
			baseViewability *= 0.9
		}
	}

	// Cap viewability at realistic max (95%)
	if baseViewability > 0.95 {
		baseViewability = 0.95
	}

	// Convert viewability to bid multiplier
	// High viewability (>70%) = 1.0-1.3x multiplier (bid more for quality)
	// Low viewability (<40%) = 0.6-0.9x multiplier (bid less for poor quality)
	if baseViewability >= 0.70 {
		return 1.0 + (baseViewability-0.70)*1.0 // Up to 1.25x for 95% viewability
	} else if baseViewability < 0.40 {
		return 0.6 + (baseViewability)*0.75 // 0.6x to 0.9x range
	}

	return 1.0 // Neutral for mid-range viewability (40-70%)
}

// calculateShadedBid implements bid shading for second-price auctions.
// In a 2nd price auction, the winner pays the 2nd highest bid + $0.01.
// Bid shading reduces our bid to save money while still winning.
// Formula: shadedBid = secondBid + (firstBid - secondBid) * 0.1 + small margin
// Ensures we bid slightly above 2nd price but well below our max willingness.
func (s *BiddingService) calculateShadedBid(firstBid, secondBid, bidFloor float64) float64 {
	// Minimum increment over second-highest bid
	minIncrement := 0.01

	// Calculate the gap between first and second bids
	gap := firstBid - secondBid

	// Shade the bid: stay closer to second bid, not first
	// We add 10% of the gap + minimum increment
	shadedBid := secondBid + (gap * 0.1) + minIncrement

	// Ensure shaded bid respects the floor
	if shadedBid < bidFloor {
		shadedBid = bidFloor + minIncrement
	}

	// Ensure shaded bid doesn't exceed original bid
	if shadedBid > firstBid {
		shadedBid = firstBid
	}

	// Ensure we're at least $0.01 above second bid
	if shadedBid < secondBid+minIncrement {
		shadedBid = secondBid + minIncrement
	}

	return shadedBid
}

// BrandSafetyResult contains the result of brand safety evaluation
type BrandSafetyResult struct {
	Blocked    bool    // Whether to completely block this placement
	Multiplier float64 // Score multiplier (1.0 = safe, 0.5-0.9 = risky)
	Reason     string  // Reason for blocking or reduction
}

// checkBrandSafety evaluates brand safety for a campaign + inventory combination
func (s *BiddingService) checkBrandSafety(campaign *model.Campaign, req *model.BidRequest) BrandSafetyResult {
	result := BrandSafetyResult{Blocked: false, Multiplier: 1.0, Reason: ""}

	// Check blocked publishers
	if len(campaign.BlockedPublishers) > 0 {
		for _, blockedPub := range campaign.BlockedPublishers {
			if blockedPub == req.PublisherID {
				return BrandSafetyResult{Blocked: true, Multiplier: 0, Reason: "blocked_publisher"}
			}
		}
	}

	// Check blocked categories (from request context if available)
	if len(campaign.BlockedCategories) > 0 {
		if categories, ok := req.Context["categories"].([]interface{}); ok {
			for _, cat := range categories {
				catStr, _ := cat.(string)
				for _, blockedCat := range campaign.BlockedCategories {
					if catStr == blockedCat {
						return BrandSafetyResult{Blocked: true, Multiplier: 0, Reason: "blocked_category:" + catStr}
					}
				}
			}
		}
	}

	// Check blocked keywords in content
	if len(campaign.BlockedKeywords) > 0 {
		if content, ok := req.Context["content"].(string); ok {
			contentLower := strings.ToLower(content)
			for _, keyword := range campaign.BlockedKeywords {
				if strings.Contains(contentLower, strings.ToLower(keyword)) {
					return BrandSafetyResult{Blocked: true, Multiplier: 0, Reason: "blocked_keyword:" + keyword}
				}
			}
		}
	}

	// Apply brand safety level multipliers
	safetyLevel := campaign.BrandSafetyLevel
	if safetyLevel == "" {
		safetyLevel = "standard"
	}

	// Check for risky categories (reduce bids but don't block unless strict)
	riskyCategories := []string{"IAB25", "IAB26", "IAB7"} // Adult, Illegal, Health (sensitive)
	if categories, ok := req.Context["categories"].([]interface{}); ok {
		for _, cat := range categories {
			catStr, _ := cat.(string)
			for _, risky := range riskyCategories {
				if strings.HasPrefix(catStr, risky) {
					switch safetyLevel {
					case "strict":
						return BrandSafetyResult{Blocked: true, Multiplier: 0, Reason: "strict_risky_category:" + catStr}
					case "standard":
						result.Multiplier *= 0.7 // 30% reduction for risky content
						result.Reason = "risky_category:" + catStr
					case "relaxed":
						result.Multiplier *= 0.9 // 10% reduction
					}
				}
			}
		}
	}

	// Check publisher reputation (from cache if available)
	fraudCount, _ := s.cache.Get(fmt.Sprintf("fraud:publisher:%s:%s:count", req.PublisherID, time.Now().Format("2006-01-02")))
	if fraudCount != "" {
		var count int64
		fmt.Sscanf(fraudCount, "%d", &count)
		if count > 10 {
			if safetyLevel == "strict" {
				return BrandSafetyResult{Blocked: true, Multiplier: 0, Reason: "high_fraud_publisher"}
			}
			result.Multiplier *= 0.5 // Heavy reduction for fraud-flagged publishers
		} else if count > 5 {
			result.Multiplier *= 0.8
		}
	}

	return result
}

// ContextualResult represents the result of contextual targeting evaluation
type ContextualResult struct {
	Blocked         bool     // If true, page contains excluded keywords
	Multiplier      float64  // Bid multiplier based on keyword matches (1.0 = neutral)
	MatchedKeywords []string // Keywords that were matched
	Reason          string   // Explanation for blocking or boost
}

// calculateContextualMultiplier evaluates page content against campaign's contextual targeting
// Returns a multiplier based on keyword matches and category alignment
func (s *BiddingService) calculateContextualMultiplier(campaign *model.Campaign, req *model.BidRequest) ContextualResult {
	result := ContextualResult{
		Blocked:         false,
		Multiplier:      1.0,
		MatchedKeywords: []string{},
	}

	// No contextual targeting configured
	if len(campaign.Targeting.ContextualKeywords) == 0 &&
		len(campaign.Targeting.ContextualCategories) == 0 &&
		len(campaign.Targeting.ContextualExcludeWords) == 0 {
		return result
	}

	// Extract page content from request context
	pageContent := s.extractPageContent(req)
	pageContentLower := strings.ToLower(pageContent)

	// Check excluded keywords first (blocklist)
	for _, excludeWord := range campaign.Targeting.ContextualExcludeWords {
		excludeWordLower := strings.ToLower(excludeWord)
		if strings.Contains(pageContentLower, excludeWordLower) {
			result.Blocked = true
			result.Multiplier = 0
			result.Reason = "excluded_keyword:" + excludeWord
			return result
		}
	}

	// Check contextual categories (from IAB taxonomy in request)
	if len(campaign.Targeting.ContextualCategories) > 0 {
		pageCategories := s.extractPageCategories(req)
		categoryMatches := 0
		for _, targetCat := range campaign.Targeting.ContextualCategories {
			for _, pageCat := range pageCategories {
				if strings.HasPrefix(pageCat, targetCat) || pageCat == targetCat {
					categoryMatches++
					break
				}
			}
		}
		if categoryMatches > 0 {
			result.Multiplier *= (1.0 + float64(categoryMatches)*0.1) // 10% boost per category match
		}
	}

	// Check positive keywords with boost multipliers
	for _, kw := range campaign.Targeting.ContextualKeywords {
		keywordLower := strings.ToLower(kw.Keyword)
		matched := false

		if kw.Exact {
			// Exact word boundary match
			words := strings.Fields(pageContentLower)
			for _, word := range words {
				if word == keywordLower {
					matched = true
					break
				}
			}
		} else {
			// Contains match
			matched = strings.Contains(pageContentLower, keywordLower)
		}

		if matched {
			result.MatchedKeywords = append(result.MatchedKeywords, kw.Keyword)
			boost := kw.Boost
			if boost <= 0 {
				boost = 1.2 // Default 20% boost
			}
			result.Multiplier *= boost
		}
	}

	// Cap the multiplier at 2.0 to prevent runaway bids
	if result.Multiplier > 2.0 {
		result.Multiplier = 2.0
	}

	return result
}

// extractPageContent extracts text content from the bid request for contextual analysis
func (s *BiddingService) extractPageContent(req *model.BidRequest) string {
	var content strings.Builder

	// Extract from context map
	if req.Context != nil {
		if pageTitle, ok := req.Context["page_title"].(string); ok {
			content.WriteString(pageTitle)
			content.WriteString(" ")
		}
		if pageKeywords, ok := req.Context["keywords"].(string); ok {
			content.WriteString(pageKeywords)
			content.WriteString(" ")
		}
		if pageContent, ok := req.Context["content"].(string); ok {
			content.WriteString(pageContent)
			content.WriteString(" ")
		}
		if pageURL, ok := req.Context["page_url"].(string); ok {
			content.WriteString(pageURL)
			content.WriteString(" ")
		}
	}

	return content.String()
}

// extractPageCategories extracts IAB categories from the bid request
func (s *BiddingService) extractPageCategories(req *model.BidRequest) []string {
	categories := []string{}

	// From context map
	if req.Context != nil {
		if cats, ok := req.Context["categories"].([]interface{}); ok {
			for _, cat := range cats {
				if catStr, ok := cat.(string); ok {
					categories = append(categories, catStr)
				}
			}
		}
		if cats, ok := req.Context["iab_categories"].([]interface{}); ok {
			for _, cat := range cats {
				if catStr, ok := cat.(string); ok {
					categories = append(categories, catStr)
				}
			}
		}
	}

	return categories
}

// AudienceSegmentResult represents the result of audience segment evaluation
type AudienceSegmentResult struct {
	Blocked         bool     // User in excluded segment or missing required segment
	Multiplier      float64  // Combined bid multiplier from matching segments
	MatchedSegments []string // Segment IDs that matched
	Reason          string   // Explanation for blocking
}

// calculateAudienceSegmentMultiplier evaluates user segments against campaign targeting
// Returns a multiplier based on segment matches with source-based weighting
func (s *BiddingService) calculateAudienceSegmentMultiplier(campaign *model.Campaign, req *model.BidRequest) AudienceSegmentResult {
	result := AudienceSegmentResult{
		Blocked:         false,
		Multiplier:      1.0,
		MatchedSegments: []string{},
	}

	// No audience targeting configured
	if len(campaign.Targeting.AudienceSegments) == 0 {
		return result
	}

	// Get user's segments from request and cache
	userSegments := s.getUserSegments(req)

	// Track if any required segment was found
	hasRequiredSegment := false
	requiresSegment := false

	for _, targetSegment := range campaign.Targeting.AudienceSegments {
		// Check if user is in this segment
		inSegment := false
		for _, userSeg := range userSegments {
			if userSeg == targetSegment.SegmentID {
				inSegment = true
				break
			}
		}

		// Handle required segments
		if targetSegment.Required {
			requiresSegment = true
			if inSegment {
				hasRequiredSegment = true
			}
		}

		// Handle excluded segments
		if targetSegment.Exclude && inSegment {
			result.Blocked = true
			result.Multiplier = 0
			result.Reason = "excluded_segment:" + targetSegment.SegmentID
			return result
		}

		// Apply weight if matched (and not excluded)
		if inSegment && !targetSegment.Exclude {
			result.MatchedSegments = append(result.MatchedSegments, targetSegment.SegmentID)

			// Get weight with source-based defaults
			weight := targetSegment.Weight
			if weight <= 0 {
				// Default weights by source type
				switch targetSegment.Source {
				case "first_party":
					weight = 1.5 // First-party data is most valuable
				case "lookalike":
					weight = 1.3 // Lookalikes perform well
				case "third_party":
					weight = 1.2 // Third-party data
				case "contextual":
					weight = 1.1 // Contextual signals
				default:
					weight = 1.2 // Default
				}
			}

			// Clamp weight to reasonable range
			if weight < 0.5 {
				weight = 0.5
			} else if weight > 3.0 {
				weight = 3.0
			}

			result.Multiplier *= weight
		}
	}

	// Block if required segment wasn't found
	if requiresSegment && !hasRequiredSegment {
		result.Blocked = true
		result.Multiplier = 0
		result.Reason = "missing_required_segment"
		return result
	}

	// Cap combined multiplier
	if result.Multiplier > 4.0 {
		result.Multiplier = 4.0
	}

	return result
}

// getUserSegments extracts user segments from request and cache
func (s *BiddingService) getUserSegments(req *model.BidRequest) []string {
	segments := []string{}

	// From request context
	if req.Context != nil {
		if segs, ok := req.Context["user_segments"].([]interface{}); ok {
			for _, seg := range segs {
				if segStr, ok := seg.(string); ok {
					segments = append(segments, segStr)
				}
			}
		}
		if segs, ok := req.Context["audience_ids"].([]interface{}); ok {
			for _, seg := range segs {
				if segStr, ok := seg.(string); ok {
					segments = append(segments, segStr)
				}
			}
		}
	}

	// From user categories
	segments = append(segments, req.User.Categories...)

	// From cache (if user ID available)
	if req.User.ID != "" {
		cachedSegments, err := s.cache.GetUserSegments(req.User.ID)
		if err == nil {
			segments = append(segments, cachedSegments...)
		}
	}

	return segments
}

// WeatherResult represents the result of weather targeting evaluation
type WeatherResult struct {
	Blocked           bool     // Required weather conditions not met
	Multiplier        float64  // Bid multiplier based on weather match
	MatchedConditions []string // Conditions that matched
	Reason            string   // Explanation
}

// calculateWeatherMultiplier evaluates weather conditions against campaign targeting
// Weather data is expected in request context from SSP/exchange
func (s *BiddingService) calculateWeatherMultiplier(campaign *model.Campaign, req *model.BidRequest) WeatherResult {
	result := WeatherResult{
		Blocked:           false,
		Multiplier:        1.0,
		MatchedConditions: []string{},
	}

	// No weather targeting configured
	if campaign.Targeting.WeatherTargeting == nil {
		return result
	}

	wt := campaign.Targeting.WeatherTargeting

	// Extract weather data from request context
	weather := s.extractWeatherData(req)
	if weather == nil {
		// No weather data available - don't block, just skip boost
		return result
	}

	// Check temperature range
	if wt.TemperatureMin != nil && weather.Temperature < *wt.TemperatureMin {
		result.Blocked = true
		result.Reason = "temperature_below_min"
		return result
	}
	if wt.TemperatureMax != nil && weather.Temperature > *wt.TemperatureMax {
		result.Blocked = true
		result.Reason = "temperature_above_max"
		return result
	}

	// Check humidity range
	if wt.HumidityMin != nil && weather.Humidity < *wt.HumidityMin {
		result.Blocked = true
		result.Reason = "humidity_below_min"
		return result
	}
	if wt.HumidityMax != nil && weather.Humidity > *wt.HumidityMax {
		result.Blocked = true
		result.Reason = "humidity_above_max"
		return result
	}

	// Check weather conditions
	hasRequiredCondition := false
	requiresCondition := false

	for _, targetCond := range wt.Conditions {
		if targetCond.Required {
			requiresCondition = true
		}

		// Check if current weather matches this condition
		matched := s.matchWeatherCondition(weather, targetCond.Condition)

		if matched {
			if targetCond.Required {
				hasRequiredCondition = true
			}
			result.MatchedConditions = append(result.MatchedConditions, targetCond.Condition)

			// Apply boost
			boost := targetCond.Boost
			if boost <= 0 {
				boost = wt.DefaultBoost
				if boost <= 0 {
					boost = 1.3 // Default 30% boost
				}
			}
			result.Multiplier *= boost
		}
	}

	// Block if required condition not met
	if requiresCondition && !hasRequiredCondition {
		result.Blocked = true
		result.Reason = "missing_required_weather"
		return result
	}

	// Cap multiplier
	if result.Multiplier > 2.5 {
		result.Multiplier = 2.5
	}

	return result
}

// WeatherData represents current weather conditions from request
type WeatherData struct {
	Condition   string  // "sunny", "cloudy", "rainy", etc.
	Temperature float64 // Temperature in Celsius
	Humidity    int     // Humidity percentage
	WindSpeed   float64 // Wind speed in km/h
}

// extractWeatherData extracts weather information from request context
func (s *BiddingService) extractWeatherData(req *model.BidRequest) *WeatherData {
	if req.Context == nil {
		return nil
	}

	weather := &WeatherData{}
	hasData := false

	// Extract weather condition
	if cond, ok := req.Context["weather"].(string); ok {
		weather.Condition = strings.ToLower(cond)
		hasData = true
	}
	if cond, ok := req.Context["weather_condition"].(string); ok {
		weather.Condition = strings.ToLower(cond)
		hasData = true
	}

	// Extract temperature
	if temp, ok := req.Context["temperature"].(float64); ok {
		weather.Temperature = temp
		hasData = true
	}
	if temp, ok := req.Context["temp"].(float64); ok {
		weather.Temperature = temp
		hasData = true
	}

	// Extract humidity
	if hum, ok := req.Context["humidity"].(float64); ok {
		weather.Humidity = int(hum)
		hasData = true
	}

	// Extract wind speed
	if wind, ok := req.Context["wind_speed"].(float64); ok {
		weather.WindSpeed = wind
		hasData = true
	}

	if !hasData {
		return nil
	}

	return weather
}

// matchWeatherCondition checks if current weather matches a target condition
func (s *BiddingService) matchWeatherCondition(weather *WeatherData, targetCondition string) bool {
	condition := strings.ToLower(weather.Condition)
	target := strings.ToLower(targetCondition)

	// Direct match
	if condition == target {
		return true
	}

	// Synonym matching
	synonyms := map[string][]string{
		"sunny":  {"clear", "fair", "bright", "sunshine"},
		"cloudy": {"overcast", "partly_cloudy", "clouds", "gray"},
		"rainy":  {"rain", "showers", "drizzle", "precipitation", "wet"},
		"snowy":  {"snow", "sleet", "blizzard", "flurries", "winter"},
		"stormy": {"storm", "thunderstorm", "thunder", "lightning", "severe"},
		"windy":  {"wind", "breezy", "gusty", "gale"},
		"foggy":  {"fog", "mist", "haze", "hazy"},
		"hot":    {"warm", "heat", "scorching"},
		"cold":   {"freezing", "frigid", "chilly", "frost"},
	}

	// Check if condition matches target or its synonyms
	if syns, ok := synonyms[target]; ok {
		for _, syn := range syns {
			if strings.Contains(condition, syn) {
				return true
			}
		}
		// Also check if condition contains target
		if strings.Contains(condition, target) {
			return true
		}
	}

	// Temperature-based conditions
	if target == "hot" && weather.Temperature > 30 {
		return true
	}
	if target == "cold" && weather.Temperature < 5 {
		return true
	}

	return false
}

// calculatePOIMultiplier evaluates user location against POI targeting configuration
// Returns a POIResult with multiplier and match information
func (s *BiddingService) calculatePOIMultiplier(campaign *model.Campaign, req *model.BidRequest) model.POIResult {
	result := model.POIResult{
		Matched:    false,
		Blocked:    false,
		Multiplier: 1.0,
		Distance:   -1, // Unknown
	}

	// No POI targeting configured
	if campaign.Targeting.POITargeting == nil {
		return result
	}

	poiConfig := campaign.Targeting.POITargeting

	// Extract user location from request
	userLat, userLon, hasLocation := s.extractUserLocation(req)
	if !hasLocation {
		// No location data - if POIs have required flag, block
		for _, poi := range poiConfig.POIs {
			if poi.Required {
				result.Blocked = true
				result.Reason = "location_required_missing"
				return result
			}
		}
		return result // No location, no targeting - neutral
	}

	var nearestPOI *model.POI
	var nearestDistance float64 = -1
	matchedPOIs := []string{}

	// Check against configured POIs
	for i, poi := range poiConfig.POIs {
		distance := haversineDistance(userLat, userLon, poi.Lat, poi.Lon)

		// Track nearest POI
		if nearestDistance < 0 || distance < nearestDistance {
			nearestDistance = distance
			nearestPOI = &poiConfig.POIs[i]
		}

		// Check if within radius
		radius := poi.Radius
		if radius <= 0 {
			radius = 1.0 // Default 1km
		}

		if distance <= radius {
			result.Matched = true
			matchedPOIs = append(matchedPOIs, poi.Name)

			// Apply POI-specific boost
			boost := poi.Boost
			if boost <= 0 {
				boost = 1.3 // Default 30% boost
			}
			result.Multiplier *= boost
		} else if poi.Required {
			// Required POI but user not within radius
			result.Blocked = true
			result.Reason = "not_near_required_poi:" + poi.Name
			return result
		}
	}

	// Apply distance-based boosts
	if nearestDistance >= 0 && len(poiConfig.DistanceBoosts) > 0 {
		for _, db := range poiConfig.DistanceBoosts {
			if nearestDistance <= db.MaxDistance {
				result.Multiplier *= db.Boost
				break // Apply only the first matching distance bracket
			}
		}
	}

	// Check min/max distance constraints
	if nearestDistance >= 0 {
		if poiConfig.MinDistance > 0 && nearestDistance < poiConfig.MinDistance {
			result.Blocked = true
			result.Reason = "too_close_to_poi"
			return result
		}
		if poiConfig.MaxDistance > 0 && nearestDistance > poiConfig.MaxDistance {
			result.Blocked = true
			result.Reason = "too_far_from_poi"
			return result
		}
	}

	result.NearestPOI = nearestPOI
	result.Distance = nearestDistance
	result.MatchedPOIs = matchedPOIs

	// Cap multiplier
	if result.Multiplier > 3.0 {
		result.Multiplier = 3.0
	}

	return result
}

// extractUserLocation extracts lat/lon from bid request
func (s *BiddingService) extractUserLocation(req *model.BidRequest) (lat, lon float64, hasLocation bool) {
	// Check device geo first (legacy format)
	if req.Device.Geo.Lat != 0 || req.Device.Geo.Lon != 0 {
		return req.Device.Geo.Lat, req.Device.Geo.Lon, true
	}

	// Check context for lat/lon
	if req.Context != nil {
		if lat, ok := req.Context["lat"].(float64); ok {
			if lon, ok := req.Context["lon"].(float64); ok {
				return lat, lon, true
			}
		}
		if lat, ok := req.Context["latitude"].(float64); ok {
			if lon, ok := req.Context["longitude"].(float64); ok {
				return lat, lon, true
			}
		}
	}

	return 0, 0, false
}

// haversineDistance calculates the distance in km between two lat/lon coordinates
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0

	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	lat1Rad := degreesToRadians(lat1)
	lat2Rad := degreesToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// degreesToRadians converts degrees to radians
func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

// calculateCarrierMultiplier evaluates user's carrier/ISP against targeting configuration
func (s *BiddingService) calculateCarrierMultiplier(campaign *model.Campaign, req *model.BidRequest) model.CarrierResult {
	result := model.CarrierResult{
		Matched:    false,
		Blocked:    false,
		Multiplier: 1.0,
	}

	// No carrier targeting configured
	if campaign.Targeting.CarrierTargeting == nil {
		return result
	}

	ct := campaign.Targeting.CarrierTargeting

	// Extract carrier/ISP info from request
	carrier, isp, connType := s.extractNetworkInfo(req)
	result.Carrier = carrier
	result.ISP = isp
	result.ConnectionType = connType

	// Check connection type requirements
	if ct.CellularOnly && connType != "cellular" {
		result.Blocked = true
		result.Reason = "cellular_only"
		return result
	}
	if ct.WiFiOnly && connType != "wifi" {
		result.Blocked = true
		result.Reason = "wifi_only"
		return result
	}

	// Check allowed connection types
	if len(ct.ConnectionTypes) > 0 {
		allowed := false
		for _, allowedType := range ct.ConnectionTypes {
			if strings.EqualFold(connType, allowedType) {
				allowed = true
				break
			}
		}
		if !allowed {
			result.Blocked = true
			result.Reason = "connection_type_not_allowed"
			return result
		}
	}

	// Check excluded carriers
	for _, excluded := range ct.ExcludeCarriers {
		if strings.EqualFold(carrier, excluded) {
			result.Blocked = true
			result.Reason = "carrier_excluded:" + excluded
			return result
		}
	}

	// Check excluded ISPs
	for _, excluded := range ct.ExcludeISPs {
		if strings.EqualFold(isp, excluded) {
			result.Blocked = true
			result.Reason = "isp_excluded:" + excluded
			return result
		}
	}

	// Check carrier rules
	hasRequiredCarrier := false
	requiresCarrier := false
	for _, rule := range ct.Carriers {
		if rule.Required {
			requiresCarrier = true
		}

		matched := s.matchCarrier(carrier, rule)
		if matched {
			result.Matched = true
			if rule.Required {
				hasRequiredCarrier = true
			}

			boost := rule.Boost
			if boost <= 0 {
				boost = 1.2 // Default 20% boost
			}
			result.Multiplier *= boost
		}
	}

	// Block if required carrier not matched
	if requiresCarrier && !hasRequiredCarrier {
		result.Blocked = true
		result.Reason = "missing_required_carrier"
		return result
	}

	// Check ISP rules (for WiFi/desktop users)
	hasRequiredISP := false
	requiresISP := false
	for _, rule := range ct.ISPs {
		if rule.Required {
			requiresISP = true
		}

		if strings.EqualFold(isp, rule.Name) {
			result.Matched = true
			if rule.Required {
				hasRequiredISP = true
			}

			boost := rule.Boost
			if boost <= 0 {
				boost = 1.2
			}
			result.Multiplier *= boost
		}
	}

	// Block if required ISP not matched
	if requiresISP && !hasRequiredISP {
		result.Blocked = true
		result.Reason = "missing_required_isp"
		return result
	}

	// Cap multiplier
	if result.Multiplier > 2.5 {
		result.Multiplier = 2.5
	}

	return result
}

// extractNetworkInfo extracts carrier, ISP, and connection type from request
func (s *BiddingService) extractNetworkInfo(req *model.BidRequest) (carrier, isp, connType string) {
	// Default connection type
	connType = "unknown"

	// Check context for network info
	if req.Context != nil {
		if c, ok := req.Context["carrier"].(string); ok {
			carrier = c
		}
		if i, ok := req.Context["isp"].(string); ok {
			isp = i
		}
		if ct, ok := req.Context["connection_type"].(string); ok {
			connType = ct
		}
		if ct, ok := req.Context["connectiontype"].(string); ok {
			connType = ct
		}
	}

	// Map numeric connection type (OpenRTB standard)
	// 0=Unknown, 1=Ethernet, 2=WiFi, 3=Cellular Unknown, 4=2G, 5=3G, 6=4G, 7=5G
	if connTypeNum, ok := req.Context["connectiontype"].(float64); ok {
		switch int(connTypeNum) {
		case 1:
			connType = "ethernet"
		case 2:
			connType = "wifi"
		case 3, 4, 5, 6, 7:
			connType = "cellular"
		}
	}

	return carrier, isp, connType
}

// matchCarrier checks if user's carrier matches a carrier rule
func (s *BiddingService) matchCarrier(userCarrier string, rule model.CarrierRule) bool {
	if userCarrier == "" {
		return false
	}

	// Match by name (case-insensitive, partial match allowed)
	carrierLower := strings.ToLower(userCarrier)
	ruleLower := strings.ToLower(rule.Name)

	if strings.Contains(carrierLower, ruleLower) || strings.Contains(ruleLower, carrierLower) {
		return true
	}

	// Common carrier name variations
	carrierAliases := map[string][]string{
		"verizon":  {"verizon", "vzw", "verizon wireless"},
		"att":      {"att", "at&t", "at and t", "cingular"},
		"t-mobile": {"t-mobile", "tmobile", "t mobile", "metro"},
		"sprint":   {"sprint", "boost"},
		"comcast":  {"comcast", "xfinity"},
		"spectrum": {"spectrum", "charter", "time warner"},
		"vodafone": {"vodafone", "voda"},
		"orange":   {"orange"},
		"ee":       {"ee", "everything everywhere"},
		"three":    {"three", "3"},
	}

	for canonical, aliases := range carrierAliases {
		if strings.Contains(ruleLower, canonical) {
			for _, alias := range aliases {
				if strings.Contains(carrierLower, alias) {
					return true
				}
			}
		}
	}

	return false
}

// calculateLanguageMultiplier evaluates user's language against targeting configuration
func (s *BiddingService) calculateLanguageMultiplier(campaign *model.Campaign, req *model.BidRequest) model.LanguageResult {
	result := model.LanguageResult{
		Matched:    false,
		Blocked:    false,
		Multiplier: 1.0,
	}

	// No language targeting configured
	if campaign.Targeting.LanguageTargeting == nil {
		return result
	}

	lt := campaign.Targeting.LanguageTargeting

	// Extract language info from request
	userLang, contentLang := s.extractLanguageInfo(req)
	result.UserLanguage = userLang
	result.ContentLanguage = contentLang

	// Check excluded languages
	for _, excluded := range lt.ExcludeLanguages {
		if s.matchLanguageCode(userLang, excluded, false) {
			result.Blocked = true
			result.Reason = "language_excluded:" + excluded
			return result
		}
		if lt.ContentLanguage && s.matchLanguageCode(contentLang, excluded, false) {
			result.Blocked = true
			result.Reason = "content_language_excluded:" + excluded
			return result
		}
	}

	// Check language rules
	hasRequiredLanguage := false
	requiresLanguage := false

	for _, rule := range lt.Languages {
		if rule.Required {
			requiresLanguage = true
		}

		// Determine if we should match locale precisely
		matchLocale := lt.LocaleMatching && rule.Locale != ""

		// Check user language
		matched := false
		if matchLocale {
			matched = s.matchLanguageCode(userLang, rule.Locale, true)
		} else {
			matched = s.matchLanguageCode(userLang, rule.Code, false)
		}

		// Also check content language if configured
		if !matched && lt.ContentLanguage && !lt.PrimaryOnly {
			if matchLocale {
				matched = s.matchLanguageCode(contentLang, rule.Locale, true)
			} else {
				matched = s.matchLanguageCode(contentLang, rule.Code, false)
			}
		}

		if matched {
			result.Matched = true
			result.MatchedCode = rule.Code
			if rule.Required {
				hasRequiredLanguage = true
			}

			boost := rule.Boost
			if boost <= 0 {
				boost = 1.2 // Default 20% boost
			}
			result.Multiplier *= boost
		}
	}

	// Block if required language not matched
	if requiresLanguage && !hasRequiredLanguage {
		result.Blocked = true
		result.Reason = "missing_required_language"
		return result
	}

	// Apply default multiplier if no match
	if !result.Matched && lt.DefaultMultiplier > 0 {
		result.Multiplier = lt.DefaultMultiplier
	}

	// Cap multiplier
	if result.Multiplier > 2.5 {
		result.Multiplier = 2.5
	}

	return result
}

// extractLanguageInfo extracts user and content language from request
func (s *BiddingService) extractLanguageInfo(req *model.BidRequest) (userLang, contentLang string) {
	// Get user language from User struct
	userLang = req.User.Language

	// Check context for additional language info
	if req.Context != nil {
		if lang, ok := req.Context["language"].(string); ok && userLang == "" {
			userLang = lang
		}
		if lang, ok := req.Context["content_language"].(string); ok {
			contentLang = lang
		}
		if lang, ok := req.Context["page_language"].(string); ok && contentLang == "" {
			contentLang = lang
		}
	}

	return userLang, contentLang
}

// matchLanguageCode matches a language code with support for locale matching
// exactLocale: true for "en-US" == "en-US", false for "en-US" matches "en"
func (s *BiddingService) matchLanguageCode(userLang, targetLang string, exactLocale bool) bool {
	if userLang == "" || targetLang == "" {
		return false
	}

	userLang = strings.ToLower(strings.TrimSpace(userLang))
	targetLang = strings.ToLower(strings.TrimSpace(targetLang))

	// Exact match
	if userLang == targetLang {
		return true
	}

	// If exact locale required, no further matching
	if exactLocale {
		return false
	}

	// Extract primary language code (e.g., "en" from "en-US")
	userPrimary := strings.Split(userLang, "-")[0]
	targetPrimary := strings.Split(targetLang, "-")[0]

	// Also handle underscore format (en_US)
	if !strings.Contains(userLang, "-") {
		userPrimary = strings.Split(userLang, "_")[0]
	}
	if !strings.Contains(targetLang, "-") {
		targetPrimary = strings.Split(targetLang, "_")[0]
	}

	return userPrimary == targetPrimary
}

// calculateDayOfWeekMultiplier evaluates current day against targeting configuration
func (s *BiddingService) calculateDayOfWeekMultiplier(campaign *model.Campaign) model.DayOfWeekResult {
	result := model.DayOfWeekResult{
		Allowed:    true,
		Multiplier: 1.0,
	}

	// No day-of-week targeting configured
	if campaign.Targeting.DayOfWeekTargeting == nil {
		return result
	}

	dt := campaign.Targeting.DayOfWeekTargeting

	// Get current time in configured timezone
	now := time.Now()
	if dt.Timezone != "" {
		if loc, err := time.LoadLocation(dt.Timezone); err == nil {
			now = now.In(loc)
		}
	}

	dayNum := int(now.Weekday()) // 0=Sunday, 1=Monday, ..., 6=Saturday
	result.DayNumber = dayNum
	result.DayName = now.Weekday().String()
	result.IsWeekend = dayNum == 0 || dayNum == 6

	// Check weekend/weekday restrictions
	if dt.WeekdaysOnly && result.IsWeekend {
		result.Allowed = false
		result.Reason = "weekdays_only"
		return result
	}
	if dt.WeekendsOnly && !result.IsWeekend {
		result.Allowed = false
		result.Reason = "weekends_only"
		return result
	}

	// Check specific day configuration
	for _, dayConfig := range dt.Days {
		if dayConfig.Day == dayNum {
			// Found configuration for this day
			if !dayConfig.Active {
				result.Allowed = false
				result.Reason = "day_not_active"
				return result
			}

			// Check day-specific hours if configured
			if len(dayConfig.Hours) > 0 {
				currentHour := now.Hour()
				hourAllowed := false
				for _, allowedHour := range dayConfig.Hours {
					if currentHour == allowedHour {
						hourAllowed = true
						break
					}
				}
				if !hourAllowed {
					result.Allowed = false
					result.Reason = "hour_not_active_for_day"
					return result
				}
			}

			// Apply day-specific boost
			if dayConfig.Boost > 0 {
				result.Multiplier = dayConfig.Boost
			}

			return result
		}
	}

	// No specific config for this day - use default boost if set
	if dt.DefaultBoost > 0 {
		result.Multiplier = dt.DefaultBoost
	}

	return result
}

// calculateAdPositionMultiplier evaluates ad position against targeting configuration
func (s *BiddingService) calculateAdPositionMultiplier(campaign *model.Campaign, req *model.BidRequest) model.AdPositionResult {
	result := model.AdPositionResult{
		Matched:    false,
		Blocked:    false,
		Multiplier: 1.0,
	}

	// No ad position targeting configured
	if campaign.Targeting.AdPositionTargeting == nil {
		return result
	}

	apt := campaign.Targeting.AdPositionTargeting

	// Extract ad position info from request
	position, isAboveFold, viewability := s.extractAdPositionInfo(req)
	result.DetectedPosition = position
	result.IsAboveFold = isAboveFold
	result.PredictedViewability = viewability

	// Check minimum viewability requirement
	if apt.MinViewability > 0 && viewability > 0 && viewability < apt.MinViewability {
		result.Blocked = true
		result.Reason = "viewability_below_minimum"
		return result
	}

	// Check above-fold only requirement
	if apt.AboveFoldOnly && !isAboveFold {
		result.Blocked = true
		result.Reason = "above_fold_only"
		return result
	}

	// Check excluded positions
	for _, excluded := range apt.ExcludePositions {
		if strings.EqualFold(position, excluded) {
			result.Blocked = true
			result.Reason = "position_excluded:" + excluded
			return result
		}
	}

	// Apply position-based boosts
	if isAboveFold && apt.AboveFoldBoost > 0 {
		result.Multiplier *= apt.AboveFoldBoost
		result.Matched = true
	} else if !isAboveFold && apt.BelowFoldDiscount > 0 {
		result.Multiplier *= apt.BelowFoldDiscount
	}

	// Apply interstitial boost
	if s.isInterstitialPosition(position) && apt.InterstitialBoost > 0 {
		result.Multiplier *= apt.InterstitialBoost
		result.Matched = true
	}

	// Apply sticky boost
	if s.isStickyPosition(position) && apt.StickyBoost > 0 {
		result.Multiplier *= apt.StickyBoost
		result.Matched = true
	}

	// Check position rules
	hasRequiredPosition := false
	requiresPosition := false

	for _, rule := range apt.Positions {
		if rule.Required {
			requiresPosition = true
		}

		if strings.EqualFold(position, rule.Position) {
			result.Matched = true
			if rule.Required {
				hasRequiredPosition = true
			}

			boost := rule.Boost
			if boost <= 0 {
				boost = 1.2
			}
			result.Multiplier *= boost
		}
	}

	// Block if required position not matched
	if requiresPosition && !hasRequiredPosition {
		result.Blocked = true
		result.Reason = "missing_required_position"
		return result
	}

	// Cap multiplier
	if result.Multiplier > 3.0 {
		result.Multiplier = 3.0
	}
	if result.Multiplier < 0.3 {
		result.Multiplier = 0.3
	}

	return result
}

// extractAdPositionInfo extracts ad position details from request
func (s *BiddingService) extractAdPositionInfo(req *model.BidRequest) (position string, isAboveFold bool, viewability float64) {
	position = "unknown"
	isAboveFold = false
	viewability = 0.0

	// Check context for position info
	if req.Context != nil {
		if pos, ok := req.Context["ad_position"].(string); ok {
			position = pos
		}
		if pos, ok := req.Context["position"].(string); ok && position == "unknown" {
			position = pos
		}
		if atf, ok := req.Context["above_fold"].(bool); ok {
			isAboveFold = atf
		}
		if atf, ok := req.Context["above_the_fold"].(bool); ok {
			isAboveFold = atf
		}
		if view, ok := req.Context["predicted_viewability"].(float64); ok {
			viewability = view
		}
		if view, ok := req.Context["viewability"].(float64); ok && viewability == 0 {
			viewability = view
		}
	}

	// Get position from AdSlot (legacy format)
	if req.AdSlot.Position != "" && position == "unknown" {
		position = req.AdSlot.Position
	}

	// Map position string to above-fold status
	posLower := strings.ToLower(position)
	switch {
	case strings.Contains(posLower, "above"):
		isAboveFold = true
	case strings.Contains(posLower, "below"):
		isAboveFold = false
	case strings.Contains(posLower, "header"):
		isAboveFold = true
	case strings.Contains(posLower, "footer"):
		isAboveFold = false
	case strings.Contains(posLower, "interstitial"):
		isAboveFold = true
	}

	// Infer above-fold from position name
	if position != "unknown" && !isAboveFold {
		aboveFoldPositions := []string{"above_fold", "header", "top", "interstitial", "sticky", "fixed"}
		for _, atfPos := range aboveFoldPositions {
			if strings.Contains(strings.ToLower(position), atfPos) {
				isAboveFold = true
				break
			}
		}
	}

	return position, isAboveFold, viewability
}

// isInterstitialPosition checks if position is an interstitial/fullscreen type
func (s *BiddingService) isInterstitialPosition(position string) bool {
	interstitialTypes := []string{"interstitial", "fullscreen", "full_screen", "overlay", "modal"}
	posLower := strings.ToLower(position)
	for _, t := range interstitialTypes {
		if strings.Contains(posLower, t) {
			return true
		}
	}
	return false
}

// isStickyPosition checks if position is a sticky/fixed type
func (s *BiddingService) isStickyPosition(position string) bool {
	stickyTypes := []string{"sticky", "fixed", "anchor", "floating"}
	posLower := strings.ToLower(position)
	for _, t := range stickyTypes {
		if strings.Contains(posLower, t) {
			return true
		}
	}
	return false
}

// calculateAppTargetingMultiplier evaluates app info against targeting configuration
func (s *BiddingService) calculateAppTargetingMultiplier(campaign *model.Campaign, req *model.BidRequest) model.AppTargetingResult {
	result := model.AppTargetingResult{
		Matched:    false,
		Blocked:    false,
		Multiplier: 1.0,
	}

	// No app targeting configured
	if campaign.Targeting.AppTargeting == nil {
		return result
	}

	at := campaign.Targeting.AppTargeting

	// Extract app info from request
	bundleID, appName, category, isInApp, appRating := s.extractAppInfo(req)
	result.BundleID = bundleID
	result.AppName = appName
	result.Category = category
	result.IsInApp = isInApp
	result.AppRating = appRating

	// Check in-app only requirement
	if at.InAppOnly && !isInApp {
		result.Blocked = true
		result.Reason = "in_app_only"
		return result
	}

	// Check mobile web only requirement
	if at.MobileWebOnly && isInApp {
		result.Blocked = true
		result.Reason = "mobile_web_only"
		return result
	}

	// Check minimum app rating
	if at.MinAppRating > 0 && appRating > 0 && appRating < at.MinAppRating {
		result.Blocked = true
		result.Reason = "app_rating_below_minimum"
		return result
	}

	// Check excluded bundle IDs
	for _, excluded := range at.ExcludeBundleIDs {
		if strings.EqualFold(bundleID, excluded) {
			result.Blocked = true
			result.Reason = "bundle_id_excluded:" + excluded
			return result
		}
	}

	// Check excluded categories
	for _, excluded := range at.ExcludeCategories {
		if strings.EqualFold(category, excluded) {
			result.Blocked = true
			result.Reason = "category_excluded:" + excluded
			return result
		}
	}

	// Check bundle ID rules
	hasRequiredBundle := false
	requiresBundle := false

	for _, rule := range at.BundleIDs {
		if rule.Required {
			requiresBundle = true
		}

		if s.matchBundleID(bundleID, rule.Value) {
			result.Matched = true
			if rule.Required {
				hasRequiredBundle = true
			}

			boost := rule.Boost
			if boost <= 0 {
				boost = 1.2
			}
			result.Multiplier *= boost
		}
	}

	// Block if required bundle not matched
	if requiresBundle && !hasRequiredBundle {
		result.Blocked = true
		result.Reason = "missing_required_bundle_id"
		return result
	}

	// Check category rules
	hasRequiredCategory := false
	requiresCategory := false

	for _, rule := range at.Categories {
		if rule.Required {
			requiresCategory = true
		}

		if strings.EqualFold(category, rule.Value) || s.matchAppCategory(category, rule.Value) {
			result.Matched = true
			if rule.Required {
				hasRequiredCategory = true
			}

			boost := rule.Boost
			if boost <= 0 {
				boost = 1.2
			}
			result.Multiplier *= boost
		}
	}

	// Block if required category not matched
	if requiresCategory && !hasRequiredCategory {
		result.Blocked = true
		result.Reason = "missing_required_category"
		return result
	}

	// Apply premium apps boost
	if at.PremiumAppsBoost > 0 && s.isPremiumApp(bundleID, appRating) {
		result.IsPremiumApp = true
		result.Multiplier *= at.PremiumAppsBoost
	}

	// Cap multiplier
	if result.Multiplier > 3.0 {
		result.Multiplier = 3.0
	}

	return result
}

// extractAppInfo extracts app details from request
func (s *BiddingService) extractAppInfo(req *model.BidRequest) (bundleID, appName, category string, isInApp bool, appRating float64) {
	// Check context for app info
	if req.Context != nil {
		if bid, ok := req.Context["bundle_id"].(string); ok {
			bundleID = bid
		}
		if bid, ok := req.Context["bundle"].(string); ok && bundleID == "" {
			bundleID = bid
		}
		if name, ok := req.Context["app_name"].(string); ok {
			appName = name
		}
		if cat, ok := req.Context["app_category"].(string); ok {
			category = cat
		}
		if cat, ok := req.Context["category"].(string); ok && category == "" {
			category = cat
		}
		if inApp, ok := req.Context["is_app"].(bool); ok {
			isInApp = inApp
		}
		if inApp, ok := req.Context["in_app"].(bool); ok {
			isInApp = inApp
		}
		if rating, ok := req.Context["app_rating"].(float64); ok {
			appRating = rating
		}
	}

	// Infer in-app from bundle ID format
	if bundleID != "" && !isInApp {
		// Bundle IDs typically look like: com.company.app, id123456789
		if strings.Contains(bundleID, ".") || strings.HasPrefix(bundleID, "id") {
			isInApp = true
		}
	}

	return bundleID, appName, category, isInApp, appRating
}

// matchBundleID checks if user's bundle ID matches a rule (supports wildcards)
func (s *BiddingService) matchBundleID(userBundle, ruleBundle string) bool {
	if userBundle == "" || ruleBundle == "" {
		return false
	}

	userBundle = strings.ToLower(userBundle)
	ruleBundle = strings.ToLower(ruleBundle)

	// Exact match
	if userBundle == ruleBundle {
		return true
	}

	// Wildcard matching (com.company.* matches com.company.app1, com.company.app2)
	if strings.HasSuffix(ruleBundle, "*") {
		prefix := strings.TrimSuffix(ruleBundle, "*")
		return strings.HasPrefix(userBundle, prefix)
	}

	return false
}

// matchAppCategory checks if app category matches rule (handles variations)
func (s *BiddingService) matchAppCategory(appCategory, ruleCategory string) bool {
	if appCategory == "" || ruleCategory == "" {
		return false
	}

	appCat := strings.ToLower(appCategory)
	ruleCat := strings.ToLower(ruleCategory)

	// Category aliases mapping
	categoryAliases := map[string][]string{
		"games":         {"games", "gaming", "game"},
		"social":        {"social", "social networking", "social media"},
		"news":          {"news", "news & magazines", "news_and_magazines"},
		"entertainment": {"entertainment", "media", "video"},
		"shopping":      {"shopping", "retail", "ecommerce"},
		"finance":       {"finance", "banking", "fintech"},
		"health":        {"health", "health & fitness", "medical", "healthcare"},
		"education":     {"education", "learning", "educational"},
		"travel":        {"travel", "travel & local", "navigation"},
		"music":         {"music", "music & audio", "audio"},
		"sports":        {"sports", "sports news"},
		"lifestyle":     {"lifestyle", "food & drink", "food"},
		"productivity":  {"productivity", "business", "tools"},
	}

	for canonical, aliases := range categoryAliases {
		if strings.Contains(ruleCat, canonical) || ruleCat == canonical {
			for _, alias := range aliases {
				if strings.Contains(appCat, alias) {
					return true
				}
			}
		}
	}

	return false
}

// isPremiumApp checks if app is considered premium based on bundle ID and rating
func (s *BiddingService) isPremiumApp(bundleID string, rating float64) bool {
	// High rating indicates premium
	if rating >= 4.5 {
		return true
	}

	// Known premium app publishers
	premiumPublishers := []string{
		"com.spotify",
		"com.netflix",
		"com.amazon",
		"com.google",
		"com.facebook",
		"com.instagram",
		"com.twitter",
		"com.snapchat",
		"com.tiktok",
		"com.uber",
		"com.lyft",
		"com.airbnb",
		"com.linkedin",
		"com.pinterest",
		"com.reddit",
		"com.discord",
	}

	bundleLower := strings.ToLower(bundleID)
	for _, publisher := range premiumPublishers {
		if strings.HasPrefix(bundleLower, publisher) {
			return true
		}
	}

	return false
}

// calculateSeasonalMultiplier evaluates current date against seasonal targeting configuration
func (s *BiddingService) calculateSeasonalMultiplier(campaign *model.Campaign) model.SeasonalResult {
	result := model.SeasonalResult{
		Matched:    false,
		Multiplier: 1.0,
	}

	// No seasonal targeting configured
	if campaign.Targeting.SeasonalTargeting == nil {
		return result
	}

	st := campaign.Targeting.SeasonalTargeting

	// Get current time in configured timezone
	now := time.Now()
	if st.Timezone != "" {
		if loc, err := time.LoadLocation(st.Timezone); err == nil {
			now = now.In(loc)
		}
	}

	// Check basic time-based boosts
	dayOfWeek := int(now.Weekday())
	month := int(now.Month())
	day := now.Day()
	daysInMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()

	// Weekend boost
	result.IsWeekend = dayOfWeek == 0 || dayOfWeek == 6
	if result.IsWeekend && st.WeekendBoost > 0 {
		result.Multiplier *= st.WeekendBoost
		result.Matched = true
	}

	// Month end boost (last 5 days of month - payday period)
	result.IsMonthEnd = day > daysInMonth-5
	if result.IsMonthEnd && st.MonthEndBoost > 0 {
		result.Multiplier *= st.MonthEndBoost
		result.Matched = true
	}

	// Q4 boost (October - December)
	result.IsQ4 = month >= 10 && month <= 12
	if result.IsQ4 && st.Q4Boost > 0 {
		result.Multiplier *= st.Q4Boost
		result.Matched = true
	}

	// Summer boost (June - August)
	if month >= 6 && month <= 8 && st.SummerBoost > 0 {
		result.Season = "summer"
		result.Multiplier *= st.SummerBoost
		result.Matched = true
	}

	// Back to school boost (August - September)
	if (month == 8 || month == 9) && st.BackToSchoolBoost > 0 {
		result.Multiplier *= st.BackToSchoolBoost
		result.Matched = true
	}

	// Check custom events
	for _, event := range st.Events {
		if !event.Active {
			continue
		}

		if s.isEventActive(event, now) {
			result.Matched = true
			result.ActiveEvents = append(result.ActiveEvents, event.Name)

			boost := event.Boost
			if boost <= 0 {
				boost = 1.5 // Default 50% boost for events
			}
			result.Multiplier *= boost
		}
	}

	// Check holidays (if enabled)
	if st.EnableHolidays {
		country := st.Country
		if country == "" {
			country = "US" // Default to US holidays
		}

		holidayName := s.getHolidayName(now, country)
		if holidayName != "" {
			result.IsHoliday = true
			result.HolidayName = holidayName
			result.Matched = true

			boost := st.HolidayBoost
			if boost <= 0 {
				boost = 1.3 // Default 30% boost for holidays
			}
			result.Multiplier *= boost
		}
	}

	// Cap multiplier
	if result.Multiplier > 3.0 {
		result.Multiplier = 3.0
	}

	return result
}

// isEventActive checks if a seasonal event is currently active
func (s *BiddingService) isEventActive(event model.SeasonalEvent, now time.Time) bool {
	var startDate, endDate time.Time
	var err error

	// Parse dates - support both full dates and recurring MM-DD format
	if event.Recurring || len(event.StartDate) == 5 { // MM-DD format
		startDate, err = time.Parse("01-02", event.StartDate)
		if err != nil {
			return false
		}
		startDate = time.Date(now.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, now.Location())

		endDate, err = time.Parse("01-02", event.EndDate)
		if err != nil {
			return false
		}
		endDate = time.Date(now.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, now.Location())

		// Handle year wrap (e.g., Dec 26 - Jan 2)
		if endDate.Before(startDate) {
			if now.Month() == 12 || now.Month() == 1 {
				if now.Month() == 1 {
					startDate = startDate.AddDate(-1, 0, 0)
				} else {
					endDate = endDate.AddDate(1, 0, 0)
				}
			}
		}
	} else {
		// Full date format (YYYY-MM-DD)
		startDate, err = time.Parse("2006-01-02", event.StartDate)
		if err != nil {
			return false
		}
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, now.Location())

		endDate, err = time.Parse("2006-01-02", event.EndDate)
		if err != nil {
			return false
		}
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, now.Location())
	}

	return !now.Before(startDate) && !now.After(endDate)
}

// getHolidayName returns the holiday name if current date is a holiday, empty string otherwise
func (s *BiddingService) getHolidayName(now time.Time, country string) string {
	month := int(now.Month())
	day := now.Day()
	dayOfWeek := int(now.Weekday())
	year := now.Year()

	// Calculate dynamic holidays
	// Thanksgiving (US): 4th Thursday of November
	thanksgivingDay := s.getNthWeekdayOfMonth(year, 11, time.Thursday, 4)

	// US Holidays
	if country == "US" {
		switch {
		// New Year's Day
		case month == 1 && day == 1:
			return "New Year's Day"
		// Martin Luther King Jr. Day (3rd Monday of January)
		case month == 1 && dayOfWeek == 1 && day >= 15 && day <= 21:
			return "MLK Day"
		// Presidents Day (3rd Monday of February)
		case month == 2 && dayOfWeek == 1 && day >= 15 && day <= 21:
			return "Presidents Day"
		// Memorial Day (Last Monday of May)
		case month == 5 && dayOfWeek == 1 && day >= 25:
			return "Memorial Day"
		// Independence Day
		case month == 7 && day == 4:
			return "Independence Day"
		// Labor Day (1st Monday of September)
		case month == 9 && dayOfWeek == 1 && day <= 7:
			return "Labor Day"
		// Columbus Day (2nd Monday of October)
		case month == 10 && dayOfWeek == 1 && day >= 8 && day <= 14:
			return "Columbus Day"
		// Veterans Day
		case month == 11 && day == 11:
			return "Veterans Day"
		// Thanksgiving (4th Thursday of November)
		case month == 11 && day == thanksgivingDay:
			return "Thanksgiving"
		// Black Friday (day after Thanksgiving)
		case month == 11 && day == thanksgivingDay+1:
			return "Black Friday"
		// Cyber Monday (Monday after Thanksgiving)
		case month == 11 && day == thanksgivingDay+4:
			return "Cyber Monday"
		// Christmas Eve
		case month == 12 && day == 24:
			return "Christmas Eve"
		// Christmas
		case month == 12 && day == 25:
			return "Christmas"
		// New Year's Eve
		case month == 12 && day == 31:
			return "New Year's Eve"
		// Valentine's Day (shopping boost)
		case month == 2 && day == 14:
			return "Valentine's Day"
		// Mother's Day (2nd Sunday of May)
		case month == 5 && dayOfWeek == 0 && day >= 8 && day <= 14:
			return "Mother's Day"
		// Father's Day (3rd Sunday of June)
		case month == 6 && dayOfWeek == 0 && day >= 15 && day <= 21:
			return "Father's Day"
		// Halloween
		case month == 10 && day == 31:
			return "Halloween"
		}
	}

	// UK Holidays
	if country == "UK" {
		switch {
		case month == 1 && day == 1:
			return "New Year's Day"
		case month == 12 && day == 25:
			return "Christmas"
		case month == 12 && day == 26:
			return "Boxing Day"
			// Add more UK holidays as needed
		}
	}

	return ""
}

// getNthWeekdayOfMonth calculates the day of month for the Nth occurrence of a weekday
func (s *BiddingService) getNthWeekdayOfMonth(year int, month int, weekday time.Weekday, n int) int {
	firstOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	firstWeekday := firstOfMonth.Weekday()

	daysUntilWeekday := int(weekday) - int(firstWeekday)
	if daysUntilWeekday < 0 {
		daysUntilWeekday += 7
	}

	return 1 + daysUntilWeekday + (n-1)*7
}

// calculateDemographicMultiplier evaluates user demographics against targeting configuration
func (s *BiddingService) calculateDemographicMultiplier(campaign *model.Campaign, req *model.BidRequest) model.DemographicResult {
	result := model.DemographicResult{
		Matched:    false,
		Blocked:    false,
		Multiplier: 1.0,
	}

	// No demographic targeting configured
	if campaign.Targeting.DemographicTargeting == nil {
		return result
	}

	dt := campaign.Targeting.DemographicTargeting

	// Extract demographic info from request
	age, gender, incomeLevel := s.extractDemographicInfo(req)
	result.UserAge = age
	result.UserGender = gender
	result.IncomeLevel = incomeLevel

	// Check excluded genders
	for _, excluded := range dt.ExcludeGenders {
		if strings.EqualFold(gender, excluded) {
			result.Blocked = true
			result.Reason = "gender_excluded:" + excluded
			return result
		}
	}

	// Check excluded age ranges
	for _, excludedRange := range dt.ExcludeAgeRanges {
		if age > 0 && age >= excludedRange.MinAge && age <= excludedRange.MaxAge {
			result.Blocked = true
			result.Reason = "age_excluded"
			return result
		}
	}

	// Check age ranges
	hasRequiredAge := false
	requiresAge := false
	ageMatched := false

	for _, ageRange := range dt.AgeRanges {
		if ageRange.Required {
			requiresAge = true
		}

		if age > 0 && age >= ageRange.MinAge && age <= ageRange.MaxAge {
			result.Matched = true
			ageMatched = true
			result.AgeRangeMatch = fmt.Sprintf("%d-%d", ageRange.MinAge, ageRange.MaxAge)

			if ageRange.Required {
				hasRequiredAge = true
			}

			boost := ageRange.Boost
			if boost <= 0 {
				boost = 1.2
			}
			result.Multiplier *= boost
		}
	}

	// Block if required age not matched
	if requiresAge && !hasRequiredAge && age > 0 {
		result.Blocked = true
		result.Reason = "missing_required_age_range"
		return result
	}

	// Apply unknown age discount if age not available
	if age == 0 && len(dt.AgeRanges) > 0 && !ageMatched {
		discount := dt.UnknownAgeBoost
		if discount <= 0 {
			discount = 0.8 // Default 20% discount for unknown age
		}
		result.Multiplier *= discount
	}

	// Check gender rules
	hasRequiredGender := false
	requiresGender := false
	genderMatched := false

	for _, genderRule := range dt.Genders {
		if genderRule.Required {
			requiresGender = true
		}

		if strings.EqualFold(gender, genderRule.Gender) {
			result.Matched = true
			genderMatched = true

			if genderRule.Required {
				hasRequiredGender = true
			}

			boost := genderRule.Boost
			if boost <= 0 {
				boost = 1.2
			}
			result.Multiplier *= boost
		}
	}

	// Block if required gender not matched
	if requiresGender && !hasRequiredGender && gender != "" && gender != "unknown" {
		result.Blocked = true
		result.Reason = "missing_required_gender"
		return result
	}

	// Apply unknown gender discount
	if (gender == "" || gender == "unknown") && len(dt.Genders) > 0 && !genderMatched {
		discount := dt.UnknownGenderBoost
		if discount <= 0 {
			discount = 0.8
		}
		result.Multiplier *= discount
	}

	// Check income level rules
	hasRequiredIncome := false
	requiresIncome := false

	for _, incomeRule := range dt.IncomeLevels {
		if incomeRule.Required {
			requiresIncome = true
		}

		if strings.EqualFold(incomeLevel, incomeRule.Level) {
			result.Matched = true

			if incomeRule.Required {
				hasRequiredIncome = true
			}

			boost := incomeRule.Boost
			if boost <= 0 {
				boost = 1.2
			}
			result.Multiplier *= boost
		}
	}

	// Block if required income not matched
	if requiresIncome && !hasRequiredIncome && incomeLevel != "" {
		result.Blocked = true
		result.Reason = "missing_required_income_level"
		return result
	}

	// Cap multiplier
	if result.Multiplier > 3.0 {
		result.Multiplier = 3.0
	}
	if result.Multiplier < 0.3 {
		result.Multiplier = 0.3
	}

	return result
}

// extractDemographicInfo extracts age, gender, and income from request
func (s *BiddingService) extractDemographicInfo(req *model.BidRequest) (age int, gender, incomeLevel string) {
	// Get from user struct
	age = req.User.Age
	gender = req.User.Gender

	// Check context for additional demographic info
	if req.Context != nil {
		if a, ok := req.Context["age"].(float64); ok && age == 0 {
			age = int(a)
		}
		if a, ok := req.Context["user_age"].(float64); ok && age == 0 {
			age = int(a)
		}
		if g, ok := req.Context["gender"].(string); ok && gender == "" {
			gender = g
		}
		if g, ok := req.Context["user_gender"].(string); ok && gender == "" {
			gender = g
		}
		if inc, ok := req.Context["income_level"].(string); ok {
			incomeLevel = inc
		}
		if inc, ok := req.Context["income"].(string); ok && incomeLevel == "" {
			incomeLevel = inc
		}

		// Calculate age from year of birth if available
		if yob, ok := req.Context["yob"].(float64); ok && age == 0 {
			age = time.Now().Year() - int(yob)
		}
	}

	// Normalize gender values
	gender = strings.ToLower(gender)
	switch gender {
	case "m", "male":
		gender = "male"
	case "f", "female":
		gender = "female"
	case "o", "other", "non-binary":
		gender = "other"
	case "":
		gender = "unknown"
	}

	return age, gender, incomeLevel
}

// calculateVideoTargetingMultiplier evaluates video ad requirements against targeting
func (s *BiddingService) calculateVideoTargetingMultiplier(campaign *model.Campaign, req *model.BidRequest) model.VideoTargetingResult {
	result := model.VideoTargetingResult{
		Matched:    false,
		Blocked:    false,
		Multiplier: 1.0,
	}

	// No video targeting configured
	if campaign.Targeting.VideoTargeting == nil {
		return result
	}

	vt := campaign.Targeting.VideoTargeting

	// Extract video context from request
	videoCtx := s.extractVideoContext(req)
	result.PlayerSize = videoCtx.playerSize
	result.Placement = videoCtx.placement
	result.Duration = videoCtx.duration
	result.Skippable = videoCtx.skippable
	result.CompletionRate = videoCtx.completionRate

	// Check if this is video inventory
	if !videoCtx.isVideo {
		// If video targeting configured but not video inventory
		if len(vt.PlayerSizes) > 0 || len(vt.Placements) > 0 {
			result.Blocked = true
			result.Reason = "not_video_inventory"
			return result
		}
		return result
	}

	// Check duration requirements
	if vt.MinDuration > 0 && videoCtx.maxDuration > 0 && videoCtx.maxDuration < vt.MinDuration {
		result.Blocked = true
		result.Reason = "duration_too_short"
		return result
	}
	if vt.MaxDuration > 0 && videoCtx.minDuration > vt.MaxDuration {
		result.Blocked = true
		result.Reason = "duration_too_long"
		return result
	}

	// Check placement targeting
	if len(vt.Placements) > 0 {
		placementMatched := false
		for _, placement := range vt.Placements {
			if strings.EqualFold(placement, videoCtx.placement) {
				placementMatched = true
				result.Matched = true
				break
			}
		}
		if !placementMatched && videoCtx.placement != "" {
			result.Blocked = true
			result.Reason = "placement_mismatch"
			return result
		}
	}

	// Check player size targeting
	if len(vt.PlayerSizes) > 0 {
		sizeMatched := false
		for _, sizeRule := range vt.PlayerSizes {
			if s.matchesPlayerSize(videoCtx.playerWidth, videoCtx.playerHeight, sizeRule) {
				sizeMatched = true
				result.Matched = true
				result.PlayerSize = sizeRule.Size

				boost := sizeRule.Boost
				if boost <= 0 {
					boost = s.getDefaultPlayerSizeBoost(sizeRule.Size)
				}
				result.Multiplier *= boost
				break
			}
		}
		if !sizeMatched {
			// Check if any size is required
			for _, sizeRule := range vt.PlayerSizes {
				if sizeRule.Required {
					result.Blocked = true
					result.Reason = "required_player_size_not_matched"
					return result
				}
			}
		}
	}

	// Check skip settings
	if vt.SkipSettings != nil {
		skipResult := s.evaluateSkipSettings(videoCtx.skippable, videoCtx.skipOffset, vt.SkipSettings)
		if skipResult.blocked {
			result.Blocked = true
			result.Reason = skipResult.reason
			return result
		}
		result.Multiplier *= skipResult.multiplier
	}

	// Check linearity
	if vt.Linearity != nil && videoCtx.linearity > 0 {
		if *vt.Linearity != videoCtx.linearity {
			result.Blocked = true
			result.Reason = "linearity_mismatch"
			return result
		}
	}

	// Check start delay (pre-roll, mid-roll, post-roll)
	if len(vt.StartDelays) > 0 && videoCtx.startDelay != -999 { // -999 = unknown
		delayMatched := false
		for _, delay := range vt.StartDelays {
			if delay == videoCtx.startDelay {
				delayMatched = true
				break
			}
			// Special handling: delay > 0 means mid-roll
			if delay > 0 && videoCtx.startDelay > 0 {
				delayMatched = true
				break
			}
		}
		if !delayMatched {
			result.Blocked = true
			result.Reason = "start_delay_mismatch"
			return result
		}
	}

	// Check completion rate requirements
	if vt.CompletionRates != nil && videoCtx.completionRate > 0 {
		cr := vt.CompletionRates

		if cr.MinCompletionRate > 0 && videoCtx.completionRate < cr.MinCompletionRate {
			result.Blocked = true
			result.Reason = "completion_rate_too_low"
			return result
		}

		// Apply completion rate-based boosts/penalties
		if videoCtx.completionRate >= 0.75 && cr.HighCompletionBoost > 0 {
			result.Multiplier *= cr.HighCompletionBoost
		} else if videoCtx.completionRate < 0.25 && cr.LowCompletionPenalty > 0 {
			result.Multiplier *= cr.LowCompletionPenalty
		}
	}

	// Check MIME types
	if len(vt.Mimes) > 0 && len(videoCtx.mimes) > 0 {
		mimeMatched := false
		for _, reqMime := range videoCtx.mimes {
			for _, targetMime := range vt.Mimes {
				if strings.EqualFold(reqMime, targetMime) {
					mimeMatched = true
					break
				}
			}
			if mimeMatched {
				break
			}
		}
		if !mimeMatched {
			result.Blocked = true
			result.Reason = "mime_type_mismatch"
			return result
		}
	}

	// Boost for large player sizes
	if videoCtx.playerWidth >= 1280 {
		result.Multiplier *= 1.3 // Large/XL player boost
	} else if videoCtx.playerWidth >= 640 {
		result.Multiplier *= 1.1 // Medium player boost
	}

	// Cap multiplier
	if result.Multiplier > 2.5 {
		result.Multiplier = 2.5
	}
	if result.Multiplier < 0.5 {
		result.Multiplier = 0.5
	}

	return result
}

// videoContext holds extracted video information from request
type videoContext struct {
	isVideo        bool
	playerWidth    int
	playerHeight   int
	playerSize     string
	placement      string
	minDuration    int
	maxDuration    int
	duration       int
	skippable      bool
	skipOffset     int
	linearity      int
	startDelay     int
	completionRate float64
	mimes          []string
}

// extractVideoContext extracts video-related info from the request
func (s *BiddingService) extractVideoContext(req *model.BidRequest) videoContext {
	ctx := videoContext{
		isVideo:    false,
		startDelay: -999, // Unknown
	}

	// Check if video inventory from context
	if req.Context != nil {
		if isVideo, ok := req.Context["video"].(bool); ok {
			ctx.isVideo = isVideo
		}
		if adType, ok := req.Context["ad_type"].(string); ok && adType == "video" {
			ctx.isVideo = true
		}
		if w, ok := req.Context["player_width"].(float64); ok {
			ctx.playerWidth = int(w)
		}
		if h, ok := req.Context["player_height"].(float64); ok {
			ctx.playerHeight = int(h)
		}
		if placement, ok := req.Context["video_placement"].(string); ok {
			ctx.placement = placement
		}
		if minDur, ok := req.Context["minduration"].(float64); ok {
			ctx.minDuration = int(minDur)
		}
		if maxDur, ok := req.Context["maxduration"].(float64); ok {
			ctx.maxDuration = int(maxDur)
		}
		if dur, ok := req.Context["duration"].(float64); ok {
			ctx.duration = int(dur)
		}
		if skip, ok := req.Context["skip"].(bool); ok {
			ctx.skippable = skip
		}
		if skipOffset, ok := req.Context["skipafter"].(float64); ok {
			ctx.skipOffset = int(skipOffset)
		}
		if lin, ok := req.Context["linearity"].(float64); ok {
			ctx.linearity = int(lin)
		}
		if delay, ok := req.Context["startdelay"].(float64); ok {
			ctx.startDelay = int(delay)
		}
		if cr, ok := req.Context["completion_rate"].(float64); ok {
			ctx.completionRate = cr
		}
		if mimes, ok := req.Context["video_mimes"].([]interface{}); ok {
			for _, m := range mimes {
				if mimeStr, ok := m.(string); ok {
					ctx.mimes = append(ctx.mimes, mimeStr)
				}
			}
		}
	}

	// Determine player size category
	ctx.playerSize = s.categorizePlayerSize(ctx.playerWidth, ctx.playerHeight)

	return ctx
}

// matchesPlayerSize checks if dimensions match a player size rule
func (s *BiddingService) matchesPlayerSize(width, height int, rule model.VideoPlayerSize) bool {
	if width == 0 {
		return rule.Size == "unknown"
	}

	// Check width range
	if rule.MinWidth > 0 && width < rule.MinWidth {
		return false
	}
	if rule.MaxWidth > 0 && width > rule.MaxWidth {
		return false
	}

	// Check by size category
	category := s.categorizePlayerSize(width, height)
	return strings.EqualFold(category, rule.Size)
}

// categorizePlayerSize determines player size category from dimensions
func (s *BiddingService) categorizePlayerSize(width, height int) string {
	if width == 0 && height == 0 {
		return "unknown"
	}

	// Prefer width for categorization
	size := width
	if size == 0 {
		size = height
	}

	switch {
	case size >= 1280:
		return "xlarge"
	case size >= 640:
		return "large"
	case size >= 400:
		return "medium"
	case size > 0:
		return "small"
	default:
		return "unknown"
	}
}

// getDefaultPlayerSizeBoost returns default boost for player size
func (s *BiddingService) getDefaultPlayerSizeBoost(size string) float64 {
	switch strings.ToLower(size) {
	case "xlarge":
		return 1.4
	case "large":
		return 1.2
	case "medium":
		return 1.0
	case "small":
		return 0.8
	default:
		return 1.0
	}
}

// skipSettingsResult holds skip evaluation result
type skipSettingsResult struct {
	blocked    bool
	multiplier float64
	reason     string
}

// evaluateSkipSettings evaluates skippability against settings
func (s *BiddingService) evaluateSkipSettings(skippable bool, skipOffset int, settings *model.VideoSkipSettings) skipSettingsResult {
	result := skipSettingsResult{
		blocked:    false,
		multiplier: 1.0,
	}

	// Check skippable only requirement
	if settings.SkippableOnly && !skippable {
		result.blocked = true
		result.reason = "requires_skippable"
		return result
	}

	// Check non-skippable only requirement
	if settings.NonSkippableOnly && skippable {
		result.blocked = true
		result.reason = "requires_non_skippable"
		return result
	}

	// Check skip offset requirements
	if skippable && skipOffset > 0 {
		if settings.MinSkipOffset > 0 && skipOffset < settings.MinSkipOffset {
			result.blocked = true
			result.reason = "skip_offset_too_short"
			return result
		}
		if settings.MaxSkipOffset > 0 && skipOffset > settings.MaxSkipOffset {
			result.blocked = true
			result.reason = "skip_offset_too_long"
			return result
		}
	}

	// Apply boosts
	if skippable && settings.SkippableBoost > 0 {
		result.multiplier = settings.SkippableBoost
	} else if !skippable && settings.NonSkipBoost > 0 {
		result.multiplier = settings.NonSkipBoost
	}

	return result
}

// calculatePerformanceGoalMultiplier optimizes bid based on campaign performance goals
func (s *BiddingService) calculatePerformanceGoalMultiplier(campaign *model.Campaign, req *model.BidRequest) model.PerformanceGoalResult {
	result := model.PerformanceGoalResult{
		Matched:           false,
		Blocked:           false,
		Multiplier:        1.0,
		OptimizationLevel: "moderate",
	}

	// No performance goals configured
	if campaign.Targeting.PerformanceGoals == nil {
		return result
	}

	pg := campaign.Targeting.PerformanceGoals
	result.GoalType = pg.PrimaryGoal

	// Get historical performance data
	perfData := s.getHistoricalPerformance(campaign.ID, req)

	// Calculate predicted rates
	result.PredictedCTR = s.predictCTR(campaign, req, perfData)
	result.PredictedCVR = s.predictCVR(campaign, req, perfData)
	result.PredictedViewRate = s.predictViewability(campaign, req, perfData)

	// Additional predictions for specific goals
	result.PredictedInstallRate = s.predictInstallRate(campaign, req, perfData)
	result.PredictedROAS = s.predictROAS(campaign, req, perfData)
	result.PredictedLTV = s.predictLTV(campaign, req, perfData)

	// Check if CTV inventory
	result.IsCTV = s.isCTVInventory(req)
	if result.IsCTV {
		result.HouseholdID = s.getHouseholdID(req)
	}

	// Check performance thresholds
	if pg.Thresholds != nil {
		blocked, reason := s.checkPerformanceThresholds(pg, &result, perfData)
		if blocked && !pg.LearningMode {
			result.Blocked = true
			result.Reason = reason
			return result
		} else if blocked {
			result.Multiplier *= 0.7 // Reduce bid in learning mode
		}
	}

	// Apply optimization based on primary goal
	switch strings.ToLower(pg.PrimaryGoal) {
	case "cpa":
		result.Multiplier *= s.optimizeForCPA(campaign, req, pg, perfData)
	case "cpc":
		result.Multiplier *= s.optimizeForCPC(campaign, req, pg, perfData)
	case "cpm":
		result.Multiplier *= s.optimizeForCPM(campaign, req, pg, perfData)
	case "cpi":
		result.Multiplier *= s.optimizeForCPI(campaign, req, pg, perfData)
	case "cps":
		result.Multiplier *= s.optimizeForCPS(campaign, req, pg, perfData)
	case "cpr":
		result.Multiplier *= s.optimizeForCPR(campaign, req, pg, perfData)
	case "ctv":
		result.Multiplier *= s.optimizeForCTV(campaign, req, pg, perfData)
	case "roas":
		result.Multiplier *= s.optimizeForROAS(campaign, req, pg, perfData)
	case "viewability":
		result.Multiplier *= s.optimizeForViewability(campaign, req, pg, perfData)
	case "completion":
		result.Multiplier *= s.optimizeForCompletion(campaign, req, pg, perfData)
	case "engagement":
		result.Multiplier *= s.optimizeForEngagement(campaign, req, pg, perfData)
	case "cpl":
		result.Multiplier *= s.optimizeForCPL(campaign, req, pg, perfData)
	case "cpv":
		result.Multiplier *= s.optimizeForCPV(campaign, req, pg, perfData)
	case "cpe":
		result.Multiplier *= s.optimizeForCPE(campaign, req, pg, perfData)
	case "vcpm":
		result.Multiplier *= s.optimizeForVCPM(campaign, req, pg, perfData)
	case "cpcv":
		result.Multiplier *= s.optimizeForCPCV(campaign, req, pg, perfData)
	case "dcpm":
		result.Multiplier *= s.optimizeForDCPM(campaign, req, pg, perfData)
	case "cpa_d", "cpad":
		result.Multiplier *= s.optimizeForCPAD(campaign, req, pg, perfData)
	case "cpiaap":
		result.Multiplier *= s.optimizeForCPIAAP(campaign, req, pg, perfData)
	}

	// Apply CTV-specific optimizations if CTV goals configured
	if pg.CTVGoals != nil && result.IsCTV {
		result.Multiplier *= s.applyCTVOptimizations(campaign, req, pg.CTVGoals, perfData)
	}

	// Apply app-specific optimizations
	if pg.AppGoals != nil {
		result.Multiplier *= s.applyAppOptimizations(campaign, req, pg.AppGoals, perfData)
	}

	// Apply e-commerce optimizations
	if pg.EcommerceGoals != nil {
		result.Multiplier *= s.applyEcommerceOptimizations(campaign, req, pg.EcommerceGoals, perfData)
	}

	// Apply bid strategy adjustments
	strategyMult := s.applyBidStrategy(pg, perfData)
	result.Multiplier *= strategyMult

	// Determine optimization level
	result.OptimizationLevel = s.determineOptimizationLevel(pg, perfData)

	// Apply Dayparting Optimization (automatic hourly bid adjustments)
	if pg.DaypartingOptimization != nil && pg.DaypartingOptimization.Enabled && s.daypartingService != nil {
		daypartResult := s.daypartingService.CalculateDaypartMultiplier(campaign, req)
		result.Multiplier *= daypartResult.Multiplier
	}

	// Apply Audience Modeling (lookalike expansion and suppression)
	if pg.AudienceModeling != nil && s.audienceModelingService != nil {
		audienceResult := s.audienceModelingService.EvaluateAudienceModeling(campaign, req)
		if audienceResult.Suppressed {
			result.Blocked = true
			result.Reason = "audience_suppressed:" + audienceResult.Reason
			return result
		}
		result.Multiplier *= audienceResult.Multiplier
	}

	// Apply Attribution Model bid adjustments
	if pg.AttributionModel != "" && s.attributionService != nil && req.User.ID != "" {
		halfLife := pg.TimeDecayHalfLife
		if halfLife <= 0 {
			halfLife = 168 // Default 7 days
		}
		attrMultiplier := s.attributionService.GetAttributionBidAdjustment(campaign.ID, req.User.ID, pg.AttributionModel, halfLife)
		result.Multiplier *= attrMultiplier
	}

	// Apply min/max bid adjustments
	if pg.MaxBidAdjust > 0 && result.Multiplier > pg.MaxBidAdjust {
		result.Multiplier = pg.MaxBidAdjust
	}
	if pg.MinBidAdjust > 0 && result.Multiplier < pg.MinBidAdjust {
		result.Multiplier = pg.MinBidAdjust
	}

	// Calculate recommended bid
	result.RecommendedBid = campaign.BidPrice * result.Multiplier

	// Cap multiplier
	if result.Multiplier > 3.0 {
		result.Multiplier = 3.0
	}
	if result.Multiplier < 0.3 {
		result.Multiplier = 0.3
	}

	result.Matched = true
	return result
}

// performanceData holds historical performance metrics
type performanceData struct {
	impressions    int64
	clicks         int64
	conversions    int64
	installs       int64
	sales          int64
	spend          float64
	revenue        float64
	ctr            float64
	cvr            float64
	cpa            float64
	cpi            float64 // Cost per install
	cps            float64 // Cost per sale
	roas           float64 // Return on ad spend
	ltv            float64 // Lifetime value
	viewability    float64
	completionRate float64
	engagementRate float64
	avgBid         float64
	winRate        float64
}

// getHistoricalPerformance retrieves campaign performance data
func (s *BiddingService) getHistoricalPerformance(campaignID string, req *model.BidRequest) performanceData {
	data := performanceData{
		ctr:         0.01, // Default 1% CTR
		cvr:         0.02, // Default 2% CVR
		viewability: 0.60, // Default 60% viewability
		winRate:     0.15, // Default 15% win rate
	}

	// Try to get from cache
	cacheKey := fmt.Sprintf("perf:%s", campaignID)
	if cached, err := s.cache.Get(cacheKey); err == nil && cached != "" {
		// Parse cached performance data
		if perfMap, ok := s.parsePerformanceCache(cached); ok {
			if imp, exists := perfMap["impressions"]; exists {
				data.impressions = int64(imp)
			}
			if clicks, exists := perfMap["clicks"]; exists {
				data.clicks = int64(clicks)
			}
			if conv, exists := perfMap["conversions"]; exists {
				data.conversions = int64(conv)
			}
			if spend, exists := perfMap["spend"]; exists {
				data.spend = spend
			}
			if ctr, exists := perfMap["ctr"]; exists {
				data.ctr = ctr
			}
			if cvr, exists := perfMap["cvr"]; exists {
				data.cvr = cvr
			}
			if cpa, exists := perfMap["cpa"]; exists {
				data.cpa = cpa
			}
			if view, exists := perfMap["viewability"]; exists {
				data.viewability = view
			}
			if comp, exists := perfMap["completion_rate"]; exists {
				data.completionRate = comp
			}
			if eng, exists := perfMap["engagement_rate"]; exists {
				data.engagementRate = eng
			}
			if avgBid, exists := perfMap["avg_bid"]; exists {
				data.avgBid = avgBid
			}
			if winRate, exists := perfMap["win_rate"]; exists {
				data.winRate = winRate
			}
		}
	}

	// Calculate derived metrics
	if data.impressions > 0 {
		data.ctr = float64(data.clicks) / float64(data.impressions)
	}
	if data.clicks > 0 {
		data.cvr = float64(data.conversions) / float64(data.clicks)
	}
	if data.conversions > 0 && data.spend > 0 {
		data.cpa = data.spend / float64(data.conversions)
	}

	return data
}

// parsePerformanceCache parses cached performance data string
func (s *BiddingService) parsePerformanceCache(cached string) (map[string]float64, bool) {
	result := make(map[string]float64)
	// Simple key:value parsing
	pairs := strings.Split(cached, ",")
	for _, pair := range pairs {
		parts := strings.Split(pair, ":")
		if len(parts) == 2 {
			if val, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
				result[strings.TrimSpace(parts[0])] = val
			}
		}
	}
	return result, len(result) > 0
}

// predictCTR predicts click-through rate for this impression
func (s *BiddingService) predictCTR(campaign *model.Campaign, req *model.BidRequest, perf performanceData) float64 {
	baseCTR := perf.ctr
	if baseCTR == 0 {
		baseCTR = 0.01 // Default 1%
	}

	// Adjust based on context signals
	multiplier := 1.0

	// Device type impact
	switch req.Device.Type {
	case "mobile":
		multiplier *= 1.1 // Mobile tends to have higher CTR
	case "tablet":
		multiplier *= 1.05
	case "desktop":
		multiplier *= 0.95
	}

	// Above fold placement boost
	if req.Context != nil {
		if pos, ok := req.Context["ad_position"].(string); ok && pos == "above_fold" {
			multiplier *= 1.3
		}
	}

	// Category match boost
	if len(campaign.Targeting.Categories) > 0 && len(req.User.Categories) > 0 {
		overlap := countOverlap(campaign.Targeting.Categories, req.User.Categories)
		if overlap > 0 {
			multiplier *= (1.0 + float64(overlap)*0.1)
		}
	}

	return baseCTR * multiplier
}

// predictCVR predicts conversion rate
func (s *BiddingService) predictCVR(campaign *model.Campaign, req *model.BidRequest, perf performanceData) float64 {
	baseCVR := perf.cvr
	if baseCVR == 0 {
		baseCVR = 0.02 // Default 2%
	}

	multiplier := 1.0

	// Previous engagers convert better
	if s.checkRetargetingEligibility(campaign, req.User.ID) {
		multiplier *= 2.0
	}

	// High-value segments from context
	if req.Context != nil {
		if segments, ok := req.Context["user_segments"].([]interface{}); ok {
			for _, seg := range segments {
				if segStr, ok := seg.(string); ok {
					if strings.Contains(strings.ToLower(segStr), "high_intent") ||
						strings.Contains(strings.ToLower(segStr), "converter") {
						multiplier *= 1.5
					}
				}
			}
		}
	}

	// Category match boosts CVR
	if len(campaign.Targeting.Categories) > 0 && len(req.User.Categories) > 0 {
		overlap := countOverlap(campaign.Targeting.Categories, req.User.Categories)
		if overlap > 0 {
			multiplier *= (1.0 + float64(overlap)*0.15)
		}
	}

	return baseCVR * multiplier
}

// predictViewability predicts viewability rate
func (s *BiddingService) predictViewability(campaign *model.Campaign, req *model.BidRequest, perf performanceData) float64 {
	baseView := perf.viewability
	if baseView == 0 {
		baseView = 0.60 // Default 60%
	}

	multiplier := 1.0

	// Position impact
	if req.Context != nil {
		if pos, ok := req.Context["ad_position"].(string); ok {
			switch pos {
			case "above_fold":
				multiplier *= 1.3
			case "below_fold":
				multiplier *= 0.7
			}
		}
	}

	return baseView * multiplier
}

// optimizeForCPA optimizes bid for target CPA
func (s *BiddingService) optimizeForCPA(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if pg.TargetCPA <= 0 {
		return 1.0
	}

	predictedCVR := s.predictCVR(campaign, req, perf)
	predictedCTR := s.predictCTR(campaign, req, perf)

	// Calculate max viable bid for target CPA
	// Expected conversions per impression = CTR * CVR
	expectedConvRate := predictedCTR * predictedCVR
	if expectedConvRate <= 0 {
		expectedConvRate = 0.0001
	}

	// Max bid = Target CPA * Expected conversion rate
	maxBidForCPA := pg.TargetCPA * expectedConvRate

	// Compare to current bid
	if campaign.BidPrice > 0 {
		ratio := maxBidForCPA / campaign.BidPrice
		if ratio > 2.0 {
			return 2.0 // Cap increase
		}
		if ratio < 0.3 {
			return 0.3 // Floor decrease
		}
		return ratio
	}

	return 1.0
}

// optimizeForCPC optimizes bid for target CPC
func (s *BiddingService) optimizeForCPC(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if pg.TargetCPC <= 0 {
		return 1.0
	}

	predictedCTR := s.predictCTR(campaign, req, perf)
	if predictedCTR <= 0 {
		predictedCTR = 0.01
	}

	// Max CPM for target CPC = Target CPC * CTR * 1000
	maxCPM := pg.TargetCPC * predictedCTR * 1000

	// Convert to bid adjustment
	if campaign.BidPrice > 0 {
		ratio := maxCPM / (campaign.BidPrice * 1000)
		if ratio > 2.0 {
			return 2.0
		}
		if ratio < 0.3 {
			return 0.3
		}
		return ratio
	}

	return 1.0
}

// optimizeForCPM simple CPM optimization
func (s *BiddingService) optimizeForCPM(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if pg.TargetCPM <= 0 {
		return 1.0
	}

	// Adjust based on predicted viewability
	viewRate := s.predictViewability(campaign, req, perf)

	// Higher viewability = willing to pay more
	if viewRate >= 0.8 {
		return 1.3
	} else if viewRate >= 0.6 {
		return 1.1
	} else if viewRate < 0.4 {
		return 0.7
	}

	return 1.0
}

// optimizeForViewability optimizes for viewability goal
func (s *BiddingService) optimizeForViewability(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if pg.ViewabilityGoal <= 0 {
		return 1.0
	}

	predictedView := s.predictViewability(campaign, req, perf)

	// Bid more for high-viewability inventory
	if predictedView >= pg.ViewabilityGoal {
		// Above target - bid up
		bonus := (predictedView - pg.ViewabilityGoal) / (1.0 - pg.ViewabilityGoal)
		return 1.0 + bonus*0.5 // Up to 50% boost
	}

	// Below target - bid down
	penalty := (pg.ViewabilityGoal - predictedView) / pg.ViewabilityGoal
	return 1.0 - penalty*0.5 // Up to 50% reduction
}

// optimizeForCompletion optimizes for video completion goal
func (s *BiddingService) optimizeForCompletion(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if pg.CompletionGoal <= 0 {
		return 1.0
	}

	// Get predicted completion rate from context
	predictedCompletion := perf.completionRate
	if req.Context != nil {
		if cr, ok := req.Context["completion_rate"].(float64); ok {
			predictedCompletion = cr
		}
	}
	if predictedCompletion == 0 {
		predictedCompletion = 0.5 // Default 50%
	}

	// Bid more for high-completion inventory
	if predictedCompletion >= pg.CompletionGoal {
		bonus := (predictedCompletion - pg.CompletionGoal) / (1.0 - pg.CompletionGoal)
		return 1.0 + bonus*0.4
	}

	penalty := (pg.CompletionGoal - predictedCompletion) / pg.CompletionGoal
	return 1.0 - penalty*0.4
}

// optimizeForEngagement optimizes for engagement metrics
func (s *BiddingService) optimizeForEngagement(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, _ performanceData) float64 {
	if pg.EngagementGoal <= 0 {
		return 1.0
	}

	// Estimate engagement potential
	engagementScore := 1.0

	// Mobile tends to have higher engagement
	if req.Device.Type == "mobile" {
		engagementScore *= 1.2
	}

	// In-app inventory often has better engagement
	if req.Context != nil {
		if env, ok := req.Context["environment"].(string); ok && env == "in-app" {
			engagementScore *= 1.15
		}
	}

	// Previous engagers
	if s.checkRetargetingEligibility(campaign, req.User.ID) {
		engagementScore *= 1.3
	}

	return engagementScore
}

// optimizeForCPI optimizes bid for target Cost Per Install (app campaigns)
func (s *BiddingService) optimizeForCPI(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	targetCPI := pg.TargetCPI
	if targetCPI <= 0 && pg.AppGoals != nil {
		targetCPI = pg.AppGoals.TargetCostPerInstall
	}
	if targetCPI <= 0 {
		return 1.0
	}

	// Predict install rate
	predictedInstallRate := s.predictInstallRate(campaign, req, perf)
	predictedCTR := s.predictCTR(campaign, req, perf)

	// Calculate max viable bid for target CPI
	// Expected installs per impression = CTR * Install Rate
	expectedInstallRate := predictedCTR * predictedInstallRate
	if expectedInstallRate <= 0 {
		expectedInstallRate = 0.0001
	}

	// Max bid = Target CPI * Expected install rate
	maxBidForCPI := targetCPI * expectedInstallRate

	// Compare to current bid
	if campaign.BidPrice > 0 {
		ratio := maxBidForCPI / campaign.BidPrice
		if ratio > 2.0 {
			return 2.0
		}
		if ratio < 0.3 {
			return 0.3
		}
		return ratio
	}

	return 1.0
}

// optimizeForCPS optimizes bid for target Cost Per Sale (e-commerce)
func (s *BiddingService) optimizeForCPS(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	targetCPS := pg.TargetCPS
	if targetCPS <= 0 && pg.EcommerceGoals != nil {
		targetCPS = pg.EcommerceGoals.TargetCostPerSale
	}
	if targetCPS <= 0 {
		return 1.0
	}

	// Predict purchase rate
	predictedCTR := s.predictCTR(campaign, req, perf)
	predictedCVR := s.predictCVR(campaign, req, perf)

	// Calculate expected purchase rate (CTR * CVR)
	expectedPurchaseRate := predictedCTR * predictedCVR
	if expectedPurchaseRate <= 0 {
		expectedPurchaseRate = 0.0001
	}

	// Max bid = Target CPS * Expected purchase rate
	maxBidForCPS := targetCPS * expectedPurchaseRate

	// Boost for cart abandoners
	if pg.EcommerceGoals != nil && pg.EcommerceGoals.CartAbandonBoost > 0 {
		if s.isCartAbandoner(req) {
			maxBidForCPS *= pg.EcommerceGoals.CartAbandonBoost
		}
	}

	// Boost for repeat customers
	if pg.EcommerceGoals != nil && pg.EcommerceGoals.RepeatCustomerBoost > 0 {
		if s.isRepeatCustomer(req) {
			maxBidForCPS *= pg.EcommerceGoals.RepeatCustomerBoost
		}
	}

	// Compare to current bid
	if campaign.BidPrice > 0 {
		ratio := maxBidForCPS / campaign.BidPrice
		if ratio > 2.5 {
			return 2.5
		}
		if ratio < 0.3 {
			return 0.3
		}
		return ratio
	}

	return 1.0
}

// optimizeForCPR optimizes bid for target Cost Per Registration/Result
func (s *BiddingService) optimizeForCPR(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if pg.TargetCPR <= 0 {
		return 1.0
	}

	// Predict registration/result rate (similar to CVR)
	predictedCTR := s.predictCTR(campaign, req, perf)
	predictedRegRate := s.predictCVR(campaign, req, perf) * 0.8 // Registration typically lower than purchase

	// Calculate expected registration rate
	expectedRegRate := predictedCTR * predictedRegRate
	if expectedRegRate <= 0 {
		expectedRegRate = 0.0001
	}

	// Max bid = Target CPR * Expected registration rate
	maxBidForCPR := pg.TargetCPR * expectedRegRate

	// Compare to current bid
	if campaign.BidPrice > 0 {
		ratio := maxBidForCPR / campaign.BidPrice
		if ratio > 2.0 {
			return 2.0
		}
		if ratio < 0.3 {
			return 0.3
		}
		return ratio
	}

	return 1.0
}

// optimizeForCPL optimizes bid for target Cost Per Lead (B2B/lead-gen campaigns)
func (s *BiddingService) optimizeForCPL(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if pg.TargetCPL <= 0 {
		return 1.0
	}

	predictedCPL := s.predictCPL(campaign, req, perf)
	if predictedCPL <= 0 {
		return 1.0
	}

	// Ratio-based optimization: bid more when predicted CPL is below target
	ratio := pg.TargetCPL / predictedCPL

	// Apply context signals for lead quality
	if req.Context != nil {
		// High-intent users (form interactions, long session duration)
		if intent, ok := req.Context["lead_intent_score"].(float64); ok && intent > 0.7 {
			ratio *= 1.3
		}
		// B2B signals boost
		if isB2B, ok := req.Context["is_b2b"].(bool); ok && isB2B {
			ratio *= 1.2
		}
	}

	// Cap multiplier
	if ratio > 2.5 {
		return 2.5
	}
	if ratio < 0.3 {
		return 0.3
	}
	return ratio
}

// optimizeForCPV optimizes bid for target Cost Per View (video campaigns)
func (s *BiddingService) optimizeForCPV(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if pg.TargetCPV <= 0 {
		return 1.0
	}

	predictedCPV := s.predictCPV(campaign, req, perf)
	if predictedCPV <= 0 {
		return 1.0
	}

	ratio := pg.TargetCPV / predictedCPV

	// Video-specific adjustments
	if req.Context != nil {
		// Sound-on inventory typically yields better view rates
		if soundOn, ok := req.Context["sound_on"].(bool); ok && soundOn {
			ratio *= 1.15
		}
		// In-stream has higher view completion
		if placement, ok := req.Context["video_placement"].(string); ok {
			if strings.EqualFold(placement, "instream") {
				ratio *= 1.2
			}
		}
	}

	if ratio > 2.5 {
		return 2.5
	}
	if ratio < 0.3 {
		return 0.3
	}
	return ratio
}

// optimizeForCPCV optimizes bid for target Cost Per Completed View (video campaigns)
func (s *BiddingService) optimizeForCPCV(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if pg.TargetCPCV <= 0 {
		return 1.0
	}

	predictedCPCV := s.predictCPCV(campaign, req, perf)
	if predictedCPCV <= 0 {
		return 1.0
	}

	ratio := pg.TargetCPCV / predictedCPCV

	// Completed view factors
	if req.Context != nil {
		// Non-skippable inventory guarantees completion
		if skippable, ok := req.Context["skippable"].(bool); ok && !skippable {
			ratio *= 1.4
		}
		// Shorter videos have higher completion rates
		if duration, ok := req.Context["video_duration"].(float64); ok {
			if duration <= 15 {
				ratio *= 1.2 // 15s or less
			} else if duration <= 30 {
				ratio *= 1.0 // 30s
			} else {
				ratio *= 0.8 // Longer videos
			}
		}
		// CTV has highest completion rates
		if s.isCTVInventory(req) {
			ratio *= 1.3
		}
	}

	if ratio > 3.0 {
		return 3.0
	}
	if ratio < 0.3 {
		return 0.3
	}
	return ratio
}

// optimizeForCPE optimizes bid for target Cost Per Engagement (social/interactive campaigns)
func (s *BiddingService) optimizeForCPE(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if pg.TargetCPE <= 0 {
		return 1.0
	}

	predictedCPE := s.predictCPE(campaign, req, perf)
	if predictedCPE <= 0 {
		return 1.0
	}

	ratio := pg.TargetCPE / predictedCPE

	// Engagement-specific signals
	if req.Device.Type == "mobile" {
		ratio *= 1.15 // Mobile has higher engagement rates
	}
	if req.Context != nil {
		// Rich media / playable ads drive higher engagement
		if creative, ok := req.Context["creative_type"].(string); ok {
			switch strings.ToLower(creative) {
			case "rich_media", "playable":
				ratio *= 1.3
			case "interactive":
				ratio *= 1.25
			}
		}
		// In-app inventory tends to have better engagement
		if env, ok := req.Context["environment"].(string); ok && env == "in-app" {
			ratio *= 1.1
		}
	}

	if ratio > 2.5 {
		return 2.5
	}
	if ratio < 0.3 {
		return 0.3
	}
	return ratio
}

// optimizeForVCPM optimizes bid for Viewable CPM (viewability-based buying)
func (s *BiddingService) optimizeForVCPM(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if pg.TargetVCPM <= 0 {
		return 1.0
	}

	predictedViewability := s.predictViewability(campaign, req, perf)
	if predictedViewability <= 0 {
		predictedViewability = 0.5
	}

	// Effective CPM = Target vCPM * viewability rate
	// We pay for all impressions but value only viewable ones
	effectiveCPM := pg.TargetVCPM * predictedViewability

	if campaign.BidPrice > 0 {
		ratio := effectiveCPM / (campaign.BidPrice * 1000)
		// Boost high-viewability inventory
		if predictedViewability >= 0.8 {
			ratio *= 1.3
		} else if predictedViewability < 0.4 {
			ratio *= 0.5 // Heavy penalty for low viewability
		}
		if ratio > 2.5 {
			return 2.5
		}
		if ratio < 0.2 {
			return 0.2
		}
		return ratio
	}

	return 1.0
}

// optimizeForDCPM optimizes bid using Dynamic CPM (ML-optimized pricing)
func (s *BiddingService) optimizeForDCPM(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if pg.TargetDCPM <= 0 {
		return 1.0
	}

	// Dynamic CPM uses multiple signals to determine optimal bid
	// Combines CTR prediction, viewability, engagement, and win rate
	predictedCTR := s.predictCTR(campaign, req, perf)
	predictedView := s.predictViewability(campaign, req, perf)
	winRate := perf.winRate
	if winRate <= 0 {
		winRate = 0.15
	}

	// Composite quality score (0-1 range)
	qualityScore := (predictedCTR*100*0.3 + predictedView*0.3 + perf.engagementRate*0.2 + winRate*0.2)
	if qualityScore > 1.0 {
		qualityScore = 1.0
	}
	if qualityScore < 0.1 {
		qualityScore = 0.1
	}

	// Dynamic adjustment: pay more for high-quality impressions
	baseDCPM := pg.TargetDCPM * qualityScore

	if campaign.BidPrice > 0 {
		ratio := baseDCPM / (campaign.BidPrice * 1000)

		// Win rate feedback: bid higher if winning too little, lower if winning too much
		if winRate < 0.1 {
			ratio *= 1.3 // Not winning enough
		} else if winRate > 0.4 {
			ratio *= 0.8 // Winning too much (overpaying)
		}

		if ratio > 2.5 {
			return 2.5
		}
		if ratio < 0.3 {
			return 0.3
		}
		return ratio
	}

	return 1.0
}

// optimizeForCPAD optimizes bid for Cost Per App Download
func (s *BiddingService) optimizeForCPAD(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	targetCPAD := pg.TargetCPAD
	if targetCPAD <= 0 {
		return 1.0
	}

	// App download is similar to CPI but includes both organic uplift and direct installs
	predictedCTR := s.predictCTR(campaign, req, perf)
	predictedDownloadRate := s.predictInstallRate(campaign, req, perf)

	// Organic uplift factor: some ad exposure leads to organic downloads
	organicUplift := 1.15 // 15% average organic uplift

	expectedDownloadRate := predictedCTR * predictedDownloadRate * organicUplift
	if expectedDownloadRate <= 0 {
		expectedDownloadRate = 0.0001
	}

	maxBid := targetCPAD * expectedDownloadRate

	// Platform-specific adjustments
	if req.Context != nil {
		// App store page optimization signals
		if appRating, ok := req.Context["app_rating"].(float64); ok && appRating >= 4.5 {
			maxBid *= 1.2 // High-rated apps convert better
		}
		// Featured in app store
		if featured, ok := req.Context["app_featured"].(bool); ok && featured {
			maxBid *= 1.15
		}
	}

	// Non-mobile devices rarely download apps
	if req.Device.Type != "mobile" && req.Device.Type != "tablet" {
		maxBid *= 0.1
	}

	if campaign.BidPrice > 0 {
		ratio := maxBid / campaign.BidPrice
		if ratio > 2.5 {
			return 2.5
		}
		if ratio < 0.2 {
			return 0.2
		}
		return ratio
	}

	return 1.0
}

// optimizeForCPIAAP optimizes bid for Cost Per In-App Purchase
func (s *BiddingService) optimizeForCPIAAP(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	targetCPIAAP := pg.TargetCPIAAP
	if targetCPIAAP <= 0 {
		return 1.0
	}

	// In-app purchase funnel: impression -> click -> install -> open -> purchase
	predictedCTR := s.predictCTR(campaign, req, perf)
	predictedInstallRate := s.predictInstallRate(campaign, req, perf)

	// Predict in-app purchase rate from installs (typically 2-5% of installs make a purchase)
	iapRate := 0.03 // Default 3% IAP rate
	if req.Context != nil {
		if rate, ok := req.Context["historical_iap_rate"].(float64); ok {
			iapRate = rate
		}
		// Spending propensity from user data
		if propensity, ok := req.Context["purchase_propensity"].(float64); ok {
			iapRate *= (1.0 + propensity)
		}
	}

	// Full funnel: CTR * InstallRate * IAPRate
	expectedIAPRate := predictedCTR * predictedInstallRate * iapRate
	if expectedIAPRate <= 0 {
		expectedIAPRate = 0.00001
	}

	maxBid := targetCPIAAP * expectedIAPRate

	// Boost for known spenders (whale targeting)
	if req.Context != nil {
		if avgPurchase, ok := req.Context["avg_iap_value"].(float64); ok && avgPurchase > 10 {
			maxBid *= 1.5 // High-value purchasers
		}
		if segments, ok := req.Context["user_segments"].([]interface{}); ok {
			for _, seg := range segments {
				if segStr, ok := seg.(string); ok {
					if strings.Contains(strings.ToLower(segStr), "spender") ||
						strings.Contains(strings.ToLower(segStr), "whale") {
						maxBid *= 1.8
						break
					}
				}
			}
		}
	}

	// Only mobile/tablet relevant for IAP
	if req.Device.Type != "mobile" && req.Device.Type != "tablet" {
		maxBid *= 0.05
	}

	if campaign.BidPrice > 0 {
		ratio := maxBid / campaign.BidPrice
		if ratio > 3.0 {
			return 3.0
		}
		if ratio < 0.2 {
			return 0.2
		}
		return ratio
	}

	return 1.0
}

// optimizeForCTV optimizes bid for Connected TV campaigns
func (s *BiddingService) optimizeForCTV(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	if !s.isCTVInventory(req) {
		return 0.5 // Significant penalty for non-CTV when CTV is goal
	}

	multiplier := 1.0

	// CTV inventory premium
	multiplier *= 1.3

	// Apply CTV-specific goals if configured
	if pg.CTVGoals != nil {
		ctv := pg.CTVGoals

		// Check completion rate target
		predictedCompletion := perf.completionRate
		if predictedCompletion == 0 {
			predictedCompletion = 0.85 // CTV typically has higher completion
		}
		if ctv.TargetCompletionRate > 0 {
			if predictedCompletion >= ctv.TargetCompletionRate {
				multiplier *= 1.2
			} else {
				multiplier *= 0.8
			}
		}

		// Primetime boost
		if ctv.PrimtimeBoost > 0 && s.isPrimetime(req) {
			multiplier *= ctv.PrimtimeBoost
		}

		// Live content boost
		if ctv.LiveContentBoost > 0 && s.isLiveContent(req) {
			multiplier *= ctv.LiveContentBoost
		}

		// Co-viewing boost
		if ctv.CoViewingBoost > 0 && s.isCoViewing(req) {
			multiplier *= ctv.CoViewingBoost
		}

		// Preferred device boost
		if len(ctv.PreferredDevices) > 0 {
			device := s.getCTVDevice(req)
			for _, preferred := range ctv.PreferredDevices {
				if strings.EqualFold(device, preferred) {
					multiplier *= 1.15
					break
				}
			}
		}
	}

	return multiplier
}

// optimizeForROAS optimizes bid for target Return On Ad Spend
func (s *BiddingService) optimizeForROAS(campaign *model.Campaign, req *model.BidRequest, pg *model.PerformanceGoals, perf performanceData) float64 {
	targetROAS := pg.TargetROAS
	if targetROAS <= 0 && pg.EcommerceGoals != nil {
		targetROAS = pg.EcommerceGoals.TargetROAS
	}
	if targetROAS <= 0 {
		return 1.0
	}

	// Predict ROAS
	predictedROAS := s.predictROAS(campaign, req, perf)
	if predictedROAS <= 0 {
		predictedROAS = 1.0 // Break-even default
	}

	// Compare predicted to target
	if predictedROAS >= targetROAS {
		// Above target - can bid more aggressively
		bonus := (predictedROAS - targetROAS) / targetROAS
		if bonus > 0.5 {
			bonus = 0.5
		}
		return 1.0 + bonus
	}

	// Below target - bid more conservatively
	penalty := (targetROAS - predictedROAS) / targetROAS
	if penalty > 0.5 {
		penalty = 0.5
	}
	return 1.0 - penalty
}

// predictInstallRate predicts app install rate
func (s *BiddingService) predictInstallRate(_ *model.Campaign, req *model.BidRequest, _ performanceData) float64 {
	baseRate := 0.05 // Default 5% install rate from clicks

	// Higher install rate for in-app inventory
	if req.Context != nil {
		if env, ok := req.Context["environment"].(string); ok && env == "in-app" {
			baseRate *= 1.5
		}

		// Historical install rate if available
		if ir, ok := req.Context["historical_install_rate"].(float64); ok {
			baseRate = ir
		}
	}

	// Mobile-only for app installs
	if req.Device.Type != "mobile" && req.Device.Type != "tablet" {
		baseRate *= 0.1 // Desktop very unlikely to convert for app
	}

	// OS matching boosts install rate
	if req.Context != nil {
		if os, ok := req.Context["device_os"].(string); ok {
			// Check if campaign targets this OS
			// This would need campaign targeting info
			_ = os
		}
	}

	return baseRate
}

// predictROAS predicts return on ad spend
func (s *BiddingService) predictROAS(_ *model.Campaign, req *model.BidRequest, _ performanceData) float64 {
	baseROAS := 2.0 // Default 2x return

	// Historical ROAS if available
	if req.Context != nil {
		if roas, ok := req.Context["historical_roas"].(float64); ok {
			baseROAS = roas
		}
	}

	// Adjust based on user signals
	multiplier := 1.0

	// Repeat customers have higher ROAS
	if s.isRepeatCustomer(req) {
		multiplier *= 1.5
	}

	// Cart abandoners have higher intent
	if s.isCartAbandoner(req) {
		multiplier *= 1.3
	}

	// High-value audience segments
	if req.Context != nil {
		if segments, ok := req.Context["user_segments"].([]interface{}); ok {
			for _, seg := range segments {
				if segStr, ok := seg.(string); ok {
					if strings.Contains(strings.ToLower(segStr), "high_value") ||
						strings.Contains(strings.ToLower(segStr), "frequent_buyer") {
						multiplier *= 1.4
						break
					}
				}
			}
		}
	}

	return baseROAS * multiplier
}

// predictLTV predicts customer lifetime value
func (s *BiddingService) predictLTV(_ *model.Campaign, req *model.BidRequest, _ performanceData) float64 {
	baseLTV := 50.0 // Default $50 LTV

	// Historical LTV if available
	if req.Context != nil {
		if ltv, ok := req.Context["predicted_ltv"].(float64); ok {
			baseLTV = ltv
		}
	}

	multiplier := 1.0

	// Repeat customers have proven LTV
	if s.isRepeatCustomer(req) {
		multiplier *= 2.0
	}

	// High-engagement users
	if req.Context != nil {
		if engagement, ok := req.Context["engagement_score"].(float64); ok {
			if engagement > 0.7 {
				multiplier *= 1.3
			}
		}
	}

	return baseLTV * multiplier
}

// predictCPL predicts cost per lead for current impression context
func (s *BiddingService) predictCPL(campaign *model.Campaign, req *model.BidRequest, perf performanceData) float64 {
	// CPL = Spend / Leads = Spend / (Clicks * LeadRate)
	// Predicted CPL from impressions = BidPrice / (CTR * LeadRate * 1000)
	predictedCTR := s.predictCTR(campaign, req, perf)
	if predictedCTR <= 0 {
		predictedCTR = 0.01
	}

	// Lead rate: typically 3-10% of clicks result in leads (form fills, signups)
	leadRate := 0.05 // Default 5%
	if req.Context != nil {
		if lr, ok := req.Context["historical_lead_rate"].(float64); ok {
			leadRate = lr
		}
	}
	// B2B has lower volume but higher quality leads
	if req.Context != nil {
		if isB2B, ok := req.Context["is_b2b"].(bool); ok && isB2B {
			leadRate *= 0.7 // Lower rate but higher value
		}
	}

	leadsPerImpression := predictedCTR * leadRate
	if leadsPerImpression <= 0 {
		return 0
	}

	return campaign.BidPrice / leadsPerImpression
}

// predictCPV predicts cost per view for video campaigns
func (s *BiddingService) predictCPV(campaign *model.Campaign, req *model.BidRequest, perf performanceData) float64 {
	// CPV = Spend / Views
	// View is typically defined as 30s or completion of shorter video
	viewRate := perf.viewability // Use viewability as proxy for view rate
	if viewRate <= 0 {
		viewRate = 0.5 // Default 50% view rate
	}

	// Adjust for video context
	if req.Context != nil {
		if vr, ok := req.Context["predicted_view_rate"].(float64); ok {
			viewRate = vr
		}
		// Auto-play inventory has higher initial view rate
		if autoplay, ok := req.Context["autoplay"].(bool); ok && autoplay {
			viewRate *= 1.2
		}
	}

	if viewRate <= 0 {
		return 0
	}

	// CPV = Cost per impression / view rate
	return campaign.BidPrice / viewRate
}

// predictCPCV predicts cost per completed view for video campaigns
func (s *BiddingService) predictCPCV(campaign *model.Campaign, req *model.BidRequest, perf performanceData) float64 {
	// CPCV = Spend / CompletedViews
	completionRate := perf.completionRate
	if completionRate <= 0 {
		completionRate = 0.4 // Default 40% completion rate
	}

	// Context adjustments
	if req.Context != nil {
		if cr, ok := req.Context["predicted_completion_rate"].(float64); ok {
			completionRate = cr
		}
		// Non-skippable has near 100% completion
		if skippable, ok := req.Context["skippable"].(bool); ok && !skippable {
			completionRate = 0.95
		}
		// CTV has higher completion
		if s.isCTVInventory(req) {
			completionRate *= 1.3
			if completionRate > 0.98 {
				completionRate = 0.98
			}
		}
	}

	if completionRate <= 0 {
		return 0
	}

	return campaign.BidPrice / completionRate
}

// predictCPE predicts cost per engagement for interactive campaigns
func (s *BiddingService) predictCPE(campaign *model.Campaign, req *model.BidRequest, perf performanceData) float64 {
	// CPE = Spend / Engagements
	// Engagements: swipes, taps, interactions, video plays, expansions
	engagementRate := perf.engagementRate
	if engagementRate <= 0 {
		engagementRate = 0.03 // Default 3% engagement rate
	}

	// Context adjustments
	if req.Context != nil {
		if er, ok := req.Context["predicted_engagement_rate"].(float64); ok {
			engagementRate = er
		}
	}

	// Mobile tends to have higher engagement
	if req.Device.Type == "mobile" {
		engagementRate *= 1.2
	}

	// Rich media / interactive formats have higher engagement
	if campaign.Creative.Type == "rich_media" || campaign.Creative.Type == "playable" {
		engagementRate *= 1.5
	}

	if engagementRate <= 0 {
		return 0
	}

	return campaign.BidPrice / engagementRate
}

// isCTVInventory checks if request is from CTV device
func (s *BiddingService) isCTVInventory(req *model.BidRequest) bool {
	// Check device type
	deviceType := strings.ToLower(req.Device.Type)
	if deviceType == "ctv" || deviceType == "tv" || deviceType == "connected_tv" {
		return true
	}

	// Check context
	if req.Context != nil {
		if ctv, ok := req.Context["is_ctv"].(bool); ok && ctv {
			return true
		}
		if env, ok := req.Context["environment"].(string); ok {
			if strings.Contains(strings.ToLower(env), "ctv") ||
				strings.Contains(strings.ToLower(env), "connected_tv") ||
				strings.Contains(strings.ToLower(env), "ott") {
				return true
			}
		}
		if dt, ok := req.Context["device_type"].(string); ok {
			dt = strings.ToLower(dt)
			if dt == "ctv" || dt == "smarttv" || dt == "connected_tv" ||
				dt == "roku" || dt == "fire_tv" || dt == "apple_tv" ||
				dt == "gaming_console" || dt == "chromecast" {
				return true
			}
		}
	}

	return false
}

// getHouseholdID extracts household ID for CTV
func (s *BiddingService) getHouseholdID(req *model.BidRequest) string {
	if req.Context != nil {
		if hhid, ok := req.Context["household_id"].(string); ok {
			return hhid
		}
		if hhid, ok := req.Context["hh_id"].(string); ok {
			return hhid
		}
		if ifa, ok := req.Context["ifa"].(string); ok {
			return ifa // IFA can serve as household proxy for CTV
		}
	}
	return ""
}

// checkPerformanceThresholds checks all performance thresholds
func (s *BiddingService) checkPerformanceThresholds(pg *model.PerformanceGoals, result *model.PerformanceGoalResult, perf performanceData) (bool, string) {
	th := pg.Thresholds
	if th == nil {
		return false, ""
	}

	if th.MinCTR > 0 && result.PredictedCTR < th.MinCTR {
		return true, "predicted_ctr_below_threshold"
	}
	if th.MinViewability > 0 && result.PredictedViewRate < th.MinViewability {
		return true, "predicted_viewability_below_threshold"
	}
	if th.MinInstallRate > 0 && result.PredictedInstallRate < th.MinInstallRate {
		return true, "predicted_install_rate_below_threshold"
	}
	if th.MinROAS > 0 && result.PredictedROAS < th.MinROAS {
		return true, "predicted_roas_below_threshold"
	}
	if th.MaxCPA > 0 && perf.cpa > th.MaxCPA {
		return true, "historical_cpa_above_threshold"
	}
	if th.MaxCPI > 0 && perf.cpi > th.MaxCPI {
		return true, "historical_cpi_above_threshold"
	}
	if th.MaxCPS > 0 && perf.cps > th.MaxCPS {
		return true, "historical_cps_above_threshold"
	}

	return false, ""
}

// applyCTVOptimizations applies CTV-specific bid adjustments
func (s *BiddingService) applyCTVOptimizations(campaign *model.Campaign, req *model.BidRequest, ctv *model.CTVOptimization, _ performanceData) float64 {
	if ctv == nil {
		return 1.0
	}

	multiplier := 1.0

	// Primetime hours boost
	if ctv.PrimtimeBoost > 0 && s.isPrimetime(req) {
		multiplier *= ctv.PrimtimeBoost
	}

	// Live content boost
	if ctv.LiveContentBoost > 0 && s.isLiveContent(req) {
		multiplier *= ctv.LiveContentBoost
	}

	// Co-viewing potential
	if ctv.CoViewingBoost > 0 && s.isCoViewing(req) {
		multiplier *= ctv.CoViewingBoost
	}

	// Household frequency check
	if ctv.HouseholdFrequencyCap > 0 {
		hhID := s.getHouseholdID(req)
		if hhID != "" {
			impressions := s.getHouseholdImpressions(campaign.ID, hhID)
			if impressions >= ctv.HouseholdFrequencyCap {
				return 0.1 // Near-block if at cap
			}
			// Reduce bid as approaching cap
			remainingCapacity := float64(ctv.HouseholdFrequencyCap-impressions) / float64(ctv.HouseholdFrequencyCap)
			multiplier *= (0.5 + 0.5*remainingCapacity)
		}
	}

	// Preferred streaming apps
	if len(ctv.PreferredApps) > 0 {
		appName := s.getAppName(req)
		for _, preferred := range ctv.PreferredApps {
			if strings.EqualFold(appName, preferred) {
				multiplier *= 1.2
				break
			}
		}
	}

	return multiplier
}

// applyAppOptimizations applies app campaign-specific optimizations
func (s *BiddingService) applyAppOptimizations(_ *model.Campaign, req *model.BidRequest, app *model.AppOptimization, _ performanceData) float64 {
	if app == nil {
		return 1.0
	}

	multiplier := 1.0

	// Preferred placement boost
	if len(app.PreferredPlacements) > 0 {
		placement := s.getPlacement(req)
		for _, preferred := range app.PreferredPlacements {
			if strings.EqualFold(placement, preferred) {
				// Rewarded video typically performs best
				if strings.EqualFold(placement, "rewarded") {
					multiplier *= 1.4
				} else {
					multiplier *= 1.2
				}
				break
			}
		}
	}

	// Value high-LTV sources
	if !app.ExcludeLowLTVSources {
		// Check if source is known low-LTV
		if s.isLowLTVSource(req) {
			multiplier *= 0.5
		}
	}

	// SKAdNetwork optimization for iOS
	if app.SKAdNetworkOptimized {
		if s.isIOSDevice(req) {
			// Apply SKAdNetwork-aware bidding
			multiplier *= s.getSKAdNetworkMultiplier(req)
		}
	}

	return multiplier
}

// applyEcommerceOptimizations applies e-commerce specific optimizations
func (s *BiddingService) applyEcommerceOptimizations(_ *model.Campaign, req *model.BidRequest, ecom *model.EcommerceOptimization, _ performanceData) float64 {
	if ecom == nil {
		return 1.0
	}

	multiplier := 1.0

	// Cart abandoner boost
	if ecom.CartAbandonBoost > 0 && s.isCartAbandoner(req) {
		multiplier *= ecom.CartAbandonBoost
	}

	// Repeat customer boost
	if ecom.RepeatCustomerBoost > 0 && s.isRepeatCustomer(req) {
		multiplier *= ecom.RepeatCustomerBoost
	}

	// New customer priority
	if ecom.NewCustomerPriority && !s.isRepeatCustomer(req) {
		multiplier *= 1.2
	}

	// Seasonal adjustments
	if len(ecom.SeasonalAdjustments) > 0 {
		season := s.getCurrentSeason()
		if adj, ok := ecom.SeasonalAdjustments[season]; ok {
			multiplier *= adj
		}
	}

	return multiplier
}

// Helper functions for CTV/App/Ecommerce optimization

// isPrimetime checks if current time is primetime (7pm-11pm local)
func (s *BiddingService) isPrimetime(req *model.BidRequest) bool {
	hour := time.Now().Hour()
	// Default primetime: 7pm-11pm
	if req.Context != nil {
		if tz, ok := req.Context["timezone"].(string); ok {
			if loc, err := time.LoadLocation(tz); err == nil {
				hour = time.Now().In(loc).Hour()
			}
		}
	}
	return hour >= 19 && hour <= 23
}

// isLiveContent checks if content is live streaming
func (s *BiddingService) isLiveContent(req *model.BidRequest) bool {
	if req.Context != nil {
		if live, ok := req.Context["is_live"].(bool); ok {
			return live
		}
		if content, ok := req.Context["content_type"].(string); ok {
			return strings.Contains(strings.ToLower(content), "live")
		}
	}
	return false
}

// isCoViewing checks if co-viewing household
func (s *BiddingService) isCoViewing(req *model.BidRequest) bool {
	if req.Context != nil {
		if coview, ok := req.Context["co_viewing"].(bool); ok {
			return coview
		}
		if viewers, ok := req.Context["household_viewers"].(float64); ok {
			return viewers > 1
		}
	}
	return false
}

// getCTVDevice returns the CTV device type
func (s *BiddingService) getCTVDevice(req *model.BidRequest) string {
	if req.Context != nil {
		if device, ok := req.Context["ctv_device"].(string); ok {
			return device
		}
		if device, ok := req.Context["device_make"].(string); ok {
			return device
		}
	}
	return req.Device.Type
}

// getHouseholdImpressions gets impression count for household
func (s *BiddingService) getHouseholdImpressions(campaignID, householdID string) int {
	key := fmt.Sprintf("hh_freq:%s:%s", campaignID, householdID)
	if count, err := s.cache.Get(key); err == nil && count != "" {
		if n, err := strconv.Atoi(count); err == nil {
			return n
		}
	}
	return 0
}

// getAppName extracts app name from request
func (s *BiddingService) getAppName(req *model.BidRequest) string {
	if req.Context != nil {
		if name, ok := req.Context["app_name"].(string); ok {
			return name
		}
		if name, ok := req.Context["bundle"].(string); ok {
			return name
		}
	}
	return ""
}

// getPlacement returns ad placement type
func (s *BiddingService) getPlacement(req *model.BidRequest) string {
	if req.Context != nil {
		if p, ok := req.Context["placement"].(string); ok {
			return p
		}
		if p, ok := req.Context["ad_type"].(string); ok {
			return p
		}
	}
	return ""
}

// isLowLTVSource checks if traffic source is known low-LTV
func (s *BiddingService) isLowLTVSource(req *model.BidRequest) bool {
	if req.Context != nil {
		if ltv, ok := req.Context["source_ltv_score"].(float64); ok {
			return ltv < 0.3
		}
		if lowLtv, ok := req.Context["low_ltv_source"].(bool); ok {
			return lowLtv
		}
	}
	return false
}

// isIOSDevice checks if device is iOS
func (s *BiddingService) isIOSDevice(req *model.BidRequest) bool {
	os := strings.ToLower(req.Device.OS)
	return os == "ios" || os == "iphone" || os == "ipad"
}

// getSKAdNetworkMultiplier returns bid adjustment for SKAdNetwork
func (s *BiddingService) getSKAdNetworkMultiplier(req *model.BidRequest) float64 {
	// SKAdNetwork has limited signal, adjust conservatively
	if req.Context != nil {
		if skadSupported, ok := req.Context["skadn_supported"].(bool); ok && skadSupported {
			return 1.1 // Slight boost for SKAN-supported inventory
		}
	}
	return 1.0
}

// isCartAbandoner checks if user is a cart abandoner
func (s *BiddingService) isCartAbandoner(req *model.BidRequest) bool {
	if req.Context != nil {
		if abandon, ok := req.Context["cart_abandoner"].(bool); ok {
			return abandon
		}
		if segments, ok := req.Context["user_segments"].([]interface{}); ok {
			for _, seg := range segments {
				if segStr, ok := seg.(string); ok {
					if strings.Contains(strings.ToLower(segStr), "cart_abandon") {
						return true
					}
				}
			}
		}
	}
	return false
}

// isRepeatCustomer checks if user is a repeat customer
func (s *BiddingService) isRepeatCustomer(req *model.BidRequest) bool {
	if req.Context != nil {
		if repeat, ok := req.Context["repeat_customer"].(bool); ok {
			return repeat
		}
		if purchases, ok := req.Context["purchase_count"].(float64); ok {
			return purchases > 0
		}
	}
	return false
}

// getCurrentSeason returns current season/period name
func (s *BiddingService) getCurrentSeason() string {
	now := time.Now()
	month := now.Month()

	// Check for major shopping periods
	if month == 11 && now.Day() >= 20 {
		return "black_friday"
	}
	if month == 12 {
		return "holiday"
	}
	if month == 1 && now.Day() <= 15 {
		return "new_year"
	}

	// Standard seasons
	switch {
	case month >= 3 && month <= 5:
		return "spring"
	case month >= 6 && month <= 8:
		return "summer"
	case month >= 9 && month <= 11:
		return "fall"
	default:
		return "winter"
	}
}

// applyBidStrategy applies bid strategy adjustments
func (s *BiddingService) applyBidStrategy(pg *model.PerformanceGoals, perf performanceData) float64 {
	switch strings.ToLower(pg.BidStrategy) {
	case "maximize_conversions":
		// Bid aggressively on high-converting traffic
		if perf.cvr > 0.03 { // Above average CVR
			return 1.4
		}
		return 1.0

	case "target_cpa":
		// More conservative, stay close to target
		if perf.cpa > 0 && pg.TargetCPA > 0 {
			ratio := pg.TargetCPA / perf.cpa
			if ratio > 1.2 {
				return 1.2 // Room to increase
			}
			if ratio < 0.8 {
				return 0.8 // Need to decrease
			}
			return ratio
		}
		return 1.0

	case "maximize_clicks":
		// Bid more on high-CTR inventory
		if perf.ctr > 0.015 { // Above average CTR
			return 1.3
		}
		return 1.0

	case "manual":
		return 1.0 // No automatic adjustment

	default:
		return 1.0
	}
}

// determineOptimizationLevel determines how aggressively to optimize
func (s *BiddingService) determineOptimizationLevel(pg *model.PerformanceGoals, perf performanceData) string {
	// Learning mode = conservative
	if pg.LearningMode {
		return "conservative"
	}

	// Not enough data = conservative
	if perf.impressions < 1000 {
		return "conservative"
	}

	// Performing well = aggressive
	if pg.TargetCPA > 0 && perf.cpa > 0 && perf.cpa < pg.TargetCPA*0.8 {
		return "aggressive"
	}

	return "moderate"
}

// calculateInventoryQualityMultiplier evaluates inventory quality for targeting
func (s *BiddingService) calculateInventoryQualityMultiplier(campaign *model.Campaign, req *model.BidRequest) model.InventoryQualityResult {
	result := model.InventoryQualityResult{
		Matched:    false,
		Blocked:    false,
		Multiplier: 1.0,
		BrandSafe:  true,
	}

	// No inventory quality targeting configured
	if campaign.Targeting.InventoryQuality == nil {
		return result
	}

	iq := campaign.Targeting.InventoryQuality

	// Extract inventory quality signals from request
	qualityCtx := s.extractInventoryQualityContext(req)
	result.QualityScore = qualityCtx.qualityScore
	result.TrustLevel = qualityCtx.trustLevel
	result.AdsTxtVerified = qualityCtx.adsTxtVerified
	result.ViewabilityRate = qualityCtx.viewabilityRate
	result.FraudRisk = qualityCtx.fraudRisk

	// Check minimum quality score
	if iq.MinQualityScore > 0 && qualityCtx.qualityScore < iq.MinQualityScore {
		result.Blocked = true
		result.Reason = "quality_score_too_low"
		return result
	}

	// Check maximum quality score (for cost control)
	if iq.MaxQualityScore > 0 && qualityCtx.qualityScore > iq.MaxQualityScore {
		result.Blocked = true
		result.Reason = "quality_score_too_high"
		return result
	}

	// Check trust levels
	if len(iq.TrustLevels) > 0 {
		trustMatched := false
		for _, level := range iq.TrustLevels {
			if strings.EqualFold(level, qualityCtx.trustLevel) {
				trustMatched = true
				break
			}
		}
		if !trustMatched && qualityCtx.trustLevel != "" {
			result.Blocked = true
			result.Reason = "trust_level_not_allowed"
			return result
		}
	}

	// Check excluded trust levels
	for _, excluded := range iq.ExcludeTrustLevels {
		if strings.EqualFold(excluded, qualityCtx.trustLevel) {
			result.Blocked = true
			result.Reason = "trust_level_excluded:" + excluded
			return result
		}
	}

	// Check ads.txt requirement
	if iq.RequireAdsTxt && !qualityCtx.adsTxtVerified {
		result.Blocked = true
		result.Reason = "ads_txt_not_verified"
		return result
	}

	// Check sellers.json requirement
	if iq.RequireSellerJson && !qualityCtx.sellersJsonVerified {
		result.Blocked = true
		result.Reason = "sellers_json_not_verified"
		return result
	}

	// Brand suitability checks
	if iq.BrandSuitability != nil {
		brandResult := s.evaluateBrandSuitability(req, iq.BrandSuitability)
		if brandResult.blocked {
			result.Blocked = true
			result.BrandSafe = false
			result.Reason = brandResult.reason
			return result
		}
		result.BrandSafe = brandResult.safe
		result.Multiplier *= brandResult.multiplier
	}

	// Fraud protection checks
	if iq.FraudProtection != nil {
		fraudResult := s.evaluateFraudProtection(req, iq.FraudProtection, qualityCtx)
		if fraudResult.blocked {
			result.Blocked = true
			result.Reason = fraudResult.reason
			return result
		}
		result.Multiplier *= fraudResult.multiplier
	}

	// Viewability history checks
	if iq.ViewabilityHistory != nil {
		viewResult := s.evaluateViewabilityHistory(qualityCtx, iq.ViewabilityHistory)
		if viewResult.blocked {
			result.Blocked = true
			result.Reason = viewResult.reason
			return result
		}
		result.Multiplier *= viewResult.multiplier
	}

	// Apply quality tier boosts
	if len(iq.QualityTiers) > 0 {
		tierResult := s.applyQualityTier(qualityCtx.qualityScore, iq.QualityTiers)
		result.QualityTier = tierResult.tier
		result.Multiplier *= tierResult.multiplier
		result.Matched = tierResult.matched
	}

	// Quality score-based boost
	if qualityCtx.qualityScore >= 0.8 {
		result.Multiplier *= 1.2 // Premium inventory
	} else if qualityCtx.qualityScore >= 0.6 {
		result.Multiplier *= 1.05 // Good inventory
	} else if qualityCtx.qualityScore < 0.4 {
		result.Multiplier *= 0.8 // Lower quality
	}

	// Cap multiplier
	if result.Multiplier > 2.0 {
		result.Multiplier = 2.0
	}
	if result.Multiplier < 0.4 {
		result.Multiplier = 0.4
	}

	return result
}

// inventoryQualityContext holds extracted inventory quality signals
type inventoryQualityContext struct {
	qualityScore        float64
	trustLevel          string
	adsTxtVerified      bool
	sellersJsonVerified bool
	viewabilityRate     float64
	fraudRisk           float64
	contentRating       string
	contentCategories   []string
	siteType            string
	botProbability      float64
	proxyDetected       bool
}

// extractInventoryQualityContext extracts quality signals from request
func (s *BiddingService) extractInventoryQualityContext(req *model.BidRequest) inventoryQualityContext {
	ctx := inventoryQualityContext{
		qualityScore:    0.5, // Default mid-quality
		trustLevel:      "unknown",
		viewabilityRate: 0.5,
	}

	if req.Context == nil {
		return ctx
	}

	// Extract from context
	if qs, ok := req.Context["quality_score"].(float64); ok {
		ctx.qualityScore = qs
	}
	if qs, ok := req.Context["inventory_quality"].(float64); ok {
		ctx.qualityScore = qs
	}
	if tl, ok := req.Context["trust_level"].(string); ok {
		ctx.trustLevel = tl
	}
	if tl, ok := req.Context["seller_type"].(string); ok {
		ctx.trustLevel = tl
	}
	if ads, ok := req.Context["ads_txt_verified"].(bool); ok {
		ctx.adsTxtVerified = ads
	}
	if ads, ok := req.Context["ads_txt"].(bool); ok {
		ctx.adsTxtVerified = ads
	}
	if sellers, ok := req.Context["sellers_json_verified"].(bool); ok {
		ctx.sellersJsonVerified = sellers
	}
	if vr, ok := req.Context["viewability_rate"].(float64); ok {
		ctx.viewabilityRate = vr
	}
	if vr, ok := req.Context["historical_viewability"].(float64); ok {
		ctx.viewabilityRate = vr
	}
	if fr, ok := req.Context["fraud_risk"].(float64); ok {
		ctx.fraudRisk = fr
	}
	if fr, ok := req.Context["ivt_score"].(float64); ok {
		ctx.fraudRisk = fr
	}
	if cr, ok := req.Context["content_rating"].(string); ok {
		ctx.contentRating = cr
	}
	if cats, ok := req.Context["content_categories"].([]interface{}); ok {
		for _, cat := range cats {
			if catStr, ok := cat.(string); ok {
				ctx.contentCategories = append(ctx.contentCategories, catStr)
			}
		}
	}
	if st, ok := req.Context["site_type"].(string); ok {
		ctx.siteType = st
	}
	if bp, ok := req.Context["bot_probability"].(float64); ok {
		ctx.botProbability = bp
	}
	if proxy, ok := req.Context["proxy_detected"].(bool); ok {
		ctx.proxyDetected = proxy
	}

	return ctx
}

// brandSuitabilityResult holds brand safety evaluation
type brandSuitabilityResult struct {
	blocked    bool
	safe       bool
	multiplier float64
	reason     string
}

// evaluateBrandSuitability checks brand safety requirements
func (s *BiddingService) evaluateBrandSuitability(req *model.BidRequest, bs *model.BrandSuitability) brandSuitabilityResult {
	result := brandSuitabilityResult{
		blocked:    false,
		safe:       true,
		multiplier: 1.0,
	}

	// Get content info from context
	var contentRating string
	var contentCategories []string
	var pageContent string

	if req.Context != nil {
		if cr, ok := req.Context["content_rating"].(string); ok {
			contentRating = cr
		}
		if cats, ok := req.Context["content_categories"].([]interface{}); ok {
			for _, cat := range cats {
				if catStr, ok := cat.(string); ok {
					contentCategories = append(contentCategories, catStr)
				}
			}
		}
		if pc, ok := req.Context["page_content"].(string); ok {
			pageContent = strings.ToLower(pc)
		}
	}

	// Check content rating floor
	if bs.FloorRating != "" && contentRating != "" {
		if !s.isRatingAllowed(contentRating, bs.FloorRating) {
			result.blocked = true
			result.safe = false
			result.reason = "content_rating_below_floor"
			return result
		}
	}

	// Check blocked categories
	for _, blocked := range bs.BlockedCategories {
		for _, cat := range contentCategories {
			if strings.EqualFold(blocked, cat) {
				result.blocked = true
				result.safe = false
				result.reason = "blocked_category:" + blocked
				return result
			}
		}
	}

	// Check allowed categories whitelist
	if len(bs.AllowedCategories) > 0 && len(contentCategories) > 0 {
		hasAllowed := false
		for _, allowed := range bs.AllowedCategories {
			for _, cat := range contentCategories {
				if strings.EqualFold(allowed, cat) {
					hasAllowed = true
					break
				}
			}
			if hasAllowed {
				break
			}
		}
		if !hasAllowed {
			result.blocked = true
			result.safe = false
			result.reason = "category_not_in_allowlist"
			return result
		}
	}

	// Check custom keyword blocks
	for _, keyword := range bs.CustomKeywordBlock {
		if strings.Contains(pageContent, strings.ToLower(keyword)) {
			result.blocked = true
			result.safe = false
			result.reason = "blocked_keyword:" + keyword
			return result
		}
	}

	// Check sentiment filters
	if req.Context != nil && len(bs.SentimentFilters) > 0 {
		if sentiment, ok := req.Context["sentiment"].(string); ok {
			for _, filter := range bs.SentimentFilters {
				if strings.EqualFold(sentiment, filter) {
					result.blocked = true
					result.safe = false
					result.reason = "sentiment_filtered:" + filter
					return result
				}
			}
		}
	}

	return result
}

// isRatingAllowed checks if content rating meets floor
func (s *BiddingService) isRatingAllowed(contentRating, floorRating string) bool {
	ratings := map[string]int{
		"G":     1,
		"PG":    2,
		"PG13":  3,
		"PG-13": 3,
		"R":     4,
		"NC17":  5,
		"NC-17": 5,
	}

	contentLevel, ok1 := ratings[strings.ToUpper(contentRating)]
	floorLevel, ok2 := ratings[strings.ToUpper(floorRating)]

	if !ok1 || !ok2 {
		return true // Unknown ratings pass
	}

	return contentLevel <= floorLevel
}

// fraudProtectionResult holds fraud evaluation
type fraudProtectionResult struct {
	blocked    bool
	multiplier float64
	reason     string
}

// evaluateFraudProtection checks fraud protection requirements
func (s *BiddingService) evaluateFraudProtection(req *model.BidRequest, fp *model.FraudProtection, ctx inventoryQualityContext) fraudProtectionResult {
	result := fraudProtectionResult{
		blocked:    false,
		multiplier: 1.0,
	}

	// Check minimum trust score
	if fp.MinTrustScore > 0 && (1.0-ctx.fraudRisk) < fp.MinTrustScore {
		result.blocked = true
		result.reason = "trust_score_too_low"
		return result
	}

	// Check bot traffic
	if fp.BlockBotTraffic && ctx.botProbability > 0.7 {
		result.blocked = true
		result.reason = "suspected_bot_traffic"
		return result
	}

	// Check proxy traffic
	if fp.BlockProxyTraffic && ctx.proxyDetected {
		result.blocked = true
		result.reason = "proxy_traffic_detected"
		return result
	}

	// Check blocked sources
	if req.Context != nil {
		if source, ok := req.Context["traffic_source"].(string); ok {
			for _, blocked := range fp.BlockedSources {
				if strings.EqualFold(blocked, source) {
					result.blocked = true
					result.reason = "blocked_traffic_source"
					return result
				}
			}
		}
	}

	// Apply risk-based discount
	if ctx.fraudRisk > 0.3 {
		result.multiplier = 1.0 - ctx.fraudRisk*0.5 // Up to 50% discount for high risk
	}

	return result
}

// viewabilityHistoryResult holds viewability evaluation
type viewabilityHistoryResult struct {
	blocked    bool
	multiplier float64
	reason     string
}

// evaluateViewabilityHistory checks historical viewability requirements
func (s *BiddingService) evaluateViewabilityHistory(ctx inventoryQualityContext, vh *model.ViewabilityHistory) viewabilityHistoryResult {
	result := viewabilityHistoryResult{
		blocked:    false,
		multiplier: 1.0,
	}

	// Check minimum historical viewability
	if vh.MinHistoricalRate > 0 && ctx.viewabilityRate < vh.MinHistoricalRate {
		result.blocked = true
		result.reason = "historical_viewability_too_low"
		return result
	}

	// Apply viewability-based boosts/penalties
	if ctx.viewabilityRate >= 0.7 && vh.HighViewBoost > 0 {
		result.multiplier = vh.HighViewBoost
	} else if ctx.viewabilityRate < 0.4 && vh.LowViewPenalty > 0 {
		result.multiplier = vh.LowViewPenalty
	}

	return result
}

// qualityTierResult holds quality tier evaluation
type qualityTierResult struct {
	tier       string
	multiplier float64
	matched    bool
}

// applyQualityTier determines quality tier and bid adjustment
func (s *BiddingService) applyQualityTier(score float64, tiers []model.QualityTier) qualityTierResult {
	result := qualityTierResult{
		tier:       "standard",
		multiplier: 1.0,
		matched:    false,
	}

	for _, tier := range tiers {
		inRange := true
		if tier.MinScore > 0 && score < tier.MinScore {
			inRange = false
		}
		if tier.MaxScore > 0 && score > tier.MaxScore {
			inRange = false
		}

		if inRange {
			result.tier = tier.Tier
			result.matched = true

			mult := tier.BidMultiplier
			if mult <= 0 {
				mult = 1.0
			}

			// Apply cap
			if tier.MaxBidIncrease > 0 && mult > (1.0+tier.MaxBidIncrease) {
				mult = 1.0 + tier.MaxBidIncrease
			}

			result.multiplier = mult
			break
		}
	}

	return result
}

// calculateDealTargetingMultiplier evaluates PMP/deal targeting
func (s *BiddingService) calculateDealTargetingMultiplier(campaign *model.Campaign, req *model.BidRequest) model.DealTargetingResult {
	result := model.DealTargetingResult{
		Matched:    false,
		Blocked:    false,
		Multiplier: 1.0,
	}

	// Get available deals from request
	var availableDeals []model.Deal
	if req.Pmp != nil && len(req.Pmp.Deals) > 0 {
		availableDeals = req.Pmp.Deals
	}

	// No deal targeting configured
	dt := campaign.Targeting.DealTargeting
	if dt == nil {
		// If campaign has a legacy DealID, check for it
		if campaign.DealID != "" {
			return s.evaluateLegacyDeal(campaign, availableDeals)
		}
		return result
	}

	// Check if deals are required but none available
	if dt.RequireDeal && len(availableDeals) == 0 {
		if dt.FallbackToOpen {
			// Allow open auction fallback
			result.DealType = "open"
			return result
		}
		result.Blocked = true
		result.Reason = "deal_required_but_none_available"
		return result
	}

	// If no deals available and not requiring deals
	if len(availableDeals) == 0 {
		result.DealType = "open"
		return result
	}

	// Check excluded deals
	for _, deal := range availableDeals {
		for _, excluded := range dt.ExcludedDealIDs {
			if deal.ID == excluded {
				// Skip this deal
				continue
			}
		}
	}

	// Find best matching deal
	bestDeal := s.findBestDeal(campaign, availableDeals, dt)
	if bestDeal != nil {
		result.Matched = true
		result.MatchedDealID = bestDeal.ID
		result.EffectiveFloor = bestDeal.BidFloor
		result.DealPriority = s.getDealPriority(bestDeal, dt)

		// Determine deal type
		result.DealType = s.classifyDealType(bestDeal)
		result.IsPG = result.DealType == "programmatic_guaranteed"

		// Apply deal-specific bid adjustments
		for _, adj := range dt.DealBidAdjustments {
			if adj.DealID == bestDeal.ID {
				mult := adj.BidMultiplier
				if mult <= 0 {
					mult = 1.0
				}
				result.Multiplier *= mult
				break
			}
		}

		// PG boost if preferred
		if dt.PreferPG && result.IsPG {
			result.Multiplier *= 1.3
		}

		// Preferred deal boost
		for _, preferred := range dt.PreferredDealIDs {
			if bestDeal.ID == preferred {
				result.Multiplier *= 1.2
				break
			}
		}
	} else if dt.RequireDeal {
		if dt.FallbackToOpen {
			result.DealType = "open"
		} else {
			result.Blocked = true
			result.Reason = "no_matching_deal_found"
			return result
		}
	}

	// Check publisher deals
	publisherID := s.getPublisherID(req)
	for _, pubDeal := range dt.PublisherDeals {
		if pubDeal.PublisherID == publisherID {
			// Apply publisher-specific boost
			if pubDeal.BidBoost > 0 {
				result.Multiplier *= pubDeal.BidBoost
			}

			// Check exclusivity
			if pubDeal.Exclusive && result.MatchedDealID != "" {
				dealMatched := false
				for _, dealID := range pubDeal.DealIDs {
					if dealID == result.MatchedDealID {
						dealMatched = true
						break
					}
				}
				if !dealMatched {
					result.Blocked = true
					result.Reason = "publisher_exclusive_deal_not_matched"
					return result
				}
			}
			break
		}
	}

	// Cap multiplier
	if result.Multiplier > 2.0 {
		result.Multiplier = 2.0
	}
	if result.Multiplier < 0.5 {
		result.Multiplier = 0.5
	}

	return result
}

// evaluateLegacyDeal handles legacy DealID on campaign
func (s *BiddingService) evaluateLegacyDeal(campaign *model.Campaign, deals []model.Deal) model.DealTargetingResult {
	result := model.DealTargetingResult{
		Matched:    false,
		Blocked:    false,
		Multiplier: 1.0,
	}

	if campaign.DealID == "" {
		return result
	}

	// Look for matching deal
	for _, deal := range deals {
		if deal.ID == campaign.DealID {
			// Check floor price
			if campaign.BidPrice >= deal.BidFloor {
				result.Matched = true
				result.MatchedDealID = deal.ID
				result.EffectiveFloor = deal.BidFloor
				result.DealType = s.classifyDealType(&deal)
				result.Multiplier = 1.2 // Boost for deal match
			}
			break
		}
	}

	return result
}

// findBestDeal finds the best matching deal from available deals
func (s *BiddingService) findBestDeal(campaign *model.Campaign, deals []model.Deal, dt *model.DealTargeting) *model.Deal {
	var bestDeal *model.Deal
	bestPriority := -1

	for i := range deals {
		deal := &deals[i]

		// Check if deal is excluded
		excluded := false
		for _, excl := range dt.ExcludedDealIDs {
			if deal.ID == excl {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Check if bid price meets floor
		if campaign.BidPrice < deal.BidFloor {
			continue
		}

		// Check deal type filter
		if len(dt.DealTypes) > 0 {
			dealType := s.classifyDealType(deal)
			typeMatched := false
			for _, allowedType := range dt.DealTypes {
				if strings.EqualFold(dealType, allowedType) {
					typeMatched = true
					break
				}
			}
			if !typeMatched {
				continue
			}
		}

		// Calculate priority
		priority := s.getDealPriority(deal, dt)

		// Check minimum priority
		if dt.MinDealPriority > 0 && priority < dt.MinDealPriority {
			continue
		}

		// Prefer PG deals if configured
		if dt.PreferPG {
			dealType := s.classifyDealType(deal)
			if dealType == "programmatic_guaranteed" {
				priority += 100
			}
		}

		// Prefer preferred deals
		for _, preferred := range dt.PreferredDealIDs {
			if deal.ID == preferred {
				priority += 50
				break
			}
		}

		// Select best deal
		if bestDeal == nil || priority > bestPriority {
			bestDeal = deal
			bestPriority = priority
		}
	}

	return bestDeal
}

// classifyDealType determines the type of deal
func (s *BiddingService) classifyDealType(deal *model.Deal) string {
	if deal == nil {
		return "open"
	}

	// Check At (Auction Type) field if available
	switch deal.At {
	case 1:
		return "first_price"
	case 2:
		return "second_price"
	case 3:
		return "programmatic_guaranteed"
	}

	// Infer from other signals
	// PG deals typically have fixed price and guaranteed delivery
	if deal.BidFloor > 0 && deal.WSeat != nil && len(deal.WSeat) == 1 {
		return "programmatic_guaranteed"
	}

	// Private auction typically has multiple allowed seats
	if len(deal.WSeat) > 1 {
		return "private_auction"
	}

	// Preferred deals have whitelist but are auction-based
	if len(deal.WSeat) > 0 {
		return "preferred"
	}

	return "private_auction" // Default for any deal
}

// getDealPriority calculates deal priority
func (s *BiddingService) getDealPriority(deal *model.Deal, dt *model.DealTargeting) int {
	if deal == nil {
		return 0
	}

	// Check for override in adjustments
	for _, adj := range dt.DealBidAdjustments {
		if adj.DealID == deal.ID && adj.Priority > 0 {
			return adj.Priority
		}
	}

	// Calculate based on deal attributes
	priority := 5 // Base priority

	// PG deals get higher priority
	dealType := s.classifyDealType(deal)
	switch dealType {
	case "programmatic_guaranteed":
		priority += 4
	case "preferred":
		priority += 2
	case "private_auction":
		priority += 1
	}

	return priority
}

// getPublisherID extracts publisher ID from request
func (s *BiddingService) getPublisherID(req *model.BidRequest) string {
	// Check context for publisher ID
	if req.Context != nil {
		if pubID, ok := req.Context["publisher_id"].(string); ok {
			return pubID
		}
		if pubID, ok := req.Context["pub_id"].(string); ok {
			return pubID
		}
		if pubID, ok := req.Context["site_publisher_id"].(string); ok {
			return pubID
		}
		if pubID, ok := req.Context["app_publisher_id"].(string); ok {
			return pubID
		}
	}

	return ""
}

// checkRetargetingEligibility checks if a user matches retargeting criteria for a campaign
// Returns true if user has any of the specified engagement events for the target campaigns
func (s *BiddingService) checkRetargetingEligibility(campaign *model.Campaign, userID string) bool {
	if userID == "" {
		return false
	}

	eventTypes := campaign.Targeting.RetargetingEvents
	if len(eventTypes) == 0 {
		eventTypes = []string{"impression", "click"} // Default: target users who viewed or clicked
	}

	targetCampaigns := campaign.Targeting.RetargetingCampaigns
	if len(targetCampaigns) == 0 {
		targetCampaigns = []string{campaign.ID} // Default: target users from this campaign
	}

	// Check if user has any of the required events for any of the target campaigns
	for _, eventType := range eventTypes {
		for _, targetCampaignID := range targetCampaigns {
			hasEvent, err := s.cache.HasUserEvent(userID, targetCampaignID, eventType)
			if err == nil && hasEvent {
				return true // User matches retargeting criteria
			}
		}
	}

	return false
}

// calculatePacingMultiplier returns a bid score multiplier (0.0-1.0) based on budget pacing.
// Strategies:
//   - "asap": No pacing; spend as fast as possible (multiplier = 1.0)
//   - "even": Spread evenly across the day; reduce bids if ahead of schedule
//   - "front": Spend more in early hours, taper off later
//   - "back": Conserve early, accelerate spend in later hours
func (s *BiddingService) calculatePacingMultiplier(strategy string, spent, dailyBudget float64) float64 {
	if strategy == "" {
		strategy = "even"
	}

	if strategy == "asap" {
		return 1.0
	}

	// Calculate time-based expected spend
	now := time.Now()
	hourOfDay := now.Hour()
	minuteOfHour := now.Minute()
	elapsedMinutes := float64(hourOfDay*60 + minuteOfHour)
	totalMinutes := 24.0 * 60.0
	dayProgress := elapsedMinutes / totalMinutes // 0.0 at midnight, 1.0 at 23:59

	var expectedSpendRatio float64

	switch strategy {
	case "front":
		// Front-loaded: expect to spend 70% in first half of day
		if dayProgress <= 0.5 {
			expectedSpendRatio = dayProgress * 1.4 // steeper curve early
		} else {
			expectedSpendRatio = 0.7 + (dayProgress-0.5)*0.6 // flatten later
		}
	case "back":
		// Back-loaded: spend 30% in first half, 70% in second half
		if dayProgress <= 0.5 {
			expectedSpendRatio = dayProgress * 0.6 // slower early
		} else {
			expectedSpendRatio = 0.3 + (dayProgress-0.5)*1.4 // accelerate later
		}
	default: // "even"
		expectedSpendRatio = dayProgress
	}

	expectedSpend := dailyBudget * expectedSpendRatio

	// If we're ahead of schedule, reduce bid aggressiveness
	// If we're behind schedule, increase bid aggressiveness (up to 1.2x)
	if expectedSpend <= 0 {
		return 1.0
	}

	pacingRatio := spent / expectedSpend

	if pacingRatio > 1.2 {
		// Way ahead of pace: significantly reduce bids
		return 0.3
	} else if pacingRatio > 1.0 {
		// Slightly ahead: moderate reduction
		return 0.7
	} else if pacingRatio < 0.5 {
		// Way behind: boost bids (cap at 1.2 to avoid overspend)
		return 1.2
	} else if pacingRatio < 0.8 {
		// Slightly behind: slight boost
		return 1.1
	}

	return 1.0 // On pace
}

// calculateGoalPacingMultiplier adjusts bids based on progress toward delivery goals.
// Goals can be impressions, clicks, or conversions with target dates.
// If behind schedule, increases multiplier; if ahead, decreases it.
func (s *BiddingService) calculateGoalPacingMultiplier(campaign *model.Campaign) float64 {
	if campaign.GoalTarget <= 0 || campaign.GoalEndDate == "" {
		return 1.0
	}

	// Parse goal end date
	endDate, err := time.Parse("2006-01-02", campaign.GoalEndDate)
	if err != nil {
		return 1.0 // Default if date is invalid
	}

	today := time.Now().Truncate(24 * time.Hour)
	endDateTrunc := endDate.Truncate(24 * time.Hour)

	// If deadline has passed, dial down bids
	if today.After(endDateTrunc) {
		return 0.5 // Campaign is past deadline
	}

	// Calculate days elapsed and remaining
	startDate := today // Assume campaign started today for simplicity
	totalDays := int(endDateTrunc.Sub(startDate.Truncate(24*time.Hour)).Hours() / 24)
	if totalDays <= 0 {
		totalDays = 1
	}
	daysRemaining := int(endDateTrunc.Sub(today).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	// Calculate expected daily delivery
	remainingGoal := campaign.GoalTarget - campaign.GoalDelivered
	if remainingGoal <= 0 {
		return 0.3 // Goal already met, reduce bids
	}

	requiredDaily := float64(remainingGoal) / float64(daysRemaining+1) // +1 to avoid division by zero
	currentDaily := campaign.GoalDelivered / int64(time.Since(today.Add(-1*24*time.Hour)).Hours()/24+1)

	// Compare actual daily delivery to required daily delivery
	if currentDaily == 0 {
		currentDaily = 1
	}
	deliveryRatio := float64(currentDaily) / requiredDaily

	// Adjust bids based on delivery pace
	if deliveryRatio > 1.5 {
		// Way ahead of goal pace: reduce bids
		return 0.5
	} else if deliveryRatio > 1.1 {
		// Slightly ahead: moderate reduction
		return 0.8
	} else if deliveryRatio < 0.5 {
		// Way behind goal pace: boost bids
		return 1.5
	} else if deliveryRatio < 0.8 {
		// Slightly behind: boost bids
		return 1.2
	}

	return 1.0 // On track
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
		"total_bids":     bidCount,
		"total_wins":     winCount,
		"win_rate":       winRate,
		"avg_latency_ms": avgLatency,
		"timestamp":      time.Now(),
	}, nil
}

// RefreshCampaigns fetches fresh campaigns from backend API
func (s *BiddingService) RefreshCampaigns(backendURL string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("%s/internal/campaigns/active", backendURL))

	// Fallback for local testing if backend is down
	if (err != nil || (resp != nil && resp.StatusCode != http.StatusOK)) && os.Getenv("ENV") == "development" {
		fmt.Printf("Warning: Backend unreachable or error (%v). Using dummy campaigns for development.\n", err)

		dummyCampaigns := []*model.Campaign{
			{
				ID:       "camp-native-1",
				Name:     "Test Native Campaign",
				Type:     "cpm",
				BidPrice: 2.50,
				Status:   "active",
				Budget:   1000.0, // High Budget
				Spent:    0.0,
				Creative: model.Creative{
					Type:        "native",
					Title:       "Native Ad Title",
					Description: "This is a native ad description",
					IconURL:     "https://example.com/icon.png",
					URL:         "https://example.com/image.jpg",
					CTAText:     "Install Now",
					Width:       1200,
					Height:      627,
				},
				Targeting: model.Targeting{
					Countries: []string{"US"},
				},
			},
			{
				ID:       "camp-banner-1",
				Name:     "Test Banner Campaign",
				Type:     "cpm",
				BidPrice: 1.50,
				Status:   "active",
				Budget:   1000.0,
				Spent:    0.0,
				Creative: model.Creative{
					Type:   "banner",
					Width:  300,
					Height: 250,
					URL:    "https://example.com/banner.jpg",
				},
				Targeting: model.Targeting{
					Countries: []string{"US"},
				},
			},
			{
				ID:       "camp-video-1",
				Name:     "Test Video Campaign",
				Type:     "cpm",
				BidPrice: 5.00,
				Status:   "active",
				Budget:   1000.0,
				Spent:    0.0,
				Creative: model.Creative{
					Type:     "video",
					Duration: 30,
					MimeType: "video/mp4",
					URL:      "https://example.com/video.mp4",
					Width:    640,
					Height:   480,
				},
				Targeting: model.Targeting{
					Countries: []string{"US"},
				},
			},
			{
				ID:       "camp-audio-1",
				Name:     "Test Audio Campaign",
				Type:     "cpm",
				BidPrice: 3.50,
				Status:   "active",
				Budget:   1000.0,
				Spent:    0.0,
				Creative: model.Creative{
					Type:        "audio",
					Title:       "Audio Ad",
					Description: "This is an audio ad",
					Duration:    15,
					MimeType:    "audio/mp3",
					URL:         "https://example.com/audio.mp3",
					Width:       0,
					Height:      0,
					Bitrate:     128,
				},
				Targeting: model.Targeting{
					Countries: []string{"US"},
				},
			},
		}
		return s.cache.SetActiveCampaigns(dummyCampaigns)
	}

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

	return s.cache.SetActiveCampaigns(campaigns)
}

// callFraudService calls the Fraud Detection Service
func (s *BiddingService) callFraudService(req *model.BidRequest) (bool, error, *model.SupplyPathHop) {
	hopStart := time.Now()

	// 1. Check Redis Cache for IP Reputation
	// Key: "ip_rep:{ip}" -> "block" or "allow"
	// TTL: 1 hour
	cacheKey := fmt.Sprintf("ip_rep:%s", req.Device.IP)
	action, err := s.cache.Get(cacheKey)
	if err == nil {
		if action == "block" {
			hop := &model.SupplyPathHop{
				ServiceName: "fraud-detection-cache",
				ServiceType: "cache",
				Endpoint:    cacheKey,
				LatencyMs:   time.Since(hopStart).Milliseconds(),
				StatusCode:  200,
				Success:     true,
				Fee:         0.0, // Cache hits are free
				Timestamp:   time.Now(),
			}
			return true, nil, hop // Fraud (Blocked)
		}
		if action == "allow" {
			hop := &model.SupplyPathHop{
				ServiceName: "fraud-detection-cache",
				ServiceType: "cache",
				Endpoint:    cacheKey,
				LatencyMs:   time.Since(hopStart).Milliseconds(),
				StatusCode:  200,
				Success:     true,
				Fee:         0.0, // Cache hits are free
				Timestamp:   time.Now(),
			}
			return false, nil, hop // Safe (Allowed)
		}
	}

	// 2. Fallback to API Call if not in cache
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
		hop := &model.SupplyPathHop{
			ServiceName:  "fraud-detection",
			ServiceType:  "internal",
			Endpoint:     s.fraudServiceURL + "/detect",
			RequestSize:  len(jsonData),
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   0,
			Success:      false,
			ErrorMessage: err.Error(),
			Fee:          0.001, // Small fee for API calls
			Timestamp:    time.Now(),
		}
		return false, err, hop
	}

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Post(s.fraudServiceURL+"/detect", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		hop := &model.SupplyPathHop{
			ServiceName:  "fraud-detection",
			ServiceType:  "internal",
			Endpoint:     s.fraudServiceURL + "/detect",
			RequestSize:  len(jsonData),
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   0,
			Success:      false,
			ErrorMessage: err.Error(),
			Fee:          0.001,
			Timestamp:    time.Now(),
		}
		return false, err, hop
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		hop := &model.SupplyPathHop{
			ServiceName:  "fraud-detection",
			ServiceType:  "internal",
			Endpoint:     s.fraudServiceURL + "/detect",
			RequestSize:  len(jsonData),
			ResponseSize: 0,
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   resp.StatusCode,
			Success:      false,
			ErrorMessage: fmt.Sprintf("HTTP %d", resp.StatusCode),
			Fee:          0.001,
			Timestamp:    time.Now(),
		}
		return false, fmt.Errorf("fraud service status: %d", resp.StatusCode), hop
	}

	var fraudResp model.FraudCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&fraudResp); err != nil {
		hop := &model.SupplyPathHop{
			ServiceName:  "fraud-detection",
			ServiceType:  "internal",
			Endpoint:     s.fraudServiceURL + "/detect",
			RequestSize:  len(jsonData),
			ResponseSize: 0,
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   resp.StatusCode,
			Success:      false,
			ErrorMessage: err.Error(),
			Fee:          0.001,
			Timestamp:    time.Now(),
		}
		return false, err, hop
	}

	// 3. Cache the Result
	result := "allow"
	isFraud := false
	if fraudResp.RecommendedAction == "block" || fraudResp.IsFraud {
		result = "block"
		isFraud = true
	}

	// Set with 1 hour TTL
	// We ignore cache set errors as it's non-critical
	_ = s.cache.Set(cacheKey, result, 3600)

	hop := &model.SupplyPathHop{
		ServiceName:  "fraud-detection",
		ServiceType:  "internal",
		Endpoint:     s.fraudServiceURL + "/detect",
		RequestSize:  len(jsonData),
		ResponseSize: 0, // Approximate
		LatencyMs:    time.Since(hopStart).Milliseconds(),
		StatusCode:   resp.StatusCode,
		Success:      true,
		Fee:          0.001,
		Timestamp:    time.Now(),
	}

	return isFraud, nil, hop
}

// callOptimizationService calls the Bid Optimization Service
func (s *BiddingService) callOptimizationService(bid *model.BidResult, req *model.BidRequest) (*model.BidRecommendation, error, *model.SupplyPathHop) {
	hopStart := time.Now()

	// Circuit Breaker Check
	s.optMutex.RLock()
	isCircuitOpen := s.optFailureCount > 5 && time.Since(s.optLastFailure) < 30*time.Second
	s.optMutex.RUnlock()

	if isCircuitOpen {
		hop := &model.SupplyPathHop{
			ServiceName:  "bid-optimizer",
			ServiceType:  "internal",
			Endpoint:     s.optServiceURL + "/optimize",
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   0,
			Success:      false,
			ErrorMessage: "circuit breaker active",
			Fee:          0.0015, // Optimization service fee
			Timestamp:    time.Now(),
		}
		return nil, fmt.Errorf("circuit breaker active: optimization service temporarily unavailable"), hop
	}

	// Fetch real performance metrics from Redis
	realCTR, _ := s.cache.GetCampaignCTR(bid.Campaign.ID)
	realWinRate, _ := s.cache.GetCampaignWinRate(bid.Campaign.ID)
	if realCTR == 0 {
		realCTR = 0.02 // Fallback default
	}
	if realWinRate == 0 {
		realWinRate = 0.5 // Fallback default
	}

	// Calculate real daily budget and pacing ratio
	dailyBudget := bid.Campaign.DailyBudget
	if dailyBudget <= 0 {
		dailyBudget = bid.Campaign.Budget / 30.0
	}
	dailySpend, _ := s.cache.GetCampaignSpend(bid.Campaign.ID)
	pacingRatio := 1.0
	if dailyBudget > 0 {
		pacingRatio = dailySpend / dailyBudget
	}

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
				WinRate:    realWinRate,
				CTR:        realCTR,
			},
			Budget: model.BudgetStatus{
				CampaignID:  bid.Campaign.ID,
				DailyBudget: dailyBudget,
				PacingRatio: pacingRatio,
			},
		},
	}

	jsonData, err := json.Marshal(optReq)
	if err != nil {
		hop := &model.SupplyPathHop{
			ServiceName:  "bid-optimizer",
			ServiceType:  "internal",
			Endpoint:     s.optServiceURL + "/optimize",
			RequestSize:  len(jsonData),
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   0,
			Success:      false,
			ErrorMessage: err.Error(),
			Fee:          0.0015,
			Timestamp:    time.Now(),
		}
		return nil, err, hop
	}

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Post(s.optServiceURL+"/optimize", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// Circuit Breaker: Record failure
		s.optMutex.Lock()
		s.optFailureCount++
		s.optLastFailure = time.Now()
		s.optMutex.Unlock()

		hop := &model.SupplyPathHop{
			ServiceName:  "bid-optimizer",
			ServiceType:  "internal",
			Endpoint:     s.optServiceURL + "/optimize",
			RequestSize:  len(jsonData),
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   0,
			Success:      false,
			ErrorMessage: err.Error(),
			Fee:          0.0015,
			Timestamp:    time.Now(),
		}
		return nil, err, hop
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		hop := &model.SupplyPathHop{
			ServiceName:  "bid-optimizer",
			ServiceType:  "internal",
			Endpoint:     s.optServiceURL + "/optimize",
			RequestSize:  len(jsonData),
			ResponseSize: 0,
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   resp.StatusCode,
			Success:      false,
			ErrorMessage: fmt.Sprintf("HTTP %d", resp.StatusCode),
			Fee:          0.0015,
			Timestamp:    time.Now(),
		}
		return nil, fmt.Errorf("optimization service status: %d", resp.StatusCode), hop
	}

	var optResp model.BidRecommendation
	if err := json.NewDecoder(resp.Body).Decode(&optResp); err != nil {
		hop := &model.SupplyPathHop{
			ServiceName:  "bid-optimizer",
			ServiceType:  "internal",
			Endpoint:     s.optServiceURL + "/optimize",
			RequestSize:  len(jsonData),
			ResponseSize: 0,
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   resp.StatusCode,
			Success:      false,
			ErrorMessage: err.Error(),
			Fee:          0.0015,
			Timestamp:    time.Now(),
		}
		return nil, err, hop
	}

	// Circuit Breaker: Reset on success
	s.optMutex.Lock()
	s.optFailureCount = 0
	s.optLastFailure = time.Time{}
	s.optMutex.Unlock()

	hop := &model.SupplyPathHop{
		ServiceName:  "bid-optimizer",
		ServiceType:  "internal",
		Endpoint:     s.optServiceURL + "/optimize",
		RequestSize:  len(jsonData),
		ResponseSize: 0, // Approximate
		LatencyMs:    time.Since(hopStart).Milliseconds(),
		StatusCode:   resp.StatusCode,
		Success:      true,
		Fee:          0.0015,
		Timestamp:    time.Now(),
	}

	return &optResp, nil, hop
}

// callAIMatchingService calls the external Python AI service
func (s *BiddingService) callAIMatchingService(req *model.BidRequest) ([]model.AIAdRecommendation, error, *model.SupplyPathHop) {
	hopStart := time.Now()

	// Circuit Breaker Logic
	s.aiMutex.RLock()
	failures := s.aiFailureCount
	lastFailure := s.aiLastFailure
	s.aiMutex.RUnlock()

	// Check if we should trip the circuit
	if failures >= 3 && time.Since(lastFailure) < 10*time.Minute {
		hop := &model.SupplyPathHop{
			ServiceName:  "ad-matching",
			ServiceType:  "internal",
			Endpoint:     s.aiServiceURL + "/match",
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   0,
			Success:      false,
			ErrorMessage: "circuit breaker active",
			Fee:          0.002, // AI service fee
			Timestamp:    time.Now(),
		}
		return nil, fmt.Errorf("circuit breaker active: AI service temporarily unavailable"), hop
	}

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
			SlotID:     "slot_default",  // In real RTB, would come from req.Imp
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
		hop := &model.SupplyPathHop{
			ServiceName:  "ad-matching",
			ServiceType:  "internal",
			Endpoint:     s.aiServiceURL + "/match",
			RequestSize:  len(jsonData),
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   0,
			Success:      false,
			ErrorMessage: err.Error(),
			Fee:          0.002,
			Timestamp:    time.Now(),
		}
		return nil, err, hop
	}

	// 2. Execute Request with specific timeout
	// AI Service expected at /match endpoint
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Post(s.aiServiceURL+"/match", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// Circuit Breaker: Record failure
		s.aiMutex.Lock()
		s.aiFailureCount++
		s.aiLastFailure = time.Now()
		s.aiMutex.Unlock()

		hop := &model.SupplyPathHop{
			ServiceName:  "ad-matching",
			ServiceType:  "internal",
			Endpoint:     s.aiServiceURL + "/match",
			RequestSize:  len(jsonData),
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   0,
			Success:      false,
			ErrorMessage: err.Error(),
			Fee:          0.002,
			Timestamp:    time.Now(),
		}
		return nil, err, hop
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		hop := &model.SupplyPathHop{
			ServiceName:  "ad-matching",
			ServiceType:  "internal",
			Endpoint:     s.aiServiceURL + "/match",
			RequestSize:  len(jsonData),
			ResponseSize: 0,
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   resp.StatusCode,
			Success:      false,
			ErrorMessage: fmt.Sprintf("HTTP %d", resp.StatusCode),
			Fee:          0.002,
			Timestamp:    time.Now(),
		}
		return nil, fmt.Errorf("AI service returned status: %d", resp.StatusCode), hop
	}

	// 3. Decode Response
	var aiResp model.AIMatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		hop := &model.SupplyPathHop{
			ServiceName:  "ad-matching",
			ServiceType:  "internal",
			Endpoint:     s.aiServiceURL + "/match",
			RequestSize:  len(jsonData),
			ResponseSize: 0,
			LatencyMs:    time.Since(hopStart).Milliseconds(),
			StatusCode:   resp.StatusCode,
			Success:      false,
			ErrorMessage: err.Error(),
			Fee:          0.002,
			Timestamp:    time.Now(),
		}
		return nil, err, hop
	}

	// Circuit Breaker: Reset on success
	s.aiMutex.Lock()
	s.aiFailureCount = 0
	s.aiLastFailure = time.Time{} // Reset
	s.aiMutex.Unlock()

	hop := &model.SupplyPathHop{
		ServiceName:  "ad-matching",
		ServiceType:  "internal",
		Endpoint:     s.aiServiceURL + "/match",
		RequestSize:  len(jsonData),
		ResponseSize: 0, // Approximate
		LatencyMs:    time.Since(hopStart).Milliseconds(),
		StatusCode:   resp.StatusCode,
		Success:      true,
		Fee:          0.002,
		Timestamp:    time.Now(),
	}

	return aiResp.Recommendations, nil, hop
}

// IncrementFormatStats increments the Redis counter for a specific format
func (s *BiddingService) IncrementFormatStats(format string) {
	// Fire and forget - don't block main thread
	go func() {
		_ = s.cache.IncrementBidFormat(format)
	}()
}

// TrackClick records a click event for a campaign in Redis for CTR tracking.
// It also increments the Prometheus click event counter.
func (s *BiddingService) TrackClick(campaignID string) error {
	return s.cache.IncrementCampaignClicks(campaignID)
}

// TrackImpression records an impression event for a campaign in Redis.
// It also increments the Prometheus impression event counter.
func (s *BiddingService) TrackImpression(campaignID string) error {
	return s.cache.IncrementCampaignImpressions(campaignID)
}

// GetBidLandscape returns bid and win statistics across price buckets
func (s *BiddingService) GetBidLandscape() (map[string]map[string]int64, error) {
	return s.cache.GetBidLandscape()
}

// GetSegmentPerformance returns performance metrics for a specific segment type (device, os, geo)
func (s *BiddingService) GetSegmentPerformance(segmentType string) (map[string]map[string]int64, error) {
	return s.cache.GetSegmentPerformance(segmentType)
}

// GetOptimalBidFloor calculates the optimal bid floor for a publisher
func (s *BiddingService) GetOptimalBidFloor(publisherID string, targetWinRate float64) (float64, error) {
	return s.cache.GetOptimalBidFloor(publisherID, targetWinRate)
}

// RecordImpression records an impression for conversion attribution (view-through)
// Default TTL is 24 hours for view-through attribution window
func (s *BiddingService) RecordImpression(userID, campaignID, requestID string) error {
	return s.cache.RecordImpression(userID, campaignID, requestID, 24) // 24 hour VTA window
}

// RecordClick records a click for conversion attribution (click-through)
// Default TTL is 168 hours (7 days) for click-through attribution window
func (s *BiddingService) RecordClick(userID, campaignID, requestID string) error {
	return s.cache.RecordClick(userID, campaignID, requestID, 168) // 7 day CTA window
}

// GetAttribution determines if a conversion should be attributed via CTA, VTA, or none
func (s *BiddingService) GetAttribution(userID, campaignID string) (string, string, error) {
	return s.cache.GetAttribution(userID, campaignID)
}

// RecordUserEvent stores a user engagement event for retargeting
// Default TTL is 30 days for retargeting window
func (s *BiddingService) RecordUserEvent(userID, campaignID, eventType string) error {
	return s.cache.RecordUserEvent(userID, campaignID, eventType, 30) // 30 day retargeting window
}

// GetUserEvents returns all campaigns a user has engaged with by event type
func (s *BiddingService) GetUserEvents(userID string, eventTypes []string) (map[string][]string, error) {
	return s.cache.GetUserEvents(userID, eventTypes)
}

// RecordTouchpoint stores a touchpoint in a user's conversion journey for multi-touch attribution
func (s *BiddingService) RecordTouchpoint(userID, campaignID, touchpointType, requestID string) error {
	return s.cache.RecordTouchpoint(userID, campaignID, touchpointType, requestID, 30) // 30 day lookback
}

// GetMultiTouchAttribution calculates attribution credit using the specified model
// Models: "linear", "time_decay", "position_based", "last_touch", "first_touch"
func (s *BiddingService) GetMultiTouchAttribution(userID, campaignID, modelType string) ([]model.AttributionCredit, error) {
	return s.cache.GetMultiTouchAttribution(userID, campaignID, modelType)
}

// GetAutoBidRecommendations returns bid optimization recommendations for active campaigns
func (s *BiddingService) GetAutoBidRecommendations() ([]map[string]interface{}, error) {
	campaigns, err := s.cache.GetActiveCampaigns()
	if err != nil {
		return nil, err
	}

	recommendations := []map[string]interface{}{}

	for _, campaign := range campaigns {
		ctr, _ := s.cache.GetCampaignCTR(campaign.ID)
		winRate, _ := s.cache.GetCampaignWinRate(campaign.ID)

		multiplier := s.calculateAutoBidMultiplier(campaign)

		if multiplier == 1.0 {
			continue // Skip campaigns with no recommendation
		}

		recommendedBid := campaign.BidPrice * multiplier
		action := "maintain"
		reason := ""

		if multiplier > 1.0 {
			action = "increase"
			reason = fmt.Sprintf("High CTR (%.2f%%) but low win rate (%.1f%%) - underbidding", ctr*100, winRate*100)
		} else if multiplier < 1.0 {
			action = "decrease"
			reason = fmt.Sprintf("Low CTR (%.2f%%) but high win rate (%.1f%%) - overbidding", ctr*100, winRate*100)
		}

		recommendations = append(recommendations, map[string]interface{}{
			"campaignId":     campaign.ID,
			"currentBid":     campaign.BidPrice,
			"recommendedBid": recommendedBid,
			"multiplier":     multiplier,
			"action":         action,
			"reason":         reason,
			"metrics": map[string]interface{}{
				"ctr":     ctr,
				"winRate": winRate,
			},
		})
	}

	return recommendations, nil
}

// getPriceBucket returns the price bucket string for bid landscape analytics
// Buckets: 0.00-0.50, 0.50-1.00, 1.00-2.00, 2.00-5.00, 5.00-10.00, 10.00+
func getPriceBucket(price float64) string {
	switch {
	case price < 0.50:
		return "0.00-0.50"
	case price < 1.00:
		return "0.50-1.00"
	case price < 2.00:
		return "1.00-2.00"
	case price < 5.00:
		return "2.00-5.00"
	case price < 10.00:
		return "5.00-10.00"
	default:
		return "10.00+"
	}
}

// CrossDeviceResult represents the result of cross-device resolution
type CrossDeviceResult struct {
	PrimaryUserID   string   // Unified user ID across devices
	LinkedDevices   []string // All device IDs linked to this user
	DeviceCount     int      // Total number of devices
	IsNewDevice     bool     // Whether this device was just linked
	UnifiedFreq     int64    // Aggregated frequency across all devices
	FreqCapExceeded bool     // Whether unified frequency exceeds cap
}

// ResolveCrossDeviceUser resolves a device ID to a unified primary user ID
// This enables frequency capping and targeting across mobile, desktop, CTV, etc.
func (s *BiddingService) ResolveCrossDeviceUser(deviceID string, additionalSignals map[string]string) *CrossDeviceResult {
	result := &CrossDeviceResult{
		PrimaryUserID: deviceID, // Default to device ID if no graph exists
		LinkedDevices: []string{deviceID},
		DeviceCount:   1,
		IsNewDevice:   false,
	}

	if deviceID == "" {
		return result
	}

	// Try to get existing primary user ID
	primaryID, err := s.cache.GetPrimaryUserID(deviceID)
	if err == nil && primaryID != "" {
		result.PrimaryUserID = primaryID

		// Get all linked devices
		linkedDevices, err := s.cache.GetLinkedDevices(primaryID)
		if err == nil && len(linkedDevices) > 0 {
			result.LinkedDevices = linkedDevices
			result.DeviceCount = len(linkedDevices)
		}
	} else {
		// Device not in graph yet - check for deterministic linking signals
		newPrimaryID := s.findDeterministicLink(deviceID, additionalSignals)
		if newPrimaryID != "" {
			// Link this device to existing graph (90 days default TTL)
			_ = s.cache.LinkDevices(newPrimaryID, []string{deviceID}, 90)
			result.PrimaryUserID = newPrimaryID
			result.IsNewDevice = true

			// Refresh linked devices
			linkedDevices, _ := s.cache.GetLinkedDevices(newPrimaryID)
			result.LinkedDevices = linkedDevices
			result.DeviceCount = len(linkedDevices)
		}
	}

	return result
}

// findDeterministicLink attempts to find deterministic signals to link devices
// Uses email hash, phone hash, login ID, or other authenticated identifiers
func (s *BiddingService) findDeterministicLink(deviceID string, signals map[string]string) string {
	if signals == nil {
		return ""
	}

	// Priority order for deterministic matching
	deterministicKeys := []string{
		"email_hash", // Hashed email (SHA256)
		"phone_hash", // Hashed phone number
		"login_id",   // Authenticated login ID
		"hh_id",      // Household ID (CTV)
		"uid2",       // Unified ID 2.0
		"rampid",     // LiveRamp RampID
	}

	for _, key := range deterministicKeys {
		if value, ok := signals[key]; ok && value != "" {
			// Look up existing graph by this identifier
			existingPrimary, err := s.cache.GetPrimaryUserID(key + ":" + value)
			if err == nil && existingPrimary != "" {
				return existingPrimary
			}
		}
	}

	return ""
}

// GetCrossDeviceFrequency returns aggregated impression count across all linked devices
func (s *BiddingService) GetCrossDeviceFrequency(primaryUserID, campaignID string) int64 {
	freq, err := s.cache.GetCrossDeviceFrequency(primaryUserID, campaignID)
	if err != nil {
		return 0
	}
	return freq
}

// checkCrossDeviceFreqCap checks if unified frequency cap is exceeded across devices
func (s *BiddingService) checkCrossDeviceFreqCap(campaign *model.Campaign, req *model.BidRequest) (bool, *CrossDeviceResult) {
	// Cross-device not enabled - fall back to standard check
	if !campaign.Targeting.CrossDeviceEnabled {
		return false, nil
	}

	if req.User.ID == "" && req.Device.DeviceID == "" {
		return false, nil
	}

	// Get device identifier
	deviceID := req.User.ID
	if deviceID == "" {
		deviceID = req.Device.DeviceID
	}

	// Build additional signals for device graph linking
	signals := make(map[string]string)
	// Note: InternalUser doesn't have BuyerUID - would need model update for full OpenRTB support

	// Resolve cross-device identity
	xdevResult := s.ResolveCrossDeviceUser(deviceID, signals)

	// Check unified frequency cap
	if campaign.Targeting.FreqCapImpressions > 0 {
		xdevResult.UnifiedFreq = s.GetCrossDeviceFrequency(xdevResult.PrimaryUserID, campaign.ID)
		if xdevResult.UnifiedFreq >= int64(campaign.Targeting.FreqCapImpressions) {
			xdevResult.FreqCapExceeded = true
			return true, xdevResult // Exceeded cap across devices
		}
	}

	return false, xdevResult
}

// LinkUserDevices explicitly links multiple device IDs under a primary user ID
// This is typically called when user authenticates on multiple devices
func (s *BiddingService) LinkUserDevices(primaryUserID string, deviceIDs []string) error {
	if primaryUserID == "" || len(deviceIDs) == 0 {
		return fmt.Errorf("primaryUserID and deviceIDs required")
	}
	return s.cache.LinkDevices(primaryUserID, deviceIDs, 90)
}

// GetUserDeviceGraph returns the complete device graph for a user
func (s *BiddingService) GetUserDeviceGraph(userID string) (*CrossDeviceResult, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID required")
	}

	// First check if this is a primary ID or device ID
	primaryID, err := s.cache.GetPrimaryUserID(userID)
	if err != nil || primaryID == "" {
		primaryID = userID // Might already be the primary ID
	}

	devices, err := s.cache.GetLinkedDevices(primaryID)
	if err != nil {
		return nil, err
	}

	return &CrossDeviceResult{
		PrimaryUserID: primaryID,
		LinkedDevices: devices,
		DeviceCount:   len(devices),
	}, nil
}
