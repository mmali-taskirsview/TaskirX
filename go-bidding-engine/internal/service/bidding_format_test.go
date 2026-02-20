package service

import (
	"strings"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ==================== Ad Format Generation Tests ====================

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Empty string",
			input: "",
			want:  "taskirx-ad.com",
		},
		{
			name:  "Any URL",
			input: "https://example.com/path",
			want:  "taskirx-ad.com",
		},
		{
			name:  "Random string",
			input: "random-input",
			want:  "taskirx-ad.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDomain(tt.input)
			if got != tt.want {
				t.Errorf("extractDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateAudioVAST(t *testing.T) {
	campaign := &model.Campaign{
		ID:   "audio-123",
		Name: "Test Audio Campaign",
		Creative: model.Creative{
			Duration: 30,
			MimeType: "audio/mpeg",
			Bitrate:  128,
			URL:      "https://cdn.example.com/audio.mp3",
		},
	}
	impURL := "https://track.example.com/impression"
	clickURL := "https://click.example.com/click"

	result := generateAudioVAST(campaign, impURL, clickURL)

	// Verify VAST structure
	if !strings.Contains(result, `<VAST version="4.0">`) {
		t.Error("Missing VAST version declaration")
	}
	if !strings.Contains(result, campaign.ID) {
		t.Error("Missing campaign ID in VAST")
	}
	if !strings.Contains(result, campaign.Name) {
		t.Error("Missing campaign name in VAST")
	}
	if !strings.Contains(result, impURL) {
		t.Error("Missing impression URL in VAST")
	}
	if !strings.Contains(result, clickURL) {
		t.Error("Missing click URL in VAST")
	}
	if !strings.Contains(result, campaign.Creative.MimeType) {
		t.Error("Missing mime type in VAST")
	}
	if !strings.Contains(result, campaign.Creative.URL) {
		t.Error("Missing media URL in VAST")
	}
	if !strings.Contains(result, `<Duration>00:00:30</Duration>`) {
		t.Error("Missing or incorrect duration in VAST")
	}
}

func TestGenerateRichMedia(t *testing.T) {
	tests := []struct {
		name           string
		campaign       *model.Campaign
		impURL         string
		clickURL       string
		wantContains   []string
		wantNotContain []string
	}{
		{
			name: "With HTML snippet",
			campaign: &model.Campaign{
				ID: "rich-1",
				Creative: model.Creative{
					Width:       300,
					Height:      250,
					HTMLSnippet: "<div>Custom Ad Content</div>",
				},
			},
			impURL:   "https://imp.example.com",
			clickURL: "https://click.example.com",
			wantContains: []string{
				"Custom Ad Content",
				"300px",
				"250px",
				"rich-1",
			},
		},
		{
			name: "Without HTML snippet (fallback iframe)",
			campaign: &model.Campaign{
				ID: "rich-2",
				Creative: model.Creative{
					Width:  728,
					Height: 90,
					URL:    "https://cdn.example.com/ad.html",
				},
			},
			impURL:   "https://imp.example.com",
			clickURL: "https://click.example.com",
			wantContains: []string{
				"<iframe",
				"https://cdn.example.com/ad.html",
				"728",
				"90",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateRichMedia(tt.campaign, tt.impURL, tt.clickURL)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("generateRichMedia() missing expected content: %s", want)
				}
			}
		})
	}
}

func TestGeneratePlayable(t *testing.T) {
	campaign := &model.Campaign{
		ID: "playable-1",
		Creative: model.Creative{
			URL: "https://cdn.example.com/playable.html",
		},
	}
	impURL := "https://imp.example.com"
	clickURL := "https://click.example.com"

	result := generatePlayable(campaign, impURL, clickURL)

	// Verify MRAID structure
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("Missing DOCTYPE")
	}
	if !strings.Contains(result, "mraid.js") {
		t.Error("Missing MRAID script")
	}
	if !strings.Contains(result, campaign.Creative.URL) {
		t.Error("Missing creative URL")
	}
	if !strings.Contains(result, clickURL) {
		t.Error("Missing click URL")
	}
	if !strings.Contains(result, impURL) {
		t.Error("Missing impression URL")
	}
	if !strings.Contains(result, "useCustomClose") {
		t.Error("Missing MRAID custom close handling")
	}
}

func TestGeneratePop(t *testing.T) {
	campaign := &model.Campaign{ID: "pop-1"}
	impURL := "https://imp.example.com"
	clickURL := "https://click.example.com"

	result := generatePop(campaign, impURL, clickURL)

	// Verify pop script structure
	if !strings.Contains(result, "<script>") {
		t.Error("Missing script tag")
	}
	if !strings.Contains(result, clickURL) {
		t.Error("Missing click URL")
	}
	if !strings.Contains(result, impURL) {
		t.Error("Missing impression URL")
	}
	if !strings.Contains(result, "window.open") {
		t.Error("Missing popup window.open call")
	}
	if !strings.Contains(result, "blur") {
		t.Error("Missing popunder blur")
	}
	if !strings.Contains(result, "addEventListener") {
		t.Error("Missing event listener")
	}
}

func TestGeneratePush(t *testing.T) {
	campaign := &model.Campaign{
		Creative: model.Creative{
			Title:       "Push Notification Title",
			Description: "Click here for great deals!",
			IconURL:     "https://cdn.example.com/icon.png",
			URL:         "https://cdn.example.com/image.png",
		},
	}
	impURL := "https://imp.example.com"
	clickURL := "https://click.example.com"

	result := generatePush(campaign, impURL, clickURL)

	// Should be valid JSON
	if !strings.HasPrefix(result, "{") || !strings.HasSuffix(result, "}") {
		t.Error("Result should be JSON object")
	}
	if !strings.Contains(result, "Push Notification Title") {
		t.Error("Missing title in push JSON")
	}
	if !strings.Contains(result, campaign.Creative.Description) {
		t.Error("Missing body in push JSON")
	}
	if !strings.Contains(result, campaign.Creative.IconURL) {
		t.Error("Missing icon in push JSON")
	}
	if !strings.Contains(result, clickURL) {
		t.Error("Missing click URL in push JSON")
	}
	if !strings.Contains(result, impURL) {
		t.Error("Missing impression URL in push JSON")
	}
}

// ==================== Language Tests ====================

func TestExtractLanguageInfo(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name            string
		req             *model.BidRequest
		wantUserLang    string
		wantContentLang string
	}{
		{
			name: "User language from User struct",
			req: &model.BidRequest{
				User: model.InternalUser{Language: "en"},
			},
			wantUserLang:    "en",
			wantContentLang: "",
		},
		{
			name: "Language from context",
			req: &model.BidRequest{
				User: model.InternalUser{},
				Context: map[string]interface{}{
					"language":         "fr",
					"content_language": "de",
				},
			},
			wantUserLang:    "fr",
			wantContentLang: "de",
		},
		{
			name: "Page language as content language",
			req: &model.BidRequest{
				User: model.InternalUser{Language: "es"},
				Context: map[string]interface{}{
					"page_language": "pt",
				},
			},
			wantUserLang:    "es",
			wantContentLang: "pt",
		},
		{
			name: "User language takes precedence over context",
			req: &model.BidRequest{
				User: model.InternalUser{Language: "ja"},
				Context: map[string]interface{}{
					"language": "zh",
				},
			},
			wantUserLang:    "ja",
			wantContentLang: "",
		},
		{
			name: "Empty request",
			req: &model.BidRequest{
				User: model.InternalUser{},
			},
			wantUserLang:    "",
			wantContentLang: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userLang, contentLang := service.extractLanguageInfo(tt.req)
			if userLang != tt.wantUserLang {
				t.Errorf("extractLanguageInfo() userLang = %v, want %v", userLang, tt.wantUserLang)
			}
			if contentLang != tt.wantContentLang {
				t.Errorf("extractLanguageInfo() contentLang = %v, want %v", contentLang, tt.wantContentLang)
			}
		})
	}
}

func TestMatchLanguageCode(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name        string
		userLang    string
		targetLang  string
		exactLocale bool
		want        bool
	}{
		{
			name:        "Exact match",
			userLang:    "en-US",
			targetLang:  "en-US",
			exactLocale: true,
			want:        true,
		},
		{
			name:        "Exact match required but different locale",
			userLang:    "en-US",
			targetLang:  "en-GB",
			exactLocale: true,
			want:        false,
		},
		{
			name:        "Primary language match",
			userLang:    "en-US",
			targetLang:  "en",
			exactLocale: false,
			want:        true,
		},
		{
			name:        "Different locale same language",
			userLang:    "en-US",
			targetLang:  "en-GB",
			exactLocale: false,
			want:        true,
		},
		{
			name:        "Underscore format matching",
			userLang:    "en_US",
			targetLang:  "en",
			exactLocale: false,
			want:        true,
		},
		{
			name:        "Case insensitive",
			userLang:    "EN-us",
			targetLang:  "en-US",
			exactLocale: true,
			want:        true,
		},
		{
			name:        "Different languages",
			userLang:    "en",
			targetLang:  "fr",
			exactLocale: false,
			want:        false,
		},
		{
			name:        "Empty user language",
			userLang:    "",
			targetLang:  "en",
			exactLocale: false,
			want:        false,
		},
		{
			name:        "Empty target language",
			userLang:    "en",
			targetLang:  "",
			exactLocale: false,
			want:        false,
		},
		{
			name:        "Whitespace handling",
			userLang:    "  en-US  ",
			targetLang:  "en-us",
			exactLocale: true,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.matchLanguageCode(tt.userLang, tt.targetLang, tt.exactLocale)
			if got != tt.want {
				t.Errorf("matchLanguageCode(%q, %q, %v) = %v, want %v",
					tt.userLang, tt.targetLang, tt.exactLocale, got, tt.want)
			}
		})
	}
}

// ==================== Ad Position Tests ====================

func TestExtractAdPositionInfo(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name            string
		req             *model.BidRequest
		wantPosition    string
		wantAboveFold   bool
		wantViewability float64
	}{
		{
			name: "Position from context",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"ad_position": "header",
					"above_fold":  true,
					"viewability": 0.85,
				},
			},
			wantPosition:    "header",
			wantAboveFold:   true,
			wantViewability: 0.85,
		},
		{
			name: "Position from AdSlot",
			req: &model.BidRequest{
				AdSlot: model.AdSlot{Position: "sidebar"},
			},
			wantPosition:    "sidebar",
			wantAboveFold:   false,
			wantViewability: 0.0,
		},
		{
			name: "Above fold inferred from header position",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"position": "Header_Banner",
				},
			},
			wantPosition:    "Header_Banner",
			wantAboveFold:   true,
			wantViewability: 0.0,
		},
		{
			name: "Below fold inferred from footer position",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"position": "footer",
				},
			},
			wantPosition:    "footer",
			wantAboveFold:   false,
			wantViewability: 0.0,
		},
		{
			name: "Interstitial is above fold",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"position": "interstitial",
				},
			},
			wantPosition:    "interstitial",
			wantAboveFold:   true,
			wantViewability: 0.0,
		},
		{
			name: "Predicted viewability",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"predicted_viewability": 0.72,
				},
			},
			wantPosition:    "unknown",
			wantAboveFold:   false,
			wantViewability: 0.72,
		},
		{
			name:            "Empty request",
			req:             &model.BidRequest{},
			wantPosition:    "unknown",
			wantAboveFold:   false,
			wantViewability: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position, aboveFold, viewability := service.extractAdPositionInfo(tt.req)
			if position != tt.wantPosition {
				t.Errorf("extractAdPositionInfo() position = %v, want %v", position, tt.wantPosition)
			}
			if aboveFold != tt.wantAboveFold {
				t.Errorf("extractAdPositionInfo() aboveFold = %v, want %v", aboveFold, tt.wantAboveFold)
			}
			if viewability != tt.wantViewability {
				t.Errorf("extractAdPositionInfo() viewability = %v, want %v", viewability, tt.wantViewability)
			}
		})
	}
}

func TestIsInterstitialPosition(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name     string
		position string
		want     bool
	}{
		{"Interstitial", "interstitial", true},
		{"Fullscreen", "fullscreen", true},
		{"Full_screen", "full_screen", true},
		{"Overlay", "overlay", true},
		{"Modal", "modal", true},
		{"Header (not interstitial)", "header", false},
		{"Sidebar (not interstitial)", "sidebar", false},
		{"Banner (not interstitial)", "banner", false},
		{"Mixed case interstitial", "INTERSTITIAL_AD", true},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.isInterstitialPosition(tt.position)
			if got != tt.want {
				t.Errorf("isInterstitialPosition(%q) = %v, want %v", tt.position, got, tt.want)
			}
		})
	}
}

func TestIsStickyPosition(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name     string
		position string
		want     bool
	}{
		{"Sticky", "sticky", true},
		{"Fixed", "fixed", true},
		{"Anchor", "anchor", true},
		{"Floating", "floating", true},
		{"Header (not sticky)", "header", false},
		{"Sidebar (not sticky)", "sidebar", false},
		{"Mixed case sticky", "STICKY_BANNER", true},
		{"Fixed footer", "fixed_footer", true},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.isStickyPosition(tt.position)
			if got != tt.want {
				t.Errorf("isStickyPosition(%q) = %v, want %v", tt.position, got, tt.want)
			}
		})
	}
}

// ==================== App Info Tests ====================

func TestExtractAppInfo(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name         string
		req          *model.BidRequest
		wantBundle   string
		wantAppName  string
		wantCategory string
		wantIsInApp  bool
		wantRating   float64
	}{
		{
			name: "Full app info from context",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"bundle_id":    "com.example.app",
					"app_name":     "Example App",
					"app_category": "Games",
					"is_app":       true,
					"app_rating":   4.5,
				},
			},
			wantBundle:   "com.example.app",
			wantAppName:  "Example App",
			wantCategory: "Games",
			wantIsInApp:  true,
			wantRating:   4.5,
		},
		{
			name: "Bundle without explicit is_app flag",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"bundle": "com.company.myapp",
				},
			},
			wantBundle:   "com.company.myapp",
			wantAppName:  "",
			wantCategory: "",
			wantIsInApp:  true, // Inferred from bundle format
			wantRating:   0.0,
		},
		{
			name: "iOS app ID format",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"bundle_id": "id123456789",
				},
			},
			wantBundle:   "id123456789",
			wantAppName:  "",
			wantCategory: "",
			wantIsInApp:  true,
			wantRating:   0.0,
		},
		{
			name: "Alternative context keys",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"category": "Entertainment",
					"in_app":   true,
				},
			},
			wantBundle:   "",
			wantAppName:  "",
			wantCategory: "Entertainment",
			wantIsInApp:  true,
			wantRating:   0.0,
		},
		{
			name:         "Empty request",
			req:          &model.BidRequest{},
			wantBundle:   "",
			wantAppName:  "",
			wantCategory: "",
			wantIsInApp:  false,
			wantRating:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundleID, appName, category, isInApp, appRating := service.extractAppInfo(tt.req)
			if bundleID != tt.wantBundle {
				t.Errorf("extractAppInfo() bundleID = %v, want %v", bundleID, tt.wantBundle)
			}
			if appName != tt.wantAppName {
				t.Errorf("extractAppInfo() appName = %v, want %v", appName, tt.wantAppName)
			}
			if category != tt.wantCategory {
				t.Errorf("extractAppInfo() category = %v, want %v", category, tt.wantCategory)
			}
			if isInApp != tt.wantIsInApp {
				t.Errorf("extractAppInfo() isInApp = %v, want %v", isInApp, tt.wantIsInApp)
			}
			if appRating != tt.wantRating {
				t.Errorf("extractAppInfo() appRating = %v, want %v", appRating, tt.wantRating)
			}
		})
	}
}

func TestMatchBundleID(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name       string
		userBundle string
		ruleBundle string
		want       bool
	}{
		{
			name:       "Exact match",
			userBundle: "com.example.app",
			ruleBundle: "com.example.app",
			want:       true,
		},
		{
			name:       "Wildcard match",
			userBundle: "com.example.app1",
			ruleBundle: "com.example.*",
			want:       true,
		},
		{
			name:       "Wildcard no match",
			userBundle: "com.other.app",
			ruleBundle: "com.example.*",
			want:       false,
		},
		{
			name:       "Case insensitive",
			userBundle: "COM.EXAMPLE.APP",
			ruleBundle: "com.example.app",
			want:       true,
		},
		{
			name:       "Empty user bundle",
			userBundle: "",
			ruleBundle: "com.example.app",
			want:       false,
		},
		{
			name:       "Empty rule bundle",
			userBundle: "com.example.app",
			ruleBundle: "",
			want:       false,
		},
		{
			name:       "Different bundles",
			userBundle: "com.example.app1",
			ruleBundle: "com.example.app2",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.matchBundleID(tt.userBundle, tt.ruleBundle)
			if got != tt.want {
				t.Errorf("matchBundleID(%q, %q) = %v, want %v",
					tt.userBundle, tt.ruleBundle, got, tt.want)
			}
		})
	}
}

func TestMatchAppCategory(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name         string
		appCategory  string
		ruleCategory string
		want         bool
	}{
		{
			name:         "Games category match",
			appCategory:  "gaming",
			ruleCategory: "games",
			want:         true,
		},
		{
			name:         "Social category match",
			appCategory:  "social networking",
			ruleCategory: "social",
			want:         true,
		},
		{
			name:         "News category match",
			appCategory:  "news & magazines",
			ruleCategory: "news",
			want:         true,
		},
		{
			name:         "Finance category match",
			appCategory:  "banking",
			ruleCategory: "finance",
			want:         true,
		},
		{
			name:         "Health category match",
			appCategory:  "health & fitness",
			ruleCategory: "health",
			want:         true,
		},
		{
			name:         "Shopping category match",
			appCategory:  "retail",
			ruleCategory: "shopping",
			want:         true,
		},
		{
			name:         "No match",
			appCategory:  "games",
			ruleCategory: "finance",
			want:         false,
		},
		{
			name:         "Empty app category",
			appCategory:  "",
			ruleCategory: "games",
			want:         false,
		},
		{
			name:         "Empty rule category",
			appCategory:  "games",
			ruleCategory: "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.matchAppCategory(tt.appCategory, tt.ruleCategory)
			if got != tt.want {
				t.Errorf("matchAppCategory(%q, %q) = %v, want %v",
					tt.appCategory, tt.ruleCategory, got, tt.want)
			}
		})
	}
}

func TestIsPremiumApp(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name     string
		bundleID string
		rating   float64
		want     bool
	}{
		{
			name:     "High rating is premium",
			bundleID: "com.unknown.app",
			rating:   4.6,
			want:     true,
		},
		{
			name:     "Spotify is premium",
			bundleID: "com.spotify.music",
			rating:   3.5,
			want:     true,
		},
		{
			name:     "Netflix is premium",
			bundleID: "com.netflix.app",
			rating:   3.0,
			want:     true,
		},
		{
			name:     "Amazon is premium",
			bundleID: "com.amazon.shopping",
			rating:   3.5,
			want:     true,
		},
		{
			name:     "Facebook is premium",
			bundleID: "com.facebook.katana",
			rating:   2.0,
			want:     true,
		},
		{
			name:     "TikTok is premium",
			bundleID: "com.tiktok.app",
			rating:   2.0,
			want:     true,
		},
		{
			name:     "Unknown low rated is not premium",
			bundleID: "com.unknown.app",
			rating:   3.5,
			want:     false,
		},
		{
			name:     "Empty bundle low rating",
			bundleID: "",
			rating:   3.0,
			want:     false,
		},
		{
			name:     "4.5 rating threshold",
			bundleID: "com.random.app",
			rating:   4.5,
			want:     true,
		},
		{
			name:     "Just below threshold",
			bundleID: "com.random.app",
			rating:   4.4,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.isPremiumApp(tt.bundleID, tt.rating)
			if got != tt.want {
				t.Errorf("isPremiumApp(%q, %.1f) = %v, want %v",
					tt.bundleID, tt.rating, got, tt.want)
			}
		})
	}
}

// ==================== Event/Holiday Tests ====================

func TestIsEventActive(t *testing.T) {
	service := createBiddingUtilsService()

	now := time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		event model.SeasonalEvent
		now   time.Time
		want  bool
	}{
		{
			name: "Active recurring event (MM-DD format)",
			event: model.SeasonalEvent{
				StartDate: "07-10",
				EndDate:   "07-20",
				Recurring: true,
			},
			now:  now,
			want: true,
		},
		{
			name: "Inactive recurring event",
			event: model.SeasonalEvent{
				StartDate: "08-01",
				EndDate:   "08-15",
				Recurring: true,
			},
			now:  now,
			want: false,
		},
		{
			name: "Active full date event",
			event: model.SeasonalEvent{
				StartDate: "2024-07-01",
				EndDate:   "2024-07-31",
				Recurring: false,
			},
			now:  now,
			want: true,
		},
		{
			name: "Expired full date event",
			event: model.SeasonalEvent{
				StartDate: "2024-06-01",
				EndDate:   "2024-06-30",
				Recurring: false,
			},
			now:  now,
			want: false,
		},
		{
			name: "Event on start date",
			event: model.SeasonalEvent{
				StartDate: "07-15",
				EndDate:   "07-20",
				Recurring: true,
			},
			now:  now,
			want: true,
		},
		{
			name: "Event on end date",
			event: model.SeasonalEvent{
				StartDate: "07-10",
				EndDate:   "07-15",
				Recurring: true,
			},
			now:  now,
			want: true,
		},
		{
			name: "Invalid date format",
			event: model.SeasonalEvent{
				StartDate: "invalid",
				EndDate:   "invalid",
				Recurring: true,
			},
			now:  now,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.isEventActive(tt.event, tt.now)
			if got != tt.want {
				t.Errorf("isEventActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsEventActive_YearWrap(t *testing.T) {
	service := createBiddingUtilsService()

	// Test year wrap scenario (e.g., holiday season Dec 26 - Jan 2)
	event := model.SeasonalEvent{
		StartDate: "12-26",
		EndDate:   "01-02",
		Recurring: true,
	}

	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{
			name: "December 27 during wrap",
			now:  time.Date(2024, 12, 27, 12, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "January 1 during wrap",
			now:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "December 20 before wrap",
			now:  time.Date(2024, 12, 20, 12, 0, 0, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.isEventActive(event, tt.now)
			if got != tt.want {
				t.Errorf("isEventActive() with year wrap = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHolidayName(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name    string
		date    time.Time
		country string
		want    string
	}{
		{
			name:    "US New Year's Day",
			date:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			country: "US",
			want:    "New Year's Day",
		},
		{
			name:    "US Independence Day",
			date:    time.Date(2024, 7, 4, 12, 0, 0, 0, time.UTC),
			country: "US",
			want:    "Independence Day",
		},
		{
			name:    "US Christmas",
			date:    time.Date(2024, 12, 25, 12, 0, 0, 0, time.UTC),
			country: "US",
			want:    "Christmas",
		},
		{
			name:    "US Christmas Eve",
			date:    time.Date(2024, 12, 24, 12, 0, 0, 0, time.UTC),
			country: "US",
			want:    "Christmas Eve",
		},
		{
			name:    "US New Year's Eve",
			date:    time.Date(2024, 12, 31, 12, 0, 0, 0, time.UTC),
			country: "US",
			want:    "New Year's Eve",
		},
		{
			name:    "US Valentine's Day",
			date:    time.Date(2024, 2, 14, 12, 0, 0, 0, time.UTC),
			country: "US",
			want:    "Valentine's Day",
		},
		{
			name:    "US Halloween",
			date:    time.Date(2024, 10, 31, 12, 0, 0, 0, time.UTC),
			country: "US",
			want:    "Halloween",
		},
		{
			name:    "US Veterans Day",
			date:    time.Date(2024, 11, 11, 12, 0, 0, 0, time.UTC),
			country: "US",
			want:    "Veterans Day",
		},
		{
			name:    "UK Boxing Day",
			date:    time.Date(2024, 12, 26, 12, 0, 0, 0, time.UTC),
			country: "UK",
			want:    "Boxing Day",
		},
		{
			name:    "UK Christmas",
			date:    time.Date(2024, 12, 25, 12, 0, 0, 0, time.UTC),
			country: "UK",
			want:    "Christmas",
		},
		{
			name:    "Regular day (no holiday)",
			date:    time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC),
			country: "US",
			want:    "",
		},
		{
			name:    "Unknown country",
			date:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			country: "XX",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.getHolidayName(tt.date, tt.country)
			if got != tt.want {
				t.Errorf("getHolidayName(%v, %q) = %q, want %q", tt.date, tt.country, got, tt.want)
			}
		})
	}
}

func TestGetHolidayName_DynamicHolidays(t *testing.T) {
	service := createBiddingUtilsService()

	// Test dynamic holidays that vary by year
	tests := []struct {
		name    string
		date    time.Time
		country string
		want    string
	}{
		{
			name:    "Thanksgiving 2024 (4th Thursday Nov)",
			date:    time.Date(2024, 11, 28, 12, 0, 0, 0, time.UTC),
			country: "US",
			want:    "Thanksgiving",
		},
		{
			name:    "Black Friday 2024",
			date:    time.Date(2024, 11, 29, 12, 0, 0, 0, time.UTC),
			country: "US",
			want:    "Black Friday",
		},
		{
			name:    "Labor Day 2024 (1st Monday Sept)",
			date:    time.Date(2024, 9, 2, 12, 0, 0, 0, time.UTC),
			country: "US",
			want:    "Labor Day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.getHolidayName(tt.date, tt.country)
			if got != tt.want {
				t.Errorf("getHolidayName(%v, %q) = %q, want %q", tt.date, tt.country, got, tt.want)
			}
		})
	}
}

func TestGetNthWeekdayOfMonth(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name    string
		year    int
		month   int
		weekday time.Weekday
		n       int
		want    int
	}{
		{
			name:    "1st Monday of September 2024 (Labor Day)",
			year:    2024,
			month:   9,
			weekday: time.Monday,
			n:       1,
			want:    2,
		},
		{
			name:    "4th Thursday of November 2024 (Thanksgiving)",
			year:    2024,
			month:   11,
			weekday: time.Thursday,
			n:       4,
			want:    28,
		},
		{
			name:    "3rd Monday of January 2024 (MLK Day)",
			year:    2024,
			month:   1,
			weekday: time.Monday,
			n:       3,
			want:    15,
		},
		{
			name:    "2nd Sunday of May 2024 (Mother's Day)",
			year:    2024,
			month:   5,
			weekday: time.Sunday,
			n:       2,
			want:    12,
		},
		{
			name:    "3rd Sunday of June 2024 (Father's Day)",
			year:    2024,
			month:   6,
			weekday: time.Sunday,
			n:       3,
			want:    16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.getNthWeekdayOfMonth(tt.year, tt.month, tt.weekday, tt.n)
			if got != tt.want {
				t.Errorf("getNthWeekdayOfMonth(%d, %d, %v, %d) = %d, want %d",
					tt.year, tt.month, tt.weekday, tt.n, got, tt.want)
			}
		})
	}
}

// ==================== Demographic Tests ====================

func TestExtractDemographicInfo(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name            string
		req             *model.BidRequest
		wantAge         int
		wantGender      string
		wantIncomeLevel string
	}{
		{
			name: "From User struct",
			req: &model.BidRequest{
				User: model.InternalUser{
					Age:    35,
					Gender: "M",
				},
			},
			wantAge:         35,
			wantGender:      "male",
			wantIncomeLevel: "",
		},
		{
			name: "From context",
			req: &model.BidRequest{
				User: model.InternalUser{},
				Context: map[string]interface{}{
					"age":          float64(28),
					"gender":       "female",
					"income_level": "high",
				},
			},
			wantAge:         28,
			wantGender:      "female",
			wantIncomeLevel: "high",
		},
		{
			name: "Year of birth calculation",
			req: &model.BidRequest{
				User: model.InternalUser{},
				Context: map[string]interface{}{
					"yob": float64(1990),
				},
			},
			wantAge:         time.Now().Year() - 1990,
			wantGender:      "unknown",
			wantIncomeLevel: "",
		},
		{
			name: "Gender normalization - male variants",
			req: &model.BidRequest{
				User: model.InternalUser{Gender: "Male"},
			},
			wantAge:         0,
			wantGender:      "male",
			wantIncomeLevel: "",
		},
		{
			name: "Gender normalization - female variants",
			req: &model.BidRequest{
				User: model.InternalUser{Gender: "F"},
			},
			wantAge:         0,
			wantGender:      "female",
			wantIncomeLevel: "",
		},
		{
			name: "Gender normalization - other",
			req: &model.BidRequest{
				User: model.InternalUser{Gender: "non-binary"},
			},
			wantAge:         0,
			wantGender:      "other",
			wantIncomeLevel: "",
		},
		{
			name: "Alternative context keys",
			req: &model.BidRequest{
				User: model.InternalUser{},
				Context: map[string]interface{}{
					"user_age":    float64(45),
					"user_gender": "male",
					"income":      "medium",
				},
			},
			wantAge:         45,
			wantGender:      "male",
			wantIncomeLevel: "medium",
		},
		{
			name:            "Empty request",
			req:             &model.BidRequest{},
			wantAge:         0,
			wantGender:      "unknown",
			wantIncomeLevel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			age, gender, incomeLevel := service.extractDemographicInfo(tt.req)
			if age != tt.wantAge {
				t.Errorf("extractDemographicInfo() age = %v, want %v", age, tt.wantAge)
			}
			if gender != tt.wantGender {
				t.Errorf("extractDemographicInfo() gender = %v, want %v", gender, tt.wantGender)
			}
			if incomeLevel != tt.wantIncomeLevel {
				t.Errorf("extractDemographicInfo() incomeLevel = %v, want %v", incomeLevel, tt.wantIncomeLevel)
			}
		})
	}
}
