package service

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/taskirx/go-bidding-engine/internal/cache"
)

// InsertTestPublisher is a test helper to insert a publisher for testing purposes only.
func (s *DirectPublisherService) InsertTestPublisher(pub *DirectPublisher) {
	s.publishers.Store(pub.ID, pub)
}

// DirectPublisherService manages direct publisher relationships
// for supply path optimization and reduced intermediary fees
type DirectPublisherService struct {
	cache         cache.Cache
	publishers    sync.Map // publisherID -> *DirectPublisher
	integrations  sync.Map // integrationID -> *PublisherIntegration
	pathAnalytics sync.Map // pathKey -> *SupplyPathMetrics
	mu            sync.RWMutex
	config        DirectPublisherConfig
}

// DirectPublisherConfig holds configuration
type DirectPublisherConfig struct {
	MinQualityScore         float64       `json:"min_quality_score"`
	MaxPathHops             int           `json:"max_path_hops"`
	TargetDirectRate        float64       `json:"target_direct_rate"`
	FeeTransparencyRequired bool          `json:"fee_transparency_required"`
	AuditInterval           time.Duration `json:"audit_interval"`
}

// DirectPublisher represents a direct publisher relationship
type DirectPublisher struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Domain          string `json:"domain"`
	Status          string `json:"status"`           // pending, active, suspended
	IntegrationType string `json:"integration_type"` // direct, reseller, header_bidding

	// Quality Metrics
	QualityScore     float64 `json:"quality_score"`
	ViewabilityRate  float64 `json:"viewability_rate"`
	IVTRate          float64 `json:"ivt_rate"` // Invalid traffic rate
	BrandSafetyScore float64 `json:"brand_safety_score"`

	// Performance Metrics
	AvgBidFloor      float64 `json:"avg_bid_floor"`
	AvgWinRate       float64 `json:"avg_win_rate"`
	AvgCPM           float64 `json:"avg_cpm"`
	TotalImpressions int64   `json:"total_impressions"`
	TotalRevenue     float64 `json:"total_revenue"`

	// Supply Path Info
	AdsText        string            `json:"ads_txt"` // ads.txt seller ID
	SellerID       string            `json:"seller_id"`
	IsDirectSeller bool              `json:"is_direct_seller"`
	SupplyChain    []SupplyChainNode `json:"supply_chain"`

	// Fee Structure
	FeeStructure FeeStructure `json:"fee_structure"`

	// Inventory
	AvailableFormats []string `json:"available_formats"`
	DailyCapacity    int64    `json:"daily_capacity"`
	GeoAvailability  []string `json:"geo_availability"`

	// Contract
	ContractStart time.Time `json:"contract_start"`
	ContractEnd   time.Time `json:"contract_end"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SupplyChainNode represents a node in the supply chain
type SupplyChainNode struct {
	ASI    string  `json:"asi"`  // Advertising System Identifier
	SID    string  `json:"sid"`  // Seller ID
	HP     int     `json:"hp"`   // Whether the node is authorized (1) or not (0)
	RID    string  `json:"rid"`  // Request ID
	Name   string  `json:"name"` // Optional name
	Domain string  `json:"domain"`
	Fee    float64 `json:"fee"` // Fee percentage taken at this node
}

// FeeStructure represents the fee breakdown
type FeeStructure struct {
	TechFee         float64 `json:"tech_fee"`
	DataFee         float64 `json:"data_fee"`
	VerificationFee float64 `json:"verification_fee"`
	MediaCost       float64 `json:"media_cost"`
	TotalTakeRate   float64 `json:"total_take_rate"`
	NetToPublisher  float64 `json:"net_to_publisher"` // Percentage that reaches publisher
}

// PublisherIntegration represents an integration with a publisher
type PublisherIntegration struct {
	ID              string `json:"id"`
	PublisherID     string `json:"publisher_id"`
	IntegrationType string `json:"integration_type"` // tag, api, prebid_server, oRTB
	Endpoint        string `json:"endpoint"`
	APIKey          string `json:"api_key,omitempty"`
	Status          string `json:"status"` // active, testing, disabled

	// Performance
	AvgLatency    float64 `json:"avg_latency_ms"`
	SuccessRate   float64 `json:"success_rate"`
	ErrorRate     float64 `json:"error_rate"`
	TotalRequests int64   `json:"total_requests"`

	// Config
	Timeout time.Duration `json:"timeout"`
	MaxQPS  int           `json:"max_qps"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SupplyPathMetrics tracks metrics for a supply path
type SupplyPathMetrics struct {
	PathKey        string    `json:"path_key"`
	PublisherID    string    `json:"publisher_id"`
	PathLength     int       `json:"path_length"`
	TotalFees      float64   `json:"total_fees"`
	AvgLatency     float64   `json:"avg_latency_ms"`
	WinRate        float64   `json:"win_rate"`
	Impressions    int64     `json:"impressions"`
	Spend          float64   `json:"spend"`
	EffectiveCPM   float64   `json:"effective_cpm"`
	IsOptimal      bool      `json:"is_optimal"`
	Recommendation string    `json:"recommendation"`
	LastUpdated    time.Time `json:"last_updated"`
}

// PathOptimizationResult represents an optimization recommendation
type PathOptimizationResult struct {
	PublisherID        string   `json:"publisher_id"`
	CurrentPath        []string `json:"current_path"`
	RecommendedPath    []string `json:"recommended_path"`
	CurrentFees        float64  `json:"current_fees"`
	ProjectedFees      float64  `json:"projected_fees"`
	FeeSavings         float64  `json:"fee_savings"`
	LatencyReduction   float64  `json:"latency_reduction_ms"`
	QualityImprovement float64  `json:"quality_improvement"`
	Priority           string   `json:"priority"`
	Implementation     string   `json:"implementation"`
}

// NewDirectPublisherService creates a new direct publisher service
func NewDirectPublisherService(cache cache.Cache) *DirectPublisherService {
	return &DirectPublisherService{
		cache: cache,
		config: DirectPublisherConfig{
			MinQualityScore:         0.7,
			MaxPathHops:             3,
			TargetDirectRate:        0.7, // 70% direct relationships target
			FeeTransparencyRequired: true,
			AuditInterval:           24 * time.Hour,
		},
	}
}

// RegisterPublisher registers a new direct publisher
func (s *DirectPublisherService) RegisterPublisher(pub *DirectPublisher) (*DirectPublisher, error) {
	if pub.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	if pub.ID == "" {
		pub.ID = uuid.New().String()
	}

	pub.Status = "pending"
	pub.CreatedAt = time.Now()
	pub.UpdatedAt = time.Now()

	// Calculate initial quality score
	pub.QualityScore = s.calculateQualityScore(pub)

	s.publishers.Store(pub.ID, pub)
	return pub, nil
}

// GetPublisher retrieves a publisher by ID
func (s *DirectPublisherService) GetPublisher(publisherID string) (*DirectPublisher, error) {
	if val, ok := s.publishers.Load(publisherID); ok {
		return val.(*DirectPublisher), nil
	}
	return nil, fmt.Errorf("publisher not found: %s", publisherID)
}

// UpdatePublisher updates a publisher
func (s *DirectPublisherService) UpdatePublisher(pub *DirectPublisher) error {
	if _, ok := s.publishers.Load(pub.ID); !ok {
		return fmt.Errorf("publisher not found: %s", pub.ID)
	}

	pub.UpdatedAt = time.Now()
	pub.QualityScore = s.calculateQualityScore(pub)
	s.publishers.Store(pub.ID, pub)
	return nil
}

// ActivatePublisher activates a publisher relationship
func (s *DirectPublisherService) ActivatePublisher(publisherID string) error {
	pub, err := s.GetPublisher(publisherID)
	if err != nil {
		return err
	}

	if pub.QualityScore < s.config.MinQualityScore {
		return fmt.Errorf("publisher quality score (%.2f) below minimum threshold (%.2f)",
			pub.QualityScore, s.config.MinQualityScore)
	}

	pub.Status = "active"
	pub.UpdatedAt = time.Now()
	s.publishers.Store(publisherID, pub)
	return nil
}

// SuspendPublisher suspends a publisher
func (s *DirectPublisherService) SuspendPublisher(publisherID, reason string) error {
	pub, err := s.GetPublisher(publisherID)
	if err != nil {
		return err
	}

	pub.Status = "suspended"
	pub.UpdatedAt = time.Now()
	s.publishers.Store(publisherID, pub)
	return nil
}

// ListPublishers returns all publishers, optionally filtered
func (s *DirectPublisherService) ListPublishers(status string, minQuality float64) []*DirectPublisher {
	var publishers []*DirectPublisher

	s.publishers.Range(func(key, value interface{}) bool {
		pub := value.(*DirectPublisher)

		if status != "" && pub.Status != status {
			return true
		}
		if pub.QualityScore < minQuality {
			return true
		}

		publishers = append(publishers, pub)
		return true
	})

	// Sort by quality score descending
	sort.Slice(publishers, func(i, j int) bool {
		return publishers[i].QualityScore > publishers[j].QualityScore
	})

	return publishers
}

// AddIntegration adds an integration for a publisher
func (s *DirectPublisherService) AddIntegration(integration *PublisherIntegration) (*PublisherIntegration, error) {
	if integration.PublisherID == "" {
		return nil, fmt.Errorf("publisher_id is required")
	}

	// Verify publisher exists
	if _, err := s.GetPublisher(integration.PublisherID); err != nil {
		return nil, err
	}

	if integration.ID == "" {
		integration.ID = uuid.New().String()
	}

	integration.Status = "testing"
	integration.SuccessRate = 1.0
	integration.CreatedAt = time.Now()
	integration.UpdatedAt = time.Now()

	if integration.Timeout == 0 {
		integration.Timeout = 150 * time.Millisecond
	}
	if integration.MaxQPS == 0 {
		integration.MaxQPS = 1000
	}

	s.integrations.Store(integration.ID, integration)
	return integration, nil
}

// GetIntegration retrieves an integration
func (s *DirectPublisherService) GetIntegration(integrationID string) (*PublisherIntegration, error) {
	if val, ok := s.integrations.Load(integrationID); ok {
		return val.(*PublisherIntegration), nil
	}
	return nil, fmt.Errorf("integration not found: %s", integrationID)
}

// AnalyzeSupplyPath analyzes a supply path for optimization opportunities
func (s *DirectPublisherService) AnalyzeSupplyPath(publisherID string) (*PathOptimizationResult, error) {
	pub, err := s.GetPublisher(publisherID)
	if err != nil {
		return nil, err
	}

	result := &PathOptimizationResult{
		PublisherID: publisherID,
	}

	// Analyze current supply chain
	currentPath := make([]string, 0, len(pub.SupplyChain))
	var totalFees float64

	for _, node := range pub.SupplyChain {
		currentPath = append(currentPath, node.ASI)
		totalFees += node.Fee
	}

	result.CurrentPath = currentPath
	result.CurrentFees = totalFees

	// Recommend optimizations
	if pub.IsDirectSeller {
		// Already direct, minor optimizations possible
		result.RecommendedPath = []string{pub.SellerID}
		result.ProjectedFees = pub.FeeStructure.TechFee
		result.Priority = "low"
		result.Implementation = "Already direct seller. Monitor for fee changes."
	} else if len(pub.SupplyChain) > s.config.MaxPathHops {
		// Long supply chain, recommend direct integration
		result.RecommendedPath = []string{pub.SellerID}
		result.ProjectedFees = totalFees * 0.5 // Estimate 50% fee reduction
		result.Priority = "high"
		result.Implementation = "Establish direct integration. Contact publisher for API access."
	} else {
		// Moderate chain, some optimization possible
		result.RecommendedPath = currentPath[:len(currentPath)/2+1]
		result.ProjectedFees = totalFees * 0.7 // Estimate 30% fee reduction
		result.Priority = "medium"
		result.Implementation = "Remove unnecessary intermediaries. Consider header bidding."
	}

	result.FeeSavings = result.CurrentFees - result.ProjectedFees
	result.LatencyReduction = float64(len(currentPath)-len(result.RecommendedPath)) * 20.0 // ~20ms per hop
	result.QualityImprovement = 0.05 * float64(len(currentPath)-len(result.RecommendedPath))

	return result, nil
}

// RecordPathMetrics records metrics for a supply path
func (s *DirectPublisherService) RecordPathMetrics(metrics *SupplyPathMetrics) {
	metrics.LastUpdated = time.Now()
	metrics.EffectiveCPM = (metrics.Spend / float64(metrics.Impressions)) * 1000

	// Determine if path is optimal
	metrics.IsOptimal = metrics.PathLength <= s.config.MaxPathHops &&
		metrics.TotalFees < 0.3 && // Less than 30% fees
		metrics.WinRate > 0.1 // Greater than 10% win rate

	if !metrics.IsOptimal {
		if metrics.PathLength > s.config.MaxPathHops {
			metrics.Recommendation = "Reduce supply chain length"
		} else if metrics.TotalFees >= 0.3 {
			metrics.Recommendation = "Negotiate lower fees or go direct"
		} else {
			metrics.Recommendation = "Improve bid strategy for higher win rate"
		}
	}

	s.pathAnalytics.Store(metrics.PathKey, metrics)
}

// GetPathMetrics retrieves path metrics
func (s *DirectPublisherService) GetPathMetrics(pathKey string) (*SupplyPathMetrics, error) {
	if val, ok := s.pathAnalytics.Load(pathKey); ok {
		return val.(*SupplyPathMetrics), nil
	}
	return nil, fmt.Errorf("path metrics not found: %s", pathKey)
}

// GetDirectRate calculates the direct publisher rate
func (s *DirectPublisherService) GetDirectRate() float64 {
	var total, direct int

	s.publishers.Range(func(key, value interface{}) bool {
		pub := value.(*DirectPublisher)
		if pub.Status == "active" {
			total++
			if pub.IsDirectSeller {
				direct++
			}
		}
		return true
	})

	if total == 0 {
		return 0
	}
	return float64(direct) / float64(total)
}

// GetStats returns service statistics
func (s *DirectPublisherService) GetStats() map[string]interface{} {
	var totalPublishers, activePublishers, directSellers int
	var totalIntegrations, activeIntegrations int
	var totalImpressions int64
	var totalRevenue float64
	var avgQuality float64

	s.publishers.Range(func(key, value interface{}) bool {
		pub := value.(*DirectPublisher)
		totalPublishers++
		avgQuality += pub.QualityScore
		totalImpressions += pub.TotalImpressions
		totalRevenue += pub.TotalRevenue

		if pub.Status == "active" {
			activePublishers++
			if pub.IsDirectSeller {
				directSellers++
			}
		}
		return true
	})

	s.integrations.Range(func(key, value interface{}) bool {
		integration := value.(*PublisherIntegration)
		totalIntegrations++
		if integration.Status == "active" {
			activeIntegrations++
		}
		return true
	})

	if totalPublishers > 0 {
		avgQuality /= float64(totalPublishers)
	}

	directRate := s.GetDirectRate()

	return map[string]interface{}{
		"total_publishers":    totalPublishers,
		"active_publishers":   activePublishers,
		"direct_sellers":      directSellers,
		"direct_rate":         directRate,
		"target_direct_rate":  s.config.TargetDirectRate,
		"avg_quality_score":   avgQuality,
		"total_integrations":  totalIntegrations,
		"active_integrations": activeIntegrations,
		"total_impressions":   totalImpressions,
		"total_revenue":       totalRevenue,
	}
}

// Helper functions

func (s *DirectPublisherService) calculateQualityScore(pub *DirectPublisher) float64 {
	score := 0.0
	weights := 0.0

	// Viewability (0-1)
	if pub.ViewabilityRate > 0 {
		score += pub.ViewabilityRate * 0.3
		weights += 0.3
	}

	// Invalid traffic (lower is better, 0-0.1 range typical)
	if pub.IVTRate >= 0 {
		ivtScore := 1.0 - (pub.IVTRate * 10) // Convert to 0-1 where 0 IVT = 1.0
		if ivtScore < 0 {
			ivtScore = 0
		}
		score += ivtScore * 0.25
		weights += 0.25
	}

	// Brand safety (0-1)
	if pub.BrandSafetyScore > 0 {
		score += pub.BrandSafetyScore * 0.25
		weights += 0.25
	}

	// Direct seller bonus
	if pub.IsDirectSeller {
		score += 0.1
		weights += 0.1
	}

	// Fee transparency bonus
	if pub.FeeStructure.TotalTakeRate > 0 {
		score += 0.1
		weights += 0.1
	}

	if weights == 0 {
		return 0.5 // Default score if no metrics
	}

	return score / weights
}
