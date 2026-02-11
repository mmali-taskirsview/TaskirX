-- TaskirX v3.0 - Database Seed Data
-- This SQL script creates initial test data

-- Admin User
-- Email: admin@taskirx.com
-- Password: Admin123! (hashed)
INSERT INTO users (id, email, "passwordHash", role, "companyName", "tenantId", "isActive", "createdAt", "updatedAt")
VALUES (
  gen_random_uuid(),
  'admin@taskirx.com',
  '$2a$10$rqGKjYGPEMHp5z7zqNMwOeHQKNZQJHvx8zLp5n9wLz8WqYXJYZQXe', -- Admin123!
  'admin',
  'TaskirX Inc',
  gen_random_uuid(),
  true,
  NOW(),
  NOW()
) ON CONFLICT (email) DO NOTHING;

-- Test Advertiser
-- Email: advertiser@test.com
-- Password: Test123!
INSERT INTO users (id, email, "passwordHash", role, "companyName", "tenantId", "isActive", "createdAt", "updatedAt")
VALUES (
  gen_random_uuid(),
  'advertiser@test.com',
  '$2a$10$rqGKjYGPEMHp5z7zqNMwOeHQKNZQJHvx8zLp5n9wLz8WqYXJYZQXe', -- Test123!
  'advertiser',
  'Test Advertiser Corp',
  gen_random_uuid(),
  true,
  NOW(),
  NOW()
) ON CONFLICT (email) DO NOTHING;

-- Test Publisher
-- Email: publisher@test.com
-- Password: Test123!
INSERT INTO users (id, email, "passwordHash", role, "companyName", "tenantId", "isActive", "createdAt", "updatedAt")
VALUES (
  gen_random_uuid(),
  'publisher@test.com',
  '$2a$10$rqGKjYGPEMHp5z7zqNMwOeHQKNZQJHvx8zLp5n9wLz8WqYXJYZQXe', -- Test123!
  'publisher',
  'Test Publisher Network',
  gen_random_uuid(),
  true,
  NOW(),
  NOW()
) ON CONFLICT (email) DO NOTHING;

-- Note: Campaigns, wallets, and transactions will be created through the API
-- This seed data only creates initial users for testing

-- Display created users
SELECT 
  id,
  email,
  role,
  "companyName",
  "isActive",
  "createdAt"
FROM users
ORDER BY "createdAt" DESC;
