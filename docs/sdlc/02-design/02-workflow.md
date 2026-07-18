---
id: DES-0002
title: Workflows
status: shipped
updated: 2026-07-17
sprint: SPRINT-030
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

## 40. Tenant uploads KM document (Sprint 15)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant B as Browser /tenant/km
  participant G as Go :8091
  participant Auth as internal/auth
  participant Q as internal/quota
  participant K as internal/km
  participant P as Postgres
  participant M as MinIO
  participant C as ClickHouse
  participant Emb as Gemini embed

  A->>B: Select agent + scope + file
  B->>G: POST /api/tenant/km/agents/{agent_id}/documents multipart Bearer
  G->>Auth: RequireTenantAdminActive
  alt unauthorized or inactive
    G-->>B: 401/403
  else ok
    G->>Q: AllowRate(BucketKM) + CheckKMDocument
    alt quota exceeded
      G-->>B: 429/403 quota code
    else allowed
      G->>K: Ingest(jwt.tenant_id, agent, file, scope)
      K->>M: PUT km/{tenant}/{agent}/{doc}/original/{file}
      K->>P: INSERT knowledge_documents status=uploaded
      K->>P: status=indexing
      K->>Emb: embed chunks
      K->>P: ReplaceKnowledgeChunks
      K->>C: INSERT km_embeddings
      K->>P: status=indexed chunk_count=N
      G-->>B: 201 document JSON
    end
  end
```

## 41. Tenant deletes KM document (Sprint 15)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant G as Go :8091
  participant K as internal/km
  participant P as Postgres
  participant M as MinIO
  participant C as ClickHouse

  A->>G: DELETE /api/tenant/km/documents/{id} Bearer
  G->>K: DeleteDocument(jwt.tenant_id, id)
  K->>P: GetKnowledgeDocument(id)
  alt missing or tenant_id mismatch
    G-->>A: 404 document not found
  else ok
    K->>C: DELETE km_embeddings WHERE tenant_id AND document_id
    K->>M: Delete object_key
    K->>P: DELETE knowledge_chunks CASCADE / document
    G-->>A: 200 { deleted: true, id }
  end
```

## 42. Tenant lists agents and documents (Sprint 15)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant G as Go :8091
  participant P as Postgres

  A->>G: GET /api/tenant/km/agents Bearer
  G->>P: tenant workforce agents + COUNT docs by agent
  G-->>A: { agents: [{id,name,doc_count,by_scope}] }
  A->>G: GET /api/tenant/km/agents/{agent_id}/documents
  G->>P: ListKnowledgeDocuments(tenant, agent)
  G-->>A: { agent_id, documents: [...] }
  A->>G: GET /api/tenant/km/scopes
  G-->>A: { scopes: [general,billing,technical] }
```

## 43. Tenant updates document scope (Sprint 15)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant G as Go :8091
  participant K as internal/km
  participant P as Postgres
  participant C as ClickHouse

  A->>G: PATCH /api/tenant/km/documents/{id} {km_scope}
  G->>K: UpdateDocumentScope(tenant, id, scope)
  alt invalid scope
    G-->>A: 400
  else not found
    G-->>A: 404
  else ok
    Note over K: Prefer re-tag chunks + CH without full re-embed when vector text unchanged; if implementation re-ingests, document it
    K->>P: UPDATE knowledge_documents.km_scope + chunks.km_scope
    K->>C: UPDATE/rewrite km_scope on embeddings for document_id
    G-->>A: 200 document JSON
  end
```

## 44. Tenant resets agent knowledge (Sprint 15)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant G as Go :8091
  participant K as internal/km
  participant P as Postgres
  participant M as MinIO
  participant C as ClickHouse

  A->>G: POST /api/tenant/km/agents/{agent_id}/reset Bearer
  G->>K: ResetAgent(jwt.tenant_id, agent_id)
  K->>P: list object_keys for tenant+agent
  K->>C: DELETE embeddings tenant+agent
  K->>M: delete objects
  K->>P: DeleteAgentKnowledge
  G-->>A: 200 { agent_id, status: reset }
```

### Document status state (S2 + S15)

| From | To | Trigger |
| --- | --- | --- |
| — | `uploaded` | Create after MinIO put |
| `uploaded` | `indexing` | Chunk/embed start |
| `indexing` | `indexed` | Embeddings + chunks ready (RAG) |
| `indexing` | `failed` | Embed/chunk error |
| any | *(deleted)* | DELETE document or agent reset |

UI labels: show **ready** for `indexed`; **processing** for `uploaded`/`indexing`; **failed** for `failed`.

## 45. Record knowledge gap on missing KM (Sprint 15)

```mermaid
sequenceDiagram
  actor C as Caller
  participant G as Go :8091
  participant R as internal/rag
  participant CH as ClickHouse qa_events
  participant P as Postgres km_gaps

  C->>G: POST /api/chat message
  G->>R: Retrieve(agent, topic, question)
  R-->>G: MissingKM=true, sources=[]
  G->>CH: InsertQAEvent missing_km
  G->>P: RecordKMGap upsert by question_hash
  Note over P: occurrence_count++ if same tenant+agent+hash
  G-->>C: reply + missing_km true
```

Tenant later: `GET /api/tenant/km/gaps` → write FAQ doc → `PATCH` gap `status=converted`.

## 46. Tenant updates settings (Sprint 16)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant B as Browser /tenant/settings
  participant G as Go :8091
  participant P as Postgres

  A->>B: Edit locale, timezone, Save
  B->>G: PUT /api/tenant/settings Bearer
  G->>G: RequireTenantAdminActive
  G->>P: UPSERT tenant_settings
  G-->>B: 200 settings
```

## 47. Tenant views usage (Sprint 16)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant G as Go :8091
  participant Q as internal/quota
  participant R as Redis
  participant P as Postgres

  A->>G: GET /api/tenant/usage
  G->>Q: Snapshot(jwt.tenant_id)
  Q->>R: monthly minutes, concurrent
  Q->>P: avatars, km counts
  G-->>A: package + limits + usage
```

## 48. Voice open with daily/per-call caps (Sprint 16)

```mermaid
sequenceDiagram
  actor C as Caller
  participant G as Go :8091
  participant Q as internal/quota
  participant L as tenant call limits
  participant R as Redis

  C->>G: WS /ws/voice
  G->>Q: S13 voice_enabled, concurrent, monthly
  G->>L: load tenant_call_limits
  G->>R: GET call_daily:{tenant}:{day}
  alt daily cap exceeded
    G-->>C: 403 daily_call_limit
  else ok
    G->>G: session start timer
    Note over G: if elapsed >= max_minutes_per_call → end
    G->>R: INCR daily on end
    G->>Q: AddCallMinutes monthly
  end
```

## 49. Tenant preview chat (Sprint 17)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant G as Go :8091
  participant Auth as RequireTenantAdminActive
  participant Q as internal/quota
  participant RAG as internal/rag
  participant AI as Gemini

  A->>G: POST /api/tenant/preview/chat
  G->>Auth: JWT tenant_admin + active
  G->>Q: AllowRate(chat) only
  Note over G,Q: Skip monthly minutes check / AddCallMinutes
  G->>RAG: WithTenant(jwt.tenant_id).Retrieve
  G->>AI: Reply + locale hint
  G-->>A: reply + sources + mode=preview
```

## 50. Tenant preview voice (Sprint 17)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant G as Go :8091
  participant Q as internal/quota
  participant R as Redis
  participant Live as Gemini Live

  A->>G: WS preview voice
  G->>Q: AllowRate(voice)
  G->>R: INCR preview:concurrent:{tenant}
  alt concurrent > PREVIEW_MAX
    G-->>A: 429 preview_concurrent
  else ok
    G->>Live: relay (tenant RAG + locale)
    Note over G: No daily/monthly minute accounting
    G->>R: DECR concurrent on close
  end
```

## 51. Preview vs production quota (Sprint 17)

```mermaid
flowchart LR
  subgraph production
    P1[embed / customer chat-voice] --> Q1[S13 rate + package minutes]
    Q1 --> Q2[S16 daily/per-call]
  end
  subgraph preview
    V1[tenant /preview] --> R1[S13 rate only]
    R1 --> R2[preview concurrent soft cap]
  end
```

## 52. Create customer tier (Sprint 18)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant G as Go :8091
  participant Auth as RequireTenantAdminActive
  participant P as Postgres

  A->>G: POST /api/tenant/tiers
  G->>Auth: JWT tenant_admin + active
  G->>G: validate slug, caps ≥ 0, locale
  G->>P: INSERT customer_tiers
  G-->>A: 201 tier
```

## 53. Preview with tier_id (Sprint 18)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant G as Go :8091
  participant T as customer_tiers
  participant L as call limits resolve
  participant AI as Gemini

  A->>G: POST /api/tenant/preview/chat {tier_id}
  G->>T: load tier by id + jwt.tenant
  G->>L: effective caps = tier override or tenant defaults
  G->>AI: system prompt + tier locale if set
  G-->>A: reply mode=preview
```

## 54. Future: customer assigned to tier (S19+)

```mermaid
flowchart LR
  C[customer account] --> T[customer_tiers]
  C --> G[customer_groups]
  T --> R[limits + agent + locale]
```

## 55. Tenant creates or updates a customer (Sprint 19)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant B as Browser /tenant/customers
  participant G as Go :8091
  participant Auth as internal/auth
  participant S as internal/store
  participant P as Postgres callcenter

  A->>B: Save customer + tier/group ids
  B->>G: POST/PUT /api/tenant/customers Bearer
  G->>Auth: RequireTenantAdminActive
  alt missing, inactive, or wrong role
    Auth-->>B: 401/403
  else authorized
    G->>S: Validate tenant tier/groups + normalize identity
    S->>P: UPSERT customer + replace group memberships
    alt resource belongs to another tenant
      S-->>B: 404 not_found
    else duplicate without idempotent match
      S-->>B: 409 conflict
    else saved
      G-->>B: 200/201 customer
    end
  end
```

## 56. CSV dry-run then commit (Sprint 19)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant B as Browser /tenant/customers
  participant G as Go :8091
  participant I as internal/customerimport
  participant S as internal/store
  participant P as Postgres callcenter

  A->>B: Select UTF-8 CSV
  B->>G: POST /api/tenant/customer-imports dry_run=true
  G->>I: Parse with byte/row limits
  I->>S: Resolve tier/group/domain refs by JWT tenant
  I-->>B: 200 validated + accepted/rejected row summary
  A->>B: Confirm import
  B->>G: POST same CSV dry_run=false
  G->>I: Parse + validate again
  alt parser/file failure
    I-->>B: 400 import_invalid
  else valid rows
    I->>P: transaction: upsert customers + memberships + import job
    P-->>I: created/updated counts
    I-->>B: 201 completed import summary
  end
```

### Customer import state

| Status | Meaning |
| --- | --- |
| `validating` | CSV parsing and reference validation in progress |
| `validated` | Dry-run complete; no customer writes |
| `completed` | Commit transaction completed |
| `failed` | Parser or transaction failure |

## 57. Domain default resolution (Sprint 19)

```mermaid
sequenceDiagram
  participant G as Go :8091
  participant I as internal/customerimport
  participant D as customer_domain_rules
  participant T as customer_tiers/groups
  participant P as Postgres

  G->>I: Upsert row {email, tier_slug?, group_slugs?}
  I->>D: Find active normalized email domain in JWT tenant
  I->>T: Resolve explicit and default assignments
  alt explicit assignment provided
    I->>P: persist explicit tier/groups
  else matching allow/deny rule has defaults
    Note over I: policy stored for S20; not an auth decision in S19
    I->>P: persist domain default tier/group
  else no defaults
    I->>P: persist unassigned customer
  end
```

## 58. Integration-safe upsert event (Sprint 19)

```mermaid
sequenceDiagram
  participant API as Customer API/import
  participant S as internal/store
  participant P as Postgres
  participant N as NATS optional

  API->>S: Upsert(tenant, source, external_id)
  S->>P: match source/external_id then normalized email fallback
  P-->>S: stable customer_id
  opt NATS configured
    S->>N: monti.customer.upserted
    Note over S,N: publish failure is non-fatal; no external webhook delivery
  end
  S-->>API: customer + created/updated outcome
```

## 59. Tenant enables customer auth (Sprint 20)

```mermaid
sequenceDiagram
  actor A as TenantAdmin
  participant B as Browser /tenant/settings
  participant G as Go :8091
  participant Auth as internal/auth
  participant S as internal/store
  participant P as Postgres callcenter

  A->>B: Enable customer auth + choose domain mode
  B->>G: PUT /api/tenant/customer-auth/settings Bearer
  G->>Auth: RequireTenantAdminActive
  alt missing, inactive, or wrong role
    Auth-->>B: 401/403
  else authorized
    G->>S: Validate mode disabled|optional|required
    S->>P: UPSERT tenant_customer_auth_settings
    G-->>B: 200 settings
  end
```

## 60. Customer requests OTP by email (Sprint 20)

```mermaid
sequenceDiagram
  actor C as Customer
  participant B as Browser /
  participant G as Go :8091
  participant CA as internal/customerauth
  participant Mail as internal/mailer
  participant D as customer_domain_rules
  participant P as Postgres callcenter

  C->>B: Enter email to sign in or claim account
  B->>G: POST /api/customer/auth/request-otp
  G->>CA: Resolve tenant from host/embed key/body-safe context
  CA->>P: Load tenant_customer_auth_settings
  alt auth disabled
    CA-->>B: 403 customer_auth_disabled
  else enabled
    CA->>D: Match normalized email domain
    alt denied domain
      CA-->>B: 403 domain_denied
    else required allow and no allow rule
      CA-->>B: 403 domain_not_allowed
    else allowed
      CA->>P: Find or create tenant-scoped customer
      CA->>P: INSERT customer_otp_challenges otp_hash
      CA->>Mail: Send short OTP to customer email
      alt mail delivery fails
        CA-->>B: 502 otp_delivery_failed
      else sent
        CA-->>B: 202 challenge_id + masked email + expires_in
      end
    end
  end
```

## 61. Customer verifies OTP and refreshes session (Sprint 20)

```mermaid
sequenceDiagram
  actor C as Customer
  participant B as Browser /
  participant G as Go :8091
  participant CA as internal/customerauth
  participant R as Redis
  participant P as Postgres callcenter

  C->>B: Submit OTP from email
  B->>G: POST /api/customer/auth/verify-otp
  G->>CA: Load challenge by challenge_id
  CA->>P: Check pending challenge, expiry, attempts
  alt expired, locked, or wrong code
    CA-->>B: 401 otp_invalid_or_expired
  else ok
    CA->>P: UPSERT customer_auth_identities
    CA->>P: INSERT customer_sessions
    CA->>R: SET customer_session:{sid} TTL
    CA-->>B: access_token + refresh_token + customer
  end
  B->>G: POST /api/customer/auth/refresh
  G->>CA: Validate refresh token + session
  CA-->>B: rotated tokens
```

### Customer auth states

| Entity | Status | Meaning |
| --- | --- | --- |
| `tenant_customer_auth_settings` | `disabled` | Public no-auth portal only; customer auth endpoints return disabled |
| `tenant_customer_auth_settings` | `optional` | Public no-auth portal remains; customers may sign in |
| `tenant_customer_auth_settings` | `required` | Customer auth required for customer-specific account context |
| `customer_auth_identities` | `active` | Email identity can request OTP/login |
| `customer_auth_identities` | `locked` | OTP login blocked after security/manual action |
| `customer_otp_challenges` | `pending` | OTP sent and awaiting verification |
| `customer_otp_challenges` | `verified` | OTP accepted and exchanged for a session |
| `customer_otp_challenges` | `expired` | Past OTP expiry |
| `customer_otp_challenges` | `locked` | Too many wrong OTP attempts |
| `customer_sessions` | `active` | Access/refresh pair valid until expiry |
| `customer_sessions` | `revoked` | Logout or security revocation |
| `customer_sessions` | `expired` | Past expiry; refresh rejected |

## 62. Authenticated chat and quota attribution (Sprint 20)

```mermaid
sequenceDiagram
  actor C as Customer
  participant B as Browser /
  participant G as Go :8091
  participant Auth as internal/customerauth
  participant Q as internal/quota
  participant S as internal/store
  participant RAG as internal/rag
  participant AI as Gemini

  C->>B: Ask question while signed in
  B->>G: POST /api/chat Bearer customer token
  G->>Auth: Validate customer session
  Auth->>S: Load tenant customer + tier/groups
  G->>Q: Apply tenant package + rate + tier/group context
  alt quota or rate exceeded
    Q-->>B: 429/403 quota error
  else allowed
    G->>RAG: Retrieve with tenant + customer context
    G->>AI: Prompt with tier/group/locale hints
    G-->>B: reply + sources + customer_context
  end
```

## 63. Multi-tenant customer auth UAT (Sprint 20)

```mermaid
flowchart LR
  A1[Tenant A customer login] --> A2[Session tenant=A]
  B1[Tenant B customer login] --> B2[Session tenant=B]
  A2 --> A3[Chat consumes A quota keys]
  B2 --> B3[Chat consumes B quota keys]
  A2 -. cross-tenant id .-> X[404 or 403]
  B2 -. cross-tenant id .-> X
```

## 64. Required-auth workforce selection (Sprint 21)

```mermaid
sequenceDiagram
  actor C as Customer
  participant B as Browser
  participant G as Go :8091
  participant Auth as internal/auth
  participant W as internal/workforce
  participant S as internal/store
  participant R as Redis
  participant P as Postgres callcenter

  C->>B: Open tenant portal with tenant_id
  B->>G: GET /api/customer/portal-policy?tenant_id=...
  G->>S: Load tenant auth + workforce policy
  G-->>B: auth_required + quota summary + selected tenant
  alt auth required and no customer token
    B-->>C: Show OTP gate; disable workforce picker/start call
  else optional or signed in
    B->>G: GET /api/customer/workforce?tenant_id=...
    G->>Auth: Validate optional customer Bearer
    G->>W: List active tenant-assigned avatars
    G->>R: Read customer/tenant quota counters
    W->>P: Load tenant avatar assignments
    G-->>B: workforce[] + selected_avatar + quota_state
  end
```

## 65. Customer quota and call duration enforcement (Sprint 21)

```mermaid
sequenceDiagram
  actor C as Customer
  participant B as Browser
  participant G as Go :8091
  participant Auth as internal/auth
  participant Q as internal/quota
  participant Calls as internal/calls
  participant R as Redis
  participant P as Postgres callcenter

  C->>B: Start chat or voice with selected avatar
  B->>G: POST /api/chat or POST /api/calls Bearer customer token
  G->>Auth: Resolve tenant/customer/session when present
  G->>Q: Check tenant package + customer daily + per-call limits
  Q->>R: INCR/READ monti_jarvis:quota:{tenant}:{customer}:...
  alt tenant requires auth and token missing
    G-->>B: 401 customer_auth_required
  else daily limit exhausted
    G-->>B: 429 customer_quota_exhausted + reset_at
  else per-call duration invalid/exceeded
    G-->>B: 403 call_duration_limit_exceeded
  else allowed
    Calls->>P: Create tenant/customer/avatar call or chat record
    G-->>B: call/chat response + quota_remaining
  end
```

### Customer workforce/quota states

| State | Meaning |
| --- | --- |
| `auth_optional` | Tenant allows anonymous use; customer can sign in for account-aware quota |
| `auth_required` | Customer must verify OTP before workforce selection or call/chat start |
| `quota_available` | Customer can start chat/voice for selected avatar |
| `quota_exhausted` | Daily customer call/chat time is depleted until reset |
| `avatar_unavailable` | Avatar is disabled, unassigned, or not allowed for tenant/customer context |

## 66. Conversation archive write (Sprint 22)

```mermaid
sequenceDiagram
  participant G as Go :8091
  participant Calls as internal/calls
  participant Arch as internal/archive
  participant KM as internal/km
  participant M as MinIO monti-jarvis/calls
  participant P as Postgres callcenter

  G->>Calls: Chat/voice turn completed
  Calls->>Arch: Build transcript/artifact envelope
  Arch->>P: INSERT conversation_records status=recording
  Arch->>M: PUT calls/{tenant_id}/{call_id}/transcript.json
  alt object write fails
    Arch->>P: UPDATE archive status=failed retry metadata
  else object stored
    Arch->>P: INSERT conversation_archive_objects status=stored
    Arch->>P: UPDATE conversation_records status=archived
  end
  Calls->>KM: Evaluate RAG/fallback confidence
```

## 67. Knowledge-gap review lifecycle (Sprint 22)

```mermaid
sequenceDiagram
  participant RAG as internal/rag
  participant Gaps as internal/km
  participant P as Postgres callcenter
  actor T as Tenant admin
  participant B as Browser
  participant G as Go :8091

  RAG->>Gaps: Low confidence, fallback answer, or no source
  Gaps->>P: UPSERT knowledge_gap_candidates status=open
  T->>B: Open Knowledge Gaps
  B->>G: GET /api/tenant/knowledge-gaps
  G->>P: Query tenant-scoped gaps
  G-->>B: gaps[]
  T->>B: Resolve, snooze, or ignore
  B->>G: PATCH /api/tenant/knowledge-gaps/{id}
  alt cross-tenant id
    G-->>B: 404 not_found
  else ok
    G->>P: Update status + reviewer notes
    G-->>B: updated gap
  end
```

### Conversation archive and gap states

| Entity | Status | Meaning |
| --- | --- | --- |
| `conversation_records` | `recording` | Conversation is active or archive not complete |
| `conversation_records` | `archived` | Required transcript/artifact metadata is stored |
| `conversation_records` | `archive_failed` | Archive write failed and can be retried |
| `conversation_archive_objects` | `stored` | MinIO object exists and checksum/size were recorded |
| `conversation_archive_objects` | `deleted` | Object was removed by retention/admin action |
| `knowledge_gap_candidates` | `open` | Needs tenant review |
| `knowledge_gap_candidates` | `snoozed` | Hidden until review date |
| `knowledge_gap_candidates` | `resolved` | Tenant addressed the gap or linked KM work |
| `knowledge_gap_candidates` | `ignored` | Tenant decided no action is needed |

## 68. Customer confirms human follow-up ticket (Sprint 23)

```mermaid
sequenceDiagram
  actor C as Customer
  participant B as Browser / voice client
  participant G as Go :8091
  participant T as internal/tickets
  participant A as internal/auth
  participant P as Postgres callcenter
  participant R as Redis
  participant N as NATS

  T->>B: ticket_offer {subject, category, source_call_id}
  C->>B: Confirm human follow-up
  B->>G: POST /api/customer/tickets {confirm_escalation:true}
  G->>A: Resolve tenant/customer policy and session
  alt auth required and no customer session
    A-->>G: customer_auth_required
    G-->>B: 401 customer_auth_required
  else anonymous request has no contact email
    G-->>B: 400 contact_required
  else confirmed
    G->>R: SETNX ticket idempotency key with TTL
    alt duplicate confirmation
      R-->>G: existing ticket id
      G-->>B: 200 existing ticket
    else first confirmation
      G->>P: Validate source call/record in tenant scope
      G->>P: INSERT tickets + ticket_events(created, customer_confirmed)
      G->>N: ticket.created {tenant_id, ticket_id, source_id, status}
      G-->>B: 201 ticket {id, status: open}
    end
  end
```

### Ticket offer states

| State | Meaning |
| --- | --- |
| `offered` | Runtime asked the customer whether human follow-up is wanted; no ticket exists |
| `confirmed` | Customer explicitly confirmed; create request may proceed |
| `declined` | Customer declined or dismissed the offer; conversation continues normally |
| `created` | Ticket and first event were persisted |

## 69. Tenant triages ticket queue (Sprint 23)

```mermaid
sequenceDiagram
  actor T as Tenant admin
  participant B as Browser
  participant G as Go :8091
  participant S as internal/tickets
  participant P as Postgres callcenter
  participant N as NATS

  T->>B: Open Tickets
  B->>G: GET /api/tenant/tickets?status=open&filters...
  G->>S: Resolve tenant from JWT and validate filters
  S->>P: SELECT tickets WHERE tenant_id = token.tenant_id
  P-->>S: ticket queue rows
  S-->>G: safe queue rows
  G-->>B: 200 tickets[]
  T->>B: Open ticket detail
  B->>G: GET /api/tenant/tickets/{id}
  G->>P: Load ticket + source record + events in tenant scope
  G-->>B: 200 ticket detail
  T->>B: Change status, priority, assignee, or add note
  B->>G: PATCH /api/tenant/tickets/{id} or POST /events
  alt invalid transition or cross-tenant id
    G-->>B: 400 validation_error or 404 not_found
  else valid update
    G->>P: UPDATE ticket + INSERT ticket_events
    G->>N: ticket.updated {tenant_id, ticket_id, status}
    G-->>B: 200 updated ticket
  end
```

## 70. Ticket lifecycle (Sprint 23)

```mermaid
stateDiagram-v2
  [*] --> open: customer confirms
  open --> in_progress: tenant starts work
  open --> closed: tenant closes without work
  in_progress --> waiting_customer: more information needed
  waiting_customer --> in_progress: tenant resumes
  in_progress --> resolved: follow-up completed
  in_progress --> closed: tenant closes
  resolved --> closed: final close
  resolved --> in_progress: reopen for correction
```

See [06-auth-spec.md](06-auth-spec.md), [08-packages-spec.md](08-packages-spec.md), [10-avatars-spec.md](10-avatars-spec.md), [11-tenant-register-spec.md](11-tenant-register-spec.md), [12-kyc-tenant-spec.md](12-kyc-tenant-spec.md), [13-payment-gateway-spec.md](13-payment-gateway-spec.md), [14-buy-package-spec.md](14-buy-package-spec.md), [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md), [17-embed-to-web-spec.md](17-embed-to-web-spec.md), [18-tenant-scope-km-spec.md](18-tenant-scope-km-spec.md), [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md), [20-tenant-test-preview-spec.md](20-tenant-test-preview-spec.md), [21-customer-tier-spec.md](21-customer-tier-spec.md), [22-customer-account-import-spec.md](22-customer-account-import-spec.md), [23-customer-auth-spec.md](23-customer-auth-spec.md), [24-authenticated-workforce-selection-spec.md](24-authenticated-workforce-selection-spec.md), [25-conversation-records-knowledge-gaps-spec.md](25-conversation-records-knowledge-gaps-spec.md), [09-platform-admin-portal-spec.md](09-platform-admin-portal-spec.md), [04-api-spec.md](04-api-spec.md), [05-ux-ui.md](05-ux-ui.md).

## 71. Customer submits satisfaction review (Sprint 24)

```mermaid
sequenceDiagram
  actor C as Customer
  participant B as Browser / voice client
  participant G as Go :8091
  participant A as internal/auth
  participant S as internal/store
  participant P as Postgres callcenter

  G->>B: call/session ended
  B->>C: Show 1-5 star review prompt
  alt customer selects a star
    C->>B: Select score 1..5
    B->>G: POST /api/calls/{id}/rating {score}
    G->>A: Resolve call tenant and optional customer context
    alt invalid score or unknown/cross-tenant call
      G-->>B: 400 validation_error or 404 not_found
    else valid score
      G->>S: SaveConversationRating(tenant_id, call_id, score)
      S->>P: INSERT ... ON CONFLICT (tenant_id, call_id) DO UPDATE
      P-->>S: rating saved
      S-->>G: saved
      G-->>B: 201 {status: saved}
    end
  else customer skips or closes prompt
    B->>B: Keep call/archive closed; mark review as unrated locally
    B->>C: Follow-up prompt remains available without reopening call
  end
```

### Satisfaction review state

| State | Meaning |
| --- | --- |
| `unrated` | Conversation ended and no score has been saved |
| `prompted` | Customer was shown or told about the review |
| `submitted` | A valid 1-5 score is stored; repeated submission updates the same row |
| `skipped` | Customer dismissed the prompt; no rating row is required |

The prompt state is client/session UX state. Postgres stores only submitted ratings, so closing a call never waits for a review.

## 72. Tenant views satisfaction statistics (Sprint 24)

```mermaid
sequenceDiagram
  actor T as Tenant admin
  participant B as Browser
  participant G as Go :8091
  participant S as internal/store
  participant P as Postgres callcenter

  T->>B: Open Satisfaction
  B->>G: GET /api/tenant/satisfaction/statistics?start_date&end_date&avatar_id&channel
  G->>S: Resolve tenant from JWT and normalize date range
  alt missing/invalid tenant role or date range
    G-->>B: 403 forbidden or 400 validation_error
  else valid request
    S->>P: Aggregate completed conversations and ratings by tenant/date
    P-->>S: KPI, distribution, avatar and channel buckets
    alt no completed conversations in range
      S-->>G: Empty aggregate with zero counts
      G-->>B: 200 empty statistics
    else data available
      S-->>G: Tenant-scoped aggregate
      G-->>B: 200 statistics
    end
  end
```

See [27-customer-satisfaction-statistics-spec.md](27-customer-satisfaction-statistics-spec.md), [03-er-diagram.md](03-er-diagram.md), [04-api-spec.md](04-api-spec.md), and [05-ux-ui.md](05-ux-ui.md).

## 73. Project completed call facts (Sprint 25)

```mermaid
sequenceDiagram
  participant G as Go :8091
  participant A as internal/analytics
  participant P as Postgres callcenter
  participant C as internal/clickhouse
  participant H as ClickHouse monti_jarvis

  G->>P: Conversation close/archive update
  A->>P: Read conversation_records + call_sessions source
  alt ended_at is missing
    A-->>A: Ignore non-terminal source record
  else completed record
    A->>C: Build fact_id from conversation_record_id
    C->>H: INSERT call_center_usage_facts FORMAT JSONEachRow
    Note over C,H: ReplacingMergeTree(updated_at) makes retries logically idempotent
    alt ClickHouse unavailable
      H-->>C: connection or insert error
      C-->>A: retryable projection error
      A-->>G: Keep source close/archive successful; record freshness gap
    else projection stored
      H-->>C: insert accepted
      C-->>A: projected
    end
  end
```

The projection is a read-model side effect. It never blocks call closure, MinIO archive writes, rating submission, or Redis quota enforcement.

## 74. Tenant opens the call-center dashboard (Sprint 25)

```mermaid
sequenceDiagram
  actor T as Tenant admin
  participant B as Browser
  participant G as Go :8091
  participant A as internal/auth
  participant P as Postgres callcenter
  participant R as Redis monti_jarvis:
  participant C as internal/clickhouse
  participant H as ClickHouse monti_jarvis

  T->>B: Open /tenant/dashboard
  B->>G: GET /api/tenant/call-center/statistics?start_date&end_date
  G->>A: Resolve active tenant-admin context
  alt missing or invalid tenant-admin context
    A-->>G: unauthorized or forbidden
    G-->>B: 401/403 stable error
  else authorized
    G->>P: Resolve tenant timezone and current call limits
    G->>R: Read current package quota usage and limits
    G->>C: Query facts by tenant and inclusive usage_date range
    C->>H: SELECT FINAL aggregate by channel and avatar
    H-->>C: KPI and breakdown rows
    C-->>G: Tenant-scoped activity aggregate
    alt ClickHouse unavailable
      C-->>G: analytics unavailable
      G-->>B: 503 analytics_unavailable
    else query succeeds
      G-->>B: 200 statistics + quota + freshness
      B-->>T: Render KPI, breakdowns, filters, and state
    end
  end
```

The activity range applies to ClickHouse facts. The quota block reports the current package period and today's operational cap, so a historical range cannot be mistaken for the current enforcement counter.

## 75. Operator replays a bounded analytics range (Sprint 25)

```mermaid
sequenceDiagram
  actor O as Operator
  participant A as internal/analytics replay job
  participant P as Postgres callcenter
  participant C as internal/clickhouse
  participant H as ClickHouse monti_jarvis

  O->>A: Run replay for tenant/date range
  A->>P: Read ended conversation_records and call_sessions
  loop each source record
    A->>C: Build deterministic usage fact
    C->>H: INSERT same fact_id with source_updated_at
    H-->>C: Replace latest logical version during FINAL query
  end
  A-->>O: Report scanned, projected, skipped, and failed counts
```

Replay is an operator job and is intentionally not exposed as a tenant or customer HTTP endpoint.

### Analytics projection outcomes

| Outcome | Meaning | Dashboard behavior |
| --- | --- | --- |
| `not_eligible` | Source has no `ended_at` yet | Excluded from facts and counts |
| `projected` | Latest fact inserted into ClickHouse | Included after ClickHouse visibility delay |
| `retryable_error` | Source is complete but ClickHouse write failed | Source remains valid; dashboard may report stale freshness or `analytics_unavailable` |
| `replayed` | Existing source range was safely projected again | No logical double count |

## 76. Tenant opens system performance monitoring (Sprint 26)

```mermaid
sequenceDiagram
  actor T as Tenant admin
  participant B as Browser
  participant G as Go :8091
  participant A as internal/auth
  participant O as internal/observability
  participant P as Postgres callcenter
  participant R as Redis monti_jarvis:
  participant M as MinIO monti-jarvis
  participant C as internal/clickhouse
  participant H as ClickHouse monti_jarvis

  T->>B: Open /tenant/monitoring
  B->>G: GET /api/tenant/system-performance
  G->>A: Resolve active tenant-admin context
  alt missing, invalid, or inactive tenant context
    A-->>G: unauthorized or forbidden
    G-->>B: 401/403 safe error
  else authorized
    G->>O: Read or build bounded monitoring snapshot
    par dependency probes
      O->>P: Ping with shared deadline
      O->>R: PING with shared deadline
      O->>M: Check configured bucket with shared deadline
      O->>C: Ping with shared deadline
    and configured service states
      O->>O: Read NATS, LiveKit, and Gemini enabled state
    end
    O->>C: Read allowlisted analytics freshness only
    alt one or more probes timeout or fail
      O-->>G: Normalized degraded/unavailable snapshot
      G-->>B: 200 snapshot with safe statuses
    else snapshot succeeds
      O-->>G: Operational snapshot
      G-->>B: 200 snapshot with status and latency
    end
    B-->>T: Render status, latency, freshness, and retry action
  end
```

Probes use a shared bounded deadline and are isolated from call creation, voice relay, quota enforcement, archive writes, and `/healthz`. A dependency failure is represented in the monitoring response; it is not allowed to fail a live customer operation.

### Monitoring status lifecycle

| Status | Meaning | UI treatment |
| --- | --- | --- |
| `operational` | Configured probe completed successfully. | Green operational state. |
| `degraded` | Probe completed with a warning or analytics freshness is outside the current threshold. | Amber warning with retry guidance. |
| `unavailable` | Required configured probe failed or snapshot generation failed. | Red unavailable state; preserve last checked time. |
| `disabled` | Optional dependency is not configured. | Neutral disabled state; no latency. |
| `stale` | Analytics source is reachable but projection freshness is outside the threshold. | Amber stale label separate from dependency health. |

See [29-tenant-system-performance-spec.md](29-tenant-system-performance-spec.md), [03-er-diagram.md](03-er-diagram.md), [04-api-spec.md](04-api-spec.md), and [05-ux-ui.md](05-ux-ui.md).

## 77. Mobile SDK creates a call session (Sprint 27)

~~~mermaid
sequenceDiagram
  actor M as Mobile caller
  participant K as Mobile SDK
  participant G as Go :8091
  participant A as internal/auth
  participant W as internal/workforce
  participant Q as internal/quota
  participant S as Postgres callcenter
  participant R as Redis monti_jarvis:

  M->>K: getBootstrap()
  K->>G: GET /api/mobile/v1/bootstrap
  G-->>K: Auth policy, locale, limits, and active assigned avatars
  M->>K: createCall(avatar_id, locale)
  K->>G: POST /api/mobile/v1/calls with Bearer, Idempotency-Key
  G->>A: Resolve customer, tenant, and auth policy
  alt missing or expired customer session
    A-->>G: unauthorized or customer_auth_required
    G-->>K: 401/403 safe error
  else authorized or allowed public development flow
    G->>W: Validate active avatar assignment for resolved tenant
    alt avatar is not assigned or inactive
      W-->>G: avatar_not_available
      G-->>K: 404/409 safe error
    else assigned
      G->>Q: Check rate limit, concurrent calls, and remaining quota
      alt quota or rate limit denied
        Q-->>G: quota_exceeded or rate_limited
        G-->>K: 429 safe error
      else allowed
        G->>S: Insert call_sessions with tenant, customer, and avatar context
        G->>R: Save bounded idempotency response
        S-->>G: call_id and active session
        G-->>K: 201 call session and WebSocket path
      end
    end
  end
~~~

A repeated create request with the same idempotency key returns the original session response. A different key creates a new session only when concurrency and quota policy allow it.

## 78. Mobile SDK connects and exchanges voice events (Sprint 27)

~~~mermaid
sequenceDiagram
  actor M as Mobile caller
  participant K as Mobile SDK
  participant G as Go :8091
  participant A as internal/auth
  participant S as call session store
  participant L as internal/live
  participant X as Gemini Live

  K->>G: GET /ws/mobile/v1/calls/{call_id} with Bearer
  G->>A: Validate caller, tenant, session, and active status
  alt caller or call is not authorized
    A-->>G: unauthorized or forbidden
    G-->>K: Close with safe error envelope
  else active call
    G->>S: Verify call belongs to resolved tenant and caller
    G->>L: Start bounded mobile relay for call
    L-->>K: ready {call_id, avatar_id, encoding}
    loop audio and control frames
      K->>L: PCM16 audio or typed text
      L->>X: Provider-neutral relay input
      X-->>L: Audio and transcript output
      L-->>K: audio, transcript, turn_complete, call_status
    end
    alt retryable transport interruption
      K->>G: Reconnect with same call_id and Bearer
      G->>S: Re-check active session and caller
      G-->>K: ready with resumed lifecycle
    else customer or server end
      K->>L: end control frame
      L->>S: End call session idempotently
      L-->>K: call_status ended
    end
  end
~~~

The relay maps provider output to the versioned Monti event envelope. Provider errors, room names, credentials, and internal connection details stay server-side.

## 79. Mobile SDK ends and rates a call (Sprint 27)

~~~mermaid
sequenceDiagram
  actor M as Mobile caller
  participant K as Mobile SDK
  participant G as Go :8091
  participant S as Postgres callcenter

  M->>K: end(reason)
  K->>G: POST /api/mobile/v1/calls/{call_id}/end with Idempotency-Key
  G->>S: Verify tenant and caller ownership
  alt active session
    S-->>G: Update status ended and ended_at
    G-->>K: 200 {status: ended}
  else already ended
    S-->>G: Return existing ended state
    G-->>K: 200 idempotent response
  end
  K-->>M: Stop microphone, close socket, expose review state
  M->>K: Select 1-5 stars
  K->>G: POST /api/mobile/v1/calls/{call_id}/rating {score}
  G->>S: Upsert conversation_ratings by tenant and call
  S-->>G: Rating saved
  G-->>K: 201 {status: saved}
~~~

### Mobile call lifecycle

| State | Entry | Allowed next state | Client behavior |
| --- | --- | --- | --- |
| idle | No call handle | creating | Show avatar and Start call. |
| creating | Create request pending | connecting, failed | Disable duplicate start. |
| connecting | Session created, WebSocket pending | active, failed | Request microphone and connect. |
| active | Ready event received | ending, failed | Stream bounded audio frames and transcript events. |
| ending | End requested or timeout | ended, failed | Stop capture and wait for final status. |
| ended | Server confirms close | idle | Offer star review; never reconnect. |
| failed | Non-retryable error | idle, connecting | Map safe error and allow explicit retry. |

See 30-mobile-call-api-sdk-spec.md, 03-er-diagram.md, 04-api-spec.md, and 05-ux-ui.md.

## 80. Request emits an audit event (Sprint 28)

```mermaid
sequenceDiagram
  actor U as Platform or tenant user
  participant B as Browser
  participant G as Go :8091
  participant A as internal/auth + auditctx
  participant Q as internal/audit queue
  participant F as Backend local spool
  participant P as Postgres callcenter

  U->>B: Submit an authenticated mutation
  B->>G: POST/PUT/DELETE protected resource
  G->>A: Resolve tenant, actor, request_id, and authorization
  alt denied
    A-->>G: Safe forbidden result
    G->>Q: Enqueue redacted denied event when context is safe
    G-->>B: 403 safe error
  else authorized
    G->>P: Commit existing domain mutation
    P-->>G: Mutation result
    G->>Q: Enqueue immutable success event with stable event_id
    Q->>F: Append one bounded JSON line to active audit file
    alt local writer accepts event
      F-->>Q: Accepted locally
      G-->>B: Existing domain response
    else queue or writer unavailable
      F-->>Q: Write failure
      Q-->>G: Observable audit acceptance failure
      G-->>B: Existing route-specific safe failure policy
    end
  end
```

The domain mutation and audit event are not a distributed transaction. The local writer is the durability boundary in `spool` mode; ClickHouse is not called from the request handler. Event construction uses allowlisted metadata and strips secrets, tokens, OTPs, raw bodies, audio, transcripts, and unbounded personal data before enqueueing.

## 81. Audit worker transfers and retains local files (Sprint 28)

```mermaid
sequenceDiagram
  participant W as Go audit worker
  participant F as Backend local spool
  participant C as ClickHouse monti_jarvis
  participant M as Transfer marker

  loop every AUDIT_LOG_FLUSH_INTERVAL (default 5s)
    W->>F: Close/rotate active file and scan closed files oldest-first
    W->>F: Claim one file with exclusive lock
    W->>C: INSERT audit_events in bounded JSONEachRow batches
    alt complete insert acknowledged
      C-->>W: HTTP 200 success for all batches
      W->>M: Atomically write count, checksum, and acknowledged_at
      alt file age >= AUDIT_LOG_RETENTION (default 1h)
        W->>F: Delete file and marker
      else retention not reached
        W-->>F: Keep acknowledged file until retention threshold
      end
    else timeout, partial response, or ClickHouse unavailable
      C-->>W: Retryable or unknown result
      W->>F: Release claim and retain unacknowledged file
      W-->>W: Record failure and retry on next bounded backoff
    end
  end
```

The event ID is deterministic across retries and ClickHouse queries deduplicate logical events. A successful HTTP response is required before writing the acknowledgement marker; an ambiguous timeout is never treated as delivered. Shutdown stops intake, drains the bounded queue, closes the active file, attempts a bounded final transfer, and leaves failed files on disk.

## 82. Platform administrator searches audit events (Sprint 28)

```mermaid
sequenceDiagram
  actor A as Platform admin
  participant B as Browser
  participant G as Go :8091
  participant H as ClickHouse audit_events
  participant O as Audit health state

  A->>B: Open /admin/audit-logs and choose filters
  B->>G: GET /api/platform/audit-logs?filters
  G->>G: Require platform_admin and validate bounded filters
  alt unauthorized or invalid filter
    G-->>B: 401/403 or validation_error
  else authorized
    G->>H: Query deduplicated audit events with FINAL and LIMIT
    alt ClickHouse unavailable
      G-->>B: 503 analytics_unavailable without raw infrastructure detail
    else query succeeds
      H-->>G: Events and next cursor
      G->>O: Read last transfer and pending-file summary
      G-->>B: Events, filters, pagination, and safe delivery health
    end
  end
  B-->>A: Render tenant/action/resource/outcome audit rows
```

See 31-cross-tenant-audit-log-spec.md, 03-er-diagram.md, 04-api-spec.md, and 05-ux-ui.md.

## 83. Platform administrator opens system performance monitoring (Sprint 29)

```mermaid
sequenceDiagram
  actor A as Platform admin
  participant B as Browser
  participant G as Go :8091
  participant X as internal/auth
  participant O as internal/observability
  participant P as Postgres callcenter
  participant R as Redis monti_jarvis:
  participant M as MinIO monti-jarvis
  participant C as ClickHouse monti_jarvis
  participant L as Audit health state

  A->>B: Open /admin/monitoring
  B->>G: GET /api/platform/system-performance?filters
  G->>X: Require platform_admin and validate query bounds
  alt missing or non-platform session
    X-->>G: unauthorized or forbidden
    G-->>B: 401/403 safe error without tenant metadata
  else authorized
    G->>O: Start request-time snapshot with bounded deadline
    par shared dependency probes
      O->>P: Ping with shared deadline
      O->>R: PING with shared deadline
      O->>M: Check configured bucket with shared deadline
      O->>C: Ping with shared deadline
    and configured service states
      O->>O: Normalize NATS, LiveKit, and Gemini state
    end
    G->>P: List active tenant metadata with filter and page
    G->>C: Read allowlisted analytics freshness per tenant
    G->>L: Read audit delivery health without local paths
    alt partial probe or analytics failure
      O-->>G: Safe degraded/unavailable state
      G-->>B: 200 normalized snapshot with bounded rows
    else all required checks complete
      O-->>G: Operational or degraded snapshot
      G-->>B: 200 snapshot with summary, components, rows, and page
    end
    B-->>A: Render health matrix, tenant rows, freshness, and retry action
  end
```

Shared dependency probes run once per request and use one bounded deadline. Tenant analytics and audit state are read-only enrichments; their failure changes the normalized status but does not expose raw infrastructure errors. The monitoring route is never called from customer call, voice relay, quota, archive, or chat handlers.

See 32-platform-system-performance-spec.md, 03-er-diagram.md, 04-api-spec.md, and 05-ux-ui.md.

## 84. Platform administrator opens overall call-center statistics (Sprint 30)

```mermaid
sequenceDiagram
  actor A as Platform admin
  participant B as Browser
  participant G as Go :8091
  participant X as internal/auth
  participant P as Postgres callcenter
  participant C as internal/clickhouse
  participant H as ClickHouse monti_jarvis

  A->>B: Open /admin/call-center
  B->>G: GET /api/platform/call-center/statistics?start_date&end_date&tenant_id&limit&offset
  G->>X: Require platform_admin and validate bounded query
  alt missing or non-platform session
    X-->>G: unauthorized or forbidden
    G-->>B: 401/403 safe error without tenant or customer data
  else authorized
    G->>G: Resolve default today range in deployment timezone
    G->>P: List active tenant metadata for the bounded tenant page
    G->>C: Query FINAL call_center_usage_facts for overall and tenant aggregates
    C->>H: Aggregate completed facts by tenant, channel, and avatar
    H-->>C: Totals, tenant rows, breakdowns, and max updated_at
    C-->>G: Activity aggregate and freshness
    G->>P: Read redacted satisfaction and active package metadata
    alt ClickHouse unavailable or query timeout
      C-->>G: analytics unavailable
      G-->>B: 503 analytics_unavailable without provider detail
    else activity query succeeds
      alt ratings or package enrichment unavailable
        P-->>G: Partial enrichment failure
        G-->>B: 200 aggregate with explicit enrichment unavailable state
      else all enrichments succeed
        P-->>G: Satisfaction and package metadata
        G-->>B: 200 aggregate, tenant rows, freshness, and pagination
      end
    end
    B-->>A: Render KPIs, breakdowns, tenant table, and state
  end
```

The activity query and tenant-page query share the same inclusive date range and `FINAL` semantics. Postgres enrichment is aggregate-only: rating scores are reduced to count, average, completion rate, and distribution; package data is limited to package name and range usage labels. No transcript, contact, ticket, audio, or audit payload is read.

### Statistics state transitions

| State | Trigger | UI/API behavior |
| --- | --- | --- |
| `loading` | Initial request or filter change | Keep filters stable and show fixed-height placeholders. |
| `current` | ClickHouse query succeeds and freshness is within the existing 5-minute threshold | Render normal totals and tenant rows. |
| `stale` | Query succeeds but latest fact projection exceeds the freshness threshold | Return 200 with stale freshness; keep values visible and label them stale. |
| `empty` | Valid range has no completed facts | Return 200 with zero totals and an explicit empty state. |
| `unavailable` | ClickHouse is disabled, times out, or rejects the query | Return 503 `analytics_unavailable`; do not present zero activity. |
| `enrichment_unavailable` | Ratings or package metadata cannot be read while facts are available | Return 200 with activity intact and the affected enrichment marked unavailable. |

See 33-platform-call-center-statistics-spec.md, 03-er-diagram.md, 04-api-spec.md, and 05-ux-ui.md.

## 85. Platform administrator opens billing and AI usage (Sprint 31)

```mermaid
sequenceDiagram
  actor A as Platform admin
  participant B as Browser
  participant G as Go :8091
  participant X as internal/auth
  participant P as Postgres callcenter
  participant R as Redis DB 4
  participant C as internal/clickhouse
  participant H as ClickHouse monti_jarvis

  A->>B: Open /admin/billing/usage
  B->>G: GET /api/platform/billing/usage?start_date&end_date&tenant_id&limit&offset
  G->>X: Require platform_admin and validate bounded query
  alt missing or non-platform session
    X-->>G: unauthorized or forbidden
    G-->>B: 401/403 safe error without tenant or cost details
  else authorized
    G->>P: Read paid order and active entitlement aggregates
    G->>R: Read bounded current quota enforcement snapshot
    G->>C: Query grouped call and AI usage facts
    C->>H: FINAL aggregate by tenant, provider, model, and measurement state
    H-->>C: Usage minutes, observed/estimated cost, coverage, and freshness
    C-->>G: Reporting aggregates
    alt one source is unavailable
      G-->>B: 200 partial sections with safe unavailable state
    else required aggregate succeeds
      G-->>B: 200 billing, quota, AI cost, reconciliation, and tenant page
    end
    B-->>A: Render paid value, quota state, coverage, cost, and warnings
  end
```

## 86. Completed interaction emits an AI usage meter event (Sprint 31)

```mermaid
sequenceDiagram
  participant I as Chat or voice interaction
  participant G as Go :8091
  participant M as internal/metering
  participant C as internal/clickhouse
  participant H as ClickHouse monti_jarvis

  I->>G: Complete committed interaction
  G->>M: Normalize provider usage with deterministic event_id
  alt provider units returned
    M->>M: Mark observed and resolve immutable rate_version
  else units missing but duration fallback configured
    M->>M: Mark estimated and apply documented fallback rate
  else no units or applicable rate
    M->>M: Mark unavailable without pricing at zero
  end
  M->>C: Insert or replace ai_cost_usage_facts by event_id
  C->>H: Idempotent ClickHouse write
  alt analytics write fails
    H-->>C: Retryable failure
    C-->>M: Record bounded replay gap; do not fail interaction
  else write succeeds
    H-->>C: Acknowledge one logical fact
  end
```

### Billing usage state transitions

| State | Trigger | Reporting behavior |
| --- | --- | --- |
| `observed` | Provider units and rate are present | Include in observed cost and coverage. |
| `estimated` | Documented fallback units and rate are present | Include separately and label as estimated. |
| `unavailable` | Units or rate are missing | Preserve the gap; never show exact zero cost. |
| `replay_pending` | ClickHouse write failed after interaction commit | Keep customer path successful and retry by deterministic event id. |

See 34-platform-billing-quota-ai-cost-spec.md, 03-er-diagram.md, 04-api-spec.md, and 05-ux-ui.md.
