Write-Host "Build and Push Updates for TaskirX.com..." -ForegroundColor Cyan

# 1. Build and push Backend (CORS update)
Write-Host "Building NestJS Backend..."
docker build -t taskirsview/nestjs-backend:latest ./nestjs-backend
if ($LASTEXITCODE -ne 0) { exit 1 }
docker push taskirsview/nestjs-backend:latest
if ($LASTEXITCODE -ne 0) { exit 1 }

# 2. Build and push AI Services (Redis persistence)
Write-Host "Building Ad Matching Service..."
docker build -t taskirsview/ad-matching-service:latest ./python-ai-agents/ad-matching-service
if ($LASTEXITCODE -ne 0) { exit 1 }
docker push taskirsview/ad-matching-service:latest
if ($LASTEXITCODE -ne 0) { exit 1 }

Write-Host "Building Bid Optimization Service..."
docker build -t taskirsview/bid-optimization-service:latest ./python-ai-agents/bid-optimization-service
if ($LASTEXITCODE -ne 0) { exit 1 }
docker push taskirsview/bid-optimization-service:latest
if ($LASTEXITCODE -ne 0) { exit 1 }

# 3. Update Kubernetes Deployments to use new images
kubectl set image deployment/nestjs-backend nestjs=taskirsview/nestjs-backend:latest -n taskir
kubectl set image deployment/ad-matching ad-matching=taskirsview/ad-matching-service:latest -n taskir
kubectl set image deployment/bid-optimization bid-optimization=taskirsview/bid-optimization-service:latest -n taskir

# 4. Restart Deployments
Write-Host "Restarting Kubernetes Deployments..."
kubectl rollout restart deployment nestjs-backend -n taskir
kubectl rollout restart deployment ad-matching -n taskir
kubectl rollout restart deployment bid-optimization -n taskir
kubectl rollout restart deployment next-dashboard -n taskir

Write-Host "✅ Update Complete! Your cluster is now using TaskirX.com and persisted AI models." -ForegroundColor Green
