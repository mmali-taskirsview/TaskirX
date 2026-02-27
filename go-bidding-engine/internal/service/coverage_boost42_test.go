package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// Coverage Boost 42: Target remaining 90-94% functions with simple edge cases
// Functions targeted:
// 1. calculateLanguageMultiplier (93.8%) - exact locale matching, content language paths
// 2. calculateDayOfWeekMultiplier (93.0%) - timezone edge cases
// 3. matchLanguageCode (helper) - underscore format handling

// ===============================================================================
// calculateLanguageMultiplier Tests
// ===============================================================================

// TestB42_Language_ExactLocaleMatch tests locale-specific matching
func TestB42_Language_ExactLocaleMatch(t *testing.T) {
	mc := cache.NewMockCache()
	svc := &BiddingService{cache: mc}

	camp := &model.Campaign{
		ID: "camp-1",
		Targeting: model.Targeting{
			LanguageTargeting: &model.LanguageTargeting{
				LocaleMatching: true, // Enable locale matching
				Languages: []model.LanguageRule{
					{
						Code:   "en",
						Locale: "en-US", // Exact locale required
						Boost:  1.5,
					},
				},
			},
		},
	}

	req := &model.BidRequest{
		User: model.InternalUser{Language: "en-US"}, // Exact match
	}

	result := svc.calculateLanguageMultiplier(camp, req)

	if !result.Matched {
		t.Errorf("Expected match for exact locale en-US")
	}
	if result.Multiplier != 1.5 {
		t.Errorf("Expected multiplier 1.5, got %f", result.Multiplier)
	}
}

// TestB42_Language_LocaleNoMatch tests exact locale mismatch
func TestB42_Language_LocaleNoMatch(t *testing.T) {
	mc := cache.NewMockCache()
	svc := &BiddingService{cache: mc}

	camp := &model.Campaign{
		ID: "camp-2",
		Targeting: model.Targeting{
			LanguageTargeting: &model.LanguageTargeting{
				LocaleMatching: true,
				Languages: []model.LanguageRule{
					{
						Code:   "en",
						Locale: "en-US",
						Boost:  1.5,
					},
				},
			},
		},
	}

	req := &model.BidRequest{
		User: model.InternalUser{Language: "en-GB"}, // Different locale
	}

	result := svc.calculateLanguageMultiplier(camp, req)

	if result.Matched {
		t.Errorf("Expected no match for en-GB when en-US required")
	}
}

// TestB42_Language_ContentLanguageMatch tests content language matching
func TestB42_Language_ContentLanguageMatch(t *testing.T) {
	mc := cache.NewMockCache()
	svc := &BiddingService{cache: mc}

	camp := &model.Campaign{
		ID: "camp-3",
		Targeting: model.Targeting{
			LanguageTargeting: &model.LanguageTargeting{
				ContentLanguage: true,  // Enable content language
				PrimaryOnly:     false, // Check both user and content
				Languages: []model.LanguageRule{
					{
						Code:  "es",
						Boost: 1.3,
					},
				},
			},
		},
	}

	req := &model.BidRequest{
		User: model.InternalUser{Language: "en"}, // User language doesn't match
		Context: map[string]interface{}{
			"content_language": "es", // Content language matches
		},
	}

	result := svc.calculateLanguageMultiplier(camp, req)

	if !result.Matched {
		t.Errorf("Expected match via content language")
	}
	if result.Multiplier != 1.3 {
		t.Errorf("Expected multiplier 1.3, got %f", result.Multiplier)
	}
}

// TestB42_Language_ContentExcluded tests content language exclusion
func TestB42_Language_ContentExcluded(t *testing.T) {
	mc := cache.NewMockCache()
	svc := &BiddingService{cache: mc}

	camp := &model.Campaign{
		ID: "camp-4",
		Targeting: model.Targeting{
			LanguageTargeting: &model.LanguageTargeting{
				ContentLanguage:  true,
				ExcludeLanguages: []string{"fr"},
			},
		},
	}

	req := &model.BidRequest{
		User: model.InternalUser{Language: "en"},
		Context: map[string]interface{}{
			"content_language": "fr", // Excluded content language
		},
	}

	result := svc.calculateLanguageMultiplier(camp, req)

	if !result.Blocked {
		t.Errorf("Expected block for excluded content language")
	}
	if result.Reason != "content_language_excluded:fr" {
		t.Errorf("Expected reason 'content_language_excluded:fr', got %s", result.Reason)
	}
}

// TestB42_Language_DefaultMultiplier tests default multiplier when no match
func TestB42_Language_DefaultMultiplier(t *testing.T) {
	mc := cache.NewMockCache()
	svc := &BiddingService{cache: mc}

	camp := &model.Campaign{
		ID: "camp-5",
		Targeting: model.Targeting{
			LanguageTargeting: &model.LanguageTargeting{
				Languages: []model.LanguageRule{
					{
						Code:  "es",
						Boost: 1.5,
					},
				},
				DefaultMultiplier: 0.8, // Penalty for non-matching
			},
		},
	}

	req := &model.BidRequest{
		User: model.InternalUser{Language: "en"}, // Doesn't match es
	}

	result := svc.calculateLanguageMultiplier(camp, req)

	if result.Matched {
		t.Errorf("Expected no match")
	}
	if result.Multiplier != 0.8 {
		t.Errorf("Expected default multiplier 0.8, got %f", result.Multiplier)
	}
}

// TestB42_Language_UnderscoreFormat tests en_US format handling
func TestB42_Language_UnderscoreFormat(t *testing.T) {
	mc := cache.NewMockCache()
	svc := &BiddingService{cache: mc}

	camp := &model.Campaign{
		ID: "camp-6",
		Targeting: model.Targeting{
			LanguageTargeting: &model.LanguageTargeting{
				Languages: []model.LanguageRule{
					{
						Code:  "en",
						Boost: 1.2,
					},
				},
			},
		},
	}

	req := &model.BidRequest{
		User: model.InternalUser{Language: "en_US"}, // Underscore format
	}

	result := svc.calculateLanguageMultiplier(camp, req)

	if !result.Matched {
		t.Errorf("Expected match for en_US to en")
	}
}

// TestB42_Language_MultiplierCap tests 2.5 cap
func TestB42_Language_MultiplierCap(t *testing.T) {
	mc := cache.NewMockCache()
	svc := &BiddingService{cache: mc}

	camp := &model.Campaign{
		ID: "camp-7",
		Targeting: model.Targeting{
			LanguageTargeting: &model.LanguageTargeting{
				Languages: []model.LanguageRule{
					{
						Code:  "en",
						Boost: 3.0, // Very high boost
					},
				},
			},
		},
	}

	req := &model.BidRequest{
		User: model.InternalUser{Language: "en"},
	}

	result := svc.calculateLanguageMultiplier(camp, req)

	if result.Multiplier != 2.5 {
		t.Errorf("Expected cap at 2.5, got %f", result.Multiplier)
	}
}

// ===============================================================================
// extractLanguageInfo Tests
// ===============================================================================

// TestB42_ExtractLanguage_PageLanguageFallback tests page_language fallback
func TestB42_ExtractLanguage_PageLanguageFallback(t *testing.T) {
	svc := &BiddingService{cache: cache.NewMockCache()}

	req := &model.BidRequest{
		User: model.InternalUser{Language: ""}, // No user language
		Context: map[string]interface{}{
			"page_language": "de", // Fallback to page language
			// No content_language set
		},
	}

	userLang, contentLang := svc.extractLanguageInfo(req)

	if userLang != "" {
		t.Errorf("Expected empty user language, got %s", userLang)
	}
	if contentLang != "de" {
		t.Errorf("Expected content language 'de' from page_language, got %s", contentLang)
	}
}

// TestB42_ExtractLanguage_ContextLanguageFallback tests context language fallback
func TestB42_ExtractLanguage_ContextLanguageFallback(t *testing.T) {
	svc := &BiddingService{cache: cache.NewMockCache()}

	req := &model.BidRequest{
		User: model.InternalUser{Language: ""}, // No user language
		Context: map[string]interface{}{
			"language": "fr", // Fallback to context language
		},
	}

	userLang, _ := svc.extractLanguageInfo(req)

	if userLang != "fr" {
		t.Errorf("Expected user language 'fr' from context, got %s", userLang)
	}
}

// ===============================================================================
// matchLanguageCode Tests
// ===============================================================================

// TestB42_MatchLanguage_EmptyStrings tests empty string handling
func TestB42_MatchLanguage_EmptyStrings(t *testing.T) {
	svc := &BiddingService{cache: cache.NewMockCache()}

	// Empty user language
	match := svc.matchLanguageCode("", "en", false)
	if match {
		t.Errorf("Expected no match for empty user language")
	}

	// Empty target language
	match = svc.matchLanguageCode("en", "", false)
	if match {
		t.Errorf("Expected no match for empty target language")
	}
}

// TestB42_MatchLanguage_ExactLocaleRequired tests exact locale requirement
func TestB42_MatchLanguage_ExactLocaleRequired(t *testing.T) {
	svc := &BiddingService{cache: cache.NewMockCache()}

	// Should NOT match when exact locale required and they differ
	match := svc.matchLanguageCode("en-US", "en-GB", true)
	if match {
		t.Errorf("Expected no match for en-US vs en-GB with exact locale")
	}

	// Should match when exact
	match = svc.matchLanguageCode("en-US", "en-US", true)
	if !match {
		t.Errorf("Expected match for exact en-US")
	}
}

// TestB42_MatchLanguage_UnderscoreFormat tests underscore format
func TestB42_MatchLanguage_UnderscoreFormat(t *testing.T) {
	svc := &BiddingService{cache: cache.NewMockCache()}

	// en_US should match en
	match := svc.matchLanguageCode("en_US", "en", false)
	if !match {
		t.Errorf("Expected match for en_US to en")
	}

	// en should match en_GB
	match = svc.matchLanguageCode("en", "en_GB", false)
	if !match {
		t.Errorf("Expected match for en to en_GB")
	}
}
