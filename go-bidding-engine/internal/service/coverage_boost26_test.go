package service

import (
	"fmt"
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ============================================================
// analyzeSentiment — positive, negative, neutral, no match
// ============================================================

func TestAnalyzeSentiment_Positive_B26(t *testing.T) {
	svc := NewContextualAIService(nil)
	sentiment, score := svc.analyzeSentiment("great excellent amazing wonderful content")
	if sentiment != "positive" {
		t.Errorf("Expected positive, got %s", sentiment)
	}
	if score <= 0.2 {
		t.Errorf("Expected score > 0.2, got %f", score)
	}
}

func TestAnalyzeSentiment_Negative_B26(t *testing.T) {
	svc := NewContextualAIService(nil)
	sentiment, score := svc.analyzeSentiment("terrible awful horrible disaster crisis")
	if sentiment != "negative" {
		t.Errorf("Expected negative, got %s", sentiment)
	}
	if score >= -0.2 {
		t.Errorf("Expected score < -0.2, got %f", score)
	}
}

func TestAnalyzeSentiment_Neutral_B26(t *testing.T) {
	svc := NewContextualAIService(nil)
	sentiment, score := svc.analyzeSentiment("the quick brown fox")
	if sentiment != "neutral" {
		t.Errorf("Expected neutral, got %s", sentiment)
	}
	_ = score
}

func TestAnalyzeSentiment_NoMatch_B26(t *testing.T) {
	svc := NewContextualAIService(nil)
	sentiment, score := svc.analyzeSentiment("xyzzy foo bar baz qux")
	if sentiment != "neutral" || score != 0.0 {
		t.Errorf("Expected neutral/0.0 for no-match, got %s/%f", sentiment, score)
	}
}

// ============================================================
// assessContentQuality — long, short, normal content
// ============================================================

func TestAssessContentQuality_LongContent_B26(t *testing.T) {
	svc := NewContextualAIService(nil)

	// Generate 110-word content with high diversity
	words := make([]string, 110)
	for i := range words {
		words[i] = "word" + string(rune('a'+i%26)) + string(rune('0'+i%10))
	}
	content := ""
	for _, w := range words {
		content += w + " "
	}
	quality := svc.assessContentQuality(content)
	if quality <= 0.5 {
		t.Errorf("Expected quality > 0.5 for long diverse content, got %f", quality)
	}
}

func TestAssessContentQuality_ShortContent_B26(t *testing.T) {
	svc := NewContextualAIService(nil)
	// < 20 words → quality -= 0.2, but diversity bonus partly offsets it
	// "short content": 2 words < 20 → -0.2; diversity=1.0 → +0.2; net=0.5
	// Use a single repeated word to get low diversity: quality still < 0.5 expected
	repeated := "word word word word word word word word word word word word word word word word word word word"
	// 19 words, all same → < 20 → -0.2; diversity = 1/19 ≈ 0.05 → +0.01; total ≈ 0.31
	quality := svc.assessContentQuality(repeated)
	if quality >= 0.5 {
		t.Errorf("Expected quality < 0.5 for short repetitive content, got %f", quality)
	}
}

func TestAssessContentQuality_EmptyContent_B26(t *testing.T) {
	svc := NewContextualAIService(nil)
	quality := svc.assessContentQuality("")
	// quality=0.5, words=0 → no diversity change (skipped), result=0.5
	if quality < 0 || quality > 1 {
		t.Errorf("Expected quality in [0,1], got %f", quality)
	}
}

// ============================================================
// evaluateEntityTargeting — exclude, match with multiplier, no match
// ============================================================

func TestEvaluateEntityTargeting_Exclude_B26(t *testing.T) {
	svc := NewContextualAIService(nil)

	entities := []model.DetectedEntity{
		{Type: "organization", Name: "Google"},
	}
	targets := []model.EntityTarget{
		{EntityType: "organization", Entities: []string{"Google"}, Exclude: true},
	}
	result := svc.evaluateEntityTargeting(entities, targets)
	if result != 0 {
		t.Errorf("Expected 0 (blocked) for excluded entity, got %f", result)
	}
}

func TestEvaluateEntityTargeting_MatchWithMultiplier_B26(t *testing.T) {
	svc := NewContextualAIService(nil)

	entities := []model.DetectedEntity{
		{Type: "location", Name: "New York"},
	}
	targets := []model.EntityTarget{
		{EntityType: "location", Entities: []string{"New York"}, Multiplier: 1.5},
	}
	result := svc.evaluateEntityTargeting(entities, targets)
	if result != 1.5 {
		t.Errorf("Expected 1.5 multiplier, got %f", result)
	}
}

func TestEvaluateEntityTargeting_MatchDefaultMultiplier_B26(t *testing.T) {
	svc := NewContextualAIService(nil)

	entities := []model.DetectedEntity{
		{Type: "organization", Name: "Apple"},
	}
	targets := []model.EntityTarget{
		{EntityType: "organization", Entities: []string{"Apple"}, Multiplier: 0}, // → default 1.2
	}
	result := svc.evaluateEntityTargeting(entities, targets)
	if result != 1.2 {
		t.Errorf("Expected default 1.2 multiplier, got %f", result)
	}
}

func TestEvaluateEntityTargeting_NoMatch_B26(t *testing.T) {
	svc := NewContextualAIService(nil)

	entities := []model.DetectedEntity{
		{Type: "organization", Name: "Tesla"},
	}
	targets := []model.EntityTarget{
		{EntityType: "person", Entities: []string{"Tesla"}}, // type mismatch
	}
	result := svc.evaluateEntityTargeting(entities, targets)
	if result != 1.0 {
		t.Errorf("Expected 1.0 for no match, got %f", result)
	}
}

// ============================================================
// AnalyzeContext — disabled, no content, brand safety block, entity targeting
// ============================================================

func TestAnalyzeContext_Disabled_B26(t *testing.T) {
	svc := NewContextualAIService(nil)

	campaign := &model.Campaign{
		Targeting: model.Targeting{
			ContextualAI: &model.ContextualAI{Enabled: false},
		},
	}
	result := svc.AnalyzeContext(campaign, &model.BidRequest{})
	if result.Analyzed {
		t.Error("Expected not analyzed when disabled")
	}
	if result.BidMultiplier != 1.0 {
		t.Errorf("Expected 1.0 multiplier, got %f", result.BidMultiplier)
	}
}

func TestAnalyzeContext_NilConfig_B26(t *testing.T) {
	svc := NewContextualAIService(nil)
	campaign := &model.Campaign{
		Targeting: model.Targeting{},
	}
	result := svc.AnalyzeContext(campaign, &model.BidRequest{})
	if result.Analyzed {
		t.Error("Expected not analyzed for nil config")
	}
}

func TestAnalyzeContext_NoPageContent_B26(t *testing.T) {
	svc := NewContextualAIService(nil)

	campaign := &model.Campaign{
		Targeting: model.Targeting{
			ContextualAI: &model.ContextualAI{Enabled: true},
		},
	}
	// No page content in request → "no_page_content" reason
	result := svc.AnalyzeContext(campaign, &model.BidRequest{})
	if result.Analyzed {
		t.Error("Expected not analyzed for empty content")
	}
	if result.Reason != "no_page_content" {
		t.Errorf("Expected 'no_page_content', got '%s'", result.Reason)
	}
}

func TestAnalyzeContext_BrandSafetyBlock_B26(t *testing.T) {
	svc := NewContextualAIService(nil)

	campaign := &model.Campaign{
		Targeting: model.Targeting{
			ContextualAI: &model.ContextualAI{
				Enabled:          true,
				AnalyzeSentiment: true,
			},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"page_content": "violence killing drugs illegal murder attack terrorist",
		},
	}
	result := svc.AnalyzeContext(campaign, req)
	if result.BrandSafe {
		t.Error("Expected brand safety block for violent content")
	}
	if result.BidMultiplier != 0 {
		t.Errorf("Expected bid multiplier 0 for blocked content, got %f", result.BidMultiplier)
	}
}

func TestAnalyzeContext_EntityTargeting_B26(t *testing.T) {
	svc := NewContextualAIService(nil)

	campaign := &model.Campaign{
		Targeting: model.Targeting{
			ContextualAI: &model.ContextualAI{
				Enabled:         true,
				AnalyzeEntities: true,
				EntityTargeting: []model.EntityTarget{
					{EntityType: "organization", Entities: []string{"Google"}, Multiplier: 1.5},
				},
			},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"page_content": "google is a tech company based in california",
		},
	}
	result := svc.AnalyzeContext(campaign, req)
	if !result.Analyzed {
		t.Error("Expected analyzed=true")
	}
}

func TestAnalyzeContext_SentimentTargeting_Negative_B26(t *testing.T) {
	svc := NewContextualAIService(nil)

	campaign := &model.Campaign{
		Targeting: model.Targeting{
			ContextualAI: &model.ContextualAI{
				Enabled:          true,
				AnalyzeSentiment: true,
				SentimentTargeting: &model.SentimentTargeting{
					TargetNegative: false, // Don't target negative → heavy penalty
				},
			},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"page_content": "bad terrible awful horrible worst fail problem wrong",
		},
	}
	result := svc.AnalyzeContext(campaign, req)
	if result.BidMultiplier > 0.5 {
		t.Errorf("Expected reduced multiplier for negative untargeted content, got %f", result.BidMultiplier)
	}
}

func TestAnalyzeContext_SemanticTargeting_B26(t *testing.T) {
	svc := NewContextualAIService(nil)

	campaign := &model.Campaign{
		Targeting: model.Targeting{
			ContextualAI: &model.ContextualAI{
				Enabled: true,
				SemanticTargeting: &model.SemanticTargeting{
					Enabled:             true,
					SeedContent:         []string{"sports football basketball baseball"},
					SimilarityThreshold: 0.1, // Low threshold to trigger boost
				},
			},
		},
	}
	req := &model.BidRequest{
		Context: map[string]interface{}{
			"page_content": "sports football basketball baseball game",
		},
	}
	result := svc.AnalyzeContext(campaign, req)
	if result.SemanticMatch <= 0 {
		t.Errorf("Expected semantic match > 0, got %f", result.SemanticMatch)
	}
}

// ============================================================
// GetCostAnalysis — with GetSupplyChainMetrics returning error
// ============================================================

type mockCacheWithErrorSPA struct {
	*MockCache
}

func (m *mockCacheWithErrorSPA) GetSupplyChainMetrics(timeRange string) (*model.SupplyChainMetrics, error) {
	return nil, nil // returns nil metrics
}

func TestGetCostAnalysis_NilMetrics_B26(t *testing.T) {
	mc := &mockCacheWithErrorSPA{MockCache: NewMockCache()}
	svc := NewSupplyPathAnalyticsService(mc)

	costs, err := svc.GetCostAnalysis("1h")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(costs) != 0 {
		t.Errorf("Expected empty costs for nil metrics, got %v", costs)
	}
}

// ============================================================
// CalculateCostBenefitAnalysis — nil metrics
// ============================================================

func TestCalculateCostBenefitAnalysis_NilMetrics_B26(t *testing.T) {
	mc := &mockCacheWithErrorSPA{MockCache: NewMockCache()}
	svc := NewSupplyPathAnalyticsService(mc)

	analysis, err := svc.CalculateCostBenefitAnalysis("24h")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if analysis == nil {
		t.Fatal("Expected non-nil analysis")
	}
	if len(analysis.Scenarios) != 0 {
		t.Error("Expected empty scenarios for nil metrics")
	}
}

// ============================================================
// AnalyzeDirectPublisherOpportunities — error path
// ============================================================

type mockCacheWithSPAError struct {
	*MockCache
}

func (m *mockCacheWithSPAError) GetSupplyChainMetrics(timeRange string) (*model.SupplyChainMetrics, error) {
	return nil, fmt.Errorf("storage error")
}

func TestAnalyzeDirectPublisher_Error_B26(t *testing.T) {
	mc := &mockCacheWithSPAError{MockCache: NewMockCache()}
	svc := NewSupplyPathAnalyticsService(mc)

	_, err := svc.AnalyzeDirectPublisherOpportunities("1h")
	if err == nil {
		t.Error("Expected error when GetSupplyChainMetrics fails")
	}
}

// ============================================================
// CheckPerformanceAlerts — enabled with threshold violations
// ============================================================

func TestCheckPerformanceAlerts_LowCTR_B26(t *testing.T) {
	svc := NewRealTimeAlertService(NewMockCache())

	campaign := &model.Campaign{
		ID: "camp-perf-ctr",
		Targeting: model.Targeting{
			AlertConfig: &model.AlertConfig{
				Enabled: true,
				PerformanceAlerts: &model.PerformanceAlerts{
					Enabled:        true,
					CTRDropPercent: 20.0, // Alert on 20% drop
				},
			},
		},
	}
	// Seed baseline CTR
	svc.mu.Lock()
	svc.campaignMetrics[campaign.ID] = &campaignMetricsHistory{
		baselineCTR: 0.10, // baseline 10% CTR
	}
	svc.mu.Unlock()

	svc.CheckPerformanceAlerts(campaign, 0.05, 0.02, 0.3) // CTR=0.05 → 50% drop > 20% threshold

	svc.mu.RLock()
	var found bool
	for _, alert := range svc.activeAlerts {
		if alert.CampaignID == campaign.ID && alert.Metric == "ctr" {
			found = true
		}
	}
	svc.mu.RUnlock()
	if !found {
		t.Error("Expected CTR drop alert to be created")
	}
}

func TestCheckPerformanceAlerts_Disabled_B26(t *testing.T) {
	svc := NewRealTimeAlertService(NewMockCache())

	campaign := &model.Campaign{
		ID: "camp-perf-disabled",
		Targeting: model.Targeting{
			AlertConfig: &model.AlertConfig{
				Enabled: true,
				PerformanceAlerts: &model.PerformanceAlerts{
					Enabled: false,
				},
			},
		},
	}
	svc.CheckPerformanceAlerts(campaign, 0.001, 0.001, 0.001)
	svc.mu.RLock()
	count := 0
	for _, alert := range svc.activeAlerts {
		if alert.CampaignID == campaign.ID {
			count++
		}
	}
	svc.mu.RUnlock()
	if count != 0 {
		t.Error("Expected no alerts when PerformanceAlerts disabled")
	}
}

func TestCheckPerformanceAlerts_NilConfig_B26(t *testing.T) {
	svc := NewRealTimeAlertService(NewMockCache())
	campaign := &model.Campaign{
		ID:        "camp-nilcfg",
		Targeting: model.Targeting{},
	}
	// Should not panic
	svc.CheckPerformanceAlerts(campaign, 0.5, 0.1, 0.3)
}
