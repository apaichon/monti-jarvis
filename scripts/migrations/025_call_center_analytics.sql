-- Sprint 25: tenant call-center usage projection.
-- Runtime bootstrap also applies this shape through ClickHouse EnsureSchema.
CREATE TABLE IF NOT EXISTS monti_jarvis.call_center_usage_facts (
  fact_id String,
  tenant_id String,
  call_id String,
  conversation_record_id String,
  avatar_id String,
  channel String,
  source String,
  status String,
  started_at DateTime,
  ended_at DateTime,
  usage_date Date,
  duration_seconds UInt32,
  source_updated_at DateTime,
  created_at DateTime DEFAULT now(),
  updated_at DateTime DEFAULT now(),
  created_by String DEFAULT 'system',
  updated_by String DEFAULT 'system'
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (tenant_id, usage_date, call_id, fact_id);
