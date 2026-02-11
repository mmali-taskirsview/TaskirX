# TaskirX Message: Docker Hub Block Bypass Script
# This script manually downloads images from a working mirror and retags them.

$Mirror = "docker.m.daocloud.io"

$Images = @(
    "library/python:3.10-slim",
    "library/node:18-alpine",
    "library/golang:1.21-alpine",
    "library/alpine:latest",
    "prom/prometheus:latest",
    "prom/alertmanager:latest",
    "grafana/grafana:latest",
    "prom/node-exporter:latest"
)

# Special handling for gcr.io
# gcr.io/cadvisor/cadvisor -> gcr.m.daocloud.io/cadvisor/cadvisor ? 
# Let's try basic ones first.

foreach ($Img in $Images) {
    $Source = "$Mirror/$Img"
    # Remove 'library/' for official retag regular usage if needed, but 'library/' is standard.
    # docker tag mirror/library/python:3.10-slim python:3.10-slim
    
    # Target name (Official)
    $Target = $Img -replace "^library/", "" 
    
    Write-Host "Pulling $Source..."
    docker pull $Source
    
    if ($?) {
        Write-Host "Re-tagging to $Target..."
        docker tag $Source $Target
        Write-Host "Success: $Target is ready." -ForegroundColor Green
    } else {
        Write-Host "Failed to pull $Source" -ForegroundColor Red
    }
}

# Handle cadvisor separately if needed
# Try pulling cadvisor from docker hub mirror if available (google/cadvisor)
Write-Host "Attempting google/cadvisor fallback..."
docker pull "$Mirror/google/cadvisor:latest"
docker tag "$Mirror/google/cadvisor:latest" "gcr.io/cadvisor/cadvisor:latest"
