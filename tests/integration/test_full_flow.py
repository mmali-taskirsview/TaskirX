import time
import requests
import json
import logging
import sys

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Base URLs (as defined in docker-compose.yml)
# If running inside docker, these would be service names.
# If running locally (with ports mapped), these are localhost.
BIDDING_ENGINE_URL = "http://localhost:8080"
AD_MATCHING_URL = "http://localhost:6002"
FRAUD_DETECTION_URL = "http://localhost:6001"
BID_OPTIMIZATION_URL = "http://localhost:6003"
BACKEND_URL = "http://localhost:3000"

import datetime

# ... (rest of imports)

# Test data (aligned with Go BidRequest struct)
SAMPLE_BID_REQUEST = {
    "id": "test-req-full-flow-001",
    "timestamp": datetime.datetime.now().isoformat() + "Z", # "2026-02-17T12:34:56.789Z"
    "publisher_id": "pub-001",
    "ad_slot": {
        "id": "slot-001",
        "dimensions": [300, 250],
        "position": "above-fold",
        "formats": ["banner"]
    },
    "user": {
        "id": "test-user-001",
        "country": "US",
        "age": 30,
        "gender": "M"
    },
    "device": {
        "type": "mobile",
        "os": "Android",
        "browser": "Chrome",
        "ip": "203.0.113.1",
        "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
        "geo": {
            "country": "US",
            "city": "New York",
            "lat": 40.7128,
            "lon": -74.0060
        }
    },
    "context": {
        "site_domain": "news.com",
        "site_category": "IAB1"
    }
}

def wait_for_service(url, name, retries=5, delay=2):
    """Wait for a service to become available."""
    logger.info(f"Waiting for {name} ({url})...")
    for i in range(retries):
        try:
            # Try health endpoint first (most services usually expose /health or check root)
            # Adjust endpoints based on service
            
            check_url = f"{url}/health"
            if name == "Go Bidding Engine":
                check_url = f"{url}/health"
            elif name == "Ad Matching Service":
                check_url = f"{url}/api/health" 
            elif name == "Fraud Detection Service":
                check_url = f"{url}/api/health"
            elif name == "Bid Optimization Service":
                check_url = f"{url}/api/health"
            
            response = requests.get(check_url, timeout=2)
            if response.status_code == 200:
                logger.info(f"✅ {name} is UP!")
                return True
        except requests.RequestException:
            pass
        
        logger.warning(f"Waiting for {name}... ({i+1}/{retries})")
        time.sleep(delay)
    
    logger.error(f"❌ {name} failed to start.")
    return False

def test_full_flow():
    """Execute a full end-to-end bid request test."""
    
    logger.info("🚀 Starting End-to-End Integration Test")
    
    # 1. Verify all services are reachable
    services_up = True
    services_up &= wait_for_service(BIDDING_ENGINE_URL, "Go Bidding Engine")
    # For a true E2E, we might not have direct access to internal AI services if not port mapped,
    # but docker-compose exposes them on 6001, 6002, 6003, so we can check them.
    services_up &= wait_for_service(FRAUD_DETECTION_URL, "Fraud Detection Service")
    services_up &= wait_for_service(AD_MATCHING_URL, "Ad Matching Service")
    services_up &= wait_for_service(BID_OPTIMIZATION_URL, "Bid Optimization Service")
    
    if not services_up:
        logger.error("🛑 One or more services are down. Aborting test.")
        sys.exit(1)

    # 2. Check Service Health status specifically
    # Ensuring they are not just reachable but "healthy" (e.g. database connected)
    # (Skipping deep health check for brevity, relying on /health 200 OK)

    # 3. Trigger Campaign Refresh to load seeded data
    logger.info("🔄 Triggering Campaign Refresh...")
    try:
        refresh_resp = requests.post(f"{BIDDING_ENGINE_URL}/refresh", timeout=5)
        if refresh_resp.status_code == 200:
            logger.info("✅ Campaigns refreshed successfully.")
        else:
            logger.warning(f"⚠️ Campaign refresh returned status {refresh_resp.status_code}")
            logger.warning(f"Response: {refresh_resp.text}")

    except Exception as e:
        logger.warning(f"⚠️ Failed to refresh campaigns: {e}")

    # 4. Send Bid Request to Go Engine
    logger.info("📤 Sending Bid Request to Go Bidding Engine...")
    try:
        response = requests.post(f"{BIDDING_ENGINE_URL}/bid", json=SAMPLE_BID_REQUEST, timeout=5)
        
        logger.info(f"📥 Received Response: Status {response.status_code}")
        
        if response.status_code == 200:
            bid_response = response.json()
            logger.info("✅ Valid Bid Response Received!")
            logger.info(json.dumps(bid_response, indent=2))
            
            # 4. Analyze Logic Execution
            # Check if bid_price exists (simplified response format)
            if "bid_price" in bid_response and bid_response["bid_price"] > 0:
                logger.info("✅ Bid was generated (Ad Matching worked!)")
                logger.info(f"   Bid Price: {bid_response['bid_price']}")
                logger.info(f"   Campaign ID: {bid_response.get('campaign_id', 'N/A')}")
                
                # Check for AI-driven modifications (logs or specific headers/metadata if available)
                # Without direct log access, we infer success from non-empty bid.
            elif "seatbid" in bid_response and len(bid_response["seatbid"]) > 0:
                 # Support OpenRTB format just in case
                logger.info("✅ Bids were generated (OpenRTB format!)")
                bid = bid_response["seatbid"][0]["bid"][0]
                logger.info(f"   Bid Price: {bid['price']}")
                logger.info(f"   Ad ID: {bid['adid']}")
            else:
                logger.warning("⚠️ No bids returned. Check if campaigns exist or if Fraud/Optimization blocked it.")
                
        elif response.status_code == 204:
            logger.info("ℹ️ No Content (Valid, but no bid matched).")
        else:
            logger.error(f"❌ Unexpected Status Code: {response.status_code}")
            logger.error(response.text)
            
    except requests.RequestException as e:
        logger.error(f"❌ Request failed: {e}")

if __name__ == "__main__":
    test_full_flow()
