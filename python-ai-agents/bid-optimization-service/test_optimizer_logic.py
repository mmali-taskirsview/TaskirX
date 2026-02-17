
import sys
import os
import time
from datetime import datetime
import asyncio

# Add the current directory to sys.path so we can import app
sys.path.insert(0, os.path.abspath(os.path.dirname(__file__)))

from app.models.schemas import (
    BidOptimizationRequest,
    OptimizationContext,
    CampaignPerformance,
    BudgetStatus,
    BidStrategy,
    AuctionType
)
from app.services.optimizer import BidOptimizer

def test_bid_optimization():
    print("Initializing BidOptimizer...")
    optimizer = BidOptimizer()
    
    # Create a mock request
    context = OptimizationContext(
        campaign_id="test-campaign-123",
        base_bid=2.50,
        performance=CampaignPerformance(
            campaign_id="test-campaign-123",
            impressions=1000,
            clicks=50,
            conversions=5
        ),
        budget_status=BudgetStatus(
            campaign_id="test-campaign-123",
            daily_budget=100.0,
            today_spend=20.0,
            remaining_daily=80.0
        ),
        hour_of_day=14,
        day_of_week=2
    )

    request = BidOptimizationRequest(
        request_id="req-001",
        context=context,
        strategy=BidStrategy.MAXIMIZE_CONVERSIONS,
        target_cpa=10.0
    )

    print(f"Optimizing bid for campaign {context.campaign_id} with base bid ${context.base_bid}")
    
    # Run optimization
    recommendation = optimizer.optimize_bid(request)
    
    print("\n--- Bid Recommendation ---")
    print(f"Request ID: {recommendation.request_id}")
    print(f"Recommended Bid: ${recommendation.recommended_bid:.2f}")
    print(f"Multiplier: {recommendation.bid_multiplier:.2f}x")
    print(f"Confidence: {recommendation.confidence:.2f}")
    print(f"Reasoning: {recommendation.reasoning}")
    print(f"Processing Time: {recommendation.processing_time_ms:.2f}ms")
    
    # Validate result
    assert recommendation.recommended_bid > 0
    assert 0.5 <= recommendation.bid_multiplier <= 2.0
    
    print("\n✅ Bid Optimization Test Passed!")

def test_budget_pacing():
    from app.models.schemas import BudgetPacingRequest, PacingStrategy
    
    print("\nTesting Budget Pacing...")
    optimizer = BidOptimizer()
    
    pacing_request = BudgetPacingRequest(
        request_id="pace-001",
        campaign_id="test-campaign-123",
        budget_status=BudgetStatus(
            campaign_id="test-campaign-123",
            daily_budget=100.0,
            today_spend=40.0, # 40% spent
            daily_remaining=60.0, # 100 - 40 = 60 remaining
            expected_daily_spend=100.0
        ),
        hours_remaining_today=10.0, # 14 hours passed (approx)
        pacing_strategy=PacingStrategy.EVEN
    )
    
    # Run pacing
    rec = optimizer.calculate_budget_pacing(pacing_request)

    print("\n--- Pacing Recommendation ---")
    print(f"Health: {rec.pacing_health}")
    print(f"Recommended Hourly Spend: ${rec.recommended_hourly_spend:.2f}")
    print(f"Predicted EOD Spend: ${rec.predicted_eod_spend:.2f}")
    
    assert rec.recommended_hourly_spend > 0
    print("\n✅ Budget Pacing Test Passed!")

if __name__ == "__main__":
    try:
        test_bid_optimization()
        test_budget_pacing()
    except Exception as e:
        print(f"\n❌ Test Failed: {e}")
        import traceback
        traceback.print_exc()
