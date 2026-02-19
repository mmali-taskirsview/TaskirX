# Test Fraud Detection Integration
# Verifies that the Fraud Service is responding and blocking mock IPs correctly.

$FraudServiceUrl = "http://localhost:6001/api/detect"
$Headers = @{ "Content-Type" = "application/json" }


Write-Host "Testing Fraud Detection Service Integration..." -ForegroundColor Cyan

# 1. Test Clean IP
$CleanPayload = @{
    request_id = "test-clean-1"
    timestamp = [DateTime]::UtcNow.ToString("yyyy-MM-ddTHH:mm:ssZ")
    event_type = "bid_request"
    ip_address = "8.8.8.8"
    campaign_id = "camp-001"

    publisher_id = "pub-001"
    advertiser_id = "adv-001"
    device = @{
        type = "mobile"
        os = "Android"
        user_agent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
    }
    geo = @{
        country = "US"
        city = "New York"
    }
} | ConvertTo-Json -Depth 5

try {
    Write-Host "1. Testing Clean IP (8.8.8.8)... " -NoNewline
    $Response = Invoke-RestMethod -Uri $FraudServiceUrl -Method Post -Body $CleanPayload -Headers $Headers
    if ($Response.risk_level -eq "LOW") {
        Write-Host "PASS (Risk: $($Response.risk_level))" -ForegroundColor Green
    } else {
        Write-Host "FAIL (Risk: $($Response.risk_level))" -ForegroundColor Red
        $Response | Format-List
    }
} catch {
    Write-Host "ERROR: $_" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader $_.Exception.Response.GetResponseStream()
        $reader.ReadToEnd()
    }
}

# 2. Test Mock Blocked IP (.99)
$BlockedPayload = @{
    request_id = "test-blocked-1"
    timestamp = [DateTime]::UtcNow.ToString("yyyy-MM-ddTHH:mm:ssZ")
    event_type = "bid_request"
    ip_address = "192.168.1.99"

    campaign_id = "camp-001"
    publisher_id = "pub-001"
    advertiser_id = "adv-001"
    device = @{
        type = "desktop"
        os = "Windows"
        user_agent = "Mozilla/5.0"
    }
    geo = @{
        country = "CN"
        city = "Beijing"
    }
} | ConvertTo-Json -Depth 5


try {
    Write-Host "2. Testing Mock Blocked IP (*.99)... " -NoNewline
    $Response = Invoke-RestMethod -Uri $FraudServiceUrl -Method Post -Body $BlockedPayload -Headers $Headers
    # Expecting HIGH or CRITICAL risk
    if ($Response.risk_level -eq "CRITICAL" -or $Response.recommended_action -eq "block") {
        Write-Host "PASS (Action: $($Response.recommended_action))" -ForegroundColor Green
    } else {
        Write-Host "FAIL (Action: $($Response.recommended_action))" -ForegroundColor Red
        $Response | Format-List
    }
} catch {
    Write-Host "ERROR: $_" -ForegroundColor Red
}

Write-Host "`nTo enable real AbuseIPDB checking:"
Write-Host "1. Get an API key from https://www.abuseipdb.com/"
Write-Host "2. Add IP_REPUTATION_API_KEY=your_key to .env"
Write-Host "3. Restart services: docker-compose up -d --build fraud-detection"
