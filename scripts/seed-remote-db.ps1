# Seed OCI Databases (Postgres & ClickHouse)
# Usage: .\scripts\seed-remote-db.ps1

function Write-Status {
    param($Message, $Color="Cyan")
    Write-Host "[$((Get-Date).ToString('HH:mm:ss'))] $Message" -ForegroundColor $Color
}

$namespace = "taskir"

# 1. Seed Postgres
Write-Status "Locating Postgres Pod..."
$pgPod = kubectl get pods -n $namespace -l app=postgres -o jsonpath="{.items[0].metadata.name}"

if ($pgPod) {
    Write-Status "Found Postgres Pod: $pgPod"
    Write-Status "Copying seed data..."
    kubectl cp scripts/seed-data.sql "$($namespace)/$($pgPod):/tmp/seed-data.sql"
    
    Write-Status "Executing seed script..."
    # ConfigMap defines POSTGRES_USER=taskir and POSTGRES_DB=taskir_adx
    kubectl exec -n $namespace $pgPod -- psql -U taskir -d taskir_adx -f /tmp/seed-data.sql
    
    if ($LASTEXITCODE -eq 0) {
        Write-Status "Postgres seeded successfully!" "Green"
    } else {
        Write-Status "Failed to seed Postgres." "Red"
    }
} else {
    Write-Status "Postgres pod not found!" "Red"
}

# 2. Seed ClickHouse (Floor Prices / Dictionaries)
if (Test-Path "seed-floor-prices.sql") {
    Write-Status "Locating ClickHouse Pod..."
    $chPod = kubectl get pods -n $namespace -l app=clickhouse -o jsonpath="{.items[0].metadata.name}"
    
    if ($chPod) {
        Write-Status "Found ClickHouse Pod: $chPod"
        Write-Status "Copying floor price seed data..."
        kubectl cp seed-floor-prices.sql "$($namespace)/$($chPod):/tmp/seed-floor-prices.sql"
        
        Write-Status "Executing ClickHouse seed..."
        # ClickHouse usually requires authentication if set, using password from secret (env var) if available
        # Assuming default user or looking up secret logic if implemented in the pod, but commonly passed via CLIENT
        # For simplicity, we assume environment variable access or default user for this script, 
        # but in production, we might need to retrieve the password.
        
        # NOTE: Using --password flag requires knowing the password here. 
        # As a workaround, we use the env var inside the pod if available or default to known secret for this setup.
        kubectl exec -n $namespace $chPod -- clickhouse-client --password=clickhouse_password_2026 --multiquery --queries-file=/tmp/seed-floor-prices.sql
        
        if ($LASTEXITCODE -eq 0) {
            Write-Status "ClickHouse seeded successfully!" "Green"
        } else {
            Write-Status "Failed to seed ClickHouse." "Red"
        }
    } else {
        Write-Status "ClickHouse pod not found!" "Red"
    }
}
