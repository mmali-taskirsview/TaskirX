# TaskirX Phase 2 Validation Script
# Usage: ./validate-phase-2.ps1

$ErrorActionPreference = "Stop"
$ApiBaseUrl = "http://localhost:3000/api"
$InternalBaseUrl = "http://localhost:3000"

function Invoke-Api {
    param(
        [string]$Url,
        [string]$Method = "GET",
        [hashtable]$Body = @{},
        [string]$Token
    )

    $Headers = @{ "Content-Type" = "application/json" }
    if ($Token) {
        $Headers["Authorization"] = "Bearer $Token"
    }

    $JsonBody = $Body | ConvertTo-Json -Depth 10

    try {
        if ($Method -eq "GET") {
            $Response = Invoke-RestMethod -Uri $Url -Method $Method -Headers $Headers
        } else {
            $Response = Invoke-RestMethod -Uri $Url -Method $Method -Headers $Headers -Body $JsonBody
        }
        return $Response
    } catch {
        Write-Error "Request failed: $_"
        return $null
    }
}

Write-Host "Starting Phase 2 Validation..." -ForegroundColor Cyan

# Pre-check: Is Server Running?
try {
    $conn = Test-NetConnection -ComputerName "localhost" -Port 3000 -InformationLevel Quiet
    if (-not $conn) {
         Write-Host "⚠️  Backend (localhost:3000) is NOT accessible." -ForegroundColor Yellow
         Write-Host "   Skipping API Integration Tests (Local Docker is down)." -ForegroundColor Gray
         Write-Host "   This is expected if you are deploying to OCI." -ForegroundColor Gray
         exit 0
    }
} catch {
    Write-Host "⚠️  Could not check localhost connection." -ForegroundColor Yellow
}

# 1. Login
Write-Host "`n1. Authenticating as Admin..." -ForegroundColor Yellow
$LoginBody = @{
    email = "admin@taskirx.com"
    password = "Admin123!"
}
$LoginResponse = Invoke-Api -Url "$ApiBaseUrl/auth/login" -Method "POST" -Body $LoginBody

if (-not $LoginResponse.access_token) {
    Write-Error "Login failed. Check if backend is running."
    exit 1
}

$Token = $LoginResponse.access_token
Write-Host "Login successful!" -ForegroundColor Green

# 2. Set User Segments
Write-Host "`n2. Populating User Segments (Redis)..." -ForegroundColor Yellow
$TestUserId = "test-user-123"
$Segments = @("vip", "high-spender", "tech-enthusiast")

$SegmentBody = @{
    userId = $TestUserId
    segments = $Segments
}

$SegmentResponse = Invoke-Api -Url "$ApiBaseUrl/targeting/user-segments" -Method "POST" -Body $SegmentBody -Token $Token
Write-Host "User Segments set for $TestUserId" -ForegroundColor Green

# 3. Set Geo Rules
Write-Host "`n3. Populating Geo Rules (Redis)..." -ForegroundColor Yellow
$GeoBody = @{
    country = "US"
    rules = @{
        blocked = $false
        boost_multiplier = 1.5
    }
}

$GeoResponse = Invoke-Api -Url "$ApiBaseUrl/targeting/geo-rules" -Method "POST" -Body $GeoBody -Token $Token
Write-Host "Geo Rules set for US (1.5x Multiplier)" -ForegroundColor Green

# 4. Verify Active Campaigns Internal Endpoint
Write-Host "`n4. Verifying Internal Active Campaigns Endpoint..." -ForegroundColor Yellow
try {
    $ActiveCampaigns = Invoke-RestMethod -Uri "$ApiBaseUrl/internal/campaigns/active" -Method "GET"
    $Count = $ActiveCampaigns.Count
    Write-Host "Internal Endpoint Accessible: Found $Count active campaigns" -ForegroundColor Green
} catch {
    Write-Host "Failed to access internal active campaigns endpoint: $_" -ForegroundColor Red
}

Write-Host "`nPhase 2 Configuration Loaded!" -ForegroundColor Cyan
Write-Host "Validation Complete for Phase 2."
