---
id: DES-0005
title: UX/UI — ASCII Wireframes
status: approved
updated: 2026-07-16
sprint: SPRINT-028
---

# UX/UI — ASCII Wireframes

**Surfaces:** Customer portal `apps/customer-web` at `/` · Platform admin `apps/platform-admin-web` at `/admin` *(Sprint 4)* · Tenant portal `apps/tenant-web` at `/tenant` *(Sprint 6)*.

## Customer portal

Primary surface at `/`. Layout matches legacy dark-neon two-panel design.

## Screen map → API

| UI zone | User action | API / WS |
| --- | --- | --- |
| A1 Brand header | (static) | `GET /images/monti-logo.png` |
| A2 Avatar halo | Select agent | `GET /api/workforce` (on load) |
| A3 Agent cards | Click agent | local state; greeting via chat optional |
| A4 Start/End call | Toggle voice | `POST /api/calls` → `WS /ws/voice?agent&topic` → `POST .../end` |
| A5 Timer | (display) | client timer on call start |
| B1 Topic tabs | General/Billing/Technical | `topic` field on chat + voice WS |
| B2 Chat stream | View messages | SSE turns + chat responses + voice transcripts |
| B3 Composer Send | Text question | `POST /api/chat` |
| B4 Infra line | (display) | `GET /api/infra` |
| B5 Session line | (display) | `session_id` from chat/call |
| Citations | Under assistant bubble | `sources[]` from `/api/chat` |

## Full layout (desktop)

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  http://localhost:8091/                                                          │
├─────────────────────────────┬──────────────────────────────────────────────────┤
│  CONTROL PANEL              │  WORKSPACE — Caller Desk                          │
│                             │                                                  │
│  ┌──┐ MONTI                 │  Caller Desk          Postgres ok · Redis ok …  │
│  │██│ Inbound Call Center   │  ┌────────┐ ┌────────┐ ┌──────────┐            │
│  └──┘                       │  │General*│ │Billing │ │Technical │  ← topic    │
│                             │  └────────┘ └────────┘ └──────────┘            │
│      ╭───────────╮          │                                                  │
│     │  ┌───────┐  │         │  ┌──┐  Welcome to Monti…                       │
│     │  │ AVATAR│  │         │  │A │  Choose an agent…                          │
│     │  │ image │  │         │  └──┘                                             │
│     │  └───────┘  │         │  ┌──┐  ┌─────────────────────────────────────┐  │
│      ╰───────────╮          │  │C │  │ When are invoices due?              │  │
│       waveform    │         │  └──┘  └─────────────────────────────────────┘  │
│                             │  ┌──┐  ┌─────────────────────────────────────┐  │
│       Ava                   │  │M │  │ Invoices are due within 15 days.    │  │
│  General Support · trait    │  └──┘  │ [billing · KB]                      │  │
│                             │        └─────────────────────────────────────┘  │
│  ┌──────────┬──────────┐  │                                                  │
│  │ 00:01:23 │ End call │  │  ┌──────────────────────────────┬────────┐       │
│  └──────────┴──────────┘  │  │ Ask your question...         │ Send   │       │
│  On call with Ava.        │  └──────────────────────────────┴────────┘       │
│                             │  Call abc123 · Max          ← session label        │
│  ┌────────────────────────┐ │                                                  │
│  │[img] Ava      [Select] │ │                                                  │
│  │ General · Popular      │ │                                                  │
│  ├────────────────────────┤ │                                                  │
│  │[img] Max       Select   │ │                                                  │
│  ├────────────────────────┤ │                                                  │
│  │[img] Luna      Select   │ │                                                  │
│  ├────────────────────────┤ │                                                  │
│  │[img] Neo       Select   │ │                                                  │
│  └────────────────────────┘ │                                                  │
│   ↑ GET /api/workforce      │                                                  │
└─────────────────────────────┴──────────────────────────────────────────────────┘
```

## Flow A — Text chat with RAG

```
User types in composer ──► POST /api/chat
                              │
                              ├─ agent_id ← selected card (A3)
                              ├─ topic    ← active tab (B1)
                              └─ history  ← prior turns
                              │
                              ▼
                         Append user bubble (C)
                         Append thinking bubble
                              │
                              ▼
                         Replace with reply + citation chips
                         sources[] → [scope · KB] tags
```

## Flow B — Voice call

```
User clicks Start call (A4)
    │
    ├─► POST /api/calls          → session.id, room_name
    ├─► GET /api/calls/{id}/events (SSE) → live turns
    └─► WS /ws/voice?agent=ava&topic=general
            │
            ├─ mic → audio JSON → Gemini Live
            └─ transcript events → chat stream (B2)

User clicks End call
    └─► POST /api/calls/{id}/end
```

## Flow C — Agent selection

```
Click agent card (A3)
    │
    ├─ Update halo portrait (A2)     image from agent.image
    ├─ Update waveform color         agent.color
    ├─ If live call → hang up first
    └─ Optional greeting in chat     agent.greeting (on select)
```

## Mobile (≤980px)

```
┌─────────────────────┐
│  CONTROL PANEL      │  stacked full width
│  (avatar + agents)  │
├─────────────────────┤
│  WORKSPACE          │
│  chat + composer    │
└─────────────────────┘
```

## Legacy UI (`/legacy/`)

Same API mapping; vanilla JS in `internal/web/public/index.html`. Useful reference for visual parity.

## Design tokens

| Token | Value |
| --- | --- |
| Background | `#020712` + blue/purple radial gradients |
| Panel | glass dark `rgb(5 16 31 / 82%)` |
| Accent | cyan `#00b7ff`, agent `--assistant-color` |
| Radius | panels `26px`, bubbles `16px` |
| Font | Inter |

## Platform Admin Portal (Sprint 4)

**App:** `apps/platform-admin-web` · **URL:** `http://localhost:8091/admin`  
Customer portal `/` unchanged. Requires `AUTH_DISABLED=false` for login.

### Global screen index

| Screen | Route | Primary API |
| --- | --- | --- |
| Login | `/admin/login` | `POST /api/auth/login` |
| Packages list | `/admin/packages` | `GET /api/platform/packages` |
| Package create | `/admin/packages/new` | `GET /api/platform/rule-schemas`, `POST /api/platform/packages` |
| Package edit | `/admin/packages/[id]` | `GET/PUT/DELETE /api/platform/packages/{id}` |
| Profile | `/admin/profile` | `GET /api/auth/me` |
| Tenant entitlement | `/admin/tenants/[id]/entitlement` | `GET/POST/DELETE /api/platform/tenants/{id}/entitlement` |
| Avatars list *(S5)* | `/admin/avatars` | `GET /api/platform/avatars` |
| Avatar create *(S5)* | `/admin/avatars/new` | `POST /api/platform/avatars` |
| Avatar edit *(S5)* | `/admin/avatars/[id]` | `GET/PUT/DELETE /api/platform/avatars/{id}` |
| Tenant avatars *(S5)* | `/admin/tenants/[id]/avatars` | `GET/POST/DELETE /api/platform/tenants/{id}/avatars*` |
| Tenants list *(S6)* | `/admin/tenants` | `GET /api/platform/tenants?status=` |
| Tenant KYC review *(S7)* | `/admin/tenants/[id]/kyc` | `GET/POST /api/platform/tenants/{id}/kyc*` |

### Design tokens (platform admin)

| Token | Value |
| --- | --- |
| Background | `#0a0f1a` solid + subtle grid |
| Panel | `rgb(12 20 36 / 92%)`, border `rgb(0 183 255 / 18%)` |
| Accent | cyan `#00b7ff` (primary button, active nav) |
| Danger | `#ff5c7a` (archive, revoke, errors) |
| Success | `#3dd68c` (save toast) |
| Radius | cards `16px`, inputs `10px` |
| Font | Inter · labels `13px` · table `14px` |

---

### Screen P0 — Login (`/admin/login`)

**Purpose:** Authenticate `platform_admin` only. No header nav (unauthenticated shell).

#### Screen map → API

| Zone | Element | Action | API / behavior |
| --- | --- | --- | --- |
| P0-L1 | Logo + title | static | Monti wordmark |
| P0-L2 | Email input | type email | bound `email` → login body |
| P0-L3 | Password input | type password | bound `password` |
| P0-L4 | Sign in button | click | `POST /api/auth/login` |
| P0-L5 | Error banner | display | API `{error}` on 401 |
| P0-L6 | Wrong-portal note | static | tenant_admin hint |
| P0-L7 | Loading state | display | disable form while pending |

#### Full layout (desktop)

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  http://localhost:8091/admin/login                                               │
│                                                                                  │
│                         ┌─────────────────────────────┐                          │
│                         │  ┌──┐  MONTI                │  P0-L1                   │
│                         │  └──┘  Platform Admin       │                          │
│                         │                             │                          │
│                         │  Sign in to manage packages │                          │
│                         │  and tenant entitlements.   │                          │
│                         │                             │                          │
│                         │  ┌─ Error (P0-L5) ───────┐ │  ← invalid credentials   │
│                         │  │ Invalid email or pass. │ │                          │
│                         │  └────────────────────────┘ │                          │
│                         │                             │                          │
│                         │  Email                      │                          │
│                         │  [ platform@monti.local  ]  │  P0-L2                   │
│                         │                             │                          │
│                         │  Password                   │                          │
│                         │  [ ••••••••••••••••••    ]  │  P0-L3                   │
│                         │                             │                          │
│                         │       [ Sign in ] (P0-L4)     │  primary cyan            │
│                         │                             │                          │
│                         │  Tenant admins: use API or  │  P0-L6                   │
│                         │  tenant portal (Sprint 15+).│                          │
│                         └─────────────────────────────┘                          │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

#### States

| State | UI |
| --- | --- |
| Default | Empty fields; Sign in enabled |
| Submitting | Button spinner; inputs disabled (P0-L7) |
| Error 401 | Red banner P0-L5; fields retain values |
| Success | Store `access_token` + `refresh_token` in `sessionStorage` → redirect `/admin/packages` or `?next=` |
| Already logged in | If valid token + `platform_admin` → redirect `/admin/packages` |

#### Validation (client)

- Email required, contains `@`
- Password required, min 1 char
- Do not echo which field failed (match API generic 401)

---

### Screen P1 — App shell (authenticated)

Shared chrome for P2–P6. Not a standalone route.

#### Screen map → API

| Zone | Element | Action | API / behavior |
| --- | --- | --- | --- |
| P1-N1 | Logo | click | navigate `/admin/packages` |
| P1-N2 | Packages nav | click | `/admin/packages` |
| P1-N3 | Profile nav | click | `/admin/profile` |
| P1-N4 | User chip | display | email from login response or `/api/auth/me` |
| P1-N5 | Logout | click | `POST /api/auth/logout` → clear storage → `/admin/login` |
| P1-N6 | Active nav | display | underline Packages or Profile |

#### Full layout

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  P1-N1 ◆ MONTI Admin    P1-N2 Packages*   P1-N3 Profile     P1-N4 platform@…  P1-N5 Logout │
├──────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│   « main content slot: packages | profile | entitlement »                        │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

### Screen P2 — Profile (`/admin/profile`)

**Purpose:** Read-only operator identity from JWT session.

#### Screen map → API

| Zone | Element | Action | API field |
| --- | --- | --- | --- |
| P2-H1 | Page title | static | "Profile" |
| P2-R1 | Email row | display | `email` |
| P2-R2 | Display name | display | `display_name` |
| P2-R3 | Role badge | display | `role` (`platform_admin`) |
| P2-R4 | Tenant row | display | `tenant_id` or `—` if empty |
| P2-R5 | User id | display | `id` (monospace, copy optional) |
| P2-E1 | Load error | display | failed `GET /api/auth/me` |

#### Full layout

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  [ App shell P1 ]                                                                │
├──────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  Profile (P2-H1)                                                                 │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │  Account                                                                   │  │
│  │                                                                            │  │
│  │  Email          platform@monti.local                          P2-R1        │  │
│  │  ────────────────────────────────────────────────────────────────────────  │  │
│  │  Display name   Monti Platform                                P2-R2        │  │
│  │  ────────────────────────────────────────────────────────────────────────  │  │
│  │  Role           ┌─────────────────┐                           P2-R3        │  │
│  │                 │ platform_admin  │  cyan badge                              │  │
│  │                 └─────────────────┘                                        │  │
│  │  ────────────────────────────────────────────────────────────────────────  │  │
│  │  Tenant         —                                           P2-R4        │  │
│  │  ────────────────────────────────────────────────────────────────────────  │  │
│  │  User ID        usr_platform                                P2-R5        │  │
│  │                                                                            │  │
│  │  Password change and MFA are not in Sprint 4 scope.                      │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

On mount: `GET /api/auth/me`. 401 → logout redirect.

---

### Screen P3 — Packages list (`/admin/packages`)

**Purpose:** Browse catalog; entry to create, edit, archive, assign demo tenant.

#### Screen map → API

| Zone | Element | Action | API |
| --- | --- | --- | --- |
| P3-H1 | Title | static | "Packages" |
| P3-H2 | New package | click | navigate `/admin/packages/new` |
| P3-F1 | Status filter | change | `GET /api/platform/packages?status=` |
| P3-T1 | Table rows | display | `packages[]` |
| P3-A1 | Edit link | click | `/admin/packages/{id}` |
| P3-A2 | Assign demo | click | `/admin/tenants/demo/entitlement` |
| P3-A3 | Archive | confirm + click | `DELETE /api/platform/packages/{id}` |
| P3-E1 | Empty state | display | no rows → link create |
| P3-E2 | Load error | display | API error banner |

#### Full layout

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  [ App shell P1 — Packages active P1-N6 ]                                        │
├──────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  Packages (P3-H1)                    Status [ active ▼] P3-F1   [+ New] P3-H2   │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │ slug       │ name        │ status  │ schema   │ price    │ actions          │  │
│  ├────────────┼─────────────┼─────────┼──────────┼──────────┼──────────────────┤  │
│  │ starter    │ Starter     │ active  │ rules-v1 │ $0/mo    │ P3-A1 Edit       │  │
│  │            │             │         │          │          │ P3-A2 Assign demo│  │
│  │            │             │         │          │          │ P3-A3 Archive    │  │
│  ├────────────┼─────────────┼─────────┼──────────┼──────────┼──────────────────┤  │
│  │ pro        │ Pro         │ active  │ rules-v1 │ —        │ …                │  │
│  ├────────────┼─────────────┼─────────┼──────────┼──────────┼──────────────────┤  │
│  │ enterprise │ Enterprise  │ active  │ rules-v1 │ —        │ …                │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  Hover row: subtle cyan border · Archive opens confirm modal                     │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

Table columns map from API: `slug`, `name`, `status`, `rules_schema_id`, `price_cents` + `billing_period`.

---

### Screen P4 — Package create (`/admin/packages/new`)

**Purpose:** Create catalog entry with schema-driven `rules` jsonb form.

#### Screen map → API

| Zone | Element | Action | API |
| --- | --- | --- | --- |
| P4-H1 | Breadcrumb | click | ← Back to list |
| P4-M1 | Slug | input | `slug` (required, lowercase) |
| P4-M2 | Name | input | `name` |
| P4-M3 | Description | textarea | `description` |
| P4-M4 | Status | select | `draft` \| `active` |
| P4-M5 | Price cents | number | `price_cents` |
| P4-M6 | Currency | select | `currency` default USD |
| P4-M7 | Billing period | select | `billing_period` |
| P4-S1 | Schema select | display | `rules_schema_id` from `GET /api/platform/rule-schemas` |
| P4-R* | Rule fields | dynamic | rendered from `fields` jsonb (int/bool) |
| P4-B1 | Cancel | click | back to list |
| P4-B2 | Create | click | `POST /api/platform/packages` |
| P4-E1 | Field errors | inline | 400 validation per field |

#### Full layout

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  [ App shell P1 ]                                                                │
├──────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ← Packages (P4-H1)                                                              │
│  New package                                                                     │
│                                                                                  │
│  ┌─ Metadata ────────────────────────────────────────────────────────────────┐  │
│  │ Slug *        [ my-package        ]  P4-M1   lowercase, unique             │  │
│  │ Name *        [ My Package        ]  P4-M2                                 │  │
│  │ Description   [ Optional text…     ]  P4-M3                                 │  │
│  │ Status        ( draft ) ( active )  P4-M4                                 │  │
│  │ Price (¢)     [ 0                 ]  P4-M5   Currency [ USD ▼] P4-M6      │  │
│  │ Billing       [ monthly ▼ ]         P4-M7                                 │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌─ Rules — schema rules-v1 (P4-S1) ─────────────────────────────────────────┐  │
│  │ Max AI employees      [ 2    ]  int   ← fields.max_ai_employees           │  │
│  │ Max monthly minutes   [ 500  ]  int                                         │  │
│  │ Max KM documents      [ 50   ]  int                                         │  │
│  │ Max concurrent calls  [ 2    ]  int                                         │  │
│  │ Voice enabled         [x] yes   bool  P4-R*                                 │  │
│  │ RAG enabled           [x] yes   bool                                        │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│              [ Cancel P4-B1 ]              [ Create package P4-B2 ]              │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

Success 201 → toast → redirect `/admin/packages/[id]`.

---

### Screen P5 — Package edit (`/admin/packages/[id]`)

**Purpose:** Update metadata and rules; archive package.

Same zones as P4 with pre-filled values from `GET /api/platform/packages/{id}`.

#### Full layout (delta from P4)

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  ← Packages    Edit: Starter (pkg-starter)                                         │
│                                                                                  │
│  [ Metadata panel — same fields as P4, prefilled ]                               │
│  [ Rules panel — same dynamic fields, prefilled from package.rules ]             │
│                                                                                  │
│  [ Save changes ]                              [ Archive package ]  danger P3-A3  │
│                                                                                  │
│  Archive confirm modal:                                                          │
│  ┌────────────────────────────────────────┐                                      │
│  │ Archive "Starter"? Active entitlements │                                      │
│  │ block archive (409).                   │                                      │
│  │        [ Cancel ]  [ Archive ]         │                                      │
│  └────────────────────────────────────────┘                                      │
└──────────────────────────────────────────────────────────────────────────────────┘
```

`PUT /api/platform/packages/{id}` on Save · `DELETE` on Archive.

---

### Screen P6 — Tenant entitlement (`/admin/tenants/[id]/entitlement`)

**Purpose:** Assign or revoke package for a tenant (dev default: `demo`).

#### Screen map → API

| Zone | Element | Action | API |
| --- | --- | --- | --- |
| P6-H1 | Tenant id | display | route param `id` |
| P6-C1 | Current package | display | `GET …/entitlement` → `package.name` |
| P6-C2 | Status badge | display | `status` active/revoked |
| P6-C3 | Rules summary | display | `rules` key/value list |
| P6-S1 | Package select | change | dropdown from `GET /api/platform/packages` |
| P6-B1 | Assign | click | `POST …/entitlement {package_id}` |
| P6-B2 | Revoke | click | `DELETE …/entitlement` |
| P6-E1 | No entitlement | display | empty state + assign CTA |

#### Full layout

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  [ App shell P1 ]                                                                │
├──────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  Tenant entitlement — demo (P6-H1)                                               │
│                                                                                  │
│  ┌─ Current (P6-C*) ─────────────────────────────────────────────────────────┐  │
│  │ Package      Starter (pkg-starter)                                          │  │
│  │ Status       ┌────────┐                                                   │  │
│  │              │ active │  green                                            │  │
│  │              └────────┘                                                   │  │
│  │ Schema       rules-v1                                                     │  │
│  │ Rules        max_ai_employees: 2 · max_km_documents: 50 · voice: on …     │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌─ Assign package (P6-S1, P6-B1) ───────────────────────────────────────────┐  │
│  │ Package [ Pro (pkg-pro)        ▼ ]                                          │  │
│  │                    [ Assign to tenant ]                                     │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  [ Revoke entitlement ]  P6-B2  danger · confirm modal                           │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

### Platform admin — interaction flows

```
Flow 1 — Login → packages → logout
  /admin/login (P0) → POST login → /admin/packages (P3)
  → Profile (P2) GET /me → Logout (P1-N5) → /admin/login

Flow 2 — Create package
  P3-H2 → P4 → POST package → P5 edit view

Flow 3 — Assign demo tenant
  P3-A2 or P6 → POST entitlement → refresh P6 current panel

Flow 4 — Customer portal (unchanged)
  / — no auth · limits not enforced until Sprint 13
```

### Platform admin — component → file

| Component | Path |
| --- | --- |
| App layout + nav | `apps/platform-admin-web/src/routes/+layout.svelte` |
| Login page | `src/routes/login/+page.svelte` |
| Profile page | `src/routes/profile/+page.svelte` |
| Packages list | `src/routes/packages/+page.svelte` |
| Package form | `src/routes/packages/new/+page.svelte`, `packages/[id]/+page.svelte` |
| Entitlement | `src/routes/tenants/[id]/entitlement/+page.svelte` |
| Rules form | `src/lib/components/RulesForm.svelte` |
| Auth guard | `src/lib/auth/guard.ts` |
| API clients | `src/lib/api/auth.ts`, `packages.ts` |
| Styles | `src/app.css` |

---

### Screen P7 — Avatars list (`/admin/avatars`) *(Sprint 5)*

**Purpose:** Browse platform avatar catalog; entry to create, edit, archive, assign to demo tenant.

#### Screen map → API

| Zone | Element | Action | API |
| --- | --- | --- | --- |
| P7-H1 | Title | static | "Avatars" |
| P7-H2 | New avatar | click | `/admin/avatars/new` |
| P7-F1 | Status filter | change | `GET /api/platform/avatars?status=` |
| P7-T1 | Table rows | display | `avatars[]` |
| P7-A1 | Edit link | click | `/admin/avatars/{id}` |
| P7-A2 | Assign demo | click | `/admin/tenants/demo/avatars` |
| P7-A3 | Archive | confirm + click | `DELETE /api/platform/avatars/{id}` |

#### Full layout

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  [ App shell P1 — Avatars active ]                                               │
├──────────────────────────────────────────────────────────────────────────────────┤
│  Avatars (P7-H1)              Status [ active ▼] P7-F1        [+ New] P7-H2     │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │ slug │ name  │ role              │ voice  │ status │ actions               │  │
│  ├──────┼───────┼───────────────────┼────────┼────────┼───────────────────────┤  │
│  │ ava  │ Ava   │ General Support   │ Aoede  │ active │ P7-A1 Edit            │  │
│  │      │       │                   │        │        │ P7-A2 Assign demo     │  │
│  │      │       │                   │        │        │ P7-A3 Archive         │  │
│  ├──────┼───────┼───────────────────┼────────┼────────┼───────────────────────┤  │
│  │ max  │ Max   │ Billing Specialist│ Charon │ active │ …                     │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

### Screen P8 — Avatar create (`/admin/avatars/new`) *(Sprint 5)*

| Zone | Element | API field |
| --- | --- | --- |
| P8-M1 | Slug | `slug` |
| P8-M2 | Name | `name` |
| P8-M3 | Role / trait | `role`, `trait` |
| P8-M4 | Color | `color` |
| P8-M5 | Image URL | `image_url` |
| P8-M6 | Greeting | `greeting` |
| P8-M7 | Flags | `flags.popular`, `flags.robot` |
| P8-V* | Voice profiles table | `voices[]`: `voice_provider_id`, `voice_id`, `voice`, `priority`, `status` |
| P8-B1 | Create | `POST /api/platform/avatars` (requires ≥1 voice row) |

---

### Screen P9 — Avatar edit (`/admin/avatars/[id]`) *(Sprint 5)*

Same fields as P8 prefilled from `GET /api/platform/avatars/{id}` including **Voice profiles** table (add row, reorder priority, disable alternate). **Save** → `PUT` with `voices[]`. **Archive** → `DELETE` with confirm modal (409 if tenant assignments active).

```
┌─ Voice profiles (P8-V*) ─────────────────────────────────────────────────────┐
│ priority │ provider          │ voice_id                          │ voice   │
│ 1        │ voice-gemini-live │ gemini-2.5-flash-native-audio-…   │ Aoede   │
│ 2        │ voice-grok-stub   │ grok-voice-stub                   │ Aoede   │  ← disabled
│ [ + Add alternate provider ]                                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

### Screen P10 — Tenant avatars (`/admin/tenants/[id]/avatars`) *(Sprint 5)*

| Zone | Element | API |
| --- | --- | --- |
| P10-H1 | Tenant id | route param |
| P10-C1 | Cap hint | `cap.max_ai_employees`, `active_count`, and demo `override_allowed` from GET |
| P10-C2 | Assigned list | `assignments[]` |
| P10-S1 | Avatar select | catalog dropdown |
| P10-B1 | Assign | `POST …/avatars {avatar_id}` |
| P10-B2 | Demote | `DELETE …/avatars/{avatar_id}` |
| P10-B3 | Promote | `POST …/avatars {avatar_id}` for a disabled assignment |

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  Tenant avatars — demo (P10-H1)     Cap: 2 max · 4 assigned (demo override)      │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌─ Assigned ────────────────────────────────────────────────────────────────┐  │
│  │ ava  Ava   General Support   [active]     [ Demote ]                        │  │
│  │ max  Max   Billing          [disabled]   [ Promote ]                        │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│  ┌─ Assign ──────────────────────────────────────────────────────────────────┐  │
│  │ Avatar [ luna ▼ ]              [ Assign to tenant ]  P10-B1                 │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

### Platform admin — Sprint 5 flows

```
Flow 1 — Manage catalog
  P7 list → P8 create → P9 edit

Flow 2 — Assign demo tenant
  P7-A2 or P10 → POST assignment → customer GET /api/workforce reflects list

Flow 3 — Customer portal (UI unchanged)
  / loads → GET /api/workforce (tenant demo) → same agent cards, DB-backed data
```

### Platform admin — Sprint 5 component → file

| Component | Path |
| --- | --- |
| Avatars list | `src/routes/avatars/+page.svelte` |
| Avatar form | `src/routes/avatars/new/+page.svelte`, `avatars/[id]/+page.svelte` |
| Tenant avatars | `src/routes/tenants/[id]/avatars/+page.svelte` |
| API client | `src/lib/api/avatars.ts` |
| Shell nav | `src/routes/+layout.svelte` (add Avatars link) |

## Tenant Portal (Sprint 6)

**App:** `apps/tenant-web` · **URL:** `http://localhost:8091/tenant`  
Prospect-facing signup. Customer portal `/` unchanged. Platform admin gains optional tenants list (P11).

### Global screen index (tenant)

| Screen | Route | Primary API |
| --- | --- | --- |
| Register | `/tenant/register` | `POST /api/public/tenant/register` |
| Success | `/tenant/register/success` | (client state + stored tokens) |
| Login stub | `/tenant/login` | `POST /api/auth/login` *(stub redirect to success/me)* |

### Screen map → API (tenant)

| Zone | Element | Action | API / behavior |
| --- | --- | --- | --- |
| T1-H1 | Header | static | Monti logo + “Tenant signup” |
| T1-F1 | Company name | input | `company_name` |
| T1-F2 | Workspace slug | input | `slug` (auto-suggest from company name) |
| T1-F3 | Admin email | input | `admin_email` |
| T1-F4 | Display name | input | `admin_display_name` |
| T1-F5 | Password | input | `admin_password` |
| T1-F6 | Confirm password | input | client-side match check |
| T1-B1 | Create account | submit | `POST /api/public/tenant/register` |
| T1-E1 | Error banner | display | API `400`/`409`/`429` |
| T2-S1 | Success title | static | “Account created — pending verification” |
| T2-S2 | Tenant id | display | `tenant_id` / slug |
| T2-L1 | Continue | link | `/tenant/login` stub or profile placeholder |
| T2-L2 | Customer portal | link | `/` |

### Screen T1 — Register (`/tenant/register`)

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  http://localhost:8091/tenant/register                                           │
│                                                                                  │
│                         ┌─────────────────────────────┐                          │
│                         │  ┌──┐  MONTI                │  T1-H1                   │
│                         │  └──┘  Start your AI call   │                          │
│                         │        center workspace     │                          │
│                         │                             │                          │
│                         │  ┌─ Error (T1-E1) ───────┐ │                          │
│                         │  │ Slug already taken     │ │                          │
│                         │  └────────────────────────┘ │                          │
│                         │                             │                          │
│                         │  Company name    (T1-F1)    │                          │
│                         │  [ Acme Corporation    ]  │                          │
│                         │                             │                          │
│                         │  Workspace URL   (T1-F2)    │                          │
│                         │  monti.app/ [ acme       ]  │                          │
│                         │                             │                          │
│                         │  Admin email     (T1-F3)    │                          │
│                         │  [ admin@acme.test     ]    │                          │
│                         │                             │                          │
│                         │  Your name       (T1-F4)    │                          │
│                         │  [ Acme Admin          ]    │                          │
│                         │                             │                          │
│                         │  Password        (T1-F5)    │                          │
│                         │  [ ••••••••••••        ]    │                          │
│                         │  Confirm         (T1-F6)    │                          │
│                         │  [ ••••••••••••        ]    │                          │
│                         │                             │                          │
│                         │    [ Create account ] T1-B1 │  primary cyan            │
│                         │                             │                          │
│                         │  Already have an account?   │                          │
│                         │  Sign in · Back to caller   │                          │
│                         │  portal (/)                 │                          │
│                         └─────────────────────────────┘                          │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

### Screen T2 — Success (`/tenant/register/success`)

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  http://localhost:8091/tenant/register/success                                   │
│                                                                                  │
│                         ┌─────────────────────────────┐                          │
│                         │  ✓  Account created  T2-S1  │                          │
│                         │                             │                          │
│                         │  Workspace: acme   T2-S2    │                          │
│                         │  Status: Pending verification                          │
│                         │                             │                          │
│                         │  We’ll review your signup.  │                          │
│                         │  You can sign in now; full  │                          │
│                         │  features unlock after KYC. │                          │
│                         │                             │                          │
│                         │  [ Continue ] T2-L1         │                          │
│                         │  [ Try caller demo → / ]    │  T2-L2                   │
│                         └─────────────────────────────┘                          │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

### Screen P11 — Tenants list (`/admin/tenants`) *(stretch / Sprint 6)*

Minimal table for platform_admin; API required even if UI deferred.

| Zone | Action | API |
| --- | --- | --- |
| P11-T1 | Filter status | `GET /api/platform/tenants?status=pending_kyc` |
| P11-T2 | Row link → KYC review | `/admin/tenants/{id}/kyc` |
| P11-T3 | Filter KYC queue | `GET /api/platform/tenants?kyc_status=submitted` |

### Screen P12 — Tenant KYC review (`/admin/tenants/{id}/kyc`) *(Sprint 7)*

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  MONTI ADMIN   Packages · Tenants · Avatars · Profile          platform@… Logout │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ← Tenants   P12-L1                                                              │
│                                                                                  │
│  KYC review — acme (Acme Corp)                              P12-H1               │
│  Tenant status: pending_kyc   ·   KYC: submitted   ·   Reg: submitted            │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │ Registration                                                               │  │
│  │  Company     Acme Corp                              P12-R1                 │  │
│  │  Workspace   monti.app/acme                         P12-R2                 │  │
│  │  Admin email admin@acme.test                        P12-R3                 │  │
│  │  Submitted   2026-07-09 08:00                       P12-R4                 │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌──────────────────────────┐  ┌────────────────────────────────────────────┐  │
│  │ Photo  P12-P1              │  │ Contact  P12-C1                            │  │
│  │ ┌────────────┐           │  │  Name     Jane Doe                         │  │
│  │ │  portrait  │           │  │  Phone    +66 2 000 0000                   │  │
│  │ │  preview   │           │  │  Address  Bangkok, TH                      │  │
│  │ └────────────┘           │  └────────────────────────────────────────────┘  │
│  └──────────────────────────┘                                                  │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │ Documents  P12-D1                                                          │  │
│  │  • license.pdf   [ Open ]  → GET /api/assets/kyc/acme/docs/license.pdf     │  │
│  │  • tax-id.png    [ Open ]                                                  │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │ Reject reason (required for reject)  P12-F1                                │  │
│  │  [ Document is illegible — please re-upload a clearer scan.            ]   │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│         [ Reject ]  P12-B2 danger          [ Approve ]  P12-B1 primary cyan      │
│                                                                                  │
│  Success/error → shadcn-style FeedbackDialog (not inline banners)                │
└──────────────────────────────────────────────────────────────────────────────────┘
```

| Zone | Action | API |
| --- | --- | --- |
| P12-L1 | Back to tenants list | navigate `/admin/tenants` |
| P12-H1 | Page load | `GET /api/platform/tenants/{id}/kyc` |
| P12-P1 | Photo preview | `GET /api/assets/kyc/{id}/photo/{file}` |
| P12-D1 | Open document | `GET /api/assets/kyc/{id}/docs/{file}` |
| P12-B1 | Approve | `POST /api/platform/tenants/{id}/kyc/approve` |
| P12-B2 | Reject | `POST /api/platform/tenants/{id}/kyc/reject` `{reason}` |
| P12-F1 | Reject reason field | required before reject; `400` if empty |

### Flow P-C — KYC review queue

```text
P-C  platform_admin login
  → /admin/tenants?status=pending_kyc&kyc_status=submitted
  → click workspace row → /admin/tenants/{id}/kyc
  → review photo + documents + contact
  → Approve → feedback dialog "Tenant activated"
  → tenant removed from pending_kyc filter
```

### Flow P-D — Reject with reason

```text
P-D  /admin/tenants/{id}/kyc
  → enter rejection reason in P12-F1
  → Reject → POST .../kyc/reject
  → feedback dialog success
  → Resend email to admin@acme.test (when configured)
  → tenant stays pending_kyc; registration rejected
```

### Platform admin — component → file (Sprint 7)

| Component | Path |
| --- | --- |
| Tenants list (link column) | `apps/platform-admin-web/src/routes/tenants/+page.svelte` |
| KYC review page | `apps/platform-admin-web/src/routes/tenants/[id]/kyc/+page.svelte` |
| KYC API client | `apps/platform-admin-web/src/lib/api/kyc.ts` |
| Feedback dialog | `apps/platform-admin-web/src/lib/components/FeedbackDialog.svelte` |

### Screen P13 — Payment gateway settings (`/admin/settings/payment`) *(Sprint 8)*

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  MONTI ADMIN   Packages · Tenants · Avatars · Settings · Profile    Logout       │
├──────────────────────────────────────────────────────────────────────────────────┤
│  Settings › Payment                                         P13-H1               │
│                                                                                  │
│  Payment gateway (ChillPay)                                                      │
│  Configure platform-wide payment provider for tenant checkout (Sprint 9).        │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │ Provider        [ chillpay ▼ ]              P13-F1   mock | chillpay       │  │
│  │ Mode            [ test ▼ ]                  P13-F2   test | live           │  │
│  │ Merchant code   [ M123456                    ]  P13-F3                     │  │
│  │ API key         [ ••••••••••••abcd          ]  P13-F4   password input     │  │
│  │ MD5 secret key  [ ••••••••••••••••          ]  P13-F5   password input     │  │
│  │ Base URL        [ https://sandbox-api…/Payment/ ]  P13-F6                │  │
│  │ Route no        [ 1 ]                       P13-F7                       │  │
│  │ Currency        [ THB ]                     P13-F8                       │  │
│  │ Return URL      [ http://localhost:8091/tenant/billing/return ]  P13-F9  │  │
│  │ Callback URL    http://localhost:8091/api/callbacks/chillpay   P13-R1   │  │
│  │                 (read-only — derived from APP_PUBLIC_URL)                │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  Connection: unknown · Last callback: —                    P13-S1                │
│                                                                                  │
│         [ Test connection ]  P13-B1 outline        [ Save ]  P13-B2 primary      │
│                                                                                  │
│  Success/error → shadcn-style FeedbackDialog (not inline banners)                │
└──────────────────────────────────────────────────────────────────────────────────┘
```

| Zone | Action | API |
| --- | --- | --- |
| P13-H1 | Page load | `GET /api/platform/payment-gateway` |
| P13-F1–F9 | Edit fields | local state |
| P13-R1 | Callback URL preview | from GET response `callback_url` |
| P13-B2 | Save | `PUT /api/platform/payment-gateway` |
| P13-B1 | Test connection | `POST /api/platform/payment-gateway/test` |
| P13-S1 | Status line | `connection_status`, `last_callback_at` from GET |

### Flow P-E — Configure ChillPay

```text
P-E  platform_admin login
  → /admin/settings/payment
  → select provider chillpay (or mock for dev)
  → fill merchant code, API key, MD5 key, base URL, return URL
  → Save → PUT /api/platform/payment-gateway
  → feedback dialog "Payment gateway saved"
```

### Flow P-F — Test connection + callback ops

```text
P-F  /admin/settings/payment
  → Test connection → POST .../test
  → feedback dialog success or error message from 502 body

Ops (curl):
  → POST /api/callbacks/chillpay with form body + valid CheckSum
  → or PAYMENT_CALLBACK_DEV_BYPASS=true for local smoke
  → verify row in payment_callback_events
```

### Platform admin — component → file (Sprint 8)

| Component | Path |
| --- | --- |
| Payment settings page | `apps/platform-admin-web/src/routes/settings/payment/+page.svelte` |
| Payment API client | `apps/platform-admin-web/src/lib/api/payment.ts` |
| Shell nav (Settings link) | `apps/platform-admin-web/src/routes/+layout.svelte` |

### Screen T4 — Billing catalog (`/tenant/billing`) *(Sprint 9)*

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  http://localhost:8091/tenant/billing                                            │
│                                                                                  │
│  ┌──┐  MONTI TENANT     Backoffice · Billing · Profile              Logout       │
│                                                                                  │
│  Billing & packages                                           T4-H1              │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │ Current plan  T4-C1                                                        │  │
│  │  Starter (active) · assigned 2026-07-01                                    │  │
│  │  max_ai_employees: 2 · max_monthly_call_minutes: 500                       │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  Available packages                                                              │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────┐       │
│  │ Starter      T4-P1  │  │ Pro          T4-P2  │  │ Enterprise   T4-P3  │       │
│  │ ฿0 / mo             │  │ ฿2,990 / mo         │  │ Contact sales       │       │
│  │ 2 avatars           │  │ 5 avatars           │  │ …                   │       │
│  │ [ Current ]         │  │ [ Buy Pro ]  T4-B1  │  │ [ Buy ]      T4-B2  │       │
│  └─────────────────────┘  └─────────────────────┘  └─────────────────────┘       │
│                                                                                  │
│  Success/error → FeedbackDialog (not inline banners)                             │
└──────────────────────────────────────────────────────────────────────────────────┘
```

| Zone | Action | API |
| --- | --- | --- |
| T4-H1 | Page load | `GET /api/tenant/packages` + `GET /api/entitlements/me` |
| T4-C1 | Current plan card | from `current_entitlement` in packages response |
| T4-B1/B2 | Buy | `POST /api/tenant/checkout` → redirect `payment_url` |
| T4-P1–P3 | Package cards | catalog row; disable Buy on current package |

### Screen T5 — Payment return (`/tenant/billing/return`) *(Sprint 9)*

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  http://localhost:8091/tenant/billing/return?order_id=ord_…                      │
│                                                                                  │
│                         ┌─────────────────────────────┐                          │
│                         │  Payment status      T5-H1    │                          │
│                         │                             │                          │
│                         │  ⏳ Processing…             │  pending — poll T5-P1    │
│                         │  ✓ Payment successful       │  paid                    │
│                         │  ✗ Payment failed           │  failed                  │
│                         │                             │                          │
│                         │  Package: Pro               │                          │
│                         │  Order: mj_demo_x7k2m9       │                          │
│                         │                             │                          │
│                         │  [ Back to billing ] T5-L1  │                          │
│                         └─────────────────────────────┘                          │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

| Zone | Action | API |
| --- | --- | --- |
| T5-H1 | Page load | `GET /api/tenant/orders/{id}` (poll every 2s while `pending`) |
| T5-P1 | Spinner | stops when `paid` or `failed` |
| T5-L1 | Back | navigate `/tenant/billing` |

Query: `order_id` from ChillPay `ReturnUrl` or checkout response stored in sessionStorage.

### Screen T6 — Mock pay (`/tenant/billing/mock-pay`) *(Sprint 9 dev)*

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  Mock payment (dev only)                                      T6-H1              │
│                                                                                  │
│  Order ord_a1b2c3d4 · Pro · ฿2,990                                               │
│                                                                                  │
│         [ Complete mock payment ]  T6-B1 primary                                 │
│                                                                                  │
│  Calls POST /api/dev/mock-pay/{order_id} then redirects to T5 return page        │
└──────────────────────────────────────────────────────────────────────────────────┘
```

| Zone | Action | API |
| --- | --- | --- |
| T6-B1 | Complete | `POST /api/dev/mock-pay/{order_id}` → redirect T5 |

### Flow T-C — Buy package (ChillPay)

```text
T-C  tenant_admin login (active tenant)
  → /tenant/billing
  → Buy Pro → POST /api/tenant/checkout
  → redirect ChillPay PaymentUrl
  → pay in sandbox
  → ChillPay POST /api/callbacks/chillpay (server)
  → browser ReturnUrl /tenant/billing/return?order_id=...
  → poll order paid → entitlement shows Pro
```

### Flow T-D — Mock checkout (CI / local)

```text
T-D  platform sets gateway provider=mock
  → tenant /tenant/billing → Buy
  → /tenant/billing/mock-pay?order_id=...
  → Complete mock payment
  → /tenant/billing/return → paid
  → GET /api/entitlements/me → Pro
```

### Flow T-E — Combined E2E with Sprint 8 (manual §0+§1)

```text
T-E  §0 platform /admin/settings/payment — chillpay sandbox + test connection
  → §1 tenant /tenant/billing — real InitPayment
  → ngrok callback to /api/callbacks/chillpay
  → verify entitlement + payment_callback_events + payment_orders paid
```

### Tenant portal — component → file (Sprint 9)

| Component | Path |
| --- | --- |
| Billing catalog | `apps/tenant-web/src/routes/billing/+page.svelte` |
| Return page | `apps/tenant-web/src/routes/billing/return/+page.svelte` |
| Mock pay | `apps/tenant-web/src/routes/billing/mock-pay/+page.svelte` |
| Billing API client | `apps/tenant-web/src/lib/api/billing.ts` |
| Shell nav | `apps/tenant-web/src/routes/+layout.svelte` |

## Sprint 9 — Buy Package (customer + platform admin unchanged)

| Surface | Sprint 9 UX | API |
| --- | --- | --- |
| Customer `/` | Unchanged | Public chat/voice |
| Tenant `/tenant` | **New** T4 billing, T5 return, T6 mock | `/api/tenant/packages`, `checkout`, `orders` |
| Platform `/admin` | Unchanged (P13 gateway from Sprint 8) | `/api/platform/payment-gateway*` |
| Combined UAT | §0 gateway + §1 checkout | [14-buy-package-spec.md](14-buy-package-spec.md) §11 |

## Sprint 8 — Payment Gateway (customer + tenant UI unchanged)

| Surface | Sprint 8 UX | API |
| --- | --- | --- |
| Customer `/` | Unchanged | Public chat/voice |
| Tenant `/tenant` | Unchanged | No checkout yet |
| Platform `/admin` | **New** P13 payment settings | `/api/platform/payment-gateway*` |
| Ops curl | Callback simulate | `POST /api/callbacks/chillpay` |

```text
┌─────────────────────────────────────────┐
│  Ops / Tester (terminal)                │
│  curl -X POST .../callbacks/chillpay    │
│  -H Content-Type: application/x-www-... │
│  -d TransactionId=...&CheckSum=...      │
└─────────────────────────────────────────┘
```

## Sprint 7 — Platform KYC (customer + tenant UI unchanged)

| Surface | Sprint 7 UX | API |
| --- | --- | --- |
| Customer `/` | Unchanged | Public chat/voice |
| Tenant `/tenant` | Unchanged backoffice submit | `POST /api/tenant/kyc/submit` |
| Platform `/admin` | **New** P12 KYC review + P11 link | `/api/platform/tenants/{id}/kyc*` |
| Ops curl | Approve/reject pending tenant | [12-kyc-tenant-spec.md](12-kyc-tenant-spec.md) §9 |

### Flow T-A — Register

```text
T-A  Open /tenant/register
  → fill company + slug + admin credentials
  → POST /api/public/tenant/register
  → store tokens (sessionStorage, same pattern as platform-admin)
  → redirect /tenant/register/success
```

### Flow T-B — Platform ops (curl or P11)

```text
T-B  platform_admin login
  → GET /api/platform/tenants?status=pending_kyc
  → see acme pending (prep Sprint 7 approve)
```

### Tenant portal — component → file

| Component | Path |
| --- | --- |
| Register page | `apps/tenant-web/src/routes/register/+page.svelte` |
| Success page | `apps/tenant-web/src/routes/register/success/+page.svelte` |
| Login stub | `apps/tenant-web/src/routes/login/+page.svelte` |
| API client | `apps/tenant-web/src/lib/api/register.ts` |
| Session | `apps/tenant-web/src/lib/auth/session.ts` |
| Styles | `apps/tenant-web/src/app.css` |

## Sprint 6 — Tenant register (customer UI unchanged)

| Surface | Sprint 6 UX | API |
| --- | --- | --- |
| Customer `/` | Unchanged | Public chat/voice |
| Tenant `/tenant` | **New** register + success | `POST /api/public/tenant/register` |
| Platform `/admin` | Optional P11 tenants table | `GET /api/platform/tenants` |
| Ops curl | Register + list pending | [11-tenant-register-spec.md](11-tenant-register-spec.md) §10 |

## Sprint 5 — Avatars (customer UI unchanged)

Customer portal **`/`** layout unchanged. Agent cards still bind to `GET /api/workforce`; Sprint 5 only changes the **data source** when tenant assignments exist.

| Surface | Sprint 5 UX | API |
| --- | --- | --- |
| Customer `/` | Same cards/halo/voice | `GET /api/workforce` + `X-Tenant-Id: demo` |
| Platform `/admin` | New Avatars nav + P7–P10 | `/api/platform/avatars*` |
| Ops curl | Assign then probe workforce | See [10-avatars-spec.md](10-avatars-spec.md) §7 |

## Sprint 3 — Auth (no customer UI change)

Customer portal **unchanged** when `AUTH_DISABLED=true` (default). No login screen in Sprint 3.

| Surface | Sprint 3 UX | API |
| --- | --- | --- |
| Customer portal `/` | No login; same as v0.3.0 | Public chat/voice |
| KM admin (ops) | curl / REST client only | `POST /api/auth/login` → Bearer on KM writes |
| Future tenant admin | Deferred Sprint 15+ | — |

```text
┌─────────────────────────────────────────┐
│  Ops / Tester (terminal or REST Client) │
│  POST /api/auth/login                   │
│  → store access_token                   │
│  → Authorization: Bearer on km-seed     │
└─────────────────────────────────────────┘
```

## Customer portal — component → file

| Component | Path |
| --- | --- |
| Page layout | `apps/customer-web/src/routes/+page.svelte` |
| Portrait | `src/lib/components/Portrait.svelte` |
| Waveform | `src/lib/components/Waveform.svelte` |
| Styles | `src/app.css` |
| Chat API | `src/lib/api/chat.ts` |
| Voice | `src/lib/voice/gemini.ts` |

## Sprint 13 — Quota & rate limit (P14)

**What changed vs prior sprints**

| Surface | Change |
| --- | --- |
| Customer portal `/` | No new layout; optional error toast on 429/403 from chat/voice |
| Tenant portal `/tenant` | No change (self-service limits → S16) |
| Platform admin `/admin` | **P14** usage panel on tenant detail (read-only) |
| Ops / curl | Usage API + enforced error codes on hot paths |

### Screen map → API

| Zone | UI | User action | API / WS |
| --- | --- | --- | --- |
| A1 | Nav | Tenants list | existing list tenants |
| A2 | Tenant header | Open `demo` | existing tenant/KYC routes |
| **B1** | Usage card header | View package + period | `GET /api/platform/tenants/{id}/usage` |
| **B2** | Usage bars | Read-only meters | same payload `usage` / `limits` |
| **B3** | Flags | voice / rag on-off | same `limits.*_enabled` |
| **B4** | Refresh | Re-fetch | same GET |
| C1 | Customer chat | Send over rate | `POST /api/chat` → 429 |
| C2 | Customer voice | Over concurrent | `GET /ws/voice` → 429 |
| C3 | KM upload | Over doc limit | `POST /api/km/agents/{id}/documents` → 429 |
| C4 | Avatar assign | Over AI employees | `POST /api/platform/tenants/{id}/avatars` → 429 |

### P14 — Full layout (desktop)

```text
┌─ /admin ──────────────────────────────────────────────────────────────┐
│ [Logo] Packages  Avatars  Tenants  Billing  Settings          [user]  │
├────────────┬──────────────────────────────────────────────────────────┤
│            │  A2 Tenant: demo                                         │
│  A1 Nav    │  Status: active · Package: Starter                       │
│  · Tenants │──────────────────────────────────────────────────────────│
│  · …       │  B1 Usage · Period 2026-07                    [B4 Refresh]│
│            │  ┌────────────────────────────────────────────────────┐  │
│            │  │ B2 Dimension           Usage / Limit               │  │
│            │  │    AI employees        ████░░░░  1 / 2             │  │
│            │  │    Call minutes/mo     █░░░░░░░  12 / 500          │  │
│            │  │    KM documents        ██░░░░░░  3 / 50            │  │
│            │  │    Concurrent calls    ░░░░░░░░  0 / 2             │  │
│            │  │ B3 voice_enabled  yes   rag_enabled  yes           │  │
│            │  └────────────────────────────────────────────────────┘  │
│            │  Empty: "No active package" when status=none             │
└────────────┴──────────────────────────────────────────────────────────┘
```

**Mobile:** single column; nav collapses; bars stack full width; Refresh full-width button.

**Copy (EN primary):** “Usage”, “Limit”, “No active package”, “Refresh”.  
TH optional: “การใช้งาน”, “ขีดจำกัด”, “ไม่มีแพ็กเกจ”.

### Customer error (optional toast)

```text
┌─ Customer / chat ─────────────────────┐
│  …                                    │
│  ┌──────────────────────────────────┐ │
│  │ Limit reached. Try again later.  │ │
│  │ (code: rate_limited)             │ │
│  └──────────────────────────────────┘ │
└───────────────────────────────────────┘
```

### Flow A — KM blocked by quota

```text
1. Operator has Starter (max_km_documents=50), already 50 docs
2. POST /api/km/agents/{id}/documents
3. ← 429 { code: quota_exceeded, dimension: max_km_documents, limit:50, usage:50 }
4. Document not stored; UI shows error
```

### Flow B — Platform views usage

```text
1. Login platform_admin → /admin
2. Tenants → demo
3. GET /api/platform/tenants/demo/usage
4. Fill B1–B3 bars; B4 Refresh re-GETs
```

### Flow C — Concurrent voice denied

```text
1. Tenant at max_concurrent_calls
2. Customer opens GET /ws/voice (or session start)
3. ← 429 quota_exceeded dimension=max_concurrent_calls
4. On disconnect of another session, slot frees → retry succeeds
```

### Component → file (planned)

| Component | Path |
| --- | --- |
| Usage panel (B1–B4) | `apps/platform-admin-web/src/routes/tenants/[id]/+page.svelte` section **or** `.../tenants/[id]/usage/+page.svelte` |
| API client | `apps/platform-admin-web/src/lib/api/usage.ts` |
| Optional chat toast | `apps/customer-web/src/lib/api/chat.ts` / `+page.svelte` |

### Ops-only (no SPA)

```text
┌─ REST client ─────────────────────────────────────────────┐
│  POST /api/auth/login → platform_admin token              │
│  GET  /api/platform/tenants/demo/usage                    │
│  GET  /api/infra | jq .quota,.rate_limit                  │
└───────────────────────────────────────────────────────────┘
```

## Sprint 14 — Embed to Web (T7 tenant · E1 widget)

**What changed**

| Surface | Change |
| --- | --- |
| Tenant `/tenant/embed` | **T7** config: key, origins, snippet, rotate |
| Customer `/` | Unchanged demo |
| Embed `/embed` | **E1** compact chat UI in iframe |
| Host site | Floating launcher via `monti-embed.js` |

### Screen map → API

| Zone | Action | API |
| --- | --- | --- |
| T7 Enable | Toggle + Save | `PUT /api/tenant/embed` |
| T7 Origins | Edit list + Save | same |
| T7 Copy | Copy snippet | client-only |
| T7 Rotate | Confirm rotate | `POST /api/tenant/embed/rotate-key` |
| T7 Load | Open page | `GET /api/tenant/embed` |
| E1 Boot | iframe load | `GET /api/public/embed/{key}` |
| E1 Chat | Send message | `POST /api/chat` + `X-Tenant-Id` |
| E1 Voice | Optional | `GET /ws/voice` |

### T7 — Tenant embed settings

```text
┌─ /tenant/embed ───────────────────────────────────────────┐
│  Embed on your website                                    │
│  [✓] Enabled                                              │
│  Embed key  emb_9f3a…              [Copy] [Rotate key]    │
│  Allowed origins (one per line)                           │
│  ┌────────────────────────────────────────────────────┐   │
│  │ https://shop.example                               │   │
│  │ http://localhost:5500                              │   │
│  └────────────────────────────────────────────────────┘   │
│  Empty list = allow any origin (dev only warning)         │
│  Default agent  [ Ava ▼ ]                                 │
│  [Save]                                                   │
│  Snippet                                                  │
│  ┌────────────────────────────────────────────────────┐   │
│  │ <script src="…/embed/monti-embed.js"               │   │
│  │   data-embed-key="emb_…" async></script>           │   │
│  └────────────────────────────────────────────────────┘   │
│  [Copy snippet]                                           │
└───────────────────────────────────────────────────────────┘
```

### E1 — Embed iframe (compact)

```text
┌─ Host page ──────────────────────────────┐
│  … shop content …                        │
│                         ┌─ Monti ──────┐ │
│                         │ [Ava ▼]      │ │
│                         │ chat…        │ │
│                         │ [Send] [🎤]  │ │
│                         └──────────────┘ │
│                              ( ( ) )     │  ← launcher
└──────────────────────────────────────────┘
```

### Flow A — First-time enable

```text
Login tenant → Embed → Enable → Save → Copy snippet → paste on site → open site → chat
```

### Flow B — Origin denied

```text
Key enabled, allowlist https://shop.example only
Site http://evil.test loads snippet
→ GET /api/public/embed/KEY Origin evil → 403
→ Widget shows error state
```

### Component → file (planned)

| Piece | Path |
| --- | --- |
| Tenant embed page | `apps/tenant-web/src/routes/embed/+page.svelte` |
| Tenant API | `apps/tenant-web/src/lib/api/embed.ts` |
| Loader | static `/embed/monti-embed.js` (server or customer-web static) |
| Embed UI | `apps/customer-web` embed route or query mode |

## Sprint 15 — Tenant KM (T8)

**What changed vs S14:** new **Knowledge** nav item and `/tenant/km` page. Customer portal, embed, and billing UIs unchanged.

### Screen map → API

| Zone | UI | User action | API / WS |
| --- | --- | --- | --- |
| A0 | Shell nav | Open **Knowledge** | — route `/tenant/km` |
| A1 | Page title + help | Read scope explanation | `GET /api/tenant/km/scopes` (cache) |
| B1 | Agent chips/tabs | Select agent | `GET /api/tenant/km/agents` then docs |
| B2 | Overview matrix | View counts by scope | data from agents payload `by_scope` |
| C1 | Upload row | Choose file + scope + Upload | `POST /api/tenant/km/agents/{id}/documents` |
| D1 | Document table | Load list | `GET /api/tenant/km/agents/{id}/documents` |
| D2 | Row scope control | Change scope | `PATCH /api/tenant/km/documents/{id}` |
| D3 | Row delete | Confirm delete | `DELETE /api/tenant/km/documents/{id}` |
| E1 | Reset | Confirm wipe agent KM | `POST /api/tenant/km/agents/{id}/reset` |
| F1 | Toast / banner | Show quota/errors | response body |
| G1 | Knowledge gaps panel | List open unanswered questions | `GET /api/tenant/km/gaps?status=open` |
| G2 | Gap row | Dismiss / mark resolved | `PATCH /api/tenant/km/gaps/{id}` |

### T8 desktop layout

```
┌─ A0 Tenant shell ───────────────────────────────────────────────────┐
│ MONTI TENANT  Backoffice Billing Documents Tax Embed [Knowledge]    │
│                                                         [Logout]    │
├─────────────────────────────────────────────────────────────────────┤
│ A1 Knowledge base                                                   │
│    Upload FAQs for each AI agent. Scope tags match caller topics:   │
│    general · billing · technical.                                   │
│                                                                     │
│ B1 Agents:  ( Ava ● 3 )  ( Max 1 )  ( Luna 0 )                      │
│ B2 ┌ Overview ────────────────────────────────────────────────────┐ │
│    │ Ava  default retrieval: general                              │ │
│    │      docs by scope → general:2  billing:0  technical:1       │ │
│    └──────────────────────────────────────────────────────────────┘ │
│ C1 Upload  [ Choose .md / .txt ]  Scope [ general ▾ ]  [ Upload ]   │
│                                                                     │
│ D1 Documents                                                        │
│ ┌─ filename ────────┬─ scope ───┬─ status ─┬─ chunks ┬─ actions ──┐ │
│ │ faq.md            │ [general▾]│ Ready    │ 12      │ [Delete]   │ │
│ │ tech-notes.md     │ [techni▾] │ Ready    │  8      │ [Delete]   │ │
│ │ old.md            │ general   │ Failed   │  0      │ [Delete]   │ │
│ └───────────────────┴───────────┴──────────┴─────────┴────────────┘ │
│ E1 [ Reset Ava knowledge… ]                                         │
│ G1 Knowledge gaps (unanswered)                                      │
│ ┌─ question ──────────────────┬ agent ┬ count ┬ last ──┬ actions ─┐ │
│ │ Do you offer student disc.? │ ava   │ 3     │ 10:00  │ dismiss  │ │
│ └─────────────────────────────┴───────┴───────┴────────┴──────────┘ │
│ F1 (error/success banner)                                           │
└─────────────────────────────────────────────────────────────────────┘
```

### T8 mobile collapse

```
┌ Knowledge ─────────────┐
│ Agent [ Ava ▾ ]        │
│ general:2 bill:0 tech:1│
│ [file] [scope] [Upload]│
│ faq.md  Ready  [⋮]     │
│  ⋮ Change scope        │
│  ⋮ Delete              │
│ [Reset agent…]         │
└────────────────────────┘
```

### Flows (ASCII)

**Flow A — Upload**

```
Login active tenant_admin
  → /tenant/km
  → B1 select Ava
  → C1 choose faq.md + scope general
  → POST documents
  → D1 row status Ready (indexed)
  → Customer chat (same tenant) can cite faq
```

**Flow B — Change scope**

```
D2 open scope dropdown on row
  → PATCH km_scope=billing
  → row updates; overview B2 recount
```

**Flow C — Delete**

```
D3 Delete → confirm modal
  → DELETE document
  → row removed; B2 counts drop
```

**Flow D — Reset agent**

```
E1 Reset → type-confirm or strong confirm
  → POST reset
  → empty table + empty state CTA
```

**Flow E — Errors**

```
Quota exceeded → F1 red banner with package message
Inactive tenant → redirect / blocked shell
Wrong agent id → toast 400
```

### Empty state (B1 agent with 0 docs)

```
No documents for Ava yet.
Upload a Markdown FAQ (scope general) to ground answers.
[ Choose file ]
```

### Copy (EN primary)

| Key | Text |
| --- | --- |
| Nav | Knowledge |
| Help | Scopes match caller desk topics (General / Billing / Technical). |
| Reset confirm | Delete all knowledge for {agent}? This cannot be undone. |
| Delete confirm | Delete {filename}? Embeddings will be removed. |

### Components → files

| UI | Path |
| --- | --- |
| KM page | `apps/tenant-web/src/routes/km/+page.svelte` |
| API client | `apps/tenant-web/src/lib/api/km.ts` |
| Nav link | `apps/tenant-web/src/routes/+layout.svelte` |

## Sprint 16 — Tenant Settings (T9)

### Screen map → API

| Zone | Action | API |
| --- | --- | --- |
| A0 | Nav Settings | — |
| B1 | Save workspace | `PUT /api/tenant/settings` |
| C1 | View meters | `GET /api/tenant/usage` |
| D1 | Save call limits | `PUT /api/tenant/call-limits` |

### T9 ASCII

```
┌─ Tenant shell … Knowledge [Settings] ───────────────────────┐
│ Settings                                                     │
│ ┌ Workspace ──────────────────────────────────────────────┐ │
│ │ Display name [____________] Locale [th ▾] TZ [Bangkok▾] │ │
│ │ AI reply language [auto ▾]                    [Save]    │ │
│ └─────────────────────────────────────────────────────────┘ │
│ ┌ Usage (this month) ─────────────────────────────────────┐ │
│ │ Minutes  120 / 5000   KM docs  12 / 500   Agents 3 / 10 │ │
│ └─────────────────────────────────────────────────────────┘ │
│ ┌ Call limits ────────────────────────────────────────────┐ │
│ │ Max minutes / call [15]   Max minutes / day [120]       │ │
│ │ Package monthly ceiling: 5000 min             [Save]    │ │
│ └─────────────────────────────────────────────────────────┘ │
│ ┌ Labels (scaffold) ──────────────────────────────────────┐ │
│ │ Tier label [____]  Group label [____]  (full tiers S18) │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Components

| UI | Path |
| --- | --- |
| Settings page | `apps/tenant-web/src/routes/settings/+page.svelte` |
| API client | `apps/tenant-web/src/lib/api/settings.ts` |

## Sprint 17 — Tenant Preview (T10)

### Screen map → API

| Zone | Action | API |
| --- | --- | --- |
| A0 | Nav Preview | — |
| B1 | Select agent / topic | — |
| C1 | Send message | `POST /api/tenant/preview/chat` |
| D1 | Start / end voice | `WS /ws/tenant/preview/voice` |
| E1 | Suggested question chip | fills input |
| F1 | Open embed preview | `GET /api/tenant/embed` → open `/embed?key=` |

### T10 ASCII

```
┌─ Tenant shell … Settings [Preview] ─────────────────────────┐
│ Preview desk                                                 │
│ ┌ banner ─────────────────────────────────────────────────┐ │
│ │ ⚠ Preview mode — does not use package call minutes      │ │
│ │    Rate limits still apply.                             │ │
│ └─────────────────────────────────────────────────────────┘ │
│ Agent [Ava ▾]  Topic [General ▾]   [Start voice] [Hang up] │
│ Suggested: [Greeting] [Billing FAQ] [Tech reset]           │
│ ┌ chat ───────────────────────────────────────────────────┐ │
│ │ Ava: …                                                  │ │
│ │ You: …                                                  │ │
│ └─────────────────────────────────────────────────────────┘ │
│ [ message input ________________________ ] [Send]          │
│ [ Open embed preview ]  or  Enable embed → /tenant/embed   │
└─────────────────────────────────────────────────────────────┘
```

### Flow A — Validate KM

```
Pick agent Luna + topic technical
→ chip "How do I reset my password?"
→ Send
→ reply + sources from tenant KM
```

### Flow B — Minutes not consumed

```
Note Redis monthly minutes
→ Start voice preview ~1 min
→ Hang up
→ monthly + daily counters unchanged
```

### Components → files

| UI | Path |
| --- | --- |
| Preview page | `apps/tenant-web/src/routes/preview/+page.svelte` |
| API client | `apps/tenant-web/src/lib/api/preview.ts` |
| Nav | `apps/tenant-web/src/routes/+layout.svelte` |

## Sprint 18 — Customer Tiers (T11)

### Screen map → API

| Zone | Action | API |
| --- | --- | --- |
| A0 | Nav Tiers | — |
| B1 | List tiers | `GET /api/tenant/tiers` |
| B2 | Create / save tier | `POST/PUT /api/tenant/tiers` |
| B3 | Delete tier | `DELETE /api/tenant/tiers/{id}` |
| C1 | Groups CRUD | `/api/tenant/groups` |
| D1 | Preview with tier | `POST /api/tenant/preview/chat` + `tier_id` |

### T11 ASCII

```
┌─ Tenant shell … Preview [Tiers] ────────────────────────────┐
│ Customer tiers                                               │
│ [+ New tier]                                                 │
│ ┌ list ───────────────────────────────────────────────────┐ │
│ │ VIP   vip   prio 100   locale th   caps 30/day inherit  │ │
│ │ Standard …                                    [Edit] [×] │ │
│ └─────────────────────────────────────────────────────────┘ │
│ ┌ edit form ──────────────────────────────────────────────┐ │
│ │ Name [VIP] Slug [vip] Priority [100] Active [✓]         │ │
│ │ Default agent [Ava ▾]  AI locale [th ▾]                 │ │
│ │ Max min/call [30]  Max min/day [0=inherit]     [Save]   │ │
│ └─────────────────────────────────────────────────────────┘ │
│ Groups (ops labels)  [+ Add]  Retail · Enterprise           │
└─────────────────────────────────────────────────────────────┘
```

### Flow A — Define VIP

```
New tier → name VIP slug vip → locale th → max_minutes_per_call 30 → Save
→ appears in list → Preview tier select VIP → chat uses Thai preference
```

### Components → files

| UI | Path |
| --- | --- |
| Tiers page | `apps/tenant-web/src/routes/tiers/+page.svelte` |
| API client | `apps/tenant-web/src/lib/api/tiers.ts` |

## Sprint 19 — Customer Accounts and Import (T12)

SPRINT-019 adds a tenant-admin customer directory. It does not add a customer-facing account or login screen; the customer conversation portal remains public until SPRINT-020.

### Screen map → API

| Zone | User action | API |
| --- | --- | --- |
| A0 | Open Customers from tenant navigation | — |
| B1 | Search/filter customer directory | `GET /api/tenant/customers?q=&status=&tier_id=` |
| B2 | Create customer | `POST /api/tenant/customers` |
| B3 | Edit profile, tier, and groups | `PUT /api/tenant/customers/{id}` |
| B4 | Deactivate customer | `DELETE /api/tenant/customers/{id}` |
| C1 | Select and validate CSV | `POST /api/tenant/customer-imports` with `dry_run=true` |
| C2 | Confirm CSV commit | `POST /api/tenant/customer-imports` with `dry_run=false` |
| C3 | View import outcome | `GET /api/tenant/customer-imports/{id}` |
| D1 | List domain rules | `GET /api/tenant/customer-domain-rules` |
| D2 | Create domain rule | `POST /api/tenant/customer-domain-rules` |
| D3 | Edit/delete domain rule | `PUT/DELETE /api/tenant/customer-domain-rules/{id}` |
| E1 | Load tier/group selectors | `GET /api/tenant/tiers`, `GET /api/tenant/groups` |

### T12 desktop layout

```text
┌─ Tenant shell ──────────────────────────────────────────────────────────┐
│ MONTI       Workspace / Customers                    Search   AD Admin  │
├──────────────┬──────────────────────────────────────────────────────────┤
│ Overview     │ Customers                                                │
│ Billing      │ Manage imported customer identities before sign-in.      │
│ Documents    │                                                          │
│ Tax          │ [Customers] [CSV imports] [Domain rules]    [+ Customer] │
│ Embed        │                                                          │
│ Knowledge    │ B1 [Search name/email…] [Status ▾] [Tier ▾]              │
│ Tiers        │ ┌──────────────────────────────────────────────────────┐ │
│ Customers A0 │ │ Name      Email           Tier      Groups   Status │ │
│ Settings     │ │ Jane Doe  jane@acme.com   VIP       Retail   Active │ │
│ Preview      │ │ Somchai   s@acme.co.th    Standard  —        Active │ │
│              │ └──────────────────────────────────────────────────────┘ │
│ Enterprise   │                                           [‹] 1 2 [›]    │
│ 68% usage    │                                                          │
└──────────────┴──────────────────────────────────────────────────────────┘
```

### T12 customer drawer (B2/B3/E1)

```text
┌─ Add customer ───────────────────────────────┐
│ Display name* [___________________________]  │
│ Email         [___________________________]  │
│ Phone         [___________________________]  │
│ Locale        [Auto / English / ไทย ▾]       │
│ Tier          [VIP ▾]                        │
│ Groups        [Retail ×] [Enterprise ×] [+]  │
│ Source        [manual]  External ID [_____]  │
│                                             │
│                         [Cancel] [Save] B2   │
└─────────────────────────────────────────────┘
```

### T12 CSV import (C1/C2/C3)

```text
┌─ Import customers ───────────────────────────────────────────────────┐
│ 1 Upload              2 Validate                3 Commit             │
│ [ Download template ]  [ Choose CSV… ]                               │
│                                                                     │
│ Dry-run: 248 accepted · 2 rejected                                  │
│ ┌ Row ┬ Field      ┬ Issue                                      ┐   │
│ │ 14  │ email      │ Invalid email                              │   │
│ │ 91  │ tier_slug  │ Tier not found                            │   │
│ └─────┴────────────┴────────────────────────────────────────────┘   │
│ [Replace file]                                  [Import 248 rows] C2 │
└─────────────────────────────────────────────────────────────────────┘
```

The commit action is disabled until the current file has a successful dry-run. Re-selecting or changing the file clears validation.

### T12 domain rules (D1–D3)

```text
┌─ Domain rules ──────────────────────────────────────────────────────┐
│ Domain          Policy   Default tier   Default group   Active      │
│ acme.com        Allow    VIP            Employees       ●    [Edit] │
│ blocked.test    Deny     —              —               ●    [Edit] │
│                                                     [+ Add rule]    │
│ Note: allow/deny is enforced when customer auth ships in SPRINT-020.│
└────────────────────────────────────────────────────────────────────┘
```

### Flow A — Manual customer

```text
B2 + Customer
  → enter display name and email
  → choose tenant-owned tier/groups (E1)
  → POST customer
  → new active row appears in B1
```

### Flow B — CSV dry-run and commit

```text
C1 choose CSV
  → dry-run parses + validates without customer writes
  → show accepted/rejected counts and row errors
  → user confirms C2
  → send file again as commit
  → C3 completed summary → refresh B1
```

### Flow C — Domain defaults

```text
D2 add acme.com allow + Standard/Employees defaults
  → imported acme.com row has no explicit assignment
  → customer receives domain defaults
  → rule policy remains informational until SPRINT-020 auth
```

### Mobile collapse

- Existing responsive tenant shell becomes bottom navigation.
- Customer table becomes stacked cards with name/email/status first.
- Filters open in a compact sheet; customer editor and import occupy full width.
- Import row errors scroll horizontally only inside their table, not the page.

### Component → file

| Component | Path |
| --- | --- |
| Customers route | `apps/tenant-web/src/routes/customers/+page.svelte` |
| Customer API client | `apps/tenant-web/src/lib/api/customers.ts` |
| Tenant navigation | `apps/tenant-web/src/routes/+layout.svelte` |
| Server handlers | `cmd/server/tenant_customers.go` |
| Store/domain logic | `internal/store/customers.go`, `internal/customerimport/` |

## Sprint 20 — Customer Authentication (T13)

SPRINT-020 adds customer login/account state and tenant customer-auth controls. The public no-auth conversation path remains visible unless tenant settings require customer auth.

### Screen map → API

| Zone | User action | API / WS |
| --- | --- | --- |
| A6 | Open sign-in panel | — |
| A7 | Request email OTP | `POST /api/customer/auth/request-otp` |
| A8 | Verify OTP and sign in | `POST /api/customer/auth/verify-otp` |
| A9 | Refresh session | `POST /api/customer/auth/refresh` |
| A10 | Log out | `POST /api/customer/auth/logout` |
| B6 | Load signed-in customer context | `GET /api/customer/me` |
| B7 | Send signed-in chat | `POST /api/chat` with customer Bearer token |
| B8 | Start signed-in voice | `POST /api/calls` / `WS /ws/voice` with customer Bearer token |
| T13-1 | Tenant admin reads auth settings | `GET /api/tenant/customer-auth/settings` |
| T13-2 | Tenant admin saves auth settings | `PUT /api/tenant/customer-auth/settings` |
| T13-3 | Tenant admin edits domain rules | `GET/POST/PUT/DELETE /api/tenant/customer-domain-rules` |

### Customer portal desktop layout

```text
┌─ Customer portal ─────────────────────────────────────────────────────┐
│ MONTI Inbound Call Center                         [Email OTP] A6     │
├───────────────────────────────┬──────────────────────────────────────┤
│ Agent panel                   │ Caller Desk                          │
│ ┌ Avatar / call controls ───┐ │ Customer: Jane Doe · VIP · Retail B6 │
│ │ Mira                      │ │ [General] [Billing] [Technical]     │
│ │ [Start call] A8           │ │                                      │
│ └───────────────────────────┘ │ transcript ...                       │
│ Agents list                  │ │ [ Ask your question... ] [Send] B7  │
└───────────────────────────────┴──────────────────────────────────────┘
```

### Sign-in panel (A6-A10)

```text
┌─ Sign in to Monti ──────────────────────────────┐
│ Email [____________________________________]    │
│ [Send OTP] A7                                   │
│ OTP code [ _ _ _ _ _ _ ]                        │
│ [Verify and sign in] A8                         │
│                                                 │
│ New here? Same OTP verifies or claims account.  │
│                                                 │
│ Domain message: acme.com allowed by tenant      │
│ Error states: denied domain, inactive customer, │
│ expired OTP, delivery failed, auth disabled     │
└─────────────────────────────────────────────────┘
```

### Tenant auth settings layout

```text
┌─ Tenant shell / Settings / Customer Auth ────────────────────────────┐
│ Customer authentication                                               │
│ Enable customer sign-in  [ on/off ] T13-2                             │
│ Mode                    [ Optional ▾ ]                                │
│ Domain enforcement      [ Allowlist + denylist ▾ ]                    │
│ Public no-auth chat     [ Keep enabled ✓ ]                            │
│ Session TTL             [ 60 ] min   Refresh TTL [ 43200 ] min        │
│                                                                      │
│ Domain rules                                                         │
│ acme.com       Allow      Default tier VIP       Active               │
│ blocked.test   Deny       —                      Active               │
│ [Manage domain rules] T13-3                                          │
│                                                                      │
│ Production gate: quota/rate-limit isolation must pass before launch. │
│ [Save settings]                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

### Flow A — Optional customer sign-in

```text
Tenant admin sets mode optional
  → customer opens portal
  → public chat still available
  → customer requests OTP by email (A7)
  → customer verifies OTP (A8)
  → B6 shows tier/group context
  → B7/B8 send customer token for quota and RAG context
```

### Flow B — Domain enforcement

```text
Tenant admin enables allowlist
  → adds acme.com allow rule
  → customer jane@acme.com requests OTP
  → POST /api/customer/auth/request-otp returns challenge_id
  → email OTP delivered
  → POST /api/customer/auth/verify-otp returns session
  → customer eve@unknown.test requests OTP
  → domain_not_allowed appears in sign-in panel
```

### Flow C — Multi-user quota gate

```text
Tenant A customer starts chat and voice
  → Redis/package counters use tenant A
Tenant B customer starts chat and voice
  → Redis/package counters use tenant B
Cross-tenant customer/session ids
  → 403 or 404 without data leakage
```

### Mobile collapse

- Sign-in moves into a top-right sheet; account state compresses to name + status chip.
- Caller Desk remains the primary panel; agent picker collapses below the active avatar.
- Tenant auth settings use stacked fields and keep domain rules as cards.

### Component → file

| Component | Path |
| --- | --- |
| Customer auth client | `apps/customer-web/src/lib/api/auth.ts` |
| Customer sign-in/account UI | `apps/customer-web/src/routes/+page.svelte` or extracted auth component |
| Tenant customer-auth settings | `apps/tenant-web/src/routes/settings/+page.svelte` or `apps/tenant-web/src/routes/customers/auth/+page.svelte` |
| Tenant customer-auth API client | `apps/tenant-web/src/lib/api/customerAuth.ts` |
| Server customer auth handlers | `cmd/server/customer_auth.go` |
| Store/domain logic | `internal/store/customer_auth.go`, `internal/customerauth/` |

## Sprint 21 — Authenticated Workforce Selection (C14/T14)

S21 changes the customer entry flow: tenant policy can require OTP before the avatar/workforce picker is usable, and quota state is visible before chat/voice start.

### Screen map → API

| UI zone | User action | API / WS |
| --- | --- | --- |
| C14-1 | Load tenant portal policy | `GET /api/customer/portal-policy` |
| C14-2 | Sign in when workforce auth is required | `POST /api/customer/auth/request-otp`, `POST /api/customer/auth/verify-otp` |
| C14-3 | Open workforce picker popup | `GET /api/customer/workforce` |
| C14-4 | View remaining call/chat quota | `GET /api/customer/quota` |
| C14-5 | Start voice call with selected avatar | `POST /api/calls` |
| C14-6 | Send chat with selected avatar | `POST /api/chat` |
| T14-1 | Tenant admin edits auth-required workforce policy | `PUT /api/tenant/customer-auth/settings` |

### Customer portal layout

```text
┌─ MONTI Customer Portal ───────────────────────────────────────────────┐
│ Tenant · libra-tech-co-ltd                  [PUP signed in] [Sign out]│
├───────────────────────────────┬───────────────────────────────────────┤
│ A. Active AI workforce        │ B. Caller Desk                         │
│ ┌───────────────────────────┐ │ Policy: OTP required ✓  Quota: 20m left│
│ │ Ava portrait              │ │ [General] [Billing] [Technical]       │
│ │ General Support           │ │ transcript ...                         │
│ │ [◇ Change avatar] C14-3   │ │ [Ask question...] [Send] C14-6         │
│ │ [00:00:00] [Start] C14-5  │ │                                       │
│ └───────────────────────────┘ │                                       │
└───────────────────────────────┴───────────────────────────────────────┘
```

### Required-auth gate

```text
Tenant policy requires customer auth
  → portal loads C14-1
  → workforce picker and Start call are disabled
  → customer verifies OTP C14-2
  → portal loads workforce C14-3 and quota C14-4
  → customer can select avatar and start chat/voice
```

### Workforce picker popup

```text
┌─ Select AI workforce ────────────────────────────────┐
│ Search [____________________]                        │
│ Quota: 20m remaining today · Max call 5m             │
├──────────────────────────────────────────────────────┤
│ ● Ava    General Support      available   [Select]   │
│ ○ Luna   Technical Support    available   [Select]   │
│ ○ Max    Billing Specialist   unavailable quota      │
└──────────────────────────────────────────────────────┘
```

### Component → file

| Component | Path |
| --- | --- |
| Portal policy/quota/workforce client | `apps/customer-web/src/lib/api/*.ts` |
| Customer auth gate and picker popup | `apps/customer-web/src/routes/+page.svelte` |
| Tenant settings auth/quota controls | `apps/tenant-web/src/routes/settings/+page.svelte` |
| Server policy/workforce/quota handlers | `cmd/server/customer_*.go`, `cmd/server/calls.go` |

## Sprint 22 — Conversation Records and Knowledge Gaps (T15)

S22 adds tenant operator surfaces for archived conversations and knowledge-gap review. Customer UI has no new mandatory screen beyond conversations producing archived records.

### Screen map → API

| UI zone | User action | API / WS |
| --- | --- | --- |
| T15-1 | Open conversation records | `GET /api/tenant/conversation-records` |
| T15-2 | Inspect record detail | `GET /api/tenant/conversation-records/{id}` |
| T15-3 | Retry failed archive | `POST /api/tenant/conversation-records/{id}/archive/retry` |
| T15-4 | Open knowledge gaps | `GET /api/tenant/knowledge-gaps` |
| T15-5 | Resolve/snooze/ignore gap | `PATCH /api/tenant/knowledge-gaps/{id}` |

### Tenant records layout

```text
┌─ Tenant shell / Operations / Conversation Records ───────────────────┐
│ Filters: Date [Today ▾] Avatar [All ▾] Status [Archived ▾] [Search] │
├──────────────────────────────────────────────────────────────────────┤
│ Time        Customer       Avatar   Channel  Duration  Archive  Gaps │
│ 10:00       PUP            Ava      voice    03:00     stored   1    │
│ 09:42       Anonymous      Luna     chat     —         failed   0    │
├──────────────────────────────────────────────────────────────────────┤
│ Detail: crec_01                                                       │
│ Summary, safe transcript preview, object metadata, linked gap ids     │
│ [Retry archive] T15-3                                                 │
└──────────────────────────────────────────────────────────────────────┘
```

### Knowledge gap review layout

```text
┌─ Tenant shell / KM / Knowledge Gaps ─────────────────────────────────┐
│ Status [Open ▾] Reason [No source ▾] Avatar [All ▾]                 │
├──────────────────────────────────────────────────────────────────────┤
│ Question                                      Reason       Action    │
│ What is warranty policy?                      no_source    [Review]  │
├──────────────────────────────────────────────────────────────────────┤
│ Review panel                                                         │
│ Question, answer excerpt, confidence, conversation link              │
│ [Resolve] [Snooze] [Ignore] T15-5                                    │
└──────────────────────────────────────────────────────────────────────┘
```

### Flow A — Archive verification

```text
Customer chat/voice ends
  → server writes MinIO object calls/{tenant}/{call}/transcript.json
  → tenant opens T15-1
  → record shows archive stored or failed
  → failed archive can be retried from T15-3
```

### Flow B — Knowledge gap lifecycle

```text
RAG cannot answer from tenant KM
  → gap candidate created
  → tenant opens T15-4
  → tenant reviews linked conversation
  → tenant resolves/snoozes/ignores T15-5
```

### Component → file

| Component | Path |
| --- | --- |
| Tenant records route | `apps/tenant-web/src/routes/conversation-records/+page.svelte` |
| Tenant knowledge gaps route | `apps/tenant-web/src/routes/knowledge-gaps/+page.svelte` |
| Tenant records/gaps API client | `apps/tenant-web/src/lib/api/operations.ts` |
| Archive/gap handlers | `cmd/server/conversation_records.go`, `cmd/server/knowledge_gaps.go` |
| Archive/gap store | `internal/store/conversation_records.go`, `internal/km/` |

## Sprint 23 - Tickets and Human Escalation (T16)

Customer UI adds a confirmation step for human follow-up. Tenant UI adds a queue and detail workflow; it does not provide live transfer or customer ticket history.

### Screen map -> API

| UI zone | User action | API / WS |
| --- | --- | --- |
| C15-1 | Receive structured human-follow-up offer during chat/voice | Existing chat/voice event stream: `ticket_offer` |
| C15-2 | Confirm human follow-up | `POST /api/customer/tickets` |
| C15-3 | Decline/dismiss offer | No ticket API call; conversation continues |
| T16-1 | Open tenant ticket queue | `GET /api/tenant/tickets` |
| T16-2 | Filter by date, status, priority, category, avatar, assignee | `GET /api/tenant/tickets?...` |
| T16-3 | Inspect ticket, source conversation, and timeline | `GET /api/tenant/tickets/{id}` |
| T16-4 | Change lifecycle, priority, or assignee | `PATCH /api/tenant/tickets/{id}` |
| T16-5 | Add internal note | `POST /api/tenant/tickets/{id}/events` |

### Customer confirmation surface

```text
┌─ Caller Desk / Human follow-up ─────────────────────────────────────┐
│ Ava: I can ask the tenant team to follow up with you.               │
│ Summary: Billing question                                            │
│                                                                      │
│ Contact email [ customer@example.com                    ]           │
│ [Create ticket]                         [Continue with AI]           │
│                                                                      │
│ Your request will be sent to the tenant team. This is not a live    │
│ transfer.                                                            │
└──────────────────────────────────────────────────────────────────────┘
```

Voice clients receive the same `ticket_offer` event and the AI speaks the offer. The client only calls the create API after a confirmed customer action or a voice confirmation mapped to the same idempotency key.

### Tenant ticket queue layout

```text
┌─ Tenant shell / Operations / Tickets ─────────────────────────────────────────────┐
│ Tickets                                                        [New ticket: out] │
│ Date [Today ▾] Status [Open ▾] Priority [All ▾] Category [All ▾] [Search]        │
├───────────────────────────────────────────────┬───────────────────────────────────┤
│ Queue                                         │ Detail: tick_01                    │
│ ID       Subject             Priority Status  │ Need a human follow-up             │
│ tick_01  Billing question    Normal   Open    │ PUP · Ava · voice · 10:02          │
│ tick_02  Account access      High     Working │ Contact c***@example.com           │
│                                               │ Source: crec_01 [Open record]      │
│                                               │ Assignee [Unassigned ▾]             │
│                                               │ Status [Open ▾] Priority [Normal ▾] │
│                                               │ Timeline                           │
│                                               │ • created · customer confirmed     │
│                                               │ Note [___________________________] │
│                                               │ [Add note] [Save changes]           │
└───────────────────────────────────────────────┴───────────────────────────────────┘
```

### Mobile collapse

On narrow screens, T16-1 queue is shown first. Selecting a row opens T16-3 as a full-width detail view with a back control. Filters wrap into a two-row toolbar; the timeline and note composer remain below the ticket metadata so controls never overlap.

### Flow A - Customer confirms escalation

```text
AI detects explicit human-help request or approved escalation signal
    │
    ├─► Send ticket_offer event and speak/display the offer
    │
    ├─► Customer declines ──► Continue conversation; no ticket
    │
    └─► Customer confirms ──► POST /api/customer/tickets
                                │
                                ├─► 201 open ticket
                                └─► 409/4xx safe error state
```

### Flow B - Tenant triages ticket

```text
Tenant opens T16-1
    │
    ├─► GET /api/tenant/tickets with default Today/Open filters
    ├─► Select row ──► GET /api/tenant/tickets/{id}
    ├─► Change status/priority/assignee ──► PATCH ticket
    └─► Add internal note ──► POST ticket event
```

### Component -> file

| Component | Path |
| --- | --- |
| Customer ticket offer and confirmation | `apps/customer-web/src/routes/+page.svelte` |
| Customer ticket API client | `apps/customer-web/src/lib/api/tickets.ts` |
| Tenant tickets route | `apps/tenant-web/src/routes/tickets/+page.svelte` |
| Tenant ticket API client | `apps/tenant-web/src/lib/api/tickets.ts` |
| Ticket handlers | `cmd/server/tickets.go` |
| Ticket store/schema | `internal/store/tickets.go` |
| Ticket event publisher | `internal/tickets/` |

See [09-platform-admin-portal-spec.md](09-platform-admin-portal-spec.md) · [10-avatars-spec.md](10-avatars-spec.md) · [11-tenant-register-spec.md](11-tenant-register-spec.md) · [12-kyc-tenant-spec.md](12-kyc-tenant-spec.md) · [13-payment-gateway-spec.md](13-payment-gateway-spec.md) · [14-buy-package-spec.md](14-buy-package-spec.md) · [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md) · [17-embed-to-web-spec.md](17-embed-to-web-spec.md) · [18-tenant-scope-km-spec.md](18-tenant-scope-km-spec.md) · [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md) · [20-tenant-test-preview-spec.md](20-tenant-test-preview-spec.md) · [21-customer-tier-spec.md](21-customer-tier-spec.md) · [22-customer-account-import-spec.md](22-customer-account-import-spec.md) · [23-customer-auth-spec.md](23-customer-auth-spec.md) · [24-authenticated-workforce-selection-spec.md](24-authenticated-workforce-selection-spec.md) · [25-conversation-records-knowledge-gaps-spec.md](25-conversation-records-knowledge-gaps-spec.md) · [06-auth-spec.md](06-auth-spec.md) · [08-packages-spec.md](08-packages-spec.md) · [02-workflow.md](02-workflow.md) · [03-er-diagram.md](03-er-diagram.md) · [04-api-spec.md](04-api-spec.md).

## Sprint 24 - Customer Satisfaction Review and Tenant Statistics (T17/C16)

Customer UI adds a post-conversation star review prompt for chat and voice. Tenant UI adds an aggregate satisfaction dashboard with a default today range; no customer identity or transcript data is shown in statistics.

### Screen map -> API

| UI zone | User action | API / WS |
| --- | --- | --- |
| C16-1 | Hear/see the post-conversation review prompt | Existing call-close flow; no new request |
| C16-2 | Select 1-5 star score and submit | `POST /api/calls/{id}/rating` |
| C16-3 | Open follow-up prompt for an unrated call | `POST /api/calls/{id}/rating` |
| T17-1 | Open tenant Satisfaction dashboard | `GET /api/tenant/satisfaction/statistics` |
| T17-2 | Change start/end date, avatar, or channel filters | `GET /api/tenant/satisfaction/statistics?...` |
| T17-3 | Inspect KPI, distribution, and breakdowns | Same statistics response; no separate API |

### Customer review surface

```text
┌─ Caller Desk / Call complete ───────────────────────────────────────┐
│ Call complete                                                       │
│ How was your call?                                                  │
│ การให้บริการวันนี้เป็นอย่างไรบ้าง?                                  │
│                                                                     │
│        ☆       ☆       ☆       ☆       ☆                           │
│       1 star   2       3       4       5 stars                     │
│                                                                     │
│ [Submit rating]                                     [Not now]       │
└─────────────────────────────────────────────────────────────────────┘
```

Selecting a star uses a familiar star icon with a visible selected state. The review prompt is shown after the call/session is already closed. Selecting `Not now` leaves the archive and ticket/session state complete; the follow-up prompt may be opened later without reopening the call.

### Tenant satisfaction dashboard

```text
┌─ Tenant shell / Operations / Satisfaction ─────────────────────────────────────────┐
│ Customer satisfaction                                      [Today ▾]                │
│ Start [2026-07-14]  End [2026-07-14]  Avatar [All ▾]  Channel [All ▾] [Apply]     │
├────────────────────────────────────────────────────────────────────────────────────┤
│ Completed conversations 18     Reviewed 15     Completion 83%     Average ★ 4.47    │
├───────────────────────────────┬───────────────────────────────┬────────────────────┤
│ Rating distribution            │ By AI employee                │ By channel         │
│ ★ 5  ████████  8               │ Ava    18 calls · ★ 4.47      │ Voice 12 · ★ 4.40  │
│ ★ 4  ████      4               │                            │ Chat   6 · ★ 4.60  │
│ ★ 3  ██        2               │                            │                    │
│ ★ 2  █         1               │                            │                    │
│ ★ 1            0               │                            │                    │
└───────────────────────────────┴───────────────────────────────┴────────────────────┘
```

### Mobile collapse

On narrow screens, date and dimension filters wrap into two rows. KPI values stay in a two-column grid, then the rating distribution, avatar breakdown, and channel breakdown stack vertically. The customer star buttons retain stable equal-width targets and do not shift when a score is selected.

### Flow A - Customer rates a completed conversation

```text
Chat/voice call ends
    │
    ├─► AI speaks/displays the review request
    │
    ├─► Customer selects a star ──► POST /api/calls/{id}/rating ──► Confirmation
    │
    └─► Customer skips ──► Follow-up prompt available; call remains closed
```

### Flow B - Tenant reviews statistics

```text
Tenant opens T17-1
    │
    ├─► Default Today range loads
    ├─► Tenant adjusts date/avatar/channel ──► GET statistics with query filters
    ├─► No records ──► Show zero-value empty state
    └─► Records exist ──► Show KPI, stars, distribution, avatar and channel breakdowns
```

### Component -> file

| Component | Path |
| --- | --- |
| Customer rating dialog | `apps/customer-web/src/routes/+page.svelte` |
| Customer rating API client | `apps/customer-web/src/lib/api/calls.ts` |
| Tenant satisfaction route | `apps/tenant-web/src/routes/satisfaction/+page.svelte` |
| Tenant satisfaction API client | `apps/tenant-web/src/lib/api/satisfaction.ts` |
| Rating handler | `cmd/server/calls.go` |
| Tenant statistics handler | `cmd/server/tenant_satisfaction.go` |
| Rating schema/store | `internal/store/conversation_records.go` |

See [27-customer-satisfaction-statistics-spec.md](27-customer-satisfaction-statistics-spec.md), [02-workflow.md](02-workflow.md), [03-er-diagram.md](03-er-diagram.md), and [04-api-spec.md](04-api-spec.md).

## Sprint 25 - Tenant Call Center Statistics and Quota Usage (T18)

Sprint 25 adds a tenant operations dashboard. The customer Caller Desk and platform admin surfaces do not change. The existing tenant Settings usage panel remains the package-detail view; T18 is the date-filtered activity and quota overview.

### Screen map -> API

| UI zone | User action | API / WS |
| --- | --- | --- |
| T18-1 Tenant dashboard shell | Open `/tenant/dashboard` | `GET /api/tenant/call-center/statistics` |
| T18-2 Range controls | Set start and end dates | `GET /api/tenant/call-center/statistics?start_date&end_date` |
| T18-3 KPI row | Inspect sessions, channel totals, voice minutes, and average duration | Same statistics response |
| T18-4 Quota panel | Inspect current monthly and daily quota usage | Same statistics response; enforcement remains `/api/tenant/usage` |
| T18-5 Avatar breakdown | Compare sessions and voice minutes by AI employee | Same statistics response |
| T18-6 Channel breakdown | Compare chat and voice activity | Same statistics response |
| T18-7 Freshness/error state | Retry after projection or ClickHouse failure | Repeat `GET /api/tenant/call-center/statistics` |

### Tenant dashboard layout

```text
┌─ Tenant shell / Overview / Call center statistics ─────────────────────────────────────┐
│ Call center statistics / สถิติศูนย์บริการ                         [Tenant scoped]      │
│ Completed conversations and quota usage for the selected date range                   │
├────────────────────────────────────────────────────────────────────────────────────────┤
│ T18-2  Start [2026-07-14]  End [2026-07-14]  [Apply]  [Today]    Source: ClickHouse     │
├──────────────────────┬──────────────────────┬──────────────────────┬───────────────────┤
│ T18-3 Sessions        │ T18-3 Voice minutes │ T18-3 Chat sessions  │ T18-3 Avg duration│
│ 18                   │ 32 min              │ 6                    │ 1m 46s            │
├───────────────────────────────────────┬────────────────────────────────────────────────┤
│ T18-4 Quota usage                     │ T18-5 By AI employee                            │
│ Monthly 32 / 5000 min                 │ Ava       18 sessions · 32 voice min           │
│ Remaining 4968 min                    │                                             │
│ Today   18 / 180 min                  │                                             │
├───────────────────────────────────────┼────────────────────────────────────────────────┤
│ T18-6 By channel                      │ T18-7 Data freshness                            │
│ Voice 12 sessions · 32 min            │ Last projected 10:19:48                         │
│ Chat   6 sessions · 0 min             │ Generated 10:20:00 · Asia/Bangkok              │
└───────────────────────────────────────┴────────────────────────────────────────────────┘
```

The dashboard keeps range activity and current quota values in separate labeled zones. It does not show customer ids, emails, transcripts, ticket notes, ratings comments, or audio object paths.

### Mobile collapse

On screens below 700px, the date controls stack into two rows and the Apply/Today actions occupy a full-width row. KPI cards use a two-column grid. Quota, avatar, channel, and freshness sections stack in that order. Long avatar names wrap within their row; no metric or action overlaps another zone.

### Flow A - Tenant opens the dashboard

```text
Tenant selects Overview / Call center statistics
    |
    +--> Dashboard sets start=end=today in tenant timezone
    |
    +--> GET /api/tenant/call-center/statistics
    |
    +--> Loading state --> KPI, quota, breakdown, and freshness sections
```

### Flow B - Tenant changes the date range

```text
Tenant edits start/end dates
    |
    +--> Invalid range --> Inline validation; no request
    |
    +--> Apply --> GET statistics?start_date&end_date
                    |
                    +--> Data --> Replace all dashboard sections atomically
                    +--> Empty --> Show zero metrics and a date-range hint
```

### Flow C - Analytics dependency is unavailable

```text
Dashboard request
    |
    +--> 503 analytics_unavailable
          |
          +--> Preserve selected date range
          +--> Show "Analytics temporarily unavailable / ไม่สามารถโหลดสถิติได้"
          +--> [Retry] repeats the same request
```

### Component -> file

| Component | Path |
| --- | --- |
| Tenant dashboard route | `apps/tenant-web/src/routes/dashboard/+page.svelte` |
| Dashboard API client and types | `apps/tenant-web/src/lib/api/callCenter.ts` |
| Tenant navigation entry | `apps/tenant-web/src/routes/+layout.svelte` |
| Statistics handler | `cmd/server/tenant_call_center.go` |
| Analytics projection/read service | `internal/analytics/` |
| ClickHouse facts client | `internal/clickhouse/call_center.go` |
| ClickHouse schema/bootstrap | `scripts/migrations/025_call_center_analytics.sql` |

See [28-call-center-statistics-spec.md](28-call-center-statistics-spec.md), [02-workflow.md](02-workflow.md), [03-er-diagram.md](03-er-diagram.md), and [04-api-spec.md](04-api-spec.md).

## Sprint 26 - Tenant System Performance Monitoring (T19)

Sprint 26 adds a tenant-admin monitoring surface. The customer Caller Desk and platform admin surfaces do not change. The view presents normalized service health and analytics freshness; it never displays raw infrastructure errors or customer data.

### Screen map -> API

| UI zone | User action | API / WS |
| --- | --- | --- |
| T19-1 Monitoring shell | Open `/tenant/monitoring` | `GET /api/tenant/system-performance` |
| T19-2 Overall status | Read current overall state | Same monitoring response |
| T19-3 Dependency grid | Inspect component status and latency | Same monitoring response |
| T19-4 Analytics freshness | Inspect current/stale/unavailable projection state | Same monitoring response |
| T19-5 Retry action | Request a fresh snapshot | Repeat `GET /api/tenant/system-performance` |

### Tenant monitoring layout

```text
┌─ Tenant shell / Operations / System performance ───────────────────────────────────────┐
│ System performance / ประสิทธิภาพระบบ                              [Tenant scoped]     │
│ Current service health and analytics freshness                                      │
├────────────────────────────────────────────────────────────────────────────────────────┤
│ T19-2 Overall status   [Degraded / ระบบทำงานช้าบางส่วน]  Checked 10:30:00  [Retry]    │
├───────────────────────┬───────────────────────┬───────────────────────┬────────────────┤
│ T19-3 Postgres         │ T19-3 Redis           │ T19-3 MinIO           │ T19-3 ClickHouse│
│ Operational · 4 ms     │ Operational · 2 ms    │ Operational · 7 ms    │ Operational 9ms│
├───────────────────────┼───────────────────────┼───────────────────────┼────────────────┤
│ T19-3 NATS             │ T19-3 LiveKit         │ T19-3 Gemini          │ T19-4 Analytics │
│ Disabled               │ Operational           │ Degraded              │ Current        │
├───────────────────────┴───────────────────────┴───────────────────────┴────────────────┤
│ Last projected 10:29:48 · Freshness is separate from live dependency status            │
│ No customer, transcript, provider URL, or infrastructure error details are displayed. │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

### Status presentation

| State | Visual treatment | User action |
| --- | --- | --- |
| `operational` | Green status icon and neutral latency label. | No action required. |
| `degraded` | Amber icon and concise Thai/English guidance. | Retry when investigating. |
| `unavailable` | Red icon, no raw error, last checked time retained. | Retry; contact operator if persistent. |
| `disabled` | Neutral icon and `Not configured / ยังไม่ได้ตั้งค่า`. | No retry loop. |
| `stale` | Amber freshness label separate from component health. | Retry or inspect analytics operations. |

### Mobile collapse

Below 700px, the overall status and Retry action remain in the first row with the action wrapping below the timestamp if needed. Dependency cards use a single-column list with fixed label/value rows. Analytics freshness follows the dependency list. No status text, latency value, or Retry action may overlap or force horizontal scrolling.

### Flow A - Tenant opens monitoring

```text
Tenant selects Operations / System performance
    |
    +--> GET /api/tenant/system-performance
    |
    +--> Loading skeleton --> Render overall, dependency, and analytics states
    +--> 401/403 --> Clear protected view and use existing tenant login redirect
    +--> 503 --> Show safe unavailable state and Retry
```

### Flow B - Tenant retries a degraded snapshot

```text
Tenant selects Retry
    |
    +--> Disable Retry while request is pending
    +--> GET /api/tenant/system-performance
    +--> Partial probe failures --> Render 200 normalized degraded state
    +--> Snapshot failure --> Keep last checked time and show safe unavailable state
```

### Flow C - Analytics is stale while services are healthy

```text
Monitoring response
    |
    +--> components all operational
    +--> analytics.status = stale
          |
          +--> Overall = degraded
          +--> Keep dependency cards operational
          +--> Show freshness warning separately with Thai/English guidance
```

### Component -> file

| Component | Path |
| --- | --- |
| Tenant monitoring route | `apps/tenant-web/src/routes/monitoring/+page.svelte` |
| Monitoring API client and types | `apps/tenant-web/src/lib/api/monitoring.ts` |
| Tenant navigation entry | `apps/tenant-web/src/routes/+layout.svelte` |
| Monitoring handler | `cmd/server/tenant_monitoring.go` |
| Probe and status aggregation | `internal/observability/` |
| Existing dependency health adapters | `internal/store/store.go`, `internal/clickhouse/client.go` |

See [29-tenant-system-performance-spec.md](29-tenant-system-performance-spec.md), [02-workflow.md](02-workflow.md), [03-er-diagram.md](03-er-diagram.md), and [04-api-spec.md](04-api-spec.md).

## Sprint 27 - Mobile Call API and SDK (M1)

Sprint 27 adds no customer-web or tenant-admin route. It defines the host mobile call surface and the typed SDK contract that a mobile application embeds. Existing web Caller Desk controls remain unchanged.

### Screen map -> API

| UI zone | User action | API / WS |
| --- | --- | --- |
| M1-1 Bootstrap / avatar picker | Load policy, locale, limits, and assigned AI employees | GET /api/mobile/v1/bootstrap |
| M1-2 Start call | Select avatar and start | POST /api/mobile/v1/calls |
| M1-3 Call shell | Connect and stream audio | GET /ws/mobile/v1/calls/{call_id} |
| M1-4 Status / timer | Refresh lifecycle or remaining time | GET /api/mobile/v1/calls/{call_id} |
| M1-5 Transcript | Read caller and assistant events | transcript events from WebSocket |
| M1-6 End call | Confirm or timeout end | POST /api/mobile/v1/calls/{call_id}/end or end frame |
| M1-7 Review | Select 1-5 star icon score | POST /api/mobile/v1/calls/{call_id}/rating |
| M1-8 Session recovery | Refresh expired customer session | POST /api/customer/auth/refresh |

### Full layout - reference mobile host app

~~~text
┌──────────────────────────────┐
│ MONTI                         │
│ General Support · Ava         │
├──────────────────────────────┤
│                              │
│        [ avatar ]             │  M1-1
│        Ava · General          │
│                              │
│  ┌────────────────────────┐  │
│  │ 00:02:14   Active       │  │  M1-4
│  └────────────────────────┘  │
│                              │
│  Assistant: สวัสดีค่ะ...     │  M1-5
│  Caller: ต้องการสอบถาม...    │
│                              │
│  [ microphone ] [ End call ] │  M1-6
│                              │
│  ★ ★ ★ ★ ★                  │  M1-7 after ended
│  [ Submit review ]            │
└──────────────────────────────┘
~~~

The mobile host app owns microphone permission prompts and platform audio-session configuration. The SDK owns server connection state, frame sizing, reconnect, token refresh, and safe error presentation. No provider credential or raw infrastructure error appears in the mobile UI.

### Flow A - Start a mobile call

~~~text
Open host app
    |
    +--> GET /api/workforce
    +--> Select assigned avatar
    +--> POST /api/mobile/v1/calls with Idempotency-Key
    +--> 201 call_id and ws_path
    +--> Request microphone permission
    +--> GET /ws/mobile/v1/calls/{call_id}
    +--> ready event --> show timer and active transcript
~~~

### Flow B - Recover a session

~~~text
WebSocket or status request returns session_expired
    |
    +--> SDK pauses audio capture
    +--> POST /api/customer/auth/refresh
    +--> Retry WebSocket once with the rotated access token
    +--> If call is still active: resume
    +--> Otherwise: show ended state and offer rating
~~~

The SDK must not retry a create request after an auth failure without a new explicit user action.

### Flow C - End and review

~~~text
Customer taps End call, AI requests timeout close, or quota timer reaches zero
    |
    +--> SDK sends end frame and POST /api/mobile/v1/calls/{id}/end
    +--> Stop microphone and wait for call_status=ended
    +--> Show 1-5 star icon review
    +--> POST /api/mobile/v1/calls/{id}/rating
    +--> Return to idle avatar state
~~~

### Component -> file

| Component | Ownership / location |
| --- | --- |
| Mobile host call screen | Integrator application; not a Monti Svelte route in Sprint 27. |
| Typed SDK client | Package location is selected in the technical-spec approval; native, React Native, Flutter, or layered core-plus-adapters. |
| Mobile REST facade | Go server mobile API package to be created during implementation. |
| Mobile voice adapter | Go live relay adapter to be created during implementation; existing customer relay remains compatible. |
| Contract and UAT tests | Go API/relay tests plus SDK target tests after platform selection. |

### Responsive and accessibility constraints

- The SDK exposes state and events so host apps can keep controls stable while audio permission or reconnect is pending.
- End call remains available during active, ending, and timeout states; duplicate taps are idempotent.
- Review uses familiar star icons with accessible labels such as 5 stars.
- Host apps must reserve a fixed timer/control row so transcript growth cannot overlap the end button.
- Thai and English user-facing status/error labels are host-app concerns; the API returns stable language-neutral codes.

See 30-mobile-call-api-sdk-spec.md, 02-workflow.md §77–79, 03-er-diagram.md, and 04-api-spec.md.

## Sprint 28 - Platform Cross-Tenant Audit Log (A20)

Sprint 28 adds a platform-admin audit surface. Customer and tenant portals have no new user-visible route. The local spool and ClickHouse delivery worker are backend operations and are represented through safe health indicators, never raw file paths or infrastructure details.

### Screen map -> API

| UI zone | User action | API / WS |
| --- | --- | --- |
| A20-1 Date range | Choose UTC start/end dates | `GET /api/platform/audit-logs?start_date=&end_date=` |
| A20-2 Tenant filter | Select one tenant or all | `GET /api/platform/audit-logs?tenant_id=` |
| A20-3 Event filters | Filter actor, action, resource, outcome | `GET /api/platform/audit-logs?actor_id=&action=&resource_type=&outcome=` |
| A20-4 Results table | Browse bounded event rows | `GET /api/platform/audit-logs?limit=&cursor=` |
| A20-5 Delivery status | Inspect queue and transfer freshness | `GET /api/platform/audit-logs/health` |
| A20-6 Next page | Load older matching events | `GET /api/platform/audit-logs?cursor=` |

### Full layout - desktop

~~~text
┌──────────────────────────────────────────────────────────────────────────────┐
│ MONTI · PLATFORM ADMIN                                      ● System healthy │
├───────────────┬──────────────────────────────────────────────────────────────┤
│ Navigation     │ A20  Audit log                                               │
│                 │ Cross-tenant security and operational history                │
│ Overview        │                                                              │
│ Tenants         │ ┌────────────────────────────────────────────────────────┐ │
│ Avatars         │ │ A20-1 Date       A20-2 Tenant       A20-3 Filters       │ │
│ Packages        │ │ [start] [end]    [All tenants v]   [action] [outcome v] │ │
│ Billing         │ │                                      [Apply] [Reset]    │ │
│ Audit log  <    │ └────────────────────────────────────────────────────────┘ │
│ Settings        │                                                              │
│                 │ ┌────────────────────────────────────────────────────────┐ │
│                 │ │ A20-5 Delivery: Spool · 12 queued · Last transfer 5s  │ │
│                 │ │ ClickHouse operational · 2 pending files               │ │
│                 │ └────────────────────────────────────────────────────────┘ │
│                 │                                                              │
│                 │ A20-4 Results                                               │
│                 │ ┌──────────┬──────────┬──────────────┬─────────┬─────────┐ │
│                 │ │ Time     │ Tenant   │ Actor        │ Action  │ Outcome │ │
│                 │ ├──────────┼──────────┼──────────────┼─────────┼─────────┤ │
│                 │ │ 09:10:11│ Libra    │ admin@...    │ Avatar  │ Success │ │
│                 │ │ 09:08:44│ Demo     │ system       │ Call    │ Success │ │
│                 │ └──────────┴──────────┴──────────────┴─────────┴─────────┘ │
│                 │                         [Previous] [Next]                 │
└───────────────┴──────────────────────────────────────────────────────────────┘
~~~

The table is dense and scan-oriented. Event details open in a right-side drawer or inline expansion with the same redacted metadata contract. No nested cards, raw database errors, local paths, secrets, tokens, audio, transcripts, or full request bodies are shown.

### Mobile collapse

Below 700px, the navigation collapses to the existing platform shell menu. Filters stack in this order: date range, tenant, action/resource, outcome, Apply. Each result becomes a stable two-column event row with time/action as the primary line and tenant/actor/outcome below. Delivery status remains above the results. Pagination controls wrap below the table without horizontal scrolling.

### Flow A - Platform admin filters audit history

~~~text
Open Platform / Audit log
    |
    +--> Load default UTC range and GET /api/platform/audit-logs
    +--> GET /api/platform/audit-logs/health in parallel
    +--> Render loading state with fixed table columns
    +--> Apply tenant/action/outcome/date filters
    +--> GET filtered audit logs with cursor reset
    +--> Render rows or explicit empty state
~~~

### Flow B - ClickHouse is unavailable

~~~text
Audit page requests logs
    |
    +--> API returns 503 analytics_unavailable
    +--> Keep filter controls usable
    +--> Show safe degraded banner and last delivery health if available
    +--> Offer Retry
    +--> Never show URL, SQL, local directory, credentials, or raw server error
~~~

### Flow C - Delivery backlog is present

~~~text
Health response
    |
    +--> mode = spool
    +--> pending_files > 0 or oldest_pending_file_age_seconds increases
    +--> Show amber delivery status and last successful transfer
    +--> Keep audit search independent when ClickHouse reads are healthy
    +--> Do not offer a destructive delete/replay action in Sprint 28
~~~

### Component -> file

| Component | Path |
| --- | --- |
| Platform audit route | `apps/platform-admin-web/src/routes/audit-logs/+page.svelte` |
| Audit API client and types | `apps/platform-admin-web/src/lib/api/audit.ts` |
| Platform navigation entry | `apps/platform-admin-web/src/routes/+layout.svelte` |
| Audit list handler | `cmd/server/platform_audit.go` |
| Event contract and writer | `internal/audit/` |
| ClickHouse audit sink | `internal/clickhouse/audit_events.go` |
| Spool worker and retention | `internal/audit/spool.go` |
| Contract/UAT tests | `internal/audit/*_test.go`, platform API tests, `docs/sdlc/06-manual-tests/SPRINT-028-manual.md` |

See 31-cross-tenant-audit-log-spec.md, 02-workflow.md §80–82, 03-er-diagram.md, and 04-api-spec.md.
