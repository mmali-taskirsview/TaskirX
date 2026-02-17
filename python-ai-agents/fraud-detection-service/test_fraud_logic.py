import sys
import os
import time
from datetime import datetime, timedelta
import asyncio

# Add the current directory to sys.path
sys.path.insert(0, os.path.abspath(os.path.dirname(__file__)))

from app.models.schemas import (
    FraudCheckRequest,
    DeviceInfo,
    DeviceType,
    GeoInfo,
    UserBehavior,
    FraudRiskLevel
)
from app.services.fraud_detector import FraudDetector

def test_fraud_detection():
    print("Initializing FraudDetector (this may take a moment to train synthetic model)...")
    detector = FraudDetector()
    
    # Create legitimate request
    legit_req = FraudCheckRequest(
        request_id="req-legit-001",
        ip_address="203.0.113.45",
        campaign_id="camp-123",
        publisher_id="pub-456",
        advertiser_id="adv-789",
        device=DeviceInfo(
            type=DeviceType.DESKTOP,
            os="Windows",
            browser="Chrome",
            user_agent="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
        ),
        geo=GeoInfo(
            country="US",
            city="New York",
            timezone="America/New_York"
        ),
        behavior=UserBehavior(
            clicks_last_hour=2,
            clicks_last_24h=10,
            impressions_last_24h=500,
            avg_time_on_site=120.5
        )
    )
    
    # Create fraudulent request (bot-like behavior)
    fraud_req = FraudCheckRequest(
        request_id="req-fraud-001",
        ip_address="192.168.1.1", # Private IP often caught by rules
        campaign_id="camp-123",
        publisher_id="pub-456",
        advertiser_id="adv-789",
        device=DeviceInfo(
            type=DeviceType.UNKNOWN,
            os="Unknown",
            user_agent="GoogleBot/2.1" # Suspicious UA
        ),
        geo=GeoInfo(
            country="US"
        ),
        behavior=UserBehavior(
            clicks_last_hour=1000, # Impossible for human
            clicks_last_24h=20000,
            impressions_last_24h=20000,
            avg_time_on_site=0.1
        )
    )
    
    # Detect legit
    print("\nAnalyzing Legitimate Request...")
    # The public method is `detect_fraud`? No, let's verify.
    # Reading the file again, the method is `predict`.
    response = detector.predict(legit_req)
    
    print(f"Risk Level: {response.risk_level}")
    print(f"Fraud Probability: {response.fraud_score:.4f}")
    print(f"Action: {response.recommended_action}")
    print(f"Reasons: {response.reasons}")
    
    assert response.risk_level == FraudRiskLevel.LOW
    assert response.recommended_action == "allow"
    
    # Detect fraud
    print("\nAnalyzing Fraudulent Request...")
    response_fraud = detector.predict(fraud_req)
    
    print(f"Risk Level: {response_fraud.risk_level}")
    print(f"Fraud Probability: {response_fraud.fraud_score:.4f}")
    print(f"Action: {response_fraud.recommended_action}")
    print(f"Reasons: {response_fraud.reasons}")
    
    assert response_fraud.risk_level in [FraudRiskLevel.HIGH, FraudRiskLevel.CRITICAL]
    assert response_fraud.recommended_action in ["block", "flag"]
    
    print("\n✅ Fraud Detection Test Passed!")

if __name__ == "__main__":
    try:
        test_fraud_detection()
    except Exception as e:
        print(f"\n❌ Test Failed: {e}")
        import traceback
        traceback.print_exc()
