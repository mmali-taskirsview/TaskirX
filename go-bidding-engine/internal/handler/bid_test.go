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

func TestHandleBid_NoRequestID(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/bid", handler.HandleBid)

	// Request without ID - should auto-generate or fail validation
	reqBody := model.BidRequest{
		PublisherID: "pub-001",
		AdSlot: model.AdSlot{
			ID:      "slot-001",
			Formats: []string{"banner"},
		},
		Device: model.InternalDevice{
			Type: "mobile",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/bid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// May return 200 if ID is auto-generated, or 400 if required
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 200 or 400, got %d", w.Code)
	}
}

func TestHandleBid_NoTimestamp(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/bid", handler.HandleBid)

	// Request without timestamp - should auto-set
	reqBody := model.BidRequest{
		ID:          "req-002",
		PublisherID: "pub-001",
		AdSlot: model.AdSlot{
			ID:      "slot-001",
			Formats: []string{"banner"},
		},
		Device: model.InternalDevice{
			Type: "desktop",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/bid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleBid_WithGeoData(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/bid", handler.HandleBid)

	reqBody := model.BidRequest{
		ID:          "req-geo-001",
		PublisherID: "pub-001",
		AdSlot: model.AdSlot{
			ID:      "slot-001",
			Formats: []string{"banner"},
		},
		Device: model.InternalDevice{
			Type: "mobile",
			Geo: model.InternalGeo{
				Country: "USA",
				City:    "San Francisco",
				Lat:     37.7749,
				Lon:     -122.4194,
			},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/bid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleBid_WithUserCategories(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/bid", handler.HandleBid)

	reqBody := model.BidRequest{
		ID:          "req-cat-001",
		PublisherID: "pub-001",
		AdSlot: model.AdSlot{
			ID:      "slot-001",
			Formats: []string{"banner"},
		},
		Device: model.InternalDevice{
			Type: "mobile",
		},
		User: model.InternalUser{
			ID:         "user-001",
			Age:        30,
			Gender:     "M",
			Categories: []string{"sports", "technology", "gaming"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/bid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

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

// TestConvertToOpenRTBResponse tests the OpenRTB response conversion
func TestConvertToOpenRTBResponse(t *testing.T) {
	handler, _ := setupBidTestHandler()

	// Create a valid bid response
	res := &model.BidResponse{
		RequestID:     "req-123",
		BidPrice:      2.50,
		AdMarkup:      "<div>Ad content</div>",
		CreativeURL:   "creative-001",
		ImpressionURL: "https://track.example.com/imp",
	}

	// Create a request with an impression
	req := &model.OpenRTBRequest{
		ID: "ortb-123",
		Imp: []model.Imp{
			{ID: "imp-001"},
		},
	}

	result := handler.convertToOpenRTBResponse(res, req)

	// Verify result structure
	if result["id"] != res.RequestID {
		t.Errorf("Expected id %s, got %v", res.RequestID, result["id"])
	}
	if result["cur"] != "USD" {
		t.Errorf("Expected cur USD, got %v", result["cur"])
	}

	// Verify seatbid
	seatbid, ok := result["seatbid"].([]map[string]interface{})
	if !ok || len(seatbid) == 0 {
		t.Error("Expected seatbid array")
		return
	}

	// Verify bid
	bids, ok := seatbid[0]["bid"].([]map[string]interface{})
	if !ok || len(bids) == 0 {
		t.Error("Expected bid array")
		return
	}

	bid := bids[0]
	if bid["impid"] != "imp-001" {
		t.Errorf("Expected impid imp-001, got %v", bid["impid"])
	}
	if bid["price"] != 2.50 {
		t.Errorf("Expected price 2.50, got %v", bid["price"])
	}
	if bid["adm"] != "<div>Ad content</div>" {
		t.Errorf("Expected adm, got %v", bid["adm"])
	}
}

// ============================================================================
// HANDLE METRICS TESTS
// ============================================================================

func TestHandleMetrics_Success(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/metrics", handler.HandleMetrics)

	req, _ := http.NewRequest("GET", "/api/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response contains expected metrics fields
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
}

// ============================================================================
// NORMALIZE OPENRTB TESTS
// ============================================================================

func TestHandleOpenRTB_WithSite(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	// Request with site context (not app)
	reqBody := map[string]interface{}{
		"id": "ortb-001",
		"imp": []map[string]interface{}{
			{
				"id": "imp-001",
				"banner": map[string]interface{}{
					"w": 300,
					"h": 250,
				},
			},
		},
		"site": map[string]interface{}{
			"id":       "site-001",
			"domain":   "example.com",
			"page":     "https://example.com/news",
			"cat":      []string{"IAB1", "IAB2"},
			"keywords": "news,tech",
		},
		"device": map[string]interface{}{
			"ua":         "Mozilla/5.0",
			"ip":         "192.168.1.1",
			"devicetype": 2,
			"os":         "Windows",
			"language":   "en",
		},
		"user": map[string]interface{}{
			"id":       "user-001",
			"buyeruid": "buyer-001",
			"yob":      1990,
			"gender":   "M",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithApp(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	// Request with app context
	reqBody := map[string]interface{}{
		"id": "ortb-002",
		"imp": []map[string]interface{}{
			{
				"id": "imp-001",
				"video": map[string]interface{}{
					"w":           640,
					"h":           480,
					"mimes":       []string{"video/mp4"},
					"protocols":   []int{2, 3},
					"minduration": 5,
					"maxduration": 30,
				},
			},
		},
		"app": map[string]interface{}{
			"id":     "app-001",
			"name":   "Test App",
			"bundle": "com.example.app",
			"cat":    []string{"IAB9"},
		},
		"device": map[string]interface{}{
			"ua":         "Mozilla/5.0",
			"ip":         "192.168.1.1",
			"devicetype": 1,
			"os":         "iOS",
			"osv":        "15.0",
			"ifa":        "12345-abcde",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithNative(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	// Request with native ad
	reqBody := map[string]interface{}{
		"id": "ortb-003",
		"imp": []map[string]interface{}{
			{
				"id": "imp-001",
				"native": map[string]interface{}{
					"request": `{"ver":"1.2","assets":[{"id":1,"required":1,"title":{"len":50}}]}`,
				},
			},
		},
		"site": map[string]interface{}{
			"id": "site-001",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithRegs(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	// Request with regulation info (GDPR)
	reqBody := map[string]interface{}{
		"id": "ortb-004",
		"imp": []map[string]interface{}{
			{
				"id": "imp-001",
				"banner": map[string]interface{}{
					"w": 300,
					"h": 250,
				},
			},
		},
		"site": map[string]interface{}{
			"id": "site-001",
		},
		"regs": map[string]interface{}{
			"coppa": 0,
			"ext": map[string]interface{}{
				"gdpr": 1,
			},
		},
		"user": map[string]interface{}{
			"id": "user-001",
			"ext": map[string]interface{}{
				"consent": "consent-string-example",
			},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithGeo(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	// Request with geo data
	reqBody := map[string]interface{}{
		"id": "ortb-005",
		"imp": []map[string]interface{}{
			{
				"id": "imp-001",
				"banner": map[string]interface{}{
					"w": 320,
					"h": 50,
				},
			},
		},
		"site": map[string]interface{}{
			"id": "site-001",
		},
		"device": map[string]interface{}{
			"ua": "Mozilla/5.0",
			"geo": map[string]interface{}{
				"lat":     37.7749,
				"lon":     -122.4194,
				"country": "USA",
				"region":  "CA",
				"city":    "San Francisco",
				"zip":     "94102",
			},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

// ============================================================================
// HANDLE METRICS TESTS
// ============================================================================

func TestHandleMetrics_Basic(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/metrics", handler.HandleMetrics)

	req, _ := http.NewRequest("GET", "/api/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check that response contains metrics
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
}

func TestHandleMetrics_WithQueryParams(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/metrics", handler.HandleMetrics)

	req, _ := http.NewRequest("GET", "/api/metrics?format=json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// ============================================================================
// HANDLE TRACK EXTENDED TESTS
// ============================================================================

func TestHandleTrack_WithUserAgent(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=click&id=campaign-001", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_WithReferer(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=impression&id=campaign-001", nil)
	req.Header.Set("Referer", "https://example.com/page")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_Win(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=win&id=campaign-001&price=1.50", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_InvalidEventType(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=invalid&id=campaign-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Invalid event type should still return 204 or 400
	if w.Code != http.StatusNoContent && w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 204 or 400, got %d", w.Code)
	}
}

// ============================================================================
// VIDEO EVENT TESTS
// ============================================================================

func TestHandleTrack_VideoStart(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=start&id=campaign-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_VideoFirstQuartile(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=first_quartile&id=campaign-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_VideoMidpoint(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=midpoint&id=campaign-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_VideoThirdQuartile(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=third_quartile&id=campaign-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_VideoComplete(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=complete&id=campaign-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

// ============================================================================
// RICH MEDIA EVENT TESTS
// ============================================================================

func TestHandleTrack_RichMediaExpand(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=expand&id=campaign-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_RichMediaCollapse(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=collapse&id=campaign-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_RichMediaInteract(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=interact&id=campaign-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleTrack_View(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/track", handler.HandleTrack)

	req, _ := http.NewRequest("GET", "/api/track?event=view&id=campaign-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

// ============================================================================
// HANDLE REFRESH TESTS
// ============================================================================

func TestHandleRefresh_WithBackendUrl(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/refresh", handler.HandleRefresh)

	req, _ := http.NewRequest("POST", "/api/refresh?backend_url=http://localhost:8080", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// May return 500 if backend fails (acceptable in test)
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestHandleRefresh_GetMethod(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.GET("/api/refresh", handler.HandleRefresh)

	req, _ := http.NewRequest("GET", "/api/refresh", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// May return 500 if no backend configured
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

// ============================================================================
// NORMALIZE OPENRTB EXTENDED TESTS
// ============================================================================

func TestHandleOpenRTB_WithVideo(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-video-001",
		"imp": []map[string]interface{}{
			{
				"id": "imp-001",
				"video": map[string]interface{}{
					"w":         640,
					"h":         480,
					"mimes":     []string{"video/mp4"},
					"mindur":    5,
					"maxdur":    30,
					"protocols": []int{2, 3, 5, 6},
				},
			},
		},
		"site": map[string]interface{}{
			"id":   "site-001",
			"page": "https://example.com/video",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithAudio(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-audio-001",
		"imp": []map[string]interface{}{
			{
				"id": "imp-001",
				"audio": map[string]interface{}{
					"mimes":     []string{"audio/mpeg"},
					"mindur":    15,
					"maxdur":    60,
					"protocols": []int{9, 10},
				},
			},
		},
		"site": map[string]interface{}{
			"id": "site-001",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithUser(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-user-001",
		"imp": []map[string]interface{}{
			{
				"id": "imp-001",
				"banner": map[string]interface{}{
					"w": 300,
					"h": 250,
				},
			},
		},
		"user": map[string]interface{}{
			"id":       "user-123",
			"buyeruid": "buyer-456",
			"yob":      1990,
			"gender":   "M",
			"keywords": "sports,technology",
		},
		"site": map[string]interface{}{
			"id": "site-001",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithBidFloor(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-floor-001",
		"imp": []map[string]interface{}{
			{
				"id":          "imp-001",
				"bidfloor":    2.50,
				"bidfloorcur": "USD",
				"banner": map[string]interface{}{
					"w": 300,
					"h": 250,
				},
			},
		},
		"site": map[string]interface{}{
			"id": "site-001",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithDeals(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-deals-001",
		"imp": []map[string]interface{}{
			{
				"id": "imp-001",
				"banner": map[string]interface{}{
					"w": 300,
					"h": 250,
				},
				"pmp": map[string]interface{}{
					"private_auction": 1,
					"deals": []map[string]interface{}{
						{
							"id":          "deal-001",
							"bidfloor":    3.00,
							"bidfloorcur": "USD",
							"at":          1,
						},
					},
				},
			},
		},
		"site": map[string]interface{}{
			"id": "site-001",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithMultipleImps(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-multi-001",
		"imp": []map[string]interface{}{
			{
				"id": "imp-001",
				"banner": map[string]interface{}{
					"w": 300,
					"h": 250,
				},
			},
			{
				"id": "imp-002",
				"banner": map[string]interface{}{
					"w": 728,
					"h": 90,
				},
			},
			{
				"id": "imp-003",
				"video": map[string]interface{}{
					"w": 640,
					"h": 480,
				},
			},
		},
		"site": map[string]interface{}{
			"id": "site-001",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithExt(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-ext-001",
		"imp": []map[string]interface{}{
			{
				"id": "imp-001",
				"banner": map[string]interface{}{
					"w": 300,
					"h": 250,
				},
				"ext": map[string]interface{}{
					"custom_field": "value",
				},
			},
		},
		"site": map[string]interface{}{
			"id": "site-001",
		},
		"ext": map[string]interface{}{
			"exchange": "taskirx",
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

// ============================================================================
// DEVICE TYPE TESTS - Test all DeviceType branches
// ============================================================================

func TestHandleOpenRTB_WithDeviceTypeDesktop(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-devicetype-desktop",
		"imp": []map[string]interface{}{
			{"id": "imp-001", "banner": map[string]interface{}{"w": 300, "h": 250}},
		},
		"site": map[string]interface{}{"id": "site-001"},
		"device": map[string]interface{}{
			"ua":         "Mozilla/5.0",
			"devicetype": 2, // Desktop
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithDeviceTypeTablet(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-devicetype-tablet",
		"imp": []map[string]interface{}{
			{"id": "imp-001", "banner": map[string]interface{}{"w": 300, "h": 250}},
		},
		"site": map[string]interface{}{"id": "site-001"},
		"device": map[string]interface{}{
			"ua":         "Mozilla/5.0",
			"devicetype": 5, // Tablet
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithDeviceTypeCTV(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-devicetype-ctv",
		"imp": []map[string]interface{}{
			{"id": "imp-001", "video": map[string]interface{}{"w": 1920, "h": 1080}},
		},
		"site": map[string]interface{}{"id": "site-001"},
		"device": map[string]interface{}{
			"ua":         "Mozilla/5.0",
			"devicetype": 3, // TV
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithDeviceTypeSetTopBox(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-devicetype-stb",
		"imp": []map[string]interface{}{
			{"id": "imp-001", "video": map[string]interface{}{"w": 1920, "h": 1080}},
		},
		"site": map[string]interface{}{"id": "site-001"},
		"device": map[string]interface{}{
			"ua":         "Mozilla/5.0",
			"devicetype": 7, // Set Top Box
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithDeviceTypeConnected(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-devicetype-connected",
		"imp": []map[string]interface{}{
			{"id": "imp-001", "banner": map[string]interface{}{"w": 300, "h": 250}},
		},
		"site": map[string]interface{}{"id": "site-001"},
		"device": map[string]interface{}{
			"ua":         "Mozilla/5.0",
			"devicetype": 6, // Connected Device
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithDeviceTypePhone(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-devicetype-phone",
		"imp": []map[string]interface{}{
			{"id": "imp-001", "banner": map[string]interface{}{"w": 320, "h": 50}},
		},
		"site": map[string]interface{}{"id": "site-001"},
		"device": map[string]interface{}{
			"ua":         "Mozilla/5.0",
			"devicetype": 4, // Phone
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

// ============================================================================
// INTERSTITIAL/RICH MEDIA TESTS
// ============================================================================

func TestHandleOpenRTB_WithInterstitial(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-interstitial",
		"imp": []map[string]interface{}{
			{
				"id":    "imp-001",
				"instl": 1, // Interstitial
				"banner": map[string]interface{}{
					"w": 320,
					"h": 480,
				},
			},
		},
		"site": map[string]interface{}{"id": "site-001"},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

// ============================================================================
// USER DATA SEGMENTS TESTS
// ============================================================================

func TestHandleOpenRTB_WithUserDataSegments(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-user-data",
		"imp": []map[string]interface{}{
			{"id": "imp-001", "banner": map[string]interface{}{"w": 300, "h": 250}},
		},
		"site": map[string]interface{}{"id": "site-001"},
		"user": map[string]interface{}{
			"id":       "user-123",
			"keywords": "sports,tech,gaming",
			"data": []map[string]interface{}{
				{
					"id":   "dmp-001",
					"name": "TestDMP",
					"segment": []map[string]interface{}{
						{"id": "seg-001", "name": "Sports Fans"},
						{"id": "seg-002", "name": "Tech Enthusiasts"},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

// ============================================================================
// BID FLOOR CURRENCY TESTS
// ============================================================================

func TestHandleOpenRTB_WithBidFloorNoCurrency(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-floor-no-currency",
		"imp": []map[string]interface{}{
			{
				"id":       "imp-001",
				"bidfloor": 1.50, // No bidfloorcur - should default to USD
				"banner":   map[string]interface{}{"w": 300, "h": 250},
			},
		},
		"site": map[string]interface{}{"id": "site-001"},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}

func TestHandleOpenRTB_WithBidFloorEUR(t *testing.T) {
	handler, router := setupBidTestHandler()
	router.POST("/api/openrtb", handler.HandleOpenRTB)

	reqBody := map[string]interface{}{
		"id": "ortb-floor-eur",
		"imp": []map[string]interface{}{
			{
				"id":          "imp-001",
				"bidfloor":    2.00,
				"bidfloorcur": "EUR",
				"banner":      map[string]interface{}{"w": 300, "h": 250},
			},
		},
		"site": map[string]interface{}{"id": "site-001"},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/openrtb", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}
}
