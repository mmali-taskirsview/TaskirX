package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ==================== Bid Strategy Tests ====================

func TestApplyBidStrategy(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name    string
		pg      *model.PerformanceGoals
		perf    performanceData
		wantMin float64
		wantMax float64
	}{
		{
			name: "Maximize conversions - high CVR",
			pg: &model.PerformanceGoals{
				BidStrategy: "maximize_conversions",
			},
			perf: performanceData{
				cvr: 0.05, // Above 0.03 threshold
			},
			wantMin: 1.39,
			wantMax: 1.41,
		},
		{
			name: "Maximize conversions - low CVR",
			pg: &model.PerformanceGoals{
				BidStrategy: "maximize_conversions",
			},
			perf: performanceData{
				cvr: 0.02, // Below threshold
			},
			wantMin: 0.99,
			wantMax: 1.01,
		},
		{
			name: "Target CPA - room to increase",
			pg: &model.PerformanceGoals{
				BidStrategy: "target_cpa",
				TargetCPA:   10.0,
			},
			perf: performanceData{
				cpa: 7.0, // Ratio is 10/7 = 1.43, capped at 1.2
			},
			wantMin: 1.19,
			wantMax: 1.21,
		},
		{
			name: "Target CPA - need to decrease",
			pg: &model.PerformanceGoals{
				BidStrategy: "target_cpa",
				TargetCPA:   10.0,
			},
			perf: performanceData{
				cpa: 15.0, // Ratio is 10/15 = 0.67, capped at 0.8
			},
			wantMin: 0.79,
			wantMax: 0.81,
		},
		{
			name: "Target CPA - on target",
			pg: &model.PerformanceGoals{
				BidStrategy: "target_cpa",
				TargetCPA:   10.0,
			},
			perf: performanceData{
				cpa: 10.0, // Ratio is 1.0
			},
			wantMin: 0.99,
			wantMax: 1.01,
		},
		{
			name: "Maximize clicks - high CTR",
			pg: &model.PerformanceGoals{
				BidStrategy: "maximize_clicks",
			},
			perf: performanceData{
				ctr: 0.02, // Above 0.015 threshold
			},
			wantMin: 1.29,
			wantMax: 1.31,
		},
		{
			name: "Maximize clicks - low CTR",
			pg: &model.PerformanceGoals{
				BidStrategy: "maximize_clicks",
			},
			perf: performanceData{
				ctr: 0.01, // Below threshold
			},
			wantMin: 0.99,
			wantMax: 1.01,
		},
		{
			name: "Manual strategy",
			pg: &model.PerformanceGoals{
				BidStrategy: "manual",
			},
			perf: performanceData{
				ctr: 0.05,
				cvr: 0.10,
			},
			wantMin: 0.99,
			wantMax: 1.01,
		},
		{
			name: "Unknown strategy",
			pg: &model.PerformanceGoals{
				BidStrategy: "unknown_strategy",
			},
			perf:    performanceData{},
			wantMin: 0.99,
			wantMax: 1.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.applyBidStrategy(tt.pg, tt.perf)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("applyBidStrategy() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestDetermineOptimizationLevel(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name string
		pg   *model.PerformanceGoals
		perf performanceData
		want string
	}{
		{
			name: "Learning mode = conservative",
			pg: &model.PerformanceGoals{
				LearningMode: true,
			},
			perf: performanceData{
				impressions: 10000,
				cpa:         5.0,
			},
			want: "conservative",
		},
		{
			name: "Low impressions = conservative",
			pg: &model.PerformanceGoals{
				LearningMode: false,
			},
			perf: performanceData{
				impressions: 500, // Below 1000 threshold
			},
			want: "conservative",
		},
		{
			name: "Performing well = aggressive",
			pg: &model.PerformanceGoals{
				LearningMode: false,
				TargetCPA:    10.0,
			},
			perf: performanceData{
				impressions: 5000,
				cpa:         7.0, // Below 80% of target (8.0)
			},
			want: "aggressive",
		},
		{
			name: "Normal performance = moderate",
			pg: &model.PerformanceGoals{
				LearningMode: false,
				TargetCPA:    10.0,
			},
			perf: performanceData{
				impressions: 5000,
				cpa:         9.0, // Close to target
			},
			want: "moderate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.determineOptimizationLevel(tt.pg, tt.perf)
			if got != tt.want {
				t.Errorf("determineOptimizationLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ==================== Deal Tests ====================

func TestGetDealPriority(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name    string
		deal    *model.Deal
		dt      *model.DealTargeting
		wantMin int
		wantMax int
	}{
		{
			name:    "Nil deal",
			deal:    nil,
			dt:      &model.DealTargeting{},
			wantMin: 0,
			wantMax: 0,
		},
		{
			name: "Deal with priority override",
			deal: &model.Deal{ID: "deal-1"},
			dt: &model.DealTargeting{
				DealBidAdjustments: []model.DealBidAdjust{
					{DealID: "deal-1", Priority: 10},
				},
			},
			wantMin: 10,
			wantMax: 10,
		},
		{
			name: "PG deal (high priority)",
			deal: &model.Deal{
				ID: "pg-deal",
				At: 3, // Guaranteed price
			},
			dt:      &model.DealTargeting{},
			wantMin: 7, // 5 base + ~4 for PG
			wantMax: 10,
		},
		{
			name: "Private auction deal",
			deal: &model.Deal{
				ID: "pa-deal",
				At: 2, // 2nd price
			},
			dt:      &model.DealTargeting{},
			wantMin: 5, // Base + private auction bonus
			wantMax: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.getDealPriority(tt.deal, tt.dt)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("getDealPriority() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGetPublisherID(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name string
		req  *model.BidRequest
		want string
	}{
		{
			name: "Publisher ID from context",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"publisher_id": "pub-123",
				},
			},
			want: "pub-123",
		},
		{
			name: "Pub ID from context",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"pub_id": "pub-456",
				},
			},
			want: "pub-456",
		},
		{
			name: "Site publisher ID",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"site_publisher_id": "site-pub-789",
				},
			},
			want: "site-pub-789",
		},
		{
			name: "App publisher ID",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"app_publisher_id": "app-pub-101",
				},
			},
			want: "app-pub-101",
		},
		{
			name: "Empty context",
			req:  &model.BidRequest{},
			want: "",
		},
		{
			name: "Priority: publisher_id over others",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"publisher_id": "priority-pub",
					"pub_id":       "fallback-pub",
				},
			},
			want: "priority-pub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.getPublisherID(tt.req)
			if got != tt.want {
				t.Errorf("getPublisherID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ==================== Retargeting Tests ====================

func TestCheckRetargetingEligibility(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name     string
		campaign *model.Campaign
		userID   string
		want     bool
	}{
		{
			name: "Empty user ID",
			campaign: &model.Campaign{
				ID: "campaign-1",
				Targeting: model.Targeting{
					RetargetingEvents: []string{"click"},
				},
			},
			userID: "",
			want:   false,
		},
		{
			name: "User with no events (mock returns false)",
			campaign: &model.Campaign{
				ID: "campaign-1",
				Targeting: model.Targeting{
					RetargetingEvents:    []string{"click"},
					RetargetingCampaigns: []string{"campaign-1"},
				},
			},
			userID: "user-no-events",
			want:   false,
		},
		{
			name: "Default events (impression, click)",
			campaign: &model.Campaign{
				ID:        "campaign-2",
				Targeting: model.Targeting{},
			},
			userID: "user-456",
			want:   false, // Mock cache returns false by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.checkRetargetingEligibility(tt.campaign, tt.userID)
			if got != tt.want {
				t.Errorf("checkRetargetingEligibility() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ==================== Pacing Tests ====================

func TestCalculatePacingMultiplier(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name        string
		strategy    string
		spent       float64
		dailyBudget float64
		wantMin     float64
		wantMax     float64
	}{
		{
			name:        "ASAP strategy always returns 1.0",
			strategy:    "asap",
			spent:       500.0,
			dailyBudget: 1000.0,
			wantMin:     1.0,
			wantMax:     1.0,
		},
		{
			name:        "Empty strategy defaults to even",
			strategy:    "",
			spent:       500.0,
			dailyBudget: 1000.0,
			wantMin:     0.2, // Could vary by time of day
			wantMax:     1.3,
		},
		{
			name:        "Even pacing",
			strategy:    "even",
			spent:       500.0,
			dailyBudget: 1000.0,
			wantMin:     0.2, // Could vary by time of day
			wantMax:     1.3,
		},
		{
			name:        "Front-loaded pacing",
			strategy:    "front",
			spent:       500.0,
			dailyBudget: 1000.0,
			wantMin:     0.2,
			wantMax:     1.3,
		},
		{
			name:        "Back-loaded pacing",
			strategy:    "back",
			spent:       500.0,
			dailyBudget: 1000.0,
			wantMin:     0.2,
			wantMax:     1.3,
		},
		{
			name:        "Zero expected spend",
			strategy:    "even",
			spent:       0.0,
			dailyBudget: 0.0,
			wantMin:     1.0,
			wantMax:     1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.calculatePacingMultiplier(tt.strategy, tt.spent, tt.dailyBudget)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculatePacingMultiplier(%q, %.1f, %.1f) = %v, want between %v and %v",
					tt.strategy, tt.spent, tt.dailyBudget, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// ==================== Household/CTV Helpers ====================

func TestGetHouseholdImpressions(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name        string
		campaignID  string
		householdID string
		want        int
	}{
		{
			name:        "No cached data returns 0",
			campaignID:  "camp-123",
			householdID: "hh-456",
			want:        0,
		},
		{
			name:        "Empty campaign ID",
			campaignID:  "",
			householdID: "hh-123",
			want:        0,
		},
		{
			name:        "Empty household ID",
			campaignID:  "camp-123",
			householdID: "",
			want:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.getHouseholdImpressions(tt.campaignID, tt.householdID)
			if got != tt.want {
				t.Errorf("getHouseholdImpressions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAppName(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name string
		req  *model.BidRequest
		want string
	}{
		{
			name: "App name from context",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"app_name": "MyApp",
				},
			},
			want: "MyApp",
		},
		{
			name: "Bundle from context as fallback",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"bundle": "com.example.app",
				},
			},
			want: "com.example.app",
		},
		{
			name: "Empty context returns empty string",
			req:  &model.BidRequest{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.getAppName(tt.req)
			if got != tt.want {
				t.Errorf("getAppName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPlacement(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name string
		req  *model.BidRequest
		want string
	}{
		{
			name: "Placement from context",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"placement": "pre-roll",
				},
			},
			want: "pre-roll",
		},
		{
			name: "Ad type from context as fallback",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"ad_type": "mid-roll",
				},
			},
			want: "mid-roll",
		},
		{
			name: "Empty context returns empty string",
			req:  &model.BidRequest{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.getPlacement(tt.req)
			if got != tt.want {
				t.Errorf("getPlacement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsLowLTVSource(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name string
		req  *model.BidRequest
		want bool
	}{
		{
			name: "Low LTV from score",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"source_ltv_score": float64(0.2),
				},
			},
			want: true,
		},
		{
			name: "High LTV from score",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"source_ltv_score": float64(0.5),
				},
			},
			want: false,
		},
		{
			name: "Low LTV explicitly set",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"low_ltv_source": true,
				},
			},
			want: true,
		},
		{
			name: "Low LTV false",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"low_ltv_source": false,
				},
			},
			want: false,
		},
		{
			name: "Empty context returns false",
			req:  &model.BidRequest{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.isLowLTVSource(tt.req)
			if got != tt.want {
				t.Errorf("isLowLTVSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSKAdNetworkMultiplier(t *testing.T) {
	service := createBiddingUtilsService()

	tests := []struct {
		name    string
		req     *model.BidRequest
		wantMin float64
		wantMax float64
	}{
		{
			name: "SKAdNetwork supported",
			req: &model.BidRequest{
				Device: model.InternalDevice{
					OS: "ios",
				},
				Context: map[string]interface{}{
					"skadn_supported": true,
				},
			},
			wantMin: 1.09,
			wantMax: 1.11,
		},
		{
			name: "SKAdNetwork not supported",
			req: &model.BidRequest{
				Device: model.InternalDevice{
					OS: "ios",
				},
				Context: map[string]interface{}{
					"skadn_supported": false,
				},
			},
			wantMin: 0.99,
			wantMax: 1.01,
		},
		{
			name: "Non-iOS device",
			req: &model.BidRequest{
				Device: model.InternalDevice{
					OS: "android",
				},
			},
			wantMin: 0.99,
			wantMax: 1.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.getSKAdNetworkMultiplier(tt.req)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("getSKAdNetworkMultiplier() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}
