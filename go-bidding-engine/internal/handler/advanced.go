package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/taskirx/go-bidding-engine/internal/model"
	"github.com/taskirx/go-bidding-engine/internal/service"
)

// AdvancedHandler handles HTTP requests for advanced services
type AdvancedHandler struct {
	biddingService service.BiddingServiceAPI
}

// NewAdvancedHandler creates a new advanced handler
func NewAdvancedHandler(biddingService service.BiddingServiceAPI) *AdvancedHandler {
	return &AdvancedHandler{
		biddingService: biddingService,
	}
}

// ============================================================================
// BID LANDSCAPE ENDPOINTS
// ============================================================================

// BidLandscapeRequest represents a request to analyze bid landscape
type BidLandscapeRequest struct {
	CampaignID  string `json:"campaign_id" binding:"required"`
	PublisherID string `json:"publisher_id"`
	DeviceType  string `json:"device_type"`
}

// HandleBidLandscapeAnalysis analyzes the bid landscape for a campaign
func (h *AdvancedHandler) HandleBidLandscapeAnalysis(c *gin.Context) {
	var req BidLandscapeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetBidLandscapeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Bid landscape service not available"})
		return
	}

	// Create a mock campaign and request for analysis
	campaign := &model.Campaign{
		ID: req.CampaignID,
		Targeting: model.Targeting{
			BidLandscape: &model.BidLandscape{
				Enabled: true,
			},
		},
	}

	bidReq := &model.BidRequest{
		PublisherID: req.PublisherID,
		Device:      model.InternalDevice{Type: req.DeviceType},
	}

	result := svc.AnalyzeLandscape(campaign, bidReq)
	c.JSON(http.StatusOK, result)
}

// RecordBidRequest represents a request to record a bid outcome
type RecordBidRequest struct {
	PublisherID string  `json:"publisher_id" binding:"required"`
	DeviceType  string  `json:"device_type"`
	BidPrice    float64 `json:"bid_price" binding:"required"`
	WinPrice    float64 `json:"win_price"`
	Won         bool    `json:"won"`
}

// HandleRecordBid records a bid outcome for landscape analysis
func (h *AdvancedHandler) HandleRecordBid(c *gin.Context) {
	var req RecordBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetBidLandscapeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Bid landscape service not available"})
		return
	}

	bidReq := &model.BidRequest{
		PublisherID: req.PublisherID,
		Device:      model.InternalDevice{Type: req.DeviceType},
	}

	svc.RecordBid(bidReq, req.BidPrice, req.WinPrice, req.Won)
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

// ============================================================================
// CREATIVE OPTIMIZATION ENDPOINTS
// ============================================================================

// CreativeSelectRequest represents a request to select a creative
type CreativeSelectRequest struct {
	CampaignID string   `json:"campaign_id" binding:"required"`
	Formats    []string `json:"formats"`
	UserID     string   `json:"user_id"`
}

// HandleCreativeSelect selects the optimal creative for a request
func (h *AdvancedHandler) HandleCreativeSelect(c *gin.Context) {
	var req CreativeSelectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetCreativeOptimizationService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Creative optimization service not available"})
		return
	}

	campaign := &model.Campaign{
		ID: req.CampaignID,
		Targeting: model.Targeting{
			CreativeOptimization: &model.CreativeOptimization{
				Enabled: true,
			},
		},
	}

	bidReq := &model.BidRequest{
		AdSlot: model.AdSlot{Formats: req.Formats},
		User:   model.InternalUser{ID: req.UserID},
	}

	result := svc.SelectCreative(campaign, bidReq)
	c.JSON(http.StatusOK, result)
}

// ============================================================================
// INCREMENTALITY ENDPOINTS
// ============================================================================

// IncrementalityEvalRequest represents a request to evaluate incrementality
type IncrementalityEvalRequest struct {
	CampaignID     string  `json:"campaign_id" binding:"required"`
	ExperimentID   string  `json:"experiment_id"`
	UserID         string  `json:"user_id" binding:"required"`
	ControlPercent float64 `json:"control_percent"`
}

// HandleIncrementalityEval evaluates if a user should be in control/test group
func (h *AdvancedHandler) HandleIncrementalityEval(c *gin.Context) {
	var req IncrementalityEvalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetIncrementalityService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Incrementality service not available"})
		return
	}

	controlPercent := req.ControlPercent
	if controlPercent <= 0 {
		controlPercent = 10.0
	}

	campaign := &model.Campaign{
		ID: req.CampaignID,
		Targeting: model.Targeting{
			IncrementalityConfig: &model.IncrementalityConfig{
				Enabled:        true,
				ExperimentID:   req.ExperimentID,
				ControlPercent: controlPercent,
			},
		},
	}

	bidReq := &model.BidRequest{
		User: model.InternalUser{ID: req.UserID},
	}

	result := svc.EvaluateUser(campaign, bidReq)
	c.JSON(http.StatusOK, result)
}

// HandleGetExperimentResults returns incrementality experiment results
func (h *AdvancedHandler) HandleGetExperimentResults(c *gin.Context) {
	experimentID := c.Param("experiment_id")
	if experimentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "experiment_id required"})
		return
	}

	svc := h.biddingService.GetIncrementalityService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Incrementality service not available"})
		return
	}

	result := svc.GetExperimentResults(experimentID)
	c.JSON(http.StatusOK, result)
}

// RecordConversionRequest represents a request to record a conversion
type RecordConversionRequest struct {
	ExperimentID string  `json:"experiment_id" binding:"required"`
	UserID       string  `json:"user_id" binding:"required"`
	IsControl    bool    `json:"is_control"`
	Revenue      float64 `json:"revenue"`
}

// HandleRecordConversion records a conversion for incrementality tracking
func (h *AdvancedHandler) HandleRecordConversion(c *gin.Context) {
	var req RecordConversionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetIncrementalityService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Incrementality service not available"})
		return
	}

	svc.RecordConversion(req.ExperimentID, req.UserID, req.IsControl, req.Revenue)
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

// ============================================================================
// PRIVACY SANDBOX ENDPOINTS
// ============================================================================

// TopicRegistrationRequest represents a request to register user topics
type TopicRegistrationRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	TopicID int    `json:"topic_id" binding:"required"`
}

// HandleRegisterTopic registers a topic for a user
func (h *AdvancedHandler) HandleRegisterTopic(c *gin.Context) {
	var req TopicRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetPrivacySandboxService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Privacy sandbox service not available"})
		return
	}

	svc.RegisterUserTopic(req.UserID, req.TopicID)
	c.JSON(http.StatusOK, gin.H{"status": "registered"})
}

// InterestGroupRequest represents a request to add user to interest group
type InterestGroupRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	GroupID string `json:"group_id" binding:"required"`
}

// HandleAddToInterestGroup adds a user to an interest group
func (h *AdvancedHandler) HandleAddToInterestGroup(c *gin.Context) {
	var req InterestGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetPrivacySandboxService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Privacy sandbox service not available"})
		return
	}

	svc.AddToInterestGroup(req.UserID, req.GroupID)
	c.JSON(http.StatusOK, gin.H{"status": "added"})
}

// HandleGetInterestGroups returns user's interest groups
func (h *AdvancedHandler) HandleGetInterestGroups(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}

	svc := h.biddingService.GetPrivacySandboxService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Privacy sandbox service not available"})
		return
	}

	groups := svc.GetUserInterestGroups(userID)
	c.JSON(http.StatusOK, gin.H{"user_id": userID, "interest_groups": groups})
}

// ============================================================================
// CONTEXTUAL AI ENDPOINTS
// ============================================================================

// ContextAnalysisRequest represents a request to analyze content context
type ContextAnalysisRequest struct {
	CampaignID       string                 `json:"campaign_id" binding:"required"`
	PublisherID      string                 `json:"publisher_id"`
	Context          map[string]interface{} `json:"context"`
	BrandSafetyLevel string                 `json:"brand_safety_level"`
}

// HandleContextAnalysis analyzes content context for brand safety and targeting
func (h *AdvancedHandler) HandleContextAnalysis(c *gin.Context) {
	var req ContextAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetContextualAIService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Contextual AI service not available"})
		return
	}

	campaign := &model.Campaign{
		ID:               req.CampaignID,
		BrandSafetyLevel: req.BrandSafetyLevel,
		Targeting: model.Targeting{
			ContextualAI: &model.ContextualAI{
				Enabled:          true,
				AnalyzeContent:   true,
				AnalyzeSentiment: true,
			},
		},
	}

	bidReq := &model.BidRequest{
		PublisherID: req.PublisherID,
		Context:     req.Context,
	}

	result := svc.AnalyzeContext(campaign, bidReq)
	c.JSON(http.StatusOK, result)
}

// ============================================================================
// REAL-TIME ALERTS ENDPOINTS
// ============================================================================

// AlertCheckRequest represents a request to check alerts
type AlertCheckRequest struct {
	CampaignID   string  `json:"campaign_id" binding:"required"`
	CurrentSpend float64 `json:"current_spend"`
	Budget       float64 `json:"budget"`
}

// HandleCheckAlerts checks real-time alerts for a campaign
func (h *AdvancedHandler) HandleCheckAlerts(c *gin.Context) {
	var req AlertCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetRealTimeAlertService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Real-time alert service not available"})
		return
	}

	campaign := &model.Campaign{
		ID: req.CampaignID,
		Targeting: model.Targeting{
			AlertConfig: &model.AlertConfig{
				Enabled: true,
				BudgetAlerts: &model.BudgetAlerts{
					Enabled:           true,
					WarnAtPercent:     80.0,
					CriticalAtPercent: 95.0,
				},
			},
		},
	}

	result := svc.CheckAlerts(campaign, req.CurrentSpend, req.Budget)
	c.JSON(http.StatusOK, result)
}

// RecordMetricsRequest represents a request to record campaign metrics
type RecordMetricsRequest struct {
	CampaignID string  `json:"campaign_id" binding:"required"`
	Spend      float64 `json:"spend"`
	CTR        float64 `json:"ctr"`
	CVR        float64 `json:"cvr"`
	WinRate    float64 `json:"win_rate"`
}

// HandleRecordMetrics records metrics for anomaly detection
func (h *AdvancedHandler) HandleRecordMetrics(c *gin.Context) {
	var req RecordMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetRealTimeAlertService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Real-time alert service not available"})
		return
	}

	svc.RecordMetrics(req.CampaignID, req.Spend, req.CTR, req.CVR, req.WinRate)
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

// ============================================================================
// COMPETITIVE INTELLIGENCE ENDPOINTS
// ============================================================================

// CompetitiveAnalysisRequest represents a request for competitive analysis
type CompetitiveAnalysisRequest struct {
	CampaignID      string   `json:"campaign_id" binding:"required"`
	PublisherID     string   `json:"publisher_id"`
	AdSlotID        string   `json:"ad_slot_id"`
	CompetitiveMode string   `json:"competitive_mode"`
	Competitors     []string `json:"competitors"`
}

// HandleCompetitiveAnalysis analyzes competitive landscape
func (h *AdvancedHandler) HandleCompetitiveAnalysis(c *gin.Context) {
	var req CompetitiveAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetCompetitiveIntelligenceService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Competitive intelligence service not available"})
		return
	}

	campaign := &model.Campaign{
		ID: req.CampaignID,
		Targeting: model.Targeting{
			CompetitiveIntelligence: &model.CompetitiveIntelligence{
				Enabled:          true,
				TrackCompetitors: req.Competitors,
				CompetitiveMode:  req.CompetitiveMode,
			},
		},
	}

	bidReq := &model.BidRequest{
		PublisherID: req.PublisherID,
		AdSlot:      model.AdSlot{ID: req.AdSlotID},
	}

	result := svc.AnalyzeCompetition(campaign, bidReq)
	c.JSON(http.StatusOK, result)
}

// AuctionOutcomeRequest represents a request to record auction outcome
type AuctionOutcomeRequest struct {
	PublisherID  string  `json:"publisher_id" binding:"required"`
	AdSlotID     string  `json:"ad_slot_id"`
	BidPrice     float64 `json:"bid_price" binding:"required"`
	WinningPrice float64 `json:"winning_price"`
	Won          bool    `json:"won"`
	WinnerID     string  `json:"winner_id"`
}

// HandleRecordAuctionOutcome records an auction outcome
func (h *AdvancedHandler) HandleRecordAuctionOutcome(c *gin.Context) {
	var req AuctionOutcomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetCompetitiveIntelligenceService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Competitive intelligence service not available"})
		return
	}

	bidReq := &model.BidRequest{
		PublisherID: req.PublisherID,
		AdSlot:      model.AdSlot{ID: req.AdSlotID},
	}

	svc.RecordAuctionOutcome(bidReq, req.BidPrice, req.WinningPrice, req.Won, req.WinnerID)
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

// HandleGetMarketReport returns competitive market report
func (h *AdvancedHandler) HandleGetMarketReport(c *gin.Context) {
	svc := h.biddingService.GetCompetitiveIntelligenceService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Competitive intelligence service not available"})
		return
	}

	report := svc.GetMarketReport()
	c.JSON(http.StatusOK, report)
}

// ============================================================================
// UNIFIED ID ENDPOINTS
// ============================================================================

// IdentityResolveRequest represents a request to resolve user identity
type IdentityResolveRequest struct {
	CampaignID      string   `json:"campaign_id" binding:"required"`
	UserID          string   `json:"user_id"`
	DeviceID        string   `json:"device_id"`
	Providers       []string `json:"providers"`
	ConsentRequired bool     `json:"consent_required"`
	HasConsent      bool     `json:"has_consent"`
}

// HandleResolveIdentity resolves user identity across providers
func (h *AdvancedHandler) HandleResolveIdentity(c *gin.Context) {
	var req IdentityResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetUnifiedIDService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unified ID service not available"})
		return
	}

	// Build providers list
	providers := make([]model.IDProvider, len(req.Providers))
	for i, p := range req.Providers {
		providers[i] = model.IDProvider{Name: p, Enabled: true, Priority: i + 1}
	}

	campaign := &model.Campaign{
		ID: req.CampaignID,
		Targeting: model.Targeting{
			UnifiedIDConfig: &model.UnifiedIDConfig{
				Enabled:         true,
				Providers:       providers,
				ConsentRequired: req.ConsentRequired,
			},
		},
	}

	context := map[string]interface{}{}
	if req.HasConsent {
		context["gdpr_consent"] = true
	}

	bidReq := &model.BidRequest{
		User:    model.InternalUser{ID: req.UserID},
		Device:  model.InternalDevice{DeviceID: req.DeviceID},
		Context: context,
	}

	result := svc.ResolveIdentity(campaign, bidReq)
	c.JSON(http.StatusOK, result)
}

// LinkIdentitiesRequest represents a request to link identities
type LinkIdentitiesRequest struct {
	ID1        string  `json:"id1" binding:"required"`
	Provider1  string  `json:"provider1" binding:"required"`
	ID2        string  `json:"id2" binding:"required"`
	Provider2  string  `json:"provider2" binding:"required"`
	DeviceType string  `json:"device_type"`
	Confidence float64 `json:"confidence"`
}

// HandleLinkIdentities links two identities
func (h *AdvancedHandler) HandleLinkIdentities(c *gin.Context) {
	var req LinkIdentitiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetUnifiedIDService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unified ID service not available"})
		return
	}

	confidence := req.Confidence
	if confidence <= 0 {
		confidence = 0.8
	}

	svc.LinkIdentities(req.ID1, req.Provider1, req.ID2, req.Provider2, req.DeviceType, confidence)
	c.JSON(http.StatusOK, gin.H{"status": "linked"})
}

// HandleGetIdentityReport returns identity resolution report
func (h *AdvancedHandler) HandleGetIdentityReport(c *gin.Context) {
	svc := h.biddingService.GetUnifiedIDService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unified ID service not available"})
		return
	}

	report := svc.GetIdentityReport()
	c.JSON(http.StatusOK, report)
}

// HandleGetCrossDeviceReach returns cross-device reach metrics
func (h *AdvancedHandler) HandleGetCrossDeviceReach(c *gin.Context) {
	svc := h.biddingService.GetUnifiedIDService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unified ID service not available"})
		return
	}

	reach := svc.CalculateCrossDeviceReach()
	c.JSON(http.StatusOK, gin.H{"cross_device_reach": reach})
}

// ============================================================================
// DYNAMIC BID ADJUSTMENT ENDPOINTS
// ============================================================================

// DynamicBidRequest represents a request for dynamic bid calculation
type DynamicBidRequest struct {
	CampaignID   string  `json:"campaign_id" binding:"required"`
	PublisherID  string  `json:"publisher_id" binding:"required"`
	DeviceType   string  `json:"device_type"`
	Country      string  `json:"country"`
	BaseBid      float64 `json:"base_bid" binding:"required"`
	UserID       string  `json:"user_id"`
	AdSlotWidth  int     `json:"ad_slot_width"`
	AdSlotHeight int     `json:"ad_slot_height"`
}

// HandleCalculateDynamicBid calculates a dynamic bid price
func (h *AdvancedHandler) HandleCalculateDynamicBid(c *gin.Context) {
	var req DynamicBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetDynamicBidService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Dynamic bid service not available"})
		return
	}

	// Build campaign model
	campaign := &model.Campaign{
		ID:       req.CampaignID,
		BidPrice: req.BaseBid,
	}

	// Build bid request model
	bidReq := &model.BidRequest{
		ID:          "dynamic-" + req.CampaignID,
		PublisherID: req.PublisherID,
		User: model.InternalUser{
			ID: req.UserID,
		},
		Device: model.InternalDevice{
			Type: req.DeviceType,
			Geo: model.InternalGeo{
				Country: req.Country,
			},
		},
		AdSlot: model.AdSlot{
			Dimensions: []int{req.AdSlotWidth, req.AdSlotHeight},
		},
	}

	result := svc.CalculateDynamicBid(campaign, bidReq)
	c.JSON(http.StatusOK, result)
}

// DynamicBidOutcomeRequest represents a bid outcome to record
type DynamicBidOutcomeRequest struct {
	CampaignID  string  `json:"campaign_id" binding:"required"`
	PublisherID string  `json:"publisher_id" binding:"required"`
	UserID      string  `json:"user_id"`
	BidPrice    float64 `json:"bid_price" binding:"required"`
	Won         bool    `json:"won"`
	WinPrice    float64 `json:"win_price"`
	Clicked     bool    `json:"clicked"`
	Converted   bool    `json:"converted"`
	Revenue     float64 `json:"revenue"`
}

// HandleRecordDynamicBidOutcome records a bid outcome for learning
func (h *AdvancedHandler) HandleRecordDynamicBidOutcome(c *gin.Context) {
	var req DynamicBidOutcomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetDynamicBidService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Dynamic bid service not available"})
		return
	}

	// Build bid request for context
	bidReq := &model.BidRequest{
		ID:          "outcome-" + req.CampaignID,
		PublisherID: req.PublisherID,
		User: model.InternalUser{
			ID: req.UserID,
		},
	}

	svc.RecordOutcome(req.CampaignID, bidReq, req.BidPrice, req.WinPrice, req.Won, req.Clicked, req.Converted, req.Revenue)
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

// HandleGetDynamicBidAnalytics returns bid analytics
func (h *AdvancedHandler) HandleGetDynamicBidAnalytics(c *gin.Context) {
	svc := h.biddingService.GetDynamicBidService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Dynamic bid service not available"})
		return
	}

	analytics := svc.GetBidAnalytics()
	c.JSON(http.StatusOK, analytics)
}

// HandleGetDynamicBidConfig returns the dynamic bid service configuration
func (h *AdvancedHandler) HandleGetDynamicBidConfig(c *gin.Context) {
	svc := h.biddingService.GetDynamicBidService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Dynamic bid service not available"})
		return
	}

	config := svc.GetConfig()
	c.JSON(http.StatusOK, config)
}

// ============================================================================
// LOOKALIKE AUDIENCE ENDPOINTS
// ============================================================================

// LookalikeGenerateRequest represents a request to generate a lookalike audience
type LookalikeGenerateRequest struct {
	SeedUserIDs     []string `json:"seed_user_ids" binding:"required"`
	Name            string   `json:"name" binding:"required"`
	ExpansionFactor float64  `json:"expansion_factor"`
}

// HandleGenerateLookalike generates a lookalike audience
func (h *AdvancedHandler) HandleGenerateLookalike(c *gin.Context) {
	var req LookalikeGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetLookalikeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lookalike service not available"})
		return
	}

	expansionFactor := req.ExpansionFactor
	if expansionFactor == 0 {
		expansionFactor = 2.0 // Default
	}

	result := svc.GenerateLookalike(req.SeedUserIDs, req.Name, expansionFactor)
	c.JSON(http.StatusOK, result)
}

// UserProfileRequest represents a user profile registration request
type UserProfileRequest struct {
	UserID      string   `json:"user_id" binding:"required"`
	Segments    []string `json:"segments"`
	Interests   []string `json:"interests"`
	DeviceTypes []string `json:"device_types"`
	Country     string   `json:"country"`
	Region      string   `json:"region"`
	City        string   `json:"city"`
}

// HandleRegisterUserProfile registers a user profile for lookalike modeling
func (h *AdvancedHandler) HandleRegisterUserProfile(c *gin.Context) {
	var req UserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetLookalikeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lookalike service not available"})
		return
	}

	profile := svc.CreateUserProfile(req.Segments, req.Interests, req.DeviceTypes, req.Country, req.Region, req.City)
	svc.RegisterUserProfile(req.UserID, profile)
	c.JSON(http.StatusOK, gin.H{"status": "registered"})
}

// HandleGetLookalikeAudience retrieves a lookalike audience
func (h *AdvancedHandler) HandleGetLookalikeAudience(c *gin.Context) {
	audienceID := c.Param("audience_id")
	if audienceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "audience_id is required"})
		return
	}

	svc := h.biddingService.GetLookalikeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lookalike service not available"})
		return
	}

	audience := svc.GetLookalikeAudience(audienceID)
	if audience == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audience not found"})
		return
	}

	c.JSON(http.StatusOK, audience)
}

// HandleIsUserInLookalike checks if a user is in a lookalike audience
func (h *AdvancedHandler) HandleIsUserInLookalike(c *gin.Context) {
	userID := c.Query("user_id")
	audienceID := c.Query("audience_id")
	if userID == "" || audienceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and audience_id are required"})
		return
	}

	svc := h.biddingService.GetLookalikeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lookalike service not available"})
		return
	}

	isMember, score := svc.IsUserInLookalike(userID, audienceID)
	c.JSON(http.StatusOK, gin.H{
		"is_member":        isMember,
		"similarity_score": score,
	})
}

// HandleGetLookalikeStats returns lookalike audience statistics
func (h *AdvancedHandler) HandleGetLookalikeStats(c *gin.Context) {
	svc := h.biddingService.GetLookalikeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lookalike service not available"})
		return
	}

	stats := svc.GetLookalikeStats()
	c.JSON(http.StatusOK, stats)
}

// ============================================================================
// USER CLUSTERING ENDPOINTS
// ============================================================================

// ClusterUserRequest represents a user clustering registration request
type ClusterUserRequest struct {
	UserID        string    `json:"user_id" binding:"required"`
	FeatureVector []float64 `json:"feature_vector" binding:"required"`
}

// HandleRegisterClusterUser registers a user for clustering
func (h *AdvancedHandler) HandleRegisterClusterUser(c *gin.Context) {
	var req ClusterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetUserClusteringService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "User clustering service not available"})
		return
	}

	svc.RegisterUser(req.UserID, req.FeatureVector)
	c.JSON(http.StatusOK, gin.H{"status": "registered"})
}

// HandleRunClustering triggers the clustering algorithm
func (h *AdvancedHandler) HandleRunClustering(c *gin.Context) {
	svc := h.biddingService.GetUserClusteringService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "User clustering service not available"})
		return
	}

	result := svc.RunClustering()
	c.JSON(http.StatusOK, result)
}

// HandleGetUserCluster returns the cluster for a specific user
func (h *AdvancedHandler) HandleGetUserCluster(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	svc := h.biddingService.GetUserClusteringService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "User clustering service not available"})
		return
	}

	cluster, confidence := svc.GetUserCluster(userID)
	if cluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found or not clustered"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cluster":    cluster,
		"confidence": confidence,
	})
}

// HandleGetClusterUsers returns users in a specific cluster
func (h *AdvancedHandler) HandleGetClusterUsers(c *gin.Context) {
	clusterID := c.Param("cluster_id")
	if clusterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster_id is required"})
		return
	}

	svc := h.biddingService.GetUserClusteringService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "User clustering service not available"})
		return
	}

	users := svc.GetClusterUsers(clusterID)
	c.JSON(http.StatusOK, gin.H{
		"cluster_id": clusterID,
		"user_count": len(users),
		"users":      users,
	})
}

// HandleGetClusteringStats returns clustering statistics
func (h *AdvancedHandler) HandleGetClusteringStats(c *gin.Context) {
	svc := h.biddingService.GetUserClusteringService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "User clustering service not available"})
		return
	}

	stats := svc.GetClusteringStats()
	c.JSON(http.StatusOK, stats)
}

// ============================================================================
// CHURN PREDICTION ENDPOINTS
// ============================================================================

// ChurnActivityRequest represents a request to record user activity
type ChurnActivityRequest struct {
	UserID    string                 `json:"user_id" binding:"required"`
	EventType string                 `json:"event_type" binding:"required"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// HandleRecordChurnActivity records user activity for churn prediction
func (h *AdvancedHandler) HandleRecordChurnActivity(c *gin.Context) {
	var req ChurnActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetChurnPredictionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Churn prediction service not available"})
		return
	}

	svc.RecordUserActivity(req.UserID, req.EventType, req.Metadata)
	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

// HandlePredictChurn predicts churn probability for a user
func (h *AdvancedHandler) HandlePredictChurn(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	svc := h.biddingService.GetChurnPredictionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Churn prediction service not available"})
		return
	}

	result := svc.PredictChurn(userID)
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// HandleBatchPredictChurn predicts churn for multiple users
func (h *AdvancedHandler) HandleBatchPredictChurn(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"user_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetChurnPredictionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Churn prediction service not available"})
		return
	}

	results := svc.BatchPredict(req.UserIDs)
	c.JSON(http.StatusOK, gin.H{
		"predictions": results,
		"count":       len(results),
	})
}

// HandleGetHighRiskUsers returns users with high churn risk
func (h *AdvancedHandler) HandleGetHighRiskUsers(c *gin.Context) {
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	svc := h.biddingService.GetChurnPredictionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Churn prediction service not available"})
		return
	}

	users := svc.GetHighRiskUsers(limit)
	c.JSON(http.StatusOK, gin.H{
		"high_risk_users": users,
		"count":           len(users),
	})
}

// HandleGetChurnStats returns churn prediction statistics
func (h *AdvancedHandler) HandleGetChurnStats(c *gin.Context) {
	svc := h.biddingService.GetChurnPredictionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Churn prediction service not available"})
		return
	}

	stats := svc.GetChurnStats()
	c.JSON(http.StatusOK, stats)
}

// ============================================================================
// A/B TESTING ENDPOINTS
// ============================================================================

// HandleCreateExperiment creates a new A/B test experiment
func (h *AdvancedHandler) HandleCreateExperiment(c *gin.Context) {
	var req service.CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetABTestingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "A/B testing service not available"})
		return
	}

	exp, err := svc.CreateExperiment(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, exp)
}

// HandleStartExperiment starts an experiment
func (h *AdvancedHandler) HandleStartExperiment(c *gin.Context) {
	experimentID := c.Param("experiment_id")
	if experimentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "experiment_id is required"})
		return
	}

	svc := h.biddingService.GetABTestingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "A/B testing service not available"})
		return
	}

	if err := svc.StartExperiment(experimentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "started", "experiment_id": experimentID})
}

// HandleStopExperiment stops an experiment
func (h *AdvancedHandler) HandleStopExperiment(c *gin.Context) {
	experimentID := c.Param("experiment_id")
	if experimentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "experiment_id is required"})
		return
	}

	svc := h.biddingService.GetABTestingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "A/B testing service not available"})
		return
	}

	if err := svc.StopExperiment(experimentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "stopped", "experiment_id": experimentID})
}

// HandleGetExperiment retrieves an experiment by ID
func (h *AdvancedHandler) HandleGetExperiment(c *gin.Context) {
	experimentID := c.Param("experiment_id")
	if experimentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "experiment_id is required"})
		return
	}

	svc := h.biddingService.GetABTestingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "A/B testing service not available"})
		return
	}

	exp, err := svc.GetExperiment(experimentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, exp)
}

// HandleListExperiments lists all experiments
func (h *AdvancedHandler) HandleListExperiments(c *gin.Context) {
	status := c.Query("status")

	svc := h.biddingService.GetABTestingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "A/B testing service not available"})
		return
	}

	experiments := svc.ListExperiments(status)
	c.JSON(http.StatusOK, gin.H{
		"experiments": experiments,
		"count":       len(experiments),
	})
}

// HandleGetVariantForUser assigns and returns a variant for a user
func (h *AdvancedHandler) HandleGetVariantForUser(c *gin.Context) {
	experimentID := c.Param("experiment_id")
	userID := c.Param("user_id")

	if experimentID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "experiment_id and user_id are required"})
		return
	}

	svc := h.biddingService.GetABTestingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "A/B testing service not available"})
		return
	}

	variant, err := svc.GetVariantForUser(experimentID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variant)
}

// ABEventRequest represents a request to record an A/B test event
type ABEventRequest struct {
	VariantID string  `json:"variant_id" binding:"required"`
	EventType string  `json:"event_type" binding:"required"`
	Value     float64 `json:"value"`
}

// HandleRecordABEvent records an event for an experiment variant
func (h *AdvancedHandler) HandleRecordABEvent(c *gin.Context) {
	experimentID := c.Param("experiment_id")
	if experimentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "experiment_id is required"})
		return
	}

	var req ABEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetABTestingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "A/B testing service not available"})
		return
	}

	if err := svc.RecordEvent(experimentID, req.VariantID, req.EventType, req.Value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

// HandleAnalyzeExperiment performs statistical analysis on an experiment
func (h *AdvancedHandler) HandleAnalyzeExperiment(c *gin.Context) {
	experimentID := c.Param("experiment_id")
	if experimentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "experiment_id is required"})
		return
	}

	svc := h.biddingService.GetABTestingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "A/B testing service not available"})
		return
	}

	result, err := svc.AnalyzeExperiment(experimentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// HandleGetBanditRecommendation gets Thompson Sampling recommendation
func (h *AdvancedHandler) HandleGetBanditRecommendation(c *gin.Context) {
	experimentID := c.Param("experiment_id")
	userID := c.Param("user_id")

	if experimentID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "experiment_id and user_id are required"})
		return
	}

	svc := h.biddingService.GetABTestingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "A/B testing service not available"})
		return
	}

	variant, err := svc.GetBanditRecommendation(experimentID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variant)
}

// HandleGetABTestingStats returns A/B testing statistics
func (h *AdvancedHandler) HandleGetABTestingStats(c *gin.Context) {
	svc := h.biddingService.GetABTestingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "A/B testing service not available"})
		return
	}

	stats := svc.GetStats()
	c.JSON(http.StatusOK, stats)
}

// ============================================================================
// DYNAMIC CREATIVE OPTIMIZATION (DCO) ENDPOINTS
// ============================================================================

// HandleCreateTemplate creates a new DCO template
func (h *AdvancedHandler) HandleCreateTemplate(c *gin.Context) {
	var template service.CreativeTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetDynamicCreativeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "DCO service not available"})
		return
	}

	result, err := svc.CreateTemplate(&template)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// HandleGetTemplate retrieves a DCO template
func (h *AdvancedHandler) HandleGetTemplate(c *gin.Context) {
	templateID := c.Param("id")

	svc := h.biddingService.GetDynamicCreativeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "DCO service not available"})
		return
	}

	template, err := svc.GetTemplate(templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// HandleCreateElement creates a new DCO element
func (h *AdvancedHandler) HandleCreateElement(c *gin.Context) {
	var element service.CreativeElement
	if err := c.ShouldBindJSON(&element); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetDynamicCreativeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "DCO service not available"})
		return
	}

	result, err := svc.CreateElement(&element)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// HandleGetElement retrieves a DCO element
func (h *AdvancedHandler) HandleGetElement(c *gin.Context) {
	elementID := c.Param("id")

	svc := h.biddingService.GetDynamicCreativeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "DCO service not available"})
		return
	}

	element, err := svc.GetElement(elementID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, element)
}

// GenerateCreativeRequest represents request to generate optimized creative
type GenerateCreativeRequest struct {
	TemplateID string             `json:"template_id" binding:"required"`
	UserID     string             `json:"user_id" binding:"required"`
	Context    service.DCOContext `json:"context"`
}

// HandleGenerateOptimizedCreative generates optimized creative
func (h *AdvancedHandler) HandleGenerateOptimizedCreative(c *gin.Context) {
	var req GenerateCreativeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetDynamicCreativeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "DCO service not available"})
		return
	}

	dcoReq := service.DCORequest{
		TemplateID: req.TemplateID,
		UserID:     req.UserID,
		Context:    req.Context,
	}

	result, err := svc.GenerateOptimizedCreative(dcoReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// HandleRecordDCOImpression records an impression for a combination
func (h *AdvancedHandler) HandleRecordDCOImpression(c *gin.Context) {
	combinationID := c.Param("combination_id")

	svc := h.biddingService.GetDynamicCreativeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "DCO service not available"})
		return
	}

	if err := svc.RecordImpression(combinationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Impression recorded"})
}

// HandleRecordDCOClick records a click for a combination
func (h *AdvancedHandler) HandleRecordDCOClick(c *gin.Context) {
	combinationID := c.Param("combination_id")

	var req struct {
		UserID string `json:"user_id"`
	}
	c.ShouldBindJSON(&req)

	svc := h.biddingService.GetDynamicCreativeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "DCO service not available"})
		return
	}

	if err := svc.RecordClick(combinationID, req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Click recorded"})
}

// HandleRecordDCOConversion records a conversion for a combination
func (h *AdvancedHandler) HandleRecordDCOConversion(c *gin.Context) {
	combinationID := c.Param("combination_id")

	var req struct {
		Revenue float64 `json:"revenue"`
	}
	c.ShouldBindJSON(&req)

	svc := h.biddingService.GetDynamicCreativeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "DCO service not available"})
		return
	}

	if err := svc.RecordConversion(combinationID, req.Revenue); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversion recorded"})
}

// HandleGetTopCombinations returns top performing combinations
func (h *AdvancedHandler) HandleGetTopCombinations(c *gin.Context) {
	templateID := c.Param("template_id")

	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 10
	}

	svc := h.biddingService.GetDynamicCreativeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "DCO service not available"})
		return
	}

	combinations := svc.GetTopCombinations(templateID, limit)
	c.JSON(http.StatusOK, gin.H{"combinations": combinations, "count": len(combinations)})
}

// HandleGetDCOStats returns DCO statistics
func (h *AdvancedHandler) HandleGetDCOStats(c *gin.Context) {
	svc := h.biddingService.GetDynamicCreativeService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "DCO service not available"})
		return
	}

	stats := svc.GetDCOStats()
	c.JSON(http.StatusOK, stats)
}

// ============================================================================
// PERFORMANCE PREDICTION ENDPOINTS
// ============================================================================

// HandleRecordPerformance records performance data
func (h *AdvancedHandler) HandleRecordPerformance(c *gin.Context) {
	var record service.PerformanceRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetPerformancePredictionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Performance prediction service not available"})
		return
	}

	svc.RecordPerformance(&record)
	c.JSON(http.StatusOK, gin.H{"message": "Performance recorded"})
}

// HandlePredictPerformance predicts future performance
func (h *AdvancedHandler) HandlePredictPerformance(c *gin.Context) {
	var req service.PredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetPerformancePredictionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Performance prediction service not available"})
		return
	}

	prediction, err := svc.Predict(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prediction)
}

// HandleForecastPerformance generates a forecast
func (h *AdvancedHandler) HandleForecastPerformance(c *gin.Context) {
	entityID := c.Query("entity_id")
	entityType := c.Query("entity_type")
	hoursStr := c.DefaultQuery("hours", "24")

	if entityID == "" || entityType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_id and entity_type are required"})
		return
	}

	hours, _ := strconv.Atoi(hoursStr)
	if hours <= 0 {
		hours = 24
	}

	svc := h.biddingService.GetPerformancePredictionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Performance prediction service not available"})
		return
	}

	forecast, err := svc.Forecast(entityID, entityType, hours)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, forecast)
}

// HandleGetPredictionAccuracy returns prediction accuracy for an entity
func (h *AdvancedHandler) HandleGetPredictionAccuracy(c *gin.Context) {
	entityID := c.Param("entity_id")
	lookbackStr := c.DefaultQuery("lookback_hours", "24")

	lookbackHours, _ := strconv.Atoi(lookbackStr)
	if lookbackHours <= 0 {
		lookbackHours = 24
	}

	svc := h.biddingService.GetPerformancePredictionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Performance prediction service not available"})
		return
	}

	accuracy := svc.GetPredictionAccuracy(entityID, lookbackHours)
	c.JSON(http.StatusOK, gin.H{"entity_id": entityID, "lookback_hours": lookbackHours, "accuracy": accuracy})
}

// HandleGetPredictionStats returns performance prediction statistics
func (h *AdvancedHandler) HandleGetPredictionStats(c *gin.Context) {
	svc := h.biddingService.GetPerformancePredictionService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Performance prediction service not available"})
		return
	}

	stats := svc.GetStats()
	c.JSON(http.StatusOK, stats)
}

// ============================================================================
// S2S BIDDING ENDPOINTS
// ============================================================================

// S2SPartnerRequest represents a request to register/update a demand partner
type S2SPartnerRequest struct {
	ID       string            `json:"id" binding:"required"`
	Name     string            `json:"name" binding:"required"`
	Endpoint string            `json:"endpoint" binding:"required"`
	BidFloor float64           `json:"bid_floor"`
	QPS      int               `json:"qps"`
	Headers  map[string]string `json:"headers"`
}

// HandleRegisterS2SPartner registers a new demand partner
func (h *AdvancedHandler) HandleRegisterS2SPartner(c *gin.Context) {
	var req S2SPartnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetS2SBiddingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "S2S bidding service not available"})
		return
	}

	partner := &service.DemandPartner{
		ID:       req.ID,
		Name:     req.Name,
		Endpoint: req.Endpoint,
		BidFloor: req.BidFloor,
		QPS:      req.QPS,
		Headers:  req.Headers,
	}

	if err := svc.RegisterPartner(partner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Partner registered", "partner_id": req.ID})
}

// HandleGetS2SPartner retrieves a demand partner
func (h *AdvancedHandler) HandleGetS2SPartner(c *gin.Context) {
	partnerID := c.Param("id")

	svc := h.biddingService.GetS2SBiddingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "S2S bidding service not available"})
		return
	}

	partner, err := svc.GetPartner(partnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, partner)
}

// HandleListS2SPartners lists all demand partners
func (h *AdvancedHandler) HandleListS2SPartners(c *gin.Context) {
	svc := h.biddingService.GetS2SBiddingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "S2S bidding service not available"})
		return
	}

	partners := svc.ListPartners()
	c.JSON(http.StatusOK, gin.H{"partners": partners, "count": len(partners)})
}

// HandleRemoveS2SPartner removes a demand partner
func (h *AdvancedHandler) HandleRemoveS2SPartner(c *gin.Context) {
	partnerID := c.Param("id")

	svc := h.biddingService.GetS2SBiddingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "S2S bidding service not available"})
		return
	}

	if err := svc.RemovePartner(partnerID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Partner removed"})
}

// HandleS2SBidRequest processes a server-to-server bid request
func (h *AdvancedHandler) HandleS2SBidRequest(c *gin.Context) {
	var req service.S2SBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetS2SBiddingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "S2S bidding service not available"})
		return
	}

	resp, err := svc.ProcessBidRequest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// HandleGetS2SStats returns S2S bidding statistics
func (h *AdvancedHandler) HandleGetS2SStats(c *gin.Context) {
	svc := h.biddingService.GetS2SBiddingService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "S2S bidding service not available"})
		return
	}

	stats := svc.GetStats()
	c.JSON(http.StatusOK, stats)
}

// ============================================================================
// BID CACHE ENDPOINTS
// ============================================================================

// HandleGetBidCacheStats returns bid cache statistics
func (h *AdvancedHandler) HandleGetBidCacheStats(c *gin.Context) {
	svc := h.biddingService.GetBidCacheService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Bid cache service not available"})
		return
	}

	stats := svc.GetStats()
	c.JSON(http.StatusOK, stats)
}

// HandleGetBidCacheHitRate returns cache hit rate
func (h *AdvancedHandler) HandleGetBidCacheHitRate(c *gin.Context) {
	svc := h.biddingService.GetBidCacheService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Bid cache service not available"})
		return
	}

	hitRate := svc.GetHitRate()
	c.JSON(http.StatusOK, gin.H{"hit_rate": hitRate, "hit_rate_percent": hitRate * 100})
}

// HandleClearBidCache clears the bid cache
func (h *AdvancedHandler) HandleClearBidCache(c *gin.Context) {
	svc := h.biddingService.GetBidCacheService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Bid cache service not available"})
		return
	}

	svc.Clear()
	c.JSON(http.StatusOK, gin.H{"message": "Cache cleared"})
}

// HandleInvalidateBidCachePartner invalidates cache for a partner
func (h *AdvancedHandler) HandleInvalidateBidCachePartner(c *gin.Context) {
	partnerID := c.Param("partner_id")

	svc := h.biddingService.GetBidCacheService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Bid cache service not available"})
		return
	}

	count := svc.InvalidatePartner(partnerID)
	c.JSON(http.StatusOK, gin.H{"message": "Cache invalidated", "entries_removed": count})
}

// HandleCleanExpiredBidCache removes expired entries
func (h *AdvancedHandler) HandleCleanExpiredBidCache(c *gin.Context) {
	svc := h.biddingService.GetBidCacheService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Bid cache service not available"})
		return
	}

	cleaned := svc.CleanExpired()
	c.JSON(http.StatusOK, gin.H{"message": "Expired entries cleaned", "entries_removed": cleaned})
}

// ============================================================================
// PROGRAMMATIC GUARANTEED (PG) ENDPOINTS
// ============================================================================

// HandleCreatePGDeal creates a new programmatic guaranteed deal
func (h *AdvancedHandler) HandleCreatePGDeal(c *gin.Context) {
	var deal service.PGDeal
	if err := c.ShouldBindJSON(&deal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetProgrammaticGuaranteedService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PG service not available"})
		return
	}

	result, err := svc.CreateDeal(&deal)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// HandleGetPGDeal retrieves a PG deal
func (h *AdvancedHandler) HandleGetPGDeal(c *gin.Context) {
	dealID := c.Param("id")

	svc := h.biddingService.GetProgrammaticGuaranteedService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PG service not available"})
		return
	}

	deal, err := svc.GetDeal(dealID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deal)
}

// HandleListPGDeals lists all PG deals
func (h *AdvancedHandler) HandleListPGDeals(c *gin.Context) {
	buyerID := c.Query("buyer_id")
	sellerID := c.Query("seller_id")
	status := c.Query("status")

	svc := h.biddingService.GetProgrammaticGuaranteedService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PG service not available"})
		return
	}

	deals := svc.ListDeals(buyerID, sellerID, status)
	c.JSON(http.StatusOK, gin.H{"deals": deals, "count": len(deals)})
}

// HandleActivatePGDeal activates a PG deal
func (h *AdvancedHandler) HandleActivatePGDeal(c *gin.Context) {
	dealID := c.Param("id")

	svc := h.biddingService.GetProgrammaticGuaranteedService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PG service not available"})
		return
	}

	if err := svc.ActivateDeal(dealID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deal activated"})
}

// HandlePausePGDeal pauses a PG deal
func (h *AdvancedHandler) HandlePausePGDeal(c *gin.Context) {
	dealID := c.Param("id")

	svc := h.biddingService.GetProgrammaticGuaranteedService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PG service not available"})
		return
	}

	if err := svc.PauseDeal(dealID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deal paused"})
}

// HandleGetPGDeliveryProgress returns delivery progress for a deal
func (h *AdvancedHandler) HandleGetPGDeliveryProgress(c *gin.Context) {
	dealID := c.Param("id")

	svc := h.biddingService.GetProgrammaticGuaranteedService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PG service not available"})
		return
	}

	progress, err := svc.GetDeliveryProgress(dealID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// HandleGetPGStats returns PG service statistics
func (h *AdvancedHandler) HandleGetPGStats(c *gin.Context) {
	svc := h.biddingService.GetProgrammaticGuaranteedService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PG service not available"})
		return
	}

	stats := svc.GetStats()
	c.JSON(http.StatusOK, stats)
}

// ============================================================================
// DIRECT PUBLISHER ENDPOINTS
// ============================================================================

// HandleRegisterDirectPublisher registers a new direct publisher
func (h *AdvancedHandler) HandleRegisterDirectPublisher(c *gin.Context) {
	var pub service.DirectPublisher
	if err := c.ShouldBindJSON(&pub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := h.biddingService.GetDirectPublisherService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Direct publisher service not available"})
		return
	}

	result, err := svc.RegisterPublisher(&pub)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// HandleGetDirectPublisher retrieves a direct publisher
func (h *AdvancedHandler) HandleGetDirectPublisher(c *gin.Context) {
	publisherID := c.Param("id")

	svc := h.biddingService.GetDirectPublisherService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Direct publisher service not available"})
		return
	}

	pub, err := svc.GetPublisher(publisherID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pub)
}

// HandleListDirectPublishers lists all direct publishers
func (h *AdvancedHandler) HandleListDirectPublishers(c *gin.Context) {
	status := c.Query("status")
	minQualityStr := c.DefaultQuery("min_quality", "0")
	minQuality := 0.0
	fmt.Sscanf(minQualityStr, "%f", &minQuality)

	svc := h.biddingService.GetDirectPublisherService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Direct publisher service not available"})
		return
	}

	publishers := svc.ListPublishers(status, minQuality)
	c.JSON(http.StatusOK, gin.H{"publishers": publishers, "count": len(publishers)})
}

// HandleActivateDirectPublisher activates a publisher
func (h *AdvancedHandler) HandleActivateDirectPublisher(c *gin.Context) {
	publisherID := c.Param("id")

	svc := h.biddingService.GetDirectPublisherService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Direct publisher service not available"})
		return
	}

	if err := svc.ActivatePublisher(publisherID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Publisher activated"})
}

// HandleAnalyzeSupplyPath analyzes supply path for optimization
func (h *AdvancedHandler) HandleAnalyzeSupplyPath(c *gin.Context) {
	publisherID := c.Param("id")

	svc := h.biddingService.GetDirectPublisherService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Direct publisher service not available"})
		return
	}

	result, err := svc.AnalyzeSupplyPath(publisherID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// HandleGetDirectPublisherStats returns direct publisher statistics
func (h *AdvancedHandler) HandleGetDirectPublisherStats(c *gin.Context) {
	svc := h.biddingService.GetDirectPublisherService()
	if svc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Direct publisher service not available"})
		return
	}

	stats := svc.GetStats()
	c.JSON(http.StatusOK, stats)
}

// ============================================================================
// HEALTH & STATUS ENDPOINTS
// ============================================================================

// HandleAdvancedServicesStatus returns status of all advanced services
func (h *AdvancedHandler) HandleAdvancedServicesStatus(c *gin.Context) {
	status := map[string]bool{
		"bid_landscape":            h.biddingService.GetBidLandscapeService() != nil,
		"creative_optimization":    h.biddingService.GetCreativeOptimizationService() != nil,
		"incrementality":           h.biddingService.GetIncrementalityService() != nil,
		"privacy_sandbox":          h.biddingService.GetPrivacySandboxService() != nil,
		"contextual_ai":            h.biddingService.GetContextualAIService() != nil,
		"realtime_alerts":          h.biddingService.GetRealTimeAlertService() != nil,
		"competitive_intelligence": h.biddingService.GetCompetitiveIntelligenceService() != nil,
		"unified_id":               h.biddingService.GetUnifiedIDService() != nil,
		"dynamic_bid":              h.biddingService.GetDynamicBidService() != nil,
		"lookalike":                h.biddingService.GetLookalikeService() != nil,
		"user_clustering":          h.biddingService.GetUserClusteringService() != nil,
		"churn_prediction":         h.biddingService.GetChurnPredictionService() != nil,
		"ab_testing":               h.biddingService.GetABTestingService() != nil,
		"s2s_bidding":              h.biddingService.GetS2SBiddingService() != nil,
		"bid_cache":                h.biddingService.GetBidCacheService() != nil,
		"programmatic_guaranteed":  h.biddingService.GetProgrammaticGuaranteedService() != nil,
		"direct_publisher":         h.biddingService.GetDirectPublisherService() != nil,
	}

	allHealthy := true
	for _, v := range status {
		if !v {
			allHealthy = false
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"healthy":  allHealthy,
		"services": status,
	})
}
