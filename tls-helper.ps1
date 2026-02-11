# TaskirX TLS Configuration Helper
# This script helps with the TLS setup process

param(
    [string]$Action = "help",
    [string]$LBIP = "",
    [string]$CloudflareToken = "",
    [string]$ZoneId = ""
)

function Show-Help {
    Write-Host "TaskirX TLS Configuration Helper" -ForegroundColor Cyan
    Write-Host "=================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage:" -ForegroundColor Yellow
    Write-Host "  .\tls-helper.ps1 -Action <action> [parameters]"
    Write-Host ""
    Write-Host "Actions:" -ForegroundColor Yellow
    Write-Host "  deploy      - Deploy infrastructure to OCI"
    Write-Host "  get-ip      - Get Load Balancer IP after deployment"
    Write-Host "  update-dns  - Update Cloudflare DNS records"
    Write-Host "  test-tls    - Test TLS configuration"
    Write-Host "  help        - Show this help"
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\tls-helper.ps1 -Action deploy"
    Write-Host "  .\tls-helper.ps1 -Action get-ip"
    Write-Host "  .\tls-helper.ps1 -Action update-dns -LBIP 123.456.789.0 -CloudflareToken YOUR_TOKEN -ZoneId YOUR_ZONE_ID"
    Write-Host "  .\tls-helper.ps1 -Action test-tls"
}

function Deploy-Infrastructure {
    Write-Host "Step 1: Deploying Infrastructure to OCI..." -ForegroundColor Green
    Write-Host "==========================================" -ForegroundColor Green

    # Check if registry is set
    $registry = Read-Host "Enter your OCI container registry (e.g., iad.ocir.io/your-tenancy/taskir)"

    if (-not $registry) {
        Write-Error "Registry is required. Get it from OCI Console > Developer Services > Container Registry"
        exit 1
    }

    Write-Host "Deploying with registry: $registry" -ForegroundColor Yellow

    # Run deployment
    & ".\scripts\deploy-to-oci.ps1" -Action apply -Registry $registry

    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Infrastructure deployed successfully!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Next: Run '.\tls-helper.ps1 -Action get-ip' to get the Load Balancer IP" -ForegroundColor Cyan
    } else {
        Write-Error "Deployment failed. Check the logs above."
        exit 1
    }
}

function Get-LB-IP {
    Write-Host "Step 2: Getting Load Balancer IP..." -ForegroundColor Green
    Write-Host "================================" -ForegroundColor Green

    # Run the get-ingress-ip script
    & ".\scripts\get-ingress-ip.ps1"

    Write-Host ""
    Write-Host "Copy the IP address shown above." -ForegroundColor Yellow
    Write-Host "Next: Run '.\tls-helper.ps1 -Action update-dns -LBIP <ip-address>'" -ForegroundColor Cyan
}

function Update-DNS {
    Write-Host "Step 3: Updating Cloudflare DNS Records..." -ForegroundColor Green
    Write-Host "=========================================" -ForegroundColor Green

    if (-not $LBIP) {
        $LBIP = Read-Host "Enter the Load Balancer IP from Step 2"
    }

    if (-not $CloudflareToken) {
        $CloudflareToken = Read-Host "Enter your Cloudflare API Token (create at https://dash.cloudflare.com/profile/api-tokens)"
    }

    if (-not $ZoneId) {
        $ZoneId = Read-Host "Enter your Cloudflare Zone ID (found in Overview tab of your domain)"
    }

    Write-Host "Updating DNS records for IP: $LBIP" -ForegroundColor Yellow

    # DNS records to update
    $records = @(
        @{ name = "api"; type = "A" },
        @{ name = "dashboard"; type = "A" },
        @{ name = "bidding"; type = "A" }
    )

    foreach ($record in $records) {
        Write-Host "Updating $($record.name).taskir.com..." -ForegroundColor Gray

        # Get existing record
        $existing = Invoke-RestMethod -Uri "https://api.cloudflare.com/client/v4/zones/$ZoneId/dns_records?type=$($record.type)&name=$($record.name).taskir.com" -Headers @{
            "Authorization" = "Bearer $CloudflareToken"
            "Content-Type" = "application/json"
        }

        if ($existing.result.Count -gt 0) {
            $recordId = $existing.result[0].id

            # Update record
            $body = @{
                type = $record.type
                name = "$($record.name).taskir.com"
                content = $LBIP
                ttl = 300
                proxied = $true
            } | ConvertTo-Json

            Invoke-RestMethod -Method PUT -Uri "https://api.cloudflare.com/client/v4/zones/$ZoneId/dns_records/$recordId" -Headers @{
                "Authorization" = "Bearer $CloudflareToken"
                "Content-Type" = "application/json"
            } -Body $body

            Write-Host "✅ Updated $($record.name).taskir.com" -ForegroundColor Green
        } else {
            Write-Host "⚠️  Record $($record.name).taskir.com not found, creating..." -ForegroundColor Yellow

            # Create record
            $body = @{
                type = $record.type
                name = "$($record.name).taskir.com"
                content = $LBIP
                ttl = 300
                proxied = $true
            } | ConvertTo-Json

            Invoke-RestMethod -Method POST -Uri "https://api.cloudflare.com/client/v4/zones/$ZoneId/dns_records" -Headers @{
                "Authorization" = "Bearer $CloudflareToken"
                "Content-Type" = "application/json"
            } -Body $body

            Write-Host "✅ Created $($record.name).taskir.com" -ForegroundColor Green
        }
    }

    Write-Host ""
    Write-Host "✅ DNS records updated!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "1. Wait 5-10 minutes for DNS propagation"
    Write-Host "2. Go to Cloudflare Dashboard > SSL/TLS > Overview"
    Write-Host "3. Set SSL mode to 'Full (Strict)'"
    Write-Host "4. Run '.\tls-helper.ps1 -Action test-tls'"
}

function Test-TLS {
    Write-Host "Step 4: Testing TLS Configuration..." -ForegroundColor Green
    Write-Host "===================================" -ForegroundColor Green

    Write-Host "Running TLS tests..." -ForegroundColor Yellow

    # Run the verification script
    & ".\scripts\verify-live-endpoints.ps1"

    Write-Host ""
    Write-Host "If all endpoints show 'OPEN (200 OK)', TLS is working!" -ForegroundColor Green
    Write-Host "If you see TLS errors, wait a few more minutes and try again." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Final step: Run 'scripts\validate-all.ps1' for full system test" -ForegroundColor Cyan
}

# Main logic
switch ($Action.ToLower()) {
    "deploy" { Deploy-Infrastructure }
    "get-ip" { Get-LB-IP }
    "update-dns" { Update-DNS }
    "test-tls" { Test-TLS }
    "help" { Show-Help }
    default {
        Write-Error "Unknown action: $Action"
        Show-Help
        exit 1
    }
}