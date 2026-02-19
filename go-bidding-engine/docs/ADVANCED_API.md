# TaskirX Go Bidding Engine - Advanced Services API

## Overview

The Advanced Services API provides access to 11 sophisticated advertising features:

1. **Bid Landscape Analysis** - Analyze market bid patterns and optimize bid prices
2. **Creative Optimization** - Multi-armed bandit creative selection
3. **Incrementality Testing** - Measure true lift from advertising
4. **Privacy Sandbox** - Topics API and FLEDGE/Protected Audience support
5. **Contextual AI** - Content analysis and brand safety
6. **Real-Time Alerts** - Budget monitoring and anomaly detection
7. **Competitive Intelligence** - Market analysis and competitor tracking
8. **Unified ID** - Cross-provider identity resolution
9. **Dynamic Bid Adjustments** - ML-based real-time bid optimization (NEW)
10. **Lookalike Audiences** - Audience expansion through similarity modeling (NEW)
11. **User Clustering** - K-means based user segmentation (NEW)

## Base URL

```
http://localhost:5000/api/advanced
```

## Authentication

All endpoints accept requests without authentication in development mode. In production, include the API key header:

```
Authorization: Bearer <api-key>
```

---

## Service Status

### GET /status

Check the health status of all advanced services.

**Response:**
```json
{
  "healthy": true,
  "services": {
    "bid_landscape": true,
    "creative_optimization": true,
    "incrementality": true,
    "privacy_sandbox": true,
    "contextual_ai": true,
    "realtime_alerts": true,
    "competitive_intelligence": true,
    "unified_id": true,
    "dynamic_bid": true,
    "lookalike": true,
    "user_clustering": true
  }
}
```

---

## Bid Landscape Analysis

### POST /bid-landscape/analyze

Analyze the bid landscape for a campaign to determine optimal bid pricing.

**Request:**
```json
{
  "campaign_id": "camp-123",
  "publisher_id": "pub-456",
  "device_type": "mobile"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| campaign_id | string | ✅ | Campaign identifier |
| publisher_id | string | ❌ | Publisher to analyze |
| device_type | string | ❌ | Device type filter (mobile, desktop, tablet) |

**Response:**
```json
{
  "analyzed": true,
  "bid_multiplier": 1.15,
  "recommended_bid": 2.30,
  "win_probability": 0.72,
  "market_percentile": 65,
  "confidence": 0.85
}
```

### POST /bid-landscape/record

Record a bid outcome for landscape learning.

**Request:**
```json
{
  "publisher_id": "pub-456",
  "device_type": "mobile",
  "bid_price": 2.50,
  "win_price": 2.30,
  "won": true
}
```

**Response:**
```json
{
  "status": "recorded"
}
```

---

## Creative Optimization

### POST /creative/select

Select the optimal creative using multi-armed bandit algorithm.

**Request:**
```json
{
  "campaign_id": "camp-123",
  "formats": ["banner", "video"],
  "user_id": "user-789"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| campaign_id | string | ✅ | Campaign identifier |
| formats | array | ❌ | Acceptable ad formats |
| user_id | string | ❌ | User ID for personalization |

**Response:**
```json
{
  "selected_creative_id": "creative-456",
  "selection_method": "thompson_sampling",
  "confidence": 0.82,
  "expected_ctr": 0.025,
  "exploration": false
}
```

---

## Incrementality Testing

### POST /incrementality/evaluate

Evaluate if a user should be in the control or test group.

**Request:**
```json
{
  "campaign_id": "camp-123",
  "experiment_id": "exp-001",
  "user_id": "user-789",
  "control_percent": 10.0
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| campaign_id | string | ✅ | Campaign identifier |
| experiment_id | string | ❌ | Experiment identifier (auto-generated if empty) |
| user_id | string | ✅ | User to evaluate |
| control_percent | float | ❌ | Control group percentage (default: 10%) |

**Response:**
```json
{
  "experiment_id": "exp-001",
  "status": "running",
  "user_in_control_group": false
}
```

### GET /incrementality/results/:experiment_id

Get results for an incrementality experiment.

**Response:**
```json
{
  "experiment_id": "exp-001",
  "status": "running",
  "test_group_size": 5420,
  "control_group_size": 580,
  "test_conversion_rate": 0.032,
  "control_conversion_rate": 0.018,
  "incremental_lift": 0.778,
  "statistical_significance": 0.95,
  "incremental_conversions": 76,
  "incremental_revenue": 3800.00
}
```

### POST /incrementality/conversion

Record a conversion for incrementality tracking.

**Request:**
```json
{
  "experiment_id": "exp-001",
  "user_id": "user-789",
  "is_control": false,
  "revenue": 50.00
}
```

**Response:**
```json
{
  "status": "recorded"
}
```

---

## Privacy Sandbox

### POST /privacy/topic

Register a user's topic interest (Topics API).

**Request:**
```json
{
  "user_id": "user-789",
  "topic_id": 42
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| user_id | string | ✅ | User identifier |
| topic_id | int | ✅ | Chrome Topics API topic ID |

**Response:**
```json
{
  "status": "registered"
}
```

### POST /privacy/interest-group

Add a user to a FLEDGE/Protected Audience interest group.

**Request:**
```json
{
  "user_id": "user-789",
  "group_id": "sports_fans"
}
```

**Response:**
```json
{
  "status": "added"
}
```

### GET /privacy/interest-groups/:user_id

Get a user's interest groups.

**Response:**
```json
{
  "user_id": "user-789",
  "interest_groups": ["sports_fans", "tech_enthusiasts", "travel_lovers"]
}
```

---

## Contextual AI

### POST /contextual/analyze

Analyze page context for brand safety and targeting signals.

**Request:**
```json
{
  "campaign_id": "camp-123",
  "publisher_id": "pub-456",
  "brand_safety_level": "standard",
  "context": {
    "page_title": "Top 10 Travel Destinations for 2026",
    "page_content": "Explore the best vacation spots around the world...",
    "page_url": "https://example.com/travel/top-destinations",
    "page_category": "travel"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| campaign_id | string | ✅ | Campaign identifier |
| publisher_id | string | ❌ | Publisher identifier |
| brand_safety_level | string | ❌ | Safety level: strict, standard, relaxed |
| context | object | ❌ | Page context signals |

**Response:**
```json
{
  "analyzed": true,
  "brand_safe": true,
  "bid_multiplier": 1.20,
  "sentiment": "positive",
  "topics": ["travel", "lifestyle", "adventure"],
  "keywords": ["vacation", "destinations", "travel"],
  "confidence": 0.92,
  "reason": ""
}
```

---

## Real-Time Alerts

### POST /alerts/check

Check real-time alerts for a campaign.

**Request:**
```json
{
  "campaign_id": "camp-123",
  "current_spend": 850.00,
  "budget": 1000.00
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| campaign_id | string | ✅ | Campaign identifier |
| current_spend | float | ❌ | Current spend amount |
| budget | float | ❌ | Campaign budget |

**Response:**
```json
{
  "has_active_alerts": true,
  "alerts": [
    {
      "type": "budget_warning",
      "severity": "warning",
      "message": "Campaign at 85% of budget"
    }
  ],
  "bid_adjustment": 0.85
}
```

### POST /alerts/metrics

Record campaign metrics for anomaly detection.

**Request:**
```json
{
  "campaign_id": "camp-123",
  "spend": 100.00,
  "ctr": 0.025,
  "cvr": 0.012,
  "win_rate": 0.18
}
```

**Response:**
```json
{
  "status": "recorded"
}
```

---

## Competitive Intelligence

### POST /competitive/analyze

Analyze competitive landscape for bidding.

**Request:**
```json
{
  "campaign_id": "camp-123",
  "publisher_id": "pub-456",
  "ad_slot_id": "slot-789",
  "competitive_mode": "aggressive",
  "competitors": ["competitor-1", "competitor-2"]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| campaign_id | string | ✅ | Campaign identifier |
| publisher_id | string | ❌ | Publisher to analyze |
| ad_slot_id | string | ❌ | Specific ad slot |
| competitive_mode | string | ❌ | Mode: aggressive, balanced, conservative |
| competitors | array | ❌ | Competitors to track |

**Response:**
```json
{
  "analyzed": true,
  "bid_adjustment": 1.15,
  "market_share": 0.23,
  "competitor_activity": {
    "competitor-1": {
      "estimated_spend": 5000.00,
      "win_rate": 0.35,
      "avg_bid": 2.80
    }
  },
  "recommended_action": "increase_bid"
}
```

### POST /competitive/outcome

Record an auction outcome for competitive learning.

**Request:**
```json
{
  "publisher_id": "pub-456",
  "ad_slot_id": "slot-789",
  "bid_price": 2.50,
  "winning_price": 2.80,
  "won": false,
  "winner_id": "competitor-1"
}
```

**Response:**
```json
{
  "status": "recorded"
}
```

### GET /competitive/report

Get aggregated market intelligence report.

**Response:**
```json
{
  "total_auctions": 15420,
  "win_rate": 0.28,
  "avg_bid": 2.35,
  "avg_win_price": 2.15,
  "top_competitors": [
    {"id": "competitor-1", "win_rate": 0.35, "market_share": 0.22},
    {"id": "competitor-2", "win_rate": 0.18, "market_share": 0.12}
  ],
  "trend": "stable"
}
```

---

## Unified ID

### POST /identity/resolve

Resolve user identity across ID providers.

**Request:**
```json
{
  "campaign_id": "camp-123",
  "user_id": "user-789",
  "device_id": "device-abc",
  "providers": ["uid2", "id5", "liveramp"],
  "consent_required": true,
  "has_consent": true
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| campaign_id | string | ✅ | Campaign identifier |
| user_id | string | ❌ | User identifier |
| device_id | string | ❌ | Device identifier |
| providers | array | ❌ | ID providers to use |
| consent_required | bool | ❌ | Require GDPR consent |
| has_consent | bool | ❌ | User consent status |

**Response:**
```json
{
  "resolved": true,
  "unified_id": "unified-xyz-123",
  "provider": "uid2",
  "match_type": "deterministic",
  "confidence": 0.95,
  "bid_multiplier": 1.20,
  "has_consent": true,
  "cross_device_ids": ["device-abc", "device-def"]
}
```

### POST /identity/link

Link two identities together.

**Request:**
```json
{
  "id1": "uid2-abc",
  "provider1": "uid2",
  "id2": "id5-xyz",
  "provider2": "id5",
  "device_type": "mobile",
  "confidence": 0.90
}
```

**Response:**
```json
{
  "status": "linked"
}
```

### GET /identity/report

Get identity resolution statistics.

**Response:**
```json
{
  "total_identities": 125000,
  "total_links": 85000,
  "providers": {
    "uid2": {"identities": 80000, "match_rate": 0.72},
    "id5": {"identities": 65000, "match_rate": 0.58},
    "liveramp": {"identities": 45000, "match_rate": 0.42}
  },
  "cross_device_rate": 0.35
}
```

### GET /identity/cross-device-reach

Get cross-device reach metrics.

**Response:**
```json
{
  "cross_device_reach": 0.35
}
```

---

## Error Responses

All endpoints return standard error responses:

### 400 Bad Request
```json
{
  "error": "campaign_id is required"
}
```

### 503 Service Unavailable
```json
{
  "error": "Bid landscape service not available"
}
```

---

## Rate Limits

| Endpoint Type | Limit |
|---------------|-------|
| Analysis endpoints | 1000 req/min |
| Recording endpoints | 5000 req/min |
| Report endpoints | 100 req/min |

---

## Dynamic Bid Adjustments

ML-based real-time bid price optimization using Thompson Sampling and multi-factor analysis.

### POST /dynamic-bid/calculate

Calculate an optimized bid price based on contextual factors and historical performance.

**Request:**
```json
{
  "campaign_id": "camp-123",
  "publisher_id": "pub-456",
  "device_type": "mobile",
  "country": "US",
  "base_bid": 2.50,
  "user_id": "user-789",
  "ad_slot_width": 300,
  "ad_slot_height": 250
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| campaign_id | string | ✅ | Campaign identifier |
| publisher_id | string | ✅ | Publisher identifier |
| device_type | string | ❌ | Device type (mobile, desktop, tablet) |
| country | string | ❌ | ISO country code |
| base_bid | float | ✅ | Starting bid price in CPM |
| user_id | string | ❌ | User identifier for personalization |
| ad_slot_width | int | ❌ | Ad slot width in pixels |
| ad_slot_height | int | ❌ | Ad slot height in pixels |

**Response:**
```json
{
  "campaign_id": "camp-123",
  "original_bid": 2.50,
  "adjusted_bid": 2.85,
  "adjustments": {
    "device_multiplier": 1.05,
    "geo_multiplier": 1.10,
    "time_multiplier": 0.95,
    "publisher_multiplier": 1.08
  },
  "confidence": 0.82,
  "win_probability": 0.68,
  "reason": "high_quality_publisher"
}
```

### POST /dynamic-bid/outcome

Record a bid outcome for ML model learning and adjustment.

**Request:**
```json
{
  "campaign_id": "camp-123",
  "publisher_id": "pub-456",
  "user_id": "user-789",
  "bid_price": 2.85,
  "won": true,
  "win_price": 2.60,
  "clicked": true,
  "converted": false,
  "revenue": 0.0
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| campaign_id | string | ✅ | Campaign identifier |
| publisher_id | string | ✅ | Publisher identifier |
| user_id | string | ❌ | User identifier |
| bid_price | float | ✅ | Bid price submitted |
| won | bool | ❌ | Whether the bid was won |
| win_price | float | ❌ | Clearing price if won |
| clicked | bool | ❌ | Whether ad was clicked |
| converted | bool | ❌ | Whether conversion occurred |
| revenue | float | ❌ | Revenue from conversion |

**Response:**
```json
{
  "status": "recorded"
}
```

### GET /dynamic-bid/analytics

Get bid performance analytics across all campaigns.

**Response:**
```json
{
  "total_bids": 150000,
  "total_wins": 72000,
  "win_rate": 0.48,
  "avg_bid_price": 2.65,
  "avg_win_price": 2.45,
  "efficiency": 0.92,
  "campaigns": 25,
  "top_performing_hours": [10, 11, 14, 15, 20]
}
```

### GET /dynamic-bid/config

Get the current dynamic bid service configuration.

**Response:**
```json
{
  "enabled": true,
  "min_bid": 0.10,
  "max_bid": 50.00,
  "max_multiplier": 3.0,
  "learning_rate": 0.1,
  "exploration_rate": 0.05
}
```

---

## Lookalike Audiences

Generate lookalike audiences by finding users similar to a seed audience.

### POST /lookalike/generate

Generate a lookalike audience from seed users.

**Request:**
```json
{
  "seed_user_ids": ["user-1", "user-2", "user-3", "user-4", "user-5"],
  "name": "High Value Customers Lookalike",
  "expansion_factor": 2.0
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| seed_user_ids | []string | ✅ | List of seed user IDs (min 5) |
| name | string | ✅ | Name for the lookalike audience |
| expansion_factor | float | ❌ | Audience expansion multiplier (default: 2.0) |

**Response:**
```json
{
  "audience_id": "lal-abc123",
  "status": "completed",
  "seed_size": 5,
  "audience_size": 45,
  "expansion_ratio": 9.0,
  "quality_score": 0.78,
  "created_at": "2026-02-20T10:30:00Z",
  "expires_at": "2026-02-27T10:30:00Z"
}
```

### POST /lookalike/user-profile

Register a user profile for lookalike modeling.

**Request:**
```json
{
  "user_id": "user-123",
  "segments": ["sports", "tech", "travel"],
  "interests": ["basketball", "programming", "hiking"],
  "device_types": ["mobile", "desktop"],
  "country": "US",
  "region": "CA",
  "city": "San Francisco"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| user_id | string | ✅ | User identifier |
| segments | []string | ❌ | User segment memberships |
| interests | []string | ❌ | User interests |
| device_types | []string | ❌ | User's device types |
| country | string | ❌ | ISO country code |
| region | string | ❌ | State/region code |
| city | string | ❌ | City name |

**Response:**
```json
{
  "status": "registered"
}
```

### GET /lookalike/audience/:audience_id

Retrieve a lookalike audience by ID.

**Response:**
```json
{
  "ID": "lal-abc123",
  "Name": "High Value Customers Lookalike",
  "SeedSegmentID": "seg-001",
  "SeedSize": 100,
  "AudienceSize": 850,
  "ExpansionRatio": 8.5,
  "Users": [...],
  "CreatedAt": "2026-02-20T10:30:00Z",
  "ExpiresAt": "2026-02-27T10:30:00Z",
  "QualityScore": 0.82
}
```

### GET /lookalike/check

Check if a user is in a lookalike audience.

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| user_id | string | ✅ | User identifier to check |
| audience_id | string | ✅ | Lookalike audience ID |

**Response:**
```json
{
  "is_member": true,
  "similarity_score": 0.85
}
```

### GET /lookalike/stats

Get statistics about all lookalike audiences.

**Response:**
```json
{
  "total_audiences": 12,
  "total_users": 45000,
  "avg_quality_score": 0.76,
  "avg_expansion_ratio": 6.5,
  "active_audiences": 10
}
```

---

## User Clustering

K-means based user segmentation for audience analysis and targeting.

### POST /clustering/user

Register a user with feature vectors for clustering.

**Request:**
```json
{
  "user_id": "user-123",
  "feature_vector": [0.5, 0.6, 0.7, 0.8, 0.9, ...]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| user_id | string | ✅ | User identifier |
| feature_vector | []float | ✅ | 30-dimensional feature vector |

**Note:** Feature vector should be 30 dimensions representing:
- Indices 0-9: Interest scores (sports, tech, finance, etc.)
- Index 15: Engagement score
- Index 16: Recency score
- Indices 20-24: Device preference scores

**Response:**
```json
{
  "status": "registered"
}
```

### POST /clustering/run

Trigger the K-means clustering algorithm on registered users.

**Response:**
```json
{
  "status": "completed",
  "total_users": 10000,
  "clusters_created": 10,
  "iterations": 45,
  "converged": true,
  "silhouette_score": 0.68,
  "clusters": [
    {
      "id": "cluster-1",
      "name": "High Engagement Mobile Users",
      "user_count": 1250,
      "cohesion": 0.82,
      "top_interests": ["sports", "tech"],
      "behavior_profile": "high_engagement",
      "value_segment": "premium"
    }
  ]
}
```

### GET /clustering/user/:user_id

Get the cluster assignment for a specific user.

**Response:**
```json
{
  "cluster": {
    "ID": "cluster-3",
    "Name": "Tech Enthusiasts",
    "Centroid": [...],
    "UserCount": 1580,
    "Cohesion": 0.75
  },
  "confidence": 0.88
}
```

### GET /clustering/cluster/:cluster_id/users

Get all users in a specific cluster.

**Response:**
```json
{
  "cluster_id": "cluster-3",
  "user_count": 1580,
  "users": ["user-123", "user-456", "user-789", ...]
}
```

### GET /clustering/stats

Get overall clustering statistics.

**Response:**
```json
{
  "total_users": 50000,
  "total_clusters": 10,
  "avg_cluster_size": 5000,
  "avg_cohesion": 0.72,
  "last_run": "2026-02-20T08:00:00Z",
  "cluster_distribution": {
    "cluster-1": 5200,
    "cluster-2": 4800,
    "cluster-3": 5100
  }
}
```

---

## SDKs

Official SDKs available for:
- Go: `go get github.com/taskirx/go-bidding-sdk`
- Python: `pip install taskirx-bidding`
- Node.js: `npm install @taskirx/bidding-sdk`

---

## Changelog

### v1.1.0 (2026-02-20)
- Added Dynamic Bid Adjustment service (4 endpoints)
- Added Lookalike Audience service (5 endpoints)
- Added User Clustering service (5 endpoints)
- 14 new ML-powered endpoints
- Total: 36 API endpoints

### v1.0.0 (2026-02-20)
- Initial release with 8 advanced services
- 21 API endpoints
- Full test coverage
