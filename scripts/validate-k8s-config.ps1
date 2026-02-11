Write-Host "TASKIRX KUBERNETES CONFIGURATION VALIDATION" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

# 1. Validate Python Deployment
$pyDeploy = "c:\TaskirX\k8s\python-services-deployment.yaml"
if (Select-String -Path $pyDeploy -Pattern "image: your-registry/taskir-ad-matching:latest") {
    Write-Host "[PASS] Ad Matching Service defined in K8s Deployment" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Ad Matching Service missing from python-services-deployment.yaml" -ForegroundColor Red
}

# 2. Validate Go Configuration
$goDeploy = "c:\TaskirX\k8s\go-bidding-deployment.yaml"
if (Select-String -Path $goDeploy -Pattern "AI_SERVICE_URL") {
    Write-Host "[PASS] Go Deployment has AI_SERVICE_URL env var" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Go Deployment missing AI_SERVICE_URL" -ForegroundColor Red
}

$goMain = "c:\TaskirX\go-bidding-engine\cmd\server\main.go"
if (Select-String -Path $goMain -Pattern "getEnv.*AI_SERVICE_URL") {
    Write-Host "[PASS] Go Main reads AI_SERVICE_URL" -ForegroundColor Green
} else {
    Write-Host "[FAIL] Go Main properly does not read AI_SERVICE_URL" -ForegroundColor Red
}

Write-Host "`nKUBERNETES MANIFEST VALIDATION COMPLETE" -ForegroundColor Cyan
