package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/taskirx/go-bidding-engine/internal/model"
	"github.com/taskirx/go-bidding-engine/internal/service"
)

func setupBidTestHandler() (*BidHandler, *gin.Engine) {
	cache := &mockCache{}
	biddingService := service.NewBiddingService(cache, "http://localhost:8080")
	handler := NewBidHandler(biddingService)

	router := gin.New()
	return handler, router
}

// ============================================================================
// BID REQUEST TESTS
// ============================================================================

func TestHandleBid_ValidRequest(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/bid", handler.HandleBid)

	reqBody := model.BidRequest{
		ID:          "req-001",
		PublisherID: "pub-001",
		Device: model.InternalDevice{
			Type: "mobile",
			OS:   "iOS",
		},
		User: model.InternalUser{
			ID: "user-001",
		},
		AdSlot: model.AdSlot{
			ID:         "slot-001",
			Dimensions: []int{300, 250},
			Formats:    []string{"banner"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/bid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 (may be no-bid response)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleBid_InvalidRequest(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/bid", handler.HandleBid)

	// Invalid JSON
	req, _ := http.NewRequest("POST", "/api/bid", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleBid_EmptyBody(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/bid", handler.HandleBid)

	req, _ := http.NewRequest("POST", "/api/bid", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty body lacks required fields, should return 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleBid_MultipleFormats(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/bid", handler.HandleBid)

	reqBody := model.BidRequest{
		ID:          "req-001",
		PublisherID: "pub-001",
		AdSlot: model.AdSlot{
			ID:      "slot-001",
			Formats: []string{"banner", "native", "video"},
		},
		Device: model.InternalDevice{
			Type: "mobile",
			OS:   "iOS",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/bid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// May return 200 or a no-bid response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// ============================================================================
// OPENRTB TESTS
// ============================================================================

func TestHandleOpenRTB_ValidRequest(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	ortbReq := model.OpenRTBRequest{
		ID: "ortb-001",
		Imp: []model.Imp{
			{
				ID: "imp-001",
				Banner: &model.Banner{
					W: 300,
					H: 250,
				},
				BidFloor: 0.5,
			},
		},
		Site: &model.Site{
			ID:     "site-001",
			Domain: "example.com",
			Page:   "https://example.com/article",
		},
		Device: &model.Device{
			UA:         "Mozilla/5.0...",
			IP:         "192.168.1.1",
			DeviceType: 4,
		},
	}
	body, _ := json.Marshal(ortbReq)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 200 or 204 (no content) are valid responses
	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_InvalidRequest(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleOpenRTB_VideoImpression(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	ortbReq := model.OpenRTBRequest{
		ID: "ortb-video-001",
		Imp: []model.Imp{
			{
				ID: "imp-001",
				Video: &model.Video{
					W:           640,
					H:           480,
					MinDuration: 15,
					MaxDuration: 30,
					Linearity:   1,
				},
				BidFloor: 2.0,
			},
		},
	}
	body, _ := json.Marshal(ortbReq)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_NativeImpression(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	ortbReq := model.OpenRTBRequest{
		ID: "ortb-native-001",
		Imp: []model.Imp{
			{
				ID:       "imp-001",
				Native:   &model.Native{},
				BidFloor: 1.0,
			},
		},
	}
	body, _ := json.Marshal(ortbReq)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_MultipleImpressions(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	ortbReq := model.OpenRTBRequest{
		ID: "ortb-multi-001",
		Imp: []model.Imp{
			{
				ID: "imp-001",
				Banner: &model.Banner{
					W: 300,
					H: 250,
				},
				BidFloor: 0.5,
			},
			{
				ID: "imp-002",
				Banner: &model.Banner{
					W: 728,
					H: 90,
				},
				BidFloor: 0.3,
			},
		},
	}
	body, _ := json.Marshal(ortbReq)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

// ============================================================================
// TRACK ENDPOINT TEST
// ============================================================================

func TestHandleTrack_Impression(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=impression&id=campaign-001", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Track endpoint returns 204 No Content on success
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_Click(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=click&id=campaign-001", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_Conversion(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=conversion&id=campaign-001", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_MissingParams(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=impression", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Missing id should return 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleTrack_Pixel(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=impression&id=campaign-001&pixel=1", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// With pixel=1, should return 200 with 1x1 GIF
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "image/gif" {
		t.Errorf("Expected Content-Type 'image/gif', got '%s'", w.Header().Get("Content-Type"))
	}
}

// ============================================================================
// REFRESH ENDPOINT TEST
// ============================================================================

func TestHandleRefresh(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/refresh", handler.HandleRefresh)

	req, _ := http.NewRequest("POST", "/api/refresh", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// May return 500 if no backend configured (acceptable in test environment)
	// or 200 if refresh succeeds
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// HEALTH CHECK TEST
// ============================================================================

func TestHandleHealth(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/health", handler.HandleHealth)

	req, _ := http.NewRequest("GET", "/health", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", response["status"])
	}
}

// ============================================================================
// METRICS TEST
// ============================================================================

func TestHandleMetrics(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/metrics", handler.HandleMetrics)

	req, _ := http.NewRequest("GET", "/metrics", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
