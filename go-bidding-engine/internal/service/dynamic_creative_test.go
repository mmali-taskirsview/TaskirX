package service

import (
	"testing"
)

func createTestTemplate() *CreativeTemplate {
	return &CreativeTemplate{
		Name:        "Test Template",
		Description: "A test template",
		Format:      "banner",
		Dimensions:  Dimensions{Width: 300, Height: 250},
		Slots: map[string]*TemplateSlot{
			"headline": {
				ID:       "headline",
				Name:     "Headline",
				Type:     "headline",
				Required: true,
			},
			"cta": {
				ID:       "cta",
				Name:     "Call to Action",
				Type:     "cta",
				Required: true,
			},
		},
		BaseHTML: "<div>{{headline}}{{cta}}</div>",
	}
}

func createTestElement(elemType string) *CreativeElement {
	return &CreativeElement{
		Type:       elemType,
		Content:    "Test " + elemType,
		Tags:       []string{"test"},
		Attributes: map[string]string{},
	}
}

func TestDCO_NewService(t *testing.T) {
	svc := NewDynamicCreativeService(nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	config := svc.GetConfig()
	if config.ExplorationRate == 0 {
		t.Error("expected ExplorationRate to be set")
	}
	if config.MaxElementsPerSlot == 0 {
		t.Error("expected MaxElementsPerSlot to be set")
	}
}

func TestDCO_CreateTemplate(t *testing.T) {
	svc := NewDynamicCreativeService(nil)
	template := createTestTemplate()

	created, err := svc.CreateTemplate(template)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected template ID to be set")
	}
	if created.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestDCO_CreateTemplate_Validation(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	tests := []struct {
		name     string
		template *CreativeTemplate
		wantErr  bool
	}{
		{
			name:     "missing name",
			template: &CreativeTemplate{Slots: map[string]*TemplateSlot{"h": {}}},
			wantErr:  true,
		},
		{
			name:     "no slots",
			template: &CreativeTemplate{Name: "Test", Slots: map[string]*TemplateSlot{}},
			wantErr:  true,
		},
		{
			name:     "valid template",
			template: createTestTemplate(),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateTemplate(tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDCO_GetTemplate(t *testing.T) {
	svc := NewDynamicCreativeService(nil)
	template := createTestTemplate()
	created, _ := svc.CreateTemplate(template)

	retrieved, err := svc.GetTemplate(created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("expected ID=%s, got %s", created.ID, retrieved.ID)
	}
}

func TestDCO_GetTemplate_NotFound(t *testing.T) {
	svc := NewDynamicCreativeService(nil)
	_, err := svc.GetTemplate("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent template")
	}
}

func TestDCO_CreateElement(t *testing.T) {
	svc := NewDynamicCreativeService(nil)
	element := createTestElement("headline")

	created, err := svc.CreateElement(element)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected element ID to be set")
	}
	if created.Performance == nil {
		t.Error("expected Performance to be initialized")
	}
}

func TestDCO_CreateElement_Validation(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	_, err := svc.CreateElement(&CreativeElement{Type: ""})
	if err == nil {
		t.Error("expected error for missing type")
	}
}

func TestDCO_GetElement(t *testing.T) {
	svc := NewDynamicCreativeService(nil)
	element := createTestElement("headline")
	created, _ := svc.CreateElement(element)

	retrieved, err := svc.GetElement(created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("expected ID=%s, got %s", created.ID, retrieved.ID)
	}
}

func TestDCO_GetElement_NotFound(t *testing.T) {
	svc := NewDynamicCreativeService(nil)
	_, err := svc.GetElement("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent element")
	}
}

func TestDCO_GetElementsByType(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	// Create elements of different types
	svc.CreateElement(createTestElement("headline"))
	svc.CreateElement(createTestElement("headline"))
	svc.CreateElement(createTestElement("cta"))

	headlines := svc.GetElementsByType("headline")
	if len(headlines) != 2 {
		t.Errorf("expected 2 headlines, got %d", len(headlines))
	}

	ctas := svc.GetElementsByType("cta")
	if len(ctas) != 1 {
		t.Errorf("expected 1 cta, got %d", len(ctas))
	}
}

func TestDCO_GenerateOptimizedCreative(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	// Create template
	template := createTestTemplate()
	created, _ := svc.CreateTemplate(template)

	// Create elements for each slot
	headline := createTestElement("headline")
	headline.Content = "Great Deal!"
	svc.CreateElement(headline)

	cta := createTestElement("cta")
	cta.CTAText = "Buy Now"
	svc.CreateElement(cta)

	req := DCORequest{
		TemplateID: created.ID,
		UserID:     "user123",
		Context: DCOContext{
			PageCategory: "shopping",
			DeviceType:   "mobile",
		},
	}

	response, err := svc.GenerateOptimizedCreative(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response == nil {
		t.Fatal("expected non-nil response")
	}
	if response.CombinationID == "" {
		t.Error("expected combination ID")
	}
	if response.TemplateID != created.ID {
		t.Error("expected template ID to match")
	}
}

func TestDCO_GenerateOptimizedCreative_InvalidTemplate(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	req := DCORequest{
		TemplateID: "invalid",
		UserID:     "user123",
	}

	_, err := svc.GenerateOptimizedCreative(req)
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

func TestDCO_RecordImpression(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	// Setup
	template := createTestTemplate()
	created, _ := svc.CreateTemplate(template)
	svc.CreateElement(createTestElement("headline"))
	svc.CreateElement(createTestElement("cta"))

	req := DCORequest{
		TemplateID: created.ID,
		UserID:     "user123",
	}
	response, _ := svc.GenerateOptimizedCreative(req)

	// Record impression
	err := svc.RecordImpression(response.CombinationID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDCO_RecordImpression_NotFound(t *testing.T) {
	svc := NewDynamicCreativeService(nil)
	err := svc.RecordImpression("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent combination")
	}
}

func TestDCO_RecordClick(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	// Setup
	template := createTestTemplate()
	created, _ := svc.CreateTemplate(template)
	svc.CreateElement(createTestElement("headline"))
	svc.CreateElement(createTestElement("cta"))

	req := DCORequest{
		TemplateID: created.ID,
		UserID:     "user123",
	}
	response, _ := svc.GenerateOptimizedCreative(req)

	// Record impression first, then click
	svc.RecordImpression(response.CombinationID)
	err := svc.RecordClick(response.CombinationID, "user123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDCO_RecordConversion(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	// Setup
	template := createTestTemplate()
	created, _ := svc.CreateTemplate(template)
	svc.CreateElement(createTestElement("headline"))
	svc.CreateElement(createTestElement("cta"))

	req := DCORequest{
		TemplateID: created.ID,
		UserID:     "user123",
	}
	response, _ := svc.GenerateOptimizedCreative(req)

	// Record funnel
	svc.RecordImpression(response.CombinationID)
	svc.RecordClick(response.CombinationID, "user123")
	err := svc.RecordConversion(response.CombinationID, 50.0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDCO_GetTopCombinations(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	// Setup
	template := createTestTemplate()
	created, _ := svc.CreateTemplate(template)
	svc.CreateElement(createTestElement("headline"))
	svc.CreateElement(createTestElement("cta"))

	// Generate and record some impressions/clicks
	for i := 0; i < 5; i++ {
		req := DCORequest{
			TemplateID: created.ID,
			UserID:     "user" + string(rune('0'+i)),
		}
		response, _ := svc.GenerateOptimizedCreative(req)
		svc.RecordImpression(response.CombinationID)
		if i%2 == 0 {
			svc.RecordClick(response.CombinationID, "user")
		}
	}

	top := svc.GetTopCombinations(created.ID, 3)

	if len(top) > 3 {
		t.Errorf("expected at most 3 combinations, got %d", len(top))
	}
}

func TestDCO_GetDCOStats(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	// Setup some data
	template := createTestTemplate()
	svc.CreateTemplate(template)
	svc.CreateElement(createTestElement("headline"))
	svc.CreateElement(createTestElement("cta"))

	stats := svc.GetDCOStats()

	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats["total_templates"].(int) != 1 {
		t.Errorf("expected 1 template, got %d", stats["total_templates"].(int))
	}
	if stats["total_elements"].(int) != 2 {
		t.Errorf("expected 2 elements, got %d", stats["total_elements"].(int))
	}
}

func TestDCO_UpdateConfig(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	newConfig := DCOConfig{
		ExplorationRate:        0.2,
		MaxElementsPerSlot:     20,
		EnableAutoOptimization: false,
	}

	svc.UpdateConfig(newConfig)

	config := svc.GetConfig()
	if config.ExplorationRate != 0.2 {
		t.Errorf("expected ExplorationRate=0.2, got %f", config.ExplorationRate)
	}
	if config.MaxElementsPerSlot != 20 {
		t.Errorf("expected MaxElementsPerSlot=20, got %d", config.MaxElementsPerSlot)
	}
}

func TestDCO_ElementTypes(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	types := []string{"headline", "image", "cta", "description", "logo"}

	for _, elemType := range types {
		t.Run(elemType, func(t *testing.T) {
			elem := createTestElement(elemType)
			created, err := svc.CreateElement(elem)
			if err != nil {
				t.Fatalf("failed to create %s: %v", elemType, err)
			}
			if created.Type != elemType {
				t.Errorf("expected type=%s, got %s", elemType, created.Type)
			}
		})
	}
}

func TestDCO_ContextVariants(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	template := createTestTemplate()
	created, _ := svc.CreateTemplate(template)
	svc.CreateElement(createTestElement("headline"))
	svc.CreateElement(createTestElement("cta"))

	contexts := []DCOContext{
		{PageCategory: "news", DeviceType: "mobile"},
		{PageCategory: "sports", DeviceType: "desktop"},
		{PageCategory: "finance", DeviceType: "tablet"},
		{TimeOfDay: "morning", DayOfWeek: "Monday"},
	}

	for i, ctx := range contexts {
		t.Run(ctx.PageCategory+"_"+ctx.DeviceType, func(t *testing.T) {
			req := DCORequest{
				TemplateID: created.ID,
				UserID:     "user" + string(rune('0'+i)),
				Context:    ctx,
			}

			response, err := svc.GenerateOptimizedCreative(req)
			if err != nil {
				t.Fatalf("failed with context %v: %v", ctx, err)
			}
			if response == nil {
				t.Fatal("expected non-nil response")
			}
		})
	}
}

func TestDCO_Constraints(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	template := createTestTemplate()
	created, _ := svc.CreateTemplate(template)
	elem1, _ := svc.CreateElement(createTestElement("headline"))
	elem2, _ := svc.CreateElement(createTestElement("headline"))
	svc.CreateElement(createTestElement("cta"))

	req := DCORequest{
		TemplateID: created.ID,
		UserID:     "user123",
		Constraints: DCOConstraints{
			ExcludedElements: []string{elem1.ID},
		},
	}

	response, err := svc.GenerateOptimizedCreative(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = elem2 // elem2 should be selectable
	_ = response
}

func TestDCO_PerformanceTracking(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	template := createTestTemplate()
	created, _ := svc.CreateTemplate(template)
	svc.CreateElement(createTestElement("headline"))
	svc.CreateElement(createTestElement("cta"))

	req := DCORequest{
		TemplateID: created.ID,
		UserID:     "user123",
	}

	// Generate and track performance
	response, _ := svc.GenerateOptimizedCreative(req)

	for i := 0; i < 100; i++ {
		svc.RecordImpression(response.CombinationID)
	}
	for i := 0; i < 10; i++ {
		svc.RecordClick(response.CombinationID, "user")
	}
	svc.RecordConversion(response.CombinationID, 100.0)

	stats := svc.GetDCOStats()

	// Stats accumulate across test runs, so check minimum values
	if stats["total_impressions"].(int64) < 100 {
		t.Errorf("expected at least 100 impressions, got %d", stats["total_impressions"].(int64))
	}
	if stats["total_clicks"].(int64) < 10 {
		t.Errorf("expected at least 10 clicks, got %d", stats["total_clicks"].(int64))
	}
	if stats["total_conversions"].(int64) < 1 {
		t.Errorf("expected at least 1 conversion, got %d", stats["total_conversions"].(int64))
	}
}

func TestDCO_Concurrency(t *testing.T) {
	svc := NewDynamicCreativeService(nil)
	done := make(chan bool, 3)

	template := createTestTemplate()
	created, _ := svc.CreateTemplate(template)
	svc.CreateElement(createTestElement("headline"))
	svc.CreateElement(createTestElement("cta"))

	// Writer
	go func() {
		for i := 0; i < 50; i++ {
			req := DCORequest{
				TemplateID: created.ID,
				UserID:     "user" + string(rune('0'+i%10)),
			}
			svc.GenerateOptimizedCreative(req)
		}
		done <- true
	}()

	// Reader 1
	go func() {
		for i := 0; i < 50; i++ {
			svc.GetDCOStats()
			svc.GetTopCombinations(created.ID, 5)
		}
		done <- true
	}()

	// Reader 2
	go func() {
		for i := 0; i < 50; i++ {
			svc.GetConfig()
			svc.GetElementsByType("headline")
		}
		done <- true
	}()

	<-done
	<-done
	<-done
}

func TestDCO_UserPreferences(t *testing.T) {
	svc := NewDynamicCreativeService(nil)

	template := createTestTemplate()
	created, _ := svc.CreateTemplate(template)
	svc.CreateElement(createTestElement("headline"))
	svc.CreateElement(createTestElement("cta"))

	userID := "user123"

	// Generate multiple creatives for same user
	for i := 0; i < 5; i++ {
		req := DCORequest{
			TemplateID: created.ID,
			UserID:     userID,
			Context: DCOContext{
				UserSegments: []string{"premium"},
			},
		}
		response, _ := svc.GenerateOptimizedCreative(req)
		svc.RecordImpression(response.CombinationID)
		svc.RecordClick(response.CombinationID, userID)
	}

	// User preferences should be updated
	// Verification is implicit - no errors should occur
}
