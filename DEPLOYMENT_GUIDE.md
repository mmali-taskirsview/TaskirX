# TaskirX V3 - Production Deployment Guide

## Prerequisites

- Docker Engine 24+ and Docker Compose v2+
- Kubernetes Cluster (EKS/GKE/AKS/Self-Hosted)
- Terraform installed
- Domain name with SSL certificate

---

## Architecture Overview

TaskirX V3 is a polyglot microservices platform:
- **Backend**: NestJS (Business Logic)
- **Bidding Engine**: Go (High Performance RTB)
- **AI Agents**: Python (Fraud, Ad Matching, Optimization)
- **Database**: PostgreSQL & Redis
- **Analytics**: ClickHouse

## Option 1: Docker Swarm / Compose (Small Scale)

Recommended for staging or simple production deployments.

### 1. Setup Server

```bash
# Ubuntu 22.04 LTS
# Open ports: 80, 443, 3000, 3001, 8080 (optional)

# Install Docker
curl -fsSL https://get.docker.com | sh
```

### 2. Deploy Platform

```bash
# Clone
git clone https://github.com/your-org/taskirx.git
cd taskirx

# Start all services
docker-compose up -d --build

# Verify
docker-compose ps
```

## Option 2: Kubernetes (Production)

### 1. Provision Infrastructure

Use the provided Terraform scripts to set up EKS/GKE.

```bash
cd terraform
terraform init
terraform apply -var="cluster_name=taskirx-prod"
```

### 2. Deploy Manifests

Apply the standard Kubernetes configurations.

```bash
# Apply Configs and Secrets first
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/postgres-deployment.yaml
kubectl apply -f k8s/redis-deployment.yaml

# Deploy Applications
kubectl apply -f k8s/nestjs-deployment.yaml
kubectl apply -f k8s/go-bidding-deployment.yaml
kubectl apply -f k8s/python-services-deployment.yaml

# Expose via Ingress
kubectl apply -f k8s/ingress.yaml
```

### 3. Monitoring & Observability

We use Prometheus and Grafana for full stack visibility.

```bash
# Deploy Monitoring Stack (if not using Helm)
kubectl apply -f monitoring/
```

Access dashboards:
- **Grafana**: http://grafana.yourdomain.com
- **Prometheus**: http://prometheus.yourdomain.com

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/taskirx /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 4. SSL Certificate (Let's Encrypt)

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d api.yourdomain.com
sudo systemctl restart nginx
```

### 5. MongoDB Atlas (Recommended)

1. Go to [MongoDB Atlas](https://www.mongodb.com/cloud/atlas)
2. Create cluster (M10 or larger for production)
3. Create database user
4. Whitelist IP addresses
5. Get connection string
6. Update .env with Atlas URI

### 6. Post-Deployment Setup

#### Database Seeding
The automated deployment creates the database schema but does not populate it with initial users.
Run the remote seeding tool to create the default Admin, Advertiser, and Publisher accounts:

```powershell
.\scripts\seed-remote-db.ps1
```

#### Monitoring Access
To access Grafana (Monitoring Dashboard):
```bash
# Verify the Monitoring Namespace
kubectl get pods -n monitoring

# Port-forward Grafana to your local machine
kubectl port-forward svc/kube-prometheus-stack-grafana 3000:80 -n monitoring
```
Access Grafana at `http://localhost:3000` (Default: admin/prom-operator).

---

## Option 2: Deploy with Docker

### Dockerfile

```dockerfile
FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .

EXPOSE 3000

CMD ["node", "server.js"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  backend:
    build: ./backend
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - MONGODB_URI=mongodb://mongo:27017/taskirx
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - mongo
    restart: unless-stopped

  mongo:
    image: mongo:6.0
    ports:
      - "27017:27017"
    volumes:
      - mongo-data:/data/db
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - backend
    restart: unless-stopped

volumes:
  mongo-data:
```

### Deploy with Docker

```bash
# Build and run
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

---

## Option 3: Deploy to Google Cloud Platform

### 1. Setup Google Cloud Run

```bash
# Install gcloud CLI
# https://cloud.google.com/sdk/docs/install

# Login
gcloud auth login
gcloud config set project your-project-id

# Build and push Docker image
gcloud builds submit --tag gcr.io/your-project-id/taskirx

# Deploy to Cloud Run
gcloud run deploy taskirx \
  --image gcr.io/your-project-id/taskirx \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars NODE_ENV=production,MONGODB_URI=your-atlas-uri \
  --memory 2Gi \
  --cpu 2
```

### 2. Setup MongoDB Atlas

Use MongoDB Atlas for database (required for Cloud Run)

---

## Option 4: Deploy to Heroku

```bash
# Install Heroku CLI
# https://devcenter.heroku.com/articles/heroku-cli

# Login
heroku login

# Create app
heroku create your-taskirx

# Add MongoDB addon
heroku addons:create mongolab:sandbox

# Set environment variables
heroku config:set NODE_ENV=production
heroku config:set JWT_SECRET=your-secret

# Deploy
git push heroku main

# View logs
heroku logs --tail
```

---

## Option 3: Oracle Cloud Infrastructure (OCI - Enterprise)

Recommended for enterprise deployments requiring high-performance computing, cost-effective networking, and vector search.

### 1. Prerequisites
- OCI Account & Tenancy
- Cloudflare Account
- Pinecone Account
- [Setup Guide](deploy-oci.md)

### 2. Validation
Run the validator to ensure your environment is ready:
```powershell
.\scripts\validate-oci-setup.ps1
```

### 3. Deployment
```powershell
.\scripts\deploy-to-oci.ps1 -Environment "oci" -Action "apply" -Registry "iad.ocir.io/tenancy/taskir"
```

---

## Post-Deployment Checklist

### Security

- [ ] Change all default passwords and secrets
- [ ] Enable firewall (ufw/security groups)
- [ ] Setup SSL/TLS certificates
- [ ] Enable rate limiting
- [ ] Configure CORS properly
- [ ] Setup database backups
- [ ] Enable MongoDB authentication
- [ ] Setup VPC/private networking

### Monitoring

- [ ] Setup application monitoring (DataDog, New Relic)
- [ ] Configure error tracking (Sentry)
- [ ] Setup uptime monitoring (Pingdom, UptimeRobot)
- [ ] Configure log aggregation (Loggly, Papertrail)
- [ ] Setup alerts for errors/downtime

### Performance

- [ ] Enable database indexing
- [ ] Configure caching (Redis)
- [ ] Setup CDN for static assets
- [ ] Enable gzip compression
- [ ] Optimize database queries
- [ ] Load test the API

### Backup & Recovery

- [ ] Setup automated database backups (daily)
- [ ] Test restore procedures
- [ ] Document recovery process
- [ ] Setup off-site backup storage

### Documentation

- [ ] Update API documentation with production URLs
- [ ] Document deployment process
- [ ] Create runbook for common issues
- [ ] Document scaling procedures

---

## Scaling Considerations

### Horizontal Scaling

```bash
# PM2 cluster mode (single server)
pm2 start server.js -i max --name taskirx

# Multiple servers with load balancer
# Use AWS ELB, Google Cloud Load Balancer, or Nginx
```

### Database Scaling

- Use MongoDB sharding for large datasets
- Read replicas for read-heavy workloads
- Implement caching layer (Redis)

### CDN for SDKs

Upload SDK files to CDN:
- AWS CloudFront
- Google Cloud CDN
- Cloudflare

---

## Monitoring Commands

```bash
# Check server status
pm2 status

# View logs
pm2 logs taskirx

# Monitor resources
pm2 monit

# Restart server
pm2 restart taskirx

# View MongoDB status
sudo systemctl status mongod

# Check Nginx
sudo systemctl status nginx
sudo nginx -t

# View system resources
htop
df -h
free -m
```

---

## Troubleshooting

### Server won't start

```bash
# Check logs
pm2 logs taskirx --lines 100

# Check MongoDB connection
mongosh

# Check port availability
sudo netstat -tulpn | grep 3000
```

### High memory usage

```bash
# Check process memory
pm2 monit

# Restart application
pm2 restart taskirx

# Check MongoDB memory
mongosh --eval "db.serverStatus().mem"
```

### Database connection issues

```bash
# Check MongoDB status
sudo systemctl status mongod

# Check MongoDB logs
sudo tail -f /var/log/mongodb/mongod.log

# Restart MongoDB
sudo systemctl restart mongod
```

---

## Environment Variables Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| NODE_ENV | Environment | development | Yes |
| PORT | Server port | 3000 | Yes |
| MONGODB_URI | MongoDB connection string | - | Yes |
| JWT_SECRET | JWT secret key | - | Yes |
| JWT_EXPIRES_IN | Token expiration | 7d | No |
| CORS_ORIGIN | Allowed origins | * | No |
| RATE_LIMIT_WINDOW_MS | Rate limit window | 900000 | No |
| RATE_LIMIT_MAX_REQUESTS | Max requests per window | 100 | No |
| LOG_LEVEL | Logging level | info | No |

---

## Support

For deployment support:
- Email: support@taskirx.com
- Docs: https://docs.taskirx.com
- GitHub Issues: https://github.com/your-org/taskirx/issues
