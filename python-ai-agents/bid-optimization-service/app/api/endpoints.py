"""
FastAPI endpoints for Bid Optimization Service
"""
from fastapi import APIRouter, HTTPException, BackgroundTasks
import logging
import time

from app.models.schemas import (
    BidOptimizationRequest,
    BidRecommendation,
    BudgetPacingRequest,
    BudgetPacingRecommendation,
    HealthResponse,
    MetricsResponse
)
from app.services.optimizer import optimizer
from app.config import settings

logger = logging.getLogger(__name__)
router = APIRouter()


@router.post("/optimize", response_model=BidRecommendation)
async def optimize_bid(
    request: BidOptimizationRequest,
    background_tasks: BackgroundTasks
) -> BidRecommendation:
    """
    Get optimal bid recommendation
    
    **Timeout**: 50ms (configurable)
    **Strategies**: maximize_clicks, maximize_conversions, target_cpa, target_roas, manual
    
    **Uses**:
    - Thompson Sampling (Multi-Armed Bandit)
    - Performance-based adjustments
    - Budget-aware bidding
    - Time-of-day optimization
    - Competition-aware bidding
    
    **Returns**:
    - recommended_bid: Optimal bid price
    - bid_multiplier: Adjustment factor (0.5x - 2.0x)
    - confidence: Model confidence (0-1)
    - reasoning: Explanation of adjustments
    - expected_outcomes: Win rate, CTR, CVR, ROI predictions
    """
    start_time = time.time()
    
    try:
        # Check timeout
        if (time.time() - start_time) * 1000 > settings.REQUEST_TIMEOUT:
            logger.warning(f"Request {request.request_id} exceeded timeout")
            raise HTTPException(status_code=408, detail="Request timeout")
        
        # Optimize bid
        response = optimizer.optimize_bid(request)
        
        # Log optimization
        logger.info(
            f"Optimized bid for {request.context.campaign_id}: "
            f"Base: ${request.context.base_bid:.2f} → "
            f"Recommended: ${response.recommended_bid:.2f} "
            f"({response.bid_multiplier:.2f}x) | "
            f"Strategy: {request.strategy}"
        )
        
        return response
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error optimizing bid: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Internal error: {str(e)}")


@router.post("/pace", response_model=BudgetPacingRecommendation)
async def calculate_pacing(
    request: BudgetPacingRequest
) -> BudgetPacingRecommendation:
    """
    Calculate budget pacing recommendations
    
    **Strategies**: even, aggressive, conservative, asap
    
    **Returns**:
    - recommended_hourly_spend: How much to spend per hour
    - bid_adjustment_factor: Bid multiplier for pacing
    - should_pause/increase/decrease: Action recommendations
    - pacing_health: healthy, underspending, overspending, depleted
    - predicted_eod_spend: Expected end-of-day spend
    """
    try:
        response = optimizer.calculate_budget_pacing(request)
        
        logger.info(
            f"Budget pacing for {request.campaign_id}: "
            f"Health: {response.pacing_health} | "
            f"Recommended: ${response.recommended_hourly_spend:.2f}/hour"
        )
        
        return response
        
    except Exception as e:
        logger.error(f"Error calculating pacing: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Internal error: {str(e)}")


@router.post("/feedback")
async def record_feedback(
    campaign_id: str,
    bid_multiplier: float,
    won: bool,
    converted: bool = False
):
    """
    Record bid outcome for Thompson Sampling learning
    
    **Use Case**: Update bandit state after auction result
    
    **Parameters**:
    - campaign_id: Campaign identifier
    - bid_multiplier: Multiplier that was used
    - won: Whether the auction was won
    - converted: Whether the impression converted
    """
    try:
        # Success = won auction AND converted (for maximize_conversions)
        # Or just won (for maximize_clicks)
        success = won and converted if converted else won
        
        optimizer.update_bandit(campaign_id, bid_multiplier, success)
        
        return {
            "status": "recorded",
            "campaign_id": campaign_id,
            "bid_multiplier": bid_multiplier,
            "success": success
        }
        
    except Exception as e:
        logger.error(f"Error recording feedback: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Internal error: {str(e)}")


@router.get("/strategy/{campaign_id}")
async def get_strategy(campaign_id: str):
    """
    Get current bidding strategy for campaign
    
    **Returns**: Thompson Sampling state for all bid multipliers
    """
    try:
        states = optimizer.bandit_states.get(campaign_id, {})
        
        strategy_info = {
            "campaign_id": campaign_id,
            "multipliers": {},
            "best_multiplier": None,
            "exploration_rate": settings.EXPLORATION_RATE
        }
        
        best_rate = 0
        best_mult = None
        
        for mult_str, state in states.items():
            mult = float(mult_str)
            strategy_info["multipliers"][mult] = {
                "trials": state.trials,
                "successes": state.successes,
                "success_rate": state.estimated_success_rate,
                "alpha": state.alpha,
                "beta": state.beta
            }
            
            if state.estimated_success_rate > best_rate:
                best_rate = state.estimated_success_rate
                best_mult = mult
        
        strategy_info["best_multiplier"] = best_mult
        
        return strategy_info
        
    except Exception as e:
        logger.error(f"Error getting strategy: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Internal error: {str(e)}")


@router.get("/health", response_model=HealthResponse)
async def health_check() -> HealthResponse:
    """
    Health check endpoint
    
    **Returns**:
    - status: "healthy" or "unhealthy"
    - optimizer_ready: Whether optimizer is initialized
    - redis_connected: Whether Redis is available
    """
    is_healthy = optimizer.is_healthy()
    metrics = optimizer.get_metrics()
    
    return HealthResponse(
        status="healthy" if is_healthy else "unhealthy",
        optimizer_ready=True,
        redis_connected=True,  # TODO: Implement Redis check
        uptime_seconds=metrics["uptime_seconds"]
    )


@router.get("/metrics", response_model=MetricsResponse)
async def get_metrics() -> MetricsResponse:
    """
    Get service metrics
    
    **Returns**:
    - total_requests: Total optimization requests
    - avg_bid_multiplier: Average bid adjustment
    - avg_processing_time_ms: Average latency
    - cache_hit_rate: Cache hit percentage
    """
    metrics = optimizer.get_metrics()
    return MetricsResponse(**metrics)


@router.get("/")
async def root():
    """Root endpoint with service information"""
    return {
        "service": "Bid Optimization Service",
        "version": "1.0.0",
        "status": "operational",
        "algorithms": ["Thompson Sampling", "Budget Pacing", "ROI Optimization"],
        "endpoints": {
            "optimize": "POST /api/optimize - Get optimal bid",
            "pace": "POST /api/pace - Budget pacing recommendation",
            "feedback": "POST /api/feedback - Record auction outcome",
            "strategy": "GET /api/strategy/{campaign_id} - View learning state",
            "health": "GET /api/health - Health check",
            "metrics": "GET /api/metrics - Performance metrics"
        }
    }
