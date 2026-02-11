# TaskirX JavaScript SDK

Enterprise-grade JavaScript SDK for the TaskirX advertising platform. Build, manage, and optimize ad campaigns with powerful APIs and real-time analytics.

## Features

✨ **Complete API Coverage**
- Campaign management (create, list, update, pause, resume)
- Real-time analytics and reporting
- Intelligent bidding engine with AI recommendations
- Ad placement management
- Webhook subscriptions and event handling
- User authentication and profile management

🚀 **Production Ready**
- Full TypeScript support with strict type safety
- Comprehensive error handling and retry logic
- Exponential backoff for failed requests
- Request timeout handling
- Detailed logging and debug mode
- 50+ unit and integration tests

⚡ **Performance Optimized**
- Request deduplication
- Connection pooling
- Automatic token refresh
- Minimal dependencies
- < 30KB minified bundle

## Installation

```bash
npm install @taskir/js-sdk
# or
yarn add @taskir/js-sdk
```

## Quick Start

```javascript
import TaskirXClient from '@taskir/js-sdk';

// Initialize the client
const client = new TaskirXClient({
  apiUrl: 'https://api.taskir.io',
  apiKey: 'your-api-key',
  debug: true // Enable detailed logging
});

// Get user profile
const profile = await client.getProfile();
console.log('User:', profile.name);

// List campaigns
const campaigns = await client.campaigns.list();
console.log('Campaigns:', campaigns);

// Create a new campaign
const campaign = await client.campaigns.create({
  name: 'Summer Campaign 2024',
  budget: 10000,
  startDate: '2024-06-01',
  endDate: '2024-08-31',
  targetAudience: { ageRange: '25-45', interests: ['technology', 'business'] }
});

// Get real-time analytics
const analytics = await client.analytics.getRealtime();
console.log('Impressions:', analytics.impressions);
console.log('Clicks:', analytics.clicks);
console.log('CTR:', analytics.ctr);
```

## Configuration

### Basic Configuration

```javascript
const client = new TaskirXClient({
  apiUrl: 'https://api.taskir.io',
  apiKey: 'your-api-key'
});
```

### Advanced Configuration

```javascript
const client = new TaskirXClient({
  // Required
  apiUrl: 'https://api.taskir.io',
  apiKey: 'your-api-key',
  
  // Optional
  debug: false,                    // Enable debug logging
  timeout: 30000,                 // Request timeout in ms (default: 30000)
  retryAttempts: 3,               // Number of retry attempts (default: 3)
  baseHeaders: {                  // Custom headers for all requests
    'X-Custom-Header': 'value'
  }
});
```

## Service APIs

### Authentication Service

```javascript
// Register new user
await client.auth.register({
  email: 'user@example.com',
  password: 'secure-password',
  name: 'John Doe'
});

// Login
const result = await client.auth.login({
  email: 'user@example.com',
  password: 'secure-password'
});

// Get current user profile
const profile = await client.auth.getProfile();

// Refresh authentication token
await client.auth.refreshToken();

// Logout
await client.auth.logout();
```

### Campaign Management

```javascript
// Create campaign
const campaign = await client.campaigns.create({
  name: 'Q3 Campaign',
  budget: 50000,
  startDate: '2024-07-01',
  endDate: '2024-09-30'
});

// List campaigns
const campaigns = await client.campaigns.list();

// Get campaign details
const campaign = await client.campaigns.get('campaign-id');

// Update campaign
await client.campaigns.update('campaign-id', {
  budget: 60000,
  status: 'active'
});

// Pause campaign
await client.campaigns.pause('campaign-id');

// Resume campaign
await client.campaigns.resume('campaign-id');

// Delete campaign
await client.campaigns.delete('campaign-id');
```

### Analytics

```javascript
// Real-time analytics (last hour)
const realtime = await client.analytics.getRealtime();
// Returns: { impressions, clicks, conversions, ctr, conversionRate, ... }

// Campaign analytics
const campaignStats = await client.analytics.getCampaignAnalytics('campaign-id', {
  startDate: '2024-01-01',
  endDate: '2024-12-31'
});

// Device breakdown
const deviceBreakdown = await client.analytics.getBreakdown(
  'campaign-id',
  'device'  // 'device' | 'geo' | 'browser' | 'os'
);

// Full dashboard
const dashboard = await client.analytics.getDashboard();
```

### Bidding Engine

```javascript
// Submit a bid
const bid = await client.bidding.submitBid({
  campaignId: 'campaign-id',
  adSlotId: 'slot-id',
  amount: 2.50,
  currency: 'USD'
});

// Get bid recommendations
const recommendations = await client.bidding.getRecommendations('campaign-id');
// Returns AI-powered bid suggestions based on CTR and conversions

// Get bid statistics
const stats = await client.bidding.getStats('campaign-id');

// List all bids
const bids = await client.bidding.getBids('campaign-id');

// Get specific bid
const bid = await client.bidding.getBid('bid-id');
```

### Ad Management

```javascript
// Create ad placement
const ad = await client.ads.create({
  campaignId: 'campaign-id',
  placement: 'homepage-banner',
  creativeUrl: 'https://example.com/ad.html',
  clickUrl: 'https://example.com',
  width: 728,
  height: 90
});

// List ads
const ads = await client.ads.list('campaign-id');

// Get ad details
const ad = await client.ads.get('ad-id');

// Update ad
await client.ads.update('ad-id', {
  creativeUrl: 'https://example.com/new-ad.html'
});

// Delete ad
await client.ads.delete('ad-id');
```

### Webhooks

```javascript
// Subscribe to webhook events
const webhook = await client.webhooks.subscribe({
  url: 'https://example.com/webhooks/taskir',
  events: ['campaign.created', 'bid.won', 'conversion.recorded'],
  active: true
});

// List webhooks
const webhooks = await client.webhooks.list();

// Get webhook details
const webhook = await client.webhooks.get('webhook-id');

// Update webhook
await client.webhooks.update('webhook-id', {
  events: ['campaign.updated', 'impression.recorded']
});

// Test webhook delivery
await client.webhooks.test('webhook-id');

// Get webhook delivery logs
const logs = await client.webhooks.getLogs('webhook-id', 100);

// Register local event handler
client.webhooks.onEvent('campaign.created', (event) => {
  console.log('New campaign:', event.data);
});

// Remove event handler
client.webhooks.offEvent('campaign.created');
```

## Advanced Usage

### Dashboard View

```javascript
// Get comprehensive dashboard
const dashboard = await client.getDashboard();
// Returns: { analytics, campaigns, webhooks, timestamp }
```

### Campaign Performance

```javascript
// Get complete campaign performance metrics
const performance = await client.getCampaignPerformance('campaign-id');
// Returns: { campaign, analytics, bidding }
```

### Batch Operations

```javascript
// Create multiple campaigns
const campaigns = await client.createCampaigns([
  { name: 'Campaign 1', budget: 10000 },
  { name: 'Campaign 2', budget: 20000 },
  { name: 'Campaign 3', budget: 15000 }
]);
```

### Statistics

```javascript
// Get complete platform statistics
const stats = await client.getStatistics();
// Includes dashboard + realtime analytics
```

### Error Handling

```javascript
try {
  const campaign = await client.campaigns.get('invalid-id');
} catch (error) {
  if (error.code === 'NOT_FOUND') {
    console.log('Campaign not found');
  } else if (error.code === 'UNAUTHORIZED') {
    console.log('Invalid API key');
  } else if (error.code === 'RATE_LIMIT_EXCEEDED') {
    console.log('Rate limit exceeded, retry later');
  } else {
    console.error('Error:', error.message);
  }
}
```

### Debug Mode

```javascript
// Enable debug logging
client.enableDebug(true);

// All requests and responses will be logged
// Useful for development and troubleshooting
```

### Custom API Key Management

```javascript
// Update API key at runtime
client.setApiKey('new-api-key');

// Useful for token refresh or multi-account scenarios
```

## TypeScript Support

Full TypeScript support with strict type checking:

```typescript
import TaskirXClient, { ClientConfig, Campaign, Analytics } from '@taskir/js-sdk';

const config: ClientConfig = {
  apiUrl: 'https://api.taskir.io',
  apiKey: 'your-api-key'
};

const client = new TaskirXClient(config);

const campaign: Campaign = await client.campaigns.get('campaign-id');
const analytics: Analytics = await client.analytics.getRealtime();
```

## Error Handling

The SDK includes comprehensive error handling with custom error codes:

- `BAD_REQUEST` (400) - Invalid request parameters
- `UNAUTHORIZED` (401) - Invalid or missing API key
- `FORBIDDEN` (403) - Access denied
- `NOT_FOUND` (404) - Resource not found
- `RATE_LIMIT_EXCEEDED` (429) - Too many requests
- `SERVER_ERROR` (500) - Server error
- `SERVICE_UNAVAILABLE` (503) - Service temporarily unavailable

## Logging

Control logging behavior:

```javascript
// Enable debug logging
client.enableDebug(true);

// Debug mode provides:
// - All API requests and responses
// - Service initialization logs
// - Error stack traces
// - Performance metrics
```

## Testing

Run the test suite:

```bash
npm test                    # Run all tests
npm run test:watch        # Watch mode
npm run test:coverage     # Coverage report
```

Test coverage includes:
- Unit tests for all services
- Integration tests for API workflows
- Error scenario testing
- Type safety validation

## Performance

The SDK is optimized for production use:

- **Bundle Size**: < 30KB minified
- **Dependencies**: Minimal (TypeScript only for type checking)
- **Response Time**: < 100ms average
- **Timeout**: 30 seconds default (configurable)
- **Retries**: Exponential backoff (100ms, 300ms, 900ms)

## Compatibility

- Node.js 16+
- Modern browsers (ES2020+)
- TypeScript 4.5+

## License

MIT

## Support

For issues, feature requests, or questions:
- GitHub Issues: https://github.com/taskir/js-sdk/issues
- Documentation: https://docs.taskir.io
- Email: support@taskir.io

## Changelog

### 1.0.0 (2024)
- Initial release
- Full API coverage
- Complete TypeScript support
- Comprehensive testing
- Production-ready error handling
