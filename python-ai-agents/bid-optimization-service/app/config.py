"""
Configuration settings for Bid Optimization Service
"""
from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    """Application settings loaded from environment variables"""
    
    # Server
    PORT: int = 6003
    HOST: str = "0.0.0.0"
    ENV: str = "development"
    DEBUG: bool = True
    
    # Redis
    REDIS_HOST: str = "localhost"
    REDIS_PORT: int = 6379
    REDIS_DB: int = 4
    REDIS_PASSWORD: Optional[str] = None
    
    # Optimization Configuration
    EXPLORATION_RATE: float = 0.1  # Epsilon for epsilon-greedy
    LEARNING_RATE: float = 0.01  # Alpha for Q-learning
    DISCOUNT_FACTOR: float = 0.95  # Gamma for future rewards
    MIN_BID_MULTIPLIER: float = 0.5  # Minimum bid adjustment
    MAX_BID_MULTIPLIER: float = 2.0  # Maximum bid adjustment
    
    # Budget Pacing
    PACING_STRATEGY: str = "even"  # even, aggressive, conservative
    SAFETY_MARGIN: float = 0.1  # Reserve 10% of budget
    
    # Performance
    MAX_WORKERS: int = 4
    REQUEST_TIMEOUT: int = 50  # milliseconds
    CACHE_TTL: int = 300  # seconds
    
    # Monitoring
    ENABLE_METRICS: bool = True
    LOG_LEVEL: str = "INFO"
    
    # External Services
    NESTJS_API_URL: str = "http://localhost:4000"
    CLICKHOUSE_HOST: str = "localhost"
    CLICKHOUSE_PORT: int = 8123
    CLICKHOUSE_DB: str = "taskirx_analytics"
    
    class Config:
        env_file = ".env"
        case_sensitive = True


# Global settings instance
settings = Settings()
