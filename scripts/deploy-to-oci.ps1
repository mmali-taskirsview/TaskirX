param(
    [string]$Registry = "your-registry",
    [string]$Action = "plan"
)

# Helper wrapper for OCI deployment
.\scripts\deploy.ps1 -Environment "oci" -Action $Action -Registry $Registry
