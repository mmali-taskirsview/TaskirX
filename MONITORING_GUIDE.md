# TaskirX V3 - Monitoring & Observability Guide

## Overview

This guide covers the monitoring stack for the Polyglot ad exchange, including Python AI services, Go Bidding Engine, and NestJS Backend.

---

## 1. Metrics & Prometheus

We use Prometheus to scrape metrics from all microservices.

### Scrape Targets
| Service | Internal Port | Metric Path | Description |
|---------|---------------|-------------|-------------|
| **Go Bidding** | 8080 | `/metrics` | Throughput, QPS, Latency, Fraud Block Counts |
| **Fraud AI** | 6001 | `/metrics` | Inference time, Blocked IP count |
| **Ad Matching AI** | 6002 | `/metrics` | AI Models loaded, Inference latency |
| **Bid Optimizer** | 6003 | `/metrics` | Thompson Sampling stats, Exploration/Exploitation ratios |
| **NestJS** | 3000 | `/metrics` | HTTP Latency, Error Rates |

### Accessing Prometheus
- **Local**: http://localhost:9090
- **Prod**: https://prometheus.yourdomain.com

## 2. Grafana Dashboards

We provide pre-provisioned dashboards in `monitoring/grafana/provisioning/dashboards`.

### Dashboard: Polyglot Overview
- **Throughput**: Real-time graph of bid requests vs bids placed.
- **AI Latency**: Compare latency of Fraud vs Matching vs Optimization.
- **Business Health**: Win rates, Click-through rates (CTR).

### Accessing Grafana
- **Local**: http://localhost:3002
- **Default Creds**: `admin` / `admin`

## 3. Alerts

Alert Manager is configured in `monitoring/alert-rules.yml`.

### Key Alerts
- **HighAPILatency**: > 0.5s p99 latency
- **HighFraudActivity**: > 100 blocked requests/sec
- **LowBidRate**: < 5% bid rate (Potential integration issue)
- **AIIntegrationErrors**: Failures communicating with Python services

## 4. Key Metrics Reference

### Bidding Engine
*   `bid_requests_total`: Counter. Total requests.
*   `bid_requests_by_format_total`: CounterVec (`format`). Requests by type (banner, video, native, audio).
*   `bids_placed_total`: CounterVec (`format`). Successful bids by type.
*   `fraud_blocked_total`: Counter.
*   `bid_latency_seconds`: Histogram. Processing time.

---

## Logging

All containers output to stdout/stderr. In production, we assume a log shipper (Fluentd/Logstash) collects these.

- **Go**: Structured JSON logs.
- **Python**: Standard Python logging (INFO level default).
- **NestJS**: Winston logger.
