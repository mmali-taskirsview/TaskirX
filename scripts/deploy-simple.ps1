# Simple OCI Deployment Script
# Helper for unreliable networks or when the advanced script fails

param(
    [string]$Action = "apply"
)

$ErrorActionPreference = "Stop"

# 1. Terraform
Write-Host "Running Terraform..." -ForegroundColor Cyan
cd terraform-oci

if ($Action -eq "destroy") {
    terraform destroy -auto-approve
} else {
    # Try init loop for flaky connection
    $maxRetries = 5
    $retryCount = 0
    $success = $false

    while (-not $success -and $retryCount -lt $maxRetries) {
        try {
            terraform init
            $success = $true
        } catch {
            Write-Warning "Terraform init failed. Retrying... ($retryCount/$maxRetries)"
            $retryCount++
            Start-Sleep -Seconds 2
        }
    }

    if (-not $success) {
        Write-Error "Terraform init failed after $maxRetries attempts."
        exit 1
    }

    if ($Action -eq "plan") {
        terraform plan
    } else {
        terraform apply -auto-approve
    }
}

cd ..

# 2. Kubernetes (Only on Apply)
if ($Action -eq "apply") {
    Write-Host "`nRunning Kubernetes Deployment..." -ForegroundColor Cyan
    
    # Generate Kubeconfig (needs OCI CLI working)
    # This step assumes terraform successfully created the cluster and updated local kubeconfig?
    # No, Terraform OCI provider doesn't auto-update kubeconfig usually.
    # We need to run the OCI CLI command.
    
    # We can get the cluster ID from terraform output
    $clusterId = (terraform -chdir=terraform-oci output -raw cluster_id) 
    $region = "us-ashburn-1" # Hardcoded for now based on tfvars default, ideally parsed
    
    if ($clusterId) {
       Write-Host "Updating kubeconfig for Cluster: $clusterId"
       oci ce cluster create-kubeconfig --cluster-id $clusterId --file $HOME/.kube/config --region $region --token-version 2.0.0 --kube-endpoint PUBLIC_ENDPOINT
    }

    $manifests = @(
        "k8s/namespace.yaml",
        "k8s/postgres-deployment.yaml",
        "k8s/redis-deployment.yaml",
        "k8s/python-services-deployment.yaml",
        "k8s/nestjs-deployment.yaml",
        "k8s/go-bidding-deployment.yaml",
        "k8s/next-dashboard-deployment.yaml",
        "k8s/ingress-oci.yaml"
    )

    foreach ($file in $manifests) {
        if (Test-Path $file) {
            kubectl apply -f $file
        } else {
            Write-Warning "$file not found"
        }
    }
}
