-- Seed Data for Performance Tests
-- Use these IDs in locustfile.py

-- 1. Create Publishers (forcing UUIDs to match locustfile)
INSERT INTO publishers (id, name, email, domains, status, tier, createdat, updatedat)
VALUES 
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Performance Test Publisher 1', 'perf1@example.com', 'perf-test-1.com', 'active', 'premium', NOW(), NOW()),
  ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Performance Test Publisher 2', 'perf2@example.com', 'perf-test-2.com', 'active', 'standard', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 2. Create Ad Units linked to publishers
-- Using locustfile ad_unit_ids: au-123 (mapped to uuid), au-456, au-789
INSERT INTO ad_units (id, name, publisherid, type, status, size, createdat, updatedat)
VALUES 
  ('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380123', 'Perf Header Bidder', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'banner', 'active', '300x250,728x90', NOW(), NOW()),
  ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380456', 'Perf Footer Sticky', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'banner', 'active', '320x50', NOW(), NOW()),
  ('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380789', 'Perf Sidebar Video', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'video_outstream', 'active', 'outstream', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 3. Create Demand Partners (DSP) to actually bid
INSERT INTO demand_partners (id, name, code, endpoint, type, status, createdat, updatedat)
VALUES
  ('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380999', 'Mock Internal DSP', 'mock-dsp-001', 'http://localhost:3000/api/dsp/bid', 'dsp', 'active', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 4. Create DSP Supply Partners (for DSP side logic)
-- This matches the 'supplyPartnerId' sent in the bid request
INSERT INTO dsp_supply_partners (id, name, code, endpoint_url, status, max_bid, min_bid, created_at, updated_at)
VALUES
  ('11eebc99-9c0b-4ef8-bb6d-6bb9bd380111', 'SSP Partner 1', 'ssp-001', 'http://ssp-partner-1.com/api', 'active', 20.0, 0.1, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 5. Create DSP Bid Strategies
-- Needed for DSP logic to return a bid
INSERT INTO dsp_bid_strategies (id, name, type, status, base_bid, max_bid, created_at, updated_at)
VALUES
  ('22eebc99-9c0b-4ef8-bb6d-6bb9bd380222', 'Default Strategy', 'manual', 'active', 5.0, 50.0, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 6. Create Test User for Analytics Auth
-- Requires pgcrypto extension for password hashing
CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO users (id, email, "passwordHash", role, "tenantId", "createdAt", "updatedAt")
VALUES (
  '33eebc99-9c0b-4ef8-bb6d-6bb9bd380333',
  'test@example.com',
  crypt('Test123!', gen_salt('bf')),
  'advertiser',
  '44eebc99-9c0b-4ef8-bb6d-6bb9bd380444',
  NOW(),
  NOW()
)
ON CONFLICT (email) DO NOTHING;
