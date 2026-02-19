# 0. Cleanup any old processes
$CurrentProcess = $PID
Get-Process -Name "go-bidding-engine" -ErrorAction SilentlyContinue | Where-Object { $_.Id -ne $CurrentProcess } | Stop-Process -Force

# 1. Configuration (Set Env Vars)
$env:ENV="development"
$env:PORT="8082"
$env:REDIS_HOST="localhost"
$env:REDIS_PORT="6379"
$env:REDIS_PASSWORD="taskir_redis_password_2026"

# 2. Start the Bidding Engine in background
$ExePath = Join-Path $PSScriptRoot "go-bidding-engine.exe"
if (-not (Test-Path $ExePath)) {
    Write-Error "Error: go-bidding-engine.exe not found at $ExePath. Please build it first."
    exit 1
}

Write-Host "Starting $ExePath..."
$Process = Start-Process -FilePath $ExePath -PassThru -NoNewWindow
Start-Sleep -Seconds 5

try {
    # 1. Test Standard OpenRTB Banner Request
    $BannerPayload = @{
        id = "req-123"
        imp = @(
            @{
                id = "imp-1"
                banner = @{
                    w = 300
                    h = 250
                }
            }
        )
        site = @{
            id = "site-123"
            page = "https://example.com/news"
        }
        device = @{
            ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
            ip = "127.0.0.1"
            geo = @{
                country = "US"
            }
        }
        user = @{
            id = "user-123"
        }
    } | ConvertTo-Json -Depth 5

    Write-Host "Sending Banner Request to /openrtb..."
    $Response = Invoke-RestMethod -Uri "http://localhost:8082/openrtb" -Method Post -Body $BannerPayload -ContentType "application/json"
    Write-Host "Response Received:" -ForegroundColor Green
    $Response | ConvertTo-Json -Depth 5 | Write-Host

    # 2. Test OpenRTB Video Request
    $VideoPayload = @{
        id = "req-video-123"
        imp = @(
            @{
                id = "imp-video-1"
                video = @{
                    mimes = @("video/mp4")
                    w = 640
                    h = 480
                }
            }
        )
        app = @{
            id = "app-123"
            bundle = "com.example.app"
        }
        device = @{
            ua = "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"
            ip = "127.0.0.1"
            geo = @{
                country = "US"
            }
        }
    } | ConvertTo-Json -Depth 5

    Write-Host "Sending Video Request to /openrtb..."
    $Response2 = Invoke-RestMethod -Uri "http://localhost:8082/openrtb" -Method Post -Body $VideoPayload -ContentType "application/json"
    Write-Host "Response Received:" -ForegroundColor Green
    $Response2 | ConvertTo-Json -Depth 5 | Write-Host

    # 3. Test OpenRTB Native Request (Standard 1.2 payload string)
    # The 'request' string is a JSON object itself
    $NativeRequestJson = @{
        native = @{
            ver = "1.2"
            assets = @(
                @{
                    id = 1
                    required = 1
                    title = @{ len = 140 }
                },
                @{
                    id = 123
                    required = 1
                    img = @{ type = 3; w = 300; h = 250 }
                },
                @{
                    id = 456
                    required = 0
                    data = @{ type = 2; len = 100 } # Description
                }
            )
        }
    } | ConvertTo-Json -Depth 5 -Compress

    # Escape quotes for nested JSON string
    $NativeRequestString = $NativeRequestJson.Replace('"', '\"')

    $NativePayload = @{
        id = "req-native-123"
        imp = @(
            @{
                id = "imp-native-1"
                native = @{
                    request = $NativeRequestJson
                }
            }
        )
        site = @{ id = "site-1" }
        device = @{ ua = "Mozilla/5.0"; ip = "127.0.0.1"; geo = @{ country = "US" } }
    } | ConvertTo-Json -Depth 5

    Write-Host "Sending Native Request to /openrtb..."
    $Response3 = Invoke-RestMethod -Uri "http://localhost:$env:PORT/openrtb" -Method Post -Body $NativePayload -ContentType "application/json"
    Write-Host "Response Received:" -ForegroundColor Green
    $Response3 | ConvertTo-Json -Depth 5 | Write-Host

    # 4. Test OpenRTB Audio Request
    $AudioPayload = @{
        id = "req-audio-123"
        imp = @(
            @{
                id = "imp-audio-1"
                audio = @{
                    mimes = @("audio/mp3", "audio/ogg")
                    minduration = 5
                    maxduration = 30
                }
            }
        )
        site = @{ id = "site-audio-1" }
        device = @{ ua = "Mozilla/5.0"; ip = "127.0.0.1"; geo = @{ country = "US" } }
    } | ConvertTo-Json -Depth 5

    Write-Host "Sending Audio Request to /openrtb..."
    $Response4 = Invoke-RestMethod -Uri "http://localhost:$env:PORT/openrtb" -Method Post -Body $AudioPayload -ContentType "application/json"
    Write-Host "Response Received:" -ForegroundColor Green
    $Response4 | ConvertTo-Json -Depth 5 | Write-Host

    # 5. Test Rich User/Device Data (Targeting)
    $RichPayload = @{
        id = "req-rich-123"
        imp = @(
            @{
                id = "imp-rich-1"
                banner = @{ w = 300; h = 250 }
            }
        )
        device = @{
            devicetype = 2 # PC
            ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
            ip = "127.0.0.1"
            geo = @{ country = "US" }
        }
        user = @{
            id = "user-rich-1"
            keywords = "sports,finance"
            data = @(
                @{
                    id = "dmp-1"
                    segment = @(
                        @{ id = "seg-high-income" }
                    )
                }
            )
        }
    } | ConvertTo-Json -Depth 5

    Write-Host "Sending Rich Targeting Request to /openrtb..."
    $Response5 = Invoke-RestMethod -Uri "http://localhost:$env:PORT/openrtb" -Method Post -Body $RichPayload -ContentType "application/json"
    Write-Host "Response Received:" -ForegroundColor Green
    $Response5 | ConvertTo-Json -Depth 5 | Write-Host

} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    if ($_.Exception.Response) {
        $_.Exception.Response.StatusCode
    }
} finally {
    # Cleanup
    Stop-Process -Id $Process.Id -Force
    Write-Host "Server stopped."
}
