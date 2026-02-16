# scripts/build-and-push-oci.ps1
# Script to Build, Tag, and Push TaskirX images to Oracle Cloud Container Registry (OCIR)

param(
    [string]$RegionCode = "sin", # Singapore
    [string]$TenancyNamespace = "axoodqjqcaam",
    [string]$Tag = "latest"
)

$RegistryBase = "$RegionCode.ocir.io/$TenancyNamespace"
$Images = @{
    "taskir-nestjs"           = "./nestjs-backend"
    "taskir-go-bidding"       = "./go-bidding-engine"
    "taskir-dashboard"        = "./next-dashboard"
    "taskir-fraud-detection"  = "./python-ai-agents/fraud-detection-service"
    "taskir-ad-matching"      = "./python-ai-agents/ad-matching-service"
    "taskir-bid-optimization" = "./python-ai-agents/bid-optimization-service"
}

Write-Host "TaskirX OCIR Build & Push Tool" -ForegroundColor Cyan
Write-Host "Registry: $RegistryBase" -ForegroundColor Gray

# Check Login
Write-Host "Checking Docker Login..." -ForegroundColor Yellow
$loginCheck = docker info 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Error "Docker is not running."
    exit 1
}

# Optional: Prompt for login if not logged in (hard to detect, so we assume user might need to)
Write-Host "Ensure you are logged in to $RegionCode.ocir.io" -ForegroundColor White
Write-Host "Command: docker login $RegionCode.ocir.io" -ForegroundColor Gray
Write-Host "User: <TenancyNamespace>/<Username>" -ForegroundColor Gray
Write-Host "Password: <AuthToken>" -ForegroundColor Gray
Write-Host "------------------------------------------------" -ForegroundColor Gray

foreach ($repoName in $Images.Keys) {
    $buildContext = $Images[$repoName]
    $fullImageName = "$RegistryBase/$repoName`:$Tag"

    Write-Host "`n[Building] $repoName from $buildContext..." -ForegroundColor Cyan
    
    if (-not (Test-Path $buildContext)) {
        Write-Error "Build context $buildContext not found!"
        continue
    }

    # Build
    docker build -t $repoName $buildContext
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build $repoName"
        continue
    }

    # Tag
    Write-Host "[Tagging] $repoName -> $fullImageName" -ForegroundColor Cyan
    docker tag "$repoName`:latest" $fullImageName
    
    # Push
    Write-Host "[Pushing] $fullImageName..." -ForegroundColor Yellow
    docker push $fullImageName
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Successfully pushed $fullImageName" -ForegroundColor Green
    } else {
        Write-Error "Failed to push $fullImageName. Check your login credentials."
    }
}

Write-Host "`nAll operations completed." -ForegroundColor Green
