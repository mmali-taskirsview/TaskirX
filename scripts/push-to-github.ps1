# push-to-github.ps1
# Helper script to push code to GitHub without needing the 'gh' CLI tool.

param(
    [Parameter(Mandatory=$false)]
    [string]$RepoUrl
)

if (-not $RepoUrl) {
    Write-Host "Please paste your GitHub Repository URL (e.g. https://github.com/username/repo.git):" -ForegroundColor Cyan
    $RepoUrl = Read-Host
}

if (-not $RepoUrl) {
    Write-Error "Repository URL is required."
    exit 1
}

Write-Host "`n[1/3] Setting remote origin to: $RepoUrl" -ForegroundColor Cyan
# Remove existing origin if it exists to avoid errors
git remote remove origin 2>$null 
git remote add origin $RepoUrl

Write-Host "[2/3] Renaming branch to 'main'" -ForegroundColor Cyan
git branch -M main

Write-Host "[3/3] Pushing code to GitHub..." -ForegroundColor Cyan
git push -u origin main

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nSUCCESS: Code pushed to GitHub!" -ForegroundColor Green
    Write-Host "Go to $RepoUrl/actions to see your deployment running." -ForegroundColor Cyan
} else {
    Write-Host "`nERROR: Push failed. Please check your URL and permissions." -ForegroundColor Red
}
