"""
Bid Optimization Service - FastAPI Application
"""
import logging
import sys
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.gzip import GZipMiddleware
from fastapi.responses import JSONResponse
import uvicorn

from app.config import settings
from app.api.endpoints import router
from app.services.optimizer import optimizer

# Prometheus Instrumentator
from prometheus_fastapi_instrumentator import Instrumentator

# Configure logging
logging.basicConfig(
    level=getattr(logging, settings.LOG_LEVEL),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler('bid_optimization.log')
    ]
)

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan events"""
    # Startup
    logger.info("🚀 Starting Bid Optimization Service")
    logger.info(f"Environment: {settings.ENV}")
    logger.info(f"Exploration rate: {settings.EXPLORATION_RATE}")
    logger.info(f"Bid multiplier range: {settings.MIN_BID_MULTIPLIER}x - {settings.MAX_BID_MULTIPLIER}x")
    logger.info(f"Pacing strategy: {settings.PACING_STRATEGY}")
    
    # Verify optimizer is ready
    if optimizer.is_healthy():
        logger.info(f"✅ Bid optimizer ready with {len(optimizer.bid_multipliers)} multiplier options")
    else:
        logger.error("❌ Failed to initialize bid optimizer")
    
    yield
    
    # Shutdown
    logger.info("Shutting down Bid Optimization Service")
    metrics = optimizer.get_metrics()
    logger.info(f"Final metrics: {metrics}")


# Create FastAPI app
app = FastAPI(
    title="Bid Optimization Service",
    description="AI-powered bid optimization with Thompson Sampling and budget pacing",
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, specify allowed origins
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Gzip compression
app.add_middleware(GZipMiddleware, minimum_size=1000)

# Include API router
app.include_router(router, prefix="/api")

# Instrument Prometheus
Instrumentator().instrument(app).expose(app)

if __name__ == "__main__":
    uvicorn.run("app.main:app", host=settings.HOST, port=settings.PORT, reload=settings.DEBUG)
