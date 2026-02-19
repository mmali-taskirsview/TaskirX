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

# 3. Build and Push Fraud Detection Service (New in Phase 10)
Write-Host "Building Fraud Detection Service..."
docker build -t taskirsview/fraud-detection-service:latest ./python-ai-agents/fraud-detection-service
if ($LASTEXITCODE -ne 0) { exit 1 }
docker push taskirsview/fraud-detection-service:latest
if ($LASTEXITCODE -ne 0) { exit 1 }

# 4. Build and Push Go Bidding Engine (Updated Budget Logic)
Write-Host "Building Go Bidding Engine..."
docker build -t taskirsview/go-bidding-engine:latest ./go-bidding-engine
if ($LASTEXITCODE -ne 0) { exit 1 }
docker push taskirsview/go-bidding-engine:latest
if ($LASTEXITCODE -ne 0) { exit 1 }

# 5. Build and Push Next.js Dashboard (Updated UI)
Write-Host "Building Next.js Dashboard..."
docker build -t taskirsview/next-dashboard:latest ./next-dashboard
if ($LASTEXITCODE -ne 0) { exit 1 }
docker push taskirsview/next-dashboard:latest
if ($LASTEXITCODE -ne 0) { exit 1 }

# 6. Update Kubernetes Deployments to use new images
kubectl set image deployment/nestjs-backend nestjs=taskirsview/nestjs-backend:latest -n taskir
kubectl set image deployment/ad-matching ad-matching=taskirsview/ad-matching-service:latest -n taskir
kubectl set image deployment/bid-optimization bid-optimization=taskirsview/bid-optimization-service:latest -n taskir
kubectl set image deployment/fraud-detection fraud-detection=taskirsview/fraud-detection-service:latest -n taskir
kubectl set image deployment/go-bidding go-bidding=taskirsview/go-bidding-engine:latest -n taskir
kubectl set image deployment/next-dashboard next-dashboard=taskirsview/next-dashboard:latest -n taskir

# 7. Restart Deployments
Write-Host "Restarting Kubernetes Deployments..."
kubectl rollout restart deployment nestjs-backend -n taskir
kubectl rollout restart deployment ad-matching -n taskir
kubectl rollout restart deployment bid-optimization -n taskir
kubectl rollout restart deployment fraud-detection -n taskir
kubectl rollout restart deployment go-bidding -n taskir
kubectl rollout restart deployment next-dashboard -n taskir

Write-Host "✅ Update Complete! Your cluster is now using TaskirX.com and persisted AI models." -ForegroundColor Green
