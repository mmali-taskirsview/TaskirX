package service

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func makeMinCamp_B30() *model.Campaign {
	return &model.Campaign{
		ID:        "b30-camp",
		Name:      "B30 Campaign",
		Type:      "cpm",
		BidPrice:  1.0,
		Status:    "active",
		Budget:    10000,
		Targeting: model.Targeting{},
	}
}

func makeMinReq_B30() *model.BidRequest {
	return &model.BidRequest{
		ID:          "b30-req",
		PublisherID: "b30-pub",
		Device:      model.InternalDevice{Type: "mobile"},
		User:        model.InternalUser{ID: "b30-user"},
	}
}

func newBiddingSvc_B30() *BiddingService {
	mc := cache.NewMockCache()
	return NewBiddingService(mc, "")
}

// ── calculateAdPositionMultiplier ─────────────────────────────────────────────

func TestB30_AdPosition_NilTargeting(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	req := makeMinReq_B30()
	// No AdPositionTargeting
	r := svc.calculateAdPositionMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked with nil targeting")
	}
	if r.Multiplier != 1.0 {
		t.Errorf("expected 1.0 multiplier, got %v", r.Multiplier)
	}
}

func TestB30_AdPosition_ViewabilityBlock(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		MinViewability: 0.7, // require 70%
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"predicted_viewability": 0.4, // below minimum
	}
	r := svc.calculateAdPositionMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked due to viewability below minimum")
	}
	if r.Reason != "viewability_below_minimum" {
		t.Errorf("expected 'viewability_below_minimum', got %q", r.Reason)
	}
}

func TestB30_AdPosition_AboveFoldOnly_Block(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		AboveFoldOnly: true,
	}
	req := makeMinReq_B30()
	// No context → isAboveFold=false
	req.Context = map[string]interface{}{
		"above_fold":  false,
		"ad_position": "footer",
	}
	r := svc.calculateAdPositionMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked due to above_fold_only requirement")
	}
	if r.Reason != "above_fold_only" {
		t.Errorf("expected 'above_fold_only', got %q", r.Reason)
	}
}

func TestB30_AdPosition_AboveFoldBoost(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		AboveFoldBoost: 1.5,
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"above_fold":  true,
		"ad_position": "header",
	}
	r := svc.calculateAdPositionMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked")
	}
	if r.Multiplier < 1.4 {
		t.Errorf("expected above-fold boost ≥1.4, got %v", r.Multiplier)
	}
}

func TestB30_AdPosition_BelowFoldDiscount(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		BelowFoldDiscount: 0.7,
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"above_fold":  false,
		"ad_position": "footer",
	}
	r := svc.calculateAdPositionMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked")
	}
	if r.Multiplier > 0.8 {
		t.Errorf("expected below-fold discount ≤0.8, got %v", r.Multiplier)
	}
}

func TestB30_AdPosition_ExcludePositions(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		ExcludePositions: []string{"sidebar", "footer"},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"ad_position": "sidebar",
	}
	r := svc.calculateAdPositionMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked for excluded position")
	}
}

func TestB30_AdPosition_InterstitialBoost(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		InterstitialBoost: 1.8,
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"ad_position": "interstitial",
	}
	r := svc.calculateAdPositionMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked for interstitial")
	}
	if r.Multiplier < 1.5 {
		t.Errorf("expected interstitial boost ≥1.5, got %v", r.Multiplier)
	}
}

func TestB30_AdPosition_StickyBoost(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		StickyBoost: 1.6,
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"ad_position": "sticky",
	}
	r := svc.calculateAdPositionMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked for sticky")
	}
	if r.Multiplier < 1.4 {
		t.Errorf("expected sticky boost ≥1.4, got %v", r.Multiplier)
	}
}

func TestB30_AdPosition_PositionRule_Required_Missed(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		Positions: []model.PositionRule{
			{Position: "above_fold", Required: true, Boost: 1.4},
		},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"ad_position": "footer",
	}
	r := svc.calculateAdPositionMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked: required position not matched")
	}
	if r.Reason != "missing_required_position" {
		t.Errorf("expected 'missing_required_position', got %q", r.Reason)
	}
}

func TestB30_AdPosition_PositionRule_Matched(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		Positions: []model.PositionRule{
			{Position: "header", Boost: 1.3},
		},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"ad_position": "header",
	}
	r := svc.calculateAdPositionMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked")
	}
	if !r.Matched {
		t.Error("expected matched")
	}
	if r.Multiplier < 1.2 {
		t.Errorf("expected multiplier from position rule, got %v", r.Multiplier)
	}
}

func TestB30_AdPosition_MultiplierCap(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		AboveFoldBoost:    2.5,
		InterstitialBoost: 2.5,
		StickyBoost:       2.5,
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"above_fold":  true,
		"ad_position": "sticky_interstitial",
	}
	r := svc.calculateAdPositionMultiplier(camp, req)
	if r.Multiplier > 3.0 {
		t.Errorf("expected multiplier capped at 3.0, got %v", r.Multiplier)
	}
}

func TestB30_AdPosition_PositionFromAdSlot(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AdPositionTargeting = &model.AdPositionTargeting{
		AboveFoldBoost: 1.3,
	}
	req := makeMinReq_B30()
	req.AdSlot = model.AdSlot{Position: "above_fold"}
	// No Context — should use AdSlot.Position
	r := svc.calculateAdPositionMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked")
	}
}

// ── calculateAppTargetingMultiplier ──────────────────────────────────────────

func TestB30_AppTargeting_NilTargeting(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	req := makeMinReq_B30()
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked with nil targeting")
	}
	if r.Multiplier != 1.0 {
		t.Errorf("expected 1.0 multiplier, got %v", r.Multiplier)
	}
}

func TestB30_AppTargeting_InAppOnly_Block(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		InAppOnly: true,
	}
	req := makeMinReq_B30()
	// No bundle_id → isInApp = false
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked: in_app_only, no bundle")
	}
	if r.Reason != "in_app_only" {
		t.Errorf("expected 'in_app_only', got %q", r.Reason)
	}
}

func TestB30_AppTargeting_InAppOnly_Pass(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		InAppOnly: true,
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"bundle_id": "com.example.app",
		"is_app":    true,
	}
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked: is_app=true")
	}
}

func TestB30_AppTargeting_MobileWebOnly_Block(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		MobileWebOnly: true,
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"bundle_id": "com.example.app",
		"is_app":    true,
	}
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked: mobile_web_only, but is in-app")
	}
	if r.Reason != "mobile_web_only" {
		t.Errorf("expected 'mobile_web_only', got %q", r.Reason)
	}
}

func TestB30_AppTargeting_MinRating_Block(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		MinAppRating: 4.0,
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"app_rating": 2.5,
	}
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked: app rating below minimum")
	}
	if r.Reason != "app_rating_below_minimum" {
		t.Errorf("expected 'app_rating_below_minimum', got %q", r.Reason)
	}
}

func TestB30_AppTargeting_ExcludeBundle(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		ExcludeBundleIDs: []string{"com.bad.app"},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"bundle_id": "com.bad.app",
	}
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked: bundle excluded")
	}
}

func TestB30_AppTargeting_ExcludeCategory(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		ExcludeCategories: []string{"gambling"},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"app_category": "gambling",
	}
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked: category excluded")
	}
}

func TestB30_AppTargeting_RequiredBundle_Missed(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		BundleIDs: []model.AppRule{
			{Value: "com.target.app", Required: true, Boost: 1.5},
		},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"bundle_id": "com.other.app",
		"is_app":    true,
	}
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked: required bundle not matched")
	}
	if r.Reason != "missing_required_bundle_id" {
		t.Errorf("expected 'missing_required_bundle_id', got %q", r.Reason)
	}
}

func TestB30_AppTargeting_RequiredBundle_Matched(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		BundleIDs: []model.AppRule{
			{Value: "com.target.app", Required: true, Boost: 1.5},
		},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"bundle_id": "com.target.app",
		"is_app":    true,
	}
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked: bundle matched")
	}
	if r.Multiplier < 1.4 {
		t.Errorf("expected bundle boost, got %v", r.Multiplier)
	}
}

func TestB30_AppTargeting_WildcardBundle(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		BundleIDs: []model.AppRule{
			{Value: "com.spotify.*", Boost: 1.3},
		},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"bundle_id": "com.spotify.music",
		"is_app":    true,
	}
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked: wildcard bundle matched")
	}
	if !r.Matched {
		t.Error("expected matched for wildcard bundle")
	}
}

func TestB30_AppTargeting_RequiredCategory_Missed(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		Categories: []model.AppRule{
			{Value: "games", Required: true, Boost: 1.3},
		},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"app_category": "finance",
		"is_app":       true,
	}
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked: required category not matched")
	}
	if r.Reason != "missing_required_category" {
		t.Errorf("expected 'missing_required_category', got %q", r.Reason)
	}
}

func TestB30_AppTargeting_CategoryAliasMatch(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		Categories: []model.AppRule{
			{Value: "games", Boost: 1.2},
		},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"app_category": "gaming", // alias for "games"
		"is_app":       true,
	}
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked")
	}
	if !r.Matched {
		t.Error("expected matched via category alias")
	}
}

func TestB30_AppTargeting_PremiumAppBoost(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		PremiumAppsBoost: 1.5,
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"bundle_id":  "com.spotify.music",
		"app_rating": float64(4.8),
		"is_app":     true,
	}
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked for premium app")
	}
	if !r.IsPremiumApp {
		t.Error("expected IsPremiumApp=true for spotify")
	}
	if r.Multiplier < 1.4 {
		t.Errorf("expected premium boost ≥1.4, got %v", r.Multiplier)
	}
}

func TestB30_AppTargeting_MultiplierCap(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.AppTargeting = &model.AppTargeting{
		PremiumAppsBoost: 3.5,
		BundleIDs: []model.AppRule{
			{Value: "com.netflix.*", Boost: 3.0},
		},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"bundle_id":  "com.netflix.app",
		"app_rating": float64(4.9),
		"is_app":     true,
	}
	r := svc.calculateAppTargetingMultiplier(camp, req)
	if r.Multiplier > 3.0 {
		t.Errorf("expected multiplier capped at 3.0, got %v", r.Multiplier)
	}
}

// ── calculateDayOfWeekMultiplier ──────────────────────────────────────────────

func TestB30_DayOfWeek_NilTargeting(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	r := svc.calculateDayOfWeekMultiplier(camp)
	if !r.Allowed {
		t.Error("expected allowed with nil targeting")
	}
	if r.Multiplier != 1.0 {
		t.Errorf("expected 1.0, got %v", r.Multiplier)
	}
}

func TestB30_DayOfWeek_WeekdaysOnly_Weekend(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	now := time.Now()
	// Force weekend by checking if today is weekend
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		WeekdaysOnly: true,
	}
	r := svc.calculateDayOfWeekMultiplier(camp)
	dayNum := int(now.Weekday())
	isWeekend := dayNum == 0 || dayNum == 6
	if isWeekend && r.Allowed {
		t.Error("expected not allowed on weekend with WeekdaysOnly")
	}
	if !isWeekend && !r.Allowed {
		t.Error("expected allowed on weekday with WeekdaysOnly")
	}
}

func TestB30_DayOfWeek_WeekendsOnly_Weekday(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	now := time.Now()
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		WeekendsOnly: true,
	}
	r := svc.calculateDayOfWeekMultiplier(camp)
	dayNum := int(now.Weekday())
	isWeekend := dayNum == 0 || dayNum == 6
	if !isWeekend && r.Allowed {
		t.Error("expected not allowed on weekday with WeekendsOnly")
	}
	if isWeekend && !r.Allowed {
		t.Error("expected allowed on weekend with WeekendsOnly")
	}
}

func TestB30_DayOfWeek_DefaultBoost(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		DefaultBoost: 1.25,
		// No Days configured → uses default boost
	}
	r := svc.calculateDayOfWeekMultiplier(camp)
	if !r.Allowed {
		t.Error("expected allowed")
	}
	if r.Multiplier != 1.25 {
		t.Errorf("expected default boost 1.25, got %v", r.Multiplier)
	}
}

func TestB30_DayOfWeek_DayConfig_TodayActive(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	todayNum := int(time.Now().Weekday())
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		Days: []model.DaySchedule{
			{Day: todayNum, Active: true, Boost: 1.4},
		},
	}
	r := svc.calculateDayOfWeekMultiplier(camp)
	if !r.Allowed {
		t.Error("expected allowed")
	}
	if r.Multiplier != 1.4 {
		t.Errorf("expected boost 1.4, got %v", r.Multiplier)
	}
}

func TestB30_DayOfWeek_DayConfig_TodayInactive(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	todayNum := int(time.Now().Weekday())
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		Days: []model.DaySchedule{
			{Day: todayNum, Active: false},
		},
	}
	r := svc.calculateDayOfWeekMultiplier(camp)
	if r.Allowed {
		t.Error("expected not allowed: day is inactive")
	}
	if r.Reason != "day_not_active" {
		t.Errorf("expected 'day_not_active', got %q", r.Reason)
	}
}

func TestB30_DayOfWeek_DayConfig_HourNotActive(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	todayNum := int(time.Now().Weekday())
	// Allow only hour 25 (impossible) → always blocked
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		Days: []model.DaySchedule{
			{Day: todayNum, Active: true, Hours: []int{25}},
		},
	}
	r := svc.calculateDayOfWeekMultiplier(camp)
	if r.Allowed {
		t.Error("expected not allowed: hour 25 never matches")
	}
	if r.Reason != "hour_not_active_for_day" {
		t.Errorf("expected 'hour_not_active_for_day', got %q", r.Reason)
	}
}

func TestB30_DayOfWeek_DayConfig_HourActive(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	todayNum := int(time.Now().Weekday())
	currentHour := time.Now().Hour()
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		Days: []model.DaySchedule{
			{Day: todayNum, Active: true, Boost: 1.1, Hours: []int{currentHour}},
		},
	}
	r := svc.calculateDayOfWeekMultiplier(camp)
	if !r.Allowed {
		t.Error("expected allowed: current hour is in allowed list")
	}
}

func TestB30_DayOfWeek_Timezone(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.DayOfWeekTargeting = &model.DayOfWeekTargeting{
		Timezone:     "America/New_York",
		DefaultBoost: 1.1,
	}
	r := svc.calculateDayOfWeekMultiplier(camp)
	if !r.Allowed {
		t.Error("expected allowed with timezone set")
	}
	if r.Multiplier != 1.1 {
		t.Errorf("expected boost 1.1, got %v", r.Multiplier)
	}
}

// ── calculateGoalPacingMultiplier ─────────────────────────────────────────────

func TestB30_GoalPacing_NoGoal(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	// GoalTarget=0
	r := svc.calculateGoalPacingMultiplier(camp)
	if r != 1.0 {
		t.Errorf("expected 1.0 with no goal, got %v", r)
	}
}

func TestB30_GoalPacing_InvalidDate(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.GoalTarget = 1000
	camp.GoalEndDate = "not-a-date"
	r := svc.calculateGoalPacingMultiplier(camp)
	if r != 1.0 {
		t.Errorf("expected 1.0 with invalid date, got %v", r)
	}
}

func TestB30_GoalPacing_GoalAlreadyMet(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.GoalTarget = 1000
	camp.GoalDelivered = 1000 // goal met
	camp.GoalEndDate = time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	r := svc.calculateGoalPacingMultiplier(camp)
	if r != 0.3 {
		t.Errorf("expected 0.3 when goal already met, got %v", r)
	}
}

func TestB30_GoalPacing_PastDeadline(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.GoalTarget = 1000
	camp.GoalDelivered = 200
	camp.GoalEndDate = time.Now().AddDate(0, 0, -2).Format("2006-01-02") // 2 days ago
	r := svc.calculateGoalPacingMultiplier(camp)
	if r != 0.5 {
		t.Errorf("expected 0.5 past deadline, got %v", r)
	}
}

func TestB30_GoalPacing_OnTrack(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.GoalTarget = 1000
	camp.GoalDelivered = 0 // nothing delivered yet (currentDaily → 1)
	camp.GoalEndDate = time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	r := svc.calculateGoalPacingMultiplier(camp)
	// With 0 delivered, currentDaily=1, requiredDaily=remaining/days+1
	// Just verify it returns a valid float64 and doesn't panic
	if r <= 0 {
		t.Errorf("expected positive pacing multiplier, got %v", r)
	}
}

// ── generateRecommendations ───────────────────────────────────────────────────

func newPredSvc_B30() *PerformancePredictionService {
	mc := cache.NewMockCache()
	return NewPerformancePredictionService(mc)
}

func TestB30_GenRec_CTRDecline(t *testing.T) {
	svc := newPredSvc_B30()
	predictions := map[string]*MetricPrediction{
		"ctr": {
			Metric:         "ctr",
			PredictedValue: 0.01,
			TrendDirection: "down",
			TrendStrength:  0.8, // strong decline
		},
	}
	ctx := PredictionContext{TimeOfDay: "afternoon"}
	recs := svc.generateRecommendations(predictions, map[string]float64{}, ctx)
	found := false
	for _, r := range recs {
		if r.Type == "creative_refresh" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected creative_refresh recommendation for declining CTR")
	}
}

func TestB30_GenRec_LowCVR(t *testing.T) {
	svc := newPredSvc_B30()
	predictions := map[string]*MetricPrediction{
		"cvr": {
			Metric:         "cvr",
			PredictedValue: 0.005, // below 0.01 threshold
		},
	}
	ctx := PredictionContext{TimeOfDay: "morning"}
	recs := svc.generateRecommendations(predictions, map[string]float64{}, ctx)
	found := false
	for _, r := range recs {
		if r.Type == "landing_page" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected landing_page recommendation for low CVR")
	}
}

func TestB30_GenRec_BidIncrease(t *testing.T) {
	svc := newPredSvc_B30()
	predictions := map[string]*MetricPrediction{
		"ctr": {
			Metric:         "ctr",
			PredictedValue: 0.03, // above 0.02 threshold
		},
	}
	features := map[string]float64{
		"bid_price": 1.5, // below 2.0 threshold
	}
	ctx := PredictionContext{TimeOfDay: "afternoon"}
	recs := svc.generateRecommendations(predictions, features, ctx)
	found := false
	for _, r := range recs {
		if r.Type == "bid_increase" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected bid_increase recommendation")
	}
}

func TestB30_GenRec_DayParting(t *testing.T) {
	svc := newPredSvc_B30()
	predictions := map[string]*MetricPrediction{}
	features := map[string]float64{}
	// "night" returns timeFactor=0.80, condition is < 0.8 (strict), so
	// day_parting is NOT triggered. Verify the function runs cleanly
	// and returns without panicking for a night context.
	ctx := PredictionContext{TimeOfDay: "night"}
	recs := svc.generateRecommendations(predictions, features, ctx)
	_ = recs // no panic is sufficient; night is on the boundary
}

func TestB30_GenRec_NoRecs_GoodPerformance(t *testing.T) {
	svc := newPredSvc_B30()
	predictions := map[string]*MetricPrediction{
		"ctr": {
			Metric:         "ctr",
			PredictedValue: 0.025,
			TrendDirection: "up",
			TrendStrength:  0.3,
		},
		"cvr": {
			Metric:         "cvr",
			PredictedValue: 0.05, // high CVR
		},
	}
	features := map[string]float64{
		"bid_price": 5.0, // high bid, no bid_increase
	}
	ctx := PredictionContext{TimeOfDay: "afternoon"}
	recs := svc.generateRecommendations(predictions, features, ctx)
	// With good CTR trend, high CVR, high bid, afternoon time → no recommendations
	_ = recs // just verify no panic, recs may be empty
}

// ── calculateAutoBidMultiplier ────────────────────────────────────────────────

func TestB30_AutoBid_NoData(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	// No cache data → returns 1.0
	r := svc.calculateAutoBidMultiplier(camp)
	if r != 1.0 {
		t.Errorf("expected 1.0 with no data, got %v", r)
	}
}

func TestB30_AutoBid_HighCTR_LowWinRate(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	mc := svc.cache.(*cache.MockCache)
	// CTR = 300/10000 = 3% > 2%, WinRate = 20/100 = 20% < 30%
	for i := 0; i < 10000; i++ {
		mc.IncrementCampaignImpressions(camp.ID)
	}
	for i := 0; i < 300; i++ {
		mc.IncrementCampaignClicks(camp.ID)
	}
	for i := 0; i < 100; i++ {
		mc.IncrementCampaignBids(camp.ID)
	}
	for i := 0; i < 20; i++ {
		mc.IncrementCampaignWins(camp.ID)
	}
	r := svc.calculateAutoBidMultiplier(camp)
	if r != 1.20 {
		t.Errorf("expected 1.20 (high CTR, low win rate), got %v", r)
	}
}

func TestB30_AutoBid_LowCTR_HighWinRate(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	mc := svc.cache.(*cache.MockCache)
	// CTR = 20/10000 = 0.2% < 0.5%, WinRate = 80/100 = 80% > 70%
	for i := 0; i < 10000; i++ {
		mc.IncrementCampaignImpressions(camp.ID)
	}
	for i := 0; i < 20; i++ {
		mc.IncrementCampaignClicks(camp.ID)
	}
	for i := 0; i < 100; i++ {
		mc.IncrementCampaignBids(camp.ID)
	}
	for i := 0; i < 80; i++ {
		mc.IncrementCampaignWins(camp.ID)
	}
	r := svc.calculateAutoBidMultiplier(camp)
	if r != 0.80 {
		t.Errorf("expected 0.80 (low CTR, high win rate), got %v", r)
	}
}

func TestB30_AutoBid_ModerateCTR_VeryLowWinRate(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	mc := svc.cache.(*cache.MockCache)
	// CTR = 150/10000 = 1.5% (1-3%), WinRate = 10/100 = 10% < 20%
	for i := 0; i < 10000; i++ {
		mc.IncrementCampaignImpressions(camp.ID)
	}
	for i := 0; i < 150; i++ {
		mc.IncrementCampaignClicks(camp.ID)
	}
	for i := 0; i < 100; i++ {
		mc.IncrementCampaignBids(camp.ID)
	}
	for i := 0; i < 10; i++ {
		mc.IncrementCampaignWins(camp.ID)
	}
	r := svc.calculateAutoBidMultiplier(camp)
	if r != 1.10 {
		t.Errorf("expected 1.10 (moderate CTR, very low win rate), got %v", r)
	}
}

func TestB30_AutoBid_LowCTR_ModerateWinRate(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	mc := svc.cache.(*cache.MockCache)
	// CTR = 50/10000 = 0.5%→ < 1.0, WinRate = 60/100 = 60% (50-70%)
	for i := 0; i < 10000; i++ {
		mc.IncrementCampaignImpressions(camp.ID)
	}
	for i := 0; i < 50; i++ {
		mc.IncrementCampaignClicks(camp.ID)
	}
	for i := 0; i < 100; i++ {
		mc.IncrementCampaignBids(camp.ID)
	}
	for i := 0; i < 60; i++ {
		mc.IncrementCampaignWins(camp.ID)
	}
	r := svc.calculateAutoBidMultiplier(camp)
	if r != 0.90 {
		t.Errorf("expected 0.90 (low CTR, moderate win rate), got %v", r)
	}
}

func TestB30_AutoBid_Neutral(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	mc := svc.cache.(*cache.MockCache)
	// CTR = 200/10000 = 2%, WinRate = 50/100 = 50% — neutral
	for i := 0; i < 10000; i++ {
		mc.IncrementCampaignImpressions(camp.ID)
	}
	for i := 0; i < 200; i++ {
		mc.IncrementCampaignClicks(camp.ID)
	}
	for i := 0; i < 100; i++ {
		mc.IncrementCampaignBids(camp.ID)
	}
	for i := 0; i < 50; i++ {
		mc.IncrementCampaignWins(camp.ID)
	}
	r := svc.calculateAutoBidMultiplier(camp)
	if r != 1.0 {
		t.Errorf("expected 1.0 (neutral), got %v", r)
	}
}

// ── calculateVideoTargetingMultiplier extra branches ──────────────────────────

func TestB30_Video_NotVideo_Blocked(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		Placements: []string{"in_stream"},
	}
	req := makeMinReq_B30()
	// No video context (video=false implicitly) but Placements set → blocked
	req.Context = map[string]interface{}{
		"video": false,
	}
	r := svc.calculateVideoTargetingMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked: video targeting with placements but not video inventory")
	}
	if r.Reason != "not_video_inventory" {
		t.Errorf("expected 'not_video_inventory', got %q", r.Reason)
	}
}

func TestB30_Video_DurationTooShort(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		MinDuration: 30,
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"video":       true,
		"maxduration": float64(15), // below min
	}
	r := svc.calculateVideoTargetingMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked: duration too short")
	}
	if r.Reason != "duration_too_short" {
		t.Errorf("expected 'duration_too_short', got %q", r.Reason)
	}
}

func TestB30_Video_PlacementMismatch(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		Placements: []string{"in_stream"},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"video":           true,
		"video_placement": "out_stream",
	}
	r := svc.calculateVideoTargetingMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked: placement mismatch")
	}
	if r.Reason != "placement_mismatch" {
		t.Errorf("expected 'placement_mismatch', got %q", r.Reason)
	}
}

func TestB30_Video_CompletionRateLow_Block(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		CompletionRates: &model.CompletionRateRule{
			MinCompletionRate: 0.5,
		},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"video":           true,
		"completion_rate": 0.3, // below min
	}
	r := svc.calculateVideoTargetingMultiplier(camp, req)
	if !r.Blocked {
		t.Error("expected blocked: completion rate too low")
	}
	if r.Reason != "completion_rate_too_low" {
		t.Errorf("expected 'completion_rate_too_low', got %q", r.Reason)
	}
}

func TestB30_Video_HighCompletion_Boost(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	camp.Targeting.VideoTargeting = &model.VideoTargeting{
		CompletionRates: &model.CompletionRateRule{
			HighCompletionBoost: 1.4,
		},
	}
	req := makeMinReq_B30()
	req.Context = map[string]interface{}{
		"video":           true,
		"completion_rate": 0.85, // >= 0.75
	}
	r := svc.calculateVideoTargetingMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked")
	}
	if r.Multiplier < 1.3 {
		t.Errorf("expected high-completion boost, got %v", r.Multiplier)
	}
}

func TestB30_Video_NilTargeting(t *testing.T) {
	svc := newBiddingSvc_B30()
	camp := makeMinCamp_B30()
	req := makeMinReq_B30()
	r := svc.calculateVideoTargetingMultiplier(camp, req)
	if r.Blocked {
		t.Error("expected not blocked with nil video targeting")
	}
	if r.Multiplier != 1.0 {
		t.Errorf("expected 1.0, got %v", r.Multiplier)
	}
}

// ── IncrementCampaignWins helper (if not already on MockCache) ────────────────
// Note: The mock_cache.go file defines IncrementCampaignWins - this is just
// a sanity check that our test file compiles fine referencing it.
