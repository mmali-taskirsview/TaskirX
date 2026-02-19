package service

import (
	"math"
	"testing"
	"time"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// --- Helper ---

func makeTouchpoints(count int, campaignID string, startTime time.Time, interval time.Duration) []model.Touchpoint {
	tps := make([]model.Touchpoint, count)
	for i := 0; i < count; i++ {
		tps[i] = model.Touchpoint{
			Type:       "impression",
			RequestID:  "req-" + string(rune('a'+i)),
			CampaignID: campaignID,
			Timestamp:  startTime.Add(time.Duration(i) * interval),
		}
	}
	return tps
}

func creditSum(credits []model.AttributionCredit) float64 {
	sum := 0.0
	for _, c := range credits {
		sum += c.Credit
	}
	return sum
}

func assertNear(t *testing.T, name string, got, want, eps float64) {
	t.Helper()
	if math.Abs(got-want) > eps {
		t.Errorf("%s: got %.6f, want %.6f (eps %.6f)", name, got, want, eps)
	}
}

// --- Tests ---

func TestAttribution_EmptyUserID(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	_, err := svc.CalculateAttribution("", "camp1", "last_click", 0)
	if err == nil {
		t.Fatal("expected error for empty userID")
	}
}

func TestAttribution_NoTouchpoints(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	credits, err := svc.CalculateAttribution("user1", "camp1", "last_click", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if credits != nil {
		t.Fatalf("expected nil credits for no touchpoints, got %d", len(credits))
	}
}

func TestAttribution_LastClick(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tps := makeTouchpoints(4, "camp1", base, time.Hour)
	mc.SetTouchpoints("user1", "camp1", tps)

	credits, err := svc.CalculateAttribution("user1", "camp1", "last_click", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 4 {
		t.Fatalf("expected 4 credits, got %d", len(credits))
	}

	// Last touchpoint should get 100% credit
	for i, c := range credits {
		if i == 3 {
			assertNear(t, "last_credit", c.Credit, 1.0, 0.001)
		} else {
			assertNear(t, "non_last_credit", c.Credit, 0.0, 0.001)
		}
		if c.Model != "last_click" {
			t.Errorf("expected model last_click, got %s", c.Model)
		}
	}

	assertNear(t, "total_credit", creditSum(credits), 1.0, 0.001)
}

func TestAttribution_FirstClick(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tps := makeTouchpoints(3, "camp1", base, time.Hour*24)
	mc.SetTouchpoints("user1", "camp1", tps)

	credits, err := svc.CalculateAttribution("user1", "camp1", "first_click", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 3 {
		t.Fatalf("expected 3 credits, got %d", len(credits))
	}

	// First touchpoint gets 100%
	assertNear(t, "first_credit", credits[0].Credit, 1.0, 0.001)
	assertNear(t, "second_credit", credits[1].Credit, 0.0, 0.001)
	assertNear(t, "third_credit", credits[2].Credit, 0.0, 0.001)
	assertNear(t, "total_credit", creditSum(credits), 1.0, 0.001)
}

func TestAttribution_Linear(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tps := makeTouchpoints(5, "camp1", base, time.Hour)
	mc.SetTouchpoints("user1", "camp1", tps)

	credits, err := svc.CalculateAttribution("user1", "camp1", "linear", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 5 {
		t.Fatalf("expected 5 credits, got %d", len(credits))
	}

	expected := 1.0 / 5.0
	for i, c := range credits {
		assertNear(t, "linear_credit_"+string(rune('0'+i)), c.Credit, expected, 0.001)
	}
	assertNear(t, "total_credit", creditSum(credits), 1.0, 0.001)
}

func TestAttribution_TimeDecay(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// 3 touchpoints spread 168 hours (1 half-life) apart
	tps := []model.Touchpoint{
		{Type: "impression", RequestID: "r1", CampaignID: "camp1", Timestamp: base},
		{Type: "click", RequestID: "r2", CampaignID: "camp1", Timestamp: base.Add(168 * time.Hour)},
		{Type: "click", RequestID: "r3", CampaignID: "camp1", Timestamp: base.Add(336 * time.Hour)},
	}
	mc.SetTouchpoints("user1", "camp1", tps)

	credits, err := svc.CalculateAttribution("user1", "camp1", "time_decay", 168)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 3 {
		t.Fatalf("expected 3 credits, got %d", len(credits))
	}

	// Most recent touchpoint should have highest credit
	if credits[2].Credit <= credits[1].Credit {
		t.Error("most recent touchpoint should have highest credit in time_decay")
	}
	if credits[1].Credit <= credits[0].Credit {
		t.Error("middle touchpoint should have more credit than earliest")
	}

	// Verify credits sum to 1
	assertNear(t, "total_credit", creditSum(credits), 1.0, 0.001)

	// Verify exponential decay property: ratio between consecutive should be ~2.0
	ratio := credits[2].Credit / credits[1].Credit
	assertNear(t, "decay_ratio", ratio, 2.0, 0.01)
}

func TestAttribution_TimeDecay_DefaultHalfLife(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tps := makeTouchpoints(2, "camp1", base, 168*time.Hour)
	mc.SetTouchpoints("user1", "camp1", tps)

	// Pass 0 for halfLife, should default to 168
	credits, err := svc.CalculateAttribution("user1", "camp1", "time_decay", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNear(t, "total_credit", creditSum(credits), 1.0, 0.001)
}

func TestAttribution_PositionBased_SingleTouchpoint(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tps := makeTouchpoints(1, "camp1", base, time.Hour)
	mc.SetTouchpoints("user1", "camp1", tps)

	credits, err := svc.CalculateAttribution("user1", "camp1", "position_based", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 1 {
		t.Fatalf("expected 1 credit, got %d", len(credits))
	}
	assertNear(t, "single_credit", credits[0].Credit, 1.0, 0.001)
}

func TestAttribution_PositionBased_TwoTouchpoints(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tps := makeTouchpoints(2, "camp1", base, time.Hour)
	mc.SetTouchpoints("user1", "camp1", tps)

	credits, err := svc.CalculateAttribution("user1", "camp1", "position_based", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNear(t, "first_credit", credits[0].Credit, 0.5, 0.001)
	assertNear(t, "last_credit", credits[1].Credit, 0.5, 0.001)
}

func TestAttribution_PositionBased_ManyTouchpoints(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tps := makeTouchpoints(5, "camp1", base, time.Hour)
	mc.SetTouchpoints("user1", "camp1", tps)

	credits, err := svc.CalculateAttribution("user1", "camp1", "position_based", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credits) != 5 {
		t.Fatalf("expected 5 credits, got %d", len(credits))
	}

	// First and last get 40% each
	assertNear(t, "first_credit", credits[0].Credit, 0.40, 0.001)
	assertNear(t, "last_credit", credits[4].Credit, 0.40, 0.001)

	// Middle 3 share the remaining 20%
	middleEach := 0.20 / 3.0
	for i := 1; i <= 3; i++ {
		assertNear(t, "middle_credit", credits[i].Credit, middleEach, 0.001)
	}

	assertNear(t, "total_credit", creditSum(credits), 1.0, 0.001)
}

func TestAttribution_UnknownModelFallsBackToLastClick(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tps := makeTouchpoints(3, "camp1", base, time.Hour)
	mc.SetTouchpoints("user1", "camp1", tps)

	credits, err := svc.CalculateAttribution("user1", "camp1", "unknown_model", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should fallback to last_click
	assertNear(t, "last_credit", credits[2].Credit, 1.0, 0.001)
}

func TestAttribution_BidAdjustment_NoData(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	// No touchpoints - should return 1.0
	mult := svc.GetAttributionBidAdjustment("camp1", "user1", "linear", 0)
	assertNear(t, "no_data_multiplier", mult, 1.0, 0.001)
}

func TestAttribution_BidAdjustment_EmptyUser(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	mult := svc.GetAttributionBidAdjustment("camp1", "", "linear", 0)
	assertNear(t, "empty_user_multiplier", mult, 1.0, 0.001)
}

func TestAttribution_BidAdjustment_WithData(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// All touchpoints from same campaign - should get full attribution
	tps := makeTouchpoints(3, "camp1", base, time.Hour)
	mc.SetTouchpoints("user1", "", tps)

	mult := svc.GetAttributionBidAdjustment("camp1", "user1", "linear", 0)
	// With all touchpoints from camp1, campaignCredit = 1.0 (sum of all credits)
	// avgCredit = 1.0/3, ratio = 1.0 / (1.0/3) = 3.0
	// multiplier = 0.5 + 3.0*0.5 = 2.0 (capped)
	if mult < 1.0 {
		t.Errorf("expected multiplier >= 1.0 for high-attribution campaign, got %.2f", mult)
	}
}

func TestAttribution_RecordConversionTouchpoint(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	err := svc.RecordConversionTouchpoint("user1", "camp1", "click", "req1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tps, _ := mc.GetTouchpoints("user1", "camp1")
	if len(tps) != 1 {
		t.Fatalf("expected 1 touchpoint, got %d", len(tps))
	}
	if tps[0].Type != "click" {
		t.Errorf("expected type click, got %s", tps[0].Type)
	}
}

func TestAttribution_CompareModels(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tps := makeTouchpoints(4, "camp1", base, time.Hour*24)
	mc.SetTouchpoints("user1", "camp1", tps)

	results, err := svc.CompareModels("user1", "camp1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedModels := []string{"last_click", "first_click", "linear", "time_decay", "position_based"}
	for _, m := range expectedModels {
		credits, ok := results[m]
		if !ok {
			t.Errorf("missing model %s in comparison", m)
			continue
		}
		if len(credits) != 4 {
			t.Errorf("model %s: expected 4 credits, got %d", m, len(credits))
			continue
		}
		assertNear(t, m+"_total", creditSum(credits), 1.0, 0.001)
	}
}

func TestAttribution_GetAttributionSummary(t *testing.T) {
	mc := NewMockCache()
	svc := NewAttributionService(mc)

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Mix of campaigns
	tps := []model.Touchpoint{
		{Type: "impression", RequestID: "r1", CampaignID: "campA", Timestamp: base},
		{Type: "click", RequestID: "r2", CampaignID: "campB", Timestamp: base.Add(time.Hour)},
		{Type: "click", RequestID: "r3", CampaignID: "campA", Timestamp: base.Add(2 * time.Hour)},
	}
	mc.SetTouchpoints("user1", "", tps)

	summary, err := svc.GetAttributionSummary("user1", "linear", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Linear: each touchpoint gets 1/3
	// campA has 2 touchpoints = 2/3, campB has 1 = 1/3
	assertNear(t, "campA_credit", summary["campA"], 2.0/3.0, 0.001)
	assertNear(t, "campB_credit", summary["campB"], 1.0/3.0, 0.001)
}
