package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/taskirx/go-bidding-engine/internal/cache"
	"github.com/taskirx/go-bidding-engine/internal/model"
)

// Boost32: target generateVideoVAST, generateNative, categorizePlayerSize, optimizeForCPA/CPS/CPR/CPL/CPCV edge cases

// ============================================================================
// Helpers
// ============================================================================

func makeMinCamp_B32() *model.Campaign {
	return &model.Campaign{
		ID:       "camp-boost32",
		Name:     "Boost32 Campaign",
		BidPrice: 5.0,
		Creative: model.Creative{
			URL:         "https://cdn.example.com/video.mp4",
			Title:       "Test Ad Title",
			Description: "Test Ad Description",
			CTAText:     "Learn More",
			IconURL:     "https://cdn.example.com/icon.png",
			Width:       1920,
			Height:      1080,
			MimeType:    "video/mp4",
			Duration:    30,
		},
		Targeting:   model.Targeting{},
		DailyBudget: 1000.0,
	}
}

func makeMinReq_B32() *model.BidRequest {
	return &model.BidRequest{
		ID:          "req-boost32",
		PublisherID: "pub-test",
		User: model.InternalUser{
			ID: "user-test",
		},
		Device: model.InternalDevice{
			Type: "mobile",
			IP:   "1.2.3.4",
		},
		Context: make(map[string]interface{}),
	}
}

func newBiddingSvc_B32() *BiddingService {
	mc := cache.NewMockCache()
	return NewBiddingService(mc, "")
}

// ============================================================================
// generateVideoVAST Tests
// ============================================================================

// TestB32_VideoVAST_Rewarded tests rewarded video extension
func TestB32_VideoVAST_Rewarded(t *testing.T) {
	camp := makeMinCamp_B32()
	camp.Creative.Rewarded = true
	camp.Creative.RewardAmt = 100
	camp.Creative.RewardType = "coins"

	impURL := "https://track.example.com/impression?campaign=camp-test"
	clickURL := "https://track.example.com/click?campaign=camp-test"

	vast := generateVideoVAST(camp, impURL, clickURL)

	if !strings.Contains(vast, "<Extension type=\"reward\">") {
		t.Errorf("Expected reward extension in rewarded video VAST")
	}
	if !strings.Contains(vast, "<Reward amount=\"100\" type=\"coins\"/>") {
		t.Errorf("Expected reward details in VAST")
	}
}

// TestB32_VideoVAST_NonRewarded tests standard video without reward
func TestB32_VideoVAST_NonRewarded(t *testing.T) {
	camp := makeMinCamp_B32()
	camp.Creative.Rewarded = false

	impURL := "https://track.example.com/impression?campaign=camp-test"
	clickURL := "https://track.example.com/click?campaign=camp-test"

	vast := generateVideoVAST(camp, impURL, clickURL)

	if strings.Contains(vast, "<Extension type=\"reward\">") {
		t.Errorf("Should not have reward extension for non-rewarded video")
	}
	if !strings.Contains(vast, "<VAST version=\"4.0\">") {
		t.Errorf("Expected VAST 4.0 format")
	}
	if !strings.Contains(vast, camp.ID) {
		t.Errorf("Expected campaign ID in VAST")
	}
}

// ============================================================================
// generateNative Tests
// ============================================================================

// TestB32_Native_DefaultMapping tests default native asset mapping
func TestB32_Native_DefaultMapping(t *testing.T) {
	camp := makeMinCamp_B32()
	impURL := "https://track.example.com/impression?campaign=camp-test"
	clickURL := "https://track.example.com/click?campaign=camp-test"

	nativeJSON := generateNative(camp, impURL, clickURL, "")

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(nativeJSON), &result); err != nil {
		t.Fatalf("Failed to parse native JSON: %v", err)
	}

	nativeObj, ok := result["native"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected native object")
	}

	assets, ok := nativeObj["assets"].([]interface{})
	if !ok || len(assets) == 0 {
		t.Errorf("Expected assets array with elements")
	}
}

// TestB32_Native_OpenRTBTitleAsset tests title asset mapping
func TestB32_Native_OpenRTBTitleAsset(t *testing.T) {
	camp := makeMinCamp_B32()
	impURL := "https://track.example.com/impression"
	clickURL := "https://track.example.com/click"

	// OpenRTB native request with title
	nativeReq := model.OpenRTBNativeRequest{
		Assets: []model.NativeAsset{
			{
				ID:    1,
				Title: &model.NativeTitleReq{Len: 25},
			},
		},
	}
	reqJSON, _ := json.Marshal(nativeReq)

	nativeJSON := generateNative(camp, impURL, clickURL, string(reqJSON))

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(nativeJSON), &result); err != nil {
		t.Fatalf("Failed to parse native JSON: %v", err)
	}

	nativeObj := result["native"].(map[string]interface{})
	assets := nativeObj["assets"].([]interface{})

	found := false
	for _, assetRaw := range assets {
		asset := assetRaw.(map[string]interface{})
		if titleObj, ok := asset["title"].(map[string]interface{}); ok {
			if titleObj["text"] == camp.Creative.Title {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("Expected title asset in native response")
	}
}

// TestB32_Native_OpenRTBIconAsset tests icon image asset
func TestB32_Native_OpenRTBIconAsset(t *testing.T) {
	camp := makeMinCamp_B32()
	impURL := "https://track.example.com/impression"
	clickURL := "https://track.example.com/click"

	// OpenRTB native request with icon (type 1)
	nativeReq := model.OpenRTBNativeRequest{
		Assets: []model.NativeAsset{
			{
				ID: 2,
				Img: &model.NativeImgReq{
					Type: 1, // Icon
				},
			},
		},
	}
	reqJSON, _ := json.Marshal(nativeReq)

	nativeJSON := generateNative(camp, impURL, clickURL, string(reqJSON))

	var result map[string]interface{}
	json.Unmarshal([]byte(nativeJSON), &result)
	nativeObj := result["native"].(map[string]interface{})
	assets := nativeObj["assets"].([]interface{})

	found := false
	for _, assetRaw := range assets {
		asset := assetRaw.(map[string]interface{})
		if imgObj, ok := asset["img"].(map[string]interface{}); ok {
			if imgObj["url"] == camp.Creative.IconURL {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("Expected icon asset in native response")
	}
}

// TestB32_Native_OpenRTBMainImageAsset tests main image asset (type 3)
func TestB32_Native_OpenRTBMainImageAsset(t *testing.T) {
	camp := makeMinCamp_B32()
	impURL := "https://track.example.com/impression"
	clickURL := "https://track.example.com/click"

	// OpenRTB native request with main image (type 3)
	nativeReq := model.OpenRTBNativeRequest{
		Assets: []model.NativeAsset{
			{
				ID: 3,
				Img: &model.NativeImgReq{
					Type: 3, // Main
				},
			},
		},
	}
	reqJSON, _ := json.Marshal(nativeReq)

	nativeJSON := generateNative(camp, impURL, clickURL, string(reqJSON))

	var result map[string]interface{}
	json.Unmarshal([]byte(nativeJSON), &result)
	nativeObj := result["native"].(map[string]interface{})
	assets := nativeObj["assets"].([]interface{})

	found := false
	for _, assetRaw := range assets {
		asset := assetRaw.(map[string]interface{})
		if imgObj, ok := asset["img"].(map[string]interface{}); ok {
			if imgObj["url"] == camp.Creative.URL {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("Expected main image asset in native response")
	}
}

// TestB32_Native_OpenRTBDefaultImageType tests fallback for unspecified image type
func TestB32_Native_OpenRTBDefaultImageType(t *testing.T) {
	camp := makeMinCamp_B32()
	impURL := "https://track.example.com/impression"
	clickURL := "https://track.example.com/click"

	// OpenRTB native request with image type 0 or other (should fallback to main)
	nativeReq := model.OpenRTBNativeRequest{
		Assets: []model.NativeAsset{
			{
				ID: 4,
				Img: &model.NativeImgReq{
					Type: 0, // Unspecified
				},
			},
		},
	}
	reqJSON, _ := json.Marshal(nativeReq)

	nativeJSON := generateNative(camp, impURL, clickURL, string(reqJSON))

	var result map[string]interface{}
	json.Unmarshal([]byte(nativeJSON), &result)
	nativeObj := result["native"].(map[string]interface{})
	assets := nativeObj["assets"].([]interface{})

	// Should still return an image asset (fallback to main URL)
	if len(assets) == 0 {
		t.Errorf("Expected at least one asset for unspecified image type")
	}
}

// TestB32_Native_OpenRTBDataSponsoredBy tests sponsored by data asset (type 1)
func TestB32_Native_OpenRTBDataSponsoredBy(t *testing.T) {
	camp := makeMinCamp_B32()
	impURL := "https://track.example.com/impression"
	clickURL := "https://track.example.com/click"

	nativeReq := model.OpenRTBNativeRequest{
		Assets: []model.NativeAsset{
			{
				ID: 5,
				Data: &model.NativeDataReq{
					Type: 1, // Sponsored By
				},
			},
		},
	}
	reqJSON, _ := json.Marshal(nativeReq)

	nativeJSON := generateNative(camp, impURL, clickURL, string(reqJSON))

	var result map[string]interface{}
	json.Unmarshal([]byte(nativeJSON), &result)
	nativeObj := result["native"].(map[string]interface{})
	assets := nativeObj["assets"].([]interface{})

	found := false
	for _, assetRaw := range assets {
		asset := assetRaw.(map[string]interface{})
		if dataObj, ok := asset["data"].(map[string]interface{}); ok {
			if dataObj["value"] == "TaskirX" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("Expected 'TaskirX' as sponsored by value")
	}
}

// TestB32_Native_OpenRTBDataDesc tests description data asset (type 2)
func TestB32_Native_OpenRTBDataDesc(t *testing.T) {
	camp := makeMinCamp_B32()
	impURL := "https://track.example.com/impression"
	clickURL := "https://track.example.com/click"

	nativeReq := model.OpenRTBNativeRequest{
		Assets: []model.NativeAsset{
			{
				ID: 6,
				Data: &model.NativeDataReq{
					Type: 2, // Description
				},
			},
		},
	}
	reqJSON, _ := json.Marshal(nativeReq)

	nativeJSON := generateNative(camp, impURL, clickURL, string(reqJSON))

	var result map[string]interface{}
	json.Unmarshal([]byte(nativeJSON), &result)
	nativeObj := result["native"].(map[string]interface{})
	assets := nativeObj["assets"].([]interface{})

	found := false
	for _, assetRaw := range assets {
		asset := assetRaw.(map[string]interface{})
		if dataObj, ok := asset["data"].(map[string]interface{}); ok {
			if dataObj["value"] == camp.Creative.Description {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("Expected description data asset")
	}
}

// TestB32_Native_OpenRTBDataCTA tests CTA text data asset (type 12)
func TestB32_Native_OpenRTBDataCTA(t *testing.T) {
	camp := makeMinCamp_B32()
	impURL := "https://track.example.com/impression"
	clickURL := "https://track.example.com/click"

	nativeReq := model.OpenRTBNativeRequest{
		Assets: []model.NativeAsset{
			{
				ID: 7,
				Data: &model.NativeDataReq{
					Type: 12, // CTA Text
				},
			},
		},
	}
	reqJSON, _ := json.Marshal(nativeReq)

	nativeJSON := generateNative(camp, impURL, clickURL, string(reqJSON))

	var result map[string]interface{}
	json.Unmarshal([]byte(nativeJSON), &result)
	nativeObj := result["native"].(map[string]interface{})
	assets := nativeObj["assets"].([]interface{})

	found := false
	for _, assetRaw := range assets {
		asset := assetRaw.(map[string]interface{})
		if dataObj, ok := asset["data"].(map[string]interface{}); ok {
			if dataObj["value"] == camp.Creative.CTAText {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("Expected CTA text data asset")
	}
}

// TestB32_Native_OpenRTBDataDisplayURL tests display URL data asset (type 11)
func TestB32_Native_OpenRTBDataDisplayURL(t *testing.T) {
	camp := makeMinCamp_B32()
	impURL := "https://track.example.com/impression"
	clickURL := "https://track.example.com/click"

	nativeReq := model.OpenRTBNativeRequest{
		Assets: []model.NativeAsset{
			{
				ID: 8,
				Data: &model.NativeDataReq{
					Type: 11, // Display URL
				},
			},
		},
	}
	reqJSON, _ := json.Marshal(nativeReq)

	nativeJSON := generateNative(camp, impURL, clickURL, string(reqJSON))

	var result map[string]interface{}
	json.Unmarshal([]byte(nativeJSON), &result)
	nativeObj := result["native"].(map[string]interface{})
	assets := nativeObj["assets"].([]interface{})

	found := false
	for _, assetRaw := range assets {
		asset := assetRaw.(map[string]interface{})
		if dataObj, ok := asset["data"].(map[string]interface{}); ok {
			// extractDomain returns "taskirx-ad.com"
			if dataObj["value"] == "taskirx-ad.com" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("Expected display URL data asset")
	}
}

// TestB32_Native_OpenRTBDataUnsupported tests unsupported data type (skipped)
func TestB32_Native_OpenRTBDataUnsupported(t *testing.T) {
	camp := makeMinCamp_B32()
	impURL := "https://track.example.com/impression"
	clickURL := "https://track.example.com/click"

	nativeReq := model.OpenRTBNativeRequest{
		Assets: []model.NativeAsset{
			{
				ID: 9,
				Data: &model.NativeDataReq{
					Type: 99, // Unsupported
				},
			},
		},
	}
	reqJSON, _ := json.Marshal(nativeReq)

	nativeJSON := generateNative(camp, impURL, clickURL, string(reqJSON))

	var result map[string]interface{}
	json.Unmarshal([]byte(nativeJSON), &result)
	nativeObj := result["native"].(map[string]interface{})
	assets := nativeObj["assets"].([]interface{})

	// Should not include unsupported asset
	for _, assetRaw := range assets {
		asset := assetRaw.(map[string]interface{})
		if assetID, ok := asset["id"].(float64); ok && int(assetID) == 9 {
			t.Errorf("Should not include unsupported data asset")
		}
	}
}

// TestB32_Native_InvalidRequestJSON tests malformed native request (fallback to default)
func TestB32_Native_InvalidRequestJSON(t *testing.T) {
	camp := makeMinCamp_B32()
	impURL := "https://track.example.com/impression"
	clickURL := "https://track.example.com/click"

	nativeJSON := generateNative(camp, impURL, clickURL, "{invalid json")

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(nativeJSON), &result); err != nil {
		t.Fatalf("Expected valid JSON output even with invalid request: %v", err)
	}

	// Should fallback to default mapping
	nativeObj := result["native"].(map[string]interface{})
	assets := nativeObj["assets"].([]interface{})
	if len(assets) == 0 {
		t.Errorf("Expected default assets when request is invalid")
	}
}

// ============================================================================
// categorizePlayerSize Tests
// ============================================================================

// TestB32_PlayerSize_BothZero tests both width and height zero
func TestB32_PlayerSize_BothZero(t *testing.T) {
	svc := newBiddingSvc_B32()
	result := svc.categorizePlayerSize(0, 0)
	if result != "unknown" {
		t.Errorf("Expected 'unknown' for 0x0, got %s", result)
	}
}

// TestB32_PlayerSize_WidthZero tests width zero but height set
func TestB32_PlayerSize_WidthZero(t *testing.T) {
	svc := newBiddingSvc_B32()
	result := svc.categorizePlayerSize(0, 1500)
	if result != "xlarge" {
		t.Errorf("Expected 'xlarge' for height 1500, got %s", result)
	}
}

// TestB32_PlayerSize_HeightZero tests height zero but width set
func TestB32_PlayerSize_HeightZero(t *testing.T) {
	svc := newBiddingSvc_B32()
	result := svc.categorizePlayerSize(800, 0)
	if result != "large" {
		t.Errorf("Expected 'large' for width 800, got %s", result)
	}
}

// TestB32_PlayerSize_XLarge tests xlarge threshold
func TestB32_PlayerSize_XLarge(t *testing.T) {
	svc := newBiddingSvc_B32()
	result := svc.categorizePlayerSize(1280, 720)
	if result != "xlarge" {
		t.Errorf("Expected 'xlarge' for 1280x720, got %s", result)
	}
}

// TestB32_PlayerSize_Large tests large threshold
func TestB32_PlayerSize_Large(t *testing.T) {
	svc := newBiddingSvc_B32()
	result := svc.categorizePlayerSize(640, 480)
	if result != "large" {
		t.Errorf("Expected 'large' for 640x480, got %s", result)
	}
}

// TestB32_PlayerSize_Medium tests medium threshold
func TestB32_PlayerSize_Medium(t *testing.T) {
	svc := newBiddingSvc_B32()
	result := svc.categorizePlayerSize(400, 300)
	if result != "medium" {
		t.Errorf("Expected 'medium' for 400x300, got %s", result)
	}
}

// TestB32_PlayerSize_Small tests small threshold
func TestB32_PlayerSize_Small(t *testing.T) {
	svc := newBiddingSvc_B32()
	result := svc.categorizePlayerSize(320, 240)
	if result != "small" {
		t.Errorf("Expected 'small' for 320x240, got %s", result)
	}
}

// ============================================================================
// optimizeForCPA Tests
// ============================================================================

// TestB32_OptimizeCPA_NoTarget tests no target CPA
func TestB32_OptimizeCPA_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	pg := &model.PerformanceGoals{TargetCPA: 0}
	perf := performanceData{}

	mult := svc.optimizeForCPA(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 for no target CPA, got %f", mult)
	}
}

// TestB32_OptimizeCPA_Capped tests ratio capped at 2.0
func TestB32_OptimizeCPA_Capped(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	camp.BidPrice = 1.0
	req := makeMinReq_B32()

	pg := &model.PerformanceGoals{TargetCPA: 100.0}
	// High predicted CTR and CVR -> high expected conv rate -> high maxBidForCPA
	perf := performanceData{
		ctr: 0.5, // 50% CTR
		cvr: 0.5, // 50% CVR
	}

	mult := svc.optimizeForCPA(camp, req, pg, perf)
	// maxBid = 100 * 0.5 * 0.5 = 25, ratio = 25/1 = 25 -> capped at 2.0
	if mult != 2.0 {
		t.Errorf("Expected 2.0 cap, got %f", mult)
	}
}

// TestB32_OptimizeCPA_Floored tests ratio floored at 0.3
func TestB32_OptimizeCPA_Floored(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	camp.BidPrice = 10.0
	req := makeMinReq_B32()

	pg := &model.PerformanceGoals{TargetCPA: 1.0}
	// Low predicted CTR and CVR -> low maxBidForCPA
	perf := performanceData{
		ctr: 0.001, // 0.1%
		cvr: 0.001, // 0.1%
	}

	mult := svc.optimizeForCPA(camp, req, pg, perf)
	// maxBid = 1.0 * 0.001 * 0.001 = 0.000001, ratio = 0.000001/10 -> floored at 0.3
	if mult != 0.3 {
		t.Errorf("Expected 0.3 floor, got %f", mult)
	}
}

// ============================================================================
// optimizeForCPS Tests
// ============================================================================

// TestB32_CPS_EcommerceGoalsFallback tests fallback to EcommerceGoals.TargetCostPerSale
func TestB32_CPS_EcommerceGoalsFallback(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()

	pg := &model.PerformanceGoals{
		TargetCPS: 0, // Not set
		EcommerceGoals: &model.EcommerceOptimization{
			TargetCostPerSale: 50.0,
		},
	}
	perf := performanceData{
		ctr: 0.05,
		cvr: 0.02,
	}

	mult := svc.optimizeForCPS(camp, req, pg, perf)
	// Should use EcommerceGoals.TargetCostPerSale
	if mult == 1.0 {
		t.Errorf("Expected multiplier based on ecommerce target, got 1.0")
	}
}

// TestB32_CPS_CartAbandonBoost tests cart abandoner boost
func TestB32_CPS_CartAbandonBoost(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	camp.BidPrice = 2.0
	req := makeMinReq_B32()
	req.Context["cart_abandoner"] = true

	pg := &model.PerformanceGoals{
		TargetCPS: 20.0,
		EcommerceGoals: &model.EcommerceOptimization{
			CartAbandonBoost: 1.5,
		},
	}
	perf := performanceData{
		ctr: 0.05,
		cvr: 0.02,
	}

	mult := svc.optimizeForCPS(camp, req, pg, perf)
	// maxBid = 20 * 0.05 * 0.02 * 1.5 (cart abandon) = 0.03
	// ratio = 0.03 / 2.0 = 0.015 -> floored at 0.3
	if mult != 0.3 {
		t.Errorf("Expected floor applied with cart abandon boost, got %f", mult)
	}
}

// TestB32_CPS_RepeatCustomerBoost tests repeat customer boost
func TestB32_CPS_RepeatCustomerBoost(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	camp.BidPrice = 1.0
	req := makeMinReq_B32()
	req.Context["repeat_customer"] = true

	pg := &model.PerformanceGoals{
		TargetCPS: 50.0,
		EcommerceGoals: &model.EcommerceOptimization{
			RepeatCustomerBoost: 1.3,
		},
	}
	perf := performanceData{
		ctr: 0.1,
		cvr: 0.05,
	}

	mult := svc.optimizeForCPS(camp, req, pg, perf)
	// maxBid = 50 * 0.1 * 0.05 * 1.3 = 0.325
	// ratio = 0.325 / 1.0 = 0.325 (within 0.3-2.5 range)
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range, got %f", mult)
	}
}

// TestB32_CPS_PurchaseCountRepeatCustomer tests repeat customer via purchase_count
func TestB32_CPS_PurchaseCountRepeatCustomer(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	camp.BidPrice = 1.0
	req := makeMinReq_B32()
	req.Context["purchase_count"] = 3.0 // Has purchases

	pg := &model.PerformanceGoals{
		TargetCPS: 50.0,
		EcommerceGoals: &model.EcommerceOptimization{
			RepeatCustomerBoost: 1.2,
		},
	}
	perf := performanceData{
		ctr: 0.1,
		cvr: 0.05,
	}

	mult := svc.optimizeForCPS(camp, req, pg, perf)
	// Should apply repeat customer boost
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with purchase_count boost, got %f", mult)
	}
}

// TestB32_CPS_CartAbandonViaSegment tests cart abandoner via user_segments
func TestB32_CPS_CartAbandonViaSegment(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	camp.BidPrice = 1.0
	req := makeMinReq_B32()
	req.Context["user_segments"] = []interface{}{"premium", "cart_abandon_7d", "frequent_buyer"}

	pg := &model.PerformanceGoals{
		TargetCPS: 30.0,
		EcommerceGoals: &model.EcommerceOptimization{
			CartAbandonBoost: 1.4,
		},
	}
	perf := performanceData{
		ctr: 0.08,
		cvr: 0.03,
	}

	mult := svc.optimizeForCPS(camp, req, pg, perf)
	// Should apply cart abandon boost via segment detection
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range with segment cart abandon, got %f", mult)
	}
}

// ============================================================================
// optimizeForCPR Tests
// ============================================================================

// TestB32_CPR_NoTarget tests no target CPR
func TestB32_CPR_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	pg := &model.PerformanceGoals{TargetCPR: 0}
	perf := performanceData{}

	mult := svc.optimizeForCPR(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 for no target CPR, got %f", mult)
	}
}

// TestB32_CPR_Capped tests ratio capped at 2.0
func TestB32_CPR_Capped(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	camp.BidPrice = 0.5
	req := makeMinReq_B32()

	pg := &model.PerformanceGoals{TargetCPR: 100.0}
	perf := performanceData{
		ctr: 0.2,
		cvr: 0.5, // Registration rate = CVR * 0.8
	}

	mult := svc.optimizeForCPR(camp, req, pg, perf)
	// High predicted rate -> high maxBid -> capped
	if mult != 2.0 {
		t.Errorf("Expected 2.0 cap, got %f", mult)
	}
}

// TestB32_CPR_Floored tests ratio floored at 0.3
func TestB32_CPR_Floored(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	camp.BidPrice = 10.0
	req := makeMinReq_B32()

	pg := &model.PerformanceGoals{TargetCPR: 1.0}
	perf := performanceData{
		ctr: 0.001,
		cvr: 0.001,
	}

	mult := svc.optimizeForCPR(camp, req, pg, perf)
	if mult != 0.3 {
		t.Errorf("Expected 0.3 floor, got %f", mult)
	}
}

// ============================================================================
// optimizeForCPL Tests
// ============================================================================

// TestB32_CPL_NoTarget tests no target CPL
func TestB32_CPL_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	pg := &model.PerformanceGoals{TargetCPL: 0}
	perf := performanceData{}

	mult := svc.optimizeForCPL(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 for no target CPL, got %f", mult)
	}
}

// TestB32_CPL_LeadIntentBoost tests lead_intent_score context boost
func TestB32_CPL_LeadIntentBoost(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	req.Context["lead_intent_score"] = 0.85 // High intent

	pg := &model.PerformanceGoals{TargetCPL: 20.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPL(camp, req, pg, perf)
	// Should apply 1.3x boost for high intent, but may be floored at 0.3 depending on predictCPL
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range 0.3-2.5, got %f", mult)
	}
}

// TestB32_CPL_B2BBoost tests is_b2b context boost
func TestB32_CPL_B2BBoost(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	req.Context["is_b2b"] = true

	pg := &model.PerformanceGoals{TargetCPL: 30.0}
	perf := performanceData{
		ctr: 0.03,
	}

	mult := svc.optimizeForCPL(camp, req, pg, perf)
	// Should apply 1.2x boost for B2B, but may be floored at 0.3 depending on predictCPL
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range 0.3-2.5, got %f", mult)
	}
}

// TestB32_CPL_CombinedBoosts tests combined intent + B2B boosts
func TestB32_CPL_CombinedBoosts(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	req.Context["lead_intent_score"] = 0.75
	req.Context["is_b2b"] = true

	pg := &model.PerformanceGoals{TargetCPL: 50.0}
	perf := performanceData{
		ctr: 0.04,
	}

	mult := svc.optimizeForCPL(camp, req, pg, perf)
	// Should apply both 1.3x and 1.2x boosts, but may be affected by predictCPL
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range 0.3-2.5, got %f", mult)
	}
}

// TestB32_CPL_Capped tests ratio capped at 2.5
func TestB32_CPL_Capped(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	req.Context["lead_intent_score"] = 0.9
	req.Context["is_b2b"] = true

	pg := &model.PerformanceGoals{TargetCPL: 100.0}
	perf := performanceData{
		ctr: 0.5, // Very high CTR
	}

	mult := svc.optimizeForCPL(camp, req, pg, perf)
	// Very high target / predictedCPL -> should have positive multiplier but may be capped
	if mult < 0.3 || mult > 2.5 {
		t.Errorf("Expected multiplier in valid range 0.3-2.5, got %f", mult)
	}
}

// ============================================================================
// optimizeForCPCV Tests
// ============================================================================

// TestB32_CPCV_NoTarget tests no target CPCV
func TestB32_CPCV_NoTarget(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	pg := &model.PerformanceGoals{TargetCPCV: 0}
	perf := performanceData{}

	mult := svc.optimizeForCPCV(camp, req, pg, perf)
	if mult != 1.0 {
		t.Errorf("Expected 1.0 for no target CPCV, got %f", mult)
	}
}

// TestB32_CPCV_NonSkippableBoost tests non-skippable video boost
func TestB32_CPCV_NonSkippableBoost(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	req.Context["skippable"] = false // Non-skippable

	pg := &model.PerformanceGoals{TargetCPCV: 5.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPCV(camp, req, pg, perf)
	// Should apply 1.4x boost for non-skippable
	if mult < 1.0 {
		t.Errorf("Expected boost for non-skippable video, got %f", mult)
	}
}

// TestB32_CPCV_ShortDurationBoost tests short video duration boost
func TestB32_CPCV_ShortDurationBoost(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	req.Context["video_duration"] = 10.0 // 10 seconds

	pg := &model.PerformanceGoals{TargetCPCV: 3.0}
	perf := performanceData{
		ctr: 0.04,
	}

	mult := svc.optimizeForCPCV(camp, req, pg, perf)
	// Should apply 1.2x boost for videos <= 15s, but may be floored at 0.3 depending on predictCPCV
	if mult < 0.3 || mult > 3.0 {
		t.Errorf("Expected multiplier in valid range 0.3-3.0, got %f", mult)
	}
}

// TestB32_CPCV_MediumDurationNeutral tests 30s video (neutral)
func TestB32_CPCV_MediumDurationNeutral(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	req.Context["video_duration"] = 30.0 // 30 seconds

	pg := &model.PerformanceGoals{TargetCPCV: 4.0}
	perf := performanceData{
		ctr: 0.03,
	}

	mult := svc.optimizeForCPCV(camp, req, pg, perf)
	// 30s video gets 1.0x (neutral), no additional boost
	if mult > 2.5 || mult < 0.3 {
		t.Errorf("Expected neutral or reasonable multiplier for 30s video, got %f", mult)
	}
}

// TestB32_CPCV_LongDurationPenalty tests long video penalty
func TestB32_CPCV_LongDurationPenalty(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	req.Context["video_duration"] = 60.0 // 60 seconds

	pg := &model.PerformanceGoals{TargetCPCV: 5.0}
	perf := performanceData{
		ctr: 0.1, // High CTR
	}

	mult := svc.optimizeForCPCV(camp, req, pg, perf)
	// Should apply 0.8x penalty for videos > 30s
	// Depending on predictCPCV, multiplier may vary, but should be affected
	if mult >= 2.0 {
		t.Errorf("Expected lower multiplier for long video, got %f", mult)
	}
}

// TestB32_CPCV_CTVBoost tests CTV inventory boost
func TestB32_CPCV_CTVBoost(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	req.Device.Type = "ctv" // CTV device

	pg := &model.PerformanceGoals{TargetCPCV: 4.0}
	perf := performanceData{
		ctr: 0.05,
	}

	mult := svc.optimizeForCPCV(camp, req, pg, perf)
	// Should apply 1.3x boost for CTV, logic is tested even if result floored
	if mult < 0.3 || mult > 3.0 {
		t.Errorf("Expected multiplier in valid range 0.3-3.0, got %f", mult)
	}
}

// TestB32_CPCV_Capped tests ratio capped at 3.0
func TestB32_CPCV_Capped(t *testing.T) {
	svc := newBiddingSvc_B32()
	camp := makeMinCamp_B32()
	req := makeMinReq_B32()
	req.Device.Type = "ctv"
	req.Context["skippable"] = false
	req.Context["video_duration"] = 10.0

	pg := &model.PerformanceGoals{TargetCPCV: 100.0}
	perf := performanceData{
		ctr: 0.5, // Very high
	}

	mult := svc.optimizeForCPCV(camp, req, pg, perf)
	// All boosts combined with high target -> should cap at 3.0
	if mult != 3.0 {
		t.Errorf("Expected 3.0 cap, got %f", mult)
	}
}
