package service

import (
	"testing"
	"time"
)

func createTestDPub(name, domain string) *DirectPublisher {
	return &DirectPublisher{
		Name:             name,
		Domain:           domain,
		IntegrationType:  "direct",
		QualityScore:     0.8,
		ViewabilityRate:  0.7,
		IVTRate:          0.02,
		BrandSafetyScore: 0.9,
		AvailableFormats: []string{"banner", "video"},
		DailyCapacity:    100000,
		GeoAvailability:  []string{"US", "CA"},
		ContractStart:    time.Now(),
		ContractEnd:      time.Now().Add(365 * 24 * time.Hour),
	}
}

func TestDirectPub_RegisterPublisher(t *testing.T) {
	service := NewDirectPublisherService(nil)
	pub := createTestDPub("Test Publisher", "example.com")

	registered, err := service.RegisterPublisher(pub)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if registered.ID == "" {
		t.Error("expected publisher ID to be set")
	}
	if registered.Status != "pending" {
		t.Errorf("expected status=pending, got %s", registered.Status)
	}
}

func TestDirectPub_RegisterPublisher_Validation(t *testing.T) {
	service := NewDirectPublisherService(nil)

	tests := []struct {
		name    string
		pub     *DirectPublisher
		wantErr bool
	}{
		{"missing domain", &DirectPublisher{Name: "Test"}, true},
		{"valid publisher", createTestDPub("Valid", "valid.com"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.RegisterPublisher(tt.pub)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterPublisher() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDirectPub_GetPublisher(t *testing.T) {
	service := NewDirectPublisherService(nil)
	pub := createTestDPub("Test Publisher", "example.com")
	registered, _ := service.RegisterPublisher(pub)

	retrieved, err := service.GetPublisher(registered.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != registered.ID {
		t.Errorf("expected ID=%s, got %s", registered.ID, retrieved.ID)
	}
}

func TestDirectPub_GetPublisher_NotFound(t *testing.T) {
	service := NewDirectPublisherService(nil)
	_, err := service.GetPublisher("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent publisher")
	}
}

func TestDirectPub_ActivatePublisher(t *testing.T) {
	service := NewDirectPublisherService(nil)
	pub := createTestDPub("Test Publisher", "example.com")
	registered, _ := service.RegisterPublisher(pub)

	err := service.ActivatePublisher(registered.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated, _ := service.GetPublisher(registered.ID)
	if updated.Status != "active" {
		t.Errorf("expected status=active, got %s", updated.Status)
	}
}

func TestDirectPub_ActivatePublisher_NotFound(t *testing.T) {
	service := NewDirectPublisherService(nil)
	err := service.ActivatePublisher("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent publisher")
	}
}

func TestDirectPub_SuspendPublisher(t *testing.T) {
	service := NewDirectPublisherService(nil)
	pub := createTestDPub("Test Publisher", "example.com")
	registered, _ := service.RegisterPublisher(pub)
	service.ActivatePublisher(registered.ID)

	err := service.SuspendPublisher(registered.ID, "quality issues")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated, _ := service.GetPublisher(registered.ID)
	if updated.Status != "suspended" {
		t.Errorf("expected status=suspended, got %s", updated.Status)
	}
}

func TestDirectPub_ListPublishers(t *testing.T) {
	service := NewDirectPublisherService(nil)

	for _, name := range []string{"A", "B", "C"} {
		pub := createTestDPub("Publisher "+name, name+".com")
		service.RegisterPublisher(pub)
	}

	all := service.ListPublishers("", 0)
	if len(all) != 3 {
		t.Errorf("expected 3 publishers, got %d", len(all))
	}
}

func TestDirectPub_ListPublishers_ByStatus(t *testing.T) {
	service := NewDirectPublisherService(nil)

	pub1 := createTestDPub("Publisher 1", "one.com")
	reg1, _ := service.RegisterPublisher(pub1)
	service.ActivatePublisher(reg1.ID)

	pub2 := createTestDPub("Publisher 2", "two.com")
	service.RegisterPublisher(pub2)

	active := service.ListPublishers("active", 0)
	if len(active) != 1 {
		t.Errorf("expected 1 active publisher, got %d", len(active))
	}

	pending := service.ListPublishers("pending", 0)
	if len(pending) != 1 {
		t.Errorf("expected 1 pending publisher, got %d", len(pending))
	}
}

func TestDirectPub_UpdatePublisher(t *testing.T) {
	service := NewDirectPublisherService(nil)
	pub := createTestDPub("Test Publisher", "example.com")
	registered, _ := service.RegisterPublisher(pub)

	registered.Name = "Updated Publisher"
	err := service.UpdatePublisher(registered)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated, _ := service.GetPublisher(registered.ID)
	if updated.Name != "Updated Publisher" {
		t.Errorf("expected name=Updated Publisher, got %s", updated.Name)
	}
}

func TestDirectPub_AddIntegration(t *testing.T) {
	service := NewDirectPublisherService(nil)
	pub := createTestDPub("Test Publisher", "example.com")
	registered, _ := service.RegisterPublisher(pub)

	integration := &PublisherIntegration{
		PublisherID:     registered.ID,
		IntegrationType: "prebid_server",
		Status:          "active",
	}

	added, err := service.AddIntegration(integration)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if added.ID == "" {
		t.Error("expected integration ID to be set")
	}
}

func TestDirectPub_GetIntegration(t *testing.T) {
	service := NewDirectPublisherService(nil)
	pub := createTestDPub("Test Publisher", "example.com")
	registered, _ := service.RegisterPublisher(pub)

	integration := &PublisherIntegration{
		PublisherID:     registered.ID,
		IntegrationType: "prebid_server",
	}
	added, _ := service.AddIntegration(integration)

	retrieved, err := service.GetIntegration(added.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != added.ID {
		t.Errorf("expected ID=%s, got %s", added.ID, retrieved.ID)
	}
}

func TestDirectPub_AnalyzeSupplyPath(t *testing.T) {
	service := NewDirectPublisherService(nil)
	pub := createTestDPub("Test Publisher", "example.com")
	pub.SupplyChain = []SupplyChainNode{
		{ASI: "ssp1.com", SID: "12345", HP: 1, Fee: 0.15},
	}
	registered, _ := service.RegisterPublisher(pub)
	service.ActivatePublisher(registered.ID)

	result, err := service.AnalyzeSupplyPath(registered.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestDirectPub_RecordPathMetrics(t *testing.T) {
	service := NewDirectPublisherService(nil)

	metrics := &SupplyPathMetrics{
		PathKey:     "pub1:ssp1",
		Impressions: 1000,
		Spend:       50.0,
		AvgLatency:  25.0,
		WinRate:     0.15,
	}

	service.RecordPathMetrics(metrics)

	retrieved, err := service.GetPathMetrics("pub1:ssp1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.Impressions != 1000 {
		t.Errorf("expected 1000 impressions, got %d", retrieved.Impressions)
	}
}

func TestDirectPub_GetPathMetrics_NotFound(t *testing.T) {
	service := NewDirectPublisherService(nil)
	_, err := service.GetPathMetrics("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestDirectPub_GetDirectRate(t *testing.T) {
	service := NewDirectPublisherService(nil)

	pub1 := createTestDPub("Direct 1", "direct1.com")
	pub1.IsDirectSeller = true
	reg1, _ := service.RegisterPublisher(pub1)
	service.ActivatePublisher(reg1.ID)

	rate := service.GetDirectRate()
	if rate < 0 || rate > 1 {
		t.Errorf("expected rate in [0,1], got %f", rate)
	}
}

func TestDirectPub_GetStats(t *testing.T) {
	service := NewDirectPublisherService(nil)
	pub := createTestDPub("Test", "test.com")
	service.RegisterPublisher(pub)

	stats := service.GetStats()

	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if _, exists := stats["total_publishers"]; !exists {
		t.Error("expected total_publishers in stats")
	}
}

func TestDirectPub_IntegrationTypes(t *testing.T) {
	service := NewDirectPublisherService(nil)

	for _, intType := range []string{"direct", "reseller", "header_bidding"} {
		t.Run(intType, func(t *testing.T) {
			pub := createTestDPub("Test "+intType, intType+".com")
			pub.IntegrationType = intType
			registered, err := service.RegisterPublisher(pub)
			if err != nil {
				t.Fatalf("failed to register: %v", err)
			}
			if registered.IntegrationType != intType {
				t.Errorf("expected type=%s, got %s", intType, registered.IntegrationType)
			}
		})
	}
}

func TestDirectPub_Concurrency(t *testing.T) {
	service := NewDirectPublisherService(nil)
	done := make(chan bool)

	pub := createTestDPub("Concurrent", "concurrent.com")
	registered, _ := service.RegisterPublisher(pub)
	service.ActivatePublisher(registered.ID)

	go func() {
		for i := 0; i < 100; i++ {
			metrics := &SupplyPathMetrics{PathKey: "concurrent:path", Impressions: int64(i)}
			service.RecordPathMetrics(metrics)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			service.GetPublisher(registered.ID)
			service.GetStats()
		}
		done <- true
	}()

	<-done
	<-done
}
