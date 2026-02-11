"""
Pydantic schemas for Bid Optimization Service
"""
from pydantic import BaseModel, Field, validator
from typing import Optional, Dict, Any, List
from datetime import datetime, date
from enum import Enum


class AuctionType(str, Enum):
    """Auction type enumeration"""
    FIRST_PRICE = "first_price"
    SECOND_PRICE = "second_price"


class BidStrategy(str, Enum):
    """Bid strategy enumeration"""
    MAXIMIZE_CLICKS = "maximize_clicks"
    MAXIMIZE_CONVERSIONS = "maximize_conversions"
    TARGET_CPA = "target_cpa"
    TARGET_ROAS = "target_roas"
    MANUAL = "manual"


class PacingStrategy(str, Enum):
    """Budget pacing strategy"""
    EVEN = "even"  # Spend evenly throughout period
    AGGRESSIVE = "aggressive"  # Spend as fast as possible
    CONSERVATIVE = "conservative"  # Spend slowly, prioritize efficiency
    ASAP = "asap"  # Spend entire budget ASAP


class CampaignPerformance(BaseModel):
    """Historical campaign performance metrics"""
    campaign_id: str
    
    # Current stats
    impressions: int = 0
    clicks: int = 0
    conversions: int = 0
    spend: float = 0.0
    revenue: float = 0.0
    
    # Calculated metrics
    ctr: float = Field(default=0.0, ge=0.0, le=1.0)
    cvr: float = Field(default=0.0, ge=0.0, le=1.0)
    cpc: float = Field(default=0.0, ge=0.0)
    cpa: float = Field(default=0.0, ge=0.0)
    roas: float = Field(default=0.0, ge=0.0)
    
    # Win rate
    bid_requests: int = 0
    wins: int = 0
    win_rate: float = Field(default=0.0, ge=0.0, le=1.0)


class BudgetStatus(BaseModel):
    """Campaign budget status"""
    campaign_id: str
    
    # Budget limits
    daily_budget: Optional[float] = None
    lifetime_budget: Optional[float] = None
    
    # Current spend
    today_spend: float = 0.0
    total_spend: float = 0.0
    
    # Remaining
    daily_remaining: Optional[float] = None
    lifetime_remaining: Optional[float] = None
    
    # Pacing
    expected_daily_spend: Optional[float] = None
    pacing_ratio: float = 1.0  # actual_spend / expected_spend
    is_underspending: bool = False
    is_overspending: bool = False


class OptimizationContext(BaseModel):
    """Context for bid optimization"""
    campaign_id: str
    base_bid: float = Field(..., gt=0)
    
    # Performance
    performance: CampaignPerformance
    
    # Budget
    budget_status: BudgetStatus
    
    # Auction context
    auction_type: AuctionType = AuctionType.FIRST_PRICE
    estimated_competition: float = Field(default=0.5, ge=0.0, le=1.0)
    
    # Time context
    hour_of_day: int = Field(..., ge=0, le=23)
    day_of_week: int = Field(..., ge=0, le=6)
    
    # Additional context
    metadata: Optional[Dict[str, Any]] = None


class BidOptimizationRequest(BaseModel):
    """Bid optimization request"""
    request_id: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    
    # Campaign & context
    context: OptimizationContext
    
    # Strategy
    strategy: BidStrategy = BidStrategy.MAXIMIZE_CONVERSIONS
    target_cpa: Optional[float] = None
    target_roas: Optional[float] = None
    
    # Constraints
    min_bid: Optional[float] = None
    max_bid: Optional[float] = None


class BidRecommendation(BaseModel):
    """Optimized bid recommendation"""
    request_id: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    
    # Recommended bid
    recommended_bid: float = Field(..., gt=0)
    bid_multiplier: float = Field(..., gt=0)
    
    # Confidence & reasoning
    confidence: float = Field(..., ge=0.0, le=1.0)
    reasoning: List[str] = Field(default_factory=list)
    
    # Expected outcomes
    expected_win_rate: float = Field(..., ge=0.0, le=1.0)
    expected_ctr: Optional[float] = None
    expected_cvr: Optional[float] = None
    expected_roi: Optional[float] = None
    
    # Processing metadata
    strategy_used: BidStrategy
    processing_time_ms: float


class BudgetPacingRequest(BaseModel):
    """Budget pacing request"""
    request_id: str
    campaign_id: str
    
    # Budget info
    budget_status: BudgetStatus
    
    # Time remaining
    hours_remaining_today: float = Field(..., gt=0, le=24)
    days_remaining_lifetime: Optional[float] = None
    
    # Strategy
    pacing_strategy: PacingStrategy = PacingStrategy.EVEN


class BudgetPacingRecommendation(BaseModel):
    """Budget pacing recommendation"""
    request_id: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    
    # Recommendations
    recommended_hourly_spend: float = Field(..., ge=0)
    recommended_daily_cap: Optional[float] = None
    bid_adjustment_factor: float = Field(default=1.0, gt=0)
    
    # Status
    should_pause: bool = False
    should_increase: bool = False
    should_decrease: bool = False
    
    # Reasoning
    reasoning: List[str] = Field(default_factory=list)
    pacing_health: str = "healthy"  # healthy, underspending, overspending, depleted
    
    # Predictions
    predicted_eod_spend: float = Field(..., ge=0)
    budget_utilization_rate: float = Field(..., ge=0.0, le=1.0)


class ThompsonSamplingState(BaseModel):
    """Thompson Sampling bandit state"""
    campaign_id: str
    action: str  # bid_multiplier_level (e.g., "0.8", "1.0", "1.2")
    
    # Beta distribution parameters
    alpha: float = 1.0  # successes + 1
    beta: float = 1.0  # failures + 1
    
    # Stats
    trials: int = 0
    successes: int = 0
    estimated_success_rate: float = 0.5


class HealthResponse(BaseModel):
    """Health check response"""
    status: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    version: str = "1.0.0"
    optimizer_ready: bool
    redis_connected: bool
    uptime_seconds: float


class MetricsResponse(BaseModel):
    """Metrics response"""
    total_requests: int
    avg_bid_multiplier: float
    avg_processing_time_ms: float
    cache_hit_rate: float
    uptime_seconds: float
