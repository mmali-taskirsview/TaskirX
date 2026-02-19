package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

func newAMCampaign(am *model.AudienceModeling) *model.Campaign {
	return &model.Campaign{
		ID:       "camp-am",
		BidPrice: 5.0,
		Budget:   1000,
		Status:   "active",
		Targeting: model.Targeting{
			PerformanceGoals: &model.PerformanceGoals{
				AudienceModeling: am,
			},
		},
	}
}

func newAMRequest(userID string, ctx map[string]interface{}) *model.BidRequest {
	return &model.BidRequest{
		ID:      "req-am",
		User:    model.InternalUser{ID: userID, Country: "US"},
		Device:  model.InternalDevice{Type: "mobile", OS: "android"},
		Context: ctx,
	}
}

func TestAudienceModeling_NilConfig(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	camp := &model.Campaign{ID: "camp1"}
	req := newAMRequest("user1", nil)

	result := svc.EvaluateAudienceModeling(camp, req)
	assertNear(t, "nil_multiplier", result.Multiplier, 1.0, 0.001)
	if result.Matched {
		t.Error("expected no match with nil config")
	}
}

func TestAudienceModeling_Suppression_Segment(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	am := &model.AudienceModeling{
		SuppressionEnabled:  true,
		SuppressionSegments: []string{"existing_customer"},
	}
	camp := newAMCampaign(am)

	ctx := map[string]interface{}{
		"user_segments": []interface{}{"existing_customer", "premium"},
	}
	req := newAMRequest("user1", ctx)

	result := svc.EvaluateAudienceModeling(camp, req)
	if !result.Suppressed {
		t.Error("expected user to be suppressed by segment match")
	}
	assertNear(t, "suppressed_multiplier", result.Multiplier, 0.0, 0.001)
}

func TestAudienceModeling_Suppression_Event(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	mc.SetUserEvent("user1", "camp-am", "purchase")

	am := &model.AudienceModeling{
		SuppressionEnabled: true,
		SuppressionEvents:  []string{"purchase"},
	}
	camp := newAMCampaign(am)
	req := newAMRequest("user1", nil)

	result := svc.EvaluateAudienceModeling(camp, req)
	if !result.Suppressed {
		t.Error("expected user to be suppressed by purchase event")
	}
}

func TestAudienceModeling_Suppression_AlreadyConverted(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	am := &model.AudienceModeling{
		SuppressionEnabled: true,
	}
	camp := newAMCampaign(am)
	ctx := map[string]interface{}{
		"has_converted": true,
	}
	req := newAMRequest("user1", ctx)

	result := svc.EvaluateAudienceModeling(camp, req)
	if !result.Suppressed {
		t.Error("expected user to be suppressed by conversion flag")
	}
}

func TestAudienceModeling_SeedAudience(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	am := &model.AudienceModeling{
		SeedSegments: []string{"vip_buyers"},
	}
	camp := newAMCampaign(am)

	ctx := map[string]interface{}{
		"user_segments": []interface{}{"vip_buyers", "tech"},
	}
	req := newAMRequest("user1", ctx)

	result := svc.EvaluateAudienceModeling(camp, req)
	if !result.Matched {
		t.Error("expected seed audience match")
	}
	if result.AudienceTier != "seed" {
		t.Errorf("expected tier seed, got %s", result.AudienceTier)
	}
	assertNear(t, "seed_multiplier", result.Multiplier, 1.5, 0.001)
}

func TestAudienceModeling_Lookalike(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	// User does NOT have seed segments, but has behavioral/demographic similarity
	am := &model.AudienceModeling{
		LookalikeEnabled:    true,
		SeedSegments:        []string{"vip_seed_only"},
		SimilarityThreshold: 0.3,
		LookalikeBoost:      1.3,
		LookalikeFeatures:   []string{"behavior", "demographics"},
	}
	camp := newAMCampaign(am)

	ctx := map[string]interface{}{
		"session_duration": float64(600),
		"pages_viewed":     float64(10),
		"engagement_score": 0.8,
		"return_visitor":   true,
		"age_bracket":      "25-34",
	}
	req := newAMRequest("user1", ctx)

	result := svc.EvaluateAudienceModeling(camp, req)
	if !result.IsLookalike {
		t.Errorf("expected lookalike match, got tier=%s reason=%s", result.AudienceTier, result.Reason)
	}
	if result.SimilarityScore <= 0 {
		t.Error("expected positive similarity score")
	}
	if result.Multiplier <= 1.0 {
		t.Errorf("expected boosted multiplier, got %.2f", result.Multiplier)
	}
}

func TestAudienceModeling_Lookalike_BelowThreshold(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	am := &model.AudienceModeling{
		LookalikeEnabled:    true,
		SeedSegments:        []string{"ultra_niche_segment"},
		SimilarityThreshold: 0.99,
		LookalikeFeatures:   []string{"interests"},
	}
	camp := newAMCampaign(am)

	ctx := map[string]interface{}{
		"user_segments": []interface{}{"completely_different"},
	}
	req := newAMRequest("user1", ctx)

	result := svc.EvaluateAudienceModeling(camp, req)
	if result.IsLookalike {
		t.Error("expected no lookalike match with high threshold and different segments")
	}
}

func TestAudienceModeling_PropensityScoring(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	am := &model.AudienceModeling{
		ScoringEnabled: true,
		ScoringModel:   "propensity",
	}
	camp := newAMCampaign(am)

	ctx := map[string]interface{}{
		"engagement_score": 0.8,
		"intent_score":     0.7,
		"return_visitor":   true,
		"cart_abandoner":   true,
	}
	req := newAMRequest("user1", ctx)

	result := svc.EvaluateAudienceModeling(camp, req)
	if !result.Matched {
		t.Error("expected propensity match")
	}
	if result.PropensityScore <= 0 {
		t.Error("expected positive propensity score")
	}
	if result.Multiplier <= 1.0 {
		t.Errorf("expected high multiplier for high-intent user, got %.2f", result.Multiplier)
	}
}

func TestAudienceModeling_PropensityScoring_LTV(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	am := &model.AudienceModeling{
		ScoringEnabled: true,
		ScoringModel:   "ltv",
	}
	camp := newAMCampaign(am)

	ctx := map[string]interface{}{
		"purchase_count":       float64(10),
		"avg_order_value":      float64(150),
		"days_since_purchase":  float64(3),
		"customer_tenure_days": float64(400),
	}
	req := newAMRequest("user1", ctx)

	result := svc.EvaluateAudienceModeling(camp, req)
	if !result.Matched {
		t.Error("expected LTV match")
	}
	if result.AudienceTier != "high_propensity" {
		t.Errorf("expected high_propensity tier, got %s", result.AudienceTier)
	}
}

func TestAudienceModeling_PropensityScoring_ChurnRisk(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	am := &model.AudienceModeling{
		ScoringEnabled: true,
		ScoringModel:   "churn_risk",
	}
	camp := newAMCampaign(am)

	// Mildly churning user: decreasing engagement but not fully gone
	// Base 0.7 - 0.3 (decreasing) = 0.4 → low_propensity
	ctx := map[string]interface{}{
		"engagement_trend": "decreasing",
	}
	req := newAMRequest("user1", ctx)

	result := svc.EvaluateAudienceModeling(camp, req)
	if !result.Matched {
		t.Errorf("expected churn risk match, got tier=%s reason=%s", result.AudienceTier, result.Reason)
	}
	if result.AudienceTier != "low_propensity" {
		t.Errorf("expected low_propensity tier for churn risk, got %s", result.AudienceTier)
	}
}

func TestAudienceModeling_ScoreBidMapping(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	am := &model.AudienceModeling{
		ScoringEnabled: true,
		ScoringModel:   "propensity",
		ScoreBidMapping: []model.ScoreBidRange{
			{MinScore: 0.0, MaxScore: 0.3, Multiplier: 0.5},
			{MinScore: 0.3, MaxScore: 0.7, Multiplier: 1.0},
			{MinScore: 0.7, MaxScore: 1.1, Multiplier: 2.0},
		},
	}
	camp := newAMCampaign(am)

	ctx := map[string]interface{}{
		"engagement_score": 0.9,
		"intent_score":     0.9,
		"return_visitor":   true,
		"cart_abandoner":   true,
	}
	req := newAMRequest("user1", ctx)

	result := svc.EvaluateAudienceModeling(camp, req)
	if result.Multiplier < 1.0 {
		t.Errorf("expected high multiplier from score-bid mapping for high-intent user, got %.2f", result.Multiplier)
	}
}

func TestAudienceModeling_MinScoreThreshold(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	am := &model.AudienceModeling{
		ScoringEnabled: true,
		ScoringModel:   "propensity",
		MinScore:       0.95,
	}
	camp := newAMCampaign(am)

	ctx := map[string]interface{}{}
	req := newAMRequest("user1", ctx)

	result := svc.EvaluateAudienceModeling(camp, req)
	if result.Multiplier > 0.5 {
		t.Errorf("expected low multiplier for below-threshold propensity, got %.2f", result.Multiplier)
	}
}

func TestAudienceModeling_Prospecting(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	am := &model.AudienceModeling{
		LookalikeEnabled:    true,
		SeedSegments:        []string{"specific_seed"},
		SimilarityThreshold: 0.99,
		LookalikeFeatures:   []string{"interests"},
	}
	camp := newAMCampaign(am)

	req := newAMRequest("user1", nil)

	result := svc.EvaluateAudienceModeling(camp, req)
	if result.AudienceTier != "prospecting" {
		t.Errorf("expected prospecting tier for unmatched user, got %s", result.AudienceTier)
	}
	assertNear(t, "prospecting_multiplier", result.Multiplier, 0.8, 0.001)
}

func TestAudienceModeling_InterestSimilarity(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	req := &model.BidRequest{
		User: model.InternalUser{Categories: []string{"tech", "gaming"}},
	}

	score := svc.interestSimilarity(req, []string{"tech", "gaming", "sports"}, []string{"tech", "gaming"})
	if score <= 0 {
		t.Error("expected positive interest similarity")
	}
	if score > 1.0 {
		t.Errorf("similarity should not exceed 1.0, got %.2f", score)
	}
}

func TestAudienceModeling_InterestSimilarity_NoOverlap(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)
	_ = mc

	req := &model.BidRequest{}
	score := svc.interestSimilarity(req, []string{"a", "b"}, []string{"c", "d"})
	assertNear(t, "no_overlap", score, 0.0, 0.001)
}

func TestAudienceModeling_SuppressionNoUserID(t *testing.T) {
	mc := NewMockCache()
	svc := NewAudienceModelingService(mc)

	am := &model.AudienceModeling{
		SuppressionEnabled: true,
		SuppressionEvents:  []string{"purchase"},
	}
	camp := newAMCampaign(am)
	req := newAMRequest("", nil)

	result := svc.EvaluateAudienceModeling(camp, req)
	if result.Suppressed {
		t.Error("should not suppress without user ID")
	}
}
