package model

import (
	"math"
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
	Type      string `json:"type" binding:"required"` // "mobile", "desktop", "tablet"
	OS        string `json:"os"`                      // "ios", "android", "windows", "macos"
	Browser   string `json:"browser"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	DeviceID  string `json:"device_id"`
	Geo       Geo    `json:"geo"`
}

// Geo represents geographic location
type Geo struct {
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Country string  `json:"country"`
	City    string  `json:"city"`
	Zip     string  `json:"zip"`
}

// BidResponse represents the bid response
type BidResponse struct {
	RequestID     string    `json:"request_id"`
	CampaignID    string    `json:"campaign_id"`
	BidPrice      float64   `json:"bid_price"`
	CreativeURL   string    `json:"creative_url"`
	ImpressionURL string    `json:"impression_url"`
	ClickURL      string    `json:"click_url"`
	TTL           int       `json:"ttl"`                 // Time to live in seconds
	AdMarkup      string    `json:"ad_markup,omitempty"` // HTML or VAST XML
	Timestamp     time.Time `json:"timestamp"`
}

// NoBidResponse indicates no suitable ad found
type NoBidResponse struct {
	RequestID string `json:"request_id"`
	Reason    string `json:"reason"`
}

// Campaign represents an active campaign in cache
type Campaign struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenantId"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "cpm", "cpc", "cpa"
	BidPrice  float64   `json:"bidPrice"`
	Budget    float64   `json:"budget"`
	Spent     float64   `json:"spent"`
	Targeting Targeting `json:"targeting"`
	Creative  Creative  `json:"creative"`
	Status    string    `json:"status"`
}

type Creative struct {
	Type        string `json:"type"` // "banner", "video", "native"
	URL         string `json:"url"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Duration    int    `json:"duration"` // seconds
	MimeType    string `json:"mimeType"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
	CTAText     string `json:"ctaText"`
}

// Targeting represents campaign targeting criteria
type Targeting struct {
	Countries  []string   `json:"countries"`
	Devices    []string   `json:"devices"`
	OS         []string   `json:"os"`
	Categories []string   `json:"categories"`
	GeoFences  []GeoFence `json:"geoFences"`
	MinAge     int        `json:"minAge"`
	MaxAge     int        `json:"maxAge"`
}

// GeoFence defines a circular region
type GeoFence struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Radius float64 `json:"radius"` // in km
	Name   string  `json:"name"`
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

	// Check GeoFences
	if len(c.Targeting.GeoFences) > 0 {
		// If request has no valid geo, we cannot match
		if req.Device.Geo.Lat == 0 && req.Device.Geo.Lon == 0 {
			return false
		}

		matchGeo := false
		for _, fence := range c.Targeting.GeoFences {
			dist := haversine(
				req.Device.Geo.Lat, req.Device.Geo.Lon,
				fence.Lat, fence.Lon,
			)
			if dist <= fence.Radius {
				matchGeo = true
				break
			}
		}
		if !matchGeo {
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

// haversine calculates distance between two points in km
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
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
