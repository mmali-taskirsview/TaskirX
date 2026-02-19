from locust import HttpUser, task, between
import uuid
import random

class MixedFormatUser(HttpUser):
    wait_time = between(0.1, 0.5)
    
    # Weights simulate traffic distribution
    # banner: 50%, native: 20%, video: 15%, rich_media: 10%, audio: 5%
    
    def on_start(self):
        self.publisher_id = "pub-001"
        # Use a safe IP to avoid fraud blocks during load testing
        self.safe_ip = "203.0.113.1" 

    def send_bid(self, formats):
        req_id = str(uuid.uuid4())
        # Add host header for production routing
        headers = {
            "Content-Type": "application/json",
            "Host": "bidding.taskirx.com"
        }
        payload = {
            "id": req_id,
            "timestamp": "2026-02-17T12:00:00Z",
            "publisher_id": self.publisher_id,
            "ad_slot": {
                "id": "slot-" + str(random.randint(1, 100)),
                "dimensions": [300, 250],
                "position": "above-fold",
                "formats": formats
            },
            "user": {
                "id": "user-" + str(random.randint(1, 1000)),
                "country": "US",
                "categories": ["tech", "gaming"]
            },
            "device": {
                "type": "mobile",
                "ip": self.safe_ip,
                "ua": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
            }
        }
        
        self.client.post("/bid", json=payload, headers=headers)

    @task(10) # 50%
    def banner_bid(self):
        self.send_bid(["banner"])

    @task(20)
    def request_native(self):
        self.send_bid(["native"])

    @task(15)
    def request_video(self):
        self.send_bid(["video"])
        
    @task(10)
    def request_rich_media(self):
        self.send_bid(["rich_media"])

    @task(5)
    def request_audio(self):
        self.send_bid(["audio"])
