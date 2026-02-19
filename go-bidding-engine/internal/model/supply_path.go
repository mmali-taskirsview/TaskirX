package model

import (
	"time"
)

// SupplyPathHop represents a single hop in the bid supply path
type SupplyPathHop struct {
	ServiceName  string    `json:"service_name"`  // e.g., "fraud-detection", "ad-matching", "bid-optimizer"
	ServiceType  string    `json:"service_type"`  // e.g., "internal", "external", "cache"
	Endpoint     string    `json:"endpoint"`      // API endpoint called
	RequestSize  int       `json:"request_size"`  // bytes sent
	ResponseSize int       `json:"response_size"` // bytes received
	LatencyMs    int64     `json:"latency_ms"`    // response time in milliseconds
	StatusCode   int       `json:"status_code"`   // HTTP status code
	Success      bool      `json:"success"`       // whether the call succeeded
	ErrorMessage string    `json:"error_message"` // error details if failed
	Fee          float64   `json:"fee"`           // cost/fees for this hop
	Timestamp    time.Time `json:"timestamp"`     // when this hop occurred
	Sequence     int       `json:"sequence"`      // order in the supply path (1, 2, 3...)
}

// BidPathAnalytics represents the complete analytics for a bid request's supply path
type BidPathAnalytics struct {
	RequestID      string                 `json:"request_id"`
	PublisherID    string                 `json:"publisher_id"`
	AdSlotID       string                 `json:"ad_slot_id"`
	TotalLatencyMs int64                  `json:"total_latency_ms"`
	TotalHops      int                    `json:"total_hops"`
	TotalFees      float64                `json:"total_fees"`
	FinalBidPrice  float64                `json:"final_bid_price"`
	DealID         string                 `json:"deal_id,omitempty"`
	WonAuction     bool                   `json:"won_auction"`
	CampaignID     string                 `json:"campaign_id,omitempty"`
	Hops           []SupplyPathHop        `json:"hops"`
	Metadata       map[string]interface{} `json:"metadata"` // additional context
	Timestamp      time.Time              `json:"timestamp"`
}

// SupplyChainMetrics aggregates metrics across multiple bid requests
type SupplyChainMetrics struct {
	TimeRange      string                    `json:"time_range"` // e.g., "1h", "24h", "7d"
	TotalRequests  int64                     `json:"total_requests"`
	SuccessfulBids int64                     `json:"successful_bids"`
	WinRate        float64                   `json:"win_rate"`
	AvgLatencyMs   float64                   `json:"avg_latency_ms"`
	AvgTotalFees   float64                   `json:"avg_total_fees"`
	ServiceMetrics map[string]ServiceMetrics `json:"service_metrics"`
	PathEfficiency float64                   `json:"path_efficiency"` // efficiency score 0-1
	Timestamp      time.Time                 `json:"timestamp"`
}

// ServiceMetrics represents performance metrics for a specific service
type ServiceMetrics struct {
	ServiceName  string  `json:"service_name"`
	TotalCalls   int64   `json:"total_calls"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	ErrorRate    float64 `json:"error_rate"`
	TotalFees    float64 `json:"total_fees"`
}

// SupplyPathOptimization represents optimization recommendations
type SupplyPathOptimization struct {
	RequestID        string                   `json:"request_id"`
	CurrentPath      []SupplyPathHop          `json:"current_path"`
	RecommendedPath  []SupplyPathHop          `json:"recommended_path"`
	Optimizations    []OptimizationSuggestion `json:"optimizations"`
	EstimatedSavings float64                  `json:"estimated_savings"`
	Timestamp        time.Time                `json:"timestamp"`
}

// OptimizationSuggestion represents a specific optimization recommendation
type OptimizationSuggestion struct {
	Type        string  `json:"type"` // "cache", "circuit_breaker", "direct_connection", "fee_negotiation"
	Service     string  `json:"service"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"` // "high", "medium", "low"
	Savings     float64 `json:"savings"`  // estimated cost/time savings
}

// DirectPublisherAnalysis represents analysis of direct publisher relationship opportunities
type DirectPublisherAnalysis struct {
	TimeRange     string                       `json:"time_range"`
	Timestamp     time.Time                    `json:"timestamp"`
	CurrentHops   int                          `json:"current_hops"`
	Opportunities []DirectPublisherOpportunity `json:"opportunities"`
}

// DirectPublisherOpportunity represents a specific opportunity for direct publisher connection
type DirectPublisherOpportunity struct {
	ServiceName        string  `json:"service_name"`
	CurrentFeeRate     float64 `json:"current_fee_rate"`
	EstimatedDirectFee float64 `json:"estimated_direct_fee"`
	SuccessRate        float64 `json:"success_rate"`
	MonthlyVolume      int64   `json:"monthly_volume"`
	Priority           string  `json:"priority"`
	RiskLevel          string  `json:"risk_level"`
	EstimatedSavings   float64 `json:"estimated_savings"`
	ROI                float64 `json:"roi"` // Return on Investment percentage
}

// CostBenefitAnalysis represents detailed cost-benefit analysis for optimization scenarios
type CostBenefitAnalysis struct {
	TimeRange         string                 `json:"time_range"`
	Timestamp         time.Time              `json:"timestamp"`
	CurrentTotalCost  float64                `json:"current_total_cost"`
	CurrentWinRate    float64                `json:"current_win_rate"`
	CurrentAvgLatency float64                `json:"current_avg_latency"`
	Scenarios         []OptimizationScenario `json:"scenarios"`
}

// OptimizationScenario represents a specific optimization scenario with cost-benefit analysis
type OptimizationScenario struct {
	Name                      string  `json:"name"`
	Description               string  `json:"description"`
	EstimatedCostReduction    float64 `json:"estimated_cost_reduction"`
	EstimatedLatencyReduction float64 `json:"estimated_latency_reduction"`
	RiskLevel                 string  `json:"risk_level"`
	ImplementationEffort      string  `json:"implementation_effort"`
	TimeToValue               string  `json:"time_to_value"`
	NetBenefit                float64 `json:"net_benefit"`
	BreakEvenMonths           int     `json:"break_even_months"`
}
