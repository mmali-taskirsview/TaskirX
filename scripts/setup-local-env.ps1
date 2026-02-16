# Setup Local Dev Environment for TaskirX
# Installs missing tools locally to ./bin and configures the shell

$ErrorActionPreference = "Stop"
$WorkDir = Get-Location
$BinDir = Join-Path $WorkDir "bin"

# 1. Create local bin directory
if (-not (Test-Path $BinDir)) {
    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
    Write-Host "Created local bin directory: $BinDir" -ForegroundColor Green
}

# 2. Install Terraform
$TerraformVersion = "1.5.7" # Stable version
$TerraformUrl = "https://releases.hashicorp.com/terraform/$TerraformVersion/terraform_${TerraformVersion}_windows_amd64.zip"
$TerraformExe = Join-Path $BinDir "terraform.exe"

if (-not (Test-Path $TerraformExe)) {
    Write-Host "Downloading Terraform $TerraformVersion..." -ForegroundColor Cyan
    $ZipPath = Join-Path $BinDir "terraform.zip"
    Invoke-WebRequest -Uri $TerraformUrl -OutFile $ZipPath
    Expand-Archive -Path $ZipPath -DestinationPath $BinDir -Force
    Remove-Item $ZipPath
    Write-Host "Terraform installed." -ForegroundColor Green
} else {
    Write-Host "Terraform already installed." -ForegroundColor Gray
}

# 3. Install Helm
$HelmVersion = "v3.12.3"
$HelmUrl = "https://get.helm.sh/helm-$HelmVersion-windows-amd64.zip"
$HelmExe = Join-Path $BinDir "helm.exe"

if (-not (Test-Path $HelmExe)) {
    Write-Host "Downloading Helm $HelmVersion..." -ForegroundColor Cyan
    $ZipPath = Join-Path $BinDir "helm.zip"
    Invoke-WebRequest -Uri $HelmUrl -OutFile $ZipPath
    Expand-Archive -Path $ZipPath -DestinationPath $BinDir -Force
    # Helm zip extracts to a subfolder
    $ExtractedHelm = Join-Path $BinDir "windows-amd64\helm.exe"
    Move-Item -Path $ExtractedHelm -Destination $BinDir -Force
    Remove-Item -Path (Join-Path $BinDir "windows-amd64") -Recurse -Force
    Remove-Item $ZipPath
    Write-Host "Helm installed." -ForegroundColor Green
} else {
    Write-Host "Helm already installed." -ForegroundColor Gray
}

# 4. Install OCI CLI via Pip
if (-not (Get-Command "oci" -ErrorAction SilentlyContinue)) {
    Write-Host "Installing OCI CLI via pip..." -ForegroundColor Cyan
    pip install oci-cli
    if ($LASTEXITCODE -eq 0) {
        Write-Host "OCI CLI installed." -ForegroundColor Green
    } else {
        Write-Host "Failed to install OCI CLI via pip." -ForegroundColor Red
    }
} else {
    Write-Host "OCI CLI already installed." -ForegroundColor Gray
}

# 5. Add bin to PATH for this session
$env:PATH = "$BinDir;$env:PATH"
Write-Host "Added $BinDir to PATH for this session." -ForegroundColor Yellow

# 6. Verify
Write-Host "`nVerifying tools:" -ForegroundColor Cyan
Get-Command terraform, helm, oci | Select-Object Name, Source

Write-Host "`nSetup Complete!" -ForegroundColor Green
Write-Host "NOTE: You must re-run this script or manually add $BinDir to your PATH if you open a new terminal." -ForegroundColor Magenta
