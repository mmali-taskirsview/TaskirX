package model

import "time"

// AI Service Request/Response Structures

type AIMatchRequest struct {
	RequestID string           `json:"request_id"`
	Timestamp time.Time        `json:"timestamp"`
	User      AIUserProfile    `json:"user"`
	AdSlot    AIAdSlotInfo     `json:"ad_slot"`
	Context   AICampaignContext `json:"campaign_context"`
	Strategy  string           `json:"strategy"`
	MaxResults int             `json:"max_results"`
}

type AIUserProfile struct {
	UserID     string   `json:"user_id"`
	Country    string   `json:"country"`
	DeviceType string   `json:"device_type"`
	Categories []string `json:"categories"`
}

type AIAdSlotInfo struct {
	SlotID     string `json:"slot_id"`
	Dimensions []int  `json:"dimensions"` // [width, height]
	Format     string `json:"format"`
}

type AICampaignContext struct {
	PublisherID string `json:"publisher_id"`
}

type AIMatchResponse struct {
	Recommendations []AIAdRecommendation `json:"recommendations"`
}

type AIAdRecommendation struct {
	CampaignID   string  `json:"campaign_id"`
	OverallScore float64 `json:"overall_score"`
	BidPrice     float64 `json:"bid_price"`
}

// Fraud Service Structs

type FraudCheckRequest struct {
	RequestID   string                 `json:"request_id"`
	Timestamp   time.Time              `json:"timestamp"`
	IPAddress   string                 `json:"ip_address"`
	CampaignID  string                 `json:"campaign_id"` // Optional in early stage?
	PublisherID string                 `json:"publisher_id"`
	AdvertiserID string                `json:"advertiser_id"`
	Device      FraudDeviceInfo        `json:"device"`
	Geo         FraudGeoInfo           `json:"geo"`
}

type FraudDeviceInfo struct {
	Type      string `json:"type"`
	OS        string `json:"os"`
	UserAgent string `json:"user_agent"`
}

type FraudGeoInfo struct {
	Country string `json:"country"`
}

type FraudCheckResponse struct {
	RequestID         string   `json:"request_id"`
	IsFraud           bool     `json:"is_fraud"`
	FraudScore        float64  `json:"fraud_score"`
	RiskLevel         string   `json:"risk_level"`
	RecommendedAction string   `json:"recommended_action"` // allow, flag, block
}

// Bid Optimization Structs

type BidOptimizationRequest struct {
	RequestID string              `json:"request_id"`
	Timestamp time.Time           `json:"timestamp"`
	Context   OptimizationContext `json:"context"`
	Strategy  string              `json:"strategy"` // e.g. "maximize_conversions"
}

type OptimizationContext struct {
	CampaignID  string              `json:"campaign_id"`
	BaseBid     float64             `json:"base_bid"`
	Performance CampaignPerformance `json:"performance"`
	Budget      BudgetStatus        `json:"budget_status"`
	HourOfDay   int                 `json:"hour_of_day"`
	DayOfWeek   int                 `json:"day_of_week"`
}

type CampaignPerformance struct {
	CampaignID  string  `json:"campaign_id"`
	WinRate     float64 `json:"win_rate"`
	CTR         float64 `json:"ctr"`
}

type BudgetStatus struct {
	CampaignID   string  `json:"campaign_id"`
	DailyBudget  float64 `json:"daily_budget"`
	TodaySpend   float64 `json:"today_spend"`
	PacingRatio  float64 `json:"pacing_ratio"`
}

type BidRecommendation struct {
	RequestID       string   `json:"request_id"`
	RecommendedBid  float64  `json:"recommended_bid"`
	BidMultiplier   float64  `json:"bid_multiplier"`
	Confidence      float64  `json:"confidence"`
	Reasoning       []string `json:"reasoning"`
}
