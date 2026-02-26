package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ─────────────────────────────────────────────────────────────────────────────
// calculateLanguageMultiplier — additional branches (66.7% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestLanguage_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	result := s.calculateLanguageMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked with no language targeting")
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0, got %f", result.Multiplier)
	}
}

func TestLanguage_ExcludedUserLanguage(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.LanguageTargeting = &model.LanguageTargeting{
		ExcludeLanguages: []string{"zh"},
	}
	req := &model.BidRequest{
		User:    model.InternalUser{Language: "zh"},
		Context: map[string]interface{}{},
	}
	result := s.calculateLanguageMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded language 'zh'")
	}
	if result.Reason != "language_excluded:zh" {
		t.Errorf("expected 'language_excluded:zh', got '%s'", result.Reason)
	}
}

func TestLanguage_RequiredLanguageNotMatched(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.LanguageTargeting = &model.LanguageTargeting{
		Languages: []model.LanguageRule{
			{Code: "en", Required: true, Boost: 1.3},
		},
	}
	req := &model.BidRequest{
		User:    model.InternalUser{Language: "fr"},
		Context: map[string]interface{}{},
	}
	result := s.calculateLanguageMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for missing required language")
	}
	if result.Reason != "missing_required_language" {
		t.Errorf("expected 'missing_required_language', got '%s'", result.Reason)
	}
}

func TestLanguage_MatchedLanguageBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.LanguageTargeting = &model.LanguageTargeting{
		Languages: []model.LanguageRule{
			{Code: "en", Boost: 1.4},
		},
	}
	req := &model.BidRequest{
		User:    model.InternalUser{Language: "en-US"},
		Context: map[string]interface{}{},
	}
	result := s.calculateLanguageMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for matched language, reason: %s", result.Reason)
	}
	if !result.Matched {
		t.Error("expected matched=true for en-US matching 'en' code")
	}
	if result.Multiplier < 1.3 {
		t.Errorf("expected boost >= 1.3, got %f", result.Multiplier)
	}
}

func TestLanguage_DefaultMultiplier(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.LanguageTargeting = &model.LanguageTargeting{
		Languages: []model.LanguageRule{
			{Code: "es", Boost: 1.3},
		},
		DefaultMultiplier: 0.8,
	}
	req := &model.BidRequest{
		User:    model.InternalUser{Language: "fr"},
		Context: map[string]interface{}{},
	}
	result := s.calculateLanguageMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for unmatched language, reason: %s", result.Reason)
	}
	if result.Multiplier != 0.8 {
		t.Errorf("expected DefaultMultiplier=0.8, got %f", result.Multiplier)
	}
}

func TestLanguage_ContentLanguageExcluded(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.LanguageTargeting = &model.LanguageTargeting{
		ContentLanguage:  true,
		ExcludeLanguages: []string{"ar"},
	}
	req := &model.BidRequest{
		User: model.InternalUser{Language: "en"},
		Context: map[string]interface{}{
			"content_language": "ar",
		},
	}
	result := s.calculateLanguageMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded content language 'ar'")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateAdPositionMultiplier — additional branches (61.1% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestAdPosition_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	result := s.calculateAdPositionMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked with no ad position targeting")
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0, got %f", result.Multiplier)
	}
}

func TestAdPosition_AboveFoldOnly_BelowFold(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		AboveFoldOnly: true,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"ad_position": "below_fold",
			"above_fold":  false,
		},
	}
	result := s.calculateAdPositionMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for below-fold when AboveFoldOnly=true")
	}
	if result.Reason != "above_fold_only" {
		t.Errorf("expected 'above_fold_only', got '%s'", result.Reason)
	}
}

func TestAdPosition_ViewabilityBelowMinimum(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		MinViewability: 0.6,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"predicted_viewability": float64(0.4),
		},
	}
	result := s.calculateAdPositionMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for viewability below minimum")
	}
	if result.Reason != "viewability_below_minimum" {
		t.Errorf("expected 'viewability_below_minimum', got '%s'", result.Reason)
	}
}

func TestAdPosition_ExcludedPosition(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		ExcludePositions: []string{"footer"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"ad_position": "footer",
		},
	}
	result := s.calculateAdPositionMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded position 'footer'")
	}
	if result.Reason != "position_excluded:footer" {
		t.Errorf("expected 'position_excluded:footer', got '%s'", result.Reason)
	}
}

func TestAdPosition_AboveFoldBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		AboveFoldBoost: 1.3,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"ad_position": "above_fold",
			"above_fold":  true,
		},
	}
	result := s.calculateAdPositionMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for above-fold, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.2 {
		t.Errorf("expected AboveFoldBoost applied, got %f", result.Multiplier)
	}
}

func TestAdPosition_BelowFoldDiscount(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		BelowFoldDiscount: 0.7,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"ad_position": "below_fold",
			"above_fold":  false,
		},
	}
	result := s.calculateAdPositionMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for below-fold with discount, reason: %s", result.Reason)
	}
	if result.Multiplier > 0.8 {
		t.Errorf("expected BelowFoldDiscount=0.7 applied, got %f", result.Multiplier)
	}
}

func TestAdPosition_InterstitialBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		InterstitialBoost: 1.4,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"ad_position": "interstitial",
		},
	}
	result := s.calculateAdPositionMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for interstitial, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.3 {
		t.Errorf("expected InterstitialBoost applied, got %f", result.Multiplier)
	}
}

func TestAdPosition_StickyBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		StickyBoost: 1.2,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"ad_position": "sticky",
		},
	}
	result := s.calculateAdPositionMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for sticky, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.1 {
		t.Errorf("expected StickyBoost applied, got %f", result.Multiplier)
	}
}

func TestAdPosition_RequiredPositionNotMatched(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		Positions: []model.PositionRule{
			{Position: "above_fold", Required: true, Boost: 1.5},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"ad_position": "sidebar",
		},
	}
	result := s.calculateAdPositionMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked when required position not matched")
	}
	if result.Reason != "missing_required_position" {
		t.Errorf("expected 'missing_required_position', got '%s'", result.Reason)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateDealTargetingMultiplier — additional branches (56.3% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestDealTargeting_NoConfigNoDeals(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	result := s.calculateDealTargetingMultiplier(camp, newReq())
	if result.Blocked {
		t.Errorf("expected not blocked with no deal targeting, reason: %s", result.Reason)
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0, got %f", result.Multiplier)
	}
}

func TestDealTargeting_RequireDeal_NoDealsAvailable(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		RequireDeal: true,
	}
	result := s.calculateDealTargetingMultiplier(camp, newReq())
	if !result.Blocked {
		t.Error("expected blocked when RequireDeal=true but no deals available")
	}
	if result.Reason != "deal_required_but_none_available" {
		t.Errorf("expected 'deal_required_but_none_available', got '%s'", result.Reason)
	}
}

func TestDealTargeting_RequireDeal_FallbackToOpen(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DealTargeting = &model.DealTargeting{
		RequireDeal:    true,
		FallbackToOpen: true,
	}
	result := s.calculateDealTargetingMultiplier(camp, newReq())
	if result.Blocked {
		t.Errorf("expected not blocked when FallbackToOpen=true, reason: %s", result.Reason)
	}
	if result.DealType != "open" {
		t.Errorf("expected DealType='open', got '%s'", result.DealType)
	}
}

func TestDealTargeting_MatchedDeal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BidPrice = 5.0
	camp.Targeting.DealTargeting = &model.DealTargeting{}
	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Type: "mobile"},
		AdSlot:      model.AdSlot{ID: "slot-1", Formats: []string{"banner"}},
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal-123", BidFloor: 2.0},
			},
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked with matching deal, reason: %s", result.Reason)
	}
	if result.MatchedDealID != "deal-123" {
		t.Errorf("expected MatchedDealID='deal-123', got '%s'", result.MatchedDealID)
	}
}

func TestDealTargeting_PreferPGBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BidPrice = 5.0
	camp.Targeting.DealTargeting = &model.DealTargeting{
		PreferPG: true,
	}
	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Type: "mobile"},
		AdSlot:      model.AdSlot{ID: "slot-1", Formats: []string{"banner"}},
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal-pg", BidFloor: 1.0, At: 4}, // AT=4 = Fixed price (PG)
			},
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	// Just verify the function runs without error
	if result.Multiplier <= 0 {
		t.Error("expected positive multiplier")
	}
}

func TestDealTargeting_PreferredDealBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BidPrice = 5.0
	camp.Targeting.DealTargeting = &model.DealTargeting{
		PreferredDealIDs: []string{"deal-pref"},
	}
	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Type: "mobile"},
		AdSlot:      model.AdSlot{ID: "slot-1", Formats: []string{"banner"}},
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal-pref", BidFloor: 1.0},
			},
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.1 {
		t.Errorf("expected preferred deal boost, got %f", result.Multiplier)
	}
}

func TestDealTargeting_BidAdjustment(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BidPrice = 5.0
	camp.Targeting.DealTargeting = &model.DealTargeting{
		DealBidAdjustments: []model.DealBidAdjust{
			{DealID: "deal-adj", BidMultiplier: 1.5},
		},
	}
	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Type: "mobile"},
		AdSlot:      model.AdSlot{ID: "slot-1", Formats: []string{"banner"}},
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal-adj", BidFloor: 1.0},
			},
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.4 {
		t.Errorf("expected bid adjustment >= 1.5, got %f", result.Multiplier)
	}
}

func TestDealTargeting_LegacyDealID_Matched(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BidPrice = 5.0
	camp.DealID = "legacy-deal"
	// No DealTargeting configured → falls through to legacy path
	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
		Device:      model.InternalDevice{Type: "mobile"},
		AdSlot:      model.AdSlot{ID: "slot-1", Formats: []string{"banner"}},
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "legacy-deal", BidFloor: 2.0},
			},
		},
	}
	result := s.calculateDealTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for legacy deal match, reason: %s", result.Reason)
	}
	if result.MatchedDealID != "legacy-deal" {
		t.Errorf("expected legacy-deal matched, got '%s'", result.MatchedDealID)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CompetitiveIntelligence.calculateBidMultiplier — additional branches (50%)
// ─────────────────────────────────────────────────────────────────────────────

func TestCompetitive_DefensiveMode(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	// Record opponent wins to drive up competition
	req := &model.BidRequest{
		ID:          "req-1",
		PublisherID: "pub-1",
		AdSlot:      model.AdSlot{ID: "slot-1", Formats: []string{"banner"}},
	}
	for i := 0; i < 20; i++ {
		svc.RecordAuctionOutcome(req, 2.0, 3.0, false, "enemy")
	}

	camp := &model.Campaign{
		ID: "camp-def",
		Targeting: model.Targeting{
			CompetitiveIntelligence: &model.CompetitiveIntelligence{
				Enabled:          true,
				TrackCompetitors: []string{"enemy"},
				CompetitiveMode:  "defensive",
			},
		},
	}
	result := svc.AnalyzeCompetition(camp, req)
	if !result.Analyzed {
		t.Error("expected analysis to be performed")
	}
	// Defensive mode → multiplier reduced
	if result.BidAdjustment >= 1.0 {
		t.Errorf("expected bid adjustment < 1.0 for defensive mode, got %f", result.BidAdjustment)
	}
}

func TestCompetitive_BalancedMode(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	req := &model.BidRequest{
		ID:          "req-2",
		PublisherID: "pub-2",
		AdSlot:      model.AdSlot{ID: "slot-2", Formats: []string{"banner"}},
	}
	camp := &model.Campaign{
		ID: "camp-bal",
		Targeting: model.Targeting{
			CompetitiveIntelligence: &model.CompetitiveIntelligence{
				Enabled:         true,
				CompetitiveMode: "balanced",
			},
		},
	}
	result := svc.AnalyzeCompetition(camp, req)
	if !result.Analyzed {
		t.Error("expected analysis for balanced mode")
	}
	if result.BidAdjustment <= 0 {
		t.Error("expected positive bid adjustment")
	}
}

func TestCompetitive_MarketShareGoalBelow(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	req := &model.BidRequest{
		ID:          "req-3",
		PublisherID: "pub-3",
		AdSlot:      model.AdSlot{ID: "slot-3", Formats: []string{"banner"}},
	}
	// Record many losses to get low share of voice
	for i := 0; i < 50; i++ {
		svc.RecordAuctionOutcome(req, 1.0, 2.0, false, "dominant")
	}
	// Record a few wins
	for i := 0; i < 5; i++ {
		svc.RecordAuctionOutcome(req, 3.0, 2.0, true, "")
	}

	camp := &model.Campaign{
		ID: "camp-mkt",
		Targeting: model.Targeting{
			CompetitiveIntelligence: &model.CompetitiveIntelligence{
				Enabled:         true,
				CompetitiveMode: "balanced",
				MarketShareGoal: 0.5, // 50% goal — likely above current share
			},
		},
	}
	result := svc.AnalyzeCompetition(camp, req)
	if !result.Analyzed {
		t.Error("expected analysis performed")
	}
	if result.BidAdjustment <= 0 {
		t.Error("expected positive bid adjustment")
	}
}

func TestCompetitive_HighMarketCondition(t *testing.T) {
	svc := NewCompetitiveIntelligenceService(nil)

	req := &model.BidRequest{
		ID:          "req-4",
		PublisherID: "pub-4",
		AdSlot:      model.AdSlot{ID: "slot-4", Formats: []string{"banner"}},
	}

	// Record many competitor outcomes to simulate high competition
	for i := 0; i < 15; i++ {
		svc.RecordAuctionOutcome(req, 2.0, 3.5, false, "comp-a")
		svc.RecordAuctionOutcome(req, 2.5, 4.0, false, "comp-b")
	}

	camp := &model.Campaign{
		ID: "camp-high",
		Targeting: model.Targeting{
			CompetitiveIntelligence: &model.CompetitiveIntelligence{
				Enabled:          true,
				TrackCompetitors: []string{"comp-a", "comp-b"},
				CompetitiveMode:  "aggressive",
			},
		},
	}
	result := svc.AnalyzeCompetition(camp, req)
	if !result.Analyzed {
		t.Error("expected analysis performed")
	}
}
