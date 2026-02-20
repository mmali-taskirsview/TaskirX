package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/taskirx/go-bidding-engine/internal/cache"
)

// ProgrammaticGuaranteedService handles programmatic guaranteed deals
// which are pre-negotiated, reserved inventory deals at fixed prices
type ProgrammaticGuaranteedService struct {
	cache           cache.Cache
	deals           sync.Map // dealID -> *PGDeal
	commitments     sync.Map // commitmentID -> *DealCommitment
	deliveryTracker sync.Map // dealID -> *DeliveryProgress
	mu              sync.RWMutex
	config          PGConfig
}

// PGConfig holds configuration for programmatic guaranteed
type PGConfig struct {
	MaxDealsPerBuyer       int           `json:"max_deals_per_buyer"`
	MinCommitmentValue     float64       `json:"min_commitment_value"`
	DefaultPriorityBoost   float64       `json:"default_priority_boost"`
	UnderdeliveryThreshold float64       `json:"underdelivery_threshold"`
	OverdeliveryAllowance  float64       `json:"overdelivery_allowance"`
	AlertThreshold         float64       `json:"alert_threshold"`
	ReconciliationInterval time.Duration `json:"reconciliation_interval"`
}

// PGDeal represents a programmatic guaranteed deal
type PGDeal struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BuyerID  string `json:"buyer_id"`
	SellerID string `json:"seller_id"`
	Status   string `json:"status"`    // pending, active, paused, completed, cancelled
	DealType string `json:"deal_type"` // guaranteed, preferred, private_auction

	// Inventory Details
	InventorySpecs InventorySpec `json:"inventory_specs"`

	// Pricing
	PriceType  string  `json:"price_type"` // fixed, floor
	FixedPrice float64 `json:"fixed_price"`
	FloorPrice float64 `json:"floor_price"`
	Currency   string  `json:"currency"`

	// Commitment
	CommittedImpressions int64   `json:"committed_impressions"`
	CommittedSpend       float64 `json:"committed_spend"`

	// Timeline
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`

	// Priority
	Priority int `json:"priority"` // Higher priority = first look

	// Tracking
	DeliveredImpressions int64   `json:"delivered_impressions"`
	ActualSpend          float64 `json:"actual_spend"`

	// Metadata
	Terms     map[string]string `json:"terms"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// InventorySpec specifies the inventory for the deal
type InventorySpec struct {
	PublisherIDs      []string `json:"publisher_ids"`
	SiteIDs           []string `json:"site_ids"`
	Placements        []string `json:"placements"`
	AdFormats         []string `json:"ad_formats"` // banner, video, native
	DeviceTypes       []string `json:"device_types"`
	GeoTargets        []string `json:"geo_targets"`
	AudienceSegments  []string `json:"audience_segments"`
	ContentCategories []string `json:"content_categories"`
	Viewability       float64  `json:"viewability"` // Minimum viewability threshold
}

// DealCommitment represents a commitment against a deal
type DealCommitment struct {
	ID              string    `json:"id"`
	DealID          string    `json:"deal_id"`
	BuyerID         string    `json:"buyer_id"`
	CommittedAmount float64   `json:"committed_amount"`
	CommittedVolume int64     `json:"committed_volume"`
	FulfilledAmount float64   `json:"fulfilled_amount"`
	FulfilledVolume int64     `json:"fulfilled_volume"`
	Status          string    `json:"status"` // active, fulfilled, underdelivered
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	CreatedAt       time.Time `json:"created_at"`
}

// DeliveryProgress tracks deal delivery progress
type DeliveryProgress struct {
	DealID               string    `json:"deal_id"`
	TargetImpressions    int64     `json:"target_impressions"`
	DeliveredImpressions int64     `json:"delivered_impressions"`
	TargetSpend          float64   `json:"target_spend"`
	ActualSpend          float64   `json:"actual_spend"`
	DeliveryRate         float64   `json:"delivery_rate"`     // Actual vs expected pace
	ExpectedDelivery     int64     `json:"expected_delivery"` // Expected by end of deal
	ProjectedShortfall   int64     `json:"projected_shortfall"`
	DaysRemaining        int       `json:"days_remaining"`
	RequiredDailyRate    int64     `json:"required_daily_rate"`
	CurrentDailyRate     int64     `json:"current_daily_rate"`
	Status               string    `json:"status"` // on_pace, underdelivering, overdelivering
	LastUpdated          time.Time `json:"last_updated"`
}

// BidEligibility represents whether a request is eligible for a PG deal
type BidEligibility struct {
	DealID   string  `json:"deal_id"`
	Eligible bool    `json:"eligible"`
	Price    float64 `json:"price"`
	Priority int     `json:"priority"`
	Reason   string  `json:"reason,omitempty"`
}

// NewProgrammaticGuaranteedService creates a new PG service
func NewProgrammaticGuaranteedService(cache cache.Cache) *ProgrammaticGuaranteedService {
	return &ProgrammaticGuaranteedService{
		cache: cache,
		config: PGConfig{
			MaxDealsPerBuyer:       100,
			MinCommitmentValue:     1000.0,
			DefaultPriorityBoost:   1.5,
			UnderdeliveryThreshold: 0.9, // 90% delivery threshold
			OverdeliveryAllowance:  1.1, // Allow 10% overdelivery
			AlertThreshold:         0.8, // Alert at 80% pacing
			ReconciliationInterval: time.Hour,
		},
	}
}

// CreateDeal creates a new programmatic guaranteed deal
func (s *ProgrammaticGuaranteedService) CreateDeal(deal *PGDeal) (*PGDeal, error) {
	if deal.BuyerID == "" || deal.SellerID == "" {
		return nil, fmt.Errorf("buyer_id and seller_id are required")
	}

	if deal.CommittedImpressions == 0 && deal.CommittedSpend == 0 {
		return nil, fmt.Errorf("either committed_impressions or committed_spend is required")
	}

	if deal.ID == "" {
		deal.ID = uuid.New().String()
	}

	deal.Status = "pending"
	deal.CreatedAt = time.Now()
	deal.UpdatedAt = time.Now()
	deal.DeliveredImpressions = 0
	deal.ActualSpend = 0

	if deal.Currency == "" {
		deal.Currency = "USD"
	}

	if deal.Priority == 0 {
		deal.Priority = 10 // Default priority
	}

	s.deals.Store(deal.ID, deal)

	// Initialize delivery tracker
	s.initDeliveryTracker(deal)

	return deal, nil
}

// GetDeal retrieves a deal by ID
func (s *ProgrammaticGuaranteedService) GetDeal(dealID string) (*PGDeal, error) {
	if val, ok := s.deals.Load(dealID); ok {
		return val.(*PGDeal), nil
	}
	return nil, fmt.Errorf("deal not found: %s", dealID)
}

// UpdateDeal updates an existing deal
func (s *ProgrammaticGuaranteedService) UpdateDeal(deal *PGDeal) error {
	if _, ok := s.deals.Load(deal.ID); !ok {
		return fmt.Errorf("deal not found: %s", deal.ID)
	}

	deal.UpdatedAt = time.Now()
	s.deals.Store(deal.ID, deal)
	return nil
}

// ActivateDeal activates a pending deal
func (s *ProgrammaticGuaranteedService) ActivateDeal(dealID string) error {
	deal, err := s.GetDeal(dealID)
	if err != nil {
		return err
	}

	if deal.Status != "pending" && deal.Status != "paused" {
		return fmt.Errorf("deal cannot be activated from status: %s", deal.Status)
	}

	deal.Status = "active"
	deal.UpdatedAt = time.Now()
	s.deals.Store(dealID, deal)
	return nil
}

// PauseDeal pauses an active deal
func (s *ProgrammaticGuaranteedService) PauseDeal(dealID string) error {
	deal, err := s.GetDeal(dealID)
	if err != nil {
		return err
	}

	if deal.Status != "active" {
		return fmt.Errorf("only active deals can be paused")
	}

	deal.Status = "paused"
	deal.UpdatedAt = time.Now()
	s.deals.Store(dealID, deal)
	return nil
}

// CancelDeal cancels a deal
func (s *ProgrammaticGuaranteedService) CancelDeal(dealID string) error {
	deal, err := s.GetDeal(dealID)
	if err != nil {
		return err
	}

	if deal.Status == "completed" || deal.Status == "cancelled" {
		return fmt.Errorf("deal is already %s", deal.Status)
	}

	deal.Status = "cancelled"
	deal.UpdatedAt = time.Now()
	s.deals.Store(dealID, deal)
	return nil
}

// ListDeals returns all deals, optionally filtered
func (s *ProgrammaticGuaranteedService) ListDeals(buyerID, sellerID, status string) []*PGDeal {
	var deals []*PGDeal

	s.deals.Range(func(key, value interface{}) bool {
		deal := value.(*PGDeal)

		// Apply filters
		if buyerID != "" && deal.BuyerID != buyerID {
			return true
		}
		if sellerID != "" && deal.SellerID != sellerID {
			return true
		}
		if status != "" && deal.Status != status {
			return true
		}

		deals = append(deals, deal)
		return true
	})

	return deals
}

// CheckEligibility checks if a bid request is eligible for any PG deals
func (s *ProgrammaticGuaranteedService) CheckEligibility(
	publisherID string,
	siteID string,
	placement string,
	adFormat string,
	deviceType string,
	geo string,
) []*BidEligibility {
	var eligibleDeals []*BidEligibility

	s.deals.Range(func(key, value interface{}) bool {
		deal := value.(*PGDeal)

		if deal.Status != "active" {
			return true
		}

		// Check if within deal timeframe
		now := time.Now()
		if now.Before(deal.StartDate) || now.After(deal.EndDate) {
			return true
		}

		// Check delivery progress
		progress := s.getDeliveryProgress(deal.ID)
		if progress != nil && progress.DeliveredImpressions >= int64(float64(deal.CommittedImpressions)*s.config.OverdeliveryAllowance) {
			eligibleDeals = append(eligibleDeals, &BidEligibility{
				DealID:   deal.ID,
				Eligible: false,
				Reason:   "Deal fully delivered",
			})
			return true
		}

		// Check inventory matching
		if !s.matchesInventorySpec(deal.InventorySpecs, publisherID, siteID, placement, adFormat, deviceType, geo) {
			return true
		}

		// Deal is eligible
		eligibleDeals = append(eligibleDeals, &BidEligibility{
			DealID:   deal.ID,
			Eligible: true,
			Price:    deal.FixedPrice,
			Priority: deal.Priority,
		})

		return true
	})

	return eligibleDeals
}

// RecordImpression records an impression against a deal
func (s *ProgrammaticGuaranteedService) RecordImpression(dealID string, price float64) error {
	deal, err := s.GetDeal(dealID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	deal.DeliveredImpressions++
	deal.ActualSpend += price
	deal.UpdatedAt = time.Now()

	// Check if deal is completed
	if deal.CommittedImpressions > 0 && deal.DeliveredImpressions >= deal.CommittedImpressions {
		deal.Status = "completed"
	}

	s.deals.Store(dealID, deal)
	s.updateDeliveryProgress(dealID)

	return nil
}

// GetDeliveryProgress returns delivery progress for a deal
func (s *ProgrammaticGuaranteedService) GetDeliveryProgress(dealID string) (*DeliveryProgress, error) {
	progress := s.getDeliveryProgress(dealID)
	if progress == nil {
		return nil, fmt.Errorf("no delivery data for deal: %s", dealID)
	}
	return progress, nil
}

// GetStats returns PG service statistics
func (s *ProgrammaticGuaranteedService) GetStats() map[string]interface{} {
	var totalDeals, activeDeals, completedDeals int
	var totalCommitted, totalDelivered int64
	var totalSpend float64

	s.deals.Range(func(key, value interface{}) bool {
		deal := value.(*PGDeal)
		totalDeals++
		totalCommitted += deal.CommittedImpressions
		totalDelivered += deal.DeliveredImpressions
		totalSpend += deal.ActualSpend

		switch deal.Status {
		case "active":
			activeDeals++
		case "completed":
			completedDeals++
		}
		return true
	})

	deliveryRate := float64(0)
	if totalCommitted > 0 {
		deliveryRate = float64(totalDelivered) / float64(totalCommitted)
	}

	return map[string]interface{}{
		"total_deals":     totalDeals,
		"active_deals":    activeDeals,
		"completed_deals": completedDeals,
		"total_committed": totalCommitted,
		"total_delivered": totalDelivered,
		"delivery_rate":   deliveryRate,
		"total_spend":     totalSpend,
	}
}

// Helper functions

func (s *ProgrammaticGuaranteedService) initDeliveryTracker(deal *PGDeal) {
	daysTotal := int(deal.EndDate.Sub(deal.StartDate).Hours() / 24)
	if daysTotal <= 0 {
		daysTotal = 1
	}

	progress := &DeliveryProgress{
		DealID:               deal.ID,
		TargetImpressions:    deal.CommittedImpressions,
		TargetSpend:          deal.CommittedSpend,
		DeliveredImpressions: 0,
		ActualSpend:          0,
		DeliveryRate:         0,
		DaysRemaining:        daysTotal,
		RequiredDailyRate:    deal.CommittedImpressions / int64(daysTotal),
		CurrentDailyRate:     0,
		Status:               "on_pace",
		LastUpdated:          time.Now(),
	}

	s.deliveryTracker.Store(deal.ID, progress)
}

func (s *ProgrammaticGuaranteedService) getDeliveryProgress(dealID string) *DeliveryProgress {
	if val, ok := s.deliveryTracker.Load(dealID); ok {
		return val.(*DeliveryProgress)
	}
	return nil
}

func (s *ProgrammaticGuaranteedService) updateDeliveryProgress(dealID string) {
	deal, err := s.GetDeal(dealID)
	if err != nil {
		return
	}

	progress := s.getDeliveryProgress(dealID)
	if progress == nil {
		return
	}

	now := time.Now()
	daysRemaining := int(deal.EndDate.Sub(now).Hours() / 24)
	if daysRemaining <= 0 {
		daysRemaining = 1
	}

	daysElapsed := int(now.Sub(deal.StartDate).Hours() / 24)
	if daysElapsed <= 0 {
		daysElapsed = 1
	}

	progress.DeliveredImpressions = deal.DeliveredImpressions
	progress.ActualSpend = deal.ActualSpend
	progress.DaysRemaining = daysRemaining
	progress.CurrentDailyRate = deal.DeliveredImpressions / int64(daysElapsed)

	// Calculate expected delivery based on current pace
	progress.ExpectedDelivery = progress.CurrentDailyRate*int64(daysRemaining) + deal.DeliveredImpressions

	// Calculate required daily rate
	remaining := deal.CommittedImpressions - deal.DeliveredImpressions
	if remaining > 0 {
		progress.RequiredDailyRate = remaining / int64(daysRemaining)
	}

	// Calculate shortfall
	if progress.ExpectedDelivery < deal.CommittedImpressions {
		progress.ProjectedShortfall = deal.CommittedImpressions - progress.ExpectedDelivery
	}

	// Delivery rate (actual vs target pace)
	expectedByNow := (deal.CommittedImpressions * int64(daysElapsed)) / int64(daysElapsed+daysRemaining)
	if expectedByNow > 0 {
		progress.DeliveryRate = float64(deal.DeliveredImpressions) / float64(expectedByNow)
	}

	// Update status
	if progress.DeliveryRate >= 1.0 {
		progress.Status = "on_pace"
	} else if progress.DeliveryRate >= s.config.AlertThreshold {
		progress.Status = "slightly_behind"
	} else {
		progress.Status = "underdelivering"
	}

	progress.LastUpdated = now
	s.deliveryTracker.Store(dealID, progress)
}

func (s *ProgrammaticGuaranteedService) matchesInventorySpec(
	spec InventorySpec,
	publisherID, siteID, placement, adFormat, deviceType, geo string,
) bool {
	// If no filters specified, match all
	if len(spec.PublisherIDs) == 0 && len(spec.SiteIDs) == 0 &&
		len(spec.Placements) == 0 && len(spec.AdFormats) == 0 &&
		len(spec.DeviceTypes) == 0 && len(spec.GeoTargets) == 0 {
		return true
	}

	// Check each filter
	if len(spec.PublisherIDs) > 0 && !contains(spec.PublisherIDs, publisherID) {
		return false
	}
	if len(spec.SiteIDs) > 0 && !contains(spec.SiteIDs, siteID) {
		return false
	}
	if len(spec.Placements) > 0 && !contains(spec.Placements, placement) {
		return false
	}
	if len(spec.AdFormats) > 0 && !contains(spec.AdFormats, adFormat) {
		return false
	}
	if len(spec.DeviceTypes) > 0 && !contains(spec.DeviceTypes, deviceType) {
		return false
	}
	if len(spec.GeoTargets) > 0 && !contains(spec.GeoTargets, geo) {
		return false
	}

	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
