"""
FastAPI endpoints for Ad Matching Service
"""
from fastapi import APIRouter, HTTPException, BackgroundTasks
from typing import List
import logging
import time

from app.models.schemas import (
    MatchRequest,
    MatchResponse,
    PerformancePredictRequest,
    PerformancePrediction,
    HealthResponse,
    MetricsResponse
)
from app.services.matcher import matcher
from app.config import settings

logger = logging.getLogger(__name__)
router = APIRouter()


@router.post("/match", response_model=MatchResponse)
async def match_ads(
    request: MatchRequest,
    background_tasks: BackgroundTasks
) -> MatchResponse:
    """
    Find best matching ads for user
    
    **Timeout**: 50ms (configurable)
    **Strategies**: collaborative, content_based, performance_based, hybrid
    
    **Returns**:
    - recommendations: List of matched campaigns with scores
    - overall_score: Combined relevance score
    - predicted_ctr: Expected click-through rate
    - match_reasons: Why this ad was recommended
    """
    start_time = time.time()
    
    try:
        # Check timeout
        if (time.time() - start_time) * 1000 > settings.REQUEST_TIMEOUT:
            logger.warning(f"Request {request.request_id} exceeded timeout")
            raise HTTPException(status_code=408, detail="Request timeout")
        
        # Perform ad matching
        response = matcher.match(request)
        
        # Log successful matches
        logger.info(
            f"Matched {len(response.recommendations)} ads for {request.request_id} | "
            f"Strategy: {request.strategy} | "
            f"Time: {response.processing_time_ms:.2f}ms"
        )
        
        # Background task: Record impressions for collaborative filtering
        if response.recommendations and request.user.user_id:
            for rec in response.recommendations:
                background_tasks.add_task(
                    matcher.record_interaction,
                    request.user.user_id,
                    rec.campaign_id,
                    "impression"
                )
        
        return response
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error processing match request: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Internal error: {str(e)}")


@router.post("/recommend", response_model=MatchResponse)
async def recommend_campaigns(
    request: MatchRequest
) -> MatchResponse:
    """
    Get campaign recommendations (alias for /match with hybrid strategy)
    
    **Use Case**: General-purpose ad recommendations
    """
    request.strategy = "hybrid"
    return await match_ads(request, BackgroundTasks())


@router.post("/predict", response_model=PerformancePrediction)
async def predict_performance(
    request: PerformancePredictRequest
) -> PerformancePrediction:
    """
    Predict campaign performance for specific user
    
    **Returns**:
    - predicted_ctr: Expected click-through rate
    - predicted_cvr: Expected conversion rate
    - predicted_revenue: Expected revenue per conversion
    """
    try:
        # Find campaign
        campaign = next(
            (c for c in matcher.campaigns if c["id"] == request.campaign_id),
            None
        )
        
        if not campaign:
            raise HTTPException(status_code=404, detail="Campaign not found")
        
        # Calculate personalized predictions
        _, collab_score, content_score, perf_score = matcher._calculate_hybrid_score(
            request.user,
            campaign,
            "hybrid"
        )
        
        # Adjust base CTR/CVR based on user-campaign match
        personalization_factor = (collab_score + content_score) / 2
        
        predicted_ctr = campaign["ctr"] * (1 + personalization_factor * 0.5)
        predicted_cvr = campaign["cvr"] * (1 + personalization_factor * 0.3)
        predicted_revenue = campaign["avg_revenue_per_conversion"]
        
        # Confidence based on data availability
        ctr_confidence = 0.7 if campaign["impressions"] > 1000 else 0.5
        cvr_confidence = 0.7 if campaign["clicks"] > 100 else 0.5
        
        return PerformancePrediction(
            campaign_id=request.campaign_id,
            predicted_ctr=min(predicted_ctr, 1.0),
            predicted_cvr=min(predicted_cvr, 1.0),
            predicted_revenue=predicted_revenue,
            ctr_confidence=ctr_confidence,
            cvr_confidence=cvr_confidence
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error predicting performance: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Internal error: {str(e)}")


@router.post("/interaction")
async def record_interaction(
    user_id: str,
    campaign_id: str,
    interaction_type: str  # impression, click, conversion
):
    """
    Record user interaction for collaborative filtering
    
    **Use Case**: Update interaction history to improve future recommendations
    """
    try:
        if interaction_type not in ["impression", "click", "conversion"]:
            raise HTTPException(
                status_code=400,
                detail="Invalid interaction type. Must be: impression, click, or conversion"
            )
        
        matcher.record_interaction(user_id, campaign_id, interaction_type)
        
        return {
            "status": "recorded",
            "user_id": user_id,
            "campaign_id": campaign_id,
            "interaction_type": interaction_type
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error recording interaction: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Internal error: {str(e)}")


@router.get("/health", response_model=HealthResponse)
async def health_check() -> HealthResponse:
    """
    Health check endpoint
    
    **Returns**:
    - status: "healthy" or "unhealthy"
    - models_loaded: Whether matching models are loaded
    - redis_connected: Whether Redis is available
    """
    is_healthy = matcher.is_healthy()
    metrics = matcher.get_metrics()
    
    return HealthResponse(
        status="healthy" if is_healthy else "unhealthy",
        models_loaded=matcher.campaign_vectors is not None,
        redis_connected=True,  # TODO: Implement Redis check
        uptime_seconds=metrics["uptime_seconds"]
    )


@router.get("/metrics", response_model=MetricsResponse)
async def get_metrics() -> MetricsResponse:
    """
    Get service metrics
    
    **Returns**:
    - total_requests: Total matching requests processed
    - avg_recommendations: Average recommendations per request
    - avg_processing_time_ms: Average latency
    - cache_hit_rate: Percentage of cached responses
    """
    metrics = matcher.get_metrics()
    return MetricsResponse(**metrics)


@router.post("/reload")
async def reload_campaigns(background_tasks: BackgroundTasks):
    """
    Reload campaigns from backend (admin only)
    
    **Use Case**: Refresh campaign data without restarting service
    """
    def reload():
        logger.info("Reloading campaigns...")
        matcher._load_campaigns()
        logger.info(f"Reloaded {len(matcher.campaigns)} campaigns")
    
    background_tasks.add_task(reload)
    return {"status": "reload started"}


@router.get("/")
async def root():
    """Root endpoint with service information"""
    return {
        "service": "Ad Matching Service",
        "version": "1.0.0",
        "status": "operational",
        "campaigns_loaded": len(matcher.campaigns),
        "endpoints": {
            "match": "POST /api/match - Find matching ads",
            "recommend": "POST /api/recommend - Get recommendations",
            "predict": "POST /api/predict - Predict performance",
            "interaction": "POST /api/interaction - Record interaction",
            "health": "GET /api/health - Health check",
            "metrics": "GET /api/metrics - Performance metrics",
            "reload": "POST /api/reload - Reload campaigns"
        }
    }
