-- Monti Jarvis ClickHouse audit event projection (idempotent)
CREATE TABLE IF NOT EXISTS audit_events (
  event_id String,
  occurred_at DateTime64(3, 'UTC'),
  tenant_id String,
  actor_id String,
  actor_type LowCardinality(String),
  action LowCardinality(String),
  resource_type LowCardinality(String),
  resource_id String,
  request_id String,
  source LowCardinality(String),
  outcome LowCardinality(String),
  metadata_json String,
  ingested_at DateTime64(3, 'UTC') DEFAULT now64(3)
) ENGINE = ReplacingMergeTree(ingested_at)
ORDER BY (tenant_id, occurred_at, event_id);
