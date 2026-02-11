/**
 * TaskirX JavaScript SDK - Complete Example
 * Demonstrates all major features and use cases
 */

import TaskirXClient from './src/client';
import { ClientConfig } from './src/types';

async function main() {
  // 1. Initialize the client
  const config: ClientConfig = {
    apiUrl: 'https://api.taskir.io',
    apiKey: process.env.TASKIR_API_KEY || 'your-api-key',
    debug: true,
    timeout: 30000,
    retryAttempts: 3,
  };

  const client = new TaskirXClient(config);

  try {
    console.log('=== TaskirX SDK Examples ===\n');

    // 2. Check platform health
    console.log('1. Platform Health Check');
    const health = await client.getHealth();
    console.log('✓ Platform is healthy:', health);

    // 3. Authentication
    console.log('\n2. Authentication Examples');

    // Register
    const registered = await client.auth.register(
      'developer@example.com',
      'SecurePassword123!',
      'Example Corp'
    );
    console.log('✓ User registered:', registered.email);

    // Login
    const loginResult = await client.auth.login(
      'developer@example.com',
      'SecurePassword123!'
    );
    console.log('✓ User logged in, token:', loginResult.token.substring(0, 20) + '...');

    // Get profile
    const profile = await client.auth.getProfile();
    console.log('✓ User profile:', profile.name, profile.email);

    // 4. Campaign Management
    console.log('\n3. Campaign Management Examples');

    // Create campaign
    const campaign = await client.campaigns.create({
      name: 'Summer Sale 2024',
      budget: 50000,
      startDate: '2024-06-01',
      endDate: '2024-08-31',
      targetAudience: {
        ageRange: '25-45',
        interests: ['ecommerce', 'fashion', 'technology'],
        geography: ['US', 'CA', 'UK'],
      },
    });
    console.log('✓ Campaign created:', campaign.id, campaign.name);

    // List campaigns
    const campaigns = await client.campaigns.list();
    console.log('✓ Total campaigns:', campaigns.length);

    // Get campaign details
    const campaignDetails = await client.campaigns.get(campaign.id);
    console.log('✓ Campaign budget:', campaignDetails.budget);

    // Update campaign
    const updated = await client.campaigns.update(campaign.id, {
      budget: 75000,
      description: 'Updated summer sale campaign',
    });
    console.log('✓ Campaign updated, new budget:', updated.budget);

    // 5. Analytics
    console.log('\n4. Analytics Examples');

    // Real-time analytics
    const realtime = await client.analytics.getRealtime();
    console.log('✓ Real-time analytics:');
    console.log('  - Impressions:', realtime.impressions);
    console.log('  - Clicks:', realtime.clicks);
    console.log('  - Conversions:', realtime.conversions);
    console.log('  - CTR:', (realtime.ctr * 100).toFixed(2) + '%');
    console.log('  - Conversion Rate:', (realtime.conversionRate * 100).toFixed(2) + '%');

    // Campaign analytics
    const campaignAnalytics = await client.analytics.getCampaignAnalytics(campaign.id, {
      startDate: '2024-01-01',
      endDate: '2024-12-31',
    });
    console.log('✓ Campaign analytics retrieved');

    // Device breakdown
    const deviceBreakdown = await client.analytics.getBreakdown(campaign.id, 'device');
    console.log('✓ Device breakdown:', Object.keys(deviceBreakdown).join(', '));

    // 6. Ad Management
    console.log('\n5. Ad Management Examples');

    // Create ad
    const ad = await client.ads.create({
      campaignId: campaign.id,
      placement: 'homepage-banner',
      creativeUrl: 'https://example.com/ads/summer-sale.html',
      clickUrl: 'https://example.com/summer-sale',
      width: 728,
      height: 90,
      status: 'active',
    });
    console.log('✓ Ad created:', ad.id);

    // List ads
    const ads = await client.ads.list(campaign.id);
    console.log('✓ Total ads:', ads.length);

    // 7. Bidding Engine
    console.log('\n6. Bidding Engine Examples');

    // Submit bid
    const bid = await client.bidding.submitBid({
      campaignId: campaign.id,
      adSlotId: 'slot-premium-001',
      amount: 2.5,
      currency: 'USD',
    });
    console.log('✓ Bid submitted:', bid.id, '$' + bid.amount);

    // Get recommendations
    const recommendations = await client.bidding.getRecommendations(campaign.id);
    console.log('✓ Bid recommendations:');
    recommendations.forEach((rec: any, index: number) => {
      console.log(`  ${index + 1}. $${rec.recommendedBid.toFixed(2)} (${rec.reasoning})`);
    });

    // Get bid stats
    const bidStats = await client.bidding.getStats(campaign.id);
    console.log('✓ Bid statistics:', bidStats);

    // 8. Webhook Management
    console.log('\n7. Webhook Management Examples');

    // Subscribe to webhook
    const webhook = await client.webhooks.subscribe({
      url: 'https://example.com/webhooks/taskir',
      events: [
        'campaign.created',
        'campaign.updated',
        'bid.won',
        'conversion.recorded',
      ],
      active: true,
    });
    console.log('✓ Webhook subscribed:', webhook.id);

    // List webhooks
    const webhooks = await client.webhooks.list();
    console.log('✓ Total webhooks:', webhooks.length);

    // Register event handler
    client.webhooks.onEvent('conversion.recorded', (event) => {
      console.log('✓ Conversion recorded:', event.data);
    });

    // Test webhook
    const testResult = await client.webhooks.test(webhook.id);
    console.log('✓ Webhook test result:', testResult.status);

    // 9. Advanced Operations
    console.log('\n8. Advanced Operations Examples');

    // Dashboard
    const dashboard = await client.getDashboard();
    console.log('✓ Dashboard retrieved with', dashboard.campaigns.length, 'campaigns');

    // Campaign performance
    const performance = await client.getCampaignPerformance(campaign.id);
    console.log('✓ Campaign performance:');
    console.log('  - Status:', performance.campaign.status);
    console.log('  - Budget spent:', performance.analytics.spend);
    console.log('  - ROI:', (performance.analytics.roi * 100).toFixed(2) + '%');

    // Batch create campaigns
    const batchCampaigns = await client.createCampaigns([
      {
        name: 'Spring Sale',
        budget: 30000,
        startDate: '2024-03-01',
        endDate: '2024-05-31',
      },
      {
        name: 'Fall Sale',
        budget: 40000,
        startDate: '2024-09-01',
        endDate: '2024-11-30',
      },
    ]);
    console.log('✓ Batch created', batchCampaigns.length, 'campaigns');

    // Platform statistics
    const stats = await client.getStatistics();
    console.log('✓ Platform statistics retrieved');

    // 10. Campaign Control
    console.log('\n9. Campaign Control Examples');

    // Pause campaign
    await client.campaigns.pause(campaign.id);
    console.log('✓ Campaign paused');

    // Resume campaign
    await client.campaigns.resume(campaign.id);
    console.log('✓ Campaign resumed');

    // 11. Debug Mode
    console.log('\n10. Debug Features');
    client.enableDebug(true);
    console.log('✓ Debug mode enabled - detailed logs will show for all requests');
    client.enableDebug(false);
    console.log('✓ Debug mode disabled');

    // 12. Cleanup
    console.log('\n11. Cleanup');

    // Delete ad
    await client.ads.delete(ad.id);
    console.log('✓ Ad deleted');

    // Delete campaign
    await client.campaigns.delete(campaign.id);
    console.log('✓ Campaign deleted');

    // Logout
    await client.logout();
    console.log('✓ User logged out');

    console.log('\n=== All Examples Completed Successfully ===');
  } catch (error: any) {
    console.error('Error:', error.code || error.message);
    console.error('Details:', error.details);
  }
}

// Run examples
main().catch(console.error);
