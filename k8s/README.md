# TaskirX Kubernetes Deployment

Complete Kubernetes manifests for deploying TaskirX v3.0 to AWS EKS.

## 📁 Files

- `namespace.yaml` - TaskirX namespace
- `postgres-deployment.yaml` - PostgreSQL database with PVC
- `redis-deployment.yaml` - Redis cache with PVC
- `nestjs-deployment.yaml` - NestJS backend (3 replicas + HPA)
- `go-bidding-deployment.yaml` - Go bidding engine (5 replicas + HPA)
- `python-services-deployment.yaml` - 3 AI services (3 replicas each)
- `next-dashboard-deployment.yaml` - Next.js dashboard (2 replicas)
- `ingress.yaml` - ALB ingress controller

## 🚀 Quick Deploy

### Prerequisites

```powershell
# Install kubectl
choco install kubernetes-cli

# Install AWS CLI
choco install awscli

# Configure AWS
aws configure

# Install eksctl
choco install eksctl
```

### Create EKS Cluster

```powershell
# Create cluster (takes ~15 minutes)
eksctl create cluster `
  --name taskir-prod `
  --region us-east-1 `
  --node-type t3.xlarge `
  --nodes 3 `
  --nodes-min 3 `
  --nodes-max 10 `
  --managed

# Verify cluster
kubectl get nodes
```

### Deploy Application

```powershell
# Create namespace
kubectl apply -f namespace.yaml

# Deploy databases
kubectl apply -f postgres-deployment.yaml
kubectl apply -f redis-deployment.yaml

# Wait for databases to be ready
kubectl wait --for=condition=ready pod -l app=postgres -n taskir --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n taskir --timeout=300s

# Deploy backend services
kubectl apply -f nestjs-deployment.yaml
kubectl apply -f go-bidding-deployment.yaml
kubectl apply -f python-services-deployment.yaml

# Deploy frontend
kubectl apply -f next-dashboard-deployment.yaml

# Setup ingress
kubectl apply -f ingress.yaml

# Check status
kubectl get all -n taskir
```

## 🔍 Monitoring

### Check Pods

```powershell
# All pods
kubectl get pods -n taskir

# Specific service
kubectl get pods -l app=nestjs-backend -n taskir

# Watch pods
kubectl get pods -n taskir --watch
```

### View Logs

```powershell
# Specific pod
kubectl logs <pod-name> -n taskir

# Follow logs
kubectl logs -f <pod-name> -n taskir

# All pods of a deployment
kubectl logs -l app=nestjs-backend -n taskir --tail=100
```

### Check Services

```powershell
# All services
kubectl get svc -n taskir

# Service details
kubectl describe svc nestjs-backend -n taskir

# Check endpoints
kubectl get endpoints -n taskir
```

## 📊 Scaling

### Manual Scaling

```powershell
# Scale deployment
kubectl scale deployment nestjs-backend --replicas=5 -n taskir

# Scale all
kubectl scale deployment --all --replicas=5 -n taskir
```

### Auto-Scaling (HPA)

```powershell
# Check HPA status
kubectl get hpa -n taskir

# HPA details
kubectl describe hpa nestjs-hpa -n taskir

# Edit HPA
kubectl edit hpa nestjs-hpa -n taskir
```

## 🔧 Configuration

### Update ConfigMaps

```powershell
# Edit ConfigMap
kubectl edit configmap nestjs-config -n taskir

# Restart deployment to apply changes
kubectl rollout restart deployment nestjs-backend -n taskir
```

### Update Secrets

```powershell
# Create/update secret
kubectl create secret generic nestjs-secret `
  --from-literal=JWT_SECRET=your-new-secret `
  --dry-run=client -o yaml | kubectl apply -n taskir -f -

# Restart deployment
kubectl rollout restart deployment nestjs-backend -n taskir
```

## 🔄 Updates & Rollbacks

### Rolling Update

```powershell
# Update image
kubectl set image deployment/nestjs-backend `
  nestjs=your-registry/taskir-nestjs:v2.0 `
  -n taskir

# Check rollout status
kubectl rollout status deployment/nestjs-backend -n taskir

# Check history
kubectl rollout history deployment/nestjs-backend -n taskir
```

### Rollback

```powershell
# Rollback to previous version
kubectl rollout undo deployment/nestjs-backend -n taskir

# Rollback to specific revision
kubectl rollout undo deployment/nestjs-backend --to-revision=2 -n taskir
```

## 🐛 Debugging

### Exec into Pod

```powershell
# Bash/sh shell
kubectl exec -it <pod-name> -n taskir -- /bin/sh

# Run command
kubectl exec <pod-name> -n taskir -- env
```

### Port Forwarding

```powershell
# Forward service to localhost
kubectl port-forward svc/nestjs-backend 3000:3000 -n taskir

# Forward pod
kubectl port-forward <pod-name> 3000:3000 -n taskir
```

### Describe Resources

```powershell
# Pod details
kubectl describe pod <pod-name> -n taskir

# Deployment details
kubectl describe deployment nestjs-backend -n taskir

# Events
kubectl get events -n taskir --sort-by='.lastTimestamp'
```

## 📈 Performance

### Resource Usage

```powershell
# Pod resource usage
kubectl top pods -n taskir

# Node resource usage
kubectl top nodes

# Specific pod metrics
kubectl top pod <pod-name> -n taskir --containers
```

## 🔐 Security

### Network Policies

```yaml
# Create network policy
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-internal
  namespace: taskir
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: taskir
```

### Pod Security

```powershell
# Check pod security
kubectl auth can-i --list -n taskir

# Apply pod security policy
kubectl apply -f pod-security-policy.yaml
```

## 🗑️ Cleanup

### Delete Everything

```powershell
# Delete namespace (deletes all resources)
kubectl delete namespace taskir

# Delete EKS cluster
eksctl delete cluster --name taskir-prod --region us-east-1
```

### Delete Specific Resources

```powershell
# Delete deployment
kubectl delete deployment nestjs-backend -n taskir

# Delete service
kubectl delete service nestjs-backend -n taskir

# Delete all deployments
kubectl delete deployments --all -n taskir
```

## 💡 Tips

1. **Use labels** for easier management
2. **Set resource limits** to prevent resource starvation
3. **Enable HPA** for auto-scaling
4. **Use secrets** for sensitive data
5. **Implement health checks** for reliability
6. **Monitor logs** for issues
7. **Use namespaces** for isolation
8. **Regular backups** of persistent data

## 🔗 Useful Links

- [Kubernetes Docs](https://kubernetes.io/docs/)
- [AWS EKS](https://aws.amazon.com/eks/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [Helm Charts](https://helm.sh/)

---

**Version**: v3.0  
**Last Updated**: January 28, 2026
