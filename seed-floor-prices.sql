INSERT INTO floor_prices (name, rule_type, price, action, conditions, publisher_id, priority, is_active)
SELECT 'US Premium Floor', 'geo', 2.50, 'set_floor', '{"countries": ["US", "CA", "GB"]}'::jsonb, id, 10, true FROM publishers WHERE email='demo@publisher.com';

INSERT INTO floor_prices (name, rule_type, price, action, conditions, publisher_id, priority, is_active)
SELECT 'Mobile Multiplier', 'device', 0.8, 'multiply', '{"devices": ["mobile"]}'::jsonb, id, 5, true FROM publishers WHERE email='demo@publisher.com';

INSERT INTO floor_prices (name, rule_type, price, action, conditions, publisher_id, priority, is_active)
SELECT 'Global Minimum', 'global', 0.50, 'set_floor', '{}'::jsonb, id, 1, true FROM publishers WHERE email='demo@publisher.com';
