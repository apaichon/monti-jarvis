---
id: DES-0034
title: Platform Billing, Quota, and AI Infrastructure Cost Usage Specification
status: approved
updated: 2026-07-17
sprint: SPRINT-031
owner: SA
depends_on: [SPRINT-010, SPRINT-013, SPRINT-025, SPRINT-030]
---

# Platform Billing, Quota, and AI Infrastructure Cost Usage - Design Spec

**Sprint:** SPRINT-031 · **Release target:** v2.12.0 (proposed)  
**Feature:** [FEAT-0033](../01-features/FEAT-0033-platform-billing-quota-ai-cost-usage.md)  
**Roadmap:** Platform monitoring for billing, quota usage, and AI infrastructure cost usage  
**Depends on:** [15-commerce-chain-plan.md](15-commerce-chain-plan.md), [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md), [28-call-center-statistics-spec.md](28-call-center-statistics-spec.md), [33-platform-call-center-statistics-spec.md](33-platform-call-center-statistics-spec.md)

## 1. Goals

- Give platform administrators one read-only operational view of paid package value, entitlement/quota usage, and AI infrastructure usage by tenant and date range.
- Reconcile reporting metrics against the existing authorities: `payment_orders`, `tenant_entitlements`, `package_limits`, Redis enforcement counters, and ClickHouse call-center facts.
- Add an idempotent AI usage meter that can distinguish provider-observed usage from duration-based estimates and unavailable usage.
- Keep reporting outside customer call, voice relay, archive, quota enforcement, payment callback, and chat critical paths.
- Make every monetary or unit total traceable to a source, rate version, period, and freshness state without exposing customer data.

## 2. Non-goals

- Charging tenants, collecting payment, changing entitlements, auto-upgrading packages, invoice generation, refunds, or tax documents.
- Replacing Redis quota enforcement, package entitlement resolution, or the existing billing ledger.
- Customer-facing cost display, customer-level attribution, raw prompts, transcripts, audio, provider credentials, or provider request/response payloads.
- Historical forecasting, alerts, scheduled reports, exports, or a general-purpose data warehouse.
- Treating estimates as invoices or silently converting missing provider usage into zero cost.

## 3. Design decision

Use the existing operational stores as authorities and add an append-only ClickHouse usage projection for AI metering. The platform usage API is a read model assembled at request time from:

```text
Postgres payment_orders / tenant_entitlements / packages
                 │
                 ├── billing and entitlement summaries
                 │
ClickHouse call_center_usage_facts ──┐
                                     ├── platform usage read model
ClickHouse ai_cost_usage_facts ─────┘
                 │
Redis quota counters ── current enforcement snapshot only
```

The new projection is reporting-only. It never authorizes a call, increments a quota, mutates an entitlement, or changes a payment order.

## 4. Source-of-truth contract

| Concern | Authority | Reporting use | Failure behavior |
| --- | --- | --- | --- |
| Paid package value | `callcenter.payment_orders` with paid status | Paid orders, paid amount, currency, date | Ledger section unavailable; do not infer payment from an entitlement |
| Active package | `callcenter.tenant_entitlements` joined to `packages` | Package name, billing period, entitlement status | Tenant package shown as unavailable |
| Package ceilings | `package_limits.rules` / entitlement `rules_snapshot` | Read-only limit labels | Do not expose raw rules JSON; limit is unavailable if parsing fails |
| Live quota enforcement | Redis DB 4, `monti_jarvis:` keys | Current counter snapshot | Show `enforcement_unavailable`; never rebuild or write counters |
| Completed activity | `monti_jarvis.call_center_usage_facts` | Conversation count, duration, minutes | `analytics_unavailable` for activity totals |
| AI usage | `monti_jarvis.ai_cost_usage_facts` | Observed/estimated provider units and cost | Mark each metric `unavailable` when no safe usage fact exists |

The date range applies to reporting facts and paid-order creation dates. Current quota enforcement is a point-in-time snapshot and is explicitly labeled separately from historical range usage.

## 5. AI usage measurement

The current text Gemini client returns generated text but discards `usageMetadata`; the live relay does not expose a normalized billing usage event. Sprint 31 must add a provider-neutral meter boundary before presenting AI cost totals.

### 5.1 Measurement states

| State | Meaning | Can contribute to cost total? |
| --- | --- | --- |
| `observed` | Provider returned normalized input/output tokens or audio units. | Yes, with the matching rate version. |
| `estimated` | No provider units were returned; a documented duration or configured fallback was used. | Yes, but shown separately and never presented as exact. |
| `unavailable` | No safe unit measurement or applicable rate exists. | No; preserve the missing state. |

The aggregate response must return observed and estimated amounts separately. A total cost may be shown only as `observed_cost + estimated_cost`, with a coverage percentage and measurement state visible to the operator.

### 5.2 Meter boundary

Each completed text or voice interaction emits one logical usage event after the interaction is committed. The event contains no prompt, response, transcript, audio, customer id, email, or phone number.

```go
type AIUsageEvent struct {
    EventID              string // deterministic provider + interaction id
    TenantID             string
    CallID               string
    ConversationRecordID string
    Provider             string // gemini
    Model                string
    Modality             string // text | voice
    MeasurementState     string // observed | estimated | unavailable
    InputTokens          uint64
    OutputTokens         uint64
    AudioSeconds         uint32
    UsageDate            time.Time
    RateVersion          string
    SourceUpdatedAt      time.Time
}
```

For retries, the same `EventID` replaces the prior projection row. A failed analytics write does not fail the customer interaction; the source interaction remains authoritative and a replay path records the gap.

## 6. ClickHouse projection

Create `monti_jarvis.ai_cost_usage_facts` as an append/replacing projection. The projection is not a quota ledger and does not replace `call_center_usage_facts`.

| Column | Type | Notes |
| --- | --- | --- |
| `fact_id` | `String` | Deterministic `aic_` plus interaction/event id; idempotency key. |
| `tenant_id` | `String` | Required reporting scope. |
| `call_id` | `String` | Call/session reference only; never exposed in aggregate responses. |
| `conversation_record_id` | `String` | Source reference for replay/reconciliation. |
| `provider` | `LowCardinality(String)` | `gemini`. |
| `model` | `LowCardinality(String)` | Model identifier used for rate lookup. |
| `modality` | `LowCardinality(String)` | `text` or `voice`. |
| `measurement_state` | `LowCardinality(String)` | `observed`, `estimated`, or `unavailable`. |
| `input_units` | `UInt64` | Token or provider input unit count. |
| `output_units` | `UInt64` | Token or provider output unit count. |
| `audio_seconds` | `UInt32` | Voice duration when available. |
| `rate_version` | `String` | Immutable pricing catalog version, empty when unavailable. |
| `cost_microunits` | `Int64` | Cost in one-millionth of the currency unit; zero when unavailable. |
| `currency` | `FixedString(3)` | ISO currency, normally `USD` for provider cost. |
| `usage_date` | `Date` | Deployment/tenant reporting date bucket. |
| `source_updated_at` | `DateTime` | Source event timestamp used for replay ordering. |
| `created_at` / `updated_at` | `DateTime` | Projection timestamps and replacing version. |

Recommended engine and order:

```sql
CREATE TABLE IF NOT EXISTS monti_jarvis.ai_cost_usage_facts (
  fact_id String,
  tenant_id String,
  call_id String,
  conversation_record_id String,
  provider LowCardinality(String),
  model LowCardinality(String),
  modality LowCardinality(String),
  measurement_state LowCardinality(String),
  input_units UInt64,
  output_units UInt64,
  audio_seconds UInt32,
  rate_version String,
  cost_microunits Int64,
  currency FixedString(3),
  usage_date Date,
  source_updated_at DateTime,
  created_at DateTime DEFAULT now(),
  updated_at DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY (tenant_id, usage_date, provider, model, fact_id);
```

All aggregate queries use `FINAL` or an equivalent latest-row strategy and group by tenant, provider, model, and measurement state. Queries never return event-level identifiers to the platform UI.

## 7. Rate catalog

Rates must be versioned and effective-dated; they must not be hard-coded in a handler or silently changed for historical records. The initial catalog may be a controlled Postgres table or an explicit deployment configuration, but the API must return the `rate_version` and `pricing_as_of` metadata.

Minimum rate fields:

| Field | Contract |
| --- | --- |
| `rate_version` | Immutable identifier. |
| `provider`, `model`, `modality` | Exact lookup dimensions. |
| `input_unit_price_microunits` | Price per one million input units. |
| `output_unit_price_microunits` | Price per one million output units. |
| `audio_second_price_microunits` | Optional voice duration rate. |
| `currency` | ISO currency. |
| `effective_from`, `effective_until` | Non-overlapping validity window. |
| `status` | `active` or `retired`. |

If a rate lookup misses, the event remains `unavailable`; it is never priced at zero and never dropped.

## 8. Reconciliation model

The platform page presents three separate reconciliations:

1. **Activity vs quota:** ClickHouse range minutes compared with the current Redis monthly/daily enforcement counters. Differences are expected when the range and current period differ; the UI labels the periods instead of declaring a failure.
2. **Paid orders vs entitlements:** paid order amount and status compared with active entitlement package/status. A mismatch is an operator warning, not an automatic entitlement mutation.
3. **AI coverage:** measured interactions compared with eligible completed call-center facts. Observed, estimated, and unavailable counts are shown independently.

Reconciliation status values are `ok`, `warning`, `unavailable`, and `not_comparable`. The API must include a short safe reason code, never raw SQL, Redis keys, provider responses, or payment callback payloads.

## 9. API contract

### 9.1 Usage summary

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| `GET` | `/api/platform/billing/usage` | `platform_admin` | Read platform billing, quota, and AI usage summaries. |

Query fields:

| Field | Type | Default | Validation |
| --- | --- | --- | --- |
| `start_date` | date | today | `YYYY-MM-DD`, valid calendar date. |
| `end_date` | date | today | Valid and on/after start; bounded to 366 days. |
| `tenant_id` | string | empty | Exact optional filter, maximum 128 characters. |
| `limit` | integer | 50 | `1..100`. |
| `offset` | integer | 0 | `0..1000000`. |

Response shape:

```json
{
  "range": {"start_date": "2026-07-17", "end_date": "2026-07-17", "timezone": "Asia/Bangkok"},
  "freshness": {"status": "current", "generated_at": "...", "activity_last_projected_at": "..."},
  "billing": {
    "paid_orders": 0,
    "paid_amount_minor": 0,
    "currency": "THB",
    "status": "current"
  },
  "quota": {
    "reporting_minutes": 0,
    "enforcement": {"status": "current", "monthly_used": 0, "monthly_limit": 0, "daily_used": 0, "daily_limit": 0}
  },
  "ai_cost": {
    "currency": "USD",
    "observed_cost_microunits": 0,
    "estimated_cost_microunits": 0,
    "observed_events": 0,
    "estimated_events": 0,
    "unavailable_events": 0,
    "coverage_percent": 0,
    "status": "current"
  },
  "reconciliation": {"activity_quota": "not_comparable", "orders_entitlements": "ok", "ai_coverage": "warning"},
  "tenants": [],
  "pagination": {"total": 0, "limit": 50, "offset": 0}
}
```

Tenant rows contain only `tenant_id`, safe tenant name/slug, package name/status, reporting minutes, quota status, AI cost observed/estimated totals, coverage, and reconciliation codes. No customer or event-level fields are returned.

### 9.2 Existing billing ledger

`GET /api/platform/billing/orders` remains the order ledger authority. Sprint 31 may add date filters or a summary link, but must not change paid-status transitions, payment callbacks, receipt generation, or entitlement mutation semantics.

### 9.3 Errors

| HTTP | Code | Meaning |
| ---: | --- | --- |
| 400 | `validation_error` | Invalid range, filter, limit, or offset. |
| 401 | `unauthorized` / `session_expired` | Missing or expired platform session. |
| 403 | `forbidden` | Caller is not a platform administrator. |
| 503 | `usage_unavailable` | No safe aggregate can be produced from one or more required sources. |
| 500 | `billing_usage_unavailable` | Unexpected safe-response failure. |

Partial source failures should return `200` with per-section status when the platform can still provide a truthful response. Raw provider, SQL, Redis, and payment details are server logs only.

## 10. Workflow and performance bounds

1. Normalize the date range once in the deployment timezone and return it in the response.
2. Complete the customer interaction first; meter and project usage asynchronously or through a bounded post-commit path.
3. Capture text Gemini `usageMetadata` at the client boundary. For voice, use provider units when available and duration-based estimates only when the estimate formula and rate are configured.
4. Write usage facts idempotently to ClickHouse. Retry/replay uses the same event id and never duplicates a logical interaction.
5. Read ClickHouse aggregates in grouped queries; do not issue one analytics query per tenant row.
6. Read Postgres order/entitlement summaries in bounded aggregate queries and Redis counters only for the requested tenant page or a bounded current snapshot.
7. Apply a bounded context deadline to every source. A source outage must not block customer traffic or mutate enforcement state.
8. Do not create an unbounded goroutine per tenant, persist dashboard snapshots, or introduce a public replay endpoint.

## 11. Authorization and privacy

- Apply the existing platform-admin guard before all source queries.
- Treat `tenant_id` as a filter, never as an authorization grant.
- Return allowlisted aggregate response types only.
- Exclude customer ids, emails, phone numbers, prompts, responses, transcripts, audio paths, ticket notes, provider URLs, tokens, API keys, Redis keys, SQL, package rules JSON, and payment callback payloads.
- Keep AI cost in provider-cost currency and label it as reporting/estimated; never expose it as a tenant invoice balance.
- Read-only dashboard views do not create new audit events unless the existing platform audit policy explicitly requires it.

## 12. UX/UI contract

Add a Usage view under the existing platform billing area at `/admin/billing/usage` or as a tab within `/admin/billing`.

```text
+----------------------------------------------------------------+
| Billing & Usage        [Orders] [Usage]                        |
| [Start date] [End date] [Tenant] [Apply] [Today]               |
+----------------------------------------------------------------+
| Paid package value | Reporting minutes | Quota status | AI cost |
| THB ...            | ...                | ...           | USD ... |
+----------------------------------------------------------------+
| AI coverage: observed ...% | estimated ... | unavailable ...  |
| Reconciliation: activity/quota | orders/entitlements | AI      |
+----------------------------------------------------------------+
| Tenant        Package       Paid   Usage/quota   AI cost  State |
| ...                                                           |
+----------------------------------------------------------------+
| < Previous                                      Next >         |
+----------------------------------------------------------------+
```

Required states:

- Loading and retry without clearing the last successful range unnecessarily.
- Empty range distinct from source unavailable.
- Partial source failure displayed per section.
- Estimated AI cost visually distinct from observed cost.
- Session expiry redirects through the existing platform session flow.
- Responsive layout below 700px with stacked filters, readable tenant rows, and no horizontal metric overlap.

No invoice payment action, quota-edit action, customer drill-down, or raw event viewer is included.

## 13. Verification and acceptance criteria

```bash
make test
make build
cd apps/platform-admin-web && npm run check && npm run build
git diff --check
```

1. Duplicate meter delivery produces one logical AI usage fact.
2. Text Gemini usage metadata is captured when returned; missing metadata is `estimated` or `unavailable`, never silently zero.
3. Platform totals reconcile with controlled ClickHouse activity fixtures and existing tenant statistics.
4. Paid order totals come only from paid `payment_orders`; entitlements are read-only enrichment.
5. Redis quota counters are read-only and clearly separated from historical reporting minutes.
6. Rate changes do not rewrite historical facts; each priced event retains its rate version.
7. Tenant pagination, platform RBAC, redaction, date bounds, empty, stale, unavailable, and partial-failure states are covered.
8. Existing customer call, voice relay, archive, payment callback, quota, audit, and tenant statistics paths remain compatible.
9. Manual UAT includes two tenants, observed/estimated/unavailable AI usage, paid/unpaid orders, entitlement mismatch, quota divergence, and responsive billing usage UI.

## 14. Proposed task mapping

| Task | Points | Owner | Outcome |
| --- | ---: | --- | --- |
| TASK-0140 | 5 | devops | Metering contract, ClickHouse AI usage projection, rate-version model, and replay/idempotency path. |
| TASK-0141 | 4 | dev | Platform billing/quota/AI usage API with bounded aggregates, RBAC, redaction, and safe partial failures. |
| TASK-0142 | 5 | dev | Gemini text/voice usage instrumentation and responsive platform billing usage dashboard. |
| TASK-0143 | 2 | tester | Reconciliation fixtures, quota/payment compatibility checks, failure-state verification, and manual UAT. |

This mapping is proposed for Sprint 31 and remains uncommitted until the sprint commitment is approved.

## 15. Risks

| Risk | Mitigation |
| --- | --- |
| Gemini Live does not expose billable units consistently | Preserve observed/estimated/unavailable states and never present estimates as invoices. |
| Usage events are duplicated during retries | Deterministic event ids, `ReplacingMergeTree`, `FINAL`, and duplicate-delivery tests. |
| Historical costs change after a rate update | Store immutable rate version and computed micro-units on every fact. |
| Billing ledger and entitlements disagree | Show a reconciliation warning; do not mutate either source from the dashboard. |
| Redis counters and date-range facts differ | Label point-in-time enforcement versus historical reporting and avoid false equality assertions. |
| Cost instrumentation adds latency to calls | Post-commit bounded emission; analytics failure cannot fail the customer path. |

See [02-workflow.md](02-workflow.md) Sprint 31, [03-er-diagram.md](03-er-diagram.md) Sprint 31, [04-api-spec.md](04-api-spec.md) Platform Billing Usage, and [05-ux-ui.md](05-ux-ui.md) Sprint 31.
