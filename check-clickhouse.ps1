if ("$args") {
    $Query = $args
} else {
    $Query = "DESCRIBE TABLE analytics.impressions FORMAT TabSeparated"
}

$Headers = @{
    "X-ClickHouse-User" = "taskir"
    "X-ClickHouse-Key"  = "clickhouse_password_2026"
}

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8123/?query=$Query" -Method Get -Headers $Headers
    Write-Host "Result:"
    $response
} catch {
    Write-Host "Error: $_"
}
