-- ClickHouse Optimization: Materialized Views for High-Speed Dashboards

CREATE MATERIALIZED VIEW IF NOT EXISTS analytics.mv_hourly_campaign_impressions
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (hour, campaignId, publisherId, country, deviceType)
AS SELECT
    toStartOfHour(timestamp) AS hour,
    campaignId,
    publisherId,
    country,
    deviceType,
    count() AS impressions
FROM analytics.impressions
GROUP BY hour, campaignId, publisherId, country, deviceType;

CREATE MATERIALIZED VIEW IF NOT EXISTS analytics.mv_hourly_campaign_clicks
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (hour, campaignId)
AS SELECT
    toStartOfHour(timestamp) AS hour,
    campaignId,
    count() AS clicks
FROM analytics.clicks
GROUP BY hour, campaignId;

CREATE MATERIALIZED VIEW IF NOT EXISTS analytics.mv_hourly_campaign_conversions
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (hour, campaignId)
AS SELECT
    toStartOfHour(timestamp) AS hour,
    campaignId,
    count() AS conversions,
    sum(value) AS revenue
FROM analytics.conversions
GROUP BY hour, campaignId;
