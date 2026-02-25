package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ─────────────────────────────────────────────────────────────────────────────
// calculateDayOfWeekMultiplier - additional branch coverage (44.2% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestDayOfWeek_WeekdaysOnly_Weekend(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		WeekdaysOnly: true,
	}
	result := s.calculateDayOfWeekMultiplier(camp)
	// Result depends on actual day — either allowed (weekday) or blocked (weekend)
	if result.IsWeekend && result.Allowed {
		t.Error("expected blocked on weekend when WeekdaysOnly=true")
	}
	if !result.IsWeekend && !result.Allowed {
		t.Errorf("expected allowed on weekday when WeekdaysOnly=true, reason: %s", result.Reason)
	}
}

func TestDayOfWeek_WeekendsOnly_Weekday(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		WeekendsOnly: true,
	}
	result := s.calculateDayOfWeekMultiplier(camp)
	if !result.IsWeekend && result.Allowed {
		t.Error("expected blocked on weekday when WeekendsOnly=true")
	}
	if result.IsWeekend && !result.Allowed {
		t.Errorf("expected allowed on weekend when WeekendsOnly=true, reason: %s", result.Reason)
	}
}

func TestDayOfWeek_DefaultBoost_NoMatchingDay(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	// Days list empty → should fall through to DefaultBoost
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		DefaultBoost: 1.15,
		Days:         []model.DaySchedule{},
	}
	result := s.calculateDayOfWeekMultiplier(camp)
	if !result.Allowed {
		t.Errorf("expected allowed when no matching day config, reason: %s", result.Reason)
	}
	if result.Multiplier != 1.15 {
		t.Errorf("expected DefaultBoost=1.15, got %f", result.Multiplier)
	}
}

func TestDayOfWeek_DayInactive(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	// Mark all days as inactive to ensure today is blocked
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		Days: []model.DaySchedule{
			{Day: 0, Active: false},
			{Day: 1, Active: false},
			{Day: 2, Active: false},
			{Day: 3, Active: false},
			{Day: 4, Active: false},
			{Day: 5, Active: false},
			{Day: 6, Active: false},
		},
	}
	result := s.calculateDayOfWeekMultiplier(camp)
	if result.Allowed {
		t.Error("expected blocked when today's day is inactive")
	}
	if result.Reason != "day_not_active" {
		t.Errorf("expected reason='day_not_active', got '%s'", result.Reason)
	}
}

func TestDayOfWeek_TimezoneSupport(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		Timezone: "America/New_York",
		Days: []model.DaySchedule{
			{Day: 0, Active: true, Boost: 1.0},
			{Day: 1, Active: true, Boost: 1.0},
			{Day: 2, Active: true, Boost: 1.0},
			{Day: 3, Active: true, Boost: 1.0},
			{Day: 4, Active: true, Boost: 1.0},
			{Day: 5, Active: true, Boost: 1.0},
			{Day: 6, Active: true, Boost: 1.0},
		},
	}
	result := s.calculateDayOfWeekMultiplier(camp)
	if !result.Allowed {
		t.Errorf("expected allowed with NY timezone, reason: %s", result.Reason)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateSeasonalMultiplier - additional branch coverage (47.4% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestSeasonal_WeekendBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		WeekendBoost: 1.25,
	}
	result := s.calculateSeasonalMultiplier(camp)
	// On weekends: should apply boost; on weekdays: multiplier stays 1.0
	if result.IsWeekend && result.Multiplier < 1.24 {
		t.Errorf("expected WeekendBoost applied (>=1.25) on weekend, got %f", result.Multiplier)
	}
	if !result.IsWeekend && result.Multiplier != 1.0 {
		t.Errorf("expected 1.0 on weekday with only WeekendBoost, got %f", result.Multiplier)
	}
}

func TestSeasonal_Q4Boost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		Q4Boost: 1.5,
	}
	result := s.calculateSeasonalMultiplier(camp)
	// If current month is Oct-Dec, boost applied; otherwise not
	if result.IsQ4 && result.Multiplier < 1.4 {
		t.Errorf("expected Q4Boost applied on Q4 month, got %f", result.Multiplier)
	}
}

func TestSeasonal_WithActiveEvent(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	// Use recurring event spanning the full year to guarantee it's active
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		Events: []model.SeasonalEvent{
			{
				Name:      "always-on",
				Active:    true,
				Recurring: true,
				StartDate: "01-01",
				EndDate:   "12-31",
				Boost:     1.4,
			},
		},
	}
	result := s.calculateSeasonalMultiplier(camp)
	if !result.Matched {
		t.Error("expected matched for year-round active event")
	}
	if result.Multiplier < 1.3 {
		t.Errorf("expected multiplier >= 1.4 for active event, got %f", result.Multiplier)
	}
	if len(result.ActiveEvents) == 0 {
		t.Error("expected at least one active event")
	}
}

func TestSeasonal_WithInactiveEvent(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		Events: []model.SeasonalEvent{
			{
				Name:   "inactive-event",
				Active: false,
				Boost:  2.0,
			},
		},
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Matched {
		t.Error("expected not matched for inactive event")
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected 1.0 for inactive event, got %f", result.Multiplier)
	}
}

func TestSeasonal_HolidayBoostEnabled(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		EnableHolidays: true,
		HolidayBoost:   1.3,
		Country:        "US",
	}
	// Just verify it doesn't panic and returns valid multiplier
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier <= 0 {
		t.Errorf("expected positive multiplier, got %f", result.Multiplier)
	}
}

func TestSeasonal_CapAt3x(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	// Stack multiple large boosts to hit the 3.0 cap
	camp.Targeting.SeasonalTargeting = &model.SeasonalTargeting{
		WeekendBoost: 2.0,
		Q4Boost:      2.0,
		SummerBoost:  2.0,
		HolidayBoost: 2.0,
		Events: []model.SeasonalEvent{
			{
				Name: "super-event", Active: true, Recurring: true,
				StartDate: "01-01", EndDate: "12-31", Boost: 2.0,
			},
		},
	}
	result := s.calculateSeasonalMultiplier(camp)
	if result.Multiplier > 3.0 {
		t.Errorf("expected multiplier capped at 3.0, got %f", result.Multiplier)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateVideoTargetingMultiplier - additional branches (33.6% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestVideoTargeting_NotVideoInventory_WithPlayerSize(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		PlayerSizes: []model.VideoPlayerSize{
			{Size: "large", Boost: 1.3},
		},
	}
	// No video context → non-video inventory
	req := newReq()
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for non-video inventory with PlayerSizes targeting")
	}
	if result.Reason != "not_video_inventory" {
		t.Errorf("expected reason 'not_video_inventory', got '%s'", result.Reason)
	}
}

func TestVideoTargeting_PlacementMismatch(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		Placements: []string{"instream"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":           true,
			"video_placement": "outstream",
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for placement mismatch")
	}
	if result.Reason != "placement_mismatch" {
		t.Errorf("expected 'placement_mismatch', got '%s'", result.Reason)
	}
}

func TestVideoTargeting_LargePlayerBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{}

	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":         true,
			"player_width":  float64(1920),
			"player_height": float64(1080),
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	// Large player width >= 1280 → *1.3
	if result.Multiplier < 1.2 {
		t.Errorf("expected multiplier >= 1.2 for large player, got %f", result.Multiplier)
	}
}

func TestVideoTargeting_MediumPlayerBoost(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{}

	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":         true,
			"player_width":  float64(640),
			"player_height": float64(360),
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	// 640 <= width < 1280 → *1.1
	if result.Multiplier < 1.05 {
		t.Errorf("expected multiplier >= 1.05 for medium player, got %f", result.Multiplier)
	}
}

func TestVideoTargeting_DurationTooLong(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		MaxDuration: 15, // campaign max is 15s
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":       true,
			"minduration": float64(30), // inventory min is 30s — too long
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for duration too long")
	}
	if result.Reason != "duration_too_long" {
		t.Errorf("expected 'duration_too_long', got '%s'", result.Reason)
	}
}

func TestVideoTargeting_StartDelayMismatch(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		StartDelays: []int{0}, // pre-roll only
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":      true,
			"startdelay": float64(-1), // post-roll (not 0)
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for start delay mismatch")
	}
}

func TestVideoTargeting_CompletionRateTooLow(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		CompletionRates: &model.CompletionRateRule{
			MinCompletionRate: 0.5,
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":           true,
			"completion_rate": float64(0.3),
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for completion rate too low")
	}
}

func TestVideoTargeting_MimeMismatch(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		Mimes: []string{"video/mp4"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"video":       true,
			"video_mimes": []interface{}{"video/webm"},
		},
	}
	result := s.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for MIME type mismatch")
	}
	if result.Reason != "mime_type_mismatch" {
		t.Errorf("expected 'mime_type_mismatch', got '%s'", result.Reason)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculatePerformanceGoalMultiplier - additional goal types (43% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func makeGoalCamp(goal string) *model.Campaign {
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: goal,
	}
	return camp
}

func TestPerfGoal_CPCGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("cpc"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for cpc goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CPMGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("cpm"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for cpm goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CPIGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("cpi"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for cpi goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CPSGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("cps"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for cps goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CPRGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("cpr"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for cpr goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CTVGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("ctv"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for ctv goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_ViewabilityGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("viewability"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for viewability goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CompletionGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("completion"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for completion goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_EngagementGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("engagement"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for engagement goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CPLGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("cpl"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for cpl goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CPVGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("cpv"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for cpv goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CPEGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("cpe"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for cpe goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_VCPMGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("vcpm"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for vcpm goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CPCVGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("cpcv"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for cpcv goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_DCPMGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("dcpm"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for dcpm goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CPADGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("cpa_d"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for cpa_d goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_CPIAAPGoal(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	result := s.calculatePerformanceGoalMultiplier(makeGoalCamp("cpiaap"), newReq())
	if result.Blocked {
		t.Errorf("expected not blocked for cpiaap goal, reason: %s", result.Reason)
	}
}

func TestPerfGoal_MinMaxBidAdjust(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal:  "cpa",
		TargetCPA:    1.0,
		MaxBidAdjust: 1.5,
		MinBidAdjust: 0.5,
	}
	result := s.calculatePerformanceGoalMultiplier(camp, newReq())
	if result.Blocked {
		t.Errorf("unexpected block, reason: %s", result.Reason)
	}
	if result.Multiplier > 1.5 {
		t.Errorf("multiplier should not exceed MaxBidAdjust=1.5, got %f", result.Multiplier)
	}
}

func TestPerfGoal_LearningMode_ThresholdBlocked(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal:  "cpa",
		TargetCPA:    5.0,
		LearningMode: true, // In learning mode → reduced bid instead of block
		Thresholds: &model.PerformanceThresholds{
			MinCTR: 0.99, // Impossible threshold to force "blocked" in learning
		},
	}
	mc := NewMockCache()
	mc.kv["perf:camp-1"] = "ctr:0.01,win_rate:0.10"
	camp.ID = "camp-1"
	result := s.calculatePerformanceGoalMultiplier(camp, newReq())
	// In learning mode, shouldn't be fully blocked
	if result.Blocked {
		t.Error("expected NOT blocked in learning mode even when threshold exceeded")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateCarrierMultiplier - additional branches (55.4% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestCarrier_CellularOnly_WiFiConn(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		CellularOnly: true,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"connection_type": "wifi"},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for wifi when CellularOnly=true")
	}
	if result.Reason != "cellular_only" {
		t.Errorf("expected 'cellular_only', got '%s'", result.Reason)
	}
}

func TestCarrier_WiFiOnly_CellularConn(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		WiFiOnly: true,
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"connection_type": "cellular"},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for cellular when WiFiOnly=true")
	}
	if result.Reason != "wifi_only" {
		t.Errorf("expected 'wifi_only', got '%s'", result.Reason)
	}
}

func TestCarrier_ExcludeCarrier(t *testing.T) {
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

func TestCarrier_ConnectionTypeNotAllowed(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.CarrierTargeting = &model.CarrierTargeting{
		ConnectionTypes: []string{"cellular"},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"connection_type": "wifi"},
	}
	result := s.calculateCarrierMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for non-allowed connection type")
	}
	if result.Reason != "connection_type_not_allowed" {
		t.Errorf("expected 'connection_type_not_allowed', got '%s'", result.Reason)
	}
}

func TestCarrier_AllowedCarrierBoost(t *testing.T) {
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
		t.Errorf("expected not blocked for allowed carrier, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.2 {
		t.Errorf("expected boost >= 1.2 for Verizon carrier, got %f", result.Multiplier)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateDemographicMultiplier - additional branches (56.2% → improve)
// ─────────────────────────────────────────────────────────────────────────────

func TestDemo_ExcludeAgeRange(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		ExcludeAgeRanges: []model.AgeRange{
			{MinAge: 13, MaxAge: 17},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"age": float64(15)},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked for excluded age range")
	}
	if result.Reason != "age_excluded" {
		t.Errorf("expected 'age_excluded', got '%s'", result.Reason)
	}
}

func TestDemo_RequiredAgeNotMet(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		AgeRanges: []model.AgeRange{
			{MinAge: 25, MaxAge: 45, Required: true, Boost: 1.3},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"age": float64(18)},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked when required age not met")
	}
	if result.Reason != "missing_required_age_range" {
		t.Errorf("expected 'missing_required_age_range', got '%s'", result.Reason)
	}
}

func TestDemo_UnknownAgeDiscount(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		AgeRanges: []model.AgeRange{
			{MinAge: 25, MaxAge: 45, Boost: 1.3},
		},
		UnknownAgeBoost: 0.7,
	}
	// No age in context → age=0
	req := &model.BidRequest{
		Context: map[string]interface{}{},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked for unknown age, reason: %s", result.Reason)
	}
	if result.Multiplier > 0.75 {
		t.Errorf("expected UnknownAgeBoost=0.7 applied, got %f", result.Multiplier)
	}
}

func TestDemo_RequiredGenderNotMet(t *testing.T) {
	s := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.DemographicTargeting = &model.DemographicTargeting{
		Genders: []model.GenderRule{
			{Gender: "female", Required: true, Boost: 1.2},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{"gender": "male"},
	}
	result := s.calculateDemographicMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked when required gender not met")
	}
	if result.Reason != "missing_required_gender" {
		t.Errorf("expected 'missing_required_gender', got '%s'", result.Reason)
	}
}
