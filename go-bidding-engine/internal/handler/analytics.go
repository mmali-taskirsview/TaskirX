package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taskirx/go-bidding-engine/internal/service"
	"github.com/taskirx/go-bidding-engine/pkg/metrics"
)

// AnalyticsHandler handles analytics and SPO requests
type AnalyticsHandler struct {
	biddingService *service.BiddingService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(biddingService *service.BiddingService) *AnalyticsHandler {
	return &AnalyticsHandler{
		biddingService: biddingService,
	}
}

// GetSupplyChainMetrics returns aggregated supply chain metrics
func (h *AnalyticsHandler) GetSupplyChainMetrics(c *gin.Context) {
	timeRange := c.DefaultQuery("timeRange", "1h")

	metrics, err := h.biddingService.GetSupplyPathAnalyticsService().GetSupplyChainMetrics(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get supply chain metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetSupplyPathOptimization returns optimization recommendations
func (h *AnalyticsHandler) GetSupplyPathOptimization(c *gin.Context) {
	timeRange := c.DefaultQuery("timeRange", "1h")

	optimization, err := h.biddingService.GetSupplyPathAnalyticsService().AnalyzeSupplyPathEfficiency(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get supply path optimization",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, optimization)
}

// GetBidPathAnalytics returns detailed analytics for a specific bid request
func (h *AnalyticsHandler) GetBidPathAnalytics(c *gin.Context) {
	requestID := c.Param("requestId")

	analytics, err := h.biddingService.GetSupplyPathAnalyticsService().GetBidPathAnalytics(requestID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     "Bid path analytics not found",
			"requestId": requestID,
			"details":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetServicePerformance returns performance metrics for a specific service
func (h *AnalyticsHandler) GetServicePerformance(c *gin.Context) {
	serviceName := c.Query("serviceName")
	timeRange := c.DefaultQuery("timeRange", "1h")

	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "serviceName parameter is required",
		})
		return
	}

	metrics, err := h.biddingService.GetSupplyPathAnalyticsService().GetServicePerformance(serviceName, timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "Failed to get service performance",
			"serviceName": serviceName,
			"details":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetDirectPublisherAnalysis returns analysis of direct publisher relationship opportunities
func (h *AnalyticsHandler) GetDirectPublisherAnalysis(c *gin.Context) {
	timeRange := c.DefaultQuery("timeRange", "1h")

	analysis, err := h.biddingService.GetSupplyPathAnalyticsService().AnalyzeDirectPublisherOpportunities(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get direct publisher analysis",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetCostBenefitAnalysis returns detailed cost-benefit analysis for optimization scenarios
func (h *AnalyticsHandler) GetCostBenefitAnalysis(c *gin.Context) {
	timeRange := c.DefaultQuery("timeRange", "1h")

	analysis, err := h.biddingService.GetSupplyPathAnalyticsService().CalculateCostBenefitAnalysis(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get cost-benefit analysis",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// TrackClick handles click beacon requests from ad markup.
// URL: GET /api/analytics/track/click?campaign_id=XXX&request_id=YYY&user_id=ZZZ
// Increments the per-campaign daily click counter used for CTR calculation.
// Also records click for CTA (click-through attribution), retargeting, and multi-touch attribution.
func (h *AnalyticsHandler) TrackClick(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	if campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campaign_id is required"})
		return
	}

	// Record in Redis for CTR calculation
	if err := h.biddingService.TrackClick(campaignID); err != nil {
		// Non-fatal: log and continue; don't fail the redirect
		c.Header("X-Track-Error", err.Error())
	}

	// Record for click-through attribution, retargeting, and MTA if user_id provided
	userID := c.Query("user_id")
	requestID := c.Query("request_id")
	if userID != "" {
		// Fire and forget - don't block pixel response
		go func() {
			if requestID != "" {
				_ = h.biddingService.RecordClick(userID, campaignID, requestID)
				// Record touchpoint for multi-touch attribution
				_ = h.biddingService.RecordTouchpoint(userID, campaignID, "click", requestID)
			}
			// Record click event for retargeting
			_ = h.biddingService.RecordUserEvent(userID, campaignID, "click")
		}()
	}

	// Increment Prometheus event counter
	metrics.EventsTotal.WithLabelValues("click", campaignID).Inc()

	// Return 204 or redirect; ad markup typically expects a redirect to the landing page
	redirectURL := c.Query("redirect")
	if redirectURL != "" {
		c.Redirect(http.StatusFound, redirectURL)
		return
	}
	c.Status(http.StatusNoContent)
}

// TrackImpression handles impression beacon requests from ad markup.
// URL: GET /api/analytics/track/impression?campaign_id=XXX&request_id=YYY&user_id=ZZZ&price=0.0025
// Increments the per-campaign daily impression counter used for CTR and win-rate calculation.
// Also records impression for VTA (view-through attribution), retargeting, and multi-touch attribution.
func (h *AnalyticsHandler) TrackImpression(c *gin.Context) {
	campaignID := c.Query("campaign_id")
	if campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campaign_id is required"})
		return
	}

	// Record in Redis for CTR calculation
	if err := h.biddingService.TrackImpression(campaignID); err != nil {
		c.Header("X-Track-Error", err.Error())
	}

	// Record for view-through attribution, retargeting, and MTA if user_id provided
	userID := c.Query("user_id")
	requestID := c.Query("request_id")
	if userID != "" {
		// Fire and forget - don't block pixel response
		go func() {
			if requestID != "" {
				_ = h.biddingService.RecordImpression(userID, campaignID, requestID)
				// Record touchpoint for multi-touch attribution
				_ = h.biddingService.RecordTouchpoint(userID, campaignID, "impression", requestID)
			}
			// Record impression event for retargeting
			_ = h.biddingService.RecordUserEvent(userID, campaignID, "impression")
		}()
	}

	// Increment Prometheus event counter
	metrics.EventsTotal.WithLabelValues("impression", campaignID).Inc()

	// Return 1x1 transparent GIF pixel
	c.Data(http.StatusOK, "image/gif", []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00,
		0x80, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x21,
		0xF9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2C, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
		0x01, 0x00, 0x3B,
	})
}

// GetBidLandscape returns bid/win distribution across price buckets for today.
// URL: GET /api/analytics/bid-landscape
// Returns: { "0.50-1.00": { "bids": 450, "wins": 120, "winRate": 0.267 }, ... }
// Use case: Help advertisers understand "bid $X to win Y% of auctions"
func (h *AnalyticsHandler) GetBidLandscape(c *gin.Context) {
	landscape, err := h.biddingService.GetBidLandscape()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get bid landscape",
			"details": err.Error(),
		})
		return
	}

	// Calculate win rates for each bucket
	response := make(map[string]gin.H)
	for bucket, stats := range landscape {
		bids := stats["bids"]
		wins := stats["wins"]
		winRate := 0.0
		if bids > 0 {
			winRate = float64(wins) / float64(bids)
		}
		response[bucket] = gin.H{
			"bids":    bids,
			"wins":    wins,
			"winRate": winRate,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"date":      "today",
		"landscape": response,
	})
}

// GetAutoBidRecommendations returns bid optimization recommendations based on real-time performance.
// URL: GET /api/analytics/auto-bid-recommendations
// Returns: Array of campaigns with bid adjustment recommendations
func (h *AnalyticsHandler) GetAutoBidRecommendations(c *gin.Context) {
	recommendations, err := h.biddingService.GetAutoBidRecommendations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get auto-bid recommendations",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"timestamp":       time.Now(),
		"recommendations": recommendations,
		"count":           len(recommendations),
	})
}

// GetSegmentPerformance returns performance metrics by segment (device/os/geo).
// URL: GET /api/analytics/segment-performance?type=device
// Returns: { "mobile": { "imps": 5000, "clicks": 150, "ctr": 0.03 }, ... }
func (h *AnalyticsHandler) GetSegmentPerformance(c *gin.Context) {
	segmentType := c.Query("type")
	if segmentType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "type parameter is required (device, os, or geo)",
		})
		return
	}

	if segmentType != "device" && segmentType != "os" && segmentType != "geo" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "type must be one of: device, os, geo",
		})
		return
	}

	performance, err := h.biddingService.GetSegmentPerformance(segmentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get segment performance",
			"details": err.Error(),
		})
		return
	}

	// Calculate CTR for each segment
	response := make(map[string]gin.H)
	for segment, metrics := range performance {
		imps := metrics["imps"]
		clicks := metrics["clicks"]
		ctr := 0.0
		if imps > 0 {
			ctr = float64(clicks) / float64(imps)
		}
		response[segment] = gin.H{
			"impressions": imps,
			"clicks":      clicks,
			"ctr":         ctr,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"date":        "today",
		"segmentType": segmentType,
		"performance": response,
	})
}

// GetOptimalBidFloor returns the recommended bid floor for a publisher based on historical win rates.
// URL: GET /api/analytics/optimal-bid-floor?publisher_id=pub123&target_win_rate=0.6
// Returns: { "publisherId": "pub123", "targetWinRate": 0.6, "recommendedFloor": 1.50 }
func (h *AnalyticsHandler) GetOptimalBidFloor(c *gin.Context) {
	publisherID := c.Query("publisher_id")
	if publisherID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "publisher_id parameter is required",
		})
		return
	}

	targetWinRateStr := c.DefaultQuery("target_win_rate", "0.6")
	var targetWinRate float64
	if _, err := fmt.Sscanf(targetWinRateStr, "%f", &targetWinRate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid target_win_rate parameter",
		})
		return
	}

	if targetWinRate < 0.1 || targetWinRate > 0.95 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "target_win_rate must be between 0.1 and 0.95",
		})
		return
	}

	optimalFloor, err := h.biddingService.GetOptimalBidFloor(publisherID, targetWinRate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to calculate optimal bid floor",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"publisherId":      publisherID,
		"targetWinRate":    targetWinRate,
		"recommendedFloor": optimalFloor,
		"timestamp":        time.Now(),
	})
}

// TrackConversion handles conversion postback and performs attribution
// URL: POST /api/analytics/track/conversion
// Body: { "user_id": "xyz", "campaign_id": "abc", "conversion_type": "purchase", "value": 29.99 }
// Returns: Attribution result (CTA, VTA, or no attribution)
// Also records conversion event for retargeting (can target converters or exclude them)
func (h *AnalyticsHandler) TrackConversion(c *gin.Context) {
	var req struct {
		UserID         string  `json:"user_id" binding:"required"`
		CampaignID     string  `json:"campaign_id" binding:"required"`
		ConversionType string  `json:"conversion_type"` // e.g., "purchase", "signup", "lead"
		Value          float64 `json:"value"`           // Conversion value in dollars
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get attribution
	attrType, requestID, err := h.biddingService.GetAttribution(req.UserID, req.CampaignID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get attribution",
			"details": err.Error(),
		})
		return
	}

	// Record conversion metrics and retargeting event
	if attrType != "none" {
		metrics.EventsTotal.WithLabelValues("conversion_"+attrType, req.CampaignID).Inc()
		// Record conversion event for retargeting (allows targeting or excluding converters)
		go func() {
			_ = h.biddingService.RecordUserEvent(req.UserID, req.CampaignID, "conversion")
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"attribution":     attrType,
		"requestId":       requestID,
		"campaignId":      req.CampaignID,
		"conversionType":  req.ConversionType,
		"conversionValue": req.Value,
		"timestamp":       time.Now(),
	})
}

// GetMultiTouchAttribution returns attribution credit distributed across touchpoints
// URL: GET /api/analytics/multi-touch-attribution?user_id=XXX&campaign_id=YYY&model=linear
// Supported models: linear, time_decay, position_based, first_touch, last_touch
// Returns: Array of touchpoints with attributed credit (sum = 1.0)
func (h *AnalyticsHandler) GetMultiTouchAttribution(c *gin.Context) {
	userID := c.Query("user_id")
	campaignID := c.Query("campaign_id")
	modelType := c.DefaultQuery("model", "linear")

	if userID == "" || campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and campaign_id are required"})
		return
	}

	credits, err := h.biddingService.GetMultiTouchAttribution(userID, campaignID, modelType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get attribution",
			"details": err.Error(),
		})
		return
	}

	// Calculate total credit (should be 1.0)
	totalCredit := 0.0
	for _, credit := range credits {
		totalCredit += credit.Credit
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":      userID,
		"campaignId":  campaignID,
		"model":       modelType,
		"touchpoints": len(credits),
		"totalCredit": totalCredit,
		"attribution": credits,
		"timestamp":   time.Now(),
	})
}

// LinkDevices links multiple device IDs under a primary user ID for cross-device targeting
// URL: POST /api/analytics/cross-device/link
// Body: {"primary_user_id": "user123", "device_ids": ["mobile_abc", "desktop_xyz", "ctv_123"]}
func (h *AnalyticsHandler) LinkDevices(c *gin.Context) {
	var req struct {
		PrimaryUserID string   `json:"primary_user_id" binding:"required"`
		DeviceIDs     []string `json:"device_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.DeviceIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one device_id required"})
		return
	}

	err := h.biddingService.LinkUserDevices(req.PrimaryUserID, req.DeviceIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "linked",
		"primaryUserId": req.PrimaryUserID,
		"linkedDevices": req.DeviceIDs,
		"timestamp":     time.Now(),
	})
}

// GetDeviceGraph returns all linked devices for a user
// URL: GET /api/analytics/cross-device/graph?user_id=XXX
func (h *AnalyticsHandler) GetDeviceGraph(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	graph, err := h.biddingService.GetUserDeviceGraph(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"primaryUserId": graph.PrimaryUserID,
		"linkedDevices": graph.LinkedDevices,
		"deviceCount":   graph.DeviceCount,
		"timestamp":     time.Now(),
	})
}

// GetCrossDeviceFrequency returns aggregated impression frequency across all linked devices
// URL: GET /api/analytics/cross-device/frequency?user_id=XXX&campaign_id=YYY
func (h *AnalyticsHandler) GetCrossDeviceFrequency(c *gin.Context) {
	userID := c.Query("user_id")
	campaignID := c.Query("campaign_id")

	if userID == "" || campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and campaign_id are required"})
		return
	}

	// First resolve to primary user ID
	graph, err := h.biddingService.GetUserDeviceGraph(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get unified frequency
	freq := h.biddingService.GetCrossDeviceFrequency(graph.PrimaryUserID, campaignID)

	c.JSON(http.StatusOK, gin.H{
		"userId":        userID,
		"primaryUserId": graph.PrimaryUserID,
		"campaignId":    campaignID,
		"unifiedFreq":   freq,
		"deviceCount":   graph.DeviceCount,
		"linkedDevices": graph.LinkedDevices,
		"timestamp":     time.Now(),
	})
}
