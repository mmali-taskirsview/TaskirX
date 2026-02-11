# TaskirX Docker Setup

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  TaskirX v3.0 - Docker Setup" -ForegroundColor Cyan
Write-Host "======================================`n" -ForegroundColor Cyan

# Check Docker
Write-Host "Checking Docker..." -ForegroundColor Yellow
$dockerVersion = docker --version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker not found. Please install Docker Desktop." -ForegroundColor Red
    Write-Host "   Download: https://www.docker.com/products/docker-desktop" -ForegroundColor Yellow
    exit 1
}
Write-Host "✅ $dockerVersion" -ForegroundColor Green

# Check Docker Compose
$composeVersion = docker-compose --version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker Compose not found." -ForegroundColor Red
    exit 1
}
Write-Host "✅ $composeVersion`n" -ForegroundColor Green

# Check if Docker is running
Write-Host "Checking Docker daemon..." -ForegroundColor Yellow
docker ps >$null 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker daemon is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}
Write-Host "✅ Docker daemon is running`n" -ForegroundColor Green

# Create .env file if it doesn't exist
if (-not (Test-Path ".env")) {
    Write-Host "Creating .env file..." -ForegroundColor Yellow
    Copy-Item ".env.docker" ".env"
    Write-Host "⚠️  Please update .env with your secrets!" -ForegroundColor Yellow
    Write-Host ""
}

# Build all services
Write-Host "Building Docker images..." -ForegroundColor Yellow
Write-Host "This may take 5-10 minutes...`n" -ForegroundColor Gray

docker-compose build --no-cache

if ($LASTEXITCODE -ne 0) {
    Write-Host "`n❌ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n✅ All images built successfully!`n" -ForegroundColor Green

# Start services
Write-Host "Starting services..." -ForegroundColor Yellow
docker-compose up -d

if ($LASTEXITCODE -ne 0) {
    Write-Host "`n❌ Failed to start services!" -ForegroundColor Red
    exit 1
}

Write-Host "`n✅ All services started!`n" -ForegroundColor Green

# Wait for services to be healthy
Write-Host "Waiting for services to be healthy..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Check service health
Write-Host "`nService Status:`n" -ForegroundColor Cyan
docker-compose ps

Write-Host "`n======================================" -ForegroundColor Cyan
Write-Host "  TaskirX v3.0 is Ready! 🚀" -ForegroundColor Cyan
Write-Host "======================================`n" -ForegroundColor Cyan

Write-Host "Services:" -ForegroundColor Yellow
Write-Host "  • Dashboard:      http://localhost:3001" -ForegroundColor White
Write-Host "  • NestJS API:     http://localhost:3000" -ForegroundColor White
Write-Host "  • Go Bidding:     http://localhost:8080" -ForegroundColor White
Write-Host "  • Fraud Detection: http://localhost:6001" -ForegroundColor White
Write-Host "  • Ad Matching:     http://localhost:6002" -ForegroundColor White
Write-Host "  • Bid Optimization: http://localhost:6003" -ForegroundColor White
Write-Host "  • PostgreSQL:      localhost:5432" -ForegroundColor White
Write-Host "  • Redis:           localhost:6379" -ForegroundColor White
Write-Host "  • ClickHouse:      localhost:8123" -ForegroundColor White

Write-Host "`nUseful Commands:" -ForegroundColor Yellow
Write-Host "  • View logs:       docker-compose logs -f" -ForegroundColor Gray
Write-Host "  • Stop services:   docker-compose down" -ForegroundColor Gray
Write-Host "  • Restart:         docker-compose restart" -ForegroundColor Gray
Write-Host "  • Check status:    docker-compose ps" -ForegroundColor Gray

Write-Host "`nHappy coding! 🎉`n" -ForegroundColor Green
