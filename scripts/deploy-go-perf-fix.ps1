# TaskirX - Deploy Go Bidding Performance Fix
# Builds and pushes the optimized Go Bidding Engine to OCIR and restarts the K8S deployment.

param(
    [string]$Registry = "sin.ocir.io/axoodqjqcaam",
    [string]$Version = "latest"
)

$Service = "go-bidding"
$Path = "./go-bidding-engine"
$ImageName = "taskir-$Service"
$Tag = "$Registry/$ImageName`:$Version"

Write-Host "Deploying Performance Fix for $Service..." -ForegroundColor Cyan
Write-Host "Registry: $Registry" -ForegroundColor Yellow

# 1. Build
Write-Host "`n[1/4] Building Docker Image..." -ForegroundColor Cyan
docker build -t $ImageName $Path
if ($LASTEXITCODE -ne 0) { Write-Error "Build failed"; exit 1 }

# 2. Tag
Write-Host "`n[2/4] Tagging Image..." -ForegroundColor Cyan
docker tag "$ImageName`:latest" $Tag

# 3. Push
Write-Host "`n[3/4] Pushing to OCIR..." -ForegroundColor Cyan
docker push $Tag
if ($LASTEXITCODE -ne 0) { 
    Write-Warning "Push failed. Ensure you are logged in to OCIR."
    Write-Warning "Run: docker login $Registry"
    exit 1 
}

# 4. Rollout Restart
Write-Host "`n[4/4] Restarting Kubernetes Deployment..." -ForegroundColor Cyan
kubectl rollout restart deployment/go-bidding -n taskir
if ($LASTEXITCODE -ne 0) { Write-Error "Kubernetes rollout failed"; exit 1 }

Write-Host "`nSUCCESS! Performance fix deployed." -ForegroundColor Green
Write-Host "Monitor stats with: kubectl top pods -n taskir" -ForegroundColor Gray
