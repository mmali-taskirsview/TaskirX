param(
    [string]$TargetHost = "http://localhost:8080",
    [int]$Users = 20,
    [int]$SpawnRate = 5,
    [string]$RunTime = "2m",
    [string]$ReportPrefix = "performance-stats-mixed"
)

Write-Host "Running Mixed Format Load Test against $TargetHost" -ForegroundColor Cyan
Write-Host "Simulating: 50% Banner, 20% Native, 15% Video, 10% Rich Media, 5% Audio" -ForegroundColor Green

# Use provided ReportPrefix for CSV and HTML output
locust -f performance-tests/mixed_format_load.py --host $TargetHost --headless -u $Users -r $SpawnRate -t $RunTime --html "$ReportPrefix.html" --csv "$ReportPrefix"

Write-Host "Test Complete. Check log/report: $ReportPrefix.html" -ForegroundColor Cyan
