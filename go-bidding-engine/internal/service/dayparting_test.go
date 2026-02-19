package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

func newDPCampaign(dp *model.DaypartingOptimization) *model.Campaign {
	return &model.Campaign{
		ID:       "camp-dp",
		BidPrice: 5.0,
		Budget:   1000,
		Status:   "active",
		Targeting: model.Targeting{
			PerformanceGoals: &model.PerformanceGoals{
				DaypartingOptimization: dp,
			},
		},
	}
}

func newDPRequest() *model.BidRequest {
	return &model.BidRequest{
		ID:     "req-dp",
		User:   model.InternalUser{Country: "US"},
		Device: model.InternalDevice{Type: "mobile"},
	}
}

func makeFixedTime(hour, weekday int) time.Time {
	return time.Date(2025, 1, 5+weekday, hour, 0, 0, 0, time.UTC)
}

func TestDayparting_Disabled(t *testing.T) {
	mc := NewMockCache()
	svc := NewDaypartingService(mc)
	req := newDPRequest()

	camp := &model.Campaign{ID: "camp1"}
	result := svc.CalculateDaypartMultiplier(camp, req)
	assertNear(t, "disabled_nil", result.Multiplier, 1.0, 0.001)

	camp2 := newDPCampaign(&model.DaypartingOptimization{Enabled: false})
	result2 := svc.CalculateDaypartMultiplier(camp2, req)
	assertNear(t, "disabled_false", result2.Multiplier, 1.0, 0.001)
}

func TestDayparting_ManualHourlyMultipliers(t *testing.T) {
	mc := NewMockCache()
	svc := NewDaypartingService(mc)
	req := newDPRequest()

	hourlyMults := make(map[int]float64)
	for h := 0; h < 24; h++ {
		hourlyMults[h] = 0.5 + float64(h)*0.05
	}

	camp := newDPCampaign(&model.DaypartingOptimization{
		Enabled:           true,
		HourlyMultipliers: hourlyMults,
	})

	result := svc.CalculateDaypartMultiplier(camp, req)
	if result.Multiplier < 0.3 || result.Multiplier > 2.0 {
		t.Errorf("multiplier out of range: %.2f", result.Multiplier)
	}
	if result.Reason == "" {
		t.Error("expected a reason string for manual hourly multiplier")
	}
}

func TestDayparting_AutoOptimize_InsufficientData(t *testing.T) {
	mc := NewMockCache()
	svc := NewDaypartingService(mc)
	req := newDPRequest()

	camp := newDPCampaign(&model.DaypartingOptimization{
		Enabled:      true,
		AutoOptimize: true,
	})

	result := svc.CalculateDaypartMultiplier(camp, req)
	assertNear(t, "insufficient_data", result.Multiplier, 1.0, 0.001)
	if result.Reason != "insufficient_data" {
		t.Errorf("expected reason insufficient_data, got %s", result.Reason)
	}
}

func TestDayparting_AutoOptimize_WithData(t *testing.T) {
	mc := NewMockCache()
	svc := NewDaypartingService(mc)
	req := newDPRequest()

	for h := 0; h < 24; h++ {
		key := fmt.Sprintf("daypart_perf:camp-dp:%d", h)
		mc.SetKV(key, "impressions:500,clicks:25,conversions:5,spend:10.0,ctr:0.050000,cvr:0.200000")
	}

	camp := newDPCampaign(&model.DaypartingOptimization{
		Enabled:      true,
		AutoOptimize: true,
		LookbackDays: 14,
	})

	result := svc.CalculateDaypartMultiplier(camp, req)
	if result.Multiplier < 0.3 || result.Multiplier > 2.0 {
		t.Errorf("auto-optimized multiplier out of bounds: %.2f", result.Multiplier)
	}
	if result.Reason == "" || result.Reason == "insufficient_data" {
		t.Errorf("expected auto-optimized reason, got %s", result.Reason)
	}
}

func TestDayparting_ClampMin(t *testing.T) {
	mc := NewMockCache()
	svc := NewDaypartingService(mc)
	req := newDPRequest()

	hourlyMults := make(map[int]float64)
	for h := 0; h < 24; h++ {
		hourlyMults[h] = 0.01
	}

	camp := newDPCampaign(&model.DaypartingOptimization{
		Enabled:           true,
		HourlyMultipliers: hourlyMults,
		MinMultiplier:     0.5,
		MaxMultiplier:     1.5,
	})

	result := svc.CalculateDaypartMultiplier(camp, req)
	assertNear(t, "clamped_min", result.Multiplier, 0.5, 0.001)
}

func TestDayparting_ClampMax(t *testing.T) {
	mc := NewMockCache()
	svc := NewDaypartingService(mc)
	req := newDPRequest()

	hourlyMults := make(map[int]float64)
	for h := 0; h < 24; h++ {
		hourlyMults[h] = 10.0
	}

	camp := newDPCampaign(&model.DaypartingOptimization{
		Enabled:           true,
		HourlyMultipliers: hourlyMults,
		MinMultiplier:     0.5,
		MaxMultiplier:     1.5,
	})

	result := svc.CalculateDaypartMultiplier(camp, req)
	assertNear(t, "clamped_max", result.Multiplier, 1.5, 0.001)
}

func TestDayparting_ClampDefaultBounds(t *testing.T) {
	mc := NewMockCache()
	svc := NewDaypartingService(mc)
	req := newDPRequest()

	hourlyMults := make(map[int]float64)
	for h := 0; h < 24; h++ {
		hourlyMults[h] = 0.1
	}

	camp := newDPCampaign(&model.DaypartingOptimization{
		Enabled:           true,
		HourlyMultipliers: hourlyMults,
	})

	result := svc.CalculateDaypartMultiplier(camp, req)
	assertNear(t, "clamped_default_min", result.Multiplier, 0.3, 0.001)
}

func TestDayparting_RecordHourlyPerf(t *testing.T) {
	mc := NewMockCache()
	svc := NewDaypartingService(mc)

	now := makeFixedTime(10, 3)

	err := svc.RecordHourlyPerformance("camp1", "impression", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hourKey := "daypart_perf:camp1:10"
	val, _ := mc.Get(hourKey)
	if val == "" {
		t.Error("expected hour performance data to be stored")
	}

	dayKey := "daypart_dow:camp1:3"
	dayVal, _ := mc.Get(dayKey)
	if dayVal == "" {
		t.Error("expected day performance data to be stored")
	}
}

func TestDayparting_RecordMultiEvents(t *testing.T) {
	mc := NewMockCache()
	svc := NewDaypartingService(mc)

	now := makeFixedTime(14, 1)

	_ = svc.RecordHourlyPerformance("camp1", "impression", now)
	_ = svc.RecordHourlyPerformance("camp1", "impression", now)
	_ = svc.RecordHourlyPerformance("camp1", "click", now)

	hourKey := "daypart_perf:camp1:14"
	val, _ := mc.Get(hourKey)
	if val == "" {
		t.Fatal("expected performance data")
	}

	perf := svc.parseDaypartCache(val)
	if perf.impressions != 2 {
		t.Errorf("expected 2 impressions, got %d", perf.impressions)
	}
	if perf.clicks != 1 {
		t.Errorf("expected 1 click, got %d", perf.clicks)
	}
}

func TestDayparting_OptimalHours_NoData(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	hours := svc.GetOptimalHours("camp1", 6)
	if len(hours) != 0 {
		t.Errorf("expected 0 optimal hours with no data, got %d", len(hours))
	}
}

func TestDayparting_OptimalHours_WithData(t *testing.T) {
	mc := NewMockCache()
	svc := NewDaypartingService(mc)

	for h := 0; h < 24; h++ {
		key := fmt.Sprintf("daypart_perf:camp1:%d", h)
		ctr := 0.02 + float64(h)*0.002
		cvr := 0.1 + float64(h)*0.01
		mc.SetKV(key, fmt.Sprintf("impressions:100,clicks:%d,conversions:%d,spend:5.0,ctr:%.6f,cvr:%.6f",
			int(100*ctr), int(100*ctr*cvr), ctr, cvr))
	}

	hours := svc.GetOptimalHours("camp1", 3)
	if len(hours) > 3 {
		t.Errorf("expected at most 3 optimal hours, got %d", len(hours))
	}

	for i := 1; i < len(hours); i++ {
		if hours[i].Multiplier > hours[i-1].Multiplier {
			t.Error("optimal hours should be sorted descending by multiplier")
		}
	}
}

func TestDayparting_ParseCacheEmpty(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	perf := svc.parseDaypartCache("")
	if perf.impressions != 0 || perf.clicks != 0 || perf.ctr != 0 {
		t.Error("empty cache should produce zero performance")
	}
}

func TestDayparting_ParseCacheValid(t *testing.T) {
	svc := NewDaypartingService(NewMockCache())
	perf := svc.parseDaypartCache("impressions:1000,clicks:50,conversions:10,spend:25.5000,ctr:0.050000,cvr:0.200000")
	if perf.impressions != 1000 {
		t.Errorf("expected 1000 impressions, got %d", perf.impressions)
	}
	if perf.clicks != 50 {
		t.Errorf("expected 50 clicks, got %d", perf.clicks)
	}
	assertNear(t, "spend", perf.spend, 25.5, 0.01)
	assertNear(t, "ctr", perf.ctr, 0.05, 0.001)
	assertNear(t, "cvr", perf.cvr, 0.2, 0.001)
}

func TestDayparting_ManualPriority(t *testing.T) {
	mc := NewMockCache()
	svc := NewDaypartingService(mc)
	req := newDPRequest()

	hourlyMults := make(map[int]float64)
	for h := 0; h < 24; h++ {
		hourlyMults[h] = 1.8
	}

	camp := newDPCampaign(&model.DaypartingOptimization{
		Enabled:           true,
		HourlyMultipliers: hourlyMults,
		AutoOptimize:      true,
	})

	result := svc.CalculateDaypartMultiplier(camp, req)
	assertNear(t, "manual_priority", result.Multiplier, 1.8, 0.001)
}
