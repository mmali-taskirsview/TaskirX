param([string]$SqlFile)

if (-not (Test-Path $SqlFile)) {
    Write-Error "File not found: $SqlFile"
    exit 1
}

$Content = Get-Content -Raw $SqlFile
# rudimentary splitting by ; at end of line or specific pattern
$Commands = $Content -split ";\r?\n"

$Headers = @{
    "X-ClickHouse-User" = "taskir"
    "X-ClickHouse-Key"  = "clickhouse_password_2026"
}

foreach ($cmd in $Commands) {
    $cmd = $cmd.Trim()
    if ($cmd) {
        Write-Host "Executing: $cmd"
        try {
            # Use POST body for the query
            Invoke-RestMethod -Uri "http://localhost:8123/" -Method Post -Body $cmd -Headers $Headers
            Write-Host "Success" -ForegroundColor Green
        } catch {
            Write-Error "Failed: $_"
        }
    }
}
