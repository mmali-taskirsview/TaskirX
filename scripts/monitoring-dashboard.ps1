# Helper script to launch monitoring dashboard
Write-Host "Setting up port forwarding to Grafana..." -ForegroundColor Cyan
Write-Host "Please keep this window open." -ForegroundColor Yellow
Write-Host "Grafana URL: http://localhost:3000" -ForegroundColor Green
Write-Host "Username: admin" -ForegroundColor Green
Write-Host "Password: taskir_admin" -ForegroundColor Green

kubectl port-forward svc/grafana 3000:3000 -n monitoring
