-- Insert specific user to link campaign
INSERT INTO users (id, email, "passwordHash", role, "companyName", "tenantId", "isActive", "createdAt", "updatedAt")
VALUES (
  '3fa85f64-5717-4562-b3fc-2c963f66afa6',
  'campaign_owner@test.com',
  '$2a$10$rqGKjYGPEMHp5z7zqNMwOeHQKNZQJHvx8zLp5n9wLz8WqYXJYZQXe',
  'advertiser',
  'Test Corp',
  '3fa85f64-5717-4562-b3fc-2c963f66afa6',
  true,
  NOW(),
  NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert Test Campaign
INSERT INTO campaigns (id, name, "userId", "tenantId", status, type, budget, spent, "bidPrice", targeting, creative, "createdAt", "updatedAt")
VALUES (
    '3fa85f64-5717-4562-b3fc-2c963f66afa7', -- Valid UUID
    'Test Tech Campaign',
    '3fa85f64-5717-4562-b3fc-2c963f66afa6', -- userId
    '3fa85f64-5717-4562-b3fc-2c963f66afa6', -- tenantId
    'active',
    'cpm',
    1000.00,
    0.00,
    1.50,
    '{"countries": ["US"], "categories": ["tech", "sports"]}', -- Targeting US and Tech
    '{"type": "banner", "url": "http://ads.com/banner.jpg", "width": 300, "height": 250}',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;
