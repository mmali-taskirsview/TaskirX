Write-Host "TASKIRX PHASE 4 START VALIDATION" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan

# 1. Validate Dashboard WebSocket Client
$rtbMonitor = "c:\TaskirX\frontend\src\pages\RTBMonitor.jsx"
if (Select-String -Path $rtbMonitor -Pattern "socket.io-client") {
    Write-Host "[PASS] RTBMonitor.jsx uses socket.io-client" -ForegroundColor Green
} else {
    Write-Host "[FAIL] RTBMonitor.jsx does not import socket.io-client" -ForegroundColor Red
}

# 2. Validate Ad Matching Service Code
$matcherFile = "c:\TaskirX\python-ai-agents\ad-matching-service\app\services\matcher.py"
if (Select-String -Path $matcherFile -Pattern "_calculate_content_score") {
    Write-Host "[PASS] Ad Matcher Content Scoring Implemented" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Ad Matcher Content Scoring Missing" -ForegroundColor Red
}

if (Select-String -Path $matcherFile -Pattern "_calculate_hybrid_score") {
    Write-Host "[PASS] Ad Matcher Hybrid Logic Implemented" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Ad Matcher Hybrid Logic Missing" -ForegroundColor Red
}

# 3. Validate Docker Compose
$dockerCompose = "c:\TaskirX\docker-compose.yml"
if (Select-String -Path $dockerCompose -Pattern "ad-matching-service") {
    Write-Host "[PASS] Ad Matching Service added to Docker Compose" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Docker Compose missing ad-matching-service" -ForegroundColor Red
}

Write-Host "`nPHASE 4 INITIALIZATION COMPLETE" -ForegroundColor Cyan
