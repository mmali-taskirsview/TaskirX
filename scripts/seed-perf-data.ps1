param(
    [string]$DbHost = "localhost",
    [string]$Port = "5432",
    [string]$User = "postgres",
    [string]$Password = "postgres",
    [string]$Database = "taskirx"
)

$env:POSTGRES_HOST = $DbHost
$env:POSTGRES_PORT = $Port
$env:POSTGRES_USER = $User
$env:POSTGRES_PASSWORD = $Password
$env:POSTGRES_DB = $Database

Write-Host "Seeding performance data into '$Database' on $Host..." -ForegroundColor Cyan

# Check for node
if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
    Write-Error "Node.js is required to run this script (since psql is missing)."
    exit 1
}

# Install pg driver if missing in a temp way, or assume it's in node_modules if we run from root?
# Better: use the project's local node_modules if available, or try to run with npx if possible, 
# but npx pg might not work for a script file. 
# Simplest: We'll assume the user can run `npm install pg` in the scripts dir or global.

# Actually, let's try to reuse nestjs-backend dependencies if possible, or just require the user to install pg.
# But to make it seamless, let's check if we can skip the install if 'pg' is resolvable.

$scriptPath = Join-Path $PSScriptRoot "seed-perf-db.js"

if (-not (Test-Path "node_modules/pg")) {
    Write-Host "Installing 'pg' driver locally to run seed script..." -ForegroundColor Yellow
    npm install pg --no-save --silent
}

node $scriptPath

if ($LASTEXITCODE -eq 0) {
    Write-Host "Seeding completed." -ForegroundColor Green
} else {
    Write-Host "Seeding failed." -ForegroundColor Red
}
