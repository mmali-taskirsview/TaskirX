package service

import (
	"testing"
)

// ==================== Evaluate Rule Tests ====================

func TestEvaluateRule(t *testing.T) {
	service := NewDynamicCreativeService(nil)

	tests := []struct {
		name    string
		rule    PersonalizationRule
		context DCOContext
		want    bool
	}{
		{
			name: "All conditions pass",
			rule: PersonalizationRule{
				Conditions: []RuleCondition{
					{Field: "device.type", Operator: "equals", Value: "mobile"},
				},
			},
			context: DCOContext{DeviceType: "mobile"},
			want:    true,
		},
		{
			name: "Condition fails",
			rule: PersonalizationRule{
				Conditions: []RuleCondition{
					{Field: "device.type", Operator: "equals", Value: "desktop"},
				},
			},
			context: DCOContext{DeviceType: "mobile"},
			want:    false,
		},
		{
			name:    "Empty conditions - should pass",
			rule:    PersonalizationRule{Conditions: []RuleCondition{}},
			context: DCOContext{},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.evaluateRule(tt.rule, tt.context)
			if result != tt.want {
				t.Errorf("evaluateRule() = %v, want %v", result, tt.want)
			}
		})
	}
}

// ==================== Evaluate Condition Tests ====================

func TestEvaluateCondition(t *testing.T) {
	service := NewDynamicCreativeService(nil)

	tests := []struct {
		name    string
		cond    RuleCondition
		context DCOContext
		want    bool
	}{
		{
			name:    "Equals - device type match",
			cond:    RuleCondition{Field: "device.type", Operator: "equals", Value: "mobile"},
			context: DCOContext{DeviceType: "mobile"},
			want:    true,
		},
		{
			name:    "Equals - device type no match",
			cond:    RuleCondition{Field: "device.type", Operator: "equals", Value: "desktop"},
			context: DCOContext{DeviceType: "mobile"},
			want:    false,
		},
		{
			name:    "Equals - context category",
			cond:    RuleCondition{Field: "context.category", Operator: "equals", Value: "sports"},
			context: DCOContext{PageCategory: "sports"},
			want:    true,
		},
		{
			name:    "Equals - time of day",
			cond:    RuleCondition{Field: "time.day", Operator: "equals", Value: "morning"},
			context: DCOContext{TimeOfDay: "morning"},
			want:    true,
		},
		{
			name:    "Equals - geo location",
			cond:    RuleCondition{Field: "geo.location", Operator: "equals", Value: "US"},
			context: DCOContext{GeoLocation: "US"},
			want:    true,
		},
		{
			name:    "Contains - category",
			cond:    RuleCondition{Field: "context.category", Operator: "contains", Value: "news"},
			context: DCOContext{PageCategory: "news"},
			want:    true,
		},
		{
			name:    "In - device type in list",
			cond:    RuleCondition{Field: "device.type", Operator: "in", Value: []string{"mobile", "tablet"}},
			context: DCOContext{DeviceType: "mobile"},
			want:    true,
		},
		{
			name:    "In - device type not in list",
			cond:    RuleCondition{Field: "device.type", Operator: "in", Value: []string{"mobile", "tablet"}},
			context: DCOContext{DeviceType: "desktop"},
			want:    false,
		},
		{
			name:    "Unknown field",
			cond:    RuleCondition{Field: "unknown.field", Operator: "equals", Value: "test"},
			context: DCOContext{},
			want:    false,
		},
		{
			name:    "Unknown operator",
			cond:    RuleCondition{Field: "device.type", Operator: "unknown", Value: "mobile"},
			context: DCOContext{DeviceType: "mobile"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.evaluateCondition(tt.cond, tt.context)
			if result != tt.want {
				t.Errorf("evaluateCondition() = %v, want %v", result, tt.want)
			}
		})
	}
}

// ==================== Get Total Impressions Tests ====================

func TestGetTotalImpressions(t *testing.T) {
	service := NewDynamicCreativeService(nil)

	// Test with no combinations - should return 1 (minimum)
	result := service.getTotalImpressions()
	if result < 1 {
		t.Errorf("getTotalImpressions() with no data = %v, want >= 1", result)
	}
}

// ==================== Hash Combination Tests ====================

func TestHashCombination(t *testing.T) {
	service := NewDynamicCreativeService(nil)

	tests := []struct {
		name       string
		templateID string
		elements   map[string]string
	}{
		{
			name:       "Single element",
			templateID: "template1",
			elements:   map[string]string{"headline": "elem1"},
		},
		{
			name:       "Multiple elements",
			templateID: "template2",
			elements:   map[string]string{"headline": "elem1", "image": "elem2", "cta": "elem3"},
		},
		{
			name:       "Empty elements",
			templateID: "template3",
			elements:   map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.hashCombination(tt.templateID, tt.elements)
			// Should start with template ID
			if len(result) < len(tt.templateID) {
				t.Errorf("hashCombination() = %q, should contain templateID %q", result, tt.templateID)
			}

			// Same input should give same hash
			result2 := service.hashCombination(tt.templateID, tt.elements)
			if result != result2 {
				t.Errorf("hashCombination() not deterministic: %q vs %q", result, result2)
			}
		})
	}
}

func TestHashCombination_Deterministic(t *testing.T) {
	service := NewDynamicCreativeService(nil)

	// Order of elements should not affect hash (sorted internally)
	elements1 := map[string]string{"b": "2", "a": "1", "c": "3"}
	elements2 := map[string]string{"a": "1", "b": "2", "c": "3"}

	hash1 := service.hashCombination("template1", elements1)
	hash2 := service.hashCombination("template1", elements2)

	if hash1 != hash2 {
		t.Errorf("hashCombination() not order-independent: %q vs %q", hash1, hash2)
	}
}

// ==================== Calculate Personalization Score Tests ====================

func TestCalculatePersonalizationScore(t *testing.T) {
	service := NewDynamicCreativeService(nil)

	tests := []struct {
		name     string
		elements map[string]*CreativeElement
		userPref *UserCreativePreference
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "Empty elements",
			elements: map[string]*CreativeElement{},
			userPref: &UserCreativePreference{},
			wantMin:  0,
			wantMax:  0,
		},
		{
			name: "User has engaged with element",
			elements: map[string]*CreativeElement{
				"headline": {ID: "elem1"},
			},
			userPref: &UserCreativePreference{
				EngagedElements: map[string]int{"elem1": 10},
			},
			wantMin: 0.5,
			wantMax: 1.0,
		},
		{
			name: "User has converted with element",
			elements: map[string]*CreativeElement{
				"headline": {ID: "elem1"},
			},
			userPref: &UserCreativePreference{
				ConvertedElements: map[string]int{"elem1": 2},
			},
			wantMin: 0.5,
			wantMax: 1.0,
		},
		{
			name: "No engagement history",
			elements: map[string]*CreativeElement{
				"headline": {ID: "elem1"},
			},
			userPref: &UserCreativePreference{},
			wantMin:  0,
			wantMax:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculatePersonalizationScore(tt.elements, tt.userPref)
			if result < tt.wantMin {
				t.Errorf("calculatePersonalizationScore() = %v, want >= %v", result, tt.wantMin)
			}
			if result > tt.wantMax {
				t.Errorf("calculatePersonalizationScore() = %v, want <= %v", result, tt.wantMax)
			}
		})
	}
}

// ==================== Predict CTR Tests ====================

func TestDCOPredictCTR(t *testing.T) {
	service := NewDynamicCreativeService(nil)
	service.config.MinImpressionsForStats = 100

	tests := []struct {
		name    string
		combo   *CreativeCombination
		context DCOContext
		wantMin float64
		wantMax float64
	}{
		{
			name: "Low impressions - default CTR",
			combo: &CreativeCombination{
				Performance: &CombinationPerformance{Impressions: 50, CTR: 0.05},
			},
			context: DCOContext{},
			wantMin: 0.005,
			wantMax: 0.02,
		},
		{
			name: "Sufficient impressions - use actual CTR",
			combo: &CreativeCombination{
				Performance: &CombinationPerformance{Impressions: 200, CTR: 0.03},
			},
			context: DCOContext{},
			wantMin: 0.02,
			wantMax: 0.05,
		},
		{
			name: "Mobile device boost",
			combo: &CreativeCombination{
				Performance: &CombinationPerformance{Impressions: 200, CTR: 0.02},
			},
			context: DCOContext{DeviceType: "mobile"},
			wantMin: 0.02,
			wantMax: 0.03,
		},
		{
			name: "Desktop device decrease",
			combo: &CreativeCombination{
				Performance: &CombinationPerformance{Impressions: 200, CTR: 0.02},
			},
			context: DCOContext{DeviceType: "desktop"},
			wantMin: 0.01,
			wantMax: 0.025,
		},
		{
			name: "Evening time boost",
			combo: &CreativeCombination{
				Performance: &CombinationPerformance{Impressions: 200, CTR: 0.02},
			},
			context: DCOContext{TimeOfDay: "evening"},
			wantMin: 0.02,
			wantMax: 0.03,
		},
		{
			name: "Night time decrease",
			combo: &CreativeCombination{
				Performance: &CombinationPerformance{Impressions: 200, CTR: 0.02},
			},
			context: DCOContext{TimeOfDay: "night"},
			wantMin: 0.015,
			wantMax: 0.025,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.predictCTR(tt.combo, tt.context)
			if result < tt.wantMin {
				t.Errorf("predictCTR() = %v, want >= %v", result, tt.wantMin)
			}
			if result > tt.wantMax {
				t.Errorf("predictCTR() = %v, want <= %v", result, tt.wantMax)
			}
		})
	}
}

// ==================== Render Creative Tests ====================

func TestRenderCreative(t *testing.T) {
	service := NewDynamicCreativeService(nil)

	tests := []struct {
		name         string
		template     *CreativeTemplate
		elements     map[string]*CreativeElement
		wantContains string
	}{
		{
			name: "With base HTML",
			template: &CreativeTemplate{
				BaseHTML: "<div class='ad'>{{headline}}</div>",
			},
			elements: map[string]*CreativeElement{
				"headline": {ID: "h1", Content: "Buy Now!"},
			},
			wantContains: "ad",
		},
		{
			name:     "Without base HTML - auto generate",
			template: &CreativeTemplate{},
			elements: map[string]*CreativeElement{
				"headline": {ID: "h1", Content: "Buy Now!"},
			},
			wantContains: "dco-creative",
		},
		{
			name:     "Multiple elements",
			template: &CreativeTemplate{},
			elements: map[string]*CreativeElement{
				"headline": {ID: "h1", Content: "Title"},
				"cta":      {ID: "c1", Content: "Click"},
			},
			wantContains: "dco-creative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.renderCreative(tt.template, tt.elements)
			if len(result) == 0 {
				t.Errorf("renderCreative() returned empty string")
			}
			// Check for expected content
			found := false
			for i := 0; i <= len(result)-len(tt.wantContains); i++ {
				if result[i:i+len(tt.wantContains)] == tt.wantContains {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("renderCreative() = %q, should contain %q", result, tt.wantContains)
			}
		})
	}
}

// ==================== Determine Selection Method Tests ====================

func TestDetermineSelectionMethod(t *testing.T) {
	service := NewDynamicCreativeService(nil)

	tests := []struct {
		name         string
		rulesApplied []string
		want         string
	}{
		{
			name:         "No rules - ML optimized",
			rulesApplied: []string{},
			want:         "ml_optimized",
		},
		{
			name:         "Exploration rule",
			rulesApplied: []string{"exploration"},
			want:         "exploration",
		},
		{
			name:         "Other rules - rule based",
			rulesApplied: []string{"personalization"},
			want:         "rule_based",
		},
		{
			name:         "Multiple rules with exploration",
			rulesApplied: []string{"personalization", "exploration"},
			want:         "exploration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.determineSelectionMethod(tt.rulesApplied)
			if result != tt.want {
				t.Errorf("determineSelectionMethod() = %q, want %q", result, tt.want)
			}
		})
	}
}

// ==================== Score Element Tests ====================

func TestScoreElement(t *testing.T) {
	service := NewDynamicCreativeService(nil)
	service.config.MinImpressionsForStats = 100
	service.config.PerformanceWeight = 0.5
	service.config.PersonalizationWeight = 0.3
	service.config.ContextWeight = 0.2

	tests := []struct {
		name     string
		elem     *CreativeElement
		userPref *UserCreativePreference
		context  DCOContext
		wantMin  float64
		wantMax  float64
	}{
		{
			name: "New element - exploration bonus",
			elem: &CreativeElement{
				ID:          "elem1",
				Performance: &ElementPerformance{Impressions: 10, CTR: 0.02},
			},
			userPref: &UserCreativePreference{},
			context:  DCOContext{},
			wantMin:  0,
			wantMax:  0.5,
		},
		{
			name: "Element with engagement",
			elem: &CreativeElement{
				ID:          "elem1",
				Performance: &ElementPerformance{Impressions: 10, CTR: 0.02},
			},
			userPref: &UserCreativePreference{
				EngagedElements: map[string]int{"elem1": 5},
			},
			context: DCOContext{},
			wantMin: 0.1,
			wantMax: 0.6,
		},
		{
			name: "Element with matching tags",
			elem: &CreativeElement{
				ID:          "elem1",
				Tags:        []string{"sports", "fitness"},
				Performance: &ElementPerformance{Impressions: 10, CTR: 0.02},
			},
			userPref: &UserCreativePreference{},
			context:  DCOContext{ContentKeywords: []string{"sports", "health"}},
			wantMin:  0.1,
			wantMax:  0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.scoreElement(tt.elem, tt.userPref, tt.context)
			if result < tt.wantMin {
				t.Errorf("scoreElement() = %v, want >= %v", result, tt.wantMin)
			}
			if result > tt.wantMax {
				t.Errorf("scoreElement() = %v, want <= %v", result, tt.wantMax)
			}
		})
	}
}
