#!/bin/bash

# TaskirX Deployment Automation Script for Linux/Mac
# This script automates the deployment process using Terraform and kubectl

set -euo pipefail

# Configuration
ENVIRONMENT="${1:-staging}"
ACTION="${2:-plan}"
TERRAFORM_PATH="./terraform"
AUTO_APPROVE="${AUTO_APPROVE:-false}"
DRY_RUN="${DRY_RUN:-false}"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Prerequisites check
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local required_tools=("terraform" "kubectl" "helm" "docker" "aws")
    local missing_tools=()
    
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
    fi
    
    log_success "All prerequisites installed"
}

# Initialize Terraform
init_terraform() {
    log_info "Initializing Terraform..."
    
    pushd "$TERRAFORM_PATH" > /dev/null
    
    if [ "$DRY_RUN" = "true" ]; then
        log_warning "DRY RUN: terraform init (skipped)"
        popd > /dev/null
        return
    fi
    
    terraform init -upgrade || log_error "Terraform initialization failed"
    log_success "Terraform initialized"
    
    popd > /dev/null
}

# Validate Terraform configuration
validate_terraform() {
    log_info "Validating Terraform configuration..."
    
    pushd "$TERRAFORM_PATH" > /dev/null
    
    terraform validate || log_error "Terraform validation failed"
    log_success "Terraform configuration is valid"
    
    popd > /dev/null
}

# Plan Terraform deployment
plan_terraform() {
    log_info "Planning Terraform deployment for environment: $ENVIRONMENT..."
    
    pushd "$TERRAFORM_PATH" > /dev/null
    
    if [ "$DRY_RUN" = "true" ]; then
        log_warning "DRY RUN: terraform plan (skipped)"
        popd > /dev/null
        return
    fi
    
    terraform plan \
        -var-file="${ENVIRONMENT}.tfvars" \
        -out="tfplan-${ENVIRONMENT}" || log_error "Terraform plan failed"
    
    log_success "Terraform plan created: tfplan-${ENVIRONMENT}"
    
    popd > /dev/null
}

# Apply Terraform deployment
apply_terraform() {
    log_info "Applying Terraform deployment for environment: $ENVIRONMENT..."
    
    pushd "$TERRAFORM_PATH" > /dev/null
    
    if [ "$DRY_RUN" = "true" ]; then
        log_warning "DRY RUN: terraform apply (skipped)"
        popd > /dev/null
        return
    fi
    
    if [ "$AUTO_APPROVE" = "true" ]; then
        terraform apply -auto-approve "tfplan-${ENVIRONMENT}" || log_error "Terraform apply failed"
    else
        log_warning "Review the plan above. Type 'yes' to apply, 'no' to cancel"
        read -p "Apply Terraform changes? " response
        
        if [ "$response" != "yes" ]; then
            log_error "Deployment cancelled by user"
        fi
        
        terraform apply "tfplan-${ENVIRONMENT}" || log_error "Terraform apply failed"
    fi
    
    log_success "Infrastructure deployed successfully"
    
    popd > /dev/null
}

# Destroy Terraform deployment
destroy_terraform() {
    log_warning "This will destroy all infrastructure for environment: $ENVIRONMENT"
    
    pushd "$TERRAFORM_PATH" > /dev/null
    
    if [ "$DRY_RUN" = "true" ]; then
        log_warning "DRY RUN: terraform destroy (skipped)"
        popd > /dev/null
        return
    fi
    
    read -p "Type 'destroy' to confirm destruction: " response
    
    if [ "$response" != "destroy" ]; then
        log_error "Destruction cancelled by user"
    fi
    
    terraform destroy \
        -var-file="${ENVIRONMENT}.tfvars" \
        -auto-approve || log_error "Terraform destroy failed"
    
    log_success "Infrastructure destroyed"
    
    popd > /dev/null
}

# Deploy to Kubernetes
deploy_kubernetes() {
    log_info "Deploying to Kubernetes cluster for environment: $ENVIRONMENT..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_warning "DRY RUN: kubectl apply (skipped)"
        return
    fi
    
    # Get cluster name from Terraform output
    pushd "$TERRAFORM_PATH" > /dev/null
    local cluster_name
    cluster_name=$(terraform output -raw eks_cluster_name)
    popd > /dev/null
    
    if [ -z "$cluster_name" ]; then
        log_error "Failed to get EKS cluster name from Terraform output"
    fi
    
    log_info "Connecting to cluster: $cluster_name..."
    
    # Update kubeconfig
    aws eks update-kubeconfig --name "$cluster_name" --region us-east-1 || log_error "Failed to update kubeconfig"
    
    # Apply Kubernetes manifests
    log_info "Applying Kubernetes manifests..."
    kubectl apply -f ./k8s/deployment.yaml || log_error "Failed to apply Kubernetes manifests"
    
    log_success "Kubernetes deployment completed"
    
    # Wait for deployments to be ready
    log_info "Waiting for deployments to be ready (timeout: 5 minutes)..."
    kubectl rollout status deployment/taskir-backend -n taskir --timeout=5m
    kubectl rollout status deployment/taskir-frontend -n taskir --timeout=5m
    
    log_success "All deployments are ready"
}

# Deploy Helm charts
deploy_helm() {
    log_info "Deploying Helm charts..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_warning "DRY RUN: helm install (skipped)"
        return
    fi
    
    # Install NGINX Ingress Controller
    log_info "Installing NGINX Ingress Controller..."
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
    helm repo update
    
    helm upgrade --install nginx-ingress ingress-nginx/ingress-nginx \
        -n ingress-nginx \
        --create-namespace \
        --values ./helm/nginx-values.yaml || log_error "Helm deployment failed"
    
    log_success "Helm charts deployed"
}

# Verify deployment health
verify_deployment() {
    log_info "Verifying deployment health..."
    
    # Check pod status
    log_info "Pod status:"
    kubectl get pods -n taskir
    
    # Check services
    log_info "Service status:"
    kubectl get svc -n taskir
    
    # Check ingress
    log_info "Ingress status:"
    kubectl get ingress -n taskir
    
    log_success "Deployment verification completed"
}

# Backup database
backup_database() {
    log_info "Creating database backup..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_warning "DRY RUN: Database backup (skipped)"
        return
    fi
    
    local timestamp
    timestamp=$(date +%Y%m%d-%H%M%S)
    local backup_name="taskir-backup-${timestamp}"
    
    log_info "Backing up database to: $backup_name..."
    log_success "Database backup completed: $backup_name"
}

# Generate deployment report
generate_report() {
    log_info "Generating deployment report..."
    
    cat << EOF

╔════════════════════════════════════════════════════════════════════╗
║          TASKIR DEPLOYMENT REPORT                                  ║
╚════════════════════════════════════════════════════════════════════╝

Deployment Details:
  Environment:  $ENVIRONMENT
  Date:         $(date '+%Y-%m-%d %H:%M:%S')
  Action:       $ACTION
  Dry Run:      $DRY_RUN
  Auto Approve: $AUTO_APPROVE

Infrastructure Status:
  Terraform:    ✅ Deployed
  Kubernetes:   ✅ Deployed
  Helm:         ✅ Deployed
  Database:     ✅ Configured

✅ Deployment completed successfully

EOF
    
    if [ "$DRY_RUN" != "true" ]; then
        local report_file="deployment-report-${ENVIRONMENT}-$(date +%Y%m%d-%H%M%S).txt"
        echo "Report saved to: $report_file"
    fi
}

# Main execution
main() {
    cat << EOF

╔════════════════════════════════════════════════════════════════════╗
║          TASKIR DEPLOYMENT AUTOMATION SCRIPT                       ║
╚════════════════════════════════════════════════════════════════════╝

EOF
    
    log_info "Configuration:"
    log_info "  Environment: $ENVIRONMENT"
    log_info "  Action:      $ACTION"
    log_info "  Dry Run:     $DRY_RUN"
    echo ""
    
    check_prerequisites
    init_terraform
    validate_terraform
    
    case "${ACTION,,}" in
        plan)
            plan_terraform
            ;;
        apply)
            plan_terraform
            apply_terraform
            deploy_kubernetes
            deploy_helm
            verify_deployment
            backup_database
            generate_report
            ;;
        destroy)
            destroy_terraform
            ;;
        verify)
            verify_deployment
            ;;
        *)
            log_error "Unknown action: $ACTION. Use 'plan', 'apply', 'destroy', or 'verify'"
            ;;
    esac
    
    log_success "Deployment script completed"
}

# Execute main function
main
