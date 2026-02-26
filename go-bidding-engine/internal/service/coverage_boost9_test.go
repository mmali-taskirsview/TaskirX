package service

// coverage_boost9_test.go – additional branch coverage targeting functions 70-79%:
//   - calculatePOIMultiplier: no location + required POI, within radius default boost,
//       required POI not near, distance boost, minDistance block, maxDistance block, cap
//   - evaluateSentimentTargeting: positive (default boost), negative (targeting), neutral,
//       TargetNegative=false, MinSentimentScore
//   - isCTVInventory: env strings (ctv/ott), device_type strings, context is_ctv
//   - calculatePerformanceGoalMultiplier: threshold blocked (no learning), CTVGoals path,
//       AppGoals path, EcommerceGoals path, Threshold+LearningMode, AudienceModeling suppression,
//       MinBidAdjust applied, MultiplierCap/Floor
//   - getHolidayName: US holiday branches (New Year's Day, MLK, Presidents Day, etc.)
//   - calculateSeasonalMultiplier: remaining branches via getHolidayName call path

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ─────────────────────────────────────────────────────────────────────────────
// calculatePOIMultiplier
// ─────────────────────────────────────────────────────────────────────────────

func newPOICampaign(pois []model.POI, distanceBoosts []model.DistanceBoost, minDist, maxDist float64) *model.Campaign {
	c := newCampaign(1.0)
	c.Targeting.POITargeting = &model.POITargeting{
		POIs:           pois,
		DistanceBoosts: distanceBoosts,
		MinDistance:    minDist,
		MaxDistance:    maxDist,
	}
	return c
}

func reqWithGeo(lat, lon float64) *model.BidRequest {
	req := newReq()
	req.Device.Geo.Lat = lat
	req.Device.Geo.Lon = lon
	return req
}

// nil POI targeting → multiplier 1.0
func TestPOI_NilConfig(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.POITargeting = nil
	result := svc.calculatePOIMultiplier(camp, newReq())
	if result.Multiplier != 1.0 || result.Blocked {
		t.Errorf("expected default for nil POITargeting, got %+v", result)
	}
}

// No location data, no required POI → neutral pass
func TestPOI_NoLocation_NoRequired(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newPOICampaign([]model.POI{
		{ID: "p1", Name: "Store", Lat: 40.7128, Lon: -74.0060, Radius: 1.0, Required: false},
	}, nil, 0, 0)
	result := svc.calculatePOIMultiplier(camp, newReq()) // no geo in req
	if result.Blocked {
		t.Errorf("expected not blocked when no location and POI not required, got reason=%s", result.Reason)
	}
}

// No location but a Required POI → blocked "location_required_missing"
func TestPOI_NoLocation_RequiredPOI_Blocked(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newPOICampaign([]model.POI{
		{ID: "p1", Name: "HQ", Lat: 40.7128, Lon: -74.0060, Radius: 1.0, Required: true},
	}, nil, 0, 0)
	result := svc.calculatePOIMultiplier(camp, newReq())
	if !result.Blocked {
		t.Error("expected blocked when location missing and POI is required")
	}
	if result.Reason != "location_required_missing" {
		t.Errorf("expected 'location_required_missing', got '%s'", result.Reason)
	}
}

// User within radius → matched + default boost 1.3
func TestPOI_WithinRadius_DefaultBoost(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	// POI at Times Square (40.7580, -73.9855); place user nearby
	camp := newPOICampaign([]model.POI{
		{ID: "ts", Name: "Times Square", Lat: 40.7580, Lon: -73.9855, Radius: 2.0, Boost: 0 /*default*/},
	}, nil, 0, 0)
	req := reqWithGeo(40.7590, -73.9860) // ~0.12 km away

	result := svc.calculatePOIMultiplier(camp, req)
	if !result.Matched {
		t.Error("expected matched when user is within radius")
	}
	// Default boost = 1.3
	if result.Multiplier < 1.25 {
		t.Errorf("expected default POI boost ~1.3, got %.2f", result.Multiplier)
	}
}

// User within radius + explicit boost
func TestPOI_WithinRadius_ExplicitBoost(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newPOICampaign([]model.POI{
		{ID: "mall", Name: "Mall", Lat: 40.7580, Lon: -73.9855, Radius: 2.0, Boost: 1.5},
	}, nil, 0, 0)
	req := reqWithGeo(40.7590, -73.9860)

	result := svc.calculatePOIMultiplier(camp, req)
	if result.Multiplier < 1.4 {
		t.Errorf("expected explicit POI boost 1.5, got %.2f", result.Multiplier)
	}
}

// Required POI outside radius → blocked "not_near_required_poi"
func TestPOI_RequiredPOI_OutsideRadius_Blocked(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newPOICampaign([]model.POI{
		{ID: "far", Name: "FarPlace", Lat: 51.5074, Lon: -0.1278, Radius: 0.5, Required: true}, // London
	}, nil, 0, 0)
	req := reqWithGeo(40.7128, -74.0060) // New York - very far

	result := svc.calculatePOIMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked when user is not near required POI")
	}
}

// Distance boost applied (user near POI, distance bracket matches)
func TestPOI_DistanceBoost(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newPOICampaign(
		[]model.POI{
			{ID: "cafe", Name: "Cafe", Lat: 40.7580, Lon: -73.9855, Radius: 10.0},
		},
		[]model.DistanceBoost{
			{MaxDistance: 0.5, Boost: 2.0}, // very close
			{MaxDistance: 2.0, Boost: 1.4}, // nearby
		},
		0, 0,
	)
	req := reqWithGeo(40.7590, -73.9860) // ~0.12 km

	result := svc.calculatePOIMultiplier(camp, req)
	// matched (within 10km), distance boost 2.0 applied (< 0.5km), default POI boost 1.3
	if result.Multiplier < 1.5 {
		t.Errorf("expected distance boost applied, got %.2f", result.Multiplier)
	}
}

// MinDistance violation → blocked "too_close_to_poi"
func TestPOI_TooClose_Blocked(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newPOICampaign([]model.POI{
		{ID: "competitor", Name: "CompetitorStore", Lat: 40.7580, Lon: -73.9855, Radius: 10.0},
	}, nil, 5.0 /*minDist=5km*/, 0)
	req := reqWithGeo(40.7590, -73.9860) // ~0.12 km - too close

	result := svc.calculatePOIMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked when user is closer than MinDistance")
	}
	if result.Reason != "too_close_to_poi" {
		t.Errorf("expected 'too_close_to_poi', got '%s'", result.Reason)
	}
}

// MaxDistance violation → blocked "too_far_from_poi"
func TestPOI_TooFar_Blocked(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newPOICampaign([]model.POI{
		{ID: "local", Name: "LocalStore", Lat: 40.7580, Lon: -73.9855, Radius: 100.0},
	}, nil, 0, 0.05 /*maxDist=50m*/)
	req := reqWithGeo(40.7128, -74.0060) // ~6.8 km - too far

	result := svc.calculatePOIMultiplier(camp, req)
	if !result.Blocked {
		t.Error("expected blocked when user is farther than MaxDistance")
	}
	if result.Reason != "too_far_from_poi" {
		t.Errorf("expected 'too_far_from_poi', got '%s'", result.Reason)
	}
}

// POI multiplier cap at 3.0 (stack multiple POI boosts)
func TestPOI_MultiplierCapped(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newPOICampaign([]model.POI{
		{ID: "p1", Name: "POI1", Lat: 40.7580, Lon: -73.9855, Radius: 10.0, Boost: 2.0},
		{ID: "p2", Name: "POI2", Lat: 40.7580, Lon: -73.9855, Radius: 10.0, Boost: 2.0},
	}, nil, 0, 0)
	req := reqWithGeo(40.7590, -73.9860)

	result := svc.calculatePOIMultiplier(camp, req)
	if result.Multiplier > 3.0 {
		t.Errorf("expected POI multiplier capped at 3.0, got %.2f", result.Multiplier)
	}
}

// Location from request context (lat/lon keys)
func TestPOI_LocationFromContext(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newPOICampaign([]model.POI{
		{ID: "ctx_poi", Name: "CtxPOI", Lat: 40.7580, Lon: -73.9855, Radius: 2.0},
	}, nil, 0, 0)
	req := newReq()
	req.Context = map[string]interface{}{
		"lat": 40.7590,
		"lon": -73.9860,
	}
	result := svc.calculatePOIMultiplier(camp, req)
	if !result.Matched {
		t.Error("expected matched when location provided via context lat/lon")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// evaluateSentimentTargeting
// ─────────────────────────────────────────────────────────────────────────────

func newContextualAISvc() *ContextualAIService {
	return NewContextualAIService(NewMockCache())
}

// Positive sentiment + TargetPositive=true, PositiveBoost > 0
func TestSentiment_Positive_WithBoost(t *testing.T) {
	svc := newContextualAISvc()
	config := &model.SentimentTargeting{
		TargetPositive: true,
		PositiveBoost:  1.4,
	}
	m := svc.evaluateSentimentTargeting("positive", 0.8, config)
	if m != 1.4 {
		t.Errorf("expected 1.4, got %.2f", m)
	}
}

// Positive sentiment + TargetPositive=true, PositiveBoost==0 → default 1.2
func TestSentiment_Positive_DefaultBoost(t *testing.T) {
	svc := newContextualAISvc()
	config := &model.SentimentTargeting{
		TargetPositive: true,
		PositiveBoost:  0,
	}
	m := svc.evaluateSentimentTargeting("positive", 0.8, config)
	if m != 1.2 {
		t.Errorf("expected default 1.2, got %.2f", m)
	}
}

// Positive sentiment + TargetPositive=false → multiplier stays 1.0
func TestSentiment_Positive_NotTargeted(t *testing.T) {
	svc := newContextualAISvc()
	config := &model.SentimentTargeting{TargetPositive: false}
	m := svc.evaluateSentimentTargeting("positive", 0.8, config)
	if m != 1.0 {
		t.Errorf("expected 1.0 when not targeting positive, got %.2f", m)
	}
}

// Negative sentiment + TargetNegative=true, NegativePenalty set
func TestSentiment_Negative_Targeted_WithPenalty(t *testing.T) {
	svc := newContextualAISvc()
	config := &model.SentimentTargeting{
		TargetNegative:  true,
		NegativePenalty: 0.6,
	}
	m := svc.evaluateSentimentTargeting("negative", -0.5, config)
	if m != 0.6 {
		t.Errorf("expected 0.6, got %.2f", m)
	}
}

// Negative sentiment + TargetNegative=true, NegativePenalty==0 → default 0.5
func TestSentiment_Negative_Targeted_DefaultPenalty(t *testing.T) {
	svc := newContextualAISvc()
	config := &model.SentimentTargeting{
		TargetNegative:  true,
		NegativePenalty: 0,
	}
	m := svc.evaluateSentimentTargeting("negative", -0.5, config)
	if m != 0.5 {
		t.Errorf("expected default penalty 0.5, got %.2f", m)
	}
}

// Negative sentiment + TargetNegative=false → heavy penalty 0.3
func TestSentiment_Negative_NotTargeted(t *testing.T) {
	svc := newContextualAISvc()
	config := &model.SentimentTargeting{TargetNegative: false}
	m := svc.evaluateSentimentTargeting("negative", -0.5, config)
	if m != 0.3 {
		t.Errorf("expected 0.3 heavy penalty, got %.2f", m)
	}
}

// Neutral sentiment + TargetNeutral=true → 1.0
func TestSentiment_Neutral_Targeted(t *testing.T) {
	svc := newContextualAISvc()
	config := &model.SentimentTargeting{TargetNeutral: true}
	m := svc.evaluateSentimentTargeting("neutral", 0.0, config)
	if m != 1.0 {
		t.Errorf("expected 1.0 for neutral targeted, got %.2f", m)
	}
}

// Neutral sentiment + TargetNeutral=false → 1.0 (no-op case in switch)
func TestSentiment_Neutral_NotTargeted(t *testing.T) {
	svc := newContextualAISvc()
	config := &model.SentimentTargeting{TargetNeutral: false}
	m := svc.evaluateSentimentTargeting("neutral", 0.0, config)
	if m != 1.0 {
		t.Errorf("expected 1.0 for neutral not targeted, got %.2f", m)
	}
}

// MinSentimentScore: score below min → *0.5 penalty
func TestSentiment_BelowMinScore(t *testing.T) {
	svc := newContextualAISvc()
	config := &model.SentimentTargeting{
		TargetPositive:    true,
		PositiveBoost:     1.4,
		MinSentimentScore: 0.8,
	}
	m := svc.evaluateSentimentTargeting("positive", 0.5, config) // 0.5 < 0.8
	// 1.4 * 0.5 = 0.7
	if m > 0.75 {
		t.Errorf("expected score penalty applied (~0.7), got %.2f", m)
	}
}

// MinSentimentScore: score above min → no penalty
func TestSentiment_AboveMinScore(t *testing.T) {
	svc := newContextualAISvc()
	config := &model.SentimentTargeting{
		TargetPositive:    true,
		PositiveBoost:     1.4,
		MinSentimentScore: 0.5,
	}
	m := svc.evaluateSentimentTargeting("positive", 0.9, config) // 0.9 >= 0.5
	if m != 1.4 {
		t.Errorf("expected no penalty when score above min, got %.2f", m)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// isCTVInventory – additional branches
// ─────────────────────────────────────────────────────────────────────────────

func TestIsCTV_DeviceType_TV(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := newReq()
	req.Device.Type = "tv"
	if !svc.isCTVInventory(req) {
		t.Error("expected true for device type 'tv'")
	}
}

func TestIsCTV_DeviceType_ConnectedTV(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := newReq()
	req.Device.Type = "connected_tv"
	if !svc.isCTVInventory(req) {
		t.Error("expected true for device type 'connected_tv'")
	}
}

func TestIsCTV_Context_IsCtv_True(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := newReq()
	req.Context = map[string]interface{}{"is_ctv": true}
	if !svc.isCTVInventory(req) {
		t.Error("expected true when context is_ctv=true")
	}
}

func TestIsCTV_Context_Environment_OTT(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := newReq()
	req.Context = map[string]interface{}{"environment": "ott_streaming"}
	if !svc.isCTVInventory(req) {
		t.Error("expected true for OTT environment")
	}
}

func TestIsCTV_Context_Environment_ConnectedTV(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := newReq()
	req.Context = map[string]interface{}{"environment": "connected_tv_app"}
	if !svc.isCTVInventory(req) {
		t.Error("expected true for connected_tv environment string")
	}
}

func TestIsCTV_Context_DeviceType_Roku(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := newReq()
	req.Context = map[string]interface{}{"device_type": "roku"}
	if !svc.isCTVInventory(req) {
		t.Error("expected true for device_type=roku")
	}
}

func TestIsCTV_Context_DeviceType_FireTV(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := newReq()
	req.Context = map[string]interface{}{"device_type": "fire_tv"}
	if !svc.isCTVInventory(req) {
		t.Error("expected true for device_type=fire_tv")
	}
}

func TestIsCTV_Context_DeviceType_AppleTV(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := newReq()
	req.Context = map[string]interface{}{"device_type": "apple_tv"}
	if !svc.isCTVInventory(req) {
		t.Error("expected true for device_type=apple_tv")
	}
}

func TestIsCTV_Context_DeviceType_GamingConsole(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := newReq()
	req.Context = map[string]interface{}{"device_type": "gaming_console"}
	if !svc.isCTVInventory(req) {
		t.Error("expected true for device_type=gaming_console")
	}
}

func TestIsCTV_Context_DeviceType_Chromecast(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := newReq()
	req.Context = map[string]interface{}{"device_type": "chromecast"}
	if !svc.isCTVInventory(req) {
		t.Error("expected true for device_type=chromecast")
	}
}

func TestIsCTV_NotCTV(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := newReq()
	req.Device.Type = "mobile"
	if svc.isCTVInventory(req) {
		t.Error("expected false for mobile device")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculatePerformanceGoalMultiplier – additional branches
// ─────────────────────────────────────────────────────────────────────────────

// Threshold blocked (no learning mode) → Blocked=true
func TestPerfGoal_ThresholdBlocked_NoLearning(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal:  "cpa",
		LearningMode: false,
		Thresholds: &model.PerformanceThresholds{
			MinCTR: 0.9999, // impossible → always blocked
		},
	}
	result := svc.calculatePerformanceGoalMultiplier(camp, newReq())
	if !result.Blocked {
		t.Error("expected blocked when threshold not met and not in learning mode")
	}
}

// CTVGoals applied when inventory is CTV
func TestPerfGoal_CTVGoals_Applied(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpm",
		CTVGoals: &model.CTVOptimization{
			TargetCompletionRate: 0.7,
			PrimtimeBoost:        1.3,
		},
	}
	req := newReq()
	req.Device.Type = "ctv" // make it CTV inventory

	result := svc.calculatePerformanceGoalMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("expected not blocked, reason=%s", result.Reason)
	}
	if !result.IsCTV {
		t.Error("expected IsCTV=true for ctv device")
	}
}

// CTVGoals present but NOT CTV inventory → CTVGoals branch skipped
func TestPerfGoal_CTVGoals_NotCTV(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpm",
		CTVGoals:    &model.CTVOptimization{PrimtimeBoost: 1.5},
	}
	req := newReq()
	req.Device.Type = "mobile" // not CTV

	result := svc.calculatePerformanceGoalMultiplier(camp, req)
	if result.IsCTV {
		t.Error("expected IsCTV=false for mobile")
	}
}

// AppGoals applied
func TestPerfGoal_AppGoals_Applied(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpi",
		TargetCPI:   2.0,
		AppGoals: &model.AppOptimization{
			TargetInstallRate:    0.05,
			TargetCostPerInstall: 2.0,
		},
	}
	result := svc.calculatePerformanceGoalMultiplier(camp, newReq())
	if result.Blocked {
		t.Errorf("expected not blocked with AppGoals, reason=%s", result.Reason)
	}
}

// EcommerceGoals applied (via calculatePerformanceGoalMultiplier)
func TestPerfGoal_EcommerceGoals_Applied(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cps",
		TargetCPS:   5.0,
		EcommerceGoals: &model.EcommerceOptimization{
			TargetROAS:       3.0,
			CartAbandonBoost: 1.5,
		},
	}
	result := svc.calculatePerformanceGoalMultiplier(camp, newReq())
	if result.Blocked {
		t.Errorf("expected not blocked with EcommerceGoals, reason=%s", result.Reason)
	}
}

// MinBidAdjust applied when multiplier drops too low
func TestPerfGoal_MinBidAdjust_Applied(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(0.001) // tiny bid → optimizers will produce low ratio
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal:  "cpm",
		TargetCPM:    0.0001, // very low target → multiplier will be small
		MinBidAdjust: 0.5,
	}
	result := svc.calculatePerformanceGoalMultiplier(camp, newReq())
	if result.Multiplier < 0.5 {
		t.Errorf("expected MinBidAdjust=0.5 floor applied, got %.4f", result.Multiplier)
	}
}

// Multiplier cap at 3.0
func TestPerfGoal_MultiplierCap(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpm",
		TargetCPM:   99999.0, // huge target → high multiplier
	}
	result := svc.calculatePerformanceGoalMultiplier(camp, newReq())
	if result.Multiplier > 3.0 {
		t.Errorf("expected multiplier capped at 3.0, got %.4f", result.Multiplier)
	}
}

// Multiplier floor at 0.3
func TestPerfGoal_MultiplierFloor(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(99999.0) // huge bid price → ratio will be tiny
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{
		PrimaryGoal: "cpa",
		TargetCPA:   0.001, // tiny target
	}
	result := svc.calculatePerformanceGoalMultiplier(camp, newReq())
	if result.Multiplier < 0.3 {
		t.Errorf("expected multiplier floored at 0.3, got %.4f", result.Multiplier)
	}
}

// CTV inventory path via getHouseholdID (household_id from context)
func TestPerfGoal_CTV_HouseholdID(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	camp.Targeting.PerformanceGoals = &model.PerformanceGoals{PrimaryGoal: "cpm"}
	req := newReq()
	req.Device.Type = "ctv"
	req.Context = map[string]interface{}{
		"household_id": "hh-abc-123",
	}
	result := svc.calculatePerformanceGoalMultiplier(camp, req)
	if result.HouseholdID != "hh-abc-123" {
		t.Errorf("expected HouseholdID='hh-abc-123', got '%s'", result.HouseholdID)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// getHolidayName – US holiday branches (exercise via calculateSeasonalMultiplier)
// ─────────────────────────────────────────────────────────────────────────────

// Call getHolidayName directly for various US dates to cover switch branches
func TestGetHolidayName_NewYearsDay(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	d := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "New Year's Day" {
		t.Errorf("expected 'New Year's Day', got '%s'", h)
	}
}

func TestGetHolidayName_ValentinesDay(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	d := time.Date(2026, 2, 14, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Valentine's Day" {
		t.Errorf("expected 'Valentine's Day', got '%s'", h)
	}
}

func TestGetHolidayName_IndependenceDay(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	d := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Independence Day" {
		t.Errorf("expected 'Independence Day', got '%s'", h)
	}
}

func TestGetHolidayName_VeteransDay(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	d := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Veterans Day" {
		t.Errorf("expected 'Veterans Day', got '%s'", h)
	}
}

func TestGetHolidayName_ChristmasEve(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	d := time.Date(2026, 12, 24, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Christmas Eve" {
		t.Errorf("expected 'Christmas Eve', got '%s'", h)
	}
}

func TestGetHolidayName_Christmas(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	d := time.Date(2026, 12, 25, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Christmas" {
		t.Errorf("expected 'Christmas', got '%s'", h)
	}
}

func TestGetHolidayName_NewYearsEve(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	d := time.Date(2026, 12, 31, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "New Year's Eve" {
		t.Errorf("expected 'New Year's Eve', got '%s'", h)
	}
}

func TestGetHolidayName_Halloween(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	d := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Halloween" {
		t.Errorf("expected 'Halloween', got '%s'", h)
	}
}

func TestGetHolidayName_MLKDay(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	// MLK Day = 3rd Monday of January 2026 = Jan 19
	d := time.Date(2026, 1, 19, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "MLK Day" {
		t.Errorf("expected 'MLK Day', got '%s'", h)
	}
}

func TestGetHolidayName_PresidentsDay(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	// Presidents Day = 3rd Monday of February 2026 = Feb 16
	d := time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Presidents Day" {
		t.Errorf("expected 'Presidents Day', got '%s'", h)
	}
}

func TestGetHolidayName_MemorialDay(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	// Memorial Day = last Monday of May 2026 = May 25
	d := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Memorial Day" {
		t.Errorf("expected 'Memorial Day', got '%s'", h)
	}
}

func TestGetHolidayName_LaborDay(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	// Labor Day = 1st Monday of September 2026 = Sep 7
	d := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Labor Day" {
		t.Errorf("expected 'Labor Day', got '%s'", h)
	}
}

func TestGetHolidayName_ColumbusDay(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	// Columbus Day = 2nd Monday of October 2026 = Oct 12
	d := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Columbus Day" {
		t.Errorf("expected 'Columbus Day', got '%s'", h)
	}
}

func TestGetHolidayName_Thanksgiving(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	// Thanksgiving = 4th Thursday of November 2026 = Nov 26
	d := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Thanksgiving" {
		t.Errorf("expected 'Thanksgiving', got '%s'", h)
	}
}

func TestGetHolidayName_BlackFriday(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	// Black Friday = day after Thanksgiving 2026 = Nov 27
	d := time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Black Friday" {
		t.Errorf("expected 'Black Friday', got '%s'", h)
	}
}

func TestGetHolidayName_CyberMonday(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	// Cyber Monday = Mon after Thanksgiving 2026 = Nov 30
	d := time.Date(2026, 11, 30, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Cyber Monday" {
		t.Errorf("expected 'Cyber Monday', got '%s'", h)
	}
}

func TestGetHolidayName_MotherDay(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	// Mother's Day = 2nd Sunday of May 2026 = May 10
	d := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Mother's Day" {
		t.Errorf("expected 'Mother's Day', got '%s'", h)
	}
}

func TestGetHolidayName_FathersDay(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	// Father's Day = 3rd Sunday of June 2026 = June 21
	d := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "US")
	if h != "Father's Day" {
		t.Errorf("expected 'Father's Day', got '%s'", h)
	}
}

func TestGetHolidayName_UK_Christmas(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	d := time.Date(2026, 12, 25, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "UK")
	if h != "Christmas" {
		t.Errorf("expected UK Christmas, got '%s'", h)
	}
}

func TestGetHolidayName_UK_BoxingDay(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	d := time.Date(2026, 12, 26, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "UK")
	if h != "Boxing Day" {
		t.Errorf("expected 'Boxing Day', got '%s'", h)
	}
}

func TestGetHolidayName_UK_NewYear(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	d := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	h := svc.getHolidayName(d, "UK")
	if h != "New Year's Day" {
		t.Errorf("expected UK New Year's Day, got '%s'", h)
	}
}

func TestGetHolidayName_NoMatch(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	// Ordinary Tuesday in February (not a holiday)
	d := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC) // regular Tuesday
	h := svc.getHolidayName(d, "US")
	if h != "" {
		t.Errorf("expected empty string for non-holiday date, got '%s'", h)
	}
}
