package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// TestPMP_Support verifies Private Marketplace logic
func TestPMP_Support(t *testing.T) {
	// Setup Mock Services (Fraud, AI, Optimization)
	// These are required because ProcessBid calls them.
	// We can reuse helper functions or setup minimal stubs.

	// 1. Mock Fraud Service (Allow All)
	fraudServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(model.FraudCheckResponse{IsFraud: false})
	}))
	defer fraudServer.Close()

	// 2. Mock AI Service (No recommendations)
	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return struct wrapper
		json.NewEncoder(w).Encode(model.AIMatchResponse{
			Recommendations: []model.AIAdRecommendation{},
		})
	}))
	defer aiServer.Close()

	// 3. Mock Optimization Service (Pass-through)
	optServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Just echo back a small boost or default
		rec := model.BidRecommendation{
			RecommendedBid: 1.0, // Will be ignored if multiplier is 1.0
			BidMultiplier:  1.0,
		}
		json.NewEncoder(w).Encode(rec)
	}))
	defer optServer.Close()

	// Setup Bidding Service with Mock Cache
	mockCache := NewMockCache()
	svc := NewBiddingService(mockCache, "http://legacy-backend")

	// Configure service URLs
	svc.SetFraudServiceURL(fraudServer.URL)
	svc.SetAIServiceURL(aiServer.URL)
	// Note: SetOptimizationServiceURL might not be exposed, need check
	// Looking at previous code, it seems hardcoded or maybe exposed?
	// Assuming SetOptimizationServiceURL exists or we can patch the struct directly if exposed.
	// Let's assume standard NewBiddingService sets defaults.
	// To override optimization URL, we might need a setter or just rely on default failure behavior (which is fine).
	// Actually, ProcessBid calls optService. If it fails, it just uses base bid. That's OK.

	// --- Scenario 1: Private Auction (Must Match Deal) ---
	// Campaign A: No Deal ID
	// Campaign B: Deal ID "deal-123"
	// Request: Private Auction ("private_auction": 1), Deal ID "deal-123"

	// Setup Campaigns in Cache
	campA := &model.Campaign{
		ID:       "camp-open",
		Status:   "active",
		Budget:   1000.0,
		BidPrice: 5.0,
		Creative: model.Creative{Type: "banner"},
		Targeting: model.Targeting{
			Countries: []string{"US"},
		},
	}
	campB := &model.Campaign{
		ID:       "camp-deal",
		Status:   "active",
		BidPrice: 10.0,
		Budget:   1000.0, // Should be enough
		DealID:   "deal-123",
		Creative: model.Creative{Type: "banner"},
		Targeting: model.Targeting{
			Countries: []string{"US"},
		},
	}
	mockCache.SetActiveCampaigns([]*model.Campaign{campA, campB})

	// Create Request
	req := &model.BidRequest{
		ID:          "req-pmp-1",
		Timestamp:   time.Now(),
		PublisherID: "pub-1",
		User: model.InternalUser{
			Country: "US",
		},
		Device: model.InternalDevice{
			Type: "mobile",
			Geo:  model.InternalGeo{Country: "US"},
			IP:   "1.2.3.4",
		},
		AdSlot: model.AdSlot{
			Formats: []string{"banner"},
		},
		Pmp: &model.Pmp{
			PrivateAuction: 1, // Enforce Deals!
			Deals: []model.Deal{
				{ID: "deal-123", BidFloor: 8.0},
			},
		},
	}

	// Execute
	resp, err := svc.ProcessBid(req)
	if err != nil {
		t.Fatalf("ProcessBid failed: %v", err)
	}

	// Verify
	if resp.CampaignID != "camp-deal" {
		t.Errorf("Expected campaign 'camp-deal' (ID: %s), got: %s", "camp-deal", resp.CampaignID)
	}

	// --- Scenario 2: Private Auction Mismatch ---
	// Request only has "deal-999" (doesn't exist)
	reqMismatch := &model.BidRequest{
		ID:          "req-pmp-2",
		Timestamp:   time.Now(),
		PublisherID: "pub-1",
		Device: model.InternalDevice{
			Type: "mobile",
			Geo:  model.InternalGeo{Country: "US"},
			IP:   "1.2.3.4",
		},
		AdSlot: model.AdSlot{Formats: []string{"banner"}},
		Pmp: &model.Pmp{
			PrivateAuction: 1,
			Deals: []model.Deal{
				{ID: "deal-999", BidFloor: 5.0},
			},
		},
	}

	_, errMismatch := svc.ProcessBid(reqMismatch)
	if errMismatch == nil {
		t.Error("Expected error (no matching campaigns) for mismatched private auction deal, got success")
	}

	// --- Scenario 3: Preferred Deal (Optional) ---
	// Request is NOT private (private_auction: 0)
	// But it lists deal-123.
	// Both campaigns match.
	// But Campaign B matches the Preferred Deal.
	// Ideally, we should pick B because it often has higher bid price or priority.
	// Here, B (10.0) > A (5.0), so B wins naturally.
	// But let's lower B's price to 4.0 (below A) to see if Deal logic boosts it?
	// Currently, our logic only FILTERS. It doesn't BOOST.
	// So purely based on price, A wins (5.0 > 4.0).
	// UNLESS filtering applies:
	// "If campaign has deal ID, does it HAVE to be in the request?"
	// Yes, `match := false; for ... if match { break }; if !match { return false }`
	// So Campaign B (DealID="deal-123") matches the deal.
	// Campaign A (No DealID) passes through (DealID == "").
	// So both are valid candidates.

	// Let's modify B to 15.0 to ensure it wins.

	req2 := &model.BidRequest{
		ID:          "req-pmp-2",
		Timestamp:   time.Now(),
		PublisherID: "pub-1",
		User: model.InternalUser{
			Country: "US",
		},
		Device: model.InternalDevice{
			Type: "mobile",
			Geo:  model.InternalGeo{Country: "US"},
			IP:   "1.2.3.4",
		},
		AdSlot: model.AdSlot{
			Formats: []string{"banner"},
		},
		Pmp: &model.Pmp{
			PrivateAuction: 0, // Not exclusive
			Deals: []model.Deal{
				{ID: "deal-123", BidFloor: 8.0},
			},
		},
	}

	resp2, err := svc.ProcessBid(req2)
	if err != nil {
		t.Fatalf("ProcessBid scenario 2 failed: %v", err)
	}
	if resp2.CampaignID != "camp-deal" {
		t.Errorf("Expected campaign 'camp-deal', got '%s'", resp2.CampaignID)
	}
	if resp2.DealID != "deal-123" {
		t.Errorf("Expected DealID 'deal-123', got '%s'", resp2.DealID)
	}

	// --- Scenario 3: Open Market Request (No PMP) ---
	// Campaign A (Open) should match.
	// Campaign B (Deal 123) should theoretically match if we allow PMP campaigns in open market?
	// Current logic: If req.Pmp is nil, no PMP check runs. So Campaign B allows it.
	// This might be unintended. A campaign with a DealID usually shouldn't bid on open market requests.
	// But for now, let's just checking CampA matches.

	req3 := &model.BidRequest{
		ID:          "req-open",
		Timestamp:   time.Now(),
		PublisherID: "pub-1",
		User: model.InternalUser{
			Country: "US",
		},
		Device: model.InternalDevice{
			Type: "mobile",
			Geo:  model.InternalGeo{Country: "US"},
			IP:   "1.2.3.4",
		},
		AdSlot: model.AdSlot{
			Formats: []string{"banner"},
		},
		// No PMP
	}

	// Disable CampB to simplify assertion (or assert one of them wins)
	// Or check if CampA wins (since price 5 < 10, typically CampB would win if allowed)
	// Let's see behavior. logic says CampB DOES match.
	// If CampB matches, it wins (price 10).

	resp3, err := svc.ProcessBid(req3)
	if err != nil {
		t.Fatalf("ProcessBid scenario 3 failed: %v", err)
	}

	// If CampB matches, we should probably FIX the logic to prevent Deal campaigns from open market.
	// But let's first see what happens.
	if resp3.CampaignID != "camp-open" {
		t.Errorf("Expected 'camp-open', got '%s'", resp3.CampaignID)
	}
}
