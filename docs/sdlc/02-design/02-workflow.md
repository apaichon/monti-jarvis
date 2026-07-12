---
id: DES-0002
title: Workflows
status: approved
updated: 2026-07-11
sprint: SPRINT-014
---

# Workflows — Monti Jarvis

## 1. Portal load

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go :8091
  participant W as workforce

  B->>G: GET /
  G-->>B: Svelte SPA
  B->>G: GET /api/workforce
  G->>W: All()
  G-->>B: agents[]
  B->>G: GET /api/infra
  G-->>B: postgres, redis, minio, clickhouse
```

## 2. Text chat (with RAG)

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go server
  participant R as rag
  participant CH as ClickHouse
  participant AI as Gemini text

  B->>G: POST /api/chat {agent_id, topic, message, history}
  G->>R: Retrieve(agent, topic, message)
  R->>CH: SearchScoped (embed query)
  CH-->>R: top-k chunks
  R-->>G: sources + context block
  G->>AI: Reply(augmented prompt, history)
  AI-->>G: reply text
  G->>G: SaveExchange (Postgres)
  G-->>B: {reply, sources[], missing_km}
```

## 3. Voice call

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go server
  participant R as rag
  participant L as Gemini Live
  participant P as Postgres

  B->>G: POST /api/calls
  G->>P: CreateCallSession
  G-->>B: session {id, room_name}
  B->>G: WS /ws/voice?agent=&topic=
  G->>R: Retrieve(agent, topic, "") preload KB
  G->>L: setup(systemInstruction + KB)
  loop conversation
    B->>G: audio frames (JSON)
    G->>L: realtimeInput
    L-->>G: audio + transcript
    G-->>B: transcript events
    G->>P: addTurn (optional)
  end
  B->>G: POST /api/calls/{id}/end
```

## 4. KM ingest (per avatar)

```mermaid
sequenceDiagram
  participant Op as Operator/curl
  participant G as Go server
  participant M as MinIO
  participant P as Postgres
  participant AI as Gemini embed
  participant CH as ClickHouse

  Op->>G: POST /api/km/agents/{id}/documents (multipart)
  G->>M: Put km/demo/{agent}/...
  G->>P: knowledge_documents row
  G->>G: ChunkText
  loop each chunk
    G->>AI: Embed(chunk)
    AI-->>G: vector
  end
  G->>P: knowledge_chunks rows
  G->>CH: km_embeddings upsert
  G-->>Op: document {status: indexed, chunk_count}
```

## 5. KM reset (per avatar)

```mermaid
sequenceDiagram
  participant Op as Operator
  participant G as Go server
  participant P as Postgres
  participant M as MinIO
  participant CH as ClickHouse

  Op->>G: POST /api/km/agents/{id}/reset
  G->>P: DELETE documents + chunks
  G->>M: DELETE object keys
  G->>CH: DELETE embeddings for agent
  G-->>Op: {status: reset}
```

## 6. Call events (SSE)

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go server
  participant N as NATS

  B->>G: GET /api/calls/{id}/events (SSE)
  Note over G,N: turn persisted
  G->>N: call.turn.created (optional)
  G-->>B: event: turn {role, content}
```

## 6. Auth login (Sprint 3 — draft)

```mermaid
sequenceDiagram
  participant C as Client (curl/admin)
  participant G as Go :8091
  participant DB as Postgres
  participant A as internal/auth

  C->>G: POST /api/auth/login {email, password}
  G->>DB: lookup user + role
  G->>A: verify bcrypt
  A->>A: issue access JWT + refresh opaque
  G->>DB: insert refresh_tokens (hash)
  G-->>C: {access_token, refresh_token, user}
```

## 7. Protected KM upload (auth enabled)

```mermaid
sequenceDiagram
  participant C as Client
  participant G as Go server
  participant M as auth middleware
  participant KM as internal/km

  C->>G: POST /api/km/.../documents + Bearer
  G->>M: validate JWT, check tenant_admin
  alt forbidden
    M-->>C: 403
  else ok
    M->>KM: Ingest(tenant_id from context)
    KM-->>C: 201 document
  end
```

## 8. Dev bypass (`AUTH_DISABLED=true`)

No login required. All handlers use `tenant_id = DEMO_TENANT_ID`. Identical to v0.3.0 flows above.

## State: call session

| Status | Meaning |
| --- | --- |
| `active` | Call in progress; Redis key `monti_jarvis:call:active:{id}` |
| `ended` | `ended_at` set; Redis key removed |

## State: knowledge document

| Status | Meaning |
| --- | --- |
| `uploaded` | MinIO object stored |
| `indexing` | Chunk + embed in progress |
| `indexed` | Postgres + ClickHouse ready |
| `failed` | Embed or index error |

## 9. Package catalog CRUD (Sprint 4)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant G as Go :8091
  participant M as auth middleware
  participant P as internal/packages
  participant DB as Postgres

  Op->>G: POST /api/platform/packages {slug, name, limits, ...}
  G->>M: validate JWT + platform_admin
  alt forbidden
    M-->>Op: 403
  else ok
    M->>P: Create(ctx, input)
    P->>DB: validate rules vs package_rule_schemas
    P->>DB: INSERT packages + package_limits (rules jsonb)
    G-->>Op: 201 {id, slug, limits, ...}
  end
```

## 10. Assign tenant entitlement (Sprint 4)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant G as Go :8091
  participant E as internal/entitlements
  participant DB as Postgres
  participant R as Redis

  Op->>G: POST /api/platform/tenants/demo/entitlement {package_id}
  G->>G: RBAC platform_admin
  G->>DB: revoke prior active row (if any)
  G->>DB: INSERT tenant_entitlements + rules_snapshot + rules_schema_id
  G->>R: DEL monti_jarvis:entitlement:demo
  G-->>Op: 200 effective entitlement JSON
```

## 11. Entitlement resolve + cache (Sprint 4)

```mermaid
sequenceDiagram
  participant C as Client (tenant_admin)
  participant G as Go :8091
  participant E as internal/entitlements
  participant R as Redis
  participant DB as Postgres

  C->>G: GET /api/entitlements/me + Bearer
  G->>G: tenant_id from JWT
  G->>E: Resolve(tenant_id)
  E->>R: GET monti_jarvis:entitlement:{tenant_id}
  alt cache hit
    R-->>E: cached JSON
  else cache miss
    E->>DB: tenant_entitlements JOIN packages JOIN package_limits (rules jsonb)
    E->>R: SETEX key TTL payload
  end
  E-->>G: effective limits
  G-->>C: 200 {tenant_id, package, limits, status}
```

## State: package (Sprint 4)

| Status | Meaning |
| --- | --- |
| `draft` | Not assignable; hidden from default list |
| `active` | Assignable to tenants |
| `archived` | No new assignments; existing entitlements honored until revoked |

## State: tenant entitlement (Sprint 4)

| Status | Meaning |
| --- | --- |
| `active` | Tenant receives package limits (at most one per tenant) |
| `suspended` | Limits withheld; row kept for audit |
| `revoked` | Operator ended entitlement; resolver returns fallback |
| `expired` | `valid_until` passed (Sprint 9+ subscriptions) |

## 12. Platform admin login (Sprint 4)

```mermaid
sequenceDiagram
  participant B as Browser /admin
  participant G as Go :8091
  participant A as internal/auth

  B->>G: POST /api/auth/login {email, password}
  G->>A: Login + issue tokens
  G-->>B: {access_token, refresh_token, user}
  B->>B: sessionStorage tokens
  B->>G: GET /admin/packages (SPA)
  G-->>B: platform-admin-web index.html
  B->>G: GET /api/platform/packages + Bearer
  G-->>B: packages[]
```

## 13. Platform admin logout (Sprint 4)

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go :8091

  B->>G: POST /api/auth/logout + Bearer
  G-->>B: 200
  B->>B: clear sessionStorage
  B->>B: navigate /admin/login
```

## 14. Avatar catalog CRUD (Sprint 5)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant G as Go :8091
  participant M as auth middleware
  participant S as internal/store avatars
  participant DB as Postgres

  Op->>G: POST /api/platform/avatars {slug, name, role, voice, ...}
  G->>M: validate JWT + platform_admin
  alt forbidden
    M-->>Op: 403
  else ok
    M->>S: CreateAvatar(ctx, input)
    S->>DB: INSERT ai_avatars (flags jsonb)
    G-->>Op: 201 {id, slug, name, status, ...}
  end
```

## 15. Assign tenant avatar (Sprint 5)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant G as Go :8091
  participant E as internal/entitlements
  participant S as internal/store avatars
  participant DB as Postgres

  Op->>G: POST /api/platform/tenants/demo/avatars {avatar_id}
  G->>G: RBAC platform_admin
  G->>E: GetEffective(demo) → rules.max_ai_employees
  G->>S: CountActiveAssignments(demo)
  alt at cap
    G-->>Op: 409 max_ai_employees exceeded
  else ok
    S->>DB: UPSERT tenant_avatar_assignments status=active
    G-->>Op: 200 {tenant_id, avatar, status}
  end
```

## 16. Workforce resolve (Sprint 5)

```mermaid
sequenceDiagram
  participant B as Browser (customer /)
  participant G as Go :8091
  participant W as internal/workforce
  participant S as internal/store avatars
  participant A as internal/auth

  B->>G: GET /api/workforce
  Note over B,G: Optional X-Tenant-Id or Bearer tenant
  G->>A: ResolveTenant(ctx, header, authDisabled, demo)
  G->>S: ListAssignedAvatars(tenant_id, active)
  alt has assignments
    S-->>G: ai_avatars + primary ai_avatar_voices row
    G->>W: map to Agent JSON (image_url → image, voice from priority 1)
  else no assignments
    G->>W: All() static fallback
  end
  G-->>B: 200 {agents: [...]}
```

## State: avatar (Sprint 5)

| Status | Meaning |
| --- | --- |
| `draft` | Not assignable; hidden from default platform list |
| `active` | Assignable to tenants; eligible for workforce when assigned |
| `archived` | No new assignments; existing assignments may be disabled by operator |

## State: tenant avatar assignment (Sprint 5)

| Status | Meaning |
| --- | --- |
| `active` | Avatar appears in tenant `/api/workforce` list |
| `disabled` | Assignment revoked; avatar hidden from tenant workforce |

## State: avatar voice profile (Sprint 5)

| Status | Meaning |
| --- | --- |
| `active` | Eligible for primary selection or failover (by `priority`) |
| `disabled` | Skipped by resolver; kept for audit / future enable |

**Failover order:** ascending `priority` among `active` rows for the same `avatar_id`. Sprint 21 applies this during live calls.

## 17. Customer portal agent pick (unchanged UI, Sprint 5 data)

Customer portal still calls `GET /api/workforce` on load. Sprint 5 only changes **data source** when tenant has assignments; UI components unchanged.

## 18. Tenant self-registration (Sprint 6)

```mermaid
sequenceDiagram
  participant B as Browser (/tenant/register)
  participant G as Go :8091
  participant R as Redis (rate limit)
  participant Reg as internal/tenantregister
  participant A as internal/auth
  participant DB as Postgres

  B->>G: POST /api/public/tenant/register {company_name, slug, email, password, display_name}
  G->>R: INCR monti_jarvis:register:ip:{ip}
  alt rate exceeded
    G-->>B: 429 Too many requests
  else ok
    G->>Reg: Register(ctx, input)
    Reg->>DB: BEGIN
    Reg->>DB: INSERT tenants status=pending_kyc
    Reg->>DB: INSERT brands (default)
    Reg->>DB: INSERT users + user_roles tenant_admin
    Reg->>DB: INSERT tenant_registrations status=submitted
    Reg->>DB: COMMIT
    G->>A: IssueTokenPair(new user)
    A->>DB: INSERT refresh_tokens
    G-->>B: 201 {tenant_id, registration_id, access_token, refresh_token, user}
  end
```

## 19. Registration validation errors (Sprint 6)

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go :8091
  participant Reg as internal/tenantregister
  participant DB as Postgres

  B->>G: POST /api/public/tenant/register
  alt invalid slug or password
    G-->>B: 400 {error}
  else slug taken
    Reg->>DB: SELECT tenants WHERE slug=?
    G-->>B: 409 slug already taken
  else email taken
    Reg->>DB: SELECT users WHERE email=?
    G-->>B: 409 email already registered
  else register disabled
    G-->>B: 503 registration disabled
  end
```

## 20. Platform list pending tenants (Sprint 6)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant G as Go :8091
  participant M as auth middleware
  participant S as internal/store tenants
  participant DB as Postgres

  Op->>G: GET /api/platform/tenants?status=pending_kyc
  G->>M: validate JWT + platform_admin
  alt forbidden
    M-->>Op: 403
  else ok
    S->>DB: SELECT tenants JOIN tenant_registrations ORDER BY created_at DESC
    G-->>Op: 200 {tenants[], total, limit, offset}
  end
```

## 21. Pending tenant login (Sprint 6)

```mermaid
sequenceDiagram
  participant B as Browser (/tenant)
  participant G as Go :8091
  participant A as internal/auth

  Note over B,G: After registration, tokens stored client-side (same as /admin)

  B->>G: GET /api/auth/me (Bearer)
  G->>A: Parse JWT → tenant_admin + tenant_id
  G-->>B: 200 {role, tenant_id, email}

  B->>G: POST /api/km/agents/ava/documents (pending_kyc)
  G-->>B: 403 tenant not active
```

## State: tenant (Sprint 6 extension)

| Status | Meaning |
| --- | --- |
| `pending_kyc` | Self-registered; login OK; KM writes blocked; awaits Sprint 7 approval |
| `active` | Production tenant (seeds, post-KYC) |
| `suspended` | Operator block |

## State: tenant_registration (Sprint 6)

| Status | Meaning |
| --- | --- |
| `submitted` | Signup complete; visible in platform tenant list |

Sprint 7 adds `approved`, `rejected`, reviewer metadata.

## 22. Platform review KYC package (Sprint 7)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant B as Browser (/admin)
  participant G as Go :8091
  participant M as auth middleware
  participant S as internal/store
  participant DB as Postgres

  Op->>B: Open /admin/tenants/{id}/kyc
  B->>G: GET /api/platform/tenants/{tenant_id}/kyc
  G->>M: validate JWT + platform_admin
  alt forbidden
    M-->>B: 403
  else ok
    S->>DB: SELECT tenants, tenant_registrations, tenant_kyc_profiles
    G-->>B: 200 {tenant, registration, kyc, photo_url, documents[]}
    B->>G: GET /api/assets/kyc/{tenant_id}/photo/...
    G-->>B: image bytes (preview)
  end
```

## 23. Approve KYC (Sprint 7)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant G as Go :8091
  participant S as internal/store
  participant DB as Postgres
  participant R as internal/resend

  Op->>G: POST /api/platform/tenants/{tenant_id}/kyc/approve
  G->>S: ApproveTenantKYC(ctx, tenant_id, reviewer_id)
  alt tenant not pending_kyc or kyc not submitted
    S-->>G: conflict
    G-->>Op: 409
  else ok
    S->>DB: BEGIN
    S->>DB: UPDATE tenants SET status=active
    S->>DB: UPDATE tenant_registrations SET status=approved, reviewed_*
    S->>DB: UPDATE tenant_kyc_profiles SET status=approved, reviewed_*
    S->>DB: COMMIT
    G->>R: SendKYCApprovedEmail(admin_email) (async, best-effort)
    G-->>Op: 200 {tenant_id, status: active, kyc_status: approved}
  end
```

## 24. Reject KYC (Sprint 7)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant G as Go :8091
  participant S as internal/store
  participant DB as Postgres
  participant R as internal/resend

  Op->>G: POST /api/platform/tenants/{tenant_id}/kyc/reject {reason}
  alt missing reason
    G-->>Op: 400
  else ok
    G->>S: RejectTenantKYC(ctx, tenant_id, reviewer_id, reason)
    alt kyc not submitted
      S-->>G: conflict
      G-->>Op: 409
    else ok
      S->>DB: BEGIN
      S->>DB: UPDATE tenant_registrations SET status=rejected, rejection_reason
      S->>DB: UPDATE tenant_kyc_profiles SET status=rejected, reviewed_*
      Note over DB: tenants.status stays pending_kyc
      S->>DB: COMMIT
      G->>R: SendKYCRejectedEmail(admin_email, reason)
      G-->>Op: 200 {tenant_id, registration_status: rejected}
    end
  end
```

## State: tenant_kyc_profiles (Sprint 6–7)

| Status | Meaning |
| --- | --- |
| `draft` | Tenant editing contact/photo/docs |
| `submitted` | Awaiting platform review |
| `approved` | Platform approved; tenant is `active` |
| `rejected` | Platform rejected; tenant may resubmit (stretch) |

## State: tenant_registration (Sprint 7)

| Status | Meaning |
| --- | --- |
| `submitted` | Signup complete |
| `approved` | KYC approved; tenant `active` |
| `rejected` | KYC rejected; `rejection_reason` set |

## 25. Platform configure ChillPay gateway (Sprint 8)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant B as Browser (/admin)
  participant G as Go :8091
  participant M as auth middleware
  participant S as internal/store
  participant DB as Postgres

  Op->>B: Open /admin/settings/payment
  B->>G: GET /api/platform/payment-gateway
  G->>M: validate JWT + platform_admin
  alt forbidden
    M-->>B: 403
  else ok
    S->>DB: SELECT payment_gateway_configs WHERE id=default
    G-->>B: 200 {provider, merchant_code, api_key_masked, md5_key_set, callback_url, ...}
    Op->>B: Edit merchant code, API key, MD5 key, base URL, return URL
    B->>G: PUT /api/platform/payment-gateway {provider: chillpay, ...}
    G->>S: UpsertPaymentGatewayConfig
    S->>DB: INSERT/UPDATE payment_gateway_configs
    G-->>B: 200 updated config (secrets masked)
    B-->>Op: FeedbackDialog success
  end
```

## 26. Test ChillPay connection (Sprint 8)

```mermaid
sequenceDiagram
  participant Op as Operator (platform_admin)
  participant G as Go :8091
  participant P as internal/payment/chillpay
  participant CP as ChillPay API

  Op->>G: POST /api/platform/payment-gateway/test
  alt provider = mock
    G-->>Op: 200 {ok: true, provider: mock}
  else provider = chillpay
    G->>P: Ping(ctx) — checksum self-test + optional InquiryPaymentStatus
    alt credentials invalid
      P-->>G: error
      G-->>Op: 502 {ok: false, message}
    else ok
      P->>CP: POST /api/v2/PaymentStatus/ (optional probe)
      G-->>Op: 200 {ok: true}
    end
  end
```

## 27. ChillPay payment callback (Sprint 8)

```mermaid
sequenceDiagram
  participant CP as ChillPay
  participant G as Go :8091
  participant P as internal/payment/chillpay
  participant S as internal/store
  participant DB as Postgres

  CP->>G: POST /api/callbacks/chillpay (form-urlencoded)
  G->>G: Parse ChillPayCallbackForm
  alt PAYMENT_CALLBACK_DEV_BYPASS
    Note over G: skip verify (local only)
  else verify
    G->>P: VerifyCallback(form)
    alt bad CheckSum
      P-->>G: false
      G-->>CP: 400
    end
  end
  G->>S: InsertPaymentCallbackEvent(transaction_id, ...)
  alt duplicate transaction_id
    S->>DB: ON CONFLICT DO NOTHING
  else new
    S->>DB: INSERT payment_callback_events
  end
  Note over G: No entitlement update (Sprint 9)
  G-->>CP: 200
```

## State: payment_gateway_configs (Sprint 8)

| Status | Meaning |
| --- | --- |
| `inactive` | Not configured or disabled |
| `active` | Operator saved valid config; callbacks accepted |

## State: payment_callback_events (Sprint 8)

| payment_status | Meaning |
| --- | --- |
| `0` | Success — Sprint 9 will fulfill order |
| `1` | Pending |
| `2` | Failed |

## 28. Tenant browse packages + start checkout (Sprint 9)

```mermaid
sequenceDiagram
  participant U as Tenant admin
  participant B as Browser (/tenant/billing)
  participant G as Go :8091
  participant M as auth middleware
  participant S as internal/store
  participant DB as Postgres

  U->>B: Open /tenant/billing
  B->>G: GET /api/tenant/packages
  B->>G: GET /api/entitlements/me
  G->>M: JWT + tenant_admin + active tenant
  alt tenant not active
    M-->>B: 403
  else ok
    S->>DB: SELECT packages WHERE status=active
    G-->>B: 200 packages[] + current entitlement
    U->>B: Click Buy on Pro
    B->>G: POST /api/tenant/checkout {package_id}
    S->>DB: INSERT payment_orders pending
    G-->>B: 200 {order_id, payment_url}
    B->>B: window.location = payment_url
  end
```

## 29. ChillPay InitPayment redirect (Sprint 9)

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go :8091
  participant CP as internal/payment/chillpay
  participant CH as ChillPay API
  participant S as internal/store

  B->>G: POST /api/tenant/checkout
  G->>S: Load gateway config + package price
  G->>CP: InitPayment(order_no, amount, customer_id, ...)
  CP->>CH: POST form-urlencoded + CheckSum
  CH-->>CP: JSON {PaymentUrl, TransactionId}
  CP-->>G: payment_url, transaction_id
  G->>S: UPDATE payment_orders SET payment_url, transaction_id
  G-->>B: {payment_url}
  B->>CH: Redirect to PaymentUrl (hosted pay page)
  Note over CH: Customer completes payment on ChillPay
  CH->>G: POST /api/callbacks/chillpay (async)
  CH-->>B: Redirect to ReturnUrl (/tenant/billing/return)
```

## 30. Callback fulfillment + entitlement (Sprint 9)

```mermaid
sequenceDiagram
  participant CP as ChillPay
  participant G as Go :8091
  participant P as internal/payment
  participant S as internal/store
  participant E as internal/entitlements
  participant DB as Postgres
  participant R as Redis

  CP->>G: POST /api/callbacks/chillpay (PaymentStatus=0)
  G->>P: VerifyCallback(form)
  G->>S: InsertPaymentCallbackEvent (idempotent)
  G->>S: GetPaymentOrderByOrderNo(OrderNo)
  alt order already paid
    G-->>CP: 200 ack
  else PaymentStatus = 0
    S->>DB: BEGIN
    S->>DB: UPDATE payment_orders SET status=paid, paid_at
    S->>DB: UPDATE tenant_entitlements SET status=revoked WHERE active
    S->>DB: INSERT tenant_entitlements active (package snapshot)
    S->>DB: COMMIT
    G->>E: InvalidateCache(tenant_id)
    E->>R: DEL monti_jarvis:entitlement:{tenant_id}
    G-->>CP: 200
  else PaymentStatus = 2
    S->>DB: UPDATE payment_orders SET status=failed
    G-->>CP: 200
  end
```

## 31. Mock checkout path (Sprint 9)

```mermaid
sequenceDiagram
  participant U as Tenant admin
  participant B as Browser
  participant G as Go :8091
  participant S as internal/store

  U->>B: POST /api/tenant/checkout (gateway provider=mock)
  G-->>B: payment_url = /tenant/billing/mock-pay?order_id=...
  B->>B: Open mock-pay page
  alt PAYMENT_MOCK_AUTO_FULFILL
    B->>G: POST /api/dev/mock-pay/{order_id}
  else manual
    U->>B: Click Complete mock payment
    B->>G: POST /api/dev/mock-pay/{order_id}
  end
  G->>S: FulfillOrder (same as callback success path)
  G-->>B: 200 {status: paid}
  B->>G: GET /api/entitlements/me
  G-->>B: new package entitlement
```

## State: payment_orders (Sprint 9)

| Status | Meaning |
| --- | --- |
| `pending` | Awaiting ChillPay callback |
| `paid` | Success; entitlement assigned |
| `failed` | ChillPay reported failure |
| `cancelled` | Abandoned (future ops) |

## 32. Quota check on KM ingest (Sprint 13)

```mermaid
sequenceDiagram
  participant U as Operator / tenant
  participant G as Go :8091
  participant E as internal/entitlements
  participant Q as internal/quota
  participant R as Redis
  participant DB as Postgres
  participant KM as internal/km

  U->>G: POST /api/km/agents/{id}/documents
  G->>Q: AllowRate(tenant, km)
  alt rate limited
    Q-->>G: ErrRateLimited
    G-->>U: 429 rate_limited
  end
  G->>E: GetEffective(tenant)
  E-->>G: rules incl max_km_documents
  G->>Q: CheckKMDocument(tenant)
  Q->>DB: COUNT km documents
  alt usage >= limit
    Q-->>G: ErrLimitExceeded
    G-->>U: 429 quota_exceeded
  else under limit
    G->>KM: Ingest(...)
    KM-->>G: doc
    G-->>U: 201
  end
```

## 33. Concurrent voice slot (Sprint 13)

```mermaid
sequenceDiagram
  participant C as Customer browser
  participant G as Go :8091
  participant Q as internal/quota
  participant R as Redis
  participant V as voice / LiveKit relay

  C->>G: GET /ws/voice (or session start)
  G->>Q: Check voice_enabled + AcquireConcurrent(tenant)
  Q->>R: INCR concurrent
  alt over max_concurrent_calls
    Q->>R: DECR
    Q-->>G: ErrLimitExceeded
    G-->>C: 429 / WS close
  else ok
    G->>V: open session
    Note over C,V: call in progress
    C-->>G: disconnect
    G->>Q: release concurrent
    Q->>R: DECR + optional AddCallMinutes
  end
```

## 34. Platform usage snapshot (Sprint 13)

```mermaid
sequenceDiagram
  participant A as Platform admin
  participant B as Browser /admin
  participant G as Go :8091
  participant Q as internal/quota
  participant E as internal/entitlements
  participant R as Redis
  participant DB as Postgres

  A->>B: Open tenant usage (P14)
  B->>G: GET /api/platform/tenants/{id}/usage
  alt not platform_admin
    G-->>B: 403
  end
  G->>E: GetEffective(tenant)
  G->>Q: Snapshot(tenant)
  Q->>R: GET concurrent, minutes
  Q->>DB: COUNT km docs, avatar assignments
  Q-->>G: limits + usage
  G-->>B: 200 JSON
  B-->>A: bars / table
```

## 35. Chat rate limit + RAG flag (Sprint 13)

```mermaid
sequenceDiagram
  participant C as Customer browser
  participant G as Go :8091
  participant Q as internal/quota
  participant E as internal/entitlements
  participant R as Redis
  participant Chat as chat / RAG

  C->>G: POST /api/chat
  Note over G: tenant = JWT or DEMO_TENANT_ID if AUTH_DISABLED
  G->>Q: AllowRate(tenant, chat)
  Q->>R: INCR rl:chat:minute
  alt count > RATE_LIMIT_CHAT_PER_MIN
    Q-->>G: ErrRateLimited
    G-->>C: 429 rate_limited + Retry-After
  end
  alt request uses RAG
    G->>Q: CheckFeature(rag_enabled)
    alt rag_enabled false
      Q-->>G: ErrFeatureDisabled
      G-->>C: 403 feature_disabled
    end
  end
  G->>Chat: handle message
  Chat-->>C: 200 answer
```

## 36. Avatar assign vs max_ai_employees (Sprint 13)

```mermaid
sequenceDiagram
  participant A as Platform admin
  participant B as Browser /admin
  participant G as Go :8091
  participant Q as internal/quota
  participant DB as Postgres

  A->>B: Assign avatar to tenant
  B->>G: POST /api/platform/tenants/{id}/avatars
  G->>Q: CheckAIEmployees(tenant, count+1)
  Q->>DB: COUNT active assignments
  alt count+1 > max_ai_employees
    Q-->>G: ErrLimitExceeded
    G-->>B: 429 quota_exceeded
  else ok
    G->>DB: INSERT assignment
    G-->>B: 200/201
  end
```

## State notes — quota counters (Sprint 13)

| Key / counter | Lifecycle |
| --- | --- |
| `concurrent` | +1 acquire · −1 release / disconnect · TTL safety |
| `minutes:YYYYMM` | +elapsed on session end · resets by key month |
| `rl:*:minute` | INCR per request · expires ~2m |
| KM / avatar usage | Derived from Postgres (not Redis primary) |

## 37. Tenant configures embed (Sprint 14)

```mermaid
sequenceDiagram
  participant A as Tenant admin
  participant B as Browser /tenant/embed
  participant G as Go :8091
  participant DB as Postgres

  A->>B: Open Embed settings
  B->>G: GET /api/tenant/embed
  alt no row
    G->>DB: INSERT tenant_embed_configs (key, enabled=false)
  end
  G-->>B: config + embed_key
  A->>B: Enable + set allowed_origins + Save
  B->>G: PUT /api/tenant/embed
  G->>DB: UPDATE
  G-->>B: 200
  A->>B: Copy snippet
```

## 38. Host page loads widget (Sprint 14)

```mermaid
sequenceDiagram
  participant V as Visitor browser
  participant H as Tenant website
  participant L as monti-embed.js
  participant G as Go :8091
  participant E as Embed UI /embed
  participant Q as internal/quota

  V->>H: Open shop page
  H->>L: Load script data-embed-key
  L->>L: Inject launcher + iframe src=/embed?key=
  E->>G: GET /api/public/embed/{key} Origin host
  alt disabled or unknown
    G-->>E: 404
  else origin not allowed
    G-->>E: 403 origin_not_allowed
  else ok
    G-->>E: tenant_id, agents, default_agent
    V->>E: Select agent + chat
    E->>G: POST /api/chat X-Tenant-Id
    G->>Q: AllowRate + checks
    G-->>E: reply
  end
```

## 39. Rotate embed key (Sprint 14)

```mermaid
sequenceDiagram
  participant A as Tenant admin
  participant G as Go :8091
  participant DB as Postgres

  A->>G: POST /api/tenant/embed/rotate-key
  G->>DB: UPDATE embed_key = new
  G-->>A: new key
  Note over A: Old key returns 404 on public resolve
```

See [06-auth-spec.md](06-auth-spec.md), [08-packages-spec.md](08-packages-spec.md), [10-avatars-spec.md](10-avatars-spec.md), [11-tenant-register-spec.md](11-tenant-register-spec.md), [12-kyc-tenant-spec.md](12-kyc-tenant-spec.md), [13-payment-gateway-spec.md](13-payment-gateway-spec.md), [14-buy-package-spec.md](14-buy-package-spec.md), [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md), [17-embed-to-web-spec.md](17-embed-to-web-spec.md), [09-platform-admin-portal-spec.md](09-platform-admin-portal-spec.md), [04-api-spec.md](04-api-spec.md), [05-ux-ui.md](05-ux-ui.md).