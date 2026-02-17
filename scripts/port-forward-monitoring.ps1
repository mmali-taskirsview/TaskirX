Write-Host "Starting Port Forwarding for Monitoring Stack..." -ForegroundColor Cyan

$job = Start-Job -ScriptBlock {
    kubectl port-forward svc/grafana 3000:3000 -n monitoring
}
Write-Host "Forwarding Grafana on http://localhost:3000 (User: admin / Pass: taskir_admin)"

$job2 = Start-Job -ScriptBlock {
    kubectl port-forward svc/prometheus 9090:9090 -n monitoring
}
Write-Host "Forwarding Prometheus on http://localhost:9090"

Write-Host "Press Enter to stop forwarding..."
Read-Host
Stop-Job $job
Stop-Job $job2
Remove-Job $job
Remove-Job $job2
