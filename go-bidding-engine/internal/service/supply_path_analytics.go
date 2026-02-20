package service

import (
	"fmt"
	"sort"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// SupplyPathAnalyticsService provides analytics and insights for supply path optimization
type SupplyPathAnalyticsService struct {
	cache cache.Cache
}

// NewSupplyPathAnalyticsService creates a new SPO analytics service
func NewSupplyPathAnalyticsService(cache cache.Cache) *SupplyPathAnalyticsService {
	return &SupplyPathAnalyticsService{
		cache: cache,
	}
}

// GetSupplyChainMetrics returns aggregated metrics for supply chain performance
func (s *SupplyPathAnalyticsService) GetSupplyChainMetrics(timeRange string) (*model.SupplyChainMetrics, error) {
	return s.cache.GetSupplyChainMetrics(timeRange)
}

// GetServicePerformance returns performance metrics for a specific service
func (s *SupplyPathAnalyticsService) GetServicePerformance(serviceName string, timeRange string) (*model.ServiceMetrics, error) {
	return s.cache.GetServiceMetrics(serviceName, timeRange)
}

// GetBidPathAnalytics returns detailed analytics for a specific bid request
func (s *SupplyPathAnalyticsService) GetBidPathAnalytics(requestID string) (*model.BidPathAnalytics, error) {
	return s.cache.GetBidPathAnalytics(requestID)
}

// AnalyzeSupplyPathEfficiency analyzes the efficiency of the current supply path
func (s *SupplyPathAnalyticsService) AnalyzeSupplyPathEfficiency(timeRange string) (*model.SupplyPathOptimization, error) {
	metrics, err := s.GetSupplyChainMetrics(timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get supply chain metrics: %w", err)
	}

	optimization := &model.SupplyPathOptimization{
		Timestamp: time.Now(),
	}

	// If no metrics data available, return empty optimization with no suggestions
	if metrics == nil {
		optimization.Optimizations = []model.OptimizationSuggestion{}
		return optimization, nil
	}

	// Analyze service performance and suggest optimizations
	var suggestions []model.OptimizationSuggestion

	// Check for high-latency services
	for serviceName, serviceMetrics := range metrics.ServiceMetrics {
		if serviceMetrics.AvgLatencyMs > 200 { // High latency threshold
			suggestions = append(suggestions, model.OptimizationSuggestion{
				Type:        "cache",
				Service:     serviceName,
				Description: fmt.Sprintf("High latency detected (%.2fms avg). Consider caching responses.", serviceMetrics.AvgLatencyMs),
				Priority:    "medium",
				Savings:     serviceMetrics.AvgLatencyMs * 0.3, // Estimated 30% reduction
			})
		}

		if serviceMetrics.SuccessRate < 0.95 { // Low success rate
			suggestions = append(suggestions, model.OptimizationSuggestion{
				Type:        "circuit_breaker",
				Service:     serviceName,
				Description: fmt.Sprintf("Low success rate (%.2f%%). Circuit breaker may be tripping too frequently.", serviceMetrics.SuccessRate*100),
				Priority:    "high",
				Savings:     0, // Reliability improvement, not direct cost savings
			})
		}

		if serviceMetrics.TotalFees > 0.01 { // High fee services
			suggestions = append(suggestions, model.OptimizationSuggestion{
				Type:        "fee_negotiation",
				Service:     serviceName,
				Description: fmt.Sprintf("High fees detected ($%.4f per call). Consider direct integration.", serviceMetrics.TotalFees/float64(serviceMetrics.TotalCalls)),
				Priority:    "low",
				Savings:     serviceMetrics.TotalFees * 0.2, // Estimated 20% fee reduction
			})
		}
	}

	// Check overall path efficiency
	if metrics.PathEfficiency < 0.7 {
		suggestions = append(suggestions, model.OptimizationSuggestion{
			Type:        "direct_connection",
			Service:     "supply_path",
			Description: "Low overall path efficiency. Consider direct publisher connections to reduce hops.",
			Priority:    "high",
			Savings:     metrics.AvgTotalFees * 0.4, // Estimated 40% cost reduction
		})
	}

	optimization.Optimizations = suggestions

	// Calculate estimated savings
	var totalSavings float64
	for _, suggestion := range suggestions {
		totalSavings += suggestion.Savings
	}
	optimization.EstimatedSavings = totalSavings

	return optimization, nil
}

// GetTopBottlenecks identifies the top performance bottlenecks in the supply chain
func (s *SupplyPathAnalyticsService) GetTopBottlenecks(timeRange string, limit int) ([]*model.ServiceMetrics, error) {
	metrics, err := s.GetSupplyChainMetrics(timeRange)
	if err != nil {
		return nil, err
	}

	// If no metrics data available, return empty list
	if metrics == nil {
		return []*model.ServiceMetrics{}, nil
	}

	// Sort services by latency (highest first)
	var services []*model.ServiceMetrics
	for _, service := range metrics.ServiceMetrics {
		services = append(services, &service)
	}

	// Simple sort by latency (in production, you'd use a more sophisticated ranking)
	for i := 0; i < len(services)-1; i++ {
		for j := i + 1; j < len(services); j++ {
			if services[j].AvgLatencyMs > services[i].AvgLatencyMs {
				services[i], services[j] = services[j], services[i]
			}
		}
	}

	if limit > 0 && len(services) > limit {
		services = services[:limit]
	}

	return services, nil
}

// GetCostAnalysis provides cost analysis for the supply chain
func (s *SupplyPathAnalyticsService) GetCostAnalysis(timeRange string) (map[string]float64, error) {
	metrics, err := s.GetSupplyChainMetrics(timeRange)
	if err != nil {
		return nil, err
	}

	costs := make(map[string]float64)

	// If no metrics data available, return empty costs
	if metrics == nil {
		return costs, nil
	}

	costs["total_fees"] = metrics.AvgTotalFees * float64(metrics.TotalRequests)
	costs["avg_fee_per_request"] = metrics.AvgTotalFees
	costs["fee_efficiency"] = 1.0 - (metrics.AvgTotalFees / 0.01) // Lower is better (assuming $0.01 is target)

	return costs, nil
}

// AnalyzeDirectPublisherOpportunities analyzes opportunities for direct publisher relationships
func (s *SupplyPathAnalyticsService) AnalyzeDirectPublisherOpportunities(timeRange string) (*model.DirectPublisherAnalysis, error) {
	metrics, err := s.GetSupplyChainMetrics(timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get supply chain metrics: %w", err)
	}

	analysis := &model.DirectPublisherAnalysis{
		TimeRange:     timeRange,
		Timestamp:     time.Now(),
		CurrentHops:   3, // Average hops in current supply chain
		Opportunities: []model.DirectPublisherOpportunity{},
	}

	// If no metrics data available, return empty analysis
	if metrics == nil {
		return analysis, nil
	}

	// Analyze each service for direct connection potential
	for serviceName, serviceMetrics := range metrics.ServiceMetrics {
		if serviceMetrics.SuccessRate < 0.95 {
			// Low success rate indicates potential for direct connection
			opportunity := model.DirectPublisherOpportunity{
				ServiceName:        serviceName,
				CurrentFeeRate:     serviceMetrics.TotalFees / float64(serviceMetrics.TotalCalls),
				EstimatedDirectFee: serviceMetrics.TotalFees * 0.3, // Estimated 70% reduction
				SuccessRate:        serviceMetrics.SuccessRate,
				MonthlyVolume:      serviceMetrics.TotalCalls * 30, // Rough monthly estimate
				Priority:           "high",
				RiskLevel:          "medium",
			}

			// Calculate ROI
			currentMonthlyCost := serviceMetrics.TotalFees * 30
			directMonthlyCost := opportunity.EstimatedDirectFee * 30
			opportunity.EstimatedSavings = currentMonthlyCost - directMonthlyCost
			opportunity.ROI = (opportunity.EstimatedSavings / currentMonthlyCost) * 100

			analysis.Opportunities = append(analysis.Opportunities, opportunity)
		}
	}

	// Sort by potential savings
	sort.Slice(analysis.Opportunities, func(i, j int) bool {
		return analysis.Opportunities[i].EstimatedSavings > analysis.Opportunities[j].EstimatedSavings
	})

	return analysis, nil
}

// CalculateCostBenefitAnalysis performs detailed cost-benefit analysis for supply path changes
func (s *SupplyPathAnalyticsService) CalculateCostBenefitAnalysis(timeRange string) (*model.CostBenefitAnalysis, error) {
	metrics, err := s.GetSupplyChainMetrics(timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get supply chain metrics: %w", err)
	}

	// If no metrics data available, return empty analysis
	if metrics == nil {
		return &model.CostBenefitAnalysis{
			TimeRange: timeRange,
			Timestamp: time.Now(),
			Scenarios: []model.OptimizationScenario{},
		}, nil
	}

	analysis := &model.CostBenefitAnalysis{
		TimeRange:         timeRange,
		Timestamp:         time.Now(),
		CurrentTotalCost:  metrics.AvgTotalFees * float64(metrics.TotalRequests),
		CurrentWinRate:    metrics.WinRate,
		CurrentAvgLatency: metrics.AvgLatencyMs,
		Scenarios:         []model.OptimizationScenario{},
	}

	// Scenario 1: Direct Publisher Connections
	directScenario := model.OptimizationScenario{
		Name:                      "Direct Publisher Connections",
		Description:               "Bypass intermediaries and connect directly with publishers",
		EstimatedCostReduction:    metrics.AvgTotalFees * 0.6, // 60% cost reduction
		EstimatedLatencyReduction: metrics.AvgLatencyMs * 0.4, // 40% faster
		RiskLevel:                 "medium",
		ImplementationEffort:      "high",
		TimeToValue:               "3-6 months",
	}

	directScenario.NetBenefit = directScenario.EstimatedCostReduction * float64(metrics.TotalRequests) * 30 // Monthly benefit
	directScenario.BreakEvenMonths = 2                                                                      // Estimated implementation cost recovery
	analysis.Scenarios = append(analysis.Scenarios, directScenario)

	// Scenario 2: Service Optimization
	optimizationScenario := model.OptimizationScenario{
		Name:                      "Service Performance Optimization",
		Description:               "Optimize existing services for better performance and cost efficiency",
		EstimatedCostReduction:    metrics.AvgTotalFees * 0.25, // 25% cost reduction
		EstimatedLatencyReduction: metrics.AvgLatencyMs * 0.2,  // 20% faster
		RiskLevel:                 "low",
		ImplementationEffort:      "medium",
		TimeToValue:               "1-3 months",
	}

	optimizationScenario.NetBenefit = optimizationScenario.EstimatedCostReduction * float64(metrics.TotalRequests) * 30
	optimizationScenario.BreakEvenMonths = 1
	analysis.Scenarios = append(analysis.Scenarios, optimizationScenario)

	// Scenario 3: Hybrid Approach
	hybridScenario := model.OptimizationScenario{
		Name:                      "Hybrid Optimization",
		Description:               "Combine direct connections with service optimizations",
		EstimatedCostReduction:    metrics.AvgTotalFees * 0.75, // 75% cost reduction
		EstimatedLatencyReduction: metrics.AvgLatencyMs * 0.5,  // 50% faster
		RiskLevel:                 "medium",
		ImplementationEffort:      "high",
		TimeToValue:               "2-4 months",
	}

	hybridScenario.NetBenefit = hybridScenario.EstimatedCostReduction * float64(metrics.TotalRequests) * 30
	hybridScenario.BreakEvenMonths = 3
	analysis.Scenarios = append(analysis.Scenarios, hybridScenario)

	// Sort scenarios by net benefit
	sort.Slice(analysis.Scenarios, func(i, j int) bool {
		return analysis.Scenarios[i].NetBenefit > analysis.Scenarios[j].NetBenefit
	})

	return analysis, nil
}
