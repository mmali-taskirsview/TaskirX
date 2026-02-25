package service

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// RealTimeAlertService manages real-time alerting and anomaly detection
type RealTimeAlertService struct {
	cache           cache.Cache
	mu              sync.RWMutex
	activeAlerts    map[string]*model.Alert
	campaignMetrics map[string]*campaignMetricsHistory
	alertCooldowns  map[string]time.Time
}

type campaignMetricsHistory struct {
	hourlySpend     []float64
	hourlyCTR       []float64
	hourlyCVR       []float64
	hourlyWinRate   []float64
	baselineSpend   float64
	baselineCTR     float64
	baselineCVR     float64
	baselineWinRate float64
	lastUpdated     time.Time
}

// NewRealTimeAlertService creates a new real-time alert service
func NewRealTimeAlertService(c cache.Cache) *RealTimeAlertService {
	return &RealTimeAlertService{
		cache:           c,
		activeAlerts:    make(map[string]*model.Alert),
		campaignMetrics: make(map[string]*campaignMetricsHistory),
		alertCooldowns:  make(map[string]time.Time),
	}
}

// CheckAlerts evaluates alert conditions for a campaign during bidding
func (s *RealTimeAlertService) CheckAlerts(campaign *model.Campaign, currentSpend, dailyBudget float64) *model.AlertResult {
	config := campaign.Targeting.AlertConfig
	if config == nil || !config.Enabled {
		return &model.AlertResult{
			HasActiveAlerts: false,
			BidAdjustment:   1.0,
		}
	}

	result := &model.AlertResult{
		HasActiveAlerts: false,
		BidAdjustment:   1.0,
	}

	// Check budget alerts
	if config.BudgetAlerts != nil && config.BudgetAlerts.Enabled {
		s.checkBudgetAlerts(campaign, config.BudgetAlerts, currentSpend, dailyBudget, result)
	}

	// Check pacing alerts
	if config.PacingAlerts != nil && config.PacingAlerts.Enabled {
		s.checkPacingAlerts(campaign, config.PacingAlerts, currentSpend, dailyBudget, result)
	}

	// Get active alerts for this campaign
	s.mu.RLock()
	for _, alert := range s.activeAlerts {
		if alert.CampaignID == campaign.ID && !alert.Acknowledged {
			result.HasActiveAlerts = true
			result.ActiveAlertTypes = append(result.ActiveAlertTypes, alert.Type)
			if alert.Severity == "critical" {
				result.CriticalAlerts++
			}
		}
	}
	s.mu.RUnlock()

	// Adjust bid if critical alerts
	if result.CriticalAlerts > 0 {
		result.BidAdjustment = 0.5 // Reduce bidding on critical issues
		if result.CriticalAlerts >= 3 {
			result.ShouldPauseBid = true
			result.Reason = "multiple_critical_alerts"
		}
	}

	return result
}

func (s *RealTimeAlertService) checkBudgetAlerts(campaign *model.Campaign, config *model.BudgetAlerts, currentSpend, dailyBudget float64, result *model.AlertResult) {
	if dailyBudget <= 0 {
		return
	}

	spendPercent := (currentSpend / dailyBudget) * 100

	// Warning threshold
	warnAt := config.WarnAtPercent
	if warnAt <= 0 {
		warnAt = 80.0
	}

	// Critical threshold
	criticalAt := config.CriticalAtPercent
	if criticalAt <= 0 {
		criticalAt = 95.0
	}

	if spendPercent >= criticalAt {
		s.createAlert(campaign, "budget", "critical",
			fmt.Sprintf("Budget critical: %.1f%% spent", spendPercent),
			"spend_percent", spendPercent, criticalAt)
		result.BidAdjustment *= 0.3
	} else if spendPercent >= warnAt {
		s.createAlert(campaign, "budget", "warning",
			fmt.Sprintf("Budget warning: %.1f%% spent", spendPercent),
			"spend_percent", spendPercent, warnAt)
		result.BidAdjustment *= 0.7
	}

	// Check for unexpected spike
	if config.UnexpectedSpike {
		s.checkSpendSpike(campaign, config, currentSpend, result)
	}
}

func (s *RealTimeAlertService) checkSpendSpike(campaign *model.Campaign, config *model.BudgetAlerts, currentSpend float64, _ *model.AlertResult) {
	s.mu.RLock()
	history, exists := s.campaignMetrics[campaign.ID]
	s.mu.RUnlock()

	if !exists || history.baselineSpend <= 0 {
		return
	}

	spikeThreshold := config.SpikeThreshold
	if spikeThreshold <= 0 {
		spikeThreshold = 50.0 // 50% increase default
	}

	// Compare current hour spend to baseline
	hourlyAvg := history.baselineSpend / 24
	if hourlyAvg > 0 {
		currentHourSpend := currentSpend // This would be just the current hour in production
		spikePercent := ((currentHourSpend - hourlyAvg) / hourlyAvg) * 100

		if spikePercent >= spikeThreshold {
			s.createAlert(campaign, "budget", "warning",
				fmt.Sprintf("Spend spike detected: %.1f%% above normal", spikePercent),
				"spend_spike", spikePercent, spikeThreshold)
		}
	}
}

func (s *RealTimeAlertService) checkPacingAlerts(campaign *model.Campaign, config *model.PacingAlerts, currentSpend, dailyBudget float64, result *model.AlertResult) {
	if dailyBudget <= 0 {
		return
	}

	// Calculate expected spend at this point in day
	now := time.Now()
	hoursElapsed := float64(now.Hour()) + float64(now.Minute())/60.0
	expectedSpendPercent := (hoursElapsed / 24.0) * 100
	actualSpendPercent := (currentSpend / dailyBudget) * 100

	pacingDiff := actualSpendPercent - expectedSpendPercent

	underPacing := config.UnderPacingPercent
	if underPacing <= 0 {
		underPacing = 20.0
	}

	overPacing := config.OverPacingPercent
	if overPacing <= 0 {
		overPacing = 20.0
	}

	if pacingDiff < -underPacing {
		s.createAlert(campaign, "pacing", "warning",
			fmt.Sprintf("Under-pacing: %.1f%% below target", -pacingDiff),
			"pacing_diff", -pacingDiff, underPacing)
		result.BidAdjustment *= 1.2 // Increase bids to catch up
	} else if pacingDiff > overPacing {
		s.createAlert(campaign, "pacing", "warning",
			fmt.Sprintf("Over-pacing: %.1f%% above target", pacingDiff),
			"pacing_diff", pacingDiff, overPacing)
		result.BidAdjustment *= 0.8 // Decrease bids to slow down
	}
}

// CheckPerformanceAlerts checks for performance anomalies
func (s *RealTimeAlertService) CheckPerformanceAlerts(campaign *model.Campaign, ctr, cvr, winRate float64) {
	config := campaign.Targeting.AlertConfig
	if config == nil || !config.Enabled || config.PerformanceAlerts == nil || !config.PerformanceAlerts.Enabled {
		return
	}

	perfConfig := config.PerformanceAlerts

	s.mu.RLock()
	history, exists := s.campaignMetrics[campaign.ID]
	s.mu.RUnlock()

	if !exists || history.baselineCTR <= 0 {
		return
	}

	// Check CTR drop
	if perfConfig.CTRDropPercent > 0 && history.baselineCTR > 0 {
		ctrChange := ((ctr - history.baselineCTR) / history.baselineCTR) * 100
		if ctrChange < -perfConfig.CTRDropPercent {
			s.createAlert(campaign, "performance", "warning",
				fmt.Sprintf("CTR dropped %.1f%% from baseline", -ctrChange),
				"ctr", ctr, history.baselineCTR)
		}
	}

	// Check CVR drop
	if perfConfig.CVRDropPercent > 0 && history.baselineCVR > 0 {
		cvrChange := ((cvr - history.baselineCVR) / history.baselineCVR) * 100
		if cvrChange < -perfConfig.CVRDropPercent {
			s.createAlert(campaign, "performance", "warning",
				fmt.Sprintf("CVR dropped %.1f%% from baseline", -cvrChange),
				"cvr", cvr, history.baselineCVR)
		}
	}

	// Check win rate drop
	if perfConfig.WinRateDropPercent > 0 && history.baselineWinRate > 0 {
		wrChange := ((winRate - history.baselineWinRate) / history.baselineWinRate) * 100
		if wrChange < -perfConfig.WinRateDropPercent {
			s.createAlert(campaign, "performance", "warning",
				fmt.Sprintf("Win rate dropped %.1f%% from baseline", -wrChange),
				"win_rate", winRate, history.baselineWinRate)
		}
	}
}

// DetectAnomaly performs ML-based anomaly detection
func (s *RealTimeAlertService) DetectAnomaly(campaign *model.Campaign, metricName string, value float64) bool {
	config := campaign.Targeting.AlertConfig
	if config == nil || config.AnomalyDetection == nil || !config.AnomalyDetection.Enabled {
		return false
	}

	s.mu.RLock()
	history, exists := s.campaignMetrics[campaign.ID]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	// Get historical values for the metric
	var values []float64
	switch metricName {
	case "spend":
		values = history.hourlySpend
	case "ctr":
		values = history.hourlyCTR
	case "cvr":
		values = history.hourlyCVR
	case "win_rate":
		values = history.hourlyWinRate
	default:
		return false
	}

	if len(values) < 10 {
		return false
	}

	// Calculate mean and standard deviation
	mean, stdDev := s.calculateStats(values)
	if stdDev == 0 {
		return false
	}

	// Calculate z-score
	zScore := math.Abs((value - mean) / stdDev)

	// Determine threshold based on sensitivity
	threshold := 2.5 // Default: medium sensitivity
	switch config.AnomalyDetection.Sensitivity {
	case "low":
		threshold = 3.0
	case "high":
		threshold = 2.0
	}

	isAnomaly := zScore > threshold

	if isAnomaly {
		severity := "warning"
		if zScore > threshold*1.5 {
			severity = "critical"
		}

		s.createAlert(campaign, "anomaly", severity,
			fmt.Sprintf("Anomaly detected in %s: value=%.4f, z-score=%.2f", metricName, value, zScore),
			metricName, value, mean)

		// Auto-pause if configured and severe
		if config.AnomalyDetection.AutoPause && severity == "critical" {
			s.createAlert(campaign, "anomaly", "critical",
				"Campaign auto-paused due to severe anomaly",
				metricName, value, mean)
		}
	}

	return isAnomaly
}

func (s *RealTimeAlertService) calculateStats(values []float64) (mean, stdDev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	// Calculate standard deviation
	sumSq := 0.0
	for _, v := range values {
		sumSq += (v - mean) * (v - mean)
	}
	variance := sumSq / float64(len(values))
	stdDev = math.Sqrt(variance)

	return mean, stdDev
}

func (s *RealTimeAlertService) createAlert(campaign *model.Campaign, alertType, severity, message, metric string, currentValue, threshold float64) {
	// Check cooldown
	cooldownKey := fmt.Sprintf("%s_%s_%s", campaign.ID, alertType, metric)

	s.mu.RLock()
	lastAlert, exists := s.alertCooldowns[cooldownKey]
	s.mu.RUnlock()

	cooldownMinutes := 30 // Default cooldown
	if campaign.Targeting.AlertConfig != nil && campaign.Targeting.AlertConfig.AlertCooldown > 0 {
		cooldownMinutes = campaign.Targeting.AlertConfig.AlertCooldown
	}

	if exists && time.Since(lastAlert) < time.Duration(cooldownMinutes)*time.Minute {
		return // Still in cooldown
	}

	alert := &model.Alert{
		ID:           uuid.New().String(),
		Type:         alertType,
		Severity:     severity,
		CampaignID:   campaign.ID,
		CampaignName: campaign.Name,
		Message:      message,
		Metric:       metric,
		CurrentValue: currentValue,
		Threshold:    threshold,
		Timestamp:    time.Now(),
		Acknowledged: false,
	}

	s.mu.Lock()
	s.activeAlerts[alert.ID] = alert
	s.alertCooldowns[cooldownKey] = time.Now()
	s.mu.Unlock()
}

// RecordMetrics records campaign metrics for baseline calculation
func (s *RealTimeAlertService) RecordMetrics(campaignID string, spend, ctr, cvr, winRate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.campaignMetrics[campaignID]; !exists {
		s.campaignMetrics[campaignID] = &campaignMetricsHistory{
			hourlySpend:   make([]float64, 0, 168), // 7 days of hourly data
			hourlyCTR:     make([]float64, 0, 168),
			hourlyCVR:     make([]float64, 0, 168),
			hourlyWinRate: make([]float64, 0, 168),
		}
	}

	history := s.campaignMetrics[campaignID]

	// Append new values
	history.hourlySpend = append(history.hourlySpend, spend)
	history.hourlyCTR = append(history.hourlyCTR, ctr)
	history.hourlyCVR = append(history.hourlyCVR, cvr)
	history.hourlyWinRate = append(history.hourlyWinRate, winRate)

	// Keep only last 168 hours (7 days)
	maxLen := 168
	if len(history.hourlySpend) > maxLen {
		history.hourlySpend = history.hourlySpend[len(history.hourlySpend)-maxLen:]
		history.hourlyCTR = history.hourlyCTR[len(history.hourlyCTR)-maxLen:]
		history.hourlyCVR = history.hourlyCVR[len(history.hourlyCVR)-maxLen:]
		history.hourlyWinRate = history.hourlyWinRate[len(history.hourlyWinRate)-maxLen:]
	}

	// Update baselines
	history.baselineSpend, _ = s.calculateStats(history.hourlySpend)
	history.baselineCTR, _ = s.calculateStats(history.hourlyCTR)
	history.baselineCVR, _ = s.calculateStats(history.hourlyCVR)
	history.baselineWinRate, _ = s.calculateStats(history.hourlyWinRate)
	history.lastUpdated = time.Now()
}

// GetActiveAlerts returns all active alerts for a campaign
func (s *RealTimeAlertService) GetActiveAlerts(campaignID string) []*model.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var alerts []*model.Alert
	for _, alert := range s.activeAlerts {
		if alert.CampaignID == campaignID && !alert.Acknowledged {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// AcknowledgeAlert marks an alert as acknowledged
func (s *RealTimeAlertService) AcknowledgeAlert(alertID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if alert, exists := s.activeAlerts[alertID]; exists {
		alert.Acknowledged = true
		return true
	}

	return false
}

// ClearOldAlerts removes alerts older than retention period
func (s *RealTimeAlertService) ClearOldAlerts(retentionHours int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-time.Duration(retentionHours) * time.Hour)

	for id, alert := range s.activeAlerts {
		if alert.Timestamp.Before(cutoff) {
			delete(s.activeAlerts, id)
		}
	}
}
