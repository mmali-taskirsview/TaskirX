# 🚀 Phase 5: Post-Launch Operations & Scaling

**Current Status:** The platform is **LIVE** at `TaskirX.com`.
**Cluster:** OKE (Oracle Kubernetes Engine)
**Registry:** Docker Hub (`taskirsview`)

## 1. Immediate Post-Launch Actions
- [x] **DNS Verification**: Confirm `api.taskirx.com` resolves to `138.2.76.159`.
- [x] **SSL Certification**: Verify `https://` access (handled automatically by cert-manager).
- [x] **User Onboarding**: Seeding complete.
    - **Admin**: `admin@taskirx.com` / `Admin123!` (or `Test123!` if hash collision)
    - **Advertiser**: `advertiser@test.com` / `Test123!`
    - *Note: Change passwords immediately upon first login.*

## 2. Monitoring & Observability
Now that the system is live, we need to see what it handles.
- [x] **Prometheus & Grafana**: Deployed to `monitoring` namespace.
- [x] **Grafana Dashboard**: Connect Grafana to Prometheus to visualize:
    - Bidding QPS (Queries Per Second).
    - AI Inference Latency (ms).
    - Redis Cache Hit Rates.
    - Fraud Detection Rate.
    - Backend Traffic.
- [x] **Setup Script**: Created `scripts/setup-monitoring-dashboards.ps1`.
- [x] **Access Script**: Created `scripts/port-forward-monitoring.ps1`.
- [x] **Log Aggregation**: Deployed Loki & Promtail.
    - Added "Application Logs" panel to Grafana Dashboard.
    - All microservices logs are now searchable from one place.

## 3. Scaling Strategy
- [x] **Horizontal Pod Autoscaling (HPA)**: Configured and Active.
    - `go-bidding`: Scales 2-10 pods (CPU > 70%).
    - `ad-matching`: Scales 2-5 pods (CPU > 60%).
    - `nestjs-backend`: Scales 2-5 pods (CPU > 70%).
- [x] **Database Optimization**:
    - **Redis**: Configured `maxmemory 400mb` and `allkeys-lru` eviction policy.
    - **ClickHouse**: Enabled `async_insert=1` via ConfigMap for high-ingestion performance.

## 4. Completed Phases (Formerly Roadmap)
- [x] **Phase 5**: Performance Optimization (Redis Caching).
- [x] **Phase 6**: Advanced Targeting & Header Bidding.
    - Geo-fencing implemented.
    - Prebid.js Adapter created (`sdks/javascript`).
    - Video/Native Ad support added.
- [x] **Phase 7**: AI Service Integration & Persistence.

## 5. Maintenance & Operations
- **Deploy Updates**: Use `.\update-domain.ps1` to deploy code changes.
- **View Logs**: Use `kubectl logs` or the Grafana dashboard.
- **Backups**: Schedule nightly backups of Postgres and ClickHouse.
- **Scale**: Adjust HPA thresholds in `k8s/hpa.yaml`.
