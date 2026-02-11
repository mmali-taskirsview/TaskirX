# Prepare GitOps
Write-Host "Preparing repository for GitHub Actions..." -ForegroundColor Cyan

# 1. Config
if (-not (Test-Path ".gitignore")) {
    Write-Host "Creating .gitignore..."
    Set-Content .gitignore "node_modules/`n.env`n.terraform/`n*.tfstate`n*.tfstate.backup`n.kube/`n"
}

# 2. Add and Commit
git add .
git commit -m "Initial commit for TaskirX OCI Deployment"

Write-Host "`nRepository initialized and files committed." -ForegroundColor Green
Write-Host "NEXT STEPS:" -ForegroundColor Yellow
Write-Host "1. Create a new repository on GitHub (https://github.com/new)"
Write-Host "2. Run the following commands:"
Write-Host "   git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPO_NAME.git"
Write-Host "   git branch -M main"
Write-Host "   git push -u origin main"
Write-Host "`nOnce pushed, the 'CI/CD Pipeline' workflow will appear in the 'Actions' tab on GitHub."
