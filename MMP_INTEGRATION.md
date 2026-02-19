# Mobile Measurement Partner (MMP) Integration

TaskirX supports integration with major MMPs (AppsFlyer, Adjust, Branch, Kochava, internal) via a unified callback endpoint.

## Endpoint Details

**URL**: `POST https://<your-domain>/api/mmp/events/track`
**Headers**:
- `Content-Type`: `application/json`
- `Authorization`: `Bearer <token>` (Optional, depending on configuration)

### Supported Payload

```json
{
  "provider": "appsflyer",       // or "adjust", "branch", "generic"
  "eventType": "install",        // or "purchase", "register", "custom_event"
  "campaignId": "cmp-123456",    // Your campaign ID in TaskirX
  "userId": "user-789",          // Optional: Internal User ID
  "deviceId": "idfa-uuid",       // Optional: GAID/IDFA
  "revenue": 1.99,               // Optional: Transaction value
  "currency": "USD",             // Optional: Currency code
  "timestamp": "2023-10-27T10:00:00Z", // ISO 8601
  "metadata": {                  // Any additional data
    "af_status": "organic",
    "adgroup": "adg-1"
  }
}
```

## Postback URL Configuration (Server-to-Server)

For MMPs that support Postbacks (Callbacks), configure the following URL template in their dashboard:

**Template**:
`https://<your-domain>/api/mmp/postback?provider={provider}&campaign_id={campaign_id}&event_name={event_name}&revenue={revenue}&currency={currency}&idfa={idfa}&gaid={gaid}`

Example for AppsFlyer:
`https://api.taskirx.com/api/mmp/postback?provider=appsflyer&campaign_id={c}&event_name={event-name}&revenue={event-value}&currency={currency}`

## Features

1.  **Unified Ingestion**: All provider events are normalized into `mmp_events` ClickHouse table.
2.  **Real-Time Attribution**: Installs and Purchases automatically update Campaign stats in Redis (visible in Dashboard instantly).
3.  **Raw Data Access**: Full event payloads are stored in `analytics.mmp_events` for deep dive analysis.

## Verification

To test the integration locally or in production:

1.  **Send Test Event**:
    ```bash
    curl -X POST http://localhost:3000/api/mmp/events/track \
      -H "Content-Type: application/json" \
      -d '{
        "provider": "manual_test",
        "eventType": "install",
        "campaignId": "cmp-test-1",
        "revenue": 0.00
      }'
    ```

2.  **Check Logs**:
    Backend logs should show "MMP Event Received".

3.  **Verify Database**:
    Query ClickHouse: `SELECT * FROM analytics.mmp_events ORDER BY timestamp DESC LIMIT 5`
