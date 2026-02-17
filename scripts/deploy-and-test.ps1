Write-Host "🚀 Starting TaskirX End-to-End Test Deployment..." -ForegroundColor Cyan

# Check for Docker
if (-not (Get-Command docker-compose -ErrorAction SilentlyContinue)) {
    Write-Error "❌ docker-compose not found. Please install Docker Desktop."
    exit 1
}

# 1. Clean up old containers
Write-Host "🧹 Cleaning up previous deployment..." -ForegroundColor Yellow
docker-compose down

# 2. Build and Start Services
Write-Host "🏗️  Building and starting services (this may take a few minutes)..." -ForegroundColor Yellow
docker-compose up -d --build

if ($LASTEXITCODE -ne 0) {
    Write-Error "❌ Failed to start docker services."
    exit 1
}

Write-Host "⏳ Waiting for services to initialize (30 seconds)..." -ForegroundColor Yellow
Start-Sleep -Seconds 30

# 3. Check container status
docker-compose ps

# 4. Run Integration Test
Write-Host "🧪 Running Python Integration Test..." -ForegroundColor Cyan

# Ensure requests library is installed locally for the test script
pip install requests

$env:PYTHONPATH = "."
python tests/integration/test_full_flow.py

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ End-to-End Test PASSED!" -ForegroundColor Green
} else {
    Write-Error "❌ End-to-End Test FAILED!"
}

# 5. Optional Cleanup
# Uncomment to auto-shutdown
# docker-compose down
