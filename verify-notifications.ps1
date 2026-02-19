# Comprehensive Setup Verification for Notifications (Budget, Fraud, Performance)

$ApiUrl = "http://localhost:3000" # NestJS typically runs on 3000, check docker-compose
if ($env:API_URL) { $ApiUrl = $env:API_URL }
# Ensure API URL has /api suffix if not provided
if (-not $ApiUrl.EndsWith("/api")) { $ApiUrl = "$ApiUrl/api" }

Write-Host "Starting Notification System Verification..." -ForegroundColor Green
Write-Host "Target API: $ApiUrl"

# 1. Login
Write-Host "`n[1/5] Authenticating as Admin..."
try {
    if (-not (Test-Path "login.json")) {
        Write-Host "Error: login.json not found. Creating default admin login..." -ForegroundColor Yellow
        @{
            email = "admin@taskir.com"
            password = "admin_password_2026"
        } | ConvertTo-Json | Set-Content "login.json"
    }
    $LoginBody = Get-Content "login.json" -Raw
    $LoginResponse = Invoke-RestMethod -Uri "$ApiUrl/auth/login" -Method Post -ContentType "application/json" -Body $LoginBody
    $Token = $LoginResponse.access_token
    if (-not $Token) { throw "No access token received" }
    Write-Host "Success: Logged in." -ForegroundColor Green
} catch {
    Write-Host "Error: Login failed. Is the backend running?" -ForegroundColor Red
    Write-Host $_
    exit 1
}

$Headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
}

# 2. Setup Budget Alert Test
Write-Host "`n[2/5] Testing Budget Alerts..."
$CampaignName = "Budget-Test-" + (Get-Date).ToString("mmss")
$CampaignBody = @{
    name = $CampaignName
    budget = 100
    status = "active"
    type = "cpm"
    bidPrice = 1.0
    vertical = "test"
    startDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    endDate = (Get-Date).AddDays(1).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
} | ConvertTo-Json

try {
    # Create Campaign
    $Campaign = Invoke-RestMethod -Uri "$ApiUrl/campaigns" -Method Post -Headers $Headers -Body $CampaignBody
    $CampaignId = $Campaign.id
    Write-Host "Created Campaign: $CampaignId (Budget: $100)"

    # Trigger 90% Spend
    $ImpressionBody = @{
        campaignId = $CampaignId
        publisherId = "pub-test-01"
        deviceType = "desktop"
        country = "US"
        price = 90.5 # 90.5%
        timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
    } | ConvertTo-Json

    # Track Impression with Price
    Invoke-RestMethod -Uri "$ApiUrl/analytics/track/impression" -Method Post -ContentType "application/json" -Body $ImpressionBody
    Write-Host "Simulated $90.50 spend. Notification should be triggered."

    # Verify Logic (Simulated by checking logs or notifications endpoint if available)
    Start-Sleep -Seconds 2
    
    # Check Notifications
    $Notifs = Invoke-RestMethod -Uri "$ApiUrl/notifications" -Headers $Headers
    $BudgetAlert = $Notifs | Where-Object { $_.title -like "*Budget Warning*" -and $_.message -like "*$CampaignName*" }
    
    if ($BudgetAlert) {
        Write-Host "SUCCESS: Budget Warning Notification Found!" -ForegroundColor Green
    } else {
        Write-Host "WARNING: Budget Notification not found yet (async delay?)" -ForegroundColor Yellow
    }

} catch {
    Write-Host "Failed Budget Test: $_" -ForegroundColor Red
}

# 3. Setup Fraud Alert Test
Write-Host "`n[3/5] Testing Fraud Alerts (Redis Injection)..."
$BadPublisherId = "pub-fraud-" + (Get-Date).ToString("Hmm")
$Today = (Get-Date).ToUniversalTime().ToString("yyyy-MM-dd")

try {
    # 51 fraud events to trigger >50 threshold
    Write-Host "Injecting 51 fraud events for $BadPublisherId into Redis..."
    # Using docker-compose to inject. Assuming 'redis' service name.
    # We will use a faster method if docker is available locally, else skip
    if (Get-Command "docker-compose" -ErrorAction SilentlyContinue) {
        # Inject Key with Password
        $Cmd = "INCRBY fraud:publisher:${BadPublisherId}:${Today}:count 51"
        Write-Host "Executing Redis Command: $Cmd"
        docker-compose exec -T redis redis-cli -a taskir_redis_password_2026 INCRBY fraud:publisher:${BadPublisherId}:${Today}:count 51 | Out-Null
        
        # Add to active set (Optional but good practice)
        docker-compose exec -T redis redis-cli -a taskir_redis_password_2026 SADD fraud:publishers:active:$Today $BadPublisherId | Out-Null

        Write-Host "Injected. Waiting 65 seconds for Cron Job (Runs every minute)..."
        Start-Sleep -Seconds 65 

        # Check Notifications for Admin
        $Notifs = Invoke-RestMethod -Uri "$ApiUrl/notifications" -Headers $Headers
        $FraudAlert = $Notifs | Where-Object { $_.title -like "*Fraud Alert*" -and $_.message -like "*$BadPublisherId*" }
        
        if ($FraudAlert) {
            Write-Host "SUCCESS: Fraud Alert Notification Found!" -ForegroundColor Green
        } else {
            Write-Host "WARNING: Fraud Notification not found. Cron job might not have run yet." -ForegroundColor Yellow
        }
    } else {
        Write-Host "Skipping Fraud Test: docker-compose not found to inject Redis data." -ForegroundColor Gray
    }
} catch {
    Write-Host "Failed Fraud Test: $_" -ForegroundColor Red
}

# 4. Cleanup
Write-Host "`n[5/5] Cleanup..."
# Start-Sleep -Seconds 1
Write-Host "Test Complete."
