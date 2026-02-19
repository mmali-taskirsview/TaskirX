package service

import (
	"math"
	"strings"
	"sync"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ContextualAIService provides ML-based contextual analysis
type ContextualAIService struct {
	cache            cache.Cache
	mu               sync.RWMutex
	categoryCache    map[string]*categoryResult
	sentimentLexicon map[string]float64
	entityPatterns   map[string][]string
}

type categoryResult struct {
	categories []model.ContextualCategory
	sentiment  string
	score      float64
	entities   []model.DetectedEntity
}

// Simplified sentiment lexicon
var defaultSentimentLexicon = map[string]float64{
	// Positive words
	"great": 0.8, "excellent": 0.9, "amazing": 0.9, "wonderful": 0.8,
	"fantastic": 0.9, "love": 0.7, "happy": 0.7, "best": 0.8,
	"perfect": 0.9, "beautiful": 0.7, "awesome": 0.8, "good": 0.5,
	"success": 0.7, "win": 0.6, "positive": 0.6, "joy": 0.8,
	// Negative words
	"bad": -0.6, "terrible": -0.9, "awful": -0.8, "horrible": -0.9,
	"hate": -0.8, "sad": -0.6, "worst": -0.9, "poor": -0.5,
	"fail": -0.7, "negative": -0.5, "angry": -0.7, "disappointing": -0.7,
	"disaster": -0.9, "crisis": -0.7, "problem": -0.4, "wrong": -0.5,
	// Brand safety concerns
	"violence": -1.0, "death": -0.8, "kill": -0.9, "attack": -0.7,
	"drugs": -0.8, "illegal": -0.7, "scandal": -0.6, "controversy": -0.5,
}

// IAB category keywords (simplified mapping)
var categoryKeywords = map[string][]string{
	"IAB1":  {"arts", "entertainment", "movies", "music", "television", "celebrity"},
	"IAB2":  {"automotive", "cars", "vehicles", "trucks", "motorcycles"},
	"IAB3":  {"business", "marketing", "advertising", "industry", "economy"},
	"IAB4":  {"careers", "jobs", "employment", "resume", "interview"},
	"IAB5":  {"education", "college", "university", "learning", "school"},
	"IAB6":  {"family", "parenting", "kids", "children", "babies"},
	"IAB7":  {"health", "fitness", "medical", "wellness", "nutrition"},
	"IAB8":  {"food", "drink", "cooking", "recipes", "restaurant"},
	"IAB9":  {"hobbies", "games", "crafts", "collecting"},
	"IAB10": {"home", "garden", "interior", "furniture", "appliances"},
	"IAB11": {"law", "government", "politics", "legal"},
	"IAB12": {"news", "weather", "international", "national", "local"},
	"IAB13": {"finance", "investing", "banking", "insurance", "credit"},
	"IAB14": {"society", "dating", "weddings", "divorce"},
	"IAB15": {"science", "biology", "chemistry", "physics", "space"},
	"IAB16": {"pets", "dogs", "cats", "birds", "fish"},
	"IAB17": {"sports", "football", "basketball", "baseball", "soccer"},
	"IAB18": {"style", "fashion", "beauty", "clothing"},
	"IAB19": {"technology", "computing", "internet", "mobile", "software"},
	"IAB20": {"travel", "hotels", "flights", "vacation", "tourism"},
	"IAB21": {"real estate", "apartments", "homes", "property"},
	"IAB22": {"shopping", "deals", "coupons", "comparison"},
	"IAB23": {"religion", "spirituality", "faith"},
}

// NewContextualAIService creates a new contextual AI service
func NewContextualAIService(c cache.Cache) *ContextualAIService {
	return &ContextualAIService{
		cache:            c,
		categoryCache:    make(map[string]*categoryResult),
		sentimentLexicon: defaultSentimentLexicon,
		entityPatterns:   make(map[string][]string),
	}
}

// AnalyzeContext performs contextual analysis on page content
func (s *ContextualAIService) AnalyzeContext(campaign *model.Campaign, req *model.BidRequest) *model.ContextualAIResult {
	config := campaign.Targeting.ContextualAI
	if config == nil || !config.Enabled {
		return &model.ContextualAIResult{
			Analyzed:      false,
			BidMultiplier: 1.0,
			BrandSafe:     true,
			Reason:        "contextual_ai_disabled",
		}
	}

	// Get page content from request
	pageContent := s.extractPageContent(req)
	if pageContent == "" {
		return &model.ContextualAIResult{
			Analyzed:      false,
			BidMultiplier: 1.0,
			BrandSafe:     true,
			Reason:        "no_page_content",
		}
	}

	result := &model.ContextualAIResult{
		Analyzed:      true,
		BidMultiplier: 1.0,
		BrandSafe:     true,
		Confidence:    0.8,
	}

	// Analyze content categories
	if config.AnalyzeContent {
		categories := s.categorizeContent(pageContent, config.MinConfidence)
		result.Categories = categories

		// Check category targeting
		result.BidMultiplier = s.evaluateCategoryTargeting(categories, config)
	}

	// Analyze sentiment
	if config.AnalyzeSentiment {
		sentiment, score := s.analyzeSentiment(pageContent)
		result.Sentiment = sentiment
		result.SentimentScore = score

		// Apply sentiment targeting
		if config.SentimentTargeting != nil {
			multiplier := s.evaluateSentimentTargeting(sentiment, score, config.SentimentTargeting)
			result.BidMultiplier *= multiplier
		}
	}

	// Detect entities
	if config.AnalyzeEntities {
		entities := s.detectEntities(pageContent)
		result.Entities = entities

		// Apply entity targeting
		if len(config.EntityTargeting) > 0 {
			multiplier := s.evaluateEntityTargeting(entities, config.EntityTargeting)
			result.BidMultiplier *= multiplier
		}
	}

	// Analyze emotional tone
	if config.AnalyzeEmotion {
		result.Emotion = s.detectEmotion(pageContent)
	}

	// Calculate content quality score
	result.ContentQuality = s.assessContentQuality(pageContent)

	// Check brand safety
	result.BrandSafe = s.checkBrandSafety(pageContent, result.SentimentScore)
	if !result.BrandSafe {
		result.BidMultiplier = 0 // Block unsafe content
		result.Reason = "brand_safety_block"
		return result
	}

	// Semantic matching if configured
	if config.SemanticTargeting != nil && config.SemanticTargeting.Enabled {
		semanticScore := s.calculateSemanticSimilarity(pageContent, config.SemanticTargeting.SeedContent)
		result.SemanticMatch = semanticScore

		if semanticScore >= config.SemanticTargeting.SimilarityThreshold {
			result.BidMultiplier *= 1.3 // Boost for semantic match
		}
	}

	// Cap multiplier
	if result.BidMultiplier > 2.5 {
		result.BidMultiplier = 2.5
	}
	if result.BidMultiplier < 0.2 && result.BrandSafe {
		result.BidMultiplier = 0.2
	}

	result.Reason = "analysis_complete"
	return result
}

func (s *ContextualAIService) extractPageContent(req *model.BidRequest) string {
	var content strings.Builder

	// Get content from PublisherID and Context
	if req.PublisherID != "" {
		content.WriteString(req.PublisherID + " ")
	}

	if req.Context != nil {
		if pageTitle, ok := req.Context["page_title"].(string); ok {
			content.WriteString(pageTitle + " ")
		}
		if pageContent, ok := req.Context["page_content"].(string); ok {
			content.WriteString(pageContent + " ")
		}
		if pageKeywords, ok := req.Context["page_keywords"].(string); ok {
			content.WriteString(pageKeywords + " ")
		}
		if pageDescription, ok := req.Context["page_description"].(string); ok {
			content.WriteString(pageDescription + " ")
		}
		if categories, ok := req.Context["page_categories"].([]interface{}); ok {
			for _, cat := range categories {
				if catStr, ok := cat.(string); ok {
					content.WriteString(catStr + " ")
				}
			}
		}
	}

	return strings.TrimSpace(content.String())
}

func (s *ContextualAIService) categorizeContent(content string, minConfidence float64) []model.ContextualCategory {
	if minConfidence <= 0 {
		minConfidence = 0.3
	}

	content = strings.ToLower(content)
	words := strings.Fields(content)

	categoryScores := make(map[string]float64)

	// Count keyword matches for each category
	for catID, keywords := range categoryKeywords {
		score := 0.0
		for _, word := range words {
			for _, keyword := range keywords {
				if strings.Contains(word, keyword) || strings.Contains(keyword, word) {
					score += 1.0
				}
			}
		}
		if score > 0 {
			// Normalize score
			categoryScores[catID] = math.Min(score/5.0, 1.0)
		}
	}

	// Convert to result slice
	var categories []model.ContextualCategory
	for catID, score := range categoryScores {
		if score >= minConfidence {
			categories = append(categories, model.ContextualCategory{
				ID:         catID,
				Name:       s.getCategoryName(catID),
				Taxonomy:   "iab",
				Confidence: score,
			})
		}
	}

	return categories
}

func (s *ContextualAIService) getCategoryName(catID string) string {
	names := map[string]string{
		"IAB1": "Arts & Entertainment", "IAB2": "Automotive", "IAB3": "Business",
		"IAB4": "Careers", "IAB5": "Education", "IAB6": "Family & Parenting",
		"IAB7": "Health & Fitness", "IAB8": "Food & Drink", "IAB9": "Hobbies & Interests",
		"IAB10": "Home & Garden", "IAB11": "Law, Gov't & Politics", "IAB12": "News",
		"IAB13": "Personal Finance", "IAB14": "Society", "IAB15": "Science",
		"IAB16": "Pets", "IAB17": "Sports", "IAB18": "Style & Fashion",
		"IAB19": "Technology & Computing", "IAB20": "Travel", "IAB21": "Real Estate",
		"IAB22": "Shopping", "IAB23": "Religion & Spirituality",
	}
	if name, exists := names[catID]; exists {
		return name
	}
	return catID
}

func (s *ContextualAIService) analyzeSentiment(content string) (string, float64) {
	content = strings.ToLower(content)
	words := strings.Fields(content)

	var totalScore float64
	var matchCount int

	for _, word := range words {
		// Clean punctuation
		word = strings.Trim(word, ".,!?\"'()[]{}:;")

		if score, exists := s.sentimentLexicon[word]; exists {
			totalScore += score
			matchCount++
		}
	}

	if matchCount == 0 {
		return "neutral", 0.0
	}

	avgScore := totalScore / float64(matchCount)

	// Determine sentiment label
	var sentiment string
	if avgScore > 0.2 {
		sentiment = "positive"
	} else if avgScore < -0.2 {
		sentiment = "negative"
	} else {
		sentiment = "neutral"
	}

	return sentiment, avgScore
}

func (s *ContextualAIService) detectEntities(content string) []model.DetectedEntity {
	// Simplified entity detection - in production would use NER model
	var entities []model.DetectedEntity

	content = strings.ToLower(content)

	// Location patterns
	locations := []string{"new york", "los angeles", "london", "paris", "tokyo", "chicago", "san francisco"}
	for _, loc := range locations {
		if strings.Contains(content, loc) {
			entities = append(entities, model.DetectedEntity{
				Type:       "location",
				Name:       strings.Title(loc),
				Confidence: 0.8,
			})
		}
	}

	// Organization patterns (simplified)
	orgs := []string{"google", "apple", "microsoft", "amazon", "facebook", "meta", "netflix", "tesla"}
	for _, org := range orgs {
		if strings.Contains(content, org) {
			entities = append(entities, model.DetectedEntity{
				Type:       "organization",
				Name:       strings.Title(org),
				Confidence: 0.9,
			})
		}
	}

	return entities
}

func (s *ContextualAIService) detectEmotion(content string) string {
	content = strings.ToLower(content)

	emotionKeywords := map[string][]string{
		"joy":          {"happy", "excited", "thrilled", "delighted", "wonderful"},
		"trust":        {"reliable", "honest", "trustworthy", "safe", "secure"},
		"fear":         {"scary", "afraid", "terrifying", "dangerous", "risk"},
		"surprise":     {"unexpected", "shocking", "amazing", "incredible"},
		"sadness":      {"sad", "depressing", "tragic", "heartbreaking"},
		"anger":        {"angry", "furious", "outrage", "frustrating"},
		"anticipation": {"upcoming", "awaiting", "expecting", "soon"},
	}

	emotionScores := make(map[string]int)
	words := strings.Fields(content)

	for _, word := range words {
		for emotion, keywords := range emotionKeywords {
			for _, kw := range keywords {
				if strings.Contains(word, kw) {
					emotionScores[emotion]++
				}
			}
		}
	}

	// Find dominant emotion
	maxScore := 0
	dominantEmotion := "neutral"
	for emotion, score := range emotionScores {
		if score > maxScore {
			maxScore = score
			dominantEmotion = emotion
		}
	}

	return dominantEmotion
}

func (s *ContextualAIService) assessContentQuality(content string) float64 {
	// Simple quality heuristics
	quality := 0.5

	wordCount := len(strings.Fields(content))

	// Length factor
	if wordCount > 100 {
		quality += 0.2
	} else if wordCount < 20 {
		quality -= 0.2
	}

	// Diversity factor (unique words ratio)
	words := strings.Fields(strings.ToLower(content))
	uniqueWords := make(map[string]bool)
	for _, w := range words {
		uniqueWords[w] = true
	}
	if len(words) > 0 {
		diversity := float64(len(uniqueWords)) / float64(len(words))
		quality += diversity * 0.2
	}

	return math.Max(0, math.Min(1, quality))
}

func (s *ContextualAIService) checkBrandSafety(content string, sentimentScore float64) bool {
	content = strings.ToLower(content)

	// Check for unsafe keywords
	unsafeKeywords := []string{
		"violence", "death", "kill", "murder", "attack", "terrorist",
		"drugs", "illegal", "porn", "adult", "xxx", "gambling",
		"hate", "racist", "discrimination", "extremist",
	}

	for _, keyword := range unsafeKeywords {
		if strings.Contains(content, keyword) {
			return false
		}
	}

	// Very negative sentiment is also a flag
	if sentimentScore < -0.7 {
		return false
	}

	return true
}

func (s *ContextualAIService) evaluateCategoryTargeting(categories []model.ContextualCategory, config *model.ContextualAI) float64 {
	multiplier := 1.0

	for _, detected := range categories {
		// Check exclusions
		for _, excluded := range config.ExcludeCategories {
			if detected.ID == excluded.ID {
				return 0 // Block
			}
		}

		// Check target categories
		for _, target := range config.TargetCategories {
			if detected.ID == target.ID {
				if target.Multiplier > 0 {
					multiplier *= target.Multiplier
				} else {
					multiplier *= 1.2 // Default boost
				}
			}
		}
	}

	return multiplier
}

func (s *ContextualAIService) evaluateSentimentTargeting(sentiment string, score float64, config *model.SentimentTargeting) float64 {
	multiplier := 1.0

	switch sentiment {
	case "positive":
		if config.TargetPositive {
			multiplier *= config.PositiveBoost
			if multiplier == 0 {
				multiplier = 1.2
			}
		}
	case "negative":
		if config.TargetNegative {
			penalty := config.NegativePenalty
			if penalty == 0 {
				penalty = 0.5
			}
			multiplier *= penalty
		} else {
			return 0.3 // Heavy penalty if not targeting negative
		}
	case "neutral":
		if config.TargetNeutral {
			multiplier *= 1.0 // No change
		}
	}

	// Check minimum score threshold
	if config.MinSentimentScore != 0 && score < config.MinSentimentScore {
		multiplier *= 0.5
	}

	return multiplier
}

func (s *ContextualAIService) evaluateEntityTargeting(entities []model.DetectedEntity, targets []model.EntityTarget) float64 {
	multiplier := 1.0

	for _, target := range targets {
		for _, entity := range entities {
			if entity.Type == target.EntityType {
				for _, targetName := range target.Entities {
					if strings.EqualFold(entity.Name, targetName) {
						if target.Exclude {
							return 0 // Block
						}
						if target.Multiplier > 0 {
							multiplier *= target.Multiplier
						} else {
							multiplier *= 1.2
						}
					}
				}
			}
		}
	}

	return multiplier
}

func (s *ContextualAIService) calculateSemanticSimilarity(content string, seedContent []string) float64 {
	if len(seedContent) == 0 {
		return 0
	}

	// Simplified semantic similarity using word overlap
	contentWords := make(map[string]bool)
	for _, word := range strings.Fields(strings.ToLower(content)) {
		contentWords[word] = true
	}

	var totalOverlap float64
	for _, seed := range seedContent {
		seedWords := strings.Fields(strings.ToLower(seed))
		overlap := 0
		for _, word := range seedWords {
			if contentWords[word] {
				overlap++
			}
		}
		if len(seedWords) > 0 {
			totalOverlap += float64(overlap) / float64(len(seedWords))
		}
	}

	return totalOverlap / float64(len(seedContent))
}
