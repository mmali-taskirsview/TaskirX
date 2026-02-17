Write-Host "Setting up Grafana Dashboards..." -ForegroundColor Cyan

# Ensure monitoring namespace exists
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -

# 1. Update Prometheus Config
Write-Host "Applying Prometheus Config..."
kubectl apply -f monitoring/k8s-simple/prometheus-all.yaml

# 2. Apply Loki Stack (Logging)
Write-Host "Applying Loki Stack (Logging)..."
kubectl apply -f monitoring/k8s-simple/loki-stack.yaml

# 3. Create Dashboard ConfigMap
Write-Host "Creating Dashboard ConfigMap..."
# Delete existing if any to avoid conflicts
kubectl delete configmap grafana-dashboards-json -n monitoring --ignore-not-found
kubectl create configmap grafana-dashboards-json --from-file=monitoring/k8s-simple/taskir-dashboard.json -n monitoring

# 4. Apply Grafana Deployment and other ConfigMaps
Write-Host "Applying Grafana Deployment..."
kubectl apply -f monitoring/k8s-simple/grafana-all.yaml

# 5. Restart Grafana to pick up changes
Write-Host "Restarting Grafana..."
kubectl rollout restart deployment grafana -n monitoring

Write-Host "✅ Monitoring Setup Complete!" -ForegroundColor Green
Write-Host "Run '.\scripts\port-forward-monitoring.ps1' to access Grafana."
