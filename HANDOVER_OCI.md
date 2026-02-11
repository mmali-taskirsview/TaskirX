# TaskirX OCI Migration Guide (Handover)

**Date**: February 11, 2026
**Status**: Ready for Deployment

## 1. Executive Summary
We have successfully transitioned the TaskirX "Polyglot" Ad Exchange architecture to be compatible with **Oracle Cloud Infrastructure (OCI)**. This move leverages OCI's cost-effective compute for the Kubernetes cluster (OKE), **Pinecone** for the AI vector search (Ad Matching), and **Cloudflare** for secure edge networking.

## 2. New Infrastructure Components
- **Terraform**: Complete OCI definition in `terraform-oci/`.
- **Kubernetes**: Updated manifests in `k8s/` using OCI storage classes and ingress controllers.
- **CI/CD**: GitHub Actions workflow (`.github/workflows/deploy-oci.yml`) for automated builds and deployment to OCIR/OKE.
- **Scripts**:
  - `deploy-to-oci.ps1`: One-command deployment.
  - `validate-oci-setup.ps1`: Pre-flight checks.
  - `seed-remote-db.ps1`: Populates remote Postgres/ClickHouse.
  - `get-ingress-ip.ps1`: Helps with DNS setup.

## 3. How to Deploy (Step-by-Step)

### Option A: The "GitOps" Way (Recommended)
1. Commit all your code to GitHub.
2. Add the following Secrets to your Repository:
   - `OCI_TENANCY_OCID`, `OCI_USER_OCID`, `OCI_FINGERPRINT`, `OCI_PRIVATE_KEY`, `OCI_REGION`
   - `OCI_AUTH_TOKEN` (for Docker registry)
   - `PINECONE_API_KEY`
   - `CLOUDFLARE_API_TOKEN`
3. Push to `main`. The pipeline will build images, setup the cluster, and deploy app.
4. Run `.\scripts\seed-remote-db.ps1` from your local machine (requires `kubectl` context to be set).

### Option B: The "Manual" Way (Local)
1. **Config**: Fill in `terraform-oci/terraform.tfvars`.
2. **Validate**: Run `.\scripts\validate-oci-setup.ps1`.
3. **Deploy**: Run `.\scripts\deploy-to-oci.ps1 -Action apply`.
4. **Seed**: Run `.\scripts\seed-remote-db.ps1`.
5. **DNS**: Run `.\scripts\get-ingress-ip.ps1` and update Cloudflare.

## 4. Operational Maintenance
- **Monitoring**: Access Grafana at `http://localhost:3000` (via port-forward).
- **Updates**: Change code -> Push to Git -> Auto-Deploy.
- **Scaling**: Terraform controls node pool size (`oke.tf`), K8s HPA controls pod count.

## 5. Known Verification Steps
- `https://api.taskir.com/health` should return `{"status": "ok"}`.
- `https://dashboard.taskir.com` should load the login page.

### 4. Live System Verification
Once deployed, use the verification tool to check if all public endpoints are reachable and healthy:

```powershell
.\scripts\verify-live-endpoints.ps1
```

If the system is healthy, you will see `OPEN (200 OK)` for all services.

---
**Signed Off By**: GitHub Copilot
