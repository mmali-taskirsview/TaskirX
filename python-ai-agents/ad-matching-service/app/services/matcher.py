"""
Ad Matching Service - Core Matching Engine
"""
import time
import logging
import numpy as np
from typing import List, Dict, Tuple
from collections import defaultdict
import redis
from sklearn.metrics.pairwise import cosine_similarity
from sklearn.feature_extraction.text import TfidfVectorizer
from pinecone import Pinecone, ServerlessSpec

from app.models.schemas import (
    MatchRequest,
    MatchResponse,
    AdRecommendation,
    UserProfile,
    MatchingStrategy
)
from app.config import settings

logger = logging.getLogger(__name__)


class AdMatcher:
    """
    AI-powered ad matching engine
    
    Implements multiple matching strategies:
    - Collaborative Filtering: User-user and item-item similarity
    - Content-Based: TF-IDF on categories, keywords
    - Performance-Based: Historical CTR, CVR, revenue
    - Hybrid: Weighted combination of all strategies
    """
    
    def __init__(self):
        self.campaigns = []
        self.user_campaign_matrix = None
        self.campaign_vectors = None
        self.tfidf_vectorizer = None
        
        # Performance metrics
        self.total_requests = 0
        self.total_recommendations = 0
        self.cache_hits = 0
        self.start_time = time.time()
        self.processing_times = []
        
        # Initialize Redis connection
        try:
            self.redis = redis.Redis(
                host=settings.REDIS_HOST,
                port=settings.REDIS_PORT,
                db=settings.REDIS_DB,
                password=settings.REDIS_PASSWORD,
                decode_responses=True
            )
            self.redis.ping()
            logger.info(f"Connected to Redis at {settings.REDIS_HOST}:{settings.REDIS_PORT}")
        except Exception as e:
            logger.error(f"Failed to connect to Redis: {e}")
            self.redis = None
        
        # Initialize Pinecone
        self.pc = None
        self.index = None
        if settings.USE_PINECONE and settings.PINECONE_API_KEY:
            try:
                self.pc = Pinecone(api_key=settings.PINECONE_API_KEY)
                self.index = self.pc.Index(settings.PINECONE_INDEX_NAME)
                logger.info(f"Connected to Pinecone index: {settings.PINECONE_INDEX_NAME}")
            except Exception as e:
                logger.error(f"Failed to connect to Pinecone: {e}")
        
        # User-campaign interaction history (in-memory fallback)
        self.user_interactions_fallback = defaultdict(lambda: {"viewed": set(), "clicked": set(), "converted": set()})
        
        self._initialize_models()
        self._load_campaigns()
    
    def _initialize_models(self):
        """Initialize ML models"""
        logger.info("Initializing ad matching models")
        
        # TF-IDF for content-based matching
        self.tfidf_vectorizer = TfidfVectorizer(
            max_features=100,
            ngram_range=(1, 2),
            stop_words='english'
        )
        
        logger.info("Models initialized successfully")
    
    def _load_campaigns(self):
        """Load active campaigns (mock data for now)"""
        # In production, load from NestJS API or database
        self.campaigns = self._generate_mock_campaigns()
        
        # Build campaign vectors for content-based matching
        self._build_campaign_vectors()
        
        logger.info(f"Loaded {len(self.campaigns)} campaigns")
    
    def _generate_mock_campaigns(self) -> List[Dict]:
        """Generate mock campaigns for demonstration"""
        import random
        categories = ["tech", "fashion", "travel", "food", "sports", "gaming", "finance", "health"]
        campaigns = []
        
        for i in range(100):
            cat_sample = random.sample(categories, k=random.randint(1, 3))
            campaigns.append({
                "id": f"camp-{i}",
                "name": f"Campaign {i} - {cat_sample[0].title()}",
                "advertiser_id": f"adv-{i%10}",
                "description": f"Best {cat_sample[0]} products for you",
                "categories": cat_sample,
                "ctr": random.uniform(0.01, 0.15),
                "cvr": random.uniform(0.005, 0.05),
                "bid_price": round(random.uniform(0.5, 5.0), 2),
                "status": "active",
                "avg_revenue_per_conversion": round(random.uniform(5.0, 50.0), 2),
                "creative_url": f"https://cdn.example.com/creatives/{i}.jpg",
                "landing_url": f"https://example.com/products/{i}",
                "keywords": [f"keyword-{k}" for k in range(3)]
            })
        return campaigns
    
    def _build_campaign_vectors(self):
        """Build TF-IDF vectors for campaigns"""
        if not self.campaigns:
            return
        
        # Create text representation of each campaign
        campaign_texts = []
        for campaign in self.campaigns:
            categories = campaign.get("categories", []) or []
            keywords = campaign.get("keywords", []) or []
            text = " ".join(categories + keywords).strip()
            if not text:
                text = campaign.get("description", "")
            campaign_texts.append(text)
        
        # Fit TF-IDF
        try:
            self.campaign_vectors = self.tfidf_vectorizer.fit_transform(campaign_texts)
            logger.info(f"Built campaign vectors: {self.campaign_vectors.shape}")
        except Exception as e:
            logger.error(f"Error building campaign vectors: {e}")
            self.campaign_vectors = None
    
    def _calculate_content_score(self, user_profile: UserProfile, campaign: Dict) -> float:
        """Calculate content similarity between user interests and campaign"""
        if self.campaign_vectors is None or not user_profile.categories:
            return 0.5 # Default score
            
        try:
            # Create user vector from their categories
            user_text = " ".join(user_profile.categories)
            user_vector = self.tfidf_vectorizer.transform([user_text])
            
            # Find campaign index
            try:
                # Assuming simple efficient lookup for now
                idx = next(i for i, c in enumerate(self.campaigns) if c["id"] == campaign["id"])
                
                # Calculate cosine similarity
                similarity = cosine_similarity(user_vector, self.campaign_vectors[idx]).flatten()[0]
                return float(similarity)
            except StopIteration:
                return 0.0
                
        except Exception as e:
            logger.error(f"Error calculating content score: {e}")
            return 0.0
    
    def _collaborative_filtering_score(self, user: UserProfile, campaign: Dict) -> float:
        """
        Calculate collaborative filtering score
        
        Based on:
        - Similar users who interacted with this campaign
        - User's past interaction patterns
        """
        score = 0.5  # Default neutral score
        
        campaign_id = campaign["id"]
        
        # Check if user has interacted with this campaign before
        if user.user_id:
            interactions = self._get_user_history(user.user_id)
            converted = interactions.get("converted", [])
            clicked = interactions.get("clicked", [])
            viewed = interactions.get("viewed", [])
            
            if campaign_id in converted:
                score = 0.95  # High score if user converted before
            elif campaign_id in clicked:
                score = 0.8  # Good score if user clicked before
            elif campaign_id in viewed:
                score = 0.6  # Moderate score if user viewed before
        
        # Check similar campaigns user interacted with (same categories)
        user_categories = set(user.interests + user.categories)
        campaign_categories = set(campaign["categories"])
        category_overlap = len(user_categories & campaign_categories)
        
        if category_overlap > 0:
            score += 0.1 * category_overlap
        
        # Check if user clicked/converted on similar campaigns
        if user.clicked_ads:
            similar_campaigns = [
                c for c in self.campaigns
                if c["id"] in user.clicked_ads and 
                len(set(c["categories"]) & campaign_categories) > 0
            ]
            if similar_campaigns:
                score += 0.15
        
        return min(score, 1.0)
    
    def _content_based_score(self, user: UserProfile, campaign: Dict) -> float:
        """
        Calculate content-based similarity score
        
        Based on:
        - TF-IDF similarity between user interests and campaign
        - Category/keyword matching
        """
        score = 0.0
        
        # Category matching
        user_categories = set(user.interests + user.categories)
        campaign_categories = set(campaign["categories"])
        
        if user_categories and campaign_categories:
            overlap = len(user_categories & campaign_categories)
            score += (overlap / max(len(user_categories), len(campaign_categories))) * 0.5
        
        # TF-IDF similarity (if available)
        if self.campaign_vectors is not None and user_categories:
            try:
                # Create user vector
                user_text = " ".join(list(user_categories))
                user_vector = self.tfidf_vectorizer.transform([user_text])
                
                # Find campaign index
                campaign_idx = next(
                    (i for i, c in enumerate(self.campaigns) if c["id"] == campaign["id"]),
                    None
                )
                
                if campaign_idx is not None:
                    # Calculate cosine similarity
                    similarity = cosine_similarity(
                        user_vector,
                        self.campaign_vectors[campaign_idx:campaign_idx+1]
                    )[0][0]
                    score += similarity * 0.5
            except Exception as e:
                logger.debug(f"TF-IDF similarity error: {e}")
        
        return min(score, 1.0)
    
    def _performance_score(self, campaign: Dict) -> float:
        """
        Calculate performance-based score
        
        Based on:
        - Historical CTR (click-through rate)
        - Historical CVR (conversion rate)
        - Revenue performance
        """
        # Normalize CTR (typical range 0-10%)
        ctr_score = min(campaign["ctr"] * 10, 1.0)
        
        # Normalize CVR (typical range 0-20%)
        cvr_score = min(campaign["cvr"] * 5, 1.0)
        
        # Revenue score (higher is better, normalize by max revenue)
        max_revenue = max(c["avg_revenue_per_conversion"] for c in self.campaigns)
        revenue_score = campaign["avg_revenue_per_conversion"] / max(max_revenue, 1)
        
        # Weighted combination
        performance_score = (ctr_score * 0.4) + (cvr_score * 0.3) + (revenue_score * 0.3)
        
        return performance_score
    
    def _calculate_hybrid_score(
        self, 
        user: UserProfile, 
        campaign: Dict,
        strategy: MatchingStrategy
    ) -> Tuple[float, float, float, float]:
        """
        Calculate hybrid score based on weights
        
        Returns: (overall_score, collab_score, content_score, perf_score)
        """
        # 1. Pinecone Retrieval (if enabled)
        if hasattr(self, 'index') and self.index and settings.USE_PINECONE:
            # Check if this campaign was in the Pinecone search results
            # (In a real system, we would batch this earlier)
            pass

        # 2. Content-based Score (Text similarity) (renumbering logic flow conceptually)
        content_score = self._calculate_content_score(user, campaign)
        
        # 3. Collaborative Filtering Score (Similar users/items)
        # Placeholder for real implementation (would query User-Item matrix)
        collab_score = 0.5 
        
        # 4. Performance Score (Historical CTR)
        perf_score = min(campaign.get("ctr", 0.0) * 10, 1.0) # Normalize 0.1 CTR to 1.0
        
        # Calculate weighted average based on strategy
        if strategy == MatchingStrategy.CONTENT_BASED:
            overall = (content_score * 0.8) + (perf_score * 0.2)
        elif strategy == MatchingStrategy.COLLABORATIVE:
            overall = (collab_score * 0.8) + (perf_score * 0.2)
        elif strategy == MatchingStrategy.PERFORMANCE_BASED:
            overall = perf_score
        else: # HYBRID (Default)
            overall = (
                (content_score * 0.4) + 
                (collab_score * 0.3) + 
                (perf_score * 0.3)
            )
            
        return min(overall, 1.0), collab_score, content_score, perf_score
    
    def _filter_campaigns(self, request: MatchRequest) -> List[Dict]:
        """Filter campaigns based on targeting and constraints"""
        filtered = []
        
        for campaign in self.campaigns:
            # Basic filters
            if campaign["status"] != "active":
                continue
            
            # Bid price filter
            if request.campaign_context.min_bid and campaign["bid_price"] < request.campaign_context.min_bid:
                continue
            if request.campaign_context.max_bid and campaign["bid_price"] > request.campaign_context.max_bid:
                continue
            
            # Required categories
            if request.campaign_context.required_categories:
                campaign_cats = set(campaign["categories"])
                required_cats = set(request.campaign_context.required_categories)
                if not required_cats.issubset(campaign_cats):
                    continue
            
            # Excluded categories
            if request.campaign_context.excluded_categories:
                campaign_cats = set(campaign["categories"])
                excluded_cats = set(request.campaign_context.excluded_categories)
                if campaign_cats & excluded_cats:
                    continue
            
            filtered.append(campaign)
        
        return filtered
    
    def _calculate_diversity_metrics(self, recommendations: List[AdRecommendation]) -> Tuple[float, float]:
        """Calculate diversity metrics for recommendations"""
        if not recommendations:
            return 0.0, 0.0
        
        # Category diversity
        all_categories = set()
        for rec in recommendations:
            all_categories.update(rec.categories)
        category_diversity = len(all_categories) / max(len(recommendations) * 2, 1)
        
        # Advertiser diversity
        advertisers = set(rec.advertiser_id for rec in recommendations)
        advertiser_diversity = len(advertisers) / len(recommendations)
        
        return category_diversity, advertiser_diversity
    
    def match(self, request: MatchRequest) -> MatchResponse:
        """
        Main matching function
        
        Process:
        1. Filter campaigns based on constraints
        2. Calculate scores for each campaign
        3. Sort by score
        4. Apply diversity
        5. Return top N recommendations
        """
        start_time = time.time()
        
        # Filter campaigns
        candidate_campaigns = self._filter_campaigns(request)
        total_candidates = len(candidate_campaigns)
        
        if not candidate_campaigns:
            logger.warning(f"No candidate campaigns found for request {request.request_id}")
            return MatchResponse(
                request_id=request.request_id,
                recommendations=[],
                total_candidates=0,
                strategy_used=request.strategy,
                processing_time_ms=0,
                category_diversity=0.0,
                advertiser_diversity=0.0
            )
        
        # 0. Vector Search Candidate Generation (Pinecone)
        if hasattr(self, 'index') and self.index and settings.USE_PINECONE:
            pinecone_results = self._search_pinecone(request.user_profile)
            if pinecone_results:
                pinecone_ids = {p['id'] for p in pinecone_results}
                # Boost candidates found in vector search, or use ONLY them if strict
                # Here we just flag them for boosting later
                for camp in candidate_campaigns:
                    if camp['id'] in pinecone_ids:
                        camp['_vector_boost'] = 0.2  # Add boost to vector matches
        
        # Calculate scores
        scored_campaigns = []
        for campaign in candidate_campaigns:
            overall, collab, content, perf = self._calculate_hybrid_score(
                request.user, 
                campaign,
                request.strategy
            )
            
            # Filter by minimum similarity
            if overall < settings.MIN_SIMILARITY_SCORE:
                continue
            
            # Create recommendation
            match_reasons = []
            if collab > 0.7:
                match_reasons.append("Similar to your past interactions")
            if content > 0.7:
                match_reasons.append("Matches your interests")
            if perf > 0.7:
                match_reasons.append("High-performing campaign")
            if not match_reasons:
                match_reasons.append("General relevance")
            
            recommendation = AdRecommendation(
                campaign_id=campaign["id"],
                campaign_name=campaign["name"],
                advertiser_id=campaign["advertiser_id"],
                overall_score=overall,
                collaborative_score=collab,
                content_score=content,
                performance_score=perf,
                bid_price=campaign["bid_price"],
                creative_url=campaign["creative_url"],
                landing_url=campaign["landing_url"],
                categories=campaign["categories"],
                predicted_ctr=campaign["ctr"],
                predicted_cvr=campaign["cvr"],
                predicted_revenue=campaign["avg_revenue_per_conversion"],
                match_reasons=match_reasons,
                confidence=max(overall, 0.5)  # Minimum 50% confidence
            )
            
            scored_campaigns.append((overall, recommendation))
        
        # Sort by score (descending)
        scored_campaigns.sort(key=lambda x: x[0], reverse=True)
        
        # Take top N
        max_results = min(request.max_results, settings.MAX_RECOMMENDATIONS)
        top_recommendations = [rec for score, rec in scored_campaigns[:max_results]]
        
        # Calculate diversity metrics
        cat_diversity, adv_diversity = self._calculate_diversity_metrics(top_recommendations)
        
        # Record metrics
        processing_time = (time.time() - start_time) * 1000
        self.total_requests += 1
        self.total_recommendations += len(top_recommendations)
        self.processing_times.append(processing_time)
        if len(self.processing_times) > 1000:
            self.processing_times.pop(0)
        
        # Build response
        response = MatchResponse(
            request_id=request.request_id,
            recommendations=top_recommendations,
            total_candidates=total_candidates,
            strategy_used=request.strategy,
            processing_time_ms=processing_time,
            category_diversity=cat_diversity,
            advertiser_diversity=adv_diversity
        )
        
        return response
    
    def record_interaction(self, user_id: str, campaign_id: str, interaction_type: str = "impression"):
        """Record user interaction with campaign"""
        if self.redis:
            try:
                # Store sets of campaigns per user interaction type
                key = f"user:{user_id}:interactions:{interaction_type}"
                self.redis.sadd(key, campaign_id)
                self.redis.expire(key, 86400 * 30) # 30 days retention
                
                # Also increment global campaign stats
                camp_key = f"campaign:{campaign_id}:stats"
                self.redis.hincrby(camp_key, interaction_type + "s", 1)
            except Exception as e:
                logger.error(f"Redis error in record_interaction: {e}")
        else:
             # Fallback to in-memory
             if interaction_type == "impression":
                 self.user_interactions_fallback[user_id]["viewed"].add(campaign_id)
             elif interaction_type == "click":
                 self.user_interactions_fallback[user_id]["clicked"].add(campaign_id)

    def _get_user_history(self, user_id: str) -> Dict[str, set]:
        """Get user interaction history"""
        history = {"viewed": set(), "clicked": set(), "converted": set()}
        
        if self.redis:
            try:
                history["viewed"] = self.redis.smembers(f"user:{user_id}:interactions:impression")
                history["clicked"] = self.redis.smembers(f"user:{user_id}:interactions:click")
                history["converted"] = self.redis.smembers(f"user:{user_id}:interactions:conversion")
            except Exception as e:
                logger.error(f"Redis error fetching user history: {e}")
                return self.user_interactions_fallback[user_id]
        else:
            return self.user_interactions_fallback[user_id]
        
        return history
    
    def get_metrics(self) -> Dict:
        """Get performance metrics"""
        uptime = time.time() - self.start_time
        avg_recommendations = self.total_recommendations / max(self.total_requests, 1)
        avg_time = np.mean(self.processing_times) if self.processing_times else 0
        cache_hit_rate = self.cache_hits / max(self.total_requests, 1)
        
        return {
            "total_requests": self.total_requests,
            "avg_recommendations": avg_recommendations,
            "avg_processing_time_ms": avg_time,
            "cache_hit_rate": cache_hit_rate,
            "uptime_seconds": uptime
        }
    
    def is_healthy(self) -> bool:
        """Health check"""
        return len(self.campaigns) > 0

    def _search_pinecone(self, user_profile: UserProfile, top_k: int = 20) -> List[Dict]:
        """Search similar ads in Pinecone"""
        if not self.index or not user_profile.categories:
            return []
            
        try:
            # Generate embedding for user profile (using random/mock for now as we lack an embedding model)
            # In production: vector = model.encode(" ".join(user_profile.categories)).tolist()
            user_vector = [np.random.uniform(-1, 1) for _ in range(1536)] 

            results = self.index.query(
                vector=user_vector,
                top_k=top_k,
                include_metadata=True
            )
            
            matches = []
            for match in results.matches:
                matches.append({
                    "id": match.id,
                    "score": match.score,
                    **match.metadata
                })
            
            return matches
            
        except Exception as e:
            logger.error(f"Pinecone search failed: {e}")
            return []


# Global matcher instance
matcher = AdMatcher()
