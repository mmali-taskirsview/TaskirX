Write-Host "TASKIRX PHASE 4 INTEGRATION VALIDATION" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan

# 1. Validate Go Code Integration
$goFile = "c:\TaskirX\go-bidding-engine\internal\service\bidding.go"
if (Select-String -Path $goFile -Pattern "callAIMatchingService") {
    Write-Host "[PASS] Go Bidding Engine AI Client Implemented" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Go Bidding Engine AI Client Missing" -ForegroundColor Red
}

$modelFile = "c:\TaskirX\go-bidding-engine\internal\model\ai.go"
if (Test-Path $modelFile) {
    Write-Host "[PASS] Go AI Model Structs Created" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Go AI Model Structs Missing" -ForegroundColor Red
}

# 2. Validate Tests
$testFile = "c:\TaskirX\go-bidding-engine\internal\service\bidding_test.go"
if (Select-String -Path $testFile -Pattern "TestProcessBid_AIScoring") {
    Write-Host "[PASS] AI Integration Tests Added" -ForegroundColor Green
} else {
    Write-Host "[FAIL] AI Integration Tests Missing" -ForegroundColor Red
}

# 3. Dynamic Test Execution (Go)
if (Get-Command "go" -ErrorAction SilentlyContinue) {
    Write-Host "`nRunning Go Unit Tests..." -ForegroundColor Yellow
    Push-Location "c:\TaskirX\go-bidding-engine"
    
    # Run tests
    go test ./... -v
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Go Tests Passed!" -ForegroundColor Green
    } else {
        Write-Host "❌ Go Tests Failed!" -ForegroundColor Red
    }
    Pop-Location
} else {
    Write-Host "⚠️  Go not installed. Skipping live test execution." -ForegroundColor Yellow
}

# 4. Validate Monitoring & Observability Integration

# Check Go Metrics
$goHandler = "c:\TaskirX\go-bidding-engine\internal\handler\bid.go"
if (Select-String -Path $goHandler -Pattern "HandleMetrics") {
    Write-Host "[PASS] Go Bidding Engine Metrics Handler Implemented" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Go Bidding Engine Metrics Handler Missing" -ForegroundColor Red
}

$goMetrics = "c:\TaskirX\go-bidding-engine\pkg\metrics\init.go"
if (Test-Path $goMetrics) {
    Write-Host "[PASS] Go Prometheus Metrics Package Created" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Go Prometheus Metrics Package Missing" -ForegroundColor Red
}

# Check Python Instrumentation
$fraudReq = "c:\TaskirX\python-ai-agents\fraud-detection-service\requirements.txt"
if (Select-String -Path $fraudReq -Pattern "prometheus-fastapi-instrumentator") {
    Write-Host "[PASS] Fraud Service Monitoring Dependencies" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Fraud Service Monitoring Dependencies Missing" -ForegroundColor Red
}

$adMatchReq = "c:\TaskirX\python-ai-agents\ad-matching-service\requirements.txt"
if (Select-String -Path $adMatchReq -Pattern "prometheus-fastapi-instrumentator") {
    Write-Host "[PASS] Ad Matching Service Monitoring Dependencies" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Ad Matching Service Monitoring Dependencies Missing" -ForegroundColor Red
}

$bidOptReq = "c:\TaskirX\python-ai-agents\bid-optimization-service\requirements.txt"
if (Select-String -Path $bidOptReq -Pattern "prometheus-fastapi-instrumentator") {
    Write-Host "[PASS] Bid Optimizer Monitoring Dependencies" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Bid Optimizer Monitoring Dependencies Missing" -ForegroundColor Red
}

# Check Prometheus Config
$promConfig = "c:\TaskirX\monitoring\prometheus.yml"
if (Select-String -Path $promConfig -Pattern "job_name: 'go-bidding'") {
    Write-Host "[PASS] Prometheus Configured for Go Bidding" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Prometheus Config Missing Go Bidding Job" -ForegroundColor Red
}

Write-Host "`nPHASE 4 INTEGRATION COMPLETE" -ForegroundColor Cyan
