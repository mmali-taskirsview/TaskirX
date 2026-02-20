package model

import (
	"testing"
)

// ============================================================================
// GEOFENCING TESTS
// ============================================================================

func TestIsMatch_GeoFencing(t *testing.T) {
	// Campaign with GeoFence: NYC (approx 40.7128, -74.0060) within 10km
	campaign := &Campaign{
		Creative: Creative{Type: "banner"},
		Targeting: Targeting{
			GeoFences: []GeoFence{
				{Lat: 40.7128, Lon: -74.0060, Radius: 10.0},
			},
		},
	}

	tests := []struct {
		name     string
		device   InternalDevice
		expected bool
	}{
		{
			name: "Match inside GeoFence",
			device: InternalDevice{
				Type: "mobile",
				Geo: InternalGeo{
					Lat: 40.7200, // Very close
					Lon: -74.0100,
				},
			},
			expected: true,
		},
		{
			name: "No Match outside GeoFence (London)",
			device: InternalDevice{
				Type: "mobile",
				Geo: InternalGeo{
					Lat: 51.5074,
					Lon: -0.1278,
				},
			},
			expected: false,
		},
		{
			name: "No Match border case (20km away)",
			device: InternalDevice{
				Type: "mobile",
				Geo: InternalGeo{
					Lat: 40.7128 + 0.2, // ~22km diff in lat approx
					Lon: -74.0060,
				},
			},
			expected: false,
		},
		{
			name: "No Geo in Request",
			device: InternalDevice{
				Type: "mobile",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &BidRequest{
				Device: tt.device,
				AdSlot: AdSlot{Formats: []string{"banner"}},
			}
			if got := campaign.IsMatch(req); got != tt.expected {
				t.Errorf("IsMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// REQUEST ID GENERATION TESTS
// ============================================================================

func TestGenerateRequestID(t *testing.T) {
	// Test that GenerateRequestID returns a non-empty unique string
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()

	if id1 == "" {
		t.Error("GenerateRequestID() returned empty string")
	}

	if id1 == id2 {
		t.Error("GenerateRequestID() returned duplicate IDs")
	}

	// UUID format check (36 chars with dashes)
	if len(id1) != 36 {
		t.Errorf("GenerateRequestID() expected UUID format (36 chars), got %d chars", len(id1))
	}
}

// ============================================================================
// COUNTRY TARGETING TESTS
// ============================================================================

func TestIsMatch_CountryTargeting(t *testing.T) {
	campaign := &Campaign{
		Creative: Creative{Type: "banner"},
		Targeting: Targeting{
			Countries: []string{"US", "CA", "GB"},
		},
	}

	tests := []struct {
		name     string
		country  string
		expected bool
	}{
		{"Match US", "US", true},
		{"Match CA", "CA", true},
		{"Match GB", "GB", true},
		{"No Match DE", "DE", false},
		{"No Match empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &BidRequest{
				User:   InternalUser{Country: tt.country},
				AdSlot: AdSlot{Formats: []string{"banner"}},
			}
			if got := campaign.IsMatch(req); got != tt.expected {
				t.Errorf("IsMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// DEVICE TARGETING TESTS
// ============================================================================

func TestIsMatch_DeviceTargeting(t *testing.T) {
	campaign := &Campaign{
		Creative: Creative{Type: "banner"},
		Targeting: Targeting{
			Devices: []string{"mobile", "tablet"},
		},
	}

	tests := []struct {
		name       string
		deviceType string
		expected   bool
	}{
		{"Match mobile", "mobile", true},
		{"Match tablet", "tablet", true},
		{"No Match desktop", "desktop", false},
		{"No Match ctv", "ctv", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &BidRequest{
				Device: InternalDevice{Type: tt.deviceType},
				AdSlot: AdSlot{Formats: []string{"banner"}},
			}
			if got := campaign.IsMatch(req); got != tt.expected {
				t.Errorf("IsMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// OS TARGETING TESTS
// ============================================================================

func TestIsMatch_OSTargeting(t *testing.T) {
	campaign := &Campaign{
		Creative: Creative{Type: "banner"},
		Targeting: Targeting{
			OS: []string{"ios", "android"},
		},
	}

	tests := []struct {
		name     string
		os       string
		expected bool
	}{
		{"Match iOS", "ios", true},
		{"Match Android", "android", true},
		{"No Match Windows", "windows", false},
		{"No Match MacOS", "macos", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &BidRequest{
				Device: InternalDevice{OS: tt.os},
				AdSlot: AdSlot{Formats: []string{"banner"}},
			}
			if got := campaign.IsMatch(req); got != tt.expected {
				t.Errorf("IsMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// AGE TARGETING TESTS
// ============================================================================

func TestIsMatch_AgeTargeting(t *testing.T) {
	campaign := &Campaign{
		Creative: Creative{Type: "banner"},
		Targeting: Targeting{
			MinAge: 25,
			MaxAge: 54,
		},
	}

	tests := []struct {
		name     string
		age      int
		expected bool
	}{
		{"Match middle age 35", 35, true},
		{"Match min boundary 25", 25, true},
		{"Match max boundary 54", 54, true},
		{"No Match below min 18", 18, false},
		{"No Match above max 65", 65, false},
		{"No Match zero age", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &BidRequest{
				User:   InternalUser{Age: tt.age},
				AdSlot: AdSlot{Formats: []string{"banner"}},
			}
			if got := campaign.IsMatch(req); got != tt.expected {
				t.Errorf("IsMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// CATEGORY TARGETING TESTS
// ============================================================================

func TestIsMatch_CategoryTargeting(t *testing.T) {
	campaign := &Campaign{
		Creative: Creative{Type: "banner"},
		Targeting: Targeting{
			Categories: []string{"sports", "tech", "gaming"},
		},
	}

	tests := []struct {
		name           string
		userCategories []string
		expected       bool
	}{
		{"Match single category", []string{"sports"}, true},
		{"Match multiple overlap", []string{"tech", "news"}, true},
		{"Match all categories", []string{"sports", "tech", "gaming"}, true},
		{"No Match different categories", []string{"finance", "travel"}, false},
		// Empty user categories means category check is skipped (per IsMatch logic)
		{"Empty user categories skips check", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &BidRequest{
				User:   InternalUser{Categories: tt.userCategories},
				AdSlot: AdSlot{Formats: []string{"banner"}},
			}
			if got := campaign.IsMatch(req); got != tt.expected {
				t.Errorf("IsMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// CREATIVE FORMAT TESTS
// ============================================================================

func TestIsMatch_CreativeFormat(t *testing.T) {
	tests := []struct {
		name         string
		creativeType string
		adFormats    []string
		expected     bool
	}{
		{"Banner matches banner", "banner", []string{"banner"}, true},
		{"Video matches video", "video", []string{"video"}, true},
		{"Native matches native", "native", []string{"native"}, true},
		{"Banner not in video", "banner", []string{"video"}, false},
		{"Video matches interstitial", "video", []string{"interstitial"}, true},
		{"Rich media matches interstitial", "rich_media", []string{"interstitial"}, true},
		{"Banner not in interstitial", "banner", []string{"interstitial"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign := &Campaign{
				Creative: Creative{Type: tt.creativeType},
			}
			req := &BidRequest{
				AdSlot: AdSlot{Formats: tt.adFormats},
			}
			if got := campaign.IsMatch(req); got != tt.expected {
				t.Errorf("IsMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// PMP DEAL TESTS
// ============================================================================

func TestIsMatch_PMPDeals(t *testing.T) {
	tests := []struct {
		name     string
		campaign *Campaign
		request  *BidRequest
		expected bool
	}{
		{
			name: "Campaign with deal matches request deal",
			campaign: &Campaign{
				DealID:   "deal-001",
				BidPrice: 5.00,
				Creative: Creative{Type: "banner"},
			},
			request: &BidRequest{
				AdSlot: AdSlot{Formats: []string{"banner"}},
				Pmp: &Pmp{
					Deals: []Deal{
						{ID: "deal-001", BidFloor: 4.00},
					},
				},
			},
			expected: true,
		},
		{
			name: "Campaign deal not in request",
			campaign: &Campaign{
				DealID:   "deal-001",
				BidPrice: 5.00,
				Creative: Creative{Type: "banner"},
			},
			request: &BidRequest{
				AdSlot: AdSlot{Formats: []string{"banner"}},
				Pmp: &Pmp{
					Deals: []Deal{
						{ID: "deal-002", BidFloor: 4.00},
					},
				},
			},
			expected: false,
		},
		{
			name: "Campaign bid below deal floor",
			campaign: &Campaign{
				DealID:   "deal-001",
				BidPrice: 3.00,
				Creative: Creative{Type: "banner"},
			},
			request: &BidRequest{
				AdSlot: AdSlot{Formats: []string{"banner"}},
				Pmp: &Pmp{
					Deals: []Deal{
						{ID: "deal-001", BidFloor: 5.00},
					},
				},
			},
			expected: false,
		},
		{
			name: "Private auction with no deal campaign",
			campaign: &Campaign{
				BidPrice: 5.00,
				Creative: Creative{Type: "banner"},
			},
			request: &BidRequest{
				AdSlot: AdSlot{Formats: []string{"banner"}},
				Pmp: &Pmp{
					PrivateAuction: 1,
					Deals: []Deal{
						{ID: "deal-001", BidFloor: 4.00},
					},
				},
			},
			expected: false,
		},
		{
			name: "Campaign needs deal but no PMP in request",
			campaign: &Campaign{
				DealID:   "deal-001",
				BidPrice: 5.00,
				Creative: Creative{Type: "banner"},
			},
			request: &BidRequest{
				AdSlot: AdSlot{Formats: []string{"banner"}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.campaign.IsMatch(tt.request); got != tt.expected {
				t.Errorf("IsMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// COMBINED TARGETING TESTS
// ============================================================================

func TestIsMatch_CombinedTargeting(t *testing.T) {
	campaign := &Campaign{
		Creative: Creative{Type: "video"},
		Targeting: Targeting{
			Countries: []string{"US"},
			Devices:   []string{"mobile"},
			OS:        []string{"ios"},
			MinAge:    18,
			MaxAge:    65,
		},
	}

	tests := []struct {
		name     string
		request  *BidRequest
		expected bool
	}{
		{
			name: "All criteria match",
			request: &BidRequest{
				User:   InternalUser{Country: "US", Age: 30},
				Device: InternalDevice{Type: "mobile", OS: "ios"},
				AdSlot: AdSlot{Formats: []string{"video"}},
			},
			expected: true,
		},
		{
			name: "Wrong country",
			request: &BidRequest{
				User:   InternalUser{Country: "DE", Age: 30},
				Device: InternalDevice{Type: "mobile", OS: "ios"},
				AdSlot: AdSlot{Formats: []string{"video"}},
			},
			expected: false,
		},
		{
			name: "Wrong device",
			request: &BidRequest{
				User:   InternalUser{Country: "US", Age: 30},
				Device: InternalDevice{Type: "desktop", OS: "ios"},
				AdSlot: AdSlot{Formats: []string{"video"}},
			},
			expected: false,
		},
		{
			name: "Wrong OS",
			request: &BidRequest{
				User:   InternalUser{Country: "US", Age: 30},
				Device: InternalDevice{Type: "mobile", OS: "android"},
				AdSlot: AdSlot{Formats: []string{"video"}},
			},
			expected: false,
		},
		{
			name: "Too young",
			request: &BidRequest{
				User:   InternalUser{Country: "US", Age: 16},
				Device: InternalDevice{Type: "mobile", OS: "ios"},
				AdSlot: AdSlot{Formats: []string{"video"}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := campaign.IsMatch(tt.request); got != tt.expected {
				t.Errorf("IsMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// ============================================================================
// HELPER FUNCTION TESTS
// ============================================================================

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{"Found at start", []string{"a", "b", "c"}, "a", true},
		{"Found in middle", []string{"a", "b", "c"}, "b", true},
		{"Found at end", []string{"a", "b", "c"}, "c", true},
		{"Not found", []string{"a", "b", "c"}, "d", false},
		{"Empty slice", []string{}, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.slice, tt.item); got != tt.expected {
				t.Errorf("contains() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHasOverlap(t *testing.T) {
	tests := []struct {
		name     string
		slice1   []string
		slice2   []string
		expected bool
	}{
		{"Single overlap", []string{"a", "b"}, []string{"b", "c"}, true},
		{"Multiple overlap", []string{"a", "b", "c"}, []string{"a", "c"}, true},
		{"No overlap", []string{"a", "b"}, []string{"c", "d"}, false},
		{"Empty first", []string{}, []string{"a", "b"}, false},
		{"Empty second", []string{"a", "b"}, []string{}, false},
		{"Both empty", []string{}, []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasOverlap(tt.slice1, tt.slice2); got != tt.expected {
				t.Errorf("hasOverlap() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHaversine(t *testing.T) {
	tests := []struct {
		name    string
		lat1    float64
		lon1    float64
		lat2    float64
		lon2    float64
		minDist float64
		maxDist float64
	}{
		{
			name: "NYC to itself",
			lat1: 40.7128, lon1: -74.0060,
			lat2: 40.7128, lon2: -74.0060,
			minDist: 0, maxDist: 0.001,
		},
		{
			name: "NYC to LA (approx 3940km)",
			lat1: 40.7128, lon1: -74.0060,
			lat2: 34.0522, lon2: -118.2437,
			minDist: 3900, maxDist: 4000,
		},
		{
			name: "London to Paris (approx 340km)",
			lat1: 51.5074, lon1: -0.1278,
			lat2: 48.8566, lon2: 2.3522,
			minDist: 330, maxDist: 350,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := haversine(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			if dist < tt.minDist || dist > tt.maxDist {
				t.Errorf("haversine() = %v, want between %v and %v", dist, tt.minDist, tt.maxDist)
			}
		})
	}
}

// ============================================================================
// NO TARGETING (MATCH ALL) TESTS
// ============================================================================

func TestIsMatch_NoTargeting(t *testing.T) {
	// Campaign with no targeting should match any request with correct format
	campaign := &Campaign{
		Creative: Creative{Type: "banner"},
	}

	req := &BidRequest{
		User:   InternalUser{Country: "XX", Age: 99},
		Device: InternalDevice{Type: "unknown", OS: "alien_os"},
		AdSlot: AdSlot{Formats: []string{"banner"}},
	}

	if !campaign.IsMatch(req) {
		t.Error("Campaign with no targeting should match any request")
	}
}
