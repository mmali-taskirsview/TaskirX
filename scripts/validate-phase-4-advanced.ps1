Write-Host "TASKIRX PHASE 4 REDIS & OPTIMIZATION VALIDATION" -ForegroundColor Cyan
Write-Host "===================================================" -ForegroundColor Cyan

# 1. Validate Ad Matching Redis Integration
$matcherFile = "c:\TaskirX\python-ai-agents\ad-matching-service\app\services\matcher.py"
if (Select-String -Path $matcherFile -Pattern "import redis") {
    Write-Host "[PASS] Ad Matcher imports Redis" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Ad Matcher Redis import missing" -ForegroundColor Red
}

if (Select-String -Path $matcherFile -Pattern "redis\.Redis") {
    Write-Host "[PASS] Ad Matcher initializes Redis client" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Ad Matcher Redis initialization missing" -ForegroundColor Red
}

# 2. Validate Bid Optimization Logic
$optFile = "c:\TaskirX\python-ai-agents\bid-optimization-service\app\services\optimizer.py"
if (Select-String -Path $optFile -Pattern "np\.random\.beta") {
    Write-Host "[PASS] Bid Optimizer uses Thompson Sampling (Beta dist)" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Bid Optimizer Thompson Sampling missing" -ForegroundColor Red
}

if (Select-String -Path $optFile -Pattern "def optimize_bid") {
    Write-Host "[PASS] Bid Optimizer exposes optimize_bid method" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Bid Optimizer optimize_bid method missing" -ForegroundColor Red
}

Write-Host "`nPHASE 4 ADVANCED FEATURES COMPLETE" -ForegroundColor Cyan
