import redis
import logging
import requests

logger = logging.getLogger(__name__)

class IPReputationService:
    def __init__(self, redis_client: redis.Redis, api_key: str = None):
        self.redis = redis_client
        self.api_key = api_key
        # Cache results for 24 hours
        self.cache_ttl = 86400 

    def is_ip_blacklisted(self, ip_address: str) -> bool:
        """
        Check if IP is blacklisted.
        Checks local Redis cache first, then (optional) external service.
        """
        # 1. Check Redis Cache/Blocklist
        # Keys: 'fraud:ip:blocklist:<ip>'
        cache_key = f"fraud:ip:blocklist:{ip_address}"
        try:
            is_blocked = self.redis.get(cache_key)
            if is_blocked:
                return True
        except redis.RedisError as e:
            logger.error(f"Redis error checking IP blocklist: {e}")

        # 2. (Optional) Check External API if configured
        # This is a placeholder for services like IPQualityScore, AbuseIPDB, etc.
        if self.api_key:
             return self._check_external_service(ip_address)
        
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

    def _check_external_service(self, ip_address: str) -> bool:
        # Placeholder implementation
        # In production this would make an HTTP request
        return False
