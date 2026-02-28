"""
Simplified Performance Test for Go Bidding Engine
Tests the /bid endpoint with realistic payloads
"""
from locust import HttpUser, task, between
import json
import uuid
import random

class SimpleBiddingUser(HttpUser):
    wait_time = between(0.01, 0.1)  # High frequency
    
    @task(10)
    def test_basic_bid(self):
        """Basic bid request - most common scenario"""
        payload = {
            "id": str(uuid.uuid4()),
            "timestamp": "2026-02-28T12:00:00Z",
            "publisher_id": "pub-test-001",
            "ad_slot": {
                "id": "slot-001",
                "dimensions": [300, 250],
                "position": "above-fold",
                "formats": ["banner"]
            },
            "user": {
                "id": f"user-{random.randint(1, 1000)}",
                "country": "US",
                "language": "en",
                "age": random.randint(18, 65),
                "gender": random.choice(["male", "female"])
            },
            "device": {
                "type": random.choice(["mobile", "desktop", "tablet"]),
                "os": random.choice(["iOS", "Android", "Windows"]),
                "browser": "Chrome",
                "ip": "203.0.113.1"
            },
            "context": {
                "site_domain": "news.com",
                "site_category": "IAB1"
            }
        }
        
        with self.client.post(
            "/bid",
            json=payload,
            catch_response=True,
            name="POST /bid (basic)"
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Status {response.status_code}")
    
    @task(5)
    def test_video_bid(self):
        """Video ad bid request"""
        payload = {
            "id": str(uuid.uuid4()),
            "timestamp": "2026-02-28T12:00:00Z",
            "publisher_id": "pub-test-001",
            "ad_slot": {
                "id": "slot-video-001",
                "dimensions": [1280, 720],
                "position": "in-stream",
                "formats": ["video"],
                "video": {
                    "mimes": ["video/mp4"],
                    "minduration": 5,
                    "maxduration": 30,
                    "protocols": [2, 3]
                }
            },
            "user": {
                "id": f"user-{random.randint(1, 1000)}",
                "country": "US",
                "language": "en",
                "age": random.randint(18, 65),
                "interests": ["sports", "technology"]
            },
            "device": {
                "type": "mobile",
                "os": "iOS",
                "browser": "Safari",
                "ip": "203.0.113.1"
            },
            "context": {
                "site_domain": "sports.com",
                "site_category": "IAB17"
            }
        }
        
        with self.client.post(
            "/bid",
            json=payload,
            catch_response=True,
            name="POST /bid (video)"
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Status {response.status_code}")
    
    @task(3)
    def test_native_bid(self):
        """Native ad bid request"""
        payload = {
            "id": str(uuid.uuid4()),
            "timestamp": "2026-02-28T12:00:00Z",
            "publisher_id": "pub-test-001",
            "ad_slot": {
                "id": "slot-native-001",
                "dimensions": [1, 1],
                "position": "feed",
                "formats": ["native"],
                "native": {
                    "assets": [
                        {"id": 1, "title": {"len": 140}},
                        {"id": 2, "img": {"w": 300, "h": 250}},
                        {"id": 3, "data": {"type": 1}}
                    ]
                }
            },
            "user": {
                "id": f"user-{random.randint(1, 1000)}",
                "country": "US",
                "language": "en"
            },
            "device": {
                "type": "mobile",
                "os": "Android",
                "ip": "203.0.113.1"
            }
        }
        
        with self.client.post(
            "/bid",
            json=payload,
            catch_response=True,
            name="POST /bid (native)"
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Status {response.status_code}")
    
    @task(1)
    def test_health(self):
        """Health check endpoint"""
        with self.client.get(
            "/health",
            catch_response=True,
            name="GET /health"
        ) as response:
            if response.status_code == 200:
                try:
                    data = response.json()
                    if data.get("status") == "ok":
                        response.success()
                    else:
                        response.failure("Health check failed")
                except:
                    response.failure("Invalid JSON response")
            else:
                response.failure(f"Status {response.status_code}")
