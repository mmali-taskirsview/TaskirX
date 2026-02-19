"""
Fraud Detection Service - Machine Learning Model
"""
import os
import time
import logging
import numpy as np
import joblib
import redis
from typing import Dict, Tuple, List
from datetime import datetime, timedelta
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier
from sklearn.preprocessing import StandardScaler

from app.models.schemas import (
    FraudCheckRequest, 
    FraudCheckResponse, 
    FraudIndicators, 
    FraudRiskLevel
)
from app.config import settings
from app.services.ip_reputation import IPReputationService

logger = logging.getLogger(__name__)


class FraudDetector:
    """
    AI-powered fraud detection engine
    
    Uses ensemble of machine learning models:
    - Random Forest for general pattern detection
    - Gradient Boosting for complex decision boundaries
    - Rule-based system for known fraud patterns
    """
    
    def __init__(self):
        self.model = None
        self.scaler = None
        self.model_version = "1.0.0"
        self.threshold = settings.MODEL_THRESHOLD
        
        # Performance metrics
        self.total_predictions = 0
        self.fraud_detected = 0
        self.start_time = time.time()
        self.processing_times = []
        
        # Known fraud patterns (IP blacklist, device fingerprints, etc.)
        self.blacklisted_ips = set()
        self.suspicious_user_agents = set()
        self.datacenter_ranges = set()

        # Initialize Redis
        try:
            self.redis_client = redis.Redis(
                host=settings.REDIS_HOST,
                port=settings.REDIS_PORT,
	            db=settings.REDIS_DB,
                password=settings.REDIS_PASSWORD,
                decode_responses=True
            )
            self.redis_client.ping()
            # Pass API key from settings
            self.ip_reputation = IPReputationService(self.redis_client, api_key=settings.IP_REPUTATION_API_KEY)
            logger.info(f"Connected to Redis for IP Reputation (API Key Present: {bool(settings.IP_REPUTATION_API_KEY)})")
        except Exception as e:
            logger.error(f"Failed to connect to Redis: {e}")
            self.redis_client = None

            self.ip_reputation = None
        
        self._initialize_model()
        self._load_fraud_patterns()
    
    def _initialize_model(self):
        """Initialize or load the ML model"""
        try:
            if os.path.exists(settings.MODEL_PATH):
                logger.info(f"Loading pre-trained model from {settings.MODEL_PATH}")
                self.model = joblib.load(settings.MODEL_PATH)
                self.scaler = joblib.load(settings.FEATURE_SCALER_PATH)
            else:
                logger.warning("No pre-trained model found. Creating new model.")
                self._create_default_model()
                
            logger.info("Fraud detection model initialized successfully")
        except Exception as e:
            logger.error(f"Error initializing model: {e}")
            self._create_default_model()
    
    def _create_default_model(self):
        """Create a default model for demonstration"""
        logger.info("Creating default fraud detection model")
        
        # Random Forest with balanced class weights
        self.model = RandomForestClassifier(
            n_estimators=100,
            max_depth=10,
            min_samples_split=5,
            class_weight='balanced',
            random_state=42,
            n_jobs=-1
        )
        
        self.scaler = StandardScaler()
        
        # Train on synthetic data (in production, use real historical data)
        self._train_on_synthetic_data()
    
    def _train_on_synthetic_data(self):
        """Train model on synthetic fraud data"""
        logger.info("Training model on synthetic data")
        
        # Generate synthetic features (in production, use real data)
        np.random.seed(42)
        n_samples = 10000
        
        # Create synthetic dataset
        # Features: [click_freq, ip_reputation, device_age, geo_consistency, behavior_score, ...]
        X = np.random.randn(n_samples, 15)
        
        # Labels: fraud (1) vs legitimate (0)
        # ~10% fraud rate (realistic for ad fraud)
        y = np.random.choice([0, 1], size=n_samples, p=[0.9, 0.1])
        
        # Add some patterns for fraud cases
        fraud_indices = np.where(y == 1)[0]
        X[fraud_indices, 0] += 2  # Higher click frequency
        X[fraud_indices, 1] -= 2  # Lower IP reputation
        X[fraud_indices, 4] -= 1.5  # Lower behavior score
        
        # Fit scaler and model
        X_scaled = self.scaler.fit_transform(X)
        self.model.fit(X_scaled, y)
        
        logger.info(f"Model trained on {n_samples} samples")
    
    def _load_fraud_patterns(self):
        """Load known fraud patterns (blacklists, rules)"""
        # In production, load from database or file
        self.blacklisted_ips = {
            "1.2.3.4",  # Example blocked IP
            "192.168.1.1"  # Private IP (shouldn't appear in ads)
        }
        
        self.suspicious_user_agents = {
            "bot",
            "crawler",
            "spider",
            "scraper",
            "headless"
        }
        
        logger.info("Fraud patterns loaded")
    
    def extract_features(self, request: FraudCheckRequest) -> np.ndarray:
        """
        Extract numerical features from request
        
        Features (15 total):
        1. Click frequency (clicks/hour)
        2. IP reputation score
        3. Device type score
        4. Geo consistency score
        5. User behavior score
        6. Time of day (hour)
        7. Day of week
        8. Session age
        9. Impression frequency
        10. Conversion rate
        11. Avg time on site
        12. Bounce rate
        13. Browser version age
        14. Screen resolution score
        15. Language consistency
        """
        features = []
        
        # 1. Click frequency
        clicks_hour = request.behavior.clicks_last_hour if request.behavior else 0
        features.append(clicks_hour)
        
        # 2. IP reputation (simplified - in production, use IP intelligence API)
        ip_reputation = 1.0 if request.ip_address not in self.blacklisted_ips else 0.0
        features.append(ip_reputation)
        
        # 3. Device type score (mobile=0.8, desktop=0.9, tablet=0.85, unknown=0.3)
        device_scores = {"mobile": 0.8, "desktop": 0.9, "tablet": 0.85, "tv": 0.7, "unknown": 0.3}
        features.append(device_scores.get(request.device.type, 0.5))
        
        # 4. Geo consistency (check if country matches timezone, etc.)
        geo_score = 1.0  # Simplified
        features.append(geo_score)
        
        # 5. User behavior score
        if request.behavior:
            clicks_24h = request.behavior.clicks_last_24h
            impressions_24h = request.behavior.impressions_last_24h
            ctr = clicks_24h / max(impressions_24h, 1)
            behavior_score = min(ctr * 10, 1.0)  # Normalize CTR
        else:
            behavior_score = 0.5  # Neutral if no data
        features.append(behavior_score)
        
        # 6-7. Time features
        now = request.timestamp
        features.append(now.hour / 24.0)  # Hour normalized
        features.append(now.weekday() / 7.0)  # Day of week normalized
        
        # 8. Session age (if available)
        session_age = 0.5  # Default
        if request.click_timestamp:
            age_seconds = (now - request.click_timestamp).total_seconds()
            session_age = min(age_seconds / 3600, 1.0)  # Max 1 hour
        features.append(session_age)
        
        # 9-12. Behavior metrics
        impressions_hour = request.behavior.impressions_last_hour if request.behavior else 0
        features.append(impressions_hour / 100.0)  # Normalize
        
        conversions = request.behavior.conversions_last_24h if request.behavior else 0
        conversion_rate = conversions / max(clicks_hour, 1)
        features.append(min(conversion_rate, 1.0))
        
        avg_time = request.behavior.avg_time_on_site if request.behavior else 30
        features.append(min(avg_time / 300.0, 1.0))  # Max 5 minutes
        
        bounce_rate = request.behavior.bounce_rate if request.behavior else 0.5
        features.append(bounce_rate)
        
        # 13-15. Device features
        browser_score = 0.9 if request.device.browser else 0.5
        features.append(browser_score)
        
        resolution_score = 0.9 if request.device.screen_resolution else 0.5
        features.append(resolution_score)
        
        language_score = 0.9 if request.device.language else 0.5
        features.append(language_score)
        
        return np.array(features).reshape(1, -1)
    
    def check_rules(self, request: FraudCheckRequest) -> Tuple[bool, List[str], FraudIndicators]:
        """
        Rule-based fraud detection (fast pre-screening)
        
        Returns:
            (is_fraud, reasons, indicators)
        """
        indicators = FraudIndicators()
        reasons = []
        is_fraud = False
        
        # Rule 1: Blacklisted IP
        if request.ip_address in self.blacklisted_ips:
            indicators.suspicious_ip = True
            reasons.append("IP address is blacklisted")
            is_fraud = True
        
        # Rule 2: Private/Local IP
        if request.ip_address.startswith(("192.168.", "10.", "127.")):
            indicators.suspicious_ip = True
            reasons.append("Private IP address detected")
            is_fraud = True
        
        # Rule 3: Suspicious User Agent
        user_agent = (request.device.user_agent or "").lower()
        if any(pattern in user_agent for pattern in self.suspicious_user_agents):
            indicators.bot_detected = True
            indicators.suspicious_user_agent = True
            reasons.append("Bot or crawler detected in user agent")
            is_fraud = True
        
        # Rule 4: High click frequency
        if request.behavior and request.behavior.clicks_last_hour > 50:
            indicators.high_click_frequency = True
            reasons.append(f"Abnormally high click frequency: {request.behavior.clicks_last_hour}/hour")
            is_fraud = True
        
        # Rule 5: Missing critical information
        if not request.device.type or request.device.type == "unknown":
            indicators.suspicious_device = True
            reasons.append("Unknown device type")
        
        # Rule 6: Impossible travel (if we have historical geo data)
        # This would require checking user's recent locations
        # indicators.impossible_travel = True
        
        # Rule 7: Device fingerprint mismatch
        # This would require storing device fingerprints
        # indicators.device_fingerprint_mismatch = True
        
        # Rule 8: Zero or excessive conversion rate
        if request.behavior:
            clicks = request.behavior.clicks_last_24h
            conversions = request.behavior.conversions_last_24h
            if clicks > 0:
                conv_rate = conversions / clicks
                if conv_rate > 0.5:  # >50% conversion rate is suspicious
                    indicators.behavioral_anomaly = True
                    reasons.append(f"Abnormally high conversion rate: {conv_rate:.1%}")
        
        return is_fraud, reasons, indicators
    
    def predict(self, request: FraudCheckRequest) -> FraudCheckResponse:
        """
        Main fraud detection prediction
        
        Combines rule-based and ML-based detection
        """
        start_time = time.time()
        
        # Step 1: Rule-based pre-screening (fast)
        rule_fraud, rule_reasons, indicators = self.check_rules(request)
        
        # Step 2: ML model prediction (if not already flagged)
        if not rule_fraud:
            try:
                # Extract features
                features = self.extract_features(request)
                
                # Scale features
                features_scaled = self.scaler.transform(features)
                
                # Get probability of fraud
                fraud_proba = self.model.predict_proba(features_scaled)[0][1]
                
                # Get prediction
                is_fraud = fraud_proba > self.threshold
                
                # Get feature importances for explanation (if available)
                if hasattr(self.model, 'feature_importances_'):
                    importances = self.model.feature_importances_
                    top_features = np.argsort(importances)[-3:][::-1]
                    
                    feature_names = [
                        "click_frequency", "ip_reputation", "device_type", 
                        "geo_consistency", "behavior_score", "hour", "day",
                        "session_age", "impression_freq", "conversion_rate",
                        "avg_time", "bounce_rate", "browser", "resolution", "language"
                    ]
                    
                    for idx in top_features:
                        if importances[idx] > 0.1:
                            rule_reasons.append(f"ML indicator: {feature_names[idx]}")
                
            except Exception as e:
                logger.error(f"ML prediction error: {e}")
                fraud_proba = 0.5
                is_fraud = False
        else:
            # Rule-based fraud already detected
            fraud_proba = 0.95
            is_fraud = True
        
        # Step 3: Determine risk level
        if fraud_proba >= 0.9:
            risk_level = FraudRiskLevel.CRITICAL
        elif fraud_proba >= 0.7:
            risk_level = FraudRiskLevel.HIGH
        elif fraud_proba >= 0.4:
            risk_level = FraudRiskLevel.MEDIUM
        else:
            risk_level = FraudRiskLevel.LOW
        
        # Step 4: Recommend action
        if fraud_proba >= 0.9:
            action = "block"
        elif fraud_proba >= 0.7:
            action = "flag"
        else:
            action = "allow"
        
        # Step 5: Calculate confidence
        # Higher confidence if consistent with rules or very clear ML prediction
        confidence = abs(fraud_proba - 0.5) * 2  # 0.5 = neutral, maps to 0 confidence
        if rule_fraud:
            confidence = max(confidence, 0.9)  # High confidence for rule violations
        
        # Record metrics
        processing_time = (time.time() - start_time) * 1000  # milliseconds
        self.total_predictions += 1
        if is_fraud:
            self.fraud_detected += 1
        self.processing_times.append(processing_time)
        if len(self.processing_times) > 1000:
            self.processing_times.pop(0)
        
        # Build response
        response = FraudCheckResponse(
            request_id=request.request_id,
            is_fraud=is_fraud,
            fraud_score=fraud_proba,
            risk_level=risk_level,
            confidence=confidence,
            indicators=indicators,
            reasons=rule_reasons or ["No fraud indicators detected"],
            recommended_action=action,
            processing_time_ms=processing_time,
            model_version=self.model_version
        )
        
        return response
    
    def get_metrics(self) -> Dict:
        """Get performance metrics"""
        uptime = time.time() - self.start_time
        fraud_rate = self.fraud_detected / max(self.total_predictions, 1)
        avg_time = np.mean(self.processing_times) if self.processing_times else 0
        
        return {
            "total_requests": self.total_predictions,
            "fraud_detected": self.fraud_detected,
            "fraud_rate": fraud_rate,
            "avg_processing_time_ms": avg_time,
            "uptime_seconds": uptime
        }
    
    def is_healthy(self) -> bool:
        """Health check"""
        return self.model is not None and self.scaler is not None

    def check_fraud(self, request: FraudCheckRequest) -> FraudCheckResponse:
        """
        Evaluate a bid request for fraud likelihood
        """
        start_time = time.time()
        
        # Step 1: Rule-based Filtering (Fast Check)
        rule_fraud = False
        rule_reasons = []
        
        # Initialize indicators
        indicators_obj = FraudIndicators()
        
        # 1. IP Blacklist (using Reputation Service)
        ip_blocked = False
        if self.ip_reputation:
            if self.ip_reputation.is_ip_blacklisted(request.ip_address):
                ip_blocked = True
                rule_reasons.append("IP is blacklisted (Reputation Service)")
                indicators_obj.suspicious_ip = True
        
        # Fallback to local blacklist set if not blocked by service
        if not ip_blocked and request.ip_address in self.blacklisted_ips:
            rule_reasons.append("IP is in static blacklist")
            indicators_obj.suspicious_ip = True
            ip_blocked = True
            
        if ip_blocked:
            rule_fraud = True
        
        # 2. Suspicious User Agent
        user_agent = (request.device.user_agent or "").lower()
        if any(pattern in user_agent for pattern in self.suspicious_user_agents):
            indicators_obj.suspicious_user_agent = True
            rule_reasons.append("Suspicious user agent detected")
            rule_fraud = True
        
        # 3. High Click Frequency
        if request.behavior and request.behavior.clicks_last_hour > 100:
            indicators_obj.high_click_frequency = True
            rule_reasons.append("Abnormally high click frequency")
            rule_fraud = True
        
        # 4. Missing Device Information
        if not request.device.type or request.device.type == "unknown":
            indicators_obj.suspicious_device = True
            rule_reasons.append("Device type is unknown")
            rule_fraud = True
        
        # 5. Private or Reserved IP Ranges
        if request.ip_address.startswith(("192.168.", "10.", "127.")):
            indicators_obj.suspicious_ip = True
            rule_reasons.append("Private or reserved IP address")
            rule_fraud = True
        
        # 6. Impossible Travel (if geo data is available)
        # TODO: Implement impossible travel detection based on geo history
        
        # 7. Device Fingerprint Mismatch
        # TODO: Implement device fingerprinting and mismatch detection
        
        # 8. Zero or Excessive Conversion Rate
        if request.behavior:
            clicks = request.behavior.clicks_last_24h
            conversions = request.behavior.conversions_last_24h
            if clicks > 0:
                conv_rate = conversions / clicks
                if conv_rate == 0:
                    indicators_obj.behavioral_anomaly = True
                    rule_reasons.append("Zero conversion rate")
                    rule_fraud = True
                elif conv_rate > 0.5:
                    indicators_obj.behavioral_anomaly = True
                    rule_reasons.append("Abnormally high conversion rate")
                    rule_fraud = True
        
        # Build response
        response = FraudCheckResponse(
            request_id=request.request_id,
            is_fraud=rule_fraud,
            fraud_score=1.0 if rule_fraud else 0.0,
            risk_level=FraudRiskLevel.CRITICAL if rule_fraud else FraudRiskLevel.LOW,
            confidence=1.0 if rule_fraud else 0.5,
            indicators=indicators_obj,
            reasons=rule_reasons,
            recommended_action="block" if rule_fraud else "allow",
            processing_time_ms=(time.time() - start_time) * 1000,
            model_version=self.model_version
        )
        
        return response
    

# Global detector instance
detector = FraudDetector()
