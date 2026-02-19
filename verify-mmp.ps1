# Verify MMP Integration Endpoints

Write-Host "Verifying MMP Integration via Docker Container..." -ForegroundColor Cyan

# 1. Track Install Event
Write-Host "1. Tracking Install Event (POST /events/track)..." -NoNewline
docker exec taskir-nestjs curl -s -X POST http://localhost:3000/api/mmp/events/track `
  -H "Content-Type: application/json" `
  -d '{\"provider\":\"script_test\",\"eventType\":\"install\",\"campaignId\":\"cmp-script-1\",\"revenue\":1.50}'

Write-Host "`n"

# 2. Check ClickHouse
Write-Host "2. Verifying Data in ClickHouse..." -NoNewline
$result = docker exec taskir-clickhouse clickhouse-client --user taskir --password clickhouse_password_2026 --query "SELECT * FROM analytics.mmp_events WHERE provider='script_test' ORDER BY timestamp DESC LIMIT 1 FORMAT JSONEachRow"
if ($result) {
    Write-Host " Success!" -ForegroundColor Green
    Write-Host $result
} else {
    Write-Host " Failed! No data found." -ForegroundColor Red
}

# 3. Simulate Postback
Write-Host "3. Simulating Postback (GET /postback)..." -NoNewline
docker exec taskir-nestjs curl -s "http://localhost:3000/api/mmp/postback?provider=adjust&event_name=purchase&c=cmp-adjust-script&event_revenue=10.00"
Write-Host "`n"

Write-Host "MMP Verification Complete." -ForegroundColor Yellow
