# PowerShell Integration Test Script for TaskirX Bidding Engine
# Tests server endpoints manually with proper error handling

Write-Host "🚀 Starting TaskirX Integration Tests..." -ForegroundColor Green
Write-Host "=================================================="

$ServerUrl = "http://localhost:8080"
$Results = @()

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Url,
        [string]$Method = "GET",
        [string]$Body = $null
    )
    
    $StartTime = Get-Date
    try {
        $Response = $null
        if ($Method -eq "GET") {
            $Response = Invoke-WebRequest -Uri $Url -Method $Method -TimeoutSec 5
        } else {
            $Headers = @{"Content-Type" = "application/json"}
            $Response = Invoke-WebRequest -Uri $Url -Method $Method -Body $Body -Headers $Headers -TimeoutSec 5
        }
        
        $Duration = ((Get-Date) - $StartTime).TotalMilliseconds
        
        $Result = @{
            Name = $Name
            Status = "PASS"
            StatusCode = $Response.StatusCode
            Duration = [math]::Round($Duration, 2)
            ResponseSize = $Response.Content.Length
        }
        
        Write-Host "✅ $Name - $($Result.StatusCode) ($($Result.Duration)ms)" -ForegroundColor Green
        
        if ($Response.Content.Length -lt 500) {
            Write-Host "   Response: $($Response.Content)" -ForegroundColor Cyan
        } else {
            Write-Host "   Response: $($Response.Content.Substring(0,100))..." -ForegroundColor Cyan
        }
        
    } catch {
        $Duration = ((Get-Date) - $StartTime).TotalMilliseconds
        
        $Result = @{
            Name = $Name
            Status = "FAIL"
            StatusCode = 0
            Duration = [math]::Round($Duration, 2)
            Error = $_.Exception.Message
        }
        
        Write-Host "❌ $Name - FAILED ($($Result.Duration)ms)" -ForegroundColor Red
        Write-Host "   Error: $($Result.Error)" -ForegroundColor Yellow
    }
    
    return $Result
}

# Test 1: Health Endpoint
Write-Host "`n📊 Testing Health Endpoint..."
$Results += Test-Endpoint -Name "Health Check" -Url "$ServerUrl/health"

# Test 2: Metrics Endpoint  
Write-Host "`n📊 Testing Metrics Endpoint..."
$Results += Test-Endpoint -Name "Metrics Endpoint" -Url "$ServerUrl/metrics"

# Test 3: Basic Bid Request
Write-Host "`n📊 Testing Basic Bid Request..."
$BidPayload = @{
    "id" = "integration-test-001"
    "publisher_id" = "pub-integration-001"
    "ad_slot" = @{
        "id" = "slot-001"
        "dimensions" = @(300, 250)
        "formats" = @("banner")
        "position" = "above-fold"
    }
    "user" = @{
        "id" = "user-integration-001"
        "country" = "US"
    }
    "device" = @{
        "type" = "mobile"
        "os" = "iOS"
    }
    "context" = @{
        "site_domain" = "integration-test.com"
    }
} | ConvertTo-Json -Depth 5

$Results += Test-Endpoint -Name "Basic Bid Request" -Url "$ServerUrl/bid" -Method "POST" -Body $BidPayload

# Test 4: Video Bid Request
Write-Host "`n📊 Testing Video Bid Request..."
$VideoBidPayload = @{
    "id" = "video-test-001"
    "publisher_id" = "pub-video-001"
    "ad_slot" = @{
        "id" = "slot-video-001"
        "dimensions" = @(1280, 720)
        "formats" = @("video")
        "position" = "in-stream"
        "video" = @{
            "mimes" = @("video/mp4")
            "minduration" = 5
            "maxduration" = 30
            "protocols" = @(2, 3)
        }
    }
    "user" = @{
        "id" = "user-video-001"
        "country" = "US"
        "interests" = @("sports", "technology")
    }
    "device" = @{
        "type" = "mobile"
        "os" = "iOS"
    }
} | ConvertTo-Json -Depth 5

$Results += Test-Endpoint -Name "Video Bid Request" -Url "$ServerUrl/bid" -Method "POST" -Body $VideoBidPayload

# Test 5: Campaign Refresh
Write-Host "`n📊 Testing Campaign Refresh..."
$Results += Test-Endpoint -Name "Campaign Refresh" -Url "$ServerUrl/campaigns/refresh"

# Test Summary
Write-Host "`n" + "=" * 50 -ForegroundColor Green
Write-Host "📊 Integration Test Summary" -ForegroundColor Green  
Write-Host "=" * 50 -ForegroundColor Green

$PassedTests = ($Results | Where-Object { $_.Status -eq "PASS" }).Count
$FailedTests = ($Results | Where-Object { $_.Status -eq "FAIL" }).Count
$TotalDuration = ($Results | Measure-Object Duration -Sum).Sum

Write-Host "Total Tests: $($Results.Count)" -ForegroundColor White
Write-Host "Passed: $PassedTests" -ForegroundColor Green
Write-Host "Failed: $FailedTests" -ForegroundColor Red
Write-Host "Total Duration: $([math]::Round($TotalDuration, 2))ms" -ForegroundColor White

foreach ($Result in $Results) {
    $Status = if ($Result.Status -eq "PASS") { "✅" } else { "❌" }
    Write-Host "$Status $($Result.Name): $($Result.StatusCode) ($($Result.Duration)ms)"
    
    if ($Result.Error) {
        Write-Host "   Error: $($Result.Error)" -ForegroundColor Yellow
    }
}

if ($FailedTests -eq 0) {
    Write-Host "`n🎉 All integration tests passed!" -ForegroundColor Green
    Write-Host "Server is ready for production deployment." -ForegroundColor Green
} else {
    Write-Host "`n⚠️ Some tests failed. Check server status." -ForegroundColor Yellow
}

Write-Host "`n💡 Next Steps:" -ForegroundColor Cyan
Write-Host "• Review any failed tests"
Write-Host "• Check server logs for errors"
Write-Host "• Validate campaign data and targeting"
Write-Host "• Monitor performance under load"
Write-Host "• Proceed to production deployment"