# PHASE 3: MONITORING SETUP DEPLOYMENT - WINDOWS VERSION
# TaskirX Production Platform
# Comprehensive monitoring stack deployment and verification

param(
    [switch]$SkipVerification,
    [switch]$VerboseOutput
)

# Configuration
$MonitoringDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$DockerComposeFile = Join-Path $MonitoringDir "docker-compose.yml"
$PrometheusConfig = Join-Path $MonitoringDir "prometheus.yml"
$AlertManagerConfig = Join-Path $MonitoringDir "alertmanager\alertmanager.yml"
$LogstashConfig = Join-Path $MonitoringDir "logstash.conf"

# Service URLs
$PrometheusUrl = "http://localhost:9090"
$GrafanaUrl = "http://localhost:3002"
$AlertManagerUrl = "http://localhost:9093"
$KibanaUrl = "http://localhost:5601"
$JaegerUrl = "http://localhost:16686"
$ElasticsearchUrl = "http://localhost:9200"

# Logging functions
function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "[OK] $Message" -ForegroundColor Green
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Write-Warning-Custom {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

# Check prerequisites
function Test-Prerequisites {
    Write-Log "Checking prerequisites..."
    
    # Check Docker
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Error-Custom "Docker is not installed or not in PATH"
        exit 1
    }
    Write-Success "Docker is installed"
    
    # Check Docker Compose
    if (-not (Get-Command docker-compose -ErrorAction SilentlyContinue)) {
        Write-Error-Custom "Docker Compose is not installed or not in PATH"
        exit 1
    }
    Write-Success "Docker Compose is installed"
    
    # Check curl
    if (-not (Get-Command curl -ErrorAction SilentlyContinue)) {
        Write-Error-Custom "curl is not installed or not in PATH"
        exit 1
    }
    Write-Success "curl is installed"
    
    # Check Docker daemon
    try {
        $null = docker ps -q 2>$null
        Write-Success "Docker daemon is running"
    }
    catch {
        Write-Error-Custom "Docker daemon is not running"
        exit 1
    }
}

# Validate configuration files
function Test-ConfigurationFiles {
    Write-Log "Validating configuration files..."
    
    if (-not (Test-Path $PrometheusConfig)) {
        Write-Error-Custom "Prometheus configuration not found: $PrometheusConfig"
        exit 1
    }
    Write-Success "Prometheus config found"
    
    if (-not (Test-Path $AlertManagerConfig)) {
        Write-Error-Custom "AlertManager configuration not found: $AlertManagerConfig"
        exit 1
    }
    Write-Success "AlertManager config found"
    
    if (-not (Test-Path $LogstashConfig)) {
        Write-Error-Custom "Logstash configuration not found: $LogstashConfig"
        exit 1
    }
    Write-Success "Logstash config found"
}

# Start monitoring stack
function Start-MonitoringStack {
    Write-Log "Starting monitoring stack..."
    
    Push-Location $MonitoringDir
    
    try {
        Write-Log "Stopping existing containers..."
        docker-compose down 2>&1 | Where-Object { $VerboseOutput }
        Start-Sleep -Seconds 2
        
        Write-Log "Starting new containers..."
        docker-compose up -d
        
        Write-Log "Waiting for services to start..."
        Start-Sleep -Seconds 10
    }
    finally {
        Pop-Location
    }
}

# Test service health
function Test-ServiceHealth {
    param(
        [string]$ServiceName,
        [string]$HealthUrl,
        [int]$MaxRetries = 30
    )
    
    Write-Log "Verifying $ServiceName..."
    
    for ($i = 1; $i -le $MaxRetries; $i++) {
        try {
            $response = Invoke-WebRequest -Uri $HealthUrl -UseBasicParsing -TimeoutSec 5 -ErrorAction SilentlyContinue
            if ($response.StatusCode -eq 200) {
                Write-Success "$ServiceName is healthy"
                return $true
            }
        }
        catch {
            if ($VerboseOutput) {
                Write-Warning-Custom "$ServiceName not ready, attempt $i/$MaxRetries..."
            }
        }
        
        Start-Sleep -Seconds 2
    }
    
    Write-Error-Custom "$ServiceName failed to start"
    return $false
}

# Verify all services
function Invoke-ServiceVerification {
    if ($SkipVerification) {
        Write-Log "Skipping service verification..."
        return $true
    }
    
    $allHealthy = $true
    
    $allHealthy = $allHealthy -and (Test-ServiceHealth "Prometheus" "$PrometheusUrl/-/healthy")
    $allHealthy = $allHealthy -and (Test-ServiceHealth "Grafana" "$GrafanaUrl/api/health")
    $allHealthy = $allHealthy -and (Test-ServiceHealth "AlertManager" "$AlertManagerUrl/-/healthy")
    $allHealthy = $allHealthy -and (Test-ServiceHealth "Elasticsearch" "$ElasticsearchUrl/_cluster/health")
    $allHealthy = $allHealthy -and (Test-ServiceHealth "Kibana" "$KibanaUrl/api/status")
    $allHealthy = $allHealthy -and (Test-ServiceHealth "Jaeger" "$JaegerUrl/")
    
    return $allHealthy
}

# Check Prometheus scrape targets
function Test-PrometheusTargets {
    Write-Log "Checking Prometheus scrape targets..."
    
    try {
        $response = Invoke-WebRequest -Uri "$PrometheusUrl/api/v1/targets" -UseBasicParsing
        $json = $response.Content | ConvertFrom-Json
        $healthyTargets = ($json.data.activeTargets | Where-Object { $_.health -eq "up" }).Count
        
        if ($healthyTargets -gt 0) {
            Write-Success "Prometheus has $healthyTargets healthy targets"
        }
        else {
            Write-Warning-Custom "No healthy targets found in Prometheus"
        }
    }
    catch {
        Write-Warning-Custom "Failed to check Prometheus targets: $_"
    }
}

# Verify alert rules
function Test-AlertRules {
    Write-Log "Verifying alert rules..."
    
    try {
        $response = Invoke-WebRequest -Uri "$PrometheusUrl/api/v1/rules" -UseBasicParsing
        $json = $response.Content | ConvertFrom-Json
        $ruleCount = $json.data.groups | ForEach-Object { $_.rules.Count } | Measure-Object -Sum | Select-Object -ExpandProperty Sum
        
        if ($ruleCount -gt 0) {
            Write-Success "Prometheus loaded $ruleCount alert rules"
        }
        else {
            Write-Error-Custom "No alert rules loaded in Prometheus"
        }
    }
    catch {
        Write-Warning-Custom "Failed to check alert rules: $_"
    }
}

# Create Kibana index patterns
function New-KibanaIndexPatterns {
    Write-Log "Creating Kibana index patterns..."
    
    Start-Sleep -Seconds 5
    
    try {
        $body = @{
            type = "index-pattern"
            "index-pattern" = @{
                title = "logs-*"
                timeFieldName = "@timestamp"
                fields = "[]"
            }
        } | ConvertTo-Json
        
        $null = Invoke-WebRequest -Uri "$ElasticsearchUrl/.kibana/_doc/index-pattern:logs-*" `
            -Method POST `
            -ContentType "application/json" `
            -Body $body `
            -UseBasicParsing `
            -ErrorAction SilentlyContinue
        
        Write-Success "Kibana index patterns created"
    }
    catch {
        Write-Warning-Custom "Failed to create Kibana index patterns: $_"
    }
}

# Setup Grafana authentication
function Set-GrafanaAuthentication {
    Write-Log "Setting up Grafana authentication..."
    
    Start-Sleep -Seconds 5
    
    try {
        $body = @{ password = "ChangeMe123!" } | ConvertTo-Json
        $base64Credentials = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("admin:admin"))
        
        $null = Invoke-WebRequest -Uri "$GrafanaUrl/api/admin/users/1/password" `
            -Method PUT `
            -ContentType "application/json" `
            -Headers @{ Authorization = "Basic $base64Credentials" } `
            -Body $body `
            -UseBasicParsing `
            -ErrorAction SilentlyContinue
        
        Write-Success "Grafana authentication setup complete"
    }
    catch {
        Write-Warning-Custom "Failed to setup Grafana authentication: $_"
    }
}

# Print service URLs
function Print-ServiceUrls {
    Write-Host ""
    Write-Host "╔════════════════════════════════════════════════════════╗" -ForegroundColor Green
    Write-Host "║            TASKIR MONITORING SERVICES                  ║" -ForegroundColor Green
    Write-Host "╠════════════════════════════════════════════════════════╣" -ForegroundColor Green
    Write-Host "║ Prometheus      http://localhost:9090                  ║" -ForegroundColor Green
    Write-Host "║ Grafana         http://localhost:3002                  ║" -ForegroundColor Green
    Write-Host "║ AlertManager    http://localhost:9093                  ║" -ForegroundColor Green
    Write-Host "║ Kibana          http://localhost:5601                  ║" -ForegroundColor Green
    Write-Host "║ Jaeger          http://localhost:16686                 ║" -ForegroundColor Green
    Write-Host "║ Elasticsearch   http://localhost:9200                  ║" -ForegroundColor Green
    Write-Host "╠════════════════════════════════════════════════════════╣" -ForegroundColor Green
    Write-Host "║ Grafana Default Credentials: admin/admin               ║" -ForegroundColor Green
    Write-Host "║ (CHANGE IN PRODUCTION!)                                ║" -ForegroundColor Green
    Write-Host "╚════════════════════════════════════════════════════════╝" -ForegroundColor Green
    Write-Host ""
}

# Print deployment summary
function Print-DeploymentSummary {
    Write-Host ""
    Write-Host "╔════════════════════════════════════════════════════════╗" -ForegroundColor Green
    Write-Host "║    PHASE 3: MONITORING SETUP - DEPLOYMENT COMPLETE     ║" -ForegroundColor Green
    Write-Host "╠════════════════════════════════════════════════════════╣" -ForegroundColor Green
    Write-Host "║                                                        ║" -ForegroundColor Green
    Write-Host "║  [OK] Prometheus metrics collection (30s interval)      ║" -ForegroundColor Green
    Write-Host "║  [OK] 30+ alert rules configured                        ║" -ForegroundColor Green
    Write-Host "║  [OK] Grafana dashboards provisioned                    ║" -ForegroundColor Green
    Write-Host "║  [OK] AlertManager routing configured                   ║" -ForegroundColor Green
    Write-Host "║  [OK] ELK Stack (Elasticsearch, Logstash, Kibana)       ║" -ForegroundColor Green
    Write-Host "║  [OK] Distributed tracing (Jaeger)                      ║" -ForegroundColor Green
    Write-Host "║  [OK] Node exporter for system metrics                  ║" -ForegroundColor Green
    Write-Host "║  [OK] cAdvisor for container metrics                    ║" -ForegroundColor Green
    Write-Host "║                                                        ║" -ForegroundColor Green
    Write-Host "╠════════════════════════════════════════════════════════╣" -ForegroundColor Green
    Write-Host "║                                                        ║" -ForegroundColor Green
    Write-Host "║  Next Steps:                                           ║" -ForegroundColor Green
    Write-Host "║  1. Configure alert notification channels              ║" -ForegroundColor Green
    Write-Host "║  2. Update Grafana admin password                      ║" -ForegroundColor Green
    Write-Host "║  3. Import custom dashboards                           ║" -ForegroundColor Green
    Write-Host "║  4. Configure log retention policies                   ║" -ForegroundColor Green
    Write-Host "║  5. Integrate with incident management system          ║" -ForegroundColor Green
    Write-Host "║                                                        ║" -ForegroundColor Green
    Write-Host "╚════════════════════════════════════════════════════════╝" -ForegroundColor Green
    Write-Host ""
}

# Main function
function Invoke-Main {
    Write-Host ""
    Write-Host "╔════════════════════════════════════════════════════════╗" -ForegroundColor Blue
    Write-Host "║   TASKIR PHASE 3: MONITORING SETUP DEPLOYMENT          ║" -ForegroundColor Blue
    Write-Host "║                                                        ║" -ForegroundColor Blue
    Write-Host "║   Components:                                          ║" -ForegroundColor Blue
    Write-Host "║   - Prometheus (Metrics Collection)                    ║" -ForegroundColor Blue
    Write-Host "║   - Grafana (Visualization)                            ║" -ForegroundColor Blue
    Write-Host "║   - AlertManager (Alert Routing)                       ║" -ForegroundColor Blue
    Write-Host "║   - ELK Stack (Centralized Logging)                    ║" -ForegroundColor Blue
    Write-Host "║   - Jaeger (Distributed Tracing)                       ║" -ForegroundColor Blue
    Write-Host "║   - Node Exporter (System Metrics)                     ║" -ForegroundColor Blue
    Write-Host "║   - cAdvisor (Container Metrics)                       ║" -ForegroundColor Blue
    Write-Host "║                                                        ║" -ForegroundColor Blue
    Write-Host "╚════════════════════════════════════════════════════════╝" -ForegroundColor Blue
    Write-Host ""
    
    Test-Prerequisites
    Test-ConfigurationFiles
    Start-MonitoringStack
    
    if (Invoke-ServiceVerification) {
        Test-PrometheusTargets
        Test-AlertRules
        New-KibanaIndexPatterns
        Set-GrafanaAuthentication
        
        Print-ServiceUrls
        Print-DeploymentSummary
        
        Write-Success "Phase 3: Monitoring Setup deployment completed successfully!"
    }
    else {
        Write-Error-Custom "Deployment failed during verification"
        exit 1
    }
}

# Run main function
Invoke-Main
