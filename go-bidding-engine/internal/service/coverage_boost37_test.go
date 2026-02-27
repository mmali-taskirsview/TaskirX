package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// BOOST 37: Target calculateVideoTargetingMultiplier (85.5%)

func newBiddingSvc_B37() *BiddingService {
	mc := NewMockCache()
	return NewBiddingService(mc, "")
}

// calculateVideoTargetingMultiplier Tests

func TestB37_VideoTargeting_NoConfig(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-1",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: nil, // No video targeting
		},
	}
	req := &model.BidRequest{
		ID:      "req-b37-1",
		Device:  model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	if result.Matched {
		t.Errorf("Expected not matched when no video targeting configured")
	}
	if result.Multiplier != 1.0 {
		t.Errorf("Expected 1.0 multiplier, got %f", result.Multiplier)
	}
}

func TestB37_VideoTargeting_NotVideoInventory_Blocked(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-2",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{
				PlayerSizes: []model.VideoPlayerSize{
					{Size: "large", MinWidth: 640},
				},
			},
		},
	}
	req := &model.BidRequest{
		ID:      "req-b37-2",
		Device:  model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{
			// No video context - not video inventory
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Errorf("Expected blocked when video targeting configured but not video inventory")
	}
	if result.Reason != "not_video_inventory" {
		t.Errorf("Expected 'not_video_inventory' reason, got '%s'", result.Reason)
	}
}

func TestB37_VideoTargeting_DurationTooShort(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-3",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{
				MinDuration: 30, // Require at least 30 seconds
			},
		},
	}
	req := &model.BidRequest{
		ID:     "req-b37-3",
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"video":       true,
			"maxduration": float64(15), // Only 15 seconds available
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Errorf("Expected blocked when video duration too short")
	}
	if result.Reason != "duration_too_short" {
		t.Errorf("Expected 'duration_too_short' reason, got '%s'", result.Reason)
	}
}

func TestB37_VideoTargeting_DurationTooLong(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-4",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{
				MaxDuration: 30, // Max 30 seconds
			},
		},
	}
	req := &model.BidRequest{
		ID:     "req-b37-4",
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"video":       true,
			"minduration": float64(60), // Minimum 60 seconds
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Errorf("Expected blocked when video duration too long")
	}
	if result.Reason != "duration_too_long" {
		t.Errorf("Expected 'duration_too_long' reason, got '%s'", result.Reason)
	}
}

func TestB37_VideoTargeting_PlacementMatch(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-5",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{
				Placements: []string{"in-stream"},
			},
		},
	}
	req := &model.BidRequest{
		ID:     "req-b37-5",
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"video":           true,
			"video_placement": "in-stream", // Matches
			"player_width":    float64(640),
			"player_height":   float64(480),
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("Expected not blocked on placement match, got reason: %s", result.Reason)
	}
	if !result.Matched {
		t.Errorf("Expected matched=true on placement match")
	}
}

func TestB37_VideoTargeting_PlacementMismatch(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-6",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{
				Placements: []string{"in-stream"},
			},
		},
	}
	req := &model.BidRequest{
		ID:     "req-b37-6",
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"video":           true,
			"video_placement": "in-banner", // Doesn't match
			"player_width":    float64(640),
			"player_height":   float64(480),
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Errorf("Expected blocked on placement mismatch")
	}
	if result.Reason != "placement_mismatch" {
		t.Errorf("Expected 'placement_mismatch' reason, got '%s'", result.Reason)
	}
}

func TestB37_VideoTargeting_LargePlayerBoost(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-7",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{},
		},
	}
	req := &model.BidRequest{
		ID:     "req-b37-7",
		Device: model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{
			"video":         true,
			"player_width":  float64(1920), // Large player width
			"player_height": float64(1080),
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	// Large player (>= 1280) should get 1.3x boost
	if result.Multiplier < 1.3 {
		t.Errorf("Expected at least 1.3x multiplier for large player, got %f", result.Multiplier)
	}
}

func TestB37_VideoTargeting_MediumPlayerBoost(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-8",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{},
		},
	}
	req := &model.BidRequest{
		ID:     "req-b37-8",
		Device: model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{
			"video":         true,
			"player_width":  float64(800), // Medium player width
			"player_height": float64(600),
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	// Medium player (640-1279) should get 1.1x boost
	if result.Multiplier < 1.1 {
		t.Errorf("Expected at least 1.1x multiplier for medium player, got %f", result.Multiplier)
	}
}

func TestB37_VideoTargeting_SmallPlayer(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-9",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{},
		},
	}
	req := &model.BidRequest{
		ID:     "req-b37-9",
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"video":         true,
			"player_width":  float64(400), // Small player
			"player_height": float64(300),
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	// Small player should have multiplier of 1.0 (no boost)
	if result.Multiplier != 1.0 {
		t.Errorf("Expected 1.0x multiplier for small player, got %f", result.Multiplier)
	}
}

func TestB37_VideoTargeting_CompletionRateHigh(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-10",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{
				CompletionRates: &model.CompletionRateRule{
					HighCompletionBoost: 1.5, // 1.5x boost for high completion
				},
			},
		},
	}
	req := &model.BidRequest{
		ID:     "req-b37-10",
		Device: model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{
			"video":           true,
			"player_width":    float64(800),
			"player_height":   float64(600),
			"completion_rate": 0.85, // 85% completion rate (>= 0.75)
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	// Should apply high completion boost
	if result.Multiplier < 1.5 {
		t.Errorf("Expected at least 1.5x multiplier for high completion rate, got %f", result.Multiplier)
	}
}

func TestB37_VideoTargeting_CompletionRateLow(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-11",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{
				CompletionRates: &model.CompletionRateRule{
					LowCompletionPenalty: 0.7, // 0.7x penalty for low completion
				},
			},
		},
	}
	req := &model.BidRequest{
		ID:     "req-b37-11",
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"video":           true,
			"player_width":    float64(400),
			"player_height":   float64(300),
			"completion_rate": 0.20, // 20% completion rate (< 0.25)
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	// Should apply low completion penalty
	if result.Multiplier > 0.8 {
		t.Errorf("Expected penalty (multiplier <= 0.8) for low completion rate, got %f", result.Multiplier)
	}
}

func TestB37_VideoTargeting_CompletionRateTooLow(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-12",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{
				CompletionRates: &model.CompletionRateRule{
					MinCompletionRate: 0.50, // Require at least 50%
				},
			},
		},
	}
	req := &model.BidRequest{
		ID:     "req-b37-12",
		Device: model.InternalDevice{Type: "mobile"},
		Context: map[string]interface{}{
			"video":           true,
			"player_width":    float64(640),
			"player_height":   float64(480),
			"completion_rate": 0.30, // Only 30% (< 50%)
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Errorf("Expected blocked when completion rate too low")
	}
	if result.Reason != "completion_rate_too_low" {
		t.Errorf("Expected 'completion_rate_too_low' reason, got '%s'", result.Reason)
	}
}

func TestB37_VideoTargeting_MimeTypeMatch(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-13",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{
				Mimes: []string{"video/mp4", "video/webm"},
			},
		},
	}
	req := &model.BidRequest{
		ID:     "req-b37-13",
		Device: model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{
			"video":         true,
			"player_width":  float64(800),
			"player_height": float64(600),
			"video_mimes":   []interface{}{"video/mp4"}, // Matches
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	if result.Blocked {
		t.Errorf("Expected not blocked on MIME type match, got reason: %s", result.Reason)
	}
}

func TestB37_VideoTargeting_MimeTypeMismatch(t *testing.T) {
	svc := newBiddingSvc_B37()
	camp := &model.Campaign{
		ID:       "camp-b37-14",
		BidPrice: 5.0,
		Targeting: model.Targeting{
			VideoTargeting: &model.VideoTargeting{
				Mimes: []string{"video/mp4", "video/webm"},
			},
		},
	}
	req := &model.BidRequest{
		ID:     "req-b37-14",
		Device: model.InternalDevice{Type: "desktop"},
		Context: map[string]interface{}{
			"video":         true,
			"player_width":  float64(800),
			"player_height": float64(600),
			"video_mimes":   []interface{}{"video/flv"}, // Doesn't match
		},
	}

	result := svc.calculateVideoTargetingMultiplier(camp, req)
	if !result.Blocked {
		t.Errorf("Expected blocked on MIME type mismatch")
	}
	if result.Reason != "mime_type_mismatch" {
		t.Errorf("Expected 'mime_type_mismatch' reason, got '%s'", result.Reason)
	}
}
