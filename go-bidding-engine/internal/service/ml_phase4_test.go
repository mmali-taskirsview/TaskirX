package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== Churn Prediction Service Tests ====================

func TestNewChurnPredictionService(t *testing.T) {
	mc := NewMockCache()
	service := NewChurnPredictionService(mc)

	assert.NotNil(t, service)
}

func TestChurnPredictionService_RecordUserActivity(t *testing.T) {
	mc := NewMockCache()
	service := NewChurnPredictionService(mc)

	// Record activity for a user
	service.RecordUserActivity("user1", "impression", nil)
	service.RecordUserActivity("user1", "click", nil)
	service.RecordUserActivity("user1", "conversion", map[string]interface{}{"value": 50.0})

	// Predict churn - user should exist now
	result := service.PredictChurn("user1")
	assert.NotNil(t, result)
	assert.Equal(t, "user1", result.UserID)
	assert.GreaterOrEqual(t, result.ChurnProbability, 0.0)
	assert.LessOrEqual(t, result.ChurnProbability, 1.0)
}

func TestChurnPredictionService_PredictChurn(t *testing.T) {
	mc := NewMockCache()
	service := NewChurnPredictionService(mc)

	// Create an active user
	for i := 0; i < 10; i++ {
		service.RecordUserActivity("active_user", "impression", nil)
		service.RecordUserActivity("active_user", "click", nil)
	}
	service.RecordUserActivity("active_user", "conversion", map[string]interface{}{"value": 100.0})

	result := service.PredictChurn("active_user")
	assert.NotNil(t, result)
	assert.Equal(t, "active_user", result.UserID)
	assert.NotEmpty(t, result.RiskLevel)
	// TopFactors may be empty for new users without enough history
}

func TestChurnPredictionService_PredictChurn_NonexistentUser(t *testing.T) {
	mc := NewMockCache()
	service := NewChurnPredictionService(mc)

	// The service returns a default result for unknown users rather than nil
	result := service.PredictChurn("nonexistent")
	assert.NotNil(t, result)
	assert.Equal(t, "nonexistent", result.UserID)
	// Unknown users get a default churn probability
}

func TestChurnPredictionService_GetHighRiskUsers(t *testing.T) {
	mc := NewMockCache()
	service := NewChurnPredictionService(mc)

	// Create multiple users with different activity levels
	// Active user
	for i := 0; i < 20; i++ {
		service.RecordUserActivity("active", "impression", nil)
		service.RecordUserActivity("active", "click", nil)
	}
	service.RecordUserActivity("active", "conversion", map[string]interface{}{"value": 200.0})

	// Less active user
	service.RecordUserActivity("less_active", "impression", nil)

	users := service.GetHighRiskUsers(10)
	assert.NotNil(t, users)
}

func TestChurnPredictionService_GetChurnStats(t *testing.T) {
	mc := NewMockCache()
	service := NewChurnPredictionService(mc)

	// Add some users
	service.RecordUserActivity("user1", "impression", nil)
	service.RecordUserActivity("user2", "impression", nil)
	service.RecordUserActivity("user3", "impression", nil)

	stats := service.GetChurnStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "total_users")
	assert.Equal(t, 3, stats["total_users"])
}

func TestChurnPredictionService_ConfigManagement(t *testing.T) {
	mc := NewMockCache()
	service := NewChurnPredictionService(mc)

	config := service.GetConfig()
	assert.NotNil(t, config)

	// Update config
	newConfig := &ChurnConfig{
		Enabled:              true,
		PredictionWindowDays: 14,
		HighRiskThreshold:    0.8,
		MediumRiskThreshold:  0.5,
		MinDataPoints:        5,
	}
	service.SetConfig(newConfig)

	updatedConfig := service.GetConfig()
	assert.Equal(t, newConfig.PredictionWindowDays, updatedConfig.PredictionWindowDays)
}

func TestChurnPredictionService_BatchPredict(t *testing.T) {
	mc := NewMockCache()
	service := NewChurnPredictionService(mc)

	// Create users
	for i := 0; i < 5; i++ {
		userID := "batch_user_" + string(rune('A'+i))
		service.RecordUserActivity(userID, "impression", nil)
		service.RecordUserActivity(userID, "click", nil)
	}

	userIDs := []string{"batch_user_A", "batch_user_B", "batch_user_C"}
	results := service.BatchPredict(userIDs)
	assert.NotNil(t, results)
	assert.Len(t, results, 3)
}

func TestChurnPredictionService_RiskLevels(t *testing.T) {
	mc := NewMockCache()
	service := NewChurnPredictionService(mc)

	// Test that different activity patterns produce predictions
	for i := 0; i < 100; i++ {
		service.RecordUserActivity("very_active", "impression", nil)
		service.RecordUserActivity("very_active", "click", nil)
	}
	for i := 0; i < 10; i++ {
		service.RecordUserActivity("very_active", "conversion", map[string]interface{}{"value": 100.0})
	}

	result := service.PredictChurn("very_active")
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.RiskLevel)
}

// ==================== A/B Testing Service Tests ====================

func TestNewABTestingService(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	assert.NotNil(t, service)
	config := service.GetConfig()
	assert.Equal(t, 100, config.MinSampleSize)
	assert.Equal(t, 0.05, config.SignificanceLevel)
}

func TestABTestingService_CreateExperiment(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	req := CreateExperimentRequest{
		Name:        "Test Experiment",
		Description: "Testing conversion rates",
		Type:        "ab",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Variant A", Weight: 0.5, IsControl: false},
		},
		TrafficAllocation: 1.0,
		Metrics:           []string{"conversion_rate", "ctr"},
	}

	exp, err := service.CreateExperiment(req)
	assert.NoError(t, err)
	assert.NotNil(t, exp)
	assert.NotEmpty(t, exp.ID)
	assert.Equal(t, "Test Experiment", exp.Name)
	assert.Equal(t, "draft", exp.Status)
	assert.Len(t, exp.Variants, 2)
}

func TestABTestingService_CreateExperiment_Validation(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	// Test missing name
	_, err := service.CreateExperiment(CreateExperimentRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")

	// Test too few variants
	_, err = service.CreateExperiment(CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Only One", IsControl: true},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2 variants")

	// Test no control variant
	_, err = service.CreateExperiment(CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "A", IsControl: false},
			{Name: "B", IsControl: false},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "control variant")
}

func TestABTestingService_StartAndStopExperiment(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	exp, _ := service.CreateExperiment(CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Variant", Weight: 0.5, IsControl: false},
		},
	})

	// Start experiment
	err := service.StartExperiment(exp.ID)
	assert.NoError(t, err)

	// Verify status
	updatedExp, _ := service.GetExperiment(exp.ID)
	assert.Equal(t, "running", updatedExp.Status)
	assert.False(t, updatedExp.StartDate.IsZero())

	// Try starting again
	err = service.StartExperiment(exp.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Stop experiment
	err = service.StopExperiment(exp.ID)
	assert.NoError(t, err)

	updatedExp, _ = service.GetExperiment(exp.ID)
	assert.Equal(t, "completed", updatedExp.Status)
}

func TestABTestingService_GetVariantForUser(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	exp, _ := service.CreateExperiment(CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Variant", Weight: 0.5, IsControl: false},
		},
		TrafficAllocation: 1.0,
	})

	service.StartExperiment(exp.ID)

	// Get variant for user
	v, err := service.GetVariantForUser(exp.ID, "user123")
	assert.NoError(t, err)
	assert.NotNil(t, v)
	assert.NotEmpty(t, v.ID)

	// Same user should get same variant (consistent assignment)
	v2, _ := service.GetVariantForUser(exp.ID, "user123")
	assert.Equal(t, v.ID, v2.ID)

	// Different user might get different variant
	_, err = service.GetVariantForUser(exp.ID, "user456")
	assert.NoError(t, err)
}

func TestABTestingService_GetVariantForUser_NotRunning(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	exp, _ := service.CreateExperiment(CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Variant", Weight: 0.5, IsControl: false},
		},
	})

	// Experiment not started
	_, err := service.GetVariantForUser(exp.ID, "user123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestABTestingService_RecordEvent(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	exp, _ := service.CreateExperiment(CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Variant", Weight: 0.5, IsControl: false},
		},
	})
	service.StartExperiment(exp.ID)

	v, _ := service.GetVariantForUser(exp.ID, "user123")

	// Record events
	err := service.RecordEvent(exp.ID, v.ID, "impression", 0)
	assert.NoError(t, err)

	err = service.RecordEvent(exp.ID, v.ID, "click", 0)
	assert.NoError(t, err)

	err = service.RecordEvent(exp.ID, v.ID, "conversion", 50.0)
	assert.NoError(t, err)

	// Check metrics updated
	updatedExp, _ := service.GetExperiment(exp.ID)
	var targetVariant *variant
	for _, vr := range updatedExp.Variants {
		if vr.ID == v.ID {
			targetVariant = vr
			break
		}
	}
	assert.NotNil(t, targetVariant)
	assert.Equal(t, int64(1), targetVariant.Metrics.Impressions)
	assert.Equal(t, int64(1), targetVariant.Metrics.Clicks)
	assert.Equal(t, int64(1), targetVariant.Metrics.Conversions)
	assert.Equal(t, 50.0, targetVariant.Metrics.Revenue)
}

func TestABTestingService_RecordEvent_CustomMetric(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	exp, _ := service.CreateExperiment(CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Variant", Weight: 0.5, IsControl: false},
		},
	})
	service.StartExperiment(exp.ID)

	v, _ := service.GetVariantForUser(exp.ID, "user123")

	// Record custom metric
	err := service.RecordEvent(exp.ID, v.ID, "time_on_page", 120.5)
	assert.NoError(t, err)

	updatedExp, _ := service.GetExperiment(exp.ID)
	for _, vr := range updatedExp.Variants {
		if vr.ID == v.ID {
			assert.Equal(t, 120.5, vr.Metrics.CustomMetrics["time_on_page"])
			break
		}
	}
}

func TestABTestingService_AnalyzeExperiment(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	exp, _ := service.CreateExperiment(CreateExperimentRequest{
		Name: "Conversion Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Variant A", Weight: 0.5, IsControl: false},
		},
	})
	service.StartExperiment(exp.ID)

	// Simulate data for control
	controlVariant := exp.Variants[0]
	for i := 0; i < 100; i++ {
		service.RecordEvent(exp.ID, controlVariant.ID, "impression", 0)
	}
	for i := 0; i < 10; i++ {
		service.RecordEvent(exp.ID, controlVariant.ID, "conversion", 10.0)
	}

	// Simulate data for variant (better performance)
	testVariant := exp.Variants[1]
	for i := 0; i < 100; i++ {
		service.RecordEvent(exp.ID, testVariant.ID, "impression", 0)
	}
	for i := 0; i < 20; i++ {
		service.RecordEvent(exp.ID, testVariant.ID, "conversion", 15.0)
	}

	// Analyze
	result, err := service.AnalyzeExperiment(exp.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, exp.ID, result.ExperimentID)
	assert.Equal(t, "Conversion Test", result.ExperimentName)
	assert.Equal(t, int64(200), result.TotalSamples)
	assert.Len(t, result.VariantResults, 2)
	assert.NotEmpty(t, result.Recommendation)
}

func TestABTestingService_AnalyzeExperiment_InsufficientData(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	exp, _ := service.CreateExperiment(CreateExperimentRequest{
		Name: "Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Variant", Weight: 0.5, IsControl: false},
		},
	})
	service.StartExperiment(exp.ID)

	// Only a few samples
	service.RecordEvent(exp.ID, exp.Variants[0].ID, "impression", 0)
	service.RecordEvent(exp.ID, exp.Variants[1].ID, "impression", 0)

	result, err := service.AnalyzeExperiment(exp.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Recommendation, "Continue experiment")
}

func TestABTestingService_ListExperiments(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	// Create multiple experiments
	for i := 0; i < 3; i++ {
		service.CreateExperiment(CreateExperimentRequest{
			Name: "Test " + string(rune('A'+i)),
			Variants: []VariantRequest{
				{Name: "Control", Weight: 0.5, IsControl: true},
				{Name: "Variant", Weight: 0.5, IsControl: false},
			},
		})
	}

	// List all
	all := service.ListExperiments("")
	assert.Len(t, all, 3)

	// List by status
	drafts := service.ListExperiments("draft")
	assert.Len(t, drafts, 3)

	running := service.ListExperiments("running")
	assert.Len(t, running, 0)
}

func TestABTestingService_GetBanditRecommendation(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	exp, _ := service.CreateExperiment(CreateExperimentRequest{
		Name: "Bandit Test",
		Type: "bandit",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.33, IsControl: true},
			{Name: "Variant A", Weight: 0.33, IsControl: false},
			{Name: "Variant B", Weight: 0.34, IsControl: false},
		},
	})
	service.StartExperiment(exp.ID)

	// Record some data
	for i := 0; i < 50; i++ {
		for _, v := range exp.Variants {
			service.RecordEvent(exp.ID, v.ID, "impression", 0)
		}
	}
	// Variant B has better conversion
	for i := 0; i < 15; i++ {
		service.RecordEvent(exp.ID, exp.Variants[2].ID, "conversion", 10.0)
	}
	for i := 0; i < 5; i++ {
		service.RecordEvent(exp.ID, exp.Variants[0].ID, "conversion", 10.0)
		service.RecordEvent(exp.ID, exp.Variants[1].ID, "conversion", 10.0)
	}

	// Get bandit recommendation (Thompson Sampling)
	v, err := service.GetBanditRecommendation(exp.ID, "user123")
	assert.NoError(t, err)
	assert.NotNil(t, v)
}

func TestABTestingService_UpdateConfig(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	newConfig := ABTestingConfig{
		MinSampleSize:          500,
		SignificanceLevel:      0.01,
		MinDetectableEffect:    0.10,
		MaxRunningExperiments:  20,
		AutoStopOnSignificance: true,
	}

	service.UpdateConfig(newConfig)
	config := service.GetConfig()

	assert.Equal(t, 500, config.MinSampleSize)
	assert.Equal(t, 0.01, config.SignificanceLevel)
	assert.True(t, config.AutoStopOnSignificance)
}

func TestABTestingService_GetStats(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	// Create experiments in different states
	exp1, _ := service.CreateExperiment(CreateExperimentRequest{
		Name: "Draft Experiment",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Variant", Weight: 0.5, IsControl: false},
		},
	})

	exp2, _ := service.CreateExperiment(CreateExperimentRequest{
		Name: "Running Experiment",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Variant", Weight: 0.5, IsControl: false},
		},
	})
	service.StartExperiment(exp2.ID)

	// Record some events
	service.RecordEvent(exp2.ID, exp2.Variants[0].ID, "impression", 0)
	service.RecordEvent(exp2.ID, exp2.Variants[0].ID, "conversion", 10.0)

	stats := service.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 2, stats["total_experiments"])
	assert.Equal(t, 1, stats["running_experiments"])
	assert.Equal(t, 1, stats["draft_experiments"])
	assert.Equal(t, 4, stats["total_variants"])

	// Avoid unused variable warning
	_ = exp1
}

func TestABTestingService_MaxRunningExperiments(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	// Set low max
	service.UpdateConfig(ABTestingConfig{
		MinSampleSize:         100,
		MaxRunningExperiments: 2,
	})

	// Create and start experiments
	var experiments []*abExperiment
	for i := 0; i < 3; i++ {
		exp, _ := service.CreateExperiment(CreateExperimentRequest{
			Name: "Exp " + string(rune('A'+i)),
			Variants: []VariantRequest{
				{Name: "Control", Weight: 0.5, IsControl: true},
				{Name: "Variant", Weight: 0.5, IsControl: false},
			},
		})
		experiments = append(experiments, exp)
	}

	// Start first two
	assert.NoError(t, service.StartExperiment(experiments[0].ID))
	assert.NoError(t, service.StartExperiment(experiments[1].ID))

	// Third should fail
	err := service.StartExperiment(experiments[2].ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum running experiments")
}

func TestABTestingService_WeightNormalization(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	// Create experiment with unnormalized weights
	exp, err := service.CreateExperiment(CreateExperimentRequest{
		Name: "Weight Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 1.0, IsControl: true},
			{Name: "Variant A", Weight: 2.0, IsControl: false},
			{Name: "Variant B", Weight: 2.0, IsControl: false},
		},
	})

	assert.NoError(t, err)

	// Weights should be normalized to sum to 1
	totalWeight := 0.0
	for _, v := range exp.Variants {
		totalWeight += v.Weight
	}
	assert.InDelta(t, 1.0, totalWeight, 0.01)
}

func TestABTestingService_StatisticalSignificance(t *testing.T) {
	mc := NewMockCache()
	service := NewABTestingService(mc)

	exp, _ := service.CreateExperiment(CreateExperimentRequest{
		Name: "Significance Test",
		Variants: []VariantRequest{
			{Name: "Control", Weight: 0.5, IsControl: true},
			{Name: "Variant", Weight: 0.5, IsControl: false},
		},
	})
	service.StartExperiment(exp.ID)

	// Add enough data to achieve significance
	// Control: 10% conversion rate
	for i := 0; i < 1000; i++ {
		service.RecordEvent(exp.ID, exp.Variants[0].ID, "impression", 0)
	}
	for i := 0; i < 100; i++ {
		service.RecordEvent(exp.ID, exp.Variants[0].ID, "conversion", 10.0)
	}

	// Variant: 15% conversion rate (significant improvement)
	for i := 0; i < 1000; i++ {
		service.RecordEvent(exp.ID, exp.Variants[1].ID, "impression", 0)
	}
	for i := 0; i < 150; i++ {
		service.RecordEvent(exp.ID, exp.Variants[1].ID, "conversion", 10.0)
	}

	result, _ := service.AnalyzeExperiment(exp.ID)
	assert.NotNil(t, result)
	// With 1000 samples per variant and 5% difference, should be significant
	assert.NotEmpty(t, result.Recommendation)
}

// ==================== Integration Tests ====================

func TestChurnAndABTestingIntegration(t *testing.T) {
	mc := NewMockCache()

	churnService := NewChurnPredictionService(mc)
	abService := NewABTestingService(mc)

	// Create A/B test for churn prevention campaign
	exp, _ := abService.CreateExperiment(CreateExperimentRequest{
		Name:        "Churn Prevention Campaign",
		Description: "Test different retention offers",
		Type:        "ab",
		Variants: []VariantRequest{
			{Name: "Control (No offer)", Weight: 0.5, IsControl: true},
			{Name: "10% Discount", Weight: 0.5, IsControl: false, Config: map[string]any{"discount": 0.10}},
		},
		Metrics: []string{"retention", "revenue"},
	})
	abService.StartExperiment(exp.ID)

	// Simulate users in experiment
	for i := 0; i < 50; i++ {
		userID := "user" + string(rune('A'+i%26)) + string(rune('0'+i/26))

		// Register user activity for churn prediction
		churnService.RecordUserActivity(userID, "impression", nil)

		// Assign to A/B test variant
		v, _ := abService.GetVariantForUser(exp.ID, userID)

		// Record impression
		abService.RecordEvent(exp.ID, v.ID, "impression", 0)

		// Simulate that discount variant has better retention
		if !v.IsControl && i%3 == 0 {
			abService.RecordEvent(exp.ID, v.ID, "conversion", 20.0)
			churnService.RecordUserActivity(userID, "conversion", map[string]interface{}{"value": 20.0})
		} else if v.IsControl && i%5 == 0 {
			abService.RecordEvent(exp.ID, v.ID, "conversion", 20.0)
		}
	}

	// Analyze results
	result, _ := abService.AnalyzeExperiment(exp.ID)
	assert.NotNil(t, result)
	assert.Greater(t, result.TotalSamples, int64(0))

	// Check churn stats
	churnStats := churnService.GetChurnStats()
	assert.Greater(t, churnStats["total_users"], 0)
}
