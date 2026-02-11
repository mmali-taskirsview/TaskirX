# TaskirX Cloud Dashboard Launcher
# Opens all the relevant URLs for your production OCI environment

Write-Host "Opening TaskirX Cloud Dashboards..." -ForegroundColor Cyan

# Define Cloud URLs
$Urls = @(
    "https://dashboard.taskir.com",
    "https://api.taskir.com/api",
    "https://bidding.taskir.com/health",
    "https://cloud.oracle.com",
    "https://dash.cloudflare.com",
    "https://app.pinecone.io"
)

foreach ($url in $Urls) {
    Write-Host "Opening $url..."
    Start-Process $url
}

Write-Host "`nDashboards opened in your default browser." -ForegroundColor Green
