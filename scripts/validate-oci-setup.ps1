# TaskirX OCI Setup Validator
# Checks if the local environment is ready for OCI deployment

function Write-Status {
    param($Item, $Status, $Details)
    $color = if ($Status -eq "OK") { "Green" } else { "Red" }
    Write-Host ("[{0}] {1}: {2}" -f $Status, $Item, $Details) -ForegroundColor $color
}

Write-Host "Validating OCI Deployment Environment..." -ForegroundColor Cyan

# 1. Check Tools
$tools = @("terraform", "oci", "kubectl", "docker", "helm")
foreach ($tool in $tools) {
    if (Get-Command $tool -ErrorAction SilentlyContinue) {
        Write-Status -Item $tool -Status "OK" -Details "Installed"
    } else {
        Write-Status -Item $tool -Status "MISSING" -Details "Please install $tool"
    }
}

# 2. Check Directory Structure
$dirs = @("terraform-oci", "k8s", "scripts")
foreach ($dir in $dirs) {
    if (Test-Path $dir) {
        Write-Status -Item "Dir: $dir" -Status "OK" -Details "Found"
    } else {
        Write-Status -Item "Dir: $dir" -Status "MISSING" -Details "Directory not found"
    }
}

# 3. Check Terraform Variables
$tfvars = "terraform-oci/terraform.tfvars"
if (Test-Path $tfvars) {
    $content = Get-Content $tfvars -Raw
    if ($content -match "ocid1.tenancy.oc1..aaaaaaa") {
        Write-Status -Item "Config: tfvars" -Status "WARNING" -Details "Detected placeholder values in $tfvars. Please update with real credentials."
    } else {
         Write-Status -Item "Config: tfvars" -Status "OK" -Details "Generic placeholders not found (good sign)"
    }
} else {
    Write-Status -Item "Config: tfvars" -Status "MISSING" -Details "Create $tfvars from terraform.tfvars.example or variables.tf defaults"
}

# 4. Check API Keys
$ociKeyPath = "$HOME/.oci/oci_api_key.pem" # Common default
if (Test-Path "terraform-oci/terraform.tfvars") {
    # Try to extract key path from tfvars if possible (basic regex)
    $tfvarsContent = Get-Content "terraform-oci/terraform.tfvars" -Raw
    if ($tfvarsContent -match 'private_key_path\s*=\s*"([^"]+)"') {
        $ociKeyPath = $matches[1]
    }
}

if ($orgKeyPath -and (Test-Path $ociKeyPath)) {
    Write-Status -Item "OCI Key" -Status "OK" -Details "Found at $ociKeyPath"
} else {
    Write-Status -Item "OCI Key" -Status "WARNING" -Details "Not found at $ociKeyPath. Ensure your terraform.tfvars points to a valid pem file."
}

Write-Host "`nValidation Complete. Fix any MISSING items before running deploy-to-oci.ps1." -ForegroundColor Cyan
