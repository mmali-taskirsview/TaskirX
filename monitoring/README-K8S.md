# Monitoring Setup for OCI (Kubernetes)

For OCI (and Kubernetes in general), we recommend using the **kube-prometheus-stack** Helm chart rather than a custom Docker Compose setup. This provides a robust, production-ready observability stack including:
- Prometheus (Metrics)
- Grafana (Visualization)
- AlertManager (Notification)
- Node Exporter (Infrastructure stats)

## 1. Prerequisites
- Helm installed (`choco install kubernetes-helm`)
- `kubectl` configured for your OCI cluster

## 2. Install Monitoring Stack
```powershell
# Add Prometheus Community Repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Create monitoring namespace
kubectl create namespace monitoring

# Install Stack (Prometheus + Grafana + AlertManager)
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack `
    --namespace monitoring `
    --set grafana.adminPassword="admin" `
    --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false `
    --set prometheus.prometheusSpec.serviceMonitorSelector.matchLabels.release="kube-prometheus-stack"
```

## 3. Import Dashboards
To import your Pinecone Dashboard (`monitoring/grafana/dashboards/pinecone-dashboard.json`):
1. Login to Grafana.
2. Go to **Dashboards** > **Import**.
3. Upload the JSON file.

## 4. Port Forwarding
```powershell
kubectl port-forward svc/kube-prometheus-stack-grafana 3000:80 -n monitoring
```
Then visit: `http://localhost:3000`
