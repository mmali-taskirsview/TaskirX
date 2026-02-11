-- Insert wallet for admin user if not exists
INSERT INTO wallets (id, "userId", balance, currency, "totalDeposited", "totalSpent", "tenantId")
SELECT 
  '11111111-1111-1111-1111-111111111111',
  'a585ba5c-b5f3-451e-98b5-c30e3a2295ac',
  500.00,
  'USD',
  500.00,
  0.00,
  '9206823f-cf21-4cb9-bfdd-95517ca0b189'
WHERE NOT EXISTS (
  SELECT 1 FROM wallets WHERE "userId" = 'a585ba5c-b5f3-451e-98b5-c30e3a2295ac'
);
