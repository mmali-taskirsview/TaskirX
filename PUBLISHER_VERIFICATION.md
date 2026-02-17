# Publisher Verification Guide

## 1. Publisher Login
To verify the **Publisher Dashboard** access:

- **URL**: `https://dashboard.taskirx.com` (or `http://localhost:3001` for local)
- **Email**: `publisher@test.com`
- **Password**: `Admin123!`

### Curl Verification
```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "publisher@test.com", "password": "Admin123!"}'
```

## 2. Bid Adapter Simulation (Ad Request)
The Publisher SDK uses the `taskirxBidAdapter.js` to request bids from the Go Engine.

### Test Payload
Create a file `test-publisher-payload.json` with the following content:
```json
{
    "id": "req-12345",
    "timestamp": "2026-02-17T12:00:00Z",
    "publisher_id": "pub-001",
    "ad_slot": {
        "id": "slot-001",
        "dimensions": [300, 250],
        "position": "above-fold",
        "formats": ["banner"]
    },
    "user": {
        "id": "user-001",
        "country": "US",
        "language": "en"
    },
    "device": {
        "type": "desktop", 
        "os": "windows",
        "browser": "chrome",
        "ip": "203.0.113.1",
        "geo": {
            "lat": 40.7128,
            "lon": -74.0060
        }
    },
    "context": {
        "page_url": "https://example-publisher.com/article"
    }
}
```

### Verification Command
Run this command against the Go Bidding Engine:
```bash
curl -v -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d @test-publisher-payload.json
```

**Expected Success Response (200 OK):**
```json
{
  "request_id": "req-12345",
  "bid_price": 2.0625,
  "creative_url": "...",
  "impression_url": "...",
  "click_url": "..."
}
```

## 3. Browser Simulation
Open `publisher-demo.html` in a browser. 
**Note**: You may need to update the `ENDPOINT_URL` in `sdks/javascript/taskirxBidAdapter.js` to point to the live server if testing production (`https://bidding.taskirx.com/bid` instead of `localhost`).
