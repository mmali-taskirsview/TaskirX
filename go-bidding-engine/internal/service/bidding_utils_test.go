package service

import (
	"math"
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// createBiddingUtilsService creates a BiddingService for testing utilities
func createBiddingUtilsService() *BiddingService {
	cache := NewMockCache()
	return NewBiddingService(cache, "http://localhost:8080")
}

// ==================== Bid Shading Tests ====================

func TestCalculateShadedBid(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name      string
		firstBid  float64
		secondBid float64
		bidFloor  float64
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "Standard bid shading",
			firstBid:  5.0,
			secondBid: 3.0,
			bidFloor:  0.5,
			wantMin:   3.0,
			wantMax:   4.0,
		},
		{
			name:      "Equal bids",
			firstBid:  3.0,
			secondBid: 3.0,
			bidFloor:  0.5,
			wantMin:   3.0,
			wantMax:   3.1,
		},
		{
			name:      "Second bid below floor",
			firstBid:  2.0,
			secondBid: 0.1,
			bidFloor:  0.5,
			wantMin:   0.5,
			wantMax:   1.0,
		},
		{
			name:      "Large gap between bids",
			firstBid:  10.0,
			secondBid: 1.0,
			bidFloor:  0.5,
			wantMin:   1.0,
			wantMax:   3.0,
		},
		{
			name:      "Floor exceeds second bid",
			firstBid:  1.0,
			secondBid: 0.3,
			bidFloor:  0.5,
			wantMin:   0.5,
			wantMax:   1.0,
		},
		{
			name:      "Very small gap",
			firstBid:  1.02,
			secondBid: 1.0,
			bidFloor:  0.5,
			wantMin:   1.0,
			wantMax:   1.02,
		},
		{
			name:      "Zero second bid",
			firstBid:  2.0,
			secondBid: 0.0,
			bidFloor:  0.1,
			wantMin:   0.1,
			wantMax:   0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateShadedBid(tt.firstBid, tt.secondBid, tt.bidFloor)
			if result < tt.wantMin {
				t.Errorf("calculateShadedBid() = %v, want >= %v", result, tt.wantMin)
			}
			if result > tt.wantMax {
				t.Errorf("calculateShadedBid() = %v, want <= %v", result, tt.wantMax)
			}
		})
	}
}

func TestCalculateShadedBid_RespectsFloor(t *testing.T) {
	service := createBiddingUtilsService()

	// Shaded bid should never be less than floor + minimum increment
	result := service.calculateShadedBid(0.6, 0.3, 0.5)
	if result < 0.51 {
		t.Errorf("Shaded bid %v should respect floor 0.5", result)
	}
}

func TestCalculateShadedBid_DoesNotExceedFirst(t *testing.T) {
	service := createBiddingUtilsService()

	// When second bid is very close to first, should not exceed first
	result := service.calculateShadedBid(1.0, 0.99, 0.5)
	if result > 1.0 {
		t.Errorf("Shaded bid %v should not exceed first bid 1.0", result)
	}
}

// ==================== Haversine Distance Tests ====================

func TestHaversineDistance(t *testing.T) {
	tests := []struct {
		name    string
		lat1    float64
		lon1    float64
		lat2    float64
		lon2    float64
		wantMin float64
		wantMax float64
	}{
		{
			name:    "Same point",
			lat1:    40.7128,
			lon1:    -74.0060,
			lat2:    40.7128,
			lon2:    -74.0060,
			wantMin: 0,
			wantMax: 0.01,
		},
		{
			name:    "NYC to LA",
			lat1:    40.7128,
			lon1:    -74.0060,
			lat2:    34.0522,
			lon2:    -118.2437,
			wantMin: 3900,
			wantMax: 4000,
		},
		{
			name:    "London to Paris",
			lat1:    51.5074,
			lon1:    -0.1278,
			lat2:    48.8566,
			lon2:    2.3522,
			wantMin: 330,
			wantMax: 350,
		},
		{
			name:    "Small distance - 1km apart",
			lat1:    40.0,
			lon1:    -74.0,
			lat2:    40.009,
			lon2:    -74.0,
			wantMin: 0.9,
			wantMax: 1.1,
		},
		{
			name:    "Equator points",
			lat1:    0.0,
			lon1:    0.0,
			lat2:    0.0,
			lon2:    1.0,
			wantMin: 110,
			wantMax: 112,
		},
		{
			name:    "Antipodal points",
			lat1:    0.0,
			lon1:    0.0,
			lat2:    0.0,
			lon2:    180.0,
			wantMin: 20000,
			wantMax: 20100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := haversineDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			if result < tt.wantMin {
				t.Errorf("haversineDistance() = %v km, want >= %v km", result, tt.wantMin)
			}
			if result > tt.wantMax {
				t.Errorf("haversineDistance() = %v km, want <= %v km", result, tt.wantMax)
			}
		})
	}
}

func TestHaversineDistance_Symmetric(t *testing.T) {
	// Distance A to B should equal distance B to A
	d1 := haversineDistance(40.7128, -74.0060, 34.0522, -118.2437)
	d2 := haversineDistance(34.0522, -118.2437, 40.7128, -74.0060)

	if math.Abs(d1-d2) > 0.001 {
		t.Errorf("Distance not symmetric: %v vs %v", d1, d2)
	}
}

func TestDegreesToRadians(t *testing.T) {
	tests := []struct {
		name    string
		degrees float64
		want    float64
	}{
		{"Zero", 0, 0},
		{"90 degrees", 90, math.Pi / 2},
		{"180 degrees", 180, math.Pi},
		{"360 degrees", 360, 2 * math.Pi},
		{"45 degrees", 45, math.Pi / 4},
		{"Negative", -90, -math.Pi / 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := degreesToRadians(tt.degrees)
			if math.Abs(result-tt.want) > 0.0001 {
				t.Errorf("degreesToRadians(%v) = %v, want %v", tt.degrees, result, tt.want)
			}
		})
	}
}

// ==================== Extract Page Content Tests ====================

func TestExtractPageContent(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name      string
		context   map[string]interface{}
		wantParts []string
	}{
		{
			name: "All fields present",
			context: map[string]interface{}{
				"page_title": "Best Shoes Online",
				"keywords":   "shoes, sneakers, running",
				"content":    "Buy amazing shoes today",
				"page_url":   "https://example.com/shoes",
			},
			wantParts: []string{"Best Shoes Online", "shoes, sneakers, running", "Buy amazing shoes today", "https://example.com/shoes"},
		},
		{
			name: "Only title",
			context: map[string]interface{}{
				"page_title": "News Article",
			},
			wantParts: []string{"News Article"},
		},
		{
			name: "Only keywords",
			context: map[string]interface{}{
				"keywords": "sports, football",
			},
			wantParts: []string{"sports, football"},
		},
		{
			name:      "Empty context",
			context:   map[string]interface{}{},
			wantParts: []string{},
		},
		{
			name:      "Nil context",
			context:   nil,
			wantParts: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &model.BidRequest{Context: tt.context}
			result := service.extractPageContent(req)

			for _, part := range tt.wantParts {
				if len(part) > 0 && !containsStr(result, part) {
					t.Errorf("extractPageContent() = %q, should contain %q", result, part)
				}
			}
		})
	}
}

// ==================== Extract Page Categories Tests ====================

func TestExtractPageCategories(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name    string
		context map[string]interface{}
		want    []string
	}{
		{
			name: "Categories from categories field",
			context: map[string]interface{}{
				"categories": []interface{}{"IAB1", "IAB2-1", "IAB3"},
			},
			want: []string{"IAB1", "IAB2-1", "IAB3"},
		},
		{
			name: "Categories from iab_categories field",
			context: map[string]interface{}{
				"iab_categories": []interface{}{"IAB4", "IAB5"},
			},
			want: []string{"IAB4", "IAB5"},
		},
		{
			name: "Both fields present",
			context: map[string]interface{}{
				"categories":     []interface{}{"IAB1"},
				"iab_categories": []interface{}{"IAB2"},
			},
			want: []string{"IAB1", "IAB2"},
		},
		{
			name:    "Empty context",
			context: map[string]interface{}{},
			want:    []string{},
		},
		{
			name:    "Nil context",
			context: nil,
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &model.BidRequest{Context: tt.context}
			result := service.extractPageCategories(req)

			if len(result) != len(tt.want) {
				t.Errorf("extractPageCategories() returned %d categories, want %d", len(result), len(tt.want))
				return
			}

			for i, cat := range tt.want {
				if result[i] != cat {
					t.Errorf("extractPageCategories()[%d] = %q, want %q", i, result[i], cat)
				}
			}
		})
	}
}

// ==================== Get User Segments Tests ====================

func TestGetUserSegments(t *testing.T) {
	cache := NewMockCache()
	service := &BiddingService{cache: cache}

	tests := []struct {
		name           string
		context        map[string]interface{}
		userCategories []string
		cachedSegments []string
		userID         string
		wantContains   []string
	}{
		{
			name: "From user_segments",
			context: map[string]interface{}{
				"user_segments": []interface{}{"seg1", "seg2"},
			},
			wantContains: []string{"seg1", "seg2"},
		},
		{
			name: "From audience_ids",
			context: map[string]interface{}{
				"audience_ids": []interface{}{"aud1", "aud2"},
			},
			wantContains: []string{"aud1", "aud2"},
		},
		{
			name:           "From user categories",
			context:        map[string]interface{}{},
			userCategories: []string{"cat1", "cat2"},
			wantContains:   []string{"cat1", "cat2"},
		},
		{
			name:    "Empty when no data",
			context: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &model.BidRequest{
				Context: tt.context,
				User: model.InternalUser{
					ID:         tt.userID,
					Categories: tt.userCategories,
				},
			}

			result := service.getUserSegments(req)

			for _, want := range tt.wantContains {
				found := false
				for _, got := range result {
					if got == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("getUserSegments() should contain %q, got %v", want, result)
				}
			}
		})
	}
}

// ==================== Extract Weather Data Tests ====================

func TestExtractWeatherData(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name     string
		context  map[string]interface{}
		wantNil  bool
		wantCond string
		wantTemp float64
		wantHum  int
		wantWind float64
	}{
		{
			name: "Full weather data",
			context: map[string]interface{}{
				"weather":     "sunny",
				"temperature": 25.5,
				"humidity":    60.0,
				"wind_speed":  15.0,
			},
			wantNil:  false,
			wantCond: "sunny",
			wantTemp: 25.5,
			wantHum:  60,
			wantWind: 15.0,
		},
		{
			name: "Weather condition from weather_condition",
			context: map[string]interface{}{
				"weather_condition": "Rainy",
			},
			wantNil:  false,
			wantCond: "rainy",
		},
		{
			name: "Temperature from temp field",
			context: map[string]interface{}{
				"temp": 30.0,
			},
			wantNil:  false,
			wantTemp: 30.0,
		},
		{
			name:    "Empty context returns nil",
			context: map[string]interface{}{},
			wantNil: true,
		},
		{
			name:    "Nil context returns nil",
			context: nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &model.BidRequest{Context: tt.context}
			result := service.extractWeatherData(req)

			if tt.wantNil {
				if result != nil {
					t.Errorf("extractWeatherData() should return nil")
				}
				return
			}

			if result == nil {
				t.Errorf("extractWeatherData() should not return nil")
				return
			}

			if tt.wantCond != "" && result.Condition != tt.wantCond {
				t.Errorf("Condition = %q, want %q", result.Condition, tt.wantCond)
			}
			if tt.wantTemp != 0 && result.Temperature != tt.wantTemp {
				t.Errorf("Temperature = %v, want %v", result.Temperature, tt.wantTemp)
			}
			if tt.wantHum != 0 && result.Humidity != tt.wantHum {
				t.Errorf("Humidity = %v, want %v", result.Humidity, tt.wantHum)
			}
			if tt.wantWind != 0 && result.WindSpeed != tt.wantWind {
				t.Errorf("WindSpeed = %v, want %v", result.WindSpeed, tt.wantWind)
			}
		})
	}
}

// ==================== Match Weather Condition Tests ====================

func TestMatchWeatherCondition(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name      string
		weather   *WeatherData
		condition string
		want      bool
	}{
		{
			name:      "Direct match - sunny",
			weather:   &WeatherData{Condition: "sunny"},
			condition: "sunny",
			want:      true,
		},
		{
			name:      "Synonym match - clear to sunny",
			weather:   &WeatherData{Condition: "clear"},
			condition: "sunny",
			want:      true,
		},
		{
			name:      "Synonym match - rain to rainy",
			weather:   &WeatherData{Condition: "rain"},
			condition: "rainy",
			want:      true,
		},
		{
			name:      "Synonym match - overcast to cloudy",
			weather:   &WeatherData{Condition: "overcast"},
			condition: "cloudy",
			want:      true,
		},
		{
			name:      "Temperature based - hot",
			weather:   &WeatherData{Condition: "clear", Temperature: 35},
			condition: "hot",
			want:      true,
		},
		{
			name:      "Temperature based - cold",
			weather:   &WeatherData{Condition: "clear", Temperature: 2},
			condition: "cold",
			want:      true,
		},
		{
			name:      "No match",
			weather:   &WeatherData{Condition: "sunny"},
			condition: "rainy",
			want:      false,
		},
		{
			name:      "Case insensitive",
			weather:   &WeatherData{Condition: "SUNNY"},
			condition: "Sunny",
			want:      true,
		},
		{
			name:      "Thunder to stormy",
			weather:   &WeatherData{Condition: "thunderstorm"},
			condition: "stormy",
			want:      true,
		},
		{
			name:      "Fog to foggy",
			weather:   &WeatherData{Condition: "fog"},
			condition: "foggy",
			want:      true,
		},
		{
			name:      "Snow to snowy",
			weather:   &WeatherData{Condition: "snow"},
			condition: "snowy",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.matchWeatherCondition(tt.weather, tt.condition)
			if result != tt.want {
				t.Errorf("matchWeatherCondition() = %v, want %v", result, tt.want)
			}
		})
	}
}

// ==================== Extract Network Info Tests ====================

func TestExtractNetworkInfo(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name        string
		context     map[string]interface{}
		wantCarrier string
		wantISP     string
		wantConn    string
	}{
		{
			name: "All network info present",
			context: map[string]interface{}{
				"carrier":         "Verizon",
				"isp":             "Comcast",
				"connection_type": "wifi",
			},
			wantCarrier: "Verizon",
			wantISP:     "Comcast",
			wantConn:    "wifi",
		},
		{
			name: "Numeric connection type - WiFi",
			context: map[string]interface{}{
				"connectiontype": float64(2),
			},
			wantConn: "wifi",
		},
		{
			name: "Numeric connection type - Ethernet",
			context: map[string]interface{}{
				"connectiontype": float64(1),
			},
			wantConn: "ethernet",
		},
		{
			name: "Numeric connection type - Cellular (4G)",
			context: map[string]interface{}{
				"connectiontype": float64(6),
			},
			wantConn: "cellular",
		},
		{
			name: "Numeric connection type - 5G",
			context: map[string]interface{}{
				"connectiontype": float64(7),
			},
			wantConn: "cellular",
		},
		{
			name:     "Empty context",
			context:  map[string]interface{}{},
			wantConn: "unknown",
		},
		{
			name: "String connectiontype",
			context: map[string]interface{}{
				"connectiontype": "cellular",
			},
			wantConn: "cellular",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &model.BidRequest{Context: tt.context}
			carrier, isp, connType := service.extractNetworkInfo(req)

			if tt.wantCarrier != "" && carrier != tt.wantCarrier {
				t.Errorf("carrier = %q, want %q", carrier, tt.wantCarrier)
			}
			if tt.wantISP != "" && isp != tt.wantISP {
				t.Errorf("isp = %q, want %q", isp, tt.wantISP)
			}
			if connType != tt.wantConn {
				t.Errorf("connType = %q, want %q", connType, tt.wantConn)
			}
		})
	}
}

// ==================== Match Carrier Tests ====================

func TestMatchCarrier(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name        string
		userCarrier string
		rule        model.CarrierRule
		want        bool
	}{
		{
			name:        "Direct match",
			userCarrier: "Verizon",
			rule:        model.CarrierRule{Name: "Verizon"},
			want:        true,
		},
		{
			name:        "Case insensitive match",
			userCarrier: "VERIZON",
			rule:        model.CarrierRule{Name: "verizon"},
			want:        true,
		},
		{
			name:        "Partial match",
			userCarrier: "Verizon Wireless",
			rule:        model.CarrierRule{Name: "Verizon"},
			want:        true,
		},
		{
			name:        "Alias match - ATT variations",
			userCarrier: "AT&T",
			rule:        model.CarrierRule{Name: "att"},
			want:        true,
		},
		{
			name:        "Alias match - T-Mobile",
			userCarrier: "tmobile",
			rule:        model.CarrierRule{Name: "t-mobile"},
			want:        true,
		},
		{
			name:        "No match",
			userCarrier: "Sprint",
			rule:        model.CarrierRule{Name: "Verizon"},
			want:        false,
		},
		{
			name:        "Empty user carrier",
			userCarrier: "",
			rule:        model.CarrierRule{Name: "Verizon"},
			want:        false,
		},
		{
			name:        "Vodafone alias",
			userCarrier: "voda",
			rule:        model.CarrierRule{Name: "vodafone"},
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.matchCarrier(tt.userCarrier, tt.rule)
			if result != tt.want {
				t.Errorf("matchCarrier() = %v, want %v", result, tt.want)
			}
		})
	}
}

// ==================== Brand Safety Tests ====================

func TestCheckBrandSafety(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name     string
		campaign *model.Campaign
		req      *model.BidRequest
		wantBlk  bool
		wantMult float64
	}{
		{
			name:     "No restrictions",
			campaign: &model.Campaign{ID: "camp1"},
			req:      &model.BidRequest{PublisherID: "pub1"},
			wantBlk:  false,
			wantMult: 1.0,
		},
		{
			name: "Blocked publisher",
			campaign: &model.Campaign{
				ID:                "camp1",
				BlockedPublishers: []string{"bad-pub"},
			},
			req:     &model.BidRequest{PublisherID: "bad-pub"},
			wantBlk: true,
		},
		{
			name: "Publisher not in blocked list",
			campaign: &model.Campaign{
				ID:                "camp1",
				BlockedPublishers: []string{"bad-pub"},
			},
			req:     &model.BidRequest{PublisherID: "good-pub"},
			wantBlk: false,
		},
		{
			name: "Blocked category",
			campaign: &model.Campaign{
				ID:                "camp1",
				BlockedCategories: []string{"adult"},
			},
			req: &model.BidRequest{
				PublisherID: "pub1",
				Context: map[string]interface{}{
					"categories": []interface{}{"adult", "news"},
				},
			},
			wantBlk: true,
		},
		{
			name: "Blocked keyword in content",
			campaign: &model.Campaign{
				ID:              "camp1",
				BlockedKeywords: []string{"violence"},
			},
			req: &model.BidRequest{
				PublisherID: "pub1",
				Context: map[string]interface{}{
					"content": "Article about violence in movies",
				},
			},
			wantBlk: true,
		},
		{
			name: "Blocked keyword case insensitive",
			campaign: &model.Campaign{
				ID:              "camp1",
				BlockedKeywords: []string{"VIOLENCE"},
			},
			req: &model.BidRequest{
				PublisherID: "pub1",
				Context: map[string]interface{}{
					"content": "violence in games",
				},
			},
			wantBlk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.checkBrandSafety(tt.campaign, tt.req)
			if result.Blocked != tt.wantBlk {
				t.Errorf("Blocked = %v, want %v (reason: %s)", result.Blocked, tt.wantBlk, result.Reason)
			}
			if !tt.wantBlk && tt.wantMult != 0 && result.Multiplier != tt.wantMult {
				t.Errorf("Multiplier = %v, want %v", result.Multiplier, tt.wantMult)
			}
		})
	}
}

// ==================== Extract User Location Tests ====================

func TestExtractUserLocation(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name    string
		req     *model.BidRequest
		wantLat float64
		wantLon float64
		wantOK  bool
	}{
		{
			name: "From device geo",
			req: &model.BidRequest{
				Device: model.InternalDevice{
					Geo: model.InternalGeo{
						Lat: 40.7128,
						Lon: -74.0060,
					},
				},
			},
			wantLat: 40.7128,
			wantLon: -74.0060,
			wantOK:  true,
		},
		{
			name: "From context lat/lon",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"lat": 34.0522,
					"lon": -118.2437,
				},
			},
			wantLat: 34.0522,
			wantLon: -118.2437,
			wantOK:  true,
		},
		{
			name: "From context latitude/longitude",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"latitude":  51.5074,
					"longitude": -0.1278,
				},
			},
			wantLat: 51.5074,
			wantLon: -0.1278,
			wantOK:  true,
		},
		{
			name: "No location data",
			req: &model.BidRequest{
				Context: map[string]interface{}{},
			},
			wantOK: false,
		},
		{
			name:   "Empty request",
			req:    &model.BidRequest{},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lon, ok := service.extractUserLocation(tt.req)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOK {
				if lat != tt.wantLat {
					t.Errorf("lat = %v, want %v", lat, tt.wantLat)
				}
				if lon != tt.wantLon {
					t.Errorf("lon = %v, want %v", lon, tt.wantLon)
				}
			}
		})
	}
}

// ==================== Helper Functions ====================

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) > 0 && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
