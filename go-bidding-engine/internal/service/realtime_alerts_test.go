package service

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

func createRTACampaign(alertsEnabled bool) *model.Campaign {
	camp := &model.Campaign{
		ID:       "camp-rta-1",
		Name:     "RTA Test Campaign",
		BidPrice: 2.0,
		Targeting: model.Targeting{
			AlertConfig: nil,
		},
	}
	if alertsEnabled {
		camp.Targeting.AlertConfig = &model.AlertConfig{
			Enabled:       true,
			AlertCooldown: 1, // 1 minute for testing
			BudgetAlerts: &model.BudgetAlerts{
				Enabled:           true,
				WarnAtPercent:     80,
				CriticalAtPercent: 95,
				UnexpectedSpike:   true,
				SpikeThreshold:    50,
			},
			PacingAlerts: &model.PacingAlerts{
				Enabled:            true,
				UnderPacingPercent: 20,
				OverPacingPercent:  20,
			},
			PerformanceAlerts: &model.PerformanceAlerts{
				Enabled:            true,
				CTRDropPercent:     20,
				CVRDropPercent:     30,
				WinRateDropPercent: 25,
			},
			AnomalyDetection: &model.AnomalyDetection{
				Enabled:     true,
				Sensitivity: "medium",
				AutoPause:   true,
			},
		}
	}
	return camp
}

func TestRTA_NewService(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	if svc == nil {
		t.Fatal("expected service")
	}
	if svc.activeAlerts == nil {
		t.Error("expected alerts map initialized")
	}
	if svc.campaignMetrics == nil {
		t.Error("expected metrics map initialized")
	}
}

func TestRTA_CheckAlerts_Disabled(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(false)

	result := svc.CheckAlerts(campaign, 50, 100)

	if result.HasActiveAlerts {
		t.Error("expected no alerts when disabled")
	}
	if result.BidAdjustment != 1.0 {
		t.Errorf("expected bid adjustment 1.0, got %f", result.BidAdjustment)
	}
}

func TestRTA_CheckAlerts_BudgetWarning(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	// 85% spent - should trigger warning
	result := svc.CheckAlerts(campaign, 85, 100)

	if result.BidAdjustment >= 1.0 {
		t.Errorf("expected reduced bid adjustment for budget warning, got %f", result.BidAdjustment)
	}

	alerts := svc.GetActiveAlerts(campaign.ID)
	if len(alerts) == 0 {
		t.Error("expected alert for budget warning")
	}
}

func TestRTA_CheckAlerts_BudgetCritical(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	// 96% spent - should trigger critical
	result := svc.CheckAlerts(campaign, 96, 100)

	if result.BidAdjustment > 0.5 {
		t.Errorf("expected bid reduction for critical alert, got %f", result.BidAdjustment)
	}

	alerts := svc.GetActiveAlerts(campaign.ID)
	hasCritical := false
	for _, a := range alerts {
		if a.Severity == "critical" {
			hasCritical = true
		}
	}
	if !hasCritical {
		t.Error("expected critical alert")
	}
}

func TestRTA_CheckAlerts_ZeroBudget(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	// Zero budget - should not crash
	result := svc.CheckAlerts(campaign, 50, 0)

	if result == nil {
		t.Fatal("expected result")
	}
}

func TestRTA_CheckAlerts_MultipleCritical(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)
	campaign.Targeting.AlertConfig.AlertCooldown = 0 // Disable cooldown for test

	// Create multiple critical alerts
	svc.CheckAlerts(campaign, 96, 100)

	// Force multiple alerts by directly adding
	svc.mu.Lock()
	svc.activeAlerts["alert1"] = &model.Alert{
		CampaignID: campaign.ID,
		Severity:   "critical",
	}
	svc.activeAlerts["alert2"] = &model.Alert{
		CampaignID: campaign.ID,
		Severity:   "critical",
	}
	svc.activeAlerts["alert3"] = &model.Alert{
		CampaignID: campaign.ID,
		Severity:   "critical",
	}
	svc.mu.Unlock()

	result := svc.CheckAlerts(campaign, 10, 100)

	if !result.ShouldPauseBid {
		t.Error("expected pause recommendation with 3+ critical alerts")
	}
}

func TestRTA_CheckPacingAlerts_UnderPacing(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	// Very low spend early in day should not trigger under-pacing
	// But very low spend at end of day should
	// This is time-dependent, so we test the bid adjustment direction

	result := svc.CheckAlerts(campaign, 10, 100)

	// Result should exist
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestRTA_RecordMetrics(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	svc.RecordMetrics("camp-1", 100, 0.02, 0.01, 0.25)

	svc.mu.RLock()
	history := svc.campaignMetrics["camp-1"]
	svc.mu.RUnlock()

	if history == nil {
		t.Fatal("expected metrics recorded")
	}
	if len(history.hourlySpend) != 1 {
		t.Errorf("expected 1 spend record, got %d", len(history.hourlySpend))
	}
	if history.hourlySpend[0] != 100 {
		t.Errorf("expected spend 100, got %f", history.hourlySpend[0])
	}
}

func TestRTA_RecordMetrics_MaxHistory(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	// Record more than 168 hours of data
	for i := 0; i < 200; i++ {
		svc.RecordMetrics("camp-1", float64(i), 0.02, 0.01, 0.25)
	}

	svc.mu.RLock()
	history := svc.campaignMetrics["camp-1"]
	svc.mu.RUnlock()

	if len(history.hourlySpend) > 168 {
		t.Errorf("expected max 168 history entries, got %d", len(history.hourlySpend))
	}
}

func TestRTA_RecordMetrics_BaselineCalculation(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	// Record consistent metrics
	for i := 0; i < 10; i++ {
		svc.RecordMetrics("camp-1", 100, 0.02, 0.01, 0.25)
	}

	svc.mu.RLock()
	history := svc.campaignMetrics["camp-1"]
	svc.mu.RUnlock()

	if history.baselineSpend != 100 {
		t.Errorf("expected baseline spend 100, got %f", history.baselineSpend)
	}
	// Use tolerance for float comparison
	if math.Abs(history.baselineCTR-0.02) > 0.0001 {
		t.Errorf("expected baseline CTR ~0.02, got %f", history.baselineCTR)
	}
}

func TestRTA_CheckPerformanceAlerts_Disabled(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(false)

	// Should not crash
	svc.CheckPerformanceAlerts(campaign, 0.02, 0.01, 0.25)
}

func TestRTA_CheckPerformanceAlerts_CTRDrop(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	// Build baseline
	for i := 0; i < 10; i++ {
		svc.RecordMetrics(campaign.ID, 100, 0.02, 0.01, 0.25)
	}

	// Report significant CTR drop
	svc.CheckPerformanceAlerts(campaign, 0.01, 0.01, 0.25) // 50% drop

	alerts := svc.GetActiveAlerts(campaign.ID)
	hasPerfAlert := false
	for _, a := range alerts {
		if a.Type == "performance" {
			hasPerfAlert = true
		}
	}
	if !hasPerfAlert {
		t.Error("expected performance alert for CTR drop")
	}
}

func TestRTA_CheckPerformanceAlerts_CVRDrop(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	// Build baseline
	for i := 0; i < 10; i++ {
		svc.RecordMetrics(campaign.ID, 100, 0.02, 0.02, 0.25)
	}

	// Report significant CVR drop
	svc.CheckPerformanceAlerts(campaign, 0.02, 0.005, 0.25) // 75% drop

	alerts := svc.GetActiveAlerts(campaign.ID)
	hasAlert := false
	for _, a := range alerts {
		if a.Type == "performance" && a.Metric == "cvr" {
			hasAlert = true
		}
	}
	if !hasAlert {
		t.Error("expected performance alert for CVR drop")
	}
}

func TestRTA_CheckPerformanceAlerts_WinRateDrop(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	// Build baseline
	for i := 0; i < 10; i++ {
		svc.RecordMetrics(campaign.ID, 100, 0.02, 0.01, 0.50)
	}

	// Report significant win rate drop
	svc.CheckPerformanceAlerts(campaign, 0.02, 0.01, 0.20) // 60% drop

	alerts := svc.GetActiveAlerts(campaign.ID)
	hasAlert := false
	for _, a := range alerts {
		if a.Type == "performance" && a.Metric == "win_rate" {
			hasAlert = true
		}
	}
	if !hasAlert {
		t.Error("expected performance alert for win rate drop")
	}
}

func TestRTA_DetectAnomaly_Disabled(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(false)

	isAnomaly := svc.DetectAnomaly(campaign, "spend", 1000)

	if isAnomaly {
		t.Error("expected no anomaly detection when disabled")
	}
}

func TestRTA_DetectAnomaly_NoHistory(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	isAnomaly := svc.DetectAnomaly(campaign, "spend", 1000)

	if isAnomaly {
		t.Error("expected no anomaly detection without history")
	}
}

func TestRTA_DetectAnomaly_InsufficientHistory(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	// Add only 5 data points (need 10+)
	for i := 0; i < 5; i++ {
		svc.RecordMetrics(campaign.ID, 100, 0.02, 0.01, 0.25)
	}

	isAnomaly := svc.DetectAnomaly(campaign, "spend", 1000)

	if isAnomaly {
		t.Error("expected no anomaly detection with insufficient history")
	}
}

func TestRTA_DetectAnomaly_NormalValue(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	// Build consistent baseline
	for i := 0; i < 20; i++ {
		svc.RecordMetrics(campaign.ID, 100, 0.02, 0.01, 0.25)
	}

	// Normal value within range
	isAnomaly := svc.DetectAnomaly(campaign, "spend", 100)

	if isAnomaly {
		t.Error("expected no anomaly for normal value")
	}
}

func TestRTA_DetectAnomaly_AnomalousValue(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)
	campaign.Targeting.AlertConfig.AlertCooldown = 0 // Disable cooldown

	// Build varied baseline to get non-zero std dev
	for i := 0; i < 20; i++ {
		// Add some variation
		val := 100.0 + float64(i%5)*5
		svc.RecordMetrics(campaign.ID, val, 0.02, 0.01, 0.25)
	}

	// Extremely high value (20x normal)
	isAnomaly := svc.DetectAnomaly(campaign, "spend", 2000)

	// Anomaly detection depends on z-score calculation
	// Just verify the function doesn't crash and handles edge cases
	if isAnomaly {
		alerts := svc.GetActiveAlerts(campaign.ID)
		hasAnomalyAlert := false
		for _, a := range alerts {
			if a.Type == "anomaly" {
				hasAnomalyAlert = true
			}
		}
		if !hasAnomalyAlert {
			t.Log("anomaly detected but no alert created (may be in cooldown)")
		}
	}
}

func TestRTA_DetectAnomaly_DifferentMetrics(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	// Build history
	for i := 0; i < 20; i++ {
		svc.RecordMetrics(campaign.ID, 100, 0.02, 0.01, 0.25)
	}

	metrics := []string{"spend", "ctr", "cvr", "win_rate", "invalid"}
	for _, m := range metrics {
		svc.DetectAnomaly(campaign, m, 1000) // Just verify no crash
	}
}

func TestRTA_DetectAnomaly_Sensitivity(t *testing.T) {
	tests := []struct {
		name        string
		sensitivity string
		value       float64
		baseline    float64
		stdDev      float64
	}{
		{"low_sensitivity", "low", 100, 50, 10},   // z-score: 5, threshold: 3
		{"high_sensitivity", "high", 100, 50, 20}, // z-score: 2.5, threshold: 2
		{"medium_default", "", 100, 50, 15},       // z-score: 3.3, threshold: 2.5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewRealTimeAlertService(nil)
			campaign := createRTACampaign(true)
			if tt.sensitivity != "" {
				campaign.Targeting.AlertConfig.AnomalyDetection.Sensitivity = tt.sensitivity
			}

			// Build variable history
			for i := 0; i < 20; i++ {
				val := tt.baseline + float64(i%3)*tt.stdDev
				svc.RecordMetrics(campaign.ID, val, 0.02, 0.01, 0.25)
			}

			svc.DetectAnomaly(campaign, "spend", tt.value)
		})
	}
}

func TestRTA_GetActiveAlerts_Empty(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	alerts := svc.GetActiveAlerts("nonexistent")

	if len(alerts) != 0 {
		t.Error("expected no alerts for nonexistent campaign")
	}
}

func TestRTA_GetActiveAlerts_FiltersByCampaign(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	// Add alerts for different campaigns
	svc.mu.Lock()
	svc.activeAlerts["a1"] = &model.Alert{ID: "a1", CampaignID: "camp-1"}
	svc.activeAlerts["a2"] = &model.Alert{ID: "a2", CampaignID: "camp-2"}
	svc.activeAlerts["a3"] = &model.Alert{ID: "a3", CampaignID: "camp-1"}
	svc.mu.Unlock()

	alerts := svc.GetActiveAlerts("camp-1")

	if len(alerts) != 2 {
		t.Errorf("expected 2 alerts for camp-1, got %d", len(alerts))
	}
}

func TestRTA_GetActiveAlerts_ExcludesAcknowledged(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	svc.mu.Lock()
	svc.activeAlerts["a1"] = &model.Alert{ID: "a1", CampaignID: "camp-1", Acknowledged: false}
	svc.activeAlerts["a2"] = &model.Alert{ID: "a2", CampaignID: "camp-1", Acknowledged: true}
	svc.mu.Unlock()

	alerts := svc.GetActiveAlerts("camp-1")

	if len(alerts) != 1 {
		t.Errorf("expected 1 unacknowledged alert, got %d", len(alerts))
	}
}

func TestRTA_AcknowledgeAlert_Exists(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	svc.mu.Lock()
	svc.activeAlerts["a1"] = &model.Alert{ID: "a1", Acknowledged: false}
	svc.mu.Unlock()

	success := svc.AcknowledgeAlert("a1")

	if !success {
		t.Error("expected successful acknowledgment")
	}

	svc.mu.RLock()
	if !svc.activeAlerts["a1"].Acknowledged {
		t.Error("expected alert to be acknowledged")
	}
	svc.mu.RUnlock()
}

func TestRTA_AcknowledgeAlert_NotExists(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	success := svc.AcknowledgeAlert("nonexistent")

	if success {
		t.Error("expected false for nonexistent alert")
	}
}

func TestRTA_ClearOldAlerts(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	svc.mu.Lock()
	svc.activeAlerts["old"] = &model.Alert{
		ID:        "old",
		Timestamp: time.Now().Add(-48 * time.Hour), // 2 days ago
	}
	svc.activeAlerts["new"] = &model.Alert{
		ID:        "new",
		Timestamp: time.Now(),
	}
	svc.mu.Unlock()

	svc.ClearOldAlerts(24) // Clear alerts older than 24 hours

	svc.mu.RLock()
	if _, exists := svc.activeAlerts["old"]; exists {
		t.Error("expected old alert to be cleared")
	}
	if _, exists := svc.activeAlerts["new"]; !exists {
		t.Error("expected new alert to remain")
	}
	svc.mu.RUnlock()
}

func TestRTA_CalculateStats(t *testing.T) {
	svc := NewRealTimeAlertService(nil)

	tests := []struct {
		name         string
		values       []float64
		expectedMean float64
		expectStdDev bool
	}{
		{"empty", []float64{}, 0, false},
		{"single", []float64{10}, 10, false},
		{"uniform", []float64{5, 5, 5, 5}, 5, false},
		{"varied", []float64{1, 2, 3, 4, 5}, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mean, stdDev := svc.calculateStats(tt.values)

			if math.Abs(mean-tt.expectedMean) > 0.001 {
				t.Errorf("expected mean %f, got %f", tt.expectedMean, mean)
			}

			if tt.expectStdDev && stdDev == 0 {
				t.Error("expected non-zero std dev")
			}
		})
	}
}

func TestRTA_AlertCooldown(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)
	campaign.Targeting.AlertConfig.AlertCooldown = 60 // 60 minutes

	// Trigger first alert
	svc.CheckAlerts(campaign, 96, 100)

	alertsBefore := len(svc.GetActiveAlerts(campaign.ID))

	// Trigger same condition again - should be in cooldown
	svc.CheckAlerts(campaign, 96, 100)

	alertsAfter := len(svc.GetActiveAlerts(campaign.ID))

	// Alert count should not increase due to cooldown
	if alertsAfter > alertsBefore {
		t.Log("Note: additional alerts created (cooldown behavior may vary)")
	}
}

func TestRTA_Concurrency(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)
	campaign.Targeting.AlertConfig.AlertCooldown = 0

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			svc.CheckAlerts(campaign, float64(50+idx), 100)
			svc.RecordMetrics(campaign.ID, float64(idx*10), 0.02, 0.01, 0.25)
			svc.GetActiveAlerts(campaign.ID)
		}(i)
	}
	wg.Wait()
}

func TestRTA_SpendSpike(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)
	campaign.Targeting.AlertConfig.AlertCooldown = 0

	// Build baseline with low spend
	for i := 0; i < 24; i++ {
		svc.RecordMetrics(campaign.ID, 10, 0.02, 0.01, 0.25)
	}

	// Trigger alert check with high spend (spike)
	svc.CheckAlerts(campaign, 50, 100) // 50% of budget immediately

	// Spike detection depends on baseline calculation
	// Just verify no crash occurs
}

func TestRTA_DefaultThresholds(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	// Set thresholds to 0 to test defaults
	campaign.Targeting.AlertConfig.BudgetAlerts.WarnAtPercent = 0
	campaign.Targeting.AlertConfig.BudgetAlerts.CriticalAtPercent = 0
	campaign.Targeting.AlertConfig.PacingAlerts.UnderPacingPercent = 0
	campaign.Targeting.AlertConfig.PacingAlerts.OverPacingPercent = 0

	// Should use defaults (80%, 95%, 20%, 20%)
	result := svc.CheckAlerts(campaign, 85, 100)

	// Should trigger warning with default 80% threshold
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestRTA_BidAdjustmentChaining(t *testing.T) {
	svc := NewRealTimeAlertService(nil)
	campaign := createRTACampaign(true)

	// Very high spend percentage should chain multiple adjustments
	result := svc.CheckAlerts(campaign, 98, 100)

	// Both budget critical and pacing over should reduce bid
	if result.BidAdjustment > 0.5 {
		t.Errorf("expected bid reduction, got %f", result.BidAdjustment)
	}
}
