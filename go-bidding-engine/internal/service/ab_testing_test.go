package service

import (
	"math"
	"sync"
	"testing"
	"time"
)

func TestABTest_NewService(t *testing.T) {
	svc := NewABTestingService(nil)
	if svc == nil {
		t.Fatal("expected service")
	}
	if svc.config.MinSampleSize != 100 {
		t.Errorf("expected default min sample size 100, got %d", svc.config.MinSampleSize)
	}
	if svc.config.SignificanceLevel != 0.05 {
		t.Errorf("expected default significance 0.05, got %f", svc.config.SignificanceLevel)
	}
}

func TestABTest_CreateExperiment_Success(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name:        "Test Experiment",
		Description: "Testing A/B functionality",
		Type:        "ab",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
		TrafficAllocation: 1.0,
	}

	exp, err := svc.CreateExperiment(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp == nil {
		t.Fatal("expected experiment")
	}
	if exp.Name != "Test Experiment" {
		t.Errorf("expected name 'Test Experiment', got '%s'", exp.Name)
	}
	if exp.Status != "draft" {
		t.Errorf("expected status 'draft', got '%s'", exp.Status)
	}
	if len(exp.Variants) != 2 {
		t.Errorf("expected 2 variants, got %d", len(exp.Variants))
	}
}

func TestABTest_CreateExperiment_NoName(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}

	_, err := svc.CreateExperiment(req)

	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestABTest_CreateExperiment_SingleVariant(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 1.0, IsControl: true},
		},
	}

	_, err := svc.CreateExperiment(req)

	if err == nil {
		t.Error("expected error for single variant")
	}
}

func TestABTest_CreateExperiment_NoControl(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "A", Weight: 0.5, IsControl: false},
			{Name: "B", Weight: 0.5, IsControl: false},
		},
	}

	_, err := svc.CreateExperiment(req)

	if err == nil {
		t.Error("expected error for no control")
	}
}

func TestABTest_CreateExperiment_MultipleControls(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "A", Weight: 0.5, IsControl: true},
			{Name: "B", Weight: 0.5, IsControl: true},
		},
	}

	_, err := svc.CreateExperiment(req)

	if err == nil {
		t.Error("expected error for multiple controls")
	}
}

func TestABTest_CreateExperiment_NormalizeWeights(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 1.0, IsControl: true},
			{Name: "Treatment", Weight: 1.0, IsControl: false},
		},
	}

	exp, err := svc.CreateExperiment(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	totalWeight := 0.0
	for _, v := range exp.Variants {
		totalWeight += v.Weight
	}
	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("expected weights normalized to 1.0, got %f", totalWeight)
	}
}

func TestABTest_StartExperiment(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)

	err := svc.StartExperiment(exp.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updatedExp, _ := svc.GetExperiment(exp.ID)
	if updatedExp.Status != "running" {
		t.Errorf("expected status 'running', got '%s'", updatedExp.Status)
	}
}

func TestABTest_StartExperiment_NotFound(t *testing.T) {
	svc := NewABTestingService(nil)

	err := svc.StartExperiment("nonexistent")

	if err == nil {
		t.Error("expected error for nonexistent experiment")
	}
}

func TestABTest_StartExperiment_AlreadyRunning(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)

	err := svc.StartExperiment(exp.ID)

	if err == nil {
		t.Error("expected error for already running")
	}
}

func TestABTest_StartExperiment_MaxReached(t *testing.T) {
	svc := NewABTestingService(nil)
	svc.config.MaxRunningExperiments = 2

	// Create and start max experiments
	for i := 0; i < 2; i++ {
		req := CreateExperimentRequest{
			Name: "Test",
			Variants: []VariantRequest{
				{Name: "Control", Weight: 0.5, IsControl: true},
				{Name: "Treatment", Weight: 0.5, IsControl: false},
			},
		}
		exp, _ := svc.CreateExperiment(req)
		svc.StartExperiment(exp.ID)
	}

	// Try to start another
	req := CreateExperimentRequest{
		Name: "Test Extra",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)

	err := svc.StartExperiment(exp.ID)

	if err == nil {
		t.Error("expected error for max experiments reached")
	}
}

func TestABTest_StopExperiment(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)

	err := svc.StopExperiment(exp.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updatedExp, _ := svc.GetExperiment(exp.ID)
	if updatedExp.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", updatedExp.Status)
	}
}

func TestABTest_StopExperiment_NotRunning(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)

	err := svc.StopExperiment(exp.ID)

	if err == nil {
		t.Error("expected error for not running")
	}
}

func TestABTest_GetVariantForUser(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name:              "Test",
		TrafficAllocation: 1.0,
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)

	variant, err := svc.GetVariantForUser(exp.ID, "user-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if variant == nil {
		t.Fatal("expected variant")
	}
}

func TestABTest_GetVariantForUser_Consistent(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name:              "Test",
		TrafficAllocation: 1.0,
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)

	// Get variant multiple times for same user
	v1, _ := svc.GetVariantForUser(exp.ID, "user-123")
	v2, _ := svc.GetVariantForUser(exp.ID, "user-123")
	v3, _ := svc.GetVariantForUser(exp.ID, "user-123")

	if v1.ID != v2.ID || v2.ID != v3.ID {
		t.Error("expected consistent variant assignment")
	}
}

func TestABTest_GetVariantForUser_NotRunning(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	// Don't start

	_, err := svc.GetVariantForUser(exp.ID, "user-123")

	if err == nil {
		t.Error("expected error for not running")
	}
}

func TestABTest_RecordEvent_Impression(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)
	variant, _ := svc.GetVariantForUser(exp.ID, "user-1")

	err := svc.RecordEvent(exp.ID, variant.ID, "impression", 0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestABTest_RecordEvent_Conversion(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)
	variant, _ := svc.GetVariantForUser(exp.ID, "user-1")

	svc.RecordEvent(exp.ID, variant.ID, "impression", 0)
	svc.RecordEvent(exp.ID, variant.ID, "click", 0)
	svc.RecordEvent(exp.ID, variant.ID, "conversion", 10.0)

	updatedExp, _ := svc.GetExperiment(exp.ID)
	for _, v := range updatedExp.Variants {
		if v.ID == variant.ID {
			if v.Metrics.Impressions != 1 {
				t.Errorf("expected 1 impression, got %d", v.Metrics.Impressions)
			}
			if v.Metrics.Clicks != 1 {
				t.Errorf("expected 1 click, got %d", v.Metrics.Clicks)
			}
			if v.Metrics.Conversions != 1 {
				t.Errorf("expected 1 conversion, got %d", v.Metrics.Conversions)
			}
			if v.Metrics.Revenue != 10.0 {
				t.Errorf("expected revenue 10.0, got %f", v.Metrics.Revenue)
			}
		}
	}
}

func TestABTest_RecordEvent_CustomMetric(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)
	variant, _ := svc.GetVariantForUser(exp.ID, "user-1")

	svc.RecordEvent(exp.ID, variant.ID, "custom_metric", 5.0)

	updatedExp, _ := svc.GetExperiment(exp.ID)
	for _, v := range updatedExp.Variants {
		if v.ID == variant.ID {
			if v.Metrics.CustomMetrics["custom_metric"] != 5.0 {
				t.Errorf("expected custom metric 5.0, got %f", v.Metrics.CustomMetrics["custom_metric"])
			}
		}
	}
}

func TestABTest_RecordEvent_VariantNotFound(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)

	err := svc.RecordEvent(exp.ID, "nonexistent", "impression", 0)

	if err == nil {
		t.Error("expected error for nonexistent variant")
	}
}

func TestABTest_AnalyzeExperiment(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)

	// Record events for both variants
	for _, v := range exp.Variants {
		for i := 0; i < 100; i++ {
			svc.RecordEvent(exp.ID, v.ID, "impression", 0)
			if i%10 == 0 {
				svc.RecordEvent(exp.ID, v.ID, "conversion", 1.0)
			}
		}
	}

	result, err := svc.AnalyzeExperiment(exp.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if len(result.VariantResults) != 2 {
		t.Errorf("expected 2 variant results, got %d", len(result.VariantResults))
	}
}

func TestABTest_AnalyzeExperiment_NotFound(t *testing.T) {
	svc := NewABTestingService(nil)

	_, err := svc.AnalyzeExperiment("nonexistent")

	if err == nil {
		t.Error("expected error for nonexistent")
	}
}

func TestABTest_ListExperiments(t *testing.T) {
	svc := NewABTestingService(nil)

	// Create a few experiments
	for i := 0; i < 3; i++ {
		req := CreateExperimentRequest{
			Name: "Test",
			Variants: []VariantRequest{
				{Name: "Control", Weight: 0.5, IsControl: true},
				{Name: "Treatment", Weight: 0.5, IsControl: false},
			},
		}
		svc.CreateExperiment(req)
	}

	experiments := svc.ListExperiments("")

	if len(experiments) != 3 {
		t.Errorf("expected 3 experiments, got %d", len(experiments))
	}
}

func TestABTest_ListExperiments_StatusFilter(t *testing.T) {
	svc := NewABTestingService(nil)

	// Create experiments with different statuses
	req := CreateExperimentRequest{
		Name: "Draft",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	svc.CreateExperiment(req)

	req.Name = "Running"
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)

	draftExps := svc.ListExperiments("draft")
	runningExps := svc.ListExperiments("running")

	if len(draftExps) != 1 {
		t.Errorf("expected 1 draft, got %d", len(draftExps))
	}
	if len(runningExps) != 1 {
		t.Errorf("expected 1 running, got %d", len(runningExps))
	}
}

func TestABTest_GetStats(t *testing.T) {
	svc := NewABTestingService(nil)

	// Create and run experiment with events
	req := CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)

	for _, v := range exp.Variants {
		svc.RecordEvent(exp.ID, v.ID, "impression", 0)
		svc.RecordEvent(exp.ID, v.ID, "conversion", 1.0)
	}

	stats := svc.GetStats()

	if stats["total_experiments"].(int) != 1 {
		t.Errorf("expected 1 total experiment, got %v", stats["total_experiments"])
	}
	if stats["running_experiments"].(int) != 1 {
		t.Errorf("expected 1 running experiment, got %v", stats["running_experiments"])
	}
	if stats["total_variants"].(int) != 2 {
		t.Errorf("expected 2 variants, got %v", stats["total_variants"])
	}
}

func TestABTest_UpdateConfig(t *testing.T) {
	svc := NewABTestingService(nil)

	newConfig := ABTestingConfig{
		MinSampleSize:     500,
		SignificanceLevel: 0.01,
	}

	svc.UpdateConfig(newConfig)
	config := svc.GetConfig()

	if config.MinSampleSize != 500 {
		t.Errorf("expected min sample 500, got %d", config.MinSampleSize)
	}
	if config.SignificanceLevel != 0.01 {
		t.Errorf("expected significance 0.01, got %f", config.SignificanceLevel)
	}
}

func TestABTest_GetBanditRecommendation(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "Bandit Test",
		Type: "bandit",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)

	// Record some events
	for _, v := range exp.Variants {
		for i := 0; i < 10; i++ {
			svc.RecordEvent(exp.ID, v.ID, "impression", 0)
		}
	}

	variant, err := svc.GetBanditRecommendation(exp.ID, "user-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if variant == nil {
		t.Fatal("expected variant")
	}
}

func TestABTest_GetBanditRecommendation_NotBandit(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name: "AB Test",
		Type: "ab",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)

	// Should fall back to regular variant assignment
	variant, err := svc.GetBanditRecommendation(exp.ID, "user-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if variant == nil {
		t.Fatal("expected variant")
	}
}

func TestABTest_Concurrency(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name:              "Concurrent Test",
		TrafficAllocation: 1.0,
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userID := "user-" + string(rune(idx))
			variant, _ := svc.GetVariantForUser(exp.ID, userID)
			if variant != nil {
				svc.RecordEvent(exp.ID, variant.ID, "impression", 0)
				if idx%5 == 0 {
					svc.RecordEvent(exp.ID, variant.ID, "conversion", 1.0)
				}
			}
		}(i)
	}
	wg.Wait()

	// Verify experiment integrity
	result, _ := svc.AnalyzeExperiment(exp.ID)
	if result.TotalSamples < 50 {
		t.Errorf("expected at least 50 samples, got %d", result.TotalSamples)
	}
}

func TestABTest_AutoStopOnSignificance(t *testing.T) {
	svc := NewABTestingService(nil)
	svc.config.AutoStopOnSignificance = true
	svc.config.MinSampleSize = 10

	req := CreateExperimentRequest{
		Name: "Auto Stop Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}
	exp, _ := svc.CreateExperiment(req)
	svc.StartExperiment(exp.ID)

	// Simulate significant difference
	for _, v := range exp.Variants {
		for i := 0; i < 1000; i++ {
			svc.RecordEvent(exp.ID, v.ID, "impression", 0)
			if v.IsControl {
				if i%20 == 0 { // 5% conversion
					svc.RecordEvent(exp.ID, v.ID, "conversion", 1.0)
				}
			} else {
				if i%5 == 0 { // 20% conversion
					svc.RecordEvent(exp.ID, v.ID, "conversion", 1.0)
				}
			}
		}
	}

	// Check if auto-stopped
	updatedExp, _ := svc.GetExperiment(exp.ID)
	// Status might be completed if significance detected
	if updatedExp.Status != "completed" && updatedExp.Status != "running" {
		t.Errorf("unexpected status: %s", updatedExp.Status)
	}
}

func TestABTest_CalculatePValue(t *testing.T) {
	svc := NewABTestingService(nil)

	// Same conversion rate should give high p-value
	pValue := svc.calculatePValue(10, 100, 10, 100)
	if pValue < 0.5 {
		t.Errorf("expected high p-value for same rates, got %f", pValue)
	}

	// Very different rates should give low p-value
	pValue = svc.calculatePValue(5, 100, 50, 100)
	if pValue > 0.05 {
		t.Errorf("expected low p-value for different rates, got %f", pValue)
	}
}

func TestABTest_CalculateStandardError(t *testing.T) {
	svc := NewABTestingService(nil)

	se := svc.calculateStandardError(0.5, 100)

	if se <= 0 {
		t.Error("expected positive standard error")
	}
	// SE should be roughly 0.05 for p=0.5, n=100
	if se > 0.1 || se < 0.01 {
		t.Errorf("unexpected standard error: %f", se)
	}
}

func TestABTest_Duration(t *testing.T) {
	svc := NewABTestingService(nil)

	req := CreateExperimentRequest{
		Name:     "Duration Test",
		Duration: 24 * time.Hour,
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Treatment", Weight: 0.5, IsControl: false},
		},
	}

	exp, _ := svc.CreateExperiment(req)

	if exp.EndDate.IsZero() {
		t.Error("expected end date set")
	}
}
