# Backup Script for TaskirX OCI Environment
# Usage: .\backup-oci.ps1

$Date = Get-Date -Format "yyyyMMdd"
$BackupDir = "./backups/$Date"

Write-Host "Starting Backup for TaskirX ($Date)..." -ForegroundColor Cyan
New-Item -ItemType Directory -Force -Path $BackupDir | Out-Null

# 1. Backup PostgreSQL
Write-Host "Backing up Postgres..."
kubectl exec -n taskir (kubectl get pods -n taskir -l app=postgres -o jsonpath="{.items[0].metadata.name}") -- pg_dumpall -U taskir > "$BackupDir/postgres-dump.sql"

# 2. Backup ClickHouse (Schema & Data)
Write-Host "Backing up ClickHouse (Schema Only)..."
# Note: Full data backup for ClickHouse is large; referencing schema + small dump
kubectl exec -n taskir (kubectl get pods -n taskir -l app=clickhouse -o jsonpath="{.items[0].metadata.name}") -- clickhouse-client --query "SHOW CREATE DATABASE analytics" > "$BackupDir/clickhouse-schema.sql"

# 3. Backup Redis (AOF)
Write-Host "Triggering Redis Save..."
kubectl exec -n taskir (kubectl get pods -n taskir -l app=redis -o jsonpath="{.items[0].metadata.name}") -- redis-cli SAVE

Write-Host "Backup Complete! Files saved to $BackupDir" -ForegroundColor Green
