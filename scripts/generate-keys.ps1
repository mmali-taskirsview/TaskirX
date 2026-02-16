# Generate OCI API Key Pair

$KeyDir = [System.IO.Path]::Combine($env:USERPROFILE, ".oci")
$KeyName = "oci_api_key"
$PrivPath = [System.IO.Path]::Combine($KeyDir, "$KeyName.pem")
$PubPath = [System.IO.Path]::Combine($KeyDir, "${KeyName}_public.pem")

# 1. Create directory
if (-not (Test-Path $KeyDir)) {
    New-Item -ItemType Directory -Force -Path $KeyDir | Out-Null
    Write-Host "Created directory: $KeyDir" -ForegroundColor Green
}

# 2. Check for OpenSSL
if (-not (Get-Command "openssl" -ErrorAction SilentlyContinue)) {
    Write-Host "Error: OpenSSL not found. Please install Git for Windows or OpenSSL." -ForegroundColor Red
    exit 1
}

# 3. Generate Keys
Write-Host "Generating 2048-bit RSA key pair..." -ForegroundColor Cyan

# Private Key
& openssl genrsa -out "$PrivPath" 2048

# Public Key
& openssl rsa -pubout -in "$PrivPath" -out "$PubPath"

# 4. Output Instructions
Write-Host "`nKeys Generated Successfully!" -ForegroundColor Green
Write-Host "Private Key: $PrivPath"
Write-Host "Public Key:  $PubPath"

Write-Host "`nNEXT STEPS:" -ForegroundColor Yellow
Write-Host "1. Open the content of the PUBLIC key:"
Write-Host "   Get-Content `"$PubPath`""
Write-Host "2. Go to OCI Console -> User Settings -> API Keys -> Add API Key"
Write-Host "3. Select 'Paste Public Key' and paste the content."
Write-Host "4. Copy the Fingerprint, User OCID, and Tenancy OCID shown there."
Write-Host "5. Update 'c:\TaskirX\terraform-oci\terraform.tfvars' with those values."
Write-Host "6. Update 'private_key_path' in tfvars to: `"$PrivPath`""
