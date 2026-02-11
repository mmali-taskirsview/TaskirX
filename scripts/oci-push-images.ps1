# TaskirX - OCI Image Push Helper
# Automates building, tagging, and pushing services to Oracle Container Registry (OCIR)

param(
    [Parameter(Mandatory=$true)]
    [string]$Registry,
    
    [string]$Version = "latest"
)

$Services = @{
    "nestjs" = "./nestjs-backend"
    "go-bidding" = "./go-bidding-engine"
    "dashboard" = "./next-dashboard"
    "ad-matching" = "./python-ai-agents/ad-matching-service"
    "fraud-detection" = "./python-ai-agents/fraud-detection-service"
    "bid-optimization" = "./python-ai-agents/bid-optimization-service"
}

Write-Host "TaskirX OCI Image Push" -ForegroundColor Cyan
Write-Host "Target Registry: $Registry" -ForegroundColor Yellow
Write-Host "Version: $Version" -ForegroundColor Yellow

# Check Docker
docker info >$null 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Docker not running" -ForegroundColor Red
    exit 1
}

foreach ($Service in $Services.Keys) {
    $Path = $Services[$Service]
    $ImageName = "taskir-$Service"
    $Tag = "$Registry/$ImageName`:$Version"
    
    Write-Host "`nProcessing $Service..." -ForegroundColor Cyan
    
    # Build
    Write-Host "  Building..."
    docker build -t $ImageName $Path
    if ($LASTEXITCODE -ne 0) { Write-Error "Build failed for $Service"; continue }

    # Tag
    Write-Host "  Tagging as $Tag..."
    docker tag "$ImageName`:latest" $Tag

    # Push
    Write-Host "  Pushing..."
    docker push $Tag
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Pushed: $Tag" -ForegroundColor Green
    } else {
        Write-Host "❌ Failed to push $Tag. Check `docker login`." -ForegroundColor Red
    }
}

Write-Host "`nAll operations complete." -ForegroundColor Cyan
