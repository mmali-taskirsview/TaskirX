
import sys
import os
import time
from datetime import datetime
import asyncio

# Add the current directory to sys.path
sys.path.insert(0, os.path.abspath(os.path.dirname(__file__)))

from app.models.schemas import (
    MatchRequest,
    UserProfile,
    AdSlotInfo,
    CampaignContext,
    MatchingStrategy
)
from app.services.matcher import AdMatcher

def test_ad_matching():
    print("Initializing AdMatcher (this may take a moment to fit TF-IDF)...")
    matcher = AdMatcher()
    
    # Create user profile with interests
    user = UserProfile(
        user_id="user-123",
        country="US",
        interests=["tech", "gaming"], # Matches 'tech' and 'gaming' categories exactly
        categories=["tech", "gaming"],
        device_type="desktop"
    )
    
    # Create context
    context = CampaignContext(
        publisher_id="pub-001",
        page_category="tech",
        keywords=["tech", "gaming"]
    )
    
    # Create request
    request = MatchRequest(
        request_id="req-match-001",
        user=user,
        ad_slot=AdSlotInfo(
            slot_id="slot-1",
            dimensions=[300, 250],
            format="banner"
        ),
        campaign_context=context,
        strategy=MatchingStrategy.CONTENT_BASED,
        max_results=5
    )
    
    # Run matching
    print("\n--- Running Ad Matching ---")
    response = matcher.match(request)
    
    print(f"Total Candidates: {response.total_candidates}")
    print(f"Recommendations: {len(response.recommendations)}")
    
    if response.recommendations:
        rec = response.recommendations[0]
        print("\nTop Recommendation:")
        print(f"Campaign: {rec.campaign_name}")
        print(f"Score: {rec.overall_score:.4f}")
        print(f"Bid: ${rec.bid_price}")
        print(f"Categories: {rec.categories}")
        print(f"Reasons: {rec.match_reasons}")
        
    assert response.total_candidates > 0
    # There should be at least one match since we have 100 campaigns and 'tech' is common
    assert len(response.recommendations) > 0
    
    print("\n✅ Ad Matching Test Passed!")

if __name__ == "__main__":
    test_ad_matching()
