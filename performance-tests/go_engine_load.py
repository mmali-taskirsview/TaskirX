from locust import HttpUser, task, between, events
import json
import uuid
import random

class GoBiddingUser(HttpUser):
    # Simulate high-frequency ad exchanges
    wait_time = between(0.01, 0.1) 
    
    @task
    def bid_request(self):
        req_id = str(uuid.uuid4())
        
        # Use fixed IP known to work
        fixed_ip = "203.0.113.1"
        
        payload = {
            "id": req_id,
            "timestamp": "2026-02-17T12:00:00Z",
            "publisher_id": "pub-001",
            "ad_slot": {
                "id": "slot-001",
                "dimensions": [300, 250],
                "position": "above-fold",
                "formats": ["banner"]
            },
            "user": {
                "id": "user-001",
                "country": "US",  # Match Campaign targeting
                "language": "en",
                # "categories": ["technology"], # Removed to match integration test behavior
                "age": 25,
                "gender": "male"
            },
            "device": {
                "type": "mobile", # Match Campaign targeting
                "os": "Android",
                "browser": "Chrome",
                "ip": fixed_ip,
                "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
                "geo": {
                    "lat": 40.7128,
                    "lon": -74.0060,
                    "country": "US",
                    "city": "New York"
                }
            },
            "context": {
                "site_domain": "news.com",
                "site_category": "IAB1"
            }
        }
        
        headers = {"Content-Type": "application/json"}
        
        with self.client.post("/bid", json=payload, headers=headers, catch_response=True) as response:
            if response.status_code == 200:
                resp_json = response.json()
                # Check if we actually got a bid (price > 0)
                if "bid_price" in resp_json and resp_json["bid_price"] > 0:
                     response.success()
                elif "reason" in resp_json:
                    # Valid response but no bid (logic correct, just business rule)
                    # We might want to count this as success for load testing purposes
                    # but fail it if we EXPECT bids.
                    # Given we have seeded data, we expect bids.
                    if resp_json["reason"] == "no matching campaigns":
                        response.failure("No matching campaigns (Check targeting/seed data)")
                    else:
                        response.success() # Fraud or other reason is valid system behavior
                else:
                    response.failure(f"Invalid response format: {response.text}")
            elif response.status_code == 204:
                response.success()
            else:
                response.failure(f"Status code: {response.status_code}")

