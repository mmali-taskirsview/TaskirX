# Quick Integration Validation Script
Write-Host "🚀 TaskirX Server Quick Test" -ForegroundColor Green

# Test server is running check
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -Method GET -TimeoutSec 3
    Write-Host "✅ Health Endpoint: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Response: $($response.Content)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Server not responding - $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Make sure test server is running:" -ForegroundColor Yellow
    Write-Host "cd c:\TaskirX\go-bidding-engine\cmd\test-server" -ForegroundColor Yellow
    Write-Host "go run main.go" -ForegroundColor Yellow
    exit 1
}

Write-Host "`n✅ Integration test server is ready!" -ForegroundColor Green