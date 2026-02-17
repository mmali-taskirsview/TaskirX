# TaskirX Maintenance & Operations Guide

This guide provides standard operating procedures for maintaining the TaskirX Ad Exchange platform in production (OCI Kubernetes).

## 1. Monitoring & Logs

### Viewing Logs (Real-time)
To tail the logs of the core services:
```powershell
# Bidding Engine (High Volume)
kubectl logs -n taskir -l app=go-bidding -f --tail=100

# API Backend (Errors/Business Logic)
kubectl logs -n taskir -l app=nestjs-backend -f --tail=100

# AI Services
kubectl logs -n taskir -l app=ad-matching -f
kubectl logs -n taskir -l app=fraud-detection -f
```

### Checking Service Health
```powershell
kubectl get pods -n taskir
# Look for STATUS: Running and RESTARTS: 0 (or low count)
```

## 2. Deploying Updates

### Code Changes
1.  **Commit** your changes to git.
2.  **Run** the update script (builds, pushes, and rolls out):
    ```powershell
    .\update-domain.ps1
    ```
    *Note: This script updates ALL services. For single service updates, use specific deployment scripts like `deploy-go-perf-fix.ps1`.*

### Configuration Changes (Environment Variables)
1.  **Edit** the ConfigMap in `k8s/<service>-deployment.yaml`.
2.  **Apply** the change:
    ```powershell
    kubectl apply -f k8s/<service>-deployment.yaml
    ```
3.  **Restart** the deployment to pick up changes:
    ```powershell
    kubectl rollout restart deployment/<service-name> -n taskir
    ```

## 3. Database Management

### PostgreSQL Backup (Manual)
```powershell
# Create a dump from the running pod
kubectl exec -it -n taskir (kubectl get pod -n taskir -l app=postgres -o jsonpath="{.items[0].metadata.name}") -- pg_dump -U taskir taskir_adx > backup_$(Get-Date -Format "yyyyMMdd").sql
```

### Redis Cache Clearing
If you need to purge the cache (Campaigns/User Segments):
```powershell
# Connect to Redis pod
kubectl exec -it -n taskir (kubectl get pod -n taskir -l app=redis -o jsonpath="{.items[0].metadata.name}") -- redis-cli -a taskir_redis_password_2026 FLUSHALL
```

## 4. Scaling
The platform uses Horizontal Pod Autoscaling (HPA).
To modify scaling thresholds (CPU/Memory):
1.  Edit `k8s/hpa.yaml`.
2.  Apply changes: `kubectl apply -f k8s/hpa.yaml`.

## 5. Troubleshooting Common Issues

### "Campaign Not Matching"
1.  Check Redis Cache: The campaign might be cached with old targeting rules. Run `FLUSHALL` (see above).
2.  Check Logs: `kubectl logs -l app=go-bidding` to see why a bid was filtered (e.g., "Filtered by Geo", "Filtered by Budget").

### "High Latency"
1.  Check if `fraud-detection` service is slow.
2.  Review Grafana dashboards for a spike in "Cache Misses".
