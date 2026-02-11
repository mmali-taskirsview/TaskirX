"""
Pydantic schemas for request/response validation
"""
from pydantic import BaseModel, Field, validator
from typing import Optional, Dict, Any, List
from datetime import datetime
from enum import Enum


class DeviceType(str, Enum):
    """Device type enumeration"""
    MOBILE = "mobile"
    DESKTOP = "desktop"
    TABLET = "tablet"
    TV = "tv"
    UNKNOWN = "unknown"


class FraudRiskLevel(str, Enum):
    """Risk level classification"""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class DeviceInfo(BaseModel):
    """Device information"""
    type: DeviceType
    os: str
    os_version: Optional[str] = None
    browser: Optional[str] = None
    browser_version: Optional[str] = None
    user_agent: Optional[str] = None
    screen_resolution: Optional[str] = None
    language: Optional[str] = None


class GeoInfo(BaseModel):
    """Geographic information"""
    country: str
    region: Optional[str] = None
    city: Optional[str] = None
    lat: Optional[float] = None
    lon: Optional[float] = None
    timezone: Optional[str] = None


class UserBehavior(BaseModel):
    """User behavior metrics"""
    clicks_last_hour: int = 0
    clicks_last_24h: int = 0
    impressions_last_hour: int = 0
    impressions_last_24h: int = 0
    conversions_last_24h: int = 0
    avg_time_on_site: Optional[float] = None
    bounce_rate: Optional[float] = None


class FraudCheckRequest(BaseModel):
    """Fraud detection request payload"""
    request_id: str = Field(..., description="Unique request identifier")
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    
    # Core identifiers
    user_id: Optional[str] = None
    session_id: Optional[str] = None
    ip_address: str
    
    # Campaign context
    campaign_id: str
    publisher_id: str
    advertiser_id: str
    
    # Device & Geo
    device: DeviceInfo
    geo: GeoInfo
    
    # Behavior (optional, improves accuracy)
    behavior: Optional[UserBehavior] = None
    
    # Additional context
    referrer: Optional[str] = None
    landing_page: Optional[str] = None
    click_timestamp: Optional[datetime] = None
    metadata: Optional[Dict[str, Any]] = None
    
    @validator('ip_address')
    def validate_ip(cls, v):
        """Basic IP validation"""
        if not v or v == "0.0.0.0":
            raise ValueError("Invalid IP address")
        return v


class FraudIndicators(BaseModel):
    """Detailed fraud indicators"""
    suspicious_ip: bool = False
    suspicious_device: bool = False
    suspicious_user_agent: bool = False
    high_click_frequency: bool = False
    impossible_travel: bool = False
    bot_detected: bool = False
    proxy_detected: bool = False
    datacenter_ip: bool = False
    device_fingerprint_mismatch: bool = False
    behavioral_anomaly: bool = False


class FraudCheckResponse(BaseModel):
    """Fraud detection response"""
    request_id: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    
    # Fraud assessment
    is_fraud: bool
    fraud_score: float = Field(..., ge=0.0, le=1.0, description="Fraud probability (0-1)")
    risk_level: FraudRiskLevel
    confidence: float = Field(..., ge=0.0, le=1.0, description="Model confidence")
    
    # Detailed indicators
    indicators: FraudIndicators
    
    # Reasons (human-readable)
    reasons: List[str] = Field(default_factory=list)
    
    # Action recommendation
    recommended_action: str  # "allow", "flag", "block"
    
    # Processing metadata
    processing_time_ms: float
    model_version: str = "1.0.0"


class HealthResponse(BaseModel):
    """Health check response"""
    status: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    version: str = "1.0.0"
    model_loaded: bool
    redis_connected: bool
    uptime_seconds: float


class MetricsResponse(BaseModel):
    """Metrics response"""
    total_requests: int
    fraud_detected: int
    fraud_rate: float
    avg_processing_time_ms: float
    model_accuracy: Optional[float] = None
    uptime_seconds: float
