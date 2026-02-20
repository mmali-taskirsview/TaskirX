package service

import (
"testing"

"github.com/taskirx/go-bidding-engine/internal/model"
)

func createAudienceModelingTestService() *AudienceModelingService {
cache := NewMockCache()
return NewAudienceModelingService(cache)
}

func TestGeoSimilarity(t *testing.T) {
s := createAudienceModelingTestService()

// Base score for empty request
result := s.geoSimilarity(&model.BidRequest{})
if result < 0.29 || result > 0.31 {
t.Errorf("geoSimilarity() base = %v, expected 0.3", result)
}

// With country
result = s.geoSimilarity(&model.BidRequest{User: model.InternalUser{Country: "US"}})
if result < 0.29 || result > 0.31 {
t.Errorf("geoSimilarity() with country = %v, expected 0.3", result)
}

// With city using InternalGeo
result = s.geoSimilarity(&model.BidRequest{Device: model.InternalDevice{Geo: model.InternalGeo{City: "New York"}}})
if result < 0.29 || result > 0.31 {
t.Errorf("geoSimilarity() with city = %v, expected 0.3", result)
}
}

func TestDeviceSimilarity(t *testing.T) {
s := createAudienceModelingTestService()

// Base score for empty request
result := s.deviceSimilarity(&model.BidRequest{})
if result < 0.39 || result > 0.41 {
t.Errorf("deviceSimilarity() base = %v, expected 0.4", result)
}

// With device type
result = s.deviceSimilarity(&model.BidRequest{Device: model.InternalDevice{Type: "mobile"}})
if result < 0.39 || result > 0.41 {
t.Errorf("deviceSimilarity() with type = %v, expected 0.4", result)
}

// With OS
result = s.deviceSimilarity(&model.BidRequest{Device: model.InternalDevice{Type: "mobile", OS: "ios"}})
if result < 0.39 || result > 0.41 {
t.Errorf("deviceSimilarity() with OS = %v, expected 0.4", result)
}

// With browser
result = s.deviceSimilarity(&model.BidRequest{Device: model.InternalDevice{Type: "desktop", OS: "windows", Browser: "chrome"}})
if result < 0.39 || result > 0.41 {
t.Errorf("deviceSimilarity() with browser = %v, expected 0.4", result)
}
}

func TestCalculateChurnRiskPropensity(t *testing.T) {
s := createAudienceModelingTestService()

// Test with basic request
result := s.calculateChurnRiskPropensity(&model.BidRequest{})
if result < 0.0 || result > 1.0 {
t.Errorf("calculateChurnRiskPropensity() = %v, expected between 0 and 1", result)
}

// Test with user data
result = s.calculateChurnRiskPropensity(&model.BidRequest{
User: model.InternalUser{ID: "user123"},
Device: model.InternalDevice{Type: "mobile"},
})
if result < 0.0 || result > 1.0 {
t.Errorf("calculateChurnRiskPropensity() with user = %v, expected between 0 and 1", result)
}
}

func TestEvaluateLookalike(t *testing.T) {
s := createAudienceModelingTestService()

campaign := &model.Campaign{ID: "camp1"}
req := &model.BidRequest{
User: model.InternalUser{ID: "user123"},
Device: model.InternalDevice{Type: "mobile"},
}
am := &model.AudienceModeling{
LookalikeEnabled: true,
SeedSegments: []string{"segment1"},
}
userSegments := []string{"segment2", "segment3"}

result := s.evaluateLookalike(campaign, req, am, userSegments)
if result.Multiplier < 0.5 || result.Multiplier > 2.0 {
t.Errorf("evaluateLookalike() multiplier = %v, expected between 0.5 and 2.0", result.Multiplier)
}
}

func TestCalculateSimilarityScore(t *testing.T) {
s := createAudienceModelingTestService()

req := &model.BidRequest{
User: model.InternalUser{ID: "user123", Age: 25, Gender: "male"},
Device: model.InternalDevice{Type: "mobile", OS: "ios"},
}
am := &model.AudienceModeling{
LookalikeEnabled: true,
SeedSegments: []string{"segment1"},
LookalikeFeatures: []string{"demographics", "interests"},
}
userSegments := []string{"segment2"}

result := s.calculateSimilarityScore(req, am, userSegments)
if result < 0.0 || result > 1.0 {
t.Errorf("calculateSimilarityScore() = %v, expected between 0 and 1", result)
}
}

func TestDemographicSimilarity(t *testing.T) {
s := createAudienceModelingTestService()

// Empty request
result := s.demographicSimilarity(&model.BidRequest{})
if result < 0.0 || result > 1.0 {
t.Errorf("demographicSimilarity() base = %v, expected between 0 and 1", result)
}

// With age and gender (Age is int)
result = s.demographicSimilarity(&model.BidRequest{
User: model.InternalUser{Age: 30, Gender: "male"},
})
if result < 0.0 || result > 1.0 {
t.Errorf("demographicSimilarity() with demo = %v, expected between 0 and 1", result)
}
}

func TestCalculateConversionPropensity(t *testing.T) {
s := createAudienceModelingTestService()

// Empty request
result := s.calculateConversionPropensity(&model.BidRequest{})
if result < 0.0 || result > 1.0 {
t.Errorf("calculateConversionPropensity() = %v, expected between 0 and 1", result)
}

// With user data
result = s.calculateConversionPropensity(&model.BidRequest{
User: model.InternalUser{ID: "user123"},
Device: model.InternalDevice{Type: "mobile"},
Context: map[string]interface{}{"page_type": "product"},
})
if result < 0.0 || result > 1.0 {
t.Errorf("calculateConversionPropensity() with data = %v, expected between 0 and 1", result)
}
}

func TestCalculateLTVPropensity(t *testing.T) {
s := createAudienceModelingTestService()

// Empty request
result := s.calculateLTVPropensity(&model.BidRequest{})
if result < 0.0 || result > 1.0 {
t.Errorf("calculateLTVPropensity() = %v, expected between 0 and 1", result)
}

// With user data
result = s.calculateLTVPropensity(&model.BidRequest{
User: model.InternalUser{ID: "user123"},
Device: model.InternalDevice{Type: "desktop", OS: "windows"},
})
if result < 0.0 || result > 1.0 {
t.Errorf("calculateLTVPropensity() with data = %v, expected between 0 and 1", result)
}
}
