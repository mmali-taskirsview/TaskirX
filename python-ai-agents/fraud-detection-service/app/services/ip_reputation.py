import redis
import logging
import requests
import random
from app.config import settings

logger = logging.getLogger(__name__)

class IPReputationService:
    def __init__(self, redis_client: redis.Redis, api_key: str = None):
        self.redis = redis_client
        self.api_key = api_key
        # Cache results for 24 hours
        self.cache_ttl = 86400 

        # Fallback for old style initialization
        if not api_key and settings.IP_REPUTATION_API_KEY:
            self.api_key = settings.IP_REPUTATION_API_KEY

    def is_ip_blacklisted(self, ip_address: str) -> bool:
        """
        Check if IP is blacklisted.
        Checks local Redis cache first, then (optional) external service.
        """
        # 1. Check Redis Cache/Blocklist
        # Format: 'fraud:ip:blocklist:<ip>' -> '1' (or reason string)
        cache_key = f"fraud:ip:blocklist:{ip_address}"
        try:
            is_blocked = self.redis.get(cache_key)
            if is_blocked:
                # If value is '1' or non-empty string, it's blocked
                return True
        except Exception as e:
            logger.error(f"Redis error checking IP blocklist: {e}")
            return False # Fail open on Redis error
        
        # 2. Mock external check if not in Redis
        # In a real scenario, this would call AbuseIPDB or similar
        # For simulation, we randomly block ~1%
        
        # Only simulate check if explicitly enabled via config or simple mock logic
        # For now, we assume if Redis doesn't have it, it's clean, UNLESS we call external API.
        
        # Here we can add logic to populate Redis from external API if missing
        if self.api_key:
             # Real API Call (mocked for now)
             is_suspicious = self._check_external_api(ip_address)
             if is_suspicious:
                 self.redis.setex(cache_key, self.cache_ttl, "blocked_by_api")
                 return True
        
        return False

    def _check_external_api(self, ip: str) -> bool:
        """
        Check external IP reputation database.
        Supports AbuseIPDB if a valid API key is present.
        """
        if not self.api_key or self.api_key == "demo-key":
            # Mock Behavior for testing/demo
            # Block IPs ending in .99
            return ip.endswith(".99")

        try:
            # Real AbuseIPDB Integration
            url = 'https://api.abuseipdb.com/api/v2/check'
            querystring = {
                'ipAddress': ip,
                'maxAgeInDays': '90'
            }
            headers = {
                'Accept': 'application/json',
                'Key': self.api_key
            }
            
            # Short timeout to prevent blocking bidding flow
            response = requests.request(method='GET', url=url, headers=headers, params=querystring, timeout=1.0)
            
            if response.status_code == 200:
                data = response.json()
                score = data.get('data', {}).get('abuseConfidenceScore', 0)
                # Any score > 50 is considered high risk
                if score > 50:
                    logger.warning(f"IP {ip} flagged by AbuseIPDB with score {score}")
                    return True
            elif response.status_code == 429:
                logger.warning("AbuseIPDB rate limit exceeded")
            else:
                logger.error(f"AbuseIPDB error: {response.status_code} - {response.text}")
                
        except Exception as e:
            logger.error(f"Error calling AbuseIPDB: {e}")
            
        return False


    def block_ip(self, ip_address: str, reason: str = "manual_block", ttl: int = 86400):
        """
        Manually block an IP address.
        """
        cache_key = f"fraud:ip:blocklist:{ip_address}"
        try:
            self.redis.setex(cache_key, ttl, reason)
            return True
        except redis.RedisError as e:
            logger.error(f"Redis error blocking IP: {e}")
            return False
