"""
TaskirX v3.0 - Simple Database Test API
FastAPI service to verify database connections
"""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import psycopg2
import redis
import logging
from config import settings

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Create FastAPI app
app = FastAPI(title="TaskirX Database Test API", version="1.0.0")

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/")
def root():
    return {
        "service": "TaskirX Database Test API",
        "version": "1.0.0",
        "status": "running"
    }

@app.get("/health")
def health_check():
    return {"status": "healthy"}

@app.get("/api/test/postgres")
def test_postgres():
    """Test PostgreSQL connection"""
    try:
        conn = psycopg2.connect(
            host=settings.database_host,
            port=settings.database_port,
            database=settings.database_name,
            user=settings.database_user,
            password=settings.database_password
        )
        cursor = conn.cursor()
        cursor.execute("SELECT COUNT(*) FROM users")
        count = cursor.fetchone()[0]
        cursor.close()
        conn.close()
        
        return {
            "status": "success",
            "database": "PostgreSQL",
            "connected": True,
            "users_count": count
        }
    except Exception as e:
        logger.error(f"PostgreSQL error: {e}")
        return {
            "status": "error",
            "database": "PostgreSQL",
            "connected": False,
            "error": str(e)
        }

@app.get("/api/test/redis")
def test_redis():
    """Test Redis connection"""
    try:
        r = redis.Redis(
            host=settings.redis_host,
            port=settings.redis_port,
            password=settings.redis_password,
            decode_responses=True
        )
        r.ping()
        
        # Test set and get
        r.set("test_key", "test_value", ex=10)
        value = r.get("test_key")
        
        return {
            "status": "success",
            "database": "Redis",
            "connected": True,
            "ping": "PONG",
            "test_value": value
        }
    except Exception as e:
        logger.error(f"Redis error: {e}")
        return {
            "status": "error",
            "database": "Redis",
            "connected": False,
            "error": str(e)
        }

@app.get("/api/users")
def get_users():
    """Get all users from PostgreSQL"""
    try:
        conn = psycopg2.connect(
            host=settings.database_host,
            port=settings.database_port,
            database=settings.database_name,
            user=settings.database_user,
            password=settings.database_password
        )
        cursor = conn.cursor()
        cursor.execute("""
            SELECT id, email, role, "companyName", "isActive", "createdAt"
            FROM users
            ORDER BY "createdAt" DESC
        """)
        
        users = []
        for row in cursor.fetchall():
            users.append({
                "id": str(row[0]),
                "email": row[1],
                "role": row[2],
                "companyName": row[3],
                "isActive": row[4],
                "createdAt": str(row[5])
            })
        
        cursor.close()
        conn.close()
        
        return {
            "status": "success",
            "count": len(users),
            "users": users
        }
    except Exception as e:
        logger.error(f"Error fetching users: {e}")
        return {
            "status": "error",
            "error": str(e)
        }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=settings.port)
