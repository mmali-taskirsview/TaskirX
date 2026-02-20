package service

import (
"testing"
)

func TestMean(t *testing.T) {
tests := []struct {
name     string
values   []float64
expected float64
}{
{"empty slice", []float64{}, 0.0},
{"single value", []float64{5.0}, 5.0},
{"multiple values", []float64{1.0, 2.0, 3.0, 4.0, 5.0}, 3.0},
{"negative values", []float64{-1.0, -2.0, -3.0}, -2.0},
{"mixed values", []float64{-1.0, 0.0, 1.0}, 0.0},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
result := mean(tt.values)
if result != tt.expected {
t.Errorf("mean(%v) = %v, expected %v", tt.values, result, tt.expected)
}
})
}
}
