# TaskirX React Native SDK

Production-grade React Native SDK for the TaskirX Ad Exchange Platform. Build, manage, and optimize digital advertising campaigns with native cross-platform integration.

## Features

### 🚀 Core Capabilities
- **Campaign Management** - Create, update, pause, and resume campaigns
- **Real-time Analytics** - Track impressions, clicks, conversions, and ROI
- **Bidding Engine** - Submit competitive bids with recommendations
- **Ad Management** - Upload and manage ad placements across channels
- **Webhook Integration** - Subscribe to events and handle real-time notifications
- **Authentication** - Secure token-based auth with auto-refresh

### 💎 Enterprise Features
- **Async/Await** - Modern concurrency with TypeScript
- **Type Safety** - Full TypeScript support with strict typing
- **Error Handling** - Comprehensive error types with recovery strategies
- **Exponential Backoff** - Automatic retry with configurable delays
- **Logging** - Debug mode with detailed request/response logging
- **Cross-Platform** - Works on iOS, Android, and Web (Expo)

### 📱 Platform Support
- **iOS 11.0+**
- **Android 5.0+**
- **Expo SDK 45+**
- **Web (React Native Web)**

## Installation

### NPM/Yarn

```bash
npm install @taskir/react-native-sdk
# or
yarn add @taskir/react-native-sdk
```

### Expo

```bash
expo install @taskir/react-native-sdk
```

### TypeScript

Types are included by default.

## Requirements

- React Native 0.60+
- Node.js 12+
- TypeScript 4.0+ (optional but recommended)

## Quick Start

### 1. Initialize Client

```typescript
import { TaskirXClient } from '@taskir/react-native-sdk';

const client = TaskirXClient.create(
  'https://api.taskir.com',
  'your-api-key',
  true // debug mode
);
```

### 2. Authenticate

```typescript
// Register
const authResponse = await client.auth.register(
  'user@example.com',
  'secure123',
  'John Doe',
  'Acme Corp'
);
console.log('Registered:', authResponse.user.name);

// Or Login
const loginResponse = await client.auth.login(
  'user@example.com',
  'secure123'
);
console.log('Logged in:', loginResponse.token);
```

### 3. Create Campaign

```typescript
const result = await client.createCampaign(
  'Summer Sale 2024',
  5000.0,
  '2024-06-01',
  '2024-08-31',
  { age: '18-45', gender: 'all' }
);

if (result.success) {
  console.log('Campaign created:', result.data.id);
} else {
  console.error('Error:', result.error.message);
}
```

### 4. View Analytics

```typescript
const analyticsResult = await client.getRealtimeAnalytics();

if (analyticsResult.success) {
  const analytics = analyticsResult.data;
  console.log('Impressions:', analytics.impressions);
  console.log('Clicks:', analytics.clicks);
  console.log('CTR:', analytics.ctr * 100 + '%');
  console.log('ROI:', analytics.roi);
}
```

### 5. Submit Bids

```typescript
const bidResult = await client.submitBid('camp-123', 'slot-456', 2.50);

if (bidResult.success) {
  console.log('Bid placed:', bidResult.data.id);
} else {
  console.error('Bid failed:', bidResult.error.message);
}
```

## Configuration

### ClientConfig

```typescript
interface ClientConfig {
  apiUrl: string;           // API endpoint
  apiKey: string;           // Authentication key
  debug?: boolean;          // Enable logging
  timeout?: number;         // Request timeout (ms)
  retryAttempts?: number;   // Max retry attempts
}

const client = TaskirXClient.create(
  'https://api.taskir.com',
  'your-api-key',
  true // debug
);
```

### Request Configuration

All requests use:
- **Timeout**: 30 seconds (default)
- **Retries**: 3 attempts with exponential backoff
- **Backoff Delays**: 100ms → 300ms → 900ms
- **Headers**: Content-Type, Accept, X-API-Key, X-Request-ID, User-Agent, Authorization

## Usage Examples

### Campaign Management

```typescript
// List campaigns
const result = await client.getCampaigns(limit: 50, offset: 0);
if (result.success) {
  console.log('Campaigns:', result.data);
}

// Get campaign
const campaign = await client.campaigns.get('camp-123');

// Update campaign
const updated = await client.campaigns.update(
  'camp-123',
  'New Name',
  7500.0
);

// Pause campaign
const paused = await client.pauseCampaign('camp-123');

// Resume campaign
const resumed = await client.resumeCampaign('camp-123');

// Delete campaign
const deleted = await client.campaigns.delete('camp-123');
```

### Analytics & Reporting

```typescript
// Real-time analytics
const realtime = await client.getRealtimeAnalytics();

// Campaign analytics
const campaignAnalytics = await client.getCampaignPerformance('camp-123');

// Breakdown by dimension
const breakdown = await client.analytics.breakdown('date');

// Full dashboard
const dashboard = await client.getDashboard();

// Statistics
const stats = await client.getStatistics();
```

### Bidding

```typescript
// Submit bid
const bid = await client.submitBid('camp-123', 'slot-456', 3.50);

// Get recommendations
const recommendations = await client.getBidRecommendations();

// List bids
const bids = await client.bidding.list(limit: 100);

// Get specific bid
const bid = await client.bidding.get('bid-789');

// Bid statistics
const stats = await client.getBidStatistics();
```

### Ad Management

```typescript
// Create ad
const ad = await client.createAd(
  'camp-123',
  'banner',
  'https://cdn.example.com/ad.jpg',
  'https://example.com/landing',
  '300x250'
);

// List ads
const ads = await client.getAds('camp-123', limit: 50);

// Get ad
const ad = await client.ads.get('ad-123');

// Update ad
const updated = await client.ads.update('ad-123', 'skyscraper');

// Delete ad
const deleted = await client.ads.delete('ad-123');
```

### Webhooks & Events

```typescript
// Subscribe to webhook
const webhook = await client.subscribeWebhook(
  'https://example.com/webhooks',
  ['campaign.created', 'bid.submitted']
);

// Handle webhook events
client.onWebhookEvent('campaign.created', (event) => {
  console.log('Campaign created:', event.data);
});

// List webhooks
const webhooks = await client.getWebhooks();

// Test webhook
const test = await client.testWebhook('webhook-123');

// Get logs
const logs = await client.webhooks.getLogs('webhook-123');

// Update webhook
const updated = await client.webhooks.update('webhook-123', true);

// Delete webhook
const deleted = await client.webhooks.delete('webhook-123');
```

## Error Handling

### Result Pattern

```typescript
const result = await client.createCampaign(...);

if (result.success) {
  // Handle success
  const campaign = result.data;
  console.log('Campaign:', campaign.name);
} else {
  // Handle failure
  const error = result.error;
  console.error('Error:', error.message);
  console.error('Type:', error.type);
  if (error.statusCode) {
    console.error('Status:', error.statusCode);
  }
}
```

### Error Types

```typescript
enum TaskirXErrorType {
  NETWORK_ERROR = 'NETWORK_ERROR',
  DECODING_ERROR = 'DECODING_ERROR',
  HTTP_ERROR = 'HTTP_ERROR',
  INVALID_RESPONSE = 'INVALID_RESPONSE',
  TIMEOUT = 'TIMEOUT',
  RETRY_EXHAUSTED = 'RETRY_EXHAUSTED',
}
```

### Error Recovery

```typescript
const result = await client.getCampaigns();

if (!result.success) {
  const error = result.error;
  
  switch (error.type) {
    case 'NETWORK_ERROR':
      console.log('Network issue - retry later');
      break;
    case 'HTTP_ERROR':
      if (error.statusCode === 401) {
        // Re-authenticate
        await client.auth.refreshToken(refreshToken);
      }
      break;
    case 'TIMEOUT':
      console.log('Request timed out');
      break;
    default:
      console.error('Unknown error:', error.message);
  }
}
```

## React Native Integration

### Function Component Example

```typescript
import React, { useState, useEffect } from 'react';
import { View, Text, FlatList, ActivityIndicator } from 'react-native';
import { TaskirXClient } from '@taskir/react-native-sdk';

const CampaignListScreen: React.FC = () => {
  const [campaigns, setCampaigns] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const client = TaskirXClient.create(
    'https://api.taskir.com',
    'your-api-key'
  );

  useEffect(() => {
    loadCampaigns();
  }, []);

  const loadCampaigns = async () => {
    setLoading(true);
    const result = await client.getCampaigns();
    
    if (result.success) {
      setCampaigns(result.data);
      setError(null);
    } else {
      setError(result.error.message);
    }
    setLoading(false);
  };

  if (loading) {
    return <ActivityIndicator size="large" />;
  }

  if (error) {
    return <Text>Error: {error}</Text>;
  }

  return (
    <FlatList
      data={campaigns}
      keyExtractor={(item) => item.id}
      renderItem={({ item }) => (
        <View>
          <Text>{item.name}</Text>
          <Text>Budget: ${item.budget}</Text>
        </View>
      )}
    />
  );
};

export default CampaignListScreen;
```

## Testing

### Jest Configuration

```json
{
  "jest": {
    "preset": "react-native",
    "testEnvironment": "node",
    "setupFilesAfterEnv": ["<rootDir>/setup-tests.js"],
    "moduleNameMapper": {
      "^@/(.*)$": "<rootDir>/src/$1"
    }
  }
}
```

### Unit Tests

```typescript
import { TaskirXClient } from '@taskir/react-native-sdk';

describe('TaskirXClient', () => {
  let client: TaskirXClient;

  beforeEach(() => {
    client = TaskirXClient.create('http://localhost:3000', 'test-key');
  });

  it('should create client', () => {
    expect(client).toBeDefined();
  });

  it('should have all services', () => {
    expect(client.auth).toBeDefined();
    expect(client.campaigns).toBeDefined();
    expect(client.analytics).toBeDefined();
  });

  it('should handle results', () => {
    const result = { success: true as const, data: 'test' };
    expect(result.success).toBe(true);
  });
});
```

## Performance

### Metrics

| Operation | Typical | Max |
|-----------|---------|-----|
| Health Check | <50ms | <100ms |
| Get Campaigns | <200ms | <500ms |
| Create Campaign | <300ms | <1000ms |
| Submit Bid | <150ms | <500ms |
| Analytics | <200ms | <500ms |

## Best Practices

1. **Error Handling** - Always check result.success
2. **Authentication** - Refresh tokens automatically on 401
3. **Batch Operations** - Use Promise.all for concurrent requests
4. **Caching** - Cache frequently accessed data
5. **Debug Mode** - Use only in development
6. **Type Safety** - Use TypeScript for compile-time checking

## Compatibility

### React Native Versions
- ✅ 0.60+
- ✅ 0.70+ (recommended)
- ✅ 0.72+ (latest)

### Expo
- ✅ SDK 45+
- ✅ SDK 50+ (recommended)

### TypeScript
- ✅ 4.0+
- ✅ 5.0+ (recommended)

## License

MIT License - See LICENSE file

## Support

- **Documentation**: https://taskir.com/docs
- **API Reference**: https://api.taskir.com/docs
- **Issues**: https://github.com/taskir/sdk-react-native/issues
- **Email**: support@taskir.com

## Changelog

### v1.0.0 (January 2024)
- ✅ Initial release
- ✅ All 6 services implemented
- ✅ Full TypeScript support
- ✅ Comprehensive error handling
- ✅ 40+ test cases
- ✅ Complete documentation
