# TaskirX - OCI Key Generator
# Generates a 2048-bit RSA key pair for Oracle Cloud Infrastructure API

$OpenSSLPath = "C:\Program Files\Git\usr\bin\openssl.exe"

if (-not (Test-Path $OpenSSLPath)) {
    Write-Error "OpenSSL not found at standard location: $OpenSSLPath"
    exit 1
}

$KeyDir = "$HOME\.oci"
if (-not (Test-Path $KeyDir)) {
    New-Item -ItemType Directory -Path $KeyDir -Force | Out-Null
    Write-Host "Created directory: $KeyDir" -ForegroundColor Cyan
}

$PrivKey = "$KeyDir\oci_api_key.pem"
$PubKey = "$KeyDir\oci_api_key_public.pem"

Write-Host "Generating RSA Key Pair..." -ForegroundColor Yellow

# Generate Private Key
& $OpenSSLPath genrsa -out "$PrivKey" 2048
# Fix permissions (Windows ACLs are tricky, but file structure is key)

# Generate Public Key
& $OpenSSLPath rsa -pubout -in "$PrivKey" -out "$PubKey"

Write-Host "✅ Keys Generated Successfully!" -ForegroundColor Green
Write-Host "----------------------------------------"
Write-Host "Private Key: $PrivKey"
Write-Host "Public Key:  $PubKey"
Write-Host "----------------------------------------"
Write-Host "ACTION REQUIRED:" -ForegroundColor Magenta
Write-Host "1. Copy the content below:"
Write-Host "2. Go to OCI Console -> User Settings -> API Keys -> Add API Key"
Write-Host "3. URL: https://cloud.oracle.com/identity/users"
Write-Host "4. Select 'Paste Public Key' and paste it."
Write-Host "----------------------------------------"
Get-Content $PubKey
Write-Host "----------------------------------------"
