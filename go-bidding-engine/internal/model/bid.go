package model

import (
	"time"

	"github.com/google/uuid"
)

// BidRequest represents an incoming RTB bid request
type BidRequest struct {
	ID          string                 `json:"id" binding:"required"`
	Timestamp   time.Time              `json:"timestamp"`
	PublisherID string                 `json:"publisher_id" binding:"required"`
	AdSlot      AdSlot                 `json:"ad_slot" binding:"required"`
	User        User                   `json:"user"`
	Device      Device                 `json:"device" binding:"required"`
	Context     map[string]interface{} `json:"context"`
}

// AdSlot represents the ad placement
type AdSlot struct {
	ID         string   `json:"id" binding:"required"`
	Dimensions []int    `json:"dimensions"` // [width, height]
	Position   string   `json:"position"`   // "above-fold", "below-fold"
	Formats    []string `json:"formats"`    // ["banner", "native", "video"]
}

// User represents the user/visitor
type User struct {
	ID         string   `json:"id"`
	Country    string   `json:"country"`
	Language   string   `json:"language"`
	Categories []string `json:"categories"` // Interest categories
	Age        int      `json:"age"`
	Gender     string   `json:"gender"`
}

// Device represents the user's device
type Device struct {
	Type       string `json:"type" binding:"required"` // "mobile", "desktop", "tablet"
	OS         string `json:"os"`                      // "ios", "android", "windows", "macos"
	Browser    string `json:"browser"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	DeviceID   string `json:"device_id"`
}

// BidResponse represents the bid response
type BidResponse struct {
	RequestID   string    `json:"request_id"`
	CampaignID  string    `json:"campaign_id"`
	BidPrice    float64   `json:"bid_price"`
	CreativeURL string    `json:"creative_url"`
	ImpressionURL string  `json:"impression_url"`
	ClickURL    string    `json:"click_url"`
	TTL         int       `json:"ttl"` // Time to live in seconds
	Timestamp   time.Time `json:"timestamp"`
}

// NoBidResponse indicates no suitable ad found
type NoBidResponse struct {
	RequestID string `json:"request_id"`
	Reason    string `json:"reason"`
}

// Campaign represents an active campaign in cache
type Campaign struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "cpm", "cpc", "cpa"
	BidPrice    float64                `json:"bid_price"`
	Budget      float64                `json:"budget"`
	Spent       float64                `json:"spent"`
	Targeting   Targeting              `json:"targeting"`
	CreativeURL string                 `json:"creative_url"`
	Status      string                 `json:"status"`
}

// Targeting represents campaign targeting criteria
type Targeting struct {
	Countries  []string `json:"countries"`
	Devices    []string `json:"devices"`
	OS         []string `json:"os"`
	Categories []string `json:"categories"`
	MinAge     int      `json:"min_age"`
	MaxAge     int      `json:"max_age"`
}

// BidResult represents the result of a bid evaluation
type BidResult struct {
	Campaign   *Campaign
	Score      float64
	BidPrice   float64
	MatchScore float64
}

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	return uuid.New().String()
}

// IsMatch checks if campaign targeting matches bid request
func (c *Campaign) IsMatch(req *BidRequest) bool {
	// Check country
	if len(c.Targeting.Countries) > 0 {
		if !contains(c.Targeting.Countries, req.User.Country) {
			return false
		}
	}

	// Check device type
	if len(c.Targeting.Devices) > 0 {
		if !contains(c.Targeting.Devices, req.Device.Type) {
			return false
		}
	}

	// Check OS
	if len(c.Targeting.OS) > 0 {
		if !contains(c.Targeting.OS, req.Device.OS) {
			return false
		}
	}

	// Check age
	if c.Targeting.MinAge > 0 && req.User.Age < c.Targeting.MinAge {
		return false
	}
	if c.Targeting.MaxAge > 0 && req.User.Age > c.Targeting.MaxAge {
		return false
	}

	// Check categories (interest overlap)
	if len(c.Targeting.Categories) > 0 && len(req.User.Categories) > 0 {
		if !hasOverlap(c.Targeting.Categories, req.User.Categories) {
			return false
		}
	}

	return true
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func hasOverlap(slice1, slice2 []string) bool {
	for _, item := range slice1 {
		if contains(slice2, item) {
			return true
		}
	}
	return false
}
