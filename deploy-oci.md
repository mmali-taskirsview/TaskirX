# Deploy Infrastructure to OCI

This script wraps the standard `deploy.ps1` but pre-configures it for OCI (Oracle Cloud Infrastructure).

## Prerequisites
1. [OCI CLI](https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliinstall.htm) installed and configured (`oci setup config`).
2. Terraform installed.
3. `terraform-oci/terraform.tfvars` populated with your tenancy details.

## Usage

```powershell
.\scripts\deploy-infrastructure-oci.ps1 -Registry "iad.ocir.io/mytenancy/taskir"
```

## Parameters
- `-Registry`: The base path for your OCIR registry (e.g., `iad.ocir.io/tenancy/taskir`).
- `-Action`: `plan` (default), `apply`, or `destroy`.

## Example
```powershell
.\scripts\deploy-infrastructure-oci.ps1 -Action apply -Registry "us-ashburn-1.ocir.io/mytnant/taskir"
```

# Notes

3. Cloudflare Tunnel/Ingress
    - The 'k8s/ingress-oci.yaml' requires an Nginx Ingress Controller.
    - OCI OKE can install this via Helm.
    - Cloudflare DNS records will point to the External IP of this Nginx Controller.
    - Ensure your 'terraform-oci/terraform.tfvars' has the correct Cloudflare credentials.

4. Pinecone Configuration
    - Pinecone 'us-east-1' (AWS) is configured as the default region for Serverless indexes.
    - Ensure your API key has permissions to create indexes in this region.

5. Storage Classes
    - The default storage class 'gp3' (AWS EBS) has been replaced with 'oci-bv' (Oracle Block Volume) in K8s manifests.
    - Ensure the 'oci-bv' StorageClass is present in your cluster (it is default in OKE).
    - If using a custom StorageClass, update 'k8s/*-deployment.yaml' files.

6. Image Pull Secrets
    - If using a private OCIR repository, you must create a secret in the 'taskir' namespace:
      ```bash
      kubectl create secret docker-registry ocir-secret \
        --docker-server=<region>.ocir.io \
        --docker-username=<tenancy>/<username> \
        --docker-password='<auth_token>' \
        -n taskir
      ```
    - Then patch your default service account:
      ```bash
      kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "ocir-secret"}]}' -n taskir
      ```
    - CI/CD (`deploy-oci.yml`) handles this automatically.

7. SSL Certificates (Let's Encrypt)
    - The deployment script automatically installs `cert-manager`.
    - It then applies `k8s/cluster-issuer.yaml`.
    - **Important**: Update the email in `k8s/cluster-issuer.yaml` to your own email address to receive expiry notifications.

8. Building & Pushing Images Manually
    If automated CI/CD is not possible (e.g., local testing), use the helper script to build and push all images to OCIR:

    ```powershell
    .\scripts\oci-push-images.ps1 -Registry "iad.ocir.io/tenancy/taskir"
    ```

    Then deploy using the same registry:
    ```powershell
    .\scripts\deploy-to-oci.ps1 -Registry "iad.ocir.io/tenancy/taskir"
    ```

### 4. DNS Configuration (Critical)

After the deployment finishes, OCI will assign a public IP ADdress to your Load Balancer.
The deployment script will attempt to display this IP at the end.

If you missed it, run:
```powershell
.\scripts\get-ingress-ip.ps1
```

**Action**: Login to Cloudflare and update the A records for `api`, `dashboard`, and `bidding` to this IP.

## Validation
Before deploying, run the validation script to ensure your environment is configured correctly:
```powershell
.\scripts\validate-oci-setup.ps1
```
