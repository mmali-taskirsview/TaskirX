# TaskirX Live System Verifier
# Checks the health of the live OCI endpoints

param (
    [string]$Domain = "taskir.com"
)

# Ensure modern TLS for HTTPS checks (PowerShell 5.1 defaults can be outdated)
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$Endpoints = @(
    @{ Name = "API Backend"; Url = "https://api.$Domain/health"; Method = "GET" },
    @{ Name = "Dashboard";   Url = "https://dashboard.$Domain"; Method = "GET" },
    @{ Name = "Bidding";     Url = "https://bidding.$Domain/health"; Method = "GET" }
)

Write-Host "Verifying TaskirX Live Systems ($Domain)..." -ForegroundColor Cyan
Write-Host "========================================`n"

function Test-TlsHandshake {
    param (
        [string]$HostName,
        [System.Security.Authentication.SslProtocols]$Protocol
    )

    try {
        $tcpClient = New-Object System.Net.Sockets.TcpClient($HostName, 443)
        $sslStream = New-Object System.Net.Security.SslStream($tcpClient.GetStream(), $false, ({ $true }))
        $sslStream.AuthenticateAsClient($HostName, $null, $Protocol, $false)
        $sslStream.Dispose()
        $tcpClient.Close()
        return $true
    } catch {
        return $false
    }
}

foreach ($ep in $Endpoints) {
    $hostName = ([Uri]$ep.Url).Host
    Write-Host -NoNewline "Checking $($ep.Name)... "
    try {
        $ips = @()
        try {
            $ips = [System.Net.Dns]::GetHostAddresses($hostName) | ForEach-Object { $_.IPAddressToString }
        } catch {
            $ips = @()
        }

        if ($ips.Count -gt 0) {
            Write-Host "DNS OK" -ForegroundColor DarkGreen -NoNewline
            Write-Host " (" -NoNewline
            Write-Host ($ips -join ", ") -NoNewline -ForegroundColor Gray
            Write-Host ")" -NoNewline
        } else {
            Write-Host "DNS FAIL" -ForegroundColor Yellow -NoNewline
        }

        $tcpOk = $false
        try {
            $tcpOk = Test-NetConnection -ComputerName $hostName -Port 443 -InformationLevel Quiet
        } catch {
            $tcpOk = $false
        }

        Write-Host " | TCP 443: " -NoNewline
        if ($tcpOk) {
            Write-Host "OPEN" -ForegroundColor DarkGreen -NoNewline
        } else {
            Write-Host "CLOSED" -ForegroundColor Yellow -NoNewline
        }

        Write-Host " | " -NoNewline

        $tls12Ok = Test-TlsHandshake -HostName $hostName -Protocol ([System.Security.Authentication.SslProtocols]::Tls12)
        $tls11Ok = Test-TlsHandshake -HostName $hostName -Protocol ([System.Security.Authentication.SslProtocols]::Tls11)

        Write-Host "TLS1.2: " -NoNewline
        if ($tls12Ok) {
            Write-Host "OK" -ForegroundColor DarkGreen -NoNewline
        } else {
            Write-Host "FAIL" -ForegroundColor Yellow -NoNewline
        }

        Write-Host " | TLS1.1: " -NoNewline
        if ($tls11Ok) {
            Write-Host "OK" -ForegroundColor DarkGreen -NoNewline
        } else {
            Write-Host "FAIL" -ForegroundColor Yellow -NoNewline
        }

        Write-Host " | " -NoNewline

        # TLS certificate inspection (best-effort)
        $certInfo = $null
        try {
            $tcpClient = New-Object System.Net.Sockets.TcpClient($hostName, 443)
            $sslStream = New-Object System.Net.Security.SslStream($tcpClient.GetStream(), $false, ({ $true }))
            $sslStream.AuthenticateAsClient($hostName)
            $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($sslStream.RemoteCertificate)
            $certInfo = @{
                Subject = $cert.Subject
                Issuer  = $cert.Issuer
                NotAfter = $cert.NotAfter
            }
            $sslStream.Dispose()
            $tcpClient.Close()
        } catch {
            $certInfo = $null
        }

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
        if ($certInfo) {
            Write-Host "   > Cert Subject: $($certInfo.Subject)" -ForegroundColor Gray
            Write-Host "   > Cert Issuer : $($certInfo.Issuer)" -ForegroundColor Gray
            Write-Host "   > Cert Expiry : $($certInfo.NotAfter)" -ForegroundColor Gray
        } else {
            Write-Host "   > Cert Info  : Unavailable (TLS handshake failed)" -ForegroundColor Gray
        }
    }
}

Write-Host "`n========================================"
Write-Host "Note: If Cloudflare is active, 520-526 errors mean Backend is unreachable."
Write-Host "      If Timeout, check Load Balancer Security Groups."
