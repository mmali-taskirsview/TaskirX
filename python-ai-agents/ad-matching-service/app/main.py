"""
Ad Matching Service - FastAPI Application
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
from app.services.matcher import matcher

# Configure logging
logging.basicConfig(
    level=getattr(logging, settings.LOG_LEVEL),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler('ad_matching.log')
    ]
)

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan events"""
    # Startup
    logger.info("🚀 Starting Ad Matching Service")
    logger.info(f"Environment: {settings.ENV}")
    logger.info(f"Max recommendations: {settings.MAX_RECOMMENDATIONS}")
    logger.info(f"Min similarity score: {settings.MIN_SIMILARITY_SCORE}")
    logger.info(f"Strategy weights - Collaborative: {settings.COLLABORATIVE_WEIGHT}, Content: {settings.CONTENT_WEIGHT}")
    
    # Verify matcher is ready
    if matcher.is_healthy():
        logger.info(f"✅ Ad matcher ready with {len(matcher.campaigns)} campaigns")
    else:
        logger.error("❌ Failed to initialize ad matcher")
    
    yield
    
    # Shutdown
    logger.info("Shutting down Ad Matching Service")
    metrics = matcher.get_metrics()
    logger.info(f"Final metrics: {metrics}")


# Create FastAPI app
app = FastAPI(
    title="Ad Matching Service",
    description="AI-powered ad matching and recommendation engine",
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

@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "service": "Ad Matching Service",
        "version": "1.0.0",
        "status": "operational",
        "campaigns": len(matcher.campaigns),
        "docs": "/docs",
        "health": "/api/health"
    }


@app.exception_handler(Exception)
async def global_exception_handler(request, exc):
    """Global exception handler"""
    logger.error(f"Unhandled exception: {exc}", exc_info=True)
    return JSONResponse(
        status_code=500,
        content={"detail": "Internal server error"}
    )


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
