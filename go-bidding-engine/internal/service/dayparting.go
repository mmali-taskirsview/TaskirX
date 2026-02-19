package service

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// DaypartingService provides automatic bid adjustments based on hour-of-day performance.
// Tracks historical performance by hour and day-of-week to optimize bid multipliers.
type DaypartingService struct {
	cache cache.Cache
}

// NewDaypartingService creates a new dayparting optimization service
func NewDaypartingService(cache cache.Cache) *DaypartingService {
	return &DaypartingService{cache: cache}
}

// dayNames maps day-of-week integers to names
var dayNames = map[int]string{
	0: "sunday",
	1: "monday",
	2: "tuesday",
	3: "wednesday",
	4: "thursday",
	5: "friday",
	6: "saturday",
}

// CalculateDaypartMultiplier returns a bid multiplier for the current hour based on
// campaign dayparting configuration and historical performance data.
func (d *DaypartingService) CalculateDaypartMultiplier(campaign *model.Campaign, req *model.BidRequest) model.DaypartingResult {
	result := model.DaypartingResult{
		Multiplier: 1.0,
	}

	pg := campaign.Targeting.PerformanceGoals
	if pg == nil || pg.DaypartingOptimization == nil || !pg.DaypartingOptimization.Enabled {
		return result
	}

	dpConfig := pg.DaypartingOptimization

	// Determine current time in configured timezone
	now := time.Now()
	if dpConfig.Timezone != "" {
		if loc, err := time.LoadLocation(dpConfig.Timezone); err == nil {
			now = now.In(loc)
		}
	}

	// Also check request timezone
	if req.Context != nil {
		if tz, ok := req.Context["timezone"].(string); ok && dpConfig.Timezone == "" {
			if loc, err := time.LoadLocation(tz); err == nil {
				now = now.In(loc)
			}
		}
	}

	hour := now.Hour()
	dayOfWeek := int(now.Weekday()) // 0=Sunday
	dayName := dayNames[dayOfWeek]

	result.Hour = hour
	result.DayOfWeek = dayName

	// 1. Check manual hourly multipliers first (they take precedence)
	if len(dpConfig.HourlyMultipliers) > 0 {
		if mult, ok := dpConfig.HourlyMultipliers[hour]; ok {
			result.Multiplier = mult
			result.Reason = fmt.Sprintf("manual_hourly_multiplier_h%d", hour)
			return d.clampMultiplier(result, dpConfig)
		}
	}

	// 2. Check day-specific multipliers
	if dpConfig.DaySpecific != nil {
		if dayMults, ok := dpConfig.DaySpecific[dayName]; ok {
			if mult, ok := dayMults[hour]; ok {
				result.Multiplier = mult
				result.Reason = fmt.Sprintf("day_specific_%s_h%d", dayName, hour)
				return d.clampMultiplier(result, dpConfig)
			}
		}
	}

	// 3. Auto-optimize using historical performance data
	if dpConfig.AutoOptimize {
		return d.autoOptimize(campaign.ID, hour, dayOfWeek, dpConfig, result)
	}

	return result
}

// autoOptimize uses historical hourly performance data to determine optimal bid multiplier
func (d *DaypartingService) autoOptimize(campaignID string, hour, dayOfWeek int, dpConfig *model.DaypartingOptimization, result model.DaypartingResult) model.DaypartingResult {
	lookbackDays := dpConfig.LookbackDays
	if lookbackDays <= 0 {
		lookbackDays = 14 // Default 2 weeks of data
	}

	// Get performance data for this hour across recent days
	hourPerf := d.getHourlyPerformance(campaignID, hour, lookbackDays)

	// Get overall average performance for normalization
	avgPerf := d.getAveragePerformance(campaignID, lookbackDays)

	if hourPerf.impressions < 100 || avgPerf.impressions < 100 {
		// Not enough data for auto-optimization
		result.Reason = "insufficient_data"
		return result
	}

	result.HistoricalCTR = hourPerf.ctr
	result.HistoricalCVR = hourPerf.cvr

	// Calculate composite performance score relative to average
	// Weighted by: CTR (30%), CVR (40%), Win Rate (30%)
	ctrRatio := 1.0
	if avgPerf.ctr > 0 {
		ctrRatio = hourPerf.ctr / avgPerf.ctr
	}

	cvrRatio := 1.0
	if avgPerf.cvr > 0 {
		cvrRatio = hourPerf.cvr / avgPerf.cvr
	}

	winRateRatio := 1.0
	if avgPerf.winRate > 0 {
		winRateRatio = hourPerf.winRate / avgPerf.winRate
	}

	// Composite performance index
	perfIndex := ctrRatio*0.3 + cvrRatio*0.4 + winRateRatio*0.3

	// Convert performance index to multiplier
	// perfIndex of 1.0 = average performance = multiplier 1.0
	// Higher perfIndex = better performance = bid more
	multiplier := 0.5 + perfIndex*0.5 // Maps [0, 2] -> [0.5, 1.5]

	// Apply day-of-week modifier
	dayMultiplier := d.getDayOfWeekModifier(campaignID, dayOfWeek, lookbackDays, avgPerf)
	multiplier *= dayMultiplier

	result.Multiplier = multiplier
	result.IsOptimalHour = perfIndex > 1.3 // Top 30% performance considered "optimal"
	result.Reason = fmt.Sprintf("auto_optimized_perf_index=%.2f_day_mod=%.2f", perfIndex, dayMultiplier)

	return d.clampMultiplier(result, dpConfig)
}

// hourlyPerformance holds aggregated performance for a specific hour
type hourlyPerformance struct {
	impressions int64
	clicks      int64
	conversions int64
	spend       float64
	ctr         float64
	cvr         float64
	winRate     float64
}

// getHourlyPerformance retrieves performance metrics for a specific hour
func (d *DaypartingService) getHourlyPerformance(campaignID string, hour, lookbackDays int) hourlyPerformance {
	perf := hourlyPerformance{}

	// Cache key format: daypart_perf:{campaignID}:{hour}
	cacheKey := fmt.Sprintf("daypart_perf:%s:%d", campaignID, hour)
	cached, err := d.cache.Get(cacheKey)
	if err != nil || cached == "" {
		return perf
	}

	// Parse cached data (format: "impressions:X,clicks:Y,conversions:Z,spend:W,wins:V")
	perf = d.parseDaypartCache(cached)
	return perf
}

// getAveragePerformance retrieves average performance across all hours
func (d *DaypartingService) getAveragePerformance(campaignID string, lookbackDays int) hourlyPerformance {
	perf := hourlyPerformance{}

	cacheKey := fmt.Sprintf("daypart_avg:%s", campaignID)
	cached, err := d.cache.Get(cacheKey)
	if err != nil || cached == "" {
		// Build average from all hours
		totalImps := int64(0)
		totalClicks := int64(0)
		totalConversions := int64(0)
		totalSpend := 0.0
		hourCount := 0

		for h := 0; h < 24; h++ {
			hp := d.getHourlyPerformance(campaignID, h, lookbackDays)
			if hp.impressions > 0 {
				totalImps += hp.impressions
				totalClicks += hp.clicks
				totalConversions += hp.conversions
				totalSpend += hp.spend
				hourCount++
			}
		}

		if hourCount > 0 {
			perf.impressions = totalImps / int64(hourCount)
			perf.clicks = totalClicks / int64(hourCount)
			perf.conversions = totalConversions / int64(hourCount)
			perf.spend = totalSpend / float64(hourCount)
			if perf.impressions > 0 {
				perf.ctr = float64(perf.clicks) / float64(perf.impressions)
			}
			if perf.clicks > 0 {
				perf.cvr = float64(perf.conversions) / float64(perf.clicks)
			}
		}

		return perf
	}

	perf = d.parseDaypartCache(cached)
	return perf
}

// getDayOfWeekModifier returns a bid modifier for the current day of week
func (d *DaypartingService) getDayOfWeekModifier(campaignID string, dayOfWeek, lookbackDays int, avgPerf hourlyPerformance) float64 {
	dayKey := fmt.Sprintf("daypart_dow:%s:%d", campaignID, dayOfWeek)
	cached, err := d.cache.Get(dayKey)
	if err != nil || cached == "" {
		return 1.0
	}

	dayPerf := d.parseDaypartCache(cached)
	if dayPerf.impressions < 50 || avgPerf.ctr <= 0 {
		return 1.0
	}

	// Compare day CTR to average
	if dayPerf.ctr > 0 && avgPerf.ctr > 0 {
		ratio := dayPerf.ctr / avgPerf.ctr
		// Gentle day-of-week adjustment (±20%)
		if ratio > 1.2 {
			return 1.2
		}
		if ratio < 0.8 {
			return 0.8
		}
		return ratio
	}

	return 1.0
}

// RecordHourlyPerformance updates performance tracking for the current hour.
// Should be called on impression, click, and conversion events.
func (d *DaypartingService) RecordHourlyPerformance(campaignID string, eventType string, now time.Time) error {
	hour := now.Hour()
	dayOfWeek := int(now.Weekday())

	// Update hour-specific metrics
	hourKey := fmt.Sprintf("daypart_perf:%s:%d", campaignID, hour)
	dayKey := fmt.Sprintf("daypart_dow:%s:%d", campaignID, dayOfWeek)

	// Get current values and increment
	current, _ := d.cache.Get(hourKey)
	updated := d.incrementDaypartMetric(current, eventType)
	_ = d.cache.Set(hourKey, updated, 86400*14) // 14-day TTL

	// Also update day-of-week metrics
	currentDay, _ := d.cache.Get(dayKey)
	updatedDay := d.incrementDaypartMetric(currentDay, eventType)
	_ = d.cache.Set(dayKey, updatedDay, 86400*14)

	return nil
}

// incrementDaypartMetric increments the appropriate counter in a serialized metrics string
func (d *DaypartingService) incrementDaypartMetric(current, eventType string) string {
	perf := d.parseDaypartCache(current)

	switch eventType {
	case "impression":
		perf.impressions++
	case "click":
		perf.clicks++
	case "conversion":
		perf.conversions++
	case "win":
		// Track via win rate
	}

	// Recalculate derived metrics
	if perf.impressions > 0 {
		perf.ctr = float64(perf.clicks) / float64(perf.impressions)
	}
	if perf.clicks > 0 {
		perf.cvr = float64(perf.conversions) / float64(perf.clicks)
	}

	return fmt.Sprintf("impressions:%d,clicks:%d,conversions:%d,spend:%.4f,ctr:%.6f,cvr:%.6f",
		perf.impressions, perf.clicks, perf.conversions, perf.spend, perf.ctr, perf.cvr)
}

// parseDaypartCache parses a serialized performance string
func (d *DaypartingService) parseDaypartCache(cached string) hourlyPerformance {
	perf := hourlyPerformance{}
	if cached == "" {
		return perf
	}

	pairs := strings.Split(cached, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "impressions":
			if v, err := strconv.ParseInt(val, 10, 64); err == nil {
				perf.impressions = v
			}
		case "clicks":
			if v, err := strconv.ParseInt(val, 10, 64); err == nil {
				perf.clicks = v
			}
		case "conversions":
			if v, err := strconv.ParseInt(val, 10, 64); err == nil {
				perf.conversions = v
			}
		case "spend":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				perf.spend = v
			}
		case "ctr":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				perf.ctr = v
			}
		case "cvr":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				perf.cvr = v
			}
		case "win_rate":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				perf.winRate = v
			}
		}
	}

	return perf
}

// clampMultiplier ensures the multiplier stays within configured bounds
func (d *DaypartingService) clampMultiplier(result model.DaypartingResult, dpConfig *model.DaypartingOptimization) model.DaypartingResult {
	minMult := dpConfig.MinMultiplier
	if minMult <= 0 {
		minMult = 0.3
	}
	maxMult := dpConfig.MaxMultiplier
	if maxMult <= 0 {
		maxMult = 2.0
	}

	if result.Multiplier < minMult {
		result.Multiplier = minMult
	}
	if result.Multiplier > maxMult {
		result.Multiplier = maxMult
	}

	return result
}

// GetOptimalHours returns the top N hours by performance for a campaign, useful for reporting.
func (d *DaypartingService) GetOptimalHours(campaignID string, topN int) []model.DaypartingResult {
	if topN <= 0 {
		topN = 6
	}

	lookbackDays := 14
	avgPerf := d.getAveragePerformance(campaignID, lookbackDays)

	results := make([]model.DaypartingResult, 0, 24)

	for h := 0; h < 24; h++ {
		hp := d.getHourlyPerformance(campaignID, h, lookbackDays)
		if hp.impressions < 10 {
			continue
		}

		perfIndex := 0.0
		if avgPerf.ctr > 0 {
			perfIndex += (hp.ctr / avgPerf.ctr) * 0.4
		}
		if avgPerf.cvr > 0 {
			perfIndex += (hp.cvr / avgPerf.cvr) * 0.6
		}

		results = append(results, model.DaypartingResult{
			Hour:          h,
			Multiplier:    0.5 + perfIndex*0.5,
			HistoricalCTR: hp.ctr,
			HistoricalCVR: hp.cvr,
			IsOptimalHour: perfIndex > 1.3,
		})
	}

	// Sort by multiplier descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Multiplier > results[j].Multiplier
	})

	if len(results) > topN {
		results = results[:topN]
	}

	return results
}

// Ensure math import is used
var _ = math.Abs
