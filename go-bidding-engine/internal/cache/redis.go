package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// RedisCache handles Redis operations
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(redisURL string, password string, db int) (*RedisCache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	if password != "" {
		opt.Password = password
	}
	opt.DB = db
	opt.PoolSize = 100
	opt.MaxRetries = 3

	client := redis.NewClient(opt)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
		ttl:    5 * time.Minute,
	}, nil
}

// GetActiveCampaigns retrieves all active campaigns from cache
func (r *RedisCache) GetActiveCampaigns() ([]*model.Campaign, error) {
	key := "campaigns:active"

	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return []*model.Campaign{}, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("Redis GET error: %w", err)
	}

	var campaigns []*model.Campaign
	if err := json.Unmarshal([]byte(val), &campaigns); err != nil {
		return nil, fmt.Errorf("failed to unmarshal campaigns: %w", err)
	}

	return campaigns, nil
}

// SetActiveCampaigns caches active campaigns
func (r *RedisCache) SetActiveCampaigns(campaigns []*model.Campaign) error {
	key := "campaigns:active"

	data, err := json.Marshal(campaigns)
	if err != nil {
		return fmt.Errorf("failed to marshal campaigns: %w", err)
	}

	if err := r.client.Set(r.ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("Redis SET error: %w", err)
	}

	return nil
}

// GetCampaign retrieves a specific campaign by ID
func (r *RedisCache) GetCampaign(campaignID string) (*model.Campaign, error) {
	key := fmt.Sprintf("campaign:%s", campaignID)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("Redis GET error: %w", err)
	}

	var campaign model.Campaign
	if err := json.Unmarshal([]byte(val), &campaign); err != nil {
		return nil, fmt.Errorf("failed to unmarshal campaign: %w", err)
	}

	return &campaign, nil
}

// SetCampaign caches a specific campaign
func (r *RedisCache) SetCampaign(campaign *model.Campaign) error {
	key := fmt.Sprintf("campaign:%s", campaign.ID)

	data, err := json.Marshal(campaign)
	if err != nil {
		return fmt.Errorf("failed to marshal campaign: %w", err)
	}

	if err := r.client.Set(r.ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("Redis SET error: %w", err)
	}

	return nil
}

// IncrementBidCount increments the bid counter for metrics
func (r *RedisCache) IncrementBidCount() error {
	key := "metrics:bids:total"
	return r.client.Incr(r.ctx, key).Err()
}

// IncrementWinCount increments the win counter
func (r *RedisCache) IncrementWinCount() error {
	key := "metrics:wins:total"
	return r.client.Incr(r.ctx, key).Err()
}

// GetBidCount gets total bid count
func (r *RedisCache) GetBidCount() (int64, error) {
	key := "metrics:bids:total"
	return r.client.Get(r.ctx, key).Int64()
}

// GetWinCount gets total win count
func (r *RedisCache) GetWinCount() (int64, error) {
	key := "metrics:wins:total"
	return r.client.Get(r.ctx, key).Int64()
}

// RecordLatency records bid processing latency
func (r *RedisCache) RecordLatency(latencyMs float64) error {
	key := "metrics:latency"
	timestamp := time.Now().Unix()
	score := float64(timestamp)

	return r.client.ZAdd(r.ctx, key, redis.Z{
		Score:  score,
		Member: latencyMs,
	}).Err()
}

// GetAverageLatency calculates average latency from last 1000 samples
func (r *RedisCache) GetAverageLatency() (float64, error) {
	key := "metrics:latency"

	vals, err := r.client.ZRevRange(r.ctx, key, 0, 999).Result()
	if err != nil {
		return 0, err
	}

	if len(vals) == 0 {
		return 0, nil
	}

	var sum float64
	var count int
	for _, val := range vals {
		var latency float64
		if _, err := fmt.Sscanf(val, "%f", &latency); err == nil {
			sum += latency
			count++
		}
	}

	if count == 0 {
		return 0, nil
	}

	return sum / float64(count), nil
}

// SetUserSegments caches user segments for a user
func (r *RedisCache) SetUserSegments(userID string, segments []string) error {
	key := fmt.Sprintf("user:%s:segments", userID)
	data, err := json.Marshal(segments)
	if err != nil {
		return fmt.Errorf("failed to marshal segments: %w", err)
	}
	return r.client.Set(r.ctx, key, data, r.ttl).Err()
}

// GetUserSegments retrieves cached user segments
func (r *RedisCache) GetUserSegments(userID string) ([]string, error) {
	key := fmt.Sprintf("user:%s:segments", userID)
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("Redis GET error: %w", err)
	}

	var segments []string
	if err := json.Unmarshal([]byte(val), &segments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal segments: %w", err)
	}
	return segments, nil
}

// SetGeoRules caches geo-targeting rules
func (r *RedisCache) SetGeoRules(countryCode string, rules map[string]interface{}) error {
	key := fmt.Sprintf("geo:%s:rules", countryCode)
	data, err := json.Marshal(rules)
	if err != nil {
		return fmt.Errorf("failed to marshal geo rules: %w", err)
	}
	return r.client.Set(r.ctx, key, data, r.ttl).Err()
}

// GetGeoRules retrieves cached geo rules
func (r *RedisCache) GetGeoRules(countryCode string) (map[string]interface{}, error) {
	key := fmt.Sprintf("geo:%s:rules", countryCode)
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("Redis GET error: %w", err)
	}

	var rules map[string]interface{}
	if err := json.Unmarshal([]byte(val), &rules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal geo rules: %w", err)
	}
	return rules, nil
}

// CampaignSpend represents daily spend tracking
type CampaignSpend struct {
	CampaignID string  `json:"campaign_id"`
	Spend      float64 `json:"spend"`
}

// IncrementCampaignSpend atomically increments the spend for a campaign for the current day
// Returns the new total spend for the day
func (r *RedisCache) IncrementCampaignSpend(campaignID string, amount float64) (float64, error) {
	// Key format: campaign:spend:<campaign_id>:<YYYY-MM-DD>
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("campaign:spend:%s:%s", campaignID, dateStr)

	// Redis INCRBYFLOAT is atomic
	newSpend, err := r.client.IncrByFloat(r.ctx, key, amount).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment campaign spend: %w", err)
	}

	// Set expiry for 48 hours (keep history for a bit)
	if err := r.client.Expire(r.ctx, key, 48*time.Hour).Err(); err != nil {
		// Log error but don't fail operation
		fmt.Printf("failed to set expiry for spend key: %v\n", err)
	}

	return newSpend, nil
}

// GetCampaignSpend retrieves the current day's spend for a campaign
func (r *RedisCache) GetCampaignSpend(campaignID string) (float64, error) {
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("campaign:spend:%s:%s", campaignID, dateStr)

	val, err := r.client.Get(r.ctx, key).Float64()
	if err == redis.Nil {
		return 0.0, nil
	}
	if err != nil {
		return 0.0, fmt.Errorf("failed to get campaign spend: %w", err)
	}
	return val, nil
}

// Get retrieves a value by key
func (r *RedisCache) Get(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	if err != nil {
		return "", fmt.Errorf("Redis GET error: %w", err)
	}
	return val, nil
}

// Set stores a key-value pair with TTL in seconds
func (r *RedisCache) Set(key string, value interface{}, ttl int64) error {
	expiration := time.Duration(ttl) * time.Second
	if err := r.client.Set(r.ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("Redis SET error: %w", err)
	}
	return nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Health checks Redis connection health
func (r *RedisCache) Health() error {
	return r.client.Ping(r.ctx).Err()
}

// IncrementPublisherFraud increments the fraud count for a publisher
func (r *RedisCache) IncrementPublisherFraud(publisherID string) error {
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("fraud:publisher:%s:%s:count", publisherID, dateStr)

	// Increment
	_, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to increment fraud count: %w", err)
	}

	// Add to set of flagged publishers for easy scanning
	setKey := fmt.Sprintf("fraud:publishers:active:%s", dateStr)
	r.client.SAdd(r.ctx, setKey, publisherID)

	return nil
}

// IsRequestDuplicate checks if a request ID has been seen before within TTL window.
// Returns true if duplicate (already exists), false if first time seeing it.
// Uses SetNX (SET if Not eXists) for atomic check-and-set operation.
func (r *RedisCache) IsRequestDuplicate(requestID string, ttlSeconds int) (bool, error) {
	key := fmt.Sprintf("dedup:req:%s", requestID)

	// SetNX returns true if key was set (first time seeing it), false if already exists
	wasSet, err := r.client.SetNX(r.ctx, key, "1", time.Duration(ttlSeconds)*time.Second).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check request deduplication: %w", err)
	}

	// If wasSet is true, this is NOT a duplicate (first time)
	// If wasSet is false, this IS a duplicate (already seen)
	isDuplicate := !wasSet

	return isDuplicate, nil
}

// IncrementBidFormat increments the bid count for a specific format
func (r *RedisCache) IncrementBidFormat(format string) error {
	key := fmt.Sprintf("stats:bids:format:%s", format)
	_, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to increment bid format %s: %w", format, err)
	}
	return nil
}

// GetBidFormats retrieves the bid counts for all formats
func (r *RedisCache) GetBidFormats() (map[string]int64, error) {
	formats := []string{"banner", "video", "native", "audio"}
	result := make(map[string]int64)

	for _, format := range formats {
		key := fmt.Sprintf("stats:bids:format:%s", format)
		val, err := r.client.Get(r.ctx, key).Int64()
		if err == redis.Nil {
			val = 0
		} else if err != nil {
			// Log error but continue
			fmt.Printf("Error getting format stats %s: %v\n", format, err)
			continue
		}
		result[format] = val
	}
	return result, nil
}

// StoreBidPathAnalytics stores SPO analytics for a bid request
func (r *RedisCache) StoreBidPathAnalytics(analytics *model.BidPathAnalytics) error {
	key := fmt.Sprintf("spo:analytics:%s", analytics.RequestID)

	jsonData, err := json.Marshal(analytics)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics: %w", err)
	}

	// Store with 24 hour TTL
	return r.client.Set(r.ctx, key, jsonData, 24*time.Hour).Err()
}

// GetBidPathAnalytics retrieves SPO analytics for a specific request
func (r *RedisCache) GetBidPathAnalytics(requestID string) (*model.BidPathAnalytics, error) {
	key := fmt.Sprintf("spo:analytics:%s", requestID)

	jsonData, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("analytics not found for request %s", requestID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	var analytics model.BidPathAnalytics
	if err := json.Unmarshal([]byte(jsonData), &analytics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analytics: %w", err)
	}

	return &analytics, nil
}

// GetSupplyChainMetrics aggregates SPO metrics across time ranges
func (r *RedisCache) GetSupplyChainMetrics(timeRange string) (*model.SupplyChainMetrics, error) {
	// For simplicity, we'll scan recent analytics and aggregate
	// In production, you'd want proper time-series storage

	pattern := "spo:analytics:*"
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to scan analytics keys: %w", err)
	}

	var totalRequests int64
	var successfulBids int64
	var totalLatency int64
	var totalFees float64
	serviceMetrics := make(map[string]*model.ServiceMetrics)

	for _, key := range keys {
		jsonData, err := r.client.Get(r.ctx, key).Result()
		if err != nil {
			continue // Skip if key expired or error
		}

		var analytics model.BidPathAnalytics
		if err := json.Unmarshal([]byte(jsonData), &analytics); err != nil {
			continue
		}

		totalRequests++
		if analytics.WonAuction {
			successfulBids++
		}
		totalLatency += analytics.TotalLatencyMs
		totalFees += analytics.TotalFees

		// Aggregate service metrics
		for _, hop := range analytics.Hops {
			if _, exists := serviceMetrics[hop.ServiceName]; !exists {
				serviceMetrics[hop.ServiceName] = &model.ServiceMetrics{
					ServiceName: hop.ServiceName,
				}
			}

			metrics := serviceMetrics[hop.ServiceName]
			metrics.TotalCalls++
			metrics.TotalFees += hop.Fee
			metrics.AvgLatencyMs += float64(hop.LatencyMs)

			if hop.Success {
				metrics.SuccessRate = (metrics.SuccessRate*float64(metrics.TotalCalls-1) + 1.0) / float64(metrics.TotalCalls)
			} else {
				metrics.SuccessRate = (metrics.SuccessRate*float64(metrics.TotalCalls-1) + 0.0) / float64(metrics.TotalCalls)
				metrics.ErrorRate++
			}
		}
	}

	// Convert service metrics map
	resultServiceMetrics := make(map[string]model.ServiceMetrics)
	for name, metrics := range serviceMetrics {
		if metrics.TotalCalls > 0 {
			metrics.AvgLatencyMs /= float64(metrics.TotalCalls)
		}
		resultServiceMetrics[name] = *metrics
	}

	// Calculate rates with division by zero protection
	var winRate, avgLatency, avgFees, pathEfficiency float64
	if totalRequests > 0 {
		winRate = float64(successfulBids) / float64(totalRequests)
		avgLatency = float64(totalLatency) / float64(totalRequests)
		avgFees = totalFees / float64(totalRequests)
		pathEfficiency = 1.0 - (float64(totalLatency) / float64(totalRequests) / 1000.0)
	} else {
		winRate = 0.0
		avgLatency = 0.0
		avgFees = 0.0
		pathEfficiency = 1.0
	}

	metrics := &model.SupplyChainMetrics{
		TimeRange:      timeRange,
		TotalRequests:  totalRequests,
		SuccessfulBids: successfulBids,
		WinRate:        winRate,
		AvgLatencyMs:   avgLatency,
		AvgTotalFees:   avgFees,
		ServiceMetrics: resultServiceMetrics,
		PathEfficiency: pathEfficiency,
		Timestamp:      time.Now(),
	}

	return metrics, nil
}

// GetServiceMetrics returns metrics for a specific service
func (r *RedisCache) GetServiceMetrics(serviceName string, timeRange string) (*model.ServiceMetrics, error) {
	// Get overall metrics and extract specific service
	overall, err := r.GetSupplyChainMetrics(timeRange)
	if err != nil {
		return nil, err
	}

	if metrics, exists := overall.ServiceMetrics[serviceName]; exists {
		return &metrics, nil
	}

	return nil, fmt.Errorf("service %s not found in metrics", serviceName)
}

// IncrementUserFrequency atomically increments the impression count for a user+campaign pair
// and sets a TTL on first write to enforce the frequency window.
func (r *RedisCache) IncrementUserFrequency(userID, campaignID string, windowSecs int) (int64, error) {
	key := fmt.Sprintf("freq:%s:%s", userID, campaignID)
	pipe := r.client.Pipeline()
	incr := pipe.Incr(r.ctx, key)
	// Only set expiry on first increment (NX = only if not exists)
	pipe.ExpireNX(r.ctx, key, time.Duration(windowSecs)*time.Second)
	_, err := pipe.Exec(r.ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// GetUserFrequency returns the current impression count for a user+campaign pair
func (r *RedisCache) GetUserFrequency(userID, campaignID string) (int64, error) {
	key := fmt.Sprintf("freq:%s:%s", userID, campaignID)
	val, err := r.client.Get(r.ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// IncrementCampaignClicks increments the click counter for a campaign (used for CTR calculation)
func (r *RedisCache) IncrementCampaignClicks(campaignID string) error {
	key := fmt.Sprintf("campaign:clicks:%s:%s", campaignID, time.Now().Format("2006-01-02"))
	return r.client.Incr(r.ctx, key).Err()
}

// IncrementCampaignImpressions increments the impression counter for a campaign
func (r *RedisCache) IncrementCampaignImpressions(campaignID string) error {
	key := fmt.Sprintf("campaign:imps:%s:%s", campaignID, time.Now().Format("2006-01-02"))
	return r.client.Incr(r.ctx, key).Err()
}

// GetCampaignCTR returns the click-through rate for a campaign based on today's data
func (r *RedisCache) GetCampaignCTR(campaignID string) (float64, error) {
	dateStr := time.Now().Format("2006-01-02")
	clicksKey := fmt.Sprintf("campaign:clicks:%s:%s", campaignID, dateStr)
	impsKey := fmt.Sprintf("campaign:imps:%s:%s", campaignID, dateStr)

	clicks, err := r.client.Get(r.ctx, clicksKey).Int64()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	imps, err := r.client.Get(r.ctx, impsKey).Int64()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	if imps == 0 {
		return 0, nil
	}
	return float64(clicks) / float64(imps), nil
}

// GetCampaignWinRate returns the win rate based on bids vs wins stored per campaign
func (r *RedisCache) GetCampaignWinRate(campaignID string) (float64, error) {
	dateStr := time.Now().Format("2006-01-02")
	bidsKey := fmt.Sprintf("campaign:bids:%s:%s", campaignID, dateStr)
	winsKey := fmt.Sprintf("campaign:wins:%s:%s", campaignID, dateStr)

	bids, err := r.client.Get(r.ctx, bidsKey).Int64()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	wins, err := r.client.Get(r.ctx, winsKey).Int64()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	if bids == 0 {
		return 0.5, nil // Default 50% win rate when no data
	}
	return float64(wins) / float64(bids), nil
}

// IncrementCampaignBids increments the daily bid-participation counter for a campaign.
// This is the denominator in win-rate = wins / bids.
func (r *RedisCache) IncrementCampaignBids(campaignID string) error {
	key := fmt.Sprintf("campaign:bids:%s:%s", campaignID, time.Now().Format("2006-01-02"))
	return r.client.Incr(r.ctx, key).Err()
}

// IncrementCampaignWins increments the daily win counter for a campaign.
// This is the numerator in win-rate = wins / bids.
func (r *RedisCache) IncrementCampaignWins(campaignID string) error {
	key := fmt.Sprintf("campaign:wins:%s:%s", campaignID, time.Now().Format("2006-01-02"))
	return r.client.Incr(r.ctx, key).Err()
}

// RecordBidInBucket tracks a bid in the appropriate price bucket
func (r *RedisCache) RecordBidInBucket(priceBucket string) error {
	key := fmt.Sprintf("landscape:bids:%s:%s", priceBucket, time.Now().Format("2006-01-02"))
	return r.client.Incr(r.ctx, key).Err()
}

// RecordWinInBucket tracks a win in the appropriate price bucket
func (r *RedisCache) RecordWinInBucket(priceBucket string) error {
	key := fmt.Sprintf("landscape:wins:%s:%s", priceBucket, time.Now().Format("2006-01-02"))
	return r.client.Incr(r.ctx, key).Err()
}

// GetBidLandscape returns bid and win counts across all price buckets for today
func (r *RedisCache) GetBidLandscape() (map[string]map[string]int64, error) {
	dateStr := time.Now().Format("2006-01-02")
	pattern := fmt.Sprintf("landscape:*:%s", dateStr)

	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	landscape := make(map[string]map[string]int64)

	for _, key := range keys {
		// Key format: landscape:{bids|wins}:{bucket}:{date}
		parts := strings.Split(key, ":")
		if len(parts) != 4 {
			continue
		}
		metricType := parts[1] // "bids" or "wins"
		bucket := parts[2]

		val, err := r.client.Get(r.ctx, key).Int64()
		if err != nil {
			continue
		}

		if _, exists := landscape[bucket]; !exists {
			landscape[bucket] = make(map[string]int64)
		}
		landscape[bucket][metricType] = val
	}

	return landscape, nil
}

// IncrementSegmentImpressions tracks impressions by segment (device/os/geo)
func (r *RedisCache) IncrementSegmentImpressions(segmentType, segmentValue string) error {
	key := fmt.Sprintf("segment:%s:imps:%s:%s", segmentType, segmentValue, time.Now().Format("2006-01-02"))
	return r.client.Incr(r.ctx, key).Err()
}

// IncrementSegmentClicks tracks clicks by segment (device/os/geo)
func (r *RedisCache) IncrementSegmentClicks(segmentType, segmentValue string) error {
	key := fmt.Sprintf("segment:%s:clicks:%s:%s", segmentType, segmentValue, time.Now().Format("2006-01-02"))
	return r.client.Incr(r.ctx, key).Err()
}

// GetSegmentPerformance returns impression and click counts for all segments of a given type
func (r *RedisCache) GetSegmentPerformance(segmentType string) (map[string]map[string]int64, error) {
	dateStr := time.Now().Format("2006-01-02")
	pattern := fmt.Sprintf("segment:%s:*:%s", segmentType, dateStr)

	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	performance := make(map[string]map[string]int64)

	for _, key := range keys {
		// Key format: segment:{type}:{imps|clicks}:{value}:{date}
		parts := strings.Split(key, ":")
		if len(parts) != 5 {
			continue
		}
		metricType := parts[2] // "imps" or "clicks"
		segmentValue := parts[3]

		val, err := r.client.Get(r.ctx, key).Int64()
		if err != nil {
			continue
		}

		if _, exists := performance[segmentValue]; !exists {
			performance[segmentValue] = make(map[string]int64)
		}
		performance[segmentValue][metricType] = val
	}

	return performance, nil
}

// RecordPublisherBidAttempt tracks bid attempts per publisher in price buckets
func (r *RedisCache) RecordPublisherBidAttempt(publisherID string, bidPrice float64, won bool) error {
	dateStr := time.Now().Format("2006-01-02")

	// Determine price bucket (same as bid landscape buckets)
	bucket := getPriceBucketForFloor(bidPrice)

	// Increment total bids in this bucket for this publisher
	bidsKey := fmt.Sprintf("floor:pub:%s:bids:%s:%s", publisherID, bucket, dateStr)
	if err := r.client.Incr(r.ctx, bidsKey).Err(); err != nil {
		return err
	}

	// Increment wins if auction was won
	if won {
		winsKey := fmt.Sprintf("floor:pub:%s:wins:%s:%s", publisherID, bucket, dateStr)
		if err := r.client.Incr(r.ctx, winsKey).Err(); err != nil {
			return err
		}
	}

	return nil
}

// GetOptimalBidFloor calculates the optimal bid floor for a publisher based on historical win rates
// Algorithm: Find the highest price bucket where win rate >= targetWinRate
func (r *RedisCache) GetOptimalBidFloor(publisherID string, targetWinRate float64) (float64, error) {
	dateStr := time.Now().Format("2006-01-02")
	pattern := fmt.Sprintf("floor:pub:%s:*:%s", publisherID, dateStr)

	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return 0, err
	}

	if len(keys) == 0 {
		return 0.5, nil // Default floor if no data
	}

	// Collect win rates per bucket
	bucketStats := make(map[string]struct {
		bids int64
		wins int64
	})

	for _, key := range keys {
		// Key format: floor:pub:{id}:{bids|wins}:{bucket}:{date}
		parts := strings.Split(key, ":")
		if len(parts) != 6 {
			continue
		}
		metricType := parts[3] // "bids" or "wins"
		bucket := parts[4]

		val, err := r.client.Get(r.ctx, key).Int64()
		if err != nil {
			continue
		}

		stats := bucketStats[bucket]
		if metricType == "bids" {
			stats.bids = val
		} else if metricType == "wins" {
			stats.wins = val
		}
		bucketStats[bucket] = stats
	}

	// Find highest bucket where win rate >= target
	buckets := []string{"0.00-0.50", "0.50-1.00", "1.00-2.00", "2.00-5.00", "5.00-10.00", "10.00+"}
	bucketFloors := map[string]float64{
		"0.00-0.50":  0.25,
		"0.50-1.00":  0.75,
		"1.00-2.00":  1.50,
		"2.00-5.00":  3.50,
		"5.00-10.00": 7.50,
		"10.00+":     10.00,
	}

	optimalFloor := 0.5 // Default

	for i := len(buckets) - 1; i >= 0; i-- {
		bucket := buckets[i]
		stats := bucketStats[bucket]

		if stats.bids < 10 {
			continue // Need at least 10 bids for statistical significance
		}

		winRate := float64(stats.wins) / float64(stats.bids)
		if winRate >= targetWinRate {
			optimalFloor = bucketFloors[bucket]
			break
		}
	}

	return optimalFloor, nil
}

// Helper function for price bucketing (same logic as in bidding service)
func getPriceBucketForFloor(price float64) string {
	switch {
	case price < 0.50:
		return "0.00-0.50"
	case price < 1.00:
		return "0.50-1.00"
	case price < 2.00:
		return "1.00-2.00"
	case price < 5.00:
		return "2.00-5.00"
	case price < 10.00:
		return "5.00-10.00"
	default:
		return "10.00+"
	}
}

// RecordImpression stores an impression event for view-through attribution (VTA)
func (r *RedisCache) RecordImpression(userID, campaignID, requestID string, ttlHours int) error {
	key := fmt.Sprintf("attr:imp:%s:%s", userID, campaignID)
	value := fmt.Sprintf("%s:%d", requestID, time.Now().Unix())
	return r.client.Set(r.ctx, key, value, time.Duration(ttlHours)*time.Hour).Err()
}

// RecordClick stores a click event for click-through attribution (CTA)
// Clicks take priority over impressions for attribution
func (r *RedisCache) RecordClick(userID, campaignID, requestID string, ttlHours int) error {
	key := fmt.Sprintf("attr:click:%s:%s", userID, campaignID)
	value := fmt.Sprintf("%s:%d", requestID, time.Now().Unix())
	return r.client.Set(r.ctx, key, value, time.Duration(ttlHours)*time.Hour).Err()
}

// GetAttribution checks for click-through (CTA) or view-through (VTA) attribution
// Returns attributionType ("cta", "vta", or "none") and the requestID
func (r *RedisCache) GetAttribution(userID, campaignID string) (string, string, error) {
	// Check for click first (CTA takes priority)
	clickKey := fmt.Sprintf("attr:click:%s:%s", userID, campaignID)
	clickVal, err := r.client.Get(r.ctx, clickKey).Result()
	if err == nil && clickVal != "" {
		parts := strings.Split(clickVal, ":")
		if len(parts) >= 1 {
			return "cta", parts[0], nil
		}
	}

	// Check for impression (VTA)
	impKey := fmt.Sprintf("attr:imp:%s:%s", userID, campaignID)
	impVal, err := r.client.Get(r.ctx, impKey).Result()
	if err == nil && impVal != "" {
		parts := strings.Split(impVal, ":")
		if len(parts) >= 1 {
			return "vta", parts[0], nil
		}
	}

	return "none", "", nil
}

// RecordUserEvent stores a user engagement event for retargeting
// Key format: retarget:{eventType}:{userID}:{campaignID}
// This allows querying all campaigns a user has engaged with by event type
func (r *RedisCache) RecordUserEvent(userID, campaignID, eventType string, ttlDays int) error {
	if userID == "" || campaignID == "" || eventType == "" {
		return fmt.Errorf("userID, campaignID, and eventType are required")
	}

	// Store event with TTL
	key := fmt.Sprintf("retarget:%s:%s:%s", eventType, userID, campaignID)
	ttl := time.Duration(ttlDays) * 24 * time.Hour
	if ttlDays <= 0 {
		ttl = 30 * 24 * time.Hour // Default 30 day window
	}

	return r.client.Set(r.ctx, key, time.Now().Unix(), ttl).Err()
}

// GetUserEvents returns all campaigns a user has engaged with by event type
// Returns map[eventType][]campaignIDs
func (r *RedisCache) GetUserEvents(userID string, eventTypes []string) (map[string][]string, error) {
	result := make(map[string][]string)

	for _, eventType := range eventTypes {
		pattern := fmt.Sprintf("retarget:%s:%s:*", eventType, userID)
		keys, err := r.client.Keys(r.ctx, pattern).Result()
		if err != nil {
			continue
		}

		campaignIDs := make([]string, 0, len(keys))
		for _, key := range keys {
			// Key format: retarget:{eventType}:{userID}:{campaignID}
			parts := strings.Split(key, ":")
			if len(parts) >= 4 {
				campaignIDs = append(campaignIDs, parts[3])
			}
		}
		result[eventType] = campaignIDs
	}

	return result, nil
}

// HasUserEvent checks if a user has a specific engagement event for a campaign
func (r *RedisCache) HasUserEvent(userID, campaignID, eventType string) (bool, error) {
	key := fmt.Sprintf("retarget:%s:%s:%s", eventType, userID, campaignID)
	exists, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// RecordTouchpoint stores a touchpoint in a user's conversion journey
// Uses a sorted set with timestamp as score for chronological ordering
// Key format: mta:journey:{userID}:{campaignID}
func (r *RedisCache) RecordTouchpoint(userID, campaignID, touchpointType, requestID string, ttlDays int) error {
	if userID == "" || campaignID == "" || touchpointType == "" {
		return fmt.Errorf("userID, campaignID, and touchpointType are required")
	}

	key := fmt.Sprintf("mta:journey:%s:%s", userID, campaignID)
	timestamp := time.Now().Unix()

	// Store as "type:requestID:timestamp"
	member := fmt.Sprintf("%s:%s:%d", touchpointType, requestID, timestamp)

	// Add to sorted set with timestamp as score
	err := r.client.ZAdd(r.ctx, key, redis.Z{
		Score:  float64(timestamp),
		Member: member,
	}).Err()
	if err != nil {
		return err
	}

	// Set TTL
	ttl := time.Duration(ttlDays) * 24 * time.Hour
	if ttlDays <= 0 {
		ttl = 30 * 24 * time.Hour // Default 30 day window
	}
	return r.client.Expire(r.ctx, key, ttl).Err()
}

// GetTouchpoints retrieves all touchpoints for a user/campaign journey
func (r *RedisCache) GetTouchpoints(userID, campaignID string) ([]model.Touchpoint, error) {
	key := fmt.Sprintf("mta:journey:%s:%s", userID, campaignID)

	// Get all members ordered by timestamp
	members, err := r.client.ZRangeWithScores(r.ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	touchpoints := make([]model.Touchpoint, 0, len(members))
	for i, member := range members {
		// Parse "type:requestID:timestamp"
		parts := strings.SplitN(member.Member.(string), ":", 3)
		if len(parts) < 2 {
			continue
		}

		tp := model.Touchpoint{
			Type:       parts[0],
			RequestID:  parts[1],
			CampaignID: campaignID,
			Timestamp:  time.Unix(int64(member.Score), 0),
			Position:   i + 1, // 1-indexed position
		}
		touchpoints = append(touchpoints, tp)
	}

	// Mark last position
	if len(touchpoints) > 0 {
		touchpoints[len(touchpoints)-1].Position = -1
	}

	return touchpoints, nil
}

// GetMultiTouchAttribution calculates attribution credit using the specified model
// Supported models: "linear", "time_decay", "position_based", "last_touch", "first_touch"
func (r *RedisCache) GetMultiTouchAttribution(userID, campaignID, modelType string) ([]model.AttributionCredit, error) {
	touchpoints, err := r.GetTouchpoints(userID, campaignID)
	if err != nil {
		return nil, err
	}

	if len(touchpoints) == 0 {
		return []model.AttributionCredit{}, nil
	}

	credits := make([]model.AttributionCredit, len(touchpoints))

	switch modelType {
	case "first_touch":
		// 100% credit to first touchpoint
		for i, tp := range touchpoints {
			credit := 0.0
			if i == 0 {
				credit = 1.0
			}
			credits[i] = model.AttributionCredit{
				Touchpoint: tp,
				Credit:     credit,
				Model:      modelType,
			}
		}

	case "last_touch":
		// 100% credit to last touchpoint
		for i, tp := range touchpoints {
			credit := 0.0
			if i == len(touchpoints)-1 {
				credit = 1.0
			}
			credits[i] = model.AttributionCredit{
				Touchpoint: tp,
				Credit:     credit,
				Model:      modelType,
			}
		}

	case "linear":
		// Equal credit to all touchpoints
		equalCredit := 1.0 / float64(len(touchpoints))
		for i, tp := range touchpoints {
			credits[i] = model.AttributionCredit{
				Touchpoint: tp,
				Credit:     equalCredit,
				Model:      modelType,
			}
		}

	case "time_decay":
		// More credit to recent touchpoints (exponential decay)
		// Half-life of 7 days
		halfLife := 7.0 * 24.0 * 3600.0 // 7 days in seconds
		now := time.Now()
		totalWeight := 0.0
		weights := make([]float64, len(touchpoints))

		for i, tp := range touchpoints {
			age := now.Sub(tp.Timestamp).Seconds()
			weight := math.Pow(0.5, age/halfLife)
			weights[i] = weight
			totalWeight += weight
		}

		for i, tp := range touchpoints {
			credits[i] = model.AttributionCredit{
				Touchpoint: tp,
				Credit:     weights[i] / totalWeight,
				Model:      modelType,
			}
		}

	case "position_based":
		// 40% first, 40% last, 20% distributed among middle
		n := len(touchpoints)
		for i, tp := range touchpoints {
			var credit float64
			if n == 1 {
				credit = 1.0
			} else if n == 2 {
				credit = 0.5
			} else {
				if i == 0 {
					credit = 0.4
				} else if i == n-1 {
					credit = 0.4
				} else {
					credit = 0.2 / float64(n-2)
				}
			}
			credits[i] = model.AttributionCredit{
				Touchpoint: tp,
				Credit:     credit,
				Model:      modelType,
			}
		}

	default:
		// Default to last touch
		return r.GetMultiTouchAttribution(userID, campaignID, "last_touch")
	}

	return credits, nil
}

// LinkDevices links multiple device IDs to a primary user ID for cross-device targeting
// Key format: xdev:primary:{deviceID} -> primaryUserID
// Key format: xdev:devices:{primaryUserID} -> set of deviceIDs
func (r *RedisCache) LinkDevices(primaryUserID string, deviceIDs []string, ttlDays int) error {
	if primaryUserID == "" || len(deviceIDs) == 0 {
		return fmt.Errorf("primaryUserID and deviceIDs are required")
	}

	ttl := time.Duration(ttlDays) * 24 * time.Hour
	if ttlDays <= 0 {
		ttl = 90 * 24 * time.Hour // Default 90 day device graph TTL
	}

	pipe := r.client.Pipeline()

	// Store primary user ID for each device
	for _, deviceID := range deviceIDs {
		primaryKey := fmt.Sprintf("xdev:primary:%s", deviceID)
		pipe.Set(r.ctx, primaryKey, primaryUserID, ttl)
	}

	// Store all devices under primary user
	devicesKey := fmt.Sprintf("xdev:devices:%s", primaryUserID)
	for _, deviceID := range deviceIDs {
		pipe.SAdd(r.ctx, devicesKey, deviceID)
	}
	pipe.Expire(r.ctx, devicesKey, ttl)

	_, err := pipe.Exec(r.ctx)
	return err
}

// GetLinkedDevices returns all devices linked to the same user as this device
func (r *RedisCache) GetLinkedDevices(deviceID string) ([]string, error) {
	// First get the primary user ID
	primaryUserID, err := r.GetPrimaryUserID(deviceID)
	if err != nil || primaryUserID == "" {
		return []string{deviceID}, nil // Return just this device if no graph
	}

	// Get all devices for this primary user
	devicesKey := fmt.Sprintf("xdev:devices:%s", primaryUserID)
	devices, err := r.client.SMembers(r.ctx, devicesKey).Result()
	if err != nil {
		return []string{deviceID}, nil
	}

	if len(devices) == 0 {
		return []string{deviceID}, nil
	}

	return devices, nil
}

// GetPrimaryUserID resolves a device ID to its primary user ID
func (r *RedisCache) GetPrimaryUserID(deviceID string) (string, error) {
	primaryKey := fmt.Sprintf("xdev:primary:%s", deviceID)
	primaryUserID, err := r.client.Get(r.ctx, primaryKey).Result()
	if err == redis.Nil {
		return "", nil // No primary user ID found
	}
	if err != nil {
		return "", err
	}
	return primaryUserID, nil
}

// GetCrossDeviceFrequency returns total impression frequency across all linked devices
func (r *RedisCache) GetCrossDeviceFrequency(primaryUserID, campaignID string) (int64, error) {
	// Get all devices for this primary user
	devicesKey := fmt.Sprintf("xdev:devices:%s", primaryUserID)
	devices, err := r.client.SMembers(r.ctx, devicesKey).Result()
	if err != nil {
		return 0, err
	}

	if len(devices) == 0 {
		// No device graph, get frequency for primary user ID directly
		return r.GetUserFrequency(primaryUserID, campaignID)
	}

	// Sum frequency across all devices
	var totalFreq int64
	for _, deviceID := range devices {
		freq, err := r.GetUserFrequency(deviceID, campaignID)
		if err == nil {
			totalFreq += freq
		}
	}

	return totalFreq, nil
}
