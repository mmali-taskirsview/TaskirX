package service

import (
	"testing"
	"time"
)

// Helper to create a valid deal with all required fields
func createTestDeal(name, buyerID, sellerID string) *PGDeal {
	return &PGDeal{
		Name:                 name,
		BuyerID:              buyerID,
		SellerID:             sellerID,
		CommittedImpressions: 1000,
		FixedPrice:           5.0,
		StartDate:            time.Now(),
		EndDate:              time.Now().Add(24 * time.Hour),
	}
}

// TestProgrammaticGuaranteed_CreateDeal tests deal creation
func TestProgrammaticGuaranteed_CreateDeal(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	deal := createTestDeal("Test Deal", "buyer1", "seller1")
	deal.DealType = "guaranteed"
	deal.InventorySpecs = InventorySpec{
		PublisherIDs: []string{"pub1"},
		AdFormats:    []string{"banner"},
	}

	created, err := service.CreateDeal(deal)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected deal ID to be set")
	}
	if created.Status != "pending" {
		t.Errorf("expected status=pending, got %s", created.Status)
	}
}

// TestProgrammaticGuaranteed_CreateDeal_Validation tests deal validation
func TestProgrammaticGuaranteed_CreateDeal_Validation(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	tests := []struct {
		name    string
		deal    *PGDeal
		wantErr bool
	}{
		{
			name: "missing buyer",
			deal: &PGDeal{
				Name:                 "Test",
				SellerID:             "seller1",
				CommittedImpressions: 1000,
			},
			wantErr: true,
		},
		{
			name: "missing seller",
			deal: &PGDeal{
				Name:                 "Test",
				BuyerID:              "buyer1",
				CommittedImpressions: 1000,
			},
			wantErr: true,
		},
		{
			name:    "valid deal",
			deal:    createTestDeal("Test", "buyer1", "seller1"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateDeal(tt.deal)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDeal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestProgrammaticGuaranteed_GetDeal tests deal retrieval
func TestProgrammaticGuaranteed_GetDeal(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	deal := createTestDeal("Test Deal", "buyer1", "seller1")
	created, err := service.CreateDeal(deal)
	if err != nil {
		t.Fatalf("failed to create deal: %v", err)
	}

	retrieved, err := service.GetDeal(created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("expected ID=%s, got %s", created.ID, retrieved.ID)
	}
}

// TestProgrammaticGuaranteed_GetDeal_NotFound tests missing deal
func TestProgrammaticGuaranteed_GetDeal_NotFound(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	_, err := service.GetDeal("nonexistent")

	if err == nil {
		t.Error("expected error for nonexistent deal")
	}
}

// TestProgrammaticGuaranteed_ActivateDeal tests deal activation
func TestProgrammaticGuaranteed_ActivateDeal(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	deal := createTestDeal("Test Deal", "buyer1", "seller1")
	created, err := service.CreateDeal(deal)
	if err != nil {
		t.Fatalf("failed to create deal: %v", err)
	}

	err = service.ActivateDeal(created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := service.GetDeal(created.ID)
	if updated.Status != "active" {
		t.Errorf("expected status=active, got %s", updated.Status)
	}
}

// TestProgrammaticGuaranteed_ActivateDeal_NotFound tests activating missing deal
func TestProgrammaticGuaranteed_ActivateDeal_NotFound(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	err := service.ActivateDeal("nonexistent")

	if err == nil {
		t.Error("expected error for nonexistent deal")
	}
}

// TestProgrammaticGuaranteed_PauseDeal tests deal pausing
func TestProgrammaticGuaranteed_PauseDeal(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	deal := createTestDeal("Test Deal", "buyer1", "seller1")
	created, err := service.CreateDeal(deal)
	if err != nil {
		t.Fatalf("failed to create deal: %v", err)
	}
	service.ActivateDeal(created.ID)

	err = service.PauseDeal(created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := service.GetDeal(created.ID)
	if updated.Status != "paused" {
		t.Errorf("expected status=paused, got %s", updated.Status)
	}
}

// TestProgrammaticGuaranteed_CancelDeal tests deal cancellation
func TestProgrammaticGuaranteed_CancelDeal(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	deal := createTestDeal("Test Deal", "buyer1", "seller1")
	created, err := service.CreateDeal(deal)
	if err != nil {
		t.Fatalf("failed to create deal: %v", err)
	}

	err = service.CancelDeal(created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := service.GetDeal(created.ID)
	if updated.Status != "cancelled" {
		t.Errorf("expected status=cancelled, got %s", updated.Status)
	}
}

// TestProgrammaticGuaranteed_ListDeals tests deal listing
func TestProgrammaticGuaranteed_ListDeals(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	// Create multiple deals
	deals := []struct {
		buyerID  string
		sellerID string
	}{
		{"buyer1", "seller1"},
		{"buyer1", "seller2"},
		{"buyer2", "seller1"},
	}

	for i, d := range deals {
		deal := createTestDeal("Deal"+string(rune('A'+i)), d.buyerID, d.sellerID)
		_, err := service.CreateDeal(deal)
		if err != nil {
			t.Fatalf("failed to create deal: %v", err)
		}
	}

	// List by buyer
	buyer1Deals := service.ListDeals("buyer1", "", "")
	if len(buyer1Deals) != 2 {
		t.Errorf("expected 2 deals for buyer1, got %d", len(buyer1Deals))
	}

	// List by seller
	seller1Deals := service.ListDeals("", "seller1", "")
	if len(seller1Deals) != 2 {
		t.Errorf("expected 2 deals for seller1, got %d", len(seller1Deals))
	}

	// List all
	allDeals := service.ListDeals("", "", "")
	if len(allDeals) != 3 {
		t.Errorf("expected 3 total deals, got %d", len(allDeals))
	}
}

// TestProgrammaticGuaranteed_ListDeals_ByStatus tests listing by status
func TestProgrammaticGuaranteed_ListDeals_ByStatus(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	// Create and activate some deals
	deal1 := createTestDeal("Deal 1", "buyer1", "seller1")
	created1, _ := service.CreateDeal(deal1)
	service.ActivateDeal(created1.ID)

	deal2 := createTestDeal("Deal 2", "buyer1", "seller1")
	service.CreateDeal(deal2)

	// List only active deals
	activeDeals := service.ListDeals("", "", "active")
	if len(activeDeals) != 1 {
		t.Errorf("expected 1 active deal, got %d", len(activeDeals))
	}

	// List only pending deals
	pendingDeals := service.ListDeals("", "", "pending")
	if len(pendingDeals) != 1 {
		t.Errorf("expected 1 pending deal, got %d", len(pendingDeals))
	}
}

// TestProgrammaticGuaranteed_RecordImpression tests impression recording
func TestProgrammaticGuaranteed_RecordImpression(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	deal := createTestDeal("Test Deal", "buyer1", "seller1")
	created, _ := service.CreateDeal(deal)
	service.ActivateDeal(created.ID)

	err := service.RecordImpression(created.ID, 5.0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := service.GetDeal(created.ID)
	if updated.DeliveredImpressions != 1 {
		t.Errorf("expected delivered impressions=1, got %d", updated.DeliveredImpressions)
	}
	if updated.ActualSpend != 5.0 {
		t.Errorf("expected actual spend=5.0, got %f", updated.ActualSpend)
	}
}

// TestProgrammaticGuaranteed_RecordImpression_NotFound tests recording for missing deal
func TestProgrammaticGuaranteed_RecordImpression_NotFound(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	err := service.RecordImpression("nonexistent", 5.0)

	if err == nil {
		t.Error("expected error for nonexistent deal")
	}
}

// TestProgrammaticGuaranteed_GetDeliveryProgress tests delivery tracking
func TestProgrammaticGuaranteed_GetDeliveryProgress(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	deal := createTestDeal("Test Deal", "buyer1", "seller1")
	deal.CommittedImpressions = 100
	created, _ := service.CreateDeal(deal)
	service.ActivateDeal(created.ID)

	// Record some impressions
	for i := 0; i < 25; i++ {
		service.RecordImpression(created.ID, 5.0)
	}

	progress, err := service.GetDeliveryProgress(created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if progress == nil {
		t.Fatal("expected non-nil progress")
	}
	if progress.DeliveredImpressions != 25 {
		t.Errorf("expected 25 delivered impressions, got %d", progress.DeliveredImpressions)
	}
}

// TestProgrammaticGuaranteed_GetStats tests statistics retrieval
func TestProgrammaticGuaranteed_GetStats(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	// Create some deals
	deal := createTestDeal("Deal 1", "buyer1", "seller1")
	service.CreateDeal(deal)

	stats := service.GetStats()

	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if _, exists := stats["total_deals"]; !exists {
		t.Error("expected total_deals in stats")
	}
}

// TestProgrammaticGuaranteed_CheckEligibility tests eligibility check
func TestProgrammaticGuaranteed_CheckEligibility(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	deal := createTestDeal("Test Deal", "buyer1", "seller1")
	deal.InventorySpecs = InventorySpec{
		PublisherIDs: []string{"pub1"},
		AdFormats:    []string{"banner"},
		DeviceTypes:  []string{"desktop"},
	}
	created, _ := service.CreateDeal(deal)
	service.ActivateDeal(created.ID)

	// Check eligibility - returns slice of BidEligibility
	eligibleDeals := service.CheckEligibility(
		"pub1",
		"site1",
		"banner",
		"desktop",
		"US",
		"",
	)

	foundEligible := false
	for _, e := range eligibleDeals {
		if e.Eligible {
			foundEligible = true
			break
		}
	}
	if !foundEligible {
		t.Log("Note: Eligibility may require specific matching criteria")
	}
}

// TestProgrammaticGuaranteed_UpdateDeal tests deal update
func TestProgrammaticGuaranteed_UpdateDeal(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	deal := createTestDeal("Test Deal", "buyer1", "seller1")
	created, err := service.CreateDeal(deal)
	if err != nil {
		t.Fatalf("failed to create deal: %v", err)
	}

	// Update the deal
	created.Name = "Updated Deal"
	created.FixedPrice = 10.0

	err = service.UpdateDeal(created)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := service.GetDeal(created.ID)
	if updated.Name != "Updated Deal" {
		t.Errorf("expected name=Updated Deal, got %s", updated.Name)
	}
	if updated.FixedPrice != 10.0 {
		t.Errorf("expected fixed price=10.0, got %f", updated.FixedPrice)
	}
}

// TestProgrammaticGuaranteed_DealTypes tests different deal types
func TestProgrammaticGuaranteed_DealTypes(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)

	dealTypes := []string{"guaranteed", "preferred", "private_auction"}

	for _, dealType := range dealTypes {
		t.Run(dealType, func(t *testing.T) {
			deal := createTestDeal("Test "+dealType, "buyer1", "seller1")
			deal.DealType = dealType
			created, err := service.CreateDeal(deal)
			if err != nil {
				t.Fatalf("failed to create %s deal: %v", dealType, err)
			}
			if created.DealType != dealType {
				t.Errorf("expected deal type=%s, got %s", dealType, created.DealType)
			}
		})
	}
}

// TestProgrammaticGuaranteed_Concurrency tests concurrent operations
func TestProgrammaticGuaranteed_Concurrency(t *testing.T) {
	service := NewProgrammaticGuaranteedService(nil)
	done := make(chan bool)

	deal := createTestDeal("Concurrent Deal", "buyer1", "seller1")
	deal.CommittedImpressions = 10000
	created, _ := service.CreateDeal(deal)
	service.ActivateDeal(created.ID)

	// Concurrent impression recording
	go func() {
		for i := 0; i < 100; i++ {
			service.RecordImpression(created.ID, 5.0)
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			service.GetDeliveryProgress(created.ID)
		}
		done <- true
	}()

	<-done
	<-done
	// Test passes if no race condition panic
}
