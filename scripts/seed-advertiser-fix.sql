UPDATE users 
SET "passwordHash" = (SELECT "passwordHash" FROM users WHERE email='admin@taskirx.com')
WHERE email = 'advertiser@test.com';
