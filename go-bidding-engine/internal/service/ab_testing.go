package service

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/taskirx/go-bidding-engine/internal/cache"
)

// ABTestingService provides A/B testing framework with statistical significance testing
type ABTestingService struct {
	cache       cache.Cache
	experiments sync.Map // experimentID -> *abExperiment
	assignments sync.Map // userID -> map[experimentID]variantID
	mu          sync.RWMutex
	config      ABTestingConfig
}

// ABTestingConfig holds configuration for A/B testing
type ABTestingConfig struct {
	MinSampleSize         int     `json:"min_sample_size"`
	SignificanceLevel     float64 `json:"significance_level"`
	MinDetectableEffect   float64 `json:"min_detectable_effect"`
	MaxRunningExperiments int     `json:"max_running_experiments"`
	AutoStopOnSignificance bool   `json:"auto_stop_on_significance"`
}

// abExperiment represents an A/B test experiment (prefixed to avoid collision with incrementality.go)
type abExperiment struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Type              string     `json:"type"` // "ab", "multivariate", "bandit"
	Status            string     `json:"status"` // "draft", "running", "paused", "completed"
	Variants          []*variant `json:"variants"`
	TrafficAllocation float64    `json:"traffic_allocation"` // Percentage of traffic to include
	StartDate         time.Time  `json:"start_date"`
	EndDate           time.Time  `json:"end_date"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	WinningVariant    string     `json:"winning_variant"`
	Metrics           []string   `json:"metrics"` // Metrics to track
}

// variant represents a test variant
type variant struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Weight      float64         `json:"weight"` // Traffic weight (0-1)
	IsControl   bool            `json:"is_control"`
	Config      map[string]any  `json:"config"` // Variant-specific configuration
	Metrics     *variantMetrics `json:"metrics"`
}

// variantMetrics holds performance metrics for a variant
type variantMetrics struct {
	Impressions    int64            `json:"impressions"`
	Clicks         int64            `json:"clicks"`
	Conversions    int64            `json:"conversions"`
	Revenue        float64          `json:"revenue"`
	CTR            float64          `json:"ctr"`
	ConversionRate float64          `json:"conversion_rate"`
	ARPU           float64          `json:"arpu"`
	CustomMetrics  map[string]float64 `json:"custom_metrics"`
	mu             sync.Mutex
}

// ExperimentResult holds the analysis results
type ExperimentResult struct {
	ExperimentID       string                    `json:"experiment_id"`
	ExperimentName     string                    `json:"experiment_name"`
	Status             string                    `json:"status"`
	Duration           time.Duration             `json:"duration"`
	TotalSamples       int64                     `json:"total_samples"`
	VariantResults     []VariantResult           `json:"variant_results"`
	WinningVariant     string                    `json:"winning_variant"`
	Confidence         float64                   `json:"confidence"`
	IsSignificant      bool                      `json:"is_significant"`
	Recommendation     string                    `json:"recommendation"`
	StatisticalPower   float64                   `json:"statistical_power"`
	ExpectedLift       float64                   `json:"expected_lift"`
	RiskAssessment     string                    `json:"risk_assessment"`
}

// VariantResult holds results for a single variant
type VariantResult struct {
	VariantID      string  `json:"variant_id"`
	VariantName    string  `json:"variant_name"`
	IsControl      bool    `json:"is_control"`
	SampleSize     int64   `json:"sample_size"`
	ConversionRate float64 `json:"conversion_rate"`
	CTR            float64 `json:"ctr"`
	Revenue        float64 `json:"revenue"`
	Lift           float64 `json:"lift"` // Lift over control
	PValue         float64 `json:"p_value"`
	ConfidenceInterval struct {
		Lower float64 `json:"lower"`
		Upper float64 `json:"upper"`
	} `json:"confidence_interval"`
}

// CreateExperimentRequest for creating new experiments
type CreateExperimentRequest struct {
	Name              string           `json:"name"`
	Description       string           `json:"description"`
	Type              string           `json:"type"`
	Variants          []VariantRequest `json:"variants"`
	TrafficAllocation float64          `json:"traffic_allocation"`
	Metrics           []string         `json:"metrics"`
	Duration          time.Duration    `json:"duration"`
}

// VariantRequest for creating variants
type VariantRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Weight      float64        `json:"weight"`
	IsControl   bool           `json:"is_control"`
	Config      map[string]any `json:"config"`
}

// NewABTestingService creates a new A/B testing service
func NewABTestingService(c cache.Cache) *ABTestingService {
	return &ABTestingService{
		cache: c,
		config: ABTestingConfig{
			MinSampleSize:         100,
			SignificanceLevel:     0.05,
			MinDetectableEffect:   0.05,
			MaxRunningExperiments: 10,
			AutoStopOnSignificance: false,
		},
	}
}

// CreateExperiment creates a new A/B test experiment
func (s *ABTestingService) CreateExperiment(req CreateExperimentRequest) (*abExperiment, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("experiment name is required")
	}

	if len(req.Variants) < 2 {
		return nil, fmt.Errorf("at least 2 variants are required")
	}

	// Validate variants have exactly one control
	controlCount := 0
	totalWeight := 0.0
	for _, v := range req.Variants {
		if v.IsControl {
			controlCount++
		}
		totalWeight += v.Weight
	}

	if controlCount != 1 {
		return nil, fmt.Errorf("exactly one control variant is required")
	}

	// Normalize weights if they don't sum to 1
	if math.Abs(totalWeight-1.0) > 0.01 {
		for i := range req.Variants {
			req.Variants[i].Weight /= totalWeight
		}
	}

	now := time.Now()
	exp := &abExperiment{
		ID:                uuid.New().String(),
		Name:              req.Name,
		Description:       req.Description,
		Type:              req.Type,
		Status:            "draft",
		TrafficAllocation: req.TrafficAllocation,
		CreatedAt:         now,
		UpdatedAt:         now,
		Metrics:           req.Metrics,
		Variants:          make([]*variant, len(req.Variants)),
	}

	if exp.Type == "" {
		exp.Type = "ab"
	}

	if exp.TrafficAllocation <= 0 {
		exp.TrafficAllocation = 1.0
	}

	if req.Duration > 0 {
		exp.EndDate = now.Add(req.Duration)
	}

	// Create variants
	for i, vr := range req.Variants {
		exp.Variants[i] = &variant{
			ID:          uuid.New().String(),
			Name:        vr.Name,
			Description: vr.Description,
			Weight:      vr.Weight,
			IsControl:   vr.IsControl,
			Config:      vr.Config,
			Metrics: &variantMetrics{
				CustomMetrics: make(map[string]float64),
			},
		}
	}

	s.experiments.Store(exp.ID, exp)
	return exp, nil
}

// StartExperiment starts an experiment
func (s *ABTestingService) StartExperiment(experimentID string) error {
	val, ok := s.experiments.Load(experimentID)
	if !ok {
		return fmt.Errorf("experiment not found: %s", experimentID)
	}

	exp := val.(*abExperiment)
	if exp.Status == "running" {
		return fmt.Errorf("experiment is already running")
	}

	// Check max running experiments
	runningCount := 0
	s.experiments.Range(func(key, value any) bool {
		if e := value.(*abExperiment); e.Status == "running" {
			runningCount++
		}
		return true
	})

	if runningCount >= s.config.MaxRunningExperiments {
		return fmt.Errorf("maximum running experiments (%d) reached", s.config.MaxRunningExperiments)
	}

	exp.Status = "running"
	exp.StartDate = time.Now()
	exp.UpdatedAt = time.Now()
	return nil
}

// StopExperiment stops an experiment
func (s *ABTestingService) StopExperiment(experimentID string) error {
	val, ok := s.experiments.Load(experimentID)
	if !ok {
		return fmt.Errorf("experiment not found: %s", experimentID)
	}

	exp := val.(*abExperiment)
	if exp.Status != "running" {
		return fmt.Errorf("experiment is not running")
	}

	exp.Status = "completed"
	exp.EndDate = time.Now()
	exp.UpdatedAt = time.Now()
	return nil
}

// GetVariantForUser assigns or retrieves a variant for a user
func (s *ABTestingService) GetVariantForUser(experimentID, userID string) (*variant, error) {
	val, ok := s.experiments.Load(experimentID)
	if !ok {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}

	exp := val.(*abExperiment)
	if exp.Status != "running" {
		return nil, fmt.Errorf("experiment is not running")
	}

	// Check traffic allocation
	if rand.Float64() > exp.TrafficAllocation {
		// User not included in experiment, return control
		for _, v := range exp.Variants {
			if v.IsControl {
				return v, nil
			}
		}
	}

	// Check if user already assigned
	assignmentKey := fmt.Sprintf("%s:%s", experimentID, userID)
	if cachedAssignment, exists := s.assignments.Load(assignmentKey); exists {
		variantID := cachedAssignment.(string)
		for _, v := range exp.Variants {
			if v.ID == variantID {
				return v, nil
			}
		}
	}

	// Assign variant based on weights (deterministic hash for consistency)
	hash := s.hashUserID(userID, experimentID)
	selectedVariant := s.selectVariantByWeight(exp.Variants, hash)

	s.assignments.Store(assignmentKey, selectedVariant.ID)
	return selectedVariant, nil
}

// RecordEvent records an event for a variant
func (s *ABTestingService) RecordEvent(experimentID, variantID, eventType string, value float64) error {
	val, ok := s.experiments.Load(experimentID)
	if !ok {
		return fmt.Errorf("experiment not found: %s", experimentID)
	}

	exp := val.(*abExperiment)

	var targetVariant *variant
	for _, v := range exp.Variants {
		if v.ID == variantID {
			targetVariant = v
			break
		}
	}

	if targetVariant == nil {
		return fmt.Errorf("variant not found: %s", variantID)
	}

	targetVariant.Metrics.mu.Lock()
	defer targetVariant.Metrics.mu.Unlock()

	switch eventType {
	case "impression":
		targetVariant.Metrics.Impressions++
	case "click":
		targetVariant.Metrics.Clicks++
	case "conversion":
		targetVariant.Metrics.Conversions++
		targetVariant.Metrics.Revenue += value
	case "revenue":
		targetVariant.Metrics.Revenue += value
	default:
		// Custom metric
		if targetVariant.Metrics.CustomMetrics == nil {
			targetVariant.Metrics.CustomMetrics = make(map[string]float64)
		}
		targetVariant.Metrics.CustomMetrics[eventType] += value
	}

	// Update calculated metrics
	if targetVariant.Metrics.Impressions > 0 {
		targetVariant.Metrics.CTR = float64(targetVariant.Metrics.Clicks) / float64(targetVariant.Metrics.Impressions)
		targetVariant.Metrics.ConversionRate = float64(targetVariant.Metrics.Conversions) / float64(targetVariant.Metrics.Impressions)
		targetVariant.Metrics.ARPU = targetVariant.Metrics.Revenue / float64(targetVariant.Metrics.Impressions)
	}

	// Check for auto-stop on significance
	if s.config.AutoStopOnSignificance && exp.Status == "running" {
		result := s.analyzeExperiment(exp)
		if result.IsSignificant {
			exp.Status = "completed"
			exp.WinningVariant = result.WinningVariant
			exp.EndDate = time.Now()
		}
	}

	return nil
}

// AnalyzeExperiment performs statistical analysis on an experiment
func (s *ABTestingService) AnalyzeExperiment(experimentID string) (*ExperimentResult, error) {
	val, ok := s.experiments.Load(experimentID)
	if !ok {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}

	exp := val.(*abExperiment)
	return s.analyzeExperiment(exp), nil
}

func (s *ABTestingService) analyzeExperiment(exp *abExperiment) *ExperimentResult {
	result := &ExperimentResult{
		ExperimentID:   exp.ID,
		ExperimentName: exp.Name,
		Status:         exp.Status,
		VariantResults: make([]VariantResult, 0, len(exp.Variants)),
	}

	if !exp.StartDate.IsZero() {
		if exp.Status == "running" {
			result.Duration = time.Since(exp.StartDate)
		} else {
			result.Duration = exp.EndDate.Sub(exp.StartDate)
		}
	}

	// Find control variant
	var controlVariant *variant
	var totalSamples int64
	for _, v := range exp.Variants {
		if v.IsControl {
			controlVariant = v
		}
		totalSamples += v.Metrics.Impressions
	}
	result.TotalSamples = totalSamples

	if controlVariant == nil || controlVariant.Metrics.Impressions == 0 {
		result.Recommendation = "Insufficient data - no control variant data"
		return result
	}

	controlCR := controlVariant.Metrics.ConversionRate
	bestLift := 0.0
	var bestVariant *variant

	// Analyze each variant
	for _, v := range exp.Variants {
		vr := VariantResult{
			VariantID:      v.ID,
			VariantName:    v.Name,
			IsControl:      v.IsControl,
			SampleSize:     v.Metrics.Impressions,
			ConversionRate: v.Metrics.ConversionRate,
			CTR:            v.Metrics.CTR,
			Revenue:        v.Metrics.Revenue,
		}

		if !v.IsControl && v.Metrics.Impressions > 0 {
			// Calculate lift over control
			if controlCR > 0 {
				vr.Lift = (v.Metrics.ConversionRate - controlCR) / controlCR * 100
			}

			// Calculate p-value using two-proportion z-test
			vr.PValue = s.calculatePValue(
				controlVariant.Metrics.Conversions, controlVariant.Metrics.Impressions,
				v.Metrics.Conversions, v.Metrics.Impressions,
			)

			// Calculate confidence interval
			se := s.calculateStandardError(v.Metrics.ConversionRate, v.Metrics.Impressions)
			z := 1.96 // 95% confidence
			vr.ConfidenceInterval.Lower = math.Max(0, v.Metrics.ConversionRate-z*se)
			vr.ConfidenceInterval.Upper = math.Min(1, v.Metrics.ConversionRate+z*se)

			// Track best performing variant
			if vr.Lift > bestLift && vr.PValue < s.config.SignificanceLevel {
				bestLift = vr.Lift
				bestVariant = v
			}
		} else if v.IsControl {
			// Confidence interval for control
			se := s.calculateStandardError(v.Metrics.ConversionRate, v.Metrics.Impressions)
			z := 1.96
			vr.ConfidenceInterval.Lower = math.Max(0, v.Metrics.ConversionRate-z*se)
			vr.ConfidenceInterval.Upper = math.Min(1, v.Metrics.ConversionRate+z*se)
		}

		result.VariantResults = append(result.VariantResults, vr)
	}

	// Determine winning variant and significance
	if bestVariant != nil {
		result.WinningVariant = bestVariant.ID
		result.IsSignificant = true
		result.ExpectedLift = bestLift
		result.Confidence = 1 - result.VariantResults[0].PValue

		// Calculate statistical power
		result.StatisticalPower = s.calculateStatisticalPower(
			controlCR,
			bestVariant.Metrics.ConversionRate,
			controlVariant.Metrics.Impressions,
			bestVariant.Metrics.Impressions,
		)

		if result.StatisticalPower >= 0.8 {
			result.Recommendation = fmt.Sprintf("Implement variant '%s' - statistically significant with %.1f%% lift", bestVariant.Name, bestLift)
			result.RiskAssessment = "Low risk - high statistical power"
		} else {
			result.Recommendation = fmt.Sprintf("Consider variant '%s' - significant but low power (%.1f%%)", bestVariant.Name, result.StatisticalPower*100)
			result.RiskAssessment = "Medium risk - consider extending experiment"
		}
	} else if totalSamples < int64(s.config.MinSampleSize) {
		result.Recommendation = fmt.Sprintf("Continue experiment - need at least %d samples (currently %d)", s.config.MinSampleSize, totalSamples)
		result.RiskAssessment = "Cannot assess - insufficient data"
	} else {
		result.Recommendation = "No significant difference detected - consider extending experiment or reviewing hypothesis"
		result.RiskAssessment = "Low risk - no change recommended"
	}

	return result
}

// GetExperiment retrieves an experiment by ID
func (s *ABTestingService) GetExperiment(experimentID string) (*abExperiment, error) {
	val, ok := s.experiments.Load(experimentID)
	if !ok {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}
	return val.(*abExperiment), nil
}

// ListExperiments returns all experiments with optional status filter
func (s *ABTestingService) ListExperiments(status string) []*abExperiment {
	var experiments []*abExperiment
	s.experiments.Range(func(key, value any) bool {
		exp := value.(*abExperiment)
		if status == "" || exp.Status == status {
			experiments = append(experiments, exp)
		}
		return true
	})

	// Sort by created date descending
	sort.Slice(experiments, func(i, j int) bool {
		return experiments[i].CreatedAt.After(experiments[j].CreatedAt)
	})

	return experiments
}

// GetBanditRecommendation uses Thompson Sampling for multi-armed bandit
func (s *ABTestingService) GetBanditRecommendation(experimentID, userID string) (*variant, error) {
	val, ok := s.experiments.Load(experimentID)
	if !ok {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}

	exp := val.(*abExperiment)
	if exp.Type != "bandit" {
		return s.GetVariantForUser(experimentID, userID)
	}

	// Thompson Sampling: Sample from Beta distribution for each variant
	bestSample := -1.0
	var bestVariant *variant

	for _, v := range exp.Variants {
		// Beta(successes + 1, failures + 1)
		alpha := float64(v.Metrics.Conversions + 1)
		beta := float64(v.Metrics.Impressions - v.Metrics.Conversions + 1)

		// Sample from Beta distribution using gamma distribution trick
		sample := s.sampleBeta(alpha, beta)

		if sample > bestSample {
			bestSample = sample
			bestVariant = v
		}
	}

	return bestVariant, nil
}

// UpdateConfig updates the A/B testing configuration
func (s *ABTestingService) UpdateConfig(config ABTestingConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// GetConfig returns the current configuration
func (s *ABTestingService) GetConfig() ABTestingConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// GetStats returns overall A/B testing statistics
func (s *ABTestingService) GetStats() map[string]any {
	stats := map[string]any{
		"total_experiments":     0,
		"running_experiments":   0,
		"completed_experiments": 0,
		"draft_experiments":     0,
		"total_variants":        0,
		"total_impressions":     int64(0),
		"total_conversions":     int64(0),
	}

	s.experiments.Range(func(key, value any) bool {
		exp := value.(*abExperiment)
		stats["total_experiments"] = stats["total_experiments"].(int) + 1
		stats["total_variants"] = stats["total_variants"].(int) + len(exp.Variants)

		switch exp.Status {
		case "running":
			stats["running_experiments"] = stats["running_experiments"].(int) + 1
		case "completed":
			stats["completed_experiments"] = stats["completed_experiments"].(int) + 1
		case "draft":
			stats["draft_experiments"] = stats["draft_experiments"].(int) + 1
		}

		for _, v := range exp.Variants {
			stats["total_impressions"] = stats["total_impressions"].(int64) + v.Metrics.Impressions
			stats["total_conversions"] = stats["total_conversions"].(int64) + v.Metrics.Conversions
		}

		return true
	})

	return stats
}

// Helper functions

func (s *ABTestingService) hashUserID(userID, experimentID string) float64 {
	// Simple hash for deterministic assignment
	combined := userID + experimentID
	hash := uint64(0)
	for _, c := range combined {
		hash = hash*31 + uint64(c)
	}
	return float64(hash%10000) / 10000.0
}

func (s *ABTestingService) selectVariantByWeight(variants []*variant, hash float64) *variant {
	cumulative := 0.0
	for _, v := range variants {
		cumulative += v.Weight
		if hash < cumulative {
			return v
		}
	}
	// Fallback to last variant
	return variants[len(variants)-1]
}

func (s *ABTestingService) calculatePValue(convA, samplesA, convB, samplesB int64) float64 {
	if samplesA == 0 || samplesB == 0 {
		return 1.0
	}

	// Two-proportion z-test
	p1 := float64(convA) / float64(samplesA)
	p2 := float64(convB) / float64(samplesB)

	// Pooled proportion
	pooled := float64(convA+convB) / float64(samplesA+samplesB)

	// Standard error
	se := math.Sqrt(pooled * (1 - pooled) * (1/float64(samplesA) + 1/float64(samplesB)))

	if se == 0 {
		return 1.0
	}

	// Z-score
	z := math.Abs(p1-p2) / se

	// Two-tailed p-value
	return 2 * (1 - abNormalCDF(z))
}

func (s *ABTestingService) calculateStandardError(p float64, n int64) float64 {
	if n == 0 {
		return 0
	}
	return math.Sqrt(p * (1 - p) / float64(n))
}

func (s *ABTestingService) calculateStatisticalPower(p1, p2 float64, n1, n2 int64) float64 {
	// Simplified power calculation
	if n1 == 0 || n2 == 0 {
		return 0
	}

	effect := math.Abs(p2 - p1)
	pooledVar := p1*(1-p1)/float64(n1) + p2*(1-p2)/float64(n2)

	if pooledVar == 0 {
		return 0
	}

	// Non-centrality parameter
	ncp := effect / math.Sqrt(pooledVar)

	// Approximate power using normal distribution
	criticalValue := 1.96 // 95% confidence
	power := 1 - abNormalCDF(criticalValue-ncp)

	return math.Min(1, math.Max(0, power))
}

func (s *ABTestingService) sampleBeta(alpha, beta float64) float64 {
	// Sample from Beta distribution using two Gamma samples
	x := s.sampleGamma(alpha, 1)
	y := s.sampleGamma(beta, 1)
	return x / (x + y)
}

func (s *ABTestingService) sampleGamma(shape, scale float64) float64 {
	// Marsaglia and Tsang's method for shape >= 1
	if shape < 1 {
		return s.sampleGamma(shape+1, scale) * math.Pow(rand.Float64(), 1/shape)
	}

	d := shape - 1.0/3.0
	c := 1.0 / math.Sqrt(9*d)

	for {
		var x, v float64
		for {
			x = rand.NormFloat64()
			v = 1 + c*x
			if v > 0 {
				break
			}
		}
		v = v * v * v
		u := rand.Float64()

		if u < 1-0.0331*(x*x)*(x*x) {
			return d * v * scale
		}
		if math.Log(u) < 0.5*x*x+d*(1-v+math.Log(v)) {
			return d * v * scale
		}
	}
}

// abNormalCDF calculates the cumulative distribution function of standard normal
// Prefixed with 'ab' to avoid collision with normalCDF in incrementality.go
func abNormalCDF(x float64) float64 {
	// Approximation using error function
	return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}
