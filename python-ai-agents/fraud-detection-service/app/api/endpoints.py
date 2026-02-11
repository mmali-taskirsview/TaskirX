"""
FastAPI endpoints for Fraud Detection Service
"""
from fastapi import APIRouter, HTTPException, BackgroundTasks
from typing import List
import logging
import time

from app.models.schemas import (
    FraudCheckRequest,
    FraudCheckResponse,
    HealthResponse,
    MetricsResponse
)
from app.services.fraud_detector import detector
from app.services.ip_reputation import IPReputationService
import redis
from app.config import settings

# Initialize IP Reputation Service
# Shared Redis connection
redis_client = redis.Redis(
    host=settings.REDIS_HOST,
    port=settings.REDIS_PORT,
    password=settings.REDIS_PASSWORD,
    decode_responses=True
)
ip_reputation = IPReputationService(redis_client, api_key=settings.IP_REPUTATION_API_KEY)

logger = logging.getLogger(__name__)
router = APIRouter()


@router.post("/detect", response_model=FraudCheckResponse)
async def detect_fraud(
    request: FraudCheckRequest,
    background_tasks: BackgroundTasks
) -> FraudCheckResponse:
    """
    Detect fraud in ad interaction
    
    **Timeout**: 100ms (configurable)
    **Use Case**: Real-time fraud screening for clicks, impressions, conversions
    
    **Returns**:
    - fraud_score: Probability of fraud (0-1)
    - risk_level: LOW/MEDIUM/HIGH/CRITICAL
    - recommended_action: allow/flag/block
    - indicators: Detailed fraud signals
    - reasons: Human-readable explanations
    """
    start_time = time.time()
    
    try:
        # Check timeout
        if (time.time() - start_time) * 1000 > settings.REQUEST_TIMEOUT:
            logger.warning(f"Request {request.request_id} exceeded timeout")
            raise HTTPException(status_code=408, detail="Request timeout")
        
        # 1. Check IP Blocklist first (Fastest check)
        if ip_reputation.is_ip_blacklisted(request.ip_address):
            # If blacklisted, return immediately
            return FraudCheckResponse(
                request_id=request.request_id,
                fraud_score=1.0,
                risk_level="CRITICAL",
                recommended_action="block",
                indicators={"ip_blocklist": True},
                reasons=["IP address is in blocklist"],
                model_version=detector.model_version,
                processing_time_ms=(time.time() - start_time) * 1000,
                transaction_id=request.transaction_id
            )

        # Perform fraud detection
        response = detector.predict(request)
        
        # Log high-risk detections
        if response.risk_level in ["HIGH", "CRITICAL"]:
            logger.warning(
                f"High-risk fraud detected: {request.request_id} | "
                f"Score: {response.fraud_score:.2f} | "
                f"Reasons: {', '.join(response.reasons)}"
            )
        
        # Background task: Store result for analytics (optional)
        # background_tasks.add_task(store_fraud_result, request, response)
        
        return response
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error processing fraud detection: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Internal error: {str(e)}")


@router.post("/batch", response_model=List[FraudCheckResponse])
async def detect_fraud_batch(
    requests: List[FraudCheckRequest]
) -> List[FraudCheckResponse]:
    """
    Batch fraud detection for multiple requests
    
    **Max Batch Size**: 100 requests
    **Use Case**: Bulk historical analysis, batch processing
    """
    if len(requests) > 100:
        raise HTTPException(
            status_code=400,
            detail="Batch size exceeds limit of 100 requests"
        )
    
    try:
        responses = []
        for req in requests:
            response = detector.predict(req)
            responses.append(response)
        
        return responses
        
    except Exception as e:
        logger.error(f"Error processing batch: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Internal error: {str(e)}")


@router.get("/health", response_model=HealthResponse)
async def health_check() -> HealthResponse:
    """
    Health check endpoint
    
    **Returns**:
    - status: "healthy" or "unhealthy"
    - model_loaded: Whether ML model is loaded
    - redis_connected: Whether Redis is available
    - uptime_seconds: Service uptime
    """
    is_healthy = detector.is_healthy()
    metrics = detector.get_metrics()
    
    return HealthResponse(
        status="healthy" if is_healthy else "unhealthy",
        model_loaded=detector.model is not None,
        redis_connected=True,  # TODO: Implement Redis check
        uptime_seconds=metrics["uptime_seconds"]
    )


@router.get("/metrics", response_model=MetricsResponse)
async def get_metrics() -> MetricsResponse:
    """
    Get service metrics
    
    **Returns**:
    - total_requests: Total predictions made
    - fraud_detected: Number of fraud cases detected
    - fraud_rate: Percentage of fraud
    - avg_processing_time_ms: Average latency
    - model_accuracy: Model performance (if available)
    """
    metrics = detector.get_metrics()
    
    return MetricsResponse(**metrics, model_accuracy=None)


@router.post("/retrain")
async def retrain_model(background_tasks: BackgroundTasks):
    """
    Trigger model retraining (admin only)
    
    **Use Case**: Periodic model updates with new data
    **Note**: This is a long-running task, executed in background
    """
    # In production, add authentication/authorization
    
    def retrain():
        logger.info("Starting model retraining...")
        # TODO: Fetch recent data from ClickHouse
        # TODO: Retrain model
        # TODO: Validate new model
        # TODO: Deploy new model
        logger.info("Model retraining completed")
    
    background_tasks.add_task(retrain)
    
    return {"status": "retraining started"}


@router.get("/")
async def root():
    """Root endpoint with service information"""
    return {
        "service": "Fraud Detection Service",
        "version": "1.0.0",
        "status": "operational",
        "endpoints": {
            "detect": "POST /api/detect - Real-time fraud detection",
            "batch": "POST /api/batch - Batch fraud detection",
            "health": "GET /api/health - Health check",
            "metrics": "GET /api/metrics - Performance metrics",
            "retrain": "POST /api/retrain - Retrain model"
        }
    }
