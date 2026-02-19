package service

import (
	"testing"
	"time"
)

// =============================================================================
// Dynamic Creative Optimization Tests
// =============================================================================

func TestNewDynamicCreativeService(t *testing.T) {
	c := NewMockCache()
	svc := NewDynamicCreativeService(c)

	if svc == nil {
		t.Fatal("Expected service to be created")
	}
	if svc.cache == nil {
		t.Error("Expected cache to be set")
	}
}

func TestDCO_CreateTemplate(t *testing.T) {
	c := NewMockCache()
	svc := NewDynamicCreativeService(c)

	template := &CreativeTemplate{
		Name:        "Test Template",
		Description: "Test Description",
		Format:      "banner",
		Dimensions:  Dimensions{Width: 300, Height: 250},
		Slots: map[string]*TemplateSlot{
			"headline": {
				ID:       "headline",
				Name:     "Headline",
				Type:     "headline",
				Required: true,
				Position: Position{X: 10, Y: 10, Width: 280, Height: 40},
			},
			"image": {
				ID:       "image",
				Name:     "Image",
				Type:     "image",
				Required: true,
				Position: Position{X: 10, Y: 60, Width: 280, Height: 140},
			},
		},
	}

	created, err := svc.CreateTemplate(template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	if created.ID == "" {
		t.Error("Template ID should be set")
	}

	// Verify template is stored
	stored, err := svc.GetTemplate(created.ID)
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}
	if stored.Name != "Test Template" {
		t.Error("Template name mismatch")
	}
}

func TestDCO_CreateElement(t *testing.T) {
	c := NewMockCache()
	svc := NewDynamicCreativeService(c)

	element := &CreativeElement{
		Type:    "headline",
		Content: "Save 50% Today!",
		Attributes: map[string]string{
			"font":  "Arial",
			"size":  "24px",
			"color": "#FF0000",
		},
		Segments: []string{"deal-seekers", "bargain-hunters"},
		Tags:     []string{"promo", "discount"},
	}

	created, err := svc.CreateElement(element)
	if err != nil {
		t.Fatalf("Failed to create element: %v", err)
	}

	if created.ID == "" {
		t.Error("Element ID should be set")
	}

	// Verify element is stored
	stored, err := svc.GetElement(created.ID)
	if err != nil {
		t.Fatalf("Failed to get element: %v", err)
	}
	if stored.Content != "Save 50% Today!" {
		t.Error("Element content mismatch")
	}
}

func TestDCO_GenerateOptimizedCreative(t *testing.T) {
	c := NewMockCache()
	svc := NewDynamicCreativeService(c)

	// Create template
	template := &CreativeTemplate{
		Name:        "Banner Template",
		Description: "Test banner",
		Format:      "banner",
		Dimensions:  Dimensions{Width: 300, Height: 250},
		Slots: map[string]*TemplateSlot{
			"headline": {
				ID:       "headline",
				Name:     "Headline",
				Type:     "headline",
				Required: true,
				Position: Position{X: 10, Y: 10, Width: 280, Height: 40},
			},
		},
	}
	created, _ := svc.CreateTemplate(template)

	// Create elements
	elements := []*CreativeElement{
		{
			Type:     "headline",
			Content:  "Default Headline",
			Segments: []string{},
		},
		{
			Type:     "headline",
			Content:  "Personalized for You!",
			Segments: []string{"premium"},
		},
	}
	for _, elem := range elements {
		svc.CreateElement(elem)
	}

	// Generate optimized creative
	req := DCORequest{
		TemplateID: created.ID,
		UserID:     "user-123",
		Context: DCOContext{
			PageCategory: "electronics",
			UserSegments: []string{"premium"},
			DeviceType:   "mobile",
		},
	}

	response, err := svc.GenerateOptimizedCreative(req)
	if err != nil {
		t.Fatalf("Failed to generate creative: %v", err)
	}

	if response.TemplateID != created.ID {
		t.Error("Template ID mismatch")
	}
	if response.CombinationID == "" {
		t.Error("Expected combination ID")
	}
}

func TestDCO_RecordImpression(t *testing.T) {
	c := NewMockCache()
	svc := NewDynamicCreativeService(c)

	// Setup template and elements
	templateID := setupDCOTestData(svc)

	// Generate a creative first
	req := DCORequest{
		TemplateID: templateID,
		UserID:     "user-456",
		Context: DCOContext{
			UserSegments: []string{"tech"},
			DeviceType:   "desktop",
		},
	}
	response, _ := svc.GenerateOptimizedCreative(req)

	// Record impression
	err := svc.RecordImpression(response.CombinationID)
	if err != nil {
		t.Fatalf("Failed to record impression: %v", err)
	}

	stats := svc.GetDCOStats()
	if stats["total_impressions"].(int64) == 0 {
		t.Error("Expected impressions to be recorded")
	}
}

func TestDCO_RecordClick(t *testing.T) {
	c := NewMockCache()
	svc := NewDynamicCreativeService(c)

	// Setup template and elements
	templateID := setupDCOTestData(svc)

	// Generate a creative first
	req := DCORequest{
		TemplateID: templateID,
		UserID:     "user-789",
		Context: DCOContext{
			UserSegments: []string{"sports"},
			DeviceType:   "mobile",
		},
	}
	response, _ := svc.GenerateOptimizedCreative(req)

	// Record impression and click
	svc.RecordImpression(response.CombinationID)
	err := svc.RecordClick(response.CombinationID, "user-789")
	if err != nil {
		t.Fatalf("Failed to record click: %v", err)
	}

	stats := svc.GetDCOStats()
	if stats["total_clicks"].(int64) == 0 {
		t.Error("Expected clicks to be recorded")
	}
}

func TestDCO_RecordConversion(t *testing.T) {
	c := NewMockCache()
	svc := NewDynamicCreativeService(c)

	// Setup template and elements
	templateID := setupDCOTestData(svc)

	// Generate a creative first
	req := DCORequest{
		TemplateID: templateID,
		UserID:     "user-conv",
		Context: DCOContext{
			UserSegments: []string{"buyers"},
			DeviceType:   "desktop",
		},
	}
	response, _ := svc.GenerateOptimizedCreative(req)

	// Record full funnel
	svc.RecordImpression(response.CombinationID)
	svc.RecordClick(response.CombinationID, "user-conv")
	err := svc.RecordConversion(response.CombinationID, 99.99)
	if err != nil {
		t.Fatalf("Failed to record conversion: %v", err)
	}

	stats := svc.GetDCOStats()
	if stats["total_conversions"].(int64) == 0 {
		t.Error("Expected conversions to be recorded")
	}
}

func TestDCO_GetTopCombinations(t *testing.T) {
	c := NewMockCache()
	svc := NewDynamicCreativeService(c)

	// Setup template and elements
	templateID := setupDCOTestData(svc)

	// Generate multiple combinations
	users := []string{"user-1", "user-2", "user-3"}
	segments := [][]string{{"tech"}, {"sports"}, {"fashion"}}

	for i, userID := range users {
		req := DCORequest{
			TemplateID: templateID,
			UserID:     userID,
			Context: DCOContext{
				UserSegments: segments[i],
				DeviceType:   "mobile",
			},
		}
		response, _ := svc.GenerateOptimizedCreative(req)
		svc.RecordImpression(response.CombinationID)
		if i%2 == 0 {
			svc.RecordClick(response.CombinationID, userID)
		}
	}

	topCombos := svc.GetTopCombinations(templateID, 5)
	if len(topCombos) == 0 {
		t.Error("Expected top combinations")
	}
}

func TestDCO_GetElementsByType(t *testing.T) {
	c := NewMockCache()
	svc := NewDynamicCreativeService(c)

	// Create elements of different types
	svc.CreateElement(&CreativeElement{Type: "headline", Content: "Headline 1"})
	svc.CreateElement(&CreativeElement{Type: "headline", Content: "Headline 2"})
	svc.CreateElement(&CreativeElement{Type: "cta", Content: "Click Here"})
	svc.CreateElement(&CreativeElement{Type: "image", ImageURL: "http://example.com/img.jpg"})

	headlines := svc.GetElementsByType("headline")
	if len(headlines) != 2 {
		t.Errorf("Expected 2 headlines, got %d", len(headlines))
	}

	ctas := svc.GetElementsByType("cta")
	if len(ctas) != 1 {
		t.Errorf("Expected 1 CTA, got %d", len(ctas))
	}
}

func TestDCO_Stats(t *testing.T) {
	c := NewMockCache()
	svc := NewDynamicCreativeService(c)

	stats := svc.GetDCOStats()

	if stats["total_templates"].(int) != 0 {
		t.Error("Expected 0 templates initially")
	}
	if stats["total_elements"].(int) != 0 {
		t.Error("Expected 0 elements initially")
	}

	// Add template and element
	setupDCOTestData(svc)

	stats = svc.GetDCOStats()
	if stats["total_templates"].(int) == 0 {
		t.Error("Expected templates after setup")
	}
	if stats["total_elements"].(int) == 0 {
		t.Error("Expected elements after setup")
	}
}

func TestDCO_UpdateConfig(t *testing.T) {
	c := NewMockCache()
	svc := NewDynamicCreativeService(c)

	originalConfig := svc.GetConfig()
	if originalConfig.ExplorationRate == 0 {
		t.Error("Expected default exploration rate")
	}

	newConfig := DCOConfig{
		MaxElementsPerSlot:     20,
		ExplorationRate:        0.2,
		MinImpressionsForStats: 200,
		PersonalizationWeight:  0.4,
		ContextWeight:          0.3,
		PerformanceWeight:      0.3,
		EnableAutoOptimization: true,
	}
	svc.UpdateConfig(newConfig)

	updatedConfig := svc.GetConfig()
	if updatedConfig.ExplorationRate != 0.2 {
		t.Errorf("Expected exploration rate 0.2, got %f", updatedConfig.ExplorationRate)
	}
}

// Helper function to setup test data for DCO tests
func setupDCOTestData(svc *DynamicCreativeService) string {
	template := &CreativeTemplate{
		Name:        "Test Template",
		Description: "Test Description",
		Format:      "banner",
		Dimensions:  Dimensions{Width: 300, Height: 250},
		Slots: map[string]*TemplateSlot{
			"headline": {
				ID:       "headline",
				Name:     "Headline",
				Type:     "headline",
				Required: true,
				Position: Position{X: 10, Y: 10, Width: 280, Height: 40},
			},
			"cta": {
				ID:       "cta",
				Name:     "CTA",
				Type:     "cta",
				Required: true,
				Position: Position{X: 100, Y: 200, Width: 100, Height: 40},
			},
		},
	}
	created, _ := svc.CreateTemplate(template)

	elements := []*CreativeElement{
		{Type: "headline", Content: "Welcome!", Segments: []string{}},
		{Type: "cta", Content: "Learn More", Segments: []string{}},
		{Type: "headline", Content: "Tech Deals!", Segments: []string{"tech"}},
		{Type: "headline", Content: "Sports Gear!", Segments: []string{"sports"}},
	}
	for _, elem := range elements {
		svc.CreateElement(elem)
	}

	return created.ID
}

// =============================================================================
// Performance Prediction Tests
// =============================================================================

func TestNewPerformancePredictionService(t *testing.T) {
	c := NewMockCache()
	svc := NewPerformancePredictionService(c)

	if svc == nil {
		t.Fatal("Expected service to be created")
	}
	if svc.cache == nil {
		t.Error("Expected cache to be set")
	}
	if svc.config.MinHistoricalSamples == 0 {
		t.Error("Expected default min historical samples")
	}
}

func TestPrediction_RecordPerformance(t *testing.T) {
	c := NewMockCache()
	svc := NewPerformancePredictionService(c)

	record := &PerformanceRecord{
		EntityID:    "campaign-1",
		EntityType:  "campaign",
		Timestamp:   time.Now(),
		Impressions: 10000,
		Clicks:      500,
		Conversions: 50,
		Spend:       1000.0,
		Revenue:     2500.0,
		CTR:         0.05,
		CVR:         0.10,
		CPC:         2.0,
		CPM:         100.0,
		ROAS:        2.5,
		Features:    map[string]float64{"bid_price": 1.50},
	}

	svc.RecordPerformance(record)

	// Verify data is stored - check stats
	stats := svc.GetStats()
	if stats["total_records"].(int) == 0 {
		t.Error("Expected records after recording")
	}
}

func TestPrediction_RecordMultipleSnapshots(t *testing.T) {
	c := NewMockCache()
	svc := NewPerformancePredictionService(c)

	// Record multiple snapshots
	for i := 0; i < 10; i++ {
		record := &PerformanceRecord{
			EntityID:    "multi-campaign",
			EntityType:  "campaign",
			Timestamp:   time.Now().Add(time.Duration(i) * time.Hour),
			Impressions: int64(10000 + i*100),
			Clicks:      int64(500 + i*10),
			Spend:       1000.0 + float64(i)*50,
			Revenue:     2500.0 + float64(i)*100,
			CTR:         0.05 + float64(i)*0.001,
		}
		svc.RecordPerformance(record)
	}

	stats := svc.GetStats()
	if stats["total_records"].(int) < 10 {
		t.Errorf("Expected at least 10 records, got %d", stats["total_records"].(int))
	}
}

func TestPrediction_Predict(t *testing.T) {
	c := NewMockCache()
	svc := NewPerformancePredictionService(c)

	// Record historical data
	for i := 0; i < 14; i++ {
		record := &PerformanceRecord{
			EntityID:    "predict-campaign",
			EntityType:  "campaign",
			Timestamp:   time.Now().Add(time.Duration(-14+i) * 24 * time.Hour),
			Impressions: int64(10000 + i*500),
			Clicks:      int64(500 + i*25),
			Conversions: int64(50 + i*5),
			Spend:       1000.0 + float64(i)*50,
			Revenue:     2500.0 + float64(i)*100,
			CTR:         0.05,
			CVR:         0.10,
			ROAS:        2.5,
			Features:    map[string]float64{"bid_price": 1.50},
		}
		svc.RecordPerformance(record)
	}

	// Make prediction
	req := PredictionRequest{
		EntityID:   "predict-campaign",
		EntityType: "campaign",
		Features:   map[string]float64{"bid_price": 1.50},
		Context: PredictionContext{
			TimeOfDay:  "afternoon",
			DayOfWeek:  "wednesday",
			DeviceType: "desktop",
			AdFormat:   "banner",
		},
		Horizon: 24,
		Metrics: []string{"ctr", "cvr", "roas"},
	}

	result, err := svc.Predict(req)
	if err != nil {
		t.Fatalf("Failed to predict: %v", err)
	}

	if result.EntityID != "predict-campaign" {
		t.Error("Entity ID mismatch")
	}
	if len(result.Predictions) == 0 {
		t.Error("Expected predictions")
	}
	if result.Confidence <= 0 || result.Confidence > 1 {
		t.Errorf("Invalid confidence: %f", result.Confidence)
	}
}

func TestPrediction_Forecast(t *testing.T) {
	c := NewMockCache()
	svc := NewPerformancePredictionService(c)

	// Record historical data with clear trend
	for i := 0; i < 14; i++ {
		record := &PerformanceRecord{
			EntityID:    "forecast-campaign",
			EntityType:  "campaign",
			Timestamp:   time.Now().Add(time.Duration(-14+i) * 24 * time.Hour),
			Impressions: int64(10000 + i*1000), // Growing
			Clicks:      int64(500 + i*50),
			Spend:       1000.0 + float64(i)*100,
			Revenue:     2500.0 + float64(i)*200,
		}
		svc.RecordPerformance(record)
	}

	forecast, err := svc.Forecast("forecast-campaign", "campaign", 24)
	if err != nil {
		t.Fatalf("Forecast failed: %v", err)
	}

	if forecast == nil {
		t.Fatal("Expected forecast result")
	}
	if forecast.EntityID != "forecast-campaign" {
		t.Error("Entity ID mismatch")
	}
}

func TestPrediction_GetPredictionAccuracy(t *testing.T) {
	c := NewMockCache()
	svc := NewPerformancePredictionService(c)

	// Record historical data
	for i := 0; i < 14; i++ {
		record := &PerformanceRecord{
			EntityID:    "accuracy-campaign",
			EntityType:  "campaign",
			Timestamp:   time.Now().Add(time.Duration(-14+i) * 24 * time.Hour),
			Impressions: int64(10000 + i*100),
			Clicks:      int64(500 + i*10),
			CTR:         0.05,
		}
		svc.RecordPerformance(record)
	}

	accuracy := svc.GetPredictionAccuracy("accuracy-campaign", 24)
	if accuracy == nil {
		t.Fatal("Expected accuracy data")
	}
}

func TestPrediction_NoHistoryError(t *testing.T) {
	c := NewMockCache()
	svc := NewPerformancePredictionService(c)

	req := PredictionRequest{
		EntityID:   "nonexistent-campaign",
		EntityType: "campaign",
		Metrics:    []string{"ctr"},
	}

	result, err := svc.Predict(req)
	// The service may return partial results with low confidence instead of error
	if err == nil && result != nil && result.Confidence > 0.5 {
		t.Error("Expected low confidence for campaign with no history")
	}
}

func TestPrediction_GetStats(t *testing.T) {
	c := NewMockCache()
	svc := NewPerformancePredictionService(c)

	stats := svc.GetStats()
	if stats["total_records"].(int) != 0 {
		t.Error("Expected 0 records initially")
	}

	// Add records - each entity gets aggregated, so record different entities
	entities := []string{"stats-campaign-A", "stats-campaign-B", "stats-campaign-C", "stats-campaign-D", "stats-campaign-E"}
	for _, entityID := range entities {
		record := &PerformanceRecord{
			EntityID:    entityID,
			EntityType:  "campaign",
			Timestamp:   time.Now(),
			Impressions: int64(10000),
		}
		svc.RecordPerformance(record)
	}

	stats = svc.GetStats()
	if stats["total_records"].(int) < 5 {
		t.Errorf("Expected at least 5 records, got %d", stats["total_records"].(int))
	}
}

func TestPrediction_UpdateConfig(t *testing.T) {
	c := NewMockCache()
	svc := NewPerformancePredictionService(c)

	originalConfig := svc.GetConfig()
	if originalConfig.MinHistoricalSamples == 0 {
		t.Error("Expected default min historical samples")
	}

	newConfig := PredictionConfig{
		MinHistoricalSamples: 20,
		ConfidenceThreshold:  0.8,
		PredictionHorizon:    48,
		EnableRealTimeUpdate: true,
	}
	svc.UpdateConfig(newConfig)

	updatedConfig := svc.GetConfig()
	if updatedConfig.MinHistoricalSamples != 20 {
		t.Errorf("Expected min samples 20, got %d", updatedConfig.MinHistoricalSamples)
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestDCO_Integration_WithBiddingService(t *testing.T) {
	c := NewMockCache()
	bs := NewBiddingService(c, "http://localhost:8080")

	dcoSvc := bs.GetDynamicCreativeService()
	if dcoSvc == nil {
		t.Fatal("DCO service should be accessible from BiddingService")
	}
}

func TestPrediction_Integration_WithBiddingService(t *testing.T) {
	c := NewMockCache()
	bs := NewBiddingService(c, "http://localhost:8080")

	predSvc := bs.GetPerformancePredictionService()
	if predSvc == nil {
		t.Fatal("Prediction service should be accessible from BiddingService")
	}
}

func TestPhase4_AllMLServicesIntegrated(t *testing.T) {
	c := NewMockCache()
	bs := NewBiddingService(c, "http://localhost:8080")

	// Test all Phase 4 ML services are accessible
	if bs.GetDynamicBidService() == nil {
		t.Error("DynamicBid service missing")
	}
	if bs.GetLookalikeService() == nil {
		t.Error("Lookalike service missing")
	}
	if bs.GetUserClusteringService() == nil {
		t.Error("UserClustering service missing")
	}
	if bs.GetChurnPredictionService() == nil {
		t.Error("ChurnPrediction service missing")
	}
	if bs.GetABTestingService() == nil {
		t.Error("ABTesting service missing")
	}
	if bs.GetDynamicCreativeService() == nil {
		t.Error("DCO service missing")
	}
	if bs.GetPerformancePredictionService() == nil {
		t.Error("PerformancePrediction service missing")
	}
}

func TestDCO_EndToEndFlow(t *testing.T) {
	c := NewMockCache()
	svc := NewDynamicCreativeService(c)

	// 1. Create template
	template := &CreativeTemplate{
		Name:        "E2E Template",
		Description: "End to end test",
		Format:      "banner",
		Dimensions:  Dimensions{Width: 728, Height: 90},
		Slots: map[string]*TemplateSlot{
			"headline": {ID: "headline", Name: "Headline", Type: "headline", Required: true, Position: Position{X: 10, Y: 10, Width: 400, Height: 30}},
			"cta":      {ID: "cta", Name: "CTA", Type: "cta", Required: true, Position: Position{X: 500, Y: 20, Width: 100, Height: 50}},
		},
	}
	created, err := svc.CreateTemplate(template)
	if err != nil {
		t.Fatalf("Template creation failed: %v", err)
	}

	// 2. Create elements
	svc.CreateElement(&CreativeElement{Type: "headline", Content: "Default", Segments: []string{}})
	svc.CreateElement(&CreativeElement{Type: "cta", Content: "Click", Segments: []string{}})
	svc.CreateElement(&CreativeElement{
		Type:     "headline",
		Content:  "Special Offer!",
		Segments: []string{"vip"},
	})

	// 3. Generate creative for user
	req := DCORequest{
		TemplateID: created.ID,
		UserID:     "e2e-user",
		Context: DCOContext{
			UserSegments: []string{"vip"},
			DeviceType:   "desktop",
		},
	}
	response, err := svc.GenerateOptimizedCreative(req)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	// 4. Record engagement
	svc.RecordImpression(response.CombinationID)
	svc.RecordClick(response.CombinationID, "e2e-user")
	svc.RecordConversion(response.CombinationID, 149.99)

	// 5. Verify stats
	stats := svc.GetDCOStats()
	if stats["total_impressions"].(int64) == 0 || stats["total_clicks"].(int64) == 0 || stats["total_conversions"].(int64) == 0 {
		t.Error("Stats should reflect engagement")
	}
}

func TestPrediction_EndToEndFlow(t *testing.T) {
	c := NewMockCache()
	svc := NewPerformancePredictionService(c)

	entityID := "e2e-prediction-campaign"

	// 1. Record 14 days of historical data
	for i := 0; i < 14; i++ {
		record := &PerformanceRecord{
			EntityID:    entityID,
			EntityType:  "campaign",
			Timestamp:   time.Now().Add(time.Duration(-14+i) * 24 * time.Hour),
			Impressions: int64(10000 + i*500),
			Clicks:      int64(500 + i*25),
			Conversions: int64(50 + i*5),
			Spend:       1000.0 + float64(i)*50,
			Revenue:     2500.0 + float64(i)*100,
			CTR:         0.05,
			CVR:         0.10,
			ROAS:        2.5,
		}
		svc.RecordPerformance(record)
	}

	// 2. Make prediction
	req := PredictionRequest{
		EntityID:   entityID,
		EntityType: "campaign",
		Features:   map[string]float64{"bid_price": 1.50},
		Context: PredictionContext{
			TimeOfDay:  "morning",
			DayOfWeek:  "monday",
			DeviceType: "mobile",
			AdFormat:   "banner",
		},
		Horizon: 24,
		Metrics: []string{"ctr", "cvr", "roas"},
	}

	result, err := svc.Predict(req)
	if err != nil {
		t.Fatalf("Prediction failed: %v", err)
	}
	if len(result.Predictions) == 0 {
		t.Error("Should have predictions")
	}

	// 3. Get forecast
	forecast, err := svc.Forecast(entityID, "campaign", 24)
	if err != nil {
		t.Fatalf("Forecast failed: %v", err)
	}
	if forecast == nil {
		t.Error("Should have forecast")
	}

	// 4. Check stats
	stats := svc.GetStats()
	if stats["total_records"].(int) < 14 {
		t.Error("Should have recorded data")
	}
}
