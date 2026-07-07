---
id: DES-0005
title: UX/UI — ASCII Wireframes
status: approved
updated: 2026-07-08
sprint: SPRINT-004
---

# UX/UI — ASCII Wireframes

**Surfaces:** Customer portal `apps/customer-web` at `/` · Platform admin `apps/platform-admin-web` at `/admin` *(Sprint 4)*.

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

See [09-platform-admin-portal-spec.md](09-platform-admin-portal-spec.md) · [06-auth-spec.md](06-auth-spec.md) · [08-packages-spec.md](08-packages-spec.md) · [04-api-spec.md](04-api-spec.md) · [02-workflow.md](02-workflow.md).