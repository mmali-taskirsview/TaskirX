package service

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

func createCIRequest() *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-ci-1",
		PublisherID: "pub-123",
		AdSlot: model.AdSlot{
			ID:      "slot-1",
			Formats: []string{"banner"},
		},
	}
}

func createCICampaign(enabled bool) *model.Campaign {
	return &model.Campaign{
		ID:       "camp-ci-1",
		BidPrice: 2.0,
		Targeting: model.Targeting{
			CompetitiveIntelligence: &model.CompetitiveIntelligence{
				Enabled:         enabled,
				CompetitiveMode: "balanced",
			},
		},
	}
}

func TestCI_NewService(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.competitorData == nil {
		t.Error("expected competitorData to be initialized")
	}
	if svc.segmentBidFloors == nil {
		t.Error("expected segmentBidFloors to be initialized")
	}
}

func TestCI_AnalyzeCompetition_Disabled(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	campaign := createCICampaign(false)
	request := createCIRequest()

	result := svc.AnalyzeCompetition(campaign, request)

	if result.Analyzed {
		t.Error("expected Analyzed=false when disabled")
	}
	if result.BidAdjustment != 1.0 {
		t.Errorf("expected BidAdjustment=1.0, got %f", result.BidAdjustment)
	}
}

func TestCI_AnalyzeCompetition_Enabled(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	campaign := createCICampaign(true)
	request := createCIRequest()

	result := svc.AnalyzeCompetition(campaign, request)

	if !result.Analyzed {
		t.Error("expected Analyzed=true")
	}
	if result.BidAdjustment <= 0 {
		t.Error("expected positive BidAdjustment")
	}
}

func TestCI_RecordAuctionOutcome_Win(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	svc.RecordAuctionOutcome(request, 2.5, 2.5, true, "")

	report := svc.GetMarketReport()
	if report["total_auctions"].(int) != 1 {
		t.Errorf("expected 1 auction, got %d", report["total_auctions"].(int))
	}
}

func TestCI_RecordAuctionOutcome_Loss(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	svc.RecordAuctionOutcome(request, 2.0, 2.5, false, "competitor-1")

	profile, exists := svc.GetCompetitorProfile("competitor-1")
	if !exists {
		t.Error("expected competitor profile to exist")
	}
	if profile.BidVolume != 1 {
		t.Errorf("expected BidVolume=1, got %d", profile.BidVolume)
	}
}

func TestCI_RecordAuctionOutcome_UpdatesFloor(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	// Record a loss with winning bid
	svc.RecordAuctionOutcome(request, 2.0, 3.0, false, "")

	segmentKey := svc.getSegmentKey(request)
	floor := svc.GetSegmentFloor(segmentKey)

	if floor <= 0 {
		t.Error("expected floor to be set after loss")
	}
}

func TestCI_GetCompetitorProfile(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	// Build competitor profile through outcomes
	for i := 0; i < 5; i++ {
		svc.RecordAuctionOutcome(request, 2.0, 2.5+float64(i)*0.1, false, "comp-1")
	}

	profile, exists := svc.GetCompetitorProfile("comp-1")

	if !exists {
		t.Fatal("expected profile to exist")
	}
	if profile.BidVolume != 5 {
		t.Errorf("expected BidVolume=5, got %d", profile.BidVolume)
	}
	if profile.AvgBidPrice <= 0 {
		t.Error("expected positive AvgBidPrice")
	}
}

func TestCI_GetCompetitorProfile_NotFound(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	_, exists := svc.GetCompetitorProfile("nonexistent")
	if exists {
		t.Error("expected profile to not exist")
	}
}

func TestCI_GetSegmentFloor(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	// Win sets floor to our bid
	svc.RecordAuctionOutcome(request, 2.0, 2.0, true, "")

	segmentKey := svc.getSegmentKey(request)
	floor := svc.GetSegmentFloor(segmentKey)

	if floor != 2.0 {
		t.Errorf("expected floor=2.0, got %f", floor)
	}
}

func TestCI_GetSegmentFloor_NotFound(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	floor := svc.GetSegmentFloor("nonexistent")
	if floor != 0 {
		t.Errorf("expected floor=0 for unknown segment, got %f", floor)
	}
}

func TestCI_GetMarketReport(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	// Record some auctions
	svc.RecordAuctionOutcome(request, 2.0, 2.0, true, "")
	svc.RecordAuctionOutcome(request, 2.0, 2.5, false, "comp-1")
	svc.RecordAuctionOutcome(request, 2.0, 2.3, false, "comp-2")

	report := svc.GetMarketReport()

	if report["total_auctions"].(int) != 3 {
		t.Errorf("expected 3 auctions, got %d", report["total_auctions"].(int))
	}
	if report["tracked_competitors"].(int) != 2 {
		t.Errorf("expected 2 competitors, got %d", report["tracked_competitors"].(int))
	}
}

func TestCI_MarketShare(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	// 50% win rate
	svc.RecordAuctionOutcome(request, 2.0, 2.0, true, "")
	svc.RecordAuctionOutcome(request, 2.0, 2.5, false, "")
	svc.RecordAuctionOutcome(request, 2.5, 2.5, true, "")
	svc.RecordAuctionOutcome(request, 2.0, 2.3, false, "")

	report := svc.GetMarketReport()
	winRate := report["overall_win_rate"].(float64)

	if winRate != 0.5 {
		t.Errorf("expected win rate=0.5, got %f", winRate)
	}
}

func TestCI_CompetitiveModes(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	modes := []string{"aggressive", "defensive", "balanced"}

	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			campaign := &model.Campaign{
				Targeting: model.Targeting{
					CompetitiveIntelligence: &model.CompetitiveIntelligence{
						Enabled:         true,
						CompetitiveMode: mode,
					},
				},
			}

			result := svc.AnalyzeCompetition(campaign, request)

			if !result.Analyzed {
				t.Error("expected Analyzed=true")
			}
			if result.BidAdjustment <= 0 {
				t.Error("expected positive BidAdjustment")
			}
		})
	}
}

func TestCI_TrackCompetitors(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	// Build competitor data
	for i := 0; i < 10; i++ {
		svc.RecordAuctionOutcome(request, 2.0, 2.5, false, "tracked-comp")
	}

	campaign := &model.Campaign{
		Targeting: model.Targeting{
			CompetitiveIntelligence: &model.CompetitiveIntelligence{
				Enabled:          true,
				TrackCompetitors: []string{"tracked-comp"},
			},
		},
	}

	result := svc.AnalyzeCompetition(campaign, request)

	if len(result.CompetitorProfiles) == 0 {
		t.Error("expected competitor profiles in result")
	}
}

func TestCI_CompetitorTrends(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	// Increasing bids pattern
	for i := 0; i < 15; i++ {
		bid := 2.0 + float64(i)*0.1
		svc.RecordAuctionOutcome(request, 2.0, bid, false, "trending-comp")
	}

	profile, _ := svc.GetCompetitorProfile("trending-comp")

	if profile.TrendDirection != "increasing" {
		t.Errorf("expected trend=increasing, got %s", profile.TrendDirection)
	}
}

func TestCI_MarketShareGoal(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	// Record low win rate
	for i := 0; i < 10; i++ {
		won := i == 0 // Only 10% win rate
		svc.RecordAuctionOutcome(request, 2.0, 2.5, won, "")
	}

	campaign := &model.Campaign{
		Targeting: model.Targeting{
			CompetitiveIntelligence: &model.CompetitiveIntelligence{
				Enabled:         true,
				MarketShareGoal: 0.3, // Target 30% but only have 10%
			},
		},
	}

	result := svc.AnalyzeCompetition(campaign, request)

	// Should recommend increasing bids when below goal
	if result.BidAdjustment <= 1.0 {
		t.Error("expected bid adjustment > 1.0 when below market share goal")
	}
}

func TestCI_HistoryLimit(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	// Record more than max history
	for i := 0; i < 11000; i++ {
		svc.RecordAuctionOutcome(request, 2.0, 2.5, i%2 == 0, "")
	}

	report := svc.GetMarketReport()
	if report["total_auctions"].(int) > 10000 {
		t.Error("expected auction history to be capped at 10000")
	}
}

func TestCI_Concurrency(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()
	campaign := createCICampaign(true)
	done := make(chan bool, 3)

	// Writer
	go func() {
		for i := 0; i < 100; i++ {
			svc.RecordAuctionOutcome(request, 2.0, 2.5, i%2 == 0, "comp-"+string(rune('0'+i%5)))
		}
		done <- true
	}()

	// Reader 1
	go func() {
		for i := 0; i < 100; i++ {
			svc.AnalyzeCompetition(campaign, request)
		}
		done <- true
	}()

	// Reader 2
	go func() {
		for i := 0; i < 100; i++ {
			svc.GetMarketReport()
			svc.GetSegmentFloor("pub_123")
		}
		done <- true
	}()

	<-done
	<-done
	<-done
}

func TestCI_RecommendedActions(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	// Create high competition scenario
	for i := 0; i < 20; i++ {
		competitorID := "comp-" + string(rune('A'+i%6))
		svc.RecordAuctionOutcome(request, 2.0, 2.5, false, competitorID)
	}

	campaign := &model.Campaign{
		Targeting: model.Targeting{
			CompetitiveIntelligence: &model.CompetitiveIntelligence{
				Enabled: true,
			},
		},
	}

	result := svc.AnalyzeCompetition(campaign, request)

	// With many competitors, should recommend optimization
	if result.RecommendedAction == "" {
		t.Error("expected a recommended action")
	}
}

func TestCI_SegmentKeyGeneration(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	tests := []struct {
		name    string
		request *model.BidRequest
	}{
		{
			name: "with slot and format",
			request: &model.BidRequest{
				PublisherID: "pub-1",
				AdSlot:      model.AdSlot{ID: "slot-1", Formats: []string{"banner"}},
			},
		},
		{
			name: "without slot",
			request: &model.BidRequest{
				PublisherID: "pub-2",
				AdSlot:      model.AdSlot{Formats: []string{"video"}},
			},
		},
		{
			name: "minimal",
			request: &model.BidRequest{
				PublisherID: "pub-3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := svc.getSegmentKey(tt.request)
			if key == "" {
				t.Error("expected non-empty segment key")
			}
		})
	}
}

func TestCI_CompetitorProfileUpdate(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	// First bid
	svc.RecordAuctionOutcome(request, 2.0, 3.0, false, "comp-update")
	profile1, _ := svc.GetCompetitorProfile("comp-update")
	firstAvg := profile1.AvgBidPrice

	// Second bid at different price
	svc.RecordAuctionOutcome(request, 2.0, 4.0, false, "comp-update")
	profile2, _ := svc.GetCompetitorProfile("comp-update")

	if profile2.AvgBidPrice == firstAvg {
		t.Error("expected AvgBidPrice to update")
	}
	if profile2.BidVolume != 2 {
		t.Errorf("expected BidVolume=2, got %d", profile2.BidVolume)
	}
}

func TestCI_FloorSmoothing(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()
	segmentKey := svc.getSegmentKey(request)

	// First loss sets initial floor
	svc.RecordAuctionOutcome(request, 2.0, 3.0, false, "")
	floor1 := svc.GetSegmentFloor(segmentKey)

	// Second loss should smooth the floor
	svc.RecordAuctionOutcome(request, 2.0, 4.0, false, "")
	floor2 := svc.GetSegmentFloor(segmentKey)

	// Floor should be between old floor and new winning bid
	if floor2 <= floor1 || floor2 >= 4.0 {
		t.Errorf("expected smoothed floor between %f and 4.0, got %f", floor1, floor2)
	}
}

func TestCI_LastSeenUpdate(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)
	request := createCIRequest()

	svc.RecordAuctionOutcome(request, 2.0, 2.5, false, "time-comp")
	profile, _ := svc.GetCompetitorProfile("time-comp")

	if profile.LastSeen.IsZero() {
		t.Error("expected LastSeen to be set")
	}
	if time.Since(profile.LastSeen) > time.Second {
		t.Error("expected LastSeen to be recent")
	}
}
