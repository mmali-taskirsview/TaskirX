-- Performance Optimization Indexes for TaskirX - Phase 1
-- Using correct column names (quoted for case sensitivity)

-- Index 1: Campaign filtering by tenant and status
CREATE INDEX IF NOT EXISTS idx_campaigns_tenant_status 
  ON campaigns("tenantId", status) 
  WHERE status = 'active';

-- Index 2: Impressions by campaign (filtered for recent data)
CREATE INDEX IF NOT EXISTS idx_impressions_campaign_date 
  ON impressions("campaignId", "createdAt" DESC);

-- Index 3: Attributions by campaign
CREATE INDEX IF NOT EXISTS idx_attributions_campaign_date 
  ON attributions("campaignId", "createdAt" DESC);

-- Index 4: Transactions by user
CREATE INDEX IF NOT EXISTS idx_transactions_user 
  ON transactions("userId", "createdAt" DESC);

-- Index 5: Bids by campaign with timestamp
CREATE INDEX IF NOT EXISTS idx_bids_campaign_timestamp 
  ON bids("campaignId", "createdAt" DESC);

-- Index 6: User lookups by email
CREATE INDEX IF NOT EXISTS idx_users_email 
  ON users(email);

-- Analyze tables to update statistics
ANALYZE campaigns;
ANALYZE impressions;
ANALYZE attributions;
ANALYZE transactions;
ANALYZE bids;
ANALYZE users;
