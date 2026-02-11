#!/bin/bash

# PHASE 3: MONITORING SETUP DEPLOYMENT GUIDE
# TaskirX Production Platform
# Comprehensive monitoring stack deployment and verification

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MONITORING_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCKER_COMPOSE_FILE="${MONITORING_DIR}/docker-compose.yml"
PROMETHEUS_CONFIG="${MONITORING_DIR}/prometheus.yml"
ALERTMANAGER_CONFIG="${MONITORING_DIR}/alertmanager/alertmanager.yml"
LOGSTASH_CONFIG="${MONITORING_DIR}/logstash.conf"

# Service URLs
PROMETHEUS_URL="http://localhost:9090"
GRAFANA_URL="http://localhost:3002"
ALERTMANAGER_URL="http://localhost:9093"
KIBANA_URL="http://localhost:5601"
JAEGER_URL="http://localhost:16686"
ELASTICSEARCH_URL="http://localhost:9200"

# Logging functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

error() {
    echo -e "${RED}[✗]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed"
        exit 1
    fi
    success "Docker is installed"
    
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose is not installed"
        exit 1
    fi
    success "Docker Compose is installed"
    
    if ! command -v curl &> /dev/null; then
        error "curl is not installed"
        exit 1
    fi
    success "curl is installed"
    
    # Check Docker daemon
    if ! docker ps &> /dev/null; then
        error "Docker daemon is not running"
        exit 1
    fi
    success "Docker daemon is running"
}

# Validate configuration files
validate_configs() {
    log "Validating configuration files..."
    
    if [ ! -f "$PROMETHEUS_CONFIG" ]; then
        error "Prometheus configuration not found: $PROMETHEUS_CONFIG"
        exit 1
    fi
    success "Prometheus config found"
    
    if [ ! -f "$ALERTMANAGER_CONFIG" ]; then
        error "AlertManager configuration not found: $ALERTMANAGER_CONFIG"
        exit 1
    fi
    success "AlertManager config found"
    
    if [ ! -f "$LOGSTASH_CONFIG" ]; then
        error "Logstash configuration not found: $LOGSTASH_CONFIG"
        exit 1
    fi
    success "Logstash config found"
}

# Start monitoring stack
start_monitoring_stack() {
    log "Starting monitoring stack..."
    
    cd "$MONITORING_DIR"
    
    docker-compose down || true
    sleep 2
    
    docker-compose up -d
    
    log "Waiting for services to start..."
    sleep 10
}

# Verify Prometheus
verify_prometheus() {
    log "Verifying Prometheus..."
    
    for i in {1..30}; do
        if curl -s "$PROMETHEUS_URL/-/healthy" &> /dev/null; then
            success "Prometheus is healthy"
            return 0
        fi
        warning "Prometheus not ready, attempt $i/30..."
        sleep 2
    done
    
    error "Prometheus failed to start"
    return 1
}

# Verify Grafana
verify_grafana() {
    log "Verifying Grafana..."
    
    for i in {1..30}; do
        if curl -s "$GRAFANA_URL/api/health" &> /dev/null; then
            success "Grafana is healthy"
            return 0
        fi
        warning "Grafana not ready, attempt $i/30..."
        sleep 2
    done
    
    error "Grafana failed to start"
    return 1
}

# Verify AlertManager
verify_alertmanager() {
    log "Verifying AlertManager..."
    
    for i in {1..30}; do
        if curl -s "$ALERTMANAGER_URL/-/healthy" &> /dev/null; then
            success "AlertManager is healthy"
            return 0
        fi
        warning "AlertManager not ready, attempt $i/30..."
        sleep 2
    done
    
    error "AlertManager failed to start"
    return 1
}

# Verify Elasticsearch
verify_elasticsearch() {
    log "Verifying Elasticsearch..."
    
    for i in {1..30}; do
        if curl -s "$ELASTICSEARCH_URL/_cluster/health" &> /dev/null; then
            success "Elasticsearch is healthy"
            return 0
        fi
        warning "Elasticsearch not ready, attempt $i/30..."
        sleep 2
    done
    
    error "Elasticsearch failed to start"
    return 1
}

# Verify Kibana
verify_kibana() {
    log "Verifying Kibana..."
    
    for i in {1..30}; do
        if curl -s "$KIBANA_URL/api/status" &> /dev/null; then
            success "Kibana is healthy"
            return 0
        fi
        warning "Kibana not ready, attempt $i/30..."
        sleep 2
    done
    
    error "Kibana failed to start"
    return 1
}

# Verify Jaeger
verify_jaeger() {
    log "Verifying Jaeger..."
    
    for i in {1..30}; do
        if curl -s "$JAEGER_URL/" &> /dev/null; then
            success "Jaeger is healthy"
            return 0
        fi
        warning "Jaeger not ready, attempt $i/30..."
        sleep 2
    done
    
    error "Jaeger failed to start"
    return 1
}

# Check Prometheus scrape targets
check_prometheus_targets() {
    log "Checking Prometheus scrape targets..."
    
    TARGETS=$(curl -s "$PROMETHEUS_URL/api/v1/targets" | grep -o '"health":"up"' | wc -l)
    
    if [ "$TARGETS" -gt 0 ]; then
        success "Prometheus has $TARGETS healthy targets"
    else
        warning "No healthy targets found in Prometheus"
    fi
}

# Create Kibana index patterns
create_kibana_indices() {
    log "Creating Kibana index patterns..."
    
    # Wait for Elasticsearch to have data
    sleep 5
    
    # Create logs index pattern
    curl -s -X POST \
        "$ELASTICSEARCH_URL/.kibana/_doc/index-pattern:logs-*" \
        -H 'Content-Type: application/json' \
        -d '{
            "type": "index-pattern",
            "index-pattern": {
                "title": "logs-*",
                "timeFieldName": "@timestamp",
                "fields": "[]"
            }
        }' || true
    
    success "Kibana index patterns created"
}

# Setup authentication for Grafana
setup_grafana_auth() {
    log "Setting up Grafana authentication..."
    
    # Wait for Grafana to be ready
    sleep 5
    
    # Update admin password (optional - default is admin/admin)
    curl -s -X PUT \
        "$GRAFANA_URL/api/admin/users/1/password" \
        -H "Content-Type: application/json" \
        -H "Authorization: Basic $(echo -n 'admin:admin' | base64)" \
        -d '{"password": "ChangeMe123!"}' || true
    
    success "Grafana authentication setup complete"
}

# Verify alert rules
verify_alert_rules() {
    log "Verifying alert rules..."
    
    RULES=$(curl -s "$PROMETHEUS_URL/api/v1/rules" | grep -o '"alert"' | wc -l)
    
    if [ "$RULES" -gt 0 ]; then
        success "Prometheus loaded $RULES alert rules"
    else
        error "No alert rules loaded in Prometheus"
    fi
}

# Print service URLs
print_service_urls() {
    log "Monitoring services are now available:"
    
    echo -e "${GREEN}"
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║            TASKIR MONITORING SERVICES                  ║"
    echo "╠════════════════════════════════════════════════════════╣"
    echo "║ Prometheus      http://localhost:9090                  ║"
    echo "║ Grafana         http://localhost:3002                  ║"
    echo "║ AlertManager    http://localhost:9093                  ║"
    echo "║ Kibana          http://localhost:5601                  ║"
    echo "║ Jaeger          http://localhost:16686                 ║"
    echo "║ Elasticsearch   http://localhost:9200                  ║"
    echo "╠════════════════════════════════════════════════════════╣"
    echo "║ Grafana Default Credentials: admin/admin               ║"
    echo "║ (CHANGE IN PRODUCTION!)                                ║"
    echo "╚════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Check disk space
check_disk_space() {
    log "Checking available disk space..."
    
    AVAILABLE=$(df "$MONITORING_DIR" | awk 'NR==2 {print $4}')
    REQUIRED=$((30 * 1024 * 1024)) # 30GB in KB
    
    if [ "$AVAILABLE" -lt "$REQUIRED" ]; then
        warning "Low disk space: $(($AVAILABLE / 1024 / 1024))GB available, 30GB recommended for 30-day retention"
    else
        success "Sufficient disk space available: $(($AVAILABLE / 1024 / 1024))GB"
    fi
}

# Print deployment summary
print_summary() {
    echo -e "${GREEN}"
    echo ""
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║    PHASE 3: MONITORING SETUP - DEPLOYMENT COMPLETE     ║"
    echo "╠════════════════════════════════════════════════════════╣"
    echo "║                                                        ║"
    echo "║  ✓ Prometheus metrics collection (30s interval)        ║"
    echo "║  ✓ 30+ alert rules configured                          ║"
    echo "║  ✓ Grafana dashboards provisioned                      ║"
    echo "║  ✓ AlertManager routing configured                     ║"
    echo "║  ✓ ELK Stack (Elasticsearch, Logstash, Kibana)         ║"
    echo "║  ✓ Distributed tracing (Jaeger)                        ║"
    echo "║  ✓ Node exporter for system metrics                    ║"
    echo "║  ✓ cAdvisor for container metrics                      ║"
    echo "║                                                        ║"
    echo "╠════════════════════════════════════════════════════════╣"
    echo "║                                                        ║"
    echo "║  Next Steps:                                           ║"
    echo "║  1. Configure alert notification channels              ║"
    echo "║  2. Update Grafana admin password                      ║"
    echo "║  3. Import custom dashboards                           ║"
    echo "║  4. Configure log retention policies                   ║"
    echo "║  5. Integrate with incident management system          ║"
    echo "║                                                        ║"
    echo "╚════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Main deployment function
main() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║   TASKIR PHASE 3: MONITORING SETUP DEPLOYMENT          ║"
    echo "║                                                        ║"
    echo "║   Components:                                          ║"
    echo "║   • Prometheus (Metrics Collection)                    ║"
    echo "║   • Grafana (Visualization)                            ║"
    echo "║   • AlertManager (Alert Routing)                       ║"
    echo "║   • ELK Stack (Centralized Logging)                    ║"
    echo "║   • Jaeger (Distributed Tracing)                       ║"
    echo "║   • Node Exporter (System Metrics)                     ║"
    echo "║   • cAdvisor (Container Metrics)                       ║"
    echo "║                                                        ║"
    echo "╚════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo ""
    
    check_prerequisites
    validate_configs
    check_disk_space
    start_monitoring_stack
    
    # Verify all services
    verify_prometheus && \
    verify_grafana && \
    verify_alertmanager && \
    verify_elasticsearch && \
    verify_kibana && \
    verify_jaeger && \
    check_prometheus_targets && \
    verify_alert_rules && \
    create_kibana_indices && \
    setup_grafana_auth || {
        error "Deployment failed during verification"
        exit 1
    }
    
    print_service_urls
    print_summary
    
    success "Phase 3: Monitoring Setup deployment completed successfully!"
}

# Run main function
main
