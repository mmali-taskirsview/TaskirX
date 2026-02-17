"""
Configuration settings for Fraud Detection Service
"""
from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    """Application settings loaded from environment variables"""
    
    # Server
    PORT: int = 6001
    HOST: str = "0.0.0.0"
    ENV: str = "development"
    DEBUG: bool = True
    
    # Redis
    REDIS_HOST: str = "localhost"
    REDIS_PORT: int = 6379
    REDIS_DB: int = 2
    REDIS_PASSWORD: Optional[str] = None
    
    # Model
    MODEL_PATH: str = "./models/fraud_detector.pkl"
    MODEL_THRESHOLD: float = 0.7
    FEATURE_SCALER_PATH: str = "./models/scaler.pkl"
    
    # Performance
    MAX_WORKERS: int = 4
    REQUEST_TIMEOUT: int = 100  # milliseconds
    CACHE_TTL: int = 300  # seconds
    
    # Monitoring
    ENABLE_METRICS: bool = True
    LOG_LEVEL: str = "INFO"
    
    # External Services
    NESTJS_API_URL: str = "http://localhost:4000"
    CLICKHOUSE_HOST: str = "localhost"
    CLICKHOUSE_PORT: int = 8123
    CLICKHOUSE_DB: str = "taskirx_analytics"
    
    # IP Reputation
    IP_REPUTATION_API_KEY: Optional[str] = None
    
    class Config:
        env_file = ".env"
        case_sensitive = True


# Global settings instance
settings = Settings()
