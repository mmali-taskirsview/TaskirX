package service

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ─── calculateSeasonalMultiplier Tests ───────────────────────────────────────

func makeCampaignWithSeasonal_B28(st *model.SeasonalTargeting) *model.Campaign {
	return &model.Campaign{
		ID:       "camp-seasonal",
		BidPrice: 1.0,
		Targeting: model.Targeting{
			SeasonalTargeting: st,
		},
	}
}

func TestSeasonalMultiplier_NoConfig_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	campaign := &model.Campaign{ID: "c1", BidPrice: 1.0}
	result := svc.calculateSeasonalMultiplier(campaign)
	if result.Matched || result.Multiplier != 1.0 {
		t.Errorf("expected no match and 1.0 multiplier, got matched=%v mult=%f", result.Matched, result.Multiplier)
	}
}

func TestSeasonalMultiplier_WeekendBoost_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	st := &model.SeasonalTargeting{WeekendBoost: 1.5}
	campaign := makeCampaignWithSeasonal_B28(st)
	result := svc.calculateSeasonalMultiplier(campaign)
	// Test runs any day; only assert multiplier is >=1.0 and that the config is honored
	if result.Multiplier < 1.0 {
		t.Errorf("expected multiplier >= 1.0, got %f", result.Multiplier)
	}
	// If it's a weekend, should be 1.5; if weekday, should be 1.0
	if result.IsWeekend && result.Multiplier != 1.5 {
		t.Errorf("expected 1.5 on weekend, got %f", result.Multiplier)
	}
	if !result.IsWeekend && result.Multiplier != 1.0 {
		t.Errorf("expected 1.0 on weekday with only weekend boost, got %f", result.Multiplier)
	}
}

func TestSeasonalMultiplier_Q4Boost_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	st := &model.SeasonalTargeting{Q4Boost: 1.8}
	campaign := makeCampaignWithSeasonal_B28(st)
	result := svc.calculateSeasonalMultiplier(campaign)
	now := time.Now()
	month := int(now.Month())
	isQ4 := month >= 10 && month <= 12
	if isQ4 && result.Multiplier != 1.8 {
		t.Errorf("expected 1.8 in Q4, got %f", result.Multiplier)
	}
	if !isQ4 && result.Multiplier != 1.0 {
		t.Errorf("expected 1.0 outside Q4, got %f", result.Multiplier)
	}
}

func TestSeasonalMultiplier_SummerBoost_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	st := &model.SeasonalTargeting{SummerBoost: 1.4}
	campaign := makeCampaignWithSeasonal_B28(st)
	result := svc.calculateSeasonalMultiplier(campaign)
	now := time.Now()
	month := int(now.Month())
	isSummer := month >= 6 && month <= 8
	if isSummer && result.Multiplier != 1.4 {
		t.Errorf("expected 1.4 in summer, got %f", result.Multiplier)
	}
	if !isSummer && result.Multiplier != 1.0 {
		t.Errorf("expected 1.0 outside summer with only SummerBoost, got %f", result.Multiplier)
	}
}

func TestSeasonalMultiplier_BackToSchoolBoost_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	st := &model.SeasonalTargeting{BackToSchoolBoost: 1.3}
	campaign := makeCampaignWithSeasonal_B28(st)
	result := svc.calculateSeasonalMultiplier(campaign)
	// Just verify the multiplier is non-negative
	if result.Multiplier < 1.0 {
		t.Errorf("expected multiplier >= 1.0, got %f", result.Multiplier)
	}
}

func TestSeasonalMultiplier_RecurringEvent_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// Create a recurring event that always matches (covers the year)
	now := time.Now()
	// Create event for "all year" using Jan 01 - Dec 31 recurring
	startDate := "01-01"
	endDate := "12-31"
	_ = now // suppress unused warning

	st := &model.SeasonalTargeting{
		Events: []model.SeasonalEvent{
			{
				Name:      "Year Round",
				StartDate: startDate,
				EndDate:   endDate,
				Boost:     2.0,
				Recurring: true,
				Active:    true,
			},
		},
	}
	campaign := makeCampaignWithSeasonal_B28(st)
	result := svc.calculateSeasonalMultiplier(campaign)
	if !result.Matched {
		t.Errorf("expected matched=true for year-round event")
	}
	if result.Multiplier < 1.9 || result.Multiplier > 2.1 {
		t.Errorf("expected ~2.0 multiplier for year-round event, got %f", result.Multiplier)
	}
}

func TestSeasonalMultiplier_InactiveEvent_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	st := &model.SeasonalTargeting{
		Events: []model.SeasonalEvent{
			{
				Name:      "Black Friday",
				StartDate: "11-27",
				EndDate:   "11-29",
				Boost:     3.0,
				Recurring: true,
				Active:    false, // Inactive event — should be skipped
			},
		},
	}
	campaign := makeCampaignWithSeasonal_B28(st)
	result := svc.calculateSeasonalMultiplier(campaign)
	// Active=false → event should not apply
	if result.Matched {
		t.Errorf("expected inactive event to not match")
	}
}

func TestSeasonalMultiplier_HolidayBoost_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	st := &model.SeasonalTargeting{
		EnableHolidays: true,
		HolidayBoost:   1.6,
		Country:        "US",
	}
	campaign := makeCampaignWithSeasonal_B28(st)
	result := svc.calculateSeasonalMultiplier(campaign)
	// Can't control current date, so just verify multiplier is valid
	if result.Multiplier < 1.0 {
		t.Errorf("expected multiplier >= 1.0, got %f", result.Multiplier)
	}
}

func TestSeasonalMultiplier_MultiplierCap_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// Combine multiple boosts that would exceed 3.0 cap
	st := &model.SeasonalTargeting{
		WeekendBoost: 2.0,
		Q4Boost:      2.0,
		SummerBoost:  2.0,
	}
	campaign := makeCampaignWithSeasonal_B28(st)
	result := svc.calculateSeasonalMultiplier(campaign)
	if result.Multiplier > 3.0 {
		t.Errorf("multiplier exceeded 3.0 cap: %f", result.Multiplier)
	}
}

func TestSeasonalMultiplier_TimezoneConfig_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	st := &model.SeasonalTargeting{
		WeekendBoost: 1.2,
		Timezone:     "America/New_York",
	}
	campaign := makeCampaignWithSeasonal_B28(st)
	result := svc.calculateSeasonalMultiplier(campaign)
	if result.Multiplier < 1.0 {
		t.Errorf("expected multiplier >= 1.0, got %f", result.Multiplier)
	}
}

// ─── calculateLanguageMultiplier Tests ───────────────────────────────────────

func makeCampaignWithLanguage_B28(lt *model.LanguageTargeting) *model.Campaign {
	return &model.Campaign{
		ID:       "camp-lang",
		BidPrice: 1.0,
		Targeting: model.Targeting{
			LanguageTargeting: lt,
		},
	}
}

func makeBidRequestWithLanguage_B28(lang string) *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-lang",
		PublisherID: "pub1",
		AdSlot:      model.AdSlot{ID: "slot1"},
		Device:      model.InternalDevice{Type: "desktop"},
		User:        model.InternalUser{Language: lang},
	}
}

func TestLanguageMultiplier_NoConfig_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	campaign := &model.Campaign{ID: "c1", BidPrice: 1.0}
	req := makeBidRequestWithLanguage_B28("en")
	result := svc.calculateLanguageMultiplier(campaign, req)
	if result.Blocked || result.Multiplier != 1.0 {
		t.Errorf("expected no block and 1.0 multiplier, got blocked=%v mult=%f", result.Blocked, result.Multiplier)
	}
}

func TestLanguageMultiplier_ExcludeLanguage_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	lt := &model.LanguageTargeting{
		ExcludeLanguages: []string{"zh"},
	}
	campaign := makeCampaignWithLanguage_B28(lt)
	req := makeBidRequestWithLanguage_B28("zh")
	result := svc.calculateLanguageMultiplier(campaign, req)
	if !result.Blocked {
		t.Errorf("expected blocked for excluded language zh")
	}
}

func TestLanguageMultiplier_ExcludeLanguage_ContentLanguage_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	lt := &model.LanguageTargeting{
		ExcludeLanguages: []string{"ar"},
		ContentLanguage:  true,
	}
	campaign := makeCampaignWithLanguage_B28(lt)
	req := &model.BidRequest{
		ID:          "req-lang",
		PublisherID: "pub1",
		AdSlot:      model.AdSlot{ID: "slot1"},
		Device:      model.InternalDevice{Type: "desktop"},
		User:        model.InternalUser{Language: "en"}, // user lang OK
		Context: map[string]interface{}{
			"content_language": "ar", // content lang excluded
		},
	}
	result := svc.calculateLanguageMultiplier(campaign, req)
	if !result.Blocked {
		t.Errorf("expected blocked for excluded content language ar")
	}
}

func TestLanguageMultiplier_TargetMatch_Boost_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	lt := &model.LanguageTargeting{
		Languages: []model.LanguageRule{
			{Code: "en", Boost: 1.5},
		},
	}
	campaign := makeCampaignWithLanguage_B28(lt)
	req := makeBidRequestWithLanguage_B28("en")
	result := svc.calculateLanguageMultiplier(campaign, req)
	if result.Blocked {
		t.Errorf("expected no block for matching language")
	}
	if result.Multiplier != 1.5 {
		t.Errorf("expected 1.5 multiplier, got %f", result.Multiplier)
	}
}

func TestLanguageMultiplier_RequiredLanguageMissing_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	lt := &model.LanguageTargeting{
		Languages: []model.LanguageRule{
			{Code: "fr", Required: true, Boost: 1.2},
		},
	}
	campaign := makeCampaignWithLanguage_B28(lt)
	req := makeBidRequestWithLanguage_B28("en") // User speaks en, not fr
	result := svc.calculateLanguageMultiplier(campaign, req)
	if !result.Blocked {
		t.Errorf("expected blocked for missing required language fr")
	}
	if result.Reason != "missing_required_language" {
		t.Errorf("expected reason missing_required_language, got %s", result.Reason)
	}
}

func TestLanguageMultiplier_LocaleMatching_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	lt := &model.LanguageTargeting{
		Languages: []model.LanguageRule{
			{Code: "en", Locale: "en-US", Boost: 1.3},
		},
		LocaleMatching: true,
	}
	campaign := makeCampaignWithLanguage_B28(lt)
	req := makeBidRequestWithLanguage_B28("en-US") // exact locale match
	result := svc.calculateLanguageMultiplier(campaign, req)
	if result.Blocked {
		t.Errorf("expected no block for locale match")
	}
	if result.Multiplier != 1.3 {
		t.Errorf("expected 1.3 multiplier for locale match, got %f", result.Multiplier)
	}
}

func TestLanguageMultiplier_DefaultBoost_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	lt := &model.LanguageTargeting{
		Languages: []model.LanguageRule{
			{Code: "en"}, // No Boost specified → defaults to 1.2
		},
	}
	campaign := makeCampaignWithLanguage_B28(lt)
	req := makeBidRequestWithLanguage_B28("en")
	result := svc.calculateLanguageMultiplier(campaign, req)
	if result.Multiplier != 1.2 {
		t.Errorf("expected default 1.2 boost, got %f", result.Multiplier)
	}
}

func TestLanguageMultiplier_ContentLang_NotPrimaryOnly_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	lt := &model.LanguageTargeting{
		Languages: []model.LanguageRule{
			{Code: "de", Boost: 1.4},
		},
		ContentLanguage: true,
		PrimaryOnly:     false,
	}
	campaign := makeCampaignWithLanguage_B28(lt)
	req := &model.BidRequest{
		ID:          "req-lang",
		PublisherID: "pub1",
		AdSlot:      model.AdSlot{ID: "slot1"},
		Device:      model.InternalDevice{Type: "desktop"},
		User:        model.InternalUser{Language: "en"}, // user lang = en
		Context: map[string]interface{}{
			"content_language": "de", // content lang = de (should match)
		},
	}
	result := svc.calculateLanguageMultiplier(campaign, req)
	if result.Blocked {
		t.Errorf("expected no block: content language de matches rule")
	}
	if result.Multiplier != 1.4 {
		t.Errorf("expected 1.4 from content language match, got %f", result.Multiplier)
	}
}

// ─── calculateVideoTargetingMultiplier Tests ─────────────────────────────────

func makeCampaignWithVideo_B28(vt *model.VideoTargeting) *model.Campaign {
	return &model.Campaign{
		ID:       "camp-video",
		BidPrice: 2.0,
		Targeting: model.Targeting{
			VideoTargeting: vt,
		},
	}
}

func makeVideoRequest_B28(ctx map[string]interface{}) *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-video",
		PublisherID: "pub1",
		AdSlot:      model.AdSlot{ID: "slot1"},
		Device:      model.InternalDevice{Type: "desktop"},
		Context:     ctx,
	}
}

func TestVideoTargeting_NoConfig_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	campaign := &model.Campaign{ID: "c1", BidPrice: 1.0}
	req := makeVideoRequest_B28(nil)
	result := svc.calculateVideoTargetingMultiplier(campaign, req)
	if result.Blocked || result.Multiplier != 1.0 {
		t.Errorf("expected no block and 1.0, got blocked=%v mult=%f", result.Blocked, result.Multiplier)
	}
}

func TestVideoTargeting_NotVideoInventory_Blocked_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// VideoTargeting configured with placements but request is not video
	vt := &model.VideoTargeting{
		Placements: []string{"instream"},
	}
	campaign := makeCampaignWithVideo_B28(vt)
	req := makeVideoRequest_B28(map[string]interface{}{
		"video": false, // NOT video
	})
	result := svc.calculateVideoTargetingMultiplier(campaign, req)
	if !result.Blocked {
		t.Errorf("expected blocked for non-video inventory with placement requirement")
	}
}

func TestVideoTargeting_DurationTooShort_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	vt := &model.VideoTargeting{
		MinDuration: 30, // Require at least 30 seconds
	}
	campaign := makeCampaignWithVideo_B28(vt)
	req := makeVideoRequest_B28(map[string]interface{}{
		"video":       true,
		"maxduration": float64(15), // Only 15 seconds available
	})
	result := svc.calculateVideoTargetingMultiplier(campaign, req)
	if !result.Blocked {
		t.Errorf("expected blocked for duration too short")
	}
	if result.Reason != "duration_too_short" {
		t.Errorf("expected reason duration_too_short, got %s", result.Reason)
	}
}

func TestVideoTargeting_DurationTooLong_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	vt := &model.VideoTargeting{
		MaxDuration: 15, // Max 15 seconds
	}
	campaign := makeCampaignWithVideo_B28(vt)
	req := makeVideoRequest_B28(map[string]interface{}{
		"video":       true,
		"minduration": float64(30), // Min 30 seconds — exceeds max
	})
	result := svc.calculateVideoTargetingMultiplier(campaign, req)
	if !result.Blocked {
		t.Errorf("expected blocked for duration too long")
	}
	if result.Reason != "duration_too_long" {
		t.Errorf("expected reason duration_too_long, got %s", result.Reason)
	}
}

func TestVideoTargeting_PlacementMatch_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	vt := &model.VideoTargeting{
		Placements: []string{"instream", "outstream"},
	}
	campaign := makeCampaignWithVideo_B28(vt)
	req := makeVideoRequest_B28(map[string]interface{}{
		"video":           true,
		"video_placement": "instream",
	})
	result := svc.calculateVideoTargetingMultiplier(campaign, req)
	if result.Blocked {
		t.Errorf("expected no block for matching placement")
	}
	if !result.Matched {
		t.Errorf("expected matched=true for instream placement")
	}
}

func TestVideoTargeting_PlacementMismatch_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	vt := &model.VideoTargeting{
		Placements: []string{"instream"},
	}
	campaign := makeCampaignWithVideo_B28(vt)
	req := makeVideoRequest_B28(map[string]interface{}{
		"video":           true,
		"video_placement": "outstream", // doesn't match instream
	})
	result := svc.calculateVideoTargetingMultiplier(campaign, req)
	if !result.Blocked {
		t.Errorf("expected blocked for placement mismatch")
	}
}

func TestVideoTargeting_SkipSettingsSkippableOnly_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	vt := &model.VideoTargeting{
		SkipSettings: &model.VideoSkipSettings{
			SkippableOnly: true,
		},
	}
	campaign := makeCampaignWithVideo_B28(vt)
	req := makeVideoRequest_B28(map[string]interface{}{
		"video": true,
		"skip":  false, // not skippable
	})
	result := svc.calculateVideoTargetingMultiplier(campaign, req)
	if !result.Blocked {
		t.Errorf("expected blocked: skippable_only but not skippable")
	}
}

func TestVideoTargeting_SkipSettingsNonSkippableOnly_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	vt := &model.VideoTargeting{
		SkipSettings: &model.VideoSkipSettings{
			NonSkippableOnly: true,
		},
	}
	campaign := makeCampaignWithVideo_B28(vt)
	req := makeVideoRequest_B28(map[string]interface{}{
		"video": true,
		"skip":  true, // skippable — blocked
	})
	result := svc.calculateVideoTargetingMultiplier(campaign, req)
	if !result.Blocked {
		t.Errorf("expected blocked: non_skippable_only but is skippable")
	}
}

func TestVideoTargeting_SkippableBoost_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	vt := &model.VideoTargeting{
		SkipSettings: &model.VideoSkipSettings{
			SkippableBoost: 1.4,
		},
	}
	campaign := makeCampaignWithVideo_B28(vt)
	req := makeVideoRequest_B28(map[string]interface{}{
		"video": true,
		"skip":  true,
	})
	result := svc.calculateVideoTargetingMultiplier(campaign, req)
	if result.Blocked {
		t.Errorf("expected no block for skippable with boost")
	}
	if result.Multiplier != 1.4 {
		t.Errorf("expected 1.4 skippable boost, got %f", result.Multiplier)
	}
}

// ─── evaluateSkipSettings Tests ──────────────────────────────────────────────

func TestEvaluateSkipSettings_NonSkipBoost_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	settings := &model.VideoSkipSettings{
		NonSkipBoost: 1.3,
	}
	result := svc.evaluateSkipSettings(false, 0, settings) // not skippable
	if result.blocked {
		t.Errorf("expected no block for non-skippable with NonSkipBoost")
	}
	if result.multiplier != 1.3 {
		t.Errorf("expected NonSkipBoost 1.3, got %f", result.multiplier)
	}
}

func TestEvaluateSkipSettings_SkipOffsetTooShort_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	settings := &model.VideoSkipSettings{
		MinSkipOffset: 5,
	}
	result := svc.evaluateSkipSettings(true, 2, settings) // skipOffset=2 < MinSkipOffset=5
	if !result.blocked {
		t.Errorf("expected blocked for skip offset too short")
	}
}

func TestEvaluateSkipSettings_SkipOffsetTooLong_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	settings := &model.VideoSkipSettings{
		MaxSkipOffset: 10,
	}
	result := svc.evaluateSkipSettings(true, 30, settings) // skipOffset=30 > MaxSkipOffset=10
	if !result.blocked {
		t.Errorf("expected blocked for skip offset too long")
	}
}

// ─── isEventActive Tests ─────────────────────────────────────────────────────

func TestIsEventActive_FullDateFormat_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	now := time.Now()
	event := model.SeasonalEvent{
		Name:      "Test Event",
		StartDate: now.Format("2006-01-02"),
		EndDate:   now.AddDate(0, 0, 5).Format("2006-01-02"),
		Active:    true,
	}
	active := svc.isEventActive(event, now)
	if !active {
		t.Errorf("expected event to be active for full date format within range")
	}
}

func TestIsEventActive_InvalidDate_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	now := time.Now()
	event := model.SeasonalEvent{
		Name:      "Bad Dates",
		StartDate: "not-a-date",
		EndDate:   "also-not-a-date",
		Active:    true,
		Recurring: true,
	}
	active := svc.isEventActive(event, now)
	if active {
		t.Errorf("expected event to be inactive for invalid date format")
	}
}

// ─── calculateScore Tests ────────────────────────────────────────────────────

func TestCalculateScore_DealPriceOverride_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// Campaign with a deal that matches a request deal — exercises deal price override path
	campaign := &model.Campaign{
		ID:        "deal-camp",
		BidPrice:  1.0,
		DealID:    "deal123",
		DealType:  "private_auction",
		DealPrice: 5.0,
		Priority:  5,
	}
	req := &model.BidRequest{
		ID:          "req1",
		PublisherID: "pub1",
		AdSlot:      model.AdSlot{ID: "slot1"},
		Device:      model.InternalDevice{Type: "desktop"},
		Pmp: &model.Pmp{
			Deals: []model.Deal{
				{ID: "deal123"},
			},
		},
	}
	// calculateScore exercises deal multiplier path (DealPrice override)
	score := svc.calculateScore(campaign, req)
	if score < 0 {
		t.Errorf("expected non-negative score for deal campaign, got %f", score)
	}
}

func TestCalculateScore_BrandSafetyBlock_B28(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "")

	// Campaign with brand safety that will block via blocked category
	campaign := &model.Campaign{
		ID:                "brandsafe-camp",
		BidPrice:          2.0,
		BlockedCategories: []string{"adult"},
	}
	req := &model.BidRequest{
		ID:          "req1",
		PublisherID: "pub1",
		AdSlot:      model.AdSlot{ID: "slot1"},
		Device:      model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{
			"categories": []interface{}{"adult"},
		},
	}
	score := svc.calculateScore(campaign, req)
	if score != 0 {
		t.Errorf("expected 0 score for brand safety blocked campaign, got %f", score)
	}
}
