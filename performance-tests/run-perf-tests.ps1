param(
    [string]$TargetHost = $env:LOCUST_HOST,
    [int]$Users = 10,
    [int]$SpawnRate = 2,
    [string]$RunTime = "30s",
    [switch]$Headed
)

if (-not $TargetHost) {
    $TargetHost = "http://localhost:3000"
}

Write-Host "Running Locust against $TargetHost with $Users users (spawn rate $SpawnRate, time $RunTime)" -ForegroundColor Cyan

$modeArgs = @()
if (-not $Headed) { $modeArgs += "--headless" }

locust -f locustfile.py --host $TargetHost @modeArgs -u $Users -r $SpawnRate -t $RunTime
