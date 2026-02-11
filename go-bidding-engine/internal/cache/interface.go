package cache

import "github.com/taskirx/go-bidding-engine/internal/model"

// Cache defines the interface for data access
type Cache interface {
	GetActiveCampaigns() ([]*model.Campaign, error)
	SetActiveCampaigns(campaigns []*model.Campaign) error
	GetCampaign(campaignID string) (*model.Campaign, error)
	SetCampaign(campaign *model.Campaign) error
	
	IncrementBidCount() error
	IncrementWinCount() error
	GetBidCount() (int64, error)
	GetWinCount() (int64, error)
	RecordLatency(latencyMs float64) error
	GetAverageLatency() (float64, error)
	
	SetUserSegments(userID string, segments []string) error
	GetUserSegments(userID string) ([]string, error)
	
	SetGeoRules(countryCode string, rules map[string]interface{}) error
	GetGeoRules(countryCode string) (map[string]interface{}, error)
	
	IncrementCampaignSpend(campaignID string, amount float64) (float64, error)
	GetCampaignSpend(campaignID string) (float64, error)
}
