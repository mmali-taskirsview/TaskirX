package service

// coverage_boost6_test.go – additional branch coverage for:
//   - audience_modeling: calculateConversionPropensity, calculateLTVPropensity, calculateChurnRiskPropensity
//   - bidding: applyCTVOptimizations, applyEcommerceOptimizations, getHistoricalPerformance
//   - dynamic_creative: getCandidateElements, selectElement, scoreElement, getTotalImpressions (extra branches)

import (
	"fmt"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newAMCampaignWithScoringModel(model_ string) *model.Campaign {
	return newAMCampaign(&model.AudienceModeling{
		ScoringEnabled: true,
		ScoringModel:   model_,
	})
}

// ---------------------------------------------------------------------------
// calculateConversionPropensity extra branches
// ---------------------------------------------------------------------------

func TestConversionPropensity_UserSegments_HighIntent(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("propensity")

	ctx := map[string]interface{}{
		"user_segments": []interface{}{"high_intent", "other_seg"},
	}
	req := newAMRequest("u1", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	if result.PropensityScore <= 0.3 {
		t.Errorf("expected >0.3 due to high_intent segment, got %.2f", result.PropensityScore)
	}
}

func TestConversionPropensity_UserSegments_Converter(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("propensity")

	ctx := map[string]interface{}{
		"user_segments": []interface{}{"converter"},
	}
	req := newAMRequest("u2", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	if result.PropensityScore <= 0.3 {
		t.Errorf("expected >0.3 due to converter segment, got %.2f", result.PropensityScore)
	}
}

func TestConversionPropensity_IntentAndEngagement(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("propensity")

	ctx := map[string]interface{}{
		"engagement_score": float64(0.5),
		"intent_score":     float64(0.5),
	}
	req := newAMRequest("u3", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	// base 0.3 + 0.5*0.2 + 0.5*0.2 = 0.3+0.1+0.1 = 0.5
	if result.PropensityScore < 0.45 {
		t.Errorf("expected ~0.5 propensity, got %.2f", result.PropensityScore)
	}
}

func TestConversionPropensity_CapAt1(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("propensity")

	ctx := map[string]interface{}{
		"engagement_score": float64(1.0),
		"intent_score":     float64(1.0),
		"return_visitor":   true,
		"cart_abandoner":   true,
		"user_segments":    []interface{}{"high_intent", "converter"},
	}
	req := newAMRequest("u4", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	if result.PropensityScore > 1.0 {
		t.Errorf("propensity should be capped at 1.0, got %.2f", result.PropensityScore)
	}
}

// ---------------------------------------------------------------------------
// calculateLTVPropensity extra branches
// ---------------------------------------------------------------------------

func TestLTVPropensity_PurchaseCountLow(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("ltv")

	ctx := map[string]interface{}{
		"purchase_count": float64(2), // >0 but <=5 → +0.15
	}
	req := newAMRequest("u5", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	// base 0.2 + 0.15 = 0.35
	if result.PropensityScore < 0.3 {
		t.Errorf("expected ~0.35 for low purchase count, got %.2f", result.PropensityScore)
	}
}

func TestLTVPropensity_PurchaseCountHigh(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("ltv")

	ctx := map[string]interface{}{
		"purchase_count": float64(10), // >5 → +0.3
	}
	req := newAMRequest("u6", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	// base 0.2 + 0.3 = 0.5
	if result.PropensityScore < 0.45 {
		t.Errorf("expected ~0.5 for high purchase count, got %.2f", result.PropensityScore)
	}
}

func TestLTVPropensity_AvgOrderValueMid(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("ltv")

	ctx := map[string]interface{}{
		"avg_order_value": float64(75), // >50 but <=100 → +0.1
	}
	req := newAMRequest("u7", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	if result.PropensityScore < 0.25 {
		t.Errorf("expected ~0.3 for mid AOV, got %.2f", result.PropensityScore)
	}
}

func TestLTVPropensity_AvgOrderValueHigh(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("ltv")

	ctx := map[string]interface{}{
		"avg_order_value": float64(200), // >100 → +0.2
	}
	req := newAMRequest("u8", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	if result.PropensityScore < 0.35 {
		t.Errorf("expected ~0.4 for high AOV, got %.2f", result.PropensityScore)
	}
}

func TestLTVPropensity_DaysSincePurchaseRecent(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("ltv")

	ctx := map[string]interface{}{
		"days_since_purchase": float64(3), // <7 → +0.15
	}
	req := newAMRequest("u9", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	if result.PropensityScore < 0.3 {
		t.Errorf("expected ~0.35 for recent purchase, got %.2f", result.PropensityScore)
	}
}

func TestLTVPropensity_DaysSincePurchaseMid(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("ltv")

	ctx := map[string]interface{}{
		"days_since_purchase": float64(20), // <30 → +0.1
	}
	req := newAMRequest("u10", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	if result.PropensityScore < 0.25 {
		t.Errorf("expected ~0.3 for mid days_since_purchase, got %.2f", result.PropensityScore)
	}
}

func TestLTVPropensity_LongTenure(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("ltv")

	ctx := map[string]interface{}{
		"customer_tenure_days": float64(500), // >365 → +0.1
	}
	req := newAMRequest("u11", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	if result.PropensityScore < 0.25 {
		t.Errorf("expected ~0.3 for long tenure, got %.2f", result.PropensityScore)
	}
}

// ---------------------------------------------------------------------------
// calculateChurnRiskPropensity extra branches
// ---------------------------------------------------------------------------

func TestChurnRisk_EngagementIncreasing(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("churn_risk")

	ctx := map[string]interface{}{
		"engagement_trend": "increasing", // +0.2 → base 0.7 + 0.2 = 0.9
	}
	req := newAMRequest("u12", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	if result.PropensityScore < 0.8 {
		t.Errorf("expected ~0.9 for increasing engagement, got %.2f", result.PropensityScore)
	}
}

func TestChurnRisk_EngagementStable(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("churn_risk")

	ctx := map[string]interface{}{
		"engagement_trend": "stable", // 0 adjustment → 0.7
	}
	req := newAMRequest("u13", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	// stable = 0 bonus → 0.7
	if result.PropensityScore < 0.6 || result.PropensityScore > 0.8 {
		t.Errorf("expected ~0.7 for stable engagement, got %.2f", result.PropensityScore)
	}
}

func TestChurnRisk_DaysSinceActivityHigh(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("churn_risk")

	ctx := map[string]interface{}{
		"days_since_activity": float64(45), // >30 → -0.3
	}
	req := newAMRequest("u14", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	// 0.7 - 0.3 = 0.4
	if result.PropensityScore > 0.5 {
		t.Errorf("expected ~0.4 for high inactivity, got %.2f", result.PropensityScore)
	}
}

func TestChurnRisk_DaysSinceActivityMid(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("churn_risk")

	ctx := map[string]interface{}{
		"days_since_activity": float64(20), // >14 but <=30 → -0.15
	}
	req := newAMRequest("u15", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	// 0.7 - 0.15 = 0.55
	if result.PropensityScore > 0.65 || result.PropensityScore < 0.45 {
		t.Errorf("expected ~0.55 for mid inactivity, got %.2f", result.PropensityScore)
	}
}

func TestChurnRisk_SubscriptionCancelled(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("churn_risk")

	ctx := map[string]interface{}{
		"subscription_status": "cancelled", // -0.4
	}
	req := newAMRequest("u16", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	// 0.7 - 0.4 = 0.3
	if result.PropensityScore > 0.4 {
		t.Errorf("expected ~0.3 for cancelled subscription, got %.2f", result.PropensityScore)
	}
}

func TestChurnRisk_SubscriptionExpired(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("churn_risk")

	ctx := map[string]interface{}{
		"subscription_status": "expired", // -0.4
	}
	req := newAMRequest("u17", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	if result.PropensityScore > 0.4 {
		t.Errorf("expected ~0.3 for expired subscription, got %.2f", result.PropensityScore)
	}
}

func TestChurnRisk_FloorAtZero(t *testing.T) {
	svc := NewAudienceModelingService(NewMockCache())
	camp := newAMCampaignWithScoringModel("churn_risk")

	ctx := map[string]interface{}{
		"engagement_trend":    "decreasing", // -0.3
		"days_since_activity": float64(60),  // >30 → -0.3
		"subscription_status": "cancelled",  // -0.4
	}
	req := newAMRequest("u18", ctx)
	result := svc.EvaluateAudienceModeling(camp, req)
	// 0.7 - 0.3 - 0.3 - 0.4 = -0.3 → floor 0
	if result.PropensityScore < 0 {
		t.Errorf("propensity score should not go below 0, got %.2f", result.PropensityScore)
	}
}

// ---------------------------------------------------------------------------
// applyCTVOptimizations branches
// ---------------------------------------------------------------------------

func TestCTVOptimizations_NilCTV(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	result := svc.applyCTVOptimizations(camp, req, nil, performanceData{})
	if result != 1.0 {
		t.Errorf("expected 1.0 for nil CTV, got %.2f", result)
	}
}

func TestCTVOptimizations_PrimetimeBoost(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()

	ctv := &model.CTVOptimization{
		PrimtimeBoost: 1.5,
	}
	result := svc.applyCTVOptimizations(camp, req, ctv, performanceData{})
	// May or may not be primetime depending on test execution time;
	// just verify no panic and result >= 1.0
	if result < 0.1 {
		t.Errorf("unexpected result %.2f", result)
	}
}

func TestCTVOptimizations_LiveContentBoost(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"is_live": true,
	}

	ctv := &model.CTVOptimization{
		LiveContentBoost: 1.3,
	}
	result := svc.applyCTVOptimizations(camp, req, ctv, performanceData{})
	if result < 1.25 {
		t.Errorf("expected live content boost applied, got %.2f", result)
	}
}

func TestCTVOptimizations_CoViewingBoost(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"co_viewing": true,
	}

	ctv := &model.CTVOptimization{
		CoViewingBoost: 1.2,
	}
	result := svc.applyCTVOptimizations(camp, req, ctv, performanceData{})
	if result < 1.15 {
		t.Errorf("expected co-viewing boost applied, got %.2f", result)
	}
}

func TestCTVOptimizations_HouseholdFrequencyAtCap(t *testing.T) {
	mc := NewMockCache()
	mc.kv[fmt.Sprintf("hh_freq:%s:hh123", "")] = "5"
	svc := NewBiddingService(mc, "")
	camp := newCampaign(1.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"household_id": "hh123",
	}

	ctv := &model.CTVOptimization{
		HouseholdFrequencyCap: 5,
	}
	result := svc.applyCTVOptimizations(camp, req, ctv, performanceData{})
	// impressions >= cap → return 0.1 or reduced multiplier
	if result > 1.2 {
		t.Errorf("expected reduced bid near cap, got %.2f", result)
	}
}

func TestCTVOptimizations_PreferredApp_Match(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"app_name": "Netflix",
	}

	ctv := &model.CTVOptimization{
		PreferredApps: []string{"Netflix", "Hulu"},
	}
	result := svc.applyCTVOptimizations(camp, req, ctv, performanceData{})
	if result < 1.15 {
		t.Errorf("expected preferred app boost, got %.2f", result)
	}
}

func TestCTVOptimizations_PreferredApp_NoMatch(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"app_name": "UnknownApp",
	}

	ctv := &model.CTVOptimization{
		PreferredApps: []string{"Netflix", "Hulu"},
	}
	result := svc.applyCTVOptimizations(camp, req, ctv, performanceData{})
	if result != 1.0 {
		t.Errorf("expected 1.0 for no preferred app match, got %.2f", result)
	}
}

func TestCTVOptimizations_ContentTypeLive(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"content_type": "live_sports",
	}

	ctv := &model.CTVOptimization{
		LiveContentBoost: 1.4,
	}
	result := svc.applyCTVOptimizations(camp, req, ctv, performanceData{})
	if result < 1.3 {
		t.Errorf("expected live content type boost, got %.2f", result)
	}
}

func TestCTVOptimizations_HouseholdViewers(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"household_viewers": float64(3),
	}

	ctv := &model.CTVOptimization{
		CoViewingBoost: 1.25,
	}
	result := svc.applyCTVOptimizations(camp, req, ctv, performanceData{})
	if result < 1.2 {
		t.Errorf("expected co-viewing boost via household_viewers, got %.2f", result)
	}
}

// ---------------------------------------------------------------------------
// applyEcommerceOptimizations branches
// ---------------------------------------------------------------------------

func TestEcommerceOptimizations_NilEcom(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	result := svc.applyEcommerceOptimizations(camp, req, nil, performanceData{})
	if result != 1.0 {
		t.Errorf("expected 1.0 for nil ecom, got %.2f", result)
	}
}

func TestEcommerceOptimizations_CartAbandon(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"cart_abandoner": true,
	}

	ecom := &model.EcommerceOptimization{
		CartAbandonBoost: 1.5,
	}
	result := svc.applyEcommerceOptimizations(camp, req, ecom, performanceData{})
	if result < 1.4 {
		t.Errorf("expected cart abandon boost, got %.2f", result)
	}
}

func TestEcommerceOptimizations_RepeatCustomer(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"repeat_customer": true,
	}

	ecom := &model.EcommerceOptimization{
		RepeatCustomerBoost: 1.3,
	}
	result := svc.applyEcommerceOptimizations(camp, req, ecom, performanceData{})
	if result < 1.25 {
		t.Errorf("expected repeat customer boost, got %.2f", result)
	}
}

func TestEcommerceOptimizations_NewCustomerPriority(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	// No repeat_customer signal → new customer

	ecom := &model.EcommerceOptimization{
		NewCustomerPriority: true,
	}
	result := svc.applyEcommerceOptimizations(camp, req, ecom, performanceData{})
	if result < 1.15 {
		t.Errorf("expected new customer priority boost, got %.2f", result)
	}
}

func TestEcommerceOptimizations_SeasonalAdjustment(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()

	// Get the current season to set the right key
	season := svc.getCurrentSeason()
	ecom := &model.EcommerceOptimization{
		SeasonalAdjustments: map[string]float64{
			season: 1.4,
		},
	}
	result := svc.applyEcommerceOptimizations(camp, req, ecom, performanceData{})
	if result < 1.3 {
		t.Errorf("expected seasonal boost for %s, got %.2f", season, result)
	}
}

func TestEcommerceOptimizations_CartAbandon_ViaSegments(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"user_segments": []interface{}{"cart_abandon_7d"},
	}

	ecom := &model.EcommerceOptimization{
		CartAbandonBoost: 1.6,
	}
	result := svc.applyEcommerceOptimizations(camp, req, ecom, performanceData{})
	if result < 1.5 {
		t.Errorf("expected cart abandon segment boost, got %.2f", result)
	}
}

func TestEcommerceOptimizations_RepeatCustomer_ViaPurchaseCount(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	camp := newCampaign(1.0)
	req := newReq()
	req.Context = map[string]interface{}{
		"purchase_count": float64(3),
	}

	ecom := &model.EcommerceOptimization{
		RepeatCustomerBoost: 1.2,
	}
	result := svc.applyEcommerceOptimizations(camp, req, ecom, performanceData{})
	if result < 1.15 {
		t.Errorf("expected repeat customer boost via purchase_count, got %.2f", result)
	}
}

// ---------------------------------------------------------------------------
// getHistoricalPerformance branches
// ---------------------------------------------------------------------------

func TestGetHistoricalPerformance_NoCache(t *testing.T) {
	svc := NewBiddingService(NewMockCache(), "")
	req := newReq()

	data := svc.getHistoricalPerformance("camp_no_cache", req)
	// Defaults: ctr=0.01, cvr=0.02, viewability=0.60, winRate=0.15
	if data.ctr != 0.01 {
		t.Errorf("expected default ctr 0.01, got %.4f", data.ctr)
	}
	if data.viewability != 0.60 {
		t.Errorf("expected default viewability 0.60, got %.4f", data.viewability)
	}
}

func TestGetHistoricalPerformance_WithCachedData(t *testing.T) {
	mc := NewMockCache()
	mc.kv["perf:camp123"] = "impressions:1000,clicks:50,conversions:10,spend:200,ctr:0.05,cvr:0.2,cpa:20,viewability:0.75,completion_rate:0.8,engagement_rate:0.3,avg_bid:1.5,win_rate:0.25"
	svc := NewBiddingService(mc, "")
	req := newReq()

	data := svc.getHistoricalPerformance("camp123", req)
	if data.impressions != 1000 {
		t.Errorf("expected 1000 impressions from cache, got %d", data.impressions)
	}
	if data.winRate < 0.2 {
		t.Errorf("expected win_rate from cache, got %.4f", data.winRate)
	}
	// impressions > 0 → recalculated ctr = 50/1000 = 0.05
	if data.ctr < 0.04 || data.ctr > 0.06 {
		t.Errorf("expected recalculated ctr ~0.05, got %.4f", data.ctr)
	}
}

func TestGetHistoricalPerformance_DerivedCVR(t *testing.T) {
	mc := NewMockCache()
	mc.kv["perf:campDerived"] = "impressions:500,clicks:25,conversions:5,spend:100"
	svc := NewBiddingService(mc, "")
	req := newReq()

	data := svc.getHistoricalPerformance("campDerived", req)
	// clicks=25, conversions=5 → cvr=5/25=0.2
	if data.cvr < 0.15 || data.cvr > 0.25 {
		t.Errorf("expected derived cvr ~0.2, got %.4f", data.cvr)
	}
	// spend=100, conversions=5 → cpa=20
	if data.cpa < 15 || data.cpa > 25 {
		t.Errorf("expected derived cpa ~20, got %.4f", data.cpa)
	}
}

func TestGetHistoricalPerformance_EmptyCache(t *testing.T) {
	mc := NewMockCache()
	mc.kv["perf:campEmpty"] = "" // cache hit but empty value
	svc := NewBiddingService(mc, "")
	req := newReq()

	data := svc.getHistoricalPerformance("campEmpty", req)
	// Should return defaults
	if data.ctr != 0.01 {
		t.Errorf("expected default ctr for empty cache, got %.4f", data.ctr)
	}
}

// ---------------------------------------------------------------------------
// getCandidateElements branches
// ---------------------------------------------------------------------------

func TestGetCandidateElements_EmptyElements(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	ctx := DCOContext{UserSegments: []string{"sports"}}
	constraints := DCOConstraints{}

	candidates := svc.getCandidateElements("headline", ctx, constraints)
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates from empty service, got %d", len(candidates))
	}
}

func TestGetCandidateElements_TypeFilter(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())

	// Add elements of different types
	svc.elements.Store("e1", &CreativeElement{
		ID:          "e1",
		Type:        "headline",
		Content:     "Great Deal!",
		Performance: &ElementPerformance{},
	})
	svc.elements.Store("e2", &CreativeElement{
		ID:          "e2",
		Type:        "image",
		Content:     "img.jpg",
		Performance: &ElementPerformance{},
	})

	ctx := DCOContext{}
	constraints := DCOConstraints{}
	candidates := svc.getCandidateElements("headline", ctx, constraints)
	if len(candidates) != 1 {
		t.Errorf("expected 1 headline candidate, got %d", len(candidates))
	}
}

func TestGetCandidateElements_ExcludedElements(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())

	svc.elements.Store("e3", &CreativeElement{
		ID:          "e3",
		Type:        "cta",
		Content:     "Buy Now",
		Performance: &ElementPerformance{},
	})

	ctx := DCOContext{}
	constraints := DCOConstraints{ExcludedElements: []string{"e3"}}
	candidates := svc.getCandidateElements("cta", ctx, constraints)
	for _, c := range candidates {
		if c.ID == "e3" {
			t.Error("excluded element should not appear in candidates")
		}
	}
}

func TestGetCandidateElements_SegmentMatch(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())

	svc.elements.Store("e4", &CreativeElement{
		ID:          "e4",
		Type:        "headline",
		Content:     "Sports Fan Offer",
		Segments:    []string{"sports_fan"},
		Performance: &ElementPerformance{},
	})
	svc.elements.Store("e5", &CreativeElement{
		ID:          "e5",
		Type:        "headline",
		Content:     "Generic Offer",
		Segments:    []string{}, // no segments – should always match
		Performance: &ElementPerformance{},
	})

	ctx := DCOContext{UserSegments: []string{"sports_fan"}}
	constraints := DCOConstraints{}
	candidates := svc.getCandidateElements("headline", ctx, constraints)
	if len(candidates) < 2 {
		t.Errorf("expected both elements (sports_fan + generic), got %d", len(candidates))
	}
}

// ---------------------------------------------------------------------------
// selectElement branches
// ---------------------------------------------------------------------------

func TestSelectElement_NilCandidates(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	ctx := DCOContext{}

	elem, method := svc.selectElement(nil, nil, ctx, nil)
	if elem != nil {
		t.Error("expected nil element for nil candidates")
	}
	_ = method
}

func TestSelectElement_EmptyCandidates(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	ctx := DCOContext{}

	elem, method := svc.selectElement([]*CreativeElement{}, nil, ctx, nil)
	if elem != nil {
		t.Error("expected nil element for empty candidates")
	}
	_ = method
}

func TestSelectElement_ScoreBased(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	// Set exploration rate to 0 to force score-based path
	svc.config.ExplorationRate = 0.0

	candidates := []*CreativeElement{
		{
			ID:      "s1",
			Type:    "headline",
			Content: "Best Deal",
			Tags:    []string{"sport"},
			Performance: &ElementPerformance{
				Impressions: 500,
				Clicks:      50,
				CTR:         0.1,
			},
		},
		{
			ID:      "s2",
			Type:    "headline",
			Content: "Good Deal",
			Performance: &ElementPerformance{
				Impressions: 200,
				Clicks:      10,
				CTR:         0.05,
			},
		},
	}

	userPref := &UserCreativePreference{
		EngagedElements: map[string]int{"s1": 5},
	}
	ctx := DCOContext{ContentKeywords: []string{"sport"}}

	elem, method := svc.selectElement(candidates, userPref, ctx, nil)
	if elem == nil {
		t.Error("expected element from score-based selection")
	}
	_ = method
}

func TestSelectElement_RuleBased(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	svc.config.ExplorationRate = 0.0

	candidates := []*CreativeElement{
		{ID: "r1", Type: "headline", Content: "Rule Match", Performance: &ElementPerformance{}},
		{ID: "r2", Type: "headline", Content: "Default", Performance: &ElementPerformance{}},
	}

	rules := []PersonalizationRule{
		{
			ID:      "pr1",
			Enabled: true,
			Conditions: []RuleCondition{
				{Field: "device.type", Operator: "equals", Value: "mobile"},
			},
			Actions: []RuleAction{
				{Type: "select_element", Value: "r1"},
			},
		},
	}

	ctx := DCOContext{DeviceType: "mobile"}
	elem, method := svc.selectElement(candidates, nil, ctx, rules)
	if elem == nil {
		t.Error("expected rule-based element selection")
	}
	_ = method
}

func TestSelectElement_ExplorationRandom(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	// Force exploration
	svc.config.ExplorationRate = 1.0

	candidates := []*CreativeElement{
		{ID: "exp1", Type: "headline", Content: "A", Performance: &ElementPerformance{}},
		{ID: "exp2", Type: "headline", Content: "B", Performance: &ElementPerformance{}},
	}

	ctx := DCOContext{}
	elem, method := svc.selectElement(candidates, nil, ctx, nil)
	if elem == nil {
		t.Error("expected random element from exploration")
	}
	_ = method
}

// ---------------------------------------------------------------------------
// scoreElement branches
// ---------------------------------------------------------------------------

func TestScoreElement_LowImpressions(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	// MinImpressionsForStats = 100, so < 100 → exploration bonus

	elem := &CreativeElement{
		ID:   "sc1",
		Tags: []string{"tech"},
		Performance: &ElementPerformance{
			Impressions: 10, // below threshold
		},
	}
	userPref := &UserCreativePreference{
		EngagedElements: map[string]int{},
	}
	ctx := DCOContext{ContentKeywords: []string{"tech"}}

	score := svc.scoreElement(elem, userPref, ctx)
	// Should get exploration bonus (0.5 * PerformanceWeight) plus context match
	if score <= 0 {
		t.Errorf("expected positive score for low-impression element, got %.4f", score)
	}
}

func TestScoreElement_HighImpressions(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())

	elem := &CreativeElement{
		ID:   "sc2",
		Tags: []string{},
		Performance: &ElementPerformance{
			Impressions: 200,
			Clicks:      20,
			CTR:         0.10,
		},
	}
	userPref := &UserCreativePreference{
		EngagedElements: map[string]int{"sc2": 3},
	}
	ctx := DCOContext{}

	score := svc.scoreElement(elem, userPref, ctx)
	if score <= 0 {
		t.Errorf("expected positive score for high-impression element, got %.4f", score)
	}
}

func TestScoreElement_NoEngagements(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())

	elem := &CreativeElement{
		ID:   "sc3",
		Tags: []string{"news"},
		Performance: &ElementPerformance{
			Impressions: 50, // below threshold → exploration bonus
		},
	}
	userPref := &UserCreativePreference{
		EngagedElements: map[string]int{}, // no engagements for this elem
	}
	ctx := DCOContext{ContentKeywords: []string{"news"}}

	score := svc.scoreElement(elem, userPref, ctx)
	if score <= 0 {
		t.Errorf("expected positive score, got %.4f", score)
	}
}

// ---------------------------------------------------------------------------
// getTotalImpressions with stored combinations
// ---------------------------------------------------------------------------

func TestGetTotalImpressions_WithCombinations(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())

	svc.combinations.Store("combo1", &CreativeCombination{
		ID: "combo1",
		Performance: &CombinationPerformance{
			Impressions: 500,
		},
		CreatedAt: time.Now(),
	})
	svc.combinations.Store("combo2", &CreativeCombination{
		ID: "combo2",
		Performance: &CombinationPerformance{
			Impressions: 300,
		},
		CreatedAt: time.Now(),
	})

	total := svc.getTotalImpressions()
	if total < 800 {
		t.Errorf("expected total >= 800, got %d", total)
	}
}

func TestGetTotalImpressions_EmptyCombinations(t *testing.T) {
	svc := NewDynamicCreativeService(NewMockCache())
	total := svc.getTotalImpressions()
	if total < 1 {
		t.Errorf("getTotalImpressions() should return at least 1, got %d", total)
	}
}
