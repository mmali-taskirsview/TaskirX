# OCI Deployment Handover Guide - v3 (Updated 2026-02-17)

**Environment**: Production (Oracle Cloud)
**Status**: Live & Observable

## 1. Access Credentials
*   **Cluster**: oke_taskir_cluster
*   **Namespace**: 	askir
*   **Public IP**: 138.2.76.159
*   **Monitoring**: metrics.taskirx.com (Internal/Port-Forward)

## 2. Service Map & Endpoints

| Component | Internal Service | Hostname (DNS) | Public Port | Local Bypass |
| :--- | :--- | :--- | :--- | :--- |
| **Backend API** | 
estjs-backend | pi.taskirx.com | 80/443 | kubectl port-forward svc/nestjs-backend 3000:3000 |
| **Dashboard** | 
ext-dashboard | dashboard.taskirx.com | 80/443 | kubectl port-forward svc/next-dashboard 3001:3001 |
| **Bidding Engine** | go-bidding | idding.taskirx.com | 80/443 | kubectl port-forward svc/go-bidding 8080:8080 |
| **AI Matching** | d-matching | (None - Internal) | - | kubectl port-forward svc/ad-matching 6002:6002 |
| **Monitoring** | grafana | (None - Internal) | - | kubectl port-forward svc/taskir-grafana 3000:3000 |

## 3. Required Actions (DNS Configuration)
Go to your DNS provider (Namecheap, GoDaddy, Cloudflare) and create these **A Records**:
- **Name**: pi -> **Value**: 138.2.76.159
- **Name**: dashboard -> **Value**: 138.2.76.159
- **Name**: idding -> **Value**: 138.2.76.159
- **Name**: 	askirx.com -> **Value**: 138.2.76.159 (and www)

> **Note**: HTTPS/SSL certificates will remain in "Pending" state until these DNS records are live.

## 4. Monitoring & Logs
To check live logs:
`powershell
kubectl logs -n taskir -l app=nestjs-backend -f
kubectl logs -n taskir -l app=go-bidding -f
`

## 5. Maintenance Checklist
- **Daily**: Check "Redis Cache Hit Rate" on Grafana.
- **Weekly**: Backup Postgres & ClickHouse volumes using scripts/backup-oci.ps1.

## 6. Known Issues
- **CORS Errors**: If you see CORS errors in the dashboard, ensure you have run .\update-domain.ps1 to rebuild the backend with the new TaskirX.com domain settings.

## 7. DNS Configuration (Cloudflare)
**Domain:** 	askirx.com
**Load Balancer IP:** 138.2.76.159

| Type | Name | Value | Proxy |
|---|---|---|---|
| A | @ | 138.2.76.159 | Proxied |
| A | www | 138.2.76.159 | Proxied |
| A | api | 138.2.76.159 | Proxied |
| A | dashboard | 138.2.76.159 | Proxied |
| A | bidding | 138.2.76.159 | Proxied |
