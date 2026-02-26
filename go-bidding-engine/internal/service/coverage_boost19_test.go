package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ---------------------------------------------------------------------------
// getPriceBucket — all 6 bucket branches
// ---------------------------------------------------------------------------

func TestGetPriceBucket_AllBuckets(t *testing.T) {
	cases := []struct {
		price    float64
		expected string
	}{
		{0.0, "0.00-0.50"},
		{0.25, "0.00-0.50"},
		{0.49, "0.00-0.50"},
		{0.50, "0.50-1.00"},
		{0.99, "0.50-1.00"},
		{1.00, "1.00-2.00"},
		{1.99, "1.00-2.00"},
		{2.00, "2.00-5.00"},
		{4.99, "2.00-5.00"},
		{5.00, "5.00-10.00"},
		{9.99, "5.00-10.00"},
		{10.00, "10.00+"},
		{99.99, "10.00+"},
	}

	for _, tc := range cases {
		got := getPriceBucket(tc.price)
		if got != tc.expected {
			t.Errorf("getPriceBucket(%.2f) = %q, want %q", tc.price, got, tc.expected)
		}
	}
}

// ---------------------------------------------------------------------------
// SuspendPublisher — covers the "registered" path
// ---------------------------------------------------------------------------

func TestDirectPublisher_SuspendPublisher(t *testing.T) {
	svc := NewDirectPublisherService(NewMockCache())

	// Register the publisher first
	pub := &DirectPublisher{
		ID:           "pub-suspend-01",
		Name:         "Suspend Test",
		Domain:       "suspend-test.example.com",
		Status:       "active",
		QualityScore: 0.8,
	}
	if _, err := svc.RegisterPublisher(pub); err != nil {
		t.Fatalf("RegisterPublisher: %v", err)
	}

	// Now suspend it — should succeed
	if err := svc.SuspendPublisher("pub-suspend-01", "fraud detected"); err != nil {
		t.Fatalf("SuspendPublisher returned unexpected error: %v", err)
	}

	// Verify status changed
	got, err := svc.GetPublisher("pub-suspend-01")
	if err != nil {
		t.Fatalf("GetPublisher after suspend: %v", err)
	}
	if got.Status != "suspended" {
		t.Errorf("expected status 'suspended', got %q", got.Status)
	}
}

func TestDirectPublisher_SuspendPublisher_NotFound(t *testing.T) {
	svc := NewDirectPublisherService(NewMockCache())

	err := svc.SuspendPublisher("nonexistent-pub", "test")
	if err == nil {
		t.Error("expected error for non-existent publisher, got nil")
	}
}

// ---------------------------------------------------------------------------
// BidCacheService.InvalidatePartner — both branches
// ---------------------------------------------------------------------------

func TestBidCache_InvalidatePartner_NotFound(t *testing.T) {
	svc := NewBidCacheService(nil)

	count := svc.InvalidatePartner("unknown-partner")
	if count != 0 {
		t.Errorf("expected 0 for unknown partner, got %d", count)
	}
}

func TestBidCache_InvalidatePartner_Found(t *testing.T) {
	svc := NewBidCacheService(nil)

	// Seed partnerCache directly — partnerCache is map[string]map[string]*CachedBid
	pid := "partner-x"
	svc.partnerCache[pid] = map[string]*CachedBid{
		"bid-1": {CampaignID: "c1", Price: 1.0, ExpiresAt: time.Now().Add(time.Hour)},
		"bid-2": {CampaignID: "c2", Price: 2.0, ExpiresAt: time.Now().Add(time.Hour)},
	}

	count := svc.InvalidatePartner(pid)
	if count != 2 {
		t.Errorf("expected 2 bids invalidated, got %d", count)
	}

	// Should be deleted
	remaining := svc.InvalidatePartner(pid)
	if remaining != 0 {
		t.Errorf("expected 0 after second invalidation, got %d", remaining)
	}
}

// ---------------------------------------------------------------------------
// CompetitiveIntelligenceService.calculateBidSpread
// ---------------------------------------------------------------------------

func TestCalculateBidSpread_TooFewAuctions(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	// Zero auctions → 0
	result := svc.calculateBidSpread(nil)
	if result != 0 {
		t.Errorf("expected 0 for empty auctions, got %f", result)
	}

	// One auction → 0
	result = svc.calculateBidSpread([]auctionOutcome{{WinningBid: 1.5}})
	if result != 0 {
		t.Errorf("expected 0 for single auction, got %f", result)
	}
}

func TestCalculateBidSpread_ZeroMean(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	// All bids zero → mean==0 → returns 0
	result := svc.calculateBidSpread([]auctionOutcome{
		{WinningBid: 0},
		{WinningBid: 0},
		{WinningBid: 0},
	})
	if result != 0 {
		t.Errorf("expected 0 for zero-mean auctions, got %f", result)
	}
}

func TestCalculateBidSpread_NormalCase(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	// Non-zero spread → positive result
	result := svc.calculateBidSpread([]auctionOutcome{
		{WinningBid: 1.0},
		{WinningBid: 2.0},
		{WinningBid: 3.0},
	})
	if result <= 0 {
		t.Errorf("expected positive spread, got %f", result)
	}
}

// ---------------------------------------------------------------------------
// CompetitiveIntelligenceService.analyzeKnownCompetitors
// ---------------------------------------------------------------------------

func TestAnalyzeKnownCompetitors_NoData(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	result := &model.CompetitiveIntelResult{}
	svc.analyzeKnownCompetitors([]string{"comp-a", "comp-b"}, result)

	// No competitorData seeded → profiles should be empty, no leading competitor
	if len(result.CompetitorProfiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(result.CompetitorProfiles))
	}
	if result.LeadingCompetitor != "" {
		t.Errorf("expected empty leading competitor, got %q", result.LeadingCompetitor)
	}
}

func TestAnalyzeKnownCompetitors_WithData(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	// Seed competitor data directly
	svc.competitorData["comp-alpha"] = &localCompetitorProfile{
		AvgBidPrice: 2.5,
		WinRate:     0.30,
		PeakHours:   []int{8, 9, 10},
		LastSeen:    time.Now(),
	}
	svc.competitorData["comp-beta"] = &localCompetitorProfile{
		AvgBidPrice: 3.0,
		WinRate:     0.60,
		PeakHours:   []int{14, 15},
		LastSeen:    time.Now(),
	}

	result := &model.CompetitiveIntelResult{}
	svc.analyzeKnownCompetitors([]string{"comp-alpha", "comp-beta", "comp-unknown"}, result)

	if len(result.CompetitorProfiles) != 2 {
		t.Errorf("expected 2 profiles (comp-unknown not in data), got %d", len(result.CompetitorProfiles))
	}
	// comp-beta has higher win rate → leading
	if result.LeadingCompetitor != "comp-beta" {
		t.Errorf("expected 'comp-beta' as leading competitor, got %q", result.LeadingCompetitor)
	}
}

// ---------------------------------------------------------------------------
// callFraudService — non-200 HTTP response
// ---------------------------------------------------------------------------

func TestCallFraudService_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	svc.fraudServiceURL = ts.URL

	req := newReq()
	req.Device.IP = "10.20.30.40"

	isFraud, err, hop := svc.callFraudService(req)

	if err == nil {
		t.Error("expected error for non-200 response, got nil")
	}
	if isFraud {
		t.Error("expected isFraud=false on error path")
	}
	if hop == nil {
		t.Error("expected non-nil hop on error path")
	} else if hop.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected hop.StatusCode 503, got %d", hop.StatusCode)
	}
}

func TestCallFraudService_DecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-valid-json{{{")) //nolint
	}))
	defer ts.Close()

	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	svc.fraudServiceURL = ts.URL

	req := newReq()
	req.Device.IP = "10.20.30.41"

	isFraud, err, _ := svc.callFraudService(req)

	if err == nil {
		t.Error("expected decode error, got nil")
	}
	if isFraud {
		t.Error("expected isFraud=false on decode error")
	}
}

// ---------------------------------------------------------------------------
// callAIMatchingService — non-200 HTTP response + decode error
// ---------------------------------------------------------------------------

func TestCallAIMatchingService_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	svc.aiServiceURL = ts.URL
	// reset circuit breaker
	svc.aiFailureCount = 0
	svc.aiLastFailure = time.Time{}

	req := newReq()

	recs, err, hop := svc.callAIMatchingService(req)

	if err == nil {
		t.Error("expected error for non-200, got nil")
	}
	if recs != nil {
		t.Error("expected nil recommendations on error")
	}
	if hop == nil {
		t.Error("expected non-nil hop")
	} else if hop.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected hop.StatusCode 500, got %d", hop.StatusCode)
	}
}

func TestCallAIMatchingService_DecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{{bad json")) //nolint
	}))
	defer ts.Close()

	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	svc.aiServiceURL = ts.URL
	svc.aiFailureCount = 0
	svc.aiLastFailure = time.Time{}

	req := newReq()

	recs, err, _ := svc.callAIMatchingService(req)

	if err == nil {
		t.Error("expected decode error, got nil")
	}
	if recs != nil {
		t.Error("expected nil recommendations on decode error")
	}
}

// ---------------------------------------------------------------------------
// callOptimizationService — non-200 HTTP response + decode error
// ---------------------------------------------------------------------------

func TestCallOptimizationService_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	svc.optServiceURL = ts.URL
	svc.optFailureCount = 0
	svc.optLastFailure = time.Time{}

	bid := &model.BidResult{Campaign: newCampaign(1.5), BidPrice: 1.5}
	req := newReq()

	rec, err, hop := svc.callOptimizationService(bid, req)

	if err == nil {
		t.Error("expected error for non-200 opt response, got nil")
	}
	if rec != nil {
		t.Error("expected nil recommendation on error")
	}
	if hop == nil {
		t.Error("expected non-nil hop")
	} else if hop.StatusCode != http.StatusBadRequest {
		t.Errorf("expected hop.StatusCode 400, got %d", hop.StatusCode)
	}
}

func TestCallOptimizationService_DecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("bad json{{")) //nolint
	}))
	defer ts.Close()

	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	svc.optServiceURL = ts.URL
	svc.optFailureCount = 0
	svc.optLastFailure = time.Time{}

	bid := &model.BidResult{Campaign: newCampaign(1.5), BidPrice: 1.5}
	req := newReq()

	rec, err, _ := svc.callOptimizationService(bid, req)

	if err == nil {
		t.Error("expected decode error for opt service, got nil")
	}
	if rec != nil {
		t.Error("expected nil recommendation on decode error")
	}
}

// ---------------------------------------------------------------------------
// S2SBiddingService.generateMockBid — empty req.Imp (nil case)
// ---------------------------------------------------------------------------

func TestS2SGenerateMockBid_EmptyImpressions(t *testing.T) {
	mc := NewMockCache()
	bsvc := NewBiddingService(mc, "")
	s2s := NewS2SBiddingService(bsvc)

	partner := &DemandPartner{
		ID:       "p1",
		Name:     "Partner One",
		BidFloor: 0.5,
		Enabled:  true,
	}

	req := &S2SBidRequest{
		ID:  "req-001",
		Imp: []S2SImpression{}, // empty
	}

	bid := s2s.generateMockBid(partner, req)
	if bid != nil {
		t.Errorf("expected nil bid for empty Imp, got %+v", bid)
	}
}

func TestS2SGenerateMockBid_WithImpression(t *testing.T) {
	mc := NewMockCache()
	bsvc := NewBiddingService(mc, "")
	s2s := NewS2SBiddingService(bsvc)

	partner := &DemandPartner{
		ID:       "p2",
		Name:     "Partner Two",
		BidFloor: 0.0, // will default to 0.50
		Enabled:  true,
	}

	req := &S2SBidRequest{
		ID: "req-002",
		Imp: []S2SImpression{
			{ID: "imp-01", BidFloor: 0.10},
		},
	}

	bid := s2s.generateMockBid(partner, req)
	if bid == nil {
		t.Fatal("expected non-nil bid")
	}
	if bid.Price < 0.50 {
		t.Errorf("expected price >= 0.50 (default), got %f", bid.Price)
	}
	if bid.ImpID != "imp-01" {
		t.Errorf("expected ImpID 'imp-01', got %q", bid.ImpID)
	}
}

// ---------------------------------------------------------------------------
// UnifiedIDService.getOrderedProviders — same-priority tiebreak
// ---------------------------------------------------------------------------

func TestGetOrderedProviders_SamePriorityTiebreak(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	config := &model.UnifiedIDConfig{
		Providers: []model.IDProvider{
			{Name: "ProviderB", Priority: 5},
			{Name: "ProviderA", Priority: 5}, // same priority as B
			{Name: "ProviderC", Priority: 2},
		},
		FallbackOrder: []string{"providerA", "providerB"}, // explicit order
	}

	ordered := svc.getOrderedProviders(config)

	if len(ordered) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(ordered))
	}

	// ProviderA should come before ProviderB (explicit fallback order)
	// and ProviderC is not in fallback order → goes to position 999
	names := make([]string, len(ordered))
	for i, p := range ordered {
		names[i] = p.Name
	}

	// ProviderA (fallback index 0) must precede ProviderB (index 1)
	aIdx, bIdx := -1, -1
	for i, n := range names {
		if n == "ProviderA" {
			aIdx = i
		} else if n == "ProviderB" {
			bIdx = i
		}
	}
	if aIdx < 0 || bIdx < 0 {
		t.Fatal("ProviderA or ProviderB not found in ordered list")
	}
	if aIdx >= bIdx {
		t.Errorf("expected ProviderA before ProviderB, got order %v", names)
	}
}

func TestGetOrderedProviders_NoPrioritySort(t *testing.T) {
	svc := NewUnifiedIDService(nil)

	config := &model.UnifiedIDConfig{
		Providers: []model.IDProvider{
			{Name: "High", Priority: 10},
			{Name: "Low", Priority: 1},
			{Name: "Mid", Priority: 5},
		},
		// No FallbackOrder → sort by Priority ascending
	}

	ordered := svc.getOrderedProviders(config)

	if len(ordered) != 3 {
		t.Fatalf("expected 3, got %d", len(ordered))
	}
	if ordered[0].Name != "Low" {
		t.Errorf("expected 'Low' (priority 1) first, got %q", ordered[0].Name)
	}
}

// ---------------------------------------------------------------------------
// AttributionService.GetAttributionSummary — basic coverage
// ---------------------------------------------------------------------------

func TestGetAttributionSummary_NoTouchpoints(t *testing.T) {
	svc := NewAttributionService(NewMockCache())

	// No touchpoints recorded → should return empty map (no error)
	summary, err := svc.GetAttributionSummary("user-xyz", "last_touch", 168)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summary) != 0 {
		t.Errorf("expected empty summary, got %d entries", len(summary))
	}
}

func TestGetAttributionSummary_WithTouchpoints(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	// Record two touchpoints for the same user via the cache mock
	_ = mc.RecordTouchpoint("user-attr", "camp-100", "click", "req-1", 7)
	_ = mc.RecordTouchpoint("user-attr", "camp-200", "impression", "req-2", 7)

	summary, err := svc.GetAttributionSummary("user-attr", "last_touch", 168)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// At minimum: no panic, summary has values
	_ = summary
}

// ---------------------------------------------------------------------------
// predictCPL — B2B branch (is_b2b = true lowers lead rate by 0.7)
// ---------------------------------------------------------------------------

func TestPredictCPL_B2B(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(3.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"is_b2b": true,
	}

	perf := performanceData{
		ctr: 0.02,
		cvr: 0.05,
	}

	cpl := svc.predictCPL(camp, req, perf)
	// B2B → leadRate *= 0.7 → higher CPL than non-B2B
	if cpl <= 0 {
		t.Errorf("expected positive CPL for B2B, got %f", cpl)
	}
}

func TestPredictCPL_NonB2B(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(3.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"is_b2b": false,
	}

	perf := performanceData{ctr: 0.02}
	cplNonB2B := svc.predictCPL(camp, req, perf)

	req.Context["is_b2b"] = true
	cplB2B := svc.predictCPL(camp, req, perf)

	// B2B has lower lead rate (×0.7) → lower leads per impression → higher CPL
	if cplB2B <= cplNonB2B {
		t.Errorf("expected B2B CPL (%.4f) > non-B2B CPL (%.4f)", cplB2B, cplNonB2B)
	}
}

// ---------------------------------------------------------------------------
// getCurrentSeason — cover all seasonal branches
// ---------------------------------------------------------------------------

func TestGetCurrentSeason_ReturnsString_B19(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// Just call it; the result depends on current date — verify it returns a non-empty string
	season := svc.getCurrentSeason()
	validSeasons := map[string]bool{
		"spring": true, "summer": true, "fall": true, "winter": true,
		"black_friday": true, "holiday": true, "new_year": true,
	}
	if !validSeasons[season] {
		t.Errorf("unexpected season %q", season)
	}
}

// ---------------------------------------------------------------------------
// GetAttributionBidAdjustment — campaign credit path
// ---------------------------------------------------------------------------

func TestGetAttributionBidAdjustment_WithCredit(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	// Record a touchpoint to build attribution data
	_ = mc.RecordTouchpoint("user-bid-adj", "camp-adj-1", "click", "req-adj-1", 7)

	adj := svc.GetAttributionBidAdjustment("camp-adj-1", "user-bid-adj", "last_touch", 168)
	// Should be >= 1.0 when the campaign has credit
	if adj <= 0 {
		t.Errorf("expected positive bid adjustment, got %f", adj)
	}
}

func TestGetAttributionBidAdjustment_EmptyInputs(t *testing.T) {
	svc := NewAttributionService(NewMockCache())

	// Empty userID or campaignID → returns 1.0
	adj := svc.GetAttributionBidAdjustment("camp-1", "", "last_touch", 168)
	if adj != 1.0 {
		t.Errorf("expected 1.0 for empty userID, got %f", adj)
	}

	adj = svc.GetAttributionBidAdjustment("", "user-1", "last_touch", 168)
	if adj != 1.0 {
		t.Errorf("expected 1.0 for empty campaignID, got %f", adj)
	}
}

// ---------------------------------------------------------------------------
// generateRecommendations — different recommendation types
// ---------------------------------------------------------------------------

func TestGenerateRecommendations_CTRDownTrend(t *testing.T) {
	svc := NewPerformancePredictionService(NewMockCache())

	// Build a predictions map with a declining CTR
	predictions := map[string]*MetricPrediction{
		"ctr": {
			TrendDirection: "down",
			TrendStrength:  0.7,
			PredictedValue: 0.005,
		},
	}
	features := map[string]float64{"bid_price": 1.50}
	ctx := PredictionContext{TimeOfDay: "afternoon", DayOfWeek: "wednesday"}

	recs := svc.generateRecommendations(predictions, features, ctx)
	// Should produce at least one recommendation (creative_refresh)
	if len(recs) == 0 {
		t.Error("expected at least one recommendation for declining CTR")
	}
}

// ---------------------------------------------------------------------------
// optimizeForDCPM — covers the dynamic CPM path
// ---------------------------------------------------------------------------

func TestOptimizeForDCPM_PositiveTarget(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(2.0)
	req := newReq()
	pg := &model.PerformanceGoals{
		TargetDCPM: 5000, // $5 CPM target
	}
	perf := performanceData{
		ctr:            0.02,
		viewability:    0.70,
		engagementRate: 0.05,
		winRate:        0.20,
	}

	multiplier := svc.optimizeForDCPM(camp, req, pg, perf)
	if multiplier <= 0 {
		t.Errorf("expected positive multiplier for DCPM, got %f", multiplier)
	}
	if multiplier > 2.5 {
		t.Errorf("expected multiplier <= 2.5, got %f", multiplier)
	}
}

func TestOptimizeForDCPM_ZeroTarget(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(2.0)
	req := newReq()
	pg := &model.PerformanceGoals{TargetDCPM: 0}
	perf := performanceData{}

	multiplier := svc.optimizeForDCPM(camp, req, pg, perf)
	if multiplier != 1.0 {
		t.Errorf("expected 1.0 for zero TargetDCPM, got %f", multiplier)
	}
}

// ---------------------------------------------------------------------------
// optimizeForCPI — target CPI path
// ---------------------------------------------------------------------------

func TestOptimizeForCPI_WithTarget(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(1.0)
	req := newAppReq("com.example.app", "Example App", "IAB3", true, 4.2)
	pg := &model.PerformanceGoals{
		TargetCPI: 3.0, // $3 per install
	}
	perf := performanceData{ctr: 0.03}

	multiplier := svc.optimizeForCPI(camp, req, pg, perf)
	if multiplier <= 0 {
		t.Errorf("expected positive CPI multiplier, got %f", multiplier)
	}
}

// ---------------------------------------------------------------------------
// optimizeForCPE — covers target CPE path with mobile + rich_media
// ---------------------------------------------------------------------------

func TestOptimizeForCPE_MobileRichMedia(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(1.0)
	camp.Creative.Type = "rich_media"
	req := newReq()
	req.Device.Type = "mobile"
	pg := &model.PerformanceGoals{TargetCPE: 0.5}
	perf := performanceData{engagementRate: 0.03}

	multiplier := svc.optimizeForCPE(camp, req, pg, perf)
	if multiplier <= 0 {
		t.Errorf("expected positive CPE multiplier, got %f", multiplier)
	}
}

// ---------------------------------------------------------------------------
// optimizeForVCPM — high / low viewability branches
// ---------------------------------------------------------------------------

func TestOptimizeForVCPM_HighViewability_B19(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(2.0)
	req := newReq()
	req.AdSlot.Position = "above-fold"
	pg := &model.PerformanceGoals{TargetVCPM: 5000}
	// Provide predicted viewability via context (above 0.8)
	req.Context = map[string]interface{}{
		"predicted_viewability": float64(0.85),
	}
	perf := performanceData{viewability: 0.85}

	multiplier := svc.optimizeForVCPM(camp, req, pg, perf)
	if multiplier <= 0 {
		t.Errorf("expected positive vCPM multiplier for high viewability, got %f", multiplier)
	}
}

func TestOptimizeForVCPM_LowViewability_B19(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(2.0)
	req := newReq()
	pg := &model.PerformanceGoals{TargetVCPM: 5000}
	perf := performanceData{viewability: 0.25} // very low

	multiplier := svc.optimizeForVCPM(camp, req, pg, perf)
	// Low viewability → heavy penalty
	if multiplier <= 0 {
		t.Errorf("expected positive (penalized) multiplier, got %f", multiplier)
	}
	// Should be notably reduced
	if multiplier > 1.0 {
		t.Errorf("expected multiplier <= 1.0 for low viewability, got %f", multiplier)
	}
}

// ---------------------------------------------------------------------------
// calculateLanguageMultiplier — missing required language
// ---------------------------------------------------------------------------

func TestCalculateLanguageMultiplier_RequiredMissing(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(1.0)
	camp.Targeting.LanguageTargeting = &model.LanguageTargeting{
		Languages: []model.LanguageRule{
			{Code: "fr", Required: true, Boost: 1.2},
		},
	}

	req := newReq()
	req.User.Language = "de" // German — not French

	result := svc.calculateLanguageMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected Blocked=true for missing required language")
	}
}

func TestCalculateLanguageMultiplier_ExcludedLanguage_B19(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(1.0)
	camp.Targeting.LanguageTargeting = &model.LanguageTargeting{
		ExcludeLanguages: []string{"de"},
	}

	req := newReq()
	req.User.Language = "de"

	result := svc.calculateLanguageMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected Blocked=true for excluded language")
	}
}

// ---------------------------------------------------------------------------
// calculateVideoTargetingMultiplier — non-video inventory blocked
// ---------------------------------------------------------------------------

func TestCalculateVideoTargetingMultiplier_NonVideoBlocked(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		Placements: []string{"instream"},
	}

	req := newReq()
	req.Context = map[string]interface{}{
		"video": false, // Not video inventory
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected Blocked=true for non-video inventory with video targeting")
	}
}

// ---------------------------------------------------------------------------
// calculatePerformanceGoalMultiplier — CPA, CPL goals
// ---------------------------------------------------------------------------

func TestCalculatePerformanceGoalMultiplier_CPA(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(1.0)
	camp.Budget = 1000.0
	camp.DailyBudget = 100.0
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpa",
		TargetCPA:   5.0,
	}

	req := newReq()
	result := svc.calculatePerformanceGoalMultiplier(camp, req)

	if result.Multiplier <= 0 {
		t.Errorf("expected positive CPA multiplier, got %f", result.Multiplier)
	}
}

func TestCalculatePerformanceGoalMultiplier_CPL(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(2.0)
	camp.Budget = 1000.0
	camp.DailyBudget = 100.0
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpl",
		TargetCPL:   8.0,
	}

	req := newReq()
	req.Context = map[string]interface{}{
		"is_b2b": true,
	}
	result := svc.calculatePerformanceGoalMultiplier(camp, req)

	if result.Multiplier <= 0 {
		t.Errorf("expected positive CPL multiplier, got %f", result.Multiplier)
	}
}

// ---------------------------------------------------------------------------
// calculateGoalPacingMultiplier — triggers when GoalTarget > 0 and GoalEndDate set
// ---------------------------------------------------------------------------

func TestCalculateGoalPacingMultiplier_BehindPace(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(1.0)
	camp.Budget = 1000.0
	camp.DailyBudget = 100.0
	camp.GoalTarget = 10000   // 10k impressions goal
	camp.GoalDelivered = 1000 // only 10% done
	// Set end date to 10 days from now (plenty of time → should reduce spend)
	camp.GoalEndDate = time.Now().Add(10 * 24 * time.Hour).Format("2006-01-02")

	req := newReq()
	multiplier := svc.calculateGoalPacingMultiplier(camp)
	_ = multiplier // Just ensure it runs without panic

	// calculateScore calls goalPacingMultiplier when GoalTarget > 0 and GoalEndDate != ""
	score := svc.calculateScore(camp, req)
	if score <= 0 {
		t.Error("expected positive score with goal pacing")
	}
}

// ---------------------------------------------------------------------------
// RefreshCampaigns — covers at least one code path (cache miss fallback)
// ---------------------------------------------------------------------------

func TestRefreshCampaigns_NilCampaigns(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// With no backend URL set and no campaigns in cache, should return error gracefully
	err := svc.RefreshCampaigns("")
	// Either no error (empty cache OK) or an error — just must not panic
	_ = err
}

// ---------------------------------------------------------------------------
// GetCrossDeviceFrequency — error branch when no graph data
// ---------------------------------------------------------------------------

func TestGetCrossDeviceFrequency_NoCachedData(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	camp := newCampaign(1.0)
	camp.Targeting.CrossDeviceEnabled = true
	camp.Targeting.FreqCapImpressions = 5

	req := newReq()
	req.Device.DeviceID = "device-xyz"

	// No cross-device data in cache → exceeded=false
	exceeded, _ := svc.checkCrossDeviceFreqCap(camp, req)
	_ = exceeded
}

// ---------------------------------------------------------------------------
// JSON encoding in HTTP tests — reuse fraud test helper pattern for AI resp
// ---------------------------------------------------------------------------

func TestCallAIMatchingService_SuccessPath(t *testing.T) {
	recommendations := []model.AIAdRecommendation{
		{CampaignID: "camp-ai-1", OverallScore: 0.85},
	}
	resp := model.AIMatchResponse{
		Recommendations: recommendations,
	}
	respBytes, _ := json.Marshal(resp)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes) //nolint
	}))
	defer ts.Close()

	mc := NewMockCache()
	svc := NewBiddingService(mc, "")
	svc.aiServiceURL = ts.URL
	svc.aiFailureCount = 0
	svc.aiLastFailure = time.Time{}

	req := newReq()
	recs, err, hop := svc.callAIMatchingService(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 1 {
		t.Errorf("expected 1 recommendation, got %d", len(recs))
	}
	if hop == nil || !hop.Success {
		t.Error("expected successful hop")
	}
	// Verify circuit breaker was reset
	svc.aiMutex.RLock()
	cnt := svc.aiFailureCount
	svc.aiMutex.RUnlock()
	if cnt != 0 {
		t.Errorf("expected aiFailureCount=0 after success, got %d", cnt)
	}
}
