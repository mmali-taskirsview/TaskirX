"""
Pydantic schemas for Ad Matching Service
"""
from pydantic import BaseModel, Field, validator
from typing import Optional, Dict, Any, List
from datetime import datetime
from enum import Enum


class MatchingStrategy(str, Enum):
    """Matching strategy enumeration"""
    COLLABORATIVE = "collaborative"
    CONTENT_BASED = "content_based"
    HYBRID = "hybrid"
    PERFORMANCE_BASED = "performance_based"


class UserProfile(BaseModel):
    """User profile for matching"""
    user_id: Optional[str] = None
    session_id: Optional[str] = None
    
    # Demographics
    age: Optional[int] = None
    gender: Optional[str] = None
    country: str
    region: Optional[str] = None
    city: Optional[str] = None
    
    # Interests & Categories
    interests: List[str] = Field(default_factory=list)
    categories: List[str] = Field(default_factory=list)
    
    # Behavioral Data
    viewed_ads: List[str] = Field(default_factory=list)
    clicked_ads: List[str] = Field(default_factory=list)
    converted_ads: List[str] = Field(default_factory=list)
    
    # Device & Context
    device_type: str = "mobile"
    os: str = "unknown"
    browser: Optional[str] = None
    
    # Engagement metrics
    avg_session_duration: Optional[float] = None
    total_impressions: int = 0
    total_clicks: int = 0
    total_conversions: int = 0


class AdSlotInfo(BaseModel):
    """Ad slot information"""
    slot_id: str
    dimensions: List[int] = Field(..., description="[width, height]")
    format: str = "banner"  # banner, video, native, interstitial
    position: Optional[str] = None  # above-fold, below-fold, sidebar
    
    @validator('dimensions')
    def validate_dimensions(cls, v):
        if len(v) != 2:
            raise ValueError("Dimensions must be [width, height]")
        return v


class CampaignContext(BaseModel):
    """Campaign context for filtering"""
    publisher_id: str
    page_url: Optional[str] = None
    page_category: Optional[str] = None
    keywords: List[str] = Field(default_factory=list)
    
    # Filters
    min_bid: Optional[float] = None
    max_bid: Optional[float] = None
    required_categories: List[str] = Field(default_factory=list)
    excluded_categories: List[str] = Field(default_factory=list)


class MatchRequest(BaseModel):
    """Ad matching request"""
    request_id: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    
    # User & Context
    user: UserProfile
    ad_slot: AdSlotInfo
    campaign_context: CampaignContext
    
    # Matching preferences
    strategy: MatchingStrategy = MatchingStrategy.HYBRID
    max_results: int = 10
    
    # Additional context
    metadata: Optional[Dict[str, Any]] = None


class AdRecommendation(BaseModel):
    """Single ad recommendation"""
    campaign_id: str
    campaign_name: str
    advertiser_id: str
    
    # Matching scores
    overall_score: float = Field(..., ge=0.0, le=1.0)
    collaborative_score: float = Field(..., ge=0.0, le=1.0)
    content_score: float = Field(..., ge=0.0, le=1.0)
    performance_score: float = Field(..., ge=0.0, le=1.0)
    
    # Campaign details
    bid_price: float
    creative_url: str
    landing_url: str
    categories: List[str] = Field(default_factory=list)
    
    # Predicted metrics
    predicted_ctr: float = Field(..., ge=0.0, le=1.0)
    predicted_cvr: Optional[float] = None
    predicted_revenue: Optional[float] = None
    
    # Explanation
    match_reasons: List[str] = Field(default_factory=list)
    confidence: float = Field(..., ge=0.0, le=1.0)


class MatchResponse(BaseModel):
    """Ad matching response"""
    request_id: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    
    # Recommendations
    recommendations: List[AdRecommendation]
    total_candidates: int
    
    # Processing metadata
    strategy_used: MatchingStrategy
    processing_time_ms: float
    cached: bool = False
    
    # Diversity metrics
    category_diversity: float = Field(..., ge=0.0, le=1.0)
    advertiser_diversity: float = Field(..., ge=0.0, le=1.0)


class PerformancePredictRequest(BaseModel):
    """Performance prediction request"""
    campaign_id: str
    user: UserProfile
    ad_slot: AdSlotInfo
    
    metrics_to_predict: List[str] = Field(
        default_factory=lambda: ["ctr", "cvr", "revenue"]
    )


class PerformancePrediction(BaseModel):
    """Predicted performance metrics"""
    campaign_id: str
    
    # Predictions
    predicted_ctr: float = Field(..., ge=0.0, le=1.0)
    predicted_cvr: float = Field(..., ge=0.0, le=1.0)
    predicted_revenue: float = Field(..., ge=0.0)
    
    # Confidence intervals
    ctr_confidence: float = Field(..., ge=0.0, le=1.0)
    cvr_confidence: float = Field(..., ge=0.0, le=1.0)
    
    # Model info
    model_version: str = "1.0.0"


class HealthResponse(BaseModel):
    """Health check response"""
    status: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    version: str = "1.0.0"
    models_loaded: bool
    redis_connected: bool
    uptime_seconds: float


class MetricsResponse(BaseModel):
    """Metrics response"""
    total_requests: int
    avg_recommendations: float
    avg_processing_time_ms: float
    cache_hit_rate: float
    uptime_seconds: float
