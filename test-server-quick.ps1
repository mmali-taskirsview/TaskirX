# Quick Test Script for Bidding Server
param(
    [string]$Port = "8080"
)

Write-Host "Testing Bidding Server on port $Port..." -ForegroundColor Cyan

# Test 1: Health Check
Write-Host "`n=== Test 1: Health Check ===" -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:$Port/health" -Method Get -TimeoutSec 5
    Write-Host "✓ Health check passed" -ForegroundColor Green
    Write-Host ($health | ConvertTo-Json)
} catch {
    Write-Host "✗ Health check failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 2: Simple Bid Request
Write-Host "`n=== Test 2: Bid Request ===" -ForegroundColor Yellow
try {
    $bidPayload = Get-Content "test-bid-payload.json" | ConvertFrom-Json
    $bid = Invoke-RestMethod -Uri "http://localhost:$Port/bid" -Method Post -Body ($bidPayload | ConvertTo-Json) -ContentType "application/json" -TimeoutSec 5
    Write-Host "✓ Bid request succeeded" -ForegroundColor Green
    Write-Host ($bid | ConvertTo-Json -Depth 3)
} catch {
    Write-Host "✗ Bid request failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "Error details: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
}

Write-Host "`n=== Tests Complete ===" -ForegroundColor Cyan
