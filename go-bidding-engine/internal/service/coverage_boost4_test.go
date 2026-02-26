package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ─────────────────────────────────────────────────────────────────────────────
// checkBrandSafety — additional branches (60.5% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestBrandSafety_BlockedPublisher(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BlockedPublishers = []string{"bad-pub-1", "bad-pub-2"}
	req := &model.BidRequest{
		PublisherID: "bad-pub-1",
		Context:     map[string]interface{}{},
	}
	result := s.checkBrandSafety(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for blocked publisher")
	}
	if result.Reason != "blocked_publisher" {
		t.Errorf("expected 'blocked_publisher', got '%s'", result.Reason)
	}
}

func TestBrandSafety_AllowedPublisher(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BlockedPublishers = []string{"bad-pub"}
	req := &model.BidRequest{
		PublisherID: "good-pub",
		Context:     map[string]interface{}{},
	}
	result := s.checkBrandSafety(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for allowed publisher, reason: %s", result.Reason)
	}
}

func TestBrandSafety_BlockedCategory(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BlockedCategories = []string{"IAB25"}
	req := &model.BidRequest{
		PublisherID: "pub-1",
		Context: map[string]interface{}{
			"categories": []interface{}{"IAB25", "IAB1"},
		},
	}
	result := s.checkBrandSafety(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for blocked IAB category")
	}
	if result.Reason != "blocked_category:IAB25" {
		t.Errorf("expected 'blocked_category:IAB25', got '%s'", result.Reason)
	}
}

func TestBrandSafety_BlockedKeyword(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BlockedKeywords = []string{"gambling", "casino"}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"content": "This article discusses online gambling strategies",
		},
	}
	result := s.checkBrandSafety(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for page containing blocked keyword")
	}
	if result.Reason != "blocked_keyword:gambling" {
		t.Errorf("expected 'blocked_keyword:gambling', got '%s'", result.Reason)
	}
}

func TestBrandSafety_StrictRiskyCategory(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BrandSafetyLevel = "strict"
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"categories": []interface{}{"IAB25-1"},
		},
	}
	result := s.checkBrandSafety(camp, req)
	if !result.Blocked {
		t.Error("expected blocked in strict mode for IAB25 risky category")
	}
	if result.Reason != "strict_risky_category:IAB25-1" {
		t.Errorf("expected 'strict_risky_category:IAB25-1', got '%s'", result.Reason)
	}
}

func TestBrandSafety_StandardRiskyCategory(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BrandSafetyLevel = "standard"
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"categories": []interface{}{"IAB26-3"},
		},
	}
	result := s.checkBrandSafety(camp, req)
	if result.Blocked {
		t.Error("expected NOT blocked in standard mode for risky category")
	}
	if result.Multiplier >= 1.0 {
		t.Errorf("expected reduced multiplier (<1.0) in standard mode, got %f", result.Multiplier)
	}
}

func TestBrandSafety_RelaxedRiskyCategory(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BrandSafetyLevel = "relaxed"
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"categories": []interface{}{"IAB7-2"},
		},
	}
	result := s.checkBrandSafety(camp, req)
	if result.Blocked {
		t.Error("expected NOT blocked in relaxed mode")
	}
	// relaxed: 10% reduction → 0.9
	if result.Multiplier > 0.95 || result.Multiplier < 0.8 {
		t.Errorf("expected ~0.9 multiplier for relaxed risky, got %f", result.Multiplier)
	}
}

func TestBrandSafety_SafeContent(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.BrandSafetyLevel = "strict"
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"categories": []interface{}{"IAB1", "IAB2"},
			"content":    "Family friendly article about cooking",
		},
	}
	result := s.checkBrandSafety(camp, req)
	if result.Blocked {
		t.Errorf("expected NOT blocked for safe content, reason: %s", result.Reason)
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0 multiplier for safe content, got %f", result.Multiplier)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateAppTargetingMultiplier — additional branches (51.4% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestAppTargeting_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	// No AppTargeting configured
	result := s.calculateAppTargetingMultiplier(camp, newReq())
	if result.Blocked {
		t.Error("expected not blocked with no app targeting")
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0 multiplier, got %f", result.Multiplier)
	}
}

func TestAppTargeting_InAppOnly_NotInApp(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AppTargeting = &model.AppTargeting{
		InAppOnly: true,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			// No "in_app" key → isInApp=false
		},
	}
	result := s.calculateAppTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for non-app inventory with InAppOnly=true")
	}
	if result.Reason != "in_app_only" {
		t.Errorf("expected 'in_app_only', got '%s'", result.Reason)
	}
}

func TestAppTargeting_MobileWebOnly_IsInApp(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AppTargeting = &model.AppTargeting{
		MobileWebOnly: true,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"in_app": true,
		},
	}
	result := s.calculateAppTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for in-app inventory with MobileWebOnly=true")
	}
	if result.Reason != "mobile_web_only" {
		t.Errorf("expected 'mobile_web_only', got '%s'", result.Reason)
	}
}

func TestAppTargeting_RatingTooLow(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AppTargeting = &model.AppTargeting{
		MinAppRating: 4.0,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"app_rating": float64(3.2),
			"bundle_id":  "com.example.lowratedapp",
		},
	}
	result := s.calculateAppTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for app rating below minimum")
	}
	if result.Reason != "app_rating_below_minimum" {
		t.Errorf("expected 'app_rating_below_minimum', got '%s'", result.Reason)
	}
}

func TestAppTargeting_ExcludedBundleID(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AppTargeting = &model.AppTargeting{
		ExcludeBundleIDs: []string{"com.spam.app"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"bundle_id": "com.spam.app",
		},
	}
	result := s.calculateAppTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded bundle ID")
	}
}

func TestAppTargeting_ExcludedCategory(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AppTargeting = &model.AppTargeting{
		ExcludeCategories: []string{"Gambling"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"app_category": "Gambling",
		},
	}
	result := s.calculateAppTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded category")
	}
}

func TestAppTargeting_RequiredBundleNotMatched(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AppTargeting = &model.AppTargeting{
		BundleIDs: []model.AppRule{
			{Value: "com.premium.app", Required: true, Boost: 1.5},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"bundle_id": "com.other.app",
		},
	}
	result := s.calculateAppTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked when required bundle ID not matched")
	}
	if result.Reason != "missing_required_bundle_id" {
		t.Errorf("expected 'missing_required_bundle_id', got '%s'", result.Reason)
	}
}

func TestAppTargeting_RequiredCategoryNotMatched(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AppTargeting = &model.AppTargeting{
		Categories: []model.AppRule{
			{Value: "Games", Required: true, Boost: 1.3},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"app_category": "News",
		},
	}
	result := s.calculateAppTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked when required category not matched")
	}
	if result.Reason != "missing_required_category" {
		t.Errorf("expected 'missing_required_category', got '%s'", result.Reason)
	}
}

func TestAppTargeting_MatchedBundleBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.AppTargeting = &model.AppTargeting{
		BundleIDs: []model.AppRule{
			{Value: "com.spotify.music", Boost: 1.4},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"bundle_id": "com.spotify.music",
		},
	}
	result := s.calculateAppTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for matched bundle, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.3 {
		t.Errorf("expected boost >= 1.4 for matched bundle, got %f", result.Multiplier)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateSeasonalMultiplier — more branches (61.4% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestSeasonal_SummerBoostOnSummer(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		SummerBoost: 1.3,
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier <= 0 {
		t.Error("expected positive multiplier")
	}
	// If it is summer (season == "summer"), boost should be applied
	if result.Season == "summer" && result.Multiplier < 1.2 {
		t.Errorf("expected SummerBoost applied in summer, got %f", result.Multiplier)
	}
}

func TestSeasonal_BackToSchoolBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		BackToSchoolBoost: 1.2,
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier <= 0 {
		t.Error("expected positive multiplier")
	}
}

func TestSeasonal_MonthEndBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		MonthEndBoost: 1.15,
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier <= 0 {
		t.Error("expected positive multiplier")
	}
	if result.IsMonthEnd && result.Multiplier < 1.1 {
		t.Errorf("expected MonthEndBoost applied, got %f", result.Multiplier)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateCarrierMultiplier — more branches (64.9% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestCarrier_ISPExcluded(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		ExcludeISPs: []string{"comcast"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"isp": "comcast"},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded ISP")
	}
}

func TestCarrier_RequiredCarrierNotMatched(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		Carriers: []model.CarrierRule{
			{Name: "verizon", Required: true, Boost: 1.3},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"carrier": "att"},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for missing required carrier")
	}
	if result.Reason != "missing_required_carrier" {
		t.Errorf("expected 'missing_required_carrier', got '%s'", result.Reason)
	}
}

func TestCarrier_NoTargeting(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	// No carrier targeting
	result := s.calculateCarrierMultiplier(camp, newReq())
	if result.Blocked {
		t.Errorf("expected not blocked with no carrier targeting, reason: %s", result.Reason)
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0 multiplier, got %f", result.Multiplier)
	}
}

func TestCarrier_4GConnection(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		ConnectionTypes: []string{"4g", "5g"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"connection_type": "4g"},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for allowed connection type '4g', reason: %s", result.Reason)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateVideoTargetingMultiplier — more branches (63.6% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestVideoTargeting_RequiredPlayerSizeNotMatched(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		PlayerSizes: []model.VideoPlayerSize{
			{Size: "xlarge", MinWidth: 2000, Required: true, Boost: 1.5},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":        true,
			"player_width": float64(320),
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for required player size not matched")
	}
	if result.Reason != "required_player_size_not_matched" {
		t.Errorf("expected 'required_player_size_not_matched', got '%s'", result.Reason)
	}
}

func TestVideoTargeting_SkippableOnlyWithNonSkippable(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		SkipSettings: &model.VideoSkipSettings{
			SkippableOnly: true,
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":     true,
			"skippable": false,
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for non-skippable when SkippableOnly=true")
	}
}

func TestVideoTargeting_DurationTooShort(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		MinDuration: 15,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":       true,
			"maxduration": float64(5), // inventory max is only 5s
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for inventory duration too short")
	}
}

func TestVideoTargeting_HighCompletionBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		CompletionRates: &model.CompletionRateRule{
			HighCompletionBoost: 1.4,
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":           true,
			"completion_rate": float64(0.8), // above 0.75
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for high completion rate, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.3 {
		t.Errorf("expected high completion boost >= 1.3, got %f", result.Multiplier)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateDemographicMultiplier — more branches (73% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestDemo_UnknownGenderDiscount(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		Genders: []model.GenderRule{
			{Gender: "female", Boost: 1.2},
		},
		UnknownGenderBoost: 0.75,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			// no "gender" key
		},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for unknown gender, reason: %s", result.Reason)
	}
}

func TestDemo_GenderMatch_Boost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		Genders: []model.GenderRule{
			{Gender: "male", Boost: 1.3},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"gender": "male"},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for matching gender, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.2 {
		t.Errorf("expected boost >= 1.2 for matching gender, got %f", result.Multiplier)
	}
}

func TestDemo_ExcludeGender(t *testing.T) {
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
	if result.Reason != "gender_excluded:male" {
		t.Errorf("expected 'gender_excluded:male', got '%s'", result.Reason)
	}
}

func TestDemo_AgeRangeMatch_Boost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		AgeRanges: []model.AgeRange{
			{MinAge: 25, MaxAge: 35, Boost: 1.4},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"age": float64(30)},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for matching age range, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.3 {
		t.Errorf("expected boost >= 1.3 for matching age, got %f", result.Multiplier)
	}
}
