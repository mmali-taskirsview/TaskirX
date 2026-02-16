# TaskirX - OCI Registry Management

# 1. Login to OCIR
# docker login <region-code>.ocir.io
# Username: <tenancy-namespace>/<username>
# Password: <auth-token>

# 2. Created Repositories (Managed by Terraform):
# - taskir-nestjs
# - taskir-go-bidding
# - taskir-dashboard
# - taskir-ad-matching
# - taskir-fraud-detection
# - taskir-bid-optimization

# 3. Usage:
# .\scripts\oci-push-images.ps1 -Registry "<region-code>.ocir.io/<tenancy-namespace>"
