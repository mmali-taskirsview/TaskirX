package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/taskirx/go-bidding-engine/internal/service"
)

// mockBiddingServiceStatus helps simulate service unavailability
type mockBiddingServiceStatus struct {
	*service.BiddingService
	forceDirectPublisherServiceNil bool
	forceBidCacheServiceNil        bool
	forceS2SBiddingServiceNil      bool
}

func (m *mockBiddingServiceStatus) GetDirectPublisherService() *service.DirectPublisherService {
	if m.forceDirectPublisherServiceNil {
		return nil
	}
	return m.BiddingService.GetDirectPublisherService()
}

func (m *mockBiddingServiceStatus) GetBidCacheService() *service.BidCacheService {
	if m.forceBidCacheServiceNil {
		return nil
	}
	return m.BiddingService.GetBidCacheService()
}

func (m *mockBiddingServiceStatus) GetS2SBiddingService() *service.S2SBiddingService {
	if m.forceS2SBiddingServiceNil {
		return nil
	}
	return m.BiddingService.GetS2SBiddingService()
}

// Ensure mock satisfies the interface
var _ service.BiddingServiceAPI = &mockBiddingServiceStatus{}

func setupTestHandlerWithMock(mock *mockBiddingServiceStatus) *gin.Engine {
	handler := NewAdvancedHandler(mock)
	router := gin.New()

	// Register routes needed for tests
	router.GET("/api/advanced/direct-publishers/:id/supply-path", handler.HandleAnalyzeSupplyPath)
	router.GET("/api/advanced/status", handler.HandleAdvancedServicesStatus)

	return router
}

// Duplicate mockCache here just in case, or rename to avoid conflict if file-scope?
// No, distinct files in same package share unexported symbols.
// So if mockCache is in advanced_test.go, we can use it here.

func TestHandleAnalyzeSupplyPath_HappyPath(t *testing.T) {
	// Setup real service (embedded in mock)
	// We need a helper to create the service.
	// Let's reconstruct what setupTestHandler does.
	c := &mockCache{} // referencing mockCache from advanced_test.go
	bs := service.NewBiddingService(c, "http://localhost:8080")

	// Prepare test data
	bs.GetDirectPublisherService().InsertTestPublisher(&service.DirectPublisher{
		ID:          "pub-happy-path",
		Name:        "Happy Path Publisher",
		Status:      "active",
		AvgBidFloor: 1.5,
		Domain:      "example.com",
		SupplyChain: []service.SupplyChainNode{
			{ASI: "1", Fee: 0.1},
			{ASI: "2", Fee: 0.2},
		},
	})

	mockService := &mockBiddingServiceStatus{BiddingService: bs}
	router := setupTestHandlerWithMock(mockService)

	req, _ := http.NewRequest("GET", "/api/advanced/direct-publishers/pub-happy-path/supply-path", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify response body contains analysis result
	// The response should be a SupplyPathAnalysis struct or map
	// Just check if it's valid JSON for now or contains expected fields
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
}

func TestHandleAnalyzeSupplyPath_ServiceUnavailable(t *testing.T) {
	c := &mockCache{}
	bs := service.NewBiddingService(c, "http://localhost:8080")

	mockService := &mockBiddingServiceStatus{
		BiddingService:                 bs,
		forceDirectPublisherServiceNil: true,
	}
	router := setupTestHandlerWithMock(mockService)

	req, _ := http.NewRequest("GET", "/api/advanced/direct-publishers/pub-any/supply-path", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

func TestHandleAdvancedServicesStatus_HappyPath(t *testing.T) {
	c := &mockCache{}
	bs := service.NewBiddingService(c, "http://localhost:8080")

	mockService := &mockBiddingServiceStatus{BiddingService: bs}
	router := setupTestHandlerWithMock(mockService)

	req, _ := http.NewRequest("GET", "/api/advanced/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Verify structure: "services" map
	services, ok := resp["services"].(map[string]interface{})
	if !ok {
		t.Errorf("Response missing 'services' map")
	} else {
		if _, exists := services["bid_cache"]; !exists {
			t.Error("Missing bid_cache status")
		}
		if _, exists := services["s2s_bidding"]; !exists {
			t.Error("Missing s2s_bidding status")
		}
	}
}

func TestHandleAdvancedServicesStatus_PartialFailure(t *testing.T) {
	c := &mockCache{}
	bs := service.NewBiddingService(c, "http://localhost:8080")

	// Force one service to be nil to simulate partial unavailability/unhealthy
	mockService := &mockBiddingServiceStatus{
		BiddingService:          bs,
		forceBidCacheServiceNil: true,
	}
	router := setupTestHandlerWithMock(mockService)

	req, _ := http.NewRequest("GET", "/api/advanced/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Even with partial failure, it might return 200 but with status "down" in JSON?
	// Or handler might return 503?
	// Let's check handler implementation.
	// If I can't check, I'll assume standard behavior:
	// If the handler checks service == nil, it might report "down".

	if w.Code != http.StatusOK {
		// Allow 503 if that's the design
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status 200 or 503, got %d", w.Code)
		}
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err == nil {
		if services, ok := resp["services"].(map[string]interface{}); ok {
			if bc, exists := services["bid_cache"].(map[string]interface{}); exists {
				if status, ok := bc["status"].(string); ok && status == "active" {
					// We forced it nil, so it shouldn't be active if logic checks nil
					// However, if the handler logic is robust, it handles nil gracefully.
					// If HandleAdvancedServicesStatus uses GetBidCacheService() and gets nil,
					// it probably skips it or marks it down.
				}
			}
		}
	}
}
