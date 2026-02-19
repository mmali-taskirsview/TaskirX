# TaskirX - Get OCI Ingress IP
# Retreives the Public IP assigned by OCI Load Balancer to the Nginx Ingress Controller

Write-Host "Waiting for Load Balancer IP..." -ForegroundColor Cyan

$attempts = 0
$maxAttempts = 20
$ip = $null

while (-not $ip -and $attempts -lt $maxAttempts) {
    # Try getting IP from the Service (LoadBalancer)
    # Note: Service name depends on Helm release name (nginx-ingress)
    # Checking for common names: 'ingress-nginx-controller' or 'ingress-nginx-nginx-ingress-controller'
    
    $svcName = "ingress-nginx-controller"
    $check = kubectl get svc ingress-nginx-nginx-ingress-controller -n ingress-nginx 2>$null
    if ($check) { $svcName = "ingress-nginx-nginx-ingress-controller" }

    $json = kubectl get svc $svcName -n ingress-nginx -o json 2>$null | ConvertFrom-Json
    
    if ($json) {
        $ip = $json.status.loadBalancer.ingress[0].ip
    }

    if (-not $ip) {
        Write-Host "  IP pending... ($attempts/$maxAttempts)" -ForegroundColor Gray
        Start-Sleep -Seconds 10
        $attempts++
    }
}

if ($ip) {
    Write-Host "`n✅ Load Balancer Public IP: $ip" -ForegroundColor Green
    Write-Host "`nAction Required:" -ForegroundColor Yellow
    Write-Host "1. Go to Cloudflare Dashboard"
    Write-Host "2. Update A records (api, dashboard, bidding) to: $ip"
    Write-Host "   OR Update terraform.tfvars and re-apply (if configured)"
    
    return $ip
} else {
    Write-Host "`n❌ Failed to retrieve IP. Load Balancer might still be provisioning." -ForegroundColor Red
    Write-Host "Check status: kubectl get svc -n ingress-nginx"
    exit 1
}
