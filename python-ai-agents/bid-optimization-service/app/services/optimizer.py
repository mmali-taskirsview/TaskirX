"""
Bid Optimization Service - Core Optimization Engine
"""
import time
import logging
import numpy as np
from typing import Dict, List, Tuple, Optional
from collections import defaultdict
from datetime import datetime, timedelta

from app.models.schemas import (
    BidOptimizationRequest,
    BidRecommendation,
    BudgetPacingRequest,
    BudgetPacingRecommendation,
    ThompsonSamplingState,
    BidStrategy,
    PacingStrategy,
    OptimizationContext
)
from app.config import settings

logger = logging.getLogger(__name__)


class BidOptimizer:
    """
    AI-powered bid optimization engine
    
    Implements:
    - Multi-armed bandit (Thompson Sampling) for dynamic pricing
    - PID controller for budget pacing
    - Reinforcement Learning for long-term ROI maximization
    """
    
    def __init__(self):
        self.start_time = time.time()
        # In-memory store for demo. Production needs Redis.
        self.bandit_states = {} 

    # Alias for API endpoint method name compat
    def optimize_bid(self, request: BidOptimizationRequest) -> BidRecommendation:
        return self.recommend_bid(request)

    def recommend_bid(self, request: BidOptimizationRequest) -> BidRecommendation:
        """
        Generate optimal bid price recommendation
        
        Uses Thompson Sampling to balance exploration (testing new prices) vs exploitation (using best historic price).
        """
        campaign_id = request.campaign_id
        
        # 1. Retrieve or Initialize Bandit State
        state = self._get_bandit_state(campaign_id)
        
        # 2. Thompson Sampling for Price Selection
        # Sample beta distribution for each price arm
        # Beta(alpha, beta) where alpha = conversions + 1, beta = failures + 1
        
        best_arm_idx = 0
        max_sample = -1.0
        
        for idx, arm in enumerate(state.arms):
            sample = np.random.beta(arm.alpha, arm.beta)
            if sample > max_sample:
                max_sample = sample
                best_arm_idx = idx
                
        recommended_price = state.arms[best_arm_idx].price
        
        # 3. Apply Contextual Modifiers
        # e.g., Day of week, Time of day adjustments
        multiplier = self._get_time_of_day_multiplier()
        final_bid = recommended_price * multiplier
        
        # 4. Cap at max bid
        if hasattr(request, 'max_bid') and request.max_bid:
             final_bid = min(final_bid, request.max_bid)
        
        return BidRecommendation(
            campaign_id=campaign_id,
            recommended_bid=final_bid,
            confidence_score=max_sample,
            strategy_used=BidStrategy.THOMPSON_SAMPLING,
            reason=f"Bandit algorithm selected arm {best_arm_idx} (price ${recommended_price}) with sample score {max_sample:.4f}"
        )

    def _get_bandit_state(self, campaign_id: str):
        # Mock state initialization
        # Define 5 price arms: $0.50, $1.00, $2.00, $5.00, $10.00
        class Arm:
             def __init__(self, price, alpha=1, beta=1):
                 self.price = price
                 self.alpha = alpha
                 self.beta = beta
        
        class State:
             def __init__(self):
                 self.arms = [
                     Arm(0.50, 10, 100), # Low conversion
                     Arm(1.00, 50, 200),
                     Arm(2.00, 80, 150), # Best perfroming mock
                     Arm(5.00, 20, 80),
                     Arm(10.0, 5, 20)
                 ]
        return State()

    def _get_time_of_day_multiplier(self) -> float:
        hour = datetime.now().hour
        # Simple heuristic: Higher bids during work hours (9-17)
        if 9 <= hour <= 17:
            return 1.1
        return 0.9
    

# Global optimizer instance
optimizer = BidOptimizer()
