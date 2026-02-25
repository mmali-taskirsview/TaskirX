package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================================
// extractInventoryQualityContext tests
// ============================================================================

func TestExtractInventoryQualityContext_NilContext(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{Context: nil}

	ctx := svc.extractInventoryQualityContext(req)

	if ctx.qualityScore != 0.5 {
		t.Errorf("Expected default qualityScore 0.5, got %f", ctx.qualityScore)
	}
	if ctx.trustLevel != "unknown" {
		t.Errorf("Expected default trustLevel 'unknown', got %s", ctx.trustLevel)
	}
	if ctx.viewabilityRate != 0.5 {
		t.Errorf("Expected default viewabilityRate 0.5, got %f", ctx.viewabilityRate)
	}
}

func TestExtractInventoryQualityContext_FullContext(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"quality_score":         0.9,
			"trust_level":           "premium",
			"ads_txt_verified":      true,
			"sellers_json_verified": true,
			"viewability_rate":      0.75,
			"fraud_risk":            0.05,
			"content_rating":        "PG",
			"content_categories":    []interface{}{"IAB1", "IAB2"},
			"site_type":             "news",
			"bot_probability":       0.1,
			"proxy_detected":        false,
		},
	}

	ctx := svc.extractInventoryQualityContext(req)

	if ctx.qualityScore != 0.9 {
		t.Errorf("Expected qualityScore 0.9, got %f", ctx.qualityScore)
	}
	if ctx.trustLevel != "premium" {
		t.Errorf("Expected trustLevel 'premium', got %s", ctx.trustLevel)
	}
	if !ctx.adsTxtVerified {
		t.Error("Expected adsTxtVerified true")
	}
	if !ctx.sellersJsonVerified {
		t.Error("Expected sellersJsonVerified true")
	}
	if ctx.viewabilityRate != 0.75 {
		t.Errorf("Expected viewabilityRate 0.75, got %f", ctx.viewabilityRate)
	}
	if ctx.fraudRisk != 0.05 {
		t.Errorf("Expected fraudRisk 0.05, got %f", ctx.fraudRisk)
	}
	if ctx.contentRating != "PG" {
		t.Errorf("Expected contentRating 'PG', got %s", ctx.contentRating)
	}
	if len(ctx.contentCategories) != 2 {
		t.Errorf("Expected 2 content categories, got %d", len(ctx.contentCategories))
	}
	if ctx.siteType != "news" {
		t.Errorf("Expected siteType 'news', got %s", ctx.siteType)
	}
	if ctx.botProbability != 0.1 {
		t.Errorf("Expected botProbability 0.1, got %f", ctx.botProbability)
	}
	if ctx.proxyDetected {
		t.Error("Expected proxyDetected false")
	}
}

func TestExtractInventoryQualityContext_AltKeys(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"inventory_quality":      0.8,
			"seller_type":            "direct",
			"ads_txt":                true,
			"historical_viewability": 0.65,
			"ivt_score":              0.12,
		},
	}

	ctx := svc.extractInventoryQualityContext(req)

	if ctx.qualityScore != 0.8 {
		t.Errorf("Expected qualityScore 0.8, got %f", ctx.qualityScore)
	}
	if ctx.trustLevel != "direct" {
		t.Errorf("Expected trustLevel 'direct', got %s", ctx.trustLevel)
	}
	if !ctx.adsTxtVerified {
		t.Error("Expected adsTxtVerified true")
	}
	if ctx.viewabilityRate != 0.65 {
		t.Errorf("Expected viewabilityRate 0.65, got %f", ctx.viewabilityRate)
	}
	if ctx.fraudRisk != 0.12 {
		t.Errorf("Expected fraudRisk 0.12, got %f", ctx.fraudRisk)
	}
}

// ============================================================================
// evaluateBrandSuitability tests
// ============================================================================

func TestEvaluateBrandSuitability_NoContext(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{Context: nil}
	bs := &model.BrandSuitability{
		FloorRating: "PG",
	}

	result := svc.evaluateBrandSuitability(req, bs)

	// No context means no rating to check → should pass
	if result.blocked {
		t.Errorf("Expected not blocked, got blocked: %s", result.reason)
	}
}

func TestEvaluateBrandSuitability_BlockedCategory(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"content_categories": []interface{}{"IAB25-3"}, // profanity
		},
	}
	bs := &model.BrandSuitability{
		BlockedCategories: []string{"IAB25-3"},
	}

	result := svc.evaluateBrandSuitability(req, bs)

	if !result.blocked {
		t.Error("Expected blocked due to blocked category")
	}
	if !result.safe == false {
		t.Error("Expected safe=false")
	}
}

func TestEvaluateBrandSuitability_AllowedCategoryMismatch(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"content_categories": []interface{}{"IAB9"}, // hobbies
		},
	}
	bs := &model.BrandSuitability{
		AllowedCategories: []string{"IAB1"}, // only news allowed
	}

	result := svc.evaluateBrandSuitability(req, bs)

	if !result.blocked {
		t.Error("Expected blocked due to category not in allowlist")
	}
}

func TestEvaluateBrandSuitability_AllowedCategoryMatch(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"content_categories": []interface{}{"IAB1"},
		},
	}
	bs := &model.BrandSuitability{
		AllowedCategories: []string{"IAB1"},
	}

	result := svc.evaluateBrandSuitability(req, bs)

	if result.blocked {
		t.Errorf("Expected not blocked, got: %s", result.reason)
	}
}

func TestEvaluateBrandSuitability_BlockedKeyword(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"page_content": "This page contains gambling content",
		},
	}
	bs := &model.BrandSuitability{
		CustomKeywordBlock: []string{"gambling"},
	}

	result := svc.evaluateBrandSuitability(req, bs)

	if !result.blocked {
		t.Error("Expected blocked due to keyword match")
	}
}

func TestEvaluateBrandSuitability_SentimentFilter(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"sentiment": "negative",
		},
	}
	bs := &model.BrandSuitability{
		SentimentFilters: []string{"negative", "controversial"},
	}

	result := svc.evaluateBrandSuitability(req, bs)

	if !result.blocked {
		t.Error("Expected blocked due to sentiment filter")
	}
}

func TestEvaluateBrandSuitability_FloorRatingBlocked(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"content_rating": "R",
		},
	}
	bs := &model.BrandSuitability{
		FloorRating: "PG", // only allow G or PG
	}

	result := svc.evaluateBrandSuitability(req, bs)

	if !result.blocked {
		t.Error("Expected blocked due to content rating exceeding floor")
	}
}

func TestEvaluateBrandSuitability_FloorRatingPasses(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"content_rating": "G",
		},
	}
	bs := &model.BrandSuitability{
		FloorRating: "R", // allow up to R
	}

	result := svc.evaluateBrandSuitability(req, bs)

	if result.blocked {
		t.Errorf("Expected not blocked, got: %s", result.reason)
	}
}

// ============================================================================
// isRatingAllowed tests
// ============================================================================

func TestIsRatingAllowed_Below(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	// G (1) <= PG (2) → allowed
	if !svc.isRatingAllowed("G", "PG") {
		t.Error("G should be allowed with PG floor")
	}
}

func TestIsRatingAllowed_Equal(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	// R (4) <= R (4) → allowed
	if !svc.isRatingAllowed("R", "R") {
		t.Error("R should be allowed with R floor")
	}
}

func TestIsRatingAllowed_Above(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	// NC17 (5) > PG (2) → not allowed
	if svc.isRatingAllowed("NC17", "PG") {
		t.Error("NC17 should not be allowed with PG floor")
	}
}

func TestIsRatingAllowed_UnknownRating(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	// Unknown ratings should pass
	if !svc.isRatingAllowed("UNKNOWN", "PG") {
		t.Error("Unknown rating should pass")
	}
}

func TestIsRatingAllowed_PG13Alias(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	// PG-13 (3) <= R (4) → allowed
	if !svc.isRatingAllowed("PG-13", "R") {
		t.Error("PG-13 should be allowed with R floor")
	}
	// PG-13 (3) > PG (2) → not allowed
	if svc.isRatingAllowed("PG-13", "PG") {
		t.Error("PG-13 should not be allowed with PG floor")
	}
}

// ============================================================================
// evaluateFraudProtection tests
// ============================================================================

func TestEvaluateFraudProtection_TrustScoreTooLow(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{}
	fp := &model.FraudProtection{
		MinTrustScore: 0.8,
	}
	// fraudRisk=0.5 → trustScore=0.5 < 0.8
	ctx := inventoryQualityContext{fraudRisk: 0.5}

	result := svc.evaluateFraudProtection(req, fp, ctx)

	if !result.blocked {
		t.Error("Expected blocked due to low trust score")
	}
}

func TestEvaluateFraudProtection_BotTraffic(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{}
	fp := &model.FraudProtection{BlockBotTraffic: true}
	ctx := inventoryQualityContext{botProbability: 0.9}

	result := svc.evaluateFraudProtection(req, fp, ctx)

	if !result.blocked {
		t.Error("Expected blocked due to bot traffic")
	}
}

func TestEvaluateFraudProtection_BotTrafficBelow(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{}
	fp := &model.FraudProtection{BlockBotTraffic: true}
	ctx := inventoryQualityContext{botProbability: 0.3} // below 0.7 threshold

	result := svc.evaluateFraudProtection(req, fp, ctx)

	if result.blocked {
		t.Error("Expected not blocked: bot probability is low")
	}
}

func TestEvaluateFraudProtection_ProxyTraffic(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{}
	fp := &model.FraudProtection{BlockProxyTraffic: true}
	ctx := inventoryQualityContext{proxyDetected: true}

	result := svc.evaluateFraudProtection(req, fp, ctx)

	if !result.blocked {
		t.Error("Expected blocked due to proxy traffic")
	}
}

func TestEvaluateFraudProtection_BlockedSource(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"traffic_source": "bad-network",
		},
	}
	fp := &model.FraudProtection{
		BlockedSources: []string{"bad-network"},
	}
	ctx := inventoryQualityContext{}

	result := svc.evaluateFraudProtection(req, fp, ctx)

	if !result.blocked {
		t.Error("Expected blocked due to blocked source")
	}
}

func TestEvaluateFraudProtection_HighRiskDiscount(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{}
	fp := &model.FraudProtection{}
	ctx := inventoryQualityContext{fraudRisk: 0.5}

	result := svc.evaluateFraudProtection(req, fp, ctx)

	if result.blocked {
		t.Error("Expected not blocked")
	}
	if result.multiplier >= 1.0 {
		t.Errorf("Expected multiplier < 1.0 for high risk, got %f", result.multiplier)
	}
}

func TestEvaluateFraudProtection_LowRisk(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	req := &model.BidRequest{}
	fp := &model.FraudProtection{}
	ctx := inventoryQualityContext{fraudRisk: 0.1}

	result := svc.evaluateFraudProtection(req, fp, ctx)

	if result.blocked {
		t.Error("Expected not blocked")
	}
	if result.multiplier != 1.0 {
		t.Errorf("Expected multiplier 1.0 for low risk, got %f", result.multiplier)
	}
}

// ============================================================================
// evaluateViewabilityHistory tests
// ============================================================================

func TestEvaluateViewabilityHistory_TooLow(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	ctx := inventoryQualityContext{viewabilityRate: 0.3}
	vh := &model.ViewabilityHistory{MinHistoricalRate: 0.5}

	result := svc.evaluateViewabilityHistory(ctx, vh)

	if !result.blocked {
		t.Error("Expected blocked due to low viewability")
	}
}

func TestEvaluateViewabilityHistory_HighViewBoost(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	ctx := inventoryQualityContext{viewabilityRate: 0.8}
	vh := &model.ViewabilityHistory{
		MinHistoricalRate: 0.5,
		HighViewBoost:     1.3,
	}

	result := svc.evaluateViewabilityHistory(ctx, vh)

	if result.blocked {
		t.Error("Expected not blocked")
	}
	if result.multiplier != 1.3 {
		t.Errorf("Expected multiplier 1.3, got %f", result.multiplier)
	}
}

func TestEvaluateViewabilityHistory_LowViewPenalty(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	ctx := inventoryQualityContext{viewabilityRate: 0.3}
	vh := &model.ViewabilityHistory{
		MinHistoricalRate: 0.0, // no min, but penalize low
		LowViewPenalty:    0.7,
	}

	result := svc.evaluateViewabilityHistory(ctx, vh)

	if result.blocked {
		t.Error("Expected not blocked")
	}
	if result.multiplier != 0.7 {
		t.Errorf("Expected multiplier 0.7, got %f", result.multiplier)
	}
}

func TestEvaluateViewabilityHistory_NoAdjustment(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	ctx := inventoryQualityContext{viewabilityRate: 0.55}
	vh := &model.ViewabilityHistory{MinHistoricalRate: 0.5}

	result := svc.evaluateViewabilityHistory(ctx, vh)

	if result.blocked {
		t.Error("Expected not blocked")
	}
	if result.multiplier != 1.0 {
		t.Errorf("Expected default multiplier 1.0, got %f", result.multiplier)
	}
}

// ============================================================================
// applyQualityTier tests
// ============================================================================

func TestApplyQualityTier_PremiumMatch(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	tiers := []model.QualityTier{
		{Tier: "remnant", MinScore: 0.0, MaxScore: 0.4, BidMultiplier: 0.7},
		{Tier: "standard", MinScore: 0.4, MaxScore: 0.7, BidMultiplier: 1.0},
		{Tier: "premium", MinScore: 0.7, MaxScore: 1.0, BidMultiplier: 1.4},
	}

	result := svc.applyQualityTier(0.85, tiers)

	if result.tier != "premium" {
		t.Errorf("Expected tier 'premium', got '%s'", result.tier)
	}
	if result.multiplier != 1.4 {
		t.Errorf("Expected multiplier 1.4, got %f", result.multiplier)
	}
	if !result.matched {
		t.Error("Expected matched=true")
	}
}

func TestApplyQualityTier_NoMatch(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	tiers := []model.QualityTier{
		{Tier: "premium", MinScore: 0.8, MaxScore: 1.0, BidMultiplier: 1.5},
	}

	// Score 0.5 doesn't match premium tier
	result := svc.applyQualityTier(0.5, tiers)

	if result.tier != "standard" {
		t.Errorf("Expected default tier 'standard', got '%s'", result.tier)
	}
	if result.matched {
		t.Error("Expected matched=false")
	}
	if result.multiplier != 1.0 {
		t.Errorf("Expected default multiplier 1.0, got %f", result.multiplier)
	}
}

func TestApplyQualityTier_MaxBidIncreaseCap(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	tiers := []model.QualityTier{
		{Tier: "premium", MinScore: 0.8, MaxScore: 1.0, BidMultiplier: 2.0, MaxBidIncrease: 0.3},
	}

	result := svc.applyQualityTier(0.9, tiers)

	// Max increase is 0.3, so capped at 1.3
	if result.multiplier != 1.3 {
		t.Errorf("Expected capped multiplier 1.3, got %f", result.multiplier)
	}
}

func TestApplyQualityTier_EmptyTiers(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	result := svc.applyQualityTier(0.9, nil)

	if result.tier != "standard" {
		t.Errorf("Expected default tier 'standard', got '%s'", result.tier)
	}
	if result.multiplier != 1.0 {
		t.Errorf("Expected default multiplier 1.0, got %f", result.multiplier)
	}
}

func TestApplyQualityTier_ZeroMultiplier(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	// BidMultiplier <= 0 should default to 1.0
	tiers := []model.QualityTier{
		{Tier: "broken", MinScore: 0.0, MaxScore: 1.0, BidMultiplier: 0},
	}

	result := svc.applyQualityTier(0.5, tiers)

	if result.multiplier != 1.0 {
		t.Errorf("Expected default multiplier 1.0 for zero BidMultiplier, got %f", result.multiplier)
	}
}

// ============================================================================
// findBestDeal tests
// ============================================================================

func TestFindBestDeal_EmptyDeals(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	campaign := &model.Campaign{BidPrice: 2.0}
	dt := &model.DealTargeting{}

	result := svc.findBestDeal(campaign, nil, dt)

	if result != nil {
		t.Error("Expected nil for empty deals")
	}
}

func TestFindBestDeal_BidFloorTooHigh(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	campaign := &model.Campaign{BidPrice: 1.0}
	deals := []model.Deal{
		{ID: "deal-1", BidFloor: 2.0}, // bid price too low
	}
	dt := &model.DealTargeting{}

	result := svc.findBestDeal(campaign, deals, dt)

	if result != nil {
		t.Error("Expected nil when bid below floor")
	}
}

func TestFindBestDeal_ExcludedDeal(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	campaign := &model.Campaign{BidPrice: 5.0}
	deals := []model.Deal{
		{ID: "deal-1", BidFloor: 1.0},
	}
	dt := &model.DealTargeting{
		ExcludedDealIDs: []string{"deal-1"},
	}

	result := svc.findBestDeal(campaign, deals, dt)

	if result != nil {
		t.Error("Expected nil for excluded deal")
	}
}

func TestFindBestDeal_PreferredDeal(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	campaign := &model.Campaign{BidPrice: 5.0}
	deals := []model.Deal{
		{ID: "deal-1", BidFloor: 1.0},
		{ID: "deal-preferred", BidFloor: 1.0},
	}
	dt := &model.DealTargeting{
		PreferredDealIDs: []string{"deal-preferred"},
	}

	result := svc.findBestDeal(campaign, deals, dt)

	if result == nil {
		t.Fatal("Expected a deal to be found")
	}
	if result.ID != "deal-preferred" {
		t.Errorf("Expected preferred deal, got %s", result.ID)
	}
}

func TestFindBestDeal_MinPriority(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	campaign := &model.Campaign{BidPrice: 5.0}
	deals := []model.Deal{
		{ID: "deal-low", BidFloor: 1.0}, // default priority ~6
	}
	dt := &model.DealTargeting{
		MinDealPriority: 100, // much higher than any deal's natural priority
	}

	result := svc.findBestDeal(campaign, deals, dt)

	if result != nil {
		t.Error("Expected nil when deal priority below minimum")
	}
}

func TestFindBestDeal_ValidDeal(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	campaign := &model.Campaign{BidPrice: 5.0}
	deals := []model.Deal{
		{ID: "deal-1", BidFloor: 1.0},
	}
	dt := &model.DealTargeting{}

	result := svc.findBestDeal(campaign, deals, dt)

	if result == nil {
		t.Fatal("Expected a deal to be found")
	}
	if result.ID != "deal-1" {
		t.Errorf("Expected deal-1, got %s", result.ID)
	}
}

// ============================================================================
// calculateGoalPacingMultiplier tests
// ============================================================================

func TestCalculateGoalPacingMultiplier_NoGoal(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	campaign := &model.Campaign{GoalTarget: 0}

	result := svc.calculateGoalPacingMultiplier(campaign)

	if result != 1.0 {
		t.Errorf("Expected 1.0 for no goal, got %f", result)
	}
}

func TestCalculateGoalPacingMultiplier_InvalidDate(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	campaign := &model.Campaign{
		GoalTarget:  10000,
		GoalEndDate: "not-a-date",
	}

	result := svc.calculateGoalPacingMultiplier(campaign)

	if result != 1.0 {
		t.Errorf("Expected 1.0 for invalid date, got %f", result)
	}
}

func TestCalculateGoalPacingMultiplier_PastDeadline(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	campaign := &model.Campaign{
		GoalTarget:  10000,
		GoalEndDate: "2020-01-01", // in the past
	}

	result := svc.calculateGoalPacingMultiplier(campaign)

	if result != 0.5 {
		t.Errorf("Expected 0.5 for past deadline, got %f", result)
	}
}

func TestCalculateGoalPacingMultiplier_GoalAlreadyMet(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	tomorrow := time.Now().Add(48 * time.Hour).Format("2006-01-02")
	campaign := &model.Campaign{
		GoalTarget:    1000,
		GoalDelivered: 1000, // already met
		GoalEndDate:   tomorrow,
	}

	result := svc.calculateGoalPacingMultiplier(campaign)

	if result != 0.3 {
		t.Errorf("Expected 0.3 for met goal, got %f", result)
	}
}

func TestCalculateGoalPacingMultiplier_FutureDeadline(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	nextWeek := time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02")
	campaign := &model.Campaign{
		GoalTarget:    100000,
		GoalDelivered: 0,
		GoalEndDate:   nextWeek,
	}

	result := svc.calculateGoalPacingMultiplier(campaign)

	// Far behind schedule → should get a boost
	if result < 1.0 {
		t.Errorf("Expected multiplier >= 1.0 when behind schedule, got %f", result)
	}
}

func TestCalculateGoalPacingMultiplier_NoGoalEndDate(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "http://localhost:8080")
	campaign := &model.Campaign{
		GoalTarget:  1000,
		GoalEndDate: "", // empty
	}

	result := svc.calculateGoalPacingMultiplier(campaign)

	if result != 1.0 {
		t.Errorf("Expected 1.0 for empty end date, got %f", result)
	}
}

// Ensure the helper function name reference compiles
var _ = fmt.Sprintf
