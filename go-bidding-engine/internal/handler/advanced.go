package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/taskirx/go-bidding-engine/internal/model"
	"github.com/taskirx/go-bidding-engine/internal/service"
)

// AdvancedHandler handles HTTP requests for advanced services
type AdvancedHandler struct {
	biddingService *service.BiddingService
}

// NewAdvancedHandler creates a new advanced handler
func NewAdvancedHandler(biddingService *service.BiddingService) *AdvancedHandler {
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
