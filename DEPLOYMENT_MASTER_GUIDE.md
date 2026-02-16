# ADX Platform Deployment Master Guide

This guide details the deployment of the Advertising Exchange (ADX) platform using Oracle Cloud Infrastructure (OCI), Cloudflare, and Pinecone.

## 1. Architecture Overview

**Traffic Flow:**
1.  **User Request:** Ad request originates from a user's browser/device.
2.  **Cloudflare Edge:** Request hits Cloudflare.
    *   **DDoS Protection:** Filters malicious traffic.
    *   **WAF:** Inspects for exploits.
    *   **Edge Workers:** (Optional) Geo-targeting & basic validation logic runs here.
3.  **OCI Load Balancer:** Valid traffic is proxied to the OCI Public Load Balancer.
    *   *Security:* Accepts traffic ONLY from Cloudflare IPs.
4.  **OKE Cluster (Kubernetes):**
    *   **Ingress Controller:** Routes to specific services (Bidding Engine, Ad Matching).
    *   **Microservices:** Go/Node.js/Python containers process the request.
5.  **Data Layer:**
    *   **Pinecone:** Vector search for ad-user similarity (Real-time).
    *   **PostgreSQL/Autonomous DB:** Transactional data (User profiles, budgets).
    *   **Redis:** Caching hot data (Session state).

## 2. Infrastructure as Code (Terraform)

### Prerequisites
*   OCI CLI installed & authenticated.
*   Terraform installed.
*   Cloudflare API Token.
*   Pinecone API Key.

### Structure
The `terraform-oci/` directory contains the definition for the entire stack.

*   `network.tf`: VCN, Subnets (Public/Private), NAT Gateway, Service Gateway.
*   `oke.tf`: Managed Kubernetes Cluster & Node Pools.
*   `loadbalancer.tf`: Public Load Balancer listening on 80/443.
*   `security.tf`: NSGs restricted to Cloudflare IPs.

## 3. Cloudflare Configuration

We use Terraform (or manual setup) to configure:
*   **DNS A Record:** `@` points to OCI Load Balancer Public IP.
*   **SSL/TLS:** set to "Full (Strict)".
*   **Origin Ca Certificate:** Installed on OCI Load Balancer to ensure end-to-end encryption.

## 4. Pinecone Integration

The `ad-matching-service` uses Pinecone for vector similarity.
*   **Index Name:** `ad-vectors`
*   **Dimension:** 1536 (or matching your embedding model).
*   **Metric:** Cosine Similarity.

## 5. Application Deployment

### Kubernetes (Recommended)
1.  Build Docker images.
2.  Push to OCI Registry (OCIR).
3.  Deploy using Helm/Manifests in `k8s/`.

### VM-Based (Alternative)
Use the `scripts/bootstrap_adx.sh` script to provision a raw Compute Instance. This handles dependencies, code checkout, and PM2 setup.

## 6. Security Best Practices

1.  **Restrict Ingress:** Use the `Cloudflare-Only-Ingress` NSG (defined in `terraform-oci/security_cloudflare.tf`) on your OCI Load Balancer to drop all traffic not originating from Cloudflare.
2.  **Secrets Management:** Store sensitive API keys (Pinecone, Database) in **OCI Vault** and inject them as environment variables or K8s Secrets at runtime. Do NOT commit `.env` files.
3.  **DDoS Protection:** Enable Cloudflare's "Under Attack Mode" during active threats. Rate limit API endpoints `/api/v1/bid` in Cloudflare WAF rules.

## 7. Testing and Go-Live Checklist

1.  **Infrastructure Verification:**
    *   Run `terraform apply` in `terraform-oci/`.
    *   Confirm OKE Cluster is Active.
    *   Confirm Load Balancer has a Public IP.

2.  **Connectivity Check:**
    *   Get LB IP context: `kubectl get svc adx-ingress-service`.
    *   Update Cloudflare DNS `A` record with this IP.
    *   `curl -I https://your-domain.com`. Verify `Server: cloudflare`.

3.  **Pinecone Test:**
    *   Run `python3 python-ai-agents/pinecone_example.py` to upsert sample vectors and query them.

4.  **Load Test:**
    *   Use `locust -f performance-tests/locustfile.py` to simulate ad traffic.
    *   Monitor latency in Cloudflare Analytics and OCI Metrics.

5.  **Go-Live:**
    *   Lower TTL on DNS records.
    *   Enable Cloudflare "Orange Cloud" (Proxy).
    *   Monitor 5xx errors.
