package cache

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// MockCache is an in-memory implementation of the Cache interface for testing
type MockCache struct {
	mu   sync.RWMutex
	data map[string]string

	bidCount  int64
	winCount  int64
	latencies []float64

	campaignClicks      map[string]int64
	campaignImpressions map[string]int64
	campaignBids        map[string]int64
	campaignWins        map[string]int64
	campaignSpend       map[string]float64

	userSegments  map[string][]string
	geoRules      map[string]map[string]interface{}
	bidFormats    map[string]int64
	publisherBids map[string][]bidAttempt

	segmentImpressions map[string]map[string]int64
	segmentClicks      map[string]map[string]int64

	bidLandscape map[string]map[string]int64

	userFrequency  map[string]int64
	seenRequests   map[string]bool
	publisherFraud map[string]int64

	impressions    map[string]impressionData
	clicks         map[string]clickData
	touchpoints    map[string][]model.Touchpoint
	userEvents     map[string]map[string][]string
	deviceGraph    map[string][]string
	primaryUserMap map[string]string
	bidPaths       map[string]*model.BidPathAnalytics
}

type bidAttempt struct {
	price float64
	won   bool
}

type impressionData struct {
	requestID string
	timestamp time.Time
}

type clickData struct {
	requestID string
	timestamp time.Time
}

// NewMockCache creates a new mock cache for testing
func NewMockCache() *MockCache {
	return &MockCache{
		data:                make(map[string]string),
		campaignClicks:      make(map[string]int64),
		campaignImpressions: make(map[string]int64),
		campaignBids:        make(map[string]int64),
		campaignWins:        make(map[string]int64),
		campaignSpend:       make(map[string]float64),
		userSegments:        make(map[string][]string),
		geoRules:            make(map[string]map[string]interface{}),
		bidFormats:          make(map[string]int64),
		publisherBids:       make(map[string][]bidAttempt),
		segmentImpressions:  make(map[string]map[string]int64),
		segmentClicks:       make(map[string]map[string]int64),
		bidLandscape:        make(map[string]map[string]int64),
		userFrequency:       make(map[string]int64),
		seenRequests:        make(map[string]bool),
		publisherFraud:      make(map[string]int64),
		impressions:         make(map[string]impressionData),
		clicks:              make(map[string]clickData),
		touchpoints:         make(map[string][]model.Touchpoint),
		userEvents:          make(map[string]map[string][]string),
		deviceGraph:         make(map[string][]string),
		primaryUserMap:      make(map[string]string),
		bidPaths:            make(map[string]*model.BidPathAnalytics),
	}
}

// GetActiveCampaigns retrieves all active campaigns from cache
func (m *MockCache) GetActiveCampaigns() ([]*model.Campaign, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val, ok := m.data["campaigns:active"]
	if !ok {
		return []*model.Campaign{}, nil
	}

	var campaigns []*model.Campaign
	if err := json.Unmarshal([]byte(val), &campaigns); err != nil {
		return nil, err
	}
	return campaigns, nil
}

// SetActiveCampaigns caches active campaigns
func (m *MockCache) SetActiveCampaigns(campaigns []*model.Campaign) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.Marshal(campaigns)
	if err != nil {
		return err
	}
	m.data["campaigns:active"] = string(data)
	return nil
}

// GetCampaign retrieves a specific campaign by ID
func (m *MockCache) GetCampaign(campaignID string) (*model.Campaign, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("campaign:%s", campaignID)
	val, ok := m.data[key]
	if !ok {
		return nil, nil
	}

	var campaign model.Campaign
	if err := json.Unmarshal([]byte(val), &campaign); err != nil {
		return nil, err
	}
	return &campaign, nil
}

// SetCampaign caches a specific campaign
func (m *MockCache) SetCampaign(campaign *model.Campaign) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("campaign:%s", campaign.ID)
	data, err := json.Marshal(campaign)
	if err != nil {
		return err
	}
	m.data[key] = string(data)
	return nil
}

// IncrementBidCount increments the bid counter
func (m *MockCache) IncrementBidCount() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bidCount++
	return nil
}

// IncrementWinCount increments the win counter
func (m *MockCache) IncrementWinCount() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.winCount++
	return nil
}

// GetBidCount gets total bid count
func (m *MockCache) GetBidCount() (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bidCount, nil
}

// GetWinCount gets total win count
func (m *MockCache) GetWinCount() (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.winCount, nil
}

// RecordLatency records bid processing latency
func (m *MockCache) RecordLatency(latencyMs float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.latencies = append(m.latencies, latencyMs)
	return nil
}

// GetAverageLatency gets the average latency
func (m *MockCache) GetAverageLatency() (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.latencies) == 0 {
		return 0, nil
	}

	var sum float64
	for _, l := range m.latencies {
		sum += l
	}
	return sum / float64(len(m.latencies)), nil
}

// SetUserSegments sets user segments
func (m *MockCache) SetUserSegments(userID string, segments []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userSegments[userID] = segments
	return nil
}

// GetUserSegments gets user segments
func (m *MockCache) GetUserSegments(userID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.userSegments[userID], nil
}

// SetGeoRules sets geo rules
func (m *MockCache) SetGeoRules(countryCode string, rules map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.geoRules[countryCode] = rules
	return nil
}

// GetGeoRules gets geo rules
func (m *MockCache) GetGeoRules(countryCode string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.geoRules[countryCode], nil
}

// IncrementCampaignSpend increments campaign spend
func (m *MockCache) IncrementCampaignSpend(campaignID string, amount float64) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.campaignSpend[campaignID] += amount
	return m.campaignSpend[campaignID], nil
}

// GetCampaignSpend gets campaign spend
func (m *MockCache) GetCampaignSpend(campaignID string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.campaignSpend[campaignID], nil
}

// IncrementBidFormat increments bid format counter
func (m *MockCache) IncrementBidFormat(format string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bidFormats[format]++
	return nil
}

// GetBidFormats gets all bid format counts
func (m *MockCache) GetBidFormats() (map[string]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]int64)
	for k, v := range m.bidFormats {
		result[k] = v
	}
	return result, nil
}

// Get retrieves a value by key
func (m *MockCache) Get(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[key], nil
}

// Set stores a value
func (m *MockCache) Set(key string, value interface{}, ttl int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch v := value.(type) {
	case string:
		m.data[key] = v
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		m.data[key] = string(data)
	}
	return nil
}

// IncrementPublisherFraud increments fraud counter
func (m *MockCache) IncrementPublisherFraud(publisherID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publisherFraud[publisherID]++
	return nil
}

// IsRequestDuplicate checks for duplicate requests
func (m *MockCache) IsRequestDuplicate(requestID string, ttlSeconds int) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.seenRequests[requestID] {
		return true, nil
	}
	m.seenRequests[requestID] = true
	return false, nil
}

// IncrementUserFrequency increments user frequency
func (m *MockCache) IncrementUserFrequency(userID, campaignID string, windowSecs int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", userID, campaignID)
	m.userFrequency[key]++
	return m.userFrequency[key], nil
}

// GetUserFrequency gets user frequency
func (m *MockCache) GetUserFrequency(userID, campaignID string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", userID, campaignID)
	return m.userFrequency[key], nil
}

// GetCampaignCTR gets campaign CTR
func (m *MockCache) GetCampaignCTR(campaignID string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	imps := m.campaignImpressions[campaignID]
	clicks := m.campaignClicks[campaignID]
	if imps == 0 {
		return 0, nil
	}
	return float64(clicks) / float64(imps), nil
}

// GetCampaignWinRate gets campaign win rate
func (m *MockCache) GetCampaignWinRate(campaignID string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bids := m.campaignBids[campaignID]
	wins := m.campaignWins[campaignID]
	if bids == 0 {
		return 0, nil
	}
	return float64(wins) / float64(bids), nil
}

// IncrementCampaignClicks increments campaign clicks
func (m *MockCache) IncrementCampaignClicks(campaignID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.campaignClicks[campaignID]++
	return nil
}

// IncrementCampaignImpressions increments campaign impressions
func (m *MockCache) IncrementCampaignImpressions(campaignID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.campaignImpressions[campaignID]++
	return nil
}

// IncrementCampaignBids increments campaign bids
func (m *MockCache) IncrementCampaignBids(campaignID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.campaignBids[campaignID]++
	return nil
}

// IncrementCampaignWins increments campaign wins
func (m *MockCache) IncrementCampaignWins(campaignID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.campaignWins[campaignID]++
	return nil
}

// RecordBidInBucket records a bid in a price bucket
func (m *MockCache) RecordBidInBucket(priceBucket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.bidLandscape[priceBucket] == nil {
		m.bidLandscape[priceBucket] = make(map[string]int64)
	}
	m.bidLandscape[priceBucket]["bids"]++
	return nil
}

// RecordWinInBucket records a win in a price bucket
func (m *MockCache) RecordWinInBucket(priceBucket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.bidLandscape[priceBucket] == nil {
		m.bidLandscape[priceBucket] = make(map[string]int64)
	}
	m.bidLandscape[priceBucket]["wins"]++
	return nil
}

// GetBidLandscape returns bid landscape data
func (m *MockCache) GetBidLandscape() (map[string]map[string]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]map[string]int64)
	for k, v := range m.bidLandscape {
		result[k] = make(map[string]int64)
		for ik, iv := range v {
			result[k][ik] = iv
		}
	}
	return result, nil
}

// IncrementSegmentImpressions increments segment impressions
func (m *MockCache) IncrementSegmentImpressions(segmentType, segmentValue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := segmentType
	if m.segmentImpressions[key] == nil {
		m.segmentImpressions[key] = make(map[string]int64)
	}
	m.segmentImpressions[key][segmentValue]++
	return nil
}

// IncrementSegmentClicks increments segment clicks
func (m *MockCache) IncrementSegmentClicks(segmentType, segmentValue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := segmentType
	if m.segmentClicks[key] == nil {
		m.segmentClicks[key] = make(map[string]int64)
	}
	m.segmentClicks[key][segmentValue]++
	return nil
}

// GetSegmentPerformance gets segment performance
func (m *MockCache) GetSegmentPerformance(segmentType string) (map[string]map[string]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]map[string]int64)

	imps := m.segmentImpressions[segmentType]
	clicks := m.segmentClicks[segmentType]

	for segment := range imps {
		if result[segment] == nil {
			result[segment] = make(map[string]int64)
		}
		result[segment]["impressions"] = imps[segment]
	}

	for segment := range clicks {
		if result[segment] == nil {
			result[segment] = make(map[string]int64)
		}
		result[segment]["clicks"] = clicks[segment]
	}

	return result, nil
}

// RecordPublisherBidAttempt records a bid attempt for a publisher
func (m *MockCache) RecordPublisherBidAttempt(publisherID string, bidPrice float64, won bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publisherBids[publisherID] = append(m.publisherBids[publisherID], bidAttempt{price: bidPrice, won: won})
	return nil
}

// GetOptimalBidFloor calculates optimal bid floor
func (m *MockCache) GetOptimalBidFloor(publisherID string, targetWinRate float64) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bids := m.publisherBids[publisherID]
	if len(bids) == 0 {
		return 1.0, nil // Default floor
	}

	// Simple calculation: find price where we achieve target win rate
	var totalWins, totalBids int
	var sumWinPrice float64

	for _, b := range bids {
		totalBids++
		if b.won {
			totalWins++
			sumWinPrice += b.price
		}
	}

	if totalWins == 0 {
		return 1.0, nil
	}

	avgWinPrice := sumWinPrice / float64(totalWins)
	return avgWinPrice * 0.9, nil // Floor slightly below average win price
}

// RecordImpression records an impression
func (m *MockCache) RecordImpression(userID, campaignID, requestID string, ttlHours int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", userID, campaignID)
	m.impressions[key] = impressionData{requestID: requestID, timestamp: time.Now()}
	return nil
}

// RecordClick records a click
func (m *MockCache) RecordClick(userID, campaignID, requestID string, ttlHours int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", userID, campaignID)
	m.clicks[key] = clickData{requestID: requestID, timestamp: time.Now()}
	return nil
}

// GetAttribution gets attribution
func (m *MockCache) GetAttribution(userID, campaignID string) (string, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", userID, campaignID)

	// Check clicks first (CTA)
	if click, ok := m.clicks[key]; ok {
		return "CTA", click.requestID, nil
	}

	// Check impressions (VTA)
	if imp, ok := m.impressions[key]; ok {
		return "VTA", imp.requestID, nil
	}

	return "", "", nil
}

// RecordTouchpoint records a touchpoint
func (m *MockCache) RecordTouchpoint(userID, campaignID, touchpointType, requestID string, ttlDays int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, campaignID)
	tp := model.Touchpoint{
		Type:      touchpointType,
		RequestID: requestID,
		Timestamp: time.Now(),
	}
	m.touchpoints[key] = append(m.touchpoints[key], tp)
	return nil
}

// GetTouchpoints gets touchpoints
func (m *MockCache) GetTouchpoints(userID, campaignID string) ([]model.Touchpoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", userID, campaignID)
	return m.touchpoints[key], nil
}

// GetMultiTouchAttribution calculates multi-touch attribution
func (m *MockCache) GetMultiTouchAttribution(userID, campaignID, modelType string) ([]model.AttributionCredit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", userID, campaignID)
	touchpoints := m.touchpoints[key]

	if len(touchpoints) == 0 {
		return []model.AttributionCredit{}, nil
	}

	credits := make([]model.AttributionCredit, len(touchpoints))
	creditValue := 1.0 / float64(len(touchpoints)) // Linear model

	for i, tp := range touchpoints {
		credits[i] = model.AttributionCredit{
			Touchpoint: tp,
			Credit:     creditValue,
			Model:      modelType,
		}
	}

	return credits, nil
}

// RecordUserEvent records a user event
func (m *MockCache) RecordUserEvent(userID, campaignID, eventType string, ttlDays int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.userEvents[userID] == nil {
		m.userEvents[userID] = make(map[string][]string)
	}
	m.userEvents[userID][eventType] = append(m.userEvents[userID][eventType], campaignID)
	return nil
}

// GetUserEvents gets user events
func (m *MockCache) GetUserEvents(userID string, eventTypes []string) (map[string][]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]string)
	events := m.userEvents[userID]

	for _, et := range eventTypes {
		if campaigns, ok := events[et]; ok {
			result[et] = campaigns
		}
	}

	return result, nil
}

// HasUserEvent checks if user has an event
func (m *MockCache) HasUserEvent(userID, campaignID, eventType string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := m.userEvents[userID]
	if events == nil {
		return false, nil
	}

	campaigns := events[eventType]
	for _, c := range campaigns {
		if c == campaignID {
			return true, nil
		}
	}

	return false, nil
}

// LinkDevices links devices to a primary user
func (m *MockCache) LinkDevices(primaryUserID string, deviceIDs []string, ttlDays int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.deviceGraph[primaryUserID] = deviceIDs
	for _, deviceID := range deviceIDs {
		m.primaryUserMap[deviceID] = primaryUserID
	}
	return nil
}

// GetLinkedDevices gets linked devices
func (m *MockCache) GetLinkedDevices(deviceID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	primaryUserID := m.primaryUserMap[deviceID]
	if primaryUserID == "" {
		return []string{}, nil
	}
	return m.deviceGraph[primaryUserID], nil
}

// GetPrimaryUserID gets primary user ID
func (m *MockCache) GetPrimaryUserID(deviceID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.primaryUserMap[deviceID], nil
}

// GetCrossDeviceFrequency gets cross-device frequency
func (m *MockCache) GetCrossDeviceFrequency(primaryUserID, campaignID string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := m.deviceGraph[primaryUserID]
	var total int64

	for _, deviceID := range devices {
		key := fmt.Sprintf("%s:%s", deviceID, campaignID)
		total += m.userFrequency[key]
	}

	return total, nil
}

// StoreBidPathAnalytics stores bid path analytics
func (m *MockCache) StoreBidPathAnalytics(analytics *model.BidPathAnalytics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bidPaths[analytics.RequestID] = analytics
	return nil
}

// GetBidPathAnalytics gets bid path analytics
func (m *MockCache) GetBidPathAnalytics(requestID string) (*model.BidPathAnalytics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bidPaths[requestID], nil
}

// GetSupplyChainMetrics gets supply chain metrics
func (m *MockCache) GetSupplyChainMetrics(timeRange string) (*model.SupplyChainMetrics, error) {
	return &model.SupplyChainMetrics{
		TimeRange:      timeRange,
		TotalRequests:  100,
		SuccessfulBids: 70,
		WinRate:        0.7,
		AvgLatencyMs:   50.0,
		AvgTotalFees:   0.15,
		PathEfficiency: 0.85,
	}, nil
}

// GetServiceMetrics gets service metrics
func (m *MockCache) GetServiceMetrics(serviceName string, timeRange string) (*model.ServiceMetrics, error) {
	return &model.ServiceMetrics{
		ServiceName:  serviceName,
		TotalCalls:   1000,
		SuccessRate:  0.95,
		AvgLatencyMs: 25.0,
		ErrorRate:    0.05,
		TotalFees:    150.0,
	}, nil
}
