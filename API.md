# TaskirX - API Reference

## Base URL
```
http://localhost:3000/api
```

## Authentication

All authenticated endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

---

## Authentication Endpoints

### Register User
```http
POST /api/auth/register
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "Password123!",
  "name": "John Doe",
  "company": "Acme Corp"
}
```

**Response:** `201 Created`
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "507f1f77bcf86cd799439011",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "advertiser"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Login
```http
POST /api/auth/login
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "Password123!"
}
```

**Response:** `200 OK`
```json
{
  "message": "Login successful",
  "user": {
    "id": "507f1f77bcf86cd799439011",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "advertiser"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

## Campaign Endpoints

### List Campaigns
```http
GET /api/campaigns?page=1&limit=20&status=active&search=summer
```

**Query Parameters:**
- `page` (number): Page number (default: 1)
- `limit` (number): Items per page (default: 20)
- `status` (string): Filter by status (active, paused, completed, deleted)
- `search` (string): Search by campaign name

**Response:** `200 OK`
```json
{
  "campaigns": [
    {
      "_id": "507f1f77bcf86cd799439011",
      "name": "Summer Sale Campaign",
      "status": "active",
      "budget": {
        "total": 10000,
        "spent": 2500
      },
      "performance": {
        "impressions": 50000,
        "clicks": 1250,
        "conversions": 125
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "pages": 3
  }
}
```

### Create Campaign
```http
POST /api/campaigns
```

**Request Body:**
```json
{
  "name": "Summer Sale Campaign",
  "status": "active",
  "bidding": {
    "strategy": "cpm",
    "maxBid": 5.0
  },
  "targeting": {
    "geo": {
      "countries": ["US", "CA"],
      "cities": ["New York", "Los Angeles"]
    },
    "device": {
      "types": ["mobile", "desktop"]
    },
    "demographics": {
      "ageRange": { "min": 25, "max": 45 }
    }
  },
  "creative": {
    "sizes": ["300x250", "728x90"],
    "assets": [
      {
        "type": "image",
        "url": "https://example.com/banner.jpg"
      }
    ]
  },
  "budget": {
    "total": 10000,
    "daily": 500
  },
  "schedule": {
    "startDate": "2024-01-01T00:00:00Z",
    "endDate": "2024-01-31T23:59:59Z"
  }
}
```

**Response:** `201 Created`

---

## RTB Endpoints

### Bid Request (OpenRTB 2.5)
```http
POST /api/rtb/bid-request
```

**Request Body:**
```json
{
  "id": "request-123",
  "imp": [
    {
      "id": "1",
      "bidfloor": 1.0,
      "banner": {
        "w": 300,
        "h": 250,
        "pos": 0
      }
    }
  ],
  "device": {
    "ip": "192.168.1.1",
    "ua": "Mozilla/5.0...",
    "devicetype": 1,
    "geo": {
      "country": "US",
      "city": "New York"
    }
  },
  "site": {
    "domain": "example.com",
    "page": "https://example.com/article"
  }
}
```

**Response:** `200 OK` (Bid) or `204 No Content` (No bid)
```json
{
  "id": "request-123",
  "seatbid": [
    {
      "bid": [
        {
          "id": "bid-456",
          "impid": "1",
          "price": 2.5,
          "adid": "campaign-789",
          "nurl": "http://localhost:3000/api/rtb/win?id=bid-456",
          "adm": "<a href='...'><img src='...' /></a>",
          "w": 300,
          "h": 250
        }
      ]
    }
  ],
  "cur": "USD"
}
```

### Win Notification
```http
GET /api/rtb/win?id=bid-456&price=2.5
```

**Response:** `200 OK`

### Click Tracking
```http
GET /api/rtb/click/:bidId
```

**Response:** `302 Found` (Redirect to landing page)

---

## Analytics Endpoints

### Dashboard Stats
```http
GET /api/analytics/dashboard?startDate=2024-01-01&endDate=2024-01-31
```

**Response:** `200 OK`
```json
{
  "overview": {
    "totalImpressions": 500000,
    "totalClicks": 12500,
    "totalConversions": 1250,
    "totalSpent": 25000,
    "ctr": 2.5,
    "cvr": 10.0,
    "avgCpc": 2.0,
    "avgCpm": 50.0
  },
  "activeCampaigns": 15,
  "totalCampaigns": 45
}
```

### Campaign Performance
```http
GET /api/analytics/campaigns/:campaignId/performance?groupBy=day
```

**Response:** `200 OK`
```json
{
  "campaign": {
    "id": "507f1f77bcf86cd799439011",
    "name": "Summer Sale Campaign"
  },
  "performance": [
    {
      "date": "2024-01-01",
      "impressions": 5000,
      "clicks": 125,
      "conversions": 12,
      "spent": 250,
      "ctr": 2.5,
      "cvr": 9.6
    }
  ]
}
```

### Time Series Data
```http
GET /api/analytics/timeseries?metric=impressions&groupBy=hour&startDate=2024-01-01
```

**Metrics:** impressions, clicks, conversions, spent, revenue
**GroupBy:** hour, day, week, month

**Response:** `200 OK`
```json
[
  { "timestamp": "2024-01-01T00:00:00Z", "value": 5000 },
  { "timestamp": "2024-01-01T01:00:00Z", "value": 5200 }
]
```

### Geo Performance
```http
GET /api/analytics/geo-performance?level=country
```

**Levels:** country, region, city

**Response:** `200 OK`
```json
[
  {
    "country": "US",
    "impressions": 250000,
    "clicks": 6250,
    "conversions": 625,
    "ctr": 2.5,
    "cvr": 10.0
  }
]
```

### Funnel Analysis
```http
GET /api/analytics/funnel?startDate=2024-01-01&endDate=2024-01-31
```

**Response:** `200 OK`
```json
{
  "bidRequests": 1000000,
  "bidsWon": 500000,
  "impressions": 450000,
  "clicks": 11250,
  "conversions": 1125,
  "winRate": 50.0,
  "serveRate": 90.0,
  "ctr": 2.5,
  "cvr": 10.0
}
```

---

## Health & Monitoring

### Health Check
```http
GET /health
```

**Response:** `200 OK`
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "uptime": 86400,
  "environment": "production",
  "database": "connected",
  "redis": "connected"
}
```

### Metrics (Prometheus)
```http
GET /metrics
```

**Response:** `200 OK` (text/plain)
```
# HELP rtb_bids_won_total Total number of bids won
# TYPE rtb_bids_won_total counter
rtb_bids_won_total 125000

# HELP rtb_latency_ms RTB request latency in milliseconds
# TYPE rtb_latency_ms histogram
rtb_latency_ms_bucket{le="50"} 95000
rtb_latency_ms_bucket{le="100"} 124500
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Validation Error",
  "message": "Invalid request data",
  "details": [
    {
      "field": "email",
      "message": "Valid email is required"
    }
  ]
}
```

### 401 Unauthorized
```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing authentication token"
}
```

### 403 Forbidden
```json
{
  "error": "Forbidden",
  "message": "Insufficient permissions"
}
```

### 404 Not Found
```json
{
  "error": "Not Found",
  "message": "Campaign not found"
}
```

### 429 Too Many Requests
```json
{
  "error": "Too Many Requests",
  "message": "Rate limit exceeded. Please try again later."
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal Server Error",
  "message": "An unexpected error occurred"
}
```

---

## Rate Limits

- **Standard Endpoints**: 100 requests per 15 minutes
- **RTB Endpoints**: 1000 requests per 15 minutes
- **Authentication**: 10 requests per 15 minutes

Rate limit headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248000
```

---

## Postman Collection

Import the `postman-collection.json` file for a complete collection of API requests with examples.
