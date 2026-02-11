Write-Host "TASKIRX PHASE 3 VALIDATION" -ForegroundColor Cyan
Write-Host "============================" -ForegroundColor Cyan

# 1. Validate Real-Time Dashboard Backend
$gatewayFile = "c:\TaskirX\nestjs-backend\src\gateway\metrics.gateway.ts"
if (Test-Path $gatewayFile) {
    Write-Host "[PASS] Real-Time Dashboard (Gateway) exists" -ForegroundColor Green
} else {
    Write-Host "[FAIL] metrics.gateway.ts not found" -ForegroundColor Red
}

# 2. Validate Go Bidding Engine Resilience Tests
Write-Host "`nRunning Go Bidding Engine Tests..." -ForegroundColor Yellow
Set-Location "c:\TaskirX\go-bidding-engine"
$testResult = go test -v ./internal/service/... 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "[PASS] Resilience Tests Passed" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Go Tests Failed" -ForegroundColor Red
    $testResult | Select-Object -Last 10
}

# 3. Validate Header Bidding Prototype
$demoFile = "c:\TaskirX\frontend\src\pages\PublisherDemo.jsx"
if (Test-Path $demoFile) {
    Write-Host "[PASS] Header Bidding Prototype (PublisherDemo.jsx) exists" -ForegroundColor Green
} else {
    Write-Host "[FAIL] PublisherDemo.jsx not found" -ForegroundColor Red
}

Write-Host "`nPHASE 3 VALIDATION COMPLETE" -ForegroundColor Cyan
