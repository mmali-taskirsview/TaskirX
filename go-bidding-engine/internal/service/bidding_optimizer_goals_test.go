package service

import (
"testing"

"github.com/taskirx/go-bidding-engine/internal/model"
)

func TestOptimizeForViewability(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForViewability(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{ViewabilityGoal: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForViewability() = %v, expected 1.0", result)
}
}

func TestOptimizeForCompletion(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForCompletion(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{CompletionGoal: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForCompletion() = %v, expected 1.0", result)
}
}

func TestOptimizeForEngagement(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForEngagement(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{EngagementGoal: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForEngagement() = %v, expected 1.0", result)
}
}

func TestOptimizeForCPI(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForCPI(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{TargetCPI: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForCPI() = %v, expected 1.0", result)
}
}

func TestOptimizeForCPS(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForCPS(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{TargetCPS: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForCPS() = %v, expected 1.0", result)
}
}

func TestOptimizeForCPR(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForCPR(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{TargetCPR: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForCPR() = %v, expected 1.0", result)
}
}

func TestOptimizeForCPL(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForCPL(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{TargetCPL: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForCPL() = %v, expected 1.0", result)
}
}

func TestOptimizeForCPV(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForCPV(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{TargetCPV: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForCPV() = %v, expected 1.0", result)
}
}

func TestOptimizeForCPCV(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForCPCV(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{TargetCPCV: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForCPCV() = %v, expected 1.0", result)
}
}

func TestOptimizeForCPE(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForCPE(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{TargetCPE: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForCPE() = %v, expected 1.0", result)
}
}

func TestOptimizeForVCPM(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForVCPM(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{TargetVCPM: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForVCPM() = %v, expected 1.0", result)
}
}

func TestOptimizeForDCPM(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForDCPM(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{TargetDCPM: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForDCPM() = %v, expected 1.0", result)
}
}

func TestOptimizeForCPAD(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForCPAD(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{TargetCPAD: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForCPAD() = %v, expected 1.0", result)
}
}

func TestOptimizeForCPIAAP(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForCPIAAP(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{TargetCPIAAP: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForCPIAAP() = %v, expected 1.0", result)
}
}

func TestOptimizeForCTV(t *testing.T) {
s := createBiddingUtilsService()
// Non-CTV inventory returns 0.5 penalty
result := s.optimizeForCTV(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{CTVGoals: nil}, performanceData{})
if result < 0.49 || result > 0.51 {
t.Errorf("optimizeForCTV() with non-CTV = %v, expected 0.5", result)
}
// CTV inventory gets boost
result = s.optimizeForCTV(&model.Campaign{ID: "camp1"}, &model.BidRequest{Device: model.InternalDevice{Type: "ctv"}}, &model.PerformanceGoals{CTVGoals: nil}, performanceData{})
if result < 1.2 {
t.Errorf("optimizeForCTV() with CTV = %v, expected >= 1.3", result)
}
}

func TestOptimizeForROAS(t *testing.T) {
s := createBiddingUtilsService()
result := s.optimizeForROAS(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, &model.PerformanceGoals{TargetROAS: 0}, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("optimizeForROAS() = %v, expected 1.0", result)
}
}

func TestPredictInstallRate(t *testing.T) {
s := createBiddingUtilsService()
result := s.predictInstallRate(&model.Campaign{ID: "camp1"}, &model.BidRequest{Device: model.InternalDevice{Type: "mobile"}}, performanceData{})
if result < 0.01 || result > 0.5 {
t.Errorf("predictInstallRate() = %v, expected between 0.01 and 0.5", result)
}
}

func TestPredictROAS(t *testing.T) {
s := createBiddingUtilsService()
result := s.predictROAS(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, performanceData{})
if result < 0.5 || result > 10.0 {
t.Errorf("predictROAS() = %v, expected between 0.5 and 10.0", result)
}
}

func TestPredictLTV(t *testing.T) {
s := createBiddingUtilsService()
result := s.predictLTV(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, performanceData{})
if result < 1.0 || result > 100.0 {
t.Errorf("predictLTV() = %v, expected between 1.0 and 100.0", result)
}
}

func TestPredictCPL(t *testing.T) {
s := createBiddingUtilsService()
result := s.predictCPL(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, performanceData{})
if result != 0 {
t.Errorf("predictCPL() = %v, expected 0 for missing data", result)
}
}

func TestPredictCPV(t *testing.T) {
s := createBiddingUtilsService()
result := s.predictCPV(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, performanceData{})
if result != 0 {
t.Errorf("predictCPV() = %v, expected 0 for missing data", result)
}
}

func TestPredictCPCV(t *testing.T) {
s := createBiddingUtilsService()
result := s.predictCPCV(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, performanceData{})
if result != 0 {
t.Errorf("predictCPCV() = %v, expected 0 for missing data", result)
}
}

func TestPredictCPE(t *testing.T) {
s := createBiddingUtilsService()
result := s.predictCPE(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, performanceData{})
if result != 0 {
t.Errorf("predictCPE() = %v, expected 0 for missing data", result)
}
}

func TestIsCTVInventory(t *testing.T) {
s := createBiddingUtilsService()

if !s.isCTVInventory(&model.BidRequest{Device: model.InternalDevice{Type: "ctv"}}) {
t.Error("isCTVInventory() should return true for ctv device")
}
if !s.isCTVInventory(&model.BidRequest{Device: model.InternalDevice{Type: "tv"}}) {
t.Error("isCTVInventory() should return true for tv device")
}
if !s.isCTVInventory(&model.BidRequest{Device: model.InternalDevice{Type: "connected_tv"}}) {
t.Error("isCTVInventory() should return true for connected_tv device")
}
if !s.isCTVInventory(&model.BidRequest{Device: model.InternalDevice{Type: "desktop"}, Context: map[string]interface{}{"environment": "ott"}}) {
t.Error("isCTVInventory() should return true for ott environment")
}
if s.isCTVInventory(&model.BidRequest{Device: model.InternalDevice{Type: "mobile"}}) {
t.Error("isCTVInventory() should return false for mobile device")
}
}

func TestGetHouseholdID(t *testing.T) {
s := createBiddingUtilsService()

result := s.getHouseholdID(&model.BidRequest{Context: map[string]interface{}{"household_id": "hh123"}})
if result != "hh123" {
t.Errorf("getHouseholdID() = %v, expected hh123", result)
}
result = s.getHouseholdID(&model.BidRequest{Context: map[string]interface{}{}})
if result != "" {
t.Errorf("getHouseholdID() = %v, expected empty", result)
}
}

func TestCheckPerformanceThresholds(t *testing.T) {
s := createBiddingUtilsService()

passed, _ := s.checkPerformanceThresholds(&model.PerformanceGoals{}, &model.PerformanceGoalResult{}, performanceData{})
if passed {
t.Error("checkPerformanceThresholds() should return false when no thresholds set")
}
pg := &model.PerformanceGoals{Thresholds: &model.PerformanceThresholds{MinCTR: 0.05}}
result := &model.PerformanceGoalResult{PredictedCTR: 0.01}
passed, reason := s.checkPerformanceThresholds(pg, result, performanceData{})
if !passed {
t.Error("checkPerformanceThresholds() should return true when CTR below min")
}
if reason != "predicted_ctr_below_threshold" {
t.Errorf("checkPerformanceThresholds() reason = %v, expected predicted_ctr_below_threshold", reason)
}
}

func TestApplyCTVOptimizations(t *testing.T) {
s := createBiddingUtilsService()

// nil CTV goals
result := s.applyCTVOptimizations(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, nil, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("applyCTVOptimizations() = %v, expected 1.0", result)
}
// with live content boost (is_live: true in context)
result = s.applyCTVOptimizations(&model.Campaign{ID: "camp1"}, &model.BidRequest{Context: map[string]interface{}{"is_live": true}}, &model.CTVOptimization{LiveContentBoost: 1.5}, performanceData{})
if result < 1.4 || result > 1.6 {
t.Errorf("applyCTVOptimizations() with live content = %v, expected 1.5", result)
}
}

func TestApplyAppOptimizations(t *testing.T) {
s := createBiddingUtilsService()

result := s.applyAppOptimizations(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, nil, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("applyAppOptimizations() = %v, expected 1.0", result)
}
}

func TestApplyEcommerceOptimizations(t *testing.T) {
s := createBiddingUtilsService()

result := s.applyEcommerceOptimizations(&model.Campaign{ID: "camp1"}, &model.BidRequest{}, nil, performanceData{})
if result < 0.99 || result > 1.01 {
t.Errorf("applyEcommerceOptimizations() = %v, expected 1.0", result)
}
result = s.applyEcommerceOptimizations(&model.Campaign{ID: "camp1"}, &model.BidRequest{Context: map[string]interface{}{"cart_abandoner": true}}, &model.EcommerceOptimization{CartAbandonBoost: 1.5}, performanceData{})
if result < 1.4 || result > 1.6 {
t.Errorf("applyEcommerceOptimizations() with cart abandoner = %v, expected 1.5", result)
}
}
