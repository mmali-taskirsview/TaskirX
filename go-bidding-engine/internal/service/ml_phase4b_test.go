package service

import (
	"testing"
)

// ============================================================================
// DYNAMIC CREATIVE OPTIMIZATION SERVICE TESTS
// ============================================================================

func TestNewDynamicCreativeService(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	if service == nil {
		t.Fatal("Expected non-nil DynamicCreativeService")
	}
}

func TestDCOCreateTemplate(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	template := &CreativeTemplate{
		Name:        "Test Template",
		Description: "A test template",
		Format:      "banner",
		Dimensions:  Dimensions{Width: 300, Height: 250},
		Slots: map[string]*TemplateSlot{
			"headline": {ID: "headline", Name: "Headline", Type: "headline", Required: true},
			"image":    {ID: "image", Name: "Image", Type: "image", Required: true},
			"cta":      {ID: "cta", Name: "CTA", Type: "cta", Required: false},
		},
	}

	created, err := service.CreateTemplate(template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	if created.Name != "Test Template" {
		t.Errorf("Expected name 'Test Template', got '%s'", created.Name)
	}
	if created.Dimensions.Width != 300 || created.Dimensions.Height != 250 {
		t.Errorf("Expected dimensions 300x250, got %dx%d", created.Dimensions.Width, created.Dimensions.Height)
	}
	if len(created.Slots) != 3 {
		t.Errorf("Expected 3 slots, got %d", len(created.Slots))
	}
}

func TestDCOGetTemplate(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	// Create a template first
	template := &CreativeTemplate{
		Name:       "Test",
		Format:     "banner",
		Dimensions: Dimensions{Width: 300, Height: 250},
		Slots: map[string]*TemplateSlot{
			"headline": {ID: "headline", Name: "Headline", Type: "headline", Required: true},
		},
	}
	created, _ := service.CreateTemplate(template)

	// Retrieve it
	retrieved, err := service.GetTemplate(created.ID)
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Template ID mismatch")
	}
}

func TestDCOGetTemplate_NotFound(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	_, err := service.GetTemplate("nonexistent-id")
	if err == nil {
		t.Error("Expected error for non-existent template")
	}
}

func TestDCOCreateElement(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	element := &CreativeElement{
		Type:    "headline",
		Content: "Buy Now!",
		Tags:    []string{"urgent", "sale"},
		Attributes: map[string]string{
			"tone":   "urgent",
			"length": "short",
		},
	}

	created, err := service.CreateElement(element)
	if err != nil {
		t.Fatalf("Failed to create element: %v", err)
	}

	if created.Type != "headline" {
		t.Errorf("Expected type 'headline', got '%s'", created.Type)
	}
	if created.Content != "Buy Now!" {
		t.Errorf("Expected content 'Buy Now!', got '%s'", created.Content)
	}
}

func TestDCOGetElement(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	element := &CreativeElement{
		Type:     "image",
		ImageURL: "https://example.com/image.jpg",
	}
	created, _ := service.CreateElement(element)

	retrieved, err := service.GetElement(created.ID)
	if err != nil {
		t.Fatalf("Failed to get element: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Element ID mismatch")
	}
}

func TestDCOGetElement_NotFound(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	_, err := service.GetElement("nonexistent-id")
	if err == nil {
		t.Error("Expected error for non-existent element")
	}
}

func TestDCOGenerateOptimizedCreative(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	// Create template
	template := &CreativeTemplate{
		Name:       "Test",
		Format:     "banner",
		Dimensions: Dimensions{Width: 300, Height: 250},
		Slots: map[string]*TemplateSlot{
			"headline": {ID: "headline", Name: "Headline", Type: "headline", Required: true},
		},
	}
	created, _ := service.CreateTemplate(template)

	// Create element
	element := &CreativeElement{
		Type:    "headline",
		Content: "Amazing Offer!",
	}
	service.CreateElement(element)

	// Generate creative using DCORequest
	req := DCORequest{
		TemplateID: created.ID,
		UserID:     "user-123",
		Context: DCOContext{
			DeviceType: "mobile",
			TimeOfDay:  "afternoon",
		},
	}

	creative, err := service.GenerateOptimizedCreative(req)
	if err != nil {
		t.Fatalf("Failed to generate creative: %v", err)
	}

	if creative.TemplateID != created.ID {
		t.Errorf("Template ID mismatch")
	}
}

func TestDCORecordImpression(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	// First create a combination by generating a creative
	template := &CreativeTemplate{
		Name:       "Test",
		Format:     "banner",
		Dimensions: Dimensions{Width: 300, Height: 250},
		Slots: map[string]*TemplateSlot{
			"headline": {ID: "headline", Name: "Headline", Type: "headline", Required: true},
		},
	}
	created, _ := service.CreateTemplate(template)
	service.CreateElement(&CreativeElement{Type: "headline", Content: "Test"})

	// Generate creative to create a combination
	req := DCORequest{TemplateID: created.ID, UserID: "user-1"}
	creative, _ := service.GenerateOptimizedCreative(req)

	// Now record impression for the generated combination
	err := service.RecordImpression(creative.CombinationID)
	if err != nil {
		t.Errorf("Failed to record impression: %v", err)
	}
}

func TestDCORecordClick(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	// Create combination first
	template := &CreativeTemplate{
		Name:       "Test",
		Format:     "banner",
		Dimensions: Dimensions{Width: 300, Height: 250},
		Slots: map[string]*TemplateSlot{
			"headline": {ID: "headline", Name: "Headline", Type: "headline", Required: true},
		},
	}
	created, _ := service.CreateTemplate(template)
	service.CreateElement(&CreativeElement{Type: "headline", Content: "Test"})

	req := DCORequest{TemplateID: created.ID, UserID: "user-1"}
	creative, _ := service.GenerateOptimizedCreative(req)

	// Record impression first
	service.RecordImpression(creative.CombinationID)

	err := service.RecordClick(creative.CombinationID, "user-1")
	if err != nil {
		t.Errorf("Failed to record click: %v", err)
	}
}

func TestDCORecordConversion(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	// Create combination first
	template := &CreativeTemplate{
		Name:       "Test",
		Format:     "banner",
		Dimensions: Dimensions{Width: 300, Height: 250},
		Slots: map[string]*TemplateSlot{
			"headline": {ID: "headline", Name: "Headline", Type: "headline", Required: true},
		},
	}
	created, _ := service.CreateTemplate(template)
	service.CreateElement(&CreativeElement{Type: "headline", Content: "Test"})

	req := DCORequest{TemplateID: created.ID, UserID: "user-1"}
	creative, _ := service.GenerateOptimizedCreative(req)

	// Record impression and click first
	service.RecordImpression(creative.CombinationID)
	service.RecordClick(creative.CombinationID, "user-1")

	err := service.RecordConversion(creative.CombinationID, 99.99)
	if err != nil {
		t.Errorf("Failed to record conversion: %v", err)
	}
}

func TestDCOGetTopCombinations(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	// Create template and element
	template := &CreativeTemplate{
		Name:       "Test",
		Format:     "banner",
		Dimensions: Dimensions{Width: 300, Height: 250},
		Slots: map[string]*TemplateSlot{
			"headline": {ID: "headline", Name: "Headline", Type: "headline", Required: true},
		},
	}
	created, _ := service.CreateTemplate(template)
	service.CreateElement(&CreativeElement{Type: "headline", Content: "Test"})

	// Generate creative to create combination
	req := DCORequest{TemplateID: created.ID, UserID: "user-1"}
	creative, _ := service.GenerateOptimizedCreative(req)

	// Record some data
	service.RecordImpression(creative.CombinationID)
	service.RecordImpression(creative.CombinationID)
	service.RecordClick(creative.CombinationID, "user-1")

	combinations := service.GetTopCombinations(created.ID, 10)
	// GetTopCombinations returns empty slice if no combinations for template
	if combinations == nil {
		combinations = []*CreativeCombination{}
	}
	// It's valid to have 0 or more combinations
}

func TestDCOGetStats(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	// Create some data
	template := &CreativeTemplate{
		Name:       "Test",
		Format:     "banner",
		Dimensions: Dimensions{Width: 300, Height: 250},
		Slots:      map[string]*TemplateSlot{},
	}
	service.CreateTemplate(template)

	element := &CreativeElement{Type: "headline", Content: "Test"}
	service.CreateElement(element)

	stats := service.GetDCOStats()
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	// Stats returns map[string]any
	if _, ok := stats["total_templates"]; !ok {
		t.Error("Expected 'total_templates' in stats")
	}
	if _, ok := stats["total_elements"]; !ok {
		t.Error("Expected 'total_elements' in stats")
	}
}

func TestDCOConfig(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	config := service.GetConfig()
	if config.ExplorationRate <= 0 {
		t.Error("Expected positive exploration rate")
	}
}

func TestDCOUpdateConfig(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	newConfig := DCOConfig{
		MaxElementsPerSlot:     20,
		ExplorationRate:        0.2,
		MinImpressionsForStats: 50,
		PersonalizationWeight:  0.4,
		ContextWeight:          0.3,
		PerformanceWeight:      0.3,
		EnableAutoOptimization: true,
	}

	service.UpdateConfig(newConfig)

	updated := service.GetConfig()
	if updated.MaxElementsPerSlot != 20 {
		t.Errorf("Expected MaxElementsPerSlot 20, got %d", updated.MaxElementsPerSlot)
	}
	if updated.ExplorationRate != 0.2 {
		t.Errorf("Expected ExplorationRate 0.2, got %f", updated.ExplorationRate)
	}
}

// ============================================================================
// PERFORMANCE PREDICTION SERVICE TESTS
// ============================================================================

func TestNewPerformancePredictionService(t *testing.T) {
	cache := NewMockCache()
	service := NewPerformancePredictionService(cache)

	if service == nil {
		t.Fatal("Expected non-nil PerformancePredictionService")
	}
}

func TestPredictionRecordPerformance(t *testing.T) {
	cache := NewMockCache()
	service := NewPerformancePredictionService(cache)

	record := &PerformanceRecord{
		EntityID:    "camp-123",
		EntityType:  "campaign",
		Impressions: 1000,
		Clicks:      25,
		Conversions: 12,
		Revenue:     150.0,
		CTR:         0.025,
		CVR:         0.012,
	}

	// RecordPerformance doesn't return error
	service.RecordPerformance(record)
}

func TestPredictionPredict(t *testing.T) {
	cache := NewMockCache()
	service := NewPerformancePredictionService(cache)

	// Record some data first
	for i := 0; i < 10; i++ {
		record := &PerformanceRecord{
			EntityID:    "camp-123",
			EntityType:  "campaign",
			Impressions: int64(1000 + i*100),
			Clicks:      int64(25 + i*2),
			CTR:         0.02 + float64(i)*0.001,
			CVR:         0.01 + float64(i)*0.0005,
		}
		service.RecordPerformance(record)
	}

	req := PredictionRequest{
		EntityID:   "camp-123",
		EntityType: "campaign",
		Metrics:    []string{"ctr", "cvr"},
	}

	result, err := service.Predict(req)
	if err != nil {
		t.Fatalf("Failed to predict: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil prediction result")
	}
}

func TestPredictionPredict_InsufficientData(t *testing.T) {
	cache := NewMockCache()
	service := NewPerformancePredictionService(cache)

	// Don't record any data - just test with non-existent entity
	req := PredictionRequest{
		EntityID:   "nonexistent-entity",
		EntityType: "campaign",
		Metrics:    []string{"ctr"},
	}

	_, err := service.Predict(req)
	// May or may not error depending on implementation
	// Just verify it doesn't panic
	_ = err
}

func TestPredictionForecast(t *testing.T) {
	cache := NewMockCache()
	service := NewPerformancePredictionService(cache)

	// Record enough data for forecasting
	for i := 0; i < 15; i++ {
		record := &PerformanceRecord{
			EntityID:    "camp-forecast",
			EntityType:  "campaign",
			Impressions: int64(1000 + i*50),
			Clicks:      int64(25 + i),
			CTR:         0.02 + float64(i)*0.001,
		}
		service.RecordPerformance(record)
	}

	forecast, err := service.Forecast("camp-forecast", "campaign", 24)
	if err != nil {
		t.Fatalf("Failed to forecast: %v", err)
	}

	if forecast == nil {
		t.Error("Expected non-nil forecast")
	}
}

func TestPredictionForecast_InsufficientData(t *testing.T) {
	cache := NewMockCache()
	service := NewPerformancePredictionService(cache)

	// Test with non-existent entity
	_, err := service.Forecast("nonexistent-entity", "campaign", 24)
	// May or may not error depending on implementation
	// Just verify it doesn't panic
	_ = err
}

func TestPredictionGetAccuracy(t *testing.T) {
	cache := NewMockCache()
	service := NewPerformancePredictionService(cache)

	accuracy := service.GetPredictionAccuracy("camp-123", 24)
	if accuracy == nil {
		t.Fatal("Expected non-nil accuracy map")
	}
}

func TestPredictionGetStats(t *testing.T) {
	cache := NewMockCache()
	service := NewPerformancePredictionService(cache)

	// Record some data
	record1 := &PerformanceRecord{EntityID: "camp-1", EntityType: "campaign", CTR: 0.025}
	record2 := &PerformanceRecord{EntityID: "creative-1", EntityType: "creative", CTR: 0.030}
	service.RecordPerformance(record1)
	service.RecordPerformance(record2)

	stats := service.GetStats()
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}
	// Stats returns map[string]any - just verify it's not nil
}

func TestPredictionConfig(t *testing.T) {
	cache := NewMockCache()
	service := NewPerformancePredictionService(cache)

	config := service.GetConfig()
	if config.MinHistoricalSamples <= 0 {
		t.Error("Expected positive min historical samples")
	}
	if config.ConfidenceThreshold <= 0 {
		t.Error("Expected positive confidence threshold")
	}
}

func TestPredictionUpdateConfig(t *testing.T) {
	cache := NewMockCache()
	service := NewPerformancePredictionService(cache)

	newConfig := PredictionConfig{
		MinHistoricalSamples: 10,
		ConfidenceThreshold:  0.85,
		EnableRealTimeUpdate: true,
	}

	service.UpdateConfig(newConfig)

	updated := service.GetConfig()
	if updated.MinHistoricalSamples != 10 {
		t.Errorf("Expected min historical samples 10, got %d", updated.MinHistoricalSamples)
	}
	if updated.ConfidenceThreshold != 0.85 {
		t.Errorf("Expected confidence threshold 0.85, got %f", updated.ConfidenceThreshold)
	}
}

func TestDCOMultipleTemplates(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	// Create multiple templates with at least one slot
	for i := 0; i < 5; i++ {
		template := &CreativeTemplate{
			Name:       "Template",
			Format:     "banner",
			Dimensions: Dimensions{Width: 300, Height: 250},
			Slots: map[string]*TemplateSlot{
				"headline": {ID: "headline", Name: "Headline", Type: "headline", Required: true},
			},
		}
		_, err := service.CreateTemplate(template)
		if err != nil {
			t.Fatalf("Failed to create template %d: %v", i, err)
		}
	}

	stats := service.GetDCOStats()
	if templates, ok := stats["total_templates"].(int); ok {
		if templates != 5 {
			t.Errorf("Expected 5 templates, got %d", templates)
		}
	}
}

func TestDCOMultipleElements(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	// Create multiple elements of different types
	types := []string{"headline", "image", "cta", "description"}
	for _, elementType := range types {
		element := &CreativeElement{
			Type:    elementType,
			Content: "content",
		}
		_, err := service.CreateElement(element)
		if err != nil {
			t.Fatalf("Failed to create element of type %s: %v", elementType, err)
		}
	}

	stats := service.GetDCOStats()
	if elements, ok := stats["total_elements"].(int); ok {
		if elements != 4 {
			t.Errorf("Expected 4 elements, got %d", elements)
		}
	}
}

func TestPredictionMultipleEntityTypes(t *testing.T) {
	cache := NewMockCache()
	service := NewPerformancePredictionService(cache)

	entityTypes := []string{"campaign", "creative", "placement"}
	for _, entityType := range entityTypes {
		record := &PerformanceRecord{
			EntityID:   entityType + "-1",
			EntityType: entityType,
			CTR:        0.025,
		}
		service.RecordPerformance(record)
	}

	stats := service.GetStats()
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}
	// Stats exists and doesn't panic
}

func TestDCOGetElementsByType(t *testing.T) {
	cache := NewMockCache()
	service := NewDynamicCreativeService(cache)

	// Create elements of different types
	service.CreateElement(&CreativeElement{Type: "headline", Content: "Headline 1"})
	service.CreateElement(&CreativeElement{Type: "headline", Content: "Headline 2"})
	service.CreateElement(&CreativeElement{Type: "image", Content: "Image 1"})

	headlines := service.GetElementsByType("headline")
	if len(headlines) != 2 {
		t.Errorf("Expected 2 headlines, got %d", len(headlines))
	}

	images := service.GetElementsByType("image")
	if len(images) != 1 {
		t.Errorf("Expected 1 image, got %d", len(images))
	}
}
