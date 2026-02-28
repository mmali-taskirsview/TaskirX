# TaskirX Server Debugging Script
# Helps diagnose the Windows access violation issue

Write-Host "🔍 TaskirX Server Debugging Toolkit" -ForegroundColor Cyan
Write-Host "====================================="

function Test-GoVersion {
    Write-Host "`n📊 Go Environment Check..." -ForegroundColor Yellow
    go version
    go env GOOS GOARCH
}

function Test-Build {
    Write-Host "`n🔨 Testing Build Process..." -ForegroundColor Yellow
    cd c:\TaskirX\go-bidding-engine\cmd\test-server
    
    Write-Host "Building with race detector..."
    $raceResult = go build -race -o test-server-race.exe main.go 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Race detector build: SUCCESS" -ForegroundColor Green
    } else {
        Write-Host "❌ Race detector build: FAILED" -ForegroundColor Red
        Write-Host $raceResult
    }
    
    Write-Host "`nBuilding with verbose output..."
    $verboseResult = go build -v -o test-server-verbose.exe main.go 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Verbose build: SUCCESS" -ForegroundColor Green
    } else {
        Write-Host "❌ Verbose build: FAILED" -ForegroundColor Red  
        Write-Host $verboseResult
    }
}

function Test-Dependencies {
    Write-Host "`n📦 Checking Dependencies..." -ForegroundColor Yellow
    cd c:\TaskirX\go-bidding-engine
    
    go mod verify
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Module verification: PASSED" -ForegroundColor Green
    } else {
        Write-Host "❌ Module verification: FAILED" -ForegroundColor Red
    }
    
    Write-Host "`nDependency tree:"
    go list -m all | Select-Object -First 10
}

function Start-ServerWithRaceDetection {
    Write-Host "`n🚀 Starting server with race detection..." -ForegroundColor Yellow
    cd c:\TaskirX\go-bidding-engine\cmd\test-server
    
    if (Test-Path "test-server-race.exe") {
        Write-Host "Running race detector version..."
        Start-Process -FilePath ".\test-server-race.exe" -NoNewWindow -PassThru
        Start-Sleep 2
        
        Write-Host "Testing health endpoint..."
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -TimeoutSec 3
            Write-Host "✅ Health check successful: $($response.StatusCode)" -ForegroundColor Green
        } catch {
            Write-Host "❌ Health check failed: $($_.Exception.Message)" -ForegroundColor Red
        }
    } else {
        Write-Host "❌ Race detector build not found" -ForegroundColor Red
    }
}

function Show-ServerLogs {
    Write-Host "`n📝 Recent Server Events..." -ForegroundColor Yellow
    
    # Check Windows Event Log for application crashes
    try {
        Get-EventLog -LogName Application -Source "Application Error" -Newest 5 -ErrorAction SilentlyContinue | 
        Where-Object { $_.Message -like "*test-server*" -or $_.Message -like "*go*" } |
        Select-Object TimeGenerated, EventID, Message | 
        Format-Table -AutoSize
    } catch {
        Write-Host "No recent crash events found" -ForegroundColor Green
    }
}

function Analyze-Memory {
    Write-Host "`n🧠 Memory Analysis..." -ForegroundColor Yellow
    
    # Available memory
    $memory = Get-CimInstance -ClassName Win32_OperatingSystem
    $freeMemoryGB = [math]::Round($memory.FreePhysicalMemory / 1MB, 2)
    $totalMemoryGB = [math]::Round($memory.TotalPhysicalMemory / 1GB, 2)
    
    Write-Host "Available Memory: $freeMemoryGB GB / $totalMemoryGB GB"
    
    # Check for memory pressure
    if ($freeMemoryGB -lt 1) {
        Write-Host "⚠️  Low memory condition detected" -ForegroundColor Yellow
    } else {
        Write-Host "✅ Memory levels normal" -ForegroundColor Green
    }
}

function Show-Recommendations {
    Write-Host "`n💡 Debugging Recommendations:" -ForegroundColor Cyan
    Write-Host "1. Try running with different Go versions"
    Write-Host "2. Test on Linux/WSL for comparison"
    Write-Host "3. Use delve debugger for deeper analysis:"
    Write-Host "   dlv debug cmd/test-server/main.go"
    Write-Host "4. Add debug logging to handlers"
    Write-Host "5. Test with minimal Gin setup"
    Write-Host "6. Check for CGO dependencies"
    
    Write-Host "`n🔧 Quick Commands:" -ForegroundColor Yellow
    Write-Host "go env | findstr CGO"
    Write-Host "go run -race cmd/test-server/main.go"
    Write-Host "dlv debug cmd/test-server/main.go"
}

# Run all diagnostics
Test-GoVersion
Test-Dependencies  
Test-Build
Analyze-Memory
Show-ServerLogs
Show-Recommendations

Write-Host "`n🎯 Summary: Server debugging information collected" -ForegroundColor Green
Write-Host "Check the output above for potential issues." -ForegroundColor Green