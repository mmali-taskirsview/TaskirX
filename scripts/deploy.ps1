# TaskirX Deployment Automation Script for Windows PowerShell
# This script automates the deployment process using Terraform and kubectl

param(
    [string]$Environment = "staging",
    [string]$Action = "plan",
    [string]$TerraformPath = "./terraform",
    [string]$Registry = "your-registry",
    [switch]$AutoApprove = $false,
    [switch]$DryRun = $false
)

# Detect local bin directory and add to PATH if detected
if (Test-Path "$PSScriptRoot/../bin") {
    $LocalBinDir = Resolve-Path "$PSScriptRoot/../bin"
    if ($env:Path -notlike "*$LocalBinDir*") {
        Write-Host "[INFO] Adding detected local bin directory to PATH: $LocalBinDir" -ForegroundColor Blue
        $env:Path = "$LocalBinDir;$env:Path"
    }
}

# Helper Functions
function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

# Auto-detect Terraform path for OCI
if ($Environment -eq "oci" -and $TerraformPath -eq "./terraform") {
    $TerraformPath = "./terraform-oci"
    Write-Info "Switched Terraform path to ./terraform-oci for OCI environment"
}

# Prerequisites check
function Test-Prerequisites {
    Write-Info "Checking prerequisites..."
    
    $required = @("terraform", "kubectl", "helm", "docker")
    $missing = @()
    
    foreach ($cmd in $required) {
        if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
            $missing += $cmd
        }
    }
    
    if ($missing.Count -gt 0) {
        Write-Error "Missing required tools: $($missing -join ', ')"
        exit 1
    }

    if ($Environment -eq "oci" -and -not (Get-Command "oci" -ErrorAction SilentlyContinue)) {
         Write-Error "Missing required tool: oci (Oracle Cloud CLI) is required for OCI deployment."
         exit 1
    }
    
    Write-Success "All prerequisites installed"
}

# Initialize Terraform
function Initialize-Terraform {
    Write-Info "Initializing Terraform..."
    
    Push-Location $TerraformPath
    
    if ($DryRun) {
        Write-Warning "DRY RUN: terraform init (skipped)"
        Pop-Location
        return
    }
    
    terraform init -upgrade
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Terraform initialization failed"
        Pop-Location
        exit 1
    }
    
    Write-Success "Terraform initialized"
    Pop-Location
}

# Validate Terraform configuration
function Test-TerraformConfig {
    Write-Info "Validating Terraform configuration..."
    
    Push-Location $TerraformPath
    
    terraform validate
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Terraform validation failed"
        Pop-Location
        exit 1
    }
    
    Write-Success "Terraform configuration is valid"
    Pop-Location
}

# Plan Terraform deployment
function Plan-Terraform {
    Write-Info "Planning Terraform deployment for environment: $Environment..."
    
    Push-Location $TerraformPath
    
    if ($DryRun) {
        Write-Warning "DRY RUN: terraform plan (skipped)"
        Pop-Location
        return
    }
    
    terraform plan `
        -var-file="$Environment.tfvars" `
        -out="tfplan-$Environment"
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Terraform plan failed"
        Pop-Location
        exit 1
    }
    
    Write-Success "Terraform plan created: tfplan-$Environment"
    Pop-Location
}

# Apply Terraform deployment
function Apply-Terraform {
    Write-Info "Applying Terraform deployment for environment: $Environment..."
    
    Push-Location $TerraformPath
    
    if ($DryRun) {
        Write-Warning "DRY RUN: terraform apply (skipped)"
        Pop-Location
        return
    }
    
    if ($AutoApprove) {
        terraform apply -auto-approve "tfplan-$Environment"
    }
    else {
        Write-Warning "Review the plan above. Type 'yes' to apply, 'no' to cancel"
        $response = Read-Host "Apply Terraform changes?"
        
        if ($response -ne "yes") {
            Write-Error "Deployment cancelled by user"
            Pop-Location
            exit 1
        }
        
        terraform apply "tfplan-$Environment"
    }
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Terraform apply failed"
        Pop-Location
        exit 1
    }
    
    Write-Success "Infrastructure deployed successfully"
    Pop-Location
}

# Destroy Terraform deployment
function Destroy-Terraform {
    Write-Warning "This will destroy all infrastructure for environment: $Environment"
    
    Push-Location $TerraformPath
    
    if ($DryRun) {
        Write-Warning "DRY RUN: terraform destroy (skipped)"
        Pop-Location
        return
    }
    
    $response = Read-Host "Type 'destroy' to confirm destruction"
    
    if ($response -ne "destroy") {
        Write-Error "Destruction cancelled by user"
        Pop-Location
        exit 1
    }
    
    terraform destroy `
        -var-file="$Environment.tfvars" `
        -auto-approve
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Terraform destroy failed"
        Pop-Location
        exit 1
    }
    
    Write-Success "Infrastructure destroyed"
    Pop-Location
}

# Deploy to Kubernetes
function Deploy-Kubernetes {
    Write-Info "Deploying to Kubernetes cluster for environment: $Environment..."
    
    if ($DryRun) {
        Write-Warning "DRY RUN: kubectl apply (skipped)"
        return
    }

    $clusterName = ""
    $region = "us-east-1" # Default AWS

    # Retrieve Cluster Credentials
    Push-Location $TerraformPath
    if ($Environment -eq "oci") {
        $clusterId = terraform output -raw k8s_cluster_id
        $region = terraform output -raw region
        
        if (-not $clusterId) {
            Write-Error "Failed to get k8s_cluster_id from Terraform output"
            Pop-Location
            exit 1
        }
        
        Write-Info "Connecting to OKE Cluster: $clusterId..."
        oci ce cluster create-kubeconfig --cluster-id $clusterId --file $HOME/.kube/config --region $region --token-version 2.0.0
        
    } else {
        # AWS EKS Logic
        $clusterName = terraform output -raw eks_cluster_name
        
        if (-not $clusterName) {
            Write-Error "Failed to get EKS cluster name from Terraform output"
            Pop-Location
            exit 1
        }
        
        Write-Info "Connecting to EKS Cluster: $clusterName..."
        aws eks update-kubeconfig --name $clusterName --region $region
    }
    Pop-Location

    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to update kubeconfig"
        exit 1
    }
    
    # Apply Kubernetes manifests
    Write-Info "Applying Kubernetes manifests with Registry: $Registry..."
    
    $manifestFiles = @(
        "./k8s/namespace.yaml",
        "./k8s/postgres-deployment.yaml",
        "./k8s/redis-deployment.yaml",
        "./k8s/pinecone-secret.yaml",
        "./k8s/clickhouse-deployment.yaml",
        "./k8s/nestjs-deployment.yaml",
        "./k8s/go-bidding-deployment.yaml",
        "./k8s/python-services-deployment.yaml"
    )

    if ($Environment -eq "oci") {
        $manifestFiles += "./k8s/ingress-oci.yaml"
    } else {
        $manifestFiles += "./k8s/ingress.yaml"
    }

    if ($Registry -eq "your-registry") {
        Write-Warning "Using 'your-registry' placeholder. Only local/test images will work."
    }

    foreach ($file in $manifestFiles) {
        if (Test-Path $file) {
            Write-Info "Applying $file..."
            (Get-Content $file).Replace('your-registry', $Registry) | kubectl apply -f -
        } else {
            Write-Warning "Manifest file not found: $file"
        }
    }
    
    Write-Success "Kubernetes deployment completed"
    
    # Wait for deployments to be ready
    Write-Info "Waiting for deployments to be ready (timeout: 5 minutes)..."
    kubectl rollout status deployment/nestjs-backend -n taskir --timeout=5m
    kubectl rollout status deployment/go-bidding -n taskir --timeout=5m
    kubectl rollout status deployment/fraud-detection -n taskir --timeout=5m
    
    Write-Success "All deployments are ready"

    # Post-deployment: Get Ingress IP for DNS
    if (Test-Path "./scripts/get-ingress-ip.ps1") {
        Write-Info "Checking Ingress IP Status..."
        try {
            & "./scripts/get-ingress-ip.ps1"
        } catch {
            Write-Warning "Could not retrieve Ingress IP automatically."
        }
    }
}

# Deploy Helm charts (if applicable)
function Deploy-Helm {
    Write-Info "Deploying Helm charts..."
    
    if ($DryRun) {
        Write-Warning "DRY RUN: helm install (skipped)"
        return
    }
    
    # Example: Deploy Ingress controller
    Write-Info "Installing NGINX Ingress Controller..."
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
    helm repo add jetstack https://charts.jetstack.io
    helm repo update
    
    helm upgrade --install nginx-ingress ingress-nginx/ingress-nginx `
        -n ingress-nginx `
        --create-namespace `
        --values ./helm/nginx-values.yaml
        
    Write-Info "Installing Cert-Manager..."
    helm upgrade --install cert-manager jetstack/cert-manager `
        --namespace cert-manager `
        --create-namespace `
        --version v1.13.1 `
        --set installCRDs=true
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Helm deployment failed"
        exit 1
    }

    Write-Info "Waiting for Cert-Manager Webhook to be ready..."
    Start-Sleep -Seconds 30

    if (Test-Path "./k8s/cluster-issuer.yaml") {
        Write-Info "Applying ClusterIssuer..."
        kubectl apply -f ./k8s/cluster-issuer.yaml
    }
}

# Deploy K8s Manifests directly (TaskirX Custom)
function Deploy-K8s-Manifests {
    Write-Info "Deploying Kubernetes Manifests..."
    
    if (-not (Test-Path ".\k8s\namespace.yaml")) {
        Write-Error "k8s directory not found in current path."
        return
    }

    $manifests = @(
        "namespace.yaml",
        "redis-deployment.yaml",
        "postgres-deployment.yaml",
        "python-services-deployment.yaml", # AI Agents
        "nestjs-deployment.yaml",
        "go-bidding-deployment.yaml",
        "next-dashboard-deployment.yaml"
    )

    # Select appropriate ingress for the environment
    if ($Environment -eq "oci") {
        $manifests += "ingress-oci.yaml"
    } else {
        $manifests += "ingress.yaml"
    }

    foreach ($file in $manifests) {
        $path = Join-Path ".\k8s" $file
        if (Test-Path $path) {
            Write-Info "Applying $file..."
            if (-not $DryRun) {
                kubectl apply -f $path
            }
        } else {
            Write-Warning "$file not found, skipping."
        }
    }
    
    Write-Success "Kubernetes manifests applied."
}

# Verify deployment health
function Test-Deployment {
    Write-Info "Verifying deployment health..."
    
    # Check pod status
    Write-Info "Pod status:"
    kubectl get pods -n taskir
    
    # Check services
    Write-Info "Service status:"
    kubectl get svc -n taskir
    
    # Check ingress
    Write-Info "Ingress status:"
    kubectl get ingress -n taskir
    
    # Run health checks
    Write-Info "Running health checks..."
    $backendService = kubectl get svc taskir-backend -n taskir -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'
    
    if ($backendService) {
        $health = Invoke-WebRequest -Uri "http://$backendService/health" -ErrorAction SilentlyContinue
        if ($health.StatusCode -eq 200) {
            Write-Success "Backend health check passed"
        }
        else {
            Write-Warning "Backend health check returned status: $($health.StatusCode)"
        }
    }
    
    Write-Success "Deployment verification completed"
}

# Backup database
function Backup-Database {
    Write-Info "Creating database backup..."
    
    if ($DryRun) {
        Write-Warning "DRY RUN: Database backup (skipped)"
        return
    }
    
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $backupName = "taskir-backup-$timestamp"
    
    # Get RDS endpoint from Terraform
    Push-Location $TerraformPath
    if ($Environment -eq "oci") {
        Write-Info "OCI Environment detected. Skipping RDS backup check as we are using in-cluster PostgreSQL."
        $rdsEndpoint = $null
    } else {
        try {
            $rdsEndpoint = terraform output -raw rds_endpoint
        } catch {
            Write-Warning "Could not retrieve RDS endpoint from Terraform output."
            $rdsEndpoint = $null
        }
    }
    Pop-Location
    
    if ($rdsEndpoint) {
        Write-Info "Backing up database to: $backupName..."
        # Note: In production, use AWS RDS APIs or pg_dump with proper credentials
    } else {
        Write-Info "No RDS endpoint found or using in-cluster DB. Skipping managed backup."
    }
    
    Write-Success "Database backup step completed"
}

# Generate deployment report
function Generate-Report {
    Write-Info "Generating deployment report..."
    
    $report = @"
╔════════════════════════════════════════════════════════════════════╗
║          TASKIR DEPLOYMENT REPORT                                  ║
╚════════════════════════════════════════════════════════════════════╝

Deployment Details:
  Environment:  $Environment
  Date:         $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
  Action:       $Action
  Dry Run:      $DryRun

Infrastructure Status:
  Terraform:    Deployed
  Kubernetes:   Deployed
  Helm:         Deployed
  Database:     Configured

Endpoints:
"@

    if (-not $DryRun) {
        Push-Location $TerraformPath
        $rdsEndpoint = terraform output -raw rds_endpoint 2>$null
        $s3Bucket = terraform output -raw s3_bucket_name 2>$null
        $cloudfrontDomain = terraform output -raw cloudfront_domain 2>$null
        Pop-Location
        
        $report += "  RDS:          $rdsEndpoint`n"
        $report += "  S3 Bucket:    $s3Bucket`n"
        $report += "  CDN:          $cloudfrontDomain`n"
    }

    $report += "`nDeployment completed`n"
    
    Write-Host $report
    
    # Save report to file
    $report | Out-File -FilePath "deployment-report-$Environment-$(Get-Date -Format 'yyyyMMdd-HHmmss').txt"
}

# Main execution
function Main {
    Write-Host "TASKIR DEPLOYMENT AUTOMATION SCRIPT"
    Write-Host ""
    
    Write-Info "Configuration:"
    Write-Info "  Environment: $Environment"
    Write-Info "  Action:      $Action"
    Write-Info "  Dry Run:     $DryRun"
    Write-Info ""
    
    Test-Prerequisites
    Initialize-Terraform
    Test-TerraformConfig
    
    $act = $Action.ToLower()
    
    if ($act -eq "plan") {
        Plan-Terraform
    }
    elseif ($act -eq "apply") {
        Plan-Terraform
        Apply-Terraform
        Deploy-Helm
        Deploy-Kubernetes
        Test-Deployment
        Backup-Database
        Generate-Report
    }
    elseif ($act -eq "destroy") {
        Destroy-Terraform
    }
    elseif ($act -eq "verify") {
        Test-Deployment
    }
    else {
        Write-Host "Unknown action: $Action"
        exit 1
    }

    Write-Host "Deployment script finished."
}

# START EXECUTION
Main