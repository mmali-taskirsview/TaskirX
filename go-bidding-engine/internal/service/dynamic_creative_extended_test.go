package service

import (
"testing"
)

func TestEvaluateRule(t *testing.T) {
s := NewDynamicCreativeService(NewMockCache())

// Rule with no conditions should return true
rule := PersonalizationRule{
ID:         "rule1",
Conditions: []RuleCondition{},
}
context := DCOContext{
PageCategory: "sports",
DeviceType:   "mobile",
}

result := s.evaluateRule(rule, context)
if !result {
t.Error("evaluateRule() with empty conditions should return true")
}

// Rule with matching condition
rule = PersonalizationRule{
ID: "rule2",
Conditions: []RuleCondition{
{Field: "device.type", Operator: "equals", Value: "mobile"},
},
}
result = s.evaluateRule(rule, context)
if !result {
t.Error("evaluateRule() with matching condition should return true")
}

// Rule with non-matching condition
rule = PersonalizationRule{
ID: "rule3",
Conditions: []RuleCondition{
{Field: "device.type", Operator: "equals", Value: "desktop"},
},
}
result = s.evaluateRule(rule, context)
if result {
t.Error("evaluateRule() with non-matching condition should return false")
}
}

func TestEvaluateCondition(t *testing.T) {
s := NewDynamicCreativeService(NewMockCache())

tests := []struct {
name     string
cond     RuleCondition
context  DCOContext
expected bool
}{
{
name:     "equals match",
cond:     RuleCondition{Field: "device.type", Operator: "equals", Value: "mobile"},
context:  DCOContext{DeviceType: "mobile"},
expected: true,
},
{
name:     "equals no match",
cond:     RuleCondition{Field: "device.type", Operator: "equals", Value: "desktop"},
context:  DCOContext{DeviceType: "mobile"},
expected: false,
},
{
name:     "context category match",
cond:     RuleCondition{Field: "context.category", Operator: "equals", Value: "sports"},
context:  DCOContext{PageCategory: "sports"},
expected: true,
},
{
name:     "time day",
cond:     RuleCondition{Field: "time.day", Operator: "equals", Value: "morning"},
context:  DCOContext{TimeOfDay: "morning"},
expected: true,
},
{
name:     "geo location",
cond:     RuleCondition{Field: "geo.location", Operator: "equals", Value: "US"},
context:  DCOContext{GeoLocation: "US"},
expected: true,
},
{
name:     "unknown field",
cond:     RuleCondition{Field: "unknown.field", Operator: "equals", Value: "test"},
context:  DCOContext{},
expected: false,
},
{
name:     "contains operator",
cond:     RuleCondition{Field: "device.type", Operator: "contains", Value: "mobile"},
context:  DCOContext{DeviceType: "mobile"},
expected: true,
},
{
name:     "in operator with match",
cond:     RuleCondition{Field: "device.type", Operator: "in", Value: []string{"mobile", "tablet"}},
context:  DCOContext{DeviceType: "mobile"},
expected: true,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
result := s.evaluateCondition(tt.cond, tt.context)
if result != tt.expected {
t.Errorf("evaluateCondition() = %v, expected %v", result, tt.expected)
}
})
}
}

func TestGetTotalImpressions(t *testing.T) {
s := NewDynamicCreativeService(NewMockCache())

// Initially should return at least 1
result := s.getTotalImpressions()
if result < 1 {
t.Errorf("getTotalImpressions() = %v, expected >= 1", result)
}
}
