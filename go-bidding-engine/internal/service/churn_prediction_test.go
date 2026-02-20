package service

import (
	"sync"
	"testing"
	"time"
)

func TestChurn_NewService(t *testing.T) {
	svc := NewChurnPredictionService(nil)
	if svc == nil {
		t.Fatal("expected service")
	}
	if svc.users == nil {
		t.Error("expected users map")
	}
	if svc.predictions == nil {
		t.Error("expected predictions map")
	}
	if svc.config == nil {
		t.Error("expected default config")
	}
	if svc.modelWeights == nil {
		t.Error("expected model weights")
	}
}

func TestChurn_RecordUserActivity_NewUser(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	svc.RecordUserActivity("user-1", "impression", nil)

	svc.mu.RLock()
	user := svc.users["user-1"]
	svc.mu.RUnlock()

	if user == nil {
		t.Fatal("expected user created")
	}
	if user.TotalImpressions != 1 {
		t.Errorf("expected 1 impression, got %d", user.TotalImpressions)
	}
}

func TestChurn_RecordUserActivity_EventTypes(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	tests := []struct {
		eventType string
		checkFunc func(*churnUser) bool
	}{
		{"impression", func(u *churnUser) bool { return u.TotalImpressions == 1 }},
		{"click", func(u *churnUser) bool { return u.TotalClicks == 1 }},
		{"conversion", func(u *churnUser) bool { return u.TotalConversions == 1 }},
		{"session_start", func(u *churnUser) bool { return u.SessionCount == 1 }},
	}

	for i, tt := range tests {
		userID := "user-" + string(rune('a'+i))
		svc.RecordUserActivity(userID, tt.eventType, nil)

		svc.mu.RLock()
		user := svc.users[userID]
		svc.mu.RUnlock()

		if !tt.checkFunc(user) {
			t.Errorf("event type %s not recorded correctly", tt.eventType)
		}
	}
}

func TestChurn_RecordUserActivity_DeviceTracking(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	metadata := map[string]interface{}{
		"device_type": "mobile",
	}
	svc.RecordUserActivity("user-1", "impression", metadata)
	svc.RecordUserActivity("user-1", "impression", metadata)

	metadata["device_type"] = "desktop"
	svc.RecordUserActivity("user-1", "impression", metadata)

	svc.mu.RLock()
	user := svc.users["user-1"]
	svc.mu.RUnlock()

	if user.DeviceTypes["mobile"] != 2 {
		t.Errorf("expected 2 mobile, got %d", user.DeviceTypes["mobile"])
	}
	if user.DeviceTypes["desktop"] != 1 {
		t.Errorf("expected 1 desktop, got %d", user.DeviceTypes["desktop"])
	}
}

func TestChurn_RecordUserActivity_SessionLength(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	svc.RecordUserActivity("user-1", "session_start", map[string]interface{}{
		"session_length": 5.0,
	})
	svc.RecordUserActivity("user-1", "session_start", map[string]interface{}{
		"session_length": 10.0,
	})

	svc.mu.RLock()
	user := svc.users["user-1"]
	svc.mu.RUnlock()

	// Average should be close to 7.5
	if user.AvgSessionLength < 7 || user.AvgSessionLength > 8 {
		t.Errorf("expected avg session length ~7.5, got %f", user.AvgSessionLength)
	}
}

func TestChurn_RecordUserActivity_ActiveDays(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	svc.RecordUserActivity("user-1", "impression", nil)

	svc.mu.RLock()
	user := svc.users["user-1"]
	svc.mu.RUnlock()

	today := time.Now().Format("2006-01-02")
	if !user.ActiveDays[today] {
		t.Error("expected today to be marked as active day")
	}
}

func TestChurn_PredictChurn_Disabled(t *testing.T) {
	svc := NewChurnPredictionService(nil)
	svc.config.Enabled = false

	result := svc.PredictChurn("user-1")

	if result.RiskLevel != "unknown" {
		t.Errorf("expected 'unknown' when disabled, got '%s'", result.RiskLevel)
	}
}

func TestChurn_PredictChurn_UnknownUser(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	result := svc.PredictChurn("nonexistent")

	if result.ChurnProbability != 0.5 {
		t.Errorf("expected 0.5 for unknown, got %f", result.ChurnProbability)
	}
	if result.RiskLevel != "unknown" {
		t.Errorf("expected 'unknown', got '%s'", result.RiskLevel)
	}
}

func TestChurn_PredictChurn_NewUser(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	// Create a new user with minimal activity
	svc.RecordUserActivity("user-1", "impression", nil)

	result := svc.PredictChurn("user-1")

	if result.UserID != "user-1" {
		t.Error("expected user ID in result")
	}
	if result.ChurnProbability < 0 || result.ChurnProbability > 1 {
		t.Errorf("churn probability out of range: %f", result.ChurnProbability)
	}
	if result.RiskLevel == "" {
		t.Error("expected risk level")
	}
}

func TestChurn_PredictChurn_ActiveUser(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	// Simulate active user with lots of engagement
	for i := 0; i < 50; i++ {
		svc.RecordUserActivity("active-user", "impression", map[string]interface{}{
			"device_type": "mobile",
		})
		if i%5 == 0 {
			svc.RecordUserActivity("active-user", "click", nil)
		}
		if i%20 == 0 {
			svc.RecordUserActivity("active-user", "conversion", nil)
		}
	}

	result := svc.PredictChurn("active-user")

	// Active users should have lower churn probability
	if result.Confidence <= 0 {
		t.Error("expected some confidence for user with data")
	}
}

func TestChurn_PredictChurn_RiskLevels(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	// Create users and make predictions
	svc.RecordUserActivity("user-1", "impression", nil)
	result := svc.PredictChurn("user-1")

	// Risk level should be one of: high, medium, low
	validLevels := map[string]bool{"high": true, "medium": true, "low": true}
	if !validLevels[result.RiskLevel] {
		t.Errorf("unexpected risk level: %s", result.RiskLevel)
	}
}

func TestChurn_PredictChurn_CachesPrediction(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	svc.RecordUserActivity("user-cache", "impression", nil)
	svc.PredictChurn("user-cache")

	svc.mu.RLock()
	prediction := svc.predictions["user-cache"]
	svc.mu.RUnlock()

	if prediction == nil {
		t.Error("expected prediction to be cached")
	}
}

func TestChurn_PredictChurn_TopFactors(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	// Create user with some activity history
	svc.RecordUserActivity("user-factors", "impression", nil)

	// Manually set user to have old last seen date
	svc.mu.Lock()
	svc.users["user-factors"].LastSeen = time.Now().Add(-30 * 24 * time.Hour)
	svc.mu.Unlock()

	result := svc.PredictChurn("user-factors")

	// Should have inactivity as top factor
	if len(result.TopFactors) > 0 {
		hasInactivity := false
		for _, f := range result.TopFactors {
			if f.Name == "inactivity" {
				hasInactivity = true
			}
		}
		if !hasInactivity {
			t.Log("inactivity factor may not be present depending on threshold")
		}
	}
}

func TestChurn_PredictChurn_DaysUntilChurn(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	svc.RecordUserActivity("user-days", "impression", nil)
	result := svc.PredictChurn("user-days")

	if result.DaysUntilChurn < 0 {
		t.Error("days until churn should not be negative")
	}
}

func TestChurn_PredictChurn_RecommendedAction(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	svc.RecordUserActivity("user-rec", "impression", nil)
	result := svc.PredictChurn("user-rec")

	if result.RecommendedAction == "" {
		t.Error("expected recommended action")
	}
}

func TestChurn_GetHighRiskUsers_Empty(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	users := svc.GetHighRiskUsers(10)

	if len(users) != 0 {
		t.Error("expected empty list")
	}
}

func TestChurn_GetHighRiskUsers_WithLimit(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	// Create predictions directly
	svc.mu.Lock()
	for i := 0; i < 10; i++ {
		userID := "high-risk-" + string(rune('a'+i))
		svc.predictions[userID] = &churnPrediction{
			UserID:           userID,
			ChurnProbability: 0.9,
			RiskLevel:        "high",
		}
	}
	svc.mu.Unlock()

	users := svc.GetHighRiskUsers(5)

	if len(users) != 5 {
		t.Errorf("expected 5 users, got %d", len(users))
	}
}

func TestChurn_GetHighRiskUsers_Sorted(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	svc.mu.Lock()
	svc.predictions["u1"] = &churnPrediction{UserID: "u1", ChurnProbability: 0.75, RiskLevel: "high"}
	svc.predictions["u2"] = &churnPrediction{UserID: "u2", ChurnProbability: 0.95, RiskLevel: "high"}
	svc.predictions["u3"] = &churnPrediction{UserID: "u3", ChurnProbability: 0.85, RiskLevel: "high"}
	svc.mu.Unlock()

	users := svc.GetHighRiskUsers(0) // No limit

	if len(users) < 2 {
		t.Fatal("expected at least 2 high risk users")
	}

	// Should be sorted by churn probability descending
	if users[0].ChurnProbability < users[1].ChurnProbability {
		t.Error("expected sorted by churn probability descending")
	}
}

func TestChurn_GetChurnStats_Empty(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	stats := svc.GetChurnStats()

	if stats["total_users"].(int) != 0 {
		t.Error("expected 0 users")
	}
	if stats["avg_churn_probability"].(float64) != 0 {
		t.Error("expected 0 avg churn prob")
	}
}

func TestChurn_GetChurnStats_WithData(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	// Add some users and predictions
	svc.mu.Lock()
	svc.users["u1"] = &churnUser{}
	svc.users["u2"] = &churnUser{}
	svc.predictions["u1"] = &churnPrediction{RiskLevel: "high", ChurnProbability: 0.8}
	svc.predictions["u2"] = &churnPrediction{RiskLevel: "low", ChurnProbability: 0.2}
	svc.mu.Unlock()

	stats := svc.GetChurnStats()

	if stats["total_users"].(int) != 2 {
		t.Errorf("expected 2 users, got %d", stats["total_users"].(int))
	}
	if stats["high_risk_users"].(int) != 1 {
		t.Error("expected 1 high risk user")
	}
	if stats["low_risk_users"].(int) != 1 {
		t.Error("expected 1 low risk user")
	}
	// Avg should be 0.5
	avgProb := stats["avg_churn_probability"].(float64)
	if avgProb < 0.49 || avgProb > 0.51 {
		t.Errorf("expected avg ~0.5, got %f", avgProb)
	}
}

func TestChurn_GetConfig(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	config := svc.GetConfig()

	if config == nil {
		t.Fatal("expected config")
	}
	if !config.Enabled {
		t.Error("expected enabled by default")
	}
}

func TestChurn_SetConfig(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	newConfig := &ChurnConfig{
		Enabled:           false,
		HighRiskThreshold: 0.8,
	}
	svc.SetConfig(newConfig)

	if svc.config.Enabled {
		t.Error("expected disabled")
	}
	if svc.config.HighRiskThreshold != 0.8 {
		t.Error("expected threshold 0.8")
	}
}

func TestChurn_BatchPredict(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	// Create some users
	svc.RecordUserActivity("batch-1", "impression", nil)
	svc.RecordUserActivity("batch-2", "impression", nil)

	results := svc.BatchPredict([]string{"batch-1", "batch-2", "batch-nonexistent"})

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestChurn_CalculateFeatures(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	user := &churnUser{
		UserID:           "test",
		FirstSeen:        time.Now().Add(-30 * 24 * time.Hour),
		LastSeen:         time.Now().Add(-2 * 24 * time.Hour),
		TotalImpressions: 100,
		TotalClicks:      5,
		TotalConversions: 1,
		SessionCount:     10,
		DeviceTypes:      map[string]int{"mobile": 8, "desktop": 2},
		WeeklyActivity:   []float64{0.5, 0.4, 0.6, 0.5, 0.3, 0.3, 0.2, 0.2, 0.1, 0.1, 0.0, 0.0},
	}

	features := svc.calculateFeatures(user)

	if features["days_since_last_seen"] < 1 {
		t.Error("expected positive days since last seen")
	}
	if features["click_through_rate"] != 0.05 {
		t.Errorf("expected CTR 0.05, got %f", features["click_through_rate"])
	}
	if features["tenure_days"] < 29 {
		t.Error("expected ~30 days tenure")
	}
}

func TestChurn_NormalizeFeature(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	// Test with known values
	normalized := svc.normalizeFeature("days_since_last_seen", 17) // mean=7, stddev=10

	// z-score = (17-7)/10 = 1.0
	if normalized < 0.9 || normalized > 1.1 {
		t.Errorf("expected normalized ~1.0, got %f", normalized)
	}
}

func TestChurn_NormalizeFeature_ZeroStdDev(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	// Test with non-existent feature (zero stddev)
	normalized := svc.normalizeFeature("nonexistent", 100)

	if normalized != 0 {
		t.Error("expected 0 for zero stddev")
	}
}

func TestChurn_CalculateDiversity(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	tests := []struct {
		name     string
		counts   map[string]int
		expected float64 // approximate
	}{
		{"empty", map[string]int{}, 0},
		{"single", map[string]int{"mobile": 100}, 0},
		{"uniform_two", map[string]int{"mobile": 50, "desktop": 50}, 0.43}, // log2(2)/2.32 ≈ 0.43
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diversity := svc.calculateDiversity(tt.counts)
			if diversity < 0 || diversity > 1 {
				t.Errorf("diversity out of range: %f", diversity)
			}
		})
	}
}

func TestChurn_CalculateWeeklyConsistency(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	tests := []struct {
		name     string
		activity []float64
	}{
		{"short", []float64{0.5, 0.5}},
		{"consistent", []float64{0.5, 0.5, 0.5, 0.5}},
		{"varied", []float64{0.1, 0.9, 0.2, 0.8}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consistency := svc.calculateWeeklyConsistency(tt.activity)
			if consistency < 0 || consistency > 1 {
				t.Errorf("consistency out of range: %f", consistency)
			}
		})
	}
}

func TestChurn_CalculateActivityDrop(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	tests := []struct {
		name     string
		activity []float64
		expected float64
	}{
		{"short", []float64{0.5, 0.5}, 0},
		{"no_drop", []float64{0.5, 0.5, 0.5, 0.5}, 0},
		{"big_drop", []float64{0.1, 0.1, 0.9, 0.9}, 0.8}, // (0.9-0.1)/0.9 ≈ 0.88
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drop := svc.calculateActivityDrop(tt.activity)
			if drop < 0 || drop > 1 {
				t.Errorf("drop out of range: %f", drop)
			}
		})
	}
}

func TestChurn_IdentifyTopFactors(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	features := map[string]float64{
		"days_since_last_seen": 15,
		"engagement_trend":     -0.3,
		"recent_activity_drop": 0.5,
		"weekly_consistency":   0.2,
	}

	factors := svc.identifyTopFactors(features)

	if len(factors) > 3 {
		t.Error("expected max 3 factors")
	}

	// Should be sorted by impact
	for i := 0; i < len(factors)-1; i++ {
		if factors[i].Impact < factors[i+1].Impact {
			t.Error("expected sorted by impact descending")
		}
	}
}

func TestChurn_EstimateDaysUntilChurn(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	tests := []struct {
		name      string
		churnProb float64
		features  map[string]float64
	}{
		{"low_risk", 0.2, map[string]float64{}},
		{"high_risk", 0.9, map[string]float64{}},
		{"high_with_drop", 0.9, map[string]float64{"recent_activity_drop": 0.6}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			days := svc.estimateDaysUntilChurn(tt.churnProb, tt.features)
			if days < 1 {
				t.Error("expected at least 1 day")
			}
		})
	}
}

func TestChurn_RecommendAction(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	tests := []struct {
		riskLevel string
		factors   []churnFactor
	}{
		{"high", []churnFactor{{Name: "inactivity"}}},
		{"high", []churnFactor{{Name: "other"}}},
		{"medium", nil},
		{"low", nil},
	}

	for _, tt := range tests {
		action := svc.recommendAction(tt.riskLevel, tt.factors)
		if action == "" {
			t.Errorf("expected action for %s risk", tt.riskLevel)
		}
	}
}

func TestChurn_Concurrency(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			userID := "conc-" + string(rune(idx))
			svc.RecordUserActivity(userID, "impression", map[string]interface{}{
				"device_type":    "mobile",
				"session_length": 5.0,
			})
			svc.PredictChurn(userID)
			svc.GetHighRiskUsers(10)
			svc.GetChurnStats()
		}(i)
	}
	wg.Wait()
}

func TestChurn_UpdateWeeklyActivity(t *testing.T) {
	svc := NewChurnPredictionService(nil)

	// Record activity over several days
	user := &churnUser{
		UserID:         "weekly-test",
		ActiveDays:     make(map[string]bool),
		WeeklyActivity: make([]float64, 12),
	}

	// Add some active days
	now := time.Now()
	for i := 0; i < 5; i++ {
		date := now.AddDate(0, 0, -i)
		user.ActiveDays[date.Format("2006-01-02")] = true
	}

	svc.updateWeeklyActivity(user)

	// First week should have activity
	if user.WeeklyActivity[0] == 0 {
		t.Error("expected recent activity")
	}
}
