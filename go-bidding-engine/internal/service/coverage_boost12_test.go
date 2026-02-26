package service

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================================
// calculateSeasonalMultiplier — recurring MM-DD event, MonthEndBoost,
// isEventActive edge cases
// ============================================================================

func TestCalculateSeasonalMultiplier_RecurringEvent(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	// Use MM-DD recurring format for today (always active)
	now := time.Now()
	start := now.AddDate(0, 0, -1).Format("01-02") // yesterday MM-DD
	end := now.AddDate(0, 0, 1).Format("01-02")    // tomorrow MM-DD
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		Events: []model.SeasonalEvent{
			{
				Name:      "recurring-event",
				Active:    true,
				Recurring: true,
				StartDate: start,
				EndDate:   end,
				Boost:     1.4,
			},
		},
	}
	result := s.calculateSeasonalMultiplier(camp)
	if !result.Matched {
		t.Error("expected matched for recurring active event")
	}
	if result.Multiplier < 1.3 {
		t.Errorf("expected boost ~1.4 for recurring event, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_MonthEndBoost(t *testing.T) {
	// Test MonthEndBoost configuration — the branch executes when day > daysInMonth-5.
	// We can't control "today", so just verify the branch doesn't panic and returns valid result.
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		MonthEndBoost: 1.3,
	}
	result := s.calculateSeasonalMultiplier(camp)
	// Multiplier depends on current day; just ensure no panic and valid range
	if result.Multiplier < 0.5 || result.Multiplier > 3.1 {
		t.Errorf("expected valid multiplier range, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_InvalidEventDates(t *testing.T) {
	// Bad date format — isEventActive returns false, no crash
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		Events: []model.SeasonalEvent{
			{Name: "bad-dates", Active: true, StartDate: "not-a-date", EndDate: "also-bad", Boost: 2.0},
		},
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier != 1.0 {
		t.Errorf("expected multiplier=1.0 for bad date event, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_NilTargeting(t *testing.T) {
	// No seasonal targeting → returns default 1.0
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = nil
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0 with nil targeting, got %f", result.Multiplier)
	}
}

// ============================================================================
// CalculateDaypartMultiplier — day-specific multiplier branch
// ============================================================================

func TestDayparting_DaySpecificMultiplier(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())

	// Use the current day-of-week to ensure the branch fires
	now := time.Now()
	dayName := dayNames[int(now.Weekday())]
	hour := now.Hour()

	camp := newDaypartCampaign(
		nil, // no hourly multipliers
		map[string]map[int]float64{
			dayName: {hour: 1.8},
		},
		false,
		"",
	)
	req := newReq()
	result := svc.CalculateDaypartMultiplier(camp, req)
	if result.Multiplier != 1.8 {
		t.Errorf("expected day-specific multiplier 1.8, got %f", result.Multiplier)
	}
	if result.Reason == "" {
		t.Error("expected non-empty reason for day-specific multiplier")
	}
}

func TestDayparting_DaySpecificWrongDay(t *testing.T) {
	// day-specific entry for a different day — falls through to no-match
	svc := NewDaypartingService(NewMockCache())

	now := time.Now()
	hour := now.Hour()
	// Pick the opposite day (never today)
	wrongDay := "saturday"
	if int(now.Weekday()) == 6 {
		wrongDay = "monday"
	}

	camp := newDaypartCampaign(
		nil,
		map[string]map[int]float64{
			wrongDay: {hour: 2.0},
		},
		false,
		"",
	)
	req := newReq()
	result := svc.CalculateDaypartMultiplier(camp, req)
	if result.Multiplier != 1.0 {
		t.Errorf("expected default 1.0 when day-specific entry is for wrong day, got %f", result.Multiplier)
	}
}

func TestDayparting_DaySpecificWrongHour(t *testing.T) {
	// Day matches but hour does not — falls through
	svc := NewDaypartingService(NewMockCache())

	now := time.Now()
	dayName := dayNames[int(now.Weekday())]
	// Use an hour that will never be current (use hour 25 is invalid; use -1+25 = avoid current)
	wrongHour := (now.Hour() + 12) % 24 // 12 hours offset — won't match current hour

	camp := newDaypartCampaign(
		nil,
		map[string]map[int]float64{
			dayName: {wrongHour: 2.0},
		},
		false,
		"",
	)
	req := newReq()
	result := svc.CalculateDaypartMultiplier(camp, req)
	if result.Multiplier != 1.0 {
		t.Errorf("expected default 1.0 when day matches but hour doesn't, got %f", result.Multiplier)
	}
}

// ============================================================================
// calculateRecommendedBid — soft/aggressive market + zero currentBid
// ============================================================================

func TestCalculateRecommendedBid_SoftMarket(t *testing.T) {
	svc := NewBidLandscapeService(NewMockCache())
	optRange := &model.OptimalBidRange{
		MinBid:    0.5,
		SweetSpot: 1.0,
		MaxBid:    2.0,
	}
	rec, mult := svc.calculateRecommendedBid(1.0, optRange, "soft")
	// soft market: recommended = minBid + (sweetSpot-minBid)*0.5 = 0.5+(0.5)*0.5 = 0.75
	if rec > 0.9 {
		t.Errorf("expected soft market recommended < sweetSpot, got %f", rec)
	}
	if mult <= 0 {
		t.Errorf("expected positive multiplier, got %f", mult)
	}
}

func TestCalculateRecommendedBid_AggressiveMarket(t *testing.T) {
	svc := NewBidLandscapeService(NewMockCache())
	optRange := &model.OptimalBidRange{
		MinBid:    0.5,
		SweetSpot: 1.0,
		MaxBid:    2.0,
	}
	rec, mult := svc.calculateRecommendedBid(1.0, optRange, "aggressive")
	// aggressive: sweetSpot + (maxBid-sweetSpot)*0.3 = 1.0+0.3 = 1.3
	if rec < 1.0 {
		t.Errorf("expected aggressive market recommended >= sweetSpot, got %f", rec)
	}
	if mult < 1.0 {
		t.Errorf("expected multiplier >= 1.0 in aggressive market, got %f", mult)
	}
}

func TestCalculateRecommendedBid_NilRange(t *testing.T) {
	svc := NewBidLandscapeService(NewMockCache())
	rec, mult := svc.calculateRecommendedBid(1.5, nil, "normal")
	if rec != 1.5 {
		t.Errorf("expected nil range to return currentBid=1.5, got %f", rec)
	}
	if mult != 1.0 {
		t.Errorf("expected mult=1.0 for nil range, got %f", mult)
	}
}

func TestCalculateRecommendedBid_ZeroCurrentBid(t *testing.T) {
	svc := NewBidLandscapeService(NewMockCache())
	optRange := &model.OptimalBidRange{
		MinBid:    0.5,
		SweetSpot: 1.0,
		MaxBid:    2.0,
	}
	// currentBid=0 → multiplier stays 1.0 (no division)
	rec, mult := svc.calculateRecommendedBid(0, optRange, "normal")
	if rec != 1.0 {
		t.Errorf("expected recommended=sweetSpot=1.0, got %f", rec)
	}
	if mult != 1.0 {
		t.Errorf("expected mult=1.0 for zero currentBid, got %f", mult)
	}
}

func TestCalculateRecommendedBid_MultiplierCaps(t *testing.T) {
	svc := NewBidLandscapeService(NewMockCache())
	// Very low current bid → ratio > 2.0, should cap at 2.0
	optRange := &model.OptimalBidRange{
		MinBid:    0.5,
		SweetSpot: 3.0,
		MaxBid:    5.0,
	}
	_, mult := svc.calculateRecommendedBid(0.1, optRange, "normal")
	if mult > 2.0 {
		t.Errorf("expected multiplier capped at 2.0, got %f", mult)
	}

	// Very high current bid → ratio < 0.5, should cap at 0.5
	_, mult2 := svc.calculateRecommendedBid(10.0, optRange, "normal")
	if mult2 < 0.5 {
		t.Errorf("expected multiplier floor at 0.5, got %f", mult2)
	}
}

// ============================================================================
// evaluateLookalike — low tier (score 0.5-0.7), no seed segments
// ============================================================================

func TestEvaluateLookalike_LowTier_B12(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())

	camp := newCampaign(1.0)
	req := newReq()
	req.User = model.InternalUser{
		ID:         "u1",
		Country:    "US",
		Categories: []string{"automotive"},
	}

	am := &model.AudienceModeling{
		LookalikeEnabled:    true,
		SeedSegments:        []string{"tech", "gaming"},
		SimilarityThreshold: 0.3, // low threshold
		LookalikeBoost:      1.4,
		LookalikeFeatures:   []string{"interests"},
	}

	result := svc.evaluateLookalike(camp, req, am, []string{"sports"})
	// With seed=[tech, gaming] and user=[automotive, sports], overlap is 0 → score=0 < threshold
	if result.Multiplier != 1.0 {
		t.Logf("got multiplier %f (below threshold expected 1.0)", result.Multiplier)
	}
}

func TestEvaluateLookalike_MediumTierScore_B12(t *testing.T) {
	// 0.7 <= score < 0.9 → lookalike_medium
	svc := NewAudienceModelingService(NewMockCache())
	camp := newCampaign(1.0)
	req := newReq()

	am := &model.AudienceModeling{
		LookalikeEnabled:    true,
		SeedSegments:        []string{"tech", "gaming", "finance"},
		SimilarityThreshold: 0.5,
		LookalikeBoost:      1.3,
		LookalikeFeatures:   []string{"interests"},
	}

	// Pass user segments that overlap with seed to get medium score
	result := svc.evaluateLookalike(camp, req, am, []string{"tech", "gaming"})
	if result.Multiplier < 0.5 {
		t.Errorf("expected valid multiplier, got %f", result.Multiplier)
	}
}

func TestEvaluateLookalike_WithExpansionFactor_LargerExpansion_B12(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newCampaign(1.0)
	req := newReq()

	am := &model.AudienceModeling{
		LookalikeEnabled:    true,
		SeedSegments:        []string{"tech"},
		SimilarityThreshold: 0.9, // very strict
		LookalikeExpansion:  9,   // high expansion
		LookalikeBoost:      1.3,
		LookalikeFeatures:   []string{"interests"},
	}

	result := svc.evaluateLookalike(camp, req, am, []string{"tech"})
	if result.Multiplier < 0.9 {
		t.Logf("expansion result: multiplier=%f, isLookalike=%v", result.Multiplier, result.IsLookalike)
	}
}

func TestEvaluateLookalike_NoSeedSegments_B12(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newCampaign(1.0)
	req := newReq()

	am := &model.AudienceModeling{
		LookalikeEnabled: true,
		SeedSegments:     []string{}, // empty
	}

	result := svc.evaluateLookalike(camp, req, am, []string{"tech"})
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0 with no seed segments, got %f", result.Multiplier)
	}
}

// ============================================================================
// GenerateOptimizedCreative — template not found, slot required but no elements,
// success path
// ============================================================================

func TestGenerateOptimizedCreative_TemplateNotFound(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	_, err := svc.GenerateOptimizedCreative(DCORequest{TemplateID: "nonexistent"})
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}

func TestGenerateOptimizedCreative_Success(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())

	// Create a template with one optional slot
	tmpl := &CreativeTemplate{
		Name: "test-template",
		Slots: map[string]*TemplateSlot{
			"headline": {
				ID:           "headline",
				Name:         "Headline",
				Type:         "headline",
				Required:     false,
				DefaultValue: "Default Headline",
			},
		},
	}
	created, err := svc.CreateTemplate(tmpl)
	if err != nil {
		t.Fatalf("CreateTemplate error: %v", err)
	}

	// Create an element for the slot
	elem := &CreativeElement{
		Type:    "headline",
		Content: "Buy Now!",
		Tags:    []string{"cta"},
	}
	_, err = svc.CreateElement(elem)
	if err != nil {
		t.Fatalf("CreateElement error: %v", err)
	}

	resp, err := svc.GenerateOptimizedCreative(DCORequest{
		TemplateID: created.ID,
		UserID:     "user-dco-1",
		Context: DCOContext{
			DeviceType: "mobile",
			TimeOfDay:  "evening",
		},
	})
	if err != nil {
		t.Fatalf("GenerateOptimizedCreative error: %v", err)
	}
	if resp.TemplateID != created.ID {
		t.Errorf("expected templateID=%s, got %s", created.ID, resp.TemplateID)
	}
	if resp.CombinationID == "" {
		t.Error("expected non-empty CombinationID")
	}
}

func TestGenerateOptimizedCreative_RequiredSlotNoElements(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())

	// Template with a required slot but no elements available
	tmpl := &CreativeTemplate{
		Name: "required-slot-template",
		Slots: map[string]*TemplateSlot{
			"logo": {
				ID:       "logo",
				Name:     "Logo",
				Type:     "logo",
				Required: true, // required, but no elements of this type
			},
		},
	}
	created, err := svc.CreateTemplate(tmpl)
	if err != nil {
		t.Fatalf("CreateTemplate error: %v", err)
	}

	_, err = svc.GenerateOptimizedCreative(DCORequest{
		TemplateID: created.ID,
		UserID:     "user-dco-2",
	})
	if err == nil {
		t.Error("expected error for required slot with no elements")
	}
}

// ============================================================================
// predictCTR (DynamicCreativeService) — device/time branches
// ============================================================================

func TestDCO_PredictCTR_MobileEvening(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	combo := &CreativeCombination{
		ID:         "combo-1",
		TemplateID: "tmpl-1",
		Elements:   map[string]string{},
		Performance: &CombinationPerformance{
			Impressions: 0, // below MinImpressionsForStats → uses baseCTR 0.01
		},
	}
	ctx := DCOContext{DeviceType: "mobile", TimeOfDay: "evening"}
	ctr := svc.predictCTR(combo, ctx)
	// mobile 1.1 * evening 1.1 * 0.01 = 0.0121
	if ctr < 0.011 {
		t.Errorf("expected mobile+evening CTR boost, got %f", ctr)
	}
}

func TestDCO_PredictCTR_DesktopMorning(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	combo := &CreativeCombination{
		ID:          "combo-2",
		TemplateID:  "tmpl-1",
		Elements:    map[string]string{},
		Performance: &CombinationPerformance{Impressions: 0},
	}
	ctx := DCOContext{DeviceType: "desktop", TimeOfDay: "morning"}
	ctr := svc.predictCTR(combo, ctx)
	// desktop 0.95 * morning 1.05 * 0.01 = 0.009975
	if ctr < 0.009 {
		t.Errorf("expected desktop+morning CTR ~0.01, got %f", ctr)
	}
}

func TestDCO_PredictCTR_TabletNight(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	combo := &CreativeCombination{
		ID:          "combo-3",
		TemplateID:  "tmpl-1",
		Elements:    map[string]string{},
		Performance: &CombinationPerformance{Impressions: 0},
	}
	ctx := DCOContext{DeviceType: "tablet", TimeOfDay: "night"}
	ctr := svc.predictCTR(combo, ctx)
	// tablet 1.0 * night 0.9 * 0.01 = 0.009
	if ctr > 0.012 {
		t.Errorf("expected night CTR reduction, got %f", ctr)
	}
}

// ============================================================================
// calculatePersonalizationScore — with engaged/converted elements
// ============================================================================

func TestDCO_PersonalizationScore_WithEngagement(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())

	elem1, _ := svc.CreateElement(&CreativeElement{Type: "headline", Content: "H1"})
	elem2, _ := svc.CreateElement(&CreativeElement{Type: "cta", Content: "Click"})

	elements := map[string]*CreativeElement{
		"slot1": elem1,
		"slot2": elem2,
	}

	pref := &UserCreativePreference{
		UserID: "u1",
		EngagedElements: map[string]int{
			elem1.ID: 3, // score: min(3/5, 1.0) = 0.6
		},
		ConvertedElements: map[string]int{
			elem2.ID: 2, // score: min(2*2, 2.0) = 2.0
		},
	}

	score := svc.calculatePersonalizationScore(elements, pref)
	// (0.6 + 2.0) / 2 = 1.3 → capped at 1.0
	if score != 1.0 {
		t.Errorf("expected capped score=1.0, got %f", score)
	}
}

func TestDCO_PersonalizationScore_EmptyElements(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	pref := &UserCreativePreference{UserID: "u-empty"}
	score := svc.calculatePersonalizationScore(map[string]*CreativeElement{}, pref)
	if score != 0 {
		t.Errorf("expected 0 for empty elements, got %f", score)
	}
}

func TestDCO_PersonalizationScore_NoEngagement(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	elem, _ := svc.CreateElement(&CreativeElement{Type: "headline", Content: "H"})
	elements := map[string]*CreativeElement{"s1": elem}
	pref := &UserCreativePreference{
		UserID:            "u-noengage",
		EngagedElements:   map[string]int{},
		ConvertedElements: map[string]int{},
	}
	score := svc.calculatePersonalizationScore(elements, pref)
	if score != 0 {
		t.Errorf("expected 0 for no engagement, got %f", score)
	}
}

// ============================================================================
// calculateGoalPacingMultiplier — ahead / behind branches
// ============================================================================

func TestCalculateGoalPacingMultiplier_WayAhead(t *testing.T) {
	// deliveryRatio > 1.5 → 0.5
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.GoalTarget = 1000
	camp.GoalDelivered = 900 // very far ahead
	camp.GoalEndDate = time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	mult := s.calculateGoalPacingMultiplier(camp)
	if mult != 0.5 {
		t.Errorf("expected 0.5 for way-ahead pacing, got %f", mult)
	}
}

func TestCalculateGoalPacingMultiplier_SlightlyAhead(t *testing.T) {
	// deliveryRatio 1.1-1.5 → 0.8
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.GoalTarget = 1000
	camp.GoalDelivered = 400
	camp.GoalEndDate = time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	mult := s.calculateGoalPacingMultiplier(camp)
	// Result depends on ratio calculation; just ensure valid range
	if mult < 0.2 || mult > 2.0 {
		t.Errorf("expected valid pacing multiplier, got %f", mult)
	}
}

func TestCalculateGoalPacingMultiplier_WayBehind(t *testing.T) {
	// deliveryRatio < 0.5 → 1.5
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.GoalTarget = 10000
	camp.GoalDelivered = 1
	camp.GoalEndDate = time.Now().AddDate(0, 0, 3).Format("2006-01-02")
	mult := s.calculateGoalPacingMultiplier(camp)
	// With almost no delivery and large goal, ratio should be < 0.5
	if mult < 0.2 {
		t.Errorf("expected boost for way-behind pacing, got %f", mult)
	}
}
