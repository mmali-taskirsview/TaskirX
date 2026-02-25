package service

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// ─────────────────────────────────────────────────────────────────────────────
// generateNative tests
// ─────────────────────────────────────────────────────────────────────────────

func TestGenerateNative_DefaultMapping(t *testing.T) {
	camp := newCampaign(1.0)
	camp.Creative.Title = "Best Offer"
	camp.Creative.Description = "Try it today"
	camp.Creative.IconURL = "https://cdn.example.com/icon.png"
	camp.Creative.URL = "https://cdn.example.com/img.png"

	result := generateNative(camp, "https://imp.example.com", "https://click.example.com", "")

	if result == "" {
		t.Fatal("expected non-empty native JSON")
	}

	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	native, ok := resp["native"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'native' key in response")
	}
	assets, ok := native["assets"].([]interface{})
	if !ok || len(assets) == 0 {
		t.Fatal("expected non-empty assets array")
	}
}

func TestGenerateNative_InvalidRequestRaw_FallsBackToDefault(t *testing.T) {
	camp := newCampaign(1.0)
	camp.Creative.Title = "Fallback Ad"

	result := generateNative(camp, "", "", "{invalid json}")

	if result == "" {
		t.Fatal("expected non-empty native JSON on invalid requestRaw")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	_, ok := resp["native"]
	if !ok {
		t.Fatal("expected 'native' key in fallback response")
	}
}

func TestGenerateNative_TitleAsset(t *testing.T) {
	camp := newCampaign(1.0)
	camp.Creative.Title = "Mapped Title"

	rawReq := `{"assets":[{"id":1,"title":{}}]}`
	result := generateNative(camp, "", "", rawReq)

	if result == "" {
		t.Fatal("expected non-empty result with title asset")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	native := resp["native"].(map[string]interface{})
	assets := native["assets"].([]interface{})
	found := false
	for _, a := range assets {
		asset := a.(map[string]interface{})
		if _, has := asset["title"]; has {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a title asset in response")
	}
}

func TestGenerateNative_IconImageAsset(t *testing.T) {
	camp := newCampaign(1.0)
	camp.Creative.IconURL = "https://cdn.example.com/icon.png"

	// type:1 = icon
	rawReq := `{"assets":[{"id":2,"img":{"type":1,"w":50,"h":50}}]}`
	result := generateNative(camp, "", "", rawReq)

	if result == "" {
		t.Fatal("expected non-empty result with icon img asset")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	native := resp["native"].(map[string]interface{})
	assets := native["assets"].([]interface{})
	found := false
	for _, a := range assets {
		asset := a.(map[string]interface{})
		if _, has := asset["img"]; has {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected an img asset in response")
	}
}

func TestGenerateNative_MainImageAsset(t *testing.T) {
	camp := newCampaign(1.0)
	camp.Creative.URL = "https://cdn.example.com/main.png"
	camp.Creative.Width = 300
	camp.Creative.Height = 250

	// type:3 = main image
	rawReq := `{"assets":[{"id":3,"img":{"type":3,"w":300,"h":250}}]}`
	result := generateNative(camp, "", "", rawReq)

	if result == "" {
		t.Fatal("expected non-empty result with main img asset")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	native := resp["native"].(map[string]interface{})
	assets := native["assets"].([]interface{})
	if len(assets) == 0 {
		t.Error("expected at least one asset")
	}
}

func TestGenerateNative_DataAssets(t *testing.T) {
	camp := newCampaign(1.0)
	camp.Creative.Description = "Great Product"
	camp.Creative.CTAText = "Buy Now"

	// type:2 = desc, type:12 = CTA
	rawReq := `{"assets":[{"id":4,"data":{"type":2}},{"id":5,"data":{"type":12}}]}`
	result := generateNative(camp, "", "", rawReq)

	if result == "" {
		t.Fatal("expected non-empty result with data assets")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	native := resp["native"].(map[string]interface{})
	assets := native["assets"].([]interface{})
	dataCount := 0
	for _, a := range assets {
		asset := a.(map[string]interface{})
		if _, has := asset["data"]; has {
			dataCount++
		}
	}
	if dataCount == 0 {
		t.Error("expected at least one data asset in response")
	}
}

func TestGenerateNative_SponsoredDataAsset(t *testing.T) {
	camp := newCampaign(1.0)

	// type:1 = sponsored brand
	rawReq := `{"assets":[{"id":10,"data":{"type":1}}]}`
	result := generateNative(camp, "", "", rawReq)

	if result == "" {
		t.Fatal("expected non-empty result with sponsored data asset")
	}
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	_, ok := resp["native"]
	if !ok {
		t.Fatal("expected 'native' key in response")
	}
}

func TestGenerateNative_LinkInResponse(t *testing.T) {
	camp := newCampaign(1.0)

	result := generateNative(camp, "https://imp.test", "https://click.test", "")

	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	native := resp["native"].(map[string]interface{})
	if _, hasLink := native["link"]; !hasLink {
		t.Error("expected 'link' key in native response")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// calculateInventoryQualityMultiplier tests
// ─────────────────────────────────────────────────────────────────────────────

func makeIQService() *BiddingService {
	mc := NewMockCache()
	return NewBiddingService(mc, "")
}

func makeIQReq(ctx map[string]interface{}) *model.BidRequest {
	req := newReq()
	if ctx != nil {
		req.Context = ctx
	}
	return req
}

func TestCalcInventoryQuality_NilTargeting(t *testing.T) {
	svc := makeIQService()
	camp := newCampaign(1.0)
	camp.Targeting.InventoryQuality = nil

	req := makeIQReq(nil)
	result := svc.calculateInventoryQualityMultiplier(camp, req)

	if result.Blocked {
		t.Error("expected not blocked for nil InventoryQuality")
	}
	if result.Multiplier != 1.0 {
		t.Errorf("expected multiplier=1.0, got %v", result.Multiplier)
	}
}

func TestCalcInventoryQuality_QualityScoreTooLow(t *testing.T) {
	svc := makeIQService()
	camp := newCampaign(1.0)
	camp.Targeting.InventoryQuality = &model.InventoryQuality{
		MinQualityScore: 0.7,
	}

	req := makeIQReq(map[string]interface{}{
		"quality_score": 0.5,
	})
	result := svc.calculateInventoryQualityMultiplier(camp, req)

	if !result.Blocked {
		t.Error("expected Blocked=true for quality score too low")
	}
	if result.Reason != "quality_score_too_low" {
		t.Errorf("expected reason='quality_score_too_low', got '%s'", result.Reason)
	}
}

func TestCalcInventoryQuality_QualityScoreTooHigh(t *testing.T) {
	svc := makeIQService()
	camp := newCampaign(1.0)
	camp.Targeting.InventoryQuality = &model.InventoryQuality{
		MaxQualityScore: 0.6,
	}

	req := makeIQReq(map[string]interface{}{
		"quality_score": 0.9,
	})
	result := svc.calculateInventoryQualityMultiplier(camp, req)

	if !result.Blocked {
		t.Error("expected Blocked=true for quality score too high")
	}
	if result.Reason != "quality_score_too_high" {
		t.Errorf("expected reason='quality_score_too_high', got '%s'", result.Reason)
	}
}

func TestCalcInventoryQuality_TrustLevelNotAllowed(t *testing.T) {
	svc := makeIQService()
	camp := newCampaign(1.0)
	camp.Targeting.InventoryQuality = &model.InventoryQuality{
		TrustLevels: []string{"premium", "direct"},
	}

	req := makeIQReq(map[string]interface{}{
		"trust_level": "standard",
	})
	result := svc.calculateInventoryQualityMultiplier(camp, req)

	if !result.Blocked {
		t.Error("expected Blocked=true when trust level not in allowed list")
	}
	if result.Reason != "trust_level_not_allowed" {
		t.Errorf("expected reason='trust_level_not_allowed', got '%s'", result.Reason)
	}
}

func TestCalcInventoryQuality_ExcludeTrustLevel(t *testing.T) {
	svc := makeIQService()
	camp := newCampaign(1.0)
	camp.Targeting.InventoryQuality = &model.InventoryQuality{
		ExcludeTrustLevels: []string{"unknown"},
	}

	req := makeIQReq(map[string]interface{}{
		"trust_level": "unknown",
	})
	result := svc.calculateInventoryQualityMultiplier(camp, req)

	if !result.Blocked {
		t.Error("expected Blocked=true when trust level is in excluded list")
	}
}

func TestCalcInventoryQuality_RequireAdsTxtNotVerified(t *testing.T) {
	svc := makeIQService()
	camp := newCampaign(1.0)
	camp.Targeting.InventoryQuality = &model.InventoryQuality{
		RequireAdsTxt: true,
	}

	req := makeIQReq(map[string]interface{}{
		"ads_txt_verified": false,
	})
	result := svc.calculateInventoryQualityMultiplier(camp, req)

	if !result.Blocked {
		t.Error("expected Blocked=true when ads_txt required but not verified")
	}
	if result.Reason != "ads_txt_not_verified" {
		t.Errorf("expected reason='ads_txt_not_verified', got '%s'", result.Reason)
	}
}

func TestCalcInventoryQuality_RequireSellerJsonNotVerified(t *testing.T) {
	svc := makeIQService()
	camp := newCampaign(1.0)
	camp.Targeting.InventoryQuality = &model.InventoryQuality{
		RequireSellerJson: true,
	}

	req := makeIQReq(map[string]interface{}{
		"sellers_json_verified": false,
	})
	result := svc.calculateInventoryQualityMultiplier(camp, req)

	if !result.Blocked {
		t.Error("expected Blocked=true when sellers_json required but not verified")
	}
	if result.Reason != "sellers_json_not_verified" {
		t.Errorf("expected reason='sellers_json_not_verified', got '%s'", result.Reason)
	}
}

func TestCalcInventoryQuality_HighQualityBoost(t *testing.T) {
	svc := makeIQService()
	camp := newCampaign(1.0)
	camp.Targeting.InventoryQuality = &model.InventoryQuality{}

	// quality_score >= 0.8 → *1.2
	req := makeIQReq(map[string]interface{}{
		"quality_score": 0.85,
	})
	result := svc.calculateInventoryQualityMultiplier(camp, req)

	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.1 {
		t.Errorf("expected multiplier >= 1.1 for high quality score, got %v", result.Multiplier)
	}
}

func TestCalcInventoryQuality_MidQualityBoost(t *testing.T) {
	svc := makeIQService()
	camp := newCampaign(1.0)
	camp.Targeting.InventoryQuality = &model.InventoryQuality{}

	// quality_score >= 0.6 → *1.05
	req := makeIQReq(map[string]interface{}{
		"quality_score": 0.65,
	})
	result := svc.calculateInventoryQualityMultiplier(camp, req)

	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier < 1.0 {
		t.Errorf("expected multiplier >= 1.0 for mid quality score, got %v", result.Multiplier)
	}
}

func TestCalcInventoryQuality_LowQualityPenalty(t *testing.T) {
	svc := makeIQService()
	camp := newCampaign(1.0)
	camp.Targeting.InventoryQuality = &model.InventoryQuality{}

	// quality_score < 0.4 → *0.8
	req := makeIQReq(map[string]interface{}{
		"quality_score": 0.3,
	})
	result := svc.calculateInventoryQualityMultiplier(camp, req)

	if result.Blocked {
		t.Errorf("expected not blocked, reason: %s", result.Reason)
	}
	if result.Multiplier > 1.0 {
		t.Errorf("expected multiplier <= 1.0 for low quality score, got %v", result.Multiplier)
	}
}

func TestCalcInventoryQuality_InventoryQualityContextKey(t *testing.T) {
	svc := makeIQService()
	camp := newCampaign(1.0)
	camp.Targeting.InventoryQuality = &model.InventoryQuality{}

	// alternate context key "inventory_quality"
	req := makeIQReq(map[string]interface{}{
		"inventory_quality": 0.82,
	})
	result := svc.calculateInventoryQualityMultiplier(camp, req)

	if result.Blocked {
		t.Errorf("expected not blocked with inventory_quality key, reason: %s", result.Reason)
	}
}

func TestCalcInventoryQuality_TrustLevelAllowed(t *testing.T) {
	svc := makeIQService()
	camp := newCampaign(1.0)
	camp.Targeting.InventoryQuality = &model.InventoryQuality{
		TrustLevels: []string{"premium", "direct"},
	}

	req := makeIQReq(map[string]interface{}{
		"trust_level": "premium",
	})
	result := svc.calculateInventoryQualityMultiplier(camp, req)

	if result.Blocked {
		t.Errorf("expected not blocked when trust level is allowed, reason: %s", result.Reason)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// getDayOfWeekModifier tests
// ─────────────────────────────────────────────────────────────────────────────

func TestGetDayOfWeekModifier_CacheMiss(t *testing.T) {
	mc := NewMockCache()
	svc := NewDaypartingService(mc)

	avgPerf := hourlyPerformance{ctr: 0.03}
	result := svc.getDayOfWeekModifier("camp-1", 1, 7, avgPerf)

	if result != 1.0 {
		t.Errorf("expected 1.0 on cache miss, got %v", result)
	}
}

func TestGetDayOfWeekModifier_InsufficientImpressions(t *testing.T) {
	mc := NewMockCache()
	// impressions=30 < 50 threshold
	mc.kv["daypart_dow:camp-1:1"] = "impressions:30,ctr:0.10"
	svc := NewDaypartingService(mc)

	avgPerf := hourlyPerformance{ctr: 0.03}
	result := svc.getDayOfWeekModifier("camp-1", 1, 7, avgPerf)

	if result != 1.0 {
		t.Errorf("expected 1.0 for insufficient impressions, got %v", result)
	}
}

func TestGetDayOfWeekModifier_ZeroAvgCTR(t *testing.T) {
	mc := NewMockCache()
	mc.kv["daypart_dow:camp-1:2"] = "impressions:100,ctr:0.05"
	svc := NewDaypartingService(mc)

	avgPerf := hourlyPerformance{ctr: 0.0} // zero avgPerf.ctr
	result := svc.getDayOfWeekModifier("camp-1", 2, 7, avgPerf)

	if result != 1.0 {
		t.Errorf("expected 1.0 when avgPerf.ctr=0, got %v", result)
	}
}

func TestGetDayOfWeekModifier_HighRatioCapped(t *testing.T) {
	mc := NewMockCache()
	// ratio = 0.09/0.03 = 3.0 → capped at 1.2
	mc.kv["daypart_dow:camp-1:3"] = "impressions:200,ctr:0.09"
	svc := NewDaypartingService(mc)

	avgPerf := hourlyPerformance{ctr: 0.03}
	result := svc.getDayOfWeekModifier("camp-1", 3, 7, avgPerf)

	if result != 1.2 {
		t.Errorf("expected 1.2 (cap) for high ratio, got %v", result)
	}
}

func TestGetDayOfWeekModifier_LowRatioFloored(t *testing.T) {
	mc := NewMockCache()
	// ratio = 0.01/0.03 = 0.333 → floored at 0.8
	mc.kv["daypart_dow:camp-1:4"] = "impressions:100,ctr:0.01"
	svc := NewDaypartingService(mc)

	avgPerf := hourlyPerformance{ctr: 0.03}
	result := svc.getDayOfWeekModifier("camp-1", 4, 7, avgPerf)

	if result != 0.8 {
		t.Errorf("expected 0.8 (floor) for low ratio, got %v", result)
	}
}

func TestGetDayOfWeekModifier_RatioInRange(t *testing.T) {
	mc := NewMockCache()
	// ratio = 0.033/0.03 = 1.1 → within [0.8, 1.2] → return exact ratio
	mc.kv["daypart_dow:camp-1:5"] = "impressions:100,ctr:0.033"
	svc := NewDaypartingService(mc)

	avgPerf := hourlyPerformance{ctr: 0.03}
	result := svc.getDayOfWeekModifier("camp-1", 5, 7, avgPerf)

	expected := 0.033 / 0.03
	if result < 0.8 || result > 1.2 {
		t.Errorf("expected ratio in [0.8, 1.2], got %v", result)
	}
	diff := result - expected
	if diff < -0.001 || diff > 0.001 {
		t.Errorf("expected exact ratio=%.4f, got %.4f", expected, result)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// InsertTestPublisher tests
// ─────────────────────────────────────────────────────────────────────────────

func TestInsertTestPublisher_StoresPublisher(t *testing.T) {
	mc := NewMockCache()
	svc := NewDirectPublisherService(mc)

	pub := &DirectPublisher{
		ID:     "pub-test-001",
		Name:   "Test Publisher",
		Domain: "example.com",
		Status: "active",
	}

	svc.InsertTestPublisher(pub)

	val, ok := svc.publishers.Load("pub-test-001")
	if !ok {
		t.Fatal("expected publisher to be stored in sync.Map")
	}
	stored, ok := val.(*DirectPublisher)
	if !ok {
		t.Fatal("expected value to be *DirectPublisher")
	}
	if stored.Name != "Test Publisher" {
		t.Errorf("expected Name='Test Publisher', got '%s'", stored.Name)
	}
}

func TestInsertTestPublisher_OverwritesExisting(t *testing.T) {
	mc := NewMockCache()
	svc := NewDirectPublisherService(mc)

	pub1 := &DirectPublisher{ID: "pub-x", Name: "First"}
	pub2 := &DirectPublisher{ID: "pub-x", Name: "Second"}

	svc.InsertTestPublisher(pub1)
	svc.InsertTestPublisher(pub2)

	val, ok := svc.publishers.Load("pub-x")
	if !ok {
		t.Fatal("expected publisher to be stored")
	}
	stored := val.(*DirectPublisher)
	if stored.Name != "Second" {
		t.Errorf("expected overwritten Name='Second', got '%s'", stored.Name)
	}
}

func TestInsertTestPublisher_MultiplePublishers(t *testing.T) {
	mc := NewMockCache()
	svc := NewDirectPublisherService(mc)

	for i := 0; i < 5; i++ {
		pub := &DirectPublisher{
			ID:   fmt.Sprintf("pub-%d", i),
			Name: fmt.Sprintf("Publisher %d", i),
		}
		svc.InsertTestPublisher(pub)
	}

	count := 0
	svc.publishers.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	if count != 5 {
		t.Errorf("expected 5 publishers stored, got %d", count)
	}
}
