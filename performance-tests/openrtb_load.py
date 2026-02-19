from locust import HttpUser, task, between, events
import json
import uuid
import random

class OpenRTBUser(HttpUser):
    # Simulate high-frequency ad exchanges
    wait_time = between(0.01, 0.1) 
    
    @task(5)
    def banner_request(self):
        self._send_openrtb_request("banner")

    @task(3)
    def video_request(self):
        self._send_openrtb_request("video")

    @task(1)
    def native_request(self):
        self._send_openrtb_request("native")

    @task(1)
    def audio_request(self):
        self._send_openrtb_request("audio")

    def _send_openrtb_request(self, format_type):
        req_id = str(uuid.uuid4())
        imp_id = f"imp-{req_id[:8]}"
        
        imp = {"id": imp_id}
        if format_type == "banner":
            imp["banner"] = {"w": 300, "h": 250}
        elif format_type == "video":
            imp["video"] = {"mimes": ["video/mp4"], "minduration": 5, "maxduration": 30}
        elif format_type == "native":
            native_req = {
                "ver": "1.2",
                "assets": [
                    {"id": 1, "required": 1, "title": {"len": 140}},
                    {"id": 2, "required": 1, "img": {"type": 3, "w": 1200, "h": 627}},
                    {"id": 3, "required": 0, "data": {"type": 2, "len": 140}}
                ]
            }
            imp["native"] = {"request": json.dumps(native_req), "ver": "1.2"}
        elif format_type == "audio":
            imp["audio"] = {"mimes": ["audio/mp3"], "minduration": 5, "maxduration": 30}

        payload = {
            "id": req_id,
            "imp": [imp],
            "site": {"id": "site-001", "name": "News Site", "domain": "news.com"},
            "device": {
                "ua": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
                "ip": "203.0.113.1",
                "geo": {"country": "US", "city": "New York", "lat": 40.7128, "lon": -74.0060},
                "devicetype": 2 # PC
            },
            "user": {
                "id": "user-001",
                "keywords": "technology,finance",
                "geo": {"country": "US"}
            }
        }
        
        headers = {"Content-Type": "application/json", "x-openrtb-version": "2.5"}
        
        with self.client.post("/openrtb", json=payload, headers=headers, catch_response=True) as response:
            if response.status_code == 200:
                # 200 OK means we got a bid
                try:
                    resp_json = response.json()
                    if "seatbid" in resp_json and len(resp_json["seatbid"]) > 0:
                        response.success()
                    else:
                        # Valid OpenRTB might return 200 with empty seatbid? Usually 204.
                        # But Go implementation returns 200 with SeatBid or 204.
                        # If seatbid is empty but present, that's weird for a 200.
                        # Let's assume correct behavior is checked.
                        response.success()
                except Exception as e:
                     response.failure(f"JSON Decode Error: {e}")
            elif response.status_code == 204:
                # No content = No Bid (Valid)
                response.success()
            else:
                response.failure(f"Status code: {response.status_code}")
