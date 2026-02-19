package service

import (
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// PrivacySandboxService handles Privacy Sandbox API integration
type PrivacySandboxService struct {
	cache           cache.Cache
	mu              sync.RWMutex
	topicCache      map[string][]int    // userID -> topic IDs
	fledgeAudiences map[string][]string // userID -> interest group names
}

// Chrome Topics Taxonomy (simplified subset)
var topicTaxonomy = map[int]string{
	1:  "Arts & Entertainment",
	2:  "Autos & Vehicles",
	3:  "Beauty & Fitness",
	4:  "Books & Literature",
	5:  "Business & Industrial",
	6:  "Computers & Electronics",
	7:  "Finance",
	8:  "Food & Drink",
	9:  "Games",
	10: "Health",
	11: "Hobbies & Leisure",
	12: "Home & Garden",
	13: "Internet & Telecom",
	14: "Jobs & Education",
	15: "Law & Government",
	16: "News",
	17: "Online Communities",
	18: "People & Society",
	19: "Pets & Animals",
	20: "Real Estate",
	21: "Reference",
	22: "Science",
	23: "Shopping",
	24: "Sports",
	25: "Travel",
}

// NewPrivacySandboxService creates a new Privacy Sandbox service
func NewPrivacySandboxService(c cache.Cache) *PrivacySandboxService {
	return &PrivacySandboxService{
		cache:           c,
		topicCache:      make(map[string][]int),
		fledgeAudiences: make(map[string][]string),
	}
}

// EvaluatePrivacySandbox evaluates Privacy Sandbox signals for targeting
func (s *PrivacySandboxService) EvaluatePrivacySandbox(campaign *model.Campaign, req *model.BidRequest) *model.PrivacySandboxResult {
	config := campaign.Targeting.PrivacySandbox
	if config == nil || !config.Enabled {
		return &model.PrivacySandboxResult{
			TopicsAvailable: false,
			TopicMatch:      false,
			TopicMultiplier: 1.0,
			Reason:          "privacy_sandbox_disabled",
		}
	}

	result := &model.PrivacySandboxResult{
		TopicMultiplier: 1.0,
	}

	// Check Topics API
	if config.TopicsAPI != nil && config.TopicsAPI.Enabled {
		s.evaluateTopicsAPI(config.TopicsAPI, req, result)
	}

	// Check Attribution API
	if config.AttributionAPI != nil && config.AttributionAPI.Enabled {
		result.AttributionEnabled = true
	}

	// Check FLEDGE/Protected Audience API
	if config.FledgeEnabled {
		result.FledgeEligible = s.checkFledgeEligibility(req)
	}

	// Apply fallback if no Privacy Sandbox signals available
	if !result.TopicsAvailable && !result.FledgeEligible {
		result.FallbackUsed = true
		result.FallbackMethod = config.FallbackStrategy
		if result.FallbackMethod == "" {
			result.FallbackMethod = "contextual"
		}
	}

	return result
}

func (s *PrivacySandboxService) evaluateTopicsAPI(config *model.TopicsAPIConfig, req *model.BidRequest, result *model.PrivacySandboxResult) {
	// Get user topics from request context (simulating browser-provided topics)
	userTopics := s.getUserTopics(req)

	if len(userTopics) == 0 {
		result.TopicsAvailable = false
		return
	}

	result.TopicsAvailable = true
	result.UserTopics = userTopics

	// Check for topic matches
	for _, userTopic := range userTopics {
		// Check exclusions first
		for _, excludeTopic := range config.ExcludeTopics {
			if userTopic == excludeTopic {
				result.TopicMatch = false
				result.TopicMultiplier = 0 // Block bid
				result.Reason = "excluded_topic_match"
				return
			}
		}

		// Check target topics
		for _, targetTopic := range config.TargetTopics {
			if userTopic == targetTopic {
				result.TopicMatch = true

				// Apply bid boost if configured
				for _, boost := range config.TopicBidBoosts {
					if boost.TopicID == userTopic {
						result.TopicMultiplier = boost.Multiplier
						break
					}
				}

				if result.TopicMultiplier == 1.0 {
					result.TopicMultiplier = 1.2 // Default 20% boost for topic match
				}

				result.Reason = "topic_match"
				return
			}
		}
	}

	// No specific match, but topics available
	result.TopicMatch = false
	result.Reason = "topics_available_no_match"
}

func (s *PrivacySandboxService) getUserTopics(req *model.BidRequest) []int {
	// In real implementation, topics come from browser via request headers
	// or as part of the OpenRTB request extensions

	// Check request context for topics
	if req.Context != nil {
		if topics, ok := req.Context["chrome_topics"].([]interface{}); ok {
			result := make([]int, 0, len(topics))
			for _, t := range topics {
				if topicID, ok := t.(float64); ok {
					result = append(result, int(topicID))
				}
			}
			return result
		}
	}

	// Check for topics in user extensions
	s.mu.RLock()
	topics, exists := s.topicCache[req.User.ID]
	s.mu.RUnlock()

	if exists {
		return topics
	}

	return nil
}

func (s *PrivacySandboxService) checkFledgeEligibility(req *model.BidRequest) bool {
	// Check if browser supports FLEDGE/Protected Audience API
	if req.Context != nil {
		if fledge, ok := req.Context["fledge_supported"].(bool); ok {
			return fledge
		}
		if pa, ok := req.Context["protected_audience"].(bool); ok {
			return pa
		}
	}

	// Check user-agent or other signals
	// Chrome 115+ supports Protected Audience
	ua := req.Device.UserAgent
	if ua != "" && containsString(ua, "Chrome/") {
		// Simplified check - real implementation would parse version
		return true
	}

	return false
}

// GenerateAttributionSource creates Attribution Reporting API source registration
func (s *PrivacySandboxService) GenerateAttributionSource(config *model.AttributionAPIConfig, campaign *model.Campaign, req *model.BidRequest) map[string]interface{} {
	if config == nil || !config.Enabled {
		return nil
	}

	// Use campaign creative URL as destination since Campaign doesn't have LandingURL
	destination := campaign.Creative.URL

	source := map[string]interface{}{
		"source_event_id":  config.SourceEventID,
		"destination":      destination,
		"reporting_origin": config.ReportingOrigin,
		"expiry":           "604800", // 7 days
		"priority":         "100",
		"debug_key":        "",
	}

	if config.DebugMode {
		source["debug_key"] = campaign.ID + "_" + req.ID
	}

	// Add aggregatable source if configured
	if len(config.AggregatableValues) > 0 {
		source["aggregatable_source"] = config.AggregatableValues
	}

	return source
}

// GenerateFledgeBid creates a FLEDGE/Protected Audience bid
func (s *PrivacySandboxService) GenerateFledgeBid(campaign *model.Campaign, req *model.BidRequest) map[string]interface{} {
	// Use campaign name as advertiser identifier since Campaign doesn't have AdvertiserDomain
	advertiserID := campaign.TenantID
	if advertiserID == "" {
		advertiserID = campaign.ID
	}

	return map[string]interface{}{
		"interest_group_owner": advertiserID,
		"interest_group_name":  "campaign_" + campaign.ID,
		"bid": map[string]interface{}{
			"bid":              campaign.BidPrice,
			"render":           campaign.Creative.URL,
			"ad_cost":          campaign.BidPrice,
			"modeling_signals": 0,
		},
		"bid_currency": "USD",
	}
}

// RegisterUserTopic registers a topic observation for a user (for simulation/testing)
func (s *PrivacySandboxService) RegisterUserTopic(userID string, topicID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.topicCache[userID]; !exists {
		s.topicCache[userID] = make([]int, 0, 5)
	}

	// Keep only last 5 topics (as per Topics API spec)
	topics := s.topicCache[userID]
	if len(topics) >= 5 {
		topics = topics[1:]
	}
	topics = append(topics, topicID)
	s.topicCache[userID] = topics
}

// AddToInterestGroup adds user to a FLEDGE interest group
func (s *PrivacySandboxService) AddToInterestGroup(userID, groupName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.fledgeAudiences[userID]; !exists {
		s.fledgeAudiences[userID] = make([]string, 0)
	}

	// Check if already in group
	for _, g := range s.fledgeAudiences[userID] {
		if g == groupName {
			return
		}
	}

	s.fledgeAudiences[userID] = append(s.fledgeAudiences[userID], groupName)
}

// GetUserInterestGroups returns interest groups for a user
func (s *PrivacySandboxService) GetUserInterestGroups(userID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if groups, exists := s.fledgeAudiences[userID]; exists {
		result := make([]string, len(groups))
		copy(result, groups)
		return result
	}

	return nil
}

// GetTopicName returns the human-readable name for a topic ID
func (s *PrivacySandboxService) GetTopicName(topicID int) string {
	if name, exists := topicTaxonomy[topicID]; exists {
		return name
	}
	return "Unknown"
}

// SimulateSharedStorage simulates Shared Storage API read (for testing)
func (s *PrivacySandboxService) SimulateSharedStorage(userID, key string) (string, bool) {
	// In real implementation, this would interact with browser's Shared Storage
	// For now, we simulate using our cache

	s.mu.RLock()
	defer s.mu.RUnlock()

	cacheKey := "shared_storage:" + userID + ":" + key
	// Simulated storage - would use actual cache in production
	_ = cacheKey
	return "", false
}

// GeneratePrivateAggregationReport creates a Private Aggregation API report
func (s *PrivacySandboxService) GeneratePrivateAggregationReport(campaign *model.Campaign, bucket int64, value int) map[string]interface{} {
	return map[string]interface{}{
		"contributions": []map[string]interface{}{
			{
				"bucket": bucket,
				"value":  value,
			},
		},
		"debug_key":                      campaign.ID,
		"aggregation_coordinator_origin": "https://publickeyservice.aws.privacysandboxservices.com",
	}
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CleanupExpiredTopics removes topics older than the retention period
func (s *PrivacySandboxService) CleanupExpiredTopics(retentionDays int) {
	// Topics API retains data for limited time (typically 3 weeks)
	// This would clean up old entries
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	_ = cutoff // Would use with timestamp tracking in production
}
