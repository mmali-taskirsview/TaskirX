package handler

import (
	"net/http"
	"strings"
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
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Increment metrics based on format (Legacy)
	for _, format := range req.AdSlot.Formats {
		metrics.BidRequestsByFormat.WithLabelValues(format).Inc()
		h.biddingService.IncrementFormatStats(format)
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

// HandleOpenRTB processes standard OpenRTB 2.5/2.6 requests
func (h *BidHandler) HandleOpenRTB(c *gin.Context) {
	timer := prometheus.NewTimer(metrics.BidLatency)
	defer timer.ObserveDuration()
	metrics.BidRequestsTotal.Inc()

	var ortbReq model.OpenRTBRequest
	if err := c.ShouldBindJSON(&ortbReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OpenRTB request", "details": err.Error()})
		return
	}

	// 1. Normalize OpenRTB to Internal Model
	internalReq := h.normalizeOpenRTB(&ortbReq)

	// Increment metrics based on format
	for _, format := range internalReq.AdSlot.Formats {
		metrics.BidRequestsByFormat.WithLabelValues(format).Inc()
		h.biddingService.IncrementFormatStats(format)
	}

	// 2. Process Bid using existing engine logic
	response, err := h.biddingService.ProcessBid(internalReq)

	if err != nil {
		// OpenRTB No-Bid Response (Empty 204 or specific JSON)
		// Standard is 204 No Content
		c.Status(http.StatusNoContent)
		return
	}

	// 3. Convert Internal Response to OpenRTB Response
	ortbRes := h.convertToOpenRTBResponse(response, &ortbReq)
	c.JSON(http.StatusOK, ortbRes)
}

// normalizeOpenRTB converts OpenRTB request to internal format
func (h *BidHandler) normalizeOpenRTB(req *model.OpenRTBRequest) *model.BidRequest {
	// Basic mapping logic
	internal := &model.BidRequest{
		ID:        req.ID,
		Timestamp: time.Now(),
		// PublisherID logic: From Site.ID or App.ID
		PublisherID: "unknown",
		AdSlot: model.AdSlot{
			ID:      "unknown",
			Formats: []string{},
		},
		Device: model.InternalDevice{
			Type: "mobile", // Default
			Geo:  model.InternalGeo{},
		},
		User: model.InternalUser{},
	}

	if req.Site != nil {
		internal.PublisherID = req.Site.ID
		internal.Device.Type = "desktop" // Assume desktop for site
	} else if req.App != nil {
		internal.PublisherID = req.App.ID
		internal.Device.Type = "mobile" // Assume mobile for app
	}

	if len(req.Imp) > 0 {
		imp := req.Imp[0] // Simplify: handle first impression only for now
		internal.AdSlot.ID = imp.ID

		// Map bid floor from OpenRTB impression
		internal.AdSlot.BidFloor = imp.BidFloor
		if imp.BidFloorCur != "" {
			internal.AdSlot.BidFloorCur = imp.BidFloorCur
		} else {
			internal.AdSlot.BidFloorCur = "USD"
		}

		// PMP Support
		if imp.Pmp != nil {
			internal.Pmp = imp.Pmp
		}

		if imp.Banner != nil {
			internal.AdSlot.Formats = append(internal.AdSlot.Formats, "banner")
			internal.AdSlot.Dimensions = []int{imp.Banner.W, imp.Banner.H}
		}
		if imp.Video != nil {
			internal.AdSlot.Formats = append(internal.AdSlot.Formats, "video")
		}
		if imp.Audio != nil {
			internal.AdSlot.Formats = append(internal.AdSlot.Formats, "audio")
		}
		if imp.Native != nil {
			internal.AdSlot.Formats = append(internal.AdSlot.Formats, "native")
			// Store raw Native Request payload in Context for later use
			if internal.Context == nil {
				internal.Context = make(map[string]interface{})
			}
			internal.Context["native_request"] = imp.Native.Request
		}
		// OpenRTB 2.5: "instl": 1 means interstitial/fullscreen
		if imp.Instl == 1 {
			internal.AdSlot.Formats = append(internal.AdSlot.Formats, "interstitial")
			// Interstitials often imply Rich Media or full-screen Video support
			// We can also infer "rich_media" compatibility
			internal.AdSlot.Formats = append(internal.AdSlot.Formats, "rich_media")
		}
	}

	if req.Device != nil {
		internal.Device.IP = req.Device.IP
		internal.Device.UserAgent = req.Device.UA
		internal.Device.OS = req.Device.OS
		internal.Device.Model = req.Device.Model
		internal.Device.Make = req.Device.Make

		if req.Device.Geo != nil {
			internal.Device.Geo.Country = req.Device.Geo.Country
			internal.Device.Geo.City = req.Device.Geo.City
			internal.Device.Geo.Lat = req.Device.Geo.Lat
			internal.Device.Geo.Lon = req.Device.Geo.Lon

			// Also set User Country from Device Geo if not already set (common practice)
			if internal.User.Country == "" {
				internal.User.Country = req.Device.Geo.Country
			}
		}

		// Map Devicetype ID (OpenRTB 2.5) to internal string
		// 1=Mobile/Tablet, 2=PC, 3=TV, 4=Phone, 5=Tablet, 6=Connected Device, 7=Set Top Box
		switch req.Device.DeviceType {
		case 2:
			internal.Device.Type = "desktop"
		case 1, 4:
			internal.Device.Type = "mobile"
		case 5:
			internal.Device.Type = "tablet"
		case 3, 7:
			internal.Device.Type = "ctv"
		case 6:
			internal.Device.Type = "connected_device"
		default:
			// Fallback logic if DeviceType is missing (using UA or assumed from Site/App earlier)
			if internal.Device.Type == "" {
				internal.Device.Type = "mobile"
			}
		}
	}

	if req.User != nil {
		internal.User.ID = req.User.ID
		internal.User.Gender = req.User.Gender
		if req.User.Yob > 0 {
			internal.User.Age = time.Now().Year() - req.User.Yob
		}
		// Map Keywords (comma separated)
		if req.User.Keywords != "" {
			parts := strings.Split(req.User.Keywords, ",")
			for _, p := range parts {
				if p != "" {
					internal.User.Categories = append(internal.User.Categories, strings.TrimSpace(p))
				}
			}
		}
		// Map Data Segments (e.g., from DMP)
		if len(req.User.Data) > 0 {
			for _, data := range req.User.Data {
				for _, segment := range data.Segment {
					if segment.ID != "" {
						internal.User.Categories = append(internal.User.Categories, segment.ID)
					}
					// Also add segment Value as category if useful? Or just ID.
					// Often ID is the segment ID.
				}
			}
		}
	}

	return internal
}

// convertToOpenRTBResponse converts internal response to OpenRTB
func (h *BidHandler) convertToOpenRTBResponse(res *model.BidResponse, req *model.OpenRTBRequest) map[string]interface{} {
	// Standard OpenRTB Response Structure
	return map[string]interface{}{
		"id": res.RequestID,
		"seatbid": []map[string]interface{}{
			{
				"bid": []map[string]interface{}{
					{
						"id":    model.GenerateRequestID(),
						"impid": req.Imp[0].ID, // Match request imp ID
						"price": res.BidPrice,
						"adm":   res.AdMarkup,
						"crid":  res.CreativeURL,
						"iurl":  res.ImpressionURL,
					},
				},
			},
		},
		"cur": "USD",
	}
}

// HandleTrack processes tracking pixels and events
func (h *BidHandler) HandleTrack(c *gin.Context) {
	eventType := c.Query("event") // "impression", "click", "video_start", "expand", etc.
	campaignID := c.Query("id")   // Campaign ID

	if eventType == "" || campaignID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing event or id"})
		return
	}

	// Record Metric based on type
	switch eventType {
	case "impression", "click", "view":
		metrics.EventsTotal.WithLabelValues(eventType, campaignID).Inc()
	case "start", "first_quartile", "midpoint", "third_quartile", "complete":
		metrics.VideoEventsTotal.WithLabelValues(eventType, campaignID).Inc()
	case "expand", "collapse", "interact":
		metrics.RichMediaEventsTotal.WithLabelValues(eventType, campaignID).Inc()
	default:
		// Generic event
		metrics.EventsTotal.WithLabelValues(eventType, campaignID).Inc()
	}

	// Return 1x1 GIF for pixels
	if c.Query("pixel") == "1" {
		c.Data(http.StatusOK, "image/gif", []byte{
			0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00,
			0x80, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x21,
			0xF9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2C, 0x00, 0x00,
			0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
			0x01, 0x00, 0x3B,
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// HandleHealth returns service health status
func (h *BidHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "go-bidding-engine",
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
	backendURL := c.Query("backend_url")
	if backendURL == "" {
		backendURL = h.biddingService.GetBackendBaseURL()
	}

	if err := h.biddingService.RefreshCampaigns(backendURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to refresh campaigns",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Campaigns refreshed successfully",
		"timestamp": time.Now(),
	})
}
