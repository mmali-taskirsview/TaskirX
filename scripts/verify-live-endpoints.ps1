# TaskirX Live System Verifier
# Checks the health of the live OCI endpoints

param (
    [string]$Domain = "taskir.com"
)

$Endpoints = @(
    @{ Name = "API Backend"; Url = "https://api.$Domain/health"; Method = "GET" },
    @{ Name = "Dashboard";   Url = "https://dashboard.$Domain"; Method = "GET" },
    @{ Name = "Bidding";     Url = "https://bidding.$Domain/health"; Method = "GET" }
)

Write-Host "Verifying TaskirX Live Systems ($Domain)..." -ForegroundColor Cyan
Write-Host "========================================`n"

foreach ($ep in $Endpoints) {
    Write-Host -NoNewline "Checking $($ep.Name)... "
    try {
        $response = Invoke-WebRequest -Uri $ep.Url -Method $ep.Method -ErrorAction Stop -TimeoutSec 5
        
        if ($response.StatusCode -eq 200) {
            Write-Host "OPEN (200 OK)" -ForegroundColor Green
            # Check content for health status if JSON
            if ($ep.Url -like "*health*") {
                Write-Host "   > Payload: $($response.Content)" -ForegroundColor Gray
            }
        } else {
            Write-Host "Status: $($response.StatusCode)" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "DOWN" -ForegroundColor Red
        Write-Host "   > Error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n========================================"
Write-Host "Note: If Cloudflare is active, 520-526 errors mean Backend is unreachable."
Write-Host "      If Timeout, check Load Balancer Security Groups."
