Write-Host "TASKIRX MASTER VALIDATION SUITE" -ForegroundColor Cyan
Write-Host "===============================" -ForegroundColor Cyan
Write-Host "Executing all phase validation scripts..." -ForegroundColor Gray

$ScriptRoot = $PSScriptRoot

# Phase 2: Scale & Optimize (Go, Redis)
if (Test-Path "$ScriptRoot\validate-phase-2.ps1") {
    Write-Host "`n--- PHASE 2 (Scale) ---" -ForegroundColor Yellow
    powershell -ExecutionPolicy Bypass -File "$ScriptRoot\validate-phase-2.ps1"
}

# Phase 3: Advanced Features (Dashboards, Header Bidding)
if (Test-Path "$ScriptRoot\validate-phase-3.ps1") {
    Write-Host "`n--- PHASE 3 (Dashboards) ---" -ForegroundColor Yellow
    powershell -ExecutionPolicy Bypass -File "$ScriptRoot\validate-phase-3.ps1"
}

# Phase 4: Machine Learning (Python Services, Integration)
if (Test-Path "$ScriptRoot\validate-phase-4-integration.ps1") {
    Write-Host "`n--- PHASE 4 A (Integration) ---" -ForegroundColor Yellow
    powershell -ExecutionPolicy Bypass -File "$ScriptRoot\validate-phase-4-integration.ps1"
}

if (Test-Path "$ScriptRoot\validate-phase-4-advanced.ps1") {
    Write-Host "`n--- PHASE 4 B (Advanced ML) ---" -ForegroundColor Yellow
    powershell -ExecutionPolicy Bypass -File "$ScriptRoot\validate-phase-4-advanced.ps1"
}

if (Test-Path "$ScriptRoot\validate-k8s-config.ps1") {
    Write-Host "`n--- DEPLOYMENT (Kubernetes) ---" -ForegroundColor Yellow
    powershell -ExecutionPolicy Bypass -File "$ScriptRoot\validate-k8s-config.ps1"
}

Write-Host "`nALL CHECKS COMPLETED" -ForegroundColor Cyan
