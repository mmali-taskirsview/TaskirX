package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ==================== Performance Prediction Tests ====================

func TestGetHistoricalPerformance(t *testing.T) {
	cache := NewMockCache()
	service := NewBiddingService(cache, "http://localhost:8080")

	tests := []struct {
		name       string
		campaignID string
		cacheData  string
		wantCTR    float64
		wantCVR    float64
	}{
		{
			name:       "Default values when no cache",
			campaignID: "camp1",
			wantCTR:    0.01,
			wantCVR:    0.02,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &model.BidRequest{}
			result := service.getHistoricalPerformance(tt.campaignID, req)

			if tt.wantCTR > 0 && result.ctr != tt.wantCTR {
				t.Errorf("ctr = %v, want %v", result.ctr, tt.wantCTR)
			}
			if tt.wantCVR > 0 && result.cvr != tt.wantCVR {
				t.Errorf("cvr = %v, want %v", result.cvr, tt.wantCVR)
			}
		})
	}
}

func TestParsePerformanceCache(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name    string
		cached  string
		wantOK  bool
		wantLen int
	}{
		{
			name:    "Valid cache data",
			cached:  "ctr:0.05,cvr:0.03,impressions:1000",
			wantOK:  true,
			wantLen: 3,
		},
		{
			name:    "Single value",
			cached:  "ctr:0.01",
			wantOK:  true,
			wantLen: 1,
		},
		{
			name:    "Empty string",
			cached:  "",
			wantOK:  false,
			wantLen: 0,
		},
		{
			name:    "Invalid format",
			cached:  "invalid data here",
			wantOK:  false,
			wantLen: 0,
		},
		{
			name:    "With spaces",
			cached:  "ctr: 0.05 , cvr: 0.03",
			wantOK:  true,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := service.parsePerformanceCache(tt.cached)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if len(result) != tt.wantLen {
				t.Errorf("len(result) = %d, want %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestPredictCTR(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name     string
		campaign *model.Campaign
		req      *model.BidRequest
		perf     performanceData
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "Default CTR",
			campaign: &model.Campaign{ID: "camp1"},
			req:      &model.BidRequest{},
			perf:     performanceData{ctr: 0.01},
			wantMin:  0.008,
			wantMax:  0.015,
		},
		{
			name:     "Mobile device boost",
			campaign: &model.Campaign{ID: "camp1"},
			req: &model.BidRequest{
				Device: model.InternalDevice{Type: "mobile"},
			},
			perf:    performanceData{ctr: 0.01},
			wantMin: 0.01,
			wantMax: 0.015,
		},
		{
			name:     "Desktop device",
			campaign: &model.Campaign{ID: "camp1"},
			req: &model.BidRequest{
				Device: model.InternalDevice{Type: "desktop"},
			},
			perf:    performanceData{ctr: 0.01},
			wantMin: 0.008,
			wantMax: 0.012,
		},
		{
			name:     "Above fold placement",
			campaign: &model.Campaign{ID: "camp1"},
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"ad_position": "above_fold",
				},
			},
			perf:    performanceData{ctr: 0.01},
			wantMin: 0.012,
			wantMax: 0.02,
		},
		{
			name: "Category overlap boost",
			campaign: &model.Campaign{
				ID: "camp1",
				Targeting: model.Targeting{
					Categories: []string{"sports", "fitness"},
				},
			},
			req: &model.BidRequest{
				User: model.InternalUser{
					Categories: []string{"sports", "health"},
				},
			},
			perf:    performanceData{ctr: 0.01},
			wantMin: 0.01,
			wantMax: 0.02,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.predictCTR(tt.campaign, tt.req, tt.perf)
			if result < tt.wantMin {
				t.Errorf("predictCTR() = %v, want >= %v", result, tt.wantMin)
			}
			if result > tt.wantMax {
				t.Errorf("predictCTR() = %v, want <= %v", result, tt.wantMax)
			}
		})
	}
}

func TestPredictCVR(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name     string
		campaign *model.Campaign
		req      *model.BidRequest
		perf     performanceData
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "Default CVR",
			campaign: &model.Campaign{ID: "camp1"},
			req:      &model.BidRequest{},
			perf:     performanceData{cvr: 0.02},
			wantMin:  0.01,
			wantMax:  0.05,
		},
		{
			name:     "High intent segment boost",
			campaign: &model.Campaign{ID: "camp1"},
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"user_segments": []interface{}{"high_intent_buyer"},
				},
			},
			perf:    performanceData{cvr: 0.02},
			wantMin: 0.02,
			wantMax: 0.06,
		},
		{
			name: "Category match boost",
			campaign: &model.Campaign{
				ID: "camp1",
				Targeting: model.Targeting{
					Categories: []string{"electronics"},
				},
			},
			req: &model.BidRequest{
				User: model.InternalUser{
					Categories: []string{"electronics", "gadgets"},
				},
			},
			perf:    performanceData{cvr: 0.02},
			wantMin: 0.02,
			wantMax: 0.05,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.predictCVR(tt.campaign, tt.req, tt.perf)
			if result < tt.wantMin {
				t.Errorf("predictCVR() = %v, want >= %v", result, tt.wantMin)
			}
			if result > tt.wantMax {
				t.Errorf("predictCVR() = %v, want <= %v", result, tt.wantMax)
			}
		})
	}
}

func TestPredictViewability(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name    string
		req     *model.BidRequest
		perf    performanceData
		wantMin float64
		wantMax float64
	}{
		{
			name:    "Default viewability",
			req:     &model.BidRequest{},
			perf:    performanceData{viewability: 0.60},
			wantMin: 0.55,
			wantMax: 0.65,
		},
		{
			name: "Above fold boost",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"ad_position": "above_fold",
				},
			},
			perf:    performanceData{viewability: 0.60},
			wantMin: 0.70,
			wantMax: 0.85,
		},
		{
			name: "Below fold decrease",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"ad_position": "below_fold",
				},
			},
			perf:    performanceData{viewability: 0.60},
			wantMin: 0.35,
			wantMax: 0.50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.predictViewability(&model.Campaign{}, tt.req, tt.perf)
			if result < tt.wantMin {
				t.Errorf("predictViewability() = %v, want >= %v", result, tt.wantMin)
			}
			if result > tt.wantMax {
				t.Errorf("predictViewability() = %v, want <= %v", result, tt.wantMax)
			}
		})
	}
}

// ==================== Optimizer Tests ====================

func TestOptimizeForCPA(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name    string
		pg      *model.PerformanceGoals
		perf    performanceData
		wantMin float64
		wantMax float64
	}{
		{
			name:    "No CPA target",
			pg:      &model.PerformanceGoals{TargetCPA: 0},
			perf:    performanceData{},
			wantMin: 1.0,
			wantMax: 1.0,
		},
		{
			name:    "With CPA target",
			pg:      &model.PerformanceGoals{TargetCPA: 10.0},
			perf:    performanceData{ctr: 0.01, cvr: 0.02},
			wantMin: 0.3,
			wantMax: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign := &model.Campaign{ID: "camp1", BidPrice: 1.0}
			req := &model.BidRequest{}
			result := service.optimizeForCPA(campaign, req, tt.pg, tt.perf)
			if result < tt.wantMin {
				t.Errorf("optimizeForCPA() = %v, want >= %v", result, tt.wantMin)
			}
			if result > tt.wantMax {
				t.Errorf("optimizeForCPA() = %v, want <= %v", result, tt.wantMax)
			}
		})
	}
}

func TestOptimizeForCPC(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name    string
		pg      *model.PerformanceGoals
		perf    performanceData
		wantMin float64
		wantMax float64
	}{
		{
			name:    "No CPC target",
			pg:      &model.PerformanceGoals{TargetCPC: 0},
			perf:    performanceData{},
			wantMin: 1.0,
			wantMax: 1.0,
		},
		{
			name:    "With CPC target",
			pg:      &model.PerformanceGoals{TargetCPC: 0.50},
			perf:    performanceData{ctr: 0.02},
			wantMin: 0.3,
			wantMax: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign := &model.Campaign{ID: "camp1", BidPrice: 1.0}
			req := &model.BidRequest{}
			result := service.optimizeForCPC(campaign, req, tt.pg, tt.perf)
			if result < tt.wantMin {
				t.Errorf("optimizeForCPC() = %v, want >= %v", result, tt.wantMin)
			}
			if result > tt.wantMax {
				t.Errorf("optimizeForCPC() = %v, want <= %v", result, tt.wantMax)
			}
		})
	}
}

func TestOptimizeForCPM(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name    string
		pg      *model.PerformanceGoals
		perf    performanceData
		wantMin float64
		wantMax float64
	}{
		{
			name:    "No CPM target",
			pg:      &model.PerformanceGoals{TargetCPM: 0},
			perf:    performanceData{},
			wantMin: 1.0,
			wantMax: 1.0,
		},
		{
			name:    "High viewability boost",
			pg:      &model.PerformanceGoals{TargetCPM: 5.0},
			perf:    performanceData{viewability: 0.85},
			wantMin: 1.2,
			wantMax: 1.4,
		},
		{
			name:    "Low viewability decrease",
			pg:      &model.PerformanceGoals{TargetCPM: 5.0},
			perf:    performanceData{viewability: 0.30},
			wantMin: 0.6,
			wantMax: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign := &model.Campaign{ID: "camp1"}
			req := &model.BidRequest{}
			result := service.optimizeForCPM(campaign, req, tt.pg, tt.perf)
			if result < tt.wantMin {
				t.Errorf("optimizeForCPM() = %v, want >= %v", result, tt.wantMin)
			}
			if result > tt.wantMax {
				t.Errorf("optimizeForCPM() = %v, want <= %v", result, tt.wantMax)
			}
		})
	}
}

// ==================== CTV Tests ====================

func TestIsPrimetime(t *testing.T) {
	service := createBiddingUtilsService()

	// Note: This test may be time-dependent
	// We test both with and without timezone context
	tests := []struct {
		name string
		req  *model.BidRequest
	}{
		{
			name: "No timezone context",
			req:  &model.BidRequest{},
		},
		{
			name: "With timezone context",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"timezone": "America/New_York",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic and returns a boolean
			result := service.isPrimetime(tt.req)
			_ = result // Result depends on current time
		})
	}
}

func TestIsLiveContent(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name string
		req  *model.BidRequest
		want bool
	}{
		{
			name: "is_live true",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"is_live": true,
				},
			},
			want: true,
		},
		{
			name: "is_live false",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"is_live": false,
				},
			},
			want: false,
		},
		{
			name: "Content type live",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"content_type": "live_stream",
				},
			},
			want: true,
		},
		{
			name: "Content type VOD",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"content_type": "vod",
				},
			},
			want: false,
		},
		{
			name: "No context",
			req:  &model.BidRequest{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isLiveContent(tt.req)
			if result != tt.want {
				t.Errorf("isLiveContent() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestIsCoViewing(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name string
		req  *model.BidRequest
		want bool
	}{
		{
			name: "co_viewing true",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"co_viewing": true,
				},
			},
			want: true,
		},
		{
			name: "co_viewing false",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"co_viewing": false,
				},
			},
			want: false,
		},
		{
			name: "Multiple household viewers",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"household_viewers": float64(3),
				},
			},
			want: true,
		},
		{
			name: "Single viewer",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"household_viewers": float64(1),
				},
			},
			want: false,
		},
		{
			name: "No context",
			req:  &model.BidRequest{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isCoViewing(tt.req)
			if result != tt.want {
				t.Errorf("isCoViewing() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGetCTVDevice(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name string
		req  *model.BidRequest
		want string
	}{
		{
			name: "From ctv_device context",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"ctv_device": "roku",
				},
			},
			want: "roku",
		},
		{
			name: "From device_make context",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"device_make": "samsung",
				},
			},
			want: "samsung",
		},
		{
			name: "Fallback to device type",
			req: &model.BidRequest{
				Device: model.InternalDevice{Type: "ctv"},
			},
			want: "ctv",
		},
		{
			name: "Empty request",
			req:  &model.BidRequest{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getCTVDevice(tt.req)
			if result != tt.want {
				t.Errorf("getCTVDevice() = %q, want %q", result, tt.want)
			}
		})
	}
}

// ==================== App Install Tests ====================

func TestIsIOSDevice(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name string
		req  *model.BidRequest
		want bool
	}{
		{
			name: "iOS device",
			req: &model.BidRequest{
				Device: model.InternalDevice{OS: "iOS"},
			},
			want: true,
		},
		{
			name: "iPhone device",
			req: &model.BidRequest{
				Device: model.InternalDevice{OS: "iphone"},
			},
			want: true,
		},
		{
			name: "Android device",
			req: &model.BidRequest{
				Device: model.InternalDevice{OS: "Android"},
			},
			want: false,
		},
		{
			name: "Empty OS",
			req:  &model.BidRequest{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isIOSDevice(tt.req)
			if result != tt.want {
				t.Errorf("isIOSDevice() = %v, want %v", result, tt.want)
			}
		})
	}
}

// ==================== Ecommerce Tests ====================

func TestIsCartAbandoner(t *testing.T) {
	cache := NewMockCache()
	service := NewBiddingService(cache, "http://localhost:8080")

	tests := []struct {
		name string
		req  *model.BidRequest
		want bool
	}{
		{
			name: "From context",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"cart_abandoner": true,
				},
			},
			want: true,
		},
		{
			name: "Not abandoner from context",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"cart_abandoner": false,
				},
			},
			want: false,
		},
		{
			name: "No context",
			req:  &model.BidRequest{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isCartAbandoner(tt.req)
			if result != tt.want {
				t.Errorf("isCartAbandoner() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestIsRepeatCustomer(t *testing.T) {
	cache := NewMockCache()
	service := NewBiddingService(cache, "http://localhost:8080")

	tests := []struct {
		name string
		req  *model.BidRequest
		want bool
	}{
		{
			name: "From context",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"repeat_customer": true,
				},
			},
			want: true,
		},
		{
			name: "Has purchase history",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"purchase_count": float64(5),
				},
			},
			want: true,
		},
		{
			name: "No purchase history",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"purchase_count": float64(0),
				},
			},
			want: false,
		},
		{
			name: "No context",
			req:  &model.BidRequest{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isRepeatCustomer(tt.req)
			if result != tt.want {
				t.Errorf("isRepeatCustomer() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGetCurrentSeason(t *testing.T) {
	service := createBiddingUtilsService()

	// The result depends on current date
	result := service.getCurrentSeason()

	// Verify it returns a valid season
	validSeasons := []string{"spring", "summer", "fall", "winter"}
	valid := false
	for _, s := range validSeasons {
		if result == s {
			valid = true
			break
		}
	}
	if !valid {
		t.Errorf("getCurrentSeason() = %q, want one of %v", result, validSeasons)
	}
}

// ==================== Carrier Multiplier Tests ====================

func TestCalculateCarrierMultiplier(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name     string
		campaign *model.Campaign
		req      *model.BidRequest
		wantBlk  bool
	}{
		{
			name: "No carrier targeting",
			campaign: &model.Campaign{
				ID:        "camp1",
				Targeting: model.Targeting{},
			},
			req:     &model.BidRequest{},
			wantBlk: false,
		},
		{
			name: "Cellular only - wifi user",
			campaign: &model.Campaign{
				ID: "camp1",
				Targeting: model.Targeting{
					CarrierTargeting: &model.CarrierTargeting{
						CellularOnly: true,
					},
				},
			},
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"connection_type": "wifi",
				},
			},
			wantBlk: true,
		},
		{
			name: "Cellular only - cellular user",
			campaign: &model.Campaign{
				ID: "camp1",
				Targeting: model.Targeting{
					CarrierTargeting: &model.CarrierTargeting{
						CellularOnly: true,
					},
				},
			},
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"connection_type": "cellular",
				},
			},
			wantBlk: false,
		},
		{
			name: "WiFi only - cellular user",
			campaign: &model.Campaign{
				ID: "camp1",
				Targeting: model.Targeting{
					CarrierTargeting: &model.CarrierTargeting{
						WiFiOnly: true,
					},
				},
			},
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"connection_type": "cellular",
				},
			},
			wantBlk: true,
		},
		{
			name: "Excluded carrier",
			campaign: &model.Campaign{
				ID: "camp1",
				Targeting: model.Targeting{
					CarrierTargeting: &model.CarrierTargeting{
						ExcludeCarriers: []string{"Sprint"},
					},
				},
			},
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"carrier": "Sprint",
				},
			},
			wantBlk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateCarrierMultiplier(tt.campaign, tt.req)
			if result.Blocked != tt.wantBlk {
				t.Errorf("Blocked = %v, want %v (reason: %s)", result.Blocked, tt.wantBlk, result.Reason)
			}
		})
	}
}
