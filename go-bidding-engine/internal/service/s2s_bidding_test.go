package service

import (
	"context"
	"sync"
	"testing"
	"time"
)

// ========== Helper Functions ==========

func createS2SService() *S2SBiddingService {
	return NewS2SBiddingService(nil)
}

func createTestPartner(id string) *DemandPartner {
	return &DemandPartner{
		ID:       id,
		Name:     "Test Partner " + id,
		Endpoint: "https://partner-" + id + ".example.com/bid",
		Enabled:  true,
		Timeout:  100 * time.Millisecond,
		QPS:      1000,
		Headers:  map[string]string{"X-API-Key": "test-key"},
		BidFloor: 0.5,
	}
}

func createS2SBidRequest() *S2SBidRequest {
	return &S2SBidRequest{
		ID: "s2s-req-001",
		Imp: []S2SImpression{
			{
				ID:       "imp-001",
				BidFloor: 0.25,
				Currency: "USD",
				Banner: &S2SBanner{
					W: 300,
					H: 250,
				},
			},
		},
		Site: &S2SSite{
			ID:     "site-001",
			Domain: "example.com",
			Page:   "https://example.com/article",
		},
		Device: &S2SDevice{
			UA:         "Mozilla/5.0",
			IP:         "192.168.1.1",
			OS:         "Windows",
			DeviceType: 2,
			Make:       "Dell",
			Model:      "XPS",
			Geo: &S2SGeo{
				Country: "US",
				Region:  "CA",
				City:    "San Francisco",
			},
		},
		User: &S2SUser{
			ID:       "user-001",
			BuyerUID: "buyer-001",
		},
		Timeout: 150,
	}
}

// ========== NewS2SBiddingService Tests ==========

func TestS2S_NewService_CreatesInstance(t *testing.T) {
	svc := NewS2SBiddingService(nil)

	if svc == nil {
		t.Fatal("Expected service to be created")
	}
	if svc.partners == nil {
		t.Error("Expected partners map to be initialized")
	}
	if svc.bidRequests == nil {
		t.Error("Expected bidRequests map to be initialized")
	}
	if svc.timeout == 0 {
		t.Error("Expected default timeout to be set")
	}
	if svc.maxPartners == 0 {
		t.Error("Expected maxPartners to be set")
	}
}

func TestS2S_NewService_DefaultTimeout(t *testing.T) {
	svc := createS2SService()

	expected := 200 * time.Millisecond
	if svc.timeout != expected {
		t.Errorf("Expected timeout %v, got %v", expected, svc.timeout)
	}
}

func TestS2S_NewService_HTTPClientInitialized(t *testing.T) {
	svc := createS2SService()

	if svc.httpClient == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

// ========== RegisterPartner Tests ==========

func TestS2S_RegisterPartner_Success(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")

	err := svc.RegisterPartner(partner)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	registered, _ := svc.GetPartner("partner-001")
	if registered == nil {
		t.Error("Expected partner to be registered")
	}
}

func TestS2S_RegisterPartner_MissingID(t *testing.T) {
	svc := createS2SService()
	partner := &DemandPartner{
		Endpoint: "https://example.com/bid",
	}

	err := svc.RegisterPartner(partner)

	if err == nil {
		t.Error("Expected error for missing ID")
	}
}

func TestS2S_RegisterPartner_MissingEndpoint(t *testing.T) {
	svc := createS2SService()
	partner := &DemandPartner{
		ID: "partner-001",
	}

	err := svc.RegisterPartner(partner)

	if err == nil {
		t.Error("Expected error for missing endpoint")
	}
}

func TestS2S_RegisterPartner_SetsDefaults(t *testing.T) {
	svc := createS2SService()
	partner := &DemandPartner{
		ID:       "partner-001",
		Endpoint: "https://example.com/bid",
	}

	err := svc.RegisterPartner(partner)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !partner.Enabled {
		t.Error("Expected partner to be enabled by default")
	}
	if partner.Timeout == 0 {
		t.Error("Expected default timeout to be set")
	}
	if partner.SuccessRate != 1.0 {
		t.Errorf("Expected success rate 1.0, got %v", partner.SuccessRate)
	}
	if partner.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestS2S_RegisterPartner_PreservesCustomTimeout(t *testing.T) {
	svc := createS2SService()
	partner := &DemandPartner{
		ID:       "partner-001",
		Endpoint: "https://example.com/bid",
		Timeout:  50 * time.Millisecond,
	}

	svc.RegisterPartner(partner)

	if partner.Timeout != 50*time.Millisecond {
		t.Errorf("Expected custom timeout to be preserved, got %v", partner.Timeout)
	}
}

// ========== GetPartner Tests ==========

func TestS2S_GetPartner_Success(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")
	svc.RegisterPartner(partner)

	retrieved, err := svc.GetPartner("partner-001")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if retrieved.ID != "partner-001" {
		t.Errorf("Expected partner ID partner-001, got %s", retrieved.ID)
	}
}

func TestS2S_GetPartner_NotFound(t *testing.T) {
	svc := createS2SService()

	_, err := svc.GetPartner("nonexistent")

	if err == nil {
		t.Error("Expected error for nonexistent partner")
	}
}

// ========== UpdatePartner Tests ==========

func TestS2S_UpdatePartner_Success(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")
	svc.RegisterPartner(partner)

	partner.Name = "Updated Partner"
	partner.BidFloor = 1.0
	err := svc.UpdatePartner(partner)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	updated, _ := svc.GetPartner("partner-001")
	if updated.Name != "Updated Partner" {
		t.Errorf("Expected updated name, got %s", updated.Name)
	}
	if updated.BidFloor != 1.0 {
		t.Errorf("Expected bid floor 1.0, got %v", updated.BidFloor)
	}
}

func TestS2S_UpdatePartner_NotFound(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("nonexistent")

	err := svc.UpdatePartner(partner)

	if err == nil {
		t.Error("Expected error for nonexistent partner")
	}
}

func TestS2S_UpdatePartner_SetsUpdatedAt(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")
	svc.RegisterPartner(partner)
	originalTime := partner.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	partner.Name = "Updated"
	svc.UpdatePartner(partner)

	if !partner.UpdatedAt.After(originalTime) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

// ========== RemovePartner Tests ==========

func TestS2S_RemovePartner_Success(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")
	svc.RegisterPartner(partner)

	err := svc.RemovePartner("partner-001")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	_, err = svc.GetPartner("partner-001")
	if err == nil {
		t.Error("Expected partner to be removed")
	}
}

func TestS2S_RemovePartner_NotFound(t *testing.T) {
	svc := createS2SService()

	err := svc.RemovePartner("nonexistent")

	if err == nil {
		t.Error("Expected error for nonexistent partner")
	}
}

// ========== ListPartners Tests ==========

func TestS2S_ListPartners_Empty(t *testing.T) {
	svc := createS2SService()

	partners := svc.ListPartners()

	if len(partners) != 0 {
		t.Errorf("Expected 0 partners, got %d", len(partners))
	}
}

func TestS2S_ListPartners_MultiplePartners(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	svc.RegisterPartner(createTestPartner("partner-002"))
	svc.RegisterPartner(createTestPartner("partner-003"))

	partners := svc.ListPartners()

	if len(partners) != 3 {
		t.Errorf("Expected 3 partners, got %d", len(partners))
	}
}

// ========== ProcessBidRequest Tests ==========

func TestS2S_ProcessBidRequest_Success(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	req := createS2SBidRequest()

	ctx := context.Background()
	resp, err := svc.ProcessBidRequest(ctx, req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response")
	}
	if resp.ID != req.ID {
		t.Errorf("Expected response ID %s, got %s", req.ID, resp.ID)
	}
}

func TestS2S_ProcessBidRequest_GeneratesID(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	req := &S2SBidRequest{
		Imp: []S2SImpression{{ID: "imp-001", Banner: &S2SBanner{W: 300, H: 250}}},
	}

	ctx := context.Background()
	resp, _ := svc.ProcessBidRequest(ctx, req)

	if resp.ID == "" {
		t.Error("Expected response ID to be generated")
	}
}

func TestS2S_ProcessBidRequest_NoPartners(t *testing.T) {
	svc := createS2SService()
	req := createS2SBidRequest()

	ctx := context.Background()
	_, err := svc.ProcessBidRequest(ctx, req)

	if err == nil {
		t.Error("Expected error when no partners available")
	}
}

func TestS2S_ProcessBidRequest_MultiplePartners(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	svc.RegisterPartner(createTestPartner("partner-002"))
	svc.RegisterPartner(createTestPartner("partner-003"))
	req := createS2SBidRequest()

	ctx := context.Background()
	resp, err := svc.ProcessBidRequest(ctx, req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(resp.PartnerBids) != 3 {
		t.Errorf("Expected 3 partner bids, got %d", len(resp.PartnerBids))
	}
}

func TestS2S_ProcessBidRequest_FilterByPartnerIDs(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	svc.RegisterPartner(createTestPartner("partner-002"))
	svc.RegisterPartner(createTestPartner("partner-003"))

	req := createS2SBidRequest()
	req.PartnerIDs = []string{"partner-001", "partner-003"}

	ctx := context.Background()
	resp, _ := svc.ProcessBidRequest(ctx, req)

	if len(resp.PartnerBids) != 2 {
		t.Errorf("Expected 2 partner bids, got %d", len(resp.PartnerBids))
	}
}

func TestS2S_ProcessBidRequest_RecordsLatency(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	req := createS2SBidRequest()

	ctx := context.Background()
	resp, _ := svc.ProcessBidRequest(ctx, req)

	if resp.Latency < 0 {
		t.Error("Expected non-negative latency")
	}
}

func TestS2S_ProcessBidRequest_IncrementCounters(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	req := createS2SBidRequest()

	ctx := context.Background()
	svc.ProcessBidRequest(ctx, req)

	stats := svc.GetStats()
	if stats["total_requests"].(int64) != 1 {
		t.Error("Expected request counter to increment")
	}
}

func TestS2S_ProcessBidRequest_CustomTimeout(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))

	req := createS2SBidRequest()
	req.Timeout = 50 // 50ms custom timeout

	ctx := context.Background()
	resp, err := svc.ProcessBidRequest(ctx, req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp == nil {
		t.Error("Expected response")
	}
}

// ========== SelectWinningBid Tests ==========

func TestS2S_SelectWinningBid_Success(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))

	response := &S2SBidResponse{
		ID: "resp-001",
		SeatBid: []S2SSeatBid{
			{
				Seat: "partner-001",
				Bid: []S2SBid{
					{ID: "bid-1", Price: 1.50, ImpID: "imp-001"},
				},
			},
			{
				Seat: "partner-002",
				Bid: []S2SBid{
					{ID: "bid-2", Price: 2.00, ImpID: "imp-001"},
				},
			},
		},
	}

	winner, seat, err := svc.SelectWinningBid(response)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if winner.Price != 2.00 {
		t.Errorf("Expected winning price 2.00, got %v", winner.Price)
	}
	if seat != "partner-002" {
		t.Errorf("Expected winning seat partner-002, got %s", seat)
	}
}

func TestS2S_SelectWinningBid_NoBids(t *testing.T) {
	svc := createS2SService()
	response := &S2SBidResponse{
		ID:      "resp-001",
		SeatBid: []S2SSeatBid{},
	}

	_, _, err := svc.SelectWinningBid(response)

	if err == nil {
		t.Error("Expected error when no bids")
	}
}

func TestS2S_SelectWinningBid_UpdatesWinStats(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")
	svc.RegisterPartner(partner)

	response := &S2SBidResponse{
		ID: "resp-001",
		SeatBid: []S2SSeatBid{
			{
				Seat: "partner-001",
				Bid: []S2SBid{
					{ID: "bid-1", Price: 1.50, ImpID: "imp-001"},
				},
			},
		},
	}

	svc.SelectWinningBid(response)

	updated, _ := svc.GetPartner("partner-001")
	if updated.WonBids != 1 {
		t.Errorf("Expected won bids to be 1, got %d", updated.WonBids)
	}
}

func TestS2S_SelectWinningBid_MultipleBidsPerSeat(t *testing.T) {
	svc := createS2SService()

	response := &S2SBidResponse{
		ID: "resp-001",
		SeatBid: []S2SSeatBid{
			{
				Seat: "partner-001",
				Bid: []S2SBid{
					{ID: "bid-1", Price: 1.00, ImpID: "imp-001"},
					{ID: "bid-2", Price: 3.00, ImpID: "imp-002"},
					{ID: "bid-3", Price: 2.00, ImpID: "imp-003"},
				},
			},
		},
	}

	winner, _, _ := svc.SelectWinningBid(response)

	if winner.Price != 3.00 {
		t.Errorf("Expected highest bid 3.00, got %v", winner.Price)
	}
	if winner.ID != "bid-2" {
		t.Errorf("Expected bid-2, got %s", winner.ID)
	}
}

// ========== GetStats Tests ==========

func TestS2S_GetStats_Empty(t *testing.T) {
	svc := createS2SService()

	stats := svc.GetStats()

	if stats["total_requests"].(int64) != 0 {
		t.Error("Expected 0 total requests")
	}
	if stats["partner_count"].(int) != 0 {
		t.Error("Expected 0 partners")
	}
}

func TestS2S_GetStats_WithPartners(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	svc.RegisterPartner(createTestPartner("partner-002"))

	stats := svc.GetStats()

	if stats["partner_count"].(int) != 2 {
		t.Errorf("Expected 2 partners, got %d", stats["partner_count"])
	}

	partners := stats["partners"].([]map[string]interface{})
	if len(partners) != 2 {
		t.Errorf("Expected 2 partner stats, got %d", len(partners))
	}
}

func TestS2S_GetStats_TracksWinRate(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")
	partner.TotalBids = 100
	partner.WonBids = 25
	svc.RegisterPartner(partner)

	stats := svc.GetStats()
	partners := stats["partners"].([]map[string]interface{})

	for _, p := range partners {
		if p["id"] == "partner-001" {
			winRate := p["win_rate"].(float64)
			if winRate != 0.25 {
				t.Errorf("Expected win rate 0.25, got %v", winRate)
			}
		}
	}
}

// ========== SetTimeout Tests ==========

func TestS2S_SetTimeout_UpdatesValue(t *testing.T) {
	svc := createS2SService()
	newTimeout := 500 * time.Millisecond

	svc.SetTimeout(newTimeout)

	if svc.timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, svc.timeout)
	}
}

// ========== EnablePartner/DisablePartner Tests ==========

func TestS2S_EnablePartner_Success(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")
	partner.Enabled = false
	svc.RegisterPartner(partner)

	// Partner is enabled by default after register, so disable first
	svc.DisablePartner("partner-001")

	err := svc.EnablePartner("partner-001")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retrieved, _ := svc.GetPartner("partner-001")
	if !retrieved.Enabled {
		t.Error("Expected partner to be enabled")
	}
}

func TestS2S_EnablePartner_NotFound(t *testing.T) {
	svc := createS2SService()

	err := svc.EnablePartner("nonexistent")

	if err == nil {
		t.Error("Expected error for nonexistent partner")
	}
}

func TestS2S_DisablePartner_Success(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))

	err := svc.DisablePartner("partner-001")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retrieved, _ := svc.GetPartner("partner-001")
	if retrieved.Enabled {
		t.Error("Expected partner to be disabled")
	}
}

func TestS2S_DisablePartner_NotFound(t *testing.T) {
	svc := createS2SService()

	err := svc.DisablePartner("nonexistent")

	if err == nil {
		t.Error("Expected error for nonexistent partner")
	}
}

func TestS2S_DisablePartner_ExcludesFromBidding(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	svc.RegisterPartner(createTestPartner("partner-002"))
	svc.DisablePartner("partner-001")

	req := createS2SBidRequest()
	ctx := context.Background()
	resp, _ := svc.ProcessBidRequest(ctx, req)

	// Only partner-002 should respond
	if len(resp.PartnerBids) != 1 {
		t.Errorf("Expected 1 partner bid, got %d", len(resp.PartnerBids))
	}
}

// ========== getActivePartners Tests ==========

func TestS2S_GetActivePartners_ReturnsOnlyEnabled(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	svc.RegisterPartner(createTestPartner("partner-002"))
	svc.RegisterPartner(createTestPartner("partner-003"))
	svc.DisablePartner("partner-002")

	partners := svc.getActivePartners(nil)

	if len(partners) != 2 {
		t.Errorf("Expected 2 active partners, got %d", len(partners))
	}
}

func TestS2S_GetActivePartners_FiltersByID(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	svc.RegisterPartner(createTestPartner("partner-002"))
	svc.RegisterPartner(createTestPartner("partner-003"))

	partners := svc.getActivePartners([]string{"partner-001", "partner-003"})

	if len(partners) != 2 {
		t.Errorf("Expected 2 partners, got %d", len(partners))
	}
}

func TestS2S_GetActivePartners_IgnoresDisabledInFilter(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	svc.RegisterPartner(createTestPartner("partner-002"))
	svc.DisablePartner("partner-001")

	partners := svc.getActivePartners([]string{"partner-001", "partner-002"})

	if len(partners) != 1 {
		t.Errorf("Expected 1 active partner, got %d", len(partners))
	}
}

// ========== Mock Bid Generation Tests ==========

func TestS2S_GenerateMockBid_ReturnsValidBid(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")
	req := createS2SBidRequest()

	bid := svc.generateMockBid(partner, req)

	if bid == nil {
		t.Fatal("Expected bid to be generated")
	}
	if bid.ImpID != "imp-001" {
		t.Errorf("Expected imp ID imp-001, got %s", bid.ImpID)
	}
	if bid.Price <= 0 {
		t.Error("Expected positive bid price")
	}
}

func TestS2S_GenerateMockBid_NoImpressions(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")
	req := &S2SBidRequest{
		ID:  "req-001",
		Imp: []S2SImpression{},
	}

	bid := svc.generateMockBid(partner, req)

	if bid != nil {
		t.Error("Expected nil bid for empty impressions")
	}
}

func TestS2S_GenerateMockBid_UsesBidFloor(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")
	partner.BidFloor = 2.0
	req := createS2SBidRequest()

	bid := svc.generateMockBid(partner, req)

	if bid.Price < 2.0 {
		t.Errorf("Expected bid price >= floor 2.0, got %v", bid.Price)
	}
}

// ========== Concurrency Tests ==========

func TestS2S_ConcurrentPartnerRegistration(t *testing.T) {
	svc := createS2SService()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			partner := createTestPartner(string(rune('a'+idx%26)) + "-" + string(rune('0'+idx%10)))
			svc.RegisterPartner(partner)
		}(i)
	}

	wg.Wait()

	partners := svc.ListPartners()
	if len(partners) == 0 {
		t.Error("Expected partners to be registered")
	}
}

func TestS2S_ConcurrentBidRequests(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))
	svc.RegisterPartner(createTestPartner("partner-002"))

	var wg sync.WaitGroup
	errors := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := createS2SBidRequest()
			req.ID = "s2s-concurrent-" + string(rune('0'+idx%10))

			ctx := context.Background()
			_, err := svc.ProcessBidRequest(ctx, req)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent bid request failed: %v", err)
	}
}

func TestS2S_ConcurrentStatsAccess(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))

	var wg sync.WaitGroup

	// Concurrent stats reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = svc.GetStats()
		}()
	}

	// Concurrent bid requests
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := createS2SBidRequest()
			ctx := context.Background()
			svc.ProcessBidRequest(ctx, req)
		}()
	}

	wg.Wait()
}

// ========== S2S Request Types Tests ==========

func TestS2S_BidRequest_WithVideoImpression(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))

	req := &S2SBidRequest{
		ID: "video-req-001",
		Imp: []S2SImpression{
			{
				ID: "imp-video-001",
				Video: &S2SVideo{
					W:         640,
					H:         480,
					MIMEs:     []string{"video/mp4"},
					MinDur:    5,
					MaxDur:    30,
					Protocols: []int{2, 3},
				},
				BidFloor: 1.0,
			},
		},
	}

	ctx := context.Background()
	resp, err := svc.ProcessBidRequest(ctx, req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(resp.PartnerBids) == 0 {
		t.Error("Expected partner bids for video request")
	}
}

func TestS2S_BidRequest_WithNativeImpression(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))

	req := &S2SBidRequest{
		ID: "native-req-001",
		Imp: []S2SImpression{
			{
				ID: "imp-native-001",
				Native: &S2SNative{
					Request: `{"ver":"1.2","assets":[]}`,
					Ver:     "1.2",
				},
				BidFloor: 0.5,
			},
		},
	}

	ctx := context.Background()
	resp, err := svc.ProcessBidRequest(ctx, req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp == nil {
		t.Error("Expected response")
	}
}

func TestS2S_BidRequest_WithAppContext(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))

	req := &S2SBidRequest{
		ID: "app-req-001",
		Imp: []S2SImpression{
			{ID: "imp-001", Banner: &S2SBanner{W: 320, H: 50}},
		},
		App: &S2SApp{
			ID:       "app-001",
			Name:     "Test App",
			Bundle:   "com.example.app",
			StoreURL: "https://play.google.com/store/apps/details?id=com.example.app",
			Cat:      []string{"IAB9"},
		},
		Device: &S2SDevice{
			OS:         "Android",
			OSV:        "13",
			Make:       "Samsung",
			Model:      "Galaxy S23",
			DeviceType: 4,
			IFA:        "aaid-12345",
		},
	}

	ctx := context.Background()
	resp, err := svc.ProcessBidRequest(ctx, req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(resp.PartnerBids) == 0 {
		t.Error("Expected partner bids for app request")
	}
}

// ========== Partner Stats Tracking Tests ==========

func TestS2S_QueryPartner_UpdatesStats(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")
	svc.RegisterPartner(partner)

	req := createS2SBidRequest()
	ctx := context.Background()
	svc.ProcessBidRequest(ctx, req)

	updated, _ := svc.GetPartner("partner-001")
	if updated.TotalBids == 0 {
		t.Error("Expected total bids to be updated")
	}
}

func TestS2S_QueryPartner_TracksLatency(t *testing.T) {
	svc := createS2SService()
	partner := createTestPartner("partner-001")
	svc.RegisterPartner(partner)

	req := createS2SBidRequest()
	ctx := context.Background()

	// Make multiple requests
	for i := 0; i < 5; i++ {
		svc.ProcessBidRequest(ctx, req)
	}

	updated, _ := svc.GetPartner("partner-001")
	if updated.TotalBids != 5 {
		t.Errorf("Expected 5 total bids, got %d", updated.TotalBids)
	}
}

// ========== Edge Cases Tests ==========

func TestS2S_ProcessBidRequest_EmptyPartnerIDs(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))

	req := createS2SBidRequest()
	req.PartnerIDs = []string{} // Empty but not nil

	ctx := context.Background()
	resp, _ := svc.ProcessBidRequest(ctx, req)

	// Empty slice should return all partners
	if len(resp.PartnerBids) != 1 {
		t.Errorf("Expected 1 partner bid (all partners), got %d", len(resp.PartnerBids))
	}
}

func TestS2S_ProcessBidRequest_NonexistentPartnerIDs(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))

	req := createS2SBidRequest()
	req.PartnerIDs = []string{"nonexistent-001", "nonexistent-002"}

	ctx := context.Background()
	_, err := svc.ProcessBidRequest(ctx, req)

	if err == nil {
		t.Error("Expected error when all specified partners don't exist")
	}
}

func TestS2S_PartnerResponse_Currency(t *testing.T) {
	svc := createS2SService()
	svc.RegisterPartner(createTestPartner("partner-001"))

	req := createS2SBidRequest()
	ctx := context.Background()
	resp, _ := svc.ProcessBidRequest(ctx, req)

	if resp.Cur != "USD" {
		t.Errorf("Expected currency USD, got %s", resp.Cur)
	}
}
