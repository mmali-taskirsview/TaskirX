"""
Configuration settings for Ad Matching Service
"""
from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    """Application settings loaded from environment variables"""
    
    # Server
    PORT: int = 6002
    HOST: str = "0.0.0.0"
    ENV: str = "development"
    DEBUG: bool = True
    
    # Redis
    REDIS_HOST: str = "localhost"
    REDIS_PORT: int = 6379
    REDIS_DB: int = 3
    REDIS_PASSWORD: Optional[str] = None
    
    # Matching Configuration
    MAX_RECOMMENDATIONS: int = 10
    MIN_SIMILARITY_SCORE: float = 0.3
    CONTENT_WEIGHT: float = 0.4
    COLLABORATIVE_WEIGHT: float = 0.6
    
    # Performance
    MAX_WORKERS: int = 4
    REQUEST_TIMEOUT: int = 50  # milliseconds
    CACHE_TTL: int = 600  # seconds (10 minutes)
    
    # Monitoring
    ENABLE_METRICS: bool = True
    LOG_LEVEL: str = "INFO"
    
    # External Services
    NESTJS_API_URL: str = "http://localhost:4000"
    CLICKHOUSE_HOST: str = "localhost"
    CLICKHOUSE_PORT: int = 8123
    CLICKHOUSE_DB: str = "taskirx_analytics"
    
    # Pinecone
    USE_PINECONE: bool = True
    PINECONE_API_KEY: Optional[str] = None
    PINECONE_INDEX_NAME: str = "taskir-ads"
    PINECONE_ENVIRONMENT: str = "us-east-1"
    
    class Config:
        env_file = ".env"
        case_sensitive = True


# Global settings instance
settings = Settings()
