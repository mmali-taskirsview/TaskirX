"""
Fraud Detection Service - FastAPI Application
"""
import logging
import sys
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.gzip import GZipMiddleware
from fastapi.responses import JSONResponse
import uvicorn
from prometheus_fastapi_instrumentator import Instrumentator

from app.config import settings
from app.api.endpoints import router
from app.services.fraud_detector import detector

# Configure logging
logging.basicConfig(
    level=getattr(logging, settings.LOG_LEVEL),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler('fraud_detection.log')
    ]
)

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan events"""
    # Startup
    logger.info("🚀 Starting Fraud Detection Service")
    logger.info(f"Environment: {settings.ENV}")
    logger.info(f"Model threshold: {settings.MODEL_THRESHOLD}")
    logger.info(f"Request timeout: {settings.REQUEST_TIMEOUT}ms")
    
    # Verify model is loaded
    if detector.is_healthy():
        logger.info("✅ Fraud detection model loaded successfully")
    else:
        logger.error("❌ Failed to load fraud detection model")
    
    yield
    
    # Shutdown
    logger.info("Shutting down Fraud Detection Service")
    metrics = detector.get_metrics()
    logger.info(f"Final metrics: {metrics}")


# Create FastAPI app
app = FastAPI(
    title="Fraud Detection Service",
    description="AI-powered fraud detection for ad interactions",
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
    logger.info(f"Starting server on {settings.HOST}:{settings.PORT}")
    
    uvicorn.run(
        "app.main:app",
        host=settings.HOST,
        port=settings.PORT,
        reload=settings.DEBUG,
        log_level=settings.LOG_LEVEL.lower(),
        access_log=True
    )
