package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// newTestRedisCache creates a RedisCache backed by a redismock client
func newTestRedisCache(t *testing.T) (*RedisCache, redismock.ClientMock) {
	t.Helper()
	client, mock := redismock.NewClientMock()
	rc := &RedisCache{
		client: client,
		ctx:    context.Background(),
		ttl:    5 * time.Minute,
	}
	return rc, mock
}

// keep errors import alive
var _ = errors.New

// =============================================================================
// GetActiveCampaigns / SetActiveCampaigns
// =============================================================================

func TestRedis_GetActiveCampaigns_Hit(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	campaigns := []*model.Campaign{{ID: "c1", Name: "Test"}}
	data, _ := json.Marshal(campaigns)
	mock.ExpectGet("campaigns:active").SetVal(string(data))
	result, err := rc.GetActiveCampaigns()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].ID != "c1" {
		t.Errorf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRedis_GetActiveCampaigns_Miss(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("campaigns:active").RedisNil()
	result, err := rc.GetActiveCampaigns()
	if err != nil {
		t.Fatalf("unexpected error on miss: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetActiveCampaigns_Error(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("campaigns:active").SetErr(errors.New("redis error"))
	_, err := rc.GetActiveCampaigns()
	if err == nil {
		t.Error("expected error, got nil")
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_SetActiveCampaigns(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	campaigns := []*model.Campaign{{ID: "c1"}}
	data, _ := json.Marshal(campaigns)
	mock.ExpectSet("campaigns:active", data, 5*time.Minute).SetVal("OK")
	if err := rc.SetActiveCampaigns(campaigns); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// GetCampaign / SetCampaign
// =============================================================================

func TestRedis_GetCampaign_Hit(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	c := &model.Campaign{ID: "camp-42", Name: "Foo"}
	data, _ := json.Marshal(c)
	mock.ExpectGet("campaign:camp-42").SetVal(string(data))
	result, err := rc.GetCampaign("camp-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.ID != "camp-42" {
		t.Errorf("unexpected result: %+v", result)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetCampaign_Miss(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("campaign:nope").RedisNil()
	result, err := rc.GetCampaign("nope")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil on miss")
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_SetCampaign(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	c := &model.Campaign{ID: "camp-99"}
	data, _ := json.Marshal(c)
	mock.ExpectSet("campaign:camp-99", data, 5*time.Minute).SetVal("OK")
	if err := rc.SetCampaign(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Bid Counts  (no-arg versions)
// =============================================================================

func TestRedis_IncrementBidCount(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectIncr("metrics:bids:total").SetVal(1)
	if err := rc.IncrementBidCount(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRedis_IncrementWinCount(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectIncr("metrics:wins:total").SetVal(1)
	if err := rc.IncrementWinCount(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRedis_GetBidCount(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("metrics:bids:total").SetVal("10")
	count, err := rc.GetBidCount()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 10 {
		t.Errorf("expected 10, got %d", count)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetBidCount_Miss(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("metrics:bids:total").RedisNil()
	_, err := rc.GetBidCount()
	// redis.Nil is returned as error from Int64()
	if err == nil {
		t.Error("expected error on miss")
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetWinCount(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("metrics:wins:total").SetVal("7")
	count, err := rc.GetWinCount()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 7 {
		t.Errorf("expected 7, got %d", count)
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Latency  (float64-only arg)
// =============================================================================

func TestRedis_RecordLatency(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	// Score is a dynamic timestamp; use CustomMatch to accept any ZAdd to the latency key
	mock.CustomMatch(func(expected, actual []interface{}) error {
		return nil // accept any args
	}).ExpectZAdd("metrics:latency", redis.Z{Score: 0, Member: ""}).SetVal(1)
	if err := rc.RecordLatency(12.5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRedis_GetAverageLatency_WithValues(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectZRevRange("metrics:latency", 0, 999).SetVal([]string{"10", "20", "30"})
	avg, err := rc.GetAverageLatency()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avg != 20.0 {
		t.Errorf("expected 20.0, got %f", avg)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRedis_GetAverageLatency_Empty(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectZRevRange("metrics:latency", 0, 999).SetVal([]string{})
	avg, err := rc.GetAverageLatency()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if avg != 0 {
		t.Errorf("expected 0 on empty, got %f", avg)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// =============================================================================
// User Segments
// =============================================================================

func TestRedis_SetUserSegments(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	segs := []string{"seg-a", "seg-b"}
	data, _ := json.Marshal(segs)
	mock.ExpectSet("user:user-1:segments", data, 5*time.Minute).SetVal("OK")
	if err := rc.SetUserSegments("user-1", segs); err != nil {
		t.Fatalf("set error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetUserSegments_Hit(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	segs := []string{"seg-a", "seg-b"}
	data, _ := json.Marshal(segs)
	mock.ExpectGet("user:user-1:segments").SetVal(string(data))
	result, err := rc.GetUserSegments("user-1")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 segments, got %d", len(result))
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetUserSegments_Miss(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("user:user-1:segments").RedisNil()
	result, err := rc.GetUserSegments("user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil on miss")
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Geo Rules  (countryCode string, rules map[string]interface{})
// =============================================================================

func TestRedis_SetGeoRules(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	rules := map[string]interface{}{"allow": true, "max_bid": 5.0}
	data, _ := json.Marshal(rules)
	mock.ExpectSet("geo:US:rules", data, 5*time.Minute).SetVal("OK")
	if err := rc.SetGeoRules("US", rules); err != nil {
		t.Fatalf("set error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetGeoRules_Hit(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	rules := map[string]interface{}{"allow": true}
	data, _ := json.Marshal(rules)
	mock.ExpectGet("geo:US:rules").SetVal(string(data))
	result, err := rc.GetGeoRules("US")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetGeoRules_Miss(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("geo:US:rules").RedisNil()
	result, err := rc.GetGeoRules("US")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil on miss")
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Campaign Spend  (returns float64, error)
// =============================================================================

func TestRedis_IncrementCampaignSpend(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("campaign:spend:camp-1:%s", dateStr)
	mock.ExpectIncrByFloat(key, 5.0).SetVal(105.0)
	mock.ExpectExpire(key, 48*time.Hour).SetVal(true)
	newSpend, err := rc.IncrementCampaignSpend("camp-1", 5.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newSpend != 105.0 {
		t.Errorf("expected 105.0, got %f", newSpend)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetCampaignSpend(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("campaign:spend:camp-1:%s", dateStr)
	mock.ExpectGet(key).SetVal("99.99")
	spend, err := rc.GetCampaignSpend("camp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spend != 99.99 {
		t.Errorf("expected 99.99, got %f", spend)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetCampaignSpend_Miss(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("campaign:spend:camp-x:%s", dateStr)
	mock.ExpectGet(key).RedisNil()
	spend, err := rc.GetCampaignSpend("camp-x")
	if err != nil {
		t.Fatalf("unexpected error on miss: %v", err)
	}
	if spend != 0 {
		t.Errorf("expected 0 on miss, got %f", spend)
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Generic Get / Set  (Set takes key, value, ttl int64)
// =============================================================================

func TestRedis_Get_Hit(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("some:key").SetVal("hello")
	val, err := rc.Get("some:key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hello" {
		t.Errorf("expected 'hello', got '%s'", val)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_Get_Miss(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("miss:key").RedisNil()
	val, err := rc.Get("miss:key")
	if err != nil {
		t.Fatalf("unexpected error on miss: %v", err)
	}
	if val != "" {
		t.Errorf("expected empty string, got '%s'", val)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_Get_Error(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("err:key").SetErr(errors.New("conn refused"))
	_, err := rc.Get("err:key")
	if err == nil {
		t.Error("expected error, got nil")
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_Set(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectSet("my:key", "myval", 60*time.Second).SetVal("OK")
	if err := rc.Set("my:key", "myval", 60); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Frequency Capping  (IncrementUserFrequency uses pipeline, GetUserFrequency(userID, campaignID))
// =============================================================================

func TestRedis_GetUserFrequency_Hit(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("freq:user-1:camp-1").SetVal("5")
	count, err := rc.GetUserFrequency("user-1", "camp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetUserFrequency_Miss(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("freq:user-1:camp-1").RedisNil()
	count, err := rc.GetUserFrequency("user-1", "camp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 on miss, got %d", count)
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Request Deduplication  (IsRequestDuplicate(requestID string, ttlSeconds int))
// =============================================================================

func TestRedis_IsRequestDuplicate_False(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectSetNX("dedup:req:req-1", "1", 30*time.Second).SetVal(true)
	isDup, err := rc.IsRequestDuplicate("req-1", 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isDup {
		t.Error("expected not duplicate on first call")
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_IsRequestDuplicate_True(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectSetNX("dedup:req:req-1", "1", 30*time.Second).SetVal(false)
	isDup, err := rc.IsRequestDuplicate("req-1", 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isDup {
		t.Error("expected duplicate on second call")
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Publisher Fraud  (IncrementPublisherFraud(publisherID string)  no fraud type)
// =============================================================================

func TestRedis_IncrementPublisherFraud(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	fraudKey := fmt.Sprintf("fraud:publisher:pub-1:%s:count", dateStr)
	setKey := fmt.Sprintf("fraud:publishers:active:%s", dateStr)
	mock.ExpectIncr(fraudKey).SetVal(1)
	mock.ExpectSAdd(setKey, "pub-1").SetVal(1)
	if err := rc.IncrementPublisherFraud("pub-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// =============================================================================
// IncrementBidFormat / GetBidFormats
// =============================================================================

func TestRedis_IncrementBidFormat(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectIncr("stats:bids:format:banner").SetVal(1)
	if err := rc.IncrementBidFormat("banner"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRedis_GetBidFormats(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("stats:bids:format:banner").SetVal("5")
	mock.ExpectGet("stats:bids:format:video").SetVal("3")
	mock.ExpectGet("stats:bids:format:native").RedisNil()
	mock.ExpectGet("stats:bids:format:audio").RedisNil()
	result, err := rc.GetBidFormats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["banner"] != 5 || result["video"] != 3 {
		t.Errorf("unexpected formats: %v", result)
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Campaign CTR / Win Rate
// =============================================================================

func TestRedis_IncrementCampaignClicks(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("campaign:clicks:camp-1:%s", dateStr)
	mock.ExpectIncr(key).SetVal(10)
	if err := rc.IncrementCampaignClicks("camp-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_IncrementCampaignImpressions(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("campaign:imps:camp-1:%s", dateStr)
	mock.ExpectIncr(key).SetVal(100)
	if err := rc.IncrementCampaignImpressions("camp-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_IncrementCampaignBids(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("campaign:bids:camp-1:%s", dateStr)
	mock.ExpectIncr(key).SetVal(50)
	if err := rc.IncrementCampaignBids("camp-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_IncrementCampaignWins(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("campaign:wins:camp-1:%s", dateStr)
	mock.ExpectIncr(key).SetVal(20)
	if err := rc.IncrementCampaignWins("camp-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetCampaignCTR_ZeroImpressions(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	clicksKey := fmt.Sprintf("campaign:clicks:camp-1:%s", dateStr)
	impsKey := fmt.Sprintf("campaign:imps:camp-1:%s", dateStr)
	mock.ExpectGet(clicksKey).RedisNil()
	mock.ExpectGet(impsKey).RedisNil()
	ctr, err := rc.GetCampaignCTR("camp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctr != 0 {
		t.Errorf("expected 0 CTR, got %f", ctr)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetCampaignWinRate_NoData(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	bidsKey := fmt.Sprintf("campaign:bids:camp-1:%s", dateStr)
	winsKey := fmt.Sprintf("campaign:wins:camp-1:%s", dateStr)
	mock.ExpectGet(bidsKey).RedisNil()
	mock.ExpectGet(winsKey).RedisNil()
	wr, err := rc.GetCampaignWinRate("camp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wr != 0.5 {
		t.Errorf("expected default 0.5, got %f", wr)
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Bid Landscape
// =============================================================================

func TestRedis_RecordBidInBucket(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("landscape:bids:1.00-2.00:%s", dateStr)
	mock.ExpectIncr(key).SetVal(1)
	if err := rc.RecordBidInBucket("1.00-2.00"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_RecordWinInBucket(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("landscape:wins:2.00-5.00:%s", dateStr)
	mock.ExpectIncr(key).SetVal(1)
	if err := rc.RecordWinInBucket("2.00-5.00"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Segment Performance
// =============================================================================

func TestRedis_IncrementSegmentImpressions(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("segment:device:imps:mobile:%s", dateStr)
	mock.ExpectIncr(key).SetVal(1)
	if err := rc.IncrementSegmentImpressions("device", "mobile"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_IncrementSegmentClicks(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	dateStr := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("segment:os:clicks:android:%s", dateStr)
	mock.ExpectIncr(key).SetVal(1)
	if err := rc.IncrementSegmentClicks("os", "android"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Attribution  (RecordImpression/RecordClick have ttlHours int param)
// =============================================================================

func TestRedis_RecordImpression(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	// Value contains dynamic timestamp, use CustomMatch to accept any value
	mock.CustomMatch(func(expected, actual []interface{}) error {
		return nil
	}).ExpectSet("attr:imp:user-1:camp-1", "", 24*time.Hour).SetVal("OK")
	if err := rc.RecordImpression("user-1", "camp-1", "req-1", 24); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRedis_RecordClick(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.CustomMatch(func(expected, actual []interface{}) error {
		return nil
	}).ExpectSet("attr:click:user-1:camp-1", "", 24*time.Hour).SetVal("OK")
	if err := rc.RecordClick("user-1", "camp-1", "req-1", 24); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRedis_GetAttribution_None(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("attr:click:user-1:camp-1").RedisNil()
	mock.ExpectGet("attr:imp:user-1:camp-1").RedisNil()
	attrType, reqID, err := rc.GetAttribution("user-1", "camp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attrType != "none" || reqID != "" {
		t.Errorf("expected none, got type=%s reqID=%s", attrType, reqID)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetAttribution_CTA(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("attr:click:user-1:camp-1").SetVal("req-1:12345")
	attrType, reqID, err := rc.GetAttribution("user-1", "camp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attrType != "cta" {
		t.Errorf("expected cta, got %s", attrType)
	}
	if reqID != "req-1" {
		t.Errorf("expected req-1, got %s", reqID)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetAttribution_VTA(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("attr:click:user-1:camp-1").RedisNil()
	mock.ExpectGet("attr:imp:user-1:camp-1").SetVal("req-2:99999")
	attrType, reqID, err := rc.GetAttribution("user-1", "camp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attrType != "vta" {
		t.Errorf("expected vta, got %s", attrType)
	}
	if reqID != "req-2" {
		t.Errorf("expected req-2, got %s", reqID)
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// User Events  (RecordUserEvent(userID, campaignID, eventType string, ttlDays int))
// =============================================================================

func TestRedis_RecordUserEvent(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	// Value is a dynamic Unix timestamp; accept any value
	mock.CustomMatch(func(expected, actual []interface{}) error {
		return nil
	}).ExpectSet("retarget:click:user-1:camp-1", "", 7*24*time.Hour).SetVal("OK")
	if err := rc.RecordUserEvent("user-1", "camp-1", "click", 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRedis_RecordUserEvent_MissingParams(t *testing.T) {
	rc, _ := newTestRedisCache(t)
	if err := rc.RecordUserEvent("", "camp-1", "click", 7); err == nil {
		t.Error("expected error for empty userID")
	}
}

func TestRedis_HasUserEvent_True(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectExists("retarget:click:user-1:camp-1").SetVal(1)
	has, err := rc.HasUserEvent("user-1", "camp-1", "click")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has {
		t.Error("expected true")
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_HasUserEvent_False(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectExists("retarget:click:user-1:camp-1").SetVal(0)
	has, err := rc.HasUserEvent("user-1", "camp-1", "click")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if has {
		t.Error("expected false")
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Touchpoints
// =============================================================================

func TestRedis_RecordTouchpoint_MissingParams(t *testing.T) {
	rc, _ := newTestRedisCache(t)
	if err := rc.RecordTouchpoint("", "camp-1", "view", "req-1", 30); err == nil {
		t.Error("expected error for empty userID")
	}
}

// =============================================================================
// Cross-Device Graph
// =============================================================================

func TestRedis_GetPrimaryUserID_Hit(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("xdev:primary:device-a").SetVal("user-primary")
	uid, err := rc.GetPrimaryUserID("device-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid != "user-primary" {
		t.Errorf("expected user-primary, got %s", uid)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetPrimaryUserID_Miss(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("xdev:primary:device-a").RedisNil()
	uid, err := rc.GetPrimaryUserID("device-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid != "" {
		t.Errorf("expected empty, got %s", uid)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetLinkedDevices_NoGraph(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("xdev:primary:device-a").RedisNil()
	devices, err := rc.GetLinkedDevices("device-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 1 || devices[0] != "device-a" {
		t.Errorf("expected [device-a], got %v", devices)
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// Bid Path Analytics
// =============================================================================

func TestRedis_StoreBidPathAnalytics(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	analytics := &model.BidPathAnalytics{RequestID: "req-1"}
	data, _ := json.Marshal(analytics)
	mock.ExpectSet("spo:analytics:req-1", data, 24*time.Hour).SetVal("OK")
	if err := rc.StoreBidPathAnalytics(analytics); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetBidPathAnalytics_Hit(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	analytics := &model.BidPathAnalytics{RequestID: "req-1", WonAuction: true}
	data, _ := json.Marshal(analytics)
	mock.ExpectGet("spo:analytics:req-1").SetVal(string(data))
	result, err := rc.GetBidPathAnalytics("req-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RequestID != "req-1" {
		t.Errorf("unexpected request ID: %s", result.RequestID)
	}
	_ = mock.ExpectationsWereMet()
}

func TestRedis_GetBidPathAnalytics_Miss(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectGet("spo:analytics:req-1").RedisNil()
	_, err := rc.GetBidPathAnalytics("req-1")
	if err == nil {
		t.Error("expected error on miss")
	}
	_ = mock.ExpectationsWereMet()
}

// =============================================================================
// getPriceBucketForFloor (internal helper)
// =============================================================================

func TestGetPriceBucketForFloor(t *testing.T) {
	cases := []struct {
		price    float64
		expected string
	}{
		{0.25, "0.00-0.50"},
		{0.75, "0.50-1.00"},
		{1.50, "1.00-2.00"},
		{3.00, "2.00-5.00"},
		{7.00, "5.00-10.00"},
		{15.00, "10.00+"},
	}
	for _, tc := range cases {
		got := getPriceBucketForFloor(tc.price)
		if got != tc.expected {
			t.Errorf("price %v: expected %s, got %s", tc.price, tc.expected, got)
		}
	}
}

// =============================================================================
// Health / Close
// =============================================================================

func TestRedis_Health(t *testing.T) {
	rc, mock := newTestRedisCache(t)
	mock.ExpectPing().SetVal("PONG")
	if err := rc.Health(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRedis_Close(t *testing.T) {
	rc, _ := newTestRedisCache(t)
	// Close should not return error for mock client
	rc.Close()
}
