package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

func createCtxCampaign(enabled bool) *model.Campaign {
	camp := &model.Campaign{
		ID:       "camp-ctx-1",
		Name:     "Contextual AI Campaign",
		BidPrice: 2.0,
		Targeting: model.Targeting{
			ContextualAI: nil,
		},
	}
	if enabled {
		camp.Targeting.ContextualAI = &model.ContextualAI{
			Enabled:          true,
			AnalyzeContent:   true,
			AnalyzeSentiment: true,
			AnalyzeEntities:  true,
			AnalyzeEmotion:   true,
			MinConfidence:    0.3,
		}
	}
	return camp
}

func createCtxRequest(content string) *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-ctx-1",
		PublisherID: "pub-ctx-123",
		Context: map[string]interface{}{
			"page_content": content,
		},
	}
}

func TestCtxAI_NewService(t *testing.T) {
	svc := NewContextualAIService(nil)
	if svc == nil {
		t.Fatal("expected service")
	}
	if svc.categoryCache == nil {
		t.Error("expected category cache")
	}
	if svc.sentimentLexicon == nil {
		t.Error("expected sentiment lexicon")
	}
	if len(svc.sentimentLexicon) == 0 {
		t.Error("expected default sentiment lexicon loaded")
	}
}

func TestCtxAI_AnalyzeContext_Disabled(t *testing.T) {
	svc := NewContextualAIService(nil)
	campaign := createCtxCampaign(false)
	req := createCtxRequest("some content")

	result := svc.AnalyzeContext(campaign, req)

	if result.Analyzed {
		t.Error("expected not analyzed when disabled")
	}
	if result.Reason != "contextual_ai_disabled" {
		t.Errorf("expected 'contextual_ai_disabled', got '%s'", result.Reason)
	}
	if result.BidMultiplier != 1.0 {
		t.Errorf("expected multiplier 1.0, got %f", result.BidMultiplier)
	}
}

func TestCtxAI_AnalyzeContext_NoContent(t *testing.T) {
	svc := NewContextualAIService(nil)
	campaign := createCtxCampaign(true)
	req := &model.BidRequest{ID: "req-1", Context: make(map[string]interface{})}

	result := svc.AnalyzeContext(campaign, req)

	if result.Analyzed {
		t.Error("expected not analyzed without content")
	}
	if result.Reason != "no_page_content" {
		t.Errorf("expected 'no_page_content', got '%s'", result.Reason)
	}
}

func TestCtxAI_AnalyzeContext_Success(t *testing.T) {
	svc := NewContextualAIService(nil)
	campaign := createCtxCampaign(true)
	req := createCtxRequest("This is a great article about sports and technology")

	result := svc.AnalyzeContext(campaign, req)

	if !result.Analyzed {
		t.Error("expected analyzed")
	}
	if result.Reason != "analysis_complete" {
		t.Errorf("expected 'analysis_complete', got '%s'", result.Reason)
	}
	if !result.BrandSafe {
		t.Error("expected brand safe")
	}
}

func TestCtxAI_ExtractPageContent_AllFields(t *testing.T) {
	svc := NewContextualAIService(nil)
	req := &model.BidRequest{
		PublisherID: "pub-123",
		Context: map[string]interface{}{
			"page_title":       "Test Title",
			"page_content":     "Main content here",
			"page_keywords":    "keyword1, keyword2",
			"page_description": "A description",
			"page_categories":  []interface{}{"sports", "news"},
		},
	}

	content := svc.extractPageContent(req)

	if content == "" {
		t.Error("expected content extracted")
	}
	if len(content) < 20 {
		t.Error("expected substantial content")
	}
}

func TestCtxAI_CategorizeContent_Sports(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "football basketball baseball sports athletes game score"
	categories := svc.categorizeContent(content, 0.3)

	found := false
	for _, cat := range categories {
		if cat.ID == "IAB17" { // Sports
			found = true
			break
		}
	}
	if !found {
		t.Error("expected IAB17 (Sports) category")
	}
}

func TestCtxAI_CategorizeContent_Technology(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "technology computing software internet mobile apps programming"
	categories := svc.categorizeContent(content, 0.3)

	found := false
	for _, cat := range categories {
		if cat.ID == "IAB19" { // Technology
			found = true
			break
		}
	}
	if !found {
		t.Error("expected IAB19 (Technology) category")
	}
}

func TestCtxAI_CategorizeContent_MultipleCategories(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "sports technology football mobile apps basketball computing"
	categories := svc.categorizeContent(content, 0.2)

	if len(categories) < 2 {
		t.Errorf("expected multiple categories, got %d", len(categories))
	}
}

func TestCtxAI_CategorizeContent_MinConfidence(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "sports"
	categoriesLow := svc.categorizeContent(content, 0.1)
	categoriesHigh := svc.categorizeContent(content, 0.9)

	if len(categoriesHigh) >= len(categoriesLow) {
		t.Error("expected higher confidence threshold to return fewer categories")
	}
}

func TestCtxAI_GetCategoryName(t *testing.T) {
	svc := NewContextualAIService(nil)

	tests := []struct {
		id       string
		expected string
	}{
		{"IAB1", "Arts & Entertainment"},
		{"IAB17", "Sports"},
		{"IAB19", "Technology & Computing"},
		{"UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		name := svc.getCategoryName(tt.id)
		if name != tt.expected {
			t.Errorf("%s: expected '%s', got '%s'", tt.id, tt.expected, name)
		}
	}
}

func TestCtxAI_AnalyzeSentiment_Positive(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "This is great! Amazing product, excellent quality, wonderful experience"
	sentiment, score := svc.analyzeSentiment(content)

	if sentiment != "positive" {
		t.Errorf("expected positive, got %s", sentiment)
	}
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}
}

func TestCtxAI_AnalyzeSentiment_Negative(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "This is terrible, horrible, awful product. Very bad experience."
	sentiment, score := svc.analyzeSentiment(content)

	if sentiment != "negative" {
		t.Errorf("expected negative, got %s", sentiment)
	}
	if score >= 0 {
		t.Errorf("expected negative score, got %f", score)
	}
}

func TestCtxAI_AnalyzeSentiment_Neutral(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "The product arrived yesterday. It was delivered on time."
	sentiment, _ := svc.analyzeSentiment(content)

	if sentiment != "neutral" {
		t.Errorf("expected neutral, got %s", sentiment)
	}
}

func TestCtxAI_AnalyzeSentiment_Empty(t *testing.T) {
	svc := NewContextualAIService(nil)

	sentiment, score := svc.analyzeSentiment("")

	if sentiment != "neutral" {
		t.Errorf("expected neutral for empty, got %s", sentiment)
	}
	if score != 0 {
		t.Errorf("expected 0 score for empty, got %f", score)
	}
}

func TestCtxAI_DetectEntities_Locations(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "Visit new york and los angeles for great experiences"
	entities := svc.detectEntities(content)

	locationCount := 0
	for _, e := range entities {
		if e.Type == "location" {
			locationCount++
		}
	}
	if locationCount < 2 {
		t.Errorf("expected at least 2 locations, got %d", locationCount)
	}
}

func TestCtxAI_DetectEntities_Organizations(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "Google and Apple are leading tech companies"
	entities := svc.detectEntities(content)

	orgCount := 0
	for _, e := range entities {
		if e.Type == "organization" {
			orgCount++
		}
	}
	if orgCount < 2 {
		t.Errorf("expected at least 2 organizations, got %d", orgCount)
	}
}

func TestCtxAI_DetectEntities_Empty(t *testing.T) {
	svc := NewContextualAIService(nil)

	entities := svc.detectEntities("no known entities here")

	if len(entities) != 0 {
		t.Errorf("expected 0 entities, got %d", len(entities))
	}
}

func TestCtxAI_DetectEmotion_Joy(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "I'm so happy and excited! This is wonderful and delightful!"
	emotion := svc.detectEmotion(content)

	if emotion != "joy" {
		t.Errorf("expected joy, got %s", emotion)
	}
}

func TestCtxAI_DetectEmotion_Fear(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "This is scary and terrifying. The dangerous risk is afraid."
	emotion := svc.detectEmotion(content)

	if emotion != "fear" {
		t.Errorf("expected fear, got %s", emotion)
	}
}

func TestCtxAI_DetectEmotion_Neutral(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "The report was submitted yesterday."
	emotion := svc.detectEmotion(content)

	if emotion != "neutral" {
		t.Errorf("expected neutral, got %s", emotion)
	}
}

func TestCtxAI_AssessContentQuality_High(t *testing.T) {
	svc := NewContextualAIService(nil)

	// Long content with diverse vocabulary
	content := "This is a comprehensive article about technology trends in modern computing. " +
		"The analysis covers various aspects including software development, hardware innovations, " +
		"and emerging technologies. Experts predict significant changes in the industry. " +
		"Multiple perspectives are considered in this detailed examination of current events."

	quality := svc.assessContentQuality(content)

	if quality < 0.6 {
		t.Errorf("expected high quality (>0.6), got %f", quality)
	}
}

func TestCtxAI_AssessContentQuality_Low(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "click here now"

	quality := svc.assessContentQuality(content)

	if quality > 0.5 {
		t.Errorf("expected low quality (<0.5), got %f", quality)
	}
}

func TestCtxAI_CheckBrandSafety_Safe(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "Great product review with positive feedback"
	isSafe := svc.checkBrandSafety(content, 0.5)

	if !isSafe {
		t.Error("expected safe content")
	}
}

func TestCtxAI_CheckBrandSafety_UnsafeKeywords(t *testing.T) {
	svc := NewContextualAIService(nil)

	unsafeContents := []string{
		"violence and attack reported",
		"illegal drugs found",
		"adult content xxx",
		"terrorist attack news",
	}

	for _, content := range unsafeContents {
		isSafe := svc.checkBrandSafety(content, 0.0)
		if isSafe {
			t.Errorf("expected unsafe for: %s", content)
		}
	}
}

func TestCtxAI_CheckBrandSafety_NegativeSentiment(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "Neutral content about weather"
	isSafe := svc.checkBrandSafety(content, -0.8) // Very negative sentiment

	if isSafe {
		t.Error("expected unsafe due to very negative sentiment")
	}
}

func TestCtxAI_EvaluateCategoryTargeting_Match(t *testing.T) {
	svc := NewContextualAIService(nil)

	categories := []model.ContextualCategory{
		{ID: "IAB17", Name: "Sports", Confidence: 0.8},
	}
	config := &model.ContextualAI{
		TargetCategories: []model.ContextualCategory{
			{ID: "IAB17", Multiplier: 1.5},
		},
	}

	multiplier := svc.evaluateCategoryTargeting(categories, config)

	if multiplier != 1.5 {
		t.Errorf("expected 1.5, got %f", multiplier)
	}
}

func TestCtxAI_EvaluateCategoryTargeting_Excluded(t *testing.T) {
	svc := NewContextualAIService(nil)

	categories := []model.ContextualCategory{
		{ID: "IAB17", Name: "Sports", Confidence: 0.8},
	}
	config := &model.ContextualAI{
		ExcludeCategories: []model.ContextualCategory{
			{ID: "IAB17"},
		},
	}

	multiplier := svc.evaluateCategoryTargeting(categories, config)

	if multiplier != 0 {
		t.Errorf("expected 0 (blocked), got %f", multiplier)
	}
}

func TestCtxAI_EvaluateCategoryTargeting_DefaultBoost(t *testing.T) {
	svc := NewContextualAIService(nil)

	categories := []model.ContextualCategory{
		{ID: "IAB17", Name: "Sports", Confidence: 0.8},
	}
	config := &model.ContextualAI{
		TargetCategories: []model.ContextualCategory{
			{ID: "IAB17", Multiplier: 0}, // No multiplier set
		},
	}

	multiplier := svc.evaluateCategoryTargeting(categories, config)

	if multiplier != 1.2 {
		t.Errorf("expected default 1.2, got %f", multiplier)
	}
}

func TestCtxAI_EvaluateSentimentTargeting_Positive(t *testing.T) {
	svc := NewContextualAIService(nil)

	config := &model.SentimentTargeting{
		TargetPositive: true,
		PositiveBoost:  1.3,
	}

	multiplier := svc.evaluateSentimentTargeting("positive", 0.7, config)

	if multiplier != 1.3 {
		t.Errorf("expected 1.3, got %f", multiplier)
	}
}

func TestCtxAI_EvaluateSentimentTargeting_Negative(t *testing.T) {
	svc := NewContextualAIService(nil)

	config := &model.SentimentTargeting{
		TargetNegative:  true,
		NegativePenalty: 0.6,
	}

	multiplier := svc.evaluateSentimentTargeting("negative", -0.5, config)

	if multiplier != 0.6 {
		t.Errorf("expected 0.6, got %f", multiplier)
	}
}

func TestCtxAI_EvaluateSentimentTargeting_NegativeNotTargeted(t *testing.T) {
	svc := NewContextualAIService(nil)

	config := &model.SentimentTargeting{
		TargetPositive: true,
		TargetNegative: false,
	}

	multiplier := svc.evaluateSentimentTargeting("negative", -0.5, config)

	if multiplier != 0.3 {
		t.Errorf("expected 0.3 (heavy penalty), got %f", multiplier)
	}
}

func TestCtxAI_EvaluateEntityTargeting_Match(t *testing.T) {
	svc := NewContextualAIService(nil)

	entities := []model.DetectedEntity{
		{Type: "organization", Name: "Google", Confidence: 0.9},
	}
	targets := []model.EntityTarget{
		{EntityType: "organization", Entities: []string{"Google"}, Multiplier: 1.5},
	}

	multiplier := svc.evaluateEntityTargeting(entities, targets)

	if multiplier != 1.5 {
		t.Errorf("expected 1.5, got %f", multiplier)
	}
}

func TestCtxAI_EvaluateEntityTargeting_Excluded(t *testing.T) {
	svc := NewContextualAIService(nil)

	entities := []model.DetectedEntity{
		{Type: "organization", Name: "BadCompany", Confidence: 0.9},
	}
	targets := []model.EntityTarget{
		{EntityType: "organization", Entities: []string{"BadCompany"}, Exclude: true},
	}

	multiplier := svc.evaluateEntityTargeting(entities, targets)

	if multiplier != 0 {
		t.Errorf("expected 0 (blocked), got %f", multiplier)
	}
}

func TestCtxAI_CalculateSemanticSimilarity(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "sports football basketball game score"
	seeds := []string{"sports game football", "basketball score"}

	similarity := svc.calculateSemanticSimilarity(content, seeds)

	if similarity < 0.5 {
		t.Errorf("expected high similarity, got %f", similarity)
	}
}

func TestCtxAI_CalculateSemanticSimilarity_NoMatch(t *testing.T) {
	svc := NewContextualAIService(nil)

	content := "technology software programming"
	seeds := []string{"cooking recipes food"}

	similarity := svc.calculateSemanticSimilarity(content, seeds)

	if similarity > 0.1 {
		t.Errorf("expected low similarity, got %f", similarity)
	}
}

func TestCtxAI_CalculateSemanticSimilarity_Empty(t *testing.T) {
	svc := NewContextualAIService(nil)

	similarity := svc.calculateSemanticSimilarity("content", []string{})

	if similarity != 0 {
		t.Errorf("expected 0 for empty seeds, got %f", similarity)
	}
}

func TestCtxAI_BrandSafetyBlock(t *testing.T) {
	svc := NewContextualAIService(nil)
	campaign := createCtxCampaign(true)
	req := createCtxRequest("violence and terrorist attack news")

	result := svc.AnalyzeContext(campaign, req)

	if result.BrandSafe {
		t.Error("expected brand unsafe")
	}
	if result.BidMultiplier != 0 {
		t.Errorf("expected multiplier 0 for unsafe, got %f", result.BidMultiplier)
	}
	if result.Reason != "brand_safety_block" {
		t.Errorf("expected 'brand_safety_block', got '%s'", result.Reason)
	}
}

func TestCtxAI_SemanticTargeting(t *testing.T) {
	svc := NewContextualAIService(nil)
	campaign := createCtxCampaign(true)
	campaign.Targeting.ContextualAI.SemanticTargeting = &model.SemanticTargeting{
		Enabled:             true,
		SeedContent:         []string{"sports football basketball"},
		SimilarityThreshold: 0.3,
	}
	req := createCtxRequest("football basketball sports game score")

	result := svc.AnalyzeContext(campaign, req)

	if result.SemanticMatch < 0.3 {
		t.Errorf("expected semantic match >= 0.3, got %f", result.SemanticMatch)
	}
}

func TestCtxAI_MultiplierCapped(t *testing.T) {
	svc := NewContextualAIService(nil)
	campaign := createCtxCampaign(true)
	campaign.Targeting.ContextualAI.TargetCategories = []model.ContextualCategory{
		{ID: "IAB17", Multiplier: 5.0}, // Very high multiplier
	}
	campaign.Targeting.ContextualAI.SentimentTargeting = &model.SentimentTargeting{
		TargetPositive: true,
		PositiveBoost:  5.0,
	}

	req := createCtxRequest("great amazing excellent sports football basketball wonderful")

	result := svc.AnalyzeContext(campaign, req)

	if result.BidMultiplier > 2.5 {
		t.Errorf("expected multiplier capped at 2.5, got %f", result.BidMultiplier)
	}
}
