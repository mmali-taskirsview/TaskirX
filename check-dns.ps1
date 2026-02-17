$domains = @("taskirx.com", "www.taskirx.com", "api.taskirx.com", "dashboard.taskirx.com", "bidding.taskirx.com")
$targetIP = "138.2.76.159"

Write-Host "Checking DNS propagation for TaskirX..." -ForegroundColor Cyan

foreach ($domain in $domains) {
    try {
        $ip = [System.Net.Dns]::GetHostAddresses($domain) | Where-Object { $_.AddressFamily -eq 'InterNetwork' } | Select-Object -First 1
        if ($ip) {
            Write-Host "[OK] $domain resolves to $($ip.IPAddressToString)" -ForegroundColor Green
        } else {
            Write-Host "[PENDING] $domain not resolving yet" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "[ERROR] Could not resolve $domain" -ForegroundColor Red
    }
}
