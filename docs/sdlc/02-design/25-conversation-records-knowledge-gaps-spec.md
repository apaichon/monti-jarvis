---
id: DES-0025
title: Conversation Records and Knowledge Gaps Specification
status: approved
updated: 2026-07-13
sprint: SPRINT-022
owner: SA
---

# Conversation Records and Knowledge Gaps — Design Spec

**Sprint:** SPRINT-022 · **Release target:** v2.3.0  
**Feature:** [FEAT-0024](../01-features/FEAT-0024-conversation-records-knowledge-gaps.md)  
**Depends on:** [18-tenant-scope-km-spec.md](18-tenant-scope-km-spec.md), [23-customer-auth-spec.md](23-customer-auth-spec.md), [24-authenticated-workforce-selection-spec.md](24-authenticated-workforce-selection-spec.md)

## 1. Goals

- Persist tenant-scoped conversation record metadata.
- Store transcript and available call artifacts in MinIO under `calls/`.
- Support documented archive protection/encryption mode without exposing object secrets.
- Create knowledge-gap candidates from RAG misses, low confidence, fallback answers, or tenant flags.
- Give tenant admins a review lifecycle for open, snoozed, resolved, and ignored gaps.

## 2. Non-goals

- Ticket routing, SLA management, or human support queues.
- Satisfaction surveys and analytics dashboards.
- Long-term retention automation and cold archive lifecycle.
- New ClickHouse dashboard tables in this sprint.

## 3. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `CONVERSATION_ARCHIVE_ENABLED` | `true` | Enables archive writer |
| `CONVERSATION_ARCHIVE_PROTECTION_MODE` | `none` | `none`, `sse-s3`, `sse-kms`, or `client` depending on deployment support |
| `CONVERSATION_ARCHIVE_RETRY_LIMIT` | `3` | Max retry attempts for failed archive writes |
| `KNOWLEDGE_GAP_MIN_CONFIDENCE` | `0.35` | Below this, create a gap candidate when other signals match |

## 4. Data model

### `conversation_records`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `crec_{ulid}` |
| `tenant_id` | text FK | Tenant isolation |
| `call_id` | text FK | Existing call/chat session id when present |
| `customer_id` | text FK | Nullable for anonymous optional-auth use |
| `avatar_id` | text FK | Selected AI workforce |
| `channel` | text | `chat` or `voice` |
| `status` | text | `recording`, `archived`, `archive_failed` |
| `started_at`, `ended_at` | timestamptz | Conversation window |
| `duration_seconds` | int | Voice/chat duration where applicable |
| `summary` | jsonb | Bounded safe metadata |
| audit cols | | `created_at`, `updated_at`, `created_by`, `updated_by` |

### `conversation_archive_objects`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `cobj_{ulid}` |
| `tenant_id` | text FK | Tenant isolation |
| `conversation_record_id` | text FK | Parent record |
| `object_key` | text | `calls/{tenant_id}/{call_id}/...`; no email/PII in path |
| `object_type` | text | `transcript`, `audio`, `metadata` |
| `content_type` | text | MIME type |
| `size_bytes` | bigint | Object size |
| `checksum_sha256` | text | Integrity check |
| `protection_mode` | text | Archive protection mode used |
| `status` | text | `stored`, `failed`, `deleted` |
| `error_code` | text | Safe failure code |
| `stored_at` | timestamptz | Write time |
| audit cols | | `created_at`, `updated_at`, `created_by`, `updated_by` |

### `knowledge_gap_candidates`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | `kgap_{ulid}` |
| `tenant_id` | text FK | Tenant isolation |
| `conversation_record_id` | text FK | Source conversation |
| `avatar_id` | text FK | Selected AI workforce |
| `customer_id` | text FK | Nullable |
| `source_turn_id` | text | Turn identifier from transcript |
| `question` | text | Bounded question text |
| `answer_excerpt` | text | Safe excerpt, not full transcript |
| `gap_reason` | text | `no_source`, `low_confidence`, `fallback`, `tenant_flag` |
| `confidence` | numeric | Model/RAG confidence where available |
| `status` | text | `open`, `snoozed`, `resolved`, `ignored` |
| `reviewer_note` | text | Tenant note |
| `snoozed_until`, `resolved_at` | timestamptz | Lifecycle timestamps |
| audit cols | | `created_at`, `updated_at`, `created_by`, `updated_by` |

## 5. MinIO / Redis / ClickHouse

| Store | Contract |
| --- | --- |
| MinIO | Bucket `monti-jarvis`, prefix `calls/{tenant_id}/{call_id}/` |
| Redis | Optional retry/worker lock key `monti_jarvis:archive_retry:{conversation_record_id}` |
| ClickHouse | No new required table; later dashboards may aggregate Postgres records |

## 6. API summary

See [04-api-spec.md](04-api-spec.md) § Conversation Records & Knowledge Gaps.

| Method | Path | Role |
| --- | --- | --- |
| `GET` | `/api/tenant/conversation-records` | `tenant_admin` |
| `GET` | `/api/tenant/conversation-records/{id}` | `tenant_admin` |
| `POST` | `/api/tenant/conversation-records/{id}/archive/retry` | `tenant_admin` |
| `GET` | `/api/tenant/knowledge-gaps` | `tenant_admin` |
| `PATCH` | `/api/tenant/knowledge-gaps/{id}` | `tenant_admin` |

## 7. RBAC

| Action | customer | tenant_admin | platform_admin |
| --- | --- | --- | --- |
| Create archive from live conversation | via server only | no | no |
| List tenant records | no | yes | platform support only if existing policy allows |
| Retry archive | no | yes | platform support only if existing policy allows |
| Review knowledge gaps | no | yes | platform support only if existing policy allows |
| Cross-tenant record/gap access | no | no | no by default |

## 8. Verification

```bash
make test && make build
curl -H "Authorization: Bearer $TENANT_TOKEN" \
  "http://localhost:8091/api/tenant/conversation-records"
curl -H "Authorization: Bearer $TENANT_TOKEN" \
  "http://localhost:8091/api/tenant/knowledge-gaps?status=open"
```

## Approver sign-off

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| PM | | | ☐ |
| Dev | | | ☐ |
