Write-Host "Starting Engagement Simulation..." -ForegroundColor Cyan

$endpoint = "http://localhost:8080/track"
Write-Host "Endpoint configured: $endpoint"
$campaignId = "camp-sim-001"

# Define weights for events (Impression > Click > Video Start > Complete)
$events = @(
    @{ Type="impression"; Weight=100 },
    @{ Type="click"; Weight=10 },
    @{ Type="view"; Weight=50 },
    @{ Type="video_start"; Weight=40 },
    @{ Type="first_quartile"; Weight=35 },
    @{ Type="midpoint"; Weight=30 },
    @{ Type="third_quartile"; Weight=25 },
    @{ Type="complete"; Weight=20 },
    @{ Type="expand"; Weight=5 },
    @{ Type="collapse"; Weight=4 },
    @{ Type="interact"; Weight=8 }
)

$totalRequests = 500
$count = 0

while ($count -lt $totalRequests) {
    $event = $events | Get-Random
    # Simple probability check
    if ((Get-Random -Minimum 0 -Maximum 100) -le $event.Weight) {
        $type = $event.Type
        
        # Map some types to internal metric names if needed, but handler uses raw strings
        if ($type -eq "video_start") { $type = "start" }
        
        try {
            Write-Host "DEBUG-LOOP: Endpoint is: '$endpoint'"
            $url = "$endpoint" + "?event=$type&id=$campaignId"
            Write-Host "DEBUG: $url"
            Invoke-WebRequest -Uri $url -Method Get -UseBasicParsing | Out-Null
            Write-Host -NoNewline "."
            $count++
        } catch {
            Write-Host -NoNewline "x"
            Write-Host $_.Exception.Message
        }
        
        if ($count % 50 -eq 0) { Write-Host " $count" }
        
        Start-Sleep -Milliseconds 10
    }
}

Write-Host "`nSimulation Complete. Check Grafana Dashboard." -ForegroundColor Green
