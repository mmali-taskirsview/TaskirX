import os
# Set mock env var before importing app
os.environ["IP_REPUTATION_API_KEY"] = "test_key_mock"

from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)

def test_health_check_endpoint():
    response = client.get("/api/health")
    assert response.status_code == 200
    assert response.json()["status"] == "healthy"
