package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================================
// calculateContextualMultiplier
// ============================================================================

func TestCalculateContextualMultiplier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	// No contextual keywords/categories
	result := s.calculateContextualMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked with no targeting")
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0, got %f", result.Multiplier)
	}
}

func TestCalculateContextualMultiplier_KeywordMatch(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.ContextualKeywords = []model.ContextualKeyword{
		{Keyword: "technology", Boost: 1.4},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"page_title": "Latest Technology News",
		},
	}
	result := s.calculateContextualMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.3 {
		t.Errorf("expected multiplier >= 1.3, got %f", result.Multiplier)
	}
	if len(result.MatchedKeywords) == 0 {
		t.Error("expected matched keyword")
	}
}

func TestCalculateContextualMultiplier_ExcludeWordBlock(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.ContextualExcludeWords = []string{"violence"}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"content": "article about violence in films",
		},
	}
	result := s.calculateContextualMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded keyword")
	}
}

func TestCalculateContextualMultiplier_ExactMatch(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.ContextualKeywords = []model.ContextualKeyword{
		{Keyword: "sports", Exact: true, Boost: 1.3},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"keywords": "sports news today"},
	}
	result := s.calculateContextualMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.0 {
		t.Errorf("expected boost for exact keyword match, got %f", result.Multiplier)
	}
}

func TestCalculateContextualMultiplier_CategoryMatch(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.ContextualCategories = []string{"IAB19"}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"iab_categories": []interface{}{"IAB19-1", "IAB19-2"},
		},
	}
	result := s.calculateContextualMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.0 {
		t.Errorf("expected boost for category match, got %f", result.Multiplier)
	}
}

// ============================================================================
// calculateAudienceSegmentMultiplier
// ============================================================================

func TestCalculateAudienceSegmentMultiplier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = nil
	result := s.calculateAudienceSegmentMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked with no targeting")
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0, got %f", result.Multiplier)
	}
}

func TestCalculateAudienceSegmentMultiplier_ExcludedSegment(t *testing.T) {
	mc := NewMockCache()
	mc.userSegments["user-1"] = []string{"seg-bad"}
	s := NewBiddingService(mc, "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-bad", Exclude: true},
	}
	req := &model.BidRequest{User: model.InternalUser{ID: "user-1"}}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded segment")
	}
}

func TestCalculateAudienceSegmentMultiplier_RequiredMissing(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-required", Required: true},
	}
	req := &model.BidRequest{User: model.InternalUser{ID: "user-no-segments"}}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for missing required segment")
	}
}

func TestCalculateAudienceSegmentMultiplier_FirstPartyBoost(t *testing.T) {
	mc := NewMockCache()
	mc.userSegments["user-vip"] = []string{"seg-vip"}
	s := NewBiddingService(mc, "")
	camp := newCampaign(1.0)
	camp.Targeting.AudienceSegments = []model.AudienceSegment{
		{SegmentID: "seg-vip", Source: "first_party"},
	}
	req := &model.BidRequest{User: model.InternalUser{ID: "user-vip"}}
	result := s.calculateAudienceSegmentMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.4 {
		t.Errorf("expected first-party boost >= 1.4, got %f", result.Multiplier)
	}
}

// ============================================================================
// calculateWeatherMultiplier
// ============================================================================

func TestCalculateWeatherMultiplier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.WeatherTargeting = nil
	result := s.calculateWeatherMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0, got %f", result.Multiplier)
	}
}

func TestCalculateWeatherMultiplier_TempBelowMin(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	minTemp := 15.0
	camp.Targeting.WeatherTargeting = &model.WeatherTargeting{
		TemperatureMin: &minTemp,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"temperature": 5.0},
	}
	result := s.calculateWeatherMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for temp below min")
	}
}

func TestCalculateWeatherMultiplier_TempAboveMax(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	maxTemp := 25.0
	camp.Targeting.WeatherTargeting = &model.WeatherTargeting{
		TemperatureMax: &maxTemp,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"temperature": 35.0},
	}
	result := s.calculateWeatherMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for temp above max")
	}
}

func TestCalculateWeatherMultiplier_ConditionMatch(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.WeatherTargeting = &model.WeatherTargeting{
		Conditions: []model.WeatherCondition{
			{Condition: "sunny", Boost: 1.4},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"weather": "sunny"},
	}
	result := s.calculateWeatherMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.3 {
		t.Errorf("expected boost >= 1.3, got %f", result.Multiplier)
	}
}

func TestCalculateWeatherMultiplier_RequiredMissing(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.WeatherTargeting = &model.WeatherTargeting{
		Conditions: []model.WeatherCondition{
			{Condition: "rainy", Required: true},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"weather": "sunny"},
	}
	result := s.calculateWeatherMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for missing required weather condition")
	}
}

func TestCalculateWeatherMultiplier_NoWeatherData(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.WeatherTargeting = &model.WeatherTargeting{
		Conditions: []model.WeatherCondition{
			{Condition: "sunny"},
		},
	}
	// No weather data in context
	result := s.calculateWeatherMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked when no weather data available")
	}
}

// ============================================================================
// calculatePOIMultiplier
// ============================================================================

func TestCalculatePOIMultiplier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.POITargeting = nil
	result := s.calculatePOIMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked")
	}
}

func TestCalculatePOIMultiplier_UserWithinRadius(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.POITargeting = &model.POITargeting{
		POIs: []model.POI{
			{Name: "Store A", Lat: 40.7128, Lon: -74.0060, Radius: 5.0, Boost: 1.5},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"lat": 40.7130,
			"lon": -74.0058,
		},
	}
	result := s.calculatePOIMultiplier(camp, req)
	if !result.Matched {
		t.Error("expected matched POI")
	}
	if result.Multiplier < 1.4 {
		t.Errorf("expected boost >= 1.4, got %f", result.Multiplier)
	}
}

func TestCalculatePOIMultiplier_NoLocation(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.POITargeting = &model.POITargeting{
		POIs: []model.POI{
			{Name: "Store B", Lat: 40.0, Lon: -74.0, Radius: 1.0},
		},
	}
	result := s.calculatePOIMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked when no location data")
	}
}

func TestCalculatePOIMultiplier_RequiredPOIMissed(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.POITargeting = &model.POITargeting{
		POIs: []model.POI{
			{Name: "Store C", Lat: 0.0, Lon: 0.0, Radius: 0.1, Required: true},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"lat": 40.0,
			"lon": -74.0,
		},
	}
	result := s.calculatePOIMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for required POI not within radius")
	}
}

// ============================================================================
// calculateCarrierMultiplier
// ============================================================================

func TestCalculateCarrierMultiplier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = nil
	result := s.calculateCarrierMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked")
	}
}

func TestCalculateCarrierMultiplier_CellularOnly_WifiBlocked(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{CellularOnly: true}
	req := &model.BidRequest{
		Context: map[string]interface{}{"connection_type": "wifi"},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for wifi when CellularOnly=true")
	}
}

func TestCalculateCarrierMultiplier_ExcludedCarrier(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		ExcludeCarriers: []string{"T-Mobile"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"carrier": "T-Mobile"},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded carrier")
	}
}

func TestCalculateCarrierMultiplier_AllowedCarrier(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		Carriers: []model.CarrierRule{
			{Name: "Verizon", Boost: 1.3},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"carrier": "Verizon"},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked for allowed carrier")
	}
	if result.Multiplier < 1.2 {
		t.Errorf("expected boost >= 1.2, got %f", result.Multiplier)
	}
}

// ============================================================================
// calculateLanguageMultiplier
// ============================================================================

func TestCalculateLanguageMultiplier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.LanguageTargeting = nil
	result := s.calculateLanguageMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked")
	}
}

func TestCalculateLanguageMultiplier_LanguageMatch(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.LanguageTargeting = &model.LanguageTargeting{
		Languages: []model.LanguageRule{
			{Code: "en", Boost: 1.2},
		},
	}
	req := &model.BidRequest{
		Device:  model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{"language": "en"},
	}
	result := s.calculateLanguageMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.1 {
		t.Errorf("expected boost, got %f", result.Multiplier)
	}
}

func TestCalculateLanguageMultiplier_ExcludedLanguage(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.LanguageTargeting = &model.LanguageTargeting{
		ExcludeLanguages: []string{"zh"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"language": "zh"},
	}
	result := s.calculateLanguageMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded language")
	}
}

// ============================================================================
// calculateDayOfWeekMultiplier
// ============================================================================

func TestCalculateDayOfWeekMultiplier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DayOfWeekTargeting = nil
	result := s.calculateDayOfWeekMultiplier(camp)
	if !result.Allowed {
		t.Error("expected allowed with no targeting")
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0, got %f", result.Multiplier)
	}
}

func TestCalculateDayOfWeekMultiplier_AllDaysAllowed(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		// All 7 days active
		Days: []model.DaySchedule{
			{Day: 0, Active: true, Boost: 1.0},
			{Day: 1, Active: true, Boost: 1.1},
			{Day: 2, Active: true, Boost: 1.1},
			{Day: 3, Active: true, Boost: 1.1},
			{Day: 4, Active: true, Boost: 1.3},
			{Day: 5, Active: true, Boost: 1.2},
			{Day: 6, Active: true, Boost: 1.0},
		},
	}
	result := s.calculateDayOfWeekMultiplier(camp)
	if !result.Allowed {
		t.Errorf("expected allowed (all days active), got reason: %s", result.Reason)
	}
}

// ============================================================================
// calculateAdPositionMultiplier
// ============================================================================

func TestCalculateAdPositionMultiplier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AdPositionTargeting = nil
	result := s.calculateAdPositionMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked")
	}
}

func TestCalculateAdPositionMultiplier_AboveFold(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		Positions: []model.PositionRule{
			{Position: "above_fold", Boost: 1.4},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"ad_position": "above_fold"},
	}
	result := s.calculateAdPositionMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.3 {
		t.Errorf("expected boost >= 1.3 for above_fold, got %f", result.Multiplier)
	}
}

func TestCalculateAdPositionMultiplier_ExcludedPosition(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		ExcludePositions: []string{"below_fold"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"ad_position": "below_fold"},
	}
	result := s.calculateAdPositionMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded position")
	}
}

// ============================================================================
// calculateAppTargetingMultiplier
// ============================================================================

func TestCalculateAppTargetingMultiplier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AppTargeting = nil
	result := s.calculateAppTargetingMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked")
	}
}

func TestCalculateAppTargetingMultiplier_AllowedBundle(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AppTargeting = &model.AppTargeting{
		BundleIDs: []model.AppRule{
			{Value: "com.example.app", Boost: 1.2},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"bundle": "com.example.app"},
	}
	result := s.calculateAppTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked for allowed bundle")
	}
	if result.Matched != true {
		t.Error("expected matched")
	}
}

func TestCalculateAppTargetingMultiplier_BlockedBundle(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AppTargeting = &model.AppTargeting{
		ExcludeBundleIDs: []string{"com.bad.app"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"bundle": "com.bad.app"},
	}
	result := s.calculateAppTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for blocked bundle")
	}
}

// ============================================================================
// calculateSeasonalMultiplier
// ============================================================================

func TestCalculateSeasonalMultiplier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = nil
	result := s.calculateSeasonalMultiplier(camp)
	if result.Matched {
		t.Error("expected not matched with no targeting")
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0, got %f", result.Multiplier)
	}
}

func TestCalculateSeasonalMultiplier_WithBoosts(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		HolidayBoost:   1.5,
		WeekendBoost:   1.2,
		Q4Boost:        1.3,
		SummerBoost:    1.2,
		EnableHolidays: true,
	}
	result := s.calculateSeasonalMultiplier(camp)
	// Should apply at least base multiplier without error
	if result.Multiplier < 1.0 {
		t.Errorf("expected multiplier >= 1.0, got %f", result.Multiplier)
	}
}

// ============================================================================
// calculateDemographicMultiplier
// ============================================================================

func TestCalculateDemographicMultiplier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = nil
	result := s.calculateDemographicMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked")
	}
}

func TestCalculateDemographicMultiplier_AgeInRange(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		AgeRanges: []model.AgeRange{
			{MinAge: 25, MaxAge: 45, Boost: 1.3},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"age": float64(30)},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.2 {
		t.Errorf("expected boost >= 1.2, got %f", result.Multiplier)
	}
}

func TestCalculateDemographicMultiplier_GenderMatch(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		Genders: []model.GenderRule{
			{Gender: "female", Boost: 1.2},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"gender": "female"},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
	if result.Multiplier < 1.1 {
		t.Errorf("expected boost >= 1.1, got %f", result.Multiplier)
	}
}

func TestCalculateDemographicMultiplier_ExcludedGender(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		ExcludeGenders: []string{"male"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"gender": "male"},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded gender")
	}
}

// ============================================================================
// calculateVideoTargetingMultiplier
// ============================================================================

func TestCalculateVideoTargetingMultiplier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = nil
	result := s.calculateVideoTargetingMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked")
	}
}

func TestCalculateVideoTargetingMultiplier_NonSkippable(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		SkipSettings: &model.VideoSkipSettings{
			NonSkippableOnly: true,
			NonSkipBoost:     1.3,
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"skippable": false},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked for non-skippable inventory when NonSkippableOnly=true")
	}
}

func TestCalculateVideoTargetingMultiplier_DurationFilter(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		MinDuration: 10,
		MaxDuration: 30,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":       true,
			"maxduration": float64(5), // inventory max duration below campaign min
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for video duration below minimum")
	}
}

func TestCalculateVideoTargetingMultiplier_InstreamPlacement(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		Placements: []string{"instream"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":           true,
			"video_placement": "instream",
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for matching instream placement, reason: %s", result.Reason)
	}
}

// ============================================================================
// calculatePerformanceGoalMultiplier
// ============================================================================

func TestCalculatePerformanceGoalMultiplier_NilGoals(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = nil
	result := s.calculatePerformanceGoalMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked with nil PerformanceGoals")
	}
}

func TestCalculatePerformanceGoalMultiplier_WithCPAGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpa",
		TargetCPA:   5.0,
	}
	result := s.calculatePerformanceGoalMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked")
	}
}

func TestCalculatePerformanceGoalMultiplier_WithROASGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "roas",
		TargetROAS:  3.0,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"historical_roas": 4.0},
	}
	result := s.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Error("expected not blocked")
	}
}

// ============================================================================
// getHistoricalPerformance
// ============================================================================

func TestGetHistoricalPerformance_NoCacheData(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	perf := s.getHistoricalPerformance("camp-unknown", newReq())
	// Should return defaults / zeros
	if perf.ctr < 0 {
		t.Errorf("expected non-negative ctr, got %f", perf.ctr)
	}
}

func TestGetHistoricalPerformance_WithCTRandWinRate(t *testing.T) {
	mc := NewMockCache()
	// getHistoricalPerformance reads from cache.Get("perf:<id>") with "key:val,..." format
	mc.kv["perf:camp-1"] = "ctr:0.03,win_rate:0.20"
	s := NewBiddingService(mc, "")
	perf := s.getHistoricalPerformance("camp-1", newReq())
	if perf.ctr != 0.03 {
		t.Errorf("expected ctr=0.03, got %f", perf.ctr)
	}
	if perf.winRate != 0.20 {
		t.Errorf("expected winRate=0.20, got %f", perf.winRate)
	}
}
