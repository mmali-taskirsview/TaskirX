INSERT INTO users (id, email, "passwordHash", role, "companyName", "tenantId", "isActive", "createdAt", "updatedAt")
VALUES (
  gen_random_uuid(),
  'publisher@test.com',
  (SELECT "passwordHash" FROM users WHERE email='admin@taskirx.com'),
  'publisher',
  'Test Publisher Manual',
  gen_random_uuid(),
  true,
  NOW(),
  NOW()
) ON CONFLICT (email) DO NOTHING;
