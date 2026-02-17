# OCI Deployment Handover Guide - v2 (Updated 2026-02-16)

**Environment**: Production (Oracle Cloud)
**Status**: Live - Awaiting DNS Propagation

## 1. Access Credentials
*   **Cluster**: `oke_taskir_cluster`
*   **Namespace**: `taskir`
*   **Public IP**: `138.2.76.159`

## 2. Service Map & Endpoints

| Component | Internal Service | Hostname (DNS) | Public Port | Local Bypass |
| :--- | :--- | :--- | :--- | :--- |
| **Backend API** | `nestjs-backend` | `api.taskirx.com` | 80/443 | `kubectl port-forward svc/nestjs-backend 3000:3000` |
| **Dashboard** | `next-dashboard` | `dashboard.taskirx.com` | 80/443 | `kubectl port-forward svc/next-dashboard 3001:3001` |
| **Bidding Engine** | `go-bidding` | `bidding.taskirx.com` | 80/443 | `kubectl port-forward svc/go-bidding 8080:8080` |
| **AI Matching** | `ad-matching` | (None - Internal) | - | `kubectl port-forward svc/ad-matching 6002:6002` |

## 3. Required Actions (DNS Configuration)
Go to your DNS provider (Namecheap, GoDaddy, Cloudflare) and create these **A Records**:

- **Name**: `api` -> **Value**: `138.2.76.159`
- **Name**: `dashboard` -> **Value**: `138.2.76.159`
- **Name**: `bidding` -> **Value**: `138.2.76.159`

> **Note**: HTTPS/SSL certificates will remain in "Pending" state until these DNS records are live.

## 4. Monitoring & Logs
To check live logs:
```powershell
kubectl logs -n taskir -l app=nestjs-backend -f
kubectl logs -n taskir -l app=go-bidding -f
```

## 6. Known Issues
- **CORS Errors**: If you see CORS errors in the dashboard, ensure you have run `.\update-domain.ps1` to rebuild the backend with the new `TaskirX.com` domain settings.

## 7. DNS Configuration (Cloudflare)
**Domain:** `taskirx.com`
**Load Balancer IP:** `138.2.76.159`

| Type | Name | Value | Proxy |
|---|---|---|---|
| A | @ | 138.2.76.159 | Proxied |
| A | www | 138.2.76.159 | Proxied |
| A | api | 138.2.76.159 | Proxied |
| A | dashboard | 138.2.76.159 | Proxied |
| A | bidding | 138.2.76.159 | Proxied |
