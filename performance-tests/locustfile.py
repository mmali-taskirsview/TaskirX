from locust import HttpUser, task, between
import json
import os
import random
import logging

class BiddingUser(HttpUser):
    wait_time = between(0.1, 0.5)  # Simulate high-frequency trading/bidding
    host = os.getenv("LOCUST_HOST", "http://localhost:3000")
    
    # Auth credentials (default to test user)
    email = os.getenv("LOCUST_EMAIL", "test@example.com")
    password = os.getenv("LOCUST_PASSWORD", "Test123!")

    # Seed data for requests (Matches scripts/seed-perf-data.sql)
    # Using real UUIDs from the seed file to ensure database hits
    ad_unit_ids = [
        "c0eebc99-9c0b-4ef8-bb6d-6bb9bd380123", # au-123
        "d0eebc99-9c0b-4ef8-bb6d-6bb9bd380456", # au-456
        "e0eebc99-9c0b-4ef8-bb6d-6bb9bd380789"  # au-789
    ]
    publishers = [
        "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", # pub-001
        "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22"  # pub-002
    ]

    @task(3)
    def ssp_auction(self):
        """
        Simulate an SSP Auction request.
        Target: /ssp/auction (TaskirX V3 Custom Payload)
        """
        payload = {
            "id": f"req-{random.randint(10000, 99999)}",
            "publisherId": random.choice(self.publishers),
            "adUnitId": random.choice(self.ad_unit_ids),
            "device": {
                "type": "mobile",
                "os": "ios", 
                "browser": "safari"
            },
            "geo": {
                "country": "US",
                "region": "CA",
                "city": "San Francisco"
            },
            "user": {
                "consent": True
            },
            "floor": random.uniform(0.5, 2.0)
        }
        
        # Headers typically required for RTB/SSP
        with self.client.post(
            "/api/ssp/auction",
            json=payload,
            headers=self.ssp_headers,
            catch_response=True,
        ) as response:
            if response.status_code == 200:
                response.success()
            elif response.status_code == 204:
                response.success()  # no bid is allowed
            elif response.status_code == 404:
                response.failure("Endpoint not found")
            elif response.status_code == 0:
                response.failure("Connection error - is the host up?")
            else:
                response.failure(f"Got status {response.status_code}")

    @task(1)
    def dsp_process_bid(self):
        """
        Simulate a DSP Bid Processing request.
        Target: /dsp/bid
        """
        payload = {
            "requestId": f"bid-req-{random.randint(10000, 99999)}",
            "supplyPartnerId": "11eebc99-9c0b-4ef8-bb6d-6bb9bd380111", # Mock UUID from seed-perf-data.sql
            "impressionId": f"imp-{random.randint(10000, 99999)}",
            "adUnitId": random.choice(self.ad_unit_ids),
            "floor": random.uniform(0.5, 2.0),
            "device": { "type": "desktop", "os": "windows", "browser": "chrome" },
            "geo": { "country": "US", "region": "CA", "city": "SF" },
            "user": { "id": "user_1" }
        }
        
        with self.client.post(
            "/api/dsp/bid",
            json=payload,
            headers=self.dsp_headers,
            catch_response=True,
        ) as response:
            if response.status_code in [200, 201, 204]:
                response.success()
            elif response.status_code == 0:
                response.failure("Connection error - is the host up?")
            else:
                response.failure(f"DSP Bid failed: {response.status_code}")

    @task(1)
    def analytics_dashboard(self):
        """
        Simulate Analytics Dashboard Load.
        Target: /api/analytics/dashboard
        """
        # If this requires auth, set LOCUST_TOKEN and we'll send Authorization automatically.
        with self.client.get(
            "/api/analytics/dashboard",
            headers=self.analytics_headers or None,
            catch_response=True,
        ) as response:
            if response.status_code == 200:
                response.success()
            elif response.status_code == 401:
                response.failure("Unauthorized (set LOCUST_TOKEN for auth)")
            elif response.status_code == 0:
                response.failure("Connection error - is the host up?")
            else:
                response.failure(f"Analytics Dashboard failed: {response.status_code}")

    @task(2)
    def analytics_tracking(self):
        """
        Simulate Tracking Events (Impression, Click, Conversion).
        Targets: /analytics/track/impression, /analytics/track/click, /analytics/track/conversion
        """
        # Event type distribution: many impressions, fewer clicks, very few conversions
        r = random.random()
        
        event_data = {
            "campaignId": "camp_1",
            "adUnitId": random.choice(self.ad_unit_ids),
            "timestamp": "2026-02-14T12:00:00Z"
        }

        if r < 0.90:
            # Impression
            imp_payload = {
                "campaignId": "camp_1",
                "publisherId": "pub-001",
                "deviceType": "data.device.type",
                "country": "data.geo.country",
                "timestamp": "2026-02-14T12:00:00Z"
            }
            with self.client.post("/api/analytics/track/impression", json=imp_payload, headers=self.analytics_headers, catch_response=True) as resp:
                if resp.status_code in [200, 201]: resp.success()
                else: resp.failure(f"Track Impression failed: {resp.status_code}")
        elif r < 0.98:
            # Click
            click_payload = {
                "impressionId": f"imp-{random.randint(10000, 99999)}",
                "campaignId": "camp_1",
                "timestamp": "2026-02-14T12:00:00Z"
            }
            with self.client.post("/api/analytics/track/click", json=click_payload, headers=self.analytics_headers, catch_response=True) as resp:
                if resp.status_code in [200, 201]: resp.success()
                else: resp.failure(f"Track Click failed: {resp.status_code}")
        else:
            # Conversion
            conv_payload = {
                "clickId": f"clk-{random.randint(10000, 99999)}",
                "campaignId": "camp_1",
                "conversionValue": 10.0,
                "timestamp": "2026-02-14T12:00:00Z"
            }
            with self.client.post("/api/analytics/track/conversion", json=conv_payload, headers=self.analytics_headers, catch_response=True) as resp:
                if resp.status_code in [200, 201]: resp.success()
                else: resp.failure(f"Track Conversion failed: {resp.status_code}")

    def on_start(self):
        """
        Called when a Locust user starts before any task is scheduled.
        Sets common headers, including optional auth for analytics.
        """
        token = os.getenv("LOCUST_TOKEN")

        # Initialize shared headers
        self.ssp_headers = {
            "Content-Type": "application/json",
            "X-OpenRTB-Version": "2.5",
        }

        self.dsp_headers = {
            "Content-Type": "application/json",
        }
        
        self.analytics_headers = {}

        # 1. Try env var token
        if token:
            logging.info("Using token from environment variable.")
            self.analytics_headers["Authorization"] = f"Bearer {token}"
            return
        
        # 2. Try Register (bypass manual hash issues)
        logging.info(f"Attempting registration for user {self.email}...")
        try:
            with self.client.post(
                "/api/auth/register",
                json={"email": self.email, "password": self.password, "role": "advertiser", "companyName": "Locust Corp"},
                headers={"Content-Type": "application/json"},
                name="/api/auth/register (setup)",
                catch_response=True
            ) as reg_resp:
                if reg_resp.status_code in [200, 201]:
                     data = reg_resp.json()
                     token = data.get("access_token") or data.get("accessToken")
                     if token:
                         logging.info("Registration successful. Got token.")
                         self.analytics_headers["Authorization"] = f"Bearer {token}"
                         reg_resp.success()
                         return
                elif reg_resp.status_code == 409:
                    # User already exists, not a failure for our test setup
                    reg_resp.success()
                    logging.info("User already exists, proceeding to login.")
                else:
                    reg_resp.failure(f"Registration failed with {reg_resp.status_code}")
        except Exception as e:
            logging.warning(f"Registration/Login attempt 1 failed: {e}")

        # 3. Try Login (if registration failed because user exists)
        try:
            logging.info(f"Attempting login for user {self.email}...")
            # Use immediate client to avoid messing with stats if possible, 
            # or just accept it as part of startup overhead.
            # We use self.client so it respects base host.
            with self.client.post(
                "/api/auth/login", 
                json={"email": self.email, "password": self.password},
                headers={"Content-Type": "application/json"},
                name="/api/auth/login (setup)",
                catch_response=True
            ) as response:
                if response.status_code in [200, 201]: # NestJS standard create
                    data = response.json()
                    token = data.get("access_token") or data.get("accessToken") # Check payload structure
                    if token:
                         logging.info(f"Login successful. Token: {token[:10]}...")
                         self.analytics_headers["Authorization"] = f"Bearer {token}"
                    else:
                         logging.warning("Login succeeded but no token found in response.")
                else:
                    logging.warning(f"Login failed: {response.status_code}. Analytics tasks may receive 401.")
        except Exception as e:
            logging.error(f"Login exception: {e}")
