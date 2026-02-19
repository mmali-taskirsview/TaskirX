package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// UnifiedIDService manages cross-platform identity resolution
type UnifiedIDService struct {
	cache         cache.Cache
	mu            sync.RWMutex
	idGraph       map[string]*identityNode // Maps any ID to its node in the graph
	providerStats map[string]*providerStat
}

type identityNode struct {
	PrimaryID   string
	Provider    string
	LinkedIDs   []linkedIdentity
	Segments    []string
	Attributes  map[string]string
	DeviceTypes []string
	Confidence  float64
	LastUpdated time.Time
	ConsentGiven bool
}

type linkedIdentity struct {
	ID         string
	Provider   string
	DeviceType string
	Confidence float64
	LinkTime   time.Time
}

type providerStat struct {
	TotalLookups   int
	SuccessCount   int
	AvgLatencyMs   float64
	LastError      string
	LastErrorTime  time.Time
}

// NewUnifiedIDService creates a new unified ID service
func NewUnifiedIDService(c cache.Cache) *UnifiedIDService {
	return &UnifiedIDService{
		cache:         c,
		idGraph:       make(map[string]*identityNode),
		providerStats: make(map[string]*providerStat),
	}
}

// ResolveIdentity resolves user identity across multiple ID providers
func (s *UnifiedIDService) ResolveIdentity(campaign *model.Campaign, request *model.BidRequest) *model.UnifiedIDResult {
	config := campaign.Targeting.UnifiedIDConfig
	if config == nil || !config.Enabled {
		return &model.UnifiedIDResult{
			Resolved:      false,
			BidMultiplier: 1.0,
		}
	}

	result := &model.UnifiedIDResult{
		Resolved:      false,
		BidMultiplier: 1.0,
		HasConsent:    true, // Assume consent unless proven otherwise
	}

	// Check consent if required
	if config.ConsentRequired {
		hasConsent := s.checkConsent(request)
		result.HasConsent = hasConsent
		if !hasConsent {
			result.Reason = "consent_not_granted"
			return result
		}
	}

	// Attempt to resolve identity from request
	inputID := s.extractUserID(request)
	if inputID == "" {
		result.Reason = "no_user_id"
		return result
	}

	// Check cache/graph first
	s.mu.RLock()
	node, exists := s.idGraph[inputID]
	s.mu.RUnlock()

	if exists && time.Since(node.LastUpdated) < 24*time.Hour {
		return s.buildResultFromNode(node, config, result)
	}

	// Try providers in fallback order
	providers := s.getOrderedProviders(config)

	for _, provider := range providers {
		if !provider.Enabled {
			continue
		}

		resolvedID := s.resolveWithProvider(inputID, provider.Name)
		if resolvedID != nil {
			result.Resolved = true
			result.PrimaryID = resolvedID
			result.MatchedProviders = append(result.MatchedProviders, provider.Name)

			// Apply bid boost if configured
			if provider.BidBoost > 0 {
				result.BidMultiplier *= (1.0 + provider.BidBoost)
			}

			// If ID graph is enabled, get linked IDs
			if config.IDGraphEnabled {
				s.enrichWithLinkedIDs(result, inputID)
			}

			// Enrich profile if configured
			if config.EnrichProfiles {
				s.enrichProfile(result, resolvedID)
			}

			break // Got a match
		}
	}

	if !result.Resolved {
		result.Reason = "no_provider_match"
	}

	return result
}

func (s *UnifiedIDService) checkConsent(request *model.BidRequest) bool {
	// Check GDPR consent in user data
	// In production, check request.Regs.GDPR and request.User.Consent
	
	// For internal BidRequest, check context
	if request.Context != nil {
		if consent, ok := request.Context["gdpr_consent"].(bool); ok {
			return consent
		}
		if consentStr, ok := request.Context["consent_string"].(string); ok && consentStr != "" {
			return true // Has consent string
		}
	}

	return true // Default to true if no explicit denial
}

func (s *UnifiedIDService) extractUserID(request *model.BidRequest) string {
	// Try multiple sources for user ID
	if request.User.ID != "" {
		return request.User.ID
	}

	if request.Device.DeviceID != "" {
		return request.Device.DeviceID
	}

	// Check context for IDs
	if request.Context != nil {
		for _, key := range []string{"uid2", "id5", "rampid", "idfa", "gaid"} {
			if id, ok := request.Context[key].(string); ok && id != "" {
				return id
			}
		}
	}

	return ""
}

func (s *UnifiedIDService) getOrderedProviders(config *model.UnifiedIDConfig) []model.IDProvider {
	providers := make([]model.IDProvider, len(config.Providers))
	copy(providers, config.Providers)

	// If fallback order specified, reorder
	if len(config.FallbackOrder) > 0 {
		orderMap := make(map[string]int)
		for i, name := range config.FallbackOrder {
			orderMap[strings.ToLower(name)] = i
		}

		sort.Slice(providers, func(i, j int) bool {
			orderI, okI := orderMap[strings.ToLower(providers[i].Name)]
			orderJ, okJ := orderMap[strings.ToLower(providers[j].Name)]

			if !okI {
				orderI = 999
			}
			if !okJ {
				orderJ = 999
			}

			if orderI == orderJ {
				return providers[i].Priority < providers[j].Priority
			}
			return orderI < orderJ
		})
	} else {
		// Sort by priority
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].Priority < providers[j].Priority
		})
	}

	return providers
}

func (s *UnifiedIDService) resolveWithProvider(inputID, providerName string) *model.UnifiedID {
	// Simulate provider resolution
	// In production, this would call the actual provider APIs

	s.mu.Lock()
	if _, exists := s.providerStats[providerName]; !exists {
		s.providerStats[providerName] = &providerStat{}
	}
	s.providerStats[providerName].TotalLookups++
	s.mu.Unlock()

	// Generate deterministic provider ID based on input
	hash := sha256.Sum256([]byte(inputID + providerName))
	providerID := hex.EncodeToString(hash[:16])

	// Simulate match rate (in production, actual API call)
	matchRate := s.getProviderMatchRate(providerName)
	if !s.shouldMatch(inputID, matchRate) {
		return nil
	}

	s.mu.Lock()
	s.providerStats[providerName].SuccessCount++
	s.mu.Unlock()

	return &model.UnifiedID{
		Provider:      providerName,
		ID:            providerID,
		Confidence:    matchRate,
		ConsentStatus: "granted",
		LastRefreshed: time.Now(),
	}
}

func (s *UnifiedIDService) getProviderMatchRate(providerName string) float64 {
	// Default match rates by provider
	switch strings.ToLower(providerName) {
	case "uid2":
		return 0.75
	case "id5":
		return 0.70
	case "rampid", "liveramp":
		return 0.80
	case "zeotap":
		return 0.65
	default:
		return 0.50
	}
}

func (s *UnifiedIDService) shouldMatch(id string, matchRate float64) bool {
	// Deterministic "random" based on ID for consistent behavior
	hash := sha256.Sum256([]byte(id))
	val := float64(hash[0]) / 255.0
	return val < matchRate
}

func (s *UnifiedIDService) buildResultFromNode(node *identityNode, config *model.UnifiedIDConfig, result *model.UnifiedIDResult) *model.UnifiedIDResult {
	result.Resolved = true
	result.PrimaryID = &model.UnifiedID{
		Provider:      node.Provider,
		ID:            node.PrimaryID,
		Confidence:    node.Confidence,
		Segments:      node.Segments,
		Attributes:    node.Attributes,
		ConsentStatus: s.getConsentStatus(node.ConsentGiven),
		LastRefreshed: node.LastUpdated,
	}

	// Add linked IDs
	for _, linked := range node.LinkedIDs {
		altID := model.UnifiedID{
			Provider:   linked.Provider,
			ID:         linked.ID,
			Confidence: linked.Confidence,
			LinkedIDs: []model.LinkedID{
				{
					Provider:   linked.Provider,
					ID:         linked.ID,
					DeviceType: linked.DeviceType,
					Confidence: linked.Confidence,
				},
			},
		}
		result.AlternateIDs = append(result.AlternateIDs, altID)
	}

	result.DeviceCount = len(node.DeviceTypes)
	result.MatchedProviders = []string{node.Provider}
	result.AudienceSegments = node.Segments

	// Calculate bid multiplier based on identity strength
	result.BidMultiplier = 1.0 + (node.Confidence * 0.3) // Up to 30% boost for high confidence

	result.HasConsent = node.ConsentGiven

	return result
}

func (s *UnifiedIDService) getConsentStatus(hasConsent bool) string {
	if hasConsent {
		return "granted"
	}
	return "unknown"
}

func (s *UnifiedIDService) enrichWithLinkedIDs(result *model.UnifiedIDResult, inputID string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if node, exists := s.idGraph[inputID]; exists {
		for _, linked := range node.LinkedIDs {
			altID := model.UnifiedID{
				Provider:   linked.Provider,
				ID:         linked.ID,
				Confidence: linked.Confidence,
			}
			result.AlternateIDs = append(result.AlternateIDs, altID)
		}
		result.DeviceCount = len(node.DeviceTypes)
	}
}

func (s *UnifiedIDService) enrichProfile(result *model.UnifiedIDResult, resolvedID *model.UnifiedID) {
	// In production, this would fetch audience segments from DMP
	result.EnrichedProfile = true

	// Add mock segments for demo
	result.AudienceSegments = []string{
		"demo:auto_intenders",
		"demo:high_income",
		"demo:online_shoppers",
	}
}

// LinkIdentities links two identities in the graph
func (s *UnifiedIDService) LinkIdentities(id1, provider1, id2, provider2, deviceType string, confidence float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create or update node for id1
	if _, exists := s.idGraph[id1]; !exists {
		s.idGraph[id1] = &identityNode{
			PrimaryID:   id1,
			Provider:    provider1,
			LinkedIDs:   make([]linkedIdentity, 0),
			Attributes:  make(map[string]string),
			DeviceTypes: make([]string, 0),
			Confidence:  confidence,
			LastUpdated: time.Now(),
		}
	}

	node := s.idGraph[id1]

	// Add link
	linked := linkedIdentity{
		ID:         id2,
		Provider:   provider2,
		DeviceType: deviceType,
		Confidence: confidence,
		LinkTime:   time.Now(),
	}

	// Check if already linked
	for i, existing := range node.LinkedIDs {
		if existing.ID == id2 && existing.Provider == provider2 {
			// Update existing link
			node.LinkedIDs[i] = linked
			return
		}
	}

	node.LinkedIDs = append(node.LinkedIDs, linked)
	node.LastUpdated = time.Now()

	// Add device type if new
	if deviceType != "" {
		found := false
		for _, dt := range node.DeviceTypes {
			if dt == deviceType {
				found = true
				break
			}
		}
		if !found {
			node.DeviceTypes = append(node.DeviceTypes, deviceType)
		}
	}

	// Also create reverse link for graph traversal
	s.idGraph[id2] = &identityNode{
		PrimaryID:   id2,
		Provider:    provider2,
		LinkedIDs:   []linkedIdentity{{ID: id1, Provider: provider1, Confidence: confidence, LinkTime: time.Now()}},
		DeviceTypes: []string{deviceType},
		Confidence:  confidence,
		LastUpdated: time.Now(),
	}
}

// AddSegments adds audience segments to an identity
func (s *UnifiedIDService) AddSegments(id string, segments []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if node, exists := s.idGraph[id]; exists {
		segmentMap := make(map[string]bool)
		for _, seg := range node.Segments {
			segmentMap[seg] = true
		}
		for _, seg := range segments {
			if !segmentMap[seg] {
				node.Segments = append(node.Segments, seg)
			}
		}
		node.LastUpdated = time.Now()
	}
}

// SetConsent records user consent status
func (s *UnifiedIDService) SetConsent(id string, hasConsent bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if node, exists := s.idGraph[id]; exists {
		node.ConsentGiven = hasConsent
		node.LastUpdated = time.Now()
	}
}

// GetProviderStats returns statistics for ID providers
func (s *UnifiedIDService) GetProviderStats() map[string]*providerStat {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*providerStat)
	for k, v := range s.providerStats {
		stat := *v
		result[k] = &stat
	}
	return result
}

// GetIdentityReport generates a report on identity resolution
func (s *UnifiedIDService) GetIdentityReport() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report := make(map[string]interface{})

	report["total_identities"] = len(s.idGraph)

	// Count by provider
	providerCounts := make(map[string]int)
	totalLinks := 0
	for _, node := range s.idGraph {
		providerCounts[node.Provider]++
		totalLinks += len(node.LinkedIDs)
	}
	report["identities_by_provider"] = providerCounts
	report["total_links"] = totalLinks

	// Provider stats
	providerStats := make(map[string]map[string]interface{})
	for name, stat := range s.providerStats {
		providerStats[name] = map[string]interface{}{
			"total_lookups":  stat.TotalLookups,
			"success_count":  stat.SuccessCount,
			"match_rate":     float64(stat.SuccessCount) / float64(max(stat.TotalLookups, 1)),
			"avg_latency_ms": stat.AvgLatencyMs,
		}
	}
	report["provider_stats"] = providerStats

	return report
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// CleanupStaleIdentities removes identities not updated within retention period
func (s *UnifiedIDService) CleanupStaleIdentities(retentionDays int) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	removed := 0

	for id, node := range s.idGraph {
		if node.LastUpdated.Before(cutoff) {
			delete(s.idGraph, id)
			removed++
		}
	}

	return removed
}

// CalculateCrossDeviceReach estimates cross-device reach
func (s *UnifiedIDService) CalculateCrossDeviceReach() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.idGraph) == 0 {
		return 0
	}

	multiDeviceUsers := 0
	for _, node := range s.idGraph {
		if len(node.DeviceTypes) > 1 || len(node.LinkedIDs) > 0 {
			multiDeviceUsers++
		}
	}

	return float64(multiDeviceUsers) / float64(len(s.idGraph))
}

// DebugIdentity prints debug info for an identity (for troubleshooting)
func (s *UnifiedIDService) DebugIdentity(id string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	node, exists := s.idGraph[id]
	if !exists {
		return fmt.Sprintf("Identity %s not found in graph", id)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Identity: %s\n", id))
	sb.WriteString(fmt.Sprintf("  Provider: %s\n", node.Provider))
	sb.WriteString(fmt.Sprintf("  Confidence: %.2f\n", node.Confidence))
	sb.WriteString(fmt.Sprintf("  Consent: %v\n", node.ConsentGiven))
	sb.WriteString(fmt.Sprintf("  Device Types: %v\n", node.DeviceTypes))
	sb.WriteString(fmt.Sprintf("  Segments: %v\n", node.Segments))
	sb.WriteString(fmt.Sprintf("  Linked IDs: %d\n", len(node.LinkedIDs)))
	for i, link := range node.LinkedIDs {
		sb.WriteString(fmt.Sprintf("    [%d] %s (%s) confidence=%.2f\n", i, link.ID, link.Provider, link.Confidence))
	}
	sb.WriteString(fmt.Sprintf("  Last Updated: %s\n", node.LastUpdated.Format(time.RFC3339)))

	return sb.String()
}
