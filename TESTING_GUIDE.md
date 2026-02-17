# TaskirX - Testing Guide

Complete guide for testing TaskirX before production deployment.

---

## Table of Contents

1. [Local Testing](#local-testing)
2. [API Testing](#api-testing)
3. [Performance Testing](#performance-testing)
4. [Security Testing](#security-testing)
5. [Mobile SDK Testing](#mobile-sdk-testing)
6. [Integration Testing](#integration-testing)
7. [Automated Testing](#automated-testing)
8. [OCI (Production) Testing](#oci-production-testing)

---

## Local Testing

### 1. Server Health Check

**Quick Test**:
```bash
npm run health
```

**Expected Output**:
```
[2025-11-14 15:30:47] ✓ Service is healthy (Response time: 45ms)
Health check completed successfully
```

**Manual Test**:
```bash
# PowerShell
Invoke-WebRequest -Uri "http://localhost:3000/health" -Method GET

# Expected Status Code: 200
# Expected Content: {"status":"healthy",...}
```

### 2. Pre-Launch Verification

**Run Comprehensive Checks**:
```bash
npm run verify
```

**Verification Categories**:
1. **Environment**: All required variables set
2. **Files**: Critical files exist (server.js, models, routes, config)
3. **Dependencies**: 670 packages installed
4. **Server Health**: /health endpoint responds 200
5. **API Endpoints**: All 8 route groups functional
6. **Documentation**: 12 guides complete
7. **Security**: JWT_SECRET changed, rate limiting active
8. **Scripts**: 7 operational tools ready

**Success Criteria**:
- ✅ All checks passed: Ready for production
- ⚠️ Warnings only: Review and fix recommended
- ❌ Failures: Must fix before deployment

### 3. Database Connection

**Test MongoDB**:
```bash
# PowerShell - Check MongoDB connection
$env:MONGODB_URI = "mongodb://localhost:27017/taskirx"
cd backend
node -e "const mongoose = require('mongoose'); mongoose.connect(process.env.MONGODB_URI).then(() => { console.log('✓ MongoDB Connected'); process.exit(0); }).catch(err => { console.error('✗ MongoDB Error:', err.message); process.exit(1); });"
```

**Expected Output**:
```
✓ MongoDB Connected
```

### 4. Log Files

**Check Logging**:
```bash
# View combined logs
Get-Content backend/logs/combined.log -Tail 20

# View error logs (should be empty or minimal)
Get-Content backend/logs/error.log -Tail 20

# Monitor logs in real-time
Get-Content backend/logs/combined.log -Wait
```

**Expected**:
- Logs directory exists: `backend/logs/`
- Combined log has startup messages
- Error log is empty (no critical errors)
- Log rotation working (max 5 files × 5MB)

---

## API Testing

### 1. Authentication Flow

**Register New User**:
```powershell
$headers = @{'Content-Type'='application/json'}
$body = @{
    email = "test@example.com"
    password = "Test123!"
    name = "Test User"
    role = "advertiser"
} | ConvertTo-Json

$response = Invoke-WebRequest -Uri "http://localhost:3000/api/auth/register" -Method POST -Headers $headers -Body $body
$result = $response.Content | ConvertFrom-Json
$token = $result.token
Write-Host "Token: $token"
```

**Expected Response** (200):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "673...",
    "email": "test@example.com",
    "name": "Test User",
    "role": "advertiser"
  }
}
```

**Login**:
```powershell
$body = @{
    email = "test@example.com"
    password = "Test123!"
} | ConvertTo-Json

$response = Invoke-WebRequest -Uri "http://localhost:3000/api/auth/login" -Method POST -Headers $headers -Body $body
$result = $response.Content | ConvertFrom-Json
$token = $result.token
```

**Verify Token**:
```powershell
$headers = @{
    'Authorization' = "Bearer $token"
}
Invoke-WebRequest -Uri "http://localhost:3000/api/auth/me" -Method GET -Headers $headers
```

**Test Cases**:
- ✅ Register with valid data → 201 Created
- ✅ Register with duplicate email → 400 Bad Request
- ✅ Login with correct credentials → 200 OK
- ✅ Login with wrong password → 401 Unauthorized
- ✅ Access protected route with valid token → 200 OK
- ✅ Access protected route without token → 401 Unauthorized
- ✅ Access protected route with expired token → 401 Unauthorized

### 2. Campaign Management

**Create Campaign**:
```powershell
$headers = @{
    'Authorization' = "Bearer $token"
    'Content-Type' = 'application/json'
}

$campaign = @{
    name = "Test Campaign"
    advertiserId = $result.user.id
    budget = 5000
    bidAmount = 2.0
    targeting = @{
        countries = @("US", "UK")
        deviceTypes = @("mobile", "desktop")
        ageRange = @{min = 18; max = 65}
    }
    startDate = "2025-11-15T00:00:00Z"
    endDate = "2025-12-31T23:59:59Z"
} | ConvertTo-Json -Depth 10

$response = Invoke-WebRequest -Uri "http://localhost:3000/api/campaigns" -Method POST -Headers $headers -Body $campaign
$campaignData = $response.Content | ConvertFrom-Json
$campaignId = $campaignData._id
```

**List Campaigns**:
```powershell
$response = Invoke-WebRequest -Uri "http://localhost:3000/api/campaigns" -Method GET -Headers $headers
$campaigns = $response.Content | ConvertFrom-Json
Write-Host "Total Campaigns: $($campaigns.Count)"
```

**Update Campaign**:
```powershell
$update = @{
    budget = 7500
    bidAmount = 2.5
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:3000/api/campaigns/$campaignId" -Method PUT -Headers $headers -Body $update
```

**Delete Campaign**:
```powershell
Invoke-WebRequest -Uri "http://localhost:3000/api/campaigns/$campaignId" -Method DELETE -Headers $headers
```

**Test Cases**:
- ✅ Create campaign with valid data → 201 Created
- ✅ Create campaign without authentication → 401 Unauthorized
- ✅ List campaigns with pagination → 200 OK
- ✅ Get single campaign by ID → 200 OK
- ✅ Update campaign budget → 200 OK
- ✅ Delete campaign → 204 No Content
- ✅ Get deleted campaign → 404 Not Found

### 3. RTB (Real-Time Bidding)

**Send Bid Request**:
```powershell
$bidRequest = @{
    id = "req-" + (Get-Random)
    imp = @(
        @{
            id = "imp-1"
            banner = @{
                w = 300
                h = 250
                pos = 1
            }
            bidfloor = 0.5
            bidfloorcur = "USD"
        }
    )
    site = @{
        id = "site-123"
        domain = "example.com"
        cat = @("IAB1")
        page = "https://example.com/page"
        publisher = @{
            id = "pub-456"
            name = "Example Publisher"
        }
    }
    device = @{
        ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
        ip = "203.0.113.1"
        geo = @{
            country = "USA"
            city = "New York"
            lat = 40.7128
            lon = -74.0060
        }
        devicetype = 1
        os = "Windows"
    }
    user = @{
        id = "user-789"
    }
    at = 2
    tmax = 100
} | ConvertTo-Json -Depth 10

$response = Invoke-WebRequest -Uri "http://localhost:3000/api/rtb/bid-request" -Method POST -Headers @{'Content-Type'='application/json'} -Body $bidRequest
$bidResponse = $response.Content | ConvertFrom-Json
```

**Expected Response** (200):
```json
{
  "id": "req-12345",
  "seatbid": [
    {
      "bid": [
        {
          "id": "bid-67890",
          "impid": "imp-1",
          "price": 2.5,
          "adid": "ad-123",
          "nurl": "http://example.com/win?price=${AUCTION_PRICE}",
          "adm": "<html>...</html>",
          "adomain": ["advertiser.com"],
          "cid": "camp-123",
          "crid": "creative-456"
        }
      ],
      "seat": "advertiser-123"
    }
  ],
  "bidid": "bid-response-123",
  "cur": "USD"
}
```

**Test Cases**:
- ✅ Valid bid request → 200 OK with bid response
- ✅ No matching campaigns → 204 No Content
- ✅ Invalid bid request format → 400 Bad Request
- ✅ Response time < 100ms → Performance OK

### 4. MMP Integration

**Track Install Event**:
```powershell
$event = @{
    eventType = "install"
    provider = "appsflyer"
    campaignId = $campaignId
    userId = "user-123"
    deviceId = "device-456"
    timestamp = (Get-Date).ToUniversalTime().ToString("o")
    metadata = @{
        app_id = "com.example.app"
        platform = "android"
        os_version = "13.0"
        device_model = "Pixel 7"
    }
} | ConvertTo-Json -Depth 10

Invoke-WebRequest -Uri "http://localhost:3000/api/mmp/events/track" -Method POST -Headers $headers -Body $event
```

**Track In-App Event**:
```powershell
$purchaseEvent = @{
    eventType = "purchase"
    provider = "adjust"
    campaignId = $campaignId
    userId = "user-123"
    timestamp = (Get-Date).ToUniversalTime().ToString("o")
    revenue = 9.99
    currency = "USD"
    metadata = @{
        item_id = "premium_plan"
        item_name = "Premium Subscription"
    }
} | ConvertTo-Json -Depth 10

Invoke-WebRequest -Uri "http://localhost:3000/api/mmp/events/track" -Method POST -Headers $headers -Body $event
```

**Get Campaign Attribution Stats**:
```powershell
$response = Invoke-WebRequest -Uri "http://localhost:3000/api/mmp/events/$campaignId/stats" -Method GET -Headers $headers
$stats = $response.Content | ConvertFrom-Json
Write-Host "Installs: $($stats.installs)"
Write-Host "Conversions: $($stats.conversions)"
Write-Host "Revenue: $($stats.totalRevenue)"
```

**Test Cases**:
- ✅ Track install event → 201 Created
- ✅ Track purchase event → 201 Created
- ✅ Track event without authentication → 401 Unauthorized
- ✅ Get campaign stats → 200 OK with metrics
- ✅ Receive MMP postback → 200 OK

### 5. GDPR/CCPA Compliance

**Record Consent**:
```powershell
$consent = @{
    userId = "user-123"
    consentGiven = $true
    purposes = @("advertising", "analytics", "personalization")
    ipAddress = "203.0.113.1"
    userAgent = "Mozilla/5.0..."
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:3000/api/consent" -Method POST -Headers @{'Content-Type'='application/json'} -Body $consent
```

**Get Consent Status**:
```powershell
Invoke-WebRequest -Uri "http://localhost:3000/api/consent/user-123" -Method GET
```

**Export User Data**:
```powershell
$response = Invoke-WebRequest -Uri "http://localhost:3000/api/consent/user-123/export" -Method GET
$userData = $response.Content | ConvertFrom-Json
$userData | ConvertTo-Json -Depth 10 | Out-File "user_data_export.json"
```

**Delete User Data**:
```powershell
Invoke-WebRequest -Uri "http://localhost:3000/api/consent/user-123/data" -Method DELETE -Headers $headers
```

**Test Cases**:
- ✅ Record consent → 201 Created
- ✅ Update consent → 200 OK
- ✅ Get consent status → 200 OK
- ✅ Export user data → 200 OK with JSON export
- ✅ Delete user data → 200 OK with deletion summary

---

## Performance Testing

### 1. Load Testing

**Run Full Load Test**:
```bash
npm run load-test
```

**Custom Load Test**:
```bash
# Test specific endpoint
node scripts/load-test.js --endpoint http://localhost:3000/api/campaigns --connections 200 --duration 60

# High load test
node scripts/load-test.js --connections 500 --duration 120
```

**Performance Targets**:

| Endpoint | QPS Target | P95 Latency | Current Performance |
|----------|-----------|-------------|---------------------|
| Health | >10,000 | <50ms | ~10,245 QPS, 12ms |
| RTB | >1,000 | <100ms | ~1,523 QPS, 85ms |
| Campaigns | >500 | <200ms | ~580 QPS, 120ms |
| MMP Events | >1,000 | <100ms | TBD |
| Analytics | >200 | <500ms | TBD |

**Success Criteria**:
- ✅ All targets met or exceeded
- ✅ Error rate < 0.1%
- ✅ No timeouts under load
- ✅ Memory usage stable (<512MB)

### 2. Stress Testing

**Gradual Load Increase**:
```bash
# Start with 10 connections
node scripts/load-test.js --connections 10 --duration 30

# Increase to 50
node scripts/load-test.js --connections 50 --duration 30

# Increase to 100
node scripts/load-test.js --connections 100 --duration 30

# Increase to 200
node scripts/load-test.js --connections 200 --duration 30

# Find breaking point
node scripts/load-test.js --connections 500 --duration 60
```

**Monitor During Test**:
```bash
# Terminal 1: Run load test
npm run load-test

# Terminal 2: Monitor health
npm run monitor

# Terminal 3: Monitor logs
Get-Content backend/logs/combined.log -Wait

# Terminal 4: Monitor memory
while ($true) { Get-Process -Name node | Select-Object @{Name='Memory(MB)';Expression={[math]::Round($_.WS/1MB,2)}} | Format-Table; Start-Sleep -Seconds 5 }
```

### 3. Database Performance

**Test Query Performance**:
```bash
# Seed database with test data
npm run seed

# Run benchmark
npm run benchmark
```

**Database Indexes**:
```javascript
// Verify indexes exist
mongo
> use taskirx
> db.campaigns.getIndexes()
> db.bids.getIndexes()
> db.impressions.getIndexes()
> db.mmpeventtracks.getIndexes()
```

**Expected Indexes**:
- Campaigns: advertiserId, status, startDate+endDate
- Bids: campaignId, status, createdAt
- Impressions: campaignId, timestamp, userId
- MMP Events: campaignId, userId, timestamp

---

## Security Testing

### 1. Authentication Security

**Test JWT Expiration**:
```powershell
# Get token
$response = Invoke-WebRequest -Uri "http://localhost:3000/api/auth/login" -Method POST -Headers @{'Content-Type'='application/json'} -Body '{"email":"test@example.com","password":"Test123!"}'
$token = ($response.Content | ConvertFrom-Json).token

# Use token immediately (should work)
Invoke-WebRequest -Uri "http://localhost:3000/api/auth/me" -Method GET -Headers @{'Authorization'="Bearer $token"}

# Wait for expiration (default 24h) or modify JWT_EXPIRES_IN in .env
# Try using expired token (should fail with 401)
```

**Test Invalid Tokens**:
```powershell
# No token
Invoke-WebRequest -Uri "http://localhost:3000/api/campaigns" -Method GET
# Expected: 401 Unauthorized

# Invalid token
Invoke-WebRequest -Uri "http://localhost:3000/api/campaigns" -Method GET -Headers @{'Authorization'="Bearer invalid_token"}
# Expected: 401 Unauthorized

# Malformed token
Invoke-WebRequest -Uri "http://localhost:3000/api/campaigns" -Method GET -Headers @{'Authorization'="Bearer eyJhbGciOiJIUzI1NiIsInR5cCI"}
# Expected: 401 Unauthorized
```

### 2. Rate Limiting

**Test API Rate Limit** (100 requests per 15 minutes):
```powershell
# Send 110 requests rapidly
1..110 | ForEach-Object {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:3000/health" -Method GET
        Write-Host "Request $_: $($response.StatusCode)"
    } catch {
        Write-Host "Request $_: Rate Limited (429)"
    }
}
# Expected: First 100 succeed, next 10 fail with 429 Too Many Requests
```

**Test Auth Rate Limit** (5 requests per 15 minutes):
```powershell
# Send 7 login attempts
1..7 | ForEach-Object {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:3000/api/auth/login" -Method POST -Headers @{'Content-Type'='application/json'} -Body '{"email":"test@example.com","password":"wrong"}'
        Write-Host "Attempt $_: $($response.StatusCode)"
    } catch {
        Write-Host "Attempt $_: Rate Limited"
    }
}
# Expected: First 5 fail with 401, next 2 fail with 429
```

### 3. Input Validation

**Test SQL Injection Attempts**:
```powershell
# Try SQL injection in email field
$body = @{
    email = "'; DROP TABLE users; --"
    password = "test"
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:3000/api/auth/login" -Method POST -Headers @{'Content-Type'='application/json'} -Body $body
# Expected: 400 Bad Request (validation error)
```

**Test XSS Attempts**:
```powershell
# Try XSS in campaign name
$body = @{
    name = "<script>alert('xss')</script>"
    advertiserId = "123"
    budget = 1000
    bidAmount = 1
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:3000/api/campaigns" -Method POST -Headers @{'Authorization'="Bearer $token"; 'Content-Type'='application/json'} -Body $body
# Expected: Script tags sanitized or rejected
```

### 4. CORS Testing

**Test CORS Headers**:
```powershell
# Request from different origin
$response = Invoke-WebRequest -Uri "http://localhost:3000/health" -Method GET -Headers @{'Origin'='https://example.com'}
$response.Headers['Access-Control-Allow-Origin']
# Expected: * (development) or specific origin (production)
```

### 5. Security Headers

**Verify Security Headers**:
```powershell
$response = Invoke-WebRequest -Uri "http://localhost:3000/health" -Method GET
$response.Headers | Format-Table

# Expected headers:
# - X-Content-Type-Options: nosniff
# - X-Frame-Options: DENY
# - X-XSS-Protection: 1; mode=block
# - Strict-Transport-Security: max-age=31536000 (production only)
```

---

## Mobile SDK Testing

### JavaScript SDK

**Build SDK**:
```bash
cd sdks/javascript
npm install
npm run build
npm test
```

**Integration Test**:
```html
<!DOCTYPE html>
<html>
<head>
    <title>SDK Test</title>
    <script src="dist/adx-sdk.min.js"></script>
</head>
<body>
    <script>
        const sdk = new taskirx({
            apiKey: 'test-api-key',
            endpoint: 'http://localhost:3000'
        });
        
        // Test ad request
        sdk.requestAd({
            placement: 'banner-300x250',
            targeting: { country: 'US' }
        }).then(ad => {
            console.log('Ad received:', ad);
        }).catch(err => {
            console.error('Error:', err);
        });
        
        // Test impression tracking
        sdk.trackImpression({
            campaignId: 'camp-123',
            placementId: 'placement-456'
        });
    </script>
</body>
</html>
```

**Test Cases**:
- ✅ SDK loads without errors
- ✅ Ad request returns valid response
- ✅ Impression tracking sends request
- ✅ Click tracking works
- ✅ Error handling functional
- ✅ Minified bundle size < 10KB

### Android SDK

**Build SDK**:
```bash
cd sdks/android
./gradlew clean build
./gradlew test
```

**Integration Test**:
```kotlin
// In your test Activity/Fragment
class SDKTestActivity : AppCompatActivity() {
    private lateinit var sdk: taskirx
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        sdk = taskirx.Builder(this)
            .setApiKey("test-api-key")
            .setEndpoint("http://10.0.2.2:3000") // Android emulator localhost
            .build()
        
        // Test ad request
        lifecycleScope.launch {
            try {
                val ad = sdk.requestAd(
                    placement = "banner-300x250",
                    targeting = mapOf("country" to "US")
                )
                Log.d("SDK", "Ad received: $ad")
            } catch (e: Exception) {
                Log.e("SDK", "Error: ${e.message}")
            }
        }
        
        // Test impression tracking
        sdk.trackImpression(
            campaignId = "camp-123",
            placementId = "placement-456"
        )
    }
}
```

**Test Cases**:
- ✅ SDK builds without errors
- ✅ Ad request on Android 5.0+ (API 21+)
- ✅ Impression tracking works
- ✅ Network requests use HTTPS in production
- ✅ Permissions handled correctly
- ✅ Proguard rules work

### iOS SDK

**Build SDK**:
```bash
cd sdks/ios
swift build
swift test
```

**Integration Test**:
```swift
import taskirxSDK

class SDKTestViewController: UIViewController {
    var sdk: taskirx!
    
    override func viewDidLoad() {
        super.viewDidLoad()
        
        sdk = taskirx(
            apiKey: "test-api-key",
            endpoint: URL(string: "http://localhost:3000")!
        )
        
        // Test ad request
        Task {
            do {
                let ad = try await sdk.requestAd(
                    placement: "banner-300x250",
                    targeting: ["country": "US"]
                )
                print("Ad received: \(ad)")
            } catch {
                print("Error: \(error)")
            }
        }
        
        // Test impression tracking
        Task {
            try? await sdk.trackImpression(
                campaignId: "camp-123",
                placementId: "placement-456"
            )
        }
    }
}
```

**Test Cases**:
- ✅ SDK builds for iOS 14.0+
- ✅ Ad request works with async/await
- ✅ Impression tracking functional
- ✅ Network requests secure (HTTPS)
- ✅ Privacy manifest included
- ✅ App Tracking Transparency compliance

---

## Integration Testing

### 1. End-to-End User Flow

**Complete Advertiser Flow**:
```powershell
# 1. Register advertiser
$regBody = @{
    email = "advertiser@example.com"
    password = "Secure123!"
    name = "Test Advertiser"
    role = "advertiser"
} | ConvertTo-Json

$regResponse = Invoke-WebRequest -Uri "http://localhost:3000/api/auth/register" -Method POST -Headers @{'Content-Type'='application/json'} -Body $regBody
$token = ($regResponse.Content | ConvertFrom-Json).token
$userId = ($regResponse.Content | ConvertFrom-Json).user.id

# 2. Create campaign
$campBody = @{
    name = "E2E Test Campaign"
    advertiserId = $userId
    budget = 10000
    bidAmount = 3.0
    targeting = @{
        countries = @("US")
        deviceTypes = @("mobile")
    }
    startDate = (Get-Date).ToUniversalTime().ToString("o")
    endDate = (Get-Date).AddDays(30).ToUniversalTime().ToString("o")
} | ConvertTo-Json -Depth 10

$campResponse = Invoke-WebRequest -Uri "http://localhost:3000/api/campaigns" -Method POST -Headers @{'Authorization'="Bearer $token"; 'Content-Type'='application/json'} -Body $campBody
$campaignId = ($campResponse.Content | ConvertFrom-Json)._id

# 3. Simulate RTB bid win
$bidRequest = @{
    id = "e2e-req-1"
    imp = @(@{
        id = "imp-1"
        banner = @{w = 300; h = 250}
        bidfloor = 1.0
    })
    site = @{
        id = "site-1"
        domain = "example.com"
    }
    device = @{
        ua = "Mozilla/5.0..."
        ip = "203.0.113.1"
        geo = @{country = "USA"}
    }
} | ConvertTo-Json -Depth 10

$bidResponse = Invoke-WebRequest -Uri "http://localhost:3000/api/rtb/bid-request" -Method POST -Headers @{'Content-Type'='application/json'} -Body $bidRequest

# 4. Track MMP install
$installEvent = @{
    eventType = "install"
    provider = "appsflyer"
    campaignId = $campaignId
    userId = "e2e-user-1"
    deviceId = "e2e-device-1"
    timestamp = (Get-Date).ToUniversalTime().ToString("o")
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:3000/api/mmp/events/track" -Method POST -Headers @{'Authorization'="Bearer $token"; 'Content-Type'='application/json'} -Body $installEvent

# 5. Get campaign analytics
$analyticsResponse = Invoke-WebRequest -Uri "http://localhost:3000/api/analytics/campaigns/$campaignId" -Method GET -Headers @{'Authorization'="Bearer $token"}
Write-Host ($analyticsResponse.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10)

# 6. Record GDPR consent
$consentBody = @{
    userId = "e2e-user-1"
    consentGiven = $true
    purposes = @("advertising", "analytics")
    ipAddress = "203.0.113.1"
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:3000/api/consent" -Method POST -Headers @{'Content-Type'='application/json'} -Body $consentBody
```

**Expected Results**:
- ✅ User registered successfully
- ✅ Campaign created and active
- ✅ RTB bid response received
- ✅ MMP install tracked
- ✅ Analytics show 1 install
- ✅ Consent recorded

### 2. MMP Provider Integration

**Test All 6 MMP Providers**:
```powershell
$providers = @("appsflyer", "adjust", "branch", "kochava", "singular", "tenjin")

foreach ($provider in $providers) {
    Write-Host "Testing $provider..."
    
    $event = @{
        eventType = "install"
        provider = $provider
        campaignId = $campaignId
        userId = "test-user-$provider"
        deviceId = "test-device-$provider"
        timestamp = (Get-Date).ToUniversalTime().ToString("o")
    } | ConvertTo-Json
    
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:3000/api/mmp/events/track" -Method POST -Headers @{'Authorization'="Bearer $token"; 'Content-Type'='application/json'} -Body $event
        Write-Host "✓ $provider - OK"
    } catch {
        Write-Host "✗ $provider - FAILED"
    }
}
```

---

## Automated Testing

### 1. Unit Tests

**Run Unit Tests**:
```bash
npm test
```

**Test Coverage**:
```bash
npm run test:coverage
```

**Target Coverage**:
- Overall: >80%
- Models: >90%
- Routes: >85%
- Middleware: >90%
- Utils: >85%

### 2. Continuous Integration

**GitHub Actions Workflow** (`.github/workflows/test.yml`):
```yaml
name: Test Suite

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      mongodb:
        image: mongo:6.0
        ports:
          - 27017:27017
    
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '20.x'
      
      - run: npm install
      - run: npm test
      - run: npm run load-test
      - run: npm run verify
```

### 3. Smoke Tests

**Quick Smoke Test Script** (`scripts/smoke-test.js`):
```javascript
const axios = require('axios');

const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';

async function smokeTest() {
    const tests = [
        { name: 'Health Check', url: '/health', method: 'GET' },
        { name: 'API Docs', url: '/api-docs', method: 'GET' },
    ];
    
    for (const test of tests) {
        try {
            const response = await axios({ method: test.method, url: BASE_URL + test.url });
            console.log(`✓ ${test.name}: ${response.status}`);
        } catch (error) {
            console.error(`✗ ${test.name}: ${error.message}`);
            process.exit(1);
        }
    }
    
    console.log('\\n✓ All smoke tests passed');
}

smokeTest();
```

**Run Smoke Tests**:
```bash
node scripts/smoke-test.js
```

---

## OCI (Production) Testing

When deployed to Oracle Cloud, testing endpoints change from `localhost` to your production domains.

### 1. Connectivity Check

| Service        | Local URL                | Production URL (OCI)          |
|----------------|--------------------------|-------------------------------|
| Dashboard      | http://localhost:3001    | https://dashboard.taskirx.com  |
| API Backend    | http://localhost:3000    | https://api.taskirx.com        |
| Bidding Engine | http://localhost:8080    | https://bidding.taskirx.com    |

**Manual Health Verification**:
```bash
# PowerShell
Invoke-WebRequest -Uri "https://api.taskir.com/health" -Method GET
```

### 2. Post-Deployment Validation

After running `deploy-to-oci.ps1`, verify the cloud services:

**Kubernetes Pod Status**:
```bash
kubectl get pods -n taskir
# Expect: Status "Running" for all pods
```

**Ingress IP**:
```bash
kubectl get ingress -n taskir
# Expect: An IP address assigned to the ADDRESS column
```
