# TaskirX V3 Platform Launcher
# This script starts the full Polyglot Stack (Go, NestJS, Python) using Docker Compose.

Write-Host "Starting TaskirX V3 Platform (Polyglot Edition)..." -ForegroundColor Cyan

# Check if Docker is running
docker info >$null 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

Write-Host "Spinning up application services..." -ForegroundColor Yellow
docker-compose up -d --build

Write-Host "Spinning up monitoring stack..." -ForegroundColor Yellow
docker-compose -f monitoring/docker-compose.yml up -d

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ All services started successfully!" -ForegroundColor Green
    Write-Host "`nService Endpoints:" -ForegroundColor Cyan
    Write-Host "   Frontend Dashboard:  http://localhost:3001"
    Write-Host "   Backend API (Nest):  http://localhost:3000/api"
    Write-Host "   Bidding Engine (Go): http://localhost:8080"
    Write-Host "   Prometheus:          http://localhost:9090"
    Write-Host "   Grafana:             http://localhost:3002"
    Write-Host "   ClickHouse DB:       http://localhost:8123"
    
    Write-Host "`nLogs are streaming in background. Run 'docker-compose logs -f' to view." -ForegroundColor Gray
    
    # Optional: Open Dashboard
    Start-Process "http://localhost:3001"
} else {
    Write-Host "`n❌ Failed to start services." -ForegroundColor Red
}

Read-Host "Press Enter to exit Launcher..."
