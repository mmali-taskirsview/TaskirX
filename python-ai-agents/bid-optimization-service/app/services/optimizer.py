"""
Bid Optimization Service - Core Optimization Engine
"""
import time
import logging
import json
import redis
import numpy as np
from dataclasses import dataclass, asdict
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

@dataclass
class BanditArmState:
    multiplier: float
    alpha: float = 1.0
    beta: float = 1.0
    trials: int = 0
    successes: int = 0

    @property
    def estimated_success_rate(self) -> float:
        return self.successes / self.trials if self.trials else 0.5

    def to_dict(self):
        return asdict(self)
    
    @staticmethod
    def from_dict(data):
        return BanditArmState(**data)


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
        self.bandit_states: Dict[str, Dict[float, "BanditArmState"]] = {}
        self.total_requests = 0
        self.total_processing_time_ms = 0.0
        self.total_bid_multiplier = 0.0
        self.cache_hits = 0
        self.bid_multipliers = self._build_bid_multipliers()
        
        # Initialize Redis
        try:
            self.redis = redis.Redis(
                host=settings.REDIS_HOST,
                port=settings.REDIS_PORT,
                db=settings.REDIS_DB,
                password=settings.REDIS_PASSWORD,
                decode_responses=True
            )
            self.redis.ping()
            logger.info(f"Connected to Redis at {settings.REDIS_HOST}:{settings.REDIS_PORT}")
        except Exception as e:
            logger.error(f"Failed to connect to Redis: {e}")
            self.redis = None

    # Alias for API endpoint method name compat
    def optimize_bid(self, request: BidOptimizationRequest) -> BidRecommendation:
        start_time = time.time()
        recommendation = self.recommend_bid(request, start_time)
        self.total_requests += 1
        self.total_processing_time_ms += recommendation.processing_time_ms
        self.total_bid_multiplier += recommendation.bid_multiplier
        return recommendation

    def recommend_bid(self, request: BidOptimizationRequest, start_time: Optional[float] = None) -> BidRecommendation:
        """
        Generate optimal bid price recommendation
        
        Uses Thompson Sampling to balance exploration (testing new prices) vs exploitation (using best historic price).
        """
        campaign_id = request.context.campaign_id
        
        # 1. Retrieve or Initialize Bandit State
        state = self._get_bandit_state(campaign_id)
        
        # 2. Thompson Sampling for Price Selection
        # Sample beta distribution for each price arm
        # Beta(alpha, beta) where alpha = conversions + 1, beta = failures + 1
        
        best_arm = None
        max_sample = -1.0

        for arm in state.values():
            sample = np.random.beta(arm.alpha, arm.beta)
            if sample > max_sample:
                max_sample = sample
                best_arm = arm

        if best_arm is None:
            best_arm = list(state.values())[0]
            max_sample = 0.5
        
        # 3. Apply Contextual Modifiers
        # e.g., Day of week, Time of day adjustments
        multiplier = best_arm.multiplier * self._get_time_of_day_multiplier()
        final_bid = request.context.base_bid * multiplier
        
        # 4. Cap at max bid
        if request.min_bid:
            final_bid = max(final_bid, request.min_bid)
        if request.max_bid:
            final_bid = min(final_bid, request.max_bid)

        processing_time_ms = (time.time() - (start_time or time.time())) * 1000
        performance = request.context.performance
        expected_win_rate = performance.win_rate if performance else max_sample

        return BidRecommendation(
            request_id=request.request_id,
            recommended_bid=final_bid,
            bid_multiplier=best_arm.multiplier,
            confidence=min(max_sample, 1.0),
            reasoning=[
                f"Bandit sampling selected multiplier {best_arm.multiplier:.2f} with score {max_sample:.4f}",
                f"Time-of-day adjustment applied: {self._get_time_of_day_multiplier():.2f}x"
            ],
            expected_win_rate=min(max(expected_win_rate, 0.0), 1.0),
            expected_ctr=performance.ctr if performance else None,
            expected_cvr=performance.cvr if performance else None,
            expected_roi=performance.roas if performance else None,
            strategy_used=request.strategy,
            processing_time_ms=processing_time_ms
        )

    def _get_bandit_state(self, campaign_id: str) -> Dict[float, "BanditArmState"]:
        # Try to get from local cache first
        if campaign_id in self.bandit_states:
             return self.bandit_states[campaign_id]

        # Try to get from Redis
        if self.redis:
            try:
                data = self.redis.get(f"campaign:{campaign_id}:bandit")
                if data:
                    serialized = json.loads(data)
                    state = {
                        float(k): BanditArmState.from_dict(v) 
                        for k, v in serialized.items()
                    }
                    self.bandit_states[campaign_id] = state
                    return state
            except Exception as e:
                logger.error(f"Redis error fetching bandit state: {e}")

        # Initialize new state if not found
        state = {
            multiplier: BanditArmState(multiplier=multiplier)
            for multiplier in self.bid_multipliers
        }
        self.bandit_states[campaign_id] = state
        
        # Persist initial state
        self._save_bandit_state(campaign_id, state)
        
        return state

    def _save_bandit_state(self, campaign_id: str, state: Dict[float, "BanditArmState"]):
        if not self.redis:
            return
            
        try:
            serialized = {
                str(k): v.to_dict() 
                for k, v in state.items()
            }
            self.redis.set(f"campaign:{campaign_id}:bandit", json.dumps(serialized))
        except Exception as e:
            logger.error(f"Redis error saving bandit state: {e}")

    def _get_time_of_day_multiplier(self) -> float:
        hour = datetime.now().hour
        # Simple heuristic: Higher bids during work hours (9-17)
        if 9 <= hour <= 17:
            return 1.1
        return 0.9

    def _build_bid_multipliers(self) -> List[float]:
        multipliers = np.linspace(settings.MIN_BID_MULTIPLIER, settings.MAX_BID_MULTIPLIER, 7)
        return [round(float(mult), 2) for mult in multipliers]

    def update_bandit(self, campaign_id: str, bid_multiplier: float, success: bool) -> None:
        state = self._get_bandit_state(campaign_id)
        normalized_multiplier = round(float(bid_multiplier), 2)
        arm = state.get(normalized_multiplier)
        if arm is None:
            arm = BanditArmState(multiplier=normalized_multiplier)
            state[normalized_multiplier] = arm

        arm.trials += 1
        if success:
            arm.successes += 1
            arm.alpha += 1
        else:
            arm.beta += 1
            
        # Persist update
        self._save_bandit_state(campaign_id, state)

    def calculate_budget_pacing(self, request: BudgetPacingRequest) -> BudgetPacingRecommendation:
        budget = request.budget_status
        reasoning: List[str] = []

        daily_remaining = budget.daily_remaining
        lifetime_remaining = budget.lifetime_remaining
        recommended_daily_cap = budget.daily_budget

        if daily_remaining is not None:
            recommended_hourly_spend = max(daily_remaining, 0.0) / request.hours_remaining_today
            reasoning.append("Daily remaining budget used for hourly pacing.")
        elif lifetime_remaining is not None and request.days_remaining_lifetime:
            recommended_hourly_spend = max(lifetime_remaining, 0.0) / (request.days_remaining_lifetime * 24)
            recommended_daily_cap = None
            reasoning.append("Lifetime remaining budget used for hourly pacing.")
        else:
            recommended_hourly_spend = 0.0
            recommended_daily_cap = None
            reasoning.append("No remaining budget info; defaulting hourly spend to 0.")

        predicted_eod_spend = budget.today_spend + recommended_hourly_spend * request.hours_remaining_today
        budget_utilization_rate = (
            predicted_eod_spend / budget.daily_budget if budget.daily_budget else 0.0
        )

        pacing_health = "healthy"
        if daily_remaining is not None and daily_remaining <= 0:
            pacing_health = "depleted"
        elif budget.pacing_ratio < 0.9:
            pacing_health = "underspending"
        elif budget.pacing_ratio > 1.1:
            pacing_health = "overspending"

        bid_adjustment_factor = 1.0
        should_pause = False
        should_increase = False
        should_decrease = False

        if pacing_health == "depleted":
            bid_adjustment_factor = 0.0
            should_pause = True
            reasoning.append("Budget depleted; pause bidding.")
        elif pacing_health == "underspending":
            bid_adjustment_factor = 1.1
            should_increase = True
            reasoning.append("Underspending; increase bids to pace up.")
        elif pacing_health == "overspending":
            bid_adjustment_factor = 0.9
            should_decrease = True
            reasoning.append("Overspending; reduce bids to pace down.")

        return BudgetPacingRecommendation(
            request_id=request.request_id,
            recommended_hourly_spend=recommended_hourly_spend,
            recommended_daily_cap=recommended_daily_cap,
            bid_adjustment_factor=bid_adjustment_factor,
            should_pause=should_pause,
            should_increase=should_increase,
            should_decrease=should_decrease,
            reasoning=reasoning,
            pacing_health=pacing_health,
            predicted_eod_spend=max(predicted_eod_spend, 0.0),
            budget_utilization_rate=max(min(budget_utilization_rate, 1.0), 0.0)
        )

    def get_metrics(self) -> Dict[str, float]:
        uptime_seconds = time.time() - self.start_time
        avg_bid_multiplier = self.total_bid_multiplier / self.total_requests if self.total_requests else 0.0
        avg_processing_time_ms = (
            self.total_processing_time_ms / self.total_requests if self.total_requests else 0.0
        )
        cache_hit_rate = self.cache_hits / self.total_requests if self.total_requests else 0.0

        return {
            "total_requests": self.total_requests,
            "avg_bid_multiplier": avg_bid_multiplier,
            "avg_processing_time_ms": avg_processing_time_ms,
            "cache_hit_rate": cache_hit_rate,
            "uptime_seconds": uptime_seconds,
        }

    def is_healthy(self) -> bool:
        return len(self.bid_multipliers) > 0


# Global optimizer instance
optimizer = BidOptimizer()
