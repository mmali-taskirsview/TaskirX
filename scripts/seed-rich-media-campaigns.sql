INSERT INTO campaigns (
    id, "userId", "tenantId", name, status, type, budget, spent, "bidPrice", vertical, 
    targeting, "startDate", "endDate", creative, impressions, clicks, conversions, "createdAt", "updatedAt"
) VALUES (
    'a1b2c3d4-e5f6-47a8-b9c0-123450000001',
    '3422b448-2460-4fd2-9183-8000de6f8343', -- Advertiser ID from seed-advertiser-fix.sql
    '123e4567-e89b-12d3-a456-426614174000', -- Default Tenant
    'Rich Media Expandable Campaign',
    'active',
    'cpm',
    5000.00,
    0.00,
    2.50,
    'GAMING',
    '{"countries": ["US"], "devices": ["mobile", "desktop"]}',
    NOW(),
    NOW() + INTERVAL '30 days',
    '{
        "type": "rich_media",
        "url": "https://cdn.example.com/rich-media/expandable.js",
        "width": 300,
        "height": 250,
        "htmlSnippet": "<div id=\"ad-container\"><h1>Expandable Ad</h1><button>Expand</button></div>",
        "expandable": true
    }',
    0, 0, 0, NOW(), NOW()
);

INSERT INTO campaigns (
    id, "userId", "tenantId", name, status, type, budget, spent, "bidPrice", vertical, 
    targeting, "startDate", "endDate", creative, impressions, clicks, conversions, "createdAt", "updatedAt"
) VALUES (
    'a1b2c3d4-e5f6-47a8-b9c0-123450000002',
    '3422b448-2460-4fd2-9183-8000de6f8343',
    '123e4567-e89b-12d3-a456-426614174000',
    'Podcast Audio Ad',
    'active',
    'cpm',
    3000.00,
    0.00,
    1.50,
    'ENTERTAINMENT',
    '{"countries": ["US", "UK"], "devices": ["mobile"]}',
    NOW(),
    NOW() + INTERVAL '30 days',
    '{
        "type": "audio",
        "url": "https://cdn.example.com/audio/podcast-ad.mp3",
        "duration": 30,
        "mimeType": "audio/mp3",
        "bitrate": 128
    }',
    0, 0, 0, NOW(), NOW()
);

INSERT INTO campaigns (
    id, "userId", "tenantId", name, status, type, budget, spent, "bidPrice", vertical, 
    targeting, "startDate", "endDate", creative, impressions, clicks, conversions, "createdAt", "updatedAt"
) VALUES (
    'a1b2c3d4-e5f6-47a8-b9c0-123450000003',
    '3422b448-2460-4fd2-9183-8000de6f8343',
    '123e4567-e89b-12d3-a456-426614174000',
    'Performance Popunder',
    'active',
    'cpc',
    1000.00,
    0.00,
    0.05,
    'SOFTWARE',
    '{"countries": ["US"], "devices": ["desktop"]}',
    NOW(),
    NOW() + INTERVAL '30 days',
    '{
        "type": "pop",
        "url": "https://landing-page.example.com/offer",
        "width": 0,
        "height": 0
    }',
    0, 0, 0, NOW(), NOW()
);
