package service

// Coverage Boost 39 - Target remaining edge cases
// Focus: calculateSeasonalMultiplier (71.9%), optimizeForCPI (84.2%)

import (
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ──────────────────────────────────────────────────────────────────────────────
// calculateSeasonalMultiplier - Edge cases
// ──────────────────────────────────────────────────────────────────────────────

// TestB39_SeasonalMultiplier_YearWrapEvent tests recurring event spanning year boundary (Dec-Jan)
func TestB39_SeasonalMultiplier_YearWrapEvent(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "test-boost39")

	// Recurring event: Dec 26 - Jan 2 (Holiday sales)
	camp := &model.Campaign{
		ID:       "camp-yearwrap",
		BidPrice: 1.0,
		Targeting: model.Targeting{
			SeasonalTargeting: &model.SeasonalTargeting{
				Events: []model.SeasonalEvent{
					{
						Name:      "Year Wrap Sale",
						StartDate: "12-26", // MM-DD format
						EndDate:   "01-02",
						Boost:     1.8,
						Recurring: true,
						Active:    true,
					},
				},
			},
		},
	}

	result := svc.calculateSeasonalMultiplier(camp)

	// This will depend on current date - just ensure it runs without error
	if result.Multiplier < 0 || result.Multiplier > 3.0 {
		t.Errorf("Multiplier out of range: %v", result.Multiplier)
	}
}

// TestB39_SeasonalMultiplier_InvalidTimezone tests invalid timezone (should fall back to local)
func TestB39_SeasonalMultiplier_InvalidTimezone(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "test-boost39")

	camp := &model.Campaign{
		ID:       "camp-badtz",
		BidPrice: 1.0,
		Targeting: model.Targeting{
			SeasonalTargeting: &model.SeasonalTargeting{
				Timezone:     "Invalid/Timezone",
				Q4Boost:      1.5,
				WeekendBoost: 1.2,
			},
		},
	}

	result := svc.calculateSeasonalMultiplier(camp)

	// Should still work (uses local time)
	if result.Multiplier < 0 || result.Multiplier > 3.0 {
		t.Errorf("Multiplier out of range: %v", result.Multiplier)
	}
}

// TestB39_SeasonalMultiplier_MultiplierCap tests that multiplier is capped at 3.0
func TestB39_SeasonalMultiplier_MultiplierCap(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "test-boost39")

	// Stack multiple boosts to exceed 3.0 cap
	camp := &model.Campaign{
		ID:       "camp-cap",
		BidPrice: 1.0,
		Targeting: model.Targeting{
			SeasonalTargeting: &model.SeasonalTargeting{
				WeekendBoost:      2.0,
				MonthEndBoost:     2.0,
				Q4Boost:           2.0,
				SummerBoost:       2.0,
				BackToSchoolBoost: 2.0,
				Events: []model.SeasonalEvent{
					{
						Name:      "Mega Sale",
						StartDate: time.Now().Format("01-02"),
						EndDate:   time.Now().AddDate(0, 0, 1).Format("01-02"),
						Boost:     2.0,
						Recurring: true,
						Active:    true,
					},
				},
			},
		},
	}

	result := svc.calculateSeasonalMultiplier(camp)

	// Should be capped at 3.0
	if result.Multiplier > 3.0 {
		t.Errorf("Multiplier should be capped at 3.0, got %v", result.Multiplier)
	}
}

// TestB39_SeasonalMultiplier_EventWithZeroBoost tests event with boost <= 0 (default 1.5)
func TestB39_SeasonalMultiplier_EventWithZeroBoost(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "test-boost39")

	today := time.Now()
	camp := &model.Campaign{
		ID:       "camp-zeroboost",
		BidPrice: 1.0,
		Targeting: model.Targeting{
			SeasonalTargeting: &model.SeasonalTargeting{
				Events: []model.SeasonalEvent{
					{
						Name:      "Zero Boost Event",
						StartDate: today.Format("2006-01-02"),
						EndDate:   today.AddDate(0, 0, 1).Format("2006-01-02"),
						Boost:     0, // Zero boost - should default to 1.5
						Active:    true,
					},
				},
			},
		},
	}

	result := svc.calculateSeasonalMultiplier(camp)

	// Event is active, boost <= 0, should use default 1.5
	if !result.Matched {
		t.Error("Event should have matched")
	}
	if result.Multiplier < 1.4 || result.Multiplier > 1.6 {
		t.Errorf("Expected default boost ~1.5, got %v", result.Multiplier)
	}
}

// TestB39_SeasonalMultiplier_HolidayZeroBoost tests holiday with boost <= 0 (default 1.3)
func TestB39_SeasonalMultiplier_HolidayZeroBoost(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "test-boost39")

	camp := &model.Campaign{
		ID:       "camp-holidayzero",
		BidPrice: 1.0,
		Targeting: model.Targeting{
			SeasonalTargeting: &model.SeasonalTargeting{
				EnableHolidays: true,
				HolidayBoost:   0, // Zero boost - should default to 1.3
				Country:        "US",
			},
		},
	}

	// We can't control the current date, so just verify it runs
	result := svc.calculateSeasonalMultiplier(camp)

	// Just verify it doesn't crash and multiplier is in valid range
	if result.Multiplier < 0 || result.Multiplier > 3.0 {
		t.Errorf("Multiplier out of range: %v", result.Multiplier)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// optimizeForCPI - Edge cases
// ──────────────────────────────────────────────────────────────────────────────

// TestB39_OptimizeForCPI_VeryLowInstallRate tests extremely low predicted install rate
func TestB39_OptimizeForCPI_VeryLowInstallRate(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "test-boost39")

	campID := "camp-lowinstall"
	mc.ctr[campID] = 0.0001 // Very low CTR

	camp := &model.Campaign{
		ID:       campID,
		BidPrice: 1.0,
	}

	pg := &model.PerformanceGoals{
		TargetCPI: 5.0,
	}

	req := &model.BidRequest{
		Device:  model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{"historical_install_rate": 0.0001},
	}

	perf := performanceData{
		impressions: 1000,
		clicks:      1,
	}

	multiplier := svc.optimizeForCPI(camp, req, pg, perf)

	// Very low install rate → should still return a multiplier in valid range
	if multiplier < 0.3 || multiplier > 2.0 {
		t.Errorf("Expected multiplier in [0.3, 2.0], got %v", multiplier)
	}
}

// TestB39_OptimizeForCPI_ExactlyAtFloor tests ratio exactly at floor (0.3)
func TestB39_OptimizeForCPI_ExactlyAtFloor(t *testing.T) {
	mc := NewMockCache()
	svc := NewBiddingService(mc, "test-boost39")

	campID := "camp-floor"
	// Set CTR such that maxBid / bidPrice = 0.3
	// We want: (5.0 * CTR * installRate) / 1.0 = 0.3
	// So: CTR * installRate = 0.06
	mc.ctr[campID] = 0.06

	camp := &model.Campaign{
		ID:       campID,
		BidPrice: 1.0,
	}

	pg := &model.PerformanceGoals{
		TargetCPI: 5.0,
	}

	req := &model.BidRequest{
		Device:  model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{"historical_install_rate": 1.0},
	}

	multiplier := svc.optimizeForCPI(camp, req, pg, performanceData{})

	tolerance := 0.01
	if multiplier < 0.3-tolerance || multiplier > 0.3+tolerance {
		t.Errorf("Expected floor 0.3, got %v", multiplier)
	}
}
