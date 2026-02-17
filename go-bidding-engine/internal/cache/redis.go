package cache

import (
	"context"
	"encoding/json"
	"fmt"
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
