package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/taskirx/go-bidding-engine/internal/model"
	"github.com/taskirx/go-bidding-engine/internal/service"
	"github.com/taskirx/go-bidding-engine/pkg/metrics"
)

// BidHandler handles HTTP bid requests
type BidHandler struct {
	biddingService *service.BiddingService
}

// NewBidHandler creates a new bid handler
func NewBidHandler(biddingService *service.BiddingService) *BidHandler {
	return &BidHandler{
		biddingService: biddingService,
	}
}

// HandleBid processes a bid request
func (h *BidHandler) HandleBid(c *gin.Context) {
	timer := prometheus.NewTimer(metrics.BidLatency)
	defer timer.ObserveDuration()
	metrics.BidRequestsTotal.Inc()

	var req model.BidRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Set timestamp if not provided
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	// Generate request ID if not provided
	if req.ID == "" {
		req.ID = model.GenerateRequestID()
	}

	// Process bid
	response, err := h.biddingService.ProcessBid(&req)
	if err != nil {
		// Return no-bid response
		c.JSON(http.StatusOK, model.NoBidResponse{
			RequestID: req.ID,
			Reason:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HandleHealth returns service health status
func (h *BidHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "go-bidding-engine",
		"timestamp": time.Now(),
	})
}

// HandleMetrics returns service metrics
func (h *BidHandler) HandleMetrics(c *gin.Context) {
	metrics, err := h.biddingService.GetMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get metrics",
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// HandleRefresh refreshes campaigns from backend
func (h *BidHandler) HandleRefresh(c *gin.Context) {
	backendURL := c.DefaultQuery("backend_url", "http://localhost:4000")
	
	if err := h.biddingService.RefreshCampaigns(backendURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to refresh campaigns",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Campaigns refreshed successfully",
		"timestamp": time.Now(),
	})
}
