# Test Budget Alerts Integration

$ApiUrl = "http://localhost:4000"
$AnalyticsUrl = "http://localhost:4000" # Backend port

# 1. Login
Write-Host "Logging in..."
try {
    $LoginBody = Get-Content "login.json" -Raw
    $LoginResponse = Invoke-RestMethod -Uri "$ApiUrl/api/auth/login" -Method Post -ContentType "application/json" -Body $LoginBody
    $Token = $LoginResponse.access_token
    Write-Host "Logged in successfully."
} catch {
    Write-Host "Login failed. Ensure backend is running on port 4000." -ForegroundColor Red
    exit 1
}

$Headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
}

# 2. Create Campaign with small budget ($100)
$CampaignDate = Get-Date -Format "yyyyMMddHHmmss"
$CampaignName = "Budget Alert Test $CampaignDate"
$CampaignBody = @{
    name = $CampaignName
    description = "Testing budget alerts"
    status = "active"
    type = "cpm"
    budget = 100
    bidPrice = 1.0
    startDate = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
    endDate = (Get-Date).AddDays(7).ToString("yyyy-MM-ddTHH:mm:ssZ")
} | ConvertTo-Json

Write-Host "Creating campaign..."
try {
    $Campaign = Invoke-RestMethod -Uri "$ApiUrl/api/campaigns" -Method Post -Headers $Headers -Body $CampaignBody
    $CampaignId = $Campaign.id
    Write-Host "Campaign created: $CampaignId (Budget: $100)"
} catch {
    Write-Host "Failed to create campaign: $_" -ForegroundColor Red
    exit 1
}

# 3. Simulate 90% Spend ($90)
Write-Host "Simulating 90% spend ($90)..."
$TrackingUrl = "$AnalyticsUrl/api/analytics/track/impression"
try {
    # Send 9 impressions of $10 each (assuming high price for test speed)
    for ($i = 1; $i -le 9; $i++) {
        $Params = "?campaignId=$CampaignId&publisherId=test-pub&deviceType=desktop&country=US&price=10.00"
        Invoke-RestMethod -Uri "$TrackingUrl$Params" -Method Get
        Write-Host "  - Impression $i sent ($10)"
    }
} catch {
    Write-Host "Failed to track impression: $_" -ForegroundColor Red
}

# 4. Check Notifications for Warning
Start-Sleep -Seconds 2 # Give async heavy lifting a moment
Write-Host "Checking for budget warning..."
try {
    $Notifications = Invoke-RestMethod -Uri "$ApiUrl/api/notifications" -Method Get -Headers $Headers
    $Warning = $Notifications | Where-Object { $_.campaignId -eq $CampaignId -and $_.type -eq 'warning' }
    
    # Note: Notification entity might not store campaignId explicitly in list unless title/message matches
    $Warning = $Notifications | Where-Object { $_.title -eq 'Budget Warning' -and $_.message -match $CampaignName }

    if ($Warning) {
        Write-Host "SUCCESS: Found warning notification!" -ForegroundColor Green
        Write-Host "  Title: $($Warning.title)"
        Write-Host "  Message: $($Warning.message)"
    } else {
        Write-Host "WARNING: No budget warning found yet." -ForegroundColor Yellow
    }
} catch {
    Write-Host "Failed to fetch notifications: $_" -ForegroundColor Red
}

# 5. Simulate 100% Spend (Add $10 more -> $100 total)
Write-Host "Simulating 100% spend (Total $100)..."
try {
    $Params = "?campaignId=$CampaignId&publisherId=test-pub&deviceType=desktop&country=US&price=10.00"
    Invoke-RestMethod -Uri "$TrackingUrl$Params" -Method Get
    Write-Host "  - Impression 10 sent ($10)"
} catch {
    Write-Host "Failed to track impression: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# 6. Check Notifications for Alert
Write-Host "Checking for budget exhausted alert..."
try {
    $Notifications = Invoke-RestMethod -Uri "$ApiUrl/api/notifications" -Method Get -Headers $Headers
    $Alert = $Notifications | Where-Object { $_.title -eq 'Budget Exhausted' -and $_.message -match $CampaignName }

    if ($Alert) {
        Write-Host "SUCCESS: Found exhausted alert!" -ForegroundColor Green
        Write-Host "  Title: $($Alert.title)"
        Write-Host "  Message: $($Alert.message)"
    } else {
        Write-Host "WARNING: No exhausted alert found yet." -ForegroundColor Yellow
    }
} catch {
    Write-Host "Failed to fetch notifications: $_" -ForegroundColor Red
}

Write-Host "Test complete."
