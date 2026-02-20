package service

import (
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/model"
)

// Test video context extraction functions

func TestExtractVideoContext(t *testing.T) {
	s := createBiddingUtilsService()

	tests := []struct {
		name        string
		req         *model.BidRequest
		wantIsVideo bool
		wantWidth   int
		wantHeight  int
	}{
		{
			name:        "no context",
			req:         &model.BidRequest{},
			wantIsVideo: false,
		},
		{
			name: "video flag true",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"video": true,
				},
			},
			wantIsVideo: true,
		},
		{
			name: "ad_type video",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"ad_type": "video",
				},
			},
			wantIsVideo: true,
		},
		{
			name: "with player dimensions",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"video":         true,
					"player_width":  640.0,
					"player_height": 480.0,
				},
			},
			wantIsVideo: true,
			wantWidth:   640,
			wantHeight:  480,
		},
		{
			name: "with video placement",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"video":           true,
					"video_placement": "instream",
				},
			},
			wantIsVideo: true,
		},
		{
			name: "with duration settings",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"video":       true,
					"minduration": 15.0,
					"maxduration": 30.0,
					"duration":    20.0,
				},
			},
			wantIsVideo: true,
		},
		{
			name: "with skip settings",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"video":     true,
					"skip":      true,
					"skipafter": 5.0,
				},
			},
			wantIsVideo: true,
		},
		{
			name: "with linearity and startdelay",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"video":      true,
					"linearity":  1.0,
					"startdelay": 0.0,
				},
			},
			wantIsVideo: true,
		},
		{
			name: "with completion rate",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"video":           true,
					"completion_rate": 0.85,
				},
			},
			wantIsVideo: true,
		},
		{
			name: "with video mimes",
			req: &model.BidRequest{
				Context: map[string]interface{}{
					"video":       true,
					"video_mimes": []interface{}{"video/mp4", "video/webm"},
				},
			},
			wantIsVideo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.extractVideoContext(tt.req)
			if result.isVideo != tt.wantIsVideo {
				t.Errorf("extractVideoContext().isVideo = %v, want %v", result.isVideo, tt.wantIsVideo)
			}
			if tt.wantWidth > 0 && result.playerWidth != tt.wantWidth {
				t.Errorf("extractVideoContext().playerWidth = %v, want %v", result.playerWidth, tt.wantWidth)
			}
			if tt.wantHeight > 0 && result.playerHeight != tt.wantHeight {
				t.Errorf("extractVideoContext().playerHeight = %v, want %v", result.playerHeight, tt.wantHeight)
			}
		})
	}
}

func TestMatchesPlayerSize(t *testing.T) {
	s := createBiddingUtilsService()

	tests := []struct {
		name   string
		width  int
		height int
		rule   model.VideoPlayerSize
		want   bool
	}{
		{
			name:   "unknown size matches unknown",
			width:  0,
			height: 0,
			rule:   model.VideoPlayerSize{Size: "unknown"},
			want:   true,
		},
		{
			name:   "large width matches large",
			width:  800,
			height: 600,
			rule:   model.VideoPlayerSize{Size: "large"},
			want:   true,
		},
		{
			name:   "xlarge width matches xlarge",
			width:  1920,
			height: 1080,
			rule:   model.VideoPlayerSize{Size: "xlarge"},
			want:   true,
		},
		{
			name:   "small width matches small",
			width:  320,
			height: 240,
			rule:   model.VideoPlayerSize{Size: "small"},
			want:   true,
		},
		{
			name:   "medium width matches medium",
			width:  500,
			height: 400,
			rule:   model.VideoPlayerSize{Size: "medium"},
			want:   true,
		},
		{
			name:   "below min width fails",
			width:  300,
			height: 200,
			rule:   model.VideoPlayerSize{Size: "large", MinWidth: 640},
			want:   false,
		},
		{
			name:   "above max width fails",
			width:  1920,
			height: 1080,
			rule:   model.VideoPlayerSize{Size: "large", MaxWidth: 1000},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.matchesPlayerSize(tt.width, tt.height, tt.rule); got != tt.want {
				t.Errorf("matchesPlayerSize(%d, %d) = %v, want %v", tt.width, tt.height, got, tt.want)
			}
		})
	}
}

func TestCategorizePlayerSize(t *testing.T) {
	s := createBiddingUtilsService()

	tests := []struct {
		name   string
		width  int
		height int
		want   string
	}{
		{
			name:   "unknown when zero",
			width:  0,
			height: 0,
			want:   "unknown",
		},
		{
			name:   "xlarge width",
			width:  1920,
			height: 1080,
			want:   "xlarge",
		},
		{
			name:   "large width",
			width:  800,
			height: 600,
			want:   "large",
		},
		{
			name:   "medium width",
			width:  500,
			height: 400,
			want:   "medium",
		},
		{
			name:   "small width",
			width:  300,
			height: 200,
			want:   "small",
		},
		{
			name:   "use height when width is zero",
			width:  0,
			height: 720,
			want:   "large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.categorizePlayerSize(tt.width, tt.height); got != tt.want {
				t.Errorf("categorizePlayerSize(%d, %d) = %v, want %v", tt.width, tt.height, got, tt.want)
			}
		})
	}
}

func TestGetDefaultPlayerSizeBoost(t *testing.T) {
	s := createBiddingUtilsService()

	tests := []struct {
		size string
		want float64
	}{
		{"xlarge", 1.4},
		{"XLARGE", 1.4},
		{"large", 1.2},
		{"Large", 1.2},
		{"medium", 1.0},
		{"MEDIUM", 1.0},
		{"small", 0.8},
		{"Small", 0.8},
		{"unknown", 1.0},
		{"other", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.size, func(t *testing.T) {
			if got := s.getDefaultPlayerSizeBoost(tt.size); got != tt.want {
				t.Errorf("getDefaultPlayerSizeBoost(%q) = %v, want %v", tt.size, got, tt.want)
			}
		})
	}
}

func TestEvaluateSkipSettings(t *testing.T) {
	s := createBiddingUtilsService()

	tests := []struct {
		name       string
		skippable  bool
		skipOffset int
		settings   *model.VideoSkipSettings
		wantBlock  bool
		wantReason string
	}{
		{
			name:       "skippable only - blocked non-skippable",
			skippable:  false,
			skipOffset: 0,
			settings:   &model.VideoSkipSettings{SkippableOnly: true},
			wantBlock:  true,
			wantReason: "requires_skippable",
		},
		{
			name:       "skippable only - allowed skippable",
			skippable:  true,
			skipOffset: 5,
			settings:   &model.VideoSkipSettings{SkippableOnly: true},
			wantBlock:  false,
		},
		{
			name:       "non-skippable only - blocked skippable",
			skippable:  true,
			skipOffset: 5,
			settings:   &model.VideoSkipSettings{NonSkippableOnly: true},
			wantBlock:  true,
			wantReason: "requires_non_skippable",
		},
		{
			name:       "non-skippable only - allowed non-skippable",
			skippable:  false,
			skipOffset: 0,
			settings:   &model.VideoSkipSettings{NonSkippableOnly: true},
			wantBlock:  false,
		},
		{
			name:       "skip offset too short",
			skippable:  true,
			skipOffset: 3,
			settings:   &model.VideoSkipSettings{MinSkipOffset: 5},
			wantBlock:  true,
			wantReason: "skip_offset_too_short",
		},
		{
			name:       "skip offset acceptable",
			skippable:  true,
			skipOffset: 10,
			settings:   &model.VideoSkipSettings{MinSkipOffset: 5},
			wantBlock:  false,
		},
		{
			name:       "skip offset too long",
			skippable:  true,
			skipOffset: 20,
			settings:   &model.VideoSkipSettings{MaxSkipOffset: 15},
			wantBlock:  true,
			wantReason: "skip_offset_too_long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.evaluateSkipSettings(tt.skippable, tt.skipOffset, tt.settings)
			if result.blocked != tt.wantBlock {
				t.Errorf("evaluateSkipSettings().blocked = %v, want %v", result.blocked, tt.wantBlock)
			}
			if tt.wantReason != "" && result.reason != tt.wantReason {
				t.Errorf("evaluateSkipSettings().reason = %v, want %v", result.reason, tt.wantReason)
			}
		})
	}
}

// Note: TestExtractPageContent and TestExtractPageCategories are in bidding_utils_test.go

// Test cross-device functions

func TestResolveCrossDeviceUser(t *testing.T) {
	s := createBiddingUtilsService()

	tests := []struct {
		name     string
		deviceID string
		signals  map[string]string
		wantPrim string
	}{
		{
			name:     "empty device ID returns default",
			deviceID: "",
			signals:  nil,
			wantPrim: "",
		},
		{
			name:     "device ID without signals returns device ID",
			deviceID: "device123",
			signals:  nil,
			wantPrim: "device123",
		},
		{
			name:     "device ID with empty signals",
			deviceID: "device456",
			signals:  map[string]string{},
			wantPrim: "device456",
		},
		{
			name:     "device ID with email hash signal",
			deviceID: "device789",
			signals: map[string]string{
				"email_hash": "abc123hash",
			},
			wantPrim: "device789", // No match in cache
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.ResolveCrossDeviceUser(tt.deviceID, tt.signals)
			if result == nil {
				t.Fatal("ResolveCrossDeviceUser() returned nil")
			}
			if result.PrimaryUserID != tt.wantPrim {
				t.Errorf("ResolveCrossDeviceUser().PrimaryUserID = %v, want %v", result.PrimaryUserID, tt.wantPrim)
			}
			// Verify linked devices contains at least the original device
			if tt.deviceID != "" && len(result.LinkedDevices) == 0 {
				t.Errorf("ResolveCrossDeviceUser().LinkedDevices should not be empty")
			}
		})
	}
}

func TestFindDeterministicLink(t *testing.T) {
	s := createBiddingUtilsService()

	tests := []struct {
		name     string
		deviceID string
		signals  map[string]string
		want     string
	}{
		{
			name:     "nil signals returns empty",
			deviceID: "device1",
			signals:  nil,
			want:     "",
		},
		{
			name:     "empty signals returns empty",
			deviceID: "device2",
			signals:  map[string]string{},
			want:     "",
		},
		{
			name:     "email hash not in cache",
			deviceID: "device3",
			signals: map[string]string{
				"email_hash": "newhash123",
			},
			want: "", // Not found in cache
		},
		{
			name:     "phone hash not in cache",
			deviceID: "device4",
			signals: map[string]string{
				"phone_hash": "phonehash456",
			},
			want: "",
		},
		{
			name:     "login_id not in cache",
			deviceID: "device5",
			signals: map[string]string{
				"login_id": "user@example.com",
			},
			want: "",
		},
		{
			name:     "hh_id not in cache",
			deviceID: "device6",
			signals: map[string]string{
				"hh_id": "household123",
			},
			want: "",
		},
		{
			name:     "uid2 not in cache",
			deviceID: "device7",
			signals: map[string]string{
				"uid2": "unifiedid2value",
			},
			want: "",
		},
		{
			name:     "rampid not in cache",
			deviceID: "device8",
			signals: map[string]string{
				"rampid": "liverampid123",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.findDeterministicLink(tt.deviceID, tt.signals); got != tt.want {
				t.Errorf("findDeterministicLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckCrossDeviceFreqCap(t *testing.T) {
	s := createBiddingUtilsService()

	tests := []struct {
		name       string
		campaign   *model.Campaign
		req        *model.BidRequest
		wantExceed bool
	}{
		{
			name: "cross-device not enabled",
			campaign: &model.Campaign{
				Targeting: model.Targeting{
					CrossDeviceEnabled: false,
				},
			},
			req:        &model.BidRequest{},
			wantExceed: false,
		},
		{
			name: "no user or device ID",
			campaign: &model.Campaign{
				Targeting: model.Targeting{
					CrossDeviceEnabled: true,
				},
			},
			req: &model.BidRequest{
				User:   model.InternalUser{ID: ""},
				Device: model.InternalDevice{DeviceID: ""},
			},
			wantExceed: false,
		},
		{
			name: "with user ID and freq cap",
			campaign: &model.Campaign{
				ID: "camp1",
				Targeting: model.Targeting{
					CrossDeviceEnabled: true,
					FreqCapImpressions: 10,
				},
			},
			req: &model.BidRequest{
				User: model.InternalUser{ID: "user123"},
			},
			wantExceed: false, // No frequency in cache yet
		},
		{
			name: "with device ID and freq cap",
			campaign: &model.Campaign{
				ID: "camp2",
				Targeting: model.Targeting{
					CrossDeviceEnabled: true,
					FreqCapImpressions: 5,
				},
			},
			req: &model.BidRequest{
				User:   model.InternalUser{ID: ""},
				Device: model.InternalDevice{DeviceID: "device456"},
			},
			wantExceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exceeded, _ := s.checkCrossDeviceFreqCap(tt.campaign, tt.req)
			if exceeded != tt.wantExceed {
				t.Errorf("checkCrossDeviceFreqCap() exceeded = %v, want %v", exceeded, tt.wantExceed)
			}
		})
	}
}
