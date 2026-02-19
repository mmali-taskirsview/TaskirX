param(
    [string]$ContainerName = "taskir-postgres",
    [string]$DbUser = "taskir",
    [string]$DbName = "taskir_adx",
    [string]$SqlFile = "./performance-indexes.sql"
)

Write-Host "Applying Database Performance Indexes..." -ForegroundColor Cyan

if (-not (Test-Path $SqlFile)) {
    Write-Error "SQL file not found: $SqlFile"
    exit 1
}

# Copy SQL file to container temporarily
docker cp $SqlFile "$ContainerName`:/tmp/indexes.sql"

if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to copy SQL file to container."
    exit 1
}

# Execute SQL file inside container
docker exec -u postgres $ContainerName psql -U $DbUser -d $DbName -f /tmp/indexes.sql

if ($LASTEXITCODE -eq 0) {
    Write-Host "Success: Indexes applied." -ForegroundColor Green
} else {
    Write-Error "Failed to apply indexes."
}
